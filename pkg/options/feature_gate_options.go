package options

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/pflag"
)

// FeatureSpec 功能特性规格定义
type FeatureSpec struct {
	// Default 默认是否启用
	Default bool

	// LockToDefault 是否锁定为默认值(不允许修改)
	LockToDefault bool

	// PreRelease 功能成熟度级别
	// Alpha: 默认关闭,可能有bug,未来可能删除
	// Beta: 默认关闭,功能基本稳定
	// GA: 默认开启,功能稳定
	PreRelease string
}

// FeatureGateOptions 功能开关配置选项
// 用于控制实验性功能、新功能的启用/禁用
type FeatureGateOptions struct {
	// Known 已知的功能特性定义
	Known map[string]FeatureSpec `mapstructure:"known" yaml:"known" json:"known"`

	// Enabled 当前启用的功能特性
	Enabled map[string]bool `mapstructure:"enabled" yaml:"enabled" json:"enabled"`

	// Override 从命令行/配置文件覆盖的功能特性
	Override map[string]bool `mapstructure:"override" yaml:"override" json:"override"`
}

// NewFeatureGateOptions 创建默认的功能开关配置
func NewFeatureGateOptions() *FeatureGateOptions {
	return &FeatureGateOptions{
		Known:    make(map[string]FeatureSpec),
		Enabled:  make(map[string]bool),
		Override: make(map[string]bool),
	}
}

// RegisterFeature 注册功能特性
func (o *FeatureGateOptions) RegisterFeature(name string, spec FeatureSpec) {
	if o.Known == nil {
		o.Known = make(map[string]FeatureSpec)
	}
	o.Known[name] = spec
}

// RegisterFeatures 批量注册功能特性
func (o *FeatureGateOptions) RegisterFeatures(features map[string]FeatureSpec) {
	for name, spec := range features {
		o.RegisterFeature(name, spec)
	}
}

// Validate 验证配置
func (o *FeatureGateOptions) Validate() error {
	// 验证所有覆盖的功能是否已注册
	for name := range o.Override {
		if _, exists := o.Known[name]; !exists {
			return fmt.Errorf("unknown feature gate: %s", name)
		}
	}

	// 验证锁定的功能是否被修改
	for name, enabled := range o.Override {
		if spec, exists := o.Known[name]; exists {
			if spec.LockToDefault && enabled != spec.Default {
				return fmt.Errorf("feature gate %s is locked to %v, cannot override to %v",
					name, spec.Default, enabled)
			}
		}
	}

	return nil
}

// AddFlags 添加命令行参数
func (o *FeatureGateOptions) AddFlags(fs *pflag.FlagSet, prefix string) {
	if prefix != "" {
		prefix = prefix + "."
	}

	fs.Var(NewFeatureGateFlag(o), prefix+"feature-gates",
		"A set of key=value pairs that describe feature gates for "+
			"alpha/experimental features. Options are:\n"+
			o.String())
}

// Complete 完成配置初始化
func (o *FeatureGateOptions) Complete() error {
	if o.Enabled == nil {
		o.Enabled = make(map[string]bool)
	}

	// 1. 首先设置所有已知功能的默认值
	for name, spec := range o.Known {
		o.Enabled[name] = spec.Default
	}

	// 2. 然后应用覆盖配置
	for name, enabled := range o.Override {
		o.Enabled[name] = enabled
	}

	return nil
}

// ApplyTo 将配置应用到目标接口
func (o *FeatureGateOptions) ApplyTo(target interface{}) error {
	if target == nil {
		return nil
	}

	switch v := target.(type) {
	case *[]interface{}:
		*v = append(*v,
			map[string]interface{}{
				"enabled": o.Enabled,
			},
		)
	}

	return nil
}

// Enabled 检查功能是否启用
func (o *FeatureGateOptions) IsEnabled(feature string) bool {
	if o.Enabled == nil {
		return false
	}
	return o.Enabled[feature]
}

// String 返回功能开关的字符串表示
func (o *FeatureGateOptions) String() string {
	if len(o.Known) == 0 {
		return "No features registered"
	}

	// 按名称排序
	names := make([]string, 0, len(o.Known))
	for name := range o.Known {
		names = append(names, name)
	}
	sort.Strings(names)

	var lines []string
	for _, name := range names {
		spec := o.Known[name]
		enabled := "false"
		if spec.Default {
			enabled = "true"
		}

		locked := ""
		if spec.LockToDefault {
			locked = " [locked]"
		}

		line := fmt.Sprintf("  %s=%s (%s)%s", name, enabled, spec.PreRelease, locked)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// KnownFeatures 返回所有已知功能列表
func (o *FeatureGateOptions) KnownFeatures() []string {
	names := make([]string, 0, len(o.Known))
	for name := range o.Known {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// FeatureGateFlag 实现 pflag.Value 接口，用于命令行解析
type FeatureGateFlag struct {
	options *FeatureGateOptions
}

func NewFeatureGateFlag(options *FeatureGateOptions) *FeatureGateFlag {
	return &FeatureGateFlag{options: options}
}

// String 返回字符串表示
func (f *FeatureGateFlag) String() string {
	if len(f.options.Override) == 0 {
		return ""
	}

	var pairs []string
	for k, v := range f.options.Override {
		pairs = append(pairs, fmt.Sprintf("%s=%v", k, v))
	}
	sort.Strings(pairs)
	return strings.Join(pairs, ",")
}

// Set 设置功能开关值
// 格式: "feature1=true,feature2=false"
func (f *FeatureGateFlag) Set(value string) error {
	if f.options.Override == nil {
		f.options.Override = make(map[string]bool)
	}

	if value == "" {
		return nil
	}

	pairs := strings.Split(value, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid feature gate format: %s (expected name=value)", pair)
		}

		name := strings.TrimSpace(parts[0])
		valueStr := strings.TrimSpace(parts[1])

		var enabled bool
		switch strings.ToLower(valueStr) {
		case "true", "1", "yes":
			enabled = true
		case "false", "0", "no":
			enabled = false
		default:
			return fmt.Errorf("invalid value for feature %s: %s (expected true/false)", name, valueStr)
		}

		// 检查是否已注册
		if len(f.options.Known) > 0 {
			if _, exists := f.options.Known[name]; !exists {
				return fmt.Errorf("unknown feature gate: %s", name)
			}

			// 检查是否锁定
			if spec, exists := f.options.Known[name]; exists && spec.LockToDefault {
				if enabled != spec.Default {
					return fmt.Errorf("feature gate %s is locked to %v", name, spec.Default)
				}
			}
		}

		f.options.Override[name] = enabled
	}

	return nil
}

// Type 返回类型名称
func (f *FeatureGateFlag) Type() string {
	return "featureGates"
}

// WithFeatureGate 设置单个功能开关
func WithFeatureGate(feature string, enabled bool) func(*FeatureGateOptions) {
	return func(o *FeatureGateOptions) {
		if o.Override == nil {
			o.Override = make(map[string]bool)
		}
		o.Override[feature] = enabled
	}
}

// WithFeatureGates 批量设置功能开关
func WithFeatureGates(gates map[string]bool) func(*FeatureGateOptions) {
	return func(o *FeatureGateOptions) {
		if o.Override == nil {
			o.Override = make(map[string]bool)
		}
		for k, v := range gates {
			o.Override[k] = v
		}
	}
}
