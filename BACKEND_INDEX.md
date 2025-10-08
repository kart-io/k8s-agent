# 后台管理系统文档索引

> 快速查找后台管理系统相关文档和资源

## 📚 核心文档

| 文档 | 说明 | 适合人群 |
|------|------|----------|
| [实施总结](BACKEND_IMPLEMENTATION_SUMMARY.md) | 📋 已完成工作总览 | 所有人 |
| [快速启动](QUICK_START_BACKEND.md) | 🚀 快速上手指南 | 新手必读 |
| [架构设计](BACKEND_MANAGEMENT_ARCHITECTURE.md) | 🏗️ 系统架构详解 | 架构师/开发者 |
| [服务说明](BACKEND_SERVICES.md) | 🎯 各服务功能说明 | 开发者/运维 |

---

## 🔧 服务文档

### 后端微服务

| 服务 | 端口 | README | 状态 |
|------|------|--------|------|
| auth-service | 8090 | [查看](auth-service/README.md) | ✅ 完善 |
| monitor-service | 8081 | [查看](monitor-service/README.md) | ✅ 新建 |
| cluster-service | 8082 | [查看](cluster-service/README.md) | ✅ 新建 |
| gateway-service | 8080 | [查看](gateway-service/README.md) | ✅ 已有 |

### 前端应用

| 应用 | 端口 | README | 状态 |
|------|------|--------|------|
| agent-manager-ui | 3000 | [查看](agent-manager-ui/README.md) | ✅ 扩展 |

---

## 🚀 快速开始

### 一键启动 (Docker Compose)

```bash
docker-compose -f docker-compose.backend.yml up -d
open http://localhost:3000
# 账号: admin / admin123
```

### 本地开发

```bash
# 1. 基础设施
docker run -d --name postgres -p 5432:5432 -e POSTGRES_PASSWORD=postgres postgres:14-alpine
docker run -d --name redis -p 6379:6379 redis:7-alpine

# 2. 初始化数据库
psql -U postgres -h localhost -f scripts/init-databases.sql

# 3. 启动服务 (在不同 Terminal)
cd auth-service && make run
cd monitor-service && make run
cd cluster-service && make run
cd gateway-service && make run
cd agent-manager-ui && npm run dev
```

详细说明: [QUICK_START_BACKEND.md](QUICK_START_BACKEND.md)

---

## 📁 目录结构

```
k8s-agent/
├── auth-service/              # ✅ 认证授权服务
│   ├── cmd/
│   ├── configs/
│   ├── internal/
│   ├── pkg/
│   └── README.md
│
├── monitor-service/           # ✅ 监控管理服务 (新建)
│   ├── cmd/
│   ├── configs/
│   ├── internal/
│   ├── pkg/
│   └── README.md
│
├── cluster-service/           # ✅ 集群管理服务 (新建)
│   ├── cmd/
│   ├── configs/
│   ├── internal/
│   ├── pkg/
│   └── README.md
│
├── gateway-service/           # ✅ API 网关
│   ├── cmd/
│   ├── configs/
│   ├── internal/
│   └── README.md
│
├── agent-manager-ui/          # ✅ 前端应用 (扩展)
│   ├── src/
│   │   ├── api/
│   │   │   ├── auth.js        # 新增
│   │   │   └── user.js        # 新增
│   │   ├── store/
│   │   │   └── user.js        # 新增
│   │   ├── views/
│   │   │   └── Login.vue      # 新增
│   │   └── directives/
│   │       └── permission.js  # 新增
│   └── README.md
│
├── scripts/
│   └── init-databases.sql     # ✅ 数据库初始化 (新建)
│
├── docker-compose.backend.yml # ✅ Docker Compose 配置 (新建)
│
└── 📚 文档
    ├── BACKEND_IMPLEMENTATION_SUMMARY.md  # ✅ 实施总结
    ├── QUICK_START_BACKEND.md             # ✅ 快速启动
    ├── BACKEND_MANAGEMENT_ARCHITECTURE.md # ✅ 架构设计
    ├── BACKEND_SERVICES.md                # ✅ 服务说明
    └── BACKEND_INDEX.md                   # ✅ 文档索引 (本文档)
```

---

## 🎯 功能概览

### 认证与授权
- ✅ JWT Token 认证
- ✅ 用户登录/登出
- ✅ RBAC 权限模型
- ✅ 用户/角色/权限管理
- ✅ 前端权限控制

### 监控管理
- ✅ 实时监控数据采集
- ✅ 指标聚合与统计
- ✅ 告警规则管理
- ✅ 多渠道告警通知
- ✅ 仪表盘数据

### 集群管理
- ✅ 多 K8s 集群管理
- ✅ 集群健康检查
- ✅ Pod/Deployment 管理
- ✅ 资源统计
- ✅ 集群事件监控

### API 网关
- ✅ 统一入口
- ✅ 路由转发
- ✅ 认证验证
- ✅ 限流保护
- ✅ CORS 处理

---

## 🌐 API 文档

### 通过网关访问

所有 API 通过网关统一访问: `http://localhost:8080/api/v1/...`

### 主要端点

