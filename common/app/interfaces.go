package app

import "github.com/kart-io/k8s-agent/common/options"

// ==================== 可选接口定义 ====================
// 这些接口是可选的，通过类型断言检测是否实现

// HealthCheckProvider 提供健康检查功能的接口
// 如果 Options 实现了此接口，App 会自动调用其健康检查方法
type HealthCheckProvider interface {
	// HealthCheck 执行健康检查
	HealthCheck() error
}

// ConfigWatcher 提供配置文件监听的接口
// 如果 Options 实现了此接口，App 会根据其返回值决定是否监听配置文件
type ConfigWatcher interface {
	// WatchConfig 返回是否需要监听配置文件变化
	WatchConfig() bool
}

// StartupInfoPrinter 提供启动信息打印的接口
// 如果 Options 实现了此接口，App 会在启动时调用它
type StartupInfoPrinter interface {
	// PrintStartupInfo 打印自定义的启动信息
	PrintStartupInfo()
}

// SilenceMode 提供静默模式控制的接口
// 如果 Options 实现了此接口，App 会根据其返回值决定是否静默
type SilenceMode interface {
	// IsSilence 返回是否启用静默模式
	IsSilence() bool
}

// HealthOptionsProvider 提供健康检查完整配置的接口（推荐使用）
// 如果 Options 实现了此接口，App 会使用其返回的健康检查配置
type HealthOptionsProvider interface {
	// GetHealthOptions 返回健康检查配置
	GetHealthOptions() *options.HealthOptions
}

// HealthPortProvider 提供健康检查端口的接口（向后兼容，不推荐使用）
// 如果 Options 实现了此接口，App 会使用其返回的端口号
// 注意：推荐实现 HealthOptionsProvider 接口以获得更完整的配置支持
type HealthPortProvider interface {
	// GetHealthPort 返回健康检查端口
	GetHealthPort() int
}
