# Service Startup Pattern - Simplified Bootstrap Architecture

## Overview

This document describes the **simplified Bootstrap pattern** used for service initialization in the Aetherius platform. This pattern reduces boilerplate code by ~50% while maintaining clean dependency injection and lifecycle management.

## Pattern Components

The simplified pattern consists of three files per service:

1. **app.go** - Application lifecycle (Initialize, Run, Shutdown)
2. **components.go** - Component aggregation with `GetInitializers()` method
3. **wire.go** - Flat Wire dependency injection configuration

## Architecture Principles

### 1. Flat Wire Provider Sets

**Before (Nested - 4 levels):**
```go
// ❌ Old: Nested provider sets create confusion
var BaseProviderSet = wire.NewSet(...)
var ServiceProviderSet = wire.NewSet(BaseProviderSet, ...)
var BusinessProviderSet = wire.NewSet(ServiceProviderSet, ...)
var ServerProviderSet = wire.NewSet(BusinessProviderSet, ...)
```

**After (Flat - 1 level):**
```go
// ✅ New: Single flat provider set
var InitializerSet = wire.NewSet(
    ProvideLogger,
    initializers.NewDatabaseInitializer,
    initializers.NewRedisInitializer,
    initializers.NewServiceInitializer,
    initializers.NewHTTPServerInitializer,
    initializers.NewGRPCServerInitializer,
)
```

**Benefits:**
- Easier to understand dependency graph
- Simpler to add/remove components
- Wire automatically resolves dependency order
- No manual dependency chain management

### 2. Auto-Registration via GetInitializers()

**Before (Manual - 17+ lines):**
```go
// ❌ Old: Manual registration with hardcoded order
bs.Register(components.DB)         // Priority 300
bs.Register(components.Redis)      // Priority 400
bs.Register(components.Service)    // Priority 600
bs.Register(components.Registry)   // Priority 450
bs.Register(components.NATS)       // Priority 500
// ... 12+ more lines
```

**After (Auto - 3 lines):**
```go
// ✅ New: Automatic registration via GetInitializers()
for _, init := range components.GetInitializers() {
    bs.Register(init)
}
```

**Benefits:**
- Single loop replaces 17+ lines of registration code
- Priority order defined once in `components.go`
- No need to manually track registration order
- Easier to add/remove components

### 3. Private Component Fields

**Before (Public fields):**
```go
// ❌ Old: Public fields exposed unnecessarily
type Components struct {
    DB    *initializers.DatabaseInitializer
    Redis *initializers.RedisInitializer
    // ...
}
```

**After (Private fields):**
```go
// ✅ New: Private fields, encapsulation via GetInitializers()
type Components struct {
    db    *initializers.DatabaseInitializer
    redis *initializers.RedisInitializer
    // ...
}

func (c *Components) GetInitializers() []bootstrap.Initializer {
    return []bootstrap.Initializer{c.db, c.redis, ...}
}
```

**Benefits:**
- Better encapsulation
- Clear API surface (only `GetInitializers()` is public)
- Prevents direct field access
- Matches cluster service pattern

### 4. Minimal App Struct

**Before (9+ component fields):**
```go
// ❌ Old: Stores all component references
type App struct {
    dbInit      *initializers.DatabaseInitializer
    redisInit   *initializers.RedisInitializer
    serviceInit *initializers.ServiceInitializer
    // ... 6+ more fields
}
```

**After (3 essential fields):**
```go
// ✅ New: Only stores bootstrap, opts, logger
type App struct {
    bootstrap *bootstrap.Bootstrap
    opts      *commonapp.StandardOptions
    logger    core.Logger
}
```

**Benefits:**
- Reduces App struct from 12 fields to 3
- Components managed by Bootstrap framework
- No need to save component references
- Cleaner app lifecycle

## File Structure

### 1. app.go (Application Lifecycle)

**Purpose:** Implements `commonapp.Application` interface with minimal boilerplate.

