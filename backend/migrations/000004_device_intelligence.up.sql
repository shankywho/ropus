-- 000004_device_intelligence.up.sql
-- PostgreSQL Schema for Durable Relational Device Intelligence, Entity Linkage, Event Ledger, and Reputation

-- 1. Primary Device Registry
CREATE TABLE IF NOT EXISTS devices (
    device_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    device_hash VARCHAR(64) NOT NULL, -- Canonical tenant-isolated SHA-256 identity
    encrypted_fingerprint TEXT,        -- Encrypted canonical fingerprint (AES-256-GCM)
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    total_tx_count BIGINT NOT NULL DEFAULT 1,
    unique_account_count INT NOT NULL DEFAULT 1,
    trust_score INT NOT NULL DEFAULT 50, -- 0 (High Risk / Fraud) to 100 (Trusted Baseline)
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_tenant_device_hash UNIQUE (tenant_id, device_hash)
);

CREATE INDEX IF NOT EXISTS idx_devices_tenant_hash ON devices(tenant_id, device_hash);
CREATE INDEX IF NOT EXISTS idx_devices_tenant_last_seen ON devices(tenant_id, last_seen_at DESC);

-- 2. Device ↔ Account Graph Linkage
CREATE TABLE IF NOT EXISTS device_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    device_id UUID NOT NULL REFERENCES devices(device_id) ON DELETE CASCADE,
    account_id VARCHAR(255) NOT NULL,
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    transaction_count BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_tenant_device_account UNIQUE (tenant_id, device_id, account_id)
);

CREATE INDEX IF NOT EXISTS idx_dev_acc_device ON device_accounts(tenant_id, device_id);
CREATE INDEX IF NOT EXISTS idx_dev_acc_account ON device_accounts(tenant_id, account_id);
CREATE INDEX IF NOT EXISTS idx_dev_acc_last_seen ON device_accounts(tenant_id, last_seen_at DESC);

-- 3. Device ↔ Payment Token Linkage
CREATE TABLE IF NOT EXISTS device_payment_instruments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    device_id UUID NOT NULL REFERENCES devices(device_id) ON DELETE CASCADE,
    payment_token VARCHAR(255) NOT NULL, -- Safe tokenized representation (tok_...)
    first_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    transaction_count BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_tenant_device_token UNIQUE (tenant_id, device_id, payment_token)
);

CREATE INDEX IF NOT EXISTS idx_dev_token_device ON device_payment_instruments(tenant_id, device_id);
CREATE INDEX IF NOT EXISTS idx_dev_token_token ON device_payment_instruments(tenant_id, payment_token);

-- 4. Durable Historical Device Event Ledger
CREATE TABLE IF NOT EXISTS device_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    device_id UUID NOT NULL REFERENCES devices(device_id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL, -- TRANSACTION, LOGIN, ACCOUNT_LINK, PAYMENT_LINK, CHARGEBACK, FRAUD_CONFIRMATION
    account_id VARCHAR(255),
    payment_token VARCHAR(255),
    event_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    amount BIGINT,
    currency VARCHAR(3),
    risk_decision_id UUID REFERENCES risk_decisions(decision_id) ON DELETE SET NULL,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_dev_events_timeline ON device_events(tenant_id, device_id, event_time DESC);
CREATE INDEX IF NOT EXISTS idx_dev_events_type_time ON device_events(tenant_id, event_type, event_time DESC);

-- 5. Durable Device Reputation & Dispute Ledger
CREATE TABLE IF NOT EXISTS device_reputation (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    device_id UUID NOT NULL REFERENCES devices(device_id) ON DELETE CASCADE UNIQUE,
    fraud_count INT NOT NULL DEFAULT 0,
    chargeback_count INT NOT NULL DEFAULT 0,
    confirmed_legitimate_count INT NOT NULL DEFAULT 0,
    reputation_score INT NOT NULL DEFAULT 50, -- 0 (Confirmed Malicious) to 100 (Highly Trusted)
    risk_band VARCHAR(20) NOT NULL DEFAULT 'NEUTRAL', -- TRUSTED, NEUTRAL, SUSPICIOUS, BLACKLISTED
    last_fraud_at TIMESTAMPTZ,
    last_chargeback_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_dev_rep_band ON device_reputation(tenant_id, risk_band);
