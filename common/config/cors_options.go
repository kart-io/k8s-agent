package config

import (
	"fmt"

	"github.com/spf13/pflag"
)

// CORSOptions CORS跨域配置
type CORSOptions struct {
	Enabled          bool     `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	AllowOrigins     []string `mapstructure:"allow_origins" yaml:"allow_origins" json:"allow_origins"`
	AllowMethods     []string `mapstructure:"allow_methods" yaml:"allow_methods" json:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers" yaml:"allow_headers" json:"allow_headers"`
	ExposeHeaders    []string `mapstructure:"expose_headers" yaml:"expose_headers" json:"expose_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials" yaml:"allow_credentials" json:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age" yaml:"max_age" json:"max_age"` // seconds
}

// NewCORSOptions 创建默认的CORS配置
func NewCORSOptions() *CORSOptions {
	return &CORSOptions{
		Enabled:          true,
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           3600,
	}
}

// Validate 验证配置
func (o *CORSOptions) Validate() error {
	if o.Enabled {
		if len(o.AllowOrigins) == 0 {
			return fmt.Errorf("at least one allow_origin is required when CORS is enabled")
		}
		if len(o.AllowMethods) == 0 {
			return fmt.Errorf("at least one allow_method is required when CORS is enabled")
		}
		if o.MaxAge < 0 {
			return fmt.Errorf("max_age must be >= 0")
		}
	}
	return nil
}

// AddFlags 添加命令行参数
func (o *CORSOptions) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.Enabled, "cors.enabled", o.Enabled, "Enable CORS")
	fs.StringSliceVar(&o.AllowOrigins, "cors.allow-origins", o.AllowOrigins, "CORS allowed origins")
	fs.StringSliceVar(&o.AllowMethods, "cors.allow-methods", o.AllowMethods, "CORS allowed methods")
	fs.StringSliceVar(&o.AllowHeaders, "cors.allow-headers", o.AllowHeaders, "CORS allowed headers")
	fs.StringSliceVar(&o.ExposeHeaders, "cors.expose-headers", o.ExposeHeaders, "CORS exposed headers")
	fs.BoolVar(&o.AllowCredentials, "cors.allow-credentials", o.AllowCredentials, "CORS allow credentials")
	fs.IntVar(&o.MaxAge, "cors.max-age", o.MaxAge, "CORS max age in seconds")
}

// ApplyTo 将配置应用到目标接口
// 接受一个配置结构指针，将CORS配置应用到目标
func (o *CORSOptions) ApplyTo(target interface{}) error {
	if target == nil {
		return nil
	}

	// 类型断言，检查是否为配置结构指针
	switch v := target.(type) {
	case *[]interface{}:
		// 将配置转换为通用选项
		*v = append(*v,
			map[string]interface{}{
				"enabled":          o.Enabled,
				"allowOrigins":     o.AllowOrigins,
				"allowMethods":     o.AllowMethods,
				"allowHeaders":     o.AllowHeaders,
				"exposeHeaders":    o.ExposeHeaders,
				"allowCredentials": o.AllowCredentials,
				"maxAge":           o.MaxAge,
			},
		)
	}

	return nil
}

// Complete 完成配置初始化
// 设置默认值和计算派生值
func (o *CORSOptions) Complete() error {
	if !o.Enabled {
		// 如果 CORS 未启用，无需完成其他配置
		return nil
	}

	// 确保至少有一个允许的来源
	if len(o.AllowOrigins) == 0 {
		o.AllowOrigins = []string{"*"}
	}

	// 确保至少有一个允许的方法
	if len(o.AllowMethods) == 0 {
		o.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}
	}

	// 确保至少有基本的允许头
	if len(o.AllowHeaders) == 0 {
		o.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	}

	// 设置合理的 MaxAge
	if o.MaxAge < 0 {
		o.MaxAge = 0
	} else if o.MaxAge == 0 {
		o.MaxAge = 3600 // 默认1小时
	}

	return nil
}


// WithCORSEnabled 启用CORS
func WithCORSEnabled(enabled bool) func(*CORSOptions) {
	return func(o *CORSOptions) {
		o.Enabled = enabled
	}
}

// WithCORSAllowOrigins 设置允许的来源
func WithCORSAllowOrigins(origins []string) func(*CORSOptions) {
	return func(o *CORSOptions) {
		o.AllowOrigins = origins
	}
}

// WithCORSAllowMethods 设置允许的方法
func WithCORSAllowMethods(methods []string) func(*CORSOptions) {
	return func(o *CORSOptions) {
		o.AllowMethods = methods
	}
}

// WithCORSAllowHeaders 设置允许的头
func WithCORSAllowHeaders(headers []string) func(*CORSOptions) {
	return func(o *CORSOptions) {
		o.AllowHeaders = headers
	}
}

// WithCORSExposeHeaders 设置暴露的头
func WithCORSExposeHeaders(headers []string) func(*CORSOptions) {
	return func(o *CORSOptions) {
		o.ExposeHeaders = headers
	}
}

// WithCORSAllowCredentials 设置是否允许凭证
func WithCORSAllowCredentials(allow bool) func(*CORSOptions) {
	return func(o *CORSOptions) {
		o.AllowCredentials = allow
	}
}

// WithCORSMaxAge 设置缓存时间
func WithCORSMaxAge(maxAge int) func(*CORSOptions) {
	return func(o *CORSOptions) {
		o.MaxAge = maxAge
	}
}