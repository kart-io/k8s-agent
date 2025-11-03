# Bootstrap + Server 集成指南

本文档说明 `common/app`、`common/bootstrap` 和 `common/server` 三者如何协作，以及如何在服务中正确使用它们。

## 架构概览

```
┌─────────────────────────────────────────────────────────────┐
│                    Service (例如: agent-manager)             │
│                                                               │
│  cmd/agent-manager/app/app.go                                │
│    ├─> AgentManagerApp (implements Application)             │
│    └─> commonapp.RunWithRunner()                             │
└─────────────────┬───────────────────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────────────────┐
│              common/app (应用启动框架)                        │
│                                                               │
│  runner.go                                                    │
│    └─> ApplicationRunner.Run()                               │
│          ├─> app.Initialize(ctx, opts)                       │
│          └─> app.Run(ctx)  ──────────────┐                  │
└───────────────────────────────────────────┼──────────────────┘
                                            │
                                            ▼
┌─────────────────────────────────────────────────────────────┐
│           common/bootstrap (生命周期管理)                     │
│                                                               │
│  bootstrap.go                                                 │
│    └─> Bootstrap.Run(ctx, runFunc)                           │
│          ├─> Initialize() - 初始化所有 initializers          │
│          ├─> 收集 ServerProvider 的服务器 ──┐                │
│          ├─> 启动所有服务器 (RunOrDie)      │                │
│          ├─> 等待信号 (SIGTERM/SIGINT)       │                │
│          └─> 优雅关停 (GracefulStop)         │                │
└──────────────────────────────────────────┼──────────────────┘
                                           │
                                           ▼
┌─────────────────────────────────────────────────────────────┐
│            common/server (服务器实现)                         │
│                                                               │
│  http/gin.go                                                  │
│    └─> GinServer (implements Server interface)               │
│          ├─> RunOrDie() - 启动 HTTP 服务器                   │
│          └─> GracefulStop() - 优雅关停                       │
│                                                               │
│  grpc/grpc.go                                                 │
│    └─> StandardGRPCServer (implements Server interface)      │
│          ├─> RunOrDie() - 启动 gRPC 服务器                   │
│          └─> GracefulStop() - 优雅关停                       │
└─────────────────────────────────────────────────────────────┘
                  ▲
                  │
┌─────────────────┴───────────────────────────────────────────┐
│         common/initializers (组件初始化器)                    │
│                                                               │
│  http_server.go                                               │
│    └─> HTTPServerInitializer                                 │
│          ├─> implements Initializer                          │
│          ├─> implements ServerProvider                       │
│          └─> GetServer() → GinServer                         │
│                                                               │
│  grpc_server.go                                               │
│    └─> GRPCServerInitializer                                 │
│          ├─> implements Initializer                          │
│          ├─> implements ServerProvider                       │
│          └─> GetServer() → StandardGRPCServer                │
└─────────────────────────────────────────────────────────────┘
```

## 核心接口

### 1. ServerProvider 接口

定义在 `common/bootstrap/bootstrap.go`:

```go
// ServerProvider represents a component that provides a server instance.
type ServerProvider interface {
    // GetServer returns the server instance (from common/server package).
    GetServer() Server
}
```

**作用**: 允许 initializers 向 bootstrap 注册服务器实例。

### 2. Server 接口

定义在 `common/bootstrap/bootstrap.go` 和 `common/server/server.go`:

```go
type Server interface {
    // RunOrDie runs the server, exits on failure
    RunOrDie()

    // GracefulStop gracefully stops the server
    GracefulStop(ctx context.Context)
}
```

**实现类**:
- `common/server/http.GinServer`
- `common/server/grpc.StandardGRPCServer`

## 集成流程详解

### 启动流程 (从服务到服务器)

1. **服务入口** (`cmd/{service}/app/app.go`)
   ```go
   func Execute() {
       opts := options.NewServerOptions()
       app := &MyServiceApp{}
       commonapp.RunWithRunner(opts, app, initLogger, config)
   }
   ```

2. **App 初始化** (`AgentManagerApp.Initialize()`)
   ```go
   func (a *AgentManagerApp) Initialize(ctx context.Context, opts Options) error {
       // 注册所有初始化器
       return a.RegisterComponents(bootstrap)
   }
   ```

