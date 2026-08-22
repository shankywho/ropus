package ml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestML_InferenceEngine(t *testing.T) {
	engine := NewRealMLInferenceEngine()

	// 1. Low risk features
	cleanFeats := TransactionFeatures{
		AmountUSD:             45.0,
		Velocity10m:           1.0,
		DeviceEntropy:         0.85,
		IsEmulator:            0.0,
		IsVPN:                 0.0,
		GeoDistanceKm:         15.0,
		GraphDegreeCentrality: 0.0,
	}
	cleanRes := engine.PredictFraud(cleanFeats)
	assert.Less(t, cleanRes.FraudProbability, 0.20)
	assert.Equal(t, "fraud-xgb-v4-prod", cleanRes.ModelVersion)

	// 2. High risk features
	fraudFeats := TransactionFeatures{
		AmountUSD:             18000.0,
		Velocity10m:           12.0,
		DeviceEntropy:         0.10,
		IsEmulator:            1.0,
		IsVPN:                 1.0,
		GeoDistanceKm:         8500.0,
		GraphDegreeCentrality: 14.0,
	}
	fraudRes := engine.PredictFraud(fraudFeats)
	assert.GreaterOrEqual(t, fraudRes.FraudProbability, 0.90)
	assert.Greater(t, fraudRes.ConfidenceScore, 0.90)
}
