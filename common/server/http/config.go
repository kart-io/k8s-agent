// Copyright 2025 Kart Project. All rights reserved.
// Gin HTTP server configuration

package server

import (
	"github.com/kart-io/k8s-agent/common/options"
)

// GinServerConfig holds complete configuration for Gin HTTP server
// including basic server options and middleware configurations
type GinServerConfig struct {
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

// NewGinServerConfig creates a GinServerConfig with default values
func NewGinServerConfig(serverOpts *options.ServerOptions) *GinServerConfig {
	return &GinServerConfig{
		Server:          serverOpts,
		EnableRecovery:  true, // Always enable recovery
		EnableRequestID: true, // Always enable request ID
		EnableLogger:    true, // Always enable logger
	}
}

// WithCORS adds CORS configuration
func (c *GinServerConfig) WithCORS(cors *options.CORSOptions) *GinServerConfig {
	c.CORS = cors
	return c
}

// WithJWT adds JWT configuration
func (c *GinServerConfig) WithJWT(jwt *options.JWTOptions) *GinServerConfig {
	c.JWT = jwt
	return c
}

// WithRateLimit adds rate limit configuration
func (c *GinServerConfig) WithRateLimit(rateLimit *options.RateLimitOptions) *GinServerConfig {
	c.RateLimit = rateLimit
	return c
}

// DisableRecovery disables recovery middleware
func (c *GinServerConfig) DisableRecovery() *GinServerConfig {
	c.EnableRecovery = false
	return c
}

// DisableRequestID disables request ID middleware
func (c *GinServerConfig) DisableRequestID() *GinServerConfig {
	c.EnableRequestID = false
	return c
}

// DisableLogger disables logger middleware
func (c *GinServerConfig) DisableLogger() *GinServerConfig {
	c.EnableLogger = false
	return c
}
