package commands

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestKnowledgeSourceCommandHelpDescribesSourceSelectors(t *testing.T) {
	for _, operation := range []string{"read-source", "list-passages"} {
		t.Run(operation, func(t *testing.T) {
			file, err := os.CreateTemp(t.TempDir(), "help")
			if err != nil {
				t.Fatal(err)
			}
			defer file.Close()
			stderr := os.Stderr
			os.Stderr = file
			defer func() { os.Stderr = stderr }()
			err = ExecuteSemanticModelCommand(NewContext(nil, "ws1", false, true, true), []string{operation, "--help"})
			if !errors.Is(err, flag.ErrHelp) {
				t.Fatalf("help result: %v", err)
			}
			raw, err := os.ReadFile(file.Name())
			if err != nil {
				t.Fatal(err)
			}
			help := string(raw)
			for _, field := range []string{"source_page.semantic_model_id: int64", "source_page.file_id: string", "source_page.cursor: string", "source_page.page_size: int32"} {
				if !strings.Contains(help, field) {
					t.Errorf("help omitted source selector %s", field)
				}
			}
			for _, field := range []string{"keywords:", "file_ids:", "max_hits:", "before:", "after:"} {
				if strings.Contains(help, field) {
					t.Errorf("help advertises unsupported source-read input %s", field)
				}
			}
		})
	}
}

func TestKnowledgeSourceCommandSchemaDescribesSourceSelectors(t *testing.T) {
	for _, operation := range []string{"read-source", "list-passages"} {
		t.Run(operation, func(t *testing.T) {
			printed := captureStdout(t, func() {
				if err := ExecuteSchemaCommand([]string{"semantic-model", operation}); err != nil {
					t.Fatal(err)
				}
			})
			var description struct {
				Schema struct {
					Properties map[string]json.RawMessage `json:"properties"`
				} `json:"schema"`
			}
			if err := json.Unmarshal([]byte(printed), &description); err != nil {
				t.Fatal(err)
			}
			for _, field := range []string{"request_id", "source_page"} {
				if description.Schema.Properties[field] == nil {
					t.Errorf("schema omitted %s", field)
				}
			}
			for _, field := range []string{"keywords", "file_ids", "max_hits", "before", "after"} {
				if description.Schema.Properties[field] != nil {
					t.Errorf("source-only schema advertises %s", field)
				}
			}
		})
	}
}

func TestKnowledgeSearchSummaryPreservesRerankStatus(t *testing.T) {
	t.Chdir(t.TempDir())
	ranking := []any{map[string]any{"semantic_model_id": float64(7), "status": "degraded", "attempts": float64(2), "candidate_count": float64(1), "reason": "http_502", "status_code": float64(502)}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "OK", "data": map[string]any{
			"call_id": "ranked", "output": map[string]any{"rerank": ranking},
			"artifacts": []any{map[string]any{"artifact_id": "ranked", "data": map[string]any{"chunks": []any{map[string]any{"chunk_id": "chunk", "file_id": "report", "content": "Revenue grew."}}}}},
		}})
	}))
	defer server.Close()
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("test"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	printed := captureStdout(t, func() {
		if err := ExecuteSemanticModelCommand(NewContext(client, "ws1", false, true, true), []string{"search", "--keywords", "Revenue", "--output", "search.json"}); err != nil {
			t.Fatal(err)
		}
	})
	var summary struct {
		Rerank []any `json:"rerank"`
	}
	if err := json.Unmarshal([]byte(printed), &summary); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(summary.Rerank, ranking) {
		t.Fatalf("CLI summary lost server ranking facts: %+v", summary.Rerank)
	}
}

func TestKnowledgeFailurePreservesOwnerErrorContract(t *testing.T) {
	for _, field := range []string{"code", "reason"} {
		t.Run(field, func(t *testing.T) {
			t.Chdir(t.TempDir())
			const ownerCode = "KNOWLEDGE_DOCUMENT_LOCATOR_NOT_FOUND"
			ownerMessage := "localized owner message"
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode(map[string]any{"code": "OK", "data": map[string]any{
					"failed": true, "call_id": "rejected", "output": map[string]any{field: ownerCode, "message": ownerMessage, "field": "markdown_file_id"},
				}})
			}))
			t.Cleanup(server.Close)
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("test"))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = client.Close() })
			captureStdout(t, func() {
				err = ExecuteSemanticModelCommand(NewContext(client, "ws", false, true, true), []string{"read-document", "--json", `{"markdown_file_id":"unknown"}`, "--output", "result.json"})
			})
			var failure *sdk.Error
			if !errors.As(err, &failure) || failure.Code != ownerCode || failure.Message != ownerMessage || failure.Details["field"] != "markdown_file_id" {
				t.Fatalf("CLI replaced the owner error contract: %v", err)
			}
			if _, err := os.Stat("result.json"); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("rejection produced a successful saved result: %v", err)
			}
		})
	}
}

func TestKnowledgeSearchExposesApplicableSemanticDefinitions(t *testing.T) {
	t.Chdir(t.TempDir())
	rule := map[string]any{"id": 23, "model_id": 7, "domain": "rag", "kind": "business_rule", "key": "revenue-unit", "rule": "Revenue is expressed in CNY."}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "OK", "data": map[string]any{
			"call_id": "rules", "output": map[string]any{"semantic_entries": []any{rule}},
			"artifacts": []any{map[string]any{"artifact_id": "rules", "data": map[string]any{"chunks": []any{map[string]any{"chunk_id": "revenue", "file_id": "report", "content": "Revenue grew by 10 percent."}}}}},
		}})
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
	printed := captureStdout(t, func() {
		if err := runKnowledgeRuntime(NewContext(client, "ws1", false, true, true), "search", []string{"--keywords", "revenue", "--output", "rules.json"}); err != nil {
			t.Fatal(err)
		}
	})
	var summary knowledgeResultSummary
	if err := json.Unmarshal([]byte(printed), &summary); err != nil {
		t.Fatal(err)
	}
	args := summary.ReadCommands["semantic_entries"]
	if len(args) == 0 {
		t.Fatalf("search summary hides applicable semantic definitions: %s", printed)
	}
	raw, err := exec.Command(args[0], args[1:]...).Output()
	if err != nil {
		t.Fatal(err)
	}
	var actual map[string]any
	if err := json.Unmarshal(raw, &actual); err != nil {
		t.Fatal(err)
	}
	if actual["key"] != rule["key"] || actual["rule"] != rule["rule"] || actual["domain"] != rule["domain"] {
		t.Fatalf("semantic reading command lost the owner's definition: %s", raw)
	}
}

