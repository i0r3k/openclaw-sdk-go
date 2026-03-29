// Package managers provides benchmarks for manager components.
package managers

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/protocol"
)

// BenchmarkRequestManagerSendRequest measures RequestManager.SendRequest() + HandleResponse() correlation overhead.
// Note: This benchmark uses a simple sequential pattern where HandleResponse is called after SendRequest registers.
func BenchmarkRequestManagerSendRequest(b *testing.B) {
	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		rm := NewRequestManager(ctx)

		sendFunc := func(req *protocol.RequestFrame) error {
			return nil
		}

		req := protocol.NewRequestFrame(
			"bench-req-001",
			"client.ping",
			[]byte(`{"key":"value"}`),
		)

		// Start SendRequest in background
		go func() {
			rm.SendRequest(ctx, req, sendFunc)
		}()

		// Ensure request is registered by sending response
		time.Sleep(10 * time.Millisecond)
		resp := protocol.NewResponseFrameSuccess("bench-req-001", []byte(`{"status":"ok"}`))
		rm.HandleResponse(resp)

		// Give time for cleanup
		time.Sleep(10 * time.Millisecond)
		rm.Close()
	}
}

// BenchmarkRequestManagerHandleResponse measures HandleResponse performance.
func BenchmarkRequestManagerHandleResponse(b *testing.B) {
	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		rm := NewRequestManager(ctx)

		req := protocol.NewRequestFrame("bench-resp", "test", nil)
		sendFunc := func(r *protocol.RequestFrame) error { return nil }

		go func() {
			rm.SendRequest(ctx, req, sendFunc)
		}()

		time.Sleep(10 * time.Millisecond)

		resp := protocol.NewResponseFrameSuccess("bench-resp", []byte(`{"ok":true}`))
		rm.HandleResponse(resp)

		time.Sleep(10 * time.Millisecond)
		rm.Close()
	}
}

// BenchmarkRequestManagerConcurrent measures concurrent request/response handling.
func BenchmarkRequestManagerConcurrent(b *testing.B) {
	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		rm := NewRequestManager(ctx)

		for j := 0; j < 10; j++ {
			go func(id int) {
				req := protocol.NewRequestFrame(
					string(rune('a'+id)),
					"test",
					nil,
				)
				sendFunc := func(r *protocol.RequestFrame) error { return nil }
				rm.SendRequest(ctx, req, sendFunc)
			}(j)
		}

		time.Sleep(10 * time.Millisecond)

		for j := 0; j < 10; j++ {
			resp := protocol.NewResponseFrameSuccess(string(rune('a'+j)), nil)
			rm.HandleResponse(resp)
		}

		time.Sleep(10 * time.Millisecond)
		rm.Close()
	}
}

// BenchmarkRequestManagerGoroutineCount measures goroutine count during request handling.
func BenchmarkRequestManagerGoroutineCount(b *testing.B) {
	ctx := context.Background()

	rm := NewRequestManager(ctx)
	defer rm.Close()

	req := protocol.NewRequestFrame("bench-req", "test", nil)
	sendFunc := func(r *protocol.RequestFrame) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	}

	go func() {
		rm.SendRequest(ctx, req, sendFunc)
	}()

	// Wait for SendRequest to block on respCh
	time.Sleep(time.Millisecond)

	// Report goroutine count while request is pending
	b.ReportMetric(float64(runtime.NumGoroutine()), "goroutine_count")

	resp := protocol.NewResponseFrameSuccess("bench-req", nil)
	rm.HandleResponse(resp)

	time.Sleep(time.Millisecond)
}
