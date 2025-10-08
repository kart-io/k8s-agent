#!/bin/bash

# Test script for reasoning service

echo "Testing Reasoning Service"
echo "========================="
echo ""

# Health check
echo "1. Health Check"
echo "---------------"
curl -s http://localhost:8082/health | jq '.'
echo ""

# Root cause analysis - OOM example
echo "2. Root Cause Analysis (OOM)"
echo "----------------------------"
curl -s -X POST http://localhost:8082/api/v1/analyze/root-cause \
  -H "Content-Type: application/json" \
  -d '{
    "request_id": "req-test-001",
    "analysis_type": "root_cause",
    "context": {
      "event": {
        "reason": "OOMKilled",
        "message": "Container killed due to OOM"
      },
      "logs": "fatal error: runtime: out of memory\ngoroutine stack exceeds 1000000000 byte limit",
      "metrics": {
        "memory": {
          "usage_percent": 98.5,
          "usage_bytes": 2048000000,
          "limit_bytes": 2097152000
        },
        "cpu": {
          "usage_percent": 75.0
        }
      },
      "cluster_id": "prod-cluster",
      "namespace": "default",
      "resource_name": "my-app-xyz"
    },
    "options": {
      "use_llm": false,
      "min_confidence": 0.7,
      "max_recommendations": 5
    }
  }' | jq '.'
echo ""

# Root cause analysis with LLM
echo "3. Root Cause Analysis (with LLM)"
echo "----------------------------------"
echo "Note: This requires a valid LLM API key"
curl -s -X POST http://localhost:8082/api/v1/analyze/root-cause \
  -H "Content-Type: application/json" \
  -d '{
    "request_id": "req-test-002",
    "analysis_type": "root_cause",
    "context": {
      "event": {
        "reason": "CrashLoopBackOff",
        "message": "Back-off restarting failed container"
      },
      "logs": "Error: Cannot find module '\''./config'\''\nRequire stack:\n- /app/index.js",
      "metrics": {
        "memory": {
          "usage_percent": 25.0
        },
        "cpu": {
          "usage_percent": 10.0
        }
      }
    },
    "options": {
      "use_llm": true,
      "llm_provider": "openai",
      "min_confidence": 0.5
    }
  }' | jq '.'
echo ""

echo "Testing complete!"
