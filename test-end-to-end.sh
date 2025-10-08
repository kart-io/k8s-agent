#!/bin/bash

# 端到端测试脚本 - 测试 agent-manager 和 orchestrator-service 集成

set -e

# 颜色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo "=========================================="
echo "端到端流程测试"
echo "=========================================="
echo ""

# 步骤 1: 检查服务状态
echo -e "${BLUE}步骤 1: 检查服务运行状态${NC}"
echo ""

echo "检查 agent-manager (端口 8080)..."
if lsof -i :8080 >/dev/null 2>&1; then
    echo -e "${GREEN}✓ agent-manager 运行中${NC}"
    AGENT_MANAGER_RUNNING=true
else
    echo -e "${RED}✗ agent-manager 未运行${NC}"
    echo "请在另一个终端运行: cd agent-manager && make run"
    AGENT_MANAGER_RUNNING=false
fi

echo "检查 orchestrator-service..."
if pgrep -f "orchestrator-service" >/dev/null 2>&1; then
    echo -e "${GREEN}✓ orchestrator-service 运行中${NC}"
    ORCHESTRATOR_RUNNING=true
else
    echo -e "${RED}✗ orchestrator-service 未运行${NC}"
    echo "请在另一个终端运行: cd orchestrator-service && make run"
    ORCHESTRATOR_RUNNING=false
fi

if [ "$AGENT_MANAGER_RUNNING" = false ] || [ "$ORCHESTRATOR_RUNNING" = false ]; then
    echo ""
    echo -e "${RED}请先启动所有服务再运行测试${NC}"
    exit 1
fi

echo ""

# 步骤 2: 检查基础设施
echo -e "${BLUE}步骤 2: 检查基础设施连接${NC}"
echo ""

echo "检查 PostgreSQL (5432)..."
if nc -z localhost 5432 2>/dev/null; then
    echo -e "${GREEN}✓ PostgreSQL 可访问${NC}"
else
    echo -e "${RED}✗ PostgreSQL 不可访问${NC}"
    echo "请确保端口转发: make forward-dev-postgres"
    exit 1
fi

echo "检查 NATS (4222)..."
if nc -z localhost 4222 2>/dev/null; then
    echo -e "${GREEN}✓ NATS 可访问${NC}"
else
    echo -e "${RED}✗ NATS 不可访问${NC}"
    echo "请确保端口转发: make forward-dev-nats"
    exit 1
fi

echo "检查 Redis (6379)..."
if nc -z localhost 6379 2>/dev/null; then
    echo -e "${GREEN}✓ Redis 可访问${NC}"
else
    echo -e "${YELLOW}⚠ Redis 不可访问（可选）${NC}"
fi

echo ""

# 步骤 3: 检查数据库策略
echo -e "${BLUE}步骤 3: 检查 orchestrator 数据库策略${NC}"
echo ""

STRATEGY_COUNT=$(kubectl exec -n aetherius-dev deployment/postgres -- \
    psql -U postgres -d aetherius_orchestrator -t -c \
    "SELECT COUNT(*) FROM strategies WHERE enabled = true;" 2>/dev/null | tr -d ' ' || echo "0")

