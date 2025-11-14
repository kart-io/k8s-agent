# Import Layering Verification Guide

This document provides tools and procedures to verify and enforce import layering compliance in `pkg/agent`.

## Quick Verification Commands

### Check for circular dependencies
```bash
go mod graph | awk '/pkg.agent/{print}' | sort -u
```

### List all imports for a specific package
```bash
grep -r "^import" /Users/costalong/code/go/src/github.com/kart/k8s-agent/pkg/agent/tools/ | grep pkg/agent
```

### Find violations of "tools should not import agents"
```bash
find /Users/costalong/code/go/src/github.com/kart/k8s-agent/pkg/agent/tools -name "*.go" -exec grep -l "pkg/agent/agents" {} \;
```

### Find violations of "production code importing examples"
```bash
find /Users/costalong/code/go/src/github.com/kart/k8s-agent/pkg/agent -path "*/examples" -prune -o -name "*.go" -exec grep -l "pkg/agent/examples" {} \;
```

## Comprehensive Dependency Map

### Layer 1: Foundation (No dependencies on pkg/agent)
```
interfaces/
├── imports: stdlib + external only
├── exports: Agent, Tool, Memory, Store, LLM, Middleware interfaces
└── no circular dependencies

errors/
├── imports: stdlib + external only
├── exports: Error types, helpers
└── independent

cache/
├── imports: stdlib + external only
├── exports: Cache implementations
└── independent

utils/
├── imports: stdlib + external only
├── exports: Utility functions
└── independent
```

### Layer 2: Business Logic (Depends on Layer 1, provides to Layer 3)

