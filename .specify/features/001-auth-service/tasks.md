# Implementation Tasks: Forced Logout Functionality

**Feature**: 001-auth-service-a
**Created**: 2025-10-10
**Status**: Ready for Implementation
**Source**: [spec.md](./spec.md) | [plan.md](./plan.md) | [data-model.md](./data-model.md)

## Task Organization

Tasks are organized in phases and can be executed in the order listed. Tasks marked with `[P]` can be executed in parallel with other `[P]` tasks in the same phase.

**Estimated Total Effort**: 16-25 days

---

## Phase 0: Foundation & Database Setup (2-3 days)

### T001: Database Migration Files [P] ✅ COMPLETED

**Description**: Create PostgreSQL migration files for new forced logout tables

**Files**:
- `auth-service/migrations/YYYYMMDD_add_forced_logout_tables.up.sql`
- `auth-service/migrations/YYYYMMDD_add_forced_logout_tables.down.sql`

**Acceptance Criteria**:
- [ ] Migration creates `forced_logout_events` table with all fields from data-model.md
- [ ] Migration creates `forced_logout_notifications` table
- [ ] All indexes defined (6 indexes on events, 4 on notifications)
- [ ] Foreign key constraint from notifications to events
- [ ] Table comments added for documentation
- [ ] Rollback migration drops tables in correct order
- [ ] Migration runs without errors on clean database
- [ ] Migration is idempotent (can run multiple times safely)

**Reference**: `data-model.md` lines 158-237

**Estimated Effort**: 0.5 days

---

### T002: Update Configuration Schema [P] ✅ COMPLETED

**Description**: Add email notification configuration to config.yaml

**Files**:
- `auth-service/configs/config.yaml`
- `auth-service/internal/config/config.go` (if exists)

**Acceptance Criteria**:
- [ ] Email configuration section added with SMTP settings
- [ ] Configuration struct updated to include email settings
- [ ] Default values set for development (enabled: false)
- [ ] Configuration validation added
- [ ] README updated with new config options

**Reference**: `research.md` lines 244-256

**Estimated Effort**: 0.25 days

---

### T003: Update JWT to Include JTI Claim [P] ✅ COMPLETED

**Description**: Modify JWT token generation to include unique JTI (JWT ID) for session tracking

**Files**:
- `auth-service/pkg/auth/jwt.go` (or equivalent)
- `auth-service/pkg/types/types.go`

**Acceptance Criteria**:
- [ ] JWT claims struct includes `JTI` field
- [ ] Token generation creates unique JTI (UUID v4)
- [ ] JTI is included in JWT payload
- [ ] Login response includes JTI for client reference
- [ ] Existing tests updated for new claim
- [ ] Backward compatibility maintained for existing tokens

**Reference**: `research.md` lines 148-174

**Estimated Effort**: 0.5 days

---

### T004: Create Core Type Definitions ✅ COMPLETED

**Description**: Define Go structs for all forced logout entities

**Files**:
- `auth-service/pkg/types/session.go` (new)
- `auth-service/pkg/types/forced_logout.go` (new)

**Acceptance Criteria**:
- [ ] SessionInfo struct defined with all fields
- [ ] RevokedSession struct defined
- [ ] ForcedLogoutEvent struct with GORM tags
- [ ] SessionMetadata custom type with JSON marshaling
- [ ] ForcedLogoutNotification struct defined
- [ ] NotificationVariables struct defined
- [ ] All validation tags added (required, min, max, enum)
- [ ] TableName() methods implemented

**Reference**: `data-model.md` lines 45-100, 127-190, 248-295

**Estimated Effort**: 0.5 days

---

## Phase 1: Session Management Infrastructure (3-4 days)

### T005: Redis Session Repository ✅ COMPLETED

**Description**: Implement Redis operations for session storage and revocation

**Files**:
- `auth-service/pkg/forced-logout/session/redis_repository.go` (new)
- `auth-service/pkg/forced-logout/session/repository.go` (interface, new)

**Acceptance Criteria**:
- [ ] Interface defines: StoreSession, GetSession, ListUserSessions, RevokeSession, IsRevoked
- [ ] StoreSession adds to user:sessions sorted set and stores metadata hash
- [ ] GetSession retrieves session metadata from Redis hash
- [ ] ListUserSessions uses ZRANGE with pagination
- [ ] RevokeSession adds to blacklist and removes from active sets
- [ ] IsRevoked checks blacklist efficiently (EXISTS command)
- [ ] All operations set appropriate TTL matching JWT expiration
- [ ] Redis pipelining used for bulk operations
- [ ] Connection pool configured
- [ ] Error handling for Redis failures

