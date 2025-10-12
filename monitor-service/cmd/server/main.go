package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kart-io/k8s-agent/monitor-service/internal/api"
	"github.com/kart-io/k8s-agent/monitor-service/internal/handler"
	"github.com/kart-io/k8s-agent/monitor-service/internal/service"
	"github.com/kart-io/k8s-agent/monitor-service/internal/storage"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port         int    `yaml:"port"`
		Mode         string `yaml:"mode"`
		ReadTimeout  string `yaml:"read_timeout"`
		WriteTimeout string `yaml:"write_timeout"`
	} `yaml:"server"`
	Database struct {
		Host         string `yaml:"host"`
		Port         int    `yaml:"port"`
		User         string `yaml:"user"`
		Password     string `yaml:"password"`
		DBName       string `yaml:"dbname"`
		SSLMode      string `yaml:"sslmode"`
		MaxOpenConns int    `yaml:"max_open_conns"`
		MaxIdleConns int    `yaml:"max_idle_conns"`
	} `yaml:"database"`
	Redis struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
		PoolSize int    `yaml:"pool_size"`
	} `yaml:"redis"`
	Prometheus struct {
		Enabled bool `yaml:"enabled"`
		Port    int  `yaml:"port"`
	} `yaml:"prometheus"`
	JWT struct {
		Secret     string `yaml:"secret"`
		Expiration string `yaml:"expiration"`
	} `yaml:"jwt"`
	Logging struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
		Output string `yaml:"output"`
	} `yaml:"logging"`
}

func main() {
	// 加载配置
	config, err := loadConfig("configs/config.yaml")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	logger := setupLogger(config.Logging.Level, config.Logging.Format)

	// 初始化存储
	pgStorage, err := storage.NewPostgresStorage(&storage.Config{
		Host:         config.Database.Host,
		Port:         config.Database.Port,
		User:         config.Database.User,
		Password:     config.Database.Password,
		DBName:       config.Database.DBName,
		SSLMode:      config.Database.SSLMode,
		MaxOpenConns: config.Database.MaxOpenConns,
		MaxIdleConns: config.Database.MaxIdleConns,
	}, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize PostgreSQL storage")
	}
	defer pgStorage.Close()

	redisStorage, err := storage.NewRedisStorage(&storage.RedisConfig{
		Host:     config.Redis.Host,
		Port:     config.Redis.Port,
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
		PoolSize: config.Redis.PoolSize,
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
	readTimeout, _ := time.ParseDuration(config.Server.ReadTimeout)
	writeTimeout, _ := time.ParseDuration(config.Server.WriteTimeout)

	// 创建服务器
	metricsPort := 0
	if config.Prometheus.Enabled {
		metricsPort = config.Prometheus.Port
	}

	server := api.NewServer(&api.ServerConfig{
		Port:         config.Server.Port,
		Mode:         config.Server.Mode,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		JWTSecret:    config.JWT.Secret,
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

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
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
