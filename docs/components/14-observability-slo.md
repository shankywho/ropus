# Component 14: Observability, Metrics & Contractual SLO Engine

---

## 1. Why It Exists
Operating mission-critical financial risk infrastructure requires full transparency into evaluation latencies, decision volume distributions, error budgets, and SLA compliance.

The **Observability Subsystem** (`backend/internal/observability/`, `backend/internal/performance/`) provides:
1. Standardized **Prometheus Metric Counters & Histograms** exported at `/metrics`.
2. **Distributed OpenTelemetry / Jaeger Traces** tracking requests from ingress to ML inference.
3. Continuous **Contractual 99.99% Availability & Latency SLO Monitoring** with real-time error budget burn alerts.

---

## 2. Key Prometheus Metrics Reference

```text
Metric Name                          Type       Labels                     Description
------------------------------------+----------+--------------------------+----------------------------------------------
ropus_risk_evaluations_total        Counter    tenant_id, decision        Total transaction evaluations by verdict
ropus_decision_latency_ms_bucket    Histogram  tenant_id, endpoint        Evaluation latency histogram (buckets: 1, 2, 5, 10, 25ms)
ropus_ml_inference_time_us          Histogram  model_version              ML model evaluation time in microseconds
ropus_graph_traversal_duration_us   Histogram  traversal_depth            3-hop graph traversal time in microseconds
ropus_circuit_breaker_state         Gauge      dependency_name            Current state (0: Closed, 1: Half-Open, 2: Open)
ropus_rate_limit_rejections_total   Counter    tenant_id, plan_tier       Requests dropped due to quota exhaustion
```

---

## 3. Contractual SLO & Error Budget Formulation

### Availability SLO Target: $99.99\%$ (Four Nines)
$$\text{Availability} = \frac{\text{Successful Evaluations (HTTP } 200\text{)}}{\text{Total Evaluation Requests}} \ge 0.9999$$

### Latency SLO Target: $P99 < 10.0\text{ms}$
$$\text{Error Budget Burn Rate} = \frac{\text{Failed Requests in Window}}{\text{Allowed Failure Budget}}$$

If the burn rate exceeds $2.0\times$ across a 1-hour window, high-priority PagerDuty alerts are fired to the on-call site reliability engineer (SRE).

---

## 4. Source Code Map
- [`backend/internal/observability/metrics.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/observability/metrics.go): Prometheus metric registration and latency recording middleware.
- [`backend/internal/observability/slo_tracker.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/observability/slo_tracker.go): Rolling error budget and SLO availability tracker.

---

## 5. Cross-Component Links
- [Component 01: Product API](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/01-product-api.md) — Instruments evaluation latency.
- [Component 12: Resilience](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/12-resilience-circuit-breaker.md) — Exports circuit breaker state transitions.
