-- Migration: Add forced logout tables
-- Version: 001
-- Date: 2025-10-10
-- Updated for MySQL

-- Create forced_logout_events table
CREATE TABLE IF NOT EXISTS forced_logout_events (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    event_id VARCHAR(36) UNIQUE NOT NULL,
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    actor_type VARCHAR(20) NOT NULL,
    actor_id VARCHAR(36),
    actor_username VARCHAR(50),
    actor_ip_address VARCHAR(45),
    target_user_id VARCHAR(36) NOT NULL,
    target_username VARCHAR(50) NOT NULL,
    session_jti VARCHAR(100),
    session_count INT NOT NULL,
    session_metadata JSON,
    reason TEXT,
    logout_type VARCHAR(20) NOT NULL,
    triggered_by VARCHAR(50) NOT NULL,
    previous_hash VARCHAR(64),
    current_hash VARCHAR(64) NOT NULL,
    correlation_id VARCHAR(36),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_fle_event_id (event_id),
    INDEX idx_fle_timestamp (timestamp DESC),
    INDEX idx_fle_target_user (target_user_id, timestamp DESC),
    INDEX idx_fle_actor (actor_id, timestamp DESC),
    INDEX idx_fle_actor_type (actor_type),
    INDEX idx_fle_logout_type (logout_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Immutable audit trail of forced logout actions';

-- Create forced_logout_notifications table
CREATE TABLE IF NOT EXISTS forced_logout_notifications (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    notification_id VARCHAR(36) UNIQUE NOT NULL,
    event_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    email_address VARCHAR(255) NOT NULL,
    channel VARCHAR(20) NOT NULL,
    template_name VARCHAR(100) NOT NULL,
    subject TEXT,
    body TEXT,
    variables JSON,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    attempts INT NOT NULL DEFAULT 0,
    last_attempt_at TIMESTAMP NULL,
    sent_at TIMESTAMP NULL,
    failed_at TIMESTAMP NULL,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_fln_event_id (event_id),
    INDEX idx_fln_user_id (user_id),
    INDEX idx_fln_status (status),
    INDEX idx_fln_created_at (created_at DESC),
    CONSTRAINT fk_fln_event_id
        FOREIGN KEY (event_id)
        REFERENCES forced_logout_events(event_id)
        ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Tracks notification delivery for forced logout events';
