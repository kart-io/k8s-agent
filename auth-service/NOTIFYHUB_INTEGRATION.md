# NotifyHub Integration Guide

本文档说明了 auth-service 中强制登出功能的通知系统集成方案。

## 概述

auth-service 的强制登出功能现已集成 **NotifyHub** 统一通知平台，替代了原先的自定义SMTP实现。NotifyHub 提供了更强大、更灵活的多平台通知能力。

## 架构变更

### 旧架构（已废弃）
```
notification/
├── email.go           # 自定义SMTP实现（已删除）
├── service.go         # 通知服务
├── repository.go      # 通知记录仓储
└── template.go        # 模板引擎
```

### 新架构（NotifyHub集成）
```
notification/
├── service.go         # 通知服务（使用NotifyHub客户端）
├── repository.go      # 通知记录仓储
└── template.go        # 模板引擎

依赖:
└── github.com/kart-io/notifyhub  # 统一通知平台
```

## 集成优势

### 1. **统一通知接口**
- 单一API入口，支持多种通知渠道
- 类型安全的消息构建
- 自动重试和错误处理

### 2. **多平台支持**
目前支持，未来可扩展：
- ✅ **Email** (SMTP)
- ✅ **飞书** (Feishu)
- ✅ **钉钉** (DingTalk)
- ✅ **企业微信** (WeChatWork)
- ✅ **Slack**
- ✅ **SMS**

### 3. **高性能**
- 异步消息处理
- Worker池并发发送
- 内置速率限制
- 连接池管理

### 4. **可观测性**
- 实时健康检查
- 详细的发送指标
- 结构化日志记录
- OpenTelemetry集成

## 使用方式

### 服务器初始化

```go
import (
    "github.com/kart-io/notifyhub/pkg/notifyhub"
    "github.com/kart-io/notifyhub/pkg/platforms/email"
)

// 初始化NotifyHub客户端（Email平台）
notifyHubClient, err := notifyhub.New(
    email.WithEmail(
        "smtp.example.com",      // SMTP主机
        587,                      // SMTP端口
        "user@example.com",       // SMTP用户名
        "password",               // SMTP密码
        "noreply@example.com",    // 发件人地址
        "System Notification",    // 发件人名称
    ),
    notifyhub.WithLogger(logger),
    notifyhub.WithTimeout(30*time.Second),
)

// 创建通知服务
notificationService := notification.NewService(
    notificationRepo,
    notifyHubClient,      // 使用NotifyHub客户端
    templateEngine,
)
```

### 发送通知

通知服务的接口保持不变，内部使用NotifyHub：

```go
// 异步发送通知
err := notificationService.NotifyUser(ctx, notification.NotifyUserParams{
    EventID:      eventID,
    UserID:       "user-123",
    EmailAddress: "user@example.com",
    Username:     "张三",
    Timestamp:    time.Now(),
    Reason:       "安全策略强制登出",
    DeviceInfo:   "Chrome on Windows",
    Location:     "深圳",
    ActorName:    "系统管理员",
    LoginURL:     "https://app.example.com/login",
})

// 同步发送（用于测试或关键通知）
err := notificationService.NotifyUserSync(ctx, params)

// 批量发送
errors := notificationService.NotifyMultipleUsers(ctx, paramsSlice)
```

## 配置说明

### 环境变量

```bash
# 通知开关
EMAIL_ENABLED=true

# SMTP配置（使用NotifyHub）
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=notifications@example.com
SMTP_PASSWORD=your-app-password
EMAIL_FROM_ADDRESS=noreply@k8s-agent.com
EMAIL_FROM_NAME="K8s Agent Security"

# 模板目录
EMAIL_TEMPLATE_DIR=templates/email
```

### config.yaml

```yaml
# Email notification configuration (NotifyHub integration)
email:
  enabled: true
  smtp_host: "smtp.gmail.com"
  smtp_port: 587
  smtp_user: "notifications@example.com"
  smtp_password: "${SMTP_PASSWORD}"
  from_address: "noreply@k8s-agent.com"
  from_name: "K8s Agent Security"
  template_dir: "templates/email"
```

## 测试模式

当 `EMAIL_ENABLED=false` 时，系统使用测试模式：

```go
// 测试模式下的NotifyHub初始化
notifyHubClient, err := notifyhub.New(
    notifyhub.WithTestDefaults(),  // 使用测试配置
    notifyhub.WithLogger(logger),
)
```

测试模式特性：
- ✅ 不发送真实邮件
- ✅ 返回成功Receipt（模拟）
- ✅ 所有API正常工作
- ✅ 适合开发和CI/CD环境

## 监控和健康检查

### 健康检查端点

```bash
# 检查服务整体健康状态
GET /health

Response:
{
  "status": "healthy",
  "service": "auth-service",
  "version": "1.0.0",
  "email_notifications": true
}
```

### NotifyHub健康检查

```go
// 检查NotifyHub客户端健康状态
health, err := notifyHubClient.Health(ctx)

fmt.Printf("Status: %s\n", health.Status)
for platform, status := range health.Platforms {
    fmt.Printf("Platform %s: Available=%v\n", platform, status.Available)
}
```

## 消息格式

NotifyHub支持多种消息格式：