func TestKnowledgeRepeatedKeywordFlagsPreserveEveryRequestedQuery(t *testing.T) {
	for _, operation := range []string{"search", "describe-tables"} {
		t.Run(operation, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request struct {
					Keywords []string `json:"keywords"`
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Error(err)
					return
				}
				if !slices.Equal(request.Keywords, []string{"復旧条件", "quota", "rate limit", "quota"}) {
					t.Errorf("repeated flags lost requested queries: %v", request.Keywords)
				}
				_, _ = w.Write([]byte(`{"code":"OK","data":{"output":{}}}`))
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
			if err := ExecuteSemanticModelCommand(NewContext(client, "ws1", false, true, true), []string{operation, "--keywords", "復旧条件,quota", "--keywords", "rate limit", "--keywords", "quota"}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestKnowledgeBatchReadsPreserveSelectorsAndStopOnOwnerFailure(t *testing.T) {
	for _, test := range []struct {
		name, rejectID string
		localRead      bool
		expected       []string
	}{
		{"local complete", "", true, []string{"first", "second"}},
		{"local first stale", "first", true, []string{"first"}},
		{"local second stale", "second", true, []string{"first", "second"}},
		{"remote complete", "", false, []string{"first", "second"}},
		{"remote first stale", "first", false, []string{"first"}},
		{"remote second stale", "second", false, []string{"first", "second"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			firstID := "first"
			var seen []string
			ids := map[string]string{}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/read-evidence") {
					t.Errorf("unexpected batch read route: %s %s", r.Method, r.URL.Path)
				}
				var request sdk.KnowledgeReadEvidenceRequest
				raw, err := io.ReadAll(r.Body)
				if err != nil || protojson.Unmarshal(raw, &request) != nil {
					t.Errorf("invalid read request: %s %v", raw, err)
					return
				}
				inputs := []*sdk.KnowledgeLocalEvidenceRead{{ArtifactId: request.ArtifactId, ChunkId: request.ChunkId, FileId: request.FileId, SemanticModelId: request.SemanticModelId, LocalContentSha256: request.LocalContentSha256}}
				if test.localRead {
					inputs = request.LocalReads
					if request.RequestId != "batch" || request.ChunkId != "" || request.ArtifactId != "" {
						t.Errorf("local batch lost atomic operation identity: %s", raw)
					}
				}
				var outputs []any
				for _, input := range inputs {
					if input.ChunkId != firstID && input.ChunkId != "second" {
						t.Errorf("batch changed selector: %s", input.ChunkId)
					}
					if input.ArtifactId != "saved-artifact" || input.FileId != "manual" || input.GetSemanticModelId() != 7 {
						t.Errorf("batch changed provenance: %s", raw)
					}
					content := map[string]string{firstID: "条件① & literal", "second": "Later state"}[input.ChunkId]
					digest := sha256.Sum256([]byte(content))
					if (test.localRead && input.LocalContentSha256 != hex.EncodeToString(digest[:])) || (!test.localRead && input.LocalContentSha256 != "") {
						t.Errorf("batch changed inspected content: %s", raw)
					}
					seen = append(seen, input.ChunkId)
					ids[input.ChunkId] = request.RequestId
					if input.ChunkId == test.rejectID {
						_, _ = w.Write([]byte(`{"code":"OK","data":{"failed":true,"output":{"reason":"CURRENT_TASK_EVIDENCE_REQUIRED"}}}`))
						return
					}
					output := map[string]any{"kind": "rag_local_evidence_read", "artifact_id": input.ArtifactId, "chunk_id": input.ChunkId, "file_id": input.FileId, "semantic_model_id": input.GetSemanticModelId(), "content_sha256": input.LocalContentSha256}
					if !test.localRead {
						output = map[string]any{"chunk_id": input.ChunkId, "content": content, "cursor": 0, "next_cursor": len([]rune(content)), "has_more": false}
					}
					outputs = append(outputs, output)
				}
				var output any = outputs[0]
				if test.localRead {
					output = map[string]any{"kind": "rag_local_evidence_read_batch", "reads": outputs}
				}
				if err := json.NewEncoder(w).Encode(map[string]any{"code": "OK", "data": map[string]any{"call_id": request.RequestId, "output": output}}); err != nil {
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
			directory := t.TempDir()
			path, output := filepath.Join(directory, "evidence.json"), filepath.Join(directory, "reads.json")
			chunks := []any{map[string]any{"chunk_id": firstID, "file_id": "manual", "semantic_model_id": 7, "content": "条件① & literal"}, map[string]any{"chunk_id": "second", "file_id": "manual", "semantic_model_id": 7, "content": "Later state"}}
			saved, err := json.Marshal(map[string]any{"artifacts": []any{map[string]any{"artifact_id": "saved-artifact", "data": map[string]any{"chunks": chunks}}}})
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, saved, 0600); err != nil {
				t.Fatal(err)
			}
			args := []string{"read-evidence", "--from", path, "--indices", "0,1", "--request-id", "batch"}
			if test.localRead {
				args = append(args, "--local-read")
			}
			printed := captureStdout(t, func() {
				err = ExecuteSemanticModelCommand(NewContext(client, "ws1", false, true, true), append(slices.Clone(args), "--output", output))
			})
			if test.rejectID != "" {
				if err == nil || !slices.Equal(seen, test.expected) || !strings.Contains(printed, "CURRENT_TASK_EVIDENCE_REQUIRED") {
					t.Fatalf("batch did not expose the owner's rejection and stop: %v %v %s", err, seen, printed)
				}
				if _, err := os.Stat(output); !os.IsNotExist(err) {
					t.Fatalf("failed batch published success: %v", err)
				}
				return
			}
			if err != nil || !slices.Equal(seen, []string{"first", "second"}) || ids["first"] == "" || (test.localRead != (ids["first"] == ids["second"])) {
				t.Fatalf("batch lost requested reads or replay identity: %v %v %v", err, seen, ids)
			}
			var summary struct {
				ReceiptIndices []int               `json:"receipt_indices"`
				ReadCommands   map[string][]string `json:"read_commands"`
			}
			if err := json.Unmarshal([]byte(printed), &summary); err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(summary.ReceiptIndices, []int{0, 1}) {
				t.Fatalf("receipt omitted selectors for first and second evidence: %s", printed)
			}
			if !test.localRead {
				args := summary.ReadCommands["content"]
				if len(args) == 0 {
					t.Fatal("saved server reads have no executable content reader")
				}
				inspected, err := exec.CommandContext(t.Context(), args[0], args[1:]...).CombinedOutput()
				if err != nil {
					t.Fatalf("inspect saved reads: %v: %s", err, inspected)
				}
				decoder := json.NewDecoder(strings.NewReader(string(inspected)))
				for index, content := range []string{"条件① & literal", "Later state"} {
					var read struct {
						Index      int `json:"index"`
						InputIndex int `json:"input_index"`
						Evidence   struct {
							Content string `json:"content"`
							HasMore bool   `json:"has_more"`
						} `json:"evidence"`
					}
					if err := decoder.Decode(&read); err != nil || read.Index != index || read.InputIndex != index || read.Evidence.Content != content || read.Evidence.HasMore {
						t.Fatalf("inspection lost a passage, its selector or coverage: %+v %v", read, err)
					}
				}
			}
			body, err := os.ReadFile(output)
			var receipt struct {
				Output struct {
					Kind  string `json:"kind"`
					Reads []struct {
						Index  int `json:"input_index"`
						Output struct {
							ChunkID string `json:"chunk_id"`
							Content string `json:"content"`
						} `json:"output"`
					} `json:"reads"`
				} `json:"output"`
			}
			if err != nil || json.Unmarshal(body, &receipt) != nil {
				t.Fatalf("batch did not retain its read receipts: %v %s", err, body)
			}
			var actualIDs []string
			var actualIndices []int
			var actualContent []string
			for _, read := range receipt.Output.Reads {
				actualIDs = append(actualIDs, read.Output.ChunkID)
				actualIndices = append(actualIndices, read.Index)
				actualContent = append(actualContent, read.Output.Content)
			}
			expectedKind := "knowledge_read_batch"
			if test.localRead {
				expectedKind = "knowledge_local_read_batch"
			}
			if !test.localRead && !slices.Equal(actualContent, []string{"条件① & literal", "Later state"}) {
				t.Fatalf("batch lost owner-returned evidence text: %s", body)
			}
			if receipt.Output.Kind != expectedKind || !slices.Equal(actualIDs, []string{"first", "second"}) || !slices.Equal(actualIndices, []int{0, 1}) {
				t.Fatalf("batch receipts lost original selectors: %s", body)
			}
			firstIDs := map[string]string{"first": ids["first"], "second": ids["second"]}
			seen = nil
			captureStdout(t, func() {
				err = ExecuteSemanticModelCommand(NewContext(client, "ws1", false, true, true), append(slices.Clone(args), "--output", filepath.Join(directory, "replayed.json")))
			})
			if err != nil || !slices.Equal(seen, test.expected) || firstIDs["first"] != ids["first"] || firstIDs["second"] != ids["second"] {
				t.Fatalf("batch replay changed its operation identities: %v %v %v", err, firstIDs, ids)
			}
		})
	}
}

func TestKnowledgeBatchLocalReadValidatesAllInputsBeforeCallingOwner(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "saved.json")
	if err := os.WriteFile(path, []byte(`{"artifacts":[{"artifact_id":"saved","data":{"chunks":[{"chunk_id":"first","content":"complete first passage"},{"chunk_id":"second","content":"complete second passage"}]}}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name, code string
		args       []string
	}{
		{"bad later index", "CLI_KNOWLEDGE_INDEX_OUT_OF_RANGE", []string{"--from", path, "--indices", "0,2", "--local-read"}},
		{"negative index", "CLI_KNOWLEDGE_INDEX_OUT_OF_RANGE", []string{"--from", path, "--indices", "0,-1", "--local-read"}},
		{"not an integer", "CLI_KNOWLEDGE_INPUT_INVALID", []string{"--from", path, "--indices", "0,second", "--local-read"}},
		{"bad later remote index", "CLI_KNOWLEDGE_INDEX_OUT_OF_RANGE", []string{"--from", path, "--indices", "0,2"}},
		{"mixed selectors", "CLI_KNOWLEDGE_INPUT_INVALID", []string{"--from", path, "--indices", "0,1", "--index", "0", "--local-read"}},
		{"multiple files", "CLI_KNOWLEDGE_INPUT_INVALID", []string{"--from", path, "--from", path, "--indices", "0,1", "--local-read"}},
		{"empty selection", "CLI_KNOWLEDGE_INPUT_INVALID", []string{"--from", path, "--indices", "", "--local-read"}},
		{"no saved file", "CLI_KNOWLEDGE_INPUT_INVALID", []string{"--indices", "0,1", "--local-read"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				t.Error("invalid batch must not register any partial local reads")
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
			err = ExecuteSemanticModelCommand(NewContext(client, "ws1", false, true, true), append([]string{"read-evidence"}, test.args...))
			if err == nil || !strings.Contains(err.Error(), test.code) {
				t.Fatalf("invalid batch lost actionable input error: %v", err)
			}
		})
	}
}

func TestKnowledgeFulltextOnlyUsesTypedSearchRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var raw json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			t.Error(err)
			return
		}
		var request sdk.KnowledgeSearchRequest
		if err := protojson.Unmarshal(raw, &request); err != nil {
			t.Error(err)
			return
		}
		if !request.GetFulltextOnly() || !slices.Equal(request.GetKeywords(), []string{"distinctive-term"}) || request.GetSourcePage() != nil {
			t.Errorf("full-text verification changed mode or terms: %v", &request)
		}
		_, _ = w.Write([]byte(`{"code":"OK","data":{"output":{"routes":["fulltext"],"chunks":[]}}}`))
	}))
	defer server.Close()
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("key"), sdk.WithRuntimeAuthorization("Bearer task"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	if err := runKnowledgeRuntime(NewContext(client, "ws1", false, true, true), "search", []string{"--fulltext-only", "--keywords", "distinctive-term"}); err != nil {
		t.Fatal(err)
	}
}

func TestKnowledgeSearchCarriesBoundedContextAndHonorsExplicitWindows(t *testing.T) {
	for _, test := range []struct {
		name          string
		input         []string
		before, after int32
	}{
		{"default neighboring context", nil, 2, 1},
		{"explicit atomic read", []string{"--json", `{"before":0,"after":0}`}, 0, 0},
		{"explicit wider context", []string{"--json", `{"before":4,"after":2}`}, 4, 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var raw json.RawMessage
				if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
					t.Error(err)
					return
				}
				var request sdk.KnowledgeSearchRequest
				if err := protojson.Unmarshal(raw, &request); err != nil {
					t.Error(err)
					return
				}
				if request.GetBefore() != test.before || request.GetAfter() != test.after || !slices.Equal(request.GetKeywords(), []string{"distinctive-term"}) {
					t.Errorf("search lost its context window or terms: %v", &request)
				}
				_, _ = w.Write([]byte(`{"code":"OK","data":{"output":{"routes":["fulltext"],"chunks":[]}}}`))
			}))
			defer server.Close()
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("key"), sdk.WithRuntimeAuthorization("Bearer task"))
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()
			args := append([]string{"--fulltext-only", "--keywords", "distinctive-term"}, test.input...)
			if err := runKnowledgeRuntime(NewContext(client, "ws1", false, true, true), "search", args); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestKnowledgeSourcePageUsesSavedScopeAndExplicitContinuation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var data json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			t.Error(err)
			return
		}
		var request sdk.KnowledgeSearchRequest
		if err := protojson.Unmarshal(data, &request); err != nil {
			t.Error(err)
			return
		}
		page := request.GetSourcePage()
		if r.URL.Path != "/newmoi/semantic-models/runtime/search" || page == nil || page.GetSemanticModelId() != 5000000000 || page.GetFileId() != "bound-file" || page.GetCursor() != "next-page" || page.GetPageSize() != 1 || len(request.GetKeywords()) != 0 || len(request.GetFileIds()) != 0 || request.GetMaxHits() != 0 {
			t.Errorf("source page lost its scope/mode: %v", &request)
		}
		_, _ = w.Write([]byte(`{"code":"OK","data":{"call_id":"source-page","output":{"source_page":{"has_more":false}}}}`))
	}))
	defer server.Close()
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("key"), sdk.WithRuntimeAuthorization("Bearer task"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	path := filepath.Join(t.TempDir(), "inventory.json")
	if err := os.WriteFile(path, []byte(`{"output":{"files":[{"semantic_model_id":5000000000,"file_id":"bound-file"}]}}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := runKnowledgeRuntime(NewContext(client, "ws1", false, true, true), "list-passages", []string{"--from", path, "--index", "0", "--cursor", "next-page", "--page-size", "1"}); err != nil {
		t.Fatal(err)
	}
}

func TestKnowledgeSourcePageContinuesFromSavedPageWithoutCopyingSelectors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var raw json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			t.Error(err)
			return
		}
		var request sdk.KnowledgeSearchRequest
		if err := protojson.Unmarshal(raw, &request); err != nil {
			t.Error(err)
			return
		}
		page := request.GetSourcePage()
		if page.GetFileId() != "bound-file" || page.GetSemanticModelId() != 5000000000 || page.GetCursor() != "opaque-next-page" || page.GetPageSize() != 1 {
			t.Errorf("saved continuation lost its exact source/cursor: %v", &request)
		}
		_, _ = w.Write([]byte(`{"code":"OK","data":{"output":{"source_page":{"has_more":false}}}}`))
	}))
	defer server.Close()
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("key"), sdk.WithRuntimeAuthorization("Bearer task"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	path := filepath.Join(t.TempDir(), "previous-page.json")
	if err := os.WriteFile(path, []byte(`{"output":{"source_page":{"semantic_model_id":5000000000,"file_id":"bound-file","cursor":"previous-cursor","has_more":true,"next_cursor":"opaque-next-page"}}}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := runKnowledgeRuntime(NewContext(client, "ws1", false, true, true), "list-passages", []string{"--from", path, "--page-size", "1"}); err != nil {
		t.Fatal(err)
	}
}

func TestKnowledgeSourcePageRejectsFinishedAndAmbiguousContinuations(t *testing.T) {
	for _, example := range []struct {
		name, page, want string
		args             []string
	}{
		{name: "finished", page: `{"semantic_model_id":7,"file_id":"bound-file","has_more":false}`, want: "CLI_KNOWLEDGE_SOURCE_COMPLETE"},
		{name: "missing cursor", page: `{"semantic_model_id":7,"file_id":"bound-file","has_more":true}`, want: "CLI_KNOWLEDGE_RESULT_INVALID"},
		{name: "manual cursor", page: `{"semantic_model_id":7,"file_id":"bound-file","has_more":true,"next_cursor":"next"}`, args: []string{"--cursor", "different"}, want: "CLI_KNOWLEDGE_INPUT_INVALID"},
		{name: "manual index", page: `{"semantic_model_id":7,"file_id":"bound-file","has_more":true,"next_cursor":"next"}`, args: []string{"--index", "0"}, want: "CLI_KNOWLEDGE_INPUT_INVALID"},
	} {
		t.Run(example.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "page.json")
			if err := os.WriteFile(path, []byte(`{"output":{"source_page":`+example.page+`}}`), 0600); err != nil {
				t.Fatal(err)
			}
			request := &sdk.KnowledgeSearchRequest{SourcePage: new(sdk.KnowledgeSourcePage)}
			index := -1
			if slices.Contains(example.args, "--index") {
				index = 0
			}
			if slices.Contains(example.args, "--cursor") {
				request.SourcePage.Cursor = "different"
			}
			err := knowledgeRequestFromResult(request, []string{path}, index, nil, false, "")
			if err == nil || !strings.Contains(err.Error(), example.want) {
				t.Fatalf("continuation should fail explicitly with %s, got %v", example.want, err)
			}
		})
	}
}

func TestKnowledgeCitationsSupportPerFileAndFlattenedSelectors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Sources []map[string]any `json:"sources"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
			return
		}
		var selected []string
		for _, source := range request.Sources {
			selected = append(selected, source["artifact_id"].(string)+":"+source["chunk_id"].(string))
		}
		if !slices.Equal(selected, []string{"first:chosen-a", "second:chosen-b"}) {
			t.Errorf("citation selectors refer to different evidence: %v", selected)
		}
		_, _ = w.Write([]byte(`{"code":"OK","data":{"call_id":"cite","output":{"accepted":true}}}`))
	}))
	defer server.Close()
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("key"), sdk.WithRuntimeAuthorization("Bearer task"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	dir := t.TempDir()
	a, b := filepath.Join(dir, "first.json"), filepath.Join(dir, "second.json")
	if err := os.WriteFile(a, []byte(`{"artifacts":[{"artifact_id":"first","data":{"chunks":[{"chunk_id":"not-selected","file_id":"a"},{"chunk_id":"chosen-a","file_id":"a"}]}}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte(`{"artifacts":[{"artifact_id":"second","data":{"chunks":[{"chunk_id":"chosen-b","file_id":"b"}]}}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"--from", a, "--indices", "1", "--from", b, "--indices", "0"},
		{"--from", a, "--from", b, "--indices", "1,2"},
	} {
		if err := runKnowledgeRuntime(NewContext(client, "ws1", false, true, true), "cite", args); err != nil {
			t.Fatal(err)
		}
	}
}

func TestKnowledgeLocalReadPreservesSavedContentAndServerVerification(t *testing.T) {
	content := "表の原文😀\n引用符 \" と \\"
	digest := sha256.Sum256([]byte(content))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body sdk.KnowledgeReadEvidenceRequest
		raw := json.RawMessage{}
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			t.Error(err)
		}
		if err := protojson.Unmarshal(raw, &body); err != nil {
			t.Error(err)
		}
		if r.URL.Path != "/newmoi/semantic-models/runtime/read-evidence" || body.GetLocalContentSha256() != hex.EncodeToString(digest[:]) || body.GetArtifactId() != "search" || body.GetChunkId() != "table" || body.GetFileId() != "file" {
			t.Errorf("local read lost its exact identity/content: %v", &body)
		}
		_, _ = w.Write([]byte(`{"code":"OK","data":{"call_id":"local-read","output":{"kind":"rag_local_evidence_read"}}}`))
	}))
	defer server.Close()
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("key"), sdk.WithRuntimeAuthorization("Bearer task"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	artifact, err := structpb.NewStruct(map[string]any{"artifact_id": "search", "data": map[string]any{"chunks": []any{map[string]any{"chunk_id": "table", "file_id": "file", "content": content}}}})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := protojson.Marshal(&sdk.KnowledgeRuntimeResult{Artifacts: []*structpb.Struct{artifact}})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "search.json")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	if err := runKnowledgeRuntime(NewContext(client, "ws1", false, true, true), "read-evidence", []string{"--from", path, "--index", "0", "--local-read"}); err != nil {
		t.Fatal(err)
	}
}