**Reference**: `data-model.md` lines 301-339

**Estimated Effort**: 1.5 days

---

### T006: Session Service Layer ✅ COMPLETED

**Description**: Business logic layer for session management

**Files**:
- `auth-service/pkg/forced-logout/session/service.go` (new)
- `auth-service/pkg/forced-logout/session/service_test.go` (new)

**Acceptance Criteria**:
- [ ] Service wraps repository with business logic
- [ ] CreateSession validates and stores session info
- [ ] GetUserSessions retrieves and enriches session data
- [ ] ValidateSession checks both JWT validity and revocation status
- [ ] TerminateSession handles revocation with error handling
- [ ] TerminateUserSessions handles bulk revocation
- [ ] Device type detection from User-Agent
- [ ] Location lookup from IP address (mock for now)
- [ ] Unit tests with >80% coverage
- [ ] Redis mock for testing

**Reference**: `research.md` lines 176-202

**Estimated Effort**: 1.5 days

---

### T007: Update Login Handler to Track Sessions ✅ COMPLETED

**Description**: Modify authentication flow to store session metadata in Redis

**Files**:
- `auth-service/internal/handler/auth_handler.go` (or equivalent)

**Acceptance Criteria**:
- [ ] Login handler extracts IP address from request
- [ ] Login handler extracts User-Agent from headers
- [ ] SessionInfo created with all metadata
- [ ] Session stored in Redis after successful login
- [ ] Session JTI returned in login response
- [ ] Failed storage logged but doesn't block login
- [ ] Existing login tests updated
- [ ] Integration test for session creation

**Reference**: `quickstart.md` lines 180-210

**Estimated Effort**: 1 day

---

## Phase 2: Audit Logging System (2-3 days)

### T008: Audit Repository (PostgreSQL) ✅ COMPLETED

**Description**: Implement database operations for audit event storage

**Files**:
- `auth-service/pkg/forced-logout/audit/repository.go` (interface, new)
- `auth-service/pkg/forced-logout/audit/postgres_repository.go` (new)

**Acceptance Criteria**:
- [ ] Interface defines: CreateEvent, GetEvent, ListEvents, ExportEvents, GetLastHash
- [ ] CreateEvent inserts with hash chain validation
- [ ] GetEvent retrieves by event_id with proper error handling
- [ ] ListEvents supports filtering by user, actor, date range, type
- [ ] ListEvents supports pagination (limit/offset)
- [ ] ExportEvents returns events in specified format (JSON/CSV)
- [ ] GetLastHash retrieves most recent event hash
- [ ] Database transactions used appropriately
- [ ] Prepared statements for security
- [ ] Proper index usage verified

**Reference**: `data-model.md` lines 127-157

**Estimated Effort**: 1.5 days

---

### T009: Hash Chain Implementation ✅ COMPLETED

**Description**: Implement cryptographic hash chain for tamper detection

**Files**:
- `auth-service/pkg/forced-logout/audit/hash_chain.go` (new)
- `auth-service/pkg/forced-logout/audit/hash_chain_test.go` (new)

**Acceptance Criteria**:
- [ ] ComputeHash function using SHA-256
- [ ] Hash includes: event_id, timestamp, actor_id, target_user_id, reason, previous_hash
- [ ] ValidateHashChain verifies entire chain
- [ ] DetectTampering returns specific failure point
- [ ] Genesis event handled (previous_hash = "genesis")
- [ ] Unit tests for hash computation
- [ ] Unit tests for chain validation
- [ ] Test case for tampered event detection

**Reference**: `data-model.md` lines 370-389

**Estimated Effort**: 0.5 days

---

### T010: Audit Service Layer ✅ COMPLETED

**Description**: Business logic for audit event recording

**Files**:
- `auth-service/pkg/forced-logout/audit/service.go` (new)
- `auth-service/pkg/forced-logout/audit/service_test.go` (new)

**Acceptance Criteria**:
- [ ] RecordEvent creates event with hash chain
- [ ] RecordEvent retrieves last hash automatically
- [ ] RecordEvent validates all required fields
- [ ] GetAuditTrail retrieves filtered events
- [ ] ExportAuditLogs formats data for export
- [ ] ValidateIntegrity checks hash chain
- [ ] Async event recording option (for performance)
- [ ] Retry logic for transient DB failures
- [ ] Unit tests with database mock
- [ ] Error scenarios tested

**Reference**: `plan.md` lines 185-200

**Estimated Effort**: 1 day

---

## Phase 3: Notification System (2-3 days)

