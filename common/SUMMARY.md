# Common Package 实现总结

## 创建时间

2025-10-17

## 概述

成功创建了 `common` 公共包，为 k8s-agent 项目的所有服务提供统一的通用功能。

## 目录结构

```
common/
├── errors/              # 错误处理和错误码定义
│   └── errors.go
├── response/            # 统一的 API 响应格式
│   └── response.go
├── pagination/          # 分页功能
│   └── pagination.go
├── logger/              # 日志工具（基于 Zap）
│   └── logger.go
├── k8sutils/            # Kubernetes 资源转换工具
│   └── converter.go
├── validator/           # 数据验证工具
│   └── validator.go
├── middleware/          # Gin 中间件集合
│   ├── logging.go       # 请求日志和请求 ID
│   ├── cors.go          # CORS 跨域
│   ├── ratelimit.go     # 限流
│   ├── recovery.go      # Panic 恢复和超时
│   └── context.go       # 上下文辅助
├── examples/            # 使用示例
│   └── simple_api/
│       └── main.go      # 完整的 API 服务示例
├── go.mod               # Go 模块定义
├── go.sum               # 依赖锁定
├── README.md            # 详细文档
└── SUMMARY.md           # 本文件
```

## 实现的功能模块

### 1. response - API 响应格式 ✅

**功能**:
- 统一的 JSON 响应格式（符合 API 文档规范）
- 成功响应：`Success()`, `SuccessWithMessage()`, `SuccessList()`
- 错误响应：`BadRequest()`, `Unauthorized()`, `Forbidden()`, `NotFound()`, `Conflict()`, `InternalError()`, `ServiceUnavailable()`

### 2. errors - 错误处理 ✅

**功能**:
- 结构化错误类型 `AppError`
- 错误码定义（0, 400-5xx, 1000+）
- 错误创建：`New()`, `Wrap()`
- 预定义错误和错误判断工具

### 3. pagination - 分页功能 ✅

**功能**:
- 自动解析查询参数（page, pageSize）
- 计算偏移量和限制
- 分页响应结构

### 4. logger - 日志工具 ✅

**功能**:
- 基于 Zap 的高性能结构化日志
- 多种日志级别和输出格式
- 全局和子 logger 支持

### 5. k8sutils - Kubernetes 工具 ✅

**功能**:
- 资源转换和信息提取
- Pod、Node 状态判断

### 6. validator - 数据验证 ✅

**功能**:
- K8s 资源名称验证
- 标签和注解验证
- 副本数、容器名、镜像名验证

### 7. middleware - Gin 中间件 ✅

**功能**:
- 请求日志和 ID 追踪
- CORS 跨域支持
- 限流保护
- Panic 恢复和超时控制

## 如何使用

### 1. 添加依赖

```go
// 在服务的 go.mod 中
require github.com/kart-io/k8s-agent/common v0.0.0
replace github.com/kart-io/k8s-agent/common => ../common
```

### 2. 初始化和使用

```go
import (
    "github.com/kart-io/k8s-agent/common/response"
    "github.com/kart-io/k8s-agent/common/errors"
    "github.com/kart-io/k8s-agent/common/logger"
    "github.com/kart-io/k8s-agent/common/middleware"
)

func main() {
    // 初始化日志
    logger.Init(&logger.Config{Level: "info", Format: "json"})

    // 创建路由
    r := gin.New()
    r.Use(middleware.Recovery())
    r.Use(middleware.RequestLogger())
    r.Use(middleware.CORS())

    // 处理请求
    r.GET("/api/v1/clusters/:id", func(c *gin.Context) {
        response.Success(c, data)
    })
}
```

## 优势

✅ **统一标准** - 所有服务使用相同的响应格式和错误码
✅ **代码复用** - 避免重复实现通用功能
✅ **最佳实践** - 使用业界标准库和 K8s 规范
✅ **易于扩展** - 清晰的模块划分
✅ **生产就绪** - 完善的错误处理、日志、限流机制

## 下一步

基于 common 包，可以在 cluster-service 中实现完整的 Kubernetes API 接口。
