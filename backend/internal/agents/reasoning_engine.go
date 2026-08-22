package agents

import (
	"fmt"
	"math"
)

// ReasoningHypothesis represents the synthesized forensic deduction made by the AI reasoning engine.
type ReasoningHypothesis struct {
	Hypothesis            string                 `json:"hypothesis"`
	SupportingEvidence    []string               `json:"supporting_evidence"`
	Confidence            float64                `json:"confidence"`
	RecommendedAction     string                 `json:"recommended_action"`
	AlternativeActions    []string               `json:"alternative_actions"`
	ReasonCodes           []string               `json:"reason_codes"`
	FeatureAttributions   map[string]float64     `json:"feature_attributions"`
}

// FraudReasoningEngine evaluates multi-vector evidence to construct deterministic, explainable conclusions.
type FraudReasoningEngine struct{}

// NewFraudReasoningEngine initializes the reasoning engine.
func NewFraudReasoningEngine() *FraudReasoningEngine {
	return &FraudReasoningEngine{}
}

// Reason analyzes forensic vectors and returns an explainable hypothesis with an audit chain.
func (e *FraudReasoningEngine) Reason(
	mlScore float64,
	graphFraudNeighbors int,
	behaviorAnomalies []string,
	threatMatches []string,
	longTermHistoricalMatches int,
) *ReasoningHypothesis {
	var evidence []string
	var reasonCodes []string
	attributions := make(map[string]float64)

	baseScore := mlScore
	attributions["ml_model_score"] = mlScore

	if graphFraudNeighbors > 0 {
		evidence = append(evidence, fmt.Sprintf("Graph connectivity: Linked to %d confirmed fraudulent syndicate accounts", graphFraudNeighbors))
		reasonCodes = append(reasonCodes, "GRAPH_FRAUD_RING_LINKAGE")
		attributions["graph_linkage"] = 0.90
		baseScore = math.Max(baseScore, 0.88)
	}

	for _, b := range behaviorAnomalies {
		evidence = append(evidence, fmt.Sprintf("Behavioral anomaly: %s", b))
		reasonCodes = append(reasonCodes, "BEHAVIOR_PROFILE_DEVIATION")
		attributions["behavior_anomaly"] = 0.75
		baseScore = math.Min(1.0, baseScore+0.10)
	}

	for _, t := range threatMatches {
		evidence = append(evidence, fmt.Sprintf("Threat intelligence: Match on %s", t))
		reasonCodes = append(reasonCodes, "THREAT_INTEL_IOC_MATCH")
		attributions["threat_intel"] = 0.95
		baseScore = math.Max(baseScore, 0.95)
	}

	if longTermHistoricalMatches > 0 {
		evidence = append(evidence, fmt.Sprintf("Memory recall: %d historical fraud attack patterns matched", longTermHistoricalMatches))
		reasonCodes = append(reasonCodes, "HISTORICAL_PATTERN_RECALL")
		attributions["memory_recall"] = 0.85
	}

	// Determine recommended action and hypothesis
	recommended := "ALLOW"
	hypothesis := "Transaction displays normal behavioral characteristics with low risk"
	var alternatives []string

	if baseScore >= 0.90 {
		recommended = "BLOCK_AND_FREEZE"
		hypothesis = "High confidence fraud attack: Critical intersection of threat intel and syndicate graph linkages"
		alternatives = []string{"REQUIRE_STEP_UP_MFA", "MONITOR_ONLY"}
	} else if baseScore >= 0.70 {
		recommended = "REQUIRE_MFA"
		hypothesis = "Suspicious transaction: Notable behavioral variance or moderate graph risk"
		alternatives = []string{"BLOCK_TRANSACTION", "RELEASE_WITH_VELOCITY_LIMIT"}
	} else {
		alternatives = []string{"REQUIRE_MFA"}
	}

	return &ReasoningHypothesis{
		Hypothesis:          hypothesis,
		SupportingEvidence:  evidence,
		Confidence:          baseScore,
		RecommendedAction:   recommended,
		AlternativeActions:  alternatives,
		ReasonCodes:         reasonCodes,
		FeatureAttributions: attributions,
	}
}
