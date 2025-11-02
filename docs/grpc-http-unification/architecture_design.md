# gRPC-HTTP统一Handler架构设计

## 系统架构概览

```
┌─────────────────────────────────────────────────────────────────┐
│                         客户端层                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │ Web Browser  │  │  curl/httpie │  │ gRPC Client  │          │
│  │   (HTTP)     │  │    (HTTP)    │  │   (gRPC)     │          │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘          │
└─────────┼──────────────────┼──────────────────┼─────────────────┘
          │                  │                  │
          │ HTTP/JSON        │ HTTP/JSON        │ gRPC/Protobuf
          │ :8092/:8082      │ :8092/:8082      │ :9092/:9093
          ▼                  ▼                  ▼
┌─────────────────────────────────────────────────────────────────┐
│                       服务传输层                                  │
│                                                                  │
│  ┌──────────────────────┐        ┌──────────────────────┐      │
│  │  Orchestrator        │        │  Reasoning           │      │
│  │                      │        │                      │      │
│  │  ┌────────────────┐  │        │  ┌────────────────┐ │      │
│  │  │ HTTP Server    │  │        │  │ HTTP Server    │ │      │
│  │  │ (gRPC-Gateway) │  │        │  │ (gRPC-Gateway) │ │      │
│  │  │ Port: 8092     │  │        │  │ Port: 8082     │ │      │
│  │  └────────┬───────┘  │        │  └────────┬───────┘ │      │
│  │           │          │        │           │         │      │
│  │  ┌────────▼───────┐  │        │  ┌────────▼───────┐ │      │
│  │  │ gRPC Server    │  │        │  │ gRPC Server    │ │      │
│  │  │ Port: 9092     │  │        │  │ Port: 9093     │ │      │
│  │  └────────┬───────┘  │        │  └────────┬───────┘ │      │
│  └───────────┼──────────┘        └───────────┼─────────┘      │
└──────────────┼───────────────────────────────┼────────────────┘
               │                               │
               │ 共享Service实例                │ 共享Service实例
               ▼                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                       业务逻辑层                                  │
│                                                                  │
│  ┌──────────────────────┐        ┌──────────────────────┐      │
│  │ WorkflowService      │        │ ReasoningService     │      │
│  │                      │        │                      │      │
│  │  • CreateWorkflow    │        │  • RootCauseAnalysis │      │
│  │  • GetWorkflow       │        │  • SaveCase          │      │
│  │  • ListWorkflows     │        │                      │      │
│  │  • ExecuteWorkflow   │        │                      │      │
│  │  • GetExecutionStatus│        │                      │      │
│  └──────────┬───────────┘        └──────────┬───────────┘      │
└─────────────┼──────────────────────────────┼────────────────────┘
              │                              │
              ▼                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                       数据/依赖层                                 │
│                                                                  │
│  ┌─────────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │ PostgreSQL  │  │  Redis   │  │   NATS   │  │ LLM APIs │   │
│  │  (工作流)    │  │  (缓存)   │  │  (消息)   │  │ (AI分析)  │   │
│  └─────────────┘  └──────────┘  └──────────┘  └──────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

## 核心组件详解

### 1. 服务初始化流程

#### Orchestrator Service启动顺序
```
Priority  Component           说明
───────────────────────────────────────────────────────
 300      Database           PostgreSQL连接初始化
 400      Redis              Redis缓存初始化
 500      NATS               消息队列初始化
 550      Workflow Engine    工作流引擎初始化
 600      Strategy Manager   策略管理器初始化
 650      Subscriber         事件订阅器初始化
 700      gRPC Server        gRPC服务启动
          └─> 创建 WorkflowServiceServer (共享实例)
 800      HTTP Server        HTTP服务启动
          └─> 获取 WorkflowServiceServer (同一实例)
          └─> 注册 gRPC-Gateway handlers
 900      Health Check       健康检查服务启动
```

#### Reasoning Service启动顺序
```
Priority  Component           说明
───────────────────────────────────────────────────────
 400      LLM Clients        OpenAI/Gemini/DeepSeek客户端初始化
 450      gRPC Server        gRPC服务启动
          └─> 创建 RootCauseAnalyzer
          └─> 创建 ReasoningServiceServer (共享实例)
 500      HTTP Server        HTTP服务启动
          └─> 获取 ReasoningServiceServer (同一实例)
          └─> 注册 gRPC-Gateway handlers
 600      Health Check       健康检查服务启动
```

### 2. 请求处理流程

#### HTTP请求流程
```
1. HTTP Request (JSON)
   │
   ├─> POST /v1/workflows
   │   Content-Type: application/json
   │   Body: {"name": "...", "steps": [...]}
   │
2. gRPC-Gateway (runtime.ServeMux)
   │
   ├─> 解析HTTP请求
   ├─> 转换JSON → Protobuf
   ├─> 提取路径参数 {workflow_id}
   │
