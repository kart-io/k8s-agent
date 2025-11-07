// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package options

import (
	"fmt"

	"github.com/spf13/pflag"
)

// PrometheusOptions Prometheus 指标采集配置
// 用于暴露 Prometheus 格式的 metrics 端点
type PrometheusOptions struct {
	Enabled bool   `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Port    int    `mapstructure:"port" yaml:"port" json:"port"`
	Path    string `mapstructure:"path" yaml:"path" json:"path"`
}

// NewPrometheusOptions 创建默认的 Prometheus 配置
func NewPrometheusOptions() *PrometheusOptions {
	return &PrometheusOptions{
		Enabled: true,
		Port:    9090,
		Path:    "/metrics",
	}
}

// Validate 验证配置
func (o *PrometheusOptions) Validate() error {
	if !o.Enabled {
		return nil // Prometheus disabled, no validation needed
	}

	if o.Port < 1 || o.Port > 65535 {
		return fmt.Errorf("invalid prometheus port: %d", o.Port)
	}

	if o.Path == "" {
		return fmt.Errorf("prometheus path cannot be empty")
	}

	return nil
}

// Complete 填充默认值
func (o *PrometheusOptions) Complete() error {
	if o.Port == 0 {
		o.Port = 9090
	}
	if o.Path == "" {
		o.Path = "/metrics"
	}
	return nil
}

// AddFlags 添加 Prometheus 相关的命令行参数
func (o *PrometheusOptions) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.Enabled, "prometheus.enabled", o.Enabled,
		"Enable Prometheus metrics endpoint")

	fs.IntVar(&o.Port, "prometheus.port", o.Port,
		"Prometheus metrics port")

	fs.StringVar(&o.Path, "prometheus.path", o.Path,
		"Prometheus metrics endpoint path")
}
