package commands

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/matrixorigin/matrixflow/sdk/go-worker-sdk/agentpackage"
	"github.com/matrixorigin/matrixflow/sdk/go-worker-sdk/agentpackageauthor"
)

func runAgentInit(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent init", flag.ExitOnError)
	outputDir := fs.String("output", "", "Output source directory")
	agentID := fs.String("agent-id", "", "Agent ID")
	name := fs.String("name", "Minimal Test Agent", "Agent display name")
	description := fs.String("description", "", "Agent display description")
	version := fs.String("version", "1.0.0", "Agent package semantic version")
	model := fs.String("model", "qwen3.6-plus", "Default model name")
	instruction := fs.String("instruction", "你是一个简洁回答问题的助手。请用一句话回答普通问题。", "Initial role instruction")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: moi-cli agent init --output <dir> --agent-id <id> [options]

Create a minimal Agent package source tree. The command only writes source files;
use agent build, verify, and load for the later package lifecycle steps.

Required:
  --output <dir>          Output source directory; must not already exist
  --agent-id <id>         Agent ID

Optional:
  --name <name>           Agent display name
  --description <text>    Agent display description
  --version <version>     Agent package semantic version
  --model <model>         Default model name
  --instruction <text>    Initial role instruction

Options:
`)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *outputDir == "" {
		return fmt.Errorf("--output is required")
	}
	var descriptionValue *string
	fs.Visit(func(flag *flag.Flag) {
		if flag.Name == "description" {
			descriptionValue = description
		}
	})
	input := agentInitTemplateInput{
		AgentID:     *agentID,
		Name:        *name,
		Description: descriptionValue,
		Version:     *version,
		Model:       *model,
		Instruction: *instruction,
	}
	files, err := buildAgentInitTemplateFiles(input)
	if err != nil {
		return err
	}
	effectiveMOIVersion := resolveAgentPackageRunningMOIVersion(ctx, "")
	if _, err := agentpackage.Build(files, effectiveMOIVersion); err != nil {
		return fmt.Errorf("generated agent template is invalid: %w", err)
	}
	if ctx.PreflightOnly {
		return nil
	}
	if err := writeAgentInitTemplate(*outputDir, files); err != nil {
		return err
	}
	if ctx.OutputJSON {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"agent_id": *agentID,
			"version":  *version,
			"source":   *outputDir,
			"files":    []string{"agent.yaml", "prompts/role.md"},
			"dirs":     []string{"custom/tools", "custom/skills"},
		})
	}
	fmt.Printf("Initialized agent package source: agent_id=%s version=%s source=%s\n", *agentID, *version, *outputDir)
	fmt.Printf("Next: moi-cli agent build --source %s --output %s.moiagent\n", *outputDir, *agentID)
	return nil
}

func runAgentBuild(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent build", flag.ExitOnError)
	sourceDir := fs.String("source", "", "Package source directory")
	output := fs.String("output", "", "Output .moiagent file path")
	runningMOIVersion := fs.String("running-moi-version", "", "Override MOI version for minMOIVersion validation")
	overwrite := fs.Bool("overwrite", false, "Overwrite output file if it already exists")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: moi-cli agent build --source <dir> --output <package.moiagent> [--running-moi-version <version>] [--overwrite]

Build a deterministic .moiagent package from a source directory.

Required:
  --source <dir>          Package source directory containing agent.yaml
  --output <path>         Output package file path

Optional:
  --running-moi-version   Override version used to validate minMOIVersion
  --overwrite             Replace output file if it already exists

Options:
`)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *sourceDir == "" {
		return fmt.Errorf("--source is required")
	}
	if *output == "" {
		return fmt.Errorf("--output is required")
	}
	// Building custom tool artifacts can create temporary files next to the
	// source tree, so dry-run stops after argument validation.
	if ctx.PreflightOnly {
		return nil
	}
	effectiveMOIVersion := resolveAgentPackageRunningMOIVersion(ctx, *runningMOIVersion)
	data, pkg, err := buildAgentPackageFromSource(*sourceDir, effectiveMOIVersion)
	if err != nil {
		return err
	}
	if err := writeAgentPackageOutput(*output, data, *overwrite); err != nil {
		return err
	}
	if ctx.OutputJSON {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"agent_id": pkg.Manifest.ID,
			"version":  pkg.Manifest.Version,
			"digest":   pkg.Digest,
			"output":   *output,
			"bytes":    len(data),
		})
	}
	fmt.Printf("Built agent package: agent_id=%s version=%s output=%s digest=%s\n", pkg.Manifest.ID, pkg.Manifest.Version, *output, pkg.Digest)
	return nil
}

