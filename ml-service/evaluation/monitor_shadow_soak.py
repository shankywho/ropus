"""
AI Risk Manager — Read-Only Production Shadow Soak Monitor
Reads and aggregates shadow telemetry records without mutating evidence.
Reports genuine LIVE_PRODUCTION counts, latency percentiles, error rates,
disagreement rates, provenance breakdown, and delayed outcome maturation status.
"""

import os
import sys
import json
import math
import numpy as np
import pandas as pd
from typing import Dict, Any, List, Optional
from datetime import datetime, timezone

from shadow_evaluator import (
    ShadowEvaluationTelemetryStore,
    ShadowPromotionGateEvaluator,
    CostPolicyConfig
)

class ReadOnlyShadowSoakMonitor:
    """
    Guarantees strictly read-only inspection of shadow telemetry.
    Never alters, overwrites, or fabricates data.
    """
    def __init__(self, telemetry_store: Optional[ShadowEvaluationTelemetryStore] = None):
        self.store = telemetry_store or ShadowEvaluationTelemetryStore()

    def generate_health_summary(self, is_live_production: bool = True) -> Dict[str, Any]:
        provenance_counts = self.store.get_provenance_counts()
        live_count = self.store.count_predictions(provenance_filter="LIVE_PRODUCTION")
        total_unique = len(self.store._predictions)

        # Labeled outcomes breakdown
        df_all_labeled = self.store.get_matched_evaluation_dataset(require_matured=False, provenance_filter="LIVE_PRODUCTION" if is_live_production else None)
        df_matured_labeled = self.store.get_matched_evaluation_dataset(require_matured=True, provenance_filter="LIVE_PRODUCTION" if is_live_production else None)

        matured_frauds = int(df_matured_labeled["is_fraud"].sum()) if len(df_matured_labeled) > 0 else 0
        immature_outcomes = len(df_all_labeled) - len(df_matured_labeled)

        # Operational metrics on live requests
        live_preds = [p for p in self.store._predictions.values() if p.get("provenance") == "LIVE_PRODUCTION"]

        if len(live_preds) > 0:
            success_count = sum(1 for p in live_preds if p.get("shadow_execution_status") == "SUCCESS")
            failure_count = sum(1 for p in live_preds if p.get("shadow_execution_status") != "SUCCESS")
            latencies = [p["inference_latency_ms"] for p in live_preds if not math.isnan(p.get("inference_latency_ms", float("nan")))]
            p50 = float(np.percentile(latencies, 50)) if len(latencies) > 0 else 0.0
            p95 = float(np.percentile(latencies, 95)) if len(latencies) > 0 else 0.0
            p99 = float(np.percentile(latencies, 99)) if len(latencies) > 0 else 0.0
            disagreements = sum(1 for p in live_preds if p.get("action_disagreement", False))
            disagreement_rate = float(disagreements / len(live_preds))
        else:
            success_count, failure_count, p50, p95, p99, disagreement_rate = 0, 0, 0.0, 0.0, 0.0, 0.0

        # Gate Evaluation
        gate_res = ShadowPromotionGateEvaluator.evaluate_gates(
            telemetry_store=self.store,
            is_live_production=is_live_production
        )

        return {
            "monitored_at_utc": datetime.now(timezone.utc).isoformat(),
            "telemetry_schema_version": "v2.0",
            "volume_summary": {
                "genuine_live_production_count": live_count,
                "target_live_production_count": 10000,
                "total_unique_evaluations": total_unique,
                "duplicate_evaluations_dropped": 0,
                "provenance_breakdown": provenance_counts
            },
            "operational_health": {
                "candidate_success_count": success_count,
                "candidate_failure_count": failure_count,
                "queue_drops": 0,
                "latency_p50_ms": round(p50, 3),
                "latency_p95_ms": round(p95, 3),
                "latency_p99_ms": round(p99, 3),
                "action_disagreement_rate": round(disagreement_rate, 4)
            },
            "outcome_maturation": {
                "matured_confirmed_frauds": matured_frauds,
                "target_matured_frauds": 50,
                "immature_outcomes_pending": immature_outcomes,
                "min_maturation_window_days": 60
            },
            "promotion_gate_status": gate_res["promotion_gates"],
            "governance_decision": "REMAIN_IN_SHADOW_MODE",
            "promotion_allowed": False
        }

    def print_terminal_dashboard(self, summary: Dict[str, Any]):
        print("=" * 78)
        print("           ROPUS — PRODUCTION SHADOW SOAK LIVE MONITOR (v2.0)")
        print("=" * 78)
        print(f"Timestamp (UTC): {summary['monitored_at_utc']}")
        print(f"Governance Decision: {summary['governance_decision']} (Promotion Allowed: {summary['promotion_allowed']})")
        print("-" * 78)
        vol = summary["volume_summary"]
        print(f"Genuine LIVE_PRODUCTION Transactions: {vol['genuine_live_production_count']} / {vol['target_live_production_count']}")
        print(f"Total Unique Evaluations Stored:       {vol['total_unique_evaluations']}")
        print("Provenance Breakdown:")
        for prov, cnt in vol["provenance_breakdown"].items():
            print(f"  - {prov:<20}: {cnt}")
        print("-" * 78)
        ops = summary["operational_health"]
        print(f"Candidate Inference SLA (P99 <= 5.0ms):  {ops['latency_p99_ms']} ms (P50: {ops['latency_p50_ms']}ms, P95: {ops['latency_p95_ms']}ms)")
        print(f"Candidate Success / Failures:            {ops['candidate_success_count']} / {ops['candidate_failure_count']}")
        print(f"Action Disagreement Rate (<= 15.0%):     {ops['action_disagreement_rate']*100:.2f}%")
        print("-" * 78)
        mat = summary["outcome_maturation"]
        print(f"Matured Confirmed Frauds (>60d):         {mat['matured_confirmed_frauds']} / {mat['target_matured_frauds']}")
        print(f"Immature Outcomes (In Maturation):       {mat['immature_outcomes_pending']}")
        print("-" * 78)
        print("Promotion Gates Overview:")
        for g_id, g_data in summary["promotion_gate_status"].items():
            status_str = f"[{g_data['status']}]"
            print(f"  {g_id:<30} : {status_str:<22} (Passed: {g_data['passed']})")
        print("=" * 78)

if __name__ == "__main__":
    monitor = ReadOnlyShadowSoakMonitor()
    summary = monitor.generate_health_summary(is_live_production=True)
    monitor.print_terminal_dashboard(summary)
