package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"sync/atomic"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestKnowledgeCreateOptionsCLIContract(t *testing.T) {
	var cases []struct {
		Name     string         `json:"name"`
		Request  map[string]any `json:"request"`
		Status   int            `json:"status"`
		Response map[string]any `json:"response"`
	}
	value, err := os.ReadFile("../testdata/product-contracts/knowledge_create_options_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(value, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				if r.Method != http.MethodPost || r.URL.Path != "/newmoi/semantic-models/create-with-sources" || r.Header.Get("X-Workspace-ID") != "workspace-1" {
					t.Errorf("unexpected target: %s %s", r.Method, r.URL.Path)
				}
				var got map[string]any
				if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
					t.Error(err)
				}
				if !reflect.DeepEqual(got, scenario.Request) {
					t.Errorf("create body=%#v, want=%#v", got, scenario.Request)
				}
				w.WriteHeader(scenario.Status)
				_ = json.NewEncoder(w).Encode(scenario.Response)
			}))
			t.Cleanup(server.Close)
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = client.Close() })
			sources := make([]sdk.SemanticModelSourceInput, 0)
			for _, v := range scenario.Request["sources"].([]any) {
				source := v.(map[string]any)
				sources = append(sources, sdk.SemanticModelSourceInput{SourceType: source["source_type"].(string), FileID: source["file_id"].(string)})
			}
			encoded, err := json.Marshal(sources)
			if err != nil {
				t.Fatal(err)
			}
			args := []string{"create-with-sources", "--name", scenario.Request["name"].(string), "--json", string(encoded)}
			if catalogID, present := scenario.Request["target_catalog_id"]; present {
				args = append(args, "--target-catalog-id", fmt.Sprint(catalogID))
			}
			if values, present := scenario.Request["source_selections"]; present {
				selections := make([]sdk.SemanticModelSourceSelectionInput, 0)
				for _, v := range values.([]any) {
					selection := v.(map[string]any)
					selections = append(selections, sdk.SemanticModelSourceSelectionInput{Kind: selection["kind"].(string), VolumeID: int64(selection["volume_id"].(float64)), AllSelected: selection["all_selected"].(bool)})
				}
				encoded, err := json.Marshal(selections)
				if err != nil {
					t.Fatal(err)
				}
				path := filepath.Join(t.TempDir(), "selections.json")
				if err := os.WriteFile(path, encoded, 0o600); err != nil {
					t.Fatal(err)
				}
				args = append(args, "--source-selections-file", path)
			}
			var callErr error
			output := captureStdout(t, func() {
				callErr = ExecuteSemanticModelCommand(NewContext(client, "workspace-1", false, true, true), args)
			})
			if scenario.Status >= 400 {
				var ownerError *sdk.Error
				if !errors.As(callErr, &ownerError) || ownerError.Code != scenario.Response["code"] || ownerError.Message != scenario.Response["msg"] || ownerError.Status != scenario.Status || output != "" {
					t.Fatalf("owner error changed: err=%v output=%s", callErr, output)
				}
			} else {
				if callErr != nil {
					t.Fatal(callErr)
				}
				var receipt struct {
					Model struct {
						ID any `json:"id"`
					} `json:"model"`
				}
				if err := json.Unmarshal([]byte(output), &receipt); err != nil || fmt.Sprint(receipt.Model.ID) != "101" {
					t.Fatalf("create receipt=%s, error=%v", output, err)
				}
			}
			if calls.Load() != 1 {
				t.Errorf("create requests=%d, want one", calls.Load())
			}
		})
	}
}
