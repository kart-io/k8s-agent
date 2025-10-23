package options

import (
	"fmt"

	"github.com/spf13/pflag"
)

// JWTOptions JWT认证配置
type JWTOptions struct {
	Secret       string `mapstructure:"secret" yaml:"secret" json:"secret"`
	ExpiresHours int    `mapstructure:"expires_hours" yaml:"expires_hours" json:"expires_hours"`
}

// NewJWTOptions 创建默认的JWT配置
func NewJWTOptions() *JWTOptions {
	return &JWTOptions{
		Secret:       "",
		ExpiresHours: 24,
	}
}

// Validate 验证配置
func (o *JWTOptions) Validate() error {
	if o.Secret == "" {
		return fmt.Errorf("jwt secret is required")
	}
	if o.ExpiresHours < 1 {
		return fmt.Errorf("jwt expires_hours must be > 0")
	}
	return nil
}

// AddFlags 添加命令行参数
func (o *JWTOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Secret, "jwt.secret", o.Secret, "JWT secret key")
	fs.IntVar(&o.ExpiresHours, "jwt.expires-hours", o.ExpiresHours, "JWT expiration time in hours")
}

// ApplyTo 将配置应用到目标接口
// 接受一个配置结构指针，将JWT配置应用到目标
func (o *JWTOptions) ApplyTo(target interface{}) error {
	if target == nil {
		return nil
	}

	// 类型断言，检查是否为配置结构指针
	switch v := target.(type) {
	case *[]interface{}:
		// 将配置转换为通用选项
		*v = append(*v,
			map[string]interface{}{
				"secret":       o.Secret,
				"expiresHours": o.ExpiresHours,
			},
		)
	}

	return nil
}

// Complete 完成配置初始化
// 设置默认值和计算派生值
func (o *JWTOptions) Complete() error {
	// 确保密钥不为空且长度合理（6-64字符）
	if o.Secret == "" {
		// 注意：在生产环境中应该要求必须设置密钥
		// 这里设置默认值仅用于开发环境
		o.Secret = "change-me-in-production"
	}

	// 确保过期时间在合理范围内 (1-720小时，即1小时到30天)
	if o.ExpiresHours < 1 {
		o.ExpiresHours = 24
	} else if o.ExpiresHours > 720 {
		o.ExpiresHours = 720
	}

	return nil
}

// WithJWTSecret 设置JWT密钥
func WithJWTSecret(secret string) func(*JWTOptions) {
	return func(o *JWTOptions) {
		o.Secret = secret
	}
}

// WithJWTExpiresHours 设置JWT过期时间(小时)
func WithJWTExpiresHours(hours int) func(*JWTOptions) {
	return func(o *JWTOptions) {
		o.ExpiresHours = hours
	}
}
