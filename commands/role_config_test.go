package commands

import (
	"strings"
	"testing"
)

func TestWorkspaceCurrentRoleIssuesAreSortedAndActionable(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	cfg := &Config{
		APIKey: "opaque-pat-that-must-not-appear",
		WorkspaceCurrentRoleIDs: rawCurrentRoleMappings(t, map[string]any{
			"ws-z": "01",
			"ws-a": "9223372036854775808",
			"":     "12",
		}),
	}

	issues := WorkspaceCurrentRoleIssues(cfg)
	if len(issues) != 3 {
		t.Fatalf("issues = %#v, want three invalid mappings", issues)
	}
	if issues[0].WorkspaceID != "" || issues[1].WorkspaceID != "ws-a" || issues[2].WorkspaceID != "ws-z" {
		t.Fatalf("issues are not sorted: %#v", issues)
	}
	_, err := WorkspaceCurrentRoleSnapshot(cfg)
	if err == nil {
		t.Fatal("WorkspaceCurrentRoleSnapshot() error = nil")
	}
	message := err.Error()
	if strings.Contains(message, cfg.APIKey) {
		t.Fatalf("validation error leaked PAT: %s", message)
	}
	for _, expected := range []string{
		"No request was sent.",
		"config set-role",
		"config unset-role",
		"ws-a",
		"ws-z",
	} {
		if !strings.Contains(message, expected) {
			t.Fatalf("validation error missing %q: %s", expected, message)
		}
	}
	if strings.Index(message, "ws-a") > strings.Index(message, "ws-z") {
		t.Fatalf("validation error order = %s", message)
	}
}

func TestWorkspaceCurrentRoleSnapshotRejectsNonStringMappings(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	cfg := &Config{
		WorkspaceCurrentRoleIDs: rawCurrentRoleMappings(t, map[string]any{
			"ws-number": 12,
			"ws-object": map[string]any{"role": "27"},
		}),
	}

	roles, err := WorkspaceCurrentRoleSnapshot(cfg)
	if roles != nil || err == nil {
		t.Fatalf("snapshot=%#v error=%v, want nil snapshot and validation error", roles, err)
	}
	message := err.Error()
	for _, expected := range []string{"ws-number", "ws-object", "JSON string", "No request was sent."} {
		if !strings.Contains(message, expected) {
			t.Fatalf("validation error missing %q: %s", expected, message)
		}
	}
}

func TestInvalidWorkspaceCurrentRoleValuesAreRedacted(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	const secret = "moi_mapping_pat_must_not_appear"
	cfg := &Config{
		WorkspaceCurrentRoleIDs: rawCurrentRoleMappings(t, map[string]any{
			"ws-string": secret,
			"ws-object": map[string]any{"token": secret},
		}),
	}

	issues := WorkspaceCurrentRoleIssues(cfg)
	if len(issues) != 2 {
		t.Fatalf("issues = %#v, want two invalid mappings", issues)
	}
	for _, issue := range issues {
		if issue.RoleID != redactedWorkspaceCurrentRole {
			t.Fatalf("issue role = %q, want %q", issue.RoleID, redactedWorkspaceCurrentRole)
		}
	}
	_, err := WorkspaceCurrentRoleSnapshot(cfg)
	if err == nil {
		t.Fatal("WorkspaceCurrentRoleSnapshot() error = nil")
	}
	message := err.Error()
	if strings.Contains(message, secret) {
		t.Fatalf("validation error leaked invalid mapping: %s", message)
	}
	if strings.Count(message, `role "<redacted>"`) != 2 {
		t.Fatalf("validation error did not redact every invalid mapping: %s", message)
	}
}
