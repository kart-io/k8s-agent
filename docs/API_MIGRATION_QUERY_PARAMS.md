# K8s Agent API 迁移指南 - 查询参数风格

**迁移日期**: 2025-10-21
**影响范围**: 所有 `/api/k8s` 路径下的 API 端点

## 概述

本次迁移将 K8s Agent 的 API 从**路径参数风格**改为**查询参数风格**,以提供更扁平化的 URL 结构和更灵活的参数传递方式。

### 主要变更

- **URL 结构扁平化**: 移除嵌套的路由层级
- **参数传递方式**: 从路径参数改为查询参数
- **资源命名**: 单一资源使用单数形式,列表资源使用复数形式

## API 对比表

### 集群管理 API

| 操作 | 旧 API (路径参数) | 新 API (查询参数) |
|------|------------------|------------------|
| 列出集群 | `GET /api/k8s/clusters` | `GET /api/k8s/clusters` |
| 获取集群 | `GET /api/k8s/clusters/:clusterId` | `GET /api/k8s/cluster?clusterId=xxx` |
| 更新集群 | `PUT /api/k8s/clusters/:clusterId` | `PUT /api/k8s/cluster?clusterId=xxx` |
| 删除集群 | `DELETE /api/k8s/clusters/:clusterId` | `DELETE /api/k8s/cluster?clusterId=xxx` |
| 集群健康 | `GET /api/k8s/clusters/:clusterId/health` | `GET /api/k8s/cluster/health?clusterId=xxx` |

### 命名空间管理 API

| 操作 | 旧 API (路径参数) | 新 API (查询参数) |
|------|------------------|------------------|
| 列出命名空间 | `GET /api/k8s/clusters/:clusterId/namespaces` | `GET /api/k8s/namespaces?clusterId=xxx` |
| 创建命名空间 | `POST /api/k8s/clusters/:clusterId/namespaces` | `POST /api/k8s/namespaces?clusterId=xxx` |
| 获取命名空间 | `GET /api/k8s/clusters/:clusterId/ns/:namespace` | `GET /api/k8s/namespace?clusterId=xxx&namespace=default` |
| 删除命名空间 | `DELETE /api/k8s/clusters/:clusterId/ns/:namespace` | `DELETE /api/k8s/namespace?clusterId=xxx&namespace=default` |

### Pod 管理 API

| 操作 | 旧 API (路径参数) | 新 API (查询参数) |
|------|------------------|------------------|
| 列出 Pods | `GET /api/k8s/clusters/:clusterId/namespaces/:namespace/pods` | `GET /api/k8s/pods?clusterId=xxx&namespace=default` |
| 获取 Pod | `GET /api/k8s/clusters/:clusterId/namespaces/:namespace/pods/:name` | `GET /api/k8s/pod?clusterId=xxx&namespace=default&name=pod-1` |
| 删除 Pod | `DELETE /api/k8s/clusters/:clusterId/namespaces/:namespace/pods/:name` | `DELETE /api/k8s/pod?clusterId=xxx&namespace=default&name=pod-1` |
| Pod 日志 | `GET /api/k8s/clusters/:clusterId/namespaces/:namespace/pods/:name/logs?container=xxx` | `GET /api/k8s/pod/logs?clusterId=xxx&namespace=default&name=pod-1&container=app` |

### Deployment 管理 API

| 操作 | 旧 API (路径参数) | 新 API (查询参数) |
|------|------------------|------------------|
| 列出 Deployments | `GET /api/k8s/clusters/:clusterId/namespaces/:namespace/deployments` | `GET /api/k8s/deployments?clusterId=xxx&namespace=default` |
| 获取 Deployment | `GET /api/k8s/clusters/:clusterId/namespaces/:namespace/deployments/:name` | `GET /api/k8s/deployment?clusterId=xxx&namespace=default&name=app` |
| 扩容 Deployment | `PUT /api/k8s/clusters/:clusterId/namespaces/:namespace/deployments/:name/scale` | `PUT /api/k8s/deployment/scale?clusterId=xxx&namespace=default&name=app` |
| 重启 Deployment | `POST /api/k8s/clusters/:clusterId/namespaces/:namespace/deployments/:name/restart` | `POST /api/k8s/deployment/restart?clusterId=xxx&namespace=default&name=app` |

