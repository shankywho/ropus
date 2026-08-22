# AI Risk Manager 🛡️

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Next.js Version](https://img.shields.io/badge/Next.js-15_App_Router-black?style=flat&logo=next.js)](https://nextjs.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=flat&logo=redis)](https://redis.io/)
[![ONNX Runtime](https://img.shields.io/badge/ONNX_Runtime-1.17+-005CED?style=flat&logo=onnx)](https://onnxruntime.ai/)
[![Redpanda](https://img.shields.io/badge/Redpanda-Kafka_Compatible-FF0055?style=flat&logo=apachekafka)](https://redpanda.com/)
[![ClickHouse](https://img.shields.io/badge/ClickHouse-24_OLAP-FFCC01?style=flat&logo=clickhouse)](https://clickhouse.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

An enterprise-grade, real-time **Payments Fraud, Abuse, and Chargeback Defense Platform**. Built with **Go**, **Next.js 15**, **ONNX Runtime**, **PostgreSQL 16**, **Redis 7**, **Redpanda (Kafka)**, **Debezium CDC**, and **ClickHouse OLAP**.

The system orchestrates multi-stage synchronous fraud decisioning (**<100ms p95 SLA**), declarative JSON-AST rules with strict **Maker-Checker dual-control**, asynchronous **24-hour SLA analyst review queues**, **AES-256-GCM Envelope Encryption & Crypto-Shredding**, and analytical telemetry.

---

## 🏛️ System Architecture

```mermaid
flowchart TD
    subgraph Edge_and_Ingestion [Edge & Ingestion]
        M[Merchant / Payment Gateway] -->|POST /v1/risk-evaluations| API[Go API Gateway :8080]
        PG[Payment Provider] -->|POST /webhooks/provider| API
        UI[Next.js 15 Analyst Dashboard :3000] <-->|REST API| API
    end

    subgraph Synchronous_Path [Synchronous Decision Pipeline (<100ms SLA)]
        API --> ORCH[Risk Orchestrator]
        ORCH <-->|ZADD / ZCOUNT| REDIS[(Redis 7 Feature Store)]
        ORCH <-->|Fetch Active Rules| PG_DB[(PostgreSQL 16)]
        ORCH -->|POST /predict (50ms Deadline)| ONNX[ONNX ML Sidecar :8000]
        ORCH -->|Derive DEK & Encrypt PII| KMS[Mock KMS AES-256]
        ORCH -->|Atomic Commit (Decision + Outbox)| PG_DB
    end

    subgraph Asynchronous_Streaming [Asynchronous CDC & Event Streaming]
        PG_DB -.->|Logical WAL Replication| DEB[Debezium Connect :8083]
        DEB -->|EventRouter| REDP[Redpanda / Kafka :9092]
        REDP -->|risk.events| CASE_C[Case Manager Consumer]
        REDP -->|risk.events| AUDIT_C[Audit OLAP Consumer]
        CASE_C -->|Provision 24h SLA Case| PG_DB
        AUDIT_C -->|Batch / Stream Insert| CH[(ClickHouse OLAP :9000)]
    end
```

---

## ⚡ Key Engineering Highlights & Design Decisions

### 1. Synchronous Multi-Stage Decision Pipeline (`<100ms SLA`)
* **Context Aggregation:** Queries real-time sliding-window counters (`velocity.ip.1hr`, `velocity.token.24hr`) from Redis Sorted Sets in `<5ms`.
* **Pre-Rules (Hard Guardrails):** Evaluates deterministic JSON-AST rules. If a hard `DECLINE` or `ALLOW` rule triggers, pipeline evaluation halts immediately to save downstream compute.
* **ONNX Runtime ML Serving (50ms Timeout):** Language-agnostic, sub-millisecond fraud scoring sidecar running compiled XGBoost graphs with **local SHAP-like feature attributions**. If the ML sidecar times out or errors, Go gracefully degrades (`is_degraded: true`) to heuristic rules without blocking the transaction.
* **Post-Rules & Dynamic Thresholds:** Evaluates risk score thresholds to determine the final outcome: `ALLOW_RECOMMENDATION`, `STEP_UP_RECOMMENDATION` (3DS), `MANUAL_REVIEW`, or `DECLINE_RECOMMENDATION`.

### 2. Transactional Outbox Pattern & Zero Data Loss
* Uses a single PostgreSQL ACID transaction (`pgx.Tx`) to commit both the `risk_decisions` record and the `outbox_events` (`risk.decisioned`) entry.
* **Debezium CDC** captures Postgres WAL changes and streams events to **Redpanda (Kafka)**, preventing dual-write inconsistencies.

### 3. Declarative JSON-AST Rules Engine & Maker-Checker Flow
* Zero arbitrary code execution (`eval()` is strictly prohibited). The pure Go AST parser evaluates nested `AND`/`OR`/`NOT` trees and comparison operators (`==`, `!=`, `>`, `<`, `>=`, `<=`, `IN`, `CONTAINS`, `STARTS_WITH`).
* **Dual-Control Governance:** State machine (`DRAFT` $\rightarrow$ `PENDING_APPROVAL` $\rightarrow$ `ACTIVE`) enforces that a rule's creator cannot approve their own rule (`ErrMakerCheckerViolation` / HTTP 403).

### 4. AES-256-GCM Envelope Encryption & Crypto-Shredding
* **At-Rest Protection:** Encrypts PII (`ip_address`, `device_fingerprint`) inside stored feature snapshots using per-tenant 32-byte AES keys derived via KMS.
* **Crypto-Shredding (`ShredTenantKey`):** Permanently destroying a tenant's KMS key mathematically renders all historical PII irrecoverable, fulfilling DPDP Act / GDPR erasure rights without corrupting immutable financial ledgers.
* **Stream Sanitization:** Enforces that raw decrypted PII is never emitted to Kafka or ClickHouse streams.

### 5. ClickHouse OLAP Analytical Sink
* A dedicated Go consumer (`risk-audit-group`) streams all evaluation events into a ClickHouse `MergeTree` table (`risk_audit_log`) for offline auditing, fraud pattern detection, and model retraining datasets.

---

## 🚀 Quickstart Guide

### Prerequisites
* [Docker Desktop](https://www.docker.com/) (version 24.0+) & Docker Compose
* [Go 1.22+](https://go.dev/) (optional, for local CLI development)
* [Node.js 20+](https://nodejs.org/) (optional, for local frontend development)

### Running the Entire Platform
To start all 7 services (API Gateway, PostgreSQL, Redis, Redpanda, Debezium, ClickHouse, ML Sidecar, and Next.js Frontend):

```bash
# Clone the repository
git clone https://github.com/shankywho/ropus.git
cd ropus

# Copy environment variables
cp .env.example .env

# Launch full Docker stack
docker compose up --build -d
```

### Accessing Platform Endpoints

| Service | URL / Port | Description |
|---|---|---|
| **Analyst Dashboard** | [http://localhost:3000](http://localhost:3000) | Next.js 15 UI (Playground, Review Queue, Rules Builder) |
| **API Gateway** | [http://localhost:8080](http://localhost:8080) | Go Risk Engine REST API |
| **ML Inference Sidecar** | [http://localhost:8000/docs](http://localhost:8000/docs) | FastAPI ONNX Runtime Swagger UI |
| **ClickHouse HTTP** | [http://localhost:8123](http://localhost:8123) | ClickHouse OLAP HTTP Interface |
| **Redpanda Console** | [http://localhost:8082](http://localhost:8082) | Redpanda / Kafka Proxy UI |
| **Debezium Connect** | [http://localhost:8083](http://localhost:8083) | Kafka Connect REST API |

---

## 🖥️ Frontend Walkthrough & Demo Guide

### 1. Risk Evaluation Playground (`/playground`) — *Priority #1 Demo*
* Input test transaction parameters (Amount, Currency, Card Token, IP, Device Fingerprint).
* Click **"Fill Synthetic Fraud"** to simulate a high-risk velocity attack (₹95,000, unrecognised device, repeat IP).
* Inspect the real-time outcome badge, **Risk Score (0–100)**, **SHAP Reason Code Chips** (e.g. `HIGH_IP_VELOCITY_BLOCK`), latency budget in milliseconds, and raw contract JSON.

### 2. Manual Review Queue (`/cases`)
* View flagged transactions routed to human analysts with **24-hour SLA countdown timers**.
* Filter by `ALL`, `OPEN`, `UNDER_REVIEW`, and `RESOLVED`.
* Click **"Claim Case"** to assign a case and transition into the investigation view.

### 3. Case Detail & Evidence Inspector (`/cases/[id]`)
* Inspect the point-in-time immutable `feature_snapshot` captured during evaluation.
* Review contributing anomaly codes and provide mandatory resolution rationale.
* Execute **"Approve & Allow"** or **"Reject & Decline"** decisions with automated audit logging.

### 4. Rules Governance & Maker-Checker (`/rules` & `/rules/new`)
* Visual multi-row condition builder compiling into JSON-AST format.
* Switch between **"Analyst A (Creator)"** and **"Analyst B (Reviewer)"** to demonstrate Maker-Checker dual-control enforcement preventing unauthorized self-approvals.

---

## 📡 API Reference

See the full OpenAPI 3.1 specification in [`docs/openapi.yaml`](docs/openapi.yaml).

### Real-Time Risk Evaluation (`POST /v1/risk-evaluations`)
```bash
curl -X POST http://localhost:8080/v1/risk-evaluations \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: 00000000-0000-0000-0000-000000000001" \
  -d '{
    "transaction_id": "txn_demo_12345",
    "amount": 95000,
    "currency": "INR",
    "payment_method": {
      "type": "card",
      "token": "tok_visa_fraud_99"
    },
    "device_fingerprint": "new_device",
    "ip_address": "192.168.1.100"
  }'
```

**Response (`<100ms`):**
```json
{
  "decision_id": "dec_8f7b2c1a-4d3e-4b2a",
  "transaction_id": "txn_demo_12345",
  "recommended_action": "DECLINE_RECOMMENDATION",
  "risk_score": 95,
  "reason_codes": [
    "HIGH_IP_VELOCITY_BLOCK",
    "HIGH_TRANSACTION_AMOUNT"
  ],
  "feature_snapshot_ref": "snap_88a91c2b",
  "latency_ms": 14,
  "is_degraded": false,
  "evaluated_at": "2026-08-21T10:40:00Z"
}
```

---

## 🧪 Testing & Verification

Run backend unit test suites across all packages (`ingestion`, `riskengine`, `rules`, `utils`):

```bash
cd backend
go test -v ./...
```

Verify Next.js frontend production build:

```bash
cd frontend
npm run build
```
## 🛡️ Defense-Only Threat Model & Safety Guarantees

Per the track criteria (*strictly defense-only: anything offense-capable is disqualified*), the system enforces strict physical and architectural boundaries against weaponization:

- **Zero Execution Authority:** Emits risk recommendations only (`ALLOW`, `MANUAL_REVIEW`, `DECLINE`). Contains zero integrations with payment rails, preventing autonomous fund movement.
- **Anti-Probing & Oracle Prevention:** Granular SHAP reason codes are exposed exclusively to authenticated merchant backend webhooks, preventing attackers from using consumer-facing responses to reverse-engineer decision thresholds.
- **Fail-Safe Degradation (Anti-DoS):** If the ML sidecar times out or suffers a DoS attack, fallback scoring defaults strictly to conservative deterministic rules—never a lenient fail-open path.
- **AST Sandbox:** Rules execute via a closed JSON-AST interpreter with zero dynamic `eval()` capability.
- **Maker-Checker Governance & Residual Risk:** Single-user rule activation is blocked in software. Collusion between two accounts is acknowledged as a residual risk and addressed via append-only, immutable audit logging.
- **Model Boundary:** Model artifacts are restricted to private container registries and internal VPC subnets to prevent offline adversarial extraction.
---

## 📄 License
This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.
