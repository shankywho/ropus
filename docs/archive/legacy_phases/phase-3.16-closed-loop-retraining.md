# Phase 3.16 — Automated Model Retraining Triggering & Closed-Loop Feedback

## 1. Executive Summary

Phase 3.16 closes the ML observability loop for the AI Risk Manager platform by orchestrating an automated, auditable, and resilient model retraining and candidate promotion lifecycle.

The subsystem transforms continuous monitoring into a closed-loop operational pipeline:
$$\text{Traffic} \longrightarrow \text{Risk Evaluation} \longrightarrow \text{Drift Detection} \longrightarrow \text{Sustained Drift Gate} \longrightarrow \text{Candidate Training} \longrightarrow \text{Offline Validation} \longrightarrow \text{Shadow Evaluation} \longrightarrow \text{Operator Approval} \longrightarrow \text{Staged Canary} \longrightarrow \text{Model Promotion}$$

At any point during validation, shadow evaluation, or canary rollout, safety gate breaches or circuit breaker trips immediately trigger an **automated rollback to 0%** and restore the verified production baseline.

---

## 2. Architecture & Data Flow

```
                                  SYNCHRONOUS INFERENCE (Zero Overhead)
┌─────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│ Transaction Request ──► Context Stores ──► Canonical 25F Feature Builder ──► ML Sidecar (fraud-xgb-25f-v3.0)│
│                                                                        │                                    │
│                                                                        ▼                                    │
│                                                            DriftCollector (1.58 µs)                         │
└────────────────────────────────────────────────────────────────────────┼────────────────────────────────────┘
                                                                         │
                                   ASYNCHRONOUS CLOSED-LOOP ORCHESTRATION│
┌────────────────────────────────────────────────────────────────────────┼────────────────────────────────────┐
│                                                                        ▼                                    │
│                                                           DriftDetector Evaluation                          │
│                                                                        │ (Every 5m)                         │
│                                                                        ▼                                    │
│                                                            RetrainingTriggerEngine                          │
│                                                              (Eligibility Gates)                            │
│                                                                        │                                    │
│                                                                        ▼                                    │
│                                                           Retraining State Machine                          │
│                                                                        │                                    │
│         ┌───────────────────┬───────────────────┬──────────────────────┴────────────┬───────────────────┐   │
│         ▼                   ▼                   ▼                                   ▼                   ▼   │
│   LocalTrainingAdapter   OfflineValidator   ShadowScorer (Async)             Operator Approval     CanaryRouter
│  (Controlled Candidate)  (AUC, FPR, Brier)  (Zero Production Decision Risk) (X-Admin-API-Key)     (0% -> 100%)│
│         │                   │                   │                                   │                   │   │
│         └───────────────────┴───────────────────┴───────────────────────────────────┴───────────────────┘   │
│                                                                │                                            │
│                                    ┌───────────────────────────┴───────────────────────────┐                │
│                                    ▼                                                       ▼                │
│                         All Safety Gates Pass                                   Safety Gate / CB Breach     │
│                                    │                                                       │                │
│                                    ▼                                                       ▼                │
│                          AUTOMATIC PROMOTION                                     AUTOMATIC ROLLBACK         │
│                    (Atomic model metadata update,                           (Rollback to 0% traffic,        │
│                     preserve fallback model)                                 restore parent baseline)       │
└─────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 3. Explicit State Machine

Every retraining lifecycle transition is strictly governed by the state machine and logged to ClickHouse:

```
                          ┌─────────────┐
                          │    IDLE     │
                          └──────┬──────┘
                                 │ Trigger Criteria / Manual Trigger
                                 ▼
                          ┌─────────────┐
               ┌──────────┤  TRIGGERED  ├──────────┐
               │          └──────┬──────┘          │
               │                 │                 │
               │                 ▼                 │
               │          ┌─────────────┐          │
               │          │   QUEUED    │          │
               │          └──────┬──────┘          │
               │                 │                 │
               │                 ▼                 │
               │          ┌─────────────┐          │
               │          │  TRAINING   │          │
               │          └──────┬──────┘          │
               │                 │                 │
               │                 ▼                 │
               │          ┌─────────────┐          │
               │          │ VALIDATING  │          │
               │          └──────┬──────┘          │
               │                 │                 │
               │                 ▼                 │
               │          ┌─────────────┐          │
               │          │   SHADOW    │          │
               │          │ EVALUATION  │          │
               │          └──────┬──────┘          │
               │                 │                 │
               │                 ▼                 │
               │          ┌─────────────┐          │
               │          │  AWAITING   │          │
               │          │  APPROVAL   │          │
               │          └──────┬──────┘          │
               │                 │ Operator Approval (X-Admin-API-Key)
               │                 ▼                 │
               │          ┌─────────────┐          │
               │          │   CANARY    ├──────────┼──────────────┐
               │          └──────┬──────┘          │              │
               │                 │ 100% Passed     │              │
               │                 ▼                 │              │
               │          ┌─────────────┐          │              │
               │          │  PROMOTED   │          │              │
               │          └─────────────┘          ▼              ▼
               ▼                             ┌──────────┐  ┌─────────────┐
         ┌───────────┐                       │ REJECTED │  │ ROLLED_BACK │
         │  FAILED   │                       └──────────┘  └─────────────┘
         └───────────┘
