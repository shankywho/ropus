package riskengine

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModelRegistry_InitialState(t *testing.T) {
	registry := NewModelRegistry()

	prod, err := registry.GetProductionModel()
	require.NoError(t, err)
	assert.Equal(t, "fraud-xgb-25f-v3.0", prod.Version)
	assert.True(t, prod.IsProductionActive)

	fb, err := registry.GetFallbackModel()
	require.NoError(t, err)
	assert.Equal(t, "fraud-xgb-15f-v1.5", fb.Version)
	assert.True(t, fb.IsFallbackActive)

	models := registry.ListModels()
	assert.GreaterOrEqual(t, len(models), 2)
}

func TestModelRegistry_LifecycleTransitions(t *testing.T) {
	registry := NewModelRegistry()

	cand := ModelCandidate{
		ModelID:            "model_cand_reg_01",
		Version:            "fraud-xgb-25f-v3.1-candidate-01",
		ParentModelVersion: "fraud-xgb-25f-v3.0",
		FeatureContract:    "fraud-risk-25f-v2.5",
		CalibrationVersion: "beta-calibrated-v2.5",
		TrainingJobID:      "job_01",
		DatasetID:          "dataset_01",
	}

	// 1. Register candidate
	err := registry.RegisterCandidate(cand, "file:///app/model/candidates/cand.onnx", "sha256_checksum_01")
	require.NoError(t, err)

	m, err := registry.GetModel(cand.Version)
	require.NoError(t, err)
	assert.Equal(t, LifecycleCandidate, m.LifecycleState)

	// 2. Duplicate registration rejected
	errDup := registry.RegisterCandidate(cand, "file:///app/model/candidates/cand.onnx", "sha256_checksum_01")
	assert.Error(t, errDup)

	// 3. Advance to VALIDATED
	err = registry.UpdateLifecycleState(cand.Version, LifecycleValidated, "Passed offline gates")
	require.NoError(t, err)
	mVal, _ := registry.GetModel(cand.Version)
	assert.Equal(t, LifecycleValidated, mVal.LifecycleState)
	assert.Equal(t, "PASSED", mVal.ValidationStatus)

	// 4. Advance to SHADOW
	err = registry.UpdateLifecycleState(cand.Version, LifecycleShadow, "Passed shadow scoring")
	require.NoError(t, err)
	mShadow, _ := registry.GetModel(cand.Version)
	assert.Equal(t, LifecycleShadow, mShadow.LifecycleState)
	assert.Equal(t, "PASSED", mShadow.ShadowStatus)

	// 5. Promote Model: Candidate becomes Primary, previous primary becomes Fallback
	err = registry.PromoteModel(cand.Version, "ADMIN_TEST", "Canary 100% completed")
	require.NoError(t, err)

	newProd, _ := registry.GetProductionModel()
	assert.Equal(t, cand.Version, newProd.Version)
	assert.True(t, newProd.IsProductionActive)
	assert.False(t, newProd.IsFallbackActive)

	newFallback, _ := registry.GetFallbackModel()
	assert.Equal(t, "fraud-xgb-25f-v3.0", newFallback.Version)
	assert.False(t, newFallback.IsProductionActive)
	assert.True(t, newFallback.IsFallbackActive)
}

func TestModelRegistry_Rollback(t *testing.T) {
	registry := NewModelRegistry()

	cand := ModelCandidate{
		ModelID:            "model_cand_rollback",
		Version:            "fraud-xgb-25f-v3.1-candidate-rb",
		ParentModelVersion: "fraud-xgb-25f-v3.0",
		FeatureContract:    "fraud-risk-25f-v2.5",
	}

	err := registry.RegisterCandidate(cand, "file:///cand.onnx", "sha256_cand")
	require.NoError(t, err)

	err = registry.RollbackModel(cand.Version, "ADMIN_TEST", "Circuit breaker tripped")
	require.NoError(t, err)

	m, _ := registry.GetModel(cand.Version)
	assert.Equal(t, LifecycleRolledBack, m.LifecycleState)
	assert.False(t, m.IsProductionActive)
}

func TestModelRegistry_ConcurrentAccess(t *testing.T) {
	registry := NewModelRegistry()
	var wg sync.WaitGroup

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				_ = registry.ListModels()
				_, _ = registry.GetProductionModel()
				_, _ = registry.GetFallbackModel()
			}
		}(i)
	}

	wg.Wait()
}
