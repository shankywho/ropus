package cases

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrCaseNotFound    = errors.New("case not found")
	ErrInvalidStatus   = errors.New("invalid case status")
	ErrCaseAlreadyDone = errors.New("case already resolved or closed")
)

type CaseStatus string

const (
	StatusOpen            CaseStatus = "OPEN"
	StatusUnderReview     CaseStatus = "UNDER_REVIEW"
	StatusResolvedAllow   CaseStatus = "RESOLVED_ALLOW"
	StatusResolvedDecline CaseStatus = "RESOLVED_DECLINE"
	StatusClosed          CaseStatus = "CLOSED"
)

type Case struct {
	ID               string     `json:"case_id"`
	TenantID         string     `json:"tenant_id"`
	DecisionID       string     `json:"decision_id"`
	TransactionID    string     `json:"transaction_id"`
	Status           CaseStatus `json:"status"`
	Priority         string     `json:"priority"`
	AssignedTo       *string    `json:"assigned_to,omitempty"`
	ResolutionReason *string    `json:"resolution_reason,omitempty"`
	ResolvedAt       *time.Time `json:"resolved_at,omitempty"`
	SLAExpiresAt     time.Time  `json:"sla_expires_at"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type CaseWithDecision struct {
	Case
	Amount             int64           `json:"amount"`
	Currency           string          `json:"currency"`
	RiskScore          int             `json:"risk_score"`
	RecommendedAction  string          `json:"recommended_action"`
	ReasonCodes        json.RawMessage `json:"reason_codes"`
	FeatureSnapshot    json.RawMessage `json:"feature_snapshot"`
	RawPayload         json.RawMessage `json:"raw_payload"`
}

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// CreateCaseFromDecision automatically creates an OPEN case with a 24-hour SLA.
func (s *Service) CreateCaseFromDecision(ctx context.Context, tenantID, decisionID, transactionID string, priority string) (*Case, error) {
	if s.db == nil {
		return nil, nil
	}

	if priority == "" {
		priority = "MEDIUM"
	}

	caseID := uuid.New().String()
	now := time.Now().UTC()
	slaExpiresAt := now.Add(24 * time.Hour) // 24-hour analyst SLA

	query := `
		INSERT INTO cases (case_id, tenant_id, decision_id, transaction_id, status, priority, sla_expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING case_id, tenant_id, decision_id, transaction_id, status, priority, assigned_to, resolution_reason, resolved_at, sla_expires_at, created_at, updated_at
	`

	var c Case
	err := s.db.QueryRow(ctx, query,
		caseID,
		tenantID,
		decisionID,
		transactionID,
		string(StatusOpen),
		priority,
		slaExpiresAt,
		now,
		now,
	).Scan(
		&c.ID,
		&c.TenantID,
		&c.DecisionID,
		&c.TransactionID,
		&c.Status,
		&c.Priority,
		&c.AssignedTo,
		&c.ResolutionReason,
		&c.ResolvedAt,
		&c.SLAExpiresAt,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create case: %w", err)
	}

	return &c, nil
}

// ListCases returns cases for a tenant, optionally filtered by status.
func (s *Service) ListCases(ctx context.Context, tenantID string, status *CaseStatus) ([]Case, error) {
	if s.db == nil {
		return []Case{}, nil
	}

	var query string
	var args []interface{}

	if status != nil && *status != "" {
		query = `
			SELECT case_id, tenant_id, decision_id, transaction_id, status, priority, assigned_to, resolution_reason, resolved_at, sla_expires_at, created_at, updated_at
			FROM cases
			WHERE tenant_id = $1 AND status = $2
			ORDER BY created_at DESC
		`
		args = []interface{}{tenantID, string(*status)}
	} else {
		query = `
			SELECT case_id, tenant_id, decision_id, transaction_id, status, priority, assigned_to, resolution_reason, resolved_at, sla_expires_at, created_at, updated_at
			FROM cases
			WHERE tenant_id = $1
			ORDER BY created_at DESC
		`
		args = []interface{}{tenantID}
	}

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list cases: %w", err)
	}
	defer rows.Close()

	results := make([]Case, 0)
	for rows.Next() {
		var c Case
		if err := rows.Scan(
			&c.ID,
			&c.TenantID,
			&c.DecisionID,
			&c.TransactionID,
			&c.Status,
			&c.Priority,
			&c.AssignedTo,
			&c.ResolutionReason,
			&c.ResolvedAt,
			&c.SLAExpiresAt,
			&c.CreatedAt,
			&c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan case row: %w", err)
		}
		results = append(results, c)
	}

	return results, nil
}

// GetCase retrieves a case joined with its associated risk decision and evidence snapshot.
func (s *Service) GetCase(ctx context.Context, tenantID, caseID string) (*CaseWithDecision, error) {
	if s.db == nil {
		return nil, ErrCaseNotFound
	}

	query := `
		SELECT 
			c.case_id, c.tenant_id, c.decision_id, c.transaction_id, c.status, c.priority,
			c.assigned_to, c.resolution_reason, c.resolved_at, c.sla_expires_at, c.created_at, c.updated_at,
			d.amount, d.currency, d.risk_score, d.recommended_action, d.reason_codes, d.feature_snapshot, d.raw_payload
		FROM cases c
		JOIN risk_decisions d ON c.decision_id = d.decision_id
		WHERE c.tenant_id = $1 AND c.case_id = $2
	`

	var cd CaseWithDecision
	err := s.db.QueryRow(ctx, query, tenantID, caseID).Scan(
		&cd.ID,
		&cd.TenantID,
		&cd.DecisionID,
		&cd.TransactionID,
		&cd.Status,
		&cd.Priority,
		&cd.AssignedTo,
		&cd.ResolutionReason,
		&cd.ResolvedAt,
		&cd.SLAExpiresAt,
		&cd.CreatedAt,
		&cd.UpdatedAt,
		&cd.Amount,
		&cd.Currency,
		&cd.RiskScore,
		&cd.RecommendedAction,
		&cd.ReasonCodes,
		&cd.FeatureSnapshot,
		&cd.RawPayload,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCaseNotFound
		}
		return nil, fmt.Errorf("failed to get case details: %w", err)
	}

	return &cd, nil
}

// ClaimCase assigns an analyst to an OPEN case and moves it to UNDER_REVIEW.
func (s *Service) ClaimCase(ctx context.Context, tenantID, caseID, analystID string) (*Case, error) {
	if s.db == nil {
		return nil, ErrCaseNotFound
	}

	now := time.Now().UTC()
	query := `
		UPDATE cases
		SET assigned_to = $1, status = $2, updated_at = $3
		WHERE tenant_id = $4 AND case_id = $5
		RETURNING case_id, tenant_id, decision_id, transaction_id, status, priority, assigned_to, resolution_reason, resolved_at, sla_expires_at, created_at, updated_at
	`

	var c Case
	err := s.db.QueryRow(ctx, query,
		analystID,
		string(StatusUnderReview),
		now,
		tenantID,
		caseID,
	).Scan(
		&c.ID,
		&c.TenantID,
		&c.DecisionID,
		&c.TransactionID,
		&c.Status,
		&c.Priority,
		&c.AssignedTo,
		&c.ResolutionReason,
		&c.ResolvedAt,
		&c.SLAExpiresAt,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCaseNotFound
		}
		return nil, fmt.Errorf("failed to claim case: %w", err)
	}

	return &c, nil
}

// ResolveCase sets the human decision override (ALLOW or DECLINE) and closes the review.
func (s *Service) ResolveCase(ctx context.Context, tenantID, caseID, resolutionAction, reason, analystID string) (*Case, error) {
	if s.db == nil {
		return nil, ErrCaseNotFound
	}

	var targetStatus CaseStatus
	if resolutionAction == "ALLOW" || resolutionAction == "RESOLVED_ALLOW" {
		targetStatus = StatusResolvedAllow
	} else if resolutionAction == "DECLINE" || resolutionAction == "RESOLVED_DECLINE" {
		targetStatus = StatusResolvedDecline
	} else {
		return nil, fmt.Errorf("%w: must be ALLOW or DECLINE", ErrInvalidStatus)
	}

	now := time.Now().UTC()
	query := `
		UPDATE cases
		SET status = $1, resolution_reason = $2, assigned_to = COALESCE(assigned_to, $3), resolved_at = $4, updated_at = $5
		WHERE tenant_id = $6 AND case_id = $7
		RETURNING case_id, tenant_id, decision_id, transaction_id, status, priority, assigned_to, resolution_reason, resolved_at, sla_expires_at, created_at, updated_at
	`

	var c Case
	err := s.db.QueryRow(ctx, query,
		string(targetStatus),
		reason,
		analystID,
		now,
		now,
		tenantID,
		caseID,
	).Scan(
		&c.ID,
		&c.TenantID,
		&c.DecisionID,
		&c.TransactionID,
		&c.Status,
		&c.Priority,
		&c.AssignedTo,
		&c.ResolutionReason,
		&c.ResolvedAt,
		&c.SLAExpiresAt,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCaseNotFound
		}
		return nil, fmt.Errorf("failed to resolve case: %w", err)
	}

	// Insert into audit log
	_, _ = s.db.Exec(ctx, `
		INSERT INTO audit_log (tenant_id, actor_id, action, entity_type, entity_id, changes, created_at)
		VALUES ($1, $2, 'CASE_RESOLVED', 'CASE', $3, $4, $5)
	`, tenantID, analystID, caseID, fmt.Sprintf(`{"action":"%s","reason":"%s"}`, targetStatus, reason), now)

	return &c, nil
}
