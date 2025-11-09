// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/gorm"

	"github.com/kart-io/k8s-agent/common/db"
	"github.com/kart-io/k8s-agent/internal/cluster/handler"
	"github.com/kart-io/k8s-agent/internal/cluster/service"
	"github.com/kart-io/k8s-agent/internal/cluster/storage"
	clusterv1 "github.com/kart-io/k8s-agent/pkg/api/cluster/v1"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/logger/core"
)

const (
	// UserAgent is the User-Agent string for the cluster service.
	UserAgent = "aetherius-cluster"
)

// Execute runs the cluster service command.
func Execute() {
	opts := commonapp.NewStandardOptions("Cluster", UserAgent).
		WithDatabase().
		WithJWT()

	app := &ClusterApp{}

	commonapp.Run(
		app,
		opts,
		commonapp.Config{
			Use:       "cluster",
			Short:     "Cluster Service",
			Long:      "Cluster Service provides multi-cluster management and K8s resource API",
			EnvPrefix: "CLUSTER",
		},
	)
}

// ClusterApp implements commonapp.Application interface.
// This is the SIMPLE pattern: no Bootstrap, no Wire, direct initialization.
type ClusterApp struct {
	opts   *commonapp.StandardOptions
	logger core.Logger

	// Dependencies
	db                 *gorm.DB
	mysqlClient        *db.MySQLClient
	clusterService     *service.ClusterService
	k8sServiceRegistry *service.K8sServiceRegistry

	// Servers
	httpServer *http.Server
	grpcServer *grpc.Server
}

// Name returns the application name.
func (a *ClusterApp) Name() string {
	return "Cluster Service"
}

// Initialize initializes the application components directly.
func (a *ClusterApp) Initialize(ctx context.Context, opts commonapp.Options) error {
	a.opts = opts.(*commonapp.StandardOptions)

	// Initialize logger
	logger, err := a.opts.InitLogger()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	a.logger = logger

	a.logger.Infow("Initializing Cluster Service",
		"http_port", a.opts.Server.Port,
		"grpc_port", a.opts.GRPC.Port,
		"grpc_enabled", a.opts.GRPC.Enable,
	)

	// Initialize database
	if err := a.initDatabase(ctx); err != nil {
		return err
	}

	// Initialize business services
	if err := a.initServices(ctx); err != nil {
		return err
	}

	// Initialize HTTP server
	if err := a.initHTTPServer(ctx); err != nil {
		return err
	}

	// Initialize gRPC server if enabled
	if a.opts.GRPC.Enable {
		if err := a.initGRPCServer(ctx); err != nil {
			return err
		}
	}

	a.logger.Infow("Cluster Service initialized successfully")
	return nil
}

// Run runs the application.
func (a *ClusterApp) Run(ctx context.Context) error {
	a.logger.Infow("Cluster Service running")

	// Wait for context cancellation
	<-ctx.Done()
	return nil
}

// Shutdown gracefully shuts down the application.
func (a *ClusterApp) Shutdown(ctx context.Context) error {
	a.logger.Infow("Shutting down Cluster Service...")

	// Shutdown HTTP server
	if a.httpServer != nil {
		if err := a.httpServer.Shutdown(ctx); err != nil {
			a.logger.Errorw("HTTP server shutdown error", "error", err)
		}
	}

	// Shutdown gRPC server
	if a.grpcServer != nil {
		a.grpcServer.GracefulStop()
		a.logger.Infow("gRPC server stopped")
	}

	// Close database connection
	if a.mysqlClient != nil {
		if err := a.mysqlClient.Close(); err != nil {
			a.logger.Errorw("Failed to close database", "error", err)
		}
	}

	a.logger.Infow("Cluster Service stopped")
	return nil
}

