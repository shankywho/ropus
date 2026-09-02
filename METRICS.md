# 📊 ROPUS — Machine Learning & Economic Decisioning Metrics Report

> **Track 02: AI Risk Manager (Razorpay AI Buildathon)**
> **Evaluation Dataset**: Out-of-Time (OOT) Chronological Held-Out Split ($N=1,200$, 52 Positive Labels in a frozen synthetic/benchmark-derived holdout fixture, 1,148 Negative Labels).
> **Evaluation Fixture**: `ml-service/data/sample_ieee_fixture.csv` (Strictly frozen; zero leakage from training or calibration sets).
> **Canonical Evaluator Output**: `ml-service/evaluation/frozen_holdout_evaluation_report.json`
> **Dataset Checksum**: SHA-256 `af554e5a8e82ec767e60fc9c2cc3054240b1a5edd330bf3b6d5bb06809ed4702`

---

## 1. Executive Summary & Philosophy

ROPUS evaluates real-time payment transactions using **gradient boosted trees (XGBoost)** paired with **post-hoc Beta probability calibration** and **Bayes Minimum Risk (BMR) cost loss optimization**.

Rather than relying purely on static heuristics, ROPUS computes transaction-level expected financial loss:
$$\mathcal{L}(a, y) = \mathbb{E}_{y \sim P(\text{fraud} \mid x)}\big[\text{Cost}(a, y)\big]$$

Where:
- $a \in \{\text{ALLOW}, \text{STEP\_UP}, \text{MANUAL\_REVIEW}, \text{DECLINE}\}$
- Calibrated probability $P(\text{fraud} \mid x)$ accurately reflects empirical posterior risk.

---

## 2. Discrimination Performance (Frozen Held-Out Test Split)

Evaluated strictly on the held-out temporal partition ($T_{\text{test}} > T_{\text{val}} > T_{\text{train}}$) with zero future lookahead:

| Model Configuration | Model Type | ROC-AUC | PR-AUC | Recall | Precision | F1-Score | FPR | Brier Loss | ECE |
|:---|:---|:---:|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| **Random Baseline** | Random | 0.5000 | 0.0433 | 50.00% | 4.33% | 0.0797 | 50.00% | 0.2500 | 0.4500 |
| **15F Production Model (Raw)** | XGBoost 15F | 0.9376 | 0.6684 | 84.62% | 67.69% | 0.7521 | 1.83% | 0.0297 | 0.0688 |
| **15F Production Champion (Beta Calibrated)** | XGBoost 15F + Beta | **0.9376** | **0.6684** | **84.62%** | **70.97%** | **0.7719** | **1.57%** | **0.0180** | **0.0119** |
| **25F Candidate Model (Raw)** | XGBoost 25F | 0.9466 | 0.6487 | 84.62% | 67.69% | 0.7521 | 1.83% | 0.0280 | 0.0666 |
| **25F Candidate (Beta Calibrated)** | XGBoost 25F + Beta | 0.9466 | 0.6487 | 63.46% | 67.35% | 0.6535 | 1.39% | 0.0212 | 0.0180 |

### Champion Model Confusion Matrix ($N=1,200$, 52 Positive Labels, 1,148 Negative Labels at $p \ge 0.50$):
```
                       Predicted Legitimate    Predicted Positive (Fraud)
Actual Legitimate:           1130                   18       (FPR: 1.57%)
Actual Fraud:                   8                   44       (Recall: 84.62%)
```

---

## 3. Probability Calibration & Expected Calibration Error (ECE)

High class imbalance ($\approx 4.33\%$ test prevalence) causes raw tree ensemble margins to be overconfident or poorly calibrated. ROPUS applies post-hoc **Beta Calibration**:
$$\ln\left(\frac{p_{\text{cal}}}{1 - p_{\text{cal}}}\right) = a \ln(p) - b \ln(1 - p) + c$$

| Calibration Method | ECE (Expected Error) | MCE (Max Error) | Brier Score Loss | Log Loss |
|:---|:---:|:---:|:---:|:---:|
| **Raw Uncalibrated (15F)** | 0.0688 (6.88%) | 0.6340 | 0.0297 | 0.1717 |
| **Platt Scaling (Sigmoid)** | 0.0245 (2.45%) | 0.4120 | 0.0210 | 0.0910 |
| **Isotonic Regression** | 0.0210 (2.10%) | 0.3800 | 0.0205 | 0.0880 |
| **Beta Calibration (Champion)** | **0.0119 (1.19%)** | **0.3455** | **0.0180** | **0.0738** |

---

## 4. False-Positive Cost & Bayes Minimum Risk (BMR) Analysis

### Financial Cost Matrix Specification (Modeled in INR):
- **False Negative (Missed Fraud)**: $1.05 \times \text{Amount}$ (Chargeback amount + gateway fee).
- **False Positive (Declining Good Customer)**: $\min(₹500, 0.15 \times \text{Amount}) + ₹150\text{ friction penalty}$.
- **Manual Review Action**: ₹100 fixed operational triage cost.

### Policy Comparison (Held-Out Test Set, $N=1,200$ Transactions):

| Metric / Outcome | Fixed Baseline ($p \ge 0.50$) | ROPUS Calibrated BMR Policy | Policy Difference |
|:---|:---:|:---:|:---:|
| **Action Distribution** | 1,138 ALLOW / 62 DECLINE | 1,137 ALLOW / 4 REVIEW / 59 DECLINE | Dynamic triage routing |
| **Frauds Intercepted (TP)** | 44 / 52 (84.62%) | 44 / 52 (84.62%) | Parity on caught fraud |
| **False Interventions (FP)** | 18 (1.57% FPR) | 19 (1.66% FPR) | +1 review on boundary item |
| **Missed Fraud (FN)** | 8 (15.38% FNR) | 8 (15.38% FNR) | Parity on missed fraud |
| **Modeled Realized Cost** | **₹5,716.39** | **₹6,116.39** | **+7.0% Cost Difference** |

### Honest Cost Model Assessment:
> **Governance Finding**: ROPUS compares a calibrated fixed-threshold baseline with a configurable cost-sensitive policy. Under the current local cost assumptions and frozen benchmark fixture, the BMR policy increased modeled cost by 7.0% (₹6,116.39 vs ₹5,716.39) due to the ₹100 operational cost incurred on 4 manual reviews without catching additional fraud beyond the 44 already intercepted. This demonstrates why cost models require rigorous validation and governance before rollout.

---

## 5. Reproducibility & Audit Verification

To independently reproduce and verify these published numbers from the repository:

```bash
# 1. Run canonical holdout evaluation script
make evaluate
# Or: python3 ml-service/evaluation/evaluate_frozen_holdout.py

# 2. Run automated metrics consistency assertion
pytest ml-service/tests/test_metrics_consistency.py

# 3. Run full test suite across all 3 tiers
make test
```

> **Data Disclosure**: Evaluated on 52 positive labels in a frozen synthetic/benchmark-derived holdout fixture ($N=1,200$). Local prototype; no live Razorpay merchant data. No live payment-rail execution or money movement.
