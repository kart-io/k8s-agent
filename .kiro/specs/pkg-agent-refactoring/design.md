# Design Document: pkg/agent Directory Refactoring

## Overview

This document provides a detailed technical design for refactoring the pkg/agent directory structure. The refactoring addresses organizational debt while maintaining full backward compatibility and follows a phased approach to minimize risk.

## 1. Package Structure Design

### 1.1 Current Structure (Before Refactoring)

```text
pkg/agent/
├── *.md (26 files - chaos)           # Documentation scattered
├── README.md
├── ARCHITECTURE.md
├── core/ (24 files, 9465 lines)      # Bloated package
│   ├── agent.go
│   ├── chain.go
│   ├── checkpointer*.go (5 files)    # Should be sub-package
│   ├── state*.go (2 files)           # Should be sub-package
│   ├── runtime.go
│   ├── middleware*.go (3 files)
│   ├── orchestrator.go
│   ├── streaming.go
│   └── ...
├── tools/
│   ├── executor_tool.go              # Misplaced Agent logic
│   ├── runtime.go                    # Name collision
│   └── ...
├── retrieval/
│   └── vector_store.go               # VectorStore interface #1
├── memory/
│   └── manager.go                    # VectorStore interface #2
├── example/ (15+ dirs)               # Disorganized examples
│   ├── basic/
│   ├── langchain_*/
│   ├── main.go
│   └── ...
└── ... (69 total directories)
```

**Critical Issues**:

- Core package: 24 files, 9,465 lines (target: ≤15 files, ≤5,000 lines)
- 26 Markdown files in root (target: 2 files)
- Duplicate filenames: runtime.go (2×), config.go (2×), main.go (21×)
- VectorStore interface defined in 2 locations
- Agent implementation in tools package

### 1.2 Target Structure (After Refactoring)

```text
pkg/agent/
├── README.md                         # Main documentation
├── ARCHITECTURE.md                   # Architecture overview
│
├── docs/                             # Organized documentation
│   ├── archive/                      # Completed implementation docs
│   │   ├── human-in-loop-complete.md
│   │   ├── parallel-execution-complete.md
│   │   ├── streaming-complete.md
│   │   └── ... (8 files)
│   ├── analysis/                     # Analysis and planning docs
│   │   ├── code-structure-analysis.md
│   │   ├── comprehensive-analysis.md
│   │   └── ... (4 files)
│   ├── refactoring/                  # Refactoring documentation
│   │   ├── refactoring-guide.md
│   │   ├── refactoring-complete.md
│   │   └── migration-guide.md
│   └── guides/                       # User guides
│       ├── quickstart-improvements.md
│       └── langchain-improvements.md
│
├── interfaces/                       # Unified interface package (NEW)
│   ├── agent.go                      # Agent, Runnable interfaces
│   ├── store.go                      # Store, VectorStore (canonical)
│   ├── memory.go                     # Memory interfaces
│   ├── tool.go                       # Tool interfaces
│   ├── checkpoint.go                 # Checkpointer interfaces
│   └── doc.go                        # Package documentation
│
├── core/                             # Refactored core (≤15 files)
│   ├── agent.go                      # Core Agent implementation
│   ├── agent_test.go
│   ├── chain.go                      # Chain abstraction
│   ├── chain_test.go
│   ├── runnable.go                   # Runnable implementation
│   ├── orchestrator.go               # High-level orchestration
│   ├── callback.go                   # Callback system
│   ├── errors.go                     # Error definitions
│   ├── interrupt.go                  # Interrupt handling
│   ├── interrupt_test.go
│   │
│   ├── state/                        # State management sub-package
│   │   ├── state.go                  # State types and operations
│   │   ├── state_test.go
│   │   ├── manager.go                # State lifecycle
│   │   └── serializer.go             # State serialization
│   │
│   ├── checkpoint/                   # Checkpointing sub-package
│   │   ├── checkpointer.go           # Base interface/types
│   │   ├── checkpointer_test.go
│   │   ├── memory.go                 # In-memory implementation
│   │   ├── redis.go                  # Redis implementation
│   │   ├── redis_test.go
│   │   ├── distributed.go            # Distributed checkpointer
│   │   └── saver.go                  # Checkpoint saving logic
│   │
│   ├── execution/                    # Execution runtime sub-package
│   │   ├── runtime.go                # Agent runtime (renamed from core/)
│   │   ├── runtime_test.go
│   │   ├── executor.go               # Execution coordinator
│   │   ├── context.go                # Execution context
│   │   └── streaming.go              # Streaming execution
│   │
│   └── middleware/                   # Middleware sub-package
│       ├── middleware.go             # Core middleware types
│       ├── middleware_test.go
│       ├── advanced.go               # Advanced middleware
│       ├── chain.go                  # Middleware chaining
│       └── builtin.go                # Built-in middleware
│
├── agents/                           # Agent implementations
│   ├── executor/                     # Executor agent (moved from tools/)
│   │   ├── executor_agent.go         # Renamed from executor_tool.go
│   │   ├── executor_agent_test.go
│   │   ├── config.go
│   │   └── options.go
│   ├── react/
│   │   └── react_agent.go
│   └── specialized/
│       └── specialized_agent.go
│
├── tools/                            # Tool definitions only
│   ├── tool.go                       # Base tool types
│   ├── tool_test.go
│   ├── registry.go
│   ├── tool_runtime.go               # Renamed from runtime.go
│   ├── compute/
│   ├── http/
│   ├── practical/
│   ├── search/
│   └── shell/
│
├── retrieval/                        # RAG and retrieval
│   ├── vector_store.go               # VectorStore alias → interfaces/
│   ├── retriever.go
│   ├── embeddings.go
│   ├── rag.go
│   └── ...
│
├── memory/                           # Memory management
│   ├── manager.go                    # Memory manager (uses interfaces/)
│   ├── conversation.go
│   ├── case.go
│   └── ...
│
├── store/                            # State store implementations
│   ├── store.go                      # Store alias → interfaces/
│   ├── adapters/
│   ├── memory/
│   ├── postgres/
│   │   ├── postgres_store.go         # Renamed from config.go
│   │   └── ...
│   └── redis/
│       ├── redis_store.go            # Renamed from config.go
│       └── ...
│
├── examples/                         # Reorganized examples
│   ├── basic/                        # Single-feature examples
│   │   ├── 01-simple-agent/
│   │   │   └── simple_agent.go       # Renamed from main.go
│   │   ├── 02-chain/
│   │   │   └── chain_demo.go
│   │   ├── 03-tools/
│   │   │   └── tools_demo.go
│   │   └── README.md
│   │
│   ├── advanced/                     # Multi-feature examples
│   │   ├── streaming/
│   │   │   └── streaming_demo.go     # Renamed from main.go
│   │   ├── multi-mode-streaming/
│   │   │   └── multi_mode_demo.go
│   │   ├── observability/
│   │   │   └── observability_demo.go
│   │   ├── react/
│   │   │   └── react_demo.go
│   │   └── README.md
│   │
│   └── integration/                  # Full-system examples
│       ├── langchain-inspired/
│       │   └── langchain_demo.go     # Renamed from main.go
│       ├── multiagent/
│       │   └── multiagent_demo.go
│       ├── human-in-loop/
│       │   └── hitl_demo.go
│       └── README.md
│
├── llm/                              # LLM client
├── mcp/                              # MCP integration
├── multiagent/                       # Multi-agent systems
├── observability/                    # Metrics, tracing
├── planning/                         # Planning algorithms
├── prompt/                           # Prompt templates
├── stream/                           # Streaming utilities
└── ... (other packages unchanged)
```

