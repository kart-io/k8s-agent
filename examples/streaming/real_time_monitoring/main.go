package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/kart-io/k8s-agent/pkg/agent/core"
	"github.com/kart-io/k8s-agent/pkg/agent/stream"
)

func main() {
	fmt.Println("=== Real-Time Monitoring Server ===")
	fmt.Println()

	// 创建进度 Agent
	progressAgent := stream.NewProgressAgent(stream.DefaultProgressConfig())

	// 创建数据管道 Agent
	pipelineAgent := stream.NewDataPipelineAgent(stream.DefaultDataPipelineConfig())

	// 创建路由
	router := mux.NewRouter()

	// SSE 端点
	router.HandleFunc("/api/stream/sse/progress", SSEProgressHandler(progressAgent)).Methods("POST")
	router.HandleFunc("/api/stream/sse/data", SSEDataProcessingHandler(pipelineAgent)).Methods("POST")

	// WebSocket 端点
	router.HandleFunc("/api/stream/ws/progress", stream.WebSocketStreamHandler(func(ctx context.Context, input *core.AgentInput) (core.StreamOutput, error) {
		return progressAgent.ExecuteStream(ctx, input)
	})).Methods("GET")

	router.HandleFunc("/api/stream/ws/data", stream.WebSocketStreamHandler(func(ctx context.Context, input *core.AgentInput) (core.StreamOutput, error) {
		return pipelineAgent.ExecuteStream(ctx, input)
	})).Methods("GET")

	// Chunked Transfer 端点
	router.HandleFunc("/api/stream/chunked/progress", ChunkedProgressHandler(progressAgent)).Methods("POST")

	// 健康检查
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	}).Methods("GET")

	// 主页
	router.HandleFunc("/", HomeHandler).Methods("GET")

	// 启动服务器
	addr := ":8080"
	fmt.Printf("Server starting on %s\n", addr)
	fmt.Println("\nAvailable endpoints:")
	fmt.Println("  GET  /                             - Home page with examples")
	fmt.Println("  GET  /health                        - Health check")
	fmt.Println("  POST /api/stream/sse/progress      - SSE progress stream")
	fmt.Println("  POST /api/stream/sse/data          - SSE data processing stream")
	fmt.Println("  GET  /api/stream/ws/progress       - WebSocket progress stream")
	fmt.Println("  GET  /api/stream/ws/data           - WebSocket data stream")
	fmt.Println("  POST /api/stream/chunked/progress  - Chunked transfer progress stream")
	fmt.Println("\nPress Ctrl+C to stop")
	fmt.Println()

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// SSEProgressHandler SSE 进度流处理器
func SSEProgressHandler(agent *stream.ProgressAgent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// 解析输入
		var input core.AgentInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// 设置默认值
		if input.Context == nil {
			input.Context = make(map[string]interface{})
		}
		if _, ok := input.Context["total_steps"]; !ok {
			input.Context["total_steps"] = 100
		}
		if _, ok := input.Context["step_duration"]; !ok {
			input.Context["step_duration"] = 100 * time.Millisecond
		}
		input.Timestamp = time.Now()

		// 执行流式任务
		source, err := agent.ExecuteStream(ctx, &input)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer source.Close()

		// 转换为 SSE
		if err := stream.StreamToSSE(ctx, w, source); err != nil {
			log.Printf("SSE error: %v", err)
		}
	}
}

// SSEDataProcessingHandler SSE 数据处理流处理器
func SSEDataProcessingHandler(agent *stream.DataPipelineAgent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// 解析输入
		var input core.AgentInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// 生成示例数据
		dataSize := 200
		if size, ok := input.Context["data_size"].(float64); ok {
			dataSize = int(size)
		}

		dataSource := make([]interface{}, dataSize)
		for i := 0; i < dataSize; i++ {
			dataSource[i] = map[string]interface{}{
				"id":    i,
				"value": fmt.Sprintf("item-%d", i),
			}
		}

		input.Context["data_source"] = dataSource
		input.Timestamp = time.Now()

		// 执行流式任务
		source, err := agent.ExecuteStream(ctx, &input)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer source.Close()

		// 转换为 SSE
		if err := stream.StreamToSSE(ctx, w, source); err != nil {
			log.Printf("SSE error: %v", err)
		}
	}
}