### T011: Email Template Engine [P] ✅ COMPLETED

**Description**: Template rendering for forced logout notification emails

**Files**:
- `auth-service/templates/email/forced-logout.html` (new)
- `auth-service/templates/email/forced-logout.txt` (new, plain text version)
- `auth-service/pkg/forced-logout/notification/template.go` (new)

**Acceptance Criteria**:
- [ ] HTML email template created with professional design
- [ ] Plain text version for email clients without HTML support
- [ ] Template includes: username, timestamp, reason, device_info, location, actor_name, login_url
- [ ] Template engine uses Go html/template
- [ ] RenderTemplate function validates all variables
- [ ] Template injection prevented (proper escaping)
- [ ] Test email rendering with sample data
- [ ] Template supports internationalization hooks (future)

**Reference**: `research.md` lines 258-284

**Estimated Effort**: 0.5 days

---

### T012: Email Delivery Service [P] ✅ COMPLETED (NotifyHub Integration)

**Description**: ~~SMTP integration for sending notification emails~~ **UPDATED**: Integrated NotifyHub unified notification platform

**Files**:
- ~~`auth-service/pkg/forced-logout/notification/email.go` (deleted - replaced by NotifyHub)~~
- `auth-service/cmd/server/main.go` (updated for NotifyHub initialization)
- `auth-service/NOTIFYHUB_INTEGRATION.md` (new - integration documentation)

**Acceptance Criteria**:
- [X] ~~SMTP connection pool configuration~~ NotifyHub client configuration
- [X] ~~SendEmail sends both HTML and plain text versions~~ Message.Message with HTML body and text metadata
- [X] ~~Email queue for async delivery~~ NotifyHub async sending with SendAsync()
- [X] ~~Retry mechanism (3 attempts, exponential backoff)~~ Built-in NotifyHub retry logic
- [X] Delivery status tracking via Receipt results
- [X] ~~Connection pooling for performance~~ NotifyHub internal connection management
- [X] ~~TLS/SSL support~~ Email platform configuration includes TLS
- [X] Test mode support via WithTestDefaults()
- [X] Health check integration via notifyHubClient.Health()
- [X] Integration documentation created

**Reference**: `NOTIFYHUB_INTEGRATION.md`, `research.md` lines 244-284

**Estimated Effort**: 1.5 days (actual)

---

### T013: Notification Repository ✅ COMPLETED

**Description**: Database operations for notification tracking

**Files**:
- `auth-service/pkg/forced-logout/notification/repository.go` (interface, new)
- `auth-service/pkg/forced-logout/notification/postgres_repository.go` (new)

**Acceptance Criteria**:
- [ ] Interface defines: CreateNotification, UpdateStatus, GetPendingNotifications
- [ ] CreateNotification inserts with foreign key to event
- [ ] UpdateStatus marks as sent/failed with timestamp
- [ ] GetPendingNotifications retrieves notifications for retry
- [ ] IncrementAttempts updates attempt counter
- [ ] Proper error handling
- [ ] Database indexes utilized
- [ ] Unit tests with mock database

**Reference**: `data-model.md` lines 248-295

**Estimated Effort**: 0.5 days

---

### T014: Notification Service Layer ✅ COMPLETED

**Description**: Orchestrate notification delivery and tracking

**Files**:
- `auth-service/pkg/forced-logout/notification/service.go` (new)
- `auth-service/pkg/forced-logout/notification/service_test.go` (new)

**Acceptance Criteria**:
- [ ] NotifyUser creates notification record and sends email
- [ ] Async dispatch using goroutines with wait groups
- [ ] Retry failed notifications (background worker)
- [ ] Track delivery status in database
- [ ] Email disabled mode for testing (config-based)
- [ ] Multiple notification channels (email only for now, extensible)
- [ ] Error aggregation for multiple recipients
- [ ] Unit tests with email mock
- [ ] Integration test end-to-end

**Reference**: `plan.md` lines 223-238

**Estimated Effort**: 0.5 days

---

## Phase 4: Forced Logout Core Logic (3-4 days)

### T015: Forced Logout Service ✅ COMPLETED

**Description**: Core business logic for forced logout operations

**Files**:
- `auth-service/pkg/forced-logout/service.go` (new)
- `auth-service/pkg/forced-logout/service_test.go` (new)