**Improvements**:

- Core package: 24 → 10 files (-58%), 9,465 → ~2,500 lines (-73%)
- Root Markdown: 26 → 0 files (-100%, moved to docs/)
- Zero filename collisions (100% elimination)
- Single source of truth for interfaces
- Clear separation: agents/ vs tools/

### 1.3 Package Responsibilities

| Package | Responsibility | Size Target | Dependencies |
|---------|---------------|-------------|--------------|
| `interfaces/` | Canonical interface definitions | 6 files, <500 lines | None (foundation) |
| `core/` | Core orchestration, agent base | 10 files, <2500 lines | interfaces/ |
| `core/state/` | State management | 4 files, <800 lines | interfaces/ |
| `core/checkpoint/` | Checkpoint persistence | 6 files, <2000 lines | interfaces/, state/ |
| `core/execution/` | Runtime and execution | 5 files, <1500 lines | interfaces/, state/ |
| `core/middleware/` | Middleware chain | 5 files, <1200 lines | interfaces/ |
| `agents/` | Agent implementations | Per-agent | core/, interfaces/ |
| `tools/` | Tool definitions & registry | Current size | interfaces/ |
| `retrieval/` | RAG and vector search | Current size | interfaces/ |
| `memory/` | Memory management | Current size | interfaces/ |
| `store/` | Store implementations | Current size | interfaces/ |

## 2. File Migration Mapping

### 2.1 Core Package Reorganization

#### Phase 1: Extract Sub-packages

| Source File | Target File | Rationale |
|------------|-------------|-----------|
| `core/state.go` | `core/state/state.go` | State management sub-package |
| `core/state_test.go` | `core/state/state_test.go` | Test follows code |
| `core/checkpointer.go` | `core/checkpoint/checkpointer.go` | Checkpoint sub-package |
| `core/checkpointer_test.go` | `core/checkpoint/checkpointer_test.go` | Test follows code |
| `core/checkpointer_redis.go` | `core/checkpoint/redis.go` | More concise name |
| `core/checkpointer_redis_test.go` | `core/checkpoint/redis_test.go` | Test follows code |
| `core/checkpointer_distributed.go` | `core/checkpoint/distributed.go` | More concise name |
| `core/runtime.go` | `core/execution/runtime.go` | Execution sub-package |
| `core/runtime_test.go` | `core/execution/runtime_test.go` | Test follows code |
| `core/streaming.go` | `core/execution/streaming.go` | Streaming is execution concern |
| `core/middleware.go` | `core/middleware/middleware.go` | Middleware sub-package |
| `core/middleware_test.go` | `core/middleware/middleware_test.go` | Test follows code |
| `core/middleware_advanced.go` | `core/middleware/advanced.go` | More concise name |

#### Phase 2: Core Root Files (Remain)

| File | Lines | Purpose | Changes |
|------|-------|---------|---------|
| `core/agent.go` | ~600 | Core Agent implementation | Update imports |
| `core/agent_test.go` | ~400 | Agent tests | Update imports |
| `core/chain.go` | ~400 | Chain abstraction | Update imports |
| `core/chain_test.go` | ~200 | Chain tests | Update imports |
| `core/chain_example_test.go` | ~150 | Chain examples | Update imports |
| `core/runnable.go` | ~300 | Runnable interface impl | Update imports |
| `core/orchestrator.go` | ~250 | High-level orchestration | Update imports |
| `core/callback.go` | ~150 | Callback system | Update imports |
| `core/errors.go` | ~50 | Error definitions | None |
| `core/interrupt.go` | ~200 | Interrupt handling | None |
| `core/interrupt_test.go` | ~150 | Interrupt tests | None |
| **Total** | **~2,850** | **11 files** | **Within 5K target** |

### 2.2 Interface Unification

#### New Files Created

| Target File | Extracted From | Purpose |
|------------|----------------|---------|
| `interfaces/agent.go` | `core/agent.go` | Agent, Runnable interfaces |
| `interfaces/store.go` | `retrieval/vector_store.go`, `memory/manager.go` | VectorStore, Store interfaces |
| `interfaces/memory.go` | `memory/manager.go` | Memory manager interfaces |
| `interfaces/tool.go` | `tools/tool.go` | Tool interfaces |
| `interfaces/checkpoint.go` | `core/checkpoint/checkpointer.go` | Checkpointer interface |
| `interfaces/doc.go` | New | Package documentation |

#### Backward Compatibility Aliases

| Old Location | New Location | Compatibility |
|-------------|--------------|---------------|
| `retrieval.VectorStore` | `interfaces.VectorStore` | Type alias for 1 version |
| `memory.Manager` | `interfaces.MemoryManager` | Type alias for 1 version |
| `core.Agent` | `interfaces.Agent` | Type alias for 1 version |
| `core.Runnable` | `interfaces.Runnable` | Type alias for 1 version |
| `core.Checkpointer` | `interfaces.Checkpointer` | Type alias for 1 version |

### 2.3 Agent/Tool Separation

| Source File | Target File | Rationale |
|------------|-------------|-----------|
| `tools/executor_tool.go` | `agents/executor/executor_agent.go` | Implements Agent interface |
| `tools/runtime.go` | `tools/tool_runtime.go` | Eliminate name collision |

### 2.4 Documentation Reorganization

| Source File | Target File | Category |
|------------|-------------|----------|
| `HUMAN_IN_THE_LOOP_IMPLEMENTATION_COMPLETE.md` | `docs/archive/human-in-loop-complete.md` | Archive |
| `PARALLEL_TOOL_EXECUTION_IMPLEMENTATION_COMPLETE.md` | `docs/archive/parallel-execution-complete.md` | Archive |
| `MULTI_MODE_STREAMING_IMPLEMENTATION_COMPLETE.md` | `docs/archive/streaming-complete.md` | Archive |
| `TOOLRUNTIME_IMPLEMENTATION_COMPLETE.md` | `docs/archive/tool-runtime-complete.md` | Archive |
| `TOOL_SELECTOR_MIDDLEWARE_IMPLEMENTATION_COMPLETE.md` | `docs/archive/tool-selector-complete.md` | Archive |
| `LANGCHAIN_INSPIRED_IMPROVEMENTS.md` | `docs/archive/langchain-inspired-complete.md` | Archive |
| `PROJECT_COMPLETION_SUMMARY.md` | `docs/archive/project-summary.md` | Archive |
| `IMPLEMENTATION_SUMMARY.md` | `docs/archive/implementation-summary.md` | Archive |
| `CODE_STRUCTURE_ANALYSIS.md` | `docs/analysis/code-structure.md` | Analysis |
| `COMPREHENSIVE_ANALYSIS.md` | `docs/analysis/comprehensive.md` | Analysis |
| `ANALYSIS_INDEX.md` | `docs/analysis/index.md` | Analysis |
| `ANALYSIS_DOCUMENTS_INDEX.md` | `docs/analysis/documents-index.md` | Analysis |
| `REFACTORING_GUIDE.md` | `docs/refactoring/guide.md` | Refactoring |
| `REFACTORING_COMPLETE.md` | `docs/refactoring/complete.md` | Refactoring |
| `QUICKSTART_IMPROVEMENTS.md` | `docs/guides/quickstart.md` | Guides |
| `LANGCHAIN_IMPROVEMENTS.md` | `docs/guides/langchain.md` | Guides |
| `LANGCHAIN_IMPROVEMENTS_SUMMARY.md` | `docs/guides/langchain-summary.md` | Guides |
| `LANGCHAIN_V2_IMPROVEMENT_PLAN.md` | `docs/guides/langchain-v2-plan.md` | Guides |
| `LANGCHAIN_FINAL_SUMMARY.md` | `docs/guides/langchain-final.md` | Guides |
| `COMPLETE_IMPLEMENTATION_SUMMARY.md` | `docs/archive/complete-summary.md` | Archive |
| `README.md` | `README.md` | Keep (updated) |
| `ARCHITECTURE.md` | `ARCHITECTURE.md` | Keep (updated) |

