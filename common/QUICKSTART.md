# Common Package 快速开始

本指南帮助您快速在服务中集成 `common` 包。

## 步骤 1: 添加依赖

在您的服务目录下（如 `cluster-service/`），编辑 `go.mod`：

```go
module github.com/kart-io/k8s-agent/cluster-service

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/kart-io/k8s-agent/common v0.0.0
    // ... 其他依赖
)

// 使用本地路径
replace github.com/kart-io/k8s-agent/common => ../common
```

然后运行：

```bash
cd cluster-service
go mod tidy
```

## 步骤 2: 初始化日志

在 `main.go` 中初始化日志：

```go
package main

import (
    "github.com/kart-io/k8s-agent/common/logger"
)

func main() {
    // 初始化日志
    logConfig := &logger.Config{
        Level:        "info",
        Format:       "json",
        OutputPath:   "stdout",
        EnableCaller: true,
    }
    if err := logger.Init(logConfig); err != nil {
        panic(err)
    }
    defer logger.Sync()

    // 启动服务...
}
```

## 步骤 3: 设置路由和中间件

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/kart-io/k8s-agent/common/middleware"
)

func setupRouter() *gin.Engine {
    r := gin.New()

    // 注册中间件（按顺序）
    r.Use(middleware.Recovery())           // Panic 恢复
    r.Use(middleware.RequestID())          // 请求 ID
    r.Use(middleware.RequestLogger())      // 请求日志
    r.Use(middleware.CORS())               // 跨域
    r.Use(middleware.RateLimitByIP(10, 20)) // 限流

    return r
}
```

## 步骤 4: 编写 Handler

```go
package handler

import (
    "github.com/gin-gonic/gin"
    "github.com/kart-io/k8s-agent/common/response"
    "github.com/kart-io/k8s-agent/common/errors"
    "github.com/kart-io/k8s-agent/common/pagination"
    "github.com/kart-io/k8s-agent/common/validator"
    "github.com/kart-io/k8s-agent/common/logger"
    "go.uber.org/zap"
)

// GetCluster 获取集群详情
func GetCluster(c *gin.Context) {
    clusterID := c.Param("id")

    // 1. 验证参数
    if err := validator.ValidateClusterID(clusterID); err != nil {
        response.BadRequest(c, "Invalid cluster ID", err)
        return
    }

    // 2. 记录日志
    logger.Info("Getting cluster", zap.String("cluster_id", clusterID))

    // 3. 调用服务层
    cluster, err := service.GetCluster(clusterID)
    if err != nil {
        // 判断错误类型
        if errors.IsNotFound(err) {
            response.NotFound(c, "Cluster not found", err)
        } else {
            response.InternalError(c, "Failed to get cluster", err)
        }
        return
    }

    // 4. 返回成功响应
    response.Success(c, cluster)
}

// ListClusters 获取集群列表（带分页）
func ListClusters(c *gin.Context) {
    // 1. 解析分页参数
    params := pagination.Parse(c)

    logger.Info("Listing clusters",
        zap.Int("page", params.Page),
        zap.Int("page_size", params.GetPageSize()),
    )

    // 2. 查询数据
    clusters, total, err := service.ListClusters(
        params.GetOffset(),
        params.GetLimit(),
    )
    if err != nil {
        response.InternalError(c, "Failed to list clusters", err)
        return
    }

    // 3. 返回分页结果
    resp := pagination.NewResponse(clusters, total, params)
    response.Success(c, resp)
}

// CreateCluster 创建集群
func CreateCluster(c *gin.Context) {
    var req CreateClusterRequest

    // 1. 绑定 JSON
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, "Invalid request body", err)
        return
    }

    // 2. 验证参数
    if err := validator.ValidateK8sName(req.Name); err != nil {
        response.BadRequest(c, "Invalid cluster name", err)
        return
    }

    // 3. 创建集群
    cluster, err := service.CreateCluster(req)
    if err != nil {
        if errors.IsConflict(err) {
            response.Conflict(c, "Cluster already exists", err)
        } else {
            response.InternalError(c, "Failed to create cluster", err)
        }
        return
    }

    // 4. 返回成功
    response.SuccessWithMessage(c, "Cluster created successfully", cluster)
}
```

## 步骤 5: 服务层错误处理

```go
package service

