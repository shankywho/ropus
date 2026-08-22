# Phase 3 — Comprehensive Verification & Test Strategy

**Document Version:** 1.0  
**Quality & Reliability Target:** 100% Automated Coverage of All 42 Edge Cases  

---

## 1. Test Layer Breakdown

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                          PHASE 3 TEST STRATEGY SUITE                            │
├────────────────────┬──────────────────────────────────┬─────────────────────────┤
│ Test Suite         │ Scope & Methodology              │ Target File Locations   │
├────────────────────┼──────────────────────────────────┼─────────────────────────┤
│ 1. Unit Tests      │ • Device hashing & sanitization  │ backend/internal/       │
│    (Go & Python)   │ • Beta calibration math bounds   │   features/device_test.go│
│                    │ • Cost policy calculation logic  │ ml-service/tests/       │
│                    │ • Rule DSL AST evaluation        │   test_phase3_device.py │
├────────────────────┼──────────────────────────────────┼─────────────────────────┤
│ 2. Redis Pipeline  │ • ZCOUNT sliding-window windows  │ backend/internal/       │
│    Store Tests     │ • SCARD multi-account sets       │   features/velocity_test│
│                    │ • TTL eviction & key memory      │   .go                   │
├────────────────────┼──────────────────────────────────┼─────────────────────────┤
│ 3. Database Schema │ • Migration up/down validity     │ backend/migrations/     │
│    & Foreign Keys  │ • Unique constraint enforcement  │ tests/integration/      │
│                    │ • Multi-tenant query isolation   │   postgres_test.go      │
├────────────────────┼──────────────────────────────────┼─────────────────────────┤
│ 4. Concurrency &   │ • 100 concurrent requests from   │ tests/concurrency/      │
│    Race Tests      │   same novel device in <10ms     │   device_stampede_test  │
│                    │ • Atomic SETNX registration      │   .go                   │
├────────────────────┼──────────────────────────────────┼─────────────────────────┤
│ 5. Temporal Point- │ • Sequential events T1 < T2 < T3 │ ml-service/tests/       │
│    in-Time Leakage │ • Verify future events do not    │   test_temporal_leakage │
│                    │   alter feature values at T1     │   .py                   │
├────────────────────┼──────────────────────────────────┼─────────────────────────┤
│ 6. Fault Injection │ • Redis connection refusal       │ tests/resilience/       │
│    & Degradation   │ • ML sidecar 100ms lag / 500     │   fault_injection_test  │
│                    │ • Verify is_degraded=true & score│   .go                   │
├────────────────────┼──────────────────────────────────┼─────────────────────────┤
│ 7. Performance &   │ • 500 concurrent RPS benchmark   │ scripts/                │
│    Latency Load    │ • Measure p50, p95, p99 latency  │   benchmark_device_load │
│                    │ • Verify target < 15ms p95       │   .py                   │
└────────────────────┴──────────────────────────────────┴─────────────────────────┘
```

---

## 2. Temporal Point-in-Time Leakage Protection Protocol

Every device feature must undergo rigorous automated leakage testing:

```python
def test_device_feature_temporal_leakage_isolation():
    # 1. Simulate Device D1 first transaction at T1 = 1000 with Account A1
    #    Expected: device_seen_before = 0, device_unique_accounts_24h = 1
    feat_t1 = compute_device_features_at_time(device_id="D1", account_id="A1", current_time=1000)
    assert feat_t1["device_seen_before"] == 0
    assert feat_t1["device_unique_accounts_24h"] == 1

    # 2. Simulate subsequent transactions at T2 = 2000 (Account A2) and T3 = 3000 (Account A3)
    record_event(device_id="D1", account_id="A2", timestamp=2000)
    record_event(device_id="D1", account_id="A3", timestamp=3000)

    # 3. RE-EVALUATE historical state at T1 = 1000
    #    CRITICAL INVARIANT: Future events T2 and T3 MUST NOT alter historical features at T1!
    feat_t1_replay = compute_device_features_at_time(device_id="D1", account_id="A1", current_time=1000)
    assert feat_t1_replay["device_seen_before"] == 0
    assert feat_t1_replay["device_unique_accounts_24h"] == 1
    assert feat_t1_replay == feat_t1, "TEMPORAL FUTURE LEAKAGE DETECTED!"
```
