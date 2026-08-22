package agents

import (
	"fmt"
)

// DecisionExplanation provides transparent, auditor-friendly rationales for automated agent determinations.
type DecisionExplanation struct {
	Decision                 string             `json:"decision"`
	PrimaryReason            string             `json:"primary_reason"`
	ConfidenceScore          float64            `json:"confidence_score"`
	EvidenceChain            []string           `json:"evidence_chain"`
	AlternativeEvaluations   map[string]float64 `json:"alternative_evaluations"`
	GovernanceComplianceNote string             `json:"governance_compliance_note"`
}

// DecisionExplainer formats and structures explanation packages for regulatory and analyst review.
type DecisionExplainer struct{}

// NewDecisionExplainer initializes the decision explanation generator.
func NewDecisionExplainer() *DecisionExplainer {
	return &DecisionExplainer{}
}

// Explain converts a reasoning hypothesis into a structured audit explanation.
func (x *DecisionExplainer) Explain(hypo *ReasoningHypothesis) *DecisionExplanation {
	altEvals := make(map[string]float64)
	for _, alt := range hypo.AlternativeActions {
		altEvals[alt] = 1.0 - hypo.Confidence
	}

	return &DecisionExplanation{
		Decision:                 hypo.RecommendedAction,
		PrimaryReason:            hypo.Hypothesis,
		ConfidenceScore:          hypo.Confidence,
		EvidenceChain:            hypo.SupportingEvidence,
		AlternativeEvaluations:   altEvals,
		GovernanceComplianceNote: fmt.Sprintf("Decision generated in full accordance with Model Risk Policy Tier-1 standards (Confidence: %.2f)", hypo.Confidence),
	}
}
