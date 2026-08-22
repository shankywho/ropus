# ROPUS Enterprise Customer Onboarding Guide

Welcome to **ROPUS AI Risk Manager**. This guide walks through setting up your organization, inviting your security team, and integrating real-time fraud decisioning in under 15 minutes.

---

## 7-Step Quickstart Checklist

### Step 1: Create Your Organization
1. Log in to the **ROPUS Admin Console** at `https://app.ropus.ai`.
2. Select your industry vertical (**Digital Banking**, **Global Marketplace**, or **Payment Gateway**).
3. Choose your subscription plan (**Growth** or **Enterprise Dedicated**).

---

### Step 2: Invite Your Risk Team & Assign Roles
Navigate to `/team` and invite your engineers and fraud analysts:
- **`OWNER`**: Full billing, API key generation, and organization deletion rights.
- **`ADMIN`**: Policy tuning, custom ML model selection, and team invites.
- **`ANALYST`**: Case review queue, graph investigation, and manual overrides.
- **`VIEWER`**: Read-only dashboard telemetry and compliance reporting.

---

### Step 3: Generate Production API Keys
Navigate to `/api-keys`:
1. Click **"Create Secret Key"**.
2. Label your key (e.g., `Production Checkout Service`).
3. Copy your `rop_live_...` secret token immediately (stored as a one-way SHA-256 hash).

---

### Step 4: Send Your First Transaction Evaluation
Use our Python or Node.js SDK:

```python
from ropus_client import RopusClient

client = RopusClient(api_key="rop_live_your_secret_key_here", base_url="https://api.ropus.ai")

transaction = {
    "customer_id": "usr_john_doe_99",
    "transaction_id": "tx_order_1001",
    "amount": 250.00,
    "currency": "USD",
    "merchant": "AcmeElectronics",
    "device_id": "dev_macbook_pro_16",
}

decision = client.evaluate_risk(transaction)
print(f"Risk Verdict: {decision['decision']} (Score: {decision['risk_score']}%)")
```

---

### Step 5: Configure Custom Risk Thresholds
Navigate to `/settings`:
- Set **Hard Block Threshold** (default: $\ge 80\%$).
- Set **Manual Review / Step-Up MFA Threshold** (default: $\ge 30\%$).
- Register your webhook URL to receive instant HMAC-SHA256 signed event dispatches.

---

### Step 6: Enable Autonomous AI Investigation Agents
- Turn on **Autonomous Agent Council**.
- Borderline decisions will automatically spawn multi-agent investigations, evidentiary graph searches, and pre-compiled regulatory SAR packages.

---

### Step 7: Go Live & Monitor Real-Time Telemetry
Navigate to `/` (Overview) or `/transactions` to watch live transaction evaluations, factor attributions, and prevented fraud metrics.
