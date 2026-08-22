package riskengine

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestCanaryRouter_0Percent(t *testing.T) {
	cfg := CanaryRouterConfig{
		Enabled:    false,
		Percentage: 0,
	}
	router := NewCanaryRouter(cfg, nil)

	for i := 0; i < 1000; i++ {
		tenantID := fmt.Sprintf("tenant_%d", i%5)
		txID := fmt.Sprintf("txn_%d", i)
		route := router.Route(tenantID, txID)
		if route != RouteLegacy {
			t.Fatalf("Expected RouteLegacy for 0%% canary at iteration %d, got %v", i, route)
		}
	}

	status := router.GetStatus()
	if status["enabled"] != false {
		t.Errorf("Expected enabled=false in status, got %v", status["enabled"])
	}
	if status["target_percentage"] != 0 {
		t.Errorf("Expected target_percentage=0, got %v", status["target_percentage"])
	}
}

func TestCanaryRouter_100Percent(t *testing.T) {
	cfg := CanaryRouterConfig{
		Enabled:    true,
		Percentage: 100,
	}
	router := NewCanaryRouter(cfg, nil)

	for i := 0; i < 1000; i++ {
		tenantID := fmt.Sprintf("tenant_%d", i%5)
		txID := fmt.Sprintf("txn_%d", i)
		route := router.Route(tenantID, txID)
		if route != RouteCandidate {
			t.Fatalf("Expected RouteCandidate for 100%% canary at iteration %d, got %v", i, route)
		}
	}
}

func TestCanaryRouter_StatisticalDistribution(t *testing.T) {
	stages := []struct {
		percentage int
		tolerance  float64 // acceptable absolute % error over 10k samples
	}{
		{1, 0.6},   // [0.4% - 1.6%]
		{5, 1.0},   // [4.0% - 6.0%]
		{10, 1.5},  // [8.5% - 11.5%]
		{25, 2.0},  // [23.0% - 27.0%]
		{50, 2.5},  // [47.5% - 52.5%]
	}

	sampleSize := 10000

	for _, stage := range stages {
		t.Run(fmt.Sprintf("Stage_%dPercent", stage.percentage), func(t *testing.T) {
			cfg := CanaryRouterConfig{
				Enabled:    true,
				Percentage: stage.percentage,
			}
			router := NewCanaryRouter(cfg, nil)

			candidateCount := 0
			for i := 0; i < sampleSize; i++ {
				tenantID := "tenant_prod_1"
				txID := fmt.Sprintf("tx_stat_sample_%d_%d", stage.percentage, i)
				if router.Route(tenantID, txID) == RouteCandidate {
					candidateCount++
				}
			}

			actualPct := (float64(candidateCount) / float64(sampleSize)) * 100.0
			delta := math.Abs(actualPct - float64(stage.percentage))

			if delta > stage.tolerance {
				t.Errorf("Target percentage %d%%: actual %.2f%% exceeded tolerance ±%.2f%% (delta=%.2f%%)",
					stage.percentage, actualPct, stage.tolerance, delta)
			}
		})
	}
}

func TestCanaryRouter_DeterministicIdempotency(t *testing.T) {
	cfg := CanaryRouterConfig{
		Enabled:    true,
		Percentage: 25,
	}
	router := NewCanaryRouter(cfg, nil)

	// Test that repeated calls for the exact same transaction return the exact same route
	testCases := []struct {
		tenantID string
		txID     string
	}{
		{"tenant_a", "txn_001"},
		{"tenant_a", "txn_002"},
		{"tenant_b", "txn_001"},
		{"tenant_b", "txn_999"},
		{"tenant_c", "txn_alpha_beta"},
	}

	for _, tc := range testCases {
		initialRoute := router.Route(tc.tenantID, tc.txID)
		for i := 0; i < 50; i++ {
			repeatRoute := router.Route(tc.tenantID, tc.txID)
			if repeatRoute != initialRoute {
				t.Fatalf("Non-deterministic routing for (%s, %s): initial=%v, iteration_%d=%v",
					tc.tenantID, tc.txID, initialRoute, i, repeatRoute)
			}
		}
	}
}

func TestCanaryRouter_CrossTransactionDispersion(t *testing.T) {
	cfg := CanaryRouterConfig{
		Enabled:    true,
		Percentage: 50,
	}
	router := NewCanaryRouter(cfg, nil)

	// Verify that buckets are well-dispersed and not all hashing to the same bucket
	bucketsHit := make(map[ModelRoute]int)
	for i := 0; i < 1000; i++ {
		r := router.Route("tenant_dispersion", fmt.Sprintf("txn_disp_%d", i))
		bucketsHit[r]++
	}

	if bucketsHit[RouteLegacy] == 0 || bucketsHit[RouteCandidate] == 0 {
		t.Errorf("Expected both legacy and candidate routes to be hit, got: %v", bucketsHit)
	}
}

