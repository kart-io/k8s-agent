# Implementation Tasks: Auth Service

**Feature**: Authentication and Authorization Service
**Branch**: `001-auth-service-a`
**Date**: 2025-10-09
**Status**: Ready for Implementation

## Overview

This document provides a complete, ordered task list for implementing the auth-service. Tasks are organized by user story to enable incremental, independently testable delivery.

## Task Summary

- **Total Tasks**: 47
- **Estimated Complexity**: High (Enterprise RBAC system)
- **Recommended Approach**: Implement by user story (MVP = US1 only)
- **Parallel Opportunities**: 15+ parallelizable tasks identified

## User Stories (From README and Design Docs)

Based on the existing documentation, the following user stories have been inferred:

### US1 (Priority P1): User Authentication
**Goal**: Users can securely log in, log out, and access their profile information
**Value**: Core authentication functionality - blocking for all other features
**Independent Test**: Login with admin/admin123, get JWT token, retrieve user info, logout

### US2 (Priority P2): User Management
**Goal**: Administrators can create, read, update, and delete user accounts
**Value**: Essential for system administration
**Independent Test**: Create new user, list users, update user details, deactivate user

### US3 (Priority P2): Role & Permission Management
**Goal**: Administrators can manage roles and permissions (RBAC)
**Value**: Fine-grained access control
**Independent Test**: Create role, assign permissions, verify user menu based on permissions

### US4 (Priority P3): API Key Management
**Goal**: System administrators can generate API keys for service-to-service authentication
**Value**: Enables microservice integration
**Independent Test**: Generate API key, authenticate using key, revoke key

---

## Phase 1: Project Setup & Infrastructure

These tasks initialize the project structure and configure shared dependencies.

### T001: Initialize Go Module and Dependencies [P]
**File**: `auth-service/go.mod`
**Description**: Initialize Go module and add all required dependencies

```bash
cd auth-service
go mod init github.com/kart-io/k8s-agent/auth-service
go get github.com/gin-gonic/gin@v1.9.1
go get github.com/golang-jwt/jwt/v5@v5.2.0
go get github.com/google/uuid@v1.4.0
go get github.com/lib/pq@v1.10.9
go get github.com/redis/go-redis/v9@v9.3.0
go get github.com/sirupsen/logrus@v1.9.3
go get golang.org/x/crypto@v0.14.0
go get gopkg.in/yaml.v3@v3.0.1
go get github.com/spf13/viper
go mod tidy
```

**Validation**: `go mod verify` succeeds

### T002: Create Project Directory Structure [P]
**File**: Multiple directories
**Description**: Create the complete directory structure per plan.md

```bash
mkdir -p cmd/server
mkdir -p internal/{handler,middleware,model,service,storage}
mkdir -p pkg/types
mkdir -p configs
mkdir -p tests/{integration,unit}
mkdir -p scripts
```

**Validation**: All directories exist

### T003: Create Configuration File Template [P]
**File**: `auth-service/configs/config.yaml`
**Description**: Create configuration file with all required settings

```yaml
server:
  port: 8090
  mode: debug

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: k8s_agent_auth
  sslmode: disable
  max_open_conns: 25
  max_idle_conns: 5

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0
  pool_size: 10

jwt:
  secret: "change-this-in-production-min-32-chars"
  expires_hours: 24

logging:
  level: info
  format: json
```

**Validation**: File exists and is valid YAML

### T004: Create Makefile for Build Automation [P]
**File**: `auth-service/Makefile`
**Description**: Create Makefile with common development commands

Include targets: `build`, `run`, `test`, `clean`, `docker-build`, `deps`, `init-db`

**Validation**: `make help` shows all targets

---

## Phase 2: Foundational Infrastructure

These tasks MUST complete before any user story implementation can begin.

### T005: Implement Shared Type Definitions
**File**: `auth-service/pkg/types/types.go`
**Description**: Define all request/response types and domain models

Include types for:
- User, Role, Permission, APIKey structs
- LoginRequest, LoginResponse
- UserCreateRequest, UserUpdateRequest
- RoleRequest, PermissionRequest
- PaginationParams, PaginatedResponse
- ErrorResponse

**Validation**: `go build ./pkg/types` succeeds

### T006: Implement PostgreSQL Connection Management
**File**: `auth-service/internal/storage/postgres.go`
**Description**: Implement database connection with connection pooling

