# Configuration Files Naming Standardization

## Summary

This document describes the standardization of all configuration files in the k8s-agent project to use the `*_options.go` naming convention. This standardization was completed on 2025-11-10.

## Motivation

The project had inconsistent naming conventions for configuration files:
- Some used `*_options.go` (standard pattern)
- Some used `config.go` (generic pattern)
- This inconsistency made it harder to understand the codebase structure

## Changes Made

### 1. File Renames

#### common/server/http/config.go → common/server/http/gin_server_options.go

**Rationale**: The file contains HTTP server configuration for Gin, so `gin_server_options.go` is more descriptive and follows the `*_options.go` pattern.

**Changes**:
- Renamed file from `config.go` to `gin_server_options.go`
- Renamed struct from `GinServerConfig` to `GinServerOptions`
- Renamed function from `NewGinServerConfig()` to `NewGinServerOptions()`
- Updated all method receivers from `c` to `o` (for "options")
- Updated all references across the codebase

**Files Updated**:
- `common/server/http/gin_server_options.go` (renamed and updated)
- `common/server/http/gin.go` (updated function signatures)
- `pkg/initializers/http_server.go` (updated function calls)
- `internal/reasoning/api/server.go` (updated function calls)
- `internal/cluster/api/server.go` (updated function calls)
- `internal/monitor/api/server.go` (updated function calls)
- `docs/refactoring/INITIALIZER_UNIFICATION_SUMMARY.md` (documentation)
- `docs/refactoring/INITIALIZER_AUDIT_REPORT.md` (documentation)

#### internal/reasoning/config/config.go → internal/reasoning/config/reasoning_options.go

**Rationale**: The file contains reasoning service specific options, so `reasoning_options.go` is more descriptive and follows the `*_options.go` pattern.

**Changes**:
- Renamed file from `config.go` to `reasoning_options.go`
- Renamed struct from `Config` to `ReasoningOptions`
- Added backward compatibility alias: `type Config = ReasoningOptions`
- Renamed function from `NewConfigFromOptions()` to `NewReasoningOptionsFromStandardOptions()`
- Kept old function name for backward compatibility
- Updated method receivers from `c` to `o`

**Note**: All existing code continues to work due to the backward compatibility alias. No changes to consuming code were needed.

### 2. Already Compliant Files

The following directories already followed the `*_options.go` naming convention:

#### common/options/ (100% compliant)
- `authentication_options.go`
- `cors_options.go`
- `grpc_options.go`
- `health_options.go`
- `http_client_options.go`
- `http_server_options.go`
- `jwt_options.go`
- `logging_options.go`
- `memory_options.go`
- `metrics_options.go`
- `mysql_options.go`
- `nats_options.go`
- `options.go`
- `performance_options.go`
- `prometheus_options.go`
- `rate_limit_options.go`
- `redis_options.go`
- `server_options.go`
- `tls_options.go`
- `tracer_options.go`
- `workflow_options.go`

#### pkg/options/ (100% compliant)
- `agent_options.go`
- `ai_options.go`
- `alert_options.go`
- `analysis_options.go`
- `email_options.go`
- `feature_gate_options.go`
- `learning_options.go`
- `llm_options.go`
- `prediction_options.go`

#### cmd/gateway/app/options/ (100% compliant)
- `options.go` (contains `ServerOptions` struct)

#### common/server/grpc/ (100% compliant)
- `options.go` (contains gRPC server options)

#### common/server/http/ (100% compliant after changes)
- `options.go` (contains HTTP server options)
- `gin_server_options.go` (renamed from config.go)

## Naming Convention Rules

### For Options Files

1. **File naming**: Use `*_options.go` pattern
   - Example: `server_options.go`, `database_options.go`, `llm_options.go`

2. **Struct naming**: Use `*Options` suffix
   - Example: `ServerOptions`, `DatabaseOptions`, `LLMOptions`

3. **Constructor naming**: Use `New*Options()` pattern
   - Example: `NewServerOptions()`, `NewDatabaseOptions()`

4. **Method receivers**: Use `o` for "options"
   - Example: `func (o *ServerOptions) Validate() error`

### For Config Files (Legacy)

If backward compatibility is required:
- Keep the `Config` type as an alias: `type Config = Options`
- Keep the old constructor as a wrapper
- Update comments to indicate the preferred naming

## Impact Analysis

### Build Verification

All services build successfully after the changes:
- agent-manager ✓
- orchestrator ✓
- reasoning ✓
- auth ✓
- gateway ✓
- monitor ✓
- cluster ✓
- collect-agent ✓

### Test Results

All existing tests pass. No test changes were required.

## Benefits

1. **Consistency**: All configuration files now follow the same naming pattern
2. **Discoverability**: Easier to find configuration files using glob patterns
3. **Clarity**: The `*_options.go` suffix clearly indicates the file's purpose
4. **Maintainability**: Consistent patterns reduce cognitive load for developers
5. **Documentation**: File names are self-documenting

## Migration Guide

If you need to add new configuration files:

1. **Name the file**: Use `*_options.go` pattern
   ```
   my_feature_options.go
   ```

2. **Name the struct**: Use `*Options` suffix
   ```go
   type MyFeatureOptions struct {
       // fields...
   }
   ```

3. **Create constructor**: Use `New*Options()` pattern
   ```go
   func NewMyFeatureOptions() *MyFeatureOptions {
       return &MyFeatureOptions{
           // defaults...
       }
   }
   ```

4. **Use consistent receivers**: Use `o` for options methods
   ```go
   func (o *MyFeatureOptions) Validate() error {
       // validation...
   }
   ```

## References

- [common/options/README.md](../../common/options/README.md) - Common options documentation
- [pkg/options/](../../pkg/options/) - Project-specific options
- [CLAUDE.md](../../CLAUDE.md) - Project coding standards

## Conclusion

This standardization effort brings consistency to the codebase and makes it easier for developers to understand the project structure. All configuration files now follow the `*_options.go` naming convention, with backward compatibility maintained where necessary.

The changes are non-breaking, and all services continue to work as expected. This is a foundation for future development and helps maintain code quality across the project.
