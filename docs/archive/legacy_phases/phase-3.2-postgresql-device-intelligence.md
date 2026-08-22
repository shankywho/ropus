# Phase 3.2 — PostgreSQL Relational Device Intelligence Persistence

**Repository:** `AI Risk Manager / Ropus`  
**Execution Date:** August 21, 2026  
**Document Version:** 1.0 (Phase 3.2 Implementation Specification)  
**Status:** COMPLETED & VERIFIED  

---

## 1. Existing Database Architecture & Integration Context

The AI Risk Manager relies on PostgreSQL 16 as its authoritative, ACID-compliant ledger. Prior to Phase 3.2, the schema consisted of:
- `tenants`, `rules`, `risk_decisions`, `cases`, `audit_log`, `disputes`, and `outbox_events`.
- Device fingerprints were stored solely as encrypted JSONB fields within `risk_decisions.feature_snapshot`.

Phase 3.2 establishes durable relational tables for device identity, multi-account linkage, payment token linkage, historical event timelines, and dispute reputation.

---

## 2. New PostgreSQL Relational Schemas (Migration `000004_device_intelligence.up.sql`)

### 1. `devices` (Primary Hardware Registry)
- `device_id UUID PRIMARY KEY DEFAULT gen_random_uuid()`
- `tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE`
- `device_hash VARCHAR(64) NOT NULL` (Canonical SHA-256 hash from Phase 3.1)
- `encrypted_fingerprint TEXT` (AES-256-GCM encrypted canonical visitor ID)
- `first_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `total_tx_count BIGINT NOT NULL DEFAULT 1`
- `unique_account_count INT NOT NULL DEFAULT 1`
- `trust_score INT NOT NULL DEFAULT 50` (0 = Confirmed Fraud to 100 = Trusted Baseline)
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- **Constraint:** `CONSTRAINT uq_tenant_device_hash UNIQUE (tenant_id, device_hash)`
- **Indexes:** `idx_devices_tenant_hash (tenant_id, device_hash)`, `idx_devices_tenant_last_seen (tenant_id, last_seen_at DESC)`

### 2. `device_accounts` (Device ↔ Account Graph Linkage)
- `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`
- `tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE`
- `device_id UUID NOT NULL REFERENCES devices(device_id) ON DELETE CASCADE`
- `account_id VARCHAR(255) NOT NULL`
- `first_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `transaction_count BIGINT NOT NULL DEFAULT 1`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- **Constraint:** `CONSTRAINT uq_tenant_device_account UNIQUE (tenant_id, device_id, account_id)`
- **Indexes:** `idx_dev_acc_device (tenant_id, device_id)`, `idx_dev_acc_account (tenant_id, account_id)`, `idx_dev_acc_last_seen (tenant_id, last_seen_at DESC)`

### 3. `device_payment_instruments` (Device ↔ Token Linkage)
- `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`
- `tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE`
- `device_id UUID NOT NULL REFERENCES devices(device_id) ON DELETE CASCADE`
- `payment_token VARCHAR(255) NOT NULL` (Tokenized card representation e.g. `tok_...`)
- `first_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `transaction_count BIGINT NOT NULL DEFAULT 1`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- **Constraint:** `CONSTRAINT uq_tenant_device_token UNIQUE (tenant_id, device_id, payment_token)`
- **Indexes:** `idx_dev_token_device (tenant_id, device_id)`, `idx_dev_token_token (tenant_id, payment_token)`

### 4. `device_events` (Historical Device Event Ledger)
- `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`
- `tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE`
- `device_id UUID NOT NULL REFERENCES devices(device_id) ON DELETE CASCADE`
- `event_type VARCHAR(50) NOT NULL` (`TRANSACTION`, `LOGIN`, `ACCOUNT_LINK`, `PAYMENT_LINK`, `CHARGEBACK`, `FRAUD_CONFIRMATION`)
- `account_id VARCHAR(255)`
- `payment_token VARCHAR(255)`
- `event_time TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `amount BIGINT`
- `currency VARCHAR(3)`
- `risk_decision_id UUID REFERENCES risk_decisions(decision_id) ON DELETE SET NULL`
- `metadata JSONB NOT NULL DEFAULT '{}'::jsonb`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- **Indexes:** `idx_dev_events_timeline (tenant_id, device_id, event_time DESC)`, `idx_dev_events_type_time (tenant_id, event_type, event_time DESC)`

