# Risk Score & Probability Semantics Contract

**Document Version:** 1.0 (Phase 2 Specification)  
**Contract Author:** Antigravity Autonomous Pair Programmer  

---

## 1. Core Terminology & Architectural Separation

To prevent confusion between statistical machine learning outputs and business decision policies, the system explicitly defines three distinct metrics:

```
  [ RAW MODEL PREDICTION ] ──► raw_probability (0.0 to 1.0)
             │
             ▼ (Beta Calibration: logit(p_cal) = a*ln(p) - b*ln(1-p) + c)
  [ EMPIRICAL POSTERIOR ]  ──► calibrated_probability (0.0001 to 0.9999)
             │
             ▼ (Business Policy & Cost Mapping)
  [ PRESENTATION METRIC ]  ──► risk_score (0 to 100 integer)
```

---

## 2. Formal Semantic Definitions

| Term | Symbol / Type | Mathematical Definition | Purpose & Permitted Usages |
| :--- | :--- | :--- | :--- |
| **Raw Model Probability** | $p_{\text{raw}} \in [0.0, 1.0]$ (`float64`) | Direct output of the uncalibrated XGBoost model: $\sigma\big(\sum \text{trees}(x)\big)$. | **Model diagnostics only.** Distorted by class rebalancing (`scale_pos_weight`). **NEVER** use directly for financial expected loss calculations. |
| **Calibrated Probability** | $p_{\text{cal}} \in [0.0001, 0.9999]$ (`float64`) | Empirical posterior $P(\text{fraud} = 1 \mid x)$ computed via Beta Calibration fitted on the validation set. | **Financial risk calculation.** Use directly in cost-sensitive loss formulas: $E[\text{loss}] = p_{\text{cal}} \times \text{Amount}$. |
| **Risk Score** | $\text{score} \in [0, 100]$ (`int`) | Monotonic integer representation: $\text{int}\big(\text{round}(p_{\text{cal}} \times 100)\big)$. | **Human presentation & analyst UI.** Displayed on dashboards, cases, and rules thresholds for intuitive decisioning. |

---

## 3. Serving Schema Field Contract

The ML sidecar HTTP response (`POST /predict`) returns all three metrics to maintain full transparency and backward compatibility:

```json
{
  "risk_score": 12,
  "probability": 0.1198,
  "raw_probability": 0.6402,
  "calibrated_probability": 0.1198,
  "expected_costs": {
    "ALLOW": 239.60,
    "MANUAL_REVIEW": 111.98,
    "DECLINE": 440.10
  },
  "policy_recommended_action": "MANUAL_REVIEW",
  "reason_codes": [
    "MODEL_SIGNAL:HIGH_FRAUD_PROBABILITY",
    "POLICY_SIGNAL:EXPECTED_LOSS_ABOVE_REVIEW_THRESHOLD"
  ],
  "feature_attributions": {
    "amount": 0.45,
    "ip_velocity_1h": 0.12
  },
  "latency_ms": 1.41,
  "runtime": "onnxruntime+calibrated"
}
```
