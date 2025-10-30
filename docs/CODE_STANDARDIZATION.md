# 代码规范统一方案

## 文档说明

本文档旨在解决 k8s-agent (Aetherius) 项目中各服务实现不一致的问题，制定统一的架构规范和重构计划。

**创建时间**: 2025-10-30
**状态**: 草案
**维护者**: 开发团队

---

## 目录

- [1. 问题诊断](#1-问题诊断)
- [2. 统一架构规范](#2-统一架构规范)
- [3. 重构计划](#3-重构计划)
- [4. 实施指南](#4-实施指南)

---

## 1. 问题诊断

### 1.1 服务入口架构不一致

#### 问题描述

项目中存在两种不同的服务启动模式：

**模式 A**：使用 `commonapp.RunWithRunner()` + `Application` 接口 + `bootstrap.Bootstrap`
- 使用服务：agent-manager, orchestrator, auth
- 特点：
  - 实现 `commonapp.Application` 接口（Initialize/Run/Shutdown）
  - 使用 `pkg/bootstrap` 进行组件生命周期管理
  - 使用 `initializers/` 包管理依赖初始化
  - 支持优先级排序的组件启动

**模式 B**：使用 `commonapp.RunWithOptions()` + 简单 `run()` 函数
- 使用服务：reasoning, collect-agent
- 特点：
  - 直接使用简单的 `run()` 函数
  - 手动管理服务器生命周期
  - 没有标准化的组件初始化流程

**影响**：
- 新服务开发时需要选择不同的模式，缺乏统一标准
- 代码风格不一致，增加学习成本
- 难以统一添加全局功能（如监控、追踪）

**示例对比**：

```go
// 模式 A (agent-manager)
func Execute() {
    opts := options.NewServerOptions()
    commonapp.RunWithRunner(
        opts,
        &AgentManagerApp{},  // 实现 Application 接口
        initLogger,
        commonapp.CommandConfig{...},
    )
}

// 模式 B (reasoning)
func Execute() {
    opts := config.NewOptions()
    runFunc := func(opts commonapp.Options) error {
        return run(opts.(*config.Options))  // 简单函数
    }
    commonapp.RunWithOptions(opts, runFunc, ...)
}
```

### 1.2 日志系统不一致

#### 问题描述

项目中混用了两套日志系统：

**日志系统 A**：`github.com/kart-io/logger/core` (推荐)
- 使用服务：agent-manager, orchestrator, auth
- 特点：
  - 双引擎支持（Zap/Slog）
  - 统一接口
  - OTLP 集成
  - 性能优化

**日志系统 B**：`github.com/kart-io/k8s-agent/common/logger` (旧版)
- 使用服务：reasoning, collect-agent
- 特点：
  - 基于 Logrus
  - 项目内部封装
  - 功能较简单

**影响**：
- 日志格式不统一
- 无法统一配置日志输出
- 难以实现全局日志追踪
- 旧版日志库性能较低

**示例对比**：

```go
// 日志系统 A (agent-manager)
import "github.com/kart-io/logger/core"

func (a *AgentManagerApp) Initialize(ctx context.Context, opts commonapp.Options) error {
    logger, err := initLogger(opts)  // 返回 core.Logger
    a.logger = logger
    a.logger.Infow("Starting service", "port", 8080)
}

// 日志系统 B (reasoning)
import "github.com/kart-io/k8s-agent/common/logger"

func run(opts *config.Options) error {
    log, err := logger.InitFromOptions(opts.Logging)
    log.Infow("Starting service", "port", 8082)
}
```

### 1.3 数据库访问层不一致

#### 问题描述

不同服务使用了不同的数据库访问模式：

**模式 A**：使用 `common/db.MySQLClient` 封装
- 使用服务：agent-manager
- 特点：
  - 统一的 Options 模式配置
  - 封装了连接池管理
  - 内嵌 `gorm.DB`
  - 自动重连机制

**模式 B**：直接使用 `gorm.DB`
- 使用服务：auth, orchestrator
- 特点：
  - 手动管理 GORM 连接
  - 每个服务自己配置连接池
  - 没有统一的错误处理

**影响**：
- 数据库配置分散
- 连接池参数不统一
- 错误处理方式不同
- 难以统一添加数据库监控

**示例对比**：

```go
// 模式 A (agent-manager/storage/postgres.go)
type PostgresStore struct {
    *db.MySQLClient  // 使用封装的客户端
    logger core.Logger
}

func NewPostgresStore(config types.DatabaseConfig, log core.Logger) (*PostgresStore, error) {
    mysqlClient, err := db.NewMySQL(log,
        db.WithHost(config.Host),
        db.WithPort(config.Port),
        db.WithMaxOpenConns(config.MaxOpenConns),
        // ...更多 Options
    )
    // ...
}

// 模式 B (auth/storage/postgres.go)
type PostgresDB struct {
    DB *gorm.DB  // 直接使用 GORM
}

func NewPostgresDB(cfg *commonoptions.DatabaseOptions) (*PostgresDB, error) {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?...", ...)
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{...})
    sqlDB, _ := db.DB()
    sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
    // ...
}
```

### 1.4 配置结构不一致

#### 问题描述

配置选项的命名和结构存在差异：

**配置类型 A**：`options.ServerOptions`
- 使用服务：agent-manager, orchestrator, auth
- 结构：`cmd/<service>/app/options/options.go`

**配置类型 B**：`config.Options`
- 使用服务：reasoning, collect-agent
- 结构：`internal/<service>/config/config.go`

**影响**：
- 配置加载逻辑不统一
- 难以实现统一的配置验证
- 环境变量命名可能冲突

### 1.5 依赖初始化模式不一致

#### 问题描述

**有 initializers 包的服务**：agent-manager, orchestrator, auth
- 特点：
  - 每个依赖有独立的 Initializer
  - 实现 `pkg/bootstrap.Initializer` 接口
  - 支持优先级排序
  - 统一的错误处理

**无 initializers 包的服务**：reasoning, collect-agent
- 特点：
  - 在 `run()` 函数中手动初始化
  - 依赖顺序不明确
  - 没有统一的生命周期管理

### 1.6 目录结构差异

```
internal/agent-manager/         internal/auth/
├── agent/                      ├── cache/
├── api/                        ├── config/
├── command/                    ├── crypto/
├── config/                     ├── email/
├── initializers/   ✓          ├── filter/
├── storage/        ✓          ├── forced-logout/
└── ...                         ├── handler/
                                ├── initializers/   ✓
internal/reasoning/             ├── jwt/
├── agents/                     ├── logger/         ⚠️ (应该用 common)
├── analyzer/                   ├── metrics/        ⚠️ (应该在 pkg)
├── api/                        ├── middleware/     ⚠️ (应该在 common)
├── config/                     ├── model/
├── llm/                        ├── pagination/     ⚠️ (应该在 common)
├── memory/                     ├── response/       ⚠️ (应该在 common)
└── ...                         ├── storage/        ✓
(无 initializers)  ✗           └── ...
```

**问题总结**：
1. auth 服务包含了大量应该在 `common/` 的通用代码
2. reasoning 和 collect-agent 缺少 initializers 包
3. 目录命名不统一（如 `api/` vs `handler/` vs `routes/`）

---

## 2. 统一架构规范

### 2.1 标准服务架构

#### 2.1.1 服务入口标准（推荐模式）

**必须使用**：`commonapp.RunWithRunner()` + `Application` 接口 + `bootstrap.Bootstrap`

**标准结构**：

```
cmd/<service>/
├── main.go                    # 入口点（最小化）
└── app/
    ├── app.go                 # Application 实现
    ├── server.go              # Server 创建逻辑（可选）
    └── options/
        └── options.go         # 配置选项
```

**标准实现模板**：

```go
// cmd/<service>/main.go
package main

import "github.com/kart-io/k8s-agent/cmd/<service>/app"

func main() {
    app.Execute()
}

// cmd/<service>/app/app.go
package app

import (
    "context"
    "fmt"

    "github.com/kart-io/k8s-agent/cmd/<service>/app/options"
    "github.com/kart-io/k8s-agent/internal/<service>/initializers"
    commonapp "github.com/kart-io/k8s-agent/pkg/app"
    "github.com/kart-io/k8s-agent/pkg/bootstrap"
    pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
    "github.com/kart-io/logger/core"
)

// Execute 运行服务命令
func Execute() {
    opts := options.NewServerOptions()

    commonapp.RunWithRunner(
        opts,
        &ServiceApp{},
        initLogger,
        commonapp.CommandConfig{
            Use:       "<service>",
            Short:     "<Service> Service",
            Long:      "<Service description>",
            EnvPrefix: "<SERVICE>",
        },
    )
}

// ServiceApp 实现 commonapp.Application 接口
type ServiceApp struct {
    bootstrap *bootstrap.Bootstrap
    opts      *options.ServerOptions
    logger    core.Logger

    // 组件初始化器
    dbInit     *initializers.DatabaseInitializer
    redisInit  *initializers.RedisInitializer
    httpInit   *initializers.HTTPServerInitializer
    healthInit *pkginitializers.HealthCheckInitializer
}

// Initialize 初始化应用程序
func (a *ServiceApp) Initialize(ctx context.Context, opts commonapp.Options) error {
    a.opts = opts.(*options.ServerOptions)

    // 初始化日志系统
    logger, err := initLogger(opts)
    if err != nil {
        return fmt.Errorf("failed to initialize logger: %w", err)
    }
    a.logger = logger

    a.logger.Infow("Initializing <Service> Service",
        "http_port", a.opts.Server.Port,
        "health_port", a.opts.Health.Port,
    )

    // 创建 bootstrap 实例
    a.bootstrap = bootstrap.New(a.logger)

    // 注册所有组件初始化器
    a.registerComponents()

    a.logger.Infow("Components registered, ready to start")
    return nil
}

// Run 运行应用程序主逻辑
func (a *ServiceApp) Run(ctx context.Context) error {
    a.logger.Infow("<Service> Service started successfully",
        "http_address", fmt.Sprintf("%s:%d", a.opts.Server.Host, a.opts.Server.Port),
    )

    // 使用 bootstrap 的 Run 方法，它会等待信号
    return a.bootstrap.Run(ctx, nil)
}

// Shutdown 优雅关闭应用程序
func (a *ServiceApp) Shutdown(ctx context.Context) error {
    a.logger.Infow("Shutting down <Service> Service")
    return a.bootstrap.Shutdown(ctx)
}

// registerComponents 注册所有组件初始化器
func (a *ServiceApp) registerComponents() {
    // 按优先级顺序注册组件
    // 1. Database (优先级 300)
    a.dbInit = initializers.NewDatabaseInitializer(a.opts, a.logger)
    a.bootstrap.Register(a.dbInit)

    // 2. Redis (优先级 400)
    a.redisInit = initializers.NewRedisInitializer(a.opts, a.logger)
    a.bootstrap.Register(a.redisInit)

    // 3. HTTP Server (优先级 600)
    a.httpInit = initializers.NewHTTPServerInitializer(a.opts, a.logger, a.dbInit, a.redisInit)
    a.bootstrap.Register(a.httpInit)

    // 4. Health Check Server (优先级最低，最后启动)
    healthPort := a.opts.GetHealthPort()
    healthAddr := fmt.Sprintf(":%d", healthPort)
    a.healthInit = pkginitializers.NewHealthCheckInitializer(healthAddr, a.logger)
    a.bootstrap.Register(a.healthInit)
}

// initLogger 初始化日志系统
func initLogger(opts commonapp.Options) (core.Logger, error) {
    serverOpts := opts.(*options.ServerOptions)
    return serverOpts.InitLogger()
}
```

#### 2.1.2 配置选项标准

**命名规范**：使用 `options.ServerOptions`

**标准位置**：`cmd/<service>/app/options/options.go`

**必须包含的字段**：

```go
package options

import (
    "github.com/kart-io/k8s-agent/common/options"
    "github.com/kart-io/logger/core"
)

// ServerOptions 包含服务的所有配置选项
type ServerOptions struct {
    // 通用选项（来自 common/options）
    Server   options.ServerOptions   `json:"server" mapstructure:"server"`
    Database options.DatabaseOptions `json:"database" mapstructure:"database"`
    Redis    options.RedisOptions    `json:"redis" mapstructure:"redis"`
    Health   options.HealthOptions   `json:"health" mapstructure:"health"`
    Logging  options.LoggingOptions  `json:"logging" mapstructure:"logging"`

    // 服务特定选项
    // ... 添加服务特定配置字段
}

// NewServerOptions 创建默认配置选项
func NewServerOptions() *ServerOptions {
    return &ServerOptions{
        Server:   *options.NewServerOptions(),
        Database: *options.NewDatabaseOptions(),
        Redis:    *options.NewRedisOptions(),
        Health:   *options.NewHealthOptions(),
        Logging:  *options.NewLoggingOptions(),
    }
}

// InitLogger 初始化日志系统
func (o *ServerOptions) InitLogger() (core.Logger, error) {
    return o.Logging.InitLogger()
}

// GetHealthPort 获取健康检查端口
func (o *ServerOptions) GetHealthPort() int {
    if o.Health.Port > 0 {
        return o.Health.Port
    }
    return o.Server.Port + 1000
}

// Config 转换为业务配置
func (o *ServerOptions) Config() (*Config, error) {
    // 转换逻辑
}
```

#### 2.1.3 日志系统标准

**必须使用**：`github.com/kart-io/logger/core.Logger`

**禁止使用**：
- ❌ `github.com/kart-io/k8s-agent/common/logger`
- ❌ `logrus`
- ❌ 标准库 `log`

**标准用法**：

```go
import "github.com/kart-io/logger/core"

// 1. 在 Application 中存储 logger
type ServiceApp struct {
    logger core.Logger
}

// 2. 通过 InitLogger 创建
logger, err := serverOpts.InitLogger()

// 3. 使用结构化日志
logger.Infow("Operation completed",
    "user_id", userID,
    "duration", duration,
    "status", "success",
)

// 4. 创建子 logger（带固定字段）
componentLogger := logger.With("component", "database")
componentLogger.Infow("Connection established", "host", host)

// 5. 错误日志
logger.Errorw("Failed to process request",
    "error", err,
    "request_id", reqID,
)
```

#### 2.1.4 数据库访问标准

**必须使用**：`common/db.MySQLClient` 封装

**标准实现**：

```go
// internal/<service>/storage/postgres.go
package storage

import (
    "context"

    "github.com/kart-io/logger/core"
    "github.com/kart-io/k8s-agent/common/db"
    "github.com/kart-io/k8s-agent/pkg/types"
)

// PostgresStore implements storage using MySQL
type PostgresStore struct {
    *db.MySQLClient  // 必须嵌入封装的客户端
    logger          core.Logger
}

// NewPostgresStore 创建新的存储实例
func NewPostgresStore(config types.DatabaseConfig, log core.Logger) (*PostgresStore, error) {
    // 使用 Options 模式创建 MySQL 客户端
    mysqlClient, err := db.NewMySQL(log,
        db.WithHost(config.Host),
        db.WithPort(config.Port),
        db.WithUser(config.User),
        db.WithPassword(config.Password),
        db.WithDatabase(config.Database),
        db.WithMaxOpenConns(config.MaxOpenConns),
        db.WithMaxIdleConns(config.MaxIdleConns),
        db.WithConnMaxLifetime(config.ConnMaxLifetime),
        db.WithLogLevel("info"),
    )
    if err != nil {
        return nil, err
    }

    store := &PostgresStore{
        MySQLClient: mysqlClient,
        logger:      log.With("component", "storage"),
    }

    // 自动迁移模型
    if err := store.AutoMigrate(
        &types.Model1{},
        &types.Model2{},
    ); err != nil {
        return nil, err
    }

    store.logger.Infow("PostgreSQL store initialized",
        "host", config.Host,
        "database", config.Database)

    return store, nil
}

// 业务方法示例
func (s *PostgresStore) GetByID(ctx context.Context, id string) (*types.Model, error) {
    var model types.Model
    if err := s.DB.WithContext(ctx).First(&model, "id = ?", id).Error; err != nil {
        return nil, err
    }
    return &model, nil
}
```

#### 2.1.5 Initializers 标准

**必须实现**：每个服务都必须有 `internal/<service>/initializers/` 包

**标准结构**：

```
internal/<service>/initializers/
├── database.go      # 数据库初始化器
├── redis.go         # Redis 初始化器
├── nats.go          # NATS 初始化器（如需要）
├── servers.go       # HTTP/gRPC 服务器初始化器
└── services.go      # 业务服务初始化器
```

**标准实现模板**：

```go
// internal/<service>/initializers/database.go
package initializers

import (
    "context"
    "fmt"

    "github.com/kart-io/logger/core"
    "github.com/kart-io/k8s-agent/cmd/<service>/app/options"
    "github.com/kart-io/k8s-agent/internal/<service>/storage"
    "github.com/kart-io/k8s-agent/pkg/types"
)

// DatabaseInitializer 数据库初始化器
type DatabaseInitializer struct {
    opts   *options.ServerOptions
    logger core.Logger
    store  *storage.PostgresStore
}

// NewDatabaseInitializer 创建数据库初始化器
func NewDatabaseInitializer(opts *options.ServerOptions, logger core.Logger) *DatabaseInitializer {
    return &DatabaseInitializer{
        opts:   opts,
        logger: logger.With("initializer", "database"),
    }
}

// Name 返回初始化器名称
func (i *DatabaseInitializer) Name() string {
    return "database"
}

// Priority 返回初始化优先级（数字越小优先级越高）
func (i *DatabaseInitializer) Priority() int {
    return 300
}

// Initialize 执行初始化
func (i *DatabaseInitializer) Initialize(ctx context.Context) error {
    i.logger.Infow("Initializing database connection")

    dbConfig := types.DatabaseConfig{
        Host:            i.opts.Database.Host,
        Port:            i.opts.Database.Port,
        User:            i.opts.Database.User,
        Password:        i.opts.Database.Password,
        Database:        i.opts.Database.Database,
        MaxOpenConns:    i.opts.Database.MaxOpenConns,
        MaxIdleConns:    i.opts.Database.MaxIdleConns,
        ConnMaxLifetime: i.opts.Database.ConnMaxLifetime,
    }

    store, err := storage.NewPostgresStore(dbConfig, i.logger)
    if err != nil {
        return fmt.Errorf("failed to create store: %w", err)
    }

    i.store = store
    i.logger.Infow("Database initialized successfully")
    return nil
}

// Shutdown 执行关闭
func (i *DatabaseInitializer) Shutdown(ctx context.Context) error {
    i.logger.Infow("Shutting down database connection")
    if i.store != nil {
        return i.store.Close()
    }
    return nil
}

// GetStore 获取存储实例
func (i *DatabaseInitializer) GetStore() *storage.PostgresStore {
    return i.store
}
```

#### 2.1.6 标准目录结构

```
internal/<service>/
├── config/                    # 配置相关（如果需要）
│   └── config.go              # 业务配置结构
├── initializers/              # 必须：组件初始化器
│   ├── database.go
│   ├── redis.go
│   ├── nats.go                # 如需要
│   ├── servers.go
│   └── services.go
├── storage/                   # 必须：数据访问层
│   ├── postgres.go            # 主数据库访问
│   ├── redis.go               # Redis 缓存访问
│   └── types.go               # 存储层类型
├── api/                       # HTTP API 处理器
│   ├── handlers.go
│   ├── routes.go
│   └── middleware.go          # 服务特定中间件
├── grpc/                      # gRPC 服务实现（如需要）
│   └── service.go
├── service/                   # 业务逻辑层
│   └── service.go
├── types/                     # 服务特定类型
│   └── types.go
└── config.go                  # 配置转换（Options → Config）
```

**禁止出现的目录**（这些应该在 `common/` 或 `pkg/`）：
- ❌ `logger/` - 应该使用 `github.com/kart-io/logger`
- ❌ `middleware/` - 通用中间件应该在 `common/middleware`
- ❌ `response/` - 应该使用 `common/response`
- ❌ `pagination/` - 应该使用 `common/pagination`
- ❌ `metrics/` - 应该使用 `pkg/metrics`

### 2.2 代码组织规范

#### 2.2.1 common/ 包规范

**定位**：通用、可复用的基础功能库，任何 Go 项目都可以使用

**允许的包**：
- ✅ `cache/` - 统一缓存接口
- ✅ `client/` - 通用客户端封装
- ✅ `config/` - Options 模式配置
- ✅ `db/` - 数据库客户端封装
- ✅ `errors/` - 通用错误处理
- ✅ `middleware/` - 通用中间件
- ✅ `mq/` - 消息队列客户端
- ✅ `pagination/` - 分页工具
- ✅ `response/` - 统一响应格式
- ✅ `server/` - HTTP/gRPC 服务器封装
- ✅ `utils/` - 通用工具函数
- ✅ `validator/` - 通用数据验证

**禁止内容**：
- ❌ 业务逻辑
- ❌ 领域模型
- ❌ 项目特定类型

#### 2.2.2 pkg/ 包规范

**定位**：Aetherius 项目特定的业务逻辑和领域模型

**标准包**：
- ✅ `bootstrap/` - 应用启动框架
- ✅ `app/` - 应用框架
- ✅ `contextx/` - 上下文增强
- ✅ `idempotent/` - 幂等性处理
- ✅ `metrics/` - 项目指标
- ✅ `initializers/` - 通用初始化器（如 HealthCheck）
- ✅ `types/` - 业务领域模型
- ✅ `k8s/` - Kubernetes 业务逻辑
- ✅ `agent/` - Agent 领域模型
- ✅ `workflow/` - 工作流业务逻辑

#### 2.2.3 internal/<service>/ 规范

**定位**：服务私有实现

**必须包含**：
- ✅ `initializers/` - 服务组件初始化器
- ✅ `storage/` - 数据访问层
- ✅ `api/` 或 `handler/` - API 处理器

**可选包含**：
- `grpc/` - gRPC 服务实现
- `service/` - 业务逻辑层
- `types/` - 服务特定类型
- `config/` - 配置转换

### 2.3 错误处理规范

**必须使用**：`common/errors` 包

```go
import "github.com/kart-io/k8s-agent/common/errors"

// 创建错误
if invalid {
    return errors.New(errors.CodeInvalidArgument, "validation failed")
}

// 包装错误
if err := doSomething(); err != nil {
    return errors.Wrap(err, errors.CodeInternal, "operation failed")
}

// 错误码定义
const (
    CodeOK                 = 0
    CodeInternal           = 1
    CodeInvalidArgument    = 2
    CodeNotFound           = 3
    CodeAlreadyExists      = 4
    CodePermissionDenied   = 5
    CodeUnavailable        = 6
)
```

### 2.4 API 响应规范

**必须使用**：`common/response` 包

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/kart-io/k8s-agent/common/response"
)

// 成功响应
response.Success(c, data)

// 错误响应
response.Error(c, http.StatusBadRequest, "invalid input")
response.NotFound(c, "resource not found")
response.InternalError(c, err)

// 分页响应
response.SuccessWithPagination(c, data, pagination.Info{
    Page: 1, PageSize: 20, Total: 100,
})
```

---

## 3. 重构计划

### 3.1 重构优先级

#### P0 - 紧急（影响开发效率）

**任务 1**：统一日志系统
- **目标**：所有服务迁移到 `github.com/kart-io/logger`
- **影响服务**：reasoning, collect-agent
- **预计工时**：2-3 天
- **风险**：低

**任务 2**：统一服务入口架构
- **目标**：reasoning, collect-agent 迁移到 Application 接口
- **影响服务**：reasoning, collect-agent
- **预计工时**：3-5 天
- **风险**：中

#### P1 - 高优先级（影响代码质量）

**任务 3**：统一数据库访问层
- **目标**：auth, orchestrator 迁移到 `common/db.MySQLClient`
- **影响服务**：auth, orchestrator
- **预计工时**：3-4 天
- **风险**：中

**任务 4**：补充 initializers 包
- **目标**：为 reasoning, collect-agent 添加 initializers
- **影响服务**：reasoning, collect-agent
- **预计工时**：2-3 天
- **风险**：低

**任务 5**：清理 auth 服务重复代码
- **目标**：移除 auth 中的 logger, middleware, response, pagination 等
- **影响服务**：auth
- **预计工时**：2-3 天
- **风险**：低

#### P2 - 中优先级（改进可维护性）

**任务 6**：统一配置选项命名
- **目标**：所有服务使用 `options.ServerOptions`
- **影响服务**：reasoning, collect-agent
- **预计工时**：1-2 天
- **风险**：低

**任务 7**：标准化目录结构
- **目标**：统一 API 处理器命名（api/ vs handler/）
- **影响服务**：所有服务
- **预计工时**：1-2 天
- **风险**：低

### 3.2 分阶段实施计划

#### 阶段 1：日志和入口统一（Week 1-2）

**目标**：解决最紧迫的不一致问题

**任务列表**：
1. reasoning 服务迁移到 `kart-io/logger`
2. collect-agent 服务迁移到 `kart-io/logger`
3. reasoning 服务迁移到 Application 接口
4. collect-agent 服务迁移到 Application 接口

**验收标准**：
- ✅ 所有服务使用统一的日志系统
- ✅ 所有服务使用统一的启动框架
- ✅ 所有服务支持优雅关闭
- ✅ 日志格式统一

#### 阶段 2：数据访问层统一（Week 3）

**目标**：统一数据库访问方式

**任务列表**：
1. auth 服务迁移到 `common/db.MySQLClient`
2. orchestrator 服务迁移到 `common/db.MySQLClient`
3. 验证连接池配置统一

**验收标准**：
- ✅ 所有服务使用统一的数据库客户端
- ✅ 连接池参数统一配置
- ✅ 数据库日志格式统一

#### 阶段 3：补充初始化器（Week 4）

**目标**：完善组件生命周期管理

**任务列表**：
1. reasoning 添加 initializers 包
2. collect-agent 添加 initializers 包
3. 验证启动顺序和依赖管理

**验收标准**：
- ✅ 所有服务有 initializers 包
- ✅ 组件启动优先级明确
- ✅ 支持按顺序关闭

#### 阶段 4：代码清理（Week 5）

**目标**：移除重复代码，提升复用性

**任务列表**：
1. auth 服务移除 logger, middleware, response, pagination
2. 验证所有服务使用 common 包
3. 更新导入路径

**验收标准**：
- ✅ 没有重复的通用代码
- ✅ common 包使用率 > 90%
- ✅ 代码行数减少 > 20%

#### 阶段 5：配置和目录标准化（Week 6）

**目标**：完善代码组织

**任务列表**：
1. 统一配置选项命名
2. 统一目录结构
3. 更新文档

**验收标准**：
- ✅ 所有服务目录结构一致
- ✅ 配置选项命名统一
- ✅ 文档完整准确

### 3.3 回归测试计划

每个阶段完成后必须执行：

```bash
# 1. 单元测试
make test

# 2. 集成测试
make test-integration

# 3. 构建验证
make build

# 4. 服务启动验证
make docker-compose-up
# 验证所有服务正常启动

# 5. API 测试
make test-endpoints
```

---

## 4. 实施指南

### 4.1 重构 Reasoning 服务示例

#### 步骤 1：迁移日志系统

```bash
# 1. 更新导入
find internal/reasoning -name "*.go" -exec sed -i \
  's|github.com/kart-io/k8s-agent/common/logger|github.com/kart-io/logger/core|g' {} \;

# 2. 更新类型声明
# logger.Logger → core.Logger
```

#### 步骤 2：创建 Application 实现

创建 `cmd/reasoning/app/app.go`：

```go
package app

import (
    "context"
    "fmt"

    "github.com/kart-io/k8s-agent/cmd/reasoning/app/options"
    "github.com/kart-io/k8s-agent/internal/reasoning/initializers"
    commonapp "github.com/kart-io/k8s-agent/pkg/app"
    "github.com/kart-io/k8s-agent/pkg/bootstrap"
    pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
    "github.com/kart-io/logger/core"
)

func Execute() {
    opts := options.NewServerOptions()

    commonapp.RunWithRunner(
        opts,
        &ReasoningApp{},
        initLogger,
        commonapp.CommandConfig{
            Use:       "reasoning",
            Short:     "Reasoning Service",
            Long:      "Reasoning Service provides AI-driven root cause analysis",
            EnvPrefix: "REASONING",
        },
    )
}

type ReasoningApp struct {
    bootstrap *bootstrap.Bootstrap
    opts      *options.ServerOptions
    logger    core.Logger

    httpInit   *initializers.HTTPServerInitializer
    llmInit    *initializers.LLMInitializer
    healthInit *pkginitializers.HealthCheckInitializer
}

// 实现 Initialize, Run, Shutdown, registerComponents 方法...
```

#### 步骤 3：创建 Initializers

创建 `internal/reasoning/initializers/llm.go`：

```go
package initializers

import (
    "context"
    "fmt"

    "github.com/kart-io/logger/core"
    "github.com/kart-io/k8s-agent/cmd/reasoning/app/options"
    "github.com/kart-io/k8s-agent/internal/reasoning/llm"
)

type LLMInitializer struct {
    opts   *options.ServerOptions
    logger core.Logger
    client *llm.Client
}

func NewLLMInitializer(opts *options.ServerOptions, logger core.Logger) *LLMInitializer {
    return &LLMInitializer{
        opts:   opts,
        logger: logger.With("initializer", "llm"),
    }
}

func (i *LLMInitializer) Name() string {
    return "llm"
}

func (i *LLMInitializer) Priority() int {
    return 400
}

func (i *LLMInitializer) Initialize(ctx context.Context) error {
    i.logger.Infow("Initializing LLM client")

    if !i.opts.LLM.Enabled {
        i.logger.Infow("LLM disabled, skipping initialization")
        return nil
    }

    client, err := llm.NewClient(i.opts.LLM, i.logger)
    if err != nil {
        return fmt.Errorf("failed to create LLM client: %w", err)
    }

    i.client = client
    i.logger.Infow("LLM client initialized successfully")
    return nil
}

func (i *LLMInitializer) Shutdown(ctx context.Context) error {
    i.logger.Infow("Shutting down LLM client")
    if i.client != nil {
        return i.client.Close()
    }
    return nil
}

func (i *LLMInitializer) GetClient() *llm.Client {
    return i.client
}
```

### 4.2 验证检查清单

每次重构完成后，使用此清单验证：

#### 代码规范检查

- [ ] 使用 `commonapp.RunWithRunner()` + `Application` 接口
- [ ] 实现 `Initialize()`, `Run()`, `Shutdown()` 方法
- [ ] 使用 `github.com/kart-io/logger/core.Logger`
- [ ] 配置选项命名为 `options.ServerOptions`
- [ ] 数据库访问使用 `common/db.MySQLClient`
- [ ] 有 `internal/<service>/initializers/` 包
- [ ] 所有初始化器实现 `bootstrap.Initializer` 接口
- [ ] HTTP 响应使用 `common/response`
- [ ] 错误处理使用 `common/errors`

#### 功能测试检查

- [ ] 服务能正常启动
- [ ] 健康检查端点工作正常 (`/health`)
- [ ] 优雅关闭功能正常（SIGTERM/SIGINT）
- [ ] 日志输出格式正确
- [ ] 所有 API 端点正常工作
- [ ] 数据库连接正常
- [ ] Redis 连接正常（如有）
- [ ] NATS 连接正常（如有）

#### 测试覆盖检查

- [ ] 单元测试通过 (`make test`)
- [ ] 集成测试通过 (`make test-integration`)
- [ ] 构建成功 (`make build`)
- [ ] Docker 镜像构建成功 (`make docker-build`)
- [ ] 代码覆盖率 > 60%

### 4.3 常见问题和解决方案

#### Q1: 如何处理服务特定的初始化逻辑？

**A**: 创建服务特定的 Initializer

```go
// internal/<service>/initializers/custom.go
type CustomInitializer struct {
    opts   *options.ServerOptions
    logger core.Logger
    // 服务特定字段
}

func NewCustomInitializer(...) *CustomInitializer { ... }

func (i *CustomInitializer) Name() string { return "custom" }
func (i *CustomInitializer) Priority() int { return 500 }
func (i *CustomInitializer) Initialize(ctx context.Context) error { ... }
func (i *CustomInitializer) Shutdown(ctx context.Context) error { ... }
```

#### Q2: 如何处理配置的环境变量覆盖？

**A**: 使用 Viper 的自动环境变量支持

```go
// cmd/<service>/app/options/options.go
import "github.com/spf13/viper"

func (o *ServerOptions) LoadConfig(configPath string) error {
    v := viper.New()
    v.SetConfigFile(configPath)
    v.AutomaticEnv()  // 自动读取环境变量
    v.SetEnvPrefix("SERVICE_NAME")  // 设置前缀

    if err := v.ReadInConfig(); err != nil {
        return err
    }

    return v.Unmarshal(o)
}
```

#### Q3: 如何在不影响功能的情况下重构？

**A**: 采用渐进式重构策略

1. **先添加后删除**：先实现新模式，保留旧代码，验证后再删除
2. **功能开关**：使用环境变量控制新旧实现
3. **充分测试**：每次修改后运行完整测试套件
4. **小步快跑**：每次只重构一个服务或模块

#### Q4: 如何处理依赖注入？

**A**: 通过 Initializer 的 Getter 方法

```go
// 在 registerComponents 中传递依赖
a.httpInit = initializers.NewHTTPServerInitializer(
    a.opts,
    a.logger,
    a.dbInit,      // 传递数据库初始化器
    a.redisInit,   // 传递 Redis 初始化器
)

// 在 HTTPServerInitializer 中使用
func (i *HTTPServerInitializer) Initialize(ctx context.Context) error {
    db := i.dbInit.GetStore()  // 获取数据库实例
    redis := i.redisInit.GetClient()  // 获取 Redis 客户端
    // 使用依赖...
}
```

---

## 附录

### A. 参考文档

- [CLAUDE.md](../CLAUDE.md) - 项目总体指南
- [CODE_REORGANIZATION.md](CODE_REORGANIZATION.md) - 代码重组计划
- [pkg/bootstrap/README.md](../pkg/bootstrap/README.md) - Bootstrap 框架文档
- [common/db/README.md](../common/db/README.md) - 数据库客户端文档

### B. 联系方式

如有问题，请联系：
- 架构负责人：[待定]
- 开发团队：[待定]

---

**文档版本**: v1.0
**最后更新**: 2025-10-30
