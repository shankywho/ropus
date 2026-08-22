package store

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFeatureRegistry_BaselinesAndCustom(t *testing.T) {
	reg := NewFeatureRegistry()
	features := reg.ListFeatures()
	assert.GreaterOrEqual(t, len(features), 15, "Baseline should contain at least 15 features")

	// Get existing feature
	amountDef, err := reg.GetFeature("amount")
	require.NoError(t, err)
	assert.Equal(t, TypeFloat64, amountDef.DataType)

	// Register new custom feature
	customDef := FeatureDefinition{
		Name:         "custom_risk_factor",
		Version:      1,
		DataType:     TypeFloat64,
		Description:  "Custom composite risk weight",
		SourceEntity: "model",
	}
	err = reg.RegisterFeature(customDef)
	require.NoError(t, err)

	fetched, err := reg.GetFeature("custom_risk_factor")
	require.NoError(t, err)
	assert.Equal(t, "custom_risk_factor", fetched.Name)

	history, err := reg.GetHistory("custom_risk_factor")
	require.NoError(t, err)
	assert.Equal(t, 1, len(history))
}

func TestFeatureDefinition_ValidateValue(t *testing.T) {
	floatDef := FeatureDefinition{Name: "score", DataType: TypeFloat64}
	assert.NoError(t, floatDef.ValidateValue(0.75))
	assert.NoError(t, floatDef.ValidateValue(100))
	assert.Error(t, floatDef.ValidateValue("invalid_string"))

	intDef := FeatureDefinition{Name: "count", DataType: TypeInt64}
	assert.NoError(t, intDef.ValidateValue(42))
	assert.Error(t, intDef.ValidateValue("bad"))

	boolDef := FeatureDefinition{Name: "is_fraud", DataType: TypeBool}
	assert.NoError(t, boolDef.ValidateValue(true))
	assert.Error(t, boolDef.ValidateValue(123))
}

func TestOnlineStore_PutAndGet(t *testing.T) {
	ctx := context.Background()
	store := NewInMemoryOnlineStore()

	err := store.PutOnlineFeatures(ctx, "user_123", map[string]interface{}{
		"user_txn_count_1h": 5,
		"user_txn_sum_1h":   1500.50,
	}, 1*time.Hour)
	require.NoError(t, err)

	vals, err := store.GetOnlineFeatures(ctx, "user_123", []string{"user_txn_count_1h", "user_txn_sum_1h", "non_existent"})
	require.NoError(t, err)
	assert.Equal(t, 5, vals["user_txn_count_1h"])
	assert.Equal(t, 1500.50, vals["user_txn_sum_1h"])
	assert.NotContains(t, vals, "non_existent")
}

func TestOfflineStore_Snapshot(t *testing.T) {
	ctx := context.Background()
	offline := NewInMemoryOfflineStore()

	records := []map[string]interface{}{
		{"amount": 100.0, "user_txn_count_1h": 1, "is_fraud": 0},
		{"amount": 500.0, "user_txn_count_1h": 8, "is_fraud": 1},
	}

	snap, err := NewFeatureSnapshot("snap_2026_01", "transaction", []string{"amount", "user_txn_count_1h", "is_fraud"}, records)
	require.NoError(t, err)
	assert.NotEmpty(t, snap.ChecksumSHA256)
	assert.Equal(t, 2, snap.SampleCount)

	err = offline.SaveSnapshot(ctx, snap)
	require.NoError(t, err)

	loaded, err := offline.GetSnapshot(ctx, "snap_2026_01")
	require.NoError(t, err)
	assert.Equal(t, snap.ChecksumSHA256, loaded.ChecksumSHA256)
}
