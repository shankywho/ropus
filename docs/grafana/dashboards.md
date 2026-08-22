# Grafana Production Observability Dashboards

This document specifies the complete Grafana dashboard suite for monitoring and operating the **AI Risk Manager / Ropus** platform in production.

---

## 1. Executive Risk Overview Dashboard

| Panel Title | Visualization | PromQL Metric / Query | Description |
| :--- | :--- | :--- | :--- |
| **Total Evaluation Rate** | Stat / Graph | `sum(rate(risk_evaluations_total[1m]))` | Real-time synchronous decisions/sec |
| **Approval vs Decline Ratio** | Pie Chart / Timeseries | `sum by (action) (rate(risk_model_decisions_total[1m]))` | Real-time decision distribution |
| **Active Production Model** | Stat Gauge | `model_active_info{role="production"}` | Current primary active model version |
| **Emergency Fallback In-Use** | Stat / Alert | `sum(rate(risk_evaluation_fallback_total[1m]))` | Triggers warning if fallback rate $> 0.1\%$ |
| **Platform Availability** | Gauge | `slo_availability * 100` | Target: $\ge 99.95\%$ |

---

## 2. API & Latency Dashboard

| Panel Title | Visualization | PromQL Metric / Query | Description |
| :--- | :--- | :--- | :--- |
| **P50 Latency** | Timeseries | `histogram_quantile(0.50, sum(rate(risk_evaluation_latency_ms_bucket[1m])) by (le))` | P50 synchronous inference latency |
| **P95 Latency** | Timeseries | `histogram_quantile(0.95, sum(rate(risk_evaluation_latency_ms_bucket[1m])) by (le))` | P95 latency (SLO Target: $< 15\text{ms}$) |
| **P99 Latency** | Timeseries | `histogram_quantile(0.99, sum(rate(risk_evaluation_latency_ms_bucket[1m])) by (le))` | P99 tail latency (SLO Target: $< 30\text{ms}$) |
| **Error Rate %** | Timeseries | `sum(rate(risk_evaluation_errors_total[1m])) / sum(rate(risk_evaluations_total[1m])) * 100` | Total synchronous HTTP error rate |

---

## 3. Drift & Retraining Pipeline Dashboard

| Panel Title | Visualization | PromQL Metric / Query | Description |
| :--- | :--- | :--- | :--- |
| **Max PSI (Population Stability)** | Gauge / Graph | `drift_max_psi` | Green $< 0.10$, Yellow $0.10 - 0.25$, Red $> 0.25$ |
| **Drift Severity Status** | Discrete Stat | `drift_status` | 0: Healthy, 1: Warning, 2: Degraded, 3: Critical |
| **Active Retraining Job** | Boolean Stat | `retraining_active` | 1 if candidate training is running in background |
| **Retraining Success / Failure** | Counter Graph | `sum(increase(retraining_jobs_total[1h]))`, `sum(increase(retraining_jobs_failed_total[1h]))` | Retraining throughput and reliability |

---

## 4. Canary Rollouts & Circuit Breaker Dashboard

| Panel Title | Visualization | PromQL Metric / Query | Description |
| :--- | :--- | :--- | :--- |
| **Canary Traffic Percentage** | Gauge | `canary_stage` | $0\% - 100\%$ candidate traffic allocation |
| **Circuit Breaker Status** | Status Indicator | `circuit_breaker_state` | 0: Healthy, 1: Warning, 2: Failed, 3: Rolled Back |
| **Automated Rollback Events** | Counter Graph | `sum(increase(canary_rollbacks_total[1h]))` | Automatic rollbacks triggered by safety breaches |
| **Model Promotions** | Counter Graph | `sum(increase(model_promotions_total[24h]))` | Production model promotions |

---

## 5. SLO & Error Budget Management Dashboard

| Panel Title | Visualization | PromQL Metric / Query | Description |
| :--- | :--- | :--- | :--- |
| **Availability Error Budget Remaining** | Gauge Bar | `slo_error_budget_remaining{slo="slo_availability"}` | Remaining error budget percentage |
| **Latency Error Budget Remaining** | Gauge Bar | `slo_error_budget_remaining{slo="slo_p99_latency"}` | Remaining latency budget percentage |
| **SLO Burn Rate** | Timeseries | `slo_burn_rate` | Fast burn alerts at $> 14.4\times$ |
| **Promotion Freeze State** | Alert Stat | `retraining_model_frozen` | 1 if auto-locked due to budget exhaustion |
