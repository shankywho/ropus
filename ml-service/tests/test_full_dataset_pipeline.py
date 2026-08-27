"""
Unit Tests for Full-Dataset Ingestion & Preflight Validation Pipeline
Tests: Missing raw file exception, SchemaValidationError, Chronological partitioning.
"""

import os
import sys
import tempfile
import pytest
import pandas as pd

CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
ML_SERVICE_DIR = os.path.abspath(os.path.join(CURRENT_DIR, ".."))
if ML_SERVICE_DIR not in sys.path:
    sys.path.insert(0, ML_SERVICE_DIR)

from data_pipeline.full_dataset_pipeline import (
    FullDatasetIngestionPipeline,
    FullDatasetPreflightValidator,
    DatasetUnavailableError,
    SchemaValidationError
)

def test_missing_raw_files_raises_dataset_unavailable_error():
    """Asserts that missing raw files raise DatasetUnavailableError with FULL_DATASET_UNAVAILABLE."""
    with tempfile.TemporaryDirectory() as tmp_dir:
        pipeline = FullDatasetIngestionPipeline(raw_data_dir=tmp_dir)
        with pytest.raises(DatasetUnavailableError) as exc_info:
            pipeline.load_full_dataset_strictly()
        assert "FULL_DATASET_UNAVAILABLE" in str(exc_info.value)

def test_missing_required_column_raises_schema_validation_error():
    """Asserts that missing mandatory columns in raw files raise SchemaValidationError."""
    with tempfile.TemporaryDirectory() as tmp_dir:
        txn_path = os.path.join(tmp_dir, "train_transaction.csv")
        id_path = os.path.join(tmp_dir, "train_identity.csv")

        # Write invalid header (missing ProductCD and isFraud)
        pd.DataFrame({"TransactionID": [101], "TransactionAmt": [50.0]}).to_csv(txn_path, index=False)
        pd.DataFrame({"TransactionID": [101], "id_01": [-5.0], "DeviceType": ["desktop"], "DeviceInfo": ["MacOS"]}).to_csv(id_path, index=False)

        with pytest.raises(SchemaValidationError) as exc_info:
            FullDatasetPreflightValidator.validate_preflight(txn_path, id_path)
        assert "missing required columns" in str(exc_info.value)

def test_valid_preflight_and_chronological_split():
    """Asserts that valid raw input files load and split strictly chronologically."""
    with tempfile.TemporaryDirectory() as tmp_dir:
        txn_path = os.path.join(tmp_dir, "train_transaction.csv")
        id_path = os.path.join(tmp_dir, "train_identity.csv")

        # Create minimal valid synthetic mock
        n_rows = 100
        df_txn = pd.DataFrame({
            "TransactionID": range(1000, 1000 + n_rows),
            "isFraud": [1 if i % 10 == 0 else 0 for i in range(n_rows)],
            "TransactionDT": [i * 100 for i in range(n_rows)],
            "TransactionAmt": [50.0 + i for i in range(n_rows)],
            "ProductCD": ["W"] * n_rows,
            "card1": [1001] * n_rows, "card2": [200] * n_rows, "card3": [150] * n_rows,
            "card4": ["visa"] * n_rows, "card5": [226] * n_rows, "card6": ["debit"] * n_rows,
            "addr1": [300] * n_rows, "addr2": [87] * n_rows,
            "P_emaildomain": ["gmail.com"] * n_rows, "R_emaildomain": ["gmail.com"] * n_rows
        })
        df_id = pd.DataFrame({
            "TransactionID": range(1000, 1000 + n_rows),
            "id_01": [-5.0] * n_rows,
            "DeviceType": ["desktop"] * n_rows,
            "DeviceInfo": ["MacOS"] * n_rows
        })
        df_txn.to_csv(txn_path, index=False)
        df_id.to_csv(id_path, index=False)

        pipeline = FullDatasetIngestionPipeline(raw_data_dir=tmp_dir)
        df_tr, df_va, df_te, manifest = pipeline.load_full_dataset_strictly(0.70, 0.15, 0.15)

        assert len(df_tr) == 70
        assert len(df_va) == 15
        assert len(df_te) == 15
        # Verify strict time monotonicity across splits
        assert df_tr["TransactionDT"].max() < df_va["TransactionDT"].min()
        assert df_va["TransactionDT"].max() < df_te["TransactionDT"].min()
