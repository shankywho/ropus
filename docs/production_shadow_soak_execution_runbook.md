# ROPUS — Production Shadow Soak Execution Runbook

---

## 1. Governance & Execution Boundaries

This runbook defines the operational procedures for running the **Live Production Shadow Soak** on the ROPUS fraud platform.

### Non-Negotiable Governance Invariants:
- **Production Champion:** `fraud-xgb-25f-v3.0` retains **100% customer decision authority**.
- **Shadow Candidate:** `extended_catboost_58f` retains **0% customer decision authority**.
- **Canary Routing:** Strictly **DISABLED** (`RISK_MODEL_CANARY_ENABLED=false`, `RISK_MODEL_CANARY_PERCENT=0`).
- **Promotion Status:** **BLOCKED (`promotion_allowed = false`)**.
- **Fail-Open Policy:** Any candidate error, timeout, or queue saturation drops the shadow task without delaying or modifying customer risk decisions.
- **Strict Provenance:** Only genuine merchant ingress is tagged `LIVE_PRODUCTION`. Synthetic, replay, or test fixtures cannot satisfy Gate 1.

---

## 2. Ingress & Telemetry Trace

```
Merchant Checkout Request (POST /v1/risk-evaluations)
  │
  ├── 1. Auth & Ingress Rate Limiter
  ├── 2. Feature Aggregation (Velocity, Device, Centrality)
  ├── 3. Synchronous Champion Scoring (fraud-xgb-25f-v3.0)
  ├── 4. Business Rules & BMR Loss Arbitration
  ├── 5. Transaction & Decision Persistence (Postgres + ClickHouse)
  ├── 6. Synchronous Response Serialization (Returned to Merchant)
  │
  └── 7. [Asynchronous Goroutine] Enqueue ShadowScoreTask
        │ (Non-blocking bounded buffer, cap=1,000)
        │
        ├── 8. CatBoost Shadow Sidecar Inference (POST /predict/shadow, timeout=100ms)
        ├── 9. Divergence & Latency Metrics Calculation
        └── 10. Immutable Telemetry Persistence (ClickHouse shadow_evaluations)
```

---

## 3. Operator Execution Protocol

### Step 1: Preflight Environmental Verification

Before routing merchant ingress traffic, verify the environment:

```bash
# Verify frozen holdout fixture integrity
shasum -a 256 ml-service/data/sample_ieee_fixture.csv
# Expected: a30a387ad0fa8743599d6043120be6bd66ac17184b8eee4fb9ce764970201d44

# Verify startup configuration safety locks
go test -v ./backend/cmd/api -run "TestValidateProductionSafetyConfig"

# Verify fault isolation and non-blocking chaos resilience
go test -v ./backend/internal/riskengine -run "TestShadow"

# Verify delayed outcome attribution and 60-day maturation logic
pytest ml-service/tests/test_shadow_outcome_attribution.py
```

Required environment variables:
```env
SHADOW_SCORING_ENABLED=true
SHADOW_CANDIDATE_MODEL_VERSION=extended_catboost_58f
SHADOW_WORKER_COUNT=4
SHADOW_QUEUE_CAPACITY=1000
SHADOW_SAMPLE_RATE=1.0
RISK_MODEL_CANARY_ENABLED=false
RISK_MODEL_CANARY_PERCENT=0
```

---

### Step 2: Continuous Observability During Soak

Run the real-time soak monitor:

```bash
python3 ml-service/evaluation/live_shadow_soak_monitor.py
```

Monitor Prometheus alerts on `GET /metrics`:
- `shadow_inference_latency_ms{quantile="0.99"} > 5.0` (Warning: Candidate latency approaching SLA)
- `shadow_queue_drops_total > 0` (Warning: Queue saturation, increase worker count)
- `shadow_errors_total` (Categorized by `TIMEOUT`, `HTTP_500`, `HTTP_503`, `CONNECTION_REFUSED`)
- `shadow_decision_divergence_total` (Track action disagreement)

---

### Step 3: Emergency Deactivation Conditions

Immediately disable shadow scoring if:
1. Champion synchronous latency regresses by $>1.0\text{ ms}$.
2. Synchronous transaction errors increase.
3. Candidate memory/CPU utilization impacts the host system.

**Emergency Deactivation Command:**
```bash
export SHADOW_SCORING_ENABLED=false
```

---

### Step 4: Outcome Maturation & Checkpoints

- For every live transaction, the immutable `evaluation_id` is recorded.
- As chargebacks arrive, attribute them via `attribute_outcome(evaluation_id, is_fraud=1)`.
- Enforce the mandatory **60-day maturation window** before counting outcomes toward Gate 2.
- Generate checkpoint audit reports at $N = 1,000$, $N = 5,000$, and $N = 10,000$ transactions.
