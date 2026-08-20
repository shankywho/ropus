package rules

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
	ErrNotFound              = errors.New("rule not found")
	ErrMakerCheckerViolation = errors.New("maker-checker violation: rule creator cannot approve their own rule")
	ErrInvalidStatusChange   = errors.New("invalid status transition")
	ErrInvalidDSL            = errors.New("invalid rule AST DSL")
)

type RuleStatus string

const (
	StatusDraft           RuleStatus = "DRAFT"
	StatusPendingApproval RuleStatus = "PENDING_APPROVAL"
	StatusShadow          RuleStatus = "SHADOW"
	StatusActive          RuleStatus = "ACTIVE"
	StatusArchived        RuleStatus = "ARCHIVED"
)

type Rule struct {
	ID          string          `json:"rule_id"`
	TenantID    string          `json:"tenant_id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	DSLAST      json.RawMessage `json:"dsl_ast"`
	Status      RuleStatus      `json:"status"`
	Version     int             `json:"version"`
	CreatedBy   string          `json:"created_by"`
	ApprovedBy  *string         `json:"approved_by,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type CreateRuleInput struct {
	TenantID    string          `json:"tenant_id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	DSLAST      json.RawMessage `json:"dsl_ast"`
	CreatedBy   string          `json:"created_by"`
}

type UpdateRuleInput struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	DSLAST      json.RawMessage `json:"dsl_ast"`
	UpdatedBy   string          `json:"updated_by"`
}

type Service struct {
	db *pgxpool.Pool
}

func NewService(db *pgxpool.Pool) *Service {
	return &Service{db: db}
}

// ensureTenantExists ensures the tenant record exists before foreign key insert.
func (s *Service) ensureTenantExists(ctx context.Context, tenantID string) error {
	if s.db == nil {
		return nil
	}
	_, err := s.db.Exec(ctx, `
		INSERT INTO tenants (tenant_id, name, api_key_hash, status)
		VALUES ($1, 'Default Tenant', $2, 'ACTIVE')
		ON CONFLICT (tenant_id) DO NOTHING
	`, tenantID, fmt.Sprintf("key_%s", tenantID))
	return err
}

// CreateRule creates a new rule in DRAFT status.
func (s *Service) CreateRule(ctx context.Context, input CreateRuleInput) (*Rule, error) {
	// Validate AST
	if _, err := ParseRuleDefinition(input.DSLAST); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidDSL, err)
	}

	if input.TenantID == "" {
		input.TenantID = "00000000-0000-0000-0000-000000000001"
	}
	if input.CreatedBy == "" {
		input.CreatedBy = "system"
	}

	if err := s.ensureTenantExists(ctx, input.TenantID); err != nil {
		return nil, fmt.Errorf("failed to ensure tenant: %w", err)
	}

	ruleID := uuid.New().String()
	now := time.Now().UTC()

	query := `
		INSERT INTO rules (rule_id, tenant_id, name, description, dsl_ast, status, version, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING rule_id, tenant_id, name, description, dsl_ast, status, version, created_by, approved_by, created_at, updated_at
	`

	var rule Rule
	err := s.db.QueryRow(ctx, query,
		ruleID,
		input.TenantID,
		input.Name,
		input.Description,
		input.DSLAST,
		string(StatusDraft),
		1,
		input.CreatedBy,
		now,
		now,
	).Scan(
		&rule.ID,
		&rule.TenantID,
		&rule.Name,
		&rule.Description,
		&rule.DSLAST,
		&rule.Status,
		&rule.Version,
		&rule.CreatedBy,
		&rule.ApprovedBy,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create rule: %w", err)
	}

	return &rule, nil
}

// GetRule retrieves a single rule by ID.
func (s *Service) GetRule(ctx context.Context, tenantID, ruleID string) (*Rule, error) {
	query := `
		SELECT rule_id, tenant_id, name, description, dsl_ast, status, version, created_by, approved_by, created_at, updated_at
		FROM rules
		WHERE tenant_id = $1 AND rule_id = $2
	`

	var rule Rule
	err := s.db.QueryRow(ctx, query, tenantID, ruleID).Scan(
		&rule.ID,
		&rule.TenantID,
		&rule.Name,
		&rule.Description,
		&rule.DSLAST,
		&rule.Status,
		&rule.Version,
		&rule.CreatedBy,
		&rule.ApprovedBy,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get rule: %w", err)
	}

	return &rule, nil
}

// ListRules lists rules for a tenant, optionally filtered by status.
func (s *Service) ListRules(ctx context.Context, tenantID string, status *RuleStatus) ([]Rule, error) {
	var query string
	var args []interface{}

	if status != nil && *status != "" {
		query = `
			SELECT rule_id, tenant_id, name, description, dsl_ast, status, version, created_by, approved_by, created_at, updated_at
			FROM rules
			WHERE tenant_id = $1 AND status = $2
			ORDER BY updated_at DESC
		`
		args = []interface{}{tenantID, string(*status)}
	} else {
		query = `
			SELECT rule_id, tenant_id, name, description, dsl_ast, status, version, created_by, approved_by, created_at, updated_at
			FROM rules
			WHERE tenant_id = $1
			ORDER BY updated_at DESC
		`
		args = []interface{}{tenantID}
	}

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query rules: %w", err)
	}
	defer rows.Close()

	rules := make([]Rule, 0)
	for rows.Next() {
		var rule Rule
		if err := rows.Scan(
			&rule.ID,
			&rule.TenantID,
			&rule.Name,
			&rule.Description,
			&rule.DSLAST,
			&rule.Status,
			&rule.Version,
			&rule.CreatedBy,
			&rule.ApprovedBy,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan rule row: %w", err)
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

// UpdateRule updates a rule's definition. If already approved, resets to DRAFT and increments version.
func (s *Service) UpdateRule(ctx context.Context, tenantID, ruleID string, input UpdateRuleInput) (*Rule, error) {
	current, err := s.GetRule(ctx, tenantID, ruleID)
	if err != nil {
		return nil, err
	}

	if len(input.DSLAST) > 0 {
		if _, err := ParseRuleDefinition(input.DSLAST); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrInvalidDSL, err)
		}
	} else {
		input.DSLAST = current.DSLAST
	}

	newStatus := current.Status
	// If rule was ACTIVE or PENDING_APPROVAL, editing it reverts status to DRAFT for review
	if current.Status != StatusDraft {
		newStatus = StatusDraft
	}

	newVersion := current.Version + 1
	now := time.Now().UTC()

	query := `
		UPDATE rules
		SET name = $1, description = $2, dsl_ast = $3, status = $4, version = $5, updated_at = $6
		WHERE tenant_id = $7 AND rule_id = $8
		RETURNING rule_id, tenant_id, name, description, dsl_ast, status, version, created_by, approved_by, created_at, updated_at
	`

	var rule Rule
	err = s.db.QueryRow(ctx, query,
		input.Name,
		input.Description,
		input.DSLAST,
		string(newStatus),
		newVersion,
		now,
		tenantID,
		ruleID,
	).Scan(
		&rule.ID,
		&rule.TenantID,
		&rule.Name,
		&rule.Description,
		&rule.DSLAST,
		&rule.Status,
		&rule.Version,
		&rule.CreatedBy,
		&rule.ApprovedBy,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update rule: %w", err)
	}

	return &rule, nil
}

// TransitionStatus executes the Maker-Checker state machine transitions.
func (s *Service) TransitionStatus(ctx context.Context, tenantID, ruleID string, targetStatus RuleStatus, actorID string) (*Rule, error) {
	current, err := s.GetRule(ctx, tenantID, ruleID)
	if err != nil {
		return nil, err
	}

	var approvedBy *string
	now := time.Now().UTC()

	switch targetStatus {
	case StatusPendingApproval:
		// Any creator/editor can submit a DRAFT rule for approval
		if current.Status != StatusDraft {
			return nil, fmt.Errorf("%w: can only submit DRAFT rules for approval (current: %s)", ErrInvalidStatusChange, current.Status)
		}

	case StatusActive, StatusShadow:
		// MAKER-CHECKER ENFORCEMENT:
		// 1. Must be in PENDING_APPROVAL or SHADOW state
		if current.Status != StatusPendingApproval && current.Status != StatusShadow {
			return nil, fmt.Errorf("%w: can only approve rules in PENDING_APPROVAL state (current: %s)", ErrInvalidStatusChange, current.Status)
		}
		// 2. The approver CANNOT be the rule creator
		if actorID != "" && current.CreatedBy != "" && actorID == current.CreatedBy {
			return nil, ErrMakerCheckerViolation
		}
		approvedBy = &actorID

	case StatusDraft:
		// Rejection / revert back to draft
		approvedBy = nil

	case StatusArchived:
		// Archive rule from any state
		approvedBy = current.ApprovedBy

	default:
		return nil, fmt.Errorf("%w: unknown target status %s", ErrInvalidStatusChange, targetStatus)
	}

	query := `
		UPDATE rules
		SET status = $1, approved_by = $2, updated_at = $3
		WHERE tenant_id = $4 AND rule_id = $5
		RETURNING rule_id, tenant_id, name, description, dsl_ast, status, version, created_by, approved_by, created_at, updated_at
	`

	var rule Rule
	err = s.db.QueryRow(ctx, query,
		string(targetStatus),
		approvedBy,
		now,
		tenantID,
		ruleID,
	).Scan(
		&rule.ID,
		&rule.TenantID,
		&rule.Name,
		&rule.Description,
		&rule.DSLAST,
		&rule.Status,
		&rule.Version,
		&rule.CreatedBy,
		&rule.ApprovedBy,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to transition rule status: %w", err)
	}

	return &rule, nil
}
