// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package options

import (
	"time"

	"github.com/spf13/pflag"
)

const (
	// Default timeout and retry settings
	defaultAITimeout  = 30 * time.Second
	defaultMaxRetries = 3
)

// AIOptions contains AI service configuration options.
type AIOptions struct {
	// ReasoningServiceURL is the URL of the reasoning service.
	ReasoningServiceURL string `json:"reasoning_service_url" mapstructure:"reasoning_service_url"`
	// AgentManagerURL is the URL of the agent manager service.
	AgentManagerURL string `json:"agent_manager_url" mapstructure:"agent_manager_url"`
	// Timeout is the request timeout for AI service calls.
	Timeout time.Duration `json:"timeout" mapstructure:"timeout"`
	// MaxRetries is the maximum number of retry attempts.
	MaxRetries int `json:"max_retries" mapstructure:"max_retries"`
}

// NewAIOptions creates default AIOptions.
func NewAIOptions() *AIOptions {
	return &AIOptions{
		ReasoningServiceURL: "http://localhost:8083", // Reasoning service 端口
		AgentManagerURL:     "http://localhost:8081", // Agent Manager 端口
		Timeout:             defaultAITimeout,
		MaxRetries:          defaultMaxRetries,
	}
}

// AddFlags adds flags for AI options.
func (o *AIOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.ReasoningServiceURL, "ai.reasoning-service-url", o.ReasoningServiceURL,
		"URL of the reasoning service")
	fs.StringVar(&o.AgentManagerURL, "ai.agent-manager-url", o.AgentManagerURL,
		"URL of the agent manager service")
	fs.DurationVar(&o.Timeout, "ai.timeout", o.Timeout,
		"Request timeout for AI service calls")
	fs.IntVar(&o.MaxRetries, "ai.max-retries", o.MaxRetries,
		"Maximum number of retry attempts for AI service calls")
}

// Validate validates AI options.
func (o *AIOptions) Validate() error {
	return nil // Add validation logic if needed
}

// Complete fills in any fields not set that are required to have valid data.
func (o *AIOptions) Complete() error {
	if o.ReasoningServiceURL == "" {
		o.ReasoningServiceURL = "http://localhost:8083"
	}
	if o.AgentManagerURL == "" {
		o.AgentManagerURL = "http://localhost:8081"
	}
	if o.Timeout == 0 {
		o.Timeout = defaultAITimeout
	}
	if o.MaxRetries == 0 {
		o.MaxRetries = defaultMaxRetries
	}
	return nil
}
