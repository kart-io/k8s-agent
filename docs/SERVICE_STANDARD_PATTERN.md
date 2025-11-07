# 服务标准实现模式（综合最佳实践）

> 综合 agent-manager 和 auth 的优点，定义统一的服务实现标准

---

## 📋 执行摘要

本文档定义了 Aetherius 项目中所有服务必须遵循的**统一实现模式**，该模式综合了 `agent-manager` 和 `auth` 服务的最佳实践：

- ✅ **从 agent-manager 继承**：Application 接口 + Bootstrap 框架（完整生命周期管理）
- ✅ **从 auth 继承**：Options 目录结构 + Config 转换模式（清晰的分层）
- ✅ **新增改进**：更清晰的职责分离、更好的可测试性

---

## 🎯 核心原则

### 1. 三层分离原则

```
┌─────────────────────────────────────────────────────────────┐
│  启动层 (cmd/<service>/app/)                                │
│  - options/options.go: 命令行配置（实现 commonapp.Options） │
│  - app.go: Application 实现（生命周期管理）                  │
│  - server.go: 辅助函数（可选）                               │
└─────────────────────────────────────────────────────────────┘
                            ↓ Config() 转换
┌─────────────────────────────────────────────────────────────┐
│  业务层 (internal/<service>/)                               │
│  - config.go: 业务配置结构                                   │
│  - server.go: 业务服务器实现                                 │
└─────────────────────────────────────────────────────────────┘
                            ↓ 使用
┌─────────────────────────────────────────────────────────────┐
│  组件层 (internal/<service>/initializers/)                  │
│  - database.go, redis.go: 组件初始化器                       │
│  - servers.go, services.go: 服务器和业务服务初始化器         │
└─────────────────────────────────────────────────────────────┘
```

### 2. 职责分离

| 层级 | 位置 | 职责 | 依赖 |
|------|------|------|------|
| **启动层** | `cmd/<service>/app/` | 命令行参数、配置加载、应用启动 | pkg/app, common/options |
| **业务层** | `internal/<service>/` | 业务逻辑、领域模型、服务实现 | common/options |
| **组件层** | `internal/<service>/initializers/` | 组件初始化、依赖管理 | pkg/bootstrap |

---

## 📐 标准目录结构

```
<service>/
├── cmd/<service>/
│   ├── main.go                          # 入口点（极简）
│   └── app/
│       ├── options/
│       │   └── options.go               # Options 配置（实现 commonapp.Options）
│       ├── app.go                       # Application 实现（生命周期管理）
│       └── server.go                    # 辅助函数（可选，如 initLogger）
│
├── internal/<service>/
│   ├── config.go                        # 业务配置结构 + NewServer
│   ├── server.go                        # 业务服务器实现（可选，可合并到 config.go）
│   ├── api/
│   │   ├── handlers/                    # HTTP 处理器
│   │   └── routes.go                    # 路由注册
│   ├── initializers/
│   │   ├── database.go                  # 数据库初始化器
│   │   ├── redis.go                     # Redis 初始化器
│   │   ├── servers.go                   # HTTP/gRPC 服务器初始化器
│   │   └── services.go                  # 业务服务初始化器
│   ├── service/                         # 业务逻辑层
│   │   └── <domain>.go
│   ├── storage/                         # 持久化层
│   │   ├── repository.go
│   │   └── models.go
│   └── <domain>/                        # 领域特定逻辑
│
└── configs/
    └── config.yaml                      # 配置文件示例
```

---

## 💻 标准实现模板

### 1. 入口点 (`cmd/<service>/main.go`)

**极简实现，只负责调用 app.Execute()**

```go
package main

import (
	// 自动配置 GOMAXPROCS 以匹配容器 CPU 配额
	_ "go.uber.org/automaxprocs/maxprocs"

	"github.com/kart-io/k8s-agent/cmd/<service>/app"
)

func main() {
	app.Execute()
}
```

**要点**:
- ✅ 导入 automaxprocs（容器环境优化）
- ✅ 只调用 `app.Execute()`
- ❌ 不包含任何业务逻辑

---

### 2. Options 配置 (`cmd/<service>/app/options/options.go`)

**实现 commonapp.Options 接口，提供辅助方法**

