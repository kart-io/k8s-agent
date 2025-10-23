#!/bin/bash

# Aetherius 快速启动脚本
# 自动检测环境并启动所有服务

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# 显示欢迎信息
show_banner() {
    cat << 'EOF'
    _         _   _                _
   / \   ___ | |_| |__   ___ _ __ (_)_   _ ___
  / _ \ / _ \| __| '_ \ / _ \ '__| | | | / __|
 / ___ \  __/| |_| | | |  __/ |  | | |_| \__ \
/_/   \_\___| \__|_| |_|\___|_|  |_|\__,_|___/

智能 Kubernetes 运维平台
EOF
    echo ""
    echo "版本: v1.0.0"
    echo "作者: Aetherius Team"
    echo ""
}

# 检查依赖
check_dependencies() {
    log_step "检查系统依赖..."

    local missing_deps=()

    # 检查 Docker
    if ! command -v docker &> /dev/null; then
        missing_deps+=("docker")
    fi

    # 检查 Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        missing_deps+=("docker-compose")
    fi

    # 检查 kubectl (可选)
    if ! command -v kubectl &> /dev/null; then
        log_warn "kubectl 未安装 (仅在部署到 K8s 时需要)"
    fi

    if [ ${#missing_deps[@]} -gt 0 ]; then
        log_error "缺少必要的依赖: ${missing_deps[*]}"
        echo ""
        echo "请安装以下依赖:"
        echo "  - Docker: https://docs.docker.com/get-docker/"
        echo "  - Docker Compose: https://docs.docker.com/compose/install/"
        exit 1
    fi

    log_info "✓ 所有依赖已安装"
}

# 检查 Docker 服务
check_docker_service() {
    log_step "检查 Docker 服务..."

    if ! docker info &> /dev/null; then
        log_error "Docker 服务未运行"
        echo ""
        echo "请启动 Docker 服务:"
        echo "  - macOS/Windows: 启动 Docker Desktop"
        echo "  - Linux: sudo systemctl start docker"
        exit 1
    fi

    log_info "✓ Docker 服务正在运行"
}

# 检查端口占用
check_ports() {
    log_step "检查端口占用..."

    local ports=(5432 6379 4222 7687 8080 8081 8082)
    local occupied_ports=()

    for port in "${ports[@]}"; do
        if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1 || \
           netstat -an 2>/dev/null | grep ":$port " | grep LISTEN >/dev/null 2>&1; then
            occupied_ports+=($port)
        fi
    done

    if [ ${#occupied_ports[@]} -gt 0 ]; then
        log_warn "以下端口已被占用: ${occupied_ports[*]}"
        echo ""
        read -p "是否继续? (y/n) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    else
        log_info "✓ 所有必需端口可用"
    fi
}

# 选择启动模式
select_mode() {
    log_step "选择启动模式..."
    echo ""
    echo "1) Docker Compose (推荐用于开发和测试)"
    echo "2) Kubernetes (推荐用于生产环境)"
    echo "3) 本地开发模式 (仅启动依赖服务)"
    echo ""
    read -p "请选择 (1-3): " mode

    case $mode in
        1)
            start_docker_compose
            ;;
        2)
            start_kubernetes
            ;;
        3)
            start_local_dev
            ;;
        *)
            log_error "无效选择"
            exit 1
            ;;
    esac
}

# Docker Compose 模式
start_docker_compose() {
    log_step "使用 Docker Compose 启动..."

    cd deployments/docker-compose

    # 检查配置文件
    if [ ! -f docker-compose.yml ]; then
        log_error "找不到 docker-compose.yml"
        exit 1
    fi

    # 拉取镜像
    log_info "拉取 Docker 镜像..."
    docker-compose pull || true

    # 启动服务
    log_info "启动服务..."
    docker-compose up -d

    # 等待服务启动
    log_info "等待服务启动..."
    sleep 10

    # 检查服务状态
    check_services_docker

    # 显示访问信息
    show_access_info_docker
}

# Kubernetes 模式
start_kubernetes() {
    log_step "部署到 Kubernetes..."

    # 检查 kubectl
    if ! command -v kubectl &> /dev/null; then
        log_error "kubectl 未安装"
        exit 1
    fi

    # 检查集群连接
    if ! kubectl cluster-info &> /dev/null; then
        log_error "无法连接到 Kubernetes 集群"
        exit 1
    fi

    cd deployments/k8s

    # 创建命名空间
    log_info "创建命名空间..."
    kubectl apply -f namespace.yaml

    # 部署依赖
    log_info "部署依赖服务..."
    kubectl apply -f dependencies.yaml

    # 等待依赖就绪
    log_info "等待依赖服务就绪 (这可能需要几分钟)..."
    kubectl -n aetherius wait --for=condition=ready pod -l app=postgres --timeout=300s || true
    kubectl -n aetherius wait --for=condition=ready pod -l app=redis --timeout=300s || true
    kubectl -n aetherius wait --for=condition=ready pod -l app=nats --timeout=300s || true

    # 部署应用服务
    log_info "部署应用服务..."
    kubectl apply -f agent-manager.yaml
    kubectl apply -f orchestrator-service.yaml
    kubectl apply -f reasoning-service.yaml

    # 等待应用就绪
    log_info "等待应用服务就绪..."
    sleep 20

    # 检查服务状态
    check_services_k8s

    # 显示访问信息
    show_access_info_k8s
}

