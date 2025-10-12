# Data Model Design: Forced Logout Functionality

**Feature**: 001-auth-service
**Date**: 2025-10-10
**Status**: Design Complete

## Overview

This document defines the data structures required for forced logout functionality, including session tracking, audit logging, and notification management. The design extends the existing auth-service schema while maintaining backward compatibility.

## Entity Relationship Diagram

```
┌─────────────────┐
│     users       │
│  (existing)     │
└────────┬────────┘
         │ 1
         │
         │ N
┌────────┴──────────────┐
│  user_sessions        │
│  (new - Redis)        │
└───────────────────────┘
         │
         │
         │ N
┌────────┴──────────────────────┐
│  forced_logout_events         │
│  (new - PostgreSQL)           │
└───────────────────────────────┘
         │
         │
         │ N
┌────────┴──────────────────────┐
│  forced_logout_notifications  │
│  (new - PostgreSQL)           │
└───────────────────────────────┘
```

## Data Entities

### 1. UserSession (Redis)

**Storage**: Redis sorted set + hash
**Purpose**: Track active user sessions for fast lookup and revocation
**TTL**: Matches JWT expiration (24 hours default)

**Redis Keys**:

```
# User's active sessions (sorted set by login timestamp)
user:sessions:{user_id} → SortedSet[session_jti]
  Score: login timestamp (Unix epoch)
  Member: JWT JTI (session identifier)

# Individual session metadata (hash)
session:{jti} → Hash {
  "jti": "unique-session-id",
  "user_id": "user-uuid",
  "username": "admin",
  "ip_address": "192.0.2.100",
  "user_agent": "Mozilla/5.0...",
  "device_type": "desktop",
  "device_name": "Chrome 118 on Windows",
  "location": "San Francisco, CA, US",
  "login_at": "2025-10-10T10:00:00Z",
  "last_activity_at": "2025-10-10T11:30:00Z",
  "expires_at": "2025-10-11T10:00:00Z"
}

# Revoked sessions blacklist (string with metadata)
revoked:{jti} → JSON {
  "jti": "session-id",
  "user_id": "user-uuid",
  "revoked_at": "2025-10-10T12:00:00Z",
  "revoked_by": "admin-uuid",
  "reason": "Security policy violation",
  "event_id": "forced-logout-event-uuid"
}
  TTL: Until JWT natural expiration
```

**Go Struct** (for application use):

```go
package types

import "time"

// SessionInfo represents an active user session
type SessionInfo struct {
    JTI             string    `json:"jti"`
    UserID          string    `json:"user_id"`
    Username        string    `json:"username"`
    IPAddress       string    `json:"ip_address"`
    UserAgent       string    `json:"user_agent"`
    DeviceType      string    `json:"device_type"`      // desktop, mobile, tablet
    DeviceName      string    `json:"device_name"`      // "Chrome 118 on Windows"
    Location        string    `json:"location"`         // "City, State, Country"
    LoginAt         time.Time `json:"login_at"`
    LastActivityAt  time.Time `json:"last_activity_at"`
    ExpiresAt       time.Time `json:"expires_at"`
}

// RevokedSession represents a blacklisted session
type RevokedSession struct {
    JTI        string    `json:"jti"`
    UserID     string    `json:"user_id"`
    RevokedAt  time.Time `json:"revoked_at"`
    RevokedBy  string    `json:"revoked_by"`      // admin user ID or "system"
    Reason     string    `json:"reason"`
    EventID    string    `json:"event_id"`        // Links to audit event
}
```

**Indexes**: Redis sorted sets provide O(log N) lookup
**Retention**: Automatic via Redis TTL

### 2. ForcedLogoutEvent (PostgreSQL)

**Table**: `forced_logout_events`
**Purpose**: Immutable audit trail of all forced logout actions
**Retention**: 90 days minimum (configurable)

**Schema**:

