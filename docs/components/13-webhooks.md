# Component 13: Webhook Egress & Cryptographic Delivery

---

## 1. Why It Exists
When ROPUS evaluates a transaction or opens a high-priority fraud case, the customer's downstream systems (core banking ledgers, customer support tools, or automated card blockers) need to be notified in real time.

The **Webhook Subsystem** (`backend/internal/webhooks/`) provides reliable, tamper-evident event dispatching with:
1. **HMAC-SHA256 Request Signatures** to verify origin authenticity.
2. **Idempotency Keys** to prevent duplicate processing.
3. **Exponential Backoff Retries** with jitter across 5 delivery attempts.

---

## 2. Event Types & Payloads

```json
// Example: risk.decision.created
{
  "event_id": "evt_9f8a1e9c7a2b",
  "event_type": "risk.decision.created",
  "timestamp": "2026-08-22T17:42:10Z",
  "tenant_id": "org_bank_acme",
  "data": {
    "transaction_id": "tx_order_88419",
    "decision": "BLOCK",
    "risk_score": 0.96,
    "case_id": "CASE-88419"
  }
}
```

---

## 3. Cryptographic Signature Verification (Python Client)

Every outgoing HTTP POST includes the header:
`X-Ropus-Signature: sha256=d5a3...e91c`

```python
import hmac
import hashlib

def verify_ropus_webhook(payload_bytes: bytes, signature_header: str, secret: str) -> bool:
    expected_hash = hmac.new(
        secret.encode('utf-8'),
        payload_bytes,
        hashlib.sha256
    ).hexdigest()
    
    expected_header = f"sha256={expected_hash}"
    return hmac.compare_digest(expected_header, signature_header)
```

---

## 4. Exponential Backoff Retry Schedule

If the customer's endpoint returns a $5xx$ server error or network timeout, the webhook dispatcher retries according to the following schedule:

| Attempt | Delay | Max Jitter | Total Cumulative Delay |
| :---: | :---: | :---: | :---: |
| 1 | Immediate | 0s | 0s |
| 2 | 2 seconds | $\pm 0.5\text{s}$ | 2.5s |
| 3 | 8 seconds | $\pm 1.0\text{s}$ | 10.5s |
| 4 | 32 seconds | $\pm 3.0\text{s}$ | 45.5s |
| 5 | 128 seconds | $\pm 10.0\text{s}$ | ~3.0 mins |

---

## 5. Source Code Map
- [`backend/internal/webhooks/webhook_manager.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/webhooks/webhook_manager.go): Subscription registry, payload signing, and delivery loop.
- [`backend/internal/webhooks/webhook_reliability_test.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/webhooks/webhook_reliability_test.go): HMAC verification and retry test suite.

---

## 6. Cross-Component Links
- [Component 01: Product API](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/01-product-api.md) — Emits risk decision events.
- [Component 08: Case Management](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/docs/components/08-cases-governance.md) — Emits case creation and resolution events.
