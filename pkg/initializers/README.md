# pkg/initializers - Quick Reference Guide

**Location**: `/Users/costalong/code/go/src/github.com/kart/k8s-agent/pkg/initializers`

**Purpose**: Centralized infrastructure component initializers that eliminate code duplication across all services.

**Last Updated**: 2025-11-09 (After Orchestrator & Gateway unification)

## Table of Contents

- [Available Initializers](#available-initializers)
  - [DatabaseInitializer](#1-databaseinitializer)
  - [RedisInitializer](#2-redisinitializer)
  - [NATSInitializer](#3-natsinitializer)
  - [HTTPServerInitializer](#4-httpserverinitializer)
  - [GRPCServerInitializer](#5-grpcserverinitializer)
  - [HealthCheckInitializer](#6-healthcheckinitializer)
- [Usage Patterns](#usage-patterns)
- [Common Recipes](#common-recipes)
- [Troubleshooting](#troubleshooting)
- [Migration Guide](#migration-guide)
- [Best Practices](#best-practices)
- [Examples](#examples-in-the-wild)

## Available Initializers

### 1. DatabaseInitializer

**File**: `database.go`

**When to use**: Service needs MySQL database

**Import**:
```go
import pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
```

**Basic Usage**:
```go
dbInit := pkginitializers.NewDatabaseInitializer(opts.Database, logger)

// Optional: Enable auto-migration
dbInit.WithAutoMigrate(&models.User{}, &models.Post{})

// Access database
db := dbInit.DB()  // *gorm.DB
client := dbInit.Client()  // *db.MySQLClient
```

**Methods**:
- `Initialize(ctx)` - Connect to database
- `Close(ctx)` - Graceful shutdown
- `HealthCheck(ctx)` - Check connection health
- `DB()` - Get GORM DB instance
- `Client()` - Get MySQL client with helpers
- `WithAutoMigrate(models...)` - Enable auto-migration

---

### 2. RedisInitializer

**File**: `redis.go`

**When to use**: Service needs Redis cache/sessions

**Basic Usage**:
```go
redisInit := pkginitializers.NewRedisInitializer(opts.Redis, logger)

// Access Redis client
client := redisInit.Client()  // *redis.Client
redisClient := redisInit.RedisClient()  // *db.RedisClient
```

**Methods**:
- `Initialize(ctx)` - Connect to Redis
- `Close(ctx)` - Graceful shutdown
- `HealthCheck(ctx)` - Check connection health
- `Client()` - Get Redis client
- `RedisClient()` - Get wrapped Redis client

---

### 3. NATSInitializer

**File**: `nats.go`

**When to use**: Service needs message queue

**Basic Usage**:
```go
natsInit := pkginitializers.NewNATSInitializer(opts.NATS, logger)

// Publish message
natsInit.Publish("subject", []byte("data"))

// Subscribe to topic
sub, err := natsInit.Subscribe("subject.*", func(msg *nats.Msg) {
    // Handle message
})

// Access connection
conn := natsInit.Conn()  // *nats.Conn
```

**Methods**:
- `Initialize(ctx)` - Connect to NATS
- `Close(ctx)` - Graceful drain and shutdown
- `HealthCheck(ctx)` - Check connection health
- `Conn()` / `Connection()` - Get NATS connection
- `Publish(subject, data)` - Publish message
- `Subscribe(subject, handler)` - Subscribe to topic

**Features**:
- Auto-reconnect on disconnect
- Event handlers (disconnect/reconnect/close)
- Configurable timeouts and buffer sizes

---

### 4. HTTPServerInitializer

**File**: `http_server.go`

**When to use**: Service needs HTTP/REST API server

**Configuration**:
```go
type HTTPServerConfig struct {
    Name       string                     // Initializer name
    Priority   int                        // Initialization priority
    Config     *options.ServerOptions     // Server host/port/timeouts
    RouteSetup func(*gin.Engine) error   // Route setup callback
    CORS       *options.CORSOptions       // CORS configuration
    JWT        *options.JWTOptions        // JWT middleware
    RateLimit  *options.RateLimitOptions  // Rate limiting
}
```

**Basic Usage**:
```go
func NewHTTPServerInitializer(opts *commonapp.StandardOptions, logger core.Logger) *HTTPServerInitializer {
    serverConfig := &pkginitializers.HTTPServerConfig{
        Name:       "my-http-server",
        Priority:   bootstrap.PriorityHTTP,
        Config:     opts.Server,
        RouteSetup: setupRoutes,  // Your route setup function
        CORS:       opts.CORS,
        JWT:        opts.JWT,
    }

    return pkginitializers.NewHTTPServerInitializer(serverConfig, logger)
}

// Route setup callback
func setupRoutes(engine *gin.Engine) error {
    engine.GET("/health", healthHandler)
    engine.POST("/api/v1/users", createUserHandler)
    return nil
}
```

**Methods**:
- `Initialize(ctx)` - Create and configure server (calls RouteSetup callback)
- `Close(ctx)` - Graceful shutdown
- `GetServer()` - Get server instance (for Bootstrap)

**Middleware Included**:
- Recovery (panic handler)
- Logger (request logging)
- CORS (if configured)
- JWT (if configured)
- Rate Limit (if configured)
- Request ID

---

### 5. GRPCServerInitializer

**File**: `grpc_server.go`

**When to use**: Service needs gRPC API server

**Configuration**:
```go
type GRPCServerConfig struct {
    Name            string                     // Initializer name
    Priority        int                        // Initialization priority
    Config          *options.GRPCOptions       // Server host/port
    ServiceRegister func(*grpc.Server) error  // Service registration callback
}
```

**Basic Usage**:
```go
func NewGRPCServerInitializer(opts *commonapp.StandardOptions, logger core.Logger) *GRPCServerInitializer {
    serverConfig := &pkginitializers.GRPCServerConfig{
        Name:     "my-grpc-server",
        Priority: bootstrap.PriorityGRPC,
        Config:   opts.GRPC,
        ServiceRegister: func(s *grpc.Server) error {
            myv1.RegisterMyServiceServer(s, myServiceImpl)
            return nil
        },
    }

    return pkginitializers.NewGRPCServerInitializer(serverConfig, logger)
}
```

**Methods**:
- `Initialize(ctx)` - Create and start gRPC server (calls ServiceRegister callback)
- `Close(ctx)` - Graceful shutdown
- `GetServer()` - Get server instance (for Bootstrap)

---

### 6. HealthCheckInitializer

**File**: `health.go`

**When to use**: All services (health endpoint)

**Basic Usage**:
```go
healthInit := pkginitializers.NewHealthCheckInitializer(opts.Health, logger)

// Bootstrap automatically manages lifecycle
```

**Methods**:
- `Initialize(ctx)` - Start health check HTTP server
- `Shutdown(ctx)` - Stop health check server

**Default Endpoint**: `http://localhost:8090/health` (configurable)

---

## Usage Patterns

### Pattern 1: Direct Usage

**Use when**: Service doesn't need custom storage abstraction

```go
import pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"

type InfrastructureInitializers struct {
    Database *pkginitializers.DatabaseInitializer
    Redis    *pkginitializers.RedisInitializer
    NATS     *pkginitializers.NATSInitializer
}

func NewInfrastructureInitializers(opts *commonapp.StandardOptions, logger core.Logger) *InfrastructureInitializers {
    return &InfrastructureInitializers{
        Database: pkginitializers.NewDatabaseInitializer(opts.Database, logger),
        Redis:    pkginitializers.NewRedisInitializer(opts.Redis, logger),
        NATS:     pkginitializers.NewNATSInitializer(opts.NATS, logger),
    }
}

// Use directly
func (i *InfrastructureInitializers) GetDB() *gorm.DB {
    return i.Database.DB()
}
```

**Services using this pattern**: auth, cluster, reasoning, orchestrator (NATS)

---

### Pattern 2: Service-Specific Wrapper

**Use when**:
- Service has existing storage layer
- Need additional service-specific methods
- Backward compatibility required

```go
import pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"

type DatabaseInitializer struct {
    *pkginitializers.DatabaseInitializer  // Embed base
    store *storage.MySQLStore              // Service-specific wrapper
}

func NewDatabaseInitializer(opts *commonapp.StandardOptions, logger core.Logger) *DatabaseInitializer {
    dbInit := pkginitializers.NewDatabaseInitializer(opts.Database, logger)

    // Configure auto-migration
    if opts.Database.AutoMigrate {
        dbInit.WithAutoMigrate(&types.Agent{}, &types.Event{})
    }

    return &DatabaseInitializer{
        DatabaseInitializer: dbInit,
    }
}

// Service-specific method
func (d *DatabaseInitializer) Store() *storage.MySQLStore {
    if d.store == nil && d.Client() != nil {
        d.store = &storage.MySQLStore{
            MySQLClient: d.Client(),
        }
    }
    return d.store
}
```

**Services using this pattern**: agent-manager, monitor, gateway

---

## Common Recipes

### Recipe 1: Database with Auto-Migration

```go
dbInit := pkginitializers.NewDatabaseInitializer(opts.Database, logger)
dbInit.WithAutoMigrate(
    &types.Agent{},
    &types.Event{},
    &types.Command{},
)
```

### Recipe 2: Optional Redis (Non-Fatal)

```go
type RedisInitializer struct {
    *pkginitializers.RedisInitializer
    isAvailable bool
}

func (r *RedisInitializer) Initialize(ctx context.Context) error {
    if r.RedisInitializer == nil {
        return nil  // Not configured
    }

    if err := r.RedisInitializer.Initialize(ctx); err != nil {
        logger.Warn("Redis unavailable, using fallback")
        return nil  // Non-fatal
    }

    r.isAvailable = true
    return nil
}

func (r *RedisInitializer) IsAvailable() bool {
    return r.isAvailable && r.Client() != nil
}
```

**Used by**: gateway service

### Recipe 3: HTTP Server with Custom Middleware

```go
func setupRoutes(engine *gin.Engine) error {
    // Add custom middleware
    engine.Use(middleware.CustomAuth())
    engine.Use(middleware.AuditLog())

    // Register routes
    api := engine.Group("/api/v1")
    api.GET("/users", getUsersHandler)
    api.POST("/users", createUserHandler)

    return nil
}

serverConfig := &pkginitializers.HTTPServerConfig{
    Name:       "my-service",
    Priority:   bootstrap.PriorityHTTP,
    Config:     opts.Server,
    RouteSetup: setupRoutes,
    CORS:       opts.CORS,
}

httpInit := pkginitializers.NewHTTPServerInitializer(serverConfig, logger)
```

### Recipe 4: NATS Event Subscriber

```go
natsInit := pkginitializers.NewNATSInitializer(opts.NATS, logger)

// Subscribe during initialization
func subscribeToEvents(natsInit *pkginitializers.NATSInitializer) error {
    _, err := natsInit.Subscribe("events.*", func(msg *nats.Msg) {
        log.Infow("Received event", "subject", msg.Subject)
        // Process event
    })
    return err
}
```

### Recipe 5: Health Check with Custom Port

```go
healthOpts := &options.HealthOptions{
    Host:    "0.0.0.0",
    Port:    9090,  // Custom health check port
    Enabled: true,
}

healthInit := pkginitializers.NewHealthCheckInitializer(healthOpts, logger)
```

---

## Troubleshooting

### Issue: "database not initialized"

**Cause**: Trying to access database before `Initialize()` is called

**Solution**: Ensure Bootstrap calls initializers in order
```go
if err := bootstrap.Run(ctx); err != nil {
    log.Fatal(err)
}
```

### Issue: "Redis connection refused"

**Cause**: Redis server not running or incorrect configuration

**Solution**: Check Redis server and configuration
```go
// For optional Redis, handle gracefully
if redisInit.Client() == nil {
    log.Warn("Redis unavailable, using local cache")
    // Use fallback
}
```

### Issue: "NATS disconnected"

**Cause**: NATS server unavailable or network issues

**Solution**: NATSInitializer has auto-reconnect, check logs
```go
// Health check will show disconnected state
if err := natsInit.HealthCheck(ctx); err != nil {
    log.Error("NATS unhealthy", err)
}
```

### Issue: "Port already in use"

**Cause**: HTTP/gRPC server port conflict

**Solution**: Change port in configuration
```go
opts.Server.Port = 8081  // Use different port
```

---

## Migration Guide

### Before (Custom Initializer):

```go
type RedisInitializer struct {
    opts   *options.RedisOptions
    logger core.Logger
    client *redis.Client
}

func (r *RedisInitializer) Initialize(ctx context.Context) error {
    r.client = redis.NewClient(&redis.Options{
        Addr: r.opts.Addr,
        // ... 20+ lines of configuration
    })
    return r.client.Ping(ctx).Err()
}
```

### After (Using pkg/initializers):

```go
type RedisInitializer struct {
    *pkginitializers.RedisInitializer  // Embed
}

func NewRedisInitializer(opts *commonapp.StandardOptions, logger core.Logger) *RedisInitializer {
    return &RedisInitializer{
        RedisInitializer: pkginitializers.NewRedisInitializer(opts.Redis, logger),
    }
}
```

**Benefits**:
- 20+ lines → 5 lines
- Standard configuration
- Health checks included
- Graceful shutdown included

---

## Priority Constants

```go
const (
    PriorityHighest  = 0
    PriorityDatabase = 200
    PriorityCache    = 300  // Redis
    PriorityMQ       = 500  // NATS
    PriorityHTTP     = 900
    PriorityGRPC     = 900
    PriorityLowest   = 9999
)
```

Lower numbers initialize first.

---

## Best Practices

1. **Always use pkg/initializers for common infrastructure**
   - Database → DatabaseInitializer
   - Redis → RedisInitializer
   - NATS → NATSInitializer
   - HTTP → HTTPServerInitializer
   - gRPC → GRPCServerInitializer

2. **Choose the right pattern**
   - Direct: Simple services without custom storage
   - Wrapper: Services with existing storage layer

3. **Configure auto-migration carefully**
   - Only enable in development/test environments
   - Use migrations for production

4. **Handle optional dependencies gracefully**
   - Return nil/skip initialization for optional components
   - Provide fallback mechanisms

5. **Test initialization order**
   - Use Bootstrap priorities correctly
   - Test with Bootstrap.Run() in integration tests

6. **Monitor health checks**
   - All initializers provide HealthCheck()
   - Aggregate health in service health endpoint

---

## Examples in the Wild

- **agent-manager**: All 5 initializers (Database, Redis, NATS, HTTP, gRPC)
- **auth**: Database, Redis, HTTP, gRPC
- **orchestrator**: Database, Redis, NATS, HTTP, gRPC (unified 2025-11-09)
- **gateway**: Redis (optional), HTTP (unified 2025-11-09)
- **monitor**: Database, Redis, HTTP, gRPC
- **collect-agent**: NATS only
- **reasoning**: HTTP only

See service implementations in `internal/{service}/` directories.

---

## Service Coverage

| Service | Database | Redis | NATS | HTTP | gRPC | Status |
|---------|----------|-------|------|------|------|--------|
| agent-manager | ✅ | ✅ | ✅ | ✅ | ✅ | Compliant |
| orchestrator | ✅ | ✅ | ✅ | ✅ | ✅ | Compliant |
| auth | ✅ | ✅ | N/A | ✅ | ✅ | Compliant |
| cluster | ✅ | ✅ | N/A | ✅ | ✅ | Compliant |
| monitor | ✅ | ✅ | N/A | ✅ | ✅ | Compliant |
| reasoning | N/A | N/A | N/A | ✅ | N/A | Compliant |
| collect-agent | N/A | N/A | ✅ | N/A | N/A | Compliant |
| gateway | N/A | ✅ | N/A | ✅ | N/A | Compliant |

**Result**: 100% coverage - all 8 services use pkg/initializers

---

## Further Reading

- [INITIALIZER_UNIFICATION_SUMMARY.md](../../docs/refactoring/INITIALIZER_UNIFICATION_SUMMARY.md) - Migration history (2025-11-09)
- [INITIALIZER_AUDIT_REPORT.md](../../docs/refactoring/INITIALIZER_AUDIT_REPORT.md) - Detailed audit report
- [Bootstrap README](../bootstrap/README.md) - Bootstrap lifecycle management
- [CLAUDE.md](../../CLAUDE.md) - Project overview
