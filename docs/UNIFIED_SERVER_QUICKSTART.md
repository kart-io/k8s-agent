# 统一服务器框架 - 快速开始指南

本指南演示如何在 Aetherius k8s-agent 项目中使用统一服务器框架。

---

## 🚀 5 分钟快速开始

### 1. 创建一个新服务

```go
package main

import (
    "context"
    "net/http"

    "github.com/kart-io/logger"
    "github.com/kart-io/k8s-agent/common/server"
    "github.com/kart-io/k8s-agent/common/contextx"
)

func main() {
    // 1. 初始化日志
    log := logger.New(logger.Config{
        Engine: logger.EngineZap,
        Level:  logger.LevelInfo,
    })

    // 2. 配置服务器选项
    opts := &server.ServerOptions{
        Host: "0.0.0.0",
        Port: 8080,
        Mode: "release",
        ReadTimeout:  30,
        WriteTimeout: 30,
        Middleware: []server.Middleware{
            server.TraceIDMiddleware(),      // 分布式追踪
            server.RequestIDMiddleware(),    // 请求 ID
            server.LoggerMiddleware(log),    // 日志记录
            server.RecoveryMiddleware(log),  // Panic 恢复
        },
    }

    // 3. 创建服务器（Gin 实现）
    httpServer, err := server.CreateGinServer(opts, log)
    if err != nil {
        log.Fatalw("Failed to create server", "error", err)
    }

    // 4. 注册路由
    httpServer.RegisterRoute("GET", "/health", healthHandler)
    httpServer.RegisterRoute("GET", "/api/v1/status", statusHandler)

    // 5. 启动服务器
    log.Infow("Starting server", "addr", httpServer.Addr())
    if err := httpServer.Start(context.Background()); err != nil {
        log.Fatalw("Server error", "error", err)
    }
}

// 健康检查处理器
func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"status":"healthy"}`))
}

// 状态检查处理器
func statusHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    traceID := contextx.GetTraceID(ctx)
    requestID := contextx.GetRequestID(ctx)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{
        "status":"running",
        "trace_id":"` + traceID + `",
        "request_id":"` + requestID + `"
    }`))
}
```

**运行**:
```bash
go run main.go
```

**测试**:
```bash
curl http://localhost:8080/health
# {"status":"healthy"}

curl http://localhost:8080/api/v1/status
# {"status":"running","trace_id":"xxx","request_id":"yyy"}
```

---

## 📦 在现有服务中使用

### 场景 1: 迁移现有 Gin 服务

**原有代码** (`cmd/myservice/main.go`):
```go
func main() {
    router := gin.New()
    router.Use(middleware.Recovery())
    router.Use(middleware.RequestID())
    router.GET("/api/v1/users", getUsersHandler)

    server := &http.Server{
        Addr:    ":8080",
        Handler: router,
    }
    server.ListenAndServe()
}
```

**迁移后** (使用统一接口):
```go
func main() {
    log := logger.New(logger.Config{Engine: logger.EngineZap})

    opts := &server.ServerOptions{
        Port: 8080,
        Middleware: []server.Middleware{
            server.RecoveryMiddleware(log),
            server.RequestIDMiddleware(),
        },
    }

    httpServer, _ := server.CreateGinServer(opts, log)
    httpServer.RegisterRoute("GET", "/api/v1/users", getUsersHandler)
    httpServer.Start(context.Background())
}
```

**优势**:
- ✅ 更清晰的配置结构
- ✅ 框架无关的中间件
- ✅ 自动的分布式追踪支持
- ✅ 更好的错误处理

---

### 场景 2: 使用配置文件选择框架

**配置文件** (`configs/config.yaml`):
```yaml
server:
  type: "gin"  # 可选: gin, kratos, http
  host: "0.0.0.0"
  port: 8080
  mode: "release"
  read_timeout: 30
  write_timeout: 30
  idle_timeout: 120
```

