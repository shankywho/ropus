package product_api

import (
	"time"
)

// CreateCaseRequest creates a new investigation case.
type CreateCaseRequest struct {
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Severity      string   `json:"severity"` // "CRITICAL", "HIGH", "MEDIUM", "LOW"
	EntityID      string   `json:"entity_id"`
	TransactionID string   `json:"transaction_id,omitempty"`
	Tags          []string `json:"tags,omitempty"`
}

// CaseTimelineEntry represents an event in the case's lifecycle.
type CaseTimelineEntry struct {
	Timestamp   time.Time `json:"timestamp"`
	Actor       string    `json:"actor"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
}

// ProductCaseResponse represents the complete case details.
type ProductCaseResponse struct {
	CaseID               string              `json:"case_id"`
	Title                string              `json:"title"`
	Description          string              `json:"description"`
	Status               string              `json:"status"` // "OPEN", "INVESTIGATING", "RESOLVED", "CLOSED"
	Severity             string              `json:"severity"`
	EntityID             string              `json:"entity_id"`
	AssignedAnalyst      string              `json:"assigned_analyst,omitempty"`
	AIInvestigatorReport string              `json:"ai_investigator_report"`
	EvidenceCount        int                 `json:"evidence_count"`
	RecommendedAction    string              `json:"recommended_action"`
	Timeline             []CaseTimelineEntry `json:"timeline"`
	CreatedAt            time.Time           `json:"created_at"`
	UpdatedAt            time.Time           `json:"updated_at"`
}
