# Agent 代理模式 - 多集群监控架构

## 文档信息

- **版本**: v1.6
- **最后更新**: 2025年9月27日
- **状态**: 正式版
- **所属系统**: Aetherius AI Agent
- **文档类型**: Agent 代理架构设计

> **文档关联**:
> - **多集群管理**: [ai_agent.md#5.5](../ai_agent.md#55-多集群管理策略-multi-cluster-management-strategy) - 业务需求和策略
> - **单集群模式**: [08_in_cluster_deployment.md](./08_in_cluster_deployment.md) - In-Cluster 部署对比
> - **微服务架构**: [06_microservices.md](./06_microservices.md) - 服务拆分设计

## 目录

- [1. Agent 代理模式概述](#1-agent-代理模式概述)
- [2. 架构设计](#2-架构设计)
- [3. 核心实现](#3-核心实现)
- [4. 通信协议](#4-通信协议)
- [5. 部署配置](#5-部署配置)
- [6. 运维管理](#6-运维管理)

## 1. Agent 代理模式概述

### 1.1 什么是 Agent 代理模式

**Agent 代理模式**是一种分布式多集群监控架构,通过在每个 Kubernetes 集群内部署轻量级 Agent,将事件和监控数据上报到中央控制平面,实现统一的多集群管理。

**与单集群In-Cluster模式的区别**:
- **单集群模式**: 全功能部署在一个集群内(参考[08_in_cluster_deployment.md](./08_in_cluster_deployment.md))
- **Agent代理模式**: 中央控制平面+多个轻量级Agent,适用于多集群场景

### 1.2 架构优势

| 优势 | 说明 |
|------|------|
| **统一管理** | 中央控制平面统一处理所有集群的事件和诊断 |
| **轻量化** | Agent 只负责数据收集和上报,逻辑简单 |
| **高可用** | Agent 与中心解耦,中心故障不影响 Agent 运行 |
| **安全隔离** | 每个集群的凭证独立管理,不互相访问 |
| **易扩展** | 新增集群只需部署 Agent,无需修改中心 |
| **网络友好** | Agent 主动上报,无需中心访问集群 API |

### 1.3 与其他模式对比

| 模式 | 部署复杂度 | 网络要求 | 适用场景 | 推荐度 |
|------|-----------|---------|---------|--------|
| **单集群 In-Cluster** | ⭐ | 无 | 1个集群 | ⭐⭐⭐ |
| **Out-of-Cluster 中心** | ⭐⭐ | 中心→集群 | 集中式kubeconfig管理 | ⭐⭐ |
| **Agent 代理模式** | ⭐⭐ | Agent→中心 | 2-10个集群 (多云/混合云) | ⭐⭐⭐⭐⭐ |
| **分层Agent模式** | ⭐⭐⭐ | 分层 | 10+个集群(超大规模) | ⭐⭐⭐⭐ |

## 2. 架构设计

### 2.1 整体架构

```
┌─────────────────────────────────────────────────────────────────────┐
│                    Central Control Plane                            │
│                    (中央控制平面 - 可部署在任意位置)                  │
│                                                                     │
│  ┌───────────────────────────────────────────────────────────────┐  │
│  │                   Core Services                               │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐   │  │
│  │  │ Agent        │  │ Orchestrator │  │  Reasoning       │   │  │
│  │  │ Manager      │  │   Service    │  │  Service         │   │  │
│  │  │ (Agent管理+  │  │              │  │                  │   │  │
│  │  │  事件聚合)    │  │              │  │                  │   │  │
│  │  └──────────────┘  └──────────────┘  └──────────────────┘   │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐   │  │
│  │  │ Knowledge    │  │  Execution   │  │  Report          │   │  │
│  │  │ Service      │  │  Service     │  │  Service         │   │  │
│  │  │              │  │              │  │                  │   │  │
│  │  └──────────────┘  └──────────────┘  └──────────────────┘   │  │
│  └───────────────────────────────────────────────────────────────┘  │
│                                                                     │
│  ┌───────────────────────────────────────────────────────────────┐  │
│  │                   Data Layer                                  │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────────┐ │  │
│  │  │PostgreSQL│  │  Redis   │  │ Weaviate │  │   NATS       │ │  │
│  │  │(任务状态) │  │(缓存队列) │  │(知识库)  │  │(消息总线)    │ │  │
│  │  └──────────┘  └──────────┘  └──────────┘  └──────────────┘ │  │
│  └───────────────────────────────────────────────────────────────┘  │
│                              ▲                                      │
│                              │ NATS / gRPC / WebSocket              │
└──────────────────────────────┼──────────────────────────────────────┘
                               │
                               │ Agent → Central (单向连接)
                               │
        ┌──────────────────────┼──────────────────────┐
        │                      │                      │
        │                      │                      │
        ▼                      ▼                      ▼
┌───────────────────┐  ┌───────────────────┐  ┌───────────────────┐
│   Cluster A       │  │   Cluster B       │  │   Cluster C       │
│   (prod-us-west)  │  │   (prod-eu-1)     │  │   (prod-asia)     │
│                   │  │                   │  │                   │
│  ┌─────────────┐  │  │  ┌─────────────┐  │  │  ┌─────────────┐  │
│  │   Agent     │  │  │  │   Agent     │  │  │  │   Agent     │  │
│  │  (轻量级)    │  │  │  │  (轻量级)    │  │  │  │  (轻量级)    │  │
│  │             │  │  │  │             │  │  │  │             │  │
│  │  - Watch    │  │  │  │  - Watch    │  │  │  │  - Watch    │  │
│  │    Events   │  │  │  │    Events   │  │  │  │    Events   │  │
│  │  - Collect  │  │  │  │  - Collect  │  │  │  │  - Collect  │  │
│  │    Metrics  │  │  │  │    Metrics  │  │  │  │    Metrics  │  │
│  │  - Execute  │  │  │  │  - Execute  │  │  │  │  - Execute  │  │
│  │    Commands │  │  │  │    Commands │  │  │  │    Commands │  │
│  │  - Report   │  │  │  │  - Report   │  │  │  │  - Report   │  │
│  │    Status   │  │  │  │    Status   │  │  │  │    Status   │  │
│  └──────┬──────┘  │  │  └──────┬──────┘  │  │  └──────┬──────┘  │
│         │         │  │         │         │  │         │         │
│    ┌────▼──────┐  │  │    ┌────▼──────┐  │  │    ┌────▼──────┐  │
│    │    K8s    │  │  │    │    K8s    │  │  │    │    K8s    │  │
│    │    API    │  │  │    │    API    │  │  │    │    API    │  │
│    └───────────┘  │  │    └───────────┘  │  │    └───────────┘  │
└───────────────────┘  └───────────────────┘  └───────────────────┘
```

### 2.2 Agent 核心职责

以下是Agent需要实现的核心功能模块:

```go
type AgentResponsibilities struct {
    // 1. 事件监听
    EventWatching struct {
        WatchK8sEvents    bool
        FilterEvents      bool
        DeduplicateEvents bool
        ReportEvents      bool
    }

    // 2. 指标收集
    MetricsCollection struct {
        CollectClusterMetrics bool
        CollectNodeMetrics    bool
        CollectPodMetrics     bool
        ReportMetrics         bool
    }

    // 3. 命令执行
    CommandExecution struct {
        ReceiveCommands   bool
        ValidateCommands  bool
        ExecuteCommands   bool
        ReportResults     bool
    }

    // 4. 健康上报
    HealthReporting struct {
        SelfHealthCheck   bool
        ClusterHealthCheck bool
        PeriodicHeartbeat  bool
    }

    // 5. 连接管理
    ConnectionManagement struct {
        MaintainConnection bool
        AutoReconnect      bool
        BackoffRetry       bool
    }
}
```

### 2.3 中央控制平面职责

以下是中央控制平面(Agent Manager)需要实现的核心功能模块:

```go
type CentralControlPlaneResponsibilities struct {
    // 1. Agent 管理
    AgentManagement struct {
        RegisterAgent      bool
        TrackAgentStatus   bool
        ConfigureAgent     bool
        DeregisterAgent    bool
    }

    // 2. 事件聚合
    EventAggregation struct {
        ReceiveEvents      bool
        NormalizeEvents    bool
        CorrelateEvents    bool
        CreateTasks        bool
    }

    // 3. 任务编排
    TaskOrchestration struct {
        CreateDiagnosticTasks bool
        ScheduleTasks         bool
        DistributeCommands    bool
        TrackTaskProgress     bool
    }

    // 4. 智能诊断
    IntelligentDiagnosis struct {
        AnalyzeEvents      bool
        GenerateStrategies bool
        ExecutePlans       bool
        GenerateReports    bool
    }

    // 5. 多集群视图
    MultiClusterView struct {
        AggregateMetrics   bool
        GlobalDashboard    bool
        CrossClusterAnalysis bool
    }
}
```

### 2.4 通信模式

#### Push 模式 (Agent → Central)

```
Agent                              Central
  │                                   │
  │  1. Register (cluster_id, info)  │
  ├──────────────────────────────────►│
  │                                   │
  │  2. Heartbeat (every 30s)        │
  ├──────────────────────────────────►│
  │                                   │
  │  3. Event Report                 │
  ├──────────────────────────────────►│
  │                                   │
  │  4. Metrics Report (every 1m)    │
  ├──────────────────────────────────►│
  │                                   │
```

#### 命令下发模式 (Central → Agent 通过订阅实现)

**通信机制说明**: Agent通过NATS订阅`agent.command.<cluster_id>`主题实现命令接收。虽然底层是订阅机制（类似Pull），但从业务视角看，这是Central主动下发命令到指定Agent的Push模式。

**工作流程**:

```
Central                            Agent
  │                                   │
  │                              (持续订阅)
  │                          agent.command.<id>
  │                                   │
  │  1. Publish Command              │
  │     to agent.command.<id>        │
  ├──────────────────────────────────►│
  │                                   │
  │                              (收到消息)
  │                                   │ Execute Command
  │                                   │
  │  2. Publish Result               │
  │     to agent.result.<id>         │
  │◄──────────────────────────────────┤
  │                                   │
```

**技术特点**:
- Agent通过NATS订阅建立长连接，无需轮询
- Central通过Publish将命令发送到指定主题
- NATS消息总线保证消息可靠传递
- 实现了类似RPC的同步调用效果

## 3. 核心实现

### 3.1 Agent 服务实现

```go
package agent

import (
    "context"
    "encoding/json"
    "fmt"
    "sync"
    "time"

    "github.com/nats-io/nats.go"
    "go.uber.org/zap"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"

    "github.com/aetherius/k8s-agent/pkg/version"
)

// 全局日志实例(实际使用时应通过依赖注入)
var log *zap.Logger

type AgentConfig struct {
    ClusterID        string        `yaml:"cluster_id"`
    CentralEndpoint  string        `yaml:"central_endpoint"`
    ReconnectDelay   time.Duration `yaml:"reconnect_delay" default:"5s"`
    HeartbeatInterval time.Duration `yaml:"heartbeat_interval" default:"30s"`
    MetricsInterval  time.Duration `yaml:"metrics_interval" default:"60s"`
    BufferSize       int           `yaml:"buffer_size" default:"1000"`
    MaxRetries       int           `yaml:"max_retries" default:"10"`
}

type Agent struct {
    config          AgentConfig
    clusterID       string
    clientset       kubernetes.Interface
    natsConn        *nats.Conn
    eventWatcher    *EventWatcher
    metricsCollector *MetricsCollector
    commandExecutor *CommandExecutor

    eventChan       chan *Event
    metricsChan     chan *Metrics
    commandChan     chan *Command
    resultChan      chan *CommandResult

    stopCh          chan struct{}
    wg              sync.WaitGroup
    running         bool
    mu              sync.RWMutex
}

func NewAgent(config AgentConfig) (*Agent, error) {
    kubeConfig, err := rest.InClusterConfig()
    if err != nil {
        return nil, fmt.Errorf("failed to create in-cluster config: %w", err)
    }

    clientset, err := kubernetes.NewForConfig(kubeConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to create clientset: %w", err)
    }

    if config.ClusterID == "" {
        detector := &ClusterIDDetector{clientset: clientset}
        config.ClusterID, err = detector.DetectClusterID(context.Background())
        if err != nil {
            return nil, fmt.Errorf("failed to detect cluster ID: %w", err)
        }
    }

    agent := &Agent{
        config:      config,
        clusterID:   config.ClusterID,
        clientset:   clientset,
        eventChan:   make(chan *Event, config.BufferSize),
        metricsChan: make(chan *Metrics, 100),
        commandChan: make(chan *Command, 100),
        resultChan:  make(chan *CommandResult, 100),
        stopCh:      make(chan struct{}),
    }

    return agent, nil
}

func (a *Agent) Start(ctx context.Context) error {
    a.mu.Lock()
    if a.running {
        a.mu.Unlock()
        return fmt.Errorf("agent already running")
    }
    a.running = true
    a.mu.Unlock()

    log.Info("Starting agent",
        zap.String("cluster_id", a.clusterID),
        zap.String("central_endpoint", a.config.CentralEndpoint))

    if err := a.connectToCentral(); err != nil {
        return fmt.Errorf("failed to connect to central: %w", err)
    }

    if err := a.register(); err != nil {
        return fmt.Errorf("failed to register agent: %w", err)
    }

    a.eventWatcher = NewEventWatcher(a.clientset, a.clusterID, a.eventChan)
    if err := a.eventWatcher.Start(ctx); err != nil {
        return fmt.Errorf("failed to start event watcher: %w", err)
    }

    a.metricsCollector = NewMetricsCollector(a.clientset, a.clusterID, a.metricsChan)
    go a.metricsCollector.Start(ctx, a.config.MetricsInterval)

    a.commandExecutor = NewCommandExecutor(a.clientset, a.clusterID)

    a.wg.Add(5)
    go a.eventReporter(ctx)
    go a.metricsReporter(ctx)
    go a.commandListener(ctx)
    go a.resultReporter(ctx)
    go a.heartbeat(ctx)

    log.Info("Agent started successfully")

    <-ctx.Done()
    return a.Stop()
}

func (a *Agent) Stop() error {
    a.mu.Lock()
    defer a.mu.Unlock()

    if !a.running {
        return nil
    }

    log.Info("Stopping agent", zap.String("cluster_id", a.clusterID))

    close(a.stopCh)

    if a.eventWatcher != nil {
        a.eventWatcher.Stop()
    }

    if a.metricsCollector != nil {
        a.metricsCollector.Stop()
    }

    a.wg.Wait()

    if a.natsConn != nil {
        a.natsConn.Close()
    }

    a.running = false
    log.Info("Agent stopped")

    return nil
}

func (a *Agent) connectToCentral() error {
    opts := []nats.Option{
        nats.Name(fmt.Sprintf("agent-%s", a.clusterID)),
        nats.ReconnectWait(a.config.ReconnectDelay),
        nats.MaxReconnects(a.config.MaxRetries),
        nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
            log.Warn("Disconnected from central",
                zap.String("cluster_id", a.clusterID),
                zap.Error(err))
        }),
        nats.ReconnectHandler(func(nc *nats.Conn) {
            log.Info("Reconnected to central",
                zap.String("cluster_id", a.clusterID),
                zap.String("url", nc.ConnectedUrl()))
        }),
        nats.ClosedHandler(func(nc *nats.Conn) {
            log.Warn("Connection closed",
                zap.String("cluster_id", a.clusterID))
        }),
    }

    nc, err := nats.Connect(a.config.CentralEndpoint, opts...)
    if err != nil {
        return fmt.Errorf("failed to connect to NATS: %w", err)
    }

    a.natsConn = nc
    log.Info("Connected to central", zap.String("url", nc.ConnectedUrl()))

    return nil
}

func (a *Agent) register() error {
    agentInfo := AgentInfo{
        ClusterID:   a.clusterID,
        Version:     version.Version,
        StartTime:   time.Now(),
        Capabilities: []string{"event_watch", "metrics_collect", "command_execute"},
    }

    data, err := json.Marshal(agentInfo)
    if err != nil {
        return fmt.Errorf("failed to marshal agent info: %w", err)
    }

    subject := fmt.Sprintf("agent.register.%s", a.clusterID)
    if err := a.natsConn.Publish(subject, data); err != nil {
        return fmt.Errorf("failed to publish register message: %w", err)
    }

    log.Info("Agent registered", zap.String("cluster_id", a.clusterID))
    return nil
}

func (a *Agent) eventReporter(ctx context.Context) {
    defer a.wg.Done()

    for {
        select {
        case <-ctx.Done():
            return
        case <-a.stopCh:
            return
        case event := <-a.eventChan:
            if err := a.reportEvent(event); err != nil {
                log.Error("Failed to report event",
                    zap.Error(err),
                    zap.String("event_id", event.ID))
            }
        }
    }
}

func (a *Agent) reportEvent(event *Event) error {
    event.ClusterID = a.clusterID
    event.ReportedAt = time.Now()

    data, err := json.Marshal(event)
    if err != nil {
        return fmt.Errorf("failed to marshal event: %w", err)
    }

    subject := fmt.Sprintf("agent.event.%s", a.clusterID)
    if err := a.natsConn.Publish(subject, data); err != nil {
        return fmt.Errorf("failed to publish event: %w", err)
    }

    log.Debug("Event reported",
        zap.String("cluster_id", a.clusterID),
        zap.String("event_id", event.ID))

    return nil
}

func (a *Agent) metricsReporter(ctx context.Context) {
    defer a.wg.Done()

    for {
        select {
        case <-ctx.Done():
            return
        case <-a.stopCh:
            return
        case metrics := <-a.metricsChan:
            if err := a.reportMetrics(metrics); err != nil {
                log.Error("Failed to report metrics", zap.Error(err))
            }
        }
    }
}

func (a *Agent) reportMetrics(metrics *Metrics) error {
    metrics.ClusterID = a.clusterID
    metrics.Timestamp = time.Now()

    data, err := json.Marshal(metrics)
    if err != nil {
        return fmt.Errorf("failed to marshal metrics: %w", err)
    }

    subject := fmt.Sprintf("agent.metrics.%s", a.clusterID)
    if err := a.natsConn.Publish(subject, data); err != nil {
        return fmt.Errorf("failed to publish metrics: %w", err)
    }

    log.Debug("Metrics reported", zap.String("cluster_id", a.clusterID))
    return nil
}

func (a *Agent) commandListener(ctx context.Context) {
    defer a.wg.Done()

    subject := fmt.Sprintf("agent.command.%s", a.clusterID)
    sub, err := a.natsConn.Subscribe(subject, func(msg *nats.Msg) {
        var cmd Command
        if err := json.Unmarshal(msg.Data, &cmd); err != nil {
            log.Error("Failed to unmarshal command", zap.Error(err))
            return
        }

        log.Info("Received command",
            zap.String("cluster_id", a.clusterID),
            zap.String("command_id", cmd.ID))

        result := a.commandExecutor.Execute(ctx, cmd)
        a.resultChan <- result
    })

    if err != nil {
        log.Error("Failed to subscribe to commands", zap.Error(err))
        return
    }
    defer sub.Unsubscribe()

    <-ctx.Done()
}

func (a *Agent) resultReporter(ctx context.Context) {
    defer a.wg.Done()

    for {
        select {
        case <-ctx.Done():
            return
        case <-a.stopCh:
            return
        case result := <-a.resultChan:
            if err := a.reportResult(result); err != nil {
                log.Error("Failed to report result",
                    zap.Error(err),
                    zap.String("command_id", result.CommandID))
            }
        }
    }
}

func (a *Agent) reportResult(result *CommandResult) error {
    result.ClusterID = a.clusterID

    data, err := json.Marshal(result)
    if err != nil {
        return fmt.Errorf("failed to marshal result: %w", err)
    }

    subject := fmt.Sprintf("agent.result.%s", a.clusterID)
    if err := a.natsConn.Publish(subject, data); err != nil {
        return fmt.Errorf("failed to publish result: %w", err)
    }

    log.Info("Result reported",
        zap.String("cluster_id", a.clusterID),
        zap.String("command_id", result.CommandID))

    return nil
}

func (a *Agent) heartbeat(ctx context.Context) {
    defer a.wg.Done()

    ticker := time.NewTicker(a.config.HeartbeatInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-a.stopCh:
            return
        case <-ticker.C:
            if err := a.sendHeartbeat(); err != nil {
                log.Error("Failed to send heartbeat",
                    zap.Error(err),
                    zap.String("cluster_id", a.clusterID))
            }
        }
    }
}

func (a *Agent) sendHeartbeat() error {
    hb := Heartbeat{
        ClusterID: a.clusterID,
        Timestamp: time.Now(),
        Status:    "healthy",
        Metrics: HeartbeatMetrics{
            EventQueueSize:   len(a.eventChan),
            MetricsQueueSize: len(a.metricsChan),
            CommandQueueSize: len(a.commandChan),
        },
    }

    data, err := json.Marshal(hb)
    if err != nil {
        return fmt.Errorf("failed to marshal heartbeat: %w", err)
    }

    subject := fmt.Sprintf("agent.heartbeat.%s", a.clusterID)
    if err := a.natsConn.Publish(subject, data); err != nil {
        return fmt.Errorf("failed to publish heartbeat: %w", err)
    }

    log.Debug("Heartbeat sent", zap.String("cluster_id", a.clusterID))
    return nil
}
```

### 3.2 中央 Agent Manager 实现

```go
package central

import (
    "context"
    "encoding/json"
    "fmt"
    "sync"
    "time"

    "github.com/nats-io/nats.go"
    "go.uber.org/zap"
)

// 全局日志实例(实际使用时应通过依赖注入)
var log *zap.Logger

type AgentManagerConfig struct {
    NATSEndpoint     string        `yaml:"nats_endpoint"`
    HeartbeatTimeout time.Duration `yaml:"heartbeat_timeout" default:"90s"`
    CleanupInterval  time.Duration `yaml:"cleanup_interval" default:"5m"`
}

type AgentManager struct {
    config    AgentManagerConfig
    natsConn  *nats.Conn
    agents    map[string]*AgentState
    mu        sync.RWMutex
    stopCh    chan struct{}
    wg        sync.WaitGroup
}

type AgentState struct {
    ClusterID    string
    Info         AgentInfo
    Status       string
    LastHeartbeat time.Time
    RegisteredAt time.Time
    EventCount   int64
    MetricsCount int64
    CommandCount int64
}

func NewAgentManager(config AgentManagerConfig) (*AgentManager, error) {
    nc, err := nats.Connect(config.NATSEndpoint)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to NATS: %w", err)
    }

    am := &AgentManager{
        config:   config,
        natsConn: nc,
        agents:   make(map[string]*AgentState),
        stopCh:   make(chan struct{}),
    }

    return am, nil
}

func (am *AgentManager) Start(ctx context.Context) error {
    log.Info("Starting Agent Manager")

    if err := am.subscribeToAgents(); err != nil {
        return fmt.Errorf("failed to subscribe to agents: %w", err)
    }

    am.wg.Add(1)
    go am.cleanupStaleAgents(ctx)

    log.Info("Agent Manager started")

    <-ctx.Done()
    return am.Stop()
}

func (am *AgentManager) Stop() error {
    log.Info("Stopping Agent Manager")

    close(am.stopCh)
    am.wg.Wait()

    if am.natsConn != nil {
        am.natsConn.Close()
    }

    log.Info("Agent Manager stopped")
    return nil
}

func (am *AgentManager) subscribeToAgents() error {
    if _, err := am.natsConn.Subscribe("agent.register.*", am.handleRegister); err != nil {
        return fmt.Errorf("failed to subscribe to register: %w", err)
    }

    if _, err := am.natsConn.Subscribe("agent.heartbeat.*", am.handleHeartbeat); err != nil {
        return fmt.Errorf("failed to subscribe to heartbeat: %w", err)
    }

    if _, err := am.natsConn.Subscribe("agent.event.*", am.handleEvent); err != nil {
        return fmt.Errorf("failed to subscribe to events: %w", err)
    }

    if _, err := am.natsConn.Subscribe("agent.metrics.*", am.handleMetrics); err != nil {
        return fmt.Errorf("failed to subscribe to metrics: %w", err)
    }

    if _, err := am.natsConn.Subscribe("agent.result.*", am.handleResult); err != nil {
        return fmt.Errorf("failed to subscribe to results: %w", err)
    }

    log.Info("Subscribed to agent topics")
    return nil
}

func (am *AgentManager) handleRegister(msg *nats.Msg) {
    var info AgentInfo
    if err := json.Unmarshal(msg.Data, &info); err != nil {
        log.Error("Failed to unmarshal agent info", zap.Error(err))
        return
    }

    am.mu.Lock()
    defer am.mu.Unlock()

    am.agents[info.ClusterID] = &AgentState{
        ClusterID:     info.ClusterID,
        Info:          info,
        Status:        "online",
        RegisteredAt:  time.Now(),
        LastHeartbeat: time.Now(),
    }

    log.Info("Agent registered",
        zap.String("cluster_id", info.ClusterID),
        zap.String("version", info.Version))
}

func (am *AgentManager) handleHeartbeat(msg *nats.Msg) {
    var hb Heartbeat
    if err := json.Unmarshal(msg.Data, &hb); err != nil {
        log.Error("Failed to unmarshal heartbeat", zap.Error(err))
        return
    }

    am.mu.Lock()
    defer am.mu.Unlock()

    if agent, exists := am.agents[hb.ClusterID]; exists {
        agent.LastHeartbeat = time.Now()
        agent.Status = hb.Status

        log.Debug("Heartbeat received",
            zap.String("cluster_id", hb.ClusterID),
            zap.String("status", hb.Status))
    }
}

func (am *AgentManager) handleEvent(msg *nats.Msg) {
    var event Event
    if err := json.Unmarshal(msg.Data, &event); err != nil {
        log.Error("Failed to unmarshal event", zap.Error(err))
        return
    }

    am.mu.Lock()
    if agent, exists := am.agents[event.ClusterID]; exists {
        agent.EventCount++
    }
    am.mu.Unlock()

    if err := am.processEvent(&event); err != nil {
        log.Error("Failed to process event",
            zap.Error(err),
            zap.String("event_id", event.ID))
    }
}

func (am *AgentManager) processEvent(event *Event) error {
    log.Info("Processing event from agent",
        zap.String("cluster_id", event.ClusterID),
        zap.String("event_id", event.ID))

    data, err := json.Marshal(event)
    if err != nil {
        return fmt.Errorf("failed to marshal event: %w", err)
    }

    if err := am.natsConn.Publish("event.received", data); err != nil {
        return fmt.Errorf("failed to publish to orchestrator: %w", err)
    }

    return nil
}

func (am *AgentManager) handleMetrics(msg *nats.Msg) {
    var metrics Metrics
    if err := json.Unmarshal(msg.Data, &metrics); err != nil {
        log.Error("Failed to unmarshal metrics", zap.Error(err))
        return
    }

    am.mu.Lock()
    if agent, exists := am.agents[metrics.ClusterID]; exists {
        agent.MetricsCount++
    }
    am.mu.Unlock()

    log.Debug("Metrics received",
        zap.String("cluster_id", metrics.ClusterID))
}

func (am *AgentManager) handleResult(msg *nats.Msg) {
    var result CommandResult
    if err := json.Unmarshal(msg.Data, &result); err != nil {
        log.Error("Failed to unmarshal result", zap.Error(err))
        return
    }

    log.Info("Command result received",
        zap.String("cluster_id", result.ClusterID),
        zap.String("command_id", result.CommandID),
        zap.String("status", result.Status))

    data, err := json.Marshal(result)
    if err != nil {
        log.Error("Failed to marshal command result", zap.Error(err))
        return
    }

    if err := am.natsConn.Publish("command.result", data); err != nil {
        log.Error("Failed to publish result", zap.Error(err))
    }
}

func (am *AgentManager) SendCommand(clusterID string, cmd Command) error {
    am.mu.RLock()
    agent, exists := am.agents[clusterID]
    am.mu.RUnlock()

    if !exists {
        return fmt.Errorf("agent not found: %s", clusterID)
    }

    if agent.Status != "online" {
        return fmt.Errorf("agent offline: %s", clusterID)
    }

    subject := fmt.Sprintf("agent.command.%s", clusterID)
    data, err := json.Marshal(cmd)
    if err != nil {
        return fmt.Errorf("failed to marshal command: %w", err)
    }

    if err := am.natsConn.Publish(subject, data); err != nil {
        return fmt.Errorf("failed to publish command: %w", err)
    }

    log.Info("Command sent to agent",
        zap.String("cluster_id", clusterID),
        zap.String("command_id", cmd.ID))

    return nil
}

func (am *AgentManager) GetAgentStatus(clusterID string) (*AgentState, error) {
    am.mu.RLock()
    defer am.mu.RUnlock()

    agent, exists := am.agents[clusterID]
    if !exists {
        return nil, fmt.Errorf("agent not found: %s", clusterID)
    }

    return agent, nil
}

func (am *AgentManager) ListAgents() []*AgentState {
    am.mu.RLock()
    defer am.mu.RUnlock()

    agents := make([]*AgentState, 0, len(am.agents))
    for _, agent := range am.agents {
        agents = append(agents, agent)
    }

    return agents
}

func (am *AgentManager) cleanupStaleAgents(ctx context.Context) {
    defer am.wg.Done()

    ticker := time.NewTicker(am.config.CleanupInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-am.stopCh:
            return
        case <-ticker.C:
            am.cleanupStale()
        }
    }
}

func (am *AgentManager) cleanupStale() {
    am.mu.Lock()
    defer am.mu.Unlock()

    now := time.Now()
    for clusterID, agent := range am.agents {
        if now.Sub(agent.LastHeartbeat) > am.config.HeartbeatTimeout {
            log.Warn("Agent timeout, marking as offline",
                zap.String("cluster_id", clusterID),
                zap.Duration("last_heartbeat", now.Sub(agent.LastHeartbeat)))

            agent.Status = "offline"
        }
    }
}
```

### 3.3 数据模型

```go
type AgentInfo struct {
    ClusterID    string    `json:"cluster_id"`
    Version      string    `json:"version"`
    StartTime    time.Time `json:"start_time"`
    Capabilities []string  `json:"capabilities"`
}

type Event struct {
    ID         string                 `json:"id"`
    ClusterID  string                 `json:"cluster_id"`
    Type       string                 `json:"type"`
    Source     string                 `json:"source"`
    Namespace  string                 `json:"namespace"`
    Severity   string                 `json:"severity"`
    Message    string                 `json:"message"`
    Timestamp  time.Time              `json:"timestamp"`
    ReportedAt time.Time              `json:"reported_at"`
    Labels     map[string]string      `json:"labels"`
    RawData    map[string]interface{} `json:"raw_data"`
}

type Metrics struct {
    ClusterID string                 `json:"cluster_id"`
    Timestamp time.Time              `json:"timestamp"`
    Data      map[string]interface{} `json:"data"`
}

type Command struct {
    ID        string            `json:"id"`
    Type      string            `json:"type"`
    Tool      string            `json:"tool"`
    Action    string            `json:"action"`
    Args      []string          `json:"args"`
    Timeout   time.Duration     `json:"timeout"`
    CreatedAt time.Time         `json:"created_at"`
}

type CommandResult struct {
    CommandID string        `json:"command_id"`
    ClusterID string        `json:"cluster_id"`
    Status    string        `json:"status"`
    Output    string        `json:"output"`
    Error     string        `json:"error,omitempty"`
    Duration  time.Duration `json:"duration"`
    Timestamp time.Time     `json:"timestamp"`
}

type Heartbeat struct {
    ClusterID string            `json:"cluster_id"`
    Timestamp time.Time         `json:"timestamp"`
    Status    string            `json:"status"`
    Metrics   HeartbeatMetrics  `json:"metrics"`
}

type HeartbeatMetrics struct {
    EventQueueSize   int `json:"event_queue_size"`
    MetricsQueueSize int `json:"metrics_queue_size"`
    CommandQueueSize int `json:"command_queue_size"`
}
```

## 4. 通信协议

### 4.1 数据流向说明

完整的事件处理链路:

```
1. Agent监听K8s事件
      ↓
2. Agent上报事件到 agent.event.<cluster_id>
      ↓
3. Agent Manager接收并转发到 event.received
      ↓
4. Orchestrator Service订阅 event.received 创建诊断任务
      ↓
5. Orchestrator分析后通过Agent Manager下发命令到 agent.command.<cluster_id>
      ↓
6. Agent执行命令并上报结果到 agent.result.<cluster_id>
      ↓
7. Agent Manager转发结果到 command.result
      ↓
8. Orchestrator完成任务处理
```

### 4.2 NATS Subject 设计

```
agent.register.<cluster_id>       # Agent注册
agent.heartbeat.<cluster_id>      # Agent心跳
agent.event.<cluster_id>          # Agent上报事件
agent.metrics.<cluster_id>        # Agent上报指标
agent.result.<cluster_id>         # Agent上报命令结果
agent.command.<cluster_id>        # Central下发命令

event.received                    # 转发给Orchestrator
command.result                    # 转发命令结果
```

### 4.3 消息格式

#### Agent注册消息

```json
{
  "cluster_id": "prod-us-west-2",
  "version": "v1.0.0",
  "start_time": "2025-09-28T10:00:00Z",
  "capabilities": ["event_watch", "metrics_collect", "command_execute"]
}
```

#### 事件上报消息

```json
{
  "id": "evt-12345",
  "cluster_id": "prod-us-west-2",
  "type": "k8s_event",
  "source": "kubernetes",
  "namespace": "default",
  "severity": "high",
  "message": "Pod CrashLoopBackOff",
  "timestamp": "2025-09-28T10:05:00Z",
  "reported_at": "2025-09-28T10:05:01Z",
  "labels": {
    "reason": "CrashLoopBackOff",
    "pod": "app-1234"
  }
}
```

#### 命令下发消息

```json
{
  "id": "cmd-67890",
  "type": "diagnostic",
  "tool": "kubectl",
  "action": "logs",
  "args": ["pod", "app-1234", "-n", "default"],
  "timeout": "30s",
  "created_at": "2025-09-28T10:05:05Z"
}
```

#### 命令结果消息

```json
{
  "command_id": "cmd-67890",
  "cluster_id": "prod-us-west-2",
  "status": "success",
  "output": "...[logs]...",
  "duration": "1.5s",
  "timestamp": "2025-09-28T10:05:06Z"
}
```

## 5. 部署配置

### 5.1 Agent 部署 (每个集群)

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: aetherius-agent
  labels:
    app.kubernetes.io/name: aetherius
    app.kubernetes.io/component: agent
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: aetherius-agent
  namespace: aetherius-agent
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: aetherius-agent
rules:
- apiGroups: [""]
  resources: ["events", "pods", "nodes", "namespaces"]
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources: ["pods/log"]
  verbs: ["get"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: aetherius-agent
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: aetherius-agent
subjects:
- kind: ServiceAccount
  name: aetherius-agent
  namespace: aetherius-agent
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: agent-config
  namespace: aetherius-agent
data:
  config.yaml: |
    cluster_id: ""  # 留空自动检测
    central_endpoint: "nats://central.aetherius.io:4222"
    reconnect_delay: 5s
    heartbeat_interval: 30s
    metrics_interval: 60s
    buffer_size: 1000
    max_retries: 10
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: aetherius-agent
  namespace: aetherius-agent
spec:
  replicas: 1
  strategy:
    type: Recreate
  selector:
    matchLabels:
      app: aetherius-agent
  template:
    metadata:
      labels:
        app: aetherius-agent
    spec:
      serviceAccountName: aetherius-agent
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
      containers:
      - name: agent
        image: aetherius/agent:v1.0
        args:
        - --config=/etc/aetherius/config.yaml
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
        - name: CLUSTER_ID
          valueFrom:
            configMapKeyRef:
              name: agent-config
              key: cluster_id
              optional: true
        ports:
        - name: http
          containerPort: 8080
        - name: metrics
          containerPort: 9090
        volumeMounts:
        - name: config
          mountPath: /etc/aetherius
          readOnly: true
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "250m"
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          capabilities:
            drop:
            - ALL
      volumes:
      - name: config
        configMap:
          name: agent-config
```

### 5.2 Central 部署

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: agent-manager
  namespace: aetherius-system
spec:
  replicas: 2
  selector:
    matchLabels:
      app: agent-manager
  template:
    metadata:
      labels:
        app: agent-manager
    spec:
      containers:
      - name: agent-manager
        image: aetherius/agent-manager:v1.0
        env:
        - name: NATS_ENDPOINT
          value: "nats://nats.aetherius-system.svc.cluster.local:4222"
        ports:
        - name: http
          containerPort: 8090
        - name: metrics
          containerPort: 9090
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: agent-manager
  namespace: aetherius-system
spec:
  selector:
    app: agent-manager
  ports:
  - name: http
    port: 8090
    targetPort: 8090
  - name: metrics
    port: 9090
    targetPort: 9090
```

### 5.3 NATS 部署 (Agent通信专用)

#### 双NATS架构说明

**重要**: Aetherius系统采用**双NATS集群架构**，分别服务不同的通信场景：

| NATS集群 | 部署位置 | 网络暴露 | 服务对象 | 主要用途 | 参考文档 |
|---------|---------|---------|---------|---------|---------|
| **微服务内部NATS** | Central集群内 | ClusterIP<br>(仅内网) | 中央控制平面各服务 | • Orchestrator ↔ Reasoning<br>• Orchestrator ↔ Execution<br>• 服务间消息传递 | [06_microservices.md](./06_microservices.md) |
| **Agent通信NATS**<br>(本节部署) | Central集群<br>或独立部署 | LoadBalancer<br>(公网暴露) | 边缘Agent集群 | • Agent注册管理<br>• 事件数据上报<br>• 命令下发接收 | 本文档 |

#### 架构设计原因

1. **安全隔离**: 微服务内部通信不对外暴露，避免安全风险
2. **性能优化**: 内部NATS优化低延迟，Agent NATS优化广域网通信
3. **独立扩展**: 两个集群可根据各自负载独立扩展
4. **故障隔离**: Agent通信故障不影响中央控制平面内部服务

#### Agent通信NATS部署要求

**网络配置**:
- 必须对外暴露（LoadBalancer或NodePort）
- 所有边缘Agent集群需能访问此NATS服务
- 建议配置TLS加密和身份认证

**部署配置**:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: nats
  namespace: aetherius-system
spec:
  selector:
    app: nats
  ports:
  - name: client
    port: 4222
    targetPort: 4222
  - name: cluster
    port: 6222
    targetPort: 6222
  - name: monitor
    port: 8222
    targetPort: 8222
  type: LoadBalancer  # 对外暴露,供Agent连接
---
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: nats
  namespace: aetherius-system
spec:
  serviceName: nats
  replicas: 3
  selector:
    matchLabels:
      app: nats
  template:
    metadata:
      labels:
        app: nats
    spec:
      containers:
      - name: nats
        image: nats:2.10-alpine
        args:
        - -c
        - /etc/nats-config/nats.conf
        ports:
        - containerPort: 4222
          name: client
        - containerPort: 6222
          name: cluster
        - containerPort: 8222
          name: monitor
        volumeMounts:
        - name: config
          mountPath: /etc/nats-config
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
      volumes:
      - name: config
        configMap:
          name: nats-config
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: nats-config
  namespace: aetherius-system
data:
  nats.conf: |
    port: 4222
    http_port: 8222

    cluster {
      name: aetherius
      port: 6222
      routes: [
        nats://nats-0.nats.aetherius-system.svc.cluster.local:6222
        nats://nats-1.nats.aetherius-system.svc.cluster.local:6222
        nats://nats-2.nats.aetherius-system.svc.cluster.local:6222
      ]
    }

    max_payload: 10485760
    max_connections: 10000
```

## 6. 运维管理

### 6.1 Agent 管理 API

```go
type AgentManagerAPI struct {
    manager *AgentManager
}

func (api *AgentManagerAPI) ListAgents(w http.ResponseWriter, r *http.Request) {
    agents := api.manager.ListAgents()
    json.NewEncoder(w).Encode(agents)
}

func (api *AgentManagerAPI) GetAgentStatus(w http.ResponseWriter, r *http.Request) {
    clusterID := mux.Vars(r)["cluster_id"]

    agent, err := api.manager.GetAgentStatus(clusterID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    json.NewEncoder(w).Encode(agent)
}

func (api *AgentManagerAPI) SendCommand(w http.ResponseWriter, r *http.Request) {
    clusterID := mux.Vars(r)["cluster_id"]

    var cmd Command
    if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    if err := api.manager.SendCommand(clusterID, cmd); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusAccepted)
}
```

### 6.2 监控指标

```go
var (
    agentTotalGauge = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "aetherius_agent_total",
            Help: "Total number of agents",
        },
        []string{"status"},
    )

    agentEventsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "aetherius_agent_events_total",
            Help: "Total events received from agents",
        },
        []string{"cluster_id"},
    )

    agentCommandsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "aetherius_agent_commands_total",
            Help: "Total commands sent to agents",
        },
        []string{"cluster_id", "status"},
    )
)
```

### 6.3 日志规范

```go
log.Info("Agent event",
    zap.String("cluster_id", clusterID),
    zap.String("event_type", eventType),
    zap.String("severity", severity))

log.Warn("Agent heartbeat timeout",
    zap.String("cluster_id", clusterID),
    zap.Duration("timeout", timeout))

log.Error("Failed to send command",
    zap.String("cluster_id", clusterID),
    zap.String("command_id", cmdID),
    zap.Error(err))
```

### 6.4 故障排查

#### Agent 无法连接到 Central

```bash
# 检查Agent日志
kubectl logs -n aetherius-agent deployment/aetherius-agent

# 检查网络连通性
kubectl exec -n aetherius-agent deployment/aetherius-agent -- \
  nc -zv central.aetherius.io 4222

# 检查NATS状态
kubectl exec -n aetherius-system statefulset/nats-0 -- \
  nats-server -sl
```

#### Agent 心跳超时

```bash
# 查看Agent Manager日志
kubectl logs -n aetherius-system deployment/agent-manager

# 查看Agent列表
curl http://agent-manager.aetherius-system.svc.cluster.local:8090/agents

# 重启Agent
kubectl rollout restart -n aetherius-agent deployment/aetherius-agent
```

## 附录

### A. Agent 部署脚本

```bash
#!/bin/bash
# deploy-agent.sh

set -e

CLUSTER_ID=${1:-""}
CENTRAL_ENDPOINT=${2:-"nats://central.aetherius.io:4222"}

if [ -z "$CLUSTER_ID" ]; then
    echo "Usage: $0 <cluster-id> [central-endpoint]"
    exit 1
fi

echo "=== Deploying Aetherius Agent to cluster: $CLUSTER_ID ==="

kubectl apply -f manifests/agent/01-namespace.yaml
kubectl apply -f manifests/agent/02-rbac.yaml

cat manifests/agent/03-configmap.yaml | \
    sed "s|cluster_id: \"\"|cluster_id: \"$CLUSTER_ID\"|" | \
    sed "s|central_endpoint:.*|central_endpoint: \"$CENTRAL_ENDPOINT\"|" | \
    kubectl apply -f -

kubectl apply -f manifests/agent/04-deployment.yaml

kubectl -n aetherius-agent rollout status deployment/aetherius-agent --timeout=5m

echo "=== Agent deployed successfully ==="
kubectl -n aetherius-agent get pods
```

### B. 相关文档

- [In-Cluster部署文档](./08_in_cluster_deployment.md)
- [微服务架构文档](./06_microservices.md)
- [K8s事件监听文档](./07_k8s_event_watcher.md)