# Phase 3 — Device Security, Privacy & Threat Modeling Audit

**Document Version:** 1.0  
**Security Standard:** PCI-DSS v4.0 & GDPR / CCPA Compliance  
**Cryptographic Architecture:** AES-256-GCM Envelope Encryption & SHA-256 Multi-Tenant Salting  

---

## 1. Data Classification Matrix: What is Raw, Hashed, Encrypted, or Never Stored

To comply with data privacy regulations (GDPR Article 32, CCPA, and PCI-DSS Requirement 3), all data attributes in the AI Risk Manager are strictly categorized:

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                           ROPUS DATA CLASSIFICATION MATRIX                      │
├───────────────────┬───────────────────────────────────┬─────────────────────────┤
│ Classification    │ Data Attributes                   │ Storage & Handling Rules│
├───────────────────┼───────────────────────────────────┼─────────────────────────┤
│ 1. RAW (Cleartext)│ • amount, currency, latency_ms    │ Standard columns & JSON;│
│                   │ • transaction_id, decision_id     │ queryable & indexable.  │
│                   │ • timestamps (created_at)         │                         │
├───────────────────┼───────────────────────────────────┼─────────────────────────┤
│ 2. HASHED (SHA256)│ • device_fingerprint_hash         │ One-way SHA-256 with    │
│                   │   = SHA256(tenant_id || ':' || fp)│ tenant salt; indexed for│
│                   │ • api_key_hash                    │ fast exact matching.    │
├───────────────────┼───────────────────────────────────┼─────────────────────────┤
│ 3. ENCRYPTED      │ • raw client device_fingerprint   │ Encrypted via AES-256-  │
│    (AES-256-GCM)  │ • client ip_address               │ GCM with 12-byte random │
│                   │ • feature_snapshot, raw_payload   │ IV; per-tenant KMS key. │
├───────────────────┼───────────────────────────────────┼─────────────────────────┤
│ 4. PSEUDONYMIZED  │ • payment_token (e.g. tok_visa..) │ Tokenized reference only│
│                   │ • account_id, customer_ref        │ Non-reversible card PAN │
├───────────────────┼───────────────────────────────────┼─────────────────────────┤
│ 5. NEVER STORED   │ • Primary Account Number (PAN)    │ Strictly forbidden.     │
│    (Prohibited)   │ • CVV / CVC / CVV2 card codes     │ Rejected immediately at │
│                   │ • Cardholder PIN / Clear Password │ API boundary.           │
└───────────────────┴───────────────────────────────────┴─────────────────────────┘
```

---

## 2. Threat Modeling & Defense Mitigations

### Threat 1: Client-Side Fingerprint Spoofing
- **Attack Vector:** An attacker crafts synthetic `device_fingerprint` strings to pretend to be a trusted corporate laptop or injects SQL/XSS payloads.
- **Defense Strategy:**
  1. API Input Validation (strictly alphanumeric strings between 16 and 64 characters).
  2. Input canonicalization & hashing: `device_id = SHA256(tenant_id || ":" || raw_fingerprint)`.
  3. Feature attribution checks: high amount + novel IP from a supposed "known device" triggers geo-velocity anomalies.

### Threat 2: Fingerprint Replay & Credential Stuffing Botnets
- **Attack Vector:** Fraudsters intercept a legitimate user's visitor ID and replay it across 500 compromised user accounts.
- **Defense Strategy:**
  1. Redis Entity Linkage: tracks `device_unique_accounts_24h`.
  2. If `unique_accounts > 5` in 24 hours, immediate rule trigger appends `RULE_SIGNAL:MULTI_ACCOUNT_DEVICE_EXCEEDED` and routes to `DECLINE` / `MANUAL_REVIEW`.

### Threat 3: Cross-Tenant Data Leakage & Identity Correlation
- **Attack Vector:** Tenant A attempts to infer or query whether a specific device has transacted with Tenant B.
- **Defense Strategy:**
  1. Cryptographic Salting: `device_fingerprint_hash` is salted with the `tenant_id`. The exact same physical browser generates two completely distinct, uncorrelated hashes under Tenant A and Tenant B.
  2. Strict Database RLS / Tenant Quarantining: All queries enforce `WHERE tenant_id = :tenant_id`.

### Threat 4: Device Identifier Exposure in Application Logs & Metrics
- **Attack Vector:** Server logs or Prometheus metrics emit raw device fingerprints, leaking personal data to logging vendors.
- **Defense Strategy:**
  1. Log Sanitization: `device_fingerprint` is masked in application logs (`fp_****_4242`).
  2. Prometheus metrics only expose aggregate counts (e.g. `device.known.count`, `device.novel.count`), never raw identifiers.

---

## 3. Cryptographic Key Management & Crypto-Shredding (GDPR)

1. **Envelope Encryption:**
   - Every tenant is assigned a 256-bit AES Master Key managed via KMS (or `MockKMS` in local development).
   - Data stored in `risk_decisions.feature_snapshot` is encrypted using AES-256-GCM with a unique 12-byte random IV per record.
2. **Crypto-Shredding on Account Deletion:**
   - When a tenant offboards or a user invokes GDPR Article 17 ("Right to Erasure"), deleting the tenant KMS key cryptographically shreds all historical encrypted decision snapshots at rest in zero milliseconds, rendering all historical ciphertexts permanently unrecoverable without requiring destructive database table rewrites.