// initDatabase initializes the database connection.
func (a *ClusterApp) initDatabase(ctx context.Context) error {
	a.logger.Infow("Initializing database connection",
		"host", a.opts.Database.Host,
		"port", a.opts.Database.Port,
		"database", a.opts.Database.Database,
	)

	// Create MySQL client using common/db package
	client, err := db.NewMySQL(a.logger,
		db.WithHost(a.opts.Database.Host),
		db.WithPort(a.opts.Database.Port),
		db.WithUser(a.opts.Database.User),
		db.WithPassword(a.opts.Database.Password),
		db.WithDatabase(a.opts.Database.Database),
		db.WithMaxOpenConns(a.opts.Database.MaxOpenConns),
		db.WithMaxIdleConns(a.opts.Database.MaxIdleConns),
		db.WithConnMaxLifetime(a.opts.Database.ConnMaxLifetime),
		db.WithLogLevel(a.opts.Database.LogLevel),
	)
	if err != nil {
		return fmt.Errorf("failed to create MySQL client: %w", err)
	}

	a.mysqlClient = client
	a.db = client.DB

	a.logger.Infow("Database connected successfully")
	return nil
}

// initServices initializes business services.
func (a *ClusterApp) initServices(ctx context.Context) error {
	a.logger.Infow("Initializing business services")

	// Initialize storage layer
	store, err := storage.NewMySQLStorageWithDB(a.db, a.logger)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Initialize business services
	a.clusterService = service.NewClusterService(store, a.logger)
	a.k8sServiceRegistry = service.NewK8sServiceRegistry(store)

	a.logger.Infow("Services initialized",
		"k8s_services_count", a.k8sServiceRegistry.Count(),
	)

	return nil
}

// initHTTPServer initializes the HTTP server.
func (a *ClusterApp) initHTTPServer(ctx context.Context) error {
	// Create Gin engine
	if a.opts.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()

	// Add middleware
	engine.Use(gin.Recovery())
	engine.Use(gin.Logger())

	// Initialize handlers
	clusterHandler := handler.NewClusterHandler(a.clusterService, a.logger)
	k8sAPIHandler := handler.NewK8sAPIHandler(a.k8sServiceRegistry)
	versionHandler := handler.NewVersionHandler()

	// Setup routes
	a.setupHTTPRoutes(engine, clusterHandler, k8sAPIHandler, versionHandler)

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", a.opts.Server.Host, a.opts.Server.Port)
	a.httpServer = &http.Server{
		Addr:    addr,
		Handler: engine,
	}

	// Start server in goroutine
	go func() {
		a.logger.Infow("HTTP server starting", "address", addr)
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatalw("HTTP server failed", "error", err)
		}
	}()

	return nil
}

// setupHTTPRoutes configures all HTTP routes.
func (a *ClusterApp) setupHTTPRoutes(
	engine *gin.Engine,
	clusterHandler *handler.ClusterHandler,
	k8sAPIHandler *handler.K8sAPIHandler,
	versionHandler *handler.VersionHandler,
) {
	// Health check
	engine.GET("/health", clusterHandler.HealthCheck)

	// Version endpoints
	version := engine.Group("/version")
	{
		version.GET("", versionHandler.GetVersion)
		version.GET("/simple", versionHandler.GetVersionSimple)
		version.GET("/text", versionHandler.GetVersionText)
	}

	// API v1 routes
	v1 := engine.Group("/api/v1")
	{
		// Cluster management
		clusters := v1.Group("/clusters")
		{
			clusters.POST("", clusterHandler.AddCluster)
			clusters.GET("/:id/health", clusterHandler.GetClusterHealth)
			clusters.GET("/:cluster_id/namespaces/:namespace/pods", clusterHandler.GetPods)
		}

		// K8s resource API
		// TODO: Register dynamic K8s API routes when RegisterRoutes method is implemented
		k8s := v1.Group("/k8s")
		{
			_ = k8s
			_ = k8sAPIHandler
			// k8sAPIHandler.RegisterRoutes(k8s) // TODO: Implement this method
		}
	}
}

