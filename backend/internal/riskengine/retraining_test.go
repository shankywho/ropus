package riskengine

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -------------------------------------------------------------
// 1. Trigger Engine Unit Tests
// -------------------------------------------------------------

func TestRetrainingTriggerEngine_StableTraffic_NoTrigger(t *testing.T) {
	cfg := DefaultRetrainingConfig()
	cfg.MinSamples = 100
	engine := NewRetrainingTriggerEngine(cfg)

	meas := &DriftMeasurement{
		SampleCount:   200,
		OverallStatus: DriftStatusHealthy,
		MaxPSI:        0.04,
	}

	shouldTrigger, triggerType, reason := engine.EvaluateDrift(meas, CircuitStateHealthy)
	assert.False(t, shouldTrigger)
	assert.Equal(t, "NONE", triggerType)
	assert.Contains(t, reason, "stable")
}

func TestRetrainingTriggerEngine_WarningDrift_NoTrigger(t *testing.T) {
	cfg := DefaultRetrainingConfig()
	cfg.MinSamples = 100
	cfg.DriftThreshold = 0.20
	engine := NewRetrainingTriggerEngine(cfg)

	meas := &DriftMeasurement{
		SampleCount:   200,
		OverallStatus: DriftStatusWarning,
		MaxPSI:        0.12,
	}

	shouldTrigger, triggerType, _ := engine.EvaluateDrift(meas, CircuitStateHealthy)
	assert.False(t, shouldTrigger)
	assert.Equal(t, "NONE", triggerType)
}

func TestRetrainingTriggerEngine_SustainedHighDrift_TriggersAfterConsecutiveWindows(t *testing.T) {
	cfg := DefaultRetrainingConfig()
	cfg.MinSamples = 100
	cfg.DriftThreshold = 0.20
	cfg.RequiredConsecutiveWindows = 2
	engine := NewRetrainingTriggerEngine(cfg)

	meas := &DriftMeasurement{
		SampleCount:   200,
		OverallStatus: DriftStatusDegraded,
		MaxPSI:        0.24,
	}

	// Window 1: Observe first window
	shouldTrigger1, _, reason1 := engine.EvaluateDrift(meas, CircuitStateHealthy)
	assert.False(t, shouldTrigger1)
	assert.Contains(t, reason1, "1/2 required consecutive windows")

	// Window 2: Sustained condition met -> Trigger!
	shouldTrigger2, triggerType2, reason2 := engine.EvaluateDrift(meas, CircuitStateHealthy)
	assert.True(t, shouldTrigger2)
	assert.Equal(t, "DRIFT_SUSTAINED", triggerType2)
	assert.Contains(t, reason2, "Sustained drift across 2 consecutive")
}

func TestRetrainingTriggerEngine_CriticalDrift_ImmediateTrigger(t *testing.T) {
	cfg := DefaultRetrainingConfig()
	cfg.MinSamples = 100
	engine := NewRetrainingTriggerEngine(cfg)

	meas := &DriftMeasurement{
		SampleCount:          250,
		OverallStatus:        DriftStatusCritical,
		MaxPSI:               0.45,
		CriticalFeatureCount: 3,
	}

	shouldTrigger, triggerType, reason := engine.EvaluateDrift(meas, CircuitStateHealthy)
	assert.True(t, shouldTrigger)
	assert.Equal(t, "DRIFT_CRITICAL", triggerType)
	assert.Contains(t, reason, "CRITICAL drift detected")
}

