# 快速测试指南

## 前提条件验证

### 1. 编译验证 ✅

代码已成功编译：

```bash
cd /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent/cluster-service
go build -o bin/cluster-service cmd/server/main.go
```

**结果**:
- 二进制文件: `bin/cluster-service` (56MB)
- 编译时间: 2025-10-17 15:00
- 状态: ✅ 成功

### 2. 代码质量检查

#### 语法检查
```bash
go vet ./...
```

#### 格式检查
```bash
gofmt -l .
```

#### 导入检查
```bash
go mod tidy
go mod verify
```

## 快速启动测试

### 选项 A: 完整环境测试 (需要数据库)

#### 1. 准备 PostgreSQL 数据库

```bash
# 使用 Docker 快速启动
docker run -d \
  --name cluster-db \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=cluster_db \
  -p 5432:5432 \
  postgres:14
```

#### 2. 运行数据库迁移

```bash
# 如果有迁移脚本
psql -h localhost -U postgres -d cluster_db -f migrations/001_init.sql
```

#### 3. 启动服务

```bash
./bin/cluster-service -config configs/config.yaml
```

#### 4. 验证服务

```bash
# 健康检查
curl http://localhost:8082/health

# 预期响应
{
  "code": 0,
  "message": "success",
  "data": {
    "status": "healthy"
  }
}
```

### 选项 B: Mock 测试 (无需数据库)

由于完整环境需要数据库和 K8s 集群,我们可以进行代码级别的验证：

#### 1. 单元测试

```bash
# 运行所有测试
go test ./... -v

# 运行特定包的测试
go test ./internal/service -v

# 运行带覆盖率的测试
go test ./... -cover
```

#### 2. Benchmark 测试

```bash
go test ./internal/service -bench=. -benchmem
```

## API 接口测试

### 前提条件
1. ✅ 服务已启动 (端口 8082)
2. ✅ 数据库已连接
3. ✅ 至少有一个 K8s 集群配置

### 使用测试脚本

```bash
# 配置环境变量
export BASE_URL="http://localhost:8082"
export CLUSTER_ID="your-cluster-id"
export NAMESPACE="default"

# 运行测试
./test-new-apis.sh
```

### 测试覆盖的接口

#### DaemonSet API (4个)
```bash
# 1. 列表
curl http://localhost:8082/api/k8s/clusters/$CLUSTER_ID/namespaces/default/daemonsets

# 2. 详情
curl http://localhost:8082/api/k8s/clusters/$CLUSTER_ID/namespaces/default/daemonsets/kube-proxy

# 3. 重启
curl -X POST http://localhost:8082/api/k8s/clusters/$CLUSTER_ID/namespaces/default/daemonsets/kube-proxy/restart

# 4. 删除
curl -X DELETE http://localhost:8082/api/k8s/clusters/$CLUSTER_ID/namespaces/default/daemonsets/test-ds
```

#### ConfigMap API (5个)
```bash
# 1. 列表
curl http://localhost:8082/api/k8s/clusters/$CLUSTER_ID/namespaces/default/configmaps

# 2. 创建
curl -X POST http://localhost:8082/api/k8s/clusters/$CLUSTER_ID/namespaces/default/configmaps \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-config",
    "namespace": "default",
    "data": {
      "key1": "value1",
      "key2": "value2"
    }
  }'

# 3. 获取
curl http://localhost:8082/api/k8s/clusters/$CLUSTER_ID/namespaces/default/configmaps/test-config

# 4. 更新
curl -X PUT http://localhost:8082/api/k8s/clusters/$CLUSTER_ID/namespaces/default/configmaps/test-config \
  -H "Content-Type: application/json" \
  -d '{
    "data": {
      "key1": "updated-value1"
    }
  }'

# 5. 删除
curl -X DELETE http://localhost:8082/api/k8s/clusters/$CLUSTER_ID/namespaces/default/configmaps/test-config
```

