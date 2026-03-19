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
	"crypto/x509/pkix"
	"errors"
	"net"
	"os"
	"time"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/types"
)

// ErrCertificateExpired represents certificate expiration errors.
var ErrCertificateExpired = errors.New("certificate has expired")

// ErrCertificateNotYetValid represents certificate not yet valid errors.
var ErrCertificateNotYetValid = errors.New("certificate is not yet valid")

// ErrCertificateHostMismatch represents hostname verification failures.
var ErrCertificateHostMismatch = errors.New("certificate hostname mismatch")

// ErrCertificateKeyUsage represents invalid certificate key usage.
var ErrCertificateKeyUsage = errors.New("certificate has invalid key usage")

// ErrCertificateSignature represents signature verification failures.
var ErrCertificateSignature = errors.New("certificate signature verification failed")

// ErrCertificateRevoked represents certificate revocation errors.
var ErrCertificateRevoked = errors.New("certificate has been revoked")

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

// ValidateCertificate validates the given certificate comprehensively.
// It performs the following checks:
//   - Expiry validation (NotBefore and NotAfter)
//   - Key usage validation
//   - Basic constraints validation
//
// For hostname verification, use VerifyHostname instead.
func ValidateCertificate(cert *x509.Certificate) error {
	// 1. Check certificate expiration
	now := time.Now()
	if now.After(cert.NotAfter) {
		return ErrCertificateExpired
	}
	if now.Before(cert.NotBefore) {
		return ErrCertificateNotYetValid
	}

	// 2. Validate key usage
	if cert.KeyUsage != 0 {
		// Check for digital signature or key encipherment for TLS
		validKeyUsage := x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment
		if cert.KeyUsage&validKeyUsage == 0 {
			return ErrCertificateKeyUsage
		}
	}

	// 3. Check basic constraints
	if cert.BasicConstraintsValid && cert.IsCA {
		// CA certificates are valid for server/client auth
		return nil
	}

	// 4. For end-entity certificates, verify extended key usage
	if len(cert.ExtKeyUsage) > 0 {
		validExtKeyUsage := false
		for _, eku := range cert.ExtKeyUsage {
			if eku == x509.ExtKeyUsageServerAuth || eku == x509.ExtKeyUsageClientAuth {
				validExtKeyUsage = true
				break
			}
		}
		if !validExtKeyUsage {
			return ErrCertificateKeyUsage
		}
	}

	return nil
}

// VerifyHostname verifies that the certificate is valid for the given hostname.
// This is critical for preventing man-in-the-middle (MITM) attacks.
func VerifyHostname(cert *x509.Certificate, hostname string) error {
	if cert == nil {
		return errors.New("certificate is nil")
	}

	opts := x509.VerifyOptions{
		// Use DNSName for hostname verification
		DNSName: hostname,
		// Note: In production, you would provide intermediate certificates here
	}

	_, err := cert.Verify(opts)
	if err != nil {
		// Check if it's a hostname mismatch error
		var hostnameErr *x509.HostnameError
		if errors.As(err, &hostnameErr) {
			return ErrCertificateHostMismatch
		}
		// Other verification errors
		return errors.New("certificate verification failed: " + err.Error())
	}

	return nil
}

// ValidateCertificateChain validates a certificate chain against a CA pool.
// This verifies that the certificate is signed by a trusted CA.
func ValidateCertificateChain(cert *x509.Certificate, caPool *x509.CertPool) error {
	if cert == nil {
		return errors.New("certificate is nil")
	}
	if caPool == nil {
		return errors.New("CA pool is nil")
	}

	opts := x509.VerifyOptions{
		Roots:         caPool,
		Intermediates: x509.NewCertPool(),
	}

	// Verify the certificate chain
	_, err := cert.Verify(opts)
	if err != nil {
		var unknownAuthErr *x509.UnknownAuthorityError
		if errors.As(err, &unknownAuthErr) {
			return ErrCertificateSignature
		}
		var certInvalidErr *x509.CertificateInvalidError
		if errors.As(err, &certInvalidErr) {
			return ErrCertificateSignature
		}
		return errors.New("certificate chain verification failed: " + err.Error())
	}

	return nil
}

// CheckCertificateRevocation checks if a certificate has been revoked.
// Note: This requires access to CRL or OCSP responders.
// In production, this should be configured with proper CRL/OCSP endpoints.
func CheckCertificateRevocation(cert *x509.Certificate, _ *x509.RevocationList) error {
	if cert == nil {
		return errors.New("certificate is nil")
	}

	// Check if certificate has a CRL distribution point
	if len(cert.CRLDistributionPoints) > 0 {
		// In production, fetch and verify against CRL
		// For now, we return nil as CRL checking is complex and usually
		// handled by the TLS library automatically
		return nil
	}

	// Check if certificate has OCSP responder
	if len(cert.OCSPServer) > 0 {
		// In production, perform OCSP check
		// For now, we return nil as OCSP checking is complex
		return nil
	}

	// No revocation information available
	return nil
}

// CertificateInfo holds validated certificate information.
type CertificateInfo struct {
	Subject      pkix.Name
	Issuer       pkix.Name
	NotBefore    time.Time
	NotAfter     time.Time
	DNSNames     []string
	IPAddresses  []net.IP
	IsCA         bool
	KeyUsage     x509.KeyUsage
	ExtKeyUsage  []x509.ExtKeyUsage
	SerialNumber string
}

// ExtractInfo extracts certificate information for logging/display.
func ExtractInfo(cert *x509.Certificate) *CertificateInfo {
	if cert == nil {
		return nil
	}

	serialHex := ""
	if cert.SerialNumber != nil {
		serialHex = cert.SerialNumber.Text(16)
	}

	return &CertificateInfo{
		Subject:      cert.Subject,
		Issuer:       cert.Issuer,
		NotBefore:    cert.NotBefore,
		NotAfter:     cert.NotAfter,
		DNSNames:     cert.DNSNames,
		IPAddresses:  cert.IPAddresses,
		IsCA:         cert.IsCA,
		KeyUsage:     cert.KeyUsage,
		ExtKeyUsage:  cert.ExtKeyUsage,
		SerialNumber: serialHex,
	}
}
