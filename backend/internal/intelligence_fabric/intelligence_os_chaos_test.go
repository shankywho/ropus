package intelligence_fabric

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFabric_SignalIngestionAndGraphEvolution(t *testing.T) {
	ingestion := NewSignalIngestionEngine()
	sig, err := ingestion.IngestSignal(SourceThreatFeed, "raw_hostile_proxy_ip", 0.95, 0.90, "CREDENTIAL_STUFFING", map[string]interface{}{"asn": "12345"})
	require.NoError(t, err)
	assert.NotEmpty(t, sig.SignalID)
	assert.Equal(t, 1, len(ingestion.ListRecentSignals()))

	// Knowledge Graph 3.0 evolution
	graph := NewThreatKnowledgeGraphV3()
	node := graph.IngestSignalAndEvolve(sig, TypeInfrastructure, "")
	assert.Equal(t, 1, node.Version)
	assert.GreaterOrEqual(t, node.ThreatScore, 0.80)

	// Evolve with linked campaign
	sig2, _ := ingestion.IngestSignal(SourceThreatFeed, "raw_hostile_proxy_ip", 0.99, 0.95, "CREDENTIAL_STUFFING", nil)
	evolvedNode := graph.IngestSignalAndEvolve(sig2, TypeInfrastructure, "CAMPAIGN_alpha")
	assert.Equal(t, 2, evolvedNode.Version)
	assert.Contains(t, graph.QueryConnected(evolvedNode.NodeID), "CAMPAIGN_alpha")
}

func TestFabric_FusionAndStrategyOptimizer(t *testing.T) {
	fusion := NewIntelligenceFusionEngine()
	signals := []*IntelligenceSignal{
		{SignalID: "s1", PrivacyHash: "hash_01", Confidence: 0.95, ReliabilityScore: 0.90, RawTopic: "BOT_WAVE"},
		{SignalID: "s2", PrivacyHash: "hash_02", Confidence: 0.90, ReliabilityScore: 0.85, RawTopic: "BOT_WAVE"},
		{SignalID: "s3", PrivacyHash: "hash_03", Confidence: 0.92, ReliabilityScore: 0.95, RawTopic: "BOT_WAVE"},
		{SignalID: "s4", PrivacyHash: "hash_04", Confidence: 0.88, ReliabilityScore: 0.90, RawTopic: "BOT_WAVE"},
		{SignalID: "s5", PrivacyHash: "hash_05", Confidence: 0.96, ReliabilityScore: 0.95, RawTopic: "BOT_WAVE"},
	}

	picture := fusion.FuseTelemetry(signals)
	assert.Equal(t, "CRITICAL", picture.ThreatLevel)
	assert.NotEmpty(t, picture.RecommendedActions)

	// Strategy Optimizer
	optimizer := NewAutonomousStrategyOptimizer()
	strat := optimizer.OptimizeStrategy(picture.ThreatLevel, 500000.0, 0.05)
	assert.Equal(t, "AUTONOMOUS_HARD_BLOCK", strat.BlockingPosture)
	assert.Equal(t, 10, strat.AllocatedComputePriority)
	assert.Greater(t, strat.ProjectedFraudSavingsUSD, 400000.0)
}

