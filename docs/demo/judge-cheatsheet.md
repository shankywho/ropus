# ROPUS — Hackathon & Investor Technical Cheatsheet

Quick, technically defensible answers to the 20 most critical judging questions:

---

### 1. What problem does ROPUS solve?
Stops coordinated, automated financial fraud (account takeovers, synthetic identities, card testing, money mule rings) that bypass traditional static rule engines.

### 2. Why can't normal rule engines solve it?
Static rules evaluate single transactions in isolation. Fraud rings operate across multiple accounts, rotating IPs and spoofing emulators, which require real-time graph traversal and behavioral modeling to detect.

### 3. What is actually AI?
Gradient-boosted decision trees (XGBoost/LightGBM) trained on transaction features, and LLM-powered investigation agents (Claude 3.7 / GPT-4o) that synthesize evidence dossiers from graph and threat intelligence.

### 4. What is actually real?
The full decisioning pipeline, ML mathematical inference, in-memory graph traversal, SHA-256 API key hashing, AES-256 GCM encryption, HMAC webhooks, rate limiting, and the Next.js portal.

### 5. What is simulated?
The upstream core banking ledger is driven by our high-scale synthetic world generator to simulate high-volume card and wire traffic without live SWIFT connections.

### 6. Why is graph intelligence useful?
It instantly reveals when a seemingly new user shares an emulator canvas, device fingerprint, or subnet with 14 previously confirmed fraud accounts.

### 7. Why use ML + rules together?
Rules enforce hard regulatory ceilings and compliance blacklists, while ML captures subtle non-linear multi-variable anomalies.

### 8. What does the LLM do?
It acts as an autonomous tier-1 fraud investigator, reading graph edges, velocity metrics, and threat feeds to draft structured investigation dossiers for human analysts.

### 9. How do you prevent hallucinations?
The AI investigator is constrained to strictly distinguish between **Observed Facts** (verifiable data in storage) and **Inferred Patterns** (hypotheses), referencing exact transaction IDs and IP addresses.

### 10. How is explainability achieved?
Every risk score is decomposed into exact additive factor contributions (e.g. Impossible Travel +0.21, Velocity +0.22, Device Novelty +0.18, ML +0.20) that sum directly to the composite score.

### 11. How does tenant isolation work?
All database queries and cache keys are scoped to Organization IDs. Cross-tenant access is rejected at the repository layer.

### 12. How does the system scale?
Stateless evaluation workers auto-scale horizontally on Kubernetes (HPA), with low-latency Redis feature caching and in-memory graph indexing.

### 13. What happens when Kafka goes down?
The `HealthManager` trips a circuit breaker and buffers events in a local fallback queue, ensuring core transaction evaluations are never blocked.

### 14. What happens when the LLM goes down?
The synchronous risk decision and case creation continue unaffected; the investigation dossier generation is queued for asynchronous retry.

### 15. What happens when PostgreSQL goes down?
Secondary read replicas take over query load while write-ahead logs archive continuously to S3 for automated point-in-time recovery.

### 16. How do you prevent false positives?
Adaptive scoring applies step-up biometric/MFA challenges on medium-risk bands (0.50–0.79) rather than immediate hard blocks.

### 17. How would a bank integrate ROPUS?
Send a JSON payload to `POST /v1/risk/evaluate` and receive a decision and factor breakdown in $< 2\text{ms}$.

### 18. What is the business model?
Tiered B2B SaaS subscriptions: Starter ($499/mo), Growth ($4,999/mo), and Enterprise Dedicated ($24,999/mo) with metered overages.

### 19. What is the competitive moat?
Sub-10ms fusion of real-time graph traversal, ML probability, and agentic investigation dossiers combined with immutable hash-chained audit trails.

### 20. What would you build next?
Federated cross-bank consortium intelligence sharing using zero-knowledge proofs to correlate fraud rings without exposing raw PII between institutions.
