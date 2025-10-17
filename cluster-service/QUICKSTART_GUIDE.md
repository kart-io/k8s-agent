# 🚀 Cluster Service 快速启动指南

本指南帮助你快速启动并测试 cluster-service。

---

## ✅ 前置条件

- Docker（用于运行 MySQL）
- Go 1.21+
- make

---

## 📦 快速启动（3步）

### 步骤 1: 启动 MySQL 数据库

```bash
./setup-mysql.sh
```

等待输出：
```
✅ MySQL container started successfully
✅ MySQL is ready
```

### 步骤 2: 启动服务

```bash
make run-local
```

看到以下输出表示成功：
```
✅ Successfully connected to MySQL
✅ Database schema initialized
✅ K8s API routes registered
✅ Server starting on port 8082
```

### 步骤 3: 测试 API

```bash
# 健康检查
curl http://localhost:8082/health

# 版本信息
curl http://localhost:8082/version/simple

# K8s API
curl http://localhost:8082/api/k8s/clusters
```

---

## 🔧 常见问题

### 问题 1: 端口被占用
**错误**: `bind: address already in use`

**解决**:
```bash
# 方法1: 杀掉占用端口的进程
fuser -k 8082/tcp

# 方法2: 查找并手动杀掉
lsof -ti:8082 | xargs kill -9
```

### 问题 2: MySQL 连接失败
**错误**: `Failed to connect to MySQL`

**解决**:
```bash
# 检查 MySQL 容器状态
docker ps | grep cluster-mysql

# 重启 MySQL 容器
docker restart cluster-mysql

# 或者重新设置
docker rm -f cluster-mysql
./setup-mysql.sh
```

### 问题 3: 数据库表不存在
**错误**: `Table 'cluster_db.clusters' doesn't exist`

**解决**: 这个问题已在代码中修复，服务启动时会自动创建表。如果仍有问题：
```bash
# 重启服务会自动创建表
make run-local
```

---

## 📚 API 端点列表

### 基础端点
- `GET /health` - 健康检查
- `GET /version` - 版本信息（完整）
- `GET /version/simple` - 版本信息（简化）
- `GET /version/text` - 版本信息（文本）
- `GET /version/json` - 版本信息（JSON）

### K8s API 端点

#### 集群管理
- `GET /api/k8s/clusters` - 获取集群列表
- `POST /api/k8s/clusters` - 创建集群
- `GET /api/k8s/clusters/:clusterId` - 获取集群详情
- `PUT /api/k8s/clusters/:clusterId` - 更新集群
- `DELETE /api/k8s/clusters/:clusterId` - 删除集群
- `GET /api/k8s/clusters/:clusterId/health` - 集群健康状态

#### 命名空间管理
- `GET /api/k8s/clusters/:clusterId/namespaces` - 获取命名空间列表
- `POST /api/k8s/clusters/:clusterId/namespaces` - 创建命名空间
- `GET /api/k8s/clusters/:clusterId/ns/:namespace` - 获取命名空间详情
- `DELETE /api/k8s/clusters/:clusterId/ns/:namespace` - 删除命名空间

#### Pod 管理
- `GET /api/k8s/clusters/:clusterId/namespaces/:namespace/pods` - Pod 列表
- `GET /api/k8s/clusters/:clusterId/namespaces/:namespace/pods/:name` - Pod 详情
- `DELETE /api/k8s/clusters/:clusterId/namespaces/:namespace/pods/:name` - 删除 Pod
- `GET /api/k8s/clusters/:clusterId/namespaces/:namespace/pods/:name/logs` - Pod 日志

...（更多端点见完整文档）

---

## 🛠️ 开发命令

```bash
# 编译
make build

# 运行测试
make test

# 清理
make clean

# 查看帮助
make help
```

---

## 🔍 调试技巧

### 查看服务日志
```bash
# 如果服务在后台运行
tail -f /tmp/cluster-service.log

# 或者前台运行查看实时日志
make run-local
```

### 检查数据库
```bash
# 连接到 MySQL
docker exec -it cluster-mysql mysql -u cluster_user -pcluster_pass cluster_db

# 查看表
mysql> SHOW TABLES;

# 查看集群数据
mysql> SELECT * FROM clusters;
```

### 测试 API
```bash
# 使用 curl
curl -X POST http://localhost:8082/api/k8s/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-cluster",
    "endpoint": "https://k8s.example.com",
    "kubeconfig": "..."
  }'

# 使用 httpie（更友好）
http POST http://localhost:8082/api/k8s/clusters \
  name=test-cluster \
  endpoint=https://k8s.example.com
```

---

## 📖 更多文档

- [完整修复报告](./FINAL_FIX_REPORT.md) - 所有问题的详细修复过程
- [MySQL 迁移报告](./MYSQL_MIGRATION_REPORT.md) - PostgreSQL 到 MySQL 的迁移
- [路由修复总结](./ROUTE_FIX_SUMMARY.md) - 路由冲突修复细节
- [MySQL 故障排查](./MYSQL_TROUBLESHOOTING.md) - MySQL 相关问题排查

---

## ✅ 验证清单

启动服务后，确认以下项目：

- [ ] MySQL 容器运行中
- [ ] 服务在 8082 端口监听
- [ ] 健康检查端点返回 OK
- [ ] 版本端点返回版本信息
- [ ] K8s API 端点返回正确响应
- [ ] 数据库 clusters 表已创建
- [ ] 日志输出正常

---

**💡 提示**: 如遇到任何问题，请查看 [FINAL_FIX_REPORT.md](./FINAL_FIX_REPORT.md) 获取详细的故障排查指南。

**🎉 享受使用 Cluster Service！**