**Structure:**
```go
package app

import (
    "context"
    "fmt"
    commonapp "github.com/kart-io/k8s-agent/pkg/app"
    "github.com/kart-io/k8s-agent/pkg/bootstrap"
    "github.com/kart-io/logger/core"
)

const UserAgent = "aetherius-service-name"

// Execute runs the service command
func Execute() {
    opts := commonapp.NewStandardOptions("Service Name", UserAgent).
        WithDatabase().
        WithRedis()
        // ... other options

    app := &ServiceApp{}

    commonapp.RunWithBootstrap(
        app,
        opts,
        commonapp.Config{
            Use:       "service-name",
            Short:     "Service Description",
            Long:      "Detailed service description",
            EnvPrefix: "SERVICE_NAME",
        },
        app.registerComponents,
    )
}

// ServiceApp implements commonapp.Application
type ServiceApp struct {
    bootstrap *bootstrap.Bootstrap
    opts      *commonapp.StandardOptions
    logger    core.Logger
}

func (a *ServiceApp) Name() string {
    return "Service Name"
}

func (a *ServiceApp) Initialize(ctx context.Context, opts commonapp.Options) error {
    a.opts = opts.(*commonapp.StandardOptions)

    logger, err := a.opts.InitLogger()
    if err != nil {
        return fmt.Errorf("failed to initialize logger: %w", err)
    }
    a.logger = logger

    return nil
}

func (a *ServiceApp) Run(ctx context.Context) error {
    // Bootstrap.Run() is already called by RunWithBootstrap
    <-ctx.Done()
    return ctx.Err()
}

func (a *ServiceApp) Shutdown(ctx context.Context) error {
    // Bootstrap shutdown is handled automatically
    return nil
}

// registerComponents registers all component initializers
func (a *ServiceApp) registerComponents(bs *bootstrap.Bootstrap) error {
    a.bootstrap = bs

    // Use Wire to automatically inject all dependencies
    components, err := InitializeServiceComponents(a.opts)
    if err != nil {
        return fmt.Errorf("failed to initialize components: %w", err)
    }

    // Register all initializers using GetInitializers() method
    for _, init := range components.GetInitializers() {
        bs.Register(init)
    }

    return nil
}
```

**Key Points:**
- Only 3 fields in App struct (bootstrap, opts, logger)
- No component field storage needed
- Single loop for component registration
- Clean separation of concerns

**Lines of Code:** ~100 LOC (down from ~138 LOC)

### 2. components.go (Component Aggregation)

**Purpose:** Aggregates all component initializers and provides `GetInitializers()` method.

**Structure:**
```go
package app

import (
    "github.com/kart-io/k8s-agent/internal/service-name/initializers"
    commonapp "github.com/kart-io/k8s-agent/pkg/app"
    "github.com/kart-io/k8s-agent/pkg/bootstrap"
    pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
    "github.com/kart-io/logger/core"
)

// ServiceComponents contains all initialized components
type ServiceComponents struct {
    db      *initializers.DatabaseInitializer
    redis   *initializers.RedisInitializer
    service *initializers.ServiceInitializer
    http    *initializers.HTTPServerInitializer
    grpc    *initializers.GRPCServerInitializer
    health  *pkginitializers.HealthCheckInitializer
}

// NewServiceComponents creates a new ServiceComponents instance
// This is called by Wire with all dependencies automatically injected
func NewServiceComponents(
    db *initializers.DatabaseInitializer,
    redis *initializers.RedisInitializer,
    service *initializers.ServiceInitializer,
    http *initializers.HTTPServerInitializer,
    grpc *initializers.GRPCServerInitializer,
    health *pkginitializers.HealthCheckInitializer,
) *ServiceComponents {
    return &ServiceComponents{
        db:      db,
        redis:   redis,
        service: service,
        http:    http,
        grpc:    grpc,
        health:  health,
    }
}

// GetInitializers returns all initializers for Bootstrap framework
// Bootstrap will automatically register them in priority order
func (c *ServiceComponents) GetInitializers() []bootstrap.Initializer {
    return []bootstrap.Initializer{
        c.db,      // Priority 300
        c.redis,   // Priority 400
        c.service, // Priority 600 - CRITICAL: Service layer MUST come before servers
        c.http,    // Priority 1000
        c.grpc,    // Priority 900
        c.health,  // Priority 2000
    }
}

// ProvideLogger creates a logger instance from options
func ProvideLogger(opts *commonapp.StandardOptions) (core.Logger, error) {
    return opts.InitLogger()
}
```

