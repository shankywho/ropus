# ROPUS Engineering Documentation & Developer Guide

Welcome to the **ROPUS AI Risk Manager** technical documentation. This documentation repository is designed for backend engineers, ML practitioners, system architects, and security auditors to understand, operate, and govern the ROPUS platform.

---

## 🧭 2-Hour New Engineer Onboarding Roadmap

If you are a new engineer joining the team, follow this sequential reading path to get fully up to speed:

| Step | Time | Document | What You Will Learn |
| :---: | :---: | :--- | :--- |
| **1** | 15 mins | [`architecture/overview.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/overview.md) | High-level dataflow, 6-tier architecture, and latency budgets. |
| **2** | 20 mins | [`components/01-product-api.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/01-product-api.md) | Canonical `POST /v1/risk/evaluate` execution pipeline and factor attribution. |
| **3** | 20 mins | [`components/02-risk-engine.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/02-risk-engine.md) & [`components/03-rules-engine.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/03-rules-engine.md) | Composite scoring, decision precedence hierarchy, and rule evaluators. |
| **4** | 20 mins | [`components/04-ml-inference.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/04-ml-inference.md) & [`components/05-fraud-graph.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/05-fraud-graph.md) | 36-feature CatBoost inference and in-memory entity graph traversal. |
| **5** | 15 mins | [`components/07-ai-investigators.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/07-ai-investigators.md) & [`components/08-cases-governance.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/08-cases-governance.md) | Autonomous LLM investigator agents and analyst case review queue. |
| **6** | 15 mins | [`components/09-auth-and-tenancy.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/09-auth-and-tenancy.md) & [`architecture/security-limitations.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/security-limitations.md) | Zero-trust API keys, RBAC, and HMAC-SHA256 authentication. |
| **7** | 15 mins | [`api/quickstart.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/api/quickstart.md) & [`demo/judge-cheatsheet.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/demo/judge-cheatsheet.md) | API integration quickstart and 20 core architectural Q&As. |

---

## 📚 Complete Documentation Index

### 1. Architecture & Security Specifications
- [`architecture/overview.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/overview.md) — End-to-End System Architecture Specification
- [`architecture/product-overview.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/product-overview.md) — Product Overview & Core Capabilities
- [`architecture/product-positioning.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/product-positioning.md) — Strategic Positioning & Competitive Moat
- [`architecture/graphsage-status-and-roadmap.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/graphsage-status-and-roadmap.md) — **GraphSAGE Status, Validation Matrix & Governance Roadmap (Phase 68 Checkpoint)**
- [`architecture/graphsage-relationship-intelligence.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/graphsage-relationship-intelligence.md) — GraphSAGE Internal Relationship Intelligence & Collusion Detection
- [`architecture/graphsage-post-provisioning-connectivity.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/graphsage-post-provisioning-connectivity.md) — Phase 68 Post-Provisioning Staging Connectivity Specification
- [`architecture/graphsage-real-staging-infrastructure-gate.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/graphsage-real-staging-infrastructure-gate.md) — Phase 67 Real Staging Infrastructure Gate Architecture
- [`architecture/graphsage-real-staging-activation.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/graphsage-real-staging-activation.md) — Phase 66 Real Staging Infrastructure Activation Architecture
- [`architecture/graphsage-real-staging-shadow.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/graphsage-real-staging-shadow.md) — Phase 65 Real Staging Shadow Preflight Architecture
- [`architecture/graphsage-real-staging-connectivity.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/graphsage-real-staging-connectivity.md) — Phase 64 Real Staging Connectivity Architecture
- [`architecture/graphsage-staging-shadow-soak.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/graphsage-staging-shadow-soak.md) — Phase 63 Staging Shadow Soak Architecture
- [`architecture/graphsage-live-shadow-integration.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/graphsage-live-shadow-integration.md) — Phase 62 Live Shadow Integration Architecture
- [`architecture/graphsage-production-shadow-telemetry.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/graphsage-production-shadow-telemetry.md) — Phase 61 Production Shadow Telemetry Architecture
- [`architecture/graphsage-real-data-shadow-ingestion.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/graphsage-real-data-shadow-ingestion.md) — Phase 60 Real Data Shadow Ingestion Architecture
- [`architecture/graphsage-real-data-shadow-replay.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/graphsage-real-data-shadow-replay.md) — Phase 59 Point-in-Time Shadow Replay Architecture
- [`architecture/graphsage-real-data-readiness.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/graphsage-real-data-readiness.md) — Phase 58 Real Data Readiness Architecture
- [`architecture/automated-attack-defense.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/automated-attack-defense.md) — Automated Attack, Bot Risk & Cadence Defense Architecture
- [`architecture/production-architecture.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/production-architecture.md) — Production Deployment Topologies & Latency Budgets
- [`architecture/risk-decision-precedence.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/risk-decision-precedence.md) — Dynamic Bayes Minimum Risk & Rule Priority Matrix
- [`architecture/risk-score-contract.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/risk-score-contract.md) — 36 Causal Feature & Scoring Contract
- [`architecture/security-baseline.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/security-baseline.md) — Enterprise Security Baseline & Defenses
- [`architecture/security-limitations.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/security-limitations.md) — Security Controls & Threat Defense Boundaries
- [`architecture/security-whitepaper.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/security-whitepaper.md) — Cryptographic Integrity & Data Isolation Whitepaper
- [`architecture/scaling-strategy.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/scaling-strategy.md) — Horizontal Scaling Architecture & Benchmarks
- [`architecture/technical-limitations.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/technical-limitations.md) — Architectural Boundaries & Latency Constraints
- [`architecture/truth-audit.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/truth-audit.md) — Implementation Truth & Capability Audit

