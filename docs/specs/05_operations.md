# 运维安全文档 (Operations & Security)

## 文档信息

- **版本**: v1.6
- **最后更新**: 2025-09-28
- **状态**: 正式版
- **所属系统**: Aetherius AI Agent
- **文档类型**: 运维安全指南

## 1. 概述

### 1.1 文档目标

本文档提供 Aetherius AI Agent 的运维和安全管理指南,涵盖:

- **安全架构**: 身份认证、授权、数据保护、审计合规
- **运维管理**: 监控告警、日志管理、性能调优、故障排查
- **成本管理**: AI服务费用控制、资源优化
- **灾难恢复**: 备份策略、故障转移、业务连续性
- **最佳实践**: 安全加固、性能优化、运维自动化

### 1.2 目标受众

| 角色 | 关注重点 | 建议阅读章节 |
|------|----------|--------------|
| **安全工程师** | 身份认证、权限控制、合规审计 | 2, 3, 4 |
| **运维工程师 (SRE)** | 监控告警、故障排查、性能优化 | 5, 6, 7 |
| **系统管理员** | 日常维护、备份恢复、升级管理 | 5, 8, 9 |
| **架构师** | 整体安全架构、灾难恢复规划 | 全部 |

## 2. 安全架构

### 2.1 身份认证 (Authentication)

#### 2.1.1 多层身份认证架构

```go
type AuthenticationConfig struct {
    // OIDC 配置 (用户身份认证)
    OIDC OIDCConfig `yaml:"oidc"`

    // 服务间认证 (mTLS)
    ServiceAuth ServiceAuthConfig `yaml:"service_auth"`

    // API 密钥管理
    APIKeys APIKeyConfig `yaml:"api_keys"`

    // JWT 配置
    JWT JWTConfig `yaml:"jwt"`
}

type OIDCConfig struct {
    Enabled       bool     `yaml:"enabled"`
    IssuerURL     string   `yaml:"issuer_url"`
    ClientID      string   `yaml:"client_id"`
    ClientSecret  string   `yaml:"client_secret"`
    RedirectURL   string   `yaml:"redirect_url"`
    Scopes        []string `yaml:"scopes"`
    UsernameClaim string   `yaml:"username_claim"`
    GroupsClaim   string   `yaml:"groups_claim"`
}

type JWTConfig struct {
    SigningKey    string        `yaml:"signing_key"`
    TokenExpiry   time.Duration `yaml:"token_expiry"`
    RefreshExpiry time.Duration `yaml:"refresh_expiry"`
    Issuer        string        `yaml:"issuer"`
    Audience      []string      `yaml:"audience"`
}
```

#### 2.1.2 OIDC 集成示例

```go
type OIDCAuthenticator struct {
    provider *oidc.Provider
    verifier *oidc.IDTokenVerifier
    config   oauth2.Config
}

func NewOIDCAuthenticator(cfg OIDCConfig) (*OIDCAuthenticator, error) {
    ctx := context.Background()

    provider, err := oidc.NewProvider(ctx, cfg.IssuerURL)
    if err != nil {
        return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
    }

    oauth2Config := oauth2.Config{
        ClientID:     cfg.ClientID,
        ClientSecret: cfg.ClientSecret,
        RedirectURL:  cfg.RedirectURL,
        Endpoint:     provider.Endpoint(),
        Scopes:       cfg.Scopes,
    }

    verifier := provider.Verifier(&oidc.Config{
        ClientID: cfg.ClientID,
    })

    return &OIDCAuthenticator{
        provider: provider,
        verifier: verifier,
        config:   oauth2Config,
    }, nil
}

func (a *OIDCAuthenticator) VerifyToken(ctx context.Context, rawToken string) (*UserInfo, error) {
    idToken, err := a.verifier.Verify(ctx, rawToken)
    if err != nil {
        return nil, fmt.Errorf("token verification failed: %w", err)
    }

    var claims struct {
        Email          string   `json:"email"`
        EmailVerified  bool     `json:"email_verified"`
        Name           string   `json:"name"`
        PreferredUsername string `json:"preferred_username"`
        Groups         []string `json:"groups"`
    }

    if err := idToken.Claims(&claims); err != nil {
        return nil, fmt.Errorf("failed to parse claims: %w", err)
    }

    return &UserInfo{
        ID:       idToken.Subject,
        Email:    claims.Email,
        Username: claims.PreferredUsername,
        Groups:   claims.Groups,
        Verified: claims.EmailVerified,
    }, nil
}
```

#### 2.1.3 服务间 mTLS 认证

```yaml
# ConfigMap: mTLS 证书配置
apiVersion: v1
kind: ConfigMap
metadata:
  name: mtls-config
  namespace: aetherius
data:
  ca.crt: |
    -----BEGIN CERTIFICATE-----
    ...
    -----END CERTIFICATE-----
---
apiVersion: v1
kind: Secret
metadata:
  name: orchestrator-tls
  namespace: aetherius
type: kubernetes.io/tls
data:
  tls.crt: <base64-encoded-cert>
  tls.key: <base64-encoded-key>
  ca.crt: <base64-encoded-ca>
```

```go
// mTLS 客户端配置
func createMTLSClient(certFile, keyFile, caFile string) (*http.Client, error) {
    cert, err := tls.LoadX509KeyPair(certFile, keyFile)
    if err != nil {
        return nil, fmt.Errorf("failed to load client cert: %w", err)
    }

    caCert, err := os.ReadFile(caFile)
    if err != nil {
        return nil, fmt.Errorf("failed to load CA cert: %w", err)
    }

    caCertPool := x509.NewCertPool()
    caCertPool.AppendCertsFromPEM(caCert)

    tlsConfig := &tls.Config{
        Certificates: []tls.Certificate{cert},
        RootCAs:      caCertPool,
        MinVersion:   tls.VersionTLS12,
    }

    return &http.Client{
        Transport: &http.Transport{
            TLSClientConfig: tlsConfig,
        },
        Timeout: 30 * time.Second,
    }, nil
}
```

### 2.2 授权与访问控制 (Authorization)

#### 2.2.1 RBAC 权限模型

