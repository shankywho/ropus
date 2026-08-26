# GraphSAGE Relationship Intelligence — Implementation Status & Governance Roadmap

**Document Reference:** `ROPUS-GRAPHSAGE-STATUS-ROADMAP-PHASE-68`
**Current Governance State:** `SYNTHETICALLY_VALIDATED` $\to$ `REAL_DATA_SHADOW_READY` $\to$ `REAL_DATA_REQUIRED` (PROMOTION BLOCKED)
**Last Updated:** 2026-08-26 (Phase 68 Verification Checkpoint)
**Active Production Champion:** [`ml-service/model/candidates/production_model_v8_bmr.joblib`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/ml-service/model/candidates/production_model_v8_bmr.joblib)
**Production Champion SHA-256:** `d473d1ef0c50f232b376c408be37e34c68a258df224277ee1357396e4e627cd7` (**BYTE-FOR-BYTE IMMUTABLE**)
**Frozen 52-Case Real Holdout:** [`ml-service/data/sample_ieee_fixture.csv`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/ml-service/data/sample_ieee_fixture.csv)
**Frozen Holdout SHA-256:** `a30a387ad0fa8743599d6043120be6bd66ac17184b8eee4fb9ce764970201d44` (**BYTE-FOR-BYTE IMMUTABLE**)

---

## 1. Executive Summary & Current Status

The **core engineering and automated software validation phase** of the GraphSAGE Relationship Intelligence subsystem is **substantially complete**.

All required application-side software components—including heterogeneous graph construction, 2-hop GNN inductive sampling, point-in-time causality defenses, privacy/DLQ quarantine, asynchronous shadow telemetry, ClickHouse ledger sinks, Prometheus observability exporters, and 7-scenario chaos resilience—have been implemented and verified across Go and Python suites with 100% green test results.

### Current Operating State (Phase 68 Verification):
* **Customer Decision Routing:** **100% Baseline Dynamic BMR Authoritative** (Zero alteration from GraphSAGE).
* **GraphSAGE Customer Enforcement:** **0% (Strictly Non-Enforcing Shadow / Investigation Intelligence Only)**.
* **Real Staging Connectivity:** **`INFRASTRUCTURE_BLOCKED`** (AWS STS caller identity returns `InvalidClientTokenId`; AWS MSK staging brokers and ClickHouse credentials unmounted).
* **Real Staging Metrics:** All live stream metrics (`REAL_EVENTS_CONSUMED`, `REAL_EVENTS_PERSISTED`, `REAL_DLQ_EVENTS`, consumer lag, shadow latency) are classified as **`NOT_AVAILABLE`** because live staging inference is not currently running.
* **Confirmed Real Collusion Cases:** **`0 / 50`** (Real-data governance gate strictly enforces `DATA_REQUIRED`).
* **Model Promotion Status:** **`PROMOTION_BLOCKED`** (Blocked until 50 mature, confirmed real collusion cases are accumulated).

---

## 2. Four-Pillar Status Breakdown

```
┌──────────────────────────────────────────────────────────────────────────────────────────────────┐
│                                 GRAPHSAGE SUBSYSTEM STATUS                                      │
├───────────────────────────────┬───────────────────────────────┬──────────────────────────────────┤
│ PILLAR A: SOFTWARE & CODE     │ COMPLETE (100%)               │ All pipelines, engines, chaos    │
│                               │                               │ handlers, and schemas built.     │
├───────────────────────────────┼───────────────────────────────┼──────────────────────────────────┤
│ PILLAR B: AUTOMATED TESTS     │ COMPLETE (100% Green)         │ 44 Go tests, 42 Go packages,     │
│                               │                               │ 124 Python tests passing.        │
├───────────────────────────────┼───────────────────────────────┼──────────────────────────────────┤
│ PILLAR C: STAGING INFRA       │ BLOCKED (External Creds)      │ AWS STS invalid; MSK, ClickHouse │
│                               │                               │ & Prometheus endpoints needed.   │
├───────────────────────────────┼───────────────────────────────┼──────────────────────────────────┤
│ PILLAR D: REAL-DATA GATE      │ BLOCKED (0 / 50 Mature Cases) │ 90-day maturation window on real │
│                               │                               │ production internal collusion.   │
└───────────────────────────────┴───────────────────────────────┴──────────────────────────────────┘
```

