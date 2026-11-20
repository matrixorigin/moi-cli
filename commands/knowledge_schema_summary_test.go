package commands

import (
	"encoding/json"
	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestKnowledgeSchemaSummaryCommandsPreserveDistinctDefinitionsAndFullSQL(t *testing.T) {
	sql := "SELECT 'complete standard SQL' AS value"
	description := "Count settled usage only; a usage event does not establish payment."
	qa := map[string]any{"kind": "standard_qa", "key_name": "usage", "question": "Usage for region A?", "description": description, "usage_mode": "force", "sql": sql}
	other := map[string]any{"kind": "standard_qa", "key_name": "usage", "question": "Usage for region B?", "usage_mode": "reference", "sql": "SELECT 2"}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/newmoi/semantic-models/runtime/describe-tables" {
			t.Errorf("wrong route: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"code": "OK", "data": map[string]any{"call_id": "schema", "output": map[string]any{"tables": []any{
			map[string]any{"name": "sales.a", "semantic_entries": []any{qa}},
			map[string]any{"name": "sales.b", "semantic_entries": []any{qa, other}},
		}}}})
	}))
	t.Cleanup(server.Close)
	client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("key"))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	path := filepath.Join(t.TempDir(), "schema.json")
	output := captureStdout(t, func() {
		if err := runKnowledgeRuntime(NewContext(client, "workspace", false, true, true), "describe-tables", []string{"--output", path}); err != nil {
			t.Fatal(err)
		}
	})
	var summary knowledgeResultSummary
	if err = json.Unmarshal([]byte(output), &summary); err != nil {
		t.Fatal(err)
	}
	var observation map[string]any
	if err = json.Unmarshal([]byte(output), &observation); err != nil || observation["table_count"] != float64(2) {
		t.Fatalf("schema receipt did not report the returned sales.a/sales.b inventory: %v, %v", observation["table_count"], err)
	}
	tables := summary.ReadCommands["tables"]
	listedTables, err := exec.CommandContext(t.Context(), tables[0], tables[1:]...).Output()
	var names []string
	if err != nil || json.Unmarshal(listedTables, &names) != nil || !slices.Equal(names, []string{"sales.a", "sales.b"}) {
		t.Fatalf("schema inventory names changed: %s, %v", listedTables, err)
	}
	command := summary.ReadCommands["standard_questions"]
	if len(command) == 0 {
		t.Fatal("saved schema has no executable standard-question index")
	}
	listed, err := exec.CommandContext(t.Context(), command[0], command[1:]...).Output()
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(strings.NewReader(string(listed)))
	var questions []string
	var regionADescription string
	for decoder.More() {
		var row struct {
			Question    string `json:"question"`
			Description string `json:"description"`
		}
		if err = decoder.Decode(&row); err != nil {
			t.Fatal(err)
		}
		questions = append(questions, row.Question)
		if row.Question == "Usage for region A?" {
			regionADescription = row.Description
		}
	}
	slices.Sort(questions)
	if !slices.Equal(questions, []string{"Usage for region A?", "Usage for region B?"}) {
		t.Fatalf("definition index merged distinct questions or repeated identical entries: %v", questions)
	}
	if regionADescription != description {
		t.Fatalf("standard-question reader lost the query's governing description: %q", regionADescription)
	}
	if strings.Contains(string(listed), sql) {
		t.Fatal("question index dumped full SQL instead of its local locator")
	}
	saved, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(saved), sql) {
		t.Fatalf("saved schema lost full SQL: %v", err)
	}
}
