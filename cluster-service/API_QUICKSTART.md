# K8s Agent API - 快速启动指南

本指南帮助您快速启动并使用新实现的 K8s API 功能。

## 目录

- [快速开始](#快速开始)
- [API 使用示例](#api-使用示例)
- [完整功能列表](#完整功能列表)
- [故障排查](#故障排查)

## 快速开始

### 1. 启动服务（启用 K8s API）

```bash
cd /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent/cluster-service

# 启动服务（默认启用 K8s API）
go run cmd/server/main.go -config configs/config-dev.yaml

# 或禁用 K8s API（仅使用原有端点）
go run cmd/server/main.go -config configs/config-dev.yaml --enable-k8s-api=false
```

### 2. 验证服务

```bash
# 健康检查
curl http://localhost:8082/health

# 预期输出
{"status":"ok","service":"cluster-service"}
```

## API 使用示例

### 统一响应格式

所有 API 返回统一的响应格式：

```json
{
  "code": 0,
  "message": "Success",
  "data": {
    ...
  }
}
```

错误响应：

```json
{
  "code": 400,
  "message": "Invalid request",
  "error": "详细错误信息"
}
```

### 1. 集群管理 API

Base Path: `/api/k8s/clusters`

#### 创建集群

```bash
curl -X POST http://localhost:8082/api/k8s/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-cluster",
    "description": "测试集群",
    "endpoint": "https://192.168.1.100:6443",
    "region": "local",
    "provider": "kubernetes",
    "kubeconfig": "..."
  }'
```

响应：

```json
{
  "code": 0,
  "message": "Cluster created successfully",
  "data": {
    "id": "cluster-uuid",
    "name": "test-cluster",
    "endpoint": "https://192.168.1.100:6443",
    "version": "v1.28.0",
    "status": "healthy",
    "createdAt": "2025-10-17T10:00:00Z"
  }
}
```

#### 获取集群列表（分页）

```bash
# 获取第 1 页，每页 20 条
curl "http://localhost:8082/api/k8s/clusters?page=1&pageSize=20"
```

响应：

```json
{
  "code": 0,
  "message": "Success",
  "data": {
    "items": [
      {
        "id": "cluster-1",
        "name": "prod-cluster",
        "status": "healthy",
        ...
      }
    ],
    "page": 1,
    "pageSize": 20,
    "total": 50
  }
}
```

#### 获取集群详情

```bash
curl http://localhost:8082/api/k8s/clusters/cluster-1
```

#### 更新集群

```bash
curl -X PUT http://localhost:8082/api/k8s/clusters/cluster-1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "updated-name",
    "description": "updated description"
  }'
```

#### 删除集群

```bash
curl -X DELETE http://localhost:8082/api/k8s/clusters/cluster-1
```

#### 获取集群健康状态

```bash
curl http://localhost:8082/api/k8s/clusters/cluster-1/health
```

响应：

```json
{
  "code": 0,
  "message": "Success",
  "data": {
    "clusterId": "cluster-1",
    "status": "healthy",
    "totalNodes": 3,
    "readyNodes": 3,
    "totalPods": 50,
    "runningPods": 48,
    "checkedAt": "2025-10-17T10:00:00Z"
  }
}
```

### 2. 命名空间管理 API

Base Path: `/api/k8s/clusters/{clusterId}/namespaces`

#### 获取命名空间列表

```bash
curl http://localhost:8082/api/k8s/clusters/cluster-1/namespaces
```

#### 创建命名空间

```bash
curl -X POST http://localhost:8082/api/k8s/clusters/cluster-1/namespaces \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-namespace",
    "labels": {
      "env": "test",
      "team": "platform"
    }
  }'
```

#### 获取命名空间详情

```bash
curl http://localhost:8082/api/k8s/clusters/cluster-1/namespaces/test-namespace
```

#### 删除命名空间

```bash
curl -X DELETE http://localhost:8082/api/k8s/clusters/cluster-1/namespaces/test-namespace
```

### 3. Pod 管理 API

Base Path: `/api/k8s/clusters/{clusterId}/namespaces/{namespace}/pods`

#### 获取 Pod 列表

```bash
# 获取 default 命名空间的 Pod
curl http://localhost:8082/api/k8s/clusters/cluster-1/namespaces/default/pods

# 分页查询
curl "http://localhost:8082/api/k8s/clusters/cluster-1/namespaces/default/pods?page=1&pageSize=20"
```

响应：

```json
{
  "code": 0,
  "message": "Success",
  "data": {
    "items": [
      {
        "name": "nginx-7d6b7d8f-xyz",
        "namespace": "default",
        "status": "Running",
        "phase": "Running",
        "nodeName": "node-1",
        "podIP": "10.244.0.5",
        "labels": {
          "app": "nginx"
        },
        "containers": [
          {
            "name": "nginx",
            "image": "nginx:1.21",
            "ready": true,
            "restartCount": 0,
            "state": "running"
          }
        ],
        "createdAt": "2025-10-17T09:00:00Z"
      }
    ],
    "page": 1,
    "pageSize": 20,
    "total": 50
  }
}
```

#### 获取 Pod 详情

```bash
curl http://localhost:8082/api/k8s/clusters/cluster-1/namespaces/default/pods/nginx-7d6b7d8f-xyz
```

#### 获取 Pod 日志

```bash
# 获取最后 100 行日志
curl "http://localhost:8082/api/k8s/clusters/cluster-1/namespaces/default/pods/nginx-7d6b7d8f-xyz/logs?tailLines=100"

# 指定容器名称
curl "http://localhost:8082/api/k8s/clusters/cluster-1/namespaces/default/pods/nginx-7d6b7d8f-xyz/logs?container=nginx&tailLines=50"

# 实时跟踪日志
curl "http://localhost:8082/api/k8s/clusters/cluster-1/namespaces/default/pods/nginx-7d6b7d8f-xyz/logs?follow=true&tailLines=10"
```

响应：

```json
{
  "code": 0,
  "message": "Success",
  "data": {
    "logs": "2025-10-17 10:00:00 INFO Starting nginx...\n2025-10-17 10:00:01 INFO Nginx started successfully"
  }
}
```

#### 删除 Pod

```bash
curl -X DELETE http://localhost:8082/api/k8s/clusters/cluster-1/namespaces/default/pods/nginx-7d6b7d8f-xyz
```

### 4. Deployment 管理 API

Base Path: `/api/k8s/clusters/{clusterId}/namespaces/{namespace}/deployments`

#### 获取 Deployment 列表

```bash
curl http://localhost:8082/api/k8s/clusters/cluster-1/namespaces/default/deployments
```

响应：

```json
{
  "code": 0,
  "message": "Success",
  "data": {
    "items": [
      {
        "name": "nginx",
        "namespace": "default",
        "replicas": 3,
        "availableReplicas": 3,
        "readyReplicas": 3,
        "updatedReplicas": 3,
        "labels": {
          "app": "nginx"
        },
        "selector": {
          "app": "nginx"
        },
        "strategy": "RollingUpdate",
        "createdAt": "2025-10-17T08:00:00Z"
      }
    ],
    "page": 1,
    "pageSize": 20,
    "total": 10
  }
}
```

#### 获取 Deployment 详情

```bash
curl http://localhost:8082/api/k8s/clusters/cluster-1/namespaces/default/deployments/nginx
```

#### 扩缩容 Deployment

```bash
curl -X PUT http://localhost:8082/api/k8s/clusters/cluster-1/namespaces/default/deployments/nginx/scale \
  -H "Content-Type: application/json" \
  -d '{
    "replicas": 5
  }'
```

响应：

```json
{
  "code": 0,
  "message": "Deployment scaled successfully",
  "data": {
    "name": "nginx",
    "replicas": 5,
    "availableReplicas": 5,
    ...
  }
}
```

#### 重启 Deployment

```bash
curl -X POST http://localhost:8082/api/k8s/clusters/cluster-1/namespaces/default/deployments/nginx/restart
```

响应：

```json
{
  "code": 0,
  "message": "Deployment restarted successfully",
  "data": {
    "deployment": "nginx"
  }
}
```

## 完整功能列表

### ✅ 已实现（18 个接口）

#### 集群管理（6 个接口）

- `GET /api/k8s/clusters` - 获取集群列表（分页）
- `GET /api/k8s/clusters/:id` - 获取集群详情
- `POST /api/k8s/clusters` - 创建集群
- `PUT /api/k8s/clusters/:id` - 更新集群
- `DELETE /api/k8s/clusters/:id` - 删除集群
- `GET /api/k8s/clusters/:id/health` - 获取集群健康状态

#### 命名空间管理（4 个接口）

- `GET /api/k8s/clusters/:clusterId/namespaces` - 获取命名空间列表
- `GET /api/k8s/clusters/:clusterId/namespaces/:name` - 获取命名空间详情
- `POST /api/k8s/clusters/:clusterId/namespaces` - 创建命名空间
- `DELETE /api/k8s/clusters/:clusterId/namespaces/:name` - 删除命名空间

#### Pod 管理（4 个接口）

- `GET /api/k8s/clusters/:clusterId/namespaces/:namespace/pods` - 获取 Pod 列表
- `GET /api/k8s/clusters/:clusterId/namespaces/:namespace/pods/:name` - 获取 Pod 详情
- `DELETE /api/k8s/clusters/:clusterId/namespaces/:namespace/pods/:name` - 删除 Pod
- `GET /api/k8s/clusters/:clusterId/namespaces/:namespace/pods/:name/logs` - 获取 Pod 日志

#### Deployment 管理（4 个接口）

- `GET /api/k8s/clusters/:clusterId/namespaces/:namespace/deployments` - 获取 Deployment 列表
- `GET /api/k8s/clusters/:clusterId/namespaces/:namespace/deployments/:name` - 获取 Deployment 详情
- `PUT /api/k8s/clusters/:clusterId/namespaces/:namespace/deployments/:name/scale` - 扩缩容
- `POST /api/k8s/clusters/:clusterId/namespaces/:namespace/deployments/:name/restart` - 重启

### 📋 待实现（101+ 个接口）

根据 API 文档，还需实现：

- Node 管理
- StatefulSet、DaemonSet、Job、CronJob 管理
- Service、Ingress 管理
- ConfigMap、Secret 管理
- 存储管理（PV、PVC、StorageClass）
- RBAC 管理（ServiceAccount、Role、RoleBinding 等）
- 资源配额管理（ResourceQuota、LimitRange）
- HPA、PriorityClass 管理
- Event 管理

## 故障排查

### 1. 服务启动失败

**错误**: `Failed to initialize common logger`

**解决方案**:
```bash
# 检查 common 包路径
ls -la ../common

# 运行 go mod tidy
go mod tidy

# 确保 replace 指令正确
cat go.mod | grep replace
```

### 2. 数据库连接失败

**错误**: `Failed to initialize PostgreSQL storage`

**解决方案**:
```bash
# 检查数据库配置
cat configs/config-dev.yaml | grep -A 8 database

# 测试数据库连接
psql -h localhost -U postgres -d cluster_db -c "SELECT 1"
```

### 3. K8s 集群连接失败

**错误**: `failed to connect to cluster` 或 `invalid kubeconfig`

**解决方案**:
```bash
# 验证 kubeconfig
kubectl --kubeconfig=your-config.yaml cluster-info

# 测试 K8s API 访问
kubectl --kubeconfig=your-config.yaml get nodes
```

### 4. 查看详细日志

```bash
# 修改配置文件，启用 debug 级别
cat > configs/config-dev.yaml <<EOF
logging:
  level: debug
  format: console
EOF

# 重启服务
go run cmd/server/main.go -config configs/config-dev.yaml
```

### 5. 限流问题

如果遇到 `Too Many Requests` 错误，说明触发了限流。

默认限流设置：100 req/s，桶容量 200

修改限流配置：编辑 `internal/api/server.go`

```go
engine.Use(middleware.RateLimitByIP(1000, 2000)) // 提高限流
```

## 测试脚本

创建一个测试脚本来快速验证所有 API：

```bash
#!/bin/bash

BASE_URL="http://localhost:8082"

echo "=== 1. 健康检查 ==="
curl -s $BASE_URL/health | jq .

echo -e "\n=== 2. 创建集群 ==="
CLUSTER_ID=$(curl -s -X POST $BASE_URL/api/k8s/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-cluster",
    "description": "Test cluster",
    "endpoint": "https://localhost:6443",
    "kubeconfig": "..."
  }' | jq -r '.data.id')

echo "Created cluster: $CLUSTER_ID"

echo -e "\n=== 3. 获取集群列表 ==="
curl -s "$BASE_URL/api/k8s/clusters?page=1&pageSize=10" | jq .

echo -e "\n=== 4. 获取集群详情 ==="
curl -s "$BASE_URL/api/k8s/clusters/$CLUSTER_ID" | jq .

echo -e "\n=== 5. 创建命名空间 ==="
curl -s -X POST "$BASE_URL/api/k8s/clusters/$CLUSTER_ID/namespaces" \
  -H "Content-Type: application/json" \
  -d '{"name": "test-ns"}' | jq .

echo -e "\n=== 6. 获取命名空间列表 ==="
curl -s "$BASE_URL/api/k8s/clusters/$CLUSTER_ID/namespaces" | jq .

echo -e "\n=== 7. 获取 Pod 列表 ==="
curl -s "$BASE_URL/api/k8s/clusters/$CLUSTER_ID/namespaces/default/pods" | jq .

echo -e "\n=== 8. 获取 Deployment 列表 ==="
curl -s "$BASE_URL/api/k8s/clusters/$CLUSTER_ID/namespaces/default/deployments" | jq .
```

保存为 `test-api.sh`，然后运行：

```bash
chmod +x test-api.sh
./test-api.sh
```

## 性能优化建议

1. **数据库连接池**: 根据负载调整连接池大小
2. **限流配置**: 根据实际 QPS 调整限流参数
3. **日志级别**: 生产环境使用 `info` 或 `warn` 级别
4. **缓存**: K8s 客户端已实现缓存，避免重复创建连接

## 下一步

1. 阅读 [K8S_API_IMPLEMENTATION.md](./K8S_API_IMPLEMENTATION.md) 了解详细实现
2. 查看 [Common Package README](../common/README.md) 了解公共包功能
3. 参考 [Logger Migration Guide](../common/LOGGER_MIGRATION.md) 了解日志使用

## 相关文档

- [K8S_API_IMPLEMENTATION.md](./K8S_API_IMPLEMENTATION.md) - 完整实现文档
- [QUICKSTART.md](./QUICKSTART.md) - 原有快速启动指南
- [Common Package](../common/README.md) - 公共包文档