3. **注册初始化器** (`AgentManagerApp.RegisterComponents()`)
   ```go
   func (a *AgentManagerApp) RegisterComponents(bs *bootstrap.Bootstrap) error {
       // 创建 HTTP 服务器初始化器 - 唯一正确的方式
       httpConfig := &initializers.HTTPServerConfig{
           Name:       "http-server",
           Priority:   bootstrap.PriorityHTTP,
           Config:     opts.Server,
           RouteSetup: a.setupRoutes,
           CORS:       opts.CORS,
           JWT:        opts.JWT,
           RateLimit:  opts.RateLimit,
       }
       a.httpInit = initializers.NewHTTPServerInitializer(httpConfig, logger)
       bs.Register(a.httpInit) // HTTPServerInitializer 实现 ServerProvider

       // 创建 gRPC 服务器初始化器 (可选)
       if opts.GRPC.Enable {
           grpcConfig := &initializers.GRPCServerConfig{
               Name:            "grpc-server",
               Priority:        bootstrap.PriorityGRPC,
               Config:          opts.GRPC,
               ServiceRegister: a.registerGRPCServices,
           }
           a.grpcInit = initializers.NewGRPCServerInitializer(grpcConfig, logger)
           bs.Register(a.grpcInit) // GRPCServerInitializer 实现 ServerProvider
       }

       return nil
   }
   ```

4. **Bootstrap 运行** (`bootstrap.Run()`)
   ```go
   func (b *Bootstrap) Run(ctx context.Context, runFunc func() error) error {
       // 4.1 初始化所有组件
       b.Initialize(ctx)

       // 4.2 收集所有服务器 (从 ServerProvider)
       var servers []Server
       for _, init := range b.initializers {
           if provider, ok := init.(ServerProvider); ok {
               if srv := provider.GetServer(); srv != nil {
                   servers = append(servers, srv)
               }
           }
       }

       // 4.3 启动所有服务器
       for _, srv := range servers {
           go srv.RunOrDie()  // 启动 GinServer 或 StandardGRPCServer
       }

       // 4.4 等待信号
       <-sigChan

       // 4.5 优雅关停
       for _, srv := range servers {
           srv.GracefulStop(ctx)  // 优雅关停所有服务器
       }

       return b.Shutdown(ctx)
   }
   ```

5. **HTTP 服务器启动** (`GinServer.RunOrDie()`)
   ```go
   func (s *GinServer) RunOrDie() {
       if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
           s.logger.Fatalw("Server failed", "err", err)
       }
   }
   ```

### 关停流程 (从信号到服务器)

1. 接收 SIGTERM/SIGINT 信号
2. Bootstrap 调用所有 `Server.GracefulStop(ctx)`
3. HTTP/gRPC 服务器停止接收新请求
4. 等待现有请求处理完成 (最多 30秒)
5. 关闭数据库、Redis 等资源 (`Closer.Close()`)

## Middleware 与 Options 集成

### 配置流程

```
┌─────────────────────────────────────────────────────────────┐
│                  1. 服务配置文件 (YAML/ENV)                  │
│                                                               │
│  server:                                                      │
│    host: "0.0.0.0"                                            │
│    port: 8080                                                 │
│  cors:                                                        │
│    enabled: true                                              │
│    allow_origins: ["*"]                                       │
│  jwt:                                                         │
│    secret: "my-secret-key"                                    │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│              2. Options 结构体 (common/options/)              │
│                                                               │
│  options.ServerOptions                                        │
│  options.CORSOptions                                          │
│  options.JWTOptions                                           │
│  options.RateLimitOptions                                     │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│       3. HTTPServerInitializer Config (common/initializers/)  │
│                                                               │
│  HTTPServerConfig {                                           │
│      Config:     *options.ServerOptions                       │
│      CORS:       *options.CORSOptions                         │
│      JWT:        *options.JWTOptions                          │
│      RateLimit:  *options.RateLimitOptions                    │
│  }                                                            │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│          4. GinServerConfig (common/server/http/config.go)    │
│                                                               │
│  ginConfig := NewGinServerConfig(serverOpts)                  │
│      .WithCORS(corsOpts)                                      │
│      .WithJWT(jwtOpts)                                        │
│      .WithRateLimit(rateLimitOpts)                            │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│   5. 配置转换 (common/server/http/converter.go)              │
│                                                               │
│  ToCORSConfig(opts.CORSOptions) → middleware.CORSConfig       │
│  ToJWTConfig(opts.JWTOptions)   → middleware.JWTConfig        │
└─────────────────────┬───────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────────┐
│         6. 应用中间件 (common/server/http/gin.go)            │
│                                                               │
│  if config.CORS != nil && config.CORS.Enabled {               │
│      corsConfig := ToCORSConfig(config.CORS)                  │
│      engine.Use(middleware.CORSWithConfig(corsConfig))        │
│  }                                                            │
│                                                               │
│  if config.JWT != nil {                                       │
│      jwtConfig := ToJWTConfig(config.JWT)                     │
│      // JWT middleware 需要在��定路由上应用                   │
│  }                                                            │
└─────────────────────────────────────────────────────────────┘
```

