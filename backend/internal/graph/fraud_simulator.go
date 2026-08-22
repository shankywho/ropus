package graph

import (
	"fmt"
	"time"
)

// AttackScenario defines the type of synthetic fraud attack simulation.
type AttackScenario string

const (
	ScenarioAccountTakeover      AttackScenario = "ACCOUNT_TAKEOVER"
	ScenarioCardTestingBurst     AttackScenario = "CARD_TESTING_BURST"
	ScenarioSyntheticIdentityFarm AttackScenario = "SYNTHETIC_IDENTITY_FARM"
	ScenarioMuleLaunderingChain  AttackScenario = "MULE_LAUNDERING_CHAIN"
)

// SimulatedAttackResult describes the generated attack topology and transactions.
type SimulatedAttackResult struct {
	Scenario            AttackScenario `json:"scenario"`
	GeneratedNodesCount int            `json:"generated_nodes_count"`
	GeneratedEdgesCount int            `json:"generated_edges_count"`
	TargetVictimID      string         `json:"target_victim_id,omitempty"`
	AttackerDeviceID    string         `json:"attacker_device_id"`
	AttackerIP          string         `json:"attacker_ip"`
	Timestamp           time.Time      `json:"timestamp"`
}

// FraudSimulator injects synthetic fraud patterns for testing, training, and benchmarking.
type FraudSimulator struct {
	engine *GraphEngine
}

// NewFraudSimulator initializes the fraud simulation engine.
func NewFraudSimulator(engine *GraphEngine) *FraudSimulator {
	return &FraudSimulator{engine: engine}
}

// SimulateAttack generates a realistic attack scenario into the graph.
func (s *FraudSimulator) SimulateAttack(scenario AttackScenario) (*SimulatedAttackResult, error) {
	now := time.Now().UTC()
	attackerDev := fmt.Sprintf("dev_sim_bad_%d", now.UnixNano())
	attackerIP := "198.51.100.44"

	nodesCount := 0
	edgesCount := 0

	switch scenario {
	case ScenarioAccountTakeover:
		victimUser := "usr_victim_123"
		txnID := fmt.Sprintf("txn_ato_%d", now.UnixNano())

		_ = s.engine.IngestTransactionLinks(txnID, victimUser, "acc_victim_01", "card_hash_v1", attackerDev, attackerIP, "merch_crypto_99", 5000.0, true)
		nodesCount = 7
		edgesCount = 6

	case ScenarioCardTestingBurst:
		for i := 0; i < 10; i++ {
			txnID := fmt.Sprintf("txn_card_test_%d_%d", now.UnixNano(), i)
			card := fmt.Sprintf("card_test_hash_%d", i)
			_ = s.engine.IngestTransactionLinks(txnID, "usr_card_tester", "acc_bot_01", card, attackerDev, attackerIP, "merch_donation_01", 1.50, true)
			nodesCount += 3
			edgesCount += 4
		}

	case ScenarioSyntheticIdentityFarm:
		farmDevice := "dev_farm_hub_01"
		for i := 0; i < 8; i++ {
			user := fmt.Sprintf("usr_synth_%d", i)
			txnID := fmt.Sprintf("txn_synth_%d", i)
			_ = s.engine.IngestTransactionLinks(txnID, user, fmt.Sprintf("acc_synth_%d", i), fmt.Sprintf("card_synth_%d", i), farmDevice, attackerIP, "merch_retail_01", 300.0, true)
			nodesCount += 4
			edgesCount += 5
		}

	case ScenarioMuleLaunderingChain:
		// Chain of transfers
		for i := 0; i < 4; i++ {
			u1 := fmt.Sprintf("mule_usr_%d", i)
			u2 := fmt.Sprintf("mule_usr_%d", i+1)
			_ = s.engine.Store().AddEdge(&Edge{
				ID:        fmt.Sprintf("mule_edge_%d", i),
				SourceID:  u1,
				TargetID:  u2,
				Type:      EdgeTransferredTo,
				Weight:    4900.0,
				CreatedAt: now,
			})
			nodesCount += 2
			edgesCount++
		}
	}

	return &SimulatedAttackResult{
		Scenario:            scenario,
		GeneratedNodesCount: nodesCount,
		GeneratedEdgesCount: edgesCount,
		AttackerDeviceID:    attackerDev,
		AttackerIP:          attackerIP,
		Timestamp:           now,
	}, nil
}
