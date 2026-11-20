package commands

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sync/atomic"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestImportExportCreateCLIContract(t *testing.T) {
	type scenario struct {
		Name       string          `json:"name"`
		Request    map[string]any  `json:"request"`
		Response   json.RawMessage `json:"response"`
		LocalError string          `json:"local_error"`
	}
	var fixture struct {
		FileImports        []scenario `json:"file_imports"`
		ExplicitTaskImport scenario   `json:"explicit_task_import"`
		Exports            []scenario `json:"exports"`
		OwnerDenial        struct {
			Status int            `json:"status"`
			Body   map[string]any `json:"body"`
		} `json:"owner_denial"`
	}
	value, err := os.ReadFile("../testdata/product-contracts/import_export_create_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(value, &fixture); err != nil {
		t.Fatal(err)
	}
	fixture.FileImports = append(fixture.FileImports, fixture.ExplicitTaskImport)
	for _, group := range []struct {
		command string
		path    string
		cases   []scenario
	}{
		{"create-task", "/newmoi/task", fixture.FileImports},
		{"create-export-task", "/newmoi/export/task/create", fixture.Exports},
	} {
		for _, scenario := range group.cases {
			for _, denied := range []bool{false, true} {
				name := group.command + "/" + scenario.Name
				if denied {
					name += "/owner denial"
				}
				t.Run(name, func(t *testing.T) {
					var calls atomic.Int32
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						calls.Add(1)
						if r.Method != http.MethodPost || r.URL.Path != group.path || r.Header.Get("X-Workspace-ID") != "workspace-1" {
							t.Errorf("unexpected target: %s %s workspace=%s", r.Method, r.URL.Path, r.Header.Get("X-Workspace-ID"))
						}
						var body map[string]any
						if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
							t.Error(err)
						}
						if !reflect.DeepEqual(body, scenario.Request) {
							t.Errorf("request = %#v, want %#v", body, scenario.Request)
						}
						if denied {
							w.WriteHeader(fixture.OwnerDenial.Status)
							_ = json.NewEncoder(w).Encode(fixture.OwnerDenial.Body)
							return
						}
						_ = json.NewEncoder(w).Encode(map[string]any{"code": "OK", "data": scenario.Response})
					}))
					t.Cleanup(server.Close)
					client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"))
					if err != nil {
						t.Fatal(err)
					}
					t.Cleanup(func() { _ = client.Close() })
					input := importExportCLIInput(group.command, scenario.Request)
					encoded, err := json.Marshal(input)
					if err != nil {
						t.Fatal(err)
					}
					var callErr error
					output := captureStdout(t, func() {
						callErr = ExecuteConnectorCommand(NewContext(client, "workspace-1", false, true, true), []string{group.command, "--json", string(encoded)})
					})
					if scenario.LocalError != "" {
						if !sdk.IsCode(callErr, scenario.LocalError) || calls.Load() != 0 || output != "" {
							t.Fatalf("invalid input reached owner: err=%v calls=%d output=%s", callErr, calls.Load(), output)
						}
						return
					}
					if denied {
						var ownerError *sdk.Error
						if !errors.As(callErr, &ownerError) || ownerError.Code != fixture.OwnerDenial.Body["code"] || ownerError.Message != fixture.OwnerDenial.Body["msg"] || ownerError.Status != fixture.OwnerDenial.Status || output != "" {
							t.Fatalf("owner denial was changed: err=%v output=%s", callErr, output)
						}
					} else {
						if callErr != nil {
							t.Fatal(callErr)
						}
						var got, want map[string]any
						if err := json.Unmarshal([]byte(output), &got); err != nil {
							t.Fatal(err)
						}
						if err := json.Unmarshal(scenario.Response, &want); err != nil {
							t.Fatal(err)
						}
						idKey := "id"
						if group.command == "create-task" {
							idKey = "task_id"
						}
						if got[idKey] != want[idKey] {
							t.Errorf("receipt ID = %v, want %v", got[idKey], want[idKey])
						}
					}
					if calls.Load() != 1 {
						t.Errorf("requests = %d, want one create without retry", calls.Load())
					}
				})
			}
		}
	}
}

func importExportCLIInput(command string, body map[string]any) map[string]any {
	if command == "create-task" {
		input := map[string]any{"SourceConnectorID": body["source_connector_id"], "ConfigType": body["config_type"], "SourceConfig": body["source_config"]}
		if taskID, ok := body["task_id"]; ok {
			input["TaskID"] = taskID
		}
		if name, ok := body["name"]; ok {
			input["Name"] = name
		}
		if volume, ok := body["volume_id"]; ok {
			input["VolumeID"] = volume
		}
		return input
	}
	config := body["config"].(map[string]any)
	inputConfig := map[string]any{"S3Config": config["s3_config"]}
	if mergeTitle, ok := config["merge_title_to_text"]; ok {
		inputConfig["MergeTitleToText"] = mergeTitle
	}
	files := make([]map[string]any, 0)
	for _, value := range body["files"].([]any) {
		file := value.(map[string]any)
		files = append(files, map[string]any{"FileID": file["file_id"], "FullPath": file["full_path"], "IsRaw": file["is_raw"]})
	}
	input := map[string]any{"Name": body["task_name"], "ConnectorID": body["connector_id"], "ConnectorName": body["connector_name"], "Type": body["type"], "Config": inputConfig, "Files": files}
	if creator, present := body["creator"]; present {
		input["Creator"] = creator
	}
	return input
}
