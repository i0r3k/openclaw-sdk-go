package api

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/frisbee-ai/openclaw-sdk-go/pkg/protocol"
)

// mockRequest creates a RequestFn that returns the given JSON for all requests.
func mockRequest(responseJSON json.RawMessage, returnErr error) RequestFn {
	return func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		if returnErr != nil {
			return nil, returnErr
		}
		return responseJSON, nil
	}
}

// mockRequestByMethod creates a RequestFn that returns different JSON per method.
func mockRequestByMethod(responses map[string]json.RawMessage) RequestFn {
	return func(ctx context.Context, method string, params any) (json.RawMessage, error) {
		resp, ok := responses[method]
		if !ok {
			return nil, errors.New("method not found: " + method)
		}
		return resp, nil
	}
}

// --- ChatAPI Tests ---

func TestChatAPI_List(t *testing.T) {
	resp := json.RawMessage(`{"chats":[{"chatId":"chat-1"},{"chatId":"chat-2"}]}`)
	api := NewChatAPI(mockRequest(resp, nil))

	result, err := api.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Chats) != 2 {
		t.Errorf("expected 2 chats, got %d", len(result.Chats))
	}
	if result.Chats[0].ChatID != "chat-1" {
		t.Errorf("expected chat-1, got %s", result.Chats[0].ChatID)
	}
}