func TestKnowledgeReadContinuationAdvancesPastReceiptCursor(t *testing.T) {
	path := filepath.Join(t.TempDir(), "receipt.json")
	if err := os.WriteFile(path, []byte(`{"output":{"sources":[{"next_call":{"tool_id":"read_artifact","arguments":{"artifact_id":"retrieved","chunk_id":"table","cursor":4000}}}]}}`), 0600); err != nil {
		t.Fatal(err)
	}
	request := &sdk.KnowledgeReadEvidenceRequest{Cursor: structpb.NewNumberValue(6000)}
	if err := knowledgeRequestFromResult(request, []string{path}, 0, nil, false, ""); err != nil {
		t.Fatal(err)
	}
	if request.GetCursor().GetNumberValue() != 6000 || request.GetArtifactId() != "retrieved" || request.GetChunkId() != "table" {
		t.Fatalf("continuation restarted or lost its source: %v", request)
	}
}

func TestKnowledgeSavedReadReceiptsContinueAndCite(t *testing.T) {
	for _, local := range []bool{false, true} {
		t.Run(map[bool]string{false: "bounded", true: "local"}[local], func(t *testing.T) {
			directory := t.TempDir()
			path := filepath.Join(directory, "search.json")
			saved := `{"artifacts":[{"artifact_id":"search","data":{"chunks":[{"chunk_id":"a","file_id":"file-a","semantic_model_id":7,"content":"日本語😀"},{"chunk_id":"b","file_id":"file-b","semantic_model_id":8,"content":"Other"}]}}]}`
			if err := os.WriteFile(path, []byte(saved), 0600); err != nil {
				t.Fatal(err)
			}
			var cited []map[string]any
			var cursors []float64
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var request map[string]any
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Error(err)
					return
				}
				if strings.HasSuffix(r.URL.Path, "/cite") {
					for _, item := range request["sources"].([]any) {
						cited = append(cited, item.(map[string]any))
					}
					_, _ = w.Write([]byte(`{"code":"OK","data":{"output":{"accepted":true}}}`))
					return
				}
				if local {
					var receipts []any
					for _, raw := range request["local_reads"].([]any) {
						read := raw.(map[string]any)
						receipts = append(receipts, map[string]any{"kind": "rag_local_evidence_read", "artifact_id": read["artifact_id"], "chunk_id": read["chunk_id"], "file_id": read["file_id"], "semantic_model_id": read["semantic_model_id"], "content_sha256": read["local_content_sha256"]})
					}
					_ = json.NewEncoder(w).Encode(map[string]any{"code": "OK", "data": map[string]any{"call_id": "read", "output": map[string]any{"kind": "rag_local_evidence_read_batch", "reads": receipts}}})
					return
				}
				output := map[string]any{"artifact_id": "search", "chunk_id": request["chunk_id"], "file_id": request["file_id"], "semantic_model_id": request["semantic_model_id"]}
				if local {
					output["kind"], output["content_sha256"] = "rag_local_evidence_read", request["local_content_sha256"]
				} else {
					cursor, _ := request["cursor"].(float64)
					cursors = append(cursors, cursor)
					output["content"], output["cursor"], output["next_cursor"], output["has_more"], output["content_chars"] = "日本", cursor, float64(2), cursor == 0, float64(4)
					if cursor > 0 {
						output["content"], output["next_cursor"] = "語😀", float64(4)
					}
				}
				_ = json.NewEncoder(w).Encode(map[string]any{"code": "OK", "data": map[string]any{"call_id": "read", "output": output}})
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
			run := func(args ...string) {
				t.Helper()
				captureStdout(t, func() { err = ExecuteSemanticModelCommand(NewContext(client, "ws1", false, true, true), args) })
				if err != nil {
					t.Fatal(err)
				}
			}
			batch := filepath.Join(directory, "batch.json")
			args := []string{"read-evidence", "--from", path, "--indices", "1,0", "--output", batch}
			if local {
				args = append(args, "--local-read")
			}
			run(args...)
			// Indices belong to the saved receipt, not the original search order.
			run("cite", "--from", batch, "--indices", "0,1")
			want := []map[string]any{{"type": "rag_chunk", "artifact_id": "search", "chunk_id": "b", "file_id": "file-b", "semantic_model_id": float64(8)}, {"type": "rag_chunk", "artifact_id": "search", "chunk_id": "a", "file_id": "file-a", "semantic_model_id": float64(7)}}
			gotJSON, _ := json.Marshal(cited)
			wantJSON, _ := json.Marshal(want)
			if string(gotJSON) != string(wantJSON) {
				t.Fatalf("citation changed provenance or leaked read-only fields: %s", gotJSON)
			}
			if !local {
				continued := filepath.Join(directory, "continued.json")
				run("read-evidence", "--from", batch, "--index", "1", "--output", continued)
				if !slices.Equal(cursors, []float64{0, 0, 2}) {
					t.Fatalf("continuation restarted: %v", cursors)
				}
				cited = nil
				run("cite", "--from", continued, "--indices", "0")
				if cited[0]["chunk_id"] != "a" {
					t.Fatalf("single receipt lost provenance: %v", cited)
				}
				captureStdout(t, func() {
					err = ExecuteSemanticModelCommand(NewContext(client, "ws1", false, true, true), []string{"read-evidence", "--from", continued, "--index", "0"})
				})
				if err == nil || !strings.Contains(err.Error(), "CLI_KNOWLEDGE_SOURCE_COMPLETE") {
					t.Fatalf("completed receipt restarted: %v", err)
				}
			}
		})
	}
}

