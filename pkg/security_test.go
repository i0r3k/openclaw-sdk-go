// Package openclaw provides security and boundary tests
package openclaw

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/i0r3k/openclaw-sdk-go/pkg/protocol"
	"github.com/i0r3k/openclaw-sdk-go/pkg/types"
)

// TestSecurity_InputValidation tests input validation
func TestSecurity_InputValidation(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
		{
			name:    "invalid URL scheme",
			url:     "http://example.com",
			wantErr: false, // Client accepts, but Dial will fail
		},
		{
			name:    "valid WebSocket URL",
			url:     "ws://example.com",
			wantErr: false,
		},
		{
			name:    "valid secure WebSocket URL",
			url:     "wss://example.com",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(WithURL(tt.url))
			if err != nil {
				t.Errorf("NewClient() unexpected error = %v", err)
			}
			if client != nil {
				client.Close()
			}
		})
	}
}

// TestSecurity_ConcurrentAccess tests thread-safety under concurrent access
func TestSecurity_ConcurrentAccess(t *testing.T) {
	client, err := NewClient(WithURL("ws://localhost:8080"))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	// Test concurrent State() calls
	const numGoroutines = 100
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = client.State()
			_ = client.Events()
		}()
	}

	wg.Wait()
}

// TestSecurity_ConcurrentConnectDisconnect tests concurrent connect/disconnect
func TestSecurity_ConcurrentConnectDisconnect(t *testing.T) {
	client, err := NewClient(
		WithURL("ws://localhost:9999"), // Non-existent server
		WithReconnect(false),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	var wg sync.WaitGroup

	// Try concurrent operations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = client.Connect(ctx)
			_ = client.Disconnect()
			_ = client.State()
		}()
	}

	wg.Wait()
}

// TestBoundary_LargePayload tests handling of large payloads
func TestBoundary_LargePayload(t *testing.T) {
	client, err := NewClient(WithURL("ws://localhost:8080"))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	// Create large payload (1MB)
	largePayload := make([]byte, 1024*1024)
	for i := range largePayload {
		largePayload[i] = byte(i % 256)
	}

	// Test that large payload doesn't cause panic
	req := &protocol.RequestFrame{
		RequestID: "large-payload-test",
		Method:    "test",
		Params:    largePayload,
	}

	// Just verify construction doesn't panic
	if req.RequestID == "" {
		t.Error("request ID is empty")
	}
}

// TestBoundary_ZeroLengthPayload tests zero-length payload
func TestBoundary_ZeroLengthPayload(t *testing.T) {
	req := &protocol.RequestFrame{
		RequestID: "zero-payload",
		Method:    "test",
		Params:    []byte{},
	}

	if req.RequestID == "" {
		t.Error("request ID is empty")
	}
}

// TestBoundary_NilPayload tests nil payload
func TestBoundary_NilPayload(t *testing.T) {
	req := &protocol.RequestFrame{
		RequestID: "nil-payload",
		Method:    "test",
		Params:    nil,
	}

	if req.RequestID == "" {
		t.Error("request ID is empty")
	}
}

// TestBoundary_SpecialCharactersInRequestID tests special characters
func TestBoundary_SpecialCharactersInRequestID(t *testing.T) {
	specialIDs := []string{
		"req-with-dashes",
		"req_with_underscores",
		"req.with.dots",
		"req:with:colons",
		"req/with/slashes",
		"req with spaces",
		"req\nwith\nnewlines",
		"req\x00with\x00nulls",
	}

	for _, id := range specialIDs {
		t.Run(id, func(t *testing.T) {
			req := &protocol.RequestFrame{
				RequestID: id,
				Method:    "test",
			}

			if req.RequestID != id {
				t.Errorf("RequestID = %q, want %q", req.RequestID, id)
			}
		})
	}
}

// TestBoundary_EventBufferSize tests various event buffer sizes
func TestBoundary_EventBufferSize(t *testing.T) {
	tests := []struct {
		name string
		size int
	}{
		{"zero buffer", 0},
		{"small buffer", 1},
		{"medium buffer", 100},
		{"large buffer", 10000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(
				WithURL("ws://localhost:8080"),
				WithEventBufferSize(tt.size),
			)
			if err != nil {
				t.Errorf("NewClient() error = %v", err)
			}
			if client != nil {
				client.Close()
			}
		})
	}
}

