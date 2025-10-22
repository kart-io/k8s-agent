.PHONY: help deploy-all deploy-infra deploy-monitoring deploy-gateway deploy-traefik deploy-prometheus deploy-grafana deploy-neo4j deploy-nats deploy-redis deploy-postgres delete-all delete-infra delete-monitoring delete-traefik delete-prometheus delete-grafana delete-neo4j delete-nats delete-redis delete-postgres status status-all status-traefik status-prometheus status-grafana status-neo4j status-nats status-redis status-postgres logs-traefik logs-prometheus logs-grafana logs-neo4j logs-nats logs-redis logs-postgres pf-traefik pf-traefik-dashboard pf-prometheus pf-grafana pf-neo4j pf-nats pf-redis pf-postgres validate restart restart-all restart-traefik restart-prometheus restart-grafana restart-neo4j restart-nats restart-redis restart-postgres describe events export up down ps setup-hosts traefik-routes

help:
	@echo "K8s Agent Kustomize 部署命令"
	@echo ""
	@echo "快速部署:"
	@echo "  make deploy-all          - 部署所有服务（基础设施 + 监控 + 网关）"
	@echo "  make deploy-infra        - 部署基础设施服务（Neo4j, NATS, Redis, PostgreSQL）"
	@echo "  make deploy-monitoring   - 部署监控服务（Prometheus, Grafana）"
	@echo "  make deploy-gateway      - 部署网关服务（Traefik）"
	@echo ""
	@echo "基础设施部署:"
	@echo "  make deploy-neo4j        - 部署 Neo4j 图数据库"
	@echo "  make deploy-nats         - 部署 NATS 消息队列"
	@echo "  make deploy-redis        - 部署 Redis 缓存"
	@echo "  make deploy-postgres     - 部署 PostgreSQL 数据库"
	@echo ""
	@echo "监控和网关部署:"
	@echo "  make deploy-traefik      - 部署 Traefik 网关"
	@echo "  make deploy-prometheus   - 部署 Prometheus 监控"
	@echo "  make deploy-grafana      - 部署 Grafana 可视化"
	@echo ""
	@echo "删除命令:"
	@echo "  make delete-all          - 删除所有服务"
	@echo "  make delete-infra        - 删除基础设施服务"
	@echo "  make delete-monitoring   - 删除监控服务"
	@echo "  make delete-neo4j        - 删除 Neo4j"
	@echo "  make delete-nats         - 删除 NATS"
	@echo "  make delete-redis        - 删除 Redis"
	@echo "  make delete-postgres     - 删除 PostgreSQL"
	@echo "  make delete-traefik      - 删除 Traefik"
	@echo "  make delete-prometheus   - 删除 Prometheus"
	@echo "  make delete-grafana      - 删除 Grafana"
	@echo ""
	@echo "状态命令:"
	@echo "  make status              - 查看所有服务状态"
	@echo "  make status-neo4j        - 查看 Neo4j 状态"
	@echo "  make status-nats         - 查看 NATS 状态"
	@echo "  make status-redis        - 查看 Redis 状态"
	@echo "  make status-postgres     - 查看 PostgreSQL 状态"
	@echo "  make status-traefik      - 查看 Traefik 状态"
	@echo "  make status-prometheus   - 查看 Prometheus 状态"
	@echo "  make status-grafana      - 查看 Grafana 状态"
	@echo ""
	@echo "日志命令:"
	@echo "  make logs-neo4j          - 查看 Neo4j 日志"
	@echo "  make logs-nats           - 查看 NATS 日志"
	@echo "  make logs-redis          - 查看 Redis 日志"
	@echo "  make logs-postgres       - 查看 PostgreSQL 日志"
	@echo "  make logs-traefik        - 查看 Traefik 日志"
	@echo "  make logs-prometheus     - 查看 Prometheus 日志"
	@echo "  make logs-grafana        - 查看 Grafana 日志"
	@echo ""
	@echo "端口转发:"
	@echo "  make pf-neo4j            - 转发 Neo4j (7474:HTTP, 7687:Bolt)"
	@echo "  make pf-nats             - 转发 NATS (4222)"
	@echo "  make pf-redis            - 转发 Redis (6379)"
	@echo "  make pf-postgres         - 转发 PostgreSQL (5432)"
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
	@echo "  make restart-all         - 重启所有服务"
	@echo "  make describe            - 查看资源详情"
	@echo "  make events              - 查看集群事件"
	@echo "  make export              - 导出生成的 YAML"
	@echo ""
	@echo "别名:"
	@echo "  make up                  - 等同于 deploy-all"
	@echo "  make down                - 等同于 delete-all"
	@echo "  make ps                  - 等同于 status"

