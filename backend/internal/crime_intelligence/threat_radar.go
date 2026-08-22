package crime_intelligence

import (
	"time"
)

// ThreatRadarHorizon models predictive risk across discrete temporal windows.
type ThreatRadarHorizon struct {
	DaysHorizon        int       `json:"days_horizon"` // 7, 30, 90
	PrimaryThreatType  string    `json:"primary_threat_type"`
	ThreatProbability  float64   `json:"threat_probability"`
	TargetRegions      []string  `json:"target_regions"`
	ProjectedAttackers int       `json:"projected_attackers"`
	RecommendedAction  string    `json:"recommended_action"`
	ForecastedAt       time.Time `json:"forecasted_at"`
}

// PredictiveThreatRadar projects macro-level fraud vectors across global jurisdictions and rails.
type PredictiveThreatRadar struct{}

// NewPredictiveThreatRadar initializes the threat radar.
func NewPredictiveThreatRadar() *PredictiveThreatRadar {
	return &PredictiveThreatRadar{}
}

// GenerateRadarForecast generates 7-day, 30-day, and 90-day predictive threat horizons.
func (r *PredictiveThreatRadar) GenerateRadarForecast() []ThreatRadarHorizon {
	now := time.Now().UTC()

	return []ThreatRadarHorizon{
		{
			DaysHorizon:        7,
			PrimaryThreatType:  "RESIDENTIAL_BOT_CARDING_SURGE",
			ThreatProbability:  0.92,
			TargetRegions:      []string{"US-EAST", "EU-CENTRAL"},
			ProjectedAttackers: 350,
			RecommendedAction:  "Enforce hardware-level WebAuthn checks on checkout flows",
			ForecastedAt:       now,
		},
		{
			DaysHorizon:        30,
			PrimaryThreatType:  "AI_GENERATED_SYNTHETIC_IDENTITY_WAVE",
			ThreatProbability:  0.85,
			TargetRegions:      []string{"APAC", "LATAM"},
			ProjectedAttackers: 1200,
			RecommendedAction:  "Deploy graph identity clustering across newly opened accounts",
			ForecastedAt:       now,
		},
		{
			DaysHorizon:        90,
			PrimaryThreatType:  "CROSS_RAIL_MULE_LAUNDERING_EXPANSION",
			ThreatProbability:  0.78,
			TargetRegions:      []string{"GLOBAL"},
			ProjectedAttackers: 4500,
			RecommendedAction:  "Share anonymized federated laundering signatures across consortium banks",
			ForecastedAt:       now,
		},
	}
}