```go
package options

import (
	"github.com/spf13/pflag"

	commonlogger "github.com/kart-io/k8s-agent/common/logger"
	commonoptions "github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/k8s-agent/internal/<service>"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/logger/core"
)

const (
	// UserAgent is the user agent name when starting the service.
	UserAgent = "aetherius-<service>"
)

// ServerOptions contains the configuration options for the <service> server.
type ServerOptions struct {
	// 基础配置（必需）
	Server   *commonoptions.ServerOptions   `json:"server" mapstructure:"server"`
	Database *commonoptions.MySQLOptions `json:"database" mapstructure:"database"`
	Redis    *commonoptions.RedisOptions    `json:"redis" mapstructure:"redis"`
	Logging  *commonoptions.LoggingOptions  `json:"logging" mapstructure:"logging"`
	Health   *commonoptions.HealthOptions   `json:"health" mapstructure:"health"`

	// 可选配置（根据服务需求选择）
	GRPC    *commonoptions.GRPCOptions    `json:"grpc,omitempty" mapstructure:"grpc"`
	NATS    *commonoptions.NATSOptions    `json:"nats,omitempty" mapstructure:"nats"`
	JWT     *commonoptions.JWTOptions     `json:"jwt,omitempty" mapstructure:"jwt"`
	Email   *commonoptions.EmailOptions   `json:"email,omitempty" mapstructure:"email"`
	Metrics *commonoptions.MetricsOptions `json:"metrics,omitempty" mapstructure:"metrics"`

	// 服务特定配置
	// Feature *FeatureOptions `json:"feature,omitempty" mapstructure:"feature"`
}

// 确保实现了 commonapp.Options 接口
var _ commonapp.Options = (*ServerOptions)(nil)

// NewServerOptions creates a ServerOptions instance with default values.
func NewServerOptions() *ServerOptions {
	healthOpts := commonoptions.NewHealthOptions()
	healthOpts.Port = 8090 // 修改为服务特定端口: 8090(auth), 8091(agent-manager), 8092(orchestrator), 8093(reasoning)

	return &ServerOptions{
		Server:   commonoptions.NewServerOptions(),
		Database: commonoptions.NewMySQLOptions(),
		Redis:    commonoptions.NewRedisOptions(),
		Logging:  commonoptions.NewLoggingOptions(),
		Health:   healthOpts,
		// 根据需要初始化其他选项
		// GRPC:    commonoptions.NewGRPCOptions(),
		// NATS:    commonoptions.NewNATSOptions(),
		// JWT:     commonoptions.NewJWTOptions(),
	}
}

// GetHealthPort 实现 commonapp.HealthPortProvider 接口
func (o *ServerOptions) GetHealthPort() int {
	if o.Health != nil {
		return o.Health.Port
	}
	return 8090 // 默认端口，与 NewServerOptions 保持一致
}

// AddFlags adds flags to the specified FlagSet.
func (o *ServerOptions) AddFlags(fs *pflag.FlagSet) {
	// 添加所有子选项的 flags
	o.Server.AddFlags(fs)
	o.Database.AddFlags(fs)
	o.Redis.AddFlags(fs)
	o.Logging.AddFlags(fs)
	if o.Health != nil {
		o.Health.AddFlags(fs, "")
	}

	// 根据需要添加其他选项
	// if o.GRPC != nil {
	//     o.GRPC.AddFlags(fs)
	// }
	// if o.NATS != nil {
	//     o.NATS.AddFlags(fs)
	// }
}

// Complete completes all the required options.
func (o *ServerOptions) Complete() error {
	// 设置服务名称到日志初始字段
	if o.Logging.InitialFields == nil {
		o.Logging.InitialFields = make(map[string]interface{})
	}
	if _, ok := o.Logging.InitialFields["service.name"]; !ok {
		o.Logging.InitialFields["service.name"] = UserAgent
	}

	// 完成所有子选项
	if err := o.Server.Complete(); err != nil {
		return err
	}

	if err := o.Database.Complete(); err != nil {
		return err
	}

	if err := o.Redis.Complete(); err != nil {
		return err
	}

	if err := o.Logging.Complete(); err != nil {
		return err
	}

	if o.Health != nil {
		if err := o.Health.Complete(); err != nil {
			return err
		}
	}

	// 根据需要完成其他选项
	// if o.GRPC != nil {
	//     if err := o.GRPC.Complete(); err != nil {
	//         return err
	//     }
	// }

	return nil
}

// Validate checks whether the options are valid.
func (o *ServerOptions) Validate() []error {
	var errs []error

	if err := o.Server.Validate(); err != nil {
		errs = append(errs, err)
	}

	if err := o.Database.Validate(); err != nil {
		errs = append(errs, err)
	}

	if err := o.Redis.Validate(); err != nil {
		errs = append(errs, err)
	}

	if err := o.Logging.Validate(); err != nil {
		errs = append(errs, err)
	}

	if o.Health != nil {
		if err := o.Health.Validate(); err != nil {
			errs = append(errs, err)
		}
	}

	// 根据需要验证其他选项

	return errs
}

// Config builds a <service>.Config based on ServerOptions.
// 这是将启动层配置转换为业务层配置的关键方法
func (o *ServerOptions) Config() (*<service>.Config, error) {
	return &<service>.Config{
		Server:   o.Server,
		Database: o.Database,
		Redis:    o.Redis,
		Logging:  o.Logging,
		// 根据需要添加其他配置
		// GRPC:    o.GRPC,
		// NATS:    o.NATS,
	}, nil
}

// InitLogger initializes logger based on the options.
func (o *ServerOptions) InitLogger() (core.Logger, error) {
	return commonlogger.InitFromOptions(o.Logging)
}
```