```

---

## 4. Multi-Layer Safety Gates & Verification Thresholds

| Safety Layer | Metric / Condition | Threshold / Requirement | Action on Breach |
| :--- | :--- | :--- | :--- |
| **Eligibility Gate** | Consecutive Drift Windows | $\ge 2$ consecutive windows with $\text{PSI} \ge 0.20$ | Suppress trigger |
| **Eligibility Gate** | Sample Quorum | $\ge 200$ live observations in monitoring window | Suppress trigger |
| **Eligibility Gate** | Cooldown Duration | $\ge 30\text{ minutes}$ since last retraining execution | Suppress trigger |
| **Eligibility Gate** | Circuit Breaker State | Must be `HEALTHY` (no active rollback in progress) | Suppress trigger |
| **Dataset Gate** | Data Quality Score | $\ge 0.80$ with zero positive-label omissions | Fail training job |
| **Offline Gate** | ROC-AUC Discrimination | $\text{AUC}_{\text{cand}} \ge \text{AUC}_{\text{base}} - 0.02$ | Reject candidate |
| **Offline Gate** | False Positive Rate (FPR) | $\text{FPR}_{\text{cand}} \le 0.05$ (5% max) | Reject candidate |
| **Offline Gate** | Calibration Brier Score | $\text{Brier}_{\text{cand}} \le 0.10$ | Reject candidate |
| **Offline Gate** | P95 Inference Latency | $p_{95} \le 15.0\text{ ms}$ | Reject candidate |
| **Offline Gate** | Arithmetic Integrity | 0 inference errors, 0 NaNs, 0 Infs | Reject candidate |
| **Shadow Gate** | Decision Divergence Rate | $\le 10.0\%$ decision change vs. production | Reject candidate |
| **Shadow Gate** | Score Divergence Rate | $\le 5.0\%$ significant score divergence | Reject candidate |
| **Shadow Gate** | Shadow Error / Fallback | $\le 1.0\%$ error or fallback rate | Reject candidate |
| **Human-in-Loop** | Operator Approval | Explicit `POST /v1/models/candidates/{id}/approve` | Awaits approval |
| **Canary Gate** | Staged Progression | $1\% \to 5\% \to 10\% \to 25\% \to 50\% \to 100\%$ | Rollback to 0% |
| **Canary Gate** | Circuit Breaker | 0 tripped states during canary observation window | Rollback to 0% |

---

## 5. Training Pipeline Architecture

> [!NOTE]
> **Controlled Adapter vs Production Pipeline Distinction:**
> "Production training pipeline integration remains pending; Phase 3.16 currently validates the closed-loop orchestration using a controlled training adapter."
>
> The `LocalTrainingAdapter` validates real dataset metadata, computes SHA-256 artifact checksums (`daa70da4...`) and configuration hashes, and evaluates immutable candidate models (`fraud-xgb-25f-v3.1-candidate-*`). It provides the exact contract needed for external orchestrators (Airflow, Prefect, Kubernetes Jobs) without modifying the risk engine.

---

## 6. Zero-PII ClickHouse Storage Schemas

No cardholder data, PAN, CVV, raw tokens, or IP addresses are stored.

### 6.1 `retraining_jobs`
```sql
CREATE TABLE IF NOT EXISTS retraining_jobs (
    job_id String,
    triggered_at DateTime,
    state String,
    trigger_type String,
    trigger_reason String,
    parent_model_version String,
    candidate_model_version String,
    dataset_id String,
    sample_count UInt32,
    completed_at DateTime,
    duration_ms Float64,
    error String
) ENGINE = MergeTree()
ORDER BY (triggered_at, job_id);
```

### 6.2 `model_candidates`
```sql
CREATE TABLE IF NOT EXISTS model_candidates (
    model_id String,
    version String,
    parent_model_version String,
    feature_contract String,
    calibration_version String,
    training_job_id String,
    dataset_id String,
    created_at DateTime,
    artifact_checksum String,
    config_hash String,
    state String
) ENGINE = MergeTree()
ORDER BY (created_at, model_id);
```

### 6.3 `model_validation_results`
```sql
CREATE TABLE IF NOT EXISTS model_validation_results (
    validation_id String,
    timestamp DateTime,
    model_id String,
    model_version String,
    parent_model_version String,
    roc_auc Float64,
    pr_auc Float64,
    precision Float64,
    recall Float64,
    fpr Float64,
    fnr Float64,
    brier_score Float64,
    calibration_error Float64,
    p95_latency_ms Float64,
    passed UInt8,
    gate_details String
) ENGINE = MergeTree()
ORDER BY (timestamp, validation_id);
```

### 6.4 `model_shadow_evaluations`
```sql
CREATE TABLE IF NOT EXISTS model_shadow_evaluations (
    evaluation_id String,
    timestamp DateTime,
    candidate_model_version String,
    production_model_version String,
    samples_evaluated UInt32,
    score_divergence_rate Float64,
    decision_change_rate Float64,
    error_rate Float64,
    fallback_rate Float64,
    avg_score_delta Float64,
    p95_latency_ms Float64,
    passed UInt8,
    gate_details String
) ENGINE = MergeTree()
ORDER BY (timestamp, evaluation_id);
```

### 6.5 `model_lifecycle_events`
```sql
CREATE TABLE IF NOT EXISTS model_lifecycle_events (
    event_id String,
    timestamp DateTime,
    model_id String,
    model_version String,
    previous_state String,
    new_state String,
    trigger String,
    actor String,
    reason String
) ENGINE = MergeTree()
ORDER BY (timestamp, event_id);
```

### 6.6 `candidate_canary_metrics`
```sql
CREATE TABLE IF NOT EXISTS candidate_canary_metrics (
    metric_id String,
    timestamp DateTime,
    rollout_id String,
    candidate_model_version String,
    stage_percentage UInt8,
    sample_count UInt32,
    error_rate Float64,
    fallback_rate Float64,
    p95_latency_ms Float64,
    p99_latency_ms Float64,
    decision_change_rate Float64,
    passed UInt8,
    action String
) ENGINE = MergeTree()
ORDER BY (timestamp, metric_id);
```

---

## 7. HTTP API Contracts

### 7.1 `GET /v1/retraining/status`
Returns real-time status of the retraining state machine and active candidate:
```json
{
  "state": "AWAITING_APPROVAL",
  "active_candidate": {
    "model_id": "model_cand_20260821130601",
    "version": "fraud-xgb-25f-v3.1-candidate-20260821130601",
    "parent_model_version": "fraud-xgb-25f-v3.0",
    "feature_contract": "v2.5",
    "calibration_version": "beta-calibrated-v2.5",
    "state": "AWAITING_APPROVAL",
    "artifact_checksum": "daa70da4392598bf...",
    "validation_result": { "passed": true },
    "shadow_result": { "passed": true }
  },
  "summary": {
    "enabled": true,
    "state": "AWAITING_APPROVAL",
    "candidate_model": "fraud-xgb-25f-v3.1-candidate-20260821130601",
    "cooldown_remaining_sec": 1780,
    "training_adapter_status": "LOCAL_EXECUTION_ADAPTER"
  }
}
```

### 7.2 `POST /v1/retraining/trigger` (Admin Protected)
Requires `X-Admin-API-Key` and non-empty `reason`.
```bash
curl -X POST http://localhost:8080/v1/retraining/trigger \
  -H "X-Admin-API-Key: <ADMIN_KEY>" \
  -H "Content-Type: application/json" \
  -d '{"reason": "Manual operator trigger for high fraud volume shift", "actor": "MLOPS_ADMIN"}'
