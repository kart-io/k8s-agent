# pkg/app 框架增强方案

## 概述

通过分析 onexstack/pkg/app 的实现，我们可以增强 k8s-agent 的 pkg/app 框架，使其更加强大和功能完整。

## onexstack/pkg/app 核心特性分析

### 1. App 结构设计

```go
type App struct {
    name        string              // 应用名称
    shortDesc   string              // 短描述
    description string              // 详细描述
    run         RunFunc             // 运行函数
    cmd         *cobra.Command      // Cobra 命令
    args        cobra.PositionalArgs // 参数验证

    // 可选功能
    healthCheckFunc   HealthCheckFunc // 健康检查
    options          any              // 配置选项（支持两种接口）
    silence          bool             // 静默模式
    noConfig         bool             // 禁用配置文件
    watch            bool             // 配置文件监听
    contextExtractors map[string]func(context.Context) string // 日志上下文提取
}
```

**优势**:
- 功能选项模式 (Functional Options Pattern)
- 高度可配置
- 支持多种选项接口
- 内置健康检查支持

### 2. 两种配置接口支持

#### NamedFlagSetOptions 接口
```go
type NamedFlagSetOptions interface {
    Flags() cliflag.NamedFlagSets  // 返回分组的 flag sets
    Complete() error                 // 完成配置
    Validate() error                 // 验证配置
}
```

**优势**: 支持 Kubernetes 风格的分组帮助信息

#### FlagSetOptions 接口
```go
type FlagSetOptions interface {
    AddFlags(fs *pflag.FlagSet)  // 添加 flags
    Complete() error              // 完成配置
    Validate() error              // 验证配置
}
```

**优势**: 简单直接，适合普通应用

### 3. 关键功能特性

#### 3.1 自动配置管理
- Viper 集成
- 自动环境变量绑定
- 配置文件监听（可选）
- 多路径配置搜索

#### 3.2 健康检查
- 内置健康检查服务器
- 自定义健康检查函数
- 默认健康检查实现

#### 3.3 日志集成
- 自动日志初始化
- 支持自定义上下文提取器
- 配置信息打印
- Golang 运行时信息打印

#### 3.4 版本管理
- 自动版本信息打印
- --version 标志支持
- 版本信息结构化输出

#### 3.5 错误处理
- 优化的 Cobra 错误显示
- 先显示用法，再显示错误
- 静默模式支持

## 当前 k8s-agent/pkg/app 分析

### 现有功能
```go
type Options interface {
    Complete() error
    Validate() []error  // 注意：返回 []error
    AddFlags(fs *pflag.FlagSet)
}

type RunFunc func(opts Options) error  // 注意：接收 Options

func Run(opts Options, runFunc RunFunc, cfg CommandConfig)
func NewCommand(opts Options, runFunc RunFunc, cfg CommandConfig) *cobra.Command
```

**优势**:
- 简单直接
- 符合项目现有风格
- Viper 集成良好

**不足**:
- 缺少健康检查支持
- 缺少日志上下文提取
- 缺少配置文件监听
- 缺少静默模式
- 不支持 NamedFlagSets

## 增强方案设计

### 方案1: 完全复制 onexstack 模式 (不推荐)

**原因**:
- 引入 Kubernetes 依赖 (component-base)
- 与现有代码风格差异大
- 迁移成本高
- 过度工程化

### 方案2: 渐进式增强 (推荐)

保持当前简洁的设计，逐步添加有价值的特性。

#### 2.1 增强的 Options 接口

```go
// 保持现有接口
type Options interface {
    Complete() error
    Validate() []error
    AddFlags(fs *pflag.FlagSet)
}

// 添加可选接口（通过类型断言检测）
type HealthCheckOptions interface {
    HealthCheck() error
}

type LoggerOptions interface {
    InitLogger() (core.Logger, error)
}

type WatchConfigOptions interface {
    WatchConfig() bool
}
```

#### 2.2 增强的 App 结构

```go
type App struct {
    opts      Options
    runFunc   RunFunc
    config    CommandConfig
    cmd       *cobra.Command

    // 新增可选功能
    healthCheckFunc func() error
    logger          core.Logger
    silence         bool
    watch           bool
}

// 功能选项
type AppOption func(*App)

func WithHealthCheck(fn func() error) AppOption
func WithSilence() AppOption
func WithWatch() AppOption
func WithLogger(logger core.Logger) AppOption
```

#### 2.3 增强的 Run 函数

```go
// 保持兼容性
func Run(opts Options, runFunc RunFunc, cfg CommandConfig) {
    RunWithOptions(opts, runFunc, cfg)
}

// 新增带选项的版本
func RunWithOptions(opts Options, runFunc RunFunc, cfg CommandConfig, appOpts ...AppOption)

// 新增 App 构建器模式
func NewApp(opts Options, runFunc RunFunc, cfg CommandConfig, appOpts ...AppOption) *App
```

