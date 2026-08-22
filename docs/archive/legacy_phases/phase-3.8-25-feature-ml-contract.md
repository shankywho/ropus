# Phase 3.8 — 25-Feature ML Feature Contract Expansion

**Repository:** `AI Risk Manager / Ropus`  
**Execution Date:** August 21, 2026  
**Document Version:** 1.0 (Phase 3.8 Implementation Specification)  
**Status:** COMPLETED & VERIFIED  

---

## 1. Executive Summary & Objective

Phase 3.8 establishes a unified, strongly typed, and canonical **25-feature ML feature vector** (`v2.5`) for production real-time fraud scoring.

### Compatibility Architecture:
```
[ TRANSACTION INGESTION ]
           │
           ▼
[ REAL-TIME FEATURE STORES ] (Phases 3.1–3.7)
  - Point-in-time multi-window velocity (10s, 1m, 5m, 15m, 1h, 6h, 24h)
  - Token-to-device graph & card testing signals (5m, 1h, 24h)
  - Deterministic device reputation & dispute attribution (30d, 90d, lifetime)
           │
           ▼
[ BUILD CANONICAL 25-FEATURE VECTOR (Contract V2.5) ]
  - Sanitized, bounded, zero-NaN guarantee, point-in-time safe
  - Stored in postgres risk_decisions feature snapshot
           │
           ▼
┌────────────────────────────────────────────────────────┐
│ EXISTING MODEL ADAPTER (ExtractLegacy15FeatureVector)  │
│ Selects exactly 15 features in original canonical order │
└────────────────────────────────────────────────────────┘
           │
           ▼
[ ML INFERENCE SIDECAR (Contract V1.5 / 15 Features) ]
  - Consumes exact 15 floats (ONNX XGBoost Model)
  - Preserves Beta Calibration mapping and Cost-Sensitive Policy
```

---

## 2. Canonical 25-Feature ML Schema (`v2.5`)

| # | Feature Name | Data Type | Source Store | Default | Bounds | Point-in-Time Safe | Description |
|---|---|---|---|:---:|:---:|:---:|---|
| **0** | `amount` | `float64` | `transaction_request` | `100.0` | `[0.01, 1e9]` | YES | Transaction monetary amount in base currency |
| **1** | `ip_velocity_1h` | `float64` | `redis:ip_velocity` | `0.0` | `[0.0, 1e5]` | YES | Prior IP transactions in last 1 hour |
| **2** | `ip_velocity_24h` | `float64` | `redis:ip_velocity` | `0.0` | `[0.0, 5e5]` | YES | Prior IP transactions in last 24 hours |
| **3** | `token_velocity_24h` | `float64` | `redis:token_velocity` | `0.0` | `[0.0, 1e5]` | YES | Prior token transactions in last 24 hours |
| **4** | `device_seen_before` | `int64` | `redis:device_feature_store` | `0.0` | `[0.0, 1.0]` | YES | 1 if device recorded prior tx, 0 if novel |
| **5** | `transaction_hour` | `int64` | `transaction_context` | `12.0` | `[0.0, 23.0]` | YES | UTC hour of day (0–23) |
| **6** | `transaction_day` | `int64` | `transaction_context` | `0.0` | `[0.0, 6.0]` | YES | UTC day of week (0–6) |
| **7** | `product_cd_encoded` | `int64` | `transaction_request` | `0.0` | `[0.0, 100.0]` | YES | Categorical product code encoding |
| **8** | `card_type_encoded` | `int64` | `transaction_request` | `0.0` | `[0.0, 100.0]` | YES | Categorical card network encoding |
| **9** | `card_category_encoded` | `int64` | `transaction_request` | `0.0` | `[0.0, 100.0]` | YES | Categorical card tier encoding |
| **10** | `email_domain_risk` | `float64` | `transaction_request` | `0.035` | `[0.0, 1.0]` | YES | Target-encoded domain risk score |
| **11** | `dist1_missing` | `int64` | `transaction_request` | `1.0` | `[0.0, 1.0]` | YES | 1 if distance to address is missing |
| **12** | `device_type_mobile` | `int64` | `transaction_request` | `0.0` | `[0.0, 1.0]` | YES | 1 if mobile device, 0 for desktop |
| **13** | `device_info_missing` | `int64` | `transaction_request` | `0.0` | `[0.0, 1.0]` | YES | 1 if device fingerprint/info missing |
| **14** | `amount_to_mean_ratio` | `float64` | `redis:device_velocity` | `1.0` | `[0.0, 1000.0]` | YES | Ratio of amount to device 24h mean |
| **15** | `device_tx_count_5m` | `int64` | `redis:device_velocity (P3.6)` | `0.0` | `[0.0, 1e4]` | YES | Point-in-time 5m device transaction count |
| **16** | `device_tx_count_1h` | `int64` | `redis:device_velocity (P3.6)` | `0.0` | `[0.0, 5e4]` | YES | Point-in-time 1h device transaction count |
| **17** | `device_amount_sum_24h` | `float64` | `redis:device_velocity (P3.6)` | `0.0` | `[0.0, 1e9]` | YES | Point-in-time 24h cumulative device volume |
| **18** | `tx_acceleration_5m_1h` | `float64` | `redis:device_velocity (P3.6)` | `0.0` | `[0.0, 1000.0]` | YES | Velocity acceleration ratio (5m vs 1h) |
| **19** | `device_amount_concentration_5m_1h` | `float64` | `redis:device_velocity (P3.6)` | `0.0` | `[0.0, 1.0]` | YES | Volume concentration ratio in 5m vs 1h |
| **20** | `device_unique_tokens_1h` | `int64` | `redis:payment_token (P3.5)` | `0.0` | `[0.0, 1000.0]` | YES | Distinct card tokens on device in 1 hour |
| **21** | `token_unique_devices_1h` | `int64` | `redis:payment_token (P3.5)` | `0.0` | `[0.0, 1000.0]` | YES | Distinct devices using token in 1 hour |
| **22** | `device_reputation_score` | `float64` | `redis:device_reputation (P3.7)` | `0.50` | `[0.0, 1.0]` | YES | Bounded deterministic reputation score |
| **23** | `device_fraud_rate` | `float64` | `redis:device_reputation (P3.7)` | `0.0` | `[0.0, 1.0]` | YES | Historical confirmed fraud rate |
| **24** | `device_dispute_rate` | `float64` | `redis:device_reputation (P3.7)` | `0.0` | `[0.0, 1.0]` | YES | Historical chargeback/dispute rate |

