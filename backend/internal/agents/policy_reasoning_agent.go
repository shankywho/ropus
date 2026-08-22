package agents

// PolicyEvaluationOutcome details the compliance and governance assessment for a candidate defense action.
type PolicyEvaluationOutcome struct {
	IsAllowed            bool     `json:"is_allowed"`
	ApplicablePolicies   []string `json:"applicable_policies"`
	CustomerImpactLevel  string   `json:"customer_impact_level"` // "LOW", "MEDIUM", "HIGH"
	ComplianceViolations []string `json:"compliance_violations,omitempty"`
	RequiresHumanSignOff bool     `json:"requires_human_sign_off"`
}

// PolicyReasoningAgent verifies whether proposed autonomous containment complies with corporate policies and banking regulations.
type PolicyReasoningAgent struct {
	AgentID string
}

// NewPolicyReasoningAgent initializes the policy reasoning agent.
func NewPolicyReasoningAgent() *PolicyReasoningAgent {
	return &PolicyReasoningAgent{AgentID: "agent_policy_reasoner_v1"}
}

// EvaluateActionCompliance checks regulatory requirements and customer impact for an intended containment control.
func (a *PolicyReasoningAgent) EvaluateActionCompliance(actionType string, estimatedCustomerImpact int, confidence float64) *PolicyEvaluationOutcome {
	var policies []string
	var violations []string
	impactLevel := "LOW"
	requiresSignOff := false

	policies = append(policies, "FCRA_ADVERSE_ACTION_POLICY", "MODEL_RISK_GOVERNANCE_TIER_1")

	if estimatedCustomerImpact > 500 {
		impactLevel = "HIGH"
		requiresSignOff = true
		policies = append(policies, "EXECUTIVE_COLLATERAL_IMPACT_POLICY")
	} else if estimatedCustomerImpact > 50 {
		impactLevel = "MEDIUM"
	}

	if actionType == "PERMANENT_ACCOUNT_TERMINATION" && confidence < 0.99 {
		violations = append(violations, "VIOLATION: Permanent account closure requires 99%+ confidence or human review")
	}

	isAllowed := len(violations) == 0

	return &PolicyEvaluationOutcome{
		IsAllowed:            isAllowed,
		ApplicablePolicies:   policies,
		CustomerImpactLevel:  impactLevel,
		ComplianceViolations: violations,
		RequiresHumanSignOff: requiresSignOff,
	}
}
