-- Migration: Add forced logout tables
-- Version: 001
-- Date: 2025-10-10

BEGIN;

-- Create forced_logout_events table
CREATE TABLE IF NOT EXISTS forced_logout_events (
    id BIGSERIAL PRIMARY KEY,
    event_id VARCHAR(36) UNIQUE NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    actor_type VARCHAR(20) NOT NULL,
    actor_id VARCHAR(36),
    actor_username VARCHAR(50),
    actor_ip_address VARCHAR(45),
    target_user_id VARCHAR(36) NOT NULL,
    target_username VARCHAR(50) NOT NULL,
    session_jti VARCHAR(100),
    session_count INTEGER NOT NULL,
    session_metadata JSONB,
    reason TEXT,
    logout_type VARCHAR(20) NOT NULL,
    triggered_by VARCHAR(50) NOT NULL,
    previous_hash VARCHAR(64),
    current_hash VARCHAR(64) NOT NULL,
    correlation_id VARCHAR(36),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_fle_event_id ON forced_logout_events(event_id);
CREATE INDEX idx_fle_timestamp ON forced_logout_events(timestamp DESC);
CREATE INDEX idx_fle_target_user ON forced_logout_events(target_user_id, timestamp DESC);
CREATE INDEX idx_fle_actor ON forced_logout_events(actor_id, timestamp DESC);
CREATE INDEX idx_fle_actor_type ON forced_logout_events(actor_type);
CREATE INDEX idx_fle_logout_type ON forced_logout_events(logout_type);

-- Create forced_logout_notifications table
CREATE TABLE IF NOT EXISTS forced_logout_notifications (
    id BIGSERIAL PRIMARY KEY,
    notification_id VARCHAR(36) UNIQUE NOT NULL,
    event_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    email_address VARCHAR(255) NOT NULL,
    channel VARCHAR(20) NOT NULL,
    template_name VARCHAR(100) NOT NULL,
    subject TEXT,
    body TEXT,
    variables JSONB,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    attempts INTEGER NOT NULL DEFAULT 0,
    last_attempt_at TIMESTAMP,
    sent_at TIMESTAMP,
    failed_at TIMESTAMP,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_fln_event_id ON forced_logout_notifications(event_id);
CREATE INDEX idx_fln_user_id ON forced_logout_notifications(user_id);
CREATE INDEX idx_fln_status ON forced_logout_notifications(status);
CREATE INDEX idx_fln_created_at ON forced_logout_notifications(created_at DESC);

-- Add foreign key constraint
ALTER TABLE forced_logout_notifications
    ADD CONSTRAINT fk_fln_event_id
    FOREIGN KEY (event_id)
    REFERENCES forced_logout_events(event_id)
    ON DELETE CASCADE;

-- Add comments
COMMENT ON TABLE forced_logout_events IS 'Immutable audit trail of forced logout actions';
COMMENT ON TABLE forced_logout_notifications IS 'Tracks notification delivery for forced logout events';
COMMENT ON COLUMN forced_logout_events.previous_hash IS 'SHA-256 hash of previous event for tamper detection';
COMMENT ON COLUMN forced_logout_events.current_hash IS 'SHA-256(event_id || timestamp || actor_id || target_user_id || reason || previous_hash)';
COMMENT ON COLUMN forced_logout_events.session_metadata IS 'JSON array of {jti, ip_address, device_name, login_at} for each terminated session';

COMMIT;
