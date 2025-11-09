// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package options

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
)

// AlertOptions 告警配置
// 支持多种告警通道：Email, Webhook, Slack
type AlertOptions struct {
	Enabled       bool            `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	CheckInterval time.Duration   `mapstructure:"check_interval" yaml:"check_interval" json:"check_interval"`
	Email         *EmailOptions   `mapstructure:"email" yaml:"email" json:"email,omitempty"`
	Webhook       *WebhookOptions `mapstructure:"webhook" yaml:"webhook" json:"webhook,omitempty"`
	Slack         *SlackOptions   `mapstructure:"slack" yaml:"slack" json:"slack,omitempty"`
}

// WebhookOptions Webhook 告警配置
type WebhookOptions struct {
	Enabled bool              `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	URL     string            `mapstructure:"url" yaml:"url" json:"url"`
	Timeout time.Duration     `mapstructure:"timeout" yaml:"timeout" json:"timeout"`
	Headers map[string]string `mapstructure:"headers" yaml:"headers" json:"headers,omitempty"`
}

// SlackOptions Slack 告警配置
type SlackOptions struct {
	Enabled    bool   `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	WebhookURL string `mapstructure:"webhook_url" yaml:"webhook_url" json:"webhook_url"`
	Channel    string `mapstructure:"channel" yaml:"channel" json:"channel,omitempty"`
	Username   string `mapstructure:"username" yaml:"username" json:"username,omitempty"`
	IconEmoji  string `mapstructure:"icon_emoji" yaml:"icon_emoji" json:"icon_emoji,omitempty"`
}

// NewAlertOptions 创建默认的告警配置
func NewAlertOptions() *AlertOptions {
	return &AlertOptions{
		Enabled:       false,
		CheckInterval: 30 * time.Second,
		Email:         NewEmailOptions(),
		Webhook: &WebhookOptions{
			Enabled: false,
			Timeout: 10 * time.Second,
			Headers: make(map[string]string),
		},
		Slack: &SlackOptions{
			Enabled:   false,
			Username:  "Alert Bot",
			IconEmoji: ":warning:",
		},
	}
}

// Validate 验证配置
func (o *AlertOptions) Validate() error {
	if !o.Enabled {
		return nil // Alert disabled, no validation needed
	}

	if o.CheckInterval <= 0 {
		return fmt.Errorf("check_interval must be positive")
	}

	// Validate Email if enabled
	if o.Email != nil && o.Email.Enabled {
		if err := o.Email.Validate(); err != nil {
			return fmt.Errorf("email alert validation failed: %w", err)
		}
	}

	// Validate Webhook if enabled
	if o.Webhook != nil && o.Webhook.Enabled {
		if o.Webhook.URL == "" {
			return fmt.Errorf("webhook url is required when webhook alert is enabled")
		}
		if o.Webhook.Timeout <= 0 {
			return fmt.Errorf("webhook timeout must be positive")
		}
	}

	// Validate Slack if enabled
	if o.Slack != nil && o.Slack.Enabled {
		if o.Slack.WebhookURL == "" {
			return fmt.Errorf("slack webhook_url is required when slack alert is enabled")
		}
	}

	// At least one channel should be enabled
	hasChannel := false
	if o.Email != nil && o.Email.Enabled {
		hasChannel = true
	}
	if o.Webhook != nil && o.Webhook.Enabled {
		hasChannel = true
	}
	if o.Slack != nil && o.Slack.Enabled {
		hasChannel = true
	}

	if !hasChannel {
		return fmt.Errorf("at least one alert channel (email, webhook, or slack) must be enabled")
	}

	return nil
}

// Complete 填充默认值
func (o *AlertOptions) Complete() error {
	if o.CheckInterval == 0 {
		o.CheckInterval = 30 * time.Second
	}

	if o.Email != nil {
		if err := o.Email.Complete(); err != nil {
			return err
		}
	}

	if o.Webhook != nil {
		if o.Webhook.Timeout == 0 {
			o.Webhook.Timeout = 10 * time.Second
		}
		if o.Webhook.Headers == nil {
			o.Webhook.Headers = make(map[string]string)
		}
	}

	if o.Slack != nil {
		if o.Slack.Username == "" {
			o.Slack.Username = "Alert Bot"
		}
		if o.Slack.IconEmoji == "" {
			o.Slack.IconEmoji = ":warning:"
		}
	}

	return nil
}

// AddFlags 添加告警相关的命令行参数
func (o *AlertOptions) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.Enabled, "alert.enabled", o.Enabled,
		"Enable alert notifications")

	fs.DurationVar(&o.CheckInterval, "alert.check-interval", o.CheckInterval,
		"Alert check interval")

	// Email alert flags
	if o.Email != nil {
		o.Email.AddFlags(fs)
	}

	// Webhook alert flags
	if o.Webhook != nil {
		fs.BoolVar(&o.Webhook.Enabled, "alert.webhook.enabled", o.Webhook.Enabled,
			"Enable webhook alerts")
		fs.StringVar(&o.Webhook.URL, "alert.webhook.url", o.Webhook.URL,
			"Webhook URL for alerts")
		fs.DurationVar(&o.Webhook.Timeout, "alert.webhook.timeout", o.Webhook.Timeout,
			"Webhook request timeout")
	}

	// Slack alert flags
	if o.Slack != nil {
		fs.BoolVar(&o.Slack.Enabled, "alert.slack.enabled", o.Slack.Enabled,
			"Enable Slack alerts")
		fs.StringVar(&o.Slack.WebhookURL, "alert.slack.webhook-url", o.Slack.WebhookURL,
			"Slack webhook URL for alerts")
		fs.StringVar(&o.Slack.Channel, "alert.slack.channel", o.Slack.Channel,
			"Slack channel for alerts")
		fs.StringVar(&o.Slack.Username, "alert.slack.username", o.Slack.Username,
			"Slack bot username")
	}
}
