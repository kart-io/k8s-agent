# Initializer Usage Audit Report

**Date**: 2025-11-09
**Objective**: Audit and unify the usage of pkg/initializers across all services to eliminate duplicate code

## Executive Summary

The audit reveals that **3 out of 8 services** still have duplicate implementations of common initializers that should be replaced with pkg/initializers versions. The other 5 services are already properly using pkg/initializers.

### Overall Status

| Service | Database | Redis | NATS | HTTP Server | gRPC Server | Status |
|---------|----------|-------|------|-------------|-------------|--------|
| **monitor** | ✅ Wrapper | ✅ Wrapper | N/A | ✅ Wrapper | ✅ Wrapper | **COMPLIANT** |
| **auth** | ✅ Direct | ✅ Direct | N/A | ✅ Wrapper | ✅ Wrapper | **COMPLIANT** |
| **agent-manager** | ✅ Wrapper | ✅ Wrapper | ✅ Wrapper | ✅ Wrapper | ✅ Wrapper | **COMPLIANT** |
| **orchestrator** | ✅ Wrapper | ✅ Wrapper | ❌ Custom | ✅ Wrapper | ✅ Wrapper | **NEEDS WORK** |
| **gateway** | N/A | ❌ Custom | N/A | ❌ Custom | N/A | **NEEDS WORK** |
| **collect-agent** | N/A | N/A | ✅ Direct | N/A | N/A | **COMPLIANT** |
| **reasoning** | N/A | N/A | N/A | ✅ Direct | N/A | **COMPLIANT** |
| **cluster** | ✅ Direct | ✅ Direct | N/A | ✅ Direct | ✅ Direct | **COMPLIANT** |

**Legend**:
- ✅ **Direct**: Directly uses pkg/initializers (no wrapper)
- ✅ **Wrapper**: Uses pkg/initializers with service-specific wrapper (acceptable pattern)
- ❌ **Custom**: Has custom implementation that duplicates pkg/initializers functionality
- N/A: Service doesn't use this infrastructure component

## Detailed Findings

### 1. COMPLIANT Services (5/8 - 62.5%)

These services are already properly using pkg/initializers:

#### 1.1 Monitor Service ✅
**Location**: `internal/monitor/initializers/`

**Status**: FULLY COMPLIANT

**Implementation Pattern**: Service-specific wrappers around pkg/initializers

```go
// Properly wraps pkg/initializers.DatabaseInitializer
type DatabaseInitializer struct {
    *pkginitializers.DatabaseInitializer
    opts   *commonapp.StandardOptions
    logger core.Logger
    store  *storage.MySQLStorage
}

// Properly wraps pkg/initializers.RedisInitializer
type RedisInitializer struct {
    *pkginitializers.RedisInitializer
    logger core.Logger
    store  *storage.RedisStorage
}

// Properly wraps pkg/initializers.HTTPServerInitializer
type HTTPServerInitializer struct {
    *pkginitializers.HTTPServerInitializer
    cfg        *commonapp.StandardOptions
    logger     core.Logger
    dbInit     *DatabaseInitializer
    redisInit  *RedisInitializer
    monitorSvc *service.MonitorService
}

// Properly wraps pkg/initializers.GRPCServerInitializer
type GRPCServerInitializer struct {
    standardInit *pkginitializers.GRPCServerInitializer
    opts         *commonapp.StandardOptions
    logger       core.Logger
    dbInit       *DatabaseInitializer
    redisInit    *RedisInitializer
}
```

**Design Pattern**:
- Uses composition (embedding) to extend base initializers
- Adds service-specific storage wrappers
- Maintains backward compatibility with existing code
- No code duplication

#### 1.2 Auth Service ✅
**Location**: `internal/auth/startup/infrastructure.go`

**Status**: FULLY COMPLIANT

**Implementation Pattern**: Direct usage of pkg/initializers

