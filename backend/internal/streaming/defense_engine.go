package streaming

import (
	"fmt"
	"sync"
	"time"
)

// DefenseActionType defines the automated containment action types.
type DefenseActionType string

const (
	DefenseDeviceBlock          DefenseActionType = "DEVICE_BLOCK"
	DefenseAccountLock          DefenseActionType = "ACCOUNT_LOCK"
	DefenseMerchantRestriction  DefenseActionType = "MERCHANT_RESTRICTION"
	DefenseNetworkBlock         DefenseActionType = "NETWORK_BLOCK"
	DefenseStepUpAuth           DefenseActionType = "STEP_UP_AUTH"
)

// AutonomousDefenseRecord captures the audit trail and state of executed defense controls.
type AutonomousDefenseRecord struct {
	DefenseID       string            `json:"defense_id"`
	ActionType      DefenseActionType `json:"action_type"`
	TargetEntityID  string            `json:"target_entity_id"`
	Reason          string            `json:"reason"`
	Confidence      float64           `json:"confidence"`
	ExecutedAt      time.Time         `json:"executed_at"`
	IsRolledBack    bool              `json:"is_rolled_back"`
	RolledBackAt    *time.Time        `json:"rolled_back_at,omitempty"`
}

// AutonomousDefenseEngine orchestrates real-time containment with impact validation.
type AutonomousDefenseEngine struct {
	mu             sync.RWMutex
	records        map[string]*AutonomousDefenseRecord
	impactAnalyzer *ImpactAnalyzer
}

// NewAutonomousDefenseEngine initializes the autonomous defense engine.
func NewAutonomousDefenseEngine(ia *ImpactAnalyzer) *AutonomousDefenseEngine {
	if ia == nil {
		ia = NewImpactAnalyzer()
	}
	return &AutonomousDefenseEngine{
		records:        make(map[string]*AutonomousDefenseRecord),
		impactAnalyzer: ia,
	}
}

// TriggerAutonomousDefense validates safety and executes containment.
func (e *AutonomousDefenseEngine) TriggerAutonomousDefense(
	actionType DefenseActionType,
	targetEntityID, reason string,
	estimatedUsers int,
	exposure, confidence float64,
) (*AutonomousDefenseRecord, error) {
	// 1. Impact & blast radius verification
	report := e.impactAnalyzer.AnalyzeImpact(string(actionType), targetEntityID, estimatedUsers, exposure, confidence)
	if !report.IsSafeToDeploy {
		return nil, fmt.Errorf("autonomous defense rejected by impact guard: %s", report.RejectionReason)
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now().UTC()
	defenseID := fmt.Sprintf("def_%d_%s", now.UnixNano(), targetEntityID)

	rec := &AutonomousDefenseRecord{
		DefenseID:      defenseID,
		ActionType:     actionType,
		TargetEntityID: targetEntityID,
		Reason:         reason,
		Confidence:     confidence,
		ExecutedAt:     now,
		IsRolledBack:   false,
	}

	e.records[defenseID] = rec
	return rec, nil
}

// RollbackDefense reverses an autonomous defense action.
func (e *AutonomousDefenseEngine) RollbackDefense(defenseID, reason string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	rec, exists := e.records[defenseID]
	if !exists {
		return fmt.Errorf("defense record '%s' not found", defenseID)
	}
	if rec.IsRolledBack {
		return fmt.Errorf("defense action '%s' is already rolled back", defenseID)
	}

	now := time.Now().UTC()
	rec.IsRolledBack = true
	rec.RolledBackAt = &now
	return nil
}
