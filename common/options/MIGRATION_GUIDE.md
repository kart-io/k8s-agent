# Options 配置使用指南

本指南介绍 `common/options` 包中推荐使用的配置选项，以及如何正确使用它们。

## ⚠️ 重要说明

**已移除的配置选项**：
- ❌ `InsecureServingOptions` - 已删除，功能过于简化
- ❌ `SecureServingOptions` - 已删除，未被实际使用

如果您的代码中使用了这些配置，请参考本文档迁移到推荐的配置选项。

## 📋 配置选项清单

### ✅ 推荐使用的配置

#### HTTP 服务器配置

1. **ServerOptions** - 适合 Gin/Kratos 框架
   ```go
   opts := options.NewServerOptions()
   opts.Host = "0.0.0.0"
   opts.Port = 8080
   opts.Mode = "release"
   opts.ReadTimeout = 10 * time.Second
   opts.WriteTimeout = 10 * time.Second
   opts.IdleTimeout = 60 * time.Second

   // 创建 Gin 服务器
   ginSrv := server.NewGinServerFromConfig(log, opts)
   ```

2. **HTTPServerOptions** - 适合原生 HTTP 服务器（OneX 风格）
   ```go
   opts := options.NewHTTPServerOptions()
   opts.Network = "tcp"
   opts.Addr = "0.0.0.0:8080"
   opts.ReadTimeout = 15 * time.Second
   opts.WriteTimeout = 15 * time.Second
   opts.IdleTimeout = 60 * time.Second
   opts.MaxHeaderBytes = 1 << 20 // 1MB

   // 创建 HTTP 服务器
   httpSrv := server.NewHTTPOptionsServer(opts, nil, handler, log)
   ```

#### HTTPS/TLS 配置

**TLSOptions** - 统一的 TLS 配置
```go
tlsOpts := options.NewTLSOptions()
tlsOpts.UseTLS = true
tlsOpts.Cert = "/path/to/cert.pem"
tlsOpts.Key = "/path/to/key.pem"
tlsOpts.CaCert = "/path/to/ca.pem"  // 可选，用于 mTLS
tlsOpts.MinVersion = tls.VersionTLS12

// 与 HTTP 服务器组合使用
httpSrv := server.NewHTTPOptionsServer(httpOpts, tlsOpts, handler, log)
```

#### gRPC 服务器配置

**GRPCOptions** - gRPC 服务器配置
```go
opts := options.NewGRPCOptions()
opts.Host = "0.0.0.0"
opts.Port = 9090
opts.MaxRecvMsgSize = 10 * 1024 * 1024  // 10MB
opts.MaxSendMsgSize = 10 * 1024 * 1024
opts.KeepAliveTime = 30 * time.Second
opts.EnableReflection = true
opts.EnableHealthCheck = true

// 创建 gRPC 服务器
grpcSrv, _ := server.NewStandardGRPCServerFromConfig(log, opts)
```

### ❌ 已移除的配置（迁移指南）

以下配置选项已从代码库中移除，因为它们未被实际使用或功能过于简化。

#### 1. InsecureServingOptions（已移除）

**移除原因**：功能过于简化，仅包含一个 `Addr` 字段，无法配置超时等重要参数。

**如果您的代码使用了此配置，请迁移到**：

```go
// ❌ 旧代码（已不可用）
opts := options.NewInsecureServingOptions()
opts.Addr = "0.0.0.0:8080"

// ✅ 新代码 - 方式1: 使用 ServerOptions（推荐用于 Gin/Kratos）
opts := options.NewServerOptions()
opts.Host = "0.0.0.0"
opts.Port = 8080

// ✅ 新代码 - 方式2: 使用 HTTPServerOptions（推荐用于原生 HTTP）
opts := options.NewHTTPServerOptions()
opts.Addr = "0.0.0.0:8080"
```

#### 2. SecureServingOptions（已移除）

**移除原因**：未被 `common/server` 实际使用，功能可以用 `HTTPServerOptions + TLSOptions` 更好地替代。