```sql
CREATE TABLE forced_logout_events (
    id BIGSERIAL PRIMARY KEY,
    event_id VARCHAR(36) UNIQUE NOT NULL,          -- UUID v4
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Actor information (who performed the logout)
    actor_type VARCHAR(20) NOT NULL,               -- 'admin' | 'system'
    actor_id VARCHAR(36),                          -- User ID (NULL for system)
    actor_username VARCHAR(50),                    -- For quick reference
    actor_ip_address VARCHAR(45),                  -- IPv4 or IPv6

    -- Target information (who was logged out)
    target_user_id VARCHAR(36) NOT NULL,
    target_username VARCHAR(50) NOT NULL,

    -- Session information
    session_jti VARCHAR(100),                      -- NULL = all sessions
    session_count INTEGER NOT NULL,                -- How many sessions terminated
    session_metadata JSONB,                        -- Array of session details

    -- Reason and context
    reason TEXT,
    logout_type VARCHAR(20) NOT NULL,              -- 'single' | 'all' | 'bulk'
    triggered_by VARCHAR(50),                      -- 'manual' | 'policy' | 'security_incident'

    -- Tamper-proof hash chain
    previous_hash VARCHAR(64),                     -- SHA-256 of previous event
    current_hash VARCHAR(64) NOT NULL,             -- SHA-256 of this event

    -- Correlation
    correlation_id VARCHAR(36),                    -- Links to related security events

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Indexes
    INDEX idx_event_id (event_id),
    INDEX idx_timestamp (timestamp DESC),
    INDEX idx_target_user (target_user_id, timestamp DESC),
    INDEX idx_actor (actor_id, timestamp DESC),
    INDEX idx_actor_type (actor_type),
    INDEX idx_logout_type (logout_type)
);

-- Comments for documentation
COMMENT ON TABLE forced_logout_events IS 'Immutable audit trail of forced logout actions';
COMMENT ON COLUMN forced_logout_events.previous_hash IS 'SHA-256 hash of previous event for tamper detection';
COMMENT ON COLUMN forced_logout_events.current_hash IS 'SHA-256(event_id || timestamp || actor_id || target_user_id || reason || previous_hash)';
COMMENT ON COLUMN forced_logout_events.session_metadata IS 'JSON array of {jti, ip_address, device_name, login_at} for each terminated session';
```

**Go Struct**:

```go
package types

import (
    "time"
    "database/sql/driver"
    "encoding/json"
)

// ForcedLogoutEvent represents an audit event
type ForcedLogoutEvent struct {
    ID                int64           `json:"id" gorm:"primaryKey"`
    EventID           string          `json:"event_id" gorm:"uniqueIndex;not null"`
    Timestamp         time.Time       `json:"timestamp" gorm:"not null"`

    // Actor
    ActorType         string          `json:"actor_type" gorm:"not null"`  // admin, system
    ActorID           *string         `json:"actor_id,omitempty"`
    ActorUsername     *string         `json:"actor_username,omitempty"`
    ActorIPAddress    *string         `json:"actor_ip_address,omitempty"`

    // Target
    TargetUserID      string          `json:"target_user_id" gorm:"not null"`
    TargetUsername    string          `json:"target_username" gorm:"not null"`

    // Session
    SessionJTI        *string         `json:"session_jti,omitempty"`
    SessionCount      int             `json:"session_count" gorm:"not null"`
    SessionMetadata   SessionMetadata `json:"session_metadata" gorm:"type:jsonb"`

    // Context
    Reason            *string         `json:"reason,omitempty"`
    LogoutType        string          `json:"logout_type" gorm:"not null"` // single, all, bulk
    TriggeredBy       string          `json:"triggered_by" gorm:"not null"` // manual, policy, security_incident

    // Tamper detection
    PreviousHash      *string         `json:"previous_hash,omitempty"`
    CurrentHash       string          `json:"current_hash" gorm:"not null"`

    // Correlation
    CorrelationID     *string         `json:"correlation_id,omitempty"`

    CreatedAt         time.Time       `json:"created_at"`
}

// SessionMetadata is a JSON array of session details
type SessionMetadata []SessionDetail

// SessionDetail represents a single session in the metadata
type SessionDetail struct {
    JTI        string    `json:"jti"`
    IPAddress  string    `json:"ip_address"`
    DeviceName string    `json:"device_name"`
    LoginAt    time.Time `json:"login_at"`
}

// Scan implements sql.Scanner for JSONB
func (sm *SessionMetadata) Scan(value interface{}) error {
    bytes, ok := value.([]byte)
    if !ok {
        return nil
    }
    return json.Unmarshal(bytes, sm)
}

// Value implements driver.Valuer for JSONB
func (sm SessionMetadata) Value() (driver.Value, error) {
    return json.Marshal(sm)
}

// TableName specifies the table name
func (ForcedLogoutEvent) TableName() string {
    return "forced_logout_events"
}
```

**Validation Rules**:

- `actor_type` ∈ {admin, system}
- `logout_type` ∈ {single, all, bulk}
- `triggered_by` ∈ {manual, policy, security_incident}
- `session_count` > 0
- `session_jti` NULL ⟺ `logout_type` ∈ {all, bulk}
- `current_hash` = SHA256(event_id ∥ timestamp ∥ actor_id ∥ target_user_id ∥ reason ∥ previous_hash)

### 3. ForcedLogoutNotification (PostgreSQL)

