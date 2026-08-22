# Security & Data-Flow Baseline Specification

**Document Version:** 1.0 (Phase -1 Baseline)  
**Evaluated Code:** `backend/internal/utils/crypto.go`, `backend/internal/utils/kms.go`, `backend/internal/riskengine/orchestrator.go`, `backend/internal/ingestion/webhook_handler.go`, `backend/internal/rules/service.go`  

---

## 1. Security Architecture & Threat Boundaries

```
[External Client / Merchant]
       │ (TLS / HTTPS)
       ▼
[Go API Server (Chi Router)] ──► Tenant ID Header ("X-Tenant-ID") / Default
       │
       ├──► [HMAC-SHA256 Verification] (Provider Webhooks: X-Signature-SHA256)
       │
       ├──► [Maker-Checker Authorization] (Rules Service: created_by != approved_by)
       │
       ├──► [AES-256-GCM Envelope Encryption] (KMS Derived 256-bit Tenant Key)
       │       │
       │       └──► [PostgreSQL 16 Storage] (Encrypted IP & Fingerprint in JSONB)
       │
       ├──► [PII Masking Filter] (192.168.1.45 ──► 192.168.***.***)
       │       │
       │       └──► [Kafka Outbox Payload] (Sanitized / Masked Event Stream)
       │
       └──► [Crypto-Shredding Engine] (Zeroes tenant key ──► irreversible data erasure)
```

---

## 2. Security Domain Audits & Current Reality

| Security Domain | Current Implementation Reality | Strengths | Vulnerabilities & Critical Gaps |
| :--- | :--- | :--- | :--- |
| **Tenant Isolation** | Multi-tenancy enforced at database query level (`tenant_id = $1`) and Redis key prefixes (`vel:ip:{tenant}:{ip}`). | Strong logical data segregation across SQL tables and cache keys. | Default tenant (`00000000-0000-0000-0000-000000000001`) is auto-created if header is omitted. No hard tenant validation against an auth service. |
| **API Authentication** | API currently permits requests without active Bearer tokens or API Key validation (stubbed default tenant fallback). | Fast developer ergonomics for testing and playground prototyping. | **CRITICAL GAP:** Anyone with network access to port 8080 can submit risk evaluation and rule mutation requests. |
| **Maker-Checker Dual Control** | Enforced in `rules/service.go:TransitionStatus()`: rule cannot be transitioned from `PENDING_APPROVAL` to `ACTIVE` if `created_by_id == approved_by_id`. | Hard mathematical block against self-approval (returns `403 Forbidden` with `maker_checker_violation`). | Frontend currently passes simulated `approved_by_id` parameter without server-side JWT verification. |
| **HMAC Webhook Ingestion** | Validates SHA-256 HMAC signatures on headers `X-Signature-SHA256`, `X-Hub-Signature-256`, or `X-Webhook-Signature` using `subtle.ConstantTimeCompare`. | Timing-attack resistant signature comparison. | Replay protection is missing: duplicate webhook event IDs are not deduplicated against a Redis nonce cache. |
| **Envelope Encryption (PII at Rest)** | AES-256-GCM encryption for `ip_address` and `device_fingerprint` inside `feature_snapshot` and `raw_payload`. | Uses 12-byte random IV per record and standard GCM authenticated encryption. | If KMS key is unavailable, falls back to storing plaintext (logged as warning). |
| **Key Management (KMS)** | `backend/internal/utils/kms.go` defines `KMS` interface with `GetTenantKey(tenantID)` and `ShredTenantKey(tenantID)`. | Clean interface decoupling storage from crypto implementation. | **MOCK IMPLEMENTATION:** `MockKMS` stores 32-byte keys in an in-memory Go map. No AWS KMS, GCP KMS, or HashiCorp Vault driver. |
| **Crypto-Shredding (DPDP/GDPR)** | `MockKMS.ShredTenantKey(tenantID)` explicitly overwrites the 32-byte key slice with zeros (`0x00`) and removes it from the map. | Instant, irreversible cryptographic erasure of all historically encrypted tenant PII. | In-memory keys are lost on container restart unless seeded. |
| **PII in Event Streams & Logs** | `maskIPAddress()` replaces the last two octets of IPv4 addresses (`192.168.***.***`) before serializing into `outbox_events`. | Decrypted raw IP addresses are never transmitted to Kafka, Debezium, or ClickHouse. | Device fingerprints in outbox are stored in encrypted format; raw tokens are present in outbox payload. |
| **SQL Injection Prevention** | All database interactions in `internal/riskengine`, `internal/rules`, and `internal/cases` use parameterized queries (`$1, $2, ...`) via `jackc/pgx/v5`. | Zero string interpolation in SQL statements. | None observed. |

---

## 3. Environment & Secret Configuration Audit

- **Environment File:** `.env` (configured with non-conflicting host ports: `5433`, `6380`, `8124`, `9001`, `8080`, `3000`).
- **Secrets Managed via Environment:**
  - `POSTGRES_PASSWORD`: `postgres`
  - `WEBHOOK_SECRET`: `whsec_test_secret_key_8821`
  - `CLICKHOUSE_PASSWORD`: `clickhouse`
- **Exposure Risk:** Default demo secrets are configured in `.env.example` and `docker-compose.yml`. Production deployment requires external secret injection (AWS Secrets Manager / Vault).
