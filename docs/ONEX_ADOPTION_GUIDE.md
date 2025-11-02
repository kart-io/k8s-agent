# OneX Architecture Adoption Guide for Aetherius k8s-agent

This guide provides practical recommendations for adopting proven patterns from the OneX codebase into the Aetherius k8s-agent project.

## Quick Reference: Which Patterns to Adopt

### 1. Error Handling (HIGH PRIORITY)

**Current State**: Needs enhancement  
**Recommendation**: Adopt ErrorX pattern from OneX  
**Effort**: Medium  
**Impact**: Standardized error responses, better debugging

**Action Items**:
1. Create `pkg/errors/` package with ErrorX implementation
2. Define error code constants (similar to OneX)
3. Update all error returns to use ErrorX
4. Support metadata and request ID tracking

**Example Implementation**:
```go
// pkg/errors/errors.go
type Error struct {
    Code     int               `json:"code"`
    Reason   string            `json:"reason"`
    Message  string            `json:"message"`
    Metadata map[string]string `json:"metadata,omitempty"`
}

var (
    ErrInternal     = New(500, "InternalError", "internal server error")
    ErrNotFound     = New(404, "NotFound", "resource not found")
    ErrUnauthorized = New(401, "Unauthorized", "unauthorized")
)

// In handlers
if agent == nil {
    return ErrNotFound.
        WithMessage("agent %s not found", agentID).
        WithRequestID(requestID)
}
```

---

### 2. Context Management (HIGH PRIORITY)

**Current State**: Basic context usage  
**Recommendation**: Implement type-safe context helpers  
**Effort**: Low  
**Impact**: Type-safe context operations, fewer runtime errors

**Action Items**:
1. Enhance `pkg/contextx/` with helper functions
2. Define context key types (unexported structs)
3. Implement type-safe getters/setters
4. Document context value propagation

**Example Implementation**:
```go
// pkg/contextx/contextx.go
type (
    agentIDKey    struct{}
    clusterIDKey  struct{}
    requestIDKey  struct{}
    traceIDKey    struct{}
)

func WithAgentID(ctx context.Context, agentID string) context.Context {
    return context.WithValue(ctx, agentIDKey{}, agentID)
}

func AgentID(ctx context.Context) string {
    agentID, _ := ctx.Value(agentIDKey{}).(string)
    return agentID
}

// Usage in middleware
router.Use(func(c *gin.Context) {
    ctx := contextx.WithTraceID(c.Request.Context(), traceID)
    ctx = contextx.WithAgentID(ctx, agentID)
    c.Request = c.Request.WithContext(ctx)
    c.Next()
})
```

---

### 3. Middleware Standardization (MEDIUM PRIORITY)

**Current State**: Ad-hoc middleware implementations  
**Recommendation**: Use Gin middleware pattern from OneX  
**Effort**: Medium  
**Impact**: Consistent request processing, better logging

**Key Middleware to Implement**:
1. TraceID - Inject trace IDs in all requests
2. RequestID - Generate/propagate request IDs
3. Logging - Structured request logging with context
4. Authentication - JWT token validation
5. ErrorHandling - Unified error response format

**Example: TraceID Middleware**:
```go
// internal/pkg/middleware/traceid.go
func TraceID() gin.HandlerFunc {
    return func(c *gin.Context) {
        traceID := c.Request.Header.Get("X-Trace-ID")
        if traceID == "" {
            traceID = uuid.New().String()
        }
        
        c.Writer.Header().Set("X-Trace-ID", traceID)
        ctx := contextx.WithTraceID(c.Request.Context(), traceID)
        c.Request = c.Request.WithContext(ctx)
        
        c.Next()
    }
}

// In server setup
router.Use(middleware.TraceID())
router.Use(middleware.RequestID())
router.Use(middleware.Logging())
```

---

### 4. Dependency Injection (MEDIUM PRIORITY)

**Current State**: Manual initialization  
**Recommendation**: Evaluate Google Wire for complex services  
**Effort**: High (upfront), but pays off with scale  
**Impact**: Clear dependency graphs, testability

**When to Use Wire**:
- agent-manager: YES (complex dependencies)
- orchestrator: YES (multiple initializers)
- auth: YES (multiple external connections)
- cluster: MAYBE (smaller scope)
- reasoning: MAYBE (complex AI integration)

