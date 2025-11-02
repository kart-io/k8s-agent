package options

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"

	"github.com/kart-io/k8s-agent/common/options/validation"
)

// RateLimitOptions 速率限制配置选项
// 用于控制 API 请求速率，防止服务过载
type RateLimitOptions struct {
	// Enable 是否启用速率限制
	Enable bool `mapstructure:"enable" yaml:"enable" json:"enable"`

	// Strategy 限流策略类型
	// - "token_bucket": 令牌桶算法（默认）
	// - "leaky_bucket": 漏桶算法
	// - "fixed_window": 固定窗口计数
	// - "sliding_window": 滑动窗口计数
	Strategy string `mapstructure:"strategy" yaml:"strategy" json:"strategy"`

	// Rate 每秒生成的令牌数（令牌桶算法）或请求速率
	// 例如: 100 表示每秒最多 100 个请求
	Rate float64 `mapstructure:"rate" yaml:"rate" json:"rate"`

	// Burst 令牌桶容量（允许的突发请求数）
	// 例如: 200 表示允许突发 200 个请求
	Burst int `mapstructure:"burst" yaml:"burst" json:"burst"`

	// Window 时间窗口大小（用于窗口算法）
	// 例如: 1m 表示 1 分钟窗口
	Window time.Duration `mapstructure:"window" yaml:"window" json:"window"`

	// MaxRequests 时间窗口内最大请求数（用于窗口算法）
	MaxRequests int `mapstructure:"max_requests" yaml:"max_requests" json:"max_requests"`

	// KeyStrategy 限流键策略
	// - "ip": 基于客户端 IP（默认）
	// - "user": 基于用户 ID
	// - "api": 基于 API 路径
	// - "global": 全局限流
	KeyStrategy string `mapstructure:"key_strategy" yaml:"key_strategy" json:"key_strategy"`

	// IPWhitelist IP 白名单（不受限流限制）
	IPWhitelist []string `mapstructure:"ip_whitelist" yaml:"ip_whitelist" json:"ip_whitelist"`

	// UserWhitelist 用户白名单（不受限流限制）
	UserWhitelist []string `mapstructure:"user_whitelist" yaml:"user_whitelist" json:"user_whitelist"`

	// ExemptPaths 豁免路径（不受限流限制）
	// 例如: ["/health", "/metrics"]
	ExemptPaths []string `mapstructure:"exempt_paths" yaml:"exempt_paths" json:"exempt_paths"`

	// CustomRules 自定义限流规则（针对特定路径的不同速率）
	// 例如: {"/api/v1/heavy": {rate: 10, burst: 20}}
	CustomRules map[string]RuleConfig `mapstructure:"custom_rules" yaml:"custom_rules" json:"custom_rules"`

	// RetryAfter 返回 Retry-After 头的时间（秒）
	RetryAfter int `mapstructure:"retry_after" yaml:"retry_after" json:"retry_after"`

	// FailureThreshold 失败次数阈值（超过后进入惩罚模式）
	FailureThreshold int `mapstructure:"failure_threshold" yaml:"failure_threshold" json:"failure_threshold"`

	// PenaltyDuration 惩罚持续时间（失败后的冷却期）
	PenaltyDuration time.Duration `mapstructure:"penalty_duration" yaml:"penalty_duration" json:"penalty_duration"`
}

// RuleConfig 自定义限流规则配置
type RuleConfig struct {
	Rate  float64 `mapstructure:"rate" yaml:"rate" json:"rate"`
	Burst int     `mapstructure:"burst" yaml:"burst" json:"burst"`
}

// NewRateLimitOptions 创建默认的速率限制配置
func NewRateLimitOptions() *RateLimitOptions {
	return &RateLimitOptions{
		Enable:           false, // 默认关闭
		Strategy:         "token_bucket",
		Rate:             100, // 每秒 100 个请求
		Burst:            200, // 允许突发 200 个请求
		Window:           1 * time.Minute,
		MaxRequests:      6000, // 1 分钟 6000 个请求（100 req/s）
		KeyStrategy:      "ip",
		IPWhitelist:      []string{},
		UserWhitelist:    []string{},
		ExemptPaths:      []string{"/health", "/healthz", "/readyz", "/livez", "/metrics"},
		CustomRules:      make(map[string]RuleConfig),
		RetryAfter:       60, // 60 秒后重试
		FailureThreshold: 0,  // 默认不启用惩罚模式
		PenaltyDuration:  5 * time.Minute,
	}
}

