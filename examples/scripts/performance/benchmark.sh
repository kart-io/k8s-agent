#!/bin/bash

# Aetherius 性能基准测试脚本
# 测试各个服务的关键性能指标

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 配置
AGENT_MANAGER_URL="${AGENT_MANAGER_URL:-http://localhost:8080}"
ORCHESTRATOR_URL="${ORCHESTRATOR_URL:-http://localhost:8081}"
REASONING_URL="${REASONING_URL:-http://localhost:8082}"

SAMPLES="${SAMPLES:-100}"
OUTPUT_DIR="${OUTPUT_DIR:-benchmark-results}"

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_section() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

# 创建输出目录
mkdir -p "$OUTPUT_DIR"

# 测量 API 响应时间
measure_response_time() {
    local name=$1
    local url=$2
    local method=${3:-GET}
    local data=$4

    log_info "测试: $name"

    local total=0
    local min=999999
    local max=0
    local success=0
    local failed=0

    local results_file="${OUTPUT_DIR}/${name// /_}.csv"
    echo "sample,time_ms,status" > "$results_file"

    for i in $(seq 1 $SAMPLES); do
        if [ -n "$data" ]; then
            result=$(curl -s -w "\n%{http_code}\n%{time_total}" -o /dev/null \
                -X "$method" \
                -H "Content-Type: application/json" \
                -d "$data" \
                "$url" 2>/dev/null)
        else
            result=$(curl -s -w "\n%{http_code}\n%{time_total}" -o /dev/null \
                -X "$method" \
                "$url" 2>/dev/null)
        fi

        status=$(echo "$result" | sed -n '1p')
        time_total=$(echo "$result" | sed -n '2p')
        time_ms=$(echo "$time_total * 1000" | bc -l | xargs printf "%.0f")

        echo "$i,$time_ms,$status" >> "$results_file"

        if [ "$status" = "200" ] || [ "$status" = "201" ]; then
            success=$((success + 1))
            total=$((total + time_ms))

            if [ $time_ms -lt $min ]; then
                min=$time_ms
            fi

            if [ $time_ms -gt $max ]; then
                max=$time_ms
            fi
        else
            failed=$((failed + 1))
        fi

        printf "\r  进度: %d/%d (成功: %d, 失败: %d)" $i $SAMPLES $success $failed
    done

    echo ""

    if [ $success -gt 0 ]; then
        local avg=$((total / success))
        echo "  结果:"
        echo "    成功率: $(echo "scale=2; $success * 100 / $SAMPLES" | bc)%"
        echo "    平均响应时间: ${avg}ms"
        echo "    最小响应时间: ${min}ms"
        echo "    最大响应时间: ${max}ms"
    else
        echo "  结果: 所有请求失败"
    fi

    echo ""
}

# Agent Manager 基准测试
benchmark_agent_manager() {
    log_section "Agent Manager 性能基准"

    # 健康检查
    measure_response_time \
        "Agent Manager Health Check" \
        "${AGENT_MANAGER_URL}/health"

    # 列出 Agents
    measure_response_time \
        "Agent Manager List Agents" \
        "${AGENT_MANAGER_URL}/api/v1/agents"

    # 列出 Clusters
    measure_response_time \
        "Agent Manager List Clusters" \
        "${AGENT_MANAGER_URL}/api/v1/clusters"

    # 查询事件
    measure_response_time \
        "Agent Manager Query Events" \
        "${AGENT_MANAGER_URL}/api/v1/events?page=1&page_size=10"
}

# Orchestrator Service 基准测试
benchmark_orchestrator() {
    log_section "Orchestrator Service 性能基准"

    # 健康检查
    measure_response_time \
        "Orchestrator Health Check" \
        "${ORCHESTRATOR_URL}/health"

    # 列出工作流
    measure_response_time \
        "Orchestrator List Workflows" \
        "${ORCHESTRATOR_URL}/api/v1/workflows"

    # 列出策略
    measure_response_time \
        "Orchestrator List Strategies" \
        "${ORCHESTRATOR_URL}/api/v1/strategies"

    # 查询工作流执行
    measure_response_time \
        "Orchestrator Query Executions" \
        "${ORCHESTRATOR_URL}/api/v1/workflows/executions?page=1&page_size=10"
}

