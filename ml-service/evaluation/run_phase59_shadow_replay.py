#!/usr/bin/env python3
"""
Phase 59: Real-Data Shadow Replay, Relationship Intelligence & Latency Benchmark Runner
"""

import os
import sys
import json
import time
import hashlib
import numpy as np
from datetime import datetime, timezone

CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
ML_SERVICE_DIR = os.path.abspath(os.path.join(CURRENT_DIR, ".."))
if ML_SERVICE_DIR not in sys.path:
    sys.path.insert(0, ML_SERVICE_DIR)

from graphsage.schema import (
    HeteroNodeType, HeteroEdgeType, LabelMaturityState, DatasetType,
    GraphSAGELifecycleState, ShadowInvestigationOutput, RelationshipRiskLevel
)
from graphsage.shadow_replay import generate_realistic_shadow_fixture
from graphsage.data_source import PointInTimeReplayHarness
from graphsage.path_extractor import RelationshipPathExtractor
from graphsage.relationship_features import RelationshipFeatureExtractor
from graphsage.employee_baseline import EmployeeBehavioralBaselineEngine
from graphsage.calibration import ShadowCollusionCalibrator
from graphsage.drift_monitor import ShadowGraphDriftMonitor
from graphsage.dossier_generator import CollusionDossierGenerator
from graphsage.model import GraphSAGEModel

PROD_CHAMPION_PATH = os.path.join(ML_SERVICE_DIR, "model", "candidates", "production_model_v8_bmr.joblib")
PROD_EXPECTED_SHA256 = "d473d1ef0c50f232b376c408be37e34c68a258df224277ee1357396e4e627cd7"


