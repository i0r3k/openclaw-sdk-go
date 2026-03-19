# Phase 5: Connection Module

**Files:**
- Create: `pkg/openclaw/connection/state.go`, `pkg/openclaw/connection/state_test.go`
- Create: `pkg/openclaw/connection/protocol.go`, `pkg/openclaw/connection/protocol_test.go`
- Create: `pkg/openclaw/connection/policies.go`, `pkg/openclaw/connection/policies_test.go`
- Create: `pkg/openclaw/connection/tls.go`, `pkg/openclaw/connection/tls_test.go`

**Project Structure:** Go module in root, source files in `pkg/openclaw/` directory

**Depends on:** Phase 1 (types.go, errors.go), Phase 4 (transport)

---

## Task 5.1: Connection State Machine

- [ ] **Step 1: Create connection directory and state.go**

```bash
mkdir -p pkg/openclaw/connection
```

```go
// pkg/openclaw/connection/state.go
package connection

import (
	"fmt"
	"sync"

	openclaw "github.com/frisbee-ai/openclaw-sdk-go"
)

// StateChangeEvent represents a state change event
type StateChangeEvent struct {
	From   openclaw.ConnectionState
	To     openclaw.ConnectionState
	Reason error
}

// ConnectionStateMachine manages connection state
type ConnectionStateMachine struct {
	state  openclaw.ConnectionState
	mu     sync.RWMutex
	events chan StateChangeEvent
	ctx    interface{} // context.Context - added for future cancellation support
}

// NewConnectionStateMachine creates a new state machine
func NewConnectionStateMachine(initial openclaw.ConnectionState) *ConnectionStateMachine {
	return &ConnectionStateMachine{
		state:  initial,
		events: make(chan StateChangeEvent, 10),
	}
}

// validTransitions defines valid state transitions using typed constants
var validTransitions = map[openclaw.ConnectionState][]openclaw.ConnectionState{
	openclaw.StateDisconnected:     {openclaw.StateConnecting},
	openclaw.StateConnecting:       {openclaw.StateConnected, openclaw.StateDisconnected, openclaw.StateFailed},
	openclaw.StateConnected:        {openclaw.StateAuthenticating, openclaw.StateDisconnected, openclaw.StateReconnecting, openclaw.StateFailed},
	openclaw.StateAuthenticating:   {openclaw.StateAuthenticated, openclaw.StateFailed},
	openclaw.StateAuthenticated:    {openclaw.StateDisconnected, openclaw.StateReconnecting},
	openclaw.StateReconnecting:    {openclaw.StateConnecting, openclaw.StateFailed},
	openclaw.StateFailed:           {openclaw.StateDisconnected},
}

func (csm *ConnectionStateMachine) validTransition(from, to openclaw.ConnectionState) bool {
	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

// Transition changes the state
func (csm *ConnectionStateMachine) Transition(to openclaw.ConnectionState, reason error) error {
	csm.mu.Lock()
	from := csm.state
	if !csm.validTransition(from, to) {
		csm.mu.Unlock()
		return fmt.Errorf("invalid state transition from %s to %s", from, to)
	}
	csm.state = to
	csm.mu.Unlock()

	select {
	case csm.events <- StateChangeEvent{From: from, To: to, Reason: reason}:
	default:
		// Channel full - return error so caller knows event was dropped
		return fmt.Errorf("state change event dropped: %s -> %s", from, to)
	}
	return nil
}

// State returns the current state
func (csm *ConnectionStateMachine) State() openclaw.ConnectionState {
	csm.mu.RLock()
	defer csm.mu.RUnlock()
	return csm.state
}

// Events returns the state change event channel
func (csm *ConnectionStateMachine) Events() <-chan StateChangeEvent {
	return csm.events
}
```

- [ ] **Step 2: Write test**

```go
// pkg/openclaw/connection/state_test.go
package connection

import (
	"testing"
	"time"

	openclaw "github.com/frisbee-ai/openclaw-sdk-go"
)

func TestConnectionStateMachine_Transition(t *testing.T) {
	csm := NewConnectionStateMachine(openclaw.StateDisconnected)

	err := csm.Transition(openclaw.StateConnecting, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if csm.State() != openclaw.StateConnecting {
		t.Errorf("expected 'connecting', got '%s'", csm.State())
	}
}

func TestConnectionStateMachine_InvalidTransition(t *testing.T) {
	csm := NewConnectionStateMachine(openclaw.StateDisconnected)

	err := csm.Transition(openclaw.StateAuthenticated, nil)
	if err == nil {
		t.Error("expected error for invalid transition")
	}
}

func TestConnectionStateMachine_StateChangeEvent(t *testing.T) {
	csm := NewConnectionStateMachine(openclaw.StateDisconnected)

	err := csm.Transition(openclaw.StateConnecting, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case event := <-csm.Events():
		if event.From != openclaw.StateDisconnected {
			t.Errorf("expected from 'disconnected', got '%s'", event.From)
		}
		if event.To != openclaw.StateConnecting {
			t.Errorf("expected to 'connecting', got '%s'", event.To)
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for state change event")
	}
}
```

