# Kustomize 部署快速入门

## 前置条件

```bash
# 检查所需工具
make check-tools

# 确保已连接到 Kubernetes 集群
kubectl cluster-info
```

## 快速部署

### 部署开发环境

```bash
# 部署所有服务
make dev-all

# 或者逐个部署
make dev-postgres
make dev-redis
make dev-neo4j
make dev-nats

# 查看部署状态
make dev-status
```

### 部署生产环境

```bash
# ⚠️  部署前必须修改默认密码！
# 编辑以下文件中的密码：
# - overlays/prod/postgres/kustomization.yaml
# - overlays/prod/redis/kustomization.yaml
# - overlays/prod/neo4j/kustomization.yaml

# 部署所有服务（会有安全确认）
make prod-all

# 或者逐个部署
make prod-postgres
make prod-redis
make prod-neo4j
make prod-nats

# 查看部署状态
make prod-status
```

## 常用操作

### 查看日志

```bash
# 开发环境日志
make logs-dev-postgres
make logs-dev-redis
make logs-dev-neo4j
make logs-dev-nats

# 生产环境日志
make logs-prod-postgres
make logs-prod-redis
make logs-prod-neo4j
make logs-prod-nats
```

### 端口转发（本地访问）

```bash
# 开发环境端口转发
make forward-dev-postgres    # localhost:5432
make forward-dev-redis       # localhost:6379
make forward-dev-neo4j       # localhost:7474, 7687
make forward-dev-nats        # localhost:4222, 8222

# 生产环境端口转发（使用不同端口避免冲突）
make forward-prod-postgres   # localhost:15432
make forward-prod-redis      # localhost:16379
make forward-prod-neo4j      # localhost:17474, 17687
make forward-prod-nats       # localhost:14222, 18222
```

### 连接到数据库

```bash
# PostgreSQL
make db-connect-dev
make db-connect-prod

# Redis
make redis-cli-dev
make redis-cli-prod

# Neo4j
make neo4j-shell-dev
make neo4j-shell-prod
```

### 监控和调试

```bash
# 实时监控资源
make watch-dev
make watch-prod

# 查看资源使用
make top-dev
make top-prod

# 查看事件
make events-dev
make events-prod
```

## 验证和测试

```bash
# 验证配置文件
make validate-dev
make validate-prod
make validate-all

# 查看部署差异
make diff-dev
make diff-prod

# 干运行（不实际部署）
make dry-run-dev
make dry-run-prod

# 渲染最终 YAML
make render-dev
make render-prod
```

## 备份和恢复

### PostgreSQL 备份

```bash
# 备份开发环境数据库
make backup-postgres-dev

# 备份生产环境数据库
make backup-postgres-prod

# 恢复数据库（需指定备份文件）
make restore-postgres-dev BACKUP_FILE=backups/postgres-dev-20240101-120000.sql
```

## 部署单个服务

### 开发环境

```bash
# 只部署 PostgreSQL
make dev-postgres

# 只部署 Redis
make dev-redis

# 只部署 Neo4j
make dev-neo4j

# 只部署 NATS
make dev-nats
```

### 生产环境

```bash
# 只部署 PostgreSQL
make prod-postgres

# 只部署 Redis
make prod-redis

# 只部署 Neo4j
make prod-neo4j

# 只部署 NATS
make prod-nats
```

## 清理环境

```bash
# 清理开发环境（需要确认）
make dev-clean

# 清理生产环境（需要输入 'delete-production' 确认）
make prod-clean
```

## 查看连接信息

```bash
# 显示所有服务的连接信息
make info
```

输出示例：
```
=== Development Environment ===
Namespace: aetherius-dev

PostgreSQL:
  Host: postgres.aetherius-dev.svc.cluster.local
  Port: 5432
  User: postgres
  Password: dev-postgres-password
  Database: aetherius

Redis:
  Host: redis.aetherius-dev.svc.cluster.local
  Port: 6379
  Password: dev-redis-password

Neo4j:
  HTTP: http://neo4j.aetherius-dev.svc.cluster.local:7474
  Bolt: bolt://neo4j.aetherius-dev.svc.cluster.local:7687
  User: neo4j
  Password: dev-neo4j-password

NATS:
  URL: nats://nats.aetherius-dev.svc.cluster.local:4222
  Monitor: http://nats.aetherius-dev.svc.cluster.local:8222
```

## 渲染特定服务的 YAML

```bash
# 只查看某个服务的最终 YAML
make render-dev-postgres
make render-dev-redis
make render-dev-neo4j
make render-dev-nats
```

## 故障排查

### Pod 无法启动

```bash
# 查看 Pod 状态
make dev-status

# 查看详细事件
make events-dev

# 查看 Pod 日志
make logs-dev-postgres  # 或其他服务

# 查看 Pod 详情
kubectl describe pod -n aetherius-dev <pod-name>
```

### 存储问题

```bash
# 查看 PVC 状态
kubectl get pvc -n aetherius-dev

# 查看 PV
kubectl get pv

# 检查存储类
kubectl get storageclass
```

### 连接问题

```bash
# 测试 PostgreSQL 连接
kubectl run -it --rm psql-test --image=postgres:15-alpine -n aetherius-dev -- \
  psql -h postgres -U postgres -d aetherius

# 测试 Redis 连接
kubectl run -it --rm redis-test --image=redis:7-alpine -n aetherius-dev -- \
  redis-cli -h redis -p 6379 -a dev-redis-password ping

# 测试 NATS 连接
kubectl run -it --rm nats-test --image=natsio/nats-box:latest -n aetherius-dev -- \
  nats-bench -s nats://nats:4222 pub test --msgs=1
```

## 高级用法

### 自定义命名空间

```bash
# 编辑 overlay 配置文件
vi overlays/dev/kustomization.yaml

# 修改 namespace 字段
namespace: custom-namespace
```

### 调整资源配置

```bash
# 编辑 patch 文件
vi overlays/dev/postgres/deployment-patch.yaml

# 修改 resources 部分
resources:
  requests:
    memory: "512Mi"
    cpu: "500m"
  limits:
    memory: "1Gi"
    cpu: "1000m"
```

### 修改存储大小

```bash
# 编辑 PVC patch
vi overlays/dev/postgres/pvc-patch.yaml

# 修改 storage
spec:
  resources:
    requests:
      storage: 20Gi
```

## 最佳实践

### 开发环境

1. ✅ 使用较小的资源配置
2. ✅ 可以使用简单密码
3. ✅ 使用 emptyDir 或小容量 PVC
4. ✅ 单副本部署

### 生产环境

1. ⚠️ **必须修改默认密码**
2. ✅ 使用充足的资源配置
3. ✅ 启用持久化存储
4. ✅ 配置高可用（多副本）
5. ✅ 启用 TLS 加密
6. ✅ 配置备份策略
7. ✅ 设置资源限制和请求

## 获取帮助

```bash
# 显示所有可用命令
make help

# 或者直接运行
make
```

## 相关文档

- [README.md](./README.md) - 完整文档
- [Kustomize 官方文档](https://kustomize.io/)
- [Kubectl Kustomize 文档](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/)
