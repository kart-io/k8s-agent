# 代码冗余清理 - 具体行动项

本文档列出每个清理任务的具体步骤和代码示例。

## 高优先级 - Phase 1（1-2 天）

### Task 1.1: 删除 NewK8sAPIHandlerLegacy()

**位置**：`internal/cluster/handler/k8s_api.go:88-154`

**当前代码**：
```go
// NewK8sAPIHandlerLegacy 保留旧的构造函数用于向后兼容（已废弃）.
// Deprecated: Use NewK8sAPIHandler with K8sServiceRegistry instead.
func NewK8sAPIHandlerLegacy(
    clusterService *service.K8sClusterService,
    namespaceService *service.K8sNamespaceService,
    // ... 30 个参数 ...
    resourcequotaService *service.K8sResourceQuotaService,
) *K8sAPIHandler {
    return &K8sAPIHandler{
        clusterService:            clusterService,
        // ... 30 个字段赋值 ...
    }
}
```

**清理步骤**：
1. 全局搜索验证：`grep -r "NewK8sAPIHandlerLegacy" .`
2. 如果无任何引用，直接删除该函数（第 88-154 行）
3. 运行测试：`make test`

**预期成果**：删除 67 行代码

---

### Task 1.2: 重命名 PostgresStore → MySQLStore

**受影响文件**：
```
./internal/agent-manager/storage/postgres.go
./internal/orchestrator/storage/postgres.go
./internal/auth/storage/postgres.go
./internal/monitor/storage/postgres.go
./internal/auth/forced-logout/audit/postgres_repository.go
./internal/auth/forced-logout/notification/postgres_repository.go
```

**当前模式**：
```go
// File: internal/agent-manager/storage/postgres.go
type PostgresStore struct {
    *db.MySQLClient // 实际使用 MySQL 客户端
    logger          core.Logger
}

func NewPostgresStore(config types.DatabaseConfig, log core.Logger) (*PostgresStore, error) {
    // 使用 db.NewMySQL() 创建客户端
    mysqlClient, err := db.NewMySQL(...)
    // ...
}
```

**清理步骤**：

对每个文件执行：
```bash
# 1. 重命名类型
sed -i 's/PostgresStore/MySQLStore/g' internal/agent-manager/storage/postgres.go

# 2. 更新构造函数名
sed -i 's/NewPostgresStore/NewMySQLStore/g' internal/agent-manager/storage/postgres.go

# 3. 更新所有导入此类型的文件
grep -r "PostgresStore\|NewPostgresStore" internal/ | grep -v ".pb.go"
# 对每个引用都进行更新

# 4. 更新注释
```

**示例 - 文件的变化**：

Before:
```go
// PostgresStore implements storage using MySQL
// Note: Kept the name for backward compatibility, but now using MySQL.
type PostgresStore struct {
    *db.MySQLClient
    logger core.Logger
}

func NewPostgresStore(config types.DatabaseConfig, log core.Logger) (*PostgresStore, error) {
```

After:
```go
// MySQLStore implements storage using MySQL
type MySQLStore struct {
    *db.MySQLClient
    logger core.Logger
}

func NewMySQLStore(config types.DatabaseConfig, log core.Logger) (*MySQLStore, error) {
```

**验证**：
```bash
# 确保没有遗漏的 PostgresStore 引用
grep -r "PostgresStore" . --include="*.go" | grep -v ".pb.go" | grep -v "test"
# 应该输出为空

# 运行测试
make test
```

**预期成果**：减少 802 行误导性代码

---

### Task 1.3: 合并 Response 处理器

**涉及文件**：
- `common/response/response.go` - 保留，扩展
- `internal/auth/response/response.go` - 删除

**第一步：扩展 common/response**

在 `common/response/response.go` 中添加 Auth 特定的错误码：