**Acceptance Criteria**:
- [ ] ForceLogoutSession terminates single session by JTI
- [ ] ForceLogoutUser terminates all user sessions
- [ ] BulkForceLogout terminates multiple specified sessions
- [ ] Each operation records audit event
- [ ] Each operation sends notification asynchronously
- [ ] Transaction-like behavior (revoke + audit + notify)
- [ ] Proper error handling with rollback
- [ ] Idempotent operations (safe to retry)
- [ ] Session-admin role validation
- [ ] Unit tests for all three logout types
- [ ] Test error scenarios (Redis down, DB down, etc.)
- [ ] Performance test for bulk operations (100+ sessions)

**Reference**: `spec.md` lines 115-181, `plan.md` lines 93-105

**Estimated Effort**: 2 days

---

### T016: Authorization Middleware ✅ COMPLETED

**Description**: Middleware to enforce session-admin role requirement

**Files**:
- `auth-service/internal/middleware/forced_logout_auth.go` (new)

**Acceptance Criteria**:
- [ ] RequireSessionAdmin middleware checks JWT claims
- [ ] Validates user has session-admin role or higher
- [ ] Allows superadmin role (per Q1 clarification)
- [ ] Returns 403 Forbidden if insufficient permissions
- [ ] Logs authorization attempts (success and failure)
- [ ] Extracts actor information for audit trail
- [ ] Unit tests with mock JWT tokens
- [ ] Integration test with real JWT

**Reference**: `spec.md` lines 128-136

**Estimated Effort**: 0.5 days

---

### T017: Rate Limiting Middleware ✅ COMPLETED

**Description**: Rate limit forced logout API to prevent abuse

**Files**:
- `auth-service/internal/middleware/rate_limit.go` (new or enhance existing)

**Acceptance Criteria**:
- [ ] Rate limiter configured for 100 requests/minute per admin
- [ ] Uses Redis for distributed rate limiting
- [ ] Returns 429 Too Many Requests when exceeded
- [ ] Response headers: X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset, Retry-After
- [ ] Different limits for different endpoints (if needed)
- [ ] Admin bypass option for emergency situations
- [ ] Unit tests with time mocking
- [ ] Load test to verify limit enforcement

**Reference**: `spec.md` line 181, `contracts/forced-logout-api.yaml` lines 459-477

**Estimated Effort**: 1 day

---

### T018: Session Validation Middleware Enhancement ✅ COMPLETED

**Description**: Update JWT middleware to check revocation status

**Files**:
- `auth-service/internal/middleware/jwt.go` (existing, modify)

**Acceptance Criteria**:
- [ ] JWTAuth middleware extracts JTI from token
- [ ] Middleware calls sessionService.IsRevoked(jti)
- [ ] Returns 401 Unauthorized if session revoked
- [ ] Includes X-Session-Terminated: forced-logout header
- [ ] Response body explains termination reason
- [ ] Performance optimized (Redis cache lookup < 5ms)
- [ ] Backward compatible with tokens without JTI (graceful degradation)
- [ ] Unit tests for revoked vs active sessions
- [ ] Integration test full flow

**Reference**: `research.md` lines 176-202, `quickstart.md` lines 338-352

**Estimated Effort**: 0.5 days

---

## Phase 5: API Handlers (2-3 days)

### T019: Session List Handler [P] ✅ COMPLETED

**Description**: API endpoint to list user's active sessions

**Files**:
- `auth-service/internal/handler/session_handler.go` (new)
- `auth-service/internal/handler/session_handler_test.go` (new)

**Acceptance Criteria**:
- [ ] GET /api/v1/sessions/users/:userId endpoint
- [ ] Validates userId parameter (UUID format)
- [ ] Supports pagination (limit, offset query params)
- [ ] Returns SessionListResponse per contract
- [ ] Requires session-admin authorization
- [ ] Handles user not found (404)
- [ ] Handles no active sessions (200 with empty array)
- [ ] Unit tests with mock service
- [ ] Integration test with real Redis

**Reference**: `contracts/forced-logout-api.yaml` lines 24-69

**Estimated Effort**: 0.5 days

---

### T020: Force Logout Session Handler [P] ✅ COMPLETED

**Description**: API endpoint to force logout single session

**Files**:
- `auth-service/internal/handler/forced_logout_handler.go` (new)
- `auth-service/internal/handler/forced_logout_handler_test.go` (new)

**Acceptance Criteria**:
- [ ] POST /api/v1/forced-logout/session/:jti endpoint
- [ ] Validates JTI parameter
- [ ] Parses ForceLogoutRequest from body
- [ ] Calls forcedLogoutService.ForceLogoutSession
- [ ] Returns ForceLogoutResponse per contract
- [ ] Handles session not found (404)
- [ ] Handles invalid request (400)
- [ ] Requires authorization and rate limiting
- [ ] Unit tests for all response codes
- [ ] Integration test end-to-end

