# 综合最佳实践总结

> agent-manager + auth 综合方案

---

## 🎯 核心设计

### 三层架构

```
┌──────────────────────────────────────────────────────────┐
│  Layer 1: 启动层 (cmd/<service>/app/)                   │
│  ├─ options/options.go  ← 命令行配置（commonapp.Options）│
│  ├─ app.go              ← Application 生命周期管理        │
│  └─ server.go           ← 辅助函数（可选）                │
└──────────────────────────────────────────────────────────┘
                    ↓ Config() 转换
┌──────────────────────────────────────────────────────────┐
│  Layer 2: 业务层 (internal/<service>/)                  │
│  ├─ config.go           ← 业务配置结构                    │
│  └─ server.go           ← 业务服务器（可选）              │
└──────────────────────────────────────────────────────────┘
                    ↓ 注册组件
┌──────────────────────────────────────────────────────────┐
│  Layer 3: 组件层 (internal/<service>/initializers/)     │
│  ├─ database.go         ← 数据库初始化器                  │
│  ├─ redis.go            ← Redis 初始化器                  │
│  └─ servers.go          ← HTTP/gRPC 服务器初始化器        │
└──────────────────────────────────────────────────────────┘
```

---

## ✅ 继承的优点

### 从 agent-manager 继承

✅ **Application 接口** - 完整的生命周期管理
```go
type Application interface {
    Initialize(ctx context.Context, opts Options) error
    Run(ctx context.Context) error
    Shutdown(ctx context.Context) error
}
```

✅ **Bootstrap 框架** - 自动依赖管理
```go
a.bootstrap.Register(a.dbInit)      // 无依赖
a.bootstrap.Register(a.redisInit)   // 无依赖
a.bootstrap.Register(a.serviceInit) // 依赖 db + redis
a.bootstrap.Register(a.httpInit)    // 依赖 service
```

### 从 auth 继承

✅ **Options 目录结构** - 清晰的分层
```
cmd/auth/app/options/options.go     ← 启动层配置
internal/auth/config.go              ← 业务层配置
```

✅ **Config() 转换方法** - 分离关注点
```go
func (o *ServerOptions) Config() (*auth.Config, error) {
    return &auth.Config{
        Server:   o.Server,
        Database: o.Database,
        // ...
    }, nil
}
```

✅ **辅助方法** - 简化使用
```go
func (o *ServerOptions) InitLogger() (core.Logger, error)
func (o *ServerOptions) GetHealthPort() int
```

---

## 📦 完整示例（关键代码）

### 1. Options (cmd/<service>/app/options/options.go)

```go
type ServerOptions struct {
    Server   *commonoptions.ServerOptions   `json:"server"`
    Database *commonoptions.DatabaseOptions `json:"database"`
    Redis    *commonoptions.RedisOptions    `json:"redis"`
    Logging  *commonoptions.LoggingOptions  `json:"logging"`
    Health   *commonoptions.HealthOptions   `json:"health"`
}

func (o *ServerOptions) Config() (*service.Config, error) {
    return &service.Config{
        Server:   o.Server,
        Database: o.Database,
        // ...
    }, nil
}

func (o *ServerOptions) InitLogger() (core.Logger, error) {
    return commonlogger.InitFromOptions(o.Logging)
}

func (o *ServerOptions) GetHealthPort() int {
    return o.Health.Port
}
```

### 2. Application (cmd/<service>/app/app.go)

```go
func Execute() {
    opts := options.NewServerOptions()

    commonapp.RunWithRunner(
        opts,
        &ServiceApp{},
        initLogger,
        commonapp.CommandConfig{
            Use:       "service",
            EnvPrefix: "SERVICE",
        },
    )
}

type ServiceApp struct {
    bootstrap *bootstrap.Bootstrap
    opts      *options.ServerOptions
    config    *service.Config
    logger    core.Logger

    // 初始化器
    dbInit    *initializers.DatabaseInitializer
    httpInit  *initializers.HTTPServerInitializer
}

func (a *ServiceApp) Initialize(ctx context.Context, opts commonapp.Options) error {
    a.opts = opts.(*options.ServerOptions)

    // 初始化日志
    logger, err := initLogger(opts)
    a.logger = logger

    // 转换配置
    config, err := a.opts.Config()
    a.config = config

    // 创建 Bootstrap
    a.bootstrap = bootstrap.New(a.logger)

    // 注册组件
    a.registerComponents()
    return nil
}

func (a *ServiceApp) Run(ctx context.Context) error {
    return a.bootstrap.Run(ctx, nil)
}

func (a *ServiceApp) Shutdown(ctx context.Context) error {
    return a.bootstrap.Shutdown(ctx)
}
```

### 3. Config (internal/<service>/config.go)

```go
type Config struct {
    Server   *commonoptions.ServerOptions
    Database *commonoptions.DatabaseOptions
    Redis    *commonoptions.RedisOptions
}

func (c *Config) NewServer(ctx context.Context, logger core.Logger) (*Server, error) {
    return &Server{
        cfg:    c,
        logger: logger,
    }, nil
}
```

### 4. Initializer (internal/<service>/initializers/database.go)

```go
type DatabaseInitializer struct {
    opts   *commonoptions.DatabaseOptions
    logger core.Logger
    db     *gorm.DB
}

func (i *DatabaseInitializer) Name() string { return "database" }

func (i *DatabaseInitializer) Initialize() error {
    db, err := commondb.NewMySQL(i.opts)
    i.db = db
    return err
}

func (i *DatabaseInitializer) Cleanup() error {
    sqlDB, _ := i.db.DB()
    return sqlDB.Close()
}

func (i *DatabaseInitializer) Dependencies() []string {
    return []string{} // 无依赖
}

func (i *DatabaseInitializer) DB() *gorm.DB {
    return i.db
}
```

