# AI Risk Manager — Full End-to-End Build Plan
### From Architecture Doc → Working Demo → Internship-Ready Project

**Version:** 1.0 | **Date:** August 21, 2026
**Goal:** Turn the existing backend architecture doc into a real, running, demoable full-stack project — scoped so one person can actually finish it.

---

## 0. Read This First — How to Use This Document

Your existing PDF (`AI Risk Manager Backend Architecture and Delivery Plan`) is a **14-week, multi-team, production-grade enterprise spec**. That's the vision doc — keep it, it's genuinely impressive and shows systems-design maturity. But nobody builds all of it solo for an internship application.

This document does three things:

1. **Scopes** the architecture into a **4-week MVP** you can actually build and demo, plus optional "stretch" layers if you have more time.
2. **Maps every architectural component to a specific open-source tool or GitHub repo** so you're never starting from a blank file.
3. **Gives you the frontend** (which the original doc explicitly excludes — "Host or render frontend UI screens" is out of backend scope, but you need one to demo).

Use this as the source you feed back into an AI assistant (or your own notes) to generate detailed sub-documents per component — e.g. "generate the Postgres schema.sql from section 4," "generate the Go project skeleton from section 3."

---

## 1. Project Framing for Interviews

When you present this, frame it as:

> "I designed and partially implemented a real-time fraud/risk decisioning platform inspired by how payment companies like Stripe Radar and Razorpay Thirdwatch work. I built a working MVP covering the synchronous risk-decision API, a rules engine, a feature store, and an analyst dashboard, and I documented the full production architecture — including multi-tenancy, ML model governance, and RBI/PCI-DSS compliance considerations — as the roadmap for scaling it."

This is a strong story because it shows: **systems thinking beyond the code**, **awareness of real-world constraints** (compliance, latency budgets, failure modes), and a **working artifact**, not just a spec.

---

## 2. Scoped MVP (What You Will Actually Build in ~4 Weeks)

Cut ruthlessly. The MVP proves the *hard* parts of the architecture work, not every feature.

| Original Component | MVP Version | Cut for Now |
|---|---|---|
| Go microservices (6 services) | **1 Go monolith**, cleanly modularized into packages that mirror the services | Service split, gRPC between them |
| Kafka event bus | **Postgres LISTEN/NOTIFY or a simple outbox table + polling** | Real Kafka cluster (add later if time allows — Redpanda is a 1-command Docker alternative) |
| ClickHouse analytics | Skip — use Postgres for everything, mark as "Phase 3" in docs | ClickHouse |
| ML model (ONNX + drift detection) | **One real, small ML model** (XGBoost/LightGBM on a public fraud dataset), served via a Python FastAPI sidecar | Retraining pipelines, drift monitors, SHAP in production |
| Multi-tenancy | Single-tenant, but **schema designed for multi-tenant** (tenant_id everywhere) | Per-tenant KMS keys, RLS enforcement testing |
| Rules engine (JSON AST DSL) | **Fully build this** — it's the most interview-impressive, self-contained piece | Turing-complete anything (correctly excluded either way) |
| Maker-checker workflow | **Fully build this** — simple but shows real engineering judgment | Full audit-log immutability guarantees |
| Case management UI | **Fully build this** — this is your demo centerpiece | Polished UX, real-time collaboration |
| PCI tokenization proxy | **Simulate it** — mock endpoint that swaps a fake PAN for a token; document that a real PCI-validated proxy (e.g., Basis Theory, Skyflow) would sit here in production | Actual PCI compliance work |
| RBI data localization | Document the requirement, deploy to a single region, note `ap-south-1` as the production target | Actual multi-region infra |

**Your MVP pitch, concretely, is:**
A merchant sends a transaction → hits a Go API → rules engine evaluates → ML model scores it → decision + reason codes returned in <150ms → risky ones create a case → an analyst reviews it in a dashboard → their decision is logged immutably.

That loop, working end to end, is enough to get you hired.

---

## 3. Full Stack Selection (with Open Source / GitHub Anchors)

### 3.1 Backend Core