# Reasoning Service 基准测试
benchmark_reasoning() {
    log_section "Reasoning Service 性能基准"

    # 健康检查
    measure_response_time \
        "Reasoning Health Check" \
        "${REASONING_URL}/health"

    # 准备根因分析请求
    local root_cause_data='{"request_id":"bench-001","analysis_type":"root_cause","context":{"event":{"reason":"OOMKilled","message":"Container was OOM killed"},"logs":"fatal error: runtime: out of memory","metrics":{"memory":{"usage_percent":98}}}}'

    # 根因分析
    measure_response_time \
        "Reasoning Root Cause Analysis" \
        "${REASONING_URL}/api/v1/analyze/root-cause" \
        "POST" \
        "$root_cause_data"

    # 准备故障预测请求
    local prediction_data='{"cluster_id":"test-cluster","resource_type":"pod","resource_name":"test-pod","metrics":{"memory":{"usage_percent":85},"cpu":{"usage_percent":75}},"time_window":"24h"}'

    # 故障预测
    measure_response_time \
        "Reasoning Failure Prediction" \
        "${REASONING_URL}/api/v1/analyze/predict" \
        "POST" \
        "$prediction_data"

    # 查询相似案例
    measure_response_time \
        "Reasoning Similar Cases" \
        "${REASONING_URL}/api/v1/cases/similar?event_reason=OOMKilled&limit=5"

    # 准确率指标
    measure_response_time \
        "Reasoning Accuracy Metrics" \
        "${REASONING_URL}/api/v1/metrics/accuracy"
}

