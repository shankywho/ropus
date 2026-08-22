package riskengine

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// BudgetPolicyAction describes the automated action enforced by error budget policies.
type BudgetPolicyAction string

const (
	ActionBudgetHealthy            BudgetPolicyAction = "BUDGET_HEALTHY"
	ActionThrottlePromotions       BudgetPolicyAction = "THROTTLE_PROMOTIONS"
	ActionFreezePromotions         BudgetPolicyAction = "FREEZE_PROMOTIONS"
	ActionEmergencyModelFreeze     BudgetPolicyAction = "EMERGENCY_MODEL_FREEZE"
)

// BudgetPolicyEvaluation encapsulates the outcome of error budget evaluation.
type BudgetPolicyEvaluation struct {
	Timestamp            time.Time          `json:"timestamp"`
	MinRemainingBudget   float64            `json:"min_remaining_budget_percent"`
	ConstrainedSLO       string             `json:"constrained_slo"`
	Action               BudgetPolicyAction `json:"action"`
	PromotionPermitted   bool               `json:"promotion_permitted"`
	AutoFrozenEnacted    bool               `json:"auto_frozen_enacted"`
	Reason               string             `json:"reason"`
}

// ErrorBudgetPolicyEngine monitors SLO error budgets and enforces automated safeguards on model lifecycle.
type ErrorBudgetPolicyEngine struct {
	mu          sync.RWMutex
	sloEngine   *SLOEngine
	coordinator *RetrainingCoordinator
	lastEval    BudgetPolicyEvaluation
}

// NewErrorBudgetPolicyEngine initializes the error budget automation engine.
func NewErrorBudgetPolicyEngine(sloEngine *SLOEngine, coordinator *RetrainingCoordinator) *ErrorBudgetPolicyEngine {
	return &ErrorBudgetPolicyEngine{
		sloEngine:   sloEngine,
		coordinator: coordinator,
		lastEval: BudgetPolicyEvaluation{
			Timestamp:          time.Now().UTC(),
			MinRemainingBudget: 100.0,
			Action:             ActionBudgetHealthy,
			PromotionPermitted: true,
		},
	}
}

// Evaluate evaluates current error budgets across all SLOs and enacts automated safety controls.
func (pe *ErrorBudgetPolicyEngine) Evaluate(ctx context.Context) BudgetPolicyEvaluation {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	now := time.Now().UTC()
	eval := BudgetPolicyEvaluation{
		Timestamp:          now,
		MinRemainingBudget: 100.0,
		Action:             ActionBudgetHealthy,
		PromotionPermitted: true,
	}

	if pe.sloEngine == nil {
		eval.Reason = "SLO engine not configured; defaulting to permitted"
		pe.lastEval = eval
		return eval
	}

	summary := pe.sloEngine.Evaluate(now)
	for id, m := range summary.Measurements {
		if m.ErrorBudgetRemaining < eval.MinRemainingBudget {
			eval.MinRemainingBudget = m.ErrorBudgetRemaining
			eval.ConstrainedSLO = id
		}
	}

	// Policy Rules
	if eval.MinRemainingBudget <= 0.0 {
		// Budget exhausted: Emergency Model Freeze
		eval.Action = ActionEmergencyModelFreeze
		eval.PromotionPermitted = false
		eval.Reason = fmt.Sprintf("Error budget exhausted on %s (%.2f%% remaining); enacting emergency model freeze",
			eval.ConstrainedSLO, eval.MinRemainingBudget)

		if pe.coordinator != nil {
			controls := pe.coordinator.GetOperationalControls()
			if frozen, ok := controls["model_frozen"].(bool); !ok || !frozen {
				_ = pe.coordinator.SetModelFrozen(ctx, true, "ERROR_BUDGET_POLICY", eval.Reason)
				eval.AutoFrozenEnacted = true
				log.Printf("[ERROR_BUDGET_POLICY] Automatically enabled model freeze: %s", eval.Reason)
			}
		}
	} else if eval.MinRemainingBudget < 10.0 {
		// Budget < 10%: Freeze promotions
		eval.Action = ActionFreezePromotions
		eval.PromotionPermitted = false
		eval.Reason = fmt.Sprintf("Error budget critically low on %s (%.2f%% remaining); freezing model promotions",
			eval.ConstrainedSLO, eval.MinRemainingBudget)

		if pe.coordinator != nil {
			controls := pe.coordinator.GetOperationalControls()
			if frozen, ok := controls["model_frozen"].(bool); !ok || !frozen {
				_ = pe.coordinator.SetModelFrozen(ctx, true, "ERROR_BUDGET_POLICY", eval.Reason)
				eval.AutoFrozenEnacted = true
			}
		}
	} else if eval.MinRemainingBudget < 25.0 {
		// Budget < 25%: Throttle promotions
		eval.Action = ActionThrottlePromotions
		eval.PromotionPermitted = false
		eval.Reason = fmt.Sprintf("Error budget warning on %s (%.2f%% remaining); non-essential promotions blocked",
			eval.ConstrainedSLO, eval.MinRemainingBudget)
	} else {
		eval.Action = ActionBudgetHealthy
		eval.PromotionPermitted = true
		eval.Reason = "All SLO error budgets healthy"
	}

	pe.lastEval = eval
	return eval
}

// GetLastEvaluation returns the most recent error budget policy evaluation.
func (pe *ErrorBudgetPolicyEngine) GetLastEvaluation() BudgetPolicyEvaluation {
	pe.mu.RLock()
	defer pe.mu.RUnlock()
	return pe.lastEval
}