**要点**:
- ✅ 实现 `commonapp.Options` 接口（AddFlags, Complete, Validate）
- ✅ 提供 `Config()` 方法（转换为业务配置）
- ✅ 提供 `InitLogger()` 方法（初始化日志）
- ✅ 提供 `GetHealthPort()` 方法（获取健康检查端口）
- ✅ 使用 `common/options` 标准配置
- ✅ 设置服务名称到日志字段

---

### 3. Application 实现 (`cmd/<service>/app/app.go`)

**实现 Application 接口，管理完整生命周期**

```go
package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/cmd/<service>/app/options"
	"github.com/kart-io/k8s-agent/internal/<service>"
	"github.com/kart-io/k8s-agent/internal/<service>/initializers"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// Execute runs the <service> command
func Execute() {
	// 创建 Options 实例
	opts := options.NewServerOptions()

	// 使用 RunWithRunner 启动应用
	commonapp.RunWithRunner(
		opts,
		&<Service>App{},
		initLogger,
		commonapp.CommandConfig{
			Use:       "<service>",
			Short:     "<Service> Service",
			Long:      "<Service> Service provides ...",
			EnvPrefix: "<SERVICE>",
		},
	)
}

// <Service>App 实现 commonapp.Application 接口
type <Service>App struct {
	bootstrap *bootstrap.Bootstrap
	opts      *options.ServerOptions
	logger    core.Logger
	config    *<service>.Config

	// 组件初始化器
	dbInit      *initializers.DatabaseInitializer
	redisInit   *initializers.RedisInitializer
	serviceInit *initializers.ServiceInitializer
	httpInit    *initializers.HTTPServerInitializer
	healthInit  *pkginitializers.HealthCheckInitializer
	// 根据需要添加其他初始化器
	// grpcInit    *initializers.GRPCServerInitializer
	// natsInit    *initializers.NATSInitializer
}

// 确保实现了 Application 接口
var _ commonapp.Application = (*<Service>App)(nil)

// Initialize 初始化应用程序
func (a *<Service>App) Initialize(ctx context.Context, opts commonapp.Options) error {
	// 类型断言
	a.opts = opts.(*options.ServerOptions)

	// 初始化日志
	logger, err := initLogger(opts)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	a.logger = logger

	a.logger.Infow("Initializing <Service> Service",
		"http_port", a.opts.Server.Port,
		"health_port", a.opts.Health.Port,
	)

	// 转换为业务配置
	config, err := a.opts.Config()
	if err != nil {
		return fmt.Errorf("failed to build config: %w", err)
	}
	a.config = config

	// 创建 Bootstrap 实例
	a.bootstrap = bootstrap.New(a.logger)

	// 注册所有组件初始化器
	if err := a.registerComponents(); err != nil {
		return fmt.Errorf("failed to register components: %w", err)
	}

	a.logger.Infow("Components registered, ready to start")
	return nil
}

// Run 运行应用程序主逻辑
func (a *<Service>App) Run(ctx context.Context) error {
	a.logger.Infow("<Service> Service starting",
		"address", fmt.Sprintf("%s:%d", a.opts.Server.Host, a.opts.Server.Port),
	)

	// 使用 Bootstrap 运行所有组件
	// Bootstrap 会：
	// 1. 按依赖顺序初始化所有组件
	// 2. 等待上下文取消信号
	// 3. 优雅关闭所有组件
	return a.bootstrap.Run(ctx, nil)
}

// Shutdown 优雅关闭应用程序
func (a *<Service>App) Shutdown(ctx context.Context) error {
	a.logger.Infow("Shutting down <Service> Service")

	// Bootstrap 会自动调用所有初始化器的 Cleanup 方法
	return a.bootstrap.Shutdown(ctx)
}

// registerComponents 注册所有组件初始化器
func (a *<Service>App) registerComponents() error {
	// 1. 创建数据库初始化器
	a.dbInit = initializers.NewDatabaseInitializer(a.opts.Database, a.logger)

	// 2. 创建 Redis 初始化器
	a.redisInit = initializers.NewRedisInitializer(a.opts.Redis, a.logger)

	// 3. 创建业务服务初始化器（依赖 db 和 redis）
	a.serviceInit = initializers.NewServiceInitializer(
		a.config,
		a.logger,
		a.dbInit,
		a.redisInit,
	)

	// 4. 创建 HTTP 服务器初始化器（依赖业务服务）
	a.httpInit = initializers.NewHTTPServerInitializer(
		a.opts.Server,
		a.logger,
		a.serviceInit,
	)

	// 5. 创建健康检查初始化器（依赖 HTTP 服务器）
	a.healthInit = pkginitializers.NewHealthCheckInitializer(
		a.opts.Health,
		a.logger,
		a.httpInit,
	)

	// 按依赖顺序注册组件
	// Bootstrap 会根据 Dependencies() 方法自动排序
	a.bootstrap.Register(a.dbInit)
	a.bootstrap.Register(a.redisInit)
	a.bootstrap.Register(a.serviceInit)
	a.bootstrap.Register(a.httpInit)
	a.bootstrap.Register(a.healthInit)

	return nil
}

// initLogger 初始化日志系统
func initLogger(opts commonapp.Options) (core.Logger, error) {
	serverOpts := opts.(*options.ServerOptions)
	return serverOpts.InitLogger()
}
```

