#!/bin/bash
# examples/test-examples.sh
# 测试所有示例代码

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

echo -e "${BLUE}=== Testing Proto Examples ===${NC}\n"

# 1. 检查生成的代码
echo -e "${YELLOW}1. Checking generated code...${NC}"
if [ ! -d "pkg/api/agent/v1" ]; then
    echo -e "${RED}✗ Generated code not found${NC}"
    echo "Run: make proto.generate"
    exit 1
fi
echo -e "${GREEN}✓ Generated code found${NC}\n"

# 2. 编译 gRPC Server
echo -e "${YELLOW}2. Building gRPC Server...${NC}"
if go build -o /tmp/grpc-server examples/grpc-server/main.go; then
    echo -e "${GREEN}✓ gRPC Server built successfully${NC}\n"
else
    echo -e "${RED}✗ Failed to build gRPC Server${NC}"
    exit 1
fi

# 3. 编译 gRPC Client
echo -e "${YELLOW}3. Building gRPC Client...${NC}"
if go build -o /tmp/grpc-client examples/grpc-client/main.go; then
    echo -e "${GREEN}✓ gRPC Client built successfully${NC}\n"
else
    echo -e "${RED}✗ Failed to build gRPC Client${NC}"
    exit 1
fi

# 4. 编译 HTTP Gateway
echo -e "${YELLOW}4. Building HTTP Gateway...${NC}"
if go build -o /tmp/http-gateway examples/http-gateway/main.go; then
    echo -e "${GREEN}✓ HTTP Gateway built successfully${NC}\n"
else
    echo -e "${RED}✗ Failed to build HTTP Gateway${NC}"
    exit 1
fi

# 5. 测试 gRPC Server + Client
echo -e "${YELLOW}5. Testing gRPC Server + Client...${NC}"

# 启动服务器
/tmp/grpc-server > /tmp/grpc-server.log 2>&1 &
SERVER_PID=$!
echo "Started gRPC Server (PID: $SERVER_PID)"

# 等待服务器启动
sleep 2

# 运行客户端
if /tmp/grpc-client > /tmp/grpc-client.log 2>&1; then
    echo -e "${GREEN}✓ gRPC Client test passed${NC}"
    cat /tmp/grpc-client.log | grep "✓" || true
else
    echo -e "${RED}✗ gRPC Client test failed${NC}"
    cat /tmp/grpc-client.log
    kill $SERVER_PID 2>/dev/null || true
    exit 1
fi

# 清理
kill $SERVER_PID 2>/dev/null || true
echo ""

# 6. 测试 HTTP Gateway
echo -e "${YELLOW}6. Testing HTTP Gateway...${NC}"

# 启动 Gateway
/tmp/http-gateway > /tmp/http-gateway.log 2>&1 &
GATEWAY_PID=$!
echo "Started HTTP Gateway (PID: $GATEWAY_PID)"

# 等待 Gateway 启动
sleep 3

# 测试健康检查
echo -n "  Testing health endpoint... "
if curl -s http://localhost:8080/health | grep -q "ok"; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗${NC}"
    kill $GATEWAY_PID 2>/dev/null || true
    exit 1
fi

# 测试注册 Agent
echo -n "  Testing RegisterAgent API... "
REGISTER_RESPONSE=$(curl -s -X POST http://localhost:8080/agent.v1.AgentService/RegisterAgent \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-agent",
    "cluster_id": "cluster-1",
    "cluster_name": "test-cluster",
    "version": "v1.0.0"
  }')

if echo "$REGISTER_RESPONSE" | grep -q "agent"; then
    echo -e "${GREEN}✓${NC}"
    AGENT_ID=$(echo "$REGISTER_RESPONSE" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    echo "    Agent ID: $AGENT_ID"
else
    echo -e "${RED}✗${NC}"
    echo "$REGISTER_RESPONSE"
    kill $GATEWAY_PID 2>/dev/null || true
    exit 1
fi

# 测试列出 Agent
echo -n "  Testing ListAgents API... "
if curl -s -X POST http://localhost:8080/agent.v1.AgentService/ListAgents \
  -H "Content-Type: application/json" \
  -d '{}' | grep -q "agents"; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗${NC}"
    kill $GATEWAY_PID 2>/dev/null || true
    exit 1
fi

# 清理
kill $GATEWAY_PID 2>/dev/null || true
echo ""

# 7. 清理临时文件
echo -e "${YELLOW}7. Cleaning up...${NC}"
rm -f /tmp/grpc-server /tmp/grpc-client /tmp/http-gateway
rm -f /tmp/grpc-server.log /tmp/grpc-client.log /tmp/http-gateway.log
echo -e "${GREEN}✓ Cleanup complete${NC}\n"

# 总结
echo -e "${GREEN}=== All Tests Passed! ===${NC}\n"
echo "Examples are working correctly:"
echo "  ✓ gRPC Server"
echo "  ✓ gRPC Client"
echo "  ✓ HTTP Gateway"
echo ""
echo "Try running the examples:"
echo "  go run examples/grpc-server/main.go"
echo "  go run examples/grpc-client/main.go"
echo "  go run examples/http-gateway/main.go"
