# Service Startup Guide

**Last Updated**: 2025-01-15
**Status**: Complete
**Examples**: cmd/agent-manager, cmd/auth, cmd/collect-agent

## Table of Contents

1. [Overview](#overview)
2. [Pattern Selection](#pattern-selection)
3. [Bootstrap Pattern (Complex Services)](#bootstrap-pattern-complex-services)
4. [Ultra-Simple Pattern (Simple Services)](#ultra-simple-pattern-simple-services)
5. [Step-by-Step Implementation](#step-by-step-implementation)
6. [Common Patterns](#common-patterns)
7. [Testing Strategies](#testing-strategies)
8. [Troubleshooting](#troubleshooting)
9. [Checklists](#checklists)

## Overview

The Aetherius project uses two standardized service startup patterns based on service complexity:

- **Bootstrap Pattern**: For complex services with multiple interdependent components (5+ dependencies)
- **Ultra-Simple Pattern**: For lightweight services with minimal dependencies (<5)

Both patterns implement the `Application` interface from `pkg/app` and follow Go best practices for graceful shutdown and resource management.

### Service Application Interface

All services implement this interface:

```go
type Application interface {
    // Name returns the service name
    Name() string

    // Initialize sets up components before running
    Initialize(ctx context.Context, opts Options) error

    // Run starts the service and blocks until context cancelled
    Run(ctx context.Context) error

    // Shutdown performs graceful cleanup
    Shutdown(ctx context.Context) error
}
```

### Startup Flow (Both Patterns)

```
main.go (minimal)
    ↓
app.Execute()
    ↓
commonapp.RunWithBootstrap() or commonapp.Run()
    ↓
app.Initialize()
    ↓
app.Run() [blocks]
    ↓
[Shutdown Signal: SIGTERM/SIGINT]
    ↓
app.Shutdown()
    ↓
Exit
```

## Pattern Selection

### Decision Matrix

| Factor | Bootstrap Pattern | Ultra-Simple Pattern |
|--------|-------------------|----------------------|
| **External Dependencies** | 5+ | 0-3 |
| **Initialization Order** | Complex | Linear |
| **Feature Modules** | Multiple | Single |
| **Complexity Score** | >= 10 | < 10 |
| **Lifecycle Management** | Fine-grained | Simple |
| **Code Size** | 150-300 lines | 50-100 lines |
| **Testing** | Complex setup | Straightforward |

### Current Service Mapping

**Bootstrap Pattern** (5/8 services):
- `agent-manager` - Central control plane with 6+ initializers
- `orchestrator` - Workflow engine with complex dependencies
- `auth` - User management with session/audit services
- `cluster` - Multi-cluster management with database
- `reasoning` - AI service with LLM integration

**Ultra-Simple Pattern** (3/8 services):
- `collect-agent` - Edge agent with direct initialization
- `gateway` - API gateway, stateless
- `monitor` - Monitoring service, stateless

### Quick Decision Guide

```
1. Count external dependencies:
   - Database
   - Redis
   - NATS
   - gRPC
   - Email
   - External APIs

   Count >= 5?  → Use BOOTSTRAP PATTERN
   Count < 5?   → Check step 2

2. Are there complex initialization interdependencies?
   (e.g., Service A needs initialized ServiceB, which needs initialized ServiceC)

   Yes? → Use BOOTSTRAP PATTERN
   No?  → Use ULTRA-SIMPLE PATTERN

3. Does the service coordinate multiple feature domains?
   (e.g., auth has sessions, audit, notifications, forced-logout)

   Yes? → Use BOOTSTRAP PATTERN
   No?  → Use ULTRA-SIMPLE PATTERN
```

## Bootstrap Pattern (Complex Services)

### Structure

```
cmd/{service}/
├── main.go                          # Entry point
└── app/
    ├── app.go                       # Application implementation (200+ lines)
    ├── components.go                # Optional: Component registration helpers
    └── wire.go                      # Optional: Wire dependency injection

internal/{service}/
├── startup/                         # Component initializers
│   ├── infrastructure.go            # Database, Redis, external clients
│   ├── core_services.go             # Main business logic
│   ├── servers.go                   # HTTP/gRPC servers
│   └── [feature]_services.go        # Feature-specific services
├── api/                             # HTTP handlers
├── grpc/                            # gRPC implementations
├── service/                         # Business logic
├── storage/                         # Data access layer
└── [domain]/                        # Domain models
```

### Core Components

#### 1. Entry Point (`cmd/{service}/main.go`)

Minimal, consistent across all services:

```go
package main

import (
    _ "go.uber.org/automaxprocs/maxprocs"
    "github.com/kart-io/k8s-agent/cmd/{service}/app"
)

func main() {
    app.Execute()
}
```

**Note**: The `automaxprocs` import ensures the Go runtime properly detects CPU quotas in containers.

#### 2. Application File (`cmd/{service}/app/app.go`)

Contains the service startup logic. Key sections:

**A. Execute Function**

```go
func Execute() {
    // 1. Create options with required capabilities
    opts := commonapp.NewStandardOptions("Service Name", "user-agent").
        WithDatabase().
        WithRedis().
        WithNATS().
        WithMetrics()

    // 2. Create app instance
    app := &{Service}App{}

    // 3. Run with Bootstrap framework
    commonapp.RunWithBootstrap(
        app,
        opts,
        commonapp.Config{
            Use:       "{service}",
            Short:     "{Service} Service",
            Long:      "Description",
            EnvPrefix: "SERVICE_NAME",
        },
        app.registerComponents,
    )
}
```

**StandardOptions Capabilities**:
- `.WithServer()` - HTTP server (port 8080+, health checks)
- `.WithDatabase()` - MySQL connection pooling
- `.WithRedis()` - Redis client
- `.WithNATS()` - NATS messaging
- `.WithGRPC()` - gRPC server
- `.WithJWT()` - JWT authentication
- `.WithEmail()` - Email client
- `.WithMetrics()` - Prometheus metrics
- `.WithAgent()` - Agent configuration
- `.WithCORS()` - CORS middleware

**B. Application Struct**

```go
type {Service}App struct {
    bootstrap *bootstrap.Bootstrap        // Manages component lifecycle
    opts      *commonapp.StandardOptions  // Configuration
    logger    core.Logger                 // Logger instance
}
```

**C. Initialize Method**

```go
func (a *{Service}App) Initialize(ctx context.Context, opts commonapp.Options) error {
    a.opts = opts.(*commonapp.StandardOptions)

    // Initialize logger FIRST
    logger, err := a.opts.InitLogger()
    if err != nil {
        return fmt.Errorf("failed to initialize logger: %w", err)
    }
    a.logger = logger

    return nil
}
```

**Important**: Initialize the logger before anything else. It's needed for debugging subsequent initialization failures.

**D. Run Method**

```go
func (a *{Service}App) Run(ctx context.Context) error {
    // Bootstrap.Run() is already called by RunWithBootstrap
    // Just wait for shutdown signal
    <-ctx.Done()
    return ctx.Err()
}
```

The Bootstrap framework handles component startup automatically. The app just waits for context cancellation.

**E. Shutdown Method**

```go
func (a *{Service}App) Shutdown(ctx context.Context) error {
    // Bootstrap shutdown is handled automatically
    return nil
}
```

The Bootstrap framework handles component shutdown in reverse registration order.

**F. registerComponents Method**

Defines initialization order using priority-based layering:

```go
func (a *{Service}App) registerComponents(bs *bootstrap.Bootstrap) error {
    a.bootstrap = bs

    // LAYER 1: Infrastructure (Priority 300-500)
    // Database and Redis clients
    infra := startup.NewInfrastructureInitializers(a.opts, a.logger)
    bs.Register(infra.Database)
    bs.Register(infra.Redis)

    // LAYER 2: Core Business Services (Priority 600-700)
    // Main service logic
    coreServices := startup.NewCoreServicesInitializer(
        a.opts, a.logger, infra.Database, infra.Redis,
    )
    bs.Register(coreServices)

    // LAYER 3: Advanced Services (Priority 650-800) - optional
    if a.opts.NATS.Enable {
        natsInit := startup.NewNATSInitializer(a.opts, a.logger, coreServices)
        bs.Register(natsInit)
    }

    // LAYER 4: Servers (Priority 900-1000)
    if a.opts.GRPC.Enable {
        grpcInit := startup.NewGRPCServerInitializer(
            a.opts, a.logger, infra, coreServices,
        )
        bs.Register(grpcInit)
    }

    httpInit := startup.NewHTTPServerInitializer(
        a.opts, a.logger, infra, coreServices,
    )
    bs.Register(httpInit)

    // LAYER 5: Monitoring (Priority 2000)
    healthInit := pkginitializers.NewHealthCheckInitializer(
        a.opts.Health, a.logger,
    )
    bs.Register(healthInit)

    return nil
}
```

### Initialization Priority Guidelines

```
Priority 100-200:   Reserved for framework
Priority 300-400:   Primary infrastructure (Database)
Priority 350-400:   Secondary infrastructure (Redis)
Priority 450-500:   Tertiary infrastructure (Email, external APIs)
Priority 600-700:   Core services (repositories, main business logic)
Priority 650-800:   Feature services (sessions, audit, advanced features)
Priority 800-900:   Specialized services (NATS, event bus)
Priority 900-950:   gRPC servers
Priority 950-1000:  HTTP servers
Priority 1000+:     Post-server services
Priority 2000:      Monitoring (health checks, metrics)
```

### Service-Specific Startup Package

Create `internal/{service}/startup/` with grouped initializers:

#### infrastructure.go

```go
package startup

import (
    "github.com/kart-io/k8s-agent/internal/{service}/storage"
    commonapp "github.com/kart-io/k8s-agent/pkg/app"
    pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
    "github.com/kart-io/k8s-agent/pkg/types"
    "github.com/kart-io/logger/core"
)

// InfrastructureInitializers contains database and external service clients
type InfrastructureInitializers struct {
    Database *DatabaseInitializer
    Redis    *RedisInitializer
    Email    *EmailInitializer  // if needed
}

func NewInfrastructureInitializers(
    opts *commonapp.StandardOptions,
    logger core.Logger,
) *InfrastructureInitializers {
    // Database with auto-migration
    dbInit := &DatabaseInitializer{
        DatabaseInitializer: pkginitializers.NewDatabaseInitializer(
            opts.Database, logger,
        ),
    }
    if opts.Database.AutoMigrate {
        dbInit.WithAutoMigrate(
            &types.Agent{},
            &types.Event{},
            &types.Command{},
            // All models for auto-migration
        )
    }

    // Redis client
    redisInit := &RedisInitializer{
        RedisInitializer: pkginitializers.NewRedisInitializer(
            opts.Redis, logger,
        ),
    }

    return &InfrastructureInitializers{
        Database: dbInit,
        Redis:    redisInit,
    }
}

// DatabaseInitializer wraps generic initializer with service-specific config
type DatabaseInitializer struct {
    *pkginitializers.DatabaseInitializer
    store *storage.MySQLStore
}

func (d *DatabaseInitializer) Store() *storage.MySQLStore {
    if d.store == nil && d.Client() != nil {
        d.store = &storage.MySQLStore{
            MySQLClient: d.Client(),
        }
    }
    return d.store
}

// RedisInitializer wraps generic initializer with service-specific config
type RedisInitializer struct {
    *pkginitializers.RedisInitializer
    store *storage.RedisStore
}

func (r *RedisInitializer) Store() *storage.RedisStore {
    if r.store == nil && r.RedisClient() != nil {
        r.store = &storage.RedisStore{
            RedisClient: r.RedisClient(),
        }
    }
    return r.store
}
```

#### core_services.go

```go
package startup

import (
    "context"
    "fmt"

    "github.com/kart-io/k8s-agent/internal/{service}/service"
    commonapp "github.com/kart-io/k8s-agent/pkg/app"
    "github.com/kart-io/k8s-agent/pkg/bootstrap"
    "github.com/kart-io/logger/core"
)

// CoreServicesInitializer creates main business logic services
type CoreServicesInitializer struct {
    *bootstrap.BaseInitializer
    logger             core.Logger
    infraDbInit        *DatabaseInitializer
    infraRedisInit     *RedisInitializer

    // Service instances
    agentService      *service.AgentService
    eventService      *service.EventService
    commandService    *service.CommandService
}

func NewCoreServicesInitializer(
    opts *commonapp.StandardOptions,
    logger core.Logger,
    dbInit *DatabaseInitializer,
    redisInit *RedisInitializer,
) *CoreServicesInitializer {
    return &CoreServicesInitializer{
        BaseInitializer: bootstrap.NewBaseInitializer("CoreServices", 600),
        logger:          logger,
        infraDbInit:     dbInit,
        infraRedisInit:  redisInit,
    }
}

func (c *CoreServicesInitializer) Initialize(ctx context.Context) error {
    c.logger.Infow("Initializing core services")

    // Create services in dependency order
    store := c.infraDbInit.Store()
    redisStore := c.infraRedisInit.Store()

    c.agentService = service.NewAgentService(store, c.logger)
    c.eventService = service.NewEventService(store, redisStore, c.logger)
    c.commandService = service.NewCommandService(store, c.logger)

    c.logger.Info("Core services initialized")
    return nil
}

func (c *CoreServicesInitializer) Close(ctx context.Context) error {
    // Cleanup if needed
    return nil
}

// Accessor methods for dependent components
func (c *CoreServicesInitializer) AgentService() *service.AgentService {
    return c.agentService
}

func (c *CoreServicesInitializer) EventService() *service.EventService {
    return c.eventService
}

func (c *CoreServicesInitializer) CommandService() *service.CommandService {
    return c.commandService
}
```

#### servers.go

```go
package startup

import (
    "context"
    "fmt"

    "github.com/kart-io/k8s-agent/internal/{service}/api"
    "github.com/kart-io/k8s-agent/internal/{service}/grpc"
    commonapp "github.com/kart-io/k8s-agent/pkg/app"
    "github.com/kart-io/k8s-agent/pkg/bootstrap"
    "github.com/kart-io/logger/core"
    "github.com/gin-gonic/gin"
)

// HTTPServerInitializer creates and manages the HTTP server
type HTTPServerInitializer struct {
    *bootstrap.BaseInitializer
    logger   core.Logger
    opts     *commonapp.StandardOptions
    coreInit *CoreServicesInitializer
    server   *gin.Engine
}

func NewHTTPServerInitializer(
    opts *commonapp.StandardOptions,
    logger core.Logger,
    coreInit *CoreServicesInitializer,
) *HTTPServerInitializer {
    return &HTTPServerInitializer{
        BaseInitializer: bootstrap.NewBaseInitializer("HTTPServer", 950),
        logger:          logger,
        opts:            opts,
        coreInit:        coreInit,
    }
}

func (h *HTTPServerInitializer) Initialize(ctx context.Context) error {
    h.logger.Infow("Initializing HTTP server",
        "port", h.opts.Server.Port,
    )

    engine := gin.New()

    // Register routes
    apiHandler := api.NewHandler(
        h.coreInit.AgentService(),
        h.coreInit.EventService(),
        h.logger,
    )
    apiHandler.RegisterRoutes(engine)

    h.server = engine

    // Start server in background
    go func() {
        addr := fmt.Sprintf(":%d", h.opts.Server.Port)
        if err := h.server.Run(addr); err != nil {
            h.logger.Errorw("HTTP server error", "error", err)
        }
    }()

    h.logger.Info("HTTP server started")
    return nil
}

func (h *HTTPServerInitializer) Close(ctx context.Context) error {
    h.logger.Info("Closing HTTP server")
    // Shutdown logic
    return nil
}

// GRPCServerInitializer creates and manages the gRPC server
type GRPCServerInitializer struct {
    *bootstrap.BaseInitializer
    logger   core.Logger
    opts     *commonapp.StandardOptions
    coreInit *CoreServicesInitializer
    // server implementation
}

func NewGRPCServerInitializer(
    opts *commonapp.StandardOptions,
    logger core.Logger,
    coreInit *CoreServicesInitializer,
) *GRPCServerInitializer {
    return &GRPCServerInitializer{
        BaseInitializer: bootstrap.NewBaseInitializer("GRPCServer", 900),
        logger:          logger,
        opts:            opts,
        coreInit:        coreInit,
    }
}

func (g *GRPCServerInitializer) Initialize(ctx context.Context) error {
    g.logger.Infow("Initializing gRPC server",
        "port", g.opts.GRPC.Port,
    )

    // Create and start gRPC server
    // ...

    return nil
}

func (g *GRPCServerInitializer) Close(ctx context.Context) error {
    g.logger.Info("Closing gRPC server")
    return nil
}
```

### Example: agent-manager Service

See `/Users/costalong/code/go/src/github.com/kart/k8s-agent/cmd/agent-manager/app/app.go` for a complete example with:
- 5 initialization layers
- 6+ initializers
- Database with auto-migration
- NATS integration
- HTTP and gRPC servers
- Comprehensive error handling

## Ultra-Simple Pattern (Simple Services)

### Structure

```
cmd/{service}/
├── main.go          # Entry point (5 lines)
└── app/
    └── app.go       # Application implementation (100-150 lines)

internal/{service}/
├── agent/           # Main service logic
├── config/          # Configuration
├── types/           # Domain types
└── health/          # Health checks
```

### Implementation Guide

#### 1. Main Entry Point

```go
package main

import (
    _ "go.uber.org/automaxprocs/maxprocs"
    "github.com/kart-io/k8s-agent/cmd/{service}/app"
)

func main() {
    app.Execute()
}
```

#### 2. Application File

```go
package app

import (
    "context"
    "fmt"

    "github.com/kart-io/k8s-agent/internal/{service}/agent"
    "github.com/kart-io/k8s-agent/internal/{service}/types"
    commonapp "github.com/kart-io/k8s-agent/pkg/app"
    "github.com/kart-io/logger/core"
)

const (
    UserAgent = "aetherius-{service}"
)

// Execute runs the service command.
func Execute() {
    // Create options with only required capabilities
    opts := commonapp.NewStandardOptions("{Service}", UserAgent).
        WithAgent()

    // Create application instance
    app := &{Service}App{}

    // Run with standard Run() - no Bootstrap framework
    commonapp.Run(
        app,
        opts,
        commonapp.Config{
            Use:       "{service}",
            Short:     "{Service}",
            Long:      "Description of {Service}",
            EnvPrefix: "{SERVICE_NAME}",
        },
    )
}

// {Service}App implements commonapp.Application interface.
type {Service}App struct {
    opts   *commonapp.StandardOptions
    logger core.Logger
    // Service-specific components
    agent  *agent.Agent
    server *HealthServer
}

func (a *{Service}App) Name() string {
    return "{Service}"
}

func (a *{Service}App) Initialize(ctx context.Context, opts commonapp.Options) error {
    a.opts = opts.(*commonapp.StandardOptions)

    // Initialize logger first
    logger, err := a.opts.InitLogger()
    if err != nil {
        return fmt.Errorf("failed to initialize logger: %w", err)
    }
    a.logger = logger

    a.logger.Infow("Initializing {Service}",
        "cluster_id", a.opts.Agent.ClusterID,
        "health_port", a.opts.Health.Port,
    )

    // Create agent instance
    agentConfig := &types.AgentConfig{
        ClusterID:         a.opts.Agent.ClusterID,
        CentralEndpoint:   a.opts.Agent.CentralEndpoint,
        // ... other fields
    }

    a.agent, err = agent.New(agentConfig, logger)
    if err != nil {
        return fmt.Errorf("failed to create agent: %w", err)
    }

    // Create health check server
    a.server = NewHealthServer(a.opts.Health, logger)
    if err := a.server.Start(ctx); err != nil {
        return fmt.Errorf("failed to start health server: %w", err)
    }

    return nil
}

func (a *{Service}App) Run(ctx context.Context) error {
    a.logger.Info("Starting {Service}")
    // Agent.Start() blocks until context is cancelled
    return a.agent.Start(ctx)
}

func (a *{Service}App) Shutdown(ctx context.Context) error {
    a.logger.Info("Shutting down {Service}")

    // Shutdown in reverse order of initialization
    if a.server != nil {
        if err := a.server.Stop(ctx); err != nil {
            a.logger.Warnw("Health server shutdown error", "error", err)
        }
    }

    // Agent cleanup is handled by context cancellation
    a.logger.Info("{Service} shutdown complete")
    return nil
}
```

### Key Differences from Bootstrap Pattern

| Aspect | Bootstrap | Ultra-Simple |
|--------|-----------|--------------|
| **Initializers** | Multiple files in startup/ | Single app.go |
| **Priority** | Explicit integer priority | Linear initialization |
| **Dependencies** | Passed through initializer chain | Direct instantiation |
| **Complexity** | Abstract via layers | Direct and obvious |
| **Shutdown Order** | Automatic reverse registration | Manual in Shutdown() |
| **Testing** | Mock entire Bootstrap framework | Mock individual components |

## Step-by-Step Implementation

### Adding a New Bootstrap Service

#### Step 1: Create Directory Structure

```bash
mkdir -p cmd/myservice/app
mkdir -p internal/myservice/startup
mkdir -p internal/myservice/{api,service,storage,grpc}
```

#### Step 2: Create main.go

```go
package main

import (
    _ "go.uber.org/automaxprocs/maxprocs"
    "github.com/kart-io/k8s-agent/cmd/myservice/app"
)

func main() {
    app.Execute()
}
```

#### Step 3: Create app/app.go

```go
package app

import (
    "context"
    "fmt"

    "github.com/kart-io/k8s-agent/internal/myservice/startup"
    commonapp "github.com/kart-io/k8s-agent/pkg/app"
    "github.com/kart-io/k8s-agent/pkg/bootstrap"
    pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
    "github.com/kart-io/logger/core"
)

const (
    UserAgent = "aetherius-myservice"
)

func Execute() {
    opts := commonapp.NewStandardOptions("MyService", UserAgent).
        WithDatabase().
        WithRedis()

    app := &MyServiceApp{}

    commonapp.RunWithBootstrap(
        app,
        opts,
        commonapp.Config{
            Use:       "myservice",
            Short:     "My Service",
            Long:      "Description",
            EnvPrefix: "MYSERVICE",
        },
        app.registerComponents,
    )
}

type MyServiceApp struct {
    bootstrap *bootstrap.Bootstrap
    opts      *commonapp.StandardOptions
    logger    core.Logger
}

func (a *MyServiceApp) Name() string {
    return "My Service"
}

func (a *MyServiceApp) Initialize(ctx context.Context, opts commonapp.Options) error {
    a.opts = opts.(*commonapp.StandardOptions)
    logger, err := a.opts.InitLogger()
    if err != nil {
        return fmt.Errorf("failed to initialize logger: %w", err)
    }
    a.logger = logger
    return nil
}

func (a *MyServiceApp) Run(ctx context.Context) error {
    <-ctx.Done()
    return ctx.Err()
}

func (a *MyServiceApp) Shutdown(ctx context.Context) error {
    return nil
}

func (a *MyServiceApp) registerComponents(bs *bootstrap.Bootstrap) error {
    a.bootstrap = bs

    // Layer 1: Infrastructure
    infra := startup.NewInfrastructureInitializers(a.opts, a.logger)
    bs.Register(infra.Database)
    bs.Register(infra.Redis)

    // Layer 2: Core Services
    coreInit := startup.NewCoreServicesInitializer(
        a.opts, a.logger, infra.Database, infra.Redis,
    )
    bs.Register(coreInit)

    // Layer 4: Servers
    httpInit := startup.NewHTTPServerInitializer(
        a.opts, a.logger, coreInit,
    )
    bs.Register(httpInit)

    // Layer 5: Monitoring
    healthInit := pkginitializers.NewHealthCheckInitializer(
        a.opts.Health, a.logger,
    )
    bs.Register(healthInit)

    return nil
}
```

#### Step 4: Create startup/infrastructure.go

```go
package startup

import (
    "github.com/kart-io/k8s-agent/internal/myservice/storage"
    commonapp "github.com/kart-io/k8s-agent/pkg/app"
    pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
    "github.com/kart-io/k8s-agent/pkg/types"
    "github.com/kart-io/logger/core"
)

type InfrastructureInitializers struct {
    Database *DatabaseInitializer
    Redis    *RedisInitializer
}

func NewInfrastructureInitializers(
    opts *commonapp.StandardOptions,
    logger core.Logger,
) *InfrastructureInitializers {
    dbInit := &DatabaseInitializer{
        DatabaseInitializer: pkginitializers.NewDatabaseInitializer(
            opts.Database, logger,
        ),
    }
    if opts.Database.AutoMigrate {
        // Add your models
        dbInit.WithAutoMigrate(
            &types.Agent{},
            &types.Event{},
        )
    }

    redisInit := &RedisInitializer{
        RedisInitializer: pkginitializers.NewRedisInitializer(
            opts.Redis, logger,
        ),
    }

    return &InfrastructureInitializers{
        Database: dbInit,
        Redis:    redisInit,
    }
}

type DatabaseInitializer struct {
    *pkginitializers.DatabaseInitializer
    store *storage.MySQLStore
}

func (d *DatabaseInitializer) Store() *storage.MySQLStore {
    if d.store == nil && d.Client() != nil {
        d.store = &storage.MySQLStore{MySQLClient: d.Client()}
    }
    return d.store
}

type RedisInitializer struct {
    *pkginitializers.RedisInitializer
    store *storage.RedisStore
}

func (r *RedisInitializer) Store() *storage.RedisStore {
    if r.store == nil && r.RedisClient() != nil {
        r.store = &storage.RedisStore{RedisClient: r.RedisClient()}
    }
    return r.store
}
```

#### Step 5: Create startup/core_services.go

```go
package startup

import (
    "context"
    "fmt"

    "github.com/kart-io/k8s-agent/internal/myservice/service"
    commonapp "github.com/kart-io/k8s-agent/pkg/app"
    "github.com/kart-io/k8s-agent/pkg/bootstrap"
    "github.com/kart-io/logger/core"
)

type CoreServicesInitializer struct {
    *bootstrap.BaseInitializer
    logger     core.Logger
    dbInit     *DatabaseInitializer
    redisInit  *RedisInitializer

    myService  *service.MyService
}

func NewCoreServicesInitializer(
    opts *commonapp.StandardOptions,
    logger core.Logger,
    dbInit *DatabaseInitializer,
    redisInit *RedisInitializer,
) *CoreServicesInitializer {
    return &CoreServicesInitializer{
        BaseInitializer: bootstrap.NewBaseInitializer("CoreServices", 600),
        logger:          logger,
        dbInit:          dbInit,
        redisInit:       redisInit,
    }
}

func (c *CoreServicesInitializer) Initialize(ctx context.Context) error {
    c.logger.Info("Initializing core services")

    store := c.dbInit.Store()
    redisStore := c.redisInit.Store()

    c.myService = service.New(store, redisStore, c.logger)

    return nil
}

func (c *CoreServicesInitializer) Close(ctx context.Context) error {
    return nil
}

func (c *CoreServicesInitializer) MyService() *service.MyService {
    return c.myService
}
```

#### Step 6: Create startup/servers.go

```go
package startup

import (
    "context"
    "fmt"

    "github.com/kart-io/k8s-agent/internal/myservice/api"
    commonapp "github.com/kart-io/k8s-agent/pkg/app"
    "github.com/kart-io/k8s-agent/pkg/bootstrap"
    "github.com/kart-io/logger/core"
    "github.com/gin-gonic/gin"
)

type HTTPServerInitializer struct {
    *bootstrap.BaseInitializer
    logger   core.Logger
    opts     *commonapp.StandardOptions
    coreInit *CoreServicesInitializer
}

func NewHTTPServerInitializer(
    opts *commonapp.StandardOptions,
    logger core.Logger,
    coreInit *CoreServicesInitializer,
) *HTTPServerInitializer {
    return &HTTPServerInitializer{
        BaseInitializer: bootstrap.NewBaseInitializer("HTTPServer", 950),
        logger:          logger,
        opts:            opts,
        coreInit:        coreInit,
    }
}

func (h *HTTPServerInitializer) Initialize(ctx context.Context) error {
    h.logger.Infow("Initializing HTTP server", "port", h.opts.Server.Port)

    engine := gin.New()
    handler := api.NewHandler(h.coreInit.MyService(), h.logger)
    handler.RegisterRoutes(engine)

    go func() {
        addr := fmt.Sprintf(":%d", h.opts.Server.Port)
        if err := engine.Run(addr); err != nil {
            h.logger.Errorw("HTTP server error", "error", err)
        }
    }()

    return nil
}

func (h *HTTPServerInitializer) Close(ctx context.Context) error {
    return nil
}
```

#### Step 7: Update Makefile

Add your service to the `SERVICES` variable in the root Makefile:

```makefile
SERVICES ?= agent-manager orchestrator auth gateway collect-agent monitor cluster reasoning myservice
```

#### Step 8: Build and Test

```bash
# Build the service
make go.build.myservice

# Run the service
make run-myservice

# Test the service
make go.test.myservice
```

### Adding an Ultra-Simple Service

#### Step 1: Create Directory Structure

```bash
mkdir -p cmd/myagent/app
mkdir -p internal/myagent/{agent,types}
```

#### Step 2: Create main.go and app/app.go

Follow the template from `/docs/templates/service_startup_template.go` (Ultra-Simple Pattern section).

#### Step 3: Add to Makefile

```makefile
SERVICES ?= ... myagent
```

## Common Patterns

### Pattern 1: Service with Optional NATS Support

```go
// In registerComponents()
if a.opts.NATS.Enable {
    natsInit := startup.NewNATSInitializer(a.opts, a.logger, coreInit)
    bs.Register(natsInit)
}
```

### Pattern 2: Service with Both HTTP and gRPC

```go
// Register gRPC first (priority 900)
if a.opts.GRPC.Enable {
    grpcInit := startup.NewGRPCServerInitializer(...)
    bs.Register(grpcInit)
}

// Register HTTP after (priority 950)
httpInit := startup.NewHTTPServerInitializer(...)
bs.Register(httpInit)
```

### Pattern 3: Service with Database Migration

```go
dbInit := &DatabaseInitializer{
    DatabaseInitializer: pkginitializers.NewDatabaseInitializer(
        opts.Database, logger,
    ),
}

// Add auto-migration models
if opts.Database.AutoMigrate {
    dbInit.WithAutoMigrate(
        &Model1{},
        &Model2{},
        &Model3{},
    )
}

bs.Register(dbInit)
```

### Pattern 4: Accessing Configuration in registerComponents

```go
func (a *MyServiceApp) registerComponents(bs *bootstrap.Bootstrap) error {
    // All options are available via a.opts
    if a.opts.Server.EnableMetrics {
        // Register metrics
    }

    if a.opts.Database.AutoMigrate {
        // Setup migrations
    }

    // Use environment overrides
    // DATABASE_HOST=custom-host override in config

    return nil
}
```

### Pattern 5: Feature Toggles

```go
// In registerComponents()
if a.opts.NATS.Enable {
    natsInit := startup.NewNATSInitializer(...)
    bs.Register(natsInit)
}

if a.opts.GRPC.Enable {
    grpcInit := startup.NewGRPCServerInitializer(...)
    bs.Register(grpcInit)
}

// Services always registered
coreInit := startup.NewCoreServicesInitializer(...)
bs.Register(coreInit)
```

## Testing Strategies

### Bootstrap Pattern Testing

```go
func TestMyService_Bootstrap(t *testing.T) {
    // Create test logger
    logger := zap.NewNop()

    // Create test options
    opts := &commonapp.StandardOptions{
        Server: commonapp.ServerOptions{Port: 8080},
        Database: commonapp.MySQLOptions{
            Host: "localhost",
            Port: 3306,
            // ...
        },
        // ...
    }

    // Create app
    app := &MyServiceApp{}

    // Create bootstrap
    bs := bootstrap.New(logger)

    // Register components
    err := app.registerComponents(bs)
    require.NoError(t, err)

    // Initialize all components
    ctx := context.Background()
    err = bs.Initialize(ctx)
    require.NoError(t, err)

    // Verify initialization
    // ... assertions ...

    // Shutdown
    err = bs.Shutdown(ctx, 5*time.Second)
    require.NoError(t, err)
}
```

### Ultra-Simple Pattern Testing

```go
func TestMyAgent_Initialize(t *testing.T) {
    logger := zap.NewNop()
    opts := &commonapp.StandardOptions{
        Agent: commonapp.AgentOptions{
            ClusterID: "test-cluster",
            // ...
        },
        // ...
    }

    app := &MyAgentApp{}
    err := app.Initialize(context.Background(), opts)
    require.NoError(t, err)

    // Verify components
    require.NotNil(t, app.agent)
    require.NotNil(t, app.server)

    // Test Run (should block)
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()

    err = app.Run(ctx)
    require.Equal(t, context.DeadlineExceeded, err)

    // Test Shutdown
    err = app.Shutdown(context.Background())
    require.NoError(t, err)
}
```

## Troubleshooting

### Issue 1: "cannot find package" errors

**Cause**: Initialization dependencies not in correct order
**Solution**: Move initializer to later priority level and ensure dependencies are registered first

```go
// WRONG: HTTP server depends on core services (600) but registered at 400
httpInit := startup.NewHTTPServerInitializer(...)
bs.Register(httpInit)  // Priority 950 - will fail

// RIGHT: Register core services first
coreInit := startup.NewCoreServicesInitializer(...)
bs.Register(coreInit)  // Priority 600
httpInit := startup.NewHTTPServerInitializer(...)
bs.Register(httpInit)  // Priority 950
```

### Issue 2: Service hangs during shutdown

**Cause**: Components not properly implementing Closer interface
**Solution**: Ensure all components that open resources implement Close()

```go
// Add to initializer
func (i *MyInitializer) Close(ctx context.Context) error {
    if i.resource != nil {
        return i.resource.Close()
    }
    return nil
}
```

### Issue 3: Configuration not applied

**Cause**: Feature not enabled in options
**Solution**: Check options and enable required features

```go
// app/app.go
opts := commonapp.NewStandardOptions("Name", "ua").
    WithDatabase().     // Must enable to use database
    WithRedis()         // Must enable to use Redis
    // ...
```

### Issue 4: Logger not available early

**Cause**: Logger initialization deferred too late
**Solution**: Initialize logger first in Initialize() method

```go
func (a *MyApp) Initialize(ctx context.Context, opts commonapp.Options) error {
    a.opts = opts.(*commonapp.StandardOptions)

    // ALWAYS first
    logger, err := a.opts.InitLogger()
    if err != nil {
        return fmt.Errorf("failed to initialize logger: %w", err)
    }
    a.logger = logger

    // Use logger after
    a.logger.Info("Initializing...")
}
```

### Issue 5: Circular dependency

**Cause**: ServiceA depends on ServiceB which depends on ServiceA
**Solution**: Refactor to break the cycle

```go
// Use interface-based dependencies
type ServiceADependency interface {
    GetData() string
}

type ServiceA struct {
    dep ServiceADependency
}

// Inject through initializer, not struct
```

## Checklists

### New Bootstrap Service Checklist

- [ ] Create `cmd/{service}/main.go`
- [ ] Create `cmd/{service}/app/app.go` with Execute() function
- [ ] Create `cmd/{service}/app/app.go` with Application implementation
- [ ] Create `internal/{service}/startup/infrastructure.go`
- [ ] Create `internal/{service}/startup/core_services.go`
- [ ] Create `internal/{service}/startup/servers.go`
- [ ] Implement registerComponents() with 5 layers
- [ ] Set correct Priority() values
- [ ] Implement Initialize() in all initializers
- [ ] Implement Close() in all initializers
- [ ] Add service to SERVICES in Makefile
- [ ] Test: `make build`
- [ ] Test: `make run-{service}`
- [ ] Write unit tests for all services
- [ ] Test shutdown gracefully with SIGTERM

### New Ultra-Simple Service Checklist

- [ ] Create `cmd/{service}/main.go`
- [ ] Create `cmd/{service}/app/app.go` with Execute() function
- [ ] Create `cmd/{service}/app/app.go` with Application implementation
- [ ] Implement Initialize() method
- [ ] Implement Run() method
- [ ] Implement Shutdown() method
- [ ] Add service to SERVICES in Makefile
- [ ] Test: `make build`
- [ ] Test: `make run-{service}`
- [ ] Write unit tests
- [ ] Test shutdown gracefully with SIGTERM

### Code Review Checklist

- [ ] Logger initialized first in Initialize()
- [ ] Correct initialization priority values
- [ ] All resources implement Close() if needed
- [ ] Shutdown() properly cleans up resources
- [ ] Configuration options properly declared with WithXXX()
- [ ] Error handling consistent (wrapped with context)
- [ ] No circular dependencies between initializers
- [ ] Tests cover happy path and error cases
- [ ] Health checks implemented for complex services
- [ ] Documentation updated if service adds new options

---

## Related Files

- [/docs/templates/service_startup_template.go](/docs/templates/service_startup_template.go) - Code templates
- [/CLAUDE.md](/CLAUDE.md) - Project guidelines
- [/cmd/agent-manager/app/app.go](/cmd/agent-manager/app/app.go) - Bootstrap pattern example
- [/cmd/auth/app/app.go](/cmd/auth/app/app.go) - Bootstrap pattern with features
- [/cmd/collect-agent/app/app.go](/cmd/collect-agent/app/app.go) - Ultra-simple pattern example
- [/pkg/bootstrap/bootstrap.go](/pkg/bootstrap/bootstrap.go) - Bootstrap framework
- [/pkg/app/base.go](/pkg/app/base.go) - Application interface