### 关键组件

#### 1. GinServerConfig (配置聚合)

`common/server/http/config.go`:

```go
type GinServerConfig struct {
    Server *options.ServerOptions

    // 中间件配置
    CORS      *options.CORSOptions
    JWT       *options.JWTOptions
    RateLimit *options.RateLimitOptions

    // 中间件启用标志
    EnableRecovery  bool
    EnableRequestID bool
    EnableLogger    bool
}

// Fluent API
func (c *GinServerConfig) WithCORS(cors *options.CORSOptions) *GinServerConfig
func (c *GinServerConfig) WithJWT(jwt *options.JWTOptions) *GinServerConfig
func (c *GinServerConfig) WithRateLimit(rl *options.RateLimitOptions) *GinServerConfig
```

#### 2. 配置转换函数

`common/server/http/converter.go`:

```go
// ToCORSConfig converts options.CORSOptions to middleware.CORSConfig
func ToCORSConfig(opts *options.CORSOptions) middleware.CORSConfig

// ToJWTConfig converts options.JWTOptions to middleware.JWTConfig
func ToJWTConfig(opts *options.JWTOptions) *middleware.JWTConfig
```

## 使用示例

### 示例 1: 创建带 CORS 和 JWT 的 HTTP 服务器

```go
// cmd/myservice/app/app.go
func (a *MyServiceApp) RegisterComponents(bs *bootstrap.Bootstrap) error {
    opts := a.GetOptions().(*options.ServerOptions)

    // 创建 HTTP 服务器初始化器
    httpConfig := &initializers.HTTPServerConfig{
        Name:     "http-server",
        Priority: bootstrap.PriorityHTTP,
        Config:   opts.Server,
        CORS:     opts.CORS,
        JWT:      opts.JWT,
        RouteSetup: func(engine *gin.Engine) error {
            // 公开路由
            engine.GET("/health", handleHealth)

            // 受保护的路由
            api := engine.Group("/api/v1")
            if opts.JWT != nil {
                ginConfig := httpserver.NewGinServerConfig(opts.Server).
                    WithJWT(opts.JWT)
                jwtMiddleware := httpserver.GetJWTMiddleware(ginConfig)
                api.Use(jwtMiddleware.Handler())
            }
            api.GET("/users", handleListUsers)

            return nil
        },
    }

    httpInit := initializers.NewHTTPServerInitializer(httpConfig, a.GetLogger())
    bs.Register(httpInit)

    return nil
}
```

### 示例 2: 创建 gRPC 服务器

```go
func (a *MyServiceApp) RegisterComponents(bs *bootstrap.Bootstrap) error {
    opts := a.GetOptions().(*options.ServerOptions)

    // 创建 gRPC 服务器初始化器
    if opts.GRPC.Enable {
        grpcConfig := &initializers.GRPCServerConfig{
            Name:     "grpc-server",
            Priority: bootstrap.PriorityGRPC,
            Config:   opts.GRPC,
            ServiceRegister: func(srv *grpc.Server) error {
                mypb.RegisterMyServiceServer(srv, a.grpcHandler)
                return nil
            },
        }

        grpcInit := initializers.NewGRPCServerInitializer(grpcConfig, a.GetLogger())
        bs.Register(grpcInit)
    }

    return nil
}
```

### 示例 3: 完整的服务配置文件

