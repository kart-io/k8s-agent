#!/bin/bash
# scripts/refactor/find-old-logger.sh
# 查找所有使用旧版 logger 的文件

set -e

echo "========================================"
echo "Finding files using old logger package"
echo "========================================"
echo ""

OLD_LOGGER="github.com/kart-io/k8s-agent/common/logger"

echo "=== Files using old logger ==="
FILES=$(find . -name "*.go" -type f \
    -not -path "./vendor/*" \
    -not -path "./.git/*" \
    -not -path "./_output/*" \
    -not -path "*.backup*" \
    -exec grep -l "$OLD_LOGGER" {} \;)

if [ -z "$FILES" ]; then
    echo "✓ No files found using old logger!"
    exit 0
fi

echo "$FILES"
echo ""

echo "=== Count by directory ==="
echo "$FILES" | xargs dirname | sort | uniq -c | sort -rn
echo ""

echo "=== Count by service ==="
for service in agent-manager orchestrator reasoning auth gateway monitor cluster collect-agent; do
    count=$(echo "$FILES" | grep -c "internal/$service" || true)
    if [ "$count" -gt 0 ]; then
        printf "%-20s: %d files\n" "$service" "$count"
    fi
done
echo ""

TOTAL=$(echo "$FILES" | wc -l)
echo "Total files to migrate: $TOTAL"
