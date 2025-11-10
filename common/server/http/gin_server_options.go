// Copyright 2025 Kart Project. All rights reserved.
// Gin HTTP server options

package server

import (
	"github.com/kart-io/k8s-agent/common/options"
)

// GinServerOptions holds complete options for Gin HTTP server
// including basic server options and middleware configurations
type GinServerOptions struct {
	// Server basic server configuration (host, port, timeouts, etc.)
	Server *options.ServerOptions

	// Middleware configurations (optional)
	CORS      *options.CORSOptions
	JWT       *options.JWTOptions
	RateLimit *options.RateLimitOptions

	// Middleware enable flags (default: true for basic middleware)
	EnableRecovery  bool
	EnableRequestID bool
	EnableLogger    bool
}

// NewGinServerOptions creates a GinServerOptions with default values
func NewGinServerOptions(serverOpts *options.ServerOptions) *GinServerOptions {
	return &GinServerOptions{
		Server:          serverOpts,
		EnableRecovery:  true, // Always enable recovery
		EnableRequestID: true, // Always enable request ID
		EnableLogger:    true, // Always enable logger
	}
}

// WithCORS adds CORS configuration
func (o *GinServerOptions) WithCORS(cors *options.CORSOptions) *GinServerOptions {
	o.CORS = cors
	return o
}

// WithJWT adds JWT configuration
func (o *GinServerOptions) WithJWT(jwt *options.JWTOptions) *GinServerOptions {
	o.JWT = jwt
	return o
}

// WithRateLimit adds rate limit configuration
func (o *GinServerOptions) WithRateLimit(rateLimit *options.RateLimitOptions) *GinServerOptions {
	o.RateLimit = rateLimit
	return o
}

// DisableRecovery disables recovery middleware
func (o *GinServerOptions) DisableRecovery() *GinServerOptions {
	o.EnableRecovery = false
	return o
}

// DisableRequestID disables request ID middleware
func (o *GinServerOptions) DisableRequestID() *GinServerOptions {
	o.EnableRequestID = false
	return o
}

// DisableLogger disables logger middleware
func (o *GinServerOptions) DisableLogger() *GinServerOptions {
	o.EnableLogger = false
	return o
}