### 2.4 增强功能清单

#### 必须添加的功能 ✅
1. **健康检查支持**
   - 可选的健康检查函数
   - 默认健康检查服务器

2. **配置文件监听**
   - 可选启用 Viper.WatchConfig
   - 配置变更通知

3. **静默模式**
   - 禁止启动信息打印
   - 禁止配置信息打印

4. **日志增强**
   - Golang 运行时信息打印
   - 启动信息结构化

#### 可选添加的功能 ⚠️
1. **NamedFlagSets 支持**
   - 需要权衡是否引入 K8s 依赖
   - 可以实现轻量级版本

2. **Context Extractors**
   - 日志上下文提取
   - Trace ID, User ID 等

## 实现优先级

### Phase 1: 核心增强 (本次实现)
- [x] 健康检查支持
- [x] 静默模式
- [x] 配置文件监听
- [x] 运行时信息打印

### Phase 2: 高级功能 (后续)
- [ ] NamedFlagSets 轻量级实现
- [ ] Context Extractors
- [ ] Signal 处理增强
- [ ] Graceful Shutdown

### Phase 3: 工具增强 (可选)
- [ ] 配置验证器
- [ ] 配置迁移工具
- [ ] 配置文档生成

## 向后兼容性保证

### 1. 保持现有API
```go
// 现有函数保持不变
func Run(opts Options, runFunc RunFunc, cfg CommandConfig)
func NewCommand(opts Options, runFunc RunFunc, cfg CommandConfig) *cobra.Command
```

### 2. 新增可选API
```go
// 新函数，不影响现有代码
func RunWithOptions(opts Options, runFunc RunFunc, cfg CommandConfig, appOpts ...AppOption)
func NewApp(...) *App
```

### 3. 可选接口
```go
// 通过类型断言检测，不强制实现
if hc, ok := opts.(HealthCheckOptions); ok {
    // 使用健康检查
}
```

## 实现示例

### 增强后的使用方式

```go
// 方式1: 保持原有简单方式（完全兼容）
func main() {
    opts := options.NewServerOptions()
    app.Run(opts, run, app.CommandConfig{
        Use:       "my-service",
        Short:     "My Service",
        EnvPrefix: "MY",
    })
}

// 方式2: 使用新增的功能选项
func main() {
    opts := options.NewServerOptions()
    app.RunWithOptions(opts, run, app.CommandConfig{
        Use:       "my-service",
        Short:     "My Service",
        EnvPrefix: "MY",
    },
        app.WithHealthCheck(healthCheck),
        app.WithWatch(),
        app.WithSilence(),
    )
}

// 方式3: 使用 App 构建器（最灵活）
func main() {
    opts := options.NewServerOptions()
    application := app.NewApp(opts, run, app.CommandConfig{
        Use:       "my-service",
        Short:     "My Service",
        EnvPrefix: "MY",
    },
        app.WithHealthCheck(healthCheck),
        app.WithWatch(),
    )

    // 可以进一步配置
    application.Silence(true)

    // 运行
    application.Run()
}
```

## 代码量估算

- `app/app.go` 增强: ~150 行
- `app/options.go` 新增接口: ~30 行
- `app/health.go` 健康检查: ~50 行
- `app/config.go` 配置监听增强: ~30 行
- 总计: ~260 行代码

## 风险评估

### 低风险 ✅
- 向后兼容性强
- 可选功能，不影响现有代码
- 渐进式实现

### 中风险 ⚠️
- 代码复杂度增加
- 需要更多测试
- 文档需要更新

### 高风险 ❌
- 无（因为保持向后兼容）

## 推荐决策

### 推荐方案: Phase 1 渐进式增强

**理由**:
1. **最小化风险**: 保持向后兼容
2. **实用主义**: 添加真正需要的功能
3. **渐进式**: 可以分阶段实现
4. **项目适配**: 符合 k8s-agent 的代码风格

**不推荐**:
- 完全复制 onexstack 实现（过度工程化）
- 引入 Kubernetes component-base 依赖（依赖过重）
- 破坏现有 API（影响其他服务）

## 实现检查清单

- [ ] 设计增强的 App 结构
- [ ] 实现功能选项模式
- [ ] 添加健康检查支持
- [ ] 添加静默模式
- [ ] 添加配置监听
- [ ] 添加运行时信息打印
- [ ] 编写单元测试
- [ ] 更新文档
- [ ] 验证向后兼容性
- [ ] 在 auth 服务中试用

## 结论

**建议**: 采用 Phase 1 渐进式增强方案，为 pkg/app 框架添加健康检查、静默模式、配置监听等实用功能，同时保持向后兼容性。

这样既能学习 onexstack 的优秀设计，又不会引入不必要的复杂性和依赖。
