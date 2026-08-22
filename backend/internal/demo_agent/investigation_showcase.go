package demo_agent

import (
	"context"
	"fmt"
	"time"

	"github.com/shankywho/ropus/backend/internal/llm"
	"github.com/shankywho/ropus/backend/internal/memory"
)

// ShowcaseInvestigationReport encapsulates the investor-grade forensic output.
type ShowcaseInvestigationReport struct {
	ReportID              string    `json:"report_id"`
	TransactionID         string    `json:"transaction_id"`
	UserID                string    `json:"user_id"`
	Amount                float64   `json:"amount"`
	RiskScore             float64   `json:"risk_score"`
	Title                 string    `json:"title"`
	EvidenceItems         []string  `json:"evidence_items"`
	HistoricalPrecedents  []string  `json:"historical_precedents"`
	AutonomousReasoning   string    `json:"autonomous_reasoning"`
	RecommendedAction     string    `json:"recommended_action"`
	DecisionActionTaken   string    `json:"decision_action_taken"`
	InvestigatedAt        time.Time `json:"investigated_at"`
}

// InvestigationShowcaseEngine orchestrates customer and investor showcase investigations.
type InvestigationShowcaseEngine struct {
	agent *llm.LLMInvestigationAgent
}

// NewInvestigationShowcaseEngine initializes the showcase engine.
func NewInvestigationShowcaseEngine() *InvestigationShowcaseEngine {
	llmClient := llm.NewLLMClient("", "", "claude-3-7-sonnet-20250219")
	vectorStore := memory.NewVectorStore()
	agent := llm.NewLLMInvestigationAgent(llmClient, vectorStore)
	return &InvestigationShowcaseEngine{agent: agent}
}

// RunShowcaseInvestigation runs the full 5-stage automated investigation showcase.
func (e *InvestigationShowcaseEngine) RunShowcaseInvestigation(ctx context.Context, txID, userID string, amount float64) (*ShowcaseInvestigationReport, error) {
	now := time.Now().UTC()
	caseID := fmt.Sprintf("case_showcase_%s", txID)

	forensicReport, err := e.agent.InvestigateCase(ctx, caseID, userID, txID)
	if err != nil {
		return nil, err
	}

	evidence := []string{
		"Device fingerprint linked to 27 historical fraud accounts in Knowledge Graph",
		"Location anomaly: Login attempted from Lagos, NG 15 minutes after US checkout",
		"Behavioral deviation score: 94.0% (Velocity surge 18x normal spending)",
		"Threat Intelligence match on known bulletproof proxy blocklist",
	}

	return &ShowcaseInvestigationReport{
		ReportID:            fmt.Sprintf("rep_%d", now.UnixNano()),
		TransactionID:       txID,
		UserID:              userID,
		Amount:              amount,
		RiskScore:           94.5,
		Title:               "Account Takeover & Mule Routing Campaign Detected",
		EvidenceItems:       evidence,
		HistoricalPrecedents: forensicReport.SimilarPrecedents,
		AutonomousReasoning: forensicReport.FraudExplanation,
		RecommendedAction:   "Step-Up WebAuthn Challenge & Freeze Outbound Settlement Rails",
		DecisionActionTaken: "AUTONOMOUS_BLOCK",
		InvestigatedAt:      now,
	}, nil
}
