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

func TestProviderBackendCLIForwardsSDKOptions(t *testing.T) {
	type recorded struct {
		path string
		body map[string]any
	}
	var calls []recorded
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		decoder := json.NewDecoder(r.Body)
		decoder.UseNumber()
		_ = decoder.Decode(&body)
		calls = append(calls, recorded{r.URL.Path, body})
		if r.Header.Get("X-Workspace-ID") != "workspace-1" || r.Header.Get("X-API-Key") != "user-key" {
			t.Errorf("scope headers=%v", r.Header)
		}
		switch r.URL.Path {
		case "/newmoi/llm/backends/1/probe-models":
			_, _ = w.Write([]byte(`{"models":[]}`))
		case "/newmoi/llm/router":
			_, _ = w.Write([]byte(`{"strategy":"round_robin"}`))
		default:
			_, _ = w.Write([]byte(`{"id":1,"name":"backend"}`))
		}
	}))
	defer server.Close()
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewContext(client, "workspace-1", false, true, true)
	dir := t.TempDir()
	secret := filepath.Join(dir, "provider-key")
	if err := os.WriteFile(secret, []byte("provider-secret"), 0600); err != nil {
		t.Fatal(err)
	}
	empty := filepath.Join(dir, "empty-key")
	if err := os.WriteFile(empty, nil, 0600); err != nil {
		t.Fatal(err)
	}
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdout
	os.Stdout = write
	commands := [][]string{
		{"create", "--name", "backend", "--type", "openai", "--reasoning-control-protocol", "responses", "--api-key-file", secret, "--address", "https://provider.test", "--timeout-seconds", "0", "--models", "m1,m2", "--supported-mime-types", "application/pdf"},
		{"create", "--kind", "parser", "--name", "parser", "--supported-mime-types", "application/pdf,text/plain"},
		{"update", "--id", "1", "--name", "", "--api-key-file", empty, "--reasoning-control-protocol", "", "--timeout-seconds", "0", "--models", "m2", "--supported-mime-types", "text/plain"},
		{"update-router", "--strategy", "", "--health-check-interval", "0", "--max-retries", "0", "--session-affinity=false"},
		{"update-router"},
		{"probe-backend-models", "--backend-id", "1", "--api-key-file", secret},
	}
	for _, args := range commands {
		if err := ExecuteProviderBackendCommand(ctx, args); err != nil {
			os.Stdout = original
			t.Fatalf("%v: %v", args, err)
		}
	}
	os.Stdout = original
	_ = write.Close()
	output, _ := io.ReadAll(read)
	_ = read.Close()
	if bytes.Contains(output, []byte("provider-secret")) {
		t.Fatal("provider secret leaked to command output")
	}
	want := []recorded{
		{"/newmoi/llm/backends", map[string]any{"name": "backend", "type": "openai", "reasoning_control_protocol": "responses", "api_key": "provider-secret", "address": "https://provider.test", "timeout_seconds": json.Number("0"), "models": []any{"m1", "m2"}}},
		{"/newmoi/parser/backends", map[string]any{"name": "parser", "supported_mime_types": []any{"application/pdf", "text/plain"}}},
		{"/newmoi/llm/backends/1", map[string]any{"name": "", "api_key": "", "reasoning_control_protocol": "", "timeout_seconds": json.Number("0"), "models": []any{"m2"}, "supported_mime_types": []any{"text/plain"}}},
		{"/newmoi/llm/router", map[string]any{"strategy": "", "health_check_interval": json.Number("0"), "max_retries": json.Number("0"), "session_affinity": false}},
		{"/newmoi/llm/router", map[string]any{}},
		{"/newmoi/llm/backends/1/probe-models", map[string]any{"api_key": "provider-secret"}},
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls=%#v\nwant=%#v", calls, want)
	}
}

func TestModelSetupAndProbeCLIForwardProviderOptions(t *testing.T) {
	type recorded struct {
		path string
		body map[string]any
	}
	var calls []recorded
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		decoder := json.NewDecoder(r.Body)
		decoder.UseNumber()
		_ = decoder.Decode(&body)
		calls = append(calls, recorded{r.URL.Path, body})
		if r.URL.Path == "/newmoi/llm/setup" {
			_, _ = w.Write([]byte(`{"backend":{"id":1},"endpoint":{"id":2}}`))
			return
		}
		_, _ = w.Write([]byte(`{"models":[]}`))
	}))
	defer server.Close()
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewContext(client, "workspace-1", false, true, true)
	secret := filepath.Join(t.TempDir(), "provider-key")
	if err := os.WriteFile(secret, []byte("provider-secret"), 0600); err != nil {
		t.Fatal(err)
	}
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdout
	os.Stdout = write
	commands := [][]string{
		{"setup-llm-backend", "--name", "backend", "--address", "https://provider.test", "--models", "m1,m2", "--type", "openai", "--reasoning-control-protocol", "responses", "--api-key-file", secret, "--timeout-seconds", "0"},
		{"probe-llm-models", "--address", "https://provider.test", "--api-key-file", secret},
		{"probe-embedding-models", "--address", "https://embedding.test", "--api-key-file", secret},
	}
	for _, args := range commands {
		if err := ExecuteModelCommand(ctx, args); err != nil {
			os.Stdout = original
			t.Fatalf("%v: %v", args, err)
		}
	}
	os.Stdout = original
	_ = write.Close()
	_, _ = io.ReadAll(read)
	_ = read.Close()
	want := []recorded{
		{"/newmoi/llm/setup", map[string]any{"name": "backend", "address": "https://provider.test", "models": []any{"m1", "m2"}, "type": "openai", "reasoning_control_protocol": "responses", "api_key": "provider-secret", "timeout_seconds": json.Number("0")}},
		{"/newmoi/llm/probe-models", map[string]any{"address": "https://provider.test", "api_key": "provider-secret"}},
		{"/newmoi/embedding/probe-models", map[string]any{"address": "https://embedding.test", "api_key": "provider-secret"}},
	}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls=%#v want=%#v", calls, want)
	}
}
