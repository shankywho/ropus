# Phase 3 — Device Intelligence Detailed Implementation Plan

**Document Version:** 1.0 (Master Execution Plan)  
**Implementation Strategy:** Incremental, Test-Driven, Zero-Downtime Rollout (Phases 3.0 through 3.13)  

---

## 1. Execution Roadmap Overview

```
Phase 3.0: Audit & Architecture Design (COMPLETED)
   │
   ├──► Phase 3.1: Device Identity Ingestion & Normalization
   │
   ├──► Phase 3.2: PostgreSQL Relational Persistence Schemas (Migration 000004)
   │
   ├──► Phase 3.3: Redis Real-Time Device Feature Store & Sliding Windows
   │
   ├──► Phase 3.4: Account ↔ Device Relationship Graph & Multi-Accounting Detection
   │
   ├──► Phase 3.5: Payment Token ↔ Device Linkage & Card Testing Defense
   │
   ├──► Phase 3.6: Multi-Window Device Velocity & Anomaly Burst Tracking
   │
   ├──► Phase 3.7: Device Reputation & Chargeback Dispute Attribution
   │
   ├──► Phase 3.8: ML Feature Pipeline Expansion (15 ──► 25 Canonical Features)
   │
   ├──► Phase 3.9: XGBoost Model Retraining with Strict Temporal Split
   │
   ├──► Phase 3.10: Beta Calibration & Cost Policy Re-Evaluation
   │
   ├──► Phase 3.11: Fault Injection & Graceful Degradation Testing
   │
   ├──► Phase 3.12: Security Hardening, Log Sanitization & Crypto-Shredding
   │
   └──► Phase 3.13: Shadow Scoring & Staged Production Rollout
```

---

## 2. Phase-by-Phase Specification

### Phase 3.0 — Audit & Architecture Design (CURRENT)
- **Objective:** Establish ground-truth baseline, threat model, 42-case edge case matrix, and zero-leakage feature contract.
- **Affected Files:** `docs/phase-3-*.md`.
- **Production Changes:** **NONE (0 code changes)**.
- **Acceptance Criteria:** 12 comprehensive audit artifacts created; all existing tests pass.

---

### Phase 3.1 — Device Identity Ingestion & Normalization
- **Objective:** Implement secure client fingerprint validation, string sanitization, and tenant-salted SHA-256 hashing.
- **Files Affected:** `backend/internal/features/device.go`, `backend/internal/riskengine/handler.go`.
- **Database / Redis Changes:** None.
- **API Changes:** Format validation on `device_fingerprint` (16–64 alphanumeric chars).
- **Tests:** Unit tests verifying SQL injection immunity, whitespace handling, and tenant salting.
- **Rollback:** Revert handler validation middleware.

---

### Phase 3.2 — PostgreSQL Relational Persistence Schemas
- **Objective:** Deploy migration `000004_device_intelligence.up.sql` creating `devices`, `device_accounts`, `device_payment_instruments`, `device_events`, `device_reputation`.
- **Files Affected:** `backend/migrations/000004_device_intelligence.up.sql`, `backend/internal/features/store.go`.
- **Database Changes:** 5 new partitioned/indexed PostgreSQL tables with foreign keys and unique constraints.
- **API Changes:** None (background persistence).
- **Tests:** Migration up/down tests; unique constraint conflict resolution tests.
- **Rollback:** `migrate down` to 000003.

---

### Phase 3.3 — Redis Real-Time Device Feature Store
- **Objective:** Implement sub-millisecond Redis pipelining for device sliding windows and novelty checks.
- **Files Affected:** `backend/internal/features/velocity.go`, `backend/internal/features/device_cache.go`.
- **Redis Changes:** Keys `{tenant}:vel:dev:1h`, `{tenant}:vel:dev:24h`, `{tenant}:dev:known:{hash}`.
- **Tests:** Pipelined latency benchmark ($< 2.5\text{ms}$); TTL eviction tests.
- **Rollback:** Disable device Redis keys via feature flag.

---

### Phase 3.4 — Account ↔ Device Relationship Graph
- **Objective:** Track distinct accounts per device (`device_unique_accounts_24h`) to flag credential stuffing and fraud farms.
- **Files Affected:** `backend/internal/features/graph.go`, `backend/internal/riskengine/orchestrator.go`.
- **Redis Changes:** `{tenant}:dev:acc24:{hash}` (Redis Set with 25h TTL).
- **Tests:** Multi-account threshold detection tests (2 accounts vs 20 accounts vs 100 accounts).
- **Rollback:** Fallback feature value to `1`.

