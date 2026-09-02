# 🛡️ Buildathon Submission Evidence & Requirements Map

> **Track 02: AI Risk Manager (Razorpay AI Buildathon)**
> **Repository**: `shankywho/ropus`
> **Core Submission Proposition**: "ROPUS receives Razorpay payment events, estimates fraud risk using deterministic signals and a calibrated model, evaluates cost-sensitive policy recommendations, and produces an auditable analyst-ready result."

---

## 1. Track Requirements Compliance Matrix

| Track 02 Requirement | Implementation Component | Verification Artifact / Test | Status |
|:---|:---|:---|:---:|
| **1. Complete Defense-Only Fraud Detection Loop** | `backend/internal/ingestion/razorpay_adapter.go`<br/>`backend/internal/riskengine/orchestrator.go` | `cd backend && go test -v ./internal/ingestion/...`<br/>`make demo-webhook` | ✅ **VERIFIED** |
| **2. Real Razorpay Webhook Integration & Auth** | HMAC-SHA256 signature verification in `RazorpayWebhookAdapter`<br/>(`X-Razorpay-Signature`) | `backend/internal/ingestion/razorpay_adapter_test.go`<br/>(Tests: valid sig, invalid sig, missing sig) | ✅ **VERIFIED** |
| **3. Held-Out Evaluation Metrics (Precision/Recall)** | Frozen 1,200-row held-out split (`sample_ieee_fixture.csv`) | `python3 ml-service/evaluation/evaluate_frozen_holdout.py`<br/>`pytest ml-service/tests/test_metrics_consistency.py` | ✅ **VERIFIED** |
| **4. Explicit False-Positive Cost & BMR Policy** | Bayes Minimum Risk Loss evaluation in `orchestrator.go` & `cost_policy.go` | `METRICS.md` (Cost comparison table & transparent policy analysis) | ✅ **VERIFIED** |
| **5. Thoughtful Failure Recovery & Isolation** | 8ms context deadline timeout circuit breaker fallback to in-memory AST rules | `backend/internal/orchestrator_test.go`<br/>(`TestHighPerformanceOrchestratorSuite`) | ✅ **VERIFIED** |
| **6. Non-Blocking Memory Saturation Protection** | Non-blocking `select` channel write on SHA-256 audit ledger with atomic drop metric | `backend/internal/riskengine/orchestrator.go`<br/>(`GetDroppedAuditEventsTotal`) | ✅ **VERIFIED** |
| **7. Immutable Audit Trail** | SHA-256 chained Merkle ledger | `backend/internal/governance/audit_trail.go`<br/>`backend/internal/orchestrator_test.go` | ✅ **VERIFIED** |
| **8. Honest Disclosure of Boundaries** | `docs/limitations.md`<br/>`README.md` disclaimers | Explicit disclosure of synthetic test dataset and sandbox prototype scope | ✅ **VERIFIED** |

---

## 2. Verification Commands & Concrete Outputs

### A. Razorpay Webhook Ingestion & Auth Test
```bash
cd backend && go test -v ./internal/ingestion/...
```
**Output**:
```text
=== RUN   TestRazorpayWebhookAdapter_ValidSignedPayment
--- PASS: TestRazorpayWebhookAdapter_ValidSignedPayment (0.00s)
=== RUN   TestRazorpayWebhookAdapter_InvalidSignature
--- PASS: TestRazorpayWebhookAdapter_InvalidSignature (0.00s)
=== RUN   TestRazorpayWebhookAdapter_MissingSignature
--- PASS: TestRazorpayWebhookAdapter_MissingSignature (0.00s)
=== RUN   TestRazorpayWebhookAdapter_MalformedPayload
--- PASS: TestRazorpayWebhookAdapter_MalformedPayload (0.00s)
=== RUN   TestRazorpayWebhookAdapter_IdempotentReplay
--- PASS: TestRazorpayWebhookAdapter_IdempotentReplay (0.00s)
=== RUN   TestRazorpayWebhookAdapter_HighRiskEventRejection
--- PASS: TestRazorpayWebhookAdapter_HighRiskEventRejection (0.00s)
PASS
ok  	github.com/shankywho/ropus/backend/internal/ingestion	0.816s
```