#### Secret API (5个)
```bash
# 1. 列表 (不含敏感数据)
curl http://localhost:8082/api/k8s/clusters/$CLUSTER_ID/namespaces/default/secrets

# 2. 创建
curl -X POST http://localhost:8082/api/k8s/clusters/$CLUSTER_ID/namespaces/default/secrets \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-secret",
    "namespace": "default",
    "type": "Opaque",
    "stringData": {
      "username": "admin",
      "password": "secret123"
    }
  }'

# 3. 获取 (不含敏感数据)
curl http://localhost:8082/api/k8s/clusters/$CLUSTER_ID/namespaces/default/secrets/test-secret

# 4. 获取 (含敏感数据)
curl http://localhost:8082/api/k8s/clusters/$CLUSTER_ID/namespaces/default/secrets/test-secret?includeData=true

# 5. 更新
curl -X PUT http://localhost:8082/api/k8s/clusters/$CLUSTER_ID/namespaces/default/secrets/test-secret \
  -H "Content-Type: application/json" \
  -d '{
    "stringData": {
      "username": "admin",
      "password": "newsecret456"
    }
  }'

# 6. 删除
curl -X DELETE http://localhost:8082/api/k8s/clusters/$CLUSTER_ID/namespaces/default/secrets/test-secret
```

## 代码验证清单

### ✅ 已完成
- [x] 代码编译成功
- [x] Handler 层实现完整 (14 个方法)
- [x] Service 层已存在并可用
- [x] 路由注册正确
- [x] main.go 服务初始化正确
- [x] 测试脚本已创建

### ⏳ 需要运行环境
- [ ] 数据库连接测试
- [ ] K8s 集群连接测试
- [ ] 端到端 API 测试
- [ ] 并发压力测试

### 📋 推荐的测试顺序

#### 第一阶段: 代码级验证 (无需环境)
1. 运行 `go vet ./...` - 静态分析
2. 运行 `go test ./...` - 单元测试
3. 检查编译警告和错误

#### 第二阶段: 集成测试 (需要数据库)
1. 启动 PostgreSQL
2. 运行数据库迁移
3. 启动服务
4. 验证健康检查端点

#### 第三阶段: 功能测试 (需要 K8s 集群)
1. 配置至少一个 K8s 集群
2. 运行 `test-new-apis.sh`
3. 手动测试每个端点
4. 验证错误处理

#### 第四阶段: 性能测试
1. 使用 Apache Bench 或 wrk 进行负载测试
2. 监控资源使用
3. 检查日志输出
4. 验证响应时间

## 常见问题

### Q1: 编译失败
```bash
# 清理并重新编译
go clean -cache
go mod tidy
go build -o bin/cluster-service cmd/server/main.go
```

### Q2: 找不到包
```bash
# 更新依赖
cd /home/hellotalk/code/go/src/github.com/kart-io/k8s-agent
go work sync
cd cluster-service
go mod download
```

### Q3: 数据库连接失败
- 检查 PostgreSQL 是否运行: `docker ps | grep postgres`
- 检查端口: `netstat -tlnp | grep 5432`
- 验证配置: `cat configs/config.yaml`

### Q4: K8s 集群连接失败
- 检查 kubeconfig: `kubectl cluster-info`
- 验证集群配置在数据库中
- 检查服务日志

## 性能基准

### 预期性能指标
- 健康检查: < 10ms
- 列表接口 (10 items): < 100ms
- 详情接口: < 50ms
- 创建/更新接口: < 200ms
- 删除接口: < 100ms

### 压力测试
```bash
# 使用 Apache Bench
ab -n 1000 -c 10 http://localhost:8082/health

# 使用 wrk
wrk -t4 -c100 -d30s http://localhost:8082/health
```

## 下一步

### 短期 (本周)
1. 在开发环境运行完整测试
2. 修复发现的任何问题
3. 添加单元测试

### 中期 (下周)
1. 在测试环境部署
2. 集成到 CI/CD
3. 添加监控告警

### 长期 (本月)
1. 生产环境部署
2. 性能优化
3. 开始 Phase 2 实现

## 参考文档

- [PHASE1_COMPLETION_REPORT.md](./PHASE1_COMPLETION_REPORT.md) - Phase 1 详细报告
- [API_IMPLEMENTATION_PLAN.md](./API_IMPLEMENTATION_PLAN.md) - 完整实现计划
- [LATEST_UPDATE.md](./LATEST_UPDATE.md) - 最新更新说明
- [test-new-apis.sh](./test-new-apis.sh) - 自动化测试脚本

---

**文档创建时间**: 2025-10-17 15:00
**状态**: ✅ 代码验证完成，等待环境测试
**作者**: Claude (AI Assistant)
