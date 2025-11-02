# OneX Architecture Pattern - Quick Reference Guide

## Overview

This guide provides a step-by-step process for refactoring services to use the OneX architecture pattern where gRPC and HTTP share the same handler methods.

## Prerequisites

- Go 1.25+ installed
- Buf CLI installed
- `protoc-gen-go-http` installed (`go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest`)
- Kratos v2 dependencies added to project

## Step-by-Step Process

### Step 1: Verify Proto Files Have HTTP Annotations

Check your `.proto` files in `pkg/api/<service>/v1/`:

```protobuf
syntax = "proto3";

package <service>.v1;

import "google/api/annotations.proto";  // ✅ Required

service MyService {
  rpc MyMethod(MyRequest) returns (MyResponse) {
    option (google.api.http) = {      // ✅ Required
      post: "/v1/my-endpoint",
      body: "*",
    };
  }
}
```

**If missing**: Add HTTP annotations to each RPC method.

### Step 2: Update buf.gen.yaml

Ensure `buf.gen.yaml` includes the Kratos HTTP plugin:

```yaml
plugins:
  # ... other plugins ...

  # Add this plugin
  - local: protoc-gen-go-http
    out: pkg/api
    opt:
      - paths=source_relative
```

### Step 3: Generate Proto Code

```bash
buf generate
```

This will generate `*_http.pb.go` files with interfaces like:
- `<Service>HTTPServer` interface (for HTTP)
- `<Service>Server` interface (for gRPC) - already exists

### Step 4: Create Unified Handler

Create `internal/<service>/handler/<service>_handler.go`:

```go
package handler

import (
    "context"

    pb "github.com/kart-io/k8s-agent/pkg/api/<service>/v1"
    "github.com/kart-io/logger/core"
)

// Handler implements both gRPC and HTTP interfaces
type Handler struct {
    pb.Unimplemented<Service>Server  // For gRPC

    // Dependencies
    logger core.Logger
    // ... other dependencies
}

// NewHandler creates a new unified handler
func NewHandler(logger core.Logger) *Handler {
    return &Handler{
        logger: logger.With("component", "<service>-handler"),
    }
}

// MyMethod implements both:
// - pb.<Service>Server.MyMethod (gRPC)
// - pb.<Service>HTTPServer.MyMethod (HTTP)
func (h *Handler) MyMethod(
    ctx context.Context,
    req *pb.MyRequest,
) (*pb.MyResponse, error) {
    // Single implementation for both gRPC and HTTP
    return &pb.MyResponse{
        // ...
    }, nil
}
```

**Key Points**:
- Struct name: `Handler` (simple and clear)
- Embeds `Unimplemented<Service>Server` for forward compatibility
- Methods use `context.Context` and protobuf types
- Same method serves both gRPC and HTTP

### Step 5: Create Unified Server

Create `internal/<service>/server/server.go`:

```go
package server

import (
    "context"
    "fmt"
    "time"

    kratosgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
    kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
    "google.golang.org/grpc"

    "github.com/kart-io/k8s-agent/internal/<service>/handler"
    pb "github.com/kart-io/k8s-agent/pkg/api/<service>/v1"
    "github.com/kart-io/logger/core"
)

type ServerOptions struct {
    HTTPHost string
    HTTPPort int
    GRPCHost string
    GRPCPort int
    Handler  *handler.Handler
}

type Server struct {
    httpServer *kratoshttp.Server
    grpcServer *kratosgrpc.Server
    logger     core.Logger
}

func NewServer(opts *ServerOptions, logger core.Logger) (*Server, error) {
    // Create gRPC server
    grpcAddr := fmt.Sprintf("%s:%d", opts.GRPCHost, opts.GRPCPort)
    grpcServer := kratosgrpc.NewServer(
        kratosgrpc.Address(grpcAddr),
        kratosgrpc.Timeout(30 * time.Second),
    )
    pb.Register<Service>Server(grpcServer, opts.Handler)

    // Create HTTP server
    httpAddr := fmt.Sprintf("%s:%d", opts.HTTPHost, opts.HTTPPort)
    httpServer := kratoshttp.NewServer(
        kratoshttp.Address(httpAddr),
        kratoshttp.Timeout(30 * time.Second),
    )
    pb.Register<Service>HTTPServer(httpServer, opts.Handler)

    return &Server{
        httpServer: httpServer,
        grpcServer: grpcServer,
        logger:     logger,
    }, nil
}

func (s *Server) Start(ctx context.Context) error {
    errCh := make(chan error, 2)

    go func() {
        if err := s.grpcServer.Start(ctx); err != nil {
            errCh <- err
        }
    }()

    go func() {
        if err := s.httpServer.Start(ctx); err != nil {
            errCh <- err
        }
    }()

    select {
    case <-ctx.Done():
        return s.Stop()
    case err := <-errCh:
        return err
    }
}

func (s *Server) Stop() error {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    s.httpServer.Stop(ctx)
    s.grpcServer.Stop(ctx)

    return nil
}
```

