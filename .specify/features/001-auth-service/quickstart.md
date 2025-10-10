# Quickstart Guide: Forced Logout Development

**Feature**: 001-auth-service (Forced Logout Functionality)
**Audience**: Developers implementing and testing forced logout
**Updated**: 2025-10-10

## Overview

This guide provides step-by-step instructions for developing and testing the forced logout functionality locally. It covers environment setup, running scenarios, and API usage examples.

## Prerequisites

Before starting, ensure you have:

- **Go 1.21+** installed
- **Docker** and **docker-compose** for PostgreSQL and Redis
- **Make** utility
- **curl** or **Postman** for API testing
- **Git** (already cloned the repository)

## Environment Setup

### 1. Start Dependencies

Use the existing docker-compose setup:

```bash
cd auth-service

# Start PostgreSQL and Redis
docker-compose up -d postgres redis

# Verify services are running
docker-compose ps

# Expected output:
#   postgres   Up   0.0.0.0:5432->5432/tcp
#   redis      Up   0.0.0.0:6379->6379/tcp
```

### 2. Run Database Migrations

```bash
# Install migration tool (if not already installed)
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations (including new forced logout tables)
migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/k8s_agent_auth?sslmode=disable" up

# Verify tables exist
psql -h localhost -U postgres -d k8s_agent_auth -c "\dt"

# Expected tables:
#   users
#   roles
#   permissions
#   user_roles
#   role_permissions
#   api_keys
#   forced_logout_events         (NEW)
#   forced_logout_notifications  (NEW)
```

### 3. Install Dependencies

```bash
# Install Go dependencies
go mod tidy
go mod download

# Verify no errors
go build ./...
```

### 4. Configure Environment

Edit `configs/config.yaml` or set environment variables:

```yaml
# configs/config.yaml
server:
  host: "0.0.0.0"
  port: 8090
  mode: "debug"

database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "postgres"
  dbname: "k8s_agent_auth"
  sslmode: "disable"

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 2

jwt:
  secret: "dev-secret-key-change-in-production"
  expires_hours: 24

logging:
  level: "debug"
  format: "json"

# NEW: Email configuration for forced logout notifications
email:
  enabled: false  # Set true when SMTP configured
  smtp_host: "smtp.example.com"
  smtp_port: 587
  smtp_user: "notifications@example.com"
  smtp_password: ""
  from_address: "noreply@k8s-agent.com"
  from_name: "K8s Agent Security"
```

### 5. Start Auth Service

```bash
# Run the service
go run cmd/server/main.go

# Or using Make
make run

# Expected output:
# [GIN-debug] Listening and serving HTTP on 0.0.0.0:8090
```

## Testing Scenarios

### Scenario 1: Basic Forced Logout Flow

This scenario tests the complete forced logout workflow from admin action to user notification.

#### Step 1: Create Test Users

```bash
# Login as super admin (created during initialization)
curl -X POST http://localhost:8090/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }' | jq

# Save the token
export ADMIN_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Create a test user
curl -X POST http://localhost:8090/api/v1/users \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123",
    "email": "testuser@example.com",
    "real_name": "Test User"
  }' | jq

# Save the user ID
export TEST_USER_ID="550e8400-e29b-41d4-a716-446655440000"
```

#### Step 2: Create Multiple Sessions for Test User

```bash
# Login as test user (Session 1 - Desktop)
curl -X POST http://localhost:8090/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -H "User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/118" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }' | jq

export SESSION1_TOKEN="..."
export SESSION1_JTI="abc123xyz"  # Extract from JWT

# Login again (Session 2 - Mobile)
curl -X POST http://localhost:8090/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -H "User-Agent: Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X)" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }' | jq

export SESSION2_TOKEN="..."
export SESSION2_JTI="def456uvw"

# Login again (Session 3 - Tablet)
curl -X POST http://localhost:8090/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -H "User-Agent: Mozilla/5.0 (iPad; CPU OS 16_0 like Mac OS X)" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }' | jq

export SESSION3_TOKEN="..."
```

#### Step 3: Verify Sessions Are Active

```bash
# Test user makes authenticated request (should succeed)
curl -X GET http://localhost:8090/api/v1/auth/me \
  -H "Authorization: Bearer $SESSION1_TOKEN" | jq

# Expected: 200 OK with user info

# List all active sessions for test user (as admin)
curl -X GET "http://localhost:8090/api/v1/sessions/users/$TEST_USER_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq

# Expected output:
# {
#   "user_id": "550e8400...",
#   "username": "testuser",
#   "total": 3,
#   "sessions": [
#     {
#       "jti": "abc123xyz",
#       "device_type": "desktop",
#       "device_name": "Chrome 118 on Windows 10",
#       "ip_address": "127.0.0.1",
#       "login_at": "2025-10-10T10:00:00Z",
#       ...
#     },
#     ...
#   ]
# }
```

#### Step 4: Force Logout Single Session

