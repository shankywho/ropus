package governance

import (
	"fmt"
	"sync"
	"time"
)

// PolicyRule defines an individual conditional logic override over the ML risk score.
type PolicyRule struct {
	RuleID         string  `json:"rule_id"`
	Description    string  `json:"description"`
	MinScore       float64 `json:"min_score"`
	MinAmount      float64 `json:"min_amount"`
	ActionOverride string  `json:"action_override"` // "REQUIRE_MANUAL_REVIEW", "BLOCK", "ALLOW"
	Priority       int     `json:"priority"`
	Enabled        bool    `json:"enabled"`
}

// GovernancePolicy encapsulates a versioned ruleset for regulatory threshold enforcement.
type GovernancePolicy struct {
	PolicyVersion string       `json:"policy_version"`
	Rules         []PolicyRule `json:"rules"`
	CreatedAt     time.Time    `json:"created_at"`
	Author        string       `json:"author"`
	Active        bool         `json:"active"`
}

// PolicyEngine enforces versioned governance rules and business policy constraints.
type PolicyEngine struct {
	mu             sync.RWMutex
	policies       map[string]*GovernancePolicy
	activeVersion  string
	versionHistory []string
}

// NewPolicyEngine initializes the policy engine with standard production governance policies.
func NewPolicyEngine() *PolicyEngine {
	pe := &PolicyEngine{
		policies: make(map[string]*GovernancePolicy),
	}
	pe.initializeBaselinePolicy()
	return pe
}

func (pe *PolicyEngine) initializeBaselinePolicy() {
	p1 := &GovernancePolicy{
		PolicyVersion: "policy_v1.0_standard",
		Author:        "chief_compliance_officer",
		CreatedAt:     time.Now().UTC(),
		Active:        true,
		Rules: []PolicyRule{
			{
				RuleID:         "rule_high_value_elevated_risk",
				Description:    "Flag transactions > $10,000 with risk score > 0.70 for mandatory manual review",
				MinScore:       0.70,
				MinAmount:      10000.0,
				ActionOverride: "REQUIRE_MANUAL_REVIEW",
				Priority:       1,
				Enabled:        true,
			},
			{
				RuleID:         "rule_extreme_risk_auto_block",
				Description:    "Hard block any transaction with risk score > 0.95",
				MinScore:       0.95,
				MinAmount:      0.0,
				ActionOverride: "BLOCK",
				Priority:       2,
				Enabled:        true,
			},
		},
	}
	pe.policies[p1.PolicyVersion] = p1
	pe.activeVersion = p1.PolicyVersion
	pe.versionHistory = append(pe.versionHistory, p1.PolicyVersion)
}

// EvaluatePolicies checks input transaction characteristics against active governance policies.
func (pe *PolicyEngine) EvaluatePolicies(riskScore, amount float64) (string, string, bool) {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	activePol, exists := pe.policies[pe.activeVersion]
	if !exists {
		return "", pe.activeVersion, false
	}

	for _, r := range activePol.Rules {
		if !r.Enabled {
			continue
		}
		if riskScore >= r.MinScore && amount >= r.MinAmount {
			return r.ActionOverride, activePol.PolicyVersion, true
		}
	}

	return "", pe.activeVersion, false
}

// DeployPolicy activates a new policy version with audit history.
func (pe *PolicyEngine) DeployPolicy(policy *GovernancePolicy) error {
	if policy == nil || policy.PolicyVersion == "" {
		return fmt.Errorf("invalid governance policy")
	}

	pe.mu.Lock()
	defer pe.mu.Unlock()

	// Deactivate existing active policies
	if cur, exists := pe.policies[pe.activeVersion]; exists {
		cur.Active = false
	}

	policy.Active = true
	policy.CreatedAt = time.Now().UTC()
	pe.policies[policy.PolicyVersion] = policy
	pe.activeVersion = policy.PolicyVersion
	pe.versionHistory = append(pe.versionHistory, policy.PolicyVersion)
	return nil
}

// RollbackPolicy rolls back to the previous policy version in history.
func (pe *PolicyEngine) RollbackPolicy() (string, error) {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	if len(pe.versionHistory) < 2 {
		return "", fmt.Errorf("no previous policy version available for rollback")
	}

	// Current becomes inactive
	if cur, exists := pe.policies[pe.activeVersion]; exists {
		cur.Active = false
	}

	// Pop current version and point to previous
	pe.versionHistory = pe.versionHistory[:len(pe.versionHistory)-1]
	prevVersion := pe.versionHistory[len(pe.versionHistory)-1]

	prevPol, exists := pe.policies[prevVersion]
	if !exists {
		return "", fmt.Errorf("target rollback policy '%s' not found", prevVersion)
	}

	prevPol.Active = true
	pe.activeVersion = prevVersion
	return prevVersion, nil
}
