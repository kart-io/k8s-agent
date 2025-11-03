// Copyright 2025 Kart Project. All rights reserved.
// Configuration converters from options to middleware types

package server

import (
	"github.com/kart-io/k8s-agent/common/middleware"
	"github.com/kart-io/k8s-agent/common/options"
)

// ToCORSConfig converts options.CORSOptions to middleware.CORSConfig
func ToCORSConfig(opts *options.CORSOptions) middleware.CORSConfig {
	if opts == nil {
		return middleware.DefaultCORSConfig()
	}

	// Check if allow all origins
	allowAllOrigins := false
	for _, origin := range opts.AllowOrigins {
		if origin == "*" {
			allowAllOrigins = true
			break
		}
	}

	return middleware.CORSConfig{
		AllowAllOrigins:  allowAllOrigins,
		AllowOrigins:     opts.AllowOrigins,
		AllowMethods:     opts.AllowMethods,
		AllowHeaders:     opts.AllowHeaders,
		ExposeHeaders:    opts.ExposeHeaders,
		AllowCredentials: opts.AllowCredentials,
		MaxAge:           opts.MaxAge,
	}
}

// ToJWTConfig converts options.JWTOptions to middleware.JWTConfig
func ToJWTConfig(opts *options.JWTOptions) *middleware.JWTConfig {
	if opts == nil {
		return nil
	}

	return &middleware.JWTConfig{
		Secret: []byte(opts.Secret),
		// SessionValidator can be set by the caller if needed
	}
}
