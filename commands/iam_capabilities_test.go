package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestIAMCurrentPrincipalCapabilitiesCommand(t *testing.T) {
	for _, test := range []struct {
		name string
		args []string
		ids  []string
	}{
		{name: "full projection"},
		{name: "single", args: []string{"--capability-id", "action.role.create"}, ids: []string{"action.role.create"}},
		{name: "batch", args: []string{"--capability-id", "navigation.roles", "--capability-id", "action.role.create"}, ids: []string{"navigation.roles", "action.role.create"}},
		{name: "empty is sent for server validation", args: []string{"--capability-id="}, ids: []string{""}},
		{name: "duplicate is sent for server validation", args: []string{"--capability-id", "navigation.roles", "--capability-id", "navigation.roles"}, ids: []string{"navigation.roles", "navigation.roles"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			var received []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/newmoi/workspaces/ws-1/iam/me/capabilities" || r.Header.Get("X-MOI-Role-ID") != "7" {
					t.Errorf("request = %s %s, role=%q", r.Method, r.URL, r.Header.Get("X-MOI-Role-ID"))
				}
				received = r.URL.Query()["capability_id"]
				_, _ = w.Write([]byte(`{"code":"OK","msg":"OK","data":{"effective_role_id":"7","capability_version":"v1","request_id":"request-1","trace_id":"trace-1","items":[{"capability_id":"action.role.create","action_id":"role.create","resource_type":"workspace","allowed":false,"reason_code":"PERMISSION_DENIED","resource_ids":[]}]}}`))
			}))
			t.Cleanup(server.Close)
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("key"), sdk.WithWorkspaceCurrentRoleID("ws-1", "7"))
			if err != nil {
				t.Fatal(err)
			}
			ctx := NewContext(client, "ws-1", false, true, true)
			output := captureStdout(t, func() {
				if err := ExecuteIAMCommand(ctx, append([]string{"current-principal-capabilities"}, test.args...)); err != nil {
					t.Fatal(err)
				}
			})
			if !slices.Equal(received, test.ids) {
				t.Fatalf("capability IDs = %q, want %q", received, test.ids)
			}
			var result sdk.IAMCurrentPrincipalCapabilitiesResult
			if err := json.Unmarshal([]byte(output), &result); err != nil {
				t.Fatal(err)
			}
			if result.GetEffectiveRoleId() != "7" || result.GetCapabilityVersion() != "v1" || result.GetRequestId() != "request-1" || result.GetTraceId() != "trace-1" {
				t.Fatalf("output metadata = %s", output)
			}
			var ids []string
			for _, item := range result.GetItems() {
				ids = append(ids, item.GetCapabilityId())
				if item.GetAllowed() || item.GetReasonCode() != "PERMISSION_DENIED" {
					t.Fatalf("output denial = %s", output)
				}
			}
			if !slices.Equal(ids, []string{"action.role.create"}) {
				t.Fatalf("output IDs = %v", ids)
			}
		})
	}
}

func TestIAMCurrentPrincipalCapabilitiesCommandPreservesServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !slices.Equal(r.URL.Query()["capability_id"], []string{"unknown"}) {
			t.Errorf("selection = %v", r.URL.Query())
		}
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"code":"IAM_ERROR_CODE_PERMISSION_DENIED","msg":"Permission denied","data":null}`))
	}))
	t.Cleanup(server.Close)
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("key"))
	if err != nil {
		t.Fatal(err)
	}
	err = ExecuteIAMCommand(NewContext(client, "ws-1", false, true, true), []string{"current-principal-capabilities", "--capability-id", "unknown"})
	var productErr *sdk.Error
	if !errors.As(err, &productErr) || productErr.Status != http.StatusForbidden || productErr.Code != "IAM_ERROR_CODE_PERMISSION_DENIED" || productErr.Message != "Permission denied" {
		t.Fatalf("server error = %#v", err)
	}
}

func TestIAMCurrentPrincipalCapabilitiesCommandUnsupportedIDs(t *testing.T) {
	for _, test := range []struct {
		name    string
		args    []string
		items   string
		wantIDs []string
	}{
		{name: "mixed", args: []string{"--capability-id", "action.role.create"}, items: `[{"capability_id":"action.role.create","allowed":false,"reason_code":"PERMISSION_DENIED"}]`, wantIDs: []string{"action.role.create"}},
		{name: "all unsupported", items: `[]`},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, `{"code":"OK","msg":"OK","data":{"items":%s,"unsupported_capability_ids":["future.id"]}}`, test.items)
			}))
			t.Cleanup(server.Close)
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("key"))
			if err != nil {
				t.Fatal(err)
			}
			output := captureStdout(t, func() {
				args := append([]string{"current-principal-capabilities", "--capability-id", "future.id"}, test.args...)
				if err := ExecuteIAMCommand(NewContext(client, "ws-1", false, true, true), args); err != nil {
					t.Fatal(err)
				}
			})
			var result sdk.IAMCurrentPrincipalCapabilitiesResult
			if err := json.Unmarshal([]byte(output), &result); err != nil {
				t.Fatal(err)
			}
			var ids []string
			for _, item := range result.GetItems() {
				ids = append(ids, item.GetCapabilityId())
			}
			if !slices.Equal(ids, test.wantIDs) || !slices.Equal(result.GetUnsupportedCapabilityIds(), []string{"future.id"}) {
				t.Fatalf("partial capability output = %s", output)
			}
		})
	}
}