**如果您的代码使用了此配置，请迁移到**：

```go
// ❌ 旧代码（已不可用）
opts := options.NewSecureServingOptions()
opts.BindAddress = "0.0.0.0"
opts.BindPort = 8443
opts.ServerCert.CertKey.CertFile = "/path/to/cert.pem"
opts.ServerCert.CertKey.KeyFile = "/path/to/key.pem"

// ✅ 新代码：使用 HTTPServerOptions + TLSOptions
httpOpts := options.NewHTTPServerOptions()
httpOpts.Addr = "0.0.0.0:8443"

tlsOpts := options.NewTLSOptions()
tlsOpts.UseTLS = true
tlsOpts.Cert = "/path/to/cert.pem"
tlsOpts.Key = "/path/to/key.pem"

// 创建 HTTPS 服务器
httpsSrv := server.NewHTTPOptionsServer(httpOpts, tlsOpts, handler, log)
```

## 🔧 配置选项对比

### ServerOptions vs HTTPServerOptions

| 特性 | ServerOptions | HTTPServerOptions |
|------|---------------|-------------------|
| **适用场景** | Gin/Kratos 框架 | 原生 HTTP 服务器 |
| **地址配置** | Host + Port (分离) | Addr (组合) |
| **网络类型** | ❌ 固定 TCP | ✅ 支持 tcp/tcp4/tcp6/unix |
| **运行模式** | ✅ debug/release/test | ❌ 无 |
| **超时配置** | ReadTimeout, WriteTimeout, IdleTimeout | 同左 + Timeout (客户端) |
| **高级配置** | GracefulStop 超时 | MaxHeaderBytes |
| **推荐使用** | ✅ Gin/Kratos 场景 | ✅ 原生 HTTP 场景 |

### 选择建议

```go
// 场景1：使用 Gin 框架
opts := options.NewServerOptions()
ginSrv := server.NewGinServerFromConfig(log, opts)

// 场景2：使用 Kratos 框架
opts := options.NewServerOptions()
kratosSrv := server.NewKratosServerFromConfig(log, opts)

// 场景3：使用原生 HTTP 服务器（OneX 风格）
httpOpts := options.NewHTTPServerOptions()
httpSrv := server.NewHTTPOptionsServer(httpOpts, nil, handler, log)

// 场景4：需要 HTTPS
httpOpts := options.NewHTTPServerOptions()
tlsOpts := options.NewTLSOptions()
tlsOpts.UseTLS = true
httpsSrv := server.NewHTTPOptionsServer(httpOpts, tlsOpts, handler, log)
```

## 📚 完整示例

### 示例1：Gin HTTP 服务器（生产环境）

```go
package main

import (
    "context"
    "github.com/kart-io/k8s-agent/common/options"
    "github.com/kart-io/k8s-agent/common/server"
    "github.com/spf13/viper"
)

func main() {
    // 1. 加载配置文件
    v := viper.New()
    v.SetConfigFile("config.yaml")
    v.ReadInConfig()

    // 2. 解析 HTTP 服务器配置
    var httpOpts options.ServerOptions
    v.UnmarshalKey("server.http", &httpOpts)
    httpOpts.Complete()
    httpOpts.Validate()

    // 3. 创建 Gin 服务器
    log := logger.New(logger.Config{Level: logger.LevelInfo})
    ginSrv := server.NewGinServerFromConfig(log, &httpOpts)

    // 4. 注册路由
    ginSrv.Engine.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    // 5. 启动服务器
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    if err := server.Serve(ctx, ginSrv, log); err != nil {
        log.Fatalw("Server failed", "err", err)
    }
}
```

**配置文件 (config.yaml)**:
```yaml
server:
  http:
    host: 0.0.0.0
    port: 8080
    mode: release
    read_timeout: 10s
    write_timeout: 10s
    idle_timeout: 60s
```

### 示例2：HTTPS 服务器（使用 TLS）

