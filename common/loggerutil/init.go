// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package loggerutil provides utility functions for initializing the kart-io/logger
// from common configuration options.
package loggerutil

import (
	"github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/logger"
	"github.com/kart-io/logger/core"
	"github.com/kart-io/logger/option"
)

// InitFromOptions initializes a logger from LoggingOptions.
// This function converts the common/options.LoggingOptions to logger/option.LogOption
// and creates a new logger instance.
func InitFromOptions(opts *options.LoggingOptions) (core.Logger, error) {
	if opts == nil {
		return logger.NewWithDefaults()
	}

	logOpt := &option.LogOption{
		Engine:      opts.Engine,
		Level:       opts.Level,
		Format:      opts.Format,
		OutputPaths: opts.OutputPaths,
	}

	// Convert OTLP options if provided
	if opts.OTLP != nil {
		logOpt.OTLP = &option.OTLPOption{
			Enabled:  opts.OTLP.Enabled,
			Endpoint: opts.OTLP.Endpoint,
			Insecure: opts.OTLP.Insecure,
			Headers:  opts.OTLP.Headers,
		}
	}

	return logger.New(logOpt)
}
