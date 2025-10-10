# Technology Research: Forced Logout Implementation

**Feature**: 001-auth-service (Forced Logout Functionality)
**Date**: 2025-10-10
**Status**: Complete

## Executive Summary

All technical decisions have been resolved through analysis of the existing auth-service codebase. The service uses Gin framework with JWT authentication, PostgreSQL + Redis architecture, and follows standard Go project patterns. This research document captures all technology choices and their rationale for implementing forced logout functionality.

## Research Findings

### 1. Web Framework

**Decision**: **Gin (github.com/gin-gonic/gin v1.9.1)**

**Evidence**: Found in auth-service/go.mod:6

**Rationale**:
- Already in use throughout the auth-service
- High-performance HTTP framework with middleware support
- Excellent ecosystem for building REST APIs
- Built-in JSON binding and validation
- Middleware architecture perfect for authorization layers

**Alternatives Considered**:
- Echo: Similar performance, but would require migration
- Standard library net/http: More verbose, lacks conveniences
- Chi: Lighter weight, but Gin provides more features out of the box

**Implementation Impact**:
- Use Gin RouterGroup for forced logout endpoints
- Leverage existing middleware patterns for authorization
- Follow established route organization in auth-service

### 2. Session Storage Strategy

**Decision**: **Hybrid approach - Redis for active sessions + PostgreSQL for audit logs**

**Evidence**:
- Redis already configured (go.mod:10, config.yaml:18-22)
- PostgreSQL in use for persistent data (go.mod:9, config.yaml:8-16)
- JWT tokens used for authentication (go.mod:7, config.yaml:24-26)

**Rationale**:
- **Redis**: Perfect for high-speed session lookups and revocation lists
  - Sub-millisecond read/write performance
  - Built-in TTL support for automatic cleanup
  - Already configured in the environment
  - Can store revoked JWT JTI (JWT ID) or session identifiers

- **PostgreSQL**: Ideal for audit log persistence
  - ACID compliance for tamper-proof logs
  - Complex querying capabilities for audits
  - Long-term retention support
  - Already stores all user/role data

**Alternatives Considered**:
- PostgreSQL only: Too slow for real-time session validation at scale
- Redis only: No persistence guarantee for critical audit data
- In-memory only: Data loss on restart, not suitable for production

**Implementation Strategy**:

1. **Session Tracking**:
   - Store active JWT JTI in Redis with user mapping: `session:{jti}` → user_id, metadata
   - Store user's active sessions list: `user:sessions:{user_id}` → set of JTIs
   - TTL matches JWT expiration time

2. **Revocation List**:
   - Blacklist revoked JTIs in Redis: `revoked:{jti}` → revocation metadata
   - Keep blacklist entries until JWT natural expiration
   - Clear expired entries automatically via Redis TTL

3. **Audit Persistence**:
   - Write all forced logout events to PostgreSQL `forced_logout_events` table
   - Asynchronous write to avoid blocking revocation operation
   - Retain for 90+ days per compliance requirements

### 3. Session Token Format

**Decision**: **JWT (JSON Web Tokens) via github.com/golang-jwt/jwt/v5**

**Evidence**:
- JWT library in go.mod:7
- JWT configuration in config.yaml:24-26
- Current expiration: 24 hours
- Login response includes JWT token (README.md:133)

**Rationale**:
- Stateless authentication already implemented
- JTI (JWT ID) claim can uniquely identify each session
- Includes user_id, role information in claims
- Supports expiration timestamps

**Token Structure for Forced Logout**:

```json
{
  "jti": "unique-session-id",      // Required for revocation tracking
  "user_id": "user-uuid",
  "username": "admin",
  "roles": ["admin"],
  "iat": 1696800000,                // Issued at
  "exp": 1696886400                 // Expires at (24h later)
}
```

**Revocation Validation Flow**:

