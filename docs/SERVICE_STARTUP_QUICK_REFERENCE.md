# Service Startup Pattern - Quick Reference

**TL;DR:** Use 1 flat Wire provider set + `GetInitializers()` method for auto-registration. Reduces boilerplate by 82%, maintains Wire + Bootstrap benefits.

## Pattern Structure (3 files)

### 1. wire.go - Flat Provider Set

```go
//go:build wireinject
// +build wireinject

package app

import (
    "github.com/google/wire"
    "github.com/kart-io/k8s-agent/internal/{service}/initializers"
    commonapp "github.com/kart-io/k8s-agent/pkg/app"
    pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
)

// ✅ Single flat set - Wire auto-resolves dependencies
var InitializerSet = wire.NewSet(
    ProvideLogger,
    initializers.NewDatabaseInitializer,
    initializers.NewRedisInitializer,
    initializers.NewServiceInitializer,
    initializers.NewHTTPServerInitializer,
    // ... add all initializer constructors
)

var HealthInitializerSet = wire.NewSet(
    pkginitializers.NewHealthCheckInitializer,
    wire.FieldsOf(new(*commonapp.StandardOptions), "Health"),
)

func Initialize{Service}Components(opts *commonapp.StandardOptions) (*{Service}Components, error) {
    wire.Build(InitializerSet, HealthInitializerSet, New{Service}Components)
    return nil, nil
}
```

**Key:** One flat `InitializerSet`, no nesting. Wire handles dependency order.

### 2. components.go - Private Fields + GetInitializers()

```go
package app

import (
    "github.com/kart-io/k8s-agent/internal/{service}/initializers"
    commonapp "github.com/kart-io/k8s-agent/pkg/app"
    "github.com/kart-io/k8s-agent/pkg/bootstrap"
    pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
    "github.com/kart-io/logger/core"
)

// ✅ Private fields (lowercase)
type {Service}Components struct {
    db      *initializers.DatabaseInitializer
    redis   *initializers.RedisInitializer
    service *initializers.ServiceInitializer
    http    *initializers.HTTPServerInitializer
    health  *pkginitializers.HealthCheckInitializer
}

// Wire calls this constructor
func New{Service}Components(
    db *initializers.DatabaseInitializer,
    redis *initializers.RedisInitializer,
    service *initializers.ServiceInitializer,
    http *initializers.HTTPServerInitializer,
    health *pkginitializers.HealthCheckInitializer,
) *{Service}Components {
    return &{Service}Components{
        db: db, redis: redis, service: service, http: http, health: health,
    }
}

// ✅ GetInitializers() - Only public method
func (c *{Service}Components) GetInitializers() []bootstrap.Initializer {
    return []bootstrap.Initializer{
        c.db,      // Priority 300 - Database
        c.redis,   // Priority 400 - Cache
        c.service, // Priority 600 - Business services (CRITICAL: before servers)
        c.http,    // Priority 1000 - HTTP server
        c.health,  // Priority 2000 - Health checks
    }
}

func ProvideLogger(opts *commonapp.StandardOptions) (core.Logger, error) {
    return opts.InitLogger()
}
```

**Key:** Private fields + `GetInitializers()` method with documented priorities.

### 3. app.go - Minimal App Struct + Auto-Registration

```go
package app

import (
    "context"
    "fmt"
    commonapp "github.com/kart-io/k8s-agent/pkg/app"
    "github.com/kart-io/k8s-agent/pkg/bootstrap"
    "github.com/kart-io/logger/core"
)

const UserAgent = "aetherius-{service}"

func Execute() {
    opts := commonapp.NewStandardOptions("{Service}", UserAgent).
        WithDatabase().WithRedis()

    app := &{Service}App{}

    commonapp.RunWithBootstrap(app, opts, commonapp.Config{
        Use: "{service}", Short: "{Service} Service",
        Long: "Detailed description", EnvPrefix: "{SERVICE}",
    }, app.registerComponents)
}

// ✅ Only 3 fields
type {Service}App struct {
    bootstrap *bootstrap.Bootstrap
    opts      *commonapp.StandardOptions
    logger    core.Logger
}

func (a *{Service}App) Name() string { return "{Service}" }

func (a *{Service}App) Initialize(ctx context.Context, opts commonapp.Options) error {
    a.opts = opts.(*commonapp.StandardOptions)
    logger, err := a.opts.InitLogger()
    if err != nil {
        return fmt.Errorf("failed to initialize logger: %w", err)
    }
    a.logger = logger
    return nil
}

func (a *{Service}App) Run(ctx context.Context) error {
    <-ctx.Done()
    return ctx.Err()
}

func (a *{Service}App) Shutdown(ctx context.Context) error {
    return nil
}

// ✅ Auto-registration (3 lines)
func (a *{Service}App) registerComponents(bs *bootstrap.Bootstrap) error {
    a.bootstrap = bs
    components, err := Initialize{Service}Components(a.opts)
    if err != nil {
        return fmt.Errorf("failed to initialize components: %w", err)
    }

    for _, init := range components.GetInitializers() {
        bs.Register(init)
    }

    return nil
}
```

**Key:** 3-field App struct, 3-line auto-registration loop.

