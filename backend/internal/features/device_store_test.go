package features_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shankywho/ropus/backend/internal/features"
)

func getTestDBPool(t *testing.T) *pgxpool.Pool {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://risk_user:risk_password@localhost:5433/risk_engine?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Skipf("Skipping PostgreSQL integration test: unable to connect to %s (%v)", dbURL, err)
		return nil
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("Skipping PostgreSQL integration test: ping failed to %s (%v)", dbURL, err)
		return nil
	}

	return pool
}

func ensureTestTenant(ctx context.Context, pool *pgxpool.Pool, tenantID string) error {
	query := `
		INSERT INTO tenants (tenant_id, name, api_key_hash, status)
		VALUES ($1, $2, $3, 'ACTIVE')
		ON CONFLICT (tenant_id) DO NOTHING
	`
	_, err := pool.Exec(ctx, query, tenantID, "Test Tenant "+tenantID[:8], "key_"+tenantID)
	return err
}

func TestDeviceStore_Integration(t *testing.T) {
	pool := getTestDBPool(t)
	if pool == nil {
		return
	}
	defer pool.Close()

	ctx := context.Background()
	store := features.NewDeviceStore(pool)

	tenantA := "00000000-0000-0000-0000-000000000001"
	tenantB := "00000000-0000-0000-0000-000000000002"

	_ = ensureTestTenant(ctx, pool, tenantA)
	_ = ensureTestTenant(ctx, pool, tenantB)

	// Clean up any existing data for these test tenants before running
	_, _ = pool.Exec(ctx, "DELETE FROM device_events WHERE tenant_id IN ($1, $2)", tenantA, tenantB)
	_, _ = pool.Exec(ctx, "DELETE FROM device_payment_instruments WHERE tenant_id IN ($1, $2)", tenantA, tenantB)
	_, _ = pool.Exec(ctx, "DELETE FROM device_accounts WHERE tenant_id IN ($1, $2)", tenantA, tenantB)
	_, _ = pool.Exec(ctx, "DELETE FROM device_reputation WHERE tenant_id IN ($1, $2)", tenantA, tenantB)
	_, _ = pool.Exec(ctx, "DELETE FROM devices WHERE tenant_id IN ($1, $2)", tenantA, tenantB)

	canonicalFP := "fp_test_macbook_intel_hash_9999"
	deviceHashA := features.HashDeviceID(tenantA, canonicalFP)
	deviceHashB := features.HashDeviceID(tenantB, canonicalFP)

	t.Run("1. Upsert Device First Seen", func(t *testing.T) {
		devA, err := store.UpsertDeviceSeen(ctx, tenantA, deviceHashA, "enc_ciphertext_dummy_123")
		if err != nil {
			t.Fatalf("failed to upsert device: %v", err)
		}

		if devA.DeviceID == "" {
			t.Errorf("expected generated device UUID")
		}
		if devA.DeviceHash != deviceHashA {
			t.Errorf("expected deviceHash %s, got %s", deviceHashA, devA.DeviceHash)
		}
		if devA.TotalTxCount != 1 {
			t.Errorf("expected initial total_tx_count=1, got %d", devA.TotalTxCount)
		}
		if devA.TrustScore != 50 {
			t.Errorf("expected baseline trust_score=50, got %d", devA.TrustScore)
		}
	})

	t.Run("2. Upsert Device Repeat Seen (Deduplication & Counter Increment)", func(t *testing.T) {
		// Second upsert should increment counter to 2 and update last_seen_at
		devA2, err := store.UpsertDeviceSeen(ctx, tenantA, deviceHashA, "enc_ciphertext_dummy_123")
		if err != nil {
			t.Fatalf("failed to re-upsert device: %v", err)
		}

		if devA2.TotalTxCount != 2 {
			t.Errorf("expected total_tx_count=2 after second seen, got %d", devA2.TotalTxCount)
		}

		// Verify count in database is still exactly 1 row for (tenantA, deviceHashA)
		var count int
		err = pool.QueryRow(ctx, "SELECT count(*) FROM devices WHERE tenant_id = $1 AND device_hash = $2", tenantA, deviceHashA).Scan(&count)
		if err != nil || count != 1 {
			t.Errorf("expected exactly 1 row in devices table, got count=%d, err=%v", count, err)
		}
	})

	t.Run("3. Tenant Isolation: Same Fingerprint Across Tenant A vs Tenant B", func(t *testing.T) {
		devB, err := store.UpsertDeviceSeen(ctx, tenantB, deviceHashB, "enc_ciphertext_dummy_b")
		if err != nil {
			t.Fatalf("failed to upsert device for tenant B: %v", err)
		}

		devA, err := store.GetDeviceByHash(ctx, tenantA, deviceHashA)
		if err != nil {
			t.Fatalf("failed to get device A: %v", err)
		}

		// Devices must have distinct database UUIDs and distinct hashes
		if devA.DeviceID == devB.DeviceID {
			t.Errorf("CRITICAL SECURITY FLAW: Tenant A and Tenant B shared identical device UUID!")
		}
		if devA.DeviceHash == devB.DeviceHash {
			t.Errorf("CRITICAL SECURITY FLAW: Tenant A and Tenant B shared identical deviceHash!")
		}

		// Querying under Tenant A with Tenant B's hash must return ErrDeviceNotFound
		_, err = store.GetDeviceByHash(ctx, tenantA, deviceHashB)
		if err != features.ErrDeviceNotFound {
			t.Errorf("expected ErrDeviceNotFound when cross-querying tenant B hash under tenant A, got %v", err)
		}
	})

	t.Run("4. Device ↔ Account Linkage & Deduplication", func(t *testing.T) {
		devA, err := store.GetDeviceByHash(ctx, tenantA, deviceHashA)
		if err != nil {
			t.Fatalf("failed to fetch device A: %v", err)
		}

		account1 := "acc_user_customer_42"
		err = store.LinkAccount(ctx, tenantA, devA.DeviceID, account1)
		if err != nil {
			t.Fatalf("failed to link account: %v", err)
		}

		// Link again (should increment tx count and update timestamp without error)
		err = store.LinkAccount(ctx, tenantA, devA.DeviceID, account1)
		if err != nil {
			t.Fatalf("failed to re-link account: %v", err)
		}

		var linkCount int
		var txCount int64
		err = pool.QueryRow(ctx, "SELECT count(*), transaction_count FROM device_accounts WHERE tenant_id = $1 AND device_id = $2 AND account_id = $3 GROUP BY transaction_count",
			tenantA, devA.DeviceID, account1).Scan(&linkCount, &txCount)
		if err != nil || linkCount != 1 || txCount != 2 {
			t.Errorf("expected 1 row with transaction_count=2, got count=%d, txCount=%d, err=%v", linkCount, txCount, err)
		}
	})

	t.Run("5. Device ↔ Payment Token Linkage & Deduplication", func(t *testing.T) {
		devA, err := store.GetDeviceByHash(ctx, tenantA, deviceHashA)
		if err != nil {
			t.Fatalf("failed to fetch device A: %v", err)
		}

		paymentToken := "tok_visa_gold_4242"
		err = store.LinkPaymentToken(ctx, tenantA, devA.DeviceID, paymentToken)
		if err != nil {
			t.Fatalf("failed to link payment token: %v", err)
		}

		// Re-link same token
		err = store.LinkPaymentToken(ctx, tenantA, devA.DeviceID, paymentToken)
		if err != nil {
			t.Fatalf("failed to re-link token: %v", err)
		}

		var rowCount int
		var txCount int64
		err = pool.QueryRow(ctx, "SELECT count(*), transaction_count FROM device_payment_instruments WHERE tenant_id = $1 AND device_id = $2 AND payment_token = $3 GROUP BY transaction_count",
			tenantA, devA.DeviceID, paymentToken).Scan(&rowCount, &txCount)
		if err != nil || rowCount != 1 || txCount != 2 {
			t.Errorf("expected 1 row with transaction_count=2, got count=%d, txCount=%d, err=%v", rowCount, txCount, err)
		}
	})

	t.Run("6. Device Event Ledger Insertion", func(t *testing.T) {
		devA, err := store.GetDeviceByHash(ctx, tenantA, deviceHashA)
		if err != nil {
			t.Fatalf("failed to fetch device A: %v", err)
		}

		event := &features.DeviceEventRecord{
			TenantID:     tenantA,
			DeviceID:     devA.DeviceID,
			EventType:    "TRANSACTION",
			AccountID:    "acc_user_customer_42",
			PaymentToken: "tok_visa_gold_4242",
			EventTime:    time.Now().UTC(),
			Amount:       15000,
			Currency:     "INR",
			Metadata: map[string]interface{}{
				"risk_score": 12,
				"action":     "ALLOW_RECOMMENDATION",
			},
		}

		err = store.RecordDeviceEvent(ctx, event)
		if err != nil {
			t.Fatalf("failed to record device event: %v", err)
		}

		var eventCount int
		err = pool.QueryRow(ctx, "SELECT count(*) FROM device_events WHERE tenant_id = $1 AND device_id = $2", tenantA, devA.DeviceID).Scan(&eventCount)
		if err != nil || eventCount < 1 {
			t.Errorf("expected at least 1 device event, got %d, err=%v", eventCount, err)
		}
	})

	t.Run("7. Privacy: Raw Fingerprint Is Never Stored in Device Table", func(t *testing.T) {
		var storedHash, storedEncrypted string
		err := pool.QueryRow(ctx, "SELECT device_hash, COALESCE(encrypted_fingerprint, '') FROM devices WHERE tenant_id = $1 AND device_hash = $2", tenantA, deviceHashA).Scan(&storedHash, &storedEncrypted)
		if err != nil {
			t.Fatalf("failed to query device row: %v", err)
		}

		if storedHash == canonicalFP {
			t.Errorf("CRITICAL PRIVACY FLAW: Raw fingerprint stored directly in device_hash column!")
		}
		if storedEncrypted == canonicalFP {
			t.Errorf("CRITICAL PRIVACY FLAW: Raw fingerprint stored unencrypted in encrypted_fingerprint column!")
		}
		if len(storedHash) != 64 {
			t.Errorf("expected 64-char SHA256 hex hash, got %d chars", len(storedHash))
		}
	})

	// Cleanup test data
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, "DELETE FROM device_events WHERE tenant_id IN ($1, $2)", tenantA, tenantB)
		_, _ = pool.Exec(ctx, "DELETE FROM device_payment_instruments WHERE tenant_id IN ($1, $2)", tenantA, tenantB)
		_, _ = pool.Exec(ctx, "DELETE FROM device_accounts WHERE tenant_id IN ($1, $2)", tenantA, tenantB)
		_, _ = pool.Exec(ctx, "DELETE FROM device_reputation WHERE tenant_id IN ($1, $2)", tenantA, tenantB)
		_, _ = pool.Exec(ctx, "DELETE FROM devices WHERE tenant_id IN ($1, $2)", tenantA, tenantB)
	})
}