**Implementation Pattern**:
```go
// cmd/agent-manager/app/wire.go
//go:build wireinject
// +build wireinject

func initializeApp(cfg *config.Config) (*app.App, error) {
    wire.Build(
        db.ProviderSet,        // MySQL client
        cache.ProviderSet,     // Redis client
        nats.ProviderSet,      // NATS messaging
        initializers.ProviderSet,
        app.ProviderSet,
    )
    return nil, nil
}
```

---

### 5. Configuration Management Enhancement (LOW PRIORITY)

**Current State**: Already using YAML + Viper ✓  
**Recommendation**: Enhance with structured options  
**Effort**: Low  
**Impact**: Better config validation, type safety

**Implementation**:
```go
// cmd/<service>/app/options/options.go
type ServerOptions struct {
    Server *ServerConfig `mapstructure:"server"`
    DB     *DBConfig     `mapstructure:"database"`
    NATS   *NATSConfig   `mapstructure:"nats"`
}

func (o *ServerOptions) Complete() error {
    if o.Server.Port == 0 {
        o.Server.Port = 8080
    }
    return nil
}

func (o *ServerOptions) Validate() error {
    if o.Server.Port < 1 || o.Server.Port > 65535 {
        return fmt.Errorf("invalid port: %d", o.Server.Port)
    }
    return nil
}
```

---

### 6. API Response Standardization (LOW PRIORITY)

**Current State**: Varies by service  
**Recommendation**: Unified response format  
**Effort**: Low  
**Impact**: Consistent client expectations

**Implementation**:
```go
// pkg/response/response.go
type Response struct {
    Code      int         `json:"code"`
    Message   string      `json:"message"`
    Data      interface{} `json:"data,omitempty"`
    TraceID   string      `json:"traceId,omitempty"`
    Timestamp int64       `json:"timestamp"`
}

// In handlers
func (h *Handler) GetAgent(c *gin.Context, id string) {
    agent, err := h.service.GetAgent(ctx, id)
    if err != nil {
        response.Error(c, err)
        return
    }
    response.Success(c, agent)
}
```

---

### 7. Build System Enhancement (MEDIUM PRIORITY)

**Current State**: Functional Makefile  
**Recommendation**: Adopt modular make-rules from OneX  
**Effort**: Medium  
**Impact**: Cleaner build logic, easier maintenance

**Key Improvements**:
1. Split Makefile into modular `.mk` files
2. Use consistent naming: `go.build.agent-manager`
3. Add service-specific build targets
4. Improve Docker build organization

**Structure**:
```
scripts/make-rules/
├── common.mk          # Shared variables
├── golang.mk          # Go build targets
├── docker.mk          # Docker targets
├── tools.mk           # Tool management
└── lint.mk            # Linting config
```

---

## Implementation Roadmap

### Phase 1: Foundation (Week 1-2)
**Priority**: HIGH
- [ ] Adopt ErrorX pattern (errors package)
- [ ] Enhance contextx with typed helpers
- [ ] Implement TraceID middleware

**Effort**: 2-3 days  
**Review**: Code review + testing

### Phase 2: Standardization (Week 3-4)
**Priority**: MEDIUM
- [ ] Standardize middleware stack
- [ ] Implement RequestID middleware
- [ ] Add structured logging middleware
- [ ] Implement unified response format

**Effort**: 3-5 days  
**Review**: API contract verification

### Phase 3: Optimization (Week 5-6)
**Priority**: MEDIUM
- [ ] Evaluate Wire for agent-manager
- [ ] Enhance config validation
- [ ] Refactor Makefile to modular system
- [ ] Add health check middleware

**Effort**: 5-7 days  
**Review**: Performance testing

### Phase 4: Polish (Week 7-8)
**Priority**: LOW
- [ ] Documentation updates
- [ ] Error handling comprehensive review
- [ ] Integration testing
- [ ] Performance profiling

**Effort**: 3-5 days  
**Review**: Final QA

---

## Code Organization Comparison

### OneX Pattern
```
cmd/onex-gateway/
├── main.go
└── app/
    ├── server.go              (NewApp, run function)
    ├── options/
    │   └── options.go        (ServerOptions, Flags, Complete, Validate)
    └── config.go             (Config struct, NewServer)

internal/pkg/
├── contextx/               (Type-safe context values)
├── middleware/
│   └── gin/                (Gin middleware functions)
├── errors/                 (ErrorX implementation)
└── bootstrap/              (Wire setup)
```

