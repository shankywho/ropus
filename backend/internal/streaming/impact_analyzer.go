package streaming

import (
	"fmt"
)

// ImpactReport assesses the blast radius and potential business disruption of an autonomous defense action.
type ImpactReport struct {
	TargetEntityID          string  `json:"target_entity_id"`
	ActionType              string  `json:"action_type"`
	AffectedUsersCount      int     `json:"affected_users_count"`
	AffectedMerchantsCount  int     `json:"affected_merchants_count"`
	FinancialExposure       float64 `json:"financial_exposure"`
	ExpectedRiskReduction   float64 `json:"expected_risk_reduction"` // 0.0 to 1.0 (Higher is better)
	FalsePositiveRisk       float64 `json:"false_positive_risk"`     // 0.0 to 1.0 (Lower is better)
	IsSafeToDeploy          bool    `json:"is_safe_to_deploy"`
	RejectionReason         string  `json:"rejection_reason,omitempty"`
}

// ImpactAnalyzer calculates collateral impact and safety limits before mass containment deployment.
type ImpactAnalyzer struct {
	MaxAllowedFPRisk      float64
	MaxAllowedUserImpact  int
}

// NewImpactAnalyzer initializes the blast radius impact analyzer.
func NewImpactAnalyzer() *ImpactAnalyzer {
	return &ImpactAnalyzer{
		MaxAllowedFPRisk:     0.05, // 5% max false positive risk for autonomous mass actions
		MaxAllowedUserImpact: 1000, // Maximum users impacted without manual executive sign-off
	}
}

// AnalyzeImpact calculates blast radius and assesses deployment safety.
func (a *ImpactAnalyzer) AnalyzeImpact(actionType, targetEntityID string, estimatedUsers int, exposure float64, confidence float64) *ImpactReport {
	fpRisk := 1.0 - confidence
	riskReduction := confidence * 0.95

	safe := true
	var reason string

	if fpRisk > a.MaxAllowedFPRisk {
		safe = false
		reason = fmt.Sprintf("False positive risk %.2f exceeds maximum tolerance %.2f", fpRisk, a.MaxAllowedFPRisk)
	} else if estimatedUsers > a.MaxAllowedUserImpact {
		safe = false
		reason = fmt.Sprintf("Impact radius of %d users exceeds autonomous limit %d (requires manual sign-off)", estimatedUsers, a.MaxAllowedUserImpact)
	}

	return &ImpactReport{
		TargetEntityID:         targetEntityID,
		ActionType:             actionType,
		AffectedUsersCount:     estimatedUsers,
		AffectedMerchantsCount: 1,
		FinancialExposure:      exposure,
		ExpectedRiskReduction:  riskReduction,
		FalsePositiveRisk:      fpRisk,
		IsSafeToDeploy:         safe,
		RejectionReason:        reason,
	}
}