**要点**:
- ✅ 实现 `Application` 接口（Initialize, Run, Shutdown）
- ✅ 使用 `Bootstrap` 框架管理组件生命周期
- ✅ 清晰的依赖关系（通过初始化器构造函数传递）
- ✅ 按依赖顺序注册组件
- ✅ 完整的错误处理

---

### 4. 业务配置 (`internal/<service>/config.go`)

**定义业务配置结构和服务器创建方法**

```go
package <service>

import (
	"context"
	"fmt"
	"os"

	commonoptions "github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/logger/core"
)

var (
	// Name is the name of the compiled software.
	Name = "aetherius-<service>"

	// ID is the hostname of the machine running the service.
	ID, _ = os.Hostname()
)

// Config contains application-related configurations for the <service> service.
// 这是业务层配置，由启动层 Options 转换而来
type Config struct {
	Server   *commonoptions.ServerOptions
	Database *commonoptions.DatabaseOptions
	Redis    *commonoptions.RedisOptions
	Logging  *commonoptions.LoggingOptions

	// 根据服务需要添加其他配置
	// GRPC    *commonoptions.GRPCOptions
	// NATS    *commonoptions.NATSOptions
	// JWT     *commonoptions.JWTOptions
}

// Server represents the <service> service server.
// 这个结构可选，如果业务逻辑简单可以省略
type Server struct {
	cfg    *Config
	logger core.Logger

	// 业务组件（如果需要）
	// db      *gorm.DB
	// redis   *redis.Client
	// service *service.Service
}

// NewServer initializes and returns a new Server instance.
// 这个方法可选，也可以直接在 Application 中初始化组件
func (c *Config) NewServer(ctx context.Context, logger core.Logger) (*Server, error) {
	s := &Server{
		cfg:    c,
		logger: logger,
	}

	logger.Infow("<Service> server initialized successfully",
		"id", ID,
		"name", Name,
	)

	return s, nil
}
```

