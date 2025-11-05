// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package bootstrap

// DatabaseProvider 定义数据库初始化器应该实现的标准接口.
//
// 这个接口标准化了不同服务的数据库初始化器，统一了获取存储实例的方法名。
// 之前不同服务使用了不同的方法名（GetStorage/Storage/Store），导致代码不一致。
type DatabaseProvider interface {
	Initializer

	// Store 返回数据库存储实例
	// 统一使用 Store() 作为标准方法名，替代之前的 GetStorage/Storage/Store
	Store() interface{}
}

// RedisProvider 定义 Redis 初始化器应该实现的标准接口.
type RedisProvider interface {
	Initializer

	// Store 返回 Redis 存储实例
	Store() interface{}
}

// ComponentProvider 定义所有提供组件实例的初始化器的通用接口.
//
// 设计原则：
// - 统一的方法命名约定
// - 类型安全的接口设计
// - 易于测试和模拟
type ComponentProvider interface {
	Initializer

	// GetComponent 返回组件实例
	// 返回 interface{} 以支持不同类型的组件
	GetComponent() interface{}
}


