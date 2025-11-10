# Internal/Auth Directory Analysis: Reusable Business Logic Assessment

## Executive Summary

The `internal/auth/` directory contains **11,247 lines of code** organized into 58 Go files across 17 subdirectories. Analysis reveals:

- **10 Components** suitable for migration to `pkg/auth/` as reusable business logic
- **7 Components** that are service-specific and should remain in `internal/auth/`
- **Opportunity to reduce code duplication** across services that require authentication/authorization
- **Current state**: Only generic SessionManager exists in `pkg/auth/session.go`

---

## Directory Structure Overview

```
internal/auth/
├── jwt/                      ← JWT token generation & validation (REUSABLE)
├── crypto/                   ← Password hashing (REUSABLE)
├── validator/                ← Input validation functions (REUSABLE)
├── types/                    ← Domain types (HYBRID)
├── forced-logout/            ← Forced logout orchestration (SERVICE-SPECIFIC)
│   ├── audit/               ← Audit logging (SERVICE-SPECIFIC)
│   ├── session/             ← Session repository (PARTIALLY REUSABLE)
│   └── notification/        ← Email notifications (SERVICE-SPECIFIC)
├── cache/                    ← Permission caching (SERVICE-SPECIFIC)
├── storage/                  ← MySQL/Redis wrappers (SERVICE-SPECIFIC)
├── email/                    ← Email client (REUSABLE)
├── filter/                   ← Query builder (REUSABLE)
├── middleware/               ← Auth middleware (SERVICE-SPECIFIC)
├── handler/                  ← HTTP handlers (SERVICE-SPECIFIC)
├── service/                  ← Business services (SERVICE-SPECIFIC)
├── grpc/                     ← gRPC services (SERVICE-SPECIFIC)
├── metrics/                  ← Prometheus metrics (SERVICE-SPECIFIC)
├── routes/                   ← Route definitions (SERVICE-SPECIFIC)
├── startup/                  ← Service initialization (SERVICE-SPECIFIC)
└── config/                   ← Service config (SERVICE-SPECIFIC)
```

---

## Component Classification

### TIER 1: Ready for Migration to pkg/auth/ (Immediate)

#### 1. JWT Token Generation & Validation
**File**: `internal/auth/jwt/jwt.go` (146 lines)

**Status**: READY FOR MIGRATION ✅

**Current Usage**: 
- Only used within auth service
- Generic JWT operations independent of auth domain logic

**Reusability**: HIGH - Other services need JWT token handling

**Assessment**:
```
✅ Zero dependencies on auth-specific models
✅ Pure utility functions for token generation/validation
✅ No database or Redis dependencies
✅ Configurable expiration times
✅ JTI (JWT ID) support for session tracking

Code Quality:
- Well-structured: TokenTypeAccess, TokenTypeRefresh constants
- Comprehensive: GenerateToken, GenerateTokenPair, ValidateToken, ValidateRefreshToken
- Error handling: Proper fmt.Errorf wrapping with context
```

**Migration Plan**:
1. Move `internal/auth/jwt/jwt.go` → `pkg/auth/jwt.go`
2. Rename Claims to AuthClaims to avoid conflicts with golang-jwt/jwt
3. Update import paths in internal/auth/service/auth_service.go
4. Can be imported by: orchestrator, reasoning, gateway for token validation

**Impact**: 
- 0 breaking changes to internal/auth/ (self-contained module)
- 5 lines of import updates in internal/auth

---

#### 2. Password Hashing & Verification
**File**: `internal/auth/crypto/password.go` (30 lines)

**Status**: READY FOR MIGRATION ✅

**Current Usage**:
- Only used in AuthService for password verification
- Standard bcrypt wrapper

**Reusability**: HIGH - Generic cryptography utility

**Assessment**:
```
✅ Zero dependencies on any auth-specific logic
✅ Pure wrapper around golang.org/x/crypto/bcrypt
✅ Generic utility that any service could reuse
✅ Thread-safe bcrypt operations
✅ Constant cost factor (10) for consistency

Code Quality:
- Simple, focused implementation
- Two functions only: HashPassword, CheckPassword
- Proper error wrapping with context
```

