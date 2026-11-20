package main

import (
	"encoding/json"
	"errors"
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

// Existing runtime coverage exercises workspace-scoped SQL and file requests.
// Account operations must also pass through the real CLI entry and SDK client.
func TestSandboxAccountCommandsUseProductIdentity(t *testing.T) {
	for _, tc := range []struct {
		name, method, path, body, response string
		args                               []string
		status                             int
		withoutWorkspace                   bool
	}{
		{name: "workspace list", method: http.MethodGet, path: "/newmoi/workspaces", args: []string{"workspace", "list"}, response: `{"workspaces":[{"id":"ws-existing","name":"Existing"}],"total":1}`},
		{name: "workspace list without workspace", method: http.MethodGet, path: "/newmoi/workspaces", args: []string{"workspace", "list"}, response: `{"workspaces":[{"id":"ws-existing","name":"Existing"}],"total":1}`, withoutWorkspace: true},
		{name: "workspace create", method: http.MethodPost, path: "/newmoi/workspaces", args: []string{"workspace", "create", "--name", "新工作区"}, body: `{"name":"新工作区"}`, response: `{"id":"ws-created"}`},
		{name: "start creation", method: http.MethodPost, path: "/newmoi/workspaces/creations", args: []string{"workspace", "start-creation", "--name", "新工作区"}, body: `{"name":"新工作区"}`, response: `{"id":"creation-job","name":"新工作区","status":"pending"}`},
		{name: "creation progress", method: http.MethodGet, path: "/newmoi/workspaces/creations", args: []string{"workspace", "creations"}, response: `{"creations":[{"id":"creation-job","status":"succeeded","workspace_id":"ws-created"}]}`},
		{name: "account preferences", method: http.MethodGet, path: "/newmoi/auth/preferences", args: []string{"account", "preferences"}, response: `{"locale":"zh-CN","timezone":"Asia/Shanghai"}`},
		{name: "creation denied", method: http.MethodPost, path: "/newmoi/workspaces", args: []string{"workspace", "create", "--name", "新工作区"}, body: `{"name":"新工作区"}`, status: http.StatusForbidden, response: `{"code":"PERMISSION_DENIED","msg":"工作区创建被拒绝"}`},
		{name: "account dependency unavailable", method: http.MethodGet, path: "/newmoi/auth/preferences", args: []string{"account", "preferences"}, status: http.StatusServiceUnavailable, response: `{"code":"ErrServer","msg":"账号服务暂不可用"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			received := make(chan struct{}, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != tc.method || r.URL.Path != tc.path {
					t.Errorf("request = %s %s, want %s %s", r.Method, r.URL.Path, tc.method, tc.path)
				}
				if r.Header.Get("X-API-Key") != "synthetic-child-pat" || r.Header.Get("X-MOI-Runtime-Authorization") != "" ||
					r.Header.Get("X-Workspace-ID") != "" || r.Header.Get("X-MOI-Role-ID") != "" || r.Header.Get("Authorization") != "" {
					t.Error("account request did not use only its Product identity")
				}
				if tc.body != "" {
					var got, want map[string]any
					if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
						t.Error(err)
					}
					if err := json.Unmarshal([]byte(tc.body), &want); err != nil {
						t.Error(err)
					}
					if !reflect.DeepEqual(got, want) {
						t.Errorf("request body = %v, want %v", got, want)
					}
				}
				received <- struct{}{}
				w.Header().Set("Content-Type", "application/json")
				if tc.status != 0 {
					w.WriteHeader(tc.status)
					_, _ = w.Write([]byte(tc.response))
				} else {
					_, _ = w.Write([]byte(`{"code":"OK","data":` + tc.response + `}`))
				}
			}))
			defer server.Close()
			t.Setenv("MOI_ENDPOINT", server.URL+"/newmoi")
			t.Setenv("MOI_API_KEY", "synthetic-child-pat")
			t.Setenv("MOI_WORKSPACE_ID", "ws-runtime")
			if tc.withoutWorkspace {
				t.Setenv("MOI_WORKSPACE_ID", "")
			}
			t.Setenv("MOI_RUNTIME_AUTHORIZATION", "Bearer synthetic-child-proof")
			t.Setenv("MOI_CLI_CONFIG_FILE", filepath.Join(t.TempDir(), "absent-config.json"))
			binary, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			output, err := exec.CommandContext(t.Context(), binary, append([]string{"test-cli-entry", "-o"}, tc.args...)...).CombinedOutput()
			if tc.status == 0 && err != nil {
				t.Fatalf("account command failed: %v: %s", err, output)
			}
			if tc.status != 0 && err == nil {
				t.Fatalf("failed account operation reported success: %s", output)
			}
			var got, want map[string]any
			if err := json.Unmarshal(output, &got); err != nil {
				t.Fatalf("invalid command output: %s: %v", output, err)
			}
			if err := json.Unmarshal([]byte(tc.response), &want); err != nil {
				t.Fatal(err)
			}
			if tc.status != 0 {
				want = map[string]any{"code": want["code"], "message": want["msg"], "status": float64(tc.status)}
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("command result = %v, want %v", got, want)
			}
			select {
			case <-received:
			default:
				t.Fatal("account command never reached the Product API")
			}
		})
	}
}

func TestRuntimeWorkspaceAuthorizationBelongsToProductAPI(t *testing.T) {
	for _, tc := range []struct{ name, workspace, role string }{
		{name: "other workspace", workspace: "ws-other"},
		{name: "explicit Role", workspace: "ws-runtime", role: "17"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			received := make(chan struct{}, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/newmoi/catalog/healthz" || r.Header.Get("X-Workspace-ID") != tc.workspace ||
					r.Header.Get("X-MOI-Role-ID") != tc.role || r.Header.Get("X-MOI-Runtime-Authorization") != "Bearer synthetic-child-proof" {
					t.Error("request scope or task proof changed before reaching MOI")
				}
				received <- struct{}{}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"code":"PERMISSION_DENIED","msg":"当前操作无权限"}`))
			}))
			defer server.Close()
			opts := []sdk.Option{sdk.WithAPIKey("synthetic-child-pat"), sdk.WithRuntimeAuthorization("Bearer synthetic-child-proof"), sdk.WithHTTPClient(runtimeHTTPClient())}
			if tc.role != "" {
				opts = append(opts, sdk.WithWorkspaceCurrentRoleID(tc.workspace, tc.role))
			}
			client, err := sdk.New(server.URL+"/newmoi", opts...)
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()
			_, err = client.Catalog(tc.workspace).Health(t.Context())
			var failure *sdk.Error
			if !errors.As(err, &failure) || failure.Code != "PERMISSION_DENIED" || failure.Message != "当前操作无权限" || failure.Status != http.StatusForbidden {
				t.Fatalf("MOI denial was not preserved: %v", err)
			}
			select {
			case <-received:
			default:
				t.Fatal("CLI blocked the request before MOI could authorize it")
			}
		})
	}
}

