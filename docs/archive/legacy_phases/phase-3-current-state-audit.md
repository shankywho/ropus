# Phase 3 — Device Identity & Ingestion Current State Audit

**Repository:** `AI Risk Manager / Ropus`  
**Execution Date:** August 21, 2026  
**Document Version:** 1.0 (Phase 3 Audit & Discovery)  
**Audit Scope:** End-to-End Dependency & Data-Flow Mapping of Device Fingerprints  

---

## 1. End-to-End Request & Data-Flow Trace

The lifecycle of a risk evaluation request from client ingestion to final ledger persistence was traced across all 13 pipeline stages (A through M):

```
[A. Client Request / Ingestion]
  │   • Ingests JSON payload: { amount, currency, payment_method, ip_address, device_fingerprint }
  │   • Source: Client browser / API gateway (e.g., FingerprintJS visitorId hash).
  ▼
[B. HTTP Handler (backend/internal/riskengine/handler.go)]
  │   • Unmarshals into EvaluateRiskRequest DTO.
  │   • ZERO validation or format sanitization performed on device_fingerprint.
  ▼
[C. Request DTO (backend/internal/riskengine/orchestrator.go)]
  │   • Extracts req.DeviceFingerprint (raw string).
  │   • Extracts tenantID (currently hardcoded/defaulted to "00000000-0000-0000-0000-000000000001").
  ▼
[D. Risk Engine Orchestrator Entry]
  │   • Initializes in-memory evalContext.
  ▼
[E. Feature Extraction & Parsing]
  │   • Extracts IP and Payment Token.
  │   • Ignores device_fingerprint for velocity calculations.
  ▼
[F. Redis Feature Lookup (backend/internal/features/velocity.go)]
  │   • Queries: {tenant_id}:velocity:ip:{ip} (1-hour sliding window)
  │   • Queries: {tenant_id}:velocity:token:{token} (24-hour sliding window)
  │   • CRITICAL GAP: NO Redis lookup for device velocity or device first-seen status.
  ▼
[G. ML Feature Construction (backend/internal/riskengine/orchestrator.go:L189-198)]
  │   • Builds MLPredictRequest payload.
  │   • Inferred heuristic: IsNewDevice = 1 IF len(device_fingerprint) < 8 || device_fingerprint == "new_device", ELSE 0.
  │   • CRITICAL SKEW: Any valid 32-character browser hash is treated as IsNewDevice = 0 (known/trusted device)!
  ▼
[H. ML Sidecar Inference (ml-service/serve.py)]
  │   • Invokes ONNX Runtime with 15 canonical features.
  │   • Applies default heuristic: device_seen_before = (0 if req.is_new_device == 1 else 1).
  │   • Yields raw model prediction p_raw.
  ▼
[I. Calibration Engine (ml-service/calibration/calibrator.py)]
  │   • Passes p_raw through BetaCalibrator to produce calibrated empirical posterior p_cal.
  │   • Computes presentation risk_score = int(round(p_cal * 100)).
  ▼
[J. Cost Policy Engine (ml-service/calibration/cost_policy.py)]
  │   • Evaluates expected loss: E[Cost(ALLOW)] vs E[Cost(REVIEW)] vs E[Cost(DECLINE)].
  │   • Selects optimal action and emits structured reason codes.
  ▼
[K. Database Persistence (PostgreSQL risk_decisions Table)]
  │   • Encrypts device_fingerprint using per-tenant AES-256-GCM envelope encryption.
  │   • Stores encrypted ciphertext inside JSONB feature_snapshot and raw_payload.
  │   • CRITICAL GAP: NO dedicated devices or device_accounts relational table exists.
  ▼
[L. Transactional Outbox (PostgreSQL outbox_events Table)]
  │   • Writes encrypted decision event for asynchronous Debezium CDC publication to Redpanda.
  ▼
[M. Audit Logging (PostgreSQL audit_log Table)]
  │   • Logs administrative actor operations (rules, cases).
  │   • Decision ledger acts as immutable runtime audit log.
```

---

## 2. Device Identity Audit Findings

| # | Question | Current System Reality | Security / Risk Assessment |
| :--- | :--- | :--- | :--- |
| **1** | **Where does `device_fingerprint` currently enter?** | Ingested via JSON in `POST /v1/risk-evaluations` payload as a raw string. | Unauthenticated client payload input. |
| **2** | **Is `device_fingerprint` trusted or validated?** | **NO validation.** Accepts empty strings, arbitrary strings, HTML/script injections, or spoofed values. | 🔴 **HIGH RISK:** Attackers can pass arbitrary strings without format checks. |
| **3** | **Is it normalized / hashed?** | **NO.** Ingested as raw string without lowercase canonicalization, whitespace trimming, or SHA-256 hashing. | Risk of duplicate representations for the same physical device. |
| **4** | **Is it tenant-scoped?** | Encrypted using tenant KMS key at rest, but **not scoped** in velocity/lookup caches (since no device cache exists). | Needs strict tenant isolation in Redis and PostgreSQL. |
| **5** | **Can the same fingerprint exist across tenants?** | Yes, multiple merchants could see the same physical device, but current architecture lacks multi-tenant cross-referencing safeguards. | Requires tenant isolation to prevent data leakage. |
| **6** | **Can device identifiers be replayed or spoofed?** | **YES.** An attacker who intercepts a legitimate user's `device_fingerprint` can replay it in any payload. | System does not cross-verify browser IP/ASN, TLS cipher fingerprint, or User-Agent consistency. |
| **7** | **Can an attacker bypass novelty logic?** | **YES.** Because `is_new_device` is derived via `len(fp) < 8`, sending any valid 32-char string (e.g. `"abcdef0123456789abcdef0123456789"`) causes the system to mark it as a known/trusted device (`is_new_device = 0`). | 🔴 **CRITICAL VULNERABILITY:** Completely disables new-device takeover fraud rules. |
| **8** | **Does the API distinguish device states?** | **NO.** The API treats missing, first-seen, known, changed, and spoofed fingerprints identically. | Blind to device lifecycle state transitions. |

---

## 3. Concrete Architectural Weaknesses Identified

1. **Absence of a Persistent Device Registry:**
   PostgreSQL contains no `devices` or `device_accounts` tables. Once a transaction finishes, the device fingerprint is buried inside an encrypted JSONB string, preventing fast indexed historical lookups.
2. **Zero Real-Time Device Velocity:**
   Redis sliding-window velocity only tracks IP (`{tenant}:velocity:ip:{ip}`) and Token (`{tenant}:velocity:token:{token}`). An attacker rotating IPs and card tokens on a single device encounters zero velocity friction.
3. **Absence of Shared-Device & Account Graph Linkage:**
   The system cannot detect when a single device is used to log into or execute payments across 50 different accounts (credential stuffing / card testing botnets).
4. **No Device Reputation Tracking:**
   Previous chargebacks or confirmed fraud events are never attributed back to the offending hardware fingerprint.
