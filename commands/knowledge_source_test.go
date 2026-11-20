package commands

import (
	"encoding/json"
	"errors"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestKnowledgeSourceReadingViewSeparatesHTMLRowsWithoutRewritingEvidence(t *testing.T) {
	content := `<table data-note="literal <tr> in attribute"><tr><td>Branch A &amp; B</td><td>条件①</td></tr><TR><td>STEP-9</td><td>Keep  2 spaces &lt;tr&gt;</td></TR></table><script>const example = '<tr>';</script>`
	artifact, err := structpb.NewStruct(map[string]any{"data": map[string]any{"chunks": []any{map[string]any{
		"file_name": "Procedure", "chunk_index": 4, "chunk_index_scope": "file", "content": content,
	}}}})
	if err != nil {
		t.Fatal(err)
	}
	result := &sdk.KnowledgeRuntimeResult{Artifacts: []*structpb.Struct{artifact}}
	projection, err := knowledgeReadingView(result, "source.json", true)
	if err != nil {
		t.Fatal(err)
	}
	expected := `<table data-note="literal <tr> in attribute">` + "\n" + `<tr><td>Branch A &amp; B</td><td>条件①</td></tr>` + "\n" + `<TR><td>STEP-9</td><td>Keep  2 spaces &lt;tr&gt;</td></TR></table><script>const example = '<tr>';</script>`
	if !strings.Contains(string(projection.text), expected) {
		t.Fatalf("table rows cannot be inspected separately without changing literals: %s", projection.text)
	}
	if actual := artifact.AsMap()["data"].(map[string]any)["chunks"].([]any)[0].(map[string]any)["content"]; actual != content {
		t.Fatalf("reading view changed canonical evidence: %v", actual)
	}
}

func TestKnowledgeSourceReadingViewPreservesPageScopeAndUnpositionedEvidence(t *testing.T) {
	for _, tc := range []struct {
		name, order       string
		first, second     map[string]any
		secondBeforeFirst bool
	}{
		{"page-local indices", "page_then_chunk_index", map[string]any{"page_number": 2, "chunk_index": 0, "chunk_index_scope": "page"}, map[string]any{"page_number": 1, "chunk_index": 9, "chunk_index_scope": "page"}, true},
		{"missing page position", "canonical_artifact_order", map[string]any{"page_number": 2, "chunk_index": 0, "chunk_index_scope": "page"}, map[string]any{"chunk_index": 9, "chunk_index_scope": "page"}, false},
		{"mixed scopes", "canonical_artifact_order", map[string]any{"chunk_index": 20, "chunk_index_scope": "file"}, map[string]any{"page_number": 1, "chunk_index": 9, "chunk_index_scope": "page"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.first["content"], tc.second["content"] = "First passage\n同一値😀", "Second passage\n同一値😀"
			artifact, err := structpb.NewStruct(map[string]any{"data": map[string]any{"chunks": []any{tc.first, tc.second}}})
			if err != nil {
				t.Fatal(err)
			}
			projection, err := knowledgeReadingView(&sdk.KnowledgeRuntimeResult{Artifacts: []*structpb.Struct{artifact}}, "receipt.json", true)
			if err != nil || projection.order != tc.order {
				t.Fatalf("incorrect source-position semantics: %s %v", projection.order, err)
			}
			text := string(projection.text)
			if !strings.Contains(text, tc.first["content"].(string)) || !strings.Contains(text, tc.second["content"].(string)) || !strings.Contains(text, `"index":0`) || !strings.Contains(text, `"index":1`) {
				t.Fatalf("reading view lost a canonical passage or reference: %s", text)
			}
			if (strings.Index(text, "Second passage") < strings.Index(text, "First passage")) != tc.secondBeforeFirst {
				t.Fatalf("incorrect source reading order: %s", text)
			}
		})
	}
}

func TestKnowledgeReadSourceKeepsCompletePagesAndOriginalSelectors(t *testing.T) {
	const table = `<table><tr><td>Initial entry</td><td>Automatic</td></tr><tr><td>Later edit</td><td>Manual &amp; optional</td></tr></table>`
	for _, tc := range []struct {
		name, content string
		byName, table bool
	}{
		{"source name", table, true, true},
		{"saved inventory", table, false, true},
		{"plain source", "CREATE TABLE example (value INT); Later edit is manual.", true, false},
		{"escaped table example", "Use &lt;table&gt; to show markup as text.", true, false},
		{"script table literal", `<script>const example = '<table><tr><td>literal</td></tr></table>';</script>`, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			requests := map[string]string{}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, "/find-files") {
					var request sdk.KnowledgeFindFilesRequest
					raw := json.RawMessage{}
					if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
						t.Error(err)
						return
					}
					if err := protojson.Unmarshal(raw, &request); err != nil || request.Query != "Manual" {
						t.Errorf("source-name lookup changed the requested name: %v %v", &request, err)
					}
					_, _ = w.Write([]byte(`{"code":"OK","data":{"output":{"row_count":1,"files":[{"semantic_model_id":5000000000,"file_id":"manual","file_name":"Manual.xlsx"}]}}}`))
					return
				}
				var request sdk.KnowledgeSearchRequest
				raw := json.RawMessage{}
				if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
					t.Error(err)
					return
				}
				if err := protojson.Unmarshal(raw, &request); err != nil {
					t.Error(err)
					return
				}
				page := request.GetSourcePage()
				if !strings.HasSuffix(r.URL.Path, "/search") || page.GetFileId() != "manual" || page.GetSemanticModelId() != 5000000000 || page.GetPageSize() != 100 || request.MaxHits != 0 || request.MaxRows != 0 {
					t.Errorf("complete-source read changed source or mode: %v", &request)
				}
				if request.RequestId == "" {
					t.Error("source-page request has no replay identity")
				}
				if prior, exists := requests[request.RequestId]; exists && prior != page.GetCursor() {
					t.Error("different pages reused one operation identity")
				}
				requests[request.RequestId] = page.GetCursor()
				cursor, next, id, content := "", "next", "initial", "Input opens the quantity popup."
				if page.GetCursor() == "next" {
					cursor, next, id, content = "next", "rows", "conditions", "Single-item input does not open that popup."
				} else if page.GetCursor() == "rows" {
					cursor, next, id, content = "rows", "", "branches", tc.content
				} else if page.GetCursor() != "" {
					t.Errorf("invented continuation: %q", page.GetCursor())
				}
				coverage := map[string]any{"semantic_model_id": int64(5000000000), "file_id": "manual", "cursor": cursor, "has_more": next != "", "next_cursor": next}
				chunk := map[string]any{"semantic_model_id": int64(5000000000), "file_id": "manual", "file_name": "Manual.xlsx", "chunk_id": id, "content": content}
				chunk["file_name"] = "PRD-DEV-Manual.xlsx"
				chunk["chunk_index_scope"], chunk["chunk_index"] = "file", 20
				if id == "conditions" {
					chunk["chunk_index"] = 10
				} else if id == "branches" {
					chunk["chunk_index"] = 30
				}
				result := map[string]any{"call_id": "call-" + id, "output": map[string]any{"source_page": coverage, "row_count": 1, "chunks": []any{map[string]any{"content_preview": "prefix only"}}}, "artifacts": []any{map[string]any{"artifact_id": "artifact-" + id, "data": map[string]any{"source_page": coverage, "chunks": []any{chunk}}}}}
				if err := json.NewEncoder(w).Encode(map[string]any{"code": "OK", "data": result}); err != nil {
					t.Error(err)
				}
			}))
			t.Cleanup(server.Close)
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("key"), sdk.WithRuntimeAuthorization("Bearer task"))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := client.Close(); err != nil {
					t.Error(err)
				}
			})
			directory := t.TempDir()
			output := filepath.Join(directory, "source.json")
			args := []string{"read-source", "--output", output}
			if tc.byName {
				args = append(args, "--source-name", "Manual")
			} else {
				inventory := filepath.Join(directory, "files.json")
				if err := os.WriteFile(inventory, []byte(`{"output":{"files":[{"semantic_model_id":5000000000,"file_id":"manual"}]}}`), 0600); err != nil {
					t.Fatal(err)
				}
				args = append(args, "--from", inventory, "--index", "0")
			}
			printed := captureStdout(t, func() {
				if err := ExecuteSemanticModelCommand(NewContext(client, "ws1", false, true, true), args); err != nil {
					t.Fatal(err)
				}
			})
			var summary struct {
				Analysis *struct {
					SkillID  string   `json:"skill_id"`
					Script   string   `json:"script"`
					Args     []string `json:"base_arguments"`
					Required []string `json:"required_arguments"`
				} `json:"local_analysis"`
			}
			if err := json.Unmarshal([]byte(printed), &summary); err != nil {
				t.Fatal(err)
			}
			if tc.table {
				if summary.Analysis == nil || summary.Analysis.SkillID != "moi.agent.momo.skill.moi-cli-semantic-model" || summary.Analysis.Script != "scripts/compare_passages.py" {
					t.Fatalf("complete table source does not identify its frozen analysis asset: %+v", summary.Analysis)
				}
				if !reflect.DeepEqual(summary.Analysis.Required, []string{"--contains"}) {
					t.Fatalf("analysis must declare its required literal selector: %v", summary.Analysis.Required)
				}
				if !reflect.DeepEqual(summary.Analysis.Args, []string{"--source", output, "--output", output + ".inspection.json"}) {
					t.Fatalf("analysis must consume this complete receipt and save its own report: %v", summary.Analysis.Args)
				}
			} else if summary.Analysis != nil {
				t.Fatalf("non-table source was redirected to table analysis: %+v", summary.Analysis)
			}
			raw, err := os.ReadFile(output)
			if err != nil {
				t.Fatal(err)
			}
			var saved sdk.KnowledgeRuntimeResult
			if err := protojson.Unmarshal(raw, &saved); err != nil {
				t.Fatal(err)
			}
			contents := map[string]string{}
			for _, artifact := range saved.Artifacts {
				data := artifact.AsMap()["data"].(map[string]any)
				for _, item := range data["chunks"].([]any) {
					chunk := item.(map[string]any)
					contents[chunk["chunk_id"].(string)] = chunk["content"].(string)
				}
			}
			if !reflect.DeepEqual(contents, map[string]string{
				"initial": "Input opens the quantity popup.", "conditions": "Single-item input does not open that popup.",
				"branches": tc.content,
			}) {
				t.Fatalf("complete source lost original content: %v", contents)
			}
			if saved.Output.AsMap()["source_page"].(map[string]any)["has_more"] != false {
				t.Fatal("source read ended before terminal coverage")
			}
			view, err := os.ReadFile(output + ".txt")
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(view), "Single-item input does not open that popup.") || !strings.Contains(string(view), "Input opens the quantity popup.") || strings.Index(string(view), "Single-item input") > strings.Index(string(view), "Input opens") {
				t.Fatalf("reading view lost complete text or original document order: %s", view)
			}
			if !strings.Contains(string(view), `"index":1`) || !strings.Contains(string(view), `"index":0`) {
				t.Fatal("reading view lost original receipt selectors")
			}
			if tc.table {
				if !strings.Contains(string(view), "\n<tr><td>Initial entry</td><td>Automatic</td></tr>\n<tr><td>Later edit</td><td>Manual &amp; optional</td></tr>") {
					t.Fatalf("CLI reading view did not separate procedure branches into inspectable rows: %s", view)
				}
			} else if !strings.Contains(string(view), tc.content) {
				t.Fatalf("non-table reading view changed source text: %s", view)
			}
			for line := range strings.SplitSeq(string(view), "\n") {
				if strings.Contains(line, "PRD") && strings.HasPrefix(line, "===== PASSAGE") {
					t.Fatalf("filename lookup falsely matched passage content: %s", line)
				}
			}
			read := &sdk.KnowledgeReadEvidenceRequest{}
			if err := knowledgeRequestFromResult(read, []string{output}, 1, nil, true, ""); err != nil || read.ArtifactId != "artifact-conditions" || read.ChunkId != "conditions" || read.FileId != "manual" {
				t.Fatalf("combined receipt lost its original evidence selector: %v %v", read, err)
			}
		})
	}
}

