-- ==============================================================================
-- ROPUS AI RISK MANAGER ENTERPRISE POSTGRESQL SCHEMA DDL
-- ==============================================================================

-- 1. Multi-Tenant Organizations
CREATE TABLE IF NOT EXISTS organizations (
    org_id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    plan_tier VARCHAR(64) NOT NULL DEFAULT 'ENTERPRISE',
    monthly_quota BIGINT NOT NULL DEFAULT 10000000,
    used_this_month BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 2. Transactions
CREATE TABLE IF NOT EXISTS transactions (
    id VARCHAR(128) PRIMARY KEY,
    tenant_id VARCHAR(64) NOT NULL,
    customer_hash VARCHAR(64) NOT NULL,
    merchant_id VARCHAR(128) NOT NULL,
    amount NUMERIC(18, 4) NOT NULL,
    currency VARCHAR(16) NOT NULL,
    location_hash VARCHAR(64) NOT NULL,
    device_hash VARCHAR(64) NOT NULL,
    risk_score NUMERIC(5, 2) NOT NULL,
    decision VARCHAR(32) NOT NULL,
    model_version VARCHAR(64) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transactions_tenant_created ON transactions (tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_transactions_customer_hash ON transactions (customer_hash);
CREATE INDEX IF NOT EXISTS idx_transactions_device_hash ON transactions (device_hash);
CREATE INDEX IF NOT EXISTS idx_transactions_decision ON transactions (decision);

-- 3. Fraud Cases
CREATE TABLE IF NOT EXISTS fraud_cases (
    id VARCHAR(128) PRIMARY KEY,
    transaction_id VARCHAR(128) NOT NULL REFERENCES transactions(id),
    priority VARCHAR(32) NOT NULL DEFAULT 'HIGH',
    status VARCHAR(32) NOT NULL DEFAULT 'OPEN',
    assigned_agent VARCHAR(128),
    evidence TEXT,
    resolution TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cases_status_priority ON fraud_cases (status, priority);
CREATE INDEX IF NOT EXISTS idx_cases_transaction_id ON fraud_cases (transaction_id);

-- 4. Model Registry
CREATE TABLE IF NOT EXISTS model_registry (
    id VARCHAR(128) PRIMARY KEY,
    version VARCHAR(64) NOT NULL,
    algorithm VARCHAR(64) NOT NULL,
    metrics TEXT,
    approval_status VARCHAR(32) NOT NULL DEFAULT 'PENDING',
    artifact_location VARCHAR(512) NOT NULL,
    deployed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 5. Audit Log (Cryptographic Hash-Chain)
CREATE TABLE IF NOT EXISTS audit_logs (
    id VARCHAR(128) PRIMARY KEY,
    actor VARCHAR(128) NOT NULL,
    action VARCHAR(128) NOT NULL,
    resource VARCHAR(255) NOT NULL,
    previous_hash VARCHAR(64) NOT NULL,
    current_hash VARCHAR(64) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON audit_logs (timestamp DESC);