```yaml
# ClusterRole: 只读诊断权限
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: aetherius-reader
rules:
# 核心资源只读
- apiGroups: [""]
  resources: ["pods", "services", "endpoints", "configmaps", "secrets"]
  verbs: ["get", "list", "watch"]

# 工作负载资源只读
- apiGroups: ["apps"]
  resources: ["deployments", "replicasets", "daemonsets", "statefulsets"]
  verbs: ["get", "list", "watch"]

# 批处理资源只读
- apiGroups: ["batch"]
  resources: ["jobs", "cronjobs"]
  verbs: ["get", "list", "watch"]

# 网络资源只读
- apiGroups: ["networking.k8s.io"]
  resources: ["networkpolicies", "ingresses"]
  verbs: ["get", "list", "watch"]

# 事件读取
- apiGroups: [""]
  resources: ["events"]
  verbs: ["get", "list", "watch"]

# 指标读取
- apiGroups: ["metrics.k8s.io"]
  resources: ["pods", "nodes"]
  verbs: ["get", "list"]

# 明确禁止写操作
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["create", "update", "patch", "delete", "deletecollection"]
  resourceNames: []
---
# ClusterRoleBinding: 绑定服务账户
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: aetherius-reader-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: aetherius-reader
subjects:
- kind: ServiceAccount
  name: aetherius-service-account
  namespace: aetherius
```

#### 2.2.2 应用层 RBAC 实现

```go
type RBACManager struct {
    policies map[string][]Permission
    cache    *cache.Cache
}

type Permission struct {
    Resource string   `json:"resource"`
    Actions  []string `json:"actions"`
    Effect   string   `json:"effect"` // allow/deny
}

type Policy struct {
    Name        string       `json:"name"`
    Description string       `json:"description"`
    Permissions []Permission `json:"permissions"`
}

func (r *RBACManager) CheckPermission(user *UserInfo, resource, action string) bool {
    cacheKey := fmt.Sprintf("perm:%s:%s:%s", user.ID, resource, action)

    if cached, found := r.cache.Get(cacheKey); found {
        return cached.(bool)
    }

    allowed := false
    for _, role := range user.Roles {
        permissions, exists := r.policies[role]
        if !exists {
            continue
        }

        for _, perm := range permissions {
            if !r.matchResource(perm.Resource, resource) {
                continue
            }

            if r.containsAction(perm.Actions, action) {
                if perm.Effect == "deny" {
                    r.cache.Set(cacheKey, false, 5*time.Minute)
                    return false
                }
                allowed = true
            }
        }
    }

    r.cache.Set(cacheKey, allowed, 5*time.Minute)
    return allowed
}

func (r *RBACManager) matchResource(pattern, resource string) bool {
    if pattern == "*" {
        return true
    }

    matched, _ := filepath.Match(pattern, resource)
    return matched
}
```

#### 2.2.3 权限检查中间件

```go
func (m *AuthMiddleware) RequirePermission(resource, action string) gin.HandlerFunc {
    return func(c *gin.Context) {
        user, exists := c.Get("user")
        if !exists {
            c.AbortWithStatusJSON(401, gin.H{
                "error": "unauthorized",
                "message": "authentication required",
            })
            return
        }

        userInfo := user.(*UserInfo)

        if !m.rbac.CheckPermission(userInfo, resource, action) {
            m.auditLogger.LogDenied(userInfo.ID, resource, action)
            c.AbortWithStatusJSON(403, gin.H{
                "error": "forbidden",
                "message": fmt.Sprintf("user does not have permission to %s %s", action, resource),
            })
            return
        }

        m.auditLogger.LogAllowed(userInfo.ID, resource, action)
        c.Next()
    }
}
```

### 2.3 数据保护与加密

#### 2.3.1 传输层安全 (TLS)

```yaml
# Ingress TLS 配置
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: aetherius-ingress
  namespace: aetherius
  annotations:
    # 强制 HTTPS
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"

    # TLS 版本控制
    nginx.ingress.kubernetes.io/ssl-protocols: "TLSv1.2 TLSv1.3"
    nginx.ingress.kubernetes.io/ssl-ciphers: "ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256"

    # HSTS
    nginx.ingress.kubernetes.io/hsts: "true"
    nginx.ingress.kubernetes.io/hsts-max-age: "31536000"
    nginx.ingress.kubernetes.io/hsts-include-subdomains: "true"

    # 证书管理器
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - aetherius.example.com
    secretName: aetherius-tls
  rules:
  - host: aetherius.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: aetherius-orchestrator
            port:
              number: 80
```

#### 2.3.2 静态数据加密

```go
// 数据库字段加密
type EncryptedField struct {
    data          []byte
    encryptionKey []byte
}

func (e *EncryptedField) Encrypt(plaintext string) error {
    block, err := aes.NewCipher(e.encryptionKey)
    if err != nil {
        return fmt.Errorf("failed to create cipher: %w", err)
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return fmt.Errorf("failed to create GCM: %w", err)
    }

    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return fmt.Errorf("failed to generate nonce: %w", err)
    }

    e.data = gcm.Seal(nonce, nonce, []byte(plaintext), nil)
    return nil
}

func (e *EncryptedField) Decrypt() (string, error) {
    block, err := aes.NewCipher(e.encryptionKey)
    if err != nil {
        return "", fmt.Errorf("failed to create cipher: %w", err)
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", fmt.Errorf("failed to create GCM: %w", err)
    }

    nonceSize := gcm.NonceSize()
    if len(e.data) < nonceSize {
        return "", fmt.Errorf("ciphertext too short")
    }

    nonce, ciphertext := e.data[:nonceSize], e.data[nonceSize:]
    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return "", fmt.Errorf("failed to decrypt: %w", err)
    }

    return string(plaintext), nil
}
```

#### 2.3.3 密钥管理 (Vault 集成)

```go
type VaultClient struct {
    client      *api.Client
    authMethod  auth.AuthMethod
    tokenPath   string
    mountPath   string
}

func NewVaultClient(cfg VaultConfig) (*VaultClient, error) {
    config := api.DefaultConfig()
    config.Address = cfg.Address

    client, err := api.NewClient(config)
    if err != nil {
        return nil, fmt.Errorf("failed to create vault client: %w", err)
    }

    // Kubernetes 认证
    k8sAuth, err := auth.NewKubernetesAuth(
        cfg.RoleName,
        auth.WithServiceAccountTokenPath(cfg.SATokenPath),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create k8s auth: %w", err)
    }

    authInfo, err := client.Auth().Login(context.Background(), k8sAuth)
    if err != nil {
        return nil, fmt.Errorf("failed to login: %w", err)
    }

    if authInfo == nil {
        return nil, fmt.Errorf("no auth info returned")
    }

    return &VaultClient{
        client:     client,
        authMethod: k8sAuth,
        mountPath:  cfg.MountPath,
    }, nil
}

func (v *VaultClient) GetK8sCredentials(clusterID string) (*K8sCredentials, error) {
    path := fmt.Sprintf("%s/kubernetes/%s/credentials", v.mountPath, clusterID)

    secret, err := v.client.Logical().Read(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read credentials: %w", err)
    }

    if secret == nil {
        return nil, fmt.Errorf("no credentials found for cluster %s", clusterID)
    }

    data := secret.Data["data"].(map[string]interface{})

    return &K8sCredentials{
        Token:     data["token"].(string),
        CA:        data["ca"].(string),
        Endpoint:  data["endpoint"].(string),
        ExpiresAt: time.Now().Add(1 * time.Hour), // 短期凭证
    }, nil
}

func (v *VaultClient) RenewToken(ctx context.Context) error {
    ticker := time.NewTicker(30 * time.Minute)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            secret, err := v.client.Auth().Token().RenewSelf(3600)
            if err != nil {
                return fmt.Errorf("failed to renew token: %w", err)
            }

            log.Info("Token renewed successfully",
                zap.Duration("ttl", time.Duration(secret.Auth.LeaseDuration)*time.Second))
        }
    }
}
```