```go
// 在现有代码后添加
const (
    // Auth specific error codes
    CodeDatabaseError     = 5001
    CodeValidationError   = 4001
    CodeAuthenticationErr = 4011
    CodePermissionDenied  = 4031
)

// 添加 Auth 特定的响应函数
func ValidationError(c *gin.Context, message string, err error) {
    Error(c, http.StatusBadRequest, CodeValidationError, "Validation Error", err)
}

func AuthenticationError(c *gin.Context, message string, err error) {
    Error(c, http.StatusUnauthorized, CodeAuthenticationErr, "Authentication Failed", err)
}

func PermissionDenied(c *gin.Context, message string, err error) {
    Error(c, http.StatusForbidden, CodePermissionDenied, "Permission Denied", err)
}

func DatabaseError(c *gin.Context, message string, err error) {
    Error(c, http.StatusInternalServerError, CodeDatabaseError, "Database Error", err)
}
```

**第二步：更新 Auth 服务导入**

在所有 auth handler 中：
```go
// Before
import "github.com/kart-io/k8s-agent/internal/auth/response"
// ...
response.ValidationError(c, msg)

// After
import "github.com/kart-io/k8s-agent/common/response"
// ...
response.ValidationError(c, msg)
```

**第三步：删除 internal/auth/response/response.go**

验证后删除整个文件：
```bash
grep -r "internal/auth/response" . --include="*.go"
# 确保没有直接导入（应该都改为 common/response）

rm internal/auth/response/response.go
```

**验证**：
```bash
make test
# 特别关注 auth 相关测试
make go.test.auth
```

**预期成果**：删除 128 行，统一响应处理

---

## 中优先级 - Phase 2（3-5 天）

### Task 2.1: 合并 Pagination 处理器 ✅ COMPLETED

**状态**: 已完成 (2025-11-06)

**执行结果**：
- 扩展了 `common/pagination/pagination.go`，添加了以下功能：
  - Sort 和 Order 字段支持（排序功能）
  - TotalPages 字段（总页数计算）
  - `CalculateTotalPages()` 函数
  - `BuildOrderBy()` 函数（SQL ORDER BY 子句构建）
- 删除了 `internal/auth/pagination/` 目录（98 行代码）
- **发现**：auth pagination 包实际上未被任何代码使用，可以安全删除
- 更新前 common/pagination: 87 行
- 更新后 common/pagination: 138 行（净增加 51 行，但删除了 98 行重复代码）
- **净删除**: 98 - 51 = 47 行代码
- 编译验证通过：`make go.build.auth`
- 测试验证通过：`go test ./internal/auth/...`

**功能差异分析**：
- Auth 版本有 Sort/Order 支持 → 已迁移到 common
- Auth 版本有 TotalPages 计算 → 已迁移到 common
- Auth 版本有 BuildOrderBy() → 已迁移到 common
- Common 版本的 DefaultPageSize=10，Auth 版本是 20 → 保持 common 版本（更合理）
- **破坏性变更**: 无（因为 auth pagination 未被使用）

---

### Task 2.2: 原 Task 2.2（待执行）

`internal/auth/pagination/pagination.go` → 改为 thin wrapper：

```go
package pagination

import (
    "github.com/gin-gonic/gin"
    commonpaging "github.com/kart-io/k8s-agent/common/pagination"
    "github.com/kart-io/k8s-agent/internal/auth/types"
)

// 保留 Auth 特定常量
const (
    DefaultPage     = 1
    DefaultPageSize = 20
    MaxPageSize     = 100
)

// GetPaginationParams 包装 common 版本
func GetPaginationParams(c *gin.Context) types.PaginationParams {
    params := commonpaging.Parse(c)
    return types.PaginationParams{
        Page:     params.Page,
        PageSize: ValidatePageSize(params.PageSize, MaxPageSize),
        Sort:     params.Sort,
        Order:    params.Order,
    }
}

// 其他函数保持调用 common 版本
```

**最终步骤：可选 - 完全删除 internal/auth/pagination**

如果 auth 不需要特定常量：
```bash
# 更新所有导入
grep -r "internal/auth/pagination" . --include="*.go"
sed -i 's|internal/auth/pagination|common/pagination|g' internal/auth/handler/*.go

# 删除文件
rm -rf internal/auth/pagination/
```

