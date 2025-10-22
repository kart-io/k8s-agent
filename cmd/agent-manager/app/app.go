package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/internal/agent-manager/config"
	commonapp "github.com/kart-io/k8s-agent/common/app"
	"github.com/kart-io/k8s-agent/common/logger"
)

// Execute runs the agent-manager command
func Execute() {
	// 创建配置选项
	opts := config.NewOptions()

	// 定义运行函数
	runFunc := func(opts commonapp.Options) error {
		return run(opts.(*config.Options))
	}

	// 使用通用框架运行应用
	commonapp.Run(opts, runFunc, commonapp.CommandConfig{
		Use:       "agent-manager",
		Short:     "Agent Manager Service",
		Long:      "Agent Manager Service manages k8s agents across multiple clusters",
		EnvPrefix: "AGENT_MANAGER",
	})
}

// run runs the agent-manager service
func run(opts *config.Options) error {
	// Initialize logger
	log, err := logger.InitFromOptions(opts.Logging)
	if err != nil {
		return fmt.Errorf("failed to init logger: %w", err)
	}
	defer log.Flush()

	log.Infow("Starting Agent Manager Service",
		"host", opts.Server.Host,
		"port", opts.Server.Port,
		"grpc_enabled", opts.GRPC.Enable,
		"grpc_port", opts.GRPC.Port,
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