**New Documentation**:

- `docs/refactoring/migration-guide.md` - Guide for migrating to new structure
- `docs/README.md` - Documentation index
- `examples/basic/README.md` - Basic examples guide
- `examples/advanced/README.md` - Advanced examples guide
- `examples/integration/README.md` - Integration examples guide

### 2.5 Example Reorganization

| Source Path | Target Path | Category |
|------------|-------------|----------|
| `example/basic/` | `examples/basic/01-simple-agent/` | Basic |
| `example/tools/main.go` | `examples/basic/03-tools/tools_demo.go` | Basic |
| `example/streaming/main.go` | `examples/advanced/streaming/streaming_demo.go` | Advanced |
| `example/multi_mode_streaming/main.go` | `examples/advanced/multi-mode-streaming/multi_mode_demo.go` | Advanced |
| `example/observability/main.go` | `examples/advanced/observability/observability_demo.go` | Advanced |
| `example/react_example/main.go` | `examples/advanced/react/react_demo.go` | Advanced |
| `example/parallel_execution/main.go` | `examples/advanced/parallel-execution/parallel_demo.go` | Advanced |
| `example/tool_runtime/main.go` | `examples/advanced/tool-runtime/runtime_demo.go` | Advanced |
| `example/tool_selector/main.go` | `examples/advanced/tool-selector/selector_demo.go` | Advanced |
| `example/langchain_inspired/main.go` | `examples/integration/langchain-inspired/langchain_demo.go` | Integration |
| `example/langchain_complete/main.go` | `examples/integration/langchain-complete/complete_demo.go` | Integration |
| `example/langchain_phase1/main.go` | `examples/integration/langchain-phase1/phase1_demo.go` | Integration |
| `example/langchain_phase2/main.go` | `examples/integration/langchain-phase2/phase2_demo.go` | Integration |
| `example/multiagent/main.go` | `examples/integration/multiagent/multiagent_demo.go` | Integration |
| `example/human_in_the_loop/main.go` | `examples/integration/human-in-loop/hitl_demo.go` | Integration |
| `example/preconfig_agents/main.go` | `examples/integration/preconfig-agents/preconfig_demo.go` | Integration |

## 3. Interface Unification Architecture

### 3.1 Canonical Interface Package Design

**File: `interfaces/agent.go`**

```go
// Package interfaces provides canonical interface definitions for the agent framework.
//
// All concrete implementations should reference these interfaces to ensure
// type compatibility across packages.
package interfaces

import "context"

// Agent represents an autonomous agent that can process inputs and produce outputs.
//
// All agent implementations (react, executor, specialized) should implement this interface.
type Agent interface {
    Runnable

    // Name returns the agent's identifier.
    Name() string

    // Description returns what the agent does.
    Description() string

    // Plan generates an execution plan for the given input.
    Plan(ctx context.Context, input *Input) (*Plan, error)
}

// Runnable represents any component that can be invoked with input to produce output.
//
// This is the foundation interface implemented by agents, chains, and tools.
type Runnable interface {
    // Invoke executes the runnable with the given input.
    Invoke(ctx context.Context, input *Input) (*Output, error)

    // Stream executes with streaming output support.
    Stream(ctx context.Context, input *Input) (<-chan *StreamChunk, error)
}

// Input represents standardized input to a runnable.
type Input struct {
    Messages []Message
    State    State
    Config   map[string]interface{}
}

// Output represents standardized output from a runnable.
type Output struct {
    Messages []Message
    State    State
    Metadata map[string]interface{}
}

// Message represents a single message in a conversation.
type Message struct {
    Role    string
    Content string
    Name    string
}

// StreamChunk represents a chunk of streaming output.
type StreamChunk struct {
    Content  string
    Metadata map[string]interface{}
    Done     bool
}

// Plan represents an agent's execution plan.
type Plan struct {
    Steps    []Step
    Metadata map[string]interface{}
}

// Step represents a single step in an execution plan.
type Step struct {
    Action   string
    Input    map[string]interface{}
    ToolName string
}
```

**File: `interfaces/store.go`**

```go
package interfaces

import "context"

// VectorStore is the canonical interface for vector storage and similarity search.
//
// Implementations: memory.MemoryVectorStore, qdrant.QdrantStore, etc.
//
// Previously defined in: retrieval/vector_store.go, memory/manager.go
type VectorStore interface {
    // SimilaritySearch performs vector similarity search.
    SimilaritySearch(ctx context.Context, query string, topK int) ([]*Document, error)

    // SimilaritySearchWithScore returns documents with similarity scores.
    SimilaritySearchWithScore(ctx context.Context, query string, topK int) ([]*Document, error)

    // AddDocuments adds documents to the vector store.
    AddDocuments(ctx context.Context, docs []*Document) error

    // Delete removes documents by ID.
    Delete(ctx context.Context, ids []string) error
}

// Store is the canonical interface for general key-value storage.
//
// Implementations: memory.MemoryStore, postgres.PostgresStore, redis.RedisStore
type Store interface {
    // Get retrieves a value by key.
    Get(ctx context.Context, key string) (interface{}, error)

    // Set stores a key-value pair.
    Set(ctx context.Context, key string, value interface{}) error

    // Delete removes a key.
    Delete(ctx context.Context, key string) error

    // Clear removes all keys.
    Clear(ctx context.Context) error
}

// Document represents a document with optional vector embedding.
type Document struct {
    ID          string
    PageContent string
    Metadata    map[string]interface{}
    Embedding   []float64
    Score       float64
}

// State represents agent state that can be persisted.
type State map[string]interface{}
```

**File: `interfaces/checkpoint.go`**

```go
package interfaces

import (
    "context"
    "time"
)

// Checkpointer is the canonical interface for saving/loading agent state.
//
// Implementations: checkpoint.MemoryCheckpointer, checkpoint.RedisCheckpointer,
// checkpoint.DistributedCheckpointer
type Checkpointer interface {
    // SaveCheckpoint persists a checkpoint.
    SaveCheckpoint(ctx context.Context, checkpoint *Checkpoint) error

    // LoadCheckpoint retrieves a checkpoint.
    LoadCheckpoint(ctx context.Context, checkpointID string) (*Checkpoint, error)

    // ListCheckpoints lists checkpoints for a thread.
    ListCheckpoints(ctx context.Context, threadID string, limit int) ([]*CheckpointMetadata, error)

    // DeleteCheckpoint removes a checkpoint.
    DeleteCheckpoint(ctx context.Context, checkpointID string) error
}

// Checkpoint represents a saved state snapshot.
type Checkpoint struct {
    ID        string
    ThreadID  string
    State     State
    Metadata  map[string]interface{}
    CreatedAt time.Time
}

// CheckpointMetadata contains checkpoint summary information.
type CheckpointMetadata struct {
    ID        string
    ThreadID  string
    CreatedAt time.Time
    Size      int64
}
```

**File: `interfaces/tool.go`**

