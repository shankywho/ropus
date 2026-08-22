package agent_council

import (
	"testing"
)

func BenchmarkCouncil_DebateConsensus(b *testing.B) {
	debate := NewDebateEngine()
	opinions := []AgentOpinion{
		{AgentID: "inv_1", Role: "INVESTIGATOR", RecommendedAction: "BLOCK_AND_FREEZE", Confidence: 0.95},
		{AgentID: "threat_1", Role: "THREAT_HUNTER", RecommendedAction: "BLOCK_AND_FREEZE", Confidence: 0.96},
		{AgentID: "resp_1", Role: "RESPONSE", RecommendedAction: "BLOCK_AND_FREEZE", Confidence: 0.92},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = debate.Deliberate("inc_bench", opinions)
	}
}

func BenchmarkCouncil_FraudForecasting(b *testing.B) {
	forecaster := NewFraudForecastingEngine()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = forecaster.ForecastTrajectory(0.75, 100000.0)
	}
}

func BenchmarkCouncil_SimulationDuel(b *testing.B) {
	lab := NewDigitalCrimeSimulationLab()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = lab.SimulateAdversarialDuel("CARDING_BOT_BURST", "ENFORCE_STEP_UP_MFA", 500, 200.0, 0.95)
	}
}

func BenchmarkCouncil_StrategyEvaluation(b *testing.B) {
	strategist := NewAutonomousDefenseStrategist()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = strategist.SelectOptimalStrategy(50000.0, 0.95, 4)
	}
}

func BenchmarkCouncil_ExecutiveDashboard(b *testing.B) {
	dash := NewExecutiveDashboard(NewDebateEngine(), NewFraudForecastingEngine())

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = dash.GetOverview()
	}
}
