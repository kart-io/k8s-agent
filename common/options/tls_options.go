package options

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/spf13/pflag"

	"github.com/kart-io/k8s-agent/common/options/validation"
)

// TLSOptions TLS配置选项
// 参考 OneX 项目设计,用于配置 TLS 加密连接
type TLSOptions struct {
	// UseTLS 是否启用TLS加密
	UseTLS bool `mapstructure:"use_tls" yaml:"use_tls" json:"use_tls"`

	// InsecureSkipVerify 是否跳过证书验证(不推荐在生产环境使用)
	InsecureSkipVerify bool `mapstructure:"insecure_skip_verify" yaml:"insecure_skip_verify" json:"insecure_skip_verify"`

	// CaCert CA证书文件路径
	CaCert string `mapstructure:"ca_cert" yaml:"ca_cert" json:"ca_cert"`

	// Cert 客户端/服务端证书文件路径
	Cert string `mapstructure:"cert" yaml:"cert" json:"cert"`

	// Key 客户端/服务端私钥文件路径
	Key string `mapstructure:"key" yaml:"key" json:"key"`

	// MinVersion 最小TLS版本 (VersionTLS12, VersionTLS13)
	MinVersion uint16 `mapstructure:"min_version" yaml:"min_version" json:"min_version"`

	// MaxVersion 最大TLS版本
	MaxVersion uint16 `mapstructure:"max_version" yaml:"max_version" json:"max_version"`
}

// NewTLSOptions 创建默认的TLS配置
func NewTLSOptions() *TLSOptions {
	return &TLSOptions{
		UseTLS:             false,
		InsecureSkipVerify: false,
		CaCert:             "",
		Cert:               "",
		Key:                "",
		MinVersion:         tls.VersionTLS12, // 默认最低TLS 1.2
		MaxVersion:         tls.VersionTLS13, // 默认最高TLS 1.3
	}
}

// Validate 验证配置
func (o *TLSOptions) Validate() error {
	if !o.UseTLS {
		return nil // TLS是可选的,如果未启用则跳过验证
	}

	// 证书和密钥必须同时设置或同时为空
	if (o.Cert != "" && o.Key == "") || (o.Cert == "" && o.Key != "") {
		return fmt.Errorf("TLS cert and key must be both set or both empty")
	}

	// 如果设置了证书文件,验证文件是否存在
	if o.Cert != "" {
		if err := validation.ValidateFileExists(o.Cert, "TLS cert file"); err != nil {
			return err
		}
	}

	if o.Key != "" {
		if err := validation.ValidateFileExists(o.Key, "TLS key file"); err != nil {
			return err
		}
	}

	if o.CaCert != "" {
		if err := validation.ValidateFileExists(o.CaCert, "TLS CA cert file"); err != nil {
			return err
		}
	}

	// 验证TLS版本范围
	if o.MinVersion > o.MaxVersion {
		return fmt.Errorf("TLS min_version (%d) cannot be greater than max_version (%d)",
			o.MinVersion, o.MaxVersion)
	}

	return nil
}

// AddFlags 添加命令行参数
func (o *TLSOptions) AddFlags(fs *pflag.FlagSet, prefix string) {
	if prefix != "" {
		prefix = prefix + "."
	}

	fs.BoolVar(&o.UseTLS, prefix+"tls.use-tls", o.UseTLS,
		"Use TLS transport to connect the server")
	fs.BoolVar(&o.InsecureSkipVerify, prefix+"tls.insecure-skip-verify", o.InsecureSkipVerify,
		"Controls whether a client verifies the server's certificate chain and host name")
	fs.StringVar(&o.CaCert, prefix+"tls.ca-cert", o.CaCert,
		"Path to CA cert for connecting to the server")
	fs.StringVar(&o.Cert, prefix+"tls.cert", o.Cert,
		"Path to cert file for connecting to the server")
	fs.StringVar(&o.Key, prefix+"tls.key", o.Key,
		"Path to key file for connecting to the server")
	fs.Uint16Var(&o.MinVersion, prefix+"tls.min-version", o.MinVersion,
		"Minimum TLS version (0x0303 for TLS 1.2, 0x0304 for TLS 1.3)")
	fs.Uint16Var(&o.MaxVersion, prefix+"tls.max-version", o.MaxVersion,
		"Maximum TLS version (0x0303 for TLS 1.2, 0x0304 for TLS 1.3)")
}

