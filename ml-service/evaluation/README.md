# ROPUS ML Service Evaluation & Model Governance Index

This directory contains the model evaluation pipelines, statistical monitoring engines, shadow evaluation telemetry, and governance audit trails for the ROPUS platform.

---

## 🏛️ Authoritative Governance & Evaluation Documentation

The following canonical documents represent the active authoritative state of the ROPUS risk decision engine:

| Document | Purpose & Key Findings |
| :--- | :--- |
| **[`docs/conditional_promotion_governance_decision.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/conditional_promotion_governance_decision.md)** | **Master Promotion Policy & Dual-Track Decision.** Formalizes Track 1 (`CONDITIONAL_PROMOTION_ELIGIBLE` for $\le 10\%$ canary with maker-checker controls) and Track 2 (`PRODUCTION_PROMOTION_BLOCKED` for full replacement until 10k live tx + 50 matured chargebacks). |
| **[`docs/production_shadow_soak_live_evidence.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/production_shadow_soak_live_evidence.md)** | **Live Shadow Evidence Ledger.** Live telemetry metrics, 6-gate evaluation status, and zero-masquerading verification. |
| **[`docs/canonical_features_specification.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/canonical_features_specification.md)** | **Point-in-Time Feature Contracts.** 58-feature schema definitions and strict point-in-time calculation rules. |
| **[`docs/architecture/graphsage-status-and-roadmap.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/graphsage-status-and-roadmap.md)** | **GraphSAGE Status & Roadmap.** Multi-relational GNN architecture, four-pillar status, and `INFRASTRUCTURE_BLOCKED` staging prerequisites. |
| **[`docs/architecture/graphsage-relationship-intelligence.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/graphsage-relationship-intelligence.md)** | **GraphSAGE Relationship Intelligence Report.** Inductive 2-hop heterogeneous graph neural network evaluation and collusion detection. |
| **[`docs/full_dataset_ml_and_production_evidence_report.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/full_dataset_ml_and_production_evidence_report.md)** | **Full Dataset ML & Production Evidence Report.** Complete offline empirical validation, BMR curves, and benchmark scores. |
| **[`MODEL_PERFORMANCE_DIAGNOSIS_REPORT.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/ml-service/evaluation/MODEL_PERFORMANCE_DIAGNOSIS_REPORT.md)** | **Historical Calibration Diagnosis.** Brier score, ECE analysis, and Bayes Minimum Risk cost-curve comparisons. |
| **[`WEEKLY_PRODUCTION_HEALTH_REPORT.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/ml-service/evaluation/WEEKLY_PRODUCTION_HEALTH_REPORT.md)** | **Production Health & Governance Monitor.** Operational telemetry checklist, drift thresholds, and alert playbooks. |

---

## 🔒 Current Model Governance State

1. **Active Production Champion**:
   - Model: `fraud-xgb-25f-v3.0` (25 Canonical Features)
   - Decision Authority: **100% Customer Decision Authority**
2. **Shadow Candidate**:
   - Model: `extended_catboost_58f` (58 Point-in-Time Features)
   - Authority: **0% Customer Decision Authority** (Asynchronous non-blocking shadow evaluation)
3. **Dual-Track Governance Status**:
   - `MODEL_VALIDATED`: **`true`**
   - `PRODUCTION_SHADOW_READY`: **`true`**
   - `LIVE_EVIDENCE_UNAVAILABLE`: **`true`**
   - `CONDITIONAL_PROMOTION_ELIGIBLE`: **`true`** (Canary $\le 10\%$ with maker-checker signoff)
   - `PRODUCTION_PROMOTION_BLOCKED`: **`true`** (Unconditional 100% replacement locked until 10,000 live tx and 50 matured chargebacks)
4. **Frozen Holdout Fixture**:
   - Checksum SHA-256: `a30a387ad0fa8743599d6043120be6bd66ac17184b8eee4fb9ce764970201d44`

---

## ⚙️ Active Evaluation Scripts

- [`evaluate_models.py`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/ml-service/evaluation/evaluate_models.py) — Baseline and candidate model evaluation harness.
- [`shadow_evaluator.py`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/ml-service/evaluation/shadow_evaluator.py) — Authoritative 6-gate promotion evaluator with explainable governance flags.
- [`live_shadow_soak_monitor.py`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/ml-service/evaluation/live_shadow_soak_monitor.py) — Real-time telemetry monitor and promotion gate manifest generator.
- [`run_reproducible_experiment.py`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/ml-service/data_pipeline/run_reproducible_experiment.py) — Deterministic offline validation pipeline.
