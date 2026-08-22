package governance

import (
	"fmt"
	"sync"
	"time"
)

// ReviewState tracks the manual risk review case lifecycle.
type ReviewState string

const (
	ReviewPending   ReviewState = "PENDING"
	ReviewAssigned  ReviewState = "ASSIGNED"
	ReviewApproved  ReviewState = "APPROVED"
	ReviewRejected  ReviewState = "REJECTED"
	ReviewEscalated ReviewState = "ESCALATED"
	ReviewClosed    ReviewState = "CLOSED"
)

// ReviewCase represents a high-risk transaction flagged for human investigation.
type ReviewCase struct {
	ReviewID        string            `json:"review_id"`
	TransactionID   string            `json:"transaction_id"`
	TenantID        string            `json:"tenant_id"`
	RiskScore       float64           `json:"risk_score"`
	FlaggedReasons  []string          `json:"flagged_reasons"`
	Status          ReviewState       `json:"status"`
	AssignedAnalyst string            `json:"assigned_analyst,omitempty"`
	AnalystDecision string            `json:"analyst_decision,omitempty"` // "CONFIRM_FRAUD", "FALSE_POSITIVE"
	ResolutionNotes string            `json:"resolution_notes,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	AssignedAt      *time.Time        `json:"assigned_at,omitempty"`
	ResolvedAt      *time.Time        `json:"resolved_at,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

// HumanReviewSystem manages analyst queues and audit records of manual interventions.
type HumanReviewSystem struct {
	mu    sync.RWMutex
	cases map[string]*ReviewCase
}

// NewHumanReviewSystem initializes the manual review case manager.
func NewHumanReviewSystem() *HumanReviewSystem {
	return &HumanReviewSystem{
		cases: make(map[string]*ReviewCase),
	}
}

// CreateReview flags a transaction for manual review.
func (s *HumanReviewSystem) CreateReview(txnID, tenantID string, riskScore float64, reasons []string) *ReviewCase {
	s.mu.Lock()
	defer s.mu.Unlock()

	caseID := fmt.Sprintf("rev_%d_%s", time.Now().UnixNano(), txnID)
	review := &ReviewCase{
		ReviewID:       caseID,
		TransactionID:  txnID,
		TenantID:       tenantID,
		RiskScore:      riskScore,
		FlaggedReasons: reasons,
		Status:         ReviewPending,
		CreatedAt:      time.Now().UTC(),
	}
	s.cases[caseID] = review
	return review
}

// AssignReview allocates a case to an analyst.
func (s *HumanReviewSystem) AssignReview(reviewID, analystID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, exists := s.cases[reviewID]
	if !exists {
		return fmt.Errorf("review case '%s' not found", reviewID)
	}

	now := time.Now().UTC()
	c.AssignedAnalyst = analystID
	c.AssignedAt = &now
	c.Status = ReviewAssigned
	return nil
}

// SubmitDecision records the analyst's conclusion.
func (s *HumanReviewSystem) SubmitDecision(reviewID, decision, notes string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, exists := s.cases[reviewID]
	if !exists {
		return fmt.Errorf("review case '%s' not found", reviewID)
	}

	now := time.Now().UTC()
	c.AnalystDecision = decision
	c.ResolutionNotes = notes
	c.ResolvedAt = &now

	if decision == "CONFIRM_FRAUD" {
		c.Status = ReviewRejected
	} else if decision == "FALSE_POSITIVE" {
		c.Status = ReviewApproved
	} else if decision == "ESCALATE" {
		c.Status = ReviewEscalated
	} else {
		c.Status = ReviewClosed
	}

	return nil
}

// ListReviews returns review cases filtered by status.
func (s *HumanReviewSystem) ListReviews(status ReviewState) []*ReviewCase {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var res []*ReviewCase
	for _, c := range s.cases {
		if status == "" || c.Status == status {
			res = append(res, c)
		}
	}
	return res
}
