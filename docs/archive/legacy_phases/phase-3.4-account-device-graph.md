# Phase 3.4 — Account ↔ Device Graph & Multi-Accounting Detection

**Repository:** `AI Risk Manager / Ropus`  
**Execution Date:** August 21, 2026  
**Document Version:** 1.0 (Phase 3.4 Implementation Specification)  
**Status:** COMPLETED & VERIFIED  

---

## 1. Architecture Overview

Phase 3.4 introduces a deterministic, low-latency, and point-in-time safe Account ↔ Device Relationship Graph. Backed by Redis 7 for synchronous real-time scoring and PostgreSQL 16 `device_accounts` for durable relational truth, it detects complex relationship patterns (multi-accounting clusters, rapid account switching, new accounts on established devices, and account fan-outs) without adding heavy external graph dependencies (such as Neo4j or Graph Neural Networks).

```
[ INCOMING TRANSACTION (Account ID + Device Telemetry) ]
                           │
                           ▼
[ STEP 1: PARSE IDENTITY & SANITIZE ACCOUNT ]
       │  device_id = SHA256(tenant_id || ":" || canonical_fingerprint)
       │  account_id = SanitizeAccountID(raw_account_id)
       ▼
[ STEP 2: PIPELINED REDIS GRAPH QUERY (Point-in-Time Safe) ]
       ├─► Device Unique Accounts (1h, 24h):  {tenant}:dev:acc1h:{dev_id}, {tenant}:dev:acc24:{dev_id}
       ├─► Account Switches (1h, 24h):        {tenant}:dev:acc_seq:{dev_id}
       ├─► Account Unique Devices (1h, 24h):  {tenant}:acc:dev1h:{acc_id}, {tenant}:acc:dev24:{acc_id}
       ├─► Specific Pair Linkage:             {tenant}:dev_acc:known:{dev_id}:{acc_id}
       └─► Established Device Status:         {tenant}:dev:known:{dev_id}
       │   (Single Pipelined Round-Trip: p95 < 2.0ms)
       ▼
[ STEP 3: SCORE & DECIDE ]
       │  ML Inference (15-feature contract preserved) + Rules + Multi-Account Reason Codes
       ▼
[ STEP 4: POST-SCORING UPDATES (Point-in-Time Safe) ]
       ├─► 1. Asynchronously Record Graph State in Redis
       └─► 2. Upsert Durable Linkage in PostgreSQL (`device_accounts`, `device_events`, `outbox_events`)
```

---

## 2. Redis Key Schema & TTL Policy

All keys are strictly tenant-isolated and keyed by the internal 64-character SHA-256 `device_id` and sanitized `account_id`. Raw client fingerprints are never used in Redis keys.

| Redis Key Template | Redis Data Structure | Member Payload | Window / TTL | Feature Produced |
|---|---|---|---|---|
| `{tenant}:dev:acc1h:{device_id}` | Sorted Set (`ZSET`) | `<account_id>` (score = last seen ms) | 1 hr / 3 hr TTL | `device_unique_accounts_1h` |
| `{tenant}:dev:acc24:{device_id}` | Sorted Set (`ZSET`) | `<account_id>` (score = last seen ms) | 24 hr / 26 hr TTL | `device_unique_accounts_24h` |
| `{tenant}:dev:acc_seq:{device_id}` | Sorted Set (`ZSET`) | `<nano>_<tx_id>:<account_id>` | 24 hr / 26 hr TTL | `device_account_switches_1h`<br>`device_account_switches_24h` |
| `{tenant}:acc:dev1h:{account_id}` | Sorted Set (`ZSET`) | `<device_id>` (score = last seen ms) | 1 hr / 3 hr TTL | `account_unique_devices_1h`<br>`account_new_device_1h` |
| `{tenant}:acc:dev24:{account_id}` | Sorted Set (`ZSET`) | `<device_id>` (score = last seen ms) | 24 hr / 26 hr TTL | `account_unique_devices_24h` |
| `{tenant}:dev_acc:known:{device_id}:{account_id}` | String Key | `"1"` | 90 days rolling TTL | `device_account_seen_before` |

---

## 3. Account Switch Detection Algorithm

An account switch is defined by consecutive transitions between distinct accounts on a physical device ($A_i \ne A_{i-1}$):

