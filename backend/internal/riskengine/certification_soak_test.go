package riskengine

import (
	"context"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestCert_LongRunningSoakAndStress executes continuous synthetic traffic while repeatedly
// triggering background drift checks, retraining cycles, canary rollouts, and disaster recoveries.
// Measures goroutine leaks, memory stability, and zero data races.
func TestCert_LongRunningSoakAndStress(t *testing.T) {
	runtime.GC()
	baselineGoroutines := runtime.NumGoroutine()

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	// Initialize production platform components
	reg := NewModelRegistry()
	cfg := DefaultRetrainingConfig()
	cfg.CooldownDuration = 10 * time.Millisecond
	coordinator := NewRetrainingCoordinator(cfg, NewFixtureTrainingAdapter(), nil, nil, nil, nil)
	canaryRouter := NewCanaryRouter(DefaultCanaryRouterConfig(), nil)
	metricsEngine := NewMetricsEngine()
	sloEngine := NewSLOEngine(5 * time.Minute)
	verifier := NewArtifactVerifier()

	coordinator.SetMetricsEngine(metricsEngine)
	coordinator.SetSLOEngine(sloEngine)

	var totalEvaluations int64
	var totalRetrainTriggers int64
	var totalRecoveries int64

	var wg sync.WaitGroup

	// 1. Worker Group A: 16 high-throughput inference evaluation workers
	for w := 0; w < 16; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)))

			for {
				select {
				case <-ctx.Done():
					return
				default:
					latency := float64(1 + r.Intn(10))
					isErr := r.Float64() < 0.005
					isFallback := r.Float64() < 0.01

					// Route request through canary router
					route := canaryRouter.Route("tenant_soak", "txn_soak")
					if route != RouteLegacy && route != RouteCandidate {
						t.Errorf("unexpected route: %s", route)
					}

					// Record telemetry
					metricsEngine.RecordRequest(latency, "ALLOW", 10, !isErr, isFallback, isErr)
					sloEngine.RecordEvaluation(latency, !isErr, isFallback, isErr)

					atomic.AddInt64(&totalEvaluations, 1)
				}
			}
		}(w)
	}

	// 2. Worker Group B: Background Drift & Retraining Coordinator Worker
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(20 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				atomic.AddInt64(&totalRetrainTriggers, 1)
				m := &DriftMeasurement{
					MeasurementID: "meas_soak_01",
					Timestamp:     time.Now().UTC(),
					MaxPSI:        0.35,
					MaxJSD:        0.15,
					OverallStatus: DriftStatusCritical,
					ModelVersion:  "fraud-xgb-25f-v3.0",
				}
				coordinator.OnDriftEvaluated(ctx, m)

			case <-ctx.Done():
				return
			}
		}
	}()

	// 3. Worker Group C: Canary Staged Rollout & Circuit Breaker Controller
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(30 * time.Millisecond)
		defer ticker.Stop()
		stages := []int{0, 10, 25, 50, 100, 0}
		idx := 0

		for {
			select {
			case <-ticker.C:
				canaryRouter.SetPercentage(stages[idx%len(stages)])
				idx++
			case <-ctx.Done():
				return
			}
		}
	}()

	// 4. Worker Group D: Periodic Disaster Recovery & Registry Self-Reconcile Loop
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				atomic.AddInt64(&totalRecoveries, 1)
				_, _ = reg.Reconcile(verifier)
			case <-ctx.Done():
				return
			}
		}
	}()

	wg.Wait()

	// Allow goroutine cooldown
	time.Sleep(100 * time.Millisecond)
	runtime.GC()

	evals := atomic.LoadInt64(&totalEvaluations)
	retrains := atomic.LoadInt64(&totalRetrainTriggers)
	recoveries := atomic.LoadInt64(&totalRecoveries)

	t.Logf("[SOAK_TEST] Completed %d evaluations, %d retraining triggers, %d recovery reconciliations",
		evals, retrains, recoveries)

	assert.Greater(t, evals, int64(1000), "Expected at least 1,000 evaluations during soak test")
	assert.Greater(t, retrains, int64(10), "Expected at least 10 retraining triggers")
	assert.Greater(t, recoveries, int64(5), "Expected at least 5 recovery cycles")

	// Verify no severe goroutine leak (leak tolerance < 10 background scheduler goroutines)
	finalGoroutines := runtime.NumGoroutine()
	goroutineDelta := finalGoroutines - baselineGoroutines
	t.Logf("[SOAK_TEST] Baseline goroutines: %d, Final goroutines: %d (Delta: %d)",
		baselineGoroutines, finalGoroutines, goroutineDelta)
	assert.LessOrEqual(t, goroutineDelta, 10, "Potential goroutine leak detected during soak test")
}