```go
type InfrastructureInitializers struct {
    Database *pkginitializers.DatabaseInitializer  // ✅ Direct usage
    Redis    *pkginitializers.RedisInitializer     // ✅ Direct usage
    Email    *EmailClientInitializer                // Custom (email-specific)
}

func NewInfrastructureInitializers(opts *commonapp.StandardOptions, logger core.Logger) *InfrastructureInitializers {
    dbInit := pkginitializers.NewDatabaseInitializer(opts.Database, logger)
    redisInit := pkginitializers.NewRedisInitializer(opts.Redis, logger)
    emailInit := &EmailClientInitializer{opts: opts, logger: logger}

    return &InfrastructureInitializers{
        Database: dbInit,
        Redis:    redisInit,
        Email:    emailInit,
    }
}
```

**Servers**: Properly wraps pkg/initializers for HTTP and gRPC servers

```go
// internal/auth/startup/servers.go
type GRPCServerInitializer struct {
    opts         *commonapp.StandardOptions
    logger       core.Logger
    coreServices *CoreServicesInitializer
    sessionInit  *SessionServiceInitializer
    standardInit *pkginitializers.GRPCServerInitializer  // ✅ Uses pkg version
}

type HTTPServerInitializer struct {
    opts         *commonapp.StandardOptions
    logger       core.Logger
    coreServices *CoreServicesInitializer
    standardInit *pkginitializers.HTTPServerInitializer  // ✅ Uses pkg version
}
```

#### 1.3 Agent-Manager Service ✅
**Location**: `internal/agent-manager/initializers/`

**Status**: FULLY COMPLIANT

**Implementation Pattern**: Service-specific wrappers with storage abstraction

```go
// internal/agent-manager/initializers/database.go
type DatabaseInitializer struct {
    *pkginitializers.DatabaseInitializer  // ✅ Embeds pkg version
    store *storage.MySQLStore
}

// internal/agent-manager/initializers/redis.go
type RedisInitializer struct {
    *pkginitializers.RedisInitializer  // ✅ Embeds pkg version
    store *storage.RedisStore
}

// internal/agent-manager/initializers/servers.go
type HTTPServerInitializer struct {
    standardInit *pkginitializers.HTTPServerInitializer  // ✅ Uses pkg version
    logger       core.Logger
    opts         *commonapp.StandardOptions
    apiServer    *api.Server
    serviceInit  *ServiceInitializer
    dbInit       *DatabaseInitializer
    redisInit    *RedisInitializer
}

type GRPCServerInitializer struct {
    standardInit *pkginitializers.GRPCServerInitializer  // ✅ Uses pkg version
    logger       core.Logger
    opts         *commonapp.StandardOptions
    serviceInit  *ServiceInitializer
}
```

#### 1.4 Collect-Agent Service ✅
**Location**: `internal/collect-agent/`

**Status**: FULLY COMPLIANT

**Pattern**: Simple Pattern service, directly uses NATS client (no custom initializer)

**Note**: Collect-agent follows the Simple Pattern and doesn't use the Bootstrap framework. It has minimal dependencies (only NATS) and doesn't need custom initializers.

#### 1.5 Reasoning Service ✅
**Location**: `cmd/reasoning/app/app.go`

**Status**: FULLY COMPLIANT

**Pattern**: Simple Pattern service, directly creates HTTP server

**Note**: Reasoning service follows the Simple Pattern with no external dependencies beyond HTTP server. It doesn't use the Bootstrap framework.

### 2. NEEDS WORK Services (2/8 - 25%)

These services have duplicate implementations that should be replaced:

#### 2.1 Orchestrator Service ⚠️
**Location**: `internal/orchestrator/startup/infrastructure.go`

**Issues Found**: 1 duplicate initializer

##### Issue 1: Custom NATSInitializer (Lines 102-165)

**Problem**: Orchestrator has a custom NATSInitializer that duplicates functionality from pkg/initializers.NATSInitializer

