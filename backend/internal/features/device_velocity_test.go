package features_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/shankywho/ropus/backend/internal/features"
)

func TestDeviceVelocityStore_Comprehensive(t *testing.T) {
	client := getTestRedisClient(t)
	if client == nil {
		return
	}
	defer client.Close()

	ctx := context.Background()
	velStore := features.NewDeviceVelocityStore(client)

	tenantA := "tenant_vel_alpha"
	tenantB := "tenant_vel_beta"
	canonicalFP_A := "fp_test_macbook_vel_alpha"
	canonicalFP_B := "fp_test_macbook_vel_beta"
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

	t.Run("1. First Transaction & Point-in-Time Zero Velocity", func(t *testing.T) {
		// Point-in-Time: Query BEFORE recording
		fBefore, err := velStore.GetVelocityFeatures(ctx, tenantA, deviceID_A)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if fBefore.DeviceTxCount10s != 0 || fBefore.DeviceTxCount1m != 0 || fBefore.DeviceTxCount24h != 0 {
			t.Errorf("expected 0 transactions before first recording, got 10s=%d, 1m=%d, 24h=%d",
				fBefore.DeviceTxCount10s, fBefore.DeviceTxCount1m, fBefore.DeviceTxCount24h)
		}
		if fBefore.DeviceAmountSum24h != 0 {
			t.Errorf("expected 0 amount sum, got %d", fBefore.DeviceAmountSum24h)
		}

		// Record first transaction
		_ = velStore.RecordVelocityTransaction(ctx, tenantA, deviceID_A, "tx_001", 5000)

		// Query AFTER recording
		fAfter, err := velStore.GetVelocityFeatures(ctx, tenantA, deviceID_A)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if fAfter.DeviceTxCount10s != 1 || fAfter.DeviceTxCount1m != 1 || fAfter.DeviceTxCount24h != 1 {
			t.Errorf("expected 1 transaction across windows after recording, got 10s=%d, 1m=%d, 24h=%d",
				fAfter.DeviceTxCount10s, fAfter.DeviceTxCount1m, fAfter.DeviceTxCount24h)
		}
		if fAfter.DeviceAmountSum10s != 5000 || fAfter.DeviceAmountSum24h != 5000 {
			t.Errorf("expected amount sum 5000, got 10s=%d, 24h=%d", fAfter.DeviceAmountSum10s, fAfter.DeviceAmountSum24h)
		}
		if fAfter.DeviceAvgAmount1m != 5000.0 || fAfter.DeviceMaxAmount24h != 5000 {
			t.Errorf("expected avg 5000 and max 5000, got avg=%f, max=%d", fAfter.DeviceAvgAmount1m, fAfter.DeviceMaxAmount24h)
		}
	})

	t.Run("2. Multi-Window Rolling Velocity & Time Injection", func(t *testing.T) {
		devMulti := "dev_vel_multi_window_target"
		baseTime := time.Date(2026, 8, 21, 14, 0, 0, 0, time.UTC)

		// Event 1: 18 hours ago, amount 10000
		_ = velStore.RecordVelocityTransactionAtTime(ctx, tenantA, devMulti, "tx_1", 10000, baseTime.Add(-18*time.Hour))
		// Event 2: 4 hours ago, amount 8000
		_ = velStore.RecordVelocityTransactionAtTime(ctx, tenantA, devMulti, "tx_2", 8000, baseTime.Add(-4*time.Hour))
		// Event 3: 45 minutes ago, amount 6000
		_ = velStore.RecordVelocityTransactionAtTime(ctx, tenantA, devMulti, "tx_3", 6000, baseTime.Add(-45*time.Minute))
		// Event 4: 10 minutes ago, amount 4000
		_ = velStore.RecordVelocityTransactionAtTime(ctx, tenantA, devMulti, "tx_4", 4000, baseTime.Add(-10*time.Minute))
		// Event 5: 3 minutes ago, amount 2000
		_ = velStore.RecordVelocityTransactionAtTime(ctx, tenantA, devMulti, "tx_5", 2000, baseTime.Add(-3*time.Minute))
		// Event 6: 30 seconds ago, amount 1500
		_ = velStore.RecordVelocityTransactionAtTime(ctx, tenantA, devMulti, "tx_6", 1500, baseTime.Add(-30*time.Second))
		// Event 7: 5 seconds ago, amount 500
		_ = velStore.RecordVelocityTransactionAtTime(ctx, tenantA, devMulti, "tx_7", 500, baseTime.Add(-5*time.Second))

		// Query at reference time
		f, err := velStore.GetVelocityFeaturesAtTime(ctx, tenantA, devMulti, baseTime)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify counts
		if f.DeviceTxCount10s != 1 {
			t.Errorf("expected 10s count = 1, got %d", f.DeviceTxCount10s)
		}
		if f.DeviceTxCount1m != 2 {
			t.Errorf("expected 1m count = 2 (5s, 30s), got %d", f.DeviceTxCount1m)
		}
		if f.DeviceTxCount5m != 3 {
			t.Errorf("expected 5m count = 3 (5s, 30s, 3m), got %d", f.DeviceTxCount5m)
		}
		if f.DeviceTxCount15m != 4 {
			t.Errorf("expected 15m count = 4, got %d", f.DeviceTxCount15m)
		}
		if f.DeviceTxCount1h != 5 {
			t.Errorf("expected 1h count = 5, got %d", f.DeviceTxCount1h)
		}
		if f.DeviceTxCount6h != 6 {
			t.Errorf("expected 6h count = 6, got %d", f.DeviceTxCount6h)
		}
		if f.DeviceTxCount24h != 7 {
			t.Errorf("expected 24h count = 7, got %d", f.DeviceTxCount24h)
		}

		// Verify sums
		if f.DeviceAmountSum10s != 500 {
			t.Errorf("expected 10s sum = 500, got %d", f.DeviceAmountSum10s)
		}
		if f.DeviceAmountSum1m != 2000 {
			t.Errorf("expected 1m sum = 2000 (500+1500), got %d", f.DeviceAmountSum1m)
		}
		if f.DeviceAmountSum5m != 4000 {
			t.Errorf("expected 5m sum = 4000 (500+1500+2000), got %d", f.DeviceAmountSum5m)
		}
		if f.DeviceAmountSum1h != 14000 {
			t.Errorf("expected 1h sum = 14000, got %d", f.DeviceAmountSum1h)
		}
		if f.DeviceAmountSum24h != 32000 {
			t.Errorf("expected 24h sum = 32000, got %d", f.DeviceAmountSum24h)
		}
		if f.DeviceMaxAmount24h != 10000 {
			t.Errorf("expected 24h max amount = 10000, got %d", f.DeviceMaxAmount24h)
		}
	})

	t.Run("3. Velocity Acceleration, Rates & Concentration", func(t *testing.T) {
		devBurst := "dev_vel_burst_accel_target"
		now := time.Now().UTC()

		// Generate a 1-minute burst: 10 transactions in 1 minute, but only 10 in 15 minutes
		for i := 0; i < 10; i++ {
			_ = velStore.RecordVelocityTransactionAtTime(ctx, tenantA, devBurst, fmt.Sprintf("tx_b_%d", i), 1000, now.Add(-time.Duration(i*5)*time.Second))
		}

		f, err := velStore.GetVelocityFeaturesAtTime(ctx, tenantA, devBurst, now)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Acceleration: 1m count (10) vs 15m count (10/15 = 0.66) => ~15.0 acceleration
		if f.TxAcceleration1m15m < 10.0 {
			t.Errorf("expected strong TxAcceleration1m15m >= 10.0, got %f", f.TxAcceleration1m15m)
		}
		if f.DeviceAmountConcentration5m1h != 1.0 {
			t.Errorf("expected DeviceAmountConcentration5m1h = 1.0, got %f", f.DeviceAmountConcentration5m1h)
		}
		if f.VelocitySignal != features.VelocitySuspicious && f.VelocitySignal != features.VelocityHighSignal {
			t.Errorf("expected SUSPICIOUS or HIGH_SIGNAL velocity, got %s", f.VelocitySignal)
		}
	})

	t.Run("4. Tenant Isolation: Independent Velocity Ledgers", func(t *testing.T) {
		// Tenant A has 10 transactions
		for i := 0; i < 10; i++ {
			_ = velStore.RecordVelocityTransaction(ctx, tenantA, deviceID_A, fmt.Sprintf("tx_ta_%d", i), 1000)
		}

		// Query under Tenant B with deviceID_B
		fB, err := velStore.GetVelocityFeatures(ctx, tenantB, deviceID_B)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if fB.DeviceTxCount24h != 0 || fB.DeviceAmountSum24h != 0 {
			t.Errorf("CRITICAL SECURITY FLAW: Tenant B observed velocity leakage from Tenant A! (count=%d, sum=%d)",
				fB.DeviceTxCount24h, fB.DeviceAmountSum24h)
		}
	})

	t.Run("5. Concurrency: 100 Simultaneous Velocity Updates", func(t *testing.T) {
		devConcur := "dev_vel_concur_100"
		var wg sync.WaitGroup
		numWorkers := 100

		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				_ = velStore.RecordVelocityTransaction(ctx, tenantA, devConcur, fmt.Sprintf("tx_w_%03d", idx), 500)
			}(i)
		}
		wg.Wait()

		f, err := velStore.GetVelocityFeatures(ctx, tenantA, devConcur)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if f.DeviceTxCount1m != 100 || f.DeviceTxCount24h != 100 {
			t.Errorf("expected 100 transactions under concurrency, got 1m=%d, 24h=%d", f.DeviceTxCount1m, f.DeviceTxCount24h)
		}
		if f.DeviceAmountSum24h != 50000 {
			t.Errorf("expected sum = 50000, got %d", f.DeviceAmountSum24h)
		}
	})

	t.Run("6. Graceful Degradation on Disconnected Store", func(t *testing.T) {
		nilStore := features.NewDeviceVelocityStore(nil)
		f, err := nilStore.GetVelocityFeatures(ctx, tenantA, deviceID_A)
		if err != nil {
			t.Fatalf("expected nil error on degradation, got %v", err)
		}
		if !f.IsDegraded || f.DegradeReason != "VELOCITY_FEATURE_STORE_UNAVAILABLE" {
			t.Errorf("expected IsDegraded=true and DegradeReason=VELOCITY_FEATURE_STORE_UNAVAILABLE")
		}
		if f.DeviceTxCount1m != 0 || f.DeviceAmountSum24h != 0 {
			t.Errorf("expected safe zero defaults on degraded store")
		}
	})
}
