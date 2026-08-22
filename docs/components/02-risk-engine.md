# Component 02: Risk Evaluation Engine & Precedence Matrix

---

## 1. Why It Exists
In complex financial institutions, multiple decision engines run simultaneously:
1. Hard regulatory compliance blacklists (OFAC, PEP, sanctions).
2. Human-written declarative business rules.
3. Graph syndicate cluster detectors.
4. Statistical machine learning anomaly models.

When these systems produce conflicting recommendations (e.g., a rule recommends `APPROVE` while the ML model scores `0.85` indicating high risk), the system must have a **mathematically deterministic and regulatory-compliant arbitration mechanism**. 

The **Risk Evaluation Engine** (`backend/internal/riskengine/`) exists to resolve these conflicts via a formal **Decision Precedence Hierarchy**, enforce calibrated threshold bounds, run zero-risk shadow policies, and trigger closed-loop model retraining when concept drift occurs.

---

## 2. Architecture & Decision Precedence Hierarchy

```text
[ Multiple Engine Evaluations (Rules, ML, Graph, Threat Intel) ]
                               │
                               ▼
┌─────────────────────────────────────────────────────────────┐
│ LEVEL 1: Regulatory & Compliance Blacklists                 │
│ • OFAC, PEP, Sanctioned Countries                           │
│ • Deterministic Hard Override: Immediate BLOCK              │
└──────────────────────────────┬──────────────────────────────┘
                               │ (No Blacklist Match)
                               ▼
┌─────────────────────────────────────────────────────────────┐
│ LEVEL 2: Declarative Policy Rules (Manual Rules)            │
│ • Explicit Hard Velocity or Threshold Matches               │
│ • Override ML: BLOCK or CHALLENGE                           │
└──────────────────────────────┬──────────────────────────────┘
                               │ (No Rule Match)
                               ▼
┌─────────────────────────────────────────────────────────────┐
│ LEVEL 3: Fraud Knowledge Graph Syndicate Signals            │
│ • Cyclic multi-account mule cluster (degree >= 4)           │
│ • Escalates score directly to >= 0.80 (BLOCK)               │
└──────────────────────────────┬──────────────────────────────┘
                               │ (No Graph Cluster)
                               ▼
┌─────────────────────────────────────────────────────────────┐
│ LEVEL 4: Calibrated Machine Learning Score (XGBoost)        │
│ • Continuous probability (0.00 to 1.00)                     │
│ • Thresholds: 0.80 (BLOCK), 0.50 (CHALLENGE), 0.30 (REVIEW) │
└──────────────────────────────┬──────────────────────────────┘
                               │ (ML Score < 0.30)
                               ▼
┌─────────────────────────────────────────────────────────────┐
│ LEVEL 5: Baseline Organic Authorization (APPROVE)           │
│ • Frictionless immediate settlement                         │
└─────────────────────────────────────────────────────────────┘
```

---

## 3. Threshold Calibration & Policy Action Matrix

| Score Band | Final Decision | Recommended Action | Automated Subsystem Trigger |
| :---: | :---: | :---: | :--- |
| **0.00 – 0.29** | `APPROVE` | `ALLOW` | Immediate settlement authorization; asynchronous telemetry logging. |
| **0.30 – 0.49** | `REVIEW` | `MANUAL_REVIEW` | Transaction allowed with step-down delay; enqueued in Analyst Case Queue. |
| **0.50 – 0.79** | `CHALLENGE` | `STEP_UP_MFA` | Intercept payment; trigger adaptive WebAuthn or SMS/Push biometrics. |
| **0.80 – 1.00** | `BLOCK` | `BLOCK_AND_REVIEW` | Immediate payment halt; session quarantine; Priority P0 Case opened. |

---

## 4. Key Data Structures (Go)

```go
type RiskEvaluationResult struct {
    EvaluationID string            `json:"evaluation_id"`
    Score        float64           `json:"score"`
    Action       PolicyAction      `json:"action"` // "APPROVE", "REVIEW", "CHALLENGE", "BLOCK"
    Confidence   float64           `json:"confidence"`
    MatchedRules []RuleMatchRecord `json:"matched_rules"`
    MLScore      float64           `json:"ml_score"`
    GraphScore   float64           `json:"graph_score"`
    EvaluatedAt  time.Time         `json:"evaluated_at"`
}

type RetrainingTriggerConfig struct {
    KSDriftThreshold float64       `json:"ks_drift_threshold"` // e.g. 0.05
    PSIThreshold     float64       `json:"psi_threshold"`      // e.g. 0.10
    MinDisputeSample int           `json:"min_dispute_sample"` // e.g. 500 records
    CooldownPeriod   time.Duration `json:"cooldown_period"`    // e.g. 6 hours
}
```

---

## 5. Drift Monitoring & Closed-Loop Retraining

The engine continuously computes the **Population Stability Index (PSI)** and **Kolmogorov-Smirnov (KS)** statistic over sliding 24-hour evaluation windows:

$$\text{PSI} = \sum_{b=1}^{B} \left( P_b - Q_b \right) \times \ln\left( \frac{P_b}{Q_b} \right)$$
where $P_b$ is the current evaluation distribution and $Q_b$ is the baseline training distribution.

- If $\text{PSI} > 0.10$, a warning alert is emitted to the Observability plane.
- If $\text{PSI} > 0.25$, the engine triggers [`retraining_trigger.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/riskengine/retraining_trigger.go) to initiate an asynchronous offline retraining pipeline.

---

## 6. Failure Modes & Edge Cases
- **Conflicting Rules**: When multiple rules match with different actions (e.g. Rule A says `REVIEW`, Rule B says `BLOCK`), the engine deterministically applies the **most restrictive action** (`BLOCK > CHALLENGE > REVIEW > APPROVE`).
- **Cold Start (New Tenant/Model)**: If no calibrated ML model is assigned to a tenant, the engine operates in pure **Rule + Graph Mode**, logging evaluation data to build historical training samples.

---

## 7. Performance & Latency
- **Arbitration Execution Time**: $< 0.10\text{ms}$ in-memory.
- **Shadow Mode Overhead**: Shadow evaluations execute asynchronously in background goroutines with zero impact on the primary response latency.

---

## 8. Source Code Map
- [`backend/internal/riskengine/engine.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/riskengine/engine.go): Core risk engine logic and decision precedence arbitration.
- [`backend/internal/riskengine/retraining_trigger.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/riskengine/retraining_trigger.go): Drift threshold monitoring and retraining dispatch.
- [`backend/internal/riskengine/shadow_evaluator.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/riskengine/shadow_evaluator.go): Zero-risk shadow policy simulation.

---

## 9. Cross-Component Links
- [Component 01: Product API](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/01-product-api.md) — Upstream caller.
- [Component 03: Rules Engine](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/03-rules-engine.md) — Declarative rule evaluation.
- [Component 04: ML Inference](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/04-ml-inference.md) — Machine learning scoring.
