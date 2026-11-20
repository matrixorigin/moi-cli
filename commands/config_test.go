package commands

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func TestConfigSetRoleValidatesAndPersists(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := ExecuteConfigCommand(nil, []string{"set-role", "ws-b", "27"}); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if got := configuredRoleID(t, cfg, "ws-b"); got != "27" {
		t.Fatalf("workspace ws-b role = %q, want 27", got)
	}

	if err := ExecuteConfigCommand(nil, []string{"set-role", "ws-b", "9223372036854775808"}); !errors.Is(err, sdk.ErrCurrentRoleInvalid) {
		t.Fatalf("overflow set-role error = %v, want %v", err, sdk.ErrCurrentRoleInvalid)
	}
	cfg, err = LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if got := configuredRoleID(t, cfg, "ws-b"); got != "27" {
		t.Fatalf("invalid set-role changed role to %q", got)
	}
}

func TestConfigUnsetRoleRejectsMissingMapping(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := ExecuteConfigCommand(nil, []string{"unset-role", "ws-a"}); err == nil {
		t.Fatal("unset-role missing mapping error = nil")
	}
	if err := ExecuteConfigCommand(nil, []string{"set-role", "ws-a", "12"}); err != nil {
		t.Fatal(err)
	}
	if err := ExecuteConfigCommand(nil, []string{"unset-role", "ws-a"}); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := cfg.WorkspaceCurrentRoleIDs["ws-a"]; ok {
		t.Fatalf("workspace role remains after unset: %#v", cfg.WorkspaceCurrentRoleIDs)
	}
}

func TestConfigShowSortsWorkspaceCurrentRolesAndMasksPAT(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := SaveConfig(&Config{
		APIKey: "opaque-personal-access-token",
		WorkspaceCurrentRoleIDs: rawCurrentRoleMappings(t, map[string]any{
			"ws-z": "27",
			"ws-a": "12",
		}),
	}); err != nil {
		t.Fatal(err)
	}

	output := captureStdout(t, func() {
		if err := ExecuteConfigCommand(nil, []string{"show"}); err != nil {
			t.Fatal(err)
		}
	})
	if strings.Contains(output, "opaque-personal-access-token") {
		t.Fatalf("config show leaked PAT: %s", output)
	}
	if !strings.Contains(output, "api_key:      ********") {
		t.Fatalf("config show did not mask PAT: %s", output)
	}
	indexA := strings.Index(output, "ws-a: 12")
	indexZ := strings.Index(output, "ws-z: 27")
	if indexA < 0 || indexZ < 0 || indexA >= indexZ {
		t.Fatalf("workspace current roles are not sorted: %s", output)
	}
}

func TestConfigShowMasksShortPAT(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := SaveConfig(&Config{APIKey: "short"}); err != nil {
		t.Fatal(err)
	}

	output := captureStdout(t, func() {
		if err := ExecuteConfigCommand(nil, []string{"show"}); err != nil {
			t.Fatal(err)
		}
	})
	if strings.Contains(output, "short") || !strings.Contains(output, "api_key:      ********") {
		t.Fatalf("config show did not mask short PAT: %s", output)
	}
}

func TestConfigShowUsesInjectedRuntimeConnection(t *testing.T) {
	t.Setenv("MOI_CLI_CONFIG_FILE", filepath.Join(t.TempDir(), "config.json"))
	if err := SaveConfig(&Config{Endpoint: "https://saved.example.test/newmoi", APIKey: "saved-pat", WorkspaceID: "ws-saved"}); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MOI_RUNTIME_AUTHORIZATION", "Bearer synthetic-proof")
	t.Setenv("MOI_ENDPOINT", "https://runtime.example.test/newmoi")
	t.Setenv("MOI_API_KEY", "runtime-pat")
	t.Setenv("MOI_WORKSPACE_ID", "ws-runtime")
	output := captureStdout(t, func() {
		if err := ExecuteConfigCommand(nil, []string{"show"}); err != nil {
			t.Fatal(err)
		}
	})
	for _, expected := range []string{"Runtime configuration (environment)", "https://runtime.example.test/newmoi", "workspace_id: ws-runtime", "api_key:      ********"} {
		if !strings.Contains(output, expected) {
			t.Fatalf("runtime configuration missing %q: %s", expected, output)
		}
	}
	for _, private := range []string{"saved.example.test", "ws-saved", "saved-pat", "runtime-pat", "synthetic-proof"} {
		if strings.Contains(output, private) {
			t.Fatal("config show exposed saved configuration or a credential")
		}
	}
	cfg, err := LoadConfig()
	if err != nil || cfg.WorkspaceID != "ws-saved" || cfg.APIKey != "saved-pat" {
		t.Fatal("show changed saved configuration")
	}
}

