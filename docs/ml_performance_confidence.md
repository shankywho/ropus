# ROPUS Platform — Statistical Robustness & Confidence Report

## 1. Executive Summary
Because the frozen chronological holdout partition contains **$N = 1,200$ transactions with $52$ fraud events ($4.33\%$)**, point estimates are subject to statistical sampling variance. This report establishes **1,000-iteration 95% bootstrap confidence intervals** and paired significance tests.

- **Frozen Holdout SHA-256**: `a30a387ad0fa8743599d6043120be6bd66ac17184b8eee4fb9ce764970201d44` (**VERIFIED UNMODIFIED**)
- **Bootstrap Iterations**: 1,000 resamples with replacement.

---

## 2. 95% Bootstrap Confidence Intervals by Model

| Model Architecture | PR-AUC Mean | PR-AUC 95% CI | ROC-AUC Mean | ROC-AUC 95% CI | F1 95% CI | Paired $p$-value vs Champ |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: |
| **logistic_regression** | 0.0690 | [0.0445, 0.1012] | 0.5841 | [0.5049, 0.6621] | [0.0716, 0.1426] | 0.7260 |
| **random_forest** | 0.1029 | [0.0589, 0.1735] | 0.6087 | [0.5286, 0.6925] | [0.0860, 0.1887] | 0.0180 |
| **hist_gradient_boosting** | 0.0819 | [0.0460, 0.1340] | 0.5703 | [0.4928, 0.6511] | [0.0702, 0.1661] | 0.4220 |
| **lightgbm** | 0.0915 | [0.0541, 0.1453] | 0.6361 | [0.5628, 0.7140] | [0.0000, 0.0000] | 0.2010 |
| **catboost** | 0.0909 | [0.0587, 0.1384] | 0.6481 | [0.5759, 0.7223] | [0.0880, 0.1814] | 0.1240 |
| **current_xgboost_25f** | 0.0766 | [0.0477, 0.1177] | 0.5989 | [0.5210, 0.6759] | [0.0725, 0.2020] | — (CHAMPION) |
| **regularized_xgboost** | 0.0814 | [0.0510, 0.1258] | 0.6156 | [0.5411, 0.6941] | [0.0901, 0.1809] | 0.2970 |
| **enhanced_temporal_xgboost_41f** | 0.0757 | [0.0451, 0.1212] | 0.5781 | [0.4985, 0.6605] | [0.0677, 0.1654] | 0.5610 |

---

## 3. Paired Model-vs-Champion Statistical Significance

Evaluating whether candidate improvements over the current champion are statistically significant ($H_0: \Delta \le 0, \alpha = 0.05$):

| Candidate Model | $\Delta$ PR-AUC Mean | $\Delta$ PR-AUC 95% CI | $p$-value | Significant Improvement? |
| :--- | :---: | :---: | :---: | :---: |
| `logistic_regression` | -0.0076 | [-0.0366, +0.0178] | 0.7260 | NO (Within Sampling Variance) |
| `random_forest` | +0.0264 | [+0.0020, +0.0701] | 0.0180 | **YES** (p < 0.05) |
| `hist_gradient_boosting` | +0.0053 | [-0.0223, +0.0466] | 0.4220 | NO (Within Sampling Variance) |
| `lightgbm` | +0.0149 | [-0.0168, +0.0613] | 0.2010 | NO (Within Sampling Variance) |
| `catboost` | +0.0143 | [-0.0079, +0.0420] | 0.1240 | NO (Within Sampling Variance) |
| `regularized_xgboost` | +0.0049 | [-0.0158, +0.0255] | 0.2970 | NO (Within Sampling Variance) |
| `enhanced_temporal_xgboost_41f` | -0.0009 | [-0.0211, +0.0231] | 0.5610 | NO (Within Sampling Variance) |

---

## 4. Temporal & Subgroup Slices Performance

| Slice Name | Sample Count | Fraud Count | Fraud Rate | Champion ROC-AUC | Champion PR-AUC | Champion F1 |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: |
| `temporal_early_month_6` | 600 | 27 | 4.50% | 0.5887 | 0.0668 | 0.1348 |
| `temporal_late_month_6` | 600 | 25 | 4.17% | 0.6179 | 0.0829 | 0.1333 |
| `amount_low_under_50` | 460 | 20 | 4.35% | 0.6328 | 0.1122 | 0.1333 |
| `amount_med_50_to_200` | 534 | 21 | 3.93% | 0.6619 | 0.0741 | 0.1684 |
| `amount_high_over_200` | 206 | 11 | 5.34% | 0.4471 | 0.0513 | 0.0000 |
| `device_novel_first_seen` | 1,076 | 49 | 4.55% | 0.5964 | 0.0754 | 0.1429 |
| `device_seen_before` | 124 | 3 | 2.42% | 0.5482 | 0.0420 | 0.0000 |
