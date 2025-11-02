# Server 目录重构方案

## 目标
将 server 包按功能分成子目录，但保持所有文件在同一个 package server 中。

## 新目录结构

```
common/server/
├── server.go              # 核心接口（package server）
├── README.md              
│
├── http/                  # HTTP 服务器（package server）
│   ├── gin.go            
│   ├── kratos.go         
│   └── options.go        
│
├── grpc/                  # gRPC 服务器（package server）
│   ├── standard.go       
│   ├── options.go        
│   ├── interceptors.go   
│   └── health.go         
│
└── internal/              # 共享代码（package internal）
    ├── middleware.go     
    └── health.go         
```

## Go 包规则

1. **http/** 和 **grpc/** 目录中的文件：`package server`
   - 与根目录的 server.go 在同一个包中
   - 可以直接访问彼此的类型和函数
   - 用户导入：`import "github.com/kart-io/k8s-agent/common/server"`

2. **internal/** 目录中的文件：`package internal`
   - 独立的内部包
   - 其他文件需要导入：`import "github.com/kart-io/k8s-agent/common/server/internal"`

## Import 规则

### http/*.go 和 grpc/*.go
```go
package server

import (
    "github.com/kart-io/k8s-agent/common/server/internal"  // 导入 internal 包
    // ... 其他导入
)
```

### internal/*.go
```go
package internal  // 独立包名

import (
    // 不能导入 server 包（循环依赖）
)
```

## 实施步骤

1. ✅ 创建子目录
2. ✅ 移动文件
3. ⚠️ 修改 internal/*.go 的 package 为 internal
4. ⚠️ 更新其他文件的 import
5. ⚠️ 验证编译
