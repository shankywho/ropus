# Phase 3.14 — Production Observability & Operational Control

**Repository:** `AI Risk Manager / Ropus`  
**Execution Date:** August 21, 2026  
**Document Version:** 1.0 (Phase 3.14 Production Telemetry & Operator Handbook)  
**Production Readiness:** READY FOR PRODUCTION  

---

## 1. Executive Summary & Architecture

Phase 3.14 delivers the operational control plane and deep telemetry suite for the AI Risk Manager platform. It introduces clean separation between container liveness and dependency readiness, exposes operator-level model and system diagnostic APIs, hardens administrative controls against operational errors, standardizes structured JSON logging (with strict zero-PII leakage guarantees), and provides a comprehensive ClickHouse OLAP analytics query library.

---

## 2. API Endpoints & Operational Interfaces

### 2.1 Diagnostics & Probes

| Endpoint | Method | Purpose | Auth Required | Fail-Safe Behavior |
|:---|:---:|:---|:---:|:---|
| `/health` | `GET` | Lightweight container liveness probe (checks process alive, uptime, versions) | No | Fast HTTP 200 without blocking queries |
| `/readiness` | `GET` | Dependency readiness probe (PostgreSQL, Redis, ClickHouse, ML sidecar) | No | HTTP 503 `NOT_READY` if critical database is disconnected |
| `/v1/models/status` | `GET` | Comprehensive status of primary 25F and fallback 15F models | No | Exposes contract versions, formats, thresholds, and sidecar URL |
| `/v1/system/status` | `GET` | Single-call operational snapshot for site reliability engineers | No | Returns `status: HEALTHY/DEGRADED`, circuit breaker state, and dependency health |
| `/v1/canary/status` | `GET` | Live rollout status, actual vs target traffic %, and latency distributions | No | Returns p50, p95, p99 latencies, error/fallback rates, and safety status |
| `/v1/canary/control` | `POST` | Authenticated dynamic rollout percentage adjustment | **Admin API Key** | Rejects missing reasons, rejects unauthorized actors, enforces cooldown rules |

---

## 3. Operator Safety & Controls (`POST /v1/canary/control`)

### Safety Hardening Rules
1. **Mandatory Non-Empty Reason:** Rollout changes without an explicit `reason` string return `400 Bad Request`.
2. **Circuit Breaker Cooldown Enforcement:** If the circuit breaker tripped to `ROLLED_BACK`, enabling canary traffic requires explicit confirmation keyword (`"override"` or `"reset"`) in the reason string.
3. **Actor Identification:** Captures `X-Admin-User` header for audit attribution.
4. **ClickHouse Audit Trail:** Every invocation emits a structured record to `canary_rollout_events`.
5. **Constant-Time Comparison:** Evaluated using `crypto/subtle.ConstantTimeCompare` against `ADMIN_API_KEY`.

---

## 4. ClickHouse Production Analytics Queries

The following queries can be executed directly in ClickHouse to monitor production operational health.

### Query 1: Candidate Error Rate & Fallback Rate (Last 1 Hour)
```sql
SELECT
    count() AS total_evaluations,
    countIf(model_route = 'CANDIDATE') AS candidate_evaluations,
    countIf(fallback_used = 1) AS fallback_count,
    round(countIf(fallback_used = 1) / nullIf(countIf(model_route = 'CANDIDATE'), 0) * 100, 2) AS candidate_fallback_rate_pct,
    countIf(error != '') AS error_count,
    round(countIf(error != '') / nullIf(countIf(model_route = 'CANDIDATE'), 0) * 100, 2) AS candidate_error_rate_pct
FROM canary_rollout_evaluations
WHERE timestamp >= now() - INTERVAL 1 HOUR;
```

### Query 2: Candidate Latency Percentiles (p50, p95, p99)
```sql
SELECT
    round(quantile(0.50)(candidate_latency_ms), 2) AS p50_latency_ms,
    round(quantile(0.95)(candidate_latency_ms), 2) AS p95_latency_ms,
    round(quantile(0.99)(candidate_latency_ms), 2) AS p99_latency_ms,
    round(avg(candidate_latency_ms), 2) AS avg_latency_ms,
    round(max(candidate_latency_ms), 2) AS max_latency_ms
FROM canary_rollout_evaluations
WHERE model_route = 'CANDIDATE'
  AND candidate_latency_ms > 0
  AND timestamp >= now() - INTERVAL 1 HOUR;
```

### Query 3: Decision Divergence Rate (Candidate vs Legacy Production)
```sql
SELECT
    count() AS total_shadow_comparisons,
    countIf(decision_divergence = 1) AS decision_changes,
    round(countIf(decision_divergence = 1) / count() * 100, 2) AS decision_divergence_pct,
    round(avg(score_delta), 4) AS avg_score_delta,
    round(avg(abs_score_delta), 4) AS avg_abs_score_delta
FROM shadow_score_evaluations
WHERE timestamp >= now() - INTERVAL 1 HOUR;
```

### Query 4: Complete Rollout & Circuit Breaker Audit History
```sql
SELECT
    event_id,
    timestamp,
    event_type,
    previous_percentage,
    new_percentage,
    trigger,
    safety_status,
    actor,
    reason
FROM canary_rollout_events
ORDER BY timestamp DESC
LIMIT 25;
```

---

## 5. Structured Logging & Security Auditing

All risk evaluations emit structured JSON log entries:
```json
{
  "event": "risk_evaluation_completed",
  "transaction_id": "tx_uuid_obs_001",
  "tenant_id": "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
  "decision_id": "dec_e90fb89f-32ed-421a-b28b-3ef7ea406461",
  "model_version": "fraud-xgb-25f-v3.0",
  "model_route": "CANDIDATE",
  "decision": "ALLOW_RECOMMENDATION",
  "risk_score": 4,
  "latency_ms": 9,
  "fallback_used": 0,
  "is_degraded": false,
  "canary_percentage": 100,
  "circuit_breaker_state": "HEALTHY",
  "safety_gate_status": "HEALTHY",
  "timestamp": "2026-08-21T11:50:29Z"
}
```
**Security Guarantee:** Raw credit card numbers, payment tokens, plaintext IP addresses, device user-agents, and authentication headers are never written to disk or container standard output.