import (
    "github.com/kart-io/k8s-agent/common/errors"
)

func GetCluster(clusterID string) (*Cluster, error) {
    cluster, err := db.GetCluster(clusterID)
    if err != nil {
        if err == sql.ErrNoRows {
            // 包装为应用错误
            return nil, errors.ErrClusterNotFound
        }
        return nil, errors.Wrap(
            errors.CodeInternalError,
            "Database query failed",
            err,
        )
    }
    return cluster, nil
}
```

## 常见场景示例

### 场景 1: 参数验证

```go
// 验证集群 ID
if err := validator.ValidateClusterID(clusterID); err != nil {
    response.BadRequest(c, "Invalid cluster ID", err)
    return
}

// 验证命名空间
if err := validator.ValidateNamespace(namespace); err != nil {
    response.BadRequest(c, "Invalid namespace", err)
    return
}

// 验证标签
if err := validator.ValidateLabels(labels); err != nil {
    response.BadRequest(c, "Invalid labels", err)
    return
}
```

### 场景 2: 分页查询

```go
func ListPods(c *gin.Context) {
    // 解析分页参数（默认 page=1, pageSize=10）
    params := pagination.Parse(c)

    // 从数据库查询
    pods, total, err := podService.List(
        params.GetOffset(),  // (page-1) * pageSize
        params.GetLimit(),   // pageSize (最大 100)
    )
    if err != nil {
        response.InternalError(c, "Query failed", err)
        return
    }

    // 返回分页结果
    resp := pagination.NewResponse(pods, total, params)
    response.Success(c, resp)
}
```

### 场景 3: 日志记录

```go
import (
    "github.com/kart-io/k8s-agent/common/logger"
    "go.uber.org/zap"
)

// 简单日志
logger.Info("Pod created", zap.String("pod_name", name))

// 错误日志
logger.Error("Failed to delete pod",
    zap.String("pod_name", name),
    zap.String("namespace", namespace),
    zap.Error(err),
)

// 创建子 logger（带固定字段）
podLogger := logger.With(
    zap.String("cluster_id", clusterID),
    zap.String("namespace", namespace),
)
podLogger.Info("Processing pod")
```

### 场景 4: K8s 资源处理

```go
import (
    "github.com/kart-io/k8s-agent/common/k8sutils"
    corev1 "k8s.io/api/core/v1"
)

func GetPod(podName string) (*corev1.Pod, error) {
    pod, err := k8sClient.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
    if err != nil {
        return nil, err
    }

    // 提取 Pod 信息（精简版）
    podInfo := k8sutils.ExtractPodInfo(pod)

    // 判断状态
    if k8sutils.IsPodReady(pod) {
        logger.Info("Pod is ready")
    }

    return pod, nil
}
```

## 完整示例

参考 `common/examples/simple_api/main.go` 查看完整的服务示例。

运行示例：

```bash
cd common/examples/simple_api
go run main.go
```

测试 API：

```bash
# 获取集群列表
curl http://localhost:8080/api/v1/clusters

# 获取单个集群
curl http://localhost:8080/api/v1/clusters/cluster-1

# 创建集群
curl -X POST http://localhost:8080/api/v1/clusters \
  -H "Content-Type: application/json" \
  -d '{"name":"test-cluster","description":"Test cluster"}'
```

## 注意事项

1. **日志初始化**: 在 `main` 函数开始时初始化日志
2. **中间件顺序**: 按照示例顺序注册中间件
3. **错误处理**: 统一使用 `errors` 包创建和判断错误
4. **参数验证**: 在 handler 层进行参数验证
5. **响应格式**: 统一使用 `response` 包返回响应

## 更多信息

- 详细文档: `common/README.md`
- 实现总结: `common/SUMMARY.md`
- 使用示例: `common/examples/simple_api/main.go`
