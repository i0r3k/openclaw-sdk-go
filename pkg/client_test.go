// pkg/openclaw/client_test.go
package openclaw

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/auth"
	"github.com/frisbee-ai/openclaw-sdk-go/pkg/protocol"
	"github.com/frisbee-ai/openclaw-sdk-go/pkg/transport"
	"github.com/frisbee-ai/openclaw-sdk-go/pkg/types"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected client to not be nil")
	}
	defer func() {
		_ = client.Close()
	}()

	if client.State() != StateDisconnected {
		t.Errorf("expected disconnected state, got %s", client.State())
	}
}

func TestClientOptions(t *testing.T) {
	creds := map[string]string{"api_key": "test123"}
	authHandler, err := auth.NewStaticAuthHandler(creds)
	if err != nil {
		t.Fatalf("failed to create auth handler: %v", err)
	}

	client, err := NewClient(
		WithURL("ws://localhost:8080"),
		WithAuthHandler(authHandler),
		WithReconnect(true),
		WithReconnectConfig(&ReconnectConfig{
			MaxAttempts:       5,
			InitialDelay:      1 * time.Second,
			MaxDelay:          30 * time.Second,
			BackoffMultiplier: 2.0,
		}),
		WithEventBufferSize(200),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()
}

func TestClientEvents(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	events := client.Events()
	if events == nil {
		t.Fatal("expected events channel to not be nil")
	}

	// Test subscription
	unsubscribe := client.Subscribe(EventError, func(e Event) {
		// Handle error event
	})
	if unsubscribe == nil {
		t.Fatal("expected unsubscribe function to not be nil")
	}

	// Unsubscribe
	unsubscribe()
}

func TestClientConnectWithoutURL(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	ctx := context.Background()
	err = client.Connect(ctx)
	if err == nil {
		t.Error("expected error when connecting without URL")
	}
}

func TestClientConnectWithoutClientID(t *testing.T) {
	client, err := NewClient(WithURL("ws://localhost:8080"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	ctx := context.Background()
	err = client.Connect(ctx)
	if err == nil {
		t.Error("expected error when connecting without ClientID")
	}
}

func TestClientState(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	state := client.State()
	if state != StateDisconnected {
		t.Errorf("expected StateDisconnected, got %s", state)
	}
}

func TestClientSendRequest_NotConnected(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	ctx := context.Background()
	req := &protocol.RequestFrame{
		Type:   protocol.FrameTypeRequest,
		ID:     "test-1",
		Method: "test.action",
	}

	_, err = client.SendRequest(ctx, req)
	if err == nil {
		t.Error("expected error when sending request without connection")
	}
}

func TestClientAPIAccessors(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	// Test that API accessors return non-nil values
	if client.Chat() == nil {
		t.Error("expected non-nil Chat API")
	}
	if client.Agents() == nil {
		t.Error("expected non-nil Agents API")
	}
	if client.Sessions() == nil {
		t.Error("expected non-nil Sessions API")
	}
	if client.Config() == nil {
		t.Error("expected non-nil Config API")
	}
	if client.Cron() == nil {
		t.Error("expected non-nil Cron API")
	}
	if client.Nodes() == nil {
		t.Error("expected non-nil Nodes API")
	}
	if client.Skills() == nil {
		t.Error("expected non-nil Skills API")
	}
	if client.DevicePairing() == nil {
		t.Error("expected non-nil DevicePairing API")
	}
}

func TestClientGetServerInfo_NotConnected(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	info := client.GetServerInfo()
	if info != nil {
		t.Error("expected nil server info when not connected")
	}
}

func TestClientGetSnapshot(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	// GetSnapshot should not panic even when not connected
	_ = client.GetSnapshot()
}

func TestClientGetPolicy(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	// GetPolicy should not panic even when not connected
	_ = client.GetPolicy()
}

func TestClientGetTickMonitor(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	// GetTickMonitor should not panic even when not connected
	_ = client.GetTickMonitor()
}

func TestClientGetGapDetector(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	// GetGapDetector should not panic even when not connected
	_ = client.GetGapDetector()
}

func TestWithLogger(t *testing.T) {
	logger := &types.DefaultLogger{}
	// WithLogger should not cause error - just verify it can be set
	client, err := NewClient(WithLogger(logger))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	// Verify client works despite logger being set
	if client.State() != StateDisconnected {
		t.Errorf("expected disconnected state")
	}
}

func TestWithHeader(t *testing.T) {
	header := map[string][]string{"X-Custom-Header": {"value1", "value2"}}
	// WithHeader should not cause error - just verify it can be set
	client, err := NewClient(WithHeader(header))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	// Verify client works despite header being set
	if client.State() != StateDisconnected {
		t.Errorf("expected disconnected state")
	}
}

func TestWithTLSConfig(t *testing.T) {
	tlsConfig := &transport.TLSConfig{InsecureSkipVerify: true}
	// WithTLSConfig should not cause error - just verify it can be set
	client, err := NewClient(WithTLSConfig(tlsConfig))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	// Verify client works despite TLS config being set
	if client.State() != StateDisconnected {
		t.Errorf("expected disconnected state")
	}
}

func BenchmarkGenerateRequestID(b *testing.B) {
	for b.Loop() {
		_ = generateRequestID()
	}
}

func TestGenerateRequestID(t *testing.T) {
	// Generate multiple IDs and verify they're unique
	ids := make(map[string]bool)
	for range 100 {
		id := generateRequestID()
		if ids[id] {
			t.Errorf("duplicate request ID generated: %s", id)
		}
		ids[id] = true

		// Verify format: "req-" + 16 bytes hex = "req-" + 32 hex chars
		if !strings.HasPrefix(id, "req-") {
			t.Errorf("request ID should start with 'req-': %s", id)
		}
		if len(id) != 36 { // "req-" (4) + 32 hex chars
			t.Errorf("request ID length should be 36, got %d: %s", len(id), id)
		}
	}
}

func TestNewClient_WithAllOptions(t *testing.T) {
	creds := map[string]string{"api_key": "test123"}
	authHandler, err := auth.NewStaticAuthHandler(creds)
	if err != nil {
		t.Fatalf("failed to create auth handler: %v", err)
	}

	logger := &types.DefaultLogger{}
	header := map[string][]string{"X-Header": {"value"}}

	// Apply all options - verify they can all be set without error
	client, err := NewClient(
		WithURL("ws://localhost:8080"),
		WithClientID("test-client-id"),
		WithAuthHandler(authHandler),
		WithReconnect(true),
		WithReconnectConfig(&ReconnectConfig{
			MaxAttempts:       3,
			InitialDelay:      1 * time.Second,
			MaxDelay:          10 * time.Second,
			BackoffMultiplier: 1.5,
		}),
		WithLogger(logger),
		WithHeader(header),
		WithEventBufferSize(200),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	// Verify client is functional with all options
	if client.State() != StateDisconnected {
		t.Errorf("expected disconnected state")
	}

	// Verify API accessors work
	_ = client.Chat()
	_ = client.Agents()
	_ = client.Sessions()
}

func TestClientDisconnect_NotConnected(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	// Disconnect when not connected should not error
	err = client.Disconnect()
	if err != nil {
		t.Errorf("unexpected error on disconnect when not connected: %v", err)
	}
}

func TestClientClose_MultipleTimes(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// First close should succeed
	err = client.Close()
	if err != nil {
		t.Errorf("unexpected error on first close: %v", err)
	}

	// Second close should also succeed (idempotent)
	err = client.Close()
	if err != nil {
		t.Errorf("unexpected error on second close: %v", err)
	}
}

func TestGetMetrics_UnconnectedClient(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	metrics := client.GetMetrics()
	if metrics.Latency != 0 {
		t.Errorf("expected Latency=0 for unconnected client, got %v", metrics.Latency)
	}
	if metrics.LastTickAge != 0 {
		t.Errorf("expected LastTickAge=0 for unconnected client, got %v", metrics.LastTickAge)
	}
	if metrics.ReconnectCount != 0 {
		t.Errorf("expected ReconnectCount=0 for unconnected client, got %d", metrics.ReconnectCount)
	}
	if metrics.IsStale {
		t.Error("expected IsStale=false for unconnected client")
	}
}

func TestGetMetrics_ReturnsConnectionMetrics(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	metrics := client.GetMetrics()
	// Verify it's a ConnectionMetrics struct (type alias works)
	var _ ConnectionMetrics = metrics
}

func TestGetMetrics_ThreadSafe(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = client.GetMetrics()
		}()
	}
	wg.Wait()
	// Should not race with -race
}
