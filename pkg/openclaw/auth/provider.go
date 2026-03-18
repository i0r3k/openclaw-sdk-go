// auth/provider.go
package auth

import "errors"

// CredentialsProvider provides credentials for authentication
type CredentialsProvider interface {
	// GetCredentials returns credentials map
	GetCredentials() (map[string]string, error)
}

// StaticCredentialsProvider provides static credentials
type StaticCredentialsProvider struct {
	credentials map[string]string
}

// NewStaticCredentialsProvider creates a new static credentials provider
// Returns error if credentials is nil or empty
func NewStaticCredentialsProvider(credentials map[string]string) (*StaticCredentialsProvider, error) {
	if credentials == nil {
		return nil, errors.New("credentials cannot be nil")
	}
	if len(credentials) == 0 {
		return nil, errors.New("credentials cannot be empty")
	}
	return &StaticCredentialsProvider{credentials: credentials}, nil
}

func (p *StaticCredentialsProvider) GetCredentials() (map[string]string, error) {
	return p.credentials, nil
}
