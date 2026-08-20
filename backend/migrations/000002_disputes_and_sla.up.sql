-- 000002_disputes_and_sla.up.sql
-- Add SLA support to cases and create disputes table

-- 1. Add SLA column to cases table
ALTER TABLE cases ADD COLUMN IF NOT EXISTS sla_expires_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_cases_sla ON cases(tenant_id, sla_expires_at);

-- 2. Disputes Table (Evidence Packet & Dispute Correlation)
CREATE TABLE IF NOT EXISTS disputes (
    dispute_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    transaction_id VARCHAR(255) NOT NULL,
    decision_id UUID REFERENCES risk_decisions(decision_id) ON DELETE SET NULL,
    amount BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL,
    dispute_reason VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'OPEN', -- OPEN, EVIDENCE_ATTACHED, UNDER_REVIEW, WON, LOST
    evidence_packet JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_disputes_txn ON disputes(tenant_id, transaction_id);
CREATE INDEX IF NOT EXISTS idx_disputes_status ON disputes(tenant_id, status);
