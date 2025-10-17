# 集群添加脚本使用指南

## 脚本概述

`add-cluster.sh` 是一个便捷的命令行工具，用于将 Kubernetes 集群添加到 cluster-service 管理系统。

**脚本位置**: `scripts/add-cluster.sh`

## 快速开始

### 1. 快速添加 Minikube 集群

```bash
# 确保 minikube 正在运行
minikube status

# 执行快速添加
./scripts/add-cluster.sh minikube
```

**输出示例**:
```json
{
  "code": 0,
  "message": "Cluster created successfully",
  "data": {
    "id": "a24170ff-492f-4431-a9d1-d16708f594d3",
    "name": "minikube-local",
    "description": "Local Minikube Cluster",
    "endpoint": "https://192.168.58.2:8443",
    "version": "v1.30.0",
    "status": "healthy",
    "region": "local",
    "provider": "minikube"
  }
}

集群 ID: a24170ff-492f-4431-a9d1-d16708f594d3

测试命令:
  # 获取集群详情
  curl -s http://127.0.0.1:8082/api/k8s/clusters/a24170ff-492f-4431-a9d1-d16708f594d3 | jq .

  # 列出命名空间
  curl -s http://127.0.0.1:8082/api/k8s/clusters/a24170ff-492f-4431-a9d1-d16708f594d3/namespaces | jq .

  # 列出 Pods
  curl -s http://127.0.0.1:8082/api/k8s/clusters/a24170ff-492f-4431-a9d1-d16708f594d3/namespaces/default/pods | jq .
```

### 2. 交互式添加集群

```bash
./scripts/add-cluster.sh interactive
# 或
./scripts/add-cluster.sh -i
```

**交互流程**:
1. 选择 kubectl context
2. 输入集群显示名称
3. 输入集群描述
4. 选择提供商 (minikube/aws/gcp/azure/self-hosted)
5. 输入区域
6. 确认并提交

**示例交互**:
```
可用的 Kubernetes contexts:
     1  minikube
     2  production-cluster
     3  dev-cluster

请选择 context (留空使用当前 context): 1

使用 context: minikube

检测到的集群信息:
  Context: minikube
  集群名称: minikube
  Endpoint: https://192.168.58.2:8443
  版本: v1.30.0

集群显示名称 [minikube]: minikube-local
集群描述: Local development cluster
提供商 (minikube/aws/gcp/azure/self-hosted) [self-hosted]: minikube
区域 [local]: local

准备添加集群，请确认信息:
  名称: minikube-local
  描述: Local development cluster
  Endpoint: https://192.168.58.2:8443
  提供商: minikube
  区域: local

确认添加？(y/N): y
```

### 3. 从 JSON 文件添加集群

```bash
./scripts/add-cluster.sh file /path/to/cluster.json
# 或
./scripts/add-cluster.sh -f /path/to/cluster.json
```

**JSON 文件格式**:
```json
{
  "name": "production-cluster",
  "description": "Production Kubernetes Cluster",
  "endpoint": "https://k8s-api.example.com:6443",
  "provider": "aws",
  "region": "us-east-1",
  "kubeconfig": "apiVersion: v1\nclusters:\n- cluster:\n    certificate-authority-data: LS0tLS1...\n    server: https://k8s-api.example.com:6443\n  name: production\n..."
}
```

**获取 kubeconfig 内容**:
```bash
# 获取完整的 kubeconfig（已展开和格式化）
kubectl config view --minify --flatten --raw

# 或针对特定 context
kubectl config view --minify --flatten --raw --context=production-cluster
```

### 4. 列出已添加的集群

```bash
./scripts/add-cluster.sh list
# 或
./scripts/add-cluster.sh -l
```

**输出示例**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": "a24170ff-492f-4431-a9d1-d16708f594d3",
        "name": "minikube-local",
        "description": "Local Minikube Cluster",
        "endpoint": "https://192.168.58.2:8443",
        "version": "v1.30.0",
        "status": "healthy",
        "region": "local",
        "provider": "minikube"
      }
    ],
    "total": 1
  }
}
```

## 高级用法

### 自定义 Cluster Service 地址

```bash
# 使用环境变量指定服务地址
CLUSTER_SERVICE_URL=http://192.168.1.100:8082 ./scripts/add-cluster.sh minikube

# 或
export CLUSTER_SERVICE_URL=http://192.168.1.100:8082
./scripts/add-cluster.sh minikube
```

### 创建 JSON 文件模板

```bash
# 自动生成当前 context 的 JSON 文件
cat > my-cluster.json <<EOF
{
  "name": "$(kubectl config current-context)",
  "description": "My Kubernetes Cluster",
  "endpoint": "$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}')",
  "provider": "self-hosted",
  "region": "local",
  "kubeconfig": $(kubectl config view --minify --flatten --raw | jq -Rs .)
}
EOF