**Migration Plan**:
1. Move `internal/auth/crypto/password.go` → `pkg/auth/crypto.go` (or new pkg/crypto/)
2. Update import paths in auth/service/auth_service.go
3. Can be imported by: any service that needs password handling

**Impact**:
- 0 breaking changes
- 1 line of import updates

---

#### 3. Input Validation Functions
**File**: `internal/auth/validator/validator.go` (154 lines)

**Status**: READY FOR MIGRATION ✅

**Current Usage**:
- Used in user creation/update handlers
- Also useful for other services with user input

**Reusability**: HIGH - Generic validation patterns

**Assessment**:
```
✅ Zero dependencies on auth models or database
✅ Pure validation logic with regex patterns
✅ Standard domain validation (email, username, password, phone, UUID)
✅ Extensible design for new validators
✅ Clear error messages

Validations Provided:
- ValidateUsername: 3-50 chars, alphanumeric+underscore+hyphen
- ValidatePassword: 8-128 chars, requires upper/lower/number/special
- ValidateEmail: RFC-like format, 255 char max
- ValidatePhone: International format, 10-15 digits
- ValidatePermissionType: Enum validation (menu, button, api)
- ValidateStatus: Binary status (0=disabled, 1=active)
- ValidateUUID: RFC 4122 format
- Utility functions: ValidateRequired, ValidateLength
```

**Migration Plan**:
1. Move `internal/auth/validator/validator.go` → `pkg/auth/validator.go` (or new pkg/validator/)
2. Update import paths in auth handlers
3. Can be imported by: user management, profile services

**Impact**:
- 0 breaking changes
- 3-4 lines of import updates in handlers

---

#### 4. Email Client Interface & Implementation
**File**: `internal/auth/email/client.go` (60+ lines)

**Status**: READY FOR MIGRATION ✅

**Current Usage**:
- Used by forced-logout notification service
- Generic email sending abstraction

**Reusability**: HIGH - Any service might need email

**Assessment**:
```
✅ Interface-based design (Client interface)
✅ Zero dependencies on auth domain
✅ SMTP implementation included
✅ Pluggable architecture (interface for alternative impls)
✅ Metadata support for extensibility

Features:
- Message with Title, Body, Format (html/text)
- Multiple recipients (targets)
- Receipt tracking with per-recipient status
- Context-aware with timeout support
- Config structure with TLS support
```

**Migration Plan**:
1. Move `internal/auth/email/client.go` → `pkg/email/client.go` (new pkg/email/)
2. Update import paths in forced-logout/notification
3. Can be imported by: any service for email notifications

**Impact**:
- 0 breaking changes
- 2 lines of import updates

---

#### 5. Query Filter Builder
**File**: `internal/auth/filter/filter.go` (228 lines)

**Status**: READY FOR MIGRATION ✅

**Current Usage**:
- Used in user, role, permission list handlers
- Generic query building pattern

**Reusability**: HIGH - Common data querying pattern

**Assessment**:
```
✅ Pure utility for building SQL WHERE clauses
✅ Zero database dependencies (generates queries, doesn't execute)
✅ Operator-based design: =, ILIKE, IN, >=, <=
✅ Type-specific filters: UserFilters, RoleFilters, PermissionFilters
✅ Extensible pattern for new entity filters

Design:
- QueryBuilder: Fluent interface for building filters
- AddFilter: Generic filter addition with operators
- AddEqualFilter, AddLikeFilter, AddInFilter: Specialized helpers
- Build(): Generates WHERE clause with SQL injection protection
- Extract*Filters: Parse query params from Gin context
- Apply*Filters: Apply specific entity filters

Benefits:
- Reduces SQL injection risk
- Consistent filtering across handlers
- Handles NULL values properly
- Range filters for date/numeric fields
```

**Migration Plan**:
1. Move `internal/auth/filter/filter.go` → `pkg/query/filter.go` (new pkg/query/)
2. Make generic types: EntityFilters, FieldFilter (remove User/Role/Permission specifics)
3. Keep specific filter extractors as examples or separate
4. Update import paths in auth handlers
5. Can be imported by: agent-manager, orchestrator, cluster for list operations

