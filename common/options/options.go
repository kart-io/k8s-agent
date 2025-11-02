package options

import "github.com/spf13/pflag"

// Options 配置接口（项目原有接口）
type Options interface {
	// Validate 验证配置
	Validate() error
}

// Option 配置选项函数（项目原有接口）
type Option interface {
	Apply(Options)
}

// === OneX 风格的接口定义 ===

// IOptions 定义标准的 Options 接口
// 参考 OneX 项目设计，所有 Options 应实现此接口
type IOptions interface {
	// Validate 验证配置并返回错误列表
	// OneX 风格返回 []error，允许收集所有验证错误
	Validate() []error

	// AddFlags 添加命令行参数到 FlagSet
	// prefixes 参数用于支持嵌套配置的前缀
	AddFlags(fs *pflag.FlagSet, prefixes ...string)
}

// CompletableOptions 可完成的 Options 接口
// 实现此接口的 Options 支持自动填充默认值和计算派生值
type CompletableOptions interface {
	IOptions

	// Complete 完成配置初始化（填充默认值和派生值）
	Complete() error
}

// ApplicableOptions 可应用的 Options 接口
// 实现此接口的 Options 支持将配置应用到目标对象
type ApplicableOptions interface {
	IOptions

	// ApplyTo 将配置应用到目标接口
	ApplyTo(target interface{}) error
}

// FullOptions 完整的 Options 接口
// 组合了所有 Options 接口，提供完整的 Options 功能
type FullOptions interface {
	// Validate 验证配置
	Validate() []error

	// Complete 完成配置初始化
	Complete() error

	// ApplyTo 将配置应用到目标
	ApplyTo(target interface{}) error

	// AddFlags 添加命令行参数
	AddFlags(fs *pflag.FlagSet, prefixes ...string)
}

// === 适配器函数 ===

// WrapValidateError 将单个 error 包装为 []error
// 用于适配项目原有的 Validate() error 接口到 OneX 的 Validate() []error
func WrapValidateError(err error) []error {
	if err == nil {
		return nil
	}
	return []error{err}
}

// UnwrapValidateErrors 将 []error 合并为单个 error
// 用于适配 OneX 的 Validate() []error 到项目原有的 Validate() error
func UnwrapValidateErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errs[0]
	}

	// 合并多个错误
	msg := "validation failed: "
	for i, err := range errs {
		if i > 0 {
			msg += "; "
		}
		msg += err.Error()
	}
	return &multiError{errors: errs, message: msg}
}

// multiError 多错误类型
type multiError struct {
	errors  []error
	message string
}

func (e *multiError) Error() string {
	return e.message
}

func (e *multiError) Unwrap() []error {
	return e.errors
}
