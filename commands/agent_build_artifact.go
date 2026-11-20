package commands

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/matrixorigin/matrixflow/sdk/go-worker-sdk/agentpackage"
)

const (
	agentToolPayloadName          = "payload"
	agentToolArtifactPlatformRoot = "platforms"

	cppAgentToolBuilderImage  = "gcc:14-bookworm"
	rustAgentToolBuilderImage = "rust:bookworm"
)

var agentToolArtifactTimestamp = time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)

type agentToolBuildPlatform struct {
	OS   string
	Arch string
}

type agentToolArtifactFile struct {
	body []byte
	mode int64
}

func buildAgentPackageSourceArtifacts(sourceDir string, files map[string][]byte) (map[string][]byte, error) {
	if sourceDir == "" {
		return nil, fmt.Errorf("source directory is required")
	}
	next := cloneAgentCommandFiles(files)
	specs, err := agentpackage.CustomToolSourceSpecs(next)
	if err != nil {
		return nil, err
	}
	for _, spec := range specs {
		if _, exists := next[agentpackage.CustomToolArtifactPath(spec.PackageDir)]; exists {
			return nil, fmt.Errorf("custom tool %s contains generated artifact file %s; remove it from source and rebuild", spec.PackageDir, agentpackage.CustomToolArtifactPath(spec.PackageDir))
		}
		if _, exists := next[agentpackage.CustomToolArtifactManifestPath(spec.PackageDir)]; exists {
			return nil, fmt.Errorf("custom tool %s contains generated artifact manifest %s; remove it from source and rebuild", spec.PackageDir, agentpackage.CustomToolArtifactManifestPath(spec.PackageDir))
		}
		artifact, manifest, err := buildOneAgentToolArtifact(sourceDir, next, spec)
		if err != nil {
			return nil, err
		}
		manifestBody, err := json.Marshal(manifest)
		if err != nil {
			return nil, fmt.Errorf("marshal custom tool artifact manifest %s: %w", spec.PackageDir, err)
		}
		next[agentpackage.CustomToolArtifactPath(spec.PackageDir)] = artifact
		next[agentpackage.CustomToolArtifactManifestPath(spec.PackageDir)] = append(manifestBody, '\n')
	}
	return next, nil
}

func buildOneAgentToolArtifact(sourceDir string, files map[string][]byte, spec agentpackage.CustomToolSpec) ([]byte, agentpackage.CustomToolArtifactManifest, error) {
	sourceDigest, err := digestAgentToolSource(files, spec.PackageDir)
	if err != nil {
		return nil, agentpackage.CustomToolArtifactManifest{}, err
	}
	platforms := agentToolBuildPlatformsForLanguage(spec.Language)
	artifactFiles := map[string]agentToolArtifactFile{}
	if err := addAgentToolSourceFiles(artifactFiles, files, spec.PackageDir); err != nil {
		return nil, agentpackage.CustomToolArtifactManifest{}, err
	}
	platformManifests := make([]agentpackage.CustomToolArtifactPlatform, 0, len(platforms))
	for _, platform := range platforms {
		platformKey := platform.OS + "_" + platform.Arch
		commandPath := agentToolArtifactPlatformRoot + "/" + platformKey + "/" + agentpackage.CustomToolExecutableName
		switch spec.Language {
		case "go":
			binaries, err := buildGoAgentToolPlatform(sourceDir, spec, platform)
			if err != nil {
				return nil, agentpackage.CustomToolArtifactManifest{}, err
			}
			for name, body := range binaries {
				artifactFiles[agentToolArtifactPlatformRoot+"/"+platformKey+"/"+name] = agentToolArtifactFile{body: body, mode: 0o755}
			}
		case "python":
			if err := validatePythonAgentToolDependencies(files, spec); err != nil {
				return nil, agentpackage.CustomToolArtifactManifest{}, err
			}
			wrapper, err := pythonAgentToolWrapper(spec)
			if err != nil {
				return nil, agentpackage.CustomToolArtifactManifest{}, err
			}
			artifactFiles[commandPath] = agentToolArtifactFile{body: []byte(wrapper), mode: 0o755}
		case "cpp", "rust":
			payload, err := buildNativeAgentToolPlatform(sourceDir, spec, platform)
			if err != nil {
				return nil, agentpackage.CustomToolArtifactManifest{}, err
			}
			wrapper, err := compiledAgentToolPythonWrapper(spec)
			if err != nil {
				return nil, agentpackage.CustomToolArtifactManifest{}, err
			}
			artifactFiles[agentToolArtifactPlatformRoot+"/"+platformKey+"/"+agentToolPayloadName] = agentToolArtifactFile{body: payload, mode: 0o755}
			artifactFiles[commandPath] = agentToolArtifactFile{body: []byte(wrapper), mode: 0o755}
		default:
			return nil, agentpackage.CustomToolArtifactManifest{}, fmt.Errorf("custom tool %s language %q is not supported by agent build", spec.PackageDir, spec.Language)
		}
		platformManifests = append(platformManifests, agentpackage.CustomToolArtifactPlatform{
			OS:      platform.OS,
			Arch:    platform.Arch,
			Command: []string{commandPath},
		})
	}
	artifact, err := buildAgentToolTar(artifactFiles)
	if err != nil {
		return nil, agentpackage.CustomToolArtifactManifest{}, err
	}
	sum := sha256.Sum256(artifact)
	manifest := agentpackage.CustomToolArtifactManifest{
		SchemaVersion:  1,
		ArtifactFormat: "tar",
		Command:        []string{agentpackage.CustomToolExecutableName},
		Platforms:      platformManifests,
		Language:       spec.Language,
		EntryFile:      spec.EntryFile,
		SourceDigest:   sourceDigest,
		ArtifactDigest: "sha256:" + hex.EncodeToString(sum[:]),
	}
	return artifact, manifest, nil
}