### Aetherius Current (Correct ✓)
```
cmd/<service>/app/
├── app.go or main.go
├── app/
│   ├── options/            (Already has this ✓)
│   └── initializers/       (Bootstrap pattern ✓)

pkg/
├── bootstrap/              (Already has this ✓)
├── contextx/               (Needs enhancement)
└── types/

internal/pkg/
└── middleware/             (Needs standardization)
```

**Recommendation**: Keep current structure, enhance existing packages rather than restructure.

---

## Testing Strategy

### Unit Tests
```go
// Test error creation and metadata
func TestErrorWithMetadata(t *testing.T) {
    err := ErrNotFound.
        WithMessage("agent %s not found", "agent-1").
        WithRequestID("req-123")
    
    assert.Equal(t, 404, err.Code)
    assert.Equal(t, "NotFound", err.Reason)
    assert.Equal(t, "req-123", err.Metadata["X-Request-ID"])
}

// Test context helpers
func TestContextAgentID(t *testing.T) {
    ctx := context.Background()
    ctx = contextx.WithAgentID(ctx, "agent-123")
    
    assert.Equal(t, "agent-123", contextx.AgentID(ctx))
}
```

### Integration Tests
```go
// Test middleware chain
func TestTraceIDMiddleware(t *testing.T) {
    router := gin.New()
    router.Use(middleware.TraceID())
    router.GET("/test", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "traceID": contextx.TraceID(c.Request.Context()),
        })
    })
    
    req := httptest.NewRequest("GET", "/test", nil)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    assert.NotEmpty(t, w.Header().Get("X-Trace-ID"))
}
```

---

## Backward Compatibility

### Migration Strategy
1. **Keep existing code working** during transition
2. **Add new patterns alongside** current implementations
3. **Gradually migrate** one module at a time
4. **Document breaking changes** clearly

### Example: Error Handling
```go
// Old way (keep working)
return fmt.Errorf("failed to create agent: %w", err)

// New way (gradually adopt)
return NewError(500, "CreateAgentFailed", 
    "failed to create agent").
    WithMetadata(map[string]string{
        "agentID": agentID,
        "error": err.Error(),
    })

// Adapter for compatibility
func legacyErrorToNew(err error) *Error {
    return ErrInternal.WithMessage(err.Error())
}
```

---

## Success Metrics

### Code Quality
- [ ] Consistent error handling across all services
- [ ] Type-safe context usage throughout codebase
- [ ] All errors include request ID tracking
- [ ] Reduced error handling boilerplate (>50%)

### Observability
- [ ] All requests have trace IDs
- [ ] Structured logging includes context
- [ ] Error metadata tracked in monitoring
- [ ] Request flow visible in logs

### Maintainability
- [ ] Clear dependency graphs (with Wire)
- [ ] Modular Makefile (easier to extend)
- [ ] Consistent middleware stack
- [ ] Reduced code duplication

### Testing
- [ ] Error handling test coverage >90%
- [ ] Middleware test coverage >85%
- [ ] Integration tests for critical paths
- [ ] Load testing with tracing enabled

---

## Resources

### OneX Reference Files
- Service Entry: `cmd/onex-apiserver/app/server.go`
- Error Handling: `staging/src/github.com/onexstack/onexstack/pkg/errorsx/`
- Context: `internal/pkg/contextx/contextx.go`
- Middleware: `internal/pkg/middleware/gin/`
- Dependency Injection: `cmd/onex-controller-manager/app/wire.go`

### Kubernetes & Go Best Practices
- [k8s.io/component-base](https://github.com/kubernetes/component-base) - Logging, flags, metrics
- [go-kratos/kratos](https://github.com/go-kratos/kratos) - App framework patterns
- [google/wire](https://github.com/google/wire) - Dependency injection
- [uber-go/zap](https://github.com/uber-go/zap) - Structured logging

---

## Questions & Support

For detailed implementation questions, refer to:
1. `docs/ONEX_ARCHITECTURE_ANALYSIS.md` - Comprehensive architectural analysis
2. OneX source code - Reference implementations
3. Project Makefile - Build system examples
4. Existing service code - Current patterns

Contact the architecture team for:
- Pattern selection guidance
- Implementation approach review
- Performance impact assessment
- Testing strategy refinement