**Table**: `forced_logout_notifications`
**Purpose**: Track notification delivery status
**Retention**: 30 days (for debugging)

**Schema**:

```sql
CREATE TABLE forced_logout_notifications (
    id BIGSERIAL PRIMARY KEY,
    notification_id VARCHAR(36) UNIQUE NOT NULL,   -- UUID v4
    event_id VARCHAR(36) NOT NULL,                 -- Links to forced_logout_events

    -- Recipient
    user_id VARCHAR(36) NOT NULL,
    email_address VARCHAR(255) NOT NULL,

    -- Notification details
    channel VARCHAR(20) NOT NULL,                  -- 'email' | 'sms' | 'push'
    template_name VARCHAR(100) NOT NULL,           -- 'forced-logout-email'

    -- Content
    subject TEXT,
    body TEXT,
    variables JSONB,                               -- Template variables

    -- Delivery status
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending' | 'sent' | 'failed'
    attempts INTEGER NOT NULL DEFAULT 0,
    last_attempt_at TIMESTAMP,
    sent_at TIMESTAMP,
    failed_at TIMESTAMP,
    error_message TEXT,

    -- Metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Indexes
    INDEX idx_event_id (event_id),
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at DESC),

    -- Foreign key
    FOREIGN KEY (event_id) REFERENCES forced_logout_events(event_id) ON DELETE CASCADE
);

COMMENT ON TABLE forced_logout_notifications IS 'Tracks notification delivery for forced logout events';
```

**Go Struct**:

```go
package types

import "time"

// ForcedLogoutNotification represents a notification delivery record
type ForcedLogoutNotification struct {
    ID              int64           `json:"id" gorm:"primaryKey"`
    NotificationID  string          `json:"notification_id" gorm:"uniqueIndex;not null"`
    EventID         string          `json:"event_id" gorm:"index;not null"`

    // Recipient
    UserID          string          `json:"user_id" gorm:"index;not null"`
    EmailAddress    string          `json:"email_address" gorm:"not null"`

    // Notification
    Channel         string          `json:"channel" gorm:"not null"` // email, sms, push
    TemplateName    string          `json:"template_name" gorm:"not null"`

    // Content
    Subject         *string         `json:"subject,omitempty"`
    Body            *string         `json:"body,omitempty"`
    Variables       *NotificationVariables `json:"variables,omitempty" gorm:"type:jsonb"`

    // Delivery
    Status          string          `json:"status" gorm:"default:pending"` // pending, sent, failed
    Attempts        int             `json:"attempts" gorm:"default:0"`
    LastAttemptAt   *time.Time      `json:"last_attempt_at,omitempty"`
    SentAt          *time.Time      `json:"sent_at,omitempty"`
    FailedAt        *time.Time      `json:"failed_at,omitempty"`
    ErrorMessage    *string         `json:"error_message,omitempty"`

    CreatedAt       time.Time       `json:"created_at"`
    UpdatedAt       time.Time       `json:"updated_at"`
}

// NotificationVariables holds template variables
type NotificationVariables struct {
    Username    string    `json:"username"`
    Timestamp   time.Time `json:"timestamp"`
    Reason      string    `json:"reason"`
    DeviceInfo  string    `json:"device_info"`
    Location    string    `json:"location"`
    ActorName   string    `json:"actor_name"`
    LoginURL    string    `json:"login_url"`
}

func (NotificationVariables) GormDataType() string {
    return "jsonb"
}

func (ForcedLogoutNotification) TableName() string {
    return "forced_logout_notifications"
}
```

**Validation Rules**:

- `channel` ∈ {email, sms, push}
- `status` ∈ {pending, sent, failed}
- `attempts` ≤ 3 (max retry attempts)
- `sent_at` NOT NULL ⟹ `status` = 'sent'
- `failed_at` NOT NULL ⟹ `status` = 'failed'

## Database Migration

**Migration File**: `migrations/YYYYMMDD_add_forced_logout_tables.up.sql`

```sql
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

COMMIT;
```

**Rollback Migration**: `migrations/YYYYMMDD_add_forced_logout_tables.down.sql`

```sql
BEGIN;

DROP TABLE IF EXISTS forced_logout_notifications CASCADE;
DROP TABLE IF EXISTS forced_logout_events CASCADE;

COMMIT;
```

## Redis Data Structures

### Session Management Operations

**Create Session** (on login):

```redis
# Add to user's session set
ZADD user:sessions:{user_id} {login_timestamp} {jti}

# Store session metadata
HSET session:{jti} field1 value1 field2 value2 ...

# Set TTL to match JWT expiration
EXPIRE session:{jti} 86400
EXPIRE user:sessions:{user_id} 86400
```

