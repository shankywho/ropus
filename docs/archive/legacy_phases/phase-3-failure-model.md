# Phase 3 — Component Failure & Graceful Degradation Model

**Document Version:** 1.0  
**High Availability Standard:** 99.99% Decision Uptime Under Infrastructure Outages  

---

## 1. Component Failure Matrix & Deterministic Degradation

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                    INFRASTRUCTURE COMPONENT FAILURE MATRIX                      │
├────────────────────┬──────────────────────────────────┬─────────────────────────┤
│ Component Outage   │ Immediate System Behavior        │ Risk Decision Outcome   │
├────────────────────┼──────────────────────────────────┼─────────────────────────┤
│ 1. Redis Cluster   │ • Catch error in <0.5ms.         │ • Hard pre-rules STILL  │
│    (Down/Timeout)  │ • Set default clean velocities:  │   EXECUTE!              │
│                    │   ip_vel=0, tok_vel=0, dev_vel=0 │ • ML evaluates fallback │
│                    │ • Set device_seen_before = 0.    │   features.             │
│                    │ • Log critical alert; proceed.   │ • is_degraded = true.   │
├────────────────────┼──────────────────────────────────┼─────────────────────────┤
│ 2. Python ML Side- │ • Go context times out at 50ms.  │ • Fallback heuristic    │
│    car (Down/500)  │ • Calculate fallback risk score: │   score (15–95).        │
│                    │   15 + (35 if amt > 100k)        │ • Pre-rules STILL halt. │
│                    │   + (30 if ip_vel >= 4) ...      │ • is_degraded = true.   │
│                    │ • Appends "ML_SERVICE_DEGRADED". │ • Pushes to REVIEW if   │
│                    │                                  │   velocity was high.    │
├────────────────────┼──────────────────────────────────┼─────────────────────────┤
│ 3. Client Finger-  │ • Client passes fallback hash or │ • Evaluated with        │
│    printJS Sensor  │   empty device_fingerprint.      │   device_info_missing=1.│
│    (Blocked/Slow)  │ • Set device_info_missing = 1.   │ • Relies on IP & Card   │
│                    │ • Proceed with normal pipeline.  │   token velocities.     │
├────────────────────┼──────────────────────────────────┼─────────────────────────┤
│ 4. PostgreSQL DB   │ • In-memory decision completes   │ • Payment authorization │
│    (Pool Exhausted)│   and returns synchronously.     │   decision SUCCEEDS.    │
│                    │ • Asynchronous write retried via │ • Buffer in memory /    │
│                    │   local queue or dead-letter log.│   Redis recovery queue. │
├────────────────────┼──────────────────────────────────┼─────────────────────────┤
│ 5. Redpanda/Kafka  │ • PostgreSQL outbox write        │ • Zero impact on real-  │
│    or Debezium CDC │   succeeds transactionally.      │   time payment auth.    │
│    (Lagging/Down)  │ • Events accumulate in outbox_   │ • Debezium resumes CDC  │
│                    │   events table until recovered.  │   drain upon recovery.  │
└────────────────────┴──────────────────────────────────┴─────────────────────────┘
```

---

## 2. Unbreakable Invariants: What MUST NEVER Fail

1. **Hard Pre-Rules Never Bypassed:**
   - Even if Redis, Python ML, PostgreSQL, and Kafka are all simultaneously unavailable, the Go in-memory declarative rule engine (`rules/ast.go`) continues evaluating hard blacklists, country bans, and merchant whitelist rules from in-memory cache.
2. **Deterministic Fallback Scores:**
   - Fallback heuristic scores strictly map to valid integers $[15, 99]$ and never return `NaN` or unhandled exceptions.
3. **Structured Degradation Reason Codes:**
   - When degraded paths execute, the response MUST emit explicit system signals:
     - `"SYSTEM_DEGRADED:REDIS_FEATURE_STORE_UNAVAILABLE"`
     - `"SYSTEM_DEGRADED:ML_SERVICE_TIMEOUT_FALLBACK"`
     - `"SYSTEM_DEGRADED:DEVICE_TELEMETRY_MISSING"`
