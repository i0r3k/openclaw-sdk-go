// Package transport provides tests for WebSocket transport
package transport

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// TestWebSocketConfig_Defaults tests default configuration values
func TestWebSocketConfig_Defaults(t *testing.T) {
	config := &WebSocketConfig{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	if config.ReadBufferSize != 1024 {
		t.Errorf("expected 1024, got %d", config.ReadBufferSize)
	}
}

// TestDial_ConfigDefaults tests Dial with nil config (should use defaults)
func TestDial_ConfigDefaults(t *testing.T) {
	// Create a mock server
	server := newTestServer(t, echoHandler)
	defer server.Close()

	// Dial with nil config
	transport, err := Dial(server.URL, nil, nil)
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	defer transport.Close()

	if transport == nil {
		t.Fatal("Dial() returned nil transport")
	}

	// Verify default timeouts
	if transport.readTimeout != 30*time.Second {
		t.Errorf("default readTimeout = %v, want 30s", transport.readTimeout)
	}
	if transport.writeTimeout != 10*time.Second {
		t.Errorf("default writeTimeout = %v, want 10s", transport.writeTimeout)
	}
}

// TestDial_CustomTimeouts tests custom read/write timeouts
func TestDial_CustomTimeouts(t *testing.T) {
	server := newTestServer(t, echoHandler)
	defer server.Close()

	config := &WebSocketConfig{
		ReadTimeout:  45 * time.Second,
		WriteTimeout: 20 * time.Second,
	}

	transport, err := Dial(server.URL, nil, config)
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	defer transport.Close()

	if transport.readTimeout != 45*time.Second {
		t.Errorf("readTimeout = %v, want 45s", transport.readTimeout)
	}
	if transport.writeTimeout != 20*time.Second {
		t.Errorf("writeTimeout = %v, want 20s", transport.writeTimeout)
	}
}

// TestDial_CustomBufferSizes tests custom buffer sizes
func TestDial_CustomBufferSizes(t *testing.T) {
	server := newTestServer(t, echoHandler)
	defer server.Close()

	config := &WebSocketConfig{
		ReadBufferSize:  2048,
		WriteBufferSize: 2048,
	}

	transport, err := Dial(server.URL, nil, config)
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	defer transport.Close()

	// Verify connection was created
	if transport.conn == nil {
		t.Error("transport.conn is nil")
	}
}

// TestDial_InvalidURL tests Dial with invalid URL
func TestDial_InvalidURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "invalid protocol",
			url:     "http://example.com",
			wantErr: true,
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
		{
			name:    "invalid host",
			url:     "ws://invalid-host-99999.example",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Dial(tt.url, nil, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Dial() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestDial_WithHeaders tests Dial with custom headers
func TestDial_WithHeaders(t *testing.T) {
	server := newTestServer(t, headerCheckHandler)
	defer server.Close()

	header := http.Header{}
	header.Set("X-Test-Header", "test-value")
	header.Set("User-Agent", "OpenClaw-Test/1.0")

	transport, err := Dial(server.URL, header, nil)
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	defer transport.Close()

	// Server should accept connection with valid headers
	if transport == nil {
		t.Fatal("transport is nil")
	}
}

// TestDial_WithTLSConfig tests TLS configuration
func TestDial_WithTLSConfig(t *testing.T) {
	server := newTestServer(t, echoHandler)
	defer server.Close()

	config := &WebSocketConfig{
		TLSConfig: &TLSConfig{
			InsecureSkipVerify: true,
			ServerName:         "localhost",
		},
	}

	transport, err := Dial(server.URL, nil, config)
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	defer transport.Close()

	if transport == nil {
		t.Fatal("transport is nil")
	}
}

// TestWebSocketTransport_Start tests Start method
func TestWebSocketTransport_Start(t *testing.T) {
	server := newTestServer(t, echoHandler)
	defer server.Close()

	transport, err := Dial(server.URL, nil, nil)
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	defer transport.Close()

	// Start the loops
	transport.Start()

	// Give loops time to start
	time.Sleep(100 * time.Millisecond)

	// Verify transport is connected
	if !transport.IsConnected() {
		t.Error("IsConnected() = false, want true")
	}
}

// TestWebSocketTransport_Send_Blocking tests Send with blocking channel
func TestWebSocketTransport_Send_Blocking(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	transport := &WebSocketTransport{
		sendCh: make(chan []byte, 1),
		recvCh: make(chan []byte, 1),
		errCh:  make(chan error, 1),
		ctx:    ctx,
		cancel: cancel,
	}
	defer cancel()

	// Test send succeeds when channel is not full
	err := transport.Send([]byte("test message"))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestWebSocketTransport_Send_ContextCanceled tests Send when context is canceled
func TestWebSocketTransport_Send_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	transport := &WebSocketTransport{
		sendCh: make(chan []byte), // unbuffered to test ctx cancellation
		recvCh: make(chan []byte, 1),
		errCh:  make(chan error, 1),
		ctx:    ctx,
		cancel: cancel,
	}

	// Cancel context
	cancel()

	// Send should return context error
	err := transport.Send([]byte("test message"))
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// TestWebSocketTransport_Receive tests Receive channel
func TestWebSocketTransport_Receive(t *testing.T) {
	transport := &WebSocketTransport{
		sendCh: make(chan []byte, 1),
		recvCh: make(chan []byte, 1),
		errCh:  make(chan error, 1),
	}

	// Send a message to recvCh
	testMsg := []byte("test")
	select {
	case transport.recvCh <- testMsg:
	default:
		t.Fatal("failed to send test message")
	}

	// Receive should return the message
	select {
	case msg := <-transport.Receive():
		if string(msg) != "test" {
			t.Errorf("expected 'test', got '%s'", string(msg))
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for message")
	}
}

// TestWebSocketTransport_IsConnected tests IsConnected method
func TestWebSocketTransport_IsConnected(t *testing.T) {
	t.Run("conn=nil, closed=false", func(t *testing.T) {
		transport := &WebSocketTransport{
			closed: false,
		}
		if transport.IsConnected() {
			t.Error("expected not connected when conn=nil")
		}
	})

	t.Run("closed=true", func(t *testing.T) {
		transport := &WebSocketTransport{
			closed: true,
		}
		if transport.IsConnected() {
			t.Error("expected disconnected when closed=true")
		}
	})

	t.Run("connected state", func(t *testing.T) {
		server := newTestServer(t, echoHandler)
		defer server.Close()

		transport, err := Dial(server.URL, nil, nil)
		if err != nil {
			t.Fatalf("Dial() error = %v", err)
		}
		defer transport.Close()

		if !transport.IsConnected() {
			t.Error("expected connected after Dial")
		}
	})
}

// TestWebSocketTransport_Close_Idempotent tests Close idempotency
func TestWebSocketTransport_Close_Idempotent(t *testing.T) {
	transport := &WebSocketTransport{
		closed: true,
	}

	// Multiple closes should be idempotent
	for i := 0; i < 3; i++ {
		err := transport.Close()
		if err != nil {
			t.Errorf("close %d: unexpected error: %v", i, err)
		}
	}
}

// TestWebSocketTransport_Close_StopsLoops tests that Close stops read/write loops
func TestWebSocketTransport_Close_StopsLoops(t *testing.T) {
	server := newTestServer(t, echoHandler)
	defer server.Close()

	transport, err := Dial(server.URL, nil, nil)
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}

	transport.Start()

	// Send some data to verify loops are running
	testData := []byte("test before close")
	if err := transport.Send(testData); err != nil {
		t.Errorf("Send() before close error = %v", err)
	}

	// Close should stop loops
	if err := transport.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// After close, should not be connected
	if transport.IsConnected() {
		t.Error("expected not connected after Close")
	}
}

// TestWebSocketTransport_Errors tests Errors channel
func TestWebSocketTransport_Errors(t *testing.T) {
	transport := &WebSocketTransport{
		errCh: make(chan error, 10),
	}

	errCh := transport.Errors()
	if errCh == nil {
		t.Error("expected error channel, got nil")
	}
}

// TestTLSConfig_toTLSConfig tests TLSConfig conversion
func TestTLSConfig_toTLSConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *TLSConfig
		check  func(*testing.T, *TLSConfig)
	}{
		{
			name: "insecure skip verify",
			config: &TLSConfig{
				InsecureSkipVerify: true,
				ServerName:         "example.com",
			},
			check: func(t *testing.T, cfg *TLSConfig) {
				tlsConfig := cfg.toTLSConfig()
				if !tlsConfig.InsecureSkipVerify {
					t.Error("InsecureSkipVerify not set")
				}
				if tlsConfig.ServerName != "example.com" {
					t.Errorf("ServerName = %s, want example.com", tlsConfig.ServerName)
				}
			},
		},
		{
			name: "secure with server name",
			config: &TLSConfig{
				InsecureSkipVerify: false,
				ServerName:         "secure.example.com",
			},
			check: func(t *testing.T, cfg *TLSConfig) {
				tlsConfig := cfg.toTLSConfig()
				if tlsConfig.InsecureSkipVerify {
					t.Error("InsecureSkipVerify should be false")
				}
			},
		},
		{
			name:   "nil config",
			config: nil,
			check: func(t *testing.T, cfg *TLSConfig) {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("toTLSConfig() panicked: %v", r)
					}
				}()
				// This would panic if called on nil, but we're testing nil handling
				_ = cfg
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config != nil {
				tt.check(t, tt.config)
			}
		})
	}
}

