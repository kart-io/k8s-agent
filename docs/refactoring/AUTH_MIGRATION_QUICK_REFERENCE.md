# Auth Migration Quick Reference Guide

This guide helps developers quickly find the new locations of migrated auth components.

## Quick Lookup Table

| What You Need | Old Location | New Location | Status |
|---------------|--------------|--------------|--------|
| **JWT Token Operations** | `internal/auth/jwt` | `pkg/auth/jwt` | ✅ Phase 1 |
| **Password Hashing** | `internal/auth/crypto` | `pkg/auth/crypto` | ✅ Phase 1 |
| **Input Validation** | `internal/auth/validator` | `pkg/auth/validator` | ✅ Phase 1 |
| **Session Types** | `internal/auth/types` | `pkg/auth/types` | ✅ Phase 1 |
| **Email Client** | `internal/auth/email` | `pkg/email` | ✅ Phase 2 |
| **Query Builder (Generic)** | `internal/auth/filter` | `pkg/query` | ✅ Phase 2 |
| **Auth Filter Helpers** | `internal/auth/filter` | `internal/auth/filter` | ✅ Phase 2 (refactored) |
| **Session Manager** | `common/storage/redis` | `pkg/auth` | ✅ Earlier |

---

## Import Path Updates

### Phase 1 Migrations (Completed 2025-11-10)

```go
// OLD imports
import (
    "github.com/kart-io/k8s-agent/internal/auth/jwt"
    "github.com/kart-io/k8s-agent/internal/auth/crypto"
    "github.com/kart-io/k8s-agent/internal/auth/validator"
)

// NEW imports
import (
    "github.com/kart-io/k8s-agent/pkg/auth/jwt"
    "github.com/kart-io/k8s-agent/pkg/auth/crypto"
    "github.com/kart-io/k8s-agent/pkg/auth/validator"
)
```

### Phase 2 Migrations (Completed 2025-11-10)

```go
// OLD email import
import "github.com/kart-io/k8s-agent/internal/auth/email"

// NEW email import
import "github.com/kart-io/k8s-agent/pkg/email"

// OLD query builder usage
import "github.com/kart-io/k8s-agent/internal/auth/filter"
qb := filter.NewQueryBuilder()

// NEW query builder usage
import "github.com/kart-io/k8s-agent/pkg/query"
qb := query.NewBuilder()

// Auth-specific filters (unchanged)
import "github.com/kart-io/k8s-agent/internal/auth/filter"
filters := filter.ExtractUserFilters(c)
```

---

## Common Use Cases

### 1. Generate JWT Token

```go
import "github.com/kart-io/k8s-agent/pkg/auth/jwt"

// Generate access token only
token, expiresAt, err := jwt.GenerateToken(userID, username, jwtSecret, 24)

// Generate access + refresh token pair
pair, err := jwt.GenerateTokenPair(userID, username, jwtSecret, 24)
// pair.AccessToken, pair.RefreshToken, pair.AccessExpiresAt, pair.RefreshExpiresAt
```

### 2. Validate JWT Token

```go
import "github.com/kart-io/k8s-agent/pkg/auth/jwt"

// Validate access token
claims, err := jwt.ValidateToken(tokenString, jwtSecret)
if err != nil {
    // Token invalid or expired
}

// Validate refresh token
claims, err := jwt.ValidateRefreshToken(refreshTokenString, jwtSecret)
```

### 3. Hash and Verify Password

```go
import "github.com/kart-io/k8s-agent/pkg/auth/crypto"

// Hash password during registration
hashedPassword, err := crypto.HashPassword(plainPassword)

// Verify password during login
err := crypto.CheckPassword(hashedPassword, plainPassword)
if err != nil {
    // Password incorrect
}
```

### 4. Validate User Input

