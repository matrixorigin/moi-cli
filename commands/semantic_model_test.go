package commands

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync/atomic"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestCreateKnowledgeBaseUsesInitializedCreationAndPreservesModelOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/newmoi/semantic-models/create-empty" {
			http.Error(w, `{"code":"UnexpectedCreatePath","msg":"knowledge base creation must initialize its model bindings"}`, http.StatusBadRequest)
			return
		}
		if r.Header.Get("X-Workspace-ID") != "ws-1" {
			t.Errorf("workspace scope = %q", r.Header.Get("X-Workspace-ID"))
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
			return
		}
		if body["name"] != "kb-cli" {
			t.Errorf("knowledge base request = %#v", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "OK", "msg": "OK", "data": map[string]any{
			"model": map[string]any{"id": 101, "name": "kb-cli", "files": map[string]any{
				"embedding_model": "bge-m3", "embedding_dimension": 1024, "vector_table": "kb_101_text_index", "file_ids": []string{},
			}},
			"data_domain": map[string]any{"model_id": 101},
		}})
	}))
	t.Cleanup(server.Close)
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
	if err != nil {
		t.Fatal(err)
	}
	var callErr error
	output := captureStdout(t, func() {
		callErr = ExecuteSemanticModelCommand(NewContext(client, "ws-1", false, true, true), []string{"create", "--name", "kb-cli"})
	})
	if callErr != nil {
		t.Fatalf("create initialized knowledge base: %v", callErr)
	}
	var result struct {
		ID    int64  `json:"id"`
		Name  string `json:"name"`
		Files struct {
			EmbeddingModel string `json:"embedding_model"`
			VectorTable    string `json:"vector_table"`
		} `json:"files"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("decode CLI model output: %v; output=%s", err, output)
	}
	if result.ID != 101 || result.Name != "kb-cli" || result.Files.EmbeddingModel != "bge-m3" || result.Files.VectorTable != "kb_101_text_index" {
		t.Fatalf("CLI must return the initialized model at the existing output path: %+v", result)
	}
}

func TestCreateKnowledgeBasePreservesInitializationFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/newmoi/semantic-models/create-empty" {
			t.Errorf("unexpected creation request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("X-Request-ID", "kb-create-denied")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"code":"ErrForbidden","msg":"无权创建知识库"}`))
	}))
	t.Cleanup(server.Close)
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
	if err != nil {
		t.Fatal(err)
	}
	var callErr error
	output := captureStdout(t, func() {
		callErr = ExecuteSemanticModelCommand(NewContext(client, "ws-1", false, true, true), []string{"create", "--name", "kb-cli"})
	})
	var failure *sdk.Error
	if !errors.As(callErr, &failure) || failure.Status != http.StatusForbidden || failure.Code != "ErrForbidden" || failure.Message != "无权创建知识库" || failure.RequestID != "kb-create-denied" {
		t.Fatalf("initialization error must be preserved: %#v", callErr)
	}
	if output != "" {
		t.Fatalf("failed creation printed a successful model: %s", output)
	}
}

func TestCreateWithSourcesImageIndexEnabledSendsRequestBody(t *testing.T) {
	t.Parallel()

	var gotPath string
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := json.Unmarshal(raw, &gotBody); err != nil {
			t.Errorf("decode body: %v raw=%s", err, string(raw))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": "OK",
			"msg":  "OK",
			"data": map[string]any{
				"model":   map[string]any{"id": 101, "name": "kb-image"},
				"sources": []any{},
				"jobs":    []any{},
			},
		})
	}))
	t.Cleanup(server.Close)

	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewContext(client, "ws-1", false, true, true)

	if err := runKnowledgeServiceCreateWithSources(ctx, []string{
		"--name", "kb-image",
		"--image-index-enabled",
		"--json", `[]`,
	}); err != nil {
		t.Fatalf("create-with-sources --image-index-enabled: %v", err)
	}

	if gotPath != "/newmoi/semantic-models/create-with-sources" {
		t.Fatalf("path = %q, want /newmoi/semantic-models/create-with-sources", gotPath)
	}
	if got := gotBody["image_index_enabled"]; got != true {
		t.Fatalf("image_index_enabled = %#v, want true; body=%#v", got, gotBody)
	}
	if got := gotBody["name"]; got != "kb-image" {
		t.Fatalf("name = %#v, want kb-image", got)
	}
}

