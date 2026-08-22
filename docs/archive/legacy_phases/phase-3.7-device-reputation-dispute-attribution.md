# Phase 3.7 — Device Reputation & Dispute Attribution

**Repository:** `AI Risk Manager / Ropus`  
**Execution Date:** August 21, 2026  
**Document Version:** 1.0 (Phase 3.7 Implementation Specification)  
**Status:** COMPLETED & VERIFIED  

---

## 1. Architecture & Reputation Topology

Phase 3.7 introduces deterministic device trust and dispute/fraud attribution. It enables the risk engine to answer point-in-time questions regarding a device's historical incident record without invoking external ML models.

```
[ INCOMING TRANSACTION (Amount + Device + Account + Token) ]
                               │
                               ▼
[ STEP 1: PARSE IDENTITY & SANITIZE TELEMETRY ]
       │  device_id = SHA256(tenant_id || ":" || canonical_fingerprint)
       ▼
[ STEP 2: PIPELINED REDIS QUERY (Point-in-Time Safe) ]
       ├─► Device Velocity Ledger:    `{tenant}:vel:dev:events:{device_id}`
       ├─► Account/Device Graph:       `{tenant}:dev:acc1h:{device_id}`, etc.
       ├─► Payment Token Linkage:      `{tenant}:dev:tok5m:{device_id}`, etc.
       └─► Device Reputation Ledger:  `{tenant}:rep:dev:events:{device_id}` (ZSET, 90d TTL)
           & First Seen Key:           `{tenant}:rep:dev:first_seen:{device_id}` (STRING, 365d TTL)
       ▼
[ STEP 3: SCORE & DECIDE ]
       │  ML Inference (15-feature contract preserved) + Rules + Reputation Risk Signals
       ▼
[ STEP 4: POST-SCORING UPDATES (Point-in-Time Safe) ]
       ├─► Asynchronously Append Current Tx to `{tenant}:rep:dev:events:{device_id}` (SUCCESS)
       └─► Upsert Durable Event in PostgreSQL (`device_reputation`, `device_events`)
```

---

## 2. Reputation Features & Metrics

All features evaluate state *prior* to recording the incoming transaction:

| Feature Name | Type | Description | Point-in-Time Safe | Tested |
|---|---|---|:---:|:---:|
| `device_total_transactions` | `int64` | Total historical transactions on this device | YES | YES |
| `device_successful_transactions` | `int64` | Total successful transactions | YES | YES |
| `device_failed_transactions` | `int64` | Total failed transactions | YES | YES |
| `device_disputed_transactions` | `int64` | Number of historical disputes linked to device | YES | YES |
| `device_fraud_transactions` | `int64` | Confirmed fraudulent transactions linked to device | YES | YES |
| `device_refunded_transactions` | `int64` | Refunded transactions | YES | YES |
| `device_chargeback_count` | `int64` | Chargeback transactions | YES | YES |
| `device_dispute_rate` | `float64` | $\text{Disputes} / \max(\text{TotalTx}, 1)$ in $[0.0, 1.0]$ | YES | YES |
| `device_fraud_rate` | `float64` | $\text{Fraud} / \max(\text{TotalTx}, 1)$ in $[0.0, 1.0]$ | YES | YES |
| `device_refund_rate` | `float64` | $\text{Refunds} / \max(\text{TotalTx}, 1)$ in $[0.0, 1.0]$ | YES | YES |
| `device_success_rate` | `float64` | $\text{Successful} / \max(\text{TotalTx}, 1)$ in $[0.0, 1.0]$ | YES | YES |
| `device_recent_dispute_count` | `int64` | Disputes within the last 30 days | YES | YES |
| `device_recent_fraud_count` | `int64` | Confirmed fraud within the last 30 days | YES | YES |
| `device_recent_chargeback_count` | `int64` | Chargebacks within the last 30 days | YES | YES |
| `device_days_since_first_seen` | `float64` | Days elapsed since device first recorded | YES | YES |
| `device_days_since_last_dispute` | `float64` | Days elapsed since last dispute ($-1.0$ if none) | YES | YES |
| `device_days_since_last_fraud` | `float64` | Days elapsed since last fraud ($-1.0$ if none) | YES | YES |
| `device_reputation_score` | `float64` | Bounded trust/risk score in $[0.0, 1.0]$ ($0.0=$ Trusted, $1.0=$ Risky) | YES | YES |

---

