package agents

import (
	"fmt"
)

// AgentSecurityGuardrails validates authority, evidence veracity, and policy boundaries for autonomous agents.
type AgentSecurityGuardrails struct {
	AllowedAgentRoles map[string]bool
}

// NewAgentSecurityGuardrails initializes agent security guardrails.
func NewAgentSecurityGuardrails() *AgentSecurityGuardrails {
	return &AgentSecurityGuardrails{
		AllowedAgentRoles: map[string]bool{
			string(AgentFraudInvestigator): true,
			string(AgentThreatHunter):      true,
			string(AgentRiskOptimizer):     true,
			string(AgentCompliance):        true,
			string(AgentResponse):          true,
			string(AgentDataQuality):       true,
		},
	}
}

// ValidateAgentAction checks identity, permission, and hallucination prevention invariants.
func (g *AgentSecurityGuardrails) ValidateAgentAction(agentType AgentType, action string, confidence float64, evidenceChain []string) error {
	if !g.AllowedAgentRoles[string(agentType)] {
		return fmt.Errorf("guardrail violation: unauthorized agent type '%s'", agentType)
	}

	if action == "BLOCK_AND_FREEZE" || action == "FREEZE_ACCOUNT" {
		if confidence < 0.85 {
			return fmt.Errorf("guardrail violation: irreversible containment requires >= 0.85 confidence (got %.2f)", confidence)
		}
		if len(evidenceChain) == 0 {
			return fmt.Errorf("guardrail violation: containment action rejected due to empty evidence chain (hallucination prevention)")
		}
	}

	return nil
}
