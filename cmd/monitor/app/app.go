package app

import (
    "github.com/kart-io/k8s-agent/common/loggerutil"
	"context"
	"fmt"

	"github.com/kart-io/logger"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/k8s-agent/internal/monitor/config"
)

// Execute runs the monitor command
func Execute() {
	// 创建配置选项
	opts := config.NewOptions()

	// 定义运行函数
	runFunc := func(opts commonapp.Options) error {
		return run(opts.(*config.Options))
	}

	// 使用增强框架运行应用
	commonapp.RunWithOptions(opts, runFunc, commonapp.CommandConfig{
		Use:       "monitor",
		Short:     "Monitor Service",
		Long:      "Monitor Service provides monitoring and metrics collection for the platform",
		EnvPrefix: "MONITOR",
	},
		commonapp.WithHealthCheck(commonapp.DefaultHealthCheckFuncFromOptions(opts)),
		commonapp.WithPrintVersion(),
		commonapp.WithPrintRuntime(),
		commonapp.WithWatch(),
	)
}

// run runs the monitor service
func run(opts *config.Options) error {
	// 转换 LoggingConfig 为 LoggingOptions
	logOpts := &options.LoggingOptions{
		Engine:      "slog",
		Level:       opts.Logging.Level,
		Format:      opts.Logging.Format,
		OutputPaths: []string{opts.Logging.Output},
	}

	// Initialize logger
	log, err := loggerutil.InitFromOptions(logOpts)
	if err != nil {
		return fmt.Errorf("failed to init logger: %w", err)
	}
	defer log.Flush()

	log.Info("Starting Monitor Service...")

	// Create service (使用 common/server)
	svc, err := NewServer(opts, log)
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	defer svc.Cleanup()

	// Start service (使用 common/server.Serve)
	ctx := context.Background()
	return svc.Run(ctx)
}
