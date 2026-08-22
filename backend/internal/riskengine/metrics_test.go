package riskengine

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsEngine_RecordAndSnapshot(t *testing.T) {
	metrics := NewMetricsEngine()
	require.NotNil(t, metrics)

	// Record sample requests
	metrics.RecordRequest(8.5, "ALLOW_RECOMMENDATION", 25, true, false, false)
	metrics.RecordRequest(15.2, "MANUAL_REVIEW", 70, true, false, false)
	metrics.RecordRequest(45.0, "REJECT_RECOMMENDATION", 95, false, true, true)

	metrics.IncrementDriftEvaluations()
	metrics.IncrementRetrainingJobs(false)
	metrics.IncrementRetrainingJobs(true)
	metrics.IncrementCanaryRollbacks()
	metrics.IncrementModelPromotions()
	metrics.IncrementCircuitBreakerTrips()
	metrics.IncrementDependencyFailures()

	snap := metrics.GetSnapshot()
	assert.Equal(t, int64(3), snap["requests_total"])
	assert.Equal(t, int64(2), snap["requests_success"])
	assert.Equal(t, int64(1), snap["requests_failed"])
	assert.Equal(t, int64(1), snap["inference_errors_total"])
	assert.Equal(t, int64(1), snap["fallbacks_total"])

	decisions := snap["decisions"].(map[string]int64)
	assert.Equal(t, int64(1), decisions["allow"])
	assert.Equal(t, int64(1), decisions["review"])
	assert.Equal(t, int64(1), decisions["reject"])

	latencies := snap["latency_percentiles_ms"].(map[string]float64)
	assert.True(t, latencies["p50"] > 0)
	assert.True(t, latencies["p95"] > 0)

	assert.Equal(t, int64(1), snap["drift_evaluations_total"])
	assert.Equal(t, int64(2), snap["retraining_jobs_total"])
	assert.Equal(t, int64(1), snap["retraining_failures_total"])
	assert.Equal(t, int64(1), snap["canary_rollbacks_total"])
	assert.Equal(t, int64(1), snap["model_promotions_total"])
	assert.Equal(t, int64(1), snap["circuit_breaker_trips"])
	assert.Equal(t, int64(1), snap["dependency_failures_total"])
}

func TestMetricsEngine_ExportPrometheus(t *testing.T) {
	metrics := NewMetricsEngine()
	metrics.RecordRequest(12.0, "ALLOW", 10, true, false, false)

	sloEngine := NewSLOEngine(5 * time.Minute)
	sloSum := sloEngine.Evaluate(time.Now().UTC())

	promText := metrics.ExportPrometheus(
		&sloSum,
		"fraud-xgb-25f-v3.0",
		"fraud-xgb-15f-v1.5",
		DriftStatusHealthy,
		0.042,
		0.015,
		25,
		CircuitStateHealthy,
		false,
	)

	assert.Contains(t, promText, "risk_evaluations_total 1")
	assert.Contains(t, promText, "risk_evaluation_success_total 1")
	assert.Contains(t, promText, `model_active_info{version="fraud-xgb-25f-v3.0",role="production"} 1`)
	assert.Contains(t, promText, `model_active_info{version="fraud-xgb-15f-v1.5",role="fallback"} 1`)
	assert.Contains(t, promText, "drift_status 0")
	assert.Contains(t, promText, "drift_max_psi 0.0420")
	assert.Contains(t, promText, "canary_stage 25")
	assert.Contains(t, promText, "circuit_breaker_state 0")
	assert.Contains(t, promText, "slo_availability 1.000000")
	assert.True(t, strings.HasPrefix(promText, "# HELP"))
}

func TestMetricsEngine_ConcurrentLoad(t *testing.T) {
	metrics := NewMetricsEngine()
	var wg sync.WaitGroup

	numGoroutines := 20
	opsPerGoroutine := 1000

	for g := 0; g < numGoroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				metrics.RecordRequest(float64(i%50), "ALLOW", 20, true, false, false)
			}
		}(g)
	}

	wg.Wait()
	snap := metrics.GetSnapshot()
	assert.Equal(t, int64(numGoroutines*opsPerGoroutine), snap["requests_total"])
	assert.Equal(t, int64(numGoroutines*opsPerGoroutine), snap["requests_success"])
}
