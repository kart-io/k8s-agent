# Feature Specification: Code Optimization - GORM and kart-io/logger Integration

**Feature Branch**: `002-gorm-kart-io`
**Created**: 2025-10-09
**Status**: Draft
**Input**: User description: "优化代码使用GORM和kart-io/logger"

## Clarifications

### Session 2025-10-09

- Q: How should the GORM migration be deployed? → A: All-at-once deployment with brief downtime (~5-10 min) for schema verification
- Q: How should existing Redis cache entries be handled during GORM migration? → A: Flush all cache on deployment, let GORM hooks repopulate fresh (brief cache miss spike accepted)
- Q: How should GORM handle incompatible schema changes? → A: Disable AutoMigrate for production (only use in dev/test), manual migrations in prod

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Database Operations with GORM (Priority: P1)

As a developer maintaining the auth-service codebase, I need to use GORM for all database operations so that I can benefit from automatic migrations, type-safe queries, and reduced boilerplate code while maintaining the same functionality.

**Why this priority**: This is the foundation for all data access. Current code uses raw SQL with database/sql, which requires manual query construction, manual scanning, and manual migration scripts. GORM provides type safety, reduces code verbosity, and prevents SQL injection vulnerabilities through its query builder.

**Independent Test**: Can be fully tested by verifying that all existing database operations (user CRUD, role management, permission queries, API key management) work identically to the current implementation, but use GORM models and methods instead of raw SQL queries.

**Acceptance Scenarios**:

1. **Given** the service is using GORM models, **When** a user login request is processed, **Then** the system retrieves user data using GORM's `First()` method and authentication succeeds with correct credentials
2. **Given** GORM auto-migration is enabled in development environment, **When** the service starts, **Then** all database tables are created/updated automatically based on model definitions
3. **Given** a user creation request with role assignments, **When** the service creates the user, **Then** GORM transactions ensure atomic creation of user and role associations
4. **Given** permission checks are needed, **When** querying user permissions, **Then** GORM eager loading (Preload) efficiently fetches related roles and permissions in a single query
5. **Given** pagination parameters are provided, **When** listing users, **Then** GORM's `Limit()` and `Offset()` methods return the correct page of results

---

### User Story 2 - Structured Logging with kart-io/logger (Priority: P2)

As a developer monitoring the service in production, I need all logging to use the kart-io/logger library so that I can leverage dual-engine support (Zap/Slog), OTLP integration, and consistent structured logging across the kart-io ecosystem.

**Why this priority**: Current code uses logrus, which doesn't integrate with the kart-io ecosystem. The kart-io/logger provides superior performance, OTLP export for distributed tracing, and standardized field formats. This enables better observability and troubleshooting.

**Independent Test**: Can be fully tested by running the service with kart-io/logger configured, performing various operations (login, user creation, permission checks), and verifying that logs are emitted with correct structure, fields, and levels through both console output and OTLP collectors.

**Acceptance Scenarios**:

1. **Given** kart-io/logger is initialized with Zap engine, **When** a request is processed, **Then** logs contain structured fields (user_id, method, path, status, latency) in JSON format
2. **Given** OTLP endpoint is configured, **When** authentication events occur, **Then** logs are exported to the OTLP collector for distributed tracing
3. **Given** log level is set to DEBUG, **When** database queries execute, **Then** GORM query logs are captured with proper context and timing information
4. **Given** an error occurs during permission check, **When** the error is logged, **Then** the log includes stack trace, error details, and contextual fields (user_id, permission_code)
5. **Given** request logging middleware is active, **When** HTTP requests complete, **Then** logs include all standard fields (method, path, status, latency_ms, client_ip) using kart-io/logger

---

### User Story 3 - GORM Caching Integration (Priority: P3)

As a developer optimizing performance, I need to integrate Redis caching with GORM hooks so that frequently accessed data (permissions, roles) is automatically cached and invalidated without manual cache management code.

**Why this priority**: Current code has manual caching logic. GORM hooks (AfterFind, AfterCreate, AfterUpdate, AfterDelete) can automate cache population and invalidation, reducing code duplication and preventing cache inconsistency bugs.

**Independent Test**: Can be fully tested by performing operations that trigger cache population (permission queries) and cache invalidation (role updates), then verifying cache hit/miss metrics and confirming cached data matches database state.

**Acceptance Scenarios**:

1. **Given** a permission query is executed, **When** GORM's AfterFind hook runs, **Then** the permission data is automatically cached in Redis with a 15-minute TTL
2. **Given** a role is updated, **When** GORM's AfterUpdate hook runs, **Then** all related permission caches are automatically invalidated
3. **Given** cached permission data exists, **When** the same permission query runs, **Then** GORM's BeforeQuery hook checks Redis first and skips database query if data is cached

---

### Edge Cases

- GORM auto-migration is disabled in production to prevent unexpected schema changes; development/test environments use AutoMigrate for rapid iteration
- How does the system handle GORM connection pool exhaustion during high concurrent load?
- What happens when kart-io/logger OTLP endpoint is unreachable?
- How does GORM handle null values in existing database records that were created with raw SQL?
- What happens when GORM transactions fail mid-operation (e.g., role assignment after user creation)?
- Existing Redis cache entries created by old caching code will be flushed during deployment to ensure consistency with new GORM hook-based caching

## Requirements *(mandatory)*

### Functional Requirements

#### Database Migration (GORM)

