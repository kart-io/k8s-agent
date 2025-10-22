# Kustomize 部署 - Neo4j & NATS 支持

本目录已更新，添加了 Neo4j 图数据库和 NATS 消息队列的完整部署支持。

## 新增功能

### 基础设施服务

现在支持完整的基础设施栈部署：

- **Neo4j** - 图数据库（用于知识图谱和推理服务）
- **NATS** - 消息队列（用于服务间通信）
- **Redis** - 缓存服务
- **PostgreSQL** - 关系数据库

### 新增命令

#### 快速部署

```bash
# 部署所有服务（基础设施 + 监控 + 网关）
make deploy-all

# 仅部署基础设施服务
make deploy-infra

# 仅部署监控服务
make deploy-monitoring
```

#### 单独部署基础设施

```bash
# 部署 Neo4j 图数据库
make deploy-neo4j

# 部署 NATS 消息队列
make deploy-nats

# 部署 Redis 缓存
make deploy-redis

# 部署 PostgreSQL 数据库
make deploy-postgres
```

#### 查看状态

```bash
# 查看 Neo4j 状态
make status-neo4j

# 查看 NATS 状态
make status-nats

# 查看所有服务状态
make status
```

#### 查看日志

```bash
# 查看 Neo4j 日志
make logs-neo4j

# 查看 NATS 日志
make logs-nats
```

#### 端口转发

```bash
# 转发 Neo4j 端口（HTTP: 7474, Bolt: 7687）
make pf-neo4j

# 转发 NATS 端口（4222）
make pf-nats

# 转发 Redis 端口（6379）
make pf-redis

# 转发 PostgreSQL 端口（5432）
make pf-postgres
```

#### 删除服务

```bash
# 删除所有基础设施服务
make delete-infra

# 单独删除服务
make delete-neo4j
make delete-nats
make delete-redis
make delete-postgres
```

## 目录结构

```
kustomize/
├── base/              # 基础配置
│   ├── neo4j/        # Neo4j StatefulSet 配置
│   ├── nats/         # NATS StatefulSet 配置
│   ├── redis/        # Redis StatefulSet 配置
│   └── postgres/     # PostgreSQL StatefulSet 配置
├── neo4j/            # Neo4j 顶层部署配置
├── nats/             # NATS 顶层部署配置
├── redis/            # Redis 顶层部署配置
├── postgres/         # PostgreSQL 顶层部署配置
└── overlays/         # 环境特定覆盖
    ├── dev/          # 开发环境配置
    └── prod/         # 生产环境配置
```

## 服务访问

### Neo4j

- **Browser UI**: `http://localhost:7474`（通过 `make pf-neo4j` 转发）
- **Bolt 协议**: `bolt://localhost:7687`
- **默认认证**: `neo4j/neo4j`（首次登录需修改密码）

### NATS

- **客户端端口**: `localhost:4222`（通过 `make pf-nats` 转发）
- **监控端口**: `8222`

### Redis

- **端口**: `localhost:6379`（通过 `make pf-redis` 转发）
- **连接命令**: `redis-cli -h localhost -p 6379`

### PostgreSQL

- **端口**: `localhost:5432`（通过 `make pf-postgres` 转发）
- **连接命令**: `psql -h localhost -p 5432 -U postgres`

## 验证配置

```bash
# 验证所有 Kustomize 配置
make validate
```

输出示例：
```
✅ 验证 Neo4j 配置...
  Neo4j: OK
✅ 验证 NATS 配置...
  NATS: OK
✅ 验证 Redis 配置...
  Redis: OK
✅ 验证 PostgreSQL 配置...
  PostgreSQL: OK
✅ 所有配置验证通过
```

## 完整命令列表

运行 `make help` 查看所有可用命令的完整列表。

## 部署顺序建议

1. **基础设施优先**
   ```bash
   make deploy-infra
   ```

2. **监控服务**
   ```bash
   make deploy-monitoring
   ```

3. **网关服务**
   ```bash
   make deploy-gateway
   ```

或者一次性部署所有服务：
```bash
make deploy-all
```

## 命名空间

- **基础设施服务**: `aetherius`
- **监控服务**: `monitoring`
- **网关服务**: `traefik`

## 注意事项

1. Neo4j 和 NATS 使用 StatefulSet 部署，提供稳定的网络标识和持久化存储
2. 所有服务默认部署在 `aetherius` 命名空间
3. 确保 Kubernetes 集群有足够的资源运行这些服务
4. 生产环境建议使用 `overlays/prod` 配置

## 相关文档

- [Neo4j 官方文档](https://neo4j.com/docs/)
- [NATS 官方文档](https://docs.nats.io/)
- [Kustomize 官方文档](https://kustomize.io/)