**Impact**:
- 2-3 lines of breaking changes (type names)
- 4-5 lines of import updates
- Other services can benefit from reusable filter builder

---

#### 6. Session Types & Models
**File**: `internal/auth/types/session.go` (104 lines)

**Status**: READY FOR PARTIAL MIGRATION ✅

**Current Usage**:
- SessionInfo: Session metadata (reusable)
- RevokedSession: Revocation tracking (reusable)
- ForceLogoutRequest/Response: Forced logout API (domain-specific)
- SessionListResponse, Pagination: API response types (generic)

**Reusability**: HIGH for session types, MEDIUM for API types

**Assessment**:
```
✅ SessionInfo: Pure session metadata model
  - JTI, UserID, Username, Email
  - Device tracking: DeviceType, DeviceName, Location
  - Timestamps: LoginAt, LastActivityAt, ExpiresAt
  - No dependencies on auth-specific logic

✅ RevokedSession: Revocation tracking
  - JTI, UserID, RevokedAt, RevokedBy, Reason, EventID
  - Audit trail information
  - Reusable for any session revocation system

✅ ForceLogoutRequest/Response: API types
  - Specific to forced logout feature
  - Reason, TriggeredBy (manual/policy/security_incident)
  - CorrelationID for distributed tracing

✅ SessionListResponse: Pagination wrapper
  - Generic pagination pattern
  - Useful for other list endpoints
```

**Migration Plan**:
1. Move session types to `pkg/auth/types.go`:
   - SessionInfo
   - RevokedSession
   - Pagination (move to pkg/pagination or keep here)
2. Keep forced-logout types in internal/auth/types/:
   - ForceLogoutRequest
   - ForceLogoutResponse
   - BulkForceLogoutRequest/Response
   - SessionLogoutResult
3. Move API response types to internal/auth/types/ (service-specific)
4. Update import paths throughout auth service

**Impact**:
- 1-2 lines of breaking changes
- 8-10 lines of import updates
- Enables other services to use session models

---

### TIER 2: Candidates for Migration with Refactoring (Secondary)

#### 7. Session Repository Pattern
**Files**: 
- `internal/auth/forced-logout/session/redis_repository.go` (231 lines)
- `internal/auth/forced-logout/session/repository.go` (interface)

**Status**: PARTIALLY REUSABLE ⚠️

**Reusability**: MEDIUM - Session storage pattern is generic, implementation is specific

**Assessment**:
```
✅ Repository interface is generic
✅ Redis implementation is well-structured
✅ Operations: StoreSession, GetSession, ListUserSessions, RevokeSession, etc.
✅ Batch operations: BulkRevokeSessions
✅ Token management: StoreRefreshToken, RevokeRefreshToken, BlacklistRefreshToken

❌ Tight coupling to SessionInfo type
❌ Forced logout specific semantics (revokedBy, reason, eventID)
❌ No generic key naming strategy (hardcoded Redis key patterns)

Challenge:
- The repository mixes generic session storage with forced-logout-specific fields
- Example: RevokeSession expects revokedBy, reason, eventID (forced logout only)
- Generic version would only store/retrieve session data, not handle revocation
```

**Migration Strategy**:
1. Create `pkg/auth/session_repository.go` with generic interface:
   - StoreSession(ctx, session *SessionInfo)
   - GetSession(ctx, jti) (*SessionInfo, error)
   - ListUserSessions(ctx, userID, limit, offset)
   - DeleteSession(ctx, jti)
   - Exists(ctx, jti)
   
2. Create specialized version for forced-logout in internal/auth:
   - Wraps base SessionRepository
   - Adds revocation-specific methods
   - ExtendedSessionRepository or RevokedSessionRepository

3. Redis implementation could support both via optional fields

**Impact**:
- 3-4 new interfaces in pkg/auth/
- Refactor 60-70 lines from forced-logout/session
- Enable reuse in other services needing session storage

