package crime_intelligence

import (
	"fmt"
	"time"
)

// GlobalSimulationReport summarizes the counter-factual evaluation of global crime scenarios.
type GlobalSimulationReport struct {
	SimulationID            string    `json:"simulation_id"`
	CampaignType            string    `json:"campaign_type"`
	SimulatedAdversaryNodes int       `json:"simulated_adversary_nodes"`
	EstimatedGrossExposure  float64   `json:"estimated_gross_exposure"`
	SimulatedLossPrevented  float64   `json:"simulated_loss_prevented"`
	DefenseEffectiveness    float64   `json:"defense_effectiveness"` // 0.0 to 1.0
	AttackSuccessLikelihood float64   `json:"attack_success_likelihood"`
	SimulatedAt             time.Time `json:"simulated_at"`
}

// GlobalFraudSimulator2 models transnational crime campaigns and consortium defense barriers.
type GlobalFraudSimulator2 struct{}

// NewGlobalFraudSimulator2 initializes the global fraud simulator.
func NewGlobalFraudSimulator2() *GlobalFraudSimulator2 {
	return &GlobalFraudSimulator2{}
}

// RunGlobalSimulation tests defender resilience against a simulated enterprise cyber-fraud offensive.
func (s *GlobalFraudSimulator2) RunGlobalSimulation(campaignType string, adversaryNodes int, avgNodeValue float64, defenseMaturity float64) *GlobalSimulationReport {
	now := time.Now().UTC()

	gross := float64(adversaryNodes) * avgNodeValue
	prevented := gross * defenseMaturity
	attackSuccess := 1.0 - defenseMaturity

	return &GlobalSimulationReport{
		SimulationID:            fmt.Sprintf("gsim_%d", now.UnixNano()),
		CampaignType:            campaignType,
		SimulatedAdversaryNodes: adversaryNodes,
		EstimatedGrossExposure:  gross,
		SimulatedLossPrevented:  prevented,
		DefenseEffectiveness:    defenseMaturity,
		AttackSuccessLikelihood: attackSuccess,
		SimulatedAt:             now,
	}
}
