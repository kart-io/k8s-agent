// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package app

import (
	"context"
	"fmt"

	"github.com/kart-io/logger/core"
)

// BaseApplication 提供应用程序的通用实现，减少各服务中的重复代码。
// 各服务可以嵌入此结构体，只需实现特定的业务逻辑。
type BaseApplication struct {
	name   string
	logger core.Logger
}

// NewBaseApplication 创建一个新的基础应用实例。
func NewBaseApplication(name string) *BaseApplication {
	return &BaseApplication{
		name: name,
	}
}

// Name 返回应用程序名称。
func (b *BaseApplication) Name() string {
	return b.name
}

// Logger 返回logger实例。
func (b *BaseApplication) Logger() core.Logger {
	return b.logger
}

// SetLogger 设置logger实例。
func (b *BaseApplication) SetLogger(logger core.Logger) {
	b.logger = logger
}

// InitializeLogger 从配置选项初始化logger。
// 这是一个辅助方法，可被具体的应用使用。
func (b *BaseApplication) InitializeLogger(opts interface{}) error {
	// 尝试从opts获取logger
	type LoggerInitializer interface {
		InitLogger() (core.Logger, error)
	}

	if li, ok := opts.(LoggerInitializer); ok {
		logger, err := li.InitLogger()
		if err != nil {
			return fmt.Errorf("failed to initialize logger: %w", err)
		}
		b.logger = logger
		return nil
	}

	return fmt.Errorf("opts does not implement InitLogger() method")
}

// Run 提供默认的运行实现 - 等待context取消。
// 大多数服务使用 bootstrap 框架，此方法只需等待关闭信号。
func (b *BaseApplication) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

// Shutdown 提供默认的关闭实现。
// Bootstrap 框架会处理组件的关闭，应用层通常不需要额外操作。
func (b *BaseApplication) Shutdown(ctx context.Context) error {
	return nil
}

// StandardInitialize 提供标准的初始化流程。
// 子类可以调用此方法来执行通用的初始化步骤。
func (b *BaseApplication) StandardInitialize(ctx context.Context, opts Options) error {
	// 初始化 logger
	if err := b.InitializeLogger(opts); err != nil {
		return err
	}

	b.logger.Infow("Application initialized", "name", b.name)
	return nil
}

// InitializeHelper 是一个辅助函数，用于实现标准的 Initialize 方法。
// 用法示例：
//   func (a *MyApp) Initialize(ctx context.Context, opts commonapp.Options) error {
//       a.opts = opts.(*options.ServerOptions)
//       return a.BaseApplication.InitializeHelper(ctx, opts)
//   }
type InitializeHelper struct {
	*BaseApplication
}

// NewInitializeHelper 创建一个初始化辅助器。
func NewInitializeHelper(name string) *InitializeHelper {
	return &InitializeHelper{
		BaseApplication: NewBaseApplication(name),
	}
}

