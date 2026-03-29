// Package managers provides benchmarks for manager components.
package managers

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/types"
)

// BenchmarkEventManagerEmit measures EventManager.Emit() channel dispatch overhead.
func BenchmarkEventManagerEmit(b *testing.B) {
	ctx := context.Background()
	em := NewEventManager(ctx, 100, 100*time.Millisecond)
	em.Start()
	defer em.Close()

	// Pre-subscribe a handler to consume events (prevents channel blocking)
	var emitCount atomic.Uint64
	handler := func(event types.Event) {
		emitCount.Add(1)
	}
	em.Subscribe(types.EventConnect, handler)
	em.Subscribe(types.EventTick, handler)
	em.Subscribe(types.EventMessage, handler)

	// Create representative events
	connectEvent := types.Event{
		Type:      types.EventConnect,
		Priority:  types.EventPriorityMedium,
		Timestamp: time.Now(),
		Payload:   map[string]string{"server": "us-east"},
	}
	tickEvent := types.Event{
		Type:      types.EventTick,
		Priority:  types.EventPriorityMedium,
		Timestamp: time.Now(),
		Payload:   map[string]int64{"ts": time.Now().Unix()},
	}
	messageEvent := types.Event{
		Type:      types.EventMessage,
		Priority:  types.EventPriorityLow,
		Timestamp: time.Now(),
		Payload:   map[string]string{"msg": "hello"},
	}

	// Scaling pattern per D-17
	sizes := []int{100, 1000, 10000}
	for _, size := range sizes {
		b.Run(string(rune('0'+size/1000)), func(b *testing.B) {
			before := time.Now()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				em.Emit(connectEvent)
				em.Emit(tickEvent)
				em.Emit(messageEvent)
			}
			elapsed := time.Since(before)
			totalEmits := int64(b.N) * 3 // 3 events per iteration
			b.ReportMetric(float64(elapsed.Nanoseconds())/float64(totalEmits), "channel_overhead_ns")
		})
	}
}

// BenchmarkEventManagerEmitHighPriority measures high-priority event emit performance.
func BenchmarkEventManagerEmitHighPriority(b *testing.B) {
	ctx := context.Background()
	em := NewEventManager(ctx, 100, 100*time.Millisecond)
	em.Start()
	defer em.Close()

	var emitCount atomic.Uint64
	handler := func(event types.Event) {
		emitCount.Add(1)
	}
	em.Subscribe(types.EventError, handler)
	em.Subscribe(types.EventDisconnect, handler)

	errorEvent := types.Event{
		Type:      types.EventError,
		Priority:  types.EventPriorityHigh,
		Timestamp: time.Now(),
		Payload:   map[string]string{"code": "ERR", "message": "test error"},
	}

	sizes := []int{100, 1000, 10000}
	for _, size := range sizes {
		b.Run(string(rune('0'+size/1000)), func(b *testing.B) {
			before := time.Now()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				em.Emit(errorEvent)
			}
			elapsed := time.Since(before)
			b.ReportMetric(float64(elapsed.Nanoseconds())/float64(b.N), "channel_overhead_ns")
		})
	}
}

// BenchmarkEventManagerSubscribe measures handler subscription overhead.
func BenchmarkEventManagerSubscribe(b *testing.B) {
	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		em := NewEventManager(ctx, 100, 100*time.Millisecond)
		handler := func(event types.Event) {}
		em.Subscribe(types.EventTick, handler)
	}
}