func validatePythonAgentToolDependencies(files map[string][]byte, spec agentpackage.CustomToolSpec) error {
	if _, hasRequirements := files[spec.PackageDir+"/requirements.txt"]; !hasRequirements {
		return nil
	}
	vendorPrefix := spec.PackageDir + "/vendor/"
	for name := range files {
		if strings.HasPrefix(name, vendorPrefix) {
			return nil
		}
	}
	return fmt.Errorf("custom Python tool %s contains requirements.txt but no vendored dependencies under vendor/; build must not install dependencies in the target MOI environment", spec.PackageDir)
}

func linuxAgentToolBuildPlatforms() []agentToolBuildPlatform {
	return []agentToolBuildPlatform{
		{OS: "linux", Arch: "amd64"},
		{OS: "linux", Arch: "arm64"},
	}
}

func agentToolBuildPlatformsForLanguage(language string) []agentToolBuildPlatform {
	platforms := linuxAgentToolBuildPlatforms()
	switch language {
	case "go", "python":
		platforms = append(platforms,
			agentToolBuildPlatform{OS: "darwin", Arch: "amd64"},
			agentToolBuildPlatform{OS: "darwin", Arch: "arm64"},
		)
	case "cpp", "rust":
		if host, ok := hostAgentToolBuildPlatform(); ok && !agentToolBuildPlatformContains(platforms, host) {
			platforms = append(platforms, host)
		}
	}
	return platforms
}

func hostAgentToolBuildPlatform() (agentToolBuildPlatform, bool) {
	if runtime.GOOS == "darwin" && (runtime.GOARCH == "amd64" || runtime.GOARCH == "arm64") {
		return agentToolBuildPlatform{OS: runtime.GOOS, Arch: runtime.GOARCH}, true
	}
	return agentToolBuildPlatform{}, false
}

func agentToolBuildPlatformContains(platforms []agentToolBuildPlatform, target agentToolBuildPlatform) bool {
	for _, platform := range platforms {
		if platform.OS == target.OS && platform.Arch == target.Arch {
			return true
		}
	}
	return false
}

func addAgentToolSourceFiles(out map[string]agentToolArtifactFile, files map[string][]byte, packageDir string) error {
	prefix := packageDir + "/"
	for name, body := range files {
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		relative := strings.TrimPrefix(name, prefix)
		if relative == "" ||
			relative == agentpackage.CustomToolArtifactFile ||
			relative == agentpackage.CustomToolArtifactManifestFile ||
			relative == agentpackage.CustomToolExecutableName {
			continue
		}
		if strings.HasPrefix(relative, agentToolArtifactPlatformRoot+"/") {
			return fmt.Errorf("custom tool %s source must not contain reserved directory %s", packageDir, agentToolArtifactPlatformRoot)
		}
		out[relative] = agentToolArtifactFile{body: bytes.Clone(body), mode: 0o644}
	}
	return nil
}

