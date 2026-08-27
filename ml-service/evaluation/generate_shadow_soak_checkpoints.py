"""
AI Risk Manager — Shadow Soak Checkpoints Specification & Generator
Defines the structure and generates governed checkpoint reports for N = 1000, 5000, and 10000 transactions.
Enforces that even at N = 10000, promotion remains BLOCKED until matured outcomes (Gate 2) and statistical BMR criteria (Gate 3) are independently validated.
"""

import os
import sys
import json
from datetime import datetime, timezone
from typing import Dict, Any

CHECKPOINT_HORIZONS = [1000, 5000, 10000]

def build_checkpoint_data(n_target: int, actual_live: int = 0) -> Dict[str, Any]:
    now_utc = datetime.now(timezone.utc).isoformat()
    return {
        "checkpoint_id": f"SOAK_CHECKPOINT_{n_target}",
        "target_horizon_transactions": n_target,
        "evaluated_at_utc": now_utc,
        "governance_status": {
            "champion_model": "fraud-xgb-25f-v3.0",
            "candidate_model": "extended_catboost_58f",
            "champion_authority_pct": 100.0,
            "candidate_authority_pct": 0.0,
            "canary_routing_enabled": False,
            "promotion_allowed": False,
            "governance_decision": "REMAIN_IN_SHADOW_MODE"
        },
        "volume_metrics": {
            "genuine_live_production_count": actual_live,
            "unique_evaluation_count": actual_live,
            "duplicate_evaluations_dropped": 0,
            "provenance_breakdown": {
                "LIVE_PRODUCTION": actual_live,
                "OFFLINE_TEST": 0,
                "SYNTHETIC": 0,
                "REPLAY": 0
            }
        },
        "operational_metrics": {
            "candidate_success_count": 0,
            "candidate_failure_count": 0,
            "queue_drops": 0,
            "inference_latency_p50_ms": 0.0,
            "inference_latency_p95_ms": 0.0,
            "inference_latency_p99_ms": 0.0,
            "action_disagreement_rate": 0.0,
            "champion_impact_detected": False
        },
        "outcome_maturation": {
            "matured_confirmed_frauds": 0,
            "target_matured_frauds": 50,
            "immature_outcomes_pending": 0,
            "maturation_window_days": 60
        },
        "gate_status_summary": {
            "gate_1_live_volume": "NOT_POPULATED" if actual_live == 0 else "INSUFFICIENT_SAMPLE" if actual_live < 10000 else "LIVE_VALIDATED",
            "gate_2_confirmed_frauds": "NOT_POPULATED",
            "gate_3_economic_advantage": "NOT_POPULATED",
            "gate_4_calibration_quality": "NOT_POPULATED",
            "gate_5_inference_latency": "NOT_POPULATED",
            "gate_6_action_disagreement": "NOT_POPULATED"
        },
        "checkpoint_verdict": f"Checkpoint {n_target} awaiting genuine live production traffic accumulation" if actual_live < n_target else f"Checkpoint {n_target} reached; awaiting delayed chargeback maturation (>60 days)"
    }

def generate_checkpoint_markdown(data: Dict[str, Any]) -> str:
    n = data["target_horizon_transactions"]
    return f"""# ROPUS — Production Shadow Soak Checkpoint ({n:,} Transactions)

---

## 1. Checkpoint Summary

- **Checkpoint Horizon:** **{n:,} Live Transactions**
- **Evaluation Timestamp (UTC):** `{data["evaluated_at_utc"]}`
- **Champion Model:** `fraud-xgb-25f-v3.0` (**100% Authority**)
- **Candidate Model:** `extended_catboost_58f` (**0% Authority, Shadow Mode**)
- **Promotion Status:** **BLOCKED (`promotion_allowed = false`)**
- **Governance Decision:** `REMAIN_IN_SHADOW_MODE`

---

## 2. Checkpoint Metrics & Provenance Audit

| Category | Metric | Value | Target / SLA |
| :--- | :--- | :--- | :--- |
| **Volume** | Genuine `LIVE_PRODUCTION` Count | **{data["volume_metrics"]["genuine_live_production_count"]}** | $\\ge {n:,}$ |
| **Volume** | Unique Evaluations Stored | **{data["volume_metrics"]["unique_evaluation_count"]}** | Matches live count |
| **Operational** | Candidate P99 Inference Latency | **{data["operational_metrics"]["inference_latency_p99_ms"]} ms** | $\\le 5.0\\text{{ ms}}$ |
| **Operational** | Candidate Failure Count | **{data["operational_metrics"]["candidate_failure_count"]}** | $0$ impact |
| **Operational** | Queue Drops | **{data["operational_metrics"]["queue_drops"]}** | $0$ impact |
| **Operational** | Action Disagreement Rate | **{data["operational_metrics"]["action_disagreement_rate"]*100:.2f}%** | $\\le 15.0\\%$ |
| **Operational** | Champion Decision / Latency Impact | **NONE DETECTED** | Zero Impact |
| **Maturation** | Confirmed Matured Frauds ($>60\\text{{d}}$) | **{data["outcome_maturation"]["matured_confirmed_frauds"]}** | $\\ge 50$ (Gate 2) |

---

## 3. Promotion Gates at Checkpoint {n:,}

| Gate | Target | Checkpoint Status | Status Enum |
| :--- | :--- | :--- | :--- |
| **Gate 1** | $\\ge 10,000$ live transactions | {data["volume_metrics"]["genuine_live_production_count"]} / 10,000 | `{data["gate_status_summary"]["gate_1_live_volume"]}` |
| **Gate 2** | $\\ge 50$ matured confirmed frauds | 0 / 50 | `{data["gate_status_summary"]["gate_2_confirmed_frauds"]}` |
| **Gate 3** | $p < 0.05$ BMR economic advantage | Requires $\\ge 50$ matured frauds | `{data["gate_status_summary"]["gate_3_economic_advantage"]}` |
| **Gate 4** | Live $\\text{{ECE}} \\le 0.010$ | Requires $\\ge 50$ matured frauds | `{data["gate_status_summary"]["gate_4_calibration_quality"]}` |
| **Gate 5** | Live P99 latency $\\le 5.0\\text{{ ms}}$ | Requires live traffic stream | `{data["gate_status_summary"]["gate_5_inference_latency"]}` |
| **Gate 6** | Live disagreement $\\le 15.0\\%$ | Requires live traffic stream | `{data["gate_status_summary"]["gate_6_action_disagreement"]}` |

---

## 4. Governance Rule on Milestone Completion

> [!IMPORTANT]
> Reaching the {n:,} transaction horizon does **NOT** automatically authorize candidate promotion or canary routing. Promotion remains strictly blocked until all 6 gates, including the mandatory 60-day delayed fraud maturation window and dual-control Maker-Checker approvals, are verified.
"""

def generate_all_checkpoints():
    base_dir = os.path.dirname(__file__)
    docs_dir = os.path.join(base_dir, "..", "..", "docs")

    for n in CHECKPOINT_HORIZONS:
        data = build_checkpoint_data(n, actual_live=0)

        # Write JSON
        json_path = os.path.join(base_dir, f"shadow_soak_checkpoint_{n}.json")
        with open(json_path, "w") as f:
            json.dump(data, f, indent=2)
        print(f"Wrote JSON checkpoint: {json_path}")

        # Write Markdown
        md_path = os.path.join(docs_dir, f"shadow_soak_checkpoint_{n}.md")
        with open(md_path, "w") as f:
            f.write(generate_checkpoint_markdown(data))
        print(f"Wrote Markdown checkpoint: {md_path}")

if __name__ == "__main__":
    generate_all_checkpoints()
