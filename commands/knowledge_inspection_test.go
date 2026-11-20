package commands

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestKnowledgeEvidenceReaderKeepsSelectorsWithReorderedDocumentPositions(t *testing.T) {
	const fact = "A shipment is complete only after all active parcels arrive."
	var selected []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/cite") {
			var request struct {
				Sources []map[string]any `json:"sources"`
			}
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Error(err)
			}
			for _, source := range request.Sources {
				selected = append(selected, source["chunk_id"].(string))
				if source["artifact_id"] != "later-page" {
					t.Errorf("lost page provenance: %v", source)
				}
			}
			_, _ = io.WriteString(w, `{"code":"OK","data":{"output":{"accepted":true}}}`)
			return
		}
		artifacts := []any{
			map[string]any{"artifact_id": "source-metadata", "data": map[string]any{"file_name": "Manual"}},
			map[string]any{"artifact_id": "first-page", "data": map[string]any{"chunks": []any{
				map[string]any{"file_id": "guide", "semantic_model_id": 7, "chunk_id": "heading", "file_name": "Manual", "chunk_index": 0, "chunk_index_scope": "file", "content": "# Delivery"},
				map[string]any{"file_id": "guide", "semantic_model_id": 7, "chunk_id": "other-rule", "file_name": "Manual", "chunk_index": 5, "chunk_index_scope": "file", "content": "Inspection must retain the parcel identity."},
			}}},
			map[string]any{"artifact_id": "later-page", "data": map[string]any{"chunks": []any{
				map[string]any{"file_id": "guide", "semantic_model_id": 7, "chunk_id": "completion-rule", "file_name": "Manual", "chunk_index": 1, "chunk_index_scope": "file", "content": fact},
			}}},
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "OK", "data": map[string]any{"call_id": "search", "output": map[string]any{}, "artifacts": artifacts}})
	}))
	t.Cleanup(server.Close)
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("key"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	ctx := NewContext(client, "workspace", false, true, true)
	path := filepath.Join(t.TempDir(), "evidence with 'quotes'.json")
	printed := captureStdout(t, func() {
		if err := ExecuteSemanticModelCommand(ctx, []string{"search", "--keywords", "completion", "--output", path}); err != nil {
			t.Fatal(err)
		}
	})
	var summary struct {
		ReadCommands map[string][]string `json:"read_commands"`
	}
	if err := json.Unmarshal([]byte(printed), &summary); err != nil {
		t.Fatal(err)
	}
	args := summary.ReadCommands["evidence"]
	if len(args) == 0 {
		t.Fatal("no executable evidence reader with stable selectors")
	}
	raw, err := exec.CommandContext(t.Context(), args[0], args[1:]...).CombinedOutput()
	if err != nil {
		t.Fatalf("inspect evidence: %s: %v", raw, err)
	}
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	index := -1
	for decoder.More() {
		var record struct {
			From     string `json:"from"`
			Index    int    `json:"index"`
			Content  string `json:"content"`
			Position struct {
				ChunkIndex int `json:"chunk_index"`
			} `json:"source_position"`
		}
		if err := decoder.Decode(&record); err != nil {
			t.Fatal(err)
		}
		if record.Content == fact {
			if record.From != path || record.Index != 2 || record.Position.ChunkIndex != 1 {
				t.Fatalf("document position replaced the evidence locator: %+v", record)
			}
			index = record.Index
		}
	}
	if index < 0 {
		t.Fatal("later page's requested fact was omitted")
	}
	captureStdout(t, func() {
		if err := ExecuteSemanticModelCommand(ctx, []string{"cite", "--from", path, "--indices", strconv.Itoa(index)}); err != nil {
			t.Fatal(err)
		}
	})
	if !reflect.DeepEqual(selected, []string{"completion-rule"}) {
		t.Fatalf("cited a heading/another passage instead of the inspected fact: %v", selected)
	}
}