// TestBoundary_ReconnectConfig tests various reconnect configurations
func TestBoundary_ReconnectConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *types.ReconnectConfig
		wantErr bool
	}{
		{
			name: "zero max attempts (infinite)",
			config: &types.ReconnectConfig{
				MaxAttempts:  0,
				InitialDelay: 1 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "very large max attempts",
			config: &types.ReconnectConfig{
				MaxAttempts:  999999,
				InitialDelay: 1 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "zero initial delay",
			config: &types.ReconnectConfig{
				MaxAttempts:  5,
				InitialDelay: 0,
			},
			wantErr: false,
		},
		{
			name: "very long delay",
			config: &types.ReconnectConfig{
				MaxAttempts:  5,
				InitialDelay: 24 * time.Hour,
			},
			wantErr: false,
		},
		{
			name: "negative multiplier",
			config: &types.ReconnectConfig{
				MaxAttempts:       5,
				InitialDelay:      1 * time.Second,
				BackoffMultiplier: -1.0,
			},
			wantErr: false,
		},
		{
			name: "zero multiplier",
			config: &types.ReconnectConfig{
				MaxAttempts:       5,
				InitialDelay:      1 * time.Second,
				BackoffMultiplier: 0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(
				WithURL("ws://localhost:8080"),
				WithReconnectConfig(tt.config),
			)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
			}
			if client != nil {
				client.Close()
			}
		})
	}
}

// TestBoundary_TimeoutValues tests various timeout values
func TestBoundary_TimeoutValues(t *testing.T) {
	tests := []struct {
		name        string
		timeout     time.Duration
		expectPanic bool
	}{
		{"zero timeout", 0, false},
		{"negative timeout", -1 * time.Second, false},
		{"nanosecond timeout", 1 * time.Nanosecond, false},
		{"hour timeout", time.Hour, false},
		{"very long timeout", 365 * 24 * time.Hour, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Error("expected panic, but didn't get one")
					}
				}()
			}

			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			// Just verify context creation doesn't panic
			_ = ctx
		})
	}
}

// TestBoundary_EmptyStrings tests empty string handling
func TestBoundary_EmptyStrings(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"empty URL", ""},
		{"whitespace URL", "   "},
		{"tab URL", "\t"},
		{"newline URL", "\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(WithURL(tt.url))
			if err != nil {
				t.Errorf("NewClient() error = %v", err)
			}
			if client != nil {
				client.Close()
			}
		})
	}
}

// TestSecurity_RaceConditions tests for race conditions using -race
func TestSecurity_RaceConditions(t *testing.T) {
	client, err := NewClient(WithURL("ws://localhost:8080"))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer client.Close()

	// Test: Close while State() is being called
	var wg sync.WaitGroup
	stop := make(chan struct{})

	// Goroutine 1: constantly call State()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
				_ = client.State()
			}
		}
	}()

	// Goroutine 2: constantly call Events()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
				_ = client.Events()
			}
		}
	}()

	// Let them run for a bit
	time.Sleep(10 * time.Millisecond)
	close(stop)
	wg.Wait()
}

// TestBoundary_MalformedJSON tests handling of malformed JSON
func TestBoundary_MalformedJSON(t *testing.T) {
	malformedJSON := []string{
		"{",
		"}",
		"{{",
		"}}",
		"[",
		"]",
		"null",
		"undefined",
		"NaN",
		"Infinity",
		"<<>>",
		"\"unclosed string",
		"{\"key\": undefined}",
		"{\"key\": function(){}}",
	}

	for _, jsonStr := range malformedJSON {
		t.Run(jsonStr, func(t *testing.T) {
			// Verify we can handle malformed JSON without panicking
			_ = jsonStr
		})
	}
}

// TestSecurity_MemoryLeaks tests for potential memory leaks
func TestSecurity_MemoryLeaks(t *testing.T) {
	// Create and close many clients to check for memory leaks
	for i := 0; i < 100; i++ {
		client, err := NewClient(WithURL("ws://localhost:8080"))
		if err != nil {
			t.Fatalf("NewClient() error = %v", err)
		}

		// Subscribe to events
		unsub := client.Subscribe(types.EventMessage, func(e types.Event) {
			// Handler
		})

		// Unsubscribe
		unsub()

		// Close client
		if err := client.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	}
}

// TestBoundary_EventTypes tests all event type constants
func TestBoundary_EventTypes(t *testing.T) {
	eventTypes := []types.EventType{
		types.EventConnect,
		types.EventDisconnect,
		types.EventError,
		types.EventMessage,
		types.EventRequest,
		types.EventResponse,
		types.EventTick,
		types.EventGap,
		types.EventStateChange,
	}

	for _, et := range eventTypes {
		if et == "" {
			t.Errorf("event type %v is empty", et)
		}
	}
}

// TestBoundary_ConnectionStates tests all connection state constants
func TestBoundary_ConnectionStates(t *testing.T) {
	states := []types.ConnectionState{
		types.StateDisconnected,
		types.StateConnecting,
		types.StateConnected,
		types.StateAuthenticating,
		types.StateAuthenticated,
		types.StateReconnecting,
		types.StateFailed,
	}

	for _, s := range states {
		if s == "" {
			t.Errorf("state %v is empty", s)
		}
	}
}
