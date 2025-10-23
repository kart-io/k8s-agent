# pkg/app 与 pkg/bootstrap 组合使用文档

## 概述

本文档说明如何在 agent-manager 中组合使用 `pkg/app/` 和 `pkg/bootstrap/` 两个框架,实现标准化的 CLI 应用启动和灵活的组件生命周期管理。

## 架构设计

### 分层架构

```
┌─────────────────────────────────────────────────────┐
│  pkg/app/ - 命令行应用框架                           │
│  ┌───────────────────────────────────────────┐      │
│  │  - Cobra 命令行解析                       │      │
│  │  - Viper 配置管理                         │      │
│  │  - Application 接口                       │      │
│  │  - 版本信息集成                           │      │
│  └─────────────┬─────────────────────────────┘      │
└────────────────┼────────────────────────────────────┘
                 │ 使用
                 ↓
┌─────────────────────────────────────────────────────┐
│  pkg/bootstrap/ - 组件生命周期管理                  │
│  ┌───────────────────────────────────────────┐      │
│  │  - Initializer 接口                       │      │
│  │  - 优先级排序 (Priority)                  │      │
│  │  - 自动关闭 (Closer)                      │      │
│  │  - 健康检查 (HealthChecker)               │      │
│  └───────────────────────────────────────────┘      │
└─────────────────────────────────────────────────────┘
```

### 组件初始化顺序

使用 `bootstrap.Priority` 常量控制初始化顺序:

```go
PriorityConfig   = 100  // 配置加载
PriorityLogger   = 200  // 日志系统
PriorityDatabase = 300  // 数据库连接
PriorityCache    = 400  // Redis 缓存
PriorityMQ       = 500  // NATS 消息队列
PriorityHTTP     = 600  // HTTP 服务器
PriorityGRPC     = 700  // gRPC 服务器
```

实际初始化顺序:

```
1. Database (300)
2. Redis (400)
3. Registry (450) - 依赖 Database + Redis
4. NATS (500)
5. Dispatcher (550) - 依赖 NATS
6. HTTP Server (600)
7. gRPC Server (700) - 可选
```

## 代码组织

### 目录结构

```
cmd/agent-manager/
├── main.go                          # 程序入口
└── app/
    └── app.go                       # 应用程序启动逻辑

internal/agent-manager/
├── config/
│   └── options.go                   # 配置选项
└── initializers/                    # 组件初始化器
    ├── database.go                  # DatabaseInitializer
    ├── redis.go                     # RedisInitializer
    ├── services.go                  # Registry/NATS/Dispatcher
    └── servers.go                   # HTTP/gRPC Server
```

### 关键文件

#### 1. cmd/agent-manager/main.go

```go
package main

import "github.com/kart-io/k8s-agent/cmd/agent-manager/app"

func main() {
    app.Execute()
}
```

#### 2. cmd/agent-manager/app/app.go

```go
// AgentManagerApp 实现 commonapp.Application 接口
type AgentManagerApp struct {
    bootstrap *bootstrap.Bootstrap
    opts      *config.Options
    logger    core.Logger

    // 组件初始化器
    dbInit         *initializers.DatabaseInitializer
    redisInit      *initializers.RedisInitializer
    // ... 其他初始化器
}

// Initialize 初始化所有组件
func (a *AgentManagerApp) Initialize(ctx context.Context, opts commonapp.Options) error {
    a.opts = opts.(*config.Options)

    // 创建 bootstrap
    a.bootstrap = bootstrap.New(a.logrusLog)

    // 注册组件
    a.registerComponents()

    // 执行初始化
    return a.bootstrap.Initialize(ctx)
}

// Run 运行应用
func (a *AgentManagerApp) Run(ctx context.Context) error {
    return a.bootstrap.Run(ctx, nil)
}

// Shutdown 优雅关闭
func (a *AgentManagerApp) Shutdown(ctx context.Context) error {
    return a.bootstrap.Shutdown(ctx)
}
```

## 组件初始化器实现

### Initializer 接口

每个组件都实现 `bootstrap.Initializer` 接口:

```go
type Initializer interface {
    Name() string                        // 组件名称
    Initialize(ctx context.Context) error // 初始化
    Priority() int                        // 优先级
}
```

可选接口:

