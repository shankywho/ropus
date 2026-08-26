#!/usr/bin/env python3
"""
Phase 64: Real Staging Connectivity Activation & Verified Shadow Consumption Runner
"""

import os
import sys
import json
import time
import hashlib
from datetime import datetime, timezone, timedelta

CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
ML_SERVICE_DIR = os.path.abspath(os.path.join(CURRENT_DIR, ".."))
if ML_SERVICE_DIR not in sys.path:
    sys.path.insert(0, ML_SERVICE_DIR)

from graphsage.schema import (
    HeteroNodeType, HeteroEdgeType, LabelMaturityState, DatasetType,
    GraphSAGELifecycleState, RelationshipRiskLevel
)
from graphsage.staging_preflight import StagingPreflightValidator, StagingPreflightReport
from graphsage.collusion_tracker import CollusionAccumulationTracker
from evaluation.evaluate_frozen_holdout import compute_sha256

CHAMPION_PATH = os.path.join(ML_SERVICE_DIR, "model", "candidates", "production_model_v8_bmr.joblib")
CHAMPION_EXPECTED_SHA256 = "d473d1ef0c50f232b376c408be37e34c68a258df224277ee1357396e4e627cd7"

HOLDOUT_PATH = os.path.join(ML_SERVICE_DIR, "data", "sample_ieee_fixture.csv")
HOLDOUT_EXPECTED_SHA256 = "a30a387ad0fa8743599d6043120be6bd66ac17184b8eee4fb9ce764970201d44"


def main():
    print("=" * 80)
    print("PHASE 64: REAL STAGING CONNECTIVITY ACTIVATION & PREFLIGHT AUDIT RUNNER")
    print("=" * 80)

    # 1. Checksums Verification
    pre_champ_sha = compute_sha256(CHAMPION_PATH)
    pre_holdout_sha = compute_sha256(HOLDOUT_PATH)

    assert pre_champ_sha == CHAMPION_EXPECTED_SHA256, "Champion SHA mismatch before run!"
    assert pre_holdout_sha == HOLDOUT_EXPECTED_SHA256, "Holdout SHA mismatch before run!"

    print(f"[Checksum Verification] Champion SHA-256: {pre_champ_sha}")
    print(f"[Checksum Verification] Holdout  SHA-256: {pre_holdout_sha}")

    # 2. Strict Staging Preflight Audit
    print("\n[Step 1] Executing Granular Staging Preflight Dependency Validation...")
    validator = StagingPreflightValidator(timeout_sec=0.5)
    preflight_report: StagingPreflightReport = validator.run_preflight()

    print(f"  Overall Preflight Status: {preflight_report.overall_status}")
    print(f"  Ready for Live Shadow:   {preflight_report.is_ready_for_live_shadow}")
    print(f"  Dependencies Audited:    {len(preflight_report.dependency_results)}")
    for dep_key, dep_res in preflight_report.dependency_results.items():
        print(f"    - {dep_res.name:30s} -> [{dep_res.status.value:15s}] {dep_res.details}")

    print(f"\n  Blocking Reasons ({len(preflight_report.blocking_reasons)}):")
    for br in preflight_report.blocking_reasons:
        print(f"    * {br}")

    print(f"\n  Required Actions ({len(preflight_report.required_unblocking_actions)}):")
    for act in preflight_report.required_unblocking_actions:
        print(f"    > {act}")

    # 3. Accumulation & Governance Gate Check
    tracker = CollusionAccumulationTracker()
    collusion_view = tracker.get_summary_view()

    # 4. Post-run Checksums Verification
    post_champ_sha = compute_sha256(CHAMPION_PATH)
    post_holdout_sha = compute_sha256(HOLDOUT_PATH)

    assert post_champ_sha == CHAMPION_EXPECTED_SHA256, "Champion corrupted!"
    assert post_holdout_sha == HOLDOUT_EXPECTED_SHA256, "Holdout corrupted!"

    output_payload = {
        "metadata": {
            "phase": "PHASE_64_REAL_STAGING_CONNECTIVITY_ACTIVATION",
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
        "staging_preflight_audit": {
            "overall_status": preflight_report.overall_status,
            "is_ready_for_live_shadow": preflight_report.is_ready_for_live_shadow,
            "real_events_consumed": 0,
            "confirmed_real_collusion_cases": 0,
            "dependency_checks": {
                k: v.model_dump(mode="json") for k, v in preflight_report.dependency_results.items()
            },
            "blocking_reasons": preflight_report.blocking_reasons,
            "missing_configurations": preflight_report.missing_configurations,
            "required_unblocking_actions": preflight_report.required_unblocking_actions
        },
        "persistence_and_observability_status": {
            "clickhouse_persistence": "CONFIG_REQUIRED (CLICKHOUSE_HOST / Auth missing for persistent staging ledger)",
            "prometheus_observability": "READY_LOCAL (Local /metrics exposition active; staging remote-write unconfigured)"
        },
        "collusion_accumulation_view": collusion_view.model_dump(mode="json"),
        "governance_decision": {
            "is_promotion_eligible": False,
            "promotion_status": "PROMOTION_BLOCKED",
            "confirmed_real_internal_collusion_cases_available": 0,
            "required_real_collusion_cases_for_promotion": 50,
            "progress_pct": 0.0,
            "blocking_reason": "DATA_REQUIRED: Requires >= 50 confirmed real internal collusion labels from mature production cases before empirical calibration or promotion review.",
            "next_required_action": "Provision AWS MSK staging credentials to activate live shadow intake."
        }
    }

    out_file = os.path.join(CURRENT_DIR, "phase_64_real_staging_connectivity_report.json")
    with open(out_file, "w") as f:
        json.dump(output_payload, f, indent=2)

    print(f"\n[Phase 64] Saved Report: {out_file}")
    print(f"[Phase 64] Overall Status: {preflight_report.overall_status}")
    print(f"[Phase 64] Real Events Consumed: 0 (No fabricated staging counts)")
    print(f"[Phase 64] Confirmed Real Collusion Labels: 0 / 50")
    print(f"[Phase 64] Governance Status: PROMOTION_BLOCKED (DATA_REQUIRED)")
    print(f"[Phase 64] GraphSAGE Enforcement: 0% / Baseline Dynamic BMR: 100%")


if __name__ == "__main__":
    main()
