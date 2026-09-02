# 📝 Razorpay AI Buildathon — Submission Guide & Form Answers

> **Track 02**: AI Risk Manager
> **Project**: ROPUS — Real-Time Payment Risk & Economic Decisioning Platform
> **Repository**: [github.com/shankywho/ropus](https://github.com/shankywho/ropus)

---

## 1. Google Form Field Answers (Ready to Copy & Paste)

### Field 1: Project Name / Title
```text
ROPUS — AI Risk Manager for Payments & Economic Loss Minimization
```

### Field 2: Project Objectives (What does it solve?)
```text
ROPUS receives Razorpay payment events, estimates fraud risk using deterministic signals and a calibrated model, chooses the lowest-cost recommendation, and produces an auditable analyst-ready result.

Traditional fraud engines rely on static binary thresholds (e.g. if probability > 0.50, block transaction), which cause high false-decline rates and costly customer friction. ROPUS solves this by combining:
1. High-throughput Go risk orchestration with point-in-time velocity feature extraction.
2. Gradient boosted model (XGBoost) with post-hoc Beta probability calibration (ECE = 0.0119 on held-out test split).
3. Bayes Minimum Risk (BMR) financial loss matrix that balances false-positive customer friction against false-negative chargeback losses across dynamic transaction tiers.
4. Native Razorpay HMAC-SHA256 signed webhook ingestion with idempotent replay protection and immutable SHA-256 audit chaining.
```

### Field 3: Build Challenges & Technical Obstacles
```text
1. Probability Calibration under Severe Class Imbalance:
Tree-based gradient boosters trained on payment fraud datasets (~4-5% prevalence) output uncalibrated score distributions that cluster near extremes. Feeding uncalibrated logits into a financial loss matrix produces distorted decision boundaries. We solved this by implementing post-hoc Beta Calibration, which fits a bivariate logistic transformation ln(p/(1-p)) = a*ln(p) - b*ln(1-p) + c strictly on the validation split, achieving an Expected Calibration Error (ECE) of 0.0119 while preserving continuous probability resolution.

2. Strict Temporal Point-in-Time Causality (Zero Future Leakage):
In payment fraud, entity aggregations (e.g., rolling device burst counts, token velocities) often suffer from subtle target leakage if calculated across future time horizons. We architected a causal time-windowed preprocessor that strictly guarantees transaction T only aggregates historical events where T_event < T_tx, verified across a 1,200-transaction frozen out-of-time holdout test set (ROC-AUC 0.9376, Recall 84.62%, Precision 70.97%).

3. Resilient Fallback & Memory Protection under High Ingress Bursts:
High-concurrency webhooks can cause downstream sidecar latency spikes. We implemented an 8ms context deadline circuit breaker that degrades seamlessly to in-memory JSON-AST rules with zero downstream SQL/network I/O, alongside a non-blocking asynchronous audit ledger that protects the hot payment path from memory saturation.
```

---

## 2. 5-Minute Pitch Video Script & Structure (Cheat Sheet)

| Time | Slide / Screen | What to Say & Demo |
|---|---|---|
| **0:00 – 0:45** | Problem Statement & Proposition | *"Payment fraud is costly in two ways: direct chargebacks, and false declines that turn away legitimate customers. ROPUS receives Razorpay payment events, scores risk with a calibrated model, and evaluates cost-sensitive BMR decision policies."* |
| **0:45 – 1:45** | Live Webhook Ingress Demo | Run `./scripts/demo_buildathon.sh`. Show Scenario A (₹480 domestic payment $\rightarrow$ ALLOW), Scenario B (₹1,50,000 international card from disposable email $\rightarrow$ DECLINE), Scenario C (Forged signature rejected with HTTP 401), and Scenario D (Idempotent replay). |
| **1:45 – 2:45** | ML Calibration & Cost Impact | Show [METRICS.md](../METRICS.md). *"Our Beta-calibrated model achieves an Expected Calibration Error of 1.19% and ROC-AUC of 0.9376 on the 1,200-transaction frozen held-out partition, providing well-calibrated probabilities for economic loss modeling."* |
| **2:45 – 3:45** | Resilient Failure Recovery | Show concurrent benchmark: 100 goroutines, 8ms context timeout gracefully falling back to deterministic AST rules with zero downstream SQL lookups. |
| **3:45 – 5:00** | Audit Ledger & Conclusion | Verify immutable SHA-256 audit ledger hash (`/v1/audit/verify`). *"ROPUS provides a complete, defense-only, reproducible risk management solution ready for Track 02."* |
