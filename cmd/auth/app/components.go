// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"github.com/kart-io/k8s-agent/cmd/auth/app/options"
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
		HTTP:         http,
		Health:       health,
	}
}

// ProvideLogger provides logger from options.
func ProvideLogger(opts *options.ServerOptions) (core.Logger, error) {
	return opts.InitLogger()
}

