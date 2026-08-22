# Decision Engine & Thresholds Baseline Specification

**Document Version:** 1.0 (Phase -1 Baseline)  
**Evaluated Code:** `backend/internal/riskengine/orchestrator.go`, `ml-service/serve.py`, `backend/internal/rules/ast.go`  

---

## 1. Complete Decision Pipeline Flow

```
1. Incoming Request (POST /v1/risk-evaluations)
   │
   ├──► Step 1: Redis Velocity Ingestion & Retrieval
   │       • ZCOUNT 1h IP attempts (vel:ip:{tenant}:{ip})
   │       • ZCOUNT 24h Token attempts (vel:tok:{tenant}:{token})
   │
   ├──► Step 2: Pre-Rules Evaluation (Hard Guardrails)
   │       • Evaluate ACTIVE rules against in-memory evalContext.
   │       • IF matched rule action is "DECLINE_RECOMMENDATION":
   │            risk_score = 95, finalAction = DECLINE, HALT (Skip ML).
   │       • IF matched rule action is "ALLOW_RECOMMENDATION":
   │            risk_score = 5, finalAction = ALLOW, HALT (Skip ML).
   │       • IF matched rule action is "MANUAL_REVIEW":
   │            Append reason code, PROCEED to ML.
   │
   ├──► Step 3: ML Inference (50ms Deadline Budget)
   │       • Call Python FastAPI /predict with 5 float features.
   │       • ML returns raw probability p in [0.0, 1.0].
   │       • Risk score transformation: risk_score = int(round(p * 100)).
   │       • IF ML times out (>50ms) or fails:
   │            is_degraded = true, compute heuristic fallback score (15–99).
   │
   ├──► Step 4: Dynamic Threshold Action Mapping (if not halted by pre-rule)
   │       • risk_score >= 85 ──► DECLINE_RECOMMENDATION
   │       • risk_score >= 65 ──► MANUAL_REVIEW (Dispatches 24h SLA Case)
   │       • risk_score >= 45 ──► STEP_UP_RECOMMENDATION (3D-Secure / OTP)
   │       • risk_score <  45 ──► ALLOW_RECOMMENDATION
   │
   ├──► Step 5: AES-256-GCM Envelope Encryption & Atomic Outbox Write
   │       • Encrypt PII (IP, Device Fingerprint) in Postgres feature_snapshot.
   │       • Write to risk_decisions and outbox_events in single pgx.Tx.
   │
   └──► Step 6: Synchronous Response (<15ms)
```

---

## 2. Risk Score & Threshold Specifications

### A. ML Probability to Risk Score Transformation
- **Raw ML Model Output:** $p \in [0.0, 1.0]$ representing probability of fraudulent intent.
- **Conversion Formula:**
  $$\text{risk\_score} = \text{int}\Big(\text{round}\big(p \times 100\big)\Big)$$
- **Score Range:** Integer from `0` (cleanest) to `100` (highest risk).

### B. Action Threshold Boundaries

| Risk Score Range | Decision Action | System Action & Consequence |
| :--- | :--- | :--- |
| **$85 \le \text{Score} \le 100$** | `DECLINE_RECOMMENDATION` | Hard automated block. Transaction rejected; zero human review. |
| **$65 \le \text{Score} < 85$** | `MANUAL_REVIEW` | Queued in Case Management with a 24-hour analyst SLA countdown. |
| **$45 \le \text{Score} < 65$** | `STEP_UP_RECOMMENDATION` | Triggers 3D-Secure, biometric step-up, or SMS OTP challenge. |
| **$0 \le \text{Score} < 45$** | `ALLOW_RECOMMENDATION` | Frictionless instant checkout. |

---

## 3. Fallback & Degraded-Mode Heuristic

When the ML sidecar is unavailable, disconnected, or exceeds the **50ms context deadline**, the Go Orchestrator activates graceful degradation (`is_degraded: true`):

```go
func (o *Orchestrator) calculateFallbackRiskScore(amount int64, velocity *features.VelocityMetrics) int {
    score := 15 // baseline
    if amount > 100000 {
        score += 35 // large transaction penalty
    }
    if velocity.TxnCountIP1h >= 4 {
        score += 30 // IP burst penalty
    }
    if velocity.TxnCountToken24h >= 6 {
        score += 25 // Card velocity penalty
    }
    if score > 99 {
        score = 99
    }
    return score
}
```

---

## 4. Concrete End-to-End Decision Trace Example

### Input Payload
```json
{
  "amount": 95000,
  "currency": "INR",
  "payment_method": {
    "type": "card",
    "token": "tok_visa_suspect_99"
  },
  "ip_address": "198.51.100.42",
  "device_fingerprint": "new_device"
}
```

### Trace Progression:
1. **Feature Aggregation:**
   - `velocity.ip.1hr` = `5` (queried from Redis)
   - `velocity.token.24hr` = `7` (queried from Redis)
   - `is_new_device` = `1` (inferred from `"new_device"`)
   - `hour_of_day` = `14` (UTC)
2. **Pre-Rules Check:**
   - Active rule `High Velocity IP Burst Protection` matches (`velocity.ip.1hr > 5` evaluates to false for exact 5; rule 2 threshold `amount > 50000` is active).
3. **ML Inference:**
   - ONNX model receives vector `[95000.0, 5.0, 7.0, 1.0, 14.0]`.
   - Raw predicted probability: $p = 0.7842$.
   - Computed Risk Score: $\text{risk\_score} = 78$.
   - Local Attribution: `HIGH_IP_VELOCITY_1H`, `HIGH_TOKEN_VELOCITY_24H`, `HIGH_TRANSACTION_AMOUNT`, `NEW_DEVICE_FINGERPRINT`.
4. **Action Determination:**
   - Since $65 \le 78 < 85$, the action maps to **`MANUAL_REVIEW`**.
5. **Persistence & Outbox:**
   - PII is encrypted with tenant key via AES-256-GCM.
   - Record committed to `risk_decisions` and `outbox_events`.
   - Debezium / Kafka triggers `risk-case-manager-group` to instantiate a case.
6. **Execution Output:**
   - Total latency: **4.2 ms**.
   - Output: `MANUAL_REVIEW`, `risk_score: 78`.
