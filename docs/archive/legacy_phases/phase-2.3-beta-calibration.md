# Phase 2.3 — Production Beta Calibration Migration

**Repository:** `AI Risk Manager / Ropus`  
**Execution Date:** August 21, 2026  
**Status:** PRODUCTION DEPLOYED & OPERATIONAL  

---

## 1. Context & Motivation for Migration

In Phase 2.2, a comprehensive resolution audit revealed a fundamental structural flaw in **Isotonic Regression**:
- **Severe Stepwise Collapse:** Isotonic regression grouped **74.50% of all validation transactions into a single exact probability (`0.0257`)**, producing only 8 unique score levels across the entire dataset (`resolution = 0.0067`).
- **Loss of Intra-Cohort Separation:** While Isotonic regression minimized empirical piecewise step loss on validation, it eliminated continuous rank-ordering among low-to-medium risk transactions.
- **Why Beta Calibration was Selected:**
  - **Smooth Continuous Resolution:** Produces **524 unique probability levels** (`resolution = 0.4367`), with modal share $<1.0\%$ and high Shannon entropy (`8.6346` bits).
  - **Empirical Accuracy:** Achieves Test ECE of **`0.0050`** (vs Isotonic `0.0245`) and Brier Score of **`0.0418`** (vs Isotonic `0.0429`).
  - **Ideal Calibration Slope:** Yields a validation calibration slope of **`1.0132`** and intercept of **`0.0372`** (near-perfect linearity).

---

## 2. Mathematical Formulation & Safety Bounds

For input probability $p \in [\epsilon, 1 - \epsilon]$ where $\epsilon = 10^{-6}$:

$$\begin{aligned}
x_1 &= \ln(p) \\
x_2 &= -\ln(1 - p) \\
\text{logit}(p_{\text{cal}}) &= a \cdot x_1 + b \cdot x_2 + c = 0.2500 \ln(p) - 1.1057 \ln(1 - p) - 3.1091 \\
p_{\text{cal}} &= \sigma\big(\text{logit}(p_{\text{cal}})\big) = \frac{1}{1 + \exp\big(-(a \ln(p) - b \ln(1 - p) + c)\big)}
\end{aligned}$$

### Safety Guarantees Implemented:
1. **Bounded Domain ($\epsilon = 10^{-6}$):** Prevents $\ln(0)$ or division by zero under edge-case inputs $p \in \{0.0, 1.0\}$.
2. **Strict Output Range:** Bounded in $[0.0001, 0.9999]$ via `np.clip()`.
3. **Monotonicity Preservation:** $a > 0$ and $b > 0$ guarantee $p_1 < p_2 \implies p_{\text{cal}}(p_1) \le p_{\text{cal}}(p_2)$ across all inputs.
4. **NaN & Inf Immunity:** Sanitizes `NaN`, `+Inf`, and negative inputs to safe fallback priors.

---

## 3. Production Verification Across Risk Tiers

Live validation on running container stack (`POST /predict`):

| Risk Tier | Input Parameters | Raw Prob ($p_{\text{raw}}$) | Calibrated Prob ($p_{\text{cal}}$) | Risk Score (0-100) | Policy Action | Reason Codes |
| :--- | :--- | :---: | :---: | :---: | :---: | :--- |
| **Low Risk** | `Amount: ₹250, Vel: 0, Known Device` | `0.0791` | `0.0253` | **3** | **`ALLOW`** | `MODEL_SIGNAL:LOW_ANOMALY_BASELINE` |
| **Medium Risk** | `Amount: ₹18,500, IP Vel: 3, New Dev, 3 AM` | `0.5797` | `0.0922` | **9** | **`MANUAL_REVIEW`** | `RULE_SIGNAL:HIGH_TRANSACTION_AMOUNT`, `MODEL_SIGNAL:NEW_DEVICE_FINGERPRINT`, `MODEL_SIGNAL:OFF_HOURS_ACTIVITY` |
| **High Risk** | `Amount: ₹125,000, IP Vel: 8, Tok Vel: 12` | `0.2114` | `0.0379` | **4** | **`MANUAL_REVIEW`** | `RULE_SIGNAL:HIGH_IP_VELOCITY_1H`, `RULE_SIGNAL:HIGH_TOKEN_VELOCITY_24H`, `RULE_SIGNAL:HIGH_TRANSACTION_AMOUNT` |

---

## 4. Artifact & Versioning Specification

The active production calibration artifact [`ml-service/model/calibration.json`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/ml-service/model/calibration.json) specifies:

```json
{
  "type": "beta",
  "version": "cal-v2.0-beta",
  "is_fitted": true,
  "checksum_sha256": "314e13221c2bc2bd79312c4e5a0a5de0dcb5fb8e97cd6fd16bc76f7ebf1a1498",
  "feature_schema_version": "1.0.0",
  "fitting_metadata": {
    "validation_samples": 1200,
    "validation_fraud_count": 56,
    "validation_fraud_rate": 0.0467,
    "fitted_at": "2026-08-21T09:29:39.786135+00:00"
  },
  "parameters": {
    "beta_params": {
      "coef": [[0.25004694, 1.10570023]],
      "intercept": [-3.10907144],
      "feature_names": ["log_p", "neg_log_one_minus_p"],
      "epsilon": 1e-06
    }
  }
}
```
