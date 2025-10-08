package main

import (
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	// 连接到 NATS
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatal("Failed to connect to NATS:", err)
	}
	defer nc.Close()

	fmt.Println("✅ Connected to NATS: nats://localhost:4222")
	fmt.Println()

	// 准备测试事件
	timestamp := time.Now().UTC().Format(time.RFC3339)
	event := fmt.Sprintf(`{
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
  "timestamp": "%s"
}`, timestamp)

	// 发送事件
	subject := "internal.event.critical"
	fmt.Printf("📤 Sending event to subject: %s\n", subject)
	fmt.Println("Event payload:")
	fmt.Println(event)
	fmt.Println()

	if err := nc.Publish(subject, []byte(event)); err != nil {
		log.Fatal("Failed to publish event:", err)
	}

	// 确保消息发送
	if err := nc.Flush(); err != nil {
		log.Fatal("Failed to flush:", err)
	}

	fmt.Println("✅ Event sent successfully!")
	fmt.Println()
	fmt.Println("Now check orchestrator-service logs for:")
	fmt.Println("  📨 Received message on critical channel")
	fmt.Println("  ========== Processing Event ==========")
	fmt.Println("  🔍 Starting strategy matching")
	fmt.Println()
}