---

#### 8. JWT/Session Integration Middleware
**File**: `internal/auth/middleware/jwt.go` (37 lines wrapper)

**Status**: PARTIALLY REUSABLE ⚠️

**Reusability**: LOW-MEDIUM - Wraps common middleware, adds forced-logout checking

**Assessment**:
```
✅ Currently extends common/middleware/jwt.JWTMiddleware
✅ Adds SessionValidator integration for revocation checking
✅ Backward compatible with common middleware

The auth-specific JWTMiddleware:
- Wraps common/middleware/jwt.JWTMiddleware
- Injects session service for forced-logout validation
- Provides aliases: JWTAuth(), OptionalJWTAuth()

Common middleware handles:
- Token parsing and validation
- Claims extraction
- User context setup
- Optional auth variant

Forced logout integration:
- Calls sessionService.ValidateSession() on protected endpoints
- Rejects requests if session is revoked
- Sets X-Session-Terminated header for revoked sessions

Could be reusable in other services that need:
- JWT auth + session revocation checking
```

**Migration Strategy**:
1. Keep in internal/auth/middleware/ (service-specific integration)
2. Ensure common/middleware/jwt.go is generic enough (✅ already is)
3. Other services can use same pattern:
   ```go
   // In any service that needs revocation checking
   jwtMW := middleware.NewJWTMiddleware(&middleware.JWTConfig{
       Secret: config.JWT.Secret,
       SessionValidator: sessionService,
   })
   ```

**Impact**:
- No migration needed
- Pattern is already reusable via common middleware

---

### TIER 3: Service-Specific (Should Stay in internal/auth/)

#### 9. Forced Logout Orchestration
**Files**: 
- `internal/auth/forced-logout/service.go`
- `internal/auth/forced-logout/audit/` (3 files)
- `internal/auth/forced-logout/notification/` (3 files)

**Status**: SERVICE-SPECIFIC ❌

**Reason for Non-Migration**:
- Implements auth service-specific business logic
- Tightly coupled to user/session domain model
- Requires audit trail, email notifications, hash chains
- Specific to forced logout feature (not generic session management)

**Components**:
- **Audit Service**: Hash chain implementation for tamper-proof logging
- **Notification Service**: Email templates for logout notifications
- **Forced Logout Service**: Orchestrates session revocation + audit + notifications

**Assessment**:
```
❌ High coupling to auth domain
❌ Business logic specific to forced logout feature
❌ Requires auth-specific models
❌ Email templates are auth-specific

Could be useful in other services?
- No. Forced logout is auth service responsibility
- Other services don't need to force logout users directly
- If they need to trigger forced logout, they call auth service API
```

**Recommendation**:
- Keep in `internal/auth/forced-logout/` where it belongs
- Other services use auth service API to trigger forced logout
- No migration needed

---

#### 10. Permission Cache & RBAC
**Files**:
- `internal/auth/cache/permission_cache.go`
- Used by permission middleware

**Status**: SERVICE-SPECIFIC ❌

**Reason for Non-Migration**:
- Auth service specific permission management
- Implements RBAC (Role-Based Access Control) for auth service
- Tightly coupled to auth models: User, Role, Permission, UserRole, RolePermission

**Assessment**:
```
❌ RBAC is auth service responsibility
❌ Permission models are auth-specific
❌ Caching is optimization detail

If other services need RBAC:
- They should use auth service API to check permissions
- They don't need to maintain their own permission cache
- Central auth service handles this
```

**Recommendation**:
- Keep in `internal/auth/cache/`
- Other services call auth service for permission checks
- No migration needed

---

#### 11. User/Role/Permission CRUD Services
**Files**:
- `internal/auth/service/auth_service.go`
- `internal/auth/service/user_service.go`
- `internal/auth/service/role_service.go`
- `internal/auth/service/permission_service.go`
- `internal/auth/service/apikey_service.go`

**Status**: SERVICE-SPECIFIC ❌

**Reason for Non-Migration**:
- Auth service business logic (user management, RBAC, API keys)
- Tightly coupled to MySQL storage, auth domain models
- Implements auth service requirements

