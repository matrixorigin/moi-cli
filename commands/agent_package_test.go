package commands

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/matrixorigin/matrixflow/sdk/go-worker-sdk/agentpackage"
)

func TestCollectAgentPackageSourceFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "prompts"), 0o755); err != nil {
		t.Fatalf("mkdir prompts: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "agent.yaml"), []byte("manifest"), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "prompts", "main.md"), []byte("prompt"), 0o644); err != nil {
		t.Fatalf("write prompt: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "assistant.moiagent"), []byte("previous package"), 0o644); err != nil {
		t.Fatalf("write package output: %v", err)
	}

	files, err := collectAgentPackageSourceFiles(dir)
	if err != nil {
		t.Fatalf("collectAgentPackageSourceFiles() error = %v", err)
	}
	if string(files["agent.yaml"]) != "manifest" {
		t.Fatalf("agent.yaml = %q", string(files["agent.yaml"]))
	}
	if string(files["prompts/main.md"]) != "prompt" {
		t.Fatalf("prompts/main.md = %q", string(files["prompts/main.md"]))
	}
	if _, ok := files["assistant.moiagent"]; ok {
		t.Fatalf("collected package output file: %+v", files)
	}
}

func TestBuildAgentInitTemplateFilesDescriptionPresence(t *testing.T) {
	withoutDescription, err := buildAgentInitTemplateFiles(agentInitTemplateInput{
		AgentID:     "com.example.agent",
		Name:        "Example Agent",
		Version:     "1.0.0",
		Model:       "qwen",
		Instruction: "Answer briefly.",
	})
	if err != nil {
		t.Fatalf("buildAgentInitTemplateFiles() without description error = %v", err)
	}
	if strings.Contains(string(withoutDescription["agent.yaml"]), "\ndescription:") {
		t.Fatalf("agent.yaml unexpectedly contains description:\n%s", string(withoutDescription["agent.yaml"]))
	}

	description := "CLI generated description"
	withDescription, err := buildAgentInitTemplateFiles(agentInitTemplateInput{
		AgentID:     "com.example.agent",
		Name:        "Example Agent",
		Description: &description,
		Version:     "1.0.0",
		Model:       "qwen",
		Instruction: "Answer briefly.",
	})
	if err != nil {
		t.Fatalf("buildAgentInitTemplateFiles() with description error = %v", err)
	}
	if !strings.Contains(string(withDescription["agent.yaml"]), "description: \"CLI generated description\"\n") {
		t.Fatalf("agent.yaml missing description:\n%s", string(withDescription["agent.yaml"]))
	}
}

func TestWriteAgentPackageOutputRejectsExistingWithoutOverwrite(t *testing.T) {
	output := filepath.Join(t.TempDir(), "assistant.moiagent")
	if err := os.WriteFile(output, []byte("existing"), 0o644); err != nil {
		t.Fatalf("write existing output: %v", err)
	}
	err := writeAgentPackageOutput(output, []byte("new"), false)
	if err == nil {
		t.Fatal("writeAgentPackageOutput() error = nil, want existing file error")
	}
	data, readErr := os.ReadFile(output)
	if readErr != nil {
		t.Fatalf("read output: %v", readErr)
	}
	if string(data) != "existing" {
		t.Fatalf("output changed to %q, want existing", string(data))
	}
}

func TestRunAgentVerifySource(t *testing.T) {
	dir := writeAgentPackageSourceForTest(t)
	if err := runAgentVerify(&Context{RunningMOIVersion: "3.0.4"}, []string{"--source", dir}); err != nil {
		t.Fatalf("runAgentVerify(--source) error = %v", err)
	}
}

func TestRunAgentVerifyRequiresExactlyOneInput(t *testing.T) {
	dir := writeAgentPackageSourceForTest(t)
	output := filepath.Join(t.TempDir(), "assistant.moiagent")
	data, _, err := buildAgentPackageFromSource(dir, "3.0.4")
	if err != nil {
		t.Fatalf("buildAgentPackageFromSource() error = %v", err)
	}
	if err := os.WriteFile(output, data, 0o644); err != nil {
		t.Fatalf("write package: %v", err)
	}

	err = runAgentVerify(&Context{RunningMOIVersion: "3.0.4"}, []string{"--source", dir, "--file", output})
	if err == nil || !strings.Contains(err.Error(), "exactly one") {
		t.Fatalf("runAgentVerify(--source --file) error = %v, want exactly one", err)
	}
}

func TestRunAgentUnpackWritesEditableSource(t *testing.T) {
	dir := writeAgentPackageSourceForTest(t)
	toolDir := filepath.Join(dir, "custom", "tools", "python-hello")
	if err := os.MkdirAll(toolDir, 0o755); err != nil {
		t.Fatalf("mkdir custom tool: %v", err)
	}
	if err := os.WriteFile(filepath.Join(toolDir, "tool.py"), []byte("print('hello')\n"), 0o644); err != nil {
		t.Fatalf("write custom tool: %v", err)
	}
	data, pkg, err := buildAgentPackageFromSource(dir, "3.0.4")
	if err != nil {
		t.Fatalf("buildAgentPackageFromSource() error = %v", err)
	}
	artifactPath := agentpackage.CustomToolArtifactPath("custom/tools/python-hello")
	manifestPath := agentpackage.CustomToolArtifactManifestPath("custom/tools/python-hello")
	if len(pkg.Files[artifactPath]) == 0 || len(pkg.Files[manifestPath]) == 0 {
		t.Fatalf("package missing generated custom tool artifacts")
	}
	packagePath := filepath.Join(t.TempDir(), "assistant.moiagent")
	if err := os.WriteFile(packagePath, data, 0o644); err != nil {
		t.Fatalf("write package: %v", err)
	}
	output := filepath.Join(t.TempDir(), "source")
	if err := runAgentUnpack(&Context{RunningMOIVersion: "3.0.4"}, []string{"--file", packagePath, "--output", output}); err != nil {
		t.Fatalf("runAgentUnpack() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(output, "checksums.txt")); !os.IsNotExist(err) {
		t.Fatalf("checksums.txt stat error = %v, want not exist", err)
	}
	if _, err := os.Stat(filepath.Join(output, filepath.FromSlash(artifactPath))); !os.IsNotExist(err) {
		t.Fatalf("artifact stat error = %v, want not exist", err)
	}
	if _, err := os.Stat(filepath.Join(output, filepath.FromSlash(manifestPath))); !os.IsNotExist(err) {
		t.Fatalf("artifact manifest stat error = %v, want not exist", err)
	}
	if _, err := os.Stat(filepath.Join(output, "custom", "tools", "python-hello", "tool.py")); err != nil {
		t.Fatalf("custom tool source missing after unpack: %v", err)
	}
	if _, _, err := buildAgentPackageFromSource(output, "3.0.4"); err != nil {
		t.Fatalf("buildAgentPackageFromSource(unpacked) error = %v", err)
	}
}

func TestBuildAgentPackageFromSourceMaterializesCustomToolArtifacts(t *testing.T) {
	dir := writeAgentPackageSourceForTest(t)
	toolDir := filepath.Join(dir, "custom", "tools", "go-hello")
	if err := os.MkdirAll(toolDir, 0o755); err != nil {
		t.Fatalf("mkdir custom tool: %v", err)
	}
	if err := os.WriteFile(filepath.Join(toolDir, "main.go"), []byte(`package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`), 0o644); err != nil {
		t.Fatalf("write custom tool: %v", err)
	}

	_, pkg, err := buildAgentPackageFromSource(dir, "3.0.4")
	if err != nil {
		t.Fatalf("buildAgentPackageFromSource() error = %v", err)
	}
	artifactPath := agentpackage.CustomToolArtifactPath("custom/tools/go-hello")
	manifestPath := agentpackage.CustomToolArtifactManifestPath("custom/tools/go-hello")
	if len(pkg.Files[artifactPath]) == 0 {
		t.Fatalf("artifact %s missing", artifactPath)
	}
	manifest, err := agentpackage.ParseCustomToolArtifactManifest(pkg.Files[manifestPath])
	if err != nil {
		t.Fatalf("ParseCustomToolArtifactManifest() error = %v", err)
	}
	if manifest.Language != "go" || manifest.EntryFile != "main.go" || len(manifest.Platforms) == 0 {
		t.Fatalf("artifact manifest = %+v", manifest)
	}
	if !customToolArtifactHasPlatform(manifest, "linux", "amd64") {
		t.Fatalf("artifact manifest missing linux/amd64 platform: %+v", manifest.Platforms)
	}
	if !customToolArtifactHasPlatform(manifest, "linux", "arm64") {
		t.Fatalf("artifact manifest missing linux/arm64 platform: %+v", manifest.Platforms)
	}
	if !customToolArtifactHasPlatform(manifest, "darwin", "amd64") {
		t.Fatalf("artifact manifest missing darwin/amd64 platform: %+v", manifest.Platforms)
	}
	if !customToolArtifactHasPlatform(manifest, "darwin", "arm64") {
		t.Fatalf("artifact manifest missing darwin/arm64 platform: %+v", manifest.Platforms)
	}
}

func TestBuildAgentPackageFromSourceMaterializesPythonToolForLocalAndLinuxPlatforms(t *testing.T) {
	dir := writeAgentPackageSourceForTest(t)
	toolDir := filepath.Join(dir, "custom", "tools", "python-hello")
	if err := os.MkdirAll(toolDir, 0o755); err != nil {
		t.Fatalf("mkdir custom tool: %v", err)
	}
	if err := os.WriteFile(filepath.Join(toolDir, "tool.py"), []byte("print('hello')\n"), 0o644); err != nil {
		t.Fatalf("write custom tool: %v", err)
	}

	_, pkg, err := buildAgentPackageFromSource(dir, "3.0.4")
	if err != nil {
		t.Fatalf("buildAgentPackageFromSource() error = %v", err)
	}
	manifestPath := agentpackage.CustomToolArtifactManifestPath("custom/tools/python-hello")
	manifest, err := agentpackage.ParseCustomToolArtifactManifest(pkg.Files[manifestPath])
	if err != nil {
		t.Fatalf("ParseCustomToolArtifactManifest() error = %v", err)
	}
	for _, platform := range []struct {
		os   string
		arch string
	}{
		{os: "linux", arch: "amd64"},
		{os: "linux", arch: "arm64"},
		{os: "darwin", arch: "amd64"},
		{os: "darwin", arch: "arm64"},
	} {
		if !customToolArtifactHasPlatform(manifest, platform.os, platform.arch) {
			t.Fatalf("artifact manifest missing %s/%s platform: %+v", platform.os, platform.arch, manifest.Platforms)
		}
	}
}

func TestBuildAgentPackageFromSourceBuildsSlackGIFCreatorSystemPackage(t *testing.T) {
	sourceDir := filepath.Join("testdata", "agent-package", "slack-gif-creator")
	_, pkg, err := buildAgentPackageFromSource(sourceDir, "3.0.4")
	if err != nil {
		t.Fatalf("buildAgentPackageFromSource() error = %v", err)
	}
	artifactPath := agentpackage.CustomToolArtifactPath("custom/tools/gif-create")
	manifestPath := agentpackage.CustomToolArtifactManifestPath("custom/tools/gif-create")
	if len(pkg.Files[artifactPath]) == 0 || len(pkg.Files[manifestPath]) == 0 {
		t.Fatalf("package missing gif-create generated artifacts")
	}
	manifest, err := agentpackage.ParseCustomToolArtifactManifest(pkg.Files[manifestPath])
	if err != nil {
		t.Fatalf("ParseCustomToolArtifactManifest() error = %v", err)
	}
	if manifest.Language != "python" || manifest.EntryFile != "tool.py" {
		t.Fatalf("artifact manifest = %+v", manifest)
	}
	skills, err := agentpackage.CustomSkillSpecs(pkg.Files)
	if err != nil {
		t.Fatalf("CustomSkillSpecs() error = %v", err)
	}
	if len(skills) != 1 ||
		skills[0].PackageName != "slack-gif-creator" ||
		len(skills[0].Requirements.ToolRefs) != 1 ||
		skills[0].Requirements.ToolRefs[0] != "gif-create" {
		t.Fatalf("custom skill specs = %+v", skills)
	}
}

func TestAgentToolBuildPlatformsForNativeToolsIncludeHostPlatform(t *testing.T) {
	for _, language := range []string{"cpp", "rust"} {
		platforms := agentToolBuildPlatformsForLanguage(language)
		if !customToolBuildPlatformHas(platforms, "linux", "amd64") {
			t.Fatalf("%s platforms missing linux/amd64: %+v", language, platforms)
		}
		if !customToolBuildPlatformHas(platforms, "linux", "arm64") {
			t.Fatalf("%s platforms missing linux/arm64: %+v", language, platforms)
		}
		if host, ok := hostAgentToolBuildPlatform(); ok {
			if !customToolBuildPlatformHas(platforms, runtime.GOOS, runtime.GOARCH) {
				t.Fatalf("%s platforms missing host %s/%s: %+v", language, host.OS, host.Arch, platforms)
			}
		}
	}
}

func TestLocalNativeAgentToolBuildUsesAbsoluteOutputWithRelativeSource(t *testing.T) {
	fakeBin := t.TempDir()
	writeFakeNativeCompiler(t, filepath.Join(fakeBin, "clang++"))
	writeFakeNativeCompiler(t, filepath.Join(fakeBin, "rustc"))
	t.Setenv("PATH", fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"))

	parent := t.TempDir()
	root := filepath.Join(parent, "relative-build")
	src := filepath.Join(root, "src")
	out := filepath.Join(root, "out")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatalf("mkdir source: %v", err)
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatalf("mkdir output: %v", err)
	}
	if err := os.WriteFile(filepath.Join(src, "main.cpp"), []byte("int main(){return 0;}\n"), 0o644); err != nil {
		t.Fatalf("write cpp source: %v", err)
	}
	if err := os.WriteFile(filepath.Join(src, "main.rs"), []byte("fn main() {}\n"), 0o644); err != nil {
		t.Fatalf("write rust source: %v", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get wd: %v", err)
	}
	if err := os.Chdir(parent); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	defer func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("restore wd: %v", err)
		}
	}()

	for _, tc := range []struct {
		name  string
		entry string
		run   func(context.Context, string, string, string) error
	}{
		{name: "cpp", entry: "main.cpp", run: runLocalCPPAgentToolBuild},
		{name: "rust", entry: "main.rs", run: runLocalRustAgentToolBuild},
	} {
		payloadPath := filepath.Join(out, agentToolPayloadName)
		if err := os.Remove(payloadPath); err != nil && !os.IsNotExist(err) {
			t.Fatalf("%s remove previous payload: %v", tc.name, err)
		}
		if err := tc.run(t.Context(), filepath.Join("relative-build", "src"), filepath.Join("relative-build", "out"), tc.entry); err != nil {
			t.Fatalf("%s local build failed: %v", tc.name, err)
		}
		payload, err := os.ReadFile(payloadPath)
		if err != nil {
			t.Fatalf("%s read payload: %v", tc.name, err)
		}
		if string(payload) != "payload\n" {
			t.Fatalf("%s payload = %q", tc.name, string(payload))
		}
	}
}

func writeFakeNativeCompiler(t *testing.T, path string) {
	t.Helper()
	script := `#!/bin/sh
set -eu
out=""
while [ "$#" -gt 0 ]; do
  if [ "$1" = "-o" ]; then
    shift
    out="$1"
    break
  fi
  shift
done
if [ -z "$out" ]; then
  echo "missing -o" >&2
  exit 1
fi
case "$out" in
  /*) ;;
  *)
    echo "output path is not absolute: $out" >&2
    exit 1
    ;;
esac
mkdir -p "$(dirname "$out")"
printf 'payload\n' > "$out"
`
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake compiler: %v", err)
	}
}

