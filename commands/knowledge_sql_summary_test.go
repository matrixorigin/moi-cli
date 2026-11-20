package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestKnowledgeSQLSavedResultsProduceExactCitations(t *testing.T) {
	for _, test := range []struct {
		name     string
		selected bool
		rows     any
		rowCount int
	}{
		{"whole result", false, []any{[]any{42}}, 1},
		{"explicit result index", true, []any{[]any{42}}, 1},
		{"empty rows", false, []any{}, 0},
		{"null rows", false, nil, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "sql result.json")
			raw, err := json.Marshal(map[string]any{"output": map[string]any{"artifact_id": "sql_result_current", "sql": "SELECT amount FROM sales.orders", "columns": []any{"amount"}, "rows": test.rows, "row_count": test.rowCount, "truncated": false}})
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, raw, 0600); err != nil {
				t.Fatal(err)
			}
			var sources []map[string]any
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/newmoi/semantic-models/runtime/cite" {
					t.Errorf("wrong operation: %s", r.URL.Path)
				}
				var request struct {
					Sources []map[string]any `json:"sources"`
				}
				if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
					t.Error(err)
					return
				}
				sources = request.Sources
				_, _ = w.Write([]byte(`{"code":"OK","data":{"output":{"accepted":true}}}`))
			}))
			t.Cleanup(server.Close)
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("test"))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = client.Close() })
			args := []string{"cite", "--from", path}
			if test.selected {
				args = append(args, "--indices", "0")
			}
			captureStdout(t, func() {
				if err := ExecuteSemanticModelCommand(NewContext(client, "ws", false, true, true), args); err != nil {
					t.Fatal(err)
				}
			})
			want := []map[string]any{{"type": "sql_result", "artifact_id": "sql_result_current"}}
			if !reflect.DeepEqual(sources, want) {
				t.Fatalf("SQL citation lost its result identity: %+v", sources)
			}
		})
	}
}

func TestKnowledgeSQLSummaryPreservesValuesAndExecutableReaders(t *testing.T) {
	for _, tc := range []struct {
		name                       string
		rows                       []any
		truncated, previewComplete bool
	}{
		{"aggregate", []any{[]any{"2026-11-02", "0.000000000000", nil}}, false, true},
		{"empty", []any{}, false, true},
		{"large cell", []any{[]any{"2026-11-02", strings.Repeat("完整内容", 2000), nil}}, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fullSQL := "SELECT '" + strings.Repeat("long query ", 1000) + "' AS value"
			columns := []any{"day", "amount", "nullable"}
			payload := map[string]any{"artifact_id": "sql_result_current", "sql": fullSQL, "columns": columns, "rows": tc.rows, "row_count": len(tc.rows), "truncated": tc.truncated, "table_names": []any{"sales.orders"}}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/newmoi/semantic-models/runtime/query" {
					t.Errorf("wrong route: %s", r.URL.Path)
				}
				if err := json.NewEncoder(w).Encode(map[string]any{"code": "OK", "data": map[string]any{"call_id": "current", "output": payload}}); err != nil {
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
			path := filepath.Join(t.TempDir(), "result.json")
			printed := captureStdout(t, func() {
				if err := runKnowledgeRuntime(NewContext(client, "workspace", false, true, true), "query", []string{"--json", `{"sql":"SELECT 1"}`, "--output", path}); err != nil {
					t.Fatal(err)
				}
			})
			var summary struct {
				Available     *bool               `json:"available"`
				NextAction    string              `json:"next_action"`
				ResultPreview map[string]any      `json:"result_preview"`
				ReadCommands  map[string][]string `json:"read_commands"`
			}
			if err := json.Unmarshal([]byte(printed), &summary); err != nil {
				t.Fatal(err)
			}
			if summary.Available != nil || summary.NextAction != "assess_requested_facts" {
				t.Fatalf("successful query changed availability or follow-up action: %s", printed)
			}
			preview := summary.ResultPreview
			if preview["artifact_id"] != "sql_result_current" || preview["row_count"] != float64(len(tc.rows)) || preview["truncated"] != tc.truncated || preview["preview_complete"] != tc.previewComplete {
				t.Fatalf("SQL receipt lost evidence identity, returned coverage or preview coverage: %s", printed)
			}
			if tc.previewComplete && !reflect.DeepEqual(preview["rows"], tc.rows) {
				t.Fatalf("preview changed zero/null/empty values: %#v", preview)
			}
			for _, reader := range []string{"result", "query"} {
				args := summary.ReadCommands[reader]
				if len(args) == 0 {
					t.Fatalf("missing %s reader", reader)
				}
				data, err := exec.CommandContext(t.Context(), args[0], args[1:]...).CombinedOutput()
				if err != nil {
					t.Fatalf("%s reader: %v: %s", reader, err, data)
				}
				var read map[string]any
				if err := json.Unmarshal(data, &read); err != nil {
					t.Fatal(err)
				}
				if reader == "result" && (!reflect.DeepEqual(read["rows"], tc.rows) || !reflect.DeepEqual(read["columns"], columns)) {
					t.Fatalf("reader lost full rows or headers: %s", data)
				}
				if reader == "query" && read["sql"] != fullSQL {
					t.Fatal("query reader changed the executed SQL")
				}
			}
			saved, err := os.ReadFile(path)
			if err != nil || !strings.Contains(string(saved), fullSQL) {
				t.Fatalf("saved SQL missing: %v", err)
			}
		})
	}
}

