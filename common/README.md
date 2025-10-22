# Common Package

通用功能包,提供 k8s-agent 项目中所有服务共享的核心功能。

## 目录结构

```
common/
├── config/            # 统一配置结构和 Options 模式
├── db/                # 数据库客户端 (MySQL, Redis)
├── mq/                # 消息队列客户端 (NATS)
├── server/            # HTTP 服务器 (Gin)
├── response/          # 统一的 API 响应格式
├── errors/            # 错误码和错误处理
├── pagination/        # 分页功能
├── logger/            # 日志工具 (基于 Zap)
├── k8sutils/          # Kubernetes 资源转换工具
├── validator/         # 数据验证工具
├── middleware/        # Gin 中间件
└── README.md          # 本文档
```

## 核心模块

### 1. config - 统一配置结构 ⭐ NEW

提供项目统一的配置结构和 Functional Options 模式。

**支持的配置类型**:
- `ServerConfig` - HTTP 服务器配置 (7 个 Options)
- `DatabaseConfig` - MySQL/PostgreSQL 配置 (10 个 Options)
- `RedisConfig` - Redis 缓存配置 (10 个 Options)
- `NATSConfig` - NATS 消息队列配置 (8 个 Options)
- `LoggingConfig` - 日志系统配置 (7 个 Options)
- `JWTConfig` - JWT 认证配置 (2 个 Options)
- `MetricsConfig` - Prometheus 指标配置 (5 个 Options)
- `CORSConfig` - CORS 跨域配置 (4 个 Options)

**总计**: 53 个配置函数,支持所有服务的通用配置需求。

**使用示例**:

```go
import "github.com/kart-io/k8s-agent/common/config"

// 使用 Options 模式创建配置
serverCfg := config.DefaultServerConfig(
    config.WithServerPort(8080),
    config.WithServerMode("release"),
)

dbCfg := config.DefaultDatabaseConfig(
    config.WithDBHost("mysql.example.com"),
    config.WithDBName("myapp"),
    config.WithDBUser("app_user"),
    config.WithDBPassword("secure_password"),
)

redisCfg := config.DefaultRedisConfig(
    config.WithRedisAddr("redis.example.com:6379"),
)
```

**详细文档**: 查看 [config/README.md](./config/README.md)

---

### 2. db - 数据库客户端

提供 MySQL 和 Redis 客户端的封装,使用 Options 模式。

#### MySQL 客户端

**使用示例**:

```go
import "github.com/kart-io/k8s-agent/common/db"

mysqlClient, err := db.NewMySQL(logger,
    db.WithHost("localhost"),
    db.WithPort(3306),
    db.WithDatabase("myapp"),
    db.WithUser("root"),
    db.WithPassword("password"),
)
if err != nil {
    log.Fatal(err)
}
defer mysqlClient.Close()

// 访问 GORM DB 实例
mysqlClient.DB.Create(&user)
```

#### Redis 客户端

**使用示例**:

```go
import "github.com/kart-io/k8s-agent/common/db"

redisClient, err := db.NewRedis(logger,
    db.WithAddr("localhost:6379"),
    db.WithRedisPassword("password"),
)
if err != nil {
    log.Fatal(err)
}
defer redisClient.Close()

// 访问 Redis 客户端
redisClient.Client.Set(ctx, "key", "value", 0)
```

---

### 3. mq - 消息队列客户端

提供 NATS 消息队列客户端封装。

**使用示例**:

```go
import "github.com/kart-io/k8s-agent/common/mq"

natsClient, err := mq.NewNATS(logger,
    mq.WithNATSURL("nats://localhost:4222"),
    mq.WithNATSMaxReconnects(10),
)
if err != nil {
    log.Fatal(err)
}
defer natsClient.Close()

// 发布消息
natsClient.Publish(ctx, "subject", []byte("message"))

// 订阅消息
natsClient.Subscribe(ctx, "subject", func(msg *nats.Msg) {
    // 处理消息
})
```

---

### 4. server - HTTP 服务器

提供 Gin HTTP 服务器封装。

**使用示例**:

```go
import "github.com/kart-io/k8s-agent/common/server"

ginServer := server.NewGinServer(logger,
    server.WithGinPort(8080),
    server.WithGinMode("release"),
)

// 注册路由
ginServer.Router.GET("/health", handleHealth)

// 启动服务器
ginServer.Start()
```

---

### 5. response - API 响应格式

提供统一的 HTTP 响应格式，符合 API 文档规范。

**功能特性**：

- 统一的成功/失败响应格式
- 支持列表响应（带 total 字段）
- 预定义的常见 HTTP 错误响应（400, 401, 403, 404, 409, 500, 503）

**使用示例**：

