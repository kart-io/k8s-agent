package config

import "errors"

// Configuration validation errors
var (
	ErrCentralEndpointRequired  = errors.New("central_endpoint is required")
	ErrInvalidReconnectDelay    = errors.New("reconnect_delay must be at least 1 second")
	ErrInvalidHeartbeatInterval = errors.New("heartbeat_interval must be at least 10 seconds")
	ErrInvalidMetricsInterval   = errors.New("metrics_interval must be at least 30 seconds")
	ErrInvalidBufferSize        = errors.New("buffer_size must be at least 10")
	ErrInvalidMaxRetries        = errors.New("max_retries must be at least 1")
)