**要点**:
- ✅ 定义业务配置结构（从 Options 转换而来）
- ✅ 定义服务器结构（可选）
- ✅ 提供 NewServer 方法（可选）
- ❌ 不直接初始化组件（由 initializers 负责）

---

### 5. 组件初始化器 (`internal/<service>/initializers/database.go`)

**实现 bootstrap.Initializer 接口**

```go
package initializers

import (
	"fmt"
	"time"

	commondb "github.com/kart-io/k8s-agent/common/db"
	commonoptions "github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	"github.com/kart-io/logger/core"
	"gorm.io/gorm"
)

// DatabaseInitializer 数据库初始化器
type DatabaseInitializer struct {
	opts   *commonoptions.MySQLOptions
	logger core.Logger
	db     *gorm.DB
}

// 确保实现了 Initializer 接口
var _ bootstrap.Initializer = (*DatabaseInitializer)(nil)

// NewDatabaseInitializer creates a new DatabaseInitializer
func NewDatabaseInitializer(opts *commonoptions.MySQLOptions, logger core.Logger) *DatabaseInitializer {
	return &DatabaseInitializer{
		opts:   opts,
		logger: logger,
	}
}

// Name returns the name of the initializer
func (i *DatabaseInitializer) Name() string {
	return "database"
}

// Initialize initializes the database connection
func (i *DatabaseInitializer) Initialize() error {
	i.logger.Infow("Initializing database connection",
		"host", i.opts.Host,
		"port", i.opts.Port,
		"database", i.opts.Database,
	)

	// 使用 common/db 包初始化数据库
	db, err := commondb.NewMySQL(i.opts)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxIdleConns(i.opts.MaxIdleConnections)
	sqlDB.SetMaxOpenConns(i.opts.MaxOpenConnections)
	sqlDB.SetConnMaxLifetime(time.Duration(i.opts.MaxConnectionLifeTime) * time.Second)

	i.db = db

	i.logger.Infow("Database connection initialized successfully")
	return nil
}

// Cleanup cleans up the database connection
func (i *DatabaseInitializer) Cleanup() error {
	if i.db == nil {
		return nil
	}

	i.logger.Infow("Closing database connection")

	sqlDB, err := i.db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	i.logger.Infow("Database connection closed successfully")
	return nil
}

// Dependencies returns the list of dependencies
func (i *DatabaseInitializer) Dependencies() []string {
	return []string{} // 数据库通常没有依赖
}

// DB returns the database instance
func (i *DatabaseInitializer) DB() *gorm.DB {
	return i.db
}
```

**要点**:
- ✅ 实现 `bootstrap.Initializer` 接口（Name, Initialize, Cleanup, Dependencies）
- ✅ 提供 Getter 方法（如 DB()）供其他组件使用
- ✅ 完整的错误处理和日志记录
- ✅ 明确声明依赖关系

---

### 6. HTTP 服务器初始化器 (`internal/<service>/initializers/servers.go`)

**初始化 HTTP 服务器，依赖业务服务**

```go
package initializers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	commonoptions "github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/k8s-agent/internal/<service>/api"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	"github.com/kart-io/logger/core"
)

// HTTPServerInitializer HTTP 服务器初始化器
type HTTPServerInitializer struct {
	opts        *commonoptions.ServerOptions
	logger      core.Logger
	serviceInit *ServiceInitializer // 依赖业务服务

	server *http.Server
	router *gin.Engine
}

var _ bootstrap.Initializer = (*HTTPServerInitializer)(nil)

func NewHTTPServerInitializer(
	opts *commonoptions.ServerOptions,
	logger core.Logger,
	serviceInit *ServiceInitializer,
) *HTTPServerInitializer {
	return &HTTPServerInitializer{
		opts:        opts,
		logger:      logger,
		serviceInit: serviceInit,
	}
}

func (i *HTTPServerInitializer) Name() string {
	return "http-server"
}

func (i *HTTPServerInitializer) Initialize() error {
	i.logger.Infow("Initializing HTTP server",
		"host", i.opts.Host,
		"port", i.opts.Port,
	)

	// 创建 Gin 路由
	if i.opts.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()

	// 注册中间件
	router.Use(gin.Recovery())
	// 根据需要添加其他中间件

	// 注册路由
	api.RegisterRoutes(router, i.serviceInit.Service(), i.logger)

	// 创建 HTTP 服务器
	i.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", i.opts.Host, i.opts.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(i.opts.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(i.opts.WriteTimeout) * time.Second,
	}

	i.router = router

	// 在后台启动服务器
	go func() {
		i.logger.Infow("HTTP server started",
			"address", i.server.Addr,
		)
		if err := i.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			i.logger.Errorw("HTTP server error", "error", err)
		}
	}()

	return nil
}

func (i *HTTPServerInitializer) Cleanup() error {
	if i.server == nil {
		return nil
	}

	i.logger.Infow("Shutting down HTTP server")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := i.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	i.logger.Infow("HTTP server stopped successfully")
	return nil
}

func (i *HTTPServerInitializer) Dependencies() []string {
	return []string{"service"} // 依赖业务服务
}

func (i *HTTPServerInitializer) Router() *gin.Engine {
	return i.router
}

func (i *HTTPServerInitializer) Server() *http.Server {
	return i.server
}
```