### B. High-Performance Concurrency & 8ms Timeout Fallback (100 Goroutines)
```bash
cd backend && go test -race -v ./internal -run TestHighPerformanceOrchestratorSuite
```
**Output**:
```text
=== RUN   TestHighPerformanceOrchestratorSuite
=== RUN   TestHighPerformanceOrchestratorSuite/Verification_Zero_Allocations_In_Hot_Filter_Path
    orchestrator_test.go:334: ✓ Zero allocations verified: 0.000000 allocs/op in hot path using sync.Pool
=== RUN   TestHighPerformanceOrchestratorSuite/Verification_100_Goroutines_Concurrent_Stress_And_8ms_Timeout_Fallback
    orchestrator_test.go:428:   Total Requests Evaluated:   5000 / 5000
    orchestrator_test.go:429:   Elapsed Time:               10.40s
    orchestrator_test.go:430:   Throughput:                 480.77 req/sec
    orchestrator_test.go:433:   Local AST Rule Resolutions: 4992
    orchestrator_test.go:457: ✓ SHA-256 Audit Ledger Final Block Hash: 79698b397f18760767c1291e9ea00670ddd98b5c6de0571045a33cc97b038598
=== RUN   TestHighPerformanceOrchestratorSuite/Verification_Audit_Channel_Saturation_Drop_And_Metric
    orchestrator_test.go:493: ✓ Non-blocking channel protection verified: safely dropped 18 audit entries during channel saturation without thread stalling
--- PASS: TestHighPerformanceOrchestratorSuite (10.91s)
PASS
```

### C. ML Frozen Holdout Evaluation & Metrics Consistency Test
```bash
make evaluate
pytest ml-service/tests/test_metrics_consistency.py
```
**Output**:
```text
Loading IEEE-CIS sample fixture from ml-service/data/sample_ieee_fixture.csv...
{
  "total_holdout_transactions": 1200,
  "positive_labels_in_holdout": 52,
  "negative_labels_in_holdout": 1148,
  "metrics": {
    "roc_auc": 0.9376,
    "pr_auc": 0.6684,
    "precision": 0.7097,
    "recall": 0.8462,
    "f1_score": 0.7719,
    "expected_calibration_error_ece": 0.0119
  }
}
pytest ml-service/tests/test_metrics_consistency.py
1 passed in 0.02s
```

---

## 3. Defense-Only Architecture Traceability

| Step | Subsystem | File & Function |
|:---|:---|:---|
| **1. Webhook Intake** | `ingestion` | [HandleRazorpayWebhook](../backend/internal/ingestion/razorpay_adapter.go#L142) |
| **2. Auth & Idempotency** | `ingestion` | [VerifyRazorpaySignature](../backend/internal/ingestion/razorpay_adapter.go#L131) |
| **3. Risk Orchestration** | `riskengine` | [Evaluate](../backend/internal/riskengine/orchestrator.go#L600) |
| **4. Feature Context** | `features` | [ExtractMetrics](../backend/internal/features/velocity.go#L50) |
| **5. Model & Calibration** | `ml` / `ml-service` | [Predict](../backend/internal/riskengine/mlclient.go) & [calibrator.py](../ml-service/calibration/calibrator.py) |
| **6. BMR Loss Engine** | `riskengine` | [EvaluateCostSensitiveDecision](../backend/internal/riskengine/cost_policy.go) |
| **7. Non-Blocking Audit** | `riskengine` | Non-blocking `select` write to `auditChannel` |

---

## 4. Operational Boundaries & Honest Disclosures

1. **Synthetic Data Fixture**: Evaluated on a 1,200-transaction frozen held-out slice generated from IEEE-CIS fraud benchmarks. It is strictly offline and does not represent live Razorpay merchant transactions.
2. **Local Sandbox Execution**: The system is designed to run locally via Docker Compose or native binaries.
3. **No Financial Settlement**: ROPUS outputs purely defensive risk decisions (`ALLOW`, `STEP_UP`, `MANUAL_REVIEW`, `DECLINE`). It does not capture card payments or transfer money.
