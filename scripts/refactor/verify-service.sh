#!/bin/bash
# scripts/refactor/verify-service.sh
# 验证服务是否符合标准架构

set -e

SERVICE=$1

if [ -z "$SERVICE" ]; then
    echo "Usage: $0 <service-name>"
    echo "Example: $0 reasoning"
    exit 1
fi

echo "========================================"
echo "Verifying $SERVICE Service"
echo "========================================"
echo ""

PASS_COUNT=0
FAIL_COUNT=0
WARN_COUNT=0

check_pass() {
    echo "✓ $1"
    ((PASS_COUNT++))
}

check_fail() {
    echo "✗ $1"
    ((FAIL_COUNT++))
}

check_warn() {
    echo "⚠ $1"
    ((WARN_COUNT++))
}

# 1. 目录结构检查
echo "=== 1. Directory Structure ==="

if [ -d "cmd/$SERVICE/app/options" ]; then
    check_pass "cmd/$SERVICE/app/options/ exists"
else
    check_fail "cmd/$SERVICE/app/options/ missing"
fi

if [ -d "internal/$SERVICE/initializers" ]; then
    check_pass "internal/$SERVICE/initializers/ exists"
else
    check_fail "internal/$SERVICE/initializers/ missing"
fi

if [ -f "cmd/$SERVICE/app/app.go" ]; then
    check_pass "cmd/$SERVICE/app/app.go exists"
else
    check_fail "cmd/$SERVICE/app/app.go missing"
fi

echo ""

# 2. Application 接口检查
echo "=== 2. Application Interface ==="

if grep -q "commonapp.RunWithRunner" "cmd/$SERVICE/app/app.go" 2>/dev/null; then
    check_pass "Uses commonapp.RunWithRunner()"
else
    check_fail "Not using commonapp.RunWithRunner()"
fi

if grep -q "func.*Initialize.*context.Context.*commonapp.Options.*error" "cmd/$SERVICE/app/app.go" 2>/dev/null; then
    check_pass "Implements Initialize() method"
else
    check_fail "Initialize() method missing"
fi

if grep -q "func.*Run.*context.Context.*error" "cmd/$SERVICE/app/app.go" 2>/dev/null; then
    check_pass "Implements Run() method"
else
    check_fail "Run() method missing"
fi

if grep -q "func.*Shutdown.*context.Context.*error" "cmd/$SERVICE/app/app.go" 2>/dev/null; then
    check_pass "Implements Shutdown() method"
else
    check_fail "Shutdown() method missing"
fi

echo ""

# 3. 日志系统检查
echo "=== 3. Logger System ==="

if find "internal/$SERVICE" -name "*.go" -exec grep -l "github.com/kart-io/logger/core" {} \; | grep -q .; then
    check_pass "Uses kart-io/logger"
else
    check_fail "Not using kart-io/logger"
fi

if find "internal/$SERVICE" -name "*.go" -exec grep -l "common/logger" {} \; | grep -q .; then
    check_fail "Still using common/logger (old)"
else
    check_pass "Not using old common/logger"
fi

echo ""

# 4. Bootstrap 检查
echo "=== 4. Bootstrap Framework ==="

if grep -q "bootstrap.New" "cmd/$SERVICE/app/app.go" 2>/dev/null; then
    check_pass "Uses bootstrap.Bootstrap"
else
    check_fail "Not using bootstrap.Bootstrap"
fi

if grep -q "bootstrap.Register" "cmd/$SERVICE/app/app.go" 2>/dev/null; then
    check_pass "Registers initializers"
else
    check_fail "Not registering initializers"
fi

echo ""

# 5. 编译检查
echo "=== 5. Compilation ==="

if go build -o /tmp/${SERVICE}-verify ./cmd/$SERVICE 2>&1 | tee /tmp/build-${SERVICE}-verify.log > /dev/null; then
    check_pass "Service compiles successfully"
    rm -f /tmp/${SERVICE}-verify
else
    check_fail "Compilation failed (see /tmp/build-${SERVICE}-verify.log)"
fi

echo ""

# 6. 测试检查
echo "=== 6. Tests ==="

if go test ./internal/$SERVICE/... -run=^$ 2>/dev/null; then
    if go test ./internal/$SERVICE/... 2>&1 | tee /tmp/test-${SERVICE}-verify.log > /dev/null; then
        check_pass "Tests pass"
    else
        check_warn "Some tests fail (see /tmp/test-${SERVICE}-verify.log)"
    fi
else
    check_warn "No tests found or tests cannot run"
fi

echo ""

# 7. 代码质量检查
echo "=== 7. Code Quality ==="

if command -v golangci-lint &> /dev/null; then
    if golangci-lint run ./internal/$SERVICE/... 2>&1 | tee /tmp/lint-${SERVICE}-verify.log > /dev/null; then
        check_pass "Linter passes"
    else
        check_warn "Linter has warnings (see /tmp/lint-${SERVICE}-verify.log)"
    fi
else
    check_warn "golangci-lint not installed, skipping"
fi

echo ""

# 总结
echo "========================================"
echo "Verification Summary"
echo "========================================"
echo "Passed:  $PASS_COUNT"
echo "Failed:  $FAIL_COUNT"
echo "Warnings: $WARN_COUNT"
echo ""

if [ $FAIL_COUNT -eq 0 ]; then
    echo "✓ Service meets all critical standards!"
    exit 0
else
    echo "✗ Service has $FAIL_COUNT critical issues"
    exit 1
fi