func runAgentVerify(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent verify", flag.ExitOnError)
	filePath := fs.String("file", "", ".moiagent package file")
	sourceDir := fs.String("source", "", "Package source directory containing agent.yaml")
	runningMOIVersion := fs.String("running-moi-version", "", "Override MOI version for minMOIVersion validation")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: moi-cli agent verify (--source <dir> | --file <package.moiagent>) [--running-moi-version <version>]

Verify an Agent package source tree or a built .moiagent package without loading it into a workspace.

Required:
  --source <dir>          Package source directory containing agent.yaml
  --file <path>           Built package file path

Exactly one of --source or --file must be set.

Optional:
  --running-moi-version   Override version used to validate minMOIVersion

Options:
`)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if (*filePath == "") == (*sourceDir == "") {
		return fmt.Errorf("exactly one of --source or --file is required")
	}
	effectiveMOIVersion := resolveAgentPackageRunningMOIVersion(ctx, *runningMOIVersion)
	var pkg *agentpackage.Package
	if *sourceDir != "" {
		_, built, err := buildAgentPackageFromSource(*sourceDir, effectiveMOIVersion)
		if err != nil {
			return err
		}
		pkg = built
	} else {
		data, err := os.ReadFile(*filePath)
		if err != nil {
			return fmt.Errorf("failed to read package file %q: %w", *filePath, err)
		}
		parsed, err := agentpackage.Parse(data, effectiveMOIVersion)
		if err != nil {
			return err
		}
		pkg = parsed
	}
	if ctx.OutputJSON {
		payload := map[string]any{
			"agent_id": pkg.Manifest.ID,
			"version":  pkg.Manifest.Version,
			"digest":   pkg.Digest,
			"files":    len(pkg.Files),
		}
		if *sourceDir != "" {
			payload["source"] = *sourceDir
		} else {
			payload["file"] = *filePath
		}
		return json.NewEncoder(os.Stdout).Encode(payload)
	}
	if *sourceDir != "" {
		fmt.Printf("Verified agent package source: agent_id=%s version=%s digest=%s files=%d\n", pkg.Manifest.ID, pkg.Manifest.Version, pkg.Digest, len(pkg.Files))
		return nil
	}
	fmt.Printf("Verified agent package: agent_id=%s version=%s digest=%s files=%d\n", pkg.Manifest.ID, pkg.Manifest.Version, pkg.Digest, len(pkg.Files))
	return nil
}

func runAgentUnpack(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent unpack", flag.ExitOnError)
	filePath := fs.String("file", "", ".moiagent package file")
	outputDir := fs.String("output", "", "Output source directory")
	runningMOIVersion := fs.String("running-moi-version", "", "Override MOI version for minMOIVersion validation")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: moi-cli agent unpack --file <package.moiagent> --output <dir> [--running-moi-version <version>]

Unpack a built .moiagent package into an editable Agent package source tree.
The command verifies the package and removes package-generated files such as
checksums.txt and custom tool artifacts from the source output.

Required:
  --file <path>           Built package file path
  --output <dir>          Output source directory; must not already exist

Optional:
  --running-moi-version   Override version used to validate minMOIVersion

Options:
`)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *filePath == "" {
		return fmt.Errorf("--file is required")
	}
	if *outputDir == "" {
		return fmt.Errorf("--output is required")
	}
	data, err := os.ReadFile(*filePath)
	if err != nil {
		return fmt.Errorf("failed to read package file %q: %w", *filePath, err)
	}
	effectiveMOIVersion := resolveAgentPackageRunningMOIVersion(ctx, *runningMOIVersion)
	pkg, err := agentpackage.Parse(data, effectiveMOIVersion)
	if err != nil {
		return err
	}
	files := agentPackageEditableSourceFiles(pkg.Files)
	if ctx.PreflightOnly {
		return nil
	}
	if err := writeAgentSourceFiles(*outputDir, files, agentSourceEmptyDirs()); err != nil {
		return err
	}
	if ctx.OutputJSON {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"agent_id": pkg.Manifest.ID,
			"version":  pkg.Manifest.Version,
			"source":   *outputDir,
			"files":    len(files),
		})
	}
	fmt.Printf("Unpacked agent package source: agent_id=%s version=%s source=%s files=%d\n", pkg.Manifest.ID, pkg.Manifest.Version, *outputDir, len(files))
	return nil
}