### Step 6: Create Unified Initializer

Create `internal/<service>/initializers/unified_server.go`:

```go
package initializers

import (
    "context"
    "fmt"

    "github.com/kart-io/k8s-agent/cmd/<service>/app/options"
    "github.com/kart-io/k8s-agent/internal/<service>/handler"
    "github.com/kart-io/k8s-agent/internal/<service>/server"
    "github.com/kart-io/logger/core"
)

type UnifiedServerInitializer struct {
    opts    *options.ServerOptions
    logger  core.Logger
    server  *server.Server
    handler *handler.Handler
}

func NewUnifiedServerInitializer(
    opts *options.ServerOptions,
    logger core.Logger,
) *UnifiedServerInitializer {
    return &UnifiedServerInitializer{
        opts:   opts,
        logger: logger,
    }
}

func (i *UnifiedServerInitializer) Name() string {
    return "UnifiedServer"
}

func (i *UnifiedServerInitializer) Priority() int {
    return 450  // Adjust based on dependencies
}

func (i *UnifiedServerInitializer) Initialize(ctx context.Context) error {
    // Create handler
    i.handler = handler.NewHandler(i.logger)

    // Create server
    serverOpts := &server.ServerOptions{
        HTTPHost: i.opts.Server.Host,
        HTTPPort: i.opts.Server.Port,
        GRPCHost: i.opts.GRPC.Host,
        GRPCPort: i.opts.GRPC.Port,
        Handler:  i.handler,
    }

    srv, err := server.NewServer(serverOpts, i.logger)
    if err != nil {
        return fmt.Errorf("failed to create server: %w", err)
    }

    i.server = srv

    // Start in background
    go func() {
        if err := srv.Start(ctx); err != nil {
            i.logger.Errorw("Server error", "error", err)
        }
    }()

    return nil
}

func (i *UnifiedServerInitializer) Shutdown(ctx context.Context) error {
    if i.server == nil {
        return nil
    }
    return i.server.Stop()
}
```

### Step 7: Update Application Entry Point

Update `cmd/<service>/app/app.go`:

```go
type ServiceApp struct {
    *commonapp.StandardBootstrapApplication

    // Replace separate grpcInit and httpInit with:
    unifiedServerInit *initializers.UnifiedServerInitializer
}

func (a *ServiceApp) RegisterComponents(bs *bootstrap.Bootstrap) error {
    // Register unified server
    a.unifiedServerInit = initializers.NewUnifiedServerInitializer(
        a.GetOptions().(*options.ServerOptions),
        a.GetLogger(),
    )
    bs.Register(a.unifiedServerInit)

    return nil
}
```

### Step 8: Build and Test

```bash
# Build
go build -o /tmp/<service>-test ./cmd/<service>

# Run
/tmp/<service>-test

# Test gRPC
grpcurl -plaintext localhost:<grpc-port> <package>.<Service>/<Method>

# Test HTTP
curl -X POST http://localhost:<http-port>/v1/<endpoint> \
  -H "Content-Type: application/json" \
  -d '{"field": "value"}'
```

## Common Patterns

### Handler with Dependencies

