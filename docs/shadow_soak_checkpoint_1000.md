# ROPUS — Production Shadow Soak Checkpoint (1,000 Transactions)

---

## 1. Checkpoint Summary

- **Checkpoint Horizon:** **1,000 Live Transactions**
- **Evaluation Timestamp (UTC):** `2026-08-27T06:54:55.966176+00:00`
- **Champion Model:** `fraud-xgb-25f-v3.0` (**100% Authority**)
- **Candidate Model:** `extended_catboost_58f` (**0% Authority, Shadow Mode**)
- **Promotion Status:** **BLOCKED (`promotion_allowed = false`)**
- **Governance Decision:** `REMAIN_IN_SHADOW_MODE`

---

## 2. Checkpoint Metrics & Provenance Audit

| Category | Metric | Value | Target / SLA |
| :--- | :--- | :--- | :--- |
| **Volume** | Genuine `LIVE_PRODUCTION` Count | **0** | $\ge 1,000$ |
| **Volume** | Unique Evaluations Stored | **0** | Matches live count |
| **Operational** | Candidate P99 Inference Latency | **0.0 ms** | $\le 5.0\text{ ms}$ |
| **Operational** | Candidate Failure Count | **0** | $0$ impact |
| **Operational** | Queue Drops | **0** | $0$ impact |
| **Operational** | Action Disagreement Rate | **0.00%** | $\le 15.0\%$ |
| **Operational** | Champion Decision / Latency Impact | **NONE DETECTED** | Zero Impact |
| **Maturation** | Confirmed Matured Frauds ($>60\text{d}$) | **0** | $\ge 50$ (Gate 2) |

---

## 3. Promotion Gates at Checkpoint 1,000

| Gate | Target | Checkpoint Status | Status Enum |
| :--- | :--- | :--- | :--- |
| **Gate 1** | $\ge 10,000$ live transactions | 0 / 10,000 | `NOT_POPULATED` |
| **Gate 2** | $\ge 50$ matured confirmed frauds | 0 / 50 | `NOT_POPULATED` |
| **Gate 3** | $p < 0.05$ BMR economic advantage | Requires $\ge 50$ matured frauds | `NOT_POPULATED` |
| **Gate 4** | Live $\text{ECE} \le 0.010$ | Requires $\ge 50$ matured frauds | `NOT_POPULATED` |
| **Gate 5** | Live P99 latency $\le 5.0\text{ ms}$ | Requires live traffic stream | `NOT_POPULATED` |
| **Gate 6** | Live disagreement $\le 15.0\%$ | Requires live traffic stream | `NOT_POPULATED` |

---

## 4. Governance Rule on Milestone Completion

> [!IMPORTANT]
> Reaching the 1,000 transaction horizon does **NOT** automatically authorize candidate promotion or canary routing. Promotion remains strictly blocked until all 6 gates, including the mandatory 60-day delayed fraud maturation window and dual-control Maker-Checker approvals, are verified.
