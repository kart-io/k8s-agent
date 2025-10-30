# 服务入口标准化重构方案

## 📋 文档信息

- **文档版本**: v1.0
- **创建日期**: 2025-10-29
- **目标范围**: cmd/ 目录下所有服务入口
- **预计工作量**: 3-5 工作日
- **风险等级**: 中等

---

## 📊 一、现状分析

### 1.1 服务清单

项目当前包含 8 个微服务：

| 序号 | 服务名 | 职责描述 | 复杂度 |
|------|--------|----------|--------|
| 1 | agent-manager | Agent 管理服务，管理多集群 agent | 高 |
| 2 | auth | 认证授权服务 | 中 |
| 3 | cluster | 集群管理服务，提供多集群管理和 K8s 资源 API | 高 |
| 4 | collect-agent | 数据收集 Agent，监控边缘集群事件和指标 | 中 |
| 5 | gateway | API 网关服务，集成 Traefik | 低 |
| 6 | monitor | 监控服务，提供指标采集 | 低 |
| 7 | orchestrator | 编排服务，管理工作流执行 | 高 |
| 8 | reasoning | 推理服务，提供 AI 驱动的根因分析 | 高 |

### 1.2 当前架构模式分布

#### 模式 A: Bootstrap 模式（2个服务）
- **服务**: agent-manager, orchestrator
- **核心特征**:
  ```go
  // 使用 RunWithRunner
  commonapp.RunWithRunner(opts, &App{}, initLogger, config)

  // 实现 Application 接口
  type App struct {
      bootstrap *bootstrap.Bootstrap
      // ... 组件初始化器
  }

  func (a *App) Initialize(ctx context.Context, opts commonapp.Options) error
  func (a *App) Run(ctx context.Context) error
  func (a *App) Shutdown(ctx context.Context) error
  ```
- **优点**: 高度结构化，组件依赖清晰，生命周期管理完善
- **缺点**: 代码量大，学习曲线陡峭

#### 模式 B: Runner 模式（1个服务）
- **服务**: cluster
- **核心特征**:
  ```go
  // 使用 RunWithRunner，但不用 bootstrap
  commonapp.RunWithRunner(opts, &ClusterApp{}, initLogger, config)

  // 手动管理组件
  type ClusterApp struct {
      storage      *storage.MySQLStorage
      server       *api.Server
      healthServer *app.DefaultHealthCheckServer
  }
  ```
- **优点**: 灵活，适合中等复杂度服务
- **缺点**: 缺少统一的组件管理机制

#### 模式 C: Simple 模式（5个服务）
- **服务**: auth, collect-agent, gateway, monitor, reasoning
- **核心特征**:
  ```go
  // 使用 RunWithOptions + 简单函数
  commonapp.RunWithOptions(opts, run, config,
      commonapp.WithHealthCheck(...),
      commonapp.WithPrintVersion(),
      commonapp.WithWatch(),
  )

  func run(opts commonapp.Options) error {
      // 直接的线性逻辑
  }
  ```
- **优点**: 简单直接，易于理解
- **缺点**: 缺少结构化，不利于复杂场景扩展

---

## 🔍 二、问题识别

### 2.1 主要问题清单

#### P1 - 高优先级问题

| 问题ID | 问题描述 | 影响范围 | 影响程度 |
|--------|----------|----------|----------|
| P1-01 | 入口函数命名不统一（Execute vs NewApp） | auth | 代码可读性差 |
| P1-02 | 三种不同的架构模式共存 | 所有服务 | 维护成本高 |
| P1-03 | 日志初始化方式不统一（3种方式） | 所有服务 | 代码不一致 |
| P1-04 | 健康检查实现方式不统一（3种方式） | 所有服务 | 运维复杂度高 |

#### P2 - 中优先级问题

| 问题ID | 问题描述 | 影响范围 | 影响程度 |
|--------|----------|----------|----------|
| P2-01 | 目录结构不一致（options/ 子目录） | 6个服务 | 项目结构混乱 |
| P2-02 | main.go 注释完整度不一致 | cluster, monitor | 文档缺失 |
| P2-03 | 配置加载方式不统一 | 部分服务 | 配置管理混乱 |

### 2.2 问题示例

#### 问题案例 1: 入口函数不统一

```go
// ❌ 不一致的调用方式

// auth/main.go
func main() {
    app.NewApp()  // 唯一使用 NewApp
}

// 其他所有服务/main.go
func main() {
    app.Execute()  // 使用 Execute
}
```

**影响**: 新开发者容易困惑，代码风格不统一。

#### 问题案例 2: 日志初始化方式混乱