func TestKnowledgeReadSourceRejectsAmbiguousSourceBeforeSearching(t *testing.T) {
	for _, inventory := range []string{
		`{"row_count":2,"files":[{"semantic_model_id":7,"file_id":"old","file_name":"Manual-v1"},{"semantic_model_id":7,"file_id":"new","file_name":"Manual-v2"}]}`,
		`{"row_count":2,"files":[{"semantic_model_id":7,"file_id":"old","file_name":"Manual-v1"}]}`,
		`{"row_count":0,"files":[]}`,
	} {
		t.Run(inventory, func(t *testing.T) {
			var searched atomic.Bool
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasSuffix(r.URL.Path, "/search") {
					searched.Store(true)
				}
				_, _ = w.Write([]byte(`{"code":"OK","data":{"output":` + inventory + `}}`))
			}))
			t.Cleanup(server.Close)
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("key"))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := client.Close(); err != nil {
					t.Error(err)
				}
			})
			output := filepath.Join(t.TempDir(), "ambiguous.json")
			printed := captureStdout(t, func() {
				err = ExecuteSemanticModelCommand(NewContext(client, "ws1", false, true, true), []string{"read-source", "--source-name", "Manual", "--output", output})
			})
			if err == nil || !strings.Contains(err.Error(), "CLI_KNOWLEDGE_SOURCE_SELECTION_REQUIRED") || searched.Load() {
				t.Fatalf("ambiguous name selected a source: %v searched=%t", err, searched.Load())
			}
			var displayed sdk.KnowledgeRuntimeResult
			if err := protojson.Unmarshal([]byte(printed), &displayed); err != nil {
				t.Fatalf("source selection discarded the available inventory: %v", err)
			}
			var expected map[string]any
			if err := json.Unmarshal([]byte(inventory), &expected); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(displayed.GetOutput().AsMap(), expected) {
				t.Fatalf("source selection changed the owner inventory: %v", displayed.GetOutput())
			}
			if _, err := os.Stat(output); !os.IsNotExist(err) {
				t.Fatalf("ambiguous source published a receipt: %v", err)
			}
		})
	}
}

