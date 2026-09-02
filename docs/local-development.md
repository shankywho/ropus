# Local development guide

This guide runs the repository on a developer machine. The provided configuration is intentionally local-only: `.env.example` contains sample database credentials, and the backend has a development default for its administrative key. Neither belongs in a shared or production environment.

## Prerequisites

- Docker Desktop with Compose v2 for the complete stack.
- Go 1.22+ for running or testing the backend directly.
- Python 3.11+ for the ML service and its tests.
- Node.js 20+ or Bun for the frontend.

## Run the complete stack

```bash
cp .env.example .env
docker compose up --build
```

Compose starts PostgreSQL, Redis, Redpanda, Debezium, ClickHouse, the ML service, the Go backend, and the frontend. It exposes the frontend on port `3000`, the backend on `8080`, and the ML service on `8000`.

Verify the primary services:

```bash
curl http://localhost:8080/health
curl http://localhost:8080/readiness
curl http://localhost:8000/health
```

`/health` checks that the API process is serving. `/readiness` also reports dependency readiness and can remain degraded while services are still starting.

Use `docker compose down` to stop containers. Add `-v` only if you explicitly want to delete the local database and service volumes.

## Run services outside Docker

Start the data dependencies first:

```bash
docker compose up -d postgres redis redpanda debezium clickhouse
```

Then open three terminals from the repository root.

```bash
# Terminal 1 — ML sidecar
cd ml-service
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
uvicorn serve:app --host 0.0.0.0 --port 8000
```

```bash
# Terminal 2 — API
cd backend
go run ./cmd/api
```

```bash
# Terminal 3 — control plane
cd frontend
bun install
bun run dev
```

The backend does not automatically load `.env`. When running it outside Compose, export the needed variables in your shell (or use the built-in local defaults). At minimum, set `ML_SERVICE_URL=http://localhost:8000` if you changed the ML port.

## Key configuration

| Variable | Default | Used by |
| --- | --- | --- |
| `PORT` | `8080` | Go API listener |
| `DATABASE_URL` | derived from `POSTGRES_*` | PostgreSQL connection |
| `REDIS_URL` | derived from `REDIS_HOST` / `REDIS_PORT` | Redis connection |
| `ML_SERVICE_URL` | `http://localhost:8000` | ML sidecar client |
| `ADMIN_API_KEY` | local sample value | administrative endpoints |
| `RISK_MODEL_CANARY_ENABLED` | `false` | canary router |
| `RISK_MODEL_CANARY_PERCENT` | `0` | candidate traffic percentage |

For the complete set, see [`loadConfig`](../backend/cmd/api/main.go) and [`.env.example`](../.env.example). Keep `RISK_MODEL_CANARY_PERCENT=0` unless you are deliberately exercising the local canary controls.

## Verify a decision

Use the request in the [API quickstart](api/quickstart.md). A minimal health and audit check is:

```bash
curl http://localhost:8080/v1/system/status
curl http://localhost:8080/v1/audit/verify
```

## Troubleshooting

| Symptom | What to check |
| --- | --- |
| Backend is up but evaluations are degraded | Confirm `http://localhost:8000/health` responds and that `ML_SERVICE_URL` points to it. |
| Backend cannot connect to PostgreSQL or Redis | Start the dependency containers, then verify `DATABASE_URL` / `REDIS_URL` or their component variables. |
| Frontend cannot reach the API in local development | The Vite proxy uses `BACKEND_INTERNAL_URL`, which defaults to `http://localhost:8080`. |
| A repeated POST returns a conflict | Reuse an idempotency key only with the same endpoint and payload; generate a new key for a different request. |
| A protected operation returns `401` | Supply `X-Admin-API-Key` with the value configured in `ADMIN_API_KEY`. |

## Test and build

```bash
(cd backend && go test ./...)
(cd ml-service && python3 -m pytest tests/ -q)
(cd frontend && npm run lint && npm run build)
```
