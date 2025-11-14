package core

import (
	"github.com/kart-io/k8s-agent/pkg/agent/core/execution"
)

// StreamingAgent is deprecated. Use execution.StreamingAgent instead.
//
// Deprecated: Use execution.StreamingAgent
type StreamingAgent = execution.StreamingAgent

// StreamOutput is deprecated. Use execution.StreamOutput instead.
//
// Deprecated: Use execution.StreamOutput
type StreamOutput = execution.StreamOutput

// LegacyStreamChunk is deprecated. Use execution.LegacyStreamChunk instead.
//
// Deprecated: Use execution.LegacyStreamChunk
type LegacyStreamChunk = execution.LegacyStreamChunk

// ChunkType is deprecated. Use execution.ChunkType instead.
//
// Deprecated: Use execution.ChunkType
type ChunkType = execution.ChunkType

const (
	// ChunkTypeText is deprecated. Use execution.ChunkTypeText instead.
	//
	// Deprecated: Use execution.ChunkTypeText
	ChunkTypeText = execution.ChunkTypeText

	// ChunkTypeJSON is deprecated. Use execution.ChunkTypeJSON instead.
	//
	// Deprecated: Use execution.ChunkTypeJSON
	ChunkTypeJSON = execution.ChunkTypeJSON

	// ChunkTypeBinary is deprecated. Use execution.ChunkTypeBinary instead.
	//
	// Deprecated: Use execution.ChunkTypeBinary
	ChunkTypeBinary = execution.ChunkTypeBinary

	// ChunkTypeProgress is deprecated. Use execution.ChunkTypeProgress instead.
	//
	// Deprecated: Use execution.ChunkTypeProgress
	ChunkTypeProgress = execution.ChunkTypeProgress

	// ChunkTypeStatus is deprecated. Use execution.ChunkTypeStatus instead.
	//
	// Deprecated: Use execution.ChunkTypeStatus
	ChunkTypeStatus = execution.ChunkTypeStatus

	// ChunkTypeError is deprecated. Use execution.ChunkTypeError instead.
	//
	// Deprecated: Use execution.ChunkTypeError
	ChunkTypeError = execution.ChunkTypeError

	// ChunkTypeMetadata is deprecated. Use execution.ChunkTypeMetadata instead.
	//
	// Deprecated: Use execution.ChunkTypeMetadata
	ChunkTypeMetadata = execution.ChunkTypeMetadata

	// ChunkTypeControl is deprecated. Use execution.ChunkTypeControl instead.
	//
	// Deprecated: Use execution.ChunkTypeControl
	ChunkTypeControl = execution.ChunkTypeControl
)

// ChunkMetadata is deprecated. Use execution.ChunkMetadata instead.
//
// Deprecated: Use execution.ChunkMetadata
type ChunkMetadata = execution.ChunkMetadata

// StreamOptions is deprecated. Use execution.StreamOptions instead.
//
// Deprecated: Use execution.StreamOptions
type StreamOptions = execution.StreamOptions

// ChunkTransformFunc is deprecated. Use execution.ChunkTransformFunc instead.
//
// Deprecated: Use execution.ChunkTransformFunc
type ChunkTransformFunc = execution.ChunkTransformFunc

// DefaultStreamOptions is deprecated. Use execution.DefaultStreamOptions instead.
//
// Deprecated: Use execution.DefaultStreamOptions
var DefaultStreamOptions = execution.DefaultStreamOptions

// StreamStatus is deprecated. Use execution.StreamStatus instead.
//
// Deprecated: Use execution.StreamStatus
type StreamStatus = execution.StreamStatus

// StreamState is deprecated. Use execution.StreamState instead.
//
// Deprecated: Use execution.StreamState
type StreamState = execution.StreamState

const (
	// StreamStateIdle is deprecated. Use execution.StreamStateIdle instead.
	//
	// Deprecated: Use execution.StreamStateIdle
	StreamStateIdle = execution.StreamStateIdle

	// StreamStateRunning is deprecated. Use execution.StreamStateRunning instead.
	//
	// Deprecated: Use execution.StreamStateRunning
	StreamStateRunning = execution.StreamStateRunning

	// StreamStatePaused is deprecated. Use execution.StreamStatePaused instead.
	//
	// Deprecated: Use execution.StreamStatePaused
	StreamStatePaused = execution.StreamStatePaused

	// StreamStateError is deprecated. Use execution.StreamStateError instead.
	//
	// Deprecated: Use execution.StreamStateError
	StreamStateError = execution.StreamStateError

	// StreamStateComplete is deprecated. Use execution.StreamStateComplete instead.
	//
	// Deprecated: Use execution.StreamStateComplete
	StreamStateComplete = execution.StreamStateComplete

	// StreamStateClosed is deprecated. Use execution.StreamStateClosed instead.
	//
	// Deprecated: Use execution.StreamStateClosed
	StreamStateClosed = execution.StreamStateClosed
)

// StreamWriter is deprecated. Use execution.StreamWriter instead.
//
// Deprecated: Use execution.StreamWriter
type StreamWriter = execution.StreamWriter

// StreamController is deprecated. Use execution.StreamController instead.
//
// Deprecated: Use execution.StreamController
type StreamController = execution.StreamController

// StreamConsumer is deprecated. Use execution.StreamConsumer instead.
//
// Deprecated: Use execution.StreamConsumer
type StreamConsumer = execution.StreamConsumer

// StreamMultiplexer is deprecated. Use execution.StreamMultiplexer instead.
//
// Deprecated: Use execution.StreamMultiplexer
type StreamMultiplexer = execution.StreamMultiplexer

// NewStreamChunk is deprecated. Use execution.NewStreamChunk instead.
//
// Deprecated: Use execution.NewStreamChunk
var NewStreamChunk = execution.NewStreamChunk

// NewTextChunk is deprecated. Use execution.NewTextChunk instead.
//
// Deprecated: Use execution.NewTextChunk
var NewTextChunk = execution.NewTextChunk

// NewProgressChunk is deprecated. Use execution.NewProgressChunk instead.
//
// Deprecated: Use execution.NewProgressChunk
var NewProgressChunk = execution.NewProgressChunk

// NewErrorChunk is deprecated. Use execution.NewErrorChunk instead.
//
// Deprecated: Use execution.NewErrorChunk
var NewErrorChunk = execution.NewErrorChunk