### 2.4 输入验证与防护

#### 2.4.1 输入验证框架

```go
type InputValidator struct {
    validator *validator.Validate
}

type Alert struct {
    Name      string            `json:"name" validate:"required,min=1,max=255"`
    Severity  string            `json:"severity" validate:"required,oneof=critical high medium low"`
    ClusterID string            `json:"cluster_id" validate:"required,uuid4"`
    Namespace string            `json:"namespace" validate:"required,dns1123label,max=63"`
    Pod       string            `json:"pod" validate:"omitempty,dns1123label,max=253"`
    Message   string            `json:"message" validate:"max=4096"`
    Labels    map[string]string `json:"labels" validate:"dive,keys,dns1123label,endkeys,max=256"`
}

func NewInputValidator() *InputValidator {
    v := validator.New()

    v.RegisterValidation("dns1123label", validateDNS1123Label)
    v.RegisterValidation("k8s_name", validateK8sName)

    return &InputValidator{
        validator: v,
    }
}

func (iv *InputValidator) ValidateAlert(alert *Alert) error {
    if err := iv.validator.Struct(alert); err != nil {
        validationErrors := err.(validator.ValidationErrors)
        return fmt.Errorf("validation failed: %s", iv.formatErrors(validationErrors))
    }

    // 业务逻辑验证
    if err := iv.validateClusterAccess(alert.ClusterID); err != nil {
        return fmt.Errorf("cluster access denied: %w", err)
    }

    if err := iv.validateNamespaceAccess(alert.ClusterID, alert.Namespace); err != nil {
        return fmt.Errorf("namespace access denied: %w", err)
    }

    return nil
}

func validateDNS1123Label(fl validator.FieldLevel) bool {
    value := fl.Field().String()
    regex := regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
    return regex.MatchString(value)
}
```

#### 2.4.2 命令注入防护

```go
type SecureExecutor struct {
    allowedCommands map[string]CommandSpec
    sandbox         SandboxConfig
}

type CommandSpec struct {
    Command     string            `yaml:"command"`
    Args        []string          `yaml:"args"`
    AllowedArgs map[string]string `yaml:"allowed_args"` // key: arg name, value: regex pattern
    Timeout     time.Duration     `yaml:"timeout"`
    MaxOutput   int64             `yaml:"max_output"`
}

func (e *SecureExecutor) Execute(ctx context.Context, toolID string, params map[string]string) (*ExecResult, error) {
    spec, exists := e.allowedCommands[toolID]
    if !exists {
        return nil, fmt.Errorf("command not allowed: %s", toolID)
    }

    // 验证所有参数
    for key, value := range params {
        pattern, exists := spec.AllowedArgs[key]
        if !exists {
            return nil, fmt.Errorf("parameter not allowed: %s", key)
        }

        matched, err := regexp.MatchString(pattern, value)
        if err != nil || !matched {
            return nil, fmt.Errorf("parameter value invalid: %s=%s", key, value)
        }
    }

    // 构建安全命令
    args := make([]string, len(spec.Args))
    copy(args, spec.Args)

    // 替换参数占位符
    for i, arg := range args {
        if strings.HasPrefix(arg, "${") && strings.HasSuffix(arg, "}") {
            paramName := strings.TrimSuffix(strings.TrimPrefix(arg, "${"), "}")
            if value, exists := params[paramName]; exists {
                args[i] = value
            }
        }
    }

    cmd := exec.CommandContext(ctx, spec.Command, args...)

    // 清空环境变量(安全沙箱)
    cmd.Env = []string{}

    // 设置工作目录
    cmd.Dir = e.sandbox.WorkDir

    // 限制资源
    cmd.SysProcAttr = &syscall.SysProcAttr{
        // 限制进程优先级
        Setpriority: 10,
    }

    // 执行命令
    return e.executeWithLimits(cmd, spec)
}

func (e *SecureExecutor) executeWithLimits(cmd *exec.Cmd, spec CommandSpec) (*ExecResult, error) {
    var stdout, stderr bytes.Buffer
    cmd.Stdout = io.LimitReader(&stdout, spec.MaxOutput)
    cmd.Stderr = io.LimitReader(&stderr, spec.MaxOutput)

    ctx, cancel := context.WithTimeout(context.Background(), spec.Timeout)
    defer cancel()

    cmd = exec.CommandContext(ctx, cmd.Path, cmd.Args[1:]...)

    start := time.Now()
    err := cmd.Run()
    duration := time.Since(start)

    result := &ExecResult{
        Stdout:   stdout.String(),
        Stderr:   stderr.String(),
        ExitCode: cmd.ProcessState.ExitCode(),
        Duration: duration,
    }

    if err != nil && ctx.Err() == context.DeadlineExceeded {
        result.Error = "execution timeout"
        return result, fmt.Errorf("command timeout after %v", spec.Timeout)
    }

    if err != nil {
        result.Error = err.Error()
        return result, fmt.Errorf("command failed: %w", err)
    }

    return result, nil
}
```

### 2.5 审计与合规

#### 2.5.1 审计日志记录