```go
import "github.com/kart-io/k8s-agent/pkg/auth/validator"

// Validate username
if err := validator.ValidateUsername(username); err != nil {
    // Invalid username
}

// Validate password strength
if err := validator.ValidatePassword(password); err != nil {
    // Password too weak
}

// Validate email format
if err := validator.ValidateEmail(email); err != nil {
    // Invalid email
}

// Validate phone number
if err := validator.ValidatePhone(phone); err != nil {
    // Invalid phone number
}
```

### 5. Send Email Notification

```go
import "github.com/kart-io/k8s-agent/pkg/email"

// Create email client
config := &email.Config{
    Host:     "smtp.gmail.com",
    Port:     587,
    Username: "user@example.com",
    Password: "app-password",
    From:     "noreply@example.com",
    UseTLS:   true,
    Timeout:  30 * time.Second,
}
client, err := email.NewClient(config)

// Send email
msg := &email.Message{
    ID:     uuid.New().String(),
    Title:  "Account Security Alert",
    Body:   "<h1>Your account was logged out</h1>",
    Format: "html",
    Targets: []email.Target{
        {Type: "email", Value: "user@example.com", Platform: "email"},
    },
}

receipt, err := client.Send(ctx, msg)
if err != nil {
    // Failed to send
}

// Check delivery results
for _, result := range receipt.Results {
    if result.Success {
        log.Printf("Email sent to %s", result.Platform)
    } else {
        log.Printf("Failed to send to %s: %s", result.Platform, result.Error)
    }
}
```

### 6. Build SQL Query with Filters

```go
import (
    "github.com/kart-io/k8s-agent/pkg/query"
    "github.com/kart-io/k8s-agent/internal/auth/filter"
)

// Generic query builder (any service can use)
qb := query.NewBuilder()
qb.AddLikeFilter("name", "john")
qb.AddEqualFilter("status", 1)
qb.AddRangeFilter("created_at", startTime, endTime)
qb.AddInFilter("role", []string{"admin", "user"})

whereClause, args := qb.Build()
// whereClause: "name ILIKE $1 AND status = $2 AND created_at >= $3 AND created_at <= $4 AND role IN ($5, $6)"
// args: []interface{}{"%john%", 1, startTime, endTime, "admin", "user"}

db.Where(whereClause, args...).Find(&results)

// Auth-specific filter helpers (auth service only)
filters := filter.ExtractUserFilters(ginContext)
qb = query.NewBuilder()
filter.ApplyUserFilters(qb, filters)
whereClause, args = qb.Build()
db.Where(whereClause, args...).Find(&users)
```

### 7. Manage Sessions

```go
import (
    "github.com/kart-io/k8s-agent/pkg/auth"
    "github.com/kart-io/k8s-agent/pkg/auth/types"
)

// Create session manager
sessionMgr := auth.NewSessionManager(redisClient, logger, "myapp")

// Store session
sessionInfo := &types.SessionInfo{
    JTI:      jti,
    UserID:   userID,
    Username: username,
    Email:    email,
    // ... other fields
}
err := sessionMgr.Set(ctx, jti, sessionInfo, 24*time.Hour)

// Get session
var session types.SessionInfo
err := sessionMgr.Get(ctx, jti, &session)

// Delete session
err := sessionMgr.Delete(ctx, jti)

// Delete all user sessions (forced logout)
err := sessionMgr.DeleteAllUserSessions(ctx, userID)
```

---

## Package Purposes

### pkg/auth/ - Generic Auth Utilities

**When to use**: Any service needing authentication functionality

**What's included**:
- JWT token generation and validation
- Password hashing and verification
- Input validation (username, email, phone, password)
- Session management (Redis-based)
- Session types (SessionInfo, RevokedSession)

**What's NOT included**:
- User/Role/Permission domain models (stay in `internal/auth/types/`)
- Auth service business logic (stay in `internal/auth/service/`)
- HTTP handlers (stay in `internal/auth/handler/`)

### pkg/email/ - Generic Email Client

**When to use**: Any service needing to send emails

**What's included**:
- SMTP email client with TLS support
- HTML and plain text email support
- Multiple recipients per message
- Delivery receipt tracking
- No-op mode for testing

