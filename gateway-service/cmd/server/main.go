package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kart-io/k8s-agent/gateway-service/internal/middleware"
	"github.com/kart-io/k8s-agent/gateway-service/internal/router"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	if err := loadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logger, err := initLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting Gateway Service...")

	// 连接 Redis
	rdb := connectRedis(logger)
	if rdb != nil {
		defer rdb.Close()
		// 初始化限流器
		middleware.InitRateLimiter(rdb)
	}

	// 设置路由
	r := router.Setup(logger)

	// 创建 HTTP 服务器
	port := viper.GetInt("server.port")
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      r,
		ReadTimeout:  viper.GetDuration("server.read_timeout"),
		WriteTimeout: viper.GetDuration("server.write_timeout"),
	}

	// 启动服务器
	go func() {
		logger.Info("Gateway server started",
			zap.Int("port", port),
			zap.String("mode", viper.GetString("server.mode")),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down gateway server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Gateway server forced to shutdown", zap.Error(err))
	}

	logger.Info("Gateway server exited")
}

// loadConfig 加载配置
func loadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// 设置默认值
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	return nil
}

// initLogger 初始化日志
func initLogger() (*zap.Logger, error) {
	logLevel := viper.GetString("log.level")

	var config zap.Config
	if viper.GetString("server.mode") == "release" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
	}

	// 设置日志级别
	switch logLevel {
	case "debug":
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	return config.Build()
}

// connectRedis 连接 Redis
func connectRedis(logger *zap.Logger) *redis.Client {
	host := viper.GetString("redis.host")
	port := viper.GetInt("redis.port")
	password := viper.GetString("redis.password")
	db := viper.GetInt("redis.db")

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       db,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Warn("Failed to connect to Redis, rate limiting will use local mode",
			zap.Error(err),
		)
		return nil
	}

	logger.Info("Connected to Redis",
		zap.String("addr", fmt.Sprintf("%s:%d", host, port)),
	)

	return rdb
}
