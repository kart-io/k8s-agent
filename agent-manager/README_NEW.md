# Agent Manager Service

Agent Manager Service 管理 k8s-agent 跨多个集群的部署和监控。

## 架构

采用 onex-usercenter 风格的 Options 模式架构：

```
agent-manager/
├── cmd/
│   ├── app/               # 应用启动逻辑
│   │   ├── app.go         # Cobra 命令和配置
│   │   └── server.go      # 服务器初始化和运行
│   └── server/
│       └── main.go        # 入口点
├── internal/
│   ├── config/
│   │   └── options.go     # Options 结构定义
│   ├── agent/             # Agent 管理
│   ├── api/               # REST API
│   ├── command/           # 命令分发
│   ├── event/             # 事件处理
│   ├── nats/              # NATS 服务器
│   └── storage/           # 存储层
├── configs/
│   ├── config.yaml        # 默认配置
│   └── config-dev.yaml    # 开发配置
└── Makefile              # 构建和运行命令
```

## 配置

### 配置文件

配置文件使用 YAML 格式，包含以下部分：

- `server` - HTTP 服务器配置
- `database` - MySQL 数据库配置
- `redis` - Redis 缓存配置
- `nats` - NATS 消息队列配置
- `logging` - 日志配置
- `metrics` - 指标采集配置

### 配置加载顺序

1. 默认值 (代码中定义)
2. 配置文件 (--config 参数)
3. 命令行参数
4. 环境变量 (AGENT_MANAGER_ 前缀)

### 环境变量

所有配置都可以通过环境变量覆盖：

```bash
AGENT_MANAGER_SERVER_PORT=8080
AGENT_MANAGER_DATABASE_HOST=mysql.example.com
AGENT_MANAGER_DATABASE_PASSWORD=secure_password
AGENT_MANAGER_REDIS_ADDR=redis.example.com:6379
```

## 运行

### 使用 Make 命令

```bash
# 使用默认配置运行
make run

# 使用开发配置运行
make run-dev

# 使用自定义配置运行
make run-config CONFIG=path/to/config.yaml

# 使用环境变量运行
make run-env

# 构建二进制文件
make build

# 运行测试
make test
```

### 直接运行

```bash
# 使用配置文件
go run cmd/server/main.go --config configs/config.yaml

# 使用命令行参数
go run cmd/server/main.go --server.port 8080 --database.host localhost

# 查看帮助
go run cmd/server/main.go --help
```

### Docker 运行

```bash
# 构建镜像
make docker-build

# 运行容器
docker run -p 8080:8080 \
  -v $(pwd)/configs:/configs \
  aetherius/agent-manager:v1.0.0 \
  --config /configs/config.yaml
```

## API 端点

- `GET /health/live` - 存活探测
- `GET /health/ready` - 就绪探测
- `GET /health/status` - 服务状态
- `GET /metrics` - Prometheus 指标

### Agent 管理
- `GET /api/v1/agents` - 列出所有 agents
- `GET /api/v1/agents/:id` - 获取 agent 详情
- `DELETE /api/v1/agents/:id` - 删除 agent

### 集群管理
- `GET /api/v1/clusters` - 列出所有集群
- `GET /api/v1/clusters/:id` - 获取集群详情
- `POST /api/v1/clusters` - 创建集群
- `PUT /api/v1/clusters/:id` - 更新集群
- `DELETE /api/v1/clusters/:id` - 删除集群

### 事件管理
- `GET /api/v1/events` - 列出事件
- `GET /api/v1/events/:id` - 获取事件详情
- `POST /api/v1/events/search` - 搜索事件

### 命令管理
- `POST /api/v1/commands` - 发送命令
- `GET /api/v1/commands/:id` - 获取命令状态
- `GET /api/v1/commands/:id/result` - 获取命令结果

## 开发

### 依赖

- Go 1.21+
- MySQL 8.0+
- Redis 6.0+
- NATS 2.0+

### 项目结构说明

- `cmd/app/` - 应用启动和配置管理 (Cobra + Viper)
- `internal/config/` - Options 配置结构
- `common/config/` - 通用配置 Options (ServerOptions, DatabaseOptions 等)

### Options 模式

采用 Functional Options Pattern：

```go
// 配置定义
type Options struct {
    Server   *config.ServerOptions
    Database *config.DatabaseOptions
    Redis    *config.RedisOptions
    // ...
}

// 创建默认配置
opts := config.NewOptions()

// 验证配置
if errs := opts.Validate(); len(errs) > 0 {
    // 处理错误
}

// 使用配置
server := NewServer(opts, logger)
```

### 添加新配置

1. 在 `common/config/` 中创建新的 Options 文件
2. 在 `internal/config/options.go` 中添加字段
3. 更新 `Validate()` 方法
4. 在 `cmd/app/app.go` 中添加命令行参数

## 监控

### 日志

- 支持 JSON 和文本格式
- 支持多个输出目标
- 支持 OTLP 导出

### 指标

- Prometheus 格式指标
- 默认端口 9090
- 路径 `/metrics`

### 健康检查

- Liveness: `/health/live`
- Readiness: `/health/ready`
- 包含数据库和 Redis 连接状态

## 许可

MIT License