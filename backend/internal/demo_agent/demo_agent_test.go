package demo_agent

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDemoAgent_ShowcaseInvestigation(t *testing.T) {
	ctx := context.Background()
	engine := NewInvestigationShowcaseEngine()

	report, err := engine.RunShowcaseInvestigation(ctx, "tx_showcase_01", "usr_mule_ring_99", 48500.0)
	require.NoError(t, err)

	assert.NotEmpty(t, report.ReportID)
	assert.Equal(t, 48500.0, report.Amount)
	assert.Equal(t, "AUTONOMOUS_BLOCK", report.DecisionActionTaken)
	assert.GreaterOrEqual(t, len(report.EvidenceItems), 3)
	assert.NotEmpty(t, report.AutonomousReasoning)
}
