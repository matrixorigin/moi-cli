package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestComputeAdoptLegacyOwnershipUsesScopedProductRoute(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.EscapedPath() != "/newmoi/compute-resources/cr-1/adopt-legacy-ownership" {
			t.Errorf("path = %q", r.URL.EscapedPath())
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": "OK",
			"data": map[string]any{
				"workspace_id":      "ws-1",
				"resource_type":     "compute_resource",
				"resource_id":       "cr-1",
				"owner_role_id":     "superadmin-role",
				"status":            "active",
				"ownership_version": 1,
			},
		})
	}))
	t.Cleanup(server.Close)

	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewContext(client, "ws-1", false, true, true)
	if err := ExecuteComputeCommand(ctx, []string{"adopt-legacy-ownership", "--id", "cr-1"}); err != nil {
		t.Fatalf("compute adoption: %v", err)
	}
}

func TestComputeAdoptLegacyOwnershipRequiresResourceID(t *testing.T) {
	ctx := &Context{WorkspaceID: "ws-1"}
	if err := ExecuteComputeCommand(ctx, []string{"adopt-legacy-ownership"}); err == nil {
		t.Fatal("compute adoption without --id succeeded")
	}
}

func TestComputeAdminUpdateUsesAdminProductRoute(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %s, want PUT", r.Method)
		}
		if r.URL.EscapedPath() != "/newmoi/admin/compute-resources/cr-1" {
			t.Errorf("path = %q", r.URL.EscapedPath())
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		for key, want := range map[string]any{
			"name":         "renamed",
			"description":  "updated description",
			"cpu":          float64(4),
			"memory_gib":   float64(16),
			"min_replicas": float64(2),
		} {
			if body[key] != want {
				t.Errorf("body[%q] = %#v, want %#v", key, body[key], want)
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": "OK",
			"data": map[string]any{
				"id":     "cr-1",
				"status": "active",
			},
		})
	}))
	t.Cleanup(server.Close)

	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewContext(client, "ws-1", false, true, true)
	if err := ExecuteComputeCommand(ctx, []string{
		"admin-update",
		"--id", "cr-1",
		"--name", "renamed",
		"--description", "updated description",
		"--cpu", "4",
		"--memory-gib", "16",
		"--min-replicas", "2",
	}); err != nil {
		t.Fatalf("compute admin update: %v", err)
	}
}

func TestComputeAdminUpdateRequiresResourceID(t *testing.T) {
	ctx := &Context{WorkspaceID: "ws-1"}
	if err := ExecuteComputeCommand(ctx, []string{"admin-update"}); err == nil {
		t.Fatal("compute admin update without --id succeeded")
	}
}

func TestComputeAdminUpdateRequiresChanges(t *testing.T) {
	ctx := &Context{WorkspaceID: "ws-1"}
	if err := ExecuteComputeCommand(ctx, []string{"admin-update", "--id", "cr-1"}); err == nil {
		t.Fatal("compute admin update without update options succeeded")
	}
}
