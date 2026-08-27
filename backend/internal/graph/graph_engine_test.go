package graph

import (
	"fmt"
	"testing"
)

func TestGraphEngine_Topologies(t *testing.T) {
	// Topology 1: Isolated Customer
	t.Run("Isolated Customer", func(t *testing.T) {
		engine := NewGraphEngine(nil)
		err := engine.IngestTransactionLinks("txn_001", "usr_iso", "acc_iso", "card_iso", "dev_iso", "106.51.0.1", "merch_01", 500, false)
		if err != nil {
			t.Fatalf("Ingest failed: %v", err)
		}

		evidence := engine.EvaluateEntityGraph("acc_iso", "dev_iso", "card_iso", "106.51.0.1")
		if evidence.ConnectedAccountCount != 0 {
			t.Errorf("Expected 0 connected accounts, got %d", evidence.ConnectedAccountCount)
		}
		if evidence.FraudNodesCount != 0 {
			t.Errorf("Expected 0 fraud nodes, got %d", evidence.FraudNodesCount)
		}
		if evidence.GraphRiskContribution > 0.30 {
			t.Errorf("Expected low risk for isolated customer, got %.2f", evidence.GraphRiskContribution)
		}
	})

	// Topology 2: Customer -> Device -> Multiple Accounts (Shared Mule Device)
	t.Run("Shared Mule Device", func(t *testing.T) {
		engine := NewGraphEngine(nil)
		devShared := "dev_emulator_shared_99"

		// Link 5 distinct accounts to the same device
		for i := 1; i <= 5; i++ {
			accID := fmt.Sprintf("acc_mule_%d", i)
			usrID := fmt.Sprintf("usr_mule_%d", i)
			txnID := fmt.Sprintf("txn_mule_%d", i)
			_ = engine.IngestTransactionLinks(txnID, usrID, accID, fmt.Sprintf("card_%d", i), devShared, "198.51.100.44", "merch_payout", 1000, false)
		}

		evidence := engine.EvaluateEntityGraph("acc_mule_1", devShared, "card_1", "198.51.100.44")
		if evidence.ConnectedAccountCount < 4 {
			t.Errorf("Expected at least 4 other connected accounts, got %d", evidence.ConnectedAccountCount)
		}
		if evidence.DegreeCentrality < 6 {
			t.Errorf("Expected degree centrality >= 6, got %d", evidence.DegreeCentrality)
		}
		if evidence.GraphRiskContribution < 0.50 {
			t.Errorf("Expected elevated graph risk >= 0.50, got %.2f", evidence.GraphRiskContribution)
		}
	})

	// Topology 3: Cluster with Known Fraudster Node
	t.Run("Cluster with Confirmed Fraud Node", func(t *testing.T) {
		engine := NewGraphEngine(nil)
		devKnownBad := "dev_fraud_ring_01"

		// Legitimate-looking account
		_ = engine.IngestTransactionLinks("txn_innocent", "usr_innocent", "acc_innocent", "card_inno", devKnownBad, "106.51.0.2", "merch_01", 200, false)
		// Known fraudster sharing the device
		_ = engine.IngestTransactionLinks("txn_bad", "usr_bad", "acc_bad", "card_bad", devKnownBad, "106.51.0.2", "merch_01", 10000, true)

		evidence := engine.EvaluateEntityGraph("acc_innocent", devKnownBad, "card_inno", "106.51.0.2")
		if evidence.FraudNodesCount < 1 {
			t.Errorf("Expected at least 1 fraud node detected, got %d", evidence.FraudNodesCount)
		}
		if evidence.GraphRiskContribution < 0.80 {
			t.Errorf("Expected high graph risk >= 0.80 for link to confirmed fraud, got %.2f", evidence.GraphRiskContribution)
		}
	})

	// Topology 4: Hub Capping & Explosion Protection
	t.Run("Hub Capping", func(t *testing.T) {
		engine := NewGraphEngine(nil)
		merchantHub := "merch_huge_retailer"

		// Ingest 60 transactions into the same merchant hub
		for i := 1; i <= 60; i++ {
			_ = engine.IngestTransactionLinks(fmt.Sprintf("txn_hub_%d", i), fmt.Sprintf("usr_%d", i), fmt.Sprintf("acc_%d", i), fmt.Sprintf("card_%d", i), fmt.Sprintf("dev_%d", i), "10.0.0.1", merchantHub, 50, false)
		}

		evidence := engine.EvaluateEntityGraph("acc_1", "dev_1", "card_1", "10.0.0.1")
		if evidence.VisitedNodesCount > MaxTotalNodesExpansion {
			t.Errorf("Visited nodes count %d exceeded MaxTotalNodesExpansion %d", evidence.VisitedNodesCount, MaxTotalNodesExpansion)
		}
	})
}
