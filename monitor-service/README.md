# Monitor Service

监控管理微服务 - 负责系统监控、指标采集、告警管理等功能。

## 功能特性

### 1. 系统监控
- 实时监控 Agent 状态
- 监控事件流
- 监控命令执行状态
- 资源使用率监控

### 2. 指标采集
- Prometheus 指标暴露
- 自定义指标采集
- 聚合统计
- 历史数据存储

### 3. 告警管理
- 告警规则配置
- 多渠道告警通知（邮件、Webhook、Slack）
- 告警历史记录
- 告警静默和恢复

### 4. 数据分析
- 趋势分析
- 异常检测
- 报表生成

## 架构设计

```
monitor-service/
├── cmd/
│   └── server/          # 服务入口
│       └── main.go
├── configs/             # 配置文件
│   └── config.yaml
├── internal/
│   ├── api/            # API 路由
│   │   └── server.go
│   ├── handler/        # HTTP 处理器
│   │   ├── metrics.go
│   │   ├── alert.go
│   │   └── dashboard.go
│   ├── service/        # 业务逻辑
│   │   ├── monitor.go
│   │   ├── metrics.go
│   │   └── alert.go
│   ├── storage/        # 数据存储
│   │   ├── postgres.go
│   │   └── redis.go
│   ├── metrics/        # Prometheus 指标
│   │   └── collector.go
│   └── middleware/     # 中间件
│       ├── auth.go
│       └── logging.go
├── pkg/
│   └── types/          # 类型定义
│       └── types.go
├── Dockerfile
├── Makefile
└── README.md
```

## API 端点

### 监控指标
- `GET /api/v1/metrics/summary` - 获取监控概览
- `GET /api/v1/metrics/agents` - Agent 指标
- `GET /api/v1/metrics/events` - 事件指标
- `GET /api/v1/metrics/commands` - 命令执行指标
- `GET /api/v1/metrics/trends` - 趋势数据

### 告警管理
- `GET /api/v1/alerts` - 获取告警列表
- `POST /api/v1/alerts` - 创建告警规则
- `PUT /api/v1/alerts/:id` - 更新告警规则
- `DELETE /api/v1/alerts/:id` - 删除告警规则
- `POST /api/v1/alerts/:id/trigger` - 触发告警

### 仪表盘
- `GET /api/v1/dashboard/overview` - 系统概览
- `GET /api/v1/dashboard/charts` - 图表数据

## 配置说明

```yaml
server:
  port: 8081
  mode: debug

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: monitor_db

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

prometheus:
  enabled: true
  port: 9091

alert:
  channels:
    email:
      enabled: true
      smtp_host: smtp.gmail.com
      smtp_port: 587
    webhook:
      enabled: true
    slack:
      enabled: false
```

## 快速开始

### 构建
```bash
make build
```

### 运行
```bash
make run
```

### 测试
```bash
make test
```

### Docker 构建
```bash
make docker-build
```

## 依赖服务

- PostgreSQL: 存储监控数据
- Redis: 缓存和实时数据
- Prometheus: 指标采集（可选）

## 环境变量

- `MONITOR_CONFIG_PATH`: 配置文件路径
- `MONITOR_LOG_LEVEL`: 日志级别
- `MONITOR_PORT`: 服务端口

## License

MIT
