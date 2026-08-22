# Phase 3.15 — Automated Drift Detection & Continuous Model Monitoring

## 1. Executive Summary

Phase 3.15 establishes an automated, real-time, non-blocking feature and distribution drift monitoring subsystem for the primary production 25-feature model (`fraud-xgb-25f-v3.0` under feature contract `fraud-risk-25f-v2.5` with `beta-calibrated-v2.5`).

The subsystem calculates statistical divergence across three complementary metrics:
- **Population Stability Index (PSI)**
- **Jensen-Shannon Divergence (JSD)**
- **Kullback-Leibler Divergence (KL)**

It captures empirical moments (mean, standard deviation, min, max) and percentiles (p01, p05, p25, p50, p75, p95, p99), missing value rates, and unseen category rates.

---

## 2. Architecture & Data Flow

```
+-------------------------------------------------------------------------+
|                        Synchronous Risk Pipeline                        |
|                                                                         |
|  Transaction Request                                                    |
|         │                                                               |
|         ▼                                                               |
|  Context & Velocity Stores (Redis, Postgres)                            |
|         │                                                               |
|         ▼                                                               |
|  Canonical 25-Feature Vector Builder (Point-in-Time Safe)               |
|         │                                                               |
|         ├───────────────────────────────────────────────┐               |
|         ▼                                               ▼               |
|  ML Sidecar Inference (1.5µs push)             Bounded Ring Buffer      |
|  (fraud-xgb-25f-v3.0)                          (DriftCollector)         |
|         │                                               │               |
|         ▼                                               │ (Async)       |
|  Risk Decision Response                                 ▼               |
+──────────────────────────────────────────┼──────────────────────────────+
                                           │ Periodic Snapshot (5m)
                                           ▼
                             +───────────────────────────+
                             |       DriftDetector       |
                             |   Background Worker Loop  |
                             +─────────────┬─────────────+
                                           │
                    ┌──────────────────────┼──────────────────────┐
                    ▼                      ▼                      ▼
           Math Calculations       State Transitions       ClickHouse
           (PSI, JSD, KL,          (Healthy, Warning,      Audit Tables
            Moments, Percentiles)   Degraded, Critical)    (Non-PII stats)
```

---

## 3. Mathematical Foundations

### 3.1 Population Stability Index (PSI)
$$\text{PSI} = \sum_{i=1}^{K} (A_i - E_i) \cdot \ln\left(\frac{A_i}{E_i}\right)$$
- $A_i$: Live/actual empirical probability in bin $i$
- $E_i$: Expected reference baseline probability in bin $i$
- Both distributions are normalized and smoothed with $\epsilon = 10^{-4}$ to prevent division-by-zero or $\ln(0)$ on empty buckets.

### 3.2 Kullback-Leibler Divergence (KL)
$$D_{KL}(P \parallel Q) = \sum_{i=1}^{K} P(i) \cdot \ln\left(\frac{P(i)}{Q(i)}\right)$$
Measures asymmetric information entropy divergence from reference distribution $Q$ to target distribution $P$.

### 3.3 Jensen-Shannon Divergence (JSD)
$$M = \frac{1}{2}(P + Q)$$
$$\text{JSD}(P \parallel Q) = \frac{1}{2} D_{KL}(P \parallel M) + \frac{1}{2} D_{KL}(Q \parallel M)$$
- Symmetric: $\text{JSD}(P \parallel Q) = \text{JSD}(Q \parallel P)$
- Strictly bounded: $\text{JSD} \in [0, \ln(2)] \approx [0, 0.69315]$

---

## 4. Severity Classification & Thresholds

