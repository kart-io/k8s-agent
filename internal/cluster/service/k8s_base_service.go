// Copyright 2024 Kart.IO. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package service

import (
	"github.com/kart-io/k8s-agent/internal/cluster/storage"
)

// K8sBaseService 是所有 K8s 资源服务的基础结构
//
// 这个基础服务封装了所有 K8s 服务共同的依赖项，消除了 30 个服务中的重复代码。
// 每个 K8s 服务（Pod, Deployment, Service等）都有相同的两个依赖：
// - storage: 用于数据持久化
// - clusterService: 用于获取 K8s 客户端
//
// 使用示例：
//
//	type K8sPodService struct {
//	    K8sBaseService  // 嵌入基础服务
//	}
//
//	func NewK8sPodService(storage *storage.MySQLStorage, clusterService *K8sClusterService) *K8sPodService {
//	    return &K8sPodService{
//	        K8sBaseService: NewK8sBaseService(storage, clusterService),
//	    }
//	}
type K8sBaseService struct {
	storage        *storage.MySQLStorage
	clusterService *K8sClusterService
}

// NewK8sBaseService 创建新的基础服务实例
//
// 参数：
//   - storage: MySQL 存储实例
//   - clusterService: 集群服务实例
//
// 返回：
//   - K8sBaseService: 基础服务实例
func NewK8sBaseService(storage *storage.MySQLStorage, clusterService *K8sClusterService) K8sBaseService {
	return K8sBaseService{
		storage:        storage,
		clusterService: clusterService,
	}
}

// Storage 返回存储实例
//
// 提供对底层存储的访问，用于数据持久化操作。
//
// 返回：
//   - *storage.MySQLStorage: MySQL 存储实例
func (b *K8sBaseService) Storage() *storage.MySQLStorage {
	return b.storage
}

// ClusterService 返回集群服务实例
//
// 提供对集群服务的访问，主要用于获取 K8s 客户端。
//
// 返回：
//   - *K8sClusterService: 集群服务实例
func (b *K8sBaseService) ClusterService() *K8sClusterService {
	return b.clusterService
}

// 设计说明：
//
// 1. 为什么使用嵌入而不是继承？
//    Go 不支持传统的继承，但通过结构体嵌入可以实现类似的效果。
//    嵌入的优势：
//    - 自动获得基础服务的所有方法
//    - 可以直接访问 storage 和 clusterService 字段
//    - 保持了类型安全
//
// 2. 为什么不使用接口？
//    虽然接口更灵活，但在这个场景下：
//    - 所有服务都需要完全相同的依赖
//    - 没有多态的需求
//    - 结构体嵌入更简单直接
//
// 3. 迁移策略：
//    现有服务可以逐步迁移到使用基础服务：
//    a) 添加 K8sBaseService 嵌入
//    b) 更新构造函数使用 NewK8sBaseService
//    c) 删除重复的字段声明
//    d) 可选：使用 b.storage 和 b.clusterService 访问器
//
// 4. 兼容性：
//    - 向后兼容：现有代码无需修改即可工作
//    - 渐进式迁移：可以一个服务一个服务地迁移
//    - 零破坏性：不影响现有的 API 和行为

