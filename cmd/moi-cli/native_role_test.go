package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/matrixorigin/matrixflow/sdk/cli/commands"
	"github.com/matrixorigin/matrixflow/sdk/cli/internal/nativeauth"
	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

// Exercise the real helper pipe without reading or changing a user's login.
func TestMain(m *testing.M) {
	if len(os.Args) > 1 && os.Args[1] == "test-cli-entry" {
		os.Args = append(os.Args[:1], os.Args[2:]...)
		main()
		os.Exit(0)
	}
	if len(os.Args) > 1 && os.Args[1] == "auth" {
		if os.Getenv("MOI_ROLE_TEST_SIGNED_OUT") == "1" {
			os.Exit(1)
		}
		workspace := "ws-a"
		pipe := os.NewFile(3, "credential-pipe")
		if pipe == nil {
			os.Exit(1)
		}
		err := json.NewEncoder(pipe).Encode(nativeauth.Credential{
			Version: 1, Environment: "test-environment", Issuer: "https://uc.example.test/realms/moi",
			Subject: "test-user", Generation: "test-generation", Endpoint: os.Getenv("MOI_ROLE_TEST_ENDPOINT"),
			ExpiresAt: time.Now().Add(time.Hour).Unix(), AccessToken: "synthetic-access", WorkspaceID: &workspace,
		})
		pipe.Close()
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestSandboxChildUploadsWithBoundIdentityWithoutInteractiveConfig(t *testing.T) {
	requests := make(chan struct{}, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/newmoi/workspaces/ws-child/runtime-artifacts" ||
			r.Header.Get("X-API-Key") != "synthetic-child-key" || r.Header.Get("X-MOI-Runtime-Authorization") != "Bearer synthetic-child-proof" ||
			r.Header.Get("X-Workspace-ID") != "ws-child" || r.Header.Get("X-MOI-Role-ID") != "" {
			t.Error("sandbox upload lost its bound endpoint or identity")
		}
		file, header, err := r.FormFile("file")
		if err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer file.Close()
		body, err := io.ReadAll(file)
		if err != nil || header.Filename != "result.txt" || string(body) != "CHILD-ARTIFACT-OK" {
			t.Error("sandbox upload changed the file")
		}
		requests <- struct{}{}
		_, _ = w.Write([]byte(`{"file_id":"file-child","original_name":"result.txt","size":17,"artifact_id":"artifact-child"}`))
	}))
	defer server.Close()
	root := t.TempDir()
	config, file := filepath.Join(root, "config.json"), filepath.Join(root, "result.txt")
	if err := os.WriteFile(config, []byte(`invalid saved config`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte("CHILD-ARTIFACT-OK"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MOI_CLI_CONFIG_FILE", config)
	t.Setenv("MOI_ASTRA_BINARY", filepath.Join(root, "missing-astra"))
	t.Setenv("MOI_RUNTIME_AUTHORIZATION", "Bearer synthetic-child-proof")
	t.Setenv("MOI_API_KEY", "synthetic-child-key")
	t.Setenv("MOI_WORKSPACE_ID", "ws-child")
	t.Setenv("MOI_ENDPOINT", server.URL+"/newmoi")
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	output, err := exec.CommandContext(t.Context(), binary, "test-cli-entry", "-o", "file", "upload", "--yes", "--file", file, "--publish-artifact").CombinedOutput()
	if err != nil {
		t.Fatalf("bound upload failed: %v: %s", err, output)
	}
	var result struct {
		FileID     string `json:"file_id"`
		ArtifactID string `json:"artifact_id"`
	}
	if err := json.Unmarshal(output, &result); err != nil || result.FileID != "file-child" || result.ArtifactID != "artifact-child" {
		t.Fatalf("upload result: %s, %v", output, err)
	}
	select {
	case <-requests:
	default:
		t.Fatal("no upload request")
	}
}

func TestSandboxChildRejectsIncompleteCredentialsBeforeRequest(t *testing.T) {
	for _, tc := range []struct {
		name, env, value, code string
	}{
		{name: "missing proof", env: "MOI_RUNTIME_AUTHORIZATION", code: "CLI_RUNTIME_AUTHORIZATION_INVALID"},
		{name: "malformed proof", env: "MOI_RUNTIME_AUTHORIZATION", value: "Bearer synthetic-child-proof\n", code: "CLI_RUNTIME_AUTHORIZATION_INVALID"},
		{name: "missing key", env: "MOI_API_KEY", code: "CLI_RUNTIME_CONNECTION_INCOMPLETE"},
		{name: "missing endpoint", env: "MOI_ENDPOINT", code: "CLI_RUNTIME_CONNECTION_INCOMPLETE"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Error("invalid child made a product request") }))
			defer server.Close()
			t.Setenv("MOI_RUNTIME_AUTHORIZATION", "Bearer synthetic-child-proof")
			t.Setenv("MOI_API_KEY", "synthetic-child-key")
			t.Setenv("MOI_WORKSPACE_ID", "ws-child")
			t.Setenv("MOI_ENDPOINT", server.URL+"/newmoi")
			if tc.env != "" {
				t.Setenv(tc.env, tc.value)
			}
			binary, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			output, err := exec.CommandContext(t.Context(), binary, "test-cli-entry", "catalog", "list").CombinedOutput()
			if err == nil || !strings.Contains(string(output), tc.code) || strings.Contains(string(output), "synthetic-child-") {
				t.Fatalf("invalid identity was not safely classified: %v: %s", err, output)
			}
		})
	}
}

