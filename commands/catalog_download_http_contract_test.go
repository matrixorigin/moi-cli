package commands

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestCatalogDownloadSharedHTTPContracts(t *testing.T) {
	data, err := os.ReadFile("../testdata/product-contracts/catalog_download_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name      string `json:"name"`
		Operation string `json:"operation"`
		Method    string `json:"method"`
		Path      string `json:"path"`
		Input     struct {
			VolumeID   string `json:"volume_id"`
			FileID     string `json:"file_id"`
			TableID    int    `json:"table_id"`
			Capability string `json:"capability"`
			AsZip      *bool  `json:"as_zip"`
		} `json:"input"`
		Request       json.RawMessage     `json:"request"`
		Query         map[string][]string `json:"query"`
		ContentBase64 string              `json:"content_base64"`
		Status        int                 `json:"status"`
	}
	if err := json.Unmarshal(data, &cases); err != nil {
		t.Fatal(err)
	}
	for _, scenario := range cases {
		t.Run(scenario.Name, func(t *testing.T) {
			content, err := base64.StdEncoding.DecodeString(scenario.ContentBase64)
			if err != nil {
				t.Fatal(err)
			}
			bodies := make(chan any, 2)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedQuery := url.Values{}
				for key, values := range scenario.Query {
					for _, value := range values {
						expectedQuery.Add(key, value)
					}
				}
				if r.Method != scenario.Method || r.URL.EscapedPath() != scenario.Path || !reflect.DeepEqual(r.URL.Query(), expectedQuery) || r.Header.Get("X-Workspace-ID") != "workspace-1" || r.Header.Get("X-API-Key") != "synthetic-key" || r.Header.Get("Accept") != "*/*" || r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" {
					t.Errorf("request=%s %s headers=%v", r.Method, r.URL.RequestURI(), r.Header)
				}
				var body any
				if len(scenario.Request) != 0 && string(scenario.Request) != "null" {
					decoder := json.NewDecoder(r.Body)
					decoder.UseNumber()
					if err := decoder.Decode(&body); err != nil {
						t.Errorf("decode request: %v", err)
					}
				}
				bodies <- body
				w.Header().Set("Content-Type", "application/octet-stream")
				w.WriteHeader(scenario.Status)
				_, _ = w.Write(content)
			}))
			defer server.Close()
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-key"))
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()
			var options []sdk.DownloadFileOption
			operation := scenario.Operation
			if operation == "" {
				operation = "download-file"
			}
			args := []string{operation}
			if operation == "download-table" {
				args = append(args, "--table-id", fmt.Sprint(scenario.Input.TableID))
			} else if operation == "download-workspace-file" || operation == "preview-workspace-file" {
				args = append(args, "--file-id", scenario.Input.FileID)
			} else if operation == "catalog-file-content" {
				args = append(args, "--file-id", scenario.Input.FileID, "--capability", scenario.Input.Capability)
			} else {
				args = append(args, "--volume-id", scenario.Input.VolumeID, "--file-id", scenario.Input.FileID)
			}
			if scenario.Input.AsZip != nil {
				options = append(options, sdk.WithFileAsZip(*scenario.Input.AsZip))
				args = append(args, "--as-zip="+boolString(*scenario.Input.AsZip))
			}
			var stream *sdk.StreamResponse
			if operation == "preview-file" {
				stream, err = client.Catalog("workspace-1").PreviewFile(t.Context(), scenario.Input.VolumeID, scenario.Input.FileID)
			} else if operation == "download-table" {
				stream, err = client.Catalog("workspace-1").DownloadTable(t.Context(), scenario.Input.TableID)
			} else if operation == "download-workspace-file" {
				stream, err = client.Catalog("workspace-1").DownloadWorkspaceFile(t.Context(), scenario.Input.FileID)
			} else if operation == "preview-workspace-file" {
				stream, err = client.Catalog("workspace-1").PreviewWorkspaceFile(t.Context(), scenario.Input.FileID)
			} else if operation == "catalog-file-content" {
				stream, err = client.Catalog("workspace-1").CatalogFileContent(t.Context(), scenario.Input.FileID, scenario.Input.Capability)
			} else {
				stream, err = client.Catalog("workspace-1").DownloadFile(t.Context(), scenario.Input.VolumeID, scenario.Input.FileID, options...)
			}
			if err != nil {
				t.Fatal(err)
			}
			goContent, err := io.ReadAll(stream.Body)
			if err != nil {
				t.Fatal(err)
			}
			if err := stream.Body.Close(); err != nil {
				t.Fatal(err)
			}
			if got := <-bodies; !reflect.DeepEqual(got, catalogFolderCLIJSON(scenario.Request)) || !reflect.DeepEqual(goContent, content) {
				t.Fatalf("Go body=%#v content=%q", got, goContent)
			}

			output := filepath.Join(t.TempDir(), "download.bin")
			args = append(args, "--output", output)
			if err := ExecuteCatalogCommand(NewContext(client, "workspace-1", false, false, true), args); err != nil {
				t.Fatal(err)
			}
			cliContent, err := os.ReadFile(output)
			if err != nil {
				t.Fatal(err)
			}
			if got := <-bodies; !reflect.DeepEqual(got, catalogFolderCLIJSON(scenario.Request)) || !reflect.DeepEqual(cliContent, content) {
				t.Fatalf("CLI body=%#v content=%q", got, cliContent)
			}
		})
	}
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