| Layer | Tool | Why | Repo |
|---|---|---|---|
| API/Orchestrator | **Go + Chi or Gin** router | Matches doc's stack choice; fast to learn | github.com/go-chi/chi, github.com/gin-gonic/gin |
| Rules Engine DSL | Hand-rolled JSON-AST evaluator, OR **Google CEL-Go** if you want a battle-tested expression engine | CEL is Google's own "safe expression language," exactly the "no Turing-complete execution" requirement the doc calls for | github.com/google/cel-go |
| Relational DB | **PostgreSQL 16** | ACID, RLS, JSONB — matches doc exactly | postgres.org / use Docker image `postgres:16` |
| DB migrations | **golang-migrate** or **Atlas** | Clean, versioned schema | github.com/golang-migrate/migrate, github.com/ariga/atlas |
| ORM/Query builder | **sqlc** (generates type-safe Go from SQL) or **pgx** raw | Idiomatic Go, avoids ORM magic, looks good in code review | github.com/sqlc-dev/sqlc, github.com/jackc/pgx |
| Cache / Velocity features | **Redis 7** (sorted sets for velocity windows) | Exactly as specced | redis.io, use `redis:7` Docker image |
| Event bus (MVP) | **Outbox table + goroutine poller**; upgrade path: **Redpanda** (Kafka-API-compatible, 1-line Docker run, no ZooKeeper) | Redpanda gives you "real Kafka" experience without JVM/ZK pain | github.com/redpanda-data/redpanda |
| Event bus (stretch) | **Apache Kafka** via `confluentinc/cp-kafka` Docker images, or **Debezium** for real CDC from Postgres | Matches doc's Phase 2 exactly | github.com/debezium/debezium |
| ML serving | **Python + FastAPI + ONNX Runtime** | Matches doc's stack; ONNX makes the model language-agnostic | github.com/microsoft/onnxruntime, fastapi.tiangolo.com |
| ML training | **XGBoost or LightGBM** on a public fraud dataset | Fast to train, explainable, industry-standard for tabular fraud | github.com/dmlc/xgboost, github.com/microsoft/LightGBM |
| Fraud dataset | **IEEE-CIS Fraud Detection** (Kaggle) or **Credit Card Fraud Detection (ULB)** dataset | Realistic, well-known, public | kaggle.com/c/ieee-fraud-detection, kaggle.com/mlg-ulb/creditcardfraud |
| Explainability | **SHAP** | Matches doc's requirement directly | github.com/shap/shap |
| API Gateway (stretch) | **Kong** or **Envoy** | Matches doc; skip for MVP, use Go middleware instead | github.com/Kong/kong, github.com/envoyproxy/envoy |
| Auth (JWT/OAuth for analyst UI) | **Keycloak** (self-hosted IdP) or simpler: **golang-jwt** hand-rolled | Keycloak is overkill for MVP — hand-roll JWT auth for analysts, document Keycloak as the production path | github.com/golang-jwt/jwt, github.com/keycloak/keycloak |
| API contract | **OpenAPI 3.1** spec, generate docs with **Swagger UI / Redoc** | Professional touch, easy to demo | github.com/swaggo/swag |
| Containerization | **Docker + Docker Compose** | One `docker-compose up` should boot your whole stack — huge for demo day | docs.docker.com |
| Observability (stretch) | **OpenTelemetry + Grafana + Prometheus + Loki** | Matches doc; even a basic Grafana dashboard showing p95 latency is a great demo moment | github.com/open-telemetry, github.com/grafana/grafana |

### 3.2 Frontend (Analyst Dashboard + Case Management UI)

The backend doc explicitly excludes the frontend — so here's what to build:

