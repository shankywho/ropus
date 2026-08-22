# Phase 3.3 — Redis Real-Time Device Feature Store & Sliding Windows

**Repository:** `AI Risk Manager / Ropus`  
**Execution Date:** August 21, 2026  
**Document Version:** 1.0 (Phase 3.3 Implementation Specification)  
**Status:** COMPLETED & VERIFIED  

---

## 1. Architecture Overview

Phase 3.3 introduces the low-latency, real-time device feature store backed by Redis 7. It complements the durable PostgreSQL relational ledger (Phase 3.2) by computing point-in-time sliding-window device metrics during synchronous transaction evaluation:

```
[ INCOMING CHECKOUT TRANSACTION ]
       │
       ▼
[ STEP 1: PARSE IDENTITY ]
       │  device_id = SHA256(tenant_id || ":" || canonical_fingerprint)
       ▼
[ STEP 2: PIPELINED REDIS QUERY (Point-in-Time Safe) ]
       ├─► 1m Device Velocity:       {tenant}:vel:dev:1m:{device_id}
       ├─► 1h Device Velocity:       {tenant}:vel:dev:1h:{device_id}
       ├─► 24h Device Velocity:      {tenant}:vel:dev:24h:{device_id}
       ├─► 24h Amount Sum:           {tenant}:vel:dev:24h:{device_id} (Parsed amounts)
       ├─► 24h Distinct Accounts:    {tenant}:dev:acc24:{device_id}
       ├─► 24h Distinct Card Tokens: {tenant}:dev:tok24:{device_id}
       └─► 90d Device Novelty:       {tenant}:dev:known:{device_id}
       │   (Single Pipelined Round-Trip: p95 < 2.5ms)
       ▼
[ STEP 3: SCORE & DECIDE ]
       │  ML Inference (15 features) + Rules + Cost Policy
       ▼
[ STEP 4: POST-SCORING UPDATES ]
       ├─► 1. Asynchronously Record Transaction in Redis Sliding Windows
       └─► 2. Persist Durable Record in PostgreSQL (devices, device_events, outbox_events)
```

---

## 2. Redis Key Schema & TTL Policy

All keys are strictly tenant-isolated and keyed by the internal 64-character SHA-256 `device_id` hash. Raw client visitor IDs are never used as keys.

| Redis Key Template | Redis Data Structure | Member Payload | Window / TTL | Feature Produced |
|---|---|---|---|---|
| `{tenant}:vel:dev:1m:{device_id}` | Sorted Set (`ZSET`) | `<nano>_<tx_id>` | 1 min / 5 min TTL | `device_tx_count_1m` |
| `{tenant}:vel:dev:1h:{device_id}` | Sorted Set (`ZSET`) | `<nano>_<tx_id>` | 1 hr / 3 hr TTL | `device_tx_count_1h` |
| `{tenant}:vel:dev:24h:{device_id}` | Sorted Set (`ZSET`) | `<nano>_<tx_id>:<amount>` | 24 hr / 26 hr TTL | `device_tx_count_24h`<br>`device_amount_sum_24h` |
| `{tenant}:dev:acc24:{device_id}` | Sorted Set (`ZSET`) | `<account_id>` (score = last seen ms) | 24 hr / 26 hr TTL | `device_unique_accounts_24h` |
| `{tenant}:dev:tok24:{device_id}` | Sorted Set (`ZSET`) | `<payment_token>` (score = last seen ms) | 24 hr / 26 hr TTL | `device_unique_tokens_24h` |
| `{tenant}:dev:known:{device_id}` | String Key | `"1"` | 90 days rolling TTL | `device_seen_before` (0 or 1) |

---

## 3. Sliding Window Semantics & Mathematical Soundness

1. **Transaction Velocity:**
   $$\text{TxCount}(W) = \text{ZCARD}\left(\{ e \in \text{ZSET} \mid \text{score}(e) \ge \text{now} - W \}\right)$$
2. **Rolling Amount Sum (Zero Floating-Point Drift):**
   $$\text{AmountSum}_{24\text{h}} = \sum_{e \in \text{ZSET}_{24\text{h}}} \text{ParseInt}(\text{amount}(e))$$
   All sums are calculated using 64-bit integer arithmetic in minor currency units (paise/cents), avoiding float accumulation bugs.
3. **Distinct Entity Tracking:**
   Using sorted sets with score = epoch timestamp of last interaction allows `ZREMRANGEBYSCORE key -inf (now - 24h)` to atomically prune inactive accounts/tokens before calling `ZCARD`.

---

## 4. Point-in-Time Safety & Current-Event Leakage Prevention

To guarantee that ML model inference and policy scoring reflect true historical priors without data leakage from the current transaction:
1. `GetDeviceFeatures` queries the feature store **prior** to scoring.
2. The ML model receives `IsNewDevice = 1` if `device_seen_before == 0`, and `0` if `device_seen_before == 1`.
3. `RecordDeviceTransaction` executes **after** scoring, updating the sorted sets and setting the 90-day novelty marker.

---

## 5. Graceful Degradation & Failure Mode

If Redis becomes unreachable or times out:
- The orchestrator catches the error, sets `is_degraded = true`, and appends `DEVICE_FEATURE_STORE_UNAVAILABLE` to `reason_codes`.
- Deterministic safe zeros are returned for all 7 metrics without crashing HTTP evaluation.
- PostgreSQL transactions and rule evaluation continue unabated.

---

## 6. Observability & Feature Snapshot Schema

Evaluated responses and outbox events include structured device intelligence:

```json
{
  "device_features": {
    "device_tx_count_1m": 1,
    "device_tx_count_1h": 3,
    "device_tx_count_24h": 5,
    "device_amount_sum_24h": 47500,
    "device_unique_accounts_24h": 2,
    "device_unique_tokens_24h": 2,
    "device_seen_before": 1
  },
  "device_feature_store": {
    "source": "redis",
    "degraded": false
  }
}
```

---

## 7. Targeted Test Results

All 26 feature tests in [`backend/internal/features/device_redis_test.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/features/device_redis_test.go) and [`device_store_test.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/features/device_store_test.go) passed:
1. *First-seen detection (`device_seen_before = 0` prior to write)*
2. *Repeat-seen detection (`device_seen_before = 1` after write)*
3. *Rolling 24h amount sum & distinct entity count*
4. *Sliding window expiration across 1m, 1h, and 24h boundaries*
5. *Tenant isolation (Tenant A vs Tenant B)*
6. *100-worker high-concurrency race safety*
7. *Privacy: Raw fingerprint never present in Redis keys*
8. *Graceful degradation on nil or disconnected Redis client*

---

## 8. What is Intentionally Deferred to Phase 3.4+

- **Phase 3.4:** Account ↔ Device Graph & Multi-Accounting Detection.
- **Phase 3.5:** Payment Token ↔ Device Linkage & Card Testing Defense.
- **Phase 3.6:** Multi-Window Velocity & Anomaly Bursts.
- **Phase 3.7:** Device Reputation & Chargeback Dispute Attribution.
- **Phase 3.8 – 3.10:** ML 25-Feature Expansion, Model Retraining, and Staged Rollout.