```go
type AuditLogger struct {
    logger *zap.Logger
    sink   AuditSink
}

type AuditEvent struct {
    Timestamp     time.Time         `json:"timestamp"`
    EventID       string            `json:"event_id"`
    EventType     string            `json:"event_type"`
    UserID        string            `json:"user_id"`
    Username      string            `json:"username"`
    SessionID     string            `json:"session_id"`
    IPAddress     string            `json:"ip_address"`
    UserAgent     string            `json:"user_agent"`
    Resource      string            `json:"resource"`
    Action        string            `json:"action"`
    ClusterID     string            `json:"cluster_id,omitempty"`
    Namespace     string            `json:"namespace,omitempty"`
    Result        string            `json:"result"`
    Reason        string            `json:"reason,omitempty"`
    Duration      time.Duration     `json:"duration_ms"`
    RequestBody   map[string]interface{} `json:"request_body,omitempty"`
    ResponseCode  int               `json:"response_code"`
}

func (a *AuditLogger) LogEvent(event AuditEvent) {
    a.logger.Info("audit_event",
        zap.String("event_id", event.EventID),
        zap.String("event_type", event.EventType),
        zap.Time("timestamp", event.Timestamp),
        zap.String("user_id", event.UserID),
        zap.String("username", event.Username),
        zap.String("action", event.Action),
        zap.String("resource", event.Resource),
        zap.String("cluster_id", event.ClusterID),
        zap.String("result", event.Result),
        zap.Duration("duration_ms", event.Duration),
        zap.Int("response_code", event.ResponseCode),
    )

    // 发送到审计存储后端
    if err := a.sink.Write(event); err != nil {
        a.logger.Error("failed to write audit log", zap.Error(err))
    }
}

func (a *AuditLogger) AuditMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()

        // 生成事件 ID
        eventID := uuid.New().String()
        c.Set("audit_event_id", eventID)

        // 提取用户信息
        userInfo := extractUserInfo(c)

        // 记录请求信息
        var requestBody map[string]interface{}
        if c.Request.Body != nil {
            bodyBytes, _ := io.ReadAll(c.Request.Body)
            c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
            json.Unmarshal(bodyBytes, &requestBody)
        }

        // 执行请求
        c.Next()

        // 记录审计事件
        duration := time.Since(start)
        event := AuditEvent{
            Timestamp:    start,
            EventID:      eventID,
            EventType:    "api_request",
            UserID:       userInfo.ID,
            Username:     userInfo.Username,
            IPAddress:    c.ClientIP(),
            UserAgent:    c.Request.UserAgent(),
            Resource:     c.Request.URL.Path,
            Action:       c.Request.Method,
            Result:       determineResult(c.Writer.Status()),
            Duration:     duration,
            RequestBody:  requestBody,
            ResponseCode: c.Writer.Status(),
        }

        a.LogEvent(event)
    }
}
```

#### 2.5.2 合规性检查

```go
// GDPR 数据保护合规
type GDPRComplianceChecker struct {
    db                *sql.DB
    retentionPolicies map[string]time.Duration
}

func (g *GDPRComplianceChecker) CheckDataRetention(ctx context.Context) error {
    cutoffDate := time.Now().AddDate(0, 0, -g.retentionPolicies["user_feedback"].Days())

    // 检查用户反馈数据保留期限
    query := `
        SELECT COUNT(*) FROM user_feedback
        WHERE created_at < $1
    `

    var count int
    if err := g.db.QueryRowContext(ctx, query, cutoffDate).Scan(&count); err != nil {
        return fmt.Errorf("failed to check retention: %w", err)
    }

    if count > 0 {
        log.Warn("Found expired user feedback data",
            zap.Int("count", count),
            zap.Time("cutoff", cutoffDate))

        // 归档或删除过期数据
        if err := g.archiveExpiredData(ctx, cutoffDate); err != nil {
            return fmt.Errorf("failed to archive data: %w", err)
        }
    }

    return nil
}

func (g *GDPRComplianceChecker) AnonymizePersonalData(ctx context.Context, userID string) error {
    tx, err := g.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback()

    // 匿名化用户反馈
    _, err = tx.ExecContext(ctx, `
        UPDATE user_feedback
        SET user_id = 'anonymized',
            comments = 'redacted',
            updated_at = NOW()
        WHERE user_id = $1
    `, userID)
    if err != nil {
        return fmt.Errorf("failed to anonymize feedback: %w", err)
    }

    // 删除审计日志中的个人信息
    _, err = tx.ExecContext(ctx, `
        UPDATE audit_log
        SET username = 'anonymized',
            ip_address = '0.0.0.0',
            user_agent = 'redacted'
        WHERE user_id = $1
    `, userID)
    if err != nil {
        return fmt.Errorf("failed to anonymize audit log: %w", err)
    }

    if err := tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    log.Info("Personal data anonymized", zap.String("user_id", userID))
    return nil
}

// SOC2 合规检查
type SOC2ComplianceChecker struct {
    controls []SecurityControl
}

type SecurityControl struct {
    ID          string
    Name        string
    Description string
    CheckFunc   func(context.Context) error
}

func (s *SOC2ComplianceChecker) ValidateSecurityControls(ctx context.Context) error {
    results := make(map[string]error)

    for _, control := range s.controls {
        if err := control.CheckFunc(ctx); err != nil {
            results[control.ID] = err
            log.Error("Security control failed",
                zap.String("control_id", control.ID),
                zap.String("control_name", control.Name),
                zap.Error(err))
        } else {
            log.Info("Security control passed",
                zap.String("control_id", control.ID),
                zap.String("control_name", control.Name))
        }
    }

    if len(results) > 0 {
        return fmt.Errorf("SOC2 compliance check failed: %d controls failed", len(results))
    }

    return nil
}
```

## 3. 威胁建模与风险评估

### 3.1 威胁识别矩阵

| 威胁ID | 威胁类别 | 威胁描述 | 影响等级 | 可能性 | 风险评级 | 缓解措施 |
|--------|----------|----------|----------|--------|----------|----------|
| T001 | 身份冒充 | 未授权用户访问系统 | 高 | 中 | **高** | 多因子认证、RBAC、会话管理 |
| T002 | 数据泄露 | 敏感诊断信息暴露 | 高 | 低 | 中 | 端到端加密、访问控制、审计 |
| T003 | 命令注入 | 恶意命令执行 | 高 | 低 | 中 | 命令白名单、参数验证、沙箱 |
| T004 | 权限提升 | 获取超出授权的权限 | 中 | 低 | 低 | 最小权限原则、审计日志 |
| T005 | DoS 攻击 | 系统服务不可用 | 中 | 中 | 中 | 限流、熔断、监控告警 |
| T006 | AI 污染 | 恶意知识库注入 | 中 | 低 | 低 | 知识库审核、置信度评分 |
| T007 | 凭证泄露 | K8s 凭证被窃取 | 高 | 低 | 中 | 短期凭证、Vault 动态管理 |
| T008 | 中间人攻击 | 网络通信被劫持 | 中 | 低 | 低 | mTLS、证书验证 |

### 3.2 安全扫描集成

