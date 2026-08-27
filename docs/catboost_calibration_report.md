# ROPUS Platform — Extended CatBoost Decision-Quality & Calibration Report

## 1. Executive Summary & Probability Scale Diagnosis

- **Candidate Model**: `extended_catboost_58f`
- **Frozen Holdout SHA-256**: `a30a387ad0fa8743599d6043120be6bd66ac17184b8eee4fb9ce764970201d44` (**VERIFIED UNMODIFIED**)
- **Training Base Rate**: $4.54\%$ (254 fraud events out of 5,600).

### Why were Precision, Recall, and F1 zero at $\tau = 0.50$?
Under moderate class weighting (`scale_pos_weight=2.0`), CatBoost minimizes cross-entropy against an empirical positive base rate of $4.54\%$. As a result:
- **Raw CatBoost Test Probabilities Span $[0.027, 0.372]$** with a mean of $0.080$.
- Because maximum predicted risk is $0.372$, an uncalibrated default decision cutoff of $\tau = 0.50$ classifies zero transactions as positive (TP = 0, FP = 0), mathematically producing Precision = 0, Recall = 0, F1 = 0.
- When evaluated at the **Validation-Optimal Threshold ($\tau^* = 0.170$)**, CatBoost delivers:
  - **Precision**: 0.1300
  - **Recall**: 0.2500
  - **F1 Score**: 0.1711
  - **Fraud Captures**: 13 out of 52 total fraud events (with 100 total alerts).

---

## 2. Multi-Method Validation-Only Calibration Comparison

All calibrators were fit **strictly on Validation ($N=1,200$, Month 5)** and evaluated on the frozen chronological Test partition:

| Calibration Method | Fitting Partition | Brier Score | Expected Calibration Error (ECE) | Output Probability Range |
| :--- | :--- | :---: | :---: | :---: |
| **Raw Uncalibrated** | None | 0.0425 | 0.0363 | [0.0270, 0.3719] |
| **Platt Scaling** | Validation ($N=1,200$) | 0.0411 | 0.0033 | [0.0401, 0.0982] |
| **Isotonic Regression** | Validation ($N=1,200$) | 0.0415 | 0.0180 | [0.0001, 0.3333] |
| **Beta Calibration** (Recommended) | Validation ($N=1,200$) | **0.0408** | **0.0035** | **[0.0095, 0.3075]** |

---

## 3. Bayes Minimum Risk (BMR) Economic Loss Comparison

Expected business monetary loss across transaction amount tiers using Beta-calibrated posterior probabilities:

| Transaction Amount | Champion Expected Loss | CatBoost Expected Loss | Monetary Delta (USD) | Economic Advantage |
| :---: | :---: | :---: | :---: | :---: |
| **$10.00** | $538.34 | $556.93 | $+18.59 | CHAMPION_LOWER_LOSS |
| **$100.00** | $4,117.61 | $3,437.12 | $-680.49 | CATBOOST_LOWER_LOSS |
| **$1,000.00** | $25,314.25 | $24,021.19 | $-1,293.06 | CATBOOST_LOWER_LOSS |
| **$10,000.00** | $146,220.45 | $145,571.26 | $-649.19 | CATBOOST_LOWER_LOSS |
| **$100,000.00** | $390,032.12 | $400,416.53 | $+10,384.41 | CHAMPION_LOWER_LOSS |

---

## 4. Multi-Seed Bootstrap Stability (5 Distinct Seeds)

| Resample Random Seed | Mean $\Delta$ PR-AUC vs Champ | 95% Confidence Interval | Empirical $p$-value |
| :---: | :---: | :---: | :---: |
| Seed `42` | +0.0381 | [-0.0007, +0.0969] | **0.0280** |
| Seed `101` | +0.0385 | [-0.0006, +0.0990] | **0.0270** |
| Seed `2024` | +0.0369 | [-0.0012, +0.0958] | **0.0320** |
| Seed `777` | +0.0398 | [-0.0014, +0.0962] | **0.0300** |
| Seed `999` | +0.0394 | [-0.0026, +0.1015] | **0.0370** |
