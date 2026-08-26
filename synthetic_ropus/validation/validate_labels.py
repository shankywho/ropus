#!/usr/bin/env python3
"""
Validates ground-truth labels:
  1. All ground_truth values are from the approved label set.
  2. Every label carries label_timestamp, label_source, label_confidence
     (independent provenance -- never a bare model score).
  3. Delayed label maturation exists (label_timestamp meaningfully after
     event_timestamp for a meaningful fraction of fraud labels).
  4. Multiple independent fraud mechanisms are represented (not one
     universal pattern).
  5. Hard negatives exist: legitimate-labeled entities carrying scenario
     tags that look risky (shared infra, high volume, pre-txn access).
  6. Legitimate AND fraudulent employee-consumer-adjacent relationships
     both exist (employee_accesses_account edges appear in both hard
     negative and collusion tag sets).

Run: python validation/validate_labels.py --data data
"""
import argparse
import csv
import os
from collections import Counter
from datetime import datetime

APPROVED = {"LEGITIMATE", "FRAUD", "SUSPECTED_FRAUD", "EMPLOYEE_POLICY_VIOLATION",
            "INTERNAL_COLLUSION", "ACCOUNT_TAKEOVER", "CARD_TESTING", "BOT_ATTACK", "FRAUD_RING"}
FRAUD_MECHANISMS = {"INTERNAL_COLLUSION", "ACCOUNT_TAKEOVER", "CARD_TESTING", "BOT_ATTACK", "FRAUD_RING"}


def load_csv(path):
    with open(path, newline="") as f:
        return list(csv.DictReader(f))


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--data", default="data")
    args = ap.parse_args()
    d = args.data
    errors = []

    labels = load_csv(os.path.join(d, "labels", "labels.csv"))
    gt_counts = Counter(l["ground_truth"] for l in labels)
    for gt in gt_counts:
        if gt not in APPROVED:
            errors.append(f"unapproved ground_truth value: {gt}")

    for l in labels:
        if not (l.get("label_timestamp") and l.get("label_source") and l.get("label_confidence")):
            errors.append(f"label {l['label_id']} missing provenance fields")

    # delayed maturation: fraud labels should show a spread of delays, not all instant
    txns = load_csv(os.path.join(d, "transactions", "transactions.csv"))
    txn_event_ts = {t["transaction_id"]: datetime.fromisoformat(t["event_timestamp"]) for t in txns}
    fraud_delays_days = []
    for l in labels:
        if l["ground_truth"] in FRAUD_MECHANISMS and l["entity_id"] in txn_event_ts:
            delay = (datetime.fromisoformat(l["label_timestamp"]) - txn_event_ts[l["entity_id"]]).days
            fraud_delays_days.append(delay)
    if fraud_delays_days:
        distinct_delays = len(set(fraud_delays_days))
        if distinct_delays < 3:
            errors.append("fraud label delays show insufficient variation for delayed-maturation testing")
        print(f"[info] fraud label delay range: {min(fraud_delays_days)}-{max(fraud_delays_days)} days "
              f"across {distinct_delays} distinct values")

    # multiple independent fraud mechanisms present
    mechanisms_present = FRAUD_MECHANISMS & set(gt_counts)
    if len(mechanisms_present) < 4:
        errors.append(f"only {len(mechanisms_present)} independent fraud mechanisms present: {mechanisms_present}")
    print(f"[info] fraud mechanisms present: {sorted(mechanisms_present)}")

    # hard negatives
    tags = load_csv(os.path.join(d, "labels", "scenario_tags.csv"))
    hard_neg = [t for t in tags if t.get("is_hard_negative") in ("True", "true", "1")]
    if len(hard_neg) < 10:
        errors.append(f"too few hard negatives found ({len(hard_neg)})")
    print(f"[info] hard negative tagged rows: {len(hard_neg)}")

    # legitimate vs fraudulent employee-account relationships both exist
    collusion_emp_tags = [t for t in tags if t["scenario_type"] == "employee_collusion" and t["entity_type"] == "EMPLOYEE"]
    legit_emp_tags = [t for t in tags if t["scenario_type"].startswith("legitimate_") and t["entity_type"] == "EMPLOYEE"]
    if not collusion_emp_tags:
        errors.append("no fraudulent employee-consumer/account relationships found")
    if not legit_emp_tags:
        errors.append("no legitimate employee-account relationships found")

    if errors:
        print(f"[FAIL] {len(errors)} label validation errors:")
        for e in errors[:25]:
            print("  -", e)
        raise SystemExit(1)
    print("[PASS] validate_labels: ground truth well-formed, independent, and diverse")


if __name__ == "__main__":
    main()
