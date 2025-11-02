package errors_test

import (
	"errors"
	"testing"

	apperrors "github.com/kart-io/k8s-agent/common/errors"
)

func TestErrorXPatternChaining(t *testing.T) {
	tests := []struct {
		name     string
		buildErr func() *apperrors.AppError
		wantCode apperrors.ErrorCode
		wantMsg  string
		wantMeta map[string]string
	}{
		{
			name: "Simple error with request ID",
			buildErr: func() *apperrors.AppError {
				return apperrors.ErrNotFound.
					WithRequestID("req-123").
					WithMessage("Agent '%s' not found", "agent-1")
			},
			wantCode: apperrors.CodeNotFound,
			wantMsg:  "Agent 'agent-1' not found",
			wantMeta: map[string]string{
				"X-Request-ID": "req-123",
			},
		},
		{
			name: "Error with reason and metadata",
			buildErr: func() *apperrors.AppError {
				return apperrors.ErrValidationFailed.
					WithReason("InvalidEmail").
					KV("field", "email", "type", "required").
					WithRequestID("req-456")
			},
			wantCode: apperrors.CodeValidationFailed,
			wantMeta: map[string]string{
				"field":        "email",
				"type":         "required",
				"X-Request-ID": "req-456",
			},
		},
		{
			name: "Error with trace ID",
			buildErr: func() *apperrors.AppError {
				return apperrors.ErrInternalError.
					WithRequestID("req-789").
					WithTraceID("trace-abc").
					KV("operation", "create_user")
			},
			wantCode: apperrors.CodeInternalError,
			wantMeta: map[string]string{
				"X-Request-ID": "req-789",
				"X-Trace-ID":   "trace-abc",
				"operation":    "create_user",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.buildErr()

			// Check code
			if err.Code != tt.wantCode {
				t.Errorf("Code = %v, want %v", err.Code, tt.wantCode)
			}

			// Check message if specified
			if tt.wantMsg != "" && err.Message != tt.wantMsg {
				t.Errorf("Message = %v, want %v", err.Message, tt.wantMsg)
			}

			// Check metadata
			for k, v := range tt.wantMeta {
				if got, ok := err.Metadata[k]; !ok || got != v {
					t.Errorf("Metadata[%s] = %v, want %v", k, got, v)
				}
			}
		})
	}
}

func TestFromError(t *testing.T) {
	tests := []struct {
		name     string
		input    error
		wantNil  bool
		wantCode apperrors.ErrorCode
	}{
		{
			name:    "nil error",
			input:   nil,
			wantNil: true,
		},
		{
			name:     "AppError",
			input:    apperrors.ErrNotFound,
			wantCode: apperrors.CodeNotFound,
		},
		{
			name:     "Standard error",
			input:    errors.New("standard error"),
			wantCode: apperrors.CodeInternalError,
		},
		{
			name:     "Wrapped AppError",
			input:    apperrors.Wrap(apperrors.CodeValidationFailed, "wrapped", errors.New("base")),
			wantCode: apperrors.CodeValidationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := apperrors.FromError(tt.input)

			if tt.wantNil {
				if got != nil {
					t.Errorf("FromError() = %v, want nil", got)
				}
				return
			}

			if got == nil {
				t.Fatal("FromError() = nil, want non-nil")
			}

			if got.Code != tt.wantCode {
				t.Errorf("FromError().Code = %v, want %v", got.Code, tt.wantCode)
			}
		})
	}
}

func TestIs(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		target *apperrors.AppError
		want   bool
	}{
		{
			name:   "Match by code only",
			err:    apperrors.ErrNotFound,
			target: apperrors.ErrNotFound,
			want:   true,
		},
		{
			name:   "Match by code and reason",
			err:    apperrors.ErrValidationFailed.WithReason("InvalidEmail"),
			target: apperrors.ErrValidationFailed.WithReason("InvalidEmail"),
			want:   true,
		},
		{
			name:   "Different codes",
			err:    apperrors.ErrNotFound,
			target: apperrors.ErrInternalError,
			want:   false,
		},
		{
			name: "Same code, different reason",
			err: apperrors.New(apperrors.CodeValidationFailed, "Validation failed").
				WithReason("InvalidEmail"),
			target: apperrors.New(apperrors.CodeValidationFailed, "Validation failed").
				WithReason("RequiredField"),
			want: false,
		},
		{
			name:   "Standard error",
			err:    errors.New("standard error"),
			target: apperrors.ErrInternalError,
			want:   true, // FromError wraps as CodeInternalError
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := apperrors.Is(tt.err, tt.target)
			if got != tt.want {
				t.Errorf("Is() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPStatus(t *testing.T) {
	tests := []struct {
		name       string
		err        *apperrors.AppError
		wantStatus int
	}{
		{
			name:       "Not Found",
			err:        apperrors.ErrNotFound,
			wantStatus: 404,
		},
		{
			name:       "Internal Error",
			err:        apperrors.ErrInternalError,
			wantStatus: 500,
		},
		{
			name:       "Bad Request",
			err:        apperrors.ErrBadRequest,
			wantStatus: 400,
		},
		{
			name:       "Custom code - AgentNotFound",
			err:        apperrors.ErrAgentNotFound,
			wantStatus: 404,
		},
		{
			name:       "Custom code - WorkflowFailed",
			err:        apperrors.New(apperrors.CodeWorkflowFailed, "Workflow failed"),
			wantStatus: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.HTTPStatus()
			if got != tt.wantStatus {
				t.Errorf("HTTPStatus() = %v, want %v", got, tt.wantStatus)
			}
		})
	}
}

func TestToMap(t *testing.T) {
	err := apperrors.ErrNotFound.
		WithReason("AgentNotFound").
		WithMessage("Agent not found").
		WithRequestID("req-123").
		WithTraceID("trace-456").
		KV("agent_id", "agent-1", "cluster", "prod")

	got := err.ToMap()

	// Check code
	if got["code"].(int) != int(apperrors.CodeNotFound) {
		t.Errorf("ToMap()['code'] = %v, want %v", got["code"], apperrors.CodeNotFound)
	}

	// Check message
	if got["message"].(string) != "Agent not found" {
		t.Errorf("ToMap()['message'] = %v, want 'Agent not found'", got["message"])
	}

	// Check reason
	if got["reason"].(string) != "AgentNotFound" {
		t.Errorf("ToMap()['reason'] = %v, want 'AgentNotFound'", got["reason"])
	}

	// Check request_id
	if got["request_id"].(string) != "req-123" {
		t.Errorf("ToMap()['request_id'] = %v, want 'req-123'", got["request_id"])
	}

	// Check trace_id
	if got["trace_id"].(string) != "trace-456" {
		t.Errorf("ToMap()['trace_id'] = %v, want 'trace-456'", got["trace_id"])
	}

	// Check metadata
	metadata := got["metadata"].(map[string]string)
	if metadata["agent_id"] != "agent-1" {
		t.Errorf("ToMap()['metadata']['agent_id'] = %v, want 'agent-1'", metadata["agent_id"])
	}
	if metadata["cluster"] != "prod" {
		t.Errorf("ToMap()['metadata']['cluster'] = %v, want 'prod'", metadata["cluster"])
	}
}