**要点**:
- ✅ 声明依赖关系（通过构造函数和 Dependencies()）
- ✅ 在后台启动服务器（Initialize 不阻塞）
- ✅ 优雅关闭（Cleanup 中处理）
- ✅ 提供 Getter 方法供其他组件使用

---

## 🔍 关键差异对比

### agent-manager vs auth vs 新标准

| 维度 | agent-manager | auth | 新标准（综合） |
|------|---------------|------|----------------|
| **Options 位置** | `internal/<service>/config/` | `cmd/<service>/app/options/` | ✅ `cmd/<service>/app/options/` |
| **Config 结构** | ❌ 无独立 Config | ✅ `internal/<service>/config.go` | ✅ `internal/<service>/config.go` |
| **Config() 方法** | ❌ 无 | ✅ 有 | ✅ 有 |
| **InitLogger() 方法** | ❌ 在 app.go 中 | ✅ 在 options.go 中 | ✅ 在 options.go 中 |
| **GetHealthPort() 方法** | ❌ 无 | ✅ 有 | ✅ 有 |
| **Application 接口** | ✅ 完整实现 | ⚠️ 使用 run 函数 | ✅ 完整实现 |
| **Bootstrap 框架** | ✅ 使用 | ❌ 不使用 | ✅ 使用 |
| **Initializer 接口** | ✅ 实现 | ⚠️ 部分实现 | ✅ 实现 |

**结论**: 新标准综合了两者的优点，提供了最清晰的分层和最完整的功能。

---

## ✅ 实施检查清单

每个服务实现完成后，使用此清单验证是否符合标准：

### 目录结构
- [ ] `cmd/<service>/main.go` 存在且极简
- [ ] `cmd/<service>/app/options/options.go` 存在
- [ ] `cmd/<service>/app/app.go` 存在
- [ ] `internal/<service>/config.go` 存在
- [ ] `internal/<service>/api/` 目录存在
- [ ] `internal/<service>/initializers/` 目录存在

### Options 实现
- [ ] 实现 `commonapp.Options` 接口（AddFlags, Complete, Validate）
- [ ] 实现 `Config()` 方法
- [ ] 实现 `InitLogger()` 方法
- [ ] 实现 `GetHealthPort()` 方法
- [ ] 使用 `common/options` 标准配置
- [ ] 设置服务名称到日志字段

### Application 实现
- [ ] 实现 `commonapp.Application` 接口（Initialize, Run, Shutdown）
- [ ] 使用 `bootstrap.Bootstrap` 管理组件
- [ ] 在 `Initialize` 中注册所有组件
- [ ] 在 `Run` 中调用 `bootstrap.Run()`
- [ ] 在 `Shutdown` 中调用 `bootstrap.Shutdown()`

### 初始化器实现
- [ ] 所有组件实现 `bootstrap.Initializer` 接口
- [ ] 明确声明依赖关系（Dependencies()）
- [ ] 有完整的错误处理和日志
- [ ] 有清理逻辑（Cleanup()）
- [ ] 提供 Getter 方法供其他组件使用

### 功能完整性
- [ ] 支持配置文件加载
- [ ] 支持环境变量
- [ ] 支持命令行参数
- [ ] 有健康检查端点
- [ ] 有优雅关闭
- [ ] 有版本信息
- [ ] 有结构化日志

---

## 📊 迁移优先级

### 高优先级：Orchestrator
- **当前状态**: ❌ 使用原始 flag 包，无框架支持
- **工作量**: 2-3 天
- **收益**: 最大（问题最多）

