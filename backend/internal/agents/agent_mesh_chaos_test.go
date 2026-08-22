package agents

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgents_MessageBusAndCommunication(t *testing.T) {
	bus := NewLocalAgentBus()

	var received []*AgentMessage
	err := bus.Subscribe(AgentFraudInvestigator, func(msg *AgentMessage) error {
		received = append(received, msg)
		return nil
	})
	require.NoError(t, err)

	msg := &AgentMessage{
		MessageID:   "msg_01",
		TraceID:     "trc_01",
		SourceAgent: "agent_orchestrator",
		TargetAgent: AgentFraudInvestigator,
		Type:        MsgTask,
	}
	err = bus.Publish(msg)
	require.NoError(t, err)
	assert.Equal(t, 1, len(received))
}

func TestAgents_DualTierMemory(t *testing.T) {
	mem := NewAgentMemory()

	// Short term working memory
	mem.RememberShortTerm("trc_100", "alice@example.com", "LOGIN_FROM_NEW_DEVICE", "dev_xyz", 0.85)
	facts := mem.RecallShortTerm("trc_100")
	require.Equal(t, 1, len(facts))
	assert.NotContains(t, facts[0].Subject, "alice@example.com", "Subject must be hashed to prevent PII leakage")

	// Long term collective pattern memory
	mem.RememberLongTerm("pattern_proxy_carding_v2", 0.96)
	conf, found := mem.RecallLongTerm("pattern_proxy_carding_v2")
	assert.True(t, found)
	assert.Equal(t, 0.96, conf)
}

func TestAgents_ReasoningEngineAndExplainer(t *testing.T) {
	engine := NewFraudReasoningEngine()
	explainer := NewDecisionExplainer()

	hypo := engine.Reason(0.92, 5, []string{"Velocity surge 10x"}, []string{"Tor exit node IP"}, 1)
	assert.Equal(t, "BLOCK_AND_FREEZE", hypo.RecommendedAction)
	assert.GreaterOrEqual(t, hypo.Confidence, 0.90)
	assert.NotEmpty(t, hypo.SupportingEvidence)

	explanation := explainer.Explain(hypo)
	assert.Equal(t, "BLOCK_AND_FREEZE", explanation.Decision)
	assert.NotEmpty(t, explanation.GovernanceComplianceNote)
}

func TestAgents_AutonomousInvestigatorDossier(t *testing.T) {
	mem := NewAgentMemory()
	investigator := NewAutonomousInvestigatorAgent(mem)

	dossier := investigator.InvestigateIncident(
		"trc_investigate_01", "usr_fraud_syndicate_node",
		0.94, 8,
		[]string{"Device fingerprint mismatch"},
		[]string{"Compromised proxy list match"},
	)

	assert.NotEmpty(t, dossier.DossierID)
	assert.Equal(t, "BLOCK_AND_FREEZE", dossier.ReasoningHypothesis.RecommendedAction)
	assert.Equal(t, "agent_auto_investigator_v1", dossier.AssignedAgentID)
}

func TestAgents_ThreatHunterAndOptimizer(t *testing.T) {
	hunter := NewAIThreatHunterAgent(nil)
	hypo := hunter.HuntSweep(150, []string{"cluster_bad_mules_01"})
	assert.NotEmpty(t, hypo.RecommendedRuleDSL)
	assert.GreaterOrEqual(t, hypo.Confidence, 0.90)

	// Optimizer
	opt := NewRiskOptimizerAgent()
	proposal := opt.ProposeThresholdTuning("velocity_limit_5m", 10.0, 0.02)
	assert.Equal(t, StageProposal, proposal.Stage)

	opt.AdvanceStage(proposal)
	assert.Equal(t, StageGovernanceReview, proposal.Stage)
	opt.AdvanceStage(proposal)
	assert.Equal(t, StageShadowTesting, proposal.Stage)
	opt.AdvanceStage(proposal)
	assert.Equal(t, StageCanaryRollout, proposal.Stage)
	opt.AdvanceStage(proposal)
	assert.Equal(t, StageProductionActive, proposal.Stage)
}

func TestAgents_PolicyReasoningAndResponsePlanner(t *testing.T) {
	policyAgent := NewPolicyReasoningAgent()
	outcome := policyAgent.EvaluateActionCompliance("BLOCK_AND_FREEZE", 2, 0.95)
	assert.True(t, outcome.IsAllowed)
	assert.False(t, outcome.RequiresHumanSignOff)

	// Excessive blast radius -> requires human review
	largeOutcome := policyAgent.EvaluateActionCompliance("BLOCK_AND_FREEZE", 800, 0.95)
	assert.True(t, largeOutcome.RequiresHumanSignOff)
	assert.Equal(t, "HIGH", largeOutcome.CustomerImpactLevel)

	// Planner
	planner := NewAutonomousResponsePlanner(policyAgent)
	plan := planner.PlanOptimalResponse(5000.0, 0.95, 2)
	assert.Equal(t, "BLOCK_DEVICE_AND_ACCOUNT", plan.SelectedAction)
	assert.Equal(t, 3, len(plan.OptionsEvaluated))
}

func TestAgents_AttackSimulatorAndGuardrails(t *testing.T) {
	sim := NewAttackSimulationAgent()
	attack := sim.GenerateAttackScenario("CARDING_BURST", 50)
	assert.NotEmpty(t, attack.SimulationID)
	assert.Equal(t, 50, attack.TargetNodeCount)

	// Guardrails
	guard := NewAgentSecurityGuardrails()
	err := guard.ValidateAgentAction(AgentFraudInvestigator, "BLOCK_AND_FREEZE", 0.95, []string{"Valid evidence item"})
	require.NoError(t, err)

	// Hallucination rejection: Empty evidence chain -> BLOCKED
	err = guard.ValidateAgentAction(AgentFraudInvestigator, "BLOCK_AND_FREEZE", 0.95, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty evidence chain")

	// Low confidence irreversible action -> BLOCKED
	err = guard.ValidateAgentAction(AgentFraudInvestigator, "BLOCK_AND_FREEZE", 0.40, []string{"Some evidence"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requires >= 0.85 confidence")
}

func TestAgents_Observability(t *testing.T) {
	obs := NewAgentObservabilityPlatform()
	obs.RecordExecution("trc_01", AgentFraudInvestigator, 2.5, 0.94, "SUCCESS")
	obs.RecordExecution("trc_02", AgentThreatHunter, 4.0, 0.88, "SUCCESS")
	obs.RecordHumanOverride()

	total, avgMs, overrides := obs.GetMetricsSummary()
	assert.Equal(t, 2, total)
	assert.Greater(t, avgMs, 0.0)
	assert.Equal(t, 1, overrides)
}
