# 服务入口标准化 - 快速参考指南

## 🎯 快速决策

### 我应该使用哪种模式？

```
┌─────────────────────────────────────┐
│ 你的服务有多个外部依赖吗？           │
│ (数据库、Redis、NATS等)              │
└─────────┬───────────────────────────┘
          │
    ┌─────┴──────┐
    YES          NO
    │            │
    │            ▼
    │     ┌─────────────────────────┐
    │     │ 使用 Simple 模式         │
    │     │ - collect-agent         │
    │     │ - gateway               │
    │     │ - monitor               │
    │     └─────────────────────────┘
    │
    ▼
┌──────────────────────────────────┐
│ 使用 Bootstrap 模式              │
│ - agent-manager                  │
│ - orchestrator                   │
│ - cluster                        │
│ - reasoning                      │
└──────────────────────────────────┘
```

---

## 📋 Bootstrap 模式清单

### 1. 创建目录结构
```bash
cmd/{service}/
├── main.go
└── app/
    ├── app.go
    └── options/
        └── options.go

internal/{service}/
└── initializers/
    ├── database.go
    ├── redis.go
    └── http_server.go
```

### 2. main.go 模板
```go
package main

import (
    _ "go.uber.org/automaxprocs/maxprocs"
    "github.com/kart-io/k8s-agent/cmd/{service}/app"
)

func main() {
    app.Execute()
}
```

### 3. options.go 必需方法
```go
type ServerOptions struct {
    // ...
}

func NewServerOptions() *ServerOptions {}
func (o *ServerOptions) Validate() error {}
func (o *ServerOptions) InitLogger() (core.Logger, error) {}
func (o *ServerOptions) GetHealthPort() int {}
func (o *ServerOptions) Config() (*{service}.Config, error) {}
```

### 4. app.go 核心结构
```go
type {Service}App struct {
    bootstrap *bootstrap.Bootstrap
    opts      *options.ServerOptions
    logger    core.Logger

    // 初始化器
    dbInit     *initializers.DatabaseInitializer
    healthInit *pkginitializers.HealthCheckInitializer
}

func Execute() {
    opts := options.NewServerOptions()
    commonapp.RunWithRunner(opts, &{Service}App{}, initLogger, config)
}

func (a *{Service}App) Initialize(ctx, opts) error {}
func (a *{Service}App) Run(ctx) error {}
func (a *{Service}App) Shutdown(ctx) error {}
func (a *{Service}App) registerComponents() {}
```

### 5. 初始化器模板
```go
type DatabaseInitializer struct {
    opts   *options.ServerOptions
    logger core.Logger
    db     *gorm.DB
}

func (i *DatabaseInitializer) Name() string { return "database" }
func (i *DatabaseInitializer) Priority() int { return 300 }
func (i *DatabaseInitializer) Initialize(ctx context.Context) error {}
func (i *DatabaseInitializer) Shutdown(ctx context.Context) error {}
```

### 6. 注册组件（按优先级）
```go
func (a *{Service}App) registerComponents() {
    // 300: Database
    a.dbInit = initializers.NewDatabaseInitializer(a.opts, a.logger)
    a.bootstrap.Register(a.dbInit)

    // 400: Redis
    // 500: NATS
    // 600: HTTP Server

    // 900: Health Check (最后)
    healthAddr := fmt.Sprintf(":%d", a.opts.GetHealthPort())
    a.healthInit = pkginitializers.NewHealthCheckInitializer(healthAddr, a.logger)
    a.bootstrap.Register(a.healthInit)
}
```

---

## 📋 Simple 模式清单

### 1. 创建目录结构
```bash
cmd/{service}/
├── main.go
└── app/
    ├── app.go
    └── server.go

internal/{service}/
└── config/
    └── options.go
```

### 2. main.go 模板
```go
package main

import (
    _ "go.uber.org/automaxprocs/maxprocs"
    "github.com/kart-io/k8s-agent/cmd/{service}/app"
)

func main() {
    app.Execute()
}
```

### 3. options.go 必需方法
```go
type Options struct {
    // ...
}

func NewOptions() *Options {}
func (o *Options) Validate() error {}
func (o *Options) InitLogger() (core.Logger, error) {}
func (o *Options) GetHealthPort() int {}
```

### 4. app.go 核心结构
```go
func Execute() {
    opts := config.NewOptions()

    runFunc := func(opts commonapp.Options) error {
        return run(opts.(*config.Options))
    }

    commonapp.RunWithOptions(opts, runFunc, commonapp.CommandConfig{
        Use:       "{service}",
        Short:     "{Service} Service",
        Long:      "...",
        EnvPrefix: "{SERVICE}",
    },
        commonapp.WithHealthCheck(commonapp.DefaultHealthCheckFuncFromOptions(opts)),
        commonapp.WithPrintVersion(),
        commonapp.WithPrintRuntime(),
        commonapp.WithWatch(),
    )
}

func run(opts *config.Options) error {
    log, err := opts.InitLogger()
    if err != nil {
        return fmt.Errorf("failed to init logger: %w", err)
    }
    defer log.Flush()

    srv, err := NewServer(opts, log)
    if err != nil {
        return fmt.Errorf("failed to create server: %w", err)
    }

    ctx := context.Background()
    return srv.Run(ctx)
}
```

### 5. server.go 结构
```go
type Server struct {
    opts   *config.Options
    logger core.Logger
    // ...
}

func NewServer(opts *config.Options, logger core.Logger) (*Server, error) {}
func (s *Server) Run(ctx context.Context) error {}
```