func digestAgentToolSource(files map[string][]byte, packageDir string) (string, error) {
	prefix := packageDir + "/"
	names := make([]string, 0)
	for name := range files {
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		relative := strings.TrimPrefix(name, prefix)
		if relative == agentpackage.CustomToolArtifactFile || relative == agentpackage.CustomToolArtifactManifestFile {
			continue
		}
		names = append(names, name)
	}
	slices.Sort(names)
	hash := sha256.New()
	for _, name := range names {
		if _, err := hash.Write([]byte(name)); err != nil {
			return "", err
		}
		if _, err := hash.Write([]byte{0}); err != nil {
			return "", err
		}
		if _, err := hash.Write(files[name]); err != nil {
			return "", err
		}
		if _, err := hash.Write([]byte{0}); err != nil {
			return "", err
		}
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), nil
}

func buildGoAgentToolPlatform(sourceDir string, spec agentpackage.CustomToolSpec, platform agentToolBuildPlatform) (map[string][]byte, error) {
	tmp, err := os.MkdirTemp("", "moi-agent-tool-build-*")
	if err != nil {
		return nil, fmt.Errorf("create custom tool build workspace: %w", err)
	}
	defer os.RemoveAll(tmp)
	src := filepath.Join(tmp, "src")
	if err := copyDir(filepath.Join(sourceDir, filepath.FromSlash(spec.PackageDir)), src); err != nil {
		return nil, fmt.Errorf("copy custom tool source %s: %w", spec.PackageDir, err)
	}
	payloadPath := filepath.Join(tmp, agentToolPayloadName)
	if err := runGoBuild(context.Background(), src, payloadPath, platform, false); err != nil {
		return nil, fmt.Errorf("build Go custom tool %s for %s/%s: %w", spec.PackageDir, platform.OS, platform.Arch, err)
	}
	wrapperSource := filepath.Join(tmp, "moi_agent_tool_wrapper.go")
	if err := os.WriteFile(wrapperSource, []byte(goAgentToolWrapperSource(agentToolPayloadName, spec.Language, spec.EntryFile)), 0o644); err != nil {
		return nil, fmt.Errorf("write Go custom tool wrapper source: %w", err)
	}
	wrapperPath := filepath.Join(tmp, agentpackage.CustomToolExecutableName)
	if err := runGoBuild(context.Background(), tmp, wrapperPath, platform, true); err != nil {
		return nil, fmt.Errorf("build Go custom tool wrapper for %s/%s: %w", platform.OS, platform.Arch, err)
	}
	payload, err := os.ReadFile(payloadPath)
	if err != nil {
		return nil, fmt.Errorf("read Go custom tool payload: %w", err)
	}
	wrapper, err := os.ReadFile(wrapperPath)
	if err != nil {
		return nil, fmt.Errorf("read Go custom tool wrapper: %w", err)
	}
	return map[string][]byte{
		agentToolPayloadName:                  payload,
		agentpackage.CustomToolExecutableName: wrapper,
	}, nil
}

