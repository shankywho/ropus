# ROPUS — Pragmatic Promotion Governance Decision & Policy Specification

---

## 1. Executive Summary & Governance Review

### The Architectural Dilemma
Under the strict single-tier promotion policy, candidate `extended_catboost_58f` is blocked from any customer routing because:
- Live commercial merchant ingress volume = `0 / 10,000` (Gate 1).
- Confirmed matured fraud chargebacks = `0 / 50` (Gate 2).

In an environment without direct live merchant gateway connectivity or a real-time chargeback stream, this creates a **circular deployment deadlock**:
1. Full unconditioned promotion requires 10,000 live production transactions and 50 matured chargebacks over a 60-day maturation window.
2. Live production traffic cannot be generated internally without violating the anti-masquerading policy (never label synthetic/fixture data as `LIVE_PRODUCTION`).
3. Model validation on frozen holdout data is already empirically and mathematically robust:
   - **BMR Loss Advantage:** Saves \$649–\$1,293 on \$1k–\$10k tiers ($p = 0.028$).
   - **Calibration Quality:** Beta calibration $\text{ECE} = 0.0035 \le 0.010$.
   - **Inference SLA:** Sidecar $\text{P99} = 0.275\text{ ms} \le 5.0\text{ ms}$.
   - **Action Disagreement:** $8.33\% \le 15.0\%$.
   - **Cold vs Warm-Start:** Point-in-time temporal features validated without forward leakage.
   - **Zero PII Exposure:** Tokenized credentials only.

---

## 2. Governance Decision: Option B (Controlled Dual-Track Policy)

Rather than remaining in an unresolvable deadlock or weakening empirical standards, the Risk Model Governance Committee establishes a **Dual-Track Governance Framework**:

```
                         ┌──────────────────────────────────────────────┐
                         │ Candidate Model: extended_catboost_58f       │
                         └──────────────────────┬───────────────────────┘
                                                │
                                                ▼
                         ┌──────────────────────────────────────────────┐
                         │ Tier 1: Offline Scientific & Empirical Gate │
                         │ • BMR Loss Advantage: p = 0.028 (<0.05)      │
                         │ • Beta Calibration: ECE = 0.0035 (<=0.010)   │
                         │ • Inference SLA: P99 = 0.275ms (<=5.0ms)     │
                         │ • Action Disagreement: 8.33% (<=15.0%)       │
                         └──────────────────────┬───────────────────────┘
                                                │
                                                ▼
                         ┌──────────────────────────────────────────────┐
                         │ Explainability Status:                       │
                         │ • MODEL_VALIDATED: true                      │
                         │ • PRODUCTION_SHADOW_READY: true              │
                         │ • LIVE_EVIDENCE_UNAVAILABLE: true            │
                         │ • CONDITIONAL_PROMOTION_ELIGIBLE: true       │
                         │ • PRODUCTION_PROMOTION_BLOCKED: true         │
                         └──────┬────────────────────────────────┬──────┘
                                │                                │
                                ▼                                ▼
               ┌─────────────────────────────────┐   ┌─────────────────────────────────┐
               │ Track 1: Conditional Canary     │   │ Track 2: Unconditional Cutover  │
               │ • Status: ELIGIBLE              │   │ • Status: HARD BLOCKED          │
               │ • Customer Exposure: 1% to 10%  │   │ • Customer Exposure: 100%       │
               │ • Maker-Checker Signoff Required│   │ • Requires: >=10k Live Tx       │
               │ • Auto-Rollback on Error > 1.0% │   │ • Requires: >=50 Matured Frauds │
               │ • Auto-Rollback on Latency > 5ms│   │ • Requires: >60-Day Soak        │
               └─────────────────────────────────┘   └─────────────────────────────────┘
```

---

## 3. Explicit 5-Way Governance Status Definitions

