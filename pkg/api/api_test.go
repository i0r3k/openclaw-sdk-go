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

func TestAgentsAPI_FilesList(t *testing.T) {
	resp := json.RawMessage(`{"files":["file1.txt","file2.txt"]}`)
	api := NewAgentsAPI(mockRequest(resp, nil))

	result, err := api.FilesList(context.Background(), protocol.AgentsFilesListParams{AgentID: "agent-1"})
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

	result, err := api.PairingList(context.Background(), protocol.NodePairListParams{NodeID: "node1"})
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

// --- AgentsAPI Tests (continued) ---

func TestAgentsAPI_Wait(t *testing.T) {
	api := NewAgentsAPI(mockRequest(nil, nil))
	err := api.Wait(context.Background(), protocol.AgentWaitParams{AgentID: "agent-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAgentsAPI_Create(t *testing.T) {
	resp := json.RawMessage(`{"agentId":"new-agent"}`)
	api := NewAgentsAPI(mockRequest(resp, nil))
	result, err := api.Create(context.Background(), protocol.AgentsCreateParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AgentID != "new-agent" {
		t.Errorf("expected new-agent, got %s", result.AgentID)
	}
}

func TestAgentsAPI_Update(t *testing.T) {
	resp := json.RawMessage(`{"agentId":"agent-1"}`)
	api := NewAgentsAPI(mockRequest(resp, nil))
	result, err := api.Update(context.Background(), protocol.AgentsUpdateParams{AgentID: "agent-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AgentID != "agent-1" {
		t.Errorf("expected agent-1, got %s", result.AgentID)
	}
}

func TestAgentsAPI_Delete(t *testing.T) {
	resp := json.RawMessage(`{"agentId":"agent-1"}`)
	api := NewAgentsAPI(mockRequest(resp, nil))
	result, err := api.Delete(context.Background(), protocol.AgentsDeleteParams{AgentID: "agent-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AgentID != "agent-1" {
		t.Errorf("expected agent-1, got %s", result.AgentID)
	}
}

func TestAgentsAPI_Files_Get(t *testing.T) {
	resp := json.RawMessage(`{"content":"file content"}`)
	api := NewAgentsAPI(mockRequest(resp, nil))
	result, err := api.FilesGet(context.Background(), protocol.AgentsFilesGetParams{AgentID: "agent-1", Path: "/test.txt"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Content != "file content" {
		t.Errorf("expected 'file content', got %s", result.Content)
	}
}

func TestAgentsAPI_Files_Set(t *testing.T) {
	resp := json.RawMessage(`{}`)
	api := NewAgentsAPI(mockRequest(resp, nil))
	_, err := api.FilesSet(context.Background(), protocol.AgentsFilesSetParams{AgentID: "agent-1", Path: "/test.txt", Content: "file content"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- SessionsAPI Tests (continued) ---

func TestSessionsAPI_Preview(t *testing.T) {
	resp := json.RawMessage(`{"preview":"preview text"}`)
	api := NewSessionsAPI(mockRequest(resp, nil))
	result, err := api.Preview(context.Background(), protocol.SessionsPreviewParams{SessionID: "s1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Preview != "preview text" {
		t.Errorf("expected 'preview text', got %s", result.Preview)
	}
}

func TestSessionsAPI_Patch(t *testing.T) {
	api := NewSessionsAPI(mockRequest(nil, nil))
	err := api.Patch(context.Background(), protocol.SessionsPatchParams{SessionID: "s1", Patch: "some patch"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSessionsAPI_Reset(t *testing.T) {
	api := NewSessionsAPI(mockRequest(nil, nil))
	err := api.Reset(context.Background(), protocol.SessionsResetParams{SessionID: "s1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSessionsAPI_Delete(t *testing.T) {
	api := NewSessionsAPI(mockRequest(nil, nil))
	err := api.Delete(context.Background(), protocol.SessionsDeleteParams{SessionID: "s1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSessionsAPI_Compact(t *testing.T) {
	api := NewSessionsAPI(mockRequest(nil, nil))
	err := api.Compact(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- CronAPI Tests (continued) ---

func TestCronAPI_Add(t *testing.T) {
	api := NewCronAPI(mockRequest(nil, nil))
	err := api.Add(context.Background(), protocol.CronAddParams{Prompt: "test prompt", Cron: "* * * * *"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCronAPI_Update(t *testing.T) {
	api := NewCronAPI(mockRequest(nil, nil))
	err := api.Update(context.Background(), protocol.CronUpdateParams{JobID: "job1", Cron: "0 * * * *"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCronAPI_Remove(t *testing.T) {
	api := NewCronAPI(mockRequest(nil, nil))
	err := api.Remove(context.Background(), protocol.CronRemoveParams{JobID: "job1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCronAPI_Run(t *testing.T) {
	api := NewCronAPI(mockRequest(nil, nil))
	err := api.Run(context.Background(), protocol.CronRunParams{JobID: "job1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCronAPI_Runs(t *testing.T) {
	resp := json.RawMessage(`[{"timestamp":1234567890}]`)
	api := NewCronAPI(mockRequest(resp, nil))
	result, err := api.Runs(context.Background(), protocol.CronRunsParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 run log entry, got %d", len(result))
	}
	if result[0].Timestamp != 1234567890 {
		t.Errorf("expected 1234567890, got %d", result[0].Timestamp)
	}
}

// --- NodesAPI Tests (continued) ---

func TestNodesAPI_Invoke(t *testing.T) {
	resp := json.RawMessage(`{"result":"success"}`)
	api := NewNodesAPI(mockRequest(resp, nil))
	result, err := api.Invoke(context.Background(), protocol.NodeInvokeParams{NodeID: "node-1", Target: "test.target"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != `{"result":"success"}` {
		t.Errorf("unexpected result: %s", string(result))
	}
}

func TestNodesAPI_Event(t *testing.T) {
	api := NewNodesAPI(mockRequest(nil, nil))
	err := api.Event(context.Background(), protocol.NodeEventParams{NodeID: "node-1", Event: "test.event", Payload: map[string]any{"key": "value"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNodesAPI_PendingDrain(t *testing.T) {
	resp := json.RawMessage(`{"items":[]}`)
	api := NewNodesAPI(mockRequest(resp, nil))
	result, err := api.PendingDrain(context.Background(), protocol.NodePendingDrainParams{NodeID: "node-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Items == nil {
		t.Error("expected items to not be nil")
	}
}

func TestNodesAPI_PendingEnqueue(t *testing.T) {
	api := NewNodesAPI(mockRequest(nil, nil))
	err := api.PendingEnqueue(context.Background(), protocol.NodePendingEnqueueParams{NodeID: "node-1", Item: map[string]any{"data": "test"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNodesAPI_Pairing_Approve(t *testing.T) {
	api := NewNodesAPI(mockRequest(nil, nil))
	err := api.PairingApprove(context.Background(), protocol.NodePairApproveParams{NodeID: "node-1", PairingID: "pair-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNodesAPI_Pairing_Reject(t *testing.T) {
	api := NewNodesAPI(mockRequest(nil, nil))
	err := api.PairingReject(context.Background(), protocol.NodePairRejectParams{NodeID: "node-1", PairingID: "pair-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNodesAPI_Pairing_Verify(t *testing.T) {
	api := NewNodesAPI(mockRequest(nil, nil))
	err := api.PairingVerify(context.Background(), protocol.NodePairVerifyParams{NodeID: "node-1", PairingID: "pair-1", Code: "123456"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- SkillsAPI Tests (continued) ---

func TestSkillsAPI_Status(t *testing.T) {
	resp := json.RawMessage(`{"status":"running","skills":["skill1","skill2"]}`)
	api := NewSkillsAPI(mockRequest(resp, nil))
	result, err := api.Status(context.Background(), protocol.SkillsStatusParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != `{"status":"running","skills":["skill1","skill2"]}` {
		t.Errorf("unexpected result: %s", string(result))
	}
}

func TestSkillsAPI_Install(t *testing.T) {
	api := NewSkillsAPI(mockRequest(nil, nil))
	err := api.Install(context.Background(), protocol.SkillsInstallParams{SkillID: "skill-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSkillsAPI_Update(t *testing.T) {
	api := NewSkillsAPI(mockRequest(nil, nil))
	err := api.Update(context.Background(), protocol.SkillsUpdateParams{SkillID: "skill-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- ConfigAPI Tests (continued) ---

func TestConfigAPI_Get(t *testing.T) {
	resp := json.RawMessage(`{"key":"value"}`)
	api := NewConfigAPI(mockRequest(resp, nil))
	result, err := api.Get(context.Background(), protocol.ConfigGetParams{Key: "test.key"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != `{"key":"value"}` {
		t.Errorf("unexpected result: %s", string(result))
	}
}

func TestConfigAPI_Set(t *testing.T) {
	api := NewConfigAPI(mockRequest(nil, nil))
	err := api.Set(context.Background(), protocol.ConfigSetParams{Key: "test.key", Value: "new value"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigAPI_Apply(t *testing.T) {
	api := NewConfigAPI(mockRequest(nil, nil))
	err := api.Apply(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigAPI_Patch(t *testing.T) {
	api := NewConfigAPI(mockRequest(nil, nil))
	err := api.Patch(context.Background(), protocol.ConfigPatchParams{Patches: []protocol.ConfigPatchOp{{Op: "replace", Path: "/key", Value: "new value"}}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- Error Handling Tests ---

func TestAgentsAPI_Error(t *testing.T) {
	api := NewAgentsAPI(mockRequest(nil, errors.New("agents error")))
	_, err := api.List(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "agents error" {
		t.Errorf("expected 'agents error', got %s", err.Error())
	}
}

func TestAgentsAPI_Create_Error(t *testing.T) {
	api := NewAgentsAPI(mockRequest(nil, errors.New("create failed")))
	_, err := api.Create(context.Background(), protocol.AgentsCreateParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAgentsAPI_Update_Error(t *testing.T) {
	api := NewAgentsAPI(mockRequest(nil, errors.New("update failed")))
	_, err := api.Update(context.Background(), protocol.AgentsUpdateParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAgentsAPI_Delete_Error(t *testing.T) {
	api := NewAgentsAPI(mockRequest(nil, errors.New("delete failed")))
	_, err := api.Delete(context.Background(), protocol.AgentsDeleteParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAgentsAPI_Files_Get_Error(t *testing.T) {
	api := NewAgentsAPI(mockRequest(nil, errors.New("get failed")))
	_, err := api.FilesGet(context.Background(), protocol.AgentsFilesGetParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAgentsAPI_Files_Set_Error(t *testing.T) {
	api := NewAgentsAPI(mockRequest(nil, errors.New("set failed")))
	_, err := api.FilesSet(context.Background(), protocol.AgentsFilesSetParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSessionsAPI_Preview_Error(t *testing.T) {
	api := NewSessionsAPI(mockRequest(nil, errors.New("preview failed")))
	_, err := api.Preview(context.Background(), protocol.SessionsPreviewParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSessionsAPI_Patch_Error(t *testing.T) {
	api := NewSessionsAPI(mockRequest(nil, errors.New("patch failed")))
	err := api.Patch(context.Background(), protocol.SessionsPatchParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSessionsAPI_Reset_Error(t *testing.T) {
	api := NewSessionsAPI(mockRequest(nil, errors.New("reset failed")))
	err := api.Reset(context.Background(), protocol.SessionsResetParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSessionsAPI_Delete_Error(t *testing.T) {
	api := NewSessionsAPI(mockRequest(nil, errors.New("delete failed")))
	err := api.Delete(context.Background(), protocol.SessionsDeleteParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSessionsAPI_Compact_Error(t *testing.T) {
	api := NewSessionsAPI(mockRequest(nil, errors.New("compact failed")))
	err := api.Compact(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCronAPI_Add_Error(t *testing.T) {
	api := NewCronAPI(mockRequest(nil, errors.New("add failed")))
	err := api.Add(context.Background(), protocol.CronAddParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCronAPI_Update_Error(t *testing.T) {
	api := NewCronAPI(mockRequest(nil, errors.New("update failed")))
	err := api.Update(context.Background(), protocol.CronUpdateParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCronAPI_Remove_Error(t *testing.T) {
	api := NewCronAPI(mockRequest(nil, errors.New("remove failed")))
	err := api.Remove(context.Background(), protocol.CronRemoveParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCronAPI_Run_Error(t *testing.T) {
	api := NewCronAPI(mockRequest(nil, errors.New("run failed")))
	err := api.Run(context.Background(), protocol.CronRunParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCronAPI_Runs_Error(t *testing.T) {
	api := NewCronAPI(mockRequest(nil, errors.New("runs failed")))
	_, err := api.Runs(context.Background(), protocol.CronRunsParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNodesAPI_Invoke_Error(t *testing.T) {
	api := NewNodesAPI(mockRequest(nil, errors.New("invoke failed")))
	_, err := api.Invoke(context.Background(), protocol.NodeInvokeParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNodesAPI_Event_Error(t *testing.T) {
	api := NewNodesAPI(mockRequest(nil, errors.New("event failed")))
	err := api.Event(context.Background(), protocol.NodeEventParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNodesAPI_PendingDrain_Error(t *testing.T) {
	api := NewNodesAPI(mockRequest(nil, errors.New("drain failed")))
	_, err := api.PendingDrain(context.Background(), protocol.NodePendingDrainParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNodesAPI_PendingEnqueue_Error(t *testing.T) {
	api := NewNodesAPI(mockRequest(nil, errors.New("enqueue failed")))
	err := api.PendingEnqueue(context.Background(), protocol.NodePendingEnqueueParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNodesAPI_Pairing_Approve_Error(t *testing.T) {
	api := NewNodesAPI(mockRequest(nil, errors.New("approve failed")))
	err := api.PairingApprove(context.Background(), protocol.NodePairApproveParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNodesAPI_Pairing_Reject_Error(t *testing.T) {
	api := NewNodesAPI(mockRequest(nil, errors.New("reject failed")))
	err := api.PairingReject(context.Background(), protocol.NodePairRejectParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNodesAPI_Pairing_Verify_Error(t *testing.T) {
	api := NewNodesAPI(mockRequest(nil, errors.New("verify failed")))
	err := api.PairingVerify(context.Background(), protocol.NodePairVerifyParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSkillsAPI_Status_Error(t *testing.T) {
	api := NewSkillsAPI(mockRequest(nil, errors.New("status failed")))
	_, err := api.Status(context.Background(), protocol.SkillsStatusParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSkillsAPI_Install_Error(t *testing.T) {
	api := NewSkillsAPI(mockRequest(nil, errors.New("install failed")))
	err := api.Install(context.Background(), protocol.SkillsInstallParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSkillsAPI_Update_Error(t *testing.T) {
	api := NewSkillsAPI(mockRequest(nil, errors.New("update failed")))
	err := api.Update(context.Background(), protocol.SkillsUpdateParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestConfigAPI_Get_Error(t *testing.T) {
	api := NewConfigAPI(mockRequest(nil, errors.New("get failed")))
	_, err := api.Get(context.Background(), protocol.ConfigGetParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestConfigAPI_Set_Error(t *testing.T) {
	api := NewConfigAPI(mockRequest(nil, errors.New("set failed")))
	err := api.Set(context.Background(), protocol.ConfigSetParams{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestConfigAPI_Apply_Error(t *testing.T) {
	api := NewConfigAPI(mockRequest(nil, errors.New("apply failed")))
	err := api.Apply(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestConfigAPI_Patch_Error(t *testing.T) {
	api := NewConfigAPI(mockRequest(nil, errors.New("patch failed")))
	err := api.Patch(context.Background(), protocol.ConfigPatchParams{})
	if err == nil {
		t.Fatal("expected error")
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
