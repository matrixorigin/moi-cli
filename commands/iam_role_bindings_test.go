package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestIAMRoleBindingsCommandsUseCanonicalRoutes(t *testing.T) {
	var requests []struct {
		method string
		path   string
		body   map[string]any
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := map[string]any{}
		if r.Body != nil && r.ContentLength != 0 {
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("decode request: %v", err)
				return
			}
		}
		requests = append(requests, struct {
			method string
			path   string
			body   map[string]any
		}{r.Method, r.URL.EscapedPath(), body})
		data := map[string]any{
			"user_id":              "user/1",
			"assigned_role_ids":    []string{"role/viewer"},
			"default_role_id":      "role/viewer",
			"binding_version":      5,
			"default_role_version": 5,
			"audit_event_id":       "audit-1",
		}
		if r.Method == http.MethodGet {
			data = map[string]any{
				"user_id":              "user/1",
				"roles":                []any{map[string]any{"role_id": "role/viewer", "role_name": "Viewer", "selectable": true}},
				"default_role_id":      "role/viewer",
				"default_role_version": 4,
				"binding_version":      "binding-v4",
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "OK", "msg": "OK", "data": data})
	}))
	t.Cleanup(server.Close)

	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewContext(client, "ws-default", false, true, true)

	if output := captureStdout(t, func() {
		if err := ExecuteIAMCommand(ctx, []string{"user-role-bindings", "--workspace-id", "ws/1", "--user-id", "user/1"}); err != nil {
			t.Fatal(err)
		}
	}); !strings.Contains(output, `"binding_version": "binding-v4"`) {
		t.Fatalf("GET output = %s", output)
	}

	if err := ExecuteIAMCommand(ctx, []string{
		"put-user-role-bindings",
		"--workspace-id", "ws/1",
		"--user-id", "user/1",
		"--use-current-versions",
		"--json", `{"assigned_role_ids":["role/viewer"],"default_role_id":"role/viewer","request_id":"write-1"}`,
	}); err != nil {
		t.Fatal(err)
	}

	if len(requests) != 3 {
		t.Fatalf("requests = %#v, want GET, version GET, and PUT", requests)
	}
	if requests[0].method != http.MethodGet || requests[0].path != "/newmoi/workspaces/ws%2F1/iam/users/user%2F1/role-bindings" {
		t.Fatalf("GET request = %#v", requests[0])
	}
	if requests[1].method != http.MethodGet || requests[1].path != requests[0].path {
		t.Fatalf("version GET request = %#v", requests[1])
	}
	if requests[2].method != http.MethodPut || requests[2].path != requests[0].path {
		t.Fatalf("PUT request = %#v", requests[2])
	}
	if requests[2].body["expected_binding_version"] != float64(4) || requests[2].body["expected_default_role_version"] != float64(4) || requests[2].body["request_id"] != "write-1" {
		t.Fatalf("PUT body = %#v", requests[2].body)
	}
}

func TestLegacyWorkspaceMemberRoleCommandIsRejected(t *testing.T) {
	ctx := &Context{}
	err := ExecuteWorkspaceMemberCommand(ctx, []string{"set-roles"})
	if err == nil || !strings.Contains(err.Error(), "unknown workspace-member subcommand: set-roles") {
		t.Fatalf("set-roles error = %v", err)
	}
}