Functions needed:
- `NewPostgresDB()` - Create connection with config
- `Close()` - Graceful shutdown
- Connection pool configuration
- Health check function

**Validation**: Connection test succeeds with test database

### T007: Implement Redis Connection Management
**File**: `auth-service/internal/storage/redis.go`
**Description**: Implement Redis client with connection pooling

Functions needed:
- `NewRedisClient()` - Create client with config
- `Close()` - Graceful shutdown
- `Ping()` - Health check
- Helper functions for token blacklist operations

**Validation**: Redis ping succeeds

### T008: Implement Database Migration System
**File**: `auth-service/internal/storage/migrate.go`
**Description**: Implement database schema creation and seed data

Functions needed:
- `AutoMigrate(db)` - Create all tables with proper indexes
- `Seed(db)` - Insert default admin user and roles
- Table definitions from data-model.md:
  - users, roles, permissions
  - user_roles, role_permissions
  - api_keys

**Seed Data**:
- Admin user: username=admin, password=admin123 (bcrypt hashed)
- Roles: super_admin, admin, user
- Default permissions for system management

**Validation**: Tables created, seed data inserted, foreign keys working

### T009: Implement Database Models (GORM)
**File**: `auth-service/internal/model/user.go`, `role.go`, `permission.go`, `api_key.go`
**Description**: Define GORM models matching database schema from data-model.md

Each model should include:
- Struct tags for GORM
- Table name specification
- Validation tags
- Relationship definitions

**Validation**: Models compile without errors

### T010: Implement JWT Utility Functions
**File**: `auth-service/pkg/jwt/jwt.go`
**Description**: JWT token generation and validation utilities

Functions needed:
- `GenerateToken(userID, username string) (string, time.Time, error)`
- `ValidateToken(tokenString string) (*Claims, error)`
- Custom Claims struct with user info

**Validation**: Unit test for token generation and validation

### T011: Implement Password Hashing Utilities
**File**: `auth-service/pkg/crypto/password.go`
**Description**: bcrypt password hashing and validation

Functions needed:
- `HashPassword(password string) (string, error)` - bcrypt with cost 10
- `CheckPassword(hashedPassword, password string) error`

**Validation**: Unit test for hashing and validation

### T012: Implement Configuration Loading
**File**: `auth-service/internal/config/config.go`
**Description**: Configuration loading using Viper

Functions needed:
- `Load() (*Config, error)` - Load from config.yaml
- Support for environment variable overrides
- Validation of required fields

**Validation**: Config loads successfully in development

### T013: Implement Logging Setup
**File**: `auth-service/pkg/logger/logger.go`
**Description**: Initialize structured logger (Logrus)

Features:
- JSON format for production
- Log level configuration
- Request ID middleware support

**Validation**: Logger outputs structured logs

---

## Phase 3: User Story 1 - User Authentication (P1)

**Goal**: Implement core authentication (login/logout/user info)
**Independent Test**: Complete authentication flow from login to logout
**[Story]: US1**

### T014: [US1] Implement Auth Service - Login Logic
**File**: `auth-service/internal/service/auth_service.go`
**Description**: Implement authentication business logic

Methods needed:
- `Login(username, password) (*LoginResponse, error)`
  - Validate credentials
  - Generate JWT token
  - Return user with roles
- `Logout(token string) error`
  - Add token to Redis blacklist
- `GetCurrentUser(userID) (*User, error)`
- `GetUserMenus(userID) ([]MenuItem, error)`
  - Build hierarchical menu from permissions

**Validation**: Service methods work with test database

### T015: [US1] Implement JWT Authentication Middleware
**File**: `auth-service/internal/middleware/jwt.go`
**Description**: Gin middleware for JWT validation

Features:
- Extract token from Authorization header
- Validate token signature and expiration
- Check Redis blacklist
- Set user context for handlers
- Return 401 for invalid/expired tokens

**Validation**: Middleware correctly validates/rejects tokens

### T016: [US1] Implement CORS Middleware [P]
**File**: `auth-service/internal/middleware/cors.go`
**Description**: CORS handling for frontend integration

**Validation**: CORS headers set correctly

### T017: [US1] Implement Auth HTTP Handlers
**File**: `auth-service/internal/handler/auth_handler.go`
**Description**: HTTP endpoints for authentication

