#!/bin/bash

# Aetherius 性能测试脚本
# 使用 wrk 进行 HTTP 负载测试

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 默认配置
AGENT_MANAGER_URL="${AGENT_MANAGER_URL:-http://localhost:8080}"
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-http://localhost:8081}"
REASONING_URL="${REASONING_URL:-http://localhost:8082}"

DURATION="${DURATION:-30s}"
THREADS="${THREADS:-10}"
CONNECTIONS="${CONNECTIONS:-100}"

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

log_section() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

# 检查依赖
check_dependencies() {
    log_info "检查依赖工具..."

    if ! command -v wrk &> /dev/null; then
        log_error "未找到 wrk 工具"
        echo ""
        echo "安装 wrk:"
        echo "  macOS:  brew install wrk"
        echo "  Ubuntu: sudo apt-get install wrk"
        echo "  从源码: git clone https://github.com/wg/wrk && cd wrk && make"
        exit 1
    fi

    if ! command -v jq &> /dev/null; then
        log_warn "未找到 jq 工具,输出将不会格式化"
    fi

    log_info "✓ 依赖检查完成"
}

# 检查服务可用性
check_services() {
    log_info "检查服务可用性..."

    local services=(
        "$AGENT_MANAGER_URL:Agent Manager"
        "$ORCHESTRATOR_URL:Orchestrator Service"
        "$REASONING_URL:Reasoning Service"
    )

    local all_ok=true

    for service in "${services[@]}"; do
        IFS=':' read -r url name <<< "$service"
        if curl -s -f "${url}/health" > /dev/null 2>&1; then
            log_info "✓ ${name} 可用"
        else
            log_error "✗ ${name} 不可用 (${url})"
            all_ok=false
        fi
    done

    if [ "$all_ok" = false ]; then
        exit 1
    fi

    log_info "✓ 所有服务可用"
}

# 显示测试配置
show_config() {
    log_section "测试配置"
    echo "Agent Manager URL: $AGENT_MANAGER_URL"
    echo "Orchestrator URL: $ORCHESTRATOR_URL"
    echo "Reasoning URL: $REASONING_URL"
    echo ""
    echo "测试时长: $DURATION"
    echo "线程数: $THREADS"
    echo "连接数: $CONNECTIONS"
    echo ""
}

# 运行负载测试
run_load_test() {
    local name=$1
    local url=$2
    local method=${3:-GET}
    local body_file=$4

    log_section "测试: ${name}"
    echo "URL: ${url}"
    echo "Method: ${method}"
    echo "Duration: ${DURATION}"
    echo "Threads: ${THREADS}"
    echo "Connections: ${CONNECTIONS}"
    echo ""

    if [ -n "$body_file" ] && [ -f "$body_file" ]; then
        echo "Body file: ${body_file}"
        echo ""
        wrk -t${THREADS} -c${CONNECTIONS} -d${DURATION} \
            -s <(cat <<EOF
wrk.method = "$method"
wrk.headers["Content-Type"] = "application/json"
wrk.body = io.open("$body_file", "r"):read("*all")
EOF
            ) \
            "$url"
    else
        wrk -t${THREADS} -c${CONNECTIONS} -d${DURATION} "$url"
    fi

    echo ""
}

# Agent Manager 测试
test_agent_manager() {
    log_section "Agent Manager 性能测试"

    # 健康检查
    run_load_test \
        "Health Check" \
        "${AGENT_MANAGER_URL}/health"

    # 列出 Agent
    run_load_test \
        "List Agents" \
        "${AGENT_MANAGER_URL}/api/v1/agents"

    # 查询事件
    run_load_test \
        "Query Events" \
        "${AGENT_MANAGER_URL}/api/v1/events?page=1&page_size=10"

    # 获取单个 Agent (假设 ID 存在)
    run_load_test \
        "Get Agent by ID" \
        "${AGENT_MANAGER_URL}/api/v1/agents/test-agent-001"
}

# Orchestrator Service 测试
test_orchestrator() {
    log_section "Orchestrator Service 性能测试"

    # 健康检查
    run_load_test \
        "Health Check" \
        "${ORCHESTRATOR_URL}/health"

    # 列出工作流
    run_load_test \
        "List Workflows" \
        "${ORCHESTRATOR_URL}/api/v1/workflows"

    # 列出策略
    run_load_test \
        "List Strategies" \
        "${ORCHESTRATOR_URL}/api/v1/strategies"

    # 查询工作流执行
    run_load_test \
        "Query Workflow Executions" \
        "${ORCHESTRATOR_URL}/api/v1/workflows/executions?page=1&page_size=10"
}

# Reasoning Service 测试
test_reasoning() {
    log_section "Reasoning Service 性能测试"

    # 健康检查
    run_load_test \
        "Health Check" \
        "${REASONING_URL}/health"

    # 准备测试数据
    cat > /tmp/root-cause-request.json <<EOF
{
  "request_id": "perf-test-001",
  "analysis_type": "root_cause",
  "context": {
    "event": {
      "reason": "OOMKilled",
      "message": "Container was OOM killed"
    },
    "logs": "fatal error: runtime: out of memory",
    "metrics": {
      "memory": {
        "usage_percent": 98
      }
    }
  }
}
EOF

    # 根因分析
    run_load_test \
        "Root Cause Analysis" \
        "${REASONING_URL}/api/v1/analyze/root-cause" \
        "POST" \
        "/tmp/root-cause-request.json"

    # 清理
    rm -f /tmp/root-cause-request.json
}