**认证服务**:
```
POST   /api/v1/auth/login       # 用户登录
GET    /api/v1/auth/me          # 获取用户信息
GET    /api/v1/auth/menus       # 获取用户菜单
GET    /api/v1/users            # 用户列表
GET    /api/v1/roles            # 角色列表
GET    /api/v1/permissions      # 权限列表
```

**监控服务**:
```
GET    /api/v1/metrics/summary  # 监控概览
GET    /api/v1/metrics/agents   # Agent 指标
GET    /api/v1/metrics/trends   # 趋势数据
GET    /api/v1/alerts           # 告警规则列表
POST   /api/v1/alerts           # 创建告警规则
GET    /api/v1/dashboard/overview  # 仪表盘概览
```

**集群服务**:
```
GET    /api/v1/clusters                      # 集群列表
POST   /api/v1/clusters                      # 添加集群
GET    /api/v1/clusters/:id/health           # 集群健康
GET    /api/v1/clusters/:id/nodes            # 节点列表
GET    /api/v1/clusters/:id/namespaces/:ns/pods  # Pod 列表
POST   /api/v1/clusters/:id/namespaces/:ns/deployments  # 创建 Deployment
```

---

## 🔐 默认账号

```
用户名: admin
密码: admin123
```

首次登录后请及时修改密码!

---

## 🛠️ 开发工具

### 构建命令

```bash
# 各服务通用命令
make build      # 编译
make run        # 运行
make test       # 测试
make clean      # 清理
make docker-build  # Docker 构建
```

### 数据库工具

```bash
# 连接 PostgreSQL
psql -U postgres -h localhost

# 查看所有数据库
\l

# 切换数据库
\c k8s_agent_auth

# 查看表
\dt
```

### Redis 工具

```bash
# 连接 Redis
redis-cli

# 查看所有键
keys *

# 获取值
get token:user_id
```

---

## 📊 监控与运维

### 健康检查

```bash
# 检查所有服务
curl http://localhost:8090/health  # auth-service
curl http://localhost:8081/health  # monitor-service
curl http://localhost:8082/health  # cluster-service
curl http://localhost:8080/health  # gateway-service

# 通过网关检查所有服务
curl http://localhost:8080/api/v1/health/services
```

### Prometheus 指标

```bash
# 监控服务指标
curl http://localhost:9091/metrics
```

### 日志查看

```bash
# Docker Compose 日志
docker-compose -f docker-compose.backend.yml logs -f

# 单个服务日志
docker-compose -f docker-compose.backend.yml logs -f auth-service
```

---

## 🐛 故障排查

### 常见问题

| 问题 | 解决方案 | 文档 |
|------|----------|------|
| 端口冲突 | 修改配置文件中的端口 | [QUICK_START_BACKEND.md](QUICK_START_BACKEND.md#常见问题) |
| 数据库连接失败 | 检查 PostgreSQL 是否启动 | [QUICK_START_BACKEND.md](QUICK_START_BACKEND.md#常见问题) |
| Redis 连接失败 | 检查 Redis 是否启动 | [QUICK_START_BACKEND.md](QUICK_START_BACKEND.md#常见问题) |
| 认证失败 | 检查 JWT secret 配置 | [QUICK_START_BACKEND.md](QUICK_START_BACKEND.md#常见问题) |
| 前端无法访问后端 | 检查网关和代理配置 | [QUICK_START_BACKEND.md](QUICK_START_BACKEND.md#常见问题) |

---

## 📖 相关资源

### 官方文档
- [Vue 3](https://vuejs.org/)
- [Ant Design Vue](https://antdv.com/)
- [Gin Framework](https://gin-gonic.com/)
- [client-go](https://github.com/kubernetes/client-go)

### 技术博客
- [微服务架构设计](https://microservices.io/)
- [JWT 认证](https://jwt.io/)
- [RBAC 权限模型](https://en.wikipedia.org/wiki/Role-based_access_control)

---

## 🎓 学习路径

### 新手入门
1. 阅读 [快速启动指南](QUICK_START_BACKEND.md)
2. 启动系统,体验功能
3. 查看 [服务说明](BACKEND_SERVICES.md)

### 开发者
1. 阅读 [架构设计文档](BACKEND_MANAGEMENT_ARCHITECTURE.md)
2. 查看各服务源码
3. 了解 API 设计

### 运维人员
1. 部署系统
2. 配置监控
3. 熟悉故障排查

---

## 🤝 贡献

欢迎提交 Issue 和 Pull Request!

---

## 📝 更新日志

### 2025-10-07
- ✅ 创建 monitor-service
- ✅ 创建 cluster-service
- ✅ 完善 auth-service
- ✅ 扩展 agent-manager-ui
- ✅ 编写完整文档
- ✅ 提供 Docker Compose 配置

---

## 📄 License

MIT

---

**快速导航**:
- [返回主 README](README.md)
- [实施总结](BACKEND_IMPLEMENTATION_SUMMARY.md)
- [快速启动](QUICK_START_BACKEND.md)
- [架构设计](BACKEND_MANAGEMENT_ARCHITECTURE.md)