func TestRetrainingTriggerEngine_CooldownAndSampleQuorum(t *testing.T) {
	cfg := DefaultRetrainingConfig()
	cfg.MinSamples = 200
	cfg.CooldownDuration = 10 * time.Minute
	engine := NewRetrainingTriggerEngine(cfg)

	// 1. Insufficient samples
	lowSampleMeas := &DriftMeasurement{
		SampleCount:   50, // < 200 min
		OverallStatus: DriftStatusCritical,
		MaxPSI:        0.50,
	}
	shouldTriggerLow, _, reasonLow := engine.EvaluateDrift(lowSampleMeas, CircuitStateHealthy)
	assert.False(t, shouldTriggerLow)
	assert.Contains(t, reasonLow, "below minimum required quorum")

	// 2. Sufficient samples triggers job
	goodMeas := &DriftMeasurement{
		SampleCount:   300,
		OverallStatus: DriftStatusCritical,
		MaxPSI:        0.50,
	}
	shouldTrigger, _, _ := engine.EvaluateDrift(goodMeas, CircuitStateHealthy)
	assert.True(t, shouldTrigger)

	// Mark active job
	engine.SetActiveJob("job_123", "CRITICAL drift")

	// 3. Duplicate trigger while active rejected
	shouldTriggerDup, _, reasonDup := engine.EvaluateDrift(goodMeas, CircuitStateHealthy)
	assert.False(t, shouldTriggerDup)
	assert.Contains(t, reasonDup, "currently active")

	// Clear active job
	engine.ClearActiveJob(true)

	// 4. Cooldown suppression
	shouldTriggerCooldown, _, reasonCooldown := engine.EvaluateDrift(goodMeas, CircuitStateHealthy)
	assert.False(t, shouldTriggerCooldown)
	assert.Contains(t, reasonCooldown, "cooldown window")
}

func TestRetrainingTriggerEngine_CircuitBreakerUnhealthy_SuppressesTrigger(t *testing.T) {
	cfg := DefaultRetrainingConfig()
	cfg.MinSamples = 100
	engine := NewRetrainingTriggerEngine(cfg)

	meas := &DriftMeasurement{
		SampleCount:   300,
		OverallStatus: DriftStatusCritical,
		MaxPSI:        0.50,
	}

	shouldTrigger, _, reason := engine.EvaluateDrift(meas, CircuitStateRolledBack)
	assert.False(t, shouldTrigger)
	assert.Contains(t, reason, "Circuit breaker is not healthy")
}

// -------------------------------------------------------------
// 2. State Machine Transition Tests
// -------------------------------------------------------------

func TestRetrainingStateMachine_ValidTransitions(t *testing.T) {
	validSequence := []struct {
		from JobState
		to   JobState
	}{
		{StateIdle, StateTriggered},
		{StateTriggered, StateQueued},
		{StateQueued, StateTraining},
		{StateTraining, StateValidating},
		{StateValidating, StateShadowEvaluation},
		{StateShadowEvaluation, StateCanary},
		{StateCanary, StatePromoted},
		{StatePromoted, StateIdle},
	}

	for _, step := range validSequence {
		t.Run(fmt.Sprintf("%s->%s", step.from, step.to), func(t *testing.T) {
			assert.True(t, step.from.CanTransitionTo(step.to))
		})
	}
}

func TestRetrainingStateMachine_InvalidTransitions(t *testing.T) {
	invalidSequence := []struct {
		from JobState
		to   JobState
	}{
		{StateIdle, StatePromoted},
		{StateIdle, StateTraining},
		{StateTraining, StatePromoted},
		{StateValidating, StatePromoted},
		{StateShadowEvaluation, StateIdle},
	}

	for _, step := range invalidSequence {
		t.Run(fmt.Sprintf("Invalid_%s->%s", step.from, step.to), func(t *testing.T) {
			assert.False(t, step.from.CanTransitionTo(step.to))
		})
	}
}

// -------------------------------------------------------------
// 3. Training Runner & Dataset Validation Tests
// -------------------------------------------------------------

func TestTrainingRunner_DatasetValidation(t *testing.T) {
	runner := NewLocalTrainingAdapter()
	ctx := context.Background()

	// 1. Valid dataset metadata
	validMeta := TrainingDatasetMetadata{
		DatasetID:          "dataset_good_01",
		SampleCount:        500,
		FeatureContract:    MLFeatureContractV25,
		ParentModelVersion: "fraud-xgb-25f-v3.0",
		DataQualityScore:   0.95,
		MissingValueRate:   0.02,
		ZeroLabelsDetected: false,
	}
	assert.NoError(t, runner.ValidateDataset(ctx, validMeta))

	// 2. Low sample count
	lowSamples := validMeta
	lowSamples.SampleCount = 20
	assert.Error(t, runner.ValidateDataset(ctx, lowSamples))

	// 3. Incompatible feature contract
	badContract := validMeta
	badContract.FeatureContract = "fraud-risk-10f-legacy"
	assert.Error(t, runner.ValidateDataset(ctx, badContract))

	// 4. Zero positive fraud labels
	zeroLabels := validMeta
	zeroLabels.ZeroLabelsDetected = true
	assert.Error(t, runner.ValidateDataset(ctx, zeroLabels))

	// 5. High missing value rate
	highMissing := validMeta
	highMissing.MissingValueRate = 0.35
	assert.Error(t, runner.ValidateDataset(ctx, highMissing))
}