| Layer | Tool | Why |
|---|---|---|
| Framework | **Next.js 15 (App Router) + React + TypeScript** | Industry standard, great for a portfolio project, SSR for fast loads |
| UI Components | **shadcn/ui** (built on Radix + Tailwind) | Looks professional out of the box, fully customizable, huge for a solo dev with limited design time |
| Styling | **Tailwind CSS** | Fast iteration |
| Data fetching / state | **TanStack Query (React Query)** | Clean async state, matches how real fintech dashboards are built |
| Tables (case queue, rule list) | **TanStack Table** | Sorting/filtering "manual review queue" like real analyst tools |
| Charts (risk score distributions, rule hit-rates) | **Recharts** or **Tremor** (dashboard-specific charting library) | Tremor is literally built for exactly this kind of ops dashboard | github.com/tremorlabs/tremor |
| Forms (rule builder) | **React Hook Form + Zod** | Type-safe rule creation form that maps directly to your JSON AST |
| Rule builder UI (visual AST editor) | **react-querybuilder** | Purpose-built for exactly your "declarative DSL / JSON AST" rules engine — huge time-saver | github.com/react-querybuilder/react-querybuilder |
| Real-time updates (new cases appearing) | **WebSockets** via Go's `gorilla/websocket`, or simplest: **polling with React Query's refetchInterval** | Start with polling, upgrade to WS if time allows | github.com/gorilla/websocket |
| Auth (frontend) | **NextAuth.js / Auth.js** | Standard, works well with your hand-rolled JWT backend | github.com/nextauthjs/next-auth |