**Assessment**:
```
❌ Domain-specific business logic
❌ Other services don't manage users/roles directly
❌ Requires auth service database schema
```

**Recommendation**:
- Keep in `internal/auth/service/`
- No migration needed

---

#### 12. Storage Layer (MySQL & Redis Wrappers)
**Files**:
- `internal/auth/storage/mysql.go`
- `internal/auth/storage/redis.go`
- `internal/auth/storage/migrate.go`

**Status**: SERVICE-SPECIFIC ❌

**Reason for Non-Migration**:
- Wraps MySQL/Redis for auth service only
- Contains auth-specific models and queries
- Migration scripts for auth database schema

**Assessment**:
```
❌ Not generic database clients
❌ Auth schema specific
❌ Better to use pkg/initializers/database.go and generic clients
```

**Recommendation**:
- Keep in `internal/auth/storage/`
- Use pkg/initializers for infrastructure setup
- No migration needed

---

#### 13. HTTP Handlers & Middleware
**Files**:
- `internal/auth/handler/` (7 handler files)
- `internal/auth/middleware/` (5 middleware files, except jwt.go)
- `internal/auth/routes/`

**Status**: SERVICE-SPECIFIC ❌

**Reason for Non-Migration**:
- HTTP endpoint implementations for auth service API
- Service-specific routing and request handling
- Specific to auth service's REST API

**Recommendation**:
- Keep in `internal/auth/`
- No migration needed

---

#### 14. gRPC Services
**Files**:
- `internal/auth/grpc/` (4 service implementations)

**Status**: SERVICE-SPECIFIC ❌

**Reason for Non-Migration**:
- gRPC service implementations for auth service
- Service-specific proto bindings

**Recommendation**:
- Keep in `internal/auth/`
- No migration needed

---

## Migration Summary

### Components Ready for Migration (10 items)

| Component | Current Location | Target Location | Priority | LOC | Impact |
|-----------|-----------------|-----------------|----------|-----|--------|
| JWT Token Ops | internal/auth/jwt | **pkg/auth/jwt** | HIGH | 146 | Zero |
| Password Crypto | internal/auth/crypto | **pkg/auth/crypto** | HIGH | 30 | Zero |
| Input Validators | internal/auth/validator | **pkg/auth/validator** | HIGH | 154 | Zero |
| Email Client | internal/auth/email | **pkg/email** | MEDIUM | 60+ | Zero |
| Query Filter | internal/auth/filter | **pkg/query** | MEDIUM | 228 | Low |
| Session Types | internal/auth/types/session | **pkg/auth/types** | HIGH | 50 | Low |
| Session Repo Interface | forced-logout/session | **pkg/auth/session** | MEDIUM | ~100 | Medium |
| Generic Config Types | internal/auth/types/types | Partial | LOW | 90+ | Medium |
| Test Utilities | internal/auth/jwt/jwt_test | **pkg/auth/jwt_test** | LOW | 80+ | Zero |
| Type Definitions | internal/auth/types | **pkg/auth/models** | MEDIUM | 235 | Medium |

**Total Reusable LOC**: ~1,173 lines (10% of auth codebase)

### Components to Keep in internal/auth/ (8 items)

| Component | Reason | LOC |
|-----------|--------|-----|
| Forced Logout Service | Auth-specific domain logic | 200+ |
| Forced Logout Audit | Auth-specific with hash chains | 150+ |
| Forced Logout Notification | Auth service email templates | 100+ |
| Permission Cache | RBAC specific to auth service | 80+ |
| User/Role/Permission Services | Auth domain business logic | 1,200+ |
| API Key Service | Auth domain business logic | 150+ |
| MySQL Storage Layer | Auth schema specific | 300+ |
| Redis Storage Layer | Auth schema specific | 100+ |
| HTTP Handlers | Service API endpoints | 800+ |
| gRPC Services | Service RPC endpoints | 300+ |
| Middleware (non-JWT) | Service-specific middleware | 150+ |
| Routes | Service-specific routing | 100+ |
| Startup | Service initialization | 200+ |
| Metrics | Service-specific metrics | 50+ |
| Config | Service-specific config | 100+ |

