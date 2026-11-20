package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

// Go already owns these options and their product scenarios. Compare the CLI
// request and output with that public SDK, without duplicating business logic.
func TestKnowledgeCLIOptionsAlignWithGoSDK(t *testing.T) {
	tables := json.RawMessage(`[{"db_name":"sales","table_names":["orders"]}]`)
	files := json.RawMessage(`{"file_ids":[],"parents":[],"vector_table":"kb_index","embedding_model":"bge-m3"}`)
	selections := []sdk.SemanticModelSourceSelectionInput{
		{Kind: "database_tables", DatabaseID: 17, AllSelected: true, ExcludedTableIDs: []int64{19}, Filters: sdk.SemanticModelSourceSelectionFilters{TableName: "orders"}},
		{Kind: "volume_files", VolumeID: 23, SelectedFileIDs: []string{"file-1"}, Filters: sdk.SemanticModelSourceSelectionFilters{FileName: "政策", FileExt: []string{"pdf"}}},
	}
	selectionJSON, err := json.Marshal(selections)
	if err != nil {
		t.Fatal(err)
	}
	for _, scenario := range []struct {
		name     string
		args     []string
		inputs   map[string][]byte
		method   string
		path     string
		response string
		invokeGo func(*sdk.KnowledgeService) (any, error)
	}{
		{
			name: "create alias keeps model output", args: []string{"create", "--name", "资料库", "--description", "说明", "--image-index-enabled=true"},
			method: http.MethodPost, path: "/semantic-models/create-empty",
			response: `{"model":{"id":101,"name":"资料库"},"data_domain":{"model_id":101}}`,
			invokeGo: func(service *sdk.KnowledgeService) (any, error) {
				result, err := service.CreateEmpty(t.Context(), "资料库", sdk.WithSemanticModelEmptyDescription("说明"), sdk.WithSemanticModelEmptyImageIndexEnabled(true))
				return result.GetModel(), err
			},
		},
		{
			name: "empty create preserves explicit false and empty description", args: []string{"create-empty", "--name", "资料库", "--description=", "--image-index-enabled=false"},
			method: http.MethodPost, path: "/semantic-models/create-empty",
			response: `{"model":{"id":101,"name":"资料库"},"data_domain":{"model_id":101}}`,
			invokeGo: func(service *sdk.KnowledgeService) (any, error) {
				return service.CreateEmpty(t.Context(), "资料库", sdk.WithSemanticModelEmptyDescription(""), sdk.WithSemanticModelEmptyImageIndexEnabled(false))
			},
		},
		{
			name: "create with sources forwards description files and false", args: []string{"create-with-sources", "--name", "资料库", "--description", "说明", "--image-index-enabled=false"},
			inputs: map[string][]byte{"files-file": files},
			method: http.MethodPost, path: "/semantic-models/create-with-sources",
			response: `{"model":{"id":101,"name":"资料库"},"data_domain":{"model_id":101},"sources":[],"jobs":[]}`,
			invokeGo: func(service *sdk.KnowledgeService) (any, error) {
				return service.CreateWithSources(t.Context(), "资料库", nil, sdk.WithSemanticModelWithSourcesDescription("说明"), sdk.WithSemanticModelWithSourcesFiles(files), sdk.WithSemanticModelWithSourcesImageIndexEnabled(false))
			},
		},
		{
			name: "update table and file definitions", args: []string{"update", "--model-id", "101", "--name", "资料库", "--description", "新说明"},
			inputs: map[string][]byte{"tables-file": tables, "files-file": files},
			method: http.MethodPut, path: "/semantic-models/101", response: `{"updated":true}`,
			invokeGo: func(service *sdk.KnowledgeService) (any, error) {
				return service.Update(t.Context(), "101", "资料库", sdk.WithSemanticModelDescription("新说明"), sdk.WithSemanticModelTables(tables), sdk.WithSemanticModelFiles(files))
			},
		},
		{
			name: "update clears description and preserves JSON null", args: []string{"update", "--model-id", "101", "--name", "资料库", "--description="},
			inputs: map[string][]byte{"files-file": []byte(`null`)},
			method: http.MethodPut, path: "/semantic-models/101", response: `{"updated":true}`,
			invokeGo: func(service *sdk.KnowledgeService) (any, error) {
				return service.Update(t.Context(), "101", "资料库", sdk.WithSemanticModelDescription(""), sdk.WithSemanticModelFiles(json.RawMessage(`null`)))
			},
		},
		{
			name: "append database and volume selections", args: []string{"append-sources", "--model-id", "101"},
			inputs: map[string][]byte{"source-selections-file": selectionJSON},
			method: http.MethodPost, path: "/semantic-models/101/sources",
			response: `{"sources":[],"jobs":[],"skipped_existing_file_ids":["file-1"],"skipped_existing_table_ids":[19]}`,
			invokeGo: func(service *sdk.KnowledgeService) (any, error) {
				return service.AppendSources(t.Context(), "101", nil, sdk.WithSemanticModelAppendSourceSelections(selections...))
			},
		},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			type request struct {
				Method, Path, Workspace string
				Body                    map[string]any
			}
			requests := make(chan request, 2)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				requests <- request{r.Method, r.URL.Path, r.Header.Get("X-Workspace-ID"), body}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"code":"OK","data":` + scenario.response + `}`))
			}))
			t.Cleanup(server.Close)
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = client.Close() })
			want, err := scenario.invokeGo(client.Knowledge("workspace-1"))
			if err != nil {
				t.Fatal(err)
			}
			args := append([]string{}, scenario.args...)
			for flag, input := range scenario.inputs {
				path := filepath.Join(t.TempDir(), flag+".json")
				if err := os.WriteFile(path, input, 0o600); err != nil {
					t.Fatal(err)
				}
				args = append(args, "--"+flag, path)
			}
			var callErr error
			output := captureStdout(t, func() {
				callErr = ExecuteSemanticModelCommand(NewContext(client, "workspace-1", false, true, true), args)
			})
			if callErr != nil {
				t.Fatal(callErr)
			}
			goRequest, cliRequest := <-requests, <-requests
			if !reflect.DeepEqual(cliRequest, goRequest) || cliRequest.Method != scenario.method || cliRequest.Path != "/newmoi"+scenario.path || cliRequest.Workspace != "workspace-1" {
				t.Fatalf("CLI request=%+v, Go SDK request=%+v", cliRequest, goRequest)
			}
			encoded, err := json.Marshal(want)
			if err != nil {
				t.Fatal(err)
			}
			var goOutput, cliOutput any
			if err := json.Unmarshal(encoded, &goOutput); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(output), &cliOutput); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(cliOutput, goOutput) {
				t.Fatalf("CLI output=%v, Go SDK output=%v", cliOutput, goOutput)
			}
		})
	}
}
