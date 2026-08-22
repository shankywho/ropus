package agents

import (
	"fmt"
	"time"
)

// FraudInvestigationDossier contains the complete forensic dossier produced by the Autonomous Investigator Agent.
type FraudInvestigationDossier struct {
	DossierID             string               `json:"dossier_id"`
	TraceID               string               `json:"trace_id"`
	TargetEntityID        string               `json:"target_entity_id"`
	ReasoningHypothesis   *ReasoningHypothesis `json:"reasoning_hypothesis"`
	Explanation           *DecisionExplanation `json:"explanation"`
	AssignedAgentID       string               `json:"assigned_agent_id"`
	InvestigationDuration float64              `json:"investigation_duration_ms"`
	CreatedAt             time.Time            `json:"created_at"`
}

// AutonomousInvestigatorAgent orchestrates autonomous forensic investigations.
type AutonomousInvestigatorAgent struct {
	AgentID          string
	memory           *AgentMemory
	reasoningEngine  *FraudReasoningEngine
	explainer        *DecisionExplainer
}

// NewAutonomousInvestigatorAgent initializes the autonomous investigator.
func NewAutonomousInvestigatorAgent(mem *AgentMemory) *AutonomousInvestigatorAgent {
	if mem == nil {
		mem = NewAgentMemory()
	}
	return &AutonomousInvestigatorAgent{
		AgentID:         "agent_auto_investigator_v1",
		memory:          mem,
		reasoningEngine: NewFraudReasoningEngine(),
		explainer:       NewDecisionExplainer(),
	}
}

// InvestigateIncident performs an end-to-end autonomous forensic inquiry.
func (a *AutonomousInvestigatorAgent) InvestigateIncident(
	traceID, rawEntityID string,
	mlScore float64,
	graphFraudNeighbors int,
	behaviorAnomalies []string,
	threatMatches []string,
) *FraudInvestigationDossier {
	start := time.Now().UTC()

	// 1. Record short-term facts into memory
	a.memory.RememberShortTerm(traceID, rawEntityID, "EVALUATED_BY_MODEL", fmt.Sprintf("Score: %.2f", mlScore), mlScore)

	// 2. Query long term memory
	historicalMatches := 0
	if _, found := a.memory.RecallLongTerm(rawEntityID); found {
		historicalMatches++
	}

	// 3. Reason over facts
	hypo := a.reasoningEngine.Reason(mlScore, graphFraudNeighbors, behaviorAnomalies, threatMatches, historicalMatches)

	// 4. Generate structured explanation
	explanation := a.explainer.Explain(hypo)

	duration := float64(time.Since(start).Microseconds()) / 1000.0

	return &FraudInvestigationDossier{
		DossierID:             fmt.Sprintf("dos_%d_%s", time.Now().UnixNano(), traceID),
		TraceID:               traceID,
		TargetEntityID:        rawEntityID,
		ReasoningHypothesis:   hypo,
		Explanation:           explanation,
		AssignedAgentID:       a.AgentID,
		InvestigationDuration: duration,
		CreatedAt:             time.Now().UTC(),
	}
}
