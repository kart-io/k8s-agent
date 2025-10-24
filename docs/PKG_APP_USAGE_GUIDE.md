# pkg/app 框架增强功能使用指南

## 概述

pkg/app 框架已经完成 Phase 1 增强，新增了以下功能：

- ✅ 健康检查支持
- ✅ 静默模式
- ✅ 配置文件监听
- ✅ 版本信息打印
- ✅ 运行时信息打印
- ✅ 向后兼容性

## 新增功能

### 1. 功能选项模式

使用 `RunWithOptions` 函数启用可选功能：

```go
commonapp.RunWithOptions(opts, run, config,
    commonapp.WithHealthCheck(healthCheckFunc),
    commonapp.WithPrintVersion(),
    commonapp.WithPrintRuntime(),
    commonapp.WithWatch(),
    commonapp.WithSilence(),
)
```

### 2. 健康检查

#### 2.1 使用默认健康检查

```go
commonapp.WithHealthCheck(commonapp.DefaultHealthCheckFunc(":8090"))
```

默认健康检查会启动一个 HTTP 服务器，提供以下端点：
- `GET /healthz` - 健康检查
- `GET /readyz` - 就绪检查

#### 2.2 自定义健康检查

```go
func myHealthCheck() error {
    // 检查数据库连接
    if err := db.Ping(); err != nil {
        return err
    }

    // 检查 Redis 连接
    if err := redis.Ping(); err != nil {
        return err
    }

    return nil
}

commonapp.WithHealthCheck(myHealthCheck)
```

### 3. 版本和运行时信息

#### 3.1 打印版本信息

```go
commonapp.WithPrintVersion()
```

输出示例：
```
Starting aetherius-auth v1.0.0-dev+e428cdb9
```

#### 3.2 打印运行时信息

```go
commonapp.WithPrintRuntime()
```

输出示例：
```
Go Version: go1.25.0
GOOS: linux, GOARCH: amd64
NumCPU: 8, GOMAXPROCS: 8
GOGC: 100
```

### 4. 配置文件监听

#### 4.1 启用配置文件监听

```go
commonapp.WithWatch()
```

当配置文件变化时，会自动重新加载：
```
Config file changed: /path/to/config.yaml
Config reloaded successfully
```

#### 4.2 通过 Options 接口控制

实现 `ConfigWatcher` 接口：

```go
type MyOptions struct {
    // ...
}

func (o *MyOptions) WatchConfig() bool {
    return true  // 启用配置监听
}
```

### 5. 静默模式

#### 5.1 使用功能选项

```go
commonapp.WithSilence()
```

#### 5.2 通过 Options 接口控制

实现 `SilenceMode` 接口：

```go
func (o *MyOptions) IsSilence() bool {
    return viper.GetBool("silent")
}
```

### 6. 可选接口

pkg/app 定义了多个可选接口，Options 可以选择性实现：

#### 6.1 HealthCheckProvider

```go
type HealthCheckProvider interface {
    HealthCheck() error
}
```

#### 6.2 ConfigWatcher

```go
type ConfigWatcher interface {
    WatchConfig() bool
}
```

#### 6.3 SilenceMode

```go
type SilenceMode interface {
    IsSilence() bool
}
```

#### 6.4 StartupInfoPrinter

```go
type StartupInfoPrinter interface {
    PrintStartupInfo()
}
```

## 使用示例

### 示例 1: 基础用法（向后兼容）

```go
func main() {
    opts := options.NewServerOptions()

    // 保持原有简单方式
    app.Run(opts, run, app.CommandConfig{
        Use:       "my-service",
        Short:     "My Service",
        EnvPrefix: "MY",
    })
}
```

### 示例 2: 使用增强功能

```go
func main() {
    opts := options.NewServerOptions()

    // 使用增强功能
    app.RunWithOptions(opts, run, app.CommandConfig{
        Use:       "my-service",
        Short:     "My Service",
        EnvPrefix: "MY",
    },
        app.WithHealthCheck(app.DefaultHealthCheckFunc(":8090")),
        app.WithPrintVersion(),
        app.WithPrintRuntime(),
        app.WithWatch(),
    )
}
```

### 示例 3: 使用 App 构建器

```go
func main() {
    opts := options.NewServerOptions()

    // 创建 App 实例
    application := app.NewApp(opts, run, app.CommandConfig{
        Use:       "my-service",
        Short:     "My Service",
        EnvPrefix: "MY",
    },
        app.WithHealthCheck(healthCheck),
        app.WithPrintVersion(),
        app.WithPrintRuntime(),
    )

    // 可以进一步配置或获取底层 Cobra 命令
    cmd := application.Command()
    cmd.AddCommand(versionCmd)

    // 运行
    application.Run()
}
```

### 示例 4: 实现可选接口

```go
type ServerOptions struct {
    Server   *commonoptions.ServerOptions
    Database *commonoptions.DatabaseOptions
    // ...
}

// 实现健康检查接口
func (o *ServerOptions) HealthCheck() error {
    // 自定义健康检查逻辑
    return nil
}

// 实现配置监听接口
func (o *ServerOptions) WatchConfig() bool {
    return true
}

// 实现静默模式接口
func (o *ServerOptions) IsSilence() bool {
    return viper.GetBool("silent")
}

// 使用时，App 会自动检测并调用这些方法
func main() {
    opts := NewServerOptions()

    app.Run(opts, run, app.CommandConfig{
        Use:   "my-service",
        Short: "My Service",
    })

    // App 会自动调用 opts.HealthCheck()
    // App 会根据 opts.WatchConfig() 决定是否监听配置
    // App 会根据 opts.IsSilence() 决定是否静默
}
```

