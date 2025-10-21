# 快速入门：5 分钟开始 API 优化

## 第一步：验证准备工作（1 分钟）

```bash
cd /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent/cluster-service

# 确认备份文件存在
ls -l internal/handler/k8s_api.go.backup

# 确认请求结构体文件存在
ls -l pkg/types/requests.go

# 确认示例代码存在
ls -l internal/handler/k8s_api_key_methods_optimized.go
```

## 第二步：修改第一个接口（2 分钟）

以 `GetCluster` 为例：

### 1. 打开原文件

```bash
vi internal/handler/k8s_api.go
```

### 2. 找到 GetCluster 方法（约 118 行）

```go
// 原代码（118-136 行）
func (h *K8sAPIHandler) GetCluster(c *gin.Context) {
	clusterID := c.Param("clusterId")

	if err := validator.ValidateClusterID(clusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	logger.Infow("Getting cluster details", "cluster_id", clusterID)

	cluster, err := h.clusterService.GetCluster(c.Request.Context(), clusterID)
	if err != nil {
		logger.Errorw("Failed to get cluster", "cluster_id", clusterID, "error", err.Error())
		response.NotFound(c, "Cluster not found", err)
		return
	}

	response.Success(c, cluster)
}
```

### 3. 替换为优化后的代码

```go
// 优化后代码
func (h *K8sAPIHandler) GetCluster(c *gin.Context) {
	var req types.GetClusterRequest

	// 绑定路径参数
	if err := c.ShouldBindUri(&req); err != nil {
		response.BadRequest(c, "Invalid request parameters", err)
		return
	}

	if err := validator.ValidateClusterID(req.ClusterID); err != nil {
		response.BadRequest(c, "Invalid cluster ID", err)
		return
	}

	logger.Infow("Getting cluster details", "cluster_id", req.ClusterID)

	cluster, err := h.clusterService.GetCluster(c.Request.Context(), req.ClusterID)
	if err != nil {
		logger.Errorw("Failed to get cluster", "cluster_id", req.ClusterID, "error", err.Error())
		response.NotFound(c, "Cluster not found", err)
		return
	}

	response.Success(c, cluster)
}
```

### 关键变化：

1. 添加了结构体定义：`var req types.GetClusterRequest`
2. 使用 `c.ShouldBindUri(&req)` 替代 `c.Param("clusterId")`
3. 所有 `clusterID` 改为 `req.ClusterID`

## 第三步：测试修改（2 分钟）

### 1. 编译检查

```bash
cd /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent/cluster-service
go build ./internal/handler/...
```

### 2. 运行单元测试

```bash
go test ./internal/handler/... -v
```

### 3. 启动服务测试

```bash
# 终端 1：启动服务
make run

# 终端 2：测试 API
curl -X GET http://localhost:8080/api/k8s/clusters/test-cluster-1
```

## 快速批量修改指南

### 模式 1：仅路径参数的接口

适用于：GetCluster, DeleteCluster, GetNamespace, DeleteNamespace 等

**修改步骤**：

1. 添加结构体声明
2. 替换 `c.Param()` 为 `c.ShouldBindUri(&req)`
3. 替换所有变量名为 `req.FieldName`

**示例对照**：

```go
// 修改前
func (h *K8sAPIHandler) GetXXX(c *gin.Context) {
    clusterID := c.Param("clusterId")
    namespace := c.Param("namespace")
    name := c.Param("name")

    // 使用 clusterID, namespace, name...
}

// 修改后
func (h *K8sAPIHandler) GetXXX(c *gin.Context) {
    var req types.GetXXXRequest

    if err := c.ShouldBindUri(&req); err != nil {
        response.BadRequest(c, "Invalid request parameters", err)
        return
    }

    // 使用 req.ClusterID, req.Namespace, req.Name...
}
```

### 模式 2：路径参数 + 请求体参数

适用于：UpdateCluster, CreateNamespace, ScaleDeployment 等

**修改步骤**：

1. 删除内联的 `struct` 定义
2. 使用预定义的 `types.XXXRequest`
3. 添加路径参数绑定（如果有）
4. 保留请求体参数绑定

**示例对照**：

```go
// 修改前
func (h *K8sAPIHandler) UpdateXXX(c *gin.Context) {
    clusterID := c.Param("clusterId")

    var req struct {
        Name string `json:"name"`
        Desc string `json:"description"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        // ...
    }

    // 使用 clusterID, req.Name, req.Desc...
}

