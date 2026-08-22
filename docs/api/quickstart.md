# ROPUS Quickstart Guide

Integrate real-time fraud risk decisioning into your application in under 15 minutes.

---

## 1. Authentication
All API requests require a secret API key passed in the `Authorization` header:

```http
Authorization: Bearer rop_live_8a19bc7f2e41...
```

---

## 2. Evaluate a Transaction (`POST /v1/risk/evaluate`)

### Request
```json
{
  "transaction_id": "tx_order_88419",
  "customer_id": "usr_sarah_connor",
  "amount": 14500.00,
  "currency": "USD",
  "merchant_id": "CryptoLiquidityExpress",
  "device_id": "dev_mule_cluster_99",
  "ip_address": "198.51.100.44",
  "country": "CY"
}
```

### Response
```json
{
  "request_id": "req_4f8a1e9c",
  "decision_id": "dec_8f4a1e9c",
  "transaction_id": "tx_order_88419",
  "verdict": "BLOCK",
  "risk_score": 0.96,
  "confidence": 0.94,
  "recommendation": "BLOCK_AND_REVIEW",
  "reasons": [
    "Cross-border impossible travel from high-risk jurisdiction (CY)",
    "Hardware fingerprint matches known emulator / spoofing framework",
    "IP address originates from commercial bulletproof proxy / VPN",
    "Entity linked to multi-account synthetic fraud cluster (degree: 14)"
  ],
  "risk_factors": [
    { "factor_name": "Transaction Velocity / Amount", "contribution": 0.22 },
    { "factor_name": "Impossible Travel Anomaly", "contribution": 0.21 },
    { "factor_name": "Device Novelty / Emulator", "contribution": 0.18 },
    { "factor_name": "IP Reputation & Proxy", "contribution": 0.18 },
    { "factor_name": "Fraud Graph Exposure", "contribution": 0.17 },
    { "factor_name": "Real ML Gradient Boosted Model", "contribution": 0.20 }
  ],
  "case_id": "CASE-88419",
  "latency_ms": 1.42
}
```

---

## 3. Python Integration Example

```python
import requests

API_KEY = "rop_live_your_secret_key"
BASE_URL = "https://api.ropus.ai"

payload = {
    "transaction_id": "tx_order_88419",
    "customer_id": "usr_sarah_connor",
    "amount": 14500.00,
    "currency": "USD",
    "merchant_id": "CryptoLiquidityExpress",
    "device_id": "dev_mule_cluster_99",
    "ip_address": "198.51.100.44",
    "country": "CY"
}

resp = requests.post(
    f"{BASE_URL}/v1/risk/evaluate",
    headers={"Authorization": f"Bearer {API_KEY}"},
    json=payload
)

data = resp.json()
print(f"Verdict: {data['verdict']} (Score: {data['risk_score']})")

if data["verdict"] == "BLOCK":
    print(f"Settlement blocked. Case opened: {data['case_id']}")
```

---

## 4. JavaScript / Node.js Integration Example

```javascript
const axios = require('axios');

async function evaluateRisk(transaction) {
  const response = await axios.post('https://api.ropus.ai/v1/risk/evaluate', transaction, {
    headers: {
      'Authorization': 'Bearer rop_live_your_secret_key',
      'Content-Type': 'application/json'
    }
  });

  return response.data;
}
```
