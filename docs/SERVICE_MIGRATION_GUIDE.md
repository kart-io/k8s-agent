# 服务迁移到增强 pkg/app 框架指南

## 概述

本指南介绍如何将现有服务迁移到增强后的 pkg/app 框架，以享受健康检查、配置监听等新功能。

## 迁移策略

项目中存在两种应用模式：

### 模式 1: 简单模式（如 auth 服务）
- 使用 `app.Run()` 或 `app.RunWithOptions()`
- 直接的 RunFunc 回调
- 适合简单服务

### 模式 2: Application 接口模式（如 agent-manager）
- 使用 `app.RunWithRunner()`
- 实现 Application 接口
- 使用 Bootstrap 组件管理
- 适合复杂服务

## 模式 1: 简单模式迁移

### 迁移前

```go
func main() {
    opts := options.NewServerOptions()

    app.Run(opts, run, app.CommandConfig{
        Use:       "my-service",
        Short:     "My Service",
        EnvPrefix: "MY",
    })
}

func run(opts app.Options) error {
    // 初始化和运行逻辑
    return nil
}
```

### 迁移后

```go
func main() {
    opts := options.NewServerOptions()

    // 使用 RunWithOptions 启用增强功能
    app.RunWithOptions(opts, run, app.CommandConfig{
        Use:       "my-service",
        Short:     "My Service",
        EnvPrefix: "MY",
    },
        // 启用健康检查
        app.WithHealthCheck(app.DefaultHealthCheckFunc(":8090")),
        // 启用版本信息
        app.WithPrintVersion(),
        // 启用运行时信息
        app.WithPrintRuntime(),
        // 启用配置监听
        app.WithWatch(),
    )
}

func run(opts app.Options) error {
    // 无需修改
    return nil
}
```

### 修改点

1. ✅ 将 `app.Run()` 改为 `app.RunWithOptions()`
2. ✅ 添加功能选项（按需）
3. ✅ 添加 automaxprocs 导入到 main.go（可选）

**修改文件**: 1 个（`cmd/xxx/app/xxx.go` 或 `main.go`）

## 模式 2: Application 接口模式迁移

### 当前状态

agent-manager、orchestrator、reasoning 等服务使用此模式。

**特点**:
- 使用 `RunWithRunner()` 函数
- 实现 `Application` 接口
- 使用 `bootstrap.Bootstrap` 管理组件
- 复杂的初始化流程

### 增强方案

为 Application 模式添加可选功能支持：

#### 方案 A: 保持现有模式，添加健康检查

在 `Initialize()` 或 `Run()` 中手动启动健康检查：

```go
func (a *MyApp) Run(ctx context.Context) error {
    // 启动健康检查服务器
    healthServer := app.NewDefaultHealthCheckServer(":8090")
    if err := healthServer.Start(); err != nil {
        return err
    }

    // 原有逻辑
    return a.bootstrap.Run(ctx, nil)
}
```

**优点**: 最小修改
**缺点**: 不统一，需要每个服务自己管理

#### 方案 B: 扩展 RunWithRunner 支持功能选项

修改 `pkg/app/runner.go`，添加选项支持：

```go
// RunWithRunnerOptions 支持功能选项的 RunWithRunner
func RunWithRunnerOptions(
    opts Options,
    app Application,
    loggerInit LoggerInitFunc,
    cfg CommandConfig,
    appOpts ...AppOption,
) {
    // 内部逻辑
}
```

**优点**: 统一的 API
**缺点**: 需要修改 runner.go

### 推荐: 方案 A + 统一健康检查组件

创建统一的健康检查初始化器，在 Bootstrap 中注册。

## 实施步骤

### Step 1: 添加 automaxprocs（所有服务）

为每个服务的 main.go 添加：

```go
package main

import (
    // 自动适配容器 CPU 限制
    _ "go.uber.org/automaxprocs/maxprocs"

    "github.com/kart-io/k8s-agent/cmd/xxx/app"
)

func main() {
    app.Execute()
}
```

### Step 2: 迁移简单模式服务

对于使用 `app.Run()` 的服务（目前只有 auth）：

1. 将 `app.Run()` 改为 `app.RunWithOptions()`
2. 添加所需的功能选项
3. 测试验证

### Step 3: 为 Application 模式添加健康检查

对于使用 `RunWithRunner()` 的服务：

#### 创建健康检查初始化器

```go
// internal/xxx/initializers/health.go
package initializers

import (
    "context"
    "fmt"

    "github.com/kart-io/k8s-agent/pkg/app"
    "github.com/kart-io/k8s-agent/pkg/bootstrap"
    "github.com/kart-io/logger/core"
)

type HealthCheckInitializer struct {
    logger core.Logger
    port   string
    server *app.DefaultHealthCheckServer
}

func NewHealthCheckInitializer(port string, logger core.Logger) *HealthCheckInitializer {
    return &HealthCheckInitializer{
        logger: logger,
        port:   port,
    }
}

func (h *HealthCheckInitializer) Name() string {
    return "HealthCheck"
}

func (h *HealthCheckInitializer) Priority() int {
    return bootstrap.PriorityLowest // 最低优先级，最后启动
}

func (h *HealthCheckInitializer) Initialize(ctx context.Context) error {
    h.server = app.NewDefaultHealthCheckServer(h.port)
    if err := h.server.Start(); err != nil {
        return fmt.Errorf("failed to start health check server: %w", err)
    }

    h.logger.Infow("Health check server started", "port", h.port)
    return nil
}

func (h *HealthCheckInitializer) Shutdown(ctx context.Context) error {
    if h.server != nil {
        return h.server.Stop()
    }
    return nil
}
```

