package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestSaveConfigRestrictsExistingPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not support Unix file permission bits")
	}
	for _, mode := range []os.FileMode{0600, 0644, 0666} {
		t.Run(fmt.Sprintf("%04o", mode), func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.json")
			t.Setenv("MOI_CLI_CONFIG_FILE", path)
			if err := os.WriteFile(path, bytes.Repeat([]byte(" "), 4096), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.Chmod(path, mode); err != nil {
				t.Fatal(err)
			}
			cfg := &Config{APIKey: "synthetic-test-token", WorkspaceID: "ws-test"}
			if err := SaveConfig(cfg); err != nil {
				t.Fatal(err)
			}
			info, err := os.Stat(path)
			if err != nil {
				t.Fatal(err)
			}
			if got := info.Mode().Perm(); got != 0600 {
				t.Errorf("config mode = %04o, want 0600", got)
			}
			got, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			want, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, want) {
				t.Fatal("saved configuration differs from the requested JSON")
			}
		})
	}
}

func TestSaveConfigDoesNotExposeCredentialsToPreexistingReader(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not support Unix file permission bits")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	t.Setenv("MOI_CLI_CONFIG_FILE", path)
	old := []byte(`{"api_key":"old-token"}`)
	if err := os.WriteFile(path, old, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0644); err != nil {
		t.Fatal(err)
	}
	// A reader that opened the broad-mode file before the save.
	reader, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	if err := SaveConfig(&Config{APIKey: "new-secret-token"}); err != nil {
		t.Fatal(err)
	}

	stale, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(stale, []byte("new-secret-token")) {
		t.Fatal("a pre-existing reader observed the rewritten credential")
	}
	if !bytes.Equal(stale, old) {
		t.Fatalf("pre-existing reader content = %q, want the old file unchanged", stale)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Errorf("config mode = %04o, want 0600", got)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("temporary files left behind: %v", entries)
	}
}

func TestInstallationConfigIsolation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "client", "moi-cli.json")
	t.Setenv("MOI_CLI_CONFIG_FILE", path)
	if got, err := GetConfigPath(); err != nil || got != path {
		t.Fatalf("GetConfigPath() = %q, %v", got, err)
	}
	if cfg, err := LoadConfig(); err != nil || cfg != nil {
		t.Fatalf("missing installation config must not load legacy config: %#v, %v", cfg, err)
	}
	if err := SaveConfig(&Config{Endpoint: "http://127.0.0.1:29050/newmoi"}); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig()
	if err != nil || cfg.Endpoint != "http://127.0.0.1:29050/newmoi" || cfg.APIKey != "" {
		t.Fatalf("installation config = %#v, %v", cfg, err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("private config mode: %v, %v", info, err)
	}
}

func TestInstallationConfigInvalidPathFailsClosed(t *testing.T) {
	for _, path := range []string{"", "relative.json", "/", "/tmp/../config.json"} {
		t.Run(path, func(t *testing.T) {
			t.Setenv("MOI_CLI_CONFIG_FILE", path)
			if _, err := LoadConfig(); err == nil || err.Error() != "CLI_CONFIG_PATH_INVALID" {
				t.Fatalf("invalid explicit path must not load legacy config: %v", err)
			}
			if err := SaveConfig(&Config{}); err == nil {
				t.Fatal("invalid explicit path accepted for write")
			}
		})
	}
}

func TestLoadLegacyConfigWithoutWorkspaceCurrentRoles(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	path := filepath.Join(home, ".moi-cli.json")
	if err := os.WriteFile(path, []byte(`{"endpoint":"https://product.example/newmoi","api_key":"opaque-pat","workspace_id":"ws-a"}`), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Endpoint != "https://product.example/newmoi" || cfg.APIKey != "opaque-pat" || cfg.WorkspaceID != "ws-a" {
		t.Fatalf("legacy config = %#v", cfg)
	}
	if len(cfg.WorkspaceCurrentRoleIDs) != 0 {
		t.Fatalf("workspace current roles = %#v, want empty", cfg.WorkspaceCurrentRoleIDs)
	}
}

func TestSaveConfigPersistsWorkspaceCurrentRolesWithPrivateMode(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	cfg := &Config{
		Endpoint:    "https://product.example/newmoi",
		APIKey:      "opaque-pat",
		WorkspaceID: "ws-a",
		WorkspaceCurrentRoleIDs: rawCurrentRoleMappings(t, map[string]any{
			"ws-a": "12",
			"ws-b": "27",
		}),
	}
	if err := SaveConfig(cfg); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if configuredRoleID(t, loaded, "ws-a") != "12" || configuredRoleID(t, loaded, "ws-b") != "27" {
		t.Fatalf("workspace current roles = %#v", loaded.WorkspaceCurrentRoleIDs)
	}
	info, err := os.Stat(filepath.Join(home, ".moi-cli.json"))
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Fatalf("config mode = %o, want 600", got)
	}
}

func TestLoadConfigDefersMalformedWorkspaceCurrentRoleContainerForRepair(t *testing.T) {
	for name, rawContainer := range map[string]string{
		"array":   `["opaque-value"]`,
		"boolean": `true`,
		"null":    `null`,
		"number":  `12`,
		"string":  `"12"`,
	} {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			path := filepath.Join(home, ".moi-cli.json")
			data := []byte(`{"endpoint":"https://product.example/newmoi","workspace_current_role_ids":` + rawContainer + `}`)
			if err := os.WriteFile(path, data, 0600); err != nil {
				t.Fatal(err)
			}

			cfg, err := LoadConfig()
			if err != nil {
				t.Fatal(err)
			}
			if !cfg.hasInvalidWorkspaceCurrentRoleContainer() {
				t.Fatalf("malformed %s container was not retained for repair", name)
			}
			if err := SaveConfig(cfg); err != nil {
				t.Fatal(err)
			}
			reloaded, err := LoadConfig()
			if err != nil {
				t.Fatal(err)
			}
			if !reloaded.hasInvalidWorkspaceCurrentRoleContainer() {
				t.Fatalf("malformed %s container was silently discarded", name)
			}
		})
	}
}

func TestLoadConfigStillRejectsMalformedExistingFields(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	path := filepath.Join(home, ".moi-cli.json")
	if err := os.WriteFile(path, []byte(`{"api_key":12}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadConfig(); err == nil {
		t.Fatal("LoadConfig() error = nil for non-string api_key")
	}
}

func rawCurrentRoleMappings(t *testing.T, values map[string]any) map[string]json.RawMessage {
	t.Helper()
	mappings := make(map[string]json.RawMessage, len(values))
	for workspaceID, value := range values {
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		mappings[workspaceID] = data
	}
	return mappings
}

func configuredRoleID(t *testing.T, cfg *Config, workspaceID string) string {
	t.Helper()
	roleID, err := decodeWorkspaceCurrentRoleID(cfg.WorkspaceCurrentRoleIDs[workspaceID])
	if err != nil {
		t.Fatal(err)
	}
	return roleID
}
