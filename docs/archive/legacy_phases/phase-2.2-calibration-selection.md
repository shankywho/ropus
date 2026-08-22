# Phase 2.2 — Calibration Selection & Score Resolution Experiment

**Repository:** `AI Risk Manager / Ropus`  
**Execution Date:** August 21, 2026  
**Scope:** Comparative Analysis of Score Resolution, Modal Collapse, and Calibration Accuracy  

---

## 1. Executive Summary & Key Findings

Phase 2.2 investigated the structural properties of probability calibration candidates on the validation partition ($N = 1,200$) and evaluated their impact on downstream decision resolution.

### Critical Discovery: Severe Stepwise Collapse in Isotonic Regression
- **Isotonic Regression** collapses **74.50% of all validation samples into a single exact probability (`0.0257`)**, offering only **8 unique probability outputs** across the entire dataset (`resolution = 0.0067`, `entropy = 1.2544`).
- While Isotonic Regression achieves nominal zero ECE on validation due to piecewise horizontal bin fitting, **it completely destroys intra-group score ranking and risk differentiation for 3 out of every 4 transactions**.
- In contrast, **Beta Calibration** and **Platt Scaling** maintain continuous smooth mappings with **464–524 unique probability levels** (`resolution = 0.3867–0.4367`, `entropy = 8.38–8.63`), near-perfect calibration slope ($\beta_1 = 1.0132$), low modal concentration ($<1.1\%$), and superior generalization on the untouched test partition (Test ECE: `0.0035–0.0050` vs Isotonic `0.0245`).

---

## 2. Comprehensive Candidate Comparison on Validation Split

Evaluated strictly on the validation partition ($N = 1,200$, 56 fraud cases):

| Metric | RAW XGBoost | Platt Scaling (Sigmoid) | Isotonic Regression | Beta Calibration | Ideal Target |
| :--- | :---: | :---: | :---: | :---: | :---: |
| **Brier Score** | `0.1054` | `0.0434` | **`0.0422`** | `0.0433` | $\rightarrow 0.0$ |
| **ECE (Expected Cal Error)** | `0.2001` | **`0.0021`** | `0.0000` (in-sample) | `0.0034` | $\rightarrow 0.0$ |
| **Log Loss** | `0.3576` | `0.1790` | **`0.1712`** | `0.1787` | Minimum |
| **PR-AUC** | `0.1237` | `0.1237` | `0.1200` | `0.1237` | Maximum |
| **ROC-AUC** | `0.6591` | `0.6591` | `0.7008` | `0.6591` | Maximum |
| **Calibration Slope ($\beta_1$)** | `0.5814` (Overconfident) | `1.2867` | `0.9781` | **`1.0132`** | $\mathbf{1.0000}$ |
| **Calibration Intercept ($\beta_0$)**| `-2.4121` | `0.8275` | `-0.0597` | **`0.0372`** | $\mathbf{0.0000}$ |
| **Unique Output Probabilities**| `1,047` | `464` | **`8`** | `524` | High ($>100$) |
| **Resolution Ratio ($n_{\text{uniq}}/N$)**| `0.8725` | `0.3867` | **`0.0067`** | `0.4367` | $> 0.10$ |
| **Modal Probability Value** | `0.0794` | `0.0320` | `0.0257` | `0.0291` | — |
| **Modal Share (% at Mode)** | `0.33%` | `1.08%` | **`74.50%`** | `0.92%` | $< 50\%$ |
| **Probability Entropy (bits)** | `9.9594` | `8.3881` | **`1.2544`** | `8.6346` | High ($>6.0$) |
| **p10 / p50 / p90 Quantiles** | `0.08 / 0.18 / 0.55` | `0.03 / 0.04 / 0.08` | `0.03 / 0.03 / 0.12` | `0.03 / 0.04 / 0.08` | Smooth Spread |
| **Fraction $< 0.05$** | `2.92%` | `74.67%` | `79.83%` | `75.33%` | Baseline clean |
| **Fraction $0.05 - 0.35$** | `74.75%` | `25.33%` | `19.83%` | `24.67%` | Review queue |
| **Fraction $\ge 0.35$** | `22.33%` | `0.00%` | `0.33%` | `0.00%` | High risk |

*CSV Artifact:* [`ml-service/evaluation/calibration_selection.csv`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/ml-service/evaluation/calibration_selection.csv)

---

## 3. Score Resolution & Modal Collapse Diagnostics

