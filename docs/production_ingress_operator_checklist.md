# ROPUS — Production Ingress & Shadow Soak Operator Checklist

---

## 1. Executive Purpose & Governance Invariant

This checklist specifies the exact prerequisites, service dependencies, environment configurations, and operational verification steps required before an operator routes genuine production traffic to the ROPUS fraud-risk platform.

### Immutable Governance Invariant:
- **Production Champion:** `fraud-xgb-25f-v3.0` retains **100% Customer Decision Authority**.
- **Candidate Model:** `extended_catboost_58f` retains **0% Customer Decision Authority** in asynchronous Shadow Mode.
- **Canary Routing:** **DISABLED (`RISK_MODEL_CANARY_ENABLED=false`)**.
- **Promotion Status:** **BLOCKED (`promotion_allowed = false`)**.
- **Zero-Fabrication Policy:** Synthetic, fixture, or replay traffic must **never** be labeled as `LIVE_PRODUCTION`.

---

## 2. Production Service Dependency Matrix

Before initiating traffic ingress, the operator must verify that the following 4 core infrastructure services are healthy and reachable:

| Service | Component / Process | Default Port / URI | Health Check Command | Purpose |
| :--- | :--- | :--- | :--- | :--- |
| **PostgreSQL 16** | Core OLTP Relational Store | `5432` / `localhost:5432` | `pg_isready -h localhost -p 5432 -U risk_user` | Persists transactions, accounts, rules, and decision records |
| **Redis 7** | Real-time In-Memory Store | `6379` / `localhost:6379` | `redis-cli ping` (Expects `PONG`) | Token bucket rate limiting, velocity counters, and graph cache |
| **ClickHouse 24** | Real-time OLAP Telemetry Store | `9000` (Native) / `8123` (HTTP) | `curl -f http://localhost:8123/ping` (Expects `Ok.`) | Asynchronous telemetry, shadow comparison logs, and audit trails |
| **ML Serving Sidecar** | FastAPI / ONNX Runtime Engine | `8000` / `http://localhost:8000`| `curl -f http://localhost:8000/health` | Champion scoring (`/predict`) and Shadow scoring (`/predict/shadow`) |

---

## 3. Environment & Configuration Checklist

Verify the following environment variables in the Go API service configuration (`.env` or container environment):

```env
# -------------------------------------------------------------
# GOVERNANCE & MODEL CONFIGURATION (STRICTLY HARD-LOCKED)
# -------------------------------------------------------------
PORT=8080
SHADOW_SCORING_ENABLED=true
SHADOW_CANDIDATE_MODEL_VERSION=extended_catboost_58f
SHADOW_WORKER_COUNT=4
SHADOW_QUEUE_CAPACITY=1000
SHADOW_SAMPLE_RATE=1.0
RISK_MODEL_CANARY_ENABLED=false
RISK_MODEL_CANARY_PERCENT=0

# -------------------------------------------------------------
# INFRASTRUCTURE CONNECTIVITY
# -------------------------------------------------------------
DATABASE_URL=postgres://risk_user:risk_password@localhost:5432/risk_engine?sslmode=disable
REDIS_URL=redis://localhost:6379/0
CLICKHOUSE_ADDR=localhost:9000
CLICKHOUSE_DB=default
CLICKHOUSE_USER=default
CLICKHOUSE_PASSWORD=
ML_SERVICE_URL=http://localhost:8000

# -------------------------------------------------------------
# SECURITY & AUTHENTICATION
# -------------------------------------------------------------
ADMIN_API_KEY=adm_risk_super_secret_key_98765
WEBHOOK_SECRET=whsec_dummy_risk_secret_12345
```

---

## 4. End-to-End API Specification for Legitimate Merchant Ingress

### Endpoint
`POST /v1/risk-evaluations`

### Required Request Headers
| Header | Type | Example / Value | Description |
| :--- | :--- | :--- | :--- |
| `Content-Type` | String | `application/json` | Standard JSON body |
| `X-Tenant-ID` | String | `tenant_merchant_acme_01` | Multi-tenant isolation partition identifier |
| `Authorization` | String | `Bearer <jwt_token>` *(optional)* | Authenticated merchant credentials |

