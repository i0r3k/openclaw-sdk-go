// auth/handler.go
package auth

import (
	"context"
	"errors"
)

// ErrNoCredentials is returned when no credentials are provided
var ErrNoCredentials = errors.New("no credentials provided")

// AuthHandler handles authentication
type AuthHandler interface {
	// Authenticate performs authentication and returns credentials
	Authenticate(ctx context.Context) (CredentialsProvider, error)
}

// StaticAuthHandler is a simple auth handler that returns static credentials
type StaticAuthHandler struct {
	credentials map[string]string
}

// NewStaticAuthHandler creates a new static auth handler
// Returns error if credentials is nil or empty
func NewStaticAuthHandler(credentials map[string]string) (*StaticAuthHandler, error) {
	if credentials == nil {
		return nil, ErrNoCredentials
	}
	if len(credentials) == 0 {
		return nil, ErrNoCredentials
	}
	return &StaticAuthHandler{credentials: credentials}, nil
}

func (h *StaticAuthHandler) Authenticate(ctx context.Context) (CredentialsProvider, error) {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	return NewStaticCredentialsProvider(h.credentials)
}
