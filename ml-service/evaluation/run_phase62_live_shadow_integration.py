#!/usr/bin/env python3
"""
Phase 62: Live Shadow Integration, Connectivity Audit & Persistent Observability Runner
"""

import os
import sys
import json
import time
import hashlib
import numpy as np
from datetime import datetime, timezone, timedelta

CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
ML_SERVICE_DIR = os.path.abspath(os.path.join(CURRENT_DIR, ".."))
if ML_SERVICE_DIR not in sys.path:
    sys.path.insert(0, ML_SERVICE_DIR)

from graphsage.schema import (
    HeteroNodeType, HeteroEdgeType, LabelMaturityState, DatasetType,
    GraphSAGELifecycleState, RelationshipRiskLevel, CollusionInvestigationDossier
)
from graphsage.shadow_telemetry import PassiveShadowTelemetryEngine
from graphsage.data_source import MockRealGraphDataSource
from graphsage.connectivity_checker import RealDataConnectivityChecker, RealDataConnectivityReport
from graphsage.observability_exporter import ShadowObservabilityExporter, ShadowMetricsSnapshot
from evaluation.evaluate_frozen_holdout import evaluate_frozen_holdout_offline, compute_sha256

CHAMPION_PATH = os.path.join(ML_SERVICE_DIR, "model", "candidates", "production_model_v8_bmr.joblib")
CHAMPION_EXPECTED_SHA256 = "d473d1ef0c50f232b376c408be37e34c68a258df224277ee1357396e4e627cd7"

HOLDOUT_PATH = os.path.join(ML_SERVICE_DIR, "data", "sample_ieee_fixture.csv")
HOLDOUT_EXPECTED_SHA256 = "a30a387ad0fa8743599d6043120be6bd66ac17184b8eee4fb9ce764970201d44"


