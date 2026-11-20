package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// Exercise the binary's real request and response-body deadlines, including
// both the sandbox task identity and the shared human-login transport.
func TestCLIExplicitRequestDeadline(t *testing.T) {
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"sandbox", "moi"} {
		t.Run(mode, func(t *testing.T) {
			var requests atomic.Int64
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests.Add(1)
				if r.URL.Path != "/newmoi/semantic-models/runtime/query" || r.Header.Get("X-Workspace-ID") != "ws-a" {
					t.Errorf("unexpected route: %s", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.(http.Flusher).Flush()
				select {
				case <-r.Context().Done():
					return
				case <-time.After(100 * time.Millisecond):
				}
				_, _ = w.Write([]byte(`{"code":"OK","data":{"call_id":"deadline-query","output":{"columns":["evidence"],"rows":[["DEADLINE_EVIDENCE"]]}}}`))
			}))
			t.Cleanup(server.Close)
			t.Setenv("MOI_CLI_CONFIG_FILE", filepath.Join(t.TempDir(), "config.json"))
			t.Setenv("MOI_AGENT_CALL", "1")
			t.Setenv("MOI_ASTRA_BINARY", binary)
			t.Setenv("MOI_ROLE_TEST_ENDPOINT", server.URL+"/newmoi")
			t.Setenv("MOI_ENDPOINT", server.URL+"/newmoi")
			t.Setenv("MOI_WORKSPACE_ID", "ws-a")
			t.Setenv("MOI_RUNTIME_AUTHORIZATION", "")
			if err := os.Unsetenv("MOI_RUNTIME_AUTHORIZATION"); err != nil {
				t.Fatal(err)
			}
			t.Setenv("MOI_API_KEY", "")
			if mode == "sandbox" {
				t.Setenv("MOI_API_KEY", "synthetic-child-key")
				t.Setenv("MOI_RUNTIME_AUTHORIZATION", "Bearer synthetic-child-proof")
			}
			for _, tc := range []struct {
				budget  string
				want    string
				success bool
				request bool
			}{
				{"1s", "DEADLINE_EVIDENCE", true, true},
				{"20ms", "ErrTimeout", false, true},
				{"0", "CLI_REQUEST_TIMEOUT_INVALID", false, false},
				{"-1s", "CLI_REQUEST_TIMEOUT_INVALID", false, false},
			} {
				t.Run(tc.budget, func(t *testing.T) {
					before := requests.Load()
					args := []string{"test-cli-entry", "--request-timeout", tc.budget, "-o", "semantic-model", "query", "--json", `{"sql":"SELECT 1"}`}
					output, err := exec.CommandContext(t.Context(), binary, args...).CombinedOutput()
					if (err == nil) != tc.success || !strings.Contains(string(output), tc.want) {
						t.Fatalf("request budget %s: %v: %s", tc.budget, err, output)
					}
					if tc.request && requests.Load() == before {
						t.Fatal("deadline test never reached product transport")
					}
					if !tc.request && requests.Load() != before {
						t.Fatal("invalid budget sent a product request")
					}
				})
			}
		})
	}
}