```go
// ❌ 方式 1: 通过 options 方法（agent-manager, auth, orchestrator）
logger, err := serverOpts.InitLogger()

// ❌ 方式 2: 直接使用 common logger（cluster, collect-agent, gateway, reasoning）
log, err := commonlogger.InitFromOptions(opts.Logging)

// ❌ 方式 3: 手动转换配置（monitor）
logOpts := &options.LoggingOptions{
    Engine:      "slog",
    Level:       opts.Logging.Level,
    Format:      opts.Logging.Format,
    OutputPaths: []string{opts.Logging.Output},
}
log, err := commonlogger.InitFromOptions(logOpts)
```

**影响**: 维护困难，容易出错。

#### 问题案例 3: 健康检查实现混乱

```go
// ❌ 方式 1: Bootstrap + 初始化器（agent-manager, orchestrator）
a.healthInit = pkginitializers.NewHealthCheckInitializer(healthAddr, a.logger)
a.bootstrap.Register(a.healthInit)

// ❌ 方式 2: 手动创建服务器（cluster）
a.healthServer = app.NewDefaultHealthCheckServer(":8096")
if err := a.healthServer.Start(); err != nil {
    return err
}

// ❌ 方式 3: 使用框架选项（auth, collect-agent, gateway, monitor, reasoning）
commonapp.WithHealthCheck(commonapp.DefaultHealthCheckFuncFromOptions(opts))
```

**影响**: 运维脚本需要处理多种健康检查端点逻辑。

---

## 🎯 三、重构目标

### 3.1 核心目标

1. **统一性**: 所有服务使用一致的架构模式和代码风格
2. **可维护性**: 代码结构清晰，易于理解和修改
3. **可扩展性**: 支持未来新增服务和功能
4. **标准化**: 建立清晰的服务开发规范

### 3.2 具体目标

| 目标 | 当前状态 | 目标状态 | 衡量标准 |
|------|----------|----------|----------|
| 入口函数命名 | 2种（Execute/NewApp） | 1种（Execute） | 100%统一 |
| 架构模式 | 3种混合 | 按复杂度分级 | 清晰分类 |
| 日志初始化 | 3种方式 | 1种方式 | 100%统一 |
| 健康检查 | 3种实现 | 1种实现 | 100%统一 |
| 目录结构 | 不一致 | 标准化 | 有明确规范 |
| 代码注释 | 不完整 | 完整 | 100%覆盖 |

### 3.3 非目标（不在本次重构范围）

- ❌ 修改业务逻辑
- ❌ 重构 internal/ 内部实现
- ❌ 修改配置文件格式
- ❌ 更改 API 接口

---

## 📐 四、重构方案

### 4.1 架构模式分级标准

根据服务复杂度，采用两级架构模式：

#### 🏗️ 高级模式 (Tier-1): Bootstrap 模式

**适用条件**（满足任意2条）:
- ✅ 有 3+ 个外部依赖（数据库、Redis、NATS等）
- ✅ 有复杂的组件依赖关系
- ✅ 需要精细控制组件启动顺序
- ✅ 有多个服务器（HTTP + gRPC + Health）
- ✅ 有消息订阅/发布逻辑

**推荐服务**:
- `agent-manager` ✅ (已实现)
- `orchestrator` ✅ (已实现)
- `cluster` 🔄 (需要重构)
- `reasoning` 🔄 (需要重构)

**核心结构**:
```go
// cmd/{service}/app/app.go
type {Service}App struct {
    bootstrap *bootstrap.Bootstrap
    opts      *options.ServerOptions
    logger    core.Logger

    // 组件初始化器
    dbInit     *initializers.DatabaseInitializer
    redisInit  *initializers.RedisInitializer
    // ... 其他初始化器
}

func Execute() {
    opts := options.NewServerOptions()
    commonapp.RunWithRunner(
        opts,
        &{Service}App{},
        initLogger,
        commonapp.CommandConfig{...},
    )
}

func (a *{Service}App) Initialize(ctx context.Context, opts commonapp.Options) error
func (a *{Service}App) Run(ctx context.Context) error
func (a *{Service}App) Shutdown(ctx context.Context) error
func (a *{Service}App) registerComponents()
```

**目录结构**:
```
cmd/{service}/
├── main.go
└── app/
    ├── app.go              # Application 实现
    └── options/
        └── options.go      # 配置选项
```

#### 🔧 标准模式 (Tier-2): Simple 模式

**适用条件**:
- ✅ 简单的 HTTP 服务
- ✅ 依赖少（0-2个外部依赖）
- ✅ 无复杂的启动顺序要求
- ✅ 单一服务器

**推荐服务**:
- `auth` 🔄 (需要调整)
- `collect-agent` ✅ (已实现)
- `gateway` ✅ (已实现)
- `monitor` ✅ (已实现)

