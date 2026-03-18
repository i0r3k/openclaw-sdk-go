package transport

import (
	"context"
	"crypto/tls"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketConfig holds WebSocket configuration
type WebSocketConfig struct {
	URL               string
	ReadBufferSize    int
	WriteBufferSize   int
	Header            http.Header
	TLSConfig         *TLSConfig
	HandshakeTimeout  time.Duration
	PingInterval      time.Duration
	EnableCompression bool
	ReadTimeout       time.Duration // Read timeout (default 30s)
	WriteTimeout      time.Duration // Write timeout (default 10s)
}

// TLSConfig holds TLS configuration
// Note: This is a local definition for transport layer.
// Phase 5 also defines TLSConfig for its TlsValidator with cert loading capabilities.
// Both serve different purposes and can coexist.
type TLSConfig struct {
	InsecureSkipVerify bool
	CertFile           string
	KeyFile            string
	CAFile             string
	ServerName         string
}

// CloseError represents a WebSocket close error
type CloseError struct {
	Code int
	Text string
}

func (e *CloseError) Error() string {
	return "websocket close: " + e.Text
}

// WebSocketTransport handles WebSocket communication
type WebSocketTransport struct {
	conn         *websocket.Conn
	sendCh       chan []byte
	recvCh       chan []byte
	errCh        chan error
	ctx          context.Context
	cancel       context.CancelFunc
	mu           sync.Mutex
	wg           sync.WaitGroup
	closed       bool
	readTimeout  time.Duration
	writeTimeout time.Duration
}

// Dial creates a new WebSocket connection
// Signature matches design spec: Dial(url string, header http.Header, config *WebSocketConfig)
func Dial(url string, header http.Header, config *WebSocketConfig) (*WebSocketTransport, error) {
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
		dialer.TLSClientConfig = config.TLSConfig.toTLSConfig()
	}

	conn, _, err := dialer.Dial(url, header)
	if err != nil {
		return nil, err
	}

	// Create transport struct first (needed for close handler)
	ctx, cancel := context.WithCancel(context.Background())

	// Set default timeouts if not configured
	readTimeout := 30 * time.Second
	writeTimeout := 10 * time.Second
	if config != nil {
		if config.ReadTimeout > 0 {
			readTimeout = config.ReadTimeout
		}
		if config.WriteTimeout > 0 {
			writeTimeout = config.WriteTimeout
		}
	}

	t := &WebSocketTransport{
		conn:         conn,
		sendCh:       make(chan []byte, 10),
		recvCh:       make(chan []byte, 10),
		errCh:        make(chan error, 10),
		ctx:          ctx,
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

// Start begins the send/receive loops
func (t *WebSocketTransport) Start() {
	t.wg.Add(2)
	go t.readLoop()
	go t.writeLoop()
}

// readLoop reads messages from the WebSocket
func (t *WebSocketTransport) readLoop() {
	defer t.wg.Done()
	for {
		select {
		case <-t.ctx.Done():
			return
		default:
			// Set read deadline to prevent blocking forever
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

// writeLoop writes messages to the WebSocket
func (t *WebSocketTransport) writeLoop() {
	defer t.wg.Done()
	for {
		select {
		case <-t.ctx.Done():
			return
		case message := <-t.sendCh:
			// Set write deadline to prevent blocking forever
			_ = t.conn.SetWriteDeadline(time.Now().Add(t.writeTimeout))

			if err := t.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				t.handleError(err)
				return
			}
		}
	}
}

// handleError handles WebSocket errors
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

// Send sends a message
func (t *WebSocketTransport) Send(data []byte) error {
	select {
	case t.sendCh <- data:
		return nil
	case <-t.ctx.Done():
		return t.ctx.Err()
	}
}

// Receive returns the receive channel
func (t *WebSocketTransport) Receive() <-chan []byte {
	return t.recvCh
}

// Errors returns the error channel
func (t *WebSocketTransport) Errors() <-chan error {
	return t.errCh
}

// Close closes the WebSocket connection
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
	_ = t.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	err := t.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		return t.conn.Close()
	}
	return nil
}

// IsConnected returns whether the transport is connected
func (t *WebSocketTransport) IsConnected() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.conn != nil && !t.closed
}

// Helper to convert TLSConfig to crypto/tls.Config
func (c *TLSConfig) toTLSConfig() *tls.Config {
	config := &tls.Config{
		InsecureSkipVerify: c.InsecureSkipVerify,
		ServerName:         c.ServerName,
	}

	// Note: For full implementation, load certs from files
	// This is handled by connection.TlsValidator.GetTLSConfig() in Phase 5

	return config
}

// compile-time check: WebSocketTransport implements transport interface
var _ Transport = (*WebSocketTransport)(nil)

// Transport is the interface for WebSocket transport
type Transport interface {
	Send(data []byte) error
	Receive() <-chan []byte
	Errors() <-chan error
	Close() error
	IsConnected() bool
}