| Status Tag | State | Definition & Policy Mandate |
| :--- | :---: | :--- |
| **`MODEL_VALIDATED`** | **`TRUE`** | The model has satisfied all statistical, calibration, discrimination, and economic criteria on the frozen holdout dataset. |
| **`PRODUCTION_SHADOW_READY`** | **`TRUE`** | Asynchronous worker pool, non-blocking queue, fail-open fallback, and ClickHouse telemetry are fully operational and tested. |
| **`LIVE_EVIDENCE_UNAVAILABLE`** | **`TRUE`** | Authentic merchant ingress stream and matured chargeback outcomes have not yet arrived ($N = 0$). |
| **`CONDITIONAL_PROMOTION_ELIGIBLE`** | **`TRUE`** | The model is authorized for low-exposure canary routing ($\le 10\%$) with automated circuit breakers, pending Maker-Checker dual approval. |
| **`PRODUCTION_PROMOTION_BLOCKED`** | **`TRUE`** | Full $100\%$ customer decision authority replacement of the champion is strictly blocked until all 6 live gates pass on live data. |

---

## 4. Gate Breakdown: Offline vs Live Production

| Gate # | Name | Threshold | Offline Basis (Tier 1) | Live Basis (Tier 2) | Track 1 Mandatory? | Track 2 Mandatory? |
| :---: | :--- | :--- | :--- | :--- | :---: | :---: |
| **Gate 1** | **Live Volume** | $\ge 10,000$ tx | Fixture $N=8,000$ | Live Ingress | **NO** (Canary accumulates live tx) | **YES** |
| **Gate 2** | **Confirmed Frauds** | $\ge 50$ frauds | Fixture $N=280$ frauds | Live $>60\text{d}$ Chargebacks | **NO** (Canary accumulates live frauds) | **YES** |
| **Gate 3** | **Economic BMR** | $p < 0.05$ | Saves \$649–\$1,293 ($p=0.028$) | Live Outcome Delta | **YES** (Offline) | **YES** (Live) |
| **Gate 4** | **Calibration** | $\text{ECE} \le 0.010$ | Beta $\text{ECE} = 0.0035$ | Live ECE | **YES** (Offline) | **YES** (Live) |
| **Gate 5** | **Inference SLA** | $\text{P99} \le 5.0\text{ ms}$ | Sidecar $\text{P99} = 0.275\text{ ms}$| Live P99 | **YES** (Offline) | **YES** (Live) |
| **Gate 6** | **Disagreement** | $\le 15.0\%$ | $8.33\%$ at $\tau^* = 0.220$ | Live Disagreement | **YES** (Offline) | **YES** (Live) |

---

## 5. Production Safety Invariants Maintained

1. **No Silent Authority Elevation:** Candidate remains at $0\%$ authority in default runtime configuration.
2. **Canary Lock:** `RISK_MODEL_CANARY_ENABLED` defaults to `false`. It can only be activated by explicit operator configuration.
3. **Automated Rollback Circuit Breakers:**
   - Error Rate $\ge 1.0\%$ $\implies$ immediate automated fallback to champion.
   - P95 Latency $\ge 5.0\text{ ms}$ $\implies$ immediate automated fallback to champion.
   - Fallback Rate $\ge 0.5\%$ $\implies$ canary disabled.
4. **Zero-Masquerading Policy:** Offline records remain strictly tagged as `OFFLINE_TEST`. They are never relabeled as `LIVE_PRODUCTION`.
5. **Maker-Checker Signoff Required:** `PromoteConditionalCanary` in Go enforces that an authorized governance actor (`ApprovalActor`) has signed off.

---

## 6. Verification Summary

- **Go Backend Test Suite:** `go test ./...` — **100% PASS (0 failures)**.
- **Python ML Test Suite:** `pytest ml-service/tests` — **100% PASS (139/139 passed)**.
- **Frozen Fixture Checksum:** `a30a387ad0fa8743599d6043120be6bd66ac17184b8eee4fb9ce764970201d44` — **VERIFIED UNMODIFIED**.