**core/**
```
core/
├── agent.go
│   └── imports: interfaces/, errors/, cache/
├── execution/
│   └── imports: interfaces/, errors/
├── state/
│   └── imports: interfaces/, errors/
├── checkpoint/
│   └── imports: interfaces/, errors/
├── middleware/
│   └── imports: interfaces/, errors/
└── NO imports from: agents/, tools/, middleware/, parsers/, builder/
```

**builder/**
```
builder/
├── builder.go
│   ├── imports: core/, llm/, store/, tools/, interfaces/, errors/
│   ├── imports: core/execution, core/middleware
│   └── cross-imports: memory/ (for type access)
└── NO imports from: agents/, middleware/, examples/
```

**llm/**
```
llm/
├── client.go
│   └── imports: interfaces/, errors/
├── providers/
│   └── imports: llm/ (parent), interfaces/, errors/
└── NO imports from: agents/, tools/, builder/
```

**memory/**
```
memory/
├── manager.go
│   └── imports: interfaces/, errors/
└── NO imports from: agents/, tools/, builder/
```

**store/**
```
store/
├── store.go
│   └── imports: interfaces/, errors/
├── memory/
│   └── imports: store/ (parent), interfaces/
├── redis/
│   └── imports: store/ (parent), interfaces/
├── postgres/
│   └── imports: store/ (parent), interfaces/
└── NO imports from: agents/, tools/, middleware/
```

**Other Layer 2:**
- `retrieval/` - imports: interfaces/, errors/, store/, llm/
- `observability/` - imports: interfaces/, errors/
- `performance/` - imports: interfaces/, errors/
- `planning/` - imports: interfaces/, errors/, core/
- `prompt/` - imports: interfaces/, errors/
- `reflection/` - imports: interfaces/, errors/

### Layer 3: Implementation (Depends on Layer 1 & 2)

**agents/**
```
agents/
├── executor/executor_agent.go
│   ├── imports: core/, interfaces/, tools/
│   └── imports: memory/, llm/ (Layer 2)
├── react/react_agent.go
│   ├── imports: core/, interfaces/
│   └── imports: parsers/ (same layer)
├── specialized/
│   ├── imports: core/, interfaces/
│   └── imports: tools/ (same layer)
└── NO imports from: builder/, middleware/, examples/
```

**tools/**
```
tools/
├── tool.go (interface only)
│   └── imports: interfaces/, errors/
├── registry.go
│   └── imports: tools.Tool (local), sync
├── shell/shell_tool.go
│   ├── imports: interfaces/, errors/
│   └── imports: core/ (for types)
├── http/http_tool.go
│   ├── imports: interfaces/, errors/
│   └── imports: cache/ (Layer 1)
└── NO imports from: agents/, middleware/, parsers/
```

**middleware/**
```
middleware/
├── observability.go
│   ├── imports: core/, core/middleware/
│   └── imports: observability/ (Layer 2)
├── tool_selector.go
│   ├── imports: core/, core/middleware/
│   └── imports: tools/ (same layer)
└── NO imports from: agents/, builder/, examples/
```

**parsers/**
```
parsers/
├── output_parser.go
│   └── imports: interfaces/, errors/
├── parser_react.go
│   └── imports: interfaces/, errors/, core/
└── NO imports from: agents/, tools/, middleware/
```

**stream/**
```
stream/
├── imports: interfaces/, errors/, core/
└── NO imports from: agents/, tools/, middleware/, builder/
```

**Other Layer 3:**
- `multiagent/` - imports: core/, interfaces/, agents/
- `distributed/` - imports: core/, interfaces/
- `mcp/` - imports: core/, interfaces/, tools/ (specialized)
- `document/` - imports: core/, interfaces/, store/, llm/
- `toolkits/` - imports: interfaces/, tools/

### Layer 4: Examples and Tests

**examples/**
```
examples/
├── basic/*.go
│   ├── imports: core/, builder/, agents/, tools/
│   └── imports: llm/, memory/, store/
├── advanced/*.go
│   ├── imports: everything
│   └── imports: middleware/, parsers/
└── integration/*.go
    ├── imports: anything (for comprehensive demos)
    └── exports: NEVER (no other code imports examples)
```

**Test Files (_test.go)**
```
All *_test.go files
├── imports: testing, testify, mocks
├── imports: all pkg/agent packages
└── cross-layer imports allowed
```

## Import Violation Detection Script

Create a file: `/Users/costalong/code/go/src/github.com/kart/k8s-agent/pkg/agent/verify_imports.sh`

```bash
#!/bin/bash
set -e

AGENT_PKG="/Users/costalong/code/go/src/github.com/kart/k8s-agent/pkg/agent"
VIOLATIONS=0

echo "=== Checking Import Layering Violations ==="
echo ""

# Rule 1: Layer 1 packages should not import other pkg/agent packages
echo "[1/5] Checking Layer 1 (interfaces, errors, cache, utils) dependencies..."
for pkg in interfaces errors cache utils; do
    if find "$AGENT_PKG/$pkg" -name "*.go" ! -name "*_test.go" -exec grep -l "pkg/agent/[^a-z]*\(interfaces\|errors\|cache\|utils\)" {} \; 2>/dev/null; then
        echo "  WARNING: $pkg imports another pkg/agent package (test files OK)"
    fi
done

# Rule 2: Layer 3 should not import Layer 4 (examples)
echo "[2/5] Checking Layer 3 doesn't import examples..."
violations=$(find "$AGENT_PKG" -path "*/examples" -prune -o -path "*_test.go" -prune -o -name "*.go" -exec grep -l "pkg/agent/examples" {} \; 2>/dev/null | grep -v examples | wc -l)
if [ "$violations" -gt 0 ]; then
    echo "  ERROR: Found $violations files importing examples in non-test code"
    find "$AGENT_PKG" -path "*/examples" -prune -o -path "*_test.go" -prune -o -name "*.go" -exec grep -l "pkg/agent/examples" {} \; 2>/dev/null | grep -v examples
    VIOLATIONS=$((VIOLATIONS + 1))
fi

# Rule 3: tools should not import agents
echo "[3/5] Checking tools don't import agents..."
violations=$(find "$AGENT_PKG/tools" -name "*.go" ! -name "*_test.go" -exec grep -l "pkg/agent/agents" {} \; 2>/dev/null | wc -l)
if [ "$violations" -gt 0 ]; then
    echo "  ERROR: Found $violations files in tools importing agents"
    VIOLATIONS=$((VIOLATIONS + 1))
fi

# Rule 4: parsers should not import tools or agents
echo "[4/5] Checking parsers don't import tools/agents..."
violations=$(find "$AGENT_PKG/parsers" -name "*.go" ! -name "*_test.go" -exec grep -l "pkg/agent/\(tools\|agents\)" {} \; 2>/dev/null | wc -l)
if [ "$violations" -gt 0 ]; then
    echo "  ERROR: Found $violations files in parsers importing tools/agents"
    VIOLATIONS=$((VIOLATIONS + 1))
fi

# Rule 5: No circular dependencies
echo "[5/5] Checking for circular dependencies..."
if go mod graph 2>/dev/null | grep -E "pkg/agent.*->.*pkg/agent.*->.*pkg/agent" | head -5; then
    echo "  WARNING: Possible circular dependencies detected (review above)"
fi

echo ""
if [ "$VIOLATIONS" -eq 0 ]; then
    echo "SUCCESS: All import layering rules verified!"
    exit 0
else
    echo "FAILURE: Found $VIOLATIONS rule violations"
    exit 1