# 添加集群
./scripts/add-cluster.sh file my-cluster.json
```

## 依赖工具

脚本需要以下工具：
- `kubectl` - Kubernetes 命令行工具
- `curl` - HTTP 客户端
- `jq` - JSON 处理工具

**安装依赖** (Ubuntu/Debian):
```bash
sudo apt-get update
sudo apt-get install -y kubectl curl jq
```

## 故障排查

### 1. Minikube 未运行

**错误**:
```
[ERROR] Minikube 未运行
```

**解决方法**:
```bash
minikube start
```

### 2. 无法连接到 cluster-service

**错误**:
```
curl: (7) Failed to connect to 127.0.0.1 port 8082: Connection refused
```

**解决方法**:
```bash
# 检查服务是否运行
curl http://127.0.0.1:8082/health

# 如果未运行，启动服务
cd cluster-service
make run-local
```

### 3. kubeconfig 解析失败

**错误**:
```
failed to parse kubeconfig: invalid configuration
```

**解决方法**:
- 确保 kubectl 能正常访问集群
- 检查 kubeconfig 格式是否正确
- 使用 `--flatten` 和 `--raw` 参数获取完整配置

### 4. 集群添加成功但状态为 unhealthy

**可能原因**:
- 集群 API server 不可达
- 证书配置问题
- 网络连接问题

**检查方法**:
```bash
# 测试集群连接
kubectl cluster-info --context=your-context

# 查看集群详情
CLUSTER_ID="your-cluster-id"
curl -s http://127.0.0.1:8082/api/k8s/clusters/$CLUSTER_ID | jq .
```

## 脚本命令参考

| 命令 | 简写 | 描述 |
|------|------|------|
| `interactive` | `-i` | 交互式添加集群 |
| `minikube` | `-m` | 快速添加 Minikube 集群 |
| `file <path>` | `-f <path>` | 从 JSON 文件添加集群 |
| `list` | `-l` | 列出已添加的集群 |
| `help` | `-h` | 显示帮助信息 |

## API 测试示例

添加集群后，可以使用以下命令测试 API：

```bash
# 设置集群 ID
CLUSTER_ID="a24170ff-492f-4431-a9d1-d16708f594d3"
BASE_URL="http://127.0.0.1:8082/api/k8s"

# 1. 获取集群详情
curl -s $BASE_URL/clusters/$CLUSTER_ID | jq .

# 2. 列出命名空间
curl -s $BASE_URL/clusters/$CLUSTER_ID/namespaces | jq .

# 3. 列出 default 命名空间的 Pods
curl -s $BASE_URL/clusters/$CLUSTER_ID/namespaces/default/pods | jq .

# 4. 列出节点
curl -s $BASE_URL/clusters/$CLUSTER_ID/nodes | jq .

# 5. 列出 default 命名空间的 Services
curl -s $BASE_URL/clusters/$CLUSTER_ID/namespaces/default/services | jq .

# 6. 列出 kube-system 命名空间的 Deployments
curl -s $BASE_URL/clusters/$CLUSTER_ID/namespaces/kube-system/deployments | jq .

# 7. 列出 StatefulSets
curl -s $BASE_URL/clusters/$CLUSTER_ID/namespaces/default/statefulsets | jq .

# 8. 列出 DaemonSets
curl -s $BASE_URL/clusters/$CLUSTER_ID/namespaces/kube-system/daemonsets | jq .

# 9. 列出 ConfigMaps
curl -s $BASE_URL/clusters/$CLUSTER_ID/namespaces/default/configmaps | jq .

# 10. 列出 Secrets
curl -s $BASE_URL/clusters/$CLUSTER_ID/namespaces/default/secrets | jq .
```

## 生产环境使用建议

1. **安全性**:
   - 不要在生产环境中使用明文存储 kubeconfig
   - 使用 RBAC 限制 cluster-service 的访问权限
   - 启用 TLS/SSL 加密通信

2. **网络**:
   - 确保 cluster-service 可以访问 K8s API server
   - 配置防火墙规则允许必要的端口
   - 考虑使用 VPN 或私有网络

3. **监控**:
   - 定期检查集群状态
   - 设置告警监控集群健康状态
   - 记录所有集群操作日志

4. **备份**:
   - 定期备份集群配置
   - 保存 kubeconfig 副本
   - 记录集群元数据信息

## 相关文档

- [API 测试报告](./api-test-report.md)
- [集群配置文件示例](../scripts/add-cluster-example.json)
- [Cluster Service API 文档](http://127.0.0.1:8082/version)
