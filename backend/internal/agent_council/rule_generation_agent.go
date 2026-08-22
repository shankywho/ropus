package agent_council

import (
	"fmt"
	"time"
)

// RuleDeploymentLifecycle tracks the 5-stage automated rule pipeline.
type RuleDeploymentLifecycle string

const (
	RuleStageDiscovery          RuleDeploymentLifecycle = "DISCOVERY"
	RuleStageSimulation         RuleDeploymentLifecycle = "SIMULATION"
	RuleStageShadowEvaluation   RuleDeploymentLifecycle = "SHADOW_EVALUATION"
	RuleStageGovernanceApproval RuleDeploymentLifecycle = "GOVERNANCE_APPROVAL"
	RuleStageProduction         RuleDeploymentLifecycle = "PRODUCTION"
)

// GeneratedRuleCandidate represents a candidate detection rule synthesized by AI.
type GeneratedRuleCandidate struct {
	RuleID          string                  `json:"rule_id"`
	RuleName        string                  `json:"rule_name"`
	RuleDSL         string                  `json:"rule_dsl"`
	TriggerPattern  string                  `json:"trigger_pattern"`
	EstimatedFPDelta float64                `json:"estimated_fp_delta"`
	Stage           RuleDeploymentLifecycle `json:"stage"`
	DiscoveredAt    time.Time               `json:"discovered_at"`
}

// AutonomousRuleGenerationAgent synthesizes new deterministic rules from emergent cluster patterns.
type AutonomousRuleGenerationAgent struct {
	AgentID string
}

// NewAutonomousRuleGenerationAgent initializes the rule synthesis agent.
func NewAutonomousRuleGenerationAgent() *AutonomousRuleGenerationAgent {
	return &AutonomousRuleGenerationAgent{AgentID: "agent_rule_generator_v1"}
}

// SynthesizeRule creates a candidate rule and places it into the discovery phase.
func (a *AutonomousRuleGenerationAgent) SynthesizeRule(patternName, conditionDSL string) *GeneratedRuleCandidate {
	now := time.Now().UTC()
	return &GeneratedRuleCandidate{
		RuleID:           fmt.Sprintf("gen_rule_%d", now.UnixNano()),
		RuleName:         fmt.Sprintf("AutoRule_%s", patternName),
		RuleDSL:          conditionDSL,
		TriggerPattern:   patternName,
		EstimatedFPDelta: -0.012,
		Stage:            RuleStageDiscovery,
		DiscoveredAt:     now,
	}
}

// AdvancePipeline moves the rule candidate through the mandatory validation pipeline.
func (a *AutonomousRuleGenerationAgent) AdvancePipeline(r *GeneratedRuleCandidate) {
	switch r.Stage {
	case RuleStageDiscovery:
		r.Stage = RuleStageSimulation
	case RuleStageSimulation:
		r.Stage = RuleStageShadowEvaluation
	case RuleStageShadowEvaluation:
		r.Stage = RuleStageGovernanceApproval
	case RuleStageGovernanceApproval:
		r.Stage = RuleStageProduction
	}
}