3. 调用gRPC Service方法
   │
   ├─> WorkflowServiceServer.CreateWorkflow(ctx, req)
   │
4. 业务逻辑处理
   │
   ├─> 验证请求
   ├─> 保存到数据库
   ├─> 返回响应
   │
5. gRPC-Gateway响应转换
   │
   ├─> Protobuf → JSON
   ├─> 设置HTTP状态码
   │
6. HTTP Response (JSON)
   │
   └─> 200 OK
       Content-Type: application/json
       Body: {"workflow_id": "...", "status": "CREATED"}
```

#### gRPC请求流程
```
1. gRPC Request (Protobuf)
   │
   ├─> orchestrator.v1.WorkflowService/CreateWorkflow
   │   Message: CreateWorkflowRequest (binary)
   │
2. gRPC Server
   │
   ├─> 反序列化Protobuf
   │
3. 调用gRPC Service方法 (与HTTP相同!)
   │
   ├─> WorkflowServiceServer.CreateWorkflow(ctx, req)
   │
4. 业务逻辑处理 (与HTTP相同!)
   │
   ├─> 验证请求
   ├─> 保存到数据库
   ├─> 返回响应
   │
5. gRPC Server响应
   │
   ├─> 序列化Protobuf
   │
6. gRPC Response (Protobuf)
   │
   └─> Message: CreateWorkflowResponse (binary)
```

**关键点**: HTTP和gRPC从步骤3开始使用**完全相同的代码**！

### 3. 代码组织结构

```
internal/orchestrator/
├── service/                    # 业务逻辑层 (协议无关)
│   └── workflow_service.go     # 实现WorkflowServiceServer接口
│                               # 同时被gRPC和HTTP使用
├── grpc/
│   └── server.go               # gRPC服务器配置
│                               # 接收共享的WorkflowServiceServer
├── initializers/
│   ├── grpc.go                 # gRPC初始化器
│   │                           # • 创建WorkflowServiceServer实例
│   │                           # • 提供GetWorkflowService()方法
│   └── http.go                 # HTTP初始化器
│                               # • 获取共享的WorkflowServiceServer
│                               # • 注册gRPC-Gateway handlers
└── ...

pkg/api/orchestrator/v1/
├── workflow.proto              # Proto定义 + HTTP注解
├── workflow.pb.go              # 生成的Protobuf代码
└── workflow.pb.gw.go           # 生成的gRPC-Gateway代码
                                # 包含HTTP路由和转换逻辑
```

### 4. Proto定义示例

```protobuf
// workflow.proto
syntax = "proto3";

package orchestrator.v1;

import "google/api/annotations.proto";

service WorkflowService {
  // HTTP注解定义RESTful路由
  rpc CreateWorkflow(CreateWorkflowRequest) returns (CreateWorkflowResponse) {
    option (google.api.http) = {
      post: "/v1/workflows"    // HTTP POST路由
      body: "*"                // 请求体映射
    };
  }

  rpc GetWorkflow(GetWorkflowRequest) returns (GetWorkflowResponse) {
    option (google.api.http) = {
      get: "/v1/workflows/{workflow_id}"  // 路径参数
    };
  }
}

message CreateWorkflowRequest {
  string name = 1;
  string description = 2;
  repeated WorkflowStep steps = 3;
}

message CreateWorkflowResponse {
  string workflow_id = 1;
  string status = 2;
}
```

### 5. Service实现示例

```go
// internal/orchestrator/service/workflow_service.go
package service

import (
    "context"
    orchestratorv1 "github.com/kart-io/k8s-agent/pkg/api/orchestrator/v1"
)

// WorkflowServiceServer实现gRPC接口
// 同时被HTTP和gRPC使用！
type WorkflowServiceServer struct {
    orchestratorv1.UnimplementedWorkflowServiceServer
    engine *workflow.Engine
    store  *storage.PostgresStore
    logger core.Logger
}

// CreateWorkflow 创建工作流
// HTTP和gRPC都调用这个方法！
func (s *WorkflowServiceServer) CreateWorkflow(
    ctx context.Context,
    req *orchestratorv1.CreateWorkflowRequest,
) (*orchestratorv1.CreateWorkflowResponse, error) {
    // 业务逻辑
    workflow := &types.Workflow{
        Name:        req.Name,
        Description: req.Description,
        Steps:       convertSteps(req.Steps),
    }

    if err := s.store.SaveWorkflow(ctx, workflow); err != nil {
        return nil, err
    }

    return &orchestratorv1.CreateWorkflowResponse{
        WorkflowId: workflow.ID,
        Status:     "CREATED",
    }, nil
}
```

### 6. HTTP Server初始化

```go
// internal/orchestrator/initializers/http.go
package initializers

