-- 000003_outbox_events.up.sql
-- Transactional Outbox Table for Debezium CDC and Asynchronous Event Streaming

CREATE TABLE IF NOT EXISTS outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type VARCHAR(255) NOT NULL,
    aggregate_id VARCHAR(255) NOT NULL,
    type VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_outbox_created_at ON outbox_events(created_at);
CREATE INDEX IF NOT EXISTS idx_outbox_aggregate ON outbox_events(aggregate_type, aggregate_id);
