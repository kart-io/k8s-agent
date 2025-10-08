package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kart/k8s-agent/cluster-service/internal/api"
	"github.com/kart/k8s-agent/cluster-service/internal/handler"
	"github.com/kart/k8s-agent/cluster-service/internal/service"
	"github.com/kart/k8s-agent/cluster-service/internal/storage"
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
	JWT struct {
		Secret string `yaml:"secret"`
	} `yaml:"jwt"`
	Logging struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
	} `yaml:"logging"`
}

func main() {
	config, err := loadConfig("configs/config.yaml")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger := setupLogger(config.Logging.Level, config.Logging.Format)

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

	clusterService := service.NewClusterService(pgStorage, logger)
	clusterHandler := handler.NewClusterHandler(clusterService, logger)

	readTimeout, _ := time.ParseDuration(config.Server.ReadTimeout)
	writeTimeout, _ := time.ParseDuration(config.Server.WriteTimeout)

	server := api.NewServer(&api.ServerConfig{
		Port:         config.Server.Port,
		Mode:         config.Server.Mode,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		JWTSecret:    config.JWT.Secret,
	}, clusterHandler, logger)

	go func() {
		if err := server.Start(); err != nil {
			logger.WithError(err).Fatal("Server failed to start")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

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
