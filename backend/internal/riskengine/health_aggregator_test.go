package riskengine

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthAggregator_BasicReport(t *testing.T) {
	sloEngine := NewSLOEngine(5 * time.Minute)
	agg := NewHealthAggregator(nil, nil, nil, nil, nil, nil, nil, sloEngine)
	defer agg.Close()

	report := agg.GetHealthReport()
	require.NotNil(t, report)
	assert.Equal(t, HealthStatusHealthy, report.OverallStatus)
	assert.Equal(t, 14, len(report.Components))

	// Verify all 14 platform components exist
	expected := []string{
		"api", "risk_engine", "model", "drift", "retraining",
		"canary", "circuit_breaker", "postgres", "redis", "clickhouse",
		"ml_runtime", "artifact_store", "model_registry", "recovery_manager",
	}

	for _, name := range expected {
		comp, exists := report.Components[name]
		assert.True(t, exists, "Component %s missing from health report", name)
		assert.NotEmpty(t, comp.Message)
	}
}

func TestHealthAggregator_ComponentFailureDegradation(t *testing.T) {
	agg := NewHealthAggregator(nil, nil, nil, nil, nil, nil, nil, nil)
	defer agg.Close()

	// 1. Degrade non-critical dependency (Redis)
	agg.updateComponent("redis", HealthStatusDegraded, "High latency connection", 150.0, time.Now().UTC())
	rep1 := agg.GetHealthReport()
	assert.Equal(t, HealthStatusDegraded, rep1.OverallStatus)
	assert.Equal(t, 1, rep1.Components["redis"].ConsecutiveFailures)

	// 2. Fail critical component (Postgres)
	agg.updateComponent("postgres", HealthStatusUnhealthy, "Connection refused", 0.0, time.Now().UTC())
	rep2 := agg.GetHealthReport()
	assert.Equal(t, HealthStatusUnhealthy, rep2.OverallStatus)

	// 3. Restore Postgres
	agg.updateComponent("postgres", HealthStatusHealthy, "Postgres healthy", 1.2, time.Now().UTC())
	rep3 := agg.GetHealthReport()
	assert.Equal(t, HealthStatusDegraded, rep3.OverallStatus) // Still degraded due to Redis

	// 4. Restore Redis
	agg.updateComponent("redis", HealthStatusHealthy, "Redis healthy", 0.8, time.Now().UTC())
	rep4 := agg.GetHealthReport()
	assert.Equal(t, HealthStatusHealthy, rep4.OverallStatus)
	assert.Equal(t, 0, rep4.Components["redis"].ConsecutiveFailures)
}
