package riskengine_test

import (
	"math"
	"testing"

	"github.com/shankywho/ropus/backend/internal/riskengine"
)

// TestBMRExhaustiveMathematicalProperties validates the rigorous mathematical integrity of Bayes Minimum Risk.
func TestBMRExhaustiveMathematicalProperties(t *testing.T) {
	cfg := riskengine.DefaultEconomicPolicyConfig()

	// 1. Monotonicity: E[Cost(ALLOW)] must increase strictly monotonically with p
	t.Run("ALLOW Cost Monotonicity with Probability", func(t *testing.T) {
		amount := 1000.0
		prevCost := -1.0
		for p := 0.0; p <= 1.0; p += 0.05 {
			res := riskengine.EvaluateCostSensitiveDecision(p, amount, cfg)
			if res.AllowCost < prevCost-1e-6 {
				t.Errorf("Violation of monotonicity: AllowCost decreased at p=%.2f (prev=%.2f, current=%.2f)",
					p, prevCost, res.AllowCost)
			}
			prevCost = res.AllowCost
		}
	})

	// 2. Monotonicity: E[Cost(DECLINE)] must decrease strictly monotonically with p
	t.Run("DECLINE Cost Monotonicity with Probability", func(t *testing.T) {
		amount := 1000.0
		prevCost := math.MaxFloat64
		for p := 0.0; p <= 1.0; p += 0.05 {
			res := riskengine.EvaluateCostSensitiveDecision(p, amount, cfg)
			if res.DeclineCost > prevCost+1e-6 {
				t.Errorf("Violation of monotonicity: DeclineCost increased at p=%.2f (prev=%.2f, current=%.2f)",
					p, prevCost, res.DeclineCost)
			}
			prevCost = res.DeclineCost
		}
	})

	// 3. Exact Mathematical Formula Equality
	t.Run("Exact Formula Validation Across Grid", func(t *testing.T) {
		probabilities := []float64{0.01, 0.05, 0.15, 0.35, 0.65, 0.85, 0.99}
		amounts := []float64{50.0, 200.0, 1500.0, 50000.0, 1000000.0}

		for _, p := range probabilities {
			for _, amt := range amounts {
				res := riskengine.EvaluateCostSensitiveDecision(p, amt, cfg)

				// Calculate expected values manually
				fpCost := cfg.FalsePositiveCostBase
				if amt > 0 && amt < 250.0 {
					fpCost = math.Max(25.0, math.Min(cfg.FalsePositiveCostBase, 2.0*amt))
				} else if amt*cfg.FalsePositiveRateLTV > cfg.FalsePositiveCostBase {
					fpCost = amt * cfg.FalsePositiveRateLTV
				}

				expectedAllow := math.Round((p*amt*cfg.FraudMultiplier)*100) / 100
				expectedDecline := math.Round(((1.0-p)*fpCost)*100) / 100
				expectedReview := math.Round((cfg.ManualReviewCost+(cfg.ResidualReviewRate*p*amt*cfg.FraudMultiplier))*100) / 100
				expectedChallenge := math.Round(((1.0-p)*(cfg.ChallengeFrictionRate*amt)+cfg.ChallengeCost)*100) / 100

				if math.Abs(res.AllowCost-expectedAllow) > 0.05 {
					t.Errorf("p=%.2f, amt=%.2f: AllowCost mismatch (got %.2f, expected %.2f)", p, amt, res.AllowCost, expectedAllow)
				}
				if math.Abs(res.DeclineCost-expectedDecline) > 0.05 {
					t.Errorf("p=%.2f, amt=%.2f: DeclineCost mismatch (got %.2f, expected %.2f)", p, amt, res.DeclineCost, expectedDecline)
				}
				if math.Abs(res.ReviewCost-expectedReview) > 0.05 {
					t.Errorf("p=%.2f, amt=%.2f: ReviewCost mismatch (got %.2f, expected %.2f)", p, amt, res.ReviewCost, expectedReview)
				}
				if math.Abs(res.ChallengeCost-expectedChallenge) > 0.05 {
					t.Errorf("p=%.2f, amt=%.2f: ChallengeCost mismatch (got %.2f, expected %.2f)", p, amt, res.ChallengeCost, expectedChallenge)
				}
			}
		}
	})

	// 4. Extreme Amount Boundary Conditions
	t.Run("Extreme Amounts (₹0.0 to ₹10 Crore)", func(t *testing.T) {
		// Zero amount
		res0 := riskengine.EvaluateCostSensitiveDecision(0.50, 0.0, cfg)
		if res0.AllowCost != 0.0 || res0.ExpectedFraudExposure != 0.0 {
			t.Errorf("Zero amount should have 0.0 allow cost and 0.0 exposure")
		}

		// Micro amount (₹1.00)
		resMicro := riskengine.EvaluateCostSensitiveDecision(0.01, 1.0, cfg)
		if resMicro.OptimalAction != "ALLOW_RECOMMENDATION" {
			t.Errorf("Clean ₹1 micro transaction should be ALLOW, got %s", resMicro.OptimalAction)
		}

		// Massive amount (₹100,000,000.00 / ₹10 Crore)
		// Even at 1% risk, exposure is ₹10 Lakhs -> Manual Review / Challenge
		resMassive := riskengine.EvaluateCostSensitiveDecision(0.01, 100000000.0, cfg)
		if resMassive.OptimalAction == "ALLOW_RECOMMENDATION" {
			t.Errorf("Blind ALLOW of ₹10 Crore at 1%% risk violates BMR loss minimization")
		}
	})
}