| Severity Level | Feature PSI Threshold | Feature JSD Threshold | Model Action / Operational Recommendation |
| :--- | :--- | :--- | :--- |
| **STABLE / HEALTHY** | $< 0.10$ | $< 0.05$ | Baseline stable; distributions aligned. |
| **WARNING** | $0.10 \le \text{PSI} < 0.20$ | $0.05 \le \text{JSD} < 0.10$ | Slight distribution divergence; continue automated monitoring. |
| **HIGH / DEGRADED** | $0.20 \le \text{PSI} < 0.30$ | $0.10 \le \text{JSD} < 0.15$ | Moderate feature drift observed. Review incoming traffic profiles. |
| **CRITICAL** | $\ge 0.30$ | $\ge 0.15$ | Major distribution shift detected across critical features. Model review & retraining recommended. |

### Overall Model Status Derivation:
- **CRITICAL**: $\ge 2$ features CRITICAL or $\text{max(PSI)} \ge 0.30$.
- **DEGRADED**: $\ge 1$ feature CRITICAL or $\ge 4$ features WARNING/HIGH or $\text{max(PSI)} \ge 0.20$.
- **WARNING**: $\ge 1$ feature WARNING or $\text{max(PSI)} \ge 0.10$.
- **HEALTHY**: Otherwise.

---

## 5. Safety Boundary & Circuit Breaker Decoupling

> [!IMPORTANT]
> **Drift detection is strictly observational in Phase 3.15.**
> It never autonomously modifies the Canary router rollout percentage, alters runtime model routing, or triggers Circuit Breaker trips.
> Runtime inference availability and failure isolation remain the sole domain of the `CircuitBreaker`.

---

## 6. Baseline Provenance & Schema

Reference statistics are embedded directly via `drift_baseline_25f.json` and loaded on subsystem initialization.

- **Baseline ID**: `base_dev_ieee_25f_v1`
- **Model Version**: `fraud-xgb-25f-v3.0`
- **Feature Contract**: `fraud-risk-25f-v2.5`
- **Calibration Version**: `beta-calibrated-v2.5`
- **Baseline Source**: `development_fixture`
- **Feature Count**: 25 canonical features

---

## 7. HTTP API Contracts

### 7.1 `GET /v1/drift/status`
Returns complete real-time status, baseline metadata, window capacities, metrics, and per-feature breakdowns:
```json
{
  "status": "HEALTHY",
  "model_version": "fraud-xgb-25f-v3.0",
  "feature_contract": "fraud-risk-25f-v2.5",
  "baseline_id": "base_dev_ieee_25f_v1",
  "baseline_source": "development_fixture",
  "total_samples_ingested": 12500,
  "window": {
    "sample_count": 10000,
    "max_capacity": 10000
  },
  "metrics": {
    "max_psi": 0.0421,
    "max_jsd": 0.0125,
    "max_kl": 0.0210
  },
  "max_psi": 0.0421,
  "max_jsd": 0.0125,
  "max_kl": 0.0210,
  "drifted_features": 0,
  "critical_features": 0,
  "last_calculated_at": "2026-08-21T12:35:00Z",
  "recommendation": "NONE",
  "features": [ ... ]
}
```

### 7.2 `GET /v1/drift/history`
Returns bounded historical measurement records (up to 50 entries) representing time-series evaluations.

### 7.3 `POST /v1/drift/evaluate`
Forces an immediate on-demand drift calculation over the current sample window.

### 7.4 `GET /v1/system/status`
Extends consolidated operational system status with the drift telemetry summary:
```json
{
  "status": "HEALTHY",
  "production_model": "fraud-xgb-25f-v3.0",
  "drift": {
    "status": "HEALTHY",
    "max_psi": 0.0421,
    "max_jsd": 0.0125,
    "drifted_features": 0,
    "critical_features": 0,
    "last_calculated_at": "2026-08-21T12:35:00Z",
    "baseline_source": "development_fixture"
  }
}
```

---

## 8. ClickHouse Storage & Schemas

No PII, PAN, CVV, raw tokens, or IP addresses are persisted. Only aggregated feature metrics and state events are written.

