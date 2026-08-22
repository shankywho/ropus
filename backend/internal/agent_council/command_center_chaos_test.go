package agent_council

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCouncil_DebateAndConsensus(t *testing.T) {
	debate := NewDebateEngine()

	// 1. Unanimous Consensus
	opinionsUnanimous := []AgentOpinion{
		{AgentID: "inv_1", Role: "INVESTIGATOR", RecommendedAction: "BLOCK_AND_FREEZE", Confidence: 0.95},
		{AgentID: "threat_1", Role: "THREAT_HUNTER", RecommendedAction: "BLOCK_AND_FREEZE", Confidence: 0.96},
		{AgentID: "resp_1", Role: "RESPONSE", RecommendedAction: "BLOCK_AND_FREEZE", Confidence: 0.92},
	}
	dec1 := debate.Deliberate("inc_001", opinionsUnanimous)
	assert.Equal(t, "BLOCK_AND_FREEZE", dec1.ConsensusAction)
	assert.Equal(t, ConsensusUnanimous, dec1.ConsensusType)
	assert.Equal(t, 3, len(dec1.SupportingAgents))

	// 2. Compliance Dissent -> Escalation
	opinionsDissent := []AgentOpinion{
		{AgentID: "inv_1", Role: "INVESTIGATOR", RecommendedAction: "HARD_BLOCK", Confidence: 0.90},
		{AgentID: "comp_1", Role: "COMPLIANCE", RecommendedAction: "STEP_UP_MFA", Confidence: 0.95, DissentReason: "Severe customer impact on verified merchant"},
	}
	dec2 := debate.Deliberate("inc_002", opinionsDissent)
	assert.Equal(t, ConsensusDissentRequiresHumanReview, dec2.ConsensusType)
	assert.Contains(t, dec2.Rationale, "Compliance objection")
}

func TestCouncil_ForecastingAndSimulationLab(t *testing.T) {
	forecaster := NewFraudForecastingEngine()
	forecast := forecaster.ForecastTrajectory(0.75, 100000.0)
	assert.Equal(t, "EMULATOR_FARM_SURGE", forecast.AttackType)
	assert.Greater(t, forecast.PredictedExposure, 200000.0)

	// Simulation Lab
	lab := NewDigitalCrimeSimulationLab()
	sim := lab.SimulateAdversarialDuel("CARDING_BOT_BURST", "ENFORCE_STEP_UP_MFA", 500, 200.0, 0.95)
	assert.Greater(t, sim.LossPrevented, 80000.0)
	assert.Greater(t, sim.NetUtilityScore, 0.0)
}

func TestCouncil_DefenseStrategist(t *testing.T) {
	strategist := NewAutonomousDefenseStrategist()
	plan := strategist.SelectOptimalStrategy(50000.0, 0.95, 4)
	assert.NotEmpty(t, plan.SelectedStrategy)
	assert.Equal(t, 4, len(plan.EvaluatedOptions))
}

func TestCouncil_IntelligenceGraphAndMemory(t *testing.T) {
	graph := NewGlobalIntelligenceGraph2()
	grp := graph.RecordThreatGroup("Syndicate_Phantom_Carder", "camp_001", []string{"T1001_EMULATOR", "T1002_PROXY"})
	assert.NotEmpty(t, grp.GroupID)

	queried, found := graph.QueryGroup("Syndicate_Phantom_Carder")
	assert.True(t, found)
	assert.Equal(t, grp.GroupID, queried.GroupID)

	// Collective Memory
	mem := NewCollectiveAgentMemory()
	entry := mem.StoreEpisodicOutcome("inc_500", "SUCCESSFUL_DEFENSE", "BLOCK_AND_FREEZE", "Rapid proxy isolation contained 98% loss", []string{"usr_mule_1"})
	assert.NotEmpty(t, entry.EntryID)

	past, foundMem := mem.QueryPastIncident("inc_500")
	assert.True(t, foundMem)
	assert.Equal(t, "SUCCESSFUL_DEFENSE", past.OutcomeType)
}

func TestCouncil_RuleGenerationAndIncidentCommander(t *testing.T) {
	ruleGen := NewAutonomousRuleGenerationAgent()
	cand := ruleGen.SynthesizeRule("RapidProxySurge", "IF proxy_velocity_10m > 50 THEN BLOCK")
	assert.Equal(t, RuleStageDiscovery, cand.Stage)

	ruleGen.AdvancePipeline(cand)
	assert.Equal(t, RuleStageSimulation, cand.Stage)
	ruleGen.AdvancePipeline(cand)
	assert.Equal(t, RuleStageShadowEvaluation, cand.Stage)
	ruleGen.AdvancePipeline(cand)
	assert.Equal(t, RuleStageGovernanceApproval, cand.Stage)
	ruleGen.AdvancePipeline(cand)
	assert.Equal(t, RuleStageProduction, cand.Stage)

	// Incident Commander
	cmdr := NewAIIncidentCommander()
	inc := cmdr.DeclareMajorIncident("Global Mule Network Surge", "SEV1")
	assert.Equal(t, IncidentDetected, inc.Stage)

	require.NoError(t, cmdr.AdvanceIncidentStage(inc.IncidentID))
	assert.Equal(t, IncidentAnalyzing, inc.Stage)
	require.NoError(t, cmdr.AdvanceIncidentStage(inc.IncidentID))
	assert.Equal(t, IncidentContaining, inc.Stage)
}

func TestCouncil_ComplianceMonitorAndDashboard(t *testing.T) {
	monitor := NewAutonomousComplianceMonitorAgent()

	decisions := []*CouncilDecision{
		{IncidentID: "dec_1", ConsensusAction: "BLOCK_AND_FREEZE", Confidence: 0.95},
		{IncidentID: "dec_2", ConsensusAction: "REQUIRE_MFA", Confidence: 0.70},
	}
	report := monitor.AuditDecisions(decisions)
	assert.Equal(t, "COMPLIANT", report.ComplianceStatus)
	assert.Empty(t, report.ViolationsFound)

	// Decision with violation (low confidence irreversible block)
	badDecisions := []*CouncilDecision{
		{IncidentID: "dec_bad", ConsensusAction: "BLOCK_AND_FREEZE", Confidence: 0.40},
	}
	badReport := monitor.AuditDecisions(badDecisions)
	assert.Equal(t, "VIOLATION_DETECTED", badReport.ComplianceStatus)
	assert.NotEmpty(t, badReport.ViolationsFound)

	// Dashboard
	dash := NewExecutiveDashboard(NewDebateEngine(), NewFraudForecastingEngine())
	overview := dash.GetOverview()
	assert.Equal(t, "OPTIMAL_DEFENSE", overview.SecurityReadinessStatus)
	assert.Greater(t, overview.AutonomousDefenseSuccessRate, 0.90)

	// Resource Optimizer
	optimizer := NewAgentResourceOptimizer()
	allocSev1 := optimizer.AllocateResources("SEV1")
	assert.Equal(t, 10, allocSev1["INVESTIGATOR"].PriorityScore)
}