### Node 管理 API

| 操作 | 旧 API (路径参数) | 新 API (查询参数) |
|------|------------------|------------------|
| 列出 Nodes | `GET /api/k8s/clusters/:clusterId/nodes` | `GET /api/k8s/nodes?clusterId=xxx` |
| 获取 Node | `GET /api/k8s/clusters/:clusterId/nodes/:name` | `GET /api/k8s/node?clusterId=xxx&name=node-1` |
| Cordon Node | `POST /api/k8s/clusters/:clusterId/nodes/:name/cordon` | `POST /api/k8s/node/cordon?clusterId=xxx&name=node-1` |
| Uncordon Node | `POST /api/k8s/clusters/:clusterId/nodes/:name/uncordon` | `POST /api/k8s/node/uncordon?clusterId=xxx&name=node-1` |
| Drain Node | `POST /api/k8s/clusters/:clusterId/nodes/:name/drain` | `POST /api/k8s/node/drain?clusterId=xxx&name=node-1` |

### Service 管理 API

| 操作 | 旧 API (路径参数) | 新 API (查询参数) |
|------|------------------|------------------|
| 列出 Services | `GET /api/k8s/clusters/:clusterId/namespaces/:namespace/services` | `GET /api/k8s/services?clusterId=xxx&namespace=default` |
| 创建 Service | `POST /api/k8s/clusters/:clusterId/namespaces/:namespace/services` | `POST /api/k8s/services?clusterId=xxx&namespace=default` |
| 获取 Service | `GET /api/k8s/clusters/:clusterId/namespaces/:namespace/services/:name` | `GET /api/k8s/service?clusterId=xxx&namespace=default&name=svc-1` |
| 更新 Service | `PUT /api/k8s/clusters/:clusterId/namespaces/:namespace/services/:name` | `PUT /api/k8s/service?clusterId=xxx&namespace=default&name=svc-1` |
| 删除 Service | `DELETE /api/k8s/clusters/:clusterId/namespaces/:namespace/services/:name` | `DELETE /api/k8s/service?clusterId=xxx&namespace=default&name=svc-1` |

### 其他资源 API

以下资源遵循相同的迁移模式:

- **StatefulSet**: `/statefulsets` → `/statefulset`
- **DaemonSet**: `/daemonsets` → `/daemonset`
- **ConfigMap**: `/configmaps` → `/configmap`
- **Secret**: `/secrets` → `/secret`
- **Endpoint**: `/endpoints` → `/endpoint`
- **PVC**: `/pvcs` → `/pvc`
- **PV**: `/pvs` → `/pv`
- **EndpointSlice**: `/endpointslices` → `/endpointslice`
- **HPA**: `/hpas` → `/hpa`
- **Event**: `/events` → `/event`
- **RoleBinding**: `/rolebindings` → `/rolebinding`
- **ClusterRole**: `/clusterroles` → `/clusterrole`
- **PriorityClass**: `/priorityclasses` → `/priorityclass`
- **Role**: `/roles` → `/role`
- **StorageClass**: `/storageclasses` → `/storageclass`

## 客户端迁移指南

### 1. URL 构建方式变更

**旧方式 (路径参数)**:

```javascript
// JavaScript/TypeScript 示例
const clusterId = 'cluster-123';
const namespace = 'default';
const podName = 'my-pod';

const url = `/api/k8s/clusters/${clusterId}/namespaces/${namespace}/pods/${podName}`;
```

**新方式 (查询参数)**:

