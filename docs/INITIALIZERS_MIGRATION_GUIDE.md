# 通用初始化器迁移指南

## 概述

本文档提供了如何将现有服务迁移到使用通用初始化器库 (`pkg/initializers/`) 的详细指南。

## 已完成的迁移

### 1. Agent-Manager 服务 ✅

**迁移前**：
- 独立的 `DatabaseInitializer` (100 行)
- 独立的 `RedisInitializer` (81 行)
- 总计：181 行

**迁移后**：
- 适配器 `DatabaseInitializer` (97 行)
- 适配器 `RedisInitializer` (82 行)
- 总计：179 行

**使用的通用初始化器**：
- `pkg/initializers.DatabaseInitializer`
- `pkg/initializers.RedisInitializer`

**代码示例**：

```go
// internal/agent-manager/initializers/database.go
func NewDatabaseInitializer(opts *config.Options, logger core.Logger) *DatabaseInitializer {
    // 使用通用初始化器
    dbInit := pkginitializers.NewDatabaseInitializer(
        opts.Database,
        logger,
    )

    // 配置自动迁移
    if opts.Database.AutoMigrate {
        dbInit.WithAutoMigrate(
            &types.Agent{},
            &types.Event{},
            // ...
        )
    }

    return &DatabaseInitializer{
        opts:   opts,
        logger: logger,
        dbInit: dbInit,
    }
}
```

### 2. Auth 服务 ✅

**迁移前**：
- 独立的 `DatabaseInitializer` (90 行)
- 独立的 `RedisInitializer` (77 行)
- 总计：167 行

**迁移后**：
- 适配器 `DatabaseInitializer` (75 行)
- 适配器 `RedisInitializer` (69 行)
- 总计：144 行

**使用的通用初始化器**：
- `pkg/initializers.DatabaseInitializer`
- `pkg/initializers.RedisInitializer`

## 待迁移的服务

### 3. Monitor 服务 🔄

**当前状态**：
- 函数式初始化
- 直接在 `NewServer()` 中创建 PostgreSQL 和 Redis
- 文件：`cmd/monitor/app/server.go:45-74`

**使用的基础设施**：
- PostgreSQL (storage.NewPostgresStorage)
- Redis (storage.NewRedisStorage)

**迁移方案**：

```go
// 创建 internal/monitor/initializers/ 目录

// 1. database.go
package initializers

import (
    pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
    "github.com/kart-io/k8s-agent/internal/monitor/storage"
)

type DatabaseInitializer struct {
    dbInit *pkginitializers.DatabaseInitializer
    store  *storage.PostgresStorage
}

func NewDatabaseInitializer(opts *options.DatabaseOptions, logger core.Logger) *DatabaseInitializer {
    dbInit := pkginitializers.NewDatabaseInitializer(opts, logger)
    // 添加 AutoMigrate 如果需要

    return &DatabaseInitializer{
        dbInit: dbInit,
    }
}

func (d *DatabaseInitializer) Initialize(ctx context.Context) error {
    if err := d.dbInit.Initialize(ctx); err != nil {
        return err
    }

    // 创建业务存储层
    d.store = &storage.PostgresStorage{
        DB: d.dbInit.DB(),
    }
    return nil
}

func (d *DatabaseInitializer) Store() *storage.PostgresStorage {
    return d.store
}

// 2. redis.go - 类似模式
```

**修改 `cmd/monitor/app/app.go`**：

```go
type MonitorApp struct {
    bootstrap *bootstrap.Bootstrap
    opts      *config.Options
    logger    core.Logger

    dbInit    *initializers.DatabaseInitializer
    redisInit *initializers.RedisInitializer
}

func (a *MonitorApp) Initialize(ctx context.Context, opts commonapp.Options) error {
    a.opts = opts.(*config.Options)
    a.logger, _ = initLogger(opts)
    a.bootstrap = bootstrap.New(a.logger)

    // 注册初始化器
    a.dbInit = initializers.NewDatabaseInitializer(dbOpts, a.logger)
    a.bootstrap.Register(a.dbInit)

    a.redisInit = initializers.NewRedisInitializer(redisOpts, a.logger)
    a.bootstrap.Register(a.redisInit)

    // 其他初始化...

    return a.bootstrap.Initialize(ctx)
}
```

