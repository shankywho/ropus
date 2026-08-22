package riskengine

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// failingOrchestrator simulates a training runtime failure.
type failingOrchestrator struct{}

func (f *failingOrchestrator) StartPipeline(ctx context.Context, req PipelineRequest) (*PipelineRun, error) {
	return nil, assert.AnError
}

func (f *failingOrchestrator) GetPipelineStatus(ctx context.Context, runID string) (*PipelineRun, error) {
	return nil, assert.AnError
}

func (f *failingOrchestrator) CancelPipeline(ctx context.Context, runID string) error {
	return nil
}

func TestMLOps_TrainingFailureSafety(t *testing.T) {
	ctx := context.Background()
	reg := NewModelRegistry()
	engine := NewTrainingPipelineEngine(nil, nil, nil, &failingOrchestrator{}, reg)

	req := PipelineRequest{
		PipelineID:      "pipe_fail_01",
		ModelVersion:    "fraud-xgb-25f-cand-fail",
		DatasetChecksum: "valid_checksum_123",
	}

	_, _, err := engine.ExecutePipeline(ctx, req)
	assert.Error(t, err)

	// Ensure no broken candidate was registered
	_, err = reg.GetModel("fraud-xgb-25f-cand-fail")
	assert.Error(t, err, "Failed training must NOT register model in registry")

	// Ensure active production model is intact
	prod, err := reg.GetProductionModel()
	require.NoError(t, err)
	assert.Equal(t, "fraud-xgb-25f-v3.0", prod.Version)
}

func TestMLOps_EvaluationGateFailure(t *testing.T) {
	evaluator := &ModelEvaluator{
		MinROCAUC:     0.95, // High threshold
		MinF1Score:    0.90,
		MaxBrierScore: 0.10,
	}

	// Predictions that perform poorly
	predictions := []float64{0.9, 0.8, 0.1, 0.2, 0.9}
	groundTruth := []int{0, 0, 1, 1, 0} // Inverse / inverted accuracy

	report, err := evaluator.EvaluateModel("candidate_poor_v1", "chk_123", predictions, groundTruth)
	require.NoError(t, err)
	assert.False(t, report.PassedGates, "Substandard model must fail quality gates")
	assert.NotEmpty(t, report.GateViolations)
}

func TestMLOps_DataQualityMonitor(t *testing.T) {
	monitor := NewDataQualityMonitor()

	// High quality batch
	cleanRecords := []map[string]interface{}{
		{"transaction_id": "txn_1", "amount": 100.0, "user_id": "u1"},
		{"transaction_id": "txn_2", "amount": 250.0, "user_id": "u2"},
	}
	cleanMetrics, err := monitor.EvaluateBatch(cleanRecords, 0.05)
	require.NoError(t, err)
	assert.Equal(t, "EXCELLENT", cleanMetrics.Status)
	assert.Equal(t, float64(0), cleanMetrics.MissingnessRate)

	// Degraded batch with missing fields and high PSI drift
	degradedRecords := []map[string]interface{}{
		{"transaction_id": "txn_1", "amount": nil, "user_id": ""},
		{"transaction_id": "txn_2", "amount": nil, "user_id": nil},
	}
	degradedMetrics, err := monitor.EvaluateBatch(degradedRecords, 0.40)
	require.NoError(t, err)
	assert.Equal(t, "DEGRADED", degradedMetrics.Status)
	assert.Greater(t, degradedMetrics.MissingnessRate, 0.30)
}

func TestMLOps_ExperimentTracker(t *testing.T) {
	tracker := NewExperimentTracker()

	run1 := &ExperimentRun{
		ExperimentID: "exp_fraud_hyperopt",
		RunID:        "run_01",
		ModelVersion: "v1.1",
		Metrics:      EvaluationMetrics{ROCAUC: 0.88, F1Score: 0.82},
		Status:       "SUCCESS",
		StartedAt:    time.Now().Add(-10 * time.Minute),
	}
	run2 := &ExperimentRun{
		ExperimentID: "exp_fraud_hyperopt",
		RunID:        "run_02",
		ModelVersion: "v1.2",
		Metrics:      EvaluationMetrics{ROCAUC: 0.94, F1Score: 0.91},
		Status:       "SUCCESS",
		StartedAt:    time.Now().Add(-5 * time.Minute),
	}

	require.NoError(t, tracker.LogRun(run1))
	require.NoError(t, tracker.LogRun(run2))

	best, err := tracker.GetBestRun("exp_fraud_hyperopt")
	require.NoError(t, err)
	assert.Equal(t, "run_02", best.RunID)
	assert.Equal(t, 0.94, best.Metrics.ROCAUC)
}

func TestMLOps_EndToEndPipelineExecution(t *testing.T) {
	ctx := context.Background()
	reg := NewModelRegistry()
	engine := NewTrainingPipelineEngine(nil, nil, nil, NewLocalOrchestrator(), reg)

	req := PipelineRequest{
		PipelineID:      "pipe_e2e_01",
		ModelVersion:    "fraud-xgb-25f-cand-e2e",
		DatasetPath:     "/app/data/train.parquet",
		DatasetChecksum: "sha256_e2e_valid_checksum",
		Hyperparameters: map[string]string{"n_estimators": "100"},
	}

	run, report, err := engine.ExecutePipeline(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, PipelineCompleted, run.State)
	assert.True(t, report.PassedGates)

	cand, err := reg.GetModel("fraud-xgb-25f-cand-e2e")
	require.NoError(t, err)
	assert.Equal(t, LifecycleCandidate, cand.LifecycleState)
	assert.Equal(t, "sha256_e2e_valid_checksum", cand.Provenance.DatasetChecksum)
}
