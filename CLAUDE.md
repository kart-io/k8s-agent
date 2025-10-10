# k8s-agent Development Guidelines

Auto-generated from all feature plans. Last updated: 2025-10-10

## Active Technologies

- Go 1.21 (001-auth-service-a)
- Go 1.21+ (002-gorm-kart-io)
- PostgreSQL 13+ (existing schema, no changes) (002-gorm-kart-io)
- Gin v1.9.1 Web Framework (001-auth-service-a)
- JWT Authentication (golang-jwt/jwt/v5) (001-auth-service-a)
- Redis (go-redis/v9) for session storage (001-auth-service-a)
- PostgreSQL for audit logging (001-auth-service-a)
- Logrus for structured logging (001-auth-service-a)

## Project Structure

```
src/
tests/
auth-service/
  ├── cmd/                # Application entry points
  ├── internal/           # Private application code
  ├── pkg/                # Public packages
  │   ├── types/          # Data type definitions
  │   └── forced-logout/  # NEW: Forced logout implementation
  ├── configs/            # Configuration files
  ├── migrations/         # Database migrations
  └── templates/          # Email templates
```

## Commands

```bash
# Run auth-service
cd auth-service && go run cmd/server/main.go

# Or using Make
cd auth-service && make run

# Run tests
go test ./...

# Run migrations
migrate -path migrations -database "postgresql://..." up
```

## Code Style

Go 1.21: Follow standard conventions

### Forced Logout Feature (001-auth-service-a)

**Architecture**:

- **Session Tracking**: Redis sorted sets for O(log N) lookups
- **Revocation**: JWT JTI blacklist in Redis with TTL
- **Audit Logging**: PostgreSQL with hash chain for tamper detection
- **Notifications**: Async email delivery with retry queue

**Key Components**:

1. **Session Service** (`pkg/forced-logout/session/`):
   - StoreSession() - Track new sessions in Redis
   - ListUserSessions() - Retrieve all active sessions
   - IsRevoked() - Check if JTI is blacklisted

2. **Forced Logout Service** (`pkg/forced-logout/service/`):
   - ForceLogoutSession() - Terminate single session
   - ForceLogoutUser() - Terminate all user sessions
   - BulkForceLogout() - Terminate multiple sessions

3. **Audit Service** (`pkg/forced-logout/audit/`):
   - RecordEvent() - Write to PostgreSQL with hash chain
   - GetEvents() - Query audit history
   - ExportEvents() - Export logs as JSON/CSV

4. **Notification Service** (`pkg/forced-logout/notification/`):
   - SendEmail() - Async email delivery
   - RenderTemplate() - Email template rendering
   - RetryFailed() - Retry failed notifications

**Data Models**:

- Redis: `user:sessions:{user_id}`, `session:{jti}`, `revoked:{jti}`
- PostgreSQL: `forced_logout_events`, `forced_logout_notifications`

**API Endpoints**:

- GET `/api/v1/sessions/users/{userId}` - List sessions
- POST `/api/v1/forced-logout/session/{jti}` - Force logout single
- POST `/api/v1/forced-logout/user/{userId}` - Force logout all
- POST `/api/v1/forced-logout/sessions` - Bulk logout
- GET `/api/v1/audit/forced-logout` - Audit history

**Security**:

- Requires `session-admin` role or higher
- Rate limiting: 100 requests/minute per admin
- All actions logged with actor, target, timestamp, reason
- Hash chain for audit log integrity

**Performance Targets**:

- Single session revocation: < 500ms (NFR-1.2)
- 100 concurrent sessions: < 5s (NFR-1.1)
- 1000+ bulk sessions: < 30s (NFR-1.3)

**Testing**:

See `.specify/features/001-auth-service/quickstart.md` for:
- Development environment setup
- API usage examples
- Testing scenarios
- Debugging tips

## Recent Changes

- 002-gorm-kart-io: Added Go 1.21+
- 001-auth-service: Added Go 1.21
- 001-auth-service-a: Added forced logout functionality with session management, audit logging, and notifications

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
