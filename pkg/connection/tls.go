// Package connection provides connection management components for OpenClaw SDK.
//
// This package provides:
//   - ConnectionStateMachine: State machine for managing connection lifecycle
//   - PolicyManager: Connection policy configuration
//   - ProtocolNegotiator: Protocol version negotiation
//   - TLS validation: Certificate and configuration validation
package connection

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"
	"time"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/types"
)

// TlsValidator validates TLS certificates and configuration.
// It provides methods to validate TLS settings and build crypto/tls.Config.
type TlsValidator struct {
	config *TLSConfig // TLS configuration to validate
}

// TLSConfig holds TLS configuration for connection layer.
// Note: This is distinct from transport.TLSConfig which is for dial-time configuration.
// This version supports certificate loading and validation.
type TLSConfig struct {
	InsecureSkipVerify bool   // Skip server certificate verification (insecure)
	CertFile           string // Path to client certificate file
	KeyFile            string // Path to client key file
	CAFile             string // Path to CA certificate file
	ServerName         string // Server name for SNI
}

// ErrInvalidTLSConfig represents TLS configuration validation errors.
var ErrInvalidTLSConfig = errors.New("invalid TLS configuration")

// ErrCertNotFound is returned when certificate file is not found.
var ErrCertNotFound = errors.New("certificate file not found")

// ErrCANotFound is returned when CA file is not found.
var ErrCANotFound = errors.New("CA certificate file not found")

// NewTlsValidator creates a new TLS validator with the given configuration.
func NewTlsValidator(config *TLSConfig) *TlsValidator {
	return &TlsValidator{config: config}
}

// Validate validates the TLS configuration.
// Checks that required files exist and that certificate/key pairs are complete.
func (v *TlsValidator) Validate() error {
	if v.config == nil {
		return nil // No config is valid (use system defaults)
	}

	// If using custom CA, verify it exists
	if v.config.CAFile != "" {
		if _, err := os.Stat(v.config.CAFile); os.IsNotExist(err) {
			return types.NewValidationError("TLS CA file does not exist", ErrCANotFound)
		}
	}

	// If using client cert, both cert and key must be present
	if v.config.CertFile != "" || v.config.KeyFile != "" {
		if v.config.CertFile == "" || v.config.KeyFile == "" {
			return types.NewValidationError("both CertFile and KeyFile are required for client authentication", ErrInvalidTLSConfig)
		}
		// Verify both files exist
		if _, err := os.Stat(v.config.CertFile); os.IsNotExist(err) {
			return types.NewValidationError("TLS certificate file does not exist", ErrCertNotFound)
		}
		if _, err := os.Stat(v.config.KeyFile); os.IsNotExist(err) {
			return types.NewValidationError("TLS key file does not exist", ErrCertNotFound)
		}
	}

	return nil
}

// GetTLSConfig returns the TLS config for the connection.
// It validates the configuration first, then builds a crypto/tls.Config.
func (v *TlsValidator) GetTLSConfig() (*tls.Config, error) {
	// First validate
	if err := v.Validate(); err != nil {
		return nil, err
	}

	// Handle nil config case
	if v.config == nil {
		return &tls.Config{}, nil
	}

	config := &tls.Config{
		InsecureSkipVerify: v.config.InsecureSkipVerify,
		ServerName:         v.config.ServerName,
	}

	// Load client certificate if provided
	if v.config.CertFile != "" && v.config.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(v.config.CertFile, v.config.KeyFile)
		if err != nil {
			return nil, types.NewTransportError("failed to load client certificate", err)
		}
		config.Certificates = []tls.Certificate{cert}
	}

	// Load CA certificate if provided
	if v.config.CAFile != "" {
		caCert, err := os.ReadFile(v.config.CAFile)
		if err != nil {
			return nil, types.NewTransportError("failed to read CA certificate", err)
		}
		caPool := x509.NewCertPool()
		caPool.AppendCertsFromPEM(caCert)
		config.RootCAs = caPool
	}

	return config, nil
}

// ValidateCertificate validates the given certificate.
// This is a basic validation that checks expiry dates (NotBefore and NotAfter).
func ValidateCertificate(cert *x509.Certificate) error {
	if time.Now().After(cert.NotAfter) {
		return errors.New("certificate has expired")
	}
	if time.Now().Before(cert.NotBefore) {
		return errors.New("certificate is not yet valid")
	}
	return nil
}