**核心结构**:
```go
// cmd/{service}/app/app.go
func Execute() {
    opts := config.NewOptions()

    runFunc := func(opts commonapp.Options) error {
        return run(opts.(*config.Options))
    }

    commonapp.RunWithOptions(opts, runFunc, commonapp.CommandConfig{...},
        commonapp.WithHealthCheck(commonapp.DefaultHealthCheckFuncFromOptions(opts)),
        commonapp.WithPrintVersion(),
        commonapp.WithPrintRuntime(),
        commonapp.WithWatch(),
    )
}

func run(opts *config.Options) error {
    // 1. 初始化日志
    // 2. 创建服务器
    // 3. 启动服务器
}
```

**目录结构**:
```
cmd/{service}/
├── main.go
└── app/
    ├── app.go      # Execute 和 run 函数
    └── server.go   # Server 实现
```

### 4.2 统一规范

#### 规范 1: 入口函数命名

```go
// ✅ 标准: 所有服务统一使用 Execute
// cmd/{service}/main.go
package main

import (
    // Import the automaxprocs package, which automatically configures the GOMAXPROCS
    // value at program startup to match the Linux container's CPU quota.
    // This avoids performance issues caused by an inappropriate default GOMAXPROCS
    // value when running in containers, ensuring that the Go program can fully utilize
    // available CPU resources and avoid CPU waste.
    _ "go.uber.org/automaxprocs/maxprocs"

    "github.com/kart-io/k8s-agent/cmd/{service}/app"
)

func main() {
    app.Execute()
}
```

#### 规范 2: 日志初始化

```go
// ✅ 标准: 统一使用 options.InitLogger() 方法

// Tier-1 (Bootstrap 模式)
func initLogger(opts commonapp.Options) (core.Logger, error) {
    serverOpts := opts.(*options.ServerOptions)
    return serverOpts.InitLogger()
}

// Tier-2 (Simple 模式)
func run(opts *config.Options) error {
    log, err := opts.InitLogger()
    if err != nil {
        return fmt.Errorf("failed to initialize logger: %w", err)
    }
    defer log.Flush()
    // ...
}
```

**要求**: 所有 options 结构必须实现 `InitLogger() (core.Logger, error)` 方法。

#### 规范 3: 健康检查

```go
// ✅ Tier-1 服务: 使用 bootstrap + HealthCheckInitializer
func (a *{Service}App) registerComponents() {
    // ... 其他组件

    // Health Check Server (优先级最低，最后启动)
    healthPort := a.opts.GetHealthPort()
    healthAddr := fmt.Sprintf(":%d", healthPort)
    a.healthInit = pkginitializers.NewHealthCheckInitializer(healthAddr, a.logger)
    a.bootstrap.Register(a.healthInit)
}

// ✅ Tier-2 服务: 使用 WithHealthCheck 选项
commonapp.RunWithOptions(opts, runFunc, config,
    commonapp.WithHealthCheck(commonapp.DefaultHealthCheckFuncFromOptions(opts)),
    // ...
)
```

**要求**: 所有 options 结构必须实现 `GetHealthPort() int` 方法。

#### 规范 4: CommandConfig 标准

```go
// ✅ 标准配置格式
commonapp.CommandConfig{
    Use:       "{service}",           // 服务名称（小写，连字符分隔）
    Short:     "{Service} Service",   // 简短描述
    Long:      "{详细的多行描述}",      // 详细描述
    EnvPrefix: "{SERVICE}",           // 环境变量前缀（大写）
}
```

#### 规范 5: 目录结构

```
cmd/{service}/
├── main.go                    # 入口文件（统一格式）
└── app/
    ├── app.go                 # Execute 函数和主要逻辑
    ├── server.go              # Server 实现（Tier-2 需要）
    └── options/               # 配置选项（Tier-1 需要）
        └── options.go
```

### 4.3 服务分级决策

| 服务 | 当前模式 | 建议模式 | 理由 | 重构工作量 |
|------|----------|----------|------|-----------|
| agent-manager | Bootstrap | Bootstrap | ✅ 已符合标准 | 小（仅调整细节） |
| orchestrator | Bootstrap | Bootstrap | ✅ 已符合标准 | 小（仅调整细节） |
| cluster | Runner | Bootstrap | 复杂度高，多组件依赖 | 中 |
| reasoning | Simple | Bootstrap | 复杂度高，需要组件管理 | 中 |
| auth | Simple | Simple | 调整为标准 Simple 模式 | 小 |
| collect-agent | Simple | Simple | ✅ 已符合标准 | 极小 |
| gateway | Simple | Simple | ✅ 已符合标准 | 极小 |
| monitor | Simple | Simple | ✅ 已符合标准 | 小 |

---

## 🔨 五、实施计划

### 5.1 分阶段实施

#### 🎯 阶段一: 基础规范统一（预计 1 天）

**目标**: 统一所有服务的基础结构

**任务清单**:
- [ ] P1-01: 统一 auth 服务入口函数（Execute）
- [ ] P2-02: 补全 cluster 和 monitor 的 main.go 注释
- [ ] 创建标准化模板文件
- [ ] 更新本文档

