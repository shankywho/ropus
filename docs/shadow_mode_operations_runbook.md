# ROPUS Platform — Shadow-Mode Operations Runbook

## 1. Operational Overview

This runbook guides SREs and MLOps engineers in monitoring, maintaining, and troubleshooting shadow model execution for `extended_catboost_58f`.

---

## 2. Telemetry & Metric Monitoring

### Key Prometheus Metrics:
- `riskengine_shadow_requests_total`: Total shadow score evaluations attempted.
- `riskengine_shadow_errors_total`: Errors encountered during shadow scoring.
- `riskengine_shadow_queue_dropped_total`: Tasks dropped due to queue saturation.
- `riskengine_shadow_decision_divergence_total`: Cumulative count of action disagreements.
- `riskengine_shadow_inference_latency_seconds`: P50, P95, and P99 shadow sidecar latency.

### Prometheus Alert Rules:
```yaml
groups:
  - name: ShadowScorerAlerts
    rules:
      - alert: ShadowQueueSaturation
        expr: rate(riskengine_shadow_queue_dropped_total[5m]) > 0.05
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "Shadow scoring queue dropping tasks (>5% drop rate)"

      - alert: ShadowCandidateHighLatency
        expr: histogram_quantile(0.99, rate(riskengine_shadow_inference_latency_seconds_bucket[5m])) > 0.010
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Shadow candidate P99 inference latency exceeds 10ms"

      - alert: ShadowHighDisagreementSpike
        expr: rate(riskengine_shadow_decision_divergence_total[10m]) / rate(riskengine_shadow_requests_total[10m]) > 0.25
        for: 10m
        labels:
          severity: critical
        annotations:
          summary: "Shadow model action disagreement spiked above 25%"
```

---

## 3. Failure Modes & Automated Recovery Procedures

### Scenario A: Shadow Candidate Service Unavailable (HTTP 503 / Crash)
1. **System Behavior**: `ShadowScorer` logs the error and increments `riskengine_shadow_errors_total`. The Go risk orchestrator returns the champion decision synchronously with zero delay.
2. **Action**: Check Python ML service pod status via `kubectl get pods -l app=ml-service`. Inspect crash logs for OOM or dependency faults.

### Scenario B: Shadow Scoring Queue Saturation
1. **System Behavior**: If the candidate execution rate falls behind ingress traffic, the worker channel drops excess tasks rather than blocking memory.
2. **Action**: Scale shadow worker count in `backend/cmd/api/main.go` from `WorkerCount: 4` to `WorkerCount: 8` or increase sidecar CPU allocation.

---

## 4. Emergency Circuit-Breaker Rollback

If shadow candidate execution consumes excess CPU/memory or exhibits pathological telemetry, disable shadow mode instantly via environment variable or API:
```bash
# Emergency disable shadow scorer without restarting Go backend
curl -X POST http://localhost:8080/v1/admin/shadow-mode/disable \
  -H "X-API-Key: $ADMIN_API_KEY" \
  -H "X-Actor-ID: MLOPS_LEAD"
```
