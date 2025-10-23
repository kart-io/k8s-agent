package app

import (
	"context"
	"fmt"

	commonlogger "github.com/kart-io/k8s-agent/common/logger"
	"github.com/kart-io/k8s-agent/internal/cluster/api"
	clusterconfig "github.com/kart-io/k8s-agent/internal/cluster/config"
	"github.com/kart-io/k8s-agent/internal/cluster/handler"
	"github.com/kart-io/k8s-agent/internal/cluster/service"
	"github.com/kart-io/k8s-agent/internal/cluster/storage"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/logger/core"
)

// Execute runs the cluster service command
func Execute() {
	// 创建配置选项
	opts := clusterconfig.NewOptions()

	// 使用组合框架运行应用
	commonapp.RunWithRunner(
		opts,
		&ClusterApp{},
		initLogger,
		commonapp.CommandConfig{
			Use:       "cluster",
			Short:     "Cluster Service",
			Long:      "Cluster Service provides multi-cluster management and K8s resource API",
			EnvPrefix: "CLUSTER",
		},
	)
}

// ClusterApp 实现 commonapp.Application 接口
type ClusterApp struct {
	opts    *clusterconfig.Options
	logger  core.Logger
	storage *storage.MySQLStorage
	server  *api.Server
}

// Initialize 初始化应用程序
func (a *ClusterApp) Initialize(ctx context.Context, opts commonapp.Options) error {
	a.opts = opts.(*clusterconfig.Options)

	// 初始化日志系统
	logger, err := initLogger(opts)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	a.logger = logger

	a.logger.Infow("Initializing Cluster Service",
		"host", a.opts.Server.Host,
		"port", a.opts.Server.Port,
	)

	// 初始化数据库
	a.storage, err = storage.NewMySQLStorage(&storage.Config{
		Host:         a.opts.Database.Host,
		Port:         a.opts.Database.Port,
		User:         a.opts.Database.User,
		Password:     a.opts.Database.Password,
		DBName:       a.opts.Database.Database,
		SSLMode:      a.opts.Database.SSLMode,
		MaxOpenConns: a.opts.Database.MaxOpenConns,
		MaxIdleConns: a.opts.Database.MaxIdleConns,
	}, a.logger)
	if err != nil {
		return fmt.Errorf("failed to initialize MySQL storage: %w", err)
	}

	a.logger.Infow("Database connected",
		"host", a.opts.Database.Host,
		"port", a.opts.Database.Port,
		"database", a.opts.Database.Database,
	)

	// 初始化数据库 schema
	if err := a.storage.InitSchema(); err != nil {
		return fmt.Errorf("failed to initialize database schema: %w", err)
	}
	a.logger.Info("Database schema initialized successfully")

	// 初始化所有服务和处理器
	if err := a.initializeServices(); err != nil {
		return fmt.Errorf("failed to initialize services: %w", err)
	}

	a.logger.Infow("All services initialized successfully")
	return nil
}