---

### Pillar A: Completed Engineering & Software Work (Phases 57–68)

1. **Heterogeneous Graph Schema & Inductive Model**:
   - 6 node types (`EMPLOYEE`, `CONSUMER`, `ACCOUNT`, `DEVICE`, `PAYMENT_METHOD`, `IP`) and 6 directed edge types (`ACCESSES`, `OWNS`, `USES_DEVICE`, `USES_PAYMENT`, `TRANSACTS_WITH`, `USES_IP`).
   - Inductive GraphSAGE 2-hop aggregation producing 64-dimensional node embeddings for dynamic and previously unseen entities.
2. **Point-in-Time Temporal Causality**:
   - Strict edge sampling enforcing $t_{\text{edge}} \le t_{\text{eval}}$, eliminating future transaction leakage.
3. **Privacy Boundary & Dead-Letter Quarantine**:
   - Automatic detection, hashing, and isolation of cardholder PANs, unmasked CVVs, and raw PII. Violating events are redirected to dead-letter queues.
4. **Passive Shadow Telemetry & Ingestion**:
   - Asynchronous worker pool with in-memory ring buffers, deduplication, and out-of-order sequence correction.
5. **Observability & Evidence Persistence**:
   - `ShadowEvidenceLedger` persisting evidentiary dossiers to ClickHouse `MergeTree` tables.
   - `ShadowObservabilityExporter` exposing OpenMetrics/Prometheus counters and histograms at `/metrics`.
6. **7-Scenario Chaos Resilience & Fail-Open Isolation**:
   - Verified that Kafka broker disconnects, malformed stream payloads, cleartext PAN injection, ClickHouse store outages, GNN timeouts, GNN model panics, and buffer drops **never alter** customer decisions.

---

### Pillar B: Completed Automated Validation

The entire platform is backed by comprehensive, passing regression test suites:

* **Go GraphSAGE Suite (`backend/internal/graph/graphsage/`):** 44 tests passed (100% green).
* **Go Backend Full Platform Suite (`backend/`):** 42 packages passed (100% green).
* **Python ML & GraphSAGE Test Suite (`ml-service/tests/`):** 124 tests passed (100% green).
* **Synthetic ROPUS Dataset Audit Suite (`synthetic_ropus/validation/`):** 16 tests passed (100% green).
* **Cryptographic Checksum Immutability:** Both the active production champion and the frozen 52-case real fraud holdout verified identical before and after all runs.

---

### Pillar C: Remaining Infrastructure & Deployment Work

The next milestone is **NOT** speculative application code or another offline simulation. The required engineering handoff is provisioning valid infrastructure credentials to the runtime environment:

1. **AWS STS Authentication**:
   - Supply valid `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, and `AWS_REGION` with `SecretsManager:GetSecretValue` permissions.
2. **Staging MSK Kafka Cluster Configuration**:
   - Export `KAFKA_BROKERS` pointing to AWS MSK staging bootstrap brokers (e.g. `b-1.ropus-kafka-staging.internal:9092`).
   - Inject SASL/SCRAM or mTLS credentials from AWS Secrets Manager (`ropus/production/credentials`).
3. **Staging Topic Permissions**:
   - Verify `READ` ACLs for consumer group `ropus-graphsage-shadow-group` on `audit.events` and `transactions.created`.
4. **Staging ClickHouse Endpoint**:
   - Export `CLICKHOUSE_HOST` pointing to the staging ClickHouse cluster with write permissions to `shadow_evidence_ledger`.
5. **Staging Prometheus Endpoint**:
   - Configure Prometheus scrape target or export `PROMETHEUS_PUSHGATEWAY_URL`.

---

### Pillar D: Remaining Real-Data Governance Requirements

Even after staging connectivity is unblocked, GraphSAGE **cannot and will not** be promoted to enforce customer decisions until all formal model governance requirements are satisfied:

1. **Confirmed Real Collusion Case Accumulation**:
   - Must accumulate $\ge 50$ confirmed, human-investigator-verified internal collusion cases from live operations.
2. **90-Day Dispute Maturation Window**:
   - Each case must mature past the mandatory 90-day dispute/chargeback cycle before being labeled as authoritative ground truth.
3. **Empirical Beta Calibration**:
   - Empirical calibration against genuine holdout distributions, proving non-inferiority over Baseline Dynamic BMR.
4. **Model Risk Management Approval**:
   - Formal review and sign-off by the Model Risk Committee under SR 11-7 / NIST AI RMF governance frameworks.

---

## 3. Authoritative Governance Invariants

```
====================================================================================================
ACTIVE PRODUCTION CHAMPION:  production_model_v8_bmr.joblib
CHAMPION SHA-256:            d473d1ef0c50f232b376c408be37e34c68a258df224277ee1357396e4e627cd7
FROZEN 52-CASE HOLDOUT:      sample_ieee_fixture.csv
HOLDOUT SHA-256:             a30a387ad0fa8743599d6043120be6bd66ac17184b8eee4fb9ce764970201d44
CUSTOMER ROUTING AUTHORITY:  100% Baseline Dynamic BMR
GRAPHSAGE ENFORCEMENT:       0% (STRICTLY NON-ENFORCING SHADOW)
PROMOTION GATE STATUS:       BLOCKED (REAL_DATA_REQUIRED — 0 / 50 Real Collusion Cases)
====================================================================================================
```

---

## 4. Phase 57–68 Architectural Documentation Index

| Phase | Focus Area | Architectural Specification | JSON Evidence Artifact |
| :--- | :--- | :--- | :--- |
| **Phase 57** | GraphSAGE Relationship Intelligence | [`docs/architecture/graphsage-relationship-intelligence.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/graphsage-relationship-intelligence.md) | `ml-service/evaluation/` benchmark suite |
| **Phase 58** | Real-Data Readiness & Maturation | [`docs/architecture/graphsage-real-data-readiness.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/graphsage-real-data-readiness.md) | `phase_58_real_data_readiness_report.json` |
| **Phase 59** | Shadow Replay & Multi-Hop Path Extraction | [`docs/architecture/graphsage-real-data-shadow-replay.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/graphsage-real-data-shadow-replay.md) | `phase_59_real_data_shadow_report.json` |
| **Phase 60** | Shadow Ingestion & Privacy Boundary | [`docs/architecture/graphsage-real-data-shadow-ingestion.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/graphsage-real-data-shadow-ingestion.md) | `phase_60_real_data_shadow_ingestion_report.json` |
| **Phase 61** | Production Shadow Telemetry Deployment | [`docs/architecture/graphsage-production-shadow-telemetry.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/graphsage-production-shadow-telemetry.md) | `phase_61_production_shadow_telemetry_report.json` |
| **Phase 62** | Live Shadow Integration & Observability | [`docs/architecture/graphsage-live-shadow-integration.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/graphsage-live-shadow-integration.md) | `phase_62_live_shadow_integration_report.json` |
| **Phase 63** | Staging Shadow Soak Validation | [`docs/architecture/graphsage-staging-shadow-soak.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/graphsage-staging-shadow-soak.md) | `phase_63_staging_shadow_soak_report.json` |
| **Phase 64** | Real Staging Connectivity Discovery | [`docs/architecture/graphsage-real-staging-connectivity.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/graphsage-real-staging-connectivity.md) | `phase_64_real_staging_connectivity_report.json` |
| **Phase 65** | Real Staging Shadow Preflight Audit | [`docs/architecture/graphsage-real-staging-shadow.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/graphsage-real-staging-shadow.md) | `phase_65_real_staging_shadow_report.json` |
| **Phase 66** | Infrastructure Unblock & Activation | [`docs/architecture/graphsage-real-staging-activation.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/graphsage-real-staging-activation.md) | `phase_66_real_staging_activation_report.json` |
| **Phase 67** | Credential Provisioning & Evidence Gate | [`docs/architecture/graphsage-real-staging-infrastructure-gate.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/graphsage-real-staging-infrastructure-gate.md) | `phase_67_real_staging_infrastructure_gate_report.json` |
| **Phase 68** | Post-Provisioning Connectivity Verification | [`docs/architecture/graphsage-post-provisioning-connectivity.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/graphsage-post-provisioning-connectivity.md) | `phase_68_post_provisioning_connectivity_report.json` |