// Validate 验证配置
func (o *RateLimitOptions) Validate() error {
	if !o.Enable {
		return nil // 速率限制是可选的，如果未启用则跳过验证
	}

	// 验证策略类型
	validStrategies := []string{"token_bucket", "leaky_bucket", "fixed_window", "sliding_window"}
	if !ContainsString(validStrategies, o.Strategy) {
		return fmt.Errorf("invalid rate limit strategy: %s (must be one of %v)", o.Strategy, validStrategies)
	}

	// 验证速率
	if o.Strategy == "token_bucket" || o.Strategy == "leaky_bucket" {
		if err := validation.ValidatePositiveFloat(o.Rate, "rate"); err != nil {
			return err
		}

		if err := validation.ValidatePositiveInt(o.Burst, "burst"); err != nil {
			return err
		}
	}

	// 验证窗口配置
	if o.Strategy == "fixed_window" || o.Strategy == "sliding_window" {
		if err := validation.ValidatePositiveDuration(o.Window, "window"); err != nil {
			return err
		}

		if err := validation.ValidatePositiveInt(o.MaxRequests, "max_requests"); err != nil {
			return err
		}
	}

	// 验证键策略
	validKeyStrategies := []string{"ip", "user", "api", "global"}
	if !ContainsString(validKeyStrategies, o.KeyStrategy) {
		return fmt.Errorf("invalid key strategy: %s (must be one of %v)", o.KeyStrategy, validKeyStrategies)
	}

	// 验证 Retry-After
	if o.RetryAfter < 0 {
		return fmt.Errorf("retry_after must be non-negative, got %d", o.RetryAfter)
	}

	// 验证惩罚配置
	if o.FailureThreshold > 0 {
		if err := validation.ValidatePositiveDuration(o.PenaltyDuration, "penalty_duration"); err != nil {
			return err
		}
	}

	// 验证自定义规则
	for path, rule := range o.CustomRules {
		if path == "" {
			return fmt.Errorf("custom rule path cannot be empty")
		}
		if rule.Rate <= 0 {
			return fmt.Errorf("custom rule rate for %s must be positive, got %f", path, rule.Rate)
		}
		if rule.Burst <= 0 {
			return fmt.Errorf("custom rule burst for %s must be positive, got %d", path, rule.Burst)
		}
	}

	return nil
}

// AddFlags 添加命令行参数
func (o *RateLimitOptions) AddFlags(fs *pflag.FlagSet, prefix string) {
	if prefix != "" {
		prefix = prefix + "."
	}

	fs.BoolVar(&o.Enable, prefix+"ratelimit.enable", o.Enable,
		"Enable rate limiting")
	fs.StringVar(&o.Strategy, prefix+"ratelimit.strategy", o.Strategy,
		"Rate limit strategy (token_bucket, leaky_bucket, fixed_window, sliding_window)")
	fs.Float64Var(&o.Rate, prefix+"ratelimit.rate", o.Rate,
		"Rate limit: tokens per second (for token_bucket/leaky_bucket)")
	fs.IntVar(&o.Burst, prefix+"ratelimit.burst", o.Burst,
		"Rate limit: burst capacity (for token_bucket)")
	fs.DurationVar(&o.Window, prefix+"ratelimit.window", o.Window,
		"Rate limit: time window (for fixed_window/sliding_window)")
	fs.IntVar(&o.MaxRequests, prefix+"ratelimit.max-requests", o.MaxRequests,
		"Rate limit: max requests per window (for fixed_window/sliding_window)")
	fs.StringVar(&o.KeyStrategy, prefix+"ratelimit.key-strategy", o.KeyStrategy,
		"Rate limit key strategy (ip, user, api, global)")
	fs.StringSliceVar(&o.IPWhitelist, prefix+"ratelimit.ip-whitelist", o.IPWhitelist,
		"Rate limit IP whitelist (not subject to rate limiting)")
	fs.StringSliceVar(&o.UserWhitelist, prefix+"ratelimit.user-whitelist", o.UserWhitelist,
		"Rate limit user whitelist (not subject to rate limiting)")
	fs.StringSliceVar(&o.ExemptPaths, prefix+"ratelimit.exempt-paths", o.ExemptPaths,
		"Rate limit exempt paths (not subject to rate limiting)")
	fs.IntVar(&o.RetryAfter, prefix+"ratelimit.retry-after", o.RetryAfter,
		"Rate limit: Retry-After header value in seconds")
	fs.IntVar(&o.FailureThreshold, prefix+"ratelimit.failure-threshold", o.FailureThreshold,
		"Rate limit: failure threshold for penalty mode (0=disabled)")
	fs.DurationVar(&o.PenaltyDuration, prefix+"ratelimit.penalty-duration", o.PenaltyDuration,
		"Rate limit: penalty duration after exceeding failure threshold")
}

