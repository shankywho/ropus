package riskengine_test

import (
	"testing"

	"github.com/shankywho/ropus/backend/internal/riskengine"
)

func TestCostSensitiveDecisionPolicy(t *testing.T) {
	cfg := riskengine.DefaultEconomicPolicyConfig()

	t.Run("Low Probability + Low Amount (Clean Micro-Transaction)", func(t *testing.T) {
		p := 0.02
		amount := 150.0 // ₹150 INR

		res := riskengine.EvaluateCostSensitiveDecision(p, amount, cfg)

		if res.OptimalAction != "ALLOW_RECOMMENDATION" {
			t.Errorf("expected ALLOW_RECOMMENDATION for clean low amount, got %s", res.OptimalAction)
		}
		if res.ExpectedFraudExposure > 10.0 {
			t.Errorf("expected low fraud exposure (<10.0), got %.2f", res.ExpectedFraudExposure)
		}
		if res.AllowCost >= res.DeclineCost {
			t.Errorf("expected AllowCost (%.2f) < DeclineCost (%.2f)", res.AllowCost, res.DeclineCost)
		}
	})

	t.Run("High Probability + Low Amount (Card Testing Fraud)", func(t *testing.T) {
		p := 0.95
		amount := 150.0 // ₹150 INR

		res := riskengine.EvaluateCostSensitiveDecision(p, amount, cfg)

		// High probability fraud should decline or review, not allow
		if res.OptimalAction != "DECLINE_RECOMMENDATION" && res.OptimalAction != "MANUAL_REVIEW" {
			t.Errorf("expected DECLINE or REVIEW for high prob fraud, got %s", res.OptimalAction)
		}
		if res.DeclineCost >= res.AllowCost {
			t.Errorf("expected DeclineCost (%.2f) < AllowCost (%.2f)", res.DeclineCost, res.AllowCost)
		}
	})

	t.Run("Low Probability + High Amount (Legitimate Corporate Payout)", func(t *testing.T) {
		p := 0.05
		amount := 1450000.0 // ₹14.5 Lakhs INR

		res := riskengine.EvaluateCostSensitiveDecision(p, amount, cfg)

		// For high amount with 5% risk, manual review minimizes expected loss (₹3,725) compared to blind allow (₹72,500) or destructive hard decline (₹68,875)
		if res.OptimalAction != "MANUAL_REVIEW" && res.OptimalAction != "STEP_UP_RECOMMENDATION" {
			t.Errorf("expected MANUAL_REVIEW or STEP_UP for high exposure, got %s", res.OptimalAction)
		}
		if res.ReviewCost >= res.AllowCost {
			t.Errorf("expected ReviewCost (%.2f) < AllowCost (%.2f)", res.ReviewCost, res.AllowCost)
		}
		if res.ReviewCost >= res.DeclineCost {
			t.Errorf("expected ReviewCost (%.2f) < DeclineCost (%.2f)", res.ReviewCost, res.DeclineCost)
		}
	})

	t.Run("High Probability + High Amount (Major Syndicate Attack)", func(t *testing.T) {
		p := 0.94
		amount := 1450000.0 // ₹14.5 Lakhs INR

		res := riskengine.EvaluateCostSensitiveDecision(p, amount, cfg)

		if res.OptimalAction != "DECLINE_RECOMMENDATION" {
			t.Errorf("expected DECLINE_RECOMMENDATION for 94%% risk on ₹14.5L, got %s", res.OptimalAction)
		}
		// Expected fraud loss on allow is ₹13.63 Lakhs
		expectedLoss := 0.94 * 1450000.0
		if res.AllowCost < expectedLoss*0.99 {
			t.Errorf("expected AllowCost close to ₹%.2f, got ₹%.2f", expectedLoss, res.AllowCost)
		}
		if res.DeclineCost >= res.AllowCost {
			t.Errorf("expected DeclineCost (%.2f) to be far lower than AllowCost (%.2f)", res.DeclineCost, res.AllowCost)
		}
	})

	t.Run("Review and Challenge Boundary (Moderate Probability)", func(t *testing.T) {
		p := 0.50
		amount := 5000.0 // ₹5,000 INR

		res := riskengine.EvaluateCostSensitiveDecision(p, amount, cfg)

		if res.OptimalAction != "STEP_UP_RECOMMENDATION" {
			t.Errorf("expected STEP_UP_RECOMMENDATION to minimize loss for moderate risk, got %s", res.OptimalAction)
		}
		if res.ChallengeCost >= res.AllowCost {
			t.Errorf("expected ChallengeCost (%.2f) < AllowCost (%.2f)", res.ChallengeCost, res.AllowCost)
		}
	})

	t.Run("Zero and Negative Amount Handling", func(t *testing.T) {
		p := 0.10
		zeroRes := riskengine.EvaluateCostSensitiveDecision(p, 0.0, cfg)
		if zeroRes.ExpectedFraudExposure != 0.0 {
			t.Errorf("expected 0.0 exposure for 0.0 amount, got %.2f", zeroRes.ExpectedFraudExposure)
		}

		negRes := riskengine.EvaluateCostSensitiveDecision(p, -500.0, cfg)
		if negRes.ExpectedFraudExposure != 0.0 {
			t.Errorf("expected negative amount clamped to 0.0 exposure, got %.2f", negRes.ExpectedFraudExposure)
		}
		if negRes.TransactionAmount != 0.0 {
			t.Errorf("expected transaction amount clamped to 0.0, got %.2f", negRes.TransactionAmount)
		}
	})

	t.Run("Probability Clamping (Out of bounds inputs)", func(t *testing.T) {
		resHigh := riskengine.EvaluateCostSensitiveDecision(1.85, 1000.0, cfg)
		if resHigh.CalibratedProbability > 1.0 {
			t.Errorf("expected probability clamped to 1.0, got %.2f", resHigh.CalibratedProbability)
		}

		resLow := riskengine.EvaluateCostSensitiveDecision(-0.45, 1000.0, cfg)
		if resLow.CalibratedProbability < 0.0 {
			t.Errorf("expected probability clamped to 0.0, got %.2f", resLow.CalibratedProbability)
		}
	})
}
