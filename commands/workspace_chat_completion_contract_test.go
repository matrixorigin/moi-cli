package commands

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestWorkspaceChatCompletionCLISharedContract(t *testing.T) {
	data, err := os.ReadFile("../testdata/product-contracts/workspace_chat_completion_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Input, Expected, Response json.RawMessage
		Query                     map[string]string
		Stream                    string
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	var expected map[string]any
	decoder := json.NewDecoder(bytes.NewReader(fixture.Expected))
	decoder.UseNumber()
	_ = decoder.Decode(&expected)
	var bodies []map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/newmoi/workspaces/workspace-1/llm/chat/completions" || r.URL.Query().Get("service_slot") != "slot/1" || r.Header.Get("X-Workspace-ID") != "workspace-1" {
			t.Errorf("request=%s headers=%v", r.URL, r.Header)
		}
		var body map[string]any
		decoder := json.NewDecoder(r.Body)
		decoder.UseNumber()
		_ = decoder.Decode(&body)
		bodies = append(bodies, body)
		if len(bodies) == 1 {
			_, _ = w.Write(fixture.Response)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(fixture.Stream))
	}))
	defer server.Close()
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewContext(client, "workspace-1", false, true, true)
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdout
	os.Stdout = write
	nonstreamErr := ExecuteModelCommand(ctx, []string{"workspace-chat-completion", "--service-slot", "slot/1", "--json", string(fixture.Input)})
	os.Stdout = original
	_ = write.Close()
	output, _ := io.ReadAll(read)
	_ = read.Close()
	if nonstreamErr != nil {
		t.Fatal(nonstreamErr)
	}
	var result map[string]any
	if err := json.Unmarshal(output, &result); err != nil || result["id"] != "chat-1" {
		t.Fatalf("output=%s error=%v", output, err)
	}
	streamPath := filepath.Join(t.TempDir(), "chat.sse")
	if err := ExecuteModelCommand(ctx, []string{"workspace-chat-completion-stream", "--service-slot", "slot/1", "--output", streamPath, "--json", string(fixture.Input)}); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(streamPath)
	if err != nil || string(content) != fixture.Stream {
		t.Fatalf("stream=%q error=%v", content, err)
	}
	streamExpected := map[string]any{}
	for key, value := range expected {
		streamExpected[key] = value
	}
	streamExpected["stream"] = true
	if !reflect.DeepEqual(bodies[0], expected) || !reflect.DeepEqual(bodies[1], streamExpected) {
		t.Fatalf("bodies=%#v expected=%#v", bodies, []map[string]any{expected, streamExpected})
	}
}