func collectAgentPackageSourceFiles(sourceDir string) (map[string][]byte, error) {
	info, err := os.Stat(sourceDir)
	if err != nil {
		return nil, fmt.Errorf("failed to stat package source directory %q: %w", sourceDir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("package source %q is not a directory", sourceDir)
	}
	files := map[string][]byte{}
	err = filepath.WalkDir(sourceDir, func(filePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(sourceDir, filePath)
		if err != nil {
			return err
		}
		packagePath := filepath.ToSlash(rel)
		if filepath.Ext(packagePath) == ".moiagent" {
			return nil
		}
		body, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		files[packagePath] = body
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to read package source directory %q: %w", sourceDir, err)
	}
	return files, nil
}

func buildAgentPackageFromSource(sourceDir, runningMOIVersion string) ([]byte, *agentpackage.Package, error) {
	files, err := collectAgentPackageSourceFiles(sourceDir)
	if err != nil {
		return nil, nil, err
	}
	files, err = buildAgentPackageSourceArtifacts(sourceDir, files)
	if err != nil {
		return nil, nil, err
	}
	data, err := agentpackage.Build(files, runningMOIVersion)
	if err != nil {
		return nil, nil, err
	}
	pkg, err := agentpackage.Parse(data, runningMOIVersion)
	if err != nil {
		return nil, nil, err
	}
	return data, pkg, nil
}

func writeAgentPackageOutput(output string, data []byte, overwrite bool) error {
	flags := os.O_WRONLY | os.O_CREATE
	if overwrite {
		flags |= os.O_TRUNC
	} else {
		flags |= os.O_EXCL
	}
	file, err := os.OpenFile(output, flags, 0o644)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("output file %q already exists; use --overwrite to replace it", output)
		}
		return fmt.Errorf("failed to open output file %q: %w", output, err)
	}
	defer file.Close()
	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write output file %q: %w", output, err)
	}
	return nil
}

func resolveAgentPackageRunningMOIVersion(ctx *Context, override string) string {
	if override != "" {
		return override
	}
	if ctx == nil {
		return ""
	}
	return ctx.RunningMOIVersion
}

type agentInitTemplateInput struct {
	AgentID     string
	Name        string
	Description *string
	Version     string
	Model       string
	Instruction string
}

func buildAgentInitTemplateFiles(input agentInitTemplateInput) (map[string][]byte, error) {
	if input.AgentID == "" {
		return nil, fmt.Errorf("--agent-id is required")
	}
	if input.Name == "" {
		return nil, fmt.Errorf("--name must not be empty")
	}
	if input.Version == "" {
		return nil, fmt.Errorf("--version must not be empty")
	}
	if input.Model == "" {
		return nil, fmt.Errorf("--model must not be empty")
	}
	if input.Instruction == "" {
		return nil, fmt.Errorf("--instruction must not be empty")
	}
	return agentpackageauthor.CreateInitialSource(agentpackageauthor.InitialSourceInput{
		AgentID:     input.AgentID,
		Name:        input.Name,
		Description: input.Description,
		Version:     input.Version,
		Model:       input.Model,
		Instruction: input.Instruction,
	})
}

func writeAgentInitTemplate(outputDir string, files map[string][]byte) error {
	for _, path := range []string{"agent.yaml", "prompts/role.md"} {
		if _, exists := files[path]; !exists {
			return fmt.Errorf("agent init template missing %s", path)
		}
	}
	return writeAgentSourceFiles(outputDir, files, agentSourceEmptyDirs())
}

func writeAgentSourceFiles(outputDir string, files map[string][]byte, extraDirs []string) error {
	if _, err := os.Stat(outputDir); err == nil {
		return fmt.Errorf("output directory %q already exists", outputDir)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to stat output directory %q: %w", outputDir, err)
	}
	names := slices.Sorted(maps.Keys(files))
	for _, name := range names {
		if err := os.MkdirAll(filepath.Join(outputDir, filepath.Dir(filepath.FromSlash(name))), 0o755); err != nil {
			return fmt.Errorf("failed to create output directory %q: %w", outputDir, err)
		}
	}
	for _, dir := range extraDirs {
		if err := os.MkdirAll(filepath.Join(outputDir, filepath.FromSlash(dir)), 0o755); err != nil {
			return fmt.Errorf("failed to create output directory %q: %w", dir, err)
		}
	}
	for _, name := range names {
		if err := os.WriteFile(filepath.Join(outputDir, filepath.FromSlash(name)), files[name], 0o644); err != nil {
			return fmt.Errorf("failed to write %s: %w", name, err)
		}
	}
	return nil
}

func agentSourceEmptyDirs() []string {
	return []string{"custom/tools", "custom/skills"}
}

func agentPackageEditableSourceFiles(files map[string][]byte) map[string][]byte {
	out := make(map[string][]byte, len(files))
	for name, body := range files {
		if name == agentpackage.ChecksumPath || isGeneratedCustomToolArtifactFile(name) {
			continue
		}
		out[name] = bytes.Clone(body)
	}
	return out
}

func isGeneratedCustomToolArtifactFile(name string) bool {
	return strings.HasPrefix(name, agentpackage.CustomToolRoot+"/") &&
		(strings.HasSuffix(name, "/"+agentpackage.CustomToolArtifactFile) ||
			strings.HasSuffix(name, "/"+agentpackage.CustomToolArtifactManifestFile))
}
