# Cluster Service

K8s 集群管理微服务 - 负责管理多个 Kubernetes 集群、集群资源监控、部署管理等功能。

## 功能特性

### 1. 集群管理
- 多集群接入和管理
- 集群状态监控
- 集群配置管理
- 集群健康检查

### 2. 资源管理
- Pod 管理（列表、详情、日志、执行命令）
- Deployment 管理（创建、更新、删除、扩缩容）
- Service 管理
- ConfigMap/Secret 管理
- Namespace 管理

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

### 构建
```bash
make build
```

### 运行
```bash
make run
```

### 测试
```bash
make test
```

### Docker 构建
```bash
make docker-build
```

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
