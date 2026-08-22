"""
Unit & Integration Tests for Phase 2.3: Production Beta Calibration, Cost Policy, Safety, and Rollback
"""

import os
import sys
import json
import unittest
import numpy as np
import pandas as pd

# Add ml-service root to path
ml_root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
if ml_root not in sys.path:
    sys.path.insert(0, ml_root)

from calibration.calibrator import ModelCalibrator
from calibration.cost_policy import (
    CostSensitivePolicyEngine,
    run_cost_threshold_analysis,
    run_review_capacity_analysis,
    run_cost_sensitivity_scenarios
)

class TestPhase2ProductionBetaCalibration(unittest.TestCase):

    def setUp(self):
        np.random.seed(42)
        # Synthetic validation probabilities and binary labels
        self.y_val = np.random.binomial(1, 0.05, size=500)
        self.y_prob_val = np.clip(self.y_val * 0.7 + np.random.uniform(0.01, 0.35, size=500), 0.01, 0.99)
        
        self.y_test = np.random.binomial(1, 0.05, size=500)
        self.y_prob_test = np.clip(self.y_test * 0.7 + np.random.uniform(0.01, 0.35, size=500), 0.01, 0.99)
        self.amounts_test = np.random.uniform(50.0, 5000.0, size=500)

    # -------------------------------------------------------------
    # 1. Beta Calibration Fit & Predict Tests
    # -------------------------------------------------------------
    def test_01_fit_predict_beta_calibrator(self):
        """Test Beta calibrator fitting and continuous probability predictions."""
        calibrator = ModelCalibrator(method="beta")
        calibrator.fit(self.y_prob_val, self.y_val)
        self.assertTrue(calibrator.is_fitted)
        
        preds = calibrator.predict_proba(self.y_prob_test)
        self.assertEqual(len(preds), len(self.y_prob_test))
        self.assertTrue(np.all(preds >= 0.0001))
        self.assertTrue(np.all(preds <= 0.9999))
        self.assertGreater(len(np.unique(np.round(preds, 4))), 50)

    def test_02_serialization_deserialization_beta(self):
        """Test Beta calibrator JSON serialization and deserialization."""
        calibrator = ModelCalibrator(method="beta")
        calibrator.fit(self.y_prob_val, self.y_val)
        
        state_dict = calibrator.to_dict()
        self.assertEqual(state_dict["type"], "beta")
        self.assertIn("parameters", state_dict)
        self.assertIn("beta_params", state_dict["parameters"])
        self.assertIn("checksum_sha256", state_dict)
        
        reloaded = ModelCalibrator.from_dict(state_dict)
        self.assertTrue(reloaded.is_fitted)
        self.assertEqual(reloaded.method, "beta")
        
        orig_preds = calibrator.predict_proba(self.y_prob_test[:20])
        reloaded_preds = reloaded.predict_proba(self.y_prob_test[:20])
        np.testing.assert_allclose(orig_preds, reloaded_preds, rtol=1e-5)

    def test_03_deterministic_serialization_and_checksum(self):
        """Test deterministic JSON artifact serialization and SHA-256 hash consistency."""
        cal1 = ModelCalibrator(method="beta").fit(self.y_prob_val, self.y_val)
        cal2 = ModelCalibrator(method="beta").fit(self.y_prob_val, self.y_val)
        
        dict1 = cal1.to_dict()
        dict2 = cal2.to_dict()
        self.assertEqual(dict1["checksum_sha256"], dict2["checksum_sha256"])

    # -------------------------------------------------------------
    # 2. Mathematical Safety & Extreme Probability Input Tests
    # -------------------------------------------------------------
    def test_04_extreme_probabilities_and_bounds(self):
        """Test extreme input probabilities: 0, 1e-12, 1e-9, 1e-6, 0.01, 0.5, 0.99, 0.999999, 1-1e-12, 1."""
        calibrator = ModelCalibrator(method="beta")
        calibrator.fit(self.y_prob_val, self.y_val)
        
        extreme_inputs = np.array([0.0, 1e-12, 1e-9, 1e-6, 0.01, 0.5, 0.99, 0.999999, 1.0 - 1e-12, 1.0])
        cal_probs = calibrator.predict_proba(extreme_inputs)
        
        self.assertEqual(len(cal_probs), len(extreme_inputs))
        self.assertTrue(np.all(np.isfinite(cal_probs)))
        self.assertTrue(np.all(cal_probs >= 0.0001))
        self.assertTrue(np.all(cal_probs <= 0.9999))
        self.assertEqual(np.sum(np.isnan(cal_probs)), 0)

    def test_05_nan_and_inf_handling(self):
        """Test safe handling of NaN, +Inf, -Inf, and out-of-range negative values."""
        calibrator = ModelCalibrator(method="beta")
        calibrator.fit(self.y_prob_val, self.y_val)
        
        anomaly_inputs = np.array([np.nan, np.inf, -np.inf, -5.0, 10.0, 999.0])
        cal_probs = calibrator.predict_proba(anomaly_inputs)
        
        self.assertEqual(len(cal_probs), len(anomaly_inputs))
        self.assertEqual(np.sum(np.isnan(cal_probs)), 0)
        self.assertEqual(np.sum(np.isinf(cal_probs)), 0)
        self.assertTrue(np.all(cal_probs >= 0.0001))
        self.assertTrue(np.all(cal_probs <= 0.9999))

    def test_06_strict_monotonicity_guarantee(self):
        """Test that Beta calibration strictly preserves monotonic probability ordering: p1 < p2 => cal(p1) <= cal(p2)."""
        calibrator = ModelCalibrator(method="beta")
        calibrator.fit(self.y_prob_val, self.y_val)
        
        dense_grid = np.linspace(1e-5, 1.0 - 1e-5, 1000)
        cal_grid = calibrator.predict_proba(dense_grid)
        
        diffs = np.diff(cal_grid)
        self.assertTrue(np.all(diffs >= -1e-7), "Beta calibration inverted risk ordering!")

    # -------------------------------------------------------------
    # 3. Rollback & Parity Tests
    # -------------------------------------------------------------
    def test_07_rollback_isotonic_artifact_loading(self):
        """Test that previous Isotonic calibration artifact can still be deserialized for rollback."""
        cal_iso = ModelCalibrator(method="isotonic").fit(self.y_prob_val, self.y_val)
        iso_dict = cal_iso.to_dict()
        
        reloaded_iso = ModelCalibrator.from_dict(iso_dict)
        self.assertEqual(reloaded_iso.method, "isotonic")
        self.assertTrue(reloaded_iso.is_fitted)
        
        preds = reloaded_iso.predict_proba(self.y_prob_test[:10])
        self.assertTrue(np.all(preds >= 0.0001))
        self.assertTrue(np.all(preds <= 0.9999))

    def test_08_cost_engine_consumes_calibrated_probability(self):
        """Verify that CostSensitivePolicyEngine calculates expected cost using calibrated probability."""
        policy = CostSensitivePolicyEngine(
            false_positive_cost=500.0,
            manual_review_cost=100.0,
            fraud_multiplier=1.0,
            residual_review_rate=0.05
        )
        
        p_raw = 0.85
        p_cal = 0.03
        amt = 2000.0
        
        # If cost engine receives p_cal = 0.03:
        # Cost(ALLOW) = 0.03 * 2000 = 60.0
        # Cost(DECLINE) = 0.97 * 500 = 485.0
        # Cost(REVIEW) = 100 + 0.05 * 0.03 * 2000 = 103.0
        # Optimal action: ALLOW
        costs = policy.calculate_expected_costs(p_calibrated=p_cal, amount=amt)
        self.assertEqual(costs["ALLOW"], 60.0)
        self.assertEqual(costs["DECLINE"], 485.0)
        action, _ = policy.select_action(p_calibrated=p_cal, amount=amt)
        self.assertEqual(action, "ALLOW")

    def test_09_expected_cost_formulas(self):
        """Test mathematical formulas for ALLOW, MANUAL_REVIEW, and DECLINE costs."""
        policy = CostSensitivePolicyEngine(
            false_positive_cost=500.0,
            manual_review_cost=100.0,
            fraud_multiplier=1.0,
            residual_review_rate=0.05
        )
        costs_high = policy.calculate_expected_costs(p_calibrated=0.90, amount=10000.0)
        self.assertEqual(costs_high["ALLOW"], 9000.0)
        self.assertEqual(costs_high["DECLINE"], 50.0)
        self.assertEqual(costs_high["MANUAL_REVIEW"], 550.0)
        action_high, _ = policy.select_action(0.90, 10000.0)
        self.assertEqual(action_high, "DECLINE")

    def test_10_review_capacity_allocations(self):
        """Test review capacity ranking and allocations at 1%, 5%, 10%, 20%."""
        policy = CostSensitivePolicyEngine()
        import tempfile
        with tempfile.NamedTemporaryFile(suffix=".csv", delete=False) as f:
            temp_csv = f.name
            
        df_cap = run_review_capacity_analysis(
            y_true=self.y_test,
            p_calibrated=self.y_prob_test,
            amounts=self.amounts_test,
            policy=policy,
            output_csv=temp_csv,
            capacities=[0.01, 0.05, 0.10, 0.20]
        )
        self.assertEqual(len(df_cap), 4)
        os.remove(temp_csv)

    def test_11_cost_threshold_sweep_generation(self):
        """Test cost threshold sweep generates valid 99-step table."""
        policy = CostSensitivePolicyEngine()
        import tempfile
        with tempfile.NamedTemporaryFile(suffix=".csv", delete=False) as f_csv, tempfile.NamedTemporaryFile(suffix=".png", delete=False) as f_png:
            temp_csv = f_csv.name
            temp_png = f_png.name
            
        df_sweep = run_cost_threshold_analysis(
            y_true=self.y_test,
            p_calibrated=self.y_prob_test,
            amounts=self.amounts_test,
            policy=policy,
            output_csv=temp_csv,
            output_png=temp_png
        )
        self.assertEqual(len(df_sweep), 99)
        os.remove(temp_csv)
        os.remove(temp_png)

if __name__ == "__main__":
    unittest.main()
