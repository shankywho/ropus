#!/usr/bin/env python3
"""
Phase 63: Real Staging Connectivity & Shadow Soak Master Runner
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
    GraphSAGELifecycleState, RelationshipRiskLevel
)
from graphsage.staging_soak_runner import StagingShadowSoakHarness, StagingSoakMetricScorecard
from graphsage.collusion_tracker import CollusionAccumulationTracker
from evaluation.evaluate_frozen_holdout import compute_sha256

CHAMPION_PATH = os.path.join(ML_SERVICE_DIR, "model", "candidates", "production_model_v8_bmr.joblib")
CHAMPION_EXPECTED_SHA256 = "d473d1ef0c50f232b376c408be37e34c68a258df224277ee1357396e4e627cd7"

HOLDOUT_PATH = os.path.join(ML_SERVICE_DIR, "data", "sample_ieee_fixture.csv")
HOLDOUT_EXPECTED_SHA256 = "a30a387ad0fa8743599d6043120be6bd66ac17184b8eee4fb9ce764970201d44"


def main():
    print("=" * 80)
    print("PHASE 63: REAL STAGING CONNECTIVITY & SHADOW SOAK RUNNER")
    print("=" * 80)

    # 1. Checksums Verification
    pre_champ_sha = compute_sha256(CHAMPION_PATH)
    pre_holdout_sha = compute_sha256(HOLDOUT_PATH)

    assert pre_champ_sha == CHAMPION_EXPECTED_SHA256, "Champion SHA mismatch before run!"
    assert pre_holdout_sha == HOLDOUT_EXPECTED_SHA256, "Holdout SHA mismatch before run!"

    print(f"[Checksum Verification] Champion SHA-256: {pre_champ_sha}")
    print(f"[Checksum Verification] Holdout  SHA-256: {pre_holdout_sha}")

    # 2. Configuration Discovery & Environment Audit
    print("\n[Step 1] Configuration Discovery & Infrastructure Validation...")
    harness = StagingShadowSoakHarness(queue_capacity=10000)
    conn_report = harness.discover_and_validate_environment()

    print(f"  Integration Status:       {conn_report.integration_status}")
    print(f"  Live Connected:           {conn_report.is_live_connected}")
    print(f"  Kafka Broker Reachable:   {conn_report.kafka_broker_reachable} ({conn_report.kafka_broker_address})")
    print(f"  ClickHouse Reachable:     {conn_report.clickhouse_reachable} ({conn_report.clickhouse_address})")
    print(f"  Missing Configurations:   {len(conn_report.missing_configurations)}")
    for mc in conn_report.missing_configurations:
        print(f"    - {mc}")

    # 3. Controlled Shadow Soak Execution
    print("\n[Step 2] Executing Controlled Shadow Soak Workload...")
    soak_scorecard: StagingSoakMetricScorecard = harness.run_controlled_soak(
        event_count=300,
        is_live_connected=conn_report.is_live_connected
    )

    print(f"  Data Source Classification:       {soak_scorecard.data_source_classification}")
    print(f"  Total Soak Events Received:        {soak_scorecard.total_events_received}")
    print(f"  Total Soak Events Processed:       {soak_scorecard.total_events_processed}")
    print(f"  Total DLQ Quarantined Events:      {soak_scorecard.total_events_quarantined_dlq}")
    print(f"  Duplicate Events Filtered:         {soak_scorecard.duplicate_events_count}")
    print(f"  Out-of-Order Events Handled:       {soak_scorecard.out_of_order_events_count}")
    print(f"  Graph Nodes in Memory:             {soak_scorecard.graph_nodes_total}")
    print(f"  Graph Edges in Memory:             {soak_scorecard.graph_edges_total}")
    print(f"  Investigation Dossiers Generated:  {soak_scorecard.dossiers_generated_total}")
    print(f"  Async Enqueue p95 Latency:         {soak_scorecard.latency_p95_us:.2f} us")
    print(f"  End-to-End Pipeline p95 Latency:   {soak_scorecard.pipeline_latency_p95_ms:.4f} ms")

    # 4. Accumulation & Governance Gate Check
    tracker = CollusionAccumulationTracker()
    collusion_view = tracker.get_summary_view()

    # 5. Post-run Checksums Verification
    post_champ_sha = compute_sha256(CHAMPION_PATH)
    post_holdout_sha = compute_sha256(HOLDOUT_PATH)

    assert post_champ_sha == CHAMPION_EXPECTED_SHA256, "Champion corrupted!"
    assert post_holdout_sha == HOLDOUT_EXPECTED_SHA256, "Holdout corrupted!"

    output_payload = {
        "metadata": {
            "phase": "PHASE_63_REAL_STAGING_CONNECTIVITY_AND_SHADOW_SOAK",
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
        "configuration_discovery_and_audit": {
            "environment_status": conn_report.integration_status,
            "is_live_staging_connected": conn_report.is_live_connected,
            "missing_configurations_count": len(conn_report.missing_configurations),
            "missing_configurations": conn_report.missing_configurations,
            "infrastructure_audit": {
                "kafka_broker_reachable": conn_report.kafka_broker_reachable,
                "kafka_broker_address": conn_report.kafka_broker_address,
                "clickhouse_reachable": conn_report.clickhouse_reachable,
                "clickhouse_address": conn_report.clickhouse_address,
                "prometheus_reachable": conn_report.prometheus_reachable,
                "prometheus_address": conn_report.prometheus_address
            },
            "stop_condition_status": "STAGING_CONNECTIVITY_BLOCKED (Minimum infrastructure credentials required)"
        },
        "controlled_shadow_soak_scorecard": soak_scorecard.model_dump(mode="json"),
        "collusion_accumulation_view": collusion_view.model_dump(mode="json"),
        "governance_decision": {
            "is_promotion_eligible": False,
            "promotion_status": "PROMOTION_BLOCKED",
            "confirmed_real_internal_collusion_cases_available": collusion_view.confirmed_real_internal_collusion_labels,
            "required_real_collusion_cases_for_promotion": 50,
            "progress_pct": 0.0,
            "blocking_reason": "DATA_REQUIRED: Requires >= 50 confirmed real internal collusion labels from mature production cases before empirical calibration or promotion review.",
            "next_required_step": "Provision AWS MSK staging credentials to connect live staging broker and begin passive real case accumulation."
        }
    }

    out_file = os.path.join(CURRENT_DIR, "phase_63_staging_shadow_soak_report.json")
    with open(out_file, "w") as f:
        json.dump(output_payload, f, indent=2)

    print(f"\n[Phase 63] Saved Report: {out_file}")
    print(f"[Phase 63] Stop Condition: STAGING_CONNECTIVITY_BLOCKED")
    print(f"[Phase 63] Confirmed Real Collusion Labels: {collusion_view.confirmed_real_internal_collusion_labels} / 50")
    print(f"[Phase 63] Governance Status: PROMOTION_BLOCKED (DATA_REQUIRED)")
    print(f"[Phase 63] GraphSAGE Enforcement: 0% / Baseline Dynamic BMR: 100%")


if __name__ == "__main__":
    main()
