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

## 2. Role-Based Approval Authority
- **Model Developer**: Authors candidate and registers dataset lineage.
- **Model Validator**: Runs offline benchmarks and statistical stability tests.
- **Compliance / Risk Officer**: Reviews fairness reports and issues cryptographic sign-off.
- **Platform Operator**: Executes canary deployment.
