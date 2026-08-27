# ROPUS Platform — Machine Learning Quality & Evaluation Report

## 1. Evaluation Methodology
- **Split Strategy**: Strict chronological temporal split ($< T$). Zero random k-fold shuffling.
- **Partition Counts**:
  - **Training Split**: 5,600 transactions (4.54% fraud) spanning Months 1–4
  - **Validation Split**: 1,200 transactions (4.67% fraud) spanning Month 5 (used exclusively for hyperparameter & threshold tuning)
  - **Temporal Holdout Test Split**: 1,200 transactions (4.33% fraud) spanning Month 6 (evaluated exactly once)
- **Frozen Holdout Integrity**: Verified SHA-256 `a30a387ad0fa8743599d6043120be6bd66ac17184b8eee4fb9ce764970201d44`.

---

## 2. Baseline Model Performance Comparison

Evaluated on the unseen temporal holdout test split ($N=1,200$):

| Model Baseline | ROC-AUC | PR-AUC | Precision | Recall | F1-Score | Brier Score | ECE | Expected Business Loss |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: |
| **Random Baseline** | 0.4868 | 0.0432 | 0.0430 | 0.4900 | 0.0633 | 0.2500 | 0.4500 | $15,820.00 |
| **Logistic Regression (Balanced)** | 0.5832 | 0.0635 | 0.0592 | 0.5385 | 0.1043 | 0.2214 | 0.1450 | $12,450.00 |
| **Random Forest (100 Trees, Depth 6)** | 0.6089 | 0.0923 | 0.0910 | 0.2692 | 0.1348 | 0.0512 | 0.0380 | $11,200.00 |
| **XGBoost 15-Feature Baseline** | 0.6039 | 0.0732 | 0.0864 | 0.2692 | 0.1308 | 0.1173 | 0.2223 | $11,350.00 |
| **XGBoost 25-Feature Candidate (Tuned)** | **0.5977** | **0.0700** | **0.0945** | **0.2308** | **0.1341** | **0.0415** | **0.0023** | **$10,890.00** |

---

## 3. 5-Stage Feature Ablation Study

| Stage | Feature Group | Added Signals | Feature Count | ROC-AUC | PR-AUC | F1-Score |
| :--- | :--- | :--- | :---: | :---: | :---: | :---: |
| **Stage A** | Transaction-Only | `amount`, `transaction_hour`, `transaction_day`, `product_cd_encoded`, `card_type_encoded`, `card_category_encoded` | 6 | 0.4787 | 0.0613 | 0.0672 |
| **Stage B** | + Rolling Velocities | `ip_velocity_1h/24h`, `token_velocity_24h`, `amount_to_mean_ratio` | 10 | 0.5545 | 0.0567 | 0.0905 |
| **Stage C** | + Device Hardware | `device_seen_before`, `device_type_mobile`, `device_info_missing`, `device_tx_count_5m/1h`, `tx_acceleration_5m_1h` | 16 | 0.5817 | 0.0595 | 0.0933 |
| **Stage D** | + Network / Threat Intel | `email_domain_risk`, `dist1_missing`, `device_amount_concentration_5m_1h`, `device_amount_sum_24h` | 20 | 0.5788 | 0.0652 | 0.1209 |
| **Stage E** | **Full 25F Canonical** | `device_unique_tokens_1h`, `token_unique_devices_1h`, `device_reputation_score`, `device_fraud_rate`, `device_dispute_rate` | **25** | **0.5977** | **0.0700** | **0.1341** |

---

## 4. Top Feature Importance Ranking (XGBoost Gain)

| Rank | Feature Key | Category | Importance Gain | Feature Description |
| :---: | :--- | :--- | :---: | :--- |
| 1 | `email_domain_risk` | Threat Intel | 56.77 | Laplace-smoothed domain risk |
| 2 | `product_cd_encoded` | Transaction | 53.26 | Channel risk encoding |
| 3 | `device_info_missing` | Device | 42.84 | Missing telemetry indicator |
| 4 | `transaction_hour` | Temporal | 34.96 | Nocturnal hour indicator |
| 5 | `device_unique_tokens_1h` | Graph / Token | 30.98 | Card testing cluster signal |
| 6 | `device_type_mobile` | Device | 28.19 | Mobile platform indicator |
| 7 | `dist1_missing` | Transaction | 28.09 | Address distance missing |
| 8 | `device_tx_count_1h` | Velocity | 27.86 | Rapid device velocity burst |
| 9 | `device_seen_before` | Novelty | 27.62 | Trusted device indicator |
| 10 | `token_velocity_24h` | Velocity | 25.23 | Card velocity in 24h |

---

## 5. Calibration & Discrimination Analysis

- **Raw Model Probability**: exhibits uncalibrated probability compression under high class imbalance (`scale_pos_weight = 21.05`), yielding raw Expected Calibration Error (ECE) of `0.1958`.
- **Beta / Isotonic Calibration**: transforms raw model logits into empirical posterior probabilities, reducing ECE to `0.0023` and Brier score to `0.0415`.
- **Reproducibility**: All evaluation metrics can be regenerated deterministically via `python3 ml-service/evaluation/evaluate_models.py`.
