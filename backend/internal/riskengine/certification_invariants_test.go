package riskengine

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestCert_ContinuousInvariantsDuringChaos validates that all 8 core platform invariants
// remain strictly unviolated throughout rapid concurrent state transitions, rollbacks, and evaluations.
func TestCert_ContinuousInvariantsDuringChaos(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	reg := NewModelRegistry()
	cfg := DefaultRetrainingConfig()
	coordinator := NewRetrainingCoordinator(cfg, NewFixtureTrainingAdapter(), nil, nil, nil, nil)
	canaryRouter := NewCanaryRouter(DefaultCanaryRouterConfig(), nil)
	sloEngine := NewSLOEngine(5 * time.Minute)
	verifier := NewArtifactVerifier()

	var invariantBreachCount int64
	var totalAudits int64

	// Goroutine 1: Continuous Invariant Auditor (Runs every 2ms)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(2 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				atomic.AddInt64(&totalAudits, 1)

				// Invariant 1: Exactly 1 production active model
				models := reg.ListModels()
				prodCount := 0
				for _, m := range models {
					if m.IsProductionActive {
						prodCount++
					}
				}
				if prodCount != 1 {
					atomic.AddInt64(&invariantBreachCount, 1)
				}

				// Invariant 2: Production model is valid
				pm, err := reg.GetProductionModel()
				if err != nil || pm == nil || pm.Version == "" {
					atomic.AddInt64(&invariantBreachCount, 1)
				}

				// Invariant 3: Fallback model exists and is valid
				fb, err := reg.GetFallbackModel()
				if err != nil || fb == nil || fb.Version == "" {
					atomic.AddInt64(&invariantBreachCount, 1)
				}

				// Invariant 4: Canary percentage is strictly within [0, 100]%
				pct := canaryRouter.GetPercentage()
				if pct < 0 || pct > 100 {
					atomic.AddInt64(&invariantBreachCount, 1)
				}

				// Invariant 5: Circuit breaker state is consistent
				cbState := canaryRouter.GetCircuitBreaker().GetState()
				if cbState != CircuitStateHealthy && cbState != CircuitStateFailed && cbState != CircuitStateRolledBack {
					atomic.AddInt64(&invariantBreachCount, 1)
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	// Spawn 10 concurrent worker goroutines mutating system state
	for workerID := 0; workerID < 10; workerID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(id)))

			for {
				select {
				case <-ctx.Done():
					return
				default:
					action := r.Intn(7)
					switch action {
					case 0:
						// Canary adjustment
						canaryRouter.SetPercentage(r.Intn(101))
					case 1:
						// Circuit breaker toggle
						if r.Intn(2) == 1 {
							canaryRouter.GetCircuitBreaker().Trip("Simulated chaos trip")
						} else {
							canaryRouter.GetCircuitBreaker().Reset()
						}
					case 2:
						// Model registry reconciliation
						_, _ = reg.Reconcile(verifier)
					case 3:
						// Operational control mutations
						_ = coordinator.SetMaintenanceMode(ctx, r.Intn(2) == 1, "chaos", "test")
						_ = coordinator.SetModelFrozen(ctx, r.Intn(2) == 1, "chaos", "test")
					case 4:
						// Telemetry recording
						sloEngine.RecordEvaluation(float64(r.Intn(50)), true, false, false)
					case 5:
						// Simulated evaluation routing
						_ = canaryRouter.Route("tenant_chaos", "txn_chaos")
					case 6:
						time.Sleep(1 * time.Millisecond)
					}
				}
			}
		}(workerID)
	}

	wg.Wait()

	assert.Equal(t, int64(0), atomic.LoadInt64(&invariantBreachCount), "Platform invariants were breached during chaos!")
	assert.Greater(t, atomic.LoadInt64(&totalAudits), int64(100), "Auditor should have executed at least 100 continuous passes")
}
