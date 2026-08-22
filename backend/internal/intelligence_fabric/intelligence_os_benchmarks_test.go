package intelligence_fabric

import (
	"testing"
)

func BenchmarkFabric_SignalIngestion(b *testing.B) {
	ingestion := NewSignalIngestionEngine()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = ingestion.IngestSignal(SourceThreatFeed, "raw_hostile_ip_address", 0.95, 0.90, "BOT_WAVE", nil)
	}
}

func BenchmarkFabric_ThreatFusion(b *testing.B) {
	fusion := NewIntelligenceFusionEngine()
	signals := []*IntelligenceSignal{
		{SignalID: "s1", PrivacyHash: "hash_01", Confidence: 0.95, ReliabilityScore: 0.90, RawTopic: "BOT_WAVE"},
		{SignalID: "s2", PrivacyHash: "hash_02", Confidence: 0.90, ReliabilityScore: 0.85, RawTopic: "BOT_WAVE"},
		{SignalID: "s3", PrivacyHash: "hash_03", Confidence: 0.92, ReliabilityScore: 0.95, RawTopic: "BOT_WAVE"},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = fusion.FuseTelemetry(signals)
	}
}

func BenchmarkFabric_StrategyOptimization(b *testing.B) {
	optimizer := NewAutonomousStrategyOptimizer()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = optimizer.OptimizeStrategy("CRITICAL", 500000.0, 0.05)
	}
}

func BenchmarkFabric_DigitalTwinSimulation(b *testing.B) {
	dt := NewDefenseDigitalTwin2()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = dt.EvaluatePolicyChange("Enforce WebAuthn on Checkout", 5000, 150.0, 0.98)
	}
}

func BenchmarkFabric_ExecutiveDashboard(b *testing.B) {
	center := NewExecutiveIntelligenceCenter(NewIntelligenceFusionEngine(), NewAutonomousStrategyOptimizer())

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = center.GetGlobalPosture()
	}
}