func TestSandboxChildRejectsExplicitConnectionFlags(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { t.Error("conflicting child made a product request") }))
	defer server.Close()
	t.Setenv("MOI_RUNTIME_AUTHORIZATION", "Bearer synthetic-child-proof")
	t.Setenv("MOI_API_KEY", "synthetic-child-key")
	t.Setenv("MOI_WORKSPACE_ID", "ws-child")
	t.Setenv("MOI_ENDPOINT", server.URL+"/newmoi")
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		args []string
	}{
		{name: "explicit endpoint", args: []string{"-endpoint", "https://elsewhere.example.test"}},
		{name: "explicit api key", args: []string{"-api-key", "synthetic-explicit-key"}},
		{name: "both connection flags", args: []string{"-endpoint", "https://elsewhere.example.test", "-api-key", "synthetic-explicit-key"}},
		{name: "matching endpoint with different api key", args: []string{"-endpoint", server.URL + "/newmoi", "-api-key", "synthetic-explicit-key"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{"test-cli-entry"}, tc.args...)
			args = append(args, "catalog", "list")
			output, err := exec.CommandContext(t.Context(), binary, args...).CombinedOutput()
			if err == nil || !strings.Contains(string(output), "CLI_RUNTIME_CONNECTION_CONFLICT") {
				t.Fatalf("conflicting flags were not rejected: %v: %s", err, output)
			}
			if !strings.Contains(string(output), "MOI_ENDPOINT") || !strings.Contains(string(output), "MOI_RUNTIME_AUTHORIZATION") {
				t.Fatalf("conflict error lacks remediation: %s", output)
			}
			if strings.Contains(string(output), "synthetic-explicit-key") || strings.Contains(string(output), "synthetic-child-key") || strings.Contains(string(output), "https://elsewhere.example.test") {
				t.Fatalf("conflict error leaked a connection value: %s", output)
			}
		})
	}
}

func TestDefaultAuthNeverFallsBackToSavedPAT(t *testing.T) {
	requests := make(chan struct{}, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests <- struct{}{}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":"OK","data":{"workspaces":[]}}`))
	}))
	defer server.Close()
	root := t.TempDir()
	config := filepath.Join(root, "config.json")
	body, err := json.Marshal(commands.Config{Endpoint: server.URL, APIKey: "old-pat"})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config, body, 0o600); err != nil {
		t.Fatal(err)
	}
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("MOI_CLI_CONFIG_FILE", config)
	t.Setenv("MOI_AUTH_DIR", filepath.Join(root, "missing-auth"))
	t.Setenv("MOI_ASTRA_BINARY", binary)
	t.Setenv("MOI_ROLE_TEST_SIGNED_OUT", "1")
	for _, mode := range []string{"auto", "moi", "pat"} {
		args := []string{"test-cli-entry"}
		if mode != "auto" {
			args = append(args, "-auth", mode)
		}
		args = append(args, "-o", "-no-interactive", "workspace", "list")
		output, err := exec.Command(binary, args...).CombinedOutput()
		if mode == "pat" {
			if err != nil {
				t.Fatalf("explicit PAT failed: %v: %s", err, output)
			}
			select {
			case <-requests:
			default:
				t.Fatal("explicit PAT made no request")
			}
		} else {
			if err == nil || !strings.Contains(string(output), "astra login") {
				t.Fatalf("expected sign-in guidance: %v: %s", err, output)
			}
			select {
			case <-requests:
				t.Fatal("signed-out native mode used legacy endpoint")
			default:
			}
		}
	}
}