func TestChatAPI_History(t *testing.T) {
	resp := json.RawMessage(`{"messages":["msg1","msg2"]}`)
	api := NewChatAPI(mockRequest(resp, nil))

	result, err := api.History(context.Background(), protocol.ChatHistoryParams{ChatID: "chat-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(result.Messages))
	}
}

func TestChatAPI_Title(t *testing.T) {
	resp := json.RawMessage(`{"title":"My Chat"}`)
	api := NewChatAPI(mockRequest(resp, nil))

	result, err := api.Title(context.Background(), protocol.ChatTitleParams{ChatID: "chat-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Title != "My Chat" {
		t.Errorf("expected 'My Chat', got %s", result.Title)
	}
}

func TestChatAPI_Delete(t *testing.T) {
	api := NewChatAPI(mockRequest(nil, nil))

	err := api.Delete(context.Background(), protocol.ChatDeleteParams{ChatID: "chat-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestChatAPI_Inject(t *testing.T) {
	api := NewChatAPI(mockRequest(nil, nil))

	err := api.Inject(context.Background(), protocol.ChatInjectParams{ChatID: "chat-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestChatAPI_Error(t *testing.T) {
	api := NewChatAPI(mockRequest(nil, errors.New("network error")))

	_, err := api.List(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "network error" {
		t.Errorf("expected 'network error', got %s", err.Error())
	}
}

func TestChatAPI_InvalidJSON(t *testing.T) {
	api := NewChatAPI(mockRequest(json.RawMessage(`invalid`), nil))

	_, err := api.List(context.Background())
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

// --- AgentsAPI Tests ---

func TestAgentsAPI_Identity(t *testing.T) {
	resp := json.RawMessage(`{"id":"agent-1","summary":{"name":"Test Agent"}}`)
	api := NewAgentsAPI(mockRequest(resp, nil))

	result, err := api.Identity(context.Background(), protocol.AgentIdentityParams{AgentID: "agent-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "agent-1" {
		t.Errorf("expected agent-1, got %s", result.ID)
	}
}

func TestAgentsAPI_List(t *testing.T) {
	resp := json.RawMessage(`{"agents":[{"id":"a1"},{"id":"a2"}]}`)
	api := NewAgentsAPI(mockRequest(resp, nil))

	result, err := api.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Agents) != 2 {
		t.Errorf("expected 2 agents, got %d", len(result.Agents))
	}
}

func TestAgentsAPI_Files(t *testing.T) {
	resp := json.RawMessage(`{"files":["file1.txt","file2.txt"]}`)
	api := NewAgentsAPI(mockRequest(resp, nil))

	filesAPI := api.Files()
	result, err := filesAPI.List(context.Background(), protocol.AgentsFilesListParams{AgentID: "agent-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Files) != 2 {
		t.Errorf("expected 2 files, got %d", len(result.Files))
	}
}

// --- SessionsAPI Tests ---

func TestSessionsAPI_List(t *testing.T) {
	resp := json.RawMessage(`{"sessions":[{"id":"s1","status":"active"}]}`)
	api := NewSessionsAPI(mockRequest(resp, nil))

	result, err := api.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Sessions) != 1 {
		t.Errorf("expected 1 session, got %d", len(result.Sessions))
	}
	if result.Sessions[0].ID != "s1" {
		t.Errorf("expected s1, got %s", result.Sessions[0].ID)
	}
}

func TestSessionsAPI_Usage(t *testing.T) {
	resp := json.RawMessage(`{"usage":{"tokens":100}}`)
	api := NewSessionsAPI(mockRequest(resp, nil))

	result, err := api.Usage(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Usage == nil {
		t.Error("expected usage data")
	}
}

// --- CronAPI Tests ---

func TestCronAPI_List(t *testing.T) {
	resp := json.RawMessage(`[{"id":"job1","cron":"* * * * *","prompt":"test"}]`)
	api := NewCronAPI(mockRequest(resp, nil))

	result, err := api.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 job, got %d", len(result))
	}
	if result[0].ID != "job1" {
		t.Errorf("expected job1, got %s", result[0].ID)
	}
}

func TestCronAPI_Status(t *testing.T) {
	resp := json.RawMessage(`{"id":"job1","cron":"* * * * *","prompt":"test"}`)
	api := NewCronAPI(mockRequest(resp, nil))

	result, err := api.Status(context.Background(), protocol.CronStatusParams{JobID: "job1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "job1" {
		t.Errorf("expected job1, got %s", result.ID)
	}
}

// --- NodesAPI Tests ---

func TestNodesAPI_List(t *testing.T) {
	resp := json.RawMessage(`[{"id":"node1","status":"online"}]`)
	api := NewNodesAPI(mockRequest(resp, nil))

	result, err := api.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 node, got %d", len(result))
	}
	if result[0].ID != "node1" {
		t.Errorf("expected node1, got %s", result[0].ID)
	}
}

func TestNodesAPI_Pairing(t *testing.T) {
	resp := json.RawMessage(`[{"pairingId":"p1","nodeId":"node1","status":"pending"}]`)
	api := NewNodesAPI(mockRequest(resp, nil))

	pairingAPI := api.Pairing()
	result, err := pairingAPI.List(context.Background(), protocol.NodePairListParams{NodeID: "node1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 pairing, got %d", len(result))
	}
}

// --- SkillsAPI Tests ---

func TestSkillsAPI_ToolsCatalog(t *testing.T) {
	resp := json.RawMessage(`{"tools":["tool1","tool2"]}`)
	api := NewSkillsAPI(mockRequest(resp, nil))

	result, err := api.ToolsCatalog(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(result.Tools))
	}
}

func TestSkillsAPI_Bins(t *testing.T) {
	resp := json.RawMessage(`{"bins":["bin1"]}`)
	api := NewSkillsAPI(mockRequest(resp, nil))

	result, err := api.Bins(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Bins) != 1 {
		t.Errorf("expected 1 bin, got %d", len(result.Bins))
	}
}

// --- ConfigAPI Tests ---

func TestConfigAPI_Schema(t *testing.T) {
	resp := json.RawMessage(`{"schema":{"type":"object"}}`)
	api := NewConfigAPI(mockRequest(resp, nil))

	result, err := api.Schema(context.Background(), protocol.ConfigSchemaParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Schema == nil {
		t.Error("expected schema data")
	}
}

// --- DevicePairingAPI Tests ---

func TestDevicePairingAPI_List(t *testing.T) {
	resp := json.RawMessage(`[{"pairingId":"p1","deviceId":"d1","status":"pending"}]`)
	api := NewDevicePairingAPI(mockRequest(resp, nil))

	result, err := api.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 pairing, got %d", len(result))
	}
	if result[0].PairingID != "p1" {
		t.Errorf("expected p1, got %s", result[0].PairingID)
	}
}

// --- Method Routing Tests ---

func TestAPI_MethodRouting(t *testing.T) {
	responses := map[string]json.RawMessage{
		"chat.list":  json.RawMessage(`{"chats":[]}`),
		"chat.title": json.RawMessage(`{"title":"Test"}`),
	}
	mock := mockRequestByMethod(responses)

	chatAPI := NewChatAPI(mock)

	// List should work
	_, err := chatAPI.List(context.Background())
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	// Title should work
	title, err := chatAPI.Title(context.Background(), protocol.ChatTitleParams{ChatID: "c1"})
	if err != nil {
		t.Fatalf("Title failed: %v", err)
	}
	if title.Title != "Test" {
		t.Errorf("expected 'Test', got %s", title.Title)
	}

	// Unknown method should fail
	_, err = chatAPI.History(context.Background(), protocol.ChatHistoryParams{ChatID: "c1"})
	if err == nil {
		t.Error("expected error for unknown method")
	}
}
