package agents

import (
	"fmt"
	"time"
)

// SimulatedAttackVector represents an adversarial scenario generated to pressure test defense readiness.
type SimulatedAttackVector struct {
	SimulationID   string                 `json:"simulation_id"`
	VectorType     string                 `json:"vector_type"` // "CARDING_BURST", "DISTRIBUTED_ATO_STORM", "MULE_CHAIN_ROTATION"
	TargetNodeCount int                   `json:"target_node_count"`
	EstimatedAmount float64               `json:"estimated_amount"`
	Payload        map[string]interface{} `json:"payload"`
	GeneratedAt    time.Time              `json:"generated_at"`
}

// AttackSimulationAgent generates synthetic adversary campaigns to train and evaluate defenses ahead of real attacks.
type AttackSimulationAgent struct {
	AgentID string
}

// NewAttackSimulationAgent initializes the synthetic attack generator agent.
func NewAttackSimulationAgent() *AttackSimulationAgent {
	return &AttackSimulationAgent{AgentID: "agent_attack_simulator_v1"}
}

// GenerateAttackScenario creates a realistic synthetic attack profile.
func (a *AttackSimulationAgent) GenerateAttackScenario(vectorType string, nodeCount int) *SimulatedAttackVector {
	now := time.Now().UTC()
	return &SimulatedAttackVector{
		SimulationID:    fmt.Sprintf("sim_%d_%s", now.UnixNano(), vectorType),
		VectorType:      vectorType,
		TargetNodeCount: nodeCount,
		EstimatedAmount: float64(nodeCount) * 150.0,
		Payload: map[string]interface{}{
			"device_fingerprint": "sim_dev_emu_root",
			"proxy_ip_pool":      []string{"198.51.100.1", "198.51.100.2"},
			"synthetic_velocity": 45,
		},
		GeneratedAt: now,
	}
}
