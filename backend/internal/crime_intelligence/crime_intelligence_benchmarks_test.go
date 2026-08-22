package crime_intelligence

import (
	"testing"
)

func BenchmarkCrime_GraphLookup(b *testing.B) {
	graph := NewCrimeIntelligenceGraph()
	_ = graph.AddNode(&CrimeNode{EntityID: "grp_bench_01", Type: EntityCriminalGroup, ThreatLevel: "CRITICAL"})
	_ = graph.AddNode(&CrimeNode{EntityID: "camp_bench_01", Type: EntityFraudCampaign, ThreatLevel: "HIGH"})
	_ = graph.AddEdge(&CrimeEdge{SourceID: "grp_bench_01", TargetID: "camp_bench_01", Type: RelOperates})

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = graph.QueryNeighbors("grp_bench_01", RelOperates)
	}
}

func BenchmarkCrime_ActorCorrelation(b *testing.B) {
	analyst := NewAICrimeAnalystAgent(nil)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = analyst.AnalyzeThreatActor("Shadow_Mule_Syndicate", 75000.0)
	}
}

func BenchmarkCrime_ThreatPrediction(b *testing.B) {
	evo := NewAttackEvolutionEngine()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = evo.PredictMutation("CREDENTIAL_STUFFING")
	}
}

func BenchmarkCrime_GlobalSimulation(b *testing.B) {
	sim := NewGlobalFraudSimulator2()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = sim.RunGlobalSimulation("TRANSNATIONAL_CARDING_OFFENSIVE", 1000, 250.0, 0.95)
	}
}

func BenchmarkCrime_ReportGeneration(b *testing.B) {
	repGen := NewThreatReportGenerator()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = repGen.GenerateExecutiveReport(15, 2500000.0)
	}
}
