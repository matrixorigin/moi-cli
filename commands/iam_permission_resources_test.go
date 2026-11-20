package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestIAMPermissionResourceCommandsUseTypedRoutes(t *testing.T) {
	type capturedRequest struct {
		method string
		path   string
		query  string
		body   map[string]any
	}
	requests := make([]capturedRequest, 0, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := map[string]any{}
		if r.Body != nil && r.ContentLength != 0 {
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("decode request: %v", err)
				return
			}
		}
		requests = append(requests, capturedRequest{
			method: r.Method, path: r.URL.EscapedPath(), query: r.URL.RawQuery, body: body,
		})
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": "OK", "msg": "OK", "data": map[string]any{"items": []any{}},
		})
	}))
	t.Cleanup(server.Close)

	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewContext(client, "ws-default", false, true, true)
	listCommand := []string{
		"list-permission-resources", "--workspace-id", "ws/1", "--resource-type", "database",
		"--action-id", "database.grant.query", "--purpose", "data_privilege",
		"--parent-id", "catalog/1", "--cursor", "next", "--q", "sales", "--limit", "1",
	}
	if output := captureStdout(t, func() {
		if err := ExecuteIAMCommand(ctx, listCommand); err != nil {
			t.Fatal(err)
		}
	}); !strings.Contains(output, `"items"`) {
		t.Fatalf("list output = %q, want JSON items", output)
	}
	batchCommand := []string{
		"batch-describe-permission-resources", "--workspace-id", "ws/1",
		"--resource-type", "database", "--action-id", "database.grant.query",
		"--purpose", "data_privilege", "--resource-id", "database/1",
	}
	if output := captureStdout(t, func() {
		if err := ExecuteIAMCommand(ctx, batchCommand); err != nil {
			t.Fatal(err)
		}
	}); !strings.Contains(output, `"items"`) {
		t.Fatalf("batch output = %q, want JSON items", output)
	}

	if requests[0].method != http.MethodGet ||
		requests[0].path != "/newmoi/workspaces/ws%2F1/iam/permissions/resources" {
		t.Fatalf("list request = %#v", requests[0])
	}
	for _, query := range []string{
		"resource_type=database", "action_id=database.grant.query", "purpose=data_privilege",
		"parent_id=catalog%2F1", "cursor=next", "q=sales", "limit=1",
	} {
		if !strings.Contains(requests[0].query, query) {
			t.Fatalf("list query %q missing %q", requests[0].query, query)
		}
	}
	if requests[1].method != http.MethodPost ||
		requests[1].path != "/newmoi/workspaces/ws%2F1/iam/permissions/resources:batchDescribe" {
		t.Fatalf("batch request = %#v", requests[1])
	}
	if requests[1].body["resource_type"] != "database" ||
		requests[1].body["action_id"] != "database.grant.query" ||
		requests[1].body["purpose"] != "data_privilege" ||
		len(requests[1].body["resource_ids"].([]any)) != 1 ||
		requests[1].body["resource_ids"].([]any)[0] != "database/1" {
		t.Fatalf("batch body = %#v", requests[1].body)
	}
}