### 中优先级：Agent-Manager
- **当前状态**: ⚠️ 使用旧 Options 位置，缺少 Config 转换
- **工作量**: 1 天
- **收益**: 中等（主要是结构调整）

### 低优先级：Auth, Reasoning
- **当前状态**: ⚠️ 使用 RunWithOptions，缺少 Application 接口
- **工作量**: 每个 0.5-1 天
- **收益**: 小（主要是补充 Application 接口）

---

## 🚀 快速开始

### 1. 创建新服务骨架

```bash
# 使用脚本生成服务骨架（待实现）
./scripts/generate-service.sh <service-name>

# 或手动创建目录结构
mkdir -p cmd/<service>/app/options
mkdir -p internal/<service>/{api,initializers,service,storage}
```

### 2. 复制模板文件

```bash
# 复制 Options 模板
cp templates/service/options.go.tmpl cmd/<service>/app/options/options.go

# 复制 Application 模板
cp templates/service/app.go.tmpl cmd/<service>/app/app.go

# 复制 Config 模板
cp templates/service/config.go.tmpl internal/<service>/config.go
```

### 3. 替换占位符

```bash
# 替换所有 <service> 为实际服务名
sed -i 's/<service>/myservice/g' cmd/myservice/**/*.go
sed -i 's/<Service>/MyService/g' cmd/myservice/**/*.go
sed -i 's/<SERVICE>/MYSERVICE/g' cmd/myservice/**/*.go
```

### 4. 实现业务逻辑

1. 在 `internal/<service>/api/` 中实现 HTTP 处理器
2. 在 `internal/<service>/service/` 中实现业务逻辑
3. 在 `internal/<service>/storage/` 中实现数据访问
4. 在 `internal/<service>/initializers/` 中实现组件初始化器

---

## 📚 参考资源

### 示例实现
- ✅ **最佳实践**: `cmd/auth/app/` + `internal/auth/`
- ⚠️ **部分参考**: `cmd/agent-manager/app/` (Application 接口)
- ❌ **反面教材**: `cmd/orchestrator/app/` (需要重构)

### 文档
- [pkg/app/README.md](../../pkg/app/README.md) - App 框架文档
- [pkg/bootstrap/](../../pkg/bootstrap/) - Bootstrap 框架
- [common/options/README.md](../../common/options/README.md) - 标准配置选项
- [SERVICE_UNIFICATION_PLAN.md](./SERVICE_UNIFICATION_PLAN.md) - 统一化计划

---

## ❓ 常见问题

### Q1: 为什么要将 Options 放在 `cmd/<service>/app/options/` 而不是 `internal/<service>/config/`？

**A**: 职责分离：
- `cmd/<service>/app/options/` - **启动层配置**，负责命令行参数、环境变量、配置文件加载
- `internal/<service>/config.go` - **业务层配置**，负责业务逻辑所需的配置
- 这种分离使得启动层和业务层解耦，更易于测试和维护

### Q2: Config() 方法的作用是什么？

**A**: 将启动层配置转换为业务层配置：
- 启动层关心的是：配置来源（文件、环境变量、命令行）
- 业务层关心的是：配置内容（数据库连接、Redis 配置等）
- Config() 是两者之间的桥梁

### Q3: 为什么要使用 Bootstrap 框架而不是直接初始化？

**A**: Bootstrap 提供：
- 自动依赖管理（按 Dependencies() 排序）
- 统一的错误处理和清理机制
- 更好的可测试性（可以 mock 初始化器）
- 更清晰的代码结构

### Q4: 服务特定配置（如 NATS、JWT）应该放在哪里？

**A**:
- 如果是 `common/options` 中已有的，直接使用
- 如果是服务特定的，在 `cmd/<service>/app/options/options.go` 中定义新的配置结构
- 如果多个服务都需要，考虑添加到 `common/options`

### Q5: 是否必须实现 Server 结构？

**A**: 不是必须的。Server 结构是可选的：
- 简单服务：可以直接在 Application 中管理组件
- 复杂服务：可以使用 Server 结构封装业务逻辑
- 看团队习惯和项目需求

---

**文档版本**: v1.0
**创建日期**: 2025-01-25
**最后更新**: 2025-01-25
**维护者**: Architecture Team
