# API quickstart

This guide targets a locally running ROPUS stack. Start it with `docker compose up --build`, then confirm `http://localhost:8080/health` responds before sending an evaluation.

## Evaluate a transaction

The canonical endpoint is `POST /v1/risk-evaluations`. `POST /v1/risk/evaluate` accepts the same request as a compatibility alias.

```bash
curl -X POST http://localhost:8080/v1/risk-evaluations \
  -H 'Content-Type: application/json' \
  -H 'X-Tenant-ID: demo-merchant' \
  -H 'X-Correlation-ID: checkout-req-1001' \
  -H 'X-Idempotency-Key: checkout-tx-1001' \
  -d '{
    "transaction_id": "tx_checkout_1001",
    "amount": 14999,
    "currency": "INR",
    "payment_method": {
      "type": "card",
      "token": "tok_demo_card"
    },
    "device_fingerprint": "device_demo_1001",
    "ip_address": "198.51.100.44",
    "account_id": "acct_demo_42"
  }'
```

`amount` is an integer amount in the currency’s smallest unit; for example, `14999` represents ₹149.99 when the integration uses paise. ROPUS currently does not validate currency minor-unit conventions, so the caller must use one convention consistently.

## Request contract

| Field | Required | Notes |
| --- | --- | --- |
| `transaction_id` | Yes | Non-empty string, maximum 128 characters. |
| `amount` | Yes | Integer; negative values are rejected. |
| `currency` | No | Defaults to `USD` when absent. |
| `payment_method.type` | No | Payment method type, such as `card`. |
| `payment_method.token` | No | Token or non-sensitive payment reference. Never send raw card data. |
| `device_fingerprint` | No | Device identifier used by velocity and graph signals. |
| `ip_address` | No | If omitted, the API uses the request’s remote address. |
| `account_id` | No | Account identifier used by graph and velocity signals. |

## Response contract

The precise score and evidence depend on the running model, rules, and historical state. A successful response has this shape:

```json
{
  "decision_id": "dec_...",
  "transaction_id": "tx_checkout_1001",
  "recommended_action": "ALLOW_RECOMMENDATION",
  "risk_score": 12,
  "reason_codes": ["..."],
  "feature_snapshot_ref": "...",
  "evaluated_at": "2026-09-02T12:00:00Z",
  "latency_ms": 8,
  "expected_fraud_exposure": 0.0,
  "expected_action_costs": {
    "ALLOW_RECOMMENDATION": 0.0,
    "MANUAL_REVIEW": 10.0,
    "DECLINE_RECOMMENDATION": 25.0
  },
  "economic_decision_reason": "..."
}
```

The action is a recommendation, not an instruction to execute payment movement. Possible actions include `ALLOW_RECOMMENDATION`, `MANUAL_REVIEW`, `STEP_UP_RECOMMENDATION`, and `DECLINE_RECOMMENDATION`. Optional response fields expose feature, threat, graph, and component-latency details when they are available.

## Headers and retries

| Header | Purpose |
| --- | --- |
| `Content-Type: application/json` | Required for JSON requests. |
| `X-Tenant-ID` | Local/demo tenant selector. It must match `^[a-zA-Z0-9_-]{1,64}$`; an omitted value uses the default demo tenant. |
| `X-Idempotency-Key` | Recommended for every POST. Repeating the same endpoint and payload replays the original result, marked with `X-Idempotency-Replayed: true`. |
| `X-Correlation-ID` | Optional request-tracing value; the API returns it when supplied. |

Using an idempotency key with a different payload or endpoint returns `409 Conflict`. Generate a new key for each new business operation.

## Useful follow-up calls

```bash
# Local service and dependency state
curl http://localhost:8080/v1/system/status

# SHA-256 audit-chain status
curl http://localhost:8080/v1/audit/verify

# Entity graph associated with an evaluation
curl 'http://localhost:8080/v1/graph?decisionId=DECISION_ID&rootId=acct_demo_42' \
  -H 'X-Tenant-ID: demo-merchant'
```

For the wider endpoint inventory, continue to the [API reference](api-reference.md). For setup and configuration, read the [local development guide](../local-development.md).
