# K8s Agent - 查询参数风格 API 重构

## 概述

本次重构将 K8s Agent 的所有 `/api/k8s` 端点从**路径参数风格**迁移到**查询参数风格**,实现更扁平化和灵活的 API 设计。

## 快速开始

### 核心变化

**旧 API**:
```
GET /api/k8s/clusters/:clusterId/namespaces/:namespace/pods/:name
```

**新 API**:
```
GET /api/k8s/pod?clusterId=xxx&namespace=default&name=pod-1
```

### 编译和运行

```bash
# 1. 进入项目目录
cd cluster-service

# 2. 安装依赖
go mod tidy

# 3. 编译
go build -o bin/cluster-service ./cmd/server

# 4. 运行
./bin/cluster-service
```

### 测试 API

```bash
# 运行测试脚本
./scripts/test_query_params_api.sh

# 或手动测试
curl -X GET "http://localhost:8080/api/k8s/clusters"
curl -X GET "http://localhost:8080/api/k8s/cluster?clusterId=test-cluster"
```

## 目录结构

```
cluster-service/
├── cmd/server/              # 服务入口
├── internal/
│   ├── api/
│   │   └── server.go        # ✓ 路由注册 (已重构为查询参数)
│   └── handler/
│       └── k8s_api.go       # ✓ Handler (已更新绑定方式)
├── pkg/types/
│   └── requests.go          # ✓ 请求类型 (uri → form 标签)
├── scripts/
│   └── test_query_params_api.sh  # API 测试脚本
├── examples/client/
│   ├── go_client.go         # Go 客户端示例
│   └── python_client.py     # Python 客户端示例
└── docs/
    ├── API_MIGRATION_QUERY_PARAMS.md  # 完整迁移文档
    └── API_QUICK_REFERENCE.md         # 快速对照表
```

## 修改清单

### 1. 请求类型重构 (`pkg/types/requests.go`)

- ✅ 修改了 **60+ 个结构体**
- ✅ 所有 `uri:"..."` 标签改为 `form:"..."`
- ✅ 覆盖所有资源类型

**示例**:
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

### 2. 路由扁平化 (`internal/api/server.go`)

- ✅ 重写了 `setupK8sAPIRoutes()` 函数
- ✅ 移除了所有嵌套的 `.Group()`
- ✅ **100+ 个端点**改为扁平化路由

**示例**:
```go
// 旧路由
pods := k8sAPI.Group("/clusters/:clusterId/namespaces/:namespace/pods")
{
    pods.GET("/:name", handler.GetPod)
}

// 新路由
k8sAPI.GET("/pod", handler.GetPod)  // ?clusterId=xxx&namespace=xxx&name=xxx
```

### 3. Handler 更新 (`internal/handler/k8s_api.go`)

- ✅ **16 处** `ShouldBindUri` → `ShouldBindQuery`
- ✅ **146 处** `c.Param()` → `c.Query()`
- ✅ 错误消息更新

**示例**:
```go
// 旧实现
if err := c.ShouldBindUri(&req); err != nil {
    response.BadRequest(c, "Invalid request parameters", err)
    return
}
clusterID := c.Param("clusterId")

// 新实现
if err := c.ShouldBindQuery(&req); err != nil {
    response.BadRequest(c, "Invalid query parameters", err)
    return
}
clusterID := c.Query("clusterId")
```

## API 端点概览

### 集群管理 (7 endpoints)
- `GET /api/k8s/clusters` - 列出集群
- `GET /api/k8s/clusters/options` - 集群选项
- `POST /api/k8s/clusters` - 创建集群
- `GET /api/k8s/cluster?clusterId=xxx` - 获取集群
- `PUT /api/k8s/cluster?clusterId=xxx` - 更新集群
- `DELETE /api/k8s/cluster?clusterId=xxx` - 删除集群
- `GET /api/k8s/cluster/health?clusterId=xxx` - 健康状态

### 命名空间管理 (4 endpoints)
- `GET /api/k8s/namespaces?clusterId=xxx`
- `POST /api/k8s/namespaces?clusterId=xxx`
- `GET /api/k8s/namespace?clusterId=xxx&namespace=xxx`
- `DELETE /api/k8s/namespace?clusterId=xxx&namespace=xxx`

### Pod 管理 (4 endpoints)
- `GET /api/k8s/pods?clusterId=xxx&namespace=xxx`
- `GET /api/k8s/pod?clusterId=xxx&namespace=xxx&name=xxx`
- `DELETE /api/k8s/pod?clusterId=xxx&namespace=xxx&name=xxx`
- `GET /api/k8s/pod/logs?clusterId=xxx&namespace=xxx&name=xxx&container=xxx`

**其他资源**: Deployment, Node, Service, StatefulSet, DaemonSet, ConfigMap, Secret, PVC, PV, HPA, Event, RBAC 等 (遵循相同模式)

## 客户端使用

### Go 客户端

```go
// 查看完整示例: examples/client/go_client.go
client := NewK8sClient("http://localhost:8080")

// 列出集群
clusters, err := client.ListClusters(1, 20)

// 获取 Pod (查询参数)
pod, err := client.GetPod("cluster-123", "default", "my-pod")

// 获取 Pod 日志
logs, err := client.GetPodLogs("cluster-123", "default", "my-pod", "app", 100)
```

