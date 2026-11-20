package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestCatalogAllowedActionsCommand(t *testing.T) {
	var ids []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/newmoi/catalog/allowed-actions" || r.Header.Get("X-Workspace-ID") != "ws-test" {
			t.Errorf("unexpected candidate request: %s %s, workspace=%s", r.Method, r.URL.Path, r.Header.Get("X-Workspace-ID"))
		}
		var body struct {
			ResourceType string   `json:"resource_type"`
			IDs          []string `json:"resource_ids"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if body.ResourceType != "catalog" {
			t.Errorf("resource_type = %s", body.ResourceType)
		}
		ids = body.IDs
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "OK", "data": map[string]any{"items": []any{}}})
	}))
	t.Cleanup(server.Close)
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("key"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewContext(client, "ws-default", false, true, true)
	captureStdout(t, func() {
		if err := ExecuteCatalogCommand(ctx, []string{"allowed-actions", "--workspace-id", "ws-test", "--resource-type", "catalog", "--resource-ids", "10,11"}); err != nil {
			t.Fatal(err)
		}
	})
	if !slices.Equal(ids, []string{"10", "11"}) {
		t.Fatalf("candidate ids = %v", ids)
	}
}
