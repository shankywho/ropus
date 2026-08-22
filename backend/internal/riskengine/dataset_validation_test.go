package riskengine

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatasetValidator_ValidMetadata(t *testing.T) {
	validator := NewDatasetValidator(100)
	ctx := context.Background()

	meta := TrainingDatasetMetadata{
		DatasetID:          "dataset_valid_01",
		SampleCount:        500,
		FeatureContract:    "fraud-risk-25f-v2.5",
		ParentModelVersion: "fraud-xgb-25f-v3.0",
		DataQualityScore:   0.95,
		MissingValueRate:   0.02,
		ZeroLabelsDetected: false,
	}

	res, err := validator.ValidateDatasetMetadata(ctx, meta)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.True(t, res.Passed)
	assert.Empty(t, res.Violations)
}

func TestDatasetValidator_Violations(t *testing.T) {
	validator := NewDatasetValidator(200)
	ctx := context.Background()

	// 1. Below sample quorum
	res1, err1 := validator.ValidateDatasetMetadata(ctx, TrainingDatasetMetadata{
		DatasetID:          "dataset_low_samples",
		SampleCount:        50, // < 200
		FeatureContract:    "fraud-risk-25f-v2.5",
		DataQualityScore:   0.90,
		MissingValueRate:   0.01,
		ZeroLabelsDetected: false,
	})
	assert.Error(t, err1)
	assert.False(t, res1.Passed)
	assert.Contains(t, res1.Violations[0], "below minimum quorum")

	// 2. Incompatible contract
	res2, err2 := validator.ValidateDatasetMetadata(ctx, TrainingDatasetMetadata{
		DatasetID:          "dataset_bad_contract",
		SampleCount:        300,
		FeatureContract:    "legacy-10f-v1.0",
		DataQualityScore:   0.90,
		MissingValueRate:   0.01,
		ZeroLabelsDetected: false,
	})
	assert.Error(t, err2)
	assert.False(t, res2.Passed)
	assert.Contains(t, res2.Violations[0], "Incompatible feature contract")

	// 3. Zero positive fraud labels
	res3, err3 := validator.ValidateDatasetMetadata(ctx, TrainingDatasetMetadata{
		DatasetID:          "dataset_zero_labels",
		SampleCount:        300,
		FeatureContract:    "fraud-risk-25f-v2.5",
		DataQualityScore:   0.90,
		MissingValueRate:   0.01,
		ZeroLabelsDetected: true,
	})
	assert.Error(t, err3)
	assert.False(t, res3.Passed)
	assert.Contains(t, res3.Violations[0], "zero positive fraud labels")

	// 4. Excessive missing value rate
	res4, err4 := validator.ValidateDatasetMetadata(ctx, TrainingDatasetMetadata{
		DatasetID:          "dataset_high_missing",
		SampleCount:        300,
		FeatureContract:    "fraud-risk-25f-v2.5",
		DataQualityScore:   0.90,
		MissingValueRate:   0.25, // > 0.15 limit
		ZeroLabelsDetected: false,
	})
	assert.Error(t, err4)
	assert.False(t, res4.Passed)
	assert.Contains(t, res4.Violations[0], "Missing value rate")
}

func TestDatasetValidator_FileValidation(t *testing.T) {
	validator := NewDatasetValidator(50)
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test_dataset.csv")
	err := os.WriteFile(testFile, []byte("amount,ip_velocity_1h,token_velocity_24h,is_fraud\n100,1,2,0\n500,4,5,1\n"), 0644)
	require.NoError(t, err)

	checksum, size, err := validator.ValidateDatasetFile(testFile)
	require.NoError(t, err)
	assert.NotEmpty(t, checksum)
	assert.Greater(t, size, int64(0))

	// Non-existent file
	_, _, err = validator.ValidateDatasetFile(filepath.Join(tmpDir, "non_existent.csv"))
	assert.Error(t, err)
}

func TestDatasetValidator_SchemaJSON(t *testing.T) {
	validator := NewDatasetValidator(50)

	validJSON := `{
		"dataset_id": "ds_json_01",
		"sample_count": 250,
		"feature_contract": "fraud-risk-25f-v2.5",
		"positive_fraud_count": 15,
		"missing_value_rate": 0.02,
		"data_quality_score": 0.94,
		"feature_names": [
			"f1","f2","f3","f4","f5","f6","f7","f8","f9","f10",
			"f11","f12","f13","f14","f15","f16","f17","f18","f19","f20",
			"f21","f22","f23","f24","f25"
		]
	}`

	res, err := validator.ValidateDatasetSchemaJSON([]byte(validJSON))
	require.NoError(t, err)
	assert.True(t, res.Passed)

	// Wrong feature count (24 features)
	invalidJSON := `{
		"dataset_id": "ds_json_bad",
		"sample_count": 250,
		"feature_contract": "fraud-risk-25f-v2.5",
		"positive_fraud_count": 15,
		"missing_value_rate": 0.02,
		"data_quality_score": 0.94,
		"feature_names": [
			"f1","f2","f3","f4","f5","f6","f7","f8","f9","f10",
			"f11","f12","f13","f14","f15","f16","f17","f18","f19","f20",
			"f21","f22","f23","f24"
		]
	}`

	resBad, errBad := validator.ValidateDatasetSchemaJSON([]byte(invalidJSON))
	assert.Error(t, errBad)
	assert.False(t, resBad.Passed)
}

func TestCheckNumericalSanity(t *testing.T) {
	cleanVec := []float64{1.0, 2.5, 0.0, 100.2}
	hasNaN, hasInf := CheckNumericalSanity(cleanVec)
	assert.False(t, hasNaN)
	assert.False(t, hasInf)

	nanVec := []float64{1.0, math.NaN(), 3.0}
	hasNaN2, _ := CheckNumericalSanity(nanVec)
	assert.True(t, hasNaN2)
}
