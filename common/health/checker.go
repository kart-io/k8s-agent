// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package health

import "context"

// Checker 统一的健康检查器接口
// 用于检查各种依赖（数据库、Redis、NATS 等）
type Checker interface {
	// Name 返回检查器名称
	Name() string
	// Check 执行健康检查
	Check(ctx context.Context) error
}

// CheckerFunc 函数式健康检查器
type CheckerFunc func(ctx context.Context) error

func (f CheckerFunc) Name() string {
	return "anonymous"
}

func (f CheckerFunc) Check(ctx context.Context) error {
	return f(ctx)
}

// NamedCheckerFunc 带名称的函数式健康检查器
type NamedCheckerFunc struct {
	name string
	fn   func(ctx context.Context) error
}

// NewNamedChecker 创建命名的健康检查器
func NewNamedChecker(name string, fn func(ctx context.Context) error) *NamedCheckerFunc {
	return &NamedCheckerFunc{name: name, fn: fn}
}

func (c *NamedCheckerFunc) Name() string {
	return c.name
}

func (c *NamedCheckerFunc) Check(ctx context.Context) error {
	return c.fn(ctx)
}
