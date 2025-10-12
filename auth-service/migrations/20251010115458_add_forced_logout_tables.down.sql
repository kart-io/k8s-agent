-- Rollback migration: Remove forced logout tables
-- Version: 001
-- Date: 2025-10-10
-- Updated for MySQL

DROP TABLE IF EXISTS forced_logout_notifications;
DROP TABLE IF EXISTS forced_logout_events;
