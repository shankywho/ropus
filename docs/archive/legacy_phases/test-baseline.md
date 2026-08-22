# Test Baseline & Edge-Case Gap Matrix

**Document Version:** 1.0 (Phase -1 Baseline)  
**Test Command:** `cd backend && go test -v ./...`  
**Execution Timestamp:** August 21, 2026  

---

## 1. Automated Test Suite Execution Summary

```
?   	github.com/shankywho/ropus/backend/cmd/api	[no test files]
?   	github.com/shankywho/ropus/backend/internal/audit	[no test files]
?   	github.com/shankywho/ropus/backend/internal/cases	[no test files]
?   	github.com/shankywho/ropus/backend/internal/features	[no test files]
=== RUN   TestWebhookHandler_HMACVerification
--- PASS: TestWebhookHandler_HMACVerification (0.00s)
PASS
ok  	github.com/shankywho/ropus/backend/internal/ingestion	1.140s
=== RUN   TestEvaluateRisk_OrchestratorPipeline
--- PASS: TestEvaluateRisk_OrchestratorPipeline (0.00s)
=== RUN   TestEvaluateRisk_MLTimeoutFallback
2026/08/21 14:39:19 ML inference degraded. Falling back to rules/heuristics.
--- PASS: TestEvaluateRisk_MLTimeoutFallback (0.00s)
PASS
ok  	github.com/shankywho/ropus/backend/internal/riskengine	1.446s
=== RUN   TestAST_FieldOperators
--- PASS: TestAST_FieldOperators (0.00s)
    --- PASS: Numeric_Greater_Than_(Int)_-_True (0.00s)
    --- PASS: Numeric_Greater_Than_-_False (0.00s)
    --- PASS: Numeric_Greater_Than_or_Equal_(GTE)_-_True (0.00s)
    --- PASS: Numeric_Less_Than_(LT)_-_True (0.00s)
    --- PASS: Numeric_Less_Than_or_Equal_(LTE)_-_True (0.00s)
    --- PASS: String_Equality_-_True (0.00s)
    --- PASS: String_Inequality_-_True (0.00s)
    --- PASS: IN_Operator_(Slice)_-_True (0.00s)
    --- PASS: IN_Operator_(Slice)_-_False (0.00s)
    --- PASS: NOT_IN_Operator_-_True (0.00s)
    --- PASS: CONTAINS_Operator_(Substring)_-_True (0.00s)
    --- PASS: STARTS_WITH_Operator_-_True (0.00s)
    --- PASS: ENDS_WITH_Operator_-_True (0.00s)
=== RUN   TestAST_NestedLogicalCombinators
--- PASS: TestAST_NestedLogicalCombinators (0.00s)
=== RUN   TestAST_ParseRuleDefinition
--- PASS: TestAST_ParseRuleDefinition (0.00s)
PASS
ok  	github.com/shankywho/ropus/backend/internal/rules	2.000s
=== RUN   TestEncryptDecrypt
--- PASS: TestEncryptDecrypt (0.00s)
=== RUN   TestMockKMS_CryptoShredding
--- PASS: TestMockKMS_CryptoShredding (0.00s)
PASS
ok  	github.com/shankywho/ropus/backend/internal/utils	2.523s
```

### Overall Results:
- **Total Test Suites Executed:** 4 packages tested, 4 packages skipped (`[no test files]`).
- **Total Tests Passed:** 19 individual unit test assertions passed (100% pass rate).
- **Total Tests Failed:** 0.

---

## 2. Package Coverage & Integration Test Audit

| Package / Domain | Unit Tests Present? | Integration Tests Present? | Untested Critical Paths |
| :--- | :--- | :--- | :--- |
| `cmd/api` | ❌ None (`0%`) | ❌ None | Main server startup, router wiring, CORS, middleware timeouts, shutdown signal handling. |
| `internal/audit` | ❌ None (`0%`) | ❌ None | Kafka audit consumer offset handling, ClickHouse batch inserts, ClickHouse reconnect loop. |
| `internal/cases` | ❌ None (`0%`) | ❌ None | Case creation SQL queries, SLA expiration calculations, claim concurrency, status transitions. |
| `internal/features` | ❌ None (`0%`) | ❌ None | Redis ZSET expiration cleanup, sorted set pipeline execution, network partition handling. |
| `internal/ingestion` | ✅ Unit test (`HMAC`) | ❌ None | Replay protection, database dispute insertion, evidence packet correlation failure. |
| `internal/riskengine` | ✅ Unit test (`Pipeline`, `Fallback`) | ❌ None | Outbox transaction rollback, Postgres lock contention, AES encryption failure path. |
| `internal/rules` | ✅ Unit test (`AST Operators`) | ❌ None | Database rule status transitions, Maker-Checker self-approval prohibition SQL level. |
| `internal/utils` | ✅ Unit test (`AES`, `KMS`) | ❌ None | Key rotation, persistent KMS integration, corrupted ciphertext handling. |

---

## 3. Edge-Case Gap Matrix (30 Critical Scenarios)

