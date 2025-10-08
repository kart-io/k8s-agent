# Traefik 域名配置说明

## 配置的域名

已配置以下域名路由：

| 域名 | 服务 | 命名空间 | 端口 | 说明 |
|------|------|----------|------|------|
| `gateway.k8s-agent.local` | gateway-service | default | 8080 | API 网关 |
| `api.k8s-agent.local` | gateway-service | default | 8080 | API 网关（别名） |
| `agent.k8s-agent.local` | agent-manager-ui | default | 80 | Agent 管理界面 |
| `agents.k8s-agent.local` | agent-manager-ui | default | 80 | Agent 管理界面（别名） |
| `traefik.k8s-agent.local` | traefik-dashboard | traefik | 8080 | Traefik 仪表板 |

## 本地开发配置

### 1. 添加 hosts 记录

编辑 `/etc/hosts` 文件，添加以下记录：

```bash
sudo nano /etc/hosts
```

添加：
```
127.0.0.1 gateway.k8s-agent.local
127.0.0.1 api.k8s-agent.local
127.0.0.1 agent.k8s-agent.local
127.0.0.1 agents.k8s-agent.local
127.0.0.1 traefik.k8s-agent.local
```

### 2. 端口转发（使用 LoadBalancer）

如果在本地 Kubernetes 集群（如 Minikube、Kind）运行：

```bash
# 转发 HTTPS 端口
kubectl port-forward -n traefik svc/traefik 443:443 80:80
```

### 3. 访问服务

- **API 网关**: https://gateway.k8s-agent.local 或 https://api.k8s-agent.local
- **Agent 管理界面**: https://agent.k8s-agent.local 或 https://agents.k8s-agent.local
- **Traefik 仪表板**: https://traefik.k8s-agent.local
  - 用户名: `admin`
  - 密码: `admin`（建议在生产环境修改）

## 生产环境配置

### 1. 修改域名

编辑 `ingressroute.yaml`，将 `.k8s-agent.local` 替换为实际域名：

```yaml
- match: Host(`gateway.yourdomain.com`)
```

### 2. 配置 SSL 证书

修改 `configmap.yaml` 中的邮箱地址：

```yaml
certificatesResolvers:
  letsencrypt:
    acme:
      email: your-email@example.com  # 修改为真实邮箱
```

### 3. 修改 Dashboard 密码

生成新密码：
```bash
htpasswd -nb admin <your-password>
```

更新 `ingressroute.yaml` 中的 Secret：
```yaml
stringData:
  users: |
    admin:<generated-hash>
```

### 4. 配置 CORS

修改 `ingressroute.yaml` 中的 CORS 允许源：

```yaml
accessControlAllowOriginList:
  - "https://yourdomain.com"
  - "https://app.yourdomain.com"
```

## Middleware 说明

### cors-headers
- 处理跨域请求
- 允许的方法：GET, POST, PUT, DELETE, PATCH, OPTIONS
- 支持凭证传递

### auth-middleware
- Traefik Dashboard 的 Basic Auth 认证
- 默认用户名/密码: admin/admin

## 部署

```bash
# 应用配置
kubectl apply -k .

# 查看 IngressRoute
kubectl get ingressroute -n traefik

# 查看 Middleware
kubectl get middleware -n traefik

# 查看服务状态
kubectl get svc -n traefik
```

## 故障排查

### 1. 查看 Traefik 日志
```bash
kubectl logs -n traefik -l app=traefik -f
```

### 2. 检查 IngressRoute 状态
```bash
kubectl describe ingressroute -n traefik gateway-https
```

### 3. 测试域名解析
```bash
curl -k https://gateway.k8s-agent.local
curl -k https://agent.k8s-agent.local
```

### 4. 检查证书
```bash
kubectl get secret -n traefik
kubectl describe secret traefik-dashboard-auth -n traefik
```

## 高级配置

### 添加新的路由

创建新的 IngressRoute：

```yaml
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  name: my-service-https
  namespace: traefik
spec:
  entryPoints:
    - websecure
  routes:
    - match: Host(`myapp.k8s-agent.local`)
      kind: Rule
      services:
        - name: my-service
          namespace: default
          port: 8080
  tls:
    certResolver: letsencrypt
```

### 添加速率限制

```yaml
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: rate-limit
  namespace: traefik
spec:
  rateLimit:
    average: 100
    burst: 50
```

### 添加重试机制

```yaml
apiVersion: traefik.containo.us/v1alpha1
kind: Middleware
metadata:
  name: retry
  namespace: traefik
spec:
  retry:
    attempts: 3
    initialInterval: 100ms
```
