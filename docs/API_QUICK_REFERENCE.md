# API 快速对照表 - 路径参数 vs 查询参数

## 常用 API 对比

### 集群管理

| 操作 | 旧风格 (路径参数) | 新风格 (查询参数) |
|------|-----------------|------------------|
| 列出集群 | `GET /api/k8s/clusters` | `GET /api/k8s/clusters` ✓ |
| 获取集群 | `GET /api/k8s/clusters/cluster-123` | `GET /api/k8s/cluster?clusterId=cluster-123` |
| 删除集群 | `DELETE /api/k8s/clusters/cluster-123` | `DELETE /api/k8s/cluster?clusterId=cluster-123` |
| 集群健康 | `GET /api/k8s/clusters/cluster-123/health` | `GET /api/k8s/cluster/health?clusterId=cluster-123` |

### 命名空间管理

| 操作 | 旧风格 | 新风格 |
|------|--------|--------|
| 列出 | `GET /api/k8s/clusters/cluster-123/namespaces` | `GET /api/k8s/namespaces?clusterId=cluster-123` |
| 获取 | `GET /api/k8s/clusters/cluster-123/ns/default` | `GET /api/k8s/namespace?clusterId=cluster-123&namespace=default` |
| 删除 | `DELETE /api/k8s/clusters/cluster-123/ns/default` | `DELETE /api/k8s/namespace?clusterId=cluster-123&namespace=default` |

### Pod 管理

| 操作 | 旧风格 | 新风格 |
|------|--------|--------|
| 列出 | `GET /api/k8s/clusters/c1/namespaces/default/pods` | `GET /api/k8s/pods?clusterId=c1&namespace=default` |
| 获取 | `GET /api/k8s/clusters/c1/namespaces/default/pods/pod-1` | `GET /api/k8s/pod?clusterId=c1&namespace=default&name=pod-1` |
| 日志 | `GET /api/k8s/clusters/c1/namespaces/default/pods/pod-1/logs?container=app` | `GET /api/k8s/pod/logs?clusterId=c1&namespace=default&name=pod-1&container=app` |

### Deployment 管理

| 操作 | 旧风格 | 新风格 |
|------|--------|--------|
| 列出 | `GET /api/k8s/clusters/c1/namespaces/default/deployments` | `GET /api/k8s/deployments?clusterId=c1&namespace=default` |
| 获取 | `GET /api/k8s/clusters/c1/namespaces/default/deployments/app` | `GET /api/k8s/deployment?clusterId=c1&namespace=default&name=app` |
| 扩容 | `PUT /api/k8s/clusters/c1/namespaces/default/deployments/app/scale` | `PUT /api/k8s/deployment/scale?clusterId=c1&namespace=default&name=app` |

### Node 管理

| 操作 | 旧风格 | 新风格 |
|------|--------|--------|
| 列出 | `GET /api/k8s/clusters/c1/nodes` | `GET /api/k8s/nodes?clusterId=c1` |
| 获取 | `GET /api/k8s/clusters/c1/nodes/node-1` | `GET /api/k8s/node?clusterId=c1&name=node-1` |
| Cordon | `POST /api/k8s/clusters/c1/nodes/node-1/cordon` | `POST /api/k8s/node/cordon?clusterId=c1&name=node-1` |

## 客户端代码示例

### cURL

#### 旧风格
```bash
curl -X GET http://localhost:8080/api/k8s/clusters/cluster-123/namespaces/default/pods/my-pod
```

#### 新风格
```bash
curl -X GET "http://localhost:8080/api/k8s/pod?clusterId=cluster-123&namespace=default&name=my-pod"
```

### Go

#### 旧风格
```go
url := fmt.Sprintf("/api/k8s/clusters/%s/namespaces/%s/pods/%s", clusterID, namespace, podName)
```

#### 新风格
```go
params := url.Values{}
params.Add("clusterId", clusterID)
params.Add("namespace", namespace)
params.Add("name", podName)
url := "/api/k8s/pod?" + params.Encode()
```

### Python

#### 旧风格
```python
url = f"/api/k8s/clusters/{cluster_id}/namespaces/{namespace}/pods/{pod_name}"
```

#### 新风格
```python
from urllib.parse import urlencode
params = {
    "clusterId": cluster_id,
    "namespace": namespace,
    "name": pod_name
}
url = f"/api/k8s/pod?{urlencode(params)}"
```

### JavaScript/TypeScript

#### 旧风格
```javascript
const url = `/api/k8s/clusters/${clusterId}/namespaces/${namespace}/pods/${podName}`;
```

#### 新风格
```javascript
const params = new URLSearchParams({
  clusterId: 'cluster-123',
  namespace: 'default',
  name: 'my-pod'
});
const url = `/api/k8s/pod?${params.toString()}`;
```

## 关键变化总结

### ✅ 优点

1. **URL 更简洁**: 扁平化结构,路径层级更少
2. **参数灵活**: 可选参数更容易处理
3. **缓存友好**: 查询参数更适合 HTTP 缓存
4. **日志清晰**: 所有参数在查询字符串中

### ⚠️ 注意事项

1. **URL 编码**: 特殊字符必须编码
2. **参数顺序**: 虽然查询参数无顺序要求,但建议保持一致性
3. **必需参数**: 所有原路径参数变为必需的查询参数
4. **向后不兼容**: 旧 API 已被移除,必须更新客户端

## 完整资源列表

所有以下资源都遵循相同的迁移模式:

- **Cluster** (集群)
- **Namespace** (命名空间)
- **Pod** (容器组)
- **Deployment** (部署)
- **Node** (节点)
- **Service** (服务)
- **StatefulSet** (有状态集)
- **DaemonSet** (守护集)
- **ConfigMap** (配置映射)
- **Secret** (密钥)
- **Endpoint** (端点)
- **PVC** (持久卷声明)
- **PV** (持久卷)
- **EndpointSlice** (端点切片)
- **HPA** (水平自动伸缩)
- **Event** (事件)
- **RoleBinding** (角色绑定)
- **ClusterRole** (集群角色)
- **PriorityClass** (优先级类)
- **Role** (角色)
- **StorageClass** (存储类)

## 测试命令

```bash
# 使用测试脚本
./scripts/test_query_params_api.sh

# 手动测试
export BASE_URL=http://localhost:8080
export CLUSTER_ID=cluster-123

# 测试集群 API
curl -X GET "${BASE_URL}/api/k8s/cluster?clusterId=${CLUSTER_ID}"

# 测试 Pod API
curl -X GET "${BASE_URL}/api/k8s/pods?clusterId=${CLUSTER_ID}&namespace=default"
```

## 错误处理

### 缺少参数

**请求**:
```bash
curl -X GET "http://localhost:8080/api/k8s/pod?clusterId=cluster-123"
```

**响应**:
```json
{
  "code": 400,
  "message": "Invalid query parameters",
  "error": "Key: 'GetPodRequest.Namespace' Error:Field validation for 'Namespace' failed on the 'required' tag"
}
```

### 特殊字符

**正确** (自动 URL 编码):
```javascript
const params = new URLSearchParams({
  namespace: "kube-system"  // 自动编码
});
```

**错误** (手动拼接):
```javascript
const url = `/api/k8s/namespace?namespace=kube-system`;  // 可能缺少编码
```

## 更多信息

- 完整迁移文档: `docs/API_MIGRATION_QUERY_PARAMS.md`
- Go 客户端示例: `examples/client/go_client.go`
- Python 客户端示例: `examples/client/python_client.py`
- 测试脚本: `scripts/test_query_params_api.sh`