---

## 🎨 代码风格速查

### 命名规范
```go
✅ Good:
type AgentManagerApp struct {}
func NewDatabaseInitializer() *DatabaseInitializer
var errInitFailed = errors.New("...")

❌ Bad:
type App struct {}
func NewDB() *DB
var failed = errors.New("...")
```

### 注释规范
```go
✅ Good:
// DatabaseInitializer 负责初始化数据库连接和 schema
// 它依赖于配置中的 Database 选项
type DatabaseInitializer struct {}

❌ Bad:
// DB init
type DatabaseInitializer struct {}
```

### 错误处理
```go
✅ Good:
if err := component.Init(); err != nil {
    return fmt.Errorf("failed to initialize %s: %w", component.Name(), err)
}

❌ Bad:
if err := component.Init(); err != nil {
    return err
}
```

---

## 🔢 优先级参考

| 组件类型 | 优先级 | 说明 |
|---------|-------|------|
| 日志系统 | 100 | 最早初始化 |
| 数据库 | 300 | 核心依赖 |
| Redis | 400 | 缓存 |
| 注册中心 | 450 | 服务发现 |
| NATS | 500 | 消息队列 |
| 业务组件 | 550 | 业务逻辑 |
| HTTP Server | 600 | Web 服务 |
| gRPC Server | 700 | RPC 服务 |
| 健康检查 | 900 | 最后初始化 |

**规则**: 数字越小优先级越高，被依赖的组件应该有更高的优先级。

---

## 🧪 快速测试

### 1. 编译测试
```bash
make build SERVICE={service}
```

### 2. 启动测试
```bash
./bin/{service} --help
./bin/{service} --version
./bin/{service} --config configs/{service}/config-dev.yaml
```

### 3. 健康检查测试
```bash
# 启动服务
./bin/{service} --config configs/{service}/config-dev.yaml &
PID=$!

# 等待启动
sleep 5

# 测试健康检查
curl http://localhost:{health_port}/healthz
curl http://localhost:{health_port}/readyz

# 关闭服务
kill $PID
```

### 4. 环境变量测试
```bash
export {SERVICE}_SERVER_PORT=9999
./bin/{service} --config configs/{service}/config-dev.yaml
```

---

## 📚 常用命令

### 创建新服务（Bootstrap 模式）
```bash
# 1. 创建目录
mkdir -p cmd/{service}/app/options
mkdir -p internal/{service}/initializers

# 2. 复制模板
cp templates/bootstrap/main.go cmd/{service}/
cp templates/bootstrap/app.go cmd/{service}/app/
cp templates/bootstrap/options.go cmd/{service}/app/options/

# 3. 替换占位符
sed -i '' 's/{service}/myservice/g' cmd/myservice/**/*.go
sed -i '' 's/{Service}/MyService/g' cmd/myservice/**/*.go
sed -i '' 's/{SERVICE}/MYSERVICE/g' cmd/myservice/**/*.go
```

### 创建新服务（Simple 模式）
```bash
# 1. 创建目录
mkdir -p cmd/{service}/app
mkdir -p internal/{service}/config

# 2. 复制模板
cp templates/simple/main.go cmd/{service}/
cp templates/simple/app.go cmd/{service}/app/
cp templates/simple/server.go cmd/{service}/app/
cp templates/simple/options.go internal/{service}/config/

# 3. 替换占位符
sed -i '' 's/{service}/myservice/g' cmd/myservice/**/*.go
```

---

## ⚠️ 常见错误

### 错误 1: 忘记实现必需方法
```
Error: undefined: opts.InitLogger

解决: 确保 Options 实现了所有必需方法：
- Validate()
- InitLogger()
- GetHealthPort()
```

### 错误 2: 组件依赖顺序错误
```
Error: component X requires Y but Y is not initialized

解决: 检查 registerComponents() 中的注册顺序
     被依赖的组件应该先注册（优先级更高）
```

### 错误 3: 健康检查端口冲突
```
Error: bind: address already in use

解决: 确保每个服务使用不同的健康检查端口
     在配置文件中设置不同的 health.port
```

### 错误 4: 日志未初始化
```
Error: nil pointer dereference in logger

解决: 确保在使用 logger 之前调用 InitLogger()
     Bootstrap 模式: 在 Initialize() 中初始化
     Simple 模式: 在 run() 开始时初始化
```

---

## 📞 获取帮助

1. **查看完整文档**: `docs/refactoring/SERVICE_ENTRY_STANDARDIZATION.md`
2. **查看示例服务**:
   - Bootstrap 模式: `cmd/agent-manager/`
   - Simple 模式: `cmd/gateway/`
3. **查看框架文档**: `pkg/app/README.md`
4. **查看 Bootstrap 文档**: `pkg/bootstrap/README.md`

---

## ✅ 检查清单

提交代码前，确保：

- [ ] main.go 有完整的 automaxprocs 注释
- [ ] 入口函数名为 `Execute()`
- [ ] Options 实现了所有必需方法
- [ ] 组件优先级设置合理
- [ ] 错误信息包含上下文
- [ ] 所有公开函数有注释
- [ ] 代码通过 `gofmt` 格式化
- [ ] 代码通过 `golangci-lint` 检查
- [ ] 服务能正常启动和关闭
- [ ] 健康检查端点正常工作

---

**最后更新**: 2025-10-29

