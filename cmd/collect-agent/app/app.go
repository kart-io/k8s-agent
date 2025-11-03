package app

import (
    "github.com/kart-io/k8s-agent/common/loggerutil"
	"context"
	"fmt"

	"github.com/kart-io/logger"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/internal/collect-agent/config"
)

// Execute runs the collect-agent command
func Execute() {
	// 创建配置选项
	opts := config.NewOptions()

	// 定义运行函数
	runFunc := func(opts commonapp.Options) error {
		return run(opts.(*config.Options))
	}

	// 使用增强框架运行应用
	commonapp.RunWithOptions(opts, runFunc, commonapp.CommandConfig{
		Use:       "collect-agent",
		Short:     "Collect Agent",
		Long:      "Collect Agent monitors K8s cluster events and collects metrics from edge clusters",
		EnvPrefix: "COLLECT_AGENT",
	},
		commonapp.WithHealthCheck(commonapp.DefaultHealthCheckFuncFromOptions(opts)),
		commonapp.WithPrintVersion(),
		commonapp.WithPrintRuntime(),
		commonapp.WithWatch(),
	)
}

// run runs the collect-agent service
func run(opts *config.Options) error {
	// Initialize logger
	log, err := loggerutil.InitFromOptions(opts.Logging)
	if err != nil {
		return fmt.Errorf("failed to init logger: %w", err)
	}
	defer log.Flush()

	log.Infow("Starting Aetherius Collect Agent",
		"cluster_id", opts.Agent.ClusterID,
		"central_endpoint", opts.Agent.CentralEndpoint,
		"health_port", opts.Agent.HealthPort,
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
