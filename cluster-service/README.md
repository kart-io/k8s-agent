# Cluster Service

K8s 集群管理微服务 - 负责管理多个 Kubernetes 集群、集群资源监控、部署管理等功能。

> 🎉 **最新更新**:
> - **Phase 1 已完成！** 新增 DaemonSet、ConfigMap、Secret 管理功能 (2025-10-17)
> - **版本管理集成完成！** 使用 kart-io/version 包进行版本管理 (2025-10-17)
>
> 📚 **完整文档**: 查看 [README_DOCS.md](./README_DOCS.md) 了解所有文档
>
> 🚀 **快速开始**: 查看 [FINAL_SUMMARY.md](./FINAL_SUMMARY.md) 了解 Phase 1 完成情况
>
> 🔖 **版本管理**: 查看 [VERSION_INTEGRATION.md](./VERSION_INTEGRATION.md) 了解版本管理

## 功能特性

### 1. 集群管理
- 多集群接入和管理
- 集群状态监控
- 集群配置管理
- 集群健康检查

### 2. 资源管理 (✅ Phase 1 完成)
- ✅ Pod 管理（列表、详情、日志、删除）
- ✅ Deployment 管理（列表、详情、扩缩容、重启）
- ✅ StatefulSet 管理（列表、详情、扩缩容、重启、删除）
- ✅ DaemonSet 管理（列表、详情、重启、删除）**← 新增**
- ✅ Service 管理（完整 CRUD）
- ✅ ConfigMap 管理（完整 CRUD）**← 新增**
- ✅ Secret 管理（完整 CRUD + 安全控制）**← 新增**
- ✅ Namespace 管理（列表、详情、创建、删除）
- ✅ Node 管理（列表、详情、cordon/uncordon/drain）

**API 覆盖率**: 47/119 (39%)

### 3. 监控统计
- 集群资源使用率（CPU、内存、存储）
- Node 状态监控
- Pod 状态统计
- 事件监控

### 4. 部署管理
- 应用部署
- 滚动更新
- 回滚操作
- HPA（水平自动扩缩容）配置

## 架构设计