// Complete 完成配置初始化
func (o *TLSOptions) Complete() error {
	// 如果未启用TLS,跳过初始化
	if !o.UseTLS {
		return nil
	}

	// 确保TLS版本在有效范围内
	if o.MinVersion == 0 {
		o.MinVersion = tls.VersionTLS12
	}
	if o.MaxVersion == 0 {
		o.MaxVersion = tls.VersionTLS13
	}

	return nil
}

// ApplyTo 将配置应用到目标接口
func (o *TLSOptions) ApplyTo(target interface{}) error {
	if target == nil {
		return nil
	}

	switch v := target.(type) {
	case *[]interface{}:
		*v = append(*v,
			map[string]interface{}{
				"useTLS":             o.UseTLS,
				"insecureSkipVerify": o.InsecureSkipVerify,
				"caCert":             o.CaCert,
				"cert":               o.Cert,
				"key":                o.Key,
				"minVersion":         o.MinVersion,
				"maxVersion":         o.MaxVersion,
			},
		)
	}

	return nil
}

// TLSConfig 构建 *tls.Config 对象
// 这是一个便捷方法,用于从 Options 生成标准的 tls.Config
func (o *TLSOptions) TLSConfig() (*tls.Config, error) {
	if !o.UseTLS {
		return nil, nil
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: o.InsecureSkipVerify,
		MinVersion:         o.MinVersion,
		MaxVersion:         o.MaxVersion,
	}

	// 加载客户端/服务端证书
	if o.Cert != "" && o.Key != "" {
		cert, err := tls.LoadX509KeyPair(o.Cert, o.Key)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS certificates: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// 加载CA证书池
	if o.CaCert != "" {
		data, err := os.ReadFile(o.CaCert)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA cert: %w", err)
		}

		caPool := x509.NewCertPool()
		for {
			var block *pem.Block
			block, data = pem.Decode(data)
			if block == nil {
				break
			}
			caCert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse CA cert: %w", err)
			}
			caPool.AddCert(caCert)
		}

		tlsConfig.RootCAs = caPool
	}

	return tlsConfig, nil
}

// MustTLSConfig 构建 *tls.Config,如果失败则返回空配置
// 仅用于不需要错误处理的场景
func (o *TLSOptions) MustTLSConfig() *tls.Config {
	tlsConf, err := o.TLSConfig()
	if err != nil {
		return &tls.Config{}
	}
	return tlsConf
}

// Scheme 根据TLS配置返回URL方案
func (o *TLSOptions) Scheme() string {
	if o.UseTLS {
		return "https"
	}
	return "http"
}

// WithUseTLS 设置是否启用TLS
func WithUseTLS(use bool) func(*TLSOptions) {
	return func(o *TLSOptions) {
		o.UseTLS = use
	}
}

// WithInsecureSkipVerify 设置是否跳过证书验证
func WithInsecureSkipVerify(skip bool) func(*TLSOptions) {
	return func(o *TLSOptions) {
		o.InsecureSkipVerify = skip
	}
}

// WithCACert 设置CA证书文件路径
func WithCACert(caCert string) func(*TLSOptions) {
	return func(o *TLSOptions) {
		o.CaCert = caCert
	}
}

// WithTLSCert 设置TLS证书文件路径
func WithTLSCert(cert string) func(*TLSOptions) {
	return func(o *TLSOptions) {
		o.Cert = cert
	}
}

// WithTLSKey 设置TLS私钥文件路径
func WithTLSKey(key string) func(*TLSOptions) {
	return func(o *TLSOptions) {
		o.Key = key
	}
}

// WithTLSMinVersion 设置最小TLS版本
func WithTLSMinVersion(version uint16) func(*TLSOptions) {
	return func(o *TLSOptions) {
		o.MinVersion = version
	}
}

// WithTLSMaxVersion 设置最大TLS版本
func WithTLSMaxVersion(version uint16) func(*TLSOptions) {
	return func(o *TLSOptions) {
		o.MaxVersion = version
	}
}
