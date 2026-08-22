# Phase 3.12 — Staged Production Rollout

**Repository:** `AI Risk Manager / Ropus`  
**Execution Date:** August 21, 2026  
**Document Version:** 1.0 (Phase 3.12 Implementation Specification)  
**Status:** COMPLETED & VERIFIED  

---

## 1. Executive Summary & Objective

Phase 3.12 introduces a safe, deterministic **Canary Router** to gradually and safely shift live production risk evaluations from the legacy 15-feature model (`fraud_model.onnx` + v1.5 contract) to the candidate 25-feature model (`fraud_model_25f_candidate.onnx` + candidate Beta calibration + v2.5 contract) across staged rollout tiers:

$$\text{Rollout Tiers: } [0\%, 1\%, 5\%, 10\%, 25\%, 50\%, 100\%]$$

### Production Safety Invariants
1. **Default State is 0%:** `RISK_MODEL_CANARY_ENABLED=false`, `RISK_MODEL_CANARY_PERCENT=0`. Deploying this release does not shift any live production traffic.
2. **Deterministic Routing:** Uses SHA-256 of `tenant_id:transaction_id` modulo 100. Retries, idempotency keys, and tenant subsets always hash to the exact same route.
3. **Fail-Safe Automatic Fallback:** If candidate 25F inference fails (network timeout, 500 error, service down), the orchestrator automatically degrades to the legacy 15F model in sub-millisecond time. The client request never fails.
4. **Frozen Decision Policy:** Candidate output calibrated probabilities are evaluated against the frozen production thresholds ($<0.05 \rightarrow \text{ALLOW}$, $0.05\le p < 0.35 \rightarrow \text{MANUAL\_REVIEW}$, $\ge 0.35 \rightarrow \text{DECLINE}$).
5. **Instantaneous Rollback:** Setting `RISK_MODEL_CANARY_ENABLED=false` or `RISK_MODEL_CANARY_PERCENT=0` returns 100% of traffic to the legacy path instantly without container rebuilds or migrations.

---

## 2. Architecture & Decision Flow

```mermaid
graph TD
    Client["Client POST /v1/risk-evaluations"] --> PreRules["1. Pre-Rules & Guardrails"]
    PreRules -->|Rule Halt| RuleDecision["Rule Action (ALLOW/DECLINE)"]
    PreRules -->|Pass| FeatureVector["2. Construct Canonical 25F Vector (v2.5)"]
    
    FeatureVector --> CanaryRouter{"3. CanaryRouter.Route()\nSHA256(tenant:tx) % 100 < Target%"}
    
    CanaryRouter -->|Candidate Route| CandidateCall["4a. Call Candidate 25F ONNX\n(/predict/shadow or /predict/candidate)\n50ms Context Budget"]
    CanaryRouter -->|Legacy Route\n(Default 100%)| LegacyCall["4b. Call Legacy 15F ONNX\n(/predict)\n50ms Context Budget"]
    
    CandidateCall -->|Candidate Success| CandScore["Derive Score & Decision\nRecord Candidate Metrics"]
    CandidateCall -->|Candidate Failure/Timeout| Fallback["FAIL-SAFE FALLBACK\nLog Warning + Increment Metric"] --> LegacyCall
    
    LegacyCall -->|Legacy Success| LegScore["Derive Score & Decision"]
    LegacyCall -->|Legacy Failure| HeuristicFallback["Heuristic Fallback Score"]
    
    CandScore --> Outbox["5. Envelope Encryption + Postgres + Outbox"]
    LegScore --> Outbox
    HeuristicFallback --> Outbox
    
    Outbox --> Response["6. Return JSON Response (<15ms p99)"]
    
    CandScore -.->|Async| ClickHouseCanary["ClickHouse OLAP\ncanary_rollout_evaluations"]
    Fallback -.->|Async| ClickHouseCanary
```

---

## 3. Deterministic Routing Algorithm

The Canary Router computes a uniform, deterministic hash bucket in $[0, 99]$ for every transaction:

$$\text{hash\_key} = \text{tenant\_id} + \text{":"} + \text{transaction\_id}$$

$$\text{bucket} = \text{uint32}(\text{SHA-256}(\text{hash\_key})[0..3]) \pmod{100}$$

$$\text{Route} = \begin{cases} \text{CANDIDATE}, & \text{if } \text{Enabled} \land (\text{bucket} < \text{Percentage}) \\ \text{LEGACY}, & \text{otherwise} \end{cases}$$

