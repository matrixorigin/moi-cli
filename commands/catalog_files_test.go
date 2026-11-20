package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestCatalogFilesSendsNativeSelectionAndPagination(t *testing.T) {
	var got map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/newmoi/catalog/file/list" {
			t.Errorf("unexpected file-list request: %s %s", r.Method, r.URL.Path)
		}
		got = nil
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Error(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":"OK","data":{"list":[],"total":0}}`))
	}))
	t.Cleanup(server.Close)
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-user-key"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	ctx := NewContext(client, "workspace", false, true, true)
	captureStdout(t, func() {
		err = ExecuteCatalogCommand(ctx, []string{"files", "--volume-id", "30", "--parent-id", "", "--name", "report", "--extensions", "pdf,docx", "--page", "2", "--page-size", "20", "--order", "desc", "--order-by", "created_at"})
	})
	if err != nil {
		t.Fatal(err)
	}
	var want map[string]any
	if err := json.Unmarshal([]byte(`{"filters":[{"name":"file_ext","values":["pdf","docx"]},{"name":"file_name","values":["report"],"fuzzy":true},{"name":"parent_id","values":[""]},{"name":"volume_id","values":["30"]}],"page":2,"page_size":20,"order":"desc","order_by":"created_at"}`), &want); err != nil {
		t.Fatal(err)
	}
	filters, ok := got["filters"].([]any)
	if !ok {
		t.Fatalf("file filters are missing: %#v", got)
	}
	sort.Slice(filters, func(i, j int) bool {
		return filters[i].(map[string]any)["name"].(string) < filters[j].(map[string]any)["name"].(string)
	})
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("file-list request = %#v, want %#v", got, want)
	}
	captureStdout(t, func() { err = ExecuteCatalogCommand(ctx, []string{"files"}) })
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, map[string]any{}) {
		t.Fatalf("omitted options must preserve server defaults: %#v", got)
	}
}
