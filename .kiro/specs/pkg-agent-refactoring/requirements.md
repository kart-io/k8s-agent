# Requirements Document: pkg/agent Directory Refactoring

## Introduction

The pkg/agent directory has evolved organically and now suffers from organizational debt that impairs maintainability, discoverability, and extensibility. This specification defines a comprehensive refactoring to establish a clean, maintainable architecture while preserving all existing functionality.

## Current State Analysis

### Critical Issues Identified

1. **Documentation Chaos**: 34 Markdown files in root directory (26 should be archived)
2. **Core Package Bloat**: 24 Go files with 9,465 lines of code in single package
3. **File Name Collisions**: 9 categories of duplicate filenames across packages
4. **Misplaced Components**: Agent implementations in tools package
5. **Interface Fragmentation**: VectorStore defined in 2+ incompatible locations
6. **Test Coverage Gaps**: Overall ~46.9% coverage, multiple packages at 0%
7. **Build Failures**: 8 example packages failing to build

### Quantitative Metrics

- Total Go files: ~280 files
- Total lines of code: 68,682 lines
- Markdown documents: 34 files
- Test files: 39 files (excluding examples)
- Core package files: 24 files (9,465 lines)
- Example packages: 15 directories (many failing)

## Requirements

### Requirement 1: Documentation Organization

**User Story**: As a developer, I want documentation organized in logical categories, so that I can quickly find relevant information without digging through 34 files.

#### Acceptance Criteria

1. WHEN the refactoring is complete THEN the root directory SHALL contain no more than 2 Markdown files (README.md and ARCHITECTURE.md)
2. IF a document describes completed implementation THEN it SHALL be moved to `docs/archive/`
3. IF a document describes analysis or planning THEN it SHALL be moved to `docs/analysis/`
4. IF a document describes refactoring process THEN it SHALL be moved to `docs/refactoring/`
5. WHEN documentation is moved THEN a `.gitkeep` or README SHALL exist in each new directory

### Requirement 2: Core Package Decomposition

**User Story**: As a maintainer, I want the core package split into focused sub-packages, so that each component has clear responsibilities and boundaries.

#### Acceptance Criteria

1. WHEN core package is refactored THEN it SHALL be split into minimum 3 sub-packages: state, checkpoint, and execution
2. IF a file handles state management THEN it SHALL reside in `core/state/`
3. IF a file handles checkpointing THEN it SHALL reside in `core/checkpoint/`
4. IF a file handles agent execution THEN it SHALL reside in `core/execution/`
5. WHEN splitting is complete THEN core root SHALL contain only orchestration and core abstractions
6. WHEN refactoring is complete THEN no single package SHALL exceed 15 files or 5,000 lines
7. WHEN imports are updated THEN backward compatibility aliases SHALL be provided for 1 release cycle

### Requirement 3: File Name Deduplication

**User Story**: As a developer, I want unique, descriptive filenames, so that I can quickly identify files by name without checking paths.

#### Acceptance Criteria

1. WHEN file renaming is complete THEN zero filename collisions SHALL exist across the entire pkg/agent tree
2. IF multiple `main.go` exist THEN each SHALL be renamed to `{purpose}_main.go` or moved to cmd-style structure
3. IF multiple `runtime.go` exist THEN each SHALL be prefixed with package context (e.g., `tool_runtime.go`, `agent_runtime.go`)
4. IF multiple `config.go` exist THEN each SHALL be prefixed with package context
5. WHEN renaming occurs THEN commit messages SHALL document old→new mappings

### Requirement 4: Component Boundary Enforcement

**User Story**: As an architect, I want clear separation between tools and agents, so that each package has a single, well-defined responsibility.

#### Acceptance Criteria

1. WHEN refactoring is complete THEN `tools/` package SHALL contain only tool definitions and execution
2. IF a file implements Agent interface THEN it SHALL NOT reside in tools package
3. IF executor_tool.go contains Agent logic THEN it SHALL be moved to `agents/executor/`
4. WHEN components are moved THEN all tests SHALL be updated and pass
5. WHEN separation is complete THEN package documentation SHALL clearly define boundaries

### Requirement 5: Interface Unification

**User Story**: As a developer, I want a single source of truth for interfaces, so that I don't have conflicting definitions causing compilation errors.

#### Acceptance Criteria

1. WHEN refactoring is complete THEN a new `interfaces/` package SHALL exist under pkg/agent
2. IF an interface is used by 2+ packages THEN it SHALL be defined in `interfaces/` package
3. WHEN VectorStore interface is unified THEN only 1 canonical definition SHALL exist
4. IF legacy interface references exist THEN they SHALL use type aliases pointing to canonical definition
5. WHEN interfaces are centralized THEN documentation SHALL list all public interfaces with usage examples

### Requirement 6: Example Package Reorganization

**User Story**: As a new user, I want examples organized by complexity and topic, so that I can learn incrementally.

#### Acceptance Criteria

1. WHEN reorganization is complete THEN examples SHALL be categorized into `basic/`, `advanced/`, and `integration/`
2. IF an example demonstrates single feature THEN it SHALL be in `basic/`
3. IF an example combines multiple features THEN it SHALL be in `advanced/`
4. IF an example shows full system integration THEN it SHALL be in `integration/`
5. WHEN examples are moved THEN each SHALL have a README explaining purpose and prerequisites
6. WHEN refactoring is complete THEN all example packages SHALL build successfully

### Requirement 7: Test Coverage Enhancement

**User Story**: As a maintainer, I want comprehensive test coverage for critical packages, so that refactoring doesn't introduce regressions.

#### Acceptance Criteria

