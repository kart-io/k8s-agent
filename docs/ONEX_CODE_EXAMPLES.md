# OneX代码实例与最佳实践

**文档版本**: v1.0
**创建日期**: 2025-11-01
**配套文档**: ONEX_IMPLEMENTATION_GUIDE.md

本文档提供OneX项目中的实际代码示例，可直接用于k8s-agent项目实施。

---

## 目录

1. [分层架构完整示例](#分层架构完整示例)
2. [Wire依赖注入完整流程](#wire依赖注入完整流程)
3. [分布式追踪集成](#分布式追踪集成)
4. [错误处理最佳实践](#错误处理最佳实践)
5. [测试策略和Mock生成](#测试策略和mock生成)

---

## 分层架构完整示例

OneX使用清晰的3层架构：Handler → Biz → Store。以下是完整实现示例。

### Store层 (数据访问)

```go
// internal/orchestrator/store/store.go
package store

import (
    "context"
    "sync"

    "github.com/google/wire"
    "gorm.io/gorm"
)

//go:generate mockgen -destination mock_store.go -package store github.com/kart-io/k8s-agent/internal/orchestrator/store IStore,WorkflowStore

// ProviderSet for Wire
var ProviderSet = wire.NewSet(
    NewStore,
    wire.Bind(new(IStore), new(*datastore)),
)

// IStore 定义Store层接口
type IStore interface {
    DB(ctx context.Context) *gorm.DB
    TX(ctx context.Context, fn func(ctx context.Context) error) error
    Workflow() WorkflowStore
    Strategy() StrategyStore
    Execution() ExecutionStore
}

// WorkflowStore 工作流存储接口
type WorkflowStore interface {
    Create(ctx context.Context, workflow *model.WorkflowM) error
    Get(ctx context.Context, id string) (*model.WorkflowM, error)
    List(ctx context.Context, opts ListOptions) ([]*model.WorkflowM, int64, error)
    Update(ctx context.Context, workflow *model.WorkflowM) error
    Delete(ctx context.Context, id string) error
}

// 事务key
type transactionKey struct{}

// datastore 实现
type datastore struct {
    db *gorm.DB
}

var (
    once sync.Once
    S    *datastore
)

// NewStore 创建Store实例（单例）
func NewStore(db *gorm.DB) *datastore {
    once.Do(func() {
        S = &datastore{db: db}
    })
    return S
}

// DB 获取数据库实例（支持事务）
func (ds *datastore) DB(ctx context.Context) *gorm.DB {
    // 从context获取事务实例
    if tx, ok := ctx.Value(transactionKey{}).(*gorm.DB); ok {
        return tx
    }
    return ds.db
}

// TX 执行事务
func (ds *datastore) TX(ctx context.Context, fn func(ctx context.Context) error) error {
    return ds.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        ctx = context.WithValue(ctx, transactionKey{}, tx)
        return fn(ctx)
    })
}

// Workflow 返回WorkflowStore实现
func (ds *datastore) Workflow() WorkflowStore {
    return newWorkflowStore(ds)
}

// Strategy 返回StrategyStore实现
func (ds *datastore) Strategy() StrategyStore {
    return newStrategyStore(ds)
}

// Execution 返回ExecutionStore实现
func (ds *datastore) Execution() ExecutionStore {
    return newExecutionStore(ds)
}
```

```go
// internal/orchestrator/store/workflow.go
package store

import (
    "context"
    "github.com/kart-io/k8s-agent/internal/orchestrator/model"
)

type workflowStore struct {
    *datastore
}

func newWorkflowStore(ds *datastore) *workflowStore {
    return &workflowStore{datastore: ds}
}

// Create 创建工作流
func (s *workflowStore) Create(ctx context.Context, workflow *model.WorkflowM) error {
    return s.DB(ctx).Create(workflow).Error
}

// Get 获取工作流
func (s *workflowStore) Get(ctx context.Context, id string) (*model.WorkflowM, error) {
    var workflow model.WorkflowM
    if err := s.DB(ctx).Where("id = ?", id).First(&workflow).Error; err != nil {
        return nil, err
    }
    return &workflow, nil
}

// List 列出工作流
func (s *workflowStore) List(ctx context.Context, opts ListOptions) ([]*model.WorkflowM, int64, error) {
    var (
        workflows []*model.WorkflowM
        total     int64
    )

    db := s.DB(ctx).Model(&model.WorkflowM{})

    // 应用过滤条件
    if opts.ClusterID != "" {
        db = db.Where("cluster_id = ?", opts.ClusterID)
    }

    // 获取总数
    if err := db.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    // 分页查询
    if err := db.Offset(opts.Offset).Limit(opts.Limit).
        Order("created_at DESC").
        Find(&workflows).Error; err != nil {
        return nil, 0, err
    }

    return workflows, total, nil
}

// Update 更新工作流
func (s *workflowStore) Update(ctx context.Context, workflow *model.WorkflowM) error {
    return s.DB(ctx).Save(workflow).Error
}

// Delete 删除工作流
func (s *workflowStore) Delete(ctx context.Context, id string) error {
    return s.DB(ctx).Where("id = ?", id).Delete(&model.WorkflowM{}).Error
}
```

### Biz层 (业务逻辑)

```go
// internal/orchestrator/biz/biz.go
package biz

import (
    "github.com/google/wire"
    "github.com/kart-io/k8s-agent/internal/orchestrator/store"
)

//go:generate mockgen -destination mock_biz.go -package biz github.com/kart-io/k8s-agent/internal/orchestrator/biz IBiz

// ProviderSet for Wire
var ProviderSet = wire.NewSet(
    NewBiz,
    wire.Bind(new(IBiz), new(*biz)),
)

// IBiz 业务层接口
type IBiz interface {
    Workflow() WorkflowBiz
    Strategy() StrategyBiz
    Execution() ExecutionBiz
}

type biz struct {
    store store.IStore
}

var _ IBiz = (*biz)(nil)

// NewBiz 创建业务层实例
func NewBiz(store store.IStore) *biz {
    return &biz{store: store}
}

func (b *biz) Workflow() WorkflowBiz {
    return newWorkflowBiz(b.store)
}

func (b *biz) Strategy() StrategyBiz {
    return newStrategyBiz(b.store)
}

func (b *biz) Execution() ExecutionBiz {
    return newExecutionBiz(b.store)
}
```

```go
// internal/orchestrator/biz/workflow.go
package biz

import (
    "context"
    "time"

    "github.com/google/uuid"
    "github.com/kart-io/logger"

    "github.com/kart-io/k8s-agent/internal/orchestrator/model"
    "github.com/kart-io/k8s-agent/internal/orchestrator/store"
    "github.com/kart-io/k8s-agent/pkg/contextx"
)

// WorkflowBiz 工作流业务接口
type WorkflowBiz interface {
    Create(ctx context.Context, req *CreateWorkflowRequest) (*model.WorkflowM, error)
    Get(ctx context.Context, id string) (*model.WorkflowM, error)
    List(ctx context.Context, req *ListWorkflowRequest) ([]*model.WorkflowM, int64, error)
    Execute(ctx context.Context, id string) error
    Delete(ctx context.Context, id string) error
}

type workflowBiz struct {
    store  store.IStore
    logger logger.Logger
}

func newWorkflowBiz(store store.IStore) *workflowBiz {
    return &workflowBiz{
        store:  store,
        logger: logger.DefaultLogger(),
    }
}

// CreateWorkflowRequest 创建工作流请求
type CreateWorkflowRequest struct {
    Name        string                 `json:"name" binding:"required"`
    Description string                 `json:"description"`
    ClusterID   string                 `json:"cluster_id" binding:"required"`
    StrategyID  string                 `json:"strategy_id" binding:"required"`
    Parameters  map[string]interface{} `json:"parameters"`
}

// Create 创建工作流（业务逻辑）
func (b *workflowBiz) Create(ctx context.Context, req *CreateWorkflowRequest) (*model.WorkflowM, error) {
    // 从context提取追踪信息
    traceID := contextx.TraceID(ctx)
    userID := contextx.UserID(ctx)

    b.logger.Info("Creating workflow",
        logger.String("trace_id", traceID),
        logger.String("user_id", userID),
        logger.String("name", req.Name),
        logger.String("cluster_id", req.ClusterID),
    )

    // 验证策略是否存在
    strategy, err := b.store.Strategy().Get(ctx, req.StrategyID)
    if err != nil {
        return nil, err
    }

    // 构建工作流模型
    workflow := &model.WorkflowM{
        ID:          uuid.NewString(),
        Name:        req.Name,
        Description: req.Description,
        ClusterID:   req.ClusterID,
        StrategyID:  req.StrategyID,
        Parameters:  req.Parameters,
        Status:      "pending",
        CreatedBy:   userID,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }

    // 使用事务创建工作流和初始执行记录
    err = b.store.TX(ctx, func(ctx context.Context) error {
        // 创建工作流
        if err := b.store.Workflow().Create(ctx, workflow); err != nil {
            return err
        }

        // 创建初始执行记录
        execution := &model.ExecutionM{
            ID:         uuid.NewString(),
            WorkflowID: workflow.ID,
            Status:     "pending",
            StartTime:  time.Now(),
        }
        if err := b.store.Execution().Create(ctx, execution); err != nil {
            return err
        }

        return nil
    })

    if err != nil {
        b.logger.Error("Failed to create workflow",
            logger.Error(err),
            logger.String("trace_id", traceID),
        )
        return nil, err
    }

    b.logger.Info("Workflow created successfully",
        logger.String("trace_id", traceID),
        logger.String("workflow_id", workflow.ID),
    )

    return workflow, nil
}

// Get 获取工作流
func (b *workflowBiz) Get(ctx context.Context, id string) (*model.WorkflowM, error) {
    return b.store.Workflow().Get(ctx, id)
}

// ListWorkflowRequest 列出工作流请求
type ListWorkflowRequest struct {
    ClusterID string `form:"cluster_id"`
    Page      int    `form:"page" binding:"min=1"`
    PageSize  int    `form:"page_size" binding:"min=1,max=100"`
}

// List 列出工作流
func (b *workflowBiz) List(ctx context.Context, req *ListWorkflowRequest) ([]*model.WorkflowM, int64, error) {
    opts := store.ListOptions{
        ClusterID: req.ClusterID,
        Offset:    (req.Page - 1) * req.PageSize,
        Limit:     req.PageSize,
    }

    return b.store.Workflow().List(ctx, opts)
}

// Execute 执行工作流
func (b *workflowBiz) Execute(ctx context.Context, id string) error {
    // 获取工作流
    workflow, err := b.store.Workflow().Get(ctx, id)
    if err != nil {
        return err
    }

    // 更新状态为执行中
    workflow.Status = "running"
    workflow.UpdatedAt = time.Now()
    if err := b.store.Workflow().Update(ctx, workflow); err != nil {
        return err
    }

    // 异步执行工作流（实际项目中可能使用消息队列）
    go b.executeWorkflowAsync(context.Background(), workflow)

    return nil
}

func (b *workflowBiz) executeWorkflowAsync(ctx context.Context, workflow *model.WorkflowM) {
    // 实际的工作流执行逻辑
    // 这里只是示例
    b.logger.Info("Executing workflow asynchronously",
        logger.String("workflow_id", workflow.ID),
    )
}

// Delete 删除工作流
func (b *workflowBiz) Delete(ctx context.Context, id string) error {
    return b.store.Workflow().Delete(ctx, id)
}
```

### Handler层 (HTTP/gRPC)

```go
// internal/orchestrator/handler/handler.go
package handler

import (
    "github.com/google/wire"
    "github.com/kart-io/k8s-agent/internal/orchestrator/biz"
)

// ProviderSet for Wire
var ProviderSet = wire.NewSet(
    NewHandler,
)

// Handler HTTP处理器
type Handler struct {
    biz biz.IBiz
}

// NewHandler 创建Handler实例
func NewHandler(biz biz.IBiz) *Handler {
    return &Handler{biz: biz}
}
```

```go
// internal/orchestrator/handler/workflow.go
package handler

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/kart-io/k8s-agent/common/response"
    "github.com/kart-io/k8s-agent/internal/orchestrator/biz"
)

// CreateWorkflow 创建工作流处理器
func (h *Handler) CreateWorkflow(c *gin.Context) {
    var req biz.CreateWorkflowRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, err.Error())
        return
    }

    workflow, err := h.biz.Workflow().Create(c.Request.Context(), &req)
    if err != nil {
        response.InternalError(c, err)
        return
    }

    response.Success(c, workflow)
}

// GetWorkflow 获取工作流处理器
func (h *Handler) GetWorkflow(c *gin.Context) {
    id := c.Param("id")

    workflow, err := h.biz.Workflow().Get(c.Request.Context(), id)
    if err != nil {
        response.NotFound(c, "Workflow not found")
        return
    }

    response.Success(c, workflow)
}

// ListWorkflows 列出工作流处理器
func (h *Handler) ListWorkflows(c *gin.Context) {
    var req biz.ListWorkflowRequest
    if err := c.ShouldBindQuery(&req); err != nil {
        response.Error(c, http.StatusBadRequest, err.Error())
        return
    }

    workflows, total, err := h.biz.Workflow().List(c.Request.Context(), &req)
    if err != nil {
        response.InternalError(c, err)
        return
    }

    response.SuccessWithPagination(c, workflows, &response.PaginationInfo{
        Page:     req.Page,
        PageSize: req.PageSize,
        Total:    total,
    })
}

// ExecuteWorkflow 执行工作流处理器
func (h *Handler) ExecuteWorkflow(c *gin.Context) {
    id := c.Param("id")

    if err := h.biz.Workflow().Execute(c.Request.Context(), id); err != nil {
        response.InternalError(c, err)
        return
    }

    response.Success(c, gin.H{"message": "Workflow execution started"})
}

// DeleteWorkflow 删除工作流处理器
func (h *Handler) DeleteWorkflow(c *gin.Context) {
    id := c.Param("id")

    if err := h.biz.Workflow().Delete(c.Request.Context(), id); err != nil {
        response.InternalError(c, err)
        return
    }

    response.Success(c, gin.H{"message": "Workflow deleted"})
}
```

---

## Wire依赖注入完整流程

### 1. 定义ProviderSet

每个包定义自己的ProviderSet：

```go
// common/db/provider.go
package db

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
    NewMySQL,  // MySQL连接provider
)
```

```go
// internal/orchestrator/store/provider.go
package store

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
    NewStore,
    wire.Bind(new(IStore), new(*datastore)),
)
```

```go
// internal/orchestrator/biz/provider.go
package biz

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
    NewBiz,
    wire.Bind(new(IBiz), new(*biz)),
)
```

```go
// internal/orchestrator/handler/provider.go
package handler

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
    NewHandler,
)
```

### 2. 创建wire.go

```go
// cmd/orchestrator/app/wire.go
//go:build wireinject
// +build wireinject

package app

//go:generate go run github.com/google/wire/cmd/wire

import (
    "github.com/google/wire"

    "github.com/kart-io/k8s-agent/common/db"
    "github.com/kart-io/k8s-agent/internal/orchestrator/biz"
    "github.com/kart-io/k8s-agent/internal/orchestrator/handler"
    "github.com/kart-io/k8s-agent/internal/orchestrator/store"
)

// wireApp 声明依赖图
func wireApp(opts *ServerOptions) (*OrchestratorApp, func(), error) {
    wire.Build(
        // Database
        provideDBOptions,  // *db.MySQLOptions
        db.ProviderSet,    // *gorm.DB

        // Store
        store.ProviderSet, // store.IStore

        // Biz
        biz.ProviderSet,   // biz.IBiz

        // Handler
        handler.ProviderSet, // *handler.Handler

        // HTTP Server
        provideHTTPServer,  // *http.Server

        // Application
        NewOrchestratorApp,
    )

    return nil, nil, nil
}

// provideDBOptions 从ServerOptions提取DB选项
func provideDBOptions(opts *ServerOptions) *db.MySQLOptions {
    return opts.MySQLOptions
}

// provideHTTPServer 创建HTTP服务器
func provideHTTPServer(opts *ServerOptions, h *handler.Handler) *http.Server {
    router := gin.New()

    // 注册路由
    v1 := router.Group("/api/v1")
    {
        v1.POST("/workflows", h.CreateWorkflow)
        v1.GET("/workflows", h.ListWorkflows)
        v1.GET("/workflows/:id", h.GetWorkflow)
        v1.POST("/workflows/:id/execute", h.ExecuteWorkflow)
        v1.DELETE("/workflows/:id", h.DeleteWorkflow)
    }

    return &http.Server{
        Addr:    fmt.Sprintf(":%d", opts.ServerPort),
        Handler: router,
    }
}
```

### 3. 生成Wire代码

```bash
cd cmd/orchestrator/app
go generate ./...
```

生成的 `wire_gen.go`:

```go
// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package app

import (
    "github.com/kart-io/k8s-agent/common/db"
    "github.com/kart-io/k8s-agent/internal/orchestrator/biz"
    "github.com/kart-io/k8s-agent/internal/orchestrator/handler"
    "github.com/kart-io/k8s-agent/internal/orchestrator/store"
)

func wireApp(opts *ServerOptions) (*OrchestratorApp, func(), error) {
    // Database
    mysqlOptions := provideDBOptions(opts)
    gormDB, cleanup, err := db.NewMySQL(mysqlOptions)
    if err != nil {
        return nil, nil, err
    }

    // Store
    datastore := store.NewStore(gormDB)

    // Biz
    bizInstance := biz.NewBiz(datastore)

    // Handler
    handlerInstance := handler.NewHandler(bizInstance)

    // HTTP Server
    server := provideHTTPServer(opts, handlerInstance)

    // Application
    app := NewOrchestratorApp(server, datastore, bizInstance, handlerInstance)

    return app, func() {
        cleanup()
    }, nil
}
```

### 4. 使用Wire构建的应用

```go
// cmd/orchestrator/app/app.go
package app

import (
    "context"
    "net/http"

    "github.com/kart-io/k8s-agent/internal/orchestrator/biz"
    "github.com/kart-io/k8s-agent/internal/orchestrator/handler"
    "github.com/kart-io/k8s-agent/internal/orchestrator/store"
)

type OrchestratorApp struct {
    server  *http.Server
    store   store.IStore
    biz     biz.IBiz
    handler *handler.Handler
}

// NewOrchestratorApp Wire注入所有依赖
func NewOrchestratorApp(
    server *http.Server,
    store store.IStore,
    biz biz.IBiz,
    handler *handler.Handler,
) *OrchestratorApp {
    return &OrchestratorApp{
        server:  server,
        store:   store,
        biz:     biz,
        handler: handler,
    }
}

// Run 启动应用
func (a *OrchestratorApp) Run(ctx context.Context) error {
    return a.server.ListenAndServe()
}

// Shutdown 优雅关闭
func (a *OrchestratorApp) Shutdown(ctx context.Context) error {
    return a.server.Shutdown(ctx)
}
```

```go
// cmd/orchestrator/main.go
package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"

    "github.com/kart-io/k8s-agent/cmd/orchestrator/app"
)

func main() {
    // 加载配置
    opts, err := app.LoadOptions()
    if err != nil {
        panic(err)
    }

    // Wire构建应用（自动依赖注入）
    orchestratorApp, cleanup, err := app.wireApp(opts)
    if err != nil {
        panic(err)
    }
    defer cleanup()

    // 启动应用
    go func() {
        if err := orchestratorApp.Run(context.Background()); err != nil {
            panic(err)
        }
    }()

    // 等待中断信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    // 优雅关闭
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := orchestratorApp.Shutdown(ctx); err != nil {
        panic(err)
    }
}
```

---

## 分布式追踪集成

### OpenTelemetry中间件实现

```go
// common/middleware/tracing.go
package middleware

import (
    "github.com/gin-gonic/gin"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/trace"

    "github.com/kart-io/k8s-agent/pkg/contextx"
)

// Tracing OpenTelemetry追踪中间件
func Tracing(serviceName string) gin.HandlerFunc {
    tracer := otel.Tracer(serviceName)
    propagator := otel.GetTextMapPropagator()

    return func(c *gin.Context) {
        // 提取上游传播的trace context
        ctx := propagator.Extract(c.Request.Context(), propagation.HeaderCarrier(c.Request.Header))

        // 创建新的span
        spanName := c.Request.Method + " " + c.FullPath()
        ctx, span := tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindServer))
        defer span.End()

        // 设置span属性
        span.SetAttributes(
            attribute.String("http.method", c.Request.Method),
            attribute.String("http.url", c.Request.URL.String()),
            attribute.String("http.scheme", c.Request.URL.Scheme),
            attribute.String("http.host", c.Request.Host),
        )

        // 提取trace ID并存入context
        traceID := span.SpanContext().TraceID().String()
        ctx = contextx.WithTraceID(ctx, traceID)

        // 更新request context
        c.Request = c.Request.WithContext(ctx)

        // 响应header添加trace ID
        c.Header("X-Trace-ID", traceID)

        // 执行下一个处理器
        c.Next()

        // 记录响应状态
        span.SetAttributes(
            attribute.Int("http.status_code", c.Writer.Status()),
        )

        // 如果有错误，记录到span
        if len(c.Errors) > 0 {
            span.RecordError(c.Errors.Last())
        }
    }
}
```

### OTEL初始化

```go
// pkg/telemetry/otel.go
package telemetry

import (
    "context"

    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/propagation"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

type Options struct {
    ServiceName    string
    ServiceVersion string
    Endpoint       string // OTLP collector endpoint
}

// InitTracer 初始化OpenTelemetry追踪
func InitTracer(opts Options) (func(context.Context) error, error) {
    // 创建OTLP exporter
    exporter, err := otlptrace.New(
        context.Background(),
        otlptracegrpc.NewClient(
            otlptracegrpc.WithEndpoint(opts.Endpoint),
            otlptracegrpc.WithInsecure(),
        ),
    )
    if err != nil {
        return nil, err
    }

    // 创建resource
    res, err := resource.New(
        context.Background(),
        resource.WithAttributes(
            semconv.ServiceName(opts.ServiceName),
            semconv.ServiceVersion(opts.ServiceVersion),
        ),
    )
    if err != nil {
        return nil, err
    }

    // 创建trace provider
    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(res),
    )

    // 设置全局trace provider
    otel.SetTracerProvider(tp)

    // 设置全局propagator
    otel.SetTextMapPropagator(
        propagation.NewCompositeTextMapPropagator(
            propagation.TraceContext{},
            propagation.Baggage{},
        ),
    )

    // 返回清理函数
    return func(ctx context.Context) error {
        return tp.Shutdown(ctx)
    }, nil
}
```

### 使用示例

```go
// cmd/orchestrator/app/app.go
import (
    "github.com/kart-io/k8s-agent/pkg/telemetry"
    "github.com/kart-io/k8s-agent/common/middleware"
)

func (a *OrchestratorApp) Initialize() error {
    // 初始化OTEL
    cleanup, err := telemetry.InitTracer(telemetry.Options{
        ServiceName:    "orchestrator",
        ServiceVersion: "v1.0.0",
        Endpoint:       "localhost:4317",  // OTLP collector
    })
    if err != nil {
        return err
    }
    a.otelCleanup = cleanup

    // 应用追踪中间件
    router := gin.New()
    router.Use(middleware.Tracing("orchestrator"))

    return nil
}
```

---

## 错误处理最佳实践

### 域特定错误定义

```go
// pkg/errors/errors.go
package errors

import (
    "errors"
    "fmt"
    "net/http"
)

// 错误码
const (
    CodeOK                  = 0
    CodeInternalError       = 10000
    CodeInvalidParameter    = 10001
    CodeNotFound            = 10002
    CodeAlreadyExists       = 10003
    CodeUnauthorized        = 10004
    CodeForbidden           = 10005

    // 工作流相关错误
    CodeWorkflowNotFound    = 20001
    CodeWorkflowInvalid     = 20002
    CodeWorkflowExecuting   = 20003

    // 策略相关错误
    CodeStrategyNotFound    = 30001
    CodeStrategyInvalid     = 30002
)

// AppError 应用错误
type AppError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Err     error  `json:"-"`
}

func (e *AppError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
    }
    return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
    return e.Err
}

// HTTPStatus 返回对应的HTTP状态码
func (e *AppError) HTTPStatus() int {
    switch e.Code {
    case CodeOK:
        return http.StatusOK
    case CodeInvalidParameter:
        return http.StatusBadRequest
    case CodeNotFound, CodeWorkflowNotFound, CodeStrategyNotFound:
        return http.StatusNotFound
    case CodeAlreadyExists:
        return http.StatusConflict
    case CodeUnauthorized:
        return http.StatusUnauthorized
    case CodeForbidden:
        return http.StatusForbidden
    default:
        return http.StatusInternalServerError
    }
}

// 预定义错误构造函数
func New(code int, message string) *AppError {
    return &AppError{Code: code, Message: message}
}

func Wrap(err error, code int, message string) *AppError {
    return &AppError{Code: code, Message: message, Err: err}
}

// 工作流错误
func WorkflowNotFound(id string) *AppError {
    return &AppError{
        Code:    CodeWorkflowNotFound,
        Message: fmt.Sprintf("workflow %s not found", id),
    }
}

func WorkflowAlreadyExecuting(id string) *AppError {
    return &AppError{
        Code:    CodeWorkflowExecuting,
        Message: fmt.Sprintf("workflow %s is already executing", id),
    }
}

// IsNotFound 判断是否为NotFound错误
func IsNotFound(err error) bool {
    var appErr *AppError
    if errors.As(err, &appErr) {
        return appErr.Code == CodeNotFound ||
            appErr.Code == CodeWorkflowNotFound ||
            appErr.Code == CodeStrategyNotFound
    }
    return false
}
```

### 错误处理中间件

```go
// common/middleware/error_handler.go
package middleware

import (
    "errors"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/kart-io/k8s-agent/pkg/errors"
)

// ErrorHandler 错误处理中间件
func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        // 检查是否有错误
        if len(c.Errors) == 0 {
            return
        }

        // 获取最后一个错误
        err := c.Errors.Last().Err

        // 判断错误类型
        var appErr *errors.AppError
        if errors.As(err, &appErr) {
            c.JSON(appErr.HTTPStatus(), gin.H{
                "code":    appErr.Code,
                "message": appErr.Message,
            })
            return
        }

        // 默认内部错误
        c.JSON(http.StatusInternalServerError, gin.H{
            "code":    errors.CodeInternalError,
            "message": "Internal server error",
        })
    }
}
```

---

## 测试策略和Mock生成

### Mock生成配置

在每个接口文件添加mock生成指令：

```go
// internal/orchestrator/store/store.go
package store

//go:generate mockgen -destination mock_store.go -package store github.com/kart-io/k8s-agent/internal/orchestrator/store IStore,WorkflowStore,StrategyStore

type IStore interface {
    // ...
}
```

生成mock:

```bash
cd internal/orchestrator/store
go generate ./...
```

### 单元测试示例

```go
// internal/orchestrator/biz/workflow_test.go
package biz

import (
    "context"
    "testing"

    "github.com/golang/mock/gomock"
    "github.com/stretchr/testify/assert"

    "github.com/kart-io/k8s-agent/internal/orchestrator/model"
    "github.com/kart-io/k8s-agent/internal/orchestrator/store"
)

func TestWorkflowBiz_Create(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    // 创建mock store
    mockStore := store.NewMockIStore(ctrl)
    mockWorkflowStore := store.NewMockWorkflowStore(ctrl)
    mockStrategyStore := store.NewMockStrategyStore(ctrl)

    // 设置mock行为
    mockStore.EXPECT().Workflow().Return(mockWorkflowStore).AnyTimes()
    mockStore.EXPECT().Strategy().Return(mockStrategyStore).AnyTimes()
    mockStore.EXPECT().Execution().Return(nil).AnyTimes()

    // 模拟策略存在
    mockStrategyStore.EXPECT().
        Get(gomock.Any(), "strategy-1").
        Return(&model.StrategyM{ID: "strategy-1"}, nil)

    // 模拟TX执行
    mockStore.EXPECT().
        TX(gomock.Any(), gomock.Any()).
        DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
            return fn(ctx)
        })

    // 模拟工作流创建
    mockWorkflowStore.EXPECT().
        Create(gomock.Any(), gomock.Any()).
        Return(nil)

    // 创建biz实例
    workflowBiz := newWorkflowBiz(mockStore)

    // 执行测试
    req := &CreateWorkflowRequest{
        Name:       "test-workflow",
        ClusterID:  "cluster-1",
        StrategyID: "strategy-1",
    }

    workflow, err := workflowBiz.Create(context.Background(), req)

    // 断言
    assert.NoError(t, err)
    assert.NotNil(t, workflow)
    assert.Equal(t, "test-workflow", workflow.Name)
    assert.Equal(t, "cluster-1", workflow.ClusterID)
    assert.Equal(t, "pending", workflow.Status)
}
```

### 集成测试示例

```go
// internal/orchestrator/store/workflow_integration_test.go
//go:build integration
// +build integration

package store_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"

    "github.com/kart-io/k8s-agent/common/db"
    "github.com/kart-io/k8s-agent/internal/orchestrator/model"
    "github.com/kart-io/k8s-agent/internal/orchestrator/store"
)

type WorkflowStoreTestSuite struct {
    suite.Suite
    store store.IStore
}

func (s *WorkflowStoreTestSuite) SetupSuite() {
    // 连接测试数据库
    gormDB, _, err := db.NewMySQL(&db.MySQLOptions{
        Host:     "localhost",
        Port:     3306,
        Username: "test",
        Password: "test",
        Database: "test_orchestrator",
    })
    assert.NoError(s.T(), err)

    // 自动迁移
    err = gormDB.AutoMigrate(&model.WorkflowM{})
    assert.NoError(s.T(), err)

    s.store = store.NewStore(gormDB)
}

func (s *WorkflowStoreTestSuite) TearDownSuite() {
    // 清理测试数据
}

func (s *WorkflowStoreTestSuite) TestCreate() {
    ctx := context.Background()

    workflow := &model.WorkflowM{
        ID:        "test-1",
        Name:      "Test Workflow",
        ClusterID: "cluster-1",
        Status:    "pending",
    }

    err := s.store.Workflow().Create(ctx, workflow)
    assert.NoError(s.T(), err)

    // 验证创建成功
    got, err := s.store.Workflow().Get(ctx, "test-1")
    assert.NoError(s.T(), err)
    assert.Equal(s.T(), workflow.Name, got.Name)
}

func TestWorkflowStoreTestSuite(t *testing.T) {
    suite.Run(t, new(WorkflowStoreTestSuite))
}
```

运行集成测试：

```bash
go test -tags=integration ./internal/orchestrator/store/...
```

---

## 总结

本文档提供了OneX项目中的实战代码示例，涵盖：

1. **完整的3层架构**：Store → Biz → Handler
2. **Wire依赖注入**：从声明到生成到使用
3. **分布式追踪**：OpenTelemetry集成
4. **错误处理**：域特定错误和中间件
5. **测试策略**：Mock生成和集成测试

这些代码可以直接应用到k8s-agent项目中，建议按照ONEX_IMPLEMENTATION_GUIDE.md中的路线图逐步实施。

---

**相关文档**:
- ONEX_IMPLEMENTATION_GUIDE.md - 完整实施指南
- ONEX_LEARNINGS.md - OneX学习总结