func TestKnowledgeSQLUnavailablePreservesOwnerOutcome(t *testing.T) {
	// The owner has classified the source as unavailable. The CLI must preserve
	// that outcome both inline and in the saved-result view used by agents.
	for _, saved := range []bool{false, true} {
		t.Run(map[bool]string{false: "inline", true: "saved"}[saved], func(t *testing.T) {
			payload := map[string]any{"available": false, "next_action": "use_available_sources"}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/newmoi/semantic-models/runtime/query" {
					t.Errorf("unexpected query request: %s %s", r.Method, r.URL.Path)
				}
				if err := json.NewEncoder(w).Encode(map[string]any{"code": "OK", "data": map[string]any{"call_id": "unavailable", "output": payload}}); err != nil {
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
			path := filepath.Join(t.TempDir(), "unavailable.json")
			args := []string{"query", "--json", `{"sql":"SELECT secret_value FROM hidden"}`}
			if saved {
				args = append(args, "--output", path)
			}
			printed := captureStdout(t, func() {
				if err := ExecuteSemanticModelCommand(NewContext(client, "workspace", false, true, true), args); err != nil {
					t.Fatal(err)
				}
			})
			data := []byte(printed)
			if saved {
				var summary struct {
					Available     *bool               `json:"available"`
					NextAction    string              `json:"next_action"`
					ResultPreview map[string]any      `json:"result_preview"`
					ReadCommands  map[string][]string `json:"read_commands"`
				}
				if err := json.Unmarshal(data, &summary); err != nil {
					t.Fatal(err)
				}
				if summary.Available == nil || *summary.Available || summary.NextAction != payload["next_action"] {
					t.Fatalf("summary lost the owner's unavailable outcome: %s", printed)
				}
				if summary.ResultPreview != nil || summary.ReadCommands["query"] != nil {
					t.Fatalf("unavailable source must not be presented as SQL evidence: %s", printed)
				}
				reader := summary.ReadCommands["result"]
				if len(reader) == 0 {
					t.Fatal("missing unavailable result reader")
				}
				read, err := exec.CommandContext(t.Context(), reader[0], reader[1:]...).CombinedOutput()
				if err != nil {
					t.Fatalf("unavailable result reader: %v: %s", err, read)
				}
				var outcome map[string]any
				if err := json.Unmarshal(read, &outcome); err != nil || !reflect.DeepEqual(outcome, payload) {
					t.Fatalf("reader lost the owner's outcome: %s (%v)", read, err)
				}
				data, err = os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
			}
			var result struct {
				Failed bool           `json:"failed"`
				Output map[string]any `json:"output"`
			}
			if err := json.Unmarshal(data, &result); err != nil || result.Failed || !reflect.DeepEqual(result.Output, payload) {
				t.Fatalf("CLI changed the owner's unavailable result: %s (%v)", data, err)
			}
		})
	}
}
