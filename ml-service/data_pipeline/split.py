"""
Temporal Train / Validation / Test Splitting Module
Strictly Chronological Splitting with Zero Future Leakage
"""

import pandas as pd
from typing import Tuple, Dict, Any

def temporal_train_val_test_split(
    df: pd.DataFrame,
    time_col: str = "TransactionDT",
    train_ratio: float = 0.70,
    val_ratio: float = 0.15,
    test_ratio: float = 0.15
) -> Tuple[pd.DataFrame, pd.DataFrame, pd.DataFrame, Dict[str, Any]]:
    """
    Performs a strictly chronological split into train, validation, and test sets.
    The test set strictly represents future unseen transactions.
    """
    assert abs((train_ratio + val_ratio + test_ratio) - 1.0) < 1e-5, "Ratios must sum to 1.0"
    
    if time_col not in df.columns:
        raise ValueError(f"Time ordering column '{time_col}' not found in dataframe.")

    # Sort strictly by time column
    df_sorted = df.sort_values(by=time_col, ascending=True).reset_index(drop=True)
    n = len(df_sorted)

    n_train = int(n * train_ratio)
    n_val = int(n * val_ratio)

    df_train = df_sorted.iloc[:n_train].copy().reset_index(drop=True)
    df_val = df_sorted.iloc[n_train:n_train + n_val].copy().reset_index(drop=True)
    df_test = df_sorted.iloc[n_train + n_val:].copy().reset_index(drop=True)

    # Verify chronological separation
    max_train_t = df_train[time_col].max()
    min_val_t = df_val[time_col].min()
    max_val_t = df_val[time_col].max()
    min_test_t = df_test[time_col].min()

    assert max_train_t <= min_val_t, f"Leakage detected: max train timestamp ({max_train_t}) > min val timestamp ({min_val_t})"
    assert max_val_t <= min_test_t, f"Leakage detected: max val timestamp ({max_val_t}) > min test timestamp ({min_test_t})"

    split_info = {
        "strategy": "strict_chronological_temporal_split",
        "time_column": time_col,
        "train_rows": len(df_train),
        "val_rows": len(df_val),
        "test_rows": len(df_test),
        "train_time_range": [int(df_train[time_col].min()), int(max_train_t)],
        "val_time_range": [int(min_val_t), int(max_val_t)],
        "test_time_range": [int(min_test_t), int(df_test[time_col].max())],
        "train_fraud_rate": float(df_train["isFraud"].mean()) if "isFraud" in df_train.columns else None,
        "val_fraud_rate": float(df_val["isFraud"].mean()) if "isFraud" in df_val.columns else None,
        "test_fraud_rate": float(df_test["isFraud"].mean()) if "isFraud" in df_test.columns else None
    }

    return df_train, df_val, df_test, split_info
