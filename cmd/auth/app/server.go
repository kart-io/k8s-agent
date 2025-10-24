// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"context"
	"fmt"

	commonapp "github.com/kart-io/k8s-agent/pkg/app"

	"github.com/kart-io/k8s-agent/cmd/auth/app/options"
	"github.com/kart-io/k8s-agent/internal/auth"
)

// Define the description of the command.
const commandDesc = `The auth server provides user authentication and authorization services.

It supports:
- JWT-based authentication
- Session management with Redis
- Role-based access control (RBAC)
- API key management
- Forced logout functionality
- Audit logging
- Email notifications
`

// NewApp creates and returns a new command with default parameters.
func NewApp() {
	opts := options.NewServerOptions()

	// Use the enhanced pkg/app framework with optional features
	commonapp.RunWithOptions(opts, run, commonapp.CommandConfig{
		Use:       auth.Name,
		Short:     "Launch an Aetherius authentication and authorization server",
		Long:      commandDesc,
		EnvPrefix: "AUTH",
	},
		// 启用健康检查（从配置中读取端口）
		commonapp.WithHealthCheck(commonapp.DefaultHealthCheckFuncFromOptions(opts)),
		// 启用版本信息打印
		commonapp.WithPrintVersion(),
		// 启用运行时信息打印
		commonapp.WithPrintRuntime(),
		// 启用配置文件监听
		commonapp.WithWatch(),
	)
}

// run contains the main logic for initializing and running the server.
func run(opts commonapp.Options) error {
	// Type assert to our specific options type
	serverOpts, ok := opts.(*options.ServerOptions)
	if !ok {
		return fmt.Errorf("invalid options type")
	}

	// Initialize logger first
	logger, err := serverOpts.InitLogger()
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	logger.Infow("Starting auth service",
		"service", auth.Name,
		"id", auth.ID,
	)

	// Load the configuration from options
	cfg, err := serverOpts.Config()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create context for the server
	ctx := context.Background()

	// Build the server using the configuration
	server, err := cfg.NewServer(ctx, logger)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	// Run the server with signal context for graceful shutdown
	return server.Run(ctx)
}