**验收标准**:
- ✅ 所有 main.go 使用统一格式
- ✅ 所有 main.go 有完整的 automaxprocs 注释
- ✅ 所有服务调用 `app.Execute()`

#### 🎯 阶段二: Simple 模式标准化（预计 1 天）

**目标**: 统一所有 Simple 模式服务

**任务清单**:
- [ ] auth: 调整为标准 Simple 模式
- [ ] monitor: 统一日志初始化方式
- [ ] 验证 collect-agent, gateway 符合标准
- [ ] 为所有 Simple 模式服务添加统一的 WithOptions

**验收标准**:
- ✅ 所有 Simple 服务使用相同的代码结构
- ✅ 日志初始化方式统一
- ✅ 健康检查方式统一

#### 🎯 阶段三: Bootstrap 模式标准化（预计 2 天）

**目标**: 统一所有 Bootstrap 模式服务

**任务清单**:
- [ ] cluster: 重构为 Bootstrap 模式
  - [ ] 创建 initializers 包
  - [ ] 实现各组件初始化器
  - [ ] 使用 bootstrap 管理生命周期
- [ ] reasoning: 重构为 Bootstrap 模式
  - [ ] 创建 initializers 包
  - [ ] 实现各组件初始化器
  - [ ] 使用 bootstrap 管理生命周期
- [ ] 验证 agent-manager, orchestrator 符合标准
- [ ] 统一健康检查实现

**验收标准**:
- ✅ 所有 Bootstrap 服务使用相同的代码结构
- ✅ 使用 bootstrap.Bootstrap 管理组件
- ✅ 组件依赖关系清晰
- ✅ 优先级设置合理

#### 🎯 阶段四: 文档和验证（预计 1 天）

**任务清单**:
- [ ] 创建服务开发规范文档
- [ ] 创建新服务开发模板
- [ ] 编写测试用例验证重构
- [ ] 更新 CONTRIBUTING.md
- [ ] 代码审查

**验收标准**:
- ✅ 所有服务通过测试
- ✅ 文档完整
- ✅ 有清晰的开发指南

### 5.2 详细实施步骤

#### 步骤 1: auth 服务重构（阶段一）

**当前问题**:
- 使用 `NewApp()` 而非 `Execute()`
- app.go 中有冗余的 `Execute()` 包装函数

**重构步骤**:

```go
// 1. 修改 cmd/auth/app/app.go
// 删除：
func Execute() {
    NewApp()
}

// 修改 NewApp 为 Execute：
func Execute() {  // 改名
    opts := options.NewServerOptions()

    commonapp.RunWithOptions(opts, run, commonapp.CommandConfig{
        Use:       auth.Name,
        Short:     "Launch an Aetherius authentication and authorization server",
        Long:      commandDesc,
        EnvPrefix: "AUTH",
    },
        commonapp.WithHealthCheck(commonapp.DefaultHealthCheckFuncFromOptions(opts)),
        commonapp.WithPrintVersion(),
        commonapp.WithPrintRuntime(),
        commonapp.WithWatch(),
    )
}

// 2. 验证 cmd/auth/main.go 已经调用 app.Execute()
```

#### 步骤 2: cluster 服务重构（阶段三）

**当前问题**:
- 使用 Runner 模式但不用 bootstrap
- 手动管理组件生命周期
- 健康检查实现不统一

**重构步骤**:

```go
// 1. 创建 internal/cluster/initializers/ 包
//    - database.go: DatabaseInitializer
//    - http_server.go: HTTPServerInitializer

// 2. 重构 cmd/cluster/app/app.go
type ClusterApp struct {
    bootstrap *bootstrap.Bootstrap
    opts      *clusterconfig.Options
    logger    core.Logger

    // 组件初始化器
    dbInit         *initializers.DatabaseInitializer
    httpInit       *initializers.HTTPServerInitializer
    healthInit     *pkginitializers.HealthCheckInitializer
}

func (a *ClusterApp) registerComponents() {
    // 1. Database (优先级 300)
    a.dbInit = initializers.NewDatabaseInitializer(a.opts, a.logger)
    a.bootstrap.Register(a.dbInit)

    // 2. HTTP Server (优先级 600)
    a.httpInit = initializers.NewHTTPServerInitializer(a.opts, a.logger, a.dbInit)
    a.bootstrap.Register(a.httpInit)

    // 3. Health Check Server (优先级最低)
    healthPort := a.opts.GetHealthPort()
    healthAddr := fmt.Sprintf(":%d", healthPort)
    a.healthInit = pkginitializers.NewHealthCheckInitializer(healthAddr, a.logger)
    a.bootstrap.Register(a.healthInit)
}

// 3. 实现 Initialize, Run, Shutdown 方法
```

