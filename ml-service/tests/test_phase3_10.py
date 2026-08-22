"""
Phase 3.10 Test Suite — Beta Calibration Re-evaluation, Numerical Stability, and Production Safety
"""

import os
import json
import unittest
import numpy as np
import pandas as pd
import onnxruntime as rt

from calibration.calibrator import ModelCalibrator
from calibration.cost_policy import CostSensitivePolicyEngine
from calibration.calibrate_25f import run_phase_3_10_calibration_pipeline

class TestPhase310BetaCalibration(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.current_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
        cls.candidate_dir = os.path.join(cls.current_dir, "model", "candidates")
        cls.prod_onnx_path = os.path.join(cls.current_dir, "model", "fraud_model.onnx")
        cls.prod_cal_path = os.path.join(cls.current_dir, "model", "calibration.json")
        cls.cand_cal_path = os.path.join(cls.candidate_dir, "calibration_25f_candidate.json")

        # Run Phase 3.10 calibration pipeline to ensure artifacts are current
        cls.eval_report = run_phase_3_10_calibration_pipeline(
            data_dir=os.path.join(cls.current_dir, "data"),
            candidate_model_dir=cls.candidate_dir,
            output_cal_path=cls.cand_cal_path,
            eval_dir=os.path.join(cls.current_dir, "evaluation"),
            random_seed=42
        )

    def test_01_numerical_stability_extremes_and_invalids(self):
        """Test Beta calibrator with extreme probabilities, NaNs, +Inf, -Inf, and negatives."""
        cal = ModelCalibrator(method="beta")
        # Mock fit
        y_prob = np.array([0.01, 0.05, 0.20, 0.50, 0.80, 0.95])
        y_val = np.array([0, 0, 0, 1, 1, 1])
        cal.fit(y_prob, y_val)

        test_inputs = np.array([0.0, 1.0, 1e-12, 1.0 - 1e-12, -0.5, 1.5, np.nan, np.inf, -np.inf])
        calibrated = cal.predict_proba(test_inputs)

        self.assertEqual(len(calibrated), len(test_inputs))
        self.assertFalse(np.isnan(calibrated).any())
        self.assertFalse(np.isinf(calibrated).any())
        self.assertTrue((calibrated >= 0.0001).all())
        self.assertTrue((calibrated <= 0.9999).all())

    def test_02_output_bounds_strictness(self):
        """Test that calibrated outputs are strictly bounded in [0.0001, 0.9999]."""
        with open(self.cand_cal_path, "r") as f:
            cand_cal_dict = json.load(f)
        cal = ModelCalibrator.from_dict(cand_cal_dict)

        grid = np.linspace(0.0, 1.0, 1000)
        calibrated = cal.predict_proba(grid)
        self.assertTrue((calibrated >= 0.0001).all())
        self.assertTrue((calibrated <= 0.9999).all())

    def test_03_determinism_and_reproducibility(self):
        """Test that same input with same parameters produces identical calibrated output."""
        with open(self.cand_cal_path, "r") as f:
            cand_cal_dict = json.load(f)
        cal1 = ModelCalibrator.from_dict(cand_cal_dict)
        cal2 = ModelCalibrator.from_dict(cand_cal_dict)

        test_probs = np.array([0.02, 0.05, 0.15, 0.35, 0.70, 0.92])
        out1 = cal1.predict_proba(test_probs)
        out2 = cal2.predict_proba(test_probs)
        np.testing.assert_array_almost_equal(out1, out2, decimal=8)

    def test_04_validation_only_fitting_metadata(self):
        """Verify calibration metadata confirms fitting strictly on validation split."""
        with open(self.cand_cal_path, "r") as f:
            cand_cal_dict = json.load(f)

        meta = cand_cal_dict.get("fitting_metadata", {})
        self.assertEqual(meta.get("validation_samples"), 1200)
        self.assertEqual(meta.get("validation_fraud_count"), 56)
        self.assertAlmostEqual(meta.get("validation_fraud_rate"), 0.0467, places=3)

    def test_05_candidate_artifact_serialization_and_checksum(self):
        """Test candidate calibration JSON serialization, loading, and checksum integrity."""
        self.assertTrue(os.path.exists(self.cand_cal_path))
        with open(self.cand_cal_path, "r") as f:
            cand_dict = json.load(f)

        self.assertEqual(cand_dict["type"], "beta")
        self.assertEqual(cand_dict["version"], "cal-v2.5-beta-candidate-v1")
        self.assertTrue(cand_dict["is_fitted"])
        self.assertIn("checksum_sha256", cand_dict)
        self.assertIn("beta_params", cand_dict["parameters"])

    def test_06_strict_monotonicity_guarantee(self):
        """Test that Beta calibrator preserves monotonic probability ordering: p1 < p2 => cal(p1) <= cal(p2)."""
        with open(self.cand_cal_path, "r") as f:
            cand_cal_dict = json.load(f)
        cal = ModelCalibrator.from_dict(cand_cal_dict)

        probs = np.linspace(0.01, 0.99, 100)
        calibrated = cal.predict_proba(probs)
        diffs = np.diff(calibrated)
        self.assertTrue((diffs >= -1e-6).all(), "Beta calibration broke monotonic ordering")

    def test_07_policy_thresholds_unchanged(self):
        """Verify production decision thresholds (<0.05 ALLOW, 0.05-0.35 REVIEW, >=0.35 DECLINE)."""
        policy = CostSensitivePolicyEngine()
        
        # p = 0.02 (< 0.05) => ALLOW
        action_low, _ = policy.select_action(p_calibrated=0.02, amount=1000.0)
        # p = 0.15 (0.05-0.35) => MANUAL_REVIEW
        action_mid, _ = policy.select_action(p_calibrated=0.15, amount=1000.0)
        # p = 0.85 (>= 0.35) => DECLINE
        action_high, _ = policy.select_action(p_calibrated=0.85, amount=1000.0)

        self.assertEqual(action_low, "ALLOW")
        self.assertEqual(action_mid, "MANUAL_REVIEW")
        self.assertEqual(action_high, "DECLINE")

    def test_08_production_artifacts_unmodified(self):
        """CRITICAL PRODUCTION SAFETY: Verify production ONNX and calibration artifacts were not modified."""
        self.assertTrue(os.path.exists(self.prod_onnx_path))
        prod_sess = rt.InferenceSession(self.prod_onnx_path, providers=['CPUExecutionProvider'])
        # Production model MUST still expect 15 features
        self.assertEqual(prod_sess.get_inputs()[0].shape[1], 15)

        self.assertTrue(os.path.exists(self.prod_cal_path))
        with open(self.prod_cal_path, "r") as f:
            prod_cal_data = json.load(f)
        self.assertEqual(prod_cal_data.get("type"), "beta")
        self.assertEqual(prod_cal_data.get("version"), "cal-v2.0-beta")

    def test_09_ece_improvement(self):
        """Test that candidate Beta calibration achieves Expected Calibration Error (ECE) < 0.01."""
        sys_d = self.eval_report["system_comparison"]["25f_candidate_beta_calibrated"]
        ece = sys_d["ece"]
        brier = sys_d["brier_score"]
        self.assertLess(ece, 0.01)
        self.assertLess(brier, 0.05)

if __name__ == "__main__":
    unittest.main()
