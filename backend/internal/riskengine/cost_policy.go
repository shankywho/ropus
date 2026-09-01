package riskengine

import (
	"fmt"
	"math"
)

// EconomicPolicyConfig specifies the parameters for Bayes Minimum Risk cost-sensitive decisioning.
type EconomicPolicyConfig struct {
	CurrencySymbol        string  `json:"currency_symbol,omitempty"` // Currency denomination for human-readable explanations (default: "₹")
	FalsePositiveCostBase float64 `json:"false_positive_cost_base"`  // Base minimum cost of false decline (default: 500.0)
	FalsePositiveRateLTV  float64 `json:"false_positive_rate_ltv"`   // Example-dependent LTV/churn loss rate on large declined volumes (default: 0.05)
	ManualReviewCost      float64 `json:"manual_review_cost"`        // Fixed operational cost of human analyst review (default: 100.0)
	ChallengeCost         float64 `json:"challenge_cost"`            // Fixed cost of Step-Up verification (SMS OTP/gateway fee, default: 2.0)
	ChallengeFrictionRate float64 `json:"challenge_friction_rate"`   // Drop-off friction rate on legitimate volume under challenge (default: 0.02)
	FraudMultiplier       float64 `json:"fraud_multiplier"`          // Multiplier including chargeback fees/dispute penalties (default: 1.0)
	ResidualReviewRate    float64 `json:"residual_review_rate"`      // Slip-through fraud rate under manual review (default: 0.05)
}

// DefaultEconomicPolicyConfig returns production standard loss parameters for Indian payments context (INR ₹).
func DefaultEconomicPolicyConfig() EconomicPolicyConfig {
	return EconomicPolicyConfig{
		CurrencySymbol:        "₹",
		FalsePositiveCostBase: 500.0,
		FalsePositiveRateLTV:  0.05,
		ManualReviewCost:      100.0,
		ChallengeCost:         2.0,
		ChallengeFrictionRate: 0.02,
		FraudMultiplier:       1.0,
		ResidualReviewRate:    0.05,
	}
}

// ActionCostBreakdown holds the expected monetary losses for each candidate risk action.
type ActionCostBreakdown struct {
	Currency              string             `json:"currency,omitempty"`
	CalibratedProbability float64            `json:"calibrated_probability"`
	TransactionAmount     float64            `json:"transaction_amount"`
	ExpectedFraudExposure float64            `json:"expected_fraud_exposure"` // p * amount
	AllowCost             float64            `json:"allow_cost"`              // E[Cost(ALLOW)]
	ChallengeCost         float64            `json:"challenge_cost"`          // E[Cost(CHALLENGE)]
	ReviewCost            float64            `json:"review_cost"`             // E[Cost(REVIEW)]
	DeclineCost           float64            `json:"decline_cost"`            // E[Cost(DECLINE)]
	OptimalAction         string             `json:"optimal_action"`          // Action with minimum expected cost
	ActionCosts           map[string]float64 `json:"action_costs"`
	DecisionReason        string             `json:"decision_reason"`
}

