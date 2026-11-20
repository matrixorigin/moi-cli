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

func TestRuntimeGatewayCLIComputeEventAuthorizationAndExecutorHealth(t *testing.T) {
	type observedRequest struct {
		path, authorization, apiKey, workspace string
		body                                   map[string]any
	}
	var requests []observedRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		if len(raw) > 0 {
			decoder := json.NewDecoder(bytes.NewReader(raw))
			decoder.UseNumber()
			_ = decoder.Decode(&body)
		}
		requests = append(requests, observedRequest{r.URL.Path, r.Header.Get("Authorization"), r.Header.Get("X-API-Key"), r.Header.Get("X-Workspace-ID"), body})
		switch r.URL.Path {
		case "/api/v1/internal/compute-bindings/events":
			_, _ = w.Write([]byte(`{"binding":{"compute_binding_id":"binding-1"},"duplicate":false}`))
		case "/api/v1/runtime-executors/authorize":
			w.WriteHeader(http.StatusNoContent)
		case "/newmoi/agent-executors/health":
			_, _ = w.Write([]byte(`{"code":"OK","data":{"executors":[{"id":"runner-1","connected":true,"connected_secs":9007199254740993,"ready":true}],"astra_reachable":true}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewContext(client, "workspace-1", false, true, true)
	ctx.Endpoint = server.URL + "/newmoi"
	grant := filepath.Join(t.TempDir(), "grant")
	if err := os.WriteFile(grant, []byte("runtime-grant\n"), 0600); err != nil {
		t.Fatal(err)
	}
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdout
	os.Stdout = write
	event := `{"compute_binding_id":"binding-1","event_type":"activation_requested","workspace_id":"workspace-1","subject_type":"agent","subject_id":"agent-1","edge_agent_id":"edge-1","compute_resource_id":"compute-1","spec_id":"spec-1","placement_profile":"default","units":2,"lifecycle_instance_id":"lifecycle-1","fencing_token":9007199254740993,"kind":"sandbox","idempotency_key":"event-1","occurred_at":"2026-11-10T01:02:03Z"}`
	applyErr := ExecuteRuntimeGatewayCommand(ctx, []string{"apply-compute-binding-event", "--runtime-token-file", grant, "--json", event})
	authorizeErr := ExecuteRuntimeGatewayCommand(ctx, []string{"authorize-executor", "--runtime-token-file", grant, "--json", `{"task_id":"task-1","executor_id":"executor-1","run_id":"run-1","turn_chain_id":"turn-1","tool_call_id":"tool-1"}`})
	healthErr := ExecuteAgentCommand(ctx, []string{"executor-health"})
	os.Stdout = original
	_ = write.Close()
	output, _ := io.ReadAll(read)
	_ = read.Close()
	if applyErr != nil || authorizeErr != nil || healthErr != nil {
		t.Fatalf("apply=%v authorize=%v health=%v", applyErr, authorizeErr, healthErr)
	}
	wantEvent := map[string]any{"compute_binding_id": "binding-1", "event_type": "activation_requested", "workspace_id": "workspace-1", "subject_type": "agent", "subject_id": "agent-1", "edge_agent_id": "edge-1", "compute_resource_id": "compute-1", "spec_id": "spec-1", "placement_profile": "default", "units": json.Number("2"), "lifecycle_instance_id": "lifecycle-1", "fencing_token": json.Number("9007199254740993"), "kind": "sandbox", "idempotency_key": "event-1", "occurred_at": "2026-11-10T01:02:03Z"}
	if !reflect.DeepEqual(requests[0].body, wantEvent) {
		t.Fatalf("event body=%#v", requests[0].body)
	}
	if got, ok := requests[0].body["fencing_token"].(json.Number); !ok || got.String() != "9007199254740993" {
		t.Fatalf("fencing_token=%#v", requests[0].body["fencing_token"])
	}
	if requests[0].authorization != "Bearer runtime-grant" || requests[1].authorization != "Bearer runtime-grant" || requests[2].apiKey != "user-key" || requests[2].workspace != "workspace-1" {
		t.Fatalf("requests=%#v", requests)
	}
	decoder := json.NewDecoder(bytes.NewReader(output))
	decoder.UseNumber()
	var applied, health map[string]any
	if err := decoder.Decode(&applied); err != nil {
		t.Fatal(err)
	}
	if err := decoder.Decode(&health); err != nil {
		t.Fatal(err)
	}
	if applied["duplicate"] != false || health["astra_reachable"] != true {
		t.Fatalf("output=%s", output)
	}
}

func TestRuntimeGatewayCLIEdgeModelRPCAndStream(t *testing.T) {
	type observed struct {
		path, authorization string
		body                map[string]any
	}
	var requests []observed
	stream := []byte("data: {\"choices\":[]}\n\ndata: [DONE]\n\n")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.UseNumber()
		_ = decoder.Decode(&body)
		requests = append(requests, observed{r.URL.Path, r.Header.Get("Authorization"), body})
		switch r.URL.Path {
		case "/api/v1/astra/edge-tokens/check":
			_, _ = w.Write([]byte(`{"code":0,"data":{"ok":true,"subject":"subject-1","workspace_id":"workspace-1","edge_agent_id":"edge-1","expires_at":"2026-11-10T02:00:00Z"}}`))
		case "/newmoi/edge-tokens/renew":
			_, _ = w.Write([]byte(`{"code":0,"data":{"token":"fresh-edge","expires_at":"2026-11-10T03:00:00Z"}}`))
		case "/api/v1/models/resolve":
			_, _ = w.Write([]byte(`{"model":"model-1","status":"ready","protocol":"openai","invoke":{"url":"http://model.test/v1"}}`))
		case "/api/v1/mcp/http":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"rpc-1","result":{"tools":[]}}`))
		case "/api/v1/models/openai/chat/completions":
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = w.Write(stream)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewContext(client, "workspace-1", false, true, true)
	ctx.Endpoint = server.URL + "/newmoi"
	grant := filepath.Join(t.TempDir(), "grant")
	if err := os.WriteFile(grant, []byte("runtime-grant"), 0600); err != nil {
		t.Fatal(err)
	}
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdout
	os.Stdout = write
	calls := [][]string{
		{"check-edge-token", "--runtime-token-file", grant, "--token", "edge-registration"},
		{"renew-edge-token", "--runtime-token-file", grant, "--token", "edge-registration"},
		{"resolve-model", "--runtime-token-file", grant, "--json", `{"model":"model-1","gateway":"control-plane"}`},
		{"mcp", "--runtime-token-file", grant, "--json", `{"jsonrpc":"2.0","id":"rpc-1","method":"tools/list","params":{"cursor":"next"}}`},
	}
	for _, args := range calls {
		if err := ExecuteRuntimeGatewayCommand(ctx, args); err != nil {
			os.Stdout = original
			t.Fatalf("%s: %v", args[0], err)
		}
	}
	outputPath := filepath.Join(t.TempDir(), "model.sse")
	invokeErr := ExecuteRuntimeGatewayCommand(ctx, []string{"invoke-model", "--runtime-token-file", grant, "--output", outputPath, "--json", `{"model":"model-1","messages":[{"role":"user","content":"hello"}],"stream":true}`})
	os.Stdout = original
	_ = write.Close()
	output, _ := io.ReadAll(read)
	_ = read.Close()
	if invokeErr != nil {
		t.Fatal(invokeErr)
	}
	content, err := os.ReadFile(outputPath)
	if err != nil || !bytes.Equal(content, stream) {
		t.Fatalf("stream=%q error=%v", content, err)
	}
	wantPaths := []string{"/api/v1/astra/edge-tokens/check", "/newmoi/edge-tokens/renew", "/api/v1/models/resolve", "/api/v1/mcp/http", "/api/v1/models/openai/chat/completions"}
	for index, path := range wantPaths {
		if requests[index].path != path || requests[index].authorization != "Bearer runtime-grant" {
			t.Fatalf("request[%d]=%#v", index, requests[index])
		}
	}
	wantBodies := []map[string]any{
		{"token": "edge-registration"}, {"token": "edge-registration"}, {"model": "model-1", "gateway": "control-plane"},
		{"jsonrpc": "2.0", "id": "rpc-1", "method": "tools/list", "params": map[string]any{"cursor": "next"}},
		{"model": "model-1", "messages": []any{map[string]any{"role": "user", "content": "hello"}}, "stream": true},
	}
	for index, want := range wantBodies {
		if !reflect.DeepEqual(requests[index].body, want) {
			t.Fatalf("body[%d]=%#v want=%#v", index, requests[index].body, want)
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(output))
	decoder.UseNumber()
	for index := range calls {
		var result map[string]any
		if err := decoder.Decode(&result); err != nil {
			t.Fatalf("output[%d]=%s error=%v", index, output, err)
		}
	}
	var extra map[string]any
	if err := decoder.Decode(&extra); err != io.EOF {
		t.Fatalf("unexpected extra output=%s error=%v", output, err)
	}
}
