# Quick Service Startup Reference

**For detailed guide**: See [docs/SERVICE_STARTUP_GUIDE.md](docs/SERVICE_STARTUP_GUIDE.md)
**For code templates**: See [docs/templates/service_startup_template.go](docs/templates/service_startup_template.go)

## Pattern Selection (30 seconds)

```
Count external dependencies (DB, Redis, NATS, APIs, etc.):
- 0-2 deps  → Use ULTRA-SIMPLE PATTERN
- 3+ deps   → Use BOOTSTRAP PATTERN
```

## Ultra-Simple Pattern (< 5 dependencies)

**Services**: collect-agent, gateway, monitor

**Structure**:
```
cmd/{service}/
├── main.go           (5 lines)
└── app/
    └── app.go        (100-150 lines)
```

**Key Code**:
```go
// Execute - entry point
func Execute() {
    opts := commonapp.NewStandardOptions("Name", "user-agent").WithAgent()
    app := &MyApp{}
    commonapp.Run(app, opts, commonapp.Config{...})
}

// Implement Application interface
func (a *MyApp) Initialize(ctx context.Context, opts commonapp.Options) error {
    a.logger, _ = opts.(*commonapp.StandardOptions).InitLogger()
    // Direct initialization: a.client = NewClient()
    return nil
}

func (a *MyApp) Run(ctx context.Context) error {
    return a.client.Start(ctx)  // Blocks until cancelled
}

func (a *MyApp) Shutdown(ctx context.Context) error {
    // Cleanup: a.client.Close()
    return nil
}
```

## Bootstrap Pattern (5+ dependencies)

**Services**: agent-manager, orchestrator, auth, cluster, reasoning

**Structure**:
```
cmd/{service}/
├── main.go           (5 lines)
└── app/
    └── app.go        (500-600 lines)

internal/{service}/
└── startup/          (Optional: for grouped initializers)
    ├── infrastructure.go
    ├── core_services.go
    └── servers.go
```

**Key Code**:
```go
// Execute - entry point
func Execute() {
    opts := commonapp.NewStandardOptions("Name", "user-agent").
        WithDatabase().
        WithRedis().
        WithNATS()

    app := &MyApp{}
    commonapp.RunWithBootstrap(
        app, opts,
        commonapp.Config{...},
        app.registerComponents,
    )
}

// Implement Application interface
func (a *MyApp) Initialize(ctx context.Context, opts commonapp.Options) error {
    a.opts = opts.(*commonapp.StandardOptions)
    a.logger, _ = a.opts.InitLogger()
    return nil
}

func (a *MyApp) Run(ctx context.Context) error {
    <-ctx.Done()  // Just wait for shutdown
    return ctx.Err()
}

func (a *MyApp) Shutdown(ctx context.Context) error {
    return nil  // Bootstrap handles everything
}

// Core method: defines initialization order via Priority
func (a *MyApp) registerComponents(bs *bootstrap.Bootstrap) error {
    a.bootstrap = bs

    // Layer 1: Infrastructure (Priority 300-500)
    dbInit := pkginitializers.NewDatabaseInitializer(a.opts.Database, a.logger)
    bs.Register(dbInit)

    // Layer 2: Core Services (Priority 600-700)
    serviceInit := &myServiceInitializer{dbInit: dbInit, logger: a.logger}
    bs.Register(serviceInit)

    // Layer 3: Servers (Priority 900-1000)
    httpInit := &httpServerInitializer{coreInit: serviceInit, logger: a.logger}
    bs.Register(httpInit)

    // Layer 4: Monitoring (Priority 2000)
    healthInit := pkginitializers.NewHealthCheckInitializer(
        a.opts.Health, a.logger,
    )
    bs.Register(healthInit)

    return nil
}

// Inline initializer for service-specific logic
type myServiceInitializer struct {
    *bootstrap.BaseInitializer
    dbInit *pkginitializers.DatabaseInitializer
    logger core.Logger
    myService *MyService
}

func (m *myServiceInitializer) Initialize(ctx context.Context) error {
    m.myService = NewMyService(m.dbInit.Client(), m.logger)
    return nil
}

func (m *myServiceInitializer) Close(ctx context.Context) error {
    return m.myService.Close()
}

func (m *myServiceInitializer) MyService() *MyService {
    return m.myService
}
```

## Initialization Priority Guidelines

```
100-200:   Framework reserved
300-400:   Primary infrastructure (Database)
350-400:   Secondary infrastructure (Redis)
450-500:   Tertiary infrastructure (Email, APIs)
600-700:   Core services (main business logic)
650-800:   Feature services (advanced features)
800-900:   Specialized services (NATS, event bus)
900:       gRPC servers
950:       HTTP servers
2000:      Monitoring (health checks)
```

## StandardOptions Capabilities

```go
opts := commonapp.NewStandardOptions("Service", "user-agent").
    WithServer()      // HTTP server
    .WithDatabase()   // MySQL
    .WithRedis()      // Redis
    .WithNATS()       // NATS
    .WithGRPC()       // gRPC
    .WithJWT()        // JWT auth
    .WithEmail()      // Email
    .WithMetrics()    // Prometheus
    .WithAgent()      // Agent config
    .WithCORS()       // CORS

// All options available via a.opts
// - a.opts.Server.Port
// - a.opts.Database.Host
// - a.opts.Redis.Addr
// - etc.
```

