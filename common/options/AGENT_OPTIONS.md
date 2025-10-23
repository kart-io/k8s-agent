# Agent Options - 通用 Agent 配置选项

## 概述

`AgentOptions` 是一个可重用的配置选项集,适用于所有 agent 类型的实现(collect-agent, monitor-agent等)。这个选项集被放置在 `common/options/` 中,遵循项目的 Options Pattern。

## 文件位置

```
common/options/agent_options.go
```

## 使用方式

### 1. 在服务中使用

```go
package config

import (
    "github.com/kart-io/k8s-agent/common/options"
)

type Options struct {
    Logging *options.LoggingOptions `json:"logging" mapstructure:"logging"`
    Agent   *options.AgentOptions   `json:"agent" mapstructure:"agent"`
}

func NewOptions() *Options {
    return &Options{
        Logging: options.NewLoggingOptions(),
        Agent:   options.NewAgentOptions(),
    }
}
```

### 2. 配置示例 (YAML)

```yaml
agent:
  # 集群标识
  cluster_id: "prod-cluster-1"
  cluster_name: "Production Cluster 1"

  # 中心端点
  central_endpoint: "nats://agent-manager:4222"

  # 连接管理
  reconnect_delay: 5s
  heartbeat_interval: 30s
  connection_timeout: 10s

  # 数据采集间隔
  metrics_interval: 60s
  event_interval: 5s

  # 缓冲区配置
  buffer_size: 1000
  event_queue_size: 500
  metrics_queue_size: 500

  # 重试配置
  max_retries: 10
  retry_backoff: 1s
  max_retry_backoff: 60s

  # 功能开关
  enable_metrics: true
  enable_events: true
  enable_tracing: false

  # 健康检查
  health_port: 8080
  enable_pprof: false
  pprof_port: 6060

  # 资源限制
  max_concurrent_requests: 100
  request_timeout: 30s
```

### 3. 命令行参数

```bash
# 集群配置
--cluster-id string                  集群唯一标识 (必需)
--cluster-name string                集群可读名称

# 连接配置
--central-endpoint string            NATS 中心端点 (默认: nats://localhost:4222)
--reconnect-delay duration           重连延迟 (默认: 5s, 最小: 1s)
--heartbeat-interval duration        心跳间隔 (默认: 30s, 最小: 10s)
--connection-timeout duration        连接超时 (默认: 10s, 最小: 1s)

# 数据采集
--metrics-interval duration          指标采集间隔 (默认: 60s, 最小: 30s)
--event-interval duration            事件处理间隔 (默认: 5s)

# 缓冲区
--buffer-size int                    主缓冲区大小 (默认: 1000, 最小: 10)
--event-queue-size int               事件队列大小 (默认: 500, 最小: 10)
--metrics-queue-size int             指标队列大小 (默认: 500, 最小: 10)

# 重试
--max-retries int                    最大重试次数 (默认: 10, 最小: 1)
--retry-backoff duration             初始退避时间 (默认: 1s)
--max-retry-backoff duration         最大退避时间 (默认: 60s)

# 功能开关
--enable-metrics                     启用指标采集 (默认: true)
--enable-events                      启用事件监控 (默认: true)
--enable-tracing                     启用分布式追踪 (默认: false)

# 监控
--health-port int                    健康检查端口 (默认: 8080)
--enable-pprof                       启用 pprof 分析 (默认: false)
--pprof-port int                     pprof 端口 (默认: 6060)

# 资源限制
--max-concurrent-requests int        最大并发请求数 (默认: 100)
--request-timeout duration           请求超时时间 (默认: 30s)
```

## 配置项说明

### 集群标识
- **cluster_id**: 集群的唯一标识符,通常使用 UUID 或集群编号
- **cluster_name**: 人类可读的集群名称,用于显示和日志

### 连接管理
- **central_endpoint**: NATS 消息队列的连接地址
- **reconnect_delay**: 断开连接后重连的等待时间
- **heartbeat_interval**: 向 agent-manager 发送心跳的间隔
- **connection_timeout**: 建立连接的超时时间

### 数据采集
- **metrics_interval**: 采集系统指标的间隔(CPU、内存等)
- **event_interval**: 处理 Kubernetes 事件的间隔

### 缓冲区配置
- **buffer_size**: 主缓冲区大小,用于临时存储数据
- **event_queue_size**: 事件队列的大小
- **metrics_queue_size**: 指标队列的大小

