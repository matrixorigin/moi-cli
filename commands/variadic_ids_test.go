package commands

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

// Each case drives a CLI command whose Product SDK method takes a variadic ID
// list and asserts that the IDs reach the request body (matrixflow#18266).
func TestVariadicIDFlagsReachRequestBody(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		run      func(*Context, []string) error
		args     []string
		path     string
		bodyKey  string
		wantBody any
	}{
		{"create-file", runCatalogServiceCreateFile, []string{"--volume-id", "2", "--file-ids", "f1, f2"}, "/newmoi/catalog/file/create", "file_ids", []any{"f1", "f2"}},
		{"delete-file-ref", runCatalogServiceDeleteFileRef, []string{"--volume-id", "2", "--file-ids", "f1"}, "/newmoi/catalog/file/delete_ref", "file_ids", []any{"f1"}},
		{"trigger-file-workflow", runCatalogServiceTriggerFileWorkflow, []string{"--volume-id", "2", "--file-ids", "f1,f2"}, "/newmoi/catalog/file/trigger_workflow", "file_ids", []any{"f1", "f2"}},
		{"batch-get-data-assets", runCatalogServiceBatchGetDataAssets, []string{"--raw-file-ids", "r1,r2"}, "/newmoi/catalog/data-assets/batch-get", "raw_file_ids", []any{"r1", "r2"}},
		{"batch-resolve-file-artifacts", runCatalogServiceBatchResolveFileArtifacts, []string{"--file-ids", "f1"}, "/newmoi/catalog/file-artifacts:batchResolve", "file_ids", []any{"f1"}},
		{"multi-table-info", runCatalogServiceMultiTableInfo, []string{"--table-ids", "7,9"}, "/newmoi/catalog/table/multi_info", "ids", []any{float64(7), float64(9)}},
		{"retry-task-files", runConnectorServiceRetryTaskFiles, []string{"--task-id", "t1", "--ids", "a,b"}, "/newmoi/task/retry", "ids", []any{"a", "b"}},
		{"retry-export-task", runConnectorServiceRetryExportTask, []string{"--task-id", "t1", "--file-ids", "f1"}, "/newmoi/export/task/t1/rerun", "files", []any{"f1"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var gotPath string
			var gotBody map[string]any
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.EscapedPath()
				raw, err := io.ReadAll(r.Body)
				if err != nil {
					t.Errorf("read body: %v", err)
				}
				if err := json.Unmarshal(raw, &gotBody); err != nil {
					t.Errorf("decode body: %v raw=%s", err, raw)
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"code": "OK", "msg": "OK", "data": map[string]any{}})
			}))
			t.Cleanup(server.Close)

			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
			if err != nil {
				t.Fatal(err)
			}
			ctx := NewContext(client, "ws-1", false, true, true)
			if err := tc.run(ctx, tc.args); err != nil {
				t.Fatalf("%s %v: %v", tc.name, tc.args, err)
			}
			if gotPath != tc.path {
				t.Fatalf("path = %q, want %q", gotPath, tc.path)
			}
			if got := gotBody[tc.bodyKey]; !reflect.DeepEqual(got, tc.wantBody) {
				t.Fatalf("body[%q] = %#v, want %#v (body=%v)", tc.bodyKey, got, tc.wantBody, gotBody)
			}
		})
	}
}

func TestCreateFileWithoutFileIDsReturnsSDKError(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected request %s", r.URL.Path)
	}))
	t.Cleanup(server.Close)
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewContext(client, "ws-1", false, true, true)
	err = runCatalogServiceCreateFile(ctx, []string{"--volume-id", "2"})
	if !errors.Is(err, sdk.ErrFileIDsRequired) {
		t.Fatalf("err = %v, want ErrFileIDsRequired", err)
	}
}

func TestMultiTableInfoRejectsNonIntegerTableID(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected request %s", r.URL.Path)
	}))
	t.Cleanup(server.Close)
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewContext(client, "ws-1", false, true, true)
	err = runCatalogServiceMultiTableInfo(ctx, []string{"--table-ids", "7,x"})
	if err == nil || !strings.Contains(err.Error(), "table-ids") {
		t.Fatalf("err = %v, want table-ids parse error", err)
	}
}
