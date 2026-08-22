# ML Baseline & Reality Specification

**Document Version:** 1.0 (Phase -1 Baseline)  
**Evaluated Artifacts:** `ml-service/train.py`, `ml-service/export_onnx.py`, `ml-service/serve.py`, `ml-service/model/fraud_model.onnx`  
**Evaluation Engine:** Python 3.11 + ONNX Runtime 1.29.0 + XGBoost 3.2.0 + Scikit-Learn 1.9.0  

---

## 1. Dataset & Ground-Truth Profile

| Dimension | Ground-Truth Reality | Code Reference |
| :--- | :--- | :--- |
| **Dataset Source** | **100% Synthetic Data Generator** | `ml-service/train.py:generate_synthetic_data` |
| **Data Generation Method** | Exponential distribution for amounts, Poisson for velocities, Binomial for device novelty, Uniform for hour. | `train.py:lines 20-35` |
| **Total Sample Size** | **30,000 rows** | `train.py:train_and_export_model` |
| **Train / Test Split** | **80% Train (24,000 rows) / 20% Test (6,000 rows)** | Stratified split (`random_state=42`) |
| **Random Seed** | **42** (fixed in NumPy and Scikit-Learn) | `train.py:line 20, line 64` |
| **Total Fraud Cases** | **591 fraud rows (1.97% fraud rate)** | 29,409 legitimate (98.03%) |
| **Test Set Fraud Cases** | **118 fraud rows / 5,882 legitimate rows** | Stratified split |
| **Label Definition** | `is_fraud` = `1` if random draw $<$ `min(0.95, base_prob + velocity_risk + amount_device_risk + night_risk)`, else `0`. | `train.py:lines 37-48` |

---

## 2. Model Architecture & Hyperparameters

- **Algorithm:** Extreme Gradient Boosting Classifier (`xgboost.XGBClassifier`)
- **Objective:** `binary:logistic` (`logloss`)
- **Hyperparameters:**
  - `n_estimators`: `120`
  - `max_depth`: `4`
  - `learning_rate`: `0.08`
  - `scale_pos_weight`: `49.7` (dynamically calculated as `(len(y_train) - sum(y_train)) / sum(y_train)`)
  - `tree_method`: `"hist"`
  - `random_state`: `42`
- **Serialization Formats:**
  1. `ml-service/model/fraud_model.joblib`: Python pickle containing model object, feature metadata, medians, stds, and importances.
  2. `ml-service/model/fraud_model.onnx`: Open Neural Network Exchange graph (opset 15, size: **102.1 KB**).
  3. `ml-service/model/model_metadata.json`: Feature statistics and global importance weights.
- **Serving Engine:** `onnxruntime.InferenceSession` with `CPUExecutionProvider` (intra-op threads: 2, graph optimization: `ORT_ENABLE_ALL`).

---

## 3. Measured Performance & Baseline Metrics

Evaluated on the held-out test split ($N = 6,000$) using the production ONNX Runtime session:

| Metric | Measured Value | Benchmark Interpretation |
| :--- | :--- | :--- |
| **ROC-AUC** | **0.6110** | Weak discriminative capability on complex boundaries due to simplistic synthetic generator distributions. |
| **PR-AUC (Average Precision)** | **0.2098** | Significantly above random baseline (0.0197), but high false positive rate. |
| **Precision** | **0.0474 (4.74%)** | At 0.50 decision threshold, only ~5 out of every 100 flagged transactions are true fraud. |
| **Recall (Sensitivity)** | **0.3305 (33.05%)** | Catches 39 out of 118 fraud cases in the test set. |
| **F1 Score** | **0.0829** | Low harmonic mean reflecting high false alarm volume. |
| **Specificity** | **0.8667 (86.67%)** | Correctly clears 5,098 out of 5,882 legitimate transactions. |
| **True Positives (TP)** | **39** | Correctly identified fraud cases. |
| **False Positives (FP)** | **784** | Legitimate transactions mistakenly predicted as fraud at 0.50 threshold. |
| **False Negatives (FN)** | **79** | Fraud cases missed by the model. |
| **True Negatives (TN)** | **5,098** | Clean transactions correctly identified. |

### Raw ONNX Engine Latency Benchmark (1,000 iterations):
- **p50 Latency:** `0.014 ms` (14 microseconds)
- **p95 Latency:** `0.019 ms` (19 microseconds)
- **p99 Latency:** `0.043 ms` (43 microseconds)
- **Mean Latency:** `0.016 ms`

---

## 4. Exact Feature Table & Training-Serving Skew Audit

| Feature Name | Data Type | Training Source | Serving Source (Go $\rightarrow$ ML) | Present in Go Context? | Training-Serving Skew Risk |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `amount` | `float32` / `float64` | `np.random.exponential(2500) + 100` | Extracted directly from incoming JSON payload (`req.Amount`). | **YES** | 🟡 **YELLOW** — Training uses continuous exponential values (e.g. 2641.52), while Go payload represents integer minor units (paise/cents). No currency normalization or scaling is performed. |
| `ip_velocity_1h` | `float32` | `np.random.poisson(0.8)` | Queried from Redis 7 Sorted Set: `ZCOUNT vel:ip:{tenant}:{ip} [now-1h, now]`. | **YES** | 🟢 **GREEN** — Matching temporal semantics (1-hour count of transactions from client IP). |
| `token_velocity_24h` | `float32` | `np.random.poisson(1.5)` | Queried from Redis 7 Sorted Set: `ZCOUNT vel:tok:{tenant}:{tok} [now-24h, now]`. | **YES** | 🟢 **GREEN** — Matching temporal semantics (24-hour count of transactions for instrument token). |
| `is_new_device` | `int32` (`0` or `1`) | `np.random.binomial(1, p=0.15)` | Inferred in Go: `1` if `len(device_fingerprint) < 8` or `device_fingerprint == "new_device"`, else `0`. | **YES** | 🔴 **RED (Critical Skew)** — Go uses crude length/string heuristics rather than checking a historical device fingerprint store. In production, a real visitor hash like `"9c8e1a..."` is interpreted as `is_new_device = 0` regardless of whether the user has ever seen it before! |
| `hour_of_day` | `int32` (`0-23`) | `np.random.randint(0, 24)` | Extracted from current UTC hour in Go or defaulted in Python: `datetime.utcnow().hour`. | **YES** | 🟡 **YELLOW** — Uses server UTC hour rather than merchant or cardholder local timezone. |

---

## 5. Explainability & SHAP Reality

- **Documentation Claim:** Real-time SHAP TreeExplainer kernel computing exact Shapley values on every prediction.
- **Code Reality:** `ml-service/serve.py` uses a lightweight **heuristic Z-score attribution**:
  $$\text{attribution}_i = \max\left(0, \frac{x_i - \mu_i}{\sigma_i}\right) \times w_i$$
  where $\mu_i$ is feature median, $\sigma_i$ is feature standard deviation, and $w_i$ is global XGBoost feature importance.
- **Why this was done:** Full Python `shap.TreeExplainer` adds 15–40ms of latency per request, violating the strict <100ms p99 budget. The Z-score heuristic runs in under 0.05ms.