// initGRPCServer initializes the gRPC server.
func (a *ClusterApp) initGRPCServer(ctx context.Context) error {
	// Create gRPC server
	a.grpcServer = grpc.NewServer()

	// Register Cluster service
	clusterGRPCService := &ClusterGRPCService{
		clusterService: a.clusterService,
		logger:         a.logger,
	}
	clusterv1.RegisterClusterServiceServer(a.grpcServer, clusterGRPCService)

	// Register K8s Resource service (note: capital S in K8S)
	k8sGRPCService := &K8sResourceGRPCService{
		registry: a.k8sServiceRegistry,
		logger:   a.logger,
	}
	clusterv1.RegisterK8SResourceServiceServer(a.grpcServer, k8sGRPCService)

	// Register reflection service (useful for debugging)
	reflection.Register(a.grpcServer)

	// Start server in goroutine
	go func() {
		addr := fmt.Sprintf("%s:%d", a.opts.GRPC.Host, a.opts.GRPC.Port)
		a.logger.Infow("gRPC server starting", "address", addr)

		listener, err := net.Listen("tcp", addr)
		if err != nil {
			a.logger.Fatalw("Failed to create gRPC listener", "error", err)
		}

		if err := a.grpcServer.Serve(listener); err != nil {
			a.logger.Fatalw("gRPC server failed", "error", err)
		}
	}()

	return nil
}

// ============================================================================
// gRPC Service Implementations
// ============================================================================

// ClusterGRPCService implements the gRPC ClusterService interface.
type ClusterGRPCService struct {
	clusterv1.UnimplementedClusterServiceServer
	clusterService *service.ClusterService
	logger         core.Logger
}

// GetCluster retrieves cluster information by ID.
func (s *ClusterGRPCService) GetCluster(ctx context.Context, req *clusterv1.GetClusterRequest) (*clusterv1.Cluster, error) {
	s.logger.Infow("GetCluster RPC called", "cluster_id", req.ClusterId)
	// TODO: Call clusterService to get cluster details
	return &clusterv1.Cluster{
		Id:   req.ClusterId,
		Name: "TODO: Implement",
	}, nil
}

// ListClusters lists all clusters.
func (s *ClusterGRPCService) ListClusters(ctx context.Context, req *clusterv1.ListClustersRequest) (*clusterv1.ListClustersResponse, error) {
	s.logger.Infow("ListClusters RPC called")
	// TODO: Implement listing logic
	return &clusterv1.ListClustersResponse{
		Clusters: []*clusterv1.Cluster{},
	}, nil
}

// CreateCluster creates a new cluster.
func (s *ClusterGRPCService) CreateCluster(ctx context.Context, req *clusterv1.CreateClusterRequest) (*clusterv1.Cluster, error) {
	s.logger.Infow("CreateCluster RPC called", "name", req.Name)
	// TODO: Implement creation logic
	return &clusterv1.Cluster{
		Name: req.Name,
	}, nil
}

// UpdateCluster updates an existing cluster.
func (s *ClusterGRPCService) UpdateCluster(ctx context.Context, req *clusterv1.UpdateClusterRequest) (*clusterv1.Cluster, error) {
	s.logger.Infow("UpdateCluster RPC called", "cluster_id", req.ClusterId)
	// TODO: Implement update logic
	return &clusterv1.Cluster{
		Id: req.ClusterId,
	}, nil
}

// DeleteCluster deletes a cluster.
func (s *ClusterGRPCService) DeleteCluster(ctx context.Context, req *clusterv1.DeleteClusterRequest) (*clusterv1.DeleteClusterResponse, error) {
	s.logger.Infow("DeleteCluster RPC called", "cluster_id", req.ClusterId)
	// TODO: Implement deletion logic
	return &clusterv1.DeleteClusterResponse{
		Message: "Cluster deleted successfully",
	}, nil
}

// GetClusterHealth retrieves cluster health status.
func (s *ClusterGRPCService) GetClusterHealth(ctx context.Context, req *clusterv1.GetClusterHealthRequest) (*clusterv1.ClusterHealth, error) {
	s.logger.Infow("GetClusterHealth RPC called", "cluster_id", req.ClusterId)
	health, err := s.clusterService.GetClusterHealth(ctx, req.ClusterId)
	if err != nil {
		return nil, err
	}
	_ = health // TODO: Convert to Proto message
	return &clusterv1.ClusterHealth{
		Status: clusterv1.ClusterStatus_HEALTHY,
	}, nil
}

