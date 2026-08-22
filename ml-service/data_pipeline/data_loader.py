"""
Data Loader Module for IEEE-CIS Fraud Detection Dataset & Fixtures
"""

import os
import pandas as pd
from typing import Tuple, Dict, Any

def load_raw_dataset(data_dir: str = None) -> Tuple[pd.DataFrame, Dict[str, Any]]:
    """
    Loads raw IEEE-CIS dataset from data_dir/raw if available,
    otherwise falls back to the deterministic sample fixture in data_dir/sample_ieee_fixture.csv.
    Ensures dataset is strictly ordered by TransactionDT chronologically.
    """
    if data_dir is None:
        data_dir = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "data")

    raw_dir = os.path.join(data_dir, "raw")
    train_txn_path = os.path.join(raw_dir, "train_transaction.csv")
    train_id_path = os.path.join(raw_dir, "train_identity.csv")
    fixture_path = os.path.join(data_dir, "sample_ieee_fixture.csv")

    metadata = {
        "dataset_source": "sample_ieee_fixture",
        "is_full_raw_dataset": False,
        "raw_paths_checked": [train_txn_path, train_id_path]
    }

    if os.path.exists(train_txn_path):
        print(f"Loading full IEEE-CIS transaction dataset from {train_txn_path}...")
        df_txn = pd.read_csv(train_txn_path)
        if os.path.exists(train_id_path):
            print(f"Merging IEEE-CIS identity table from {train_id_path}...")
            df_id = pd.read_csv(train_id_path)
            df = pd.merge(df_txn, df_id, on="TransactionID", how="left")
        else:
            df = df_txn
        metadata["dataset_source"] = "ieee_cis_raw_full"
        metadata["is_full_raw_dataset"] = True
    elif os.path.exists(fixture_path):
        print(f"Loading IEEE-CIS sample fixture from {fixture_path}...")
        df = pd.read_csv(fixture_path)
        metadata["dataset_source"] = "sample_ieee_fixture"
        metadata["is_full_raw_dataset"] = False
    else:
        # Generate deterministic fixture on-the-fly
        from data.generate_fixture import generate_sample_ieee_fixture
        print(f"Generating deterministic IEEE-CIS fixture at {fixture_path}...")
        df = generate_sample_ieee_fixture(output_path=fixture_path, n_rows=8000, random_seed=42)
        metadata["dataset_source"] = "sample_ieee_fixture_generated"
        metadata["is_full_raw_dataset"] = False

    # Ensure chronological sort by TransactionDT
    if "TransactionDT" in df.columns:
        df = df.sort_values(by="TransactionDT", ascending=True).reset_index(drop=True)

    metadata["total_rows"] = len(df)
    metadata["total_columns"] = len(df.columns)
    if "isFraud" in df.columns:
        fraud_count = int(df["isFraud"].sum())
        metadata["fraud_count"] = fraud_count
        metadata["non_fraud_count"] = int(len(df) - fraud_count)
        metadata["fraud_ratio"] = float(fraud_count / len(df))

    return df, metadata
