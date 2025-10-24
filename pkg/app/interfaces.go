package app

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

// HealthPortProvider 提供健康检查端口的接口
// 如果 Options 实现了此接口，App 会使用其返回的端口号
type HealthPortProvider interface {
	// GetHealthPort 返回健康检查端口
	GetHealthPort() int
}

// ==================== 辅助函数 ====================

// hasHealthCheck 检查 Options 是否实现了健康检查接口
func hasHealthCheck(opts Options) (HealthCheckProvider, bool) {
	hc, ok := opts.(HealthCheckProvider)
	return hc, ok
}

// hasConfigWatch 检查 Options 是否实现了配置监听接口
func hasConfigWatch(opts Options) (ConfigWatcher, bool) {
	cw, ok := opts.(ConfigWatcher)
	return cw, ok
}

// hasStartupInfoPrinter 检查 Options 是否实现了启动信息打印接口
func hasStartupInfoPrinter(opts Options) (StartupInfoPrinter, bool) {
	si, ok := opts.(StartupInfoPrinter)
	return si, ok
}

// hasSilenceMode 检查 Options 是否实现了静默模式接口
func hasSilenceMode(opts Options) (SilenceMode, bool) {
	sm, ok := opts.(SilenceMode)
	return sm, ok
}