func TestBuildAgentPackageFromSourceRejectsPythonRequirementsWithoutVendor(t *testing.T) {
	dir := writeAgentPackageSourceForTest(t)
	toolDir := filepath.Join(dir, "custom", "tools", "python-hello")
	if err := os.MkdirAll(toolDir, 0o755); err != nil {
		t.Fatalf("mkdir custom tool: %v", err)
	}
	if err := os.WriteFile(filepath.Join(toolDir, "tool.py"), []byte("print('hello')\n"), 0o644); err != nil {
		t.Fatalf("write custom tool: %v", err)
	}
	if err := os.WriteFile(filepath.Join(toolDir, "requirements.txt"), []byte("requests==2.32.3\n"), 0o644); err != nil {
		t.Fatalf("write requirements: %v", err)
	}

	_, _, err := buildAgentPackageFromSource(dir, "3.0.4")
	if err == nil || !strings.Contains(err.Error(), "no vendored dependencies under vendor/") {
		t.Fatalf("buildAgentPackageFromSource() error = %v, want vendored dependency error", err)
	}
}

func TestBuildAgentPackageFromSourceRejectsRustCargoPackage(t *testing.T) {
	dir := writeAgentPackageSourceForTest(t)
	toolDir := filepath.Join(dir, "custom", "tools", "rust-hello")
	if err := os.MkdirAll(toolDir, 0o755); err != nil {
		t.Fatalf("mkdir custom tool: %v", err)
	}
	if err := os.WriteFile(filepath.Join(toolDir, "main.rs"), []byte("fn main() {}\n"), 0o644); err != nil {
		t.Fatalf("write custom tool: %v", err)
	}
	if err := os.WriteFile(filepath.Join(toolDir, "Cargo.toml"), []byte("[package]\nname = \"rust-hello\"\nversion = \"0.1.0\"\n"), 0o644); err != nil {
		t.Fatalf("write Cargo.toml: %v", err)
	}

	_, _, err := buildAgentPackageFromSource(dir, "3.0.4")
	if err == nil || !strings.Contains(err.Error(), "Cargo packages are not supported") {
		t.Fatalf("buildAgentPackageFromSource() error = %v, want Cargo package rejection", err)
	}
}

func customToolBuildPlatformHas(platforms []agentToolBuildPlatform, osName, arch string) bool {
	for _, platform := range platforms {
		if platform.OS == osName && platform.Arch == arch {
			return true
		}
	}
	return false
}

func customToolArtifactHasPlatform(manifest agentpackage.CustomToolArtifactManifest, osName, arch string) bool {
	for _, platform := range manifest.Platforms {
		if platform.OS == osName && platform.Arch == arch {
			return true
		}
	}
	return false
}

func writeAgentPackageSourceForTest(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "prompts"), 0o755); err != nil {
		t.Fatalf("mkdir prompts: %v", err)
	}
	manifest := []byte(`schemaVersion: 1
id: com.example.agent
version: 1.0.0
name: Example Agent
model: qwen3.6-plus
minMOIVersion: 3.0.0
instruction:
  path: prompts/role.md
moiResourceRefs:
  tools: []
  skills: []
  knowledgeBases: []
`)
	if err := os.WriteFile(filepath.Join(dir, "agent.yaml"), manifest, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "prompts", "role.md"), []byte("You are concise."), 0o644); err != nil {
		t.Fatalf("write prompt: %v", err)
	}
	return dir
}