// ChunkedProgressHandler Chunked Transfer 进度处理器
func ChunkedProgressHandler(agent *stream.ProgressAgent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// 解析输入
		var input core.AgentInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// 设置默认值
		if input.Context == nil {
			input.Context = make(map[string]interface{})
		}
		if _, ok := input.Context["total_steps"]; !ok {
			input.Context["total_steps"] = 50
		}
		if _, ok := input.Context["step_duration"]; !ok {
			input.Context["step_duration"] = 200 * time.Millisecond
		}
		input.Timestamp = time.Now()

		// 执行流式任务
		source, err := agent.ExecuteStream(ctx, &input)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer source.Close()

		// 转换为 Chunked Transfer
		if err := stream.StreamToChunkedTransfer(ctx, w, source); err != nil {
			log.Printf("Chunked transfer error: %v", err)
		}
	}
}

// HomeHandler 主页处理器
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>Real-Time Monitoring Server</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f5f5f5;
        }
        .container {
            background-color: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        h1 {
            color: #333;
            border-bottom: 2px solid #4CAF50;
            padding-bottom: 10px;
        }
        h2 {
            color: #666;
            margin-top: 30px;
        }
        .endpoint {
            background-color: #f9f9f9;
            padding: 15px;
            margin: 10px 0;
            border-left: 4px solid #4CAF50;
            border-radius: 4px;
        }
        .method {
            display: inline-block;
            padding: 2px 8px;
            border-radius: 3px;
            font-weight: bold;
            margin-right: 10px;
        }
        .get { background-color: #61affe; color: white; }
        .post { background-color: #49cc90; color: white; }
        code {
            background-color: #f4f4f4;
            padding: 2px 6px;
            border-radius: 3px;
            font-family: monospace;
        }
        .example {
            background-color: #fffbf0;
            padding: 15px;
            margin: 10px 0;
            border-radius: 4px;
            border: 1px solid #ffe58f;
        }
        button {
            background-color: #4CAF50;
            color: white;
            padding: 10px 20px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            margin: 5px;
        }
        button:hover {
            background-color: #45a049;
        }
        #output {
            background-color: #1e1e1e;
            color: #d4d4d4;
            padding: 20px;
            border-radius: 4px;
            margin-top: 20px;
            font-family: monospace;
            white-space: pre-wrap;
            max-height: 400px;
            overflow-y: auto;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 Real-Time Monitoring Server</h1>
        <p>This server demonstrates streaming capabilities using SSE, WebSocket, and Chunked Transfer Encoding.</p>

        <h2>📡 Available Endpoints</h2>

        <div class="endpoint">
            <span class="method post">POST</span>
            <code>/api/stream/sse/progress</code>
            <p>Server-Sent Events (SSE) stream with real-time progress updates.</p>
        </div>

        <div class="endpoint">
            <span class="method post">POST</span>
            <code>/api/stream/sse/data</code>
            <p>SSE stream for data processing with batch updates.</p>
        </div>

        <div class="endpoint">
            <span class="method get">GET</span>
            <code>/api/stream/ws/progress</code>
            <p>WebSocket stream with bidirectional communication support.</p>
        </div>

        <div class="endpoint">
            <span class="method post">POST</span>
            <code>/api/stream/chunked/progress</code>
            <p>HTTP Chunked Transfer Encoding for streaming JSON data.</p>
        </div>

        <h2>🧪 Try It Out</h2>
        <button onclick="testSSEProgress()">Test SSE Progress</button>
        <button onclick="testSSEData()">Test SSE Data Processing</button>
        <button onclick="testChunked()">Test Chunked Transfer</button>
        <button onclick="clearOutput()">Clear Output</button>

        <div id="output"></div>

        <div class="example">
            <h3>cURL Examples:</h3>
            <pre>
# SSE Progress Stream
curl -N -X POST http://localhost:8080/api/stream/sse/progress \
  -H "Content-Type: application/json" \
  -d '{"task": "test", "context": {"total_steps": 50}}'

# SSE Data Processing
curl -N -X POST http://localhost:8080/api/stream/sse/data \
  -H "Content-Type: application/json" \
  -d '{"task": "process", "context": {"data_size": 100}}'

# Chunked Transfer
curl -N -X POST http://localhost:8080/api/stream/chunked/progress \
  -H "Content-Type: application/json" \
  -d '{"task": "test"}'
            </pre>
        </div>
    </div>

    <script>
        function appendOutput(text) {
            const output = document.getElementById('output');
            output.textContent += text + '\n';
            output.scrollTop = output.scrollHeight;
        }

        function clearOutput() {
            document.getElementById('output').textContent = '';
        }

        function testSSEProgress() {
            clearOutput();
            appendOutput('=== Testing SSE Progress Stream ===\n');

            const eventSource = new EventSource('/api/stream/sse/progress?task=test&total_steps=30');

            eventSource.addEventListener('start', (e) => {
                appendOutput('[START] ' + e.data);
            });

            eventSource.addEventListener('progress', (e) => {
                const data = JSON.parse(e.data);
                appendOutput('[PROGRESS] ' + data.metadata.progress.toFixed(1) + '%');
            });

            eventSource.addEventListener('status', (e) => {
                const data = JSON.parse(e.data);
                appendOutput('[STATUS] ' + data.data);
            });

            eventSource.addEventListener('close', (e) => {
                appendOutput('[CLOSE] Stream ended');
                eventSource.close();
            });

            eventSource.onerror = (e) => {
                appendOutput('[ERROR] Connection error');
                eventSource.close();
            };
        }

        function testSSEData() {
            clearOutput();
            appendOutput('=== Testing SSE Data Processing ===\n');

            fetch('/api/stream/sse/data', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({task: 'process', context: {data_size: 100}})
            }).then(response => {
                const reader = response.body.getReader();
                const decoder = new TextDecoder();

                function read() {
                    reader.read().then(({done, value}) => {
                        if (done) {
                            appendOutput('\n[COMPLETE] Stream ended');
                            return;
                        }

                        const text = decoder.decode(value, {stream: true});
                        const lines = text.split('\n');

                        for (const line of lines) {
                            if (line.startsWith('data: ')) {
                                try {
                                    const data = JSON.parse(line.substring(6));
                                    if (data.type === 'progress') {
                                        appendOutput('[PROGRESS] ' + data.metadata.progress.toFixed(1) + '%');
                                    } else if (data.type === 'json') {
                                        appendOutput('[DATA] Batch processed');
                                    }
                                } catch (e) {}
                            }
                        }

                        read();
                    });
                }

                read();
            });
        }

        function testChunked() {
            clearOutput();
            appendOutput('=== Testing Chunked Transfer ===\n');

            fetch('/api/stream/chunked/progress', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({task: 'test', context: {total_steps: 20}})
            }).then(response => {
                const reader = response.body.getReader();
                const decoder = new TextDecoder();

                function read() {
                    reader.read().then(({done, value}) => {
                        if (done) {
                            appendOutput('\n[COMPLETE] Stream ended');
                            return;
                        }

                        const text = decoder.decode(value, {stream: true});
                        const lines = text.split('\n').filter(l => l.trim());

                        for (const line of lines) {
                            try {
                                const chunk = JSON.parse(line);
                                if (chunk.type === 'progress') {
                                    appendOutput('[PROGRESS] ' + chunk.metadata.progress.toFixed(1) + '%');
                                }
                            } catch (e) {}
                        }

                        read();
                    });
                }

                read();
            });
        }
    </script>
</body>
</html>
`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}
