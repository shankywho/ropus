package graphsage

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGraphSAGE_HeterogeneousGraphConstruction(t *testing.T) {
	store := NewTemporalHeteroGraphStore()

	// 1. Add all 10 Node Types
	nodes := []*HeteroNode{
		{ID: "emp_101", Type: NodeEmployee, RiskScore: 0.05},
		{ID: "usr_alice", Type: NodeConsumer, RiskScore: 0.10},
		{ID: "acc_9001", Type: NodeAccount, RiskScore: 0.10},
		{ID: "dev_iphone_15", Type: NodeDevice, RiskScore: 0.05},
		{ID: "ip_192_0_2_1", Type: NodeIP, RiskScore: 0.05},
		{ID: "pm_visa_8844", Type: NodePaymentMethod, RiskScore: 0.05},
		{ID: "tx_998811", Type: NodeTransaction, RiskScore: 0.15},
		{ID: "case_4411", Type: NodeCase, RiskScore: 0.20},
		{ID: "mch_apple_store", Type: NodeMerchant, RiskScore: 0.01},
		{ID: "ses_xyz_99", Type: NodeSession, RiskScore: 0.05},
	}
	for _, n := range nodes {
		require.NoError(t, store.AddNode(n))
	}

	// 2. Add sample edges
	now := time.Now().UTC()
	edges := []*HeteroEdge{
		{ID: "e1", SourceID: "emp_101", TargetID: "usr_alice", Type: EdgeAccessesConsumer, Timestamp: now},
		{ID: "e2", SourceID: "usr_alice", TargetID: "acc_9001", Type: EdgeOwns, Timestamp: now},
		{ID: "e3", SourceID: "acc_9001", TargetID: "dev_iphone_15", Type: EdgeUsesDevice, Timestamp: now},
		{ID: "e4", SourceID: "dev_iphone_15", TargetID: "ip_192_0_2_1", Type: EdgeUsesIP, Timestamp: now},
		{ID: "e5", SourceID: "acc_9001", TargetID: "pm_visa_8844", Type: EdgeUsesPayment, Timestamp: now},
		{ID: "e6", SourceID: "acc_9001", TargetID: "tx_998811", Type: EdgeCreatesTx, Timestamp: now},
		{ID: "e7", SourceID: "tx_998811", TargetID: "mch_apple_store", Type: EdgeTargets, Timestamp: now},
		{ID: "e8", SourceID: "emp_101", TargetID: "case_4411", Type: EdgeReviewsCase, Timestamp: now},
		{ID: "e9", SourceID: "acc_9001", TargetID: "ses_xyz_99", Type: EdgeHasSession, Timestamp: now},
	}
	for _, e := range edges {
		require.NoError(t, store.AddEdge(e))
	}

	nodeCount, edgeCount := store.Count()
	assert.Equal(t, 10, nodeCount)
	assert.Equal(t, 9, edgeCount)
}

