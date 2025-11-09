# Bootstrap Startup Pattern - Quick Reference

**Last Updated**: 2025-11-09

## Services Using This Pattern

- ✅ agent-manager (5 Bootstrap-based services total)
- ✅ orchestrator
- ✅ auth
- ✅ cluster
- ✅ reasoning

## The Pattern (3 Files)

### 1. app.go - Application Lifecycle

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

// Execute runs the service command.
func Execute() {
    opts := commonapp.NewStandardOptions("Service", UserAgent).
        WithDatabase().WithRedis() // Configure as needed

    app := &{Service}App{}

    commonapp.RunWithBootstrap(
        app, opts,
        commonapp.Config{
            Use:       "{service}",
            Short:     "{Service} Service",
            Long:      "Description...",
            EnvPrefix: "{SERVICE}",
        },
        app.registerComponents,
    )
}

// {Service}App implements commonapp.Application interface.
type {Service}App struct {
    bootstrap *bootstrap.Bootstrap
    opts      *commonapp.StandardOptions
    logger    core.Logger
}

func (a *{Service}App) Name() string {
    return "{Service} Service"
}

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
    // Bootstrap.Run() is already called by RunWithBootstrap
    <-ctx.Done()
    return ctx.Err()
}

func (a *{Service}App) Shutdown(ctx context.Context) error {
    // Bootstrap shutdown is handled automatically
    return nil
}

