// auth/provider_test.go
package auth

import (
	"testing"
)

func TestStaticCredentialsProvider(t *testing.T) {
	creds := map[string]string{
		"username": "testuser",
		"password": "testpass",
	}
	provider, err := NewStaticCredentialsProvider(creds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := provider.GetCredentials()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got["username"] != "testuser" {
		t.Errorf("expected 'testuser', got '%s'", got["username"])
	}
}

func TestStaticCredentialsProvider_Nil(t *testing.T) {
	_, err := NewStaticCredentialsProvider(nil)
	if err == nil {
		t.Error("expected error for nil credentials")
	}
}

func TestStaticCredentialsProvider_Empty(t *testing.T) {
	_, err := NewStaticCredentialsProvider(map[string]string{})
	if err == nil {
		t.Error("expected error for empty credentials")
	}
}

// Compile-time check: StaticCredentialsProvider implements CredentialsProvider
var _ CredentialsProvider = (*StaticCredentialsProvider)(nil)
