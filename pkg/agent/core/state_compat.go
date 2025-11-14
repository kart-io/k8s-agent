package core

import "github.com/kart-io/k8s-agent/pkg/agent/core/state"

// State is deprecated. Use core/state.State instead.
//
// This type alias provides backward compatibility during the refactoring period.
// All new code should import and use github.com/kart-io/k8s-agent/pkg/agent/core/state directly.
//
// Deprecated: Use github.com/kart-io/k8s-agent/pkg/agent/core/state.State
type State = state.State

// AgentState is deprecated. Use core/state.AgentState instead.
//
// Deprecated: Use github.com/kart-io/k8s-agent/pkg/agent/core/state.AgentState
type AgentState = state.AgentState

// NewAgentState is deprecated. Use core/state.NewAgentState instead.
//
// Deprecated: Use github.com/kart-io/k8s-agent/pkg/agent/core/state.NewAgentState
var NewAgentState = state.NewAgentState

// NewAgentStateWithData is deprecated. Use core/state.NewAgentStateWithData instead.
//
// Deprecated: Use github.com/kart-io/k8s-agent/pkg/agent/core/state.NewAgentStateWithData
var NewAgentStateWithData = state.NewAgentStateWithData