### 4. Orchestrator 服务 🔄

**当前状态**：
- 最简单的模式
- 直接加载配置，创建 Server
- 文件：`cmd/orchestrator/app/app.go`

**使用的基础设施**：
- PostgreSQL (待确认)
- Redis (storage.NewRedisStore)
- NATS (待确认)

**迁移方案**：

```go
// 改造为使用 RunWithRunner 模式

type OrchestratorApp struct {
    bootstrap *bootstrap.Bootstrap
    opts      *config.Config
    logger    core.Logger

    dbInit    *pkginitializers.DatabaseInitializer
    redisInit *pkginitializers.RedisInitializer
    natsInit  *pkginitializers.NATSInitializer
}

func Execute() {
    opts := config.NewOptions()  // 创建 Options 结构

    commonapp.RunWithRunner(
        opts,
        &OrchestratorApp{},
        initLogger,
        commonapp.CommandConfig{
            Use:   "orchestrator",
            Short: "Orchestrator Service",
        },
    )
}

func (a *OrchestratorApp) Initialize(ctx context.Context, opts commonapp.Options) error {
    a.opts = opts.(*config.Config)
    a.logger, _ = initLogger(opts)
    a.bootstrap = bootstrap.New(a.logger)

    // 直接使用通用初始化器
    a.dbInit = pkginitializers.NewDatabaseInitializer(a.opts.Database, a.logger)
    a.bootstrap.Register(a.dbInit)

    a.redisInit = pkginitializers.NewRedisInitializer(a.opts.Redis, a.logger)
    a.bootstrap.Register(a.redisInit)

    a.natsInit = pkginitializers.NewNATSInitializer(a.opts.NATS, a.logger)
    a.bootstrap.Register(a.natsInit)

    return a.bootstrap.Initialize(ctx)
}

func (a *OrchestratorApp) Run(ctx context.Context) error {
    return a.bootstrap.Run(ctx, nil)
}
```

### 5. Cluster 服务 🔄

**当前状态**：
- 已使用 `RunWithRunner` 模式
- 直接在 Initialize 中创建存储
- 文件：`cmd/cluster/app/app.go:46-86`

**使用的基础设施**：
- MySQL (storage.NewMySQLStorage)

**迁移方案**：

```go
// 只需添加 bootstrap 和初始化器

type ClusterApp struct {
    bootstrap *bootstrap.Bootstrap  // 新增
    opts      *clusterconfig.Options
    logger    core.Logger

    dbInit    *pkginitializers.DatabaseInitializer  // 新增
    storage   *storage.MySQLStorage  // 保留
    server    *api.Server
}

func (a *ClusterApp) Initialize(ctx context.Context, opts commonapp.Options) error {
    a.opts = opts.(*clusterconfig.Options)
    a.logger, _ = initLogger(opts)

    // 创建 bootstrap
    a.bootstrap = bootstrap.New(a.logger)

    // 使用通用数据库初始化器
    a.dbInit = pkginitializers.NewDatabaseInitializer(a.opts.Database, a.logger)
    a.dbInit.WithAutoMigrate(/* models */)
    a.bootstrap.Register(a.dbInit)

    // 初始化
    if err := a.bootstrap.Initialize(ctx); err != nil {
        return err
    }

    // 创建业务存储（使用通用初始化器的 DB）
    a.storage = &storage.MySQLStorage{
        DB: a.dbInit.DB(),
    }

    // 初始化服务
    return a.initializeServices()
}
```

### 6. Gateway 服务 🔄

**待探索**：需要查看当前实现

### 7. Reasoning 服务 🔄

**待探索**：需要查看当前实现