func buildNativeAgentToolPlatform(sourceDir string, spec agentpackage.CustomToolSpec, platform agentToolBuildPlatform) ([]byte, error) {
	tmp, err := os.MkdirTemp(filepath.Dir(sourceDir), ".moi-agent-tool-build-*")
	if err != nil {
		return nil, fmt.Errorf("create custom tool build workspace: %w", err)
	}
	defer os.RemoveAll(tmp)
	src := filepath.Join(tmp, "src")
	if err := copyDir(filepath.Join(sourceDir, filepath.FromSlash(spec.PackageDir)), src); err != nil {
		return nil, fmt.Errorf("copy custom tool source %s: %w", spec.PackageDir, err)
	}
	out := filepath.Join(tmp, "out")
	if err := os.MkdirAll(out, 0o755); err != nil {
		return nil, fmt.Errorf("create custom tool output directory: %w", err)
	}
	switch spec.Language {
	case "cpp":
		if err := buildCPPAgentToolPlatform(context.Background(), src, out, platform, spec.EntryFile); err != nil {
			return nil, fmt.Errorf("build C++ custom tool %s for %s/%s: %w", spec.PackageDir, platform.OS, platform.Arch, err)
		}
	case "rust":
		files, err := filesInDir(src)
		if err != nil {
			return nil, fmt.Errorf("list Rust custom tool source files: %w", err)
		}
		if err := validateRustAgentToolSource(files); err != nil {
			return nil, fmt.Errorf("validate Rust custom tool %s: %w", spec.PackageDir, err)
		}
		if err := buildRustAgentToolPlatform(context.Background(), src, out, platform, spec.EntryFile); err != nil {
			return nil, fmt.Errorf("build Rust custom tool %s for %s/%s: %w", spec.PackageDir, platform.OS, platform.Arch, err)
		}
	default:
		return nil, fmt.Errorf("language %q is not a native compiled custom tool", spec.Language)
	}
	payload, err := os.ReadFile(filepath.Join(out, agentToolPayloadName))
	if err != nil {
		return nil, fmt.Errorf("read native custom tool payload: %w", err)
	}
	if len(payload) == 0 {
		return nil, fmt.Errorf("native custom tool payload is empty")
	}
	return payload, nil
}

func buildCPPAgentToolPlatform(ctx context.Context, src, out string, platform agentToolBuildPlatform, entryFile string) error {
	if platform.OS == "linux" {
		return runDockerAgentToolBuild(ctx, src, out, platform, cppAgentToolBuilderImage, cppAgentToolBuildScript(entryFile))
	}
	if platform.OS == runtime.GOOS && platform.Arch == runtime.GOARCH {
		return runLocalCPPAgentToolBuild(ctx, src, out, entryFile)
	}
	return fmt.Errorf("unsupported C++ custom tool build platform %s/%s", platform.OS, platform.Arch)
}

func buildRustAgentToolPlatform(ctx context.Context, src, out string, platform agentToolBuildPlatform, entryFile string) error {
	if platform.OS == "linux" {
		return runDockerAgentToolBuild(ctx, src, out, platform, rustAgentToolBuilderImage, rustAgentToolBuildScript(entryFile))
	}
	if platform.OS == runtime.GOOS && platform.Arch == runtime.GOARCH {
		return runLocalRustAgentToolBuild(ctx, src, out, entryFile)
	}
	return fmt.Errorf("unsupported Rust custom tool build platform %s/%s", platform.OS, platform.Arch)
}

func runDockerAgentToolBuild(ctx context.Context, src, out string, platform agentToolBuildPlatform, image, script string) error {
	srcAbs, err := filepath.Abs(src)
	if err != nil {
		return err
	}
	outAbs, err := filepath.Abs(out)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "docker", "run", "--rm",
		"--platform=linux/"+platform.Arch,
		"-v", srcAbs+":/src:ro",
		"-v", outAbs+":/out",
		"-w", "/src",
		image,
		"bash", "-lc", script,
	)
	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, strings.TrimSpace(string(outputBytes)))
	}
	return nil
}

func runLocalCPPAgentToolBuild(ctx context.Context, src, out, entryFile string) error {
	const compiler = "clang++"
	if _, err := exec.LookPath(compiler); err != nil {
		return fmt.Errorf("%s is required to build C++ custom tools for %s/%s: %w", compiler, runtime.GOOS, runtime.GOARCH, err)
	}
	outAbs, err := filepath.Abs(out)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, compiler,
		"-std=c++17",
		"-O2",
		"-o", filepath.Join(outAbs, agentToolPayloadName),
		entryFile,
	)
	cmd.Dir = src
	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, strings.TrimSpace(string(outputBytes)))
	}
	return nil
}

func runLocalRustAgentToolBuild(ctx context.Context, src, out, entryFile string) error {
	if _, err := exec.LookPath("rustc"); err != nil {
		return fmt.Errorf("rustc is required to build Rust custom tools for %s/%s: %w", runtime.GOOS, runtime.GOARCH, err)
	}
	outAbs, err := filepath.Abs(out)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "rustc",
		"--edition=2021",
		"-O",
		"-C", "strip=symbols",
		"-o", filepath.Join(outAbs, agentToolPayloadName),
		entryFile,
	)
	cmd.Dir = src
	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, strings.TrimSpace(string(outputBytes)))
	}
	return nil
}

