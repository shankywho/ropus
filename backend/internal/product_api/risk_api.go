package product_api

import (
	"time"
)

// EvaluateRiskRequest represents the customer API payload for evaluating a transaction.
type EvaluateRiskRequest struct {
	TransactionID string                 `json:"transaction_id"`
	UserID        string                 `json:"user_id"`
	Amount        float64                `json:"amount"`
	Currency      string                 `json:"currency"`
	Merchant      string                 `json:"merchant"`
	Device        DeviceDetails          `json:"device"`
	Location      LocationDetails        `json:"location"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// DeviceDetails captures client hardware and network telemetry.
type DeviceDetails struct {
	DeviceFingerprint string `json:"device_fingerprint"`
	IPAddress         string `json:"ip_address"`
	UserAgent         string `json:"user_agent"`
	IsEmulator        bool   `json:"is_emulator"`
	IsVPN             bool   `json:"is_vpn"`
}

// LocationDetails captures geolocation telemetry.
type LocationDetails struct {
	Country string  `json:"country"`
	City    string  `json:"city"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
}

// RiskDecision represents the actionable verdict.
type RiskDecision string

const (
	DecisionApprove   RiskDecision = "APPROVE"
	DecisionReview    RiskDecision = "REVIEW"
	DecisionChallenge RiskDecision = "CHALLENGE"
	DecisionBlock     RiskDecision = "BLOCK"
)

// ExplanationBreakdown provides transparent factor contributions to the risk score.
type ExplanationBreakdown struct {
	GraphIntelligenceWeight  float64 `json:"graph_intelligence_weight"`  // e.g. 0.90
	BehaviorAnalysisWeight   float64 `json:"behavior_analysis_weight"`   // e.g. 0.82
	ThreatIntelligenceWeight float64 `json:"threat_intelligence_weight"` // e.g. 0.95
	MachineLearningWeight    float64 `json:"machine_learning_weight"`    // e.g. 0.88
}

// EvaluateRiskResponse represents the comprehensive customer response.
type EvaluateRiskResponse struct {
	TransactionID     string               `json:"transaction_id"`
	RiskScore         float64              `json:"risk_score"` // 0.0 to 1.0 (or 0 to 100)
	Decision          RiskDecision         `json:"decision"`
	Confidence        float64              `json:"confidence"`
	Reasons           []string             `json:"reasons"`
	HumanExplanation  string               `json:"human_explanation"`
	Breakdown         ExplanationBreakdown `json:"breakdown"`
	ModelVersion      string               `json:"model_version"`
	GraphSignals      []string             `json:"graph_signals"`
	RecommendedAction string               `json:"recommended_action"`
	EvaluatedAt       time.Time            `json:"evaluated_at"`
}
