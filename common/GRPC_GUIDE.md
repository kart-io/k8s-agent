# gRPC Server 使用指南

本指南展示如何在 common 库中使用 gRPC 服务器，并使用 buf 管理 protobuf 文件。

## 目录

1. [快速开始](#快速开始)
2. [使用 buf 生成代码](#使用-buf-生成代码)
3. [创建 gRPC 服务](#创建-grpc-服务)
4. [配置选项](#配置选项)
5. [拦截器](#拦截器)
6. [健康检查](#健康检查)
7. [完整示例](#完整示例)

## 快速开始

### 1. 安装 buf

```bash
# macOS
brew install bufbuild/buf/buf

# Linux
BIN="/usr/local/bin" && \
VERSION="1.28.1" && \
  curl -sSL \
    "https://github.com/bufbuild/buf/releases/download/v${VERSION}/buf-$(uname -s)-$(uname -m)" \
    -o "${BIN}/buf" && \
  chmod +x "${BIN}/buf"

# 或使用 Go
go install github.com/bufbuild/buf/cmd/buf@latest
```

### 2. 基础使用

```go
import (
    "github.com/kart-io/k8s-agent/common/server"
    "go.uber.org/zap"
)

func main() {
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    // 创建 gRPC 服务器（使用默认配置）
    grpcServer, err := server.NewGRPCServer(logger)
    if err != nil {
        logger.Fatal("Failed to create gRPC server", zap.Error(err))
    }

    // 注册你的服务
    // pb.RegisterYourServiceServer(grpcServer.Server(), yourServiceImpl)

    // 启动服务器
    if err := grpcServer.Run(); err != nil {
        logger.Fatal("Failed to run gRPC server", zap.Error(err))
    }
}
```

## 使用 buf 生成代码

### 目录结构

```
common/
├── proto/
│   ├── buf.yaml          # buf 配置文件
│   ├── buf.gen.yaml      # 代码生成配置
│   ├── buf.work.yaml     # workspace 配置
│   ├── common/
│   │   ├── health/v1/
│   │   │   └── health.proto
│   │   └── example/v1/
│   │       └── example.proto
│   └── gen/              # 生成的代码（由 buf 生成）
│       ├── common/
│       │   ├── health/v1/
│       │   └── example/v1/
│       └── openapiv2/
└── server/
    ├── grpc.go
    ├── grpc_interceptors.go
    └── grpc_health.go
```

### 生成代码

```bash
# 在 common/proto 目录下执行
cd common/proto

# 生成 Go 代码
buf generate

# 或使用 Makefile（见下文）
make proto-gen
```

### 验证 protobuf 定义

```bash
# 检查格式
buf format -w

# 检查 lint 规则
buf lint

# 检查 breaking changes
buf breaking --against '.git#branch=master'
```

## 创建 gRPC 服务

### 1. 定义 Protobuf

创建文件 `proto/myservice/v1/myservice.proto`:

```protobuf
syntax = "proto3";

package myservice.v1;

option go_package = "github.com/kart-io/k8s-agent/common/proto/gen/myservice/v1;myservicev1";

service MyService {
  rpc GetData(GetDataRequest) returns (GetDataResponse) {}
}

message GetDataRequest {
  string id = 1;
}

message GetDataResponse {
  string data = 1;
}
```

### 2. 生成代码

```bash
cd proto
buf generate
```

### 3. 实现服务

```go
package myservice

import (
    "context"

    pb "github.com/kart-io/k8s-agent/common/proto/gen/myservice/v1"
)

type MyServiceImpl struct {
    pb.UnimplementedMyServiceServer
}

func (s *MyServiceImpl) GetData(ctx context.Context, req *pb.GetDataRequest) (*pb.GetDataResponse, error) {
    return &pb.GetDataResponse{
        Data: "Hello, " + req.Id,
    }, nil
}
```

### 4. 注册服务

```go
grpcServer, _ := server.NewGRPCServer(logger)

// 注册服务
pb.RegisterMyServiceServer(grpcServer.Server(), &MyServiceImpl{})

// 启动服务器
grpcServer.Run()
```

## 配置选项

### 基础配置

```go
grpcServer, err := server.NewGRPCServer(logger,
    server.WithGRPCHost("0.0.0.0"),
    server.WithGRPCPort(9090),
)
```

### 消息大小限制

```go
grpcServer, err := server.NewGRPCServer(logger,
    server.WithGRPCMaxRecvMsgSize(10 * 1024 * 1024), // 10MB
    server.WithGRPCMaxSendMsgSize(10 * 1024 * 1024), // 10MB
)
```

### 连接超时和 KeepAlive

```go
grpcServer, err := server.NewGRPCServer(logger,
    server.WithGRPCConnectionTimeout(2 * time.Minute),
    server.WithGRPCKeepAliveTime(30 * time.Second),
    server.WithGRPCKeepAliveTimeout(10 * time.Second),
    server.WithGRPCMaxConnectionIdle(5 * time.Minute),
    server.WithGRPCMaxConnectionAge(30 * time.Minute),
)
```

### TLS 配置

```go
import "crypto/tls"

tlsConfig := &tls.Config{
    Certificates: []tls.Certificate{cert},
    ClientAuth:   tls.RequireAndVerifyClientCert,
    ClientCAs:    caCertPool,
}

grpcServer, err := server.NewGRPCServer(logger,
    server.WithGRPCTLS(tlsConfig),
)
```

### 启用/禁用特性

```go
grpcServer, err := server.NewGRPCServer(logger,
    server.WithGRPCReflection(true),    // 启用反射服务（用于 grpcurl）
    server.WithGRPCHealthCheck(true),   // 启用健康检查
)
```

## 拦截器

### 使用内置拦截器

```go
grpcServer, err := server.NewGRPCServer(logger,
    // 添加日志拦截器
    server.WithGRPCUnaryInterceptor(server.LoggingUnaryInterceptor(logger)),
    server.WithGRPCStreamInterceptor(server.LoggingStreamInterceptor(logger)),

    // 添加恢复拦截器
    server.WithGRPCUnaryInterceptor(server.RecoveryUnaryInterceptor(logger)),
    server.WithGRPCStreamInterceptor(server.RecoveryStreamInterceptor(logger)),

    // 添加请求 ID 拦截器
    server.WithGRPCUnaryInterceptor(server.RequestIDUnaryInterceptor()),
    server.WithGRPCStreamInterceptor(server.RequestIDStreamInterceptor()),
)
```

### 自定义拦截器

```go
func MyCustomInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        // 前置处理
        logger.Info("Before handler")

        // 调用实际处理器
        resp, err := handler(ctx, req)

        // 后置处理
        logger.Info("After handler")

        return resp, err
    }
}

// 使用自定义拦截器
grpcServer, err := server.NewGRPCServer(logger,
    server.WithGRPCUnaryInterceptor(MyCustomInterceptor(logger)),
)
```

## 健康检查

### 标准健康检查

gRPC 服务器默认启用标准健康检查协议：

```bash
# 使用 grpc-health-probe
grpc-health-probe -addr=localhost:9090

# 使用 grpcurl
grpcurl -plaintext localhost:9090 grpc.health.v1.Health/Check
```

### 自定义健康检查

```go
// 实现健康检查器
type MySQLHealthChecker struct {
    db *gorm.DB
}

func (c *MySQLHealthChecker) HealthCheck() (bool, error) {
    sqlDB, err := c.db.DB()
    if err != nil {
        return false, err
    }
    return sqlDB.Ping() == nil, nil
}

// 注册健康检查器
healthManager := server.NewHealthCheckManager()
healthManager.RegisterChecker("mysql", &MySQLHealthChecker{db: mysqlDB})

// 手动设置服务状态
grpcServer.HealthServer().SetServingStatus("myservice", grpc_health_v1.HealthCheckResponse_SERVING)
grpcServer.HealthServer().SetServingStatus("myservice", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
```

## 完整示例

### 服务端完整示例

```go
package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"
    "time"

    "go.uber.org/zap"
    "google.golang.org/grpc/health/grpc_health_v1"

    "github.com/kart-io/k8s-agent/common/server"
    pb "github.com/kart-io/k8s-agent/common/proto/gen/myservice/v1"
)

type MyServiceImpl struct {
    pb.UnimplementedMyServiceServer
    logger *zap.Logger
}

func (s *MyServiceImpl) GetData(ctx context.Context, req *pb.GetDataRequest) (*pb.GetDataResponse, error) {
    s.logger.Info("GetData called", zap.String("id", req.Id))
    return &pb.GetDataResponse{
        Data: "Hello, " + req.Id,
    }, nil
}

func main() {
    // 初始化日志
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    // 创建 gRPC 服务器
    grpcServer, err := server.NewGRPCServer(logger,
        server.WithGRPCPort(9090),
        server.WithGRPCReflection(true),
        server.WithGRPCHealthCheck(true),
        // 添加拦截器
        server.WithGRPCUnaryInterceptor(server.LoggingUnaryInterceptor(logger)),
        server.WithGRPCUnaryInterceptor(server.RecoveryUnaryInterceptor(logger)),
        server.WithGRPCUnaryInterceptor(server.RequestIDUnaryInterceptor()),
    )
    if err != nil {
        logger.Fatal("Failed to create gRPC server", zap.Error(err))
    }

    // 注册服务
    myService := &MyServiceImpl{logger: logger}
    pb.RegisterMyServiceServer(grpcServer.Server(), myService)

    // 设置健康状态
    grpcServer.HealthServer().SetServingStatus("myservice", grpc_health_v1.HealthCheckResponse_SERVING)

    // 在 goroutine 中启动服务器
    go func() {
        if err := grpcServer.Run(); err != nil {
            logger.Fatal("Failed to run gRPC server", zap.Error(err))
        }
    }()

    logger.Info("gRPC server started on :9090")

    // 等待中断信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    logger.Info("Shutting down server...")

    // 优雅关闭
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := grpcServer.Shutdown(ctx); err != nil {
        logger.Error("Server forced to shutdown", zap.Error(err))
    }

    logger.Info("Server exited")
}
```

### 客户端示例

```go
package main

import (
    "context"
    "log"
    "time"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"

    pb "github.com/kart-io/k8s-agent/common/proto/gen/myservice/v1"
)

func main() {
    // 连接到 gRPC 服务器
    conn, err := grpc.Dial("localhost:9090",
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithBlock(),
    )
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()

    // 创建客户端
    client := pb.NewMyServiceClient(conn)

    // 调用 RPC
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()

    resp, err := client.GetData(ctx, &pb.GetDataRequest{Id: "123"})
    if err != nil {
        log.Fatalf("Failed to call GetData: %v", err)
    }

    log.Printf("Response: %s", resp.Data)
}
```

## 默认配置值

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| Host | `0.0.0.0` | 监听地址 |
| Port | `9090` | 监听端口 |
| MaxRecvMsgSize | `4MB` | 最大接收消息大小 |
| MaxSendMsgSize | `4MB` | 最大发送消息大小 |
| ConnectionTimeout | `120s` | 连接超时 |
| KeepAliveTime | `30s` | KeepAlive 时间间隔 |
| KeepAliveTimeout | `10s` | KeepAlive 超时 |
| MaxConnectionIdle | `5m` | 最大连接空闲时间 |
| MaxConnectionAge | `30m` | 最大连接存活时间 |
| MaxConnectionAgeGrace | `5s` | 连接关闭宽限时间 |
| EnableReflection | `true` | 启用反射服务 |
| EnableHealthCheck | `true` | 启用健康检查 |

## 调试工具

### 使用 grpcurl

```bash
# 列出所有服务
grpcurl -plaintext localhost:9090 list

# 列出服务方法
grpcurl -plaintext localhost:9090 list myservice.v1.MyService

# 调用方法
grpcurl -plaintext -d '{"id": "123"}' localhost:9090 myservice.v1.MyService/GetData

# 健康检查
grpcurl -plaintext localhost:9090 grpc.health.v1.Health/Check
```

### 使用 grpcui

```bash
# 安装 grpcui
go install github.com/fullstorydev/grpcui/cmd/grpcui@latest

# 启动 Web UI
grpcui -plaintext localhost:9090
```

## 性能优化建议

1. **合理设置消息大小限制**：根据实际业务需求调整 `MaxRecvMsgSize` 和 `MaxSendMsgSize`
2. **使用连接池**：客户端应该复用 gRPC 连接，避免频繁创建连接
3. **启用 KeepAlive**：在长连接场景下启用 KeepAlive 防止连接被中间设备断开
4. **使用流式 RPC**：对于大量数据传输，使用流式 RPC 可以提高性能
5. **合理配置拦截器**：拦截器会影响性能，只添加必要的拦截器
6. **生产环境禁用反射**：反射服务会暴露服务信息，生产环境建议禁用

## 常见问题

### 1. 如何同时运行 HTTP 和 gRPC 服务？

```go
// 创建 Gin HTTP 服务器
ginServer := server.NewGinServer(logger, server.WithGinPort(8080))

// 创建 gRPC 服务器
grpcServer, _ := server.NewGRPCServer(logger, server.WithGRPCPort(9090))

// 分别在 goroutine 中启动
go ginServer.Run()
go grpcServer.Run()
```

### 2. 如何实现 gRPC-Gateway（HTTP 转 gRPC）？

buf.gen.yaml 已经配置了 gRPC-Gateway 插件，生成的代码会包含 HTTP 网关支持。

### 3. 如何处理 context 超时？

```go
func (s *MyServiceImpl) GetData(ctx context.Context, req *pb.GetDataRequest) (*pb.GetDataResponse, error) {
    // 检查 context 是否已取消
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    // 业务逻辑...
}
```

## 参考资料

- [gRPC 官方文档](https://grpc.io/docs/)
- [buf 官方文档](https://buf.build/docs/)
- [Protocol Buffers 指南](https://protobuf.dev/)
- [gRPC Health Checking Protocol](https://github.com/grpc/grpc/blob/master/doc/health-checking.md)