#### 步骤 3: reasoning 服务重构（阶段三）

**当前问题**:
- 使用 Simple 模式，但服务复杂度较高
- 有 LLM、向量存储等复杂依赖
- 需要更好的组件管理

**重构步骤**:

```go
// 1. 创建 internal/reasoning/initializers/ 包
//    - database.go: DatabaseInitializer
//    - llm.go: LLMInitializer
//    - memory.go: MemoryInitializer
//    - http_server.go: HTTPServerInitializer

// 2. 创建 cmd/reasoning/app/options/options.go
//    - 定义 ServerOptions
//    - 实现 InitLogger(), GetHealthPort() 等方法

// 3. 重构 cmd/reasoning/app/app.go 为 Bootstrap 模式
type ReasoningApp struct {
    bootstrap *bootstrap.Bootstrap
    opts      *options.ServerOptions
    logger    core.Logger

    // 组件初始化器
    dbInit       *initializers.DatabaseInitializer
    llmInit      *initializers.LLMInitializer
    memoryInit   *initializers.MemoryInitializer
    httpInit     *initializers.HTTPServerInitializer
    healthInit   *pkginitializers.HealthCheckInitializer
}
```

#### 步骤 4: monitor 服务调整（阶段二）

**当前问题**:
- 日志初始化需要手动转换配置

**重构步骤**:

```go
// 1. 在 internal/monitor/config/options.go 添加 InitLogger 方法
func (o *Options) InitLogger() (core.Logger, error) {
    logOpts := &options.LoggingOptions{
        Engine:      "slog",
        Level:       o.Logging.Level,
        Format:      o.Logging.Format,
        OutputPaths: []string{o.Logging.Output},
    }
    return commonlogger.InitFromOptions(logOpts)
}

// 2. 简化 cmd/monitor/app/app.go 中的 run 函数
func run(opts *config.Options) error {
    // 直接调用
    log, err := opts.InitLogger()
    if err != nil {
        return fmt.Errorf("failed to init logger: %w", err)
    }
    defer log.Flush()
    // ...
}
```

### 5.3 时间线

```
Week 1
├─ Day 1: 阶段一 - 基础规范统一
│  ├─ Morning: auth 入口重构 + main.go 注释补全
│  └─ Afternoon: 创建模板文件 + 文档更新
│
├─ Day 2: 阶段二 - Simple 模式标准化
│  ├─ Morning: auth 和 monitor 调整
│  └─ Afternoon: 验证其他 Simple 服务
│
├─ Day 3: 阶段三开始 - cluster 重构
│  ├─ Morning: 创建 cluster initializers
│  └─ Afternoon: 重构 cluster app.go
│
├─ Day 4: 阶段三继续 - reasoning 重构
│  ├─ Morning: 创建 reasoning initializers
│  └─ Afternoon: 重构 reasoning app.go
│
└─ Day 5: 阶段四 - 文档和验证
   ├─ Morning: 编写文档和测试
   └─ Afternoon: 代码审查和优化
```

---

## 🧪 六、测试计划

### 6.1 测试策略

#### 单元测试
- 每个初始化器的独立测试
- 配置加载测试
- 日志初始化测试

#### 集成测试
- 服务启动/关闭测试
- 健康检查端点测试
- 组件依赖关系测试

#### 回归测试
- 所有现有功能必须正常工作
- API 兼容性测试
- 配置文件兼容性测试

### 6.2 测试用例

#### 测试用例 1: 服务启动测试

```bash
# 测试所有服务能正常启动
for service in agent-manager auth cluster collect-agent gateway monitor orchestrator reasoning; do
    echo "Testing $service..."
    ./_output/bin/$service --help
    ./_output/bin/$service --version
done
```

#### 测试用例 2: 健康检查测试

```bash
# 启动服务并测试健康检查
./bin/agent-manager --config configs/agent-manager/config-dev.yaml &
PID=$!
sleep 5

# 测试健康检查端点
curl -f http://localhost:8091/healthz || exit 1
curl -f http://localhost:8091/readyz || exit 1

# 清理
kill $PID
```

#### 测试用例 3: 配置加载测试

```bash
# 测试环境变量配置
export AGENT_MANAGER_SERVER_PORT=9999
./bin/agent-manager --config configs/agent-manager/config-dev.yaml &
PID=$!
sleep 5

# 验证端口是否为 9999
netstat -tuln | grep 9999 || exit 1

kill $PID
```

### 6.3 验收标准

所有测试必须通过：

- ✅ 所有服务能正常启动和关闭
- ✅ 健康检查端点返回 200
- ✅ 日志输出格式正确
- ✅ 配置文件能正确加载
- ✅ 环境变量能正确覆盖配置
- ✅ 信号处理正常（SIGINT/SIGTERM）
- ✅ 所有原有功能正常工作

