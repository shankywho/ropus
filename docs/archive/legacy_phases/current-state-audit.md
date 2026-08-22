# Current State Repository Audit — Ground-Truth Baseline

**Repository:** `AI Risk Manager / Ropus`  
**Git Commit:** `d0f85100f21527b24d27ce183117db1604226bbf`  
**Audit Date:** August 21, 2026  
**Auditor:** Antigravity Autonomous Pair Programmer (Phase -1 Baseline)

---

## 1. Executive Summary

This document establishes an uncompromising, source-code-verified baseline of the **AI Risk Manager** repository. Every claim in previous documentation and ADRs was checked directly against executable code in `backend/`, `ml-service/`, `infra/`, and `frontend/`.

### Classification Key:
- **`IMPLEMENTED`**: Fully implemented in executable source code and verified with tests or live execution.
- **`PARTIALLY IMPLEMENTED`**: Core logic exists, but key production paths, background consumers, or edge-case handling are incomplete or unattached.
- **`DOCUMENTED BUT NOT IMPLEMENTED`**: Architecture diagrams, ADRs, or READMEs claim the feature, but no active code exists.
- **`STUB / MOCK BEHAVIOR`**: Simulated in-memory or heuristic logic used in place of production systems.
- **`UNUSED / DEAD CODE`**: Present in the repository but never invoked by runtime entry points.

---

## 2. End-to-End Component Audit

```
HTTP Request (POST /v1/risk-evaluations)
  │
  ├──► [Go Orchestrator] (backend/internal/riskengine/orchestrator.go)
  │       │
  │       ├──► [Redis 7] (backend/internal/features/velocity.go) ── ZSET Sliding Windows
  │       │
  │       ├──► [JSON-AST Rules Engine] (backend/internal/rules/ast.go) ── Pre-Rules Filter
  │       │
  │       ├──► [FastAPI + ONNX Sidecar] (ml-service/serve.py) ── XGBoost ONNX Serving (<2ms)
  │       │
  │       ├──► [Decision Mapping & Thresholds] ── Risk Score (0-100) & Recommendations
  │       │
  │       ├──► [AES-256-GCM Envelope Encryption] (backend/internal/utils/crypto.go)
  │       │
  │       └──► [PostgreSQL 16 Transaction] (risk_decisions + outbox_events tables)
  │
  ├──► [Kafka / Redpanda: risk.events] ── Outbox CDC / Stream
  │       │
  │       ├──► [Case Consumer Group] (backend/internal/cases/consumer.go) ── Auto Case Gen
  │       │
  │       └──► [ClickHouse Audit Group] (backend/internal/audit/consumer.go) ── OLAP Sink
  │
  ├──► [Provider Webhooks] (backend/internal/ingestion/webhook_handler.go) ── HMAC Dispute Packet
  │
  └──► [Next.js 15 Frontend] (frontend/src/app) ── Playground, Cases, Rules, Analytics
```

---

## 3. Detailed Component Breakdown