### Python 客户端

```python
# 查看完整示例: examples/client/python_client.py
from python_client import K8sClient

client = K8sClient("http://localhost:8080")

# 列出集群
clusters = client.list_clusters(page=1, page_size=20)

# 获取 Pod (查询参数)
pod = client.get_pod("cluster-123", "default", "my-pod")

# 获取 Pod 日志
logs = client.get_pod_logs("cluster-123", "default", "my-pod",
                           container="app", tail_lines=100)
```

### cURL

```bash
# 获取集群 (查询参数)
curl -X GET "http://localhost:8080/api/k8s/cluster?clusterId=cluster-123"

# 列出 Pods (查询参数)
curl -X GET "http://localhost:8080/api/k8s/pods?clusterId=cluster-123&namespace=default"

# 获取 Pod 日志 (查询参数)
curl -X GET "http://localhost:8080/api/k8s/pod/logs?clusterId=cluster-123&namespace=default&name=my-pod&container=app&tailLines=100"
```

## 测试

### 自动化测试

```bash
# 运行完整测试套件
./scripts/test_query_params_api.sh

# 测试输出示例:
# [INFO] Testing: 列出所有集群
# [PASS] HTTP 200 - 列出所有集群
# [INFO] Testing: 获取集群详情 (查询参数)
# [PASS] HTTP 200 - 获取集群详情 (查询参数)
# ...
# 总测试数: 30
# 通过: 28
# 失败: 2
```

### 手动测试

```bash
# 设置环境变量
export BASE_URL=http://localhost:8080
export CLUSTER_ID=test-cluster

# 测试集群 API
curl "${BASE_URL}/api/k8s/cluster?clusterId=${CLUSTER_ID}"

# 测试命名空间 API
curl "${BASE_URL}/api/k8s/namespaces?clusterId=${CLUSTER_ID}"

# 测试 Pod API
curl "${BASE_URL}/api/k8s/pods?clusterId=${CLUSTER_ID}&namespace=default"
```

## 重要注意事项

### ⚠️ 向后不兼容

本次重构**不保持向后兼容**:
- 旧的路径参数风格 API 已被完全移除
- 所有客户端必须更新为查询参数风格
- 建议在生产环境部署前完成客户端更新

### ✓ URL 编码

查询参数中的特殊字符需要 URL 编码:

```javascript
// ✓ 正确: 使用 URLSearchParams (自动编码)
const params = new URLSearchParams({
  namespace: "kube-system",
  name: "coredns-abc-123"
});

// ✗ 错误: 手动拼接 (缺少编码)
const url = `/api/k8s/pod?namespace=kube-system&name=coredns-abc-123`;
```

### ✓ 必需参数

所有原路径参数现在都是必需的查询参数:

```json
// 缺少参数会返回 400 错误
{
  "code": 400,
  "message": "Invalid query parameters",
  "error": "Key: 'GetPodRequest.Namespace' Error:Field validation for 'Namespace' failed on the 'required' tag"
}
```

## 文档

- **完整迁移指南**: [docs/API_MIGRATION_QUERY_PARAMS.md](docs/API_MIGRATION_QUERY_PARAMS.md)
  - 详细的新旧 API 对比
  - 多语言客户端示例
  - 错误处理指南
  - FAQ

- **快速对照表**: [docs/API_QUICK_REFERENCE.md](docs/API_QUICK_REFERENCE.md)
  - 常用 API 对比
  - 代码示例速查
  - 测试命令

## 开发

### 编译

```bash
# 开发环境编译
go build -o bin/cluster-service ./cmd/server

# 生产环境编译 (带版本信息)
make build
```

### 代码检查

```bash
# 格式化代码
go fmt ./...

# 静态检查
go vet ./...

# 运行测试
go test ./...
```

### 添加新端点

遵循查询参数模式:

1. **定义请求类型** (pkg/types/requests.go):
```go
type GetResourceRequest struct {
    ClusterID string `form:"clusterId" binding:"required"`
    Namespace string `form:"namespace" binding:"required"`
    Name      string `form:"name" binding:"required"`
}
```

2. **注册路由** (internal/api/server.go):
```go
k8sAPI.GET("/resource", s.k8sAPIHandler.GetResource)
```

3. **实现 Handler** (internal/handler/k8s_api.go):
```go
func (h *K8sAPIHandler) GetResource(c *gin.Context) {
    var req types.GetResourceRequest
    if err := c.ShouldBindQuery(&req); err != nil {
        response.BadRequest(c, "Invalid query parameters", err)
        return
    }
    // ... 业务逻辑
}
```

## 性能

### 编译产物

- **二进制大小**: ~57MB
- **编译时间**: ~30s (首次), ~5s (增量)

### 路由性能

查询参数风格相比路径参数:
- **路由查找**: 略慢 (扁平路由减少了路由树深度)
- **参数解析**: 相当 (Gin 高效处理查询参数)
- **整体影响**: 可忽略不计

## 贡献

如发现问题或有改进建议:
1. 创建 GitHub Issue
2. 提交 Pull Request
3. 联系开发团队

## 许可证

[根据项目实际情况添加]

---

**最后更新**: 2025-10-21
**版本**: v2.0.0 (查询参数风格)
