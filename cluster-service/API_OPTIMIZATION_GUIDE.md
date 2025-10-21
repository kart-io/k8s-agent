# API 参数优化指南

## 优化目标

将所有 API 接口改为使用结构体接收参数，避免参数缺失导致的系统错误。

## 设计原则

1. **路径参数使用 `uri` 标签**：使用 `c.ShouldBindUri()` 绑定
2. **查询参数使用 `form` 标签**：使用 `c.ShouldBindQuery()` 绑定
3. **请求体参数使用 `json` 标签**：使用 `c.ShouldBindJSON()` 绑定
4. **组合参数**：路径参数和请求体参数可组合在同一结构体中

## 修改前后对比

### 示例 1：GetCluster (仅路径参数)

#### 修改前

```go
func (h *K8sAPIHandler) GetCluster(c *gin.Context) {
	clusterID := c.Param("clusterId")  // 直接获取，无参数校验

	if err := validator.ValidateClusterID(clusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}
	// ... 业务逻辑
}
```

#### 修改后

```go
func (h *K8sAPIHandler) GetCluster(c *gin.Context) {
	var req types.GetClusterRequest

	// 自动绑定和校验路径参数
	if err := c.ShouldBindUri(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters", err)
		return
	}

	// 使用 req.ClusterID 而不是 clusterID
	cluster, err := h.clusterService.GetCluster(c.Request.Context(), req.ClusterID)
	// ... 业务逻辑
}
```

### 示例 2：GetPodLogs (路径参数 + 查询参数)

#### 修改前

```go
func (h *K8sAPIHandler) GetPodLogs(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespace := c.Param("namespace")
	podName := c.Param("name")

	// 查询参数
	container := c.Query("container")
	tailLines := c.DefaultQuery("tailLines", "100")
	follow := c.Query("follow") == "true"

	// ... 业务逻辑
}
```

#### 修改后

```go
func (h *K8sAPIHandler) GetPodLogs(c *gin.Context) {
	var req types.GetPodLogsRequest

	// 绑定路径参数
	if err := c.ShouldBindUri(&req); err != nil {
		response.BadRequest(c, "Invalid path parameters", err)
		return
	}

	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters", err)
		return
	}

	// 设置默认值
	if req.TailLines == "" {
		req.TailLines = "100"
	}

	// 使用结构体字段
	logs, err := h.podService.GetPodLogs(
		c.Request.Context(),
		req.ClusterID,
		req.Namespace,
		req.Name,
		req.Container,
		req.TailLines,
		req.Follow,
	)
	// ... 业务逻辑
}
```

### 示例 3：CreateCluster (请求体参数)

#### 修改前

```go
func (h *K8sAPIHandler) CreateCluster(c *gin.Context) {
	var req struct {
		Name        string            `json:"name" binding:"required"`
		Description string            `json:"description"`
		Endpoint    string            `json:"endpoint" binding:"required"`
		KubeConfig  string            `json:"kubeconfig" binding:"required"`
		Region      string            `json:"region"`
		Provider    string            `json:"provider"`
		Labels      map[string]string `json:"labels"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}
	// ... 业务逻辑
}
```

#### 修改后

```go
func (h *K8sAPIHandler) CreateCluster(c *gin.Context) {
	var req types.CreateClusterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}

	// 验证集群名称
	if err := validator.ValidateK8sName(req.Name); err != nil {
		response.BadRequest(c, "Invalid cluster name", err)
		return
	}

	cluster, err := h.clusterService.CreateCluster(
		c.Request.Context(),
		req.Name,
		req.Description,
		req.Endpoint,
		req.KubeConfig,
		req.Region,
		req.Provider,
		req.Labels,
	)
	// ... 业务逻辑
}
```

### 示例 4：UpdateCluster (路径参数 + 请求体参数)

#### 修改前

```go
func (h *K8sAPIHandler) UpdateCluster(c *gin.Context) {
	clusterID := c.Param("clusterId")

	if err := validator.ValidateClusterID(clusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	var req struct {
		Name        string            `json:"name"`
		Description string            `json:"description"`
		Labels      map[string]string `json:"labels"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}
	// ... 业务逻辑
}
```

#### 修改后

```go
func (h *K8sAPIHandler) UpdateCluster(c *gin.Context) {
	var req types.UpdateClusterRequest

	// 绑定路径参数
	if err := c.ShouldBindUri(&req); err != nil {
		response.BadRequest(c, "Invalid path parameters", err)
		return
	}

	// 绑定请求体参数
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}

	cluster, err := h.clusterService.UpdateCluster(
		c.Request.Context(),
		req.ClusterID,
		req.Name,
		req.Description,
		req.Labels,
	)
	// ... 业务逻辑
}
```

### 示例 5：ScaleDeployment (路径参数 + 请求体参数)

#### 修改前

```go
func (h *K8sAPIHandler) ScaleDeployment(c *gin.Context) {
	clusterID := c.Param("clusterId")
	namespace := c.Param("namespace")
	deploymentName := c.Param("name")

	var req struct {
		Replicas int32 `json:"replicas" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}
	// ... 业务逻辑
}
```

#### 修改后

```go
func (h *K8sAPIHandler) ScaleDeployment(c *gin.Context) {
	var req types.ScaleDeploymentRequest

	// 绑定路径参数
	if err := c.ShouldBindUri(&req); err != nil {
		response.BadRequest(c, "Invalid path parameters", err)
		return
	}

	// 绑定请求体参数
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body", err)
		return
	}

	err := h.deploymentService.ScaleDeployment(
		c.Request.Context(),
		req.ClusterID,
		req.Namespace,
		req.Name,
		req.Replicas,
	)
	// ... 业务逻辑
}
```

## 优化收益

1. **类型安全**：所有参数通过结构体定义，编译时即可发现类型错误
2. **参数校验**：使用 `binding` 标签自动校验必填参数
3. **代码复用**：请求结构体可在文档生成、测试等场景复用
4. **错误提示**：Gin 自动生成详细的参数校验错误信息
5. **可维护性**：集中管理所有请求结构体定义

## 实施步骤

### 步骤 1：定义请求结构体

已创建 `pkg/types/requests.go` 文件，包含所有 API 的请求结构体定义。

### 步骤 2：修改 Handler 方法

逐个修改 `internal/handler/k8s_api.go` 中的 Handler 方法，使用新的请求结构体。

### 步骤 3：更新前端调用

修改 `k8s-agent-web/apps/web-k8s` 中的 API 调用，确保传递的参数符合新的结构体定义。

### 步骤 4：测试验证

运行测试用例，确保所有 API 调用正常工作。

## 前端修改示例

### GetPodLogs API 调用修改

#### 修改前

```typescript
// 可能会漏传参数，导致后端错误
const logs = await api.get(`/api/k8s/clusters/${clusterId}/namespaces/${namespace}/pods/${podName}/logs`);
```

#### 修改后

```typescript
// 使用查询参数对象，清晰且不易遗漏
const logs = await api.get(`/api/k8s/clusters/${clusterId}/namespaces/${namespace}/pods/${podName}/logs`, {
  params: {
    container: containerName,
    tailLines: '100',
    follow: false
  }
});
```

### CreateCluster API 调用修改

#### 修改前

```typescript
// 参数可能不完整
const cluster = await api.post('/api/k8s/clusters', {
  name: clusterName,
  endpoint: endpoint,
  // kubeconfig 可能被遗漏
});
```

#### 修改后

```typescript
// 使用完整的请求对象
const request: CreateClusterRequest = {
  name: clusterName,
  description: description,
  endpoint: endpoint,
  kubeconfig: kubeconfigContent,  // 必填，不会遗漏
  region: region,
  provider: provider,
  labels: labels
};

