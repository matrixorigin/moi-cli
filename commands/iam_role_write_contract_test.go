package commands

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestIAMUpdateRoleCommandPreservesAggregateContract(t *testing.T) {
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.EscapedPath() != "/newmoi/workspaces/ws%2F1/iam/roles/role%2F1" {
			t.Errorf("request = %s %s", r.Method, r.URL.EscapedPath())
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "OK", "msg": "OK", "data": map[string]any{
			"role_id": "role/1", "role_version": 4, "policy_version": 8, "data_privilege_version": 5,
			"provisioning_state": "ready", "provisioning_version": 2, "provisioning_error_code": "",
			"idempotent_replay": true, "session_invalidation_status": "succeeded", "request_id": "response-request", "trace_id": "response-trace",
		}})
	}))
	t.Cleanup(server.Close)
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewContext(client, "ws-default", false, true, true)
	output := captureStdout(t, func() {
		if err := ExecuteIAMCommand(ctx, []string{"update-role", "--workspace-id", "ws/1", "--role-id", "role/1", "--json", `{
			"name":"analyst","description":"updated role","schema_version":"v1","request_id":"update-1",
			"expected_lifecycle_version":3,"expected_policy_version":7,"expected_data_privilege_version":4,
			"statements":[{"action_id":"workflow.run","resource_type":"workflow","scope_mode":"all"}],
			"data_privilege_mutations":[{"operation":"grant","action_id":"table.query","resource_type":"table","resource_id":"41","privilege":"select"}]
		}`}); err != nil {
			t.Fatal(err)
		}
	})
	if body["expected_lifecycle_version"] != float64(3) || body["expected_policy_version"] != float64(7) || body["expected_data_privilege_version"] != float64(4) || body["schema_version"] != "v1" || body["request_id"] != "update-1" {
		t.Fatalf("request body = %#v", body)
	}
	if statements, ok := body["statements"].([]any); !ok || len(statements) != 1 || statements[0].(map[string]any)["action_id"] != "workflow.run" {
		t.Fatalf("statements = %#v", body["statements"])
	}
	if mutations, ok := body["data_privilege_mutations"].([]any); !ok || len(mutations) != 1 || mutations[0].(map[string]any)["resource_id"] != "41" {
		t.Fatalf("data mutations = %#v", body["data_privilege_mutations"])
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatal(err)
	}
	if result["role_id"] != "role/1" || result["role_version"] != float64(4) || result["policy_version"] != float64(8) || result["data_privilege_version"] != float64(5) || result["provisioning_state"] != "ready" || result["provisioning_version"] != float64(2) || result["idempotent_replay"] != true || result["session_invalidation_status"] != "succeeded" || result["request_id"] != "response-request" || result["trace_id"] != "response-trace" {
		t.Fatalf("printed role write result = %s", output)
	}
}

func TestIAMUpdateRoleCommandPreservesStoppedFailure(t *testing.T) {
	const message = "The role change was not completed and the operation has stopped. Contact an administrator before continuing."
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-ID", "public-request")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "ErrIAMRoleWriteFailed", "msg": message})
	}))
	t.Cleanup(server.Close)
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
	if err != nil {
		t.Fatal(err)
	}
	err = ExecuteIAMCommand(NewContext(client, "ws-1", false, true, true), []string{"update-role", "--role-id", "role-1", "--json", `{"request_id":"update-1"}`})
	var public *sdk.Error
	if !errors.As(err, &public) || public.Status != 500 || public.Code != "ErrIAMRoleWriteFailed" || public.Message != message || public.RequestID != "public-request" {
		t.Fatalf("CLI error = %v", err)
	}
}