- **FR-001**: System MUST replace all raw SQL queries with GORM query methods while maintaining identical query results
- **FR-002**: System MUST define GORM models for all existing tables (users, roles, permissions, user_roles, role_permissions, api_keys)
- **FR-003**: System MUST use GORM's AutoMigrate for table creation and schema updates in development/test environments only; production uses manual migration scripts
- **FR-004**: System MUST maintain all existing foreign key constraints and indexes in GORM model definitions
- **FR-005**: System MUST use GORM transactions for multi-table operations (e.g., user creation with role assignment)
- **FR-006**: System MUST configure GORM connection pool settings to match current database/sql configuration (MaxOpenConns: 25, MaxIdleConns: 5)
- **FR-007**: System MUST use GORM's Preload/Joins for fetching related data (user roles, role permissions) to avoid N+1 queries
- **FR-008**: System MUST preserve existing soft delete behavior using GORM's `gorm.DeletedAt` field
- **FR-009**: System MUST handle nullable fields (expires_at, last_used_at in api_keys) using pointer types in GORM models

#### Logging Migration (kart-io/logger)

- **FR-010**: System MUST replace all logrus calls with kart-io/logger calls throughout the codebase
- **FR-011**: System MUST initialize kart-io/logger with configuration from config.yaml (engine type, level, format)
- **FR-012**: System MUST configure kart-io/logger to use Zap engine by default for high performance
- **FR-013**: System MUST support OTLP export configuration through environment variables or config file
- **FR-014**: System MUST preserve all existing log levels (Debug, Info, Warn, Error) and structured fields
- **FR-015**: System MUST log GORM query execution with timing information when debug logging is enabled
- **FR-016**: System MUST maintain request logging middleware to use kart-io/logger with structured fields (method, path, status, latency_ms, user_id)

#### Integration & Compatibility

- **FR-017**: System MUST maintain backward compatibility with existing database schema (no data loss during migration)
- **FR-018**: System MUST maintain all existing API endpoints and response formats
- **FR-019**: System MUST preserve existing Prometheus metrics collection
- **FR-020**: System MUST ensure Docker Compose setup works with GORM and kart-io/logger
- **FR-021**: System MUST update go.mod dependencies to include GORM (`gorm.io/gorm`, `gorm.io/driver/postgres`) and remove direct database/sql usage
- **FR-022**: System MUST update go.mod dependencies to include kart-io/logger and remove logrus dependency
- **FR-023**: Deployment MUST include rollback plan to revert to previous version if GORM auto-migration fails or critical issues detected within 10 minutes post-deployment
- **FR-024**: Deployment script MUST flush all Redis cache entries before starting the new service version to ensure cache consistency with GORM hook-based caching

### Key Entities

**Note**: These entities already exist in the current system. This spec describes how they will be represented as GORM models.

- **User**: Represents system users with attributes (id, username, password_hash, email, real_name, phone, avatar, status, created_at, updated_at). Has many-to-many relationship with Roles through UserRoles.

- **Role**: Represents user roles with attributes (id, name, code, description, status, sort, created_at, updated_at). Has many-to-many relationship with Users and Permissions.

- **Permission**: Represents system permissions with hierarchical structure (id, parent_id, name, code, type, path, method, component, icon, sort, status, description, created_at, updated_at). Has many-to-many relationship with Roles.

- **APIKey**: Represents programmatic access keys with attributes (id, name, key, secret_hash, user_id, description, expires_at, status, last_used_at, created_at, updated_at). Belongs to a User.

- **UserRole**: Junction entity for User-Role many-to-many relationship (user_id, role_id).

- **RolePermission**: Junction entity for Role-Permission many-to-many relationship (role_id, permission_id).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All existing database operations complete with identical results (verified by running existing test suite or manual testing)
- **SC-002**: Code reduction of at least 30% in database access layers (measured by line count comparison in service and storage packages)
- **SC-003**: Database query performance matches or exceeds current implementation (no queries slower than 10ms compared to current raw SQL)
- **SC-004**: Service startup time remains under 5 seconds with GORM auto-migration enabled
- **SC-005**: All logs emit in structured JSON format with required fields (timestamp, level, message, context fields)
- **SC-006**: Log output performance is equivalent to or better than logrus (measured by logging benchmark tests)
- **SC-007**: OTLP integration allows logs to be viewable in Jaeger/VictoriaLogs within 1 second of emission
- **SC-008**: Zero data corruption or loss during migration (verified by database record count comparison before/after)
- **SC-009**: All existing API endpoints return identical responses (verified by API integration tests)
- **SC-010**: Docker Compose environment starts successfully with all services healthy within 30 seconds

### Assumptions

- PostgreSQL database version is 13+ (supports all GORM features)
- Redis is available for caching (already in use)
- Current database schema is compatible with GORM conventions (primary keys, foreign keys)
- Existing environment variable configuration structure can be extended for kart-io/logger settings
- OTLP collector endpoint is available in production environment (optional, can be disabled)
- All existing database records have valid data (no orphaned foreign keys)
- Service can tolerate brief downtime (5-10 minutes) during all-at-once deployment with GORM auto-migration and verification

### Dependencies

- GORM library: `gorm.io/gorm` and `gorm.io/driver/postgres`
- kart-io/logger library: `github.com/kart-io/logger` (must be compatible with Go 1.21+)
- Existing PostgreSQL database with current schema
- Existing Redis instance for caching
- OTLP collector (optional, for distributed tracing)

### Out of Scope

- Database schema changes (this is a code refactoring, not a feature addition)
- New API endpoints or functionality
- Performance optimization beyond GORM's built-in capabilities
- Migration of existing production data (data remains unchanged)
- Frontend changes (no UI impact)
- Authentication/authorization logic changes (only implementation changes)