| # | Scenario | Current Behavior | Expected Behavior | Tested? | Severity |
| :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | **Normal transaction** | Evaluates via Redis $\rightarrow$ Rules $\rightarrow$ ML $\rightarrow$ Postgres $\rightarrow$ Outbox. Returns `<45` ALLOW. | Clean evaluation under 20ms. | ✅ Yes | Low |
| 2 | **High-value legitimate transaction** | Evaluated by ML; high amount alone may trigger `HIGH_TRANSACTION_AMOUNT` flag. | Risk score accurately calibrated against user history without false positive block. | ❌ No | High |
| 3 | **New customer (no historical data)** | Handled with `velocity = 0`; ML evaluates based on baseline distributions. | Smooth onboarding with calibrated first-txn risk threshold. | ❌ No | Medium |
| 4 | **Missing optional fields** | Default values populated (`Currency="INR"`, `Token="tok_default_123"`, `IsNewDevice=0`). | Graceful schema validation with explicit error or default fallback. | ❌ No | Low |
| 5 | **Duplicate transaction ID** | Postgres `ON CONFLICT (tenant_id, transaction_id)` executes `DO UPDATE`. | Idempotent response with original decision or explicit 409 conflict. | ❌ No | High |
| 6 | **Repeated rapid transactions** | Redis Sorted Set counts increase; velocity flags trigger at threshold. | Velocity window catches rapid repeat requests within 1 second. | ❌ No | High |
| 7 | **Extreme transaction amount ($10^9$)** | Handled as `int64`; passed to ML as float64 without bounds checking. | Amount validated against currency ceilings; triggers immediate review if above maximum allowed. | ❌ No | High |
| 8 | **Zero or negative amount** | Accepted by Go struct (unless negative triggers uint/int bounds). Passed to ML. | HTTP `400 Bad Request` (`amount must be > 0`). | ❌ No | Critical |
| 9 | **Malformed request JSON** | HTTP 400 returned by JSON decoder in `handler.go`. | HTTP `400 Bad Request` with structured error JSON. | ✅ Yes | Medium |
| 10 | **Unknown tenant ID** | Auto-creates tenant in database via `ensureTenantExists()`. | Returns `401 Unauthorized` or `404 Not Found` if tenant does not exist. | ❌ No | Critical |
| 11 | **Invalid API key / Unauthenticated** | API currently does not validate API key headers (stubbed default tenant). | Strict Bearer/API Key authentication check rejecting unknown callers (`401`). | ❌ No | Critical |
| 12 | **Redis unavailable / disconnected** | Logged; defaults velocity to `{TxnCountIP1h: 0, TxnCountToken24h: 0}`; continues pipeline. | Gracefully proceeds with degraded velocity flag `velocity_degraded: true`. | ❌ No | High |
| 13 | **PostgreSQL unavailable** | Logged as error in tx; API returns decision in-memory without saving to database. | Returns HTTP `500` or queued in fallback local cache to prevent lost decisions. | ❌ No | Critical |
| 14 | **ML service unavailable** | Activates heuristic fallback (`calculateFallbackRiskScore`); sets `is_degraded: true`. | Fast heuristic fallback under 5ms. | ✅ Yes | Medium |
| 15 | **ML sidecar timeout (>50ms)** | Context timeout triggers; cancels request and uses fallback heuristic. | Guarantees strict <100ms API response SLA. | ✅ Yes | Medium |
| 16 | **Invalid ML response JSON** | Decodes with error; orchestrator logs and falls back to heuristic score. | Handled gracefully without crashing Go API. | ❌ No | High |
| 17 | **Malformed JSON-AST rule** | `ParseRuleDefinition()` returns error; rule is skipped during loop. | Rule compilation skipped with error logged; pipeline continues. | ✅ Yes | Low |
| 18 | **Conflicting rules (Decline vs Allow)** | First matched hard pre-rule wins (loop order dependent). | Deterministic priority resolution (e.g. DECLINE > MANUAL_REVIEW > ALLOW). | ❌ No | High |
| 19 | **No active rules present** | Pipeline proceeds directly to ML inference and threshold mapping. | Clean fallback to pure ML evaluation without error. | ✅ Yes | Low |
| 20 | **Shared IP (NAT / University / Corporate)** | IP velocity increases for all users sharing IP; may trigger `HIGH_IP_VELOCITY`. | IP velocity tempered by device fingerprint and payment token correlation. | ❌ No | High |
| 21 | **High IP velocity (>5 attempts/hour)** | Caught by Redis ZSET query and rule/ML flags `HIGH_IP_VELOCITY_1H`. | High velocity detected and flagged. | ❌ No | Medium |
| 22 | **High token velocity (>6 attempts/24h)** | Caught by Redis ZSET query and flagged by ML and rules. | Token velocity detected and flagged. | ❌ No | Medium |
| 23 | **Dispute for unknown transaction** | Webhook inserts dispute record with `decision_id = NULL` and empty evidence packet. | Dispute stored; flagged as unlinked for analyst manual reconciliation. | ❌ No | Medium |
| 24 | **Duplicate webhook event ID** | Re-processed; inserts additional row in `disputes` table without unique constraint check. | Idempotent webhook handling (returns `200 OK` without duplicating dispute record). | ❌ No | High |
| 25 | **Invalid webhook HMAC signature** | Returns `401 Unauthorized` (`invalid_signature`). | Rejects spoofed or unsigned webhooks with HTTP 401. | ✅ Yes | High |
| 26 | **Case already claimed by another analyst** | `ClaimCase` overwrites `assigned_to` with current analyst without optimistic lock. | Optimistic concurrency check (`WHERE assigned_to IS NULL` or explicit re-assignment). | ❌ No | Medium |
| 27 | **Case already resolved / closed** | `ResolveCase` allows updating status of already resolved cases. | State machine guard (`400 Case Already Resolved`). | ❌ No | Medium |
| 28 | **Missing fingerprint / blank string** | `IsNewDevice` set to 0 or 1; empty string encrypted at rest. | Handled gracefully; flags `MISSING_DEVICE_TELEMETRY`. | ❌ No | Low |
| 29 | **Low-confidence device identity** | Processed as regular device hash without confidence score weighting. | Incorporates confidence score from telemetry provider. | ❌ No | Medium |
| 30 | **Future timestamp / Clock skew** | Redis uses client/server `time.Now()`; outbox records `time.Now().UTC()`. | Rejects transactions with timestamps >5 minutes in the future. | ❌ No | Low |
