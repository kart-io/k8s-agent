// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	commonapp "github.com/kart-io/k8s-agent/pkg/app"
	"github.com/kart-io/k8s-agent/internal/auth/initializers"
	pkginitializers "github.com/kart-io/k8s-agent/pkg/initializers"
	"github.com/kart-io/logger/core"
)

// AuthComponents contains all component initializers.
type AuthComponents struct {
	DB           *initializers.DatabaseInitializer
	Redis        *initializers.RedisInitializer
	Session      *initializers.SessionServiceInitializer
	Email        *initializers.EmailClientInitializer
	Audit        *initializers.AuditServiceInitializer
	Notification *initializers.NotificationServiceInitializer
	ForcedLogout *initializers.ForcedLogoutServiceInitializer
	GRPC         *initializers.GRPCServerInitializer // gRPC 服务器
	HTTP         *initializers.HTTPServerInitializer
	Health       *pkginitializers.HealthCheckInitializer
}

// NewAuthComponents creates a new AuthComponents.
func NewAuthComponents(
	db *initializers.DatabaseInitializer,
	redis *initializers.RedisInitializer,
	session *initializers.SessionServiceInitializer,
	email *initializers.EmailClientInitializer,
	audit *initializers.AuditServiceInitializer,
	notification *initializers.NotificationServiceInitializer,
	forcedLogout *initializers.ForcedLogoutServiceInitializer,
	grpc *initializers.GRPCServerInitializer, // gRPC 初始化器
	http *initializers.HTTPServerInitializer,
	health *pkginitializers.HealthCheckInitializer,
) *AuthComponents {
	return &AuthComponents{
		DB:           db,
		Redis:        redis,
		Session:      session,
		Email:        email,
		Audit:        audit,
		Notification: notification,
		ForcedLogout: forcedLogout,
		GRPC:         grpc,
		HTTP:         http,
		Health:       health,
	}
}

// ProvideLogger provides logger from options.
func ProvideLogger(opts *commonapp.StandardOptions) (core.Logger, error) {
	return opts.InitLogger()
}

