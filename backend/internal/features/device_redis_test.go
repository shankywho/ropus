package features_test

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shankywho/ropus/backend/internal/features"
)

func getTestRedisClient(t *testing.T) *redis.Client {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6380"
	}

	client := redis.NewClient(&redis.Options{
		Addr:        redisAddr,
		DialTimeout: 2 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		t.Skipf("Skipping Redis integration tests: unable to connect to %s (%v)", redisAddr, err)
		return nil
	}

	return client
}

func TestDeviceFeatureStore_Comprehensive(t *testing.T) {
	client := getTestRedisClient(t)
	if client == nil {
		return
	}
	defer client.Close()

	ctx := context.Background()
	store := features.NewDeviceFeatureStore(client)

	tenantA := "tenant_alpha_001"
	tenantB := "tenant_beta_002"
	canonicalFP := "fp_test_macbook_device_pro_99"
	deviceID_A := features.HashDeviceID(tenantA, canonicalFP)
	deviceID_B := features.HashDeviceID(tenantB, canonicalFP)

	// Clean up keys before tests
	cleanup := func() {
		iter := client.Scan(ctx, 0, fmt.Sprintf("%s:*", tenantA), 0).Iterator()
		for iter.Next(ctx) {
			_ = client.Del(ctx, iter.Val()).Err()
		}
		iterB := client.Scan(ctx, 0, fmt.Sprintf("%s:*", tenantB), 0).Iterator()
		for iterB.Next(ctx) {
			_ = client.Del(ctx, iterB.Val()).Err()
		}
	}
	cleanup()
	defer cleanup()

	t.Run("1. First-Seen Detection & Point-in-Time Ordering", func(t *testing.T) {
		// Point-in-Time: Query BEFORE recording any transaction
		feats, err := store.GetDeviceFeatures(ctx, tenantA, deviceID_A)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if feats.DeviceSeenBefore != 0 {
			t.Errorf("expected device_seen_before = 0 for brand new device, got %d", feats.DeviceSeenBefore)
		}
		if feats.DeviceTxCount1m != 0 || feats.DeviceTxCount1h != 0 || feats.DeviceTxCount24h != 0 {
			t.Errorf("expected 0 velocity before first transaction, got 1m=%d, 1h=%d, 24h=%d",
				feats.DeviceTxCount1m, feats.DeviceTxCount1h, feats.DeviceTxCount24h)
		}

		// Now record transaction 1
		err = store.RecordDeviceTransaction(ctx, tenantA, deviceID_A, "txn_001", 2500, "acc_alice", "tok_visa_1")
		if err != nil {
			t.Fatalf("failed to record transaction: %v", err)
		}

		// Query AFTER recording transaction 1
		featsAfter, err := store.GetDeviceFeatures(ctx, tenantA, deviceID_A)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if featsAfter.DeviceSeenBefore != 1 {
			t.Errorf("expected device_seen_before = 1 after first transaction, got %d", featsAfter.DeviceSeenBefore)
		}
		if featsAfter.DeviceTxCount1m != 1 || featsAfter.DeviceTxCount1h != 1 || featsAfter.DeviceTxCount24h != 1 {
			t.Errorf("expected 1 tx count across windows, got 1m=%d, 1h=%d, 24h=%d",
				featsAfter.DeviceTxCount1m, featsAfter.DeviceTxCount1h, featsAfter.DeviceTxCount24h)
		}
		if featsAfter.DeviceAmountSum24h != 2500 {
			t.Errorf("expected amount_sum_24h = 2500, got %d", featsAfter.DeviceAmountSum24h)
		}
	})

	t.Run("2. Rolling 24h Amount Sum & Distinct Entity Velocity", func(t *testing.T) {
		// Record second transaction with different account and different payment token
		err := store.RecordDeviceTransaction(ctx, tenantA, deviceID_A, "txn_002", 5000, "acc_bob", "tok_visa_2")
		if err != nil {
			t.Fatalf("failed to record transaction 2: %v", err)
		}

		// Record third transaction with duplicate account (acc_alice) and new token (tok_visa_3)
		err = store.RecordDeviceTransaction(ctx, tenantA, deviceID_A, "txn_003", 7500, "acc_alice", "tok_visa_3")
		if err != nil {
			t.Fatalf("failed to record transaction 3: %v", err)
		}

		feats, err := store.GetDeviceFeatures(ctx, tenantA, deviceID_A)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// 3 total transactions
		if feats.DeviceTxCount24h != 3 {
			t.Errorf("expected DeviceTxCount24h = 3, got %d", feats.DeviceTxCount24h)
		}
		// Amount sum: 2500 + 5000 + 7500 = 15000
		if feats.DeviceAmountSum24h != 15000 {
			t.Errorf("expected DeviceAmountSum24h = 15000, got %d", feats.DeviceAmountSum24h)
		}
		// Distinct accounts: acc_alice, acc_bob = 2
		if feats.DeviceUniqueAccounts24h != 2 {
			t.Errorf("expected DeviceUniqueAccounts24h = 2, got %d", feats.DeviceUniqueAccounts24h)
		}
		// Distinct tokens: tok_visa_1, tok_visa_2, tok_visa_3 = 3
		if feats.DeviceUniqueTokens24h != 3 {
			t.Errorf("expected DeviceUniqueTokens24h = 3, got %d", feats.DeviceUniqueTokens24h)
		}
	})

	t.Run("3. Sliding Window Expiration (Time Injection)", func(t *testing.T) {
		tempDevID := "test_sliding_window_dev_001"
		baseTime := time.Date(2026, 8, 21, 12, 0, 0, 0, time.UTC)

		// T0: ₹1000
		_ = store.RecordDeviceTransactionAtTime(ctx, tenantA, tempDevID, "tx_t0", 1000, "acc_1", "tok_1", baseTime)
		// T + 30s: ₹2000
		_ = store.RecordDeviceTransactionAtTime(ctx, tenantA, tempDevID, "tx_t30", 2000, "acc_2", "tok_2", baseTime.Add(30*time.Second))

		// Check at T + 45s: 1m window should have 2 tx (t0, t30) = ₹3000
		f45s, _ := store.GetDeviceFeaturesAtTime(ctx, tenantA, tempDevID, baseTime.Add(45*time.Second))
		if f45s.DeviceTxCount1m != 2 {
			t.Errorf("expected 1m tx count at T+45s = 2, got %d", f45s.DeviceTxCount1m)
		}

		// Check at T + 90s: 1m window should have 1 tx (t30 only; t0 expired)
		f90s, _ := store.GetDeviceFeaturesAtTime(ctx, tenantA, tempDevID, baseTime.Add(90*time.Second))
		if f90s.DeviceTxCount1m != 1 {
			t.Errorf("expected 1m tx count at T+90s = 1, got %d", f90s.DeviceTxCount1m)
		}

		// Now record T + 10m: ₹3000
		_ = store.RecordDeviceTransactionAtTime(ctx, tenantA, tempDevID, "tx_t10m", 3000, "acc_3", "tok_3", baseTime.Add(10*time.Minute))

		// Check at T + 30m: 1h window has all 3 (t0, t30, t10m) = ₹6000
		f30m, _ := store.GetDeviceFeaturesAtTime(ctx, tenantA, tempDevID, baseTime.Add(30*time.Minute))
		if f30m.DeviceTxCount1h != 3 {
			t.Errorf("expected 1h tx count at T+30m = 3, got %d", f30m.DeviceTxCount1h)
		}
		if f30m.DeviceAmountSum24h != 6000 {
			t.Errorf("expected 24h amount at T+30m = 6000, got %d", f30m.DeviceAmountSum24h)
		}

		// Check at T + 2h: 1h window is 0; 24h window still has all 3 = ₹6000
		f2h, _ := store.GetDeviceFeaturesAtTime(ctx, tenantA, tempDevID, baseTime.Add(2*time.Hour))
		if f2h.DeviceTxCount1h != 0 {
			t.Errorf("expected 1h tx count at T+2h = 0, got %d", f2h.DeviceTxCount1h)
		}
		if f2h.DeviceTxCount24h != 3 {
			t.Errorf("expected 24h tx count at T+2h = 3, got %d", f2h.DeviceTxCount24h)
		}

		// Check at T + 25h: 24h window has 0 events and ₹0
		f25h, _ := store.GetDeviceFeaturesAtTime(ctx, tenantA, tempDevID, baseTime.Add(25*time.Hour))
		if f25h.DeviceTxCount24h != 0 || f25h.DeviceAmountSum24h != 0 {
			t.Errorf("expected 24h window expired at T+25h, got count=%d, amt=%d", f25h.DeviceTxCount24h, f25h.DeviceAmountSum24h)
		}
		if f25h.DeviceUniqueAccounts24h != 0 || f25h.DeviceUniqueTokens24h != 0 {
			t.Errorf("expected distinct entities expired at T+25h, got acc=%d, tok=%d", f25h.DeviceUniqueAccounts24h, f25h.DeviceUniqueTokens24h)
		}
	})

	t.Run("4. Tenant Isolation: Same Physical Device Across Tenant A vs Tenant B", func(t *testing.T) {
		// Tenant A has 3 transactions recorded in Test 2
		featsA, _ := store.GetDeviceFeatures(ctx, tenantA, deviceID_A)
		if featsA.DeviceTxCount24h != 3 {
			t.Fatalf("expected Tenant A to have 3 transactions, got %d", featsA.DeviceTxCount24h)
		}

		// Tenant B must see 0 transactions and device_seen_before = 0
		featsB, err := store.GetDeviceFeatures(ctx, tenantB, deviceID_B)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if featsB.DeviceSeenBefore != 0 {
			t.Errorf("CRITICAL SECURITY FLAW: Tenant B observed device_seen_before = 1 from Tenant A!")
		}
		if featsB.DeviceTxCount24h != 0 || featsB.DeviceAmountSum24h != 0 {
			t.Errorf("CRITICAL SECURITY FLAW: Tenant B observed velocity leakage from Tenant A! (24h=%d, sum=%d)",
				featsB.DeviceTxCount24h, featsB.DeviceAmountSum24h)
		}
	})

	t.Run("5. Concurrency: 100 Simultaneous Transactions on Same Device", func(t *testing.T) {
		concurrentDevID := "test_concurrent_device_100"
		var wg sync.WaitGroup
		numWorkers := 100

		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				txID := fmt.Sprintf("concur_tx_%03d", idx)
				accID := fmt.Sprintf("acc_%02d", idx%10) // 10 distinct accounts
				tokID := fmt.Sprintf("tok_%02d", idx%5)  // 5 distinct tokens
				_ = store.RecordDeviceTransaction(ctx, tenantA, concurrentDevID, txID, 100, accID, tokID)
			}(i)
		}
		wg.Wait()

		feats, err := store.GetDeviceFeatures(ctx, tenantA, concurrentDevID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if feats.DeviceTxCount1m != int64(numWorkers) {
			t.Errorf("expected 100 1m transactions, got %d", feats.DeviceTxCount1m)
		}
		if feats.DeviceAmountSum24h != int64(numWorkers*100) {
			t.Errorf("expected total amount = 10000, got %d", feats.DeviceAmountSum24h)
		}
		if feats.DeviceUniqueAccounts24h != 10 {
			t.Errorf("expected 10 distinct accounts, got %d", feats.DeviceUniqueAccounts24h)
		}
		if feats.DeviceUniqueTokens24h != 5 {
			t.Errorf("expected 5 distinct tokens, got %d", feats.DeviceUniqueTokens24h)
		}
		if feats.DeviceSeenBefore != 1 {
			t.Errorf("expected device_seen_before = 1, got %d", feats.DeviceSeenBefore)
		}
	})

	t.Run("6. Redis Key Privacy & TTL Verification", func(t *testing.T) {
		// Verify no Redis key contains the raw fingerprint
		iter := client.Scan(ctx, 0, fmt.Sprintf("*%s*", canonicalFP), 0).Iterator()
		var leakCount int
		for iter.Next(ctx) {
			leakCount++
		}
		if leakCount > 0 {
			t.Errorf("CRITICAL PRIVACY FLAW: Found %d Redis keys containing raw fingerprint string!", leakCount)
		}

		// Verify TTL exists on known key
		knownKey := fmt.Sprintf("%s:dev:known:%s", tenantA, deviceID_A)
		ttl, err := client.TTL(ctx, knownKey).Result()
		if err != nil || ttl <= 0 {
			t.Errorf("expected positive TTL on known device key, got %v, err=%v", ttl, err)
		}
		// Should be close to 90 days (> 80 days)
		if ttl < 80*24*time.Hour {
			t.Errorf("expected ~90-day TTL, got %v", ttl)
		}
	})

	t.Run("7. Graceful Degradation on Nil or Disconnected Store", func(t *testing.T) {
		nilStore := features.NewDeviceFeatureStore(nil)
		feats, err := nilStore.GetDeviceFeatures(ctx, tenantA, deviceID_A)
		if err != nil {
			t.Fatalf("expected nil error on graceful degradation, got %v", err)
		}
		if !feats.IsDegraded {
			t.Errorf("expected IsDegraded = true for nil Redis store")
		}
		if feats.DegradeReason != "DEVICE_FEATURE_STORE_UNAVAILABLE" {
			t.Errorf("expected DegradeReason = DEVICE_FEATURE_STORE_UNAVAILABLE, got %s", feats.DegradeReason)
		}
		if feats.DeviceTxCount1m != 0 || feats.DeviceAmountSum24h != 0 {
			t.Errorf("expected safe zero defaults on degraded feature store")
		}
	})
}