```yaml
# configs/config.yaml

# HTTP 服务器配置
server:
  host: "0.0.0.0"
  port: 8080
  mode: "release"
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s

# CORS 配置
cors:
  enabled: true
  allow_origins:
    - "https://example.com"
    - "https://app.example.com"
  allow_methods:
    - "GET"
    - "POST"
    - "PUT"
    - "DELETE"
  allow_headers:
    - "Content-Type"
    - "Authorization"
  allow_credentials: true
  max_age: 3600

# JWT 配置
jwt:
  secret: "${JWT_SECRET}"  # 从环境变量读取
  expires_hours: 24

# Rate Limiting 配置
rate_limit:
  enable: true
  requests_per_minute: 100

# gRPC 服务器配置 (可选)
grpc:
  enable: true
  port: 9090
  max_recv_msg_size: 10485760  # 10MB
```

## 常见���题

### Q1: 服务器何时启动？

**A**: 在 `bootstrap.Run()` 方法中自动启动。Bootstrap 会：
1. 收集所有实现了 `ServerProvider` 的 initializers
2. 调用 `GetServer()` 获取服务器实例
3. 在后台 goroutine 中启动每个服务器 (`srv.RunOrDie()`)

### Q2: 如何配置中间件？

**A**: 通过 `HTTPServerConfig` 传递中间件配置:

```go
httpConfig := &initializers.HTTPServerConfig{
    CORS:      opts.CORS,      // CORS 配置
    JWT:       opts.JWT,       // JWT 配置
    RateLimit: opts.RateLimit, // Rate Limit 配置
}
httpInit := initializers.NewHTTPServerInitializer(httpConfig, logger)
```

### Q3: JWT 中间件为什么不自动应用？

**A**: JWT 通常只需要在受保护的路由上应用，而不是全局应用。建议在 `RouteSetup` 中手动应用:

```go
RouteSetup: func(engine *gin.Engine) error {
    // 公开路由
    engine.GET("/health", handleHealth)

    // 受保护的路由
    protected := engine.Group("/api/v1")
    if opts.JWT != nil {
        ginConfig := httpserver.NewGinServerConfig(opts.Server).
            WithJWT(opts.JWT)
        jwtMiddleware := httpserver.GetJWTMiddleware(ginConfig)
        protected.Use(jwtMiddleware.Handler())
    }
    protected.GET("/users", handleListUsers)

    return nil
}
```

### Q4: 如何同时运行 HTTP 和 gRPC 服务器？

**A**: 注册两个初始化器即可:

```go
func (a *MyServiceApp) RegisterComponents(bs *bootstrap.Bootstrap) error {
    // 注册 HTTP 服务器
    httpInit := initializers.NewHTTPServerInitializer(httpConfig, logger)
    bs.Register(httpInit)

    // 注册 gRPC 服务器
    if opts.GRPC.Enable {
        grpcInit := initializers.NewGRPCServerInitializer(grpcConfig, logger)
        bs.Register(grpcInit)
    }

    return nil
}
```

Bootstrap 会自动启动和管理两个服务器。

## 最佳实践

1. **使用配置文件**: 将中间件配置放在 YAML 配置文件中，便于不同环境使用不同配置
2. **敏感信息使用环境变量**: JWT secret、数据库密码等使用环境变量
3. **JWT 只应用在需要保护的路由**: 避免全局应用，保持 health check 等公开端点可访问
4. **CORS 生产环境不使用 `*`**: 生产环境应该明确指定允许的 origins
5. **启用 Rate Limiting**: 防止 API 滥用
6. **使用 Recovery 中间件**: 防止 panic 导致服务器崩溃 (默认启用)
7. **启用 RequestID**: 便于追踪请求 (默认启用)

## 总结

✅ **集成已完成**: `common/app` 和 `common/bootstrap` 通过 `ServerProvider` 接口正确使用 `common/server`

✅ **配置已关联**: `common/middleware` 与 `common/options` 通过配置结构和转换函数关联

✅ **架构清晰**: Bootstrap 自动管理服务器生命周期，服务开发者只需注册 initializers

✅ **无兼容性代码**: 删除了所有旧的兼容性函数，只保留一种正确的创建方式

如有问题或建议，请在项目 issue 中提出。
