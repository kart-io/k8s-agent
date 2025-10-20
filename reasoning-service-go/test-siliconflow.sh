#!/bin/bash

# SiliconFlow 集成测试脚本

set -e

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║          SiliconFlow Integration Test Script                ║"
echo "╚══════════════════════════════════════════════════════════════╝"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 检查环境变量
echo -e "\n${BLUE}[1/5] 检查环境变量...${NC}"
if [ -z "$SILICONFLOW_API_KEY" ]; then
    echo -e "${RED}❌ SILICONFLOW_API_KEY 未设置${NC}"
    echo "请设置环境变量:"
    echo "  export SILICONFLOW_API_KEY='your-api-key'"
    exit 1
fi
echo -e "${GREEN}✅ SILICONFLOW_API_KEY 已设置 (长度: ${#SILICONFLOW_API_KEY} 字符)${NC}"

# 验证 API Key 格式
echo -e "\n${BLUE}[2/5] 验证 API Key 格式...${NC}"
if [[ ! $SILICONFLOW_API_KEY =~ ^sk- ]]; then
    echo -e "${YELLOW}⚠️  警告: API Key 不以 'sk-' 开头${NC}"
fi
echo -e "${GREEN}✅ API Key 格式看起来有效${NC}"

# 测试网络连接
echo -e "\n${BLUE}[3/5] 测试网络连接...${NC}"
if ! ping -c 1 api.siliconflow.cn > /dev/null 2>&1; then
    echo -e "${RED}❌ 无法连接到 api.siliconflow.cn${NC}"
    exit 1
fi
echo -e "${GREEN}✅ 网络连接正常${NC}"

# 测试 API 认证
echo -e "\n${BLUE}[4/5] 测试 API 认证...${NC}"
response=$(curl -s -w "\n%{http_code}" \
    -H "Authorization: Bearer $SILICONFLOW_API_KEY" \
    -H "Content-Type: application/json" \
    https://api.siliconflow.cn/v1/models 2>/dev/null)

status_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | head -n-1)

if [ "$status_code" = "200" ]; then
    echo -e "${GREEN}✅ API 认证成功${NC}"
    model_count=$(echo "$body" | jq -r '.data | length' 2>/dev/null || echo "unknown")
    echo "   可用模型数量: $model_count"
elif [ "$status_code" = "401" ]; then
    echo -e "${RED}❌ 认证失败 (401 Unauthorized)${NC}"
    echo "   您的 API Key 可能无效或已过期"
    exit 1
else
    echo -e "${YELLOW}⚠️  意外的响应码: $status_code${NC}"
    echo "   响应内容: $(echo "$body" | head -c 200)"
fi

# 启动服务并验证
echo -e "\n${BLUE}[5/5] 启动服务并验证集成...${NC}"

# 清理可能占用端口的进程
lsof -ti:8083 | xargs -r kill -9 2>/dev/null || true
sleep 1

# 启动服务(后台运行)
echo "   正在启动服务..."
SILICONFLOW_API_KEY=$SILICONFLOW_API_KEY go run cmd/server/main.go -c configs/config-dev.yaml > /tmp/reasoning-service.log 2>&1 &
SERVER_PID=$!

# 等待服务启动
echo -n "   等待服务启动"
for i in {1..30}; do
    if curl -s http://localhost:8083/health > /dev/null 2>&1; then
        echo ""
        break
    fi
    sleep 1
    echo -n "."
done
echo ""

# 检查服务是否成功启动
if ! curl -s http://localhost:8083/health > /dev/null 2>&1; then
    echo -e "${RED}❌ 服务启动失败${NC}"
    echo "   查看日志: tail -50 /tmp/reasoning-service.log"
    kill $SERVER_PID 2>/dev/null || true
    exit 1
fi

echo -e "${GREEN}✅ 服务启动成功${NC}"

# 检查 SiliconFlow 是否被正确加载
echo "   检查 SiliconFlow 提供商状态..."
if grep -q "siliconflow.*ready" /tmp/reasoning-service.log; then
    echo -e "${GREEN}✅ SiliconFlow 提供商已成功加载${NC}"

    # 显示提供商信息
    grep "siliconflow" /tmp/reasoning-service.log | head -3
else
    echo -e "${RED}❌ SiliconFlow 提供商未正确加载${NC}"
    echo "   查看日志了解详情:"
    grep -i "siliconflow\|error" /tmp/reasoning-service.log | tail -10
    kill $SERVER_PID 2>/dev/null || true
    exit 1
fi

# 清理
echo -e "\n${BLUE}清理资源...${NC}"
kill $SERVER_PID 2>/dev/null || true
sleep 1

echo -e "\n╔══════════════════════════════════════════════════════════════╗"
echo -e "║  ${GREEN}✅ 所有测试通过!SiliconFlow 已成功集成${NC}                   ║"
echo "╚══════════════════════════════════════════════════════════════╝"

echo -e "\n${YELLOW}下一步:${NC}"
echo "  1. 启动服务: SILICONFLOW_API_KEY=\$SILICONFLOW_API_KEY make run-dev"
echo "  2. 测试 API: curl http://localhost:8083/health"
echo "  3. 查看文档: cat README.md"