### 重试策略
- **max_retries**: 失败操作的最大重试次数
- **retry_backoff**: 首次重试前的等待时间
- **max_retry_backoff**: 重试等待时间的上限(指数退避)

### 功能开关
- **enable_metrics**: 是否启用指标采集功能
- **enable_events**: 是否启用事件监控功能
- **enable_tracing**: 是否启用分布式追踪

### 健康监控
- **health_port**: 健康检查 HTTP 端点的端口号
- **enable_pprof**: 是否启用 pprof 性能分析
- **pprof_port**: pprof HTTP 端点的端口号

### 资源限制
- **max_concurrent_requests**: 允许的最大并发请求数
- **request_timeout**: 单个请求的超时时间

## 验证规则

`AgentOptions` 实现了完整的验证逻辑:

- ✅ 所有时间间隔都有最小值限制
- ✅ 缓冲区大小必须足够大(≥10)
- ✅ 端口号范围验证(1-65535)
- ✅ 必填字段检查(central_endpoint)
- ✅ 逻辑一致性检查

## 默认值

所有配置项都有合理的默认值,可以开箱即用:

```go
opts := options.NewAgentOptions()
// opts 包含所有默认配置
```

## 向后兼容性

为了保持向后兼容,`internal/collect-agent/config/options.go` 提供了便捷的 getter 方法:

```go
opts := config.NewOptions()

// 通过 Agent 字段访问
clusterID := opts.Agent.ClusterID

// 或使用便捷方法
clusterID := opts.GetClusterID()
```

## 与其他 Options 的关系

`AgentOptions` 遵循项目的 Options Pattern,与其他选项配合使用:

```go
type ServiceOptions struct {
    Logging   *options.LoggingOptions   // 日志配置
    Agent     *options.AgentOptions     // Agent 配置
    Database  *options.DatabaseOptions  // 数据库配置(如需要)
    Redis     *options.RedisOptions     // Redis 配置(如需要)
}
```

## 迁移指南

### 从旧配置迁移

如果你的代码使用旧的配置结构:

```go
// 旧代码
type Options struct {
    ClusterID string
    CentralEndpoint string
    // ...
}
```

迁移到新结构:

```go
// 新代码
type Options struct {
    Agent *options.AgentOptions
}

// 访问字段
opts.Agent.ClusterID
opts.Agent.CentralEndpoint
```

### 配置文件迁移

旧的 YAML 配置:

```yaml
cluster_id: "prod-1"
central_endpoint: "nats://localhost:4222"
heartbeat_interval: 30s
```

新的 YAML 配置:

```yaml
agent:
  cluster_id: "prod-1"
  central_endpoint: "nats://localhost:4222"
  heartbeat_interval: 30s
```

## 扩展性

如果需要添加 agent 特定的配置,可以在服务的 Options 中扩展:

```go
type Options struct {
    Agent   *options.AgentOptions   // 通用 agent 配置

    // 服务特定配置
    CustomField string `json:"custom_field"`
}
```

## 最佳实践

1. **使用默认值**: 调用 `NewAgentOptions()` 获取合理的默认配置
2. **验证配置**: 启动前调用 `Validate()` 确保配置有效
3. **完成配置**: 使用 `Complete()` 填充缺失的默认值
4. **环境变量**: 通过 Viper 自动支持环境变量覆盖
5. **文档化**: 为每个配置项添加清晰的说明

## 示例代码

完整的使用示例:

```go
package main

import (
    "github.com/kart-io/k8s-agent/common/options"
)

func main() {
    // 创建配置
    opts := options.NewAgentOptions()

    // 自定义配置
    opts.ClusterID = "my-cluster"
    opts.CentralEndpoint = "nats://production:4222"
    opts.EnableMetrics = true

    // 完成配置(填充默认值)
    if err := opts.Complete(); err != nil {
        panic(err)
    }

    // 验证配置
    if err := opts.Validate(); err != nil {
        panic(err)
    }

    // 使用配置
    startAgent(opts)
}
```

## 相关文档

- [OPTIONS_PATTERN.md](../OPTIONS_PATTERN.md) - Options 模式说明
- [README.md](../README.md) - Common 包整体说明
- [CLAUDE.md](../../CLAUDE.md) - 项目配置指南
