# Risk Decision Precedence & Rule-ML Interaction

**Document Version:** 1.0 (Phase 2 Specification)  
**Evaluated Code:** `backend/internal/riskengine/orchestrator.go`, `backend/internal/rules/ast.go`  

---

## 1. Authoritative Decision Precedence Hierarchy

The AI Risk Manager enforces a strict, multi-tiered precedence hierarchy ensuring legal compliance and hard deterministic guardrails supersede probabilistic machine learning signals:

```
+-------------------------------------------------------------------------------+
| TIER 1: HARD BLOCK PRE-RULES (Sanctions, Blacklists, Stolen Cards)            |
| -> Matched rule with action="DECLINE_RECOMMENDATION"                          |
| -> Sets risk_score=95, finalAction="DECLINE_RECOMMENDATION", HALTS PIPELINE. |
+-------------------------------------------------------------------------------+
                                      │ (No match)
                                      ▼
+-------------------------------------------------------------------------------+
| TIER 2: HARD ALLOW PRE-RULES (Whitelists, Internal System Test Tokens)        |
| -> Matched rule with action="ALLOW_RECOMMENDATION"                            |
| -> Sets risk_score=5, finalAction="ALLOW_RECOMMENDATION", HALTS PIPELINE.    |
+-------------------------------------------------------------------------------+
                                      │ (No match)
                                      ▼
+-------------------------------------------------------------------------------+
| TIER 3: CALIBRATED ML + COST-SENSITIVE POLICY OPTIMIZATION                     |
| -> Passes features to FastAPI ONNX Inference (50ms context deadline).         |
| -> Computes calibrated empirical posterior P(fraud | x).                      |
| -> Evaluates expected loss: E[Cost(ALLOW)] vs E[Cost(REVIEW)] vs E[Cost(DEC)].|
| -> Selects optimal action subject to operational review capacity.             |
+-------------------------------------------------------------------------------+
                                      │ (ML timeout / network failure)
                                      ▼
+-------------------------------------------------------------------------------+
| TIER 4: DEGRADED MODE HEURISTIC FALLBACK                                      |
| -> Activates deterministic fallback risk score (15-99).                       |
| -> Appends reason code "ML_SERVICE_DEGRADED", sets is_degraded=true.          |
+-------------------------------------------------------------------------------+
```

---

## 2. Conflict Matrix & Deterministic Resolutions

| Scenario # | Condition Description | Pre-Rule Action | ML / Cost Signal | Degraded? | Final Resolution | Rationale |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **1** | Blacklisted IP / Sanctioned Country | `DECLINE` | `ALLOW` ($p=0.01$) | No | **`DECLINE_RECOMMENDATION`** | Hard pre-rule halts pipeline in Tier 1 before ML inference occurs. |
| **2** | VIP Whitelisted Corporate Card | `ALLOW` | `DECLINE` ($p=0.92$) | No | **`ALLOW_RECOMMENDATION`** | Tier 2 Whitelist explicitly bypasses probabilistic scoring. |
| **3** | Borderline Velocity Rule Match | `MANUAL_REVIEW` | `DECLINE` ($p=0.88$) | No | **`DECLINE_RECOMMENDATION`** | Soft rule appends reason code; Tier 3 ML/Cost policy determines final action. |
| **4** | ML Sidecar Unreachable (>50ms) + Blacklisted Card | `DECLINE` | Unreachable | Yes | **`DECLINE_RECOMMENDATION`** | Pre-rule halts execution; system never depends on ML for hard blocks. |
| **5** | ML Sidecar Unreachable (>50ms) + High Velocity Amount | None | Unreachable | Yes | **`MANUAL_REVIEW`** or **`DECLINE`** | Heuristic fallback assigns high risk score (85–99) based on Redis velocity counters. |
