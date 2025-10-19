package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kart-io/k8s-agent/cluster-service/internal/api"
	"github.com/kart-io/k8s-agent/cluster-service/internal/config"
	"github.com/kart-io/k8s-agent/cluster-service/internal/handler"
	"github.com/kart-io/k8s-agent/cluster-service/internal/service"
	"github.com/kart-io/k8s-agent/cluster-service/internal/storage"
	"github.com/kart-io/k8s-agent/common/logger"
	"github.com/kart-io/version"
	"github.com/sirupsen/logrus"
)

func main() {
	// Parse command-line flags
	var configPath string
	var enableK8sAPI bool
	flag.StringVar(&configPath, "config", "", "Path to configuration file (defaults to ./configs/config.yaml)")
	flag.StringVar(&configPath, "c", "", "Path to configuration file (shorthand)")
	flag.BoolVar(&enableK8sAPI, "enable-k8s-api", true, "Enable full K8s API endpoints (default: true)")
	flag.Parse()

	// Load configuration from config file
	var cfg *config.Config
	var err error

	if configPath != "" {
		cfg, err = config.LoadFromPath(configPath)
		if err != nil {
			log.Fatalf("Failed to load configuration from %s: %v", configPath, err)
		}
		log.Printf("Loaded configuration from: %s", configPath)
	} else {
		cfg, err = config.Load()
		if err != nil {
			log.Fatalf("Failed to load configuration: %v", err)
		}
	}

	// 初始化 logrus logger（用于兼容现有代码）
	logrusLogger := setupLogrusLogger(cfg.Logging.Level, cfg.Logging.Format)

	// 初始化 common/logger（用于新代码）
	if err := initCommonLogger(cfg); err != nil {
		logrusLogger.WithError(err).Fatal("Failed to initialize common logger")
	}

	// 获取版本信息
	versionInfo := version.Get()

	logger.Infow("Application starting",
		"service", versionInfo.ServiceName,
		"version", versionInfo.GitVersion,
		"commit", versionInfo.GitCommit,
		"branch", versionInfo.GitBranch,
		"build_date", versionInfo.BuildDate,
		"config_path", configPath,
		"enable_k8s_api", enableK8sAPI,
	)

	// 初始化数据库
	pgStorage, err := storage.NewMySQLStorage(&storage.Config{
		Host:         cfg.Database.Host,
		Port:         cfg.Database.Port,
		User:         cfg.Database.User,
		Password:     cfg.Database.Password,
		DBName:       cfg.Database.DBName,
		SSLMode:      cfg.Database.SSLMode,
		MaxOpenConns: cfg.Database.MaxOpenConns,
		MaxIdleConns: cfg.Database.MaxIdleConns,
	}, logrusLogger)
	if err != nil {
		logger.Fatalw("Failed to initialize PostgreSQL storage", "error", err.Error())
	}
	defer pgStorage.Close()

	logger.Infow("Database connected",
		"host", cfg.Database.Host,
		"port", cfg.Database.Port,
		"database", cfg.Database.DBName,
	)

	// 初始化数据库 schema
	if err := pgStorage.InitSchema(); err != nil {
		logger.Fatalw("Failed to initialize database schema", "error", err.Error())
	}
	logger.Info("Database schema initialized successfully")

	// 初始化服务层
	clusterService := service.NewClusterService(pgStorage, logrusLogger)
	clusterHandler := handler.NewClusterHandler(clusterService, logrusLogger)

	readTimeout, _ := time.ParseDuration(cfg.Server.ReadTimeout)
	writeTimeout, _ := time.ParseDuration(cfg.Server.WriteTimeout)

	serverConfig := &api.ServerConfig{
		Port:         cfg.Server.Port,
		Mode:         cfg.Server.Mode,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		JWTSecret:    cfg.JWT.Secret,
	}

	var server *api.Server

	if enableK8sAPI {
		// 初始化 K8s API 相关服务
		k8sClusterService := service.NewK8sClusterService(pgStorage)
		k8sNamespaceService := service.NewK8sNamespaceService(pgStorage, k8sClusterService)
		k8sPodService := service.NewK8sPodService(pgStorage, k8sClusterService)
		k8sDeploymentService := service.NewK8sDeploymentService(pgStorage, k8sClusterService)
		k8sNodeService := service.NewK8sNodeService(pgStorage, k8sClusterService)
		k8sServiceService := service.NewK8sServiceService(pgStorage, k8sClusterService)
		k8sStatefulSetService := service.NewK8sStatefulSetService(pgStorage, k8sClusterService)
		k8sDaemonSetService := service.NewK8sDaemonSetService(pgStorage, k8sClusterService)
		k8sConfigMapService := service.NewK8sConfigMapService(pgStorage, k8sClusterService)
		k8sSecretService := service.NewK8sSecretService(pgStorage, k8sClusterService)
		k8sEndpointService := service.NewK8sEndpointService(pgStorage, k8sClusterService)
		k8sPVCService := service.NewK8sPVCService(pgStorage, k8sClusterService)
		k8sPVService := service.NewK8sPVService(pgStorage, k8sClusterService)
		k8sEndpointSliceService := service.NewK8sEndpointSliceService(pgStorage, k8sClusterService)
		k8sHPAService := service.NewK8sHPAService(pgStorage, k8sClusterService)
		k8sEventService := service.NewK8sEventService(pgStorage, k8sClusterService)
		k8sRoleBindingService := service.NewK8sRoleBindingService(pgStorage, k8sClusterService)
		k8sClusterRoleService := service.NewK8sClusterRoleService(pgStorage, k8sClusterService)
		k8sPriorityClassService := service.NewK8sPriorityClassService(pgStorage, k8sClusterService)
		k8sRoleService := service.NewK8sRoleService(pgStorage, k8sClusterService)
		k8sStorageClassService := service.NewK8sStorageClassService(pgStorage, k8sClusterService)

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
		)

		// 创建支持完整 K8s API 的服务器
		server = api.NewServerWithK8sAPI(serverConfig, clusterHandler, k8sAPIHandler, logrusLogger)

		logger.Infow("K8s API enabled",
			"endpoints", "cluster, namespace, pod, deployment, node, service, statefulset, daemonset, configmap, secret, endpoint, pvc, pv, endpointslice, hpa, event, rolebinding, clusterrole, priorityclass",
			"base_path", "/api/k8s",
		)
	} else {
		// 创建原有的服务器（保持向后兼容）
		server = api.NewServer(serverConfig, clusterHandler, logrusLogger)
		logger.Info("K8s API disabled, using legacy endpoints only")
	}

	// 启动服务器
	go func() {
		if err := server.Start(); err != nil {
			logger.Fatalw("Server failed to start", "error", err.Error())
		}
	}()

	logger.Infow("Cluster service started successfully",
		"port", cfg.Server.Port,
		"mode", cfg.Server.Mode,
	)

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	logger.Infow("Received shutdown signal", "signal", sig.String())
	logrusLogger.WithField("signal", sig).Info("Shutting down server...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Errorw("Server forced to shutdown", "error", err.Error())
		logrusLogger.WithError(err).Error("Server forced to shutdown")
	}

	// Sync logger before exit
	logger.Sync()
	logger.Info("Server exited gracefully")
	logrusLogger.Info("Server exited")
}

