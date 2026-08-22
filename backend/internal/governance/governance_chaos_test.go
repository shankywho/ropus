package governance

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGovernance_UnauthorizedModelApprovalBlocked(t *testing.T) {
	mgr := NewModelRiskManager()

	candidate := &GovernanceModelRecord{
		ModelID:  "cand_unverified_01",
		Version:  "fraud-xgb-25f-cand-unverified",
		RiskTier: RiskTier1High,
		ApprovalChain: ApprovalChainStatus{
			ValidationPassed:     false, // Failed validation!
			ExplainabilityPassed: true,
			FairnessPassed:       true,
			SecurityReviewPassed: true,
		},
	}
	require.NoError(t, mgr.RegisterModel(candidate))

	// Attempt approval without passed validation
	err := mgr.ApproveModel("fraud-xgb-25f-cand-unverified", "malicious_or_accidental_actor")
	assert.Error(t, err, "Model without passed validation gate must NOT be approved")
	assert.Contains(t, err.Error(), "validation gate not passed")

	// Ensure lifecycle remains non-approved
	m, err := mgr.GetModel("fraud-xgb-25f-cand-unverified")
	require.NoError(t, err)
	assert.False(t, m.ApprovalChain.IsFullyApproved)
	assert.NotEqual(t, StateApproved, m.LifecycleState)
}

func TestGovernance_AuditTrailTamperingDetection(t *testing.T) {
	trail := NewDecisionAuditTrail()

	trail.AppendDecision("dec_01", "req_hash_1", "v3.0", "contract_v2", "ALLOW", "exp_01", "pol_v1", 0.12)
	trail.AppendDecision("dec_02", "req_hash_2", "v3.0", "contract_v2", "BLOCK", "exp_02", "pol_v1", 0.94)
	trail.AppendDecision("dec_03", "req_hash_3", "v3.0", "contract_v2", "MANUAL_REVIEW", "exp_03", "pol_v1", 0.76)

	// Verify pristine chain
	valid, err := trail.VerifyIntegrity()
	require.NoError(t, err)
	assert.True(t, valid)

	// Tamper with record 1 payload (simulate malicious DB update)
	trail.records[1].RiskScore = 0.10 // Altered score

	// Verification MUST detect tampering
	valid, err = trail.VerifyIntegrity()
	assert.False(t, valid, "Audit chain must detect tampered record")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audit record tampered at index 1")
}

func TestGovernance_PolicyEngineDeploymentAndRollback(t *testing.T) {
	pe := NewPolicyEngine()

	// Initial baseline policy
	override, ver, triggered := pe.EvaluatePolicies(0.96, 50.0)
	assert.True(t, triggered)
	assert.Equal(t, "BLOCK", override)
	assert.Equal(t, "policy_v1.0_standard", ver)

	// Deploy new policy v2.0
	p2 := &GovernancePolicy{
		PolicyVersion: "policy_v2.0_experimental",
		Author:        "lead_risk_officer",
		Rules: []PolicyRule{
			{
				RuleID:         "rule_strict_block",
				MinScore:       0.50,
				MinAmount:      0.0,
				ActionOverride: "BLOCK",
				Enabled:        true,
			},
		},
	}
	require.NoError(t, pe.DeployPolicy(p2))

	override, ver, triggered = pe.EvaluatePolicies(0.60, 50.0)
	assert.True(t, triggered)
	assert.Equal(t, "BLOCK", override)
	assert.Equal(t, "policy_v2.0_experimental", ver)

	// Rollback policy safely
	prevVer, err := pe.RollbackPolicy()
	require.NoError(t, err)
	assert.Equal(t, "policy_v1.0_standard", prevVer)

	// Confirm rollback active
	override, ver, triggered = pe.EvaluatePolicies(0.60, 50.0)
	assert.False(t, triggered, "Rolled back policy should not trigger at 0.60")
	assert.Equal(t, "policy_v1.0_standard", ver)
}

func TestGovernance_ExplainabilityAndHumanReview(t *testing.T) {
	expEngine := NewExplainabilityEngine()
	features := map[string]float64{
		"amount":            15000.0,
		"user_txn_count_1h": 6.0,
		"device_age_days":   0.2,
	}

	exp := expEngine.ExplainDecision("dec_test_01", "BLOCK", 0.94, features)
	assert.Equal(t, "BLOCK", exp.Decision)
	assert.Equal(t, 0.94, exp.RiskScore)
	assert.GreaterOrEqual(t, len(exp.TopExplanations), 2)

	// Human Review Case
	reviewSys := NewHumanReviewSystem()
	c := reviewSys.CreateReview("txn_high_risk_01", "tenant_01", 0.94, []string{"High amount", "Velocity burst"})
	assert.Equal(t, ReviewPending, c.Status)

	require.NoError(t, reviewSys.AssignReview(c.ReviewID, "analyst_alice"))
	assert.Equal(t, ReviewAssigned, c.Status)

	require.NoError(t, reviewSys.SubmitDecision(c.ReviewID, "CONFIRM_FRAUD", "Account takeover confirmed via phone verify"))
	assert.Equal(t, ReviewRejected, c.Status)

	// Feedback Loop Ingestion
	fbLoop := NewFeedbackLearningLoop()
	outcome := fbLoop.IngestAnalystFeedback("txn_high_risk_01", 0.94, "BLOCK", "CONFIRM_FRAUD")
	assert.True(t, outcome.ConfirmedFraud)
	assert.False(t, outcome.IsFalsePositive)

	samples := fbLoop.GetGoldTrainingSamples()
	assert.Equal(t, 1, len(samples))
}

func TestGovernance_DashboardAndModelCardReport(t *testing.T) {
	mgr := NewModelRiskManager()
	audit := NewDecisionAuditTrail()
	fairness := NewFairnessMonitor()
	reviews := NewHumanReviewSystem()
	pe := NewPolicyEngine()

	dash := NewGovernanceDashboardAggregator(mgr, audit, fairness, reviews, pe)
	overview := dash.GetOverview()

	assert.GreaterOrEqual(t, overview.ActiveModelsCount, 1)
	assert.True(t, overview.FairnessCompliant)
	assert.True(t, overview.AuditChainIntegrityVerified)
	assert.Equal(t, "COMPLIANT_READY", overview.RegulatoryReadinessStatus)

	// Model Card Generation
	prodModel, err := mgr.GetModel("fraud-xgb-25f-v3.0")
	require.NoError(t, err)

	reporter := NewComplianceReportGenerator()
	doc := reporter.GenerateModelCard(prodModel)
	assert.Equal(t, "model_prod_25f_v3", doc.ModelName)

	md := reporter.RenderModelCardMarkdown(doc)
	assert.Contains(t, md, "AI Model Governance Card")
	assert.Contains(t, md, "Quantitative Performance & Validation")
}
