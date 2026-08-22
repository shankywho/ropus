package graph

import (
	"fmt"
	"math"
	"sync"
)

// AdaptiveScoreWeights defines the multi-dimensional attribution weights for final decisioning.
type AdaptiveScoreWeights struct {
	MLModelWeight  float64 `json:"ml_model_weight"`  // default: 0.35
	GraphWeight    float64 `json:"graph_weight"`     // default: 0.25
	BehaviorWeight float64 `json:"behavior_weight"`  // default: 0.20
	ThreatWeight   float64 `json:"threat_weight"`    // default: 0.15
	RulesWeight    float64 `json:"rules_weight"`     // default: 0.05
}

// AdaptiveRiskResult encapsulates the synthesized multi-layered risk decision.
type AdaptiveRiskResult struct {
	FinalScore        float64                 `json:"final_score"` // 0.0 to 1.0
	Decision          string                  `json:"decision"`    // "ALLOW", "MANUAL_REVIEW", "BLOCK"
	MLScore           float64                 `json:"ml_score"`
	GraphScore        float64                 `json:"graph_score"`
	BehaviorScore     float64                 `json:"behavior_score"`
	ThreatScore       float64                 `json:"threat_score"`
	RulesScore        float64                 `json:"rules_score"`
	ContributingReasons []AdaptiveRiskReason `json:"contributing_reasons"`
}

// AdaptiveRiskReason details explanations categorized by sub-engine type.
type AdaptiveRiskReason struct {
	Type    string `json:"type"` // "GRAPH", "BEHAVIOR", "THREAT_INTEL", "RULES", "ML_MODEL"
	Message string `json:"message"`
}

// AdaptiveRiskEngine synthesizes ML, Graph, Behavioral, Threat Intel, and Rule outputs.
type AdaptiveRiskEngine struct {
	mu             sync.RWMutex
	weights        AdaptiveScoreWeights
	graphExtractor *GraphFeatureExtractor
	behaviorEngine *BehaviorEngine
	threatEngine   *ThreatIntelligenceEngine
}

// NewAdaptiveRiskEngine initializes the adaptive multi-layered risk engine.
func NewAdaptiveRiskEngine(
	gx *GraphFeatureExtractor,
	be *BehaviorEngine,
	te *ThreatIntelligenceEngine,
) *AdaptiveRiskEngine {
	return &AdaptiveRiskEngine{
		weights: AdaptiveScoreWeights{
			MLModelWeight:  0.35,
			GraphWeight:    0.25,
			BehaviorWeight: 0.20,
			ThreatWeight:   0.15,
			RulesWeight:    0.05,
		},
		graphExtractor: gx,
		behaviorEngine: be,
		threatEngine:   te,
	}
}

// EvaluateAdaptiveRisk runs comprehensive multi-engine evaluation.
func (e *AdaptiveRiskEngine) EvaluateAdaptiveRisk(
	userID, deviceFingerprint, ipAddress, emailDomain, location string,
	amount, mlModelScore, rulesScore float64,
) (*AdaptiveRiskResult, error) {
	e.mu.RLock()
	w := e.weights
	e.mu.RUnlock()

	var reasons []AdaptiveRiskReason

	// 1. Graph Risk Component
	graphScore := 0.0
	if e.graphExtractor != nil {
		gf, _ := e.graphExtractor.ExtractFeatures(userID, deviceFingerprint, ipAddress)
		if gf != nil {
			graphScore = gf.GraphRiskScore
			if gf.FraudNeighborCount > 0 {
				reasons = append(reasons, AdaptiveRiskReason{
					Type:    "GRAPH",
					Message: fmt.Sprintf("Device/IP connected to %d confirmed fraudulent entities", gf.FraudNeighborCount),
				})
			}
		}
	}

	// 2. Behavioral Anomaly Component
	behaviorScore := 0.0
	if e.behaviorEngine != nil {
		bScore, bAnomalies := e.behaviorEngine.EvaluateBehavior(userID, amount, deviceFingerprint, ipAddress, location)
		behaviorScore = bScore
		for _, a := range bAnomalies {
			reasons = append(reasons, AdaptiveRiskReason{Type: "BEHAVIOR", Message: a})
		}
	}

	// 3. Threat Intelligence Component
	threatScore := 0.0
	if e.threatEngine != nil {
		tScore, tMatches := e.threatEngine.CheckThreat(ipAddress, deviceFingerprint, emailDomain)
		threatScore = tScore
		for _, m := range tMatches {
			reasons = append(reasons, AdaptiveRiskReason{Type: "THREAT_INTEL", Message: m})
		}
	}

	// 4. Synthesize Final Weighted Score
	finalScore := w.MLModelWeight*mlModelScore +
		w.GraphWeight*graphScore +
		w.BehaviorWeight*behaviorScore +
		w.ThreatWeight*threatScore +
		w.RulesWeight*rulesScore

	// Severe threat or graph override clamps to high risk
	if threatScore >= 0.90 || graphScore >= 0.90 {
		finalScore = math.Max(finalScore, 0.92)
	}

	if finalScore > 1.0 {
		finalScore = 1.0
	}

	// Decision threshold classification
	decision := "ALLOW"
	if finalScore >= 0.85 {
		decision = "BLOCK"
	} else if finalScore >= 0.50 {
		decision = "MANUAL_REVIEW"
	}

	return &AdaptiveRiskResult{
		FinalScore:          finalScore,
		Decision:            decision,
		MLScore:             mlModelScore,
		GraphScore:          graphScore,
		BehaviorScore:       behaviorScore,
		ThreatScore:         threatScore,
		RulesScore:          rulesScore,
		ContributingReasons: reasons,
	}, nil
}