**Reference**: `contracts/forced-logout-api.yaml` lines 71-124

**Estimated Effort**: 0.5 days

---

### T021: Force Logout User Handler [P] ✅ COMPLETED

**Description**: API endpoint to force logout all user sessions

**Files**:
- `auth-service/internal/handler/forced_logout_handler.go` (add to existing)

**Acceptance Criteria**:
- [ ] POST /api/v1/forced-logout/user/:userId endpoint
- [ ] Validates userId parameter (UUID format)
- [ ] Parses ForceLogoutRequest from body
- [ ] Calls forcedLogoutService.ForceLogoutUser
- [ ] Returns session count in response
- [ ] Handles user not found (404)
- [ ] Handles user with no active sessions (200, count=0)
- [ ] Requires authorization and rate limiting
- [ ] Unit tests covering edge cases
- [ ] Integration test with multiple sessions

**Reference**: `contracts/forced-logout-api.yaml` lines 126-174

**Estimated Effort**: 0.5 days

---

### T022: Bulk Force Logout Handler [P] ✅ COMPLETED

**Description**: API endpoint to force logout multiple sessions

**Files**:
- `auth-service/internal/handler/forced_logout_handler.go` (add to existing)

**Acceptance Criteria**:
- [ ] POST /api/v1/forced-logout/sessions endpoint
- [ ] Parses BulkForceLogoutRequest (max 100 JTIs)
- [ ] Validates array length (1-100)
- [ ] Calls forcedLogoutService.BulkForceLogout
- [ ] Returns BulkForceLogoutResponse with per-session results
- [ ] Handles partial failures gracefully
- [ ] Returns 200 even with partial failures (details in response)
- [ ] Requires authorization and rate limiting
- [ ] Unit tests for success and partial failure
- [ ] Performance test with 100 JTIs

**Reference**: `contracts/forced-logout-api.yaml` lines 176-220

**Estimated Effort**: 0.5 days

---

### T023: Audit Query Handler [P] ✅ COMPLETED

**Description**: API endpoint to query forced logout audit events

**Files**:
- `auth-service/internal/handler/audit_handler.go` (new)
- `auth-service/internal/handler/audit_handler_test.go` (new)

**Acceptance Criteria**:
- [ ] GET /api/v1/audit/forced-logout endpoint
- [ ] Supports query params: target_user_id, actor_id, actor_type, logout_type, from_date, to_date, limit, offset
- [ ] Validates date format (ISO 8601)
- [ ] Validates enum values (actor_type, logout_type)
- [ ] Returns AuditEventListResponse per contract
- [ ] Pagination included in response
- [ ] Requires session-admin authorization
- [ ] Unit tests for all filter combinations
- [ ] Integration test with sample data

**Reference**: `contracts/forced-logout-api.yaml` lines 222-294

**Estimated Effort**: 0.5 days

---

### T024: Audit Export Handler [P] ✅ COMPLETED

**Description**: API endpoint to export audit logs (JSON/CSV)

**Files**:
- `auth-service/internal/handler/audit_handler.go` (add to existing)

**Acceptance Criteria**:
- [ ] GET /api/v1/audit/forced-logout/export endpoint
- [ ] Supports format query param (json, csv)
- [ ] Supports same filters as list endpoint
- [ ] Returns JSON array for format=json
- [ ] Returns CSV with headers for format=csv
- [ ] Sets appropriate Content-Type header
- [ ] Sets Content-Disposition for file download
- [ ] Handles large exports (streaming if > 1000 events)
- [ ] Requires session-admin authorization
- [ ] Unit tests for both formats
- [ ] Integration test with export download

**Reference**: `contracts/forced-logout-api.yaml` lines 331-365

**Estimated Effort**: 0.5 days

---

### T025: Route Registration ✅ COMPLETED

**Description**: Register all forced logout routes with Gin router

**Files**:
- `auth-service/internal/routes/routes.go` (or cmd/server/main.go)

**Acceptance Criteria**:
- [ ] All 6 endpoints registered under /api/v1
- [ ] Middleware chain configured: JWT → SessionAdmin → RateLimit → Handler
- [ ] Route groups used for organization
- [ ] Swagger/OpenAPI documentation generated (if used)
- [ ] All routes listed in startup log
- [ ] Health check endpoint unaffected
- [ ] Integration test for all routes accessible

**Reference**: `plan.md` lines 93-105

**Estimated Effort**: 0.5 days

---

## Phase 6: Testing & Quality (3-4 days)

### T026: Unit Test Suite [P]

**Description**: Comprehensive unit tests for all components

