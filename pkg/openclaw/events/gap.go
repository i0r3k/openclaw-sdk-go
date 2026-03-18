package events

import (
	"sync"
)

// GapDetector detects message gaps
type GapDetector struct {
	mu           sync.Mutex
	expectedSeq  uint64
	detectedGaps []Gap
	onGap        func(start, end uint64)
}

// Gap represents a detected gap
type Gap struct {
	Start uint64
	End   uint64
}

// NewGapDetector creates a new gap detector
func NewGapDetector() *GapDetector {
	return &GapDetector{
		detectedGaps: make([]Gap, 0),
	}
}

// SetOnGap sets the gap callback (thread-safe)
func (gd *GapDetector) SetOnGap(f func(start, end uint64)) {
	gd.mu.Lock()
	defer gd.mu.Unlock()
	gd.onGap = f
}

// Record records a message sequence number
func (gd *GapDetector) Record(seq uint64) {
	gd.mu.Lock()
	defer gd.mu.Unlock()

	// Handle first message
	if gd.expectedSeq == 0 {
		gd.expectedSeq = seq + 1
		return
	}

	// Detect gap (but not for duplicate sequences)
	if seq > gd.expectedSeq {
		gap := Gap{Start: gd.expectedSeq, End: seq - 1}
		gd.detectedGaps = append(gd.detectedGaps, gap)

		// Call callback OUTSIDE the lock to prevent deadlock
		onGap := gd.onGap
		if onGap != nil {
			gd.mu.Unlock()
			onGap(gap.Start, gap.End)
			gd.mu.Lock()
		}
	}

	// Advance expected (handle duplicates by not going backwards)
	if seq+1 > gd.expectedSeq {
		gd.expectedSeq = seq + 1
	}
}

// Gaps returns a copy of detected gaps (thread-safe)
func (gd *GapDetector) Gaps() []Gap {
	gd.mu.Lock()
	defer gd.mu.Unlock()

	// Return a copy to prevent external mutation
	result := make([]Gap, len(gd.detectedGaps))
	copy(result, gd.detectedGaps)
	return result
}

// GapCount returns the number of detected gaps
func (gd *GapDetector) GapCount() int {
	gd.mu.Lock()
	defer gd.mu.Unlock()
	return len(gd.detectedGaps)
}

// Reset resets the gap detector (thread-safe)
func (gd *GapDetector) Reset() {
	gd.mu.Lock()
	defer gd.mu.Unlock()
	gd.expectedSeq = 0
	gd.detectedGaps = nil
	gd.detectedGaps = make([]Gap, 0)
}

// ExpectedSequence returns the next expected sequence number
func (gd *GapDetector) ExpectedSequence() uint64 {
	gd.mu.Lock()
	defer gd.mu.Unlock()
	return gd.expectedSeq
}
