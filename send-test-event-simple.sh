#!/bin/bash

# 简化版测试事件发送脚本 - 使用 kubectl 直接在 NATS pod 中执行

echo "=========================================="
echo "通过 NATS Pod 发送测试事件"
echo "=========================================="
echo ""

TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

cat <<EOF | kubectl exec -i -n aetherius-dev nats-0 -- nats pub internal.event.critical
{
  "type": "pod_failure",
  "cluster_id": "test-cluster-001",
  "severity": "critical",
  "payload": {
    "namespace": "default",
    "pod_name": "crashloop-test-pod",
    "container": "crash-container",
    "reason": "CrashLoopBackOff",
    "message": "Container exited with code 1",
    "restart_count": 6
  },
  "timestamp": "$TIMESTAMP"
}
EOF

echo ""
echo "✅ 测试事件已发送到 internal.event.critical"
echo ""
echo "现在检查 orchestrator-service 的日志，应该看到："
echo "  📨 Received message on critical channel"
echo "  ========== Processing Event =========="
echo "  🔍 Starting strategy matching"
echo ""