**Files**:
- All `*_test.go` files created in previous tasks

**Acceptance Criteria**:
- [ ] All services have >80% code coverage
- [ ] All handlers have >80% code coverage
- [ ] Test uses mocks for external dependencies (Redis, PostgreSQL, SMTP)
- [ ] Table-driven tests for multiple scenarios
- [ ] Error cases tested explicitly
- [ ] Edge cases covered (empty lists, nil values, etc.)
- [ ] `go test ./...` passes all tests
- [ ] `go test -cover ./...` shows >80% coverage
- [ ] CI pipeline configured to run tests

**Reference**: General testing best practices

**Estimated Effort**: 2 days

---

### T027: Integration Test Suite [P]

**Description**: End-to-end integration tests with real dependencies

**Files**:
- `auth-service/tests/integration/forced_logout_test.go` (new)
- `auth-service/tests/integration/audit_test.go` (new)

**Acceptance Criteria**:
- [ ] Test uses docker-compose for PostgreSQL and Redis
- [ ] Test creates real sessions and revokes them
- [ ] Test verifies audit log entries
- [ ] Test verifies notification records
- [ ] Test validates hash chain integrity
- [ ] Test covers all 6 API endpoints
- [ ] Test cleanup after execution
- [ ] Tests can run in CI environment
- [ ] `go test -tags=integration ./tests/integration` passes

**Reference**: `quickstart.md` lines 142-285

**Estimated Effort**: 1.5 days

---

### T028: Performance Benchmarks [P]

**Description**: Benchmark tests for performance requirements

**Files**:
- `auth-service/pkg/forced-logout/benchmark_test.go` (new)

**Acceptance Criteria**:
- [ ] Benchmark single session revocation (target: <500ms)
- [ ] Benchmark 100 concurrent sessions (target: <5s)
- [ ] Benchmark 1000 bulk sessions (target: <30s)
- [ ] Benchmark session validation overhead (<5ms per request)
- [ ] Benchmark audit log query performance
- [ ] Results documented with timestamps
- [ ] Redis pipelining effectiveness measured
- [ ] `go test -bench=. -benchmem` runs successfully

**Reference**: `spec.md` NFR-1.1, NFR-1.2, NFR-1.3

**Estimated Effort**: 0.5 days

---

## Phase 7: Monitoring & Observability (1-2 days)

### T029: Prometheus Metrics [P]

**Description**: Instrument code with Prometheus metrics

**Files**:
- `auth-service/pkg/forced-logout/metrics.go` (new)
- `auth-service/cmd/server/main.go` (add /metrics endpoint)

**Acceptance Criteria**:
- [ ] Counter: forced_logout_requests_total{actor_type, result}
- [ ] Counter: forced_logout_sessions_terminated_total
- [ ] Histogram: forced_logout_duration_seconds{operation}
- [ ] Counter: forced_logout_notifications_sent_total{channel, result}
- [ ] Counter: session_revocation_errors_total{error_type}
- [ ] Metrics exposed on /metrics endpoint
- [ ] Metrics incremented in all relevant code paths
- [ ] Grafana dashboard JSON template created
- [ ] Documentation for metric meanings

**Reference**: `research.md` lines 313-336

**Estimated Effort**: 1 day

---

### T030: Structured Logging Enhancement [P]

**Description**: Add detailed structured logging for forced logout operations

**Files**:
- All service files (enhance existing logging)

**Acceptance Criteria**:
- [ ] All forced logout operations logged with structured fields
- [ ] Log fields: event_id, actor_id, target_user_id, session_count, operation, duration_ms
- [ ] Error logs include stack traces
- [ ] Correlation ID propagated through all layers
- [ ] Log levels appropriate (info for success, error for failures)
- [ ] No sensitive data logged (passwords, tokens)
- [ ] JSON format for machine parsing
- [ ] Sample queries documented for common debugging scenarios

**Reference**: `plan.md` lines 223-225

**Estimated Effort**: 0.5 days

---

### T031: Alerting Rules [P]

**Description**: Define Prometheus alerting rules for operational issues

**Files**:
- `auth-service/deployments/prometheus/alerts.yml` (new)

**Acceptance Criteria**:
- [ ] Alert: ForcedLogoutSuccessRateLow (< 99%)
- [ ] Alert: ForcedLogoutAPIErrorRateHigh (> 1%)
- [ ] Alert: SessionStoreUnavailable
- [ ] Alert: NotificationDeliveryFailureSpike
- [ ] Alert: AuditLogWriteFailures
- [ ] All alerts have severity labels (warning, critical)
- [ ] All alerts have runbook annotations
- [ ] Alerts tested with alertmanager