# ============================================================================
# 部署命令 - 组合部署
# ============================================================================

deploy-all: deploy-infra deploy-monitoring deploy-gateway
	@echo "✅ 所有服务部署完成"

deploy-infra: deploy-neo4j deploy-nats deploy-redis deploy-postgres
	@echo "✅ 基础设施服务部署完成"

deploy-monitoring: deploy-prometheus deploy-grafana
	@echo "✅ 监控服务部署完成"

deploy-gateway: deploy-traefik
	@echo "✅ 网关服务部署完成"

# ============================================================================
# 部署命令 - 基础设施
# ============================================================================

deploy-neo4j:
	@echo "🚀 部署 Neo4j 图数据库..."
	kubectl apply -k base/neo4j/
	@echo "✅ Neo4j 部署完成"
	@echo "   访问: kubectl port-forward -n aetherius svc/neo4j 7474:7474 7687:7687"

deploy-nats:
	@echo "🚀 部署 NATS 消息队列..."
	kubectl apply -k base/nats/
	@echo "✅ NATS 部署完成"
	@echo "   访问: kubectl port-forward -n aetherius svc/nats 4222:4222"

deploy-redis:
	@echo "🚀 部署 Redis 缓存..."
	kubectl apply -k base/redis/
	kubectl apply -k base/redis/
	@echo "✅ Redis 部署完成"
	@echo "   访问: kubectl port-forward -n aetherius svc/redis 6379:6379"

deploy-postgres:
	@echo "🚀 部署 PostgreSQL 数据库..."
	kubectl apply -k base/postgres/
	@echo "✅ PostgreSQL 部署完成"
	@echo "   访问: kubectl port-forward -n aetherius svc/postgres 5432:5432"

# ============================================================================
# 部署命令 - 监控和网关
# ============================================================================

deploy-traefik:
	@echo "🚀 部署 Traefik 网关..."
	kubectl apply -k traefik/
	@echo "✅ Traefik 部署完成"

deploy-prometheus:
	@echo "🚀 部署 Prometheus 监控..."
	kubectl apply -k prometheus/
	@echo "✅ Prometheus 部署完成"

deploy-grafana:
	@echo "🚀 部署 Grafana 可视化..."
	kubectl apply -k grafana/
	@echo "✅ Grafana 部署完成"

# ============================================================================
# 删除命令 - 组合删除
# ============================================================================

delete-all: delete-gateway delete-monitoring delete-infra
	@echo "✅ 所有服务删除完成"

delete-infra: delete-postgres delete-redis delete-nats delete-neo4j
	@echo "✅ 基础设施服务删除完成"

delete-monitoring: delete-grafana delete-prometheus
	@echo "✅ 监控服务删除完成"

delete-gateway: delete-traefik
	@echo "✅ 网关服务删除完成"

# ============================================================================
# 删除命令 - 基础设施
# ============================================================================

delete-neo4j:
	@echo "🗑️  删除 Neo4j..."
	kubectl delete -k base/neo4j/
	@echo "✅ Neo4j 删除完成"

delete-nats:
	@echo "🗑️  删除 NATS..."
	kubectl delete -k base/nats/
	@echo "✅ NATS 删除完成"

delete-redis:
	@echo "🗑️  删除 Redis..."
	kubectl delete -k base/redis/
	@echo "✅ Redis 删除完成"

delete-postgres:
	@echo "🗑️  删除 PostgreSQL..."
	kubectl delete -k base/postgres/
	@echo "✅ PostgreSQL 删除完成"

# ============================================================================
# 删除命令 - 监控和网关
# ============================================================================

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

# ============================================================================
# 状态命令
# ============================================================================