---

## ⚠️ 七、风险评估

### 7.1 技术风险

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|----------|
| 重构破坏现有功能 | 中 | 高 | 完整的回归测试 |
| Bootstrap 模式引入复杂度 | 低 | 中 | 充分的文档和示例 |
| 配置兼容性问题 | 低 | 中 | 保持配置文件格式不变 |
| 依赖关系处理错误 | 中 | 高 | 仔细设计初始化器优先级 |

### 7.2 时间风险

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|----------|
| cluster 重构时间超预期 | 中 | 中 | 预留缓冲时间 |
| reasoning 重构复杂度高 | 中 | 中 | 分步实施，可延后 |
| 测试覆盖不足 | 低 | 高 | 提前制定测试计划 |

### 7.3 回滚计划

如果重构出现严重问题：

1. **Git 回滚**: 使用 git revert 回滚所有提交
2. **分支策略**: 在 feature 分支进行重构，测试通过后再合并
3. **灰度发布**: 优先重构低风险服务（gateway, monitor）

---

## 📚 八、最佳实践

### 8.1 新服务开发指南

#### 决策树: 选择合适的模式

```
开始
  │
  ├─ 是否有 3+ 个外部依赖？
  │  ├─ 是 ─→ 使用 Bootstrap 模式
  │  └─ 否
  │      │
  │      ├─ 是否有复杂的组件依赖？
  │      │  ├─ 是 ─→ 使用 Bootstrap 模式
  │      │  └─ 否 ─→ 使用 Simple 模式
```

#### Bootstrap 模式开发清单

- [ ] 创建 `cmd/{service}/app/options/options.go`
  - [ ] 定义 ServerOptions 结构
  - [ ] 实现 `Validate() error`
  - [ ] 实现 `InitLogger() (core.Logger, error)`
  - [ ] 实现 `GetHealthPort() int`
  - [ ] 实现 `Config()` 转换方法

- [ ] 创建 `internal/{service}/initializers/` 包
  - [ ] 为每个组件创建独立的初始化器
  - [ ] 设置合理的优先级
  - [ ] 实现 `Initialize()`, `Priority()`, `Name()` 方法
  - [ ] 实现 `Shutdown()` 方法（如需要）

- [ ] 创建 `cmd/{service}/app/app.go`
  - [ ] 定义 App 结构体（包含 bootstrap）
  - [ ] 实现 `Initialize()` 方法
  - [ ] 实现 `Run()` 方法
  - [ ] 实现 `Shutdown()` 方法
  - [ ] 实现 `registerComponents()` 方法
  - [ ] 实现 `Execute()` 函数

#### Simple 模式开发清单

- [ ] 创建 `internal/{service}/config/options.go`
  - [ ] 定义 Options 结构
  - [ ] 实现 `Validate() error`
  - [ ] 实现 `InitLogger() (core.Logger, error)`
  - [ ] 实现 `GetHealthPort() int`

- [ ] 创建 `cmd/{service}/app/server.go`
  - [ ] 定义 Server 结构
  - [ ] 实现 `NewServer()` 构造函数
  - [ ] 实现 `Run()` 方法

- [ ] 创建 `cmd/{service}/app/app.go`
  - [ ] 实现 `Execute()` 函数
  - [ ] 实现 `run()` 函数

### 8.2 代码风格规范

#### 命名规范

```go
// ✅ 好的命名
type AgentManagerApp struct {}          // 使用完整服务名
func NewDatabaseInitializer()          // 明确的构造函数名
var errInitFailed = errors.New("...")  // 错误变量以 err 开头

// ❌ 避免的命名
type App struct {}                      // 太通用
func NewDB()                            // 缩写不清晰
var failed = errors.New("...")         // 不符合约定
```

#### 注释规范

```go
// ✅ 好的注释
// DatabaseInitializer 负责初始化数据库连接和 schema
// 它依赖于配置中的 Database 选项，并在初始化失败时返回详细的错误信息
type DatabaseInitializer struct {
    // ...
}

// ❌ 避免的注释
// DB init
type DatabaseInitializer struct {
    // ...
}
```

#### 错误处理规范

```go
// ✅ 好的错误处理
if err := component.Init(); err != nil {
    return fmt.Errorf("failed to initialize %s: %w", component.Name(), err)
}

// ❌ 避免的错误处理
if err := component.Init(); err != nil {
    return err  // 丢失上下文
}
```

### 8.3 组件优先级规范

标准的初始化优先级分配：

```go
const (
    PriorityInfrastructure = 100  // 基础设施（日志等）
    PriorityDatabase       = 300  // 数据库
    PriorityCache          = 400  // 缓存（Redis）
    PriorityRegistry       = 450  // 注册中心
    PriorityMessageQueue   = 500  // 消息队列
    PriorityBusinessLogic  = 550  // 业务逻辑组件
    PriorityHTTPServer     = 600  // HTTP 服务器
    PriorityGRPCServer     = 700  // gRPC 服务器
    PriorityHealthCheck    = 900  // 健康检查（最后）
)
```

