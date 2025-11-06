package initializers

import (
	"context"

	"google.golang.org/grpc"

	"github.com/kart-io/k8s-agent/cmd/auth/app/options"
	commonserver "github.com/kart-io/k8s-agent/common/server"
	authv1 "github.com/kart-io/k8s-agent/pkg/api/auth/v1"
	"github.com/kart-io/k8s-agent/pkg/bootstrap"
	commoninitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// GRPCServerInitializer gRPC服务器初始化器
type GRPCServerInitializer struct {
	standardInit *commoninitializers.GRPCServerInitializer
	opts         *options.ServerOptions
	logger       core.Logger

	// 依赖的初始化器
	dbInit      *DatabaseInitializer
	redisInit   *RedisInitializer
	sessionInit *SessionServiceInitializer
}

// NewGRPCServerInitializer 创建gRPC服务器初始化器
func NewGRPCServerInitializer(
	opts *options.ServerOptions,
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

	serverConfig := &commoninitializers.GRPCServerConfig{
		Name:     g.Name(),
		Priority: g.Priority(),
		Config:   g.opts.GRPC,
		ServiceRegister: func(s *grpc.Server) error {
			// TODO: 注册gRPC服务
			// authv1.RegisterAuthServiceServer(s, authService)
			// authv1.RegisterUserServiceServer(s, userService)
			// authv1.RegisterRoleServiceServer(s, roleService)
			// authv1.RegisterPermissionServiceServer(s, permissionService)
			// authv1.RegisterSessionServiceServer(s, sessionService)

			g.logger.Info("gRPC services registered (TODO: implement service handlers)")
			_ = authv1.UnimplementedAuthServiceServer{}
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
