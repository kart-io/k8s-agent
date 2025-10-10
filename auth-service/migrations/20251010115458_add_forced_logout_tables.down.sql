-- Rollback migration: Remove forced logout tables
-- Version: 001
-- Date: 2025-10-10

BEGIN;

DROP TABLE IF EXISTS forced_logout_notifications CASCADE;
DROP TABLE IF EXISTS forced_logout_events CASCADE;

COMMIT;