```javascript
// JavaScript/TypeScript 示例
const params = new URLSearchParams({
  clusterId: 'cluster-123',
  namespace: 'default',
  name: 'my-pod'
});

const url = `/api/k8s/pod?${params.toString()}`;
```

### 2. Go 客户端示例

**旧方式**:

```go
import "fmt"

clusterId := "cluster-123"
namespace := "default"
podName := "my-pod"

url := fmt.Sprintf("/api/k8s/clusters/%s/namespaces/%s/pods/%s",
    clusterId, namespace, podName)
```

**新方式**:

```go
import (
    "fmt"
    "net/url"
)

params := url.Values{}
params.Add("clusterId", "cluster-123")
params.Add("namespace", "default")
params.Add("name", "my-pod")

url := fmt.Sprintf("/api/k8s/pod?%s", params.Encode())
```

### 3. cURL 示例

**旧方式**:

```bash
# 获取 Pod 详情
curl -X GET http://localhost:8080/api/k8s/clusters/cluster-123/namespaces/default/pods/my-pod

# 获取 Pod 日志
curl -X GET "http://localhost:8080/api/k8s/clusters/cluster-123/namespaces/default/pods/my-pod/logs?container=app"
```

**新方式**:

```bash
# 获取 Pod 详情
curl -X GET "http://localhost:8080/api/k8s/pod?clusterId=cluster-123&namespace=default&name=my-pod"

# 获取 Pod 日志
curl -X GET "http://localhost:8080/api/k8s/pod/logs?clusterId=cluster-123&namespace=default&name=my-pod&container=app"
```

### 4. Python 客户端示例

**旧方式**:

```python
import requests

cluster_id = "cluster-123"
namespace = "default"
pod_name = "my-pod"

url = f"/api/k8s/clusters/{cluster_id}/namespaces/{namespace}/pods/{pod_name}"
response = requests.get(f"http://localhost:8080{url}")
```

**新方式**:

```python
import requests

params = {
    "clusterId": "cluster-123",
    "namespace": "default",
    "name": "my-pod"
}

response = requests.get(
    "http://localhost:8080/api/k8s/pod",
    params=params
)
```

## 错误处理

### 查询参数缺失

新 API 使用查询参数绑定,如果必需参数缺失,将返回:

```json
{
  "code": 400,
  "message": "Invalid query parameters",
  "error": "Key: 'GetPodRequest.ClusterID' Error:Field validation for 'ClusterID' failed on the 'required' tag"
}
```

**客户端应确保**:

- 所有必需的查询参数都已提供
- 参数名称使用 `camelCase` 格式 (`clusterId`, `namespace`, `name`)
- 特殊字符需要 URL 编码

### URL 编码注意事项

某些资源名称可能包含特殊字符,需要正确的 URL 编码:

```javascript
// 正确的做法
const namespace = "kube-system";
const podName = "coredns-abc-123";

const params = new URLSearchParams({
  clusterId: "cluster-123",
  namespace: namespace,  // URLSearchParams 自动编码
  name: podName
});
```

```go
// Go 中使用 url.Values 自动处理编码
params := url.Values{}
params.Add("namespace", "kube-system")
params.Add("name", "coredns-abc-123")
```

## 迁移步骤

### 对于 API 客户端开发者

1. **更新 URL 构建逻辑**
   - 将所有路径参数改为查询参数
   - 使用各语言的标准 URL 编码工具

2. **测试所有 API 端点**
   - 验证查询参数正确传递
   - 测试错误处理(参数缺失、参数错误)

3. **更新文档和示例代码**
   - 修改 API 调用示例
   - 更新集成测试

### 对于运维人员

1. **无需服务器配置变更**
   - 路由已在应用层处理
   - 无需修改反向代理配置

2. **监控新旧 API 调用**
   - 旧 API (路径参数) 已被移除
   - 仅支持新 API (查询参数)

## 向后兼容性

**注意**: 本次迁移**不保持向后兼容**。

- 旧的路径参数风格 API 已被完全替换
- 客户端必须更新以使用新的查询参数风格
- 建议在生产环境部署前完成所有客户端更新