def main():
    print("=" * 80)
    print("PHASE 59: REAL-DATA SHADOW REPLAY & RELATIONSHIP INTELLIGENCE BENCHMARK")
    print("=" * 80)

    # 1. Check production champion checksum
    hasher = hashlib.sha256()
    with open(PROD_CHAMPION_PATH, "rb") as f:
        while chunk := f.read(65536):
            hasher.update(chunk)
    champion_sha256 = hasher.hexdigest()
    assert champion_sha256 == PROD_EXPECTED_SHA256, "CRITICAL: Production model checksum altered!"

    # 2. Benchmark Graph Construction
    t_gc_0 = time.perf_counter()
    fixture = generate_realistic_shadow_fixture()
    t_gc_1 = time.perf_counter()
    gc_us = (t_gc_1 - t_gc_0) * 1_000_000

    model = GraphSAGEModel(input_dim=32, hidden_dim=64, output_dim=64, seed=42)
    path_extractor = RelationshipPathExtractor()
    feature_extractor = RelationshipFeatureExtractor()
    baseline_engine = EmployeeBehavioralBaselineEngine()
    calibrator = ShadowCollusionCalibrator(confirmed_labels_count=0)
    drift_monitor = ShadowGraphDriftMonitor()
    dossier_gen = CollusionDossierGenerator()

    # 3. Benchmark Execution Latencies
    latencies = {
        "graph_construction_us": [gc_us],
        "sampling_us": [],
        "gnn_forward_us": [],
        "path_extraction_us": [],
        "dossier_generation_us": []
    }

    all_nodes = {n.id: n for n in fixture.get_nodes()}
    all_edges = fixture.get_edges()
    txns = fixture.get_transactions()

    evaluated_dossiers = []
    evaluated_scores = []

    employees = [n for n in all_nodes.values() if n.type == HeteroNodeType.EMPLOYEE]

    # Warmup + benchmark across 20 iterations to collect smooth distributions
    for _ in range(20):
        for emp in employees:
            t0 = time.perf_counter()
            emp_edges = [e for e in all_edges if e.source_id == emp.id or e.target_id == emp.id]
            l1_nodes = []
            for e in emp_edges:
                target_id = e.target_id if e.source_id == emp.id else e.source_id
                if target_id in all_nodes:
                    l1_nodes.append(all_nodes[target_id])
            t1 = time.perf_counter()
            latencies["sampling_us"].append((t1 - t0) * 1_000_000)

            # GNN forward pass
            t0 = time.perf_counter()
            emb, signals = model.forward(emp, l1_nodes, [])
            t1 = time.perf_counter()
            latencies["gnn_forward_us"].append((t1 - t0) * 1_000_000)

            # Path extraction
            t0 = time.perf_counter()
            paths = path_extractor.find_paths(emp.id, None, all_nodes, emp_edges, max_depth=3)
            t1 = time.perf_counter()
            latencies["path_extraction_us"].append((t1 - t0) * 1_000_000)

            # Structural features
            rel_feats = feature_extractor.extract_features(emp.id, all_nodes, emp_edges, txns)
            struct_score = rel_feats["neighborhood_anomaly_score"]
            final_risk = max(signals["relationship_collusion_risk"], struct_score)
            evaluated_scores.append(final_risk)

            shadow_out = ShadowInvestigationOutput(
                root_node_id=emp.id,
                root_node_type=HeteroNodeType.EMPLOYEE,
                graph_risk_score=final_risk,
                collusion_risk_score=final_risk,
                employee_risk_score=final_risk,
                consumer_risk_score=0.10,
                neighborhood_anomaly_score=struct_score,
                graph_poisoning_score=0.01,
                embedding=emb.tolist(),
                top_related_entities=[],
                suspicious_paths=[p["path_string"] for p in paths[:3]],
                relationship_evidence=[],
                point_in_time_timestamp=datetime.now(timezone.utc)
            )

            # Dossier generation
            t0 = time.perf_counter()
            dossier = dossier_gen.generate_dossier(
                employee_node=emp,
                shadow_output=shadow_out,
                incident_edges=emp_edges,
                associated_nodes=l1_nodes,
                transaction_records=txns
            )
            t1 = time.perf_counter()
            latencies["dossier_generation_us"].append((t1 - t0) * 1_000_000)

            if len(evaluated_dossiers) < len(employees):
                evaluated_dossiers.append(dossier.model_dump(mode="json"))

    # Calibration result
    cal_res = calibrator.calibrate(raw_score=0.85, relationship_score=0.80).model_dump(mode="json")

    # Drift report
    drift_report = drift_monitor.compute_drift_report(
        current_nodes=list(all_nodes.values()),
        current_edges=all_edges,
        current_scores=evaluated_scores,
        baseline_stats={"average_degree": 3.8, "mean_score": 0.20},
        current_labels=fixture.get_labels(mature_only=False)
    )

    # Compute latency percentiles (in milliseconds)
    perf_summary = {}
    for k, v in latencies.items():
        arr = np.array(v) / 1000.0  # Convert to ms
        if len(arr) > 0:
            perf_summary[k.replace("_us", "_ms")] = {
                "p50_ms": round(float(np.percentile(arr, 50)), 3),
                "p95_ms": round(float(np.percentile(arr, 95)), 3),
                "p99_ms": round(float(np.percentile(arr, 99)), 3),
                "mean_ms": round(float(np.mean(arr)), 3)
            }

    total_shadow_latency_p95 = sum(p["p95_ms"] for k, p in perf_summary.items() if k != "graph_construction_ms")

    output_payload = {
        "metadata": {
            "phase": "PHASE_59_REAL_DATA_SHADOW_REPLAY",
            "executed_at": datetime.now(timezone.utc).isoformat(),
            "target_system": "GraphSAGE Real-Data Shadow Replay & Relationship Intelligence",
            "active_production_champion": "production_model_v8_bmr.joblib",
            "champion_sha256": champion_sha256,
            "production_routing": "UNCHANGED (Baseline Dynamic BMR Authoritative)",
            "operational_mode": "STRICTLY NON-ENFORCING SHADOW / INVESTIGATION ONLY",
            "lifecycle_state": "SYNTHETICALLY_VALIDATED / REAL_DATA_SHADOW_READY / REAL_DATA_REQUIRED / PROMOTION_BLOCKED"
        },
        "shadow_replay_fixture_summary": {
            "total_nodes": len(all_nodes),
            "total_edges": len(all_edges),
            "total_transactions": len(txns),
            "scenarios_evaluated": [
                "1. Legitimate Customer Support (Alice: 20 accounts, normal hours, low concentration)",
                "2. Legitimate Fraud Analyst (Bob: case reviews, benign outcomes)",
                "3. Rogue Support Employee (Charlie: 3-account concentration, rapid pre-tx access, self-approval)",
                "4. Off-Hours Collusion (David: 02:30 AM access, shared mobile hardware colocation)",
                "5. Low-and-Slow Collusion (Eve: multi-week spaced access)"
            ]
        },
        "performance_latency_benchmarks": {
            "percentiles": perf_summary,
            "total_shadow_inference_p95_ms": round(total_shadow_latency_p95, 3),
            "is_within_shadow_budget": total_shadow_latency_p95 < 15.0
        },
        "shadow_calibration_status": cal_res,
        "shadow_graph_drift_report": drift_report,
        "sample_investigation_dossiers": evaluated_dossiers[:3],
        "governance_decision": {
            "is_promotion_eligible": False,
            "promotion_status": "PROMOTION_BLOCKED",
            "reason": "GraphSAGE is shadow-ready but lacks >= 50 confirmed real internal collusion labels.",
            "next_required_step": "Deploy passive Kafka consumer to ingest shadow audit telemetry."
        }
    }

    out_file = os.path.join(CURRENT_DIR, "phase_59_real_data_shadow_report.json")
    with open(out_file, "w") as f:
        json.dump(output_payload, f, indent=2)

    print(f"[Phase 59] Saved Report: {out_file}")
    print(f"[Phase 59] Total Shadow Pipeline p95 Latency: {total_shadow_latency_p95:.3f} ms")
    print(f"[Phase 59] Calibration Status: {cal_res['calibration_status']}")
    print(f"[Phase 59] Governance Status: PROMOTION_BLOCKED (DATA_REQUIRED)")


if __name__ == "__main__":
    main()
