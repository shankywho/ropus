#!/usr/bin/env python3
"""
synthetic_ropus/generate.py

Generates a large, temporally-consistent, heterogeneous synthetic graph
dataset for ROPUS GraphSAGE development, fraud-ring / collusion / bot /
poisoning simulation, and leakage testing.

*** THIS DATA IS SYNTHETIC — DEVELOPMENT / ENGINEERING USE ONLY ***
It must never be used to claim production accuracy, recall, or real-world
fraud/collusion prevalence. The real ROPUS benchmark's 52 confirmed fraud
cases remain the only real evidence and are entirely separate from this
generator and its output.

Usage:
    python generate.py --events 100000 --seed 42
    python generate.py --events 500000 --seed 7 --outdir data_run2
"""
import argparse
import csv
import os
import random
import sys
from datetime import datetime

import yaml

from scenarios.common import IdCounters, build_entity_pools, GenResult, parse_date
from scenarios import (legitimate, employee_collusion, fraud_ring, account_takeover,
                        card_testing, bot_attack, sybil, graph_poisoning)

SCENARIO_MODULES = {
    "legitimate": legitimate,
    "employee_collusion": employee_collusion,
    "fraud_ring": fraud_ring,
    "account_takeover": account_takeover,
    "card_testing": card_testing,
    "bot_attack": bot_attack,
    "sybil": sybil,
    "graph_poisoning": graph_poisoning,
}

EDGE_FIELD_MAP = {
    "employee_accesses_account": ["employee_id", "account_id", "edge_timestamp"],
    "employee_reviews_case": ["employee_id", "case_id", "edge_timestamp"],
    "employee_modifies_transaction": ["employee_id", "transaction_id", "edge_timestamp"],
    "employee_approves_transaction": ["employee_id", "transaction_id", "edge_timestamp"],
    "employee_uses_device": ["employee_id", "device_id", "edge_timestamp"],
    "employee_uses_ip": ["employee_id", "ip_id", "edge_timestamp"],
    "consumer_owns_account": ["consumer_id", "account_id", "edge_timestamp"],
    "account_uses_device": ["account_id", "device_id", "edge_timestamp"],
    "account_uses_ip": ["account_id", "ip_id", "edge_timestamp"],
    "account_uses_payment_token": ["account_id", "payment_token_id", "edge_timestamp"],
    "account_transacts_with_merchant": ["account_id", "merchant_id", "edge_timestamp"],
    "case_references_account": ["case_id", "account_id", "edge_timestamp"],
    "case_references_consumer": ["case_id", "consumer_id", "edge_timestamp"],
    "case_references_employee": ["case_id", "employee_id", "edge_timestamp"],
}


def write_csv(path, rows, fieldnames):
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "w", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames)
        writer.writeheader()
        for r in rows:
            writer.writerow(r)


