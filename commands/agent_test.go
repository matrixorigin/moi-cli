package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestAgentCardRequiresSelector(t *testing.T) {
	t.Parallel()

	ctx := &Context{WorkspaceID: "ws-1"}
	err := ExecuteAgentCommand(ctx, []string{"agent-card"})
	if err == nil {
		t.Fatal("agent-card without selector succeeded")
	}
	if !strings.Contains(err.Error(), "--agent-code or --agent-id is required") {
		t.Fatalf("error = %v, want selector requirement", err)
	}
}

func TestAgentCardForwardsSelectorQuery(t *testing.T) {
	t.Parallel()

	var got http.Request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = *r
		got.URL = r.URL
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": "OK",
			"data": map[string]any{"name": "Explore", "url": "https://example.invalid/a2a"},
		})
	}))
	t.Cleanup(server.Close)

	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewContext(client, "ws-caller", false, true, true)
	if err := ExecuteAgentCommand(ctx, []string{
		"agent-card",
		"--agent-code", "explore",
		"--agent-workspace-id", "system",
	}); err != nil {
		t.Fatalf("agent-card: %v", err)
	}

	if got.Method != http.MethodGet {
		t.Errorf("method = %s, want GET", got.Method)
	}
	if got.URL.EscapedPath() != "/newmoi/agents/card" {
		t.Errorf("path = %q", got.URL.EscapedPath())
	}
	if got.Header.Get("X-Workspace-ID") != "ws-caller" {
		t.Errorf("X-Workspace-ID = %q, want caller workspace", got.Header.Get("X-Workspace-ID"))
	}
	query := got.URL.Query()
	if query.Get("agent_code") != "explore" {
		t.Errorf("agent_code = %q", query.Get("agent_code"))
	}
	if query.Get("agent_workspace_id") != "system" {
		t.Errorf("agent_workspace_id = %q", query.Get("agent_workspace_id"))
	}
	if query.Has("agent_id") {
		t.Errorf("agent_id = %q, want omitted", query.Get("agent_id"))
	}
}

func TestAgentCardAliasForwardsAgentID(t *testing.T) {
	t.Parallel()

	var gotQuery string
	var gotWorkspace string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		gotWorkspace = r.Header.Get("X-Workspace-ID")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": "OK",
			"data": map[string]any{"name": "custom"},
		})
	}))
	t.Cleanup(server.Close)

	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewContext(client, "ws-default", false, true, true)
	if err := ExecuteAgentCommand(ctx, []string{
		"card",
		"--workspace-id", "ws-override",
		"--agent-id", "agt-1",
	}); err != nil {
		t.Fatalf("agent card alias: %v", err)
	}
	if gotWorkspace != "ws-override" {
		t.Errorf("X-Workspace-ID = %q, want --workspace-id override", gotWorkspace)
	}
	if gotQuery != "agent_id=agt-1" {
		t.Errorf("query = %q, want agent_id=agt-1", gotQuery)
	}
}

func TestAgentCatalogCardsForwardsPagination(t *testing.T) {
	t.Parallel()

	var got http.Request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = *r
		got.URL = r.URL
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": "OK",
			"data": map[string]any{"items": []any{}, "total": 0, "limit": 20, "offset": 5},
		})
	}))
	t.Cleanup(server.Close)

	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewContext(client, "ws-1", false, true, true)
	if err := ExecuteAgentCommand(ctx, []string{
		"list-agent-catalog-cards",
		"--limit", "20",
		"--offset", "5",
	}); err != nil {
		t.Fatalf("list-agent-catalog-cards: %v", err)
	}

	if got.Method != http.MethodGet || got.URL.EscapedPath() != "/newmoi/workspaces/ws-1/agent-cards" {
		t.Fatalf("request = %s %s", got.Method, got.URL.EscapedPath())
	}
	if got.URL.Query().Get("limit") != "20" || got.URL.Query().Get("offset") != "5" {
		t.Fatalf("query = %q", got.URL.RawQuery)
	}
}
