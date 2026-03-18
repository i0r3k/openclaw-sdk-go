package events

import (
	"sync"
	"testing"
)

func TestGapDetector_Basic(t *testing.T) {
	detector := NewGapDetector()
	gapDetected := false

	detector.SetOnGap(func(start, end uint64) {
		gapDetected = true
		if start != 2 || end != 3 {
			t.Errorf("expected gap 2-3, got %d-%d", start, end)
		}
	})

	detector.Record(1)
	detector.Record(4) // Should detect gap 2-3

	if !gapDetected {
		t.Error("expected gap to be detected")
	}
}

func TestGapDetector_NoGap(t *testing.T) {
	detector := NewGapDetector()

	detector.Record(1)
	detector.Record(2)
	detector.Record(3)

	gaps := detector.Gaps()
	if len(gaps) != 0 {
		t.Errorf("expected no gaps, got %d", len(gaps))
	}
}

func TestGapDetector_DuplicateSequence(t *testing.T) {
	detector := NewGapDetector()

	detector.Record(1)
	detector.Record(1) // Duplicate
	detector.Record(2)

	gaps := detector.Gaps()
	if len(gaps) != 0 {
		t.Errorf("expected no gaps, got %d", len(gaps))
	}
}

func TestGapDetector_GapsReturnsCopy(t *testing.T) {
	detector := NewGapDetector()

	detector.Record(1)
	detector.Record(4)

	gaps1 := detector.Gaps()
	gaps1[0].Start = 999 // Modify copy

	gaps2 := detector.Gaps()
	if gaps2[0].Start == 999 {
		t.Error("Gaps() should return a copy, not original")
	}
}

func TestGapDetector_GapCount(t *testing.T) {
	detector := NewGapDetector()

	detector.Record(1)
	detector.Record(4)
	detector.Record(10)

	if detector.GapCount() != 2 {
		t.Errorf("expected 2 gaps, got %d", detector.GapCount())
	}
}

func TestGapDetector_Reset(t *testing.T) {
	detector := NewGapDetector()

	detector.Record(1)
	detector.Record(4)

	detector.Reset()

	if detector.GapCount() != 0 {
		t.Error("expected 0 gaps after reset")
	}
	if detector.ExpectedSequence() != 0 {
		t.Error("expected 0 expectedSeq after reset")
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
			detector.Record(i * 2) // Even numbers
		}
	}()

	go func() {
		defer wg.Done()
		for i := uint64(0); i < 1000; i++ {
			detector.Record(i*2 + 1) // Odd numbers
		}
	}()

	wg.Wait()

	// Should have gaps when interleaved
	gaps := detector.Gaps()
	_ = gaps // Verify no panic
}

func TestGapDetector_ConcurrentCallbacks(t *testing.T) {
	detector := NewGapDetector()

	var wg sync.WaitGroup
	wg.Add(2)

	// Set callback while recording
	detector.SetOnGap(func(start, end uint64) {})

	go func() {
		defer wg.Done()
		for i := uint64(0); i < 100; i++ {
			detector.Record(i)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			detector.SetOnGap(func(start, end uint64) {})
		}
	}()

	wg.Wait()
}
