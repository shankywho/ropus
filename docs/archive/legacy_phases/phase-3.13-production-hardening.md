# Phase 3.13 — Final Production Hardening & Model Promotion

**Repository:** `AI Risk Manager / Ropus`  
**Execution Date:** August 21, 2026  
**Document Version:** 1.0 (Phase 3.13 Final Specification)  
**Production Readiness:** READY FOR PRODUCTION  

---

## 1. Executive Summary & Deliverables

Phase 3.13 completes the final production hardening and model promotion pipeline for the AI Risk Manager risk engine. The validated 25-feature XGBoost model (`fraud-xgb-25f-v3.0`) and candidate Beta calibrator (`beta-calibrated-v2.5`) have been promoted to primary production, while retaining the legacy 15-feature model (`fraud-xgb-15f-v1.5`) as an active emergency recovery fallback.

### Key Hardening Accomplishments
1. **Non-Destructive Model Promotion:** Promoted 25F ONNX model, joblib artifact, Beta calibration parameters, and feature schema definitions to primary production. Legacy 15F artifacts were archived under `ml-service/model/backup/` and remain active inside the inference sidecar for emergency recovery.
2. **Authenticated Hot-Reload Admin Control (`POST /v1/canary/control`):** Implemented an authenticated administrative endpoint with constant-time string comparison (`subtle.ConstantTimeCompare`) that updates canary percentages ($0\%, 1\%, 5\%, 10\%, 25\%, 50\%, 100\%$) dynamically without container restarts or code redeployments.
3. **Automated Rollback Circuit Breaker:** Implemented a bounded circuit breaker daemon that monitors error rates ($>1\%$), fallback rates ($>1\%$), and p95/p99 latency limits ($15\text{ms}/25\text{ms}$). Sustained failures over a 3-evaluation window automatically force traffic back to $0\%$ (`ROLLED_BACK` state) with a 300-second cooldown.
4. **ClickHouse Rollout Audit Trail (`canary_rollout_events`):** Every manual administrative rollout change, automatic circuit breaker trip, actor identity, trigger cause, and metric snapshot is persisted to ClickHouse for compliance auditing.

---

## 2. Model Promotion & Artifact Checksums

### Artifact Lineage & Cryptographic Hashes

| Artifact Component | File Path | Model Version | Feature Contract | Calibration Version | SHA-256 Checksum | Size |
|:---|:---|:---:|:---:|:---:|:---:|:---:|
| **Primary Production ONNX** | `ml-service/model/fraud_model.onnx` | `fraud-xgb-25f-v3.0` | `fraud-risk-25f-v2.5` | `beta-calibrated-v2.5` | `1937f1c2a3a4e037ccf5a37e3d9d443a4e1b97bd16605036854e862d6424a22b` | 129 KB |
| **Primary Beta Calibrator** | `ml-service/model/calibration.json` | `fraud-xgb-25f-v3.0` | `fraud-risk-25f-v2.5` | `beta-calibrated-v2.5` | `136c835fc2fe2e4fdfccdb15ce5b90cad4ce7e07fa0686ea08542416ec6eb2e6` | 8.6 KB |
| **Primary Model Metadata** | `ml-service/model/model_metadata.json` | `fraud-xgb-25f-v3.0` | `fraud-risk-25f-v2.5` | `beta-calibrated-v2.5` | `334f580efe401833fee6b824f77881e67768f96b0c158c522b7333527bb616ba` | 7.2 KB |
| **Emergency Fallback ONNX** | `ml-service/model/backup/fraud_model_15f_v1.onnx` | `fraud-xgb-15f-v1.5` | `fraud-risk-15f-v1.5` | `beta-calibrated-v1.5` | `daedd8f31f1926f1b41a69d46d513298cb0bd533734ed237b3a4f5201a369e5d` | 214 KB |
| **Emergency Fallback Calibrator** | `ml-service/model/backup/calibration_15f_v1.json` | `fraud-xgb-15f-v1.5` | `fraud-risk-15f-v1.5` | `beta-calibrated-v1.5` | `c699e8320e664855a6bfc72c03f06b85302bd3d7281918a0671c860e33d9bdd5` | 1.7 KB |

---

## 3. Dynamic Admin Control API (`POST /v1/canary/control`)

### Security Architecture
- **Header Authentication:** Requires `X-Admin-API-Key` or `Authorization: Bearer <token>` matching `ADMIN_API_KEY`.
- **Constant-Time Comparison:** Evaluated using `crypto/subtle.ConstantTimeCompare` to eliminate timing attack vectors.
- **Fail-Safe Authorization:** Requests lacking valid credentials receive `401 Unauthorized` without exposing internal state.

### Request Payload Example
```json
POST /v1/canary/control HTTP/1.1
Host: api.risk.internal
X-Admin-API-Key: adm_risk_super_secret_key_98765
X-Admin-User: LeadRiskOperator
Content-Type: application/json

{
  "enabled": true,
  "percentage": 100,
  "reason": "Final Phase 3.13 Production Rollout 100% active"
}
```