func TestSandboxCommandUsesOnlyItsRuntimeIdentity(t *testing.T) {
	requests := make(chan http.Header, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- r.Header.Clone()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":"OK","data":{"db_list":[]}}`))
	}))
	t.Cleanup(server.Close)
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	// Runtime identity must not consume the human's saved connection/Role config.
	config := filepath.Join(t.TempDir(), "invalid-config.json")
	if err := os.WriteFile(config, []byte(`not-json`), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MOI_CLI_CONFIG_FILE", config)
	t.Setenv("MOI_ASTRA_BINARY", filepath.Join(t.TempDir(), "missing-astra"))
	t.Setenv("MOI_ENDPOINT", server.URL+"/newmoi")
	t.Setenv("MOI_API_KEY", "synthetic-child-pat")
	t.Setenv("MOI_WORKSPACE_ID", "ws-runtime")
	t.Setenv("MOI_RUNTIME_AUTHORIZATION", "Bearer synthetic-child-proof")
	cmd := exec.CommandContext(t.Context(), binary, "test-cli-entry", "-o", "sql", "meta-db-tree")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%v: %s", err, output)
	}
	select {
	case header := <-requests:
		if header.Get("X-Workspace-ID") != "ws-runtime" || header.Get("X-API-Key") != "synthetic-child-pat" ||
			header.Get("X-MOI-Runtime-Authorization") != "Bearer synthetic-child-proof" || header.Get("X-MOI-Role-ID") != "" || header.Get("Authorization") != "" {
			t.Fatal("runtime request did not retain its exact credential/workspace binding")
		}
	default:
		t.Fatal("no runtime request")
	}
	if strings.Contains(string(output), "synthetic-child") {
		t.Fatal("runtime credential leaked to command output")
	}
}

