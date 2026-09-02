# ⚠️ Technical Scope & Known Limitations

> **Track 02: AI Risk Manager (Razorpay AI Buildathon)**
> This document establishes the explicit technical, environmental, and regulatory boundaries of the ROPUS submission.

---

## 1. Prototype & Hackathon Scope Boundaries

1. **Non-Production Prototype**:
   - ROPUS is an engineering submission developed for the Razorpay AI Buildathon.
   - It is designed to demonstrate an end-to-end, defense-only fraud evaluation loop running on local development machines or containerized stacks.
   - It has **not** undergone third-party penetration testing, enterprise disaster recovery drills, or production SOC2 / ISO-27001 audits.

2. **No Live Payment Rail Connection**:
   - The system receives and validates synthetic or sandboxed Razorpay-format webhook payloads (`payment.authorized`, `payment.failed`, `dispute.created`).
   - ROPUS produces advisory risk recommendations (`ALLOW_RECOMMENDATION`, `STEP_UP_RECOMMENDATION`, `MANUAL_REVIEW`, `DECLINE_RECOMMENDATION`).
   - It does **not** capture credit cards, execute UPI debits, or settle money on banking rails.

3. **Synthetic Evaluation Dataset**:
   - The ML models (15F Production Champion and 25F Candidate) are trained and evaluated on an out-of-time frozen partition derived from canonical IEEE-CIS fraud detection distributions.
   - The held-out test split consists of 1,200 transactions (52 positive labels in a frozen synthetic/benchmark-derived holdout fixture, 1,148 legitimate cases).
   - While statistically rigorous, this dataset reflects benchmark distributions and should not be construed as live production data from Razorpay merchants.

4. **Cost Model Sensitivity & Governance**:
   - The Bayes Minimum Risk (BMR) loss model demonstrates policy arbitration under explicit, user-configured cost parameters.
   - On the current frozen fixture and default local assumptions (₹100 manual review cost, ₹500 false decline cost), the BMR policy incurred 7.0% higher modeled costs than the fixed 0.50 threshold due to operational triage overhead on 4 borderline legitimate items, illustrating why production rollout requires merchant-specific empirical cost validation.

---

## 2. Infrastructure & Architectural Boundaries

1. **Local Storage & Persistence**:
   - Primary state storage uses local PostgreSQL 16, Redis 7, and ClickHouse containers.
   - In single-node development mode, data is persisted to local Docker volumes. Distributed replication across multi-region clusters is not modeled in the hackathon artifact.

2. **Asynchronous Ledger Backpressure**:
   - The SHA-256 audit ledger employs an asynchronous ring channel with a capacity of 20,000 events.
   - Under sustained extreme saturation where the queue fills completely, the orchestrator safely drops audit events to prevent stalling the real-time payment evaluation path, incrementing `dropped_audit_events_total` and logging a rate-limited alert.

3. **Fallback Degradation Path**:
   - When downstream ML sidecar response latency exceeds 8ms or the connection drops, the orchestrator gracefully degrades to in-memory AST deterministic rules.
   - In degraded mode, the system guarantees sub-millisecond deterministic decisions without downstream database or network I/O.

---

## 3. Compliance & Security Boundaries

- **PCI-DSS Scope**: ROPUS expects tokenized payment identifiers (`tok_...`) and masked card numbers (`last4: 4321`). Raw Primary Account Numbers (PAN) and CVVs must never be sent to the API.
- **TLS & Network**: In local Docker Compose mode, services communicate over HTTP within an internal bridge network (`risk_network`). In production deployments, TLS termination would be managed by an API Gateway or reverse proxy.
