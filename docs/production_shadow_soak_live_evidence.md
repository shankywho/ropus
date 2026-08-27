# ROPUS — Production Shadow Soak Live Evidence & Governance Report

---

## 1. Governance Status & Explainability Tiering

| Governance Status Tier | Value / State | Meaning / Verification Basis |
| :--- | :--- | :--- |
| **`MODEL_VALIDATED`** | **`True`** | Verified on frozen holdout fixture (BMR $p=0.028$, $\text{ECE}=0.0035$, $\text{P99}=0.275\text{ ms}$, Disag $=8.33\%$) |
| **`PRODUCTION_SHADOW_READY`** | **`True`** | Asynchronous worker pool, non-blocking queue, and fail-open ClickHouse telemetry verified |
| **`LIVE_EVIDENCE_UNAVAILABLE`** | **`True`** | Live commercial merchant ingress currently $0 / 10,000$, confirmed matured frauds $0 / 50$ |
| **`CONDITIONAL_PROMOTION_ELIGIBLE`** | **`True`** | Eligible for controlled canary routing ($\le 10\%$) under Maker-Checker approval & automated rollback guards |
| **`PRODUCTION_PROMOTION_BLOCKED`** | **`True`** | Full unconditioned $100\%$ customer decision authority remains hard-blocked until $60$-day soak completes |

- **Champion Model:** `fraud-xgb-25f-v3.0` (**100.0% Authority**)
- **Candidate Model:** `extended_catboost_58f` (**0.0% Authority, Shadow Mode**)
- **Canary Routing:** **DISABLED (`canary_routing_enabled = false`)**
- **Unconditional Promotion Status:** **`PRODUCTION_PROMOTION_BLOCKED`**
- **Conditional Canary Eligibility:** **`CONDITIONAL_PROMOTION_ELIGIBLE`**
- **Report Generated (UTC):** `2026-08-27T17:23:08.679185+00:00`

---

## 2. Live Evidence Summary

| Dimension | Metric | Live Actual | Target Horizon | Status |
| :--- | :--- | :--- | :--- | :--- |
| **Volume** | Genuine `LIVE_PRODUCTION` Volume | **0** | $\ge 10,000$ | `NOT_POPULATED` |
| **Outcomes** | Confirmed Matured Frauds ($>60\text{d}$) | **0** | $\ge 50$ | `{'description': 'Total confirmed matured fraud chargebacks >= 50', 'target': 50, 'actual': 0, 'evidence_basis': 'NOT_POPULATED', 'status': 'NOT_POPULATED', 'passed': False}` |
| **SLA** | Candidate P99 Inference Latency | **0.0 ms** | $\le 5.0\text{ ms}$ | `NOT_POPULATED` |
| **Alignment** | Action Disagreement Bound | **0.00%** | $\le 15.0\%$ | `NOT_POPULATED` |
| **Economic** | BMR Monetary Loss Delta | **N/A** (0 live frauds) | $p < 0.05$ | `NOT_POPULATED` |
| **Calibration**| Expected Calibration Error (ECE) | **N/A** (0 live outcomes)| $\text{ECE} \le 0.010$ | `NOT_POPULATED` |

---

## 3. Provenance Breakdown

| Provenance Category | Count | Eligible for Gate 1? | Purpose |
| :--- | :--- | :--- | :--- |
| `LIVE_PRODUCTION` | **0** | **YES** | Genuine merchant ingress evaluations |
| `OFFLINE_TEST` | **0** | **NO** | Offline holdout fixtures |
| `SYNTHETIC` | **0** | **NO** | Chaos / soak fault injection |
| `REPLAY` | **0** | **NO** | Historical re-evaluations |

---

## 4. Promotion Gates Audit Ledger

| Gate | Target | Live Evidence | Offline Benchmark Reference | Live Gate Status |
| :--- | :--- | :--- | :--- | :--- |
| **Gate 1: Live Volume** | $\ge 10,000$ tx | **0** | Fixture $N=8,000$ | `NOT_POPULATED` |
| **Gate 2: Confirmed Frauds** | $\ge 50$ frauds ($>60\text{d}$) | **0** | Fixture $N=280$ frauds | `{'description': 'Total confirmed matured fraud chargebacks >= 50', 'target': 50, 'actual': 0, 'evidence_basis': 'NOT_POPULATED', 'status': 'NOT_POPULATED', 'passed': False}` |
| **Gate 3: Economic Advantage** | $p < 0.05$ BMR delta | Awaiting $\ge 50$ live frauds | Saves \$649–\$1,293 ($p=0.028$) | `NOT_POPULATED` |
| **Gate 4: Calibration Quality** | $\text{ECE} \le 0.010$ | Awaiting $\ge 50$ live frauds | Beta $\text{ECE} = 0.0035$ | `NOT_POPULATED` |
| **Gate 5: Inference Latency** | $\text{P99} \le 5.0\text{ ms}$ | Awaiting live stream | Sidecar $\text{P99} = 0.275\text{ ms}$ | `NOT_POPULATED` |
| **Gate 6: Action Disagreement** | $\le 15.0\%$ | Awaiting live stream | $8.33\%$ at $\tau^* = 0.220$ | `NOT_POPULATED` |

---

## 5. Dual-Track Governance Policy

### Track 1: Conditional Canary Deployment (Option B - APPROVED FOR ELIGIBILITY)
- **Eligibility:** **`CONDITIONAL_PROMOTION_ELIGIBLE`**
- **Preconditions:** Completed Tier 1 offline validation (BMR $p < 0.05$, $\text{ECE} \le 0.010$, SLA $\le 5.0\text{ ms}$, Disagreement $\le 15.0\%$) + Maker-Checker authorization.
- **Exposure Cap:** Maximum $5\%$–$10\%$ of traffic routed through Canary Router.
- **Rollback Invariant:** Automated immediate rollback if candidate error rate exceeds $1.0\%$ or P95 latency exceeds $5.0\text{ ms}$.

### Track 2: Full Unconditional Promotion (STRICTLY BLOCKED)
- **Status:** **`PRODUCTION_PROMOTION_BLOCKED`**
- **Reason:** Requires $10,000$ genuine live production transactions and $50$ confirmed matured fraud chargebacks over the $60$-day maturation window.