## Priority Ranges (Standard)

| Range | Purpose | Examples |
|-------|---------|----------|
| 100-299 | Foundation | Logging, config |
| 300-499 | Infrastructure | Database (300), Redis (400), NATS (500) |
| 500-699 | Business Logic | Services (600), Workflows (650) |
| 700-899 | Processors | Subscribers (700) |
| 900-1099 | Servers | gRPC (900), HTTP (1000) |
| 2000+ | Auxiliary | Health (2000), Metrics (2100) |

**CRITICAL:** Service layer (600) MUST be before servers (900-1000).

## Migration Checklist

### Step 1: Simplify wire.go

- [ ] Replace 4 nested provider sets with 1 flat `InitializerSet`
- [ ] List all `initializers.New*()` constructors in flat set
- [ ] Keep `HealthInitializerSet` separate
- [ ] Verify `Initialize{Service}Components()` function signature

### Step 2: Add GetInitializers() to components.go

- [ ] Change all fields from public (uppercase) to private (lowercase)
- [ ] Add `GetInitializers()` method returning `[]bootstrap.Initializer`
- [ ] Document priority for each component in comments
- [ ] Ensure CRITICAL components (service layer) are before servers

### Step 3: Simplify app.go

- [ ] Remove all component field storage (keep only 3: bootstrap, opts, logger)
- [ ] Replace manual `bs.Register()` calls with auto-registration loop
- [ ] Remove component reference assignments (9+ lines deleted)
- [ ] Verify `registerComponents()` uses `GetInitializers()`

### Step 4: Regenerate and Test

```bash
# From repository root
wire gen ./cmd/{service}/app
make go.build.{service}
make go.test.{service}
```

## Common Mistakes

### ❌ DON'T: Nest Wire provider sets

```go
var BaseSet = wire.NewSet(...)
var ServiceSet = wire.NewSet(BaseSet, ...)  // ← Bad: nesting
```

### ✅ DO: Use flat Wire provider set

```go
var InitializerSet = wire.NewSet(
    ProvideLogger,
    initializers.New...,  // ← Good: all at same level
)
```

### ❌ DON'T: Manual registration

```go
bs.Register(components.DB)      // ← Bad: 17+ manual lines
bs.Register(components.Redis)
// ... 15 more lines
```

### ✅ DO: Auto-registration

```go
for _, init := range components.GetInitializers() {
    bs.Register(init)  // ← Good: 3 lines, auto-order
}
```

### ❌ DON'T: Public component fields

```go
type Components struct {
    DB *DatabaseInitializer  // ← Bad: public field
}
```

### ✅ DO: Private fields + GetInitializers()

```go
type Components struct {
    db *DatabaseInitializer  // ← Good: private field
}

func (c *Components) GetInitializers() []bootstrap.Initializer {
    return []bootstrap.Initializer{c.db}
}
```

### ❌ DON'T: Store component references in App

```go
type App struct {
    dbInit    *DatabaseInitializer    // ← Bad: unused storage
    redisInit *RedisInitializer       // ← Bad: unused storage
    // ... 7 more unused fields
}
```

### ✅ DO: Only store essential fields

```go
type App struct {
    bootstrap *bootstrap.Bootstrap    // ← Good: used
    opts      *StandardOptions        // ← Good: used
    logger    core.Logger             // ← Good: used
}
```

## Quick Stats

| Aspect | Before | After | Improvement |
|--------|--------|-------|-------------|
| Wire provider sets | 4 nested levels | 1 flat level | -75% complexity |
| Registration lines | 17+ manual | 3 auto-loop | -82% boilerplate |
| App struct fields | 11 fields | 3 fields | -75% fields |
| Total LOC | ~257 | ~218 | -15% code |
| Encapsulation | Public fields | Private + method | Better API |

## Examples

- **Reference:** `/cmd/cluster/app/` (best practice)
- **Refactored:** `/cmd/agent-manager/app/`, `/cmd/orchestrator/app/`
- **Pending:** `/cmd/auth/app/`, `/cmd/reasoning/app/`

## Documentation

- **Full Guide:** [docs/SERVICE_STARTUP_PATTERN.md](SERVICE_STARTUP_PATTERN.md)
- **Comparison:** [docs/REFACTORING_COMPARISON.md](REFACTORING_COMPARISON.md)
- **Architecture:** [CLAUDE.md](../CLAUDE.md) (Service Entry Architecture Patterns section)

## When to Use

**Use Simplified Bootstrap Pattern when:**
- Service has 3+ external dependencies (database, Redis, NATS, etc.)
- Service needs complex initialization order (priority system)
- Service requires fine-grained lifecycle management
- Service complexity score ≥ 10

**Use Simple Pattern when:**
- Service has 0-2 dependencies
- Simple linear initialization
- Lightweight service (gateway, monitoring)
- Service complexity score < 10

## Support

**Questions?** Check:
1. [docs/SERVICE_STARTUP_PATTERN.md](SERVICE_STARTUP_PATTERN.md) - Complete guide
2. `/cmd/cluster/app/` - Reference implementation
3. [docs/REFACTORING_COMPARISON.md](REFACTORING_COMPARISON.md) - Before/after comparison
