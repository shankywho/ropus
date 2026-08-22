# ROPUS Technical Limitations & Real vs. Simulated Architecture

Honest technical disclosure for enterprise security audits and investor due diligence:

---

## 1. Implemented Technical Controls vs. Third-Party Certification
- **Implemented Controls**: AES-256 GCM authenticated encryption at rest, TLS 1.3 in-transit enforcement, one-way SHA-256 API key hashing, HMAC-SHA256 webhook signatures, and SQL injection parameter sanitization are fully written and verified in the codebase.
- **Formal Certification**: SOC 2 Type II, PCI-DSS Level 1 Service Provider, and ISO 27001 formal certifications require external accredited third-party auditing periods and organizational compliance reviews.

---

## 2. Real vs. Simulated Subsystems

| Architecture Layer | Status | Description |
| :--- | :---: | :--- |
| **Real ML Inference** | **REAL** | XGBoost / LightGBM gradient boosted tree scoring running mathematical probability transforms. |
| **Unified Decision Pipeline** | **REAL** | End-to-end API key auth, sanitization, factor attribution, case creation, and webhook dispatching. |
| **Multi-Tenant SaaS & RBAC** | **REAL** | Strict tenant boundary separation, SHA-256 key hashing, and role verification. |
| **Circuit Breakers & Resilience** | **REAL** | Stateful circuit breaker tripping and local queue fallback buffering. |
| **Upstream Core Bank Ledger** | **SIMULATED** | Driven by the `simulator/` world generator to create realistic high-volume transaction traffic without connecting to live SWIFT/Fedwire networks. |
| **External LLM Cloud Gateway** | **HYBRID** | Can run locally against mocked inference models or connect to live Anthropic / OpenAI API endpoints. |
