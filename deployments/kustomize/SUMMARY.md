# Kustomize 部署总结

## ✅ 已完成的工作

### 1. 目录结构

已创建完整的 Kustomize 部署结构：

```
deployments/kustomize/
├── README.md                   # 完整文档
├── QUICKSTART.md              # 快速入门指南
├── SUMMARY.md                 # 本文件
├── Makefile                   # 部署管理工具
├── base/                      # 基础配置（4个中间件）
│   ├── postgres/             # PostgreSQL
│   ├── redis/                # Redis
│   ├── neo4j/                # Neo4j
│   └── nats/                 # NATS
└── overlays/                 # 环境特定配置
    ├── dev/                  # 开发环境
    └── prod/                 # 生产环境
```

### 2. 支持的中间件

#### PostgreSQL (主数据库)
- ✅ Deployment 配置
- ✅ Service (ClusterIP)
- ✅ PersistentVolumeClaim (持久化存储)
- ✅ ConfigMap (初始化脚本和配置)
- ✅ Secret (用户名密码)
- ✅ 健康检查探针
- ✅ 资源限制配置

**开发环境**: 5Gi 存储, 512Mi 内存
**生产环境**: 50Gi 存储, 2Gi 内存, 性能优化

#### Redis (缓存)
- ✅ Deployment 配置
- ✅ Service (ClusterIP)
- ✅ ConfigMap (Redis 配置)
- ✅ Secret (密码)
- ✅ 健康检查探针
- ✅ 内存管理和淘汰策略

**开发环境**: emptyDir, 256Mi 内存
**生产环境**: 1Gi 内存, 性能优化

#### Neo4j (知识图谱)
- ✅ StatefulSet 配置
- ✅ Service (ClusterIP + Headless)
- ✅ VolumeClaimTemplates (data/logs/plugins)
- ✅ Secret (认证信息)
- ✅ 健康检查探针
- ✅ 内存堆配置

**开发环境**: 5Gi 存储, 1Gi 内存
**生产环境**: 50Gi 存储, 4Gi 内存

#### NATS (消息队列)
- ✅ StatefulSet 配置
- ✅ Service (ClusterIP + Headless)
- ✅ VolumeClaimTemplate (JetStream 持久化)
- ✅ ConfigMap (服务器配置)
- ✅ JetStream 支持
- ✅ 监控端口

**开发环境**: 1副本, 2Gi 存储
**生产环境**: 3副本集群 (HA), 20Gi 存储

### 3. Makefile 功能

#### 部署管理 (18个命令)
```bash
# 开发环境
make dev-all          # 部署所有服务
make dev-postgres     # 单独部署 PostgreSQL
make dev-redis        # 单独部署 Redis
make dev-neo4j        # 单独部署 Neo4j
make dev-nats         # 单独部署 NATS
make dev-status       # 查看状态
make dev-clean        # 清理环境

# 生产环境
make prod-all         # 部署所有服务（带安全确认）
make prod-postgres    # 单独部署 PostgreSQL
make prod-redis       # 单独部署 Redis
make prod-neo4j       # 单独部署 Neo4j
make prod-nats        # 单独部署 NATS
make prod-status      # 查看状态
make prod-clean       # 清理环境（需输入确认）
```

#### 验证和测试 (7个命令)
```bash
make validate-dev     # 验证开发配置
make validate-prod    # 验证生产配置
make validate-all     # 验证所有配置
make diff-dev         # 查看开发环境差异
make diff-prod        # 查看生产环境差异
make dry-run-dev      # 干运行（开发）
make dry-run-prod     # 干运行（生产）
```

#### 日志和调试 (8个命令)
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

#### 端口转发 (8个命令)
```bash
# 开发环境（标准端口）
make forward-dev-postgres    # :5432
make forward-dev-redis       # :6379
make forward-dev-neo4j       # :7474, :7687
make forward-dev-nats        # :4222, :8222

# 生产环境（高位端口）
make forward-prod-postgres   # :15432
make forward-prod-redis      # :16379
make forward-prod-neo4j      # :17474, :17687
make forward-prod-nats       # :14222, :18222
```

#### 数据库操作 (6个命令)
```bash
make db-connect-dev      # 连接 PostgreSQL (dev)
make db-connect-prod     # 连接 PostgreSQL (prod)
make redis-cli-dev       # 连接 Redis (dev)
make redis-cli-prod      # 连接 Redis (prod)
make neo4j-shell-dev     # 连接 Neo4j (dev)
make neo4j-shell-prod    # 连接 Neo4j (prod)
```

#### 备份和恢复 (3个命令)
```bash
make backup-postgres-dev              # 备份开发环境
make backup-postgres-prod             # 备份生产环境
make restore-postgres-dev BACKUP_FILE=<file>  # 恢复数据库
```

#### 监控 (6个命令)
```bash
make watch-dev        # 实时监控开发环境
make watch-prod       # 实时监控生产环境
make top-dev          # 资源使用（开发）
make top-prod         # 资源使用（生产）
make events-dev       # 事件列表（开发）
make events-prod      # 事件列表（生产）
```

#### 工具命令 (7个命令)
```bash
make render-dev           # 渲染开发环境 YAML
make render-prod          # 渲染生产环境 YAML
make render-dev-postgres  # 渲染单个服务 YAML
make render-dev-redis
make render-dev-neo4j
make render-dev-nats
make check-tools          # 检查所需工具
make info                 # 显示连接信息
make help                 # 显示帮助信息
```

**总计**: 60+ 个 Make 命令