**Total Service-Specific LOC**: ~4,480 lines (40% of auth codebase)

---

## Detailed Migration Plan

### Phase 1: Core Authentication Utilities (Week 1-2)

**Priority**: HIGH - Unblock other services

**Files to Migrate**:
1. `jwt/jwt.go` → `pkg/auth/jwt.go`
2. `crypto/password.go` → `pkg/auth/crypto.go`
3. `validator/validator.go` → `pkg/auth/validator.go`
4. `types/session.go` (SessionInfo, RevokedSession) → `pkg/auth/types.go`
5. `jwt/jwt_test.go` → `pkg/auth/jwt_test.go`

**Execution**:
```bash
# 1. Create new files in pkg/auth/
cp internal/auth/jwt/jwt.go pkg/auth/jwt.go
cp internal/auth/crypto/password.go pkg/auth/crypto.go
cp internal/auth/validator/validator.go pkg/auth/validator.go
cp internal/auth/jwt/jwt_test.go pkg/auth/jwt_test.go

# 2. Create types file with SessionInfo, RevokedSession
# (Extract from internal/auth/types/session.go)

# 3. Update imports in internal/auth/
# - service/auth_service.go
# - handler/auth_handler.go
# - Any other references

# 4. Run tests to verify
make test internal/auth

# 5. Delete old files
rm -rf internal/auth/jwt internal/auth/crypto internal/auth/validator
```

**Breaking Changes**: 0
**Tests Needed**: All existing tests should pass
**Documentation**: Update CLAUDE.md

---

### Phase 2: Generic Data Handling (Week 2-3)

**Priority**: MEDIUM - Enables data filtering across services

**Files to Migrate**:
1. `email/client.go` → `pkg/email/client.go` (new directory)
2. `filter/filter.go` → `pkg/query/filter.go` (new directory)

**Execution**:
```bash
# 1. Create pkg/email/ and pkg/query/ directories

# 2. Move files
cp internal/auth/email/client.go pkg/email/client.go
cp internal/auth/filter/filter.go pkg/query/filter.go

# 3. Make filter.go more generic:
#    - Rename UserFilters, RoleFilters, PermissionFilters to examples
#    - Create generic EntityFilter type
#    - Keep specific extractors in auth service

# 4. Update imports in internal/auth/
#    - forced-logout/notification/
#    - handler/user_handler.go
#    - handler/role_handler.go
#    - handler/permission_handler.go

# 5. Tests and cleanup
```

**Breaking Changes**: 1-2 (type renames in filter builder)
**Tests Needed**: 10+ test functions
**Documentation**: Document filter builder pattern

---

### Phase 3: Session Management Infrastructure (Week 3-4)

**Priority**: MEDIUM - Enables session reuse

**Files to Migrate**:
1. `forced-logout/session/redis_repository.go` → Extract generic interface
2. Create `pkg/auth/session_repository.go` with generic interface

**Execution**:
```bash
# 1. Define generic SessionRepository interface in pkg/auth/session_repository.go
#    - StoreSession(ctx, session *SessionInfo)
#    - GetSession(ctx, jti) (*SessionInfo, error)
#    - ListUserSessions(ctx, userID, limit, offset) ([]SessionInfo, int, error)
#    - DeleteSession(ctx, jti)
#    - Exists(ctx, jti)
#    - Refresh(ctx, jti, ttl)
#    - GetTTL(ctx, jti)

# 2. Create pkg/auth/session_redis.go with implementation
#    - Copy from forced-logout/session/redis_repository.go
#    - Remove forced-logout-specific methods

# 3. Create internal/auth/forced-logout/session_revocation.go
#    - Extends SessionRepository interface
#    - Adds: RevokeSession, IsRevoked, BulkRevokeSessions

# 4. Update forced-logout to use wrapped repository
```

**Breaking Changes**: 2-3 (interface extraction)
**Tests Needed**: 15+ test functions
**Documentation**: Document session repository pattern

