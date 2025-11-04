// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package initializers

import (
	"context"
	"fmt"

	commonoptions "github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/k8s-agent/internal/cluster/api"
	"github.com/kart-io/k8s-agent/internal/cluster/handler"
	"github.com/kart-io/k8s-agent/internal/cluster/service"
	"github.com/kart-io/logger/core"
)

// HTTPServerInitializer initializes the HTTP server with all services and handlers.
type HTTPServerInitializer struct {
	serverOpts *commonoptions.ServerOptions
	jwtOpts    *commonoptions.JWTOptions
	logger     core.Logger
	dbInit     *DatabaseInitializer
	server     *api.Server
}

// NewHTTPServerInitializer creates a new HTTP server initializer.
func NewHTTPServerInitializer(
	serverOpts *commonoptions.ServerOptions,
	jwtOpts *commonoptions.JWTOptions,
	logger core.Logger,
	dbInit *DatabaseInitializer,
) *HTTPServerInitializer {
	return &HTTPServerInitializer{
		serverOpts: serverOpts,
		jwtOpts:    jwtOpts,
		logger:     logger,
		dbInit:     dbInit,
	}
}

// Initialize initializes the HTTP server.
func (i *HTTPServerInitializer) Initialize(ctx context.Context) error {
	i.logger.Infow("Initializing HTTP server",
		"host", i.serverOpts.Host,
		"port", i.serverOpts.Port,
		"mode", i.serverOpts.Mode,
	)

	// 获取数据库存储实例
	storage := i.dbInit.GetStorage()
	if storage == nil {
		return fmt.Errorf("database storage not initialized")
	}

	// 初始化所有服务和处理器
	if err := i.initializeServices(storage); err != nil {
		return fmt.Errorf("failed to initialize services: %w", err)
	}

	i.logger.Infow("HTTP server initialized",
		"port", i.serverOpts.Port,
		"mode", i.serverOpts.Mode,
	)

	return nil
}

// initializeServices 初始化所有服务层和处理器.
func (i *HTTPServerInitializer) initializeServices(storage interface{}) error {
	// 使用实际的存储实例
	mysqlStorage := i.dbInit.GetStorage()

	// 初始化基础服务
	clusterService := service.NewClusterService(mysqlStorage, i.logger)
	clusterHandler := handler.NewClusterHandler(clusterService, i.logger)

	// 初始化 K8s API 相关服务
	k8sClusterService := service.NewK8sClusterService(mysqlStorage)
	k8sNamespaceService := service.NewK8sNamespaceService(mysqlStorage, k8sClusterService)
	k8sPodService := service.NewK8sPodService(mysqlStorage, k8sClusterService)
	k8sDeploymentService := service.NewK8sDeploymentService(mysqlStorage, k8sClusterService)
	k8sNodeService := service.NewK8sNodeService(mysqlStorage, k8sClusterService)
	k8sServiceService := service.NewK8sServiceService(mysqlStorage, k8sClusterService)
	k8sStatefulSetService := service.NewK8sStatefulSetService(mysqlStorage, k8sClusterService)
	k8sDaemonSetService := service.NewK8sDaemonSetService(mysqlStorage, k8sClusterService)
	k8sConfigMapService := service.NewK8sConfigMapService(mysqlStorage, k8sClusterService)
	k8sSecretService := service.NewK8sSecretService(mysqlStorage, k8sClusterService)
	k8sEndpointService := service.NewK8sEndpointService(mysqlStorage, k8sClusterService)
	k8sPVCService := service.NewK8sPVCService(mysqlStorage, k8sClusterService)
	k8sPVService := service.NewK8sPVService(mysqlStorage, k8sClusterService)
	k8sEndpointSliceService := service.NewK8sEndpointSliceService(mysqlStorage, k8sClusterService)
	k8sHPAService := service.NewK8sHPAService(mysqlStorage, k8sClusterService)
	k8sEventService := service.NewK8sEventService(mysqlStorage, k8sClusterService)
	k8sRoleBindingService := service.NewK8sRoleBindingService(mysqlStorage, k8sClusterService)
	k8sClusterRoleService := service.NewK8sClusterRoleService(mysqlStorage, k8sClusterService)
	k8sPriorityClassService := service.NewK8sPriorityClassService(mysqlStorage, k8sClusterService)
	k8sRoleService := service.NewK8sRoleService(mysqlStorage, k8sClusterService)
	k8sStorageClassService := service.NewK8sStorageClassService(mysqlStorage, k8sClusterService)
	k8sJobService := service.NewK8sJobService(mysqlStorage, k8sClusterService)
	k8sCronJobService := service.NewK8sCronJobService(mysqlStorage, k8sClusterService)
	k8sIngressService := service.NewK8sIngressService(mysqlStorage, k8sClusterService)
	k8sNetworkPolicyService := service.NewK8sNetworkPolicyService(mysqlStorage, k8sClusterService)
	k8sReplicaSetService := service.NewK8sReplicaSetService(mysqlStorage, k8sClusterService)
	k8sLimitRangeService := service.NewK8sLimitRangeService(mysqlStorage, k8sClusterService)
	k8sServiceAccountService := service.NewK8sServiceAccountService(mysqlStorage, k8sClusterService)
	k8sClusterRoleBindingService := service.NewK8sClusterRoleBindingService(mysqlStorage, k8sClusterService)
	k8sResourceQuotaService := service.NewK8sResourceQuotaService(mysqlStorage, k8sClusterService)

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

	i.logger.Infow("K8s API endpoints initialized",
		"count", 25,
		"base_path", "/api/k8s",
	)

	// 创建服务器配置
	serverConfig := &api.ServerConfig{
		Port:         i.serverOpts.Port,
		Mode:         i.serverOpts.Mode,
		ReadTimeout:  i.serverOpts.ReadTimeout,
		WriteTimeout: i.serverOpts.WriteTimeout,
		JWTSecret:    i.jwtOpts.Secret,
	}

	// 创建服务器实例
	i.server = api.NewServer(serverConfig, clusterHandler, k8sAPIHandler, i.logger)

	return nil
}

// Shutdown stops the HTTP server.
func (i *HTTPServerInitializer) Shutdown(ctx context.Context) error {
	i.logger.Info("Shutting down HTTP server")
	if i.server != nil {
		return i.server.Shutdown(ctx)
	}
	return nil
}

// Priority returns the initialization priority (higher = earlier).
func (i *HTTPServerInitializer) Priority() int {
	return 500 // HTTP server should be initialized after database
}

// Name returns the name of this initializer.
func (i *HTTPServerInitializer) Name() string {
	return "HTTPServer"
}

// GetServer returns the initialized server instance.
func (i *HTTPServerInitializer) GetServer() *api.Server {
	return i.server
}

// Start starts the HTTP server.
func (i *HTTPServerInitializer) Start() error {
	if i.server == nil {
		return fmt.Errorf("server not initialized")
	}
	return i.server.Start()
}