func TestKnowledgeReadReceiptDoesNotAuthorizeOrRegisterPartialContent(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "read.json")
	receipt := `{"output":{"artifact_id":"retrieved","chunk_id":"partial","file_id":"manual","semantic_model_id":7,"content":"a prefix","cursor":0,"next_cursor":8,"has_more":true,"content_chars":100}}`
	if err := os.WriteFile(path, []byte(receipt), 0600); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/cite") {
			t.Error("receipt must not register partial content")
			return
		}
		_, _ = w.Write([]byte(`{"code":"OK","data":{"failed":true,"output":{"reason":"CURRENT_TASK_EVIDENCE_REQUIRED"}}}`))
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
	ctx := NewContext(client, "ws1", false, true, true)
	printed := captureStdout(t, func() {
		err = ExecuteSemanticModelCommand(ctx, []string{"cite", "--from", path, "--indices", "0", "--output", filepath.Join(directory, "cited.json")})
	})
	if err == nil || !strings.Contains(printed, "CURRENT_TASK_EVIDENCE_REQUIRED") {
		t.Fatalf("receipt bypassed owner rejection: %s %v", printed, err)
	}
	if _, err := os.Stat(filepath.Join(directory, "cited.json")); !os.IsNotExist(err) {
		t.Fatalf("rejected receipt published success: %v", err)
	}
	err = ExecuteSemanticModelCommand(ctx, []string{"read-evidence", "--from", path, "--index", "0", "--local-read"})
	if err == nil || !strings.Contains(err.Error(), "CLI_KNOWLEDGE_INPUT_INVALID") {
		t.Fatalf("partial content was registered as complete: %v", err)
	}
	for _, malformed := range []string{
		`{"output":{"kind":"knowledge_read_batch","reads":[]}}`,
		`{"output":{"kind":"knowledge_read_batch","reads":[{"output":{"content":"unattributed"}}]}}`,
		`{"output":{"kind":"rag_local_evidence_read","artifact_id":"missing-source"}}`,
		`{"output":{"artifact_id":"retrieved","chunk_id":"partial","file_id":"manual","semantic_model_id":7.5,"has_more":true,"next_cursor":8}}`,
	} {
		if err := os.WriteFile(path, []byte(malformed), 0600); err != nil {
			t.Fatal(err)
		}
		err := knowledgeRequestFromResult(&sdk.KnowledgeCiteRequest{}, []string{path}, -1, []string{"0"}, false, "")
		if err == nil || !strings.Contains(err.Error(), "CLI_KNOWLEDGE_RESULT_INVALID") {
			t.Fatalf("malformed receipt accepted: %s %v", malformed, err)
		}
	}
}