```
cluster-service/
├── cmd/
│   └── server/          # 服务入口
│       └── main.go
├── configs/             # 配置文件
│   └── config.yaml
├── internal/
│   ├── api/            # API 路由
│   │   └── server.go
│   ├── handler/        # HTTP 处理器
│   │   ├── cluster.go
│   │   ├── pod.go
│   │   ├── deployment.go
│   │   └── resource.go
│   ├── service/        # 业务逻辑
│   │   ├── cluster.go
│   │   └── k8s_manager.go
│   ├── storage/        # 数据存储
│   │   └── postgres.go
│   ├── k8s/            # K8s 客户端封装
│   │   ├── client.go
│   │   └── manager.go
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

### 版本管理 API (新增 ⭐)

- `GET /version` - 获取完整版本信息 (JSON格式)
- `GET /version/simple` - 获取简化版本信息
- `GET /version/text` - 获取文本格式版本信息
- `GET /version/json` - 获取原始 JSON 格式版本信息

> 📖 **版本管理文档**: 查看 [VERSION_INTEGRATION.md](./VERSION_INTEGRATION.md) 了解详细使用方法

---

### 新 K8s API (基于 `/api/k8s`, Phase 1 完成)

#### 集群管理 (6个接口)
- `GET /api/k8s/clusters` - 获取集群列表
- `POST /api/k8s/clusters` - 创建集群
- `GET /api/k8s/clusters/:id` - 获取集群详情
- `PUT /api/k8s/clusters/:id` - 更新集群
- `DELETE /api/k8s/clusters/:id` - 删除集群
- `GET /api/k8s/clusters/:id/health` - 集群健康检查

#### DaemonSet 管理 (4个接口) ⭐ 新增
- `GET /api/k8s/clusters/:clusterId/namespaces/:namespace/daemonsets`
- `GET /api/k8s/clusters/:clusterId/namespaces/:namespace/daemonsets/:name`
- `POST /api/k8s/clusters/:clusterId/namespaces/:namespace/daemonsets/:name/restart`
- `DELETE /api/k8s/clusters/:clusterId/namespaces/:namespace/daemonsets/:name`

#### ConfigMap 管理 (5个接口) ⭐ 新增
- `GET /api/k8s/clusters/:clusterId/namespaces/:namespace/configmaps`
- `GET /api/k8s/clusters/:clusterId/namespaces/:namespace/configmaps/:name`
- `POST /api/k8s/clusters/:clusterId/namespaces/:namespace/configmaps`
- `PUT /api/k8s/clusters/:clusterId/namespaces/:namespace/configmaps/:name`
- `DELETE /api/k8s/clusters/:clusterId/namespaces/:namespace/configmaps/:name`

#### Secret 管理 (5个接口) ⭐ 新增
- `GET /api/k8s/clusters/:clusterId/namespaces/:namespace/secrets`
- `GET /api/k8s/clusters/:clusterId/namespaces/:namespace/secrets/:name?includeData=true`
- `POST /api/k8s/clusters/:clusterId/namespaces/:namespace/secrets`
- `PUT /api/k8s/clusters/:clusterId/namespaces/:namespace/secrets/:name`
- `DELETE /api/k8s/clusters/:clusterId/namespaces/:namespace/secrets/:name`

> 📖 **更多接口**: 查看 [API_IMPLEMENTATION_PLAN.md](./API_IMPLEMENTATION_PLAN.md) 了解所有 47 个 API

---

### 原有 API (基于 `/api/v1`, 保持向后兼容)

### 集群管理
- `GET /api/v1/clusters` - 获取集群列表
- `POST /api/v1/clusters` - 添加集群
- `GET /api/v1/clusters/:id` - 获取集群详情
- `PUT /api/v1/clusters/:id` - 更新集群
- `DELETE /api/v1/clusters/:id` - 删除集群
- `GET /api/v1/clusters/:id/health` - 集群健康检查
- `GET /api/v1/clusters/:id/nodes` - 获取节点列表

### Pod 管理
- `GET /api/v1/clusters/:cluster_id/namespaces/:namespace/pods` - Pod 列表
- `GET /api/v1/clusters/:cluster_id/namespaces/:namespace/pods/:name` - Pod 详情
- `DELETE /api/v1/clusters/:cluster_id/namespaces/:namespace/pods/:name` - 删除 Pod
- `GET /api/v1/clusters/:cluster_id/namespaces/:namespace/pods/:name/logs` - Pod 日志
- `POST /api/v1/clusters/:cluster_id/namespaces/:namespace/pods/:name/exec` - 执行命令

### Deployment 管理
- `GET /api/v1/clusters/:cluster_id/namespaces/:namespace/deployments` - Deployment 列表
- `POST /api/v1/clusters/:cluster_id/namespaces/:namespace/deployments` - 创建 Deployment
- `GET /api/v1/clusters/:cluster_id/namespaces/:namespace/deployments/:name` - Deployment 详情
- `PUT /api/v1/clusters/:cluster_id/namespaces/:namespace/deployments/:name` - 更新 Deployment
- `DELETE /api/v1/clusters/:cluster_id/namespaces/:namespace/deployments/:name` - 删除 Deployment
- `POST /api/v1/clusters/:cluster_id/namespaces/:namespace/deployments/:name/scale` - 扩缩容

### 资源统计
- `GET /api/v1/clusters/:id/stats` - 集群资源统计
- `GET /api/v1/clusters/:id/events` - 集群事件

## 配置说明

```yaml
server:
  port: 8082
  mode: debug

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: cluster_db

jwt:
  secret: your-secret-key-change-in-production

k8s:
  config_path: ~/.kube/config
  in_cluster: false
```

## 快速开始

### 方式 1: 快速测试 (推荐)

```bash
# 1. 构建服务 (带版本注入)
cd /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent/cluster-service
make build

# 2. 查看版本信息
make version

# 3. 配置环境(如果需要)
# 编辑 configs/config.yaml 或设置环境变量

# 4. 启动服务
./bin/cluster-service -config configs/config.yaml

# 5. 测试版本端点 (在另一个终端)
curl http://localhost:8082/version
curl http://localhost:8082/version/simple

