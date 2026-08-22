package intelligence_fabric

import (
	"fmt"
	"time"
)

// DigitalTwinScenarioResult models the simulated impact of a hypothetical policy modification.
type DigitalTwinScenarioResult struct {
	ScenarioID            string    `json:"scenario_id"`
	PolicyChange          string    `json:"policy_change"`
	SimulatedFraudSaved   float64   `json:"simulated_fraud_saved"`
	SimulatedFPRate       float64   `json:"simulated_fp_rate"`
	ExpectedROI           float64   `json:"expected_roi"`
	DecisionRecommendation string   `json:"decision_recommendation"` // "SAFE_TO_DEPLOY", "REQUIRES_CANARY", "DANGEROUS_HIGH_FP"
	SimulatedAt           time.Time `json:"simulated_at"`
}

// DefenseDigitalTwin2 provides counter-factual sandbox evaluations for policy adjustments.
type DefenseDigitalTwin2 struct{}

// NewDefenseDigitalTwin2 initializes Digital Twin 2.0.
func NewDefenseDigitalTwin2() *DefenseDigitalTwin2 {
	return &DefenseDigitalTwin2{}
}

// EvaluatePolicyChange simulates the systemic impact of enacting or removing a defense rule.
func (dt *DefenseDigitalTwin2) EvaluatePolicyChange(changeDescription string, baseVolume int, avgAmount float64, confidence float64) *DigitalTwinScenarioResult {
	now := time.Now().UTC()

	gross := float64(baseVolume) * avgAmount
	fraudSaved := gross * confidence * 0.90
	fpRate := (1.0 - confidence) * 0.05
	frictionCost := float64(baseVolume) * fpRate * 20.0

	roi := 12.0
	if frictionCost > 0 {
		roi = fraudSaved / frictionCost
	}

	rec := "SAFE_TO_DEPLOY"
	if confidence < 0.90 {
		rec = "REQUIRES_CANARY"
	}
	if fpRate > 0.03 {
		rec = "DANGEROUS_HIGH_FP"
	}

	return &DigitalTwinScenarioResult{
		ScenarioID:             fmt.Sprintf("dt2_%d", now.UnixNano()),
		PolicyChange:           changeDescription,
		SimulatedFraudSaved:    fraudSaved,
		SimulatedFPRate:        fpRate,
		ExpectedROI:            roi,
		DecisionRecommendation: fmt.Sprintf("%s (Simulated ROI: %.1fx)", rec, roi),
		SimulatedAt:            now,
	}
}
