package options

import (
	"fmt"

	"github.com/spf13/pflag"
)

// HealthOptions 健康检查配置
type HealthOptions struct {
	Port int `mapstructure:"port" yaml:"port" json:"port"` // 健康检查端口
}

// NewHealthOptions 创建默认的健康检查配置
func NewHealthOptions() *HealthOptions {
	return &HealthOptions{
		Port: 8090, // 默认健康检查端口
	}
}

// Validate 验证配置
func (o *HealthOptions) Validate() error {
	if o.Port <= 0 || o.Port > 65535 {
		return fmt.Errorf("health check port must be between 1-65535, got %d", o.Port)
	}
	return nil
}

// AddFlags 添加命令行参数
func (o *HealthOptions) AddFlags(fs *pflag.FlagSet, prefix string) {
	if prefix != "" {
		prefix = prefix + "-"
	}

	fs.IntVar(&o.Port, prefix+"health-port", o.Port,
		"Health check server port")
}

// ApplyDefaults 应用默认值
func (o *HealthOptions) ApplyDefaults() {
	if o.Port == 0 {
		o.Port = 8090
	}
}

// Complete 完成配置初始化（实现 Options 接口）
func (o *HealthOptions) Complete() error {
	o.ApplyDefaults()
	return nil
}

// GetAddr 获取监听地址
func (o *HealthOptions) GetAddr() string {
	return fmt.Sprintf(":%d", o.Port)
}
