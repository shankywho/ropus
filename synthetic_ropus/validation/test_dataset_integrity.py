"""
Automated Pytest Suite: Dataset Referential Integrity, Schemas, and Privacy
"""
import os
import re
import glob
import pandas as pd
import pytest

DATA_DIR = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "data"))
ID_PATTERN = re.compile(r"^[a-z_]+_\d{6}$")
FORBIDDEN_COLUMNS = {"pan", "card_number", "cvv", "cvv2", "password", "ssn", "full_name", "email", "phone", "dob", "address"}
APPROVED_LABELS = {"LEGITIMATE", "FRAUD", "SUSPECTED_FRAUD", "EMPLOYEE_POLICY_VIOLATION",
                   "INTERNAL_COLLUSION", "ACCOUNT_TAKEOVER", "CARD_TESTING", "BOT_ATTACK", "FRAUD_RING"}

def test_node_tables_exist_and_not_empty():
    expected_nodes = ["employee", "consumer", "account", "device", "ip", "payment_token", "merchant", "case", "session"]
    for n in expected_nodes:
        path = os.path.join(DATA_DIR, "nodes", f"{n}.csv")
        assert os.path.exists(path), f"Missing node table: {path}"
        df = pd.read_csv(path)
        assert len(df) > 0, f"Node table {n}.csv is empty"

def test_edge_tables_exist_and_not_empty():
    expected_edges = [
        "account_transacts_with_merchant", "account_uses_device", "account_uses_ip",
        "account_uses_payment_token", "case_references_account", "case_references_consumer",
        "case_references_employee", "consumer_owns_account", "employee_accesses_account",
        "employee_approves_transaction", "employee_modifies_transaction", "employee_reviews_case",
        "employee_uses_device", "employee_uses_ip"
    ]
    for e in expected_edges:
        path = os.path.join(DATA_DIR, "edges", f"{e}.csv")
        assert os.path.exists(path), f"Missing edge table: {path}"
        df = pd.read_csv(path)
        assert len(df) > 0, f"Edge table {e}.csv is empty"

def test_transactions_table():
    path = os.path.join(DATA_DIR, "transactions", "transactions.csv")
    assert os.path.exists(path), f"Missing transactions file: {path}"
    df = pd.read_csv(path)
    assert len(df) >= 50000, f"Unexpectedly low transaction count: {len(df)}"
    required_cols = ["transaction_id", "account_id", "device_id", "ip_id", "payment_token_id", "merchant_id", "amount", "event_timestamp", "ingestion_timestamp"]
    for c in required_cols:
        assert c in df.columns, f"Missing required transaction column: {c}"

def test_synthetic_id_format():
    sample_files = [
        os.path.join(DATA_DIR, "nodes", "employee.csv"),
        os.path.join(DATA_DIR, "nodes", "account.csv"),
        os.path.join(DATA_DIR, "nodes", "consumer.csv"),
        os.path.join(DATA_DIR, "transactions", "transactions.csv")
    ]
    for path in sample_files:
        df = pd.read_csv(path)
        id_col = df.columns[0]
        invalid_ids = [val for val in df[id_col].dropna() if not ID_PATTERN.match(str(val))]
        assert len(invalid_ids) == 0, f"Found non-synthetic IDs in {path}: {invalid_ids[:5]}"

def test_no_forbidden_columns():
    all_csvs = glob.glob(os.path.join(DATA_DIR, "**", "*.csv"), recursive=True)
    for path in all_csvs:
        df = pd.read_csv(path, nrows=5)
        for col in df.columns:
            assert col.lower() not in FORBIDDEN_COLUMNS, f"Forbidden column '{col}' found in {path}"

def test_approved_ground_truth_labels():
    labels_path = os.path.join(DATA_DIR, "labels", "labels.csv")
    df = pd.read_csv(labels_path)
    unapproved = set(df["ground_truth"]) - APPROVED_LABELS
    assert len(unapproved) == 0, f"Found unapproved ground truth labels: {unapproved}"
