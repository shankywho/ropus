-- 000004_device_intelligence.down.sql
-- Revert device intelligence relational schema

DROP TABLE IF EXISTS device_reputation;
DROP TABLE IF EXISTS device_events;
DROP TABLE IF EXISTS device_payment_instruments;
DROP TABLE IF EXISTS device_accounts;
DROP TABLE IF EXISTS devices;
