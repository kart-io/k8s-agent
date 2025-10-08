.PHONY: help deploy-all deploy-traefik deploy-prometheus deploy-grafana delete-all delete-traefik delete-prometheus delete-grafana status status-traefik status-prometheus status-grafana logs-traefik logs-prometheus logs-grafana pf-traefik pf-traefik-dashboard pf-prometheus pf-grafana validate restart restart-traefik restart-prometheus restart-grafana describe events export up down ps setup-hosts traefik-routes

help:
	@echo "K8s Agent Kustomize 部署命令"
	@echo ""
	@echo "部署命令:"
	@echo "  make deploy-all          - 部署所有服务"
	@echo "  make deploy-traefik      - 部署 Traefik"
	@echo "  make deploy-prometheus   - 部署 Prometheus"
	@echo "  make deploy-grafana      - 部署 Grafana"
	@echo ""
	@echo "删除命令:"
	@echo "  make delete-all          - 删除所有服务"
	@echo "  make delete-traefik      - 删除 Traefik"
	@echo "  make delete-prometheus   - 删除 Prometheus"
	@echo "  make delete-grafana      - 删除 Grafana"
	@echo ""
	@echo "状态命令:"
	@echo "  make status              - 查看所有服务状态"
	@echo "  make status-traefik      - 查看 Traefik 状态"
	@echo "  make status-prometheus   - 查看 Prometheus 状态"
	@echo "  make status-grafana      - 查看 Grafana 状态"
	@echo ""
	@echo "日志命令:"
	@echo "  make logs-traefik        - 查看 Traefik 日志"
	@echo "  make logs-prometheus     - 查看 Prometheus 日志"
	@echo "  make logs-grafana        - 查看 Grafana 日志"
	@echo ""
	@echo "端口转发:"
	@echo "  make pf-traefik          - 转发 Traefik 服务 (80/443)"
	@echo "  make pf-traefik-dashboard- 转发 Traefik Dashboard (8080)"
	@echo "  make pf-prometheus       - 转发 Prometheus (9090)"
	@echo "  make pf-grafana          - 转发 Grafana (3000)"
	@echo ""
	@echo "Traefik 域名配置:"
	@echo "  make setup-hosts         - 配置本地 /etc/hosts"
	@echo "  make traefik-routes      - 查看所有路由配置"
	@echo ""
	@echo "工具命令:"
	@echo "  make validate            - 验证 Kustomize 配置"
	@echo "  make restart             - 重启所有服务"
	@echo "  make describe            - 查看资源详情"
	@echo "  make events              - 查看集群事件"
	@echo "  make export              - 导出生成的 YAML"

# 部署命令
deploy-all: deploy-traefik deploy-prometheus deploy-grafana
	@echo "✅ 所有服务部署完成"

deploy-traefik:
	@echo "🚀 部署 Traefik..."
	kubectl apply -k traefik/
	@echo "✅ Traefik 部署完成"

deploy-prometheus:
	@echo "🚀 部署 Prometheus..."
	kubectl apply -k prometheus/
	@echo "✅ Prometheus 部署完成"

deploy-grafana:
	@echo "🚀 部署 Grafana..."
	kubectl apply -k grafana/
	@echo "✅ Grafana 部署完成"

# 删除命令
delete-all: delete-grafana delete-prometheus delete-traefik
	@echo "✅ 所有服务删除完成"

delete-traefik:
	@echo "🗑️  删除 Traefik..."
	kubectl delete -k traefik/
	@echo "✅ Traefik 删除完成"

delete-prometheus:
	@echo "🗑️  删除 Prometheus..."
	kubectl delete -k prometheus/
	@echo "✅ Prometheus 删除完成"

delete-grafana:
	@echo "🗑️  删除 Grafana..."
	kubectl delete -k grafana/
	@echo "✅ Grafana 删除完成"

# 状态命令
status:
	@echo "📊 Traefik 状态:"
	@kubectl get all -n traefik
	@echo ""
	@echo "📊 Monitoring 状态:"
	@kubectl get all -n monitoring

status-traefik:
	@kubectl get all -n traefik

status-prometheus:
	@kubectl get all -n monitoring -l app.kubernetes.io/name=prometheus

status-grafana:
	@kubectl get all -n monitoring -l app.kubernetes.io/name=grafana

# 日志命令
logs-traefik:
	kubectl logs -n traefik -l app=traefik --tail=100 -f

logs-prometheus:
	kubectl logs -n monitoring -l app=prometheus --tail=100 -f