```go
import "github.com/kart-io/k8s-agent/common/response"

// 成功响应
response.Success(c, data)

// 成功响应（自定义消息）
response.SuccessWithMessage(c, "Cluster created successfully", cluster)

// 列表响应
response.SuccessList(c, items, total)

// 错误响应
response.BadRequest(c, "Invalid parameters", err)
response.NotFound(c, "Cluster not found", err)
response.InternalError(c, "Database error", err)
```

**响应格式**：

```json
{
  "code": 0,
  "message": "success",
  "data": {...}
}
```

### 2. errors - 错误处理

提供统一的错误码和错误类型定义。

**功能特性**：

- 预定义的标准错误码（HTTP 错误码 + 业务错误码）
- 结构化的错误类型（AppError）
- 错误包装和解包功能
- 错误类型判断工具

**使用示例**：

```go
import "github.com/kart-io/k8s-agent/common/errors"

// 创建新错误
err := errors.New(errors.CodeClusterNotFound, "Cluster not found")

// 包装已有错误
err := errors.Wrap(errors.CodeK8sAPIError, "Failed to list pods", originalErr)

// 使用预定义错误
err := errors.ErrClusterNotFound

// 判断错误类型
if errors.IsNotFound(err) {
    // 处理资源不存在的情况
}
```

**错误码定义**：

- `0` - 成功
- `400-5xx` - HTTP 标准错误码
- `1000+` - 业务错误码（集群、资源、K8s API 相关）

### 3. pagination - 分页功能

提供标准化的分页参数解析和响应格式。

**功能特性**：

- 从查询参数自动解析分页信息
- 支持自定义默认值
- 自动限制最大页面大小（防止滥用）
- 计算 offset 和 limit

**使用示例**：

```go
import "github.com/kart-io/k8s-agent/common/pagination"

// 解析分页参数（默认 page=1, pageSize=10）
params := pagination.Parse(c)

// 使用自定义默认值
params := pagination.ParseWithDefaults(c, 1, 20)

// 获取计算后的值
offset := params.GetOffset()  // (page-1) * pageSize
limit := params.GetLimit()    // pageSize，最大不超过 100

// 创建分页响应
response := pagination.NewResponse(items, total, params)
```

**查询参数**：

- `page` - 页码（从 1 开始，默认 1）
- `pageSize` - 每页数量（默认 10，最大 100）

### 4. logger - 日志工具

基于 Zap 的高性能结构化日志工具。

**功能特性**：

- 支持多种日志级别（debug, info, warn, error, fatal）
- JSON 和 Console 两种输出格式
- 灵活的输出目标（stdout, stderr, 文件）
- 自动包含调用者信息
- 全局和局部 logger 支持

**使用示例**：

```go
import (
    "github.com/kart-io/k8s-agent/common/logger"
    "go.uber.org/zap"
)

// 初始化日志（应用启动时）
config := &logger.Config{
    Level:        "info",
    Format:       "json",
    OutputPath:   "stdout",
    EnableCaller: true,
}
logger.Init(config)

// 记录日志
logger.Info("Cluster connected",
    zap.String("cluster_id", "prod-01"),
    zap.Int("node_count", 10),
)

logger.Error("Failed to create pod",
    zap.String("namespace", "default"),
    zap.String("pod_name", "nginx"),
    zap.Error(err),
)

// 创建子 logger（带固定字段）
clusterLogger := logger.With(zap.String("cluster_id", "prod-01"))
clusterLogger.Info("Processing request")
```

**配置说明**：

- `level` - 日志级别：debug, info, warn, error, fatal
- `format` - 输出格式：json（生产环境）, console（开发环境）
- `output_path` - 输出目标：stdout, stderr, 或文件路径
- `enable_caller` - 是否显示文件名和行号

### 5. k8sutils - Kubernetes 工具

Kubernetes 资源转换和操作的实用工具。

**功能特性**：

- 资源对象转 map 工具
- 提取 metadata、Pod、Node 等核心信息
- 资源状态判断（Ready、Running 等）

**使用示例**：

```go
import "github.com/kart-io/k8s-agent/common/k8sutils"

// 转换 K8s 资源为 map
dataMap, err := k8sutils.ConvertToMap(pod)

// 提取 metadata
metadata := k8sutils.ExtractMetadata(pod)

// 提取 Pod 信息（精简版）
podInfo := k8sutils.ExtractPodInfo(pod)

// 提取 Node 信息
nodeInfo := k8sutils.ExtractNodeInfo(node)

// 判断状态
if k8sutils.IsPodReady(pod) {
    // Pod 已就绪
}

if k8sutils.IsNodeReady(node) {
    // Node 已就绪
}
```

### 6. validator - 数据验证

Kubernetes 资源名称和标签的验证工具。

**功能特性**：

- K8s 资源名称验证（符合 DNS-1123 规范）
- 标签键/值验证
- 注解验证
- 副本数、容器名、镜像名验证

**使用示例**：

