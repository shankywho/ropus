package crime_intelligence

import (
	"fmt"
	"time"
)

// MoneyFlowPatternType identifies specific laundering topological signatures.
type MoneyFlowPatternType string

const (
	FlowLayeringRapidHops MoneyFlowPatternType = "LAYERING_RAPID_HOPS"
	FlowCircularMovement  MoneyFlowPatternType = "CIRCULAR_MOVEMENT"
	FlowStructuringSmurf  MoneyFlowPatternType = "STRUCTURING_SMURFING"
	FlowMuleChainCashout  MoneyFlowPatternType = "MULE_CHAIN_CASHOUT"
)

// MoneyFlowAnalysisResult details detected laundering structures.
type MoneyFlowAnalysisResult struct {
	AnalysisID          string               `json:"analysis_id"`
	PatternDetected     MoneyFlowPatternType `json:"pattern_detected"`
	HashedSourceAccount string               `json:"hashed_source_account"`
	HashedDestination   string               `json:"hashed_destination"`
	TotalAmountUSD      float64              `json:"total_amount_usd"`
	HopCount            int                  `json:"hop_count"`
	LaunderingRiskScore float64              `json:"laundering_risk_score"`
	FlaggedReason       string               `json:"flagged_reason"`
	AnalyzedAt          time.Time            `json:"analyzed_at"`
}

// MoneyFlowAnalyzer detects illicit fund movements and laundering chains across accounts and rails.
type MoneyFlowAnalyzer struct{}

// NewMoneyFlowAnalyzer initializes the money flow analyzer.
func NewMoneyFlowAnalyzer() *MoneyFlowAnalyzer {
	return &MoneyFlowAnalyzer{}
}

// AnalyzeFlowTrajectory traces fund velocity and detects layering topologies.
func (a *MoneyFlowAnalyzer) AnalyzeFlowTrajectory(rawSrc, rawDst string, amount float64, hopCount int, timeSpanMinutes int) *MoneyFlowAnalysisResult {
	now := time.Now().UTC()
	hashedSrc := HashID(rawSrc)
	hashedDst := HashID(rawDst)

	pattern := FlowMuleChainCashout
	risk := 0.70
	reason := "Standard multi-hop fund transfer"

	if hopCount >= 3 && timeSpanMinutes <= 10 {
		pattern = FlowLayeringRapidHops
		risk = 0.96
		reason = fmt.Sprintf("High velocity layering: %d hops in %d minutes across mule routing nodes", hopCount, timeSpanMinutes)
	} else if rawSrc == rawDst && hopCount >= 2 {
		pattern = FlowCircularMovement
		risk = 0.98
		reason = "Circular fund rotation detected: Funds returned to origin entity via intermediaries"
	}

	return &MoneyFlowAnalysisResult{
		AnalysisID:          fmt.Sprintf("mfa_%d", now.UnixNano()),
		PatternDetected:     pattern,
		HashedSourceAccount: hashedSrc,
		HashedDestination:   hashedDst,
		TotalAmountUSD:      amount,
		HopCount:            hopCount,
		LaunderingRiskScore: risk,
		FlaggedReason:       reason,
		AnalyzedAt:          now,
	}
}
