# Training-to-Production Feature Parity & Data Flow

**Document Version:** 1.0 (Phase -1 Baseline)  
**Evaluated Components:** `backend/internal/riskengine/orchestrator.go`, `backend/internal/features/velocity.go`, `ml-service/train.py`, `ml-service/serve.py`  

---

## 1. Authoritative Feature Data Flow Diagram

```
                       [ TRANSACTION REQUEST ]
             (amount, currency, ip_address, device_fingerprint, token)
                                   │
                                   ▼
                   [ Go Orchestrator Feature Extraction ]
              (backend/internal/riskengine/orchestrator.go)
                                   │
       ┌───────────────────────────┴───────────────────────────┐
       ▼                                                       ▼
 [ In-Memory Request Fields ]                     [ Redis 7 Sorted Set Queries ]
 • amount (int64)                                 • ip_velocity_1h (ZCOUNT 1h)
 • currency (string)                              • token_velocity_24h (ZCOUNT 24h)
 • device_fingerprint (string)
 • ip_address (string)
       │                                                       │
       └───────────────────────────┬───────────────────────────┘
                                   │
                                   ▼
                      [ Go ML Client Payload Builder ]
                     (MLPredictRequest Struct Serializer)
                                   │
                                   ▼
                    [ HTTP POST /predict (FastAPI) ]
                                   │
                                   ▼
                   [ Python Feature Vector Constructor ]
                   (ml-service/serve.py: lines 118-136)
                   Constructs 1x5 Float32 Tensor:
                   [amount, ip_vel, token_vel, new_dev, hour]
                                   │
                                   ▼
                      [ ONNX Runtime XGBoost Model ]
                    (ml-service/model/fraud_model.onnx)
                                   │
                                   ▼
                      [ Raw Probability p in [0, 1] ]
                                   │
                                   ▼
             [ Risk Score Calculation: int(round(p * 100)) ]
                                   │
                                   ▼
                     [ Go Decision Engine Thresholds ]
            (ALLOW < 45 | STEP_UP < 65 | REVIEW < 85 | DECLINE >= 85)
```

---

## 2. Comprehensive Feature Parity Matrix

Every feature in the repository is audited below and categorized with color status:

- 🟢 **GREEN**: Identical semantic definition, units, and distribution between training and production serving.
- 🟡 **YELLOW**: Minor discrepancy, indirect estimation, or localized timezone assumption.
- 🔴 **RED**: Severe training-serving skew or architectural mismatch that distorts inference accuracy.
- ⚪ **GRAY**: Documented in architecture or ADRs, but completely missing/unimplemented in code.

---

### Feature Status Audit Table

| Feature Name | Status | Training Generator Semantics | Production Serving Semantics | Severity & Impact Analysis |
| :--- | :--- | :--- | :--- | :--- |
| `ip_velocity_1h` | 🟢 **GREEN** | Generated via `np.random.poisson(0.8)`. Represents transaction count from IP in 1 hour. | Queried from Redis Sorted Set `vel:ip:{tenant}:{ip}` with sliding window `[now - 1h, now]`. | **Low Risk:** Semantics match. Sliding window accurately captures rolling hourly rate. |
| `token_velocity_24h` | 🟢 **GREEN** | Generated via `np.random.poisson(1.5)`. Represents transaction count for card/token in 24 hours. | Queried from Redis Sorted Set `vel:tok:{tenant}:{token}` with sliding window `[now - 24h, now]`. | **Low Risk:** Semantics match. Sliding window accurately captures rolling 24-hour card velocity. |
| `hour_of_day` | 🟡 **YELLOW** | Uniform integer in range `[0, 23]`. High night-risk injected for hours `1..4`. | Extracted from current server UTC hour: `datetime.utcnow().hour`. | **Medium Risk:** Server UTC time is used instead of user or merchant local timezone. A 2:00 PM IST transaction is evaluated as 8:30 AM UTC. |
| `amount` | 🟡 **YELLOW** | Synthetic continuous values from `np.random.exponential(2500) + 100`. | Input payload `req.Amount` (integer base units / paise / cents). | **Medium Risk:** Training dataset has no multi-currency exchange rate normalization. A 5,000 USD transaction is numerically smaller than a 10,000 INR transaction despite being worth ~40x more. |
| `is_new_device` | 🔴 **RED** | Binary flag `np.random.binomial(1, p=0.15)` simulating unknown device hardware. | Inferred in Go using crude string heuristic: `len(device_fingerprint) < 8 || device_fingerprint == "new_device"`. | **CRITICAL SKEW:** Production does not look up whether the device fingerprint has been seen before for the user/tenant. Any realistic 32-character browser hash is treated as `is_new_device = 0` (known device), disabling fraud detection for new device takeovers! |
| `user_account_age_days` | ⚪ **GRAY** | Documented in `docs/architecture.md` feature store diagram. | **Not Implemented:** 0 code references in Go or Python. | **Documented Gap:** No account age feature is extracted or queried. |
| `chargeback_rate_30d` | ⚪ **GRAY** | Documented in `docs/architecture.md` analytical sink section. | **Not Implemented:** Dispute webhook stores records in PostgreSQL, but no rolling 30-day chargeback rate feature is served to ML. | **Documented Gap:** Feedback loop does not update real-time feature vector. |
| `device_ip_graph_degree` | ⚪ **GRAY** | Documented in initial roadmap. | **Not Implemented:** No graph database or entity resolution engine. | **Documented Gap:** Graph network intelligence is not implemented. |
| `billing_shipping_mismatch`| ⚪ **GRAY** | Documented in fraud rule catalog. | **Not Implemented:** Request schema does not contain separate billing and shipping addresses. | **Documented Gap:** Address verification features are absent. |

---

## 3. Summary of Critical ML Vulnerabilities

1. **New Device Detection Failure (🔴 RED)**:
   Because `is_new_device` relies on string matching `"new_device"`, real telemetry hashes generated by FingerprintJS are always classified as known trusted devices.
2. **Currency Unit Normalization Missing (🟡 YELLOW)**:
   Amounts are passed without currency exchange normalization to a common baseline (e.g. USD or INR).
3. **Synthetic Baseline Simplicity (🔴 RED)**:
   The model was trained on synthetic Poisson/Exponential random variables with simplistic linear risk additions, resulting in low ROC-AUC (0.6110) on complex multi-factor fraud patterns.
