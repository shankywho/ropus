# ROPUS Engineering Documentation & Backend Developer Guide

Welcome to the **ROPUS AI Risk Manager** technical documentation. This documentation repository is designed for backend engineers, system architects, and security auditors to understand, contribute to, and operate the ROPUS platform.

---

## 🧭 2-Hour New Backend Engineer Onboarding Roadmap

If you are a new engineer joining the team, follow this sequential reading path to get fully up to speed:

| Step | Time | Document | What You Will Learn |
| :---: | :---: | :--- | :--- |
| **1** | 15 mins | [`architecture/overview.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/overview.md) | High-level dataflow, 6-tier architecture, and latency budgets. |
| **2** | 20 mins | [`components/01-product-api.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/01-product-api.md) | The canonical `POST /v1/risk/evaluate` execution pipeline and factor attribution. |
| **3** | 20 mins | [`components/02-risk-engine.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/02-risk-engine.md) & [`components/03-rules-engine.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/03-rules-engine.md) | Composite scoring, decision precedence hierarchy, and rule evaluators. |
| **4** | 20 mins | [`components/04-ml-inference.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/04-ml-inference.md) & [`components/05-fraud-graph.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/05-fraud-graph.md) | 25-feature XGBoost inference and in-memory 3-hop entity graph traversal. |
| **5** | 15 mins | [`components/07-ai-investigators.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/07-ai-investigators.md) & [`components/08-cases-governance.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/08-cases-governance.md) | Autonomous LLM investigator agents and analyst case review queue. |
| **6** | 15 mins | [`components/09-auth-and-tenancy.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/09-auth-and-tenancy.md) & [`architecture/security-model.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/security-limitations.md) | Zero-trust API keys, RBAC, and AES-256 GCM encryption. |
| **7** | 15 mins | [`api/quickstart.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/api/quickstart.md) & [`demo/judge-cheatsheet.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/demo/judge-cheatsheet.md) | API integration quickstart and 20 core architectural Q&As. |

---

## 📚 Complete Documentation Index

### 1. Architecture & Security
- [`architecture/overview.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/overview.md) — End-to-End System Architecture Specification
- [`architecture/ropus-architecture.excalidraw`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/ropus-architecture.excalidraw) — Visual Excalidraw Architecture Diagram
- [`architecture/truth-audit.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/truth-audit.md) — Implementation Truth & Capability Audit
- [`architecture/security-limitations.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/security-limitations.md) — Security Controls & Threat Defense Boundaries
- [`architecture/production-architecture.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/production-architecture.md) — Production Deployment Topologies & Latency Budgets

---

### 2. Component Deep-Dives
1. [`components/01-product-api.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/01-product-api.md) — Product API Layer & Unified Evaluation Pipeline
2. [`components/02-risk-engine.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/02-risk-engine.md) — Risk Evaluation Engine & Precedence Matrix
3. [`components/03-rules-engine.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/03-rules-engine.md) — Rules Engine & Velocity Condition Evaluator
4. [`components/04-ml-inference.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/04-ml-inference.md) — Real ML Inference Engine & Feature Vectors
5. [`components/05-fraud-graph.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/05-fraud-graph.md) — Fraud Knowledge Graph 3.0 & Syndicate Discovery
6. [`components/06-threat-intelligence.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/06-threat-intelligence.md) — Threat Intelligence & Behavioral Telemetry
7. [`components/07-ai-investigators.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/07-ai-investigators.md) — Autonomous AI Investigation Agents & Council
8. [`components/08-cases-governance.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/08-cases-governance.md) — Case Review Management & Human Governance
9. [`components/09-auth-and-tenancy.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/09-auth-and-tenancy.md) — Authentication, API Keys & Multi-Tenant Isolation
10. [`components/10-streaming-kafka.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/10-streaming-kafka.md) — Event Streaming & Apache Kafka Architecture
11. [`components/11-storage-persistence.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/11-storage-persistence.md) — Data Storage, Redis Feature Store & Field Encryption
12. [`components/12-resilience-circuit-breaker.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/12-resilience-circuit-breaker.md) — Fault Tolerance, Circuit Breakers & Fallback Buffers
13. [`components/13-webhooks.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/13-webhooks.md) — Webhook Egress & Cryptographic Delivery
14. [`components/14-observability-slo.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/14-observability-slo.md) — Observability, Metrics & Contractual SLO Engine

---

### 3. API & Integration
- [`api/quickstart.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/api/quickstart.md) — 15-Minute Integration Quickstart (Python & Node.js)
- [`api/api-reference.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/api/api-reference.md) — OpenAPI & REST Endpoint Specifications
- [`api/openapi.yaml`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/api/openapi.yaml) — OpenAPI 3.0 Specification YAML

---

### 4. Operations & Reliability
- [`operations/disaster-recovery.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/operations/disaster-recovery.md) — Multi-AZ Failover & Continuous S3 WAL Archiving
- [`operations/incident-response.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/operations/incident-response.md) — Security Incident Playbooks & Escalation Paths
- [`operations/readiness-audit.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/operations/final-readiness.md) — Institutional Readiness Audit Report

---

### 5. Compliance & Governance
- [`compliance/SOC2.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/compliance/SOC2.md) — SOC 2 Trust Services Criteria Mapping
- [`compliance/PCI_DSS.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/compliance/PCI_DSS.md) — PCI-DSS Level 1 Service Provider Alignment
- [`compliance/ISO27001.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/compliance/ISO27001.md) — ISO 27001 Information Security Controls
- [`compliance/NIST_AI_RMF.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/compliance/NIST_AI_RMF.md) — NIST AI Risk Management Framework
- [`compliance/MODEL_RISK_MANAGEMENT.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/compliance/MODEL_RISK_MANAGEMENT.md) — SR 11-7 Supervisory Guidance on Model Risk

---

### 6. Presentation, Demonstrations & Audits
- [`demo/final-demo-script.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/demo/final-demo-script.md) — 5-Minute Timed Presenter Runbook
- [`demo/judge-cheatsheet.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/demo/judge-cheatsheet.md) — 20 Technical Q&As for Investors and Judges
- [`demo/investor-story.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/demo/investor-story.md) — Problem, Solution, Differentiation, and Moats
- [`demo/customer-story.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/demo/customer-story.md) — Customer Value Proposition & Onboarding
- [`demo/fraud-scenarios.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/demo/fraud-scenarios.md) — Canonical Fraud Attack Topologies

---

### 7. Historical Archive
- [`archive/legacy_phases/`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/archive/legacy_phases/) — Historical milestone reports for Phases 1 through 3.39.
