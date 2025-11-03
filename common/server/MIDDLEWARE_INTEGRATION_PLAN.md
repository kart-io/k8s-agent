# HTTP Server 中间件配置优化方案

## 问题分析

### 当前问题

1. **配置断层**
   - `common/options` 有 `CORSOptions`, `JWTOptions`
   - `common/middleware` 有对应的中间件实现
   - `common/server/http` **没有连接两者**，硬编码中间件

2. **缺少中间件配置**
   - `ServerOptions` 只有 HTTP 基础配置
   - 没有中间件相关配置字段

3. **硬编码中间件**
   ```go
   // common/server/http/gin.go:132-136
   engine.Use(middleware.Recovery())      // 不可配置
   engine.Use(middleware.RequestID())     // 不可配置
   engine.Use(middleware.RequestLogger()) // 不可配置
   engine.Use(middleware.CORS())          // 硬编码，不使用 CORSOptions
   ```

## 优化目标

1. 连接 `options` → `server` → `middleware`
2. 使所有中间件可配置
3. 保持向后兼容
4. 简化服务器启动流程

## 解决方案

### 方案 1：扩展 ServerOptions（推荐）

在 `ServerOptions` 中添加中间件配置字段：

```go
// common/options/server_options.go
type ServerOptions struct {
    Host         string
    Port         int
    Mode         string
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
    IdleTimeout  time.Duration
    GracefulStop time.Duration

    // 中间件配置
    Middleware MiddlewareOptions `mapstructure:"middleware"`
}

type MiddlewareOptions struct {
    EnableRecovery bool         `mapstructure:"enable_recovery"`
    EnableRequestID bool        `mapstructure:"enable_request_id"`
    EnableLogger   bool         `mapstructure:"enable_logger"`
    CORS           *CORSOptions `mapstructure:"cors"`
    JWT            *JWTOptions  `mapstructure:"jwt"`
}
```

**优点**：
- 配置集中管理
- 易于序列化（YAML/JSON）
- 符合现有模式

**缺点**：
- `ServerOptions` 变得更大
- 需要修改现有配置文件

### 方案 2：独立 MiddlewareConfig（采用）

创建独立的中间件配置，通过函数式选项传递：

```go
// common/server/http/middleware_config.go
type MiddlewareConfig struct {
    EnableRecovery  bool
    EnableRequestID bool
    EnableLogger    bool
    CORS            *CORSConfig    // 从 options.CORSOptions 转换
    JWT             *JWTConfig     // 从 options.JWTOptions 转换
}

// common/server/http/gin.go
func NewGinServerFromConfig(log core.Logger, opts *options.ServerOptions, middlewareOpts ...MiddlewareOption) *GinServer
```

**优点**：
- 解耦，不污染 `ServerOptions`
- 灵活，易于扩展
- 向后兼容（中间件配置可选）

**缺点**：
- 配置分散在两个地方

### 方案 3：使用 common/options 中的配置（最佳）

直接在 `GinServer` 初始化时使用 `common/options` 中的配置：

```go
// common/server/http/gin.go
type GinServerConfig struct {
    Server     *options.ServerOptions
    CORS       *options.CORSOptions    // 可选
    JWT        *options.JWTOptions     // 可选
    RateLimit  *options.RateLimitOptions // 可选
}

func NewGinServerFromConfig(log core.Logger, config *GinServerConfig) *GinServer {
    gin.SetMode(config.Server.Mode)
    engine := gin.New()

    // 应用默认中间件
    engine.Use(middleware.Recovery())
    engine.Use(middleware.RequestID())
    engine.Use(middleware.RequestLogger())

    // 应用可选中间件
    if config.CORS != nil && config.CORS.Enabled {
        engine.Use(middleware.CORSWithConfig(toMiddlewareCORSConfig(config.CORS)))
    }

    if config.JWT != nil {
        jwtMiddleware := middleware.NewJWTMiddleware(&middleware.JWTConfig{
            Secret: []byte(config.JWT.Secret),
        })
        // JWT 需要按需在路由上使用
    }

    // ...
}
```

**优点**：
- 直接使用已有的 `options` 配置
- 不引入新的配置结构
- 配置统一，易于管理

**缺点**：
- 需要写转换函数（`CORSOptions` → `middleware.CORSConfig`）

## 实施计划

采用**方案 3**：

### 步骤 1：创建 GinServerConfig

```go
// common/server/http/config.go
type GinServerConfig struct {
    Server    *options.ServerOptions
    CORS      *options.CORSOptions
    JWT       *options.JWTOptions
    RateLimit *options.RateLimitOptions
}
```

### 步骤 2：修改 GinServer 初始化

```go
// common/server/http/gin.go
func NewGinServerFromFullConfig(log core.Logger, config *GinServerConfig) *GinServer
```

### 步骤 3：添加配置转换函数

```go
// common/server/http/converter.go
func toMiddlewareCORSConfig(opts *options.CORSOptions) middleware.CORSConfig
```

### 步骤 4：更新 HTTPServerInitializer

```go
// common/initializers/http_server.go
type HTTPServerConfig struct {
    Name         string
    Priority     int
    Server       *options.ServerOptions
    CORS         *options.CORSOptions    // 新增
    JWT          *options.JWTOptions     // 新增
    RateLimit    *options.RateLimitOptions // 新增
    RouteSetup   func(*gin.Engine) error
}
```

### 步骤 5：向后兼容

保留原有的 `NewGinServerFromConfig(log, opts)` 函数，使用默认中间件配置。

## 效果

### 优化前

```go
// 硬编码中间件，不可配置
ginServer := server.NewGinServerFromConfig(logger, serverOpts)
```

### 优化后

```go
// 使用完整配置，所有中间件可配置
ginServer := server.NewGinServerFromFullConfig(logger, &server.GinServerConfig{
    Server: serverOpts,
    CORS:   corsOpts,    // 来自配置文件
    JWT:    jwtOpts,     // 来自配置文件
})

// 或使用 HTTPServerInitializer（推荐）
httpInit := initializers.NewHTTPServerInitializer(
    &initializers.HTTPServerConfig{
        Name:   "MyHTTP",
        Server: serverOpts,
        CORS:   corsOpts,    // 中间件配置
        JWT:    jwtOpts,     // 中间件配置
        RouteSetup: setupRoutes,
    },
    logger,
)
```

## 优势

1. ✅ 所有中间件都可以通过配置文件配置
2. ✅ 配置来源统一（`common/options`）
3. ✅ 向后兼容（默认行为不变）
4. ✅ 易于扩展（添加新中间件只需添加新字段）
5. ✅ 符合依赖注入原则

## 配置文件示例

```yaml
# config.yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "release"

cors:
  enabled: true
  allow_origins: ["https://example.com"]
  allow_methods: ["GET", "POST", "PUT", "DELETE"]
  allow_credentials: true

jwt:
  secret: "your-secret-key"
  expires_hours: 24

rate_limit:
  enabled: true
  requests_per_minute: 100
```

## 总结

通过此优化，实现了：
- **配置驱动**: 所有中间件通过配置文件控制
- **统一管理**: 配置集中在 `common/options`
- **灵活扩展**: 易于添加新中间件
- **向后兼容**: 不影响现有代码
