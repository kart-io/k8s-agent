// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/gorm"

	clustergrpc "github.com/kart-io/k8s-agent/internal/cluster/grpc"
	"github.com/kart-io/k8s-agent/internal/cluster/handler"
	"github.com/kart-io/k8s-agent/internal/cluster/service"
	clustermodel "github.com/kart-io/k8s-agent/internal/models/cluster"
	clusterv1 "github.com/kart-io/k8s-agent/pkg/api/cluster/v1"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
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

	// Dependencies - using pkg/initializers
	dbInit             *pkginitializers.DatabaseInitializer
	db                 *gorm.DB
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

	// Close database connection via pkg/initializers
	if a.dbInit != nil {
		if err := a.dbInit.Close(ctx); err != nil {
			a.logger.Errorw("Failed to close database", "error", err)
		}
	}

	a.logger.Infow("Cluster Service stopped")
	return nil
}

// initDatabase initializes the database connection using pkg/initializers.
func (a *ClusterApp) initDatabase(ctx context.Context) error {
	a.logger.Infow("Initializing database connection",
		"host", a.opts.Database.Host,
		"port", a.opts.Database.Port,
		"database", a.opts.Database.Database,
	)

	// Create database initializer from pkg/initializers with GORM AutoMigrate
	a.dbInit = pkginitializers.NewDatabaseInitializer(a.opts.Database, a.logger).
		WithAutoMigrate(&clustermodel.Cluster{})

	// Initialize the database
	if err := a.dbInit.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Get GORM DB instance
	a.db = a.dbInit.DB()
	if a.db == nil {
		return fmt.Errorf("failed to get GORM DB instance")
	}

	a.logger.Infow("Database initialized successfully")
	return nil
}

// initServices initializes business services with direct GORM DB access.
func (a *ClusterApp) initServices(ctx context.Context) error {
	a.logger.Infow("Initializing business services")

	// Create services with direct GORM DB access (no storage wrapper)
	a.clusterService = service.NewClusterService(a.db, a.logger)
	a.k8sServiceRegistry = service.NewK8sServiceRegistryWithDB(a.db, a.logger)

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
		Addr:              addr,
		Handler:           engine,
		ReadHeaderTimeout: 10 * time.Second, // Prevent Slowloris attacks
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
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

	// Register Cluster service using the grpc package
	clusterGRPCService := clustergrpc.NewClusterGRPCService(a.clusterService, a.logger)
	clusterv1.RegisterClusterServiceServer(a.grpcServer, clusterGRPCService)

	// Register K8s Resource service using the grpc package
	k8sGRPCService := clustergrpc.NewK8sResourceGRPCService(a.k8sServiceRegistry, a.logger)
	clusterv1.RegisterK8SResourceServiceServer(a.grpcServer, k8sGRPCService)

	// Register reflection service (useful for debugging)
	reflection.Register(a.grpcServer)

	// Start server in goroutine
	go func() {
		addr := fmt.Sprintf("%s:%d", a.opts.GRPC.Host, a.opts.GRPC.Port)
		a.logger.Infow("gRPC server starting", "address", addr)

		lc := net.ListenConfig{}
		listener, err := lc.Listen(ctx, "tcp", addr)
		if err != nil {
			a.logger.Fatalw("Failed to create gRPC listener", "error", err)
		}

		if err := a.grpcServer.Serve(listener); err != nil {
			a.logger.Fatalw("gRPC server failed", "error", err)
		}
	}()

	return nil
}