### 8. Collect-Agent 服务 🔄

**待探索**：需要查看当前实现

## 迁移步骤模板

### 步骤 1：评估当前服务

1. 确认服务使用的基础设施
   - [ ] Database (MySQL/PostgreSQL)
   - [ ] Redis
   - [ ] NATS
   - [ ] 其他

2. 查看当前初始化方式
   - 是否已有 `internal/<service>/initializers/` 目录？
   - 是否使用 `bootstrap` 框架？
   - 是否实现了 `Runner` 接口？

### 步骤 2：创建适配器（如果需要）

**场景 A**：服务已有 initializers 目录

```bash
# 修改现有文件，使用通用初始化器
# 参考 Agent-Manager 和 Auth 的模式
```

**场景 B**：服务没有 initializers 目录

```bash
# 创建新目录
mkdir -p internal/<service>/initializers

# 创建适配器文件
# database.go
# redis.go
# nats.go (如果需要)
```

### 步骤 3：修改服务入口

**修改 `cmd/<service>/app/app.go`**：

```go
type ServiceApp struct {
    bootstrap *bootstrap.Bootstrap
    opts      *config.Options
    logger    core.Logger

    // 通用初始化器
    dbInit    *pkginitializers.DatabaseInitializer
    redisInit *pkginitializers.RedisInitializer
    natsInit  *pkginitializers.NATSInitializer

    // 或适配器
    dbInit    *initializers.DatabaseInitializer
    redisInit *initializers.RedisInitializer
}

func (a *ServiceApp) Initialize(ctx context.Context, opts commonapp.Options) error {
    a.opts = opts.(*config.Options)
    a.logger, _ = initLogger(opts)

    // 创建 bootstrap
    a.bootstrap = bootstrap.New(a.logger)

    // 注册通用初始化器
    a.registerComponents()

    // 初始化
    return a.bootstrap.Initialize(ctx)
}

func (a *ServiceApp) registerComponents() {
    // Database
    a.dbInit = pkginitializers.NewDatabaseInitializer(
        a.opts.Database,
        a.logger,
    )
    a.bootstrap.Register(a.dbInit)

    // Redis
    a.redisInit = pkginitializers.NewRedisInitializer(
        a.opts.Redis,
        a.logger,
    )
    a.bootstrap.Register(a.redisInit)

    // NATS
    a.natsInit = pkginitializers.NewNATSInitializer(
        a.opts.NATS,
        a.logger,
    )
    a.bootstrap.Register(a.natsInit)
}

func (a *ServiceApp) Run(ctx context.Context) error {
    return a.bootstrap.Run(ctx, nil)
}

func (a *ServiceApp) Shutdown(ctx context.Context) error {
    return a.bootstrap.Shutdown(ctx)
}
```

### 步骤 4：编译验证

```bash
# 编译服务
go build ./cmd/<service>

# 运行单元测试
go test ./internal/<service>/...

# 验证功能
# 启动服务，测试基本功能
```

### 步骤 5：清理旧代码

```bash
# 如果使用了适配器，删除旧的独立实现
# 如果直接使用通用初始化器，删除旧的初始化代码
```

## 迁移收益

### 代码减少

| 服务 | 迁移前 | 迁移后 | 减少 |
|-----|--------|--------|------|
| Agent-Manager | 181 行 | 179 行 | 2 行 |
| Auth | 167 行 | 144 行 | 23 行 |
| Monitor | ~150 行 | ~80 行 | ~70 行 |
| Orchestrator | ~200 行 | ~100 行 | ~100 行 |
| Cluster | ~100 行 | ~60 行 | ~40 行 |
| **总计** | ~798 行 | ~563 行 | **~235 行 (29%)** |

### 维护性提升

- **统一初始化模式**：所有服务使用相同的初始化流程
- **集中维护**：基础设施变更只需修改 `pkg/initializers/`
- **自动健康检查**：所有初始化器自动实现健康检查
- **优雅关闭**：统一的关闭逻辑

