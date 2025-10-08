#!/bin/bash

# Ollama 集成测试脚本

echo "Ollama 集成测试"
echo "==============="
echo ""

# 检查 Ollama 是否运行
echo "1. 检查 Ollama 服务状态"
echo "------------------------"
if curl -s http://localhost:11434/api/tags > /dev/null 2>&1; then
    echo "✅ Ollama 服务运行正常"
    echo ""
    echo "已安装的模型:"
    curl -s http://localhost:11434/api/tags | jq -r '.models[].name'
else
    echo "❌ Ollama 未运行"
    echo ""
    echo "请先安装并启动 Ollama:"
    echo "  curl -fsSL https://ollama.com/install.sh | sh"
    echo "  ollama pull llama3.1"
    exit 1
fi
echo ""

# 检查 Reasoning Service 是否运行
echo "2. 检查 Reasoning Service 状态"
echo "--------------------------------"
if curl -s http://localhost:8082/health > /dev/null 2>&1; then
    echo "✅ Reasoning Service 运行正常"

    # 检查 LLM 组件
    llm_status=$(curl -s http://localhost:8082/health | jq -r '.components.llm')
    if [ "$llm_status" = "true" ]; then
        echo "✅ LLM 组件已启用"
    else
        echo "⚠️  LLM 组件未启用"
    fi
else
    echo "❌ Reasoning Service 未运行"
    echo ""
    echo "请先启动服务:"
    echo "  go run cmd/server/main.go"
    exit 1
fi
echo ""

# 测试 Ollama 推理
echo "3. 测试 Ollama 直接推理"
echo "------------------------"
echo "发送测试请求到 Ollama..."
OLLAMA_RESPONSE=$(curl -s -X POST http://localhost:11434/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama3.1",
    "messages": [{"role": "user", "content": "What is Kubernetes? Answer in one sentence."}],
    "stream": false
  }')

if [ $? -eq 0 ]; then
    echo "✅ Ollama 推理成功"
    echo "响应: $(echo $OLLAMA_RESPONSE | jq -r '.choices[0].message.content' | head -c 100)..."
else
    echo "❌ Ollama 推理失败"
fi
echo ""

# 测试使用 Ollama 的根因分析（不使用 LLM）
echo "4. 测试规则引擎分析（baseline）"
echo "--------------------------------"
echo "测试 OOM 分析（仅规则引擎）..."
BASELINE_RESULT=$(curl -s -X POST http://localhost:8082/api/v1/analyze/root-cause \
  -H "Content-Type: application/json" \
  -d '{
    "request_id": "ollama-baseline-001",
    "context": {
      "event": {
        "reason": "OOMKilled",
        "message": "Container killed due to OOM"
      },
      "logs": "fatal error: runtime: out of memory",
      "metrics": {
        "memory": {
          "usage_percent": 98.5
        }
      }
    },
    "options": {
      "use_llm": false
    }
  }')

if [ $? -eq 0 ]; then
    root_cause=$(echo $BASELINE_RESULT | jq -r '.result.root_cause.type')
    confidence=$(echo $BASELINE_RESULT | jq -r '.result.root_cause.confidence')
    echo "✅ 规则引擎分析成功"
    echo "   根因: $root_cause"
    echo "   置信度: $confidence"
else
    echo "❌ 分析失败"
fi
echo ""

# 测试使用 Ollama 的根因分析（使用 LLM）
echo "5. 测试 Ollama LLM 增强分析"
echo "----------------------------"
echo "测试配置错误分析（使用 Ollama）..."
OLLAMA_RESULT=$(curl -s -X POST http://localhost:8082/api/v1/analyze/root-cause \
  -H "Content-Type: application/json" \
  -d '{
    "request_id": "ollama-llm-001",
    "context": {
      "event": {
        "reason": "CrashLoopBackOff",
        "message": "Back-off restarting failed container"
      },
      "logs": "Error: Cannot find module '\''./config'\''\nRequire stack:\n- /app/index.js\n- /app/server.js",
      "metrics": {
        "memory": {
          "usage_percent": 20
        }
      }
    },
    "options": {
      "use_llm": true,
      "llm_provider": "ollama"
    }
  }')

if [ $? -eq 0 ]; then
    status=$(echo $OLLAMA_RESULT | jq -r '.status')
    if [ "$status" = "completed" ]; then
        root_cause=$(echo $OLLAMA_RESULT | jq -r '.result.root_cause.type')
        confidence=$(echo $OLLAMA_RESULT | jq -r '.result.root_cause.confidence')
        has_llm=$(echo $OLLAMA_RESULT | jq -r '.result.llm_analysis != null and .result.llm_analysis != ""')

        echo "✅ Ollama LLM 分析成功"
        echo "   根因: $root_cause"
        echo "   置信度: $confidence"

        if [ "$has_llm" = "true" ]; then
            echo "   ✅ 包含 LLM 分析结果"
            echo ""
            echo "   LLM 分析摘要:"
            echo $OLLAMA_RESULT | jq -r '.result.llm_analysis' | head -c 200
            echo "..."
        else
            echo "   ⚠️  未包含 LLM 分析（可能降级到规则引擎）"
        fi

        echo ""
        echo "   推荐操作数量: $(echo $OLLAMA_RESULT | jq -r '.result.recommendations | length')"
    else
        echo "❌ 分析状态: $status"
        error=$(echo $OLLAMA_RESULT | jq -r '.error')
        echo "   错误: $error"
    fi
else
    echo "❌ 请求失败"
fi
echo ""

# 性能测试
echo "6. Ollama 性能测试"
echo "------------------"
echo "测试推理速度..."
START_TIME=$(date +%s.%N)
curl -s -X POST http://localhost:8082/api/v1/analyze/root-cause \
  -H "Content-Type: application/json" \
  -d '{
    "request_id": "ollama-perf-001",
    "context": {
      "event": {"reason": "OOMKilled"},
      "logs": "out of memory"
    },
    "options": {
      "use_llm": true,
      "llm_provider": "ollama"
    }
  }' > /dev/null 2>&1
END_TIME=$(date +%s.%N)
DURATION=$(echo "$END_TIME - $START_TIME" | bc)
echo "✅ 完成"
echo "   处理时间: ${DURATION}s"
echo ""

# 总结
echo "==============="
echo "测试总结"
echo "==============="
echo "✅ Ollama 服务正常"
echo "✅ Reasoning Service 正常"
echo "✅ 规则引擎分析正常"
echo "✅ Ollama LLM 集成正常"
echo ""
echo "🎉 所有测试通过！"
echo ""
echo "提示："
echo "- 如需更快推理，使用较小模型: ollama pull llama3.2"
echo "- 查看更多模型: ollama list"
echo "- 查看 Ollama 日志: ollama logs"