```yaml
# GitHub Actions: 安全扫描工作流
name: Security Scan
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]
  schedule:
    - cron: '0 0 * * 0' # 每周日午夜运行

jobs:
  security-scan:
    runs-on: ubuntu-latest
    permissions:
      security-events: write

    steps:
    - uses: actions/checkout@v3

    # 静态代码安全扫描 (gosec)
    - name: Run gosec
      run: |
        go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
        gosec -fmt sarif -out gosec-results.sarif ./...

    - name: Upload gosec results
      uses: github/codeql-action/upload-sarif@v2
      with:
        sarif_file: gosec-results.sarif

    # 依赖漏洞扫描 (govulncheck)
    - name: Run govulncheck
      run: |
        go install golang.org/x/vuln/cmd/govulncheck@latest
        govulncheck -json ./... > govulncheck-results.json

    # 容器镜像安全扫描 (Trivy)
    - name: Build Docker image
      run: |
        docker build -t aetherius/orchestrator:${{ github.sha }} .

    - name: Run Trivy
      uses: aquasecurity/trivy-action@master
      with:
        image-ref: 'aetherius/orchestrator:${{ github.sha }}'
        format: 'sarif'
        output: 'trivy-results.sarif'
        severity: 'CRITICAL,HIGH'

    - name: Upload Trivy results
      uses: github/codeql-action/upload-sarif@v2
      with:
        sarif_file: trivy-results.sarif

    # Kubernetes 配置安全检查 (Polaris)
    - name: Run Polaris
      uses: fairwindsops/polaris@master
      with:
        audit-path: ./k8s-manifests/

    # 密钥泄露扫描 (Gitleaks)
    - name: Run Gitleaks
      uses: gitleaks/gitleaks-action@v2
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

    # OWASP Dependency Check
    - name: OWASP Dependency Check
      uses: dependency-check/Dependency-Check_Action@main
      with:
        project: 'aetherius'
        path: '.'
        format: 'ALL'

    - name: Upload Dependency Check results
      uses: actions/upload-artifact@v3
      with:
        name: dependency-check-report
        path: reports/
```

## 4. 监控与可观测性

### 4.1 监控指标体系

#### 4.1.1 业务指标定义

```go
// Prometheus 指标定义
var (
    // 任务执行指标
    TasksTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "aetherius_tasks_total",
            Help: "Total number of diagnostic tasks",
        },
        []string{"status", "priority", "cluster"},
    )

    TaskDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "aetherius_task_duration_seconds",
            Help: "Task execution duration in seconds",
            Buckets: []float64{1, 5, 10, 30, 60, 300, 600},
        },
        []string{"status", "cluster"},
    )

    // AI 服务指标
    AITokenUsage = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "aetherius_ai_tokens_total",
            Help: "Total AI tokens consumed",
        },
        []string{"model", "task_type"},
    )

    AILatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "aetherius_ai_latency_seconds",
            Help: "AI service latency in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"model", "operation"},
    )

    AICost = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "aetherius_ai_cost_usd_total",
            Help: "Total AI service cost in USD",
        },
        []string{"model"},
    )

    // 知识库指标
    KnowledgeBaseHits = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "aetherius_kb_hits_total",
            Help: "Total knowledge base hits",
        },
        []string{"category", "result"},
    )

    // 系统资源指标
    QueueDepth = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "aetherius_queue_depth",
            Help: "Current task queue depth",
        },
        []string{"priority"},
    )
)

func init() {
    prometheus.MustRegister(
        TasksTotal,
        TaskDuration,
        AITokenUsage,
        AILatency,
        AICost,
        KnowledgeBaseHits,
        QueueDepth,
    )
}
```

#### 4.1.2 告警规则配置

```yaml
# Prometheus 告警规则
groups:
- name: aetherius.rules
  interval: 30s
  rules:
  # 高任务失败率
  - alert: HighTaskFailureRate
    expr: |
      (
        rate(aetherius_tasks_total{status="failed"}[5m]) /
        rate(aetherius_tasks_total[5m])
      ) > 0.1
    for: 2m
    labels:
      severity: warning
      component: orchestrator
    annotations:
      summary: "High task failure rate detected"
      description: "Task failure rate is {{ $value | humanizePercentage }} over the last 5 minutes"

  # AI 服务不可用
  - alert: AIServiceDown
    expr: up{job="aetherius-reasoning"} == 0
    for: 1m
    labels:
      severity: critical
      component: reasoning
    annotations:
      summary: "AI reasoning service is down"
      description: "The AI reasoning service has been down for more than 1 minute"

  # AI 成本超限
  - alert: AIBudgetExceeded
    expr: |
      (
        sum(increase(aetherius_ai_cost_usd_total[1d]))
      ) > 1000
    labels:
      severity: warning
      component: cost
    annotations:
      summary: "Daily AI budget exceeded"
      description: "AI service cost is ${{ $value | humanize }} today, exceeding daily budget"

  # 队列积压
  - alert: QueueBacklog
    expr: sum(aetherius_queue_depth) > 100
    for: 5m
    labels:
      severity: warning
      component: orchestrator
    annotations:
      summary: "Task queue backlog detected"
      description: "Task queue depth is {{ $value }}, indicating processing bottleneck"

  # 高响应延迟
  - alert: HighLatency
    expr: |
      histogram_quantile(0.95,
        rate(aetherius_task_duration_seconds_bucket[5m])
      ) > 600
    for: 5m
    labels:
      severity: warning
      component: orchestrator
    annotations:
      summary: "High task latency detected"
      description: "P95 task duration is {{ $value | humanizeDuration }}, exceeding 10 minutes"

  # 数据库连接池耗尽
  - alert: DatabasePoolExhausted
    expr: |
      (
        sum(go_sql_stats_max_open_connections) - sum(go_sql_stats_open_connections)
      ) < 5
    for: 2m
    labels:
      severity: critical
      component: database
    annotations:
      summary: "Database connection pool near exhaustion"
      description: "Only {{ $value }} database connections available"
```

### 4.2 分布式追踪

#### 4.2.1 OpenTelemetry 集成

```go
// 初始化追踪器
func initTracer(serviceName, endpoint string) (*sdktrace.TracerProvider, error) {
    exporter, err := otlptracegrpc.New(
        context.Background(),
        otlptracegrpc.WithEndpoint(endpoint),
        otlptracegrpc.WithInsecure(),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create exporter: %w", err)
    }

    resource := resource.NewWithAttributes(
        semconv.SchemaURL,
        semconv.ServiceNameKey.String(serviceName),
        semconv.ServiceVersionKey.String(version.Version),
        attribute.String("environment", os.Getenv("ENVIRONMENT")),
    )

    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(resource),
        sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(0.1))),
    )

    otel.SetTracerProvider(tp)
    otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
        propagation.TraceContext{},
        propagation.Baggage{},
    ))

    return tp, nil
}

// 追踪中间件
func TracingMiddleware(tracer trace.Tracer) gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx, span := tracer.Start(c.Request.Context(), c.Request.URL.Path,
            trace.WithAttributes(
                attribute.String("http.method", c.Request.Method),
                attribute.String("http.url", c.Request.URL.String()),
                attribute.String("http.user_agent", c.Request.UserAgent()),
                attribute.String("http.client_ip", c.ClientIP()),
            ),
        )
        defer span.End()

        c.Request = c.Request.WithContext(ctx)
        c.Next()

        span.SetAttributes(
            attribute.Int("http.status_code", c.Writer.Status()),
        )

        if c.Writer.Status() >= 400 {
            span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", c.Writer.Status()))
        }
    }
}
```

