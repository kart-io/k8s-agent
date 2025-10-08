# Agent-Manager 日志分析报告

## 🔍 问题诊断

### 症状
`make run` 后 agent-manager **一直在输出日志**，可能感觉像在"刷屏"。

### 根本原因

#### 1. **日志级别设置为 DEBUG** ⚠️
**位置**: `configs/config.yaml:46`
```yaml
logging:
  level: "debug"  # ← 问题所在
```

**影响**:
- DEBUG 级别会输出所有详细日志
- 包括每次心跳检查、缓存操作、数据库查询等
- 适合开发调试，但会产生大量日志

#### 2. **后台定时任务正常运行** ✅
Agent-manager 有两个后台任务持续运行：

**a. heartbeatMonitor (心跳监控)**
- **频率**: 每 30 秒
- **位置**: `internal/agent/registry.go:294-308`
- **作用**: 检查所有 agent 的心跳是否超时
- **日志输出**:
  ```go
  r.logger.Warn("Agent heartbeat timeout", ...)  // 如果有超时
  r.logger.Error("Failed to update agent status", ...) // 如果更新失败
  ```

**b. cleanupStaleAgents (清理过期 agent)**
- **频率**: 每 30 秒
- **位置**: `internal/agent/registry.go:343-357`
- **作用**: 删除离线超过 24 小时的 agent
- **日志输出**:
  ```go
  r.logger.Info("Cleaning up stale agent", ...) // 清理时
  ```

#### 3. **NATS 连接重连机制**
**配置**: `configs/config.yaml:15`
```yaml
nats:
  max_reconnect: -1  # 无限重试
  reconnect_wait: 2s
```

如果 NATS 连接不稳定，会产生重连日志。

## 📊 正常 vs 异常判断

### ✅ 正常情况

**启动时**:
```json
{"level":"info","msg":"Starting Aetherius Agent Manager"}
{"level":"info","msg":"Initializing PostgreSQL storage"}
{"level":"info","msg":"Initializing Redis cache"}
{"level":"info","msg":"Initializing agent registry"}
{"level":"info","msg":"Starting agent registry"}
{"level":"info","msg":"Loaded agents from database","count":0}
{"level":"info","msg":"Initializing NATS server"}
{"level":"info","msg":"Initializing API server"}
{"level":"info","msg":"API server starting","addr":"0.0.0.0:8080"}
```

**运行时** (每 30 秒):
- 如果有 agent 连接，会有心跳日志
- 如果日志级别是 debug，会有大量调试信息

### ⚠️ 异常情况

**高频错误日志**:
```json
{"level":"error","msg":"Failed to connect to database"}
{"level":"error","msg":"Failed to connect to NATS"}
{"level":"error","msg":"Failed to update agent status"}
```

**无限重连**:
```json
{"level":"warn","msg":"NATS connection lost, reconnecting..."}
{"level":"warn","msg":"NATS connection lost, reconnecting..."}
{"level":"warn","msg":"NATS connection lost, reconnecting..."}
```

## 🔧 解决方案

### 方案 1: 降低日志级别（推荐）

修改 `configs/config.yaml`:
```yaml
logging:
  level: "info"  # 从 debug 改为 info
  format: "json"
  output_path: "stdout"
```

**效果**:
- ✅ 只显示重要信息、警告和错误
- ✅ 过滤掉调试细节
- ✅ 日志量减少 70-80%

**对比**:
| 级别 | 输出内容 | 适用场景 |
|------|---------|---------|
| debug | 所有日志（非常详细） | 开发调试 |
| info | 重要操作 + 警告 + 错误 | **生产环境（推荐）** |
| warn | 警告 + 错误 | 只关注问题 |
| error | 只有错误 | 最小化日志 |

### 方案 2: 改为更友好的格式

修改 `configs/config.yaml`:
```yaml
logging:
  level: "info"
  format: "console"  # 从 json 改为 console
  output_path: "stdout"
```

**效果**:
- ✅ 彩色输出，更易读
- ✅ 格式更简洁
- ✅ 适合本地开发

**示例输出**:
```
2025-10-04T00:13:13.095+0800    INFO    agent-registry  Loaded agents from database    count=0
2025-10-04T00:13:13.095+0800    INFO    nats-server     Connected to NATS       url=nats://localhost:4222
```

### 方案 3: 输出到文件

修改 `configs/config.yaml`:
```yaml
logging:
  level: "info"
  format: "json"
  output_path: "/tmp/agent-manager.log"  # 输出到文件
```

**效果**:
- ✅ 终端干净
- ✅ 日志持久化
- ✅ 可以用 `tail -f` 查看

**查看日志**:
```bash
# 实时查看
tail -f /tmp/agent-manager.log | jq .

# 只看错误
grep '"level":"error"' /tmp/agent-manager.log | jq .

# 统计日志
wc -l /tmp/agent-manager.log
```

### 方案 4: 调整定时任务频率（可选）

如果心跳检查太频繁，可以修改代码：

**位置**: `internal/agent/registry.go:45-46`
```go
heartbeatTimeout: 60 * time.Second,  // 当前 60 秒
cleanupInterval:  30 * time.Second,  // 当前 30 秒
```

**改为**:
```go
heartbeatTimeout: 120 * time.Second,  // 改为 2 分钟
cleanupInterval:  300 * time.Second,  // 改为 5 分钟
```

**心跳监控定时器**: `internal/agent/registry.go:297`
```go
ticker := time.NewTicker(30 * time.Second)  // 当前 30 秒
```

**改为**:
```go
ticker := time.NewTicker(60 * time.Second)  // 改为 60 秒
```

## 🎯 推荐配置

### 开发环境
```yaml
logging:
  level: "info"      # 重要信息
  format: "console"  # 彩色输出
  output_path: "stdout"
```

### 生产环境
```yaml
logging:
  level: "info"      # 重要信息
  format: "json"     # 结构化日志
  output_path: "/var/log/agent-manager/app.log"  # 持久化
```

### 调试问题时
```yaml
logging:
  level: "debug"     # 详细日志
  format: "console"  # 易读格式
  output_path: "stdout"
```

## 📈 日志优化检查清单

- [ ] 修改日志级别为 `info`
- [ ] 选择合适的日志格式
- [ ] 考虑是否输出到文件
- [ ] 检查是否有高频错误日志
- [ ] 验证 NATS 连接稳定性
- [ ] 确认 PostgreSQL 连接正常
- [ ] 确认 Redis 连接正常

## 🔍 日志分析技巧

### 查找高频日志
```bash
# 统计每种消息出现次数
make run 2>&1 | grep -o '"msg":"[^"]*"' | sort | uniq -c | sort -rn | head -10
```

### 监控特定组件
```bash
# 只看 agent-registry 的日志
make run 2>&1 | grep '"component":"agent-registry"'

# 只看错误
make run 2>&1 | grep '"level":"error"'

# 只看警告和错误
make run 2>&1 | grep -E '"level":"(error|warn)"'
```

### 实时查看并美化
```bash
# 需要安装 jq
make run 2>&1 | jq -r '"\(.timestamp) [\(.level)] \(.component // "main"): \(.msg)"'
```

## ✅ 验证修改

修改配置后重启：
```bash
cd agent-manager
make run
```

**预期结果**:
- 启动时只有 6-8 行初始化日志
- 运行时安静（除非有实际操作）
- 每 30-60 秒可能有一次心跳检查（如果有 agent）
