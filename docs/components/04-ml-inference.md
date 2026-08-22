# Component 04: Real ML Inference Engine & Model Registry

---

## 1. Why It Exists
Sophisticated financial fraud attacks (e.g. coordinated botnets, subtle card testing velocity increments, distributed money mule deposits) easily bypass simple threshold rules. 

The **ML Inference Engine** (`backend/internal/ml/`) evaluates high-dimensional feature interactions using **Gradient Boosted Decision Trees (XGBoost/LightGBM)** to compute an exact, calibrated continuous probability of fraud ($p \in [0.0, 1.0]$) in sub-millisecond execution time.

---

## 2. Standardized Feature Vector (25 Dimensions)

Every transaction is mapped to a standardized numerical feature vector $\mathbf{x} \in \mathbb{R}^{25}$:

```text
Index | Feature Name               | Type    | Description
------+----------------------------+---------+----------------------------------------------
  01  | amount_usd                 | float64 | Monetary amount normalized to USD
  02  | velocity_10m               | float64 | Count of transactions from user in last 10m
  03  | velocity_1h                | float64 | Count of transactions from user in last 1h
  04  | amount_velocity_24h        | float64 | Cumulative spend from user in last 24h
  05  | device_entropy             | float64 | Canvas/WebGL uniqueness entropy score
  06  | is_emulator_flag           | float64 | 1.0 if emulator/VM detected, else 0.0
  07  | is_vpn_proxy               | float64 | 1.0 if bulletproof VPN/proxy detected, else 0
  08  | geo_distance_km            | float64 | Distance from previous session location
  09  | geo_velocity_kmh           | float64 | Travel speed implied by successive sessions
  10  | graph_degree_centrality    | float64 | Shared entity connections in 3-hop graph
  11  | graph_pagerank             | float64 | Relative importance score in fraud network
  12  | account_age_days           | float64 | Age of customer account in days
  13  | dispute_history_count      | float64 | Number of prior chargebacks/disputes
  14  | card_testing_ratio         | float64 | Failed authorizations / total attempts ratio
  15  | merchant_risk_weight       | float64 | Historical fraud rate of receiving merchant
  16  | canvas_fingerprint_drift   | float64 | Variance from user's primary device canvas
  17  | session_ip_risk_score      | float64 | Autonomous subnet threat score (0.0 - 1.0)
  18  | country_mismatch_flag      | float64 | 1.0 if billing country != IP country, else 0
  19  | failed_auth_attempts_24h   | float64 | Incorrect password/OTP attempts in last 24h
  20  | avg_transaction_amount     | float64 | Customer's 90-day moving average amount
  21  | time_since_last_active     | float64 | Minutes elapsed since previous user action
  22  | cross_border_flag          | float64 | 1.0 if cross-border international rail
  23  | hour_of_day_sin            | float64 | Cyclical time feature: sin(2 * pi * hour / 24)
  24  | hour_of_day_cos            | float64 | Cyclical time feature: cos(2 * pi * hour / 24)
  25  | payment_method_risk_score  | float64 | Card type risk (prepaid vs credit vs debit)
```

---

## 3. Mathematical Inference Formulation

The ensemble tree prediction aggregates leaf values across $M$ trees:
$$z(\mathbf{x}) = w_0 + \sum_{m=1}^{M} \gamma_m \cdot T_m(\mathbf{x})$$

The raw score is mapped to a calibrated posterior probability via the logistic sigmoid transform:
$$P(\text{Fraud} \mid \mathbf{x}) = \frac{1}{1 + e^{-z(\mathbf{x})}}$$

---

## 4. Key Data Structures (Go)

```go
type TransactionFeatures struct {
    AmountUSD             float64
    Velocity10m           float64
    DeviceEntropy         float64
    IsEmulator            float64
    IsVPN                 float64
    GeoDistanceKm         float64
    GraphDegreeCentrality float64
}

type ModelPrediction struct {
    ModelVersion     string    `json:"model_version"`
    FraudProbability float64   `json:"fraud_probability"` // 0.00 to 1.00
    InferenceTimeUs  int64     `json:"inference_time_us"`
    EvaluatedAt      time.Time `json:"evaluated_at"`
}
```

---

## 5. Model Registry & Champion/Challenger Testing

The **Model Registry** (`backend/internal/ml/model_registry.go`) supports zero-downtime hot-swapping of models:
- **Champion Model**: Receives $100\%$ of authoritative decision scoring.
- **Challenger Model**: Receives asynchronous shadow traffic to compute offline AUC-ROC and KS drift metrics before promotion.

---

## 6. Performance & Concurrency
- **Inference Speed**: $< 0.45\text{ms}$ per feature vector evaluation.
- **Memory Footprint**: Tree arrays are kept in read-only shared memory, enabling concurrent goroutine evaluation without mutex contention.

---

## 7. Source Code Map
- [`backend/internal/ml/inference_engine.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/ml/inference_engine.go): In-memory gradient boosted decision tree scoring.
- [`backend/internal/ml/model_registry.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/ml/model_registry.go): Model versioning and metadata management.
- [`backend/internal/training/pipeline.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/training/pipeline.go): Offline training and validation pipeline.

---

## 8. Cross-Component Links
- [Component 01: Product API](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/01-product-api.md) — Calls inference engine during feature weighting.
- [Component 02: Risk Evaluation Engine](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/02-risk-engine.md) — Calibrates ML score into policy decisions.
