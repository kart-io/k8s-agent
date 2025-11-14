package core

import (
	"github.com/kart-io/k8s-agent/pkg/agent/core/checkpoint"
)

// Checkpointer is deprecated. Use checkpoint.Checkpointer instead.
//
// This type alias provides backward compatibility. The checkpointing functionality
// has been moved to the core/checkpoint sub-package for better organization.
//
// Migration guide: Replace imports of "github.com/kart-io/k8s-agent/pkg/agent/core"
// with "github.com/kart-io/k8s-agent/pkg/agent/core/checkpoint" for checkpoint-related types.
//
// Deprecated: Use checkpoint.Checkpointer
type Checkpointer = checkpoint.Checkpointer

// CheckpointInfo is deprecated. Use checkpoint.CheckpointInfo instead.
//
// Deprecated: Use checkpoint.CheckpointInfo
type CheckpointInfo = checkpoint.CheckpointInfo

// InMemorySaver is deprecated. Use checkpoint.InMemorySaver instead.
//
// Deprecated: Use checkpoint.InMemorySaver
type InMemorySaver = checkpoint.InMemorySaver

// RedisCheckpointer is deprecated. Use checkpoint.RedisCheckpointer instead.
//
// Deprecated: Use checkpoint.RedisCheckpointer
type RedisCheckpointer = checkpoint.RedisCheckpointer

// RedisCheckpointerConfig is deprecated. Use checkpoint.RedisCheckpointerConfig instead.
//
// Deprecated: Use checkpoint.RedisCheckpointerConfig
type RedisCheckpointerConfig = checkpoint.RedisCheckpointerConfig

// DistributedCheckpointer is deprecated. Use checkpoint.DistributedCheckpointer instead.
//
// Deprecated: Use checkpoint.DistributedCheckpointer
type DistributedCheckpointer = checkpoint.DistributedCheckpointer

// DistributedCheckpointerConfig is deprecated. Use checkpoint.DistributedCheckpointerConfig instead.
//
// Deprecated: Use checkpoint.DistributedCheckpointerConfig
type DistributedCheckpointerConfig = checkpoint.DistributedCheckpointerConfig

// CheckpointerWithAutoCleanup is deprecated. Use checkpoint.CheckpointerWithAutoCleanup instead.
//
// Deprecated: Use checkpoint.CheckpointerWithAutoCleanup
type CheckpointerWithAutoCleanup = checkpoint.CheckpointerWithAutoCleanup

// NewInMemorySaver is deprecated. Use checkpoint.NewInMemorySaver instead.
//
// Deprecated: Use checkpoint.NewInMemorySaver
var NewInMemorySaver = checkpoint.NewInMemorySaver

// NewRedisCheckpointer is deprecated. Use checkpoint.NewRedisCheckpointer instead.
//
// Deprecated: Use checkpoint.NewRedisCheckpointer
var NewRedisCheckpointer = checkpoint.NewRedisCheckpointer

// NewDistributedCheckpointer is deprecated. Use checkpoint.NewDistributedCheckpointer instead.
//
// Deprecated: Use checkpoint.NewDistributedCheckpointer
var NewDistributedCheckpointer = checkpoint.NewDistributedCheckpointer

// NewCheckpointerWithAutoCleanup is deprecated. Use checkpoint.NewCheckpointerWithAutoCleanup instead.
//
// Deprecated: Use checkpoint.NewCheckpointerWithAutoCleanup
var NewCheckpointerWithAutoCleanup = checkpoint.NewCheckpointerWithAutoCleanup
