package agent_council

import (
	"fmt"
	"time"
)

// SimulationEvaluation contains the outcome of an adversarial simulation duel.
type SimulationEvaluation struct {
	ScenarioID            string    `json:"scenario_id"`
	AttackerVector        string    `json:"attacker_vector"`
	DefenderPolicy        string    `json:"defender_policy"`
	LossPrevented         float64   `json:"loss_prevented"`
	ProjectedFPRate       float64   `json:"projected_fp_rate"`
	CustomerFrictionIndex float64   `json:"customer_friction_index"`
	NetUtilityScore       float64   `json:"net_utility_score"`
	SimulatedAt           time.Time `json:"simulated_at"`
}

// DigitalCrimeSimulationLab tests defender configurations against aggressive synthetic fraud adversaries.
type DigitalCrimeSimulationLab struct{}

// NewDigitalCrimeSimulationLab initializes the crime simulation sandbox.
func NewDigitalCrimeSimulationLab() *DigitalCrimeSimulationLab {
	return &DigitalCrimeSimulationLab{}
}

// SimulateAdversarialDuel pits synthetic attack campaigns against active or candidate defense controls.
func (lab *DigitalCrimeSimulationLab) SimulateAdversarialDuel(
	attackerVector string,
	defenderPolicy string,
	attackVolume int,
	avgTxnAmount float64,
	policyConfidence float64,
) *SimulationEvaluation {
	now := time.Now().UTC()

	grossExposure := float64(attackVolume) * avgTxnAmount
	lossPrevented := grossExposure * policyConfidence
	fpRate := (1.0 - policyConfidence) * 0.40
	frictionIndex := float64(attackVolume) * fpRate * 15.0 // $15 friction cost per good user impacted
	netUtility := lossPrevented - frictionIndex

	return &SimulationEvaluation{
		ScenarioID:            fmt.Sprintf("sim_%d", now.UnixNano()),
		AttackerVector:        attackerVector,
		DefenderPolicy:        defenderPolicy,
		LossPrevented:         lossPrevented,
		ProjectedFPRate:       fpRate,
		CustomerFrictionIndex: frictionIndex,
		NetUtilityScore:       netUtility,
		SimulatedAt:           now,
	}
}
