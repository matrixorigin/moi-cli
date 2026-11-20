package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/matrixorigin/matrixflow/sdk/go-worker-sdk/agentpackage"
)

// Preserve the original cli_agent_package_test's unauthenticated local
// lifecycle, and exercise the installed CLI entry point beyond help output.
func TestCLIAgentPackageLocalLifecycleNeedsNoAuthentication(t *testing.T) {
	root := t.TempDir()
	binary := filepath.Join(root, "moi-cli")
	build := exec.CommandContext(t.Context(), "go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build CLI: %v\n%s", err, output)
	}
	home := filepath.Join(root, "home")
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(home, ".moi-cli.json")
	before := []byte("{broken-config")
	if err := os.WriteFile(configPath, before, 0o600); err != nil {
		t.Fatal(err)
	}
	var requests atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		http.Error(w, "local package commands must not contact a server", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)
	run := func(args ...string) ([]byte, error) {
		command := exec.CommandContext(t.Context(), binary, append([]string{"-endpoint", server.URL, "-auth", "moi", "-o"}, args...)...)
		command.Env = append(os.Environ(), "HOME="+home, "MOI_AUTH_DIR="+home, "MOI_ASTRA_BINARY="+filepath.Join(home, "missing-helper"), "MOI_AGENT_CALL=1")
		return command.CombinedOutput()
	}
	source := filepath.Join(root, "source")
	archive := filepath.Join(root, "agent.moiagent")
	unpacked := filepath.Join(root, "unpacked")
	for _, step := range [][]string{
		{"agent", "init", "--output", source, "--agent-id", "com.example.cli-local-agent", "--name", "CLI Local Agent", "--model", "qwen3.6-plus"},
		{"agent", "verify", "--source", source},
	} {
		output, err := run(step...)
		if err != nil || !json.Valid(output) {
			t.Fatalf("%v: %v\n%s", step, err, output)
		}
	}
	manifest, err := os.ReadFile(filepath.Join(source, "agent.yaml"))
	if err != nil || !bytes.Contains(manifest, []byte(`id: "com.example.cli-local-agent"`)) || !bytes.Contains(manifest, []byte(`model: "qwen3.6-plus"`)) || !bytes.Contains(manifest, []byte("moiResourceRefs:")) {
		t.Fatalf("initial manifest: %v\n%s", err, manifest)
	}
	role, err := os.ReadFile(filepath.Join(source, "prompts", "role.md"))
	if err != nil || !bytes.Contains(role, []byte("简洁回答问题")) {
		t.Fatalf("initial instruction: %v\n%s", err, role)
	}
	toolDir := filepath.Join(source, "custom", "tools", "python-hello")
	if err := os.MkdirAll(toolDir, 0o755); err != nil {
		t.Fatal(err)
	}
	tool := []byte("import json,sys\nprint(json.load(sys.stdin)['message'])\n")
	if err := os.WriteFile(filepath.Join(toolDir, "tool.py"), tool, 0o644); err != nil {
		t.Fatal(err)
	}
	for _, step := range [][]string{
		{"agent", "verify", "--source", source},
		{"agent", "build", "--source", source, "--output", archive},
		{"agent", "verify", "--file", archive},
		{"agent", "unpack", "--file", archive, "--output", unpacked},
	} {
		output, err := run(step...)
		if err != nil || !json.Valid(output) {
			t.Fatalf("%v: %v\n%s", step, err, output)
		}
	}
	data, err := os.ReadFile(archive)
	if err != nil {
		t.Fatal(err)
	}
	pkg, err := agentpackage.Parse(data, version)
	if err != nil || pkg.Manifest.ID != "com.example.cli-local-agent" {
		t.Fatalf("actual package=%+v error=%v", pkg, err)
	}
	artifactPath := agentpackage.CustomToolArtifactPath("custom/tools/python-hello")
	artifactManifest := agentpackage.CustomToolArtifactManifestPath("custom/tools/python-hello")
	if len(pkg.Files[artifactPath]) == 0 || len(pkg.Files[artifactManifest]) == 0 {
		t.Fatal("built package lacks executable custom tool artifacts")
	}
	for _, path := range []string{agentpackage.ChecksumPath, artifactPath, artifactManifest} {
		if _, err := os.Stat(filepath.Join(unpacked, filepath.FromSlash(path))); !os.IsNotExist(err) {
			t.Fatalf("unpacked generated file %s error=%v", path, err)
		}
	}
	unpackedTool, err := os.ReadFile(filepath.Join(unpacked, "custom", "tools", "python-hello", "tool.py"))
	if err != nil || !bytes.Equal(tool, unpackedTool) {
		t.Fatalf("editable custom tool differs: %v\n%s", err, unpackedTool)
	}
	rebuilt := filepath.Join(root, "rebuilt.moiagent")
	if output, err := run("agent", "build", "--source", unpacked, "--output", rebuilt); err != nil {
		t.Fatalf("rebuild unpacked source: %v\n%s", err, output)
	}
	rebuiltData, err := os.ReadFile(rebuilt)
	if err != nil || !bytes.Equal(data, rebuiltData) {
		t.Fatalf("non-deterministic editable source rebuild: %v", err)
	}
	for _, failure := range []struct {
		args []string
		want string
	}{
		{[]string{"agent", "init", "--output", source, "--agent-id", "com.example.cli-local-agent"}, "already exists"},
		{[]string{"agent", "build", "--source", source, "--output", archive}, "already exists"},
		{[]string{"agent", "verify", "--source", source, "--file", archive}, "exactly one"},
		{[]string{"agent", "unpack", "--file", archive, "--output", unpacked}, "already exists"},
	} {
		output, err := run(failure.args...)
		if err == nil || !strings.Contains(string(output), failure.want) {
			t.Fatalf("expected failure %v: %v\n%s", failure.args, err, output)
		}
	}
	corrupted := filepath.Join(root, "corrupted.moiagent")
	if err := os.WriteFile(corrupted, []byte("invalid package"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, sub := range []string{"verify", "unpack"} {
		args := []string{"agent", sub, "--file", corrupted}
		if sub == "unpack" {
			args = append(args, "--output", filepath.Join(root, "must-not-exist"))
		}
		if output, err := run(args...); err == nil || bytes.Contains(output, []byte("api-key")) {
			t.Fatalf("corrupt package %s must fail locally: %v\n%s", sub, err, output)
		}
	}
	if requests.Load() != 0 {
		t.Fatalf("local commands contacted server %d times", requests.Load())
	}
	if after, err := os.ReadFile(configPath); err != nil || !bytes.Equal(before, after) {
		t.Fatalf("local commands changed user config: %v", err)
	}
}
