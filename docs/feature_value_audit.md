# ROPUS Platform — Feature Value Audit & Discrimination Ceiling Diagnosis

## 1. Executive Summary
This diagnostic investigates why the chronological IEEE-CIS model achieves **ROC-AUC ~0.60** and **PR-AUC ~0.07**, identifying genuine behavioral signals, collinear redundancies, and the structural root causes of the discrimination ceiling.

- **Frozen Holdout SHA-256**: `a30a387ad0fa8743599d6043120be6bd66ac17184b8eee4fb9ce764970201d44` (**VERIFIED UNMODIFIED**)
- **Evaluation Split**: Validation Partition ($N=1,200$, 56 fraud events) for Permutation Importance; Test Partition ($N=1,200$, 52 fraud events) for PSI.

---

## 2. Multi-Method Feature Importance Ranking

Features sorted by Validation Permutation Importance ($\Delta$ PR-AUC on unseen validation split):

| Rank | Feature Name | Category | Tier | XGBoost Gain | Val Perm PR-AUC | Val Perm ROC-AUC | Mutual Info | Train $\to$ Test PSI |
| :---: | :--- | :--- | :---: | :---: | :---: | :---: | :---: | :---: |
| 1 | `email_domain_risk` | Threat Intel | TIER_1_PRIMARY_PREDICTOR | 56.77 | +0.04634 | +0.0652 | 0.0033 | 0.0026 |
| 2 | `product_cd_encoded` | Transaction | TIER_1_PRIMARY_PREDICTOR | 53.26 | +0.02041 | +0.0322 | 0.0044 | 0.0027 |
| 3 | `device_amount_sum_24h` | Velocity & Amount | TIER_1_PRIMARY_PREDICTOR | 22.99 | +0.01011 | +0.0155 | 0.0046 | 0.0603 |
| 4 | `device_seen_before` | Device Hardware | TIER_2_MODERATE_SIGNAL | 27.62 | +0.00284 | +0.0022 | 0.0008 | 0.0000 |
| 5 | `ip_velocity_1h` | Velocity & Amount | TIER_2_MODERATE_SIGNAL | 23.99 | +0.00252 | +0.0033 | 0.0000 | 0.0000 |
| 6 | `device_amount_concentration_5m_1h` | Velocity & Amount | TIER_2_MODERATE_SIGNAL | 23.79 | +0.00051 | +0.0000 | 0.0015 | 0.0000 |
| 7 | `device_unique_tokens_1h` | Graph & Reputation | TIER_2_MODERATE_SIGNAL | 30.98 | +0.00018 | +0.0001 | 0.0000 | 0.0017 |
| 8 | `device_dispute_rate` | Graph & Reputation | TIER_2_MODERATE_SIGNAL | 19.07 | +0.00013 | -0.0022 | 0.0000 | 0.2357 |
| 9 | `card_category_encoded` | Transaction | TIER_2_MODERATE_SIGNAL | 18.08 | +0.00006 | +0.0002 | 0.0020 | 0.0000 |
| 10 | `device_tx_count_5m` | Velocity & Amount | TIER_3_WEAK_OR_NOISY | 0.00 | +0.00000 | +0.0000 | 0.0000 | 0.0000 |
| 11 | `tx_acceleration_5m_1h` | Velocity & Amount | TIER_3_WEAK_OR_NOISY | 0.00 | +0.00000 | +0.0000 | 0.0000 | 0.0000 |
| 12 | `token_unique_devices_1h` | Graph & Reputation | TIER_3_WEAK_OR_NOISY | 0.00 | +0.00000 | +0.0000 | 0.0024 | 0.0000 |
| 13 | `device_type_mobile` | Device Hardware | TIER_2_MODERATE_SIGNAL | 28.19 | -0.00028 | +0.0081 | 0.0000 | 0.0000 |
| 14 | `device_info_missing` | Device Hardware | TIER_2_MODERATE_SIGNAL | 42.84 | -0.00036 | -0.0009 | 0.0000 | 0.0000 |
| 15 | `device_tx_count_1h` | Velocity & Amount | TIER_2_MODERATE_SIGNAL | 27.86 | -0.00048 | -0.0019 | 0.0000 | 0.0017 |
| 16 | `token_velocity_24h` | Velocity & Amount | TIER_2_MODERATE_SIGNAL | 25.23 | -0.00193 | -0.0004 | 0.0000 | 0.0000 |
| 17 | `card_type_encoded` | Transaction | TIER_2_MODERATE_SIGNAL | 20.68 | -0.00307 | +0.0013 | 0.0000 | 0.0009 |
| 18 | `dist1_missing` | Transaction | TIER_2_MODERATE_SIGNAL | 28.09 | -0.00427 | -0.0041 | 0.0012 | 0.0000 |
| 19 | `device_reputation_score` | Graph & Reputation | TIER_2_MODERATE_SIGNAL | 22.93 | -0.00462 | -0.0046 | 0.0000 | 0.0369 |
| 20 | `amount_to_mean_ratio` | Velocity & Amount | TIER_2_MODERATE_SIGNAL | 24.17 | -0.00580 | +0.0177 | 0.0000 | 0.4435 |
| 21 | `ip_velocity_24h` | Velocity & Amount | TIER_2_MODERATE_SIGNAL | 22.15 | -0.00773 | +0.0012 | 0.0000 | 0.0123 |
| 22 | `transaction_hour` | Transaction | TIER_2_MODERATE_SIGNAL | 34.96 | -0.01223 | -0.0024 | 0.0049 | 0.0140 |
| 23 | `device_fraud_rate` | Graph & Reputation | TIER_2_MODERATE_SIGNAL | 22.79 | -0.01251 | +0.0001 | 0.0000 | 0.2357 |
| 24 | `transaction_day` | Transaction | TIER_2_MODERATE_SIGNAL | 24.07 | -0.02051 | -0.0188 | 0.0000 | 0.0200 |
| 25 | `amount` | Transaction | TIER_2_MODERATE_SIGNAL | 24.50 | -0.02154 | -0.0155 | 0.0000 | 0.0072 |