**Reference**: `plan.md` lines 256-273

**Estimated Effort**: 0.5 days

---

## Phase 8: Documentation & Deployment (2-3 days)

### T032: API Documentation [P]

**Description**: Generate and publish API documentation

**Files**:
- `auth-service/docs/api/forced-logout.md` (new)
- README update

**Acceptance Criteria**:
- [ ] API documentation generated from OpenAPI spec
- [ ] All endpoints documented with examples
- [ ] Authentication requirements clearly stated
- [ ] Error responses documented
- [ ] Code samples in curl and Go
- [ ] Postman collection created and exported
- [ ] Documentation published (GitHub Pages or internal docs site)

**Reference**: `contracts/forced-logout-api.yaml`

**Estimated Effort**: 0.5 days

---

### T033: Admin User Guide [P]

**Description**: User guide for security administrators

**Files**:
- `auth-service/docs/admin/forced-logout-guide.md` (new)

**Acceptance Criteria**:
- [ ] Guide explains when to use forced logout
- [ ] Step-by-step procedures for common scenarios
- [ ] Troubleshooting section for common issues
- [ ] Best practices for security incident response
- [ ] Audit log interpretation guide
- [ ] FAQ section
- [ ] Screenshots or examples of typical workflows

**Reference**: `spec.md` User Scenarios

**Estimated Effort**: 0.5 days

---

### T034: Deployment Configuration [P]

**Description**: Production-ready deployment configuration

**Files**:
- `auth-service/deployments/kubernetes/forced-logout-config.yaml` (new)
- `auth-service/Dockerfile` (update if needed)

**Acceptance Criteria**:
- [ ] Kubernetes ConfigMap for forced logout settings
- [ ] Environment variables for production config
- [ ] Redis connection pooling tuned for production
- [ ] PostgreSQL connection limits configured
- [ ] Email SMTP settings templated with secrets
- [ ] Resource limits defined (CPU, memory)
- [ ] Health check endpoints configured
- [ ] Readiness probe includes dependency checks
- [ ] Deployment tested in staging environment

**Reference**: `plan.md` lines 240-275

**Estimated Effort**: 1 day

---

### T035: Database Migration Execution Plan

**Description**: Plan and document database migration execution for production

**Files**:
- `auth-service/docs/deployment/migration-plan.md` (new)

**Acceptance Criteria**:
- [ ] Migration execution steps documented
- [ ] Rollback procedure documented
- [ ] Downtime estimate calculated (if any)
- [ ] Data validation queries provided
- [ ] Pre-migration checklist created
- [ ] Post-migration verification steps
- [ ] Migration tested on production-like dataset
- [ ] Database backup procedure verified

**Reference**: `data-model.md` lines 158-237

**Estimated Effort**: 0.5 days

---

### T036: Runbook Creation [P]

**Description**: Operational runbook for on-call engineers

**Files**:
- `auth-service/docs/operations/forced-logout-runbook.md` (new)

**Acceptance Criteria**:
- [ ] Common alerts and their remediation steps
- [ ] Incident response procedures
- [ ] Escalation contacts
- [ ] Common queries for debugging
- [ ] Performance troubleshooting guide
- [ ] Redis/PostgreSQL failure recovery procedures
- [ ] Data consistency verification steps
- [ ] Rollback procedures for emergencies

**Reference**: `plan.md` lines 292-308

**Estimated Effort**: 0.5 days

---

## Phase 9: Security & Compliance (1-2 days)

### T037: Security Audit & Penetration Testing [P]

**Description**: Security review of forced logout implementation

**Tasks**:
- [ ] Code review for security vulnerabilities
- [ ] SQL injection prevention verified (prepared statements)
- [ ] XSS prevention in email templates (proper escaping)
- [ ] CSRF protection validated (if applicable)
- [ ] Authorization bypass attempts (try logout without session-admin)
- [ ] Rate limit bypass attempts
- [ ] Hash chain tampering attempts
- [ ] Session hijacking prevention verified
- [ ] Secrets management reviewed (no hardcoded credentials)
- [ ] Security findings documented and remediated

**Reference**: `spec.md` NFR-3, `plan.md` lines 310-334

**Estimated Effort**: 1 day

---

### T038: Compliance Verification [P]

**Description**: Verify compliance with audit requirements

**Tasks**:
- [ ] Verify 100% of forced logout actions are logged (FR-3.1)
- [ ] Verify audit logs retained for 90+ days (FR-3.4)
- [ ] Verify audit log integrity (hash chain validation)
- [ ] Verify audit log export functionality (FR-3.5)
- [ ] Verify user notification within 1 minute (FR-4.2)
- [ ] Verify no unauthorized access to audit logs
- [ ] Compliance checklist completed
- [ ] Compliance documentation generated for review