func TestKnowledgeResultPublicationIsCompleteAndExclusive(t *testing.T) {
	path := filepath.Join(t.TempDir(), "results", "search.json")
	original := []byte(`{"content":"日本語😀"}`)
	if err := writeCommandResultFile(path, original, false); err != nil {
		t.Fatal(err)
	}
	if err := writeCommandResultFile(path, []byte(`{"content":"changed"}`), false); !os.IsExist(err) {
		t.Fatalf("overwrite error=%v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != string(original) {
		t.Fatalf("result=%s error=%v", got, err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("mode=%v error=%v", info, err)
	}
}

func TestKnowledgeEvidenceUsesSavedIdentity(t *testing.T) {
	artifact, err := structpb.NewStruct(map[string]any{"artifact_id": "opaque-id-should-never-be-retyped", "data": map[string]any{"chunks": []any{map[string]any{"semantic_model_id": 5000000000, "chunk_id": "c1", "file_id": "f1", "markdown_file_id": "m1", "content": "日本語😀"}}}})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := protojson.Marshal(&sdk.KnowledgeRuntimeResult{Artifacts: []*structpb.Struct{artifact}})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "search.json")
	if err := os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	read := &sdk.KnowledgeReadEvidenceRequest{}
	if err := knowledgeRequestFromResult(read, []string{path}, 0, nil, false, ""); err != nil {
		t.Fatal(err)
	}
	if read.GetArtifactId() != "opaque-id-should-never-be-retyped" || read.GetChunkId() != "c1" || read.GetSemanticModelId() != 5000000000 {
		t.Fatalf("read=%v", read)
	}
	if err := knowledgeRequestFromResult(read, []string{path}, 1, nil, false, ""); err == nil {
		t.Fatal("out-of-range index was accepted")
	}
	part := int32(0)
	page := &sdk.KnowledgeReadEvidenceRequest{PartIndex: &part}
	if err := knowledgeRequestFromResult(page, []string{path}, 0, nil, false, ""); err != nil {
		t.Fatal(err)
	}
	if page.GetArtifactId() != read.GetArtifactId() || page.GetChunkId() != "" || page.GetFileId() != "" || page.SemanticModelId != nil {
		t.Fatalf("page reader received chunk-only fields: %v", page)
	}
	cite := &sdk.KnowledgeCiteRequest{}
	second := filepath.Join(t.TempDir(), "second.json")
	if err := os.WriteFile(second, []byte(strings.ReplaceAll(string(raw), `"c1"`, `"c2"`)), 0600); err != nil {
		t.Fatal(err)
	}
	if err := knowledgeRequestFromResult(cite, []string{path, second}, -1, []string{"0", "1"}, false, ""); err != nil {
		t.Fatal(err)
	}
	var cited []string
	for _, source := range cite.GetSources() {
		cited = append(cited, source.AsMap()["chunk_id"].(string))
	}
	if !slices.Equal(cited, []string{"c1", "c2"}) {
		t.Fatalf("sources=%v", cite.GetSources())
	}
}

func TestKnowledgeSearchRetainsRuntimeIdentityAndCompleteFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/newmoi/semantic-models/runtime/search" || r.Header.Get("X-Workspace-ID") != "ws1" || r.Header.Get("X-MOI-Runtime-Authorization") != "Bearer task" {
			t.Errorf("incorrect request identity or route")
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		if body["request_id"] == "" {
			t.Error("missing replay identity")
		}
		_, _ = w.Write([]byte(`{"code":"OK","data":{"call_id":"saved-call","output":{"item_count":1},"artifacts":[{"artifact_id":"a1","data":{"chunks":[{"content":"完全な証拠😀"}]}}]}}`))
	}))
	defer server.Close()
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("key"), sdk.WithRuntimeAuthorization("Bearer task"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	path := filepath.Join(t.TempDir(), "search.json")
	if err := runKnowledgeRuntime(NewContext(client, "ws1", false, true, true), "search", []string{"--keywords", "test", "--output", path}); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	result := &sdk.KnowledgeRuntimeResult{}
	if err := protojson.Unmarshal(raw, result); err != nil {
		t.Fatal(err)
	}
	chunks := result.GetArtifacts()[0].AsMap()["data"].(map[string]any)["chunks"].([]any)
	if chunks[0].(map[string]any)["content"] != "完全な証拠😀" {
		t.Fatalf("lost content: %s", raw)
	}
}

