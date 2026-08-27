# ROPUS Platform — Model Promotion Governance Record

## 1. Candidate Model Identification & Metadata

- **Candidate Model Name**: `extended_catboost_58f`
- **Candidate Feature Contract**: `MLFeatureContractV25` (58 Point-in-Time Features)
- **Champion Model Name**: `fraud-xgb-25f-v3.0` (25 Canonical Features)
- **Evaluation Mode**: **Asynchronous Shadow Evaluation (0% Customer Decision Authority)**
- **Unconditional Production Cutover**: **`PRODUCTION_PROMOTION_BLOCKED`** (Requires 10k live tx + 50 matured chargebacks)
- **Conditional Canary Eligibility**: **`CONDITIONAL_PROMOTION_ELIGIBLE`** (Tier 1 offline validation complete, max 10% exposure)
- **Authoritative Policy**: See [`docs/conditional_promotion_governance_decision.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/conditional_promotion_governance_decision.md)

---

## 2. Six Promotion Gates Evaluation Matrix

| Gate # | Promotion Gate Name | Threshold Mandate | Operational Evaluation Status | Evidence Basis & Metric |
| :---: | :--- | :--- | :---: | :--- |
| **Gate 1** | **Live Traffic Volume** | $\ge 10,000$ transactions | `NOT_POPULATED` | Currently 0 / 10,000 live production transactions accumulated. |
| **Gate 2** | **Confirmed Fraud Outcomes** | $\ge 50$ chargebacks | `NOT_POPULATED` | Currently 0 / 50 matured confirmed chargeback labels accumulated ($>60\text{d}$ maturation). |
| **Gate 3** | **Economic BMR Advantage** | Statistically significant loss reduction ($p < 0.05$) | `OFFLINE_VALIDATED` | Candidate reduces expected loss by \$649–\$1,293 on \$1k–\$10k tiers ($p = 0.028$ on frozen holdout). Awaiting live verification. |
| **Gate 4** | **Calibration Quality** | Test Expected Calibration Error ($\text{ECE} \le 0.010$) | `OFFLINE_VALIDATED` | Beta calibration achieves $\text{ECE} = 0.0035 \le 0.010$ on validation/test holdout. Awaiting live verification. |
| **Gate 5** | **Inference Latency SLA** | P99 latency $\le 5.0\text{ ms}$ | `OFFLINE_VALIDATED` | Sidecar benchmark demonstrates P99 = 0.275 ms $\le 5.0\text{ ms}$. Awaiting multi-tenant live traffic stream. |
| **Gate 6** | **Action Disagreement Bound** | Disagreement Rate $\le 15.0\%$ | `OFFLINE_VALIDATED` | Holdout evaluation measures 8.33% action disagreement at $\tau^* = 0.220 \le 15.0\%$. Awaiting live merchant stream. |

---

## 3. Cryptographic Governance & Dual-Control Log

```json
{
  "governance_event": "shadow_evaluation_gate_check",
  "evaluated_at": "2026-08-27T17:22:40Z",
  "champion_model": "fraud-xgb-25f-v3.0",
  "candidate_model": "extended_catboost_58f",
  "frozen_holdout_sha256": "a30a387ad0fa8743599d6043120be6bd66ac17184b8eee4fb9ce764970201d44",
  "governance_status_flags": {
    "MODEL_VALIDATED": true,
    "PRODUCTION_SHADOW_READY": true,
    "LIVE_EVIDENCE_UNAVAILABLE": true,
    "CONDITIONAL_PROMOTION_ELIGIBLE": true,
    "PRODUCTION_PROMOTION_BLOCKED": true
  },
  "gates_status": {
    "gate_1_live_volume": "NOT_POPULATED",
    "gate_2_confirmed_frauds": "NOT_POPULATED",
    "gate_3_economic_advantage": "OFFLINE_VALIDATED",
    "gate_4_calibration_quality": "OFFLINE_VALIDATED",
    "gate_5_inference_latency": "OFFLINE_VALIDATED",
    "gate_6_action_disagreement": "OFFLINE_VALIDATED"
  },
  "overall_verdict": "CONDITIONAL_CANARY_ELIGIBLE",
  "unconditional_promotion_allowed": false,
  "conditional_canary_eligible": true,
  "max_canary_percentage": 10
}
```

---

## 4. Promotion Criteria Checklist (Dual-Track)

### Track 1: Conditional Canary Deployment ($\le 10\%$)
- [x] Tier 1 Offline Validation Passed (BMR $p = 0.028$, Beta $\text{ECE} = 0.0035$, $\text{P99} = 0.275\text{ ms}$, Disagreement $8.33\%$).
- [x] Non-blocking fail-open shadow infrastructure verified.
- [ ] Cryptographic dual-control maker-checker signoff by Risk Director and MLOps Lead.
- [ ] Automated rollback triggers armed: Candidate error rate $\ge 1.0\%$ or P95 latency $\ge 5.0\text{ ms}$.

### Track 2: Full Unconditional Production Promotion (100% Authority)
- [ ] Gate 1: Total live production volume $\ge 10,000$ transactions.
- [ ] Gate 2: Total matured chargeback outcomes $\ge 50$ (minimum 60-day maturation window).
- [ ] Gate 3: Live BMR economic loss delta $p < 0.05$ against active champion on live outcomes.
- [ ] Gate 4: Live Expected Calibration Error $\le 0.010$ on live outcomes.
- [ ] Gate 5: Live P99 inference latency $\le 5.0\text{ ms}$ under live traffic.
- [ ] Gate 6: Live action disagreement $\le 15.0\%$ under live traffic.