// EvaluateCostSensitiveDecision computes expected monetary loss across candidate actions
// using the Bayes Minimum Risk principle: E[Cost(Action)] and selects the optimal action.
func EvaluateCostSensitiveDecision(p float64, amount float64, cfg EconomicPolicyConfig) ActionCostBreakdown {
	// Guard against NaN or infinite inputs from upstream telemetry/anomaly
	if math.IsNaN(p) || math.IsInf(p, 0) {
		p = 0.5 // Default to cautious median probability on numerical error
	}
	if math.IsNaN(amount) || math.IsInf(amount, 0) {
		amount = 0.0
	}

	// Clamp probability strictly to [0.0, 1.0]
	p = math.Max(0.0, math.Min(1.0, p))
	amt := math.Max(0.0, amount)

	// Example-dependent False Positive Cost:
	// For micro-transactions (amt < 250), dampens down to 2x amount (min 25).
	// For large transactions, scales with LTV/merchant churn impact: max(base, ltv_rate * amount).
	fpCost := cfg.FalsePositiveCostBase
	if amt > 0 && amt < 250.0 {
		fpCost = math.Max(25.0, math.Min(cfg.FalsePositiveCostBase, 2.0*amt))
	} else if amt*cfg.FalsePositiveRateLTV > cfg.FalsePositiveCostBase {
		fpCost = amt * cfg.FalsePositiveRateLTV
	}

	// 1. Expected Fraud Exposure: E[Exposure] = P(Fraud | x) * Amount
	expectedExposure := p * amt

	// 2. Candidate Action Costs (Bahnsen et al., 2015):
	// E[Cost(ALLOW)]     = p * amount * fraud_multiplier
	// E[Cost(DECLINE)]   = (1 - p) * false_positive_cost
	// E[Cost(REVIEW)]    = review_cost + (residual_review_rate * p * amount * fraud_multiplier)
	// E[Cost(CHALLENGE)] = (1 - p) * (challenge_friction_rate * amount) + cfg.ChallengeCost
	costAllow := p * amt * cfg.FraudMultiplier
	costDecline := (1.0 - p) * fpCost
	costReview := cfg.ManualReviewCost + (cfg.ResidualReviewRate * p * amt * cfg.FraudMultiplier)
	costChallenge := (1.0-p)*(cfg.ChallengeFrictionRate*amt) + cfg.ChallengeCost

	// Format rounded costs
	costs := map[string]float64{
		"ALLOW_RECOMMENDATION":   math.Round(costAllow*100) / 100,
		"STEP_UP_RECOMMENDATION": math.Round(costChallenge*100) / 100,
		"MANUAL_REVIEW":          math.Round(costReview*100) / 100,
		"DECLINE_RECOMMENDATION": math.Round(costDecline*100) / 100,
	}

	// 3. Select action minimizing expected cost
	bestAction := "ALLOW_RECOMMENDATION"
	minCost := costAllow

	// Step-Up challenge is activated when risk probability justifies friction (p >= 0.15)
	// OR when expected fraud exposure is elevated (exposure >= 2500.0)
	if (p >= 0.15 || expectedExposure >= 2500.0) && costChallenge < minCost {
		minCost = costChallenge
		bestAction = "STEP_UP_RECOMMENDATION"
	}
	// Manual review is activated when risk probability is elevated (p >= 0.25)
	// OR when expected fraud exposure is significant (exposure >= 5000.0)
	if (p >= 0.25 || expectedExposure >= 5000.0) && costReview < minCost {
		minCost = costReview
		bestAction = "MANUAL_REVIEW"
	}
	if costDecline < minCost {
		minCost = costDecline
		bestAction = "DECLINE_RECOMMENDATION"
	}

	// High probability safety floor: When probability is critical (>= 0.80),
	// never allow step-up or raw allow; must DECLINE or REVIEW
	if p >= 0.80 {
		if costDecline <= costReview {
			bestAction = "DECLINE_RECOMMENDATION"
			minCost = costDecline
		} else {
			bestAction = "MANUAL_REVIEW"
			minCost = costReview
		}
	}

	currency := cfg.CurrencySymbol
	if currency == "" {
		currency = "₹"
	}

	reason := fmt.Sprintf("Bayes Minimum Risk: %s minimizes expected loss (ExpLoss: %s%.2f, ExpFraud: %s%.2f for Amount: %s%.2f, P(Fraud): %.2f)",
		bestAction, currency, minCost, currency, expectedExposure, currency, amt, p)

	return ActionCostBreakdown{
		Currency:              currency,
		CalibratedProbability: math.Round(p*1000) / 1000,
		TransactionAmount:     math.Round(amt*100) / 100,
		ExpectedFraudExposure: math.Round(expectedExposure*100) / 100,
		AllowCost:             costs["ALLOW_RECOMMENDATION"],
		ChallengeCost:         costs["STEP_UP_RECOMMENDATION"],
		ReviewCost:            costs["MANUAL_REVIEW"],
		DeclineCost:           costs["DECLINE_RECOMMENDATION"],
		OptimalAction:         bestAction,
		ActionCosts:           costs,
		DecisionReason:        reason,
	}
}
