package crime_intelligence

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCrime_GraphAndLineage(t *testing.T) {
	graph := NewCrimeIntelligenceGraph()

	group := &CrimeNode{EntityID: "grp_phantom", Type: EntityCriminalGroup, ThreatLevel: "CRITICAL"}
	camp1 := &CrimeNode{EntityID: "camp_v1", Type: EntityFraudCampaign, ThreatLevel: "HIGH"}
	camp2 := &CrimeNode{EntityID: "camp_v2", Type: EntityFraudCampaign, ThreatLevel: "CRITICAL"}

	require.NoError(t, graph.AddNode(group))
	require.NoError(t, graph.AddNode(camp1))
	require.NoError(t, graph.AddNode(camp2))

	require.NoError(t, graph.AddEdge(&CrimeEdge{SourceID: "grp_phantom", TargetID: "camp_v1", Type: RelOperates}))
	require.NoError(t, graph.AddEdge(&CrimeEdge{SourceID: "camp_v2", TargetID: "camp_v1", Type: RelEvolvedFrom}))

	// Query neighbors
	neighbors := graph.QueryNeighbors("grp_phantom", RelOperates)
	require.Equal(t, 1, len(neighbors))
	assert.Equal(t, "camp_v1", neighbors[0].EntityID)

	// Lineage
	lineage := graph.GetCampaignLineage("camp_v2")
	assert.GreaterOrEqual(t, len(lineage), 2)
}

func TestCrime_EcosystemModelAndAnalyst(t *testing.T) {
	model := NewFinancialCrimeEcosystemModel()
	analysis := model.ModelSyndicateEcosystem("syn_apex_01", 500, 15, 3, 750000.0)
	assert.Equal(t, "TIER_1_APG_TRANSNATIONAL", analysis.ClassificationTier)
	assert.GreaterOrEqual(t, analysis.ThreatMaturityScore, 0.80)
	assert.Greater(t, analysis.ProjectedAnnualExposure, 5000000.0)

	// Analyst Agent
	analyst := NewAICrimeAnalystAgent(nil)
	report := analyst.AnalyzeThreatActor("Shadow_Mule_Cartel", 50000.0)
	assert.NotEmpty(t, report.ReportID)
	assert.Equal(t, 0.96, report.Confidence)
	assert.NotEmpty(t, report.ObservedTechniques)
}

func TestCrime_AttackEvolutionAndMoneyFlow(t *testing.T) {
	evo := NewAttackEvolutionEngine()
	mut := evo.PredictMutation("CREDENTIAL_STUFFING")
	assert.Equal(t, "DISTRIBUTED_RESIDENTIAL_BOT_STUFFING", mut.MutatedTechnique)
	assert.NotEmpty(t, mut.CountermeasureDSL)

	// Money Flow
	mfa := NewMoneyFlowAnalyzer()
	resLayering := mfa.AnalyzeFlowTrajectory("acc_src_1", "acc_dst_9", 15000.0, 4, 5)
	assert.Equal(t, FlowLayeringRapidHops, resLayering.PatternDetected)
	assert.GreaterOrEqual(t, resLayering.LaunderingRiskScore, 0.90)

	resCircular := mfa.AnalyzeFlowTrajectory("acc_origin", "acc_origin", 25000.0, 3, 60)
	assert.Equal(t, FlowCircularMovement, resCircular.PatternDetected)
}

func TestCrime_InfrastructureAndThreatRadar(t *testing.T) {
	infra := NewInfrastructureIntelligenceEngine()
	infra.RecordInfrastructure("dev_emulator_root_pool", "EMULATOR_FARM", "grp_phantom", true)

	score, hostile := infra.QueryReputation("dev_emulator_root_pool")
	assert.True(t, hostile)
	assert.GreaterOrEqual(t, score, 0.90)

	// Threat Radar
	radar := NewPredictiveThreatRadar()
	forecasts := radar.GenerateRadarForecast()
	assert.Equal(t, 3, len(forecasts))
	assert.Equal(t, 7, forecasts[0].DaysHorizon)
	assert.Equal(t, 30, forecasts[1].DaysHorizon)
	assert.Equal(t, 90, forecasts[2].DaysHorizon)
}

func TestCrime_SimulatorAndStrategicCouncil(t *testing.T) {
	sim := NewGlobalFraudSimulator2()
	res := sim.RunGlobalSimulation("TRANSNATIONAL_CARDING_OFFENSIVE", 1000, 250.0, 0.95)
	assert.Greater(t, res.SimulatedLossPrevented, 200000.0)

	// Strategic Council
	council := NewStrategicIntelligenceCouncil(nil)
	directive := council.IssueMacroDirective(12, 1500000.0)
	assert.Equal(t, "DEFCON_1_CRITICAL", directive.EcosystemThreatLevel)
	assert.NotEmpty(t, directive.ConsortiumPriorityActions)

	// Response Network
	respNet := NewGlobalThreatResponseNetwork()
	bcast := respNet.BroadcastDefenseAction("BLOCK_INFRASTRUCTURE", "Isolate rogue ASN 99999", "consortium_hub_alpha")
	assert.NotEmpty(t, bcast.BroadcastID)
	assert.Equal(t, 24, bcast.ParticipatingPeersCount)
}

func TestCrime_MemoryReportsAndSecurityGuard(t *testing.T) {
	mem := NewCrimeKnowledgeMemory()
	mem.StoreMemory("Syndicate_Alpha", "Residential proxy hopping", "Biometric challenge", 0.98)
	recalled, found := mem.QueryMemory("Syndicate_Alpha")
	assert.True(t, found)
	assert.Equal(t, 0.98, recalled.OutcomeScore)

	// Reports
	repGen := NewThreatReportGenerator()
	execRep := repGen.GenerateExecutiveReport(15, 2500000.0)
	assert.NotEmpty(t, execRep.ReportID)

	analystRep := repGen.GenerateAnalystReport("Syndicate_Beta", []string{"acc_1", "acc_2"}, 0.95)
	assert.NotEmpty(t, analystRep.ReportID)

	// Security Guard
	guard := NewCrimeSecurityGuard()
	require.NoError(t, guard.ValidateIntelligenceAccess("consortium_master", "ALL"))
	assert.Error(t, guard.ValidateIntelligenceAccess("unauthorized_party", "ALL"))

	// Poison protection
	assert.Error(t, guard.ValidatePoisonProtection(0.95, 0.40), "Low reputation reporter must be rejected")
	require.NoError(t, guard.ValidatePoisonProtection(0.95, 0.85))
}
