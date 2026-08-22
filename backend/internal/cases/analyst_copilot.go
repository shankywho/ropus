package cases

import (
	"fmt"
)

// AnalystCopilot provides high-level investigation and decision APIs for human analysts.
type AnalystCopilot struct {
	caseManager *CaseManager
}

// NewAnalystCopilot initializes the analyst assistant copilot.
func NewAnalystCopilot(cm *CaseManager) *AnalystCopilot {
	return &AnalystCopilot{caseManager: cm}
}

// GetCaseDetails fetches full case context.
func (c *AnalystCopilot) GetCaseDetails(caseID string) (*FraudCase, error) {
	return c.caseManager.GetCase(caseID)
}

// GetCaseTimeline retrieves chronological events for a case.
func (c *AnalystCopilot) GetCaseTimeline(caseID string) ([]TimelineEvent, error) {
	fc, err := c.caseManager.GetCase(caseID)
	if err != nil {
		return nil, err
	}
	return fc.Timeline, nil
}

// GetCaseEvidence retrieves evidence artifacts.
func (c *AnalystCopilot) GetCaseEvidence(caseID string) ([]EvidenceItem, error) {
	fc, err := c.caseManager.GetCase(caseID)
	if err != nil {
		return nil, err
	}
	return fc.Evidence, nil
}

// AssignCase allocates a case to an analyst.
func (c *AnalystCopilot) AssignCase(caseID, analystID string) error {
	return c.caseManager.AssignCase(caseID, analystID)
}

// ResolveCase processes the analyst's final determination.
func (c *AnalystCopilot) ResolveCase(caseID string, isFraud bool, notes, analystID string) error {
	status := StatusConfirmedFraud
	if !isFraud {
		status = StatusFalsePositive
	}
	if notes == "" {
		notes = fmt.Sprintf("Resolved by %s", analystID)
	}
	return c.caseManager.ResolveCase(caseID, status, notes, analystID)
}