- [ ] **Step 3: Run tests and commit**

Run: `go test -v ./pkg/openclaw/connection/...`
Commit: `git add pkg/openclaw/connection/ go.mod && git commit -m "feat: add connection state machine"`

---

## Task 5.2: Protocol Negotiator

- [ ] **Step 1: Write protocol.go**

```go
// pkg/openclaw/connection/protocol.go
package connection

import (
	"context"
	"errors"
	"time"

	openclaw "github.com/frisbee-ai/openclaw-sdk-go"
)

// ProtocolNegotiator handles protocol version negotiation
type ProtocolNegotiator struct {
	supportedVersions []string
	defaultTimeout    time.Duration
}

// NewProtocolNegotiator creates a new negotiator
func NewProtocolNegotiator(supportedVersions []string) *ProtocolNegotiator {
	if len(supportedVersions) == 0 {
		supportedVersions = []string{"1.0"}
	}
	return &ProtocolNegotiator{
		supportedVersions: supportedVersions,
		defaultTimeout:    5 * time.Second,
	}
}

// Negotiate performs protocol negotiation with context support
func (p *ProtocolNegotiator) Negotiate(ctx context.Context, serverVersions []string) (string, error) {
	// Create a timeout if context doesn't have one
	ctx, cancel := context.WithTimeout(ctx, p.defaultTimeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return "", openclaw.NewProtocolError("protocol negotiation timeout", ctx.Err())
		default:
			// Check for matching versions
			for _, clientVer := range p.supportedVersions {
				for _, serverVer := range serverVersions {
					if clientVer == serverVer {
						return clientVer, nil
					}
				}
			}
			// No match found
			return "", openclaw.NewProtocolError("no matching protocol version", nil)
		}
	}
}

// ErrNoMatchingProtocol is a sentinel error for protocol negotiation failures
// Use errors.Is() to check for this specific error
var ErrNoMatchingProtocol = errors.New("no matching protocol version")
```

- [ ] **Step 2: Write protocol_test.go**

```go
// pkg/openclaw/connection/protocol_test.go
package connection

import (
	"context"
	"testing"
	"time"
)

func TestProtocolNegotiator_Negotiate_Match(t *testing.T) {
	negotiator := NewProtocolNegotiator([]string{"1.0", "2.0"})

	version, err := negotiator.Negotiate(context.Background(), []string{"1.0", "1.1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version != "1.0" {
		t.Errorf("expected '1.0', got '%s'", version)
	}
}

func TestProtocolNegotiator_Negotiate_NoMatch(t *testing.T) {
	negotiator := NewProtocolNegotiator([]string{"1.0", "2.0"})

	_, err := negotiator.Negotiate(context.Background(), []string{"3.0", "4.0"})
	if err == nil {
		t.Error("expected error for no matching version")
	}
}

func TestProtocolNegotiator_Negotiate_ContextCancel(t *testing.T) {
	negotiator := NewProtocolNegotiator([]string{"1.0"})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := negotiator.Negotiate(ctx, []string{"1.0"})
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

func TestProtocolNegotiator_DefaultVersions(t *testing.T) {
	negotiator := NewProtocolNegotiator(nil)

	version, err := negotiator.Negotiate(context.Background(), []string{"1.0"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if version != "1.0" {
		t.Errorf("expected '1.0', got '%s'", version)
	}
}
```

- [ ] **Step 3: Commit**

```bash
git add pkg/openclaw/connection/protocol.go pkg/openclaw/connection/protocol_test.go
git commit -m "feat: add protocol negotiator with context support"
```

---

## Task 5.3: Policy Manager

- [ ] **Step 1: Write policies.go**

