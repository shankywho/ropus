# Phase 3.5 — Payment Token ↔ Device Linkage & Card Testing Defense

**Repository:** `AI Risk Manager / Ropus`  
**Execution Date:** August 21, 2026  
**Document Version:** 1.0 (Phase 3.5 Implementation Specification)  
**Status:** COMPLETED & VERIFIED  

---

## 1. Architecture & Objective

Phase 3.5 introduces deterministic payment instrument intelligence and card testing attack detection. Operating on opaque, sanitized synthetic payment tokens (`tok_...`), it detects card testing velocity bursts, rapid token rotations, and compromised token fan-outs across physical devices without storing primary account numbers (PANs) or card credentials.

```
[ INCOMING TRANSACTION (Payment Token + Device Telemetry) ]
                           │
                           ▼
[ STEP 1: PARSE IDENTITY & SANITIZE TOKEN ]
       │  device_id = SHA256(tenant_id || ":" || canonical_fingerprint)
       │  token_id  = SHA256(tenant_id || ":" || canonical_payment_token)
       ▼
[ STEP 2: PIPELINED REDIS QUERY (Point-in-Time Safe) ]
       ├─► Device Unique Tokens (5m, 1h, 24h):  {tenant}:dev:tok5m:{dev_id}, 1h, 24h
       ├─► Device Token Tx Counts (5m, 1h, 24h): {tenant}:dev:tok_tx5m:{dev_id}, 1h, 24h
       ├─► Device Token Amount Sum (24h):       {tenant}:dev:tok_amt24h:{dev_id}
       ├─► Token Unique Devices (1h, 24h):      {tenant}:tok:dev1h:{tok_id}, 24h
       ├─► Token Tx Counts (1h, 24h):           {tenant}:tok:tx1h:{tok_id}, 24h
       └─► Specific Pair Linkage (90d):         {tenant}:dev_tok:known:{dev_id}:{tok_id}
       │   (Single Pipelined Round-Trip: p95 < 2.0ms)
       ▼
[ STEP 3: SCORE & DECIDE ]
       │  ML Inference (15-feature contract preserved) + Rules + Card Testing Risk Signals
       ▼
[ STEP 4: POST-SCORING UPDATES (Point-in-Time Safe) ]
       ├─► 1. Asynchronously Record State in Redis
       └─► 2. Upsert Durable Linkage in PostgreSQL (`device_payment_instruments`, `device_events`)
```

---

## 2. Redis Key Schema & TTL Policy

All keys are strictly tenant-isolated and keyed by internal 64-character SHA-256 identifiers (`device_id`, `token_id`). Plaintext payment tokens and raw fingerprints are never present in Redis keys.

| Redis Key Template | Redis Data Structure | Member Payload | Window / TTL | Feature Produced |
|---|---|---|---|---|
| `{tenant}:dev:tok5m:{device_id}` | Sorted Set (`ZSET`) | `<token_id>` (score = last seen ms) | 5 min / 15 min TTL | `device_unique_tokens_5m` |
| `{tenant}:dev:tok1h:{device_id}` | Sorted Set (`ZSET`) | `<token_id>` (score = last seen ms) | 1 hr / 3 hr TTL | `device_unique_tokens_1h` |
| `{tenant}:dev:tok24:{device_id}` | Sorted Set (`ZSET`) | `<token_id>` (score = last seen ms) | 24 hr / 26 hr TTL | `device_unique_tokens_24h` |
| `{tenant}:dev:tok_tx5m:{device_id}` | Sorted Set (`ZSET`) | `<timestamp>_<tx_id>` | 5 min / 15 min TTL | `device_token_tx_count_5m` |
| `{tenant}:dev:tok_tx1h:{device_id}` | Sorted Set (`ZSET`) | `<timestamp>_<tx_id>` | 1 hr / 3 hr TTL | `device_token_tx_count_1h` |
| `{tenant}:dev:tok_tx24h:{device_id}` | Sorted Set (`ZSET`) | `<timestamp>_<tx_id>` | 24 hr / 26 hr TTL | `device_token_tx_count_24h` |
| `{tenant}:dev:tok_amt24h:{device_id}` | Sorted Set (`ZSET`) | `<timestamp>_<tx_id>:<amount>` | 24 hr / 26 hr TTL | `device_token_amount_sum_24h` |
| `{tenant}:tok:dev1h:{token_id}` | Sorted Set (`ZSET`) | `<device_id>` (score = last seen ms) | 1 hr / 3 hr TTL | `token_unique_devices_1h` |
| `{tenant}:tok:dev24:{token_id}` | Sorted Set (`ZSET`) | `<device_id>` (score = last seen ms) | 24 hr / 26 hr TTL | `token_unique_devices_24h` |
| `{tenant}:tok:tx1h:{token_id}` | Sorted Set (`ZSET`) | `<timestamp>_<tx_id>` | 1 hr / 3 hr TTL | `token_tx_count_1h` |
| `{tenant}:tok:tx24h:{token_id}` | Sorted Set (`ZSET`) | `<timestamp>_<tx_id>` | 24 hr / 26 hr TTL | `token_tx_count_24h` |
| `{tenant}:dev_tok:known:{device_id}:{token_id}` | String Key | `"1"` | 90 days rolling TTL | `device_token_seen_before` |