```bash
# Admin forces logout of Session 1 (desktop)
curl -X POST "http://localhost:8090/api/v1/forced-logout/session/$SESSION1_JTI" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "reason": "Testing forced logout - suspicious activity",
    "triggered_by": "manual"
  }' | jq

# Expected output:
# {
#   "event_id": "uuid...",
#   "success": true,
#   "session_count": 1,
#   "target_user_id": "550e8400...",
#   "target_username": "testuser",
#   "timestamp": "2025-10-10T11:30:00Z",
#   "message": "Successfully terminated 1 session"
# }
```

#### Step 5: Verify Session Revocation

```bash
# Try to use the revoked session (should fail)
curl -X GET http://localhost:8090/api/v1/auth/me \
  -H "Authorization: Bearer $SESSION1_TOKEN" -i

# Expected: 401 Unauthorized
# Headers include: X-Session-Terminated: forced-logout

# Other sessions still work
curl -X GET http://localhost:8090/api/v1/auth/me \
  -H "Authorization: Bearer $SESSION2_TOKEN" | jq

# Expected: 200 OK

# Check remaining sessions
curl -X GET "http://localhost:8090/api/v1/sessions/users/$TEST_USER_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq

# Expected: total = 2 (sessions 2 and 3 remain)
```

#### Step 6: Force Logout All User Sessions

```bash
# Admin forces logout of all remaining sessions
curl -X POST "http://localhost:8090/api/v1/forced-logout/user/$TEST_USER_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "reason": "Account security review",
    "triggered_by": "security_incident"
  }' | jq

# Expected output:
# {
#   "success": true,
#   "session_count": 2,
#   "message": "Successfully terminated 2 sessions for user testuser"
# }

# Verify all sessions are terminated
curl -X GET http://localhost:8090/api/v1/auth/me \
  -H "Authorization: Bearer $SESSION2_TOKEN" -i

# Expected: 401 Unauthorized

curl -X GET http://localhost:8090/api/v1/auth/me \
  -H "Authorization: Bearer $SESSION3_TOKEN" -i

# Expected: 401 Unauthorized
```

#### Step 7: Check Audit Trail

```bash
# View audit events for the test user
curl -X GET "http://localhost:8090/api/v1/audit/forced-logout?target_user_id=$TEST_USER_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq

# Expected: List of 2 events (single session + all sessions)

# Get detailed event
curl -X GET "http://localhost:8090/api/v1/audit/forced-logout/{event_id}" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq

# Expected: Full event details including session metadata, hash chain
```

### Scenario 2: Bulk Session Termination

Test terminating multiple specific sessions at once.

```bash
# Create 5 sessions for test user
# ... (repeat login 5 times) ...

# Collect session JTIs
JTIS='["jti1", "jti2", "jti3", "jti4", "jti5"]'

# Bulk force logout
curl -X POST http://localhost:8090/api/v1/forced-logout/sessions \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "session_jtis": '"$JTIS"',
    "reason": "Bulk security policy enforcement",
    "triggered_by": "policy"
  }' | jq

# Expected output:
# {
#   "event_id": "uuid...",
#   "total_requested": 5,
#   "successful": 5,
#   "failed": 0,
#   "results": [
#     {"jti": "jti1", "success": true},
#     {"jti": "jti2", "success": true},
#     ...
#   ]
# }
```

### Scenario 3: Testing Error Cases

#### Invalid Session JTI

```bash
curl -X POST http://localhost:8090/api/v1/forced-logout/session/invalid-jti-999 \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"reason": "Test"}' | jq

# Expected: 404 Not Found
# {
#   "error": "NOT_FOUND",
#   "message": "Session not found or already expired"
# }
```

#### Insufficient Permissions

```bash
# Try forced logout without admin role
curl -X POST "http://localhost:8090/api/v1/forced-logout/user/$TEST_USER_ID" \
  -H "Authorization: Bearer $SESSION2_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"reason": "Test"}' -i

# Expected: 403 Forbidden
# {
#   "error": "FORBIDDEN",
#   "message": "Requires session-admin role or higher"
# }
```

#### Rate Limiting

```bash
# Send 101 requests rapidly (exceeds 100/minute limit)
for i in {1..101}; do
  curl -X POST "http://localhost:8090/api/v1/forced-logout/user/$TEST_USER_ID" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"reason": "Rate limit test"}' &
done
wait

# Expected: Last request returns 429 Too Many Requests
# Headers include:
#   X-RateLimit-Limit: 100
#   X-RateLimit-Remaining: 0
#   Retry-After: 60
```

## Redis Inspection

Verify Redis data structures during testing:

```bash
# Connect to Redis
redis-cli -h localhost -p 6379

# Select database (db 2 per config)
SELECT 2

# Check active sessions for user
ZRANGE user:sessions:550e8400-e29b-41d4-a716-446655440000 0 -1 WITHSCORES

# View session metadata
HGETALL session:abc123xyz

# Check revoked sessions blacklist
KEYS revoked:*
GET revoked:abc123xyz

# Check TTL
TTL session:abc123xyz
TTL revoked:abc123xyz
```

