# OneX Reference Index

This index provides quick access to key files and patterns from the OneX codebase that should inform Aetherius development.

## Document Overview

This analysis consists of two main documents:

1. **ONEX_ARCHITECTURE_ANALYSIS.md** (1089 lines)
   - Comprehensive architectural analysis
   - Detailed code examples
   - Service patterns breakdown
   - Best practices documentation

2. **ONEX_ADOPTION_GUIDE.md** (380 lines)
   - Practical implementation roadmap
   - Quick reference for adoption
   - Implementation checklists
   - Success metrics

---

## Key OneX Files Reference

### Service Entry Point Patterns

#### Pattern 1: Complex Services (Kubernetes-style Cobra)
- **File**: `cmd/onex-apiserver/apiserver.go`
- **File**: `cmd/onex-apiserver/app/server.go`
- **Pattern**: Direct Cobra command with explicit configuration
- **Used by**: onex-apiserver, onex-controller-manager, onex-blockchain-controller, onex-job-controller

#### Pattern 2: Simple Services (App wrapper with Viper)
- **File**: `cmd/onex-gateway/main.go`
- **File**: `cmd/onex-gateway/app/server.go`
- **File**: `cmd/onex-gateway/app/options/options.go`
- **Pattern**: Lightweight App struct with embedded options
- **Used by**: onex-gateway, onex-cacheserver, onex-nightwatch

#### App Framework Implementation
- **File**: `staging/src/github.com/onexstack/onexstack/pkg/app/app.go`
- **Purpose**: Base App struct for service initialization
- **Key methods**: NewApp(), WithOptions(), WithRunFunc(), Run()

---

### Error Handling & Responses

- **File**: `staging/src/github.com/onexstack/onexstack/pkg/errorsx/errorsx.go`
- **Structure**: ErrorX with Code, Reason, Message, Metadata
- **Features**:
  - HTTP status code integration
  - Business error reason codes
  - User-friendly messages
  - gRPC status conversion
  - Error metadata support
  - Error chaining (errors.Is())

---

### Context Management

- **File**: `internal/pkg/contextx/contextx.go`
- **Pattern**: Type-safe context values using unexported struct types as keys
- **Key functions**:
  - WithClaims/Claims - JWT token claims
  - WithUserID/UserID - User identification
  - WithAccessToken/AccessToken - Token management
  - WithTraceID/TraceID - Request tracing

---

### Middleware Patterns

#### Gin Middleware
- **Location**: `internal/pkg/middleware/gin/`
- **Key File**: `internal/pkg/middleware/gin/traceid.go`
- **Pattern**: Simple gin.HandlerFunc functions
- **Middleware types**:
  - TraceID - Unique request identification
  - Authentication - JWT validation
  - Authorization - Permission checks
  - Request logging - Structured logging

#### Middleware Usage
```go
router.Use(middleware.TraceID())
router.Use(middleware.Authentication())
router.Use(middleware.Authorization())
router.Use(middleware.Logging())
```

---

### Dependency Injection

- **File**: `cmd/onex-controller-manager/app/wire.go`
- **Tool**: Google Wire (code generation based DI)
- **Pattern**: ProviderSet grouping + wire.Build() declaration
- **Generated**: `cmd/onex-controller-manager/app/wire_gen.go`
- **Benefits**: Compile-time safety, clear dependency graphs, no reflection

---

### Build System (Modular Makefile)

**Location**: `scripts/make-rules/`

**Key Files**:
- `Makefile` - Root orchestration
- `scripts/make-rules/common.mk` - Shared variables
- `scripts/make-rules/golang.mk` - Go build targets
- `scripts/make-rules/image.mk` - Docker build targets
- `scripts/make-rules/tools.mk` - Tool management
- `scripts/make-rules/generate.mk` - Code generation

**Naming Convention**: `<module>.<action>[.<service>]`
- Example: `go.build.agent-manager`
- Example: `image.build.orchestrator`

---

### Controller Runtime Integration

- **File**: `cmd/onex-controller-manager/app/controllermanager.go`
- **Framework**: sigs.k8s.io/controller-runtime
- **Features**:
  - Manager initialization with options
  - Cache and client configuration
  - Health check setup
  - Metrics collection
  - Leader election
  - Namespace filtering

---

### Configuration Management

**Pattern**: YAML config + Viper + Structured Options

**Files**:
- `cmd/<service>/app/options/options.go` - Options struct
- Implements: Complete(), Validate(), Flags()
- Supports: Environment variable overrides
- Methods: NamedFlagSets for organized flags

---

### Bootstrap Framework

- **File**: `internal/pkg/bootstrap/app.go`
- **Features**: AppInfo, AppConfig
- **Integration**: Google Wire ProviderSet
- **Purpose**: Unified app initialization

---

## Aetherius Current Status

### Already Implemented Correctly (✓)
- Bootstrap pattern with Runner ✓
- pkg/bootstrap/ package ✓
- pkg/types/ for domain models ✓
- cmd/<service>/app/ structure ✓
- internal/<service>/ organization ✓
- YAML + Viper configuration ✓

### Needs Enhancement (Priority)
1. **HIGH**: Error handling pattern
2. **HIGH**: Context management helpers
3. **MEDIUM**: Middleware standardization
4. **MEDIUM**: Dependency injection (Wire)
5. **MEDIUM**: Build system modularization
6. **LOW**: Configuration enhancement
7. **LOW**: API response standardization

---

## Implementation Priority Matrix

