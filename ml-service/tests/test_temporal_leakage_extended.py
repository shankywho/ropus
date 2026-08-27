"""
Tests for Extended Point-in-Time Temporal Feature Extraction (Zero-Leakage Invariant)
"""

import os
import sys
import pytest
import numpy as np
import pandas as pd

CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
ML_SERVICE_DIR = os.path.abspath(os.path.join(CURRENT_DIR, ".."))
if ML_SERVICE_DIR not in sys.path:
    sys.path.insert(0, ML_SERVICE_DIR)

from data_pipeline.temporal_features import (
    extract_point_in_time_temporal_features,
    ENHANCED_TEMPORAL_FEATURE_COLS
)

def test_temporal_feature_columns_exist():
    df = pd.DataFrame({
        "card1": [1001, 1001, 1002],
        "addr1": [325.0, 325.0, 126.0],
        "DeviceInfo": ["iOS", "iOS", "Windows"],
        "TransactionDT": [1000, 2000, 3000],
        "TransactionAmt": [50.0, 150.0, 20.0]
    })
    res = extract_point_in_time_temporal_features(df)
    assert len(res) == 3
    for col in ENHANCED_TEMPORAL_FEATURE_COLS:
        assert col in res.columns

def test_zero_future_leakage_on_modified_future_events():
    """
    Modifying a future transaction (at T=5000) must NOT change features for transaction at T=1000 or T=2000.
    """
    df1 = pd.DataFrame({
        "card1": [5000, 5000, 5000],
        "addr1": [100.0, 100.0, 100.0],
        "DeviceInfo": ["Android", "Android", "Android"],
        "TransactionDT": [1000, 2000, 5000],
        "TransactionAmt": [100.0, 200.0, 300.0]
    })

    df2 = pd.DataFrame({
        "card1": [5000, 5000, 5000],
        "addr1": [100.0, 100.0, 100.0],
        "DeviceInfo": ["Android", "Android", "Android"],
        "TransactionDT": [1000, 2000, 5000],
        "TransactionAmt": [100.0, 200.0, 999999.0] # Massive future amount at T=5000
    })

    res1 = extract_point_in_time_temporal_features(df1)
    res2 = extract_point_in_time_temporal_features(df2)

    # First transaction (T=1000) must be identical
    np.testing.assert_array_equal(res1.iloc[0].values, res2.iloc[0].values)
    # Second transaction (T=2000) must be identical despite massive future transaction at T=5000
    np.testing.assert_array_equal(res1.iloc[1].values, res2.iloc[1].values)

def test_sliding_window_expiration():
    """
    Events older than 24h (86,400s) must expire and not be counted in 24h count or amount.
    """
    df = pd.DataFrame({
        "card1": [7777, 7777],
        "addr1": [200.0, 200.0],
        "DeviceInfo": ["Mac", "Mac"],
        "TransactionDT": [1000, 1000 + 86401], # 24h and 1 second later
        "TransactionAmt": [500.0, 250.0]
    })
    res = extract_point_in_time_temporal_features(df)
    # Second transaction has no prior events in the last 24h window
    assert res.iloc[1]["card_tx_count_24h"] == 0.0
    assert res.iloc[1]["card_amount_sum_24h"] == 0.0
    # But 7d window still includes the first event
    assert res.iloc[1]["card_tx_count_7d"] == 1.0
    assert res.iloc[1]["card_amount_sum_7d"] == 500.0

def test_novelty_indicators():
    """
    First occurrence of a card/device must be marked as new (1.0), subsequent occurrences as not new (0.0).
    """
    df = pd.DataFrame({
        "card1": [8888, 8888, 9999],
        "addr1": [300.0, 300.0, 300.0],
        "DeviceInfo": ["DevA", "DevA", "DevB"],
        "TransactionDT": [100, 200, 300],
        "TransactionAmt": [10.0, 20.0, 30.0]
    })
    res = extract_point_in_time_temporal_features(df)
    assert res.iloc[0]["is_new_card_indicator"] == 1.0
    assert res.iloc[1]["is_new_card_indicator"] == 0.0
    assert res.iloc[2]["is_new_card_indicator"] == 1.0

    assert res.iloc[0]["is_new_device_indicator"] == 1.0
    assert res.iloc[1]["is_new_device_indicator"] == 0.0
    assert res.iloc[2]["is_new_device_indicator"] == 1.0