status:
	@echo "📊 === Aetherius 命名空间状态 ==="
	@kubectl get all -n aetherius 2>/dev/null || echo "命名空间 aetherius 不存在"
	@echo ""
	@echo "📊 === Traefik 命名空间状态 ==="
	@kubectl get all -n traefik 2>/dev/null || echo "命名空间 traefik 不存在"
	@echo ""
	@echo "📊 === Monitoring 命名空间状态 ==="
	@kubectl get all -n monitoring 2>/dev/null || echo "命名空间 monitoring 不存在"

status-neo4j:
	@echo "📊 Neo4j 状态:"
	@kubectl get all -n aetherius -l app=neo4j

status-nats:
	@echo "📊 NATS 状态:"
	@kubectl get all -n aetherius -l app=nats

status-redis:
	@echo "📊 Redis 状态:"
	@kubectl get all -n aetherius -l app=redis

status-postgres:
	@echo "📊 PostgreSQL 状态:"
	@kubectl get all -n aetherius -l app=postgres

status-traefik:
	@kubectl get all -n traefik

status-prometheus:
	@kubectl get all -n monitoring -l app.kubernetes.io/name=prometheus

status-grafana:
	@kubectl get all -n monitoring -l app.kubernetes.io/name=grafana

# ============================================================================
# 日志命令
# ============================================================================

logs-neo4j:
	kubectl logs -n aetherius -l app=neo4j --tail=100 -f

logs-nats:
	kubectl logs -n aetherius -l app=nats --tail=100 -f

logs-redis:
	kubectl logs -n aetherius -l app=redis --tail=100 -f

logs-postgres:
	kubectl logs -n aetherius -l app=postgres --tail=100 -f

logs-traefik:
	kubectl logs -n traefik -l app=traefik --tail=100 -f

logs-prometheus:
	kubectl logs -n monitoring -l app=prometheus --tail=100 -f

logs-grafana:
	kubectl logs -n monitoring -l app=grafana --tail=100 -f

# ============================================================================
# 端口转发
# ============================================================================

pf-neo4j:
	@echo "🌐 访问 Neo4j Browser: http://localhost:7474"
	@echo "🌐 Neo4j Bolt 协议: bolt://localhost:7687"
	kubectl port-forward -n aetherius svc/neo4j 7474:7474 7687:7687

pf-nats:
	@echo "🌐 NATS 客户端端口: localhost:4222"
	kubectl port-forward -n aetherius svc/nats 4222:4222

pf-redis:
	@echo "🌐 Redis 端口: localhost:6379"
	@echo "连接命令: redis-cli -h localhost -p 6379"
	kubectl port-forward -n aetherius svc/redis 6379:6379

pf-postgres:
	@echo "🌐 PostgreSQL 端口: localhost:5432"
	@echo "连接命令: psql -h localhost -p 5432 -U postgres"
	kubectl port-forward -n aetherius svc/postgres 5432:5432

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

# ============================================================================
# 验证
# ============================================================================

validate:
	@echo "✅ 验证 Neo4j 配置..."
	@kubectl kustomize neo4j/ > /dev/null && echo "  Neo4j: OK" || echo "  Neo4j: FAILED"
	@echo "✅ 验证 NATS 配置..."
	@kubectl kustomize nats/ > /dev/null && echo "  NATS: OK" || echo "  NATS: FAILED"
	@echo "✅ 验证 Redis 配置..."
	@kubectl kustomize redis/ > /dev/null && echo "  Redis: OK" || echo "  Redis: FAILED"
	@echo "✅ 验证 PostgreSQL 配置..."
	@kubectl kustomize postgres/ > /dev/null && echo "  PostgreSQL: OK" || echo "  PostgreSQL: FAILED"
	@echo "✅ 验证 Traefik 配置..."
	@kubectl kustomize traefik/ > /dev/null && echo "  Traefik: OK" || echo "  Traefik: FAILED"
	@echo "✅ 验证 Prometheus 配置..."
	@kubectl kustomize prometheus/ > /dev/null && echo "  Prometheus: OK" || echo "  Prometheus: FAILED"
	@echo "✅ 验证 Grafana 配置..."
	@kubectl kustomize grafana/ > /dev/null && echo "  Grafana: OK" || echo "  Grafana: FAILED"
	@echo "✅ 所有配置验证通过"

