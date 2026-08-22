package demo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/shankywho/ropus/backend/internal/auth/api_keys"
	"github.com/shankywho/ropus/backend/internal/product_api"
)

func TestCanonicalScenario_17Stages(t *testing.T) {
	keyService := api_keys.NewAPIKeyService()
	pipeline := product_api.NewUnifiedRiskPipeline(keyService, nil, nil)
	runner := NewCanonicalScenarioRunner(pipeline)

	keyResp, err := keyService.GenerateKey("org_canonical_demo", "Canonical Attack Key", "live", []string{"risk:evaluate"}, 90)
	require.NoError(t, err)

	res, err := runner.ExecuteCanonicalAttack(context.Background(), keyResp.PlaintextKey)
	require.NoError(t, err)

	assert.Equal(t, 17, res.TotalStages)
	assert.Equal(t, "BLOCK", res.DecisionResult.Decision)
	assert.GreaterOrEqual(t, res.DecisionResult.RiskScore, 0.80)
	assert.NotEmpty(t, res.DecisionResult.CaseID)
}