func TestCreateWithSourcesOmitsImageIndexEnabledByDefault(t *testing.T) {
	t.Parallel()

	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := json.Unmarshal(raw, &gotBody); err != nil {
			t.Errorf("decode body: %v raw=%s", err, string(raw))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": "OK",
			"msg":  "OK",
			"data": map[string]any{
				"model":   map[string]any{"id": 101, "name": "kb-text"},
				"sources": []any{},
				"jobs":    []any{},
			},
		})
	}))
	t.Cleanup(server.Close)

	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewContext(client, "ws-1", false, true, true)

	if err := runKnowledgeServiceCreateWithSources(ctx, []string{
		"--name", "kb-text",
		"--json", `[]`,
	}); err != nil {
		t.Fatalf("create-with-sources default: %v", err)
	}

	if _, ok := gotBody["image_index_enabled"]; ok {
		t.Fatalf("image_index_enabled must be omitted by default; body=%#v", gotBody)
	}
}

func TestCreateEmptySendsDataKnowledgeBaseOptions(t *testing.T) {
	t.Parallel()

	var gotPath string
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.EscapedPath()
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := json.Unmarshal(raw, &gotBody); err != nil {
			t.Errorf("decode body: %v raw=%s", err, string(raw))
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": "OK",
			"msg":  "OK",
			"data": map[string]any{
				"model":       map[string]any{"id": 101, "name": "kb-empty"},
				"data_domain": map[string]any{"model_id": 101},
			},
		})
	}))
	t.Cleanup(server.Close)

	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("user-key"))
	if err != nil {
		t.Fatal(err)
	}
	ctx := NewContext(client, "ws-1", false, true, true)

	if err := ExecuteSemanticModelCommand(ctx, []string{
		"create-empty",
		"--name", "kb-empty",
		"--description", "data knowledge base",
		"--image-index-enabled",
	}); err != nil {
		t.Fatalf("create-empty: %v", err)
	}

	if gotPath != "/newmoi/semantic-models/create-empty" {
		t.Fatalf("path = %q, want /newmoi/semantic-models/create-empty", gotPath)
	}
	if got := gotBody["name"]; got != "kb-empty" {
		t.Fatalf("name = %#v, want kb-empty", got)
	}
	if got := gotBody["description"]; got != "data knowledge base" {
		t.Fatalf("description = %#v, want data knowledge base", got)
	}
	if got := gotBody["image_index_enabled"]; got != true {
		t.Fatalf("image_index_enabled = %#v, want true", got)
	}
}

// #18283: invalid table IDs must fail before any source-existence request.
func TestCheckSourceExistenceTableIDs(t *testing.T) {
	for _, tt := range []struct {
		name    string
		input   string
		want    []int64
		invalid bool
	}{
		{name: "mixed", input: "1,not-an-id,2", invalid: true},
		{name: "overflow", input: "9223372036854775808", invalid: true},
		{name: "underflow", input: "-9223372036854775809", invalid: true},
		{name: "invalid", input: "xyz", invalid: true},
		{name: "fraction", input: "1,2.5", invalid: true},
		{name: "whitespace and duplicates", input: " 2, 1,2 ", want: []int64{2, 1, 2}},
		{name: "integer boundaries", input: "-9223372036854775808,9223372036854775807,0", want: []int64{-9223372036854775808, 9223372036854775807, 0}},
		{name: "empty", input: ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var requests atomic.Int64
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				var body struct {
					TableIDs []int64  `json:"table_ids"`
					FileIDs  []string `json:"file_ids"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				if !tt.invalid && (!slices.Equal(body.TableIDs, tt.want) || !slices.Equal(body.FileIDs, []string{"file-1"})) {
					t.Errorf("request IDs = %v / %v, want %v / [file-1]", body.TableIDs, body.FileIDs, tt.want)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, `{"code":"OK","msg":"OK","data":{"file_ids":[],"table_ids":[]}}`)
			}))
			t.Cleanup(server.Close)
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("synthetic-test-key"))
			if err != nil {
				t.Fatal(err)
			}
			ctx := NewContext(client, "ws-1", false, true, true)
			err = ExecuteSemanticModelCommand(ctx, []string{"check-source-existence", "--model-id", "101", "--file-ids", "file-1", "--table-ids", tt.input})
			if tt.invalid {
				if err == nil || !strings.Contains(err.Error(), "CLI_INVALID_INT64_CSV") {
					t.Errorf("error = %v, want CLI_INVALID_INT64_CSV", err)
				}
				if requests.Load() != 0 {
					t.Errorf("invalid input sent %d requests", requests.Load())
				}
			} else if err != nil || requests.Load() != 1 {
				t.Errorf("error = %v, requests = %d; want one successful request", err, requests.Load())
			}
		})
	}
}
