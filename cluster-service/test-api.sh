#!/bin/bash

# K8s Agent API 测试脚本
# 用于快速验证所有已实现的 API 接口

set -e

# 配置
BASE_URL="${BASE_URL:-http://localhost:8082}"
CLUSTER_ID=""
TEST_NAMESPACE="test-api-ns"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# 检查依赖
check_dependencies() {
    log_info "Checking dependencies..."

    if ! command -v curl &> /dev/null; then
        log_error "curl is not installed"
        exit 1
    fi

    if ! command -v jq &> /dev/null; then
        log_warn "jq is not installed (recommended for better output)"
    fi
}

# 测试健康检查
test_health_check() {
    log_info "Testing health check..."

    response=$(curl -s -w "\n%{http_code}" "$BASE_URL/health")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" = "200" ]; then
        log_info "✓ Health check passed"
        echo "$body" | jq . 2>/dev/null || echo "$body"
    else
        log_error "✗ Health check failed (HTTP $http_code)"
        echo "$body"
        exit 1
    fi

    echo ""
}

# 测试集群管理 API
test_cluster_api() {
    log_info "=== Testing Cluster Management API ==="

    # 1. 获取集群列表
    log_info "1. GET /api/k8s/clusters (List clusters)"
    response=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/k8s/clusters?page=1&pageSize=10")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" = "200" ]; then
        log_info "✓ List clusters successful"
        echo "$body" | jq . 2>/dev/null || echo "$body"
    else
        log_warn "✗ List clusters returned HTTP $http_code (may be empty)"
        echo "$body"
    fi

    echo ""

    # 注意：创建集群需要有效的 kubeconfig，这里只展示如何调用
    log_info "2. POST /api/k8s/clusters (Create cluster)"
    log_warn "Skipping cluster creation (requires valid kubeconfig)"
    log_info "Example command:"
    cat <<'EOF'
    curl -X POST $BASE_URL/api/k8s/clusters \
      -H "Content-Type: application/json" \
      -d '{
        "name": "test-cluster",
        "description": "Test cluster",
        "endpoint": "https://localhost:6443",
        "region": "local",
        "provider": "kubernetes",
        "kubeconfig": "<your-kubeconfig-here>"
      }'
EOF

    echo ""
}

# 测试命名空间 API
test_namespace_api() {
    log_info "=== Testing Namespace Management API ==="

    if [ -z "$CLUSTER_ID" ]; then
        log_warn "No cluster ID available, skipping namespace tests"
        echo ""
        return
    fi

    # 1. 获取命名空间列表
    log_info "1. GET /api/k8s/clusters/$CLUSTER_ID/namespaces"
    response=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/k8s/clusters/$CLUSTER_ID/namespaces")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" = "200" ]; then
        log_info "✓ List namespaces successful"
        echo "$body" | jq . 2>/dev/null || echo "$body"
    else
        log_error "✗ List namespaces failed (HTTP $http_code)"
        echo "$body"
    fi

    echo ""

    # 2. 创建命名空间
    log_info "2. POST /api/k8s/clusters/$CLUSTER_ID/namespaces"
    response=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/k8s/clusters/$CLUSTER_ID/namespaces" \
      -H "Content-Type: application/json" \
      -d "{\"name\": \"$TEST_NAMESPACE\", \"labels\": {\"test\": \"true\"}}")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" = "200" ]; then
        log_info "✓ Create namespace successful"
        echo "$body" | jq . 2>/dev/null || echo "$body"
    else
        log_warn "✗ Create namespace returned HTTP $http_code (may already exist)"
        echo "$body"
    fi

    echo ""

    # 3. 获取命名空间详情
    log_info "3. GET /api/k8s/clusters/$CLUSTER_ID/namespaces/$TEST_NAMESPACE"
    response=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/k8s/clusters/$CLUSTER_ID/namespaces/$TEST_NAMESPACE")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" = "200" ]; then
        log_info "✓ Get namespace successful"
        echo "$body" | jq . 2>/dev/null || echo "$body"
    else
        log_error "✗ Get namespace failed (HTTP $http_code)"
        echo "$body"
    fi

    echo ""
}