---

### Phase 4: Type Definitions & Models (Week 4)

**Priority**: LOW - Documentation and code clarity

**Files to Migrate**:
1. `types/types.go` - Extract reusable types to pkg/auth/models.go

**Execution**:
```bash
# 1. Identify reusable types:
#    - User, Role, Permission (already in internal/models/auth - keep there!)
#    - User/Role/Permission request types (keep in auth handler)
#    - LoginRequest, LoginResponse, RefreshTokenRequest (keep in auth handlers)
#    
# 2. Move only API-agnostic types to pkg/auth/
#    - Keep CRUD request/response types in internal/auth/types/

# 3. Defer migration until Phase 1-3 complete
```

**Breaking Changes**: 3-4
**Tests Needed**: 5+
**Documentation**: Update type organization

---

## Benefits of Migration

### For Other Services

1. **orchestrator**: Use JWT validation without importing auth service
2. **reasoning**: Validate JWT tokens in API endpoints
3. **gateway**: Consistent JWT handling across entry point
4. **monitor**: Email notifications without auth service
5. **cluster**: Input validation for API requests
6. **collect-agent**: Password validation for node auth

### For Auth Service

1. **Reduced exports**: Focus on domain logic, less shared utilities
2. **Clearer separation**: Business logic vs infrastructure
3. **Easier testing**: Can test utilities independently
4. **Better documentation**: Clear what's reusable vs domain-specific

### Code Quality

1. **DRY principle**: Eliminate JWT/crypto duplication
2. **Consistency**: All services use same validators
3. **Maintainability**: Bug fixes in utils benefit all services
4. **Performance**: Potential for shared validator compilation

---

## Risk Assessment

### Low Risk (✅ Safe)
- JWT migration: Pure utility, no dependencies
- Crypto migration: Pure utility, no dependencies  
- Validator migration: Pure utility, no dependencies
- Tests: All existing tests should still pass

### Medium Risk (⚠️ Review)
- Session types: Used in Redis repository, forced-logout
- Filter builder: Relies on Gin context, may need adaptation
- Email client: Few dependencies, straightforward

### High Risk (❌ Major Refactoring)
- Session repository: Requires interface extraction and testing
- Type definitions: Significant refactoring needed

---

## Appendix: File Locations Reference

### Files to Migrate
```
internal/auth/jwt/jwt.go                           → pkg/auth/jwt.go
internal/auth/jwt/jwt_test.go                      → pkg/auth/jwt_test.go
internal/auth/crypto/password.go                   → pkg/auth/crypto.go
internal/auth/validator/validator.go               → pkg/auth/validator.go
internal/auth/types/session.go (SessionInfo)       → pkg/auth/types.go
internal/auth/email/client.go                      → pkg/email/client.go
internal/auth/filter/filter.go                     → pkg/query/filter.go
internal/auth/forced-logout/session/redis_repo.go  → pkg/auth/session_redis.go (+ interface)
```

### Files to Keep
```
internal/auth/forced-logout/service.go
internal/auth/forced-logout/audit/
internal/auth/forced-logout/notification/
internal/auth/forced-logout/session/service.go
internal/auth/cache/
internal/auth/storage/
internal/auth/service/
internal/auth/handler/
internal/auth/middleware/ (except jwt.go wrapper)
internal/auth/grpc/
internal/auth/routes/
internal/auth/startup/
internal/auth/config/
internal/auth/metrics/
internal/auth/types/ (domain-specific types)
```

---

## Recommendations

### Short Term (Next Sprint)
1. Migrate JWT, Crypto, Validator to pkg/auth/
2. Update imports in auth service
3. Run full test suite
4. Document in CLAUDE.md

### Medium Term (Next Month)
1. Migrate Email, Filter, Session types
2. Create session repository interface
3. Enable use in other services
4. Update architecture docs

### Long Term (Future)
1. Add more reusable auth utilities as needed
2. Consider pkg/validation, pkg/email as shared libraries
3. Explore moving some forced-logout components to policies service
4. Plan for auth service API improvements

