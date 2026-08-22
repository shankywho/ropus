# Phase 3 — FingerprintJS Capability, Stability & Security Analysis

**Document Version:** 1.0  
**Resource Assessed:** [FingerprintJS Open Source Library (v4+)](https://github.com/fingerprintjs/fingerprintjs)  
**Role in Ropus Architecture:** Untrusted Client Telemetry Sensor / Peripheral Signal  

---

## 1. Underlying Entropy Signals & Identification Model

Open-source FingerprintJS extracts a 32-character hexadecimal `visitorId` by collecting and hashing 30+ browser entropy sources:

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                      FINGERPRINTJS BROWSER ENTROPY SIGNALS                      │
├────────────────────────┬────────────────────────────┬───────────────────────────┤
│ Hardware & Platform    │ Graphics & Media Subsystem │ Browser Engine & Locale   │
├────────────────────────┼────────────────────────────┼───────────────────────────┤
│ • CPU Cores (concurrency) • Canvas 2D Rendering Hash  │ • User-Agent String       │
│ • Device Memory (GB)   │ • WebGL Vendor & Renderer  │ • Language & System Locale│
│ • Screen Resolution    │ • WebGL Shader Precision   │ • Timezone & DST Offset   │
│ • Color Depth & Gamut  │ • AudioContext Oscillator  │ • Touch Support / MaxPts  │
│ • Platform (Mac/Win/X) │ • Installed Fonts Metrics  │ • Session / Local Storage │
└────────────────────────┴────────────────────────────┴───────────────────────────┘
```

---

## 2. Identifier Stability & Entropy Characteristics

| Scenario | Stability Behavior | Accuracy / Entropy Expectation | Risk Assessment |
| :--- | :--- | :--- | :--- |
| **Standard Session Reload** | **100% Stable.** Identical hash produced. | High entropy (~90–95% distinct in general desktop population). | Normal expected baseline. |
| **Incognito / Private Browsing** | **Stable.** Hardware, Canvas, and WebGL entropy remain active. | Preserves hash across Incognito tabs on standard Chromium/Firefox. | Legitimate users and fraudsters can be correlated across Incognito. |
| **Browser Version Updates** | **Unstable (Hash Drifts).** Engine updates or font changes alter canvas hash. | `visitorId` mutates into a novel string for the same physical device. | Must not treat updated browser as an immediate malicious takeover. |
| **Mobile Ecosystems (iOS Safari)** | **Low Entropy / High Collision.** All iPhone 15 Pro devices running iOS 17.5 share identical Canvas, GPU, and screen specs. | Multiple distinct users share identical `visitorId`. | **NEVER** use `visitorId` as the sole deterministic identity on iOS mobile traffic! |
| **Anti-Fingerprinting Browsers** | **Randomized per session.** (Brave, Firefox Strict, Tor, CanvasBlocker). | Injects noise into Canvas/Audio rendering, producing a new hash on every page reload. | Generates high device churn; must trigger `MODEL_SIGNAL:RANDOMIZED_FINGERPRINT_ANOMALY`. |

---

## 3. Threat Model: Spoofing, Evasion & Manipulation

1. **Client-Side Parameter Tampering:**
   - A malicious actor calling `POST /v1/risk-evaluations` directly via `curl` or Postman can pass any arbitrary string (e.g. `device_fingerprint: "trusted_admin_macbook"`).
   - **Defense:** The risk engine must validate format (alphanumeric, length 16–64), normalize via SHA-256 with tenant salt, and cross-reference with IP/User-Agent telemetry.
2. **Headless Automation & Fraud Farms:**
   - Puppeteer / Playwright bots running on AWS EC2 share identical headless Linux GPU renderers (`SwiftShader / Mesa`).
   - **Defense:** Track global and tenant velocity on shared headless GPU hashes.
3. **Fingerprint Replay Across Accounts:**
   - Fraudsters using stolen credit cards replay a single clean browser hash across 100 accounts.
   - **Defense:** Implement `device_unique_accounts_24h` and `device_unique_tokens_24h` velocity counters in Redis.

---

## 4. Privacy, Compliance & Multi-Tenancy

- **Regulatory Status (GDPR / CCPA / ePrivacy):** Browser fingerprints constitute **pseudo-anonymous personal data (PII)** because they can uniquely identify a natural person's terminal equipment.
- **Storage Directives:**
  1. Raw client-provided strings must be normalized and hashed: `device_id = SHA256(tenant_id || ":" || raw_fingerprint)`.
  2. In persistent relational storage (`risk_decisions`), raw telemetry must remain encrypted at rest via per-tenant AES-256-GCM envelope keys.
  3. Strict multi-tenant isolation: a device hash observed under Tenant A must NEVER be exposed or cross-queried directly by Tenant B.

---

## 5. Architectural Role in Ropus

FingerprintJS is strictly treated as an **UNTRUSTED PERIPHERAL SIGNAL**.

```
[ UNTRUSTED INPUT ] ──► [ NORMALIZATION & SANITIZATION ] ──► [ DURABLE HISTORY & GRAPH ]
Client visitorId         SHA-256 Hash + Tenant Salt          Redis Velocities + PostgreSQL
```

Ropus remains the authoritative source of truth for:
- State persistence and first-seen timestamps
- Entity linkage (Device ↔ Account ↔ Card Token)
- Velocity counters and sliding-window anomaly detection
- Device reputation scoring and machine learning feature vectors