Endpoints:
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/logout` - User logout (requires JWT)
- `GET /api/v1/auth/me` - Current user info (requires JWT)
- `GET /api/v1/auth/menus` - User menu tree (requires JWT)
- `POST /api/v1/auth/check` - Permission check (public)

Include:
- Request validation
- Error handling
- HTTP status codes per contracts/auth.yaml

**Validation**: All endpoints return correct responses

### T018: [US1] Implement Main Application Entry Point
**File**: `auth-service/cmd/server/main.go`
**Description**: Application bootstrap and HTTP server setup

Features:
- Load configuration
- Initialize logger
- Connect to databases
- Run migrations
- Setup Gin routes
- Graceful shutdown

Routes setup:
- Public: `/api/v1/auth/login`, `/api/v1/auth/check`
- Protected: All other endpoints

**Validation**: Server starts and accepts requests

### T019: [US1] Create Dockerfile [P]
**File**: `auth-service/Dockerfile`
**Description**: Multi-stage Docker build

**Validation**: Docker image builds and runs

### T020: [US1] Integration Test - Authentication Flow
**File**: `auth-service/tests/integration/auth_test.go`
**Description**: End-to-end test for US1

Test scenarios:
1. Successful login with admin/admin123
2. Failed login with wrong password
3. Get current user info with valid token
4. Get user menus with valid token
5. Logout and verify token blacklisted
6. Access protected endpoint with expired/invalid token

**Validation**: All scenarios pass

**✓ CHECKPOINT: US1 Complete** - MVP ready for deployment

---

## Phase 4: User Story 2 - User Management (P2)

**Goal**: Full CRUD operations for user accounts
**Independent Test**: Admin can manage user lifecycle
**[Story]: US2**

### T021: [US2] Implement User Service - CRUD Operations
**File**: `auth-service/internal/service/user_service.go`
**Description**: User management business logic

Methods:
- `List(pagination, filters) (*PaginatedResponse, error)`
- `GetByID(id) (*User, error)`
- `Create(req *UserCreateRequest) (*User, error)`
  - Hash password before saving
  - Assign default role
- `Update(id, req *UserUpdateRequest) error`
  - Don't update password here (separate endpoint)
- `Delete(id) error` - Soft delete (set status=0)
- `AssignRoles(userID, roleIDs) error`

**Validation**: All CRUD operations work with database

### T022: [US2] Implement Permission Check Middleware
**File**: `auth-service/internal/middleware/permission.go`
**Description**: Middleware for permission-based authorization

Functions:
- `RequirePermission(permissionCode string) gin.HandlerFunc`
- `RequireRole(roleCode string) gin.HandlerFunc`

Check user permissions from database/cache and return 403 if insufficient.

**Validation**: Correctly allows/denies based on permissions

### T023: [US2] Implement User HTTP Handlers
**File**: `auth-service/internal/handler/user_handler.go`
**Description**: User management endpoints

Endpoints:
- `GET /api/v1/users` - List users (pagination, filtering)
- `GET /api/v1/users/:id` - Get user by ID
- `POST /api/v1/users` - Create user (requires `user:create` permission)
- `PUT /api/v1/users/:id` - Update user (requires `user:update`)
- `DELETE /api/v1/users/:id` - Delete user (requires `user:delete`)

**Validation**: All endpoints work with proper permissions

### T024: [US2] Integration Test - User Management
**File**: `auth-service/tests/integration/user_test.go`
**Description**: End-to-end test for US2

Test scenarios:
1. List all users with pagination
2. Create new user
3. Get user by ID
4. Update user information
5. Deactivate user (soft delete)
6. Verify permission checks (403 without permission)

**Validation**: All scenarios pass

**✓ CHECKPOINT: US2 Complete** - User management operational

---

## Phase 5: User Story 3 - Role & Permission Management (P2)

**Goal**: Complete RBAC system with role and permission management
**Independent Test**: Admin can configure access control
**[Story]: US3**

### T025: [US3] Implement Role Service
**File**: `auth-service/internal/service/role_service.go`
**Description**: Role management business logic

Methods:
- `List() ([]Role, error)`
- `GetByID(id) (*Role, error)`
- `Create(req *RoleRequest) (*Role, error)`
- `Update(id, req *RoleRequest) error`
- `Delete(id) error`
  - Check if role is in use
  - Prevent deletion of system roles (super_admin, admin, user)
- `AssignPermissions(roleID, permissionIDs) error`
- `GetPermissions(roleID) ([]Permission, error)`

**Validation**: Role operations work correctly

### T026: [US3] Implement Permission Service
**File**: `auth-service/internal/service/permission_service.go`
**Description**: Permission management business logic

Methods:
- `List(filters) ([]Permission, error)`
- `GetTree() ([]PermissionNode, error)` - Build hierarchical tree
- `GetByID(id) (*Permission, error)`
- `Create(req *PermissionRequest) (*Permission, error)`
  - Validate parent exists
  - Validate type-specific fields
- `Update(id, req *PermissionRequest) error`
- `Delete(id) error`
  - Check if permission is assigned to roles
  - Cascade or prevent deletion

**Validation**: Permission operations and tree building work

### T027: [US3] Implement Role HTTP Handlers
**File**: `auth-service/internal/handler/role_handler.go`
**Description**: Role management endpoints

Endpoints:
- `GET /api/v1/roles` - List roles
- `GET /api/v1/roles/:id` - Get role by ID
- `POST /api/v1/roles` - Create role (requires `role:create`)
- `PUT /api/v1/roles/:id` - Update role (requires `role:update`)
- `DELETE /api/v1/roles/:id` - Delete role (requires `role:delete`)
- `POST /api/v1/roles/:id/permissions` - Assign permissions (requires `role:assign`)

**Validation**: All endpoints work with permissions

### T028: [US3] Implement Permission HTTP Handlers
**File**: `auth-service/internal/handler/permission_handler.go`
**Description**: Permission management endpoints

Endpoints:
- `GET /api/v1/permissions` - List permissions
- `GET /api/v1/permissions/tree` - Get permission tree
- `GET /api/v1/permissions/:id` - Get permission by ID
- `POST /api/v1/permissions` - Create permission (requires `permission:create`)
- `PUT /api/v1/permissions/:id` - Update permission (requires `permission:update`)
- `DELETE /api/v1/permissions/:id` - Delete permission (requires `permission:delete`)

**Validation**: All endpoints work correctly

### T029: [US3] Integration Test - RBAC Management
**File**: `auth-service/tests/integration/rbac_test.go`
**Description**: End-to-end test for US3

Test scenarios:
1. Create new role
2. Create menu, button, and API permissions
3. Assign permissions to role
4. Assign role to user
5. Verify user gets correct menu tree
6. Verify API access control works
7. Test permission tree hierarchy

**Validation**: Complete RBAC flow works

**✓ CHECKPOINT: US3 Complete** - Full RBAC system operational

---

## Phase 6: User Story 4 - API Key Management (P3)

**Goal**: Service-to-service authentication via API keys
**Independent Test**: Generate and use API key for authentication
**[Story]: US4**

### T030: [US4] Implement API Key Service
**File**: `auth-service/internal/service/apikey_service.go`
**Description**: API key management business logic

Methods:
- `List(userID) ([]APIKey, error)` - List user's keys (secret masked)
- `Create(req *APIKeyCreateRequest) (*APIKeyWithSecret, error)`
  - Generate random key (ak_xxxxx format)
  - Generate random secret (sk_xxxxx format)
  - Hash secret before storage
  - Return both key and secret (secret shown only once)
- `Delete(id) error`
- `Validate(key, secret) (*APIKey, error)`
  - Find by key
  - Verify secret hash
  - Check expiration
  - Update last_used_at
- `CleanupExpired() (int, error)` - Background task

**Validation**: API key generation and validation work

### T031: [US4] Implement API Key Authentication Middleware
**File**: `auth-service/internal/middleware/api_key.go`
**Description**: Middleware for API key authentication

Features:
- Extract key from `X-API-Key` and secret from `X-API-Secret` headers
- Validate against database
- Set user context
- Alternative to JWT for service accounts

**Validation**: Middleware validates API keys correctly

### T032: [US4] Implement API Key HTTP Handlers
**File**: `auth-service/internal/handler/apikey_handler.go`
**Description**: API key management endpoints

Endpoints:
- `GET /api/v1/api-keys` - List user's API keys (requires auth)
- `POST /api/v1/api-keys` - Create new API key (requires `apikey:create`)
- `DELETE /api/v1/api-keys/:id` - Delete API key (requires `apikey:delete`)

**Validation**: All endpoints work correctly

### T033: [US4] Integration Test - API Key Authentication
**File**: `auth-service/tests/integration/apikey_test.go`
**Description**: End-to-end test for US4

Test scenarios:
1. Create API key as authenticated user
2. Authenticate using API key headers
3. Access protected endpoint with API key
4. Verify expired keys are rejected
5. Delete API key and verify it no longer works

**Validation**: Complete API key flow works

**✓ CHECKPOINT: US4 Complete** - API key authentication ready

---

## Phase 7: Polish & Production Readiness

Tasks that enhance the system across all user stories.

### T034: Implement Request Logging Middleware [P]
**File**: `auth-service/internal/middleware/logging.go`
**Description**: Log all HTTP requests with structured fields

Log fields: method, path, status, latency, user_id, request_id

**Validation**: Logs show request details

### T035: Implement Error Response Standardization [P]
**File**: `auth-service/pkg/errors/errors.go`
**Description**: Consistent error response format

Standard format per contracts:
```json
{
  "error": "Error Type",
  "code": 400,
  "details": "Detailed message"
}
```

**Validation**: All endpoints return standard error format

### T036: Implement Health Check Endpoint [P]
**File**: `auth-service/internal/handler/health_handler.go`
**Description**: Health and readiness endpoints

Endpoints:
- `GET /health` - Basic health check
- `GET /health/ready` - Readiness check (DB + Redis)

**Validation**: Health endpoints work correctly

### T037: Add Input Validation [P]
**File**: Throughout handlers
**Description**: Comprehensive input validation using Gin binding

Validate:
- Required fields
- Field formats (email, phone)
- String lengths
- Enum values

**Validation**: Invalid inputs return 400 with clear messages

### T038: Implement Pagination Helper [P]
**File**: `auth-service/pkg/pagination/pagination.go`
**Description**: Reusable pagination logic

Features:
- Parse page, page_size from query params
- Calculate offset/limit
- Build paginated response

**Validation**: Pagination works across list endpoints

### T039: Add Query Filtering Support [P]
**File**: Throughout services
**Description**: Support filtering in list operations

Filters:
- Users: status, role
- Permissions: type, status

**Validation**: Filters work correctly

### T040: Implement Database Query Optimization [P]
**File**: Throughout services
**Description**: Add indexes and optimize queries

Optimizations:
- Add missing indexes from data-model.md
- Use joins instead of N+1 queries
- Implement pagination at database level
- Cache frequently accessed data

**Validation**: Query performance meets goals (<100ms)

### T041: Implement Redis Caching for Permissions [P]
**File**: `auth-service/internal/service/cache.go`
**Description**: Cache user permissions in Redis

Features:
- Cache user→permissions mapping (TTL: 15 minutes)
- Cache role→permissions mapping
- Invalidation on permission changes

**Validation**: Cache hit rate >80% for permission checks

### T042: Add Prometheus Metrics Export [P]
**File**: `auth-service/internal/metrics/metrics.go`
**Description**: Expose Prometheus metrics

Metrics:
- HTTP request duration histogram
- Request count by endpoint and status
- Active sessions gauge
- Authentication success/failure counters
- Database connection pool stats

Endpoint: `GET /metrics`

**Validation**: Prometheus can scrape metrics

### T043: Add OpenTelemetry Tracing [P]
**File**: `auth-service/internal/tracing/tracing.go`
**Description**: Distributed tracing with OpenTelemetry

Features:
- Trace HTTP requests
- Trace database operations
- Trace Redis operations
- Export to OTLP collector

**Validation**: Traces visible in Jaeger/Zipkin

### T044: Complete Remaining OpenAPI Specs [P]
**Files**: `specs/001-auth-service-a/contracts/{users,roles,permissions,api-keys}.yaml`
**Description**: Generate OpenAPI specs for remaining endpoints

Based on contracts/auth.yaml pattern, create specs for:
- User management endpoints
- Role management endpoints
- Permission management endpoints
- API key management endpoints

**Validation**: All specs are valid OpenAPI 3.0

### T045: Integrate Swagger UI [P]
**File**: `auth-service/cmd/server/main.go`
**Description**: Serve Swagger UI at /docs

Features:
- Load OpenAPI specs from contracts/
- Interactive API documentation
- Try-it-out functionality

**Validation**: `/docs` shows complete API documentation

### T046: Write Unit Tests [P]
**File**: `auth-service/tests/unit/*_test.go`
**Description**: Unit tests for services and utilities

Test coverage targets:
- Service layer: 80%+
- Utilities (JWT, crypto): 100%
- Middleware: 70%+

**Validation**: `go test -cover ./...` meets targets

### T047: Create Docker Compose for Development [P]
**File**: `auth-service/docker-compose.yml`
**Description**: Local development environment

Services:
- PostgreSQL with initialization
- Redis
- Auth service

**Validation**: `docker-compose up` starts all services

**✓ CHECKPOINT: Production Ready** - All enhancements complete

---

## Execution Strategy

### Recommended Implementation Order

1. **MVP (Weeks 1-2)**: Phase 1 + Phase 2 + Phase 3 (US1 only)
   - Delivers: Working authentication system
   - Value: Core security infrastructure
   - Testable: Complete login/logout flow

2. **Version 1.0 (Weeks 3-4)**: Phase 4 + Phase 5 (US2 + US3)
   - Delivers: Full RBAC system
   - Value: Complete access control
   - Testable: Admin can manage users and permissions

3. **Version 1.1 (Week 5)**: Phase 6 (US4)
   - Delivers: Service-to-service authentication
   - Value: Microservice integration
   - Testable: API keys work for automation

4. **Version 2.0 (Week 6)**: Phase 7 (Polish)
   - Delivers: Production-grade features
   - Value: Observability and performance
   - Testable: Metrics and tracing functional

### Parallel Execution Opportunities

**Phase 1 (All tasks can run in parallel)**:
- T001, T002, T003, T004

**Phase 2 (Groups that can run in parallel)**:
- Group A: T005 (types)
- Group B: T006, T007 (database connections)
- Group C: T010, T011, T012, T013 (utilities)
- Sequential: T008 (migrations - depends on T006), T009 (models - depends on T005)

**Phase 3 (US1)**:
- T014 (service) → T017, T018 (handlers, main)
- T015, T016 (middleware) - parallel
- T019 (Dockerfile) - parallel with T014-T018
- T020 (tests) - after all implementation

**Phase 7 (Most tasks can run in parallel)**:
- T034-T045 are largely independent

### Dependency Graph

```
Phase 1 (Setup)
  ↓
Phase 2 (Foundation)
  T005 → T009
  T006 → T008
  T007 (parallel)
  T010, T011, T012, T013 (parallel)
  ↓
Phase 3 (US1) - MVP
  T014 → T017 → T018 → T020
  T015, T016 → T018
  T019 (parallel)
  ↓
Phase 4 (US2)
  T021 → T023 → T024
  T022 → T023
  ↓
Phase 5 (US3)
  T025 → T027 → T029
  T026 → T028 → T029
  ↓
Phase 6 (US4)
  T030 → T032 → T033
  T031 → T032
  ↓
Phase 7 (Polish)
  T034-T047 (mostly parallel)
```

### Testing Strategy

**Per User Story Testing** (Integration tests after each story):
- US1 (T020): Authentication flow
- US2 (T024): User CRUD operations
- US3 (T029): RBAC configuration
- US4 (T033): API key authentication

**Final Testing** (Phase 7):
- Unit tests (T046): Service and utility coverage
- Performance tests: Load testing for 1000 concurrent requests
- Security tests: Penetration testing checklist

### Success Criteria by Phase

**Phase 3 (MVP) Success**:
- ✅ Users can log in and access protected endpoints
- ✅ JWT tokens work correctly
- ✅ Password hashing is secure
- ✅ Basic permission checks work

**Phase 5 (v1.0) Success**:
- ✅ Admins can create/manage users
- ✅ Complete RBAC with menu/button/API permissions
- ✅ User menus render based on permissions
- ✅ Permission checks enforce access control

**Phase 6 (v1.1) Success**:
- ✅ API keys can be generated and used
- ✅ Service-to-service auth works
- ✅ Key expiration and revocation work

**Phase 7 (v2.0) Success**:
- ✅ Prometheus metrics exported
- ✅ Distributed tracing operational
- ✅ API documentation complete
- ✅ Performance goals met (<100ms p95)
- ✅ Test coverage >80%

---

## Notes

- All file paths are relative to repository root
- Tasks marked [P] can be executed in parallel with other [P] tasks in the same phase
- Each user story is independently testable and deployable
- Integration tests should run after completing each user story phase
- Default credentials: admin/admin123 (change in production)
- JWT secret MUST be changed in production (min 32 characters)

---

**Generated**: 2025-10-09 by `/speckit.tasks` command
**Ready for**: `/speckit.implement` command to execute tasks