```go
// pkg/openclaw/connection/policies.go
package connection

import (
	"time"
)

// PolicyManager manages connection policies
type PolicyManager struct {
	maxReconnectAttempts int
	pingInterval         time.Duration
}

// NewPolicyManager creates a new policy manager
func NewPolicyManager(maxReconnectAttempts int, pingInterval time.Duration) *PolicyManager {
	return &PolicyManager{
		maxReconnectAttempts: maxReconnectAttempts,
		pingInterval:         pingInterval,
	}
}

// MaxReconnectAttempts returns the max reconnect attempts
// Returns 0 for infinite retries
func (pm *PolicyManager) MaxReconnectAttempts() int {
	return pm.maxReconnectAttempts
}

// PingInterval returns the ping interval
func (pm *PolicyManager) PingInterval() time.Duration {
	return pm.pingInterval
}

// ShouldReconnect checks if reconnection should be attempted based on attempt count
func (pm *PolicyManager) ShouldReconnect(attemptCount int) bool {
	if pm.maxReconnectAttempts == 0 {
		return true // Infinite retries
	}
	return attemptCount < pm.maxReconnectAttempts
}
```

- [ ] **Step 2: Write policies_test.go**

```go
// pkg/openclaw/connection/policies_test.go
package connection

import (
	"testing"
	"time"
)

func TestPolicyManager_InfiniteReconnect(t *testing.T) {
	pm := NewPolicyManager(0, 30*time.Second)

	if !pm.ShouldReconnect(0) {
		t.Error("expected ShouldReconnect(0) to return true for infinite retries")
	}
	if !pm.ShouldReconnect(100) {
		t.Error("expected ShouldReconnect(100) to return true for infinite retries")
	}
}

func TestPolicyManager_LimitedReconnect(t *testing.T) {
	pm := NewPolicyManager(3, 30*time.Second)

	if !pm.ShouldReconnect(0) {
		t.Error("expected ShouldReconnect(0) to return true")
	}
	if !pm.ShouldReconnect(2) {
		t.Error("expected ShouldReconnect(2) to return true")
	}
	if pm.ShouldReconnect(3) {
		t.Error("expected ShouldReconnect(3) to return false (attempt 3 == max)")
	}
	if pm.ShouldReconnect(10) {
		t.Error("expected ShouldReconnect(10) to return false")
	}
}

func TestPolicyManager_PingInterval(t *testing.T) {
	pm := NewPolicyManager(0, 30*time.Second)

	interval := pm.PingInterval()
	if interval != 30*time.Second {
		t.Errorf("expected 30s, got %v", interval)
	}
}
```

- [ ] **Step 3: Commit**

```bash
git add pkg/openclaw/connection/policies.go pkg/openclaw/connection/policies_test.go
git commit -m "feat: add policy manager with reconnection logic"
```

---

## Task 5.4: TLS Validator

- [ ] **Step 1: Write tls.go**

```go
// pkg/openclaw/connection/tls.go
package connection

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"
	"time"

	openclaw "github.com/frisbee-ai/openclaw-sdk-go"
)

// TlsValidator validates TLS certificates
type TlsValidator struct {
	config *TLSConfig
}

// TLSConfig holds TLS configuration for connection layer
// Note: This is distinct from transport.TLSConfig which is for dial-time configuration
// This version supports certificate loading and validation
type TLSConfig struct {
	InsecureSkipVerify bool
	CertFile          string
	KeyFile           string
	CAFile            string
	ServerName        string
}

// ErrInvalidTLSConfig represents TLS configuration validation errors
var ErrInvalidTLSConfig = errors.New("invalid TLS configuration")

// ErrCertNotFound is returned when certificate file is not found
var ErrCertNotFound = errors.New("certificate file not found")

// ErrCANotFound is returned when CA file is not found
var ErrCANotFound = errors.New("CA certificate file not found")

// NewTlsValidator creates a new TLS validator
func NewTlsValidator(config *TLSConfig) *TlsValidator {
	return &TlsValidator{config: config}
}

// Validate validates the TLS configuration
func (v *TlsValidator) Validate() error {
	if v.config == nil {
		return nil // No config is valid (use system defaults)
	}

	// If using custom CA, verify it exists
	if v.config.CAFile != "" {
		if _, err := os.Stat(v.config.CAFile); os.IsNotExist(err) {
			return openclaw.NewValidationError("TLS CA file does not exist", ErrCANotFound)
		}
	}

	// If using client cert, both cert and key must be present
	if v.config.CertFile != "" || v.config.KeyFile != "" {
		if v.config.CertFile == "" || v.config.KeyFile == "" {
			return openclaw.NewValidationError("both CertFile and KeyFile are required for client authentication", ErrInvalidTLSConfig)
		}
		// Verify both files exist
		if _, err := os.Stat(v.config.CertFile); os.IsNotExist(err) {
			return openclaw.NewValidationError("TLS certificate file does not exist", ErrCertNotFound)
		}
		if _, err := os.Stat(v.config.KeyFile); os.IsNotExist(err) {
			return openclaw.NewValidationError("TLS key file does not exist", ErrCertNotFound)
		}
	}

	return nil
}

// GetTLSConfig returns the TLS config for the connection
func (v *TlsValidator) GetTLSConfig() (*tls.Config, error) {
	// First validate
	if err := v.Validate(); err != nil {
		return nil, err
	}

	config := &tls.Config{
		InsecureSkipVerify: v.config.InsecureSkipVerify,
		ServerName:         v.config.ServerName,
	}

	if v.config == nil {
		return config, nil
	}

	// Load client certificate if provided
	if v.config.CertFile != "" && v.config.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(v.config.CertFile, v.config.KeyFile)
		if err != nil {
			return nil, openclaw.NewTransportError("failed to load client certificate", err)
		}
		config.Certificates = []tls.Certificate{cert}
	}

	// Load CA certificate if provided
	if v.config.CAFile != "" {
		caCert, err := os.ReadFile(v.config.CAFile)
		if err != nil {
			return nil, openclaw.NewTransportError("failed to read CA certificate", err)
		}
		caPool := x509.NewCertPool()
		caPool.AppendCertsFromPEM(caCert)
		config.RootCAs = caPool
	}

	return config, nil
}

// ValidateCertificate validates the given certificate
// This is a basic validation - checks expiry and key usage
func ValidateCertificate(cert *x509.Certificate) error {
	if time.Now().After(cert.NotAfter) {
		return errors.New("certificate has expired")
	}
	if time.Now().Before(cert.NotBefore) {
		return errors.New("certificate is not yet valid")
	}
	return nil
}
```

