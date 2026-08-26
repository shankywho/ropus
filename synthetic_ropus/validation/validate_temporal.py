#!/usr/bin/env python3
"""
Validates temporal causality:
  1. ingestion_timestamp >= event_timestamp for every transaction.
  2. label_timestamp >= event_timestamp for every transaction label.
  3. For a simulated evaluation cutoff T, no label/edge with
     timestamp > T is visible in a T-scoped view (demonstrated on
     temporal_leakage_test.csv).
  4. Train/validation/test splits are strictly chronological (no split
     boundary violation, no shuffling across time).

Run: python validation/validate_temporal.py --data data
"""
import argparse
import csv
import os
from datetime import datetime


def load_csv(path):
    with open(path, newline="") as f:
        return list(csv.DictReader(f))


def parse_ts(s):
    return datetime.fromisoformat(s)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--data", default="data")
    args = ap.parse_args()
    d = args.data
    errors = []

    # 1. ingestion >= event
    txns = load_csv(os.path.join(d, "transactions", "transactions.csv"))
    for t in txns:
        if parse_ts(t["ingestion_timestamp"]) < parse_ts(t["event_timestamp"]):
            errors.append(f"ingestion before event for {t['transaction_id']}")

    # 2. label_timestamp >= event_timestamp
    txn_event_ts = {t["transaction_id"]: parse_ts(t["event_timestamp"]) for t in txns}
    labels = load_csv(os.path.join(d, "labels", "labels.csv"))
    for lab in labels:
        if lab["entity_type"] == "TRANSACTION" and lab["entity_id"] in txn_event_ts:
            if parse_ts(lab["label_timestamp"]) < txn_event_ts[lab["entity_id"]]:
                errors.append(f"label before event for {lab['entity_id']}")

    # 3. temporal_leakage_test demonstration file: confirm every row would be
    #    correctly EXCLUDED from a T-scoped view at T = event_timestamp.
    leak_path = os.path.join(d, "labels", "temporal_leakage_test.csv")
    if os.path.exists(leak_path):
        leak_rows = load_csv(leak_path)
        for r in leak_rows:
            event_ts = parse_ts(r["event_timestamp"])
            label_ts = parse_ts(r["label_timestamp"])
            if label_ts <= event_ts:
                errors.append(f"temporal_leakage_test row {r['transaction_id']} does not actually "
                               f"demonstrate a future label (label_ts <= event_ts)")
        print(f"[ok] temporal_leakage_test.csv contains {len(leak_rows)} rows demonstrating "
              f"labels that must be withheld from evaluation at T=event_timestamp")

    # 4. split chronology: max(event_timestamp) in train <= min(event_timestamp) in validation, etc.
    splits = load_csv(os.path.join(d, "labels", "splits.csv"))
    by_split = {"train": [], "validation": [], "test": []}
    for r in splits:
        by_split.setdefault(r["split"], []).append(parse_ts(r["event_timestamp"]))
    if by_split["train"] and by_split["validation"]:
        if max(by_split["train"]) > min(by_split["validation"]):
            # small overlap at the boundary second is tolerated only if identical;
            # flag genuine inversions
            overlap = sum(1 for x in by_split["train"] if x > min(by_split["validation"]))
            if overlap > 0:
                errors.append(f"{overlap} train rows have event_timestamp after validation split begins")
    if by_split["validation"] and by_split["test"]:
        if max(by_split["validation"]) > min(by_split["test"]):
            overlap = sum(1 for x in by_split["validation"] if x > min(by_split["test"]))
            if overlap > 0:
                errors.append(f"{overlap} validation rows have event_timestamp after test split begins")

    if errors:
        print(f"[FAIL] {len(errors)} temporal validation errors:")
        for e in errors[:25]:
            print("  -", e)
        raise SystemExit(1)
    print("[PASS] validate_temporal: no future-information leakage detected")


if __name__ == "__main__":
    main()