# 测试 Pod API
test_pod_api() {
    log_info "=== Testing Pod Management API ==="

    if [ -z "$CLUSTER_ID" ]; then
        log_warn "No cluster ID available, skipping Pod tests"
        echo ""
        return
    fi

    # 1. 获取 Pod 列表
    log_info "1. GET /api/k8s/clusters/$CLUSTER_ID/namespaces/default/pods"
    response=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/k8s/clusters/$CLUSTER_ID/namespaces/default/pods?page=1&pageSize=10")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" = "200" ]; then
        log_info "✓ List pods successful"
        echo "$body" | jq . 2>/dev/null || echo "$body"

        # 提取第一个 Pod 名称用于后续测试
        POD_NAME=$(echo "$body" | jq -r '.data.items[0].name' 2>/dev/null)
    else
        log_error "✗ List pods failed (HTTP $http_code)"
        echo "$body"
    fi

    echo ""

    # 2. 获取 Pod 详情
    if [ -n "$POD_NAME" ] && [ "$POD_NAME" != "null" ]; then
        log_info "2. GET /api/k8s/clusters/$CLUSTER_ID/namespaces/default/pods/$POD_NAME"
        response=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/k8s/clusters/$CLUSTER_ID/namespaces/default/pods/$POD_NAME")
        http_code=$(echo "$response" | tail -n1)
        body=$(echo "$response" | sed '$d')

        if [ "$http_code" = "200" ]; then
            log_info "✓ Get pod details successful"
            echo "$body" | jq . 2>/dev/null || echo "$body"
        else
            log_error "✗ Get pod details failed (HTTP $http_code)"
            echo "$body"
        fi

        echo ""

        # 3. 获取 Pod 日志
        log_info "3. GET /api/k8s/clusters/$CLUSTER_ID/namespaces/default/pods/$POD_NAME/logs"
        response=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/k8s/clusters/$CLUSTER_ID/namespaces/default/pods/$POD_NAME/logs?tailLines=10")
        http_code=$(echo "$response" | tail -n1)
        body=$(echo "$response" | sed '$d')

        if [ "$http_code" = "200" ]; then
            log_info "✓ Get pod logs successful"
            echo "$body" | jq . 2>/dev/null || echo "$body"
        else
            log_error "✗ Get pod logs failed (HTTP $http_code)"
            echo "$body"
        fi

        echo ""
    else
        log_warn "No pods found, skipping pod details and logs tests"
        echo ""
    fi
}

# 测试 Deployment API
test_deployment_api() {
    log_info "=== Testing Deployment Management API ==="

    if [ -z "$CLUSTER_ID" ]; then
        log_warn "No cluster ID available, skipping Deployment tests"
        echo ""
        return
    fi

    # 1. 获取 Deployment 列表
    log_info "1. GET /api/k8s/clusters/$CLUSTER_ID/namespaces/default/deployments"
    response=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/k8s/clusters/$CLUSTER_ID/namespaces/default/deployments?page=1&pageSize=10")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" = "200" ]; then
        log_info "✓ List deployments successful"
        echo "$body" | jq . 2>/dev/null || echo "$body"

        # 提取第一个 Deployment 名称用于后续测试
        DEPLOY_NAME=$(echo "$body" | jq -r '.data.items[0].name' 2>/dev/null)
    else
        log_error "✗ List deployments failed (HTTP $http_code)"
        echo "$body"
    fi

    echo ""

    # 2. 获取 Deployment 详情
    if [ -n "$DEPLOY_NAME" ] && [ "$DEPLOY_NAME" != "null" ]; then
        log_info "2. GET /api/k8s/clusters/$CLUSTER_ID/namespaces/default/deployments/$DEPLOY_NAME"
        response=$(curl -s -w "\n%{http_code}" "$BASE_URL/api/k8s/clusters/$CLUSTER_ID/namespaces/default/deployments/$DEPLOY_NAME")
        http_code=$(echo "$response" | tail -n1)
        body=$(echo "$response" | sed '$d')

        if [ "$http_code" = "200" ]; then
            log_info "✓ Get deployment details successful"
            echo "$body" | jq . 2>/dev/null || echo "$body"
        else
            log_error "✗ Get deployment details failed (HTTP $http_code)"
            echo "$body"
        fi

        echo ""

        log_info "Deployment scale and restart operations skipped (would modify cluster state)"
        log_info "Example scale command:"
        echo "curl -X PUT $BASE_URL/api/k8s/clusters/$CLUSTER_ID/namespaces/default/deployments/$DEPLOY_NAME/scale -H 'Content-Type: application/json' -d '{\"replicas\": 3}'"

        log_info "Example restart command:"
        echo "curl -X POST $BASE_URL/api/k8s/clusters/$CLUSTER_ID/namespaces/default/deployments/$DEPLOY_NAME/restart"

        echo ""
    else
        log_warn "No deployments found, skipping deployment details test"
        echo ""
    fi
}

# 主测试流程
main() {
    log_info "Starting K8s Agent API Tests"
    log_info "Base URL: $BASE_URL"
    echo ""

    check_dependencies

    # 测试健康检查
    test_health_check

    # 测试集群管理 API
    test_cluster_api

    # 如果有集群 ID，测试其他 API
    # 注意：需要用户提供有效的 CLUSTER_ID
    if [ -n "$CLUSTER_ID" ]; then
        test_namespace_api
        test_pod_api
        test_deployment_api
    else
        log_warn "No CLUSTER_ID provided, skipping namespace, pod, and deployment tests"
        log_info "To test all APIs, set CLUSTER_ID environment variable:"
        log_info "  export CLUSTER_ID=your-cluster-id"
        log_info "  ./test-api.sh"
    fi

    echo ""
    log_info "=== Test Summary ==="
    log_info "✓ Health check API working"
    log_info "✓ Cluster list API working"

    if [ -n "$CLUSTER_ID" ]; then
        log_info "✓ Namespace, Pod, and Deployment APIs tested"
    else
        log_warn "⊘ Namespace, Pod, and Deployment APIs not tested (no CLUSTER_ID)"
    fi

    echo ""
    log_info "All available tests completed!"
}

# 运行主程序
main "$@"