# 本地开发模式
start_local_dev() {
    log_step "启动本地开发环境..."

    cd deployments/docker-compose

    # 只启动依赖服务
    log_info "启动依赖服务 (PostgreSQL, Redis, NATS, Neo4j)..."
    docker-compose up -d postgres redis nats neo4j

    # 等待服务启动
    log_info "等待服务启动..."
    sleep 10

    # 检查服务状态
    docker-compose ps

    echo ""
    log_info "✓ 依赖服务已启动"
    echo ""
    echo "现在可以在各个终端中运行应用服务:"
    echo ""
    echo "  终端 1 - Agent Manager:"
    echo "    cd agent-manager && make run"
    echo ""
    echo "  终端 2 - Orchestrator Service:"
    echo "    cd orchestrator-service && make run"
    echo ""
    echo "  终端 3 - Reasoning Service:"
    echo "    cd reasoning-service && make run"
    echo ""
    echo "  终端 4 - Collect Agent (可选):"
    echo "    cd collect-agent && make run"
    echo ""
}

# 检查 Docker Compose 服务状态
check_services_docker() {
    log_step "检查服务状态..."

    cd deployments/docker-compose

    local services=(
        "aetherius-postgres:PostgreSQL"
        "aetherius-redis:Redis"
        "aetherius-nats:NATS"
        "aetherius-neo4j:Neo4j"
        "aetherius-agent-manager:Agent Manager"
        "aetherius-orchestrator:Orchestrator"
        "aetherius-reasoning:Reasoning Service"
    )

    echo ""
    for service in "${services[@]}"; do
        IFS=':' read -r container name <<< "$service"
        if docker ps --filter "name=$container" --filter "status=running" | grep -q "$container"; then
            log_info "✓ $name 运行中"
        else
            log_warn "✗ $name 未运行"
        fi
    done
    echo ""
}

# 检查 Kubernetes 服务状态
check_services_k8s() {
    log_step "检查服务状态..."

    echo ""
    kubectl -n aetherius get pods
    echo ""

    local ready_pods=$(kubectl -n aetherius get pods --no-headers | grep "Running" | wc -l)
    log_info "$ready_pods 个 Pod 正在运行"
}

# 显示 Docker Compose 访问信息
show_access_info_docker() {
    cat << 'EOF'

╔═══════════════════════════════════════════════════════════════╗
║                     🎉 启动成功！                              ║
╚═══════════════════════════════════════════════════════════════╝

服务访问地址:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  📊 Agent Manager API:
     http://localhost:8080
     健康检查: curl http://localhost:8080/health

  🔄 Orchestrator Service API:
     http://localhost:8081
     健康检查: curl http://localhost:8081/health

  🤖 Reasoning Service API:
     http://localhost:8082
     健康检查: curl http://localhost:8082/health

  🗄️  Neo4j Browser:
     http://localhost:7474
     用户名: neo4j
     密码: neo4j_pass

  📡 NATS Monitoring:
     http://localhost:8222

数据库连接:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  PostgreSQL: localhost:5432
    用户名: aetherius
    密码: aetherius_pass

  Redis: localhost:6379
    密码: redis_pass

常用命令:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  查看日志:
    docker-compose logs -f

  停止服务:
    docker-compose down

  重启服务:
    docker-compose restart

  查看状态:
    docker-compose ps

测试 API:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  运行测试脚本:
    ./examples/scripts/test-api.sh

下一步:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  1. 查看文档: docs/
  2. 部署 Collect Agent 到 K8s 集群
  3. 配置工作流: examples/workflows/
  4. 访问 Neo4j Browser 查看知识图谱

EOF
}

# 显示 Kubernetes 访问信息
show_access_info_k8s() {
    cat << 'EOF'

╔═══════════════════════════════════════════════════════════════╗
║                  🎉 部署成功！                                 ║
╚═══════════════════════════════════════════════════════════════╝

访问服务:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  使用 port-forward 访问:

    Agent Manager:
      kubectl -n aetherius port-forward svc/agent-manager 8080:8080
      访问: http://localhost:8080

    Orchestrator Service:
      kubectl -n aetherius port-forward svc/orchestrator-service 8081:8081
      访问: http://localhost:8081

    Reasoning Service:
      kubectl -n aetherius port-forward svc/reasoning-service 8082:8082
      访问: http://localhost:8082

常用命令:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  查看 Pod:
    kubectl -n aetherius get pods

  查看日志:
    kubectl -n aetherius logs -l app=agent-manager -f

  查看服务:
    kubectl -n aetherius get svc

  删除部署:
    kubectl delete namespace aetherius

下一步:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  1. 部署 Collect Agent 到目标集群
  2. 配置 Ingress 用于外部访问
  3. 配置监控和告警
  4. 查看文档: docs/

EOF
}

# 主函数
main() {
    show_banner
    check_dependencies
    check_docker_service
    check_ports
    select_mode
}

# 运行主函数
main "$@"