// auth/handler_test.go
package auth

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestStaticAuthHandler(t *testing.T) {
	creds := map[string]string{
		"username": "testuser",
		"password": "testpass",
	}
	handler, err := NewStaticAuthHandler(creds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	provider, err := handler.Authenticate(context.Background())
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

func TestStaticAuthHandler_NilCredentials(t *testing.T) {
	_, err := NewStaticAuthHandler(nil)
	if err == nil {
		t.Error("expected error for nil credentials")
	}
}

func TestStaticAuthHandler_EmptyCredentials(t *testing.T) {
	_, err := NewStaticAuthHandler(map[string]string{})
	if err == nil {
		t.Error("expected error for empty credentials")
	}
}

func TestStaticAuthHandler_ContextCancellation(t *testing.T) {
	creds := map[string]string{
		"username": "testuser",
		"password": "testpass",
	}
	handler, err := NewStaticAuthHandler(creds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = handler.Authenticate(ctx)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

func TestStaticAuthHandler_ContextTimeout(t *testing.T) {
	creds := map[string]string{
		"username": "testuser",
		"password": "testpass",
	}
	handler, err := NewStaticAuthHandler(creds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Wait for context to expire
	time.Sleep(100 * time.Millisecond)

	_, err = handler.Authenticate(ctx)
	if err == nil {
		t.Error("expected error for timed out context")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded error, got %v", err)
	}
}

// Compile-time check: StaticAuthHandler implements AuthHandler
var _ AuthHandler = (*StaticAuthHandler)(nil)
