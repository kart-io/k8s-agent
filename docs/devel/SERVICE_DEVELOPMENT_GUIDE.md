# 服务开发指南

## 目录

- [服务开发指南](#服务开发指南)
  - [目录](#目录)
  - [1. 概述](#1-概述)
  - [2. 架构说明](#2-架构说明)
    - [2.1 Bootstrap 框架](#21-bootstrap-框架)
    - [2.2 组件初始化器](#22-组件初始化器)
    - [2.3 服务类型](#23-服务类型)
  - [3. 标准项目结构](#3-标准项目结构)
  - [4. 开发新服务](#4-开发新服务)
    - [4.1 创建 ServerOptions](#41-创建-serveroptions)
    - [4.2 实现 Application 接口](#42-实现-application-接口)
    - [4.3 创建 Initializers](#43-创建-initializers)
  - [5. Options 结构规范](#5-options-结构规范)
    - [5.1 必须包含的字段](#51-必须包含的字段)
    - [5.2 可选字段列表](#52-可选字段列表)
    - [5.3 方法实现标准](#53-方法实现标准)
  - [6. Initializers 开发规范](#6-initializers-开发规范)
    - [6.1 接口定义](#61-接口定义)
    - [6.2 优先级设置指南](#62-优先级设置指南)
    - [6.3 依赖管理](#63-依赖管理)
  - [7. 服务类型示例](#7-服务类型示例)
    - [7.1 标准 HTTP 服务 (auth, cluster)](#71-标准-http-服务-auth-cluster)
    - [7.2 后台任务服务 (collect-agent)](#72-后台任务服务-collect-agent)
    - [7.3 gRPC + HTTP 统一服务 (reasoning)](#73-grpc--http-统一服务-reasoning)
    - [7.4 反向代理服务 (gateway)](#74-反向代理服务-gateway)
  - [8. 最佳实践](#8-最佳实践)
  - [9. 常见问题](#9-常见问题)
  - [10. 迁移指南](#10-迁移指南)

---

## 1. 概述

本指南介绍如何在 k8s-agent 项目中开发新服务或维护现有服务。所有服务统一使用 **Bootstrap 框架**，确保一致性和可维护性。

### 核心原则

1. **统一框架**：所有服务使用 Bootstrap 框架
2. **标准化配置**：Options 结构遵循统一规范
3. **组件化设计**：使用 Initializers 管理组件生命周期
4. **依赖管理**：通过优先级控制组件初始化顺序

---

## 2. 架构说明

### 2.1 Bootstrap 框架

Bootstrap 框架负责：

- **组件注册**：注册所有 Initializer
- **依赖解析**：按优先级初始化组件
- **生命周期管理**：Initialize -> Run -> Close
- **优雅关停**：处理信号并协调组件关闭

### 2.2 组件初始化器

每个 Initializer 负责一个组件的生命周期：

```go
type Initializer interface {
    Name() string                      // 组件名称
    Priority() int                     // 优先级（低数字先执行）
    Initialize(ctx context.Context) error
    Run(ctx context.Context) error
    Close(ctx context.Context) error
}
```

### 2.3 服务类型

- **HTTP REST API 服务**：auth, cluster, monitor, agent-manager
- **gRPC + HTTP 统一服务**：reasoning, orchestrator
- **API 网关服务**：gateway（反向代理）
- **后台任务服务**：collect-agent

---

## 3. 标准项目结构

```
cmd/service-name/
  ├─ main.go                  # 入口文件
  └─ app/
      ├─ app.go               # Application 实现
      └─ options/
          └─ options.go       # ServerOptions 定义

internal/service-name/
  ├─ initializers/            # Bootstrap 初始化器
  │   ├─ database.go          # 数据库初始化器
  │   ├─ redis.go             # Redis 初始化器
  │   ├─ http_server.go       # HTTP 服务器初始化器
  │   └─ ...
  ├─ handler/                 # HTTP handlers
  ├─ service/                 # 业务逻辑
  └─ storage/                 # 数据访问层

configs/service-name/
  └─ service-name.yaml        # 配置文件示例
```

---

## 4. 开发新服务

### 4.1 创建 ServerOptions

```go
// cmd/myservice/app/options/options.go
package options

import (
	"github.com/spf13/pflag"
	"github.com/kart-io/k8s-agent/common/loggerutil"
	commonoptions "github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/logger/core"
)

// ServerOptions 定义服务配置
type ServerOptions struct {
	// 核心配置（必须）
	Server  *commonoptions.ServerOptions  `json:"server" mapstructure:"server"`
	Logging *commonoptions.LoggingOptions `json:"logging" mapstructure:"logging"`
	Health  *commonoptions.HealthOptions  `json:"health" mapstructure:"health"`

	// 可选配置（按需）
	Database *commonoptions.DatabaseOptions `json:"database,omitempty" mapstructure:"database"`
	Redis    *commonoptions.RedisOptions    `json:"redis,omitempty" mapstructure:"redis"`

	// 服务特定配置
	MyConfig MyConfigOptions `json:"my_config" mapstructure:"my_config"`
}

// NewServerOptions 创建默认配置
func NewServerOptions() *ServerOptions {
	return &ServerOptions{
		Server:   commonoptions.NewServerOptions(),
		Logging:  commonoptions.NewLoggingOptions(),
		Health:   commonoptions.NewHealthOptions(),
		Database: commonoptions.NewDatabaseOptions(),
		Redis:    commonoptions.NewRedisOptions(),
		MyConfig: MyConfigOptions{
			// 默认值
		},
	}
}

// Validate 验证配置
func (o *ServerOptions) Validate() []error {
	// 使用通用工具函数验证基础配置
	errs := commonoptions.ValidateAll(o)

	// 添加服务特定验证
	// ...

	return errs
}

// Complete 完成配置
func (o *ServerOptions) Complete() error {
	// 使用通用工具函数完成基础配置
	if err := commonoptions.CompleteAll(o); err != nil {
		return err
	}

	// 添加服务特定默认值
	// ...

	return nil
}

// AddFlags 添加命令行参数
func (o *ServerOptions) AddFlags(fs *pflag.FlagSet) {
	// 使用通用工具函数添加基础参数
	commonoptions.AddFlagsAll(o, fs)

	// 添加服务特定参数
	// ...
}

// InitLogger 初始化日志
func (o *ServerOptions) InitLogger() (core.Logger, error) {
	return loggerutil.InitFromOptions(o.Logging)
}

// GetServiceName 返回服务名称
func (o *ServerOptions) GetServiceName() string {
	return "MyService"
}

// GetLogFields 返回日志字段
func (o *ServerOptions) GetLogFields() []interface{} {
	return []interface{}{
		"http_port", o.Server.Port,
		"health_port", o.Health.Port,
	}
}
```

### 4.2 实现 Application 接口

```go
// cmd/myservice/app/app.go
package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/cmd/myservice/app/options"
	commonoptions "github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/k8s-agent/internal/myservice/initializers"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// Execute 运行服务
func Execute() {
	opts := options.NewServerOptions()
	app := &MyServiceApp{}

	commonapp.RunWithBootstrap(
		app,
		opts,
		commonapp.Config{
			Use:       "myservice",
			Short:     "My Service",
			Long:      "My Service does something awesome",
			EnvPrefix: "MYSERVICE",
		},
		app.registerComponents,
	)
}

// MyServiceApp 实现 Application 接口
type MyServiceApp struct {
	opts   *options.ServerOptions
	logger core.Logger

	// Component initializers
	dbInit     *initializers.DatabaseInitializer
	redisInit  *initializers.RedisInitializer
	httpInit   *initializers.HTTPServerInitializer
	healthInit *pkginitializers.HealthCheckInitializer
}

// Name 返回服务名称
func (a *MyServiceApp) Name() string {
	return "My Service"
}

// Initialize 初始化应用
func (a *MyServiceApp) Initialize(ctx context.Context, opts commonapp.Options) error {
	a.opts = opts.(*options.ServerOptions)

	logger, err := a.opts.InitLogger()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	a.logger = logger

	a.logger.Info("Starting My Service...")
	return nil
}

// Run 运行应用
func (a *MyServiceApp) Run(ctx context.Context) error {
	// Bootstrap 框架会运行所有服务器
	<-ctx.Done()
	return nil
}

// Shutdown 关闭应用
func (a *MyServiceApp) Shutdown(ctx context.Context) error {
	// Bootstrap 框架会关闭所有组件
	return nil
}

// registerComponents 注册组件
func (a *MyServiceApp) registerComponents(bs *bootstrap.Bootstrap) error {
	// 1. Database (priority 300)
	a.dbInit = initializers.NewDatabaseInitializer(a.opts, a.logger)
	bs.Register(a.dbInit)

	// 2. Redis (priority 400)
	a.redisInit = initializers.NewRedisInitializer(a.opts, a.logger)
	bs.Register(a.redisInit)

	// 3. HTTP Server (priority 500)
	a.httpInit = initializers.NewHTTPServerInitializer(
		a.opts,
		a.logger,
		a.dbInit,
		a.redisInit,
	)
	bs.Register(a.httpInit)

	// 4. Health Check (priority 2000)
	a.healthInit = pkginitializers.NewHealthCheckInitializer(
		a.opts.Health,
		a.logger,
	)
	bs.Register(a.healthInit)

	return nil
}
```

### 4.3 创建 Initializers

#### HTTP 服务器初始化器

```go
// internal/myservice/initializers/http_server.go
package initializers

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/cmd/myservice/app/options"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// HTTPServerInitializer 封装 HTTP 服务器初始化
type HTTPServerInitializer struct {
	*pkginitializers.HTTPServerInitializer
	cfg       *options.ServerOptions
	logger    core.Logger
	dbInit    *DatabaseInitializer
	redisInit *RedisInitializer
}

// NewHTTPServerInitializer 创建 HTTP 服务器初始化器
func NewHTTPServerInitializer(
	cfg *options.ServerOptions,
	logger core.Logger,
	dbInit *DatabaseInitializer,
	redisInit *RedisInitializer,
) *HTTPServerInitializer {
	h := &HTTPServerInitializer{
		cfg:       cfg,
		logger:    logger,
		dbInit:    dbInit,
		redisInit: redisInit,
	}

	// 创建标准 HTTP 服务器配置
	serverConfig := &pkginitializers.HTTPServerConfig{
		Name:       "myservice-http-server",
		Priority:   bootstrap.PriorityHTTP,
		Config:     cfg.Server,
		RouteSetup: h.setupRoutes,
	}

	h.HTTPServerInitializer = pkginitializers.NewHTTPServerInitializer(serverConfig, logger)
	return h
}

// setupRoutes 配置路由
func (h *HTTPServerInitializer) setupRoutes(engine *gin.Engine) error {
	h.logger.Infow("Setting up myservice routes")

	// 健康检查
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API 路由
	api := engine.Group("/api/v1")
	{
		// 添加你的路由
	}

	h.logger.Infow("Routes registered")
	return nil
}
```

---

## 5. Options 结构规范

### 5.1 必须包含的字段

```go
type ServerOptions struct {
    Server  *commonoptions.ServerOptions  `json:"server" mapstructure:"server"`
    Logging *commonoptions.LoggingOptions `json:"logging" mapstructure:"logging"`
    Health  *commonoptions.HealthOptions  `json:"health" mapstructure:"health"`
    // ...
}
```

### 5.2 可选字段列表

根据服务需求选择：

- `Database *DatabaseOptions` - 需要数据库
- `Redis *RedisOptions` - 需要 Redis
- `GRPC *GRPCOptions` - 需要 gRPC 服务器
- `NATS *NATSOptions` - 需要消息队列
- `JWT *JWTOptions` - 需要 JWT 认证
- `Metrics *MetricsOptions` - 需要 Prometheus metrics

### 5.3 方法实现标准

**必须实现的方法：**

- `Complete() error` - 使用 `commonoptions.CompleteAll()`
- `Validate() []error` - 使用 `commonoptions.ValidateAll()`
- `AddFlags(fs *pflag.FlagSet)` - 使用 `commonoptions.AddFlagsAll()`
- `InitLogger() (core.Logger, error)` - 使用 `loggerutil.InitFromOptions()`
- `GetServiceName() string` - 返回服务名称
- `GetLogFields() []interface{}` - 返回日志字段

---

## 6. Initializers 开发规范

### 6.1 接口定义

所有 Initializer 必须实现：

```go
type Initializer interface {
    Name() string                      // 返回组件名称
    Priority() int                     // 返回优先级
    Initialize(ctx context.Context) error  // 初始化
    Run(ctx context.Context) error         // 运行（可阻塞）
    Close(ctx context.Context) error       // 关闭
}
```

### 6.2 优先级设置指南

```go
// 标准优先级（在 pkg/bootstrap/constants.go 中定义）
const (
    PriorityDatabase    = 300   // 数据库
    PriorityRedis       = 400   // Redis
    PriorityNATS        = 500   // 消息队列
    PriorityHTTP        = 1000  // HTTP 服务器
    PriorityGRPC        = 1100  // gRPC 服务器
    PriorityApplication = 1500  // 应用逻辑
    PriorityHealthCheck = 2000  // 健康检查
)
```

**原则：**
- 依赖先初始化（低优先级数字）
- 服务器后启动（高优先级数字）
- 健康检查最后启动

### 6.3 依赖管理

通过构造函数传递依赖：

```go
type HTTPServerInitializer struct {
    dbInit    *DatabaseInitializer
    redisInit *RedisInitializer
}

func NewHTTPServerInitializer(
    cfg *options.ServerOptions,
    logger core.Logger,
    dbInit *DatabaseInitializer,      // 依赖注入
    redisInit *RedisInitializer,      // 依赖注入
) *HTTPServerInitializer {
    // ...
}

func (h *HTTPServerInitializer) Initialize(ctx context.Context) error {
    // 使用依赖
    db := h.dbInit.DB()
    redis := h.redisInit.Client()
    // ...
}
```

---

## 7. 服务类型示例

### 7.1 标准 HTTP 服务 (auth, cluster)

**特点：**
- REST API
- 数据库 + Redis
- JWT 认证

**参考：** `cmd/auth/app/app.go`

### 7.2 后台任务服务 (collect-agent)

**特点：**
- 无 HTTP 服务器（仅健康检查）
- 长期运行的后台任务
- 连接外部系统

**参考：** `cmd/collect-agent/app/app.go`

**关键点：**
- Agent 在 `Run()` 中阻塞运行
- 通过 context 取消来停止

### 7.3 gRPC + HTTP 统一服务 (reasoning)

**特点：**
- gRPC 服务器
- HTTP 网关（grpc-gateway）
- 共享 handler

**参考：** `cmd/reasoning/app/app.go`

### 7.4 反向代理服务 (gateway)

**特点：**
- 路由到多个后端服务
- CORS、限流、认证中间件
- 动态路由配置

**参考：** `cmd/gateway/app/app.go`

---

## 8. 最佳实践

1. **日志使用**：所有日志使用结构化字段（`logger.Infow("msg", "key", value)`）
2. **错误处理**：返回带上下文的错误（`fmt.Errorf("context: %w", err)`）
3. **配置默认值**：在 `NewServerOptions()` 中设置合理默认值
4. **优雅关停**：依赖 Bootstrap 框架，不要自己处理信号
5. **健康检查**：所有服务必须包含 Health Initializer
6. **依赖注入**：通过构造函数传递，不要使用全局变量

---

## 9. 常见问题

**Q: 为什么所有服务都用 Bootstrap？**
A: 统一框架提高可维护性、减少重复代码、易于添加全局功能（如 tracing）

**Q: 简单服务不会太重吗？**
A: Bootstrap 开销很小，且提供了统一的生命周期管理，值得

**Q: 如何添加新的 Initializer？**
A: 实现 `Initializer` 接口，在 `registerComponents()` 中注册

**Q: 优先级如何选择？**
A: 参考优先级指南，依赖先初始化，服务器后启动

**Q: 可以不使用 Health Initializer 吗？**
A: 不可以，这是标准要求

---

## 10. 迁移指南

**从旧模式迁移到 Bootstrap：**

1. **创建 initializers 目录**
   ```
   internal/service-name/initializers/
   ```

2. **将 server.go 逻辑拆分到 Initializers**
   - 数据库连接 → `database.go`
   - Redis 连接 → `redis.go`
   - 路由设置 → `http_server.go`

3. **修改 app.go**
   - 使用 `RunWithBootstrap()` 而不是 `Run()`
   - 实现 `registerComponents()`
   - 移除手动服务器管理代码

4. **删除旧的 server.go**

5. **测试**
   - 确保服务可以启动
   - 确保健康检查可用
   - 确保优雅关停正常

---

**完成开发后，请参考模板目录中的示例进行验证。**

