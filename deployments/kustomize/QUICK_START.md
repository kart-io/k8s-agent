# Kustomize 部署快速入门

## 目录结构

```
deployments/kustomize/
├── Makefile           # 部署命令集合
├── traefik/           # Traefik Ingress 控制器
│   ├── ingressroute.yaml   # 域名路由配置
│   └── README.md
├── prometheus/        # Prometheus 监控
└── grafana/          # Grafana 可视化
```

## 快速部署

### 1. 部署 Traefik（含域名路由）

```bash
cd deployments/kustomize

# 部署 Traefik
make deploy-traefik

# 配置本地域名解析
make setup-hosts

# 启动端口转发
make pf-traefik
```

### 2. 访问服务

配置完成后，可通过以下域名访问：

| 服务 | 域名 | 说明 |
|------|------|------|
| API 网关 | https://gateway.k8s-agent.local | 后端 API 网关 |
| API 别名 | https://api.k8s-agent.local | 网关别名 |
| Agent UI | https://agent.k8s-agent.local | Agent 管理界面 |
| Agent UI 别名 | https://agents.k8s-agent.local | 管理界面别名 |
| Traefik Dashboard | https://traefik.k8s-agent.local | Traefik 控制台 (admin/admin) |

### 3. 部署监控栈

```bash
# 部署 Prometheus
make deploy-prometheus

# 部署 Grafana
make deploy-grafana

# 或一次性部署所有服务
make deploy-all
```

## 常用命令

### 部署命令

```bash
make deploy-all          # 部署所有服务（Traefik + Prometheus + Grafana）
make deploy-traefik      # 仅部署 Traefik
make deploy-prometheus   # 仅部署 Prometheus
make deploy-grafana      # 仅部署 Grafana
```

### 删除命令

```bash
make delete-all          # 删除所有服务
make delete-traefik      # 删除 Traefik
make delete-prometheus   # 删除 Prometheus
make delete-grafana      # 删除 Grafana
```

### 状态查看

```bash
make status              # 查看所有服务状态
make status-traefik      # 查看 Traefik 状态
make traefik-routes      # 查看 Traefik 路由配置
```

### 日志查看

```bash
make logs-traefik        # 查看 Traefik 日志（实时）
make logs-prometheus     # 查看 Prometheus 日志
make logs-grafana        # 查看 Grafana 日志
```

### 端口转发

```bash
make pf-traefik          # 转发 Traefik 服务（80/443）
make pf-traefik-dashboard # 转发 Traefik Dashboard（8080）
make pf-prometheus       # 转发 Prometheus（9090）
make pf-grafana          # 转发 Grafana（3000）
```

### 域名配置

```bash
make setup-hosts         # 配置 /etc/hosts（需要 sudo）
make traefik-routes      # 查看配置的域名路由
```

### 工具命令

```bash
make validate            # 验证 Kustomize 配置
make restart             # 重启所有服务
make describe            # 查看资源详情
make events              # 查看集群事件
make export              # 导出生成的 YAML 文件
```

### 别名

```bash
make up                  # = deploy-all
make down                # = delete-all
make ps                  # = status
```

## 完整部署流程

```bash
# 1. 进入 kustomize 目录
cd deployments/kustomize

# 2. 查看可用命令
make help

# 3. 验证配置
make validate

# 4. 部署所有服务
make deploy-all

# 5. 配置本地域名
make setup-hosts

# 6. 等待服务就绪
make status

# 7. 查看 Traefik 路由
make traefik-routes

# 8. 启动端口转发（新终端窗口）
make pf-traefik

# 9. 访问服务
# - Gateway: https://gateway.k8s-agent.local
# - Agent UI: https://agent.k8s-agent.local
# - Traefik: https://traefik.k8s-agent.local
```

## 故障排查

### 1. 查看 Pod 状态

```bash
make status-traefik
```

### 2. 查看日志

```bash
make logs-traefik
```

### 3. 查看事件

```bash
make events
```

### 4. 查看详细信息

```bash
make describe
```

### 5. 重启服务

```bash
make restart-traefik
```

## 域名配置详解

### 自动配置（推荐）

```bash
make setup-hosts
```

### 手动配置

编辑 `/etc/hosts`：

```bash
sudo nano /etc/hosts
```

添加：

```
127.0.0.1 gateway.k8s-agent.local api.k8s-agent.local
127.0.0.1 agent.k8s-agent.local agents.k8s-agent.local
127.0.0.1 traefik.k8s-agent.local
```

### 验证配置

```bash
# 测试域名解析
ping gateway.k8s-agent.local

# 测试 HTTPS 访问
curl -k https://gateway.k8s-agent.local
```

## 生产环境配置

### 修改域名

编辑 `traefik/ingressroute.yaml`，替换 `.k8s-agent.local` 为实际域名：

```yaml
- match: Host(`gateway.yourdomain.com`)
```

### 配置 SSL 证书

编辑 `traefik/configmap.yaml`，修改 Let's Encrypt 邮箱：

```yaml
certificatesResolvers:
  letsencrypt:
    acme:
      email: your-email@example.com
```

### 修改 Dashboard 密码

```bash
# 生成新密码
htpasswd -nb admin <your-password>

# 更新 traefik/ingressroute.yaml 中的 Secret
```

## 监控访问

### Prometheus

```bash
# 端口转发
make pf-prometheus

# 访问
open http://localhost:9090
```

### Grafana

```bash
# 端口转发
make pf-grafana

# 访问（默认密码 admin/admin）
open http://localhost:3000
```

## 卸载

```bash
# 删除所有服务
make delete-all

# 验证删除
kubectl get all -n traefik
kubectl get all -n monitoring
```

## 从根目录使用

如果在项目根目录，可以使用：

```bash
# 部署 Kustomize 服务
make k8s-kustomize

# 或直接调用 kustomize Makefile
cd deployments/kustomize && make deploy-all
```

## 注意事项

1. **端口冲突**：确保本地 80/443 端口未被占用
2. **Kubernetes 版本**：需要 Kubernetes 1.19+
3. **kubectl 配置**：确保 kubectl 已正确配置并连接到集群
4. **权限**：`setup-hosts` 需要 sudo 权限
5. **证书**：本地开发使用自签名证书，浏览器会显示安全警告

## 更多信息

- [Traefik 配置详解](traefik/README.md)
- [项目根目录 Makefile](../../Makefile)
- [Traefik 官方文档](https://doc.traefik.io/traefik/)