```go
package main

import (
    "context"
    "net/http"
    "github.com/kart-io/k8s-agent/common/options"
    "github.com/kart-io/k8s-agent/common/server"
)

func main() {
    log := logger.New(logger.Config{Level: logger.LevelInfo})

    // HTTP 服务器配置
    httpOpts := options.NewHTTPServerOptions()
    httpOpts.Addr = "0.0.0.0:8443"

    // TLS 配置
    tlsOpts := options.NewTLSOptions()
    tlsOpts.UseTLS = true
    tlsOpts.Cert = "/etc/certs/server.crt"
    tlsOpts.Key = "/etc/certs/server.key"

    // 创建请求处理器
    mux := http.NewServeMux()
    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(200)
        w.Write([]byte(`{"status":"ok"}`))
    })

    // 创建 HTTPS 服务器
    httpsSrv := server.NewHTTPOptionsServer(httpOpts, tlsOpts, mux, log)

    // 启动服务器
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    if err := server.Serve(ctx, httpsSrv, log); err != nil {
        log.Fatalw("HTTPS server failed", "err", err)
    }
}
```

### 示例3：gRPC 服务器

```go
package main

import (
    "context"
    "github.com/kart-io/k8s-agent/common/options"
    "github.com/kart-io/k8s-agent/common/server"
)

func main() {
    log := logger.New(logger.Config{Level: logger.LevelInfo})

    // gRPC 服务器配置
    grpcOpts := options.NewGRPCOptions()
    grpcOpts.Port = 9090
    grpcOpts.EnableReflection = true
    grpcOpts.EnableHealthCheck = true

    // 创建 gRPC 服务器
    grpcSrv, err := server.NewStandardGRPCServerFromConfig(log, grpcOpts)
    if err != nil {
        log.Fatalw("Failed to create gRPC server", "err", err)
    }

    // 注册 gRPC 服务
    // mypb.RegisterMyServiceServer(grpcSrv.GetServer(), myServiceImpl)

    // 启动服务器
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    if err := server.Serve(ctx, grpcSrv, log); err != nil {
        log.Fatalw("gRPC server failed", "err", err)
    }
}
```

## 🔍 常见问题

### Q1: ServerOptions 和 HTTPServerOptions 有什么区别？

**A**:
- `ServerOptions`: 简化版，适合 Gin/Kratos 框架，使用 Host+Port 分离配置
- `HTTPServerOptions`: 完整版，适合原生 HTTP 服务器，使用 Addr 组合配置，支持更多网络类型

### Q2: 为什么废弃 InsecureServingOptions？

**A**: 功能过于简化，只有一个 `Addr` 字段，无法配置超时、网络类型等重要参数。`ServerOptions` 和 `HTTPServerOptions` 提供了更完整的配置。

### Q3: 如何配置 HTTPS？

**A**: 使用 `HTTPServerOptions + TLSOptions` 组合：
```go
httpOpts := options.NewHTTPServerOptions()
tlsOpts := options.NewTLSOptions()
tlsOpts.UseTLS = true
tlsOpts.Cert = "/path/to/cert.pem"
tlsOpts.Key = "/path/to/key.pem"

srv := server.NewHTTPOptionsServer(httpOpts, tlsOpts, handler, log)
```

### Q4: 旧代码还能用吗？

**A**: 可以，但已标记为 `Deprecated`。建议尽快迁移到新配置，旧配置可能在未来版本中移除。

## 📖 相关文档

- [common/server/README.md](../server/README.md) - 服务器包使用指南
- [common/options/README.md](README.md) - 配置选项包文档（如果存在）

---

**更新时间**: 2025-11-02
**状态**: ✅ 配置清理完成
**变更**:
- ❌ 已删除 `InsecureServingOptions` (未使用)
- ❌ 已删除 `SecureServingOptions` (未使用)
- ✅ 保留 `ServerOptions` (用于 Gin/Kratos)
- ✅ 保留 `HTTPServerOptions` (用于原生 HTTP)
- ✅ 保留 `GRPCOptions` (用于 gRPC)
