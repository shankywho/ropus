# Phase 3 — Device Data Model & Relational Schema Design

**Document Version:** 1.0 (Design Specification)  
**Database Target:** PostgreSQL 16  
**Security & Multi-Tenancy:** Strict Per-Tenant Scoping, AES-256-GCM Envelope Encryption  

---

## 1. Entity-Relationship Diagram

```
       ┌────────────────────────┐
       │        tenants         │
       └───────────┬────────────┘
                   │ 1:N
                   ▼
       ┌────────────────────────┐
       │        devices         │◄───────────────────────┐
       └───────────┬────────────┘                        │
                   │ 1:N                                 │ 1:1
       ┌───────────┼─────────────────────────┐           │
       ▼           ▼                         ▼           ▼
┌──────────────┐ ┌─────────────────────────┐ ┌──────────────┐ ┌───────────────────┐
│device_accounts││device_payment_instruments││device_events │ │ device_reputation │
└──────────────┘ └─────────────────────────┘ └──────────────┘ └───────────────────┘
```

---

## 2. Proposed PostgreSQL Relational Schemas

### 1. `devices` (Durable Hardware Identity & Lifetime Metrics)
```sql
CREATE TABLE IF NOT EXISTS devices (
    device_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    device_fingerprint_hash CHAR(64) NOT NULL, -- SHA256(tenant_id || ':' || raw_fingerprint)
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    total_tx_count BIGINT NOT NULL DEFAULT 1,
    total_amount_cents BIGINT NOT NULL DEFAULT 0,
    unique_account_count INT NOT NULL DEFAULT 1,
    unique_token_count INT NOT NULL DEFAULT 1,
    is_blocked BOOLEAN NOT NULL DEFAULT FALSE,
    trust_score INT NOT NULL DEFAULT 50, -- 0 (High Risk / Fraud) to 100 (Trusted Baseline)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_tenant_device_hash UNIQUE (tenant_id, device_fingerprint_hash)
);

CREATE INDEX IF NOT EXISTS idx_devices_tenant_hash ON devices(tenant_id, device_fingerprint_hash);
CREATE INDEX IF NOT EXISTS idx_devices_last_seen ON devices(tenant_id, last_seen_at DESC);
CREATE INDEX IF NOT EXISTS idx_devices_blocked ON devices(tenant_id, is_blocked) WHERE is_blocked = TRUE;
```

### 2. `device_accounts` (Device ↔ Account Graph Linkage)
```sql
CREATE TABLE IF NOT EXISTS device_accounts (
    link_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    device_id UUID NOT NULL REFERENCES devices(device_id) ON DELETE CASCADE,
    account_id VARCHAR(255) NOT NULL,
    first_linked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_linked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    tx_count BIGINT NOT NULL DEFAULT 1,
    is_primary_device BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_tenant_device_account UNIQUE (tenant_id, device_id, account_id)
);

CREATE INDEX IF NOT EXISTS idx_dev_acc_device ON device_accounts(tenant_id, device_id);
CREATE INDEX IF NOT EXISTS idx_dev_acc_account ON device_accounts(tenant_id, account_id);
CREATE INDEX IF NOT EXISTS idx_dev_acc_last_linked ON device_accounts(tenant_id, last_linked_at DESC);
```

### 3. `device_payment_instruments` (Device ↔ Card/Token Linkage)
```sql
CREATE TABLE IF NOT EXISTS device_payment_instruments (
    link_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    device_id UUID NOT NULL REFERENCES devices(device_id) ON DELETE CASCADE,
    payment_token VARCHAR(255) NOT NULL, -- Tokenized PAN representation (tok_...)
    first_used_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    tx_count BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_tenant_device_token UNIQUE (tenant_id, device_id, payment_token)
);

CREATE INDEX IF NOT EXISTS idx_dev_token_device ON device_payment_instruments(tenant_id, device_id);
CREATE INDEX IF NOT EXISTS idx_dev_token_token ON device_payment_instruments(tenant_id, payment_token);
```

### 4. `device_events` (Historical Audit Trail & State Transitions)
```sql
CREATE TABLE IF NOT EXISTS device_events (
    event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    device_id UUID NOT NULL REFERENCES devices(device_id) ON DELETE CASCADE,
    transaction_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(50) NOT NULL, -- TRANSACTION, LOGIN, PASSWORD_RESET, CHARGEBACK_DISPUTE, MANUAL_BLOCK
    risk_score INT NOT NULL,
    action_taken VARCHAR(50) NOT NULL,
    ip_address VARCHAR(45),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_dev_events_device_time ON device_events(tenant_id, device_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_dev_events_txn ON device_events(tenant_id, transaction_id);
```

### 5. `device_reputation` (Confirmed Fraud & Chargeback Ledger)
```sql
CREATE TABLE IF NOT EXISTS device_reputation (
    reputation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    device_id UUID NOT NULL REFERENCES devices(device_id) ON DELETE CASCADE UNIQUE,
    confirmed_fraud_count INT NOT NULL DEFAULT 0,
    chargeback_count INT NOT NULL DEFAULT 0,
    last_fraud_at TIMESTAMPTZ,
    reputation_band VARCHAR(20) NOT NULL DEFAULT 'NEUTRAL', -- TRUSTED, NEUTRAL, SUSPICIOUS, BLACKLISTED
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_dev_rep_band ON device_reputation(tenant_id, reputation_band);
```

---

## 3. Multi-Tenancy & Data Retention Rules

1. **Strict Multi-Tenant Scoping:**
   Every query must include `WHERE tenant_id = :tenant_id`. Unique constraints `(tenant_id, ...)` guarantee that tenant identity namespaces remain strictly quarantined.
2. **Data Retention & Pruning Strategy:**
   - `device_events`: 90-day rolling partition via `pg_partman` or scheduled TTL cleanup.
   - `devices`, `device_accounts`, `device_reputation`: Indefinite retention with automated anonymization of inactive records ($> 730$ days) per GDPR right-to-be-forgotten directives.
