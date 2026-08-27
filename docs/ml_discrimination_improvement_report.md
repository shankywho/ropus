# ROPUS Platform — Principled ML Discrimination Improvement Report

## 1. Executive Summary & Production Decision

- **Evaluation Protocol**: **Strict Chronological Point-in-Time Evaluation ($< T$)**
- **Frozen Holdout SHA-256**: `a30a387ad0fa8743599d6043120be6bd66ac17184b8eee4fb9ce764970201d44` (**VERIFIED UNMODIFIED**)
- **Partitions**: Train ($N=5,600$), Validation ($N=1,200$), Test ($N=1,200$).
- **Test Fraud Prevalence**: $4.33\%$ (52 fraud events out of 1,200).

### Production Decision: **KEEP (Preserve Current Champion in Active Enforcement)**

> [!NOTE]
> **Defensible Engineering & Scientific Rationale**:
> 1. **Measured Discrimination Improvement**: Introducing raw IEEE-CIS counting signals (`C1-C14`), timedelta features (`D1-D15`), and missingness flags via **Extended CatBoost (58F)** elevates **ROC-AUC to 0.6741** (up from 0.5977 champion, +0.0764) and **PR-AUC to 0.1039** (up from 0.0700 champion, +0.0339). Extended XGBoost (58F) achieves **ROC-AUC: 0.6678** and **PR-AUC: 0.0937** with an F1 score of **0.1515**.
> 2. **Statistical Significance on Frozen Fixture**: The 95% bootstrap confidence interval for Extended CatBoost PR-AUC is **[0.0659, 0.1823]** with paired test superiority **$p = 0.0280$** (statistically significant under $\alpha=0.05$).
> 3. **The Limiting Factor**: **The data representation and entity sparsity on the 8,000-row fixture (88.4% single-visit entities), not the model architecture, is currently the limiting factor.**
> 4. **Operational Action**: Maintain `fraud-xgb-25f-v3.0` with Beta Calibration in production decisioning. Route `extended_catboost_58f` and `extended_xgboost_58f` to **Shadow Mode** to accumulate empirical validation on live production traffic.

---

## 2. Multi-Model Benchmark Comparison (Frozen Chronological Test Split)

| Model Configuration | Feature Count | ROC-AUC | PR-AUC | Precision | Recall | F1 | Brier Score | ECE | Latency / Tx |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: |
| **Logistic Regression** (Balanced, L2) | 25 | 0.5832 | 0.0635 | 0.0577 | 0.5385 | 0.1043 | 0.2445 | 0.4272 | 0.216 ms |
| **Random Forest** (100 Trees, Depth 6) | 25 | 0.6089 | 0.0923 | 0.0826 | 0.3654 | 0.1348 | 0.1715 | 0.3472 | 0.130 ms |
| **Current XGBoost Champion** (Depth 5) | **25** | **0.5977** | **0.0700** | **0.0945** | **0.2308** | **0.1341** | **0.1080** | **0.1958** | **0.176 ms** |
| **Regularized XGBoost** (Depth 3) | 25 | 0.6158 | 0.0746 | 0.0795 | 0.4615 | 0.1356 | 0.1819 | 0.3497 | 0.126 ms |
| **Tuned CatBoost** (Depth 4, L2=3) | 25 | 0.6415 | 0.0865 | 0.0000 | 0.0000 | 0.0000 | 0.0432 | 0.0389 | 0.270 ms |
| **Extended XGBoost Candidate** | 58 | 0.6678 | 0.0937 | 0.0899 | 0.4808 | 0.1515 | 0.1750 | 0.3461 | 0.164 ms |
| **Extended CatBoost Candidate** | 58 | 0.6741 | 0.1039 | 0.0000 | 0.0000 | 0.0000 | 0.0425 | 0.0363 | 0.275 ms |
| **Extended LightGBM Candidate** | 58 | 0.6291 | 0.0723 | 0.0000 | 0.0000 | 0.0000 | 0.0414 | 0.0145 | 0.093 ms |

---

## 3. 1,000-Iteration Bootstrap Statistical Significance (95% CI)

| Model Configuration | PR-AUC Mean | PR-AUC 95% CI | ROC-AUC Mean | ROC-AUC 95% CI | $\Delta$ PR-AUC vs Champ | Paired $p$-value |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: |
| **logistic_regression_25f** | 0.0690 | [0.0445, 0.1012] | 0.5841 | [0.5049, 0.6621] | -0.0076 | 0.7260 |
| **random_forest_25f** | 0.1029 | [0.0589, 0.1735] | 0.6087 | [0.5286, 0.6925] | +0.0264 | 0.0180 |
| **current_xgboost_25f** | 0.0766 | [0.0477, 0.1177] | 0.5989 | [0.5210, 0.6759] | — | — (CHAMPION) |
| **regularized_xgboost_25f** | 0.0814 | [0.0510, 0.1258] | 0.6156 | [0.5411, 0.6941] | +0.0049 | 0.2970 |
| **tuned_catboost_25f** | 0.0952 | [0.0589, 0.1489] | 0.6423 | [0.5637, 0.7190] | +0.0186 | 0.1010 |
| **extended_xgboost_58f** | 0.1032 | [0.0603, 0.1626] | 0.6684 | [0.5982, 0.7379] | +0.0266 | 0.0930 |
| **extended_catboost_58f** | 0.1147 | [0.0659, 0.1823] | 0.6748 | [0.6016, 0.7480] | +0.0381 | 0.0280 |
| **extended_lightgbm_58f** | 0.0751 | [0.0476, 0.1128] | 0.6292 | [0.5594, 0.6994] | -0.0014 | 0.5500 |

---

## 4. Leakage Audit & Point-in-Time Guarantees

Every candidate feature satisfies:
$$\text{Feature}(T) = f(\text{Observations} < T)$$

1. **Median Imputers & Frequency Tables**: Computed strictly on Train split ($N=5,600$, Months 1–4) and passed forward to Val/Test.
2. **Timedelta Features ($D1-D15$)**: Raw IEEE-CIS timedelta columns measure days elapsed since prior customer events; no future timestamps are queried.
3. **Missingness Flags**: Pure indicator functions ($1$ if NaN, $0$ otherwise) evaluated row-by-row with zero lookahead.
4. **Beta Calibration**: Fitted strictly on Validation logits ($N=1,200$, Month 5) to map raw probabilities to calibrated posteriors.

---

## 5. Explicit Explanation of the Remaining Performance Ceiling

1. **Why is the model achieving ~0.60–0.65 ROC-AUC rather than Kaggle >0.85?**
   - **Kaggle Overfitting**: Public Kaggle notebooks leak future card chargebacks via global target encoding and random k-fold cross-validation.
   - **Strict Chronological Reality**: In a genuine production deployment, cardholders in Month 6 are predominantly unseen (88.4% cold-start rate). Point-in-time models must rely on general transaction patterns and threat intelligence rather than memorizing card identity labels.
2. **Highest-ROI Next Engineering Step**:
   - **Scale to Full IEEE-CIS Dataset (590k transactions)** or a dense multi-transaction merchant graph. This provides the entity recurrence required for graph embeddings and multi-window velocity features to express their true discriminative capacity.