## 技术实现细节

### 请求类型变更

所有请求结构体中的标签已从 `uri` 改为 `form`:

```go
// 旧定义
type GetPodRequest struct {
    ClusterID string `uri:"clusterId" binding:"required"`
    Namespace string `uri:"namespace" binding:"required"`
    Name      string `uri:"name" binding:"required"`
}

// 新定义
type GetPodRequest struct {
    ClusterID string `form:"clusterId" binding:"required"`
    Namespace string `form:"namespace" binding:"required"`
    Name      string `form:"name" binding:"required"`
}
```

### Handler 绑定方法变更

所有 Handler 中的参数绑定从 `ShouldBindUri` 改为 `ShouldBindQuery`:

```go
// 旧实现
if err := c.ShouldBindUri(&req); err != nil {
    response.BadRequest(c, "Invalid request parameters", err)
    return
}

// 新实现
if err := c.ShouldBindQuery(&req); err != nil {
    response.BadRequest(c, "Invalid query parameters", err)
    return
}
```

### 直接参数提取变更

某些 Handler 中使用 `c.Param()` 直接提取参数的代码已改为 `c.Query()`:

```go
// 旧实现
clusterID := c.Param("clusterId")
namespace := c.Param("namespace")
nodeName := c.Param("name")

// 新实现
clusterID := c.Query("clusterId")
namespace := c.Query("namespace")
nodeName := c.Query("name")
```

## 常见问题 (FAQ)

### Q1: 为什么要从路径参数改为查询参数?

**A**: 主要原因包括:

- **URL 扁平化**: 查询参数提供更简洁的 URL 结构
- **参数灵活性**: 更容易支持可选参数和参数组合
- **缓存友好**: 查询参数风格更适合 HTTP 缓存机制
- **日志清晰**: 所有参数都在查询字符串中,日志记录更清晰

### Q2: 查询参数是否需要 URL 编码?

**A**: 是的,特殊字符必须进行 URL 编码:

- 使用标准库的 URL 编码工具 (`URLSearchParams`, `url.Values`, etc.)
- 避免手动拼接查询字符串

### Q3: 如何处理包含特殊字符的资源名称?

**A**: 使用 URL 编码工具自动处理:

```javascript
// 资源名称包含 '/' 或 '=' 等特殊字符
const podName = "app/v1=latest";

const params = new URLSearchParams({
  name: podName  // 自动编码为 app%2Fv1%3Dlatest
});
```

### Q4: 列表接口是否也使用查询参数?

**A**: 是的,列表接口也使用查询参数传递过滤条件:

```bash
# 列出指定命名空间的 Pods
GET /api/k8s/pods?clusterId=cluster-123&namespace=default

# 列出所有 Nodes
GET /api/k8s/nodes?clusterId=cluster-123
```

### Q5: 分页参数如何传递?

**A**: 分页参数通过查询参数传递:

```bash
# 第 2 页,每页 20 条
GET /api/k8s/pods?clusterId=cluster-123&namespace=default&page=2&pageSize=20
```

## 测试示例

### 完整测试脚本 (Bash)

```bash
#!/bin/bash

BASE_URL="http://localhost:8080"
CLUSTER_ID="cluster-123"
NAMESPACE="default"
POD_NAME="my-pod"

echo "=== 测试集群 API ==="
curl -X GET "${BASE_URL}/api/k8s/clusters"
curl -X GET "${BASE_URL}/api/k8s/cluster?clusterId=${CLUSTER_ID}"

echo "=== 测试命名空间 API ==="
curl -X GET "${BASE_URL}/api/k8s/namespaces?clusterId=${CLUSTER_ID}"
curl -X GET "${BASE_URL}/api/k8s/namespace?clusterId=${CLUSTER_ID}&namespace=${NAMESPACE}"

echo "=== 测试 Pod API ==="
curl -X GET "${BASE_URL}/api/k8s/pods?clusterId=${CLUSTER_ID}&namespace=${NAMESPACE}"
curl -X GET "${BASE_URL}/api/k8s/pod?clusterId=${CLUSTER_ID}&namespace=${NAMESPACE}&name=${POD_NAME}"

echo "=== 测试 Pod 日志 API ==="
curl -X GET "${BASE_URL}/api/k8s/pod/logs?clusterId=${CLUSTER_ID}&namespace=${NAMESPACE}&name=${POD_NAME}&container=app&tailLines=100"
```

