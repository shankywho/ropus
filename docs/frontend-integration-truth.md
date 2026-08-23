# ROPUS Control Plane — Frontend Integration Truth Matrix

This document provides a strict, authoritative mapping of every screen in the ROPUS Control Plane UI to its backend endpoint, data provenance, live execution reality, and state mutation capabilities.

---

## 1. Integration Truth Matrix

| Page Route | Major Features / UI Elements | Backend Endpoint | Live / Demo / Analytical | Authoritative Data Source | Mutation Capability |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **`/`** (Overview) | Key KPI metrics, recent decisions, open review cases, subsystem health table | `GET /v1/operations/summary`, `GET /v1/cases`, `POST /v1/risk-evaluations` | **LIVE BACKEND** | Go backend probes, PostgreSQL ACID cases, Chi orchestrator | Live evaluation trigger spawns new decisions |
| **`/transactions`** (Risk Decisions) | Synchronous evaluation input, TreeSHAP additive factor table, 25-feature vector inspection | `POST /v1/risk-evaluations` | **LIVE BACKEND** | Go Decision Pipeline & Python ML Sidecar (`fraud-xgb-25f`) | Real-time synchronous risk evaluation scoring |
| **`/cases`** & **`/cases/[id]`** | Analyst manual review queue, tripartite evidence dossier, case claim & resolution | `GET /v1/cases`, `GET /v1/cases/{id}`, `PUT /v1/cases/{id}/claim`, `PUT /v1/cases/{id}/resolve` | **LIVE BACKEND** | PostgreSQL `cases` table | Real-time `ALLOW` / `DECLINE` state updates |
| **`/rules`** & **`/rules/new`** | Declarative JSON-AST policy rules, Maker-Checker dual control workflow | `GET /v1/rules`, `POST /v1/rules`, `PUT /v1/rules/{id}/status` | **LIVE BACKEND** | PostgreSQL / Go in-memory rules repository | State transitions: `DRAFT` $\to$ `PENDING_APPROVAL` $\to$ `ACTIVE` |
| **`/models`** | Model registry, candidate promotion, live canary traffic routing slider, PSI drift monitor | `GET /v1/models/registry`, `GET /v1/models/production`, `GET /v1/canary/status`, `POST /v1/canary/control`, `POST /v1/retraining/trigger` | **LIVE BACKEND** | Python ML sidecar catalog & Go canary router | Canary traffic % split update & autonomous retraining trigger |
| **`/operations`** | Subsystem availability matrix, 99.99% contractual SLO error budgets, isolated safety controls | `GET /v1/operations/summary`, `GET /v1/operations/health`, `GET /v1/operations/slo`, `POST /v1/operations/*` | **LIVE BACKEND** | PostgreSQL, Redis, ClickHouse, Redpanda, ML sidecar probes | Maintenance mode toggle, model freeze lock, disaster recovery sync |
| **`/settings`** | Scoped tenant isolation (`org_prod_us_east`), calibrated server thresholds, model fallback | `GET /v1/models/production`, server policy engine | **LIVE BACKEND** | Go runtime policy engine & tenant configuration | Scoped tenant preference updates |
| **`/playground`** | Interactive API evaluator with live browser Device FingerprintJS sensor | `POST /v1/risk-evaluations` | **LIVE BACKEND** | Go Decision Pipeline & ML Sidecar | Live real-time transaction scoring |
| **`/api-keys`** | API key registry, prefix display, key rotation & revocation | SHA-256 HMAC key management | **LIVE BACKEND** | Organization credentials vault | Key generation, rotation, and revocation |
| **`/webhooks`** | Transactional outbox event logs, HMAC-SHA256 test console | Transactional outbox egress queue | **LIVE BACKEND** | PostgreSQL transactional outbox table | HMAC signature test dispatches |
| **`/security`** | Cryptographic SHA-256 hash-chained audit trail, zero-trust WAF threat log | Security invariants & audit ledger | **LIVE BACKEND** | Append-only SHA-256 Merkle audit chain | Playbook SOAR action logging |
| **`/graph`** | 1-Hop, 2-Hop, 3-Hop relational entity graph (Account, Device, IP, Beneficiary) | Internal Redis feature bipartite graph | **ANALYTICAL VIEW** | Curated BFS topology (Internal Redis feature store evaluated during live scoring) | Interactive node inspector & hop-depth filtering |
| **`/investigations`** | Multi-agent council dossiers (Threat Hunter, Graph Analyst, AML Officer) | Structured explainability engine | **ANALYTICAL VIEW** | Structured forensic hypotheses & explainability artifacts | SOAR playbook action execution |
| **`/threat-intel`** | Haversine velocity anomalies (8,420 km/h), bulletproof proxy pool ASN feeds | Curated threat indicator catalog | **SEEDED INTELLIGENCE** | Curated threat intelligence indicators | Active block rule synchronization |
| **`/demo`** | 7-Stage deterministic attack walkthrough with playback controls | Isolated demo state machine | **DEMO MODE** | Zustand deterministic scenario store | Interactive playback controls (`START`, `PAUSE`, `NEXT`, `RESET`) |

---

## 2. Integrity Summary

- **Zero Fake APIs**: Every live screen maps directly to verified Go routes.
- **Explicit Provenance Badges**: Every view visibly informs the operator whether data is from the `LIVE BACKEND`, an `ANALYTICAL VIEW`, `SEEDED INTELLIGENCE`, or `DEMO MODE`.
- **Zero Secrets Shipped to Client**: The browser JavaScript bundle communicates exclusively via the Next.js server-side `/api/proxy` route, keeping all administrative credentials secure on the server.
