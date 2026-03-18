// WebSocket echo server example for OpenClaw SDK testing
package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// upgrader upgrades HTTP connections to WebSocket
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins for demo purposes
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	// Register WebSocket handler
	http.HandleFunc("/ws", handleWebSocket)

	// Start server
	addr := ":8080"
	log.Printf("🚀 Echo server started on %s", addr)
	log.Printf("📡 WebSocket endpoint: ws://localhost%s/ws", addr)
	log.Println("Press Ctrl+C to stop")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("❌ Server failed: %v", err)
	}
}

// handleWebSocket handles WebSocket connections
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("⚠️  Upgrade error from %s: %v", r.RemoteAddr, err)
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("⚠️  Close error from %s: %v", r.RemoteAddr, err)
		}
	}()

	log.Printf("✓ Client connected from %s", r.RemoteAddr)

	// Message loop
	for {
		// Read message from client
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("⚠️  Read error from %s: %v", r.RemoteAddr, err)
			} else {
				log.Printf("✗ Client disconnected from %s", r.RemoteAddr)
			}
			break
		}

		log.Printf("📨 Received from %s: %s", r.RemoteAddr, message)

		// Echo back the message
		if err := conn.WriteMessage(messageType, message); err != nil {
			log.Printf("⚠️  Write error to %s: %v", r.RemoteAddr, err)
			break
		}
	}

	log.Printf("🔌 Connection closed for %s", r.RemoteAddr)
}
