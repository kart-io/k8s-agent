//go:build wireinject
// +build wireinject

// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"github.com/google/wire"
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/internal/auth/initializers"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
)

// BaseProviderSet Wire dependency set for base infrastructure.
var BaseProviderSet = wire.NewSet(
	ProvideLogger,
	initializers.NewDatabaseInitializer,
	initializers.NewRedisInitializer,
)

// ServiceProviderSet Wire dependency set for business services.
var ServiceProviderSet = wire.NewSet(
	BaseProviderSet,
	initializers.NewSessionServiceInitializer,
	initializers.NewEmailClientInitializer,
	initializers.NewAuditServiceInitializer,
	initializers.NewNotificationServiceInitializer,
	initializers.NewForcedLogoutServiceInitializer,
)

// ServerProviderSet Wire dependency set for servers (gRPC and HTTP).
var ServerProviderSet = wire.NewSet(
	ServiceProviderSet,
	initializers.NewGRPCServerInitializer, // gRPC 服务器（先初始化，创建共享 Service）
	initializers.NewHTTPServerInitializer, // HTTP 服务器（依赖 gRPC 初始化器）
)

// HealthProviderSet Wire dependency set for health check.
var HealthProviderSet = wire.NewSet(
	pkginitializers.NewHealthCheckInitializer,
	wire.FieldsOf(new(*commonapp.StandardOptions), "Health"),
)

// InitializeAuthComponents automatically injects all components using Wire.
func InitializeAuthComponents(opts *commonapp.StandardOptions) (*AuthComponents, error) {
	wire.Build(
		ServerProviderSet,
		HealthProviderSet,
		NewAuthComponents,
	)
	return nil, nil
}

