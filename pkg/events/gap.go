// Package events provides event handling utilities for the OpenClaw SDK.
//
// This package provides:
//   - TickMonitor: Connection heartbeat monitoring with timeout detection
//   - GapDetector: Message gap detection for ordered message streams
package events

import (
	"sync"
	"time"
)

// GapRecoveryMode defines how gap recovery is handled.
type GapRecoveryMode string

const (
	GapRecoveryModeReconnect GapRecoveryMode = "reconnect"
	GapRecoveryModeSnapshot  GapRecoveryMode = "snapshot"
	GapRecoveryModeSkip      GapRecoveryMode = "skip"
)

// GapRecoveryConfig contains configuration for gap recovery behavior.
type GapRecoveryConfig struct {
	Mode             GapRecoveryMode // Recovery mode
	OnGap            func([]GapInfo) // Callback when gaps are detected
	SnapshotEndpoint string          // Endpoint for snapshot recovery (optional)
}

// GapDetectorConfig contains configuration for the gap detector.
type GapDetectorConfig struct {
	Recovery GapRecoveryConfig // Recovery configuration
	MaxGaps  int               // Maximum gaps to track (default: 100)
}

// GapInfo represents a detected gap in the message sequence.
type GapInfo struct {
	Expected   uint64 // Expected sequence number
	Received   uint64 // Received sequence number
	DetectedAt int64  // Timestamp when gap was detected
}

// GapDetector detects message gaps in ordered message streams.
// It tracks sequence numbers and identifies when expected messages are missing.
type GapDetector struct {
	mu           sync.Mutex        // Mutex for thread-safety
	recovery     GapRecoveryConfig // Recovery configuration
	maxGaps      int               // Maximum gaps to track
	lastSequence uint64            // Last recorded sequence number
	gaps         []GapInfo         // List of detected gaps
}

// NewGapDetector creates a new gap detector with default configuration.
func NewGapDetector() *GapDetector {
	return NewGapDetectorWithConfig(GapDetectorConfig{
		Recovery: GapRecoveryConfig{
			Mode: GapRecoveryModeSkip,
		},
		MaxGaps: 100,
	})
}

// NewGapDetectorWithConfig creates a new gap detector with the given configuration.
func NewGapDetectorWithConfig(config GapDetectorConfig) *GapDetector {
	maxGaps := config.MaxGaps
	if maxGaps <= 0 {
		maxGaps = 100
	}
	return &GapDetector{
		recovery: config.Recovery,
		maxGaps:  maxGaps,
		gaps:     make([]GapInfo, 0),
	}
}

// SetOnGap sets the callback function to be called when a gap is detected.
func (gd *GapDetector) SetOnGap(f func([]GapInfo)) {
	gd.mu.Lock()
	defer gd.mu.Unlock()
	gd.recovery.OnGap = f
}

// RecordSequence records a message sequence number.
// It detects gaps when sequence numbers are skipped.
func (gd *GapDetector) RecordSequence(seq uint64) {
	// Deferred side-effects list — executed after all state updates
	var deferred []func()

	gd.mu.Lock()

	// Check for gap if we have a previous sequence
	if gd.lastSequence > 0 {
		expected := gd.lastSequence + 1

		// Only detect gaps for sequences after the last one
		// Duplicate or old sequences are ignored
		if seq > gd.lastSequence && seq > expected {
			gap := GapInfo{
				Expected:   expected,
				Received:   seq,
				DetectedAt: int64(getTimeNow()),
			}

			// State mutation FIRST
			gd.gaps = append(gd.gaps, gap)

			// Trim if exceeds max (use splice to avoid array copy)
			if len(gd.gaps) > gd.maxGaps {
				gd.gaps = gd.gaps[len(gd.gaps)-gd.maxGaps:]
			}

			// Capture side-effects (do not execute yet)
			if gd.recovery.OnGap != nil {
				deferred = append(deferred, func() {
					gd.recovery.OnGap(gd.gaps)
				})
			}
		}
	}

	// Update lastSequence BEFORE executing side-effects
	gd.lastSequence = seq

	gd.mu.Unlock()

	// Execute deferred side-effects — state is already consistent
	for _, op := range deferred {
		op()
	}
}

// HasGap returns true if gaps have been detected.
func (gd *GapDetector) HasGap() bool {
	gd.mu.Lock()
	defer gd.mu.Unlock()
	return len(gd.gaps) > 0
}

// GapCount returns the number of detected gaps.
func (gd *GapDetector) GapCount() int {
	gd.mu.Lock()
	defer gd.mu.Unlock()
	return len(gd.gaps)
}

// GetGaps returns a copy of detected gaps.
func (gd *GapDetector) GetGaps() []GapInfo {
	gd.mu.Lock()
	defer gd.mu.Unlock()
	result := make([]GapInfo, len(gd.gaps))
	copy(result, gd.gaps)
	return result
}

// GetLastSequence returns the last recorded sequence number.
func (gd *GapDetector) GetLastSequence() uint64 {
	gd.mu.Lock()
	defer gd.mu.Unlock()
	return gd.lastSequence
}

// Reset resets the gap detector to its initial state.
// Clears all detected gaps and resets the last sequence.
func (gd *GapDetector) Reset() {
	gd.mu.Lock()
	defer gd.mu.Unlock()
	gd.lastSequence = 0
	gd.gaps = gd.gaps[:0]
}

// getTimeNow returns current Unix timestamp in milliseconds.
func getTimeNow() int64 {
	return time.Now().UnixMilli()
}