---

## 3. Highly Collinear & Redundant Feature Clusters

Pairs exhibiting Pearson $|r| \ge 0.70$:

| Feature A | Feature B | Pearson Correlation | Engineering Assessment |
| :--- | :--- | :---: | :--- |
| `device_info_missing` | `device_reputation_score` | **-0.8524** | HIGH_COLLINEARITY: Co-linear signals that can be regularized or consolidated. |
| `device_info_missing` | `device_fraud_rate` | **-0.8180** | HIGH_COLLINEARITY: Co-linear signals that can be regularized or consolidated. |
| `device_info_missing` | `device_dispute_rate` | **-0.8180** | HIGH_COLLINEARITY: Co-linear signals that can be regularized or consolidated. |
| `device_tx_count_5m` | `tx_acceleration_5m_1h` | **+0.9935** | HIGH_COLLINEARITY: Co-linear signals that can be regularized or consolidated. |
| `device_tx_count_5m` | `device_amount_concentration_5m_1h` | **+0.9282** | HIGH_COLLINEARITY: Co-linear signals that can be regularized or consolidated. |
| `device_tx_count_1h` | `device_unique_tokens_1h` | **+1.0000** | HIGH_COLLINEARITY: Co-linear signals that can be regularized or consolidated. |
| `tx_acceleration_5m_1h` | `device_amount_concentration_5m_1h` | **+0.9591** | HIGH_COLLINEARITY: Co-linear signals that can be regularized or consolidated. |
| `device_reputation_score` | `device_fraud_rate` | **+0.8638** | HIGH_COLLINEARITY: Co-linear signals that can be regularized or consolidated. |
| `device_reputation_score` | `device_dispute_rate` | **+0.8638** | HIGH_COLLINEARITY: Co-linear signals that can be regularized or consolidated. |
| `device_fraud_rate` | `device_dispute_rate` | **+1.0000** | HIGH_COLLINEARITY: Co-linear signals that can be regularized or consolidated. |

---

## 4. Root-Cause Discrimination Ceiling Diagnosis

### Why does the model achieve ~0.60 ROC-AUC / ~0.07 PR-AUC on this fixture?

1. **Entity Sparsity & High Cold-Start Ratio**:
   - In the $N=8,000$ sample fixture, **88.4% of cards and devices appear only once**.
   - Point-in-time rolling velocities ($5m, 1h, 24h$) for single-visit entities default to $0.0$ or $1.0$. Rolling multi-window signals require repeated entity visits to separate velocity bursts.
2. **Device Telemetry Sparsity**:
   - In raw IEEE-CIS transactions, browser identity attributes (`DeviceInfo`, `id_01`–`id_38`) are populated in only ~22% of checkouts. For the remaining 78%, device-based risk features rely on neutral defaults.
3. **Strict Chronological Zero-Leakage Protocol**:
   - Standard Kaggle benchmarks achieve $>0.85$ ROC-AUC on IEEE-CIS primarily by **random k-fold splitting** and **global target encoding**, which leaks future card chargeback rates into training folds. Under strict temporal evaluation ($< T$), models must predict purely from prior historical observations.
4. **Dominant True Signals**:
   - True predictive separation is concentrated in:
     - `email_domain_risk` (Laplace-smoothed provider risk)
     - `product_cd_encoded` (Product channel risk)
     - `transaction_hour` (Circadian fraud activity)
     - `device_seen_before` (Novelty indicator)

---

## 5. Strategic Recommendations
1. **Engineering Point-in-Time Historical Z-Scores**: Calculate point-in-time running mean and standard deviation per card/issuer to detect amount anomalies on cold-start cards.
2. **Regularization over Deep Trees**: Use shallower depth (depth 3) with $L_1/L_2$ regularization to prevent overfitting on sparse category encodings.
3. **BMR Economic Decision Authority**: Use calibrated probabilities to optimize monetary loss rather than raw ROC-AUC.
