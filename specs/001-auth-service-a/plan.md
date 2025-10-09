# Implementation Plan: Auth Service

**Branch**: `001-auth-service-a` | **Date**: 2025-10-09 | **Spec**: [spec.md](./spec.md)
**Input**: Based on existing auth-service implementation at `/auth-service/README.md`

**Note**: This plan documents the existing auth-service implementation and guides future development.

## Summary

Authentication and authorization service providing user authentication, RBAC (Role-Based Access Control) permission management, and API authorization. The service implements JWT token-based authentication, user/role/permission management, and API Key authentication for service-to-service communication.

## Technical Context

**Language/Version**: Go 1.21
**Primary Dependencies**:
  - Gin (v1.9.1) - HTTP web framework
  - golang-jwt/jwt (v5.2.0) - JWT authentication
  - PostgreSQL driver (lib/pq v1.10.9) - Database access
  - Redis client (go-redis/v9 v9.3.0) - Token caching and blacklist
  - bcrypt (golang.org/x/crypto) - Password encryption
  - Logrus (v1.9.3) - Structured logging

**Storage**:
  - PostgreSQL - Primary data store (users, roles, permissions, API keys)
  - Redis - Session cache and token blacklist

**Testing**: Go testing framework (go test), NEEDS CLARIFICATION: integration test strategy
**Target Platform**: Linux server (containerized via Docker)
**Project Type**: Single microservice (RESTful API server)
**Performance Goals**:
  - 1000 concurrent requests
  - <100ms response time for authentication checks
  - Support 10k+ active users

**Constraints**:
  - JWT token expiry: 24 hours (configurable)
  - HTTPS required in production
  - Bcrypt for password hashing (security standard)

**Scale/Scope**:
  - Enterprise-level user management (10k-100k users)
  - Multi-tenant capable through role separation
  - API gateway integration ready

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Note**: Constitution template is not yet populated. Applying standard microservice development principles:

### Initial Check (Pre-Research)

| Principle | Status | Notes |
|-----------|--------|-------|
| Test-First Development | ⚠️ NEEDS ATTENTION | Integration tests not yet defined |
| Security Best Practices | ✅ PASS | JWT + bcrypt + HTTPS enforced |
| API Documentation | ⚠️ NEEDS ATTENTION | OpenAPI/Swagger spec needed |
| Observability | ⚠️ NEEDS ATTENTION | Structured logging present, metrics/tracing needed |
| Error Handling | ✅ PASS | Standard HTTP error codes |

**Action Items Before Implementation**:
1. Define integration test strategy
2. Generate OpenAPI specification
3. Add metrics and distributed tracing support

### Post-Design Check

| Principle | Status | Notes |
|-----------|--------|-------|
| Test-First Development | ✅ IMPROVED | Integration test strategy defined in research.md |
| Security Best Practices | ✅ PASS | JWT + bcrypt + HTTPS + API Key auth |
| API Documentation | ✅ IMPROVED | OpenAPI spec generated in contracts/ |
| Observability | ⚠️ PARTIAL | Logging present, metrics/tracing documented but not implemented |
| Error Handling | ✅ PASS | Standard HTTP error codes, comprehensive error responses |
| Data Model | ✅ PASS | Complete RBAC model documented in data-model.md |

**Remaining Action Items**:
1. Implement Prometheus metrics export
2. Implement OpenTelemetry tracing
3. Complete remaining OpenAPI specs (users, roles, permissions, api-keys)
4. Implement integration test suite based on strategy

## Project Structure

### Documentation (this feature)

```
specs/001-auth-service-a/
├── spec.md              # Feature specification
├── plan.md              # This file (implementation plan)
├── research.md          # Technology decisions and research
├── data-model.md        # Data entity definitions
├── quickstart.md        # Developer quickstart guide
├── contracts/           # API contracts (OpenAPI specs)
│   ├── auth.yaml        # Authentication endpoints
│   ├── users.yaml       # User management endpoints
│   ├── roles.yaml       # Role management endpoints
│   └── permissions.yaml # Permission management endpoints
└── tasks.md             # Implementation tasks
```

### Source Code (repository root)

```
auth-service/
├── cmd/
│   └── server/
│       └── main.go           # Application entry point
├── internal/
│   ├── handler/              # HTTP request handlers
│   │   ├── auth_handler.go
│   │   ├── user_handler.go
│   │   ├── role_handler.go
│   │   └── permission_handler.go
│   ├── middleware/           # HTTP middlewares
│   │   ├── jwt.go            # JWT authentication
│   │   ├── permission.go     # Permission check
│   │   ├── api_key.go        # API Key auth
│   │   └── cors.go           # CORS handling
│   ├── model/                # Database models (GORM)
│   │   ├── user.go
│   │   ├── role.go
│   │   ├── permission.go
│   │   └── api_key.go
│   ├── service/              # Business logic
│   │   ├── auth_service.go
│   │   ├── user_service.go
│   │   ├── role_service.go
│   │   └── permission_service.go
│   └── storage/              # Data access layer
│       ├── postgres.go       # PostgreSQL connection
│       ├── redis.go          # Redis connection
│       └── migrate.go        # Database migrations
├── pkg/
│   └── types/                # Shared type definitions
│       └── types.go
├── configs/
│   └── config.yaml           # Configuration file
├── tests/
│   ├── integration/          # Integration tests
│   └── unit/                 # Unit tests
├── scripts/                  # Utility scripts
├── Dockerfile
├── Makefile
├── go.mod
├── go.sum
├── README.md
└── QUICKSTART.md
```

**Structure Decision**: Single project structure chosen because auth-service is a standalone microservice with RESTful API. The `internal/` directory follows Go best practices for private packages, while `pkg/` contains reusable types that could be imported by other services.

## Complexity Tracking

*Fill ONLY if Constitution Check has violations that must be justified*

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| N/A | No constitutional violations identified | Constitution template not yet populated |

**Notes**: Once constitution is established, revisit this section to document any justified complexity additions.
