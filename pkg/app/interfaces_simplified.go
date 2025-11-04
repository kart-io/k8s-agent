package app

// 核心接口定义

// HealthCheckProvider 提供健康检查功能的接口
type HealthCheckProvider interface {
	// HealthCheck 执行健康检查
	HealthCheck() error
}

// ConfigWatcher 提供配置文件监听的接口
type ConfigWatcher interface {
	// WatchConfig 返回是否需要监听配置文件变化
	WatchConfig() bool
}

// StartupInfoPrinter 提供启动信息打印的接口
type StartupInfoPrinter interface {
	// PrintStartupInfo 打印自定义的启动信息
	PrintStartupInfo()
}

// SilenceMode 提供静默模式控制的接口
type SilenceMode interface {
	// IsSilence 返回是否启用静默模式
	IsSilence() bool
}

// 删除了以下兼容接口：
// - HealthPortProvider (已废弃，使用 HealthOptionsProvider)
// - HealthOptionsProvider (直接使用配置结构体)