## 3. Deterministic Reputation Scoring Formula

Reputation scoring operates on a continuous, bounded $[0.0, 1.0]$ scale:

- **Baseline Neutral Score:** $0.50$ (assigned to brand-new devices with zero history).
- **Trust Accumulation (Discounts):**
  - Successful transaction volume discount: $-\min(0.30, \text{SuccessfulTx} \times 0.03)$ (up to $-0.30$ at 10+ successful tx).
  - Clean tenure discount: $-\min(0.15, \text{DaysSinceFirstSeen} \times 0.005)$ (up to $-0.15$ at 30+ days without incidents).
- **Risk Penalties:**
  - Base confirmed fraud penalty: $+0.50$.
  - Recent confirmed fraud penalty ($< 30\text{ days}$): $+0.30$ additional.
  - Dispute / chargeback penalty: $+\min(0.50, \text{DisputeCount} \times 0.25)$.
  - Recent dispute penalty ($< 30\text{ days}$): $+\min(0.20, \text{RecentDisputes} \times 0.10)$.
  - Repeated failure penalty: $+\min(0.20, \text{FailedTx} \times 0.05)$.
  - Elevated dispute rate penalty: $+0.20$ if $\text{TotalTx} \ge 3$ and $\text{DisputeRate} \ge 0.10$.
- **Boundary Clamping:** Clamped strictly to $[0.0, 1.0]$.

---

## 4. Reason Codes & Precedence

- **High Severity:**
  - `DEVICE_FRAUD_HISTORY`: Device has $\ge 1$ confirmed fraud in its lifetime.
  - `DEVICE_RECENT_FRAUD_ACTIVITY`: Confirmed fraud occurred within the last 30 days.
  - `DEVICE_BAD_REPUTATION`: Overall `DeviceReputationScore >= 0.75`.
- **Medium Severity:**
  - `DEVICE_HIGH_FRAUD_RATE`: Lifetime or recent fraud rate $\ge 5\%$.
  - `DEVICE_HIGH_DISPUTE_RATE`: Lifetime or recent dispute rate $\ge 10\%$.
  - `DEVICE_RECENT_DISPUTE_BURST`: $\ge 2$ disputes within the last 30 days.
  - `DEVICE_DISPUTE_HISTORY`: Device has $\ge 1$ dispute in its history.
- **Trust Signal:**
  - `DEVICE_LONG_TRUSTED_HISTORY`: Tenure $\ge 14$ days, $\ge 5$ successful transactions, 0 disputes, 0 fraud (`ReputationScore <= 0.20`).
- **Reliability:**
  - `DEVICE_REPUTATION_FEATURE_STORE_UNAVAILABLE`: When degraded.

---

## 5. Security & Idempotency Guarantees

1. **Idempotent Ingestion:** Duplicate outcome events (e.g. re-ingesting a webhook for `DISPUTE:tx_123`) do not inflate dispute or fraud counts. Redis Sorted Set members are formatted as `<OUTCOME_TYPE>:<transaction_id>`, ensuring atomic deduplication.
2. **Tenant Isolation:** All reputation Redis keys are scoped by `{tenant_id}` (`{tenant}:rep:dev:events:{device_id}`).
3. **Key Privacy:** Uses internal 64-character SHA-256 `device_id`. Raw fingerprints, PANs, and sensitive user data are never written to Redis keys.
4. **Graceful Degradation:** Disconnected Redis returns safe neutral defaults ($0.50$ score, 0 counts) with `DEVICE_REPUTATION_FEATURE_STORE_UNAVAILABLE` without failing the transaction evaluation.

---

## 6. Performance Benchmark (200 Live Requests)

- **HTTP End-to-End Latency:**
  - p50: **`8.65ms`**
  - p95: **`12.50ms`**
  - p99: **`21.45ms`**
- **Server Internal Decision Pipeline Latency:**
  - p50: **`3.00ms`**
  - p95: **`4.00ms`**
  - p99: **`5.05ms`**

---

## 7. What is Intentionally Deferred to Phase 3.8+

- **Phase 3.8:** 25-Feature ML Feature Contract Expansion (combining identity, velocity, graph, token, and reputation features).
- **Phase 3.9:** XGBoost Retraining on canonical datasets.
- **Phase 3.10:** Beta Calibration Re-evaluation.
- **Phase 3.11 – 3.13:** Shadow Scoring & Staged Production Rollout.