# 生成 HTML 报告
generate_html_report() {
    log_section "生成性能报告"

    local report_file="${OUTPUT_DIR}/benchmark-report.html"

    cat > "$report_file" <<EOF
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Aetherius 性能基准测试报告</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            margin: 20px;
            background-color: #f5f5f5;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background-color: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            border-bottom: 2px solid #4CAF50;
            padding-bottom: 10px;
        }
        h2 {
            color: #555;
            margin-top: 30px;
        }
        .metric {
            display: inline-block;
            margin: 10px 20px 10px 0;
            padding: 15px;
            background-color: #f9f9f9;
            border-left: 4px solid #4CAF50;
            min-width: 200px;
        }
        .metric-label {
            font-size: 12px;
            color: #777;
            text-transform: uppercase;
        }
        .metric-value {
            font-size: 24px;
            font-weight: bold;
            color: #333;
        }
        .chart {
            margin: 20px 0;
            padding: 20px;
            background-color: #fafafa;
            border-radius: 4px;
        }
        .timestamp {
            color: #999;
            font-size: 14px;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            margin: 20px 0;
        }
        th, td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #ddd;
        }
        th {
            background-color: #4CAF50;
            color: white;
        }
        tr:hover {
            background-color: #f5f5f5;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Aetherius 性能基准测试报告</h1>
        <p class="timestamp">生成时间: $(date '+%Y-%m-%d %H:%M:%S')</p>

        <h2>测试配置</h2>
        <div class="metric">
            <div class="metric-label">样本数</div>
            <div class="metric-value">$SAMPLES</div>
        </div>
        <div class="metric">
            <div class="metric-label">Agent Manager</div>
            <div class="metric-value">$AGENT_MANAGER_URL</div>
        </div>
        <div class="metric">
            <div class="metric-label">Orchestrator</div>
            <div class="metric-value">$ORCHESTRATOR_URL</div>
        </div>
        <div class="metric">
            <div class="metric-label">Reasoning</div>
            <div class="metric-value">$REASONING_URL</div>
        </div>

        <h2>性能摘要</h2>
        <table>
            <thead>
                <tr>
                    <th>测试项</th>
                    <th>平均响应时间</th>
                    <th>最小响应时间</th>
                    <th>最大响应时间</th>
                    <th>成功率</th>
                </tr>
            </thead>
            <tbody>
EOF

    # 处理每个 CSV 文件
    for csv_file in "$OUTPUT_DIR"/*.csv; do
        if [ -f "$csv_file" ]; then
            local name=$(basename "$csv_file" .csv | tr '_' ' ')
            local success=$(awk -F',' '$3 == 200 || $3 == 201' "$csv_file" | wc -l)
            local total=$(tail -n +2 "$csv_file" | wc -l)

            if [ $success -gt 0 ]; then
                local avg=$(awk -F',' 'NR>1 && ($3 == 200 || $3 == 201) {sum+=$2; count++} END {if (count>0) print int(sum/count); else print 0}' "$csv_file")
                local min=$(awk -F',' 'NR>1 && ($3 == 200 || $3 == 201) {if (min=="" || $2<min) min=$2} END {print min}' "$csv_file")
                local max=$(awk -F',' 'NR>1 && ($3 == 200 || $3 == 201) {if ($2>max) max=$2} END {print max}' "$csv_file")
                local success_rate=$(echo "scale=1; $success * 100 / $total" | bc)

                cat >> "$report_file" <<EOF
                <tr>
                    <td>$name</td>
                    <td>${avg}ms</td>
                    <td>${min}ms</td>
                    <td>${max}ms</td>
                    <td>${success_rate}%</td>
                </tr>
EOF
            fi
        fi
    done

    cat >> "$report_file" <<EOF
            </tbody>
        </table>

        <h2>建议</h2>
        <ul>
            <li>平均响应时间应小于 100ms (健康检查和简单查询)</li>
            <li>平均响应时间应小于 500ms (复杂分析如根因分析)</li>
            <li>成功率应达到 99% 以上</li>
            <li>P95 响应时间不应超过平均值的 3 倍</li>
        </ul>

        <h2>详细数据</h2>
        <p>详细的 CSV 数据文件位于: <code>$OUTPUT_DIR/</code></p>
    </div>
</body>
</html>
EOF

    log_info "HTML 报告已生成: $report_file"
}

# 生成 Markdown 报告
generate_markdown_report() {
    local report_file="${OUTPUT_DIR}/benchmark-report.md"

    cat > "$report_file" <<EOF
# Aetherius 性能基准测试报告

**生成时间**: $(date '+%Y-%m-%d %H:%M:%S')

---

## 测试配置

- **样本数**: $SAMPLES
- **Agent Manager URL**: $AGENT_MANAGER_URL
- **Orchestrator URL**: $ORCHESTRATOR_URL
- **Reasoning URL**: $REASONING_URL

---

## 性能摘要

| 测试项 | 平均响应时间 | 最小响应时间 | 最大响应时间 | 成功率 |
|--------|-------------|-------------|-------------|--------|
EOF

    # 处理每个 CSV 文件
    for csv_file in "$OUTPUT_DIR"/*.csv; do
        if [ -f "$csv_file" ]; then
            local name=$(basename "$csv_file" .csv | tr '_' ' ')
            local success=$(awk -F',' '$3 == 200 || $3 == 201' "$csv_file" | wc -l)
            local total=$(tail -n +2 "$csv_file" | wc -l)

            if [ $success -gt 0 ]; then
                local avg=$(awk -F',' 'NR>1 && ($3 == 200 || $3 == 201) {sum+=$2; count++} END {if (count>0) print int(sum/count); else print 0}' "$csv_file")
                local min=$(awk -F',' 'NR>1 && ($3 == 200 || $3 == 201) {if (min=="" || $2<min) min=$2} END {print min}' "$csv_file")
                local max=$(awk -F',' 'NR>1 && ($3 == 200 || $3 == 201) {if ($2>max) max=$2} END {print max}' "$csv_file")
                local success_rate=$(echo "scale=1; $success * 100 / $total" | bc)

                echo "| $name | ${avg}ms | ${min}ms | ${max}ms | ${success_rate}% |" >> "$report_file"
            fi
        fi
    done

    cat >> "$report_file" <<EOF

---

## 建议

- 平均响应时间应小于 100ms (健康检查和简单查询)
- 平均响应时间应小于 500ms (复杂分析如根因分析)
- 成功率应达到 99% 以上
- P95 响应时间不应超过平均值的 3 倍

---

## 详细数据

详细的 CSV 数据文件位于: \`$OUTPUT_DIR/\`
EOF

    log_info "Markdown 报告已生成: $report_file"
}

# 主函数
main() {
    echo "========================================="
    echo "Aetherius Performance Benchmarking Tool"
    echo "========================================="
    echo ""

    # 解析参数
    while [[ $# -gt 0 ]]; do
        case $1 in
            --samples)
                SAMPLES="$2"
                shift 2
                ;;
            --output)
                OUTPUT_DIR="$2"
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
            --all)
                test_type="all"
                shift
                ;;
            --help)
                cat <<EOF
用法: $0 [选项] [测试类型]

选项:
  --samples <数量>        测试样本数 (默认: 100)
  --output <目录>         输出目录 (默认: benchmark-results)

测试类型:
  --agent-manager         仅测试 Agent Manager
  --orchestrator          仅测试 Orchestrator Service
  --reasoning             仅测试 Reasoning Service
  --all                   运行所有测试 (默认)

环境变量:
  AGENT_MANAGER_URL       Agent Manager URL
  ORCHESTRATOR_URL        Orchestrator URL
  REASONING_URL           Reasoning URL

示例:
  # 运行所有基准测试,100 个样本
  $0 --samples 100

  # 仅测试 Reasoning Service,1000 个样本
  $0 --reasoning --samples 1000
EOF
                exit 0
                ;;
            *)
                echo "未知参数: $1"
                exit 1
                ;;
        esac
    done

    # 默认运行所有测试
    test_type="${test_type:-all}"

    log_info "开始性能基准测试..."
    log_info "样本数: $SAMPLES"
    log_info "输出目录: $OUTPUT_DIR"
    echo ""

    # 运行测试
    case $test_type in
        agent-manager)
            benchmark_agent_manager
            ;;
        orchestrator)
            benchmark_orchestrator
            ;;
        reasoning)
            benchmark_reasoning
            ;;
        all)
            benchmark_agent_manager
            benchmark_orchestrator
            benchmark_reasoning
            ;;
    esac

    # 生成报告
    generate_html_report
    generate_markdown_report

    log_info "基准测试完成!"
    log_info "查看报告:"
    log_info "  HTML: $OUTPUT_DIR/benchmark-report.html"
    log_info "  Markdown: $OUTPUT_DIR/benchmark-report.md"
}

# 运行主函数
main "$@"