func cppAgentToolBuildScript(entryFile string) string {
	return fmt.Sprintf("g++ -std=c++17 -O2 -pipe -static -static-libstdc++ -static-libgcc -s -Wl,--build-id=none -o /out/%s %s", shellQuote(agentToolPayloadName), shellQuote(entryFile))
}

func rustAgentToolBuildScript(entryFile string) string {
	return fmt.Sprintf("export PATH=/usr/local/cargo/bin:$PATH; rustc --edition=2021 -O -C target-feature=+crt-static -C strip=symbols -C link-arg=-Wl,--build-id=none -o /out/%s %s", shellQuote(agentToolPayloadName), shellQuote(entryFile))
}

func validateRustAgentToolSource(files []string) error {
	for _, name := range files {
		if name == "Cargo.toml" || name == "Cargo.lock" {
			return fmt.Errorf("Cargo packages are not supported by v1 custom tool build; provide a self-contained main.rs with local modules")
		}
	}
	return nil
}

func filesInDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	slices.Sort(names)
	return names, nil
}

func runGoBuild(ctx context.Context, dir, output string, platform agentToolBuildPlatform, singleFile bool) error {
	args := []string{"build", "-trimpath", "-ldflags=-s -w", "-o", output}
	if singleFile {
		args = append(args, filepath.Join(dir, "moi_agent_tool_wrapper.go"))
	} else if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
		args = append(args, ".")
	} else {
		goFiles, err := rootGoFiles(dir)
		if err != nil {
			return err
		}
		args = append(args, goFiles...)
	}
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"CGO_ENABLED=0",
		"GOOS="+platform.OS,
		"GOARCH="+platform.Arch,
	)
	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err, strings.TrimSpace(string(outputBytes)))
	}
	return nil
}

func rootGoFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	files := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".go") {
			files = append(files, name)
		}
	}
	slices.Sort(files)
	if len(files) == 0 {
		return nil, fmt.Errorf("no Go files found")
	}
	return files, nil
}

func buildAgentToolTar(files map[string]agentToolArtifactFile) ([]byte, error) {
	names := slices.Sorted(maps.Keys(files))
	var buf bytes.Buffer
	writer := tar.NewWriter(&buf)
	for _, name := range names {
		file := files[name]
		if err := writer.WriteHeader(&tar.Header{
			Name:    name,
			Mode:    file.mode,
			Size:    int64(len(file.body)),
			ModTime: agentToolArtifactTimestamp,
			Format:  tar.FormatUSTAR,
		}); err != nil {
			_ = writer.Close()
			return nil, fmt.Errorf("write custom tool artifact header %s: %w", name, err)
		}
		if _, err := writer.Write(file.body); err != nil {
			_ = writer.Close()
			return nil, fmt.Errorf("write custom tool artifact file %s: %w", name, err)
		}
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close custom tool artifact: %w", err)
	}
	return buf.Bytes(), nil
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(dst, 0o755)
		}
		target := filepath.Join(dst, rel)
		if entry.IsDir() {
			if rel == agentToolArtifactPlatformRoot {
				return fmt.Errorf("source contains reserved directory %s", agentToolArtifactPlatformRoot)
			}
			return os.MkdirAll(target, 0o755)
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, body, 0o644)
	})
}

