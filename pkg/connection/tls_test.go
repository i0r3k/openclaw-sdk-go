// pkg/openclaw/connection/tls_test.go
package connection

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Helper function to generate temporary certificate files
func generateTestCerts(t *testing.T) (certFile, keyFile, caFile string, cleanup func()) {
	t.Helper()

	// Generate CA certificate
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate CA key: %v", err)
	}

	caTemplate := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test CA"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	caCertDER, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("failed to create CA certificate: %v", err)
	}

	caFile = filepath.Join(t.TempDir(), "ca.pem")
	if err := os.WriteFile(caFile, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caCertDER}), 0644); err != nil {
		t.Fatalf("failed to write CA file: %v", err)
	}

	// Generate client certificate
	clientKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate client key: %v", err)
	}

	clientKeyBytes, err := x509.MarshalECPrivateKey(clientKey)
	if err != nil {
		t.Fatalf("failed to marshal client key: %v", err)
	}

	keyFile = filepath.Join(t.TempDir(), "client.key")
	if err := os.WriteFile(keyFile, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: clientKeyBytes}), 0600); err != nil {
		t.Fatalf("failed to write key file: %v", err)
	}

	clientTemplate := x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Organization: []string{"Test Client"},
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	clientCertDER, err := x509.CreateCertificate(rand.Reader, &clientTemplate, &caTemplate, &clientKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("failed to create client certificate: %v", err)
	}

	certFile = filepath.Join(t.TempDir(), "client.pem")
	if err := os.WriteFile(certFile, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: clientCertDER}), 0644); err != nil {
		t.Fatalf("failed to write cert file: %v", err)
	}

	return certFile, keyFile, caFile, func() {
		os.Remove(caFile)
		os.Remove(keyFile)
		os.Remove(certFile)
	}
}

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
		ServerName:         "example.com",
	})

	config, err := v.GetTLSConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !config.InsecureSkipVerify {
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

func TestTlsValidator_GetTLSConfig_WithClientCert(t *testing.T) {
	certFile, keyFile, _, cleanup := generateTestCerts(t)
	defer cleanup()

	v := NewTlsValidator(&TLSConfig{
		CertFile:   certFile,
		KeyFile:    keyFile,
		ServerName: "test.example.com",
	})

	config, err := v.GetTLSConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(config.Certificates) != 1 {
		t.Fatalf("expected 1 certificate, got %d", len(config.Certificates))
	}

	cert := config.Certificates[0]
	if len(cert.Certificate) == 0 {
		t.Error("expected non-empty certificate")
	}
}

func TestTlsValidator_GetTLSConfig_WithCA(t *testing.T) {
	_, _, caFile, cleanup := generateTestCerts(t)
	defer cleanup()

	v := NewTlsValidator(&TLSConfig{
		CAFile:     caFile,
		ServerName: "test.example.com",
	})

	config, err := v.GetTLSConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if config.RootCAs == nil {
		t.Fatal("expected non-nil RootCAs")
	}
}

func TestTlsValidator_GetTLSConfig_WithClientCertAndCA(t *testing.T) {
	certFile, keyFile, caFile, cleanup := generateTestCerts(t)
	defer cleanup()

	v := NewTlsValidator(&TLSConfig{
		CertFile:   certFile,
		KeyFile:    keyFile,
		CAFile:     caFile,
		ServerName: "test.example.com",
	})

	config, err := v.GetTLSConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(config.Certificates) != 1 {
		t.Fatalf("expected 1 certificate, got %d", len(config.Certificates))
	}

	if config.RootCAs == nil {
		t.Fatal("expected non-nil RootCAs")
	}

	if config.ServerName != "test.example.com" {
		t.Errorf("ServerName = %s, want test.example.com", config.ServerName)
	}
}

func TestTlsValidator_GetTLSConfig_InvalidCertFile(t *testing.T) {
	v := NewTlsValidator(&TLSConfig{
		CertFile: "/nonexistent/cert.pem",
		KeyFile:  "/nonexistent/key.pem",
	})

	_, err := v.GetTLSConfig()
	if err == nil {
		t.Error("expected error for invalid cert file")
	}
}

func TestTlsValidator_GetTLSConfig_InvalidCAFile(t *testing.T) {
	v := NewTlsValidator(&TLSConfig{
		CAFile: "/nonexistent/ca.pem",
	})

	_, err := v.GetTLSConfig()
	if err == nil {
		t.Error("expected error for invalid CA file")
	}
}

func TestTlsValidator_GetTLSConfig_IncompleteClientCert(t *testing.T) {
	certFile, _, _, cleanup := generateTestCerts(t)
	defer cleanup()

	v := NewTlsValidator(&TLSConfig{
		CertFile: certFile,
		KeyFile:  "/nonexistent/key.pem",
	})

	_, err := v.GetTLSConfig()
	if err == nil {
		t.Error("expected error for incomplete client cert")
	}
}