### 4. 文档

- ✅ **README.md**: 完整的部署文档，包含故障排查、性能调优、安全建议
- ✅ **QUICKSTART.md**: 快速入门指南，涵盖常用操作
- ✅ **SUMMARY.md**: 本总结文档
- ✅ Makefile 内置帮助系统 (`make help`)

## 📊 配置对比

| 项目 | 开发环境 | 生产环境 |
|------|---------|---------|
| **命名空间** | aetherius-dev | aetherius-prod |
| **PostgreSQL 存储** | 5Gi | 50Gi |
| **PostgreSQL 内存** | 512Mi | 2Gi |
| **Redis 内存** | 256Mi | 1Gi |
| **Neo4j 存储** | 5Gi | 50Gi |
| **Neo4j 内存** | 1Gi | 4Gi |
| **NATS 副本数** | 1 | 3 (HA集群) |
| **NATS 存储** | 2Gi | 20Gi |
| **密码策略** | 简单密码 | 强密码（需修改） |

## 🚀 快速开始

### 1. 检查环境

```bash
cd deployments/kustomize
make check-tools
```

### 2. 部署开发环境

```bash
make dev-all
make dev-status
```

### 3. 查看连接信息

```bash
make info
```

### 4. 端口转发测试

```bash
# 打开新终端窗口
make forward-dev-postgres

# 在另一个终端测试连接
psql -h localhost -U postgres -d aetherius
# 密码: dev-postgres-password
```

## 🔒 生产环境部署注意事项

### ⚠️ 部署前必做

1. **修改所有默认密码**:
   ```bash
   vi overlays/prod/postgres/kustomization.yaml
   vi overlays/prod/redis/kustomization.yaml
   vi overlays/prod/neo4j/kustomization.yaml
   ```

2. **检查存储类**:
   ```bash
   kubectl get storageclass
   # 确保集群有可用的 StorageClass
   ```

3. **验证资源配额**:
   ```bash
   kubectl describe quota -n aetherius-prod
   ```

4. **验证配置**:
   ```bash
   make validate-prod
   make dry-run-prod
   ```

### 部署

```bash
make prod-all
# 会提示确认密码已修改和部署确认
```

## 📈 监控和维护

### 日常监控

```bash
# 实时监控
make watch-prod

# 查看资源使用
make top-prod

# 查看最近事件
make events-prod
```

### 日志查看

```bash
# 查看特定服务日志
make logs-prod-postgres
make logs-prod-redis
make logs-prod-neo4j
make logs-prod-nats
```

### 定期备份

```bash
# 每日备份 PostgreSQL
make backup-postgres-prod

# 备份文件位置: ./backups/postgres-prod-YYYYMMDD-HHMMSS.sql
```

## 🛠️ 自定义配置

### 修改资源配置

编辑对应环境的 patch 文件：

```bash
# 开发环境
vi overlays/dev/postgres/deployment-patch.yaml

# 生产环境
vi overlays/prod/postgres/deployment-patch.yaml
```

### 修改存储大小

```bash
# 开发环境
vi overlays/dev/postgres/pvc-patch.yaml

# 生产环境
vi overlays/prod/postgres/pvc-patch.yaml
```

### 修改密码

```bash
# 编辑 kustomization.yaml 中的 secretGenerator
vi overlays/prod/postgres/kustomization.yaml
```

重新部署后新密码生效。

## 🔧 故障排查

### Pod 无法启动

```bash
# 查看状态
make dev-status

# 查看事件
make events-dev

# 查看日志
make logs-dev-postgres

# 查看详情
kubectl describe pod -n aetherius-dev <pod-name>
```

### 存储问题

```bash
# 查看 PVC
kubectl get pvc -n aetherius-dev

# 查看 PV
kubectl get pv

# 检查存储类
kubectl get storageclass
```

### 连接测试

```bash
# PostgreSQL
kubectl run -it --rm psql-test --image=postgres:15-alpine -n aetherius-dev -- \
  psql -h postgres -U postgres -d aetherius

# Redis
kubectl run -it --rm redis-test --image=redis:7-alpine -n aetherius-dev -- \
  redis-cli -h redis -p 6379 -a dev-redis-password ping

# NATS
kubectl run -it --rm nats-test --image=natsio/nats-box:latest -n aetherius-dev -- \
  nats-bench -s nats://nats:4222 pub test --msgs=1
```

## 📚 相关资源

- [Kustomize 官方文档](https://kustomize.io/)
- [Kubectl Kustomize](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/)
- [PostgreSQL Docker](https://hub.docker.com/_/postgres)
- [Redis Docker](https://hub.docker.com/_/redis)
- [Neo4j Docker](https://hub.docker.com/_/neo4j)
- [NATS Docker](https://hub.docker.com/_/nats)

## 🎯 下一步

1. ✅ 中间件部署配置已完成
2. ⏭️ 部署 Aetherius 应用服务
3. ⏭️ 配置 Ingress 和服务暴露
4. ⏭️ 设置监控和告警
5. ⏭️ 配置备份和灾难恢复

## 🤝 贡献

如需改进此部署配置，请遵循以下流程：

1. 修改 base/ 或 overlays/ 中的配置
2. 运行 `make validate-all` 验证
3. 运行 `make dry-run-dev` 测试
4. 提交 PR 并说明改动原因

## 📝 版本历史

- **v1.0.0** (2025-01-XX): 初始版本
  - 完整的 Kustomize 配置
  - 4个中间件支持
  - 60+ Make 命令
  - 开发/生产环境支持
