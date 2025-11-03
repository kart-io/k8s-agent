# Server Integration Refactoring

## 概述

此重构将 `common/server` 的标准 HTTP/gRPC 服务器实现集成到 `common/bootstrap` 系统中，实现了服务器的统一管理和自动启动。同时修复了 `ApplicationRunner` 和 `Bootstrap` 之间的信号处理冲突问题。

## 重构日期

2025-11-03

## 问题描述

### 重构前的问题

1. **HTTP/gRPC 服务器实现分散**
   - 每个服务（如 `orchestrator`）都自己实现 HTTP/gRPC 服务器
   - 服务器启动和关闭逻辑重复
   - 没有利用 `common/server` 中已有的标准实现

2. **Bootstrap 系统未集成服务器**
   - `common/bootstrap` 只负责组件初始化
   - 服务器需要在初始化器中手动启动（`go srv.Start()`）
   - 服务器关闭逻辑也需要手动实现

3. **代码重复**
   - 每个服务都要实现服务器配置、中间件、拦截器
   - 服务器生命周期管理代码重复

4. **信号处理冲突** ⚠️
   - `ApplicationRunner.Run()` 监听信号
   - `bootstrap.Run()` 也监听信号
   - 导致 `bootstrap.Shutdown()` 可能被调用两次

## 解决方案

### 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                      common/app                             │
│  RunWithRunner() -> Bootstrap.Run()                         │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   common/bootstrap                          │
│  • Bootstrap.Run()                                          │
│  • 自动收集 ServerProvider                                   │
│  • 自动启动所有服务器                                        │
│  • 自动关闭所有服务器                                        │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│               common/initializers                           │
│  • HTTPServerInitializer                                    │
│  • GRPCServerInitializer                                    │
│  (包装 common/server 的标准实现)                            │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   common/server                             │
│  • Server 接口                                              │
│  • GinServer (HTTP)                                         │
│  • StandardGRPCServer (gRPC)                                │
│  • Serve() / MultiServe()                                   │
└─────────────────────────────────────────────────────────────┘
```

### 核心改动

#### 1. 添加 `ServerProvider` 接口 (`common/bootstrap/bootstrap.go`)

```go
// ServerProvider represents a component that provides a server instance.
type ServerProvider interface {
    GetServer() Server
}

// Server is the interface from common/server package.
type Server interface {
    RunOrDie()
    GracefulStop(ctx context.Context)
}
```

#### 2. 修改 `Bootstrap.Run()` 方法

```go
func (b *Bootstrap) Run(ctx context.Context, runFunc func() error) error {
    // 1. Initialize all components
    if err := b.Initialize(ctx); err != nil {
        return err
    }

    // 2. Collect all servers from ServerProvider initializers
    var servers []Server
    for _, init := range b.initializers {
        if provider, ok := init.(ServerProvider); ok {
            if srv := provider.GetServer(); srv != nil {
                servers = append(servers, srv)
            }
        }
    }

    // 3. Start all servers in background
    for _, srv := range servers {
        go srv.RunOrDie()
    }

    // 4. Wait for signals...
    // 5. Stop servers gracefully before shutdown
}
```

#### 3. 创建标准服务器初始化器

**HTTP 服务器初始化器** (`common/initializers/http_server.go`):

```go
type HTTPServerInitializer struct {
    name         string
    priority     int
    logger       core.Logger
    serverConfig *options.ServerOptions
    routeSetup   func(*gin.Engine) error
    server       commonserver.Server
}

// 实现 ServerProvider 接口
func (i *HTTPServerInitializer) GetServer() bootstrap.Server {
    return i.server
}
```

**gRPC 服务器初始化器** (`common/initializers/grpc_server.go`):

```go
type GRPCServerInitializer struct {
    name            string
    priority        int
    logger          core.Logger
    serverConfig    *options.GRPCOptions
    serviceRegister func(*grpc.Server) error
    server          commonserver.Server
}

// 实现 ServerProvider 接口
func (i *GRPCServerInitializer) GetServer() bootstrap.Server {
    return i.server
}
```

#### 4. 修改服务的初始化器

**Orchestrator gRPC 初始化器** (`internal/orchestrator/initializers/grpc.go`):

```go
type GRPCServerInitializer struct {
    // ...
    standardInit    *commoninitializers.GRPCServerInitializer
    workflowService *service.WorkflowServiceServer
}

func (i *GRPCServerInitializer) Initialize(ctx context.Context) error {
    // 创建业务服务
    i.workflowService = service.NewWorkflowServiceServer(...)

    // 使用标准初始化器
    i.standardInit = commoninitializers.NewGRPCServerInitializer(
        &commoninitializers.GRPCServerConfig{
            Config: i.opts.GRPC,
            ServiceRegister: func(grpcServer *grpc.Server) error {
                orchestratorv1.RegisterWorkflowServiceServer(grpcServer, i.workflowService)
                return nil
            },
        },
        i.logger,
    )

    return i.standardInit.Initialize(ctx)
}

