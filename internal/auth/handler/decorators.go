package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/kart-io/k8s-agent/common/response"
)

// HandlerFunc defines a typed handler function that returns data and error
// This allows us to write handlers that focus only on business logic
type HandlerFunc[Req any, Res any] func(c *gin.Context, req *Req) (*Res, error)

// NoRequestHandlerFunc defines a handler that doesn't need request body
type NoRequestHandlerFunc[Res any] func(c *gin.Context) (*Res, error)

// NoResponseHandlerFunc defines a handler that doesn't return data
type NoResponseHandlerFunc[Req any] func(c *gin.Context, req *Req) error

// WithJSONRequest creates a decorator for handlers with JSON request body
// It handles request binding, validation, business logic execution, and response formatting
func WithJSONRequest[Req any, Res any](handler HandlerFunc[Req, Res]) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Req

		// Bind JSON request
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "Invalid request format", err)
			return
		}

		// Execute business logic
		result, err := handler(c, &req)
		if err != nil {
			handleBusinessError(c, err)
			return
		}

		// Success response
		response.Success(c, result)
	}
}

// WithQueryParams creates a decorator for handlers with query parameters
func WithQueryParams[Req any, Res any](handler HandlerFunc[Req, Res]) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Req

		// Bind query parameters
		if err := c.ShouldBindQuery(&req); err != nil {
			response.BadRequest(c, "Invalid query parameters", err)
			return
		}

		// Execute business logic
		result, err := handler(c, &req)
		if err != nil {
			handleBusinessError(c, err)
			return
		}

		// Success response
		response.Success(c, result)
	}
}

// WithURIParams creates a decorator for handlers with URI parameters
func WithURIParams[Req any, Res any](handler HandlerFunc[Req, Res]) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Req

		// Bind URI parameters
		if err := c.ShouldBindUri(&req); err != nil {
			response.BadRequest(c, "Invalid URI parameters", err)
			return
		}

		// Execute business logic
		result, err := handler(c, &req)
		if err != nil {
			handleBusinessError(c, err)
			return
		}

		// Success response
		response.Success(c, result)
	}
}

// WithNoRequest creates a decorator for handlers without request body
// Useful for GET endpoints that only need context
func WithNoRequest[Res any](handler NoRequestHandlerFunc[Res]) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Execute business logic
		result, err := handler(c)
		if err != nil {
			handleBusinessError(c, err)
			return
		}

		// Success response
		response.Success(c, result)
	}
}

// WithJSONRequestNoResponse creates a decorator for handlers with request but no response data
// Useful for DELETE, UPDATE endpoints that return success message only
func WithJSONRequestNoResponse[Req any](handler NoResponseHandlerFunc[Req]) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Req

		// Bind JSON request
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "Invalid request format", err)
			return
		}

		// Execute business logic
		if err := handler(c, &req); err != nil {
			handleBusinessError(c, err)
			return
		}

		// Success response without data
		response.SuccessWithMessage(c, "Operation completed successfully", nil)
	}
}

// WithJSONRequestCreated creates a decorator for POST handlers that create resources
// Returns 201 Created status instead of 200 OK
func WithJSONRequestCreated[Req any, Res any](handler HandlerFunc[Req, Res]) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Req

		// Bind JSON request
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "Invalid request format", err)
			return
		}

		// Execute business logic
		result, err := handler(c, &req)
		if err != nil {
			handleBusinessError(c, err)
			return
		}

		// Created response (201)
		response.Created(c, result)
	}
}

// WithURIParamsNoResponse creates a decorator for handlers with URI params but no response data
func WithURIParamsNoResponse[Req any](handler NoResponseHandlerFunc[Req]) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Req

		// Bind URI parameters
		if err := c.ShouldBindUri(&req); err != nil {
			response.BadRequest(c, "Invalid URI parameters", err)
			return
		}

		// Execute business logic
		if err := handler(c, &req); err != nil {
			handleBusinessError(c, err)
			return
		}

		// Success response without data
		response.SuccessWithMessage(c, "Operation completed successfully", nil)
	}
}

// WithURIAndJSONRequest creates a decorator for handlers with both URI params and JSON body
func WithURIAndJSONRequest[Req any](handler NoResponseHandlerFunc[Req]) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Req

		// Bind URI parameters first
		if err := c.ShouldBindUri(&req); err != nil {
			response.BadRequest(c, "Invalid URI parameters", err)
			return
		}

		// Bind JSON body
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "Invalid request format", err)
			return
		}

		// Execute business logic
		if err := handler(c, &req); err != nil {
			handleBusinessError(c, err)
			return
		}

		// Success response without data
		response.SuccessWithMessage(c, "Operation completed successfully", nil)
	}
}

// WithURIAndJSONRequestNoResponse creates a decorator for handlers with URI params and JSON body but no response
func WithURIAndJSONRequestNoResponse[Req any](handler NoResponseHandlerFunc[Req]) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Req

		// Bind URI parameters first
		if err := c.ShouldBindUri(&req); err != nil {
			response.BadRequest(c, "Invalid URI parameters", err)
			return
		}

		// Bind JSON body
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "Invalid request format", err)
			return
		}

		// Execute business logic
		if err := handler(c, &req); err != nil {
			handleBusinessError(c, err)
			return
		}

		// Success response without data
		response.SuccessWithMessage(c, "Operation completed successfully", nil)
	}
}

// handleBusinessError handles errors returned from business logic
// This function can be extended to handle custom error types
func handleBusinessError(c *gin.Context, err error) {
	// For now, return as internal server error
	// TODO: Implement custom error types to return specific HTTP status codes
	// For example:
	// - NotFoundError -> 404
	// - ValidationError -> 400
	// - UnauthorizedError -> 401
	// - ForbiddenError -> 403
	response.InternalError(c, "Operation failed", err)
}
