package riskengine

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSLOEngine_DefaultEvaluation(t *testing.T) {
	engine := NewSLOEngine(5 * time.Minute)
	require.NotNil(t, engine)

	summary := engine.Evaluate(time.Now().UTC())
	assert.Equal(t, 10, summary.TotalSLOs)
	assert.Equal(t, SLOStatusHealthy, summary.OverallStatus)
	assert.Equal(t, 10, summary.HealthyCount)
	assert.Equal(t, 0, summary.BreachedCount)
}

func TestSLOEngine_AvailabilityAndLatency(t *testing.T) {
	engine := NewSLOEngine(5 * time.Minute)

	// Record 990 successful requests with 10ms latency, 10 failed requests
	for i := 0; i < 990; i++ {
		engine.RecordEvaluation(10.0, true, false, false)
	}
	for i := 0; i < 10; i++ {
		engine.RecordEvaluation(120.0, false, true, true)
	}

	summary := engine.Evaluate(time.Now().UTC())
	assert.Equal(t, 10, summary.TotalSLOs)

	// Availability = 990 / 1000 = 0.990 (Breaches target of 0.999)
	mAvail, ok := summary.Measurements["slo_availability"]
	require.True(t, ok)
	assert.Equal(t, SLOStatusBreached, mAvail.Status)
	assert.Equal(t, 0.99, mAvail.CurrentValue)
	assert.True(t, mAvail.BurnRate > 1.0)
	assert.Equal(t, 0.0, mAvail.ErrorBudgetRemaining)

	// P95 Latency = 10ms (Healthy)
	mP95, ok := summary.Measurements["slo_p95_latency"]
	require.True(t, ok)
	assert.Equal(t, SLOStatusHealthy, mP95.Status)
	assert.Equal(t, 10.0, mP95.CurrentValue)

	// Overall should be BREACHED
	assert.Equal(t, SLOStatusBreached, summary.OverallStatus)
}

func TestSLOEngine_DriftAndRetraining(t *testing.T) {
	engine := NewSLOEngine(5 * time.Minute)

	// Drift Max PSI in warning range
	engine.RecordDriftMetrics(0.15, DriftStatusWarning)

	summary := engine.Evaluate(time.Now().UTC())
	mDrift, ok := summary.Measurements["slo_drift_health"]
	require.True(t, ok)
	assert.Equal(t, SLOStatusWarning, mDrift.Status)
	assert.Equal(t, 0.15, mDrift.CurrentValue)

	// Retraining failure
	engine.RecordRetrainingOutcome(false)
	summary2 := engine.Evaluate(time.Now().UTC())
	mRetrain, ok := summary2.Measurements["slo_retraining_success_rate"]
	require.True(t, ok)
	assert.Equal(t, SLOStatusBreached, mRetrain.Status)
	assert.Equal(t, 0.0, mRetrain.CurrentValue)
}

func TestSLOEngine_CanaryRollbackAndDependencies(t *testing.T) {
	engine := NewSLOEngine(5 * time.Minute)

	engine.RecordCanaryRollback()
	engine.RecordDependencyCheck("postgres", true)
	engine.RecordDependencyCheck("postgres", false)

	summary := engine.Evaluate(time.Now().UTC())
	mCanary, ok := summary.Measurements["slo_canary_rollback_rate"]
	require.True(t, ok)
	assert.Equal(t, SLOStatusBreached, mCanary.Status)

	mDep, ok := summary.Measurements["slo_dependency_availability"]
	require.True(t, ok)
	assert.Equal(t, SLOStatusBreached, mDep.Status)
	assert.Equal(t, 0.50, mDep.CurrentValue)
}