// 实现 ServerProvider 接口
func (i *GRPCServerInitializer) GetServer() bootstrap.Server {
    if i.standardInit == nil {
        return nil
    }
    return i.standardInit.GetServer()
}
```

**Orchestrator HTTP 初始化器** (`internal/orchestrator/initializers/http.go`):

```go
type HTTPServerInitializer struct {
    // ...
    standardInit *commoninitializers.HTTPServerInitializer
    mux          *runtime.ServeMux
}

func (i *HTTPServerInitializer) Initialize(ctx context.Context) error {
    // 创建 gRPC-Gateway mux
    i.mux = runtime.NewServeMux(...)

    // 注册 gRPC service handler
    workflowService := i.grpcInit.GetWorkflowService()
    orchestratorv1.RegisterWorkflowServiceHandlerServer(ctx, i.mux, workflowService)

    // 使用标准初始化器
    i.standardInit = commoninitializers.NewHTTPServerInitializer(
        &commoninitializers.HTTPServerConfig{
            Config: i.opts.Server,
            RouteSetup: func(engine *gin.Engine) error {
                // 将 gRPC-Gateway mux 挂载到 Gin
                engine.Any("/api/*path", gin.WrapH(i.mux))
                return nil
            },
        },
        i.logger,
    )

    return i.standardInit.Initialize(ctx)
}

// 实现 ServerProvider 接口
func (i *HTTPServerInitializer) GetServer() bootstrap.Server {
    if i.standardInit == nil {
        return nil
    }
    return i.standardInit.GetServer()
}
```

#### 5. 修复 ApplicationRunner 信号处理冲突

**问题**：ApplicationRunner 和 bootstrap.Run() 都在监听信号，导致重复关闭。

**修复前** (`common/app/runner.go`):

```go
func (r *ApplicationRunner) Run() error {
    // ...

    // 设置信号处理
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

    // 在 goroutine 中运行应用程序
    errCh := make(chan error, 1)
    go func() {
        if err := r.app.Run(ctx); err != nil {
            errCh <- err
        }
    }()

    // 等待信号或错误
    select {
    case sig := <-sigCh:  // ❌ 问题：这里监听信号
        r.logger.Infow("Received signal, shutting down", "signal", sig.String())
        cancel()
        return r.app.Shutdown(context.Background())
    case err := <-errCh:
        return err
    }
}
```

**修复后**:

```go
func (r *ApplicationRunner) Run() error {
    // 1. 初始化日志
    if r.loggerInit != nil {
        logger, err := r.loggerInit(r.opts)
        if err != nil {
            return fmt.Errorf("failed to initialize logger: %w", err)
        }
        r.logger = logger
    }

    // 2. 创建上下文
    ctx := context.Background()

    // 3. 初始化应用程序
    if err := r.app.Initialize(ctx, r.opts); err != nil {
        return fmt.Errorf("failed to initialize application: %w", err)
    }

    // 4. 运行应用程序
    // ✅ 修复：app.Run() 会阻塞直到收到信号或发生错误
    // 信号处理由 bootstrap.Run() 负责，不需要在这里处理
    if err := r.app.Run(ctx); err != nil {
        return fmt.Errorf("application run failed: %w", err)
    }

    // 5. 正常退出（由信号触发）
    if r.logger != nil {
        r.logger.Infow("Application exited normally")
    }

    return nil
}
```

**信号处理流程**：

```
RunWithRunner()
  → ApplicationRunner.Run()
    → app.Initialize()
    → app.Run() [阻塞] ──────┐
                              │
                              ├─ bootstrap.Run()
                              │    ├─ 启动服务器
                              │    ├─ 等待信号 ◄─── ✅ 唯一的信号处理点
                              │    └─ 关闭服务器
                              │
                              └─ 返回到 ApplicationRunner
  ← 正常退出