### Go 集成测试示例

```go
package api_test

import (
    "net/http"
    "net/http/httptest"
    "net/url"
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestGetPodQueryParams(t *testing.T) {
    // 准备测试服务器
    server := setupTestServer()
    defer server.Close()

    // 构建查询参数
    params := url.Values{}
    params.Add("clusterId", "cluster-123")
    params.Add("namespace", "default")
    params.Add("name", "my-pod")

    // 发送请求
    url := server.URL + "/api/k8s/pod?" + params.Encode()
    resp, err := http.Get(url)

    // 验证
    assert.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}
```

## 附录

### 完整的 API 端点列表

#### 集群管理

- `GET /api/k8s/clusters` - 列出所有集群
- `GET /api/k8s/clusters/options` - 获取集群选择器
- `POST /api/k8s/clusters` - 创建集群
- `GET /api/k8s/cluster?clusterId=xxx` - 获取集群详情
- `PUT /api/k8s/cluster?clusterId=xxx` - 更新集群
- `DELETE /api/k8s/cluster?clusterId=xxx` - 删除集群
- `GET /api/k8s/cluster/health?clusterId=xxx` - 集群健康状态

#### 命名空间管理

- `GET /api/k8s/namespaces?clusterId=xxx` - 列出命名空间
- `POST /api/k8s/namespaces?clusterId=xxx` - 创建命名空间
- `GET /api/k8s/namespace?clusterId=xxx&namespace=xxx` - 获取命名空间
- `DELETE /api/k8s/namespace?clusterId=xxx&namespace=xxx` - 删除命名空间

#### Pod 管理

- `GET /api/k8s/pods?clusterId=xxx&namespace=xxx` - 列出 Pods
- `GET /api/k8s/pod?clusterId=xxx&namespace=xxx&name=xxx` - 获取 Pod
- `DELETE /api/k8s/pod?clusterId=xxx&namespace=xxx&name=xxx` - 删除 Pod
- `GET /api/k8s/pod/logs?clusterId=xxx&namespace=xxx&name=xxx&container=xxx` - Pod 日志

#### Deployment 管理

- `GET /api/k8s/deployments?clusterId=xxx&namespace=xxx` - 列出 Deployments
- `GET /api/k8s/deployment?clusterId=xxx&namespace=xxx&name=xxx` - 获取 Deployment
- `PUT /api/k8s/deployment/scale?clusterId=xxx&namespace=xxx&name=xxx` - 扩容
- `POST /api/k8s/deployment/restart?clusterId=xxx&namespace=xxx&name=xxx` - 重启

#### Node 管理

- `GET /api/k8s/nodes?clusterId=xxx` - 列出 Nodes
- `GET /api/k8s/node?clusterId=xxx&name=xxx` - 获取 Node
- `POST /api/k8s/node/cordon?clusterId=xxx&name=xxx` - Cordon
- `POST /api/k8s/node/uncordon?clusterId=xxx&name=xxx` - Uncordon
- `POST /api/k8s/node/drain?clusterId=xxx&name=xxx` - Drain

*(其他资源端点省略,遵循相同模式)*

## 修改历史

- **2025-10-21**: 初始版本,完成从路径参数到查询参数的全面迁移

## 联系方式

如有问题或需要支持,请联系:

- **开发团队**: k8s-agent-dev@example.com
- **GitHub Issues**: https://github.com/kart-io/k8s-agent/issues