**代码** (`cmd/myservice/main.go`):
```go
package main

import (
    "github.com/spf13/viper"
    "github.com/kart-io/logger"
    "github.com/kart-io/k8s-agent/common/server"
)

type Config struct {
    Server struct {
        Type         string `mapstructure:"type"`
        Host         string `mapstructure:"host"`
        Port         int    `mapstructure:"port"`
        Mode         string `mapstructure:"mode"`
        ReadTimeout  int    `mapstructure:"read_timeout"`
        WriteTimeout int    `mapstructure:"write_timeout"`
        IdleTimeout  int    `mapstructure:"idle_timeout"`
    } `mapstructure:"server"`
}

func main() {
    // 加载配置
    viper.SetConfigFile("configs/config.yaml")
    viper.ReadInConfig()

    var cfg Config
    viper.Unmarshal(&cfg)

    // 初始化日志
    log := logger.New(logger.Config{Engine: logger.EngineZap})

    // 构建服务器选项
    opts := &server.ServerOptions{
        Host:         cfg.Server.Host,
        Port:         cfg.Server.Port,
        Mode:         cfg.Server.Mode,
        ReadTimeout:  cfg.Server.ReadTimeout,
        WriteTimeout: cfg.Server.WriteTimeout,
        IdleTimeout:  cfg.Server.IdleTimeout,
        Middleware:   buildMiddleware(log),
    }

    // 根据配置创建服务器
    serverType := server.ServerType(cfg.Server.Type)
    httpServer, err := server.CreateWithType(serverType, opts, log)
    if err != nil {
        log.Fatalw("Failed to create server", "error", err)
    }

    // 注册路由
    registerRoutes(httpServer)

    // 启动服务器
    if err := httpServer.Start(context.Background()); err != nil {
        log.Fatalw("Server error", "error", err)
    }
}

func buildMiddleware(log core.Logger) []server.Middleware {
    return []server.Middleware{
        server.TraceIDMiddleware(),
        server.RequestIDMiddleware(),
        server.LoggerMiddleware(log),
        server.RecoveryMiddleware(log),
        server.CORSMiddleware([]string{"*"}, nil, nil),
    }
}

func registerRoutes(httpServer server.HTTPServer) {
    httpServer.RegisterRoute("GET", "/health", healthHandler)
    httpServer.RegisterRoute("GET", "/api/v1/users", getUsersHandler)
    httpServer.RegisterRoute("POST", "/api/v1/users", createUserHandler)
}
```

**优势**:
- ✅ 通过修改配置文件切换框架
- ✅ 无需修改代码
- ✅ 适合不同环境使用不同框架

---

### 场景 3: 添加自定义中间件

**创建认证中间件**:
```go
package middleware

import (
    "net/http"
    "github.com/kart-io/k8s-agent/common/server"
    "github.com/kart-io/k8s-agent/common/contextx"
)

// AuthMiddleware 创建 JWT 认证中间件
func AuthMiddleware(jwtSecret string) server.Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // 提取 token
            token := r.Header.Get("Authorization")
            if token == "" {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            // 验证 token
            claims, err := validateJWT(token, jwtSecret)
            if err != nil {
                http.Error(w, "Forbidden", http.StatusForbidden)
                return
            }

            // 将用户信息添加到 context
            ctx := r.Context()
            ctx = contextx.WithUserID(ctx, claims.UserID)
            ctx = contextx.WithUsername(ctx, claims.Username)
            r = r.WithContext(ctx)

            next.ServeHTTP(w, r)
        })
    }
}

// RateLimitMiddleware 创建速率限制中间件
func RateLimitMiddleware(requestsPerMinute int) server.Middleware {
    limiter := rate.NewLimiter(rate.Limit(requestsPerMinute)/60, requestsPerMinute)

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !limiter.Allow() {
                http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

**使用自定义中间件**:
```go
opts := &server.ServerOptions{
    Port: 8080,
    Middleware: []server.Middleware{
        server.TraceIDMiddleware(),
        server.RequestIDMiddleware(),
        middleware.RateLimitMiddleware(100),      // 限速: 100 req/min
        middleware.AuthMiddleware(jwtSecret),     // JWT 认证
        server.LoggerMiddleware(log),
        server.RecoveryMiddleware(log),
    },
}
```

---

## 🔧 高级功能

### 1. 访问底层框架特性

有时需要使用框架特定的功能（如 Gin 的模板渲染）：

```go
httpServer, _ := server.CreateGinServer(opts, log)

// 类型断言获取 Gin 引擎
ginServer := httpServer.(*server.GinServer)
ginEngine := ginServer.Engine

// 使用 Gin 特定功能
ginEngine.LoadHTMLGlob("templates/*")
ginEngine.Static("/static", "./public")

// 使用 Gin 特定中间件
ginEngine.Use(ginSpecificMiddleware())
```

### 2. 优雅关闭

```go
package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"
    "time"
)

