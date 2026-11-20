package commands

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestKnowledgeQuoteSelectsInspectedContentWithoutPositionalAssumptions(t *testing.T) {
	const content = "已签收包裹才计入完成交付。 Keep original IDs."
	for _, first := range []bool{false, true} {
		t.Run(map[bool]string{false: "later page", true: "first position"}[first], func(t *testing.T) {
			chunks := []any{
				map[string]any{"chunk_id": "heading", "file_id": "manual", "semantic_model_id": 7, "content": "# Delivery policy", "chunk_index": 0},
				map[string]any{"chunk_id": "target", "file_id": "manual", "semantic_model_id": 7, "content": content, "chunk_index": 1},
			}
			if first {
				chunks[0], chunks[1] = chunks[1], chunks[0]
			}
			path := filepath.Join(t.TempDir(), "saved source.json")
			raw, err := json.Marshal(map[string]any{"artifacts": []any{
				map[string]any{"artifact_id": "page1", "data": map[string]any{"chunks": []any{map[string]any{"chunk_id": "other", "file_id": "manual", "semantic_model_id": 7, "content": "Do not reuse parcel IDs.", "chunk_index": 9}}}},
				map[string]any{"artifact_id": "page2", "data": map[string]any{"chunks": chunks}},
			}})
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, raw, 0600); err != nil {
				t.Fatal(err)
			}
			var observed []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
					return
				}
				if strings.HasSuffix(r.URL.Path, "/read-evidence") {
					digest := sha256.Sum256([]byte(content))
					if body["artifact_id"] != "page2" || body["chunk_id"] != "target" || body["local_content_sha256"] != hex.EncodeToString(digest[:]) {
						t.Errorf("quote lost canonical identity/content digest: %v", body)
					}
					observed = append(observed, "read-target")
				} else if strings.HasSuffix(r.URL.Path, "/cite") {
					for _, raw := range body["sources"].([]any) {
						source := raw.(map[string]any)
						if source["type"] == "sql_result" {
							observed = append(observed, source["artifact_id"].(string))
							continue
						}
						pages := map[string]string{"target": "page2", "other": "page1"}
						if source["artifact_id"] != pages[source["chunk_id"].(string)] {
							t.Errorf("wrong citation page: %v", source)
						}
						observed = append(observed, source["chunk_id"].(string))
					}
				}
				_, _ = io.WriteString(w, `{"code":"OK","data":{"output":{"accepted":true}}}`)
			}))
			t.Cleanup(server.Close)
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("key"))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = client.Close() })
			ctx := NewContext(client, "workspace", false, true, true)
			for _, args := range [][]string{
				{"read-evidence", "--from", path, "--quote", "已签收包裹才计入完成交付", "--local-read"},
				{"cite", "--json", `{"sources":[{"type":"sql_result","artifact_id":"saved-query"}]}`, "--from", path, "--quote", "已签收包裹才计入完成交付", "--from", path, "--quote", "Do not reuse parcel IDs."},
			} {
				captureStdout(t, func() {
					if err := ExecuteSemanticModelCommand(ctx, args); err != nil {
						t.Fatal(err)
					}
				})
			}
			if !reflect.DeepEqual(observed, []string{"read-target", "saved-query", "target", "other"}) {
				t.Fatalf("wrong read/citation identities: %v", observed)
			}
		})
	}
}

func TestKnowledgeQuoteRejectsMissingAmbiguousAndMixedSelectorsBeforeRequest(t *testing.T) {
	path := filepath.Join(t.TempDir(), "saved.json")
	if err := os.WriteFile(path, []byte(`{"artifacts":[{"artifact_id":"a","data":{"chunks":[{"chunk_id":"first","content":"Shared rule. Alpha."},{"chunk_id":"second","content":"Shared rule. Beta."}]}}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Error("invalid/ambiguous quote reached owner") }))
	t.Cleanup(server.Close)
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("key"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	for _, tc := range []struct {
		args []string
		code string
	}{
		{[]string{"cite", "--from", path, "--quote", "absent"}, "CLI_KNOWLEDGE_QUOTE_NOT_FOUND"},
		{[]string{"cite", "--from", path, "--quote", "Shared rule."}, "CLI_KNOWLEDGE_QUOTE_AMBIGUOUS"},
		{[]string{"cite", "--json", `{"sources":[{"type":"sql_result","artifact_id":"saved-query"}]}`, "--from", path, "--quote", "Shared rule."}, "CLI_KNOWLEDGE_QUOTE_AMBIGUOUS"},
		{[]string{"cite", "--from", path, "--quote", "shared rule."}, "CLI_KNOWLEDGE_QUOTE_NOT_FOUND"},
		{[]string{"cite", "--from", path, "--quote", "   "}, "CLI_KNOWLEDGE_INPUT_INVALID"},
		{[]string{"cite", "--from", path, "--quote", "Alpha", "--indices", "0"}, "CLI_KNOWLEDGE_INPUT_INVALID"},
		{[]string{"read-evidence", "--from", path, "--quote", "Alpha", "--index", "0"}, "CLI_KNOWLEDGE_INPUT_INVALID"},
		{[]string{"cite", "--quote", "Alpha"}, "CLI_KNOWLEDGE_INPUT_INVALID"},
		{[]string{"cite", "--from", path, "--quote", "Alpha", "--from", path}, "CLI_KNOWLEDGE_INPUT_INVALID"},
	} {
		err := ExecuteSemanticModelCommand(NewContext(client, "workspace", false, true, true), tc.args)
		var failure *sdk.Error
		if !errors.As(err, &failure) || failure.Code != tc.code {
			t.Fatalf("quote failure = %v, want %s", err, tc.code)
		}
		if tc.code == "CLI_KNOWLEDGE_QUOTE_AMBIGUOUS" && !reflect.DeepEqual(failure.Details["candidate_indices"], []int{0, 1}) {
			t.Fatalf("ambiguity lost exact choices: %+v", failure)
		}
	}
}