---

### Phase 3.5 — Payment Token ↔ Device Linkage
- **Objective:** Detect distributed card testing and BIN attacks across devices (`device_unique_tokens_24h`).
- **Files Affected:** `backend/internal/features/payment_link.go`.
- **Redis Changes:** `{tenant}:dev:tok24:{hash}` (Redis Set with 25h TTL).
- **Tests:** Card testing simulation tests (10 tokens on 1 device in 5 minutes).
- **Rollback:** Fallback feature value to `1`.

---

### Phase 3.6 — Multi-Window Device Velocity & Anomaly Bursts
- **Objective:** Monitor 1m, 1h, and 24h device transaction velocity and aggregate monetary volume.
- **Files Affected:** `backend/internal/features/velocity.go`.
- **Tests:** High-frequency transaction burst tests.
- **Rollback:** Omit device velocity from orchestrator payload.

---

### Phase 3.7 — Device Reputation & Chargeback Dispute Integration
- **Objective:** Attribute confirmed fraud and chargeback events from `disputes` table back to `device_reputation`.
- **Files Affected:** `backend/internal/ingestion/webhook.go`, `backend/internal/cases/service.go`.
- **Database Changes:** Writes to `device_reputation`.
- **Tests:** Webhook chargeback dispute replay test updating device trust score.
- **Rollback:** Clear `{tenant}:dev:bad_rep` Redis set.

---

### Phase 3.8 — ML Feature Pipeline Expansion (15 → 25 Features)
- **Objective:** Update `ml-service/data_pipeline/` to extract all 10 device features offline from historical partitions.
- **Files Affected:** `ml-service/data_pipeline/features.py`, `ml-service/features/feature_schema.json`.
- **ML Changes:** Feature vector length expands from 15 to 25.
- **Tests:** Point-in-time future leakage isolation unit tests.
- **Rollback:** Revert schema to 15 canonical features.

---

### Phase 3.9 — XGBoost Model Retraining on 25 Canonical Features
- **Objective:** Train Model C on 25 features with strict 70/15/15 chronological temporal split.
- **Files Affected:** `ml-service/train.py`, `ml-service/model/fraud_model.onnx`.
- **Tests:** Compare ROC-AUC, PR-AUC, Precision@5%, Recall@5% vs Phase 2 Baseline.
- **Rollback:** Retain Phase 2 `fraud_model.onnx` as backup.

---

### Phase 3.10 — Beta Calibration & Cost Policy Re-Evaluation
- **Objective:** Fit BetaCalibrator on 25-feature validation predictions and re-verify calibration slope and ECE.
- **Files Affected:** `ml-service/calibration/calibrator.py`, `ml-service/model/calibration.json`.
- **Tests:** 95% bootstrap confidence intervals for Brier, ECE, Log Loss, and resolution.
- **Rollback:** `calibration.isotonic.backup.json` or Phase 2 Beta artifact.

---

### Phase 3.11 — Fault Injection & Graceful Degradation Testing
- **Objective:** Validate system resiliency under simulated Redis, PostgreSQL, and ML sidecar outages.
- **Files Affected:** `tests/resilience/fault_injection_test.go`.
- **Tests:** Automated chaos injection asserting `is_degraded = true` and $< 15\text{ms}$ response times.
- **Rollback:** N/A (test suite).

---

### Phase 3.12 — Security Hardening, Log Sanitization & Crypto-Shredding
- **Objective:** Mask device identifiers in logs and verify AES-256-GCM envelope encryption and GDPR erasure.
- **Files Affected:** `backend/internal/utils/crypto.go`, `backend/internal/riskengine/orchestrator.go`.
- **Tests:** Cryptographic shredding key deletion verification.
- **Rollback:** N/A.

---

### Phase 3.13 — Shadow Scoring & Staged Production Rollout
- **Objective:** Deploy 25-feature Beta-calibrated pipeline in shadow mode, monitor live latency, and execute cutover.
- **Files Affected:** `docker-compose.yml`, production configs.
- **Rollback:** One-command rollback to Phase 2.3 deployment.
- **Acceptance Criteria:** E2E latency $< 15\text{ms}$ p95; zero regressions across all test suites.