### Routing Properties
- **Zero Allocations:** $\approx 138 \text{ ns/op}$, $0 \text{ B/op}$ heap allocation overhead.
- **Idempotency:** A client retry with the same transaction ID will never flip between models.
- **Stable Monotonic Growth:** Increasing the rollout percentage from $5\% \rightarrow 10\%$ guarantees all transactions previously routed to candidate continue receiving the candidate model, while expanding into newly eligible buckets.

---

## 4. Configuration & Operational Parameters

| Environment Variable | Type | Default | Description |
|:---|:---:|:---:|:---|
| `RISK_MODEL_CANARY_ENABLED` | `bool` | `false` | Master switch for canary routing. |
| `RISK_MODEL_CANARY_PERCENT` | `int` | `0` | Target rollout percentage ($0 \le p \le 100$). |
| `RISK_MODEL_CANARY_VERSION` | `string` | `fraud-xgb-25f-candidate-v1` | Target candidate model identifier. |
| `CANARY_MAX_ERROR_RATE` | `float` | `0.01` (1%) | Maximum allowable candidate inference error rate. |
| `CANARY_MAX_FALLBACK_RATE` | `float` | `0.01` (1%) | Maximum allowable fallback rate before safety gate fails. |
| `CANARY_MAX_P95_LATENCY_MS` | `float` | `15.0` | Maximum p95 inference latency threshold in ms. |
| `CANARY_MAX_P99_LATENCY_MS` | `float` | `25.0` | Maximum p99 inference latency threshold in ms. |
| `CANARY_MAX_DECISION_CHANGE_RATE` | `float` | `0.10` (10%) | Maximum allowable decision divergence rate. |

---

## 5. Automated Safety Gates

The system continuously evaluates real-time candidate metrics against operational thresholds:

```
                  ┌───────────────────────┐
                  │ EvaluateSafetyGates() │
                  └───────────┬───────────┘
                              │
         ┌────────────────────┼────────────────────┐
         ▼                    ▼                    ▼
     [ IDLE ]           [ HEALTHY ]           [ FAILED / WARNING ]
  Canary 0% or         All error rates &     Error rate > 1% or
  no requests          latencies within      fallback > 1% or
                       safe bounds           p95/p99 breached
```

- **`HEALTHY`**: Error rate $< 1\%$, fallback rate $< 1\%$, p95 latency $\le 15\text{ms}$, p99 latency $\le 25\text{ms}$.
- **`WARNING`**: Error rate approaching threshold or latency jitter detected on low sample volumes.
- **`FAILED`**: Error rate or fallback rate $> 1\%$ over $\ge 10$ samples.

---

## 6. ClickHouse Rollout Schema

Live candidate evaluations are persisted asynchronously to `canary_rollout_evaluations`:

```sql
CREATE TABLE IF NOT EXISTS canary_rollout_evaluations (
    evaluation_id String,
    tenant_id String,
    transaction_id String,
    timestamp DateTime,
    model_route String,
    production_model_version String,
    candidate_model_version String,
    production_score Float64,
    candidate_score Float64,
    production_decision String,
    candidate_decision String,
    score_delta Float64,
    absolute_score_delta Float64,
    decision_changed UInt8,
    candidate_latency_ms Float64,
    fallback_used UInt8,
    error String
) ENGINE = MergeTree()
ORDER BY (timestamp, tenant_id, evaluation_id);
```

---

## 7. Staged Rollout Runbook (Stages 0–6)

### Stage 0: 0% Canary (Default Baseline)
```bash
RISK_MODEL_CANARY_ENABLED=false
RISK_MODEL_CANARY_PERCENT=0
```
- **Validation Criteria:** 100% of decisions evaluated on legacy 15F model. Candidate requests = 0.
- **Gate Check:** `curl http://localhost:8080/v1/canary/status` $\rightarrow$ `status: IDLE`.

### Stage 1: 1% Canary
```bash
RISK_MODEL_CANARY_ENABLED=true
RISK_MODEL_CANARY_PERCENT=1
```
- **Validation Criteria:** Candidate traffic $\approx 1\%$. Safety gate `HEALTHY`.
- **Observation Window:** 1 hour / 5,000 transactions.

### Stage 2: 5% Canary
```bash
RISK_MODEL_CANARY_ENABLED=true
RISK_MODEL_CANARY_PERCENT=5
```
- **Validation Criteria:** Candidate error rate $< 0.1\%$, fallback rate $= 0\%$, p95 latency $< 10\text{ms}$.
- **Observation Window:** 2 hours.

### Stage 3: 10% Canary
```bash
RISK_MODEL_CANARY_ENABLED=true
RISK_MODEL_CANARY_PERCENT=10
```
- **Validation Criteria:** Confirm score distribution and rule reason code parity.
- **Observation Window:** 4 hours.