```

## 优势

### 1. 统一服务器管理

- **自动启动**: Bootstrap 自动收集和启动所有服务器
- **自动关闭**: 统一的优雅关闭流程
- **一致性**: 所有服务器使用相同的生命周期管理

### 2. 代码重用

- **标准实现**: 所有服务器使用 `common/server` 的标准实现
- **减少重复**: 不再需要每个服务重复实现服务器启动/关闭逻辑
- **配置统一**: 使用 `options.ServerOptions` / `options.GRPCOptions`

### 3. 灵活性

- **可组合**: 支持 HTTP-only、gRPC-only、HTTP+gRPC 等多种组合
- **易扩展**: 新服务只需实现 `ServerProvider` 接口
- **保持兼容**: gRPC-Gateway 等特殊需求仍可支持

### 4. 更好的错误处理

- **统一日志**: 所有服务器使用统一的日志记录
- **优雅关闭**: 自动处理超时和强制关闭
- **健康检查**: 与 Bootstrap 的健康检查系统集成

## 影响范围

### 已修改的文件

1. **common/bootstrap/bootstrap.go**
   - 添加 `ServerProvider` 和 `Server` 接口
   - 修改 `Run()` 方法以支持自动服务器管理

2. **common/initializers/http_server.go** (新建)
   - 标准 HTTP 服务器初始化器

3. **common/initializers/grpc_server.go** (新建)
   - 标准 gRPC 服务器初始化器

4. **internal/orchestrator/initializers/grpc.go**
   - 使用标准 gRPC 初始化器
   - 实现 `ServerProvider` 接口

5. **internal/orchestrator/initializers/http.go**
   - 使用标准 HTTP 初始化器
   - 实现 `ServerProvider` 接口
   - 支持 gRPC-Gateway

### 需要修改的其他服务

以下服务也应该按照相同的模式进行重构：

- `agent-manager` (HTTP + NATS)
- `reasoning` (HTTP + AI)
- `auth` (HTTP)
- `cluster` (HTTP)
- `collect-agent` (NATS client, 不需要 HTTP/gRPC 服务器)

## 迁移指南

### 为现有服务添加 ServerProvider 支持

1. **修改服务的初始化器**

```go
// 内部使用标准初始化器
type MyGRPCInitializer struct {
    standardInit *commoninitializers.GRPCServerInitializer
    // ... 其他字段
}

func (i *MyGRPCInitializer) Initialize(ctx context.Context) error {
    // 创建标准初始化器
    i.standardInit = commoninitializers.NewGRPCServerInitializer(
        &commoninitializers.GRPCServerConfig{
            Name:     "MyServiceGRPC",
            Priority: i.Priority(),
            Config:   i.opts.GRPC,
            ServiceRegister: func(grpcServer *grpc.Server) error {
                // 注册你的 gRPC 服务
                mypb.RegisterMyServiceServer(grpcServer, myService)
                return nil
            },
        },
        i.logger,
    )

    return i.standardInit.Initialize(ctx)
}

// 实现 ServerProvider 接口
func (i *MyGRPCInitializer) GetServer() bootstrap.Server {
    if i.standardInit == nil {
        return nil
    }
    return i.standardInit.GetServer()
}
```

2. **删除手动启动服务器的代码**

```go
// 删除这些代码：
// go func() {
//     if err := server.Start(ctx); err != nil {
//         // ...
//     }
// }()
```

3. **删除 `Shutdown()` 方法中的服务器关闭代码**

```go
// 删除或简化 Close/Shutdown 方法
func (i *MyGRPCInitializer) Close(ctx context.Context) error {
    // Bootstrap 会自动调用 server.GracefulStop()
    // 这里只需要清理其他资源
    return nil
}
```

## 测试

### 编译测试

```bash
# 构建 orchestrator 服务
make go.build.orchestrator

# 或直接使用 go build
go build -o /tmp/orchestrator ./cmd/orchestrator
```

### 运行测试

```bash
# 运行 orchestrator 服务
./orchestrator --config configs/orchestrator.yaml

# 检查日志输出，应该看到：
# - "Registered server from initializer" (自动注册)
# - "Starting servers" count=2 (HTTP + gRPC)
# - "Starting Gin HTTP server"
# - "Starting standard gRPC server"
```

### 验证服务器

```bash
# 测试 HTTP 端点
curl http://localhost:8081/health

# 测试 gRPC-Gateway HTTP API
curl http://localhost:8081/api/v1/workflows

# 测试 gRPC (使用 grpcurl)
grpcurl -plaintext localhost:9091 list
```

## 未来改进

1. **添加服务器生命周期钩子**
   - `PreStart()` / `PostStart()`
   - `PreStop()` / `PostStop()`

2. **支持服务器依赖关系**
   - 确保服务器按正确顺序启动和关闭

3. **改进错误处理**
   - 如果某个服务器启动失败，是否应该停止其他服务器？

4. **添加服务器状态监控**
   - 服务器健康检查
   - 自动重启失败的服务器

5. **支持动态服务器**
   - 运行时添加/移除服务器

## 总结

此重构成功地将 `common/server` 的标准实现集成到 `common/bootstrap` 系统中，实现了：

✅ 服务器的统一管理和自动启动
✅ 减少了大量重复代码
✅ 保持了灵活性和可扩展性
✅ Orchestrator 服务已成功迁移并测试通过

这为后续其他服务的迁移提供了清晰的模式和参考。
