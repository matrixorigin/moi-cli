package commands

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestKnowledgeDownloadSharedHTTPContracts(t *testing.T) {
	data, err := os.ReadFile("../testdata/product-contracts/knowledge_download_http_contract.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name, Operation, Method, Path string
		Input                         struct {
			ModelID string `json:"model_id"`
			FileID  string `json:"file_id"`
		} `json:"input"`
		ContentBase64   string            `json:"content_base64"`
		Status          int               `json:"status"`
		Response        json.RawMessage   `json:"response"`
		ResponseHeaders map[string]string `json:"response_headers"`
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
			requests := make(chan struct{}, 2)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != scenario.Method || r.URL.EscapedPath() != scenario.Path || r.URL.RawQuery != "" || r.Header.Get("X-Workspace-ID") != "workspace-1" || r.Header.Get("X-API-Key") != "synthetic-key" || r.Header.Get("Accept") != "*/*" || r.Header.Get("Authorization") != "" || r.Header.Get("Cookie") != "" {
					t.Errorf("request=%s %s headers=%v", r.Method, r.URL.RequestURI(), r.Header)
				}
				body, _ := io.ReadAll(r.Body)
				if len(body) != 0 {
					t.Errorf("request body=%q", body)
				}
				requests <- struct{}{}
				for key, value := range scenario.ResponseHeaders {
					w.Header().Set(key, value)
				}
				if scenario.Status >= 400 {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(scenario.Status)
					_, _ = w.Write(scenario.Response)
					return
				}
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
			service := client.Knowledge("workspace-1")
			var stream *sdk.StreamResponse
			if scenario.Operation == "preview-source-file" {
				stream, err = service.PreviewSourceFile(t.Context(), scenario.Input.ModelID, scenario.Input.FileID)
			} else {
				stream, err = service.PreviewArtifact(t.Context(), scenario.Input.ModelID, scenario.Input.FileID)
			}
			<-requests
			assertKnowledgeDownloadError(t, err, scenario.Status, scenario.Response, scenario.ResponseHeaders)
			if scenario.Status < 400 {
				got, readErr := io.ReadAll(stream.Body)
				if readErr != nil {
					t.Fatal(readErr)
				}
				if closeErr := stream.Body.Close(); closeErr != nil {
					t.Fatal(closeErr)
				}
				if string(got) != string(content) {
					t.Fatalf("Go content=%q, want=%q", got, content)
				}
			}

			output := filepath.Join(t.TempDir(), "download.bin")
			args := []string{scenario.Operation, "--model-id", scenario.Input.ModelID, "--file-id", scenario.Input.FileID, "--output", output}
			var cliErr error
			stdout := captureStdout(t, func() {
				cliErr = ExecuteSemanticModelCommand(NewContext(client, "workspace-1", false, true, true), args)
			})
			<-requests
			assertKnowledgeDownloadError(t, cliErr, scenario.Status, scenario.Response, scenario.ResponseHeaders)
			if scenario.Status < 400 {
				got, readErr := os.ReadFile(output)
				if readErr != nil {
					t.Fatal(readErr)
				}
				if string(got) != string(content) || stdout != "" {
					t.Fatalf("CLI content=%q stdout=%q", got, stdout)
				}
			} else if stdout != "" {
				t.Fatalf("CLI error output=%q", stdout)
			}
		})
	}
}

func assertKnowledgeDownloadError(t *testing.T, err error, status int, raw json.RawMessage, headers map[string]string) {
	t.Helper()
	if status < 400 {
		if err != nil {
			t.Fatal(err)
		}
		return
	}
	var failure *sdk.Error
	if !errors.As(err, &failure) {
		t.Fatalf("error=%T %v", err, err)
	}
	var envelope struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatal(err)
	}
	if failure.Code != envelope.Code || failure.Message != envelope.Msg || failure.Status != status || failure.RequestID != headers["X-Request-ID"] {
		t.Fatalf("error=%+v", failure)
	}
}