func main() {
    log := logger.New(logger.Config{Engine: logger.EngineZap})
    httpServer, _ := server.CreateGinServer(opts, log)

    // 在 goroutine 中启动服务器
    errChan := make(chan error, 1)
    go func() {
        if err := httpServer.Start(context.Background()); err != nil {
            errChan <- err
        }
    }()

    // 等待中断信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    select {
    case err := <-errChan:
        log.Fatalw("Server error", "error", err)
    case sig := <-quit:
        log.Infow("Received shutdown signal", "signal", sig)
    }

    // 优雅关闭（30 秒超时）
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    log.Infow("Shutting down server")
    if err := httpServer.Shutdown(ctx); err != nil {
        log.Errorw("Shutdown error", "error", err)
    }

    log.Infow("Server stopped gracefully")
}
```

### 3. 在处理器中使用 Context

```go
func getUserHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // 提取追踪信息
    traceID := contextx.GetTraceID(ctx)
    requestID := contextx.GetRequestID(ctx)
    userID := contextx.GetUserID(ctx)  // 来自认证中间件

    // 记录日志（自动包含追踪信息）
    log.Infow("Getting user",
        "trace_id", traceID,
        "request_id", requestID,
        "user_id", userID,
    )

    // 调用业务逻辑
    user, err := userService.Get(ctx, userID)
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // 返回响应
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}
```

---

## 📝 最佳实践

### 1. 中间件顺序

推荐的中间件顺序（从外到内）：

```go
opts.Middleware = []server.Middleware{
    server.TraceIDMiddleware(),       // 1. 最先生成 Trace ID
    server.RequestIDMiddleware(),     // 2. 生成 Request ID
    RateLimitMiddleware(100),         // 3. 限速（早期拦截）
    AuthMiddleware(jwtSecret),        // 4. 认证
    server.LoggerMiddleware(log),     // 5. 日志（记录已认证请求）
    server.RecoveryMiddleware(log),   // 6. Panic 恢复（最后保护）
    server.CORSMiddleware(...),       // 7. CORS（最后设置头）
}
```

### 2. 错误处理

```go
func handler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    traceID := contextx.GetTraceID(ctx)

    // 业务逻辑
    result, err := doSomething(ctx)
    if err != nil {
        // 记录错误（包含 trace ID）
        log.Errorw("Operation failed",
            "trace_id", traceID,
            "error", err,
        )

        // 返回错误响应
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // 成功响应
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}
```

### 3. 配置管理

```go
// config/config.go
type ServerConfig struct {
    Type         string `yaml:"type" env:"SERVER_TYPE" default:"gin"`
    Host         string `yaml:"host" env:"SERVER_HOST" default:"0.0.0.0"`
    Port         int    `yaml:"port" env:"SERVER_PORT" default:"8080"`
    Mode         string `yaml:"mode" env:"SERVER_MODE" default:"release"`
    ReadTimeout  int    `yaml:"read_timeout" env:"SERVER_READ_TIMEOUT" default:"30"`
    WriteTimeout int    `yaml:"write_timeout" env:"SERVER_WRITE_TIMEOUT" default:"30"`
}

// 支持环境变量覆盖
// SERVER_TYPE=kratos SERVER_PORT=9090 ./myservice
```

---

## 🐛 常见问题

### Q1: 中间件不生效？

**A**: 确保中间件在 `ServerOptions` 中注册，而不是在创建服务器后：

```go
// ✅ 正确
opts := &server.ServerOptions{
    Middleware: []server.Middleware{
        server.LoggerMiddleware(log),
    },
}
httpServer, _ := server.CreateGinServer(opts, log)

// ❌ 错误 - 太晚了
httpServer, _ := server.CreateGinServer(opts, log)
httpServer.Use(server.LoggerMiddleware(log))  // 可能不会应用到所有路由
```

### Q2: Trace ID 为空？

**A**: 确保 `TraceIDMiddleware` 在中间件链的最前面：

```go
opts.Middleware = []server.Middleware{
    server.TraceIDMiddleware(),  // 必须在第一位
    server.RequestIDMiddleware(),
    // ... 其他中间件
}
```

### Q3: 如何测试使用统一接口的服务？

**A**: 使用 `httptest` 包：

```go
func TestHandler(t *testing.T) {
    // 创建测试服务器
    opts := &server.ServerOptions{Port: 0}
    httpServer, _ := server.CreateGinServer(opts, testLogger)
    httpServer.RegisterRoute("GET", "/test", testHandler)

    // 创建测试请求
    req := httptest.NewRequest("GET", "/test", nil)
    w := httptest.NewRecorder()

    // 执行请求（通过 Gin 引擎）
    ginServer := httpServer.(*server.GinServer)
    ginServer.Engine.ServeHTTP(w, req)

    // 验证响应
    assert.Equal(t, http.StatusOK, w.Code)
}
```

---

## 📚 更多资源

- **完整文档**: `common/server/README.md`
- **代码示例**: `common/server/example_test.go`
- **实施报告**: `docs/FRAMEWORK_UNIFICATION_COMPLETE.md`
- **OneX 分析**: `docs/ONEX_ARCHITECTURE_ANALYSIS.md`

---

## 🎯 下一步

1. **在新服务中使用**
   - 创建服务时优先使用统一接口
   - 参考本指南的示例代码

2. **迁移现有服务**
   - 从最简单的服务开始（如 gateway）
   - 逐步迁移其他服务

3. **等待 Phase 2**
   - Kratos 服务器完整实现
   - 标准 HTTP 服务器支持
   - 更多框架选择

---

**版本**: 1.0.0
**状态**: 生产就绪 (Gin)
**更新日期**: 2025-11-02
