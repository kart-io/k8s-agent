// Package server provides server configuration options.
package server

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
)

// Options defines HTTP/gRPC server configuration.
type Options struct {
	// HTTP server options
	HTTPHost           string        `json:"httpHost" mapstructure:"http_host"`
	HTTPPort           int           `json:"httpPort" mapstructure:"http_port"`
	HTTPMode           string        `json:"httpMode" mapstructure:"http_mode"` // debug, release, test
	HTTPReadTimeout    time.Duration `json:"httpReadTimeout" mapstructure:"http_read_timeout"`
	HTTPWriteTimeout   time.Duration `json:"httpWriteTimeout" mapstructure:"http_write_timeout"`
	HTTPMaxHeaderBytes int           `json:"httpMaxHeaderBytes" mapstructure:"http_max_header_bytes"`

	// gRPC server options
	GRPCHost string `json:"grpcHost" mapstructure:"grpc_host"`
	GRPCPort int    `json:"grpcPort" mapstructure:"grpc_port"`

	// Common options
	EnableProfiling bool `json:"enableProfiling" mapstructure:"enable_profiling"`
	EnableMetrics   bool `json:"enableMetrics" mapstructure:"enable_metrics"`
}

// DefaultOptions returns default server options.
func DefaultOptions() *Options {
	return &Options{
		HTTPHost:           "0.0.0.0",
		HTTPPort:           8080,
		HTTPMode:           "release",
		HTTPReadTimeout:    30 * time.Second,
		HTTPWriteTimeout:   30 * time.Second,
		HTTPMaxHeaderBytes: 1 << 20, // 1MB

		GRPCHost: "0.0.0.0",
		GRPCPort: 9090,

		EnableProfiling: false,
		EnableMetrics:   true,
	}
}

// Validate validates the server options.
func (o *Options) Validate() error {
	if o.HTTPPort <= 0 || o.HTTPPort > 65535 {
		return fmt.Errorf("invalid HTTP port: %d", o.HTTPPort)
	}
	if o.GRPCPort <= 0 || o.GRPCPort > 65535 {
		return fmt.Errorf("invalid gRPC port: %d", o.GRPCPort)
	}
	if o.HTTPMode != "debug" && o.HTTPMode != "release" && o.HTTPMode != "test" {
		return fmt.Errorf("invalid HTTP mode: %s", o.HTTPMode)
	}
	return nil
}

// Complete completes the server options with defaults.
func (o *Options) Complete() error {
	if o.HTTPHost == "" {
		o.HTTPHost = "0.0.0.0"
	}
	if o.HTTPPort == 0 {
		o.HTTPPort = 8080
	}
	if o.HTTPMode == "" {
		o.HTTPMode = "release"
	}
	if o.HTTPReadTimeout == 0 {
		o.HTTPReadTimeout = 30 * time.Second
	}
	if o.HTTPWriteTimeout == 0 {
		o.HTTPWriteTimeout = 30 * time.Second
	}
	if o.HTTPMaxHeaderBytes == 0 {
		o.HTTPMaxHeaderBytes = 1 << 20
	}
	if o.GRPCHost == "" {
		o.GRPCHost = "0.0.0.0"
	}
	if o.GRPCPort == 0 {
		o.GRPCPort = 9090
	}
	return nil
}

// AddFlags adds server configuration flags.
func (o *Options) AddFlags(fs *pflag.FlagSet, prefixes ...string) {
	prefix := ""
	if len(prefixes) > 0 {
		prefix = prefixes[0] + "-"
	}

	fs.StringVar(&o.HTTPHost, prefix+"http-host", o.HTTPHost, "HTTP server host")
	fs.IntVar(&o.HTTPPort, prefix+"http-port", o.HTTPPort, "HTTP server port")
	fs.StringVar(&o.HTTPMode, prefix+"http-mode", o.HTTPMode, "HTTP server mode (debug, release, test)")
	fs.DurationVar(&o.HTTPReadTimeout, prefix+"http-read-timeout", o.HTTPReadTimeout, "HTTP read timeout")
	fs.DurationVar(&o.HTTPWriteTimeout, prefix+"http-write-timeout", o.HTTPWriteTimeout, "HTTP write timeout")

	fs.StringVar(&o.GRPCHost, prefix+"grpc-host", o.GRPCHost, "gRPC server host")
	fs.IntVar(&o.GRPCPort, prefix+"grpc-port", o.GRPCPort, "gRPC server port")

	fs.BoolVar(&o.EnableProfiling, prefix+"enable-profiling", o.EnableProfiling, "Enable profiling")
	fs.BoolVar(&o.EnableMetrics, prefix+"enable-metrics", o.EnableMetrics, "Enable metrics")
}

// Functional options

// Option is a functional option for server configuration.
type Option func(*Options)

// WithHTTPPort sets the HTTP port.
func WithHTTPPort(port int) Option {
	return func(o *Options) {
		o.HTTPPort = port
	}
}

// WithHTTPHost sets the HTTP host.
func WithHTTPHost(host string) Option {
	return func(o *Options) {
		o.HTTPHost = host
	}
}

// WithGRPCPort sets the gRPC port.
func WithGRPCPort(port int) Option {
	return func(o *Options) {
		o.GRPCPort = port
	}
}

// WithGRPCHost sets the gRPC host.
func WithGRPCHost(host string) Option {
	return func(o *Options) {
		o.GRPCHost = host
	}
}

// WithHTTPMode sets the HTTP mode.
func WithHTTPMode(mode string) Option {
	return func(o *Options) {
		o.HTTPMode = mode
	}
}

// WithEnableProfiling enables profiling.
func WithEnableProfiling(enable bool) Option {
	return func(o *Options) {
		o.EnableProfiling = enable
	}
}

// WithEnableMetrics enables metrics.
func WithEnableMetrics(enable bool) Option {
	return func(o *Options) {
		o.EnableMetrics = enable
	}
}