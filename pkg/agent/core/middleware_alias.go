package core

import "github.com/kart-io/k8s-agent/pkg/agent/core/middleware"

// Middleware type aliases for backward compatibility.
// These aliases allow existing code to continue using middleware types from the core package.
//
// Deprecated: Use github.com/kart-io/k8s-agent/pkg/agent/core/middleware instead.
// These aliases will be removed in a future major version.

// Middleware is deprecated. Use middleware.Middleware instead.
//
// Deprecated: Use middleware.Middleware
type Middleware = middleware.Middleware

// MiddlewareRequest is deprecated. Use middleware.MiddlewareRequest instead.
//
// Deprecated: Use middleware.MiddlewareRequest
type MiddlewareRequest = middleware.MiddlewareRequest

// MiddlewareResponse is deprecated. Use middleware.MiddlewareResponse instead.
//
// Deprecated: Use middleware.MiddlewareResponse
type MiddlewareResponse = middleware.MiddlewareResponse

// Handler is deprecated. Use middleware.Handler instead.
//
// Deprecated: Use middleware.Handler
type Handler = middleware.Handler

// MiddlewareChain is deprecated. Use middleware.MiddlewareChain instead.
//
// Deprecated: Use middleware.MiddlewareChain
type MiddlewareChain = middleware.MiddlewareChain

// BaseMiddleware is deprecated. Use middleware.BaseMiddleware instead.
//
// Deprecated: Use middleware.BaseMiddleware
type BaseMiddleware = middleware.BaseMiddleware

// MiddlewareFunc is deprecated. Use middleware.MiddlewareFunc instead.
//
// Deprecated: Use middleware.MiddlewareFunc
type MiddlewareFunc = middleware.MiddlewareFunc

// LoggingMiddleware is deprecated. Use middleware.LoggingMiddleware instead.
//
// Deprecated: Use middleware.LoggingMiddleware
type LoggingMiddleware = middleware.LoggingMiddleware

// TimingMiddleware is deprecated. Use middleware.TimingMiddleware instead.
//
// Deprecated: Use middleware.TimingMiddleware
type TimingMiddleware = middleware.TimingMiddleware

// RetryMiddleware is deprecated. Use middleware.RetryMiddleware instead.
//
// Deprecated: Use middleware.RetryMiddleware
type RetryMiddleware = middleware.RetryMiddleware

// CacheMiddleware is deprecated. Use middleware.CacheMiddleware instead.
//
// Deprecated: Use middleware.CacheMiddleware
type CacheMiddleware = middleware.CacheMiddleware

// CacheEntry is deprecated. Use middleware.CacheEntry instead.
//
// Deprecated: Use middleware.CacheEntry
type CacheEntry = middleware.CacheEntry

// DynamicPromptMiddleware is deprecated. Use middleware.DynamicPromptMiddleware instead.
//
// Deprecated: Use middleware.DynamicPromptMiddleware
type DynamicPromptMiddleware = middleware.DynamicPromptMiddleware

// ToolSelectorMiddleware is deprecated. Use middleware.ToolSelectorMiddleware instead.
//
// Deprecated: Use middleware.ToolSelectorMiddleware
type ToolSelectorMiddleware = middleware.ToolSelectorMiddleware

// RateLimiterMiddleware is deprecated. Use middleware.RateLimiterMiddleware instead.
//
// Deprecated: Use middleware.RateLimiterMiddleware
type RateLimiterMiddleware = middleware.RateLimiterMiddleware

// AuthenticationMiddleware is deprecated. Use middleware.AuthenticationMiddleware instead.
//
// Deprecated: Use middleware.AuthenticationMiddleware
type AuthenticationMiddleware = middleware.AuthenticationMiddleware

// ValidationMiddleware is deprecated. Use middleware.ValidationMiddleware instead.
//
// Deprecated: Use middleware.ValidationMiddleware
type ValidationMiddleware = middleware.ValidationMiddleware

// TransformMiddleware is deprecated. Use middleware.TransformMiddleware instead.
//
// Deprecated: Use middleware.TransformMiddleware
type TransformMiddleware = middleware.TransformMiddleware