func (i *HTTPServerInitializer) Initialize(ctx context.Context) error {
    // 1. 创建gRPC-Gateway mux
    i.mux = runtime.NewServeMux()

    // 2. 获取共享的service实例 (与gRPC相同!)
    workflowService := i.grpcInit.GetWorkflowService()

    // 3. 注册HTTP handlers (自动生成的代码)
    if err := orchestratorv1.RegisterWorkflowServiceHandlerServer(
        ctx,
        i.mux,
        workflowService,  // 传入共享的service
    ); err != nil {
        return err
    }

    // 4. 创建HTTP server
    i.server = &http.Server{
        Addr:    fmt.Sprintf(":%d", port),
        Handler: i.mux,
    }

    // 5. 启动服务器
    go i.server.ListenAndServe()

    return nil
}
```

### 7. 关键技术点

#### 7.1 共享Service模式
```
┌─────────────────────────────────┐
│  gRPC Initializer               │
│  ┌───────────────────────────┐  │
│  │ 创建 ServiceServer实例     │  │
│  │ service = NewService(...)  │  │
│  └───────────┬───────────────┘  │
│              │                   │
│              │ 存储到字段         │
│              ▼                   │
│  ┌───────────────────────────┐  │
│  │ GetService() *Service     │◄─┼── HTTP Initializer调用
│  │   return i.service        │  │
│  └───────────────────────────┘  │
└─────────────────────────────────┘
```

#### 7.2 依赖注入
```go
// app.go
func (a *App) RegisterComponents(bs *bootstrap.Bootstrap) error {
    // 1. gRPC先初始化 (优先级700)
    a.grpcInit = initializers.NewGRPCServerInitializer(...)
    bs.Register(a.grpcInit)

    // 2. HTTP后初始化 (优先级800)
    a.httpInit = initializers.NewHTTPServerInitializer(
        opts,
        logger,
        a.grpcInit,  // 传入gRPC initializer
    )
    bs.Register(a.httpInit)

    return nil
}
```

#### 7.3 自动代码生成
```bash
# buf.gen.yaml配置
plugins:
  # 生成标准Protobuf代码
  - remote: buf.build/protocolbuffers/go
    out: pkg/api

  # 生成gRPC代码
  - remote: buf.build/grpc/go
    out: pkg/api

  # 生成gRPC-Gateway代码 (关键!)
  - remote: buf.build/grpc-ecosystem/gateway
    out: pkg/api
    opt:
      - paths=source_relative
      - generate_unbound_methods=true

# 执行生成
$ buf generate
```

## 技术优势

### 1. 零代码重复
- ✅ 业务逻辑只实现一次
- ✅ HTTP和gRPC共享相同代码
- ✅ 修改一处两端生效

### 2. 类型安全
- ✅ Proto定义作为单一数据源
- ✅ 编译时类型检查
- ✅ 自动生成客户端代码

### 3. 易于维护
- ✅ 清晰的分层架构
- ✅ 协议细节隐藏
- ✅ 业务逻辑与传输解耦

### 4. 灵活性
- ✅ 支持多种客户端
- ✅ 可独立扩展
- ✅ 向后兼容

### 5. 性能优化
- ✅ gRPC高性能二进制协议
- ✅ HTTP/2多路复用
- ✅ 流式传输支持

## 最佳实践

### 1. Service设计
```go
// ✅ 好的实践: service包独立于传输层
package service

type WorkflowServiceServer struct {
    // 只依赖业务组件，不依赖HTTP/gRPC
    engine *workflow.Engine
    store  storage.Store
}

// ❌ 避免: 混合HTTP和gRPC逻辑
package grpc

type WorkflowService struct {
    ginRouter *gin.Engine  // 不应该在这里!
}
```

### 2. 错误处理
```go
// ✅ 使用gRPC标准错误码
import "google.golang.org/grpc/status"

func (s *Service) CreateWorkflow(...) (*Response, error) {
    if req.Name == "" {
        return nil, status.Error(codes.InvalidArgument, "name is required")
    }

    if err := s.store.Save(workflow); err != nil {
        return nil, status.Error(codes.Internal, "failed to save")
    }

    return &Response{...}, nil
}
```

### 3. Proto设计
```protobuf
// ✅ 好的设计: RESTful路由
rpc GetWorkflow(GetWorkflowRequest) returns (GetWorkflowResponse) {
  option (google.api.http) = {
    get: "/v1/workflows/{workflow_id}"
  };
}

// ❌ 避免: 非RESTful路由
rpc GetWorkflow(GetWorkflowRequest) returns (GetWorkflowResponse) {
  option (google.api.http) = {
    post: "/v1/getWorkflow"  // 应该用GET
  };
}
```

### 4. 优先级设置
```
数据库/缓存 (300-400)
  ↓
业务组件 (500-600)
  ↓
gRPC Server (700)
  ↓
HTTP Server (800)
  ↓
健康检查 (900)
```

## 总结

这个架构实现了**真正的统一handler模式**:
- 一个service实现
- 两种协议支持
- 零代码重复
- 高性能 + 易用性

完美结合了gRPC的高性能和HTTP的易用性！

---

**文档版本**: v1.0
**更新时间**: 2025-11-01
**作者**: Claude Code
