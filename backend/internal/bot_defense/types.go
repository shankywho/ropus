package bot_defense

import (
	"time"
)

// AutomationRiskLevel categorizes the severity of detected robotic or scripted activity.
type AutomationRiskLevel string

const (
	RiskLevelLow      AutomationRiskLevel = "LOW"      // Score < 0.30: Normal human-like behavior
	RiskLevelMedium   AutomationRiskLevel = "MEDIUM"   // Score 0.30-0.69: Elevated scrutiny / telemetry
	RiskLevelHigh     AutomationRiskLevel = "HIGH"     // Score 0.70-0.89: High likelihood of automated abuse
	RiskLevelCritical AutomationRiskLevel = "CRITICAL" // Score >= 0.90: Immediate containment / block
)

// DefenseAction dictates the operational recommendation from the bot layer.
type DefenseAction string

const (
	ActionAllow       DefenseAction = "ALLOW"
	ActionScrutinize  DefenseAction = "SCRUTINIZE"
	ActionElevateRisk DefenseAction = "ELEVATE_RISK"
	ActionContain     DefenseAction = "CONTAIN"
)

// BotDefenseContext contains all observable request attributes for automation analysis.
type BotDefenseContext struct {
	TenantID          string                 `json:"tenant_id"`
	PlanTier          string                 `json:"plan_tier"` // "STARTER", "GROWTH", "ENTERPRISE"
	TransactionID     string                 `json:"transaction_id"`
	AccountID         string                 `json:"account_id"`
	DeviceFingerprint string                 `json:"device_fingerprint"`
	IPAddress         string                 `json:"ip_address"`
	UserAgent         string                 `json:"user_agent"`
	CardHash          string                 `json:"card_hash"`
	Amount            float64                `json:"amount"`
	Currency          string                 `json:"currency"`
	Nonce             string                 `json:"nonce,omitempty"`
	Timestamp         time.Time              `json:"timestamp"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// BotSignals captures the granular observable metrics computed during evaluation.
type BotSignals struct {
	// Cadence & Timing Metrics
	InterArrivalTimeMeanMs float64 `json:"inter_arrival_time_mean_ms"`
	InterArrivalTimeStdDev float64 `json:"inter_arrival_time_std_dev"`
	CoefficientOfVariation float64 `json:"coefficient_of_variation"` // Low (<0.15) indicates robotic cadence
	RequestCadenceHz       float64 `json:"request_cadence_hz"`        // Instantaneous requests/second
	IsDeterministicCadence bool    `json:"is_deterministic_cadence"`  // True if fixed timer loop detected
	IsBurstPacing          bool    `json:"is_burst_pacing"`           // True if rapid burst (<200ms)

	// Entity Fan-Out & Clustering
	AccountFanOut1h     int64 `json:"account_fan_out_1h"`     // Accounts accessed by this Device in 1h
	DeviceFanOut1h      int64 `json:"device_fan_out_1h"`      // Devices used by this Account in 1h
	IPDeviceFanOut5m    int64 `json:"ip_device_fan_out_5m"`   // Devices observed from this IP in 5m
	CardTestingTokens5m int64 `json:"card_testing_tokens_5m"` // Distinct card hashes tested in 5m
	IsCardTestingProbing bool  `json:"is_card_testing_probing"`// Small amounts + high unique cards

	// Replay & Header Anomalies
	ReplayDetected bool    `json:"replay_detected"`
	TimeDriftSec   float64 `json:"time_drift_sec"`
	IsHeadlessUA   bool    `json:"is_headless_ua"`

	// Layered Rate Limiting Flags
	TenantLimited  bool `json:"tenant_limited"`
	IPLimited      bool `json:"ip_limited"`
	DeviceLimited  bool `json:"device_limited"`
	AccountLimited bool `json:"account_limited"`
	CardLimited    bool `json:"card_limited"`
}

// BotDefenseResult encapsulates the composite score, triggered rules, and telemetry.
type BotDefenseResult struct {
	AutomationRiskScore float64             `json:"automation_risk_score"` // 0.00 to 1.00
	RiskLevel           AutomationRiskLevel `json:"risk_level"`
	RecommendedAction   DefenseAction       `json:"recommended_action"`
	TriggeredRules      []string            `json:"triggered_rules"`
	ObservedPatterns    []string            `json:"observed_patterns"`
	Signals             BotSignals          `json:"signals"`
	EvaluatedAt         time.Time           `json:"evaluated_at"`
	LatencyMs           float64             `json:"latency_ms"`
	IsDegraded          bool                `json:"is_degraded"`
	DegradeReason       string              `json:"degrade_reason,omitempty"`
}
