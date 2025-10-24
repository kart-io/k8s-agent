package app

import (
	"context"
	"fmt"

	"github.com/kart-io/k8s-agent/common/logger"
	"github.com/kart-io/k8s-agent/internal/reasoning/config"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
)

// Execute runs the reasoning command
func Execute() {
	// 创建配置选项
	opts := config.NewOptions()

	// 定义运行函数
	runFunc := func(opts commonapp.Options) error {
		return run(opts.(*config.Options))
	}

	// 使用增强框架运行应用
	commonapp.RunWithOptions(opts, runFunc, commonapp.CommandConfig{
		Use:       "reasoning",
		Short:     "Reasoning Service",
		Long:      "Reasoning Service provides AI-driven root cause analysis and intelligent recommendations",
		EnvPrefix: "REASONING",
	},
		// 启用健康检查
		commonapp.WithHealthCheck(commonapp.DefaultHealthCheckFuncFromOptions(opts)),
		// 启用版本信息
		commonapp.WithPrintVersion(),
		// 启用运行时信息
		commonapp.WithPrintRuntime(),
		// 启用配置监听
		commonapp.WithWatch(),
	)
}

// run runs the reasoning service
func run(opts *config.Options) error {
	// Initialize logger
	log, err := logger.InitFromOptions(opts.Logging)
	if err != nil {
		return fmt.Errorf("failed to init logger: %w", err)
	}
	defer log.Flush()

	log.Info("Aetherius Reasoning Service (Go)")
	log.Info("=================================")
	log.Infow("Starting Reasoning Service",
		"host", opts.Server.Host,
		"port", opts.Server.Port,
		"llm_enabled", opts.LLM.Enabled,
		"memory_enabled", opts.Memory.EnableVectorStore,
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
