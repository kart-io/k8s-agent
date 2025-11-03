package logger

import (
	config "github.com/kart-io/k8s-agent/common/options"
	"github.com/kart-io/logger"
	"github.com/kart-io/logger/core"
	"github.com/kart-io/logger/option"
)

// InitFromOptions 从 options.LoggingOptions 初始化日志
func InitFromOptions(opts *config.LoggingOptions) (core.Logger, error) {
	if opts == nil {
		opts = config.NewLoggingOptions()
	}

	// 转换为 kart-io/logger 的配置
	opt := &option.LogOption{
		Engine:            opts.Engine,
		Level:             opts.Level,
		Format:            opts.Format,
		OutputPaths:       opts.OutputPaths,
		Development:       opts.Development,
		DisableCaller:     opts.DisableCaller,
		DisableStacktrace: opts.DisableStacktrace,
		OTLPEndpoint:      opts.OTLPEndpoint,
	}

	// 添加初始字段
	if len(opts.InitialFields) > 0 {
		opt.WithInitialFields(opts.InitialFields)
	}

	// 设置 OTLP 配置
	if opts.OTLP != nil {
		opt.OTLP = &option.OTLPOption{
			Enabled:  opts.OTLP.Enabled,
			Endpoint: opts.OTLP.Endpoint,
			Protocol: opts.OTLP.Protocol,
			Timeout:  opts.OTLP.Timeout,
			Headers:  opts.OTLP.Headers,
			Insecure: opts.OTLP.Insecure,
		}
	}

	// 设置 Rotation 配置
	if opts.Rotation != nil {
		opt.Rotation = &option.RotationOption{
			MaxSize:        opts.Rotation.MaxSize,
			MaxAge:         opts.Rotation.MaxAge,
			MaxBackups:     opts.Rotation.MaxBackups,
			Compress:       opts.Rotation.Compress,
			RotateInterval: opts.Rotation.RotateInterval,
		}
	}

	// 创建 logger
	coreLogger, err := logger.New(opt)
	if err != nil {
		return nil, err
	}

	return coreLogger, nil
}

// InitGlobalFromOptions 从 options.LoggingOptions 初始化全局日志
func InitGlobalFromOptions(opts *config.LoggingOptions) error {
	coreLogger, err := InitFromOptions(opts)
	if err != nil {
		return err
	}

	// 设置为全局 logger
	logger.SetGlobal(coreLogger)
	return nil
}

// InitWithDefaults 使用默认配置初始化日志
func InitWithDefaults() error {
	opts := config.NewLoggingOptions()
	_, err := InitFromOptions(opts)
	return err
}

// Get 获取全局 logger（返回 kart-io/logger 的 core.Logger）
func Get() core.Logger {
	return logger.Global()
}

// Debug 输出 Debug 级别日志（简单参数风格）
func Debug(args ...interface{}) {
	logger.Debug(args...)
}

// Info 输出 Info 级别日志（简单参数风格）
func Info(args ...interface{}) {
	logger.Info(args...)
}

// Warn 输出 Warn 级别日志（简单参数风格）
func Warn(args ...interface{}) {
	logger.Warn(args...)
}

// Error 输出 Error 级别日志（简单参数风格）
func Error(args ...interface{}) {
	logger.Error(args...)
}

// Fatal 输出 Fatal 级别日志并退出程序（简单参数风格）
func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

// Debugf 输出 Debug 级别日志（格式化风格）
func Debugf(template string, args ...interface{}) {
	logger.Debugf(template, args...)
}

// Infof 输出 Info 级别日志（格式化风格）
func Infof(template string, args ...interface{}) {
	logger.Infof(template, args...)
}

// Warnf 输出 Warn 级别日志（格式化风格）
func Warnf(template string, args ...interface{}) {
	logger.Warnf(template, args...)
}

// Errorf 输出 Error 级别日志（格式化风格）
func Errorf(template string, args ...interface{}) {
	logger.Errorf(template, args...)
}

// Fatalf 输出 Fatal 级别日志并退出程序（格式化风格）
func Fatalf(template string, args ...interface{}) {
	logger.Fatalf(template, args...)
}

// Debugw 输出 Debug 级别日志（结构化风格）
func Debugw(msg string, keysAndValues ...interface{}) {
	logger.Debugw(msg, keysAndValues...)
}

// Infow 输出 Info 级别日志（结构化风格）
func Infow(msg string, keysAndValues ...interface{}) {
	logger.Infow(msg, keysAndValues...)
}

// Warnw 输出 Warn 级别日志（结构化风格）
func Warnw(msg string, keysAndValues ...interface{}) {
	logger.Warnw(msg, keysAndValues...)
}

// Errorw 输出 Error 级别日志（结构化风格）
func Errorw(msg string, keysAndValues ...interface{}) {
	logger.Errorw(msg, keysAndValues...)
}

// Fatalw 输出 Fatal 级别日志并退出程序（结构化风格）
func Fatalw(msg string, keysAndValues ...interface{}) {
	logger.Fatalw(msg, keysAndValues...)
}

// With 创建带字段的子 logger
func With(keysAndValues ...interface{}) core.Logger {
	return logger.With(keysAndValues...)
}

// Sync 刷新日志缓冲区
func Sync() error {
	return logger.Flush()
}
