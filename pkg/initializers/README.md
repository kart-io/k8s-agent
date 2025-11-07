# pkg/initializers - 通用初始化器库

本包提供了一组通用的初始化器，用于简化服务启动过程中的组件初始化。这些初始化器遵循 `pkg/bootstrap` 定义的接口，可以在所有服务中复用。

## 目录

- [简介](#简介)
- [可用初始化器](#可用初始化器)
- [使用指南](#使用指南)
- [最佳实践](#最佳实践)
- [迁移指南](#迁移指南)

## 简介

通用初始化器库解决了以下问题：

- 减少重复代码：数据库、Redis、NATS 等基础设施的初始化逻辑在多个服务中重复
- 统一初始化模式：所有服务使用相同的初始化方式
- 简化服务创建：新服务只需导入并注册初始化器即可
- 提高可维护性：基础设施变更只需修改一处

## 可用初始化器

### 1. DatabaseInitializer

数据库初始化器，负责 MySQL 连接初始化。

**功能特性**：

- 数据库连接初始化
- 连接池配置
- 可选的自动迁移（AutoMigrate）
- 健康检查
- 优雅关闭

**优先级**: `bootstrap.PriorityDatabase (300)`

**使用示例**：

```go
import (
    "github.com/kart-io/k8s-agent/pkg/initializers"
    "github.com/kart-io/k8s-agent/common/options"
)

// 创建数据库初始化器
dbInit := initializers.NewDatabaseInitializer(
    dbOptions,    // *options.MySQLOptions
    logger,       // core.Logger
)

// （可选）启用自动迁移
dbInit.WithAutoMigrate(
    &types.Agent{},
    &types.Event{},
    &types.Command{},
)

// 注册到 bootstrap
bootstrap.Register(dbInit)

// 获取数据库实例
db := dbInit.DB()  // *gorm.DB
```

### 2. RedisInitializer

Redis 初始化器，负责 Redis 连接初始化。

**功能特性**：

- Redis 连接初始化
- 连接池配置
- 健康检查
- 优雅关闭

**优先级**: `bootstrap.PriorityCache (400)`

**使用示例**：

```go
// 创建 Redis 初始化器
redisInit := initializers.NewRedisInitializer(
    redisOptions,  // *options.RedisOptions
    logger,        // core.Logger
)

// 注册到 bootstrap
bootstrap.Register(redisInit)

// 获取 Redis 客户端
client := redisInit.Client()  // *redis.Client
```

### 3. NATSInitializer

NATS 初始化器，负责 NATS 消息队列连接初始化。

**功能特性**：

- NATS 连接初始化
- 自动重连配置
- 健康检查
- 优雅关闭（Drain）

**优先级**: `bootstrap.PriorityMQ (500)`

**使用示例**：

```go
// 创建 NATS 初始化器
natsInit := initializers.NewNATSInitializer(
    natsOptions,  // *options.NATSOptions
    logger,       // core.Logger
)

// 注册到 bootstrap
bootstrap.Register(natsInit)

// 获取 NATS 连接
conn := natsInit.Connection()  // *nats.Conn

// 或使用便捷方法发布/订阅
natsInit.Publish("subject", data)
natsInit.Subscribe("subject", handler)
```

### 4. HealthCheckInitializer

健康检查初始化器，提供独立的健康检查 HTTP 服务器。

**功能特性**：

- 独立的健康检查服务器
- 提供 `/healthz` 和 `/readyz` 端点
- 最低优先级初始化

**优先级**: `bootstrap.PriorityLowest (1000)`

**使用示例**：

```go
// 创建健康检查初始化器
healthInit := initializers.NewHealthCheckInitializer(
    ":8091",  // 健康检查端口
    logger,
)

// 注册到 bootstrap
bootstrap.Register(healthInit)
```

## 使用指南

### 完整示例

以下是一个完整的服务初始化示例：

```go
package main

import (
    "context"
    "fmt"

    "github.com/kart-io/k8s-agent/common/options"
    "github.com/kart-io/k8s-agent/pkg/bootstrap"
    "github.com/kart-io/k8s-agent/pkg/initializers"
    "github.com/kart-io/logger"
)

type MyServiceApp struct {
    bootstrap *bootstrap.Bootstrap
    logger    core.Logger

    // 初始化器
    dbInit    *initializers.DatabaseInitializer
    redisInit *initializers.RedisInitializer
    natsInit  *initializers.NATSInitializer
}

func NewMyServiceApp() *MyServiceApp {
    // 创建日志
    log := logger.New(logger.Config{
        Engine: logger.EngineZap,
        Level:  logger.LevelInfo,
    })

    return &MyServiceApp{
        bootstrap: bootstrap.New(log),
        logger:    log,
    }
}

func (app *MyServiceApp) Initialize(ctx context.Context) error {
    // 1. 加载配置
    dbOptions := options.NewMySQLOptions()
    redisOptions := options.NewRedisOptions()
    natsOptions := options.NewNATSOptions()

    // 2. 创建并注册通用初始化器

    // 数据库
    app.dbInit = initializers.NewDatabaseInitializer(dbOptions, app.logger)
    app.dbInit.WithAutoMigrate(&MyModel{})
    app.bootstrap.Register(app.dbInit)

    // Redis
    app.redisInit = initializers.NewRedisInitializer(redisOptions, app.logger)
    app.bootstrap.Register(app.redisInit)

    // NATS
    app.natsInit = initializers.NewNATSInitializer(natsOptions, app.logger)
    app.bootstrap.Register(app.natsInit)

    // 3. 注册业务特定初始化器
    // myServiceInit := NewMyServiceInitializer(...)
    // app.bootstrap.Register(myServiceInit)

    // 4. 执行初始化
    return app.bootstrap.Initialize(ctx)
}

func (app *MyServiceApp) Run(ctx context.Context) error {
    return app.bootstrap.Run(ctx, func() error {
        // 主业务逻辑
        app.logger.Info("Service running...")
        <-ctx.Done()
        return nil
    })
}

func main() {
    app := NewMyServiceApp()

    ctx := context.Background()
    if err := app.Initialize(ctx); err != nil {
        panic(fmt.Sprintf("Failed to initialize: %v", err))
    }

    if err := app.Run(ctx); err != nil {
        panic(fmt.Sprintf("Failed to run: %v", err))
    }
}
```

## 最佳实践

### 1. 初始化器生命周期管理

**推荐做法**：

```go
type MyApp struct {
    // 将初始化器保存为字段
    dbInit    *initializers.DatabaseInitializer
    redisInit *initializers.RedisInitializer
}

func (app *MyApp) registerComponents() {
    // 创建并注册
    app.dbInit = initializers.NewDatabaseInitializer(opts, logger)
    app.bootstrap.Register(app.dbInit)

    // 后续可以直接使用
    db := app.dbInit.DB()
}
```

**不推荐做法**：

```go
// ❌ 不要创建后丢弃引用
dbInit := initializers.NewDatabaseInitializer(opts, logger)
bootstrap.Register(dbInit)
// 无法在后续代码中获取 DB 实例
```

### 2. 自动迁移的使用

**开发环境**：

```go
// 开发环境启用自动迁移
if config.Environment == "development" {
    dbInit.WithAutoMigrate(models...)
}
```

**生产环境**：

```go
// 生产环境使用专门的迁移工具
// 不要在生产环境使用 AutoMigrate
```

### 3. 错误处理

**推荐做法**：

```go
// 在 Initialize 时检查错误
if err := app.bootstrap.Initialize(ctx); err != nil {
    app.logger.Errorw("Initialization failed", "error", err)
    return fmt.Errorf("failed to initialize: %w", err)
}
```

### 4. 健康检查

**推荐做法**：

```go
// 使用独立的健康检查端口
healthInit := initializers.NewHealthCheckInitializer(":8091", logger)
bootstrap.Register(healthInit)

// 主服务使用不同端口
// HTTP Server: :8080
// gRPC Server: :9090
// Health Check: :8091
```

## 迁移指南

### 从服务特定初始化器迁移

**迁移前**（Agent-Manager）：

```go
// internal/agent-manager/initializers/database.go
type DatabaseInitializer struct {
    opts  *config.Options
    logger core.Logger
    store *storage.PostgresStore
}

func (d *DatabaseInitializer) Initialize(ctx context.Context) error {
    mysqlClient, err := db.NewMySQLFromOptions(d.logger, d.opts.Database)
    // ... 30 行重复代码
}
```

**迁移后**：

```go
// cmd/agent-manager/app/app.go
import pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"

func (app *AgentManagerApp) registerComponents() {
    // 使用通用初始化器，只需 5 行代码
    app.dbInit = pkginitializers.NewDatabaseInitializer(
        app.opts.Database,
        app.logger,
    ).WithAutoMigrate(
        &types.Agent{},
        &types.Event{},
    )

    app.bootstrap.Register(app.dbInit)
}
```

**收益**：

- 代码减少：从 30 行减少到 5 行
- 维护成本降低：80%
- 一致性：所有服务使用相同的初始化逻辑

### 迁移检查清单

- [ ] 替换 DatabaseInitializer
- [ ] 替换 RedisInitializer
- [ ] 替换 NATSInitializer（如果使用）
- [ ] 更新业务初始化器依赖
- [ ] 更新测试代码
- [ ] 运行全量测试
- [ ] 删除旧的初始化器文件

## 相关文档

- [pkg/bootstrap](../bootstrap/) - 初始化器框架
- [common/options](../../common/options/) - 配置选项
- [common/db](../../common/db/) - 数据库客户端
- [docs/INITIALIZERS_ANALYSIS.md](../../docs/INITIALIZERS_ANALYSIS.md) - 初始化器分析报告

## 贡献指南

如果需要添加新的通用初始化器，请遵循以下步骤：

1. 确保初始化器是通用的，适用于多个服务
2. 实现 `bootstrap.Initializer` 接口
3. 实现 `bootstrap.Closer` 接口（可选）
4. 实现 `bootstrap.HealthChecker` 接口（可选）
5. 添加完整的文档注释
6. 添加使用示例
7. 更新本 README