logs-grafana:
	kubectl logs -n monitoring -l app=grafana --tail=100 -f

# 端口转发
pf-traefik:
	@echo "🌐 访问 Traefik Dashboard: https://traefik.k8s-agent.local"
	@echo "🌐 转发 Traefik 服务端口..."
	kubectl port-forward -n traefik svc/traefik 80:80 443:443

pf-traefik-dashboard:
	@echo "🌐 访问 Traefik Dashboard: http://localhost:8080"
	kubectl port-forward -n traefik svc/traefik-dashboard 8080:8080

pf-prometheus:
	@echo "🌐 访问 Prometheus: http://localhost:9090"
	kubectl port-forward -n monitoring svc/prometheus 9090:9090

pf-grafana:
	@echo "🌐 访问 Grafana: http://localhost:3000 (admin/admin)"
	kubectl port-forward -n monitoring svc/grafana 3000:3000

# 验证
validate:
	@echo "✅ 验证 Traefik 配置..."
	kubectl kustomize traefik/ > /dev/null && echo "  Traefik: OK"
	@echo "✅ 验证 Prometheus 配置..."
	kubectl kustomize prometheus/ > /dev/null && echo "  Prometheus: OK"
	@echo "✅ 验证 Grafana 配置..."
	kubectl kustomize grafana/ > /dev/null && echo "  Grafana: OK"
	@echo "✅ 所有配置验证通过"

# 重启命令
restart: restart-traefik restart-prometheus restart-grafana
	@echo "✅ 所有服务重启完成"

restart-traefik:
	kubectl rollout restart deployment/traefik -n traefik

restart-prometheus:
	kubectl rollout restart deployment/prometheus -n monitoring

restart-grafana:
	kubectl rollout restart deployment/grafana -n monitoring

# 调试命令
describe:
	@echo "📝 Traefik 资源详情:"
	kubectl describe deployment/traefik -n traefik
	@echo ""
	@echo "📝 Prometheus 资源详情:"
	kubectl describe deployment/prometheus -n monitoring
	@echo ""
	@echo "📝 Grafana 资源详情:"
	kubectl describe deployment/grafana -n monitoring

events:
	@echo "📋 Traefik 命名空间事件:"
	kubectl get events -n traefik --sort-by='.lastTimestamp'
	@echo ""
	@echo "📋 Monitoring 命名空间事件:"
	kubectl get events -n monitoring --sort-by='.lastTimestamp'

# 导出生成的 YAML
export:
	@echo "📦 导出 Traefik YAML..."
	kubectl kustomize traefik/ > traefik-generated.yaml
	@echo "📦 导出 Prometheus YAML..."
	kubectl kustomize prometheus/ > prometheus-generated.yaml
	@echo "📦 导出 Grafana YAML..."
	kubectl kustomize grafana/ > grafana-generated.yaml
	@echo "✅ YAML 导出完成"

# Traefik 域名配置
setup-hosts:
	@echo "🌐 配置本地 /etc/hosts..."
	@echo "需要 sudo 权限"
	@sudo sh -c 'grep -q "k8s-agent.local" /etc/hosts || echo "\n# K8s Agent Platform\n127.0.0.1 gateway.k8s-agent.local api.k8s-agent.local\n127.0.0.1 agent.k8s-agent.local agents.k8s-agent.local\n127.0.0.1 traefik.k8s-agent.local" >> /etc/hosts'
	@echo "✅ Hosts 配置完成"
	@echo ""
	@echo "已添加域名:"
	@echo "  - gateway.k8s-agent.local"
	@echo "  - api.k8s-agent.local"
	@echo "  - agent.k8s-agent.local"
	@echo "  - agents.k8s-agent.local"
	@echo "  - traefik.k8s-agent.local"

traefik-routes:
	@echo "📋 Traefik IngressRoutes:"
	@kubectl get ingressroute -n traefik
	@echo ""
	@echo "📋 Traefik Middlewares:"
	@kubectl get middleware -n traefik
	@echo ""
	@echo "🌐 配置的域名路由:"
	@echo "  - https://gateway.k8s-agent.local    -> gateway-service:8080"
	@echo "  - https://api.k8s-agent.local        -> gateway-service:8080"
	@echo "  - https://agent.k8s-agent.local      -> agent-manager-ui:80"
	@echo "  - https://agents.k8s-agent.local     -> agent-manager-ui:80"
	@echo "  - https://traefik.k8s-agent.local    -> traefik-dashboard:8080"

# 别名
up: deploy-all
down: delete-all
ps: status
