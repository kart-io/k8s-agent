package main

import (
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("发送测试事件到 NATS")
	fmt.Println("========================================")
	fmt.Println()

	// 连接到 NATS
	fmt.Println("🔌 连接到 NATS: nats://localhost:4222")
	nc, err := nats.Connect("nats://localhost:4222",
		nats.Name("e2e-test-sender"),
		nats.Timeout(5*time.Second))
	if err != nil {
		log.Fatalf("❌ 无法连接到 NATS: %v", err)
	}
	defer nc.Close()
	fmt.Println("✅ NATS 连接成功")
	fmt.Println()

	// 准备测试事件
	timestamp := time.Now().UTC().Format(time.RFC3339)
	testID := time.Now().Unix()

	event := fmt.Sprintf(`{
  "type": "pod_failure",
  "cluster_id": "test-cluster-001",
  "severity": "critical",
  "payload": {
    "namespace": "default",
    "pod_name": "test-pod-e2e-%d",
    "container": "nginx",
    "reason": "CrashLoopBackOff",
    "message": "Container exited with code 1",
    "restart_count": 5,
    "test_marker": "E2E_TEST"
  },
  "timestamp": "%s"
}`, testID, timestamp)

	fmt.Println("📤 事件内容:")
	fmt.Println(event)
	fmt.Println()

	// 发送到 critical 频道
	subject := "internal.event.critical"
	fmt.Printf("📨 发送到频道: %s\n", subject)

	if err := nc.Publish(subject, []byte(event)); err != nil {
		log.Fatalf("❌ 发送失败: %v", err)
	}

	// 确保消息发送
	if err := nc.Flush(); err != nil {
		log.Fatalf("❌ Flush 失败: %v", err)
	}

	// 等待一下确保消息送达
	time.Sleep(100 * time.Millisecond)

	fmt.Println("✅ 事件发送成功！")
	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("现在检查 orchestrator-service 日志")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("应该看到以下日志:")
	fmt.Println("  📨 Received message on critical channel")
	fmt.Println("  ========== Processing Event ==========")
	fmt.Println("  ✓ Event parsed successfully")
	fmt.Println("  🔍 Starting strategy matching")
	fmt.Println("  📋 Retrieved strategies from database")
	fmt.Println("  ✅ Strategy matched successfully")
	fmt.Println("  🚀 Starting strategy execution")
	fmt.Println("  🎬 Starting workflow execution")
	fmt.Println("  ========== Strategy execution started successfully ==========")
	fmt.Println()
}
