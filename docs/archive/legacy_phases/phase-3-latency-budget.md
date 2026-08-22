# Phase 3 — End-to-End Latency Budget & Critical Path Breakdown

**Document Version:** 1.0  
**Synchronous Target Budget:** $< 100\text{ms}$ p95 (Operational Target: $< 15\text{ms}$ p95)  
**ML Inference Deadline Budget:** $50\text{ms}$ hard context timeout  

---

## 1. Synchronous Critical Path Latency Breakdown

Every microsecond on the payment authorization critical path directly affects checkout conversion rates. The Phase 3 architecture allocates strict sub-millisecond latency budgets across all synchronous stages:

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│              SYNCHRONOUS PAYMENT DECISION FAST-PATH BUDGET (< 15ms)             │
├───────────────────────────────────────────────────────┬────────────┬────────────┤
│ Stage / Operation                                     │ p50 Budget │ p95 Budget │
├───────────────────────────────────────────────────────┼────────────┼────────────┤
│ 1. HTTP Ingestion & JSON Unmarshaling                 │ 0.15 ms    │ 0.35 ms    │
│ 2. Tenant Validation & Device SHA-256 Hashing         │ 0.10 ms    │ 0.20 ms    │
│ 3. Redis Pipelined Feature Query (5 Operations):      │ 1.20 ms    │ 2.50 ms    │
│    • ZCOUNT {tenant}:vel:ip:1h                        │            │            │
│    • ZCOUNT {tenant}:vel:token:24h                    │            │            │
│    • ZCOUNT {tenant}:vel:device:1h                    │            │            │
│    • SCARD  {tenant}:dev:accounts:24h                 │            │            │
│    • EXISTS {tenant}:dev:known:{hash}                 │            │            │
│ 4. Pre-Rules Declarative AST Evaluation               │ 0.20 ms    │ 0.45 ms    │
│ 5. Feature Vector Construction & Formatting           │ 0.10 ms    │ 0.20 ms    │
│ 6. Python ML Sidecar Inference (HTTP Keep-Alive):     │ 2.00 ms    │ 4.50 ms    │
│    • ONNX Runtime 25-Feature Forward Pass             │ (1.20 ms)  │ (1.80 ms)  │
│    • Beta Calibration logit mapping                   │ (0.05 ms)  │ (0.08 ms)  │
│    • Cost-Sensitive Expected Loss & Action Selection  │ (0.10 ms)  │ (0.15 ms)  │
│ 7. Redis Pipelined Event Recording (ZADD + EXPIRE)    │ 0.80 ms    │ 1.50 ms    │
│ 8. JSON Response Serialization & HTTP Flush           │ 0.15 ms    │ 0.30 ms    │
├───────────────────────────────────────────────────────┼────────────┼────────────┤
│ TOTAL SYNCHRONOUS DECISION LATENCY                    │ 4.70 ms    │ 10.00 ms   │
└───────────────────────────────────────────────────────┴────────────┴────────────┘
```

---

## 2. Critical Architectural Rules for Latency Compliance

1. **ZERO Relational Database Lookups on Synchronous Fast-Path:**
   - PostgreSQL queries MUST NEVER be executed inside the synchronous decision flow.
   - All real-time signals (velocities, known status, account counts) MUST be served exclusively from Redis in a single atomic pipelined roundtrip ($< 2.5\text{ms}$).
2. **Asynchronous Ledger Persistence via Worker / Goroutines:**
   - PostgreSQL writes (`risk_decisions`, `devices`, `outbox_events`) are dispatched asynchronously or completed after response flush, ensuring database I/O latency never blocks client authorization.
3. **HTTP Connection Pooling & Keep-Alive:**
   - Go `http.Client` to Python ML sidecar maintains persistent HTTP/1.1 TCP connection pools (`MaxIdleConns = 100`, `IdleConnTimeout = 90s`), eliminating TCP handshake and TLS negotiation overhead.
4. **Hard 50ms Context Deadline:**
   - If the ML sidecar exceeds 50ms (or network is congested), Go context deadline immediately aborts the call, engages fallback heuristics in $<0.1\text{ms}$, and returns the decision marked `is_degraded: true`.