**原则**:
1. 数字越小，优先级越高（越早初始化）
2. 被依赖的组件优先级应高于依赖它的组件
3. 服务器类组件应最后初始化
4. 健康检查应该最后初始化

---

## 📝 九、附录

### 9.1 服务对比表（重构前后）

#### agent-manager

| 项目 | 重构前 | 重构后 | 变化 |
|------|--------|--------|------|
| 入口函数 | Execute | Execute | ✅ 无变化 |
| 架构模式 | Bootstrap | Bootstrap | ✅ 无变化 |
| 日志初始化 | options.InitLogger() | options.InitLogger() | ✅ 无变化 |
| 健康检查 | HealthCheckInitializer | HealthCheckInitializer | ✅ 无变化 |
| 目录结构 | app/ + options/ | app/ + options/ | ✅ 无变化 |
| 工作量 | - | - | 极小（仅文档） |

#### auth

| 项目 | 重构前 | 重构后 | 变化 |
|------|--------|--------|------|
| 入口函数 | NewApp | Execute | 🔄 重命名 |
| 架构模式 | Simple | Simple | ✅ 无变化 |
| 日志初始化 | options.InitLogger() | options.InitLogger() | ✅ 无变化 |
| 健康检查 | WithHealthCheck | WithHealthCheck | ✅ 无变化 |
| 目录结构 | app/ + options/ | app/ + options/ | ✅ 无变化 |
| 工作量 | - | - | 小 |

#### cluster

| 项目 | 重构前 | 重构后 | 变化 |
|------|--------|--------|------|
| 入口函数 | Execute | Execute | ✅ 无变化 |
| 架构模式 | Runner | Bootstrap | 🔄 升级 |
| 日志初始化 | commonlogger.InitFromOptions | options.InitLogger() | 🔄 统一 |
| 健康检查 | DefaultHealthCheckServer | HealthCheckInitializer | 🔄 统一 |
| 目录结构 | app/app.go | app/ + options/ | ➕ 新增 |
| 工作量 | - | - | 中 |

#### reasoning

| 项目 | 重构前 | 重构后 | 变化 |
|------|--------|--------|------|
| 入口函数 | Execute | Execute | ✅ 无变化 |
| 架构模式 | Simple | Bootstrap | 🔄 升级 |
| 日志初始化 | logger.InitFromOptions | options.InitLogger() | 🔄 统一 |
| 健康检查 | WithHealthCheck | HealthCheckInitializer | 🔄 统一 |
| 目录结构 | app/ + server.go | app/ + options/ | 🔄 调整 |
| 工作量 | - | - | 中 |

### 9.2 模板文件

#### 模板 1: main.go（通用）

```go
package main

import (
	// Import the automaxprocs package, which automatically configures the GOMAXPROCS
	// value at program startup to match the Linux container's CPU quota.
	// This avoids performance issues caused by an inappropriate default GOMAXPROCS
	// value when running in containers, ensuring that the Go program can fully utilize
	// available CPU resources and avoid CPU waste.
	_ "go.uber.org/automaxprocs/maxprocs"

	"github.com/kart-io/k8s-agent/cmd/{service}/app"
)

func main() {
	app.Execute()
}
```

#### 模板 2: Bootstrap 模式 app.go

```go
package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/cmd/{service}/app/options"
	{service} "github.com/kart-io/k8s-agent/internal/{service}"
	"github.com/kart-io/k8s-agent/internal/{service}/initializers"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// Execute runs the {service} command
func Execute() {
	opts := options.NewServerOptions()

	commonapp.RunWithRunner(
		opts,
		&{Service}App{},
		initLogger,
		commonapp.CommandConfig{
			Use:       "{service}",
			Short:     "{Service} Service",
			Long:      "{Service} Service provides {description}",
			EnvPrefix: "{SERVICE}",
		},
	)
}

// {Service}App 实现 commonapp.Application 接口
type {Service}App struct {
	bootstrap *bootstrap.Bootstrap
	opts      *options.ServerOptions
	config    *{service}.Config
	logger    core.Logger

	// 组件初始化器
	dbInit     *initializers.DatabaseInitializer
	healthInit *pkginitializers.HealthCheckInitializer
}

// Initialize 初始化应用程序
func (a *{Service}App) Initialize(ctx context.Context, opts commonapp.Options) error {
	a.opts = opts.(*options.ServerOptions)

	// 初始化日志系统
	logger, err := initLogger(opts)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	a.logger = logger

	a.logger.Infow("Initializing {Service} Service",
		"port", a.opts.Server.Port,
	)

	// 转换为业务配置
	config, err := a.opts.Config()
	if err != nil {
		return fmt.Errorf("failed to build config: %w", err)
	}
	a.config = config

	// 创建 bootstrap 实例
	a.bootstrap = bootstrap.New(a.logger)

	// 注册所有组件初始化器
	a.registerComponents()

	a.logger.Infow("Components registered, ready to start")
	return nil
}

// Run 运行应用程序主逻辑
func (a *{Service}App) Run(ctx context.Context) error {
	a.logger.Infow("{Service} Service started successfully",
		"address", fmt.Sprintf(":%d", a.opts.Server.Port),
	)

	return a.bootstrap.Run(ctx, nil)
}

// Shutdown 优雅关闭应用程序
func (a *{Service}App) Shutdown(ctx context.Context) error {
	a.logger.Infow("Shutting down {Service} Service")
	return a.bootstrap.Shutdown(ctx)
}

// registerComponents 注册所有组件初始化器
func (a *{Service}App) registerComponents() {
	// 1. Database (优先级 300)
	a.dbInit = initializers.NewDatabaseInitializer(a.opts, a.logger)
	a.bootstrap.Register(a.dbInit)

	// 2. Health Check Server (优先级最低)
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

#### 模板 3: Simple 模式 app.go

```go
package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/internal/{service}/config"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
)