func TestKnowledgeSchemaReaderPreservesDeclaredAndAbsentConstraints(t *testing.T) {
	tables := []any{
		map[string]any{"name": "scope.ledger", "columns": []any{map[string]any{"name": "entry_id", "type": "BIGINT"}}, "ddl": "CREATE TABLE ledger (entry_id BIGINT NOT NULL PRIMARY KEY)"},
		map[string]any{"name": "scope.rates", "columns": []any{map[string]any{"name": "rate", "type": "DECIMAL(12,6)"}}, "ddl": "CREATE TABLE rates (rate DECIMAL(12,6))"},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "OK", "data": map[string]any{"output": map[string]any{"tables": tables}}})
	}))
	t.Cleanup(server.Close)
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("key"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	path := filepath.Join(t.TempDir(), "schema.json")
	printed := captureStdout(t, func() {
		if err := ExecuteSemanticModelCommand(NewContext(client, "workspace", false, true, true), []string{"describe-tables", "--output", path}); err != nil {
			t.Fatal(err)
		}
	})
	var summary struct {
		ReadCommands map[string][]string `json:"read_commands"`
	}
	if err := json.Unmarshal([]byte(printed), &summary); err != nil {
		t.Fatal(err)
	}
	args := summary.ReadCommands["schema"]
	if len(args) == 0 {
		t.Fatal("schema has no executable column/DDL reader")
	}
	raw, err := exec.CommandContext(t.Context(), args[0], args[1:]...).CombinedOutput()
	if err != nil {
		t.Fatalf("inspect schema: %s: %v", raw, err)
	}
	var got []any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, tables) {
		t.Fatalf("schema reader lost columns or invented constraints: %s", raw)
	}
	fieldArgs := summary.ReadCommands["schema_fields"]
	if len(fieldArgs) == 0 {
		t.Fatal("no compact reader preserving original table and field identities")
	}
	fieldData, err := exec.CommandContext(t.Context(), fieldArgs[0], fieldArgs[1:]...).CombinedOutput()
	if err != nil {
		t.Fatalf("inspect fields: %s: %v", fieldData, err)
	}
	decoder := json.NewDecoder(strings.NewReader(string(fieldData)))
	var records []any
	for decoder.More() {
		var record map[string]any
		if err := decoder.Decode(&record); err != nil {
			t.Fatal(err)
		}
		records = append(records, record)
	}
	want := []any{
		map[string]any{"name": "scope.ledger", "columns": []any{map[string]any{"name": "entry_id", "type": "BIGINT"}}},
		map[string]any{"name": "scope.rates", "columns": []any{map[string]any{"name": "rate", "type": "DECIMAL(12,6)"}}},
	}
	if !reflect.DeepEqual(records, want) {
		t.Fatalf("compact reader changed field names/types: %s", fieldData)
	}
}

func TestKnowledgeQueryReaderLocatesCTEAndSetBranches(t *testing.T) {
	repair := map[string]any{"rule": "table_qualifier_alias", "before": "events", "after": "`e`", "start_byte": float64(7), "end_byte": float64(13)}
	blocks := []any{
		map[string]any{"id": "query_1", "cte": "eligible", "relations": []any{map[string]any{"name": "accounts"}}, "where": "`enabled`=1", "projections": []any{map[string]any{"expression": "`account_id`"}}},
		map[string]any{"id": "query_2", "set_group": float64(1), "branch": float64(1), "relations": []any{map[string]any{"name": "events", "alias": "e"}}, "where": "`e`.`occurred_at`>=1794153600", "projections": []any{map[string]any{"expression": "COUNT(1)", "alias": "total"}}},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "OK", "data": map[string]any{"output": map[string]any{"artifact_id": "result", "columns": []any{"total"}, "rows": []any{[]any{"7"}}, "row_count": 1, "truncated": false, "query_blocks": blocks, "sql_repairs": []any{repair}}}})
	}))
	t.Cleanup(server.Close)
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("key"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })
	path := filepath.Join(t.TempDir(), "query with 'quotes'.json")
	printed := captureStdout(t, func() {
		if err := ExecuteSemanticModelCommand(NewContext(client, "workspace", false, true, true), []string{"query", "--json", `{"sql":"SELECT 7 AS total"}`, "--output", path}); err != nil {
			t.Fatal(err)
		}
	})
	var summary struct {
		ReadCommands map[string][]string `json:"read_commands"`
	}
	if err := json.Unmarshal([]byte(printed), &summary); err != nil {
		t.Fatal(err)
	}
	args := summary.ReadCommands["query_blocks"]
	if len(args) == 0 {
		t.Fatal("no query-block reader")
	}
	raw, err := exec.CommandContext(t.Context(), args[0], args[1:]...).CombinedOutput()
	if err != nil {
		t.Fatalf("read query blocks: %s: %v", raw, err)
	}
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	var got []any
	for decoder.More() {
		var block map[string]any
		if err := decoder.Decode(&block); err != nil {
			t.Fatal(err)
		}
		got = append(got, block)
	}
	if !reflect.DeepEqual(got, blocks) {
		t.Fatalf("query reader lost expression, scope or branch: %s", raw)
	}
	args = summary.ReadCommands["sql_repairs"]
	if len(args) == 0 {
		t.Fatal("no executable repair reader")
	}
	raw, err = exec.CommandContext(t.Context(), args[0], args[1:]...).CombinedOutput()
	if err != nil {
		t.Fatalf("read SQL repairs: %s: %v", raw, err)
	}
	var actualRepair map[string]any
	if err := json.Unmarshal(raw, &actualRepair); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(repair, actualRepair) {
		t.Fatalf("repair provenance lost: %s", raw)
	}
}