// CircuitBreakerMiddleware is deprecated. Use middleware.CircuitBreakerMiddleware instead.
//
// Deprecated: Use middleware.CircuitBreakerMiddleware
type CircuitBreakerMiddleware = middleware.CircuitBreakerMiddleware

// RandomDelayMiddleware is deprecated. Use middleware.RandomDelayMiddleware instead.
//
// Deprecated: Use middleware.RandomDelayMiddleware
type RandomDelayMiddleware = middleware.RandomDelayMiddleware

// Function aliases for backward compatibility

// NewMiddlewareChain is deprecated. Use middleware.NewMiddlewareChain instead.
//
// Deprecated: Use middleware.NewMiddlewareChain
var NewMiddlewareChain = middleware.NewMiddlewareChain

// NewBaseMiddleware is deprecated. Use middleware.NewBaseMiddleware instead.
//
// Deprecated: Use middleware.NewBaseMiddleware
var NewBaseMiddleware = middleware.NewBaseMiddleware

// NewMiddlewareFunc is deprecated. Use middleware.NewMiddlewareFunc instead.
//
// Deprecated: Use middleware.NewMiddlewareFunc
var NewMiddlewareFunc = middleware.NewMiddlewareFunc

// NewLoggingMiddleware is deprecated. Use middleware.NewLoggingMiddleware instead.
//
// Deprecated: Use middleware.NewLoggingMiddleware
var NewLoggingMiddleware = middleware.NewLoggingMiddleware

// NewTimingMiddleware is deprecated. Use middleware.NewTimingMiddleware instead.
//
// Deprecated: Use middleware.NewTimingMiddleware
var NewTimingMiddleware = middleware.NewTimingMiddleware

// NewRetryMiddleware is deprecated. Use middleware.NewRetryMiddleware instead.
//
// Deprecated: Use middleware.NewRetryMiddleware
var NewRetryMiddleware = middleware.NewRetryMiddleware

// NewCacheMiddleware is deprecated. Use middleware.NewCacheMiddleware instead.
//
// Deprecated: Use middleware.NewCacheMiddleware
var NewCacheMiddleware = middleware.NewCacheMiddleware

// NewDynamicPromptMiddleware is deprecated. Use middleware.NewDynamicPromptMiddleware instead.
//
// Deprecated: Use middleware.NewDynamicPromptMiddleware
var NewDynamicPromptMiddleware = middleware.NewDynamicPromptMiddleware

// NewToolSelectorMiddleware is deprecated. Use middleware.NewToolSelectorMiddleware instead.
//
// Deprecated: Use middleware.NewToolSelectorMiddleware
var NewToolSelectorMiddleware = middleware.NewToolSelectorMiddleware

// NewRateLimiterMiddleware is deprecated. Use middleware.NewRateLimiterMiddleware instead.
//
// Deprecated: Use middleware.NewRateLimiterMiddleware
var NewRateLimiterMiddleware = middleware.NewRateLimiterMiddleware

// NewAuthenticationMiddleware is deprecated. Use middleware.NewAuthenticationMiddleware instead.
//
// Deprecated: Use middleware.NewAuthenticationMiddleware
var NewAuthenticationMiddleware = middleware.NewAuthenticationMiddleware

// NewValidationMiddleware is deprecated. Use middleware.NewValidationMiddleware instead.
//
// Deprecated: Use middleware.NewValidationMiddleware
var NewValidationMiddleware = middleware.NewValidationMiddleware

// NewTransformMiddleware is deprecated. Use middleware.NewTransformMiddleware instead.
//
// Deprecated: Use middleware.NewTransformMiddleware
var NewTransformMiddleware = middleware.NewTransformMiddleware

// NewCircuitBreakerMiddleware is deprecated. Use middleware.NewCircuitBreakerMiddleware instead.
//
// Deprecated: Use middleware.NewCircuitBreakerMiddleware
var NewCircuitBreakerMiddleware = middleware.NewCircuitBreakerMiddleware

// NewRandomDelayMiddleware is deprecated. Use middleware.NewRandomDelayMiddleware instead.
//
// Deprecated: Use middleware.NewRandomDelayMiddleware
var NewRandomDelayMiddleware = middleware.NewRandomDelayMiddleware
