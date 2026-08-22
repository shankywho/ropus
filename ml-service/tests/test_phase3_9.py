"""
Phase 3.9 Test Suite — Candidate 25-Feature Model, Leakage Prevention, and ONNX Parity
"""

import os
import json
import unittest
import numpy as np
import pandas as pd
import joblib
import onnxruntime as rt
import xgboost as xgb

from data_pipeline.schema import CANONICAL_25_FEATURE_COLS, CANONICAL_15_FEATURE_COLS
from data_pipeline.data_loader import load_raw_dataset
from data_pipeline.split import temporal_train_val_test_split
from data_pipeline.features import extract_canonical_25_features
from data_pipeline.preprocess import CanonicalPreprocessor
from data_pipeline.validate import validate_pipeline_integrity
from train_25f import train_and_evaluate_25f_candidate

class TestPhase39CandidateRetraining(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.current_dir = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
        cls.candidate_dir = os.path.join(cls.current_dir, "model", "candidates")
        cls.prod_onnx_path = os.path.join(cls.current_dir, "model", "fraud_model.onnx")
        cls.prod_cal_path = os.path.join(cls.current_dir, "model", "calibration.json")

        # Run candidate training pipeline to ensure fresh candidate artifacts
        cls.metadata = train_and_evaluate_25f_candidate(
            data_dir=os.path.join(cls.current_dir, "data"),
            output_dir=cls.candidate_dir,
            eval_dir=os.path.join(cls.current_dir, "evaluation"),
            random_seed=42
        )

    def test_01_canonical_25_feature_count_and_ordering(self):
        """Test that candidate feature schema defines exactly 25 features in canonical order."""
        self.assertEqual(len(CANONICAL_25_FEATURE_COLS), 25)
        # Check first 15 match legacy exactly
        self.assertEqual(CANONICAL_25_FEATURE_COLS[:15], CANONICAL_15_FEATURE_COLS)
        # Check 10 new features
        expected_new = [
            "device_tx_count_5m",
            "device_tx_count_1h",
            "device_amount_sum_24h",
            "tx_acceleration_5m_1h",
            "device_amount_concentration_5m_1h",
            "device_unique_tokens_1h",
            "token_unique_devices_1h",
            "device_reputation_score",
            "device_fraud_rate",
            "device_dispute_rate"
        ]
        self.assertEqual(CANONICAL_25_FEATURE_COLS[15:], expected_new)

    def test_02_point_in_time_leakage_isolation(self):
        """Test that event at time T cannot see future transactions or its own outcome."""
        test_data = pd.DataFrame({
            "TransactionID": [101, 102, 103],
            "TransactionDT": [1000, 1100, 1200],
            "TransactionAmt": [100.0, 200.0, 300.0],
            "ProductCD": ["W", "W", "W"],
            "card1": [1234, 1234, 1234],
            "card2": [567.0, 567.0, 567.0],
            "card4": ["visa", "visa", "visa"],
            "card6": ["debit", "debit", "debit"],
            "addr1": [300.0, 300.0, 300.0],
            "dist1": [10.0, 10.0, 10.0],
            "P_emaildomain": ["gmail.com", "gmail.com", "gmail.com"],
            "DeviceInfo": ["iOS_Device", "iOS_Device", "iOS_Device"],
            "DeviceType": ["mobile", "mobile", "mobile"],
            "isFraud": [1, 0, 0]  # First tx is fraud
        })

        feat_df = extract_canonical_25_features(test_data)
        self.assertEqual(len(feat_df), 3)

        # Tx 1: Point-in-time reputation is 0.50 (neutral), fraud_rate is 0.0 (past fraud=0)
        self.assertEqual(feat_df["device_reputation_score"].iloc[0], 0.50)
        self.assertEqual(feat_df["device_fraud_rate"].iloc[0], 0.0)
        self.assertEqual(feat_df["device_tx_count_5m"].iloc[0], 0.0)

        # Tx 2: Point-in-time reputation reflects Tx 1's fraud
        self.assertGreater(feat_df["device_reputation_score"].iloc[1], 0.50)
        self.assertEqual(feat_df["device_fraud_rate"].iloc[1], 1.0)
        self.assertEqual(feat_df["device_tx_count_5m"].iloc[1], 1.0)

    def test_03_temporal_split_integrity(self):
        """Test that train < val < test strictly holds chronologically."""
        df_raw, _ = load_raw_dataset(data_dir=os.path.join(self.current_dir, "data"))
        df_train, df_val, df_test, split_info = temporal_train_val_test_split(df_raw)

        max_train = df_train["TransactionDT"].max()
        min_val = df_val["TransactionDT"].min()
        max_val = df_val["TransactionDT"].max()
        min_test = df_test["TransactionDT"].min()

        self.assertLessEqual(max_train, min_val)
        self.assertLessEqual(max_val, min_test)

    def test_04_preprocessor_zero_nans_25f(self):
        """Test that CanonicalPreprocessor transforms to exactly 25 columns with zero NaNs."""
        df_raw, _ = load_raw_dataset(data_dir=os.path.join(self.current_dir, "data"))
        feat_df = extract_canonical_25_features(df_raw)
        prep = CanonicalPreprocessor(feature_contract="v2.5").fit(feat_df)
        X_trans = prep.transform(feat_df, feature_contract="v2.5")

        self.assertEqual(X_trans.shape[1], 25)
        self.assertEqual(list(X_trans.columns), CANONICAL_25_FEATURE_COLS)
        self.assertEqual(int(X_trans.isna().sum().sum()), 0)

    def test_05_candidate_joblib_model_loading_and_prediction(self):
        """Test that candidate joblib model bundle loads and predicts probabilities in [0, 1]."""
        joblib_path = os.path.join(self.candidate_dir, "fraud_model_25f_candidate.joblib")
        self.assertTrue(os.path.exists(joblib_path))

        bundle = joblib.load(joblib_path)
        model = bundle["model"]
        self.assertIsInstance(model, xgb.XGBClassifier)

        sample = np.ones((5, 25), dtype=np.float32)
        preds = model.predict_proba(sample)[:, 1]
        self.assertEqual(len(preds), 5)
        self.assertTrue((preds >= 0.0).all() and (preds <= 1.0).all())

    def test_06_candidate_onnx_model_loading_and_input_dim(self):
        """Test that candidate ONNX model loads with input dimension 25."""
        onnx_path = os.path.join(self.candidate_dir, "fraud_model_25f_candidate.onnx")
        self.assertTrue(os.path.exists(onnx_path))

        sess = rt.InferenceSession(onnx_path, providers=['CPUExecutionProvider'])
        input_shape = sess.get_inputs()[0].shape
        self.assertEqual(input_shape[1], 25)

    def test_07_native_xgboost_onnx_parity(self):
        """Test that native XGBoost and ONNX produce identical probabilities (< 1e-4 difference)."""
        joblib_path = os.path.join(self.candidate_dir, "fraud_model_25f_candidate.joblib")
        onnx_path = os.path.join(self.candidate_dir, "fraud_model_25f_candidate.onnx")

        bundle = joblib.load(joblib_path)
        xgb_model = bundle["model"]
        sess = rt.InferenceSession(onnx_path, providers=['CPUExecutionProvider'])

        np.random.seed(42)
        sample = np.random.uniform(0.0, 10.0, size=(20, 25)).astype(np.float32)

        xgb_probs = xgb_model.predict_proba(sample)[:, 1]
        onnx_out = sess.run(None, {sess.get_inputs()[0].name: sample})
        
        prob_output = onnx_out[1]
        if isinstance(prob_output, list) and len(prob_output) > 0 and isinstance(prob_output[0], dict):
            onnx_probs = np.array([float(d.get(1, 0.0)) for d in prob_output])
        elif isinstance(prob_output, np.ndarray) and prob_output.ndim >= 2:
            onnx_probs = onnx_out[1][:, 1]
        else:
            onnx_probs = onnx_out[0].flatten()

        max_diff = float(np.max(np.abs(xgb_probs - onnx_probs)))
        self.assertLess(max_diff, 1e-4)

    def test_08_production_artifacts_unmodified(self):
        """CRITICAL PRODUCTION SAFETY: Verify production ONNX and calibration artifacts were not overwritten."""
        self.assertTrue(os.path.exists(cls_prod := self.prod_onnx_path))
        prod_sess = rt.InferenceSession(cls_prod, providers=['CPUExecutionProvider'])
        # Production model MUST still expect 15 features
        self.assertEqual(prod_sess.get_inputs()[0].shape[1], 15)

        self.assertTrue(os.path.exists(self.prod_cal_path))
        with open(self.prod_cal_path, "r") as f:
            cal_data = json.load(f)
        self.assertEqual(cal_data.get("type"), "beta")

    def test_09_candidate_metadata_correctness(self):
        """Test candidate metadata.json contains all required schema fields."""
        meta_path = os.path.join(self.candidate_dir, "metadata.json")
        self.assertTrue(os.path.exists(meta_path))

        with open(meta_path, "r") as f:
            meta = json.load(f)

        self.assertEqual(meta["model_version"], "fraud-xgb-25f-candidate-v1")
        self.assertEqual(meta["feature_contract"], "v2.5")
        self.assertEqual(meta["feature_count"], 25)
        self.assertFalse(meta["is_production_active"])
        self.assertIn("test_metrics_25f", meta)
        self.assertIn("test_metrics_15f", meta)
        self.assertIn("ablation_delta", meta)
        self.assertIn("feature_importance_ranking", meta)

if __name__ == "__main__":
    unittest.main()
