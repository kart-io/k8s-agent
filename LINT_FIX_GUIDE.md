# Golangci-lint Fix Guide - Quick Reference

This guide provides copy-paste solutions for common lint issues in the k8s-agent codebase.

## Table of Contents
1. [errcheck - Unchecked Errors](#1-errcheck---unchecked-errors)
2. [gosec - Security Issues](#2-gosec---security-issues)
3. [staticcheck - Deprecated APIs](#3-staticcheck---deprecated-apis)
4. [noctx - Missing Context](#4-noctx---missing-context)
5. [goconst - Extract Constants](#5-goconst---extract-constants)
6. [revive - Missing Comments](#6-revive---missing-comments)

---

## 1. errcheck - Unchecked Errors

### Problem: Defer statements ignoring errors
**Locations**: 46 issues across codebase

#### Pattern 1: defer Close() in functions
```go
// ❌ Before
func process() error {
    file, err := os.Open("config.yaml")
    if err != nil {
        return err
    }
    defer file.Close()  // ❌ errcheck: unchecked error
    // ...
}

// ✅ After - Option 1: Explicit ignore
func process() error {
    file, err := os.Open("config.yaml")
    if err != nil {
        return err
    }
    defer func() { _ = file.Close() }()  // ✅ Explicit ignore
    // ...
}

// ✅ After - Option 2: With nolint comment
func process() error {
    file, err := os.Open("config.yaml")
    if err != nil {
        return err
    }
    defer file.Close() //nolint:errcheck // cleanup in defer
    // ...
}
```

#### Pattern 2: Logger Flush in shutdown
```go
// ❌ Before
func Shutdown() {
    logger.Flush()  // ❌ errcheck
}

// ✅ After
func Shutdown() {
    _ = logger.Flush()  // Best effort flush during shutdown
}
```

#### Pattern 3: SetTrustedProxies
```go
// ❌ Before
router.SetTrustedProxies(nil)  // ❌ errcheck

// ✅ After
_ = router.SetTrustedProxies(nil)  // Error unlikely with nil argument
```

---

## 2. gosec - Security Issues

### G112: Missing ReadHeaderTimeout (CRITICAL)
**Impact**: Vulnerable to Slowloris attacks
**Locations**: 5 HTTP servers

```go
// ❌ Before
server := &http.Server{
    Addr:    ":8080",
    Handler: router,
}

// ✅ After
server := &http.Server{
    Addr:              ":8080",
    Handler:           router,
    ReadHeaderTimeout: 10 * time.Second,  // Prevent Slowloris attacks
    ReadTimeout:       30 * time.Second,  // Optional: full read timeout
    WriteTimeout:      30 * time.Second,  // Optional: write timeout
}
```

**Files to fix**:
- `examples/http-gateway/main.go:191`
- `internal/auth/initializers/server.go:191`
- `internal/monitor/api/server.go:112`

### G204: Command Injection (FALSE POSITIVE)
**Note**: Already validated in code, just add comment

```go
// ❌ Linter warning
execCmd = exec.CommandContext(ctx, cmd.Tool, cmd.Args...)  // G204

// ✅ After
// Tool and args are validated against whitelist in validateCommand()
execCmd = exec.CommandContext(ctx, cmd.Tool, cmd.Args...) //nolint:gosec // G204 - validated
```

### G601: Memory Aliasing in Loop
```go
// ❌ Before
for _, item := range items {
    results = append(results, &item)  // ❌ G601: all pointers point to same address
}

// ✅ After - Option 1: Copy to local var
for _, item := range items {
    item := item  // Create new variable
    results = append(results, &item)
}

// ✅ After - Option 2: Use index
for i := range items {
    results = append(results, &items[i])
}
```

---

## 3. staticcheck - Deprecated APIs

### SA1019: grpc.Dial is deprecated
```go
// ❌ Before
conn, err := grpc.Dial(addr, opts...)  // SA1019: deprecated

// ✅ After
conn, err := grpc.NewClient(addr, opts...)
```

**Location**: `examples/grpc-client/main.go:28`

### SA1019: cache.NewInformer is deprecated
```go
// ❌ Before
_, controller := cache.NewInformer(...)  // SA1019: deprecated

// ✅ After
_, controller := cache.NewInformerWithOptions(
    lw,
    &corev1.Event{},
    time.Minute,
    cache.ResourceEventHandlerFuncs{...},
    cache.InformerOptions{},
)
```

**Location**: `internal/collect-agent/agent/event_watcher.go:64`

### ST1005: Error Messages Should Not Be Capitalized
```go
// ❌ Before
return fmt.Errorf("Gemini API key is required")  // ST1005

// ✅ After
return fmt.Errorf("gemini API key is required")
```

**Locations**:
- `internal/reasoning/llm/gemini.go:22`
- `internal/reasoning/llm/gemini.go:146`
- `internal/reasoning/llm/kimi.go:23`

### S1009: Redundant Nil Check Before len()
```go
// ❌ Before
if embedding == nil || len(embedding) == 0 {  // S1009: redundant nil check
    return nil
}

// ✅ After
if len(embedding) == 0 {  // len() of nil slice is 0
    return nil
}
```

---

## 4. noctx - Missing Context

### net.Listen → net.ListenConfig.Listen
```go
// ❌ Before
listener, err := net.Listen("tcp", addr)  // noctx

// ✅ After
lc := net.ListenConfig{}
listener, err := lc.Listen(ctx, "tcp", addr)
```

### http.NewRequest → http.NewRequestWithContext
```go
// ❌ Before
req, err := http.NewRequest("GET", url, nil)  // noctx

// ✅ After
req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
```

### db.Exec → db.ExecContext
```go
// ❌ Before
_, err := db.Exec(query, args...)  // noctx

// ✅ After
_, err := db.ExecContext(ctx, query, args...)
```

### db.Ping → db.PingContext
```go
// ❌ Before
if err := db.Ping(); err != nil {  // noctx

// ✅ After
if err := db.PingContext(ctx); err != nil {
```

---

## 5. goconst - Extract Constants

### Common Repeated Strings
**14 issues** - strings repeated 3+ times should be constants

```go
// ❌ Before
if status == "failed" { ... }
result.Status = "failed"
log.Error("Operation failed")

// ✅ After
const (
    StatusFailed  = "failed"
    StatusSuccess = "success"
    StatusPending = "pending"
)

if status == StatusFailed { ... }
result.Status = StatusFailed
```

### Recommended Constants File
Create `internal/<service>/constants/status.go`:

```go
package constants

const (
    // Status values
    StatusSuccess = "success"
    StatusFailed  = "failed"
    StatusPending = "pending"
    StatusUnknown = "unknown"

    // Common values
    EventTypeManual = "manual"
    StateHealthy    = "healthy"
    StateReady      = "Ready"

    // Email constants
    PlatformEmail        = "email"
    MessageEmailFailed   = "Email delivery failed"

    // Genesis hash for audit chain
    HashGenesis = "genesis"
)
```

**Strings to extract**:
- "genesis" (4 occurrences)
- "email" (3 occurrences)
- "failed" (5 occurrences)
- "success" (6 occurrences)
- "manual" (3 occurrences)
- "healthy" (3 occurrences)
- "Ready" (3 occurrences)
- "unknown" (3 occurrences)

---

## 6. revive - Missing Comments

### Package Comments
```go
// ❌ Before
package crypto

// ✅ After
// Package crypto provides password hashing and verification utilities.
package crypto
```

### Exported Function Comments
```go
// ❌ Before
func NewK8sAPIHandler(...) *K8sAPIHandler {

// ✅ After
// NewK8sAPIHandler creates a new K8s API handler with all service dependencies.
func NewK8sAPIHandler(...) *K8sAPIHandler {
```

### Exported Type Comments
```go
// ❌ Before
type MySQLStorage struct {

// ✅ After
// MySQLStorage implements storage operations using MySQL database.
type MySQLStorage struct {
```

### Exported Method Comments
```go
// ❌ Before
func (s *MySQLStorage) Close() error {

// ✅ After
// Close closes the MySQL database connection.
func (s *MySQLStorage) Close() error {
```

---

## Batch Fix Scripts

### Fix all defer Close() issues
```bash
# Create a script to add _ = to all defer Close() calls
find . -name "*.go" -type f ! -path "./vendor/*" -exec sed -i 's/defer \(.*\.Close()\)$/defer func() { _ = \1 }()/g' {} \;
```

### Fix all defer Flush() issues
```bash
find . -name "*.go" -type f ! -path "./vendor/*" -exec sed -i 's/^\t\(.*\.Flush()\)$/\t_ = \1/g' {} \;
```

---

## Verification Commands

After applying fixes, verify with:

```bash
# Build to ensure no compilation errors
make go.build

# Run tests
make test

# Re-run linter
make go.lint 2>&1 | tee /tmp/lint-results-after.txt

# Count remaining issues
make go.lint 2>&1 | grep "issues:" | tail -1
```

---

## Priority Order for Fixing

1. **🔴 Critical Security** (G112 - HTTP timeouts)
2. **🟠 Logic Errors** (nilerr, nilnil) - ✅ DONE
3. **🟡 Error Handling** (errcheck)
4. **🟢 Deprecated APIs** (staticcheck SA1019)
5. **🔵 Context Support** (noctx)
6. **⚪ Constants** (goconst)
7. **⚫ Documentation** (revive)

---

## Common Mistakes to Avoid

❌ **Don't blindly ignore all errors**
```go
// Bad: Hides real issues
defer func() { _ = db.Close() }()  // What if Close() fails?
```

✅ **Do consider context**
```go
// Good: Log critical errors, ignore cleanup errors
if err := db.Close(); err != nil {
    logger.Warnw("Failed to close database", "error", err)
}

// Acceptable: Cleanup in defer where error is not actionable
defer func() { _ = logger.Flush() }()  // Best effort during shutdown
```

---

**Last Updated**: 2025-11-04
**Based On**: golangci-lint analysis of k8s-agent codebase
