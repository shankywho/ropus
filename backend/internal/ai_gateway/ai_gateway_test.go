package ai_gateway

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAIGateway_RoutingAndCostAccounting(t *testing.T) {
	ctx := context.Background()
	gw := NewAIGateway()

	req := GatewayRequest{
		Prompt:         "Analyze user velocity anomaly: 14 txs in 2 mins.",
		Provider:       ProviderAnthropic,
		PreferredModel: "claude-3-7-sonnet-20250219",
		MaxTokens:      500,
	}

	resp, err := gw.GenerateCompletion(ctx, req)
	require.NoError(t, err)

	assert.NotEmpty(t, resp.Content)
	assert.Equal(t, ProviderAnthropic, resp.ProviderUsed)
	assert.Greater(t, resp.EstimatedCostUSD, 0.0)

	tokens, cost := gw.GetGatewayStats()
	assert.Greater(t, tokens, uint64(0))
	assert.Greater(t, cost, 0.0)
}