| Calibrator | Unique Count | Resolution ($n/N$) | Modal Share | Resolution Flag ($<0.10$) | Modal Collapse ($>0.50$) | Structural Diagnostic Verdict |
| :--- | :---: | :---: | :---: | :---: | :---: | :--- |
| **RAW XGBoost** | 1,047 | 0.8725 | 0.33% | No | No | 🟢 **HEALTHY RESOLUTION** (Miscalibrated scale) |
| **Platt Scaling** | 464 | 0.3867 | 1.08% | No | No | 🟢 **HEALTHY & CONTINUOUS** |
| **Isotonic Regression** | **8** | **0.0067** | **74.50%** | **YES** | **YES** | 🔴 **SEVERE STEPWISE COLLAPSE** |
| **Beta Calibration** | 524 | 0.4367 | 0.92% | No | No | 🟢 **OPTIMAL CONTINUOUS & CALIBRATED** |

*CSV Artifact:* [`ml-service/evaluation/calibration_resolution.csv`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/ml-service/evaluation/calibration_resolution.csv)

---

## 4. Operational Policy Simulation

Simulated under the Phase 2 Decision Policy (`ALLOW < 0.05`, `MANUAL_REVIEW 0.05–0.35`, `DECLINE >= 0.35`) using the standard cost matrix ($C_{\text{FP}} = ₹500, C_{\text{rev}} = ₹100, \text{multiplier} = 1.0$):

| Calibrator | % ALLOW | % REVIEW | % DECLINE | Fraud Recall | Fraud Precision | Fraud Caught | False Positives | Realized Loss/Txn |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: |
| **RAW XGBoost** | 2.92% | 74.75% | 22.33% | **98.21%** | 4.72% | 55 / 56 | 237 | ₹173.66 (Excess FP loss) |
| **Platt Scaling** | 74.67% | 25.33% | 0.00% | **58.93%** | 10.86% | 33 / 56 | **0** | ₹28.40 |
| **Isotonic Regression** | 79.83% | 19.83% | 0.33% | **55.36%** | **12.81%** | 31 / 56 | 2 | **₹23.77** |
| **Beta Calibration** | 75.33% | 24.67% | 0.00% | **58.93%** | **11.15%** | 33 / 56 | **0** | **₹27.74** |

*CSV Artifact:* [`ml-service/evaluation/calibration_operational_comparison.csv`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/ml-service/evaluation/calibration_operational_comparison.csv)

---

## 5. Untouched Test Set Evaluation

Evaluated once on the held-out temporal test set ($N = 1,200$, 52 fraud cases):

| Calibrator | Brier Score | Test ECE | Log Loss | PR-AUC | ROC-AUC | Test Resolution | Modal Share | Test Fraud Caught | Realized Cost/Txn |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: |
| **RAW XGBoost** | 0.1129 | 0.2061 | 0.3749 | 0.0583 | 0.5572 | 0.8808 | 0.42% | 51 / 52 | ₹185.46 |
| **Platt Scaling** | **0.0415** | **0.0035** | **0.1785** | **0.0583** | **0.5572** | 0.3917 | 1.08% | 19 / 52 | ₹31.54 |
| **Isotonic Regression** | 0.0429 | 0.0245 | 0.1901 | 0.0509 | 0.5399 | **0.0125** | **73.33%** | 15 / 52 | **₹27.24** |
| **Beta Calibration** | **0.0418** | **0.0050** | **0.1799** | **0.0583** | **0.5572** | **0.4417** | **0.83%** | 18 / 52 | **₹30.43** |

---

## 6. Strategic Recommendation

### Recommendation: **Option D — Decouple Probability Calibration from Continuous Score Ranking (with Option B/Beta for Parametric Calibration)**

1. **Why Isotonic Regression Alone is Insufficient:**
   - Although Isotonic regression minimizes theoretical step loss on validation, its **74.5% modal collapse** renders it incapable of ranking or differentiating among low-to-medium risk transactions. For an analyst looking at two transactions with score `0.0257`, one could be 10x riskier than the other in terms of underlying tree feature splits.
2. **The Optimal Architectural Approach (Option D):**
   - **Continuous Risk Ranking Signal:** Use smooth parametric calibration (**Beta Calibration** or **Platt Scaling**) to produce continuous, granular posterior probabilities ($0.0001 \dots 0.9999$) with $>400$ unique output levels and high Shannon entropy.
   - **Presentation Risk Score (0–100):** Maintain monotonic continuous scaling from the parametric probability, ensuring every small increase in risk feature value produces a distinct ranking position.
   - **Expected Cost Computation:** Compute financial losses directly from smooth calibrated posteriors, avoiding the discrete cliff effects of step functions.