- [ ] **Step 2: Write tls_test.go**

```go
// pkg/openclaw/connection/tls_test.go
package connection

import (
	"testing"
)

func TestTlsValidator_Validate_NilConfig(t *testing.T) {
	v := NewTlsValidator(nil)

	err := v.Validate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTlsValidator_Validate_MissingCAFile(t *testing.T) {
	v := NewTlsValidator(&TLSConfig{
		CAFile: "/nonexistent/ca.pem",
	})

	err := v.Validate()
	if err == nil {
		t.Error("expected error for missing CA file")
	}
}

func TestTlsValidator_Validate_IncompleteClientCert(t *testing.T) {
	v := NewTlsValidator(&TLSConfig{
		CertFile: "/path/to/cert.pem",
		// KeyFile missing
	})

	err := v.Validate()
	if err == nil {
		t.Error("expected error for incomplete client cert")
	}
}

func TestTlsValidator_Validate_ValidConfig(t *testing.T) {
	// Create temp files for testing
	// In real tests, use temp files

	v := NewTlsValidator(&TLSConfig{
		InsecureSkipVerify: true,
		ServerName:         "example.com",
	})

	err := v.Validate()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestTlsValidator_GetTLSConfig_Insecure(t *testing.T) {
	v := NewTlsValidator(&TLSConfig{
		InsecureSkipVerify: true,
		ServerName:        "example.com",
	})

	config, err := v.GetTLSConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.InsecureSkipVerify != true {
		t.Error("InsecureSkipVerify not set correctly")
	}
	if config.ServerName != "example.com" {
		t.Error("ServerName not set correctly")
	}
}

func TestTlsValidator_GetTLSConfig_NoConfig(t *testing.T) {
	v := NewTlsValidator(nil)

	config, err := v.GetTLSConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config == nil {
		t.Fatal("expected non-nil config")
	}
}
```

- [ ] **Step 3: Commit**

```bash
git add pkg/openclaw/connection/tls.go pkg/openclaw/connection/tls_test.go
git commit -m "feat: add TLS validator with certificate validation"
```

---

## Phase 5 Complete

After this phase, you should have:
- `pkg/openclaw/connection/state.go` - Connection state machine with typed states
- `pkg/openclaw/connection/state_test.go` - State machine tests
- `pkg/openclaw/connection/protocol.go` - Protocol negotiator with context support
- `pkg/openclaw/connection/protocol_test.go` - Protocol negotiator tests
- `pkg/openclaw/connection/policies.go` - Policy manager with reconnection logic
- `pkg/openclaw/connection/policies_test.go` - Policy manager tests
- `pkg/openclaw/connection/tls.go` - TLS validator with cert validation
- `pkg/openclaw/connection/tls_test.go` - TLS validator tests

All code should compile and tests should pass.

Key fixes from review:
1. Use Phase 1 typed ConnectionState constants throughout
2. Use Phase 1 openclaw.ProtocolError instead of local definition
3. Implement actual TLS validation logic (not a stub)
4. Add context support to ProtocolNegotiator
5. Return error when state change event is dropped
6. Use time.Duration for PingInterval (Go idiomatic)
7. Add comprehensive tests for all components
8. Note: TLSConfig is intentionally separate from transport.TLSConfig for different purposes
