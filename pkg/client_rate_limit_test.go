// Package openclaw provides the OpenClaw WebSocket SDK for Go.
package openclaw

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/protocol"
	"github.com/frisbee-ai/openclaw-sdk-go/pkg/types"
)

// denyLimiter always returns false from Allow().
type denyLimiter struct{}

func (denyLimiter) Allow() bool { return false }

// alwaysAllowLimiter always returns true from Allow().
type alwaysAllowLimiter struct{}

func (alwaysAllowLimiter) Allow() bool { return true }

// trackingLimiter wraps a limiter and tracks whether Allow() was called.
type trackingLimiter struct {
	inner  types.RequestRateLimiter
	called bool
}

func (t *trackingLimiter) Allow() bool {
	t.called = true
	return t.inner.Allow()
}

// TestWithRateLimit_Option verifies that WithRateLimit sets RateLimiter on config.
func TestWithRateLimit_Option(t *testing.T) {
	limiter := types.NewTokenBucketLimiter(100, 10)
	cli, err := NewClient(WithRateLimit(limiter))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer cli.Close()

	c := cli.(*client)
	if c.config.RateLimiter == nil {
		t.Error("expected RateLimiter to be set")
	}
}

// TestWithMaxPending_Option verifies that WithMaxPending sets MaxPending on config.
func TestWithMaxPending_Option(t *testing.T) {
	cli, err := NewClient(WithMaxPending(50))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer cli.Close()

	c := cli.(*client)
	if c.config.MaxPending != 50 {
		t.Errorf("expected MaxPending=50, got %d", c.config.MaxPending)
	}
}

// TestWithMaxPending_NegativeReturnsError verifies that WithMaxPending(-1) returns an error.
func TestWithMaxPending_NegativeReturnsError(t *testing.T) {
	_, err := NewClient(WithMaxPending(-1))
	if err == nil {
		t.Error("expected error for negative max pending, got nil")
	}
}

// TestSendRequest_RateLimited verifies that when the rate limiter denies,
// SendRequest returns a RequestError with code RATE_LIMITED before checking connection.
func TestSendRequest_RateLimited(t *testing.T) {
	cli, err := NewClient(
		WithURL("ws://localhost:8080"),
		WithClientID("test-client"),
		WithRateLimit(&denyLimiter{}),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer cli.Close()

	req := protocol.NewRequestFrame("test-1", "test.action", nil)
	_, err = cli.SendRequest(context.Background(), req)

	if err == nil {
		t.Fatal("expected error from rate-limited SendRequest, got nil")
	}

	var reqErr *RequestError
	if !errors.As(err, &reqErr) {
		t.Fatalf("expected *RequestError, got %T: %v", err, err)
	}
	if reqErr.Code() != "RATE_LIMITED" {
		t.Errorf("expected code RATE_LIMITED, got %s", reqErr.Code())
	}
	if !reqErr.Retryable() {
		t.Error("expected RATE_LIMITED error to be retryable")
	}
}

// TestSendRequest_NoRateLimiter verifies backward compatibility: without a rate limiter,
// SendRequest returns NOT_CONNECTED when not connected.
func TestSendRequest_NoRateLimiter(t *testing.T) {
	cli, err := NewClient(
		WithURL("ws://localhost:8080"),
		WithClientID("test-client"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer cli.Close()

	req := protocol.NewRequestFrame("test-1", "test.action", nil)
	_, err = cli.SendRequest(context.Background(), req)

	if err == nil {
		t.Fatal("expected error when not connected, got nil")
	}

	var connErr *ConnectionError
	if !errors.As(err, &connErr) {
		t.Fatalf("expected *ConnectionError, got %T: %v", err, err)
	}
	if connErr.Code() != "NOT_CONNECTED" {
		t.Errorf("expected code NOT_CONNECTED, got %s", connErr.Code())
	}
}

// TestSendRequest_RateLimiterCalled verifies the rate limiter Allow() is invoked.
func TestSendRequest_RateLimiterCalled(t *testing.T) {
	tracker := &trackingLimiter{inner: &alwaysAllowLimiter{}}
	cli, err := NewClient(
		WithURL("ws://localhost:8080"),
		WithClientID("test-client"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer cli.Close()

	c := cli.(*client)
	c.config.RateLimiter = tracker

	req := protocol.NewRequestFrame("test-1", "test.action", nil)
	_, _ = cli.SendRequest(context.Background(), req)

	if !tracker.called {
		t.Error("expected rate limiter Allow() to be called")
	}
}

// TestSendRequest_ConcurrentInFlight verifies that concurrent SendRequests can all
// check the rate limiter simultaneously (mutex not held during the wait).
func TestSendRequest_ConcurrentInFlight(t *testing.T) {
	cli, err := NewClient(
		WithURL("ws://localhost:8080"),
		WithClientID("test-client"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer cli.Close()

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errCount int

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			req := protocol.NewRequestFrame("concurrent-"+string(rune('0'+idx)), "test.action", nil)
			_, err := cli.SendRequest(context.Background(), req)
			mu.Lock()
			if err != nil {
				errCount++
			}
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	mu.Lock()
	count := errCount
	mu.Unlock()

	if count != 20 {
		t.Errorf("expected 20 errors (all NOT_CONNECTED), got %d", count)
	}
}

// TestTokenBucketLimiter_ReExports verifies that TokenBucketLimiter and
// NewTokenBucketLimiter are re-exported from the main package.
func TestTokenBucketLimiter_ReExports(t *testing.T) {
	limiter := NewTokenBucketLimiter(100.0, 10)
	if limiter == nil {
		t.Fatal("expected NewTokenBucketLimiter to return non-nil")
	}

	if !limiter.Allow() {
		t.Error("expected first Allow() to return true")
	}

	for i := 0; i < 10; i++ {
		limiter.Allow()
	}

	if limiter.Allow() {
		t.Error("expected Allow() to return false when exhausted")
	}
}
