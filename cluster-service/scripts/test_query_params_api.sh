#!/bin/bash

# K8s Agent API 测试脚本 - 查询参数风格
# 此脚本用于测试所有新的查询参数风格 API 端点

set -e

# 配置
BASE_URL="${BASE_URL:-http://localhost:8080}"
CLUSTER_ID="${CLUSTER_ID:-test-cluster}"
NAMESPACE="${NAMESPACE:-default}"
POD_NAME="${POD_NAME:-test-pod}"
DEPLOYMENT_NAME="${DEPLOYMENT_NAME:-test-deployment}"
NODE_NAME="${NODE_NAME:-test-node}"
SERVICE_NAME="${SERVICE_NAME:-test-service}"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 测试计数器
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((PASSED_TESTS++))
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((FAILED_TESTS++))
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_header() {
    echo ""
    echo -e "${BLUE}================================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}================================================${NC}"
}

# HTTP 请求函数
http_get() {
    local url="$1"
    local description="$2"

    ((TOTAL_TESTS++))
    log_info "Testing: ${description}"
    log_info "URL: ${url}"

    response=$(curl -s -w "\n%{http_code}" "$url")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)

    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        log_success "HTTP ${http_code} - ${description}"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
        return 0
    else
        log_error "HTTP ${http_code} - ${description}"
        echo "$body"
        return 1
    fi
}

http_post() {
    local url="$1"
    local data="$2"
    local description="$3"

    ((TOTAL_TESTS++))
    log_info "Testing: ${description}"
    log_info "URL: ${url}"

    response=$(curl -s -w "\n%{http_code}" -X POST \
        -H "Content-Type: application/json" \
        -d "$data" \
        "$url")

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)

    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        log_success "HTTP ${http_code} - ${description}"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
        return 0
    else
        log_error "HTTP ${http_code} - ${description}"
        echo "$body"
        return 1
    fi
}

http_put() {
    local url="$1"
    local data="$2"
    local description="$3"

    ((TOTAL_TESTS++))
    log_info "Testing: ${description}"
    log_info "URL: ${url}"

    response=$(curl -s -w "\n%{http_code}" -X PUT \
        -H "Content-Type: application/json" \
        -d "$data" \
        "$url")

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)

    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        log_success "HTTP ${http_code} - ${description}"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
        return 0
    else
        log_error "HTTP ${http_code} - ${description}"
        echo "$body"
        return 1
    fi
}

http_delete() {
    local url="$1"
    local description="$2"

    ((TOTAL_TESTS++))
    log_info "Testing: ${description}"
    log_info "URL: ${url}"

    response=$(curl -s -w "\n%{http_code}" -X DELETE "$url")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)

    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        log_success "HTTP ${http_code} - ${description}"
        echo "$body" | jq '.' 2>/dev/null || echo "$body"
        return 0
    else
        log_error "HTTP ${http_code} - ${description}"
        echo "$body"
        return 1
    fi
}

# 检查服务是否运行
check_service() {
    log_header "检查服务状态"

    if curl -s "${BASE_URL}/health" > /dev/null 2>&1; then
        log_success "服务运行正常: ${BASE_URL}"
        return 0
    else
        log_error "服务未运行: ${BASE_URL}"
        log_warn "请先启动 cluster-service"
        exit 1
    fi
}

# 测试集群管理 API
test_cluster_apis() {
    log_header "测试集群管理 API"

    # 列出集群
    http_get "${BASE_URL}/api/k8s/clusters" \
        "列出所有集群"

    # 获取集群选项
    http_get "${BASE_URL}/api/k8s/clusters/options" \
        "获取集群选择器列表"

    # 获取单个集群 (查询参数)
    http_get "${BASE_URL}/api/k8s/cluster?clusterId=${CLUSTER_ID}" \
        "获取集群详情 (查询参数)"

    # 获取集群健康状态 (查询参数)
    http_get "${BASE_URL}/api/k8s/cluster/health?clusterId=${CLUSTER_ID}" \
        "获取集群健康状态 (查询参数)"
}

