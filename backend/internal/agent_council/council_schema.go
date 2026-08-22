package agent_council

import (
	"time"
)

// ConsensusType defines the outcome level of the AI Agent Council deliberation.
type ConsensusType string

const (
	ConsensusUnanimous                 ConsensusType = "UNANIMOUS"
	ConsensusMajority                  ConsensusType = "MAJORITY"
	ConsensusDissentRequiresHumanReview ConsensusType = "HUMAN_REVIEW_REQUIRED"
)

// AgentOpinion encapsulates a single agent's assessment and recommendation.
type AgentOpinion struct {
	AgentID           string   `json:"agent_id"`
	Role              string   `json:"role"` // "INVESTIGATOR", "THREAT_HUNTER", "RISK_OPTIMIZER", "COMPLIANCE", "RESPONSE"
	RecommendedAction string   `json:"recommended_action"`
	Confidence        float64  `json:"confidence"`
	EvidenceChain     []string `json:"evidence_chain"`
	DissentReason     string   `json:"dissent_reason,omitempty"`
}

// CouncilDecision represents the synthesized multi-agent consensus verdict.
type CouncilDecision struct {
	IncidentID       string         `json:"incident_id"`
	ConsensusAction  string         `json:"consensus_action"`
	Confidence       float64        `json:"confidence"`
	ConsensusType    ConsensusType  `json:"consensus_type"`
	SupportingAgents []string       `json:"supporting_agents"`
	DissentingAgents []string       `json:"dissenting_agents"`
	Rationale        string         `json:"rationale"`
	DeliberatedAt    time.Time      `json:"deliberated_at"`
}
