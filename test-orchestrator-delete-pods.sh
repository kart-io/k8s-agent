#!/bin/bash

echo "=========================================="
echo "测试 Orchestrator Service - Pod 删除场景"
echo "=========================================="
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 要删除的 Pods
PODS=(
    "final-correct-test"
    "final-test"
    "test-command-correlation"
    "test-correlation-final"
    "test-event-correlation"
    "test-final-verification"
)

# 检查 orchestrator-service 是否运行
echo -e "${BLUE}步骤 1: 检查 orchestrator-service 状态${NC}"
if pgrep -f "orchestrator-service" > /dev/null || pgrep -f "go run.*orchestrator" > /dev/null; then
    echo -e "${GREEN}✓ orchestrator-service 正在运行${NC}"
else
    echo -e "${RED}✗ orchestrator-service 未运行${NC}"
    echo -e "${YELLOW}请先启动: cd orchestrator-service && make run${NC}"
    exit 1
fi
echo ""

# 检查 NATS 是否可访问
echo -e "${BLUE}步骤 2: 检查 NATS 连接${NC}"
if kubectl get pod -n aetherius-dev | grep nats | grep -q Running; then
    echo -e "${GREEN}✓ NATS 在 K8s 中运行${NC}"
else
    echo -e "${RED}✗ NATS 未运行${NC}"
    exit 1
fi
echo ""

# 检查数据库策略
echo -e "${BLUE}步骤 3: 检查数据库中的策略${NC}"
STRATEGY_COUNT=$(kubectl exec -n aetherius-dev deployment/postgres -- \
    psql -U postgres -d aetherius_orchestrator -t -c \
    "SELECT COUNT(*) FROM strategies WHERE enabled = true;" 2>/dev/null | tr -d ' ')

if [ -z "$STRATEGY_COUNT" ] || [ "$STRATEGY_COUNT" = "0" ]; then
    echo -e "${YELLOW}⚠ 数据库中没有活跃策略${NC}"
    echo -e "${YELLOW}运行以下命令创建测试策略:${NC}"
    echo "  cd orchestrator-service && ./scripts/test-strategy.sh"
    echo ""
    read -p "是否继续？(y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
else
    echo -e "${GREEN}✓ 找到 $STRATEGY_COUNT 个活跃策略${NC}"
fi
echo ""

# 显示当前 Pod 状态
echo -e "${BLUE}步骤 4: 当前 Pod 状态${NC}"
kubectl get pods -n default | grep -E "$(IFS='|'; echo "${PODS[*]}")" || echo "没有找到目标 Pods"
echo ""

# 确认删除
echo -e "${YELLOW}即将删除以下 Pods:${NC}"
for pod in "${PODS[@]}"; do
    if kubectl get pod "$pod" -n default &>/dev/null; then
        echo "  - $pod"
    fi
done
echo ""
read -p "确认删除？(y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "已取消"
    exit 0
fi
echo ""

# 创建日志文件
LOG_FILE="/tmp/orchestrator-test-$(date +%Y%m%d-%H%M%S).log"
echo -e "${BLUE}步骤 5: 开始监控和删除${NC}"
echo "日志文件: $LOG_FILE"
echo ""

# 启动后台监控
echo "开始监控 orchestrator-service 日志..."
echo "----------------------------------------"

# 删除 Pods
DELETED_COUNT=0
for pod in "${PODS[@]}"; do
    if kubectl get pod "$pod" -n default &>/dev/null; then
        echo -e "${YELLOW}删除 Pod: $pod${NC}"
        kubectl delete pod "$pod" -n default --grace-period=0 --force &
        DELETED_COUNT=$((DELETED_COUNT + 1))
        sleep 1
    else
        echo -e "${RED}Pod 不存在: $pod${NC}"
    fi
done

echo ""
echo -e "${GREEN}已触发 $DELETED_COUNT 个 Pod 删除操作${NC}"
echo ""

# 等待事件传播
echo "等待事件传播 (5秒)..."
sleep 5
echo ""

# 检查数据库中的执行记录
echo -e "${BLUE}步骤 6: 检查工作流执行记录${NC}"
kubectl exec -n aetherius-dev deployment/postgres -- \
    psql -U postgres -d aetherius_orchestrator -c \
    "SELECT id, workflow_id, status, started_at
     FROM workflow_executions
     ORDER BY started_at DESC
     LIMIT 10;" 2>/dev/null || echo "查询失败"
echo ""

# 检查事件记录
echo -e "${BLUE}步骤 7: 检查事件记录${NC}"
kubectl exec -n aetherius-dev deployment/postgres -- \
    psql -U postgres -d aetherius -c \
    "SELECT id, type, severity, created_at
     FROM events
     WHERE created_at > NOW() - INTERVAL '5 minutes'
     ORDER BY created_at DESC
     LIMIT 10;" 2>/dev/null || echo "查询失败或表不存在"
echo ""

# 显示 Pod 当前状态
echo -e "${BLUE}步骤 8: Pod 当前状态${NC}"
kubectl get pods -n default
echo ""

echo "=========================================="
echo "测试完成"
echo "=========================================="
echo ""
echo "查看 orchestrator-service 的详细日志以确认事件处理:"
echo "  1. 查找接收到的消息: grep '📨'"
echo "  2. 查找策略匹配: grep '🔍'"
echo "  3. 查找工作流执行: grep '🚀'"
echo ""
echo "如果没有看到事件处理日志，可能的原因:"
echo "  1. collect-agent 未运行或未捕获事件"
echo "  2. NATS 连接问题"
echo "  3. 事件类型不匹配策略条件"
echo "  4. 数据库中没有配置策略"
echo ""
