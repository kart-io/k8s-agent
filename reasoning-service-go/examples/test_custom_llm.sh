#!/bin/bash

# 测试自定义 LLM 集成
# 此脚本演示如何连接到自定义 LLM 服务（如 vLLM, FastChat, LocalAI 等）

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  自定义 LLM 集成测试${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 配置变量
REASONING_SERVICE_URL="${REASONING_SERVICE_URL:-http://localhost:8082}"
CUSTOM_LLM_URL="${CUSTOM_LLM_BASE_URL:-http://localhost:8000}"

# 1. 检查自定义 LLM 服务是否运行
echo -e "${YELLOW}[1/5] 检查自定义 LLM 服务状态...${NC}"
if curl -s -f "${CUSTOM_LLM_URL}/v1/models" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ 自定义 LLM 服务正在运行: ${CUSTOM_LLM_URL}${NC}"

    # 显示可用模型
    echo -e "\n可用模型列表:"
    curl -s "${CUSTOM_LLM_URL}/v1/models" | python3 -m json.tool 2>/dev/null || echo "无法解析模型列表"
else
    echo -e "${RED}✗ 自定义 LLM 服务未运行: ${CUSTOM_LLM_URL}${NC}"
    echo -e "${YELLOW}提示: 请先启动自定义 LLM 服务，例如:${NC}"
    echo -e "  - vLLM: python -m vllm.entrypoints.openai.api_server --model <model-name>"
    echo -e "  - FastChat: python -m fastchat.serve.openai_api_server"
    echo -e "  - LocalAI: docker run -p 8080:8080 localai/localai:latest"
    echo -e "  - LM Studio: 在设置中开启 'Local Server'"
    echo ""
    exit 1
fi
echo ""

# 2. 测试自定义 LLM API
echo -e "${YELLOW}[2/5] 测试自定义 LLM API...${NC}"
TEST_MODEL="${CUSTOM_LLM_MODEL:-gpt-3.5-turbo}"
TEST_RESPONSE=$(curl -s -X POST "${CUSTOM_LLM_URL}/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -d "{
    \"model\": \"${TEST_MODEL}\",
    \"messages\": [{\"role\": \"user\", \"content\": \"Hello\"}],
    \"max_tokens\": 10
  }")

if echo "$TEST_RESPONSE" | grep -q "choices"; then
    echo -e "${GREEN}✓ 自定义 LLM API 响应正常${NC}"
    echo -e "测试响应: $(echo $TEST_RESPONSE | python3 -m json.tool 2>/dev/null | head -20)"
else
    echo -e "${RED}✗ 自定义 LLM API 测试失败${NC}"
    echo "响应内容: $TEST_RESPONSE"
    exit 1
fi
echo ""

# 3. 检查 Reasoning Service
echo -e "${YELLOW}[3/5] 检查 Reasoning Service...${NC}"
HEALTH_RESPONSE=$(curl -s "${REASONING_SERVICE_URL}/health")
if echo "$HEALTH_RESPONSE" | grep -q "healthy"; then
    echo -e "${GREEN}✓ Reasoning Service 运行正常${NC}"
    echo "$HEALTH_RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$HEALTH_RESPONSE"
else
    echo -e "${RED}✗ Reasoning Service 未运行${NC}"
    echo -e "${YELLOW}提示: 请先启动 Reasoning Service:${NC}"
    echo -e "  go run cmd/server/main.go"
    exit 1
fi
echo ""

# 4. 测试使用自定义 LLM 进行根因分析
echo -e "${YELLOW}[4/5] 测试根因分析（使用自定义 LLM）...${NC}"

REQUEST_ID="custom-llm-test-$(date +%s)"

ANALYSIS_REQUEST='{
  "request_id": "'$REQUEST_ID'",
  "analysis_type": "root_cause",
  "context": {
    "event": {
      "reason": "OOMKilled",
      "message": "Container was killed due to out of memory"
    },
    "logs": "fatal error: runtime: out of memory\ngoroutine 1 [running]:\nruntime.throw(0x1a9f5e0, 0x16)\n",
    "metrics": {
      "memory": {
        "usage_percent": 98.5,
        "limit_bytes": 536870912,
        "usage_bytes": 528482304
      },
      "cpu": {
        "usage_percent": 75.0
      }
    }
  },
  "options": {
    "use_llm": true,
    "llm_provider": "custom",
    "min_confidence": 0.7,
    "max_recommendations": 5
  }
}'

echo "发送分析请求..."
ANALYSIS_RESPONSE=$(curl -s -X POST "${REASONING_SERVICE_URL}/api/v1/analyze/root-cause" \
  -H "Content-Type: application/json" \
  -d "$ANALYSIS_REQUEST")

if echo "$ANALYSIS_RESPONSE" | grep -q "completed"; then
    echo -e "${GREEN}✓ 根因分析完成${NC}"
    echo ""
    echo "=== 分析结果 ==="
    echo "$ANALYSIS_RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$ANALYSIS_RESPONSE"

    # 提取关键信息
    ROOT_CAUSE=$(echo "$ANALYSIS_RESPONSE" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data['result']['root_cause']['type'])" 2>/dev/null || echo "Unknown")
    CONFIDENCE=$(echo "$ANALYSIS_RESPONSE" | python3 -c "import sys, json; data=json.load(sys.stdin); print(data['result']['confidence'])" 2>/dev/null || echo "N/A")

    echo ""
    echo -e "${GREEN}根因类型: ${ROOT_CAUSE}${NC}"
    echo -e "${GREEN}置信度: ${CONFIDENCE}${NC}"
else
    echo -e "${RED}✗ 根因分析失败${NC}"
    echo "响应: $ANALYSIS_RESPONSE"
    exit 1
fi
echo ""

# 5. 性能测试
echo -e "${YELLOW}[5/5] 性能测试...${NC}"

PERFORMANCE_REQUEST='{
  "request_id": "perf-test-'$(date +%s)'",
  "context": {
    "event": {
      "reason": "CrashLoopBackOff",
      "message": "Container keeps crashing"
    },
    "logs": "Error: Failed to connect to database\nConnection timeout after 30s"
  },
  "options": {
    "use_llm": true,
    "llm_provider": "custom"
  }
}'

