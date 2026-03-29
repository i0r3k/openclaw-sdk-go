// Package protocol provides fuzzing tests for protocol parsing
package protocol

import (
	"bytes"
	"encoding/json"
	"testing"
)

// FuzzValidateRequestFrame tests RequestFrame validation with fuzzed input
func FuzzValidateRequestFrame(f *testing.F) {
	// Seed with valid and invalid examples
	f.Add([]byte(`{"id":"test","method":"ping"}`))
	f.Add([]byte(`{"id":"","method":"test"}`))
	f.Add([]byte(`{invalid json`))
	f.Add([]byte(``))
	f.Add([]byte(`{"id":"` + string(make([]byte, 10000)) + `"}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		// Try to parse as JSON (may fail, that's OK)
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic on input: %v", r)
			}
		}()

		var original RequestFrame
		if err := json.Unmarshal(data, &original); err != nil {
			// Invalid JSON is expected for some fuzz inputs
			return
		}

		// Round-trip assertion (D-05): Marshal -> Unmarshal -> compare fields
		marshaled, err := json.Marshal(original)
		if err != nil {
			t.Errorf("RequestFrame marshal failed: %v", err)
			return
		}

		var roundTripped RequestFrame
		if err := json.Unmarshal(marshaled, &roundTripped); err != nil {
			t.Errorf("RequestFrame round-trip unmarshal failed: %v", err)
			return
		}

		// Compare key fields that survive round-trip
		if roundTripped.ID != original.ID {
			t.Errorf("RequestFrame round-trip mismatch: ID=%q want=%q", roundTripped.ID, original.ID)
		}
		if roundTripped.Method != original.Method {
			t.Errorf("RequestFrame round-trip mismatch: Method=%q want=%q", roundTripped.Method, original.Method)
		}
		if !bytes.Equal(roundTripped.Params, original.Params) {
			t.Errorf("RequestFrame round-trip mismatch: Params differ")
		}
	})
}

// FuzzValidateResponseFrame tests ResponseFrame validation with fuzzed input
func FuzzValidateResponseFrame(f *testing.F) {
	// Seed with examples
	f.Add([]byte(`{"id":"test","ok":true,"payload":{}}`))
	f.Add([]byte(`{"id":"","ok":false,"error":{"code":"ERR","message":"test"}}`))
	f.Add([]byte(`{invalid`))
	f.Add([]byte(``))

	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic on input: %v", r)
			}
		}()

		var original ResponseFrame
		if err := json.Unmarshal(data, &original); err != nil {
			return
		}

		// Round-trip assertion (D-05)
		marshaled, err := json.Marshal(original)
		if err != nil {
			t.Errorf("ResponseFrame marshal failed: %v", err)
			return
		}

		var roundTripped ResponseFrame
		if err := json.Unmarshal(marshaled, &roundTripped); err != nil {
			t.Errorf("ResponseFrame round-trip unmarshal failed: %v", err)
			return
		}

		// Compare key fields
		if roundTripped.ID != original.ID {
			t.Errorf("ResponseFrame round-trip mismatch: ID=%q want=%q", roundTripped.ID, original.ID)
		}
		if roundTripped.Ok != original.Ok {
			t.Errorf("ResponseFrame round-trip mismatch: Ok=%v want=%v", roundTripped.Ok, original.Ok)
		}
		if !bytes.Equal(roundTripped.Payload, original.Payload) {
			t.Errorf("ResponseFrame round-trip mismatch: Payload differ")
		}
		if (roundTripped.Error == nil) != (original.Error == nil) {
			t.Errorf("ResponseFrame round-trip mismatch: Error presence mismatch")
		}
		if roundTripped.Error != nil && original.Error != nil {
			if roundTripped.Error.Code != original.Error.Code {
				t.Errorf("ResponseFrame round-trip mismatch: Error.Code=%q want=%q", roundTripped.Error.Code, original.Error.Code)
			}
		}
	})
}

// FuzzValidateEventFrame tests EventFrame validation with fuzzed input
func FuzzValidateEventFrame(f *testing.F) {
	// Seed with examples
	f.Add([]byte(`{"event":"connected","payload":{}}`))
	f.Add([]byte(`{"event":"","payload":null}`))
	f.Add([]byte(`{invalid`))

	f.Fuzz(func(t *testing.T, data []byte) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("panic on input: %v", r)
			}
		}()

		var original EventFrame
		if err := json.Unmarshal(data, &original); err != nil {
			return
		}

		// Round-trip assertion (D-05)
		marshaled, err := json.Marshal(original)
		if err != nil {
			t.Errorf("EventFrame marshal failed: %v", err)
			return
		}

		var roundTripped EventFrame
		if err := json.Unmarshal(marshaled, &roundTripped); err != nil {
			t.Errorf("EventFrame round-trip unmarshal failed: %v", err)
			return
		}

		// Compare key fields
		if roundTripped.Event != original.Event {
			t.Errorf("EventFrame round-trip mismatch: Event=%q want=%q", roundTripped.Event, original.Event)
		}
		if !bytes.Equal(roundTripped.Payload, original.Payload) {
			t.Errorf("EventFrame round-trip mismatch: Payload differ")
		}
		if roundTripped.Seq != original.Seq {
			t.Errorf("EventFrame round-trip mismatch: Seq=%v want=%v", roundTripped.Seq, original.Seq)
		}
	})
}

// FuzzFrameType tests FrameType validation with fuzzed input
func FuzzFrameType(f *testing.F) {
	// Seed with valid types
	f.Add([]byte("req"))
	f.Add([]byte("res"))
	f.Add([]byte("event"))
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

		// Round-trip: FrameType -> String -> FrameType -> IsValid
		ftStr := string(ft)
		ftRoundTrip := FrameType(ftStr)
		if ftRoundTrip.IsValid() != ft.IsValid() {
			t.Errorf("FrameType round-trip mismatch: IsValid=%v want=%v", ftRoundTrip.IsValid(), ft.IsValid())
		}
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
		frame := NewRequestFrame(string(id), "test", nil)
		if frame == nil {
			return
		}

		// Round-trip assertion
		marshaled, err := json.Marshal(frame)
		if err != nil {
			t.Errorf("RequestFrame marshal failed: %v", err)
			return
		}

		var roundTripped RequestFrame
		if err := json.Unmarshal(marshaled, &roundTripped); err != nil {
			t.Errorf("RequestFrame round-trip unmarshal failed: %v", err)
			return
		}

		if roundTripped.ID != frame.ID {
			t.Errorf("RequestFrame round-trip mismatch: ID=%q want=%q", roundTripped.ID, frame.ID)
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
		frame := NewRequestFrame("test-id", string(method), nil)
		if frame == nil {
			return
		}

		// Round-trip assertion
		marshaled, err := json.Marshal(frame)
		if err != nil {
			t.Errorf("RequestFrame marshal failed: %v", err)
			return
		}

		var roundTripped RequestFrame
		if err := json.Unmarshal(marshaled, &roundTripped); err != nil {
			t.Errorf("RequestFrame round-trip unmarshal failed: %v", err)
			return
		}

		if roundTripped.Method != frame.Method {
			t.Errorf("RequestFrame round-trip mismatch: Method=%q want=%q", roundTripped.Method, frame.Method)
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
		frame := NewEventFrame(string(eventType), nil)
		if frame == nil {
			return
		}

		// Round-trip assertion
		marshaled, err := json.Marshal(frame)
		if err != nil {
			t.Errorf("EventFrame marshal failed: %v", err)
			return
		}

		var roundTripped EventFrame
		if err := json.Unmarshal(marshaled, &roundTripped); err != nil {
			t.Errorf("EventFrame round-trip unmarshal failed: %v", err)
			return
		}

		if roundTripped.Event != frame.Event {
			t.Errorf("EventFrame round-trip mismatch: Event=%q want=%q", roundTripped.Event, frame.Event)
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

		// Attempt to unmarshal into RequestFrame and round-trip
		var reqFrame RequestFrame
		if err := json.Unmarshal(data, &reqFrame); err == nil {
			// Valid JSON - test round-trip
			marshaled, err := json.Marshal(reqFrame)
			if err == nil {
				var roundTripped RequestFrame
				if err := json.Unmarshal(marshaled, &roundTripped); err == nil {
					if roundTripped.ID != reqFrame.ID {
						t.Errorf("RequestFrame round-trip mismatch: ID=%q want=%q", roundTripped.ID, reqFrame.ID)
					}
				}
			}
		}
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

		// Test with RequestFrame and round-trip
		var reqFrame RequestFrame
		if err := json.Unmarshal(data, &reqFrame); err == nil {
			marshaled, err := json.Marshal(reqFrame)
			if err == nil {
				var roundTripped RequestFrame
				if err := json.Unmarshal(marshaled, &roundTripped); err == nil {
					if roundTripped.ID != reqFrame.ID {
						t.Errorf("RequestFrame round-trip mismatch: ID=%q want=%q", roundTripped.ID, reqFrame.ID)
					}
				}
			}
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

		// Test with special characters in RequestFrame and round-trip
		frame := NewRequestFrame(string(data), string(data), nil)
		if frame == nil {
			return
		}

		marshaled, err := json.Marshal(frame)
		if err != nil {
			t.Errorf("RequestFrame marshal failed: %v", err)
			return
		}

		var roundTripped RequestFrame
		if err := json.Unmarshal(marshaled, &roundTripped); err != nil {
			t.Errorf("RequestFrame round-trip unmarshal failed: %v", err)
			return
		}

		if roundTripped.ID != frame.ID {
			t.Errorf("RequestFrame round-trip mismatch: ID=%q want=%q", roundTripped.ID, frame.ID)
		}
	})
}