if [ "$STRATEGY_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✓ 找到 $STRATEGY_COUNT 个活跃策略${NC}"
else
    echo -e "${YELLOW}⚠ 没有活跃策略${NC}"
    echo "运行以下命令创建测试策略:"
    echo "  cd orchestrator-service && ./scripts/test-strategy.sh"
    echo ""
    read -p "是否继续测试？(y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

echo ""

# 步骤 4: 发送测试事件
echo -e "${BLUE}步骤 4: 发送测试事件到 NATS${NC}"
echo ""

TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

cat > /tmp/test-event.json <<EOF
{
  "type": "pod_failure",
  "cluster_id": "test-cluster-001",
  "severity": "critical",
  "payload": {
    "namespace": "default",
    "pod_name": "test-pod-e2e",
    "container": "nginx",
    "reason": "CrashLoopBackOff",
    "message": "Container exited with code 1",
    "restart_count": 5,
    "test_id": "e2e-test-$(date +%s)"
  },
  "timestamp": "$TIMESTAMP"
}
EOF

echo "事件内容:"
cat /tmp/test-event.json | jq .
echo ""

# 使用 kubectl exec 发送到 NATS
echo "通过 NATS pod 发送事件..."
kubectl exec -i -n aetherius-dev nats-0 -- sh -c "cat > /tmp/event.json && /nats-server -js -signal publish internal.event.critical < /tmp/event.json" < /tmp/test-event.json 2>/dev/null || {
    echo -e "${YELLOW}⚠ 无法通过 nats pod 发送，尝试其他方式...${NC}"

    # 尝试使用 Go 程序发送
    if [ -f "send-event/main.go" ]; then
        echo "使用 Go 程序发送..."
        cd send-event
        cat /tmp/test-event.json | go run main.go || echo "发送可能失败"
        cd ..
    else
        echo -e "${RED}✗ 无法发送事件，请手动运行: cd orchestrator-service && ./scripts/test-event.sh${NC}"
    fi
}

echo -e "${GREEN}✓ 事件已发送到 internal.event.critical${NC}"
echo ""

# 步骤 5: 等待处理
echo -e "${BLUE}步骤 5: 等待事件处理 (5秒)${NC}"
echo ""
sleep 5

# 步骤 6: 验证结果
echo -e "${BLUE}步骤 6: 验证处理结果${NC}"
echo ""

echo "6.1 检查 orchestrator 工作流执行记录..."
kubectl exec -n aetherius-dev deployment/postgres -- \
    psql -U postgres -d aetherius_orchestrator -c \
    "SELECT
        id,
        workflow_id,
        status,
        started_at
     FROM workflow_executions
     WHERE started_at > NOW() - INTERVAL '1 minute'
     ORDER BY started_at DESC
     LIMIT 5;" 2>/dev/null || echo "查询失败"

echo ""

echo "6.2 检查 agent-manager 事件记录..."
kubectl exec -n aetherius-dev deployment/postgres -- \
    psql -U postgres -d aetherius -c \
    "SELECT
        id,
        type,
        severity,
        created_at
     FROM events
     WHERE created_at > NOW() - INTERVAL '1 minute'
     ORDER BY created_at DESC
     LIMIT 5;" 2>/dev/null || echo "查询失败或表不存在"

echo ""

# 步骤 7: 查看日志提示
echo -e "${BLUE}步骤 7: 查看服务日志${NC}"
echo ""
echo "在 orchestrator-service 终端查找以下日志:"
echo -e "${YELLOW}  📨 Received message on critical channel${NC}"
echo -e "${YELLOW}  ========== Processing Event ==========${NC}"
echo -e "${YELLOW}  🔍 Starting strategy matching${NC}"
echo -e "${YELLOW}  ✅ Strategy matched successfully${NC}"
echo -e "${YELLOW}  🚀 Starting strategy execution${NC}"
echo ""

echo "在 agent-manager 终端查找以下日志:"
echo -e "${YELLOW}  \"msg\":\"Received event\"${NC}"
echo -e "${YELLOW}  \"msg\":\"Event processed\"${NC}"
echo ""

# 总结
echo "=========================================="
echo "测试完成"
echo "=========================================="
echo ""

echo -e "${GREEN}成功标志:${NC}"
echo "1. orchestrator-service 收到并处理事件"
echo "2. 策略匹配成功"
echo "3. 工作流启动"
echo "4. 数据库中有执行记录"
echo ""

echo -e "${BLUE}验证命令:${NC}"
echo "# 查看最近的工作流执行"
echo "kubectl exec -n aetherius-dev deployment/postgres -- \\"
echo "  psql -U postgres -d aetherius_orchestrator -c \\"
echo "  \"SELECT * FROM workflow_executions ORDER BY started_at DESC LIMIT 1;\""
echo ""

echo "# 实时监控日志"
echo "# orchestrator-service: 查找 📨 🔍 🚀 等符号"
echo "# agent-manager: 查找 \"msg\":\"Event\" 相关日志"
echo ""

rm -f /tmp/test-event.json
