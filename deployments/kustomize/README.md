# K8s Agent Kustomize 部署

## 快速开始

### 部署所有服务

```bash
# 部署 Traefik
kubectl apply -k deployments/kustomize/traefik/

# 部署 Prometheus
kubectl apply -k deployments/kustomize/prometheus/

# 部署 Grafana
kubectl apply -k deployments/kustomize/grafana/
```

### 访问服务

**Traefik Dashboard:**
```bash
kubectl port-forward -n traefik svc/traefik-dashboard 8080:8080
# 访问 http://localhost:8080
```

**Prometheus:**
```bash
kubectl port-forward -n monitoring svc/prometheus 9090:9090
# 访问 http://localhost:9090
```

**Grafana:**
```bash
kubectl port-forward -n monitoring svc/grafana 3000:3000
# 访问 http://localhost:3000 (admin/admin)
```

## 删除部署

```bash
kubectl delete -k deployments/kustomize/traefik/
kubectl delete -k deployments/kustomize/prometheus/
kubectl delete -k deployments/kustomize/grafana/
```