fi
```

Usage:
```bash
chmod +x /Users/costalong/code/go/src/github.com/kart/k8s-agent/pkg/agent/verify_imports.sh
/Users/costalong/code/go/src/github.com/kart/k8s-agent/pkg/agent/verify_imports.sh
```

## Detailed Dependency Graph

```
┌────────────────────────────────────────────────────────────────────┐
│                    LAYER 1: FOUNDATIONS                            │
│ ┌──────────┐  ┌────────┐  ┌───────┐  ┌────────┐                    │
│ │interfaces│  │ errors │  │ cache │  │ utils  │                    │
│ └────┬─────┘  └───┬────┘  └───┬───┘  └───┬────┘                    │
│      └────────────┴───────────┴─────────┘                           │
│      (All import ONLY stdlib + external)                            │
└────────────────────────┬────────────────────────────────────────────┘
                         │ (One-way dependency: Layer 1 ← All)
┌────────────────────────▼────────────────────────────────────────────┐
│                  LAYER 2: BUSINESS LOGIC                            │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │              core/ (Foundation of execution)                │   │
│  ├─────────────────────────────────────────────────────────────┤   │
│  │ ├─ agent.go: BaseAgent implementation                       │   │
│  │ ├─ execution/: Execution engine                             │   │
│  │ ├─ state/: State management                                │   │
│  │ ├─ checkpoint/: Checkpointing logic                         │   │
│  │ ├─ middleware/: Middleware framework                        │   │
│  │ └─ callback/: Callback system                               │   │
│  │                                                             │   │
│  │ Imports: interfaces/, errors/, cache/                       │   │
│  │ Exports: Core types, BaseAgent, execution infrastructure    │   │
│  └────────────────────────────────────────────────────────────┘   │
│                                                                      │
│  ┌──────────────┐  ┌────────────┐  ┌──────────┐                   │
│  │  llm/        │  │  store/    │  │ memory/  │                   │
│  ├──────────────┤  ├────────────┤  ├──────────┤                   │
│  │ • client.go  │  │ • store.go │  │ • manager│                   │
│  │ • providers/ │  │ • memory/  │  │ • types  │                   │
│  └──────────────┘  │ • redis/   │  └──────────┘                   │
│                    │ • postgres/│                                   │
│                    └────────────┘                                   │
│                                                                      │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │         builder/ (Fluent API for agent construction)          │  │
│  ├──────────────────────────────────────────────────────────────┤  │
│  │ Imports: core/, llm/, store/, memory/, tools/                │  │
│  │ Exports: AgentBuilder, fluent configuration API              │  │
│  └──────────────────────────────────────────────────────────────┘  │
│                                                                      │
│  Other: retrieval/, observability/, performance/, planning/,       │
│         prompt/, reflection/                                        │
│                                                                      │
│  All Layer 2 packages:                                              │
│  • Import from: Layer 1 only                                        │
│  • Cross-imports: Allowed (carefully managed)                       │
│  • Export to: Layer 3, Builder provides fluent API                 │
│                                                                      │
└──────────────────────────┬─────────────────────────────────────────┘
                           │ (One-way dependency: Layer 3 → Layer 2)
┌──────────────────────────▼─────────────────────────────────────────┐
│                   LAYER 3: IMPLEMENTATION                           │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌──────────────────────────────────────────────────────────────┐  │
│  │           agents/ (Agent implementations)                    │  │
│  ├──────────────────────────────────────────────────────────────┤  │
│  │ executor/   → Tool execution agent                           │  │
│  │ react/      → ReAct reasoning agent                          │  │
│  │ specialized/→ Domain-specific agents                         │  │
│  │                                                              │  │
│  │ Imports: core/, interfaces/, tools/, memory/, llm/           │  │
│  │ May import: parsers/ (same layer)                            │  │
│  │ Exports: Specific agent implementations                      │  │
│  └──────────────────────────────────────────────────────────────┘  │
│                                                                      │
│  ┌──────────────────┐  ┌─────────────────┐  ┌─────────────────┐   │
│  │ tools/           │  │ middleware/     │  │ parsers/        │   │
│  ├──────────────────┤  ├─────────────────┤  ├─────────────────┤   │
│  │ shell/           │  │ observability.go│  │ output_parser.go│   │
│  │ http/            │  │ tool_selector.go│  │ parser_react.go │   │
│  │ search/          │  │ cache_mw.go     │  └─────────────────┘   │
│  │ practical/       │  └─────────────────┘                         │
│  │ registry.go      │                                              │
│  └──────────────────┘                                              │
│                                                                      │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │ Other: stream/, multiagent/, distributed/, mcp/,            │   │
│  │ document/, toolkits/                                         │   │
│  └─────────────────────────────────────────────────────────────┘   │
│                                                                      │
│  All Layer 3 packages:                                              │
│  • Import from: Layer 1, Layer 2                                    │
│  • Cross-imports: Limited (documented exceptions)                   │
│  • CANNOT import: Layer 4 (examples), upward in Layer 3            │
│  • Exports: Specific implementations                                │
│                                                                      │
└──────────────────────────┬─────────────────────────────────────────┘
                           │ (One-way dependency: Layer 4 → Layer 3)