```go
package interfaces

import "context"

// Tool represents an executable tool that agents can invoke.
//
// All tool implementations should implement this interface.
type Tool interface {
    // Name returns the tool identifier.
    Name() string

    // Description returns what the tool does.
    Description() string

    // Invoke executes the tool with given input.
    Invoke(ctx context.Context, input *ToolInput) (*ToolOutput, error)

    // Schema returns the tool's input schema (JSON Schema format).
    Schema() map[string]interface{}
}

// ToolInput represents tool execution input.
type ToolInput struct {
    Args    map[string]interface{}
    Context context.Context
}

// ToolOutput represents tool execution output.
type ToolOutput struct {
    Result   interface{}
    Metadata map[string]interface{}
    Error    error
}
```

**File: `interfaces/memory.go`**

```go
package interfaces

import (
    "context"
    "time"
)

// MemoryManager is the canonical interface for agent memory management.
//
// Previously: memory.Manager
type MemoryManager interface {
    // AddConversation stores a conversation turn.
    AddConversation(ctx context.Context, conv *Conversation) error

    // GetConversationHistory retrieves conversation history.
    GetConversationHistory(ctx context.Context, sessionID string, limit int) ([]*Conversation, error)

    // ClearConversation removes conversation history.
    ClearConversation(ctx context.Context, sessionID string) error

    // AddCase stores a case for case-based reasoning.
    AddCase(ctx context.Context, caseMemory *Case) error

    // SearchSimilarCases finds similar cases.
    SearchSimilarCases(ctx context.Context, query string, limit int) ([]*Case, error)

    // Store persists arbitrary key-value data.
    Store(ctx context.Context, key string, value interface{}) error

    // Retrieve fetches stored data.
    Retrieve(ctx context.Context, key string) (interface{}, error)

    // Delete removes stored data.
    Delete(ctx context.Context, key string) error

    // Clear removes all memory.
    Clear(ctx context.Context) error
}

// Conversation represents a conversation turn.
type Conversation struct {
    ID        string
    SessionID string
    Role      string
    Content   string
    Timestamp time.Time
    Metadata  map[string]interface{}
}

// Case represents a stored case for reasoning.
type Case struct {
    ID          string
    Title       string
    Description string
    Problem     string
    Solution    string
    Category    string
    Tags        []string
    Embedding   []float64
    Similarity  float64
    CreatedAt   time.Time
    UpdatedAt   time.Time
    Metadata    map[string]interface{}
}
```

### 3.2 Backward Compatibility Layer

**File: `retrieval/vector_store.go` (after refactoring)**

```go
package retrieval

import "github.com/kart-io/k8s-agent/pkg/agent/interfaces"

// VectorStore is deprecated. Use interfaces.VectorStore instead.
//
// This type alias provides backward compatibility for one major version.
// Migration: import "github.com/kart-io/k8s-agent/pkg/agent/interfaces"
//
// Deprecated: Use interfaces.VectorStore
type VectorStore = interfaces.VectorStore

// Document is deprecated. Use interfaces.Document instead.
//
// Deprecated: Use interfaces.Document
type Document = interfaces.Document

// MockVectorStore remains in this package for testing purposes.
type MockVectorStore struct {
    // ... existing implementation
}

// Ensure MockVectorStore implements the canonical interface
var _ interfaces.VectorStore = (*MockVectorStore)(nil)
```

**File: `memory/manager.go` (after refactoring)**

```go
package memory

import "github.com/kart-io/k8s-agent/pkg/agent/interfaces"

// Manager is deprecated. Use interfaces.MemoryManager instead.
//
// Deprecated: Use interfaces.MemoryManager
type Manager = interfaces.MemoryManager

// Conversation is deprecated. Use interfaces.Conversation instead.
//
// Deprecated: Use interfaces.Conversation
type Conversation = interfaces.Conversation

// Case is deprecated. Use interfaces.Case instead.
//
// Deprecated: Use interfaces.Case
type Case = interfaces.Case

// DefaultManager implements MemoryManager.
type DefaultManager struct {
    // ... existing implementation
}

// Ensure DefaultManager implements the canonical interface
var _ interfaces.MemoryManager = (*DefaultManager)(nil)
```

**File: `core/agent.go` (after refactoring)**

```go
package core

import "github.com/kart-io/k8s-agent/pkg/agent/interfaces"

// Agent is deprecated. Use interfaces.Agent instead.
//
// Deprecated: Use interfaces.Agent
type Agent = interfaces.Agent

// Runnable is deprecated. Use interfaces.Runnable instead.
//
// Deprecated: Use interfaces.Runnable
type Runnable = interfaces.Runnable

// BaseAgent implements the canonical Agent interface.
type BaseAgent struct {
    // ... existing implementation
}

// Ensure BaseAgent implements the canonical interface
var _ interfaces.Agent = (*BaseAgent)(nil)
```

### 3.3 Migration Path

**Phase 1** (Immediate):

1. Create `interfaces/` package with all canonical definitions
2. Add type aliases in original locations
3. Update all internal imports to use `interfaces/`
4. All tests pass with no behavior change

**Phase 2** (1-2 versions later):

1. Add deprecation warnings to godoc
2. Update examples to use `interfaces/` directly
3. Publish migration guide

**Phase 3** (Next major version):

1. Remove type aliases
2. Force migration to `interfaces/`

## 4. Core Package Decomposition

### 4.1 Sub-package Organization

#### `core/state/` - State Management

**Purpose**: All state-related types and operations.

**Files**:

- `state.go` - Core state types (State, StateUpdate, StateDiff)
- `state_test.go` - State tests
- `manager.go` - State lifecycle management
- `serializer.go` - State serialization/deserialization

**Exports**:

```go
type State map[string]interface{}
type StateUpdate struct { ... }
type StateDiff struct { ... }

func NewState() State
func (s State) Get(key string) (interface{}, bool)
func (s State) Set(key string, value interface{})
func (s State) Merge(other State) State
func (s State) Clone() State
```

**Lines**: ~800 (extracted from core/state.go ~500 lines + new code)

#### `core/checkpoint/` - Checkpointing

**Purpose**: Checkpoint persistence and recovery.

**Files**:

- `checkpointer.go` - Base interface and types (moved to interfaces/)
- `checkpointer_test.go` - Base tests
- `memory.go` - In-memory checkpointer (extracted from checkpointer.go)
- `redis.go` - Redis checkpointer (renamed from checkpointer_redis.go)
- `redis_test.go` - Redis tests
- `distributed.go` - Distributed checkpointer (renamed)
- `saver.go` - Checkpoint saving logic

**Exports**:

```go
// Interface in interfaces/checkpoint.go
type Checkpointer = interfaces.Checkpointer

// Implementations
type MemoryCheckpointer struct { ... }
type RedisCheckpointer struct { ... }
type DistributedCheckpointer struct { ... }

func NewMemoryCheckpointer() *MemoryCheckpointer
func NewRedisCheckpointer(config RedisConfig) *RedisCheckpointer
func NewDistributedCheckpointer(config DistConfig) *DistributedCheckpointer
```

**Lines**: ~2000 (5 checkpoint files ~1800 lines + refactoring)

#### `core/execution/` - Execution Runtime

**Purpose**: Agent execution and runtime management.

**Files**:

- `runtime.go` - Agent runtime (moved from core/runtime.go)
- `runtime_test.go` - Runtime tests
- `executor.go` - Execution coordinator (new)
- `context.go` - Execution context (new)
- `streaming.go` - Streaming execution (moved from core/streaming.go)

**Exports**:

```go
type Runtime struct { ... }
type ExecutionContext struct { ... }
type StreamingExecutor struct { ... }

func NewRuntime(opts ...Option) *Runtime
func (r *Runtime) Execute(ctx context.Context, agent Agent, input *Input) (*Output, error)
func (r *Runtime) ExecuteStream(ctx context.Context, agent Agent, input *Input) (<-chan *StreamChunk, error)
```

