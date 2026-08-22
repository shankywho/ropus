package cases

import (
	"math"
)

// CasePrioritizer calculates dynamic risk priority based on multi-dimensional forensic severity.
type CasePrioritizer struct {
	CriticalAmountThreshold float64
	HighAmountThreshold     float64
}

// NewCasePrioritizer initializes the prioritization engine.
func NewCasePrioritizer() *CasePrioritizer {
	return &CasePrioritizer{
		CriticalAmountThreshold: 50000.0,
		HighAmountThreshold:     10000.0,
	}
}

// CalculatePriority computes the operational urgency tier and numeric composite score.
func (p *CasePrioritizer) CalculatePriority(
	riskScore float64,
	exposureAmount float64,
	connectedAccountsCount int,
	velocityCount int,
	hasThreatIntelMatch bool,
) (CasePriority, float64) {
	// 1. Financial Impact Factor (0.0 to 1.0)
	finFactor := math.Min(1.0, exposureAmount/p.CriticalAmountThreshold)

	// 2. Network Size Factor (0.0 to 1.0)
	networkFactor := math.Min(1.0, float64(connectedAccountsCount)/10.0)

	// 3. Velocity Factor (0.0 to 1.0)
	velocityFactor := math.Min(1.0, float64(velocityCount)/10.0)

	// 4. Threat Intel Factor
	threatFactor := 0.0
	if hasThreatIntelMatch {
		threatFactor = 1.0
	}

	// Composite Priority Score
	score := (riskScore * 0.30) +
		(finFactor * 0.30) +
		(networkFactor * 0.20) +
		(velocityFactor * 0.10) +
		(threatFactor * 0.10)

	// Immediate Critical Override for Massive Syndicate or Extreme Exposure
	if (exposureAmount >= p.CriticalAmountThreshold && riskScore >= 0.80) ||
		(connectedAccountsCount >= 20 && riskScore >= 0.80) {
		return PriorityCritical, math.Max(score, 0.95)
	}

	if score >= 0.80 {
		return PriorityCritical, score
	} else if score >= 0.60 {
		return PriorityHigh, score
	} else if score >= 0.35 {
		return PriorityMedium, score
	}
	return PriorityLow, score
}