**Current Implementation**:
```go
// internal/orchestrator/startup/infrastructure.go:102-165
type NATSInitializer struct {
    opts   *commonapp.StandardOptions
    logger core.Logger
    conn   *nats.Conn
}

func (n *NATSInitializer) Name() string { return "nats" }
func (n *NATSInitializer) Priority() int { return bootstrap.PriorityMQ }

func (n *NATSInitializer) Initialize(ctx context.Context) error {
    conn, err := nats.Connect(
        n.opts.NATS.URL,
        nats.Name("orchestrator-service"),
        nats.MaxReconnects(n.opts.NATS.MaxReconnect),
        nats.ReconnectWait(n.opts.NATS.ReconnectWait),
    )
    // ... rest of implementation
}
```

**Available in pkg/initializers**:
```go
// pkg/initializers/nats.go
type NATSInitializer struct {
    opts   *options.NATSOptions
    logger core.Logger
    conn   *nats.Conn
}

func NewNATSInitializer(opts *options.NATSOptions, logger core.Logger) *NATSInitializer
func (n *NATSInitializer) Name() string
func (n *NATSInitializer) Priority() int
func (n *NATSInitializer) Initialize(ctx context.Context) error
func (n *NATSInitializer) Close(ctx context.Context) error
func (n *NATSInitializer) HealthCheck(ctx context.Context) error
func (n *NATSInitializer) Connection() *nats.Conn
func (n *NATSInitializer) Conn() *nats.Conn
func (n *NATSInitializer) Publish(subject string, data []byte) error
func (n *NATSInitializer) Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error)
```

**Recommendation**:
```go
// internal/orchestrator/startup/infrastructure.go
type InfrastructureInitializers struct {
    Database *DatabaseInitializer
    Redis    *RedisInitializer
    NATS     *pkginitializers.NATSInitializer  // ✅ Use pkg version directly
}

func NewInfrastructureInitializers(opts *commonapp.StandardOptions, logger core.Logger) *InfrastructureInitializers {
    dbInit := &DatabaseInitializer{...}
    redisInit := &RedisInitializer{...}
    natsInit := pkginitializers.NewNATSInitializer(opts.NATS, logger)  // ✅ Use pkg factory

    return &InfrastructureInitializers{
        Database: dbInit,
        Redis:    redisInit,
        NATS:     natsInit,
    }
}
```

**Impact**:
- Lines of code saved: ~60 LOC
- Improved maintainability: NATS initialization logic maintained in one place
- Reduced bugs: All services benefit from bug fixes to pkg/initializers

#### 2.2 Gateway Service ⚠️
**Location**: `cmd/gateway/app/container.go`

**Issues Found**: 2 duplicate initializers

##### Issue 1: Custom RedisInitializer (Lines 48-113)

**Problem**: Gateway has a custom RedisInitializer that duplicates functionality from pkg/initializers.RedisInitializer

**Current Implementation**:
```go
// cmd/gateway/app/container.go:48-113
type RedisInitializer struct {
    opts   *options.ServerOptions
    logger core.Logger
    client *redis.Client
}

func (r *RedisInitializer) Name() string { return "gateway-redis" }
func (r *RedisInitializer) Priority() int { return bootstrap.PriorityRedis }

func (r *RedisInitializer) Initialize(ctx context.Context) error {
    if r.opts.Redis == nil || r.opts.Redis.Addr == "" {
        r.logger.Info("Redis is not configured")
        return nil
    }

    r.client = redis.NewClient(&redis.Options{
        Addr:         r.opts.Redis.Addr,
        Password:     r.opts.Redis.Password,
        DB:           r.opts.Redis.DB,
        PoolSize:     r.opts.Redis.PoolSize,
        MinIdleConns: r.opts.Redis.MinIdleConns,
        DialTimeout:  r.opts.Redis.DialTimeout,
        ReadTimeout:  r.opts.Redis.ReadTimeout,
        WriteTimeout: r.opts.Redis.WriteTimeout,
    })

    // Test connection (non-fatal)
    if err := r.client.Ping(ctx).Err(); err != nil {
        r.logger.Warnw("Failed to connect to Redis (rate limiting will use local mode)", "error", err)
        r.client.Close()
        r.client = nil
        return nil
    }

    r.logger.Infow("Redis connected", "addr", r.opts.Redis.Addr)
    return nil
}

func (r *RedisInitializer) Close(ctx context.Context) error {
    if r.client != nil {
        return r.client.Close()
    }
    return nil
}

func (r *RedisInitializer) Client() *redis.Client {
    return r.client
}
```

