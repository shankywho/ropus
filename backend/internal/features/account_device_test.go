package features_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/shankywho/ropus/backend/internal/features"
)

func TestAccountDeviceGraph_Comprehensive(t *testing.T) {
	client := getTestRedisClient(t)
	if client == nil {
		return
	}
	defer client.Close()

	ctx := context.Background()
	graphStore := features.NewAccountDeviceGraphStore(client)
	deviceStore := features.NewDeviceFeatureStore(client)

	tenantA := "tenant_graph_alpha"
	tenantB := "tenant_graph_beta"
	canonicalFP_A := "fp_test_macbook_graph_alpha"
	canonicalFP_B := "fp_test_macbook_graph_beta"
	deviceID_A := features.HashDeviceID(tenantA, canonicalFP_A)
	deviceID_B := features.HashDeviceID(tenantB, canonicalFP_B)

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

	t.Run("1. First Account/Device Relationship & Point-in-Time Safety", func(t *testing.T) {
		account1 := "acc_user_alpha_1"

		// Point-in-Time: Query BEFORE recording
		fBefore, err := graphStore.GetGraphFeatures(ctx, tenantA, deviceID_A, account1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if fBefore.DeviceAccountSeenBefore != 0 {
			t.Errorf("expected device_account_seen_before = 0 before write, got %d", fBefore.DeviceAccountSeenBefore)
		}
		if fBefore.DeviceUniqueAccounts1h != 0 || fBefore.DeviceUniqueAccounts24h != 0 {
			t.Errorf("expected 0 accounts before write, got 1h=%d, 24h=%d", fBefore.DeviceUniqueAccounts1h, fBefore.DeviceUniqueAccounts24h)
		}

		// Record transaction
		_ = graphStore.RecordGraphTransaction(ctx, tenantA, deviceID_A, account1, "tx_001")
		_ = deviceStore.RecordDeviceTransaction(ctx, tenantA, deviceID_A, "tx_001", 1000, account1, "tok_1")

		// Query AFTER recording
		fAfter, err := graphStore.GetGraphFeatures(ctx, tenantA, deviceID_A, account1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if fAfter.DeviceAccountSeenBefore != 1 {
			t.Errorf("expected device_account_seen_before = 1 after write, got %d", fAfter.DeviceAccountSeenBefore)
		}
		if fAfter.DeviceUniqueAccounts1h != 1 || fAfter.DeviceUniqueAccounts24h != 1 {
			t.Errorf("expected 1 unique account after write, got 1h=%d, 24h=%d", fAfter.DeviceUniqueAccounts1h, fAfter.DeviceUniqueAccounts24h)
		}
		if fAfter.AccountUniqueDevices1h != 1 || fAfter.AccountUniqueDevices24h != 1 {
			t.Errorf("expected 1 unique device for account, got 1h=%d, 24h=%d", fAfter.AccountUniqueDevices1h, fAfter.AccountUniqueDevices24h)
		}
	})

	t.Run("2. New Account on Established Device", func(t *testing.T) {
		account2 := "acc_user_alpha_2"

		// Device A is already known (established in Test 1). Query with account2 (new to device)
		f, err := graphStore.GetGraphFeatures(ctx, tenantA, deviceID_A, account2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if f.DeviceAccountSeenBefore != 0 {
			t.Errorf("expected device_account_seen_before = 0 for account2, got %d", f.DeviceAccountSeenBefore)
		}
		if f.DeviceNewAccountOnKnownDevice != 1 {
			t.Errorf("expected device_new_account_on_known_device = 1 for new account on known device, got %d", f.DeviceNewAccountOnKnownDevice)
		}

		// Record account2 on device A
		_ = graphStore.RecordGraphTransaction(ctx, tenantA, deviceID_A, account2, "tx_002")
		_ = deviceStore.RecordDeviceTransaction(ctx, tenantA, deviceID_A, "tx_002", 2000, account2, "tok_2")

		fAfter, _ := graphStore.GetGraphFeatures(ctx, tenantA, deviceID_A, account2)
		if fAfter.DeviceAccountSeenBefore != 1 {
			t.Errorf("expected device_account_seen_before = 1 for account2 after recording")
		}
		if fAfter.DeviceUniqueAccounts24h != 2 {
			t.Errorf("expected 2 unique accounts on device A, got %d", fAfter.DeviceUniqueAccounts24h)
		}
	})

	t.Run("3. Account Switching & Repeat Transactions Do Not Inflate Switches", func(t *testing.T) {
		devSwitch := "dev_switch_test_99"
		baseTime := time.Date(2026, 8, 21, 14, 0, 0, 0, time.UTC)

		// Sequence: A -> A -> B -> C -> A
		// Expected switches: A->A (0), A->B (1), B->C (2), C->A (3) = Total 3 switches!
		_ = graphStore.RecordGraphTransactionAtTime(ctx, tenantA, devSwitch, "acc_A", "tx_1", baseTime)
		_ = graphStore.RecordGraphTransactionAtTime(ctx, tenantA, devSwitch, "acc_A", "tx_2", baseTime.Add(1*time.Minute))
		_ = graphStore.RecordGraphTransactionAtTime(ctx, tenantA, devSwitch, "acc_B", "tx_3", baseTime.Add(2*time.Minute))
		_ = graphStore.RecordGraphTransactionAtTime(ctx, tenantA, devSwitch, "acc_C", "tx_4", baseTime.Add(3*time.Minute))
		_ = graphStore.RecordGraphTransactionAtTime(ctx, tenantA, devSwitch, "acc_A", "tx_5", baseTime.Add(4*time.Minute))

		f, err := graphStore.GetGraphFeaturesAtTime(ctx, tenantA, devSwitch, "acc_A", baseTime.Add(5*time.Minute))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.DeviceAccountSwitches1h != 3 || f.DeviceAccountSwitches24h != 3 {
			t.Errorf("expected 3 account switches, got 1h=%d, 24h=%d", f.DeviceAccountSwitches1h, f.DeviceAccountSwitches24h)
		}
		if f.DeviceUniqueAccounts24h != 3 {
			t.Errorf("expected 3 unique accounts (A, B, C), got %d", f.DeviceUniqueAccounts24h)
		}
	})

	t.Run("4. Account Fan-Out: One Account Across Multiple Devices", func(t *testing.T) {
		fanoutAccount := "acc_fanout_king"
		dev1 := "dev_fanout_01"
		dev2 := "dev_fanout_02"
		dev3 := "dev_fanout_03"

		_ = graphStore.RecordGraphTransaction(ctx, tenantA, dev1, fanoutAccount, "tx_f1")
		_ = graphStore.RecordGraphTransaction(ctx, tenantA, dev2, fanoutAccount, "tx_f2")
		_ = graphStore.RecordGraphTransaction(ctx, tenantA, dev3, fanoutAccount, "tx_f3")

		f, err := graphStore.GetGraphFeatures(ctx, tenantA, dev1, fanoutAccount)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.AccountUniqueDevices24h != 3 {
			t.Errorf("expected AccountUniqueDevices24h = 3, got %d", f.AccountUniqueDevices24h)
		}
		if f.MultiAccountSignal != features.MultiAccountLowSignal {
			t.Errorf("expected MultiAccountSignal = LOW_SIGNAL, got %s", f.MultiAccountSignal)
		}
	})

	t.Run("5. Multi-Accounting High Signal Clustering", func(t *testing.T) {
		devMulti := "dev_syndicate_cluster_01"
		for i := 0; i < 11; i++ {
			accID := fmt.Sprintf("acc_syndicate_%d", i)
			_ = graphStore.RecordGraphTransaction(ctx, tenantA, devMulti, accID, fmt.Sprintf("tx_syn_%d", i))
		}

		f, err := graphStore.GetGraphFeatures(ctx, tenantA, devMulti, "acc_syndicate_10")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.DeviceUniqueAccounts24h != 11 {
			t.Errorf("expected 11 unique accounts on cluster device, got %d", f.DeviceUniqueAccounts24h)
		}
		if f.MultiAccountSignal != features.MultiAccountHighSignal {
			t.Errorf("expected MultiAccountSignal = HIGH_SIGNAL, got %s", f.MultiAccountSignal)
		}
	})

	t.Run("6. Tenant Isolation: Complete Graph Namespace Boundary", func(t *testing.T) {
		sharedAccount := "acc_shared_name_across_tenants"

		// Tenant A has relationship
		_ = graphStore.RecordGraphTransaction(ctx, tenantA, deviceID_A, sharedAccount, "tx_ta")

		// Query under Tenant B with same device/account
		fB, err := graphStore.GetGraphFeatures(ctx, tenantB, deviceID_B, sharedAccount)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if fB.DeviceAccountSeenBefore != 0 {
			t.Errorf("CRITICAL SECURITY FLAW: Tenant B observed device_account_seen_before = 1 from Tenant A!")
		}
		if fB.DeviceUniqueAccounts24h != 0 || fB.AccountUniqueDevices24h != 0 {
			t.Errorf("CRITICAL SECURITY FLAW: Tenant B observed graph leakage from Tenant A! (devAcc=%d, accDev=%d)",
				fB.DeviceUniqueAccounts24h, fB.AccountUniqueDevices24h)
		}
	})

	t.Run("7. Account ID Sanitization & Validation", func(t *testing.T) {
		// Valid
		clean, ok := features.SanitizeAccountID("  acc_user_valid_123  ")
		if !ok || clean != "acc_user_valid_123" {
			t.Errorf("expected trimmed valid account, got %s (ok=%t)", clean, ok)
		}

		// Empty
		_, ok = features.SanitizeAccountID("   ")
		if ok {
			t.Errorf("expected false for whitespace-only account ID")
		}

		// Oversized (>256 chars)
		oversized := string(make([]byte, 300))
		_, ok = features.SanitizeAccountID(oversized)
		if ok {
			t.Errorf("expected false for oversized account ID")
		}

		// Embedded Null Byte / Control chars
		_, ok = features.SanitizeAccountID("acc_user\x00_hacked")
		if ok {
			t.Errorf("expected false for embedded null byte in account ID")
		}
	})

	t.Run("8. Concurrency: 100 Simultaneous Graph Updates", func(t *testing.T) {
		devConcur := "dev_graph_concur_100"
		var wg sync.WaitGroup
		numWorkers := 100

		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				accID := fmt.Sprintf("acc_worker_%02d", idx%10) // 10 distinct accounts
				txID := fmt.Sprintf("tx_worker_%03d", idx)
				_ = graphStore.RecordGraphTransaction(ctx, tenantA, devConcur, accID, txID)
			}(i)
		}
		wg.Wait()

		f, err := graphStore.GetGraphFeatures(ctx, tenantA, devConcur, "acc_worker_01")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.DeviceUniqueAccounts24h != 10 {
			t.Errorf("expected 10 distinct accounts under concurrent load, got %d", f.DeviceUniqueAccounts24h)
		}
		if f.DeviceAccountSeenBefore != 1 {
			t.Errorf("expected device_account_seen_before = 1, got %d", f.DeviceAccountSeenBefore)
		}
	})

	t.Run("9. Graceful Degradation on Disconnected Store", func(t *testing.T) {
		nilStore := features.NewAccountDeviceGraphStore(nil)
		f, err := nilStore.GetGraphFeatures(ctx, tenantA, deviceID_A, "acc_1")
		if err != nil {
			t.Fatalf("expected nil error on degradation, got %v", err)
		}
		if !f.IsDegraded || f.DegradeReason != "ACCOUNT_DEVICE_GRAPH_UNAVAILABLE" {
			t.Errorf("expected IsDegraded=true and DegradeReason=ACCOUNT_DEVICE_GRAPH_UNAVAILABLE")
		}
		if f.DeviceUniqueAccounts24h != 0 || f.DeviceAccountSeenBefore != 0 {
			t.Errorf("expected safe zero defaults on degraded graph store")
		}
	})
}