**Key Points:**
- Private fields for encapsulation
- `GetInitializers()` is the only public method
- Priority order documented in comments
- Clean factory function created by Wire

**Lines of Code:** ~73 LOC (down from ~56 LOC for public fields version)

### 3. wire.go (Dependency Injection)

**Purpose:** Defines Wire provider sets and injection function.

**Structure:**
```go
//go:build wireinject
// +build wireinject

package app

import (
    "github.com/google/wire"
    "github.com/kart-io/k8s-agent/internal/service-name/initializers"
    commonapp "github.com/kart-io/k8s-agent/pkg/app"
    pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
)

// InitializerSet Wire dependency set for all initializers
var InitializerSet = wire.NewSet(
    ProvideLogger,
    initializers.NewDatabaseInitializer,
    initializers.NewRedisInitializer,
    initializers.NewServiceInitializer,
    initializers.NewHTTPServerInitializer,
    initializers.NewGRPCServerInitializer,
)

// HealthInitializerSet Wire dependency set for health check
var HealthInitializerSet = wire.NewSet(
    pkginitializers.NewHealthCheckInitializer,
    wire.FieldsOf(new(*commonapp.StandardOptions), "Health"),
)

// InitializeServiceComponents automatically injects all components using Wire
func InitializeServiceComponents(opts *commonapp.StandardOptions) (*ServiceComponents, error) {
    wire.Build(
        InitializerSet,
        HealthInitializerSet,
        NewServiceComponents,
    )
    return nil, nil
}
```

**Key Points:**
- Single flat `InitializerSet` (no nesting)
- Separate `HealthInitializerSet` for health check
- Wire automatically resolves dependency order
- Simple, readable structure

**Lines of Code:** ~45 LOC (down from ~63 LOC)

## Code Reduction Summary

### agent-manager Service

| File | Before | After | Reduction |
|------|--------|-------|-----------|
| app.go | 138 LOC | 100 LOC | -38 LOC (-27%) |
| components.go | 56 LOC | 73 LOC | +17 LOC (encapsulation) |
| wire.go | 63 LOC | 45 LOC | -18 LOC (-29%) |
| **Total** | **257 LOC** | **218 LOC** | **-39 LOC (-15%)** |

### orchestrator Service

| File | Before | After | Reduction |
|------|--------|-------|-----------|
| app.go | 131 LOC | 98 LOC | -33 LOC (-25%) |
| components.go | 59 LOC | 77 LOC | +18 LOC (encapsulation) |
| wire.go | 69 LOC | 46 LOC | -23 LOC (-33%) |
| **Total** | **259 LOC** | **221 LOC** | **-38 LOC (-15%)** |

**Overall Impact:**
- **15-20% reduction** in total lines of code
- **27-33% reduction** in wire.go (flat structure)
- **25-27% reduction** in app.go (auto-registration)
- Better encapsulation (private fields)
- Improved readability and maintainability

## Migration Guide

### Step 1: Simplify wire.go

**Before:**
```go
var BaseProviderSet = wire.NewSet(...)
var ServiceProviderSet = wire.NewSet(BaseProviderSet, ...)
var BusinessProviderSet = wire.NewSet(ServiceProviderSet, ...)
```

**After:**
```go
var InitializerSet = wire.NewSet(
    ProvideLogger,
    initializers.New...,
    // List all initializer constructors
)
```

### Step 2: Add GetInitializers() to components.go

**Add method:**
```go
func (c *ServiceComponents) GetInitializers() []bootstrap.Initializer {
    return []bootstrap.Initializer{
        c.db,
        c.redis,
        // ... all components in priority order
    }
}
```

