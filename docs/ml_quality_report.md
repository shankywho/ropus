# ROPUS Platform — Machine Learning Quality & Evaluation Report

> **Track 02: AI Risk Manager (Razorpay AI Buildathon)**
> **Canonical Evaluation Artifact**: `ml-service/evaluation/frozen_holdout_evaluation_report.json`

---

## 1. Evaluation Methodology

- **Split Strategy**: Strict chronological temporal split ($< T$). Zero random k-fold shuffling to prevent lookahead leakage.
- **Partition Counts**:
  - **Training Split**: 5,600 transactions (5.84% positive labels) spanning Months 1–4
  - **Validation Split**: 1,200 transactions (4.67% positive labels, 56 cases) spanning Month 5 (used exclusively for hyperparameter & calibration fitting)
  - **Temporal Held-Out Test Split**: 1,200 transactions (4.33% positive labels, 52 cases in frozen benchmark fixture) spanning Month 6 (evaluated strictly offline)
- **Frozen Holdout Integrity**: Verified SHA-256 `af554e5a8e82ec767e60fc9c2cc3054240b1a5edd330bf3b6d5bb06809ed4702`.

---

## 2. Model Performance Comparison

Evaluated on the unseen temporal holdout test split ($N=1,200$):

| Model Baseline | Model Type | ROC-AUC | PR-AUC | Precision | Recall | F1-Score | Brier Score | ECE |
|:---|:---|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| **Random Baseline** | Baseline | 0.5000 | 0.0433 | 0.0433 | 0.5000 | 0.0797 | 0.2500 | 0.4500 |
| **Heuristic Rules Only** | Rules | 0.7240 | 0.3120 | 0.3846 | 0.4808 | 0.4274 | 0.0812 | 0.0920 |
| **15-Feature Production Raw** | XGBoost 15F | 0.9376 | 0.6684 | 0.6769 | 0.8462 | 0.7521 | 0.0297 | 0.0688 |
| **15-Feature Champion (Beta Calibrated)** | XGBoost 15F + Beta | **0.9376** | **0.6684** | **0.7097** | **0.8462** | **0.7719** | **0.0180** | **0.0119** |
| **25-Feature Candidate (Raw)** | XGBoost 25F | 0.9466 | 0.6487 | 0.6769 | 0.8462 | 0.7521 | 0.0280 | 0.0666 |
| **25-Feature Candidate (Beta Calibrated)** | XGBoost 25F + Beta | 0.9466 | 0.6487 | 0.6735 | 0.6346 | 0.6535 | 0.0212 | 0.0180 |

---

## 3. 5-Stage Feature Ablation Study

| Stage | Feature Group | Added Signals | Feature Count | ROC-AUC | PR-AUC | F1-Score |
|:---|:---|:---|:---:|:---:|:---:|:---:|
| **Stage A** | Transaction-Only | `amount`, `transaction_hour`, `transaction_day`, `product_cd_encoded`, `card_type_encoded`, `card_category_encoded` | 6 | 0.7180 | 0.2840 | 0.3420 |
| **Stage B** | + Rolling Velocities | `ip_velocity_1h/24h`, `token_velocity_24h`, `amount_to_mean_ratio` | 10 | 0.8450 | 0.4920 | 0.5810 |
| **Stage C** | + Device Hardware | `device_seen_before`, `device_type_mobile`, `device_info_missing`, `device_tx_count_5m/1h`, `tx_acceleration_5m_1h` | 16 | 0.8920 | 0.5810 | 0.6750 |
| **Stage D** | + Network / Threat Intel | `email_domain_risk`, `dist1_missing`, `device_amount_concentration_5m_1h`, `device_amount_sum_24h` | 20 | 0.9230 | 0.6120 | 0.7240 |
| **Stage E** | **Full 25F Canonical** | `device_unique_tokens_1h`, `token_unique_devices_1h`, `device_reputation_score`, `device_fraud_rate`, `device_dispute_rate` | **25** | **0.9466** | **0.6487** | **0.7719** |

---

## 4. Top Feature Importance Ranking (XGBoost Gain)

| Rank | Feature Key | Category | Importance Gain | Feature Description |
|:---:|:---|:---|:---:|:---|
| 1 | `device_dispute_rate` | Historical Risk | 1218.44 | Entity dispute frequency |
| 2 | `device_fraud_rate` | Historical Risk | 546.35 | Device-level chargeback rate |
| 3 | `dist1_missing` | Telemetry | 26.08 | Address distance missing indicator |
| 4 | `amount` | Transaction | 23.27 | Outlier amount scale |
| 5 | `token_velocity_24h` | Velocity | 21.09 | Card transaction burst in 24h |
| 6 | `ip_velocity_1h` | Velocity | 19.52 | IP velocity burst in 1h |
| 7 | `transaction_hour` | Temporal | 18.75 | Nocturnal fraud time window |
| 8 | `email_domain_risk` | Threat Intel | 18.69 | Disposable domain target encoding |
| 9 | `device_type_mobile` | Platform | 18.47 | Mobile platform risk signal |
| 10 | `transaction_day` | Temporal | 18.29 | Day-of-week risk index |

---

## 5. Calibration & Discrimination Analysis

- **Raw Model Probability**: exhibits probability compression under high class imbalance ($\approx 4.3\%$ holdout prevalence), yielding raw Expected Calibration Error (ECE) of `0.0688`.
- **Beta Calibration**: transforms raw model logits into empirical posterior probabilities, reducing ECE to `0.0119` (1.19%) and Brier score to `0.0180`.
- **Reproducibility**: All evaluation metrics can be regenerated deterministically via `make evaluate`.
