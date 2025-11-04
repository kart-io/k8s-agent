# Server Framework Standardization Analysis

## Current Situation

The project has a standardized server framework (`common/server/`) that provides:
- Unified server interface with `RunOrDie()` and `GracefulStop()` methods
- HTTP server implementation using Gin (`common/server/http/gin.go`)
- gRPC server implementation (`common/server/grpc/`)
- Lifecycle management with graceful shutdown
- Integration with bootstrap system via `pkg/initializers/HTTPServerInitializer`

However, most services are not using this framework and are instead managing servers manually.

## Server Framework Architecture

### Correct Pattern (Used by agent-manager, orchestrator)

```go
// Using pkg/initializers/HTTPServerInitializer
httpInit := NewHTTPServerInitializer(config, logger, dependencies...)
bootstrap.Register(httpInit)

// HTTPServerInitializer creates:
// 1. httpserver.GinServerFromFullConfig (implements Server interface)
// 2. Registers routes via RouteSetup callback
// 3. Server lifecycle managed by bootstrap framework
```

### Anti-Pattern (Used by auth, reasoning, cluster, etc.)

```go
// Direct server creation
server := &http.Server{
    Addr:    fmt.Sprintf("%s:%d", host, port),
    Handler: router,
}

// Manual lifecycle management
go func() {
    if err := server.ListenAndServe(); err != nil {
        // Manual error handling
    }
}()
```

## Services Analysis

### Services Using Common Server (Correct) ✅
1. **agent-manager** - Uses `pkg/initializers.HTTPServerInitializer`
2. **orchestrator** - Uses `pkg/initializers.HTTPServerInitializer` for both HTTP and gRPC

### Services NOT Using Common Server (Need Refactoring) ❌

#### Bootstrap Mode Services
1. **auth** - Creates `http.Server` directly in `internal/auth/initializers/server.go`
2. **reasoning** - Creates servers directly in `internal/reasoning/initializers/unified_server.go`
3. **cluster** - Creates `http.Server` directly in `internal/cluster/initializers/http_server.go`

#### Simple Mode Services
4. **gateway** - Creates `http.Server` directly in `cmd/gateway/app/server.go`
5. **monitor** - Creates `http.Server` directly in `cmd/monitor/app/server.go`
6. **collect-agent** - Uses `common/server` but not via standard initializer pattern

## Refactoring Plan

### Phase 1: Bootstrap Mode Services (auth, reasoning, cluster)

For each service:
1. Replace custom server creation with `pkg/initializers.HTTPServerInitializer`
2. Move route registration to `RouteSetup` callback
3. Remove manual server lifecycle management
4. Let bootstrap framework handle server start/stop

#### Example: Auth Service Refactoring

**Before:**
```go
// internal/auth/initializers/server.go
type HTTPServerInitializer struct {
    server *http.Server
    errChan chan error
}

func (h *HTTPServerInitializer) Initialize(ctx context.Context) error {
    router := gin.Default()
    // ... register routes ...

    h.server = &http.Server{
        Addr:    addr,
        Handler: router,
    }

    go func() {
        if err := h.server.ListenAndServe(); err != nil {
            h.errChan <- err
        }
    }()
}
```

**After:**
```go
// internal/auth/initializers/server.go
type HTTPServerInitializer struct {
    *pkg.initializers.HTTPServerInitializer
}

func NewHTTPServerInitializer(cfg *options.ServerOptions, ...) *HTTPServerInitializer {
    config := &pkg.initializers.HTTPServerConfig{
        Name:     "auth-http-server",
        Priority: bootstrap.PriorityHTTP,
        Config:   cfg.Server,
        CORS:     cfg.CORS,
        JWT:      cfg.JWT,
        RouteSetup: func(engine *gin.Engine) error {
            // Register all routes here
            v1 := engine.Group("/api/v1/auth")
            // ...
            return nil
        },
    }

    return &HTTPServerInitializer{
        HTTPServerInitializer: pkg.initializers.NewHTTPServerInitializer(config, logger),
    }
}
```

### Phase 2: Simple Mode Services (gateway, monitor, collect-agent)

For simple mode services:
1. Create lightweight wrapper using `common/server` directly
2. Use `commonserver.Serve()` for lifecycle management
3. Standardize server creation pattern

#### Example: Gateway Service Refactoring

**After:**
```go
func (s *GatewayService) Run(ctx context.Context) error {
    // Create Gin server using common/server
    ginConfig := httpserver.NewGinServerConfig(s.opts.Server).
        WithCORS(s.opts.CORS)

    ginServer := httpserver.NewGinServerFromFullConfig(s.logger, ginConfig)

    // Register routes
    engine := ginServer.GetEngine()
    s.setupRoutes(engine)

    // Use common server lifecycle management
    return commonserver.Serve(ctx, ginServer, s.logger)
}
```

## Benefits

1. **Consistency**: All services use the same server framework
2. **Reduced Code Duplication**: ~200 lines of server management code per service removed
3. **Better Error Handling**: Centralized error handling in framework
4. **Graceful Shutdown**: Standardized graceful shutdown across all services
5. **Middleware Management**: Centralized middleware configuration
6. **Testing**: Easier to test with standardized interfaces

## Migration Priority

1. **High Priority** (Complex servers with multiple endpoints):
   - auth (extensive API)
   - reasoning (gRPC + HTTP)
   - cluster (full CRUD API)

2. **Medium Priority** (Simpler servers):
   - gateway
   - monitor
   - collect-agent

## Implementation Steps

1. Create this refactoring plan document ✅
2. Refactor auth service as pilot
3. Refactor reasoning service (special case with gRPC+HTTP)
4. Refactor cluster service
5. Refactor simple mode services
6. Update documentation
7. Remove old server management code

## Success Criteria

- All services use `common/server` framework
- No direct `http.Server` creation outside of framework
- Server lifecycle managed by bootstrap or `commonserver.Serve()`
- All services have standardized middleware configuration
- Reduced code duplication (~1000 lines removed)