#### 4.2.2 关键追踪点

```go
func (s *OrchestratorService) HandleAlert(ctx context.Context, alert Alert) error {
    ctx, span := tracer.Start(ctx, "orchestrator.handle_alert")
    defer span.End()

    span.SetAttributes(
        attribute.String("alert.name", alert.Name),
        attribute.String("alert.severity", alert.Severity),
        attribute.String("cluster.id", alert.ClusterID),
        attribute.String("namespace", alert.Namespace),
    )

    // 知识库查询
    ctx, kbSpan := tracer.Start(ctx, "knowledge_base.query")
    solutions, err := s.kb.QuerySolutions(ctx, alert)
    kbSpan.End()
    if err != nil {
        span.RecordError(err)
        return err
    }

    // AI 推理
    ctx, aiSpan := tracer.Start(ctx, "ai.reasoning")
    plan, err := s.ai.GeneratePlan(ctx, alert, solutions)
    aiSpan.SetAttributes(
        attribute.Int("ai.tokens_used", plan.TokensUsed),
        attribute.String("ai.model", "gpt-4"),
    )
    aiSpan.End()

    // 任务执行
    ctx, execSpan := tracer.Start(ctx, "task.execution")
    result, err := s.executor.Execute(ctx, plan)
    execSpan.SetAttributes(
        attribute.Int("task.steps", len(plan.Steps)),
        attribute.String("task.status", result.Status),
    )
    execSpan.End()

    span.SetAttributes(
        attribute.String("task.id", result.TaskID),
        attribute.String("task.status", result.Status),
    )

    return nil
}
```

### 4.3 日志管理

#### 4.3.1 结构化日志格式

```go
// 初始化日志器
func initLogger(cfg LogConfig) (*zap.Logger, error) {
    encoderConfig := zap.NewProductionEncoderConfig()
    encoderConfig.TimeKey = "timestamp"
    encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

    config := zap.Config{
        Level:            zap.NewAtomicLevelAt(zapcore.InfoLevel),
        Development:      false,
        Encoding:         "json",
        EncoderConfig:    encoderConfig,
        OutputPaths:      []string{"stdout"},
        ErrorOutputPaths: []string{"stderr"},
    }

    logger, err := config.Build(
        zap.AddCaller(),
        zap.AddStacktrace(zapcore.ErrorLevel),
        zap.Fields(
            zap.String("service", "aetherius-orchestrator"),
            zap.String("version", version.Version),
            zap.String("environment", os.Getenv("ENVIRONMENT")),
        ),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to build logger: %w", err)
    }

    return logger, nil
}

// 日志示例
logger.Info("Task execution started",
    zap.String("task_id", task.ID),
    zap.String("alert_id", task.AlertID),
    zap.String("cluster_id", task.ClusterID),
    zap.String("namespace", task.Namespace),
    zap.String("priority", task.Priority.String()),
    zap.Duration("estimated_duration", task.EstimatedDuration),
)
```

## 5. 成本管理

### 5.1 AI 服务成本控制

#### 5.1.1 预算配置

```yaml
# ConfigMap: 成本预算配置
apiVersion: v1
kind: ConfigMap
metadata:
  name: aetherius-cost-budget
  namespace: aetherius
data:
  budget.yaml: |
    budgets:
      daily_limit: 1000.00  # 每日上限 $1000
      monthly_limit: 25000.00  # 每月上限 $25000
      per_task_limit: 0.50  # 单任务上限 $0.50

    thresholds:
      warning: 0.80  # 80% 预警
      critical: 0.95  # 95% 限流

    models:
      gpt-4:
        input_cost_per_1k: 0.03
        output_cost_per_1k: 0.06
        max_tokens: 4000
      gpt-3.5-turbo:
        input_cost_per_1k: 0.0015
        output_cost_per_1k: 0.002
        max_tokens: 4000
```

#### 5.1.2 成本追踪实现

```go
type CostTracker struct {
    db      *sql.DB
    budgets BudgetConfig
    mu      sync.RWMutex
}

type UsageRecord struct {
    Timestamp  time.Time
    TaskID     string
    Model      string
    InputTokens  int
    OutputTokens int
    Cost       float64
}

func (ct *CostTracker) RecordUsage(ctx context.Context, record UsageRecord) error {
    ct.mu.Lock()
    defer ct.mu.Unlock()

    // 检查预算
    dailyCost, err := ct.GetDailyCost(ctx)
    if err != nil {
        return fmt.Errorf("failed to get daily cost: %w", err)
    }

    if dailyCost+record.Cost > ct.budgets.DailyLimit {
        return fmt.Errorf("daily budget exceeded: $%.2f + $%.2f > $%.2f",
            dailyCost, record.Cost, ct.budgets.DailyLimit)
    }

    // 记录使用
    _, err = ct.db.ExecContext(ctx, `
        INSERT INTO ai_usage (
            timestamp, task_id, model, input_tokens, output_tokens, cost
        ) VALUES ($1, $2, $3, $4, $5, $6)
    `, record.Timestamp, record.TaskID, record.Model,
        record.InputTokens, record.OutputTokens, record.Cost)

    if err != nil {
        return fmt.Errorf("failed to record usage: %w", err)
    }

    // 更新指标
    AITokenUsage.WithLabelValues(record.Model, "input").Add(float64(record.InputTokens))
    AITokenUsage.WithLabelValues(record.Model, "output").Add(float64(record.OutputTokens))
    AICost.WithLabelValues(record.Model).Add(record.Cost)

    // 检查预警阈值
    ct.checkThresholds(dailyCost + record.Cost)

    return nil
}

func (ct *CostTracker) checkThresholds(currentCost float64) {
    usage := currentCost / ct.budgets.DailyLimit

    if usage >= ct.budgets.Thresholds.Critical {
        log.Error("AI budget critical threshold exceeded",
            zap.Float64("usage_percent", usage*100),
            zap.Float64("current_cost", currentCost),
            zap.Float64("daily_limit", ct.budgets.DailyLimit))

        // 触发限流
        ct.enableThrottling()

    } else if usage >= ct.budgets.Thresholds.Warning {
        log.Warn("AI budget warning threshold exceeded",
            zap.Float64("usage_percent", usage*100),
            zap.Float64("current_cost", currentCost),
            zap.Float64("daily_limit", ct.budgets.DailyLimit))
    }
}
```

## 6. 灾难恢复

### 6.1 备份策略

#### 6.1.1 自动化备份配置

