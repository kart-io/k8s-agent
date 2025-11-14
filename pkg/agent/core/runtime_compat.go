package core

import (
	"github.com/kart-io/k8s-agent/pkg/agent/core/checkpoint"
	"github.com/kart-io/k8s-agent/pkg/agent/core/execution"
	"github.com/kart-io/k8s-agent/pkg/agent/core/state"
	"github.com/kart-io/k8s-agent/pkg/agent/store"
)

// Runtime is deprecated. Use execution.Runtime instead.
//
// This type alias provides backward compatibility for one major version.
// Migration: import "github.com/kart-io/k8s-agent/pkg/agent/core/execution"
//
// Deprecated: Use execution.Runtime
type Runtime[C any, S state.State] = execution.Runtime[C, S]

// NewRuntime is deprecated. Use execution.NewRuntime instead.
//
// Deprecated: Use execution.NewRuntime
func NewRuntime[C any, S state.State](
	ctx C,
	st S,
	str store.Store,
	cp checkpoint.Checkpointer,
	sessionID string,
) *execution.Runtime[C, S] {
	return execution.NewRuntime[C, S](ctx, st, str, cp, sessionID)
}

// ToolFunc is deprecated. Use execution.ToolFunc instead.
//
// Deprecated: Use execution.ToolFunc
type ToolFunc[I, O, C any, S state.State] = execution.ToolFunc[I, O, C, S]

// ToolWithRuntime is deprecated. Use execution.ToolWithRuntime instead.
//
// Deprecated: Use execution.ToolWithRuntime
type ToolWithRuntime[I, O, C any, S state.State] = execution.ToolWithRuntime[I, O, C, S]

// NewToolWithRuntime is deprecated. Use execution.NewToolWithRuntime instead.
//
// Deprecated: Use execution.NewToolWithRuntime
func NewToolWithRuntime[I, O, C any, S state.State](
	name string,
	description string,
	fn execution.ToolFunc[I, O, C, S],
	runtime *execution.Runtime[C, S],
) *execution.ToolWithRuntime[I, O, C, S] {
	return execution.NewToolWithRuntime[I, O, C, S](name, description, fn, runtime)
}

// RuntimeConfig is deprecated. Use execution.RuntimeConfig instead.
//
// Deprecated: Use execution.RuntimeConfig
type RuntimeConfig = execution.RuntimeConfig

// DefaultRuntimeConfig is deprecated. Use execution.DefaultRuntimeConfig instead.
//
// Deprecated: Use execution.DefaultRuntimeConfig
var DefaultRuntimeConfig = execution.DefaultRuntimeConfig

// RuntimeMetrics is deprecated. Use execution.RuntimeMetrics instead.
//
// Deprecated: Use execution.RuntimeMetrics
type RuntimeMetrics = execution.RuntimeMetrics

// NewRuntimeMetrics is deprecated. Use execution.NewRuntimeMetrics instead.
//
// Deprecated: Use execution.NewRuntimeMetrics
var NewRuntimeMetrics = execution.NewRuntimeMetrics

// RuntimeManager is deprecated. Use execution.RuntimeManager instead.
//
// Deprecated: Use execution.RuntimeManager
type RuntimeManager[C any, S state.State] = execution.RuntimeManager[C, S]

// NewRuntimeManager is deprecated. Use execution.NewRuntimeManager instead.
//
// Deprecated: Use execution.NewRuntimeManager
func NewRuntimeManager[C any, S state.State](config *execution.RuntimeConfig) *execution.RuntimeManager[C, S] {
	return execution.NewRuntimeManager[C, S](config)
}
