package cases

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCases_CreationAndIntelligentPrioritization(t *testing.T) {
	cm := NewCaseManager(nil)

	// Low priority case
	cLow, err := cm.CreateCase("tenant_01", "usr_low", []string{"txn_1"}, 25.0, 0.20, 1, 1, false)
	require.NoError(t, err)
	assert.Equal(t, PriorityLow, cLow.Priority)
	assert.Equal(t, StatusOpen, cLow.Status)

	// Critical priority case ($100k exposure + 50 connected nodes + threat match)
	cCrit, err := cm.CreateCase("tenant_01", "usr_crit", []string{"txn_2", "txn_3"}, 120000.0, 0.95, 50, 12, true)
	require.NoError(t, err)
	assert.Equal(t, PriorityCritical, cCrit.Priority)
}

func TestCases_AutonomousInvestigationAgent(t *testing.T) {
	agent := NewAutonomousInvestigationAgent(nil)

	report, err := agent.Investigate(
		"txn_999", "usr_suspect", "dev_emul_root", "198.51.100.44",
		15000.0, 0.94, 12,
		[]string{"Spending spike 10x baseline", "Unrecognized device"},
		[]string{"IP flagged on proxy blocklist"},
	)
	require.NoError(t, err)
	assert.Equal(t, "FREEZE_ACCOUNT", report.RecommendedAction)
	assert.GreaterOrEqual(t, report.FraudProbability, 0.90)
	assert.NotEmpty(t, report.EvidenceItems)
	assert.NotEmpty(t, report.EvidenceTimeline)
}

func TestCases_ResponseEngineAndGuardrails(t *testing.T) {
	guard := NewResponseGuard()
	re := NewResponseEngine(guard)

	// Safe high confidence action -> SUCCESS
	rec, err := re.ExecuteAction(ActionFreezeAccount, "acc_fraud_01", "ACCOUNT", "Confirmed mule node", "SYSTEM", 0.95, 0.92)
	require.NoError(t, err)
	assert.False(t, rec.IsRolledBack)

	// Rollback action -> SUCCESS
	err = re.RollbackAction(rec.ActionID, "Analyst override after phone verification")
	require.NoError(t, err)

	updatedRec, _ := re.GetAction(rec.ActionID)
	assert.True(t, updatedRec.IsRolledBack)

	// Unsafe low confidence action -> GUARD BLOCKS
	_, err = re.ExecuteAction(ActionFreezeAccount, "acc_innocent_01", "ACCOUNT", "Weak suspicion", "SYSTEM", 0.40, 0.30)
	assert.Error(t, err, "Response guard must block low confidence account freeze")
	assert.Contains(t, err.Error(), "rejected: confidence")
}

func TestCases_SOARPlaybookExecution(t *testing.T) {
	re := NewResponseEngine(nil)
	cm := NewCaseManager(nil)
	soar := NewSOAREngine(re, cm)

	ctx := IncidentContext{
		IncidentID:         "inc_ring_01",
		PrimaryEntityID:    "acc_hub_99",
		AssociatedEntities: []string{"acc_mule_1", "acc_mule_2", "acc_mule_3"},
		RiskScore:          0.96,
		ThreatType:         "FRAUD_RING",
		TransactionIDs:     []string{"txn_ring_1", "txn_ring_2"},
		TotalExposure:      75000.0,
	}

	result, err := soar.ExecutePlaybook("PLAYBOOK_FRAUD_RING", ctx)
	require.NoError(t, err)
	assert.Equal(t, "COMPLETED", result.Status)
	assert.Equal(t, 4, len(result.ExecutedActions)) // 1 hub + 3 mules frozen
	assert.NotEmpty(t, result.CaseGeneratedID)
}

func TestCases_ThreatHunterAndAlerts(t *testing.T) {
	hunter := NewThreatHunter()
	sweep := hunter.RunHuntSweep(120, 5000)
	assert.Equal(t, "COMPLETED", sweep.Status)
	assert.NotEmpty(t, sweep.Findings)

	// Alert Engine
	ae := NewFraudAlertEngine()
	alert := ae.DispatchAlert("case_123", "Critical Syndicate Detected", "Mule network activated", AlertCritical)
	assert.Equal(t, AlertCritical, alert.Severity)
	assert.Contains(t, alert.Channels, "PAGERDUTY")
	assert.Equal(t, 1, len(ae.ListAlerts()))
}

func TestCases_AnalystCopilotAndDashboard(t *testing.T) {
	cm := NewCaseManager(nil)
	ae := NewFraudAlertEngine()
	copilot := NewAnalystCopilot(cm)
	dash := NewFraudOperationsDashboard(cm, ae)

	c, err := cm.CreateCase("tenant_01", "usr_test", []string{"txn_1"}, 5000.0, 0.85, 3, 2, false)
	require.NoError(t, err)

	require.NoError(t, copilot.AssignCase(c.CaseID, "analyst_bob"))
	require.NoError(t, copilot.ResolveCase(c.CaseID, true, "Account takeover confirmed", "analyst_bob"))

	resolvedCase, err := copilot.GetCaseDetails(c.CaseID)
	require.NoError(t, err)
	assert.Equal(t, StatusConfirmedFraud, resolvedCase.Status)

	summary := dash.GetSummary()
	assert.GreaterOrEqual(t, summary.FraudLossPrevented, 5000.0)
}
