package llm

import (
	"context"
	"testing"

	"github.com/shankywho/ropus/backend/internal/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLLM_InvestigationAgentAndRAG(t *testing.T) {
	ctx := context.Background()
	vectorStore := memory.NewVectorStore()
	llmClient := NewLLMClient("", "", "claude-3-7-sonnet-20250219")
	agent := NewLLMInvestigationAgent(llmClient, vectorStore)

	report, err := agent.InvestigateCase(ctx, "case_test_99", "usr_mule_cluster", "tx_9981")
	require.NoError(t, err)

	assert.NotEmpty(t, report.ReportID)
	assert.NotEmpty(t, report.FraudExplanation)
	assert.Equal(t, 3, len(report.ToolTraces))
	assert.GreaterOrEqual(t, len(report.SimilarPrecedents), 1)
	assert.GreaterOrEqual(t, report.Confidence, 0.95)
}