// GetClusterVersion retrieves cluster version information.
func (s *ClusterGRPCService) GetClusterVersion(ctx context.Context, req *clusterv1.GetClusterVersionRequest) (*clusterv1.ClusterVersion, error) {
	s.logger.Infow("GetClusterVersion RPC called", "cluster_id", req.ClusterId)
	// TODO: Implement version query
	return &clusterv1.ClusterVersion{
		KubernetesVersion: "TODO",
	}, nil
}

// K8sResourceGRPCService implements the gRPC K8sResourceService interface.
type K8sResourceGRPCService struct {
	clusterv1.UnimplementedK8SResourceServiceServer
	registry *service.K8sServiceRegistry
	logger   core.Logger
}

// GetResource retrieves a Kubernetes resource.
func (s *K8sResourceGRPCService) GetResource(ctx context.Context, req *clusterv1.GetResourceRequest) (*clusterv1.Resource, error) {
	s.logger.Infow("GetResource RPC called",
		"cluster_id", req.ClusterId,
		"type", req.ResourceType,
		"namespace", req.Namespace,
		"name", req.Name,
	)
	// TODO: Implement resource retrieval
	return &clusterv1.Resource{
		ClusterId: req.ClusterId,
		Type:      req.ResourceType,
		Namespace: req.Namespace,
		Name:      req.Name,
	}, nil
}

// ListResources lists Kubernetes resources.
func (s *K8sResourceGRPCService) ListResources(ctx context.Context, req *clusterv1.ListResourcesRequest) (*clusterv1.ListResourcesResponse, error) {
	s.logger.Infow("ListResources RPC called",
		"cluster_id", req.ClusterId,
		"type", req.ResourceType,
	)
	// TODO: Implement resource listing
	return &clusterv1.ListResourcesResponse{
		Resources: []*clusterv1.Resource{},
	}, nil
}

// CreateResource creates a new Kubernetes resource.
func (s *K8sResourceGRPCService) CreateResource(ctx context.Context, req *clusterv1.CreateResourceRequest) (*clusterv1.Resource, error) {
	s.logger.Infow("CreateResource RPC called",
		"cluster_id", req.ClusterId,
		"type", req.ResourceType,
	)
	// TODO: Implement resource creation
	return &clusterv1.Resource{
		ClusterId: req.ClusterId,
		Type:      req.ResourceType,
	}, nil
}

// UpdateResource updates a Kubernetes resource.
func (s *K8sResourceGRPCService) UpdateResource(ctx context.Context, req *clusterv1.UpdateResourceRequest) (*clusterv1.Resource, error) {
	s.logger.Infow("UpdateResource RPC called",
		"cluster_id", req.ClusterId,
		"type", req.ResourceType,
		"name", req.Name,
	)
	// TODO: Implement resource update
	return &clusterv1.Resource{
		ClusterId: req.ClusterId,
		Type:      req.ResourceType,
		Name:      req.Name,
	}, nil
}

// DeleteResource deletes a Kubernetes resource.
func (s *K8sResourceGRPCService) DeleteResource(ctx context.Context, req *clusterv1.DeleteResourceRequest) (*clusterv1.DeleteResourceResponse, error) {
	s.logger.Infow("DeleteResource RPC called",
		"cluster_id", req.ClusterId,
		"type", req.ResourceType,
		"name", req.Name,
	)
	// TODO: Implement resource deletion
	return &clusterv1.DeleteResourceResponse{
		Message: "Resource deleted successfully",
	}, nil
}

// WatchResources watches Kubernetes resource changes.
func (s *K8sResourceGRPCService) WatchResources(req *clusterv1.WatchResourcesRequest, stream clusterv1.K8SResourceService_WatchResourcesServer) error {
	s.logger.Infow("WatchResources RPC called",
		"cluster_id", req.ClusterId,
		"type", req.ResourceType,
	)
	// TODO: Implement streaming resource watch
	return nil
}