func TestCanaryRouter_InvalidConfiguration(t *testing.T) {
	// Negative percentage should fail safe to disabled 0%
	cfgNeg := CanaryRouterConfig{
		Enabled:    true,
		Percentage: -15,
	}
	routerNeg := NewCanaryRouter(cfgNeg, nil)
	if routerNeg.Route("tenant_1", "txn_1") != RouteLegacy {
		t.Errorf("Negative percentage must route to RouteLegacy")
	}

	// Percentage > 100 should clamp to 100
	cfgOver := CanaryRouterConfig{
		Enabled:    true,
		Percentage: 150,
	}
	routerOver := NewCanaryRouter(cfgOver, nil)
	if routerOver.Route("tenant_1", "txn_1") != RouteCandidate {
		t.Errorf("Percentage > 100 must clamp to RouteCandidate")
	}
}

func TestCanaryRouter_CandidateSuccessAndMetrics(t *testing.T) {
	cfg := CanaryRouterConfig{
		Enabled:    true,
		Percentage: 10,
	}
	router := NewCanaryRouter(cfg, nil)

	// Simulate 100 legacy requests and 10 candidate requests
	for i := 0; i < 100; i++ {
		router.RecordLegacyRequest()
	}

	latencies := []float64{2.5, 3.1, 4.0, 5.2, 3.8, 4.5, 6.1, 2.9, 3.4, 4.8}
	for i, lat := range latencies {
		router.RecordCandidateRequest()
		decision := "ALLOW_RECOMMENDATION"
		if i == 8 {
			decision = "MANUAL_REVIEW"
		} else if i == 9 {
			decision = "DECLINE_RECOMMENDATION"
		}
		router.RecordCandidateSuccess(lat, decision)
	}

	status := router.GetStatus()
	metrics := status["metrics"].(map[string]interface{})

	if metrics["legacy_requests_total"].(int64) != 100 {
		t.Errorf("Expected 100 legacy requests, got %v", metrics["legacy_requests_total"])
	}
	if metrics["candidate_requests_total"].(int64) != 10 {
		t.Errorf("Expected 10 candidate requests, got %v", metrics["candidate_requests_total"])
	}
	if metrics["candidate_success_total"].(int64) != 10 {
		t.Errorf("Expected 10 candidate successes, got %v", metrics["candidate_success_total"])
	}
	if metrics["candidate_error_total"].(int64) != 0 {
		t.Errorf("Expected 0 candidate errors, got %v", metrics["candidate_error_total"])
	}
	if metrics["candidate_decision_allow_total"].(int64) != 8 {
		t.Errorf("Expected 8 candidate allows, got %v", metrics["candidate_decision_allow_total"])
	}
	if metrics["candidate_decision_review_total"].(int64) != 1 {
		t.Errorf("Expected 1 candidate review, got %v", metrics["candidate_decision_review_total"])
	}
	if metrics["candidate_decision_decline_total"].(int64) != 1 {
		t.Errorf("Expected 1 candidate decline, got %v", metrics["candidate_decision_decline_total"])
	}

	p50 := metrics["candidate_p50_latency_ms"].(float64)
	if p50 <= 0 || p50 > 10 {
		t.Errorf("Expected valid candidate p50 latency, got %v", p50)
	}
}

func TestCanaryRouter_CandidateFallbackAndMetrics(t *testing.T) {
	cfg := CanaryRouterConfig{
		Enabled:    true,
		Percentage: 10,
	}
	router := NewCanaryRouter(cfg, nil)

	// Simulate 5 successful candidate requests and 2 fallback failures
	for i := 0; i < 5; i++ {
		router.RecordCandidateRequest()
		router.RecordCandidateSuccess(3.0, "ALLOW_RECOMMENDATION")
	}

	for i := 0; i < 2; i++ {
		router.RecordCandidateRequest()
		router.RecordCandidateFallback(50.0, "timeout: context deadline exceeded")
	}

	status := router.GetStatus()
	metrics := status["metrics"].(map[string]interface{})

	if metrics["candidate_requests_total"].(int64) != 7 {
		t.Errorf("Expected 7 total candidate requests, got %v", metrics["candidate_requests_total"])
	}
	if metrics["candidate_fallback_total"].(int64) != 2 {
		t.Errorf("Expected 2 candidate fallbacks, got %v", metrics["candidate_fallback_total"])
	}
	if metrics["candidate_error_total"].(int64) != 2 {
		t.Errorf("Expected 2 candidate errors, got %v", metrics["candidate_error_total"])
	}

	fallbackRate := metrics["candidate_fallback_rate"].(float64)
	expectedRate := math.Round((2.0/7.0)*10000) / 10000
	if math.Abs(fallbackRate-expectedRate) > 0.001 {
		t.Errorf("Expected fallback rate ~%.4f, got %.4f", expectedRate, fallbackRate)
	}
}

