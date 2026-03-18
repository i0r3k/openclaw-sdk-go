package openclaw

import (
	"context"
	"io"
	"log"
	"os"
)

// Logger interface for customizable logging
// Follows standard Go logging patterns with level support
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// DefaultLogger uses stdlib log
type DefaultLogger struct {
	debug *log.Logger
	info  *log.Logger
	warn  *log.Logger
	error *log.Logger
}

// NewDefaultLogger creates a logger that writes to stdout
func NewDefaultLogger() *DefaultLogger {
	return &DefaultLogger{
		debug: log.New(os.Stdout, "[DEBUG] ", log.Ldate|log.Ltime),
		info:  log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime),
		warn:  log.New(os.Stdout, "[WARN] ", log.Ldate|log.Ltime),
		error: log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime),
	}
}

// NewDefaultLoggerWithWriter creates a logger with custom writer
func NewDefaultLoggerWithWriter(w io.Writer) *DefaultLogger {
	return &DefaultLogger{
		debug: log.New(w, "[DEBUG] ", log.Ldate|log.Ltime),
		info:  log.New(w, "[INFO] ", log.Ldate|log.Ltime),
		warn:  log.New(w, "[WARN] ", log.Ldate|log.Ltime),
		error: log.New(w, "[ERROR] ", log.Ldate|log.Ltime),
	}
}

func (l *DefaultLogger) Debug(msg string, args ...any) { l.debug.Printf(msg, args...) }
func (l *DefaultLogger) Info(msg string, args ...any)  { l.info.Printf(msg, args...) }
func (l *DefaultLogger) Warn(msg string, args ...any)  { l.warn.Printf(msg, args...) }
func (l *DefaultLogger) Error(msg string, args ...any) { l.error.Printf(msg, args...) }

// NopLogger is a no-op implementation for testing
type NopLogger struct{}

func (l *NopLogger) Debug(msg string, args ...any) {}
func (l *NopLogger) Info(msg string, args ...any)  {}
func (l *NopLogger) Warn(msg string, args ...any)  {}
func (l *NopLogger) Error(msg string, args ...any) {}

// WithContext creates a context with logger
func WithContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// FromContext retrieves logger from context
func FromContext(ctx context.Context) (Logger, bool) {
	logger, ok := ctx.Value(loggerKey{}).(Logger)
	return logger, ok
}

type loggerKey struct{}
