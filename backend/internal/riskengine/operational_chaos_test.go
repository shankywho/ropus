package riskengine

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOperationalChaos_ConcurrentTelemetryAndSafetySwitches(t *testing.T) {
	metrics := NewMetricsEngine()
	sloEngine := NewSLOEngine(5 * time.Minute)
	alertSink := NewInMemoryAlertSink(500)
	alertMgr := NewAlertManager(alertSink)
	defer alertMgr.Close()

	incidentEngine := NewIncidentEngine(alertMgr, nil)

	cfg := DefaultRetrainingConfig()
	cfg.CanaryObservationWindow = 10 * time.Millisecond
	coordinator := NewRetrainingCoordinator(cfg, NewFixtureTrainingAdapter(), nil, nil, nil, nil)
	healthAggregator := NewHealthAggregator(nil, nil, nil, nil, nil, coordinator, nil, sloEngine)
	defer healthAggregator.Close()

	var wg sync.WaitGroup
	ctx := context.Background()

	// 1. High-throughput telemetry worker goroutines
	numInferenceWorkers := 10
	numIterations := 500

	for w := 0; w < numInferenceWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < numIterations; i++ {
				lat := float64(5 + rand.Intn(100))
				isSuccess := rand.Float64() > 0.05
				isFallback := !isSuccess && rand.Float64() > 0.5
				isInferErr := !isSuccess

				metrics.RecordRequest(lat, "ALLOW", 20, isSuccess, isFallback, isInferErr)
				sloEngine.RecordEvaluation(lat, isSuccess, isFallback, isInferErr)
			}
		}(w)
	}

	// 2. Continuous Operational Control Flips
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 50; i++ {
			_ = coordinator.SetMaintenanceMode(ctx, i%2 == 1, "chaos_agent", "toggling maintenance")
			_ = coordinator.SetModelFrozen(ctx, i%3 == 0, "chaos_agent", "toggling model freeze")
			_ = coordinator.SetRetrainingPaused(ctx, i%4 == 0, "chaos_agent", "toggling retraining pause")
			_ = coordinator.SetCanaryPaused(ctx, i%5 == 0, "chaos_agent", "toggling canary pause")
			time.Sleep(2 * time.Millisecond)
		}
	}()

	// 3. Continuous Incident Evaluations
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			health := healthAggregator.GetHealthReport()
			slo := sloEngine.Evaluate(time.Now().UTC())
			_ = incidentEngine.Evaluate(ctx, health, slo, CircuitStateHealthy, DriftStatusHealthy, StateIdle, "fraud-xgb-25f-v3.0")
			time.Sleep(5 * time.Millisecond)
		}
	}()

	wg.Wait()

	// Verify no panics and data integrity
	snap := metrics.GetSnapshot()
	assert.Equal(t, int64(numInferenceWorkers*numIterations), snap["requests_total"])

	sloSummary := sloEngine.Evaluate(time.Now().UTC())
	assert.Equal(t, 10, sloSummary.TotalSLOs)

	healthReport := healthAggregator.GetHealthReport()
	assert.Equal(t, 14, len(healthReport.Components))
}

func TestOperationalChaos_CircuitBreakerIncidentCorrelation(t *testing.T) {
	alertSink := NewInMemoryAlertSink(100)
	alertMgr := NewAlertManager(alertSink)
	defer alertMgr.Close()

	incidentEngine := NewIncidentEngine(alertMgr, nil)
	ctx := context.Background()

	health := SystemHealthReport{
		OverallStatus: HealthStatusDegraded,
		Components:    make(map[string]ComponentHealth),
	}
	slo := SLOSummary{}

	// 1. Circuit breaker trips
	incidents := incidentEngine.Evaluate(ctx, health, slo, CircuitStateRolledBack, DriftStatusHealthy, StateIdle, "v3.0")
	require.NotEmpty(t, incidents)

	var cbInc *Incident
	for _, inc := range incidents {
		if inc.Category == IncidentCategoryCircuitBreakerTrip {
			cbInc = &inc
			break
		}
	}
	require.NotNil(t, cbInc)
	assert.Equal(t, IncidentSeverityCritical, cbInc.Severity)
	assert.Equal(t, IncidentStateOpen, cbInc.Status)

	// 2. Circuit breaker restored to HEALTHY
	incidents2 := incidentEngine.Evaluate(ctx, health, slo, CircuitStateHealthy, DriftStatusHealthy, StateIdle, "v3.0")
	for _, inc := range incidents2 {
		if inc.Category == IncidentCategoryCircuitBreakerTrip {
			assert.Equal(t, IncidentStateResolved, inc.Status)
		}
	}
}
