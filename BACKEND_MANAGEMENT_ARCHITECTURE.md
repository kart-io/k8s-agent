# K8s Agent 后台管理系统架构文档

## 系统概述

K8s Agent 后台管理系统是一个基于微服务架构的 Kubernetes 集群监控和管理平台,采用前后端分离设计,通过 API 网关提供统一的服务入口。

## 架构图

```
┌─────────────────────────────────────────────────────────────────────┐
│                          前端应用层                                  │
│                      agent-manager-ui                               │
│                     (Vue 3 + Ant Design)                            │
└────────────────────────────┬────────────────────────────────────────┘
                             │ HTTP/HTTPS
                             ↓
┌─────────────────────────────────────────────────────────────────────┐
│                        API 网关层                                    │
│                      gateway-service                                │
│         (统一入口、认证、限流、路由转发、CORS)                       │
└──────┬──────────┬──────────┬──────────┬──────────┬──────────────────┘
       │          │          │          │          │
       ↓          ↓          ↓          ↓          ↓
┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────────┐
│  认证服务 │ │ 监控服务  │ │ 集群服务  │ │Agent服务 │ │ 编排服务      │
│   auth   │ │ monitor  │ │ cluster  │ │ manager │ │ orchestrator │
│ service  │ │ service  │ │ service  │ │ service │ │   service    │
│          │ │          │ │          │ │         │ │              │
│ :8090    │ │ :8081    │ │ :8082    │ │ :8080   │ │ :8083        │
└────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬────┘ └──────┬───────┘
     │            │            │            │             │
     │            │            │            │             │
     └────────────┴────────────┴────────────┴─────────────┘
                              │
                ┌─────────────┴─────────────┐
                │                           │
         ┌──────▼──────┐           ┌───────▼──────┐
         │ PostgreSQL  │           │    Redis     │
         │  (主存储)   │           │   (缓存)     │
         └─────────────┘           └──────────────┘
                │
         ┌──────▼──────┐
         │   NATS      │
         │ (消息队列)  │
         └─────────────┘
```

## 微服务列表

### 1. 认证服务 (auth-service)

**端口**: 8090
**职责**:
- 用户认证与授权
- JWT Token 管理
- RBAC 权限控制
- 用户、角色、权限管理
- API Key 管理

**主要 API**:
```
POST   /api/v1/auth/login           - 用户登录
POST   /api/v1/auth/logout          - 用户登出
GET    /api/v1/auth/me              - 获取当前用户信息
GET    /api/v1/auth/menus           - 获取用户菜单
GET    /api/v1/users                - 用户管理
GET    /api/v1/roles                - 角色管理
GET    /api/v1/permissions          - 权限管理
```

**技术栈**:
- Gin (Web 框架)
- JWT (认证)
- bcrypt (密码加密)
- PostgreSQL (数据存储)
- Redis (Token 缓存)

---

### 2. 监控服务 (monitor-service)

**端口**: 8081
**职责**:
- 系统监控数据采集
- 指标聚合与统计
- 告警规则管理
- 告警触发与通知
- 仪表盘数据提供

**主要 API**:
```
GET    /api/v1/metrics/summary      - 监控概览
GET    /api/v1/metrics/agents       - Agent 指标
GET    /api/v1/metrics/trends       - 趋势数据
GET    /api/v1/alerts               - 告警规则
POST   /api/v1/alerts               - 创建告警规则
GET    /api/v1/dashboard/overview   - 仪表盘概览
```

**技术栈**:
- Gin (Web 框架)
- PostgreSQL (指标存储)
- Redis (实时数据缓存)
- Prometheus (可选指标暴露)

---

### 3. 集群服务 (cluster-service)

**端口**: 8082
**职责**:
- K8s 集群管理
- 集群资源监控
- Pod/Deployment 管理
- 集群健康检查
- 资源统计

**主要 API**:
```
GET    /api/v1/clusters                      - 集群列表
POST   /api/v1/clusters                      - 添加集群
GET    /api/v1/clusters/:id/health           - 集群健康
GET    /api/v1/clusters/:id/nodes            - 节点列表
GET    /api/v1/clusters/:id/namespaces/:ns/pods  - Pod 列表
POST   /api/v1/clusters/:id/namespaces/:ns/deployments  - 创建 Deployment
GET    /api/v1/clusters/:id/stats            - 集群统计
```

