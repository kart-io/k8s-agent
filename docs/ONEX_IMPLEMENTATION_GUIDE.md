# OneX功能补充实施指南

**文档版本**: v1.0
**创建日期**: 2025-11-01
**基于**: OneX项目代码分析
**目标**: 为k8s-agent补充OneX中已验证的高价值功能

---

## 目录

1. [执行摘要](#执行摘要)
2. [高优先级功能](#高优先级功能)
3. [中优先级功能](#中优先级功能)
4. [实施路线图](#实施路线图)
5. [详细实施指南](#详细实施指南)

---

## 执行摘要

通过深入分析OneX项目（企业级单体仓库，33个服务），识别出k8s-agent项目当前缺失但极具价值的功能模式。本文档提供详细的实施指南，包括代码示例、优先级排序和实施路线图。

### 当前差距分析

| 功能维度 | k8s-agent | OneX | 优先级 | 收益 |
|---------|-----------|------|-------|------|
| 依赖注入 | 手动Bootstrap | Google Wire | ⭐⭐⭐⭐⭐ | 高 |
| 幂等性 | 无 | Redis+Middleware | ⭐⭐⭐⭐⭐ | 高 |
| Context管理 | 基础 | contextx包 | ⭐⭐⭐⭐ | 中 |
| 分布式追踪 | 部分 | 完整OTEL | ⭐⭐⭐⭐ | 中 |
| 配置管理 | Viper | Options模式 | ⭐⭐⭐ | 中 |
| Store层 | 手动 | 泛型+Wire | ⭐⭐⭐ | 中 |
| Middleware | Gin only | Kratos统一 | ⭐⭐ | 低 |
| RBAC | 无 | Casbin | ⭐ | 低 |

---

## 高优先级功能

### 1. 幂等性框架 ⭐⭐⭐⭐⭐

**优先级**: 最高
**实施周期**: 1周
**难度**: 低
**收益**: 防止重复操作，提升系统可靠性

#### 1.1 功能说明

幂等性确保相同的请求多次执行产生相同结果，对于诊断任务创建、工作流执行等场景至关重要。

#### 1.2 OneX实现分析

**核心组件**:
1. **pkg/idempotent/idempotent.go** - 幂等性核心逻辑
2. **pkg/idempotent/options.go** - 配置选项
3. **internal/pkg/idempotent/idempotent.go** - Wire集成
4. **internal/pkg/middleware/idempotent/idempotent.go** - Kratos中间件

**工作原理**:
```
1. 客户端请求时携带 X-Idempotent-ID header
2. 服务端检查Redis中是否存在该token
3. 如果存在，删除token并继续处理（一次性使用）
4. 如果不存在或已过期，返回错误
```

**关键代码**:

```go
// pkg/idempotent/idempotent.go
package idempotent

import (
    "context"
    "fmt"
    "time"

    "github.com/google/uuid"
    "github.com/redis/go-redis/v9"
)

// Redis Lua脚本：原子性读取+删除
const lua string = `
local current = redis.call('GET', KEYS[1])
if current == false then
    return '-1';  -- token不存在
end
local del = redis.call('DEL', KEYS[1])
if del == 1 then
     return '1';  -- 成功删除
else
     return '0';  -- 删除失败
end
`

type Idempotent struct {
    ops Options
}

type Options struct {
    redis  *redis.Client
    prefix string  // 键前缀，默认"idempotent"
    expire int     // 过期时间（分钟），默认60
}

func New(options ...func(*Options)) *Idempotent {
    ops := &Options{
        prefix: "idempotent",
        expire: 60,
    }
    for _, f := range options {
        f(ops)
    }
    return &Idempotent{ops: *ops}
}

// Token 生成新的幂等性token
func (i *Idempotent) Token(ctx context.Context) string {
    if i.ops.redis == nil {
        return ""
    }

    token := uuid.NewString()
    key := fmt.Sprintf("%s_%s", i.ops.prefix, token)
    i.ops.redis.Set(ctx, key, true, time.Duration(i.ops.expire)*time.Minute)
    return token
}

// Check 检查token是否有效（一次性使用）
func (i *Idempotent) Check(ctx context.Context, token string) bool {
    if i.ops.redis == nil {
        return true  // 无Redis时放行
    }

    key := fmt.Sprintf("%s_%s", i.ops.prefix, token)
    res, err := i.ops.redis.Eval(ctx, lua, []string{key}).Result()
    if err != nil || res != "1" {
        return false
    }

    return true
}
```

#### 1.3 k8s-agent实施方案

**目录结构**:
```
pkg/idempotent/
├── idempotent.go       # 核心逻辑
├── options.go          # 配置选项
└── idempotent_test.go  # 单元测试

common/middleware/
└── idempotent.go       # Gin中间件
```

**Gin中间件实现**:

```go
// common/middleware/idempotent.go
package middleware

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/kart-io/k8s-agent/pkg/idempotent"
    "github.com/kart-io/k8s-agent/common/response"
)

// IdempotentBlacklist 需要幂等性检查的API列表
var IdempotentBlacklist = map[string]bool{
    "POST /api/v1/workflows":     true,  // 创建工作流
    "POST /api/v1/tasks":         true,  // 创建任务
    "POST /api/v1/events":        true,  // 创建事件
}

func Idempotent(idt *idempotent.Idempotent) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 检查是否需要幂等性验证
        route := c.Request.Method + " " + c.FullPath()
        if !IdempotentBlacklist[route] {
            c.Next()
            return
        }

        // 提取幂等性token
        token := c.GetHeader("X-Idempotent-ID")
        if token == "" {
            response.Error(c, http.StatusBadRequest, "Missing X-Idempotent-ID header")
            c.Abort()
            return
        }

        // 验证token
        if !idt.Check(c.Request.Context(), token) {
            response.Error(c, http.StatusConflict, "Idempotent token is invalid or expired")
            c.Abort()
            return
        }

        c.Next()
    }
}
```

**服务集成示例**:

```go
// internal/orchestrator/initializers/http_server.go
import (
    "github.com/kart-io/k8s-agent/pkg/idempotent"
    "github.com/kart-io/k8s-agent/common/middleware"
)

func (i *HTTPServerInitializer) Initialize(ctx context.Context) error {
    // 创建幂等性实例
    idt := idempotent.New(
        idempotent.WithRedis(i.redisClient),
        idempotent.WithPrefix("orch"),
        idempotent.WithExpire(30), // 30分钟过期
    )

    // 应用中间件
    router := gin.New()
    router.Use(middleware.Idempotent(idt))

    // 定义路由
    router.POST("/api/v1/workflows", i.CreateWorkflow)

    return nil
}
```

**客户端使用示例**:

```bash
# 1. 获取幂等性token（可选，如果服务端提供token生成接口）
TOKEN=$(curl -X GET http://localhost:8081/api/v1/idempotent/token | jq -r '.token')

# 2. 使用token发起请求
curl -X POST http://localhost:8081/api/v1/workflows \
  -H "X-Idempotent-ID: $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "diagnose-pod-crash"}'

# 3. 重复请求会失败（token已被消费）
curl -X POST http://localhost:8081/api/v1/workflows \
  -H "X-Idempotent-ID: $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "diagnose-pod-crash"}'
# Response: {"error": "Idempotent token is invalid or expired"}
```

#### 1.4 测试用例

```go
// pkg/idempotent/idempotent_test.go
package idempotent_test

import (
    "context"
    "testing"
    "time"

    "github.com/alicebob/miniredis/v2"
    "github.com/redis/go-redis/v9"
    "github.com/stretchr/testify/assert"

    "github.com/kart-io/k8s-agent/pkg/idempotent"
)

func setupTestRedis(t *testing.T) (*redis.Client, func()) {
    mr, err := miniredis.Run()
    assert.NoError(t, err)

    client := redis.NewClient(&redis.Options{
        Addr: mr.Addr(),
    })

    cleanup := func() {
        client.Close()
        mr.Close()
    }

    return client, cleanup
}

func TestIdempotent_TokenGeneration(t *testing.T) {
    client, cleanup := setupTestRedis(t)
    defer cleanup()

    idt := idempotent.New(
        idempotent.WithRedis(client),
        idempotent.WithPrefix("test"),
    )

    ctx := context.Background()
    token := idt.Token(ctx)

    assert.NotEmpty(t, token)
    assert.True(t, idt.Check(ctx, token))
}

func TestIdempotent_TokenOneTimeUse(t *testing.T) {
    client, cleanup := setupTestRedis(t)
    defer cleanup()

    idt := idempotent.New(idempotent.WithRedis(client))

    ctx := context.Background()
    token := idt.Token(ctx)

    // 第一次使用成功
    assert.True(t, idt.Check(ctx, token))

    // 第二次使用失败（token已被删除）
    assert.False(t, idt.Check(ctx, token))
}

func TestIdempotent_TokenExpiration(t *testing.T) {
    client, cleanup := setupTestRedis(t)
    defer cleanup()

    idt := idempotent.New(
        idempotent.WithRedis(client),
        idempotent.WithExpire(1), // 1分钟过期
    )

    ctx := context.Background()
    token := idt.Token(ctx)

    // 模拟时间流逝（在实际测试中使用miniredis的FastForward）
    time.Sleep(61 * time.Second)

    assert.False(t, idt.Check(ctx, token))
}
```

#### 1.5 实施步骤

1. **创建基础包** (1天)
   - 复制OneX的 `pkg/idempotent/` 代码
   - 修改import路径
   - 添加单元测试

2. **实现Gin中间件** (0.5天)
   - 创建 `common/middleware/idempotent.go`
   - 定义IdempotentBlacklist
   - 处理header提取和验证

3. **集成到服务** (0.5天)
   - orchestrator服务集成
   - agent-manager服务集成
   - reasoning服务集成（如需要）

4. **测试验证** (1天)
   - 单元测试
   - 集成测试
   - 压力测试（重复请求）

5. **文档更新** (0.5天)
   - API文档更新
   - 使用示例
   - 故障排查指南

---

### 2. Context管理包 (contextx) ⭐⭐⭐⭐

**优先级**: 高
**实施周期**: 2天
**难度**: 低
**收益**: 统一context值管理，支持追踪和审计

#### 2.1 功能说明

OneX的contextx包提供类型安全的context值管理，避免魔法字符串键，支持JWT claims、用户ID、traceID等信息传递。

#### 2.2 OneX实现

```go
// internal/pkg/contextx/contextx.go
package contextx

import (
    "context"
    "github.com/golang-jwt/jwt/v4"
)

// 定义类型安全的context键
type (
    claimsKey      struct{}
    userIDKey      struct{}
    traceIDKey     struct{}
    accessTokenKey struct{}
)

// WithClaims 将JWT claims存入context
func WithClaims(ctx context.Context, claims *jwt.RegisteredClaims) context.Context {
    return context.WithValue(ctx, claimsKey{}, claims)
}

// Claims 从context提取JWT claims
func Claims(ctx context.Context) *jwt.RegisteredClaims {
    claims, _ := ctx.Value(claimsKey{}).(*jwt.RegisteredClaims)
    return claims
}

// WithUserID 将用户ID存入context
func WithUserID(ctx context.Context, userID string) context.Context {
    return context.WithValue(ctx, userIDKey{}, userID)
}

// UserID 从context提取用户ID
func UserID(ctx context.Context) string {
    userID, _ := ctx.Value(userIDKey{}).(string)
    return userID
}

// WithTraceID 将追踪ID存入context
func WithTraceID(ctx context.Context, traceID string) context.Context {
    return context.WithValue(ctx, traceIDKey{}, traceID)
}

// TraceID 从context提取追踪ID
func TraceID(ctx context.Context) string {
    traceID, _ := ctx.Value(traceIDKey{}).(string)
    return traceID
}

// WithAccessToken 将访问令牌存入context
func WithAccessToken(ctx context.Context, accessToken string) context.Context {
    return context.WithValue(ctx, accessTokenKey{}, accessToken)
}

// AccessToken 从context提取访问令牌
func AccessToken(ctx context.Context) string {
    accessToken, _ := ctx.Value(accessTokenKey{}).(string)
    return accessToken
}
```

#### 2.3 k8s-agent扩展实现

```go
// pkg/contextx/contextx.go
package contextx

import (
    "context"
    "github.com/golang-jwt/jwt/v4"
)

// k8s-agent特定的context键
type (
    claimsKey       struct{}
    userIDKey       struct{}
    traceIDKey      struct{}
    requestIDKey    struct{}
    agentIDKey      struct{}
    clusterIDKey    struct{}
    workflowIDKey   struct{}
    taskIDKey       struct{}
)

// ============ JWT Claims ============

func WithClaims(ctx context.Context, claims *jwt.RegisteredClaims) context.Context {
    return context.WithValue(ctx, claimsKey{}, claims)
}

func Claims(ctx context.Context) *jwt.RegisteredClaims {
    claims, _ := ctx.Value(claimsKey{}).(*jwt.RegisteredClaims)
    return claims
}

// ============ 用户信息 ============

func WithUserID(ctx context.Context, userID string) context.Context {
    return context.WithValue(ctx, userIDKey{}, userID)
}

func UserID(ctx context.Context) string {
    userID, _ := ctx.Value(userIDKey{}).(string)
    return userID
}

// ============ 追踪信息 ============

func WithTraceID(ctx context.Context, traceID string) context.Context {
    return context.WithValue(ctx, traceIDKey{}, traceID)
}

func TraceID(ctx context.Context) string {
    traceID, _ := ctx.Value(traceIDKey{}).(string)
    return traceID
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
    return context.WithValue(ctx, requestIDKey{}, requestID)
}

func RequestID(ctx context.Context) string {
    requestID, _ := ctx.Value(requestIDKey{}).(string)
    return requestID
}

// ============ k8s-agent特定 ============

func WithAgentID(ctx context.Context, agentID string) context.Context {
    return context.WithValue(ctx, agentIDKey{}, agentID)
}

func AgentID(ctx context.Context) string {
    agentID, _ := ctx.Value(agentIDKey{}).(string)
    return agentID
}

func WithClusterID(ctx context.Context, clusterID string) context.Context {
    return context.WithValue(ctx, clusterIDKey{}, clusterID)
}

func ClusterID(ctx context.Context) string {
    clusterID, _ := ctx.Value(clusterIDKey{}).(string)
    return clusterID
}

func WithWorkflowID(ctx context.Context, workflowID string) context.Context {
    return context.WithValue(ctx, workflowIDKey{}, workflowID)
}

func WorkflowID(ctx context.Context) string {
    workflowID, _ := ctx.Value(workflowIDKey{}).(string)
    return workflowID
}

func WithTaskID(ctx context.Context, taskID string) context.Context {
    return context.WithValue(ctx, taskIDKey{}, taskID)
}

func TaskID(ctx context.Context) string {
    taskID, _ := ctx.Value(taskIDKey{}).(string)
    return taskID
}

// ============ 辅助函数 ============

// ExtractAll 从context提取所有常用信息，用于日志
func ExtractAll(ctx context.Context) map[string]string {
    return map[string]string{
        "user_id":     UserID(ctx),
        "trace_id":    TraceID(ctx),
        "request_id":  RequestID(ctx),
        "agent_id":    AgentID(ctx),
        "cluster_id":  ClusterID(ctx),
        "workflow_id": WorkflowID(ctx),
        "task_id":     TaskID(ctx),
    }
}
```

#### 2.4 使用示例

**中间件集成**:

```go
// common/middleware/request_id.go
package middleware

import (
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"

    "github.com/kart-io/k8s-agent/pkg/contextx"
)

func RequestID() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 从header获取或生成新的request ID
        requestID := c.GetHeader("X-Request-ID")
        if requestID == "" {
            requestID = uuid.NewString()
        }

        // 存入context
        ctx := contextx.WithRequestID(c.Request.Context(), requestID)
        c.Request = c.Request.WithContext(ctx)

        // 响应header
        c.Header("X-Request-ID", requestID)

        c.Next()
    }
}
```

**日志集成**:

```go
// common/middleware/logging.go
package middleware

import (
    "time"

    "github.com/gin-gonic/gin"
    "github.com/kart-io/logger"
    "github.com/kart-io/k8s-agent/pkg/contextx"
)

func Logger(log logger.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path

        c.Next()

        // 从context提取信息进行结构化日志
        ctx := c.Request.Context()
        log.Info("HTTP request",
            logger.String("method", c.Request.Method),
            logger.String("path", path),
            logger.Int("status", c.Writer.Status()),
            logger.Duration("latency", time.Since(start)),
            logger.String("request_id", contextx.RequestID(ctx)),
            logger.String("trace_id", contextx.TraceID(ctx)),
            logger.String("user_id", contextx.UserID(ctx)),
            logger.String("cluster_id", contextx.ClusterID(ctx)),
        )
    }
}
```

**业务代码使用**:

```go
// internal/orchestrator/workflow/executor.go
package workflow

import (
    "context"
    "github.com/kart-io/logger"
    "github.com/kart-io/k8s-agent/pkg/contextx"
)

func (e *Executor) ExecuteWorkflow(ctx context.Context, workflow *Workflow) error {
    // 将workflow ID存入context
    ctx = contextx.WithWorkflowID(ctx, workflow.ID)
    ctx = contextx.WithClusterID(ctx, workflow.ClusterID)

    e.logger.Info("Executing workflow",
        logger.String("workflow_id", contextx.WorkflowID(ctx)),
        logger.String("cluster_id", contextx.ClusterID(ctx)),
        logger.String("trace_id", contextx.TraceID(ctx)),
    )

    for _, step := range workflow.Steps {
        // 传递context到各个步骤
        if err := e.ExecuteStep(ctx, step); err != nil {
            return err
        }
    }

    return nil
}

func (e *Executor) ExecuteStep(ctx context.Context, step *Step) error {
    // 添加task ID到context
    ctx = contextx.WithTaskID(ctx, step.ID)

    // 所有日志都自动包含workflow_id, cluster_id, task_id等信息
    e.logger.Debug("Executing step",
        logger.String("step_type", step.Type),
        logger.String("task_id", contextx.TaskID(ctx)),
    )

    // ... 执行逻辑
    return nil
}
```

#### 2.5 实施步骤

1. **创建pkg/contextx包** (0.5天)
   - 实现基础函数（UserID, TraceID等）
   - 添加k8s-agent特定函数
   - 添加ExtractAll辅助函数

2. **更新中间件** (0.5天)
   - RequestID中间件
   - Logger中间件集成
   - TraceID中间件（OTEL集成）

3. **重构现有代码** (0.5天)
   - 替换魔法字符串键
   - 统一context值访问方式

4. **测试和文档** (0.5天)
   - 单元测试
   - 集成测试
   - 更新使用文档

---

### 3. Google Wire依赖注入 ⭐⭐⭐⭐

**优先级**: 高
**实施周期**: 1-2周（按服务逐步迁移）
**难度**: 中
**收益**: 编译时类型安全，消除手动依赖管理错误

#### 3.1 功能说明

Google Wire是编译时依赖注入工具，通过代码生成消除运行时反射开销，提供类型安全的依赖图构建。

#### 3.2 当前k8s-agent的问题

```go
// 当前方式：手动管理依赖关系
func (a *OrchestratorApp) Initialize(ctx context.Context) error {
    // 必须手动确保初始化顺序
    if err := a.dbInit.Initialize(ctx); err != nil {
        return err
    }

    if err := a.redisInit.Initialize(ctx); err != nil {
        return err
    }

    // 容易出现初始化顺序错误
    // 难以发现循环依赖
}
```

#### 3.3 OneX Wire实现

**wire.go** (声明依赖关系):

```go
// cmd/onex-controller-manager/app/wire.go
//go:build wireinject
// +build wireinject

package app

//go:generate go run github.com/google/wire/cmd/wire

import (
    "github.com/google/wire"
    "github.com/onexstack/onex/internal/gateway/store"
    "github.com/onexstack/onexstack/pkg/db"
)

// wireStoreClient 声明如何构建IStore
func wireStoreClient(*db.MySQLOptions) (store.IStore, error) {
    wire.Build(
        db.ProviderSet,      // MySQL client provider
        store.ProviderSet,   // Store provider
    )

    return nil, nil  // 仅用于类型推断，wire会生成实际代码
}
```

**wire_gen.go** (自动生成):

```go
// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package app

import (
    "github.com/onexstack/onex/internal/gateway/store"
    "github.com/onexstack/onexstack/pkg/db"
)

// Injectors from wire.go:

func wireStoreClient(options *db.MySQLOptions) (store.IStore, error) {
    gormDB, err := db.NewMySQL(options)
    if err != nil {
        return nil, err
    }
    datastore := store.NewStore(gormDB)
    return datastore, nil
}
```

**ProviderSet模式**:

```go
// internal/gateway/store/store.go
package store

import "github.com/google/wire"

// ProviderSet 声明这个包提供什么
var ProviderSet = wire.NewSet(
    NewStore,                          // 构造函数
    wire.Bind(new(IStore), new(*datastore)),  // 接口绑定
)

// NewStore 是provider函数
func NewStore(db *gorm.DB) *datastore {
    return &datastore{core: db}
}
```

#### 3.4 k8s-agent实施方案

**阶段1: orchestrator服务试点**

```go
// cmd/orchestrator/app/wire.go
//go:build wireinject
// +build wireinject

package app

//go:generate go run github.com/google/wire/cmd/wire

import (
    "github.com/google/wire"

    "github.com/kart-io/k8s-agent/internal/orchestrator/store"
    "github.com/kart-io/k8s-agent/internal/orchestrator/biz"
    "github.com/kart-io/k8s-agent/internal/orchestrator/handler"
    "github.com/kart-io/k8s-agent/common/db"
    "github.com/kart-io/k8s-agent/common/config"
)

// wireApp 构建完整的Orchestrator应用
func wireApp(*config.ServerOptions) (*OrchestratorApp, error) {
    wire.Build(
        // Database layer
        db.ProviderSet,

        // Store layer
        store.ProviderSet,

        // Business layer
        biz.ProviderSet,

        // Handler layer
        handler.ProviderSet,

        // HTTP Server
        NewHTTPServer,

        // gRPC Server
        NewGRPCServer,

        // Application
        NewOrchestratorApp,
    )

    return nil, nil
}
```

**Store层ProviderSet**:

```go
// internal/orchestrator/store/store.go
package store

import (
    "github.com/google/wire"
    "gorm.io/gorm"
)

var ProviderSet = wire.NewSet(
    NewStore,
    wire.Bind(new(IStore), new(*store)),
)

type IStore interface {
    Workflow() WorkflowStore
    Strategy() StrategyStore
    Execution() ExecutionStore
}

type store struct {
    db *gorm.DB
}

func NewStore(db *gorm.DB) *store {
    return &store{db: db}
}
```

**Biz层ProviderSet**:

```go
// internal/orchestrator/biz/biz.go
package biz

import (
    "github.com/google/wire"
    "github.com/kart-io/k8s-agent/internal/orchestrator/store"
)

var ProviderSet = wire.NewSet(
    NewWorkflowBiz,
    wire.Bind(new(IWorkflowBiz), new(*workflowBiz)),
)

type IWorkflowBiz interface {
    Create(ctx context.Context, req *CreateWorkflowRequest) (*Workflow, error)
    Execute(ctx context.Context, workflowID string) error
}

type workflowBiz struct {
    store store.IStore
}

// NewWorkflowBiz 是provider函数，Wire自动注入store
func NewWorkflowBiz(store store.IStore) *workflowBiz {
    return &workflowBiz{store: store}
}
```

**Handler层ProviderSet**:

```go
// internal/orchestrator/handler/handler.go
package handler

import (
    "github.com/google/wire"
    "github.com/kart-io/k8s-agent/internal/orchestrator/biz"
)

var ProviderSet = wire.NewSet(
    NewWorkflowHandler,
)

type WorkflowHandler struct {
    biz biz.IWorkflowBiz
}

// NewWorkflowHandler Wire自动注入biz
func NewWorkflowHandler(biz biz.IWorkflowBiz) *WorkflowHandler {
    return &WorkflowHandler{biz: biz}
}

func (h *WorkflowHandler) CreateWorkflow(c *gin.Context) {
    // 使用h.biz
}
```

**Application构建**:

```go
// cmd/orchestrator/app/app.go
package app

import (
    "github.com/kart-io/k8s-agent/internal/orchestrator/handler"
)

type OrchestratorApp struct {
    httpServer *http.Server
    grpcServer *grpc.Server
    handler    *handler.WorkflowHandler
}

// NewOrchestratorApp Wire自动注入所有依赖
func NewOrchestratorApp(
    httpServer *http.Server,
    grpcServer *grpc.Server,
    handler *handler.WorkflowHandler,
) *OrchestratorApp {
    return &OrchestratorApp{
        httpServer: httpServer,
        grpcServer: grpcServer,
        handler:    handler,
    }
}
```

**生成Wire代码**:

```bash
cd cmd/orchestrator/app
go generate ./...

# 生成 wire_gen.go
```

#### 3.5 迁移步骤

**阶段1: orchestrator服务 (1周)**

1. 安装Wire工具
   ```bash
   go install github.com/google/wire/cmd/wire@latest
   ```

2. 重构Store层
   - 创建IStore接口
   - 添加ProviderSet
   - 编写provider函数

3. 重构Biz层
   - 创建IBiz接口
   - 依赖IStore接口
   - 添加ProviderSet

4. 重构Handler层
   - 依赖IBiz接口
   - 添加ProviderSet

5. 创建wire.go
   - 声明依赖图
   - 运行`go generate`

6. 测试验证
   - 单元测试
   - 集成测试

**阶段2: agent-manager服务 (1周)**

重复阶段1的步骤

**阶段3: 其他服务 (按需)**

- reasoning服务
- auth服务
- cluster服务

#### 3.6 Wire最佳实践

1. **ProviderSet命名**: 每个包一个`ProviderSet`
2. **接口绑定**: 使用`wire.Bind`绑定接口和实现
3. **构造函数**: Provider函数应返回`(T, error)`或`T`
4. **Cleanup**: 使用`wire.Cleanup`注册清理函数
5. **编译标签**: wire.go使用`//go:build wireinject`，wire_gen.go使用`//go:build !wireinject`

---

## 中优先级功能

### 4. Options配置模式 ⭐⭐⭐

**优先级**: 中
**实施周期**: 1周
**难度**: 低
**收益**: 统一配置管理，支持CLI和文件

（详细内容省略，参考OneX的pkg/config/options/模式）

### 5. 泛型Store/Repository ⭐⭐⭐

**优先级**: 中
**实施周期**: 2周
**难度**: 中
**收益**: 减少CRUD代码重复

（详细内容省略）

### 6. 分布式追踪增强 (OpenTelemetry) ⭐⭐⭐⭐

**优先级**: 中
**实施周期**: 1-2周
**难度**: 中
**收益**: 完整的分布式追踪能力

（详细内容省略）

---

## 实施路线图

### Sprint 1 (Week 1-2): 高优先级快速胜利

**目标**: 快速交付高价值功能

| Week | 功能 | 人天 | 产出 |
|------|-----|------|------|
| 1 | 幂等性框架 | 5 | pkg/idempotent + 中间件 |
| 1-2 | contextx包 | 2 | pkg/contextx + 集成 |
| 2 | Wire试点(orchestrator) | 5 | wire.go + 重构 |

**里程碑**:
- ✅ 幂等性框架可用
- ✅ contextx包集成到3个服务
- ✅ orchestrator服务Wire化

### Sprint 2 (Week 3-4): 扩展和优化

**目标**: 扩展到更多服务，优化现有实现

| Week | 功能 | 人天 | 产出 |
|------|-----|------|------|
| 3 | Wire扩展(agent-manager) | 5 | agent-manager Wire化 |
| 3-4 | Options模式 | 3 | 配置重构 |
| 4 | 泛型Store | 5 | 通用CRUD层 |

**里程碑**:
- ✅ 2个核心服务Wire化
- ✅ 统一配置管理
- ✅ 泛型Store可用

### Sprint 3 (Week 5-6): 高级功能

**目标**: 增强可观测性和中间件

| Week | 功能 | 人天 | 产出 |
|------|-----|------|------|
| 5 | OTEL增强 | 5 | 完整追踪 |
| 5-6 | Kratos中间件（可选） | 5 | 统一中间件 |
| 6 | RBAC（可选） | 3 | Casbin集成 |

**里程碑**:
- ✅ 完整分布式追踪
- ✅ 统一中间件系统（如选择Kratos）
- ✅ RBAC授权（如需要）

---

## 风险和缓解措施

### 风险1: Wire学习曲线

**风险**: 团队不熟悉Wire可能导致错误使用

**缓解**:
- 从单个服务试点开始
- 提供详细培训和文档
- Code Review重点关注Wire使用

### 风险2: 大规模重构影响稳定性

**风险**: 同时重构多个服务可能引入bug

**缓解**:
- 逐服务迁移，不并行重构
- 每个服务迁移后充分测试
- 保留旧代码作为回退方案

### 风险3: 性能影响

**风险**: 新功能可能影响性能

**缓解**:
- 幂等性使用Redis，性能影响小
- Wire是编译时工具，零运行时开销
- contextx是轻量级包装，影响可忽略
- 每个功能都进行性能基准测试

---

## 成功指标

### 技术指标

- **代码质量**:
  - 单元测试覆盖率 > 80%
  - golangci-lint zero warnings
  - 循环复杂度 < 15

- **性能**:
  - 幂等性检查延迟 < 5ms (P99)
  - Wire构建时间增加 < 10%
  - 内存使用无明显增加

- **可维护性**:
  - 依赖注入错误在编译时发现 (Wire)
  - 重复代码减少 > 30% (泛型Store)
  - 配置管理统一化 100%

### 业务指标

- **可靠性**:
  - 重复任务创建减少 > 95% (幂等性)
  - 初始化错误减少 100% (Wire)

- **可观测性**:
  - 全链路追踪覆盖率 100%
  - 日志结构化率 100%
  - 审计日志完整性 100%

---

## 参考资料

1. **OneX项目**:
   - GitHub: https://github.com/onexstack/onex
   - 文档: /Users/costalong/code/go/src/github.com/onexstack/onex/docs

2. **Google Wire**:
   - 官方文档: https://github.com/google/wire/blob/main/docs/guide.md
   - 最佳实践: https://github.com/google/wire/blob/main/docs/best-practices.md

3. **幂等性设计**:
   - RESTful API幂等性: https://restfulapi.net/idempotent-rest-apis/
   - 分布式系统幂等性: https://www.baeldung.com/cs/idempotent-operations

4. **Go Context最佳实践**:
   - 官方博客: https://go.dev/blog/context
   - Context使用模式: https://pkg.go.dev/context

---

**文档维护者**: Aetherius开发团队
**最后更新**: 2025-11-01
**版本**: v1.0