- Sequence: $\text{Alice} \to \text{Alice} \to \text{Bob} \to \text{Charlie} \to \text{Alice}$
- Transitions evaluated:
  1. $\text{Alice} \to \text{Alice}$ (Same account $\implies$ 0 switches)
  2. $\text{Alice} \to \text{Bob}$ (Distinct account $\implies$ Switch 1)
  3. $\text{Bob} \to \text{Charlie}$ (Distinct account $\implies$ Switch 2)
  4. $\text{Charlie} \to \text{Alice}$ (Distinct account $\implies$ Switch 3)
- Total switches: **3**. Repeated same-account transactions never artificially inflate the switch count.

---

## 4. Multi-Accounting Risk Signals & Reason Codes

The graph engine produces deterministic risk signals:

- `ACCOUNT_NEW_TO_DEVICE`: First time this specific account has transacted on this device (`device_account_seen_before = 0`).
- `ACCOUNT_NEW_TO_ESTABLISHED_DEVICE`: Known, established device (`device_seen_before = 1`) encountering a novel account.
- `DEVICE_MULTI_ACCOUNT_ACTIVITY`: Device has transacted with $\ge 3$ distinct accounts in 24 hours.
- `DEVICE_ACCOUNT_SWITCH_BURST`: Device has experienced $\ge 3$ account switches in 1 hour or $\ge 5$ in 24 hours.
- `ACCOUNT_MULTI_DEVICE_ACTIVITY`: Account has transacted from $\ge 3$ distinct devices in 24 hours.
- `ACCOUNT_DEVICE_FANOUT`: Account has transacted from $\ge 5$ distinct devices in 24 hours.
- `ACCOUNT_DEVICE_GRAPH_UNAVAILABLE`: Redis feature store unavailable, system degraded gracefully.

---

## 5. Point-in-Time Safety & Current-Event Leakage Prevention

1. **Pre-Scoring Retrieval:** `GetGraphFeatures` reads existing relationship state *prior* to evaluation.
2. **Deterministic Evaluation:** Feature values reflect true historical priors without data leakage from the current transaction.
3. **Post-Scoring Update:** `RecordGraphTransaction` updates Redis sets and PostgreSQL linkages only *after* risk decisioning completes.

---

## 6. Durable PostgreSQL Integration

In `backend/internal/features/device_store.go`, the durable relational schema supports querying and transactional updates:
- `LinkAccount(ctx, tenantID, deviceUUID, accountID)` (Idempotent upsert with counter increment)
- `GetDeviceAccounts(ctx, tenantID, deviceUUID)`
- `GetAccountDevices(ctx, tenantID, accountID)`
- `GetAccountDeviceRelationship(ctx, tenantID, deviceUUID, accountID)`

---

## 7. Targeted Test Results

All 35 feature tests passed:
- `TestAccountDeviceGraph_Comprehensive` (9 suites in [`account_device_test.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/features/account_device_test.go)):
  1. *First account/device relationship & point-in-time safety*
  2. *New account on established device*
  3. *Account switching & repeat transactions do not inflate switch count*
  4. *Account fan-out across multiple devices*
  5. *Multi-accounting high-signal clustering ($\ge 10$ accounts)*
  6. *Tenant isolation (Tenant A vs Tenant B)*
  7. *Account ID sanitization & validation (control chars / null bytes / length limits)*
  8. *Concurrency: 100 simultaneous updates on same device*
  9. *Graceful degradation on nil / disconnected store*

---

## 8. Performance Benchmark

Across 200 live synchronous requests against `http://localhost:8080/v1/risk-evaluations`:
- **HTTP End-to-End Latency:**
  - p50: **`3.65ms`**
  - p95: **`5.90ms`**
  - p99: **`11.20ms`**
- **Server Internal Decision Pipeline Latency:**
  - p50: **`1.00ms`**
  - p95: **`2.00ms`**
  - p99: **`3.00ms`**

---

## 9. What is Intentionally Deferred to Phase 3.5+

- **Phase 3.5:** Payment Token ↔ Device Linkage & Card Testing Defense.
- **Phase 3.6:** Multi-Window Device Velocity & Anomaly Bursts.
- **Phase 3.7:** Device Reputation & Dispute Attribution.
- **Phase 3.8:** 25-Feature ML Feature Contract Expansion.
- **Phase 3.9 – 3.13:** Model Retraining, Beta Calibration Re-evaluation, and Staged Rollout.