**技术栈**:
- Gin (Web 框架)
- client-go (K8s 客户端)
- PostgreSQL (集群配置存储)

---

### 4. Agent 管理服务 (agent-manager)

**端口**: 8080
**职责**:
- Agent 注册与管理
- 事件收集与处理
- 命令分发与执行
- Agent 状态监控

**主要 API**:
```
GET    /api/v1/agents               - Agent 列表
GET    /api/v1/agents/:id           - Agent 详情
DELETE /api/v1/agents/:id           - 删除 Agent
GET    /api/v1/events               - 事件列表
POST   /api/v1/commands             - 创建命令
GET    /api/v1/commands             - 命令列表
```

**技术栈**:
- Gin (Web 框架)
- NATS (消息队列)
- PostgreSQL (数据存储)
- Redis (缓存)

---

### 5. 编排服务 (orchestrator-service)

**端口**: 8083
**职责**:
- 工作流编排
- 策略管理
- 自动化任务执行
- 智能推理集成

**主要 API**:
```
GET    /api/v1/workflows            - 工作流列表
POST   /api/v1/workflows            - 创建工作流
GET    /api/v1/strategies           - 策略列表
POST   /api/v1/strategies           - 创建策略
```

**技术栈**:
- Gin (Web 框架)
- NATS (消息订阅)
- PostgreSQL (工作流存储)
- Redis (状态缓存)

---

### 6. API 网关 (gateway-service)

**端口**: 8080 (对外)
**职责**:
- 统一入口
- 请求路由转发
- JWT 认证验证
- 限流保护
- CORS 处理
- 服务健康检查

**路由规则**:
```
/api/v1/auth/*        → auth-service:8090
/api/v1/users/*       → auth-service:8090
/api/v1/roles/*       → auth-service:8090
/api/v1/metrics/*     → monitor-service:8081
/api/v1/alerts/*      → monitor-service:8081
/api/v1/dashboard/*   → monitor-service:8081
/api/v1/clusters/*    → cluster-service:8082
/api/v1/agents/*      → agent-manager:8080
/api/v1/events/*      → agent-manager:8080
/api/v1/commands/*    → agent-manager:8080
/api/v1/workflows/*   → orchestrator-service:8083
```

**技术栈**:
- Gin (Web 框架)
- 反向代理
- Redis (限流)

---

## 前端应用

### agent-manager-ui

**端口**: 3000 (开发)
**技术栈**:
- Vue 3 (渐进式框架)
- Ant Design Vue 4 (UI 组件)
- VXETable (表格组件)
- Pinia (状态管理)
- Axios (HTTP 客户端)
- Vue Router (路由)

**主要功能模块**:
1. **登录认证** - 用户登录、Token 管理
2. **仪表盘** - 系统概览、实时监控
3. **Agent 管理** - Agent 列表、详情、操作
4. **事件监控** - 事件流、过滤、详情
5. **命令执行** - 命令创建、执行、结果查看
6. **集群管理** - 集群列表、健康检查、资源管理
7. **告警规则** - 规则配置、告警历史
8. **用户管理** - 用户、角色、权限管理

---

## 数据存储

### PostgreSQL 数据库

**用途**: 主数据存储
**数据库列表**:
- `k8s_agent_auth` - 认证服务数据库
  - users (用户)
  - roles (角色)
  - permissions (权限)
  - user_roles (用户角色关联)
  - role_permissions (角色权限关联)
  - api_keys (API密钥)

- `monitor_db` - 监控服务数据库
  - metrics_summary (监控概览)
  - agent_metrics (Agent 指标)
  - event_metrics (事件指标)
  - alerts (告警规则)
  - alert_history (告警历史)
  - trend_data (趋势数据)

- `cluster_db` - 集群服务数据库
  - clusters (集群信息)
  - cluster_health (集群健康)

- `agent_manager_db` - Agent 管理数据库
  - agents (Agent 信息)
  - events (事件)
  - commands (命令)

### Redis 缓存

**用途**: 缓存、限流、实时数据
**键空间**:
- `token:{user_id}` - JWT Token
- `metrics:summary` - 监控概览缓存
- `rate_limit:{ip}` - 限流计数
- `agent:status:{agent_id}` - Agent 状态

### NATS 消息队列

