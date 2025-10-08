#!/bin/bash

echo "=========================================="
echo "验证 Orchestrator Service 功能"
echo "=========================================="
echo ""

# 检查 Pod 状态
echo "1. 检查测试 Pod 状态："
kubectl get pod crashloop-test-pod -n default
echo ""

# 检查 collect-agent 日志（如果运行）
echo "2. 检查 collect-agent 是否捕获事件："
if kubectl get pod test-event-collection -n aetherius-dev &>/dev/null; then
    echo "查看最近的 collect-agent 日志："
    kubectl logs test-event-collection -n aetherius-dev --tail=20 | grep -i "crash\|error\|event" || echo "未找到相关事件"
else
    echo "collect-agent 未运行"
fi
echo ""

# 检查数据库中的事件
echo "3. 检查数据库中的事件记录："
echo "连接到 PostgreSQL 查询最近的事件..."
kubectl exec -n aetherius-dev deployment/postgres -- psql -U postgres -d aetherius -c "
SELECT id, type, severity, created_at
FROM events
WHERE created_at > NOW() - INTERVAL '10 minutes'
ORDER BY created_at DESC
LIMIT 5;" 2>/dev/null || echo "无法连接到数据库或表不存在"
echo ""

# 检查 orchestrator-service 的策略
echo "4. 检查 orchestrator-service 的策略配置："
kubectl exec -n aetherius-dev deployment/postgres -- psql -U postgres -d aetherius_orchestrator -c "
SELECT id, name, category, enabled
FROM strategies
WHERE enabled = true
ORDER BY priority DESC;" 2>/dev/null || echo "orchestrator 数据库未初始化"
echo ""

# 检查工作流执行记录
echo "5. 检查工作流执行记录："
kubectl exec -n aetherius-dev deployment/postgres -- psql -U postgres -d aetherius_orchestrator -c "
SELECT id, workflow_id, status, started_at
FROM workflow_executions
ORDER BY started_at DESC
LIMIT 5;" 2>/dev/null || echo "无工作流执行记录"
echo ""

# 检查 NATS 连接
echo "6. 测试 NATS 连接："
if command -v nats &> /dev/null; then
    echo "NATS CLI 可用，尝试连接..."
    kubectl port-forward -n aetherius-dev svc/nats 4222:4222 &>/dev/null &
    PF_PID=$!
    sleep 2
    nats server ping --server=nats://localhost:4222 2>/dev/null && echo "✓ NATS 连接正常" || echo "✗ NATS 连接失败"
    kill $PF_PID 2>/dev/null
else
    echo "NATS CLI 未安装"
fi
echo ""

echo "=========================================="
echo "验证完成"
echo "=========================================="
echo ""
echo "下一步操作："
echo "1. 如果 orchestrator-service 未运行，执行: cd orchestrator-service && make run"
echo "2. 确保已创建测试策略: cd orchestrator-service && ./scripts/test-strategy.sh"
echo "3. 发送测试事件: cd orchestrator-service && ./scripts/test-event.sh"
echo "4. 监控 Pod 状态: kubectl get pod crashloop-test-pod -w"