func TestKnowledgeInventoryScopesSearchWithoutInventingEvidence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "files.json")
	if err := os.WriteFile(path, []byte(`{"call_id":"inventory","output":{"files":[{"file_id":"exact-file-id","file_name":"Guide.md","semantic_model_id":7}]}}`), 0600); err != nil {
		t.Fatal(err)
	}
	search := &sdk.KnowledgeSearchRequest{Keywords: []string{"retention"}}
	if err := knowledgeRequestFromResult(search, []string{path}, 0, nil, false, ""); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(search.GetFileIds(), []string{"exact-file-id"}) || !slices.Equal(search.GetKeywords(), []string{"retention"}) {
		t.Fatalf("inventory search lost its source or query: %v", search)
	}
	if err := knowledgeRequestFromResult(&sdk.KnowledgeReadEvidenceRequest{}, []string{path}, 0, nil, false, ""); err == nil || !strings.Contains(err.Error(), "CLI_KNOWLEDGE_RESULT_MISSING_RETRIEVAL") {
		t.Fatalf("file inventory was treated as retrieved evidence: %v", err)
	}
	if err := knowledgeRequestFromResult(&sdk.KnowledgeCiteRequest{}, []string{path}, -1, []string{"0"}, false, ""); err == nil || !strings.Contains(err.Error(), "CLI_KNOWLEDGE_RESULT_MISSING_RETRIEVAL") {
		t.Fatalf("file inventory was accepted as an answer citation: %v", err)
	}
}

