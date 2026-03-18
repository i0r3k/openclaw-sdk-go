// Command-line client example for OpenClaw SDK
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	openclaw "github.com/i0r3k/openclaw-sdk-go/pkg"
	"github.com/i0r3k/openclaw-sdk-go/pkg/protocol"
	"github.com/i0r3k/openclaw-sdk-go/pkg/types"
)

func main() {
	// Create client with custom logger
	client, err := openclaw.NewClient(
		openclaw.WithURL("ws://localhost:8080/ws"),
		openclaw.WithLogger(&logger{}),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer func() {
		_ = client.Close()
	}()

	// Subscribe to connection events
	client.Subscribe(types.EventConnect, func(e types.Event) {
		fmt.Println("✓ Connected to server!")
	})
	client.Subscribe(types.EventDisconnect, func(e types.Event) {
		fmt.Println("✗ Disconnected from server")
	})
	client.Subscribe(types.EventError, func(e types.Event) {
		fmt.Printf("✗ Error: %v\n", e.Err)
	})

	// Connect to server
	fmt.Println("Connecting to ws://localhost:8080/ws...")
	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	// Send a ping request
	fmt.Println("\nSending ping request...")
	resp, err := client.SendRequest(ctx, &protocol.RequestFrame{
		RequestID: "req-" + time.Now().Format("20060102150405"),
		Method:    "ping",
		Timestamp: time.Now(),
	})
	if err != nil {
		log.Printf("Request failed: %v", err)
	} else {
		fmt.Printf("Response: Success=%v, Result=%s\n", resp.Success, string(resp.Result))
	}

	// Wait for interrupt signal
	fmt.Println("\nPress Ctrl+C to exit...")
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("\nShutting down...")
	if err := client.Disconnect(); err != nil {
		log.Printf("Disconnect error: %v", err)
	}
}

// logger implements openclaw.Logger interface
type logger struct{}

func (l *logger) Debug(msg string, args ...any) {
	log.Printf("[DEBUG] "+msg, args...)
}

func (l *logger) Info(msg string, args ...any) {
	log.Printf("[INFO] "+msg, args...)
}

func (l *logger) Warn(msg string, args ...any) {
	log.Printf("[WARN] "+msg, args...)
}

func (l *logger) Error(msg string, args ...any) {
	log.Printf("[ERROR] "+msg, args...)
}
