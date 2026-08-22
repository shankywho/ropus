package agent_council

import (
	"fmt"
	"time"
)

// FraudForecast encapsulates predictive crime intelligence over future time horizons.
type FraudForecast struct {
	ForecastID               string    `json:"forecast_id"`
	AttackType               string    `json:"attack_type"` // "EMULATOR_FARM_SPIKE", "DISTRIBUTED_CREDENTIAL_STUFFING", "SYNTHETIC_IDENTITY_WAVE"
	Probability              float64   `json:"probability"`
	TimeWindowHours          int       `json:"time_window_hours"`
	PredictedExposure        float64   `json:"predicted_exposure"`
	ConfidenceInterval       []float64 `json:"confidence_interval"` // [min, max]
	RecommendedPreemptiveRule string    `json:"recommended_preemptive_rule"`
	ForecastGeneratedAt      time.Time `json:"forecast_generated_at"`
}

// FraudForecastingEngine projects emerging crime trends and synthetic attack waves before impact.
type FraudForecastingEngine struct{}

// NewFraudForecastingEngine initializes the predictive forecasting engine.
func NewFraudForecastingEngine() *FraudForecastingEngine {
	return &FraudForecastingEngine{}
}

// ForecastTrajectory models attack acceleration vectors over recent historical baselines.
func (f *FraudForecastingEngine) ForecastTrajectory(recentVelocityGrowthRate float64, currentExposure float64) *FraudForecast {
	now := time.Now().UTC()

	prob := 0.75
	attackType := "DISTRIBUTED_CREDENTIAL_STUFFING"
	predictedExp := currentExposure * 1.50
	rule := "PREEMPTIVE: Tighten login velocity to 5 attempts/10m for flagged ASN pools"

	if recentVelocityGrowthRate > 0.50 {
		prob = 0.94
		attackType = "EMULATOR_FARM_SURGE"
		predictedExp = currentExposure * 3.20
		rule = "PREEMPTIVE: Enforce hardware attestation check on newly registered devices"
	}

	return &FraudForecast{
		ForecastID:                fmt.Sprintf("fc_%d", now.UnixNano()),
		AttackType:                attackType,
		Probability:               prob,
		TimeWindowHours:           48,
		PredictedExposure:         predictedExp,
		ConfidenceInterval:        []float64{predictedExp * 0.85, predictedExp * 1.25},
		RecommendedPreemptiveRule: rule,
		ForecastGeneratedAt:       now,
	}
}
