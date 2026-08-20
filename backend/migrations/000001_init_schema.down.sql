-- 000001_init_schema.down.sql
-- AI Risk Manager Core PostgreSQL Schema Rollback

DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS cases;
DROP TABLE IF EXISTS risk_decisions;
DROP TABLE IF EXISTS rules;
DROP TABLE IF EXISTS tenants;