```yaml
# CronJob: 数据库备份
apiVersion: batch/v1
kind: CronJob
metadata:
  name: aetherius-backup
  namespace: aetherius
spec:
  schedule: "0 2 * * *"  # 每日凌晨 2 点
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 1
  jobTemplate:
    spec:
      template:
        spec:
          restartPolicy: OnFailure
          containers:
          - name: backup
            image: postgres:14
            command:
            - /bin/bash
            - -c
            - |
              # 执行备份
              BACKUP_FILE="aetherius_$(date +%Y%m%d_%H%M%S).sql.gz"

              pg_dump -h $DB_HOST -U $DB_USER $DB_NAME | \
                gzip > /tmp/$BACKUP_FILE

              # 上传到 S3
              aws s3 cp /tmp/$BACKUP_FILE s3://$S3_BUCKET/backups/$BACKUP_FILE

              # 验证备份
              aws s3 ls s3://$S3_BUCKET/backups/$BACKUP_FILE

              # 清理本地文件
              rm /tmp/$BACKUP_FILE

              echo "Backup completed: $BACKUP_FILE"
            env:
            - name: DB_HOST
              value: "postgresql.aetherius.svc.cluster.local"
            - name: DB_USER
              valueFrom:
                secretKeyRef:
                  name: postgresql-secret
                  key: username
            - name: DB_NAME
              value: "aetherius"
            - name: PGPASSWORD
              valueFrom:
                secretKeyRef:
                  name: postgresql-secret
                  key: password
            - name: S3_BUCKET
              value: "aetherius-backups"
            - name: AWS_ACCESS_KEY_ID
              valueFrom:
                secretKeyRef:
                  name: aws-credentials
                  key: access_key_id
            - name: AWS_SECRET_ACCESS_KEY
              valueFrom:
                secretKeyRef:
                  name: aws-credentials
                  key: secret_access_key
```

#### 6.1.2 备份恢复脚本

```bash
#!/bin/bash
# restore-backup.sh

set -e

BACKUP_FILE=$1
DB_HOST=${DB_HOST:-localhost}
DB_USER=${DB_USER:-aetherius}
DB_NAME=${DB_NAME:-aetherius}
S3_BUCKET=${S3_BUCKET:-aetherius-backups}

if [ -z "$BACKUP_FILE" ]; then
    echo "Usage: $0 <backup_file>"
    echo "Available backups:"
    aws s3 ls s3://$S3_BUCKET/backups/ | tail -10
    exit 1
fi

echo "=== Aetherius 数据库恢复 ==="
echo "备份文件: $BACKUP_FILE"
echo "目标数据库: $DB_HOST/$DB_NAME"

read -p "确认恢复? (yes/no): " CONFIRM
if [ "$CONFIRM" != "yes" ]; then
    echo "恢复已取消"
    exit 0
fi

# 下载备份
echo "下载备份文件..."
aws s3 cp s3://$S3_BUCKET/backups/$BACKUP_FILE /tmp/$BACKUP_FILE

# 停止应用服务
echo "停止应用服务..."
kubectl scale deployment --all --replicas=0 -n aetherius

# 删除现有数据库
echo "删除现有数据库..."
PGPASSWORD=$PGPASSWORD psql -h $DB_HOST -U $DB_USER -d postgres \
  -c "DROP DATABASE IF EXISTS $DB_NAME;"
PGPASSWORD=$PGPASSWORD psql -h $DB_HOST -U $DB_USER -d postgres \
  -c "CREATE DATABASE $DB_NAME;"

# 恢复数据
echo "恢复数据..."
gunzip -c /tmp/$BACKUP_FILE | \
  PGPASSWORD=$PGPASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME

# 验证恢复
echo "验证数据..."
RECORD_COUNT=$(PGPASSWORD=$PGPASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME \
  -t -c "SELECT COUNT(*) FROM diagnostic_tasks;")
echo "诊断任务记录数: $RECORD_COUNT"

# 重启应用服务
echo "重启应用服务..."
kubectl scale deployment aetherius-orchestrator --replicas=3 -n aetherius
kubectl scale deployment aetherius-reasoning --replicas=2 -n aetherius
kubectl scale deployment aetherius-execution --replicas=2 -n aetherius

# 清理临时文件
rm /tmp/$BACKUP_FILE

echo "=== 数据库恢复完成 ==="
```

### 6.2 故障转移机制

```go
// 自动故障转移管理器
type FailoverManager struct {
    primary    ServiceEndpoint
    secondary  ServiceEndpoint
    monitor    HealthMonitor
    notifier   AlertNotifier
}

type ServiceEndpoint struct {
    Name     string
    URL      string
    Priority int
    Healthy  bool
}

func (f *FailoverManager) MonitorAndFailover(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            f.checkAndFailover()
        }
    }
}

func (f *FailoverManager) checkAndFailover() {
    if !f.monitor.IsHealthy(f.primary) {
        log.Warn("Primary service unhealthy, initiating failover",
            zap.String("primary", f.primary.Name))

        // 发送告警
        f.notifier.SendAlert(Alert{
            Severity: "critical",
            Title:    "Service Failover Initiated",
            Message:  fmt.Sprintf("Primary service %s is unhealthy", f.primary.Name),
        })

        // 执行故障转移
        if err := f.promoteSecondary(); err != nil {
            log.Error("Failover failed",
                zap.Error(err),
                zap.String("secondary", f.secondary.Name))

            f.notifier.SendAlert(Alert{
                Severity: "critical",
                Title:    "Failover Failed",
                Message:  fmt.Sprintf("Failed to promote secondary: %v", err),
            })
            return
        }

        log.Info("Failover completed successfully",
            zap.String("new_primary", f.secondary.Name))

        f.notifier.SendAlert(Alert{
            Severity: "warning",
            Title:    "Failover Completed",
            Message:  fmt.Sprintf("Service now running on %s", f.secondary.Name),
        })
    }
}
```

## 7. 故障排查指南

### 7.1 常见问题诊断

#### 问题 1: Pod 无法启动

```bash
# 诊断步骤
# 1. 查看 Pod 状态和事件
kubectl describe pod <pod-name> -n aetherius

# 2. 查看容器日志
kubectl logs <pod-name> -n aetherius
kubectl logs <pod-name> -n aetherius --previous  # 查看崩溃前的日志

# 3. 检查资源限制
kubectl top pod <pod-name> -n aetherius
kubectl get pod <pod-name> -n aetherius -o jsonpath='{.spec.containers[*].resources}'

# 4. 检查镜像拉取
kubectl get events -n aetherius --field-selector involvedObject.name=<pod-name>

# 常见原因及解决方案:
# - ImagePullBackOff: 检查镜像名称、私有仓库认证
# - CrashLoopBackOff: 检查应用日志、配置、依赖服务
# - OOMKilled: 增加内存限制或优化应用内存使用
# - Pending: 检查节点资源、亲和性、PVC 绑定
```

