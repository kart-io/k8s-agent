package initializers

import (
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"context"

	"google.golang.org/grpc"

	
	commonserver "github.com/kart-io/k8s-agent/common/server"
	authgrpc "github.com/kart-io/k8s-agent/internal/auth/grpc"
	"github.com/kart-io/k8s-agent/internal/auth/service"
	"github.com/kart-io/k8s-agent/internal/auth/storage"
	authv1 "github.com/kart-io/k8s-agent/pkg/api/auth/v1"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	commoninitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// GRPCServerInitializer gRPC服务器初始化器
type GRPCServerInitializer struct {
	standardInit *commoninitializers.GRPCServerInitializer
	opts         *commonapp.StandardOptions
	logger       core.Logger

	// 依赖的初始化器
	dbInit      *DatabaseInitializer
	redisInit   *RedisInitializer
	sessionInit *SessionServiceInitializer

	// 共享的存储层包装实例（供 HTTP 初始化器复用）
	mysqlDB     *storage.MySQLDB
	redisClient *storage.RedisClient

	// 共享的 Service 实例（供 HTTP 初始化器复用）
	authService       *service.AuthService
	userService       *service.UserService
	roleService       *service.RoleService
	permissionService *service.PermissionService
}

// NewGRPCServerInitializer 创建gRPC服务器初始化器
func NewGRPCServerInitializer(
	opts *commonapp.StandardOptions,
	logger core.Logger,
	dbInit *DatabaseInitializer,
	redisInit *RedisInitializer,
	sessionInit *SessionServiceInitializer,
) *GRPCServerInitializer {
	return &GRPCServerInitializer{
		opts:        opts,
		logger:      logger,
		dbInit:      dbInit,
		redisInit:   redisInit,
		sessionInit: sessionInit,
	}
}

// Name 返回初始化器名称
func (g *GRPCServerInitializer) Name() string {
	return "auth-grpc-server"
}

// Priority 返回初始化优先级
func (g *GRPCServerInitializer) Priority() int {
	return bootstrap.PriorityGRPC
}

// Initialize 执行初始化
func (g *GRPCServerInitializer) Initialize(ctx context.Context) error {
	if !g.opts.GRPC.Enable {
		g.logger.Infow("gRPC server is disabled, skipping initialization")
		return nil
	}

	g.logger.Infow("Initializing gRPC server for auth service",
		"host", g.opts.GRPC.Host,
		"port", g.opts.GRPC.Port,
	)

	// TODO: 创建gRPC服务实现
	// 由于Auth服务结构复杂（多个Handler），需要重构为统一Service层
	// 当前先创建基础框架

	// 创建并保存存储层包装实例（供 HTTP 初始化器复用）
	g.mysqlDB = &storage.MySQLDB{
		DB:     g.dbInit.DB(),
		Logger: g.logger,
	}
	g.redisClient = &storage.RedisClient{
		Client: g.redisInit.Client(),
	}

	// 创建Service层实例（使用保存的包装实例，供 HTTP 初始化器复用）
	g.authService = service.NewAuthService(g.mysqlDB, g.redisClient, g.opts, g.logger)
	g.userService = service.NewUserService(g.mysqlDB, g.logger)
	g.roleService = service.NewRoleService(g.mysqlDB, g.logger)
	g.permissionService = service.NewPermissionService(g.mysqlDB, g.logger)

	// 创建gRPC服务器实例（使用保存的实例）
	authGRPC := authgrpc.NewAuthServiceServer(g.authService, g.logger)
	userGRPC := authgrpc.NewUserServiceServer(g.userService, g.logger)
	roleGRPC := authgrpc.NewRoleServiceServer(g.roleService, g.logger)
	permissionGRPC := authgrpc.NewPermissionServiceServer(g.permissionService, g.logger)
	sessionGRPC := authgrpc.NewSessionServiceServer(g.sessionInit.Service(), g.logger)

	serverConfig := &commoninitializers.GRPCServerConfig{
		Name:     g.Name(),
		Priority: g.Priority(),
		Config:   g.opts.GRPC,
		ServiceRegister: func(s *grpc.Server) error {
			// 注册所有gRPC服务
			authv1.RegisterAuthServiceServer(s, authGRPC)
			authv1.RegisterUserServiceServer(s, userGRPC)
			authv1.RegisterRoleServiceServer(s, roleGRPC)
			authv1.RegisterPermissionServiceServer(s, permissionGRPC)
			authv1.RegisterSessionServiceServer(s, sessionGRPC)

			g.logger.Infow("All gRPC services registered successfully",
				"services", []string{"Auth", "User", "Role", "Permission", "Session"},
			)
			return nil
		},
	}

	g.standardInit = commoninitializers.NewGRPCServerInitializer(serverConfig, g.logger)
	return g.standardInit.Initialize(ctx)
}

// Close 关闭gRPC服务器
func (g *GRPCServerInitializer) Close(ctx context.Context) error {
	if g.standardInit == nil {
		return nil
	}
	return g.standardInit.Close(ctx)
}

// GetServer 返回服务器实例（实现ServerProvider接口）
func (g *GRPCServerInitializer) GetServer() commonserver.Server {
	if g.standardInit == nil {
		return nil
	}
	return g.standardInit.GetServer()
}

// AuthService 返回共享的 AuthService 实例
func (g *GRPCServerInitializer) AuthService() *service.AuthService {
	return g.authService
}

// UserService 返回共享的 UserService 实例
func (g *GRPCServerInitializer) UserService() *service.UserService {
	return g.userService
}

// RoleService 返回共享的 RoleService 实例
func (g *GRPCServerInitializer) RoleService() *service.RoleService {
	return g.roleService
}

// PermissionService 返回共享的 PermissionService 实例
func (g *GRPCServerInitializer) PermissionService() *service.PermissionService {
	return g.permissionService
}

// GetMySQLDB 返回共享的 MySQLDB 包装实例
func (g *GRPCServerInitializer) GetMySQLDB() *storage.MySQLDB {
	return g.mysqlDB
}

// GetRedisClient 返回共享的 RedisClient 包装实例
func (g *GRPCServerInitializer) GetRedisClient() *storage.RedisClient {
	return g.redisClient
}
