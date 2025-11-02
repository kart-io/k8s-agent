# PKG Directory Analysis & Migration Documentation

This directory contains comprehensive analysis and migration plan documents for the `pkg/` directory refactoring.

## Documents Overview

### 1. [PKG_ANALYSIS_SUMMARY.txt](./PKG_ANALYSIS_SUMMARY.txt) - START HERE
**Quick executive overview** (3-5 minute read)

Contains:
- Current state of pkg/ directory
- List of mislocated packages
- Migration recommendations
- Timeline and effort estimates
- Key recommendations (DO's and DON'Ts)

**Who should read**: Team leads, project managers, developers planning the migration

---

### 2. [PKG_ANALYSIS_REPORT.md](./PKG_ANALYSIS_REPORT.md) - DETAILED ANALYSIS
**Comprehensive technical analysis** (15-20 minute read)

Contains:
- Package-by-package inventory
- Classification (move/keep/partial)
- Specific file recommendations
- Dependency impact assessment
- Conflicts and issues found
- Migration assessment matrix
- Summary tables

**Who should read**: Developers executing the migration, architects reviewing decisions

---

### 3. [PKG_MIGRATION_PLAN.md](./PKG_MIGRATION_PLAN.md) - EXECUTION GUIDE
**Step-by-step executable instructions** (reference document)

Contains:
- Phase-by-phase migration steps
- Specific bash commands for each move
- File counts and effort estimates
- Import path update procedures
- Validation checklist
- Rollback plan
- Success criteria

**Who should read**: Developers executing the migration, QA testing the changes

---

## Quick Reference Table

| Package | Current | Target | Move? | Priority | Status | Effort |
|---------|---------|--------|-------|----------|--------|--------|
| bootstrap | common/ | pkg/ | YES | HIGH | Ready | 1-2h |
| contextx | common/ | pkg/ | YES | HIGH | Ready | 1-2h |
| idempotent | common/ | pkg/ | YES | HIGH | Ready | 1-2h |
| metrics | common/ | pkg/ | PARTIAL | MEDIUM | Needs review | 2-3h |
| app | common/ | pkg/ | SPLIT | MEDIUM | Needs refactor | 4-6h |
| types | pkg/ | pkg/ | NO | N/A | Current | - |
| api | pkg/ | pkg/ | NO | N/A | Current | - |

---

## Current State

### In pkg/ (Correctly Located ✅)
- `pkg/types/` - Business domain models (295 lines)
- `pkg/api/` - Protocol Buffer definitions and generated code

### Should Move to pkg/ (Currently in common/ ⚠️)
- `bootstrap` - App initialization logic (3 files, ~15KB)
- `contextx` - K8s-agent specific context (4 files, ~25KB)
- `idempotent` - Business idempotency (4 files, ~17KB)

### Needs Analysis (Currently in common/ ⚠️)
- `metrics` - Mix of generic and project-specific
- `app` - Hybrid generic + project-specific

### Should Stay in common/ (Correctly Located ✅)
- cache, client, config, db, errors, logger, middleware, mq, pagination, response, server, validator, initializers, serializers, utils

---

## Decision Framework

### Belongs in pkg/ (Project-Specific Business Logic)
- Contains business logic specific to Aetherius
- Depends on domain models (Agent, Workflow, etc.)
- Not useful in other projects
- Tightly coupled with project requirements

### Belongs in common/ (Generic Utilities)
- Can be used by ANY Go project
- Contains ZERO business logic
- Pure technical implementation
- Could be published as independent library

---

## Timeline

**Phase 1 (Week 1)**: 3-5 hours
- Move bootstrap, contextx, idempotent
- Risk: Low

**Phase 2 (Week 2)**: 8-12 hours
- Analyze and refactor metrics, app
- Risk: Medium

**Phase 3 (Week 3)**: 2-3 hours
- Create planned pkg/ subdirectories
- Document structure

**Total**: 11-16 hours over 2-3 weeks

---

## How to Use This Documentation

### For Decision Makers
1. Read PKG_ANALYSIS_SUMMARY.txt (5 min)
2. Review the Quick Reference Table above
3. Decide on Phase 1 execution timeline

### For Architects
1. Read PKG_ANALYSIS_SUMMARY.txt (5 min)
2. Review PKG_ANALYSIS_REPORT.md (15 min)
3. Validate decision framework
4. Approve migration approach

### For Developers
1. Read PKG_ANALYSIS_SUMMARY.txt (5 min)
2. Review PKG_MIGRATION_PLAN.md (reference)
3. Follow step-by-step instructions
4. Use validation checklist

### For QA/Testing
1. Review PKG_ANALYSIS_SUMMARY.txt (5 min)
2. Check validation checklist in PKG_MIGRATION_PLAN.md
3. Run test suite
4. Verify service deployment

---

## Key Findings

### Finding #1: Import Path Inconsistency
CLAUDE.md architecture documentation specifies that bootstrap, contextx, idempotent, and metrics should be in pkg/, but they're currently in common/. This analysis confirms the correct location should be pkg/.

### Finding #2: Straightforward Phase 1 Migration
Three packages (bootstrap, contextx, idempotent) are ready to move immediately with minimal risk. These contain pure project-specific logic and have clear dependencies.

### Finding #3: Hybrid Package Refactoring Needed
The common/app package contains both generic patterns (RunWithRunner/RunWithOptions) and project-specific logic. It should be split, with generic parts staying in common/ and project-specific parts moving to pkg/.

### Finding #4: Metrics Clarification Needed
The common/metrics package may contain both generic Prometheus metrics and Aetherius-specific metrics. It needs analysis to determine if it should be moved, split, or kept as-is.

---

## Success Criteria

After completing this migration:

1. All three Phase 1 packages (bootstrap, contextx, idempotent) moved to pkg/
2. All import paths updated across codebase
3. All tests passing (unit, integration, e2e)
4. Services deploy and run successfully
5. Code structure matches CLAUDE.md architecture
6. Team has clear understanding of why packages are in pkg/ vs common/

---

## Next Steps

1. Review all three documents
2. Schedule stakeholder meeting to discuss recommendations
3. Approve Phase 1 execution plan
4. Assign developers to execute migration
5. Schedule code review for import path updates
6. Run comprehensive test suite
7. Document completion status in CLAUDE.md

---

## Questions & Clarifications

### Q: Why move packages from common/ to pkg/?
A: CLAUDE.md specifies that pkg/ contains Aetherius-specific business logic, while common/ contains reusable utilities. Moving these packages aligns the code structure with the documented architecture.

### Q: Will this break anything?
A: No. This is a pure refactoring. Import paths change, but no API surfaces, functionality, or behavior changes. All tests will verify this.

### Q: What about the other packages in common/?
A: They should stay in common/. They are generic utilities (cache, middleware, database wrappers, etc.) that could be used in any Go project.

### Q: What about pkg/types and pkg/api?
A: These are correctly located and require no changes.

### Q: How long will this take?
A: Phase 1 is 3-5 hours. Phase 2 (optional) is 8-12 hours. Phase 3 (future) is 2-3 hours.

---

## References

- Architecture documentation: `CLAUDE.md`
- Current code organization: `docs/CODE_REORGANIZATION.md`
- Original analysis request: Task to identify what should move to common/

---

**Document Version**: 1.0
**Created**: 2025-11-02
**Analysis Thoroughness**: MEDIUM
**Status**: Complete and ready for stakeholder review
