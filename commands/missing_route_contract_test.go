package commands

import (
	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestMissingRouteCLIContracts(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/newmoi/workflow/v2/workflow-apps/workflow/versions/version":
			io.WriteString(w, `{"code":0,"data":{"ID":"version","Version":1}}`)
		case "/newmoi/workspaces/workspace/llm/sessions/1/messages/2/modify-response":
			if r.Method != "PUT" {
				t.Errorf("method=%s", r.Method)
			}
			w.WriteHeader(204)
		case "/api/v1/service-accounts/workspaces":
			if r.Header.Get("Authorization") != "Bearer provisioner" || r.Header.Get("X-UC-Actor-Assertion") != "actor" || r.Header.Get("X-Service-Account-Contract-Version") != "1" || r.Header.Get("X-API-Key") != "" {
				t.Error("provisioner identity was not isolated")
			}
			io.WriteString(w, `{"items":[]}`)
		case "/api/v1/mcp/http":
			if r.Header.Get("Authorization") != "Bearer grant" {
				t.Error("missing runtime grant")
			}
			w.WriteHeader(401)
			io.WriteString(w, `{"jsonrpc":"2.0","id":"rpc","error":{"code":-32001,"message":"Runtime admission denied"}}`)
		default:
			t.Errorf("unexpected path=%s", r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer server.Close()
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("key"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	ctx := NewContext(client, "workspace", false, true, true)
	ctx.Endpoint = server.URL + "/newmoi"
	ctx.APIKey = "key"
	file := func(name, body string) string {
		p := filepath.Join(t.TempDir(), name)
		if err := os.WriteFile(p, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
		return p
	}
	if err := ExecuteWorkflowCommand(ctx, []string{"version", "--workflow-id", "workflow", "--version-id", "version"}); err != nil {
		t.Fatal(err)
	}
	if err := ExecuteConversationCommand(ctx, []string{"modify-response", "--session-id", "1", "--message-id", "2", "--content", "edited"}); err != nil {
		t.Fatal(err)
	}
	if err := ExecuteSystemCommand(ctx, []string{"service-account-workspaces", "--provisioner-token-file", file("token", "provisioner"), "--actor-assertion-file", file("actor", "actor")}); err != nil {
		t.Fatal(err)
	}
	err = ExecuteRuntimeGatewayCommand(ctx, []string{"mcp", "--runtime-token-file", file("grant", "grant"), "--json", `{"jsonrpc":"2.0","id":"rpc","method":"tools/list","params":{}}`})
	if !sdk.IsCode(err, "-32001") {
		t.Fatalf("runtime owner code lost: %v", err)
	}
	if calls != 4 {
		t.Fatalf("calls=%d", calls)
	}
}
