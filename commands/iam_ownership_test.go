package commands

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestIAMRoleDeleteOwnershipsCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.EscapedPath() != "/newmoi/workspaces/ws%2F1/iam/roles/role%2F1/delete-ownerships" {
			t.Errorf("request = %s %s", r.Method, r.URL.EscapedPath())
		}
		query := r.URL.Query()
		if query.Get("cursor") != "next/1" || query.Get("limit") != "10" || query.Get("transfer_requested") != "true" {
			t.Errorf("query = %v", query)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "OK", "msg": "OK", "data": map[string]any{
			"items":       []any{map[string]any{"resource_type": "connector", "resource_id": "object-1", "ownership_version": 3}},
			"next_cursor": "next/2", "ownership_version": "snapshot-1",
		}})
	}))
	t.Cleanup(server.Close)
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
	if err != nil {
		t.Fatal(err)
	}
	output := captureStdout(t, func() {
		err = ExecuteIAMCommand(NewContext(client, "ws-default", false, true, true), []string{
			"role-delete-ownerships", "--workspace-id", "ws/1", "--role-id", "role/1",
			"--cursor", "next/1", "--limit", "10", "--transfer-requested",
		})
	})
	if err != nil {
		t.Fatal(err)
	}
	var result sdk.IAMRoleDeleteOwnershipList
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatal(err)
	}
	if result.GetNextCursor() != "next/2" || result.GetOwnershipVersion() != "snapshot-1" || len(result.GetItems()) != 1 || result.GetItems()[0].GetResourceId() != "object-1" {
		t.Fatalf("printed page = %s", output)
	}
}

func TestIAMTransferObjectOwnershipCommand(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.EscapedPath() != "/newmoi/workspaces/ws%2F1/iam/permissions/objects/connector/object%2F1/owner" {
			t.Errorf("request = %s %s", r.Method, r.URL.EscapedPath())
		}
		var request sdk.IAMObjectOwnershipTransferRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}
		if request.GetTargetRoleId() != "receiver" || request.GetRequestId() != "transfer-1" || request.GetExpectedOwnershipVersion() != 3 {
			t.Errorf("request body = %v", &request)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "OK", "msg": "OK", "data": map[string]any{
			"resource_type": "connector", "resource_id": "object/1", "previous_owner_role_id": "source",
			"owner_role_id": "receiver", "ownership_version": 4, "request_id": "transfer-1", "idempotent_replay": true,
		}})
	}))
	t.Cleanup(server.Close)
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
	if err != nil {
		t.Fatal(err)
	}
	output := captureStdout(t, func() {
		err = ExecuteIAMCommand(NewContext(client, "ws-default", false, true, true), []string{
			"transfer-object-ownership", "--workspace-id", "ws/1", "--resource-type", "connector", "--resource-id", "object/1",
			"--json", `{"target_role_id":"receiver","expected_ownership_version":3,"request_id":"transfer-1"}`,
		})
	})
	if err != nil {
		t.Fatal(err)
	}
	var result sdk.IAMObjectOwnershipTransferResult
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatal(err)
	}
	if result.GetOwnerRoleId() != "receiver" || result.GetOwnershipVersion() != 4 || result.GetRequestId() != "transfer-1" || !result.GetIdempotentReplay() {
		t.Fatalf("printed receipt = %s", output)
	}
}

func TestIAMTransferObjectOwnershipCommandPreservesConflict(t *testing.T) {
	const message = "对象所有权已变化，请刷新后重试。"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-ID", "public-request")
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "ErrIAMObjectGrantConflict", "msg": message})
	}))
	t.Cleanup(server.Close)
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
	if err != nil {
		t.Fatal(err)
	}
	err = ExecuteIAMCommand(NewContext(client, "ws-1", false, true, true), []string{
		"transfer-object-ownership", "--resource-type", "connector", "--resource-id", "object-1", "--json", `{"request_id":"transfer-1"}`,
	})
	var public *sdk.Error
	if !errors.As(err, &public) || public.Status != http.StatusConflict || public.Code != "ErrIAMObjectGrantConflict" || public.Message != message || public.RequestID != "public-request" {
		t.Fatalf("CLI error = %v", err)
	}
}