```go
import "github.com/kart-io/k8s-agent/common/validator"

// 验证资源名称
err := validator.ValidateK8sName("my-deployment")

// 验证命名空间
err := validator.ValidateNamespace("default")

// 验证标签
err := validator.ValidateLabels(map[string]string{
    "app": "nginx",
    "environment": "production",
})

// 验证副本数
err := validator.ValidateReplicas(3)

// 验证容器名称
err := validator.ValidateContainerName("nginx")

// 验证镜像名称
err := validator.ValidateImageName("nginx:1.21")
```

**验证规则**：

- **资源名称**: 小写字母、数字、连字符，长度 1-253
- **标签键**: 可选前缀（域名）+ 名称，名称部分 1-63 字符
- **标签值**: 字母、数字、连字符、下划线、点，长度 0-63
- **副本数**: 0-10000

### 7. middleware - Gin 中间件

常用的 Gin 中间件集合。

**可用中间件**：

- `RequestLogger()` - 请求日志记录
- `RequestID()` - 请求 ID 追踪
- `CORS()` - 跨域资源共享（默认配置）
- `CORSWithConfig()` - 自定义配置的 CORS
- `RateLimitByIP()` - 基于 IP 的限流
- `RateLimitByUser()` - 基于用户的限流
- `Recovery()` - Panic 恢复
- `Timeout()` - 请求超时控制

**使用示例**：

```go
import (
    "github.com/gin-gonic/gin"
    "github.com/kart-io/k8s-agent/common/middleware"
)

func setupRouter() *gin.Engine {
    r := gin.New()

    // 基础中间件
    r.Use(middleware.Recovery())
    r.Use(middleware.RequestID())
    r.Use(middleware.RequestLogger())
    r.Use(middleware.CORS())

    // 限流中间件（每秒 10 个请求，桶容量 20）
    r.Use(middleware.RateLimitByIP(10, 20))

    // 超时中间件（30 秒）
    r.Use(middleware.Timeout(30 * time.Second))

    return r
}
```

## 依赖说明

本包依赖以下第三方库：

- `github.com/gin-gonic/gin` - Web 框架
- `go.uber.org/zap` - 高性能日志库
- `k8s.io/api` - Kubernetes API 类型
- `k8s.io/apimachinery` - Kubernetes 元数据类型

## 使用建议

### 1. 在服务中引入 common 包

在服务的 `go.mod` 中添加：

```go
require (
    github.com/kart-io/k8s-agent/common v0.0.0
)

replace github.com/kart-io/k8s-agent/common => ../common
```

### 2. 初始化日志

在服务启动时初始化日志：

```go
func main() {
    // 初始化日志
    logConfig := &logger.Config{
        Level:        os.Getenv("LOG_LEVEL"),
        Format:       os.Getenv("LOG_FORMAT"),
        OutputPath:   "stdout",
        EnableCaller: true,
    }
    if err := logger.Init(logConfig); err != nil {
        panic(err)
    }
    defer logger.Sync()

    // 启动服务
    // ...
}
```

### 3. 统一错误处理

在 handler 中统一使用 response 和 errors 包：

```go
func GetCluster(c *gin.Context) {
    clusterID := c.Param("id")

    // 验证参数
    if err := validator.ValidateClusterID(clusterID); err != nil {
        response.BadRequest(c, "Invalid cluster ID", err)
        return
    }

    // 查询集群
    cluster, err := service.GetCluster(clusterID)
    if err != nil {
        if errors.IsNotFound(err) {
            response.NotFound(c, "Cluster not found", err)
        } else {
            response.InternalError(c, "Failed to get cluster", err)
        }
        return
    }

    // 返回成功
    response.Success(c, cluster)
}
```

### 4. 分页查询

```go
func ListClusters(c *gin.Context) {
    // 解析分页参数
    params := pagination.Parse(c)

    // 查询数据
    clusters, total, err := service.ListClusters(
        params.GetOffset(),
        params.GetLimit(),
    )
    if err != nil {
        response.InternalError(c, "Failed to list clusters", err)
        return
    }

    // 返回分页结果
    resp := pagination.NewResponse(clusters, total, params)
    response.Success(c, resp)
}
```

## 最佳实践

1. **统一响应格式**: 所有 API 都应使用 `response` 包的方法返回响应
2. **结构化日志**: 使用 `logger` 包记录日志，避免使用 `fmt.Println`
3. **错误处理**: 使用 `errors` 包创建和处理错误，保持错误码一致
4. **参数验证**: 使用 `validator` 包验证输入参数
5. **中间件组合**: 合理使用中间件，保持代码简洁

## 版本说明

- **v0.1.0** - 初始版本，包含核心功能模块

## 贡献指南

如需添加新的通用功能，请遵循以下规范：

1. 功能应该是多个服务共享的
2. 代码应该有清晰的文档和示例
3. 应该包含单元测试
4. 更新本 README 文档

## 许可证

本项目采用 MIT 许可证。
