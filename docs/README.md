# ROPUS Technical Documentation Index

Welcome to the **ROPUS** technical documentation. This documentation repository covers the architecture, component implementations, API usage, and operational runbooks for the ROPUS payment risk decision engine.

---

## 🧭 Quick Links

- **[System Architecture Overview](architecture/overview.md)**: High-level dataflow, synchronous multi-stage pipeline, and sub-millisecond execution budget.
- **[API Quickstart](api/quickstart.md)**: How to send risk evaluation requests, create rules, and query decisions.
- **[API Reference](api/api-reference.md)**: Endpoints, request schemas, and response formats.
- **[Demo Walkthrough Script](demo/demo-script.md)**: Step-by-step guide to exercising live scenarios via the web UI and curl.
- **[Judge / Technical Q&A Cheatsheet](demo/judge-cheatsheet.md)**: Architectural defense guide and key technical trade-offs.

---

## 📚 Component Guides

| Component | Document | Description |
| :--- | :--- | :--- |
| **Product API** | [`components/01-product-api.md`](components/01-product-api.md) | Synchronous `POST /v1/risk/evaluate` execution pipeline and factor attribution. |
| **Risk Engine** | [`components/02-risk-engine.md`](components/02-risk-engine.md) | Composite scoring, decision precedence hierarchy, and Bayes Minimum Risk loss matrix. |
| **Rules Engine** | [`components/03-rules-engine.md`](components/03-rules-engine.md) | Deterministic JSON-AST rule evaluator with Maker-Checker dual control. |
| **ML Inference** | [`components/04-ml-inference.md`](components/04-ml-inference.md) | 25-feature XGBoost/ONNX model serving and probability calibration. |
| **Fraud Graph** | [`components/05-fraud-graph.md`](components/05-fraud-graph.md) | In-memory 3-hop BFS entity relationship graph with tenant namespacing. |
| **Threat Intelligence** | [`components/06-threat-intelligence.md`](components/06-threat-intelligence.md) | IP reputation, ASN anomaly detection, and synthetic identity markers. |
| **AI Investigators** | [`components/07-ai-investigators.md`](components/07-ai-investigators.md) | Asynchronous LLM triage and investigation synthesis. |
| **Cases & Governance** | [`components/08-cases-governance.md`](components/08-cases-governance.md) | Analyst case disposition queue and dual-control change approvals. |
| **Auth & Tenancy** | [`components/09-auth-and-tenancy.md`](components/09-auth-and-tenancy.md) | Tenant boundary isolation and API key authentication. |
| **Storage & Outbox** | [`components/11-storage-persistence.md`](components/11-storage-persistence.md) | PostgreSQL ACID persistence, transactional outbox, and Redis velocity counters. |
| **Resilience** | [`components/12-resilience-circuit-breaker.md`](components/12-resilience-circuit-breaker.md) | 3-state circuit breaker, fallback heuristics, and dependency fault injection. |
| **Webhooks** | [`components/13-webhooks.md`](components/13-webhooks.md) | HMAC-SHA256 event dispatch with retry exponential backoff. |
| **Observability** | [`components/14-observability-slo.md`](components/14-observability-slo.md) | Prometheus metrics, structured audit events, and latency SLO tracking. |

---

## 🏛️ Architecture & System Design Deep Dives

- **[Decision Precedence Matrix](architecture/risk-decision-precedence.md)**: Hard guardrails vs. ML Bayes Minimum Risk authority.
- **[Security Baseline & Defenses](architecture/security-baseline.md)**: Authentication model, request validation, and cryptographic hash chain.
- **[Technical Limitations](architecture/technical-limitations.md)**: Explicit boundaries, memory trade-offs, and scaling constraints.
- **[GraphSAGE Relationship Intelligence](architecture/graphsage-relationship-intelligence.md)**: Inductive GNN architecture evaluated in shadow mode.
