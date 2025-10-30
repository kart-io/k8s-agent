#!/bin/bash
# scripts/refactor/migrate-logger.sh
# 批量迁移服务的日志系统到 kart-io/logger

set -e

SERVICE=$1

if [ -z "$SERVICE" ]; then
    echo "Usage: $0 <service-name>"
    echo ""
    echo "Available services:"
    echo "  reasoning"
    echo "  collect-agent"
    echo ""
    echo "Example: $0 reasoning"
    exit 1
fi

echo "========================================"
echo "Migrating $SERVICE to kart-io/logger"
echo "========================================"
echo ""

SERVICE_DIR="internal/$SERVICE"

if [ ! -d "$SERVICE_DIR" ]; then
    echo "Error: Service directory $SERVICE_DIR not found"
    exit 1
fi

# 备份
BACKUP_DIR="${SERVICE_DIR}.logger-migration-backup-$(date +%Y%m%d-%H%M%S)"
echo "1. Creating backup at $BACKUP_DIR"
cp -r "$SERVICE_DIR" "$BACKUP_DIR"
echo "✓ Backup created"
echo ""

# 查找需要修改的文件
echo "2. Finding files to migrate..."
FILES=$(find "$SERVICE_DIR" -name "*.go" -type f -exec grep -l "common/logger" {} \;)

if [ -z "$FILES" ]; then
    echo "✓ No files need migration!"
    rm -rf "$BACKUP_DIR"
    exit 0
fi

FILE_COUNT=$(echo "$FILES" | wc -l)
echo "Found $FILE_COUNT files to migrate"
echo ""

# 执行替换
echo "3. Replacing imports..."

# 替换导入语句
echo "$FILES" | xargs sed -i \
    's|"github.com/kart-io/k8s-agent/common/logger"|"github.com/kart-io/logger/core"|g'

# 替换常见的包别名
echo "$FILES" | xargs sed -i 's|commonlogger\.|core.|g'

# 替换 logger.Logger 类型
echo "$FILES" | xargs sed -i 's|logger\.Logger|core.Logger|g'

# 替换初始化函数（如果有）
echo "$FILES" | xargs sed -i 's|logger\.InitFromOptions|// MANUAL: Update to use core.Logger initialization|g'

echo "✓ Imports replaced"
echo ""

# 编译检查
echo "4. Compiling to verify changes..."
if go build ./cmd/$SERVICE 2>&1 | tee /tmp/build-${SERVICE}.log; then
    echo "✓ Compilation successful!"
    echo ""
    echo "Migration completed successfully!"
    echo ""
    echo "Next steps:"
    echo "1. Review the changes carefully"
    echo "2. Update logger initialization code if needed"
    echo "3. Run tests: go test ./internal/$SERVICE/... -v"
    echo "4. If everything works, delete backup: rm -rf $BACKUP_DIR"
else
    echo "✗ Compilation failed!"
    echo ""
    echo "Build errors saved to /tmp/build-${SERVICE}.log"
    echo "Backup is at: $BACKUP_DIR"
    echo ""
    echo "To rollback:"
    echo "  rm -rf $SERVICE_DIR"
    echo "  mv $BACKUP_DIR $SERVICE_DIR"
    exit 1
fi
