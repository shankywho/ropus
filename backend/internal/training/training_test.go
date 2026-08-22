package training

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTraining_PipelineExecution(t *testing.T) {
	pipeline := NewMLTrainingPipeline()

	cfg := TrainingJobConfig{
		DatasetPath:      "datasets/transactions.csv",
		Algorithm:        "XGBOOST",
		TestSplitRatio:   0.20,
		MinimumTargetAUC: 0.95,
	}

	artifact, err := pipeline.RunTrainingJob(cfg)
	require.NoError(t, err)

	assert.NotEmpty(t, artifact.ModelID)
	assert.Equal(t, "APPROVED_FOR_PRODUCTION", artifact.ApprovalStatus)
	assert.GreaterOrEqual(t, artifact.Metrics.ROCAUC, 0.98)
	assert.GreaterOrEqual(t, artifact.Metrics.Precision, 0.95)
	assert.GreaterOrEqual(t, artifact.Metrics.Recall, 0.90)
	assert.Less(t, artifact.Metrics.PSIDrift, 0.05)
}
