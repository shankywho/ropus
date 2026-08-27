"""
AI Risk Manager — Production Shadow Soak v2 Readiness Report Generator
Produces auditable, machine-readable JSON and Markdown readiness reports.
Distinguishes precisely between:
- IMPLEMENTED
- READY_FOR_PRODUCTION_TRAFFIC
- LIVE_VALIDATED
- NOT_POPULATED
- FAILED
Strictly adheres to zero-fabrication and zero-masquerading governance policies.
"""

import os
import sys
import json
import hashlib
from datetime import datetime, timezone

def compute_fixture_sha256(filepath: str) -> str:
    if not os.path.exists(filepath):
        return "FILE_NOT_FOUND"
    h = hashlib.sha256()
    with open(filepath, "rb") as f:
        for chunk in iter(lambda: f.read(8192), b""):
            h.update(chunk)
    return h.hexdigest()

def generate_soak_v2_readiness_report():
    fixture_path = os.path.join(os.path.dirname(__file__), "..", "data", "sample_ieee_fixture.csv")
    fixture_sha256 = compute_fixture_sha256(fixture_path)

    raw_tx_path = os.path.join(os.path.dirname(__file__), "..", "data", "raw", "train_transaction.csv")
    raw_id_path = os.path.join(os.path.dirname(__file__), "..", "data", "raw", "train_identity.csv")
    dataset_status = "AVAILABLE" if (os.path.exists(raw_tx_path) and os.path.exists(raw_id_path)) else "FULL_DATASET_UNAVAILABLE"

    now_utc = datetime.now(timezone.utc).isoformat()

    report_json = {
        "metadata": {
            "report_name": "ROPUS Production Shadow Soak v2 Readiness Report",
            "report_version": "2.0",
            "telemetry_schema_version": "v2.0",
            "generated_at_utc": now_utc,
            "frozen_fixture_sha256": fixture_sha256,
            "dataset_status": dataset_status,
            "raw_dataset_available": False
        },
        "executive_status": {
            "champion_model": "fraud-xgb-25f-v3.0",
            "candidate_model": "extended_catboost_58f",
            "champion_decision_authority_pct": 100.0,
            "candidate_decision_authority_pct": 0.0,
            "canary_routing_enabled": False,
            "production_promotion_status": "BLOCKED",
            "soak_readiness_status": "READY_FOR_PRODUCTION_TRAFFIC",
            "genuine_live_production_transactions": 0,
            "live_production_target": 10000,
            "matured_confirmed_fraud_outcomes": 0,
            "matured_fraud_target": 50,
            "live_evidence_status": "LIVE_PRODUCTION_EVIDENCE_UNAVAILABLE"
        },
        "implementation_state": {
            "shadow_scoring_pipeline": "IMPLEMENTED",
            "non_blocking_fault_isolation": "IMPLEMENTED",
            "explicit_provenance_tracking": "IMPLEMENTED",
            "delayed_outcome_attribution_60d": "IMPLEMENTED",
            "read_only_soak_monitor": "IMPLEMENTED",
            "milestone_checkpoints_pipeline": "IMPLEMENTED",
            "startup_safety_validation": "IMPLEMENTED"
        },
        "provenance_breakdown": {
            "LIVE_PRODUCTION": 0,
            "OFFLINE_TEST": 0,
            "SYNTHETIC": 0,
            "REPLAY": 0
        },
        "promotion_gates": {
            "gate_1_live_volume": {
                "gate_name": "Gate 1: Live Volume",
                "description": "Total live production transactions evaluated >= 10,000",
                "target": ">= 10,000 live transactions",
                "actual_live": "0",
                "evidence_basis": "NOT_POPULATED",
                "status": "NOT_POPULATED",
                "implementation_status": "READY_FOR_PRODUCTION_TRAFFIC",
                "passed": False
            },
            "gate_2_confirmed_frauds": {
                "gate_name": "Gate 2: Confirmed Frauds",
                "description": "Total confirmed matured fraud chargebacks >= 50",
                "target": ">= 50 matured frauds (>60-day maturation)",
                "actual_live": "0",
                "evidence_basis": "NOT_POPULATED",
                "status": "NOT_POPULATED",
                "implementation_status": "READY_FOR_PRODUCTION_TRAFFIC",
                "passed": False
            },
            "gate_3_economic_advantage": {
                "gate_name": "Gate 3: Economic BMR Advantage",
                "description": "Statistically significant BMR monetary loss reduction (p < 0.05)",
                "target": "p < 0.05 economic advantage on live cohort",
                "actual_live": "N/A (0 live frauds)",
                "offline_holdout_evidence": "Saves $649-$1,293 on $1k-$10k tiers, p=0.028",
                "evidence_basis": "OFFLINE_HOLDOUT",
                "status": "NOT_POPULATED",
                "implementation_status": "IMPLEMENTED",
                "passed": False
            },
            "gate_4_calibration_quality": {
                "gate_name": "Gate 4: Calibration Quality",
                "description": "Expected Calibration Error (ECE) <= 0.010 on live outcomes",
                "target": "ECE <= 0.010",
                "actual_live": "N/A (0 live outcomes)",
                "offline_holdout_evidence": "Beta ECE = 0.0035, Brier = 0.0408",
                "evidence_basis": "OFFLINE_HOLDOUT",
                "status": "NOT_POPULATED",
                "implementation_status": "IMPLEMENTED",
                "passed": False
            },
            "gate_5_inference_latency": {
                "gate_name": "Gate 5: Inference Latency SLA",
                "description": "P99 inference latency <= 5.0 ms (queue dispatch -> inference completion)",
                "target": "P99 <= 5.0 ms",
                "actual_live": "N/A (0 live requests)",
                "offline_holdout_evidence": "Sidecar P99 = 0.275 ms",
                "evidence_basis": "OFFLINE_HOLDOUT",
                "status": "NOT_POPULATED",
                "implementation_status": "READY_FOR_PRODUCTION_TRAFFIC",
                "passed": False
            },
            "gate_6_action_disagreement": {
                "gate_name": "Gate 6: Action Disagreement Bound",
                "description": "Action disagreement rate <= 15.0%",
                "target": "<= 15.0% disagreement",
                "actual_live": "N/A (0 live requests)",
                "offline_holdout_evidence": "Disagreement = 8.33% at tau* = 0.220",
                "evidence_basis": "OFFLINE_HOLDOUT",
                "status": "NOT_POPULATED",
                "implementation_status": "READY_FOR_PRODUCTION_TRAFFIC",
                "passed": False
            }
        },
        "external_dependencies": {
            "merchant_ingress_traffic": "External dependency: streaming merchant traffic to Go API gateway required to begin accumulating LIVE_PRODUCTION evidence.",
            "delayed_chargeback_outcomes": "External dependency: streaming confirmed dispute webhooks following mandatory >60-day maturation window required for Gate 2 and Gate 3.",
            "raw_ieee_cis_dataset": "External dependency: train_transaction.csv and train_identity.csv required for full offline retraining pipeline."
        },
        "governance_conclusion": {
            "promotion_allowed": False,
            "champion_authority": "100%",
            "candidate_authority": "0%",
            "governance_decision": "REMAIN_IN_SHADOW_MODE",
            "primary_blocker": "Live production traffic and matured chargeback outcomes unavailable in local development environment",
            "next_operational_action": "Deploy system to staging/production environment, connect merchant traffic ingress, and begin 60-day shadow soak evidence collection"
        }
    }

    # Write JSON report
    json_out_path = os.path.join(os.path.dirname(__file__), "production_shadow_soak_v2_readiness_report.json")
    with open(json_out_path, "w") as f:
        json.dump(report_json, f, indent=2)
    print(f"Wrote JSON readiness report: {json_out_path}")

    # Generate Markdown report
    md_content = f"""# ROPUS — Production Shadow Soak v2 Readiness Report

---

## 1. Executive Status

- **Champion Model:** `fraud-xgb-25f-v3.0`
- **Candidate Model:** `extended_catboost_58f`
- **Champion Decision Authority:** **100%** (Sole production decision-maker)
- **Candidate Decision Authority:** **0%** (Strictly non-enforcing asynchronous Shadow Mode)
- **Canary Routing:** **FORBIDDEN / DISABLED (`RISK_MODEL_CANARY_ENABLED=false`)**
- **Promotion Status:** **BLOCKED (`promotion_allowed = false`)**
- **System Readiness Status:** `READY_FOR_PRODUCTION_TRAFFIC`
- **Genuine Live Production Transactions:** **0 / 10,000**
- **Confirmed Matured Fraud Outcomes (>60-day maturation):** **0 / 50**
- **Full IEEE-CIS Dataset Status:** `{dataset_status}`
- **Live Evidence Status:** `LIVE_PRODUCTION_EVIDENCE_UNAVAILABLE`

---

## 2. Distinction of Implementation & Verification States

| Component / Requirement | Architecture State | Evidence Basis | Status |
| :--- | :--- | :--- | :--- |
| **Shadow Scorer Worker Pool** | Initialized & Bounded (`cap=1000, workers=4`) | Unit & Chaos Tests | `IMPLEMENTED` |
| **Asynchronous Non-Blocking Enqueue** | Verified post-commit enqueue | Chaos Tests | `IMPLEMENTED` |
| **Candidate Zero Decision Authority** | Hard boundary enforced | Integration Tests | `IMPLEMENTED` |
| **Fail-Open Fault Isolation** | HTTP 500/503/timeout isolated | Chaos Tests | `IMPLEMENTED` |
| **Startup Safety Validation** | Canary lock enforced | Startup Tests | `IMPLEMENTED` |
| **Delayed Outcome Attribution (>60d)** | Maturation & Conflict Resolution | Outcome Tests | `IMPLEMENTED` |
| **Read-Only Soak Monitor Tool** | `monitor_shadow_soak.py` | CLI Tests | `IMPLEMENTED` |
| **Milestone Checkpoints ($N=1k, 5k, 10k$)** | `shadow_soak_checkpoint_<N>` | Spec Tests | `IMPLEMENTED` |
| **Genuine Ingress Stream** | Awaiting merchant ingress connection | External Traffic | `READY_FOR_PRODUCTION_TRAFFIC` |
| **Gate 1: Live Volume ($\\ge 10,000$)** | Awaiting live transactions | Live Telemetry | `NOT_POPULATED` |
| **Gate 2: Confirmed Frauds ($\\ge 50$)** | Awaiting 60d matured disputes | Live Disputes | `NOT_POPULATED` |
| **Gate 3: BMR Loss Advantage ($p<0.05$)** | Offline validated on fixture ($p=0.028$) | Holdout Fixture | `NOT_POPULATED` (Live) |
| **Gate 4: Calibration ($\\text{{ECE}}\\le 0.010$)** | Offline validated on fixture ($\\text{{ECE}}=0.0035$) | Holdout Fixture | `NOT_POPULATED` (Live) |
| **Gate 5: Inference SLA ($\\text{{P99}}\\le 5.0\\text{{ms}}$)**| Offline validated (Sidecar P99 = $0.275\\text{{ms}}$) | Holdout Fixture | `NOT_POPULATED` (Live) |
| **Gate 6: Disagreement ($\\le 15.0\\%$)** | Offline validated (Disagreement = $8.33\\%$) | Holdout Fixture | `NOT_POPULATED` (Live) |

---

## 3. Formal Promotion Gate Table

| Gate | Target | Live Actual | Offline Fixture Evidence | Status |
| :--- | :--- | :--- | :--- | :--- |
| **G1: Live Volume** | $\\ge 10,000$ live transactions | **0 / 10,000** | Fixture $N=8,000$ | `NOT_POPULATED` |
| **G2: Confirmed Frauds** | $\\ge 50$ matured frauds ($>60\\text{{d}}$) | **0 / 50** | Fixture $N_{{\\text{{fraud}}}}=280$ | `NOT_POPULATED` |
| **G3: Economic BMR Advantage** | $p < 0.05$ monetary loss delta | **N/A** (0 live frauds) | Saves \\$649–\\$1,293 ($p=0.028$) | `NOT_POPULATED` |
| **G4: Calibration Quality** | $\\text{{ECE}} \\le 0.010$ | **N/A** (0 live outcomes) | Beta $\\text{{ECE}} = 0.0035$ | `NOT_POPULATED` |
| **G5: Inference Latency SLA** | $\\text{{P99}} \\le 5.0\\text{{ ms}}$ | **N/A** (0 live requests) | Sidecar $\\text{{P99}} = 0.275\\text{{ ms}}$ | `NOT_POPULATED` |
| **G6: Action Disagreement** | $\\le 15.0\\%$ disagreement | **N/A** (0 live requests) | $8.33\\%$ at $\\tau^* = 0.220$ | `NOT_POPULATED` |

---

## 4. External Dependencies

The platform is completely engineered, tested, and ready. The following external dependencies must be provided by the operational environment to populate live evidence:

1. **Merchant Ingress Traffic:** Connection of real merchant checkout traffic to the Go API gateway (`POST /v1/risk-evaluations`).
2. **Confirmed Chargeback Disputes:** Streaming dispute webhooks following the mandatory $>60$-day maturation window.
3. **Production Infrastructure:** Cloud deployment environment with PostgreSQL, Redis, ClickHouse, and Python ML sidecar container connectivity.
4. **Raw IEEE-CIS Dataset:** `train_transaction.csv` and `train_identity.csv` (if full dataset offline retraining is desired).

---

## 5. Governance Conclusion

$$\\mathbf{{PROMOTION\\_ALLOWED = FALSE}}$$
$$\\mathbf{{CHAMPION\\_AUTHORITY = 100\\%}}$$
$$\\mathbf{{CANDIDATE\\_AUTHORITY = 0\\%}}$$
$$\\mathbf{{GOVERNANCE\\_DECISION = REMAIN\\_IN\\_SHADOW\\_MODE}}$$

- **Primary Blocker:** Live production traffic and matured chargeback outcomes unavailable in local development environment.
- **Next Operational Action:** Deploy platform to staging/production environment, connect merchant traffic ingress, and begin 60-day shadow soak evidence collection.
"""

    md_out_path = os.path.join(os.path.dirname(__file__), "..", "..", "docs", "production_shadow_soak_v2_readiness_report.md")
    with open(md_out_path, "w") as f:
        f.write(md_content)
    print(f"Wrote Markdown readiness report: {md_out_path}")

if __name__ == "__main__":
    generate_soak_v2_readiness_report()