### 8.1 `drift_baselines`
```sql
CREATE TABLE IF NOT EXISTS drift_baselines (
    baseline_id String,
    model_version String,
    feature_contract String,
    calibration_version String,
    dataset_version String,
    created_at DateTime,
    feature_count UInt16,
    metadata String
) ENGINE = MergeTree()
ORDER BY (created_at, baseline_id);
```

### 8.2 `drift_measurements`
```sql
CREATE TABLE IF NOT EXISTS drift_measurements (
    measurement_id String,
    timestamp DateTime,
    model_version String,
    baseline_id String,
    evaluation_window UInt32,
    sample_count UInt32,
    overall_status String,
    max_psi Float64,
    max_jsd Float64,
    max_kl Float64,
    drifted_feature_count UInt16,
    critical_feature_count UInt16
) ENGINE = MergeTree()
ORDER BY (timestamp, measurement_id);
```

### 8.3 `drift_feature_measurements`
```sql
CREATE TABLE IF NOT EXISTS drift_feature_measurements (
    measurement_id String,
    timestamp DateTime,
    feature_name String,
    sample_count UInt32,
    psi Float64,
    jsd Float64,
    kl Float64,
    baseline_mean Float64,
    live_mean Float64,
    mean_shift Float64,
    baseline_std Float64,
    live_std Float64,
    std_shift Float64,
    missing_rate Float64,
    severity String
) ENGINE = MergeTree()
ORDER BY (timestamp, measurement_id, feature_name);
```

### 8.4 `drift_events`
```sql
CREATE TABLE IF NOT EXISTS drift_events (
    event_id String,
    timestamp DateTime,
    model_version String,
    baseline_id String,
    previous_status String,
    new_status String,
    max_psi Float64,
    max_jsd Float64,
    max_kl Float64,
    affected_feature_count UInt16,
    critical_feature_count UInt16,
    trigger String,
    reason String
) ENGINE = MergeTree()
ORDER BY (timestamp, event_id);
```

---

## 9. Performance Benchmarks

Measured on Apple M4 (ARM64, 10 Cores):

| Operation | Latency / op | Memory Allocations | Impact on Request Path |
| :--- | :--- | :--- | :--- |
| `DriftCollector.PushVector` (25 Features) | **1.61 µs** | 0 B / op, 0 allocs | Synchronous (negligible, $< 2\mu s$) |
| `DriftDetector.IngestVector` | **1.58 µs** | 0 B / op, 0 allocs | Synchronous ($< 2\mu s$) |
| `CalculatePSI` (10 bins) | **90.8 ns** | 0 B / op, 0 allocs | Background worker |
| `CalculateJSDivergence` | **299.4 ns** | 240 B / op, 3 allocs | Background worker |
| `CalculateFeatureDrift` (1,000 samples) | **22.92 µs** | 8,592 B / op, 6 allocs | Background worker |
| `EvaluateLiveDrift` (Full 25-feature sweep) | **396.2 µs** | 432 KB / op, 212 allocs | Background worker (every 5m) |

---

## 10. Operational Runbook

### Diagnosis on CRITICAL / DEGRADED Drift
1. Query current drift breakdown:
   ```bash
   curl -s http://localhost:8080/v1/drift/status | jq '{status: .status, max_psi: .max_psi, features: [.features[] | select(.severity == "CRITICAL" or .severity == "HIGH")]}'
   ```
2. Check historical divergence in ClickHouse:
   ```sql
   SELECT feature_name, psi, jsd, live_mean, baseline_mean, mean_shift, severity
   FROM drift_feature_measurements
   ORDER BY timestamp DESC, psi DESC
   LIMIT 10;
   ```
3. Check state transition triggers:
   ```sql
   SELECT timestamp, previous_status, new_status, trigger, reason
   FROM drift_events
   ORDER BY timestamp DESC
   LIMIT 5;
   ```
4. If drift is caused by external macroeconomic/fraud pattern changes, trigger model retraining pipeline.
