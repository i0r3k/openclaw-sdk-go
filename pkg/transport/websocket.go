// Package transport provides WebSocket transport layer for OpenClaw SDK.
//
// This package provides:
//   - WebSocketTransport: WebSocket connection and message handling
//   - Transport interface: Abstract transport interface for connection management
//   - WebSocketConfig: Configuration for WebSocket connections
//   - TLSConfig: TLS configuration for secure connections
package transport

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/connection"
	"github.com/gorilla/websocket"
)

// WebSocketConfig holds WebSocket connection configuration.
// All fields are optional and have sensible defaults.
type WebSocketConfig struct {
	URL               string        // WebSocket server URL
	ReadBufferSize    int           // Size of read buffer (default 4096)
	WriteBufferSize   int           // Size of write buffer (default 4096)
	Header            http.Header   // Custom HTTP headers for handshake
	TLSConfig         *TLSConfig    // TLS configuration for wss:// connections
	HandshakeTimeout  time.Duration // WebSocket handshake timeout (default 10s)
	PingInterval      time.Duration // Interval for ping/pong heartbeats
	EnableCompression bool          // Enable per-message compression
	ReadTimeout       time.Duration // Read timeout (default 30s)
	WriteTimeout      time.Duration // Write timeout (default 10s)
	ChannelBufferSize int           // Size of internal send/recv channels (default 64)
}

// TLSConfig holds TLS configuration for transport layer.
// Note: This is a local definition for transport layer.
// connection.TLSConfig provides cert loading capabilities for the connection layer.
// Both serve different purposes and can coexist.
type TLSConfig struct {
	InsecureSkipVerify bool   // Skip server certificate verification (insecure)
	CertFile           string // Path to client certificate file
	KeyFile            string // Path to client key file
	CAFile             string // Path to CA certificate file
	ServerName         string // Server name for SNI
}

// CloseError represents a WebSocket close error with code and text.
// Returned when the server closes the connection.
type CloseError struct {
	Code int    // WebSocket close code
	Text string // Close reason text
}

// Error returns the string representation of the close error.
func (e *CloseError) Error() string {
	return "websocket close: " + e.Text
}

// WebSocketTransport handles WebSocket communication.
// It provides a thread-safe interface for sending and receiving messages.
type WebSocketTransport struct {
	conn         *websocket.Conn    // Underlying WebSocket connection
	sendCh       chan []byte        // Channel for outgoing messages
	recvCh       chan []byte        // Channel for incoming messages
	errCh        chan error         // Channel for errors
	ctx          context.Context    // Context for cancellation
	cancel       context.CancelFunc // Cancel function for context
	mu           sync.Mutex         // Mutex for thread-safety
	wg           sync.WaitGroup     // WaitGroup for goroutines
	closed       bool               // Flag indicating if transport is closed
	readTimeout  time.Duration      // Read timeout
	writeTimeout time.Duration      // Write timeout
}

// Dial creates a new WebSocket connection.
// It establishes a WebSocket connection to the specified URL with the given headers and configuration.
// Returns a WebSocketTransport ready to be started, or an error if the connection fails.
// The context is used to cancel the entire dial operation including DNS lookup, TCP connection, and handshake.
func Dial(ctx context.Context, url string, header http.Header, config *WebSocketConfig) (*WebSocketTransport, error) {
	// Apply defaults
	readBufSize := 4096
	writeBufSize := 4096
	handshakeTimeout := 10 * time.Second

	if config != nil {
		if config.ReadBufferSize > 0 {
			readBufSize = config.ReadBufferSize
		}
		if config.WriteBufferSize > 0 {
			writeBufSize = config.WriteBufferSize
		}
		if config.HandshakeTimeout > 0 {
			handshakeTimeout = config.HandshakeTimeout
		}
	}

	dialer := websocket.Dialer{
		ReadBufferSize:   readBufSize,
		WriteBufferSize:  writeBufSize,
		HandshakeTimeout: handshakeTimeout,
	}

	// Convert TLSConfig to crypto/tls.Config if provided
	if config != nil && config.TLSConfig != nil {
		// Warn about insecure TLS configuration at dial time
		if config.TLSConfig.InsecureSkipVerify {
			fmt.Fprintf(os.Stderr, "WARNING: InsecureSkipVerify is enabled - server certificate verification is disabled. This is insecure and should only be used for testing or in controlled environments.\n")
		}

		// Use connection.TlsValidator to properly load certificates
		tlsConfig := &connection.TLSConfig{
			InsecureSkipVerify: config.TLSConfig.InsecureSkipVerify,
			CertFile:           config.TLSConfig.CertFile,
			KeyFile:            config.TLSConfig.KeyFile,
			CAFile:             config.TLSConfig.CAFile,
			ServerName:         config.TLSConfig.ServerName,
		}
		validator := connection.NewTlsValidator(tlsConfig)
		tlsClientConfig, err := validator.GetTLSConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to build TLS config: %w", err)
		}
		dialer.TLSClientConfig = tlsClientConfig
	}

	conn, _, err := dialer.DialContext(ctx, url, header)
	if err != nil {
		return nil, err
	}

	// Create transport struct first (needed for close handler)
	transportCtx, cancel := context.WithCancel(context.Background())

	// Set default timeouts if not configured
	readTimeout := 30 * time.Second
	writeTimeout := 10 * time.Second
	channelBufSize := 64
	if config != nil {
		if config.ReadTimeout > 0 {
			readTimeout = config.ReadTimeout
		}
		if config.WriteTimeout > 0 {
			writeTimeout = config.WriteTimeout
		}
		if config.ChannelBufferSize > 0 {
			channelBufSize = config.ChannelBufferSize
		}
	}

	t := &WebSocketTransport{
		conn:         conn,
		sendCh:       make(chan []byte, channelBufSize),
		recvCh:       make(chan []byte, channelBufSize),
		errCh:        make(chan error, channelBufSize),
		ctx:          transportCtx,
		cancel:       cancel,
		readTimeout:  readTimeout,
		writeTimeout: writeTimeout,
	}

	// Set up close handler to detect server-initiated closes (needs t to be defined first)
	conn.SetCloseHandler(func(code int, text string) error {
		// Propagate close event through error channel
		err := &CloseError{Code: code, Text: text}
		select {
		case t.errCh <- err:
		default:
		}
		return nil
	})

	// Configure ping/pong if interval specified
	if config != nil && config.PingInterval > 0 {
		conn.SetPingHandler(func(appData string) error {
			err := conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(config.PingInterval))
			return err
		})
	}

	return t, nil
}

