package events

import (
	"sync"
	"testing"
)

func TestGapDetector_Basic(t *testing.T) {
	detector := NewGapDetector()
	gapDetected := false

	detector.SetOnGap(func(gaps []GapInfo) {
		gapDetected = true
		if len(gaps) != 1 {
			t.Errorf("expected 1 gap, got %d", len(gaps))
		}
		if gaps[0].Expected != 2 || gaps[0].Received != 4 {
			t.Errorf("expected gap 2-4, got %d-%d", gaps[0].Expected, gaps[0].Received)
		}
	})

	detector.RecordSequence(1)
	detector.RecordSequence(4) // Should detect gap 2-3

	if !gapDetected {
		t.Error("expected gap to be detected")
	}
}

func TestGapDetector_NoGap(t *testing.T) {
	detector := NewGapDetector()

	detector.RecordSequence(1)
	detector.RecordSequence(2)
	detector.RecordSequence(3)

	gaps := detector.GetGaps()
	if len(gaps) != 0 {
		t.Errorf("expected no gaps, got %d", len(gaps))
	}
}

func TestGapDetector_DuplicateSequence(t *testing.T) {
	detector := NewGapDetector()

	detector.RecordSequence(1)
	detector.RecordSequence(1) // Duplicate
	detector.RecordSequence(2)

	gaps := detector.GetGaps()
	if len(gaps) != 0 {
		t.Errorf("expected no gaps, got %d", len(gaps))
	}
}

func TestGapDetector_GapsReturnsCopy(t *testing.T) {
	detector := NewGapDetector()

	detector.RecordSequence(1)
	detector.RecordSequence(4)

	gaps1 := detector.GetGaps()
	gaps1[0].Expected = 999 // Modify copy

	gaps2 := detector.GetGaps()
	if gaps2[0].Expected == 999 {
		t.Error("GetGaps() should return a copy, not original")
	}
}

func TestGapDetector_GapCount(t *testing.T) {
	detector := NewGapDetector()

	detector.RecordSequence(1)
	detector.RecordSequence(4)
	detector.RecordSequence(10)

	if detector.GapCount() != 2 {
		t.Errorf("expected 2 gaps, got %d", detector.GapCount())
	}
}

func TestGapDetector_Reset(t *testing.T) {
	detector := NewGapDetector()

	detector.RecordSequence(1)
	detector.RecordSequence(4)

	detector.Reset()

	if detector.GapCount() != 0 {
		t.Error("expected 0 gaps after reset")
	}
	if detector.GetLastSequence() != 0 {
		t.Error("expected 0 lastSequence after reset")
	}
}

func TestGapDetector_Concurrent(t *testing.T) {
	detector := NewGapDetector()

	var wg sync.WaitGroup
	wg.Add(2)

	// Concurrent record from multiple goroutines
	go func() {
		defer wg.Done()
		for i := uint64(0); i < 1000; i++ {
			detector.RecordSequence(i * 2) // Even numbers
		}
	}()

	go func() {
		defer wg.Done()
		for i := uint64(0); i < 1000; i++ {
			detector.RecordSequence(i*2 + 1) // Odd numbers
		}
	}()

	wg.Wait()

	// Should have gaps when interleaved
	gaps := detector.GetGaps()
	_ = gaps // Verify no panic
}

func TestGapDetector_ConcurrentCallbacks(t *testing.T) {
	detector := NewGapDetector()

	var wg sync.WaitGroup
	wg.Add(2)

	// Set callback while recording
	detector.SetOnGap(func(gaps []GapInfo) {})

	go func() {
		defer wg.Done()
		for i := uint64(0); i < 100; i++ {
			detector.RecordSequence(i)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			detector.SetOnGap(func(gaps []GapInfo) {})
		}
	}()

	wg.Wait()
}

func TestGapDetector_HasGap(t *testing.T) {
	detector := NewGapDetector()

	if detector.HasGap() {
		t.Error("expected no gap initially")
	}

	detector.RecordSequence(1)
	detector.RecordSequence(5)

	if !detector.HasGap() {
		t.Error("expected gap after skipping sequence")
	}
}

func TestGapDetector_GetLastSequence(t *testing.T) {
	detector := NewGapDetector()

	if detector.GetLastSequence() != 0 {
		t.Error("expected 0 last sequence initially")
	}

	detector.RecordSequence(5)

	if detector.GetLastSequence() != 5 {
		t.Errorf("expected 5, got %d", detector.GetLastSequence())
	}
}
