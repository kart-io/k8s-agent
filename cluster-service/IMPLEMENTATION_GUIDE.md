# API 参数优化实施指南

## 概述

由于 `k8s_api.go` 文件有 2721 行，包含 85+ 个 Handler 方法，直接全部修改工作量较大。本指南提供分阶段实施方案。

## 实施策略

### 方案 1：渐进式迁移（推荐）

创建新的 Handler 方法，保留旧方法向后兼容，逐步迁移。

**优点**：
- 风险低，不影响现有功能
- 可以逐步迁移客户端
- 便于 A/B 测试

**缺点**：
- 需要维护两套代码一段时间
- 代码量会暂时增加

### 方案 2：一次性全量修改

直接修改所有 Handler 方法，同时修改前端调用。

**优点**：
- 一次性完成，代码整洁
- 不需要维护兼容代码

**缺点**：
- 风险较高，需要全面测试
- 必须同时部署前后端

## 推荐实施步骤（方案 1）

### 阶段 1：修改核心接口（1-2 天）

优先修改使用频率最高的接口：

1. **集群管理**（7 个方法）
   - GetCluster
   - CreateCluster
   - UpdateCluster
   - DeleteCluster
   - GetClusterHealthStatus
   - ListClusters
   - GetClusterOptions

2. **Pod 管理**（4 个方法）
   - ListPods
   - GetPod
   - DeletePod
   - GetPodLogs ⚠️ 重点（有查询参数）

3. **Deployment 管理**（4 个方法）
   - ListDeployments
   - GetDeployment
   - ScaleDeployment ⚠️ 重点（路径+请求体参数）
   - RestartDeployment

### 阶段 2：修改其他工作负载接口（1-2 天）

4. **Node 管理**（5 个方法）
5. **Service 管理**（5 个方法）
6. **StatefulSet 管理**（5 个方法）
7. **DaemonSet 管理**（4 个方法）

### 阶段 3：修改配置和存储接口（1 天）

8. **ConfigMap 管理**（5 个方法）
9. **Secret 管理**（5 个方法）
10. **PVC/PV 管理**（6 个方法）

### 阶段 4：修改其他资源接口（1 天）

11. 其他所有资源（40+ 个方法）

## 快速修改模板

### 模板 1：仅路径参数

```go
// 修改前
func (h *K8sAPIHandler) GetXXX(c *gin.Context) {
    clusterID := c.Param("clusterId")
    namespace := c.Param("namespace")
    name := c.Param("name")

    // 业务逻辑...
}

// 修改后
func (h *K8sAPIHandler) GetXXX(c *gin.Context) {
    var req types.GetXXXRequest

    if err := c.ShouldBindUri(&req); err != nil {
        response.BadRequest(c, "Invalid request parameters", err)
        return
    }

    // 使用 req.ClusterID, req.Namespace, req.Name
    // 业务逻辑...
}
```

### 模板 2：路径参数 + 请求体参数

```go
// 修改前
func (h *K8sAPIHandler) UpdateXXX(c *gin.Context) {
    clusterID := c.Param("clusterId")
    namespace := c.Param("namespace")
    name := c.Param("name")

    var req struct {
        Field1 string `json:"field1"`
        Field2 int    `json:"field2"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        // ...
    }

    // 业务逻辑...
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

    // 使用 req.ClusterID, req.Field1, req.Field2...
    // 业务逻辑...
}
```

### 模板 3：路径参数 + 查询参数

```go
// 修改前
func (h *K8sAPIHandler) GetPodLogs(c *gin.Context) {
    clusterID := c.Param("clusterId")
    namespace := c.Param("namespace")
    podName := c.Param("name")

    container := c.Query("container")
    tailLines := c.DefaultQuery("tailLines", "100")
    follow := c.Query("follow") == "true"

    // 业务逻辑...
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

    // 使用 req.ClusterID, req.Container, req.TailLines...
    // 业务逻辑...
}
```

## 批量修改脚本

我已经创建了一个示例修改文件，展示了如何修改关键的接口。你可以参考这个文件逐个修改其他接口：

**文件位置**：`internal/handler/k8s_api_optimized_example.go`

## 修改检查清单

每修改一个方法后，检查以下内容：

- [ ] 导入了 `types` 包
- [ ] 使用正确的请求结构体（`types.XXXRequest`）
- [ ] 路径参数使用 `c.ShouldBindUri()`
- [ ] 查询参数使用 `c.ShouldBindQuery()`
- [ ] 请求体参数使用 `c.ShouldBindJSON()`
- [ ] 错误处理统一使用 `response.BadRequest()`
- [ ] 业务逻辑中使用 `req.FieldName` 而不是原来的变量名
- [ ] 移除了不必要的 `c.Param()` 和 `c.Query()` 调用

## 测试策略

### 单元测试

```go
func TestGetCluster(t *testing.T) {
    // 创建测试路由
    r := gin.Default()
    handler := NewK8sAPIHandler(...)
    r.GET("/api/k8s/clusters/:clusterId", handler.GetCluster)

    // 测试正常请求
    req := httptest.NewRequest("GET", "/api/k8s/clusters/cluster-1", nil)
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)

    assert.Equal(t, 200, w.Code)

    // 测试参数缺失
    req = httptest.NewRequest("GET", "/api/k8s/clusters/", nil)
    w = httptest.NewRecorder()
    r.ServeHTTP(w, req)

    assert.Equal(t, 400, w.Code)
}
```

### 集成测试

```bash
# 1. 启动服务
cd cluster-service
make run

# 2. 测试 API
curl -X GET http://localhost:8080/api/k8s/clusters/cluster-1

# 3. 测试参数校验
curl -X POST http://localhost:8080/api/k8s/clusters \
  -H "Content-Type: application/json" \
  -d '{}'  # 应该返回 400 错误

# 4. 测试正常请求
curl -X POST http://localhost:8080/api/k8s/clusters \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-cluster",
    "endpoint": "https://api.k8s.example.com",
    "kubeconfig": "..."
  }'
```

## 自动化修改建议

如果你熟悉 Go 的 AST（抽象语法树）操作，可以编写脚本自动修改：

1. 使用 `go/parser` 解析源文件
2. 使用 `go/ast` 遍历和修改 AST
3. 使用 `go/printer` 输出修改后的代码

但考虑到复杂性和风险，还是建议人工逐个修改，确保质量。

## 实际操作建议

由于完整修改工作量较大，我建议：

### 选项 1：示例驱动

我已经创建了完整的示例文件和指南。你可以：

1. 参考示例文件修改关键接口
2. 运行测试确保功能正常
3. 逐步扩展到其他接口

### 选项 2：生成完整文件

如果需要，我可以生成完整的修改后的 `k8s_api.go` 文件，但需要分多次完成（因为文件太大）。

### 选项 3：关键路径优先

只修改最常用的 20-30 个接口，其他接口保持现状，逐步迁移。

## 推荐方案

考虑到实际情况，我推荐以下方案：

1. **立即修改**：集群管理 + Pod + Deployment（15 个接口）
2. **短期修改**：Node + Service + StatefulSet（15 个接口）
3. **长期迁移**：其他所有接口（50+ 个接口）

这样可以在 2-3 天内完成核心功能的优化，剩余接口可以根据优先级逐步迁移。

## 需要我帮你做什么？

请选择：

**A. 生成完整的修改后的 k8s_api.go 文件**（分多次完成，每次处理一部分）

**B. 只修改核心的 15-20 个接口**（集群+Pod+Deployment+Node）

**C. 提供更详细的修改脚本和工具**

**D. 其他需求**

请告诉我你的选择，我会相应调整工作方式。