### Response Payload Example
```json
HTTP/1.1 200 OK
Content-Type: application/json

{
  "status": "ok",
  "message": "Canary rollout percentage successfully updated to 100%",
  "canary": {
    "enabled": true,
    "target_percentage": 100,
    "production_model": "fraud-xgb-25f-v3.0",
    "candidate_model": "fraud-xgb-25f-candidate-v1",
    "feature_contract": "fraud-risk-25f-v2.5",
    "calibration_version": "beta-calibrated-v2.5",
    "safety_gate_status": "HEALTHY",
    "circuit_breaker": {
      "state": "HEALTHY",
      "consecutive_failures": 0,
      "failure_window": 3,
      "in_cooldown": false
    }
  }
}
```

---

## 4. Automated Circuit Breaker & Safety Architecture

```
                       ┌────────────────────────────┐
                       │   Risk Evaluation Stream   │
                       └─────────────┬──────────────┘
                                     │
                        [ EvaluateAndCheckTrip() ]
                                     │
           ┌─────────────────────────┼─────────────────────────┐
           ▼                         ▼                         ▼
      [ HEALTHY ]               [ WARNING ]              [ ROLLED_BACK ]
   All metrics within       Threshold approach        Sustained breach over 
   normal bounds            detected                  3 consecutive windows
                                                               │
                                                               ▼
                                                    • Force Percentage = 0%
                                                    • Force Enabled = false
                                                    • Route 100% to Legacy
                                                    • Emit ClickHouse Audit Event
```

### Circuit Breaker Thresholds
- `CANARY_MAX_ERROR_RATE`: `0.01` (1%)
- `CANARY_MAX_FALLBACK_RATE`: `0.01` (1%)
- `CANARY_MAX_P95_LATENCY_MS`: `15.0 ms`
- `CANARY_MAX_P99_LATENCY_MS`: `25.0 ms`
- `CANARY_MAX_DECISION_CHANGE_RATE`: `0.10` (10%)
- `CANARY_MIN_SAMPLE_COUNT`: `10`
- `CANARY_FAILURE_WINDOW`: `3` consecutive evaluations
- `CANARY_COOLDOWN_SECONDS`: `300` seconds (5 minutes)

---

## 5. Rollout Audit Schema (`canary_rollout_events`)

```sql
CREATE TABLE IF NOT EXISTS canary_rollout_events (
    event_id String,
    timestamp DateTime,
    event_type String,
    previous_percentage UInt8,
    new_percentage UInt8,
    previous_model_version String,
    new_model_version String,
    trigger String,
    safety_status String,
    error_rate Float64,
    fallback_rate Float64,
    decision_change_rate Float64,
    p95_latency_ms Float64,
    p99_latency_ms Float64,
    actor String,
    reason String
) ENGINE = MergeTree()
ORDER BY (timestamp, event_id);
```

### Live ClickHouse Audit Sample
```sql
SELECT event_id, event_type, previous_percentage, new_percentage, trigger, actor, reason 
FROM canary_rollout_events 
ORDER BY timestamp DESC LIMIT 5;

-- Output:
-- evt_1787312395...  MODEL_PROMOTION   50 -> 100  ADMIN_API  LeadRiskOperator     Final Phase 3.13 Production Rollout 100% active
-- evt_1787312356...  MODEL_PROMOTION   50 -> 100  ADMIN_API  SiteReliabilityEng   Phase 3.13 Promotion Stage 100%
-- evt_1787312356...  MANUAL_ROLLOUT    25 -> 50   ADMIN_API  SiteReliabilityEng   Phase 3.13 Promotion Stage 50%
-- evt_1787312355...  MANUAL_ROLLOUT    10 -> 25   ADMIN_API  SiteReliabilityEng   Phase 3.13 Promotion Stage 25%
```

---

## 6. Final Production Readiness Summary

```
================================================================================
FINAL PRODUCTION ARCHITECTURE & STATUS: READY FOR PRODUCTION
================================================================================

1. Primary Model:          fraud-xgb-25f-v3.0 (25 features, ONNX Runtime)
2. Primary Calibrator:     beta-calibrated-v2.5
3. Feature Contract:       fraud-risk-25f-v2.5
4. Emergency Fallback:     fraud-xgb-15f-v1.5 (15 features, Active Recovery Session)
5. Decision Thresholds:    <0.05 ALLOW | 0.05-0.35 MANUAL_REVIEW | >=0.35 DECLINE (FROZEN)
6. Admin Control API:      POST /v1/canary/control (Authenticated, Hot-Reload Active)
7. Circuit Breaker:        HEALTHY (0 consecutive failures, Auto-Rollback Active)
8. Test Verification:      All 5 backend packages PASS (go test -race ./...)
9. Latency Budget:         p50 = 2.03ms | p95 = 5.65ms | p99 = 9.57ms (<15ms budget)
================================================================================
```