#### 问题 2: 数据库连接失败

```bash
# 诊断步骤
# 1. 验证 PostgreSQL 运行状态
kubectl get pods -l app=postgresql -n aetherius
kubectl exec -it postgresql-0 -n aetherius -- pg_isready

# 2. 测试数据库连接
kubectl exec -it postgresql-0 -n aetherius -- \
  psql -U aetherius -d aetherius -c "SELECT 1;"

# 3. 检查网络策略
kubectl get networkpolicies -n aetherius
kubectl describe networkpolicy aetherius-network-policy -n aetherius

# 4. 验证密钥
kubectl get secret postgresql-secret -n aetherius -o json | \
  jq -r '.data.password' | base64 -d

# 5. 查看数据库日志
kubectl logs postgresql-0 -n aetherius

# 常见原因:
# - 密码错误: 验证 Secret 中的密码
# - 网络策略阻止: 检查 NetworkPolicy 规则
# - 连接池耗尽: 检查数据库连接数配置
# - DNS 解析失败: 验证 Service 和 DNS
```

#### 问题 3: AI 服务调用失败

```bash
# 诊断步骤
# 1. 检查 API 密钥
kubectl get secret aetherius-secrets -n aetherius -o json | \
  jq -r '.data."openai-api-key"' | base64 -d

# 2. 测试 API 连通性
kubectl exec -it deployment/aetherius-orchestrator -n aetherius -- \
  curl -s -w "\n%{http_code}" \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  https://api.openai.com/v1/models

# 3. 查看推理服务日志
kubectl logs -l app=aetherius-reasoning -n aetherius --tail=100

# 4. 检查成本预算
kubectl exec -it deployment/aetherius-orchestrator -n aetherius -- \
  curl http://localhost:8080/api/v1/cost/status

# 常见原因:
# - API 密钥失效: 更新 Secret 中的密钥
# - 超出配额: 检查 OpenAI 账户配额
# - 网络限制: 检查出站网络策略
# - 预算超限: 调整成本预算配置
```

### 7.2 性能问题排查

```bash
# 任务处理延迟分析
# 1. 查看任务队列深度
kubectl exec -it deployment/aetherius-orchestrator -n aetherius -- \
  curl http://localhost:9090/metrics | grep aetherius_queue_depth

# 2. 分析任务处理时间
kubectl exec -it deployment/aetherius-orchestrator -n aetherius -- \
  curl http://localhost:9090/metrics | grep aetherius_task_duration_seconds

# 3. 检查 AI 服务延迟
kubectl exec -it deployment/aetherius-reasoning -n aetherius -- \
  curl http://localhost:9090/metrics | grep aetherius_ai_latency

# 4. 数据库慢查询分析
kubectl exec -it postgresql-0 -n aetherius -- \
  psql -U aetherius -d aetherius -c "
    SELECT query, calls, mean_exec_time, max_exec_time
    FROM pg_stat_statements
    ORDER BY mean_exec_time DESC
    LIMIT 10;
  "

# 5. 资源使用情况
kubectl top pods -n aetherius
kubectl top nodes
```

## 8. 运维最佳实践

### 8.1 日常维护清单

```markdown
## 每日检查

- [ ] 检查所有 Pod 运行状态
- [ ] 查看告警和错误日志
- [ ] 验证核心功能可用性
- [ ] 检查资源使用情况
- [ ] 监控 AI 服务成本

## 每周任务

- [ ] 审查安全审计日志
- [ ] 分析性能指标趋势
- [ ] 检查备份完整性
- [ ] 更新知识库内容
- [ ] 清理过期数据

## 每月任务

- [ ] 安全漏洞扫描
- [ ] 容量规划评估
- [ ] 灾难恢复演练
- [ ] 成本优化审查
- [ ] 文档更新

## 每季度任务

- [ ] 系统升级计划
- [ ] 架构评审
- [ ] 合规性审计
- [ ] 用户满意度调查
- [ ] 培训和知识分享
```

### 8.2 安全加固建议

```yaml
# 安全加固检查清单
security_hardening:
  network:
    - 启用网络策略,限制 Pod 间通信
    - 使用 Ingress TLS 加密外部流量
    - 配置 mTLS 保护服务间通信
    - 限制出站网络访问

  authentication:
    - 启用 OIDC 集成进行用户认证
    - 强制使用强密码策略
    - 实施多因子认证 (MFA)
    - 定期轮换 API 密钥

  authorization:
    - 实施最小权限原则
    - 使用 RBAC 细粒度权限控制
    - 审计所有特权操作
    - 定期审查权限分配

  data_protection:
    - 加密静态数据 (数据库、存储)
    - 加密传输数据 (TLS 1.2+)
    - 使用 Vault 管理敏感凭证
    - 实施数据保留策略

  monitoring:
    - 启用审计日志记录
    - 配置实时告警
    - 监控异常行为
    - 定期安全扫描

  compliance:
    - 遵循 GDPR 数据保护要求
    - 满足 SOC2 合规标准
    - 定期合规性评估
    - 维护审计记录
```

## 附录

### A. 运维命令速查表

```bash
# Pod 管理
kubectl get pods -n aetherius
kubectl describe pod <pod-name> -n aetherius
kubectl logs -f <pod-name> -n aetherius
kubectl exec -it <pod-name> -n aetherius -- /bin/sh

# 服务管理
kubectl get svc -n aetherius
kubectl port-forward svc/aetherius-orchestrator 8080:80 -n aetherius

# 配置管理
kubectl get configmap aetherius-config -n aetherius -o yaml
kubectl edit configmap aetherius-config -n aetherius

# 密钥管理
kubectl get secrets -n aetherius
kubectl create secret generic <name> --from-literal=key=value -n aetherius

# 资源监控
kubectl top pods -n aetherius
kubectl top nodes

# 扩缩容
kubectl scale deployment aetherius-orchestrator --replicas=5 -n aetherius

# 滚动更新
kubectl set image deployment/aetherius-orchestrator \
  orchestrator=aetherius/orchestrator:v1.7 -n aetherius
kubectl rollout status deployment/aetherius-orchestrator -n aetherius
kubectl rollout undo deployment/aetherius-orchestrator -n aetherius
```

### B. 相关文档

- [架构设计文档](./02_architecture.md) - 系统架构详细设计
- [数据模型文档](./03_data_models.md) - 核心数据模型定义
- [部署配置文档](./04_deployment.md) - 部署配置指南
- [需求规格说明](../REQUIREMENTS.md) - 完整需求索引