package app

import (
	"context"
	"fmt"

	"github.com/kart-io/logger"
	"github.com/kart-io/k8s-agent/internal/orchestrator/config"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
)

// Execute runs the orchestrator command
func Execute() {
	// 创建配置选项
	opts := config.NewOptions()

	// 定义运行函数
	runFunc := func(opts commonapp.Options) error {
		return run(opts.(*config.Options))
	}

	// 使用通用框架运行应用
	commonapp.Run(opts, runFunc, commonapp.CommandConfig{
		Use:       "orchestrator",
		Short:     "Orchestrator Service",
		Long:      "Orchestrator Service manages workflow orchestration for automated diagnosis and remediation",
		EnvPrefix: "ORCHESTRATOR",
	})
}

// run runs the orchestrator service
func run(opts *config.Options) error {
	// Initialize logger
	log, err := logger.InitFromOptions(opts.Logging)
	if err != nil {
		return fmt.Errorf("failed to init logger: %w", err)
	}
	defer log.Flush()

	log.Info("==========================================================")
	log.Info("     Aetherius Orchestrator Service - Initialization      ")
	log.Info("==========================================================")

	log.Infow("Starting Orchestrator Service",
		"nats_url", opts.NATS.URL,
		"database_host", opts.Database.Host,
		"redis_addr", opts.Redis.Addr,
	)

	// Create server
	srv, err := NewServer(opts, log)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	// Start server
	ctx := context.Background()
	return srv.Run(ctx)
}