**Lines**: ~1500 (runtime.go ~600 + streaming.go ~300 + new ~600)

#### `core/middleware/` - Middleware System

**Purpose**: Request/response middleware.

**Files**:

- `middleware.go` - Core middleware types (moved from core/)
- `middleware_test.go` - Middleware tests
- `advanced.go` - Advanced middleware (renamed from middleware_advanced.go)
- `chain.go` - Middleware chaining (new)
- `builtin.go` - Built-in middleware (new)

**Exports**:

```go
type Middleware interface { ... }
type MiddlewareChain struct { ... }

// Built-in middleware
func LoggingMiddleware() Middleware
func TracingMiddleware() Middleware
func RateLimitMiddleware(limit int) Middleware
func RetryMiddleware(maxRetries int) Middleware
```

**Lines**: ~1200 (3 middleware files ~900 + new ~300)

### 4.2 Core Root Package (Remains)

**Purpose**: High-level orchestration and agent base implementation.

**Files** (11 total, ~2850 lines):

1. `agent.go` (~600 lines) - BaseAgent implementation
2. `agent_test.go` (~400 lines) - Agent tests
3. `chain.go` (~400 lines) - Chain abstraction
4. `chain_test.go` (~200 lines) - Chain tests
5. `chain_example_test.go` (~150 lines) - Chain examples
6. `runnable.go` (~300 lines) - Runnable implementation
7. `orchestrator.go` (~250 lines) - High-level orchestration
8. `callback.go` (~150 lines) - Callback system
9. `errors.go` (~50 lines) - Error definitions
10. `interrupt.go` (~200 lines) - Interrupt handling (NEW)
11. `interrupt_test.go` (~150 lines) - Interrupt tests (NEW)

**Success Criteria Met**:

- File count: 24 → 11 (-54%)
- Line count: 9,465 → 2,850 (-70%)
- All under 5,000 line target

### 4.3 Import Path Updates

**Before**:

```go
import "github.com/kart-io/k8s-agent/pkg/agent/core"

checkpoint := &core.Checkpoint{ ... }
```

**After**:

```go
import (
    "github.com/kart-io/k8s-agent/pkg/agent/core/checkpoint"
    "github.com/kart-io/k8s-agent/pkg/agent/interfaces"
)

cp := &checkpoint.Checkpoint{ ... }
var checkpointer interfaces.Checkpointer = checkpoint.NewRedisCheckpointer(...)
```

## 5. Testing Strategy

### 5.1 Test Coverage Goals

| Package | Current Coverage | Target Coverage | Gap | Priority |
|---------|-----------------|-----------------|-----|----------|
| core/ | 46.9% | >80% | +33.1% | High |
| core/state/ | 50% (estimated) | >80% | +30% | High |
| core/checkpoint/ | 60% (estimated) | >80% | +20% | High |
| core/execution/ | 40% (estimated) | >80% | +40% | High |
| core/middleware/ | 55% (estimated) | >75% | +20% | Medium |
| agents/ | 30% (estimated) | >70% | +40% | High |
| agents/executor/ | 0% (new) | >70% | +70% | High |
| tools/ | 40% (estimated) | >75% | +35% | Medium |
| interfaces/ | N/A (new) | 100% | N/A | Critical |

### 5.2 Test Types

#### Unit Tests

**Location**: Alongside implementation files (`*_test.go`)

**Coverage**:

- All public functions
- All exported types
- Error paths
- Edge cases

**Example Structure**:

```go
// core/state/state_test.go
func TestState_Get(t *testing.T) { ... }
func TestState_Set(t *testing.T) { ... }
func TestState_Merge(t *testing.T) { ... }
func TestState_Clone(t *testing.T) { ... }

// core/checkpoint/redis_test.go
func TestRedisCheckpointer_SaveCheckpoint(t *testing.T) { ... }
func TestRedisCheckpointer_LoadCheckpoint(t *testing.T) { ... }
func TestRedisCheckpointer_Concurrency(t *testing.T) { ... }
```

#### Integration Tests

**Location**: `{package}/test/integration/`

**Purpose**: Test interactions between packages

**Examples**:

```go
// core/test/integration/checkpoint_state_test.go
func TestCheckpointWithState(t *testing.T) {
    // Test checkpoint saving/loading with real state
}

// agents/executor/test/integration/executor_integration_test.go
func TestExecutorWithTools(t *testing.T) {
    // Test executor agent with real tools
}
```

#### Example Tests

**Location**: Within examples/ directories

**Purpose**: Ensure examples build and run

**Approach**:

```go
// examples/basic/01-simple-agent/simple_agent_test.go
func TestExample_SimpleAgent(t *testing.T) {
    // Run the example and verify output
}
```

### 5.3 Test Infrastructure

**Tools**:

- `testing` (standard library)
- `github.com/stretchr/testify` (assertions, mocks)
- `go-sqlmock` (database mocking)
- `golangci-lint` (static analysis)

**Coverage Commands**:

```bash
# Unit tests with coverage
make test-coverage

# Coverage report location
_output/coverage/coverage.html

# Per-package coverage
go test -coverprofile=coverage.out ./pkg/agent/core/...
go tool cover -html=coverage.out -o coverage.html
```

### 5.4 Test Milestones

**Phase 1: Emergency Fixes** (No new tests required)

- Documentation moves don't affect tests
- Filename changes update test files automatically

**Phase 2: Structural Refactoring**

- Milestone 1: Interface package tests (100% coverage)
- Milestone 2: Core sub-package tests (>80% each)
- Milestone 3: Agent/tool separation tests (>70%)

**Phase 3: Quality Improvements**

- Milestone 4: Fill gaps to reach overall >75%
- Milestone 5: Integration tests for critical paths
- Milestone 6: All examples have runnable tests

## 6. Backward Compatibility Strategy

### 6.1 Type Alias Mechanism

**Purpose**: Allow existing code to continue working while encouraging migration.

**Implementation Pattern**:

```go
// Old location (e.g., retrieval/vector_store.go)
package retrieval

import "github.com/kart-io/k8s-agent/pkg/agent/interfaces"

// VectorStore is deprecated. Use interfaces.VectorStore.
//
// This type alias provides backward compatibility for one major version.
//
// Migration guide: https://github.com/kart-io/k8s-agent/blob/master/docs/refactoring/migration-guide.md
//
// Deprecated: Use interfaces.VectorStore instead.
type VectorStore = interfaces.VectorStore
```

**Compile-Time Guidance**:

```go
// When users view godoc or IDE hints, they see:
//
// type VectorStore = interfaces.VectorStore
//     Deprecated: Use interfaces.VectorStore instead.
//
// This guides them to the new location without breaking builds.
```

### 6.2 Deprecation Timeline

**Version Strategy**: Semantic versioning (current: v0.x.x)

| Version | Phase | Actions | Breaking Changes |
|---------|-------|---------|------------------|
| v0.9.0 | Pre-refactor | Current state | None |
| v0.10.0 | Refactor Phase 1 | Docs moved, interfaces created, type aliases added | None (fully compatible) |
| v0.11.0 | Refactor Phase 2 | Core split, agents/tools separated | None (type aliases work) |
| v0.12.0 | Refactor Phase 3 | Tests added, examples reorganized | None (type aliases work) |
| v0.13.0 | Stabilization | Deprecation warnings, migration guide published | None (warnings only) |
| v0.14.0 | Final pre-1.0 | All deprecated paths still work | None (warnings only) |
| v1.0.0 | Major release | **Type aliases removed** | Yes (planned, documented) |