**Available in pkg/initializers**:
```go
// pkg/initializers/redis.go
type RedisInitializer struct {
    opts   *options.RedisOptions
    logger core.Logger
    client *db.RedisClient
}

func NewRedisInitializer(opts *options.RedisOptions, logger core.Logger) *RedisInitializer
func (r *RedisInitializer) Name() string
func (r *RedisInitializer) Priority() int
func (r *RedisInitializer) Initialize(ctx context.Context) error
func (r *RedisInitializer) Close(ctx context.Context) error
func (r *RedisInitializer) HealthCheck(ctx context.Context) error
func (r *RedisInitializer) Client() *redis.Client
func (r *RedisInitializer) RedisClient() *db.RedisClient
```

**Recommendation**:
```go
// cmd/gateway/app/container.go
type GatewayContainer struct {
    Redis  *pkginitializers.RedisInitializer           // ✅ Use pkg version
    HTTP   *HTTPServerInitializer
    Health *pkginitializers.HealthCheckInitializer
}

func NewGatewayContainer(
    redis *pkginitializers.RedisInitializer,            // ✅ Use pkg version
    http *HTTPServerInitializer,
    health *pkginitializers.HealthCheckInitializer,
) *GatewayContainer {
    return &GatewayContainer{
        Redis:  redis,
        HTTP:   http,
        Health: health,
    }
}
```

**Special Note**: Gateway's Redis is optional (non-fatal if unavailable). The pkg/initializers.RedisInitializer can be made optional by:
1. Checking if opts.Redis is nil before creating initializer
2. Or wrapping the pkg version with optional initialization logic

##### Issue 2: Custom HTTPServerInitializer (Lines 119-323)

**Problem**: Gateway has a custom HTTPServerInitializer that duplicates functionality from pkg/initializers.HTTPServerInitializer

**Current Implementation**:
```go
// cmd/gateway/app/container.go:119-323
type HTTPServerInitializer struct {
    opts         *options.ServerOptions
    logger       core.Logger
    redisInit    *RedisInitializer
    ginServer    *commonserver.GinServer
    proxyHandler *proxy.Proxy
}

func (h *HTTPServerInitializer) Name() string { return "gateway-http-server" }
func (h *HTTPServerInitializer) Priority() int { return bootstrap.PriorityHTTP }

func (h *HTTPServerInitializer) Initialize(ctx context.Context) error {
    // Initialize Redis-based rate limiter
    if h.redisInit.Client() != nil {
        middleware.InitRateLimiter(h.redisInit.Client())
    }

    // Create proxy handler
    h.proxyHandler = proxy.NewProxy(h.logger)

    // Create Gin server config
    config := commonserver.NewGinServerOptions(h.opts.Server)
    if h.opts.CORS != nil {
        config.WithCORS(h.opts.CORS)
    }

    h.ginServer = commonserver.NewGinServerFromFullConfig(h.logger, config)

    // Setup routes
    h.setupRoutes(h.ginServer.Engine)

    return nil
}

func (h *HTTPServerInitializer) setupRoutes(engine *gin.Engine) {
    // ... 150+ lines of route setup code
}
```