def main():
    ap = argparse.ArgumentParser(description="Generate synthetic ROPUS graph/event dataset.")
    ap.add_argument("--events", type=int, default=100000, help="target number of transaction/events")
    ap.add_argument("--seed", type=int, default=None, help="deterministic random seed")
    ap.add_argument("--outdir", type=str, default="data", help="output directory (relative to this script)")
    ap.add_argument("--config", type=str, default="config.yaml")
    args = ap.parse_args()

    base_dir = os.path.dirname(os.path.abspath(__file__))
    with open(os.path.join(base_dir, args.config)) as f:
        cfg = yaml.safe_load(f)

    seed = args.seed if args.seed is not None else cfg.get("random_seed_default", 42)
    rng = random.Random(seed)
    ids = IdCounters()

    start = parse_date(cfg["time_range"]["start_date"])
    end = parse_date(cfg["time_range"]["end_date"])

    print(f"[generate] seed={seed} target_events={args.events}")
    print("[generate] building entity pools...")
    pools = build_entity_pools(rng, cfg, args.events, ids)
    print(f"[generate] employees={len(pools.employees)} consumers={len(pools.consumers)} "
          f"accounts={len(pools.accounts)} devices={len(pools.devices)} ips={len(pools.ips)} "
          f"tokens={len(pools.payment_tokens)} merchants={len(pools.merchants)}")

    weights = cfg["scenario_weights"]
    total_weight = sum(weights.values())

    combined = GenResult()
    for name, module in SCENARIO_MODULES.items():
        w = weights.get(name, 0.0) / total_weight
        n_txn = max(1, int(args.events * w))
        print(f"[generate] running scenario '{name}' (~{n_txn} transactions)...")
        part = module.generate(rng, pools, ids, cfg, start, end, n_txn)
        combined.merge(part)

    print(f"[generate] total transactions generated: {len(combined.transactions)}")

    # sort transactions chronologically (temporal causality: nothing downstream
    # should assume arrival order == generation order)
    combined.transactions.sort(key=lambda t: t["event_timestamp"])

    outdir = os.path.join(base_dir, args.outdir)

    # ---------------- nodes ----------------
    write_csv(os.path.join(outdir, "nodes", "employee.csv"),
              [{"employee_id": e["id"], "role": e["role"]} for e in pools.employees],
              ["employee_id", "role"])
    write_csv(os.path.join(outdir, "nodes", "consumer.csv"),
              [{"consumer_id": c} for c in pools.consumers], ["consumer_id"])
    write_csv(os.path.join(outdir, "nodes", "account.csv"),
              [{"account_id": a, "consumer_id": c} for a, c in pools.accounts.items()],
              ["account_id", "consumer_id"])
    write_csv(os.path.join(outdir, "nodes", "device.csv"),
              [{"device_id": d["id"], "device_type": d["device_type"]} for d in pools.devices],
              ["device_id", "device_type"])
    write_csv(os.path.join(outdir, "nodes", "ip.csv"),
              [{"ip_id": i["id"], "ip_type": i["ip_type"]} for i in pools.ips],
              ["ip_id", "ip_type"])
    write_csv(os.path.join(outdir, "nodes", "payment_token.csv"),
              [{"payment_token_id": t} for t in pools.payment_tokens], ["payment_token_id"])
    write_csv(os.path.join(outdir, "nodes", "merchant.csv"),
              [{"merchant_id": m["id"], "category": m["category"]} for m in pools.merchants],
              ["merchant_id", "category"])
    write_csv(os.path.join(outdir, "nodes", "case.csv"), combined.nodes_cases,
              ["case_id", "opened_at", "status"])
    write_csv(os.path.join(outdir, "nodes", "session.csv"), combined.nodes_sessions,
              ["session_id", "account_id", "device_id", "ip_id", "started_at", "ended_at"])

    # ---------------- edges ----------------
    for edge_type, rows in combined.edges.items():
        fieldnames = EDGE_FIELD_MAP.get(edge_type)
        if fieldnames is None:
            fieldnames = sorted({k for r in rows for k in r.keys()})
        write_csv(os.path.join(outdir, "edges", f"{edge_type}.csv"), rows, fieldnames)

    # ---------------- transactions ----------------
    txn_fields = ["transaction_id", "account_id", "device_id", "ip_id", "payment_token_id",
                  "merchant_id", "amount", "currency", "txn_type", "event_timestamp", "ingestion_timestamp"]
    write_csv(os.path.join(outdir, "transactions", "transactions.csv"), combined.transactions, txn_fields)

    # ---------------- labels ----------------
    label_fields = ["label_id", "entity_type", "entity_id", "ground_truth",
                     "label_timestamp", "label_source", "label_confidence"]
    write_csv(os.path.join(outdir, "labels", "labels.csv"), combined.labels, label_fields)

    tag_fields = ["tag_id", "entity_type", "entity_id", "scenario_type", "is_adversarial", "is_hard_negative"]
    write_csv(os.path.join(outdir, "labels", "scenario_tags.csv"), combined.scenario_tags, tag_fields)

    # ---------------- temporal splits (strictly chronological, no shuffling) ----------------
    n = len(combined.transactions)
    train_frac = cfg["splits"]["train"]
    val_frac = cfg["splits"]["validation"]
    train_end = int(n * train_frac)
    val_end = int(n * (train_frac + val_frac))
    split_rows = []
    for i, t in enumerate(combined.transactions):
        if i < train_end:
            split = "train"
        elif i < val_end:
            split = "validation"
        else:
            split = "test"
        split_rows.append({"transaction_id": t["transaction_id"], "event_timestamp": t["event_timestamp"], "split": split})
    write_csv(os.path.join(outdir, "labels", "splits.csv"), split_rows,
              ["transaction_id", "event_timestamp", "split"])

    # ---------------- special evaluation subsets ----------------
    tag_by_type = {}
    for t in combined.scenario_tags:
        tag_by_type.setdefault(t["scenario_type"], []).append(t)

    def write_subset(name, scenario_types):
        rows = []
        for st in scenario_types:
            rows.extend(tag_by_type.get(st, []))
        write_csv(os.path.join(outdir, "labels", f"{name}.csv"), rows, tag_fields)

    write_subset("employee_collusion_test", ["employee_collusion"])
    write_subset("fraud_ring_test", ["fraud_ring", "sybil"])
    write_subset("bot_attack_test", ["bot_attack"])
    write_subset("graph_poisoning_test", ["graph_poisoning_flood", "graph_poisoning_injected_edge", "graph_poisoning_low_and_slow"])
    write_subset("hard_negative_test", [t for t in tag_by_type if t.startswith("legitimate_")])

    # temporal_leakage_test: a small, hand-checkable subset demonstrating
    # transactions whose label_timestamp is well after event_timestamp,
    # specifically for exercising the "no future info" validator.
    labels_by_txn = {}
    for lab in combined.labels:
        if lab["entity_type"] == "TRANSACTION":
            labels_by_txn.setdefault(lab["entity_id"], []).append(lab)
    leakage_rows = []
    for txn in combined.transactions[:2000]:
        for lab in labels_by_txn.get(txn["transaction_id"], []):
            event_dt = datetime.fromisoformat(txn["event_timestamp"])
            label_dt = datetime.fromisoformat(lab["label_timestamp"])
            if (label_dt - event_dt).total_seconds() > 86400:  # >1 day delay
                leakage_rows.append({
                    "transaction_id": txn["transaction_id"],
                    "event_timestamp": txn["event_timestamp"],
                    "label_timestamp": lab["label_timestamp"],
                    "ground_truth": lab["ground_truth"],
                    "note": "label must NOT be visible to a model evaluating at or before event_timestamp",
                })
    write_csv(os.path.join(outdir, "labels", "temporal_leakage_test.csv"), leakage_rows,
              ["transaction_id", "event_timestamp", "label_timestamp", "ground_truth", "note"])

    print(f"[generate] done. wrote output to {outdir}")
    print(f"[generate] splits: train={train_end} validation={val_end - train_end} test={n - val_end}")


if __name__ == "__main__":
    main()
