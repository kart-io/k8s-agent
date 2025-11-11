# Agent 流式输出快速开始

这是一个 5 分钟快速入门指南，帮助你快速上手 Agent 框架的流式输出功能。

## 安装依赖

```bash
# 安装 gorilla/websocket (如果使用 WebSocket)
go get github.com/gorilla/websocket

# 安装 gorilla/mux (如果使用 HTTP 路由)
go get github.com/gorilla/mux

# 安装 google/uuid (多路复用需要)
go get github.com/google/uuid
```

## 场景 1: 基本流式输出 (30 秒)

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/kart-io/k8s-agent/pkg/agent/core"
    "github.com/kart-io/k8s-agent/pkg/agent/stream/agents"
)

func main() {
    // 1. 创建 Agent
    agent := agents.NewProgressAgent(nil)

    // 2. 准备输入
    input := &core.AgentInput{
        Task: "Test task",
        Context: map[string]interface{}{
            "total_steps": 10,
        },
    }

    // 3. 执行流式任务
    stream, err := agent.ExecuteStream(context.Background(), input)
    if err != nil {
        log.Fatal(err)
    }
    defer stream.Close()

    // 4. 读取流数据
    for {
        chunk, err := stream.Next()
        if err != nil {
            break
        }

        if chunk.Type == core.ChunkTypeProgress {
            fmt.Printf("Progress: %.1f%%\n", chunk.Metadata.Progress)
        }
    }
}
```

## 场景 2: HTTP SSE 流式接口 (1 分钟)

```go
package main

import (
    "context"
    "net/http"

    "github.com/kart-io/k8s-agent/pkg/agent/core"
    "github.com/kart-io/k8s-agent/pkg/agent/stream/agents"
    "github.com/kart-io/k8s-agent/pkg/agent/stream/tools"
)

func main() {
    agent := agents.NewProgressAgent(nil)

    // 创建 SSE 端点
    http.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
        input := &core.AgentInput{Task: "Stream task"}

        stream, _ := agent.ExecuteStream(r.Context(), input)
        defer stream.Close()

        tools.StreamToSSE(r.Context(), w, stream)
    })

    http.ListenAndServe(":8080", nil)
}
```

**测试**:
```bash
curl -N http://localhost:8080/stream
```

## 场景 3: WebSocket 双向流 (2 分钟)

```go
package main

import (
    "context"
    "log"
    "net/http"

    "github.com/kart-io/k8s-agent/pkg/agent/core"
    "github.com/kart-io/k8s-agent/pkg/agent/stream/agents"
    "github.com/kart-io/k8s-agent/pkg/agent/stream/tools"
)

