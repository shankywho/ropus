package features_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/shankywho/ropus/backend/internal/features"
)

func TestDeviceReputationStore_Comprehensive(t *testing.T) {
	client := getTestRedisClient(t)
	if client == nil {
		return
	}
	defer client.Close()

	ctx := context.Background()
	repStore := features.NewDeviceReputationStore(client)

	tenantA := "tenant_rep_alpha"
	tenantB := "tenant_rep_beta"
	canonicalFP_A := "fp_test_macbook_rep_alpha"
	canonicalFP_B := "fp_test_macbook_rep_beta"
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

	t.Run("1. New Device Neutral Reputation & Point-in-Time Safety", func(t *testing.T) {
		// Query BEFORE recording
		f, err := repStore.GetReputationFeatures(ctx, tenantA, deviceID_A)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.DeviceReputationScore != 0.50 {
			t.Errorf("expected neutral reputation 0.50 for new device, got %f", f.DeviceReputationScore)
		}
		if f.DeviceTotalTransactions != 0 || f.DeviceDisputedTransactions != 0 || f.DeviceFraudTransactions != 0 {
			t.Errorf("expected 0 stats on new device, got total=%d, disp=%d, fraud=%d",
				f.DeviceTotalTransactions, f.DeviceDisputedTransactions, f.DeviceFraudTransactions)
		}
		if f.DeviceDaysSinceLastDispute != -1.0 || f.DeviceDaysSinceLastFraud != -1.0 {
			t.Errorf("expected -1.0 for never disputed/fraud, got disp=%f, fraud=%f",
				f.DeviceDaysSinceLastDispute, f.DeviceDaysSinceLastFraud)
		}
	})

	t.Run("2. Successful History Builds Trust", func(t *testing.T) {
		devTrusted := "dev_trusted_reputation_1"
		baseTime := time.Date(2026, 8, 21, 14, 0, 0, 0, time.UTC)

		// 10 successful transactions across 20 days
		for i := 1; i <= 10; i++ {
			txTime := baseTime.Add(-time.Duration(21-i) * 24 * time.Hour)
			_ = repStore.RecordOutcomeAtTime(ctx, tenantA, devTrusted, fmt.Sprintf("tx_succ_%d", i), features.OutcomeSuccess, txTime)
		}

		f, err := repStore.GetReputationFeaturesAtTime(ctx, tenantA, devTrusted, baseTime)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.DeviceTotalTransactions != 10 || f.DeviceSuccessfulTransactions != 10 {
			t.Errorf("expected 10 successful transactions, got total=%d, succ=%d", f.DeviceTotalTransactions, f.DeviceSuccessfulTransactions)
		}
		if f.DeviceSuccessRate != 1.0 {
			t.Errorf("expected success rate = 1.0, got %f", f.DeviceSuccessRate)
		}
		if f.DeviceReputationScore >= 0.20 {
			t.Errorf("expected trusted score <= 0.20, got %f", f.DeviceReputationScore)
		}
	})

	t.Run("3. Dispute & Confirmed Fraud Impact & Decay", func(t *testing.T) {
		devFraud := "dev_fraud_reputation_2"
		baseTime := time.Date(2026, 8, 21, 14, 0, 0, 0, time.UTC)

		// 1. Transaction 1: 45 days ago, SUCCESS
		_ = repStore.RecordOutcomeAtTime(ctx, tenantA, devFraud, "tx_f1", features.OutcomeSuccess, baseTime.Add(-45*24*time.Hour))
		// 2. Transaction 2: 10 days ago, DISPUTE
		_ = repStore.RecordOutcomeAtTime(ctx, tenantA, devFraud, "tx_f2", features.OutcomeDispute, baseTime.Add(-10*24*time.Hour))
		// 3. Transaction 3: 2 days ago, FRAUD
		_ = repStore.RecordOutcomeAtTime(ctx, tenantA, devFraud, "tx_f3", features.OutcomeFraud, baseTime.Add(-2*24*time.Hour))

		f, err := repStore.GetReputationFeaturesAtTime(ctx, tenantA, devFraud, baseTime)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.DeviceDisputedTransactions != 1 {
			t.Errorf("expected 1 dispute, got %d", f.DeviceDisputedTransactions)
		}
		if f.DeviceFraudTransactions != 1 {
			t.Errorf("expected 1 fraud, got %d", f.DeviceFraudTransactions)
		}
		if f.DeviceRecentFraudCount != 1 || f.DeviceRecentDisputeCount != 1 {
			t.Errorf("expected recent fraud=1, recent dispute=1, got fraud=%d, disp=%d",
				f.DeviceRecentFraudCount, f.DeviceRecentDisputeCount)
		}
		if f.DeviceReputationScore < 0.85 {
			t.Errorf("expected highly risky score >= 0.85, got %f", f.DeviceReputationScore)
		}
		if f.DeviceDaysSinceLastFraud != 2.0 {
			t.Errorf("expected DaysSinceLastFraud = 2.0, got %f", f.DeviceDaysSinceLastFraud)
		}
		if f.DeviceDaysSinceLastDispute != 10.0 {
			t.Errorf("expected DaysSinceLastDispute = 10.0, got %f", f.DeviceDaysSinceLastDispute)
		}
	})

	t.Run("4. Idempotency: Duplicate Dispute/Fraud Events Do Not Inflate Counts", func(t *testing.T) {
		devIdemp := "dev_idempotency_target"
		now := time.Now().UTC()

		// Record SAME dispute outcome 5 times
		for i := 0; i < 5; i++ {
			_ = repStore.RecordOutcomeAtTime(ctx, tenantA, devIdemp, "tx_same_dispute_99", features.OutcomeDispute, now)
		}

		f, err := repStore.GetReputationFeaturesAtTime(ctx, tenantA, devIdemp, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.DeviceDisputedTransactions != 1 {
			t.Errorf("expected exactly 1 dispute due to deduplication, got %d", f.DeviceDisputedTransactions)
		}
	})

	t.Run("5. Tenant Isolation: Complete Reputation Partitioning", func(t *testing.T) {
		// Record fraud in Tenant A
		_ = repStore.RecordOutcome(ctx, tenantA, deviceID_A, "tx_ta_fraud", features.OutcomeFraud)

		// Query under Tenant B
		fB, err := repStore.GetReputationFeatures(ctx, tenantB, deviceID_B)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if fB.DeviceFraudTransactions != 0 || fB.DeviceReputationScore != 0.50 {
			t.Errorf("CRITICAL SECURITY LEAK: Tenant B observed fraud history from Tenant A! (fraud=%d, score=%f)",
				fB.DeviceFraudTransactions, fB.DeviceReputationScore)
		}
	})

	t.Run("6. Concurrency: 100 Simultaneous Outcome Ingestions", func(t *testing.T) {
		devConcur := "dev_rep_concur_100"
		var wg sync.WaitGroup
		numWorkers := 100

		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				outcome := features.OutcomeSuccess
				if idx%10 == 0 {
					outcome = features.OutcomeDispute
				}
				_ = repStore.RecordOutcome(ctx, tenantA, devConcur, fmt.Sprintf("tx_w_%03d", idx), outcome)
			}(i)
		}
		wg.Wait()

		f, err := repStore.GetReputationFeatures(ctx, tenantA, devConcur)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.DeviceSuccessfulTransactions != 90 {
			t.Errorf("expected 90 successful transactions, got %d", f.DeviceSuccessfulTransactions)
		}
		if f.DeviceDisputedTransactions != 10 {
			t.Errorf("expected 10 disputes, got %d", f.DeviceDisputedTransactions)
		}
	})

	t.Run("7. Graceful Degradation on Disconnected Store", func(t *testing.T) {
		nilStore := features.NewDeviceReputationStore(nil)
		f, err := nilStore.GetReputationFeatures(ctx, tenantA, deviceID_A)
		if err != nil {
			t.Fatalf("expected nil error on degradation, got %v", err)
		}
		if !f.IsDegraded || f.DegradeReason != "DEVICE_REPUTATION_FEATURE_STORE_UNAVAILABLE" {
			t.Errorf("expected IsDegraded=true and DegradeReason=DEVICE_REPUTATION_FEATURE_STORE_UNAVAILABLE")
		}
		if f.DeviceReputationScore != 0.50 {
			t.Errorf("expected neutral 0.50 default on degraded store, got %f", f.DeviceReputationScore)
		}
	})
}
