package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kart-io/k8s-agent/internal/monitor/api"
	"github.com/kart-io/k8s-agent/internal/monitor/config"
	"github.com/kart-io/k8s-agent/internal/monitor/handler"
	"github.com/kart-io/k8s-agent/internal/monitor/service"
	"github.com/kart-io/k8s-agent/internal/monitor/storage"
	"github.com/sirupsen/logrus"
)

func main() {
	// Parse command-line flags
	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to configuration file (defaults to ./configs/config.yaml)")
	flag.StringVar(&configPath, "c", "", "Path to configuration file (shorthand)")
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

	// 初始化日志
	logger := setupLogger(cfg.Logging.Level, cfg.Logging.Format)

	// 初始化存储
	pgStorage, err := storage.NewPostgresStorage(&storage.Config{
		Host:         cfg.Database.Host,
		Port:         cfg.Database.Port,
		User:         cfg.Database.User,
		Password:     cfg.Database.Password,
		DBName:       cfg.Database.DBName,
		SSLMode:      cfg.Database.SSLMode,
		MaxOpenConns: cfg.Database.MaxOpenConns,
		MaxIdleConns: cfg.Database.MaxIdleConns,
	}, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize PostgreSQL storage")
	}
	defer pgStorage.Close()

	redisStorage, err := storage.NewRedisStorage(&storage.RedisConfig{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	}, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize Redis storage")
	}
	defer redisStorage.Close()

	// 初始化服务
	monitorService := service.NewMonitorService(pgStorage, redisStorage, logger)

	// 初始化处理器
	metricsHandler := handler.NewMetricsHandler(monitorService, logger)

	// 解析超时时间
	readTimeout, _ := time.ParseDuration(cfg.Server.ReadTimeout)
	writeTimeout, _ := time.ParseDuration(cfg.Server.WriteTimeout)

	// 创建服务器
	metricsPort := 0
	if cfg.Prometheus.Enabled {
		metricsPort = cfg.Prometheus.Port
	}

	server := api.NewServer(&api.ServerConfig{
		Port:         cfg.Server.Port,
		Mode:         cfg.Server.Mode,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		JWTSecret:    cfg.JWT.Secret,
		MetricsPort:  metricsPort,
	}, metricsHandler, logger)

	// 启动服务器
	go func() {
		if err := server.Start(); err != nil {
			logger.WithError(err).Fatal("Server failed to start")
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// 优雅关闭
	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
	}

	logger.Info("Server exited")
}

func setupLogger(level, format string) *logrus.Logger {
	logger := logrus.New()

	// 设置日志级别
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	logger.SetLevel(logLevel)

	// 设置日志格式
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