### Canonical Request Payload Schema
```json
{
  "transaction_id": "tx_live_prod_10001",
  "account_id": "acc_usr_99812",
  "amount": 12500,
  "currency": "USD",
  "payment_method": {
    "type": "CARD",
    "token": "tok_visa_credit_4242"
  },
  "device_fingerprint": "dev_fp_a9b8c7d6e5",
  "ip_address": "198.51.100.42"
}
```

### Production Synchronous Response Schema
```json
{
  "decision_id": "8b5d3a17-48f2-4e63-a612-9cb8c5a6104e",
  "transaction_id": "tx_live_prod_10001",
  "recommended_action": "ALLOW_RECOMMENDATION",
  "risk_score": 5,
  "reason_codes": [],
  "feature_snapshot_ref": "snap_8b5d3a17",
  "evaluated_at": "2026-08-27T07:15:00Z",
  "latency_ms": 3
}
```

---

## 5. Shadow Isolation & Telemetry Verification

1. **Provenance Tagging:**
   - Every request arriving at `POST /v1/risk-evaluations` is automatically tagged with `Provenance: "LIVE_PRODUCTION"`.
2. **Immutable Telemetry Capture:**
   - Each evaluation produces an immutable record with:
     - `evaluation_id`: UUID matching `decision_id`
     - `timestamp`: UTC transaction time
     - `amount_usd`: Decimal amount ($125.00)
     - `champion_probability` vs `candidate_calibrated_probability`
     - `champion_action` vs `candidate_action`
     - `inference_latency_ms` and `queue_wait_ms`
     - `error_category`: `NONE` (or `TIMEOUT` / `HTTP_500` / `HTTP_503`)
3. **Fault-Isolation Verification:**
   - If the CatBoost sidecar is stopped (`kill` or connection refusal), the Go API completes the checkout evaluation in $<5\text{ ms}$ with status HTTP 200 without returning any candidate error to the merchant.

---

## 6. Connectivity Smoke Test Command

The operator can execute a safe, non-sensitive connectivity smoke test using `curl`:

```bash
curl -X POST http://localhost:8080/v1/risk-evaluations \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: tenant_merchant_smoke" \
  -d '{
    "transaction_id": "tx_smoke_conn_001",
    "account_id": "acc_smoke_test_01",
    "amount": 2500,
    "currency": "USD",
    "payment_method": {
      "type": "CARD",
      "token": "tok_smoke_clean"
    },
    "device_fingerprint": "dev_smoke_fp_01",
    "ip_address": "127.0.0.1"
  }'
```

Expected HTTP Status: **`200 OK`**
Response contains: `decision_id`, `recommended_action: "ALLOW_RECOMMENDATION"` or `"STEP_UP_RECOMMENDATION"`.

---

## 7. Emergency Shutdown Procedure

If any regression occurs in production champion latency or error rates:

```bash
# Step 1: Immediately disable shadow mode via environment flag
export SHADOW_SCORING_ENABLED=false

# Step 2: Restart Go API gateway process
# System gracefully drains remaining tasks in buffer and ceases all shadow enqueue
```

---

## 8. Exact Operator Evidence Required Before Declaring Soak Active

1. **Preflight Checklist Completed:** PostgreSQL, Redis, ClickHouse, and ML sidecar verified running.
2. **Zero Canary Traffic:** Verified `RISK_MODEL_CANARY_ENABLED=false` and `RISK_MODEL_CANARY_PERCENT=0`.
3. **Genuine Merchant Ingress Flowing:** Streaming legitimate merchant transactions to `POST /v1/risk-evaluations`.
4. **Telemetry Logging Verified:** ClickHouse `shadow_evaluations` actively receiving records tagged `LIVE_PRODUCTION`.
5. **Outcome Maturation Tracking:** Maturation clock running towards the $>60$-day horizon.
