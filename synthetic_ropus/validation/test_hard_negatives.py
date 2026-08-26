"""
Automated Pytest Suite: Hard Negatives & Employee-Consumer Realism
"""
import os
import pandas as pd
import pytest

DATA_DIR = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "data"))

def test_hard_negative_tags_exist():
    tags_path = os.path.join(DATA_DIR, "labels", "scenario_tags.csv")
    assert os.path.exists(tags_path), f"Missing scenario tags file: {tags_path}"
    df = pd.read_csv(tags_path)

    hard_negs = df[df["is_hard_negative"].astype(str).str.lower().isin(["true", "1"])]
    assert len(hard_negs) >= 1000, f"Expected at least 1000 hard negative tags, got {len(hard_negs)}"

    # Must include support staff, analyst reviews, household sharing, and pre-txn access
    scenario_types = set(hard_negs["scenario_type"])
    expected_hard_types = {"legitimate_high_volume_support", "legitimate_household_sharing", "legitimate_access_before_txn"}
    assert expected_hard_types.issubset(scenario_types), f"Missing expected hard negative scenario types. Present: {scenario_types}"

def test_legitimate_and_fraudulent_employee_edges_both_exist():
    emp_access_path = os.path.join(DATA_DIR, "edges", "employee_accesses_account.csv")
    tags_path = os.path.join(DATA_DIR, "labels", "scenario_tags.csv")

    emp_access = pd.read_csv(emp_access_path)
    tags = pd.read_csv(tags_path)

    emp_tags = tags[tags["entity_type"] == "EMPLOYEE"]
    collusion_emps = set(emp_tags[emp_tags["scenario_type"] == "employee_collusion"]["entity_id"])
    legit_emps = set(emp_tags[emp_tags["scenario_type"].str.startswith("legitimate_")]["entity_id"])

    active_emps = set(emp_access["employee_id"])

    assert len(collusion_emps.intersection(active_emps)) > 0, "No collusive employees found with active account access edges"
    assert len(legit_emps.intersection(active_emps)) > 0, "No legitimate employees found with active account access edges"

def test_hard_negative_evaluation_file():
    hn_file = os.path.join(DATA_DIR, "labels", "hard_negative_test.csv")
    assert os.path.exists(hn_file), f"Missing hard_negative_test.csv: {hn_file}"
    df = pd.read_csv(hn_file)
    assert len(df) > 0, "hard_negative_test.csv is empty"