// registerComponents registers all component initializers with bootstrap.
func (a *{Service}App) registerComponents(bs *bootstrap.Bootstrap) error {
    a.bootstrap = bs

    // Use Wire to automatically inject all dependencies
    components, err := Initialize{Service}Components(a.opts)
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

### 2. components.go - Component Container

```go
package app

import (
    "github.com/kart-io/k8s-agent/internal/{service}/initializers"
    commonapp "github.com/kart-io/k8s-agent/pkg/app"
    "github.com/kart-io/k8s-agent/pkg/bootstrap"
    pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
    "github.com/kart-io/logger/core"
)

// {Service}Components contains all component initializers.
type {Service}Components struct {
    // Private fields - lowercase names
    db      *pkginitializers.DatabaseInitializer
    redis   *pkginitializers.RedisInitializer
    service *initializers.ServiceInitializer
    // ... add service-specific initializers
    http    *initializers.HTTPServerInitializer
    health  *pkginitializers.HealthCheckInitializer
}

// New{Service}Components creates a new instance.
// Called by Wire with all dependencies automatically injected.
func New{Service}Components(
    db *pkginitializers.DatabaseInitializer,
    redis *pkginitializers.RedisInitializer,
    service *initializers.ServiceInitializer,
    // ... all initializers as parameters
    http *initializers.HTTPServerInitializer,
    health *pkginitializers.HealthCheckInitializer,
) *{Service}Components {
    return &{Service}Components{
        db:      db,
        redis:   redis,
        service: service,
        // ... assign all fields
        http:    http,
        health:  health,
    }
}

// GetInitializers returns all initializers for Bootstrap framework.
// Bootstrap automatically handles priority ordering.
func (c *{Service}Components) GetInitializers() []bootstrap.Initializer {
    return []bootstrap.Initializer{
        c.db,      // Priority 300
        c.redis,   // Priority 400
        c.service, // Priority 600
        // ... list all components in any order
        c.http,    // Priority 1000
        c.health,  // Priority 2000
    }
}

// ProvideLogger provides logger from options.
func ProvideLogger(opts *commonapp.StandardOptions) (core.Logger, error) {
    return opts.InitLogger()
}
```

### 3. wire.go - Dependency Injection

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

// CoreProviderSet contains all providers in a single flat set.
var CoreProviderSet = wire.NewSet(
    // Logger
    ProvideLogger,

    // Infrastructure (Priority 300-500)
    pkginitializers.DatabaseInitializerProvider,
    pkginitializers.RedisInitializerProvider,

    // Business Services (Priority 600-800)
    initializers.NewServiceInitializer,
    // ... add service-specific initializers

    // Protocol Servers (Priority 900-1000)
    initializers.NewHTTPServerInitializer,
    initializers.NewGRPCServerInitializer,

    // Health Check (Priority 2000)
    pkginitializers.NewHealthCheckInitializer,
    wire.FieldsOf(new(*commonapp.StandardOptions), "Health"),
)

// Initialize{Service}Components automatically injects all components using Wire.
func Initialize{Service}Components(opts *commonapp.StandardOptions) (*{Service}Components, error) {
    wire.Build(
        CoreProviderSet,
        New{Service}Components,
    )
    return nil, nil
}
```

## Standard Priority Levels

| Priority | Component Type | Examples |
|----------|---------------|----------|
| 100-200  | Config        | Configuration loading |
| 300      | Database      | MySQL, PostgreSQL |
| 400      | Cache         | Redis, Memcached |
| 500      | Message Queue | NATS, Kafka |
| 600      | Core Services | Business logic layer |
| 700-800  | Aux Services  | Email, Audit, Notification |
| 900      | gRPC Server   | gRPC protocol server |
| 1000     | HTTP Server   | HTTP/REST API server |
| 2000     | Health Check  | Readiness/liveness probes |

## Implementation Checklist

When creating a new Bootstrap-based service:

### Step 1: Create app.go
- [ ] Define Execute() function with RunWithBootstrap
- [ ] Define {Service}App struct (bootstrap, opts, logger only)
- [ ] Implement Application interface (Name, Initialize, Run, Shutdown)
- [ ] Implement registerComponents() with auto-registration loop

### Step 2: Create components.go
- [ ] Define {Service}Components struct with private fields
- [ ] Implement New{Service}Components() constructor
- [ ] Implement GetInitializers() method with priority comments
- [ ] Add ProvideLogger() function

### Step 3: Create wire.go
- [ ] Add wireinject build tag
- [ ] Define CoreProviderSet with all providers
- [ ] Implement Initialize{Service}Components() function
- [ ] Add any custom provider functions if needed

### Step 4: Generate and Verify
```bash
# Generate Wire code
wire gen ./cmd/{service}/app

# Build service
make go.build.{service}

# Verify startup
make run-{service}
```

## Common Patterns

### Using Factory Providers (Database, Redis)

```go
// In wire.go CoreProviderSet
pkginitializers.DatabaseInitializerProvider,  // Factory provider
pkginitializers.RedisInitializerProvider,     // Factory provider
```

These eliminate the need for service-specific database/redis initializers.

### Conditional Components (gRPC)

```go
// In components.go
type Components struct {
    grpc *initializers.GRPCServerInitializer // May be nil if disabled
}

// In GetInitializers()
func (c *Components) GetInitializers() []bootstrap.Initializer {
    return []bootstrap.Initializer{
        c.db,
        c.grpc, // Bootstrap handles nil gracefully
        c.http,
    }
}
```

### Custom Provider Functions

```go
// In wire.go - for complex initialization
func ProvideCustomComponent(
    opts *commonapp.StandardOptions,
    logger core.Logger,
    db *gorm.DB,
) (*CustomComponent, error) {
    // Complex initialization logic
    return NewCustomComponent(opts, logger, db)
}
```

## Anti-Patterns (Don't Do This)

### ❌ Manual Registration
```go
// DON'T: Manual bs.Register() calls
bs.Register(components.DB)
bs.Register(components.Redis)
bs.Register(components.Service)
// ...repeat 10 times
```

### ❌ Storing Component References in App
```go
// DON'T: Component fields in App struct
type App struct {
    dbInit      *DatabaseInitializer
    redisInit   *RedisInitializer
    serviceInit *ServiceInitializer
    // ... lots of fields
}
```

### ❌ Public Component Fields
```go
// DON'T: Exported fields in Components
type Components struct {
    DB    *DatabaseInitializer  // Public
    Redis *RedisInitializer     // Public
}
```

### ❌ Nested Provider Sets
```go
// DON'T: Multiple provider sets
var InfraProviderSet = wire.NewSet(...)
var ServiceProviderSet = wire.NewSet(...)
var ServerProviderSet = wire.NewSet(...)

// DO: Single flat CoreProviderSet
var CoreProviderSet = wire.NewSet(
    // All providers in one place
)
```

## Benefits of This Pattern

1. **Consistency**: All services follow identical structure
2. **Simplicity**: Minimal boilerplate code
3. **Type Safety**: Wire provides compile-time DI verification
4. **Maintainability**: Single source of truth for component ordering
5. **Scalability**: Easy to add/remove components
6. **Auto-Registration**: No manual loops or field assignments

## Related Documentation

- [Bootstrap Framework](../../pkg/bootstrap/README.md) - Bootstrap architecture
- [Startup Simplification Summary](./STARTUP_SIMPLIFICATION_SUMMARY.md) - Migration details
- [CLAUDE.md](../../CLAUDE.md) - Service architecture patterns
