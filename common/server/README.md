# Server Package

统一的 HTTP 和 gRPC 服务器实现，支持多种框架和配置模式。参考 OneX 项目设计，提供统一的 Server 接口和生命周期管理。

## ⚠️ 重要说明：配置选项整合

**本包现已与 `common/options` 完全整合**：

- ✅ **HTTP 服务器** 使用 `common/options.ServerOptions`
- ✅ **gRPC 服务器** 使用 `common/options.GRPCOptions`
- ✅ **健康检查** 使用 `common/options.HealthOptions`

**设计原则**：
1. `common/options/` - 存放配置选项定义（数据结构 + 验证 + 命令行标志）
2. `common/server/` - 使用 `common/options` 的配置创建服务器实例
3. 保留函数式 options 模式用于简单场景（如单元测试）
4. 提供 `From<Options>` 工厂方法用于从配置对象创建（推荐用于生产环境）

**两种使用方式**：

```go
// 方式1: 函数式选项（简单场景、测试）
ginSrv := server.NewGinServer(log,
    server.WithGinPort(8080),
    server.WithGinMode("release"),
)

// 方式2: 配置对象（推荐用于生产环境、配置文件）
opts := options.NewServerOptions()
opts.Port = 8080
opts.Mode = "release"
ginSrv := server.NewGinServerFromConfig(log, opts)
```

**优势**：
- 🎯 统一配置来源：`common/options` 是唯一的配置定义位置
- 📝 支持配置文件：通过 Viper 加载 YAML/JSON 配置
- ✅ 自动验证：配置对象提供 `Validate()` 和 `Complete()` 方法
- 🚀 灵活性：简单场景用函数式选项，复杂场景用配置对象

## 📁 目录结构

```
common/server/
├── server.go              # 核心Server接口 + Serve/MultiServe
├── README.md
│
├── http/                  # HTTP服务器实现（package server）
│   ├── gin.go            # Gin框架实现
│   ├── kratos.go         # Kratos框架实现
│   └── options.go        # Options配置HTTP实现
│
├── grpc/                  # gRPC服务器实现（package server）
│   ├── standard.go       # 标准gRPC实现
│   ├── options.go        # Options配置gRPC实现
│   ├── interceptors.go   # gRPC拦截器
│   └── health.go         # gRPC健康检查
│
└── internal/              # 共享内部代码（package internal）
    ├── middleware.go     # 中间件定义
    └── health.go         # 通用健康检查
```

**包结构说明:**
- `http/` 和 `grpc/` 目录中的文件属于 `package server`，与根目录 `server.go` 在同一个包中
- `internal/` 目录是独立的 `package internal`，提供内部工具函数
- 用户只需导入：`import "github.com/kart-io/k8s-agent/common/server"`

## 🎯 核心设计

### Server 接口

所有服务器实现都遵循统一的 Server 接口：

```go
type Server interface {
    RunOrDie()                       // 启动服务器，失败则退出
    GracefulStop(ctx context.Context) // 优雅关停
}
```

### 生命周期管理

```go
// 单服务器
func Serve(ctx context.Context, srv Server, log core.Logger) error

// 多服务器（同时运行HTTP和gRPC）
func MultiServe(ctx context.Context, log core.Logger, servers ...Server) error
```

## 🚀 快速开始

### 1. Gin HTTP 服务器

```go
import (
    "github.com/kart-io/k8s-agent/common/server"
    "github.com/kart-io/k8s-agent/common/options"
)

// 方式1: 函数式选项（简单场景）
ginSrv := server.NewGinServer(log,
    server.WithGinPort(8080),
    server.WithGinMode("release"),
)

// 方式2: 配置对象（推荐用于生产环境）
opts := options.NewServerOptions()
opts.Port = 8080
opts.Mode = "release"
ginSrv := server.NewGinServerFromConfig(log, opts)

// 注册路由
ginSrv.Engine.GET("/health", func(c *gin.Context) {
    c.JSON(200, gin.H{"status": "ok"})
})

// 启动
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
server.Serve(ctx, ginSrv, log)
```

### 2. Kratos HTTP 服务器

```go
// 方式1: 函数式选项
kratosSrv := server.NewKratosServer(log,
    server.WithKratosPort(8080),
)

// 方式2: 配置对象
opts := options.NewServerOptions()
opts.Port = 8080
kratosSrv := server.NewKratosServerFromConfig(log, opts)

// 注册路由
kratosSrv.GetServer().Route("/").GET("/health", ...)

// 启动
server.Serve(ctx, kratosSrv, log)
```

### 3. gRPC 服务器

```go
// 方式1: 函数式选项
grpcSrv, _ := server.NewStandardGRPCServer(log,
    server.WithGRPCPort(9090),
    server.WithGRPCReflection(true),
)

// 方式2: 配置对象
opts := options.NewGRPCOptions()
opts.Port = 9090
opts.EnableReflection = true
grpcSrv, _ := server.NewStandardGRPCServerFromConfig(log, opts)

// 注册服务
mypb.RegisterMyServiceServer(grpcSrv.GetServer(), myImpl)

// 启动
server.Serve(ctx, grpcSrv, log)
```

### 4. 从配置文件加载（生产环境推荐）

