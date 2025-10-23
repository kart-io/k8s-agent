package app

import (
	"context"
	"fmt"

	commonlogger "github.com/kart-io/k8s-agent/common/logger"
	"github.com/kart-io/k8s-agent/internal/gateway/config"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
)

// Execute runs the gateway command
func Execute() {
	// 创建配置选项
	opts := config.NewOptions()

	// 定义运行函数
	runFunc := func(opts commonapp.Options) error {
		return run(opts.(*config.Options))
	}

	// 使用通用框架运行应用
	commonapp.Run(opts, runFunc, commonapp.CommandConfig{
		Use:       "gateway",
		Short:     "Gateway Service",
		Long:      "Gateway Service provides API gateway with Traefik integration",
		EnvPrefix: "GATEWAY",
	})
}

// run runs the gateway service
func run(opts *config.Options) error {
	// Initialize logger
	log, err := commonlogger.InitFromOptions(opts.Logging)
	if err != nil {
		return fmt.Errorf("failed to init logger: %w", err)
	}
	defer log.Flush()

	log.Info("Starting Gateway Service...")

	// Create server
	srv, err := NewServer(opts, log)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	// Start server
	ctx := context.Background()
	return srv.Run(ctx)
}
