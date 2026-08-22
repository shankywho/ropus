package agents

import (
	"fmt"
	"time"
)

// ResponseOption evaluates the tradeoffs of a candidate containment action.
type ResponseOption struct {
	ActionName           string  `json:"action_name"`
	FraudLossPrevented   float64 `json:"fraud_loss_prevented"`
	CustomerFrictionCost float64 `json:"customer_friction_cost"`
	NetBenefitScore      float64 `json:"net_benefit_score"`
	IsSelected           bool    `json:"is_selected"`
}

// ResponsePlanDetails details the multi-option tradeoff simulation.
type ResponsePlanDetails struct {
	PlanID          string           `json:"plan_id"`
	SelectedAction  string           `json:"selected_action"`
	OptionsEvaluated []ResponseOption `json:"options_evaluated"`
	Rationale       string           `json:"rationale"`
	PlannedAt       time.Time        `json:"planned_at"`
}

// AutonomousResponsePlanner simulates multiple containment avenues to select the option with optimal net utility.
type AutonomousResponsePlanner struct {
	policyAgent *PolicyReasoningAgent
}

// NewAutonomousResponsePlanner initializes the response planner.
func NewAutonomousResponsePlanner(pa *PolicyReasoningAgent) *AutonomousResponsePlanner {
	if pa == nil {
		pa = NewPolicyReasoningAgent()
	}
	return &AutonomousResponsePlanner{policyAgent: pa}
}

// PlanOptimalResponse evaluates Option A (Hard Block), Option B (Step-up MFA), Option C (Monitor).
func (p *AutonomousResponsePlanner) PlanOptimalResponse(exposure float64, confidence float64, affectedUsers int) *ResponsePlanDetails {
	now := time.Now().UTC()

	// Option A: Hard Block
	lossA := exposure * confidence
	fricA := float64(affectedUsers) * 20.0 // friction penalty per good user challenged
	netA := lossA - fricA

	// Option B: Step-Up MFA
	lossB := exposure * confidence * 0.95
	fricB := float64(affectedUsers) * 2.0 // minimal friction
	netB := lossB - fricB

	// Option C: Monitor
	lossC := 0.0
	fricC := 0.0
	netC := 0.0

	var selected string
	var rationale string

	if confidence >= 0.92 && affectedUsers <= 5 {
		selected = "BLOCK_DEVICE_AND_ACCOUNT"
		rationale = fmt.Sprintf("High confidence (%.2f) with minimal blast radius (%d users) justifies hard containment", confidence, affectedUsers)
	} else if confidence >= 0.65 {
		selected = "REQUIRE_STEP_UP_MFA"
		rationale = fmt.Sprintf("Moderate confidence (%.2f) or wider blast radius (%d users) indicates Step-Up MFA provides optimal net utility ($%.2f vs $%.2f)", confidence, affectedUsers, netB, netA)
	} else {
		selected = "MONITOR_ONLY"
		rationale = "Low confidence does not justify active friction"
	}

	options := []ResponseOption{
		{ActionName: "BLOCK_DEVICE_AND_ACCOUNT", FraudLossPrevented: lossA, CustomerFrictionCost: fricA, NetBenefitScore: netA, IsSelected: selected == "BLOCK_DEVICE_AND_ACCOUNT"},
		{ActionName: "REQUIRE_STEP_UP_MFA", FraudLossPrevented: lossB, CustomerFrictionCost: fricB, NetBenefitScore: netB, IsSelected: selected == "REQUIRE_STEP_UP_MFA"},
		{ActionName: "MONITOR_ONLY", FraudLossPrevented: lossC, CustomerFrictionCost: fricC, NetBenefitScore: netC, IsSelected: selected == "MONITOR_ONLY"},
	}

	return &ResponsePlanDetails{
		PlanID:           fmt.Sprintf("plan_%d", now.UnixNano()),
		SelectedAction:   selected,
		OptionsEvaluated: options,
		Rationale:        rationale,
		PlannedAt:        now,
	}
}
