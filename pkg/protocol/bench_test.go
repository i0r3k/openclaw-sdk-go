// Package protocol provides benchmarks for protocol serialization.
package protocol

import (
	"encoding/json"
	"testing"
)

// BenchmarkProtocolMarshal measures JSON marshaling performance for protocol frames.
func BenchmarkProtocolMarshal(b *testing.B) {
	// Create representative frames with payloads
	reqFrame := &RequestFrame{
		Type:   FrameTypeRequest,
		ID:     "bench-req-001",
		Method: "client.ping",
		Params: json.RawMessage(`{"key":"value string"}`),
	}
	respFrameSuccess := &ResponseFrame{
		Type:    FrameTypeResponse,
		ID:      "bench-res-001",
		Ok:      true,
		Payload: json.RawMessage(`{"result":"ok"}`),
	}
	respFrameError := &ResponseFrame{
		Type:  FrameTypeResponse,
		ID:    "bench-res-002",
		Ok:    false,
		Error: &ErrorShape{Code: "ERR", Message: "error"},
	}
	eventFrame := &EventFrame{
		Type:    FrameTypeEvent,
		Event:   "tick",
		Payload: json.RawMessage(`{"ts":1234567890}`),
	}

	// Scaling pattern per D-17: test at different operation counts
	sizes := []int{100, 1000, 10000}
	for _, size := range sizes {
		b.Run(string(rune('0'+size/1000)), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				data, _ := json.Marshal(reqFrame)
				b.ReportMetric(float64(len(data)), "bytes_per_frame")
				_ = data

				data, _ = json.Marshal(respFrameSuccess)
				b.ReportMetric(float64(len(data)), "bytes_per_frame")
				_ = data

				data, _ = json.Marshal(respFrameError)
				b.ReportMetric(float64(len(data)), "bytes_per_frame")
				_ = data

				data, _ = json.Marshal(eventFrame)
				b.ReportMetric(float64(len(data)), "bytes_per_frame")
				_ = data
			}
			b.SetBytes(int64(size * 4)) // 4 frames per iteration
		})
	}
}

// BenchmarkProtocolUnmarshal measures JSON unmarshaling performance for protocol frames.
func BenchmarkProtocolUnmarshal(b *testing.B) {
	// Pre-serialize JSON bytes for each frame type
	reqJSON := []byte(`{"type":"req","id":"bench-req-001","method":"client.ping","params":{"key":"value string"}}`)
	respSuccessJSON := []byte(`{"type":"res","id":"bench-res-001","ok":true,"payload":{"result":"ok"}}`)
	respErrorJSON := []byte(`{"type":"res","id":"bench-res-002","ok":false,"error":{"code":"ERR","message":"error"}}`)
	eventJSON := []byte(`{"type":"event","event":"tick","payload":{"ts":1234567890}}`)

	// Scaling pattern per D-17
	sizes := []int{100, 1000, 10000}
	for _, size := range sizes {
		b.Run(string(rune('0'+size/1000)), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				var req RequestFrame
				_ = json.Unmarshal(reqJSON, &req)

				var respSuccess ResponseFrame
				_ = json.Unmarshal(respSuccessJSON, &respSuccess)

				var respError ResponseFrame
				_ = json.Unmarshal(respErrorJSON, &respError)

				var event EventFrame
				_ = json.Unmarshal(eventJSON, &event)
			}
			b.SetBytes(int64(size * 4)) // 4 frames per iteration
		})
	}
}

// BenchmarkProtocolUnmarshalLargePayload measures unmarshaling with large payloads.
func BenchmarkProtocolUnmarshalLargePayload(b *testing.B) {
	// Generate a large payload (1MB string)
	largePayload := make([]byte, 1024*1024)
	for i := range largePayload {
		largePayload[i] = byte('a' + (i % 26))
	}
	largeJSON := []byte(`{"type":"req","id":"bench-large","method":"test","params":{"data":"` + string(largePayload) + `"}}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var req RequestFrame
		_ = json.Unmarshal(largeJSON, &req)
	}
	b.SetBytes(1024 * 1024)
}

// BenchmarkProtocolMarshalLargePayload measures marshaling with large payloads.
func BenchmarkProtocolMarshalLargePayload(b *testing.B) {
	params := make(map[string]string)
	params["data"] = string(make([]byte, 1024*1024))
	paramsJSON, _ := json.Marshal(params)

	reqFrame := &RequestFrame{
		Type:   FrameTypeRequest,
		ID:     "bench-large",
		Method: "test",
		Params: paramsJSON,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, _ := json.Marshal(reqFrame)
		b.ReportMetric(float64(len(data)), "bytes_per_frame")
		_ = data
	}
	b.SetBytes(1024 * 1024)
}