**Available in pkg/initializers**:
```go
// pkg/initializers/http_server.go
type HTTPServerInitializer struct {
    config *HTTPServerConfig
    logger core.Logger
    server commonserver.Server
}

type HTTPServerConfig struct {
    Name       string
    Priority   int
    Config     *options.ServerOptions
    RouteSetup func(*gin.Engine) error  // ✅ Callback for service-specific routes
    CORS       *options.CORSOptions
    JWT        *options.JWTOptions
    RateLimit  *options.RateLimitOptions
}

func NewHTTPServerInitializer(config *HTTPServerConfig, logger core.Logger) *HTTPServerInitializer
```

**Recommendation**:
```go
// cmd/gateway/app/container.go
type HTTPServerInitializer struct {
    *pkginitializers.HTTPServerInitializer  // ✅ Embed pkg version
    opts         *options.ServerOptions
    logger       core.Logger
    redisInit    *pkginitializers.RedisInitializer
    proxyHandler *proxy.Proxy
}

func NewHTTPServerInitializer(
    opts *options.ServerOptions,
    logger core.Logger,
    redisInit *pkginitializers.RedisInitializer,
) *HTTPServerInitializer {
    h := &HTTPServerInitializer{
        opts:      opts,
        logger:    logger,
        redisInit: redisInit,
    }

    // Create standard HTTP server config
    serverConfig := &pkginitializers.HTTPServerConfig{
        Name:       "gateway-http-server",
        Priority:   bootstrap.PriorityHTTP,
        Config:     opts.Server,
        RouteSetup: h.setupRoutes,  // ✅ Delegate route setup
        CORS:       opts.CORS,
    }

    h.HTTPServerInitializer = pkginitializers.NewHTTPServerInitializer(serverConfig, logger)
    return h
}

// Initialize initializes gateway-specific components before server starts
func (h *HTTPServerInitializer) Initialize(ctx context.Context) error {
    // Initialize Redis-based rate limiter (gateway-specific)
    if h.redisInit.Client() != nil {
        middleware.InitRateLimiter(h.redisInit.Client())
    }

    // Create proxy handler (gateway-specific)
    h.proxyHandler = proxy.NewProxy(h.logger)

    // Initialize standard HTTP server (calls setupRoutes callback)
    return h.HTTPServerInitializer.Initialize(ctx)
}

// setupRoutes configures gateway-specific routes
func (h *HTTPServerInitializer) setupRoutes(engine *gin.Engine) error {
    // ... existing route setup code (unchanged)
    return nil
}
```

**Impact**:
- Lines of code saved: ~80 LOC (server initialization logic)
- Improved consistency: Gateway HTTP server follows same pattern as other services
- Easier testing: Can mock pkg/initializers.HTTPServerInitializer

### 3. Cluster Service ✅
**Location**: N/A (uses pkg/initializers directly)

**Status**: FULLY COMPLIANT

**Note**: Cluster service was recently refactored to use Bootstrap pattern and directly uses pkg/initializers for all infrastructure components.

## Summary of Duplicate Code

### Total Duplicated Lines of Code (LOC)

| Service | Component | Duplicate LOC | Can Replace With |
|---------|-----------|---------------|------------------|
| orchestrator | NATSInitializer | ~60 | pkg/initializers.NATSInitializer |
| gateway | RedisInitializer | ~65 | pkg/initializers.RedisInitializer |
| gateway | HTTPServerInitializer | ~80 | pkg/initializers.HTTPServerInitializer |
| **TOTAL** | | **~205 LOC** | |

### Code Duplication Metrics

- **Services with duplicates**: 2 out of 8 (25%)
- **Services compliant**: 6 out of 8 (75%)
- **Total duplicate LOC**: ~205 lines
- **Estimated refactoring time**: 2-3 hours
- **Risk level**: LOW (wrappers can be replaced incrementally)

## pkg/initializers Inventory

### Available Common Initializers

1. **DatabaseInitializer** (`pkg/initializers/database.go`)
   - Supports MySQL with GORM
   - Auto-migration support
   - Health checks
   - Connection pooling configuration

2. **RedisInitializer** (`pkg/initializers/redis.go`)
   - Standard Redis client initialization
   - Health checks
   - Connection pooling configuration

