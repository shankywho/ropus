package riskengine

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSafetyInvariants_PropertyTest(t *testing.T) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	ctx := context.Background()

	reg := NewModelRegistry()
	cfg := DefaultRetrainingConfig()
	coordinator := NewRetrainingCoordinator(cfg, NewFixtureTrainingAdapter(), nil, nil, nil, nil)
	canaryRouter := NewCanaryRouter(DefaultCanaryRouterConfig(), nil)
	sloEngine := NewSLOEngine(5 * time.Minute)
	verifier := NewArtifactVerifier()
	auditor := NewSafetyAuditor(reg, coordinator, canaryRouter, sloEngine, verifier)

	// Execute 50 randomized state transitions and operational actions
	for i := 0; i < 50; i++ {
		action := r.Intn(8)
		switch action {
		case 0:
			// Trigger Manual
			_, _ = coordinator.TriggerManual(ctx, "property_tester", "Random trigger")
		case 1:
			// Toggle Maintenance Mode
			_ = coordinator.SetMaintenanceMode(ctx, r.Intn(2) == 1, "property_tester", "Random maintenance")
		case 2:
			// Toggle Model Freeze
			_ = coordinator.SetModelFrozen(ctx, r.Intn(2) == 1, "property_tester", "Random freeze")
		case 3:
			// Toggle Retraining Pause
			_ = coordinator.SetRetrainingPaused(ctx, r.Intn(2) == 1, "property_tester", "Random pause")
		case 4:
			// Adjust Canary
			canaryRouter.SetPercentage(r.Intn(101))
		case 5:
			// Circuit breaker trip/reset
			if r.Intn(2) == 1 {
				canaryRouter.GetCircuitBreaker().Trip("Random trip")
			} else {
				canaryRouter.GetCircuitBreaker().Reset()
			}
		case 6:
			// Model registry reconcile
			_, _ = reg.Reconcile(verifier)
		case 7:
			// Record telemetry
			sloEngine.RecordEvaluation(float64(r.Intn(100)), true, false, false)
		}

		// ASSERT INVARIANTS:
		// 1. Exactly 1 production active model in registry
		models := reg.ListModels()
		prodCount := 0
		for _, m := range models {
			if m.IsProductionActive {
				prodCount++
			}
		}
		require.Equal(t, 1, prodCount, "Invariant breached: Expected exactly 1 production active model")

		// 2. Production model version is valid
		pm, err := reg.GetProductionModel()
		require.NoError(t, err)
		require.NotEmpty(t, pm.Version)

		// 3. Fallback model is available
		fb, err := reg.GetFallbackModel()
		require.NoError(t, err)
		require.NotEmpty(t, fb.Version)

		// 4. Canary percentage within [0, 100]
		pct := canaryRouter.GetPercentage()
		require.GreaterOrEqual(t, pct, 0)
		require.LessOrEqual(t, pct, 100)
	}

	// Final full safety audit
	report := auditor.Audit(ctx)
	assert.NotEmpty(t, report.Status)
}