## PostgreSQL Inspection

Query audit data directly:

```bash
# Connect to PostgreSQL
psql -h localhost -U postgres -d k8s_agent_auth

# View recent forced logout events
SELECT
  event_id,
  timestamp,
  actor_username,
  target_username,
  session_count,
  logout_type,
  reason
FROM forced_logout_events
ORDER BY timestamp DESC
LIMIT 10;

# Check hash chain integrity
SELECT
  event_id,
  previous_hash,
  current_hash
FROM forced_logout_events
ORDER BY id;

# View notification delivery status
SELECT
  notification_id,
  user_id,
  channel,
  status,
  attempts,
  sent_at,
  error_message
FROM forced_logout_notifications
ORDER BY created_at DESC;
```

## Common Development Tasks

### Resetting Test Data

```bash
# Clear all sessions from Redis
redis-cli -h localhost -p 6379 -n 2 FLUSHDB

# Clear audit tables (PostgreSQL)
psql -h localhost -U postgres -d k8s_agent_auth <<EOF
TRUNCATE TABLE forced_logout_notifications CASCADE;
TRUNCATE TABLE forced_logout_events CASCADE;
EOF
```

### Debugging Tips

1. **Enable Debug Logging**:

   ```yaml
   # configs/config.yaml
   logging:
     level: "debug"  # Shows Redis commands, SQL queries
   ```

2. **Monitor Redis Commands**:

   ```bash
   redis-cli -h localhost -p 6379 MONITOR
   ```

3. **Watch PostgreSQL Queries**:

   ```bash
   # In psql
   ALTER SYSTEM SET log_statement = 'all';
   SELECT pg_reload_conf();
   ```

4. **Check Application Logs**:

   ```bash
   # If using systemd
   journalctl -u auth-service -f

   # Or if running directly
   go run cmd/server/main.go 2>&1 | jq '.'  # Pretty-print JSON logs
   ```

### Running Unit Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test package
go test ./pkg/forced-logout/...

# Run with verbose output
go test -v ./pkg/forced-logout/service_test.go
```

### Running Integration Tests

```bash
# Ensure dependencies are running
docker-compose up -d

# Run integration tests
go test -tags=integration ./tests/integration/...

# Cleanup after tests
docker-compose down
```

## Performance Testing

### Benchmark Single Session Revocation

```bash
# Use Apache Bench
ab -n 1000 -c 10 -H "Authorization: Bearer $ADMIN_TOKEN" \
   -p logout-request.json -T application/json \
   "http://localhost:8090/api/v1/forced-logout/session/test-jti"

# Expected: < 500ms per request (NFR-1.2)
```

### Benchmark Bulk Revocation

```bash
# Create 100 sessions
# ... (script to create sessions) ...

# Time bulk logout
time curl -X POST http://localhost:8090/api/v1/forced-logout/sessions \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"session_jtis": [...100 JTIs...], "reason": "Benchmark"}'

# Expected: < 5 seconds (NFR-1.1)
```

## Troubleshooting

### Issue: Sessions Not Appearing in List

**Cause**: Session tracking not implemented in login flow

**Solution**: Verify JWT includes `jti` claim and session is stored in Redis:

```go
// Ensure login handler stores session
sessionInfo := &types.SessionInfo{
    JTI: jti,
    UserID: user.ID,
    // ... other fields ...
}
sessionService.StoreSession(ctx, sessionInfo)
```

### Issue: Forced Logout Not Working

**Cause**: JTI not being validated in auth middleware

**Solution**: Check middleware validates against revocation list:

```go
// In JWT middleware
if revoked, _ := sessionService.IsRevoked(ctx, jti); revoked {
    c.AbortWithStatusJSON(401, gin.H{
        "error": "UNAUTHORIZED",
        "message": "Session has been terminated",
    })
    return
}
```

### Issue: Audit Logs Missing

**Cause**: Async audit logging failed

**Solution**: Check PostgreSQL connection and logs:

```bash
# Test database connection
psql -h localhost -U postgres -d k8s_agent_auth -c "SELECT 1;"

# Check application logs for DB errors
grep "audit" logs/app.log
```

## Next Steps

After completing this quickstart:

1. Review API contracts: `.specify/features/001-auth-service/contracts/forced-logout-api.yaml`
2. Explore data model: `.specify/features/001-auth-service/data-model.md`
3. Run `/speckit.tasks` to generate detailed implementation task list
4. Begin implementation following task order

## References

- **Specification**: `spec.md`
- **Implementation Plan**: `plan.md`
- **Research Document**: `research.md`
- **API Contracts**: `contracts/forced-logout-api.yaml`
- **Data Model**: `data-model.md`

---

**Quickstart Status**: ✅ Ready for Development
**Last Updated**: 2025-10-10
