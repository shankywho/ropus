package riskengine

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFixtureTrainingAdapter(t *testing.T) {
	runner := NewFixtureTrainingAdapter()
	ctx := context.Background()

	meta := TrainingDatasetMetadata{
		DatasetID:          "dataset_fix_01",
		SampleCount:        200,
		FeatureContract:    "fraud-risk-25f-v2.5",
		ParentModelVersion: "fraud-xgb-25f-v3.0",
		DataQualityScore:   0.95,
		MissingValueRate:   0.01,
	}

	req := TrainingRequest{
		JobID:              "job_fix_01",
		ParentModelVersion: "fraud-xgb-25f-v3.0",
		FeatureContract:    "fraud-risk-25f-v2.5",
		DatasetMetadata:    meta,
		TriggerReason:      "Sustained drift test",
		Actor:              "TEST_RUNNER",
	}

	job, err := runner.StartTraining(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, job)
	assert.Equal(t, TrainingJobSucceeded, job.State)
	assert.NotEmpty(t, job.Candidate.Version)
	assert.NotEmpty(t, job.ArtifactChecksum)
	assert.Greater(t, job.ValidationMetrics.ROCAUC, 0.90)

	// Status query
	status, err := runner.GetTrainingStatus(ctx, job.JobID)
	require.NoError(t, err)
	assert.Equal(t, TrainingJobSucceeded, status.State)

	// Cancel
	err = runner.CancelTraining(ctx, job.JobID)
	require.NoError(t, err)
	statusCancelled, _ := runner.GetTrainingStatus(ctx, job.JobID)
	assert.Equal(t, TrainingJobCancelled, statusCancelled.State)
}

func TestLocalProcessTrainingAdapter_DatasetValidationFailure(t *testing.T) {
	cfg := DefaultLocalProcessConfig()
	adapter := NewLocalProcessTrainingAdapter(cfg)
	ctx := context.Background()

	// Insufficient samples -> should fail immediately without spawning process
	badMeta := TrainingDatasetMetadata{
		DatasetID:          "dataset_too_small",
		SampleCount:        10, // < 50
		FeatureContract:    "fraud-risk-25f-v2.5",
		DataQualityScore:   0.90,
		MissingValueRate:   0.01,
	}

	req := TrainingRequest{
		JobID:              "job_fail_ds",
		ParentModelVersion: "fraud-xgb-25f-v3.0",
		FeatureContract:    "fraud-risk-25f-v2.5",
		DatasetMetadata:    badMeta,
	}

	_, err := adapter.StartTraining(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "below minimum quorum")
}

func TestLocalProcessTrainingAdapter_ExecutionAndCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := LocalProcessConfig{
		Command:     "sleep",
		Args:        []string{"10"},
		OutputDir:   tmpDir,
		Timeout:     5 * time.Second,
		MaxLogBytes: 4096,
	}

	adapter := NewLocalProcessTrainingAdapter(cfg)
	ctx := context.Background()

	meta := TrainingDatasetMetadata{
		DatasetID:          "dataset_proc_01",
		SampleCount:        300,
		FeatureContract:    "fraud-risk-25f-v2.5",
		ParentModelVersion: "fraud-xgb-25f-v3.0",
		DataQualityScore:   0.95,
		MissingValueRate:   0.01,
	}

	req := TrainingRequest{
		JobID:              "job_cancel_test",
		ParentModelVersion: "fraud-xgb-25f-v3.0",
		FeatureContract:    "fraud-risk-25f-v2.5",
		DatasetMetadata:    meta,
	}

	job, err := adapter.StartTraining(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, TrainingJobRunning, job.State)

	// Cancel job
	time.Sleep(20 * time.Millisecond)
	err = adapter.CancelTraining(ctx, job.JobID)
	require.NoError(t, err)

	status, err := adapter.GetTrainingStatus(ctx, job.JobID)
	require.NoError(t, err)
	assert.Equal(t, TrainingJobCancelled, status.State)
	assert.Contains(t, status.Error, "Cancelled by operator")
}

func TestLocalProcessTrainingAdapter_Timeout(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := LocalProcessConfig{
		Command:     "sleep",
		Args:        []string{"5"},
		OutputDir:   tmpDir,
		Timeout:     50 * time.Millisecond, // Strict short timeout
		MaxLogBytes: 4096,
	}

	adapter := NewLocalProcessTrainingAdapter(cfg)
	ctx := context.Background()

	meta := TrainingDatasetMetadata{
		DatasetID:          "dataset_timeout_01",
		SampleCount:        300,
		FeatureContract:    "fraud-risk-25f-v2.5",
		ParentModelVersion: "fraud-xgb-25f-v3.0",
		DataQualityScore:   0.95,
		MissingValueRate:   0.01,
	}

	req := TrainingRequest{
		JobID:              "job_timeout_test",
		ParentModelVersion: "fraud-xgb-25f-v3.0",
		FeatureContract:    "fraud-risk-25f-v2.5",
		DatasetMetadata:    meta,
	}

	job, err := adapter.StartTraining(ctx, req)
	require.NoError(t, err)

	// Wait for timeout
	time.Sleep(100 * time.Millisecond)

	status, err := adapter.GetTrainingStatus(ctx, job.JobID)
	require.NoError(t, err)
	assert.Equal(t, TrainingJobFailed, status.State)
	assert.Contains(t, status.Error, "timed out")
}

func TestLocalProcessTrainingAdapter_NonZeroExit(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := LocalProcessConfig{
		Command:     "sh",
		Args:        []string{"-c", "echo 'Training script syntax error' >&2; exit 2"},
		OutputDir:   tmpDir,
		Timeout:     5 * time.Second,
		MaxLogBytes: 4096,
	}

	adapter := NewLocalProcessTrainingAdapter(cfg)
	ctx := context.Background()

	meta := TrainingDatasetMetadata{
		DatasetID:          "dataset_err_01",
		SampleCount:        300,
		FeatureContract:    "fraud-risk-25f-v2.5",
		ParentModelVersion: "fraud-xgb-25f-v3.0",
		DataQualityScore:   0.95,
		MissingValueRate:   0.01,
	}

	req := TrainingRequest{
		JobID:              "job_err_test",
		ParentModelVersion: "fraud-xgb-25f-v3.0",
		FeatureContract:    "fraud-risk-25f-v2.5",
		DatasetMetadata:    meta,
	}

	_, err := adapter.StartTraining(ctx, req)
	require.NoError(t, err)

	var status *TrainingJob
	for i := 0; i < 50; i++ {
		time.Sleep(20 * time.Millisecond)
		status, err = adapter.GetTrainingStatus(ctx, req.JobID)
		if err == nil && status.State != TrainingJobRunning {
			break
		}
	}
	require.NoError(t, err)
	assert.Equal(t, TrainingJobFailed, status.State)
	assert.Contains(t, status.Logs, "syntax error")
}

func init() {
	_ = os.Setenv("GO_TEST_MODE", "1")
}
