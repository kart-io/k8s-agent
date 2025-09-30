#!/bin/bash

# Aetherius API 测试脚本
# 用于测试所有核心 API 端点

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# API 地址
AGENT_MANAGER_URL="${AGENT_MANAGER_URL:-http://localhost:8080}"
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-http://localhost:8081}"
REASONING_URL="${REASONING_URL:-http://localhost:8082}"

# 测试计数
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

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

# 测试函数
test_endpoint() {
    local name=$1
    local url=$2
    local method=${3:-GET}
    local data=$4

    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo ""
    echo "========================================="
    echo "Test #${TOTAL_TESTS}: ${name}"
    echo "========================================="
    echo "URL: ${url}"
    echo "Method: ${method}"

    if [ -n "$data" ]; then
        echo "Data: ${data}"
    fi

    # 执行请求
    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X ${method} \
            -H "Content-Type: application/json" \
            -d "${data}" \
            "${url}" 2>/dev/null)
    else
        response=$(curl -s -w "\n%{http_code}" -X ${method} "${url}" 2>/dev/null)
    fi

    # 提取状态码和响应体
    http_code=$(echo "$response" | tail -n 1)
    body=$(echo "$response" | sed '$d')

    echo "HTTP Status: ${http_code}"
    echo "Response:"
    echo "$body" | jq '.' 2>/dev/null || echo "$body"

    # 检查状态码
    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        log_info "✓ PASSED"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        log_error "✗ FAILED (HTTP ${http_code})"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

# 开始测试
echo "========================================="
echo "Aetherius API Test Suite"
echo "========================================="
echo ""
echo "Agent Manager URL: ${AGENT_MANAGER_URL}"
echo "Orchestrator URL: ${ORCHESTRATOR_URL}"
echo "Reasoning URL: ${REASONING_URL}"
echo ""

# ====================================
# Agent Manager API 测试
# ====================================

log_info "Testing Agent Manager API..."

# 健康检查
test_endpoint \
    "Agent Manager - Health Check" \
    "${AGENT_MANAGER_URL}/health" \
    "GET"

# 列出所有 Agent
test_endpoint \
    "Agent Manager - List Agents" \
    "${AGENT_MANAGER_URL}/api/v1/agents" \
    "GET"

# 列出所有集群
test_endpoint \
    "Agent Manager - List Clusters" \
    "${AGENT_MANAGER_URL}/api/v1/clusters" \
    "GET"

# 查询事件
test_endpoint \
    "Agent Manager - Query Events" \
    "${AGENT_MANAGER_URL}/api/v1/events?severity=high&page=1&page_size=10" \
    "GET"

# 高级事件查询
test_endpoint \
    "Agent Manager - Advanced Event Query" \
    "${AGENT_MANAGER_URL}/api/v1/events/query" \
    "POST" \
    '{
        "filters": {
            "severities": ["high", "critical"],
            "time_range": {
                "start": "2025-09-30T00:00:00Z",
                "end": "2025-09-30T23:59:59Z"
            }
        },
        "pagination": {
            "page": 1,
            "page_size": 20
        }
    }'

# 发送命令 (可能会失败,因为没有真实的 Agent)
test_endpoint \
    "Agent Manager - Send Command" \
    "${AGENT_MANAGER_URL}/api/v1/commands" \
    "POST" \
    '{
        "cluster_id": "test-cluster",
        "type": "diagnostic",
        "tool": "kubectl",
        "action": "get",
        "args": ["pods", "-n", "default"],
        "timeout": "30s"
    }'

# ====================================
# Orchestrator Service API 测试
# ====================================

log_info "Testing Orchestrator Service API..."

# 健康检查
test_endpoint \
    "Orchestrator - Health Check" \
    "${ORCHESTRATOR_URL}/health" \
    "GET"

# 列出工作流
test_endpoint \
    "Orchestrator - List Workflows" \
    "${ORCHESTRATOR_URL}/api/v1/workflows" \
    "GET"

# 列出策略
test_endpoint \
    "Orchestrator - List Strategies" \
    "${ORCHESTRATOR_URL}/api/v1/strategies" \
    "GET"

# 查询工作流执行历史
test_endpoint \
    "Orchestrator - List Workflow Executions" \
    "${ORCHESTRATOR_URL}/api/v1/workflows/executions?status=completed&page=1&page_size=10" \
    "GET"

# ====================================
# Reasoning Service API 测试
# ====================================

log_info "Testing Reasoning Service API..."

# 健康检查
test_endpoint \
    "Reasoning - Health Check" \
    "${REASONING_URL}/health" \
    "GET"

# 根因分析
test_endpoint \
    "Reasoning - Root Cause Analysis" \
    "${REASONING_URL}/api/v1/analyze/root-cause" \
    "POST" \
    '{
        "request_id": "test-req-001",
        "analysis_type": "root_cause",
        "context": {
            "event": {
                "reason": "OOMKilled",
                "message": "Container was OOM killed"
            },
            "logs": "fatal error: runtime: out of memory\\nruntime stack:\\nruntime.throw(0x1234567)",
            "metrics": {
                "memory": {
                    "usage_percent": 98
                }
            }
        },
        "options": {
            "min_confidence": 0.7,
            "max_recommendations": 5
        }
    }'

# 故障预测
test_endpoint \
    "Reasoning - Failure Prediction" \
    "${REASONING_URL}/api/v1/analyze/predict" \
    "POST" \
    '{
        "cluster_id": "test-cluster",
        "resource_type": "pod",
        "resource_name": "test-pod",
        "metrics": {
            "memory": {
                "usage_percent": 85
            },
            "cpu": {
                "usage_percent": 75,
                "throttling_percent": 60
            },
            "restart_count": 3
        },
        "time_window": "24h"
    }'

# 查找相似案例
test_endpoint \
    "Reasoning - Find Similar Cases" \
    "${REASONING_URL}/api/v1/cases/similar?event_reason=OOMKilled&limit=5" \
    "GET"

# 获取准确率指标
test_endpoint \
    "Reasoning - Get Accuracy Metrics" \
    "${REASONING_URL}/api/v1/metrics/accuracy" \
    "GET"

# 获取知识图谱统计
test_endpoint \
    "Reasoning - Get Knowledge Stats" \
    "${REASONING_URL}/api/v1/knowledge/stats" \
    "GET"

# ====================================
# 测试总结
# ====================================

echo ""
echo "========================================="
echo "Test Summary"
echo "========================================="
echo "Total Tests: ${TOTAL_TESTS}"
echo -e "Passed: ${GREEN}${PASSED_TESTS}${NC}"
echo -e "Failed: ${RED}${FAILED_TESTS}${NC}"
echo ""

if [ $FAILED_TESTS -eq 0 ]; then
    log_info "All tests passed! ✓"
    exit 0
else
    log_error "Some tests failed! ✗"
    exit 1
fi