# 测试命名空间管理 API
test_namespace_apis() {
    log_header "测试命名空间管理 API"

    # 列出命名空间 (查询参数)
    http_get "${BASE_URL}/api/k8s/namespaces?clusterId=${CLUSTER_ID}" \
        "列出命名空间 (查询参数)"

    # 获取单个命名空间 (查询参数)
    http_get "${BASE_URL}/api/k8s/namespace?clusterId=${CLUSTER_ID}&namespace=${NAMESPACE}" \
        "获取命名空间详情 (查询参数)"
}

# 测试 Pod 管理 API
test_pod_apis() {
    log_header "测试 Pod 管理 API"

    # 列出 Pods (查询参数)
    http_get "${BASE_URL}/api/k8s/pods?clusterId=${CLUSTER_ID}&namespace=${NAMESPACE}" \
        "列出 Pods (查询参数)"

    # 获取单个 Pod (查询参数)
    http_get "${BASE_URL}/api/k8s/pod?clusterId=${CLUSTER_ID}&namespace=${NAMESPACE}&name=${POD_NAME}" \
        "获取 Pod 详情 (查询参数)"

    # 获取 Pod 日志 (查询参数)
    http_get "${BASE_URL}/api/k8s/pod/logs?clusterId=${CLUSTER_ID}&namespace=${NAMESPACE}&name=${POD_NAME}&tailLines=100" \
        "获取 Pod 日志 (查询参数)"
}

# 测试 Deployment 管理 API
test_deployment_apis() {
    log_header "测试 Deployment 管理 API"

    # 列出 Deployments (查询参数)
    http_get "${BASE_URL}/api/k8s/deployments?clusterId=${CLUSTER_ID}&namespace=${NAMESPACE}" \
        "列出 Deployments (查询参数)"

    # 获取单个 Deployment (查询参数)
    http_get "${BASE_URL}/api/k8s/deployment?clusterId=${CLUSTER_ID}&namespace=${NAMESPACE}&name=${DEPLOYMENT_NAME}" \
        "获取 Deployment 详情 (查询参数)"
}

# 测试 Node 管理 API
test_node_apis() {
    log_header "测试 Node 管理 API"

    # 列出 Nodes (查询参数)
    http_get "${BASE_URL}/api/k8s/nodes?clusterId=${CLUSTER_ID}" \
        "列出 Nodes (查询参数)"

    # 获取单个 Node (查询参数)
    http_get "${BASE_URL}/api/k8s/node?clusterId=${CLUSTER_ID}&name=${NODE_NAME}" \
        "获取 Node 详情 (查询参数)"
}

# 测试 Service 管理 API
test_service_apis() {
    log_header "测试 Service 管理 API"

    # 列出 Services (查询参数)
    http_get "${BASE_URL}/api/k8s/services?clusterId=${CLUSTER_ID}&namespace=${NAMESPACE}" \
        "列出 Services (查询参数)"

    # 获取单个 Service (查询参数)
    http_get "${BASE_URL}/api/k8s/service?clusterId=${CLUSTER_ID}&namespace=${NAMESPACE}&name=${SERVICE_NAME}" \
        "获取 Service 详情 (查询参数)"
}

# 测试 StatefulSet 管理 API
test_statefulset_apis() {
    log_header "测试 StatefulSet 管理 API"

    # 列出 StatefulSets (查询参数)
    http_get "${BASE_URL}/api/k8s/statefulsets?clusterId=${CLUSTER_ID}&namespace=${NAMESPACE}" \
        "列出 StatefulSets (查询参数)"
}

# 测试 DaemonSet 管理 API
test_daemonset_apis() {
    log_header "测试 DaemonSet 管理 API"

    # 列出 DaemonSets (查询参数)
    http_get "${BASE_URL}/api/k8s/daemonsets?clusterId=${CLUSTER_ID}&namespace=${NAMESPACE}" \
        "列出 DaemonSets (查询参数)"
}

# 测试 ConfigMap 管理 API
test_configmap_apis() {
    log_header "测试 ConfigMap 管理 API"

    # 列出 ConfigMaps (查询参数)
    http_get "${BASE_URL}/api/k8s/configmaps?clusterId=${CLUSTER_ID}&namespace=${NAMESPACE}" \
        "列出 ConfigMaps (查询参数)"
}