```

### 7.3 `POST /v1/models/candidates/{id}/approve` (Admin Protected)
Approves candidate model to initiate staged canary rollout:
```bash
curl -X POST http://localhost:8080/v1/models/candidates/model_cand_20260821130601/approve \
  -H "X-Admin-API-Key: <ADMIN_KEY>" \
  -H "Content-Type: application/json" \
  -d '{"reason": "Approved candidate model for staged rollout after offline scorecard review", "actor": "MLOPS_LEAD"}'
```

### 7.4 `POST /v1/models/candidates/{id}/reject` (Admin Protected)
Rejects candidate model and records audit rationale:
```bash
curl -X POST http://localhost:8080/v1/models/candidates/model_cand_20260821130601/reject \
  -H "X-Admin-API-Key: <ADMIN_KEY>" \
  -H "Content-Type: application/json" \
  -d '{"reason": "Rejected candidate due to high FPR in merchant segment", "actor": "MLOPS_LEAD"}'
```

---

## 8. Performance Benchmarks

Measured on Apple M4 (ARM64, 10 Cores):

| Operation | Latency / op | Memory Allocations | Impact on Request Path |
| :--- | :--- | :--- | :--- |
| `EvaluateDrift` Trigger Check | **347.3 ns** | 120 B / op, 4 allocs | Background worker |
| `ValidateCandidate` Gates | **1.20 µs** | 460 B / op, 5 allocs | Background worker |
| `RetrainingCoordinator.GetStatus` | **310.1 ns** | 464 B / op, 4 allocs | Status endpoint |
| `IngestVector` (Synchronous Push) | **1.58 µs** | 0 B / op, 0 allocs | Synchronous path ($< 2\mu s$) |

---

## 9. Operational Runbook

### Handling `AWAITING_APPROVAL` Candidates
1. Inspect candidate validation scorecard:
   ```bash
   curl -s http://localhost:8080/v1/models/candidates/<MODEL_ID>/validation | jq .
   ```
2. Inspect shadow evaluation metrics:
   ```bash
   curl -s http://localhost:8080/v1/models/candidates/<MODEL_ID>/shadow | jq .
   ```
3. To approve candidate for staged rollout:
   ```bash
   curl -X POST http://localhost:8080/v1/models/candidates/<MODEL_ID>/approve \
     -H "X-Admin-API-Key: <KEY>" \
     -H "Content-Type: application/json" \
     -d '{"reason": "Approved by ML engineer after reviewing metrics", "actor": "NAME"}'
   ```
4. If candidate metrics are unsatisfactory, reject:
   ```bash
   curl -X POST http://localhost:8080/v1/models/candidates/<MODEL_ID>/reject \
     -H "X-Admin-API-Key: <KEY>" \
     -H "Content-Type: application/json" \
     -d '{"reason": "Excessive divergence in chargeback features", "actor": "NAME"}'
   ```
