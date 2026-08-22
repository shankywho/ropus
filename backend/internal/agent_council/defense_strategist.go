package agent_council

import (
	"fmt"
	"time"
)

// DefenseCandidateStrategy encapsulates a strategic defense response package.
type DefenseCandidateStrategy struct {
	StrategyName        string  `json:"strategy_name"`
	ActionPlan          string  `json:"action_plan"`
	ProjectedSavings    float64 `json:"projected_savings"`
	CustomerFrictionCost float64 `json:"customer_friction_cost"`
	ComplianceRiskScore float64 `json:"compliance_risk_score"`
	NetScore            float64 `json:"net_score"`
}

// StrategyPlanReport details the chosen strategic containment path.
type StrategyPlanReport struct {
	PlanID             string                     `json:"plan_id"`
	SelectedStrategy   string                     `json:"selected_strategy"`
	EvaluatedOptions   []DefenseCandidateStrategy `json:"evaluated_options"`
	ExecutiveRationale string                     `json:"executive_rationale"`
	PlannedAt          time.Time                  `json:"planned_at"`
}

// AutonomousDefenseStrategist evaluates multiple macro-containment strategies before high-stakes deployment.
type AutonomousDefenseStrategist struct{}

// NewAutonomousDefenseStrategist initializes the defense strategist.
func NewAutonomousDefenseStrategist() *AutonomousDefenseStrategist {
	return &AutonomousDefenseStrategist{}
}

// SelectOptimalStrategy scores Strategy A (Aggressive Block), Strategy B (Adaptive Monitor), Strategy C (Step-up Verification), Strategy D (Network Isolation).
func (s *AutonomousDefenseStrategist) SelectOptimalStrategy(exposure float64, confidence float64, impactedUsers int) *StrategyPlanReport {
	now := time.Now().UTC()

	// 1. Strategy A: Aggressive Blocking
	savA := exposure * confidence
	fricA := float64(impactedUsers) * 25.0
	compA := 0.20
	netA := savA - fricA - (compA * 1000.0)

	// 2. Strategy B: Adaptive Monitoring
	savB := exposure * 0.40
	fricB := 0.0
	compB := 0.0
	netB := savB

	// 3. Strategy C: Step-up Verification
	savC := exposure * confidence * 0.95
	fricC := float64(impactedUsers) * 3.0
	compC := 0.02
	netC := savC - fricC - (compC * 1000.0)

	// 4. Strategy D: Network Isolation
	savD := exposure * 0.99
	fricD := float64(impactedUsers) * 50.0
	compD := 0.40
	netD := savD - fricD - (compD * 1000.0)

	options := []DefenseCandidateStrategy{
		{StrategyName: "STRATEGY_AGGRESSIVE_BLOCK", ActionPlan: "Hard block device fingerprint and immediate account freeze", ProjectedSavings: savA, CustomerFrictionCost: fricA, ComplianceRiskScore: compA, NetScore: netA},
		{StrategyName: "STRATEGY_ADAPTIVE_MONITOR", ActionPlan: "Allow transactions with enhanced velocity telemetry tracking", ProjectedSavings: savB, CustomerFrictionCost: fricB, ComplianceRiskScore: compB, NetScore: netB},
		{StrategyName: "STRATEGY_STEP_UP_VERIFICATION", ActionPlan: "Challenge high risk users with biometric / hardware MFA", ProjectedSavings: savC, CustomerFrictionCost: fricC, ComplianceRiskScore: compC, NetScore: netC},
		{StrategyName: "STRATEGY_NETWORK_ISOLATION", ActionPlan: "Isolate entire proxy subnet from checkout gateways", ProjectedSavings: savD, CustomerFrictionCost: fricD, ComplianceRiskScore: compD, NetScore: netD},
	}

	selected := "STRATEGY_STEP_UP_VERIFICATION"
	bestNet := netC

	if netA > bestNet && compA < 0.30 && impactedUsers <= 10 {
		selected = "STRATEGY_AGGRESSIVE_BLOCK"
		bestNet = netA
	}

	rationale := fmt.Sprintf("Selected %s with highest net benefit ($%.2f) and low regulatory friction", selected, bestNet)

	return &StrategyPlanReport{
		PlanID:             fmt.Sprintf("strat_%d", now.UnixNano()),
		SelectedStrategy:   selected,
		EvaluatedOptions:   options,
		ExecutiveRationale: rationale,
		PlannedAt:          now,
	}
}