// Execute runs the {service} command
func Execute() {
	opts := config.NewOptions()

	runFunc := func(opts commonapp.Options) error {
		return run(opts.(*config.Options))
	}

	commonapp.RunWithOptions(opts, runFunc, commonapp.CommandConfig{
		Use:       "{service}",
		Short:     "{Service} Service",
		Long:      "{Service} Service provides {description}",
		EnvPrefix: "{SERVICE}",
	},
		commonapp.WithHealthCheck(commonapp.DefaultHealthCheckFuncFromOptions(opts)),
		commonapp.WithPrintVersion(),
		commonapp.WithPrintRuntime(),
		commonapp.WithWatch(),
	)
}

// run runs the {service} service
func run(opts *config.Options) error {
	// 初始化日志
	log, err := opts.InitLogger()
	if err != nil {
		return fmt.Errorf("failed to init logger: %w", err)
	}
	defer log.Flush()

	log.Info("Starting {Service} Service...")

	// 创建服务器
	srv, err := NewServer(opts, log)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	// 启动服务器
	ctx := context.Background()
	return srv.Run(ctx)
}
```

### 9.3 检查清单

#### 代码审查检查清单

- [ ] main.go 使用标准格式和完整注释
- [ ] 入口函数名称为 `Execute()`
- [ ] CommandConfig 格式正确
- [ ] 日志初始化方式统一
- [ ] 健康检查实现统一
- [ ] 组件优先级设置合理
- [ ] 错误处理包含上下文信息
- [ ] 所有公开函数有注释
- [ ] 代码格式化（gofmt）
- [ ] 通过 golangci-lint 检查

#### 测试检查清单

- [ ] 服务能正常启动
- [ ] 服务能优雅关闭
- [ ] 健康检查端点正常
- [ ] 配置文件能正确加载
- [ ] 环境变量能正确覆盖
- [ ] 日志输出格式正确
- [ ] 所有原有功能正常
- [ ] 性能无明显下降

---

## 🎓 十、总结

### 10.1 预期收益

#### 短期收益
- ✅ 代码风格统一，易于阅读
- ✅ 新开发者上手更快
- ✅ 代码审查更高效

#### 长期收益
- ✅ 维护成本降低 30%
- ✅ 新服务开发效率提升 50%
- ✅ Bug 率降低
- ✅ 架构清晰，易于扩展

### 10.2 成功指标

| 指标 | 当前 | 目标 | 测量方法 |
|------|------|------|----------|
| 代码风格一致性 | 60% | 100% | 静态分析 |
| 新服务开发时间 | 2天 | 1天 | 统计 |
| 代码审查时间 | 2小时 | 1小时 | 统计 |
| 新人上手时间 | 3天 | 1.5天 | 调研 |

### 10.3 后续行动

1. **立即行动**:
   - 开始阶段一：基础规范统一
   - 创建 feature 分支
   - 设置代码审查流程

2. **持续改进**:
   - 收集团队反馈
   - 优化开发工具
   - 更新文档

3. **知识分享**:
   - 组织技术分享会
   - 编写博客文章
   - 更新团队 Wiki

---

## 📞 联系方式

- **负责人**: [待定]
- **审阅者**: [待定]
- **相关文档**:
  - `CONTRIBUTING.md`
  - `DEVELOPMENT.md`
  - `pkg/app/README.md`
  - `pkg/bootstrap/README.md`

---

**文档状态**: ✅ Draft Ready for Review

**下一步**: 获得团队审批后开始实施阶段一