// setupLogrusLogger 设置 logrus logger（保持向后兼容）
func setupLogrusLogger(level, format string) *logrus.Logger {
	logger := logrus.New()
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	logger.SetLevel(logLevel)

	if format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	}

	return logger
}

// initCommonLogger 初始化 common/logger
func initCommonLogger(cfg *config.Config) error {
	// 获取服务名称，如果是默认值则使用 cluster-service
	serviceName := version.Get().ServiceName
	if serviceName == "" || serviceName == "unknown" || serviceName == "apiserver" {
		serviceName = "cluster-service"
	}

	// 获取版本信息
	gitVersion := version.Get().GitVersion
	if gitVersion == "" || gitVersion == "unknown" {
		gitVersion = "v0.0.0-dev"
	}

	logConfig := &logger.Config{
		Engine:       "zap",                    // 使用 Zap 引擎（高性能）
		Level:        cfg.Logging.Level,       // 从配置读取日志级别
		Format:       cfg.Logging.Format,      // 从配置读取输出格式
		OutputPaths:  []string{"stdout"},      // 输出到标准输出
		EnableCaller: true,                    // 启用调用者信息
		Development:  cfg.Server.Mode == "debug", // 开发模式
		InitialFields: map[string]interface{}{
			"service.name":    serviceName,  // 覆盖默认的 "unknown"
			"service.version": gitVersion,   // 覆盖默认的 "unknown"
			"service":         serviceName,  // 保留兼容字段
			"version":         gitVersion,   // 保留兼容字段
		},
	}

	if err := logger.Init(logConfig); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	return nil
}
