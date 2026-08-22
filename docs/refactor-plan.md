# Documentation Refactor Plan: Comprehensive Engineering Architecture

## Objectives
Transform the 106+ sprawling, fragmented documentation files into a pristine, structured, and deeply technical documentation repository for backend engineers, system architects, and technical auditors.

---

## Phase 1: Audit & Disposition Matrix

### 1. Primary Documentation (KEEP & REORGANIZE)
- `docs/architecture.md` -> Move to `docs/architecture/overview.md`
- `docs/production-architecture.md` -> Merge into `docs/architecture/overview.md`
- `docs/quickstart.md` -> Move to `docs/api/quickstart.md`
- `docs/api-reference.md` -> Move to `docs/api/reference.md`
- `docs/judge-cheatsheet.md` -> Move to `docs/demo/judge-cheatsheet.md`
- `docs/final-demo-script.md` -> Move to `docs/demo/demo-script.md`
- `docs/implementation-truth-audit.md` -> Move to `docs/architecture/truth-audit.md`
- `docs/security-limitations.md` -> Move to `docs/architecture/security-model.md`
- `docs/disaster-recovery.md` -> Move to `docs/operations/disaster-recovery.md`
- `docs/incident-response.md` -> Move to `docs/operations/incident-response.md`
- `docs/compliance/*` -> Retain in `docs/compliance/`
- `docs/governance/*` -> Retain in `docs/compliance/governance/`

### 2. Historical Progress Reports (ARCHIVE)
Move all 45+ incremental milestone reports (`docs/phase-1-report.md` through `docs/phase-3.39-final-validation.md`, `docs/phase-*.md`) into `docs/archive/legacy_phases/` to preserve historical trajectory while keeping the active onboarding path clean.

### 3. Redundant / Duplicated Scenarios (MERGE & DEDUPLICATE)
- Merge `docs/demo-script.md`, `docs/demo/investor-demo-script.md`, and `docs/demo/customer-demo-script.md` into a single canonical `docs/demo/demo-runbook.md`.
- Merge `docs/product-overview.md` and `docs/product-positioning.md` into `docs/architecture/product-model.md`.

---

## Phase 2: Target Clean Folder Hierarchy

```text
docs/
├── README.md                          <-- Root Documentation Index & 2-Hour Onboarding Roadmap
├── architecture/
│   ├── overview.md                    <-- System Architecture, Dataflow & Latency Budget
│   ├── product-model.md               <-- Risk Engine vs Point Solutions & Domain Model
│   ├── security-model.md              <-- Threat Boundaries, Cryptography & Tenant Isolation
│   └── truth-audit.md                 <-- Verified Implementation Capabilities vs Simulations
├── components/                        <-- Deep-Dive Engineering Guides (Phase 3)
│   ├── 01-product-api.md              <-- Canonical POST /v1/risk/evaluate Pipeline
│   ├── 02-risk-engine.md              <-- Composite Scoring, Precedence & Policy Thresholds
│   ├── 03-rules-engine.md             <-- Declarative Policies & Velocity Evaluators
│   ├── 04-ml-inference.md             <-- XGBoost Gradient Boosted Trees & Model Registry
│   ├── 05-fraud-graph.md              <-- In-Memory 3-Hop Entity Graph & Syndicate Detection
│   ├── 06-threat-intelligence.md      <-- IP Subnets, Proxies, Geo Travel & Behavioral Telemetry
│   ├── 07-ai-investigators.md         <-- Agent Council, LLM Gateway & Evidence Dossier Synthesis
│   ├── 08-cases-governance.md         <-- Review Queue, Analyst Actions & Closed-Loop Retraining
│   ├── 09-auth-and-tenancy.md         <-- SHA-256 API Keys, 4-Tier RBAC & Tenant Boundary Isolation
│   ├── 10-streaming-kafka.md          <-- Kafka Event Bus, DLQ Fallback & Telemetry Fanout
│   ├── 11-storage-persistence.md      <-- PostgreSQL Schemas, Redis Feature Store & AES-256 GCM
│   ├── 12-resilience-circuit-breaker.md <-- Health Manager, Degraded Modes & Fallback Queues
│   ├── 13-webhooks.md                 <-- HMAC-SHA256 Signatures, Delivery Retries & Dispatch
│   └── 14-observability-slo.md        <-- Prometheus Metrics, Tracing, SLO Engine & Alarming
├── api/
│   ├── quickstart.md                  <-- 15-Minute Integration (Python & Node.js)
│   ├── reference.md                   <-- Full OpenAPI Endpoint Specifications & Payloads
│   └── error-taxonomy.md              <-- Standardized Error Codes & Troubleshooting
├── operations/
│   ├── deployment.md                  <-- Kubernetes Helm Charts, Docker & AWS Terraform
│   ├── disaster-recovery.md           <-- Multi-AZ Failover, S3 WAL Archiving & Recovery Runbooks
│   └── incident-response.md           <-- Security Incidents, Rollback Procedures & Runbooks
├── compliance/
│   ├── SOC2.md
│   ├── PCI_DSS.md
│   ├── ISO27001.md
│   ├── NIST_AI_RMF.md
│   └── MODEL_RISK_MANAGEMENT.md
├── demo/
│   ├── canonical-scenario.md          <-- 17-Stage Transnational Mule Syndicate Attack
│   ├── demo-runbook.md                <-- 5-Minute Investor / Customer Presentation Script
│   └── judge-cheatsheet.md            <-- 20 Technical Q&As for Due Diligence
└── archive/
    └── legacy_phases/                 <-- Preserved Phase 1 to 3.39 Milestones
```

---

## Phase 3: Component Deep-Dive Structure (14 Subsystems)
Each component document in `docs/components/` will follow a standardized 14-section technical template:
1. **Purpose & Responsibilities**
2. **Inputs & Contract Schemas**
3. **Internal Workflow & Core Algorithms**
4. **Outputs & Side Effects**
5. **ASCII Interaction & Dataflow Diagram**
6. **Key Data Structures (Go Structs & DB/Redis Keys)**
7. **Error Handling & Degraded Fallback Behavior**
8. **Performance Budgets & Concurrency Guarantees**
9. **Security, Encryption & Tenant Isolation Controls**
10. **Code Map (Key Go Files & Line Ranges)**
11. **Concrete Execution Walkthrough**
12. **How to Extend / Implement New Capabilities**
13. **Common Developer Pitfalls & Anti-Patterns**
14. **Cross-Component Dependencies & References**