# 并发测试
test_concurrent() {
    log_section "并发压力测试"

    log_info "同时向所有服务发送请求..."

    # 启动后台任务
    wrk -t5 -c50 -d${DURATION} "${AGENT_MANAGER_URL}/api/v1/agents" > /tmp/agent-manager-concurrent.log 2>&1 &
    PID1=$!

    wrk -t5 -c50 -d${DURATION} "${ORCHESTRATOR_URL}/api/v1/workflows" > /tmp/orchestrator-concurrent.log 2>&1 &
    PID2=$!

    wrk -t5 -c50 -d${DURATION} "${REASONING_URL}/health" > /tmp/reasoning-concurrent.log 2>&1 &
    PID3=$!

    # 等待所有测试完成
    wait $PID1
    wait $PID2
    wait $PID3

    # 显示结果
    echo ""
    log_info "Agent Manager 并发测试结果:"
    cat /tmp/agent-manager-concurrent.log | grep -E "Requests/sec|Latency"

    echo ""
    log_info "Orchestrator Service 并发测试结果:"
    cat /tmp/orchestrator-concurrent.log | grep -E "Requests/sec|Latency"

    echo ""
    log_info "Reasoning Service 并发测试结果:"
    cat /tmp/reasoning-concurrent.log | grep -E "Requests/sec|Latency"

    # 清理
    rm -f /tmp/*-concurrent.log
}

# 生成性能报告
generate_report() {
    log_section "性能测试报告"

    echo "测试完成时间: $(date)"
    echo ""
    echo "测试配置:"
    echo "  - 时长: $DURATION"
    echo "  - 线程: $THREADS"
    echo "  - 连接: $CONNECTIONS"
    echo ""
    echo "关键指标:"
    echo "  - 请求成功率: 查看上述各项测试的错误率"
    echo "  - 平均响应时间: 查看各项测试的 Latency"
    echo "  - 吞吐量: 查看各项测试的 Requests/sec"
    echo ""
    echo "建议:"
    echo "  1. 如果错误率 > 1%, 检查服务资源配置"
    echo "  2. 如果平均延迟 > 100ms, 考虑优化数据库查询"
    echo "  3. 如果吞吐量不达预期, 增加服务副本数或优化代码"
    echo ""
}

# 主函数
main() {
    echo "========================================="
    echo "Aetherius Performance Testing Tool"
    echo "========================================="
    echo ""

    # 解析参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            --duration)
                DURATION="$2"
                shift 2
                ;;
            --threads)
                THREADS="$2"
                shift 2
                ;;
            --connections)
                CONNECTIONS="$2"
                shift 2
                ;;
            --agent-manager)
                test_type="agent-manager"
                shift
                ;;
            --orchestrator)
                test_type="orchestrator"
                shift
                ;;
            --reasoning)
                test_type="reasoning"
                shift
                ;;
            --concurrent)
                test_type="concurrent"
                shift
                ;;
            --all)
                test_type="all"
                shift
                ;;
            --help)
                cat <<EOF
用法: $0 [选项] [测试类型]

选项:
  --duration <时长>        测试时长 (默认: 30s)
  --threads <数量>         线程数 (默认: 10)
  --connections <数量>     并发连接数 (默认: 100)

测试类型:
  --agent-manager         仅测试 Agent Manager
  --orchestrator          仅测试 Orchestrator Service
  --reasoning             仅测试 Reasoning Service
  --concurrent            并发压力测试
  --all                   运行所有测试 (默认)

环境变量:
  AGENT_MANAGER_URL       Agent Manager URL (默认: http://localhost:8080)
  ORCHESTRATOR_URL        Orchestrator URL (默认: http://localhost:8081)
  REASONING_URL           Reasoning URL (默认: http://localhost:8082)

示例:
  # 运行所有测试,持续 1 分钟
  $0 --duration 60s --threads 20 --connections 200

  # 仅测试 Agent Manager
  $0 --agent-manager

  # 并发压力测试
  $0 --concurrent --duration 120s
EOF
                exit 0
                ;;
            *)
                log_error "未知参数: $1"
                echo "使用 --help 查看帮助"
                exit 1
                ;;
        esac
    done

    # 默认运行所有测试
    test_type="${test_type:-all}"

    # 检查依赖
    check_dependencies

    # 检查服务
    check_services

    # 显示配置
    show_config

    # 运行测试
    case $test_type in
        agent-manager)
            test_agent_manager
            ;;
        orchestrator)
            test_orchestrator
            ;;
        reasoning)
            test_reasoning
            ;;
        concurrent)
            test_concurrent
            ;;
        all)
            test_agent_manager
            test_orchestrator
            test_reasoning
            test_concurrent
            ;;
    esac

    # 生成报告
    generate_report

    log_info "性能测试完成!"
}

# 运行主函数
main "$@"