func TestKnowledgeInventoryReaderAndMissingLocalResultRecovery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/newmoi/semantic-models/runtime/find-files":
			_, _ = w.Write([]byte(`{"code":"OK","data":{"output":{"files":[{"file_id":"guide-id","file_name":"Guide.md","semantic_model_id":7}]}}}`))
		case "/newmoi/semantic-models/runtime/search":
			var request sdk.KnowledgeSearchRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil || !slices.Equal(request.FileIds, []string{"guide-id"}) || !slices.Equal(request.Keywords, []string{"retention"}) {
				t.Errorf("corrected lookup lost its source or question: %v %v", &request, err)
			}
			_, _ = w.Write([]byte(`{"code":"OK","data":{"output":{"model_view":"retention fact"}}}`))
		default:
			t.Errorf("unexpected request: %s", r.URL.Path)
		}
	}))
	t.Cleanup(server.Close)
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("key"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	ctx := NewContext(client, "ws1", false, true, true)
	for _, mistakenSource := range []string{"Guide.md", "guide-id"} {
		err := ExecuteSemanticModelCommand(ctx, []string{"search", "--from", filepath.Join(t.TempDir(), mistakenSource), "--keywords", "retention"})
		var failure *sdk.Error
		if !errors.As(err, &failure) || failure.Code != "CLI_KNOWLEDGE_LOCAL_RESULT_NOT_FOUND" || failure.Details["expected"] != "saved_local_result_path" {
			t.Fatalf("missing local receipt has no actionable classification: %v", err)
		}
	}
	path := filepath.Join(t.TempDir(), "files.json")
	printed := captureStdout(t, func() {
		if err := ExecuteSemanticModelCommand(ctx, []string{"find-files", "--query", "Guide", "--output", path}); err != nil {
			t.Fatal(err)
		}
	})
	var summary struct {
		ReadCommands map[string][]string `json:"read_commands"`
	}
	if err := json.Unmarshal([]byte(printed), &summary); err != nil {
		t.Fatal(err)
	}
	args := summary.ReadCommands["files"]
	if len(args) == 0 {
		t.Fatal("saved inventory has no inspection command")
	}
	data, err := exec.CommandContext(t.Context(), args[0], args[1:]...).CombinedOutput()
	var file struct {
		Index int    `json:"index"`
		Name  string `json:"file_name"`
	}
	if err != nil || json.Unmarshal(data, &file) != nil || file.Index != 0 || file.Name != "Guide.md" {
		t.Fatalf("inventory reader lost its selector or filename: %s %v", data, err)
	}
	captureStdout(t, func() {
		if err := ExecuteSemanticModelCommand(ctx, []string{"search", "--from", path, "--index", "0", "--keywords", "retention"}); err != nil {
			t.Fatal(err)
		}
	})
}

func TestKnowledgeRuntimeRejectsInvalidInputBeforeRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("invalid input reached Product API")
	}))
	defer server.Close()
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("key"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	for _, args := range [][]string{
		{"--json", `{"unknown":true}`},
		{"--json", `{"keywords":["x"]} {}`},
		{"--max-hits", "4294967297"},
		{"--max-hits", "-1"},
		{"--query", "must not be silently ignored"},
	} {
		if err := runKnowledgeRuntime(NewContext(client, "ws1", false, true, true), "search", args); err == nil {
			t.Fatalf("accepted invalid arguments %q", args)
		}
	}
}

func TestKnowledgeUnexpectedQueryArgumentsExplainHowToPreserveEveryQuery(t *testing.T) {
	for _, operation := range []string{"search", "describe-tables"} {
		t.Run(operation, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				t.Error("unexpected positional queries must not reach the owner")
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
			err = ExecuteSemanticModelCommand(NewContext(client, "ws1", false, true, true), []string{operation, "--keywords", "first", "second"})
			if err == nil || !strings.Contains(err.Error(), "CLI_KNOWLEDGE_INPUT_INVALID") || !strings.Contains(err.Error(), "--keywords") {
				t.Fatalf("rejected multiple queries without explaining their applicable flag: %v", err)
			}
		})
	}
}

