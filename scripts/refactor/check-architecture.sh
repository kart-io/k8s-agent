#!/bin/bash
# scripts/refactor/check-architecture.sh
# 检查整个项目的架构一致性

set -e

echo "========================================"
echo "Architecture Consistency Check"
echo "========================================"
echo ""

SERVICES=(agent-manager orchestrator reasoning auth gateway monitor cluster collect-agent)

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_status() {
    local service=$1
    local check=$2
    local status=$3

    if [ "$status" == "PASS" ]; then
        printf "${GREEN}✓${NC} %-20s: %s\n" "$service" "$check"
    elif [ "$status" == "FAIL" ]; then
        printf "${RED}✗${NC} %-20s: %s\n" "$service" "$check"
    else
        printf "${YELLOW}⚠${NC} %-20s: %s\n" "$service" "$check"
    fi
}

# 统计变量
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0

check_service() {
    local service=$1
    local has_options="FAIL"
    local has_initializers="FAIL"
    local uses_application="FAIL"
    local uses_new_logger="FAIL"

    # 检查 options 目录
    if [ -d "cmd/$service/app/options" ]; then
        has_options="PASS"
    fi

    # 检查 initializers 目录
    if [ -d "internal/$service/initializers" ]; then
        has_initializers="PASS"
    fi

    # 检查是否使用 Application 接口
    if [ -f "cmd/$service/app/app.go" ]; then
        if grep -q "commonapp.RunWithRunner" "cmd/$service/app/app.go" 2>/dev/null; then
            uses_application="PASS"
        fi
    fi

    # 检查是否使用新日志系统
    if ! find "internal/$service" -name "*.go" -exec grep -l "common/logger" {} \; 2>/dev/null | grep -q .; then
        uses_new_logger="PASS"
    fi

    # 打印结果
    print_status "$service" "options/" "$has_options"
    print_status "$service" "initializers/" "$has_initializers"
    print_status "$service" "Application interface" "$uses_application"
    print_status "$service" "New logger" "$uses_new_logger"
    echo ""

    # 统计
    for status in "$has_options" "$has_initializers" "$uses_application" "$uses_new_logger"; do
        ((TOTAL_CHECKS++))
        if [ "$status" == "PASS" ]; then
            ((PASSED_CHECKS++))
        else
            ((FAILED_CHECKS++))
        fi
    done
}

# 检查所有服务
for service in "${SERVICES[@]}"; do
    if [ -d "internal/$service" ]; then
        check_service "$service"
    fi
done

# 总结
echo "========================================"
echo "Summary"
echo "========================================"
echo "Total checks: $TOTAL_CHECKS"
printf "${GREEN}Passed: $PASSED_CHECKS${NC}\n"
printf "${RED}Failed: $FAILED_CHECKS${NC}\n"

PASS_RATE=$(awk "BEGIN {printf \"%.1f\", ($PASSED_CHECKS / $TOTAL_CHECKS) * 100}")
echo "Pass rate: $PASS_RATE%"
echo ""

if [ "$PASS_RATE" == "100.0" ]; then
    echo "✓ All services meet architecture standards!"
    exit 0
else
    echo "Services need standardization. See:"
    echo "  docs/CODE_STANDARDIZATION.md"
    echo "  docs/REFACTORING_PLAN.md"
    exit 1
fi