3. **NATSInitializer** (`pkg/initializers/nats.go`)
   - NATS connection with auto-reconnect
   - Event handlers for disconnect/reconnect
   - Publish/Subscribe convenience methods
   - Health checks

4. **HTTPServerInitializer** (`pkg/initializers/http_server.go`)
   - Gin-based HTTP server
   - Middleware configuration (CORS, JWT, RateLimit)
   - Route setup via callback
   - Graceful shutdown

5. **GRPCServerInitializer** (`pkg/initializers/grpc_server.go`)
   - Standard gRPC server
   - Service registration via callback
   - Graceful shutdown

6. **HealthCheckInitializer** (`pkg/initializers/health.go`)
   - HTTP health check endpoint
   - Integrates with Bootstrap lifecycle

### Design Patterns Used

#### Pattern 1: Direct Usage (Recommended for simple cases)
```go
type InfrastructureInitializers struct {
    Database *pkginitializers.DatabaseInitializer  // Direct usage
    Redis    *pkginitializers.RedisInitializer     // Direct usage
}
```

**Used by**: auth service, cluster service

#### Pattern 2: Service-Specific Wrapper (Recommended for custom storage abstraction)
```go
type DatabaseInitializer struct {
    *pkginitializers.DatabaseInitializer  // Embed base initializer
    store *storage.MySQLStore              // Service-specific storage wrapper
}

func (d *DatabaseInitializer) Store() *storage.MySQLStore {
    if d.store == nil && d.Client() != nil {
        d.store = &storage.MySQLStore{
            MySQLClient: d.Client(),
        }
    }
    return d.store
}
```

**Used by**: agent-manager, monitor, orchestrator

**Rationale**:
- Maintains backward compatibility with existing storage layer
- Adds service-specific methods while reusing core initialization
- Zero duplication of initialization logic

#### Pattern 3: Custom Initializer (Use only when necessary)
```go
type CustomComponentInitializer struct {
    opts   *options.CustomOptions
    logger core.Logger
    client *CustomClient
}
```

**Used by**:
- EmailClientInitializer (auth) - Email is not a standard infrastructure component
- EventSubscriberInitializer (orchestrator) - Business logic specific to orchestrator

**When to use**: Only when the component is truly service-specific and cannot be generalized

## Recommended Actions

### Priority 1: Immediate (High Impact, Low Risk)

1. **Replace orchestrator NATSInitializer** (60 LOC saved)
   - File: `internal/orchestrator/startup/infrastructure.go`
   - Replace custom NATSInitializer with `pkginitializers.NATSInitializer`
   - Update EventSubscriber to use `pkginitializers.NATSInitializer.Conn()`
   - Test: Run orchestrator tests and verify NATS connectivity

2. **Replace gateway RedisInitializer** (65 LOC saved)
   - File: `cmd/gateway/app/container.go`
   - Replace custom RedisInitializer with `pkginitializers.RedisInitializer`
   - Handle optional Redis (non-fatal failure) in Wire provider
   - Test: Verify rate limiting works with and without Redis

3. **Replace gateway HTTPServerInitializer** (80 LOC saved)
   - File: `cmd/gateway/app/container.go`
   - Wrap `pkginitializers.HTTPServerInitializer` with gateway-specific logic
   - Move route setup to callback function
   - Test: Verify all gateway routes and middleware work correctly

### Priority 2: Documentation (Medium Impact, Low Risk)

4. **Update CLAUDE.md**
   - Add section on initializer usage patterns
   - Document when to use direct vs wrapper pattern
   - Add examples from compliant services

5. **Create pkg/initializers/README.md**
   - Document all available initializers
   - Provide usage examples for each
   - Explain design patterns (Direct, Wrapper, Custom)

### Priority 3: Prevention (Low Impact, High Value)

6. **Add linting rule**
   - Create custom golangci-lint rule to detect duplicate initializer patterns
   - Check for direct Redis/NATS client creation in service initializers
   - Warn when new initializers are created instead of using pkg versions