// This follows the weather trace's CLI entry, including redundant connection
// flags and an explicit subcommand workspace. The server owns scope rejection.
func TestSandboxPublishedServiceAndResultRequestsReachMOI(t *testing.T) {
	for _, tc := range []struct {
		name, method, path, workspace, request, response string
		args                                             []string
		status                                           int
		rawResponse                                      bool
	}{
		{
			name: "read execution result", method: http.MethodGet,
			path: "/newmoi/workflow/v2/workflow-apps/executions/execution-weather/result", workspace: "ws-runtime",
			args:     []string{"workflow", "get-run-result-by-id", "--execution-id", "execution-weather", "--workspace-id", "ws-runtime"},
			response: `{"result":{"execution_id":"execution-weather","status":"completed","case_result":"{\"city\":\"南京\",\"temperature\":16.9}"}}`,
		},
		{
			name: "invoke published MCP tool", method: http.MethodPost,
			path: "/newmoi/workspaces/ws-runtime/published-mcp", workspace: "ws-runtime",
			request:  `{"jsonrpc":"2.0","id":"weather-call","method":"tools/call","params":{"name":"weather_forecast","arguments":{"city":"南京"}}}`,
			args:     []string{"agent", "resource-post", "--resource", "published-mcp", "--json", `{"jsonrpc":"2.0","id":"weather-call","method":"tools/call","params":{"name":"weather_forecast","arguments":{"city":"南京"}}}`},
			response: `{"jsonrpc":"2.0","id":"weather-call","result":{"content":[{"type":"text","text":"{\"city\":\"南京\",\"temperature\":16.9}"}]}}`, rawResponse: true,
		},
		{
			name: "other workspace denied by MOI", method: http.MethodGet,
			path: "/newmoi/workflow/v2/workflow-apps/executions/execution-weather/result", workspace: "ws-other",
			args:   []string{"workflow", "get-run-result-by-id", "--execution-id", "execution-weather", "--workspace-id", "ws-other"},
			status: http.StatusForbidden, response: `{"code":"PERMISSION_DENIED","msg":"当前操作无权限"}`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			received := make(chan struct{}, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != tc.method || r.URL.Path != tc.path || r.Header.Get("X-Workspace-ID") != tc.workspace ||
					r.Header.Get("X-API-Key") != "synthetic-child-pat" || r.Header.Get("X-MOI-Runtime-Authorization") != "Bearer synthetic-child-proof" {
					t.Error("CLI changed the explicit request scope or runtime credentials")
				}
				if tc.request != "" {
					var actual, expected any
					if err := json.NewDecoder(r.Body).Decode(&actual); err != nil {
						t.Error(err)
					}
					if err := json.Unmarshal([]byte(tc.request), &expected); err != nil {
						t.Error(err)
						return
					}
					if !reflect.DeepEqual(actual, expected) {
						t.Error("CLI changed the published tool invocation")
					}
				}
				received <- struct{}{}
				w.Header().Set("Content-Type", "application/json")
				if tc.status != 0 {
					w.WriteHeader(tc.status)
					_, _ = w.Write([]byte(tc.response))
				} else if tc.rawResponse {
					_, _ = w.Write([]byte(tc.response))
				} else {
					_, _ = w.Write([]byte(`{"code":"OK","data":` + tc.response + `}`))
				}
			}))
			defer server.Close()
			t.Setenv("MOI_ENDPOINT", server.URL+"/newmoi")
			t.Setenv("MOI_API_KEY", "synthetic-child-pat")
			t.Setenv("MOI_WORKSPACE_ID", "ws-runtime")
			t.Setenv("MOI_RUNTIME_AUTHORIZATION", "Bearer synthetic-child-proof")
			t.Setenv("MOI_CLI_CONFIG_FILE", filepath.Join(t.TempDir(), "absent.json"))
			binary, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			args := []string{"test-cli-entry", "--endpoint", server.URL + "/newmoi", "--api-key", "synthetic-child-pat", "--workspace-id", "ws-runtime", "--auth", "pat", "-o"}
			output, commandErr := exec.CommandContext(t.Context(), binary, append(args, tc.args...)...).CombinedOutput()
			if (commandErr != nil) != (tc.status != 0) {
				t.Fatalf("CLI exit = %v, output: %s", commandErr, output)
			}
			select {
			case <-received:
			default:
				t.Fatalf("CLI did not reach MOI: %s", output)
			}
			var actual, expected map[string]any
			if err := json.Unmarshal(output, &actual); err != nil {
				t.Fatalf("invalid CLI output: %v: %s", err, output)
			}
			if err := json.Unmarshal([]byte(tc.response), &expected); err != nil {
				t.Fatal(err)
			}
			if tc.status != 0 {
				expected = map[string]any{"code": expected["code"], "message": expected["msg"], "status": float64(tc.status)}
			}
			if !reflect.DeepEqual(actual, expected) {
				t.Fatalf("CLI did not preserve MOI output: %s", output)
			}
		})
	}
}

func TestSandboxCommandRejectsIncompleteCredentials(t *testing.T) {
	requests := make(chan struct{}, 16)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests <- struct{}{}
		_, _ = w.Write([]byte(`{"code":"OK","data":{}}`))
	}))
	t.Cleanup(server.Close)
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name  string
		empty string
	}{
		{name: "missing proof", empty: "MOI_RUNTIME_AUTHORIZATION"},
		{name: "missing PAT", empty: "MOI_API_KEY"},
		{name: "missing endpoint", empty: "MOI_ENDPOINT"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("MOI_ENDPOINT", server.URL+"/newmoi")
			t.Setenv("MOI_API_KEY", "synthetic-child-pat")
			t.Setenv("MOI_WORKSPACE_ID", "ws-runtime")
			t.Setenv("MOI_RUNTIME_AUTHORIZATION", "Bearer synthetic-child-proof")
			if tc.empty != "" {
				t.Setenv(tc.empty, "")
			}
			output, err := exec.CommandContext(t.Context(), binary, "test-cli-entry", "-o", "sql", "meta-db-tree").CombinedOutput()
			if err == nil {
				t.Fatalf("invalid runtime identity accepted: %s", output)
			}
			if strings.Contains(string(output), "synthetic-child") {
				t.Fatal("runtime credential leaked in failure output")
			}
			select {
			case <-requests:
				t.Fatal("invalid runtime identity sent a request")
			default:
			}
		})
	}
}
