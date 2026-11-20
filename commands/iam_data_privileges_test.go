package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestIAMDataPrivilegeCommandsUseTypedNativeRoutes(t *testing.T) {
	type capturedRequest struct {
		method string
		path   string
		body   map[string]any
	}
	requests := make([]capturedRequest, 0, 5)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := map[string]any{}
		if r.Body != nil && r.ContentLength != 0 {
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("decode request: %v", err)
				return
			}
		}
		path := r.URL.EscapedPath()
		requests = append(requests, capturedRequest{method: r.Method, path: path, body: body})
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": "OK", "msg": "OK", "data": iamDataPrivilegeCLIResponse(r.Method, path),
		})
	}))
	t.Cleanup(server.Close)

	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewContext(client, "ws-default", false, true, true)
	mutationJSON := `{
		"request_id":"request/1",
		"expected_data_privilege_version":4,
		"mutations":[{
			"operation":"grant",
			"action_id":"table.query",
			"resource_type":"table",
			"resource_id":"table/1",
			"privilege":"select",
			"row_column":{"mode":"set","draft":{"black_column_list":["secret"],"rule_list":[{"column":"id","relation":"and","expression_list":[{"operator":"=","expression":["1"]}]}]}}
		}]
	}`

	for _, command := range [][]string{
		{"data-privileges", "--workspace-id", "ws/1", "--role-id", "role/1"},
		{"validate-data-privileges", "--workspace-id", "ws/1", "--role-id", "role/1", "--json", mutationJSON},
		{"apply-data-privileges", "--workspace-id", "ws/1", "--role-id", "role/1", "--json", mutationJSON},
		{"data-privilege-operation", "--workspace-id", "ws/1", "--role-id", "role/1", "--request-id", "request/1"},
		{"repair-data-privileges", "--workspace-id", "ws/1", "--role-id", "role/1", "--repair-operation", "request/1:repair", "--json", `{"request_id":"repair-1","expected_data_privilege_version":5}`},
	} {
		if output := captureStdout(t, func() {
			if err := ExecuteIAMCommand(ctx, command); err != nil {
				t.Fatal(err)
			}
		}); !strings.Contains(output, `"`) {
			t.Fatalf("command %q output = %q, want JSON", command[0], output)
		}
	}

	if len(requests) != 5 {
		t.Fatalf("requests = %#v, want five typed SDK calls", requests)
	}
	base := "/newmoi/workspaces/ws%2F1/iam/roles/role%2F1"
	assertIAMDataPrivilegeCLIRequest(t, requests[0], http.MethodGet, base+"/data-privileges")
	assertIAMDataPrivilegeCLIRequest(t, requests[1], http.MethodPost, base+"/data-privileges:validate")
	assertIAMDataPrivilegeCLIRequest(t, requests[2], http.MethodPut, base+"/data-privileges")
	assertIAMDataPrivilegeCLIRequest(t, requests[3], http.MethodGet, base+"/data-privilege-operations/request%2F1")
	assertIAMDataPrivilegeCLIRequest(t, requests[4], http.MethodPost, base+"/data-privilege-operations/request%2F1:repair")

	for _, request := range []capturedRequest{requests[1], requests[2]} {
		if request.body["request_id"] != "request/1" || request.body["expected_data_privilege_version"] != float64(4) {
			t.Fatalf("mutation body = %#v", request.body)
		}
		mutations, ok := request.body["mutations"].([]any)
		if !ok || len(mutations) != 1 {
			t.Fatalf("mutation list = %#v", request.body["mutations"])
		}
		mutation, ok := mutations[0].(map[string]any)
		if !ok || mutation["resource_id"] != "table/1" || mutation["operation"] != "grant" {
			t.Fatalf("mutation = %#v", mutations[0])
		}
		rowColumn, ok := mutation["row_column"].(map[string]any)
		if !ok || rowColumn["mode"] != "set" {
			t.Fatalf("row_column = %#v", mutation["row_column"])
		}
	}
	if requests[4].body["request_id"] != "repair-1" || requests[4].body["expected_data_privilege_version"] != float64(5) {
		t.Fatalf("repair body = %#v", requests[4].body)
	}
}

func iamDataPrivilegeCLIResponse(method, path string) any {
	switch {
	case method == http.MethodGet && strings.HasSuffix(path, "/data-privileges"):
		return map[string]any{
			"target_role_id": "role/1", "data_privilege_version": 4, "items": []any{},
		}
	case method == http.MethodPost && strings.HasSuffix(path, "/data-privileges:validate"):
		return map[string]any{"valid": true, "errors": []any{}, "warnings": []any{}}
	default:
		return map[string]any{
			"request_id": "request/1", "operation_id": "operation-1", "state": "succeeded", "data_privilege_version": 5, "resource_results": []any{},
		}
	}
}

func assertIAMDataPrivilegeCLIRequest(t *testing.T, request struct {
	method string
	path   string
	body   map[string]any
}, method, path string) {
	t.Helper()
	if request.method != method || request.path != path {
		t.Fatalf("request = %s %s, want %s %s", request.method, request.path, method, path)
	}
}