func TestKnowledgeRuntimeLargeResultsAndRejectedCalls(t *testing.T) {
	t.Chdir(t.TempDir())
	for _, tc := range []struct {
		name   string
		failed bool
	}{
		{name: "complete"}, {name: "rejected", failed: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := map[string]any{"code": "OK", "data": map[string]any{
				"call_id": tc.name, "failed": tc.failed,
				"output": map[string]any{"content": strings.Repeat("日本語😀", 10000)},
			}}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode(body)
			}))
			defer server.Close()
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("key"))
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()
			// A rejected call is kept small: it must never publish a success file.
			if tc.failed {
				body["data"].(map[string]any)["output"] = map[string]any{"code": "invalid_evidence"}
			}
			err = runKnowledgeRuntime(NewContext(client, "ws1", false, true, true), "search", nil)
			if (err != nil) != tc.failed {
				t.Fatalf("error=%v failed=%v", err, tc.failed)
			}
			raw, readErr := os.ReadFile(filepath.Join("kb-results", tc.name+".json"))
			if tc.failed {
				if !os.IsNotExist(readErr) {
					t.Fatalf("rejected call published a result: %v", readErr)
				}
				return
			}
			result := &sdk.KnowledgeRuntimeResult{}
			if readErr != nil || protojson.Unmarshal(raw, result) != nil || result.GetOutput().AsMap()["content"] != strings.Repeat("日本語😀", 10000) {
				t.Fatalf("large result was not preserved: %v", readErr)
			}
		})
	}
}

func TestKnowledgeCitationReceiptDrivesExactReadRepair(t *testing.T) {
	t.Chdir(t.TempDir())
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/newmoi/semantic-models/runtime/cite" {
			_, _ = w.Write([]byte(`{"code":"OK","data":{"call_id":"repair-receipt","output":{"accepted":false,"sources":[{"next_call":{"tool_id":"read_artifact","arguments":{"artifact_id":"exact-artifact","chunk_id":"exact-chunk","file_id":"exact-file","semantic_model_id":5000000000,"cursor":{"item_index":2,"content_offset":37}}}}]}}}`))
			return
		}
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		if r.URL.Path != "/newmoi/semantic-models/runtime/read-evidence" || request["artifact_id"] != "exact-artifact" || request["chunk_id"] != "exact-chunk" || request["file_id"] != "exact-file" || request["semantic_model_id"] != float64(5000000000) {
			t.Errorf("read lost provenance: %v", request)
		}
		cursor := request["cursor"].(map[string]any)
		if cursor["item_index"] != float64(2) || cursor["content_offset"] != float64(37) {
			t.Errorf("read restarted the evidence window: %v", cursor)
		}
		_, _ = w.Write([]byte(`{"code":"OK","data":{"call_id":"repair-read","output":{"content":"Complete evidence"}}}`))
	}))
	defer server.Close()
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("key"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	ctx := NewContext(client, "ws1", false, true, true)
	if err := runKnowledgeRuntime(ctx, "cite", []string{"--json", `{"sources":[]}`}); err == nil {
		t.Fatal("incomplete citation was accepted")
	}
	if err := runKnowledgeRuntime(ctx, "read-evidence", []string{"--from", "kb-results/repair-receipt.json", "--index", "0"}); err != nil {
		t.Fatal(err)
	}
}

// Exercise the CLI transport and local receipt together: the displayed passage
// selector must remain usable for citation without copying opaque source IDs.
func TestKnowledgeSavedSearchProvidesReadableEvidenceAndExactSelectors(t *testing.T) {
	t.Chdir(t.TempDir())
	const content = "An account grants access. Catalog contains databases, tables and volumes. 日本語😀"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/cite") {
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
				return
			}
			source := body["sources"].([]any)[0].(map[string]any)
			if source["artifact_id"] != "opaque_artifact_9" || source["chunk_id"] != "chunk_73" || source["file_id"] != "guide" || source["type"] != "rag_chunk" {
				t.Errorf("display selector lost provenance: %v", source)
			}
			_, _ = w.Write([]byte(`{"code":"OK","data":{"output":{"accepted":true}}}`))
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "OK", "data": map[string]any{
			"call_id": "lookup", "output": map[string]any{"model_view": "INDEX 0 | FILE Guide\nCatalog contains databases", "chunk_count": 1},
			"artifacts": []any{map[string]any{"artifact_id": "opaque_artifact_9", "data": map[string]any{"chunks": []any{map[string]any{
				"semantic_model_id": 7, "chunk_id": "chunk_73", "file_id": "guide", "file_name": "Guide", "content": content,
			}}}}},
		}})
	}))
	defer server.Close()
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("key"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	ctx := NewContext(client, "ws1", false, true, true)
	printed := captureStdout(t, func() {
		if err := runKnowledgeRuntime(ctx, "search", []string{"--keywords", "Catalog", "--output", "lookup.json"}); err != nil {
			t.Fatal(err)
		}
	})
	var summary map[string]any
	if err := json.Unmarshal([]byte(printed), &summary); err != nil {
		t.Fatal(err)
	}
	path, ok := summary["reading_path"].(string)
	if !ok {
		t.Fatalf("saved evidence has no readable view: %s", printed)
	}
	view, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(view), content) || !strings.Contains(string(view), `"index":0`) || !strings.Contains(string(view), "Guide") {
		t.Fatalf("readable evidence lost content/source/selectors: %s", view)
	}
	// Execute the advertised inspection command against the saved receipt,
	// rather than assuming console-summary fields also exist in that file.
	var commands struct {
		ReadCommands map[string][]string `json:"read_commands"`
	}
	if err := json.Unmarshal([]byte(printed), &commands); err != nil {
		t.Fatal(err)
	}
	args := commands.ReadCommands["preview"]
	if len(args) == 0 {
		t.Fatal("saved search has no executable preview reader")
	}
	preview, err := exec.CommandContext(t.Context(), args[0], args[1:]...).CombinedOutput()
	if err != nil {
		t.Fatalf("inspect saved result: %v: %s", err, preview)
	}
	if !strings.Contains(string(preview), "INDEX 0 | FILE Guide") || !strings.Contains(string(preview), "Catalog contains databases") {
		t.Fatalf("saved preview lost the fact or its selector: %s", preview)
	}
	captureStdout(t, func() {
		if err := runKnowledgeRuntime(ctx, "cite", []string{"--from", summary["path"].(string), "--indices", "0"}); err != nil {
			t.Fatal(err)
		}
	})
}
