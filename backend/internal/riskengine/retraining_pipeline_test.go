package riskengine

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetrainingPipeline_FullLifecycleWithRegistry(t *testing.T) {
	cfg := DefaultRetrainingConfig()
	cfg.CanaryObservationWindow = 1 * time.Millisecond
	cfg.AutoApproveCanary = false // Explicit operator approval
	cfg.MinSamples = 50

	canaryRouter := NewCanaryRouter(DefaultCanaryRouterConfig(), nil)
	shadowScorer := NewShadowScorer(DefaultShadowScorerConfig(), nil, nil)
	runner := NewFixtureTrainingAdapter()

	var mu sync.Mutex
	var promotedVersion string
	coordinator := NewRetrainingCoordinator(
		cfg,
		runner,
		shadowScorer,
		canaryRouter,
		nil,
		func(newVer string) {
			mu.Lock()
			promotedVersion = newVer
			mu.Unlock()
		},
	)

	ctx := context.Background()

	// 1. Initial State
	reg := coordinator.GetModelRegistry()
	require.NotNil(t, reg)
	prodInitial, err := reg.GetProductionModel()
	require.NoError(t, err)
	assert.Equal(t, "fraud-xgb-25f-v3.0", prodInitial.Version)

	// 2. Trigger Job
	job, err := coordinator.TriggerManual(ctx, "ADMIN_TEST", "Pipeline integration test")
	require.NoError(t, err)
	require.NotNil(t, job)

	// 3. Wait for pipeline to reach AWAITING_APPROVAL
	time.Sleep(50 * time.Millisecond)

	status := coordinator.GetStatus()
	assert.Equal(t, StateAwaitingApproval, status["state"])

	candidates := coordinator.GetCandidates()
	require.Len(t, candidates, 1)
	cand := candidates[0]
	assert.Equal(t, StateAwaitingApproval, cand.State)

	// Model in registry must be in APPROVED/SHADOW status
	regModel, err := reg.GetModel(cand.Version)
	require.NoError(t, err)
	assert.Equal(t, "PASSED", regModel.ValidationStatus)
	assert.Equal(t, "PASSED", regModel.ShadowStatus)

	// 4. Operator issues Approval
	err = coordinator.ApproveCandidate(ctx, cand.ModelID, "ADMIN_OPERATOR", "Approved candidate for canary deployment")
	require.NoError(t, err)

	// Wait for canary staged rollout (1% -> 5% -> 10% -> 25% -> 50% -> 100%) to complete
	time.Sleep(60 * time.Millisecond)

	statusPromoted := coordinator.GetStatus()
	assert.Equal(t, StatePromoted, statusPromoted["state"])
	mu.Lock()
	pv := promotedVersion
	mu.Unlock()
	assert.NotEmpty(t, pv)
	assert.Equal(t, cand.Version, pv)

	// 5. Verify ModelRegistry Promotion
	newProd, err := reg.GetProductionModel()
	require.NoError(t, err)
	assert.Equal(t, cand.Version, newProd.Version)
	assert.True(t, newProd.IsProductionActive)

	newFallback, err := reg.GetFallbackModel()
	require.NoError(t, err)
	assert.Equal(t, "fraud-xgb-25f-v3.0", newFallback.Version)
	assert.True(t, newFallback.IsFallbackActive)

	// 6. Verify Observability Metrics
	metrics := coordinator.GetObservabilityMetrics()
	assert.GreaterOrEqual(t, metrics["training_jobs_started"], int64(1))
	assert.GreaterOrEqual(t, metrics["training_jobs_completed"], int64(1))
	assert.GreaterOrEqual(t, metrics["candidate_promotions"], int64(1))
}

func TestRetrainingPipeline_JobCancellation(t *testing.T) {
	cfg := DefaultRetrainingConfig()
	canaryRouter := NewCanaryRouter(DefaultCanaryRouterConfig(), nil)
	shadowScorer := NewShadowScorer(DefaultShadowScorerConfig(), nil, nil)
	runner := NewFixtureTrainingAdapter()

	coordinator := NewRetrainingCoordinator(
		cfg,
		runner,
		shadowScorer,
		canaryRouter,
		nil,
		nil,
	)

	ctx := context.Background()

	job, err := coordinator.TriggerManual(ctx, "ADMIN_TEST", "Cancellation test trigger")
	require.NoError(t, err)

	// Cancel job
	err = coordinator.CancelJob(ctx, job.JobID, "ADMIN_OPERATOR", "Emergency cancellation")
	require.NoError(t, err)

	status := coordinator.GetStatus()
	assert.Equal(t, StateFailed, status["state"])
}

func BenchmarkDatasetValidator(b *testing.B) {
	validator := NewDatasetValidator(100)
	ctx := context.Background()
	meta := TrainingDatasetMetadata{
		DatasetID:          "ds_bench_01",
		SampleCount:        1000,
		FeatureContract:    "fraud-risk-25f-v2.5",
		ParentModelVersion: "fraud-xgb-25f-v3.0",
		DataQualityScore:   0.98,
		MissingValueRate:   0.005,
		ZeroLabelsDetected: false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = validator.ValidateDatasetMetadata(ctx, meta)
	}
}

func BenchmarkModelRegistry_GetModel(b *testing.B) {
	registry := NewModelRegistry()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = registry.GetProductionModel()
	}
}
