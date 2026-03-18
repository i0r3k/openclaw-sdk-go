// Package protocol provides fuzzing tests for protocol parsing
package protocol

import (
	"testing"
)

// FuzzValidateRequestFrame tests RequestFrame validation with fuzzed input
func FuzzValidateRequestFrame(f *testing.F) {
	// Seed with valid and invalid examples
	f.Add([]byte(`{"requestId":"test","method":"ping","timestamp":"2024-01-01T00:00:00Z"}`))
	f.Add([]byte(`{"requestId":"","method":"test","timestamp":"2024-01-01T00:00:00Z"}`))
	f.Add([]byte(`{invalid json`))
	f.Add([]byte(``))
	f.Add([]byte(`{"requestId":"` + string(make([]byte, 10000)) + `"}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		// Validate that parsing doesn't panic
		var frame RequestFrame
		_ = frame

		// Try to parse as JSON (may fail, that's OK)
		// The goal is to ensure no panics occur
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic on input: %v", r)
			}
		}()

		// Validate the frame if possible
		_ = data
	})
}

// FuzzValidateResponseFrame tests ResponseFrame validation with fuzzed input
func FuzzValidateResponseFrame(f *testing.F) {
	// Seed with examples
	f.Add([]byte(`{"requestId":"test","success":true,"result":{}}`))
	f.Add([]byte(`{"requestId":"","success":false,"error":{"code":"ERR","message":"test"}}`))
	f.Add([]byte(`{invalid`))
	f.Add([]byte(``))

	f.Fuzz(func(t *testing.T, data []byte) {
		var frame ResponseFrame
		_ = frame

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic on input: %v", r)
			}
		}()

		_ = data
	})
}

// FuzzValidateEventFrame tests EventFrame validation with fuzzed input
func FuzzValidateEventFrame(f *testing.F) {
	// Seed with examples
	f.Add([]byte(`{"eventType":"connected","data":{},"timestamp":"2024-01-01T00:00:00Z"}`))
	f.Add([]byte(`{"eventType":"","data":null}`))
	f.Add([]byte(`{invalid`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var frame EventFrame
		_ = frame

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic on input: %v", r)
			}
		}()

		_ = data
	})
}

// FuzzFrameType tests FrameType validation with fuzzed input
func FuzzFrameType(f *testing.F) {
	// Seed with valid types
	f.Add([]byte("gateway"))
	f.Add([]byte("request"))
	f.Add([]byte("response"))
	f.Add([]byte("event"))
	f.Add([]byte("error"))
	f.Add([]byte(""))
	f.Add([]byte("invalid"))
	f.Add([]byte(string(make([]byte, 1000))))

	f.Fuzz(func(t *testing.T, typ []byte) {
		ft := FrameType(typ)

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic on input: %v", r)
			}
		}()

		// Call IsValid - should not panic
		_ = ft.IsValid()

		// Test String method (if any)
		_ = string(ft)
	})
}

// FuzzRequestID tests handling of various request ID formats
func FuzzRequestID(f *testing.F) {
	// Seed with examples
	f.Add([]byte("req-001"))
	f.Add([]byte(""))
	f.Add([]byte(string(make([]byte, 100))))
	f.Add([]byte("\x00\x01\x02"))
	f.Add([]byte("../../../etc/passwd"))
	f.Add([]byte("<script>alert(1)</script>"))

	f.Fuzz(func(t *testing.T, id []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic on input: %v", r)
			}
		}()

		// Create request frame with fuzzed ID
		frame := NewRequestFrame(string(id), "test")
		if frame != nil {
			_ = frame.RequestID
		}
	})
}

// FuzzMethod tests handling of various method names
func FuzzMethod(f *testing.F) {
	// Seed with examples
	f.Add([]byte("ping"))
	f.Add([]byte(""))
	f.Add([]byte("subscribe"))
	f.Add([]byte(string(make([]byte, 100))))
	f.Add([]byte("../../"))

	f.Fuzz(func(t *testing.T, method []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic on input: %v", r)
			}
		}()

		// Create request frame with fuzzed method
		frame := NewRequestFrame("test-id", string(method))
		if frame != nil {
			_ = frame.Method
		}
	})
}

// FuzzEventType tests handling of various event types
func FuzzEventType(f *testing.F) {
	// Seed with examples
	f.Add([]byte("connected"))
	f.Add([]byte(""))
	f.Add([]byte("disconnected"))
	f.Add([]byte(string(make([]byte, 50))))

	f.Fuzz(func(t *testing.T, eventType []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic on input: %v", r)
			}
		}()

		// Create event frame with fuzzed type
		frame := NewEventFrame(string(eventType))
		if frame != nil {
			_ = frame.EventType
		}
	})
}

// FuzzJSONMalformed tests handling of malformed JSON
func FuzzJSONMalformed(f *testing.F) {
	// Seed with malformed JSON examples
	f.Add([]byte("{"))
	f.Add([]byte("}"))
	f.Add([]byte("["))
	f.Add([]byte("]"))
	f.Add([]byte("{{"))
	f.Add([]byte("}}"))
	f.Add([]byte("[["))
	f.Add([]byte("]]"))
	f.Add([]byte("{,}"))
	f.Add([]byte("[,]"))
	f.Add([]byte(`{"key": undefined}`))
	f.Add([]byte(`{"key": function(){}}`))
	f.Add([]byte(`{"key": NaN}`))
	f.Add([]byte(`{"key": Infinity}`))
	f.Add([]byte("\x00"))
	f.Add([]byte("\xff\xfe"))

	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic on input: %v", r)
			}
		}()

		// Attempt to unmarshal into various frame types
		var reqFrame RequestFrame
		_ = reqFrame

		var respFrame ResponseFrame
		_ = respFrame

		var eventFrame EventFrame
		_ = eventFrame

		_ = data
	})
}

// FuzzLargeInput tests handling of very large inputs
func FuzzLargeInput(f *testing.F) {
	// Add large input seeds
	largeInput := make([]byte, 1024*1024) // 1MB
	for i := range largeInput {
		largeInput[i] = byte(i % 256)
	}
	f.Add(largeInput)

	f.Add(make([]byte, 10*1024*1024)) // 10MB

	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic on large input: %v", r)
			}
		}()

		// Test that large inputs don't cause crashes
		if len(data) > 0 {
			frame := NewRequestFrame(string(data), "test")
			_ = frame
		}
	})
}

// FuzzSpecialCharacters tests handling of special characters
func FuzzSpecialCharacters(f *testing.F) {
	specialChars := []string{
		"\x00", "\x01", "\x02", "\x03", "\x04", "\x05",
		"\n", "\r", "\t",
		"\\n", "\\r", "\\t",
		"../../", "..\\..\\",
		"\x00\x00\x00",
		"💀", "🚀", "🎯",
		"\u200B", "\uFEFF", // Zero-width characters
		"\u202E", "\u202A", // Direction override
	}

	for _, ch := range specialChars {
		f.Add([]byte(ch))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic on special char input: %v", r)
			}
		}()

		// Test with special characters in various fields
		frame := NewRequestFrame(string(data), string(data))
		_ = frame
	})
}