#### 在 App 中注册

```go
func (a *MyApp) registerComponents() {
    // ... 其他初始化器 ...

    // 添加健康检查
    healthInit := initializers.NewHealthCheckInitializer(":8090", a.logger)
    a.bootstrap.Register(healthInit)
}
```

### Step 4: 测试验证

```bash
# 构建
make build-xxx

# 运行
./xxx

# 测试健康检查
curl http://localhost:8090/healthz
```

## 服务迁移清单

### Auth ✅
- [x] 添加 automaxprocs
- [x] 使用 RunWithOptions
- [x] 启用健康检查
- [x] 启用版本/运行时信息
- [x] 启用配置监听
- [x] 测试验证

### Agent-Manager ⏳
- [ ] 添加 automaxprocs
- [ ] 创建健康检查初始化器
- [ ] 注册到 Bootstrap
- [ ] 测试验证

### Orchestrator ⏳
- [ ] 添加 automaxprocs
- [ ] 创建健康检查初始化器
- [ ] 注册到 Bootstrap
- [ ] 测试验证

### Reasoning ⏳
- [ ] 添加 automaxprocs
- [ ] 创建健康检查初始化器
- [ ] 注册到 Bootstrap
- [ ] 测试验证

### Gateway ⏳
- [ ] 评估迁移需求
- [ ] 按需实施

### Monitor ⏳
- [ ] 评估迁移需求
- [ ] 按需实施

### Cluster ⏳
- [ ] 评估迁移需求
- [ ] 按需实施

### Collect-Agent ⏳
- [ ] 评估迁移需求
- [ ] 按需实施

## 健康检查端口分配

为避免端口冲突，建议各服务使用不同端口：

| 服务 | 主端口 | 健康检查端口 |
|------|--------|--------------|
| auth | 8080 | 8090 |
| agent-manager | 8080 | 8091 |
| orchestrator | 8081 | 8092 |
| reasoning | 8082 | 8093 |
| gateway | 8000 | 8094 |
| monitor | 8083 | 8095 |
| cluster | 8084 | 8096 |
| collect-agent | - | 8097 |

## 配置文件更新

如果需要配置健康检查端口，可在配置文件中添加：

```yaml
# config.yaml
health:
  enabled: true
  port: 8090
```

对应的 Options：

```go
type HealthOptions struct {
    Enabled bool   `json:"enabled" mapstructure:"enabled"`
    Port    string `json:"port" mapstructure:"port"`
}
```

## Kubernetes 集成

### Deployment 配置

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth
spec:
  template:
    spec:
      containers:
      - name: auth
        image: auth:latest
        ports:
        - name: http
          containerPort: 8080
        - name: health
          containerPort: 8090
        livenessProbe:
          httpGet:
            path: /healthz
            port: health
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /readyz
            port: health
          initialDelaySeconds: 5
          periodSeconds: 5
```

## 最佳实践

### 1. 健康检查逻辑

**简单服务**: 使用默认健康检查即可

**复杂服务**: 实现自定义健康检查

```go
func healthCheck() error {
    // 检查数据库
    if err := checkDatabase(); err != nil {
        return err
    }

    // 检查依赖服务
    if err := checkDependencies(); err != nil {
        return err
    }

    return nil
}
```

### 2. 配置监听

**开发环境**: 启用配置监听，提高开发效率

**生产环境**: 谨慎使用，建议通过环境变量或重启更新配置

### 3. 版本信息

**所有服务**: 建议启用版本信息打印，便于排查问题

### 4. 运行时信息

**调试时**: 启用运行时信息
**生产环境**: 可关闭以减少日志

## 故障排除

### 健康检查端口冲突

```
Error: listen tcp :8090: bind: address already in use
```

**解决**: 为每个服务分配不同端口

### 配置重载失败

**原因**: 配置格式错误或权限问题
**解决**: 检查配置文件格式和读取权限

### Bootstrap 初始化失败

**原因**: 初始化器顺序或依赖问题
**解决**: 检查 Priority 设置和依赖关系

## 总结

- ✅ Auth 服务已完成迁移，可作为参考
- ⏳ Application 模式服务建议添加健康检查初始化器
- 📝 所有服务建议添加 automaxprocs
- 🎯 渐进式迁移，不破坏现有功能

详细使用请参考:
- [PKG_APP_USAGE_GUIDE.md](PKG_APP_USAGE_GUIDE.md)
- [PKG_APP_ENHANCEMENT_SUMMARY.md](PKG_APP_ENHANCEMENT_SUMMARY.md)
