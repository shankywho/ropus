package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGraph_IngestionAndNeighbors(t *testing.T) {
	store := NewLocalGraphStore()
	engine := NewGraphEngine(store)

	err := engine.IngestTransactionLinks(
		"txn_01", "usr_alice", "acc_1001", "card_hash_99", "dev_iphone_15", "198.51.100.1", "merch_apple", 999.0, false,
	)
	require.NoError(t, err)

	assert.Equal(t, 7, store.CountNodes())
	assert.Equal(t, 6, store.CountEdges())

	neighbors, err := store.QueryNeighbors("usr_alice", "")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(neighbors), 3)

	paths, err := store.FindPaths("usr_alice", "card_hash_99", 3)
	require.NoError(t, err)
	assert.NotEmpty(t, paths)
	assert.Equal(t, 2, paths[0].Length) // alice -> acc_1001 -> card_hash_99
}

func TestGraph_RealTimeFeatureExtraction(t *testing.T) {
	store := NewLocalGraphStore()
	engine := NewGraphEngine(store)
	extractor := NewGraphFeatureExtractor(engine)

	// Ingest clean transaction
	_ = engine.IngestTransactionLinks("txn_clean_01", "usr_clean", "acc_clean", "card_c1", "dev_clean_01", "10.0.0.1", "merch_01", 50.0, false)

	featClean, err := extractor.ExtractFeatures("usr_clean", "dev_clean_01", "10.0.0.1")
	require.NoError(t, err)
	assert.Equal(t, 0, featClean.FraudNeighborCount)
	assert.Less(t, featClean.GraphRiskScore, 0.30)

	// Ingest fraud transaction sharing same device
	_ = engine.IngestTransactionLinks("txn_bad_01", "usr_bad", "acc_bad", "card_b1", "dev_clean_01", "10.0.0.1", "merch_01", 5000.0, true)

	featDirty, err := extractor.ExtractFeatures("usr_clean", "dev_clean_01", "10.0.0.1")
	require.NoError(t, err)
	assert.Greater(t, featDirty.FraudNeighborCount, 0)
	assert.GreaterOrEqual(t, featDirty.GraphRiskScore, 0.85)
}

func TestGraph_EntityResolution(t *testing.T) {
	resEngine := NewEntityResolutionEngine()

	c1 := resEngine.ResolveEntity("usr_alice", "dev_laptop_1", "card_111", "192.168.1.1")
	assert.NotEmpty(t, c1.ClusterID)
	assert.False(t, c1.IsSuspicious)

	// Same person using different email/user but same laptop
	c2 := resEngine.ResolveEntity("usr_alice_alias", "dev_laptop_1", "card_222", "192.168.1.5")
	assert.Equal(t, c1.ClusterID, c2.ClusterID, "Entities sharing device should cluster together")
	assert.Contains(t, c2.MemberNodeIDs, "usr_alice")
	assert.Contains(t, c2.MemberNodeIDs, "usr_alice_alias")
}

func TestGraph_FraudRingDetection(t *testing.T) {
	store := NewLocalGraphStore()
	engine := NewGraphEngine(store)
	simulator := NewFraudSimulator(engine)
	detector := NewFraudRingDetector(store)

	// Simulate a synthetic identity farm sharing 1 device hub
	simResult, err := simulator.SimulateAttack(ScenarioSyntheticIdentityFarm)
	require.NoError(t, err)
	assert.Equal(t, ScenarioSyntheticIdentityFarm, simResult.Scenario)

	rings := detector.DetectRings(3)
	assert.NotEmpty(t, rings, "Should detect synthetic identity farm ring")
	assert.GreaterOrEqual(t, rings[0].RingSize, 4)
	assert.GreaterOrEqual(t, rings[0].FraudRingScore, 0.60)
}

func TestGraph_BehaviorAndThreatEngines(t *testing.T) {
	be := NewBehaviorEngine()

	// Normal baselines
	score1, _ := be.EvaluateBehavior("usr_bob", 100.0, "dev_bob_phone", "10.0.0.1", "US")
	assert.Less(t, score1, 0.20)

	_, _ = be.EvaluateBehavior("usr_bob", 120.0, "dev_bob_phone", "10.0.0.1", "US")
	_, _ = be.EvaluateBehavior("usr_bob", 90.0, "dev_bob_phone", "10.0.0.1", "US")

	// 10x spending spike + brand new device
	spikeScore, anomalies := be.EvaluateBehavior("usr_bob", 2500.0, "dev_unknown_mac", "10.0.0.1", "US")
	assert.GreaterOrEqual(t, spikeScore, 0.45)
	assert.NotEmpty(t, anomalies)

	// Threat Intelligence
	te := NewThreatIntelligenceEngine()
	threatScore, matches := te.CheckThreat("198.51.100.44", "dev_normal", "legit-corp.com")
	assert.GreaterOrEqual(t, threatScore, 0.90)
	assert.NotEmpty(t, matches)
}

func TestGraph_AdaptiveRiskEngineSynthesis(t *testing.T) {
	store := NewLocalGraphStore()
	engine := NewGraphEngine(store)
	gx := NewGraphFeatureExtractor(engine)
	be := NewBehaviorEngine()
	te := NewThreatIntelligenceEngine()

	adaptiveEngine := NewAdaptiveRiskEngine(gx, be, te)

	// 1. Low risk evaluation
	resClean, err := adaptiveEngine.EvaluateAdaptiveRisk("usr_good", "dev_good", "1.1.1.1", "gmail.com", "US", 50.0, 0.05, 0.0)
	require.NoError(t, err)
	assert.Equal(t, "ALLOW", resClean.Decision)
	assert.Less(t, resClean.FinalScore, 0.30)

	// 2. High risk evaluation (Threat IP + ML score)
	resBad, err := adaptiveEngine.EvaluateAdaptiveRisk("usr_bad", "dev_emul_root_89a", "198.51.100.44", "temp-mail.org", "RU", 10000.0, 0.90, 0.80)
	require.NoError(t, err)
	assert.Equal(t, "BLOCK", resBad.Decision)
	assert.GreaterOrEqual(t, resBad.FinalScore, 0.90)
	assert.NotEmpty(t, resBad.ContributingReasons)
}

func TestGraph_GNNAdapter(t *testing.T) {
	gnn := NewLocalGNNAdapter()
	score := gnn.PredictGraphRisk("node_1", []float64{0.1, 0.5, 0.9}, []string{"n2", "n3", "n4", "n5"})
	assert.Greater(t, score, 0.20)
}