**Reference**: `spec.md` FR-3, Success Criteria #3, #4

**Estimated Effort**: 0.5 days

---

## Phase 10: Final Integration & Release (1-2 days)

### T039: Staging Deployment & Smoke Testing

**Description**: Deploy to staging environment and run smoke tests

**Tasks**:
- [ ] Deploy to staging environment
- [ ] Run automated integration test suite
- [ ] Manual smoke test of all 6 API endpoints
- [ ] Verify audit logs being written
- [ ] Verify email notifications (test SMTP)
- [ ] Verify metrics appearing in Prometheus
- [ ] Load test with realistic traffic patterns
- [ ] Verify alerts trigger correctly
- [ ] Staging sign-off from QA team

**Reference**: `plan.md` lines 277-280

**Estimated Effort**: 1 day

---

### T040: Production Deployment

**Description**: Deploy forced logout feature to production

**Tasks**:
- [ ] Database migration executed (with backup)
- [ ] Application deployed via canary rollout (10% → 50% → 100%)
- [ ] Health checks passing
- [ ] Smoke tests on production (non-disruptive)
- [ ] Monitoring dashboards displaying metrics
- [ ] Alert rules active
- [ ] 24-hour intensive monitoring period
- [ ] No rollback required
- [ ] Feature announcement to users/admins
- [ ] Post-deployment retrospective

**Reference**: `plan.md` lines 277-290

**Estimated Effort**: 0.5 days

---

## Task Dependency Graph

```
Phase 0 (Foundation)
  T001 [P] ─────┐
  T002 [P] ─────┤
  T003 [P] ─────┼─→ Phase 1
  T004      ────┘

Phase 1 (Session Management)
  T005 → T006 → T007

Phase 2 (Audit Logging)
  T008 ─────┐
  T009 [P] ─┼→ T010
            │
Phase 3 (Notifications)
  T011 [P] ─┤
  T012 [P] ─┼→ T013 → T014
            │
Phase 4 (Core Logic)
  T015 ←────┘
  T016 [P]
  T017 [P]
  T018

Phase 5 (API Handlers)
  T019 [P] ─┐
  T020 [P] ─┤
  T021 [P] ─┼→ T025
  T022 [P] ─┤
  T023 [P] ─┤
  T024 [P] ─┘

Phase 6 (Testing)
  T026 [P]
  T027 [P]
  T028 [P]

Phase 7 (Monitoring)
  T029 [P]
  T030 [P]
  T031 [P]

Phase 8 (Documentation)
  T032 [P]
  T033 [P]
  T034 [P]
  T035
  T036 [P]

Phase 9 (Security)
  T037 [P]
  T038 [P]

Phase 10 (Release)
  T039 → T040
```

## Progress Tracking

Track progress by checking off acceptance criteria for each task. Update this document as tasks are completed.

**Overall Status**: 25/40 tasks completed (62.5%) ✅

**Phase Completion**:
- Phase 0: 4/4 tasks ✅ COMPLETE
- Phase 1: 3/3 tasks ✅ COMPLETE
- Phase 2: 3/3 tasks ✅ COMPLETE
- Phase 3: 4/4 tasks ✅ COMPLETE (Note: T012 replaced with NotifyHub integration)
- Phase 4: 4/4 tasks ✅ COMPLETE
- Phase 5: 7/7 tasks ✅ COMPLETE
- Phase 6: 0/3 tasks ⏳ NEXT
- Phase 7: 0/3 tasks
- Phase 8: 0/5 tasks
- Phase 9: 0/2 tasks
- Phase 10: 0/2 tasks

**Latest Update**: 2025-10-10
- Completed Phases 0-5 (Foundation through API Handlers)
- Integrated NotifyHub for unified notifications (replaced custom SMTP)
- Created comprehensive integration documentation (NOTIFYHUB_INTEGRATION.md)
- Ready to proceed with Phase 6: Testing & Quality

## Next Steps

1. Review this task list with the development team
2. Assign tasks to team members
3. Set up project tracking (JIRA, GitHub Projects, etc.)
4. Begin with Phase 0 tasks (can run in parallel)
5. Run `/speckit.implement` to start guided implementation

---

**Tasks Document Status**: ✅ Complete
**Ready for**: Implementation (`/speckit.implement`)
**Total Tasks**: 40
**Total Estimated Effort**: 16-25 days