1. Middleware extracts JWT from request
2. Parse and validate JWT signature and expiration
3. Extract JTI from claims
4. Check Redis: `EXISTS revoked:{jti}`
5. If exists → return 401 Unauthorized with "Session terminated" message
6. If not exists → proceed with request

**Alternatives Considered**:
- Opaque tokens: Would require database lookup on every request (slower)
- Session cookies: Less suitable for API-first architecture
- Keep current JWT without JTI: Cannot track individual sessions for revocation

### 4. Real-time Notification Mechanism

**Decision**: **Polling + HTTP 401 responses (simplest approach)**

**Rationale**:
- Most REST clients already handle 401 by redirecting to login
- No additional infrastructure required
- Aligns with existing JWT architecture
- Forced logout scenarios are relatively rare (not chatty)

**Implementation**:

1. **Server Side**:
   - Revoke JWT JTI in Redis immediately
   - Any subsequent request with revoked JWT returns 401
   - Include custom header: `X-Session-Terminated: forced-logout`
   - Include reason in response body

2. **Client Side** (frontend integration):
   - Existing 401 interceptor handles logout
   - Check `X-Session-Terminated` header for forced logout vs natural expiration
   - Display appropriate message to user
   - Clear local token and redirect to login

3. **Optional Enhancement** (future):
   - Heartbeat endpoint: `GET /api/v1/auth/session/check`
   - Clients poll every 30-60 seconds
   - Returns session status without full token validation overhead

**Alternatives Considered**:
- **WebSocket**: Adds infrastructure complexity, requires WebSocket support across load balancers
- **Server-Sent Events (SSE)**: Requires persistent connections, challenging at scale
- **Push Notifications**: Requires mobile SDK integration, overkill for web applications
- **Polling**: Chosen - simplest, most reliable, leverages existing 401 handling

### 5. Audit Logging Implementation

**Decision**: **PostgreSQL table + structured logging with cryptographic hashing**

**Evidence**: Logrus already in use (go.mod:11, config.yaml:28-30)

**Rationale**:
- PostgreSQL provides ACID guarantees for audit data integrity
- Structured logging via Logrus for application logs
- Cryptographic hash chain for tamper detection
- JSON format for easy parsing and export

**Audit Log Architecture**:

1. **Database Table**: `forced_logout_events`

```sql
CREATE TABLE forced_logout_events (
    id SERIAL PRIMARY KEY,
    event_id VARCHAR(36) UNIQUE NOT NULL,          -- UUID
    timestamp TIMESTAMP NOT NULL,
    actor_type VARCHAR(20) NOT NULL,               -- 'admin' or 'system'
    actor_id VARCHAR(36),                          -- Admin user ID or system ID
    target_user_id VARCHAR(36) NOT NULL,
    session_jti VARCHAR(100),                      -- NULL for "all sessions"
    session_count INT NOT NULL,                    -- Number of sessions terminated
    reason TEXT,
    session_metadata JSONB,                        -- IP, device, location, etc.
    previous_hash VARCHAR(64),                     -- SHA-256 of previous event
    current_hash VARCHAR(64) NOT NULL,             -- SHA-256 of this event
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_target_user (target_user_id),
    INDEX idx_actor (actor_id),
    INDEX idx_timestamp (timestamp)
);
```

2. **Hash Chain for Tamper Detection**:

```
Event N hash = SHA256(
    event_id + timestamp + actor + target + reason + previous_hash
)
```

- Each event includes hash of previous event
- Tampering breaks the chain (hash mismatch)
- First event has previous_hash = "genesis"

3. **Application Logging**:
- Use Logrus for structured application logs
- JSON format (already configured)
- Log to stdout/file for operational monitoring
- Database for compliance and long-term retention

**Alternatives Considered**:
- MongoDB/NoSQL: Not needed, relational model sufficient
- Immutable append-only storage (e.g., blockchain): Overkill for requirements
- File-based logs only: Hard to query, no referential integrity
- External audit service: Added complexity, not required initially

