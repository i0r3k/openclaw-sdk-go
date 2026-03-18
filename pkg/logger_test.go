package openclaw

import (
	"bytes"
	"context"
	"testing"
)

func TestDefaultLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewDefaultLoggerWithWriter(buf)

	logger.Info("test message %s", "world")
	logger.Debug("debug message")
	logger.Warn("warning message")
	logger.Error("error message")

	output := buf.String()
	if output == "" {
		t.Error("expected logger output")
	}
}

func TestNopLogger(t *testing.T) {
	logger := &NopLogger{}
	// Should not panic
	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")
}

func TestLoggerInterface(t *testing.T) {
	// Verify DefaultLogger implements Logger
	var _ Logger = &DefaultLogger{}
	// Verify NopLogger implements Logger
	var _ Logger = &NopLogger{}
}

func TestWithContext(t *testing.T) {
	ctx := context.Background()
	logger := &NopLogger{}

	ctx = WithContext(ctx, logger)
	retrieved, ok := FromContext(ctx)

	if !ok {
		t.Error("expected to retrieve logger from context")
	}
	if retrieved != logger {
		t.Error("expected to retrieve same logger")
	}
}