func TestKnowledgeReadSourceRejectsConflictingOrPartialSourceSelectors(t *testing.T) {
	var contacted atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		contacted.Store(true)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("key"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Error(err)
		}
	})
	for _, args := range [][]string{
		{"read-source", "--source-name", " "},
		{"read-source", "--source-name", "Manual", "--json", `{"source_page":{"semantic_model_id":7,"file_id":"manual"}}`},
		{"read-source", "--source-name", "Manual", "--from", "missing-inventory.json", "--index", "0"},
		{"read-source", "--json", `{"source_page":{"semantic_model_id":7,"file_id":"manual","cursor":"later-page"}}`},
	} {
		err := ExecuteSemanticModelCommand(NewContext(client, "ws1", false, true, true), args)
		if err == nil || !strings.Contains(err.Error(), "CLI_KNOWLEDGE_INPUT_INVALID") || contacted.Load() {
			t.Fatalf("invalid selectors reached knowledge owner: %v contacted=%t", err, contacted.Load())
		}
	}
}

func TestKnowledgeReadSourceDoesNotPublishIncompleteOrMismatchedEvidence(t *testing.T) {
	for _, mode := range []string{"artifact identity", "artifact coverage", "cyclic cursor", "missing continuation", "later owner failure"} {
		t.Run(mode, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request sdk.KnowledgeSearchRequest
				raw := json.RawMessage{}
				if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
					t.Error(err)
					return
				}
				if err := protojson.Unmarshal(raw, &request); err != nil {
					t.Error(err)
					return
				}
				cursor := request.GetSourcePage().GetCursor()
				if mode == "later owner failure" && cursor == "next" {
					_, _ = w.Write([]byte(`{"code":"OK","data":{"failed":true,"output":{"code":"OWNER_SOURCE_READ_FAILED","message":"source unavailable"}}}`))
					return
				}
				coverage := map[string]any{"semantic_model_id": 7, "file_id": "manual", "cursor": cursor, "has_more": false}
				if mode == "cyclic cursor" || mode == "later owner failure" {
					coverage["has_more"], coverage["next_cursor"] = true, "next"
				} else if mode == "missing continuation" {
					coverage["has_more"] = true
				}
				artifactCoverage := maps.Clone(coverage)
				chunk := map[string]any{"semantic_model_id": 7, "file_id": "manual", "chunk_id": "text", "content": "canonical body"}
				if mode == "artifact identity" {
					chunk["file_id"] = "another-source"
				} else if mode == "artifact coverage" {
					artifactCoverage["file_id"] = "another-source"
				}
				result := map[string]any{"call_id": "page-" + cursor, "output": map[string]any{"source_page": coverage, "row_count": 1}, "artifacts": []any{map[string]any{"artifact_id": "artifact-" + cursor, "data": map[string]any{"source_page": artifactCoverage, "chunks": []any{chunk}}}}}
				if err := json.NewEncoder(w).Encode(map[string]any{"code": "OK", "data": result}); err != nil {
					t.Error(err)
				}
			}))
			t.Cleanup(server.Close)
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("key"))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := client.Close(); err != nil {
					t.Error(err)
				}
			})
			output := filepath.Join(t.TempDir(), "source.json")
			err = ExecuteSemanticModelCommand(NewContext(client, "ws1", false, true, true), []string{"read-source", "--json", `{"source_page":{"semantic_model_id":7,"file_id":"manual"}}`, "--output", output})
			want := "CLI_KNOWLEDGE_RESULT_INVALID"
			if mode == "later owner failure" {
				var failure *sdk.Error
				if !errors.As(err, &failure) || failure.Code != "OWNER_SOURCE_READ_FAILED" {
					t.Fatalf("source enumeration lost the owner error: %v", err)
				}
				want = "OWNER_SOURCE_READ_FAILED"
			}
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("incomplete source was accepted: %v", err)
			}
			if _, err := os.Stat(output); !os.IsNotExist(err) {
				t.Fatalf("failed enumeration published a complete receipt: %v", err)
			}
		})
	}
}