## Adding a New Service (Checklist)

### 1. Choose Pattern
- [ ] Understand your dependencies (< 5 → Ultra-Simple, >= 5 → Bootstrap)
- [ ] Read applicable section in SERVICE_STARTUP_GUIDE.md

### 2. Create Directory Structure

**Ultra-Simple**:
```bash
mkdir -p cmd/{service}/app
mkdir -p internal/{service}
```

**Bootstrap**:
```bash
mkdir -p cmd/{service}/app
mkdir -p internal/{service}/{startup,api,service,storage}
```

### 3. Implement Files
- [ ] `cmd/{service}/main.go` - Copy template, minimal
- [ ] `cmd/{service}/app/app.go` - Copy template, implement Application interface
- [ ] (Bootstrap only) `internal/{service}/startup/*.go` - Infrastructure, services, servers

### 4. Register in Build System
- [ ] Add service to `SERVICES` variable in root Makefile

### 5. Test
```bash
make go.build.{service}    # Build
make run-{service}         # Run
make go.test.{service}     # Test
```

### 6. Verify Checklist
- [ ] Logger initialized first in Initialize()
- [ ] Correct Priority() values (Bootstrap only)
- [ ] All resources properly closed
- [ ] Tests pass
- [ ] Shutdown signal (SIGTERM) handled gracefully

## Common Patterns

### Pattern: Conditional Component Registration (Bootstrap)

```go
func (a *MyApp) registerComponents(bs *bootstrap.Bootstrap) error {
    // ... infrastructure ...

    // Only register NATS if enabled
    if a.opts.NATS.Enable {
        natsInit := startup.NewNATSInitializer(a.opts, a.logger)
        bs.Register(natsInit)
    }

    // Only register gRPC if enabled
    if a.opts.GRPC.Enable {
        grpcInit := startup.NewGRPCServerInitializer(...)
        bs.Register(grpcInit)
    }

    return nil
}
```

### Pattern: Database Auto-Migration (Bootstrap)

```go
dbInit := &DatabaseInitializer{
    DatabaseInitializer: pkginitializers.NewDatabaseInitializer(
        a.opts.Database, a.logger,
    ),
}

if a.opts.Database.AutoMigrate {
    dbInit.WithAutoMigrate(
        &types.Agent{},
        &types.Event{},
        &types.Metrics{},
    )
}

bs.Register(dbInit)
```

### Pattern: Accessing Configuration in registerComponents

```go
// All options available
if a.opts.Server.EnableMetrics {
    // setup metrics
}

if a.opts.Database.PoolSize > 100 {
    // large pool optimization
}

// Environment overrides work
// DATABASE_HOST=custom.host override config file
```

## Troubleshooting

### "cannot find package" Error

**Cause**: Initialization dependency not registered first
**Fix**: Move initializer to later priority level

```go
// Wrong: HTTP depends on services (600) but registered at 400
// Right: Register core services (600) before HTTP server (950)
bs.Register(coreInit)      // 600
bs.Register(httpInit)      // 950
```

### Service Hangs on Shutdown

**Cause**: Components not implementing Close()
**Fix**: Add Close() method to initializers that open resources

```go
func (i *MyInitializer) Close(ctx context.Context) error {
    if i.resource != nil {
        return i.resource.Close()
    }
    return nil
}
```

### Configuration Not Applied

**Cause**: Feature not enabled in options
**Fix**: Use WithXXX() methods to enable capabilities

```go
// Must enable to use Redis
opts := commonapp.NewStandardOptions(...).WithRedis()

// Check in code
if a.opts.Redis.Enable {
    // use Redis
}
```

## Real-World Examples

### Complex Service (agent-manager)
- File: `/cmd/agent-manager/app/app.go`
- Pattern: Bootstrap
- Dependencies: MySQL, Redis, NATS
- LOC: 504
- Initializers: 5 layers

### Simple Service (collect-agent)
- File: `/cmd/collect-agent/app/app.go`
- Pattern: Ultra-Simple
- Dependencies: NATS
- LOC: 122
- No Bootstrap framework

### Medium Service (auth)
- File: `/cmd/auth/app/app.go`
- Pattern: Bootstrap
- Dependencies: MySQL, Redis, Email, JWT
- LOC: 620
- Initializers: 7 layers with features

## Next Steps

1. **Read**: [docs/SERVICE_STARTUP_GUIDE.md](docs/SERVICE_STARTUP_GUIDE.md) (comprehensive)
2. **Copy**: Template from [docs/templates/service_startup_template.go](docs/templates/service_startup_template.go)
3. **Compare**: Example from cmd/{existing-service}/app/app.go
4. **Implement**: Follow step-by-step guide in SERVICE_STARTUP_GUIDE.md
5. **Test**: Use provided checklists to verify

---

**Last Updated**: 2025-11-09
**Guide Status**: Complete and tested
**Examples**: All 8 services refactored and working
