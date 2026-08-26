# ROPUS ML Service Evaluation & Model Governance Index

This directory contains the authoritative model evaluation pipelines, statistical monitoring engines, production shadow evaluation infrastructure, and complete governance audit trails for the ROPUS platform.

---

## 🏛️ Authoritative Production Documentation

The following documents represent the active, authoritative production state of the ROPUS fraud risk management engine:

| Document | Purpose & Key Findings |
| :--- | :--- |
| **[`ROPUS_PRODUCTION_COMPLETION_REPORT.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/ml-service/evaluation/ROPUS_PRODUCTION_COMPLETION_REPORT.md)** | **Master Authoritative Production Completion Report.** Confirms active production champion `v8.0-bmr-36f` (SHA-256 `d473d1ef0c50...`), 9-gate promotion scorecard, multi-tier evidence ledger, and `PRODUCTION_INFRASTRUCTURE_ACCESS_REQUIRED` status. |
| **[`PHASE_57_GRAPHSAGE_INTERNAL_RELATIONSHIP_INTELLIGENCE_REPORT.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/ml-service/evaluation/PHASE_57_GRAPHSAGE_INTERNAL_RELATIONSHIP_INTELLIGENCE_REPORT.md)** | **Phase 57 GraphSAGE Relationship Intelligence Report.** Inductive 2-hop heterogeneous graph neural network evaluation, employee-consumer collusion detection, and 5-defense benchmark suite. |
| **[`PHASE_56_PRODUCTION_DEPLOYMENT_VALIDATION_REPORT.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/ml-service/evaluation/PHASE_56_PRODUCTION_DEPLOYMENT_VALIDATION_REPORT.md)** | **Production Deployment Validation Report.** Comprehensive audit of Kubernetes manifests (`deploy/kubernetes/`), Terraform AWS IaC (`infra/terraform/aws/`), and cloud authentication diagnostics. |
| **[`MODEL_PERFORMANCE_DIAGNOSIS_REPORT.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/ml-service/evaluation/MODEL_PERFORMANCE_DIAGNOSIS_REPORT.md)** | **Model Performance & Calibration Diagnosis.** Comprehensive calibration, Brier score, ECE analysis, and Bayes Minimum Risk cost-curve comparisons. |
| **[`WEEKLY_PRODUCTION_HEALTH_REPORT.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/ml-service/evaluation/WEEKLY_PRODUCTION_HEALTH_REPORT.md)** | **Production Health & Governance Monitor.** Weekly telemetry checklist, drift thresholds, and alert playbooks. |

---

## 🔒 Production Governance & Safety Invariants

1. **Active Production Champion**:
   - Version: `v8.0-bmr-36f` (`CatBoostClassifier`, depth 4, 160 iterations, 36 Causal Features)
   - SHA-256 Checksum: `d473d1ef0c50f232b376c408be37e34c68a258df224277ee1357396e4e627cd7`
   - Artifact Path: [`ml-service/model/candidates/production_model_v8_bmr.joblib`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/ml-service/model/candidates/production_model_v8_bmr.joblib)
2. **Authoritative Customer Decision Policy**:
   - Baseline Dynamic Bayes Minimum Risk (BMR) with $C_{\text{FP}} = \$25.00$ and Surcharge $= 1.05$:
     $$\text{DECLINE} \iff P > P^*(A) = \frac{25.00}{1.05 \cdot \text{Amount} + 25.00}$$
3. **Shadow Challenger & GraphSAGE Intelligence**:
   - `LightGBM L20 D4` with continuous `BetaCalibrator` ($\tau_{\text{floor}} = 0.040$)
   - `GraphSAGE Heterogeneous Collusion Engine` (64-dim inductive embeddings)
   - Status: **Strictly Non-Enforcing Shadow** (100% fail-open isolation; cannot affect customer decisions)
4. **Promotion Mandate**:
   - Promotion strictly requires $\ge 5,000$ genuine mature live transactions, $\ge 86$ mature live frauds, paired bootstrap 95% CI $> 0$, and $\text{FPR} \le 6.0\%$.

---

## ⚙️ Active Evaluation & Verification Runners

- [`phase_57_graphsage_internal_relationship_intelligence.py`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/ml-service/evaluation/phase_57_graphsage_internal_relationship_intelligence.py) — GraphSAGE heterogeneous relationship evaluation, temporal point-in-time neighbor sampling, and 5-defense benchmark suite.
- [`ropus_production_completion.py`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/ml-service/evaluation/ropus_production_completion.py) — Master end-to-end audit, scorecard evaluation, and report generator.
- [`phase_56_production_deployment_validation.py`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/ml-service/evaluation/phase_56_production_deployment_validation.py) — Production infrastructure and cloud reachability diagnostics.
- [`phase_55_production_infrastructure_handoff.py`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/ml-service/evaluation/phase_55_production_infrastructure_handoff.py) — Infrastructure handoff and architecture mapping.
- [`validate_steady_state_monitoring.py`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/ml-service/evaluation/validate_steady_state_monitoring.py) — Continuous monitoring and drift verification.