```go
type Closer interface {
    Close(ctx context.Context) error     // 关闭清理
}

type HealthChecker interface {
    HealthCheck(ctx context.Context) error // 健康检查
}
```

### 示例: DatabaseInitializer

```go
package initializers

import (
    "context"
    "github.com/kart-io/k8s-agent/pkg/bootstrap"
)

type DatabaseInitializer struct {
    opts   *config.Options
    logger core.Logger
    store  *storage.PostgresStore
}

func NewDatabaseInitializer(opts *config.Options, logger core.Logger) *DatabaseInitializer {
    return &DatabaseInitializer{opts: opts, logger: logger}
}

// 实现 Initializer 接口
func (d *DatabaseInitializer) Name() string {
    return "database"
}

func (d *DatabaseInitializer) Priority() int {
    return bootstrap.PriorityDatabase
}

func (d *DatabaseInitializer) Initialize(ctx context.Context) error {
    // 创建数据库连接
    mysqlClient, err := db.NewMySQL(d.logger, ...)
    if err != nil {
        return err
    }

    d.store = &storage.PostgresStore{MySQLClient: mysqlClient}

    // 自动迁移
    if d.opts.Database.AutoMigrate {
        return d.store.AutoMigrate(...)
    }
    return nil
}

// 实现 Closer 接口
func (d *DatabaseInitializer) Close(ctx context.Context) error {
    if d.store != nil {
        return d.store.Close()
    }
    return nil
}

// 实现 HealthChecker 接口
func (d *DatabaseInitializer) HealthCheck(ctx context.Context) error {
    if d.store == nil {
        return fmt.Errorf("database not initialized")
    }
    return nil
}

// 提供访问器
func (d *DatabaseInitializer) Store() *storage.PostgresStore {
    return d.store
}
```

## 组件依赖注入

组件之间通过构造函数注入依赖:

```go
// RegistryInitializer 依赖 Database 和 Redis
func NewRegistryInitializer(
    opts *config.Options,
    logger core.Logger,
    dbInit *DatabaseInitializer,      // 依赖注入
    redisInit *RedisInitializer,      // 依赖注入
) *RegistryInitializer {
    return &RegistryInitializer{
        opts:      opts,
        logger:    logger,
        dbInit:    dbInit,
        redisInit: redisInit,
    }
}

// Initialize 时使用依赖
func (r *RegistryInitializer) Initialize(ctx context.Context) error {
    r.registry = agent.NewRegistry(
        r.dbInit.Store(),      // 获取 database store
        r.redisInit.Store(),   // 获取 redis store
        r.logger,
    )
    return r.registry.Start(ctx)
}
```

## 启动流程

### 完整启动流程

```
1. main.go 调用 app.Execute()
   ↓
2. pkg/app 创建 Cobra 命令
   ├── 解析命令行参数
   ├── 加载配置文件 (Viper)
   ├── 验证配置 (Complete + Validate)
   └── 初始化日志系统
   ↓
3. 调用 AgentManagerApp.Initialize()
   ├── 创建 bootstrap 实例
   ├── 注册所有组件初始化器
   └── 执行 bootstrap.Initialize()
       ├── 按 Priority 排序
       ├── 依次初始化各组件
       │   ├── Database (300)
       │   ├── Redis (400)
       │   ├── Registry (450)
       │   ├── NATS (500)
       │   ├── Dispatcher (550)
       │   ├── HTTP Server (600)
       │   └── gRPC Server (700)
       └── 记录初始化日志
   ↓
4. 调用 AgentManagerApp.Run()
   ├── HTTP/gRPC 服务器在后台运行
   └── bootstrap.Run() 等待信号
   ↓
5. 收到信号 (SIGINT/SIGTERM)
   ↓
6. 调用 AgentManagerApp.Shutdown()
   └── bootstrap.Shutdown()
       ├── 反向顺序关闭组件
       │   ├── gRPC Server (700)
       │   ├── HTTP Server (600)
       │   ├── Dispatcher (550)
       │   ├── NATS (500)
       │   ├── Registry (450)
       │   ├── Redis (400)
       │   └── Database (300)
       └── 记录关闭日志
```

### 日志输出示例