func TestCanaryRouter_SafetyGates(t *testing.T) {
	t.Run("Healthy Status Under Normal Traffic", func(t *testing.T) {
		cfg := CanaryRouterConfig{
			Enabled:         true,
			Percentage:      10,
			MaxErrorRate:    0.05,
			MaxFallbackRate: 0.05,
			MaxP95LatencyMs: 20.0,
		}
		router := NewCanaryRouter(cfg, nil)

		for i := 0; i < 50; i++ {
			router.RecordCandidateRequest()
			router.RecordCandidateSuccess(4.0, "ALLOW_RECOMMENDATION")
		}

		safety := router.EvaluateSafetyGates()
		if safety.Status != GateStatusHealthy {
			t.Errorf("Expected HEALTHY status, got %v (violations: %v)", safety.Status, safety.Violations)
		}
		if len(safety.Violations) > 0 {
			t.Errorf("Expected 0 violations, got %v", safety.Violations)
		}
	})

	t.Run("Failed Status On High Error Rate", func(t *testing.T) {
		cfg := CanaryRouterConfig{
			Enabled:         true,
			Percentage:      10,
			MaxErrorRate:    0.01,
			MaxFallbackRate: 0.01,
		}
		router := NewCanaryRouter(cfg, nil)

		// 15 requests with 3 fallbacks (20% fallback rate > 1% threshold)
		for i := 0; i < 12; i++ {
			router.RecordCandidateRequest()
			router.RecordCandidateSuccess(3.0, "ALLOW_RECOMMENDATION")
		}
		for i := 0; i < 3; i++ {
			router.RecordCandidateRequest()
			router.RecordCandidateFallback(45.0, "connection refused")
		}

		safety := router.EvaluateSafetyGates()
		if safety.Status != GateStatusFailed {
			t.Errorf("Expected FAILED status on high fallback rate, got %v", safety.Status)
		}
		if len(safety.Violations) == 0 {
			t.Errorf("Expected violations to be populated, got none")
		}
	})

	t.Run("Idle Status When Disabled", func(t *testing.T) {
		cfg := CanaryRouterConfig{
			Enabled:    false,
			Percentage: 0,
		}
		router := NewCanaryRouter(cfg, nil)
		safety := router.EvaluateSafetyGates()
		if safety.Status != GateStatusIdle {
			t.Errorf("Expected IDLE status for disabled canary, got %v", safety.Status)
		}
	})
}

func TestCanaryRouter_ConcurrencyAndRace(t *testing.T) {
	cfg := CanaryRouterConfig{
		Enabled:    true,
		Percentage: 25,
	}
	router := NewCanaryRouter(cfg, nil)

	var wg sync.WaitGroup
	workers := 20
	requestsPerWorker := 200

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)))

			for r := 0; r < requestsPerWorker; r++ {
				tenantID := fmt.Sprintf("tenant_conc_%d", workerID%3)
				txID := fmt.Sprintf("txn_conc_%d_%d", workerID, r)

				route := router.Route(tenantID, txID)
				if route == RouteCandidate {
					router.RecordCandidateRequest()
					lat := float64(rng.Intn(10)+1) + rng.Float64()
					if rng.Float64() < 0.05 {
						router.RecordCandidateFallback(lat, "simulated error")
					} else {
						router.RecordCandidateSuccess(lat, "ALLOW_RECOMMENDATION")
					}
				} else {
					router.RecordLegacyRequest()
				}

				if r%50 == 0 {
					_ = router.GetStatus()
					_ = router.EvaluateSafetyGates()
				}
			}
		}(w)
	}

	wg.Wait()

	status := router.GetStatus()
	metrics := status["metrics"].(map[string]interface{})
	total := metrics["total_requests"].(int64)
	expectedTotal := int64(workers * requestsPerWorker)

	if total != expectedTotal {
		t.Errorf("Expected %d total requests, got %d", expectedTotal, total)
	}
}

func BenchmarkCanaryRouter_Route(b *testing.B) {
	cfg := CanaryRouterConfig{
		Enabled:    true,
		Percentage: 25,
	}
	router := NewCanaryRouter(cfg, nil)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = router.Route("tenant_bench", "txn_bench_123456")
	}
}

func BenchmarkCanaryRouter_ReservoirPercentiles(b *testing.B) {
	res := NewLatencyReservoir(2000)
	for i := 0; i < 2000; i++ {
		res.Add(float64(i%20) + 0.5)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, _ = res.Percentiles()
	}
}

