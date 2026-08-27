package riskengine

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shankywho/ropus/backend/internal/rules"
)

// TestProvenanceAntiSpoofing verifies that client request fields cannot forge LIVE_PRODUCTION.
// Provenance is derived strictly server-side from the verified environment.
func TestProvenanceAntiSpoofing(t *testing.T) {
	rulesService := rules.NewService(nil)

	// 1. Non-production environment (e.g. development/test): must produce OFFLINE_TEST
	devOrch := NewOrchestrator(nil, nil, rulesService, nil, nil)
	devOrch.SetEnvironment("development")

	mockScorerDev := NewShadowScorer(ShadowScorerConfig{
		Enabled:                  true,
		WorkerCount:              1,
		QueueCapacity:            10,
		SampleRate:               1.0,
		CandidateModelVersion:    "extended_catboost_58f",
		CandidateFeatureContract: MLFeatureContractV25,
	}, nil, nil)

	devOrch.SetShadowScorer(mockScorerDev)

	// Evaluate in development environment
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	respDev, errDev := devOrch.Evaluate(ctx, "tenant_test", RiskEvaluationRequest{
		TransactionID: "tx_dev_spoof_attempt",
		Amount:        5000,
		Currency:      "USD",
		PaymentMethod: PaymentMethod{Type: "CARD", Token: "tok_dev"},
		IPAddress:     "127.0.0.1",
	})
	require.NoError(t, errDev)
	assert.NotEmpty(t, respDev.DecisionID)

	// Give background worker a moment to process the task
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, int64(1), mockScorerDev.metrics.OfflineTestRequestsTotal.Load(), "Non-production environment must record OFFLINE_TEST")
	assert.Equal(t, int64(0), mockScorerDev.metrics.LiveProductionRequestsTotal.Load(), "Non-production environment must NOT record LIVE_PRODUCTION")

	// 2. Production environment: derives LIVE_PRODUCTION server-side
	prodOrch := NewOrchestrator(nil, nil, rulesService, nil, nil)
	prodOrch.SetEnvironment("production")

	mockScorerProd := NewShadowScorer(ShadowScorerConfig{
		Enabled:                  true,
		WorkerCount:              1,
		QueueCapacity:            10,
		SampleRate:               1.0,
		CandidateModelVersion:    "extended_catboost_58f",
		CandidateFeatureContract: MLFeatureContractV25,
	}, nil, nil)

	prodOrch.SetShadowScorer(mockScorerProd)

	respProd, errProd := prodOrch.Evaluate(ctx, "tenant_real_prod", RiskEvaluationRequest{
		TransactionID: "tx_prod_authentic_001",
		Amount:        9900,
		Currency:      "USD",
		PaymentMethod: PaymentMethod{Type: "CARD", Token: "tok_prod"},
		IPAddress:     "198.51.100.22",
	})
	require.NoError(t, errProd)
	assert.NotEmpty(t, respProd.DecisionID)

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, int64(1), mockScorerProd.metrics.LiveProductionRequestsTotal.Load(), "Production environment derives LIVE_PRODUCTION server-side")
	assert.Equal(t, int64(0), mockScorerProd.metrics.OfflineTestRequestsTotal.Load())
}