echo "执行性能测试（3次请求）..."
TOTAL_TIME=0
for i in {1..3}; do
    START_TIME=$(date +%s%N)

    PERF_RESPONSE=$(curl -s -X POST "${REASONING_SERVICE_URL}/api/v1/analyze/root-cause" \
      -H "Content-Type: application/json" \
      -d "$PERFORMANCE_REQUEST")

    END_TIME=$(date +%s%N)
    ELAPSED=$((($END_TIME - $START_TIME) / 1000000))
    TOTAL_TIME=$(($TOTAL_TIME + $ELAPSED))

    echo "  请求 $i: ${ELAPSED}ms"
done

AVG_TIME=$(($TOTAL_TIME / 3))
echo -e "${GREEN}平均响应时间: ${AVG_TIME}ms${NC}"
echo ""

# 总结
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  测试完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "✅ 自定义 LLM 服务: ${CUSTOM_LLM_URL}"
echo -e "✅ Reasoning Service: ${REASONING_SERVICE_URL}"
echo -e "✅ 平均响应时间: ${AVG_TIME}ms"
echo ""
echo -e "${YELLOW}提示:${NC}"
echo -e "  - 查看完整文档: docs/CUSTOM_LLM_GUIDE.md"
echo -e "  - 配置文件: configs/config.yaml"
echo -e "  - 环境变量: CUSTOM_LLM_BASE_URL, CUSTOM_LLM_MODEL, CUSTOM_LLM_API_KEY"
echo ""