// initializeServices 初始化所有服务层和处理器
func (a *ClusterApp) initializeServices() error {
	// 初始化基础服务
	clusterService := service.NewClusterService(a.storage, a.logger)
	clusterHandler := handler.NewClusterHandler(clusterService, a.logger)

	// 初始化 K8s API 相关服务
	k8sClusterService := service.NewK8sClusterService(a.storage)
	k8sNamespaceService := service.NewK8sNamespaceService(a.storage, k8sClusterService)
	k8sPodService := service.NewK8sPodService(a.storage, k8sClusterService)
	k8sDeploymentService := service.NewK8sDeploymentService(a.storage, k8sClusterService)
	k8sNodeService := service.NewK8sNodeService(a.storage, k8sClusterService)
	k8sServiceService := service.NewK8sServiceService(a.storage, k8sClusterService)
	k8sStatefulSetService := service.NewK8sStatefulSetService(a.storage, k8sClusterService)
	k8sDaemonSetService := service.NewK8sDaemonSetService(a.storage, k8sClusterService)
	k8sConfigMapService := service.NewK8sConfigMapService(a.storage, k8sClusterService)
	k8sSecretService := service.NewK8sSecretService(a.storage, k8sClusterService)
	k8sEndpointService := service.NewK8sEndpointService(a.storage, k8sClusterService)
	k8sPVCService := service.NewK8sPVCService(a.storage, k8sClusterService)
	k8sPVService := service.NewK8sPVService(a.storage, k8sClusterService)
	k8sEndpointSliceService := service.NewK8sEndpointSliceService(a.storage, k8sClusterService)
	k8sHPAService := service.NewK8sHPAService(a.storage, k8sClusterService)
	k8sEventService := service.NewK8sEventService(a.storage, k8sClusterService)
	k8sRoleBindingService := service.NewK8sRoleBindingService(a.storage, k8sClusterService)
	k8sClusterRoleService := service.NewK8sClusterRoleService(a.storage, k8sClusterService)
	k8sPriorityClassService := service.NewK8sPriorityClassService(a.storage, k8sClusterService)
	k8sRoleService := service.NewK8sRoleService(a.storage, k8sClusterService)
	k8sStorageClassService := service.NewK8sStorageClassService(a.storage, k8sClusterService)
	k8sJobService := service.NewK8sJobService(a.storage, k8sClusterService)
	k8sCronJobService := service.NewK8sCronJobService(a.storage, k8sClusterService)
	k8sIngressService := service.NewK8sIngressService(a.storage, k8sClusterService)
	k8sNetworkPolicyService := service.NewK8sNetworkPolicyService(a.storage, k8sClusterService)
	k8sReplicaSetService := service.NewK8sReplicaSetService(a.storage, k8sClusterService)
	k8sLimitRangeService := service.NewK8sLimitRangeService(a.storage, k8sClusterService)
	k8sServiceAccountService := service.NewK8sServiceAccountService(a.storage, k8sClusterService)
	k8sClusterRoleBindingService := service.NewK8sClusterRoleBindingService(a.storage, k8sClusterService)
	k8sResourceQuotaService := service.NewK8sResourceQuotaService(a.storage, k8sClusterService)

	// 初始化 K8s API 处理器
	k8sAPIHandler := handler.NewK8sAPIHandler(
		k8sClusterService,
		k8sNamespaceService,
		k8sPodService,
		k8sDeploymentService,
		k8sNodeService,
		k8sServiceService,
		k8sStatefulSetService,
		k8sDaemonSetService,
		k8sConfigMapService,
		k8sSecretService,
		k8sEndpointService,
		k8sPVCService,
		k8sPVService,
		k8sEndpointSliceService,
		k8sHPAService,
		k8sEventService,
		k8sRoleBindingService,
		k8sClusterRoleService,
		k8sPriorityClassService,
		k8sRoleService,
		k8sStorageClassService,
		k8sJobService,
		k8sCronJobService,
		k8sIngressService,
		k8sNetworkPolicyService,
		k8sReplicaSetService,
		k8sLimitRangeService,
		k8sServiceAccountService,
		k8sClusterRoleBindingService,
		k8sResourceQuotaService,
	)

	// 创建服务器配置
	serverConfig := &api.ServerConfig{
		Port:         a.opts.Server.Port,
		Mode:         a.opts.Server.Mode,
		ReadTimeout:  a.opts.Server.ReadTimeout,
		WriteTimeout: a.opts.Server.WriteTimeout,
		JWTSecret:    a.opts.JWT.Secret,
	}

	// 创建服务器实例
	a.server = api.NewServer(serverConfig, clusterHandler, k8sAPIHandler, a.logger)

	a.logger.Infow("K8s API endpoints initialized",
		"count", 25,
		"base_path", "/api/k8s",
	)

	return nil
}

// Run 运行应用程序主逻辑
func (a *ClusterApp) Run(ctx context.Context) error {
	// 启动服务器
	go func() {
		if err := a.server.Start(); err != nil {
			a.logger.Fatalw("Server failed to start", "error", err)
		}
	}()

	a.logger.Infow("Cluster service started successfully",
		"port", a.opts.Server.Port,
		"mode", a.opts.Server.Mode,
	)

	// 等待 context 取消
	<-ctx.Done()

	a.logger.Info("Received shutdown signal")
	return nil
}

// Shutdown 优雅关闭应用程序
func (a *ClusterApp) Shutdown(ctx context.Context) error {
	a.logger.Infow("Shutting down Cluster Service")

	// 关闭服务器
	if a.server != nil {
		if err := a.server.Shutdown(ctx); err != nil {
			a.logger.Errorw("Server forced to shutdown", "error", err)
			return err
		}
	}

	// 关闭数据库连接
	if a.storage != nil {
		a.storage.Close()
	}

	a.logger.Info("Cluster Service shutdown complete")
	return nil
}

// initLogger 初始化日志系统
func initLogger(opts commonapp.Options) (core.Logger, error) {
	cfg := opts.(*clusterconfig.Options)
	return commonlogger.InitFromOptions(cfg.Logging)
}
