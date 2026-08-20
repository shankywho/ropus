-- 000002_disputes_and_sla.down.sql
DROP TABLE IF EXISTS disputes;
ALTER TABLE cases DROP COLUMN IF EXISTS sla_expires_at;