### 开发效率

- **新服务创建**：从 2-3 小时减少到 30 分钟
- **Bug 修复**：从 6 处修改减少到 1 处
- **功能增强**：一次实现，所有服务受益

## 最佳实践

### 1. 使用适配器模式

**何时使用**：
- 服务已有复杂的业务存储层
- 其他组件依赖特定的 `Store()` 方法
- 需要向后兼容

**示例**：
```go
type DatabaseInitializer struct {
    dbInit *pkginitializers.DatabaseInitializer
    store  *storage.BusinessStore
}

func (d *DatabaseInitializer) Store() *storage.BusinessStore {
    return d.store
}
```

### 2. 直接使用通用初始化器

**何时使用**：
- 新服务
- 简单的服务（无复杂业务存储层）
- 可以直接使用 `DB()` 和 `Client()` 方法

**示例**：
```go
type ServiceApp struct {
    dbInit *pkginitializers.DatabaseInitializer
}

func (a *ServiceApp) registerComponents() {
    a.dbInit = pkginitializers.NewDatabaseInitializer(opts, logger)
    a.bootstrap.Register(a.dbInit)
}

func (a *ServiceApp) createServices() {
    db := a.dbInit.DB()
    service := NewService(db)
}
```

### 3. 配置自动迁移

```go
dbInit := pkginitializers.NewDatabaseInitializer(opts, logger)

// 生产环境：不使用 AutoMigrate
if os.Getenv("ENVIRONMENT") == "production" {
    // 使用专门的迁移工具
} else {
    // 开发环境：启用 AutoMigrate
    dbInit.WithAutoMigrate(
        &models.User{},
        &models.Session{},
    )
}
```

### 4. 健康检查集成

```go
// 所有通用初始化器都实现了 HealthChecker 接口
// Bootstrap 会自动注册健康检查

// 在 API 中暴露健康检查端点
func healthCheckHandler(c *gin.Context) {
    if err := bootstrap.HealthCheck(ctx); err != nil {
        c.JSON(500, gin.H{"status": "unhealthy", "error": err.Error()})
        return
    }
    c.JSON(200, gin.H{"status": "healthy"})
}
```

## 常见问题

### Q1: 是否必须使用适配器？

**A**: 不是必须的。如果服务可以直接使用 `DB()` 和 `Client()` 方法，无需适配器。

### Q2: 如何处理自定义初始化逻辑？

**A**: 在适配器的 `Initialize()` 方法中添加：

```go
func (d *DatabaseInitializer) Initialize(ctx context.Context) error {
    // 调用通用初始化器
    if err := d.dbInit.Initialize(ctx); err != nil {
        return err
    }

    // 添加自定义逻辑
    db := d.dbInit.DB()
    // 执行额外的初始化

    return nil
}
```

### Q3: 如何测试初始化器？

**A**: 使用 mock 或测试数据库：

```go
func TestDatabaseInitializer(t *testing.T) {
    opts := &options.DatabaseOptions{
        Host: "localhost",
        Port: 3306,
        // ...
    }
    logger := logger.NewNoOpLogger()

    init := NewDatabaseInitializer(opts, logger)

    ctx := context.Background()
    err := init.Initialize(ctx)
    assert.NoError(t, err)

    db := init.DB()
    assert.NotNil(t, db)
}
```

## 参考资源

- **通用初始化器使用指南**: `pkg/initializers/README.md`
- **初始化器分析报告**: `docs/INITIALIZERS_ANALYSIS.md`
- **代码对比**: `docs/INITIALIZERS_CODE_COMPARISON.md`
- **Bootstrap 框架**: `pkg/bootstrap/README.md`

## 贡献指南

如果您完成了某个服务的迁移：

1. 更新本文档中的迁移状态
2. 添加代码示例
3. 运行测试并记录结果
4. 提交 PR

---

**最后更新**: 2025-10-24
**维护者**: Aetherius Team