// Start begins the send/receive loops.
// It launches two goroutines: one for reading and one for writing messages.
func (t *WebSocketTransport) Start() {
	t.wg.Add(2)
	go t.readLoop()
	go t.writeLoop()
}

// readLoop reads messages from the WebSocket in a dedicated goroutine.
// It sets read deadlines to prevent blocking and forwards messages to recvCh.
func (t *WebSocketTransport) readLoop() {
	defer t.wg.Done()
	for {
		select {
		case <-t.ctx.Done():
			return
		default:
			// Set read deadline to prevent blocking forever
			// Note: Ignore error - connection closed means deadline setting fails, loop exits anyway
			_ = t.conn.SetReadDeadline(time.Now().Add(t.readTimeout))

			_, message, err := t.conn.ReadMessage()
			if err != nil {
				t.handleError(err)
				return
			}
			select {
			case t.recvCh <- message:
			case <-t.ctx.Done():
				return
			}
		}
	}
}

// writeLoop writes messages to the WebSocket in a dedicated goroutine.
// It reads from sendCh and writes messages to the WebSocket.
func (t *WebSocketTransport) writeLoop() {
	defer t.wg.Done()
	for {
		select {
		case <-t.ctx.Done():
			return
		case message := <-t.sendCh:
			// Set write deadline to prevent blocking forever
			// Note: Ignore error - connection closed means deadline setting fails, loop exits anyway
			_ = t.conn.SetWriteDeadline(time.Now().Add(t.writeTimeout))

			if err := t.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				t.handleError(err)
				return
			}
		}
	}
}

// handleError handles WebSocket errors.
// It filters out context.Canceled and context.DeadlineExceeded errors
// and forwards other errors to the error channel.
func (t *WebSocketTransport) handleError(err error) {
	// Don't emit context cancelled errors
	if err == context.Canceled || err == context.DeadlineExceeded {
		return
	}

	select {
	case t.errCh <- err:
	default:
		// Channel full - don't block
	}
}

// Send sends a message through the WebSocket.
// It is thread-safe and non-blocking.
func (t *WebSocketTransport) Send(data []byte) error {
	select {
	case t.sendCh <- data:
		return nil
	case <-t.ctx.Done():
		return t.ctx.Err()
	}
}

// Receive returns the receive channel for incoming messages.
func (t *WebSocketTransport) Receive() <-chan []byte {
	return t.recvCh
}

// Errors returns the error channel for connection errors.
func (t *WebSocketTransport) Errors() <-chan error {
	return t.errCh
}

// Close closes the WebSocket connection gracefully.
// It sends a close frame and waits for the goroutines to finish.
func (t *WebSocketTransport) Close() error {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return nil
	}
	t.closed = true
	t.mu.Unlock()

	t.cancel()
	t.wg.Wait()

	// Send close frame with deadline
	// Note: Ignore error - connection may already be closed, we're cleaning up anyway
	_ = t.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	err := t.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		return t.conn.Close()
	}
	return nil
}

// IsConnected returns whether the transport is connected.
// Thread-safe method that checks if the connection is active.
func (t *WebSocketTransport) IsConnected() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.conn != nil && !t.closed
}

// compile-time check: WebSocketTransport implements transport interface
var _ Transport = (*WebSocketTransport)(nil)

// Transport is the interface for WebSocket transport.
// It abstracts the underlying WebSocket connection for the SDK.
type Transport interface {
	// Send sends a message through the transport.
	Send(data []byte) error
	// Receive returns the channel for incoming messages.
	Receive() <-chan []byte
	// Errors returns the channel for connection errors.
	Errors() <-chan error
	// Close closes the transport connection.
	Close() error
	// IsConnected returns whether the transport is connected.
	IsConnected() bool
}
