package agent_council

import (
	"fmt"
	"time"
)

// DebateEngine coordinates structured multi-agent debates to resolve complex, conflicting fraud scenarios.
type DebateEngine struct {
	MinUnanimousConfidence float64
}

// NewDebateEngine initializes the council debate engine.
func NewDebateEngine() *DebateEngine {
	return &DebateEngine{
		MinUnanimousConfidence: 0.90,
	}
}

// Deliberate orchestrates a multi-agent debate across independent agent opinions.
func (e *DebateEngine) Deliberate(incidentID string, opinions []AgentOpinion) *CouncilDecision {
	now := time.Now().UTC()

	if len(opinions) == 0 {
		return &CouncilDecision{
			IncidentID:      incidentID,
			ConsensusAction: "REQUIRE_HUMAN_REVIEW",
			ConsensusType:   ConsensusDissentRequiresHumanReview,
			Rationale:       "No agent opinions provided",
			DeliberatedAt:   now,
		}
	}

	// Tally action votes
	voteCounts := make(map[string]int)
	confidenceSums := make(map[string]float64)
	agentBuckets := make(map[string][]string)

	complianceVeto := false
	var complianceVetoReason string

	for _, op := range opinions {
		if op.Role == "COMPLIANCE" && op.DissentReason != "" {
			complianceVeto = true
			complianceVetoReason = op.DissentReason
		}

		voteCounts[op.RecommendedAction]++
		confidenceSums[op.RecommendedAction] += op.Confidence
		agentBuckets[op.RecommendedAction] = append(agentBuckets[op.RecommendedAction], fmt.Sprintf("%s (%s)", op.AgentID, op.Role))
	}

	// Compliance veto triggers immediate human review for high-friction actions
	if complianceVeto {
		var dissenting []string
		for _, op := range opinions {
			if op.Role == "COMPLIANCE" {
				dissenting = append(dissenting, op.AgentID)
			}
		}
		return &CouncilDecision{
			IncidentID:       incidentID,
			ConsensusAction:  "ESCALATE_TO_HUMAN_ANALYST",
			Confidence:       0.50,
			ConsensusType:    ConsensusDissentRequiresHumanReview,
			SupportingAgents: []string{"ComplianceAgent"},
			DissentingAgents: dissenting,
			Rationale:        fmt.Sprintf("Compliance objection: %s", complianceVetoReason),
			DeliberatedAt:    now,
		}
	}

	// Find dominant action
	bestAction := ""
	maxVotes := 0
	for act, count := range voteCounts {
		if count > maxVotes {
			maxVotes = count
			bestAction = act
		}
	}

	avgConf := confidenceSums[bestAction] / float64(maxVotes)
	var supporting []string
	var dissenting []string

	for _, op := range opinions {
		if op.RecommendedAction == bestAction {
			supporting = append(supporting, op.AgentID)
		} else {
			dissenting = append(dissenting, fmt.Sprintf("%s (preferred %s)", op.AgentID, op.RecommendedAction))
		}
	}

	consensusType := ConsensusMajority
	if len(dissenting) == 0 {
		consensusType = ConsensusUnanimous
	} else if float64(maxVotes) < float64(len(opinions))*0.60 {
		consensusType = ConsensusDissentRequiresHumanReview
	}

	rationale := fmt.Sprintf("Council selected '%s' with %d/%d votes (Avg Confidence: %.2f)", bestAction, maxVotes, len(opinions), avgConf)

	return &CouncilDecision{
		IncidentID:       incidentID,
		ConsensusAction:  bestAction,
		Confidence:       avgConf,
		ConsensusType:    consensusType,
		SupportingAgents: supporting,
		DissentingAgents: dissenting,
		Rationale:        rationale,
		DeliberatedAt:    now,
	}
}