func TestNativeCommandUsesConfiguredWorkspaceRoles(t *testing.T) {
	for _, tc := range []struct {
		name, workspace, role string
		wantWorkspace         string
		args                  []string
	}{
		{name: "login workspace", role: "17", wantWorkspace: "ws-a", args: []string{"list"}},
		{name: "global workspace", workspace: "ws-b", role: "29", wantWorkspace: "ws-b", args: []string{"list"}},
		{name: "subcommand workspace", role: "29", wantWorkspace: "ws-b", args: []string{"list", "--workspace-id", "ws-b"}},
		{name: "no selected role", workspace: "ws-c", wantWorkspace: "ws-c", args: []string{"list"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			requests := make(chan http.Header, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests <- r.Header.Clone()
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"code":"OK","data":{}}`))
			}))
			defer server.Close()
			binary, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			t.Setenv("MOI_ASTRA_BINARY", binary)
			t.Setenv("MOI_ROLE_TEST_ENDPOINT", server.URL+"/newmoi")
			cfg := &commands.Config{
				Endpoint: server.URL + "/newmoi", APIKey: "old-pat", WorkspaceID: "old-workspace",
				WorkspaceCurrentRoleIDs: map[string]json.RawMessage{"ws-a": json.RawMessage(`"17"`), "ws-b": json.RawMessage(`"29"`)},
			}
			if err := executeNativeCommand(cfg, "", false, "", tc.workspace, false, true, true, "catalog", tc.args, 30*time.Second); err != nil {
				t.Fatal(err)
			}
			select {
			case headers := <-requests:
				if got := headers.Get("X-Workspace-ID"); got != tc.wantWorkspace {
					t.Fatalf("workspace = %q, want %q (must not use stale PAT config)", got, tc.wantWorkspace)
				}
				if got := headers.Get("X-MOI-Role-ID"); got != tc.role {
					t.Fatalf("role = %q, want %q", got, tc.role)
				}
				if headers.Get("Authorization") != "Bearer synthetic-access" || headers.Get("X-API-Key") != "" {
					t.Fatal("native command mixed legacy and native credentials")
				}
			default:
				t.Fatal("no product request")
			}
		})
	}
}

func TestNativeCommandRejectsInvalidRolesBeforeHelper(t *testing.T) {
	t.Setenv("MOI_ASTRA_BINARY", filepath.Join(t.TempDir(), "missing-astra"))
	for _, raw := range []string{
		`{"workspace_current_role_ids":{"ws-a":"17","ws-other":"01"}}`,
		`{"workspace_current_role_ids":{"ws-a":17}}`,
		`{"workspace_current_role_ids":[]}`,
	} {
		var cfg commands.Config
		if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
			t.Fatal(err)
		}
		err := executeNativeCommand(&cfg, "", false, "", "ws-a", false, true, true, "catalog", []string{"list"}, 30*time.Second)
		var roleErr *commands.WorkspaceCurrentRoleConfigError
		if !errors.As(err, &roleErr) {
			t.Fatalf("must reject role config before invoking helper: %v", err)
		}
	}
}

func TestNativeCommandEndpointPrecedence(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":"OK","data":{}}`))
	}))
	defer server.Close()
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("MOI_ASTRA_BINARY", binary)
	endpoint := server.URL + "/newmoi"
	t.Setenv("MOI_ROLE_TEST_ENDPOINT", endpoint)
	for _, tc := range []struct {
		name, configured, flag string
		explicit, wantError    bool
	}{
		{"no saved endpoint", "", defaultEndpoint, false, false},
		{"matching saved endpoint", endpoint, defaultEndpoint, false, false},
		{"different saved endpoint ignored", "https://old.example.test", defaultEndpoint, false, false},
		{"explicit matching overrides saved", "https://old.example.test", endpoint, true, false},
		{"explicit mismatched rejected", endpoint, "https://other.example.test", true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := executeNativeCommand(&commands.Config{Endpoint: tc.configured}, tc.flag, tc.explicit, "", "ws-a", false, true, true, "catalog", []string{"list"}, 30*time.Second)
			if (err != nil) != tc.wantError {
				t.Fatalf("endpoint binding error=%v, wantError=%v", err, tc.wantError)
			}
		})
	}
}

func TestWorkspaceRoleOptionsFreezeConfig(t *testing.T) {
	cfg := &commands.Config{WorkspaceCurrentRoleIDs: map[string]json.RawMessage{"ws-a": json.RawMessage(`"17"`)}}
	opts, err := workspaceRoleOptions(cfg)
	if err != nil {
		t.Fatal(err)
	}
	cfg.WorkspaceCurrentRoleIDs["ws-a"] = json.RawMessage(`"29"`)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-MOI-Role-ID"); got != "17" {
			t.Errorf("frozen role = %q, want 17", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":"OK","data":{}}`))
	}))
	defer server.Close()
	client, err := sdk.NewWithPersonalAccessToken(server.URL, "synthetic-pat", opts...)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	if _, err := client.Catalog("ws-a").List(t.Context()); err != nil {
		t.Fatal(err)
	}
}
