# ROPUS Control Plane — Frontend & Mesh Architecture

This document describes the architectural flow from the user's browser down through the Next.js presentation and proxy layer to the Go API, domain services, and underlying storage subsystems.

---

## 1. End-to-End System Data Flow

```text
                                 OPERATOR BROWSER
                         (Next.js 16 Client-Side Bundle)
                                        │
                                        │ HTTP / JSON (No Secrets)
                                        ▼
                             NEXT.JS SERVER-SIDE PROXY
                              (/api/proxy/[...path])
                                        │
                                        │ Injects:
                                        │ - X-Admin-API-Key (Server Env)
                                        │ - X-Tenant-ID
                                        │ - X-Correlation-ID
                                        ▼
                             GO CHI ROUTER (Port 8080)
                              (backend/cmd/api/main.go)
                                        │
                   ┌────────────────────┴────────────────────┐
                   │                                         │
          SYNCHRONOUS PIPELINE                      ASYNCHRONOUS PIPELINE
                   │                                         │
        ┌──────────┴──────────┐                     ┌────────┴────────┐
        ▼                     ▼                     ▼                 ▼
   Risk Orchestrator     Rules Engine          Event Outbox     Audit Ledger
  (product_api.go)       (rules/dsl.go)       (Redpanda/Kafka)   (ClickHouse)
        │                     │                     │                 │
        ├─────────────────────┤                     ▼                 ▼
        │                     │               risk.events       OLAP Drift Log
        ▼                     ▼
   Redis Store           ML Sidecar (8000)
  (Graph/Features)      (fraud-xgb-25f ONNX)
        │
        ▼
   PostgreSQL 16
  (Cases / Rules)
```

---

## 2. Layer-by-Layer Responsibilities

### Layer 1: Presentation & Presentation State
- **Framework**: Next.js 16 (Turbopack) with App Router.
- **Design System**: Razorpay Blade aesthetic (`#070e1c` canvas, `#011638` sidebar, `#0f172a` card, `#1c2536` border, `#0d94fb` Dodger Blue brand, 4px Blade corner radius).
- **Data Provenance**: Every screen displays a standardized badge (`LIVE_BACKEND`, `ANALYTICAL_VIEW`, `SEEDED_INTELLIGENCE`, or `DEMO_MODE`).

### Layer 2: Secure Server-Side Proxy (`frontend/src/app/api/proxy/[...path]/route.ts`)
- Prevents administrative secret exposure (`ADMIN_API_KEY`) in client bundles.
- Automatically handles preflight requests, sets request timeouts, and injects tenant scoping headers (`X-Tenant-ID`).

### Layer 3: Centralized Domain Client Layer (`frontend/src/api/`)
- Domain-isolated API modules:
  - `decisions.ts`: Synchronous evaluations (`POST /v1/risk-evaluations`).
  - `cases.ts`: Case management and resolution workflows (`/v1/cases`).
  - `rules.ts`: AST policy builder and Maker-Checker transitions (`/v1/rules`).
  - `models.ts`: Model candidate registry and canary traffic controls (`/v1/models`, `/v1/canary`).
  - `operations.ts`: Cluster availability, SLOs, and isolated safety controls (`/v1/operations`).

### Layer 4: Authoritative Go Backend & Subsystems
- **Go Chi Router (`port 8080`)**: Authoritative decision engine, state machine, and orchestrator.
- **Python ML Sidecar (`port 8000`)**: ONNX runtime executing calibrated gradient boosted trees (`fraud-xgb-25f-v3.0`).
- **PostgreSQL 16 (`port 5432`)**: ACID persistence for review cases, rules, and maker-checker audit histories.
- **Redis 7 (`port 6379`)**: Hot feature cache and entity bipartite graph.
- **Redpanda / Kafka (`port 9092`)**: Asynchronous event streaming to `risk.events`.
- **ClickHouse (`port 9000`)**: Columnar storage for long-term audit trails and feature drift analytics.
