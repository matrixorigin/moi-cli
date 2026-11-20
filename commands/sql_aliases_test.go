package commands

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestSQLAliasesCarryQueryAndStatementIDs(t *testing.T) {
	for _, equals := range []bool{false, true} {
		t.Run(map[bool]string{false: "separate", true: "equals"}[equals], func(t *testing.T) {
			requests := make(chan string, 8)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var body map[string]any
				if r.Method == http.MethodPost {
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Errorf("request JSON: %v", err)
					}
				}
				var data any
				switch r.URL.Path {
				case "/newmoi/query/execute":
					if body["query"] != "SELECT 7 AS n" {
						t.Errorf("query input: %#v", body)
					}
					data = map[string]any{"id": "query-1", "query_id": "query-1"}
				case "/newmoi/query/describe", "/newmoi/query/kill":
					if body["query_id"] != "query-1" {
						t.Errorf("query ID not preserved: %#v", body)
					}
					data = map[string]any{"query_id": "query-1", "statement_id": "statement-2", "status": "success"}
				case "/newmoi/query/result":
					if body["statement_id"] != "statement-2" {
						t.Errorf("statement ID not preserved: %#v", body)
					}
					data = map[string]any{"status": "success", "total": 1, "columns": []any{map[string]any{"name": "n", "type": "INT"}}, "result": []any{[]any{map[string]any{"String": "7", "Valid": true}}}}
				case "/newmoi/meta/db/table":
					if body["name"] != "db1" {
						t.Errorf("database alias not preserved: %#v", body)
					}
					data = map[string]any{"db_name": "db1"}
				default:
					t.Errorf("unexpected request: %s", r.URL.Path)
				}
				requests <- r.URL.Path
				_ = json.NewEncoder(w).Encode(map[string]any{"code": "OK", "data": data})
			}))
			t.Cleanup(server.Close)
			client, err := sdk.New(server.URL+"/newmoi", sdk.WithAPIKey("test-only"))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = client.Close() })
			ctx := NewContext(client, "ws-1", false, true, true)
			run := func(sub, flag, value string, out any) {
				t.Helper()
				args := []string{sub, "--" + flag, value}
				if equals {
					args = []string{sub, "--" + flag + "=" + value}
				}
				output := captureStdout(t, func() {
					if err := ExecuteSQLCommand(ctx, args); err != nil {
						t.Fatalf("%s: %v", sub, err)
					}
				})
				if out != nil {
					if err := json.Unmarshal([]byte(output), out); err != nil {
						t.Fatalf("%s output: %v: %s", sub, err, output)
					}
				}
			}
			var started sdk.SQLQueryStartResult
			run("execute", "query", "SELECT 7 AS n", &started)
			if started.GetQueryId() == "" || started.GetId() != started.GetQueryId() {
				t.Fatalf("start response: %v", &started)
			}
			var operation sdk.SQLQueryOperation
			run("describe", "query-id", started.GetQueryId(), &operation)
			if operation.GetStatus() != "success" || operation.GetStatementId() == started.GetQueryId() {
				t.Fatalf("describe response: %v", &operation)
			}
			var result sdk.SQLQueryResult
			run("result", "statement-id", operation.GetStatementId(), &result)
			if len(result.GetResult()) != 1 || result.GetTotal() != 1 {
				t.Fatalf("result response: %v", &result)
			}
			run("kill", "query-id", started.GetQueryId(), nil)
			run("meta-tables", "db", "db1", nil)
			for _, suffix := range []string{"query/execute", "query/describe", "query/result", "query/kill", "meta/db/table"} {
				select {
				case path := <-requests:
					if !strings.HasSuffix(path, suffix) {
						t.Fatalf("request order: got %s, want %s", path, suffix)
					}
				default:
					t.Fatalf("missing request %s", suffix)
				}
			}
		})
	}
}

func TestSQLRunReturnsResultWithoutPollingOrReplaying(t *testing.T) {
	for _, tc := range []struct {
		name, query, response  string
		wantRows, wantAffected int
		wantError              bool
	}{
		{"query", "SELECT NULL AS n", `{"code":"OK","data":{"query_id":"q1","statement_id":"s1","status":"success","columns":[{"name":"n","type":"INT"}],"result":[[{"String":"","Valid":false}]],"total":1}}`, 1, 0, false},
		{"update", "UPDATE items SET version=2", `{"code":"OK","data":{"query_id":"q1","statement_id":"s1","status":"success","rows_affected":2,"empty_set":{"msg":"Query OK"}}}`, 0, 2, false},
		{"denied", "DELETE FROM items", `{"code":"ErrForbidden","msg":"Permission denied"}`, 0, 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			requests := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests++
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Error(err)
				}
				if r.URL.Path != "/query/run" || body["query"] != tc.query || body["db_name"] != "sales" || body["limit"] != float64(30) {
					t.Errorf("unexpected request %s: %#v", r.URL.Path, body)
				}
				w.Header().Set("Content-Type", "application/json")
				if tc.wantError {
					w.WriteHeader(http.StatusForbidden)
				}
				_, _ = w.Write([]byte(tc.response))
			}))
			t.Cleanup(server.Close)
			client, err := sdk.New(server.URL, sdk.WithAPIKey("test-only"))
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = client.Close() })
			ctx := NewContext(client, "ws-1", false, true, true)
			output := captureStdout(t, func() {
				err = ExecuteSQLCommand(ctx, []string{"run", "--query", tc.query, "--db", "sales", "--limit", "30"})
			})
			if (err != nil) != tc.wantError {
				t.Fatalf("run error = %v", err)
			}
			if requests != 1 {
				t.Fatalf("run made %d requests", requests)
			}
			if tc.wantError {
				if !sdk.IsCode(err, "ErrForbidden") || output != "" {
					t.Fatalf("failure contract: %v / %s", err, output)
				}
				return
			}
			var result sdk.SQLQueryResult
			if err := json.Unmarshal([]byte(output), &result); err != nil {
				t.Fatal(err)
			}
			if len(result.GetResult()) != tc.wantRows || result.GetRowsAffected() != int64(tc.wantAffected) || result.GetStatementId() != "s1" {
				t.Fatalf("result: %s", output)
			}
		})
	}
}
