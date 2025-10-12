# Auth Service 快速入门指南

## 项目已创建的内容

1. **目录结构**
   ```
   auth-service/
   ├── cmd/server/          # 主程序入口
   ├── internal/
   │   ├── handler/         # HTTP 处理器
   │   ├── middleware/      # 中间件（JWT、权限检查等）
   │   ├── model/           # 数据库模型
   │   ├── service/         # 业务逻辑
   │   └── storage/         # 数据存储层
   ├── pkg/types/           # 类型定义 ✅ 已创建
   ├── configs/             # 配置文件 ✅ 已创建
   ├── go.mod              # Go 模块 ✅ 已创建
   ├── Makefile            # 构建脚本 ✅ 已创建
   └── README.md           # 文档 ✅ 已创建
   ```

2. **数据模型** (`pkg/types/types.go`)
   - User (用户)
   - Role (角色)
   - Permission (权限)
   - UserRole (用户角色关联)
   - RolePermission (角色权限关联)
   - APIKey (API密钥)

## 需要实现的核心文件

### 1. 主程序 `cmd/server/main.go`

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kart-io/k8s-agent/auth-service/internal/handler"
	"github.com/kart-io/k8s-agent/auth-service/internal/middleware"
	"github.com/kart-io/k8s-agent/auth-service/internal/service"
	"github.com/kart-io/k8s-agent/auth-service/internal/storage"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	if err := loadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 连接数据库
	db, err := storage.NewMySQLDB()
	if err != nil {
		logger.Fatal("Failed to connect database", zap.Error(err))
	}

	// 自动迁移数据库
	if err := storage.AutoMigrate(db); err != nil {
		logger.Fatal("Failed to migrate database", zap.Error(err))
	}

	// 连接 Redis
	rdb := storage.NewRedisClient()

	// 初始化服务
	authService := service.NewAuthService(db, rdb, logger)
	userService := service.NewUserService(db, logger)
	roleService := service.NewRoleService(db, logger)
	permissionService := service.NewPermissionService(db, logger)

	// 初始化处理器
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	roleHandler := handler.NewRoleHandler(roleService)
	permissionHandler := handler.NewPermissionHandler(permissionService)

	// 设置路由
	router := setupRouter(authHandler, userHandler, roleHandler, permissionHandler)

	// 启动服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", viper.GetInt("server.port")),
		Handler: router,
	}

	// 优雅关闭
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	logger.Info("Server started", zap.Int("port", viper.GetInt("server.port")))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

func loadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	return viper.ReadInConfig()
}

func setupRouter(
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	roleHandler *handler.RoleHandler,
	permissionHandler *handler.PermissionHandler,
) *gin.Engine {
	router := gin.Default()

	// 公开接口
	api := router.Group("/api/v1")
	{
		// 认证
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", middleware.JWTAuth(), authHandler.Logout)
			auth.GET("/me", middleware.JWTAuth(), authHandler.GetCurrentUser)
			auth.GET("/menus", middleware.JWTAuth(), authHandler.GetUserMenus)
			auth.POST("/check", authHandler.CheckPermission)
		}

		// 需要认证的接口
		authenticated := api.Group("")
		authenticated.Use(middleware.JWTAuth())
		{
			// 用户管理
			users := authenticated.Group("/users")
			{
				users.GET("", userHandler.List)
				users.GET("/:id", userHandler.Get)
				users.POST("", middleware.RequirePermission("user:create"), userHandler.Create)
				users.PUT("/:id", middleware.RequirePermission("user:update"), userHandler.Update)
				users.DELETE("/:id", middleware.RequirePermission("user:delete"), userHandler.Delete)
			}

			// 角色管理
			roles := authenticated.Group("/roles")
			{
				roles.GET("", roleHandler.List)
				roles.GET("/:id", roleHandler.Get)
				roles.POST("", middleware.RequirePermission("role:create"), roleHandler.Create)
				roles.PUT("/:id", middleware.RequirePermission("role:update"), roleHandler.Update)
				roles.DELETE("/:id", middleware.RequirePermission("role:delete"), roleHandler.Delete)
				roles.POST("/:id/permissions", middleware.RequirePermission("role:assign"), roleHandler.AssignPermissions)
			}

			// 权限管理
			permissions := authenticated.Group("/permissions")
			{
				permissions.GET("", permissionHandler.List)
				permissions.GET("/tree", permissionHandler.GetTree)
				permissions.GET("/:id", permissionHandler.Get)
				permissions.POST("", middleware.RequirePermission("permission:create"), permissionHandler.Create)
				permissions.PUT("/:id", middleware.RequirePermission("permission:update"), permissionHandler.Update)
				permissions.DELETE("/:id", middleware.RequirePermission("permission:delete"), permissionHandler.Delete)
			}
		}
	}

	return router
}
```

## 后续开发步骤

### 第1步：实现数据存储层 (internal/storage/)

1. `mysql.go` - MySQL 连接和初始化
2. `redis.go` - Redis 连接和初始化
3. `migrate.go` - 数据库迁移和初始化数据

### 第2步：实现服务层 (internal/service/)

1. `auth_service.go` - 认证服务（登录、Token生成）
2. `user_service.go` - 用户服务（CRUD）
3. `role_service.go` - 角色服务（CRUD）
4. `permission_service.go` - 权限服务（CRUD）

### 第3步：实现中间件 (internal/middleware/)

1. `jwt.go` - JWT 认证中间件
2. `permission.go` - 权限检查中间件
3. `api_key.go` - API Key 认证中间件
4. `cors.go` - CORS 中间件

### 第4步：实现HTTP处理器 (internal/handler/)

1. `auth_handler.go` - 认证相关接口
2. `user_handler.go` - 用户管理接口
3. `role_handler.go` - 角色管理接口
4. `permission_handler.go` - 权限管理接口

## 启动步骤

```bash
# 1. 进入项目目录
cd auth-service

# 2. 初始化依赖
make deps

# 3. 创建数据库
make init-db

# 4. 运行服务
make run

# 或者编译后运行
make build
./bin/auth-service
```

## 数据库初始化

服务首次启动时会自动：
1. 创建所有表
2. 创建默认超级管理员用户（admin/admin123）
3. 创建默认角色（超级管理员、管理员、普通用户）
4. 创建默认权限（系统管理、用户管理等）

## 测试接口

```bash
# 1. 登录获取 token
curl -X POST http://localhost:8090/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# 2. 使用 token 获取用户信息
curl -X GET http://localhost:8090/api/v1/auth/me \
  -H "Authorization: Bearer <your-token>"

# 3. 获取用户菜单
curl -X GET http://localhost:8090/api/v1/auth/menus \
  -H "Authorization: Bearer <your-token>"
```

## 集成到现有系统

### 在 agent-manager-ui 中集成

1. 修改登录页面调用 auth-service 的登录接口
2. 保存返回的 token
3. 在所有请求中添加 Authorization header
4. 根据返回的菜单权限动态渲染菜单
5. 根据按钮权限控制按钮显示

### 在其他服务中集成

其他服务（agent-manager、orchestrator-service等）可以：
1. 调用 auth-service 的权限检查接口验证用户权限
2. 或者实现自己的 JWT 验证中间件（共享密钥）

## 下一步

你现在可以：
1. 按照上述步骤依次实现各个模块
2. 或者让我帮你生成具体的某个文件（比如先生成 storage 层）
3. 或者直接运行一个简化版本进行测试

需要我继续实现哪个部分？