---

## 3. Backward Compatibility & Adapter Architecture

1. **Frozen V1.5 Contract:** Indices 0 to 14 match the existing ONNX input vector exactly in name, ordering, type, and semantic meaning.
2. **Explicit Adapter:** `ExtractLegacy15FeatureVector` extracts the 15 legacy features using canonical names and indices.
3. **ML Inference:** The existing ONNX Runtime sidecar receives only the 15 features it was trained on.
4. **Offline Training Readability:** The canonical 25-feature vector is persisted inside the PostgreSQL `risk_decisions` encrypted snapshot, creating the dataset required for **Phase 3.9 XGBoost retraining**.

---

## 4. Normalization, Sanitization & Failure Semantics

- **NaN Handling:** Converted to definition `DefaultValue`.
- **+Inf Handling:** Clamped to definition `MaxBound`.
- **-Inf Handling:** Clamped to definition `MinBound`.
- **Out-of-Bounds Handling:** Clamped strictly to `[MinBound, MaxBound]`.
- **Graceful Degradation:** Missing Redis keys or degraded stores populate safe default values (e.g. `0.50` neutral reputation, `0.0` velocity counts) without halting evaluation.

---

## 5. Performance Benchmark (200 Live Requests)

- **HTTP End-to-End Latency:**
  - p50: **`6.22ms`**
  - p95: **`11.60ms`**
  - p99: **`17.95ms`**
- **Server Internal Decision Pipeline Latency:**
  - p50: **`2.00ms`**
  - p95: **`4.00ms`**
  - p99: **`5.02ms`**

---

## 6. What is Explicitly Unchanged

- **ONNX Model Artifact:** `ml-service/model/fraud_model.onnx` unchanged.
- **Calibration Artifact:** `ml-service/model/calibration.json` (Beta Calibration) unchanged.
- **Risk Thresholds:** Allow `< 0.05`, Review `0.05–0.35`, Decline `>= 0.35` unchanged.
- **Public API Contract:** `POST /v1/risk-evaluations` request/response unchanged.
- **Database Migrations:** Schema unchanged.