1. WHEN testing is complete THEN core package coverage SHALL exceed 80%
2. WHEN testing is complete THEN agents package coverage SHALL exceed 70%
3. WHEN testing is complete THEN tools package coverage SHALL exceed 75%
4. IF a package has 0% coverage AND it contains business logic THEN tests SHALL be added
5. WHEN tests are added THEN they SHALL include unit, integration, and example tests
6. WHEN coverage improves THEN coverage reports SHALL be generated in `_output/coverage/`

### Requirement 8: Backward Compatibility

**User Story**: As a user of this library, I want existing code to continue working, so that I can upgrade without breaking my applications.

#### Acceptance Criteria

1. WHEN packages are moved THEN type aliases SHALL be created at old locations
2. IF an exported symbol changes location THEN the old path SHALL remain importable for 1 major version
3. WHEN breaking changes are necessary THEN they SHALL be documented in MIGRATION.md
4. IF deprecated paths are used THEN compile-time warnings SHALL guide users to new paths
5. WHEN refactoring is complete THEN all existing examples in consuming code SHALL still compile

### Requirement 9: Build System Integration

**User Story**: As a developer, I want the refactored structure to work seamlessly with existing build tools, so that CI/CD pipelines don't break.

#### Acceptance Criteria

1. WHEN refactoring is complete THEN `make test` SHALL pass for all packages
2. WHEN refactoring is complete THEN `make lint` SHALL pass with 0 warnings
3. WHEN refactoring is complete THEN `make build` SHALL produce all binaries
4. IF package paths change THEN go.mod SHALL be updated correctly
5. WHEN builds run THEN build time SHALL NOT increase by more than 10%

### Requirement 10: Incremental Migration

**User Story**: As a project manager, I want refactoring done in safe, reviewable increments, so that we minimize risk of catastrophic breakage.

#### Acceptance Criteria

1. WHEN refactoring starts THEN it SHALL be divided into 3 phases: Emergency, Structural, Quality
2. IF a phase completes THEN all tests SHALL pass before starting next phase
3. WHEN each change is committed THEN it SHALL be atomic and independently verifiable
4. IF build breaks THEN the change SHALL be reverted immediately
5. WHEN each phase completes THEN a summary document SHALL be created

## Success Metrics

### Quantitative Goals

- Root directory Markdown files: 34 → 2 (94% reduction)
- Core package file count: 24 → ≤15 (38% reduction)
- Core package line count: 9,465 → ≤5,000 (47% reduction)
- Filename collisions: 9 → 0 (100% elimination)
- Test coverage (core): 46.9% → >80% (71% increase)
- Test coverage (overall): ~60% → >75% (25% increase)
- Failing example packages: 8 → 0 (100% fix rate)

### Qualitative Goals

- Clear package boundaries with documented responsibilities
- Single source of truth for all shared interfaces
- Logical organization that aids discoverability
- Comprehensive examples organized by complexity
- Maintainable structure that scales with growth

## Constraints

### Technical Constraints

1. Must maintain Go 1.25.0 compatibility
2. Must preserve all existing public APIs
3. Must not break existing consuming code
4. Must work with current build system (Makefile)
5. Must pass all existing tests after refactoring

### Process Constraints

1. Changes must be reviewable (max 500 lines per commit)
2. Each commit must leave codebase in buildable state
3. Must use git for all changes (no manual file moves)
4. Must run tests after each atomic change
5. Must document all breaking changes

### Time Constraints

- Phase 1 (Emergency): 1-2 days
- Phase 2 (Structural): 1-2 weeks
- Phase 3 (Quality): 2-3 weeks
- Total: 3-4 weeks maximum

## Out of Scope

The following are explicitly NOT part of this refactoring:

1. Performance optimization (unless regression occurs)
2. Feature additions or enhancements
3. Algorithm changes or improvements
4. External API changes
5. Dependency upgrades (unless required for refactoring)
6. UI/UX changes in examples
7. Internationalization
8. Security audits (beyond maintaining current security)

## Dependencies

### Internal Dependencies

- All pkg/agent packages (mutual dependencies during refactor)
- Build system (Makefile targets)
- Test infrastructure (test helpers, mocks)
- CI/CD pipeline (GitHub Actions or equivalent)

### External Dependencies

- Go toolchain 1.25.0
- Testing frameworks (testify, go-sqlmock)
- Linting tools (golangci-lint with 58 linters)
- Code formatting tools (gofumpt, gci)

## Risk Assessment

### High-Risk Items

1. **Breaking existing code**: Mitigated by type aliases and backward compatibility layer
2. **Test failures during migration**: Mitigated by incremental changes with test runs
3. **Interface incompatibilities**: Mitigated by unified interface package first
4. **Import cycle creation**: Mitigated by careful dependency analysis before moves

### Medium-Risk Items

1. **Build time increase**: Monitored with benchmarks
2. **Documentation drift**: Mitigated by updating docs with each change
3. **Example breakage**: Mitigated by testing examples in each phase

### Low-Risk Items

1. **File rename conflicts**: Easily reversible with git
2. **Directory structure changes**: Can be adjusted iteratively
3. **Markdown organization**: No code impact

## Approval Checkpoints

This requirements document requires approval before proceeding to design phase.

**Review Questions**:

1. Are the 10 requirements comprehensive and clear?
2. Do the success metrics align with project goals?
3. Are the constraints realistic and achievable?
4. Is anything missing from the scope?
5. Are the risk mitigations sufficient?

Upon approval, we will proceed to the design document which will detail:

- Package structure diagrams
- File migration mappings
- Interface unification strategy
- Testing approach
- Rollback procedures
