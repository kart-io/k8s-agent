#!/bin/bash

# Kubernetes Agent - New APIs Test Script
# 测试新增的 DaemonSet, ConfigMap, Secret API

set -e

# 配置
BASE_URL=${BASE_URL:-"http://localhost:8080"}
CLUSTER_ID=${CLUSTER_ID:-"test-cluster"}
NAMESPACE=${NAMESPACE:-"default"}

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 打印函数
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ $1${NC}"
}

print_section() {
    echo ""
    echo "========================================="
    echo "$1"
    echo "========================================="
}

# 测试API函数
test_api() {
    local method=$1
    local url=$2
    local data=$3
    local description=$4

    print_info "Testing: $description"
    echo "  Method: $method"
    echo "  URL: $url"

    if [ -n "$data" ]; then
        response=$(curl -s -X "$method" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$url")
    else
        response=$(curl -s -X "$method" "$url")
    fi

    # 检查响应中是否有 "code":0
    if echo "$response" | grep -q '"code":0'; then
        print_success "$description - SUCCESS"
        echo "  Response: $response" | head -c 200
        echo "..."
        return 0
    else
        print_error "$description - FAILED"
        echo "  Response: $response"
        return 1
    fi
}

# ===========================
# DaemonSet API Tests
# ===========================
print_section "DaemonSet API Tests"

# 1. 获取 DaemonSet 列表
test_api "GET" \
    "$BASE_URL/api/k8s/clusters/$CLUSTER_ID/namespaces/$NAMESPACE/daemonsets" \
    "" \
    "List DaemonSets"

# 2. 获取 DaemonSet 详情 (需要先有一个存在的 DaemonSet)
# print_info "Skipping Get DaemonSet detail (no daemonset created yet)"

# 3. 重启 DaemonSet (需要先有一个存在的 DaemonSet)
# print_info "Skipping Restart DaemonSet (no daemonset created yet)"

# 4. 删除 DaemonSet (需要先有一个存在的 DaemonSet)
# print_info "Skipping Delete DaemonSet (no daemonset created yet)"

# ===========================
# ConfigMap API Tests
# ===========================
print_section "ConfigMap API Tests"

# 1. 创建 ConfigMap
CONFIGMAP_NAME="test-configmap-$(date +%s)"
test_api "POST" \
    "$BASE_URL/api/k8s/clusters/$CLUSTER_ID/namespaces/$NAMESPACE/configmaps" \
    '{
        "name": "'"$CONFIGMAP_NAME"'",
        "namespace": "'"$NAMESPACE"'",
        "data": {
            "key1": "value1",
            "key2": "value2",
            "config.yaml": "app:\n  name: test\n  port: 8080"
        },
        "labels": {
            "app": "test",
            "env": "testing"
        }
    }' \
    "Create ConfigMap"

# 2. 获取 ConfigMap 列表
test_api "GET" \
    "$BASE_URL/api/k8s/clusters/$CLUSTER_ID/namespaces/$NAMESPACE/configmaps" \
    "" \
    "List ConfigMaps"

# 3. 获取 ConfigMap 详情
test_api "GET" \
    "$BASE_URL/api/k8s/clusters/$CLUSTER_ID/namespaces/$NAMESPACE/configmaps/$CONFIGMAP_NAME" \
    "" \
    "Get ConfigMap Detail"

# 4. 更新 ConfigMap
test_api "PUT" \
    "$BASE_URL/api/k8s/clusters/$CLUSTER_ID/namespaces/$NAMESPACE/configmaps/$CONFIGMAP_NAME" \
    '{
        "data": {
            "key1": "updated-value1",
            "key3": "value3"
        },
        "labels": {
            "app": "test",
            "env": "production"
        }
    }' \
    "Update ConfigMap"

# 5. 删除 ConfigMap
test_api "DELETE" \
    "$BASE_URL/api/k8s/clusters/$CLUSTER_ID/namespaces/$NAMESPACE/configmaps/$CONFIGMAP_NAME" \
    "" \
    "Delete ConfigMap"

# ===========================
# Secret API Tests
# ===========================
print_section "Secret API Tests"

# 1. 创建 Secret
SECRET_NAME="test-secret-$(date +%s)"
# Base64 编码: echo -n 'admin' | base64 => YWRtaW4=
# Base64 编码: echo -n 'password123' | base64 => cGFzc3dvcmQxMjM=
test_api "POST" \
    "$BASE_URL/api/k8s/clusters/$CLUSTER_ID/namespaces/$NAMESPACE/secrets" \
    '{
        "name": "'"$SECRET_NAME"'",
        "namespace": "'"$NAMESPACE"'",
        "type": "Opaque",
        "stringData": {
            "username": "admin",
            "password": "password123"
        },
        "labels": {
            "app": "test",
            "type": "credentials"
        }
    }' \
    "Create Secret"

# 2. 获取 Secret 列表
test_api "GET" \
    "$BASE_URL/api/k8s/clusters/$CLUSTER_ID/namespaces/$NAMESPACE/secrets" \
    "" \
    "List Secrets"

# 3. 获取 Secret 详情 (不包含敏感数据)
test_api "GET" \
    "$BASE_URL/api/k8s/clusters/$CLUSTER_ID/namespaces/$NAMESPACE/secrets/$SECRET_NAME" \
    "" \
    "Get Secret Detail (without data)"

# 4. 获取 Secret 详情 (包含敏感数据)
test_api "GET" \
    "$BASE_URL/api/k8s/clusters/$CLUSTER_ID/namespaces/$NAMESPACE/secrets/$SECRET_NAME?includeData=true" \
    "" \
    "Get Secret Detail (with data)"

# 5. 更新 Secret
test_api "PUT" \
    "$BASE_URL/api/k8s/clusters/$CLUSTER_ID/namespaces/$NAMESPACE/secrets/$SECRET_NAME" \
    '{
        "stringData": {
            "username": "admin",
            "password": "newpassword456",
            "api_key": "sk-test-123456"
        },
        "labels": {
            "app": "test",
            "type": "credentials",
            "updated": "true"
        }
    }' \
    "Update Secret"

# 6. 删除 Secret
test_api "DELETE" \
    "$BASE_URL/api/k8s/clusters/$CLUSTER_ID/namespaces/$NAMESPACE/secrets/$SECRET_NAME" \
    "" \
    "Delete Secret"

# ===========================
# 测试总结
# ===========================
print_section "Test Summary"
print_success "All new API tests completed!"
print_info "APIs tested:"
echo "  - DaemonSet: List"
echo "  - ConfigMap: Create, List, Get, Update, Delete"
echo "  - Secret: Create, List, Get, Get(with data), Update, Delete"
echo ""
print_info "Next steps:"
echo "  1. Test with real Kubernetes cluster"
echo "  2. Test DaemonSet operations (restart, delete) with existing resources"
echo "  3. Add integration tests"
echo ""