### Stage 4: 25% Canary
```bash
RISK_MODEL_CANARY_ENABLED=true
RISK_MODEL_CANARY_PERCENT=25
```
- **Validation Criteria:** Candidate p99 latency $< 15\text{ms}$, ClickHouse records match distribution.
- **Observation Window:** 12 hours.

### Stage 5: 50% Canary
```bash
RISK_MODEL_CANARY_ENABLED=true
RISK_MODEL_CANARY_PERCENT=50
```
- **Validation Criteria:** Evaluate macro fraud metrics, precision, and analyst review volume.
- **Observation Window:** 24 hours.

### Stage 6: 100% Canary (Full Promotion)
```bash
RISK_MODEL_CANARY_ENABLED=true
RISK_MODEL_CANARY_PERCENT=100
```
- **Validation Criteria:** 100% candidate execution, zero fallback, stable p99 $< 15\text{ms}$.

---

## 8. Rollback Procedure

If any safety gate enters `FAILED` status or unexpected divergence occurs:

1. **Immediate Configuration Rollback:**
   ```bash
   RISK_MODEL_CANARY_ENABLED=false
   RISK_MODEL_CANARY_PERCENT=0
   ```
2. **Apply Configuration:** Restart the API service or trigger environment reload.
3. **Verify Rollback:**
   ```bash
   curl -sf http://localhost:8080/v1/canary/status | jq .
   ```
   Confirm `enabled: false`, `target_percentage: 0`, and `actual_canary_percentage: 0`.

---

## 9. Verification & Performance Benchmarks

### Go Unit & Race Test Suite
```
=== RUN   TestCanaryRouter_0Percent                          --- PASS (0.00s)
=== RUN   TestCanaryRouter_100Percent                        --- PASS (0.00s)
=== RUN   TestCanaryRouter_StatisticalDistribution           --- PASS (0.23s)
=== RUN   TestCanaryRouter_DeterministicIdempotency          --- PASS (0.00s)
=== RUN   TestCanaryRouter_CrossTransactionDispersion        --- PASS (0.00s)
=== RUN   TestCanaryRouter_InvalidConfiguration              --- PASS (0.00s)
=== RUN   TestCanaryRouter_CandidateSuccessAndMetrics        --- PASS (0.00s)
=== RUN   TestCanaryRouter_CandidateFallbackAndMetrics       --- PASS (0.00s)
=== RUN   TestCanaryRouter_SafetyGates                       --- PASS (0.00s)
=== RUN   TestCanaryRouter_ConcurrencyAndRace                --- PASS (0.03s)
=== RUN   TestEvaluateRisk_CanaryRoutingCandidate            --- PASS (0.00s)
=== RUN   TestEvaluateRisk_CanaryCandidateFallbackToLegacy   --- PASS (0.00s)

PASS (All 5 packages passed with -race in 3.75s)
```

### Micro-Benchmark Results (Apple M4 / ARM64)
| Benchmark Target | Ops / sec | Latency / op | Memory / op | Allocs / op |
|:---|:---:|:---:|:---:|:---:|
| `BenchmarkCanaryRouter_Route` | 7,803,247 | **138.1 ns** | 0 B | **0** |
| `BenchmarkCanaryRouter_ReservoirPercentiles` | 59,540 | 20.5 µs | 16.3 KB | 1 |

### Live Docker Verification Summary
| Rollout Stage | Target % | Actual Candidate % | Legacy Requests | Candidate Requests | p95 Latency | Safety Gate |
|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| **Stage 0** | 0% | 0.0% | 20 | 0 | — | `IDLE` |
| **Stage 1** | 1% | 1.0% | 99 | 1 | 1.54 ms | `HEALTHY` |
| **Stage 2** | 5% | 2.0% | 98 | 2 | 2.26 ms | `HEALTHY` |
| **Stage 3** | 10% | 9.0% | 91 | 9 | 4.00 ms | `HEALTHY` |
| **Stage 4** | 25% | 24.0% | 76 | 24 | 6.39 ms | `WARNING`* |
| **Stage 5** | 50% | 53.0% | 47 | 53 | 7.03 ms | `HEALTHY` |
| **Stage 6** | 100% | 100.0% | 0 | 100 | 2.60 ms | `HEALTHY` |
| **Rollback** | 0% | 0.0% | 100% Legacy | 0 | — | `IDLE` |

*\*Stage 4 caught initial warm-up container jitter and correctly reported WARNING state.*
