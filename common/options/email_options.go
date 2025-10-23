package options

import (
	"fmt"

	"github.com/spf13/pflag"
)

// EmailOptions 邮件通知配置
type EmailOptions struct {
	Enabled      bool   `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	SMTPHost     string `mapstructure:"smtp_host" yaml:"smtp_host" json:"smtp_host"`
	SMTPPort     int    `mapstructure:"smtp_port" yaml:"smtp_port" json:"smtp_port"`
	SMTPUser     string `mapstructure:"smtp_user" yaml:"smtp_user" json:"smtp_user"`
	SMTPPassword string `mapstructure:"smtp_password" yaml:"smtp_password" json:"smtp_password"`
	FromAddress  string `mapstructure:"from_address" yaml:"from_address" json:"from_address"`
	FromName     string `mapstructure:"from_name" yaml:"from_name" json:"from_name"`
	TemplateDir  string `mapstructure:"template_dir" yaml:"template_dir" json:"template_dir"`
}

// NewEmailOptions 创建默认的邮件配置
func NewEmailOptions() *EmailOptions {
	return &EmailOptions{
		Enabled:      false,
		SMTPHost:     "localhost",
		SMTPPort:     25,
		SMTPUser:     "",
		SMTPPassword: "",
		FromAddress:  "noreply@example.com",
		FromName:     "System",
		TemplateDir:  "./templates/email",
	}
}

// Validate 验证配置
func (o *EmailOptions) Validate() error {
	if !o.Enabled {
		return nil // Email disabled, no validation needed
	}

	if o.SMTPHost == "" {
		return fmt.Errorf("smtp_host is required when email is enabled")
	}
	if o.SMTPPort < 1 || o.SMTPPort > 65535 {
		return fmt.Errorf("invalid smtp_port: %d", o.SMTPPort)
	}
	if o.FromAddress == "" {
		return fmt.Errorf("from_address is required when email is enabled")
	}
	if o.TemplateDir == "" {
		return fmt.Errorf("template_dir is required when email is enabled")
	}
	return nil
}

// AddFlags 添加命令行参数
func (o *EmailOptions) AddFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.Enabled, "email.enabled", o.Enabled, "Enable email notifications")
	fs.StringVar(&o.SMTPHost, "email.smtp-host", o.SMTPHost, "SMTP server host")
	fs.IntVar(&o.SMTPPort, "email.smtp-port", o.SMTPPort, "SMTP server port")
	fs.StringVar(&o.SMTPUser, "email.smtp-user", o.SMTPUser, "SMTP authentication username")
	fs.StringVar(&o.SMTPPassword, "email.smtp-password", o.SMTPPassword, "SMTP authentication password")
	fs.StringVar(&o.FromAddress, "email.from-address", o.FromAddress, "Sender email address")
	fs.StringVar(&o.FromName, "email.from-name", o.FromName, "Sender name")
	fs.StringVar(&o.TemplateDir, "email.template-dir", o.TemplateDir, "Email template directory")
}

// ApplyTo 将配置应用到目标接口
func (o *EmailOptions) ApplyTo(target interface{}) error {
	if target == nil {
		return nil
	}

	switch v := target.(type) {
	case *[]interface{}:
		*v = append(*v,
			map[string]interface{}{
				"enabled":      o.Enabled,
				"smtpHost":     o.SMTPHost,
				"smtpPort":     o.SMTPPort,
				"smtpUser":     o.SMTPUser,
				"smtpPassword": o.SMTPPassword,
				"fromAddress":  o.FromAddress,
				"fromName":     o.FromName,
				"templateDir":  o.TemplateDir,
			},
		)
	}

	return nil
}

// Complete 完成配置初始化
func (o *EmailOptions) Complete() error {
	// 设置默认值
	if o.SMTPHost == "" {
		o.SMTPHost = "localhost"
	}

	if o.SMTPPort <= 0 || o.SMTPPort > 65535 {
		o.SMTPPort = 25
	}

	if o.FromAddress == "" {
		o.FromAddress = "noreply@example.com"
	}

	if o.FromName == "" {
		o.FromName = "System"
	}

	if o.TemplateDir == "" {
		o.TemplateDir = "./templates/email"
	}

	return nil
}

// WithEmailEnabled 设置是否启用邮件通知
func WithEmailEnabled(enabled bool) func(*EmailOptions) {
	return func(o *EmailOptions) {
		o.Enabled = enabled
	}
}

// WithSMTPHost 设置SMTP服务器地址
func WithSMTPHost(host string) func(*EmailOptions) {
	return func(o *EmailOptions) {
		o.SMTPHost = host
	}
}

// WithSMTPPort 设置SMTP服务器端口
func WithSMTPPort(port int) func(*EmailOptions) {
	return func(o *EmailOptions) {
		o.SMTPPort = port
	}
}

// WithSMTPUser 设置SMTP用户名
func WithSMTPUser(user string) func(*EmailOptions) {
	return func(o *EmailOptions) {
		o.SMTPUser = user
	}
}

// WithSMTPPassword 设置SMTP密码
func WithSMTPPassword(password string) func(*EmailOptions) {
	return func(o *EmailOptions) {
		o.SMTPPassword = password
	}
}

// WithFromAddress 设置发件人地址
func WithFromAddress(address string) func(*EmailOptions) {
	return func(o *EmailOptions) {
		o.FromAddress = address
	}
}

// WithFromName 设置发件人名称
func WithFromName(name string) func(*EmailOptions) {
	return func(o *EmailOptions) {
		o.FromName = name
	}
}

// WithTemplateDir 设置邮件模板目录
func WithTemplateDir(dir string) func(*EmailOptions) {
	return func(o *EmailOptions) {
		o.TemplateDir = dir
	}
}