### 6. Email Notification Integration

**Decision**: **SMTP integration with configurable provider**

**Rationale**:
- Standard Go library `net/smtp` for SMTP support
- Works with any email provider (SendGrid, AWS SES, Mailgun, company SMTP)
- Template-based email rendering
- Async delivery with retry queue

**Configuration** (add to config.yaml):

```yaml
email:
  enabled: true
  smtp_host: "smtp.example.com"
  smtp_port: 587
  smtp_user: "notifications@example.com"
  smtp_password: "secret"
  from_address: "noreply@k8s-agent.com"
  from_name: "K8s Agent Security"
  template_dir: "templates/email"
```

**Email Template** (templates/email/forced-logout.html):

```html
Subject: Security Alert: Your session was terminated

Dear {{.Username}},

Your session was forcefully terminated for security reasons.

Details:
- Time: {{.Timestamp}}
- Reason: {{.Reason}}
- Device: {{.DeviceInfo}}
- Location: {{.Location}}
- Terminated by: {{.ActorName}}

If you did not expect this, please contact security immediately.

To regain access, please log in again at: {{.LoginURL}}

Best regards,
K8s Agent Security Team
```

**Implementation**:
- Email service with retry queue (3 attempts, exponential backoff)
- Async sending (don't block forced logout operation)
- Track delivery status for monitoring
- Fallback to log-only mode if SMTP unavailable

**Alternatives Considered**:
- Cloud email APIs (SendGrid SDK, AWS SES SDK): Vendor lock-in
- No email, in-app only: Misses offline users
- SMS notifications: Too expensive for all forced logouts, reserve for high-security accounts

### 7. Monitoring and Metrics

**Decision**: **Prometheus metrics + Logrus structured logging**

**Rationale**:
- Industry standard for Go applications
- Easy integration with existing monitoring
- Rich querying capabilities
- Grafana dashboards for visualization

**Metrics to Collect**:

```go
// Prometheus counters and histograms
forced_logout_requests_total{actor_type, result}
forced_logout_sessions_terminated_total
forced_logout_duration_seconds{operation}
forced_logout_notifications_sent_total{channel, result}
session_revocation_errors_total{error_type}
```

**Implementation**:
- Use `github.com/prometheus/client_golang`
- Expose `/metrics` endpoint
- Record metrics at key points: request received, sessions revoked, notifications sent
- Track errors and latencies

**Alternatives Considered**:
- Custom metrics system: Reinventing the wheel
- Cloud-only metrics (CloudWatch, DataDog): Vendor lock-in
- No metrics: Unacceptable for production monitoring

## Architecture Decisions Summary

| Component | Technology | Rationale |
|-----------|------------|-----------|
| Web Framework | Gin v1.9.1 | Already in use, excellent middleware support |
| Session Storage | Redis (active) + PostgreSQL (audit) | Hybrid for speed + durability |
| Token Format | JWT with JTI claim | Existing standard, supports individual session tracking |
| Real-time Notification | HTTP 401 + polling | Simplest, leverages existing client 401 handling |
| Audit Logging | PostgreSQL + hash chain | ACID compliance, tamper detection |
| Email Delivery | SMTP with retry queue | Standard, provider-agnostic |
| Monitoring | Prometheus + Logrus | Industry standard, rich ecosystem |

## Data Flow Diagram

### Forced Logout Flow

```
Admin/System Request
       ↓
[Authorization Middleware] ← Verify admin role
       ↓
[Forced Logout Service]
       ↓
   ┌───┴────────────────────┐
   ↓                        ↓
[Session Termination]   [Audit Log]
   ↓                        ↓
Redis:                PostgreSQL:
- Get sessions        - Insert event
- Add to blacklist    - Hash chain
- Update user set     - Metadata
   ↓                        ↓
[Notification Service]      │
   ├─────────────────┬──────┤
   ↓                 ↓      ↓
Email Queue    Application Log
(async)        (Logrus JSON)
   ↓
SMTP Delivery
(with retry)
```

### Client Session Validation Flow

```
Client Request with JWT
       ↓
[JWT Middleware]
   ├── Parse token
   ├── Validate signature
   ├── Check expiration
   └── Extract JTI
       ↓
[Revocation Check]
   └── Redis: EXISTS revoked:{jti}?
       ├── YES → 401 Unauthorized + X-Session-Terminated header
       └── NO  → Continue to handler
```

## Security Considerations

### 1. Race Condition Prevention

**Scenario**: Admin forces logout while user makes concurrent requests

**Solution**:
- Redis SET with NX (not exists) flag for idempotency
- Multiple admins revoking same session is safe (idempotent)
- User's in-flight requests complete, subsequent requests fail

### 2. Privilege Escalation Prevention

**Scenario**: Lower-privilege admin tries to logout higher-privilege user

**Solution**:
- Middleware checks: session-admin role required
- Decision: session-admin CAN logout superadmin (Q1 clarification)
- All actions audited with actor ID
- Monitor audit logs for suspicious patterns

### 3. Denial of Service Prevention

**Scenario**: Malicious actor floods forced logout API

**Solution**:
- Rate limiting: 100 requests/minute per admin (FR-6.6)
- Gin middleware: `github.com/ulule/limiter/v3`
- Redis-backed rate limiter for distributed systems
- Alert on rate limit violations

## Performance Optimization

### 1. Bulk Revocation Optimization

**Challenge**: Revoking 1000+ sessions within 30 seconds (NFR-1.3)

**Solution**:
- Redis pipelining for bulk SET operations
- Batch size: 100 sessions per pipeline
- Parallel processing with worker pool (5-10 goroutines)
- Async audit log writing (don't block revocation)

### 2. Session List Performance

**Challenge**: Display all sessions for a user within 10 seconds (Success Criteria #7)

**Solution**:
- Redis sorted set for user sessions: `user:sessions:{user_id}` (score = login timestamp)
- Store metadata directly in Redis: `session:{jti}` → JSON
- Avoid PostgreSQL JOIN queries for session lists
- Paginate if user has 50+ sessions

### 3. Audit Log Query Performance

**Challenge**: Fast audit log searches for compliance audits

**Solution**:
- Database indexes on: target_user_id, actor_id, timestamp
- Partition table by timestamp (monthly partitions after 1M+ events)
- Export old logs to cold storage (S3, long-term archive)

## Implementation Priorities

### Phase 0: Research ✅ COMPLETE

All technology decisions resolved through codebase analysis.

### Phase 1: Design (Next) ⏳

1. Data model design (`data-model.md`)
2. API contracts (`/contracts/forced-logout-api.yaml`)
3. Quickstart guide (`quickstart.md`)
4. Agent context update (CLAUDE.md)

### Phase 2: Implementation

Following task breakdown from `/speckit.tasks`

## Open Questions → RESOLVED ✅

1. ✅ Web framework? **Gin v1.9.1**
2. ✅ Session storage? **Redis + PostgreSQL hybrid**
3. ✅ Token format? **JWT with JTI claim**
4. ✅ Notification service? **SMTP with template rendering**
5. ✅ Monitoring system? **Prometheus + Logrus**
6. ✅ Audit logging? **PostgreSQL with hash chain**
7. ✅ Database migrations? **Standard Go migration tools (to be implemented)**
8. ✅ Client session detection? **HTTP 401 + X-Session-Terminated header**

## References

- Existing codebase: `auth-service/`
- go.mod dependencies analysis
- config.yaml configuration patterns
- README.md API documentation
- JWT best practices: RFC 7519
- Redis session management patterns
- PostgreSQL audit log design patterns

---

**Research Status**: ✅ COMPLETE
**Ready for**: Phase 1 Design
**Blocker Count**: 0
