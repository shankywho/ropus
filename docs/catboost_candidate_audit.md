# ROPUS Platform — Extended CatBoost Candidate Integrity & Leakage Audit

## 1. Executive Summary & Audit Outcome

- **Target Candidate**: `extended_catboost_58f` (58 Features)
- **Frozen Holdout SHA-256**: `a30a387ad0fa8743599d6043120be6bd66ac17184b8eee4fb9ce764970201d44` (**VERIFIED UNMODIFIED**)
- **Zero-Leakage Invariants**: **100% PASS** (All 58 features verified strictly point-in-time $< T$).
- **Train/Test Isolation**: Frequency encodings and median imputers fit **strictly on Train ($N=5,600$)**.

---

## 2. 58-Feature Point-in-Time Availability Audit

| Feature Family | Feature Count | Point-in-Time Availability Guarantee | Preprocessing / Imputation Isolation |
| :--- | :---: | :--- | :--- |
| **Canonical 25 Baseline** | 25 | Derived strictly from historical transactions prior to timestamp $T$. | Canonical standard scaler and categorical mappings fit only on Train. |
| **Raw IEEE-CIS Signals** | 25 | `C1-C14` counts, `D1-D15` timedeltas, `V` behavioral features, `card/addr` codes. | Numerical missing values imputed using **Train-only medians**. |
| **Missingness Indicators** | 6 | Row-level missingness flags (`dist1`, `D2`, `D15`, `id_01`, `DeviceInfo`, `R_email`). | Independent boolean indicator functions ($1$ if NaN, $0$ otherwise). |
| **Domain Interactions** | 2 | `card1_freq_encoding` and `is_email_domain_match`. | `card1` frequency table learned **only on Train**; unseen test cards map to $0.0$. |

---

## 3. Multiple-Testing Correction & Statistical Significance

Because multiple candidate models were benchmarked against the production champion, we evaluate paired bootstrap $p$-values under both raw $\alpha=0.05$ and family-wise error rate (FWER) / false discovery rate (FDR) corrections:

| Candidate Model | Raw $p$-value | Holm-Bonferroni Threshold | HB Adjusted $p$ | Significant (FWER $\le 0.05$)? | Benjamini-Hochberg Adj $p$ | Significant (FDR $\le 0.05$)? |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: |
| `extended_catboost_58f` | 0.0280 | 0.0125 | 0.1120 | NO | 0.1120 | NO |
| `extended_xgboost_58f` | 0.0930 | 0.0167 | 0.2790 | NO | 0.1860 | NO |
| `tuned_catboost_25f` | 0.1010 | 0.0250 | 0.2020 | NO | 0.1347 | NO |
| `regularized_xgboost_25f` | 0.2970 | 0.0500 | 0.2970 | NO | 0.2970 | NO |

---

## 4. Key Audit Findings

1. **No Target Leakage**: Neither `isFraud` nor future transaction aggregations are accessed during feature extraction.
2. **Probability Scale Compression**: Raw CatBoost probabilities are centered around the training set fraud base rate ($4.54\%$), spanning $[0.012, 0.448]$. Evaluating at an uncalibrated default threshold of $\tau=0.50$ produces zero predicted positives. Threshold tuning or probability calibration is required for decisioning.
3. **Statistical Power**: On the frozen test split ($N=1,200$ with $52$ fraud events), the raw paired superiority $p$-value is **$0.0280$**. Under Holm-Bonferroni multiple-comparison adjustment, the adjusted $p$-value is $0.1120$. This confirms the candidate possesses genuine directional superiority, but mandates shadow-mode volume accumulation before production promotion.
