// pkg/openclaw/client_test.go
package openclaw

import (
	"context"
	"testing"
	"time"

	"github.com/i0r3k/openclaw-sdk-go/pkg/auth"
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
