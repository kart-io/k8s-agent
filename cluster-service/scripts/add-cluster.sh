#!/bin/bash

# Cluster Service - Add Cluster Script
# 用于添加 Kubernetes 集群到 cluster-service 管理系统

set -e

# 配置
CLUSTER_SERVICE_URL="${CLUSTER_SERVICE_URL:-http://127.0.0.1:8082}"
API_ENDPOINT="${CLUSTER_SERVICE_URL}/api/k8s/clusters"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 打印函数
info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查依赖
check_dependencies() {
    info "检查依赖工具..."

    local missing_tools=()

    if ! command -v kubectl &> /dev/null; then
        missing_tools+=("kubectl")
    fi

    if ! command -v jq &> /dev/null; then
        missing_tools+=("jq")
    fi

    if ! command -v curl &> /dev/null; then
        missing_tools+=("curl")
    fi

    if [ ${#missing_tools[@]} -ne 0 ]; then
        error "缺少以下工具: ${missing_tools[*]}"
        error "请安装后重试"
        exit 1
    fi

    info "所有依赖工具已安装 ✓"
}

# 获取 kubeconfig
get_kubeconfig() {
    local context="$1"

    info "获取 kubeconfig (context: ${context})..."

    if [ -n "$context" ]; then
        kubectl config view --minify --flatten --raw --context="$context" 2>/dev/null
    else
        kubectl config view --minify --flatten --raw 2>/dev/null
    fi

    if [ $? -ne 0 ]; then
        error "无法获取 kubeconfig"
        return 1
    fi
}

# 获取集群信息
get_cluster_info() {
    local context="$1"

    info "获取集群信息..."

    # 获取当前 context
    if [ -z "$context" ]; then
        context=$(kubectl config current-context 2>/dev/null)
    fi

    if [ -z "$context" ]; then
        error "无法确定 Kubernetes context"
        return 1
    fi

    # 获取集群 endpoint
    local endpoint=$(kubectl config view --minify --context="$context" -o jsonpath='{.clusters[0].cluster.server}' 2>/dev/null)

    # 获取集群名称
    local cluster_name=$(kubectl config view --minify --context="$context" -o jsonpath='{.clusters[0].name}' 2>/dev/null)

    # 尝试获取集群版本（需要连接到集群）
    local version=$(kubectl version --short --context="$context" 2>/dev/null | grep 'Server Version' | awk '{print $3}')

    if [ -z "$version" ]; then
        version="unknown"
    fi

    echo "$context|$cluster_name|$endpoint|$version"
}

# 创建 JSON 请求
create_request_json() {
    local name="$1"
    local description="$2"
    local endpoint="$3"
    local provider="$4"
    local region="$5"
    local kubeconfig="$6"

    # 转义 kubeconfig 中的特殊字符
    local escaped_kubeconfig=$(echo "$kubeconfig" | jq -Rs .)

    cat <<EOF
{
  "name": "$name",
  "description": "$description",
  "endpoint": "$endpoint",
  "provider": "$provider",
  "region": "$region",
  "kubeconfig": $escaped_kubeconfig
}
EOF
}

# 添加集群
add_cluster() {
    local json_file="$1"

    info "发送请求到 cluster-service..."
    info "API Endpoint: $API_ENDPOINT"

    local response=$(curl -s -w "\n%{http_code}" -X POST "$API_ENDPOINT" \
        -H "Content-Type: application/json" \
        -d @"$json_file")

    local http_code=$(echo "$response" | tail -n 1)
    local body=$(echo "$response" | sed '$d')

    if [ "$http_code" -eq 200 ]; then
        info "集群添加成功！"
        echo ""
        echo "$body" | jq .

        # 提取集群 ID
        local cluster_id=$(echo "$body" | jq -r '.data.id')
        if [ -n "$cluster_id" ] && [ "$cluster_id" != "null" ]; then
            echo ""
            info "集群 ID: $cluster_id"
            info "测试命令:"
            echo "  # 获取集群详情"
            echo "  curl -s $CLUSTER_SERVICE_URL/api/k8s/clusters/$cluster_id | jq ."
            echo ""
            echo "  # 列出命名空间"
            echo "  curl -s $CLUSTER_SERVICE_URL/api/k8s/clusters/$cluster_id/namespaces | jq ."
            echo ""
            echo "  # 列出 Pods"
            echo "  curl -s $CLUSTER_SERVICE_URL/api/k8s/clusters/$cluster_id/namespaces/default/pods | jq ."
        fi
        return 0
    else
        error "添加集群失败 (HTTP $http_code)"
        echo "$body" | jq . 2>/dev/null || echo "$body"
        return 1
    fi
}

# 交互式添加
interactive_add() {
    info "=== 交互式添加 Kubernetes 集群 ==="
    echo ""

    # 列出可用的 contexts
    info "可用的 Kubernetes contexts:"
    kubectl config get-contexts -o name | nl
    echo ""

    # 选择 context
    read -p "请选择 context (留空使用当前 context): " context_choice

    local context=""
    if [ -n "$context_choice" ]; then
        context=$(kubectl config get-contexts -o name | sed -n "${context_choice}p")
        if [ -z "$context" ]; then
            error "无效的选择"
            return 1
        fi
    else
        context=$(kubectl config current-context)
    fi

    info "使用 context: $context"
    echo ""

    # 获取集群信息
    local cluster_info=$(get_cluster_info "$context")
    IFS='|' read -r ctx_name cluster_name endpoint version <<< "$cluster_info"

    info "检测到的集群信息:"
    echo "  Context: $ctx_name"
    echo "  集群名称: $cluster_name"
    echo "  Endpoint: $endpoint"
    echo "  版本: $version"
    echo ""

    # 输入集群元数据
    read -p "集群显示名称 [$cluster_name]: " display_name
    display_name=${display_name:-$cluster_name}

    read -p "集群描述: " description
    description=${description:-"Kubernetes Cluster"}

    read -p "提供商 (minikube/aws/gcp/azure/self-hosted) [self-hosted]: " provider
    provider=${provider:-self-hosted}

    read -p "区域 [local]: " region
    region=${region:-local}

    # 获取 kubeconfig
    local kubeconfig=$(get_kubeconfig "$context")
    if [ -z "$kubeconfig" ]; then
        error "无法获取 kubeconfig"
        return 1
    fi

    # 创建临时 JSON 文件
    local temp_json=$(mktemp)
    create_request_json "$display_name" "$description" "$endpoint" "$provider" "$region" "$kubeconfig" > "$temp_json"

    echo ""
    info "准备添加集群，请确认信息:"
    echo "  名称: $display_name"
    echo "  描述: $description"
    echo "  Endpoint: $endpoint"
    echo "  提供商: $provider"
    echo "  区域: $region"
    echo ""

    read -p "确认添加？(y/N): " confirm
    if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
        warn "已取消"
        rm -f "$temp_json"
        return 1
    fi

    # 添加集群
    add_cluster "$temp_json"
    local result=$?

    # 清理
    rm -f "$temp_json"

    return $result
}

# 从文件添加
add_from_file() {
    local json_file="$1"

    if [ ! -f "$json_file" ]; then
        error "文件不存在: $json_file"
        return 1
    fi

    info "从文件添加集群: $json_file"

    # 验证 JSON 格式
    if ! jq empty "$json_file" 2>/dev/null; then
        error "无效的 JSON 文件"
        return 1
    fi

    add_cluster "$json_file"
}

# 快速添加 minikube
add_minikube() {
    info "=== 快速添加 Minikube 集群 ==="

    # 检查 minikube 是否运行
    if ! minikube status &>/dev/null; then
        error "Minikube 未运行"
        info "请先启动 minikube: minikube start"
        return 1
    fi

    local cluster_info=$(get_cluster_info "minikube")
    IFS='|' read -r ctx_name cluster_name endpoint version <<< "$cluster_info"

    local kubeconfig=$(get_kubeconfig "minikube")

    local temp_json=$(mktemp)
    create_request_json \
        "minikube-local" \
        "Local Minikube Cluster" \
        "$endpoint" \
        "minikube" \
        "local" \
        "$kubeconfig" > "$temp_json"

    add_cluster "$temp_json"
    local result=$?

    rm -f "$temp_json"
    return $result
}

# 列出已添加的集群
list_clusters() {
    info "获取集群列表..."

    local response=$(curl -s "$API_ENDPOINT")

    echo "$response" | jq .
}

# 显示帮助
show_help() {
    cat <<EOF
Cluster Service - 集群添加工具

用法:
  $0 [选项] [命令]

命令:
  interactive, -i    交互式添加集群（默认）
  minikube, -m       快速添加 Minikube 集群
  file, -f <file>    从 JSON 文件添加集群
  list, -l           列出已添加的集群
  help, -h           显示此帮助信息

选项:
  CLUSTER_SERVICE_URL  cluster-service 地址（默认: http://127.0.0.1:8082）

示例:
  # 交互式添加
  $0 interactive

  # 快速添加 minikube
  $0 minikube

  # 从文件添加
  $0 file /path/to/cluster.json

  # 列出集群
  $0 list

  # 使用自定义服务地址
  CLUSTER_SERVICE_URL=http://192.168.1.100:8082 $0 minikube

JSON 文件格式:
{
  "name": "my-cluster",
  "description": "My Kubernetes Cluster",
  "endpoint": "https://192.168.1.100:6443",
  "provider": "self-hosted",
  "region": "local",
  "kubeconfig": "<kubeconfig-content>"
}
EOF
}

# 主函数
main() {
    local command="${1:-interactive}"

    case "$command" in
        interactive|-i)
            check_dependencies
            interactive_add
            ;;
        minikube|-m)
            check_dependencies
            add_minikube
            ;;
        file|-f)
            check_dependencies
            if [ -z "$2" ]; then
                error "请指定 JSON 文件路径"
                echo "用法: $0 file <json-file>"
                exit 1
            fi
            add_from_file "$2"
            ;;
        list|-l)
            list_clusters
            ;;
        help|-h|--help)
            show_help
            ;;
        *)
            error "未知命令: $command"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# 执行
main "$@"