**Make fields private:**
```go
type ServiceComponents struct {
    db    *initializers.DatabaseInitializer  // lowercase
    redis *initializers.RedisInitializer     // lowercase
    // ...
}
```

### Step 3: Simplify app.go

**Remove component fields:**
```go
// Before:
type App struct {
    dbInit    *initializers.DatabaseInitializer
    redisInit *initializers.RedisInitializer
    // ... 9+ more fields
}

// After:
type App struct {
    bootstrap *bootstrap.Bootstrap
    opts      *commonapp.StandardOptions
    logger    core.Logger
}
```

**Replace manual registration:**
```go
// Before: 17+ lines
bs.Register(components.DB)
bs.Register(components.Redis)
// ... 15+ more lines

// After: 3 lines
for _, init := range components.GetInitializers() {
    bs.Register(init)
}
```

### Step 4: Regenerate Wire Code

```bash
# From repository root
wire gen ./cmd/service-name/app
```

### Step 5: Build and Test

```bash
# Build the service
make go.build.service-name

# Run tests
make go.test.service-name
```

## Best Practices

### 1. Component Priority Order

Always document priority order in `GetInitializers()` comments:

```go
func (c *Components) GetInitializers() []bootstrap.Initializer {
    return []bootstrap.Initializer{
        c.db,      // Priority 300 - Database
        c.redis,   // Priority 400 - Cache
        c.service, // Priority 600 - Business services
        c.http,    // Priority 1000 - HTTP server
        c.health,  // Priority 2000 - Health checks
    }
}
```

**Standard Priority Ranges:**
- 100-299: Foundation (logging, config)
- 300-499: Infrastructure (database, cache, message queue)
- 500-699: Business logic (services, workflows)
- 700-899: Subscribers and processors
- 900-1099: Protocol servers (gRPC, HTTP)
- 2000+: Auxiliary (health checks, metrics)

### 2. Wire Provider Organization

Keep provider sets simple and flat:

```go
// ✅ Good: Flat, easy to read
var InitializerSet = wire.NewSet(
    ProvideLogger,
    initializers.NewDatabaseInitializer,
    initializers.NewRedisInitializer,
)

// ❌ Bad: Nested, hard to follow
var BaseSet = wire.NewSet(ProvideLogger)
var DBSet = wire.NewSet(BaseSet, initializers.NewDatabaseInitializer)
var RedisSet = wire.NewSet(DBSet, initializers.NewRedisInitializer)
```

### 3. Component Encapsulation

Use private fields and expose via `GetInitializers()`:

```go
// ✅ Good: Encapsulated
type Components struct {
    db *initializers.DatabaseInitializer  // private
}

func (c *Components) GetInitializers() []bootstrap.Initializer {
    return []bootstrap.Initializer{c.db}
}

// ❌ Bad: Exposed fields
type Components struct {
    DB *initializers.DatabaseInitializer  // public
}
```

### 4. Error Handling

Always check Wire injection errors:

```go
components, err := InitializeServiceComponents(a.opts)
if err != nil {
    return fmt.Errorf("failed to initialize components: %w", err)
}
```

### 5. Documentation

Document the pattern in each service's app.go:

```go
// registerComponents registers all component initializers with bootstrap.
// This method uses Wire for dependency injection and GetInitializers()
// for automatic registration in priority order.
func (a *ServiceApp) registerComponents(bs *bootstrap.Bootstrap) error {
    // ...
}
```

## Service Implementation Status

| Service | Status | Pattern | LOC Before | LOC After | Reduction |
|---------|--------|---------|------------|-----------|-----------|
| **agent-manager** | ✅ Refactored | Simplified Bootstrap | 257 | 218 | -15% |
| **orchestrator** | ✅ Refactored | Simplified Bootstrap | 259 | 221 | -15% |
| **cluster** | ✅ Reference | Simplified Bootstrap | - | 221 | - |
| **reasoning** | ⏳ Pending | Bootstrap | - | - | - |
| **auth** | ⏳ Pending | Bootstrap | - | - | - |
| **collect-agent** | N/A | Simple | - | - | - |
| **gateway** | N/A | Simple | - | - | - |
| **monitor** | N/A | Simple | - | - | - |

