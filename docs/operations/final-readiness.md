# 🛡️ Engineering Readiness & Prototype Scope Assessment

> **Track 02: AI Risk Manager (Razorpay AI Buildathon)**
> This document summarizes the engineering implementation status, tested capabilities, and prototype boundaries.

---

## 1. Implemented & Tested Subsystems

| Subsystem Area | Implementation Status | Verification Evidence |
|:---|:---|:---|
| **Razorpay Ingress & Auth** | Implemented & covered by local tests | HMAC-SHA256 signature verification (`X-Razorpay-Signature`), normalized payload parsing, and idempotent replay handling in `backend/internal/ingestion/razorpay_adapter.go`. |
| **Model & Calibration** | Implemented & covered by local tests | 15-Feature gradient boosted trees with post-hoc Beta probability calibration (ECE = 0.0119 on held-out test split). |
| **Bayes Minimum Risk Policy** | Implemented & covered by local tests | Financial loss matrix selecting optimal action (`ALLOW`, `STEP_UP`, `MANUAL_REVIEW`, `DECLINE`) to minimize expected chargeback and friction costs. |
| **Resilience & Timeout Fallback** | Implemented & covered by local tests | 8ms context deadline circuit breaker degrading to in-memory JSON-AST rules with zero downstream SQL/network I/O; non-blocking audit ledger writes. |
| **Audit Ledger** | Implemented & covered by local tests | Cryptographic SHA-256 Merkle chain verified via `GET /v1/audit/verify`. |
| **Analyst Control Plane** | Implemented & covered by local tests | React / TanStack Start interface displaying real-time evaluations, reason codes, decision logs, and metrics. |

---

## 2. Explicit Prototype Boundaries & Limitations

1. **Local Prototype Scope**:
   - ROPUS is an engineering prototype submission for Track 02 of the Razorpay AI Buildathon.
   - It runs locally via Docker Compose or native binaries.
   - It has **not** been independently audited for SOC 2 Type II or PCI-DSS compliance.

2. **No Live Payment Rail Connection**:
   - The system ingests sandboxed or synthetic Razorpay webhook events.
   - It returns defensive risk recommendations (`ALLOW_RECOMMENDATION`, `STEP_UP_RECOMMENDATION`, `MANUAL_REVIEW`, `DECLINE_RECOMMENDATION`).
   - It does **not** process live cardholder payments or settle real funds.

3. **Synthetic / Benchmark-Derived Evaluation**:
   - ML models are evaluated on a frozen 1,200-transaction holdout partition (52 positive labels, 1,148 negative labels) derived from canonical IEEE-CIS fraud detection distributions.
   - Evaluation is conducted strictly offline with zero parameter tuning on the holdout split.
