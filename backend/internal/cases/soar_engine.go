package cases

import (
	"fmt"
	"time"
)

// IncidentContext encapsulates parameters passed to a SOAR playbook.
type IncidentContext struct {
	IncidentID        string   `json:"incident_id"`
	PrimaryEntityID   string   `json:"primary_entity_id"`
	AssociatedEntities []string `json:"associated_entities"`
	RiskScore         float64  `json:"risk_score"`
	ThreatType        string   `json:"threat_type"` // "CREDENTIAL_THEFT", "CARD_TESTING", "FRAUD_RING"
	TransactionIDs    []string `json:"transaction_ids"`
	TotalExposure     float64  `json:"total_exposure"`
}

// PlaybookResult details the containment actions executed by SOAR.
type PlaybookResult struct {
	PlaybookName     string                     `json:"playbook_name"`
	IncidentID       string                     `json:"incident_id"`
	ExecutedActions  []*ResponseExecutionRecord `json:"executed_actions"`
	CaseGeneratedID  string                     `json:"case_generated_id,omitempty"`
	Status           string                     `json:"status"` // "COMPLETED", "FAILED"
	CompletedAt      time.Time                  `json:"completed_at"`
}

// SOAREngine coordinates automated incident response playbooks.
type SOAREngine struct {
	responseEngine *ResponseEngine
	caseManager    *CaseManager
}

// NewSOAREngine initializes the SOAR playbook orchestrator.
func NewSOAREngine(re *ResponseEngine, cm *CaseManager) *SOAREngine {
	return &SOAREngine{
		responseEngine: re,
		caseManager:    cm,
	}
}

// ExecutePlaybook triggers the designated security orchestration sequence.
func (s *SOAREngine) ExecutePlaybook(playbookName string, ctx IncidentContext) (*PlaybookResult, error) {
	now := time.Now().UTC()
	var executed []*ResponseExecutionRecord
	var caseID string

	switch playbookName {
	case "PLAYBOOK_CREDENTIAL_THEFT":
		// 1. Require step-up MFA
		act1, err := s.responseEngine.ExecuteAction(ActionRequireMFA, ctx.PrimaryEntityID, "USER", "Credential theft indicator", "SOAR_PLAYBOOK", 0.90, ctx.RiskScore)
		if err == nil {
			executed = append(executed, act1)
		}
		// 2. Challenge active sessions
		act2, err := s.responseEngine.ExecuteAction(ActionChallengeUser, ctx.PrimaryEntityID, "USER", "Session invalidation required", "SOAR_PLAYBOOK", 0.90, ctx.RiskScore)
		if err == nil {
			executed = append(executed, act2)
		}

	case "PLAYBOOK_CARD_TESTING":
		// 1. Block compromised device
		act1, err := s.responseEngine.ExecuteAction(ActionBlockTransaction, ctx.PrimaryEntityID, "DEVICE", "High frequency card testing detected", "SOAR_PLAYBOOK", 0.95, ctx.RiskScore)
		if err == nil {
			executed = append(executed, act1)
		}

	case "PLAYBOOK_FRAUD_RING":
		// 1. Freeze primary account
		act1, err := s.responseEngine.ExecuteAction(ActionFreezeAccount, ctx.PrimaryEntityID, "ACCOUNT", "Syndicate fraud ring node hub", "SOAR_PLAYBOOK", 0.95, ctx.RiskScore)
		if err == nil {
			executed = append(executed, act1)
		}
		// 2. Freeze connected entities
		for _, eID := range ctx.AssociatedEntities {
			act, err := s.responseEngine.ExecuteAction(ActionFreezeAccount, eID, "ACCOUNT", "Syndicate member node", "SOAR_PLAYBOOK", 0.90, ctx.RiskScore)
			if err == nil {
				executed = append(executed, act)
			}
		}
		// 3. Open high-priority investigation case
		if s.caseManager != nil {
			c, err := s.caseManager.CreateCase("default_tenant", ctx.PrimaryEntityID, ctx.TransactionIDs, ctx.TotalExposure, ctx.RiskScore, len(ctx.AssociatedEntities)+1, 15, true)
			if err == nil {
				caseID = c.CaseID
			}
		}
	default:
		return nil, fmt.Errorf("unknown playbook '%s'", playbookName)
	}

	return &PlaybookResult{
		PlaybookName:    playbookName,
		IncidentID:      ctx.IncidentID,
		ExecutedActions: executed,
		CaseGeneratedID: caseID,
		Status:          "COMPLETED",
		CompletedAt:     now,
	}, nil
}
