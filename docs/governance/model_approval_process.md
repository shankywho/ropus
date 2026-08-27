# Formal Model Approval & Promotion Process

## 1. Governance Gate Sequence
Before any candidate model can receive production traffic, it must traverse and pass five distinct governance gates:

```text
[ CANDIDATE ARTIFACT ]
          │
          ▼
   1. Validation Gate (ROC-AUC >= 0.85, F1 >= 0.80, Brier <= 0.15)
          │
          ▼
   2. Explainability Audit (Deterministic SHAP attribution & PII scrub)
          │
          ▼
   3. Fairness Check (Disparate Impact Ratio >= 0.80 across cohorts)
          │
          ▼
   4. Security & Vulnerability Scan (Checksum verify, Non-root execution)
          │
          ▼
   5. Risk Officer Four-Eyes Sign-Off (Role: RISK_APPROVER)
          │
          ▼
    [ PRODUCTION READY ]
```

## 2. Role-Based Approval Authority (Dual-Control Maker-Checker)
- **Model Developer**: Authors candidate, runs hyperparameter optimization, and registers dataset lineage.
- **Model Validator**: Runs offline benchmarks, 1,000-iteration bootstrap significance tests, and calibration checks.
- **Risk Director / Compliance Officer**: Reviews BMR loss calculations and issues cryptographic sign-off.
- **Platform Operator / SRE**: Configures canary deployment percentage ($\le 10\%$) with automated circuit breaker monitors.

---

## 3. Master Promotion Policy
For concrete threshold numbers, statistical protocols ($p < 0.05$), latency SLAs ($\text{P99} \le 5.0\text{ ms}$), and the complete dual-track framework, see:
- [`docs/conditional_promotion_governance_decision.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/conditional_promotion_governance_decision.md) — Master Promotion Policy & Governance Decision.
- [`docs/shadow_mode_promotion_policy.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/shadow_mode_promotion_policy.md) — Concrete 6-Gate Threshold Specifications.
