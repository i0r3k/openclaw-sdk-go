# Phase 10: Examples

**Files:**
- Create: `examples/cmd/main.go`
- Create: `examples/server/main.go`

Note: Examples remain in `examples/` directory (not under pkg/openclaw/) as they are standalone applications.

**Project Structure:** Go module in root, source files in `pkg/openclaw/` directory

**Depends on:** Phase 9 (main client)

---

## Task 10.1: CLI Example

- [ ] **Step 1: Create examples directory**

```bash
mkdir -p examples/cmd
```

```go
// examples/cmd/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	openclaw "github.com/frisbee-ai/openclaw-sdk-go"
	"github.com/frisbee-ai/openclaw-sdk-go/pkg/openclaw/protocol"
)

func main() {
	// Create client
	client, err := openclaw.NewClient(
		openclaw.WithURL("ws://localhost:8080/ws"),
		openclaw.WithLogger(&logger{}),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Subscribe to events
	client.Subscribe(openclaw.EventConnect, func(e openclaw.Event) {
		fmt.Println("Connected!")
	})
	client.Subscribe(openclaw.EventDisconnect, func(e openclaw.Event) {
		fmt.Println("Disconnected!")
	})
	client.Subscribe(openclaw.EventError, func(e openclaw.Event) {
		fmt.Printf("Error: %v\n", e.Err)
	})

	// Connect
	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// Send a request
	resp, err := client.SendRequest(ctx, &protocol.RequestFrame{
		RequestID: "req-1",
		Method:    "ping",
		Timestamp: time.Now(),
	})
	if err != nil {
		log.Printf("Request failed: %v", err)
	} else {
		fmt.Printf("Response: %+v\n", resp)
	}

	// Wait for interrupt
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("Shutting down...")
}

type logger struct{}

func (l *logger) Debug(msg string, args ...any) { log.Printf("[DEBUG] "+msg, args...) }
func (l *logger) Info(msg string, args ...any)  { log.Printf("[INFO] "+msg, args...) }
func (l *logger) Warn(msg string, args ...any)  { log.Printf("[WARN] "+msg, args...) }
func (l *logger) Error(msg string, args ...any) { log.Printf("[ERROR] "+msg, args...) }
```

- [ ] **Step 2: Commit**

```bash
git add examples/cmd/main.go
git commit -m "feat: add CLI example"
```

---

## Task 10.2: Server Example

- [ ] **Step 1: Create examples/server directory**

```bash
mkdir -p examples/server
```

```go
// examples/server/main.go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)

	log.Println("Server started on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("Client connected from %s", r.RemoteAddr)

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Read error: %v", err)
			break
		}

		log.Printf("Received: %s", message)

		// Echo back
		if err := conn.WriteMessage(messageType, message); err != nil {
			log.Printf("Write error: %v", err)
			break
		}
	}

	fmt.Println("Client disconnected")
}
```

- [ ] **Step 2: Commit**

```bash
git add examples/server/main.go
git commit -m "feat: add server example"
```

---

## Phase 10 Complete

After this phase, you should have:
- `examples/cmd/main.go` - CLI example
- `examples/server/main.go` - Server example

---

## Implementation Complete

All 10 phases are now documented. The complete implementation plan includes:

1. **Phase 1** - Project Setup (go.mod, types, errors, logger)
2. **Phase 2** - Auth Module (CredentialsProvider, AuthHandler)
3. **Phase 3** - Protocol Module (types, validation)
4. **Phase 4** - Transport Module (WebSocket)
5. **Phase 5** - Connection Module (state machine, negotiator, policies, TLS)
6. **Phase 6** - Events Module (tick monitor, gap detector)
7. **Phase 7** - Managers Module (event, request, connection, reconnect)
8. **Phase 8** - Utils Module (timeout manager)
9. **Phase 9** - Main Client
10. **Phase 10** - Examples
