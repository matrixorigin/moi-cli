package commands

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestFileListUsesWorkspaceDiscoveryAndRuntimeIdentity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || request.URL.Path != "/newmoi/workspaces/ws-1/files" {
			t.Errorf("request = %s %s", request.Method, request.URL.Path)
		}
		if request.Header.Get("X-MOI-Runtime-Authorization") != "Bearer runtime-proof" || request.Header.Get("X-Workspace-ID") != "ws-1" {
			t.Error("workspace discovery lost the runtime identity")
		}
		query := request.URL.Query()
		if query.Get("page_size") != "2" || query.Get("page_token") != "next" || query.Get("filter_file_name_fuzzy") != "report" ||
			strings.Join(query["filter_file_id"], ",") != "f1,f2" || strings.Join(query["filter_file_ext"], ",") != "csv,txt" {
			t.Errorf("query = %v", query)
		}
		if err := json.NewEncoder(w).Encode(map[string]any{"items": []any{map[string]any{"id": 7, "volume_id": 3, "file_id": "f1", "file_name": "report.csv", "size": 12}}, "total": 1}); err != nil {
			t.Error(err)
		}
	}))
	t.Cleanup(server.Close)
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"), sdk.WithRuntimeAuthorization("Bearer runtime-proof"))
	if err != nil {
		t.Fatal(err)
	}
	if err := ExecuteFileCommand(NewContext(client, "ws-1", false, true, true), []string{"list", "--name", "report", "--file-ids", "f1,f2", "--extensions", "csv,txt", "--page-size", "2", "--page-token", "next"}); err != nil {
		t.Fatal(err)
	}
}

func TestFileListEmptyPageIsIterableJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		_, _ = w.Write([]byte(`{"code":"OK","data":{"items":[],"total":0}}`))
	}))
	t.Cleanup(server.Close)
	client, err := sdk.New(server.URL, sdk.WithAPIKey("user-key"))
	if err != nil {
		t.Fatal(err)
	}
	output := captureStdout(t, func() {
		if err := ExecuteFileCommand(NewContext(client, "ws-1", false, true, true), []string{"list"}); err != nil {
			t.Fatal(err)
		}
	})
	var result struct {
		Items json.RawMessage `json:"items"`
		Total int64           `json:"total"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatal(err)
	}
	if string(result.Items) != "[]" || result.Total != 0 {
		t.Fatalf("empty list = %s", output)
	}
}

func TestFileUploadPublishesRuntimeArtifactThroughProductSDK(t *testing.T) {
	input := filepath.Join(t.TempDir(), "report.txt")
	if err := os.WriteFile(input, []byte("artifact body"), 0o600); err != nil {
		t.Fatal(err)
	}
	var gotWorkspace, gotRuntimeAuthorization, gotFilename, gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.EscapedPath() != "/newmoi/workspaces/ws-1/runtime-artifacts" {
			t.Fatalf("request = %s %s", request.Method, request.URL.EscapedPath())
		}
		gotWorkspace = request.Header.Get("X-Workspace-ID")
		gotRuntimeAuthorization = request.Header.Get("X-MOI-Runtime-Authorization")
		file, header, err := request.FormFile("file")
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		gotFilename = header.Filename
		body, err := io.ReadAll(file)
		if err != nil {
			t.Fatal(err)
		}
		gotBody = string(body)
		_, _ = w.Write([]byte(`{"file_id":"file-1","original_name":"renamed.txt","size":13,"md5":"md5","artifact_id":"artifact-1","sha256":"sha256:abc"}`))
	}))
	t.Cleanup(server.Close)

	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"), sdk.WithRuntimeAuthorization("Bearer runtime-proof"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewContext(client, "ws-1", false, true, true)
	if err := ExecuteFileCommand(ctx, []string{"upload", "--yes", "--file", input, "--name", "renamed.txt", "--publish-artifact"}); err != nil {
		t.Fatal(err)
	}
	if gotWorkspace != "ws-1" || gotRuntimeAuthorization != "Bearer runtime-proof" || gotFilename != "renamed.txt" || gotBody != "artifact body" {
		t.Fatalf("request = workspace=%q runtime=%q filename=%q body=%q", gotWorkspace, gotRuntimeAuthorization, gotFilename, gotBody)
	}
}

func TestFileDownloadUsesProductSDKTaskVisibleRoute(t *testing.T) {
	output := filepath.Join(t.TempDir(), "download.txt")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet || request.URL.EscapedPath() != "/newmoi/workspaces/ws-1/files/file-1/download" {
			t.Fatalf("request = %s %s", request.Method, request.URL.EscapedPath())
		}
		if got := request.Header.Get("X-MOI-Runtime-Authorization"); got != "Bearer runtime-proof" {
			t.Fatalf("runtime authorization = %q", got)
		}
		_, _ = w.Write([]byte("downloaded bytes"))
	}))
	t.Cleanup(server.Close)

	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"), sdk.WithRuntimeAuthorization("Bearer runtime-proof"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewContext(client, "ws-1", false, false, true)
	if err := ExecuteFileCommand(ctx, []string{"download", "--yes", "--file-id", "file-1", "--output", output}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "downloaded bytes" {
		t.Fatalf("download = %q", data)
	}
}

func TestFileAnalyzeImageUsesProductSDKTaskScopedRoute(t *testing.T) {
	input := filepath.Join(t.TempDir(), "prepared.jpg")
	if err := os.WriteFile(input, []byte("prepared jpeg bytes"), 0o600); err != nil {
		t.Fatal(err)
	}
	var gotWorkspace, gotRuntimeAuthorization, gotFilename, gotQuestion, gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.EscapedPath() != "/newmoi/workspaces/ws-1/runtime-image-analysis" {
			t.Fatalf("request = %s %s", request.Method, request.URL.EscapedPath())
		}
		gotWorkspace = request.Header.Get("X-Workspace-ID")
		gotRuntimeAuthorization = request.Header.Get("X-MOI-Runtime-Authorization")
		gotQuestion = request.FormValue("question")
		file, header, err := request.FormFile("file")
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		gotFilename = header.Filename
		body, err := io.ReadAll(file)
		if err != nil {
			t.Fatal(err)
		}
		gotBody = string(body)
		_, _ = w.Write([]byte(`{"answer":"a cat"}`))
	}))
	t.Cleanup(server.Close)
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"), sdk.WithRuntimeAuthorization("Bearer runtime-proof"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewContext(client, "ws-1", false, true, true)
	if err := ExecuteFileCommand(ctx, []string{"analyze-image", "--yes", "--file", input, "--question", "what is shown?"}); err != nil {
		t.Fatal(err)
	}
	if gotWorkspace != "ws-1" || gotRuntimeAuthorization != "Bearer runtime-proof" ||
		gotFilename != "prepared.jpg" || gotQuestion != "what is shown?" || gotBody != "prepared jpeg bytes" {
		t.Fatalf("request = workspace=%q runtime=%q filename=%q question=%q body=%q", gotWorkspace, gotRuntimeAuthorization, gotFilename, gotQuestion, gotBody)
	}
}

func TestFileUploadRequiresRuntimeArtifactFlag(t *testing.T) {
	err := ExecuteFileCommand(nil, []string{"upload", "--file", "report.txt"})
	if err == nil || !strings.Contains(err.Error(), "--publish-artifact is required") {
		t.Fatalf("error = %v", err)
	}
}

func TestFileUploadErrorPreservesRuntimeErrorCode(t *testing.T) {
	input := filepath.Join(t.TempDir(), "report.txt")
	if err := os.WriteFile(input, []byte("artifact body"), 0o600); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error_code":"runtime_artifact_upload_denied"}`))
	}))
	t.Cleanup(server.Close)
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"), sdk.WithRuntimeAuthorization("Bearer runtime-proof"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewContext(client, "ws-1", false, false, true)
	err = ExecuteFileCommand(ctx, []string{"upload", "--file", input, "--publish-artifact"})
	if !sdk.IsCode(err, "runtime_artifact_upload_denied") {
		t.Fatalf("error = %v", err)
	}
}