**Deprecation Period**: Minimum 4 minor versions (v0.10.0 → v1.0.0)

### 6.3 Migration Guide Structure

**File**: `docs/refactoring/migration-guide.md`

**Contents**:

```markdown
# Migration Guide: pkg/agent Refactoring

## Overview
This guide helps you migrate from the old package structure to the new refactored structure.

## Quick Reference Table

| Old Import | New Import | Version | Action |
|-----------|------------|---------|--------|
| `github.com/kart-io/k8s-agent/pkg/agent/core.Agent` | `github.com/kart-io/k8s-agent/pkg/agent/interfaces.Agent` | v0.10.0+ | Update imports |
| `github.com/kart-io/k8s-agent/pkg/agent/retrieval.VectorStore` | `github.com/kart-io/k8s-agent/pkg/agent/interfaces.VectorStore` | v0.10.0+ | Update imports |
| `github.com/kart-io/k8s-agent/pkg/agent/core.Checkpointer` | `github.com/kart-io/k8s-agent/pkg/agent/interfaces.Checkpointer` | v0.10.0+ | Update imports |
| `github.com/kart-io/k8s-agent/pkg/agent/core/runtime.go` | `github.com/kart-io/k8s-agent/pkg/agent/core/execution/runtime.go` | v0.11.0+ | Update imports |

## Migration Steps

### Step 1: Update Interface Imports (Required for v1.0.0)
... detailed steps ...

### Step 2: Update Sub-package Imports (Required for v1.0.0)
... detailed steps ...

### Step 3: Update Example References (Optional but recommended)
... detailed steps ...

## Compatibility Matrix
... version-by-version compatibility ...

## Automated Migration Tools
... scripts to help automate migration ...

## FAQ
... common questions and answers ...
```

### 6.4 Compatibility Verification

**Test Strategy**: Ensure old code still compiles

**Test File**: `pkg/agent/test/compatibility/v0_9_compatibility_test.go`

```go
// TestBackwardCompatibility_V0_9 verifies that code written against v0.9.0 still compiles
func TestBackwardCompatibility_V0_9(t *testing.T) {
    // Old way (should still work with type aliases)
    var vectorStore retrieval.VectorStore = retrieval.NewMockVectorStore()
    var agent core.Agent = core.NewBaseAgent("test")

    // Verify type compatibility
    var _ interfaces.VectorStore = vectorStore
    var _ interfaces.Agent = agent

    // Old import paths should still work
    runtime := core.NewRuntime()  // Type alias points to execution.Runtime
    _ = runtime
}
```

## 7. Rollback Procedures

### 7.1 Rollback Strategy

**Principle**: Each phase is atomic and independently revertible.

**Rollback Points**:

1. After Phase 1 (Documentation)
2. After Phase 2.1 (Interfaces)
3. After Phase 2.2 (Core split)
4. After Phase 2.3 (Agent/tool separation)
5. After Phase 3 (Testing/examples)

### 7.2 Phase-by-Phase Rollback

#### Phase 1 Rollback: Documentation Reorganization

**Trigger Conditions**:

- Documentation is in wrong locations
- README/ARCHITECTURE are broken
- User confusion

**Rollback Steps**:

```bash
# 1. Identify the commit before Phase 1
git log --oneline --grep="Phase 1: Documentation" -1

# 2. Revert the documentation commits (preserving other work)
git revert <commit-hash>

# 3. Verify rollback
ls pkg/agent/*.md  # Should show original 26 files
test -d pkg/agent/docs && echo "ERROR: docs/ still exists"

# 4. Commit rollback
git commit -m "Rollback: Phase 1 documentation reorganization"
```

**Validation**:

- [ ] All original 26 Markdown files are back
- [ ] `docs/` directory removed or empty
- [ ] README.md and ARCHITECTURE.md unchanged
- [ ] No broken links in documentation

**Time to Rollback**: < 5 minutes

#### Phase 2.1 Rollback: Interface Unification

**Trigger Conditions**:

- Type aliases causing compilation errors
- Interface incompatibilities
- Import cycle created

**Rollback Steps**:

```bash
# 1. Remove interfaces/ package
rm -rf pkg/agent/interfaces/

# 2. Revert type alias changes in packages
git checkout HEAD -- pkg/agent/retrieval/vector_store.go
git checkout HEAD -- pkg/agent/memory/manager.go
git checkout HEAD -- pkg/agent/core/agent.go

# 3. Revert import changes in all files
find pkg/agent -name "*.go" -exec \
  sed -i 's|github.com/kart-io/k8s-agent/pkg/agent/interfaces|github.com/kart-io/k8s-agent/pkg/agent/core|g' {} \;

# 4. Run tests
make test

# 5. Commit rollback
git add -A
git commit -m "Rollback: Phase 2.1 interface unification"
```

**Validation**:

- [ ] `interfaces/` package deleted
- [ ] All imports restored to original packages
- [ ] All tests pass
- [ ] `make build` succeeds
- [ ] No deprecation warnings

**Time to Rollback**: < 10 minutes

#### Phase 2.2 Rollback: Core Package Decomposition

**Trigger Conditions**:

- Import cycles introduced
- Tests failing after split
- Performance regression > 10%
- Sub-packages don't build

**Rollback Steps**:

```bash
# 1. Restore original core/ structure
git checkout HEAD -- pkg/agent/core/

# 2. Remove new sub-packages
rm -rf pkg/agent/core/state/
rm -rf pkg/agent/core/checkpoint/
rm -rf pkg/agent/core/execution/
rm -rf pkg/agent/core/middleware/

# 3. Restore original file names
cd pkg/agent/core/
git mv checkpointer_redis.go checkpointer_redis.go  # If renamed
git mv checkpointer_distributed.go checkpointer_distributed.go

# 4. Revert import changes
find pkg/agent -name "*.go" -exec \
  sed -i 's|github.com/kart-io/k8s-agent/pkg/agent/core/state|github.com/kart-io/k8s-agent/pkg/agent/core|g' {} \;
find pkg/agent -name "*.go" -exec \
  sed -i 's|github.com/kart-io/k8s-agent/pkg/agent/core/checkpoint|github.com/kart-io/k8s-agent/pkg/agent/core|g' {} \;
find pkg/agent -name "*.go" -exec \
  sed -i 's|github.com/kart-io/k8s-agent/pkg/agent/core/execution|github.com/kart-io/k8s-agent/pkg/agent/core|g' {} \;
find pkg/agent -name "*.go" -exec \
  sed -i 's|github.com/kart-io/k8s-agent/pkg/agent/core/middleware|github.com/kart-io/k8s-agent/pkg/agent/core|g' {} \;

# 5. Run full test suite
make test-all

# 6. Verify build performance
time make build  # Should be within 10% of baseline

# 7. Commit rollback
git add -A
git commit -m "Rollback: Phase 2.2 core package decomposition"
```

**Validation**:

- [ ] Core package has original 24 files
- [ ] No sub-packages exist
- [ ] All tests pass (100% of original tests)
- [ ] Build time within 10% of baseline
- [ ] No import cycles (`go list -f '{{.ImportPath}} {{.Imports}}' ./... | grep cycle`)

**Time to Rollback**: < 15 minutes

#### Phase 2.3 Rollback: Agent/Tool Separation

**Trigger Conditions**:

- Executor agent doesn't work in new location
- Tool runtime name collision issues
- Tests failing

**Rollback Steps**:

