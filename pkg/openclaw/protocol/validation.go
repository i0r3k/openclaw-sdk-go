package protocol

import (
	"errors"
	"strings"
)

// ValidationError represents a validation error (uses Phase 1 error pattern)
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// Validator validates protocol frames
type Validator struct{}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateGatewayFrame validates a gateway frame
func (v *Validator) ValidateGatewayFrame(frame *GatewayFrame) error {
	if frame == nil {
		return errors.New("frame is nil")
	}
	if frame.Type == "" {
		return &ValidationError{Field: "Type", Message: "is required"}
	}
	if !frame.Type.IsValid() {
		return &ValidationError{Field: "Type", Message: "is not a valid frame type"}
	}
	return nil
}

// ValidateRequestFrame validates a request frame
func (v *Validator) ValidateRequestFrame(frame *RequestFrame) error {
	if frame == nil {
		return errors.New("frame is nil")
	}
	if frame.RequestID == "" {
		return &ValidationError{Field: "RequestID", Message: "is required"}
	}
	if frame.Method == "" {
		return &ValidationError{Field: "Method", Message: "is required"}
	}
	// Validate method format (namespace.method or namespace.sub.method)
	parts := strings.Split(frame.Method, ".")
	if len(parts) < 2 {
		return &ValidationError{Field: "Method", Message: "must be in format 'namespace.method'"}
	}
	for _, part := range parts {
		if part == "" {
			return &ValidationError{Field: "Method", Message: "must be in format 'namespace.method'"}
		}
	}
	return nil
}

// ValidateResponseFrame validates a response frame
func (v *Validator) ValidateResponseFrame(frame *ResponseFrame) error {
	if frame == nil {
		return errors.New("frame is nil")
	}
	if frame.RequestID == "" {
		return &ValidationError{Field: "RequestID", Message: "is required"}
	}
	// Success and Error are mutually exclusive
	if frame.Success && frame.Error != nil {
		return &ValidationError{Field: "Error", Message: "must be nil when Success is true"}
	}
	if !frame.Success && frame.Error == nil {
		return &ValidationError{Field: "Error", Message: "is required when Success is false"}
	}
	return nil
}

// ValidateEventFrame validates an event frame
func (v *Validator) ValidateEventFrame(frame *EventFrame) error {
	if frame == nil {
		return errors.New("frame is nil")
	}
	if frame.EventType == "" {
		return &ValidationError{Field: "EventType", Message: "is required"}
	}
	return nil
}