┌──────────────────────────▼─────────────────────────────────────────┐
│                  LAYER 4: EXAMPLES & TESTS                          │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  examples/basic/     → Basic usage patterns                         │
│  examples/advanced/  → Advanced patterns                            │
│  examples/integration/→ Integration examples                        │
│                                                                      │
│  *_test.go files     → Unit and integration tests                  │
│                                                                      │
│  Can import: ALL layers (for teaching/testing)                      │
│  Cannot export: Nothing imports from examples                       │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

## Common Refactoring Scenarios

### Scenario 1: Moving Code Between Layers

**Problem:** You have code in Layer 3 that belongs in Layer 2.

**Solution:**
1. Create/update Layer 2 package (e.g., `planning/`)
2. Move code to Layer 2
3. Update Layer 3 imports to use Layer 2
4. Keep type aliases in Layer 3 for backward compatibility if needed

```go
// Before: In agents/planning.go (Layer 3)
package agents
func PlanExecution() { /* ... */ }

// After: In planning/executor.go (Layer 2)
package planning
func ExecutionPlan() { /* ... */ }

// Backward compat: In agents/planning.go (Layer 3)
package agents
import "github.com/kart-io/k8s-agent/pkg/agent/planning"
var PlanExecution = planning.ExecutionPlan
```

### Scenario 2: Circular Dependency

**Problem:** Package A imports B, B imports A.

**Solution:** Extract common interface to Layer 1.

```go
// Problem:
// core/agent.go imports builder/
// builder/builder.go imports core/

// Solution: Define interface in Layer 1
// interfaces/builder.go - New interface
type Builder interface {
    Build(ctx context.Context) (Agent, error)
}

// Now both can depend on interfaces without circular imports
```

### Scenario 3: Tool Needs Feature from Middleware

**Problem:** `tools/my_tool.go` needs functionality from `middleware/`.

**Solution:** Extract to Layer 2 if it's general purpose.

```go
// Problem:
// tools/my_tool.go wants middleware functionality

// Solution:
// 1. If it's general: Create performance/ or observability/ in Layer 2
// 2. Move code to Layer 2
// 3. tools/ imports from Layer 2

// Before:
import "github.com/kart-io/k8s-agent/pkg/agent/middleware"

// After:
import "github.com/kart-io/k8s-agent/pkg/agent/observability"
```

## Import Audit Checklist

When reviewing pull requests or adding new packages:

- [ ] New package has clear purpose within correct layer
- [ ] Package only imports from allowed layers
- [ ] No circular imports between packages
- [ ] No imports from Layer 4 (examples) in production code
- [ ] Test files properly named (`*_test.go`)
- [ ] Interfaces defined in Layer 1 if cross-layer
- [ ] Type aliases created for backward compatibility if needed
- [ ] Documentation updated (ARCHITECTURE.md)
- [ ] Examples added showing correct usage
- [ ] `verify_imports.sh` passes all checks

## Monitoring and Metrics

Track these metrics over time:

1. **Number of import rule violations** (should be 0)
2. **Average dependency depth** (layer count imports traverse)
3. **Cyclomatic import complexity** (should be low)
4. **Package coupling** (how many packages import each package)

```bash
# Example: Count imports per package
for pkg in core llm store memory builder agents tools middleware; do
    count=$(grep -r "import.*pkg/agent/$pkg" /Users/costalong/code/go/src/github.com/kart/k8s-agent/pkg/agent/ --include="*.go" ! -path "*/examples/*" ! -path "*_test.go" | wc -l)
    echo "$pkg: $count imports"
done
```

## References

- Main Architecture: `/Users/costalong/code/go/src/github.com/kart/k8s-agent/pkg/agent/ARCHITECTURE.md`
- Layer 1 Definitions: `/Users/costalong/code/go/src/github.com/kart/k8s-agent/pkg/agent/interfaces/`
- Migration Guide: `/Users/costalong/code/go/src/github.com/kart/k8s-agent/pkg/agent/MIGRATION_GUIDE.md`