**Revoke Session** (forced logout):

```redis
# Add to blacklist
SET revoked:{jti} "{json_metadata}" EX {ttl_seconds}

# Remove from user's active sessions
ZREM user:sessions:{user_id} {jti}

# Delete session metadata
DEL session:{jti}
```

**Check Revocation** (on each request):

```redis
# Check if JTI is blacklisted
EXISTS revoked:{jti}
# Returns 1 if revoked, 0 if active
```

**List User Sessions**:

```redis
# Get all session JTIs for user (sorted by login time)
ZRANGE user:sessions:{user_id} 0 -1 WITHSCORES

# For each JTI, get metadata
HGETALL session:{jti}
```

## Data Retention Policies

| Entity | Storage | Retention | Cleanup Method |
|--------|---------|-----------|----------------|
| Active Sessions | Redis | 24h (JWT expiration) | Redis TTL automatic |
| Revoked Sessions | Redis | Until JWT expiration | Redis TTL automatic |
| Audit Events | PostgreSQL | 90 days minimum | Cron job or partition drop |
| Notifications | PostgreSQL | 30 days | Cron job or DELETE query |

**Cleanup Query** (run daily):

```sql
-- Delete old audit events (retain 90 days)
DELETE FROM forced_logout_events
WHERE created_at < NOW() - INTERVAL '90 days';

-- Delete old notifications (retain 30 days)
DELETE FROM forced_logout_notifications
WHERE created_at < NOW() - INTERVAL '30 days';

-- Vacuum to reclaim space
VACUUM ANALYZE forced_logout_events;
VACUUM ANALYZE forced_logout_notifications;
```

## Performance Considerations

### Session Revocation Performance

**Target**: 99.9% success within 5 seconds for 100 concurrent sessions (NFR-1.1)

**Optimization**:

- Redis pipelining for bulk SET operations
- Worker pool (5-10 goroutines) for parallel processing
- Batch size: 100 sessions per pipeline
- Async audit logging (don't block revocation)

**Benchmark**:

```go
// Pseudo-code for bulk revocation
func RevokeSessions(jtis []string) error {
    pipe := redisClient.Pipeline()

    for _, jti := range jtis {
        pipe.Set(ctx, "revoked:"+jti, metadata, ttl)
        pipe.Del(ctx, "session:"+jti)
    }

    _, err := pipe.Exec(ctx)
    return err
}
```

**Expected Performance**:
- Single session revocation: < 10ms
- 100 sessions: < 1s
- 1000 sessions: < 10s (within 30s requirement)

### Audit Log Query Performance

**Indexes optimized for common queries**:

```sql
-- Query 1: All actions by admin
SELECT * FROM forced_logout_events
WHERE actor_id = 'admin-uuid'
ORDER BY timestamp DESC
LIMIT 100;
-- Uses: idx_fle_actor

-- Query 2: All logouts for user
SELECT * FROM forced_logout_events
WHERE target_user_id = 'user-uuid'
ORDER BY timestamp DESC;
-- Uses: idx_fle_target_user

-- Query 3: Recent events
SELECT * FROM forced_logout_events
WHERE timestamp > NOW() - INTERVAL '7 days'
ORDER BY timestamp DESC;
-- Uses: idx_fle_timestamp
```

## Security Considerations

### Tamper Detection

**Hash Chain Validation**:

```go
func ValidateHashChain(events []ForcedLogoutEvent) bool {
    for i := 1; i < len(events); i++ {
        expected := ComputeHash(events[i-1])
        if events[i].PreviousHash != expected {
            return false // Chain broken, tampering detected
        }
    }
    return true
}

func ComputeHash(event ForcedLogoutEvent) string {
    data := fmt.Sprintf("%s%s%s%s%s%s",
        event.EventID,
        event.Timestamp,
        event.ActorID,
        event.TargetUserID,
        event.Reason,
        event.PreviousHash,
    )
    hash := sha256.Sum256([]byte(data))
    return hex.EncodeToString(hash[:])
}
```

### Data Privacy

**PII Handling**:
- IP addresses stored for security auditing (justified by security requirements)
- Email addresses in notifications table (required for delivery)
- User agents and device info (minimal data, security relevant)
- Retention limited to compliance minimum (90/30 days)

**GDPR Compliance**:
- User can request audit log export (FR-3.5)
- Audit logs excluded from "right to be forgotten" (legal requirement)
- Notification data deleted after 30 days

---

**Data Model Status**: ✅ COMPLETE
**Schema Version**: 1.0.0
**Migration Ready**: Yes
**Next Step**: API Contracts (`/contracts/forced-logout-api.yaml`)