// 修改后
func (h *K8sAPIHandler) UpdateXXX(c *gin.Context) {
    var req types.UpdateXXXRequest

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

    // 使用 req.ClusterID, req.Name, req.Desc...
}
```

### 模式 3：路径参数 + 查询参数

适用于：GetPodLogs, ListEvents 等

**修改步骤**：

1. 添加路径参数绑定
2. 添加查询参数绑定
3. 保留默认值设置逻辑

**示例对照**：

```go
// 修改前
func (h *K8sAPIHandler) GetPodLogs(c *gin.Context) {
    clusterID := c.Param("clusterId")
    namespace := c.Param("namespace")
    podName := c.Param("name")

    container := c.Query("container")
    tailLines := c.DefaultQuery("tailLines", "100")

    // 使用变量...
}

// 修改后
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

    // 使用 req.ClusterID, req.Namespace, req.Name, req.Container...
}
```

## 一天完成核心接口优化计划

### 上午（4 小时）：修改集群和命名空间接口

```bash
# 9:00-10:00  修改集群管理（5 个）
- GetCluster
- CreateCluster
- UpdateCluster
- DeleteCluster
- GetClusterHealthStatus

# 10:00-11:00  修改命名空间管理（4 个）
- ListNamespaces
- GetNamespace
- CreateNamespace
- DeleteNamespace

# 11:00-12:00  编译测试
go build ./...
go test ./...
```

### 下午（4 小时）：修改 Pod 和 Deployment 接口

```bash
# 14:00-15:00  修改 Pod 管理（4 个）
- ListPods
- GetPod
- DeletePod
- GetPodLogs  # 重点：有查询参数

# 15:00-16:00  修改 Deployment 管理（2 个）
- ListDeployments
- GetDeployment
- ScaleDeployment  # 重点：路径+请求体参数

# 16:00-17:00  集成测试
make run
# 测试所有修改过的接口

# 17:00-18:00  前端调整（如需要）
cd k8s-agent-web/apps/web-k8s
# 更新 types.ts 和 index.ts
```

## 检查清单

每修改一个方法后，检查：

- [ ] 导入了 `types` 包（文件顶部）
- [ ] 使用正确的请求结构体
- [ ] 路径参数使用 `ShouldBindUri`
- [ ] 查询参数使用 `ShouldBindQuery`
- [ ] 请求体参数使用 `ShouldBindJSON`
- [ ] 错误信息清晰
- [ ] 移除了原来的 `c.Param()` 调用
- [ ] 代码编译通过
- [ ] 功能测试通过

## 常见问题

### Q1: 找不到 types 包？

**A**: 确保文件顶部导入了：

```go
import (
    "github.com/kart-io/k8s-agent/cluster-service/pkg/types"
    // ...
)
```

### Q2: 编译错误：undefined: types.GetClusterRequest

**A**: 确认 `pkg/types/requests.go` 文件存在并且结构体已定义。

### Q3: 参数绑定失败？

**A**: 检查路由定义中的参数名是否与结构体 `uri` 标签匹配：

```go
// 路由定义
r.GET("/api/k8s/clusters/:clusterId", handler.GetCluster)

// 结构体定义
type GetClusterRequest struct {
    ClusterID string `uri:"clusterId" binding:"required"`
    //                     ^^^^^^^^^^ 必须匹配路由中的参数名
}
```

### Q4: 如何处理可选参数？

**A**: 不使用 `binding:"required"` 标签：

```go
type GetPodLogsRequest struct {
    ClusterID string `uri:"clusterId" binding:"required"` // 必填
    Container string `form:"container"`                   // 可选
    TailLines string `form:"tailLines"`                   // 可选
}
```

## 参考文件

- **结构体定义**：`pkg/types/requests.go`
- **完整示例**：`internal/handler/k8s_api_key_methods_optimized.go`
- **详细指南**：`IMPLEMENTATION_GUIDE.md`
- **优化方案**：`API_OPTIMIZATION_GUIDE.md`

## 立即开始

```bash
# 1. 进入项目目录
cd /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent/cluster-service

# 2. 打开编辑器
vi internal/handler/k8s_api.go

# 3. 参考示例文件
vi internal/handler/k8s_api_key_methods_optimized.go

# 4. 开始修改第一个接口！
```

**祝优化顺利！** 🚀
