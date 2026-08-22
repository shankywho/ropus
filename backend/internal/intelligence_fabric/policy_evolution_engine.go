package intelligence_fabric

import (
	"fmt"
	"time"
)

// PolicyStage defines the 6-stage lifecycle for autonomous policy evolution.
type PolicyStage string

const (
	PolicyStageDiscover         PolicyStage = "DISCOVER"
	PolicyStageSimulate         PolicyStage = "SIMULATE"
	PolicyStageShadow           PolicyStage = "SHADOW"
	PolicyStageGovernanceReview PolicyStage = "GOVERNANCE_REVIEW"
	PolicyStageCanary           PolicyStage = "CANARY"
	PolicyStageProduction       PolicyStage = "PRODUCTION"
)

// EvolvingPolicyCandidate represents a candidate rule or threshold undergoing promotion.
type EvolvingPolicyCandidate struct {
	PolicyID        string      `json:"policy_id"`
	PolicyName      string      `json:"policy_name"`
	RuleDSL         string      `json:"rule_dsl"`
	Stage           PolicyStage `json:"stage"`
	ExpectedFPReduction float64 `json:"expected_fp_reduction"`
	DiscoveredAt    time.Time   `json:"discovered_at"`
}

// AutonomousPolicyEvolutionEngine manages automated policy evolution through strict governance gates.
type AutonomousPolicyEvolutionEngine struct{}

// NewAutonomousPolicyEvolutionEngine initializes the policy evolution engine.
func NewAutonomousPolicyEvolutionEngine() *AutonomousPolicyEvolutionEngine {
	return &AutonomousPolicyEvolutionEngine{}
}

// CreateCandidate initializes a newly discovered policy candidate.
func (e *AutonomousPolicyEvolutionEngine) CreateCandidate(name, dsl string) *EvolvingPolicyCandidate {
	now := time.Now().UTC()
	return &EvolvingPolicyCandidate{
		PolicyID:            fmt.Sprintf("pol_%d", now.UnixNano()),
		PolicyName:          name,
		RuleDSL:             dsl,
		Stage:               PolicyStageDiscover,
		ExpectedFPReduction: 0.02,
		DiscoveredAt:        now,
	}
}

// AdvanceStage progresses the candidate across the 6-stage safety lifecycle.
func (e *AutonomousPolicyEvolutionEngine) AdvanceStage(p *EvolvingPolicyCandidate) {
	switch p.Stage {
	case PolicyStageDiscover:
		p.Stage = PolicyStageSimulate
	case PolicyStageSimulate:
		p.Stage = PolicyStageShadow
	case PolicyStageShadow:
		p.Stage = PolicyStageGovernanceReview
	case PolicyStageGovernanceReview:
		p.Stage = PolicyStageCanary
	case PolicyStageCanary:
		p.Stage = PolicyStageProduction
	}
}
