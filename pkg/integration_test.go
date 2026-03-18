// Package openclaw provides integration tests
package openclaw

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/i0r3k/openclaw-sdk-go/pkg/types"
)

// TestClient_NewClient tests basic client creation
func TestClient_NewClient(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer func() { _ = client.Close() }()

	if client.State() != StateDisconnected {
		t.Errorf("expected StateDisconnected, got %s", client.State())
	}
}

// TestClient_Options tests various client options
func TestClient_Options(t *testing.T) {
	tests := []struct {
		name    string
		options []ClientOption
	}{
		{"default", nil},
		{"reconnect", []ClientOption{WithReconnect(true)}},
		{"buffer_size", []ClientOption{WithEventBufferSize(50)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var c OpenClawClient
			var err error

			if tt.options == nil {
				c, err = NewClient()
			} else {
				c, err = NewClient(tt.options...)
			}

			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}
			defer c.Close()
		})
	}
}

// TestClient_Subscribe tests event subscription
func TestClient_Subscribe(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer func() { _ = client.Close() }()

	// Subscribe to multiple event types
	unsub1 := client.Subscribe(types.EventConnect, func(e types.Event) {})
	unsub2 := client.Subscribe(types.EventDisconnect, func(e types.Event) {})
	unsub3 := client.Subscribe(types.EventError, func(e types.Event) {})
	unsub4 := client.Subscribe(types.EventMessage, func(e types.Event) {})

	// Verify unsubscribe functions work
	if unsub1 == nil || unsub2 == nil || unsub3 == nil || unsub4 == nil {
		t.Error("unsubscribe function is nil")
	}

	// Unsubscribe
	unsub1()
	unsub2()
	unsub3()
	unsub4()
}

// TestClient_State tests state retrieval
func TestClient_State(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer func() { _ = client.Close() }()

	// Verify initial state
	if client.State() != StateDisconnected {
		t.Errorf("initial state = %s, want %s", client.State(), StateDisconnected)
	}

	// Verify all state constants are valid
	states := []ConnectionState{
		StateDisconnected, StateConnecting, StateConnected,
		StateAuthenticating, StateAuthenticated,
		StateReconnecting, StateFailed,
	}
	for _, s := range states {
		if s == "" {
			t.Error("empty state constant")
		}
	}
}

// TestClient_EventsChannel tests events channel
func TestClient_EventsChannel(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer func() { _ = client.Close() }()

	events := client.Events()
	if events == nil {
		t.Fatal("Events() returned nil channel")
	}

	// Channel should be readable
	select {
	case <-events:
	case <-time.After(10 * time.Millisecond):
	}
}

// TestClient_ConnectWithoutURL tests connecting without URL

// TestIntegration_ClientOptions_All tests all client options together
func TestIntegration_ClientOptions_All(t *testing.T) {
	cfg := types.DefaultReconnectConfig()
	client, err := NewClient(
		WithURL("ws://localhost:8080"),
		WithReconnect(true),
		WithReconnectConfig(&cfg),
		WithEventBufferSize(100),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defer func() { _ = client.Close() }()

	if client.State() != StateDisconnected {
		t.Errorf("state = %s, want %s", client.State(), StateDisconnected)
	}
}

// Helper: test WebSocket server
type testWSServer struct {
	URL      string
	server   *http.Server
	listener net.Listener
	done     chan struct{}
}

func newTestWSServer(t *testing.T) *testWSServer {
	t.Helper()

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}
			conn.WriteMessage(websocket.TextMessage, msg)
		}
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen error: %v", err)
	}

	server := &http.Server{Handler: handler}
	done := make(chan struct{})

	go func() {
		server.Serve(listener)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)

	return &testWSServer{
		URL:      "ws://" + listener.Addr().String() + "/ws",
		server:   server,
		listener: listener,
		done:     done,
	}
}

func (s *testWSServer) Close() {
	if s.server != nil {
		s.server.Shutdown(context.Background())
	}
	if s.listener != nil {
		s.listener.Close()
	}
	<-s.done
}
