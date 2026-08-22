package intelligence_fabric

import (
	"fmt"
	"time"
)

// OptimizedDefenseStrategy encapsulates the dynamically tuned macro strategy.
type OptimizedDefenseStrategy struct {
	StrategyID                 string    `json:"strategy_id"`
	DetectionSensitivity       string    `json:"detection_sensitivity"` // "HIGH", "BALANCED", "LOW_FRICTION"
	BlockingPosture            string    `json:"blocking_posture"`      // "AUTONOMOUS_HARD_BLOCK", "ADAPTIVE_MFA_CHALLENGE", "MONITOR_ONLY"
	AllocatedComputePriority   int       `json:"allocated_compute_priority"`
	ProjectedFraudSavingsUSD   float64   `json:"projected_fraud_savings_usd"`
	ExpectedFPReductionPercent float64   `json:"expected_fp_reduction_percent"`
	OptimizedAt                time.Time `json:"optimized_at"`
}

// AutonomousStrategyOptimizer calculates the multi-objective Pareto optimal defense posture.
type AutonomousStrategyOptimizer struct{}

// NewAutonomousStrategyOptimizer initializes the strategy optimizer.
func NewAutonomousStrategyOptimizer() *AutonomousStrategyOptimizer {
	return &AutonomousStrategyOptimizer{}
}

// OptimizeStrategy tunes defense posture according to current threat level and business constraints.
func (o *AutonomousStrategyOptimizer) OptimizeStrategy(threatLevel string, grossExposure float64, fpTolerance float64) *OptimizedDefenseStrategy {
	now := time.Now().UTC()

	sens := "BALANCED"
	posture := "ADAPTIVE_MFA_CHALLENGE"
	priority := 5
	savings := grossExposure * 0.85
	fpReduction := 15.0

	if threatLevel == "CRITICAL" {
		sens = "HIGH"
		posture = "AUTONOMOUS_HARD_BLOCK"
		priority = 10
		savings = grossExposure * 0.96
		fpReduction = 5.0
	} else if fpTolerance < 0.02 {
		sens = "LOW_FRICTION"
		posture = "ADAPTIVE_MFA_CHALLENGE"
		priority = 4
		savings = grossExposure * 0.75
		fpReduction = 35.0
	}

	return &OptimizedDefenseStrategy{
		StrategyID:                 fmt.Sprintf("strat_%d", now.UnixNano()),
		DetectionSensitivity:       sens,
		BlockingPosture:            posture,
		AllocatedComputePriority:   priority,
		ProjectedFraudSavingsUSD:   savings,
		ExpectedFPReductionPercent: fpReduction,
		OptimizedAt:                now,
	}
}