// Complete 完成配置初始化
func (o *RateLimitOptions) Complete() error {
	if !o.Enable {
		return nil
	}

	// 确保默认豁免路径不为空
	if len(o.ExemptPaths) == 0 {
		o.ExemptPaths = []string{"/health", "/healthz", "/readyz", "/livez", "/metrics"}
	}

	// 确保自定义规则 map 已初始化
	if o.CustomRules == nil {
		o.CustomRules = make(map[string]RuleConfig)
	}

	// 确保白名单已初始化
	if o.IPWhitelist == nil {
		o.IPWhitelist = []string{}
	}
	if o.UserWhitelist == nil {
		o.UserWhitelist = []string{}
	}

	// 根据策略设置合理的默认值
	switch o.Strategy {
	case "token_bucket", "leaky_bucket":
		if o.Rate <= 0 {
			o.Rate = 100
		}
		if o.Burst <= 0 {
			o.Burst = int(o.Rate * 2) // Burst 默认为 Rate 的 2 倍
		}

	case "fixed_window", "sliding_window":
		if o.Window <= 0 {
			o.Window = 1 * time.Minute
		}
		if o.MaxRequests <= 0 {
			o.MaxRequests = int(o.Rate * 60) // 1 分钟窗口
		}
	}

	return nil
}

// ApplyTo 将配置应用到目标接口
func (o *RateLimitOptions) ApplyTo(target interface{}) error {
	if target == nil {
		return nil
	}

	switch v := target.(type) {
	case *[]interface{}:
		config := map[string]interface{}{
			"enable":           o.Enable,
			"strategy":         o.Strategy,
			"rate":             o.Rate,
			"burst":            o.Burst,
			"window":           o.Window,
			"maxRequests":      o.MaxRequests,
			"keyStrategy":      o.KeyStrategy,
			"ipWhitelist":      o.IPWhitelist,
			"userWhitelist":    o.UserWhitelist,
			"exemptPaths":      o.ExemptPaths,
			"customRules":      o.CustomRules,
			"retryAfter":       o.RetryAfter,
			"failureThreshold": o.FailureThreshold,
			"penaltyDuration":  o.PenaltyDuration,
		}

		*v = append(*v, config)
	}

	return nil
}

// WithRateLimitEnable 设置是否启用速率限制
func WithRateLimitEnable(enable bool) func(*RateLimitOptions) {
	return func(o *RateLimitOptions) {
		o.Enable = enable
	}
}

// WithRateLimitStrategy 设置限流策略
func WithRateLimitStrategy(strategy string) func(*RateLimitOptions) {
	return func(o *RateLimitOptions) {
		o.Strategy = strategy
	}
}

// WithRateLimitRate 设置速率
func WithRateLimitRate(rate float64) func(*RateLimitOptions) {
	return func(o *RateLimitOptions) {
		o.Rate = rate
	}
}

// WithRateLimitBurst 设置突发容量
func WithRateLimitBurst(burst int) func(*RateLimitOptions) {
	return func(o *RateLimitOptions) {
		o.Burst = burst
	}
}

// WithRateLimitWindow 设置时间窗口
func WithRateLimitWindow(window time.Duration) func(*RateLimitOptions) {
	return func(o *RateLimitOptions) {
		o.Window = window
	}
}

// WithRateLimitMaxRequests 设置窗口最大请求数
func WithRateLimitMaxRequests(maxRequests int) func(*RateLimitOptions) {
	return func(o *RateLimitOptions) {
		o.MaxRequests = maxRequests
	}
}

// WithRateLimitKeyStrategy 设置键策略
func WithRateLimitKeyStrategy(keyStrategy string) func(*RateLimitOptions) {
	return func(o *RateLimitOptions) {
		o.KeyStrategy = keyStrategy
	}
}

// WithRateLimitIPWhitelist 设置 IP 白名单
func WithRateLimitIPWhitelist(ips []string) func(*RateLimitOptions) {
	return func(o *RateLimitOptions) {
		o.IPWhitelist = ips
	}
}

// WithRateLimitExemptPaths 设置豁免路径
func WithRateLimitExemptPaths(paths []string) func(*RateLimitOptions) {
	return func(o *RateLimitOptions) {
		o.ExemptPaths = paths
	}
}

// WithRateLimitCustomRule 添加自定义规则
func WithRateLimitCustomRule(path string, rate float64, burst int) func(*RateLimitOptions) {
	return func(o *RateLimitOptions) {
		if o.CustomRules == nil {
			o.CustomRules = make(map[string]RuleConfig)
		}
		o.CustomRules[path] = RuleConfig{
			Rate:  rate,
			Burst: burst,
		}
	}
}
