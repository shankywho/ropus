# ROPUS Platform — Shadow-Mode Architecture & Model Promotion Policy

## 1. Executive Summary & Governance Principle

To eliminate confirmation bias, data leakage, and premature model deployment, ROPUS enforces a **zero-trust shadow evaluation architecture**. No model may be promoted to active production customer decision authority solely on offline benchmark scores.

- **Current Active Champion**: `fraud-xgb-25f-v3.0` (100% Customer Decision Authority)
- **Active Shadow Candidate**: `extended_catboost_58f` (0% Customer Decision Authority)
- **Frozen Holdout SHA-256**: `a30a387ad0fa8743599d6043120be6bd66ac17184b8eee4fb9ce764970201d44`

---

## 2. Shadow-Mode Runtime Architecture

```mermaid
graph TD
    Req[Incoming Transaction Request] --> Orch[Go Risk Orchestrator]
    Orch --> Feat[Extract 58-Feature Vector]
    Feat --> Champ[Active Champion: XGBoost 25F]
    Feat -.-> Shadow[Shadow Candidate: CatBoost 58F (Async)]
    Champ --> BMR[Bayes Minimum Risk Policy]
    BMR --> Dec[Active Decision: ALLOW / STEP-UP / REVIEW / DECLINE]
    Shadow -.-> ShadowLog[Shadow Telemetry Store]
    BMR -.-> ShadowLog
    ShadowLog --> Drift[Drift & Disagreement Analyzer]
```

### Runtime Telemetry Recorded Per Transaction:
1. `champion_score` & `champion_action`
2. `candidate_score` & `candidate_action`
3. `action_disagreement` (boolean flag)
4. `feature_latency_ms` & `inference_latency_ms`
5. `feature_missingness_flags`
6. `actual_outcome` (populated asynchronously upon 60-day chargeback feedback)

---

## 3. Concrete Empirical Promotion Gates & Dual-Track Policy

ROPUS establishes a **Dual-Track Governance Model** (authoritative specification in [`docs/conditional_promotion_governance_decision.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/conditional_promotion_governance_decision.md)):

### Track 1: Controlled Conditional Canary Routing ($\le 10\%$)
- **Eligibility Status**: **`CONDITIONAL_PROMOTION_ELIGIBLE`**
- **Preconditions**: Satisfy Tier 1 offline empirical validation (Gates 3–6 on frozen holdout: BMR $p < 0.05$, Beta $\text{ECE} \le 0.010$, $\text{P99} \le 5.0\text{ ms}$, Disagreement $\le 15.0\%$) + Maker-Checker signoff.
- **Traffic Exposure**: Capped at $1\%$–$10\%$ with continuous automatic rollback guards.

### Track 2: Full Unconditional Production Promotion (100% Decision Authority)
- **Status**: **`PRODUCTION_PROMOTION_BLOCKED`** (until live soak is complete)
- **Preconditions**: Candidate model will only replace the champion at $100\%$ customer authority if it satisfies **ALL SIX** quantifiable criteria on live production traffic:

| Gate # | Promotion Criterion | Required Threshold | Rationale & Justification |
| :---: | :--- | :--- | :--- |
| **Gate 1** | **Live Traffic Volume** | $\ge 10,000$ live production transactions | Ensures evaluation spans diverse geographic and merchant traffic without sample size noise. |
| **Gate 2** | **Confirmed Fraud Labels** | $\ge 50$ real confirmed chargeback outcomes | Overcomes offline fixture cold-start limitations with ground-truth production outcomes ($>60\text{d}$ maturation). |
| **Gate 3** | **Economic BMR Advantage** | Statistically significant loss reduction ($p < 0.05$) | The candidate must generate lower expected business monetary loss than the champion. |
| **Gate 4** | **Calibration Quality** | Test Expected Calibration Error ECE $\le 0.010$ | Ensures posterior probabilities safely drive cost matrix optimization. |
| **Gate 5** | **Inference Latency SLA** | P99 latency $\le 5.0\text{ ms}$ | Guarantees sidecar execution conforms to real-time checkout latency budgets. |
| **Gate 6** | **Action Disagreement Bound** | Disagreement Rate $\le 15.0\%$ | Prevents catastrophic customer friction shifts across merchant tiers. |

---

## 4. Rollback & Fail-Safe Triggers

If a promoted model exhibits any of the following anomalies in production, the orchestrator automatically triggers a circuit-breaker rollback to `fraud-xgb-25f-v3.0`:
1. Candidate Error Rate $\ge 1.0\%$ $\implies$ immediate fallback to champion.
2. Inference latency exceeding $5.0\text{ ms}$ for $> 1.0\%$ of traffic.
3. False positive rate spike $> 20\%$ over 24-hour rolling window.
4. Feature drift PSI $> 0.25$ on any primary predictor.
