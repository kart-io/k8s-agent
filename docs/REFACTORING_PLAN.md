# 代码重构执行计划

## 文档说明

本文档提供 k8s-agent 项目代码规范统一的详细执行计划，包括具体步骤、脚本和验证方法。

**创建时间**: 2025-10-30
**状态**: 执行中
**负责人**: 开发团队

---

## 目录

- [1. 重构概览](#1-重构概览)
- [2. 阶段 1：Reasoning 服务重构](#2-阶段-1reasoning-服务重构)
- [3. 阶段 2：Collect-Agent 服务重构](#3-阶段-2collect-agent-服务重构)
- [4. 阶段 3：Auth 服务清理](#4-阶段-3auth-服务清理)
- [5. 阶段 4：Orchestrator 数据库统一](#5-阶段-4orchestrator-数据库统一)
- [6. 进度跟踪](#6-进度跟踪)

---

## 1. 重构概览

### 1.1 总体目标

将所有服务统一到标准架构：
- ✅ **已标准化**：agent-manager, orchestrator, auth (部分)
- 🔄 **待重构**：reasoning, collect-agent
- 🧹 **待清理**：auth (移除重复代码)

### 1.2 重构优先级

| 阶段 | 服务 | 任务 | 预计工时 | 风险 | 优先级 |
|------|------|------|----------|------|--------|
| 1 | reasoning | 日志+入口+初始化器 | 3-4 天 | 中 | P0 |
| 2 | collect-agent | 日志+入口+初始化器 | 2-3 天 | 低 | P0 |
| 3 | auth | 清理重复代码 | 2-3 天 | 低 | P1 |
| 4 | orchestrator | 数据库层统一 | 2-3 天 | 中 | P1 |

**总预计工时**: 9-13 天

### 1.3 验收标准

每个阶段完成后必须满足：

1. **代码规范**
   - [ ] 使用 `commonapp.RunWithRunner()` + `Application` 接口
   - [ ] 使用 `github.com/kart-io/logger/core.Logger`
   - [ ] 有 `internal/<service>/initializers/` 包
   - [ ] 配置选项命名为 `options.ServerOptions`

2. **功能验证**
   - [ ] 所有单元测试通过
   - [ ] 服务能正常启动和关闭
   - [ ] API 端点正常工作
   - [ ] 健康检查正常

3. **代码质量**
   - [ ] `make lint` 无错误
   - [ ] 代码覆盖率不降低
   - [ ] 性能无明显下降

---

## 2. 阶段 1：Reasoning 服务重构

### 2.1 当前状态分析

**当前架构**：
```
cmd/reasoning/app/
├── app.go                    # 使用 RunWithOptions + 简单 run()
└── server.go                 # 手动创建服务器

internal/reasoning/
├── config/                   # config.Options
├── api/                      # HTTP 处理器
├── llm/                      # LLM 客户端
├── memory/                   # 向量存储
└── ...
(无 initializers/)            ❌
```

**使用的日志**：`common/logger` ❌

### 2.2 重构步骤

#### 步骤 1：创建新的配置选项结构

**目标**：创建标准的 `options.ServerOptions`

**文件**：`cmd/reasoning/app/options/options.go`

```bash
# 创建目录
mkdir -p cmd/reasoning/app/options

# 创建文件
cat > cmd/reasoning/app/options/options.go << 'EOF'
package options

import (
    "github.com/kart-io/k8s-agent/common/options"
    "github.com/kart-io/k8s-agent/internal/reasoning/config"
    "github.com/kart-io/logger/core"
)

// ServerOptions 包含 Reasoning Service 的所有配置选项
type ServerOptions struct {
    // 通用选项
    Server  options.ServerOptions  `json:"server" mapstructure:"server"`
    Health  options.HealthOptions  `json:"health" mapstructure:"health"`
    Logging options.LoggingOptions `json:"logging" mapstructure:"logging"`

    // Reasoning 特定选项
    LLM    config.LLMOptions    `json:"llm" mapstructure:"llm"`
    Memory config.MemoryOptions `json:"memory" mapstructure:"memory"`
}

// NewServerOptions 创建默认配置选项
func NewServerOptions() *ServerOptions {
    return &ServerOptions{
        Server:  *options.NewServerOptions(),
        Health:  *options.NewHealthOptions(),
        Logging: *options.NewLoggingOptions(),
        LLM:     config.DefaultLLMOptions(),
        Memory:  config.DefaultMemoryOptions(),
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
func (o *ServerOptions) Config() (*config.Config, error) {
    return &config.Config{
        Server:  o.Server,
        LLM:     o.LLM,
        Memory:  o.Memory,
        Logging: o.Logging,
    }, nil
}
EOF
```

#### 步骤 2：创建 Initializers

**目标**：创建标准的组件初始化器

```bash
# 创建 initializers 目录
mkdir -p internal/reasoning/initializers

# 1. LLM 初始化器
cat > internal/reasoning/initializers/llm.go << 'EOF'
package initializers

import (
    "context"
    "fmt"

    "github.com/kart-io/logger/core"
    "github.com/kart-io/k8s-agent/cmd/reasoning/app/options"
    "github.com/kart-io/k8s-agent/internal/reasoning/llm"
)

// LLMInitializer LLM 客户端初始化器
type LLMInitializer struct {
    opts   *options.ServerOptions
    logger core.Logger
    client *llm.Client
}

// NewLLMInitializer 创建 LLM 初始化器
func NewLLMInitializer(opts *options.ServerOptions, logger core.Logger) *LLMInitializer {
    return &LLMInitializer{
        opts:   opts,
        logger: logger.With("initializer", "llm"),
    }
}

// Name 返回初始化器名称
func (i *LLMInitializer) Name() string {
    return "llm"
}

// Priority 返回初始化优先级
func (i *LLMInitializer) Priority() int {
    return 400
}

// Initialize 执行初始化
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
    i.logger.Infow("LLM client initialized successfully",
        "provider", i.opts.LLM.Provider,
        "model", i.opts.LLM.Model,
    )
    return nil
}

// Shutdown 执行关闭
func (i *LLMInitializer) Shutdown(ctx context.Context) error {
    i.logger.Infow("Shutting down LLM client")
    if i.client != nil {
        return i.client.Close()
    }
    return nil
}

// GetClient 获取客户端实例
func (i *LLMInitializer) GetClient() *llm.Client {
    return i.client
}
EOF

# 2. Memory 初始化器
cat > internal/reasoning/initializers/memory.go << 'EOF'
package initializers

import (
    "context"
    "fmt"

    "github.com/kart-io/logger/core"
    "github.com/kart-io/k8s-agent/cmd/reasoning/app/options"
    "github.com/kart-io/k8s-agent/internal/reasoning/memory"
)

// MemoryInitializer 向量存储初始化器
type MemoryInitializer struct {
    opts   *options.ServerOptions
    logger core.Logger
    store  *memory.VectorStore
}

// NewMemoryInitializer 创建 Memory 初始化器
func NewMemoryInitializer(opts *options.ServerOptions, logger core.Logger) *MemoryInitializer {
    return &MemoryInitializer{
        opts:   opts,
        logger: logger.With("initializer", "memory"),
    }
}

// Name 返回初始化器名称
func (i *MemoryInitializer) Name() string {
    return "memory"
}

// Priority 返回初始化优先级
func (i *MemoryInitializer) Priority() int {
    return 450
}

// Initialize 执行初始化
func (i *MemoryInitializer) Initialize(ctx context.Context) error {
    i.logger.Infow("Initializing vector store")

    if !i.opts.Memory.EnableVectorStore {
        i.logger.Infow("Vector store disabled, skipping initialization")
        return nil
    }

    store, err := memory.NewVectorStore(i.opts.Memory, i.logger)
    if err != nil {
        return fmt.Errorf("failed to create vector store: %w", err)
    }

    i.store = store
    i.logger.Infow("Vector store initialized successfully")
    return nil
}

// Shutdown 执行关闭
func (i *MemoryInitializer) Shutdown(ctx context.Context) error {
    i.logger.Infow("Shutting down vector store")
    if i.store != nil {
        return i.store.Close()
    }
    return nil
}

// GetStore 获取存储实例
func (i *MemoryInitializer) GetStore() *memory.VectorStore {
    return i.store
}
EOF

# 3. HTTP Server 初始化器
cat > internal/reasoning/initializers/server.go << 'EOF'
package initializers

import (
    "context"
    "fmt"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/kart-io/logger/core"
    "github.com/kart-io/k8s-agent/cmd/reasoning/app/options"
    "github.com/kart-io/k8s-agent/internal/reasoning/api"
)

// HTTPServerInitializer HTTP 服务器初始化器
type HTTPServerInitializer struct {
    opts   *options.ServerOptions
    logger core.Logger
    server *http.Server

    // 依赖的初始化器
    llmInit    *LLMInitializer
    memoryInit *MemoryInitializer
}

// NewHTTPServerInitializer 创建 HTTP 服务器初始化器
func NewHTTPServerInitializer(
    opts *options.ServerOptions,
    logger core.Logger,
    llmInit *LLMInitializer,
    memoryInit *MemoryInitializer,
) *HTTPServerInitializer {
    return &HTTPServerInitializer{
        opts:       opts,
        logger:     logger.With("initializer", "http"),
        llmInit:    llmInit,
        memoryInit: memoryInit,
    }
}

// Name 返回初始化器名称
func (i *HTTPServerInitializer) Name() string {
    return "http-server"
}

// Priority 返回初始化优先级
func (i *HTTPServerInitializer) Priority() int {
    return 600
}

// Initialize 执行初始化
func (i *HTTPServerInitializer) Initialize(ctx context.Context) error {
    i.logger.Infow("Initializing HTTP server")

    // 创建 Gin 路由
    router := gin.New()
    router.Use(gin.Recovery())

    // 设置 API 路由
    api.SetupRoutes(router, i.llmInit.GetClient(), i.memoryInit.GetStore(), i.logger)

    // 创建 HTTP 服务器
    addr := fmt.Sprintf("%s:%d", i.opts.Server.Host, i.opts.Server.Port)
    i.server = &http.Server{
        Addr:           addr,
        Handler:        router,
        ReadTimeout:    30 * time.Second,
        WriteTimeout:   30 * time.Second,
        MaxHeaderBytes: 1 << 20,
    }

    // 在 goroutine 中启动服务器
    go func() {
        i.logger.Infow("HTTP server listening", "address", addr)
        if err := i.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            i.logger.Errorw("HTTP server error", "error", err)
        }
    }()

    return nil
}

// Shutdown 执行关闭
func (i *HTTPServerInitializer) Shutdown(ctx context.Context) error {
    i.logger.Infow("Shutting down HTTP server")
    if i.server != nil {
        return i.server.Shutdown(ctx)
    }
    return nil
}
EOF
```

#### 步骤 3：更新 app.go

**目标**：使用 Application 接口

```bash
# 备份旧文件
cp cmd/reasoning/app/app.go cmd/reasoning/app/app.go.bak

# 创建新的 app.go
cat > cmd/reasoning/app/app.go << 'EOF'
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

// Execute 运行 reasoning 命令
func Execute() {
    // 创建配置选项
    opts := options.NewServerOptions()

    // 使用组合框架运行应用
    commonapp.RunWithRunner(
        opts,
        &ReasoningApp{},
        initLogger,
        commonapp.CommandConfig{
            Use:       "reasoning",
            Short:     "Reasoning Service",
            Long:      "Reasoning Service provides AI-driven root cause analysis and intelligent recommendations",
            EnvPrefix: "REASONING",
        },
    )
}

// ReasoningApp 实现 commonapp.Application 接口
type ReasoningApp struct {
    bootstrap *bootstrap.Bootstrap
    opts      *options.ServerOptions
    logger    core.Logger

    // 组件初始化器
    llmInit    *initializers.LLMInitializer
    memoryInit *initializers.MemoryInitializer
    httpInit   *initializers.HTTPServerInitializer
    healthInit *pkginitializers.HealthCheckInitializer
}

// Initialize 初始化应用程序
func (a *ReasoningApp) Initialize(ctx context.Context, opts commonapp.Options) error {
    a.opts = opts.(*options.ServerOptions)

    // 初始化日志系统
    logger, err := initLogger(opts)
    if err != nil {
        return fmt.Errorf("failed to initialize logger: %w", err)
    }
    a.logger = logger

    a.logger.Infow("Initializing Reasoning Service",
        "http_port", a.opts.Server.Port,
        "health_port", a.opts.GetHealthPort(),
        "llm_enabled", a.opts.LLM.Enabled,
        "memory_enabled", a.opts.Memory.EnableVectorStore,
    )

    // 创建 bootstrap 实例
    a.bootstrap = bootstrap.New(a.logger)

    // 注册所有组件初始化器
    a.registerComponents()

    a.logger.Infow("Components registered, ready to start")
    return nil
}

// Run 运行应用程序主逻辑
func (a *ReasoningApp) Run(ctx context.Context) error {
    a.logger.Infow("Reasoning Service started successfully",
        "http_address", fmt.Sprintf("%s:%d", a.opts.Server.Host, a.opts.Server.Port),
        "health_address", fmt.Sprintf(":%d", a.opts.GetHealthPort()),
    )

    // 使用 bootstrap 的 Run 方法，它会等待信号
    return a.bootstrap.Run(ctx, nil)
}

// Shutdown 优雅关闭应用程序
func (a *ReasoningApp) Shutdown(ctx context.Context) error {
    a.logger.Infow("Shutting down Reasoning Service")
    return a.bootstrap.Shutdown(ctx)
}

// registerComponents 注册所有组件初始化器
func (a *ReasoningApp) registerComponents() {
    // 1. LLM Client (优先级 400)
    a.llmInit = initializers.NewLLMInitializer(a.opts, a.logger)
    a.bootstrap.Register(a.llmInit)

    // 2. Memory/Vector Store (优先级 450)
    a.memoryInit = initializers.NewMemoryInitializer(a.opts, a.logger)
    a.bootstrap.Register(a.memoryInit)

    // 3. HTTP Server (优先级 600)
    a.httpInit = initializers.NewHTTPServerInitializer(
        a.opts,
        a.logger,
        a.llmInit,
        a.memoryInit,
    )
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
EOF
```

#### 步骤 4：更新日志导入

**目标**：全部迁移到 `kart-io/logger`

```bash
# 查找所有使用旧 logger 的文件
find internal/reasoning -name "*.go" -type f -exec grep -l "common/logger" {} \;

# 批量替换导入语句
find internal/reasoning -name "*.go" -type f -exec sed -i \
  's|"github.com/kart-io/k8s-agent/common/logger"|"github.com/kart-io/logger/core"|g' {} \;

# 替换类型声明（如果有）
find internal/reasoning -name "*.go" -type f -exec sed -i \
  's|logger\.Logger|core.Logger|g' {} \;

# 替换初始化函数调用
find internal/reasoning -name "*.go" -type f -exec sed -i \
  's|logger\.InitFromOptions|core.InitLogger|g' {} \;
```

#### 步骤 5：验证和测试

```bash
# 1. 编译检查
cd /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent
go build -o _output/bin/reasoning ./cmd/reasoning

# 2. 运行单元测试
go test ./internal/reasoning/... -v

# 3. 运行服务（测试启动）
./_output/bin/reasoning --config configs/reasoning.yaml

# 4. 检查健康端点
curl http://localhost:8082/health
curl http://localhost:9082/health  # Health check 端点

# 5. 测试 API 端点
curl http://localhost:8082/api/v1/analyze/root-cause -X POST \
  -H "Content-Type: application/json" \
  -d '{"event": {...}}'
```

### 2.3 回滚方案

如果重构出现问题：

```bash
# 恢复备份文件
cp cmd/reasoning/app/app.go.bak cmd/reasoning/app/app.go

# 删除新创建的文件
rm -rf cmd/reasoning/app/options
rm -rf internal/reasoning/initializers

# 重新构建
make build-reasoning
```

### 2.4 完成检查清单

- [ ] 创建 `cmd/reasoning/app/options/options.go`
- [ ] 创建 `internal/reasoning/initializers/` 包
- [ ] 更新 `cmd/reasoning/app/app.go` 使用 Application 接口
- [ ] 所有文件迁移到 `kart-io/logger`
- [ ] 编译成功无警告
- [ ] 单元测试全部通过
- [ ] 服务能正常启动和关闭
- [ ] 健康检查端点正常
- [ ] API 端点正常工作
- [ ] 日志输出格式正确

---

## 3. 阶段 2：Collect-Agent 服务重构

### 3.1 当前状态分析

**当前架构**：
```
cmd/collect-agent/app/
├── app.go                    # 使用 RunWithOptions
└── server.go

internal/collect-agent/
├── config/
├── agent/                    # Agent 核心逻辑
├── event/                    # 事件监控
└── ...
(无 initializers/)            ❌
```

**使用的日志**：`common/logger` ❌

### 3.2 重构步骤

#### 步骤 1：创建配置选项

```bash
mkdir -p cmd/collect-agent/app/options

cat > cmd/collect-agent/app/options/options.go << 'EOF'
package options

import (
    "github.com/kart-io/k8s-agent/common/options"
    "github.com/kart-io/k8s-agent/internal/collect-agent/config"
    "github.com/kart-io/logger/core"
)

// ServerOptions 包含 Collect-Agent 的所有配置选项
type ServerOptions struct {
    // 通用选项
    Health  options.HealthOptions  `json:"health" mapstructure:"health"`
    Logging options.LoggingOptions `json:"logging" mapstructure:"logging"`

    // Agent 特定选项
    Agent config.AgentOptions `json:"agent" mapstructure:"agent"`
    NATS  options.NATSOptions `json:"nats" mapstructure:"nats"`
}

// NewServerOptions 创建默认配置选项
func NewServerOptions() *ServerOptions {
    return &ServerOptions{
        Health:  *options.NewHealthOptions(),
        Logging: *options.NewLoggingOptions(),
        Agent:   config.DefaultAgentOptions(),
        NATS:    *options.NewNATSOptions(),
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
    return o.Agent.HealthPort
}

// Config 转换为业务配置
func (o *ServerOptions) Config() (*config.Config, error) {
    return &config.Config{
        Agent:   o.Agent,
        NATS:    o.NATS,
        Logging: o.Logging,
    }, nil
}
EOF
```

#### 步骤 2：创建 Initializers

```bash
mkdir -p internal/collect-agent/initializers

# 1. NATS 初始化器
cat > internal/collect-agent/initializers/nats.go << 'EOF'
package initializers

import (
    "context"
    "fmt"

    "github.com/kart-io/logger/core"
    "github.com/kart-io/k8s-agent/cmd/collect-agent/app/options"
    "github.com/kart-io/k8s-agent/common/mq"
    "github.com/nats-io/nats.go"
)

// NATSInitializer NATS 连接初始化器
type NATSInitializer struct {
    opts   *options.ServerOptions
    logger core.Logger
    conn   *nats.Conn
}

// NewNATSInitializer 创建 NATS 初始化器
func NewNATSInitializer(opts *options.ServerOptions, logger core.Logger) *NATSInitializer {
    return &NATSInitializer{
        opts:   opts,
        logger: logger.With("initializer", "nats"),
    }
}

// Name 返回初始化器名称
func (i *NATSInitializer) Name() string {
    return "nats"
}

// Priority 返回初始化优先级
func (i *NATSInitializer) Priority() int {
    return 300
}

// Initialize 执行初始化
func (i *NATSInitializer) Initialize(ctx context.Context) error {
    i.logger.Infow("Initializing NATS connection")

    conn, err := mq.NewNATSConnection(i.opts.NATS)
    if err != nil {
        return fmt.Errorf("failed to create NATS connection: %w", err)
    }

    i.conn = conn
    i.logger.Infow("NATS connection established",
        "url", i.opts.NATS.URL,
    )
    return nil
}

// Shutdown 执行关闭
func (i *NATSInitializer) Shutdown(ctx context.Context) error {
    i.logger.Infow("Closing NATS connection")
    if i.conn != nil {
        i.conn.Close()
    }
    return nil
}

// GetConnection 获取 NATS 连接
func (i *NATSInitializer) GetConnection() *nats.Conn {
    return i.conn
}
EOF

# 2. K8s Watcher 初始化器
cat > internal/collect-agent/initializers/watcher.go << 'EOF'
package initializers

import (
    "context"
    "fmt"

    "github.com/kart-io/logger/core"
    "github.com/kart-io/k8s-agent/cmd/collect-agent/app/options"
    "github.com/kart-io/k8s-agent/internal/collect-agent/event"
)

// WatcherInitializer Kubernetes 事件监控初始化器
type WatcherInitializer struct {
    opts     *options.ServerOptions
    logger   core.Logger
    natsInit *NATSInitializer
    watcher  *event.Watcher
}

// NewWatcherInitializer 创建 Watcher 初始化器
func NewWatcherInitializer(
    opts *options.ServerOptions,
    logger core.Logger,
    natsInit *NATSInitializer,
) *WatcherInitializer {
    return &WatcherInitializer{
        opts:     opts,
        logger:   logger.With("initializer", "watcher"),
        natsInit: natsInit,
    }
}

// Name 返回初始化器名称
func (i *WatcherInitializer) Name() string {
    return "k8s-watcher"
}

// Priority 返回初始化优先级
func (i *WatcherInitializer) Priority() int {
    return 500
}

// Initialize 执行初始化
func (i *WatcherInitializer) Initialize(ctx context.Context) error {
    i.logger.Infow("Initializing Kubernetes event watcher")

    watcher, err := event.NewWatcher(
        i.opts.Agent,
        i.natsInit.GetConnection(),
        i.logger,
    )
    if err != nil {
        return fmt.Errorf("failed to create watcher: %w", err)
    }

    // 启动监控
    if err := watcher.Start(ctx); err != nil {
        return fmt.Errorf("failed to start watcher: %w", err)
    }

    i.watcher = watcher
    i.logger.Infow("Kubernetes event watcher started")
    return nil
}

// Shutdown 执行关闭
func (i *WatcherInitializer) Shutdown(ctx context.Context) error {
    i.logger.Infow("Stopping Kubernetes event watcher")
    if i.watcher != nil {
        return i.watcher.Stop()
    }
    return nil
}
EOF
```

#### 步骤 3：更新 app.go

```bash
cp cmd/collect-agent/app/app.go cmd/collect-agent/app/app.go.bak

cat > cmd/collect-agent/app/app.go << 'EOF'
package app

import (
    "context"
    "fmt"

    "github.com/kart-io/k8s-agent/cmd/collect-agent/app/options"
    "github.com/kart-io/k8s-agent/internal/collect-agent/initializers"
    commonapp "github.com/kart-io/k8s-agent/pkg/app"
    "github.com/kart-io/k8s-agent/pkg/bootstrap"
    pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
    "github.com/kart-io/logger/core"
)

// Execute 运行 collect-agent 命令
func Execute() {
    opts := options.NewServerOptions()

    commonapp.RunWithRunner(
        opts,
        &CollectAgentApp{},
        initLogger,
        commonapp.CommandConfig{
            Use:       "collect-agent",
            Short:     "Collect Agent",
            Long:      "Collect Agent monitors K8s cluster events and collects metrics from edge clusters",
            EnvPrefix: "COLLECT_AGENT",
        },
    )
}

// CollectAgentApp 实现 commonapp.Application 接口
type CollectAgentApp struct {
    bootstrap *bootstrap.Bootstrap
    opts      *options.ServerOptions
    logger    core.Logger

    // 组件初始化器
    natsInit    *initializers.NATSInitializer
    watcherInit *initializers.WatcherInitializer
    healthInit  *pkginitializers.HealthCheckInitializer
}

// Initialize 初始化应用程序
func (a *CollectAgentApp) Initialize(ctx context.Context, opts commonapp.Options) error {
    a.opts = opts.(*options.ServerOptions)

    // 初始化日志系统
    logger, err := initLogger(opts)
    if err != nil {
        return fmt.Errorf("failed to initialize logger: %w", err)
    }
    a.logger = logger

    a.logger.Infow("Initializing Collect Agent",
        "cluster_id", a.opts.Agent.ClusterID,
        "central_endpoint", a.opts.Agent.CentralEndpoint,
        "health_port", a.opts.GetHealthPort(),
    )

    // 创建 bootstrap 实例
    a.bootstrap = bootstrap.New(a.logger)

    // 注册所有组件初始化器
    a.registerComponents()

    a.logger.Infow("Components registered, ready to start")
    return nil
}

// Run 运行应用程序主逻辑
func (a *CollectAgentApp) Run(ctx context.Context) error {
    a.logger.Infow("Collect Agent started successfully",
        "cluster_id", a.opts.Agent.ClusterID,
        "health_address", fmt.Sprintf(":%d", a.opts.GetHealthPort()),
    )

    // 使用 bootstrap 的 Run 方法
    return a.bootstrap.Run(ctx, nil)
}

// Shutdown 优雅关闭应用程序
func (a *CollectAgentApp) Shutdown(ctx context.Context) error {
    a.logger.Infow("Shutting down Collect Agent")
    return a.bootstrap.Shutdown(ctx)
}

// registerComponents 注册所有组件初始化器
func (a *CollectAgentApp) registerComponents() {
    // 1. NATS Connection (优先级 300)
    a.natsInit = initializers.NewNATSInitializer(a.opts, a.logger)
    a.bootstrap.Register(a.natsInit)

    // 2. Kubernetes Watcher (优先级 500)
    a.watcherInit = initializers.NewWatcherInitializer(
        a.opts,
        a.logger,
        a.natsInit,
    )
    a.bootstrap.Register(a.watcherInit)

    // 3. Health Check Server (优先级最低)
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
EOF
```

#### 步骤 4：更新日志导入

```bash
# 批量替换
find internal/collect-agent -name "*.go" -type f -exec sed -i \
  's|"github.com/kart-io/k8s-agent/common/logger"|"github.com/kart-io/logger/core"|g' {} \;

find internal/collect-agent -name "*.go" -type f -exec sed -i \
  's|commonlogger\.|core.|g' {} \;
```

#### 步骤 5：验证和测试

```bash
# 编译
go build -o _output/bin/collect-agent ./cmd/collect-agent

# 测试（需要 K8s 集群）
./_output/bin/collect-agent \
  --cluster-id test-cluster \
  --central-endpoint nats://localhost:4222

# 健康检查
curl http://localhost:8090/health
```

### 3.3 完成检查清单

- [ ] 创建 `cmd/collect-agent/app/options/options.go`
- [ ] 创建 `internal/collect-agent/initializers/` 包
- [ ] 更新 `cmd/collect-agent/app/app.go`
- [ ] 迁移到 `kart-io/logger`
- [ ] 编译成功
- [ ] 服务能启动和关闭
- [ ] 能正常监控 K8s 事件

---

## 4. 阶段 3：Auth 服务清理

### 4.1 清理目标

移除以下重复包（使用 common/ 中的版本）：
- ❌ `internal/auth/logger/`
- ❌ `internal/auth/middleware/`
- ❌ `internal/auth/response/`
- ❌ `internal/auth/pagination/`
- ❌ `internal/auth/metrics/` (移到 pkg/)
- ❌ `internal/auth/cache/` (使用 common/cache)

### 4.2 清理步骤

#### 步骤 1：备份和分析

```bash
# 备份 auth 服务
cp -r internal/auth internal/auth.backup

# 分析依赖关系
grep -r "internal/auth/logger" internal/auth | wc -l
grep -r "internal/auth/middleware" internal/auth | wc -l
grep -r "internal/auth/response" internal/auth | wc -l
grep -r "internal/auth/pagination" internal/auth | wc -l
```

#### 步骤 2：批量替换导入

```bash
# 替换 logger
find internal/auth -name "*.go" -type f -exec sed -i \
  's|"github.com/kart-io/k8s-agent/internal/auth/logger"|"github.com/kart-io/logger/core"|g' {} \;

# 替换 middleware
find internal/auth -name "*.go" -type f -exec sed -i \
  's|"github.com/kart-io/k8s-agent/internal/auth/middleware"|"github.com/kart-io/k8s-agent/common/middleware"|g' {} \;

# 替换 response
find internal/auth -name "*.go" -type f -exec sed -i \
  's|"github.com/kart-io/k8s-agent/internal/auth/response"|"github.com/kart-io/k8s-agent/common/response"|g' {} \;

# 替换 pagination
find internal/auth -name "*.go" -type f -exec sed -i \
  's|"github.com/kart-io/k8s-agent/internal/auth/pagination"|"github.com/kart-io/k8s-agent/common/pagination"|g' {} \;

# 替换 cache
find internal/auth -name "*.go" -type f -exec sed -i \
  's|"github.com/kart-io/k8s-agent/internal/auth/cache"|"github.com/kart-io/k8s-agent/common/cache"|g' {} \;
```

#### 步骤 3：移动 metrics 到 pkg/

```bash
# 移动 metrics
mkdir -p pkg/metrics
mv internal/auth/metrics/* pkg/metrics/

# 更新导入
find internal/auth -name "*.go" -type f -exec sed -i \
  's|"github.com/kart-io/k8s-agent/internal/auth/metrics"|"github.com/kart-io/k8s-agent/pkg/metrics"|g' {} \;
```

#### 步骤 4：删除重复包

```bash
# 确认编译通过后再删除
go build ./cmd/auth

# 删除重复的包
rm -rf internal/auth/logger
rm -rf internal/auth/middleware
rm -rf internal/auth/response
rm -rf internal/auth/pagination
rm -rf internal/auth/cache
rm -rf internal/auth/metrics
```

#### 步骤 5：验证

```bash
# 编译
go build -o _output/bin/auth ./cmd/auth

# 运行测试
go test ./internal/auth/... -v

# 启动服务
./_output/bin/auth --config configs/auth.yaml

# 测试 API
curl http://localhost:8083/api/v1/auth/login -X POST \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'
```

### 4.3 完成检查清单

- [ ] 所有导入已替换到 common 包
- [ ] metrics 已移动到 pkg/
- [ ] 编译成功无错误
- [ ] 所有测试通过
- [ ] API 功能正常
- [ ] 重复包已删除

---

## 5. 阶段 4：Orchestrator 数据库统一

### 5.1 迁移目标

将 `internal/orchestrator/storage/postgres.go` 从直接使用 GORM 迁移到 `common/db.MySQLClient`

### 5.2 迁移步骤

#### 步骤 1：备份

```bash
cp internal/orchestrator/storage/postgres.go \
   internal/orchestrator/storage/postgres.go.bak
```

#### 步骤 2：重写 storage

```bash
cat > internal/orchestrator/storage/postgres.go << 'EOF'
package storage

import (
    "context"

    "github.com/kart-io/logger/core"
    "github.com/kart-io/k8s-agent/common/db"
    "github.com/kart-io/k8s-agent/internal/orchestrator/types"
)

// PostgresStore implements storage using MySQL
type PostgresStore struct {
    *db.MySQLClient  // 使用封装的客户端
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
        &types.Workflow{},
        &types.WorkflowExecution{},
        &types.Strategy{},
    ); err != nil {
        return nil, err
    }

    store.logger.Infow("PostgreSQL store initialized",
        "host", config.Host,
        "database", config.Database)

    return store, nil
}

// Workflow 操作方法...
func (s *PostgresStore) SaveWorkflow(ctx context.Context, workflow *types.Workflow) error {
    return s.DB.WithContext(ctx).Save(workflow).Error
}

func (s *PostgresStore) GetWorkflow(ctx context.Context, id string) (*types.Workflow, error) {
    var workflow types.Workflow
    if err := s.DB.WithContext(ctx).First(&workflow, "id = ?", id).Error; err != nil {
        return nil, err
    }
    return &workflow, nil
}

// ... 其他方法保持不变
EOF
```

#### 步骤 3：更新 initializers

```bash
# 更新 internal/orchestrator/initializers/database.go
# 确保使用新的 PostgresStore
```

#### 步骤 4：验证

```bash
# 编译
go build -o _output/bin/orchestrator ./cmd/orchestrator

# 测试
go test ./internal/orchestrator/storage/... -v

# 运行
./_output/bin/orchestrator --config configs/orchestrator.yaml
```

### 5.3 完成检查清单

- [ ] storage 使用 `common/db.MySQLClient`
- [ ] 编译成功
- [ ] 测试通过
- [ ] 数据库操作正常
- [ ] 连接池配置生效

---

## 6. 进度跟踪

### 6.1 整体进度

| 阶段 | 任务 | 状态 | 开始日期 | 完成日期 | 负责人 |
|------|------|------|----------|----------|--------|
| 1 | Reasoning 服务重构 | ⏸️ 待开始 | - | - | - |
| 2 | Collect-Agent 服务重构 | ⏸️ 待开始 | - | - | - |
| 3 | Auth 服务清理 | ⏸️ 待开始 | - | - | - |
| 4 | Orchestrator 数据库统一 | ⏸️ 待开始 | - | - | - |

### 6.2 详细任务清单

#### Reasoning 服务

- [ ] 创建 options 包
- [ ] 创建 initializers 包
  - [ ] LLM 初始化器
  - [ ] Memory 初始化器
  - [ ] HTTP Server 初始化器
- [ ] 更新 app.go
- [ ] 迁移日志系统
- [ ] 编译验证
- [ ] 测试验证
- [ ] 功能验证

#### Collect-Agent 服务

- [ ] 创建 options 包
- [ ] 创建 initializers 包
  - [ ] NATS 初始化器
  - [ ] Watcher 初始化器
- [ ] 更新 app.go
- [ ] 迁移日志系统
- [ ] 编译验证
- [ ] 测试验证
- [ ] 功能验证

#### Auth 服务

- [ ] 分析依赖关系
- [ ] 替换 logger 导入
- [ ] 替换 middleware 导入
- [ ] 替换 response 导入
- [ ] 替换 pagination 导入
- [ ] 移动 metrics 到 pkg/
- [ ] 删除重复包
- [ ] 编译验证
- [ ] 测试验证
- [ ] 功能验证

#### Orchestrator 服务

- [ ] 备份 storage
- [ ] 重写使用 MySQLClient
- [ ] 更新 initializers
- [ ] 编译验证
- [ ] 测试验证
- [ ] 功能验证

### 6.3 风险和问题跟踪

| 日期 | 阶段 | 问题描述 | 严重程度 | 状态 | 解决方案 |
|------|------|----------|----------|------|----------|
| - | - | - | - | - | - |

---

## 附录

### A. 快速参考命令

```bash
# 编译所有服务
make build

# 编译特定服务
make build-reasoning
make build-collect-agent
make build-auth
make build-orchestrator

# 运行测试
make test

# 运行特定服务测试
go test ./internal/reasoning/... -v
go test ./internal/collect-agent/... -v

# 代码检查
make lint

# 启动所有服务（Docker Compose）
make docker-compose-up

# 查看服务日志
docker-compose logs -f reasoning
docker-compose logs -f collect-agent
```

### B. 有用的脚本

#### 查找所有使用旧 logger 的文件

```bash
#!/bin/bash
# find-old-logger.sh

echo "=== Files using old logger ==="
find . -name "*.go" -type f -exec grep -l "common/logger" {} \; | \
  grep -v vendor | grep -v ".backup"

echo ""
echo "=== Count by directory ==="
find . -name "*.go" -type f -exec grep -l "common/logger" {} \; | \
  grep -v vendor | grep -v ".backup" | \
  xargs dirname | sort | uniq -c
```

#### 批量替换导入

```bash
#!/bin/bash
# replace-imports.sh

SERVICE=$1

if [ -z "$SERVICE" ]; then
    echo "Usage: $0 <service-name>"
    echo "Example: $0 reasoning"
    exit 1
fi

echo "Replacing imports in internal/$SERVICE..."

# 替换 logger
find internal/$SERVICE -name "*.go" -type f -exec sed -i \
  's|"github.com/kart-io/k8s-agent/common/logger"|"github.com/kart-io/logger/core"|g' {} \;

echo "Done! Please run: go build ./cmd/$SERVICE"
```

#### 验证服务

```bash
#!/bin/bash
# verify-service.sh

SERVICE=$1

if [ -z "$SERVICE" ]; then
    echo "Usage: $0 <service-name>"
    exit 1
fi

echo "=== Verifying $SERVICE service ==="

# 1. 编译检查
echo "1. Compiling..."
if go build -o _output/bin/$SERVICE ./cmd/$SERVICE 2>&1; then
    echo "✓ Compilation successful"
else
    echo "✗ Compilation failed"
    exit 1
fi

# 2. 测试检查
echo "2. Running tests..."
if go test ./internal/$SERVICE/... -v 2>&1; then
    echo "✓ Tests passed"
else
    echo "✗ Tests failed"
    exit 1
fi

# 3. Lint 检查
echo "3. Running linter..."
if golangci-lint run ./internal/$SERVICE/... 2>&1; then
    echo "✓ Lint passed"
else
    echo "⚠ Lint has warnings"
fi

echo ""
echo "=== Verification complete for $SERVICE ==="
```

---

**文档版本**: v1.0
**最后更新**: 2025-10-30