## Testing Strategy

### Unit Tests
- Test that pkg/initializers work with service-specific wrappers
- Test optional initialization (e.g., gateway Redis)
- Test health checks for all infrastructure components

### Integration Tests
- Test orchestrator with pkg/initializers.NATSInitializer
- Test gateway with pkg/initializers.RedisInitializer
- Verify all services start and stop gracefully

### Regression Tests
- Verify no behavior changes after refactoring
- Test NATS reconnection in orchestrator
- Test Redis failover in gateway (local fallback)

## Success Criteria

1. ✅ All services use pkg/initializers for common infrastructure (Database, Redis, NATS, HTTP, gRPC)
2. ✅ No duplicate initializer implementations exist
3. ✅ All services pass integration tests
4. ✅ Code coverage maintained or improved
5. ✅ Documentation updated to reflect best practices

## Risk Assessment

### Low Risk
- Replacing orchestrator NATSInitializer (pkg version has identical functionality)
- Replacing gateway RedisInitializer (pkg version supports all required features)

### Medium Risk
- Replacing gateway HTTPServerInitializer (complex route setup logic)
  - Mitigation: Thorough testing of all gateway routes
  - Mitigation: Incremental rollout with feature flags

### High Risk
- None identified

## Estimated Effort

| Task | Effort | Risk |
|------|--------|------|
| Replace orchestrator NATSInitializer | 30 min | Low |
| Replace gateway RedisInitializer | 30 min | Low |
| Replace gateway HTTPServerInitializer | 60 min | Medium |
| Testing all services | 60 min | Low |
| Update documentation | 30 min | Low |
| **TOTAL** | **3.5 hours** | **Low-Medium** |

## Conclusion

The audit reveals that **75% of services (6/8)** are already properly using pkg/initializers, with only **2 services needing refactoring**. The refactoring will:

- Eliminate ~205 lines of duplicate code
- Improve maintainability across all services
- Establish clear patterns for future development
- Reduce bugs by centralizing infrastructure initialization

The refactoring is **low risk** and **high value**, with an estimated effort of 3.5 hours. All identified duplicates can be safely replaced with pkg/initializers versions without changing behavior.

## Appendix A: Service Architecture Summary

| Service | Pattern | Complexity | External Deps | Initializers | pkg Usage |
|---------|---------|------------|---------------|--------------|-----------|
| agent-manager | Bootstrap | High | MySQL, Redis, NATS | 6+ | ✅ Wrapper |
| orchestrator | Bootstrap | High | MySQL, Redis, NATS | 5+ | ⚠️ 1 Custom |
| auth | Bootstrap | Medium-High | MySQL, Redis | 8+ | ✅ Direct |
| cluster | Bootstrap | High | MySQL | 3 | ✅ Direct |
| reasoning | Simple | High | LLM APIs | 1 | ✅ Direct |
| monitor | Bootstrap | Medium | MySQL, Redis | 4 | ✅ Wrapper |
| collect-agent | Simple | Medium | NATS | 1 | ✅ Direct |
| gateway | Bootstrap | Low | Redis (opt) | 2+ | ⚠️ 2 Custom |

## Appendix B: File Locations

### Compliant Services

**Monitor**:
- `internal/monitor/initializers/database.go` - Wrapper around pkg/initializers
- `internal/monitor/initializers/redis.go` - Wrapper around pkg/initializers
- `internal/monitor/initializers/http_server.go` - Wrapper around pkg/initializers
- `internal/monitor/initializers/grpc.go` - Wrapper around pkg/initializers

**Auth**:
- `internal/auth/startup/infrastructure.go` - Direct usage of pkg/initializers
- `internal/auth/startup/servers.go` - Wrappers around pkg/initializers

