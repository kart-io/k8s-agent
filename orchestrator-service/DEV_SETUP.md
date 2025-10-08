# Orchestrator Service - 开发环境设置指南

快速设置 orchestrator-service 本地开发环境。

---

## 🚀 快速开始

### 一键设置（推荐）

```bash
# 运行设置脚本
./scripts/setup-dev.sh

# 等待服务启动完成，然后运行 orchestrator-service
make run
```

### 手动设置

如果自动脚本失败，可以手动启动服务：

#### 1. 启动 PostgreSQL

```bash
docker run -d \
  --name aetherius-postgres-dev \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=dev-postgres-password \
  -e POSTGRES_DB=aetherius_orchestrator \
  -p 5432:5432 \
  postgres:14-alpine

# 等待启动
sleep 10

# 验证
docker exec aetherius-postgres-dev pg_isready -U postgres
```

#### 2. 启动 Redis

```bash
docker run -d \
  --name aetherius-redis-dev \
  -p 6379:6379 \
  redis:7-alpine \
  redis-server --requirepass dev-redis-password

# 验证
docker exec aetherius-redis-dev redis-cli -a dev-redis-password ping
```

#### 3. 启动 NATS

```bash
docker run -d \
  --name aetherius-nats-dev \
  -p 4222:4222 \
  -p 8222:8222 \
  nats:2.10-alpine \
  -js -m 8222

# 验证
curl http://localhost:8222/healthz
```

#### 4. 运行 Orchestrator Service

```bash
make run
# 或
go run ./cmd/server --config=configs/config.yaml
```

---

## 📋 服务配置

### 当前配置 (`configs/config.yaml`)

```yaml
# 数据库
database:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "dev-postgres-password"
  database: "aetherius_orchestrator"

# Redis
redis:
  addr: "localhost:6379"
  password: "dev-redis-password"
  db: 1

# NATS
nats:
  url: "nats://localhost:4222"
```

### 连接信息

| 服务 | 地址 | 认证 |
|------|------|------|
| **PostgreSQL** | `localhost:5432` | User: `postgres`<br>Pass: `dev-postgres-password` |
| **Redis** | `localhost:6379` | Pass: `dev-redis-password` |
| **NATS** | `localhost:4222` | 无需认证 |
| **NATS Monitor** | `http://localhost:8222` | Web界面 |

---

## 🔍 验证服务

### 检查容器状态

```bash
docker ps --filter "name=aetherius-"
```

**期望输出**:
```
CONTAINER ID   IMAGE                 STATUS          PORTS
xxxx           postgres:14-alpine    Up 2 minutes    0.0.0.0:5432->5432/tcp
xxxx           redis:7-alpine        Up 2 minutes    0.0.0.0:6379->6379/tcp
xxxx           nats:2.10-alpine      Up 2 minutes    0.0.0.0:4222->4222/tcp, 0.0.0.0:8222->8222/tcp
```

### 测试数据库连接

```bash
# 使用 Docker exec
docker exec -it aetherius-postgres-dev psql -U postgres -d aetherius_orchestrator -c "SELECT version();"

# 或使用 Go
go run ./cmd/server --config=configs/config.yaml
# 检查日志中是否有 "PostgreSQL connected successfully"
```

### 测试 Redis 连接

```bash
docker exec aetherius-redis-dev redis-cli -a dev-redis-password ping
# 应该返回: PONG
```

### 测试 NATS 连接

```bash
curl http://localhost:8222/healthz
# 应该返回: ok
```

---

## 🛠️ 常见问题

### Q1: `make run` 失败，提示数据库不存在

**错误**: `database "aetherius_orchestrator" does not exist`

**解决**:
```bash
# 1. 检查容器是否运行
docker ps | grep postgres

# 2. 如果没有运行，启动容器
docker start aetherius-postgres-dev

# 3. 如果容器不存在，运行设置脚本
./scripts/setup-dev.sh
```

### Q2: 端口冲突

**错误**: `bind: address already in use`