func TestGraphSAGE_AdversarialTemporalLeakage_Cases1to5(t *testing.T) {
	store := NewTemporalHeteroGraphStore()
	engine := NewRelationshipIntelligenceEngine(store)

	tEval := time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC)
	tPast := tEval.Add(-2 * time.Hour)
	tFuture := tEval.Add(2 * time.Hour)

	// Setup Base Entities at tPast
	store.AddNode(&HeteroNode{ID: "emp_temp", Type: NodeEmployee, RiskScore: 0.05, CreatedAt: tPast})
	store.AddNode(&HeteroNode{ID: "usr_victim", Type: NodeConsumer, RiskScore: 0.05, CreatedAt: tPast})
	store.AddNode(&HeteroNode{ID: "acc_victim", Type: NodeAccount, RiskScore: 0.05, CreatedAt: tPast})
	store.AddNode(&HeteroNode{ID: "dev_past", Type: NodeDevice, RiskScore: 0.05, CreatedAt: tPast})

	// Case 1: Relationship appears AFTER transaction (t_rel > t_tx/t_eval)
	store.AddEdge(&HeteroEdge{
		ID:        "edge_case1_future_rel",
		SourceID:  "emp_temp",
		TargetID:  "usr_victim",
		Type:      EdgeAccessesConsumer,
		Timestamp: tFuture,
	})

	// Case 2: Employee access appears AFTER fraud event (t_access > t_fraud/t_eval)
	store.AddEdge(&HeteroEdge{
		ID:        "edge_case2_future_access",
		SourceID:  "emp_temp",
		TargetID:  "acc_victim",
		Type:      EdgeAccessesAccount,
		Timestamp: tFuture,
	})

	// Case 3: Graph edge is future-dated (t_edge > t_eval)
	store.AddEdge(&HeteroEdge{
		ID:        "edge_case3_future_device",
		SourceID:  "acc_victim",
		TargetID:  "dev_past",
		Type:      EdgeUsesDevice,
		Timestamp: tFuture,
	})

	// Case 4: Node is future-dated (t_created > t_eval)
	store.AddNode(&HeteroNode{ID: "node_case4_future", Type: NodeConsumer, RiskScore: 0.90, CreatedAt: tFuture})
	store.AddEdge(&HeteroEdge{
		ID:        "edge_case4_to_future_node",
		SourceID:  "emp_temp",
		TargetID:  "node_case4_future",
		Type:      EdgeAccessesConsumer,
		Timestamp: tPast, // edge timestamp was set to past but node created in future
	})

	// Case 5: Consumer joins device cluster AFTER historical decision (t_join > t_eval)
	store.AddNode(&HeteroNode{ID: "dev_shared_cluster", Type: NodeDevice, RiskScore: 0.05, CreatedAt: tPast})
	store.AddEdge(&HeteroEdge{
		ID:        "edge_case5_future_join",
		SourceID:  "emp_temp",
		TargetID:  "dev_shared_cluster",
		Type:      EdgeSharesDevice,
		Timestamp: tFuture,
	})

	// Run point-in-time evaluation strictly as of tEval
	report := engine.EvaluateRelationship(context.Background(), "emp_temp", tEval)
	require.NotNil(t, report)

	// Verification 1: Risk score remains baseline low because all suspicious links occurred in the future
	assert.Equal(t, RiskLevelLow, report.RiskLevel, "Future events MUST NOT cause risk score inflation")
	assert.Equal(t, 0.0, report.Signals.ConcentrationRatio)
	assert.Equal(t, 0, report.Signals.SharedDeviceOverlap)
	assert.Empty(t, report.ExplainableSignals, "No future collusion signals should trigger at tEval")

	// Verification 2: Check neighborhood extraction strictly at tEval
	nh, err := store.GetTemporalSampledNeighborhood("emp_temp", tEval, []int{10, 5})
	require.NoError(t, err)
	assert.Empty(t, nh.Layer1, "Layer 1 must be empty because all incident edges are future-dated")
	assert.Empty(t, nh.Layer2, "Layer 2 must be empty")
}

func TestGraphSAGE_ExplainableCollusionSignals_SchemaVerification(t *testing.T) {
	store := NewTemporalHeteroGraphStore()
	engine := NewRelationshipIntelligenceEngine(store)

	now := time.Date(2026, 8, 15, 22, 30, 0, 0, time.UTC) // 22:30 is Off-Hours
	tPast := now.Add(-10 * time.Minute)

	store.AddNode(&HeteroNode{ID: "emp_rogue_analyst", Type: NodeEmployee, RiskScore: 0.05, CreatedAt: tPast})
	store.AddNode(&HeteroNode{ID: "usr_mule_target", Type: NodeConsumer, RiskScore: 0.75, IsKnownBad: true, CreatedAt: tPast})
	store.AddNode(&HeteroNode{ID: "dev_shared_hw", Type: NodeDevice, RiskScore: 0.60, CreatedAt: tPast})

	// 6 accesses in off-hours to mule target
	for i := 0; i < 6; i++ {
		store.AddEdge(&HeteroEdge{
			ID:        "e_mule_access_" + string(rune('0'+i)),
			SourceID:  "emp_rogue_analyst",
			TargetID:  "usr_mule_target",
			Type:      EdgeAccessesConsumer,
			Timestamp: now.Add(time.Duration(-i*2) * time.Minute),
		})
	}
	// Hardware colocation
	store.AddEdge(&HeteroEdge{
		ID:        "e_hw_colocate",
		SourceID:  "emp_rogue_analyst",
		TargetID:  "dev_shared_hw",
		Type:      EdgeEmployeeUsesDevice,
		Timestamp: now.Add(-5 * time.Minute),
	})

	report := engine.EvaluateRelationship(context.Background(), "emp_rogue_analyst", now)
	require.NotNil(t, report)
	assert.GreaterOrEqual(t, report.OverallRiskScore, 0.70)
	assert.NotEmpty(t, report.ExplainableSignals)

	// Verify Schema Contract of Explainable Signal
	sig := report.ExplainableSignals[0]
	assert.Equal(t, "employee_consumer_collusion", sig.RiskSignal)
	assert.Greater(t, sig.Score, 0.0)
	assert.LessOrEqual(t, sig.Score, 1.0)
	assert.Equal(t, "24h", sig.TemporalWindow)
	assert.NotEmpty(t, sig.AffectedEntities)
	assert.Greater(t, sig.Confidence, 0.5)

	// Verify that evidence contains machine-readable tokens
	expectedTokens := []string{"employee_account_access_frequency_anomaly", "shared_device_cluster", "off_hours_access_anomaly"}
	foundCount := 0
	for _, exp := range expectedTokens {
		for _, ev := range sig.Evidence {
			if ev == exp {
				foundCount++
				break
			}
		}
	}
	assert.GreaterOrEqual(t, foundCount, 2, "Expected at least 2 structured machine-readable evidence tokens")
}

