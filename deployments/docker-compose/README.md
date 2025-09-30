# Aetherius Docker Compose 部署

使用 Docker Compose 快速部署 Aetherius 智能运维平台的所有组件。

---

## 前置要求

- Docker 20.10+
- Docker Compose 2.0+
- 至少 8GB 可用内存
- 至少 20GB 可用磁盘空间

---

## 快速开始

### 1. 启动所有服务

```bash
cd deployments/docker-compose
docker-compose up -d
```

### 2. 检查服务状态

```bash
docker-compose ps
```

所有服务应显示为 `healthy` 状态:

```text
NAME                        STATUS
aetherius-postgres          Up (healthy)
aetherius-redis             Up (healthy)
aetherius-nats              Up (healthy)
aetherius-neo4j             Up (healthy)
aetherius-agent-manager     Up (healthy)
aetherius-orchestrator      Up (healthy)
aetherius-reasoning         Up (healthy)
```

### 3. 验证服务

```bash
# Agent Manager
curl http://localhost:8080/health

# Orchestrator Service
curl http://localhost:8081/health

# Reasoning Service
curl http://localhost:8082/health
```

---

## 服务端口

| 服务 | 端口 | 用途 |
|------|------|------|
| PostgreSQL | 5432 | 数据库 |
| Redis | 6379 | 缓存 |
| NATS | 4222 | 消息队列 (客户端) |
| NATS | 8222 | 监控 |
| Neo4j | 7474 | HTTP API |
| Neo4j | 7687 | Bolt 协议 |
| Agent Manager | 8080 | REST API |
| Orchestrator | 8081 | REST API |
| Reasoning Service | 8082 | REST API |

---

## 访问 Web 界面

### Neo4j Browser

访问 [http://localhost:7474](http://localhost:7474)

- 用户名: `neo4j`
- 密码: `neo4j_pass`

### NATS Monitoring

访问 [http://localhost:8222](http://localhost:8222)

---

## 部署 Collect Agent

在目标 Kubernetes 集群中部署 collect-agent:

```bash
# 编辑配置
vi ../../collect-agent/configs/config.yaml

# 修改 NATS 连接地址
nats:
  url: "nats://<YOUR_HOST_IP>:4222"  # Docker host IP

# 构建镜像
cd ../../collect-agent
docker build -t collect-agent:latest .

# 部署到 K8s
kubectl apply -f deployments/k8s/collect-agent.yaml
```

---

## 常用命令

### 查看日志

```bash
# 所有服务
docker-compose logs -f

# 特定服务
docker-compose logs -f agent-manager
docker-compose logs -f orchestrator-service
docker-compose logs -f reasoning-service
```

### 重启服务

```bash
# 重启所有服务
docker-compose restart

# 重启特定服务
docker-compose restart agent-manager
```

### 停止服务

```bash
# 停止所有服务
docker-compose down

# 停止并删除数据卷
docker-compose down -v
```

### 扩展服务

```bash
# 扩展 agent-manager 到 3 个实例
docker-compose up -d --scale agent-manager=3
```

### 更新服务

```bash
# 重新构建并启动
docker-compose up -d --build agent-manager
```

---

## 数据持久化

数据存储在 Docker volumes 中:

```bash
# 查看 volumes
docker volume ls | grep aetherius

# 备份 PostgreSQL
docker exec aetherius-postgres pg_dump -U aetherius aetherius > backup.sql

# 恢复 PostgreSQL
docker exec -i aetherius-postgres psql -U aetherius aetherius < backup.sql

# 备份 Neo4j
docker exec aetherius-neo4j neo4j-admin dump --database=neo4j --to=/tmp/neo4j-backup.dump
docker cp aetherius-neo4j:/tmp/neo4j-backup.dump ./
```

---

## 配置自定义

### 修改数据库密码

编辑 `docker-compose.yml`:

```yaml
postgres:
  environment:
    POSTGRES_PASSWORD: your_new_password

redis:
  command: redis-server --requirepass your_new_password

neo4j:
  environment:
    NEO4J_AUTH: neo4j/your_new_password
```

### 修改资源限制

```yaml
agent-manager:
  deploy:
    resources:
      limits:
        cpus: '2'
        memory: 4G
      reservations:
        cpus: '0.5'
        memory: 1G
```

### 启用持久化日志

```yaml
agent-manager:
  volumes:
    - ./logs/agent-manager:/app/logs
```

---

## 监控和调试

### 查看 NATS 统计

```bash
curl http://localhost:8222/varz
```

### 查看 Redis 信息

```bash
docker exec aetherius-redis redis-cli -a redis_pass INFO
```

### 查看 PostgreSQL 连接

```bash
docker exec aetherius-postgres psql -U aetherius -c "SELECT * FROM pg_stat_activity;"
```

### 查看 Neo4j 数据库状态

```bash
docker exec aetherius-neo4j cypher-shell -u neo4j -p neo4j_pass "CALL dbms.components();"
```

---

## 故障排查

### 问题 1: 服务无法启动

**检查**:

```bash
# 查看详细日志
docker-compose logs <service_name>

# 检查端口占用
sudo netstat -tulpn | grep <port>
```

**解决**: 确保端口未被占用,修改 `docker-compose.yml` 中的端口映射

### 问题 2: 数据库连接失败

**检查**:

```bash
# 测试数据库连接
docker exec aetherius-agent-manager nc -zv postgres 5432
```

**解决**: 等待数据库服务完全启动,健康检查通过后再启动应用服务

### 问题 3: NATS 连接失败

**检查**:

```bash
# 查看 NATS 日志
docker-compose logs nats

# 测试连接
docker exec aetherius-agent-manager nc -zv nats 4222
```

### 问题 4: 内存不足

**解决**:

```bash
# 减少服务副本数
docker-compose up -d --scale reasoning-service=1

# 或增加 Docker 内存限制
# Docker Desktop -> Settings -> Resources -> Memory
```

---

## 性能优化

### 1. PostgreSQL 优化

编辑 `docker-compose.yml`:

```yaml
postgres:
  command: postgres -c max_connections=200 -c shared_buffers=512MB
```

### 2. Redis 持久化

```yaml
redis:
  command: redis-server --requirepass redis_pass --appendonly yes
```

### 3. NATS JetStream

```yaml
nats:
  command: ["-js", "-m", "8222", "-sd", "/data"]
  volumes:
    - nats_data:/data
```

---

## 生产环境注意事项

### 安全加固

1. **修改默认密码**: 所有默认密码应该修改
2. **启用 TLS**: NATS 和 API 应启用 TLS
3. **网络隔离**: 使用 Docker 网络隔离
4. **限制访问**: 使用防火墙规则限制端口访问

### 高可用

1. **PostgreSQL**: 使用主从复制
2. **Redis**: 使用 Sentinel 或 Cluster
3. **NATS**: 使用 Cluster 模式
4. **应用服务**: 扩展到多个实例

### 监控

1. 集成 Prometheus 采集指标
2. 配置 Grafana 仪表板
3. 设置告警规则
4. 配置日志聚合 (ELK/Loki)

---

## 清理

### 停止并删除所有容器和数据

```bash
docker-compose down -v
docker volume prune -f
```

### 删除镜像

```bash
docker rmi $(docker images 'aetherius-*' -q)
```

---

## 下一步

- [部署 Collect Agent](../../collect-agent/README.md#部署)
- [配置工作流](../../orchestrator-service/README.md#工作流定义)
- [使用 API](../docs/API.md)
- [生产部署](../docs/deployment/PRODUCTION.md)