---

### 1.1 Current GraphSAGE Implementation & Governance Status (Phase 68 Checkpoint)

> [!IMPORTANT]
> **Core Software Completion Notice**: The core engineering and automated software validation for the GraphSAGE relationship intelligence subsystem are **substantially complete** (44 Go tests, 124 Python tests passing with 100% green status).
>
> **Current Governance State**:
> * **Customer Decision Routing:** **100% Baseline Dynamic BMR Authoritative** (GraphSAGE customer enforcement = 0%).
> * **Operational Status:** Strictly non-enforcing shadow / investigation intelligence only.
> * **Staging Infrastructure Connectivity:** **`INFRASTRUCTURE_BLOCKED`** (Awaiting AWS STS staging credentials and MSK broker endpoints).
> * **Model Promotion Status:** **`PROMOTION_BLOCKED`** under `REAL_DATA_REQUIRED` (0 / 50 confirmed mature real internal collusion cases accumulated).
>
> See [`architecture/graphsage-status-and-roadmap.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/graphsage-status-and-roadmap.md) for the complete four-pillar status breakdown.


---

### 2. Component Deep-Dives
1. [`components/01-product-api.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/01-product-api.md) — Product API Layer & Unified Evaluation Pipeline
2. [`components/02-risk-engine.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/02-risk-engine.md) — Risk Evaluation Engine & Precedence Matrix
3. [`components/03-rules-engine.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/03-rules-engine.md) — Rules Engine & Velocity Condition Evaluator
4. [`components/04-ml-inference.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/04-ml-inference.md) — ML Inference Engine, 36F Pipeline & Beta Calibration
5. [`components/05-fraud-graph.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/05-fraud-graph.md) — Fraud Knowledge Graph & Syndicate Discovery
6. [`components/06-threat-intelligence.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/06-threat-intelligence.md) — Threat Intelligence & Behavioral Telemetry
7. [`components/07-ai-investigators.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/07-ai-investigators.md) — Autonomous AI Investigation Agents & Council
8. [`components/08-cases-governance.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/08-cases-governance.md) — Case Review Management & Human Governance
9. [`components/09-auth-and-tenancy.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/09-auth-and-tenancy.md) — Authentication, API Keys & Multi-Tenant Isolation
10. [`components/10-streaming-kafka.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/10-streaming-kafka.md) — Event Streaming & Apache Kafka Architecture
11. [`components/11-storage-persistence.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/11-storage-persistence.md) — Data Storage, Redis Feature Store & Field Encryption
12. [`components/12-resilience-circuit-breaker.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/12-resilience-circuit-breaker.md) — Fault Tolerance, Circuit Breakers & Fail-Open Isolation
13. [`components/13-webhooks.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/13-webhooks.md) — Webhook Egress & Cryptographic Dispute Delivery
14. [`components/14-observability-slo.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/14-observability-slo.md) — Observability, Metrics & Contractual SLO Engine

---

### 3. API & Integration Guides
- [`api/quickstart.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/api/quickstart.md) — 15-Minute Integration Quickstart (Python & Node.js)
- [`api/api-reference.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/api/api-reference.md) — OpenAPI & REST Endpoint Specifications
- [`api/customer-integration.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/api/customer-integration.md) — Customer Integration Best Practices
- [`api/customer-onboarding.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/api/customer-onboarding.md) — Merchant Gateway Onboarding Guide

---

### 4. Operations, Reliability & Dashboards
- [`operations/disaster-recovery.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/operations/disaster-recovery.md) — Multi-AZ Failover & Continuous S3 WAL Archiving
- [`operations/incident-response.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/operations/incident-response.md) — Security Incident Playbooks & Escalation Paths
- [`operations/enterprise-readiness.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/operations/enterprise-readiness.md) — Enterprise Readiness Assessment
- [`operations/final-readiness.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/operations/final-readiness.md) — Production Readiness Audit Report
- [`demo-runbook.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/demo-runbook.md) — Live Verification & Demonstration Runbook
- [`grafana/dashboards.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/grafana/dashboards.md) — Grafana Telemetry & Alerting Dashboards

---

### 5. Compliance & Governance Frameworks
- [`compliance/SOC2.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/compliance/SOC2.md) — SOC 2 Type II Trust Services Criteria Mapping
- [`compliance/PCI_DSS.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/compliance/PCI_DSS.md) — PCI-DSS Level 1 Service Provider Alignment
- [`compliance/ISO27001.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/compliance/ISO27001.md) — ISO 27001 Information Security Controls
- [`compliance/NIST_AI_RMF.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/compliance/NIST_AI_RMF.md) — NIST AI Risk Management Framework 1.0
- [`compliance/MODEL_RISK_MANAGEMENT.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/compliance/MODEL_RISK_MANAGEMENT.md) — SR 11-7 Supervisory Guidance on Model Risk
- [`governance/ai_risk_policy.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/governance/ai_risk_policy.md) — AI Risk & Ethics Policy
- [`governance/model_approval_process.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/governance/model_approval_process.md) — Formal Model Promotion & Governance Gates
- [`governance/regulatory_compliance.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/governance/regulatory_compliance.md) — Global Regulatory Alignment

---

### 6. Presentation & Demonstration Assets (`docs/demo/`)
- [`demo/final-demo-script.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/demo/final-demo-script.md) — Timed Presenter Runbook
- [`demo/judge-cheatsheet.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/demo/judge-cheatsheet.md) — 20 Technical Q&As for Investors and Judges
- [`demo/investor-story.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/demo/investor-story.md) — Problem, Solution, Differentiation, and Moats
- [`demo/customer-story.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/demo/customer-story.md) — Customer Value Proposition & ROI Analysis
- [`demo/fraud-scenarios.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/demo/fraud-scenarios.md) — Canonical Fraud Attack Topologies

---

### 7. Authoritative ML Evaluation Documents
- **ML Evaluation Index**: [`ml-service/evaluation/README.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/ml-service/evaluation/README.md)
  - **Master Production Completion Report**: [`ml-service/evaluation/ROPUS_PRODUCTION_COMPLETION_REPORT.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/ml-service/evaluation/ROPUS_PRODUCTION_COMPLETION_REPORT.md)
  - **Model Performance Diagnosis Report**: [`ml-service/evaluation/MODEL_PERFORMANCE_DIAGNOSIS_REPORT.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/ml-service/evaluation/MODEL_PERFORMANCE_DIAGNOSIS_REPORT.md)
  - **Weekly Production Health Report**: [`ml-service/evaluation/WEEKLY_PRODUCTION_HEALTH_REPORT.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/ml-service/evaluation/WEEKLY_PRODUCTION_HEALTH_REPORT.md)
  - **GraphSAGE Status & Roadmap**: [`docs/architecture/graphsage-status-and-roadmap.md`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/architecture/graphsage-status-and-roadmap.md)
