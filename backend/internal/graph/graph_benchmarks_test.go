package graph

import (
	"testing"
)

func BenchmarkGraph_NeighborLookup(b *testing.B) {
	store := NewLocalGraphStore()
	engine := NewGraphEngine(store)

	_ = engine.IngestTransactionLinks("txn_01", "usr_alice", "acc_1001", "card_99", "dev_iphone", "10.0.0.1", "merch_01", 100.0, false)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = store.QueryNeighbors("usr_alice", "")
	}
}

func BenchmarkGraph_RealTimeFeatureExtraction(b *testing.B) {
	store := NewLocalGraphStore()
	engine := NewGraphEngine(store)
	extractor := NewGraphFeatureExtractor(engine)

	_ = engine.IngestTransactionLinks("txn_01", "usr_alice", "acc_1001", "card_99", "dev_iphone", "10.0.0.1", "merch_01", 100.0, false)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = extractor.ExtractFeatures("usr_alice", "dev_iphone", "10.0.0.1")
	}
}

func BenchmarkGraph_EntityResolution(b *testing.B) {
	resEngine := NewEntityResolutionEngine()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = resEngine.ResolveEntity("usr_alice", "dev_laptop_1", "card_111", "192.168.1.1")
	}
}

func BenchmarkGraph_FraudRingDetection(b *testing.B) {
	store := NewLocalGraphStore()
	engine := NewGraphEngine(store)
	simulator := NewFraudSimulator(engine)
	detector := NewFraudRingDetector(store)

	_, _ = simulator.SimulateAttack(ScenarioSyntheticIdentityFarm)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = detector.DetectRings(3)
	}
}

func BenchmarkGraph_AdaptiveRiskScoring(b *testing.B) {
	store := NewLocalGraphStore()
	engine := NewGraphEngine(store)
	gx := NewGraphFeatureExtractor(engine)
	be := NewBehaviorEngine()
	te := NewThreatIntelligenceEngine()

	adaptiveEngine := NewAdaptiveRiskEngine(gx, be, te)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = adaptiveEngine.EvaluateAdaptiveRisk("usr_alice", "dev_iphone", "10.0.0.1", "gmail.com", "US", 250.0, 0.45, 0.10)
	}
}
