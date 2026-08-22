# Phase 2 Current Decision Audit & Score Semantics

**Document Version:** 1.0 (Phase 2 Baseline Audit)  
**Evaluated Systems:** `ml-service/serve.py`, `backend/internal/riskengine/orchestrator.go`, `backend/internal/riskengine/mlclient.go`  

---

## 1. End-to-End Decision & Score Trace

```
1. RAW MODEL INFERENCE (ml-service/serve.py)
   │
   ├──► ONNX Runtime C++ Engine Output:
   │       raw_preds[1] = [[p_legit, p_fraud]] where p_fraud in [0.0, 1.0].
   │       Current Reality: p_fraud is raw uncalibrated probability.
   │
   ├──► Transformation in Python:
   │       risk_score = int(round(p_fraud * 100))  (Linear integer scaling to 0..100)
   │
   └──► JSON HTTP Response to Go Backend:
           {
             "risk_score": 78,
             "probability": 0.7842,
             "reason_codes": ["HIGH_IP_VELOCITY_1H"],
             "feature_attributions": {...},
             "latency_ms": 1.45,
             "runtime": "onnxruntime"
           }

2. GO RISK ORCHESTRATOR PROCESSING (backend/internal/riskengine/orchestrator.go)
   │
   ├──► Step 1: Redis Velocity Ingestion & Queries (1h IP, 24h Token).
   │
   ├──► Step 2: Pre-Rules Evaluation (Hard Guardrails):
   │       • Matches active rules against in-memory context before ML.
   │       • IF Rule Action is "DECLINE_RECOMMENDATION":
   │            Set risk_score = 95, finalAction = DECLINE, HALT (Bypasses ML).
   │       • IF Rule Action is "ALLOW_RECOMMENDATION":
   │            Set risk_score = 5, finalAction = ALLOW, HALT (Bypasses ML).
   │       • IF Rule Action is "MANUAL_REVIEW":
   │            Appends reason code; PROCEEDS to ML inference.
   │
   ├──► Step 3: ML Inference (50ms Deadline Budget):
   │       • Calls Python FastAPI /predict.
   │       • If ML succeeds: Adopts returned mlResp.RiskScore.
   │       • If ML times out or fails (Degraded Mode):
   │            is_degraded = true; computes heuristic score:
   │            score = 15 + (35 if amount > 100k) + (30 if ip_vel >= 4) + (25 if tok_vel >= 6).
   │            Appends reason code "ML_SERVICE_DEGRADED".
   │
   └──► Step 4: Fixed Static Threshold Mapping (if not halted by pre-rule):
           • risk_score >= 85 ──► DECLINE_RECOMMENDATION
           • risk_score >= 65 ──► MANUAL_REVIEW (Queued in Cases with 24h SLA)
           • risk_score >= 45 ──► STEP_UP_RECOMMENDATION (3DS / OTP)
           • risk_score <  45 ──► ALLOW_RECOMMENDATION
```

---

## 2. Score Semantics Audit & Key Questions

| Question | Current Implementation Reality | Risk / Critique |
| :--- | :--- | :--- |
| **Is the ONNX output a true probability?** | **NO.** It is raw uncalibrated logistic output from XGBoost. In imbalanced datasets with `scale_pos_weight`, raw output is distorted and does not reflect true empirical posterior $P(\text{fraud} \mid x)$. | Probabilities cannot be directly multiplied with financial amounts for expected loss calculations. |
| **How is risk score derived?** | Pure linear multiplication: $\text{round}(p \times 100)$. | Arbitrary assumption that $p=0.85$ corresponds to hard decline threshold regardless of transaction size. |
| **Are thresholds applied before or after rules?** | **Both.** Pre-rules evaluate *before* ML (hard guardrails). Threshold mapping evaluates *after* ML as default action. | Sound architectural precedence: deterministic legal/compliance guardrails supersede probabilistic models. |
| **Can rules override ML?** | **YES.** Pre-rules with `DECLINE` or `ALLOW` immediately halt pipeline before calling ML. | Essential for hard blacklists, VIP whitelists, and sanction screens. |
| **Can ML override rules?** | **NO.** ML output only determines action when no hard pre-rule has halted execution. | Complies with enterprise risk governance. |
| **What happens in degraded mode?** | Go orchestrator catches ML timeout/error in $<0.1\text{ms}$, sets `is_degraded: true`, applies deterministic fallback heuristic (15–99 score), and logs warning. | Zero downtime; preserves system availability under ML sidecar failure. |

---

## 3. Transition Strategy for Phase 2

1. **Implement In-Process Calibration** (`ml-service/calibration/calibrator.py`):
   Fit Platt and Isotonic calibration models on the validation split. Transform raw $p_{\text{raw}}$ into empirical posterior $P(\text{fraud} \mid x)$.
2. **Expose Dual Probabilities in API Contract**:
   Expose `raw_probability` and `calibrated_probability` in `PredictResponse` while preserving `risk_score` for backwards compatibility.
3. **Cost-Sensitive Decision Policy Engine** (`config/risk-policy.json`):
   Calculate $E[\text{cost}]$ for ALLOW vs REVIEW vs DECLINE using calibrated probabilities, transaction amount, and operational review capacity.