---

## 🔄 数据流转

```
配置文件/环境变量/命令行参数
    ↓
ServerOptions (启动层)
    ├─ AddFlags()    - 注册命令行参数
    ├─ Complete()    - 设置默认值
    ├─ Validate()    - 验证配置
    ├─ Config()      - 转换为业务配置
    ├─ InitLogger()  - 初始化日志
    └─ GetHealthPort() - 获取健康检查端口
    ↓
Config (业务层)
    └─ NewServer()   - 创建服务器（可选）
    ↓
Initializers (组件层)
    ├─ DatabaseInitializer
    ├─ RedisInitializer
    └─ HTTPServerInitializer
    ↓
Bootstrap.Run()
    ├─ 按依赖顺序初始化组件
    ├─ 等待上下文取消
    └─ 优雅关闭所有组件
```

---

## 🆚 对比表

| 特性 | agent-manager | auth | 综合方案 ✅ |
|------|---------------|------|------------|
| Options 位置 | `internal/config/` | `cmd/app/options/` | `cmd/app/options/` |
| Config 分离 | ❌ | ✅ | ✅ |
| Config() 方法 | ❌ | ✅ | ✅ |
| InitLogger() | 在 app.go | 在 options.go | 在 options.go |
| GetHealthPort() | ❌ | ✅ | ✅ |
| Application 接口 | ✅ | ⚠️ (run函数) | ✅ |
| Bootstrap 框架 | ✅ | ❌ | ✅ |
| Initializer 接口 | ✅ | ⚠️ (部分) | ✅ |
| **总评** | **70分** | **75分** | **100分** |

---

## 📋 快速检查清单

使用此清单验证服务是否符合标准：

### 目录结构 ✅
- [ ] `cmd/<service>/main.go` - 极简入口
- [ ] `cmd/<service>/app/options/options.go` - 启动层配置
- [ ] `cmd/<service>/app/app.go` - Application 实现
- [ ] `internal/<service>/config.go` - 业务层配置
- [ ] `internal/<service>/initializers/` - 组件初始化器

### Options 实现 ✅
- [ ] 实现 `commonapp.Options` 接口
- [ ] 实现 `Config()` 转换方法
- [ ] 实现 `InitLogger()` 辅助方法
- [ ] 实现 `GetHealthPort()` 辅助方法

### Application 实现 ✅
- [ ] 实现 `commonapp.Application` 接口
- [ ] 使用 `bootstrap.Bootstrap` 管理组件
- [ ] 在 `Initialize` 中注册组件
- [ ] 在 `Run` 中调用 `bootstrap.Run()`

### Initializer 实现 ✅
- [ ] 实现 `bootstrap.Initializer` 接口
- [ ] 声明依赖关系 (`Dependencies()`)
- [ ] 提供 Getter 方法
- [ ] 实现完整的 `Cleanup()`

---

## 🚀 迁移步骤

### 对于 orchestrator（从零开始）

1. 创建目录结构
2. 创建 Options (参考 auth)
3. 创建 Application (参考 agent-manager)
4. 创建 Initializers (参考 agent-manager)
5. 创建 Config (参考 auth)
6. 测试验证

### 对于 agent-manager（调整结构）

1. 移动 `internal/config/options.go` → `cmd/app/options/options.go`
2. 创建 `internal/config.go` (业务配置)
3. 添加 `Config()` 转换方法
4. 添加 `InitLogger()` 辅助方法
5. 添加 `GetHealthPort()` 辅助方法
6. 测试验证

### 对于 auth（补充 Application）

1. 修改 app.go 使用 `RunWithRunner`
2. 创建 `ServiceApp` 结构（实现 Application 接口）
3. 将 run 函数逻辑迁移到 `Initialize/Run/Shutdown`
4. 使用 `bootstrap.Bootstrap` 管理组件
5. 测试验证

---

## 💡 关键理解

### 为什么三层分离？

1. **启动层** (`cmd/app/options/`) - 关心配置来源
   - 配置文件在哪？
   - 环境变量怎么映射？
   - 命令行参数是什么？

2. **业务层** (`internal/config.go`) - 关心配置内容
   - 数据库连接信息是什么？
   - Redis 地址是什么？
   - 业务逻辑需要什么配置？

3. **组件层** (`internal/initializers/`) - 关心初始化顺序
   - 数据库要先初始化
   - Redis 可以并行初始化
   - HTTP 服务器依赖业务服务

### Config() 的价值

```go
// 启动层只关心配置加载
opts := options.NewServerOptions()
commonapp.RunWithRunner(opts, ...)

// 业务层只关心配置使用
config, _ := opts.Config()
db := config.Database  // 直接使用
```

分离后：
- 启动层可以独立测试（配置加载逻辑）
- 业务层可以独立测试（不关心配置来源）
- 两层解耦，易于维护

---

## 📖 文档索引

- **[SERVICE_STANDARD_PATTERN.md](./SERVICE_STANDARD_PATTERN.md)** - 完整的实现指南（含代码模板）
- **[SERVICE_UNIFICATION_PLAN.md](./SERVICE_UNIFICATION_PLAN.md)** - 统一化重构计划
- **[templates/service/](../templates/service/)** - 代码模板和生成脚本
- **[pkg/app/README.md](../pkg/app/README.md)** - App 框架文档

---

**总结**: 综合方案 = agent-manager (生命周期管理) + auth (清晰分层) = 最佳实践 ✅

**下一步**: 开始重构 orchestrator 服务，验证综合方案的可行性。