### 5. `device_reputation` (Persistent Reputation & Dispute Ledger)
- `id UUID PRIMARY KEY DEFAULT gen_random_uuid()`
- `tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE`
- `device_id UUID NOT NULL REFERENCES devices(device_id) ON DELETE CASCADE UNIQUE`
- `fraud_count INT NOT NULL DEFAULT 0`
- `chargeback_count INT NOT NULL DEFAULT 0`
- `confirmed_legitimate_count INT NOT NULL DEFAULT 0`
- `reputation_score INT NOT NULL DEFAULT 50`
- `risk_band VARCHAR(20) NOT NULL DEFAULT 'NEUTRAL'` (`TRUSTED`, `NEUTRAL`, `SUSPICIOUS`, `BLACKLISTED`)
- `last_fraud_at TIMESTAMPTZ`
- `last_chargeback_at TIMESTAMPTZ`
- `created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- `updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`
- **Index:** `idx_dev_rep_band (tenant_id, risk_band)`

---

## 3. Concurrency & Upsert Architecture

To guarantee zero race conditions during high-concurrency checkouts, the device registry relies on atomic PostgreSQL `ON CONFLICT DO UPDATE` upserts backed by the unique constraint `(tenant_id, device_hash)`:

```sql
INSERT INTO devices (
    tenant_id, device_hash, encrypted_fingerprint, first_seen_at, last_seen_at,
    total_tx_count, unique_account_count, trust_score, created_at, updated_at
)
VALUES ($1, $2, $3, NOW(), NOW(), 1, 1, 50, NOW(), NOW())
ON CONFLICT (tenant_id, device_hash) DO UPDATE
SET last_seen_at = NOW(),
    total_tx_count = devices.total_tx_count + 1,
    encrypted_fingerprint = COALESCE(EXCLUDED.encrypted_fingerprint, devices.encrypted_fingerprint),
    updated_at = NOW()
RETURNING device_id;
```

---

## 4. Multi-Tenant Isolation & Privacy Safeguards

1. **Tenant Quarantining:** Every device query enforces `WHERE tenant_id = $1`. The same physical device observed under two tenants generates distinct internal hashes and distinct database rows.
2. **Zero Plaintext Storage:** Raw client fingerprints and unmasked card numbers are strictly prohibited. The database only stores the SHA-256 `device_hash`, the AES-256-GCM ciphertext, and tokenized payment references (`tok_...`).
3. **Crypto-Shredding:** Shredding a tenant's KMS key immediately renders all `encrypted_fingerprint` ciphertexts permanently unrecoverable.

---

## 5. Repository Layer Specification

Implemented in [`backend/internal/features/device_store.go`](file:///Users/shankar/PROJECTS/Ai%20Risk%20Manager/backend/internal/features/device_store.go):
- `GetDeviceByHash(ctx, tenantID, deviceHash)`
- `UpsertDeviceSeen(ctx, tenantID, deviceHash, encryptedFP)`
- `LinkAccount(ctx, tenantID, deviceUUID, accountID)`
- `LinkPaymentToken(ctx, tenantID, deviceUUID, paymentToken)`
- `RecordDeviceEvent(ctx, event)`
- `GetDeviceReputation(ctx, tenantID, deviceUUID)`

---

## 6. Migration Verification & Test Results

- **Migration Applied:** `backend/migrations/000004_device_intelligence.up.sql` (Verified on live PostgreSQL instance).
- **Rollback Tested:** `backend/migrations/000004_device_intelligence.down.sql` (Verified cleanly drops all 5 tables and re-applies without error).
- **Automated Tests:**
  - `TestDeviceStore_Integration` (7 database integration scenarios testing upserts, isolation, entity linkages, event logging, and privacy).
  - `TestParseDeviceIdentity` (19 identity canonicalization and security tests).
  - All 23 Go backend unit and integration tests passed.
  - All 17 Python ML tests passed.

---

## 7. What is Intentionally Deferred to Phase 3.3+

- **Phase 3.3:** Redis real-time sliding-window feature store (`{tenant}:vel:dev:1h`, `{tenant}:dev:acc24`, `{tenant}:dev:known:{hash}`).
- **Phase 3.4:** Account ↔ Device Graph scoring & credential stuffing anomaly algorithms.
- **Phase 3.5:** Payment Token ↔ Device Linkage card testing detection heuristics.
- **Phase 3.7:** Dynamic device reputation calculation and chargeback dispute attribution.
- **Phase 3.8 – 3.10:** ML 25-feature expansion, XGBoost retraining, and Beta calibration re-evaluation.