**Screens to build (this is your whole frontend scope):**
1. **Login** (analyst)
2. **Risk Evaluation Playground** — a form to submit a test transaction and see the live decision + reason codes + latency (this is your #1 demo screen)
3. **Manual Review Queue** — table of open cases, filterable by status/risk score
4. **Case Detail View** — transaction context, feature snapshot, rule evaluations, SHAP explanation, decision buttons (Allow/Decline)
5. **Rules Management** — list of rules, create/edit via visual query builder, maker-checker approval flow (submit → pending → approve as a different user)
6. **Dashboard/Overview** — decision volume, outcome breakdown, p95 latency chart (even fake/simulated data is fine for demo)

### 3.3 Infrastructure & DevOps (Keep Minimal for MVP)

| Need | Tool |
|---|---|
| Local dev, whole stack | **Docker Compose** (Postgres, Redis, Go API, FastAPI ML service, Next.js frontend, Redpanda if used) |
| CI | **GitHub Actions** — lint, test, build on every push (free, and shows engineering discipline to reviewers) |
| Deployment (optional, for a live demo link) | **Railway** or **Render** for backend+DB, **Vercel** for Next.js frontend — both have generous free tiers and near-zero config |
| IaC (document only, don't build) | **Terraform** — write a `main.tf` stub showing intended AWS/EKS setup, even if you never `apply` it. Shows you understand the doc's Phase 4 without wasting time actually provisioning cloud infra |

---

## 4. Repository Structure

```
ai-risk-manager/
├── docs/
│   ├── architecture.md          # your original PDF content, converted
│   ├── adr/                     # Architecture Decision Records (great interview signal)
│   │   ├── 001-outbox-vs-2pc.md
│   │   ├── 002-rules-dsl-cel-vs-custom.md
│   │   └── 003-mvp-scope-cuts.md
│   └── openapi.yaml
├── backend/
│   ├── cmd/api/main.go
│   ├── internal/
│   │   ├── ingestion/
│   │   ├── riskengine/          # pre-rules -> ML -> post-rules pipeline
│   │   ├── rules/                # DSL parser + maker-checker
│   │   ├── features/             # Redis velocity features
│   │   ├── cases/
│   │   ├── audit/
│   │   └── auth/
│   ├── migrations/
│   ├── go.mod
│   └── Dockerfile
├── ml-service/
│   ├── train.py                  # trains XGBoost on Kaggle fraud dataset
│   ├── export_onnx.py
│   ├── serve.py                  # FastAPI + onnxruntime
│   ├── requirements.txt
│   └── Dockerfile
├── frontend/
│   ├── app/
│   │   ├── login/
│   │   ├── playground/
│   │   ├── cases/
│   │   ├── rules/
│   │   └── dashboard/
│   ├── components/ui/            # shadcn components
│   ├── lib/api.ts
│   └── package.json
├── docker-compose.yml
├── .github/workflows/ci.yml
└── README.md                     # THIS is what recruiters read first — invest in it
```

---

## 5. Week-by-Week Build Order (4 Weeks, Solo)

**Week 1 — Backend Core**
- Postgres schema (`tenants`, `rules`, `risk_decisions`, `cases`, `audit_log`) via migrations
- Go API skeleton: `POST /v1/risk-evaluations` returning a **hardcoded** decision (get the contract right first)
- Redis connection + one real velocity feature (e.g., `velocity.ip.1hr` using a sorted set)
- Docker Compose brings up Postgres + Redis + Go API

**Week 2 — Rules Engine + ML**
- JSON AST rule evaluator (or CEL-Go integration)
- Maker-checker state machine: `DRAFT → PENDING_APPROVAL → SHADOW → ACTIVE`
- Train XGBoost model on Kaggle fraud dataset, export to ONNX
- FastAPI service serves the model; Go orchestrator calls it via HTTP (skip gRPC for MVP, add later if time)
- Wire together: Pre-rules → ML → Post-rules pipeline, real end-to-end decision

**Week 3 — Case Management + Frontend**
- `dispute.opened`-style webhook handler + case creation on `MANUAL_REVIEW`
- Next.js app scaffolded with shadcn/ui
- Build Playground screen (submit transaction → see decision) — **this is priority #1**
- Build Case Queue + Case Detail screens

**Week 4 — Rules UI, Polish, Docs**
- Rule builder UI with react-querybuilder, wired to backend maker-checker
- Dashboard screen with basic charts
- Write the README, ADRs, record a 2-minute demo video (huge for applications — link it in your resume)
- Deploy live if possible (Railway/Render + Vercel)

**If you have extra time (stretch, in priority order):**
1. Replace outbox-polling with real Redpanda/Kafka
2. Add WebSocket live updates to the case queue
3. Add SHAP explanations to the case detail view
4. Add OpenTelemetry + a Grafana dashboard showing real p95 latency
5. Write a k6 or Locust load test proving your latency budget

---

## 6. Open Source Repos Worth Studying Directly

These aren't drop-in solutions but are the best real-world references for exactly this domain:

- **github.com/topics/fraud-detection** — general starting point, many small reference projects
- **github.com/IBM/fraud-detection-using-oracle-machine-learning** — good example of a full fraud pipeline
- **github.com/openbanking/** projects and **github.com/hyperledger** — for understanding financial-grade API patterns
- **github.com/casbin/casbin** — if you want real RBAC/ABAC for analyst permissions instead of hand-rolling
- **github.com/OpenPolicyAgent/opa** — an alternative to a custom rules DSL; Rego is a well-known policy language, worth mentioning in your docs as "considered OPA, chose custom JSON-AST for X reason" (great ADR content)
- **github.com/feast-dev/feast** — real open-source feature store; even if you don't adopt it, read its docs to describe your Redis-based feature store more precisely
- **github.com/great-expectations/great_expectations** — data quality checks, good stretch addition for the ML pipeline
- **github.com/evidentlyai/evidently** — open-source ML drift/monitoring, directly matches the doc's "adversarial drift" requirement — worth wiring in as a stretch goal, it's designed for exactly this

---

## 7. What Makes This Stand Out in an Internship Review

1. **The ADRs (Architecture Decision Records).** Recruiters skim code but *read* well-written decision docs. Three or four short ADRs (why Postgres over Mongo, why CEL over a custom DSL, why you cut Kafka for MVP) show senior-level thinking in junior-level output.
2. **The audit trail of what you *didn't* build and why.** Section 2 above ("Cut for Now") — if you paste that into your README as "Scope & Trade-offs," it turns limited time into a feature, not an excuse.
3. **One working end-to-end demo video**, even 90 seconds, linked at the top of your README. Most student projects have no video and a half-broken `npm start`.
4. **Real data, real model.** A model trained on an actual Kaggle fraud dataset with a real AUC score is miles more credible than mocked risk scores.

---

## 8. Next Steps

Feed this document back to an assistant (or use it yourself) to generate, one at a time:
1. `docs/architecture.md` — cleaned-up version of your original PDF
2. `backend/migrations/0001_init.sql` — full schema from Section 4 of your PDF
3. `backend/internal/riskengine/` — the pre-rules → ML → post-rules pipeline in Go
4. `ml-service/train.py` — full training script against the Kaggle dataset
5. `frontend/app/playground/page.tsx` — the demo centerpiece screen

Build one piece at a time, get each one actually running before moving to the next — momentum matters more than completeness right now.