func TestFabric_SelfLearningAndPolicyEvolution(t *testing.T) {
	learner := NewSelfLearningDefenseLoop()
	rec1 := learner.RecordDefenseOutcome("inc_1", 25000.0, false)
	assert.Greater(t, rec1.StrategyAdaptationDelta, 0.0)

	rec2 := learner.RecordDefenseOutcome("inc_2", 0.0, true)
	assert.Less(t, rec2.StrategyAdaptationDelta, 0.0)

	savings, fpRate, tp, fp := learner.GetLearningMetrics()
	assert.Equal(t, 25000.0, savings)
	assert.Equal(t, 0.50, fpRate)
	assert.Equal(t, 1, tp)
	assert.Equal(t, 1, fp)

	// Policy Evolution
	evo := NewAutonomousPolicyEvolutionEngine()
	cand := evo.CreateCandidate("DetectCellularProxySurge", "IF cellular_proxy_velocity > 50 THEN BLOCK")
	assert.Equal(t, PolicyStageDiscover, cand.Stage)

	evo.AdvanceStage(cand)
	assert.Equal(t, PolicyStageSimulate, cand.Stage)
	evo.AdvanceStage(cand)
	assert.Equal(t, PolicyStageShadow, cand.Stage)
	evo.AdvanceStage(cand)
	assert.Equal(t, PolicyStageGovernanceReview, cand.Stage)
	evo.AdvanceStage(cand)
	assert.Equal(t, PolicyStageCanary, cand.Stage)
	evo.AdvanceStage(cand)
	assert.Equal(t, PolicyStageProduction, cand.Stage)
}

func TestFabric_ResearcherAndRedTeam(t *testing.T) {
	researcher := NewAIFinancialCrimeResearcher()
	rep := researcher.ConductResearch("Transnational Mule Syndicates")
	assert.NotEmpty(t, rep.ReportID)
	assert.NotEmpty(t, rep.KeyFindings)

	// Red Team
	rt := NewAutonomousRedTeamEngine()
	attack := rt.ExecuteOffensiveSimulation("ADVERSARIAL_ML_EVASION", 1000, 0.95)
	assert.NotEmpty(t, attack.AttackID)
	assert.NotEmpty(t, attack.VulnerabilitiesFound)
	assert.NotEmpty(t, attack.RecommendedPatchDSL)
}

func TestFabric_DigitalTwinAndWorkspace(t *testing.T) {
	dt := NewDefenseDigitalTwin2()
	sim := dt.EvaluatePolicyChange("Enforce WebAuthn on Checkout", 5000, 150.0, 0.98)
	assert.Greater(t, sim.SimulatedFraudSaved, 500000.0)
	assert.Contains(t, sim.DecisionRecommendation, "SAFE_TO_DEPLOY")

	// Human + AI Workspace
	ws := NewHumanAIWorkspace()
	inter := ws.RecordInteraction("case_101", "analyst_alice", "Is this IP associated with known mule networks?", "High risk confirmed", 0.94, true, "Approved account freeze")
	assert.NotEmpty(t, inter.InteractionID)

	recalled, found := ws.GetInteraction(inter.InteractionID)
	assert.True(t, found)
	assert.Equal(t, "analyst_alice", recalled.AnalystID)
}

func TestFabric_ExecutiveCenterAndSecurityGuard(t *testing.T) {
	center := NewExecutiveIntelligenceCenter(NewIntelligenceFusionEngine(), NewAutonomousStrategyOptimizer())
	posture := center.GetGlobalPosture()
	assert.Equal(t, "HEALTHY_OPTIMAL", posture.AutonomousDefenseHealth)
	assert.Greater(t, posture.SignalsIngestedPerSec, 10000000.0)

	// Resource Optimizer
	resOpt := NewAutonomousResourceOptimizer()
	alloc := resOpt.BalanceResources("CRITICAL")
	assert.Equal(t, 10, alloc["SIGNAL_INGESTION"].ComputeWeight)

	// Security Guard
	guard := NewIntelligenceSecurityGuard()
	require.NoError(t, guard.AuthorizeIngestion("consortium_root", 0.90))
	assert.Error(t, guard.AuthorizeIngestion("unauthorized_party", 0.90))
	assert.Error(t, guard.AuthorizeIngestion("consortium_root", 0.30), "Low reliability signal must be rejected")

	require.NoError(t, guard.AuthorizePolicyPromotion(PolicyStageProduction, "ROLE_RISK_EXECUTIVE"))
	assert.Error(t, guard.AuthorizePolicyPromotion(PolicyStageProduction, "ROLE_ANALYST"))
}