const cluster = await api.post('/api/k8s/clusters', request);
```

## 注意事项

1. **路径参数名称匹配**：`uri` 标签中的名称必须与路由定义中的参数名一致
2. **绑定顺序**：先绑定路径参数（`ShouldBindUri`），再绑定查询参数或请求体参数
3. **默认值处理**：查询参数的默认值需要在绑定后手动设置
4. **错误处理**：统一使用 `response.BadRequest` 返回参数错误

## 迁移清单

- [ ] 集群管理接口（6个）
  - [ ] ListClusters
  - [ ] GetCluster
  - [ ] GetClusterOptions
  - [ ] CreateCluster
  - [ ] UpdateCluster
  - [ ] DeleteCluster
  - [ ] GetClusterHealthStatus

- [ ] 命名空间管理接口（4个）
  - [ ] ListNamespaces
  - [ ] GetNamespace
  - [ ] CreateNamespace
  - [ ] DeleteNamespace

- [ ] Pod 管理接口（4个）
  - [ ] ListPods
  - [ ] GetPod
  - [ ] DeletePod
  - [ ] GetPodLogs

- [ ] Deployment 管理接口（4个）
  - [ ] ListDeployments
  - [ ] GetDeployment
  - [ ] ScaleDeployment
  - [ ] RestartDeployment

- [ ] Node 管理接口（5个）
  - [ ] ListNodes
  - [ ] GetNode
  - [ ] CordonNode
  - [ ] UncordonNode
  - [ ] DrainNode

- [ ] Service 管理接口（5个）
  - [ ] ListServices
  - [ ] GetService
  - [ ] CreateService
  - [ ] UpdateService
  - [ ] DeleteService

- [ ] StatefulSet 管理接口（5个）
  - [ ] ListStatefulSets
  - [ ] GetStatefulSet
  - [ ] ScaleStatefulSet
  - [ ] RestartStatefulSet
  - [ ] DeleteStatefulSet

- [ ] DaemonSet 管理接口（4个）
  - [ ] ListDaemonSets
  - [ ] GetDaemonSet
  - [ ] RestartDaemonSet
  - [ ] DeleteDaemonSet

- [ ] ConfigMap 管理接口（5个）
  - [ ] ListConfigMaps
  - [ ] GetConfigMap
  - [ ] CreateConfigMap
  - [ ] UpdateConfigMap
  - [ ] DeleteConfigMap

- [ ] Secret 管理接口（5个）
  - [ ] ListSecrets
  - [ ] GetSecret
  - [ ] CreateSecret
  - [ ] UpdateSecret
  - [ ] DeleteSecret

- [ ] 其他资源管理接口（30+个）
  - [ ] Endpoint (3个)
  - [ ] PVC (3个)
  - [ ] PV (3个)
  - [ ] EndpointSlice (3个)
  - [ ] HPA (3个)
  - [ ] Event (2个)
  - [ ] RoleBinding (3个)
  - [ ] ClusterRole (3个)
  - [ ] PriorityClass (3个)
  - [ ] Role (3个)
  - [ ] StorageClass (3个)

## 总结

本次优化将使 API 接口更加健壮和易维护，通过结构体参数绑定可以：

1. 避免参数遗漏导致的运行时错误
2. 提供更好的类型安全保障
3. 简化参数校验逻辑
4. 提升代码可读性和可维护性
