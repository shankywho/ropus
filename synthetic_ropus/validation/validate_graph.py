#!/usr/bin/env python3
"""
Validates graph structure:
  1. No duplicate node IDs within each node file.
  2. No orphaned required entity references in edges/transactions
     (every referenced id exists in its node file).
  3. No impossible timestamps (event_timestamp within configured date range).
  4. Graph degree distributions are bounded (flags any node whose degree
     is a wild outlier vs. the rest of the population, which is expected
     ONLY for the deliberately-injected graph_poisoning flood node).

Run: python validation/validate_graph.py --data data
"""
import argparse
import csv
import glob
import os
from collections import Counter
from datetime import datetime


def load_csv(path):
    with open(path, newline="") as f:
        return list(csv.DictReader(f))


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--data", default="data")
    ap.add_argument("--start-date", default="2025-01-01")
    ap.add_argument("--end-date", default="2026-01-01")
    args = ap.parse_args()
    d = args.data
    errors = []
    warnings = []

    range_start = datetime.strptime(args.start_date, "%Y-%m-%d")
    range_end = datetime.strptime(args.end_date, "%Y-%m-%d")

    node_ids = {}
    id_cols = {
        "employee.csv": "employee_id", "consumer.csv": "consumer_id", "account.csv": "account_id",
        "device.csv": "device_id", "ip.csv": "ip_id", "payment_token.csv": "payment_token_id",
        "merchant.csv": "merchant_id", "case.csv": "case_id", "session.csv": "session_id",
    }
    for fname, col in id_cols.items():
        path = os.path.join(d, "nodes", fname)
        if not os.path.exists(path):
            continue
        rows = load_csv(path)
        ids_list = [r[col] for r in rows]
        dupes = [k for k, c in Counter(ids_list).items() if c > 1]
        if dupes:
            errors.append(f"{fname}: {len(dupes)} duplicate ids, e.g. {dupes[:3]}")
        node_ids[col] = set(ids_list)

    # transactions reference valid accounts/devices/ips/tokens/merchants; timestamps in range
    txns = load_csv(os.path.join(d, "transactions", "transactions.csv"))
    degree = Counter()
    for t in txns:
        for col, key in [("account_id", "account_id"), ("device_id", "device_id"),
                          ("ip_id", "ip_id"), ("payment_token_id", "payment_token_id"),
                          ("merchant_id", "merchant_id")]:
            if key in node_ids and t[col] not in node_ids[key]:
                errors.append(f"transaction {t['transaction_id']} references unknown {col}={t[col]}")
        degree[t["device_id"]] += 1
        ts = datetime.fromisoformat(t["event_timestamp"])
        if not (range_start <= ts <= range_end + __import__("datetime").timedelta(days=1)):
            errors.append(f"transaction {t['transaction_id']} event_timestamp {ts} outside configured range")

    # edges reference valid node ids
    edge_ref_cols = {
        "employee_id": "employee_id", "account_id": "account_id", "case_id": "case_id",
        "transaction_id": None,  # validated via txn ids separately if needed
        "device_id": "device_id", "ip_id": "ip_id", "consumer_id": "consumer_id",
        "payment_token_id": "payment_token_id", "merchant_id": "merchant_id",
    }
    txn_ids = {t["transaction_id"] for t in txns}
    for path in glob.glob(os.path.join(d, "edges", "*.csv")):
        rows = load_csv(path)
        for r in rows:
            for col, key in edge_ref_cols.items():
                if col in r and r[col]:
                    if col == "transaction_id":
                        if r[col] not in txn_ids:
                            errors.append(f"{os.path.basename(path)}: unknown transaction_id {r[col]}")
                    elif key and key in node_ids and r[col] not in node_ids[key]:
                        errors.append(f"{os.path.basename(path)}: unknown {col}={r[col]}")

    # degree distribution sanity: flag extreme outliers (informational unless
    # absurd), since one deliberate flood node from graph_poisoning is expected.
    if degree:
        values = sorted(degree.values())
        median = values[len(values) // 2]
        p99 = values[int(len(values) * 0.99)]
        extreme = [k for k, v in degree.items() if v > max(50, p99 * 5)]
        if extreme:
            warnings.append(f"{len(extreme)} device(s) with extreme transaction fan-out "
                             f"(median={median}, p99={p99}) -- expected for graph_poisoning flood node(s)")

    for w in warnings:
        print("[warn]", w)

    if errors:
        print(f"[FAIL] {len(errors)} graph validation errors:")
        for e in errors[:25]:
            print("  -", e)
        raise SystemExit(1)
    print("[PASS] validate_graph: no orphaned references, no duplicate ids, timestamps in range")


if __name__ == "__main__":
    main()