```go
type Handler struct {
    pb.Unimplemented<Service>Server

    db       *gorm.DB
    cache    cache.Cache
    analyzer *analyzer.Analyzer
    logger   core.Logger
}

func NewHandler(
    db *gorm.DB,
    cache cache.Cache,
    analyzer *analyzer.Analyzer,
    logger core.Logger,
) *Handler {
    return &Handler{
        db:       db,
        cache:    cache,
        analyzer: analyzer,
        logger:   logger,
    }
}
```

### Error Handling

```go
func (h *Handler) MyMethod(ctx context.Context, req *pb.Request) (*pb.Response, error) {
    if err := validate(req); err != nil {
        return nil, status.Errorf(codes.InvalidArgument, "validation failed: %v", err)
    }

    result, err := h.processRequest(ctx, req)
    if err != nil {
        h.logger.Errorw("Processing failed", "error", err)
        return nil, status.Errorf(codes.Internal, "processing failed: %v", err)
    }

    return &pb.Response{Result: result}, nil
}
```

### Type Conversion

```go
// Internal type to Proto
func convertToProto(internal *types.MyType) *pb.MyType {
    return &pb.MyType{
        Field1: internal.Field1,
        Field2: internal.Field2,
        // ...
    }
}

// Proto to Internal type
func convertFromProto(proto *pb.MyType) *types.MyType {
    return &types.MyType{
        Field1: proto.Field1,
        Field2: proto.Field2,
        // ...
    }
}
```

## Checklist

- [ ] Proto files have `google.api.http` annotations
- [ ] `buf.gen.yaml` includes `protoc-gen-go-http` plugin
- [ ] Run `buf generate` successfully
- [ ] `*_http.pb.go` files generated
- [ ] Created `handler/<service>_handler.go`
- [ ] Handler implements both gRPC and HTTP interfaces
- [ ] Created `server/server.go` with Kratos servers
- [ ] Created `initializers/unified_server.go`
- [ ] Updated `cmd/<service>/app/app.go`
- [ ] Build succeeds with no errors
- [ ] Both gRPC and HTTP endpoints tested
- [ ] Documentation updated

## Troubleshooting

### Issue: `*_http.pb.go` not generated
**Solution**: Ensure `protoc-gen-go-http` is in PATH and `buf.gen.yaml` has the plugin configured.

### Issue: Compilation error - method signature mismatch
**Solution**: Check that both interfaces have identical signatures. Use `context.Context` and proto types.

### Issue: Server fails to start
**Solution**: Check port conflicts, ensure proper initialization order, verify dependencies are initialized.

### Issue: HTTP endpoints return 404
**Solution**: Verify proto annotations match the routes, check server registration.

## Best Practices

1. **Single Handler**: Always use one handler for both protocols
2. **Proto Types**: Use protobuf types in method signatures
3. **Error Handling**: Use `google.golang.org/grpc/status` for errors
4. **Logging**: Add structured logging with context
5. **Validation**: Validate input at the handler level
6. **Testing**: Test both gRPC and HTTP endpoints
7. **Documentation**: Keep proto comments up to date

## Resources

- [Kratos Documentation](https://go-kratos.dev/)
- [gRPC-Gateway](https://github.com/grpc-ecosystem/grpc-gateway)
- [Google API HTTP Annotations](https://github.com/googleapis/googleapis/blob/master/google/api/http.proto)
- [Reasoning Service Example](./ONEX_REASONING_REFACTORING.md)

## Next Services to Refactor

1. **Orchestrator** (Priority 2)
   - Proto: `pkg/api/orchestrator/v1/workflow.proto`
   - Current: Separate gRPC and HTTP
   - Estimated effort: 4-6 hours

2. **Agent Manager** (Priority 3)
   - Proto: `pkg/api/agent/v1/agent.proto`
   - Current: Separate gRPC and HTTP
   - Estimated effort: 4-6 hours

## Summary

The OneX architecture pattern provides:
- **Consistency**: Same logic for both protocols
- **Maintainability**: Single handler to update
- **Type Safety**: Protobuf types throughout
- **Performance**: No conversion overhead
- **Simplicity**: Less code, clearer structure

Follow this guide to refactor any service to the OneX pattern.
