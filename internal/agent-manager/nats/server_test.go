package nats

import (
	"testing"
	"time"

	"github.com/kart-io/k8s-agent/internal/agent-manager/agent"
	"github.com/kart-io/k8s-agent/internal/agent-manager/event"
	"github.com/kart-io/logger"
	"github.com/kart-io/logger/option"
)

// TestCustomReconnectDelay tests the exponential backoff calculation.
func TestCustomReconnectDelay(t *testing.T) {
	log, err := logger.New(&option.LogOption{
		Engine: "zap",
		Level:  "info",
		Format: "console",
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Create a test server with known configuration
	server := &Server{
		options: &ServerOptions{
			reconnectDelayInitial:  1 * time.Second,
			reconnectDelayMax:      30 * time.Second,
			reconnectBackoffFactor: 2.0,
		},
		logger:                log.With("component", "nats-test"),
		currentReconnectDelay: 1 * time.Second,
	}

	tests := []struct {
		name     string
		attempts int
		expected time.Duration
	}{
		{
			name:     "First reconnection attempt",
			attempts: 1,
			expected: 2 * time.Second, // 1s * 2^0 * 2.0 = 2s
		},
		{
			name:     "Second reconnection attempt",
			attempts: 2,
			expected: 4 * time.Second, // 1s * 2^1 * 2.0 = 4s
		},
		{
			name:     "Third reconnection attempt",
			attempts: 3,
			expected: 8 * time.Second, // 1s * 2^2 * 2.0 = 8s
		},
		{
			name:     "Fourth reconnection attempt",
			attempts: 4,
			expected: 16 * time.Second, // 1s * 2^3 * 2.0 = 16s
		},
		{
			name:     "Fifth reconnection attempt",
			attempts: 5,
			expected: 30 * time.Second, // 1s * 2^4 * 2.0 = 32s, capped at 30s
		},
		{
			name:     "Sixth reconnection attempt (should cap at max)",
			attempts: 6,
			expected: 30 * time.Second, // Capped at reconnectDelayMax
		},
		{
			name:     "Large number of attempts (should stay at max)",
			attempts: 10,
			expected: 30 * time.Second, // Capped at reconnectDelayMax
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := server.customReconnectDelay(tt.attempts)

			if delay != tt.expected {
				t.Errorf("customReconnectDelay(%d) = %v, want %v", tt.attempts, delay, tt.expected)
			}

			// Verify currentReconnectDelay is updated
			if server.currentReconnectDelay != tt.expected {
				t.Errorf("currentReconnectDelay not updated correctly: got %v, want %v",
					server.currentReconnectDelay, tt.expected)
			}
		})
	}
}

// TestCustomReconnectDelayWithBackoffFactor tests different backoff factors.
func TestCustomReconnectDelayWithBackoffFactor(t *testing.T) {
	log, err := logger.New(&option.LogOption{
		Engine: "zap",
		Level:  "info",
		Format: "console",
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	tests := []struct {
		name           string
		backoffFactor  float64
		attempts       int
		expectedMin    time.Duration
		expectedMax    time.Duration
	}{
		{
			name:          "No backoff factor (1.0)",
			backoffFactor: 1.0,
			attempts:      3,
			expectedMin:   4 * time.Second,
			expectedMax:   4 * time.Second,
		},
		{
			name:          "Backoff factor 1.5",
			backoffFactor: 1.5,
			attempts:      3,
			expectedMin:   6 * time.Second, // 4s * 1.5
			expectedMax:   6 * time.Second,
		},
		{
			name:          "Backoff factor 3.0",
			backoffFactor: 3.0,
			attempts:      3,
			expectedMin:   12 * time.Second, // 4s * 3.0
			expectedMax:   12 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &Server{
				options: &ServerOptions{
					reconnectDelayInitial:  1 * time.Second,
					reconnectDelayMax:      60 * time.Second,
					reconnectBackoffFactor: tt.backoffFactor,
				},
				logger:                log.With("component", "nats-test"),
				currentReconnectDelay: 1 * time.Second,
			}

			delay := server.customReconnectDelay(tt.attempts)

			if delay < tt.expectedMin || delay > tt.expectedMax {
				t.Errorf("customReconnectDelay(%d) with factor %.1f = %v, want between %v and %v",
					tt.attempts, tt.backoffFactor, delay, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

// TestNewServerWithDefaultOptions tests server creation with default options.
func TestNewServerWithDefaultOptions(t *testing.T) {
	log, err := logger.New(&option.LogOption{
		Engine: "zap",
		Level:  "info",
		Format: "console",
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Create mock registry and event processor (nil for this test)
	var registry *agent.Registry
	var eventProc *event.Processor

	server := NewServer(registry, eventProc, log)

	if server == nil {
		t.Fatal("NewServer returned nil")
	}

	// Verify default options
	if server.options.reconnectDelayInitial != 1*time.Second {
		t.Errorf("Default reconnectDelayInitial = %v, want %v",
			server.options.reconnectDelayInitial, 1*time.Second)
	}

	if server.options.reconnectDelayMax != 30*time.Second {
		t.Errorf("Default reconnectDelayMax = %v, want %v",
			server.options.reconnectDelayMax, 30*time.Second)
	}

	if server.options.reconnectBackoffFactor != 2.0 {
		t.Errorf("Default reconnectBackoffFactor = %v, want %v",
			server.options.reconnectBackoffFactor, 2.0)
	}

	if server.currentReconnectDelay != 1*time.Second {
		t.Errorf("Initial currentReconnectDelay = %v, want %v",
			server.currentReconnectDelay, 1*time.Second)
	}
}

// TestNewServerWithCustomOptions tests server creation with custom options.
func TestNewServerWithCustomOptions(t *testing.T) {
	log, err := logger.New(&option.LogOption{
		Engine: "zap",
		Level:  "info",
		Format: "console",
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	var registry *agent.Registry
	var eventProc *event.Processor

	customInitial := 5 * time.Second
	customMax := 60 * time.Second
	customFactor := 3.0

	server := NewServer(
		registry,
		eventProc,
		log,
		WithReconnectDelayInitial(customInitial),
		WithReconnectDelayMax(customMax),
		WithReconnectBackoffFactor(customFactor),
	)

	if server == nil {
		t.Fatal("NewServer returned nil")
	}

	// Verify custom options are applied
	if server.options.reconnectDelayInitial != customInitial {
		t.Errorf("Custom reconnectDelayInitial = %v, want %v",
			server.options.reconnectDelayInitial, customInitial)
	}

	if server.options.reconnectDelayMax != customMax {
		t.Errorf("Custom reconnectDelayMax = %v, want %v",
			server.options.reconnectDelayMax, customMax)
	}

	if server.options.reconnectBackoffFactor != customFactor {
		t.Errorf("Custom reconnectBackoffFactor = %v, want %v",
			server.options.reconnectBackoffFactor, customFactor)
	}

	if server.currentReconnectDelay != customInitial {
		t.Errorf("Initial currentReconnectDelay = %v, want %v",
			server.currentReconnectDelay, customInitial)
	}
}

// TestGetStatistics tests the statistics gathering function.
func TestGetStatistics(t *testing.T) {
	log, err := logger.New(&option.LogOption{
		Engine: "zap",
		Level:  "info",
		Format: "console",
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	var registry *agent.Registry
	var eventProc *event.Processor

	server := NewServer(registry, eventProc, log)

	// Set some test values
	server.reconnectCount = 5
	server.reconnectSuccess = 4
	server.reconnectFailed = 1
	server.messagesReceived = 100
	server.messagesSent = 50
	server.errorCount = 2
	server.currentReconnectDelay = 8 * time.Second

	stats := server.GetStatistics()

	// Verify statistics
	if stats["reconnect_count"] != int64(5) {
		t.Errorf("reconnect_count = %v, want %v", stats["reconnect_count"], int64(5))
	}

	if stats["reconnect_success"] != int64(4) {
		t.Errorf("reconnect_success = %v, want %v", stats["reconnect_success"], int64(4))
	}

	if stats["reconnect_failed"] != int64(1) {
		t.Errorf("reconnect_failed = %v, want %v", stats["reconnect_failed"], int64(1))
	}

	if stats["messages_received"] != int64(100) {
		t.Errorf("messages_received = %v, want %v", stats["messages_received"], int64(100))
	}

	if stats["messages_sent"] != int64(50) {
		t.Errorf("messages_sent = %v, want %v", stats["messages_sent"], int64(50))
	}

	if stats["error_count"] != int64(2) {
		t.Errorf("error_count = %v, want %v", stats["error_count"], int64(2))
	}

	if stats["current_reconnect_delay"] != "8s" {
		t.Errorf("current_reconnect_delay = %v, want %v", stats["current_reconnect_delay"], "8s")
	}
}