**用途**: 异步消息、事件驱动
**主题**:
- `events.pod.*` - Pod 事件
- `events.node.*` - Node 事件
- `commands.execute` - 命令执行
- `alerts.trigger` - 告警触发

---

## 服务间通信

### 认证流程

```
1. 用户登录
   前端 → API网关 → auth-service
   ↓
   auth-service 验证用户名密码
   ↓
   生成 JWT Token
   ↓
   返回 Token 和用户信息

2. 访问受保护资源
   前端 (带 Token) → API网关
   ↓
   网关验证 Token (可选调用 auth-service)
   ↓
   转发到后端服务
   ↓
   后端服务处理请求
   ↓
   返回响应
```

### 数据流向

```
1. Agent 事件上报
   Collect Agent → NATS → Agent Manager
   ↓
   Event Processor
   ↓
   PostgreSQL 存储
   ↓
   通知 Monitor Service (可选)

2. 监控数据采集
   Monitor Service (定时任务)
   ↓
   从 Agent Manager 获取 Agent 状态
   从 Cluster Service 获取集群状态
   ↓
   聚合计算
   ↓
   存储到 PostgreSQL
   缓存到 Redis

3. 告警触发
   Monitor Service (告警检查)
   ↓
   判断告警条件
   ↓
   触发告警
   ↓
   发送通知 (邮件/Webhook/Slack)
```

---

## 部署架构

### 开发环境

```bash
# 1. 启动基础设施
docker-compose up -d postgres redis nats

# 2. 启动后端服务
cd auth-service && make run &
cd monitor-service && make run &
cd cluster-service && make run &
cd agent-manager && make run &
cd orchestrator-service && make run &
cd gateway-service && make run &

# 3. 启动前端
cd agent-manager-ui && npm run dev
```

### 生产环境 (Kubernetes)

```yaml
# 部署清单示例
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gateway-service
spec:
  replicas: 2
  template:
    spec:
      containers:
      - name: gateway
        image: k8s-agent/gateway-service:latest
        ports:
        - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: gateway-service
spec:
  type: LoadBalancer
  ports:
  - port: 80
    targetPort: 8080
  selector:
    app: gateway-service
```

---

## 安全考虑

### 1. 认证安全
- JWT Token 有效期限制 (24小时)
- Token 刷新机制
- 密码 bcrypt 加密存储
- API Key 支持

### 2. 授权安全
- RBAC 权限模型
- 细粒度权限控制
- API 级别权限校验

### 3. 通信安全
- HTTPS/TLS 加密 (生产环境)
- API 网关统一入口
- CORS 跨域限制
- 限流保护

### 4. 数据安全
- 敏感数据加密
- SQL 注入防护
- XSS 攻击防护

---

## 监控与日志

### 应用监控
- Prometheus 指标暴露
- 健康检查端点
- 请求统计

### 日志管理
- 结构化日志 (JSON)
- 日志级别控制
- 日志聚合 (可选 ELK)

---

## 扩展性

### 水平扩展
- 所有微服务支持多实例部署
- 无状态设计
- 负载均衡

### 功能扩展
- 插件化架构
- 新服务接入简单
- API 版本控制

---

## 快速开始

### 前置要求
- Go 1.21+
- Node.js 16+
- PostgreSQL 14+
- Redis 6+
- NATS 2.9+

### 完整启动流程

详见各服务的 README:
- [auth-service/README.md](auth-service/README.md)
- [monitor-service/README.md](monitor-service/README.md)
- [cluster-service/README.md](cluster-service/README.md)
- [agent-manager/README.md](agent-manager/README.md)
- [gateway-service/README.md](gateway-service/README.md)
- [agent-manager-ui/README.md](agent-manager-ui/README.md)

---

## 技术栈总结

**后端**:
- 语言: Go 1.21
- Web 框架: Gin
- 数据库: PostgreSQL
- 缓存: Redis
- 消息队列: NATS
- K8s 客户端: client-go
- JWT: golang-jwt/jwt
- 日志: logrus

**前端**:
- 框架: Vue 3
- UI 库: Ant Design Vue 4
- 表格: VXETable
- 状态管理: Pinia
- HTTP: Axios
- 构建工具: Vite

**基础设施**:
- 容器: Docker
- 编排: Kubernetes
- 监控: Prometheus (可选)
- 网关: Traefik (可选)

---

## 维护者

K8s Agent 开发团队

## License

MIT