```bash
# 1. Remove new agents/executor/
rm -rf pkg/agent/agents/executor/

# 2. Restore executor_tool.go to tools/
git checkout HEAD -- pkg/agent/tools/executor_tool.go

# 3. Restore tools/runtime.go
git checkout HEAD -- pkg/agent/tools/runtime.go
# OR
git mv pkg/agent/tools/tool_runtime.go pkg/agent/tools/runtime.go

# 4. Revert import changes
find pkg/agent -name "*.go" -exec \
  sed -i 's|github.com/kart-io/k8s-agent/pkg/agent/agents/executor|github.com/kart-io/k8s-agent/pkg/agent/tools|g' {} \;

# 5. Run tests
make test

# 6. Commit rollback
git add -A
git commit -m "Rollback: Phase 2.3 agent/tool separation"
```

**Validation**:

- [ ] `executor_tool.go` back in `tools/`
- [ ] `runtime.go` (not `tool_runtime.go`) in `tools/`
- [ ] All executor tests pass
- [ ] No broken imports

**Time to Rollback**: < 5 minutes

#### Phase 3 Rollback: Examples Reorganization

**Trigger Conditions**:

- Examples don't build
- Examples don't run
- User confusion about new structure

**Rollback Steps**:

```bash
# 1. Remove new examples/ structure
rm -rf pkg/agent/examples/

# 2. Restore original example/ directory
git checkout HEAD -- pkg/agent/example/

# 3. Restore main.go names
cd pkg/agent/example/
for dir in */; do
    if [ -f "$dir/$(basename $dir)_demo.go" ]; then
        git mv "$dir/$(basename $dir)_demo.go" "$dir/main.go"
    fi
done

# 4. Verify examples build
for dir in example/*/; do
    echo "Building $dir..."
    (cd "$dir" && go build -o /dev/null .)
done

# 5. Commit rollback
git add -A
git commit -m "Rollback: Phase 3 examples reorganization"
```

**Validation**:

- [ ] All original example/ directories restored
- [ ] All examples build successfully
- [ ] All `main.go` files restored (not `*_demo.go`)
- [ ] Example READMEs removed (if created)

**Time to Rollback**: < 5 minutes

### 7.3 Emergency Rollback (Full Revert)

**Trigger Conditions**:

- Critical bug introduced
- Major compatibility issue
- Production system broken
- Unable to debug quickly

**Emergency Steps**:

```bash
# 1. Identify the last known good commit (before refactoring started)
git log --oneline --all -20

# 2. Create a rollback branch
git checkout -b emergency-rollback-$(date +%Y%m%d)

# 3. Hard reset to last good commit
git reset --hard <last-good-commit-hash>

# 4. Force push to rollback branch (for review)
git push origin emergency-rollback-$(date +%Y%m%d) --force

# 5. Test rollback
make clean
make build
make test

# 6. If tests pass, merge rollback
git checkout master
git merge emergency-rollback-$(date +%Y%m%d)
git push origin master
```

**Post-Rollback Actions**:

1. Document what went wrong
2. Create issue with root cause analysis
3. Plan fix or alternative approach
4. Communicate to team and users

**Time to Rollback**: < 10 minutes

### 7.4 State Verification Checks

**Purpose**: Verify system state after any rollback.

**Checklist Script**: `scripts/verify-state.sh`

```bash
#!/bin/bash
# Verify pkg/agent state after rollback

echo "=== State Verification Check ==="

# 1. Build check
echo "1. Build check..."
if make build >/dev/null 2>&1; then
    echo "   ✓ Build succeeds"
else
    echo "   ✗ Build fails"
    exit 1
fi

# 2. Test check
echo "2. Test check..."
if make test >/dev/null 2>&1; then
    echo "   ✓ Tests pass"
else
    echo "   ✗ Tests fail"
    exit 1
fi

# 3. Lint check
echo "3. Lint check..."
if make lint >/dev/null 2>&1; then
    echo "   ✓ Linting passes"
else
    echo "   ✗ Linting fails"
    exit 1
fi

# 4. Import cycle check
echo "4. Import cycle check..."
if go list -f '{{.ImportPath}} {{.Imports}}' ./pkg/agent/... 2>&1 | grep -q "import cycle"; then
    echo "   ✗ Import cycle detected"
    exit 1
else
    echo "   ✓ No import cycles"
fi

# 5. Example build check
echo "5. Example build check..."
EXAMPLE_FAILURES=0
for dir in pkg/agent/example/*/; do
    if ! (cd "$dir" && go build -o /dev/null . 2>/dev/null); then
        echo "   ✗ Example $(basename $dir) fails to build"
        EXAMPLE_FAILURES=$((EXAMPLE_FAILURES+1))
    fi
done

if [ $EXAMPLE_FAILURES -eq 0 ]; then
    echo "   ✓ All examples build"
else
    echo "   ✗ $EXAMPLE_FAILURES example(s) fail to build"
    exit 1
fi

echo "=== All checks passed ==="
```

**Usage**:

```bash
# After any rollback
./scripts/verify-state.sh

# On success, proceed with commit
# On failure, investigate and fix before committing
```

## 8. Implementation Phases

### Phase 1: Emergency Fixes (1-2 days)

**Goals**: Quick wins, zero risk

**Tasks**:

1. Documentation reorganization
2. Create `docs/` directory structure
3. Move Markdown files to appropriate locations
4. Update root README.md and ARCHITECTURE.md

**Success Criteria**:

- [ ] Root directory has exactly 2 Markdown files
- [ ] All documentation in `docs/` subdirectories
- [ ] No broken links
- [ ] All builds still pass

**Rollback**: Trivial (move files back)

### Phase 2: Structural Refactoring (1-2 weeks)

#### Phase 2.1: Interface Unification (2-3 days)

**Tasks**:

1. Create `interfaces/` package
2. Define canonical interfaces
3. Add type aliases in original locations
4. Update internal imports
5. Test compatibility

**Success Criteria**:

- [ ] `interfaces/` package created with 6 files
- [ ] Type aliases in place
- [ ] All tests pass
- [ ] No breaking changes

**Rollback**: Delete `interfaces/`, revert aliases (10 mins)

#### Phase 2.2: Core Package Decomposition (3-5 days)

**Tasks**:

1. Create `core/state/` sub-package
2. Create `core/checkpoint/` sub-package
3. Create `core/execution/` sub-package
4. Create `core/middleware/` sub-package
5. Move files to sub-packages
6. Update imports
7. Run tests

**Success Criteria**:

- [ ] Core root has ≤15 files, ≤5000 lines
- [ ] Each sub-package < 2500 lines
- [ ] All tests pass
- [ ] No import cycles

**Rollback**: Remove sub-packages, restore files (15 mins)

#### Phase 2.3: Agent/Tool Separation (1-2 days)

**Tasks**:

1. Create `agents/executor/` directory
2. Move `executor_tool.go` → `executor_agent.go`
3. Rename `tools/runtime.go` → `tools/tool_runtime.go`
4. Update imports
5. Test executor agent

**Success Criteria**:

- [ ] Executor in `agents/executor/`
- [ ] No filename collisions
- [ ] All tests pass

**Rollback**: Move files back, rename (5 mins)

#### Phase 2.4: Store/Config Renaming (1 day)

**Tasks**:

1. Rename `store/postgres/config.go` → `store/postgres/postgres_store.go`
2. Rename `store/redis/config.go` → `store/redis/redis_store.go`
3. Update imports

**Success Criteria**:

- [ ] No filename collisions
- [ ] All tests pass

**Rollback**: Rename back (2 mins)

### Phase 3: Quality Improvements (2-3 weeks)

#### Phase 3.1: Test Coverage Enhancement (1 week)

**Tasks**:

