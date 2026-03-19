// Package auth provides authentication types and handlers for OpenClaw SDK.
//
// This package provides:
//   - CredentialsProvider: Interface for credential sources
//   - StaticCredentialsProvider: Simple static credential implementation
//   - AuthHandler: Interface for authentication logic
//   - StaticAuthHandler: Simple static authentication implementation
package auth

import (
	"errors"
	"strings"
)

// ErrInvalidCredentials is returned when credentials format is invalid.
var ErrInvalidCredentials = errors.New("invalid credentials format")

// CredentialsProvider provides credentials for authentication.
// Implement this interface to provide custom credential sources.
type CredentialsProvider interface {
	// GetCredentials returns credentials map
	GetCredentials() (map[string]string, error)
}

// StaticCredentialsProvider provides static credentials.
// It stores credentials in memory and returns them on request.
type StaticCredentialsProvider struct {
	credentials map[string]string
}

// NewStaticCredentialsProvider creates a new static credentials provider.
// Returns error if credentials is nil, empty, or contains invalid format.
func NewStaticCredentialsProvider(credentials map[string]string) (*StaticCredentialsProvider, error) {
	if err := validateCredentials(credentials); err != nil {
		return nil, err
	}
	return &StaticCredentialsProvider{credentials: credentials}, nil
}

// validateCredentials checks that credentials are valid.
func validateCredentials(credentials map[string]string) error {
	if credentials == nil {
		return errors.New("credentials cannot be nil")
	}
	if len(credentials) == 0 {
		return errors.New("credentials cannot be empty")
	}
	// Validate credential format: check for required keys with non-empty values
	for key, value := range credentials {
		if strings.TrimSpace(value) == "" {
			return errors.New("credentials cannot be empty: " + key)
		}
	}
	return nil
}

// GetCredentials returns the stored credentials map.
func (p *StaticCredentialsProvider) GetCredentials() (map[string]string, error) {
	return p.credentials, nil
}
