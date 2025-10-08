# 后台管理系统快速启动指南

## 概述

本指南帮助您快速启动完整的 K8s Agent 后台管理系统,包括所有微服务和前端应用。

## 架构组件

### 微服务
1. **auth-service** (8090) - 认证授权服务
2. **monitor-service** (8081) - 监控管理服务
3. **cluster-service** (8082) - K8s集群管理服务
4. **agent-manager** (8080) - Agent管理服务
5. **orchestrator-service** (8083) - 编排服务
6. **gateway-service** (8080) - API网关

### 前端
- **agent-manager-ui** (3000) - 管理界面

### 基础设施
- PostgreSQL (5432) - 主数据库
- Redis (6379) - 缓存
- NATS (4222) - 消息队列

## 方式一: Docker Compose (推荐)

### 1. 启动所有服务

```bash
# 克隆或进入项目目录
cd k8s-agent

# 启动完整的后台管理系统
docker-compose -f docker-compose.backend.yml up -d

# 查看服务状态
docker-compose -f docker-compose.backend.yml ps

# 查看日志
docker-compose -f docker-compose.backend.yml logs -f
```

### 2. 访问系统

- **前端界面**: http://localhost:3000
- **API网关**: http://localhost:8080
- **认证服务**: http://localhost:8090
- **监控服务**: http://localhost:8081
- **集群服务**: http://localhost:8082

### 3. 默认账号

```
用户名: admin
密码: admin123
```

### 4. 停止服务

```bash
# 停止所有服务
docker-compose -f docker-compose.backend.yml down

# 停止并删除数据卷
docker-compose -f docker-compose.backend.yml down -v
```

## 方式二: 本地开发模式

### 前置要求

- Go 1.21+
- Node.js 16+
- PostgreSQL 14+
- Redis 6+
- NATS 2.9+

### 1. 启动基础设施

```bash
# 方式 A: 使用 Docker
docker run -d --name postgres -p 5432:5432 \
  -e POSTGRES_PASSWORD=postgres \
  postgres:14-alpine

docker run -d --name redis -p 6379:6379 \
  redis:7-alpine

docker run -d --name nats -p 4222:4222 -p 8222:8222 \
  nats:2.9-alpine --http_port 8222

# 方式 B: 本地安装
# macOS
brew install postgresql@14 redis nats-server
brew services start postgresql@14
brew services start redis
brew services start nats-server
```

### 2. 初始化数据库

```bash
# 连接到 PostgreSQL
psql -U postgres -h localhost

# 执行初始化脚本
\i scripts/init-databases.sql

# 退出
\q
```

### 3. 启动后端服务

```bash
# Terminal 1: 认证服务
cd auth-service
go mod tidy
make run

# Terminal 2: 监控服务
cd monitor-service
go mod tidy
make run

# Terminal 3: 集群服务
cd cluster-service
go mod tidy
make run

# Terminal 4: Agent 管理服务
cd agent-manager
go mod tidy
make run

# Terminal 5: 编排服务
cd orchestrator-service
go mod tidy
make run

# Terminal 6: API 网关
cd gateway-service
go mod tidy
make run
```

### 4. 启动前端

```bash
# Terminal 7: 前端应用
cd agent-manager-ui
npm install
npm run dev
```

### 5. 访问系统

打开浏览器访问: http://localhost:3000

## 验证服务状态

### 检查基础设施

```bash
# PostgreSQL
psql -U postgres -h localhost -c "\l"

# Redis
redis-cli ping

# NATS
curl http://localhost:8222/healthz
```

### 检查微服务健康

```bash
# 认证服务
curl http://localhost:8090/health

# 监控服务
curl http://localhost:8081/health

# 集群服务
curl http://localhost:8082/health

# API 网关
curl http://localhost:8080/health
```

### 检查网关路由

```bash
# 通过网关访问各服务
curl http://localhost:8080/api/v1/health/services
```

## 测试完整流程

### 1. 登录系统

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'
```

保存返回的 token。

### 2. 获取用户信息

```bash
curl http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer <your-token>"
```

### 3. 获取监控概览

```bash
curl http://localhost:8080/api/v1/metrics/summary \
  -H "Authorization: Bearer <your-token>"
```

### 4. 获取集群列表

```bash
curl http://localhost:8080/api/v1/clusters \
  -H "Authorization: Bearer <your-token>"
```

## 常见问题

### 1. 端口冲突

如果端口被占用,可以修改各服务的配置文件:

```yaml
# auth-service/configs/config.yaml
server:
  port: 8090  # 改为其他端口

# 类似修改其他服务
```

### 2. 数据库连接失败

检查 PostgreSQL 是否启动:

```bash
# 查看进程
ps aux | grep postgres

# 检查端口
lsof -i :5432

# 测试连接
psql -U postgres -h localhost
```

### 3. Redis 连接失败

```bash
# 检查 Redis
redis-cli ping

# 查看配置
redis-cli config get port
```

### 4. NATS 连接失败

```bash
# 检查 NATS
curl http://localhost:8222/varz

# 查看订阅
curl http://localhost:8222/subsz
```

### 5. 前端无法访问后端

检查 API 代理配置:

```javascript
// agent-manager-ui/vite.config.js
server: {
  proxy: {
    '/api': {
      target: 'http://localhost:8080',  // 网关地址
      changeOrigin: true
    }
  }
}
```

### 6. 认证失败

确保:
1. auth-service 正常运行
2. JWT secret 配置一致
3. Token 未过期

## 开发建议

### 1. 服务启动顺序

建议按以下顺序启动:
1. 基础设施 (PostgreSQL, Redis, NATS)
2. auth-service (认证服务)
3. monitor-service (监控服务)
4. cluster-service (集群服务)
5. agent-manager (Agent 管理)
6. orchestrator-service (编排服务)
7. gateway-service (API 网关)
8. agent-manager-ui (前端)

### 2. 日志查看

```bash
# 查看服务日志
tail -f logs/auth-service.log
tail -f logs/monitor-service.log
tail -f logs/gateway-service.log

# Docker 日志
docker-compose -f docker-compose.backend.yml logs -f auth-service
```

### 3. 热重载开发

使用 Air 实现 Go 服务热重载:

```bash
# 安装 Air
go install github.com/cosmtrek/air@latest

# 在服务目录运行
cd auth-service
air
```

### 4. 数据库迁移

```bash
# 使用 golang-migrate
migrate -path migrations -database "postgresql://postgres:postgres@localhost:5432/k8s_agent_auth?sslmode=disable" up
```

## 生产部署

详见:
- [Kubernetes 部署指南](deployments/kubernetes/README.md)
- [Docker Swarm 部署指南](deployments/swarm/README.md)
- [架构文档](BACKEND_MANAGEMENT_ARCHITECTURE.md)

## 监控和维护

### 查看 Prometheus 指标

```bash
# 监控服务指标
curl http://localhost:9091/metrics
```

### 数据库备份

```bash
# 备份所有数据库
pg_dumpall -U postgres > backup.sql

# 恢复
psql -U postgres < backup.sql
```

## 下一步

- 配置告警规则
- 添加 K8s 集群
- 创建用户和角色
- 配置监控仪表盘
- 设置工作流

## 获取帮助

- 查看各服务的 README 文档
- 查看 API 文档
- 提交 Issue

## License

MIT
