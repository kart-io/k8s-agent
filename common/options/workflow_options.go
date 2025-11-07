// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package options

import (
	"time"

	"github.com/spf13/pflag"
)

const (
	// Default workflow timeout settings
	defaultGlobalTimeout      = 30 * time.Minute
	defaultStepDefaultTimeout = 5 * time.Minute
	defaultWorkflowMaxRetries = 3
)

// WorkflowOptions contains workflow timeout configuration options.
type WorkflowOptions struct {
	// GlobalTimeout is the maximum time a workflow can run.
	GlobalTimeout time.Duration `json:"global_timeout" mapstructure:"global_timeout"`
	// StepDefaultTimeout is the default timeout for workflow steps.
	StepDefaultTimeout time.Duration `json:"step_default_timeout" mapstructure:"step_default_timeout"`
	// RetryOnTimeout indicates whether to retry workflows that timeout.
	RetryOnTimeout bool `json:"retry_on_timeout" mapstructure:"retry_on_timeout"`
	// MaxRetries is the maximum number of retry attempts for timed-out workflows.
	MaxRetries int `json:"max_retries" mapstructure:"max_retries"`
}

// NewWorkflowOptions creates default WorkflowOptions.
func NewWorkflowOptions() *WorkflowOptions {
	return &WorkflowOptions{
		GlobalTimeout:      defaultGlobalTimeout,
		StepDefaultTimeout: defaultStepDefaultTimeout,
		RetryOnTimeout:     true,
		MaxRetries:         defaultWorkflowMaxRetries,
	}
}

// AddFlags adds flags for workflow options.
func (o *WorkflowOptions) AddFlags(fs *pflag.FlagSet) {
	fs.DurationVar(&o.GlobalTimeout, "workflow.global-timeout", o.GlobalTimeout,
		"Maximum time a workflow can run")
	fs.DurationVar(&o.StepDefaultTimeout, "workflow.step-default-timeout", o.StepDefaultTimeout,
		"Default timeout for workflow steps")
	fs.BoolVar(&o.RetryOnTimeout, "workflow.retry-on-timeout", o.RetryOnTimeout,
		"Whether to retry workflows that timeout")
	fs.IntVar(&o.MaxRetries, "workflow.max-retries", o.MaxRetries,
		"Maximum number of retry attempts for timed-out workflows")
}

// Validate validates workflow options.
func (o *WorkflowOptions) Validate() error {
	if o.GlobalTimeout <= 0 {
		o.GlobalTimeout = defaultGlobalTimeout
	}
	if o.StepDefaultTimeout <= 0 {
		o.StepDefaultTimeout = defaultStepDefaultTimeout
	}
	if o.MaxRetries < 0 {
		o.MaxRetries = 0
	}
	return nil
}

// Complete fills in any fields not set that are required to have valid data.
func (o *WorkflowOptions) Complete() error {
	if o.GlobalTimeout == 0 {
		o.GlobalTimeout = defaultGlobalTimeout
	}
	if o.StepDefaultTimeout == 0 {
		o.StepDefaultTimeout = defaultStepDefaultTimeout
	}
	if o.MaxRetries == 0 {
		o.MaxRetries = defaultWorkflowMaxRetries
	}
	return nil
}
