package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/matrixorigin/matrixflow/sdk/cli/commands"
	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestCLISandboxNetworkErrorsMatchSharedContract(t *testing.T) {
	var cases []struct {
		Name            string
		Status          int
		DetailBytes     int `json:"detail_bytes"`
		Response        map[string]any
		ExpectedCode    string           `json:"expected_code"`
		ExpectedDetails sdk.ErrorDetails `json:"expected_details"`
	}
	data, err := os.ReadFile("../../testdata/product-contracts/sandbox_network_error_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			if scenario.DetailBytes != 0 {
				detail := strings.Repeat("x", scenario.DetailBytes)
				scenario.Response["detail"] = detail
				if scenario.ExpectedDetails["detail_omitted"] != true {
					scenario.ExpectedDetails["detail"] = detail
				}
			}
			body, err := json.Marshal(scenario.Response)
			if err != nil {
				t.Fatal(err)
			}
			for _, operation := range []string{"list", "download"} {
				t.Run(operation, func(t *testing.T) {
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						path := "/newmoi/workspaces/ws-1/files"
						if operation == "download" {
							path += "/file-1/download"
						}
						if r.Method != http.MethodGet || r.URL.Path != path || r.Header.Get("X-MOI-Runtime-Authorization") != "Bearer task-proof" {
							t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
						}
						w.Header().Set("Content-Type", "application/json")
						w.Header().Set("X-Request-ID", "transport-request")
						w.WriteHeader(scenario.Status)
						_, _ = w.Write(body)
					}))
					t.Cleanup(server.Close)
					client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"), sdk.WithRuntimeAuthorization("Bearer task-proof"))
					if err != nil {
						t.Fatal(err)
					}
					t.Cleanup(func() { _ = client.Close() })
					outputFile := filepath.Join(t.TempDir(), "download.txt")
					args := []string{"list", "--page-size", "1"}
					if operation == "download" {
						args = []string{"download", "--file-id", "file-1", "--output", outputFile}
					}
					callErr := executeCommand(commands.NewContext(client, "ws-1", false, true, true), "file", args)
					var failure *sdk.Error
					if !errors.As(callErr, &failure) {
						t.Fatalf("unstructured failure: %v", callErr)
					}
					var output bytes.Buffer
					writeCommandError(&output, callErr, true)
					var rendered sdk.Error
					if err := json.Unmarshal(output.Bytes(), &rendered); err != nil {
						t.Fatal(err)
					}
					if rendered.Code != scenario.ExpectedCode || rendered.Status != scenario.Status || rendered.RequestID != "transport-request" || !reflect.DeepEqual(rendered.Details, scenario.ExpectedDetails) && !(len(rendered.Details) == 0 && len(scenario.ExpectedDetails) == 0) {
						t.Fatalf("CLI lost diagnostics: %s", output.Bytes())
					}
					if strings.Contains(output.String(), "RAW-SECRET") {
						t.Fatalf("CLI exposed an unrelated response field: %s", output.Bytes())
					}
					if _, err := os.Stat(outputFile); !errors.Is(err, os.ErrNotExist) {
						t.Fatalf("failed download created a file: %v", err)
					}
					output.Reset()
					writeCommandError(&output, callErr, false)
					lines := bytes.Split(bytes.TrimSpace(output.Bytes()), []byte("\n"))
					if !bytes.Contains(lines[0], []byte(scenario.ExpectedCode)) || rendered.Message == "" {
						t.Fatalf("text output lost failure: %s", output.Bytes())
					}
					if detail, ok := scenario.Response["detail"].(string); scenario.ExpectedCode == "policy_denied" && ok && bytes.Contains(lines[0], []byte(detail)) {
						t.Fatalf("raw detail became summary: %s", output.Bytes())
					}
					if len(scenario.ExpectedDetails) != 0 {
						var details sdk.ErrorDetails
						if err := json.Unmarshal(lines[len(lines)-1], &details); err != nil {
							t.Fatal(err)
						}
						if !reflect.DeepEqual(details, scenario.ExpectedDetails) {
							t.Fatalf("text output lost details: %#v", details)
						}
					}
				})
			}
		})
	}
}