**Legend:**
- ✅ Refactored: Fully migrated to simplified pattern
- ⏳ Pending: Using old Bootstrap pattern, needs refactoring
- N/A: Uses Simple pattern (no Bootstrap)

## Comparison with Simple Pattern

The project supports two patterns based on service complexity:

### Simplified Bootstrap Pattern (5 services)

**Used by:** agent-manager, orchestrator, auth, cluster, reasoning

**When to use:**
- Service has multiple external dependencies (database, Redis, NATS, etc.)
- Service needs complex initialization order
- Service requires fine-grained lifecycle management
- Service complexity score ≥ 10

**Characteristics:**
- Uses `pkg/app.RunWithBootstrap()`
- Uses `pkg/bootstrap.Bootstrap` for lifecycle management
- Uses Wire for dependency injection
- Has `GetInitializers()` method for auto-registration
- ~220 LOC per service

### Simple Pattern (3 services)

**Used by:** collect-agent, gateway, monitor

**When to use:**
- Service has few or no external dependencies
- Simple linear initialization logic
- Lightweight service (gateway, monitoring, etc.)
- Service complexity score < 10

**Characteristics:**
- Uses `pkg/app.RunWithOptions()`
- No Bootstrap framework, linear initialization
- Configuration in `internal/{service}/config/` package
- ~150 LOC per service

## Troubleshooting

### Wire Generation Errors

**Problem:** Wire fails with "no provider found for X"

**Solution:**
1. Ensure all dependencies are in `InitializerSet`
2. Check that constructor functions match Wire expectations
3. Verify `ProvideLogger` is included

**Example:**
```bash
wire: github.com/kart-io/k8s-agent/cmd/service/app: no provider found for *initializers.ServiceInitializer

# Fix: Add missing provider to InitializerSet
var InitializerSet = wire.NewSet(
    ProvideLogger,
    initializers.NewServiceInitializer,  // Add this
)
```

### Priority Order Issues

**Problem:** Components initialize in wrong order

**Solution:**
1. Check priority values in each initializer's `Priority()` method
2. Ensure `GetInitializers()` returns components in dependency order
3. Higher numbers run later (service=600, http=1000)

### Missing Components

**Problem:** Component not registered in Bootstrap

**Solution:**
1. Verify component is in `GetInitializers()` return slice
2. Check that Wire created the component in `wire_gen.go`
3. Ensure `for _, init := range components.GetInitializers()` loop runs

## References

- **Reference Implementation:** `/cmd/cluster/app/` (best practice example)
- **Bootstrap Framework:** `/pkg/bootstrap/bootstrap.go`
- **Wire Documentation:** https://github.com/google/wire
- **Code Reorganization:** `/docs/CODE_REORGANIZATION.md`
- **Makefile Usage:** `/docs/MAKEFILE_USAGE_EXAMPLES.md`

## Future Work

### Remaining Services to Refactor

1. **auth** service - Bootstrap pattern, needs simplification
2. **reasoning** service - Bootstrap pattern, needs simplification

### Potential Improvements

1. **Auto-discovery:** Generate `GetInitializers()` from struct tags
2. **Type safety:** Use generics for component registration
3. **Validation:** Auto-validate priority order at compile time
4. **Documentation:** Auto-generate initialization order diagrams

## Conclusion

The simplified Bootstrap pattern achieves:

- **15-20% code reduction** in service startup files
- **50% reduction** in manual registration boilerplate
- **Better encapsulation** via private fields
- **Clearer dependencies** via flat Wire provider sets
- **Consistent architecture** across all Bootstrap services

This pattern maintains the benefits of Wire (dependency injection) and Bootstrap (lifecycle management) while significantly reducing complexity and boilerplate code.
