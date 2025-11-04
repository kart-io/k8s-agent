package config

import "errors"

// Configuration validation errors.
var (
	ErrInvalidRateLimit = errors.New("requests_per_second must be greater than 0")
	ErrInvalidBurst     = errors.New("burst must be greater than 0")
)