**解决**:
```bash
# 检查哪个进程占用端口（以 5432 为例）
lsof -i :5432

# 停止冲突的容器
docker ps | grep 5432
docker stop <container_id>

# 或使用不同的端口（需要修改 config.yaml）
```

### Q3: Docker 容器无法启动

**错误**: `Error response from daemon: driver failed`

**解决**:
```bash
# 1. 检查 Docker Desktop 是否运行
docker info

# 2. 清理旧容器
docker rm -f aetherius-postgres-dev aetherius-redis-dev aetherius-nats-dev

# 3. 重新运行设置脚本
./scripts/setup-dev.sh
```

### Q4: orchestrator-service 无法连接到 agent-manager

**错误**: `failed to connect to agent-manager`

**检查**:
```bash
# 1. agent-manager 是否运行
curl http://localhost:8080/health

# 2. NATS 是否运行
curl http://localhost:8222/healthz

# 3. 检查配置文件中的服务地址
cat configs/config.yaml | grep -A 5 "services:"
```

---

## 🧹 清理环境

### 停止服务（保留数据）

```bash
docker stop aetherius-postgres-dev aetherius-redis-dev aetherius-nats-dev
```

### 完全清理（删除容器和数据）

```bash
docker rm -f aetherius-postgres-dev aetherius-redis-dev aetherius-nats-dev

# 清理数据卷（可选）
docker volume prune
```

### 重新开始

```bash
./scripts/setup-dev.sh
```

---

## 📊 监控和调试

### 查看日志

```bash
# PostgreSQL 日志
docker logs aetherius-postgres-dev

# Redis 日志
docker logs aetherius-redis-dev

# NATS 日志
docker logs aetherius-nats-dev

# Orchestrator Service 日志
# 会直接输出到终端
```

### NATS 监控界面

访问 http://localhost:8222 查看：
- 连接状态
- 订阅主题
- 消息统计

### PostgreSQL 管理

```bash
# 进入数据库
docker exec -it aetherius-postgres-dev psql -U postgres -d aetherius_orchestrator

# 查看表
\dt

# 查看工作流
SELECT id, name, status FROM workflows;

# 查看执行历史
SELECT * FROM workflow_executions ORDER BY started_at DESC LIMIT 10;
```

---

## 🎯 完整测试流程

```bash
# 1. 设置环境
./scripts/setup-dev.sh

# 2. 启动 agent-manager（如果还未运行）
cd ../agent-manager
make run &

# 3. 启动 orchestrator-service
cd ../orchestrator-service
make run &

# 4. 启动 reasoning-service-go（可选，用于 AI 分析）
cd ../reasoning-service-go
export OPENAI_API_KEY="sk-..."
go run cmd/server/main.go &

# 5. 发送测试事件
curl -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -d '{
    "cluster_id": "test-cluster",
    "namespace": "default",
    "reason": "CrashLoopBackOff",
    "message": "Pod is crashing repeatedly",
    "severity": "critical",
    "labels": {
      "app": "test-app",
      "pod": "test-pod-123"
    }
  }'

# 6. 检查 orchestrator-service 日志
# 应该看到工作流被触发并执行
```

---

## 🔗 相关资源

- [Orchestrator Service README](README.md)
- [Agent Manager](../agent-manager/README.md)
- [Reasoning Service (Go)](../reasoning-service-go/README.md)
- [System Architecture](../docs/architecture/SYSTEM_ARCHITECTURE.md)

---

## 💡 开发技巧

### 使用 Make 命令

```bash
make build      # 构建
make run        # 运行
make test       # 测试
make clean      # 清理
make help       # 查看所有命令
```

### 热重载开发

```bash
# 安装 air（热重载工具）
go install github.com/cosmtrek/air@latest

# 使用 air 运行（代码改动自动重启）
air
```

### 调试模式

```bash
# 修改配置启用调试日志
# configs/config.yaml
logging:
  level: "debug"

# 然后运行
make run
```

---

**开始愉快的开发吧！** 🚀

如有问题，请查看 [故障排查指南](#-常见问题) 或提交 Issue。
