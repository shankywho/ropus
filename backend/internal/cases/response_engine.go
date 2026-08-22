package cases

import (
	"fmt"
	"sync"
	"time"
)

// ResponseActionType defines the containment action triggered on high risk entities.
type ResponseActionType string

const (
	ActionBlockTransaction ResponseActionType = "BLOCK_TRANSACTION"
	ActionFreezeAccount    ResponseActionType = "FREEZE_ACCOUNT"
	ActionChallengeUser    ResponseActionType = "CHALLENGE_USER"
	ActionRequireMFA       ResponseActionType = "REQUIRE_MFA"
	ActionLimitAccount     ResponseActionType = "LIMIT_ACCOUNT"
	ActionEscalateReview   ResponseActionType = "ESCALATE_REVIEW"
)

// ResponseExecutionRecord documents the execution and rollback state of containment actions.
type ResponseExecutionRecord struct {
	ActionID        string             `json:"action_id"`
	ActionType      ResponseActionType `json:"action_type"`
	TargetEntityID  string             `json:"target_entity_id"`
	TargetType      string             `json:"target_type"` // "USER", "ACCOUNT", "DEVICE", "TRANSACTION"
	TriggerReason   string             `json:"trigger_reason"`
	Actor           string             `json:"actor"` // "SOAR_AUTO_PILOT", "ANALYST_BOB"
	ExecutedAt      time.Time          `json:"executed_at"`
	RollbackCapable bool               `json:"rollback_capable"`
	IsRolledBack    bool               `json:"is_rolled_back"`
	RolledBackAt    *time.Time         `json:"rolled_back_at,omitempty"`
}

// ResponseEngine coordinates containment actions and audit records.
type ResponseEngine struct {
	mu      sync.RWMutex
	records map[string]*ResponseExecutionRecord
	guard   *ResponseGuard
}

// NewResponseEngine initializes the response dispatcher.
func NewResponseEngine(guard *ResponseGuard) *ResponseEngine {
	if guard == nil {
		guard = NewResponseGuard()
	}
	return &ResponseEngine{
		records: make(map[string]*ResponseExecutionRecord),
		guard:   guard,
	}
}

// ExecuteAction applies containment controls with safety guardrail enforcement.
func (e *ResponseEngine) ExecuteAction(
	actionType ResponseActionType,
	targetEntityID, targetType, triggerReason, actor string,
	confidence, riskScore float64,
) (*ResponseExecutionRecord, error) {
	if err := e.guard.ValidateSafety(string(actionType), targetEntityID, confidence, riskScore); err != nil {
		return nil, fmt.Errorf("safety guardrail violation: %w", err)
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now().UTC()
	actionID := fmt.Sprintf("act_%d_%s", now.UnixNano(), targetEntityID)

	rec := &ResponseExecutionRecord{
		ActionID:        actionID,
		ActionType:      actionType,
		TargetEntityID:  targetEntityID,
		TargetType:      targetType,
		TriggerReason:   triggerReason,
		Actor:           actor,
		ExecutedAt:      now,
		RollbackCapable: true,
		IsRolledBack:    false,
	}

	e.records[actionID] = rec
	return rec, nil
}

// RollbackAction reverses a previously executed containment control.
func (e *ResponseEngine) RollbackAction(actionID, reason string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	rec, exists := e.records[actionID]
	if !exists {
		return fmt.Errorf("action record '%s' not found", actionID)
	}
	if !rec.RollbackCapable {
		return fmt.Errorf("action '%s' is not rollback capable", actionID)
	}
	if rec.IsRolledBack {
		return fmt.Errorf("action '%s' has already been rolled back", actionID)
	}

	now := time.Now().UTC()
	rec.IsRolledBack = true
	rec.RolledBackAt = &now
	return nil
}

// GetAction retrieves an action record.
func (e *ResponseEngine) GetAction(actionID string) (*ResponseExecutionRecord, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	rec, exists := e.records[actionID]
	if !exists {
		return nil, fmt.Errorf("action record '%s' not found", actionID)
	}
	return rec, nil
}
