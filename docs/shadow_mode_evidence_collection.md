# ROPUS Platform — Shadow Mode Evidence Collection & Promotion Pipeline

## 1. Executive Summary & Governance Mandate

To prevent premature promotion, ensure zero customer risk, and accumulate empirically defensible production telemetry, the ROPUS risk platform operates `extended_catboost_58f` strictly in **0%-Authority Shadow Evaluation Mode**.

- **Active Production Champion**: `fraud-xgb-25f-v3.0` (**100% Customer Decision Authority**)
- **Active Shadow Candidate**: `extended_catboost_58f` (**0% Customer Decision Authority**)
- **Dataset Availability**: **`FULL_DATASET_UNAVAILABLE`** (Pending mounting of raw 590k IEEE-CIS files)
- **Frozen Fixture Integrity**: Checksum `a30a387ad0fa8743599d6043120be6bd66ac17184b8eee4fb9ce764970201d44` (**Verified Unmodified**)
- **Current Operational Verdict**: **`REMAIN_IN_SHADOW_MODE`**

---

## 2. Six Promotion Gates Operational Status

Every promotion gate is evaluated from real telemetry and classified into an explicit machine-readable status enum (`NOT_POPULATED`, `INSUFFICIENT_SAMPLE`, `OFFLINE_VALIDATED`, `LIVE_VALIDATED`, `FAILED`, `PASSED`):

| Gate # | Promotion Gate Name | Operational Mandate / Threshold | Current Value | Sample Size ($N$) | Current Status | Blocker & Next Step |
| :---: | :--- | :--- | :---: | :---: | :---: | :--- |
| **Gate 1** | **Live Traffic Volume** | $\ge 10,000$ live production transactions | 0 tx | $N = 0$ | `NOT_POPULATED` | Live merchant traffic has not yet been accumulated. |
| **Gate 2** | **Confirmed Fraud Labels** | $\ge 50$ matured chargeback outcomes | 0 labels | $M = 0$ | `NOT_POPULATED` | Awaiting 60-day maturation of live chargeback feed. |
| **Gate 3** | **Economic BMR Advantage** | Statistically significant monetary loss reduction ($p < 0.05$) | Saves $649–$1,293 / tier ($p=0.028$) | $N = 1,200$ | `OFFLINE_VALIDATED` | Validated on frozen holdout; awaiting live cohort maturation. |
| **Gate 4** | **Calibration Quality** | Live Expected Calibration Error $\text{ECE} \le 0.010$ | $\text{ECE} = 0.0035$ | $N = 1,200$ | `OFFLINE_VALIDATED` | Beta calibrator validated on holdout; awaiting live drift verification. |
| **Gate 5** | **Inference Latency SLA** | Candidate P99 inference latency $\le 5.0\text{ ms}$ | P99 = 0.275 ms | $N = 100$ | `OFFLINE_VALIDATED` | Sidecar benchmark verified; awaiting multi-tenant concurrency soak. |
| **Gate 6** | **Action Disagreement Bound** | Action disagreement rate $\le 15.0\%$ | Disagreement = 8.33% | $N = 1,200$ | `OFFLINE_VALIDATED` | Validated on holdout; awaiting live merchant distribution. |

---

## 3. Telemetry Schema & Storage Architecture

```
+---------------------------------------------------------------------------------------------------+
|                                SHADOW TELEMETRY CONTRACT (Zero PII)                                |
+---------------------------------------------------------------------------------------------------+
|  Field Name                        Type       Description                                         |
|  -----------------------------------------------------------------------------------------------  |
|  evaluation_id                     string     Immutable prediction UUID (Primary Key)             |
|  tenant_id                         string     Tenant / merchant identifier                        |
|  timestamp                         float64    Epoch timestamp at prediction time (< T)            |
|  champion_model_version            string     "fraud-xgb-25f-v3.0"                                |
|  candidate_model_version           string     "extended_catboost_58f"                             |
|  champion_probability              float64    Champion calibrated fraud probability               |
|  candidate_raw_probability         float64    Candidate uncalibrated tree score                   |
|  candidate_calibrated_probability  float64    Candidate Beta-calibrated probability               |
|  champion_action                   string     ALLOW | STEP_UP | MANUAL_REVIEW | DECLINE           |
|  candidate_action                  string     ALLOW | STEP_UP | MANUAL_REVIEW | DECLINE           |
|  action_disagreement               bool       true if champion_action != candidate_action         |
|  amount_usd                        float64    Transaction amount in USD for BMR evaluation        |
|  latency_ms                        float64    Candidate inference latency in milliseconds         |
|  queue_wait_ms                     float64    Time task waited in buffered worker channel         |
|  shadow_execution_status           string     SUCCESS | TIMEOUT | MALFORMED | QUEUE_SATURATED     |
+---------------------------------------------------------------------------------------------------+
```

---

## 4. Invariant Verification & Fault-Tolerant Isolation

1. **Zero Customer Decision Authority**: The candidate score and recommended action are computed in asynchronous background goroutines post-persistence. They are never written to `RiskEvaluationResponse`.
2. **Non-Blocking Resilience**: Candidate HTTP 500/503 errors, socket timeouts, or queue backpressure drop cleanly without delaying the champion or returning errors to checkout clients.
3. **Outcome Attribution & Maturation**: Chargebacks arriving $>60$ days post-transaction are associated with `evaluation_id` in a separate label table. Labels never flow backward into feature vectors. Duplicate chargebacks execute as `IDEMPOTENT_NOOP`.
4. **Continuous-Peeking Protection**: Promotion evaluation requires a minimum fixed-horizon sample of $\ge 10,000$ transactions and $\ge 50$ chargebacks before hypothesis testing is permitted, eliminating multi-testing false discovery.
5. **Maker-Checker Dual Control**: System promotion code locks candidate models from production deployment without cryptographic signoff from both Risk Director and MLOps Lead.

---

## 5. Machine-Readable Governance Record

The canonical machine-readable status is updated dynamically in [`ml-service/evaluation/live_shadow_evidence_report.json`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/ml-service/evaluation/live_shadow_evidence_report.json).