**预期成果**：删除 98 行，统一分页处理

---

### Task 2.2: 实现 TODO - NATS 命令结果处理

**位置**：`internal/agent-manager/nats/server.go`

**当前代码**：
```go
// TODO: Process command result
func (s *NATSServer) handleCommandResult(result *types.CommandResult) {
    s.logger.Infow("Received command result",
        "command_id", result.CommandID,
        "agent_id", result.AgentID,
    )
    // TODO: Process command result
}
```

**实现建议**：
```go
func (s *NATSServer) handleCommandResult(result *types.CommandResult) {
    s.logger.Infow("Processing command result",
        "command_id", result.CommandID,
        "agent_id", result.AgentID,
        "status", result.Status,
    )

    // 1. 验证结果
    if err := s.validateCommandResult(result); err != nil {
        s.logger.Errorw("Invalid command result", "error", err.Error())
        return
    }

    // 2. 更新数据库中的命令状态
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := s.storage.UpdateCommandResult(ctx, result); err != nil {
        s.logger.Errorw("Failed to update command result", "error", err.Error())
        return
    }

    // 3. 发布事件供 orchestrator 订阅
    s.eventBus.Publish("command.completed", result)

    s.logger.Infow("Command result processed successfully")
}
```

---

### Task 2.3: 实现 TODO - Auth 会话撤销

**位置**：`internal/auth/handler/auth_handler.go`

**当前代码**：
```go
// TODO: Implement session revocation in Phase 4
if err := h.sessionService.RevokeSession(ctx, sessionID); err != nil {
    // Handle error
}
```

**实现建议**：
```go
func (h *AuthHandler) RevokeSession(c *gin.Context) {
    var req types.ForceLogoutRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, "Invalid request", err)
        return
    }

    ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
    defer cancel()

    // 撤销会话
    if err := h.sessionService.RevokeSession(ctx, req.SessionID); err != nil {
        h.logger.Errorw("Failed to revoke session", "error", err.Error())
        response.InternalError(c, "Failed to revoke session", err)
        return
    }

    // 记录审计日志
    h.auditLog.LogForcedLogout(c.Request.Context(), map[string]interface{}{
        "session_id": req.SessionID,
        "reason":     req.Reason,
        "timestamp":  time.Now(),
    })

    response.Success(c, gin.H{
        "message": "Session revoked successfully",
        "session_id": req.SessionID,
    })
}
```

---

## 低优先级 - Phase 3（2-4 周）

### Task 3.1: 创建 Handler 装饰器（可选）

**目标**：减少 Handler 中的重复代码

**新建文件**：`pkg/handler/decorator.go`

```go
package handler

import (
    "github.com/gin-gonic/gin"
    "github.com/kart-io/logger"
    "github.com/kart-io/k8s-agent/common/response"
    "github.com/kart-io/k8s-agent/common/validator"
)

type HandlerFunc func(c *gin.Context, req interface{}) (interface{}, error)

// WithValidation 装饰器 - 自动处理验证
func WithValidation(validator *validator.Validator) func(HandlerFunc) HandlerFunc {
    return func(next HandlerFunc) HandlerFunc {
        return func(c *gin.Context, req interface{}) (interface{}, error) {
            if err := validator.Validate(req); err != nil {
                return nil, err
            }
            return next(c, req)
        }
    }
}

// WithLogging 装饰器 - 自动处理日志
func WithLogging(action string) func(HandlerFunc) HandlerFunc {
    return func(next HandlerFunc) HandlerFunc {
        return func(c *gin.Context, req interface{}) (interface{}, error) {
            logger.Infow("Action started",
                "action", action,
                "user_id", c.GetString("user_id"),
            )

            result, err := next(c, req)

            if err != nil {
                logger.Errorw("Action failed",
                    "action", action,
                    "error", err.Error(),
                )
            } else {
                logger.Infow("Action completed", "action", action)
            }

            return result, err
        }
    }
}
```

