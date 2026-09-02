# 🎬 Five-Minute Evaluation & Video Demonstration Runbook

> **Track 02: AI Risk Manager (Razorpay AI Buildathon)**
> **Target Duration**: Exactly 5 Minutes (0:00 – 5:00)
> **Demonstration Script**: `scripts/demo_buildathon.sh`

---

## ⏱️ Video Time Budget & Script Breakdown

```
[0:00 - 0:45] Minute 1: The Problem & 60-Second Stack Overview
[0:45 - 1:45] Minute 2: Signed Razorpay Webhook Ingress (Low-Risk vs High-Risk)
[1:45 - 2:45] Minute 3: Bayes Minimum Risk (BMR) False-Positive Cost Minimization
[2:45 - 3:45] Minute 4: Controlled Failure & 8ms Timeout Circuit Breaker Fallback
[3:45 - 4:45] Minute 5: Immutable SHA-256 Audit Trail & Analyst Control Plane
[4:45 - 5:00] Wrap-up: Reproducibility & Track 02 Submission Checklist
```

---

### Minute 1 (0:00 – 0:45): The Problem & Stack Overview
- **Opening Statement**: *"Welcome to ROPUS, built for Razorpay AI Buildathon Track 02: AI Risk Manager. Online merchants lose money in two ways: through fraud chargebacks, and through false declines that reject legitimate buyers. ROPUS solves this with a calibrated model and Bayes Minimum Risk cost optimization."*
- **Visual**: Show terminal running `docker compose ps` / `make up` and opening [http://localhost:3000](http://localhost:3000).

---

### Minute 2 (0:45 – 1:45): Signed Razorpay Webhook Ingress
- **Action**: Run `./scripts/demo_buildathon.sh` (or show Step 2 & 3 in UI).
- **Scenario A (Normal Domestic Payment)**:
  - Payload: ₹480 domestic HDFC card.
  - Verification: Valid HMAC-SHA256 signature in `X-Razorpay-Signature`.
  - Output: `ALLOW_RECOMMENDATION` (Risk score: 12/100).
- **Scenario B (Coordinated Whale Attack)**:
  - Payload: ₹1,50,000 international card from a disposable `@guerrillamail.com` domain.
  - Output: `DECLINE_RECOMMENDATION` (Risk score: 99/100, Reason codes: `HIGH_VALUE_WHALE_THRESHOLD`, `DISPOSABLE_EMAIL_DOMAIN`).
- **Scenario C (Tampered Signature Rejection)**:
  - Payload with forged signature $\rightarrow$ Immediately rejected with `HTTP 401 Unauthorized`.

---

### Minute 3 (1:45 – 2:45): Bayes Minimum Risk & False-Positive Cost
- **Narrative**: *"Why is probability calibration critical? Traditional fraud engines use uncalibrated scores or naive static thresholds. ROPUS pairs Beta calibration with Bayes Minimum Risk to evaluate trade-offs between chargeback risk and customer friction."*
- **Visual**: Show `METRICS.md` or the BMR Cost Matrix in the frontend UI.
- **Evidence**:
  - Uncalibrated Model ECE: 6.88% $\rightarrow$ Beta Calibrated Model ECE: **1.19%** (Brier Score: 0.0180).
  - ROC-AUC: **0.9376**, PR-AUC: **0.6684**, Recall: **84.62%** (44/52), Precision: **70.97%** (44/62), FPR: **1.57%** (18/1148) on the frozen held-out test split.
  - Transparent Economic Analysis: Evaluates realized losses under configurable false-positive and triage cost parameters (Modeled cost: ₹5,716.39 baseline vs ₹6,116.39 BMR policy).

---

### Minute 4 (2:45 – 3:45): Controlled Failure & 8ms Timeout Fallback
- **Demonstration**: Run `cd backend && go test -v ./internal -run TestHighPerformanceOrchestratorSuite`.
- **Narrative**: *"What happens when the downstream ML model experiences a network spike or outage during high-volume flash sales?"*
- **Proof**:
  - 100 concurrent goroutines firing 5,000 requests.
  - When the 8ms context deadline is exceeded, the orchestrator trips its circuit breaker.
  - **Zero Downstream I/O**: The system degrades instantly to in-memory JSON-AST rules using recycled scratch buffers, returning safe deterministic decisions in sub-millisecond latency.

---

### Minute 5 (3:45 – 5:00): Immutable Audit Trail & Final Summary
- **Visual**: Show the SHA-256 Audit Trail block hash generated from the batch ledger (`/v1/audit/verify`).
- **Proof of Idempotency**: Re-sending identical webhooks returns `X-ROPUS-Idempotent-Replayed: true` without creating duplicate transactions.
- **Conclusion**:
  - *"In summary, ROPUS delivers a complete, defense-only fraud management loop: signed webhook ingestion, calibrated probability scoring, cost-sensitive BMR recommendations, and cryptographic auditability."*

---

## 🛠️ Step-by-Step CLI Execution Guide for Judges

```bash
# 1. Start the local stack
make up

# 2. Execute the automated 5-minute demonstration
make demo-webhook

# 3. Reproduce held-out metrics from canonical JSON artifact
make evaluate

# 4. Verify test suite (backend + ML + frontend)
make test
```
