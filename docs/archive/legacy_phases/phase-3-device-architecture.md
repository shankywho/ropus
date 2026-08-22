# Phase 3 — Device Intelligence Architecture & Pipeline Design

**Document Version:** 1.0  
**Architectural Goal:** Real-Time Device Identity, Velocity, Entity Linkage, and Reputation Engine  

---

## 1. Architectural Blueprint & Layer Separation

To prevent the device intelligence subsystem from devolving into an unmaintainable collection of ad-hoc heuristics, Ropus enforces strict separation between identity ingestion, historical state, feature extraction, machine learning, and cost policy:

```
                        1. INCOMING TRANSACTION PAYLOAD
                                      │
                                      ▼
                        2. DEVICE IDENTITY SANITIZATION
                  • Format validation (length, character set)
                  • Tenant-isolated pseudonymization:
                    device_id = SHA256(tenant_id || ":" || raw_fingerprint)
                                      │
                                      ▼
                        3. REDIS REAL-TIME FEATURE STORE (Synchronous < 5ms)
                  ┌───────────────────┼───────────────────┐
                  ▼                   ▼                   ▼
          DEVICE VELOCITY      ENTITY LINKAGE      DEVICE NOVELTY
       • 1m, 1h, 24h Txns   • Distinct Accounts • First-Seen Timestamp
       • 1h, 24h Amount     • Distinct Tokens   • Age in Days / Hours
                  │                   │                   │
                  └───────────────────┼───────────────────┘
                                      ▼
                        4. CANONICAL FEATURE VECTOR (25 Features)
                                      │
                                      ▼
                        5. ONNX MACHINE LEARNING SCORING (50ms Deadline)
                                      │
                                      ▼
                        6. BETA PROBABILITY CALIBRATION
                                      │
                                      ▼
                        7. COST-SENSITIVE DECISION ENGINE
                  • Evaluates E[Cost(ALLOW)] vs E[Cost(REVIEW)] vs E[Cost(DECLINE)]
                  • Enforces 5-10% Manual Review Queue Capacity
                                      │
                                      ▼
                        8. SYNCHRONOUS RESPONSE TO CALLER (< 15ms Total Latency)
                                      │
                        9. ASYNCHRONOUS DURABLE LEDGER & CDC (Background)
                                      ▼
                  ┌───────────────────┴───────────────────┐
                  ▼                                       ▼
          PostgreSQL Core Ledger                  Transactional Outbox
          • devices                               • Debezium CDC Pipeline
          • device_accounts                       • Redpanda / Kafka
          • device_payment_instruments            • ClickHouse Long-Term Graph
          • device_reputation
```

---

## 2. Synchronous Path vs Asynchronous Path Isolation

| Processing Pipeline | SLA Budget | Storage Engine | Operations Executed |
| :--- | :---: | :--- | :--- |
| **Synchronous Fast Path** | **$< 15\text{ms}$** | **Redis 7 (In-Memory)** | 1. Calculate device ID hash.<br>2. Pipeline Redis queries for device velocity, account count, and novelty.<br>3. Construct ML feature vector.<br>4. Execute ONNX inference & Beta calibration.<br>5. Compute cost policy action.<br>6. Record Redis sliding-window event. |
| **Asynchronous Ledger & Graph Path** | **Background ($< 500\text{ms}$)** | **PostgreSQL 16 & ClickHouse** | 1. Persist encrypted decision in `risk_decisions`.<br>2. Insert/update `devices`, `device_accounts`, and `device_payment_instruments`.<br>3. Emit Debezium outbox event to Redpanda.<br>4. Ingest into ClickHouse for historical graph clustering. |

---

## 3. Real-Time Identity Resolution & State Machine

Every incoming transaction transitions through a deterministic device state machine:

```
                          [ Incoming Transaction ]
                                     │
                                     ▼
                    [ Does Device Exist in Redis/DB? ]
                                   /   \
                             NO   /     \   YES
                                 /       \
                                ▼         ▼
                    [ STATE: FIRST_SEEN ] [ Query Last Seen Timestamp ]
                            │                     │
                            │             ┌───────┴───────┐
                            │      < 30 Days              >= 30 Days
                            │             │                       │
                            │             ▼                       ▼
                            │    [ STATE: KNOWN_ACTIVE ] [ STATE: REAPPEARED ]
                            │             │                       │
                            └─────────────┼───────────────────────┘
                                          ▼
                         [ Check Device-Account Linkage ]
                                   /             \
                   Known Account  /               \ Novel Account for Device
                                 /                 \
                                ▼                   ▼
                   [ Single-User Device ]  [ Multi-Account Shared Device ]
                                                    (Fraud farm or Family)
```
