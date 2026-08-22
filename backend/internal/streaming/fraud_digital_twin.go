package streaming

import (
	"fmt"
	"time"
)

// DigitalTwinResult models the projected impact of counter-fraud policy or containment simulations.
type DigitalTwinResult struct {
	ScenarioName             string    `json:"scenario_name"`
	ProjectedFraudPrevented  float64   `json:"projected_fraud_prevented"`
	ProjectedFalsePositives  int       `json:"projected_false_positives"`
	ProjectedLegitimateImpact float64  `json:"projected_legitimate_impact"` // $ value of good transactions blocked
	ROIProjectedRatio        float64   `json:"roi_projected_ratio"`
	Recommendation           string    `json:"recommendation"` // "HIGHLY_RECOMMENDED", "PROCEED_WITH_CANARY", "REJECT"
	SimulatedAt              time.Time `json:"simulated_at"`
}

// FraudDigitalTwin provides predictive sandbox simulations for "what-if" risk scenario testing.
type FraudDigitalTwin struct{}

// NewFraudDigitalTwin initializes the digital twin simulator.
func NewFraudDigitalTwin() *FraudDigitalTwin {
	return &FraudDigitalTwin{}
}

// SimulatePolicyChange evaluates the counter-factual impact of enacting containment or deploying new thresholds.
func (dt *FraudDigitalTwin) SimulatePolicyChange(scenarioName string, estimatedAttacks int, avgAmount float64, confidence float64) *DigitalTwinResult {
	now := time.Now().UTC()

	fraudPrevented := float64(estimatedAttacks) * avgAmount * confidence
	projectedFP := int(float64(estimatedAttacks) * (1.0 - confidence) * 0.5)
	legitImpact := float64(projectedFP) * avgAmount

	roi := 10.0
	if legitImpact > 0 {
		roi = fraudPrevented / legitImpact
	}

	recommendation := "HIGHLY_RECOMMENDED"
	if confidence < 0.90 {
		recommendation = "PROCEED_WITH_CANARY"
	}
	if roi < 2.0 {
		recommendation = "REJECT"
	}

	return &DigitalTwinResult{
		ScenarioName:              scenarioName,
		ProjectedFraudPrevented:   fraudPrevented,
		ProjectedFalsePositives:   projectedFP,
		ProjectedLegitimateImpact: legitImpact,
		ROIProjectedRatio:         roi,
		Recommendation:            fmt.Sprintf("%s (Projected ROI: %.1fx)", recommendation, roi),
		SimulatedAt:               now,
	}
}
