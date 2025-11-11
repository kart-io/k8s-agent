# Agent 框架流式输出完整指南

本文档提供 Agent 框架流式输出功能的完整介绍和使用指南。

## 目录

1. [概述](#概述)
2. [核心概念](#核心概念)
3. [架构设计](#架构设计)
4. [快速开始](#快速开始)
5. [核心组件](#核心组件)
6. [流式 Agent](#流式-agent)
7. [中间件](#中间件)
8. [协议支持](#协议支持)
9. [示例](#示例)
10. [性能优化](#性能优化)
11. [最佳实践](#最佳实践)

## 概述

Agent 框架的流式输出功能提供了高效、灵活的实时数据传输能力：

### 核心特性

- **实时响应**: LLM 逐字输出，首字符延迟从秒级降到毫秒级
- **内存高效**: 处理 GB 级数据而不会 OOM，使用流而非缓存
- **进度跟踪**: 实时进度更新、ETA 计算、阶段报告
- **多协议支持**: SSE、WebSocket、HTTP Chunked Transfer、gRPC Stream
- **背压控制**: 防止生产者过载消费者
- **流式转换**: Filter、Map、Reduce 等流操作
- **多路复用**: 单个流支持多个消费者
- **错误恢复**: 自动重试、指数退避

### 使用场景

- LLM 对话的流式响应
- 大数据集的渐进式处理
- 长时间运行任务的实时进度
- 实时监控和告警
- 日志流式分析
- 视频/音频流处理

## 核心概念

### StreamingAgent

扩展了 `Agent` 接口，添加流式输出能力：

```go
type StreamingAgent interface {
    Agent
    ExecuteStream(ctx context.Context, input *AgentInput) (StreamOutput, error)
}
```

### StreamOutput

流式输出接口，提供逐块读取能力：

```go
type StreamOutput interface {
    Next() (*StreamChunk, error)  // 读取下一个数据块
    Close() error                  // 关闭流
    IsClosed() bool                // 检查是否已关闭
    Context() context.Context      // 获取上下文
}
```

### StreamChunk

流数据块，表示流中的一个数据单元：

```go
type StreamChunk struct {
    Type     ChunkType      // 数据类型 (text, json, binary, progress, etc.)
    Data     interface{}    // 实际数据
    Text     string         // 文本数据 (仅 Type=text 时)
    Metadata ChunkMetadata  // 元数据 (序列号、时间戳、进度等)
    IsLast   bool           // 是否是最后一个块
    Error    error          // 错误信息
}
```

### ChunkType

数据块类型：

- `ChunkTypeText`: 文本数据
- `ChunkTypeJSON`: JSON 数据
- `ChunkTypeBinary`: 二进制数据
- `ChunkTypeProgress`: 进度更新
- `ChunkTypeStatus`: 状态更新
- `ChunkTypeError`: 错误信息
- `ChunkTypeMetadata`: 元数据
- `ChunkTypeControl`: 控制命令

## 架构设计

### 三层架构

```
┌─────────────────────────────────────────────┐
│         Application Layer (应用层)           │
│  - StreamingAgent                           │
│  - StreamConsumer                           │
│  - Application Logic                        │
└─────────────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────┐
│       Infrastructure Layer (基础设施层)      │
│  - Writer/Reader                            │
│  - RingBuffer                               │
│  - Multiplexer                              │
└─────────────────────────────────────────────┘
                     ↓
┌─────────────────────────────────────────────┐
│         Protocol Layer (协议层)              │
│  - SSE                                      │
│  - WebSocket                                │
│  - Chunked Transfer                         │
│  - gRPC Stream                              │
└─────────────────────────────────────────────┘
```

### 数据流向

```
Producer (Agent)
      ↓
 StreamWriter
      ↓
  Channel (buffered)
      ↓
 StreamReader
      ↓
Middleware (optional)
      ↓
Protocol Adapter
      ↓
Consumer (Client)
```

## 快速开始

### 基本流式输出

```go
// 1. 创建流式 Agent
agent := agents.NewStreamingLLMAgent(llmClient, config)

// 2. 准备输入
input := &core.AgentInput{
    Task: "Explain streaming benefits",
}

// 3. 执行流式任务
ctx := context.Background()
stream, err := agent.ExecuteStream(ctx, input)
if err != nil {
    log.Fatal(err)
}
defer stream.Close()

// 4. 读取流数据
for {
    chunk, err := stream.Next()
    if err == io.EOF {
        break
    }
    if err != nil {
        log.Fatal(err)
    }

    // 处理数据块
    if chunk.Type == core.ChunkTypeText {
        fmt.Print(chunk.Text)  // 实时输出
    }
}
```

### HTTP SSE 流式接口

```go
func SSEHandler(agent core.StreamingAgent) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 解析输入
        var input core.AgentInput
        json.NewDecoder(r.Body).Decode(&input)

        // 执行流式任务
        stream, err := agent.ExecuteStream(r.Context(), &input)
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }
        defer stream.Close()

        // 转换为 SSE
        tools.StreamToSSE(r.Context(), w, stream)
    }
}
```

## 核心组件

### Writer (流写入器)

Agent 通过 Writer 向流中写入数据：

```go
writer := stream.NewWriter(ctx, opts)

// 写入文本
writer.WriteText("Hello, world!")

// 写入进度
writer.WriteProgress(50.0, "Half complete")

// 写入自定义数据块
chunk := core.NewStreamChunk(core.ChunkTypeJSON, data)
writer.WriteChunk(chunk)

// 关闭写入器
writer.Close()
```

### Reader (流读取器)

Consumer 通过 Reader 从流中读取数据：

```go
reader := stream.NewReader(ctx, channel, opts)

// 逐块读取
for {
    chunk, err := reader.Next()
    if err == io.EOF {
        break
    }
    // 处理 chunk
}

// 或者收集所有数据
chunks, err := reader.Collect()

// 或者只收集文本
text, err := reader.CollectText()
```

### RingBuffer (环形缓冲区)

高效的固定大小缓冲：

```go
buffer := stream.NewRingBuffer(100)

// 添加元素
buffer.Push(chunk)

// 弹出元素
chunk := buffer.Pop()

// 检查状态
isEmpty := buffer.IsEmpty()
isFull := buffer.IsFull()
count := buffer.Count()

// 调整大小
buffer.Resize(200)
```

### Multiplexer (多路复用器)

支持多个消费者同时读取同一个流：

```go
multiplexer := stream.NewMultiplexer(opts)

// 添加消费者
consumer1 := &MyConsumer{}
id1, _ := multiplexer.AddConsumer(consumer1)

consumer2 := &MyConsumer{}
id2, _ := multiplexer.AddConsumer(consumer2)

// 启动多路复用
multiplexer.Start(ctx, sourceStream)

// 移除消费者
multiplexer.RemoveConsumer(id1)

// 关闭多路复用器
multiplexer.Close()
```

## 流式 Agent

### StreamingLLMAgent

LLM 流式响应 Agent：

```go
config := &agents.StreamingLLMConfig{
    ChunkSize:        10,                    // 每次发送字符数
    ChunkDelay:       50 * time.Millisecond, // 块间延迟
    EnableProgress:   true,                  // 启用进度报告
    ProgressInterval: time.Second,           // 进度更新间隔
}

agent := agents.NewStreamingLLMAgent(llmClient, config)

stream, err := agent.ExecuteStream(ctx, input)
```

**特性**：
- 逐字返回生成的文本
- 实时进度更新
- 降低首字符延迟
- 模拟打字效果

### DataPipelineAgent

数据流处理 Agent：

```go
config := &agents.DataPipelineConfig{
    BatchSize:        100,                   // 批次大小
    ProcessDelay:     100 * time.Millisecond,// 处理延迟
    EnableProgress:   true,                  // 启用进度
    MaxWorkers:       4,                     // 最大工作协程数
}

agent := agents.NewDataPipelineAgent(config)

// 基本流式处理
stream, err := agent.ExecuteStream(ctx, input)

// 使用转换函数
transform := func(item interface{}) (interface{}, error) {
    // 转换逻辑
    return transformedItem, nil
}
stream, err := agent.ProcessWithTransform(ctx, dataSource, transform)
```

**特性**：
- 分批处理大数据集
- 避免一次性加载所有数据
- 实时返回处理结果
- 支持 Filter、Map、Reduce 操作

### ProgressAgent

带进度反馈的 Agent：

```go
config := &agents.ProgressConfig{
    EnableProgress:   true,
    ProgressInterval: time.Second,
    EnableETA:        true,  // 启用 ETA 计算
    EnablePhases:     true,  // 启用阶段报告
}

agent := agents.NewProgressAgent(config)

stream, err := agent.ExecuteStream(ctx, input)
```

**特性**：
- 实时进度更新 (0-100%)
- ETA 预估
- 阶段性状态报告
- 吞吐量统计

## 中间件

### BufferMiddleware (缓冲中间件)

动态调整缓冲区大小：

```go
middleware := middleware.NewBufferMiddleware(
    minSize, maxSize, threshold,
)

stream, err := middleware.Apply(ctx, sourceStream)
```

### ThrottleMiddleware (限流中间件)

控制流的速率：

```go
middleware := middleware.NewThrottleMiddleware(
    100.0, // 最大 100 块/秒
)

stream, err := middleware.Apply(ctx, sourceStream)
```

### TransformMiddleware (转换中间件)

转换流中的数据：

```go
transform := func(chunk *core.StreamChunk) (*core.StreamChunk, error) {
    // 修改 chunk
    chunk.Data = processedData
    return chunk, nil
}

middleware := middleware.NewTransformMiddleware(transform)
stream, err := middleware.Apply(ctx, sourceStream)
```

### FilterMiddleware (过滤中间件)

过滤流中的数据块：

```go
predicate := func(chunk *core.StreamChunk) bool {
    return chunk.Type == core.ChunkTypeProgress
}

middleware := middleware.NewFilterMiddleware(predicate)
stream, err := middleware.Apply(ctx, sourceStream)
```

### BatchMiddleware (批处理中间件)

聚合多个块成批次：

```go
middleware := middleware.NewBatchMiddleware(
    batchSize,
    timeout,
)

stream, err := middleware.Apply(ctx, sourceStream)
```

### TeeMiddleware (分支中间件)

复制流到多个目标：

```go
consumer1 := &MyConsumer{}
consumer2 := &MyConsumer{}

middleware := middleware.NewTeeMiddleware(consumer1, consumer2)
stream, err := middleware.Apply(ctx, sourceStream)
```

## 协议支持

### Server-Sent Events (SSE)

单向服务器推送，适合实时更新：

```go
// 服务器端
streamer, _ := tools.NewSSEStreamer(w)
streamer.WriteText("Hello")
streamer.WriteProgress(50, "Half done")
streamer.Close()

// 或使用辅助函数
tools.StreamToSSE(ctx, w, sourceStream)

// HTTP 处理器
handler := tools.SSEHandler(func(ctx context.Context, input *core.AgentInput) (core.StreamOutput, error) {
    return agent.ExecuteStream(ctx, input)
})
```

**客户端示例 (JavaScript)**:

```javascript
const eventSource = new EventSource('/api/stream/sse');

eventSource.addEventListener('text', (e) => {
    const chunk = JSON.parse(e.data);
    console.log(chunk.text);
});

eventSource.addEventListener('progress', (e) => {
    const chunk = JSON.parse(e.data);
    console.log('Progress:', chunk.metadata.progress);
});

eventSource.onerror = () => {
    eventSource.close();
};
```

### WebSocket

双向实时通信：

```go
// 服务器端
streamer := tools.NewWebSocketStreamer(conn)
streamer.WriteText("Hello")
streamer.WriteProgress(50, "Half done")

// 读取客户端消息
chunk, err := streamer.ReadChunk()

// 双向流
bidirectional := tools.NewWebSocketBidirectionalStream(conn)
input := <-bidirectional.Input()
bidirectional.Output() <- outputChunk

// HTTP 处理器
handler := tools.WebSocketStreamHandler(func(ctx context.Context, input *core.AgentInput) (core.StreamOutput, error) {
    return agent.ExecuteStream(ctx, input)
})
```

**客户端示例 (JavaScript)**:

```javascript
const ws = new WebSocket('ws://localhost:8080/api/stream/ws');

ws.onopen = () => {
    // 发送输入
    ws.send(JSON.stringify({task: 'test'}));
};

ws.onmessage = (e) => {
    const chunk = JSON.parse(e.data);
    if (chunk.type === 'text') {
        console.log(chunk.text);
    }
};
```

### HTTP Chunked Transfer

标准 HTTP 流式传输：

```go
// 服务器端
streamer, _ := tools.NewChunkedTransferStreamer(w)
streamer.WriteChunk(chunk)
streamer.Close()

// 或使用辅助函数
tools.StreamToChunkedTransfer(ctx, w, sourceStream)
```

**客户端示例 (JavaScript)**:

```javascript
fetch('/api/stream/chunked', {method: 'POST'})
    .then(response => {
        const reader = response.body.getReader();
        const decoder = new TextDecoder();

        function read() {
            reader.read().then(({done, value}) => {
                if (done) return;
                const text = decoder.decode(value);
                console.log(text);
                read();
            });
        }

        read();
    });
```

### gRPC Stream (TODO)

高性能二进制流：

```go
// 服务器端 (TODO: 实现)
// stream.Send(&pb.Chunk{...})

// 客户端 (TODO: 实现)
// for {
//     chunk, err := stream.Recv()
//     if err == io.EOF { break }
// }
```

## 示例

### 示例 1: LLM 流式对话

```bash
cd examples/streaming/llm_streaming
go run main.go
```

**演示**：
- LLM 逐字输出
- 实时进度更新
- 模拟打字效果
- 流消费者使用

### 示例 2: 大数据流处理

```bash
cd examples/streaming/data_processing
go run main.go
```

**演示**：
- 分批处理 500 个数据项
- 实时进度和吞吐量
- Filter、Map、Reduce 操作
- 流式转换

### 示例 3: 进度跟踪

```bash
cd examples/streaming/progress_tracking
go run main.go
```

**演示**：
- 实时进度条
- ETA 计算
- 阶段报告
- 多任务并发跟踪

### 示例 4: 实时监控服务器

```bash
cd examples/streaming/real_time_monitoring
go run main.go
```

然后访问 http://localhost:8080

**演示**：
- SSE 流式接口
- WebSocket 双向通信
- Chunked Transfer
- 交互式 Web 界面

## 性能优化

### 1. 零拷贝优化

使用环形缓冲区避免内存分配：

```go
buffer := stream.NewRingBuffer(1000)
// 预分配，避免运行时分配
```

### 2. 背压处理

防止生产者过载消费者：

```go
opts := core.DefaultStreamOptions()
opts.EnableBackpressure = true
opts.BackpressureWindow = 10
opts.MaxPendingChunks = 1000
```

### 3. 内存池

复用 StreamChunk 对象：

```go
var chunkPool = sync.Pool{
    New: func() interface{} {
        return &core.StreamChunk{}
    },
}

chunk := chunkPool.Get().(*core.StreamChunk)
// 使用 chunk
chunkPool.Put(chunk)
```

### 4. 批量处理

减少系统调用和网络往返：

```go
middleware := middleware.NewBatchMiddleware(100, time.Second)
stream, _ := middleware.Apply(ctx, sourceStream)
```

### 5. 并发处理

使用多个 worker 并行处理：

```go
config := &agents.DataPipelineConfig{
    MaxWorkers: runtime.NumCPU(),
}
```

## 最佳实践

### 1. 错误处理

```go
for {
    chunk, err := stream.Next()
    if err == io.EOF {
        // 正常结束
        break
    }
    if err != nil {
        // 错误处理
        log.Printf("Stream error: %v", err)

        // 发送错误通知
        if chunk.Type == core.ChunkTypeError {
            notifyError(chunk.Error)
        }
        break
    }

    // 处理成功的 chunk
}
```

### 2. 资源清理

```go
stream, err := agent.ExecuteStream(ctx, input)
if err != nil {
    return err
}
defer stream.Close() // 确保资源被释放
```

### 3. 超时控制

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

stream, err := agent.ExecuteStream(ctx, input)
```

### 4. 进度报告

```go
opts := core.DefaultStreamOptions()
opts.EnableProgress = true
opts.ProgressInterval = time.Second

writer := stream.NewWriter(ctx, opts)

tracker := agents.NewProgressTracker(totalItems, writer, config)
for i := 0; i < totalItems; i++ {
    // 处理项目
    tracker.Increment(1)
}
tracker.Complete()
```

### 5. 流的组合

```go
// 原始流
stream1, _ := agent.ExecuteStream(ctx, input)

// 应用过滤器
stream2, _ := filterMiddleware.Apply(ctx, stream1)

// 应用转换
stream3, _ := transformMiddleware.Apply(ctx, stream2)

// 应用限流
stream4, _ := throttleMiddleware.Apply(ctx, stream3)

// 最终消费
for chunk := range stream4.Next() {
    // 处理
}
```

### 6. 监控和调试

```go
reader := stream.NewReader(ctx, channel, opts)

// 定期检查流状态
ticker := time.NewTicker(5 * time.Second)
go func() {
    for range ticker.C {
        status := reader.Status()
        log.Printf("Stream status: state=%s, chunks=%d, bytes=%d, progress=%.1f%%",
            status.State, status.ChunksRead, status.BytesRead, status.Progress)
    }
}()
```

### 7. 多路复用

```go
// 创建多个消费者
logConsumer := &LoggingConsumer{}
metricsConsumer := &MetricsConsumer{}
storageConsumer := &StorageConsumer{}

// 设置多路复用
multiplexer := stream.NewMultiplexer(opts)
multiplexer.AddConsumer(logConsumer)
multiplexer.AddConsumer(metricsConsumer)
multiplexer.AddConsumer(storageConsumer)

// 启动
multiplexer.Start(ctx, sourceStream)
```

## 性能指标

### 吞吐量

- **文本流**: > 10MB/s
- **JSON 流**: > 1000 chunks/s
- **二进制流**: > 50MB/s

### 延迟

- **首字符延迟**: < 10ms (LLM 流式)
- **块传输延迟**: < 1ms (本地)
- **端到端延迟**: < 50ms (SSE)

### 内存使用

- **环形缓冲区**: 固定大小，O(1) 操作
- **流处理**: 常量内存，不随数据大小增长
- **多路复用**: 每个消费者独立缓冲

## 故障排查

### 问题 1: 流卡住不动

**原因**: 缓冲区满，背压未启用

**解决**:
```go
opts.EnableBackpressure = true
opts.MaxPendingChunks = 1000
```

### 问题 2: 内存持续增长

**原因**: 未正确关闭流或泄漏

**解决**:
```go
defer stream.Close()
// 确保所有 goroutine 退出
```

### 问题 3: 进度不更新

**原因**: 进度间隔太长

**解决**:
```go
opts.ProgressInterval = 100 * time.Millisecond
```

### 问题 4: SSE 连接断开

**原因**: 超时或网络问题

**解决**:
```go
// 发送心跳
ticker := time.NewTicker(30 * time.Second)
go func() {
    for range ticker.C {
        streamer.WriteText("") // 心跳
    }
}()
```

## 未来改进

- [ ] gRPC 流支持
- [ ] 压缩支持 (gzip, brotli)
- [ ] 加密流 (TLS)
- [ ] 流的持久化和重放
- [ ] 更多性能优化
- [ ] 流式测试工具

## 相关资源

- [Agent 核心文档](../agent/README.md)
- [LLM 客户端文档](../llm/README.md)
- [示例代码](../../examples/streaming/)
- [性能测试](../../pkg/agent/performance/)

## 贡献

欢迎提交问题和 Pull Request！

## 许可证

MIT License
