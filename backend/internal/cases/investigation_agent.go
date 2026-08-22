package cases

import (
	"fmt"
	"math"
	"time"
)

// InvestigationReport summarizes autonomous agent forensic findings and recommended containment actions.
type InvestigationReport struct {
	ReportID               string          `json:"report_id"`
	TransactionID          string          `json:"transaction_id"`
	PrimaryUserID          string          `json:"primary_user_id"`
	RiskSummary            string          `json:"risk_summary"`
	FraudProbability       float64         `json:"fraud_probability"`
	ConnectedEntitiesCount int             `json:"connected_entities_count"`
	EvidenceItems          []EvidenceItem  `json:"evidence_items"`
	EvidenceTimeline       []TimelineEvent `json:"evidence_timeline"`
	RecommendedAction      string          `json:"recommended_action"` // "BLOCK_TRANSACTION", "FREEZE_ACCOUNT", "REQUIRE_MFA", "RELEASE"
	ConfidenceScore        float64         `json:"confidence_score"`
	GeneratedAt            time.Time       `json:"generated_at"`
}

// AutonomousInvestigationAgent analyzes multi-layer telemetry to generate complete forensic dossiers.
type AutonomousInvestigationAgent struct {
	evidenceEngine *EvidenceEngine
}

// NewAutonomousInvestigationAgent initializes the automated investigator.
func NewAutonomousInvestigationAgent(ee *EvidenceEngine) *AutonomousInvestigationAgent {
	if ee == nil {
		ee = NewEvidenceEngine()
	}
	return &AutonomousInvestigationAgent{evidenceEngine: ee}
}

// Investigate analyzes an incident and compiles the forensic investigation report.
func (a *AutonomousInvestigationAgent) Investigate(
	txnID, userID, device, ip string,
	amount, mlScore float64,
	graphFraudNeighbors int,
	behaviorAnomalies []string,
	threatMatches []string,
) (*InvestigationReport, error) {
	now := time.Now().UTC()

	evidence := a.evidenceEngine.BuildEvidencePackage(
		txnID, amount, graphFraudNeighbors, behaviorAnomalies, threatMatches, mlScore,
	)

	// Calculate holistic fraud probability
	baseProb := mlScore
	if graphFraudNeighbors > 0 {
		baseProb = math.Max(baseProb, 0.88)
	}
	if len(threatMatches) > 0 {
		baseProb = math.Max(baseProb, 0.95)
	}
	if len(behaviorAnomalies) > 0 {
		baseProb = math.Min(1.0, baseProb+0.15)
	}

	// Determine recommended action
	recommendedAction := "RELEASE"
	riskSummary := "Transaction exhibits normal patterns with no significant risk indicators"

	if baseProb >= 0.90 {
		recommendedAction = "FREEZE_ACCOUNT"
		riskSummary = fmt.Sprintf("Critical fraud threat: Confirmed malicious network or threat IOC intersection (Probability: %.2f)", baseProb)
	} else if baseProb >= 0.75 {
		recommendedAction = "BLOCK_TRANSACTION"
		riskSummary = fmt.Sprintf("High risk transaction: Significant behavioral or graph anomalies detected (Probability: %.2f)", baseProb)
	} else if baseProb >= 0.45 {
		recommendedAction = "REQUIRE_MFA"
		riskSummary = fmt.Sprintf("Moderate risk: Unrecognized device signature or baseline deviation (Probability: %.2f)", baseProb)
	}

	timeline := []TimelineEvent{
		{
			EventID:   fmt.Sprintf("tl_%d_1", now.UnixNano()),
			Actor:     "TRANSACTION_GATEWAY",
			Action:    "TRANSACTION_INITIATED",
			Details:   fmt.Sprintf("Transaction %s for $%.2f from device %s (IP: %s)", txnID, amount, device, ip),
			Timestamp: now.Add(-2 * time.Second),
		},
		{
			EventID:   fmt.Sprintf("tl_%d_2", now.UnixNano()),
			Actor:     "AUTO_INVESTIGATION_AGENT",
			Action:    "FORENSIC_TRIAGE_COMPLETED",
			Details:   riskSummary,
			Timestamp: now,
		},
	}

	return &InvestigationReport{
		ReportID:               fmt.Sprintf("inv_%d_%s", now.UnixNano(), txnID),
		TransactionID:          txnID,
		PrimaryUserID:          userID,
		RiskSummary:            riskSummary,
		FraudProbability:       baseProb,
		ConnectedEntitiesCount: graphFraudNeighbors + 1,
		EvidenceItems:          evidence,
		EvidenceTimeline:       timeline,
		RecommendedAction:      recommendedAction,
		ConfidenceScore:        0.96,
		GeneratedAt:            now,
	}, nil
}