# ============================================================================
# 重启命令
# ============================================================================

restart-all: restart-neo4j restart-nats restart-redis restart-postgres restart-traefik restart-prometheus restart-grafana
	@echo "✅ 所有服务重启完成"

restart-neo4j:
	@echo "🔄 重启 Neo4j..."
	kubectl rollout restart statefulset/neo4j -n aetherius

restart-nats:
	@echo "🔄 重启 NATS..."
	kubectl rollout restart statefulset/nats -n aetherius

restart-redis:
	@echo "🔄 重启 Redis..."
	kubectl rollout restart statefulset/redis -n aetherius

restart-postgres:
	@echo "🔄 重启 PostgreSQL..."
	kubectl rollout restart statefulset/postgres -n aetherius

restart-traefik:
	kubectl rollout restart deployment/traefik -n traefik

restart-prometheus:
	kubectl rollout restart deployment/prometheus -n monitoring

restart-grafana:
	kubectl rollout restart deployment/grafana -n monitoring

# ============================================================================
# 调试命令
# ============================================================================

describe:
	@echo "📝 === Neo4j 资源详情 ==="
	@kubectl describe statefulset/neo4j -n aetherius 2>/dev/null || echo "Neo4j 未部署"
	@echo ""
	@echo "📝 === NATS 资源详情 ==="
	@kubectl describe statefulset/nats -n aetherius 2>/dev/null || echo "NATS 未部署"
	@echo ""
	@echo "📝 === Redis 资源详情 ==="
	@kubectl describe statefulset/redis -n aetherius 2>/dev/null || echo "Redis 未部署"
	@echo ""
	@echo "📝 === PostgreSQL 资源详情 ==="
	@kubectl describe statefulset/postgres -n aetherius 2>/dev/null || echo "PostgreSQL 未部署"
	@echo ""
	@echo "📝 === Traefik 资源详情 ==="
	@kubectl describe deployment/traefik -n traefik 2>/dev/null || echo "Traefik 未部署"
	@echo ""
	@echo "📝 === Prometheus 资源详情 ==="
	@kubectl describe deployment/prometheus -n monitoring 2>/dev/null || echo "Prometheus 未部署"
	@echo ""
	@echo "📝 === Grafana 资源详情 ==="
	@kubectl describe deployment/grafana -n monitoring 2>/dev/null || echo "Grafana 未部署"

events:
	@echo "📋 === Aetherius 命名空间事件 ==="
	@kubectl get events -n aetherius --sort-by='.lastTimestamp' 2>/dev/null || echo "命名空间 aetherius 不存在"
	@echo ""
	@echo "📋 === Traefik 命名空间事件 ==="
	@kubectl get events -n traefik --sort-by='.lastTimestamp' 2>/dev/null || echo "命名空间 traefik 不存在"
	@echo ""
	@echo "📋 === Monitoring 命名空间事件 ==="
	@kubectl get events -n monitoring --sort-by='.lastTimestamp' 2>/dev/null || echo "命名空间 monitoring 不存在"

# ============================================================================
# 导出生成的 YAML
# ============================================================================

export:
	@echo "📦 导出 Neo4j YAML..."
	@kubectl kustomize neo4j/ > neo4j-generated.yaml
	@echo "📦 导出 NATS YAML..."
	@kubectl kustomize nats/ > nats-generated.yaml
	@echo "📦 导出 Redis YAML..."
	@kubectl kustomize redis/ > redis-generated.yaml
	@echo "📦 导出 PostgreSQL YAML..."
	@kubectl kustomize postgres/ > postgres-generated.yaml
	@echo "📦 导出 Traefik YAML..."
	@kubectl kustomize traefik/ > traefik-generated.yaml
	@echo "📦 导出 Prometheus YAML..."
	@kubectl kustomize prometheus/ > prometheus-generated.yaml
	@echo "📦 导出 Grafana YAML..."
	@kubectl kustomize grafana/ > grafana-generated.yaml
	@echo "✅ YAML 导出完成"

# ============================================================================
# Traefik 域名配置
# ============================================================================

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

# ============================================================================
# 别名
# ============================================================================

up: deploy-all
down: delete-all
ps: status