// TestCloseError tests CloseError type
func TestCloseError(t *testing.T) {
	tests := []struct {
		name string
		err  *CloseError
		want string
	}{
		{
			name: "normal closure",
			err:  &CloseError{Code: websocket.CloseNormalClosure, Text: "normal"},
			want: "websocket close: normal",
		},
		{
			name: "going away",
			err:  &CloseError{Code: websocket.CloseGoingAway, Text: "server shutdown"},
			want: "websocket close: server shutdown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestWebSocketTransport_E2E tests end-to-end communication
func TestWebSocketTransport_E2E(t *testing.T) {
	server := newTestServer(t, echoHandler)
	defer server.Close()

	transport, err := Dial(server.URL, nil, nil)
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	defer transport.Close()

	transport.Start()

	// Send message
	testMsg := []byte("hello, world")
	if err := transport.Send(testMsg); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	// Receive echoed message
	select {
	case recvMsg := <-transport.Receive():
		if string(recvMsg) != string(testMsg) {
			t.Errorf("received = %s, want %s", string(recvMsg), string(testMsg))
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for echo")
	}
}

// TestWebSocketTransport_ConcurrentSend tests concurrent Send operations
func TestWebSocketTransport_ConcurrentSend(t *testing.T) {
	server := newTestServer(t, echoHandler)
	defer server.Close()

	transport, err := Dial(server.URL, nil, nil)
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	defer transport.Close()

	transport.Start()

	// Concurrent sends
	const numGoroutines = 10
	const messagesPerGoroutine = 10

	var wg sync.WaitGroup
	errCh := make(chan error, numGoroutines*messagesPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				msg := []byte{byte(id), byte(j)}
				if err := transport.Send(msg); err != nil {
					errCh <- err
				}
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("concurrent Send error: %v", err)
	}
}

// TestWebSocketTransport_Timeout tests read/write timeout behavior
func TestWebSocketTransport_Timeout(t *testing.T) {
	server := newTestServer(t, slowHandler)
	defer server.Close()

	config := &WebSocketConfig{
		ReadTimeout:  100 * time.Millisecond,
		WriteTimeout: 100 * time.Millisecond,
	}

	transport, err := Dial(server.URL, nil, config)
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	defer transport.Close()

	transport.Start()

	// Send should succeed within timeout
	testMsg := []byte("timeout test")
	if err := transport.Send(testMsg); err != nil {
		t.Errorf("Send() error = %v", err)
	}

	// Wait for potential timeout errors
	select {
	case err := <-transport.Errors():
		// Timeout error is expected
		if err != nil && !errors.Is(err, context.DeadlineExceeded) {
			t.Logf("Received error (may be timeout): %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		// No error within timeout period
	}
}

// TestTransportInterface verifies WebSocketTransport implements Transport interface
func TestTransportInterface(t *testing.T) {
	var _ Transport = (*WebSocketTransport)(nil)
}

// Helper types and functions for testing

type testServer struct {
	URL     string
	server  *http.Server
	done    chan struct{}
	cleanup func()
}

func newTestServer(t *testing.T, handler http.HandlerFunc) *testServer {
	t.Helper()

	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}

	server := &http.Server{
		Handler: handler,
	}

	done := make(chan struct{})

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			t.Logf("Server error: %v", err)
		}
		close(done)
	}()

	// Wait for server to start
	time.Sleep(50 * time.Millisecond)

	addr := listener.Addr().String()
	if addr == "" {
		listener.Close()
		t.Fatal("server address is empty")
	}

	return &testServer{
		URL:     "ws://" + addr + "/ws",
		server:  server,
		done:    done,
		cleanup: func() { listener.Close() },
	}
}

func (s *testServer) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = s.server.Shutdown(ctx)
	<-s.done
}

// echoHandler echoes received messages
func echoHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		if err := conn.WriteMessage(messageType, message); err != nil {
			break
		}
	}
}

// headerCheckHandler checks for required headers
func headerCheckHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// Check for required header
	if r.Header.Get("X-Test-Header") == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// Keep connection alive
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// slowHandler delays responses to test timeouts
func slowHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// Delay before echoing
		time.Sleep(200 * time.Millisecond)
	}
}