**Use cases**:
- Forced logout notifications (auth)
- Workflow completion alerts (orchestrator)
- Analysis reports (reasoning)
- System health alerts (monitor)

### pkg/query/ - Generic SQL Query Builder

**When to use**: Any service needing dynamic SQL filtering

**What's included**:
- WHERE clause builder with fluent API
- Support for =, IN, LIKE, >=, <= operators
- SQL injection protection
- PostgreSQL placeholders ($1, $2, etc.)

**Use cases**:
- User/role/permission filtering (auth)
- Cluster/node filtering (cluster service)
- Agent/event filtering (agent-manager)
- Workflow/execution filtering (orchestrator)

---

## Backward Compatibility

### Type Aliases (Maintained for Compatibility)

```go
// internal/auth/types/session.go
package types

import authtypes "github.com/kart-io/k8s-agent/pkg/auth/types"

// Re-export types for backward compatibility
type SessionInfo = authtypes.SessionInfo
type RevokedSession = authtypes.RevokedSession
```

**What this means**:
- Existing code using `internal/auth/types.SessionInfo` still works
- New code should use `pkg/auth/types.SessionInfo`
- No breaking changes during gradual migration

### Deprecated (Still Works)

These imports still work but are deprecated:

```go
// DEPRECATED: Use pkg/auth/jwt instead
import "github.com/kart-io/k8s-agent/internal/auth/jwt"

// DEPRECATED: Use pkg/auth/crypto instead
import "github.com/kart-io/k8s-agent/internal/auth/crypto"

// DEPRECATED: Use pkg/auth/validator instead
import "github.com/kart-io/k8s-agent/internal/auth/validator"

// REMOVED: Use pkg/email instead
import "github.com/kart-io/k8s-agent/internal/auth/email"  // ❌ No longer exists
```

---

## Migration Checklist for New Code

When writing new code:

- [ ] Use `pkg/auth/jwt` for JWT operations
- [ ] Use `pkg/auth/crypto` for password hashing
- [ ] Use `pkg/auth/validator` for input validation
- [ ] Use `pkg/auth/types` for session types
- [ ] Use `pkg/email` for email sending
- [ ] Use `pkg/query` for SQL query building
- [ ] Use `internal/auth/filter` for auth-specific filter helpers only
- [ ] Use `internal/auth/types` for auth domain models (User, Role, Permission)

---

## FAQs

### Q: Where is the User model?

**A**: `internal/auth/types/types.go` - User, Role, Permission are domain models, not generic utilities.

### Q: Where is the forced logout logic?

**A**: `internal/auth/forced-logout/` - This is domain-specific business logic for the auth service.

### Q: Can I use pkg/email in the orchestrator service?

**A**: Yes! That's the whole point of moving it to `pkg/`. It's now a generic email client.

### Q: Where is the old QueryBuilder?

**A**: Renamed to `query.Builder` in `pkg/query/`. Auth-specific helpers now use it internally.

### Q: Do I need to update my existing code?

**A**: Not immediately. Backward compatibility is maintained via re-exports. But for new code, use the new import paths.

### Q: Where can I find migration examples?

**A**: Check `internal/auth/service/auth_service.go` for updated import examples.

---

## Getting Help

**Documentation**:
- Phase 1 Report: `docs/refactoring/AUTH_TO_PKG_MIGRATION_PHASE1_REPORT.md`
- Phases 2-4 Report: `docs/refactoring/AUTH_TO_PKG_MIGRATION_PHASES2-4_REPORT.md`
- Code Reorganization Plan: `docs/CODE_REORGANIZATION.md`

**Package Documentation**:
- `pkg/auth/doc.go` - Auth utilities overview
- `pkg/email/doc.go` - Email client usage
- `pkg/query/doc.go` - Query builder examples

**Questions?** Check the migration reports or ask the team.

---

**Last Updated**: 2025-11-10
**Migration Status**: Phases 1-4 Complete ✅