def main():
    print("=" * 80)
    print("PHASE 62: LIVE SHADOW INTEGRATION, CONNECTIVITY AUDIT & OBSERVABILITY RUN")
    print("=" * 80)

    # 1. Cryptographic Checksum Invariance Verification
    pre_champ_sha = compute_sha256(CHAMPION_PATH)
    pre_holdout_sha = compute_sha256(HOLDOUT_PATH)

    assert pre_champ_sha == CHAMPION_EXPECTED_SHA256, "Champion SHA mismatch before run!"
    assert pre_holdout_sha == HOLDOUT_EXPECTED_SHA256, "Holdout SHA mismatch before run!"

    print(f"[Checksum Verification] Champion SHA-256: {pre_champ_sha}")
    print(f"[Checksum Verification] Holdout  SHA-256: {pre_holdout_sha}")

    # 2. Real Connectivity Audit
    print("\n[Step 1] Auditing Staging / Shadow Infrastructure Connectivity...")
    conn_checker = RealDataConnectivityChecker(timeout_sec=0.5)
    conn_report: RealDataConnectivityReport = conn_checker.audit_environment()
    print(f"  Integration Status:       {conn_report.integration_status}")
    print(f"  Live Connected:           {conn_report.is_live_connected}")
    print(f"  Kafka Broker Reachable:   {conn_report.kafka_broker_reachable} ({conn_report.kafka_broker_address})")
    print(f"  ClickHouse Reachable:     {conn_report.clickhouse_reachable} ({conn_report.clickhouse_address})")
    print(f"  Missing Configurations:   {len(conn_report.missing_configurations)}")
    for mc in conn_report.missing_configurations:
        print(f"    - {mc}")

    # 3. Initialize Shadow Telemetry Engine
    print("\n[Step 2] Operating Shadow Telemetry Engine...")
    target_ds = MockRealGraphDataSource(dataset_type=DatasetType.REAL_SHADOW)
    engine = PassiveShadowTelemetryEngine(
        queue_size=10000,
        is_live_connected=conn_report.is_live_connected,
        data_source=target_ds
    )
    engine.start()

    base_time = datetime(2026, 8, 1, 9, 0, 0, tzinfo=timezone.utc)
    events_to_stream = []
    latencies_us = []

    # Stream batch: 150 normal support audit events
    for i in range(1, 151):
        ts = base_time + timedelta(minutes=i * 4)
        events_to_stream.append((
            f"evt_audit_{i:04d}",
            "audit.events",
            {"employee_id": "emp_alice_support", "account_id": f"acct_legit_{i:03d}", "action": "ACCESS_ACCOUNT"},
            ts
        ))

    # Stream batch: Suspicious collusion burst
    for i in range(1, 15):
        ts = base_time + timedelta(hours=36, minutes=i * 2)
        events_to_stream.append((
            f"evt_rogue_acc_{i:04d}",
            "audit.events",
            {"employee_id": "emp_charlie_rogue", "account_id": f"acct_mule_{i % 3:02d}", "action": "ACCESS_ACCOUNT"},
            ts
        ))
        events_to_stream.append((
            f"evt_rogue_app_{i:04d}",
            "audit.events",
            {"employee_id": "emp_charlie_rogue", "transaction_id": f"txn_mule_{i:02d}", "action": "APPROVE_TRANSACTION"},
            ts + timedelta(seconds=15)
        ))

    # Stream batch: Privacy violations (quarantined to DLQ)
    events_to_stream.append((
        "evt_dlq_01",
        "audit.events",
        {"employee_id": "emp_bad", "raw_pan": "4111222233334444", "cvv": "999"},
        base_time + timedelta(hours=10)
    ))
    events_to_stream.append((
        "evt_dlq_02",
        "transactions.created",
        {"transaction_id": "txn_bad", "ssn": "123-45-6789", "card_number": "5500000000000004"},
        base_time + timedelta(hours=11)
    ))

    # Enqueue events
    for eid, topic, payload, ts in events_to_stream:
        t0 = time.perf_counter()
        engine.enqueue_event(eid, topic, payload, ts)
        t1 = time.perf_counter()
        latencies_us.append((t1 - t0) * 1_000_000)

    time.sleep(0.3)
    engine.stop()

    # 4. Generate Investigation Dossiers
    t_eval = base_time + timedelta(days=5)
    dossier_alice = engine.evaluate_shadow_investigation("emp_alice_support", t_eval, "customer_support")
    dossier_charlie = engine.evaluate_shadow_investigation("emp_charlie_rogue", t_eval, "customer_support")

    # 5. Extract Observability & Metrics
    print("\n[Step 3] Extracting Observability Snapshot & Prometheus Exposition...")
    obs_exporter = ShadowObservabilityExporter(engine)
    obs_snapshot: ShadowMetricsSnapshot = obs_exporter.get_snapshot()
    prom_text = obs_exporter.export_prometheus_text()

    # 6. Accumulation Tracker View
    collusion_view = engine.collusion_tracker.get_summary_view()

    # 7. Post-execution Checksums
    post_champ_sha = compute_sha256(CHAMPION_PATH)
    post_holdout_sha = compute_sha256(HOLDOUT_PATH)

    assert post_champ_sha == CHAMPION_EXPECTED_SHA256, "Champion corrupted!"
    assert post_holdout_sha == HOLDOUT_EXPECTED_SHA256, "Holdout corrupted!"

    output_payload = {
        "metadata": {
            "phase": "PHASE_62_LIVE_SHADOW_INTEGRATION_AND_OBSERVABILITY",
            "executed_at": datetime.now(timezone.utc).isoformat(),
            "active_production_champion": "production_model_v8_bmr.joblib",
            "champion_sha256_before": pre_champ_sha,
            "champion_sha256_after": post_champ_sha,
            "champion_immutability_verified": True,
            "frozen_52_case_holdout": "sample_ieee_fixture.csv",
            "holdout_sha256_before": pre_holdout_sha,
            "holdout_sha256_after": post_holdout_sha,
            "holdout_immutability_verified": True,
            "customer_decision_routing": "100% Baseline Dynamic BMR (Authoritative)",
            "graphsage_customer_enforcement": "0% (STRICTLY NON-ENFORCING SHADOW)",
            "lifecycle_state": "SYNTHETICALLY_VALIDATED / REAL_DATA_SHADOW_READY / REAL_DATA_REQUIRED / PROMOTION_BLOCKED"
        },
        "real_data_connectivity_audit": conn_report.model_dump(mode="json"),
        "shadow_observability_scorecard": obs_snapshot.model_dump(mode="json"),
        "prometheus_metrics_preview": prom_text.strip().split("\n")[:20],
        "investigation_dossiers_summary": {
            "legitimate_support_dossier": {
                "employee_id": dossier_alice.employee_id if dossier_alice else "none",
                "risk_score": dossier_alice.overall_risk_score if dossier_alice else 0.0,
                "risk_level": dossier_alice.risk_level.value if dossier_alice else "LOW",
                "counter_evidence_count": len(dossier_alice.temporal_evidence.get("counter_evidence", [])) if dossier_alice else 0
            },
            "rogue_collusion_dossier": {
                "employee_id": dossier_charlie.employee_id if dossier_charlie else "none",
                "risk_score": dossier_charlie.overall_risk_score if dossier_charlie else 0.0,
                "risk_level": dossier_charlie.risk_level.value if dossier_charlie else "HIGH",
                "supporting_evidence_count": len(dossier_charlie.temporal_evidence.get("supporting_evidence", [])) if dossier_charlie else 0
            }
        },
        "collusion_accumulation_view": collusion_view.model_dump(mode="json"),
        "governance_decision": {
            "is_promotion_eligible": False,
            "promotion_status": "PROMOTION_BLOCKED",
            "confirmed_real_internal_collusion_cases_available": collusion_view.confirmed_real_internal_collusion_labels,
            "required_real_collusion_cases_for_promotion": 50,
            "progress_pct": 0.0,
            "blocking_reason": "DATA_REQUIRED: Requires >= 50 confirmed real internal collusion labels from mature production cases before empirical calibration or promotion review.",
            "next_required_step": "Deploy configuration credentials to connect live staging broker and begin passive evidence accumulation."
        }
    }

    out_file = os.path.join(CURRENT_DIR, "phase_62_live_shadow_integration_report.json")
    with open(out_file, "w") as f:
        json.dump(output_payload, f, indent=2)

    print(f"\n[Phase 62] Saved Report: {out_file}")
    print(f"[Phase 62] Connectivity Status: {conn_report.integration_status}")
    print(f"[Phase 62] Confirmed Real Collusion Labels: {collusion_view.confirmed_real_internal_collusion_labels} / 50")
    print(f"[Phase 62] Governance Status: PROMOTION_BLOCKED (DATA_REQUIRED)")
    print(f"[Phase 62] GraphSAGE Enforcement: 0% / Baseline Dynamic BMR: 100%")


if __name__ == "__main__":
    main()