**Agent-Manager**:
- `internal/agent-manager/initializers/database.go` - Wrapper around pkg/initializers
- `internal/agent-manager/initializers/redis.go` - Wrapper around pkg/initializers
- `internal/agent-manager/initializers/servers.go` - Wrappers around pkg/initializers

### Non-Compliant Services

**Orchestrator**:
- `internal/orchestrator/startup/infrastructure.go:102-165` - ❌ Custom NATSInitializer

**Gateway**:
- `cmd/gateway/app/container.go:48-113` - ❌ Custom RedisInitializer
- `cmd/gateway/app/container.go:119-323` - ❌ Custom HTTPServerInitializer

## Appendix C: pkg/initializers API Reference

### DatabaseInitializer
```go
func NewDatabaseInitializer(opts *options.MySQLOptions, logger core.Logger) *DatabaseInitializer
func (d *DatabaseInitializer) WithAutoMigrate(models ...interface{}) *DatabaseInitializer
func (d *DatabaseInitializer) Name() string
func (d *DatabaseInitializer) Priority() int
func (d *DatabaseInitializer) Initialize(ctx context.Context) error
func (d *DatabaseInitializer) Close(ctx context.Context) error
func (d *DatabaseInitializer) HealthCheck(ctx context.Context) error
func (d *DatabaseInitializer) DB() *gorm.DB
func (d *DatabaseInitializer) Client() *db.MySQLClient
func (d *DatabaseInitializer) Store() interface{}
```

### RedisInitializer
```go
func NewRedisInitializer(opts *options.RedisOptions, logger core.Logger) *RedisInitializer
func (r *RedisInitializer) Name() string
func (r *RedisInitializer) Priority() int
func (r *RedisInitializer) Initialize(ctx context.Context) error
func (r *RedisInitializer) Close(ctx context.Context) error
func (r *RedisInitializer) HealthCheck(ctx context.Context) error
func (r *RedisInitializer) Client() *redis.Client
func (r *RedisInitializer) RedisClient() *db.RedisClient
```

### NATSInitializer
```go
func NewNATSInitializer(opts *options.NATSOptions, logger core.Logger) *NATSInitializer
func (n *NATSInitializer) Name() string
func (n *NATSInitializer) Priority() int
func (n *NATSInitializer) Initialize(ctx context.Context) error
func (n *NATSInitializer) Close(ctx context.Context) error
func (n *NATSInitializer) HealthCheck(ctx context.Context) error
func (n *NATSInitializer) Connection() *nats.Conn
func (n *NATSInitializer) Conn() *nats.Conn
func (n *NATSInitializer) Publish(subject string, data []byte) error
func (n *NATSInitializer) Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error)
```

### HTTPServerInitializer
```go
type HTTPServerConfig struct {
    Name       string
    Priority   int
    Config     *options.ServerOptions
    RouteSetup func(*gin.Engine) error
    CORS       *options.CORSOptions
    JWT        *options.JWTOptions
    RateLimit  *options.RateLimitOptions
}

func NewHTTPServerInitializer(config *HTTPServerConfig, logger core.Logger) *HTTPServerInitializer
func (i *HTTPServerInitializer) Name() string
func (i *HTTPServerInitializer) Priority() int
func (i *HTTPServerInitializer) Initialize(ctx context.Context) error
func (i *HTTPServerInitializer) GetServer() commonserver.Server
func (i *HTTPServerInitializer) Close(ctx context.Context) error
```

### GRPCServerInitializer
```go
type GRPCServerConfig struct {
    Name            string
    Priority        int
    Config          *options.GRPCOptions
    ServiceRegister func(*grpc.Server) error
}

func NewGRPCServerInitializer(config *GRPCServerConfig, logger core.Logger) *GRPCServerInitializer
func (i *GRPCServerInitializer) Name() string
func (i *GRPCServerInitializer) Priority() int
func (i *GRPCServerInitializer) Initialize(ctx context.Context) error
func (i *GRPCServerInitializer) GetServer() commonserver.Server
func (i *GRPCServerInitializer) Close(ctx context.Context) error
```