```go
msg := &message.Message{
    ID:      "notification-001",
    Title:   "Security Alert: Session Terminated",
    Body:    renderedHTML,              // HTML格式邮件正文
    Format:  message.FormatHTML,        // 格式类型
    Targets: []target.Target{           // 目标列表
        {
            Type:     "email",
            Value:    "user@example.com",
            Platform: "email",
        },
    },
    Metadata: map[string]interface{}{   // 额外元数据
        "text_body": plainText,          // 纯文本备用
        "category":  "security-alert",
        "priority":  "high",
    },
}

receipt, err := notifyHubClient.Send(ctx, msg)
```

## 错误处理

NotifyHub提供详细的错误信息：

```go
receipt, err := notifyHubClient.Send(ctx, msg)
if err != nil {
    log.Printf("Send failed: %v", err)
    return
}

// 检查每个平台的发送结果
for _, result := range receipt.Results {
    if !result.Success {
        log.Printf("Platform %s failed: %s", result.Platform, result.Error)
    } else {
        log.Printf("Platform %s success: MessageID=%s",
            result.Platform, result.MessageID)
    }
}
```

## 未来扩展

### 添加飞书通知

```go
import "github.com/kart-io/notifyhub/pkg/platforms/feishu"

notifyHubClient, err := notifyhub.New(
    // Email平台
    email.WithEmail(...),

    // 飞书平台
    feishu.WithFeishu(
        "https://open.feishu.cn/open-apis/bot/v2/hook/xxx",
        "secret-key",
    ),

    notifyhub.WithLogger(logger),
)

// 同时发送到Email和飞书
msg.Targets = []target.Target{
    {Type: "email", Value: "user@example.com", Platform: "email"},
    {Type: "feishu", Value: "#security-alerts", Platform: "feishu"},
}
```

### 添加SMS通知

```go
import "github.com/kart-io/notifyhub/pkg/platforms/sms"

notifyHubClient, err := notifyhub.New(
    email.WithEmail(...),
    sms.WithSMS("twilio", "api-key", "api-secret"),
    notifyhub.WithLogger(logger),
)
```

## 性能优化

### 异步发送（推荐）

```go
// 使用NotifyHub的异步发送
asyncReceipt, err := notifyHubClient.SendAsync(ctx, msg)
if err != nil {
    return err
}

// 立即返回，不阻塞主流程
log.Printf("Message queued: %s", asyncReceipt.MessageID)
```

### 批量发送

```go
messages := []*message.Message{msg1, msg2, msg3}
receipts, err := notifyHubClient.SendBatch(ctx, messages)

for i, receipt := range receipts {
    log.Printf("Message %d: Status=%s", i, receipt.Status)
}
```

## 日志示例

启动日志：
```
=== Auth Service Started ===
Listening on: :8090
NotifyHub Integration: ✅ Enabled
Email Platform: smtp.gmail.com:587 (from: noreply@k8s-agent.com)

Registered Routes:
GET    /health                              - Health check
POST   /api/v1/auth/login                   - User login
GET    /api/v1/sessions/users/:userId       - List user sessions
POST   /api/v1/forced-logout/session/:jti   - Force logout single session
============================
```

通知发送日志（由NotifyHub提供）：
```
[INFO] NotifyHub: Message queued notification-001
[INFO] NotifyHub: Sending to email: user@example.com
[INFO] NotifyHub: Email sent successfully MessageID=abc123
[INFO] NotifyHub: Receipt status=delivered
```

## 故障排查

### 问题1: 邮件未发送

**检查步骤：**
1. 确认 `EMAIL_ENABLED=true`
2. 验证SMTP凭据
3. 检查NotifyHub日志
4. 调用健康检查接口

```bash
# 查看NotifyHub健康状态
curl http://localhost:8090/health

# 查看通知记录表
SELECT * FROM forced_logout_notifications WHERE status='failed';
```

### 问题2: 连接超时

```go
// 增加超时时间
notifyHubClient, err := notifyhub.New(
    email.WithEmail(...),
    notifyhub.WithTimeout(60*time.Second),  // 增加到60秒
)
```

### 问题3: 速率限制

NotifyHub内置速率限制，可配置：

```go
notifyHubClient, err := notifyhub.New(
    email.WithEmail(...),
    notifyhub.WithRateLimit(100, 10),  // 100 tokens, 10/sec refill
)
```

## 相关资源

- **NotifyHub 项目**: https://github.com/kart-io/notifyhub
- **NotifyHub 文档**: `/path/to/notifyhub/README.md`
- **Email平台文档**: `/path/to/notifyhub/pkg/platforms/email/README.md`
- **强制登出规范**: `.specify/features/001-auth-service/spec.md`

## 迁移清单

从旧SMTP实现迁移到NotifyHub的步骤：

- [x] 删除自定义SMTP实现 (`email.go`)
- [x] 更新 `notification/service.go` 使用NotifyHub客户端
- [x] 更新 `cmd/server/main.go` 初始化NotifyHub
- [x] 保持模板引擎不变（兼容）
- [x] 保持通知记录仓储不变（兼容）
- [x] 更新配置说明文档
- [x] 测试Email发送功能
- [ ] 添加集成测试
- [ ] 添加性能基准测试

## 总结

通过集成NotifyHub，auth-service的通知系统获得了：

✅ **更强大的功能** - 多平台支持、异步处理、自动重试
✅ **更好的性能** - Worker池、连接池、速率限制
✅ **更易维护** - 统一接口、结构化日志、健康检查
✅ **更好的扩展性** - 轻松添加新通知渠道

NotifyHub作为统一通知平台，为未来的多渠道通知需求提供了坚实基础。
