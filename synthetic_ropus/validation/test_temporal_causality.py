"""
Automated Pytest Suite: Temporal Causality & Leakage Prevention
"""
import os
import glob
import pandas as pd
from datetime import datetime
import pytest

DATA_DIR = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "data"))

def test_ingestion_after_event():
    txns_path = os.path.join(DATA_DIR, "transactions", "transactions.csv")
    assert os.path.exists(txns_path), f"Missing transactions file: {txns_path}"
    df = pd.read_csv(txns_path)

    event_ts = pd.to_datetime(df["event_timestamp"])
    ingest_ts = pd.to_datetime(df["ingestion_timestamp"])

    violations = (ingest_ts < event_ts).sum()
    assert violations == 0, f"Found {violations} transactions where ingestion < event timestamp"

def test_label_after_event():
    txns_path = os.path.join(DATA_DIR, "transactions", "transactions.csv")
    labels_path = os.path.join(DATA_DIR, "labels", "labels.csv")
    assert os.path.exists(labels_path), f"Missing labels file: {labels_path}"

    txns = pd.read_csv(txns_path)
    labels = pd.read_csv(labels_path)

    txn_labels = labels[labels["entity_type"] == "TRANSACTION"]
    merged = pd.merge(txn_labels, txns[["transaction_id", "event_timestamp"]], left_on="entity_id", right_on="transaction_id")

    event_ts = pd.to_datetime(merged["event_timestamp"])
    label_ts = pd.to_datetime(merged["label_timestamp"])

    violations = (label_ts < event_ts).sum()
    assert violations == 0, f"Found {violations} transaction labels where label_timestamp < event_timestamp"

def test_split_chronology():
    splits_path = os.path.join(DATA_DIR, "labels", "splits.csv")
    assert os.path.exists(splits_path), f"Missing splits file: {splits_path}"
    splits = pd.read_csv(splits_path)

    train_ts = pd.to_datetime(splits[splits["split"] == "train"]["event_timestamp"])
    val_ts = pd.to_datetime(splits[splits["split"] == "validation"]["event_timestamp"])
    test_ts = pd.to_datetime(splits[splits["split"] == "test"]["event_timestamp"])

    assert train_ts.max() <= val_ts.min(), f"Train max ({train_ts.max()}) > Val min ({val_ts.min()})"
    assert val_ts.max() <= test_ts.min(), f"Val max ({val_ts.max()}) > Test min ({test_ts.min()})"

def test_temporal_leakage_demonstration_file():
    leak_path = os.path.join(DATA_DIR, "labels", "temporal_leakage_test.csv")
    assert os.path.exists(leak_path), f"Missing temporal_leakage_test.csv: {leak_path}"
    df = pd.read_csv(leak_path)
    assert len(df) > 0, "temporal_leakage_test.csv is empty"

    event_ts = pd.to_datetime(df["event_timestamp"])
    label_ts = pd.to_datetime(df["label_timestamp"])

    # In temporal leakage test, every label should be strictly after event timestamp
    assert (label_ts > event_ts).all(), "All rows in temporal_leakage_test must have label_timestamp > event_timestamp"
