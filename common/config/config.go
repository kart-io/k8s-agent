// Package config provides base configuration interfaces and types.
package config

import "github.com/spf13/pflag"

// Config defines the base configuration interface.
type Config interface {
	// Validate validates the configuration.
	Validate() error

	// Complete completes the configuration with defaults.
	Complete() error
}

// FlaggableConfig defines configuration that can be bound to command-line flags.
type FlaggableConfig interface {
	Config

	// AddFlags adds configuration flags to the given FlagSet.
	AddFlags(fs *pflag.FlagSet, prefixes ...string)
}

// Option defines a functional option for configuration.
type Option[T any] func(*T)

// Apply applies the given options to the target configuration.
func Apply[T any](target *T, opts ...Option[T]) {
	for _, opt := range opts {
		opt(target)
	}
}