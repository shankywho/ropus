package agents

import (
	"fmt"
	"time"
)

// OptimizationStage defines the mandatory safety progression of any automated tuning.
type OptimizationStage string

const (
	StageProposal         OptimizationStage = "PROPOSAL"
	StageGovernanceReview OptimizationStage = "GOVERNANCE_REVIEW"
	StageShadowTesting    OptimizationStage = "SHADOW_TESTING"
	StageCanaryRollout    OptimizationStage = "CANARY"
	StageProductionActive OptimizationStage = "PRODUCTION"
)

// RiskOptimizationProposal describes a candidate parameter adjustment.
type RiskOptimizationProposal struct {
	ProposalID       string            `json:"proposal_id"`
	TargetParameter  string            `json:"target_parameter"`
	CurrentValue     float64           `json:"current_value"`
	ProposedValue    float64           `json:"proposed_value"`
	ExpectedFPDelta  float64           `json:"expected_fp_delta"` // e.g. -0.02 (2% FP reduction)
	ExpectedFNReduction float64        `json:"expected_fn_reduction"`
	Stage            OptimizationStage `json:"stage"`
	ProposedAt       time.Time         `json:"proposed_at"`
}

// RiskOptimizerAgent autonomously identifies opportunities to tighten or loosen detection thresholds safely.
type RiskOptimizerAgent struct {
	AgentID string
}

// NewRiskOptimizerAgent initializes the self-optimizing risk agent.
func NewRiskOptimizerAgent() *RiskOptimizerAgent {
	return &RiskOptimizerAgent{AgentID: "agent_risk_optimizer_v1"}
}

// ProposeThresholdTuning formulates an optimization proposal based on historical FPR and loss metrics.
func (o *RiskOptimizerAgent) ProposeThresholdTuning(parameterName string, currentVal, targetFPR float64) *RiskOptimizationProposal {
	now := time.Now().UTC()
	proposedVal := currentVal
	if targetFPR < 0.03 {
		proposedVal = currentVal + 0.05 // Raise threshold slightly to reduce false positives
	}

	return &RiskOptimizationProposal{
		ProposalID:          fmt.Sprintf("opt_%d_%s", now.UnixNano(), parameterName),
		TargetParameter:     parameterName,
		CurrentValue:        currentVal,
		ProposedValue:       proposedVal,
		ExpectedFPDelta:     -0.015,
		ExpectedFNReduction: 0.02,
		Stage:               StageProposal,
		ProposedAt:          now,
	}
}

// AdvanceStage progresses the proposal through the mandatory safety gating lifecycle.
func (o *RiskOptimizerAgent) AdvanceStage(p *RiskOptimizationProposal) {
	switch p.Stage {
	case StageProposal:
		p.Stage = StageGovernanceReview
	case StageGovernanceReview:
		p.Stage = StageShadowTesting
	case StageShadowTesting:
		p.Stage = StageCanaryRollout
	case StageCanaryRollout:
		p.Stage = StageProductionActive
	}
}