| Component | Documented Functionality | Source Code Implementation Status | Verification Status | Source Files |
| :--- | :--- | :--- | :--- | :--- |
| **API Server Routing & Lifecycle** | REST API endpoints on port 8080 with Chi router, healthchecks, timeouts, and graceful shutdown. | **`IMPLEMENTED`** — `main.go` registers `/health`, `/v1/risk-evaluations`, `/v1/rules`, `/v1/cases`, and `/v1/webhooks/provider`. | Verified via HTTP tests & live benchmark. | `backend/cmd/api/main.go` |
| **Risk Orchestrator Pipeline** | Synchronous decision pipeline executing velocity queries, rules evaluation, ML sidecar call (50ms budget), and atomic persistence. | **`IMPLEMENTED`** — `Orchestrator.Evaluate()` executes 6 sequential stages. | Verified (p50: 3.46ms, p95: 6.61ms). | `backend/internal/riskengine/orchestrator.go` |
| **Redis Feature Store (Velocity)** | Sliding-window transaction frequency counters for IP (1h) and Payment Token (24h) using Redis Sorted Sets (`ZREMRANGEBYSCORE`, `ZADD`, `ZCOUNT`). | **`IMPLEMENTED`** — Fully implemented with Redis connection pooling and fallback to 0 count on error. | Verified against live Redis container. | `backend/internal/features/velocity.go` |
| **JSON-AST Rules Engine** | Dynamic recursive rule evaluation supporting binary comparisons, string operators, and logical combinators (`AND`, `OR`, `NOT`). | **`IMPLEMENTED`** — Unit tests cover all 13 operators and nested AST structures. | Verified via `ast_test.go` (100% pass). | `backend/internal/rules/ast.go`, `ast_test.go` |
| **Maker-Checker Governance** | Dual-control lifecycle for rules (`DRAFT` $\rightarrow$ `PENDING_APPROVAL` $\rightarrow$ `ACTIVE`/`SHADOW`) prohibiting self-approval (`created_by != approved_by`). | **`IMPLEMENTED`** — Enforced in `Service.TransitionStatus()`. Returns `403 Forbidden` with `maker_checker_violation`. | Verified via live API tests. | `backend/internal/rules/service.go`, `handler.go` |
| **ML Inference Sidecar (ONNX)** | Sub-2ms model serving using `onnxruntime.InferenceSession` on XGBoost binary classifier with local feature attribution. | **`IMPLEMENTED`** — Replaced joblib inference with ONNX Runtime opset 15. | Verified (p50: 0.014ms raw, 1.85ms over HTTP). | `ml-service/serve.py`, `export_onnx.py` |
| **ML Graceful Degradation** | 50ms context timeout fallback to heuristic risk scoring if ML container is down or slow (`is_degraded: true`). | **`IMPLEMENTED`** — Tested with mock client and simulated timeouts in `handler_test.go`. | Verified via unit tests. | `backend/internal/riskengine/mlclient.go`, `orchestrator.go` |
| **Provider Webhooks & Disputes** | HMAC-SHA256 authenticated webhook endpoint to ingest chargeback/dispute events and generate forensic evidence packets. | **`IMPLEMENTED`** — Validates signatures against `WEBHOOK_SECRET` and creates evidence packets linked to decisions. | Verified via `webhook_handler_test.go`. | `backend/internal/ingestion/webhook_handler.go` |
| **Case Management Service** | Manual review queue with 24-hour analyst SLA timers, claim mechanism, and resolution overrides (`ALLOW`/`DECLINE`). | **`IMPLEMENTED`** — Full CRUD, SLA tracking, and audit logging in PostgreSQL. | Verified via live SQL & REST endpoints. | `backend/internal/cases/case_service.go`, `handler.go` |
| **Kafka Case Consumer** | Async background consumer reading `risk.events` topic with `risk-case-manager-group` to auto-provision manual review cases. | **`PARTIALLY IMPLEMENTED`** — Background goroutine is wired in `main.go`, but Debezium connector needs active CDC sync to stream outbox events. | Code exists; requires external Debezium CDC trigger. | `backend/internal/cases/consumer.go` |
| **ClickHouse Analytical Sink** | Columnar OLAP consumer writing sanitized decision events into `risk_audit_log` MergeTree table. | **`PARTIALLY IMPLEMENTED`** — ClickHouse client, table schema, and Kafka consumer group (`risk-audit-group`) are implemented in Go and SQL. | Verified connection & table creation; Kafka ingestion depends on Debezium producer. | `backend/internal/audit/clickhouse.go`, `consumer.go` |
| **Debezium CDC Connector** | PostgreSQL WAL logical replication capturing `outbox_events` table inserts and publishing to Kafka. | **`PARTIALLY IMPLEMENTED`** — JSON config and Docker container exist; connector registration requires REST POST to Debezium on cold boot. | Configuration present in `infra/debezium-connector.json`. | `infra/debezium-connector.json` |
| **Envelope Encryption (Crypto-Shredding)** | AES-256-GCM envelope encryption for PII at rest with per-tenant DEK derivation and key-zeroing for DPDP/GDPR compliance. | **`IMPLEMENTED`** — Unit tests verify encrypt/decrypt roundtrip and cryptographic erasure on key shredding. | Verified via `crypto_test.go`. | `backend/internal/utils/crypto.go`, `kms.go` |
| **KMS Provider Integration** | AWS KMS / HashiCorp Vault key management service for master key wrapping. | **`STUB / MOCK BEHAVIOR`** — `MockKMS` in `backend/internal/utils/kms.go` uses an in-memory map of 32-byte keys. | Production AWS KMS driver is not implemented. | `backend/internal/utils/kms.go` |
| **Real Device Fingerprinting** | Browser entropy telemetry extraction using FingerprintJS sensor. | **`IMPLEMENTED`** — Integrated in `frontend/src/app/playground/page.tsx` on component mount, bound to `device_fingerprint`. | Verified in frontend build. | `frontend/src/app/playground/page.tsx` |
| **Graph / Entity Linkage Network** | Neo4j / Amazon Neptune device-card-IP graph intelligence. | **`DOCUMENTED BUT NOT IMPLEMENTED`** — Mentioned in early design notes; no graph database or entity resolution code exists. | 0 source files present. | N/A |
| **Frontend Web Application** | Next.js 15 App Router interface with Analytics dashboard, live Risk Playground, Case Management, and AST Rule Builder. | **`IMPLEMENTED`** — 6 fully functional routes with Recharts, shadcn/ui components, and typed API client. | Verified via `npm run build` and HTTP 200 checks. | `frontend/src/app/*` |

---

## 4. Source Code vs Documentation Gaps

1. **Synthetic Data vs Claimed Production Fraud Model**:
   - *Documentation Claim:* "Trained XGBoost classifier for enterprise fraud detection."
   - *Code Reality:* The model is trained on a 30,000-row synthetic exponential/poisson dataset generated in `ml-service/train.py`. It has only 5 features and an ROC-AUC of 0.6110 on test data.
2. **KMS Provider**:
   - *Documentation Claim:* "Envelope encryption backed by enterprise KMS."
   - *Code Reality:* Backed by an in-memory `MockKMS` struct in Go with zero persistent external keystore integration.
3. **Debezium Automated Auto-Bootstrap**:
   - *Documentation Claim:* "Seamless Outbox-to-Kafka streaming via Debezium."
   - *Code Reality:* Debezium container starts, but connector registration JSON (`infra/debezium-connector.json`) must be submitted via `curl -X POST http://localhost:8083/connectors` after Postgres replication slot is ready.
4. **SHAP TreeExplainer Runtime**:
   - *Documentation Claim:* "SHAP-based reason codes computed for each transaction."
   - *Code Reality:* `ml-service/serve.py` uses heuristic Z-score deviation ($(\text{val} - \text{median}) / \text{std} \times \text{importance}$) rather than running the full Python `shap.TreeExplainer` during live ONNX evaluation.

---

## 5. Unused or Dead Code Inventory

- `ml-service/model/fraud_model.joblib`: Kept as intermediate export artifact; runtime serving exclusively loads `fraud_model.onnx`.
- `backend/migrations/*_down.sql`: Migration rollback scripts present but unused in container startup.
