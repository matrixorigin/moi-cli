package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestSQLRunExitStatusPreservesQueryResult(t *testing.T) {
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("MOI_CLI_CONFIG_FILE", filepath.Join(t.TempDir(), "config.json"))
	for _, status := range []string{"success", "failed", "killed"} {
		t.Run(status, func(t *testing.T) {
			result := sdk.SQLQueryResult{QueryId: "query-1", StatementId: "statement-1", Status: status}
			if status != "success" {
				result.ErrMsg = "server-provided SQL diagnostic"
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/query/run" {
					t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(map[string]any{"code": "OK", "data": &result})
			}))
			t.Cleanup(server.Close)
			command := exec.CommandContext(t.Context(), binary, "test-cli-entry", "-auth", "pat", "-endpoint", server.URL,
				"-api-key", "test-only", "-workspace-id", "workspace-1", "-o", "sql", "run", "--query", "SELECT missing_column FROM orders")
			var stdout, stderr bytes.Buffer
			command.Stdout, command.Stderr = &stdout, &stderr
			err := command.Run()
			if status == "success" {
				if err != nil {
					t.Fatalf("successful query: %v; %s", err, &stderr)
				}
			} else {
				var exitError *exec.ExitError
				if !errors.As(err, &exitError) || exitError.ExitCode() == 0 {
					t.Fatalf("%s query exited successfully: err=%v stdout=%s stderr=%s", status, err, &stdout, &stderr)
				}
			}
			var output sdk.SQLQueryResult
			if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
				t.Fatalf("query result JSON: %v; %s", err, &stdout)
			}
			if output.GetStatus() != status || output.GetQueryId() != result.GetQueryId() || output.GetStatementId() != result.GetStatementId() || output.GetErrMsg() != result.GetErrMsg() {
				t.Fatalf("query result was lost: %s", &stdout)
			}
		})
	}
}