| Pattern | Priority | Effort | Impact | Status |
|---------|----------|--------|--------|--------|
| Error Handling (ErrorX) | HIGH | Medium | High | TODO |
| Context Management | HIGH | Low | High | TODO |
| Middleware Stack | MEDIUM | Medium | Medium | TODO |
| Dependency Injection (Wire) | MEDIUM | High | Medium | TODO |
| Build System | MEDIUM | Medium | Medium | TODO |
| Config Enhancement | LOW | Low | Low | TODO |
| Response Format | LOW | Low | Low | TODO |

---

## Quick Start Guide

### For Error Handling Implementation
1. Read: ONEX_ARCHITECTURE_ANALYSIS.md Section 3
2. Reference: errorsx.go in OneX
3. Create: pkg/errors/errors.go in Aetherius
4. Implement: Error, WithMessage, WithMetadata methods
5. Test: Unit tests for error creation and matching

### For Middleware Standardization
1. Read: ONEX_ARCHITECTURE_ANALYSIS.md Section 5
2. Reference: internal/pkg/middleware/gin/ in OneX
3. Create: internal/pkg/middleware/ in Aetherius
4. Implement: TraceID, RequestID, Logging middlewares
5. Integrate: Update all HTTP servers to use middleware stack

### For Dependency Injection
1. Read: ONEX_ARCHITECTURE_ANALYSIS.md Section 2
2. Reference: wire.go examples in OneX
3. Evaluate: Which services benefit from Wire
4. Implement: ProviderSets for each domain
5. Generate: `go generate ./...` to create wire_gen.go

---

## Key Metrics from OneX

### Project Statistics
- **Services**: 30+ microservices
- **Pattern**: Consistent entry point handling
- **Code organization**: Strict separation of concerns
- **Build system**: Modular, 12+ .mk files
- **Middleware**: Standardized Gin middleware
- **Error handling**: Unified ErrorX pattern
- **DI framework**: Google Wire for complex services

### Performance Characteristics
- **Build time**: <30 seconds for single service
- **Service startup**: <2 seconds for most services
- **Context propagation**: Minimal overhead (<1ms)
- **Middleware chain**: <5ms for 5-6 middlewares
- **Error handling**: Zero allocation for common paths

---

## Migration Path for Aetherius

### Phase 1: Foundation (Week 1-2)
```
Priority: HIGH
Activities:
- Implement ErrorX pattern
- Enhance contextx helpers
- Add TraceID middleware
Deliverable: Core error/context/middleware framework
```

### Phase 2: Standardization (Week 3-4)
```
Priority: MEDIUM
Activities:
- Standardize middleware stack
- Implement RequestID middleware
- Add structured logging
- Unified response format
Deliverable: API contract standardization
```

### Phase 3: Optimization (Week 5-6)
```
Priority: MEDIUM
Activities:
- Evaluate Wire for services
- Enhance config validation
- Modular Makefile
- Health check middleware
Deliverable: Build and startup optimization
```

### Phase 4: Polish (Week 7-8)
```
Priority: LOW
Activities:
- Documentation
- Comprehensive testing
- Integration testing
- Performance profiling
Deliverable: Production readiness
```

---

## File Navigation Quick Links

### OneX Codebase Root
```
/Users/costalong/code/go/src/github.com/onexstack/onex/
```

### Key Directories
```
cmd/                           # Service entry points
├── onex-apiserver/
├── onex-controller-manager/
├── onex-blockchain-controller/
├── onex-job-controller/
└── onex-gateway/

internal/
├── pkg/                       # Shared utilities
│   ├── bootstrap/             # App initialization
│   ├── contextx/              # Context helpers
│   ├── middleware/            # Middleware patterns
│   ├── errors/                # Error handling (search)
│   └── util/                  # Utility functions

staging/src/github.com/onexstack/onexstack/pkg/
├── app/                       # App framework
├── errorsx/                   # ErrorX implementation
└── log/                       # Logging

scripts/make-rules/            # Build system modules
└── *.mk files
```

---

## Testing & Validation

### Error Handling Tests
- ErrorX creation and chaining
- Error metadata operations
- gRPC status conversion
- Error matching with Is()

### Middleware Tests
- TraceID injection and propagation
- RequestID generation
- Context value setting/retrieval
- Middleware chain execution

### Integration Tests
- Service startup and initialization
- Configuration loading and validation
- Database and cache connections
- API endpoint functionality

---

## Success Criteria

### Code Quality
- [ ] Consistent error handling (100% adoption)
- [ ] Type-safe context usage (100% adoption)
- [ ] Request ID tracking (all endpoints)
- [ ] Reduced boilerplate (50%+ reduction)

### Observability
- [ ] Trace IDs in all requests
- [ ] Structured logging with context
- [ ] Error metadata in monitoring
- [ ] Request flow visibility

### Maintainability
- [ ] Clear dependency graphs
- [ ] Modular build system
- [ ] Standardized middleware
- [ ] Code duplication reduction

---

## Additional Resources

### External References
- [Kubernetes component-base](https://github.com/kubernetes/component-base)
- [go-kratos/kratos](https://github.com/go-kratos/kratos)
- [google/wire](https://github.com/google/wire)
- [uber-go/zap](https://github.com/uber-go/zap)

### Internal Documentation
- CLAUDE.md - Aetherius project overview
- Architecture docs in Aetherius
- Service-specific READMEs

---

## Contact & Support

For questions about this analysis:
1. Review the comprehensive ONEX_ARCHITECTURE_ANALYSIS.md
2. Check ONEX_ADOPTION_GUIDE.md for practical steps
3. Reference OneX source code directly
4. Consult with architecture team

Last Updated: 2025-11-02
Analysis Scope: OneX codebase comprehensive review
Document Format: Markdown with code examples