func main() {
    agent := agents.NewProgressAgent(nil)

    // WebSocket 处理器
    handler := tools.WebSocketStreamHandler(
        func(ctx context.Context, input *core.AgentInput) (core.StreamOutput, error) {
            return agent.ExecuteStream(ctx, input)
        },
    )

    http.HandleFunc("/ws", handler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

**客户端测试 (JavaScript)**:
```html
<script>
const ws = new WebSocket('ws://localhost:8080/ws');
ws.onopen = () => {
    ws.send(JSON.stringify({task: 'test'}));
};
ws.onmessage = (e) => {
    const chunk = JSON.parse(e.data);
    console.log('Progress:', chunk.metadata.progress);
};
</script>
```

## 场景 4: 数据流处理 (2 分钟)

```go
package main

import (
    "context"
    "fmt"

    "github.com/kart-io/k8s-agent/pkg/agent/core"
    "github.com/kart-io/k8s-agent/pkg/agent/stream/agents"
)

func main() {
    agent := agents.NewDataPipelineAgent(nil)

    // 准备数据
    data := make([]interface{}, 1000)
    for i := 0; i < 1000; i++ {
        data[i] = i
    }

    // 使用转换函数
    transform := func(item interface{}) (interface{}, error) {
        num := item.(int)
        return num * 2, nil  // 每个数字乘以 2
    }

    stream, _ := agent.ProcessWithTransform(context.Background(), data, transform)
    defer stream.Close()

    // 收集结果
    for {
        chunk, err := stream.Next()
        if err != nil {
            break
        }
        fmt.Printf("Processed: %v\n", chunk.Data)
    }
}
```

## 场景 5: 使用中间件 (2 分钟)

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/kart-io/k8s-agent/pkg/agent/core"
    "github.com/kart-io/k8s-agent/pkg/agent/stream/agents"
    "github.com/kart-io/k8s-agent/pkg/agent/stream/middleware"
)

func main() {
    agent := agents.NewProgressAgent(nil)
    ctx := context.Background()

    // 1. 执行流式任务
    input := &core.AgentInput{Task: "test"}
    stream1, _ := agent.ExecuteStream(ctx, input)

    // 2. 应用过滤器 - 只保留进度更新
    filter := middleware.NewFilterMiddleware(func(chunk *core.StreamChunk) bool {
        return chunk.Type == core.ChunkTypeProgress
    })
    stream2, _ := filter.Apply(ctx, stream1)

    // 3. 应用限流 - 最多 10 块/秒
    throttle := middleware.NewThrottleMiddleware(10.0)
    stream3, _ := throttle.Apply(ctx, stream2)

    // 4. 消费最终流
    for {
        chunk, err := stream3.Next()
        if err != nil {
            break
        }
        fmt.Printf("Progress: %.1f%%\n", chunk.Metadata.Progress)
        time.Sleep(time.Second)
    }
}
```

## 常见模式

### 模式 1: 收集所有数据

```go
stream, _ := agent.ExecuteStream(ctx, input)
defer stream.Close()

// 方法 1: 手动收集
var results []interface{}
for {
    chunk, err := stream.Next()
    if err != nil {
        break
    }
    results = append(results, chunk.Data)
}

// 方法 2: 使用 Collect
reader := stream.(*stream.Reader)
chunks, err := reader.Collect()
```

### 模式 2: 只收集文本

```go
stream, _ := agent.ExecuteStream(ctx, input)
defer stream.Close()

reader := stream.(*stream.Reader)
text, err := reader.CollectText()
fmt.Println(text)
```

### 模式 3: 带超时的流

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

stream, err := agent.ExecuteStream(ctx, input)
if err != nil {
    log.Fatal(err)
}
defer stream.Close()

for {
    chunk, err := stream.Next()
    if err != nil {
        if ctx.Err() == context.DeadlineExceeded {
            log.Println("Timeout!")
        }
        break
    }
    // 处理 chunk
}
```

### 模式 4: 多个消费者

```go
multiplexer := stream.NewMultiplexer(nil)

// 添加消费者
consumer1 := &agents.SimpleStreamConsumer{
    OnChunkFunc: func(chunk *core.StreamChunk) error {
        fmt.Println("Consumer 1:", chunk.Type)
        return nil
    },
}
multiplexer.AddConsumer(consumer1)

consumer2 := &agents.SimpleStreamConsumer{
    OnChunkFunc: func(chunk *core.StreamChunk) error {
        fmt.Println("Consumer 2:", chunk.Type)
        return nil
    },
}
multiplexer.AddConsumer(consumer2)

// 启动多路复用
stream, _ := agent.ExecuteStream(ctx, input)
multiplexer.Start(ctx, stream)
```

## 运行示例

### 示例 1: LLM 流式对话

```bash
cd examples/streaming/llm_streaming
go run main.go
```

### 示例 2: 数据处理

```bash
cd examples/streaming/data_processing
go run main.go
```

### 示例 3: 进度跟踪

```bash
cd examples/streaming/progress_tracking
go run main.go
```

### 示例 4: HTTP 服务器

```bash
cd examples/streaming/real_time_monitoring
go run main.go

# 在浏览器访问
open http://localhost:8080
```

## 调试技巧

### 1. 打印所有块

```go
for {
    chunk, err := stream.Next()
    if err != nil {
        break
    }

    // 打印详细信息
    fmt.Printf("Type: %s, Data: %v, Progress: %.1f%%\n",
        chunk.Type, chunk.Data, chunk.Metadata.Progress)
}
```

### 2. 查看流状态

```go
reader := stream.(*stream.Reader)
status := reader.Status()
fmt.Printf("State: %s, Chunks: %d, Progress: %.1f%%\n",
    status.State, status.ChunksRead, status.Progress)
```

### 3. 启用详细日志

```go
import "log"

opts := core.DefaultStreamOptions()
// 自定义转换函数来记录日志
opts.TransformFunc = func(chunk *core.StreamChunk) (*core.StreamChunk, error) {
    log.Printf("[STREAM] Type: %s, Seq: %d", chunk.Type, chunk.Metadata.Sequence)
    return chunk, nil
}
```

## 性能建议

### 1. 调整缓冲区大小

```go
opts := core.DefaultStreamOptions()
opts.BufferSize = 1000  // 增加缓冲区
```

### 2. 启用背压

```go
opts.EnableBackpressure = true
opts.MaxPendingChunks = 5000
```

### 3. 批量处理

```go
middleware := middleware.NewBatchMiddleware(100, time.Second)
stream, _ := middleware.Apply(ctx, sourceStream)
```

## 常见问题

### Q: 如何停止流？

```go
// 方法 1: 关闭流
stream.Close()

// 方法 2: 取消上下文
ctx, cancel := context.WithCancel(context.Background())
stream, _ := agent.ExecuteStream(ctx, input)
cancel()  // 停止流
```

### Q: 如何处理错误？

```go
for {
    chunk, err := stream.Next()
    if err != nil {
        if err.Error() == "EOF" {
            // 正常结束
            break
        }
        // 错误处理
        log.Printf("Error: %v", err)
        break
    }

    // 检查错误块
    if chunk.Type == core.ChunkTypeError {
        log.Printf("Stream error: %v", chunk.Error)
    }
}
```

### Q: 如何实现心跳？

```go
ticker := time.NewTicker(30 * time.Second)
defer ticker.Stop()

for {
    select {
    case <-ticker.C:
        writer.WriteText("")  // 发送心跳
    default:
        // 正常处理
    }
}
```

## 下一步

- 阅读 [完整使用指南](AGENT_STREAMING_GUIDE.md)
- 查看 [实现总结](AGENT_STREAMING_IMPLEMENTATION_SUMMARY.md)
- 探索 [示例代码](../examples/streaming/)
- 了解 [性能优化](AGENT_STREAMING_GUIDE.md#性能优化)

## 获取帮助

- GitHub Issues: [提交问题](https://github.com/kart-io/k8s-agent/issues)
- 文档: [完整文档](AGENT_STREAMING_GUIDE.md)
- 示例: [examples/streaming/](../examples/streaming/)

祝你使用愉快！ 🚀