func TestConfigShowMarksInvalidWorkspaceCurrentRoles(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	const secret = "moi_mapping_pat_must_not_appear"
	if err := SaveConfig(&Config{
		WorkspaceCurrentRoleIDs: rawCurrentRoleMappings(t, map[string]any{
			"ws-a":      "12",
			"ws-b":      "01",
			"ws-secret": map[string]any{"token": secret},
		}),
	}); err != nil {
		t.Fatal(err)
	}

	output := captureStdout(t, func() {
		if err := ExecuteConfigCommand(nil, []string{"show"}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(output, "ws-b: <redacted> [invalid:") || !strings.Contains(output, "ws-secret: <redacted> [invalid:") {
		t.Fatalf("config show did not mark invalid role: %s", output)
	}
	if strings.Contains(output, secret) {
		t.Fatalf("config show leaked invalid mapping: %s", output)
	}
}

func TestConfigSetRoleRepairsOneInvalidMappingWithoutChangingOthers(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := SaveConfig(&Config{
		WorkspaceCurrentRoleIDs: rawCurrentRoleMappings(t, map[string]any{
			"ws-a": "01",
			"ws-b": "27",
		}),
	}); err != nil {
		t.Fatal(err)
	}

	if err := ExecuteConfigCommand(nil, []string{"set-role", "ws-a", "12"}); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if configuredRoleID(t, cfg, "ws-a") != "12" || configuredRoleID(t, cfg, "ws-b") != "27" {
		t.Fatalf("repaired mappings = %#v", cfg.WorkspaceCurrentRoleIDs)
	}
}

func TestConfigCommandsRepairMalformedWorkspaceCurrentRoleContainer(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	path := filepath.Join(home, ".moi-cli.json")
	writeMalformedContainer := func(value string) {
		t.Helper()
		data := []byte(`{"api_key":"opaque-pat","workspace_current_role_ids":` + value + `}`)
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
	}

	writeMalformedContainer(`[12]`)
	output := captureStdout(t, func() {
		if err := ExecuteConfigCommand(nil, []string{"show"}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(output, "workspace_current_role_ids: (invalid:") || !strings.Contains(output, "config set-role") || !strings.Contains(output, "config unset-role") {
		t.Fatalf("config show did not expose repair commands: %s", output)
	}
	if err := ExecuteConfigCommand(nil, []string{"set-role", "ws-a", "12"}); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.hasInvalidWorkspaceCurrentRoleContainer() || configuredRoleID(t, cfg, "ws-a") != "12" {
		t.Fatalf("set-role did not replace malformed container: %#v", cfg)
	}

	writeMalformedContainer(`12`)
	if err := ExecuteConfigCommand(nil, []string{"unset-role", "ws-a"}); err != nil {
		t.Fatal(err)
	}
	cfg, err = LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.hasInvalidWorkspaceCurrentRoleContainer() || len(cfg.WorkspaceCurrentRoleIDs) != 0 {
		t.Fatalf("unset-role did not clear malformed container: %#v", cfg)
	}
}

func captureStdout(t *testing.T, run func()) string {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdout
	os.Stdout = writer
	t.Cleanup(func() { os.Stdout = original })

	run()
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = original
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if err := reader.Close(); err != nil {
		t.Fatal(err)
	}
	return string(data)
}
