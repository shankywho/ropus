# AI Risk Manager — Customer Integration Guide

This guide walks through integrating the **AI Risk Manager (Ropus)** API into checkout, authorization, and payment processing pipelines.

---

## 1. Quick Start

### Python Integration

```python
from ropus_client import RopusClient

client = RopusClient(api_key="ropus_live_abcdef1234567890", base_url="https://api.ropus.ai")

transaction = {
    "transaction_id": "tx_order_882194",
    "user_id": "usr_customer_1234",
    "amount": 250.00,
    "currency": "USD",
    "merchant": "GlobalElectronicsStore",
    "device": {
        "device_fingerprint": "fp_client_macbook_pro_16",
        "ip_address": "198.51.100.12",
        "user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
        "is_emulator": False,
        "is_vpn": False,
    },
    "location": {
        "country": "US",
        "city": "San Francisco",
    }
}

decision = client.evaluate_risk(transaction)

print(f"Verdict: {decision['decision']}") # APPROVE, REVIEW, CHALLENGE, or BLOCK
print(f"Risk Score: {decision['risk_score']}%")
print(f"Explanation: {decision['human_explanation']}")
```

---

### Node.js / JavaScript Integration

```javascript
const { RopusClient } = require("ropus-client");

const client = new RopusClient("ropus_live_abcdef1234567890", { baseUrl: "https://api.ropus.ai" });

async function processPayment(tx) {
  const verdict = await client.evaluateRisk(tx);

  if (verdict.decision === "BLOCK") {
    throw new Error(`Transaction declined: ${verdict.human_explanation}`);
  } else if (verdict.decision === "CHALLENGE") {
    return triggerBiometricStepUpMFA(tx.user_id);
  }

  return proceedToGatewaySettlement(tx);
}
```

---

## 2. Webhook Event Consumption

Ropus delivers webhooks for async decisions, case creation, and model retraining events with `X-Ropus-Signature` verification headers:

```javascript
const isValid = RopusClient.verifyWebhookSignature(rawBodyString, req.headers["x-ropus-signature"], webhookSecret);
if (!isValid) {
  return res.status(401).send("Invalid signature");
}
```