---

## 3. Card Testing & Token Fan-Out Policy Thresholds

Configurable constants in [`backend/internal/features/payment_token.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/features/payment_token.go):

```go
type CardTestingThresholds struct {
    LowUniqueTokens5m        int64 // 3+
    SuspiciousUniqueTokens5m int64 // 5+
    SuspiciousUniqueTokens1h int64 // 8+
    HighUniqueTokens5m       int64 // 10+
    HighUniqueTokens1h       int64 // 15+
    HighTxAcrossTokens1h     int64 // 20+
}
```

- **Card Testing Signals:**
  - `NORMAL`: Standard single/dual card behavior.
  - `LOW_SIGNAL`: $\ge 3$ distinct tokens on device in 5 minutes.
  - `SUSPICIOUS`: $\ge 5$ distinct tokens in 5 minutes OR $\ge 8$ in 1 hour.
  - `HIGH_SIGNAL`: $\ge 10$ distinct tokens in 5 minutes OR $\ge 15$ in 1 hour OR $\ge 20$ tx in 1 hour.

- **Token Fan-Out Signals:**
  - `NORMAL`: Token used across 1–2 devices.
  - `SUSPICIOUS`: Single token observed across $\ge 3$ distinct devices in 1 hour OR $\ge 5$ devices in 24 hours.
  - `HIGH_SIGNAL`: Single token observed across $\ge 5$ devices in 1 hour OR $\ge 10$ devices in 24 hours.

---

## 4. Reason Codes & Risk Signals

- `DEVICE_TOKEN_NEW_RELATIONSHIP`: First time this specific payment token is used on this device (`device_token_seen_before = 0`).
- `CARD_TESTING_DEVICE_TOKEN_BURST`: Triggered when `CardTestingSignal` is `SUSPICIOUS` or `HIGH_SIGNAL`.
- `DEVICE_HIGH_TOKEN_DIVERSITY`: Device has transacted with $\ge 3$ distinct payment tokens in 24 hours.
- `DEVICE_RAPID_TOKEN_ROTATION`: Device has transacted with $\ge 3$ distinct payment tokens in 5 minutes.
- `TOKEN_MULTI_DEVICE_ACTIVITY`: Single payment token transacting across $\ge 3$ distinct devices in 24 hours.
- `TOKEN_DEVICE_FANOUT`: Triggered when `TokenFanOutSignal` is `SUSPICIOUS` or `HIGH_SIGNAL`.
- `PAYMENT_TOKEN_FEATURE_STORE_UNAVAILABLE`: Graceful degradation indicator when Redis is disconnected.

---

## 5. Security, Token Sanitization & Zero PAN Storage

- `SanitizePaymentToken`:
  - Enforces UTF-8 validity and length limit of 256 characters.
  - Rejects control characters and embedded null bytes (`\x00`).
  - Strict regex check rejects any raw 13–19 digit numeric strings (`^[0-9]{13,19}$`), guaranteeing raw PANs are never ingested or stored.
- `HashPaymentToken`: Produces tenant-isolated cryptographic SHA-256 identifier `token_id = SHA256(tenant_id + ":" + canonicalToken)`.

---

## 6. Durable PostgreSQL Integration

In `backend/internal/features/device_store.go`, the repository supports:
- `GetDevicePaymentInstruments(ctx, tenantID, deviceUUID)`
- `GetPaymentInstrumentDevices(ctx, tenantID, paymentToken)`
- `GetDevicePaymentInstrumentRelationship(ctx, tenantID, deviceUUID, paymentToken)`
- Idempotent upsert in `device_payment_instruments` within the orchestrator database transaction.

---

## 7. Performance Benchmark (200 Live Requests)

- **HTTP End-to-End Latency:**
  - p50: **`3.70ms`**
  - p95: **`5.75ms`**
  - p99: **`8.10ms`**
- **Server Internal Pipeline Latency:**
  - p50: **`1.00ms`**
  - p95: **`2.00ms`**
  - p99: **`3.00ms`**

---

## 8. What is Intentionally Deferred to Phase 3.6+

- **Phase 3.6:** Multi-Window Device Velocity & Anomaly Bursts.
- **Phase 3.7:** Device Reputation & Dispute Attribution.
- **Phase 3.8:** 25-Feature ML Feature Contract Expansion.
- **Phase 3.9 – 3.13:** XGBoost Model Retraining, Beta Calibration Re-evaluation, and Staged Rollout.