func TestTrainingRunner_SuccessfulTrainingAndArtifactGeneration(t *testing.T) {
	runner := NewLocalTrainingAdapter()
	ctx := context.Background()

	req := TrainingRequest{
		JobID:              "job_test_001",
		ParentModelVersion: "fraud-xgb-25f-v3.0",
		FeatureContract:    MLFeatureContractV25,
		DatasetMetadata: TrainingDatasetMetadata{
			DatasetID:        "dataset_test_001",
			SampleCount:      500,
			FeatureContract:  MLFeatureContractV25,
			DataQualityScore: 0.95,
			MissingValueRate: 0.01,
		},
		TriggerReason: "Sustained drift detected",
		Actor:         "AUTOMATED_DRIFT_DETECTOR",
	}

	res, err := runner.StartTraining(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.NotEmpty(t, res.Candidate.ModelID)
	assert.Contains(t, res.Candidate.Version, "fraud-xgb-25f-v3.1-candidate")
	assert.Equal(t, "fraud-xgb-25f-v3.0", res.Candidate.ParentModelVersion)
	assert.NotEmpty(t, res.Candidate.ArtifactChecksum)
	assert.NotEmpty(t, res.Candidate.ConfigHash)
	assert.Greater(t, res.ValidationMetrics.ROCAUC, 0.90)
	assert.Less(t, res.ValidationMetrics.FPR, 0.05)
	assert.Less(t, res.ValidationMetrics.BrierScore, 0.05)
}

// -------------------------------------------------------------
// 4. Offline Validation Gate Tests
// -------------------------------------------------------------

func TestOfflineValidator_CandidateComparisons(t *testing.T) {
	cfg := DefaultRetrainingConfig()
	validator := NewOfflineValidator(cfg)
	baseline := DefaultProductionBaselineMetrics()

	candidate := ModelCandidate{
		ModelID:            "model_cand_01",
		Version:            "fraud-xgb-25f-v3.1-candidate",
		ParentModelVersion: "fraud-xgb-25f-v3.0",
	}

	// 1. Passing candidate (Higher AUC, low FPR, low latency)
	goodMetrics := ValidationMetrics{
		ROCAUC:           0.9200, // > baseline 0.895
		PRAUC:            0.7500,
		Precision:        0.8400,
		Recall:           0.7800,
		FPR:              0.0200, // < 0.05
		BrierScore:       0.0380, // < 0.10
		CalibrationError: 0.0120,
		P95LatencyMs:     6.00, // < 15.0ms
		InferenceErrors:  0,
		NaNCount:         0,
	}
	resGood := validator.ValidateCandidate(candidate, goodMetrics, baseline)
	assert.True(t, resGood.Passed)
	assert.Empty(t, resGood.Violations)

	// 2. Failing candidate: severe ROC-AUC regression
	badAUCMetrics := goodMetrics
	badAUCMetrics.ROCAUC = 0.8200 // < 0.895 - 0.02 = 0.875
	resBadAUC := validator.ValidateCandidate(candidate, badAUCMetrics, baseline)
	assert.False(t, resBadAUC.Passed)
	assert.Contains(t, resBadAUC.GateDetails, "ROC-AUC")

	// 3. Failing candidate: excessive FPR
	badFPRMetrics := goodMetrics
	badFPRMetrics.FPR = 0.08 // > 0.05 limit
	resBadFPR := validator.ValidateCandidate(candidate, badFPRMetrics, baseline)
	assert.False(t, resBadFPR.Passed)
	assert.Contains(t, resBadFPR.GateDetails, "FPR")

	// 4. Failing candidate: latency regression
	badLatMetrics := goodMetrics
	badLatMetrics.P95LatencyMs = 28.5 // > 15.0ms limit
	resBadLat := validator.ValidateCandidate(candidate, badLatMetrics, baseline)
	assert.False(t, resBadLat.Passed)
	assert.Contains(t, resBadLat.GateDetails, "latency")

	// 5. Failing candidate: arithmetic error / NaNs
	badNaNMetrics := goodMetrics
	badNaNMetrics.NaNCount = 3
	resBadNaN := validator.ValidateCandidate(candidate, badNaNMetrics, baseline)
	assert.False(t, resBadNaN.Passed)
	assert.Contains(t, resBadNaN.GateDetails, "NaN")
}

// -------------------------------------------------------------
// 5. End-to-End Retraining Coordinator & Canary Promotion Tests
// -------------------------------------------------------------

func TestRetrainingCoordinator_EndToEnd_TriggerToPromotion(t *testing.T) {
	cfg := DefaultRetrainingConfig()
	cfg.CanaryObservationWindow = 1 * time.Millisecond // Fast for test
	cfg.AutoApproveCanary = true
	cfg.MinSamples = 50

	canaryRouter := NewCanaryRouter(DefaultCanaryRouterConfig(), nil)
	shadowScorer := NewShadowScorer(DefaultShadowScorerConfig(), nil, nil)

	var mu sync.Mutex
	var promotedVersion string
	coordinator := NewRetrainingCoordinator(
		cfg,
		NewLocalTrainingAdapter(),
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

	// 1. Manual trigger
	job, err := coordinator.TriggerManual(ctx, "ADMIN_TEST", "Integration test trigger")
	require.NoError(t, err)
	require.NotNil(t, job)
	assert.Equal(t, StateTriggered, job.State)

	// Wait for asynchronous pipeline completion
	time.Sleep(50 * time.Millisecond)

	status := coordinator.GetStatus()
	assert.Equal(t, StatePromoted, status["state"])
	mu.Lock()
	pv := promotedVersion
	mu.Unlock()
	assert.NotEmpty(t, pv)
	assert.Contains(t, pv, "fraud-xgb-25f-v3.1-candidate")

	// Check candidate record
	candidates := coordinator.GetCandidates()
	require.Len(t, candidates, 1)
	assert.Equal(t, StatePromoted, candidates[0].State)
	assert.True(t, candidates[0].ValidationResult.Passed)
	assert.True(t, candidates[0].ShadowResult.Passed)

	// Check history
	history := coordinator.GetHistory()
	require.Len(t, history, 1)
	assert.Equal(t, StatePromoted, history[0].State)
	assert.Greater(t, history[0].DurationMs, 0.0)
}

func TestRetrainingCoordinator_CanarySafetyBreach_RollsBackToZero(t *testing.T) {
	cfg := DefaultRetrainingConfig()
	cfg.CanaryObservationWindow = 10 * time.Millisecond
	cfg.AutoApproveCanary = true
	cfg.MinSamples = 50

	canaryRouter := NewCanaryRouter(DefaultCanaryRouterConfig(), nil)
	shadowScorer := NewShadowScorer(DefaultShadowScorerConfig(), nil, nil)

	coordinator := NewRetrainingCoordinator(
		cfg,
		NewLocalTrainingAdapter(),
		shadowScorer,
		canaryRouter,
		nil,
		nil,
	)

	ctx := context.Background()

	// Trigger job
	_, err := coordinator.TriggerManual(ctx, "ADMIN_TEST", "Rollback verification test")
	require.NoError(t, err)

	// Simulate Circuit Breaker trip during canary rollout
	time.Sleep(5 * time.Millisecond)
	canaryRouter.circuitBreaker.mu.Lock()
	canaryRouter.circuitBreaker.state = CircuitStateRolledBack
	canaryRouter.circuitBreaker.lastTripReason = "Simulated canary latency breach"
	canaryRouter.circuitBreaker.mu.Unlock()

	// Wait for coordinator to observe trip
	time.Sleep(50 * time.Millisecond)

	status := coordinator.GetStatus()
	assert.Equal(t, StateRolledBack, status["state"])

	// Canary percentage must be rolled back to 0%
	canaryStatus := canaryRouter.GetStatus()
	assert.Equal(t, 0, canaryStatus["target_percentage"])
	assert.False(t, canaryStatus["enabled"].(bool))
}

func TestRetrainingCoordinator_OperatorApprovalFlow(t *testing.T) {
	cfg := DefaultRetrainingConfig()
	cfg.CanaryObservationWindow = 1 * time.Millisecond
	cfg.AutoApproveCanary = false // Explicit operator approval required
	cfg.MinSamples = 50

	canaryRouter := NewCanaryRouter(DefaultCanaryRouterConfig(), nil)
	shadowScorer := NewShadowScorer(DefaultShadowScorerConfig(), nil, nil)

	var mu sync.Mutex
	var promotedVersion string
	coordinator := NewRetrainingCoordinator(
		cfg,
		NewLocalTrainingAdapter(),
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

	// 1. Manual trigger
	job, err := coordinator.TriggerManual(ctx, "ADMIN_TEST", "Operator approval test trigger")
	require.NoError(t, err)
	require.NotNil(t, job)

	// Wait for pipeline to reach AWAITING_APPROVAL
	time.Sleep(30 * time.Millisecond)

	status := coordinator.GetStatus()
	assert.Equal(t, StateAwaitingApproval, status["state"])

	candidates := coordinator.GetCandidates()
	require.Len(t, candidates, 1)
	cand := candidates[0]
	assert.Equal(t, StateAwaitingApproval, cand.State)

	// 2. Operator issues Approval
	err = coordinator.ApproveCandidate(ctx, cand.ModelID, "ADMIN_OPERATOR", "Approved for canary rollout")
	require.NoError(t, err)

	// Wait for canary rollout to complete
	time.Sleep(50 * time.Millisecond)

	statusPromoted := coordinator.GetStatus()
	assert.Equal(t, StatePromoted, statusPromoted["state"])
	mu.Lock()
	pv := promotedVersion
	mu.Unlock()
	assert.NotEmpty(t, pv)
}

func TestRetrainingCoordinator_OperatorRejectionFlow(t *testing.T) {
	cfg := DefaultRetrainingConfig()
	cfg.CanaryObservationWindow = 1 * time.Millisecond
	cfg.AutoApproveCanary = false
	cfg.MinSamples = 50

	canaryRouter := NewCanaryRouter(DefaultCanaryRouterConfig(), nil)
	shadowScorer := NewShadowScorer(DefaultShadowScorerConfig(), nil, nil)

	coordinator := NewRetrainingCoordinator(
		cfg,
		NewLocalTrainingAdapter(),
		shadowScorer,
		canaryRouter,
		nil,
		nil,
	)

	ctx := context.Background()

	// 1. Trigger job
	_, err := coordinator.TriggerManual(ctx, "ADMIN_TEST", "Test rejection flow")
	require.NoError(t, err)

	// Wait for pipeline to reach AWAITING_APPROVAL
	time.Sleep(30 * time.Millisecond)

	status := coordinator.GetStatus()
	assert.Equal(t, StateAwaitingApproval, status["state"])

	candidates := coordinator.GetCandidates()
	require.Len(t, candidates, 1)
	cand := candidates[0]

	// 2. Operator issues Rejection
	err = coordinator.RejectCandidate(ctx, cand.ModelID, "ADMIN_OPERATOR", "Candidate rejected due to operational risk")
	require.NoError(t, err)

	statusRejected := coordinator.GetStatus()
	assert.Equal(t, StateRejected, statusRejected["state"])
}

// -------------------------------------------------------------
// 6. Concurrency & Race Safety Tests
// -------------------------------------------------------------

func TestRetrainingCoordinator_ConcurrentAccess(t *testing.T) {
	cfg := DefaultRetrainingConfig()
	cfg.CanaryObservationWindow = 1 * time.Millisecond
	cfg.MinSamples = 50

	canaryRouter := NewCanaryRouter(DefaultCanaryRouterConfig(), nil)
	shadowScorer := NewShadowScorer(DefaultShadowScorerConfig(), nil, nil)

	coordinator := NewRetrainingCoordinator(
		cfg,
		NewLocalTrainingAdapter(),
		shadowScorer,
		canaryRouter,
		nil,
		nil,
	)

	var wg sync.WaitGroup
	ctx := context.Background()

	// Concurrent status queries
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				_ = coordinator.GetStatus()
				_ = coordinator.GetSummary()
				_ = coordinator.GetHistory()
				_ = coordinator.GetCandidates()
			}
		}()
	}

	// Concurrent trigger attempts (only 1 should succeed, rest rejected)
	var successfulTriggers int64
	var triggerMu sync.Mutex

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			job, err := coordinator.TriggerManual(ctx, "ADMIN_CONC", fmt.Sprintf("Trigger attempt %d", idx))
			if err == nil && job != nil {
				triggerMu.Lock()
				successfulTriggers++
				triggerMu.Unlock()
			}
		}(i)
	}

	wg.Wait()
	assert.Equal(t, int64(1), successfulTriggers, "Expected strictly 1 concurrent trigger to succeed")
}