# 测试 Secret 管理 API
test_secret_apis() {
    log_header "测试 Secret 管理 API"

    # 列出 Secrets (查询参数)
    http_get "${BASE_URL}/api/k8s/secrets?clusterId=${CLUSTER_ID}&namespace=${NAMESPACE}" \
        "列出 Secrets (查询参数)"
}

# 测试查询参数编码
test_query_encoding() {
    log_header "测试查询参数 URL 编码"

    # 测试包含特殊字符的命名空间
    local special_namespace="kube-system"
    http_get "${BASE_URL}/api/k8s/namespace?clusterId=${CLUSTER_ID}&namespace=${special_namespace}" \
        "测试特殊字符命名空间 (kube-system)"

    # 测试包含特殊字符的 Pod 名称
    local special_pod="coredns-abc-123"
    http_get "${BASE_URL}/api/k8s/pod?clusterId=${CLUSTER_ID}&namespace=${special_namespace}&name=${special_pod}" \
        "测试特殊字符 Pod 名称"
}

# 测试错误处理
test_error_handling() {
    log_header "测试错误处理"

    # 测试缺少必需参数
    log_info "测试缺少 clusterId 参数"
    ((TOTAL_TESTS++))
    response=$(curl -s -w "\n%{http_code}" "${BASE_URL}/api/k8s/cluster")
    http_code=$(echo "$response" | tail -n1)

    if [ "$http_code" -eq 400 ]; then
        log_success "正确返回 400 - 缺少必需参数"
    else
        log_error "期望 400,实际 ${http_code}"
    fi

    # 测试缺少多个参数
    log_info "测试缺少 namespace 参数"
    ((TOTAL_TESTS++))
    response=$(curl -s -w "\n%{http_code}" "${BASE_URL}/api/k8s/pod?clusterId=${CLUSTER_ID}")
    http_code=$(echo "$response" | tail -n1)

    if [ "$http_code" -eq 400 ]; then
        log_success "正确返回 400 - 缺少多个必需参数"
    else
        log_error "期望 400,实际 ${http_code}"
    fi
}

# 测试分页
test_pagination() {
    log_header "测试分页功能"

    # 测试第 1 页
    http_get "${BASE_URL}/api/k8s/clusters?page=1&pageSize=10" \
        "测试集群列表分页 (第 1 页,每页 10 条)"

    # 测试第 2 页
    http_get "${BASE_URL}/api/k8s/clusters?page=2&pageSize=10" \
        "测试集群列表分页 (第 2 页,每页 10 条)"

    # 测试 Pod 列表分页
    http_get "${BASE_URL}/api/k8s/pods?clusterId=${CLUSTER_ID}&namespace=${NAMESPACE}&page=1&pageSize=20" \
        "测试 Pod 列表分页"
}

# 生成测试报告
generate_report() {
    log_header "测试报告"

    echo ""
    echo -e "${BLUE}总测试数:${NC} ${TOTAL_TESTS}"
    echo -e "${GREEN}通过:${NC} ${PASSED_TESTS}"
    echo -e "${RED}失败:${NC} ${FAILED_TESTS}"
    echo ""

    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "${GREEN}✓ 所有测试通过!${NC}"
        return 0
    else
        echo -e "${RED}✗ 有 ${FAILED_TESTS} 个测试失败${NC}"
        return 1
    fi
}

# 主函数
main() {
    log_header "K8s Agent API 测试 - 查询参数风格"

    echo "Base URL: ${BASE_URL}"
    echo "Cluster ID: ${CLUSTER_ID}"
    echo "Namespace: ${NAMESPACE}"
    echo ""

    # 检查依赖
    if ! command -v curl &> /dev/null; then
        log_error "curl 未安装,请先安装 curl"
        exit 1
    fi

    if ! command -v jq &> /dev/null; then
        log_warn "jq 未安装,JSON 输出将不会格式化"
    fi

    # 检查服务
    check_service

    # 执行测试
    test_cluster_apis || true
    test_namespace_apis || true
    test_pod_apis || true
    test_deployment_apis || true
    test_node_apis || true
    test_service_apis || true
    test_statefulset_apis || true
    test_daemonset_apis || true
    test_configmap_apis || true
    test_secret_apis || true
    test_query_encoding || true
    test_error_handling || true
    test_pagination || true

    # 生成报告
    generate_report
}

# 运行主函数
main "$@"
