# AI Risk Manager — Customer API Reference

Base URL: `https://api.ropus.ai/v1`

---

## 1. Evaluate Transaction Risk

`POST /v1/risk/evaluate`

Evaluates transaction telemetry against multi-model ML, fraud graph, and rules engines.

### Headers
- `Authorization: Bearer <API_KEY>`
- `Content-Type: application/json`

### Request Body
```json
{
  "transaction_id": "tx_8819203",
  "user_id": "usr_99812",
  "amount": 2500.00,
  "currency": "USD",
  "merchant": "LuxuryGoodsHub",
  "device": {
    "device_fingerprint": "fp_88a91c2b",
    "ip_address": "198.51.100.44",
    "user_agent": "Mozilla/5.0",
    "is_emulator": true,
    "is_vpn": true
  },
  "location": {
    "country": "CY",
    "city": "Limassol",
    "lat": 34.68,
    "lon": 33.04
  },
  "metadata": {
    "cart_items_count": 3
  }
}
```

### Response Body (`200 OK`)
```json
{
  "transaction_id": "tx_8819203",
  "risk_score": 94.0,
  "decision": "BLOCK",
  "confidence": 0.96,
  "reasons": [
    "Device telemetry indicates VPN or emulator environment",
    "Fraud knowledge graph detected connection to active syndicate ring",
    "High risk geolocation observed: CY"
  ],
  "human_explanation": "Transaction tx_8819203 blocked due to critical risk score (94.0%). High correlation with malicious fraud cluster and emulator spoofing.",
  "breakdown": {
    "graph_intelligence_weight": 0.92,
    "behavior_analysis_weight": 0.55,
    "threat_intelligence_weight": 0.95,
    "machine_learning_weight": 0.25
  },
  "model_version": "v3.34-ensemble-prod",
  "graph_signals": [
    "Entity linked to known transnational carding cluster (degree: 14)"
  ],
  "recommended_action": "Block transaction immediately and trigger account freeze review",
  "evaluated_at": "2026-08-22T12:00:00Z"
}
```

---

## 2. Case Management Endpoints

### Create Case
`POST /v1/cases/create`

### Get Case Details
`GET /v1/cases/{id}`

### Get Case Investigation Timeline
`GET /v1/cases/{id}/timeline`
