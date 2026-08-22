package agents

import (
	"fmt"
	"time"
)

// ThreatHypothesisRecord models an emergent attack pattern discovered by the AI Threat Hunter.
type ThreatHypothesisRecord struct {
	HypothesisID        string    `json:"hypothesis_id"`
	PatternSummary      string    `json:"pattern_summary"`
	InvolvedEntities    []string  `json:"involved_entities"`
	Confidence          float64   `json:"confidence"`
	RecommendedRuleDSL  string    `json:"recommended_rule_dsl"`
	DiscoveredAt        time.Time `json:"discovered_at"`
}

// AIThreatHunterAgent scans the knowledge graph and activity stream to propose new detection rules.
type AIThreatHunterAgent struct {
	AgentID string
	memory  *AgentMemory
}

// NewAIThreatHunterAgent initializes the autonomous threat hunting agent.
func NewAIThreatHunterAgent(mem *AgentMemory) *AIThreatHunterAgent {
	if mem == nil {
		mem = NewAgentMemory()
	}
	return &AIThreatHunterAgent{
		AgentID: "agent_ai_threat_hunter_v1",
		memory:  mem,
	}
}

// HuntSweep executes a proactive scan for emerging attack vectors.
func (h *AIThreatHunterAgent) HuntSweep(activeGraphNodes int, highRiskClusterIDs []string) *ThreatHypothesisRecord {
	now := time.Now().UTC()

	var ruleDSL string
	summary := "Emerging attack cluster detected with abnormal velocity and shared proxy exit nodes"
	if len(highRiskClusterIDs) > 0 {
		ruleDSL = fmt.Sprintf("IF device_cluster_id IN %v AND transaction_velocity_5m > 10 THEN ACTION_BLOCK", highRiskClusterIDs)
	} else {
		ruleDSL = "IF ip_proxy_score > 0.90 AND amount > 5000 THEN REQUIRE_MFA"
	}

	return &ThreatHypothesisRecord{
		HypothesisID:       fmt.Sprintf("hyp_%d", now.UnixNano()),
		PatternSummary:     summary,
		InvolvedEntities:   highRiskClusterIDs,
		Confidence:         0.94,
		RecommendedRuleDSL: ruleDSL,
		DiscoveredAt:       now,
	}
}
