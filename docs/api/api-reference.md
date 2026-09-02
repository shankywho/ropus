# API reference

**Local base URL:** `http://localhost:8080`
**Versioned prefix:** `/v1`

This reference covers the principal routes registered by the current Go API. It is a local-development contract, not a public hosted API promise. The authoritative implementation is in [`backend/cmd/api/main.go`](../../backend/cmd/api/main.go) and its handlers.

## Conventions

- Send JSON bodies with `Content-Type: application/json`.
- `X-Tenant-ID` selects the local/demo tenant. Omit it to use the default tenant.
- `X-Idempotency-Key` is supported for `POST`, `PUT`, `PATCH`, and `DELETE` requests under `/v1`. A replayed response includes `X-Idempotency-Replayed: true`.
- Administrative mutations require `X-Admin-API-Key` (or `Authorization: Bearer <admin key>`).
- `X-Actor-ID` identifies the actor for rule and case workflows in the local environment.

## Health and status

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/health`, `/healthz` | API process health. |
| `GET` | `/readiness`, `/readyz` | Dependency readiness. |
| `GET` | `/v1/system/status` | Consolidated health, model, canary, drift, and dependency snapshot. |
| `GET` | `/v1/operations/health` | Operations health report. |
| `GET` | `/v1/operations/slo` | SLO report. |
| `GET` | `/v1/operations/metrics` | Metrics snapshot. |
| `GET` | `/v1/audit/verify` | SHA-256 audit-chain verification status. |

## Risk evaluation

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/v1/risk-evaluations` | Canonical synchronous risk evaluation. |
| `POST` | `/v1/risk/evaluate` | Alias for the canonical evaluation endpoint. |
| `GET` | `/v1/graph` | Entity graph export; accepts `decisionId` and `rootId` query parameters. |

See the [API quickstart](quickstart.md) for the request and response schema. The service returns `200 OK` for a completed evaluation, `400 Bad Request` for malformed input, `429 Too Many Requests` when a tenant’s quota is exceeded, and `500 Internal Server Error` when the pipeline cannot complete.

## Rules and cases

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/v1/rules/` | Create a rule. Body requires `name` and `dsl_ast`. |
| `GET` | `/v1/rules/` | List rules; optional `status` query filter. |
| `GET` | `/v1/rules/{id}` | Retrieve a rule. |
| `PUT` | `/v1/rules/{id}` | Update a rule. |
| `PUT` | `/v1/rules/{id}/status` | Transition a rule lifecycle state. |
| `GET` | `/v1/cases/` | List cases; optional `status` query filter. |
| `GET` | `/v1/cases/{id}` | Retrieve a case. |
| `PUT` | `/v1/cases/{id}/claim` | Claim a case for the current actor. |
| `PUT` | `/v1/cases/{id}/resolve` | Resolve a case; body requires `action` and `reason`. |

Rule approval is maker-checker controlled: the creator cannot approve their own rule. See the [rules component guide](../components/03-rules-engine.md) for the supported AST structure and lifecycle behavior.

## Models, drift, and retraining

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/v1/models/status` | Static model subsystem status. |
| `GET` | `/v1/models/registry` | Registered models. |
| `GET` | `/v1/models/production` | Production and fallback model information. |
| `GET` | `/v1/models/candidates` | Candidate model inventory. |
| `GET` | `/v1/drift/status` | Current drift state. |
| `GET` | `/v1/drift/history` | Drift measurements. |
| `POST` | `/v1/drift/evaluate` | Trigger an on-demand drift measurement. |
| `GET` | `/v1/retraining/status` | Retraining subsystem status. |
| `GET` | `/v1/retraining/history` | Retraining history. |
| `POST` | `/v1/retraining/trigger` | Trigger retraining; admin key required, with `reason` required. |
| `POST` | `/v1/retraining/jobs/{id}/cancel` | Cancel a retraining job; admin key required. |

## Operations and controlled mutations

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/v1/canary/status` | Canary routing and safety-gate status. |
| `POST` | `/v1/canary/control` | Change canary settings; admin key required. Body requires `percentage` and non-empty `reason`; optional `enabled`. |
| `POST` | `/v1/chaos/drill` | Run a local resilience drill. |
| `POST` | `/admin/chaos/drill` | Admin-protected chaos drill alias. |
| `GET` | `/v1/operations/incidents` | Current incidents. |
| `POST` | `/v1/operations/maintenance/enable` | Enable maintenance mode; admin key required. |
| `POST` | `/v1/operations/maintenance/disable` | Disable maintenance mode; admin key required. |
| `POST` | `/v1/operations/recovery/trigger` | Trigger recovery handling; admin key required. |

Use these endpoints only in a controlled local or authorized test environment. Review [incident response](../operations/incident-response.md) and [technical limitations](../architecture/technical-limitations.md) before use.

## Inbound webhooks

| Method | Path | Description |
| --- | --- | --- |
| `POST` | `/webhooks/provider` | Provider webhook ingestion. |
| `POST` | `/webhooks/razorpay` | Razorpay webhook ingestion. |
| `POST` | `/v1/webhooks/provider` | Versioned provider webhook alias. |
| `POST` | `/v1/webhooks/razorpay` | Versioned Razorpay webhook alias. |
| `POST` | `/v1/razorpay/webhook` | Razorpay webhook alias. |

Webhook authentication and expected event shapes are implemented by the ingestion handlers. Do not expose any endpoint until its secret validation and routing have been configured for the relevant provider.
