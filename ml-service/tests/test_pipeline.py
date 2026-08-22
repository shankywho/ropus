"""
Unit and Integration Tests for Canonical Feature Pipeline and Point-in-Time Correctness
"""

import os
import sys
import unittest
import numpy as np
import pandas as pd

# Add ml-service root to path
ml_root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
if ml_root not in sys.path:
    sys.path.insert(0, ml_root)

from data_pipeline.schema import CANONICAL_FEATURE_COLS, RAW_SCHEMA
from data_pipeline.data_loader import load_raw_dataset
from data_pipeline.split import temporal_train_val_test_split
from data_pipeline.features import (
    extract_canonical_features,
    compute_point_in_time_velocities,
    compute_point_in_time_device_novelty,
    compute_point_in_time_amount_ratio
)
from data_pipeline.preprocess import CanonicalPreprocessor
from data_pipeline.validate import validate_pipeline_integrity

class TestCanonicalDataPipeline(unittest.TestCase):

    def setUp(self):
        self.sample_df, self.metadata = load_raw_dataset()

    def test_01_feature_schema_columns(self):
        """Test that canonical feature schema defines exact 15 features."""
        self.assertEqual(len(CANONICAL_FEATURE_COLS), 15)
        self.assertIn("amount", CANONICAL_FEATURE_COLS)
        self.assertIn("ip_velocity_1h", CANONICAL_FEATURE_COLS)
        self.assertIn("token_velocity_24h", CANONICAL_FEATURE_COLS)
        self.assertIn("device_seen_before", CANONICAL_FEATURE_COLS)
        self.assertIn("amount_to_mean_ratio", CANONICAL_FEATURE_COLS)

    def test_02_temporal_split_no_leakage(self):
        """Test that temporal splitting guarantees train < val < test time boundaries."""
        df_train, df_val, df_test, info = temporal_train_val_test_split(
            self.sample_df, time_col="TransactionDT", train_ratio=0.70, val_ratio=0.15, test_ratio=0.15
        )
        self.assertGreater(len(df_train), 0)
        self.assertGreater(len(df_val), 0)
        self.assertGreater(len(df_test), 0)
        
        max_train = df_train["TransactionDT"].max()
        min_val = df_val["TransactionDT"].min()
        max_val = df_val["TransactionDT"].max()
        min_test = df_test["TransactionDT"].min()
        
        self.assertLessEqual(max_train, min_val, "Train max time must be <= Val min time")
        self.assertLessEqual(max_val, min_test, "Val max time must be <= Test min time")

    def test_03_point_in_time_future_leakage_isolation(self):
        """
        CRITICAL TEST:
        Synthetic events T1 = 10:00 (36000s), T2 = 10:10 (36600s), T3 = 10:20 (37200s).
        A feature for T1 MUST NOT change whether T2 and T3 exist or do not exist in the dataset!
        """
        df_single = pd.DataFrame({
            "TransactionID": [101],
            "TransactionDT": [36000],
            "addr1": [100.0],
            "card1": [5000],
            "card2": [111.0],
            "TransactionAmt": [150.0],
            "ProductCD": ["W"],
            "card4": ["visa"],
            "card6": ["debit"],
            "P_emaildomain": ["gmail.com"],
            "dist1": [np.nan],
            "DeviceInfo": ["iOS Device"],
            "DeviceType": ["mobile"]
        })
        
        df_multiple = pd.DataFrame({
            "TransactionID": [101, 102, 103],
            "TransactionDT": [36000, 36600, 37200], # T1, T2, T3
            "addr1": [100.0, 100.0, 100.0],
            "card1": [5000, 5000, 5000],
            "card2": [111.0, 111.0, 111.0],
            "TransactionAmt": [150.0, 300.0, 450.0],
            "ProductCD": ["W", "W", "W"],
            "card4": ["visa", "visa", "visa"],
            "card6": ["debit", "debit", "debit"],
            "P_emaildomain": ["gmail.com", "gmail.com", "gmail.com"],
            "dist1": [np.nan, np.nan, np.nan],
            "DeviceInfo": ["iOS Device", "iOS Device", "iOS Device"],
            "DeviceType": ["mobile", "mobile", "mobile"]
        })
        
        feat_single = extract_canonical_features(df_single)
        feat_multiple = extract_canonical_features(df_multiple)
        
        # Check T1 features in both evaluations
        t1_from_single = feat_single.iloc[0]
        t1_from_multiple = feat_multiple.iloc[0]
        
        self.assertEqual(t1_from_single["ip_velocity_1h"], t1_from_multiple["ip_velocity_1h"])
        self.assertEqual(t1_from_single["token_velocity_24h"], t1_from_multiple["token_velocity_24h"])
        self.assertEqual(t1_from_single["device_seen_before"], t1_from_multiple["device_seen_before"])
        self.assertEqual(t1_from_single["amount_to_mean_ratio"], t1_from_multiple["amount_to_mean_ratio"])
        
        # In multiple, T2 should have token_velocity_24h = 2 and device_seen_before = 1
        self.assertEqual(feat_multiple.iloc[1]["token_velocity_24h"], 2.0)
        self.assertEqual(feat_multiple.iloc[1]["device_seen_before"], 1)
        # T3 should have token_velocity_24h = 3 and device_seen_before = 1
        self.assertEqual(feat_multiple.iloc[2]["token_velocity_24h"], 3.0)
        self.assertEqual(feat_multiple.iloc[2]["device_seen_before"], 1)

    def test_04_missing_value_handling_and_indicators(self):
        """Test that missing values generate explicit indicators and zero NaNs in preprocessed output."""
        df_feats = extract_canonical_features(self.sample_df)
        prep = CanonicalPreprocessor()
        prep.fit(df_feats)
        X = prep.transform(df_feats)
        
        self.assertEqual(X.isna().sum().sum(), 0, "Transformed feature matrix must contain 0 NaNs")
        self.assertIn("dist1_missing", X.columns)
        self.assertIn("device_info_missing", X.columns)
        self.assertTrue(set(X["dist1_missing"].unique()).issubset({0, 1}))

    def test_05_strict_train_fitted_target_encoding(self):
        """Test that domain risk mapping is fitted exclusively on train data."""
        df_train, df_val, df_test, _ = temporal_train_val_test_split(self.sample_df)
        train_feats = extract_canonical_features(df_train)
        test_feats = extract_canonical_features(df_test)
        
        prep = CanonicalPreprocessor()
        prep.fit(train_feats)
        
        # Test transformation produces valid numeric risk scores
        X_test = prep.transform(test_feats)
        self.assertTrue(all(X_test["email_domain_risk"] >= 0.0))
        self.assertTrue(all(X_test["email_domain_risk"] <= 1.0))

    def test_06_pipeline_validation_integrity(self):
        """Test full pipeline validation utility."""
        df_train, df_val, df_test, _ = temporal_train_val_test_split(self.sample_df)
        train_feats = extract_canonical_features(df_train)
        val_feats = extract_canonical_features(df_val)
        test_feats = extract_canonical_features(df_test)
        
        prep = CanonicalPreprocessor()
        prep.fit(train_feats)
        
        X_train = prep.transform(train_feats)
        X_val = prep.transform(val_feats)
        X_test = prep.transform(test_feats)
        
        is_valid, errors = validate_pipeline_integrity(
            X_train, X_val, X_test,
            df_train["TransactionDT"].values,
            df_val["TransactionDT"].values,
            df_test["TransactionDT"].values
        )
        self.assertTrue(is_valid, f"Pipeline validation failed: {errors}")
        self.assertEqual(len(errors), 0)

if __name__ == "__main__":
    unittest.main()