# 6. 运行 API 测试 (在另一个终端)
export BASE_URL="http://localhost:8082"
export CLUSTER_ID="your-cluster-id"
export NAMESPACE="default"
./test-new-apis.sh
```

> 📖 **详细指南**:
> - [QUICKSTART_TEST.md](./QUICKSTART_TEST.md) - API 测试指南
> - [VERSION_INTEGRATION.md](./VERSION_INTEGRATION.md) - 版本管理指南

---

### 方式 2: 使用 Make

```bash
# 显示所有可用命令
make help

# 构建 (带版本注入)
make build

# 查看版本信息
make version

# 运行
make run

# 测试
make test

# 清理
make clean
```

---

### 方式 3: Docker

```bash
# Docker 构建
make docker-build

# Docker 运行
docker run -d -p 8082:8082 \
  -v $(pwd)/configs:/app/configs \
  cluster-service:latest
```

## 文档导航

### 📚 核心文档
- **[FINAL_SUMMARY.md](./FINAL_SUMMARY.md)** ⭐ - Phase 1 完整总结
- **[README_DOCS.md](./README_DOCS.md)** ⭐ - 所有文档索引
- **[CODE_VERIFICATION_REPORT.md](./CODE_VERIFICATION_REPORT.md)** - 代码质量报告
- **[VERSION_INTEGRATION.md](./VERSION_INTEGRATION.md)** ⭐ 新增 - 版本管理指南

### 📖 实施文档
- [API_IMPLEMENTATION_PLAN.md](./API_IMPLEMENTATION_PLAN.md) - 4阶段实现计划
- [PHASE1_COMPLETION_REPORT.md](./PHASE1_COMPLETION_REPORT.md) - Phase 1 详细报告
- [LATEST_UPDATE.md](./LATEST_UPDATE.md) - 最新更新说明

### 🧪 测试文档
- [QUICKSTART_TEST.md](./QUICKSTART_TEST.md) - 快速测试指南
- [test-new-apis.sh](./test-new-apis.sh) - 自动化测试脚本

### 🚀 部署文档
- [DEPLOYMENT.md](./DEPLOYMENT.md) - 生产部署指南
- [PROJECT_COMPLETION.md](./PROJECT_COMPLETION.md) - 项目总结

## 技术特性

### 版本管理 ✨ 新增
- ✅ 构建时版本注入 (使用 kart-io/version)
- ✅ Git 信息跟踪 (commit, branch, tree state)
- ✅ 多种输出格式 (JSON, text, simplified)
- ✅ 4 个版本查询 API 端点
- ✅ 日志中自动包含版本信息

### K8s 资源管理
- ✅ 10 种 K8s 资源类型完整支持
- ✅ 47 个 REST API 接口
- ✅ 统一的错误处理和日志记录
- ✅ Secret 安全控制
- ✅ 分页支持

## 项目状态

### ✅ 已完成
- Phase 1: DaemonSet、ConfigMap、Secret API (2025-10-17)
- 版本管理集成: kart-io/version 包 (2025-10-17)
- 10 种 K8s 资源类型完整支持
- 47 个 REST API 接口 + 4 个版本 API 接口
- 完整的文档和测试

### 🚧 进行中
- 集成测试验证
- 单元测试编写
- 性能测试

### 📋 计划中
- Phase 2: P0 优先级资源 (ReplicaSet, Job, CronJob, Ingress, PV/PVC)
- Phase 3: P1 优先级资源 (RBAC, NetworkPolicy)
- Phase 4: P2 优先级资源 (ResourceQuota, HPA, Events)

## K8s 配置

服务支持两种方式连接 K8s 集群：

1. **外部访问**：使用 kubeconfig 文件
2. **In-Cluster**：以 Pod 形式部署在 K8s 集群内

## 依赖服务

- PostgreSQL: 存储集群配置信息

## 环境变量

- `CLUSTER_CONFIG_PATH`: 配置文件路径
- `CLUSTER_LOG_LEVEL`: 日志级别
- `CLUSTER_PORT`: 服务端口
- `KUBECONFIG`: K8s 配置文件路径

## License

MIT
