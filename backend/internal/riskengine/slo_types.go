package riskengine

import "time"

// SLOStatus represents the health status of an individual Service Level Objective.
type SLOStatus string

const (
	SLOStatusHealthy  SLOStatus = "HEALTHY"
	SLOStatusWarning  SLOStatus = "WARNING"
	SLOStatusBreached SLOStatus = "BREACHED"
	SLOStatusUnknown  SLOStatus = "UNKNOWN"
)

// SLODefinition defines the target threshold and parameters for an SLO.
type SLODefinition struct {
	ID                string        `json:"id"`
	Name              string        `json:"name"`
	Description       string        `json:"description"`
	Target            float64       `json:"target"`
	WarningThreshold  float64       `json:"warning_threshold"`
	BreachThreshold   float64       `json:"breach_threshold"`
	WindowDuration    time.Duration `json:"window_duration"`
	IsUpperBound      bool          `json:"is_upper_bound"` // true for latency/error rate (lower is better), false for availability (higher is better)
	Unit              string        `json:"unit"`
}

// SLOMetricRecord represents the real-time computed health, error budget, and burn rate for an SLO.
type SLOMetricRecord struct {
	SLOID                string    `json:"slo_id"`
	Name                 string    `json:"name"`
	CurrentValue         float64   `json:"current_value"`
	Target               float64   `json:"target"`
	Status               SLOStatus `json:"status"`
	ErrorBudget          float64   `json:"error_budget"`           // Total error budget fraction (e.g. 0.001 for 99.9%)
	ErrorBudgetRemaining float64   `json:"error_budget_remaining"` // 0.0% to 100.0%
	BurnRate             float64   `json:"burn_rate"`              // Multiple of acceptable burn (1.0 = on track to exhaust budget over window)
	MeasurementWindow    string    `json:"measurement_window"`
	SampleCount          int64     `json:"sample_count"`
	ViolationsCount      int64     `json:"violations_count"`
	CalculatedAt         time.Time `json:"calculated_at"`
	Violations           []string  `json:"violations,omitempty"`
}

// SLOSummary provides an aggregate snapshot of all SLO evaluations across the system.
type SLOSummary struct {
	OverallStatus        SLOStatus                  `json:"overall_status"`
	TotalSLOs            int                        `json:"total_slos"`
	HealthyCount         int                        `json:"healthy_count"`
	WarningCount         int                        `json:"warning_count"`
	BreachedCount        int                        `json:"breached_count"`
	CalculatedAt         time.Time                  `json:"calculated_at"`
	Measurements         map[string]SLOMetricRecord `json:"measurements"`
}
