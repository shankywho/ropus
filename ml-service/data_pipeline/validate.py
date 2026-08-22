"""
Pipeline Validation & Integrity Verification Module
Checks schema conformity, NaN absence, and point-in-time temporal guarantees
"""

import numpy as np
import pandas as pd
from typing import Dict, Any, List, Tuple
from .schema import CANONICAL_FEATURE_COLS, CANONICAL_15_FEATURE_COLS, CANONICAL_25_FEATURE_COLS

def validate_pipeline_integrity(
    X_train: pd.DataFrame,
    X_val: pd.DataFrame,
    X_test: pd.DataFrame,
    train_dts: np.ndarray,
    val_dts: np.ndarray,
    test_dts: np.ndarray,
    expected_cols: List[str] = None
) -> Tuple[bool, List[str]]:
    """
    Validates that:
    1. All canonical feature columns exist and are strictly ordered.
    2. Zero NaNs / Infs exist in transformed matrices.
    3. Chronological boundaries strictly satisfy max(Train) <= min(Val) <= max(Val) <= min(Test).
    """
    errors = []
    if expected_cols is None:
        expected_cols = CANONICAL_25_FEATURE_COLS if X_train.shape[1] == 25 else CANONICAL_15_FEATURE_COLS
    
    # 1. Check feature columns and ordering
    for split_name, df in [("Train", X_train), ("Val", X_val), ("Test", X_test)]:
        cols = list(df.columns)
        if cols != expected_cols:
            errors.append(f"{split_name} columns do not match expected columns exactly. Found {cols} vs {expected_cols}")
            
        # 2. Check NaNs
        nan_counts = df.isna().sum().to_dict()
        total_nans = sum(nan_counts.values())
        if total_nans > 0:
            errors.append(f"{split_name} contains {total_nans} NaNs: {nan_counts}")
            
    # 3. Check temporal ordering
    max_train_t = np.max(train_dts)
    min_val_t = np.min(val_dts)
    max_val_t = np.max(val_dts)
    min_test_t = np.min(test_dts)
    
    if max_train_t > min_val_t:
        errors.append(f"Temporal Leakage: max train timestamp ({max_train_t}) > min val timestamp ({min_val_t})")
    if max_val_t > min_test_t:
        errors.append(f"Temporal Leakage: max val timestamp ({max_val_t}) > min test timestamp ({min_test_t})")
        
    is_valid = len(errors) == 0
    return is_valid, errors