## Auth 服务示例

auth 服务已经使用了增强功能：

```go
func NewApp() {
    opts := options.NewServerOptions()

    commonapp.RunWithOptions(opts, run, commonapp.CommandConfig{
        Use:       auth.Name,
        Short:     "Launch an Aetherius authentication and authorization server",
        Long:      commandDesc,
        EnvPrefix: "AUTH",
    },
        // 启用健康检查（监听 :8090）
        commonapp.WithHealthCheck(commonapp.DefaultHealthCheckFunc(":8090")),
        // 启用版本信息打印
        commonapp.WithPrintVersion(),
        // 启用运行时信息打印
        commonapp.WithPrintRuntime(),
        // 启用配置文件监听
        commonapp.WithWatch(),
    )
}
```

### 测试健康检查

```bash
# 启动 auth 服务
./auth

# 在另一个终端测试健康检查
curl http://localhost:8090/healthz
# 输出: ok

curl http://localhost:8090/readyz
# 输出: ready
```

### 测试配置监听

```bash
# 使用配置文件启动
./auth -c config.yaml

# 修改配置文件
vim config.yaml

# 观察日志输出
# Config file changed: config.yaml
# Config reloaded successfully
```

## API 对比

### 向后兼容的 API

```go
// 原有函数，完全兼容
func Run(opts Options, runFunc RunFunc, cfg CommandConfig)
func NewCommand(opts Options, runFunc RunFunc, cfg CommandConfig) *cobra.Command
func Execute(cmd *cobra.Command)
```

### 新增的 API

```go
// 新增：支持功能选项的运行函数
func RunWithOptions(opts Options, runFunc RunFunc, cfg CommandConfig, appOpts ...AppOption)

// 新增：App 构建器
func NewApp(opts Options, runFunc RunFunc, cfg CommandConfig, appOpts ...AppOption) *App

// App 方法
func (app *App) Run()
func (app *App) Command() *cobra.Command

// 功能选项
func WithHealthCheck(fn HealthCheckFunc) AppOption
func WithSilence() AppOption
func WithWatch() AppOption
func WithPrintVersion() AppOption
func WithPrintRuntime() AppOption

// 默认健康检查
func DefaultHealthCheckFunc(addr string) HealthCheckFunc
```

## 文件结构

```
pkg/app/
├── app.go           # 核心 App 结构和逻辑（增强版）
├── app.go.bak       # 原有实现备份
├── interfaces.go    # 可选接口定义
├── health.go        # 健康检查支持
├── runner.go        # ApplicationRunner（现有）
└── README.md        # 使用指南（现有）
```

## 迁移指南

### 从旧版本迁移

1. **不需要任何修改**：旧代码完全兼容，继续使用 `app.Run()` 即可

2. **启用新功能**：将 `app.Run()` 改为 `app.RunWithOptions()`

```diff
- app.Run(opts, run, config)
+ app.RunWithOptions(opts, run, config,
+     app.WithHealthCheck(app.DefaultHealthCheckFunc(":8090")),
+     app.WithPrintVersion(),
+ )
```

3. **实现可选接口**：在 Options 中实现需要的接口

```go
func (o *ServerOptions) HealthCheck() error {
    return checkDatabase()
}
```

## 性能影响

- **健康检查**: 启动一个轻量级 HTTP 服务器，几乎无性能影响
- **配置监听**: 使用 fsnotify，文件系统事件驱动，无性能影响
- **版本/运行时打印**: 仅在启动时执行一次，无影响
- **向后兼容**: 未使用的功能不会被执行，零开销

## 测试

```bash
# 构建
make build-auth

# 测试帮助信息
./auth --help

# 测试版本信息
./auth --version

# 测试运行（会看到启动信息和健康检查端口）
./auth

# 测试健康检查
curl http://localhost:8090/healthz

# 测试配置监听
./auth -c config.yaml
# 修改 config.yaml 并观察日志
```

## 故障排除

### 健康检查端口被占用

```
Error: listen tcp :8090: bind: address already in use
```

**解决方案**: 更改健康检查端口

```go
commonapp.WithHealthCheck(commonapp.DefaultHealthCheckFunc(":9090"))
```

### 配置重载失败

```
Failed to reload config: ...
```

**解决方案**: 检查配置文件格式和权限

## 最佳实践

1. **生产环境**: 启用健康检查，便于 Kubernetes liveness/readiness 探针
2. **开发环境**: 启用配置监听，提高开发效率
3. **调试**: 启用版本和运行时信息打印
4. **静默模式**: 在容器化部署时使用，减少日志输出

## 未来计划

### Phase 2 (可选)

- [ ] NamedFlagSets 轻量级实现
- [ ] Context Extractors（日志上下文）
- [ ] Signal 处理增强
- [ ] Graceful Shutdown 优化

### Phase 3 (可选)

- [ ] 配置验证增强
- [ ] 配置迁移工具
- [ ] 配置文档生成

## 参考

- [功能增强计划](PKG_APP_ENHANCEMENT_PLAN.md)
- [onexstack/pkg/app](https://github.com/onexstack/onex/tree/master/staging/src/github.com/onexstack/onexstack/pkg/app)
- [Cobra 文档](https://github.com/spf13/cobra)
- [Viper 文档](https://github.com/spf13/viper)
