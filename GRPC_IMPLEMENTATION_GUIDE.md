# gRPC实施指南

**版本**: 1.0
**创建日期**: 2025-11-06
**适用范围**: k8s-agent项目的所有gRPC服务

---

## 目录

1. [概述](#概述)
2. [架构设计](#架构设计)
3. [实施步骤](#实施步骤)
4. [代码示例](#代码示例)
5. [最佳实践](#最佳实践)
6. [故障排查](#故障排查)
7. [API文档](#api文档)

---

## 概述

本项目采用gRPC作为服务间通信协议，同时使用gRPC-Gateway提供HTTP RESTful接口。本指南介绍如何为服务添加gRPC支持。

### 核心特性

- **统一Handler模式**: HTTP和gRPC共享业务逻辑
- **标准化初始化**: 使用统一的初始化器模式
- **自动HTTP转换**: gRPC-Gateway自动生成RESTful API
- **依赖注入**: 使用Google Wire管理依赖
- **生命周期管理**: Bootstrap框架统一管理

### 已实现的服务

| 服务 | gRPC端口 | HTTP端口 | 状态 |
|------|----------|----------|------|
| reasoning | 50051 | 8081 | ✅ 完整实现 |
| orchestrator | 50052 | 8082 | ✅ 完整实现 |
| agent-manager | 50053 | 8083 | ✅ 完整实现 |
| cluster | 50054 | 8084 | ✅ 框架完成 |
| monitor | 50055 | 8085 | ✅ 框架完成 |
| auth | 50056 | 8090 | ⏳ 框架就绪 |

---

## 架构设计

### 整体架构

```
┌─────────────────────────────────────────────────────────┐
│                     Client Layer                         │
├─────────────────────────────────────────────────────────┤
│  HTTP Client    │  gRPC Client   │   Browser (REST)     │
└────────┬─────────┴────────┬──────┴──────────┬───────────┘
         │                  │                  │
         ▼                  ▼                  ▼
┌─────────────────────────────────────────────────────────┐
│                   Gateway / Proxy                        │
│  ┌──────────────┐  ┌────────────┐  ┌─────────────────┐ │
│  │ gRPC-Gateway │  │ gRPC Proxy │  │   HTTP Router   │ │
│  └──────┬───────┘  └─────┬──────┘  └────────┬────────┘ │
└─────────┼────────────────┼──────────────────┼──────────┘
          │                │                  │
          └────────────────┼──────────────────┘
                           ▼
┌─────────────────────────────────────────────────────────┐
│                   Service Layer                          │
│  ┌─────────────────────────────────────────────────┐   │
│  │         Unified Handler (OneX Pattern)          │   │
│  │  ┌──────────────────┐  ┌──────────────────┐    │   │
│  │  │ gRPC Interface   │  │  HTTP Interface  │    │   │
│  │  │ (Proto Server)   │  │  (via Gateway)   │    │   │
│  │  └────────┬─────────┘  └────────┬─────────┘    │   │
│  │           └──────────┬───────────┘              │   │
│  │                      ▼                          │   │
│  │           Shared Business Logic                 │   │
│  └─────────────────────────────────────────────────┘   │
│                          │                              │
│                          ▼                              │
│  ┌──────────────────────────────────────────────┐     │
│  │            Service / Repository              │     │
│  └──────────────────────────────────────────────┘     │
└──────────────────────────┬──────────────────────────────┘
                           ▼
                    ┌─────────────┐
                    │  Database   │
                    └─────────────┘
```

### OneX架构模式

核心思想：一个Handler同时实现gRPC和HTTP接口

```go
type ServiceHandler struct {
    // gRPC接口
    servicev1.UnimplementedServiceServer

    // 共享的业务逻辑
    service *BusinessService
    logger  core.Logger
}

// gRPC方法
func (h *ServiceHandler) GetData(ctx context.Context, req *servicev1.GetDataRequest) (*servicev1.Data, error) {
    return h.service.GetData(ctx, req.Id)
}

// HTTP通过gRPC-Gateway自动调用gRPC方法
```

---

## 实施步骤

### 步骤1: 创建Proto文件

在 `pkg/api/<service>/v1/` 创建 `.proto` 文件：

```protobuf
syntax = "proto3";

package myservice.v1;

import "google/api/annotations.proto";
import "common/pagination/v1/pagination.proto";

option go_package = "github.com/kart-io/k8s-agent/pkg/api/myservice/v1;myservicev1";

service MyService {
  // 获取数据
  rpc GetData(GetDataRequest) returns (DataResponse) {
    option (google.api.http) = {
      get: "/api/v1/data/{id}"
    };
  }

  // 列出数据
  rpc ListData(ListDataRequest) returns (ListDataResponse) {
    option (google.api.http) = {
      get: "/api/v1/data"
    };
  }
}

message GetDataRequest {
  string id = 1;
}

message DataResponse {
  string id = 1;
  string name = 2;
}

message ListDataRequest {
  common.pagination.v1.PaginationRequest pagination = 1;
}

message ListDataResponse {
  repeated DataResponse data = 1;
  common.pagination.v1.PaginationMetadata pagination = 2;
}
```

**注意事项**:
- 使用 `google.api.http` 注解定义HTTP映射
- 使用 `common.pagination.v1.PaginationMetadata` （不是PaginationResponse）
- package名称格式: `<service>.v1`
- go_package路径要正确

### 步骤2: 生成代码

```bash
cd /path/to/k8s-agent
buf generate
```

会生成以下文件：
- `<service>.pb.go` - 消息定义
- `<service>_grpc.pb.go` - gRPC服务接口
- `<service>.pb.gw.go` - gRPC-Gateway代码
- `<service>_http.pb.go` - HTTP处理代码

### 步骤3: 添加GRPCOptions配置

在 `cmd/<service>/app/options/options.go`:

```go
type ServerOptions struct {
    Server   *commonoptions.ServerOptions   `json:"server" mapstructure:"server"`
    GRPC     *commonoptions.GRPCOptions     `json:"grpc" mapstructure:"grpc"`  // 添加此行
    // ... 其他配置
}

func NewServerOptions() *ServerOptions {
    return &ServerOptions{
        Server:   commonoptions.NewServerOptions(),
        GRPC:     commonoptions.NewGRPCOptions(),  // 添加此行
        // ... 其他初始化
    }
}
```

### 步骤4: 创建gRPC初始化器

在 `internal/<service>/initializers/grpc.go`:

```go
package initializers

import (
    "context"
    "google.golang.org/grpc"

    "github.com/kart-io/k8s-agent/cmd/<service>/app/options"
    servicev1 "github.com/kart-io/k8s-agent/pkg/api/<service>/v1"
    "github.com/kart-io/k8s-agent/pkg/bootstrap"
    commoninitializers "github.com/kart-io/k8s-agent/pkg/initializers"
    commonserver "github.com/kart-io/k8s-agent/common/server"
    "github.com/kart-io/logger/core"
)

// GRPCServerInitializer gRPC服务器初始化器
type GRPCServerInitializer struct {
    standardInit *commoninitializers.GRPCServerInitializer
    opts         *options.ServerOptions
    logger       core.Logger

    // 依赖的初始化器
    dbInit *DatabaseInitializer

    // gRPC服务实现
    myService *MyGRPCService
}

func NewGRPCServerInitializer(
    opts *options.ServerOptions,
    logger core.Logger,
    dbInit *DatabaseInitializer,
) *GRPCServerInitializer {
    return &GRPCServerInitializer{
        opts:   opts,
        logger: logger,
        dbInit: dbInit,
    }
}

func (g *GRPCServerInitializer) Name() string {
    return "<service>-grpc-server"
}

func (g *GRPCServerInitializer) Priority() int {
    return bootstrap.PriorityGRPC
}

func (g *GRPCServerInitializer) Initialize(ctx context.Context) error {
    if !g.opts.GRPC.Enable {
        g.logger.Infow("gRPC server is disabled")
        return nil
    }

    g.logger.Infow("Initializing gRPC server",
        "host", g.opts.GRPC.Host,
        "port", g.opts.GRPC.Port,
    )

    // 创建Service实现
    g.myService = NewMyGRPCService(g.logger)

    serverConfig := &commoninitializers.GRPCServerConfig{
        Name:     g.Name(),
        Priority: g.Priority(),
        Config:   g.opts.GRPC,
        ServiceRegister: func(s *grpc.Server) error {
            // 注册服务
            servicev1.RegisterMyServiceServer(s, g.myService)
            g.logger.Info("Services registered")
            return nil
        },
    }

    g.standardInit = commoninitializers.NewGRPCServerInitializer(serverConfig, g.logger)
    return g.standardInit.Initialize(ctx)
}

func (g *GRPCServerInitializer) Close(ctx context.Context) error {
    if g.standardInit == nil {
        return nil
    }
    return g.standardInit.Close(ctx)
}

func (g *GRPCServerInitializer) GetServer() commonserver.Server {
    if g.standardInit == nil {
        return nil
    }
    return g.standardInit.GetServer()
}

// MyGRPCService 实现 servicev1.MyServiceServer
type MyGRPCService struct {
    servicev1.UnimplementedMyServiceServer
    logger core.Logger
}

func NewMyGRPCService(logger core.Logger) *MyGRPCService {
    return &MyGRPCService{
        logger: logger.With("component", "my-grpc-service"),
    }
}

func (s *MyGRPCService) GetData(ctx context.Context, req *servicev1.GetDataRequest) (*servicev1.DataResponse, error) {
    s.logger.Infow("GetData RPC called", "id", req.Id)

    // 实现业务逻辑
    return &servicev1.DataResponse{
        Id:   req.Id,
        Name: "Example",
    }, nil
}

func (s *MyGRPCService) ListData(ctx context.Context, req *servicev1.ListDataRequest) (*servicev1.ListDataResponse, error) {
    s.logger.Infow("ListData RPC called")

    // 实现业务逻辑
    return &servicev1.ListDataResponse{
        Data: []*servicev1.DataResponse{},
    }, nil
}
```

### 步骤5: 更新Wire配置

在 `cmd/<service>/app/wire.go`:

```go
var InitializerSet = wire.NewSet(
    ProvideLogger,
    initializers.NewDatabaseInitializer,
    initializers.NewHTTPServerInitializer,
    initializers.NewGRPCServerInitializer,  // 添加此行
)
```

在 `cmd/<service>/app/components.go`:

```go
type ServiceComponents struct {
    DB     *initializers.DatabaseInitializer
    HTTP   *initializers.HTTPServerInitializer
    GRPC   *initializers.GRPCServerInitializer  // 添加此行
    Health *pkginitializers.HealthCheckInitializer
}

func NewServiceComponents(
    db *initializers.DatabaseInitializer,
    http *initializers.HTTPServerInitializer,
    grpc *initializers.GRPCServerInitializer,  // 添加此行
    health *pkginitializers.HealthCheckInitializer,
) *ServiceComponents {
    return &ServiceComponents{
        DB:     db,
        HTTP:   http,
        GRPC:   grpc,  // 添加此行
        Health: health,
    }
}
```

### 步骤6: 生成Wire代码

```bash
cd cmd/<service>
wire
```

或者手动创建 `wire_gen.go`（参考cluster/monitor的实现）

### 步骤7: 更新app.go注册组件

在 `cmd/<service>/app/app.go`:

```go
func (a *App) registerComponents(bs *bootstrap.Bootstrap) error {
    components, err := InitializeComponents(a.opts)
    if err != nil {
        return err
    }

    bs.Register(components.DB)
    bs.Register(components.HTTP)
    bs.Register(components.GRPC)  // 添加此行
    bs.Register(components.Health)

    return nil
}
```

### 步骤8: 编译测试

```bash
cd /path/to/k8s-agent
go build -o _output/bin/<service> ./cmd/<service>/main.go
```

---

## 代码示例

### 完整的gRPC Service实现

```go
// ReasoningHandler 实现reasoning服务（参考项目中的实际实现）
type ReasoningHandler struct {
    reasoningv1.UnimplementedReasoningServiceServer
    analyzer *analyzer.RootCauseAnalyzer
    logger   core.Logger
}

func NewReasoningHandler(
    analyzer *analyzer.RootCauseAnalyzer,
    logger core.Logger,
) *ReasoningHandler {
    return &ReasoningHandler{
        analyzer: analyzer,
        logger:   logger.With("component", "reasoning-handler"),
    }
}

// AnalyzeRootCause 实现根因分析
func (h *ReasoningHandler) AnalyzeRootCause(
    ctx context.Context,
    req *reasoningv1.RootCauseAnalysisRequest,
) (*reasoningv1.RootCauseAnalysisResponse, error) {
    h.logger.Infow("Analyzing root cause",
        "event_id", req.EventId,
    )

    // 调用业务逻辑
    result, err := h.analyzer.Analyze(ctx, req)
    if err != nil {
        h.logger.Errorw("Analysis failed", "error", err)
        return nil, err
    }

    return result, nil
}
```

### 流式RPC示例

```go
// ReportMetrics 流式接收指标数据
func (s *MetricsCollectorGRPCService) ReportMetrics(
    stream monitorv1.MetricsCollectorService_ReportMetricsServer,
) error {
    s.logger.Infow("ReportMetrics stream started")

    receivedCount := 0
    for {
        metricData, err := stream.Recv()
        if err == io.EOF {
            // 客户端关闭流
            break
        }
        if err != nil {
            return err
        }

        // 处理接收到的数据
        s.logger.Debugw("Received metric",
            "agent_id", metricData.AgentId,
            "metric", metricData.MetricName,
        )

        // TODO: 存储指标数据
        receivedCount++
    }

    // 发送响应并关闭流
    return stream.SendAndClose(&monitorv1.ReportMetricsResponse{
        ReceivedCount: int32(receivedCount),
        Message:       "Metrics received",
    })
}
```

### 监听资源变化（服务端流式）

```go
// WatchResources 监听K8s资源变化
func (s *K8sResourceGRPCService) WatchResources(
    req *clusterv1.WatchResourcesRequest,
    stream clusterv1.K8SResourceService_WatchResourcesServer,
) error {
    s.logger.Infow("Watch started",
        "cluster", req.ClusterId,
        "type", req.ResourceType,
    )

    // 创建K8s watch
    watcher, err := s.createK8sWatcher(req)
    if err != nil {
        return err
    }
    defer watcher.Stop()

    // 持续发送事件
    for event := range watcher.ResultChan() {
        resourceEvent := &clusterv1.ResourceEvent{
            Type:     convertEventType(event.Type),
            Resource: convertToProtoResource(event.Object),
        }

        if err := stream.Send(resourceEvent); err != nil {
            return err
        }
    }

    return nil
}
```

---

## 最佳实践

### 1. Proto文件设计

**DO ✅**:
- 使用明确的命名: `GetUser`, `ListUsers`, `CreateUser`
- 包含分页支持: 使用 `PaginationRequest` 和 `PaginationMetadata`
- 添加HTTP注解: 每个RPC都应有 `google.api.http` 注解
- 使用版本化: package名包含版本 `myservice.v1`

**DON'T ❌**:
- 不要使用模糊的命名: `Get`, `Do`, `Process`
- 不要忘记timestamp字段: 使用 `google.protobuf.Timestamp`
- 不要硬编码枚举从1开始: 0值应为 `UNSPECIFIED`

### 2. 错误处理

使用gRPC标准错误码:

```go
import "google.golang.org/grpc/codes"
import "google.golang.org/grpc/status"

func (s *Service) GetData(ctx context.Context, req *v1.Request) (*v1.Response, error) {
    data, err := s.repo.Find(req.Id)
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            return nil, status.Error(codes.NotFound, "data not found")
        }
        return nil, status.Error(codes.Internal, "database error")
    }
    return data, nil
}
```

常用错误码:
- `codes.NotFound` - 资源不存在
- `codes.InvalidArgument` - 参数错误
- `codes.Unauthenticated` - 未认证
- `codes.PermissionDenied` - 无权限
- `codes.Internal` - 内部错误

### 3. 日志记录

```go
func (s *Service) Method(ctx context.Context, req *v1.Request) (*v1.Response, error) {
    // 记录请求
    s.logger.Infow("Method called",
        "request_id", extractRequestID(ctx),
        "param1", req.Param1,
    )

    // 业务逻辑
    result, err := s.doSomething(req)
    if err != nil {
        s.logger.Errorw("Method failed",
            "error", err,
            "param1", req.Param1,
        )
        return nil, err
    }

    // 记录成功
    s.logger.Infow("Method completed",
        "result_count", len(result),
    )

    return result, nil
}
```

### 4. 上下文传递

```go
// 传递超时
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

response, err := client.Method(ctx, request)

// 传递metadata
md := metadata.New(map[string]string{
    "request-id": "12345",
    "user-id":    "user-1",
})
ctx := metadata.NewOutgoingContext(context.Background(), md)
```

### 5. 性能优化

```go
// 使用连接池
conn, err := grpc.Dial(
    address,
    grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
    grpc.WithKeepaliveParams(keepalive.ClientParameters{
        Time:                10 * time.Second,
        Timeout:             3 * time.Second,
        PermitWithoutStream: true,
    }),
)

// 使用拦截器
opts := []grpc.ServerOption{
    grpc.UnaryInterceptor(loggingInterceptor),
    grpc.StreamInterceptor(streamLoggingInterceptor),
}
```

---

## 故障排查

### 常见问题

#### 1. 编译错误: "undefined: PaginationResponse"

**原因**: 应该使用 `PaginationMetadata` 而不是 `PaginationResponse`

**解决**:
```protobuf
// 错误
message ListResponse {
  repeated Item items = 1;
  common.pagination.v1.PaginationResponse pagination = 2;
}

// 正确
message ListResponse {
  repeated Item items = 1;
  common.pagination.v1.PaginationMetadata pagination = 2;
}
```

#### 2. 编译错误: "cannot use store (type interface{}) as type *Storage"

**原因**: 需要类型断言

**解决**:
```go
// 错误
store := g.dbInit.Store()
service := NewService(store)

// 正确
store := g.dbInit.Store().(*storage.MySQLStorage)
service := NewService(store)
```

#### 3. Wire生成失败

**解决步骤**:
1. 删除旧的 `wire_gen.go`
2. 确保 `wire.go` 中的 `InitializerSet` 包含所有初始化器
3. 确保 `components.go` 中的构造函数参数顺序正确
4. 运行 `cd cmd/<service> && wire`

#### 4. gRPC服务无法启动

**检查清单**:
- [ ] `GRPC.Enable` 是否设置为 `true`
- [ ] 端口是否被占用
- [ ] 初始化器是否注册到Bootstrap
- [ ] Wire依赖是否正确生成

**调试方法**:
```bash
# 检查端口
netstat -tuln | grep 50051

# 查看日志
./service --config config.yaml

# 测试连接
grpcurl -plaintext localhost:50051 list
```

---

## API文档

### 测试gRPC服务

#### 使用grpcurl

```bash
# 列出服务
grpcurl -plaintext localhost:50051 list

# 列出方法
grpcurl -plaintext localhost:50051 list myservice.v1.MyService

# 调用方法
grpcurl -plaintext \
  -d '{"id": "123"}' \
  localhost:50051 \
  myservice.v1.MyService/GetData

# 使用JSON文件
grpcurl -plaintext \
  -d @request.json \
  localhost:50051 \
  myservice.v1.MyService/ListData
```

#### 使用HTTP（通过gRPC-Gateway）

```bash
# GET请求
curl http://localhost:8081/api/v1/data/123

# POST请求
curl -X POST http://localhost:8081/api/v1/data \
  -H "Content-Type: application/json" \
  -d '{"name": "test"}'

# 分页查询
curl "http://localhost:8081/api/v1/data?page=1&page_size=10"
```

### 配置示例

在 `configs/<service>/config.yaml`:

```yaml
grpc:
  enable: true
  host: "0.0.0.0"
  port: 50051
  max_connection_idle: "5m"
  max_connection_age: "2h"
  timeout: "30s"
  max_recv_msg_size: 4194304  # 4MB
  max_send_msg_size: 4194304  # 4MB
```

---

## 附录

### A. 项目中的gRPC服务清单

| 服务 | Proto文件 | 端口 | 状态 |
|------|-----------|------|------|
| reasoning | pkg/api/reasoning/v1/analysis.proto | 50051 | ✅ |
| orchestrator | pkg/api/orchestrator/v1/workflow.proto | 50052 | ✅ |
| agent-manager | pkg/api/agent/v1/agent.proto | 50053 | ✅ |
| cluster | pkg/api/cluster/v1/cluster.proto | 50054 | ✅ |
| monitor | pkg/api/monitor/v1/monitor.proto | 50055 | ✅ |
| auth | pkg/api/auth/v1/auth.proto | 50056 | ⏳ |

### B. 相关文档

- [gRPC官方文档](https://grpc.io/docs/)
- [gRPC-Gateway文档](https://grpc-ecosystem.github.io/grpc-gateway/)
- [Protocol Buffers指南](https://developers.google.com/protocol-buffers)
- [Google Wire文档](https://github.com/google/wire)

### C. 快速参考

```bash
# 生成Proto代码
buf generate

# 运行Wire
cd cmd/<service> && wire

# 编译服务
go build -o _output/bin/<service> ./cmd/<service>/main.go

# 测试gRPC
grpcurl -plaintext localhost:50051 list

# 查看日志
tail -f logs/service.log
```

---

**文档版本**: 1.0
**最后更新**: 2025-11-06
**维护者**: k8s-agent团队