func TestGraphSAGE_BoundedMemoryAndDegreeDefenses(t *testing.T) {
	// Configure store with small limits to verify safety bounding
	cfg := GraphStoreConfig{
		MaxNodes:         50,
		MaxEdges:         100,
		MaxDegreePerNode: 15,
	}
	store := NewTemporalHeteroGraphStoreWithConfig(cfg)

	now := time.Now().UTC()
	store.AddNode(&HeteroNode{ID: "node_hub", Type: NodeAccount, CreatedAt: now})

	// Sybil / High-Degree Flood Simulation: Add 100 incident edges
	for i := 0; i < 100; i++ {
		targetID := "node_sybil_" + string(rune('a'+(i%26))) + string(rune('0'+(i/26)))
		_ = store.AddNode(&HeteroNode{ID: targetID, Type: NodeTransaction, CreatedAt: now})
		_ = store.AddEdge(&HeteroEdge{
			ID:        "edge_flood_" + targetID,
			SourceID:  "node_hub",
			TargetID:  targetID,
			Type:      EdgeCreatesTx,
			Timestamp: now.Add(time.Duration(-i) * time.Second),
		})
	}

	nodeCount, edgeCount := store.Count()
	assert.LessOrEqual(t, nodeCount, 50, "Node count must be bounded by MaxNodes")
	assert.LessOrEqual(t, edgeCount, 100, "Edge count must be bounded by MaxEdges")

	// Extraction must be strictly bounded by MaxDegreePerNode (15) and sample sizes (10, 5)
	nh, err := store.GetTemporalSampledNeighborhood("node_hub", now, []int{10, 5})
	require.NoError(t, err)
	assert.LessOrEqual(t, len(nh.Layer1), 10, "Layer 1 neighbors must be bounded by sample size")
}

func TestGraphSAGE_ContextTimeoutAndFailOpen(t *testing.T) {
	store := NewTemporalHeteroGraphStore()
	engine := NewRelationshipIntelligenceEngine(store)

	// Create an already-cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	report := engine.EvaluateRelationship(ctx, "emp_timeout_test", time.Now().UTC())
	require.NotNil(t, report)
	assert.True(t, report.IsDegraded)
	assert.Equal(t, RiskLevelLow, report.RiskLevel)
	assert.Equal(t, 0.05, report.OverallRiskScore)
	assert.Contains(t, report.DegradeReason, "context error")
}

func TestGraphSAGE_DeterministicEmbeddings(t *testing.T) {
	store := NewTemporalHeteroGraphStore()
	engine := NewRelationshipIntelligenceEngine(store)

	now := time.Now().UTC()
	store.AddNode(&HeteroNode{ID: "usr_alice", Type: NodeConsumer, RiskScore: 0.20, CreatedAt: now})

	report1 := engine.EvaluateRelationship(context.Background(), "usr_alice", now)
	report2 := engine.EvaluateRelationship(context.Background(), "usr_alice", now)

	require.Equal(t, len(report1.Embedding), len(report2.Embedding))
	for i := range report1.Embedding {
		assert.InDelta(t, report1.Embedding[i], report2.Embedding[i], 1e-6, "Embeddings must be strictly deterministic")
	}
}

func TestGraphSAGE_InductiveUnseenEntity(t *testing.T) {
	store := NewTemporalHeteroGraphStore()
	engine := NewRelationshipIntelligenceEngine(store)

	// Evaluate brand new unseen node that doesn't exist in store
	report := engine.EvaluateRelationship(context.Background(), "emp_completely_unseen_999", time.Now().UTC())
	require.NotNil(t, report)
	assert.Equal(t, RiskLevelLow, report.RiskLevel)
	assert.Equal(t, 64, len(report.Embedding))
	assert.False(t, report.IsDegraded)
}

func TestGraphSAGE_PrivacyAndAnonymization(t *testing.T) {
	store := NewTemporalHeteroGraphStore()
	engine := NewRelationshipIntelligenceEngine(store)

	now := time.Now().UTC()
	store.AddNode(&HeteroNode{
		ID:   "tok_sha256_99388a1b",
		Type: NodePaymentMethod,
		Properties: map[string]interface{}{
			"card_brand": "VISA",
			"token_ref":  "tok_sha256_99388a1b",
		},
		CreatedAt: now,
	})

	report := engine.EvaluateRelationship(context.Background(), "tok_sha256_99388a1b", now)
	require.NotNil(t, report)

	// Verify no PAN/CVV is leaked in evidence
	for _, fact := range report.ObservedFacts {
		assert.NotContains(t, fact, "4111")
		assert.NotContains(t, fact, "cvv")
	}
	for _, pattern := range report.InferredPatterns {
		assert.NotContains(t, pattern, "4111")
		assert.NotContains(t, pattern, "cvv")
	}
}
