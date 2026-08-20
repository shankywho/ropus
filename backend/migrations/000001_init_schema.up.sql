-- 000001_init_schema.up.sql
-- AI Risk Manager Core PostgreSQL Schema

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 1. Tenants Table
CREATE TABLE IF NOT EXISTS tenants (
    tenant_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    api_key_hash VARCHAR(255) NOT NULL UNIQUE,
    status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tenants_api_key_hash ON tenants(api_key_hash);
CREATE INDEX IF NOT EXISTS idx_tenants_status ON tenants(status);

-- 2. Rules Table (Declarative JSON AST DSL & Maker-Checker State Machine)
CREATE TABLE IF NOT EXISTS rules (
    rule_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    dsl_ast JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'DRAFT', -- DRAFT, PENDING_APPROVAL, SHADOW, ACTIVE, ARCHIVED
    version INT NOT NULL DEFAULT 1,
    created_by VARCHAR(255) NOT NULL,
    approved_by VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_rules_tenant_status ON rules(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_rules_tenant_name ON rules(tenant_id, name);

-- 3. Risk Decisions Table (Synchronous Decision Store)
CREATE TABLE IF NOT EXISTS risk_decisions (
    decision_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    transaction_id VARCHAR(255) NOT NULL,
    amount BIGINT NOT NULL, -- in smallest currency unit (e.g. paise / cents)
    currency VARCHAR(3) NOT NULL,
    recommended_action VARCHAR(50) NOT NULL, -- ALLOW_RECOMMENDATION, STEP_UP_RECOMMENDATION, MANUAL_REVIEW, HOLD_RECOMMENDATION, DECLINE_RECOMMENDATION, SHADOW_ONLY, INSUFFICIENT_CONTEXT
    risk_score INT NOT NULL, -- 0 - 100
    reason_codes JSONB NOT NULL DEFAULT '[]'::jsonb,
    feature_snapshot_ref VARCHAR(255),
    feature_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    raw_payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    latency_ms INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_tenant_transaction UNIQUE (tenant_id, transaction_id)
);

CREATE INDEX IF NOT EXISTS idx_risk_decisions_txn ON risk_decisions(tenant_id, transaction_id);
CREATE INDEX IF NOT EXISTS idx_risk_decisions_created_at ON risk_decisions(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_risk_decisions_action ON risk_decisions(tenant_id, recommended_action);

-- 4. Cases Table (Manual Review & Dispute Queue)
CREATE TABLE IF NOT EXISTS cases (
    case_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    decision_id UUID NOT NULL REFERENCES risk_decisions(decision_id) ON DELETE CASCADE,
    transaction_id VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'OPEN', -- OPEN, UNDER_REVIEW, RESOLVED_ALLOW, RESOLVED_DECLINE, CLOSED
    priority VARCHAR(20) NOT NULL DEFAULT 'MEDIUM', -- LOW, MEDIUM, HIGH, CRITICAL
    assigned_to VARCHAR(255),
    resolution_reason TEXT,
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cases_status_owner ON cases(tenant_id, status, assigned_to);
CREATE INDEX IF NOT EXISTS idx_cases_decision ON cases(decision_id);
CREATE INDEX IF NOT EXISTS idx_cases_txn ON cases(tenant_id, transaction_id);

-- 5. Audit Log Table (Immutable Audit Trail)
CREATE TABLE IF NOT EXISTS audit_log (
    log_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    actor_id VARCHAR(255) NOT NULL,
    action VARCHAR(100) NOT NULL, -- RULE_CREATED, RULE_APPROVED, CASE_OVERRIDE, CONFIG_CHANGE, etc.
    entity_type VARCHAR(100) NOT NULL,
    entity_id VARCHAR(255) NOT NULL,
    changes JSONB NOT NULL DEFAULT '{}'::jsonb,
    ip_address VARCHAR(45),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_log_tenant_time ON audit_log(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_log_entity ON audit_log(tenant_id, entity_type, entity_id);
