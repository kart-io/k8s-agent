package options

import (
	"github.com/spf13/pflag"

	"github.com/kart-io/k8s-agent/common/options/validation"
)

// AuthenticationOptions 客户端证书认证配置选项
// 参考 OneX 项目设计,用于配置基于客户端证书的双向认证 (mTLS)
type AuthenticationOptions struct {
	// Enable 是否启用客户端证书认证
	Enable bool `mapstructure:"enable" yaml:"enable" json:"enable"`

	// ClientCA CA证书文件路径，用于验证客户端证书
	// 如果设置，任何由该CA签名的客户端证书都将被认证
	// 客户端身份对应证书的 CommonName (CN)
	ClientCA string `mapstructure:"client_ca_file" yaml:"client_ca_file" json:"client_ca_file"`

	// RequireClientCert 是否强制要求客户端证书
	// true: 拒绝没有有效客户端证书的请求
	// false: 允许没有证书的请求，但有证书时会验证
	RequireClientCert bool `mapstructure:"require_client_cert" yaml:"require_client_cert" json:"require_client_cert"`
}

// NewAuthenticationOptions 创建默认的认证配置
func NewAuthenticationOptions() *AuthenticationOptions {
	return &AuthenticationOptions{
		Enable:            false, // 默认关闭
		ClientCA:          "",
		RequireClientCert: false,
	}
}

// Validate 验证配置
func (o *AuthenticationOptions) Validate() error {
	if !o.Enable {
		return nil // 认证是可选的,如果未启用则跳过验证
	}

	// 如果启用了客户端证书认证，必须提供 CA 证书
	if err := validation.ValidateRequired(o.ClientCA, "client CA file"); err != nil {
		return err
	}

	// 验证 CA 证书文件是否存在
	if err := validation.ValidateFileExists(o.ClientCA, "client CA file"); err != nil {
		return err
	}

	return nil
}

// AddFlags 添加命令行参数
func (o *AuthenticationOptions) AddFlags(fs *pflag.FlagSet, prefix string) {
	if prefix != "" {
		prefix = prefix + "."
	}

	fs.BoolVar(&o.Enable, prefix+"auth.enable", o.Enable,
		"Enable client certificate authentication")
	fs.StringVar(&o.ClientCA, prefix+"auth.client-ca-file", o.ClientCA,
		"If set, any request presenting a client certificate signed by one of "+
			"the authorities in the client-ca-file is authenticated with an identity "+
			"corresponding to the CommonName of the client certificate")
	fs.BoolVar(&o.RequireClientCert, prefix+"auth.require-client-cert", o.RequireClientCert,
		"Require client certificate for all requests (reject requests without valid client cert)")
}

// Complete 完成配置初始化
func (o *AuthenticationOptions) Complete() error {
	// 认证配置无需额外初始化
	return nil
}

// ApplyTo 将配置应用到目标接口
func (o *AuthenticationOptions) ApplyTo(target interface{}) error {
	if target == nil {
		return nil
	}

	switch v := target.(type) {
	case *[]interface{}:
		*v = append(*v,
			map[string]interface{}{
				"enable":            o.Enable,
				"clientCA":          o.ClientCA,
				"requireClientCert": o.RequireClientCert,
			},
		)
	}

	return nil
}

// WithAuthenticationEnable 设置是否启用认证
func WithAuthenticationEnable(enable bool) func(*AuthenticationOptions) {
	return func(o *AuthenticationOptions) {
		o.Enable = enable
	}
}

// WithClientCA 设置客户端 CA 证书文件
func WithClientCA(clientCA string) func(*AuthenticationOptions) {
	return func(o *AuthenticationOptions) {
		o.ClientCA = clientCA
	}
}

// WithRequireClientCert 设置是否强制要求客户端证书
func WithRequireClientCert(require bool) func(*AuthenticationOptions) {
	return func(o *AuthenticationOptions) {
		o.RequireClientCert = require
	}
}