**使用示例**：
```go
// Before: handler 中有 40 行重复代码
func (h *Handler) CreateUser(c *gin.Context) {
    var req types.UserCreateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, "Invalid input", err)
        return
    }
    if err := validator.Validate(req); err != nil {
        response.BadRequest(c, "Validation failed", err)
        return
    }
    logger.Infow("Creating user", "name", req.Name)
    
    result, err := h.service.CreateUser(c.Request.Context(), &req)
    if err != nil {
        logger.Errorw("Create user failed", "error", err.Error())
        response.InternalError(c, "Create failed", err)
        return
    }
    
    response.Success(c, result)
}

// After: 清晰的业务逻辑
func (h *Handler) CreateUserDecorated(c *gin.Context) {
    var req types.UserCreateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.BadRequest(c, "Invalid input", err)
        return
    }

    result, err := h.createUserLogic(c, &req)
    if err != nil {
        response.InternalError(c, "Create failed", err)
        return
    }

    response.Success(c, result)
}

func (h *Handler) createUserLogic(c *gin.Context, req *types.UserCreateRequest) (interface{}, error) {
    return h.service.CreateUser(c.Request.Context(), req)
}
```

---

### Task 3.2: 拆分大型文件

**示例 - 拆分 k8s_api.go (4072 行)**

当前结构：
```
k8s_api.go (所有 30 个资源类型的 handler)
```

目标结构：
```
k8s_api_base.go        (公共接口和初始化)
k8s_cluster.go         (Cluster 相关 handler)
k8s_pod.go             (Pod 相关 handler)
k8s_deployment.go      (Deployment 相关 handler)
k8s_node.go            (Node 相关 handler)
...（每个资源类型一个文件）
```

**拆分步骤**：
1. 创建 `k8s_api_base.go`，包含 K8sAPIHandler 结构体和初始化
2. 创建 `k8s_cluster.go`，包含所有 Cluster 相关方法
3. 对其他 29 个资源类型重复步骤 2
4. 删除原始 `k8s_api.go`
5. 运行测试确认功能完整

---

## 验证清单

清理完成后的检查清单：

```bash
# 1. 编译检查
make build
# 错误 → 检查是否有遗漏的导入或引用

# 2. 测试
make test
# 失败 → 检查是否有测试需要更新

# 3. Linting
make lint
# 警告 → 修复任何代码风格问题

# 4. 验证无遗漏引用
grep -r "PostgresStore\|NewK8sAPIHandlerLegacy\|internal/auth/response" . --include="*.go" | grep -v test | grep -v ".pb.go"
# 应该输出为空

# 5. 运行特定服务测试
make go.test.auth
make go.test.agent-manager
make go.test.cluster

# 6. 端到端测试
make test-e2e

# 7. 代码覆盖检查
make test-coverage
# 验证覆盖率未下降
```

---

## Git 提交建议

分离提交有利于 review：

```bash
# 提交 1: 删除废弃代码
git add internal/cluster/handler/k8s_api.go
git commit -m "chore: remove deprecated NewK8sAPIHandlerLegacy function

- Removed unused legacy constructor (67 lines)
- NewK8sAPIHandler with K8sServiceRegistry is the standard approach"

# 提交 2: 重命名 PostgresStore
git add internal/*/storage/postgres.go
git commit -m "chore: rename PostgresStore to MySQLStore for clarity

- Renamed class from PostgresStore to MySQLStore (802 lines)
- Updated all constructors and imports
- Updated comments to reflect MySQL usage
- Addresses code clarity: implementation uses MySQL, naming should reflect that"

# 提交 3: 合并 Response
git add common/response/ internal/auth/
git commit -m "chore: consolidate response handlers into common package

- Moved auth-specific error codes to common/response
- Removed duplicate functions (Success, Error, BadRequest, etc.)
- Auth handlers now import from common/response
- Reduced code duplication: 128 lines deleted"

# 提交 4: 实现 TODO
git add internal/agent-manager/nats/server.go
git commit -m "feat: implement NATS command result processing

- Implemented TODO: Process command result
- Added result validation
- Added database update for command status
- Published events for orchestrator subscription"
```

