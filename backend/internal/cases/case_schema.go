package cases

import (
	"time"
)

// CasePriority categorizes the urgency and financial exposure of a fraud case.
type CasePriority string

const (
	PriorityLow      CasePriority = "LOW"
	PriorityMedium   CasePriority = "MEDIUM"
	PriorityHigh     CasePriority = "HIGH"
	PriorityCritical CasePriority = "CRITICAL"
)

// EvidenceItem encapsulates an individual forensic proof artifact.
type EvidenceItem struct {
	EvidenceID string    `json:"evidence_id"`
	Type       string    `json:"type"` // "TRANSACTION", "GRAPH", "BEHAVIOR", "THREAT_INTEL", "DEVICE"
	Summary    string    `json:"summary"`
	Details    string    `json:"details"`
	RiskWeight float64   `json:"risk_weight"`
	CapturedAt time.Time `json:"captured_at"`
}

// TimelineEvent records a chronological action or system event on the case.
type TimelineEvent struct {
	EventID   string    `json:"event_id"`
	Actor     string    `json:"actor"` // "SYSTEM_AUTO_INVESTIGATOR", "ANALYST_ALICE", "SOAR_ENGINE"
	Action    string    `json:"action"`
	Details   string    `json:"details"`
	Timestamp time.Time `json:"timestamp"`
}

// FraudCase represents an end-to-end investigation case.
type FraudCase struct {
	CaseID          string                 `json:"case_id"`
	TenantID        string                 `json:"tenant_id"`
	TransactionIDs  []string               `json:"transaction_ids"`
	PrimaryUserID   string                 `json:"primary_user_id"`
	TotalExposure   float64                `json:"total_exposure"`
	RiskScore       float64                `json:"risk_score"`
	Priority        CasePriority           `json:"priority"`
	Status          CaseStatus             `json:"status"`
	AssignedAnalyst string                 `json:"assigned_analyst,omitempty"`
	Evidence        []EvidenceItem         `json:"evidence"`
	Timeline        []TimelineEvent        `json:"timeline"`
	ResolutionNotes string                 `json:"resolution_notes,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	ResolvedAt      *time.Time             `json:"resolved_at,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}
