// Package auth provides authentication types and handlers for OpenClaw SDK.
//
// This package provides:
//   - CredentialsProvider: Interface for credential sources
//   - StaticCredentialsProvider: Simple static credential implementation
//   - AuthHandler: Interface for authentication logic
//   - StaticAuthHandler: Simple static authentication implementation
package auth

import (
	"context"
	"errors"
)

// ErrNoCredentials is returned when no credentials are provided.
var ErrNoCredentials = errors.New("no credentials provided")

// AuthHandler handles authentication.
// Implement this interface to provide custom authentication logic.
type AuthHandler interface {
	// Authenticate performs authentication and returns credentials.
	// The context can be used for cancellation and timeout.
	Authenticate(ctx context.Context) (CredentialsProvider, error)
}

// StaticAuthHandler is a simple auth handler that returns static credentials.
// It returns the same credentials on every call without any actual authentication.
type StaticAuthHandler struct {
	credentials map[string]string
}

// NewStaticAuthHandler creates a new static auth handler.
// Returns error if credentials is nil, empty, or contains empty values.
func NewStaticAuthHandler(credentials map[string]string) (*StaticAuthHandler, error) {
	if err := validateCredentials(credentials); err != nil {
		return nil, err
	}
	return &StaticAuthHandler{credentials: credentials}, nil
}

// Authenticate returns a static credentials provider.
// It checks for context cancellation before returning credentials.
func (h *StaticAuthHandler) Authenticate(ctx context.Context) (CredentialsProvider, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	return NewStaticCredentialsProvider(h.credentials)
}