```go
import (
    "github.com/spf13/viper"
    "github.com/kart-io/k8s-agent/common/options"
    "github.com/kart-io/k8s-agent/common/server"
)

// 加载配置文件
v := viper.New()
v.SetConfigFile("config.yaml")
v.ReadInConfig()

// 解析 HTTP 服务器配置
var httpOpts options.ServerOptions
v.UnmarshalKey("server.http", &httpOpts)
httpOpts.Complete()  // 填充默认值
httpOpts.Validate()  // 验证配置

// 解析 gRPC 服务器配置
var grpcOpts options.GRPCOptions
v.UnmarshalKey("server.grpc", &grpcOpts)
grpcOpts.Complete()
grpcOpts.Validate()

// 创建服务器
httpSrv := server.NewGinServerFromConfig(log, &httpOpts)
grpcSrv, _ := server.NewStandardGRPCServerFromConfig(log, &grpcOpts)

// 同时启动
server.MultiServe(ctx, log, httpSrv, grpcSrv)
```

**配置文件示例 (config.yaml)**:

```yaml
server:
  http:
    host: 0.0.0.0
    port: 8080
    mode: release
    read_timeout: 10s
    write_timeout: 10s
    idle_timeout: 60s

  grpc:
    enable: true
    host: 0.0.0.0
    port: 9090
    max_recv_msg_size: 10485760  # 10MB
    max_send_msg_size: 10485760
    keep_alive_time: 30s
    enable_reflection: true
    enable_health_check: true
```

### 5. 同时运行 HTTP 和 gRPC

```go
// 使用函数式选项
httpSrv := server.NewGinServer(log, server.WithGinPort(8080))
grpcSrv, _ := server.NewStandardGRPCServer(log, server.WithGRPCPort(9090))

// 或使用配置对象
httpOpts := options.NewServerOptions()
grpcOpts := options.NewGRPCOptions()
httpSrv := server.NewGinServerFromConfig(log, httpOpts)
grpcSrv, _ := server.NewStandardGRPCServerFromConfig(log, grpcOpts)

// 同时启动
server.MultiServe(ctx, log, httpSrv, grpcSrv)
```


## 📦 可用的服务器实现

| 实现 | 文件 | 用途 | 特性 |
|------|------|------|------|
| **GinServer** | `http/gin.go` | Gin Web应用 | 内置中间件，开箱即用 |
| **KratosServer** | `http/kratos.go` | Kratos微服务 | 与Kratos深度集成 |
| **HTTPOptionsServer** | `http/options.go` | HTTP/HTTPS | TLS支持，Options配置 |
| **StandardGRPCServer** | `grpc/standard.go` | gRPC服务 | 函数式配置，灵活 |
| **GRPCOptionsServer** | `grpc/options.go` | gRPC服务 | Options配置驱动 |

## 🔧 配置选项

### Gin 配置

```go
server.WithGinHost(host string)
server.WithGinPort(port int)
server.WithGinMode(mode string)         // debug/release/test
server.WithGinReadTimeout(d time.Duration)
server.WithGinWriteTimeout(d time.Duration)
server.WithGinIdleTimeout(d time.Duration)
```

### Kratos 配置

```go
server.WithKratosHost(host string)
server.WithKratosPort(port int)
server.WithKratosReadTimeout(d time.Duration)
server.WithKratosWriteTimeout(d time.Duration)
server.WithKratosIdleTimeout(d time.Duration)
```

### gRPC 配置

```go
server.WithGRPCHost(host string)
server.WithGRPCPort(port int)
server.WithGRPCMaxRecvMsgSize(size int)
server.WithGRPCMaxSendMsgSize(size int)
server.WithGRPCKeepAliveTime(d time.Duration)
server.WithGRPCReflection(enabled bool)
server.WithGRPCHealthCheck(enabled bool)
server.WithGRPCTLS(tlsConfig *tls.Config)
server.WithGRPCUnaryInterceptor(interceptor)
server.WithGRPCStreamInterceptor(interceptor)
```

## 💡 使用示例

### 完整示例：HTTP + gRPC + 信号处理

```go
package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"

    "github.com/kart-io/k8s-agent/common/server"
    "github.com/kart-io/logger"
)

func main() {
    log := logger.New(logger.Config{Level: logger.LevelInfo})

    // 创建HTTP服务器
    httpSrv := server.NewGinServer(log, server.WithGinPort(8080))
    httpSrv.Engine.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    // 创建gRPC服务器
    grpcSrv, _ := server.NewStandardGRPCServer(log,
        server.WithGRPCPort(9090),
        server.WithGRPCReflection(true),
    )
    // mypb.RegisterMyServiceServer(grpcSrv.GetServer(), myImpl)

    // 设置信号监听
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    go func() {
        sigCh := make(chan os.Signal, 1)
        signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
        <-sigCh
        log.Infow("Received shutdown signal")
        cancel()
    }()

    // 同时启动所有服务器
    log.Infow("Starting servers...")
    if err := server.MultiServe(ctx, log, httpSrv, grpcSrv); err != nil {
        log.Fatalw("Servers failed", "err", err)
    }

    log.Infow("All servers shut down successfully")
}
```

## 🔄 迁移指南

### 从旧代码迁移

```go
// 旧代码（已弃用）
ginSrv.Run()  // 阻塞启动

// 新代码
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
server.Serve(ctx, ginSrv, log)  // 统一生命周期管理
```

## 📚 参考

- OneX 项目: https://github.com/onexstack/onex
- Options 配置: `common/options/`
- 示例代码: `example_test.go`

---

**重构完成时间**: 2025-11-02
**状态**: ✅ 生产就绪