func pythonAgentToolWrapper(spec agentpackage.CustomToolSpec) (string, error) {
	entryJSON, err := json.Marshal(spec.EntryFile)
	if err != nil {
		return "", err
	}
	languageJSON, err := json.Marshal(spec.Language)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`#!/usr/bin/env python3
import json
import os
import subprocess
import sys

entry_file = %s
language = %s
wrapper_dir = os.path.dirname(os.path.abspath(__file__))
artifact_root = os.path.abspath(os.path.join(wrapper_dir, "..", ".."))
env = os.environ.copy()
vendor_dir = os.path.join(artifact_root, "vendor")
if os.path.isdir(vendor_dir):
    env["PYTHONPATH"] = vendor_dir + (os.pathsep + env["PYTHONPATH"] if env.get("PYTHONPATH") else "")
body = sys.stdin.buffer.read()
process = subprocess.run(["python3", entry_file], input=body, stdout=subprocess.PIPE, stderr=subprocess.PIPE, cwd=artifact_root, env=env)
try:
    stdout = process.stdout.decode("utf-8")
    stderr = process.stderr.decode("utf-8")
except UnicodeDecodeError as exc:
    sys.stderr.write("custom tool output must be valid UTF-8: " + str(exc) + "\n")
    sys.exit(1)
if process.returncode != 0:
    sys.stderr.write(stderr)
    sys.exit(process.returncode)
json.dump({
    "output": {
        "stdout": stdout,
        "stderr": stderr,
        "exit_code": process.returncode,
    },
    "metadata": {
        "runner": "moi-agent-tool-wrapper",
        "language": language,
        "entry_file": entry_file,
    },
}, sys.stdout, ensure_ascii=False, separators=(",", ":"))
sys.stdout.write("\n")
`, string(entryJSON), string(languageJSON)), nil
}

func compiledAgentToolPythonWrapper(spec agentpackage.CustomToolSpec) (string, error) {
	entryJSON, err := json.Marshal(spec.EntryFile)
	if err != nil {
		return "", err
	}
	languageJSON, err := json.Marshal(spec.Language)
	if err != nil {
		return "", err
	}
	payloadJSON, err := json.Marshal(agentToolPayloadName)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`#!/usr/bin/env python3
import json
import os
import subprocess
import sys

entry_file = %s
language = %s
payload_name = %s
wrapper_dir = os.path.dirname(os.path.abspath(__file__))
artifact_root = os.path.abspath(os.path.join(wrapper_dir, "..", ".."))
body = sys.stdin.buffer.read()
process = subprocess.run([os.path.join(wrapper_dir, payload_name)], input=body, stdout=subprocess.PIPE, stderr=subprocess.PIPE, cwd=artifact_root)
try:
    stdout = process.stdout.decode("utf-8")
    stderr = process.stderr.decode("utf-8")
except UnicodeDecodeError as exc:
    sys.stderr.write("custom tool output must be valid UTF-8: " + str(exc) + "\n")
    sys.exit(1)
if process.returncode != 0:
    sys.stderr.write(stderr)
    sys.exit(process.returncode)
json.dump({
    "output": {
        "stdout": stdout,
        "stderr": stderr,
        "exit_code": process.returncode,
    },
    "metadata": {
        "runner": "moi-agent-tool-wrapper",
        "language": language,
        "entry_file": entry_file,
    },
}, sys.stdout, ensure_ascii=False, separators=(",", ":"))
sys.stdout.write("\n")
`, string(entryJSON), string(languageJSON), string(payloadJSON)), nil
}

func goAgentToolWrapperSource(payloadName, language, entryFile string) string {
	payloadJSON, _ := json.Marshal(payloadName)
	languageJSON, _ := json.Marshal(language)
	entryJSON, _ := json.Marshal(entryFile)
	return fmt.Sprintf(`package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	payloadName := %s
	language := %s
	entryFile := %s
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	wrapperDir := filepath.Dir(exe)
	artifactRoot := filepath.Clean(filepath.Join(wrapperDir, "..", ".."))
	body, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	cmd := exec.Command(filepath.Join(wrapperDir, payloadName))
	cmd.Dir = artifactRoot
	cmd.Stdin = bytes.NewReader(body)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		_, _ = os.Stderr.Write(stderr.Bytes())
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	response := map[string]any{
		"output": map[string]any{
			"stdout": stdout.String(),
			"stderr": stderr.String(),
			"exit_code": 0,
		},
		"metadata": map[string]any{
			"runner": "moi-agent-tool-wrapper",
			"language": language,
			"entry_file": entryFile,
		},
	}
	if err := json.NewEncoder(os.Stdout).Encode(response); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
`, string(payloadJSON), string(languageJSON), string(entryJSON))
}

func cloneAgentCommandFiles(files map[string][]byte) map[string][]byte {
	out := make(map[string][]byte, len(files))
	for name, body := range files {
		out[name] = bytes.Clone(body)
	}
	return out
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}
