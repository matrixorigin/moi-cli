package commands

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestModelWorkspaceChatStreamWritesExactSSE(t *testing.T) {
	const events = "data: {\"choices\":[{\"delta\":{\"content\":\"hello\"}}]}\n\ndata: [DONE]\n\n"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/newmoi/workspaces/workspace-1/llm/chat/completions" || r.URL.Query().Get("service_slot") != "agent_chat" || r.Header.Get("Authorization") != "Bearer user-token" || r.Header.Get("Accept") != "text/event-stream" {
			t.Errorf("stream request=%s %s headers=%v", r.Method, r.URL, r.Header)
		}
		var body struct {
			Stream bool   `json:"stream"`
			Model  string `json:"model"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if !body.Stream || body.Model != "integration-model" {
			t.Errorf("stream request body=%+v", body)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		if _, err := w.Write([]byte(events)); err != nil {
			t.Error(err)
		}
	}))
	t.Cleanup(server.Close)
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithToken("user-token"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Error(err)
		}
	})
	output := filepath.Join(t.TempDir(), "events.sse")
	err = ExecuteModelCommand(NewContext(client, "workspace-1", false, false, true), []string{
		"workspace-chat-completion-stream", "--service-slot", "agent_chat", "--output", output,
		"--json", `{"model":"integration-model","messages":[{"role":"user","content":"hello"}],"stream":false}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(output)
	if err != nil || string(got) != events {
		t.Fatalf("events=%q error=%v", got, err)
	}
}

func TestModelWorkspaceChatStreamPreservesPublicFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-ID", "chat-stream-denied")
		w.WriteHeader(http.StatusForbidden)
		if _, err := w.Write([]byte(`{"code":"ErrForbidden","msg":"无权调用模型"}`)); err != nil {
			t.Error(err)
		}
	}))
	t.Cleanup(server.Close)
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithToken("user-token"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Error(err)
		}
	})
	output := filepath.Join(t.TempDir(), "denied.sse")
	err = ExecuteModelCommand(NewContext(client, "workspace-1", false, false, true), []string{
		"workspace-chat-completion-stream", "--output", output, "--json", `{"messages":[{"role":"user","content":"hello"}]}`,
	})
	var failure *sdk.Error
	if !errors.As(err, &failure) || failure.Status != http.StatusForbidden || failure.Code != "ErrForbidden" || failure.Message != "无权调用模型" || failure.RequestID != "chat-stream-denied" {
		t.Fatalf("public failure=%#v", err)
	}
	if _, err := os.Stat(output); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("failed request created output: %v", err)
	}
}