```
INFO Starting Agent Manager Service  http_port=8080 grpc_enabled=true grpc_port=9090
INFO Initializing database (priority: 300)
INFO Initialized database in 125ms
INFO Initializing redis (priority: 400)
INFO Initialized redis in 45ms
INFO Initializing registry (priority: 450)
INFO Initialized registry in 82ms
INFO Initializing nats (priority: 500)
INFO Initialized nats in 156ms
INFO Initializing dispatcher (priority: 550)
INFO Initialized dispatcher in 12ms
INFO Initializing http-server (priority: 600)
INFO Initialized http-server in 35ms
INFO Initializing grpc-server (priority: 700)
INFO Initialized grpc-server in 28ms
INFO All components initialized successfully
INFO Agent Manager Service started successfully  http_address=0.0.0.0:8080 grpc_enabled=true

<收到信号 Ctrl+C>

INFO Shutting down Agent Manager Service
INFO Shutting down grpc-server
INFO Closed grpc-server in 15ms
INFO Shutting down http-server
INFO Closed http-server in 32ms
INFO Shutting down dispatcher
INFO Shutting down nats
INFO Closed nats in 45ms
INFO Shutting down registry
INFO Closed registry in 18ms
INFO Shutting down redis
INFO Closed redis in 5ms
INFO Shutting down database
INFO Closed database in 12ms
INFO All components shut down successfully
```

## 优势

### 1. 清晰的职责分离

- **pkg/app**: 处理 CLI、配置、版本管理
- **pkg/bootstrap**: 处理组件生命周期
- **initializers**: 具体组件初始化逻辑

### 2. 可维护性

- 每个组件独立的 Initializer 文件
- 依赖关系显式声明
- 优先级一目了然

### 3. 可测试性

```go
// 单元测试示例
func TestDatabaseInitializer(t *testing.T) {
    opts := &config.Options{
        Database: config.DatabaseOptions{
            Host: "localhost",
            Port: 3306,
        },
    }
    logger := logrus.New()

    init := NewDatabaseInitializer(opts, logger)

    ctx := context.Background()
    err := init.Initialize(ctx)
    assert.NoError(t, err)

    // 验证初始化结果
    assert.NotNil(t, init.Store())

    // 清理
    err = init.Close(ctx)
    assert.NoError(t, err)
}
```

### 4. 灵活性

- 轻松添加/移除组件
- 支持条件初始化 (如 gRPC 可选)
- 支持重试机制 (`RetryInitializer`)

### 5. 自动化

- 自动按优先级排序
- 自动反向顺序关闭
- 自动健康检查

## 扩展

### 添加新组件

1. 创建 Initializer:

```go
// internal/agent-manager/initializers/mycomponent.go
type MyComponentInitializer struct {
    opts   *config.Options
    logger core.Logger
    component *MyComponent
}

func NewMyComponentInitializer(opts *config.Options, logger core.Logger) *MyComponentInitializer {
    return &MyComponentInitializer{opts: opts, logger: logger}
}

func (m *MyComponentInitializer) Name() string {
    return "my-component"
}

func (m *MyComponentInitializer) Priority() int {
    return bootstrap.PriorityMQ + 100 // 自定义优先级
}

func (m *MyComponentInitializer) Initialize(ctx context.Context) error {
    m.component = NewMyComponent(m.opts)
    return m.component.Start()
}

func (m *MyComponentInitializer) Close(ctx context.Context) error {
    return m.component.Stop()
}
```

2. 在 `app.go` 中注册:

```go
func (a *AgentManagerApp) registerComponents() {
    // ... 其他组件

    // 添加新组件
    myComponentInit := initializers.NewMyComponentInitializer(a.opts, a.logger)
    a.bootstrap.Register(myComponentInit)
}
```

## 总结

通过组合使用 `pkg/app/` 和 `pkg/bootstrap/`,agent-manager 实现了:

✅ **标准化的 CLI 应用框架** - Cobra + Viper
✅ **优雅的组件生命周期管理** - 初始化 → 运行 → 关闭
✅ **清晰的依赖关系** - 显式注入, 优先级控制
✅ **完整的可观测性** - 日志记录、健康检查
✅ **易于扩展** - 添加新组件只需实现接口
✅ **生产就绪** - 信号处理、优雅关闭、错误处理

这种架构模式可以应用到其他服务 (orchestrator, reasoning 等),实现统一的启动管理模式。
