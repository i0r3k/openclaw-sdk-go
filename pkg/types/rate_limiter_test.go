package types

import (
	"sync"
	"testing"
	"time"
)

// TestTokenBucketLimiter_AllowWithinBurst verifies that TokenBucketLimiter allows
// exactly burst calls then denies subsequent ones.
func TestTokenBucketLimiter_AllowWithinBurst(t *testing.T) {
	limiter := NewTokenBucketLimiter(10, 5)

	// First 5 should all succeed (burst capacity)
	for i := 0; i < 5; i++ {
		if !limiter.Allow() {
			t.Errorf("Allow() = false on call %d, want true (within burst)", i+1)
		}
	}

	// 6th call should be denied
	if limiter.Allow() {
		t.Error("Allow() = true on 6th call, want false (burst exhausted)")
	}
}

// TestTokenBucketLimiter_TokenRefill verifies that tokens refill over time.
func TestTokenBucketLimiter_TokenRefill(t *testing.T) {
	// rate=1000 tokens/sec, burst=10
	limiter := NewTokenBucketLimiter(1000, 10)

	// Drain the bucket
	for i := 0; i < 10; i++ {
		if !limiter.Allow() {
			t.Errorf("Allow() = false on drain call %d, want true", i+1)
		}
	}

	// Verify bucket is empty
	if limiter.Allow() {
		t.Error("Allow() = true immediately after drain, want false")
	}

	// Wait for tokens to refill (15ms should give us ~15 tokens at 1000/sec)
	time.Sleep(15 * time.Millisecond)

	// Should be allowed again
	if !limiter.Allow() {
		t.Error("Allow() = false after refill wait, want true")
	}
}

// TestTokenBucketLimiter_BurstCap verifies the limiter never exceeds burst capacity.
func TestTokenBucketLimiter_BurstCap(t *testing.T) {
	// Low rate (100/sec) with burst=3 ensures refill can't outpace burst between rapid calls.
	// At 100/sec, ~0.01 tokens refill per millisecond. Even 0.001s between calls adds only 0.1 tokens.
	limiter := NewTokenBucketLimiter(100, 3)

	allowed := 0
	for i := 0; i < 100; i++ {
		if limiter.Allow() {
			allowed++
		}
	}

	if allowed > 3 {
		t.Errorf("Allowed %d calls, want max 3 (burst cap)", allowed)
	}
}

// TestTokenBucketLimiter_ConcurrentSafety verifies the limiter is safe for concurrent use.
func TestTokenBucketLimiter_ConcurrentSafety(t *testing.T) {
	limiter := NewTokenBucketLimiter(10000, 100)

	const goroutines = 50
	const callsPerGoroutine = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < callsPerGoroutine; j++ {
				limiter.Allow()
			}
		}()
	}

	wg.Wait()

	// No panic = test passes (race detector will catch races if any)
}

// TestTokenBucketLimiter_InterfaceCompliance verifies TokenBucketLimiter implements RequestRateLimiter.
func TestTokenBucketLimiter_InterfaceCompliance(t *testing.T) {
	var limiter RequestRateLimiter = NewTokenBucketLimiter(10, 5)
	if limiter == nil {
		t.Error("TokenBucketLimiter does not satisfy RequestRateLimiter interface")
	}
}