1. Add unit tests to reach coverage targets
2. Create integration tests
3. Generate coverage reports

**Success Criteria**:

- [ ] Core package >80% coverage
- [ ] Agents package >70% coverage
- [ ] Tools package >75% coverage

**Rollback**: Delete new test files (no impact on functionality)

#### Phase 3.2: Example Reorganization (3-5 days)

**Tasks**:

1. Create `examples/` directory structure
2. Move examples to basic/advanced/integration
3. Rename `main.go` files
4. Add README files
5. Test all examples

**Success Criteria**:

- [ ] All examples build
- [ ] Examples organized by complexity
- [ ] Each category has README

**Rollback**: Restore `example/` directory (5 mins)

#### Phase 3.3: Documentation Updates (2-3 days)

**Tasks**:

1. Update ARCHITECTURE.md with new structure
2. Create migration guide
3. Update README.md
4. Update example READMEs

**Success Criteria**:

- [ ] All documentation accurate
- [ ] Migration guide complete
- [ ] No broken links

**Rollback**: Revert documentation commits (2 mins)

## 9. Success Metrics

### Quantitative Metrics

| Metric | Before | After | Target | Status |
|--------|--------|-------|--------|--------|
| Root Markdown files | 26 | 0 | ≤2 | TBD |
| Core package files | 24 | 11 | ≤15 | TBD |
| Core package lines | 9,465 | 2,850 | ≤5,000 | TBD |
| Filename collisions | 9+ | 0 | 0 | TBD |
| Core test coverage | 46.9% | TBD | >80% | TBD |
| Agent test coverage | ~30% | TBD | >70% | TBD |
| Tool test coverage | ~40% | TBD | >75% | TBD |
| Overall test coverage | ~60% | TBD | >75% | TBD |
| Failing examples | 8+ | 0 | 0 | TBD |
| Build time increase | 0% | TBD | <10% | TBD |

### Qualitative Goals

- [ ] Clear package boundaries
- [ ] Single source of truth for interfaces
- [ ] Logical organization
- [ ] Comprehensive examples
- [ ] Maintainable structure
- [ ] Backward compatible
- [ ] Well documented

## 10. Risk Mitigation

### High-Risk Mitigations

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|------------|
| Breaking existing code | Medium | High | Type aliases, compatibility tests, rollback procedures |
| Test failures | Medium | Medium | Incremental changes, run tests after each commit |
| Interface incompatibilities | Low | High | Unified interface package first, type checking |
| Import cycles | Low | High | Dependency analysis before moves, gradual refactoring |

### Medium-Risk Mitigations

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|------------|
| Build time increase | Low | Medium | Benchmark builds, monitor CI times |
| Documentation drift | Medium | Low | Update docs with each change, automated link checks |
| Example breakage | Medium | Medium | Test examples in each phase, CI integration |

## 11. Appendix

### A. Dependency Graph

```text
interfaces/
    ↓
core/ ← core/state/ ← core/checkpoint/
    ↓                      ↓
core/execution/        (Redis, etc.)
    ↓
core/middleware/
    ↓
agents/ ← tools/
    ↓
retrieval/ ← memory/ ← store/
```

### B. File Size Analysis

**Current Core Package**:

```text
core/agent.go                    600 lines
core/agent_test.go               400 lines
core/chain.go                    400 lines
core/chain_test.go               200 lines
core/chain_example_test.go       150 lines
core/checkpointer.go             500 lines
core/checkpointer_test.go        300 lines
core/checkpointer_redis.go       400 lines
core/checkpointer_redis_test.go  300 lines
core/checkpointer_distributed.go 600 lines
core/state.go                    500 lines
core/state_test.go               200 lines
core/runtime.go                  600 lines
core/runtime_test.go             300 lines
core/middleware.go               400 lines
core/middleware_test.go          200 lines
core/middleware_advanced.go      500 lines
core/orchestrator.go             250 lines
core/runnable.go                 300 lines
core/streaming.go                300 lines
core/callback.go                 150 lines
core/errors.go                    50 lines
core/interrupt.go                200 lines
core/interrupt_test.go           150 lines
-------------------------------------------
Total: 24 files, 9,465 lines
```

**Target Core Package** (after refactoring):

```text
core/
  agent.go                    600 lines
  agent_test.go               400 lines
  chain.go                    400 lines
  chain_test.go               200 lines
  chain_example_test.go       150 lines
  runnable.go                 300 lines
  orchestrator.go             250 lines
  callback.go                 150 lines
  errors.go                    50 lines
  interrupt.go                200 lines
  interrupt_test.go           150 lines
  ---------------------------------------
  Subtotal: 11 files, 2,850 lines

core/state/
  state.go                    550 lines
  state_test.go               200 lines
  manager.go                  200 lines
  serializer.go               150 lines
  ---------------------------------------
  Subtotal: 4 files, 1,100 lines

core/checkpoint/
  checkpointer.go             200 lines
  checkpointer_test.go        300 lines
  memory.go                   300 lines
  redis.go                    400 lines
  redis_test.go               300 lines
  distributed.go              600 lines
  saver.go                    200 lines
  ---------------------------------------
  Subtotal: 7 files, 2,300 lines

core/execution/
  runtime.go                  600 lines
  runtime_test.go             300 lines
  executor.go                 300 lines
  context.go                  200 lines
  streaming.go                300 lines
  ---------------------------------------
  Subtotal: 5 files, 1,700 lines

core/middleware/
  middleware.go               400 lines
  middleware_test.go          200 lines
  advanced.go                 500 lines
  chain.go                    200 lines
  builtin.go                  200 lines
  ---------------------------------------
  Subtotal: 5 files, 1,500 lines

===========================================
Grand Total: 32 files, 9,450 lines

Per-package max: 2,850 lines ✓ (under 5,000)
Files per package: 4-11 ✓ (under 15)
```

### C. Example Rename Mapping

| Old Path | New Path |
|----------|----------|
| `example/main.go` | `examples/basic/01-simple-agent/simple_agent.go` |
| `example/tools/main.go` | `examples/basic/03-tools/tools_demo.go` |
| `example/streaming/main.go` | `examples/advanced/streaming/streaming_demo.go` |
| `example/multi_mode_streaming/main.go` | `examples/advanced/multi-mode-streaming/multi_mode_demo.go` |
| `example/observability/main.go` | `examples/advanced/observability/observability_demo.go` |
| `example/react_example/main.go` | `examples/advanced/react/react_demo.go` |
| `example/parallel_execution/main.go` | `examples/advanced/parallel-execution/parallel_demo.go` |
| `example/tool_runtime/main.go` | `examples/advanced/tool-runtime/runtime_demo.go` |
| `example/tool_selector/main.go` | `examples/advanced/tool-selector/selector_demo.go` |
| `example/langchain_inspired/main.go` | `examples/integration/langchain-inspired/langchain_demo.go` |
| `example/langchain_complete/main.go` | `examples/integration/langchain-complete/complete_demo.go` |
| `example/langchain_phase1/main.go` | `examples/integration/langchain-phase1/phase1_demo.go` |
| `example/langchain_phase2/main.go` | `examples/integration/langchain-phase2/phase2_demo.go` |
| `example/multiagent/main.go` | `examples/integration/multiagent/multiagent_demo.go` |
| `example/human_in_the_loop/main.go` | `examples/integration/human-in-loop/hitl_demo.go` |
| `example/preconfig_agents/main.go` | `examples/integration/preconfig-agents/preconfig_demo.go` |

---

**Document Status**: Ready for Review

**Next Steps**: Upon approval, proceed to Task List (tasks.md) for implementation breakdown.
