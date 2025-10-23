package config

// Options 配置接口
type Options interface {
	// Validate 验证配置
	Validate() error
}

// Option 配置选项函数
type Option interface {
	Apply(Options)
}
