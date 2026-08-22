package features_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/shankywho/ropus/backend/internal/features"
)

func TestPaymentTokenStore_Comprehensive(t *testing.T) {
	client := getTestRedisClient(t)
	if client == nil {
		return
	}
	defer client.Close()

	ctx := context.Background()
	tokenStore := features.NewPaymentTokenStore(client)
	deviceStore := features.NewDeviceFeatureStore(client)

	tenantA := "tenant_tok_alpha"
	tenantB := "tenant_tok_beta"
	canonicalFP_A := "fp_test_macbook_tok_alpha"
	canonicalFP_B := "fp_test_macbook_tok_beta"
	deviceID_A := features.HashDeviceID(tenantA, canonicalFP_A)
	deviceID_B := features.HashDeviceID(tenantB, canonicalFP_B)

	cleanup := func() {
		iterA := client.Scan(ctx, 0, fmt.Sprintf("%s:*", tenantA), 0).Iterator()
		for iterA.Next(ctx) {
			_ = client.Del(ctx, iterA.Val()).Err()
		}
		iterB := client.Scan(ctx, 0, fmt.Sprintf("%s:*", tenantB), 0).Iterator()
		for iterB.Next(ctx) {
			_ = client.Del(ctx, iterB.Val()).Err()
		}
	}
	cleanup()
	defer cleanup()

	t.Run("1. First Device/Token Relationship & Point-in-Time Safety", func(t *testing.T) {
		rawToken1 := "tok_visa_card_001"

		// Query BEFORE recording (Point-in-Time Safe)
		fBefore, err := tokenStore.GetPaymentTokenFeatures(ctx, tenantA, deviceID_A, rawToken1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if fBefore.DeviceTokenSeenBefore != 0 {
			t.Errorf("expected device_token_seen_before = 0 before write, got %d", fBefore.DeviceTokenSeenBefore)
		}
		if fBefore.DeviceUniqueTokens5m != 0 || fBefore.DeviceUniqueTokens24h != 0 {
			t.Errorf("expected 0 tokens on device before write, got 5m=%d, 24h=%d", fBefore.DeviceUniqueTokens5m, fBefore.DeviceUniqueTokens24h)
		}

		// Record transaction
		_ = tokenStore.RecordPaymentTokenTransaction(ctx, tenantA, deviceID_A, rawToken1, "tx_01", 5000)
		_ = deviceStore.RecordDeviceTransaction(ctx, tenantA, deviceID_A, "tx_01", 5000, "acc_1", rawToken1)

		// Query AFTER recording
		fAfter, err := tokenStore.GetPaymentTokenFeatures(ctx, tenantA, deviceID_A, rawToken1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if fAfter.DeviceTokenSeenBefore != 1 {
			t.Errorf("expected device_token_seen_before = 1 after write, got %d", fAfter.DeviceTokenSeenBefore)
		}
		if fAfter.DeviceUniqueTokens5m != 1 || fAfter.DeviceUniqueTokens24h != 1 {
			t.Errorf("expected 1 token on device after write, got 5m=%d, 24h=%d", fAfter.DeviceUniqueTokens5m, fAfter.DeviceUniqueTokens24h)
		}
		if fAfter.TokenUniqueDevices1h != 1 || fAfter.TokenUniqueDevices24h != 1 {
			t.Errorf("expected 1 device on token, got 1h=%d, 24h=%d", fAfter.TokenUniqueDevices1h, fAfter.TokenUniqueDevices24h)
		}
		if fAfter.DeviceTokenAmountSum24h != 5000 {
			t.Errorf("expected DeviceTokenAmountSum24h = 5000, got %d", fAfter.DeviceTokenAmountSum24h)
		}
	})

	t.Run("2. Repeat Transaction With Same Token Does Not Inflate Unique Count", func(t *testing.T) {
		rawToken1 := "tok_visa_card_001"

		// Record second transaction with SAME token
		_ = tokenStore.RecordPaymentTokenTransaction(ctx, tenantA, deviceID_A, rawToken1, "tx_02", 3000)

		f, err := tokenStore.GetPaymentTokenFeatures(ctx, tenantA, deviceID_A, rawToken1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if f.DeviceUniqueTokens5m != 1 || f.DeviceUniqueTokens24h != 1 {
			t.Errorf("expected unique tokens to remain 1, got 5m=%d, 24h=%d", f.DeviceUniqueTokens5m, f.DeviceUniqueTokens24h)
		}
		if f.DeviceTokenTxCount5m != 2 {
			t.Errorf("expected DeviceTokenTxCount5m = 2, got %d", f.DeviceTokenTxCount5m)
		}
		if f.DeviceTokenAmountSum24h != 8000 {
			t.Errorf("expected amount sum = 8000 (5000+3000), got %d", f.DeviceTokenAmountSum24h)
		}
	})

	t.Run("3. Card Testing Detection: Rapid Token Rotation & Thresholds", func(t *testing.T) {
		devCardTester := "dev_card_testing_target_99"
		baseTime := time.Date(2026, 8, 21, 14, 0, 0, 0, time.UTC)

		// 1. Low Signal: 3 distinct tokens within 5 minutes
		for i := 1; i <= 3; i++ {
			_ = tokenStore.RecordPaymentTokenTransactionAtTime(ctx, tenantA, devCardTester, fmt.Sprintf("tok_card_test_%02d", i), fmt.Sprintf("tx_ct_%02d", i), 100, baseTime.Add(time.Duration(i)*time.Second))
		}
		fLow, _ := tokenStore.GetPaymentTokenFeaturesAtTime(ctx, tenantA, devCardTester, "tok_card_test_01", baseTime.Add(1*time.Minute))
		if fLow.DeviceUniqueTokens5m != 3 {
			t.Errorf("expected 3 unique tokens in 5m, got %d", fLow.DeviceUniqueTokens5m)
		}
		if fLow.CardTestingSignal != features.CardTestingLowSignal {
			t.Errorf("expected CardTestingSignal = LOW_SIGNAL, got %s", fLow.CardTestingSignal)
		}

		// 2. Suspicious Signal: 5 distinct tokens within 5 minutes
		for i := 4; i <= 5; i++ {
			_ = tokenStore.RecordPaymentTokenTransactionAtTime(ctx, tenantA, devCardTester, fmt.Sprintf("tok_card_test_%02d", i), fmt.Sprintf("tx_ct_%02d", i), 100, baseTime.Add(time.Duration(i)*time.Second))
		}
		fSusp, _ := tokenStore.GetPaymentTokenFeaturesAtTime(ctx, tenantA, devCardTester, "tok_card_test_01", baseTime.Add(2*time.Minute))
		if fSusp.DeviceUniqueTokens5m != 5 {
			t.Errorf("expected 5 unique tokens in 5m, got %d", fSusp.DeviceUniqueTokens5m)
		}
		if fSusp.CardTestingSignal != features.CardTestingSuspicious {
			t.Errorf("expected CardTestingSignal = SUSPICIOUS, got %s", fSusp.CardTestingSignal)
		}

		// 3. High Signal: 10 distinct tokens within 5 minutes
		for i := 6; i <= 10; i++ {
			_ = tokenStore.RecordPaymentTokenTransactionAtTime(ctx, tenantA, devCardTester, fmt.Sprintf("tok_card_test_%02d", i), fmt.Sprintf("tx_ct_%02d", i), 100, baseTime.Add(time.Duration(i)*time.Second))
		}
		fHigh, _ := tokenStore.GetPaymentTokenFeaturesAtTime(ctx, tenantA, devCardTester, "tok_card_test_01", baseTime.Add(3*time.Minute))
		if fHigh.DeviceUniqueTokens5m != 10 {
			t.Errorf("expected 10 unique tokens in 5m, got %d", fHigh.DeviceUniqueTokens5m)
		}
		if fHigh.CardTestingSignal != features.CardTestingHighSignal {
			t.Errorf("expected CardTestingSignal = HIGH_SIGNAL, got %s", fHigh.CardTestingSignal)
		}
	})

	t.Run("4. Token Fan-Out: Single Token Across Multiple Devices", func(t *testing.T) {
		sharedToken := "tok_compromised_visa_4242"
		dev1 := "dev_stolen_01"
		dev2 := "dev_stolen_02"
		dev3 := "dev_stolen_03"
		dev4 := "dev_stolen_04"
		dev5 := "dev_stolen_05"

		_ = tokenStore.RecordPaymentTokenTransaction(ctx, tenantA, dev1, sharedToken, "tx_s1", 1000)
		_ = tokenStore.RecordPaymentTokenTransaction(ctx, tenantA, dev2, sharedToken, "tx_s2", 1000)
		_ = tokenStore.RecordPaymentTokenTransaction(ctx, tenantA, dev3, sharedToken, "tx_s3", 1000)

		f3, _ := tokenStore.GetPaymentTokenFeatures(ctx, tenantA, dev1, sharedToken)
		if f3.TokenUniqueDevices1h != 3 || f3.TokenUniqueDevices24h != 3 {
			t.Errorf("expected TokenUniqueDevices = 3, got 1h=%d, 24h=%d", f3.TokenUniqueDevices1h, f3.TokenUniqueDevices24h)
		}
		if f3.TokenFanOutSignal != features.TokenFanOutSuspicious {
			t.Errorf("expected TokenFanOutSignal = SUSPICIOUS, got %s", f3.TokenFanOutSignal)
		}

		_ = tokenStore.RecordPaymentTokenTransaction(ctx, tenantA, dev4, sharedToken, "tx_s4", 1000)
		_ = tokenStore.RecordPaymentTokenTransaction(ctx, tenantA, dev5, sharedToken, "tx_s5", 1000)

		f5, _ := tokenStore.GetPaymentTokenFeatures(ctx, tenantA, dev1, sharedToken)
		if f5.TokenUniqueDevices1h != 5 {
			t.Errorf("expected TokenUniqueDevices = 5, got %d", f5.TokenUniqueDevices1h)
		}
		if f5.TokenFanOutSignal != features.TokenFanOutHighSignal {
			t.Errorf("expected TokenFanOutSignal = HIGH_SIGNAL, got %s", f5.TokenFanOutSignal)
		}
	})

	t.Run("5. Tenant Isolation: Complete Separation of Token Namespaces", func(t *testing.T) {
		tokenName := "tok_common_shared_name"

		// Record in Tenant A
		_ = tokenStore.RecordPaymentTokenTransaction(ctx, tenantA, deviceID_A, tokenName, "tx_ta", 5000)

		// Query in Tenant B
		fB, err := tokenStore.GetPaymentTokenFeatures(ctx, tenantB, deviceID_B, tokenName)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if fB.DeviceTokenSeenBefore != 0 {
			t.Errorf("CRITICAL SECURITY LEAK: Tenant B observed device_token_seen_before = 1 from Tenant A!")
		}
		if fB.DeviceUniqueTokens24h != 0 || fB.TokenUniqueDevices24h != 0 || fB.DeviceTokenAmountSum24h != 0 {
			t.Errorf("CRITICAL SECURITY LEAK: Tenant B observed token metrics from Tenant A! (toks=%d, devs=%d, amt=%d)",
				fB.DeviceUniqueTokens24h, fB.TokenUniqueDevices24h, fB.DeviceTokenAmountSum24h)
		}
	})

	t.Run("6. Token Sanitization & Raw PAN Rejection", func(t *testing.T) {
		// Valid synthetic token
		clean, ok := features.SanitizePaymentToken("  tok_visa_valid_123  ")
		if !ok || clean != "tok_visa_valid_123" {
			t.Errorf("expected clean token, got %s (ok=%t)", clean, ok)
		}

		// Empty
		_, ok = features.SanitizePaymentToken("   ")
		if ok {
			t.Errorf("expected false for whitespace token")
		}

		// Raw 16-digit PAN (MUST BE REJECTED)
		_, ok = features.SanitizePaymentToken("4111111111111111")
		if ok {
			t.Errorf("CRITICAL SECURITY FLAW: raw 16-digit PAN was accepted!")
		}

		// Control characters & embedded null bytes
		_, ok = features.SanitizePaymentToken("tok_card\x00_inject")
		if ok {
			t.Errorf("expected false for embedded null byte")
		}

		// Oversized (>256 chars)
		oversized := string(make([]byte, 300))
		_, ok = features.SanitizePaymentToken(oversized)
		if ok {
			t.Errorf("expected false for oversized token")
		}
	})

	t.Run("7. Redis Key Privacy: No Plaintext Tokens in Keys", func(t *testing.T) {
		secretToken := "tok_secret_super_private_token"
		_ = tokenStore.RecordPaymentTokenTransaction(ctx, tenantA, deviceID_A, secretToken, "tx_sec", 1000)

		iter := client.Scan(ctx, 0, fmt.Sprintf("%s:*", tenantA), 0).Iterator()
		for iter.Next(ctx) {
			key := iter.Val()
			if containsSubstring(key, secretToken) {
				t.Fatalf("CRITICAL SECURITY FLAW: Plaintext token %q was found in Redis key %q", secretToken, key)
			}
		}
	})

	t.Run("8. Concurrency: 100 Simultaneous Token Updates", func(t *testing.T) {
		devConcur := "dev_tok_concur_100"
		var wg sync.WaitGroup
		numWorkers := 100

		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				tok := fmt.Sprintf("tok_worker_%02d", idx%10) // 10 distinct tokens
				_ = tokenStore.RecordPaymentTokenTransaction(ctx, tenantA, devConcur, tok, fmt.Sprintf("tx_w_%03d", idx), 100)
			}(i)
		}
		wg.Wait()

		f, err := tokenStore.GetPaymentTokenFeatures(ctx, tenantA, devConcur, "tok_worker_00")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if f.DeviceUniqueTokens24h != 10 {
			t.Errorf("expected 10 distinct tokens under concurrent load, got %d", f.DeviceUniqueTokens24h)
		}
		if f.DeviceTokenSeenBefore != 1 {
			t.Errorf("expected device_token_seen_before = 1, got %d", f.DeviceTokenSeenBefore)
		}
	})

	t.Run("9. Graceful Degradation on Disconnected Store", func(t *testing.T) {
		nilStore := features.NewPaymentTokenStore(nil)
		f, err := nilStore.GetPaymentTokenFeatures(ctx, tenantA, deviceID_A, "tok_1")
		if err != nil {
			t.Fatalf("expected nil error on degradation, got %v", err)
		}
		if !f.IsDegraded || f.DegradeReason != "PAYMENT_TOKEN_FEATURE_STORE_UNAVAILABLE" {
			t.Errorf("expected IsDegraded=true and DegradeReason=PAYMENT_TOKEN_FEATURE_STORE_UNAVAILABLE")
		}
		if f.DeviceUniqueTokens5m != 0 || f.DeviceTokenSeenBefore != 0 {
			t.Errorf("expected safe zero defaults on degraded payment token store")
		}
	})
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && (s[:len(substr)] == substr || containsSubstring(s[1:], substr))))
}
