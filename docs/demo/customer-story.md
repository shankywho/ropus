# ROPUS Customer Onboarding & Integration Story

Integrate ROPUS into your banking or checkout rails in **under 15 minutes**.

---

## 1. Single API Call Evaluation (`POST /v1/risk/evaluate`)

### Request
```bash
curl -X POST https://api.ropus.ai/v1/risk/evaluate \
  -H "Authorization: Bearer rop_live_8a19bc7f2e41..." \
  -H "Content-Type: application/json" \
  -d '{
    "transaction_id": "tx_order_88419",
    "customer_id": "usr_sarah_connor",
    "amount": 14500.00,
    "currency": "USD",
    "merchant_id": "CryptoLiquidityExpress",
    "device_id": "dev_mule_cluster_99",
    "ip_address": "198.51.100.44",
    "country": "CY"
  }'
```

### Response
```json
{
  "decision_id": "dec_8f4a1e9c7a",
  "decision": "BLOCK",
  "risk_score": 0.96,
  "confidence": 0.94,
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
    { "factor_name": "Fraud Graph Relationship", "contribution": 0.17 },
    { "factor_name": "XGBoost ML Inference", "contribution": 0.20 }
  ],
  "case_id": "CASE-88419",
  "latency_ms": 1.42
}
```

---

## 2. Python & Node.js Drop-in SDKs

```python
from ropus_client import RopusClient

client = RopusClient(api_key="rop_live_...")
decision = client.evaluate_risk({
    "transaction_id": "tx_order_88419",
    "customer_id": "usr_sarah_connor",
    "amount": 14500.00,
    "currency": "USD"
})

if decision["decision"] == "BLOCK":
    halt_settlement()
```
