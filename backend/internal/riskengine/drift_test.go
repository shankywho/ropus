package riskengine

import (
	"context"
	"math"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestDriftCalculator_PSI_IdenticalAndShifted(t *testing.T) {
	// Identical uniform distributions
	a := []float64{0.25, 0.25, 0.25, 0.25}
	b := []float64{0.25, 0.25, 0.25, 0.25}
	psiIdentical := CalculatePSI(a, b, 1e-6)
	if psiIdentical > 0.001 {
		t.Fatalf("expected PSI for identical distributions to be ~0, got %f", psiIdentical)
	}

	// Moderately shifted distribution
	shifted := []float64{0.40, 0.30, 0.20, 0.10}
	psiModerate := CalculatePSI(shifted, b, 1e-6)
	if psiModerate <= 0.01 {
		t.Fatalf("expected moderate PSI to be > 0.01, got %f", psiModerate)
	}

	// Severely shifted distribution
	severe := []float64{0.90, 0.05, 0.03, 0.02}
	psiSevere := CalculatePSI(severe, b, 1e-6)
	if psiSevere <= psiModerate {
		t.Fatalf("expected severe PSI (%f) to be > moderate PSI (%f)", psiSevere, psiModerate)
	}

	// Mismatched lengths
	if psiMismatch := CalculatePSI(a, []float64{0.5, 0.5}, 1e-6); psiMismatch != 0.0 {
		t.Fatalf("expected 0.0 on mismatched lengths, got %f", psiMismatch)
	}

	// Empty distributions
	if psiEmpty := CalculatePSI(nil, nil, 1e-6); psiEmpty != 0.0 {
		t.Fatalf("expected 0.0 on empty distributions, got %f", psiEmpty)
	}
}

func TestDriftCalculator_KLDivergence(t *testing.T) {
	a := []float64{0.25, 0.25, 0.25, 0.25}
	b := []float64{0.25, 0.25, 0.25, 0.25}
	klIdentical := CalculateKLDivergence(a, b, 1e-6)
	if klIdentical > 0.001 {
		t.Fatalf("expected KL for identical distributions to be ~0, got %f", klIdentical)
	}

	shifted := []float64{0.60, 0.20, 0.10, 0.10}
	klShifted := CalculateKLDivergence(shifted, b, 1e-6)
	if klShifted <= 0.05 {
		t.Fatalf("expected shifted KL to be > 0.05, got %f", klShifted)
	}
}

func TestDriftCalculator_JSDivergence(t *testing.T) {
	a := []float64{0.25, 0.25, 0.25, 0.25}
	b := []float64{0.25, 0.25, 0.25, 0.25}
	jsdIdentical := CalculateJSDivergence(a, b)
	if jsdIdentical > 0.001 {
		t.Fatalf("expected JSD for identical distributions to be ~0, got %f", jsdIdentical)
	}

	// Symmetric property: JSD(P, Q) == JSD(Q, P)
	p := []float64{0.7, 0.1, 0.1, 0.1}
	q := []float64{0.1, 0.2, 0.3, 0.4}
	jsdPQ := CalculateJSDivergence(p, q)
	jsdQP := CalculateJSDivergence(q, p)
	if math.Abs(jsdPQ-jsdQP) > 1e-4 {
		t.Fatalf("expected symmetric JSD, got JSD(P,Q)=%f != JSD(Q,P)=%f", jsdPQ, jsdQP)
	}

	// Boundedness: JSD must be in [0, ln(2)] ≈ [0, 0.69315]
	if jsdPQ < 0 || jsdPQ > 0.70 {
		t.Fatalf("expected JSD to be bounded in [0, 0.69315], got %f", jsdPQ)
	}
}

func TestDriftCalculator_ZeroBinsAndEpsilon(t *testing.T) {
	// One distribution has 0 count in some bins
	actualWithZeros := []float64{1.0, 0.0, 0.0, 0.0}
	expectedUniform := []float64{0.25, 0.25, 0.25, 0.25}

	psi := CalculatePSI(actualWithZeros, expectedUniform, 1e-6)
	if math.IsNaN(psi) || math.IsInf(psi, 0) {
		t.Fatalf("PSI calculation produced NaN/Inf with zero bins: %f", psi)
	}
	if psi <= 0.5 {
		t.Fatalf("expected high PSI for extreme delta with zero bins, got %f", psi)
	}

	jsd := CalculateJSDivergence(actualWithZeros, expectedUniform)
	if math.IsNaN(jsd) || math.IsInf(jsd, 0) {
		t.Fatalf("JSD calculation produced NaN/Inf with zero bins: %f", jsd)
	}
}

func TestDriftCalculator_BinContinuousFeature(t *testing.T) {
	edges := []float64{0.0, 100.0, 200.0, 300.0}
	values := []float64{-50.0, 50.0, 150.0, 250.0, 500.0, math.NaN()}

	probs := BinContinuousFeature(values, edges)
	if len(probs) != 3 {
		t.Fatalf("expected 3 bins, got %d", len(probs))
	}

	// -50.0 goes to bin 0, 50.0 to bin 0, 150.0 to bin 1, 250.0 to bin 2, 500.0 to bin 2
	// Total valid values = 5
	// Bin 0 count = 2 -> 2/5 = 0.4
	// Bin 1 count = 1 -> 1/5 = 0.2
	// Bin 2 count = 2 -> 2/5 = 0.4
	if math.Abs(probs[0]-0.4) > 1e-4 || math.Abs(probs[1]-0.2) > 1e-4 || math.Abs(probs[2]-0.4) > 1e-4 {
		t.Fatalf("unexpected bin probabilities: %v", probs)
	}
}

func TestDriftCalculator_BinCategoricalFeature(t *testing.T) {
	categories := []string{"0", "1", "2"}
	baseProbs := map[string]float64{"0": 0.6, "1": 0.3, "2": 0.1}

	// Live values with known and unseen categories
	liveVals := []float64{0, 0, 1, 99} // 99 is unseen
	liveProbs, expProbs, unseenRate := BinCategoricalFeature(liveVals, categories, baseProbs)

	if len(liveProbs) != 4 || len(expProbs) != 4 {
		t.Fatalf("expected 4 buckets (3 known + 1 OTHER), got %d / %d", len(liveProbs), len(expProbs))
	}

	// 1 unseen out of 4 -> unseenRate = 0.25
	if math.Abs(unseenRate-0.25) > 1e-4 {
		t.Fatalf("expected unseenRate 0.25, got %f", unseenRate)
	}

	// Check PSI calculation
	psi := CalculatePSI(liveProbs, expProbs, 1e-6)
	if psi < 0 {
		t.Fatalf("negative PSI from categorical binning: %f", psi)
	}
}

func TestDriftCalculator_DescriptiveStatsAndPercentiles(t *testing.T) {
	cfg := DefaultDriftConfig()
	baseline := FeatureDistribution{
		Name:     "amount",
		DataType: "float64",
		Mean:     100.0,
		Std:      50.0,
		BinEdges: []float64{0.0, 50.0, 100.0, 150.0, 200.0},
		BinProbs: []float64{0.25, 0.25, 0.25, 0.25},
	}

	// Generate 100 synthetic observations around mean=100
	vals := make([]float64, 100)
	for i := 0; i < 100; i++ {
		vals[i] = float64(i + 1) * 2.0 // 2 to 200, mean ≈ 101, min=2, max=200
	}
	vals = append(vals, math.NaN()) // 1 missing

	res := CalculateFeatureDrift(vals, baseline, cfg)
	if res.FeatureName != "amount" {
		t.Fatalf("expected feature name 'amount', got %s", res.FeatureName)
	}
	if res.SampleCount != 101 {
		t.Fatalf("expected sample count 101, got %d", res.SampleCount)
	}
	if res.MissingRate <= 0.0 {
		t.Fatalf("expected non-zero missing rate due to NaN, got %f", res.MissingRate)
	}
	if res.LiveMin != 2.0 || res.LiveMax != 200.0 {
		t.Fatalf("expected min=2.0 max=200.0, got min=%f max=%f", res.LiveMin, res.LiveMax)
	}
	if res.LiveP50 < 90.0 || res.LiveP50 > 115.0 {
		t.Fatalf("expected p50 around 100, got %f", res.LiveP50)
	}
	if res.LiveP95 < 180.0 {
		t.Fatalf("expected p95 > 180, got %f", res.LiveP95)
	}
}

func TestDriftCollector_RingBufferAndRollingWindows(t *testing.T) {
	collector := NewDriftCollector(100, []string{"amount", "ip_velocity_1h"})

	// Push 150 values (exceeding capacity 100)
	for i := 1; i <= 150; i++ {
		collector.PushVector(map[string]float64{
			"amount":         float64(i),
			"ip_velocity_1h": float64(i % 10),
		})
	}

	if total := collector.TotalCollected(); total != 150 {
		t.Fatalf("expected total collected 150, got %d", total)
	}

	// Snapshot with window 50 (should get values 101 to 150)
	snap50, count50 := collector.Snapshot(50)
	if count50 != 50 || len(snap50["amount"]) != 50 {
		t.Fatalf("expected 50 samples, got %d", count50)
	}
	if snap50["amount"][0] != 101 || snap50["amount"][49] != 150 {
		t.Fatalf("unexpected window contents: first=%f, last=%f", snap50["amount"][0], snap50["amount"][49])
	}

	// Snapshot with window 100 (full ring buffer)
	snap100, count100 := collector.Snapshot(100)
	if count100 != 100 || len(snap100["amount"]) != 100 {
		t.Fatalf("expected 100 samples, got %d", count100)
	}
	if snap100["amount"][0] != 51 || snap100["amount"][99] != 150 {
		t.Fatalf("unexpected full buffer contents: first=%f, last=%f", snap100["amount"][0], snap100["amount"][99])
	}
}

func TestDriftDetector_LifecycleAndStatus(t *testing.T) {
	cfg := DriftConfig{
		PSIWarnThreshold:    0.10,
		PSIHighThreshold:    0.20,
		PSICritThreshold:    0.30,
		MinSamplesForDrift:  10,
		MaxWindowSize:       500,
		CalculationInterval: 100 * time.Millisecond,
		Epsilon:             1e-6,
	}

	detector, err := NewDriftDetector(cfg, nil)
	if err != nil {
		t.Fatalf("failed to initialize drift detector: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go detector.Start(ctx)

	// Status before samples
	initialStatus := detector.GetStatus()
	if initialStatus["status"] != DriftStatusHealthy {
		t.Fatalf("expected initial status HEALTHY, got %v", initialStatus["status"])
	}
	if initialStatus["baseline_source"] != "development_fixture" {
		t.Fatalf("expected baseline_source 'development_fixture', got %v", initialStatus["baseline_source"])
	}

	// Ingest 20 normal vectors matching baseline means
	for i := 0; i < 20; i++ {
		detector.IngestVector(map[string]float64{
			"amount":                 120.0,
			"ip_velocity_1h":         1.0,
			"ip_velocity_24h":        8.0,
			"token_velocity_24h":     1.0,
			"device_seen_before":     0.0,
			"transaction_hour":       12.0,
			"transaction_day":        2.0,
			"product_cd_encoded":     0.0,
			"card_type_encoded":      1.0,
			"card_category_encoded":  0.0,
			"email_domain_risk":      0.03,
			"dist1_missing":          1.0,
			"device_type_mobile":     0.0,
			"device_info_missing":    0.0,
			"amount_to_mean_ratio":   1.0,
			"device_tx_count_5m":     0.0,
			"device_tx_count_1h":     0.0,
			"device_amount_sum_24h":  0.0,
			"tx_acceleration_5m_1h":  0.0,
			"device_amount_concentration_5m_1h": 0.0,
			"device_unique_tokens_1h": 1.0,
			"token_unique_devices_1h": 1.0,
			"device_reputation_score": 0.5,
			"device_fraud_rate":       0.0,
			"device_dispute_rate":     0.0,
		})
	}

	meas := detector.EvaluateLiveDrift(context.Background())
	if meas == nil {
		t.Fatal("expected non-nil drift measurement")
	}
	if meas.SampleCount != 20 {
		t.Fatalf("expected sample count 20, got %d", meas.SampleCount)
	}

	// Check Summary
	summary := detector.GetSummary()
	if summary["status"] == "" {
		t.Fatalf("expected non-empty summary status, got %v", summary)
	}

	// Check History
	history := detector.GetHistory()
	if len(history) == 0 {
		t.Fatal("expected history to record measurement")
	}

	detector.Stop()
}

func TestDriftDetector_SyntheticDriftAndRecovery(t *testing.T) {
	cfg := DefaultDriftConfig()
	cfg.MinSamplesForDrift = 20
	cfg.MaxWindowSize = 200
	cfg.CalculationInterval = 1 * time.Hour

	detector, err := NewDriftDetector(cfg, nil)
	if err != nil {
		t.Fatalf("failed to initialize drift detector: %v", err)
	}

	// 1. INGEST SYNTHETIC SHIFTED DATA: amount = $3000 (extreme shift from baseline ~$123)
	for i := 0; i < 50; i++ {
		detector.IngestVector(map[string]float64{
			"amount":                 3000.0,
			"ip_velocity_1h":         10.0, // baseline max ~5
			"ip_velocity_24h":        50.0, // baseline max ~19
			"token_velocity_24h":     5.0,
			"device_seen_before":     1.0,
			"transaction_hour":       3.0,
			"transaction_day":        6.0,
			"product_cd_encoded":     9.0,
			"card_type_encoded":      4.0,
			"card_category_encoded":  3.0,
			"email_domain_risk":      0.95,
			"dist1_missing":          0.0,
			"device_type_mobile":     1.0,
			"device_info_missing":    1.0,
			"amount_to_mean_ratio":   10.0,
			"device_tx_count_5m":     50.0,
			"device_tx_count_1h":     100.0,
			"device_amount_sum_24h":  10000.0,
			"tx_acceleration_5m_1h":  5.0,
			"device_amount_concentration_5m_1h": 0.8,
			"device_unique_tokens_1h": 10.0,
			"token_unique_devices_1h": 10.0,
			"device_reputation_score": 0.99,
			"device_fraud_rate":       0.9,
			"device_dispute_rate":     0.8,
		})
	}

	driftedMeas := detector.EvaluateLiveDrift(context.Background())
	if driftedMeas.OverallStatus != DriftStatusCritical && driftedMeas.OverallStatus != DriftStatusDegraded {
		t.Fatalf("expected status CRITICAL or DEGRADED on extreme synthetic shift, got %s (max PSI: %f)",
			driftedMeas.OverallStatus, driftedMeas.MaxPSI)
	}
	if driftedMeas.MaxPSI < 0.20 {
		t.Fatalf("expected max PSI > 0.20 on extreme shift, got %f", driftedMeas.MaxPSI)
	}

	// 2. INGEST RECOVERY DATA: Fill buffer with baseline-aligned samples sampled across baseline bins
	for i := 0; i < 200; i++ {
		sampleVec := make(map[string]float64, len(detector.baseline.Features))
		r := (float64(i) + 0.5) / 200.0
		for name, feat := range detector.baseline.Features {
			if len(feat.BinEdges) >= 2 && len(feat.BinProbs) > 0 {
				binIdx := 0
				cum := 0.0
				for bIdx, p := range feat.BinProbs {
					cum += p
					if r <= cum || bIdx == len(feat.BinProbs)-1 {
						binIdx = bIdx
						break
					}
				}
				mid := (feat.BinEdges[binIdx] + feat.BinEdges[binIdx+1]) / 2.0
				sampleVec[name] = mid
			} else {
				sampleVec[name] = feat.Mean
			}
		}
		detector.IngestVector(sampleVec)
	}

	recoveredMeas := detector.EvaluateLiveDrift(context.Background())
	if recoveredMeas.OverallStatus != DriftStatusHealthy {
		t.Fatalf("expected recovered status HEALTHY, got %s (max PSI: %f)",
			recoveredMeas.OverallStatus, recoveredMeas.MaxPSI)
	}
}

func TestDriftDetector_ConcurrentIngestAndEvaluate(t *testing.T) {
	cfg := DriftConfig{
		PSIWarnThreshold:    0.10,
		PSIHighThreshold:    0.20,
		PSICritThreshold:    0.30,
		MinSamplesForDrift:  10,
		MaxWindowSize:       1000,
		CalculationInterval: 5 * time.Millisecond,
		Epsilon:             1e-6,
	}

	detector, err := NewDriftDetector(cfg, nil)
	if err != nil {
		t.Fatalf("failed to initialize drift detector: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	go detector.Start(ctx)

	var wg sync.WaitGroup
	// 10 concurrent ingestion workers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)))
			for j := 0; j < 100; j++ {
				detector.IngestVector(map[string]float64{
					"amount":         50.0 + r.Float64()*100.0,
					"ip_velocity_1h": r.Float64() * 3.0,
				})
				_ = detector.GetStatus()
				_ = detector.GetSummary()
				_ = detector.GetHistory()
			}
		}(i)
	}

	wg.Wait()
	detector.Stop()
}

func TestDriftDetector_FailureIsolation(t *testing.T) {
	// Nil detector should not panic
	var nilDetector *DriftDetector
	nilDetector.IngestVector(map[string]float64{"amount": 100.0})

	status := nilDetector.GetStatus()
	if status["status"] != "UNAVAILABLE" {
		t.Fatalf("expected UNAVAILABLE status for nil detector, got %v", status["status"])
	}

	summary := nilDetector.GetSummary()
	if summary["status"] != "DEGRADED" {
		t.Fatalf("expected DEGRADED summary for nil detector, got %v", summary["status"])
	}

	history := nilDetector.GetHistory()
	if len(history) != 0 {
		t.Fatalf("expected empty history for nil detector, got %d", len(history))
	}
}
