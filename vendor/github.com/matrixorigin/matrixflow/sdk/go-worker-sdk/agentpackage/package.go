// Package agentpackage provides local .moiagent archive parsing and building.
package agentpackage

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

const (
	SchemaVersion = 1
	ManifestPath  = "agent.yaml"
	ChecksumPath  = "checksums.txt"

	CustomToolRoot  = "custom/tools"
	CustomSkillRoot = "custom/skills"

	CustomToolArtifactFile         = "moi-agent-tool-artifact.tar"
	CustomToolArtifactManifestFile = "moi-agent-tool-artifact.json"
	CustomToolExecutableName       = "moi-agent-tool"
	CustomSkillManifestFile        = "skill.yaml"

	MaxArchiveBytes          int64 = 128 * 1024 * 1024
	MaxEntryBytes            int64 = 64 * 1024 * 1024
	MaxUncompressedBytes     int64 = 256 * 1024 * 1024
	MaxEntries                     = 4096
	MaxAgentDescriptionBytes       = 4096
)

var (
	ErrInvalidPackage = errors.New("invalid agent package")

	zipTimestamp = time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)
	semverRE     = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)

	customToolNameRE     = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)
	customToolMaxNameLen = 51
	customSkillMaxDirLen = 51

	customToolEntryCandidates = []customToolEntryCandidate{
		{Language: "go", EntryFile: "main.go"},
		{Language: "python", EntryFile: "tool.py"},
		{Language: "cpp", EntryFile: "main.cpp"},
		{Language: "rust", EntryFile: "main.rs"},
	}

	forbiddenYAMLKeys = map[string]struct{}{
		"workspace_id":            {},
		"api_key":                 {},
		"apiKey":                  {},
		"user_token":              {},
		"userToken":               {},
		"endpoint":                {},
		"endpoint_url":            {},
		"endpointURL":             {},
		"url":                     {},
		"endpoint_secret":         {},
		"endpointSecret":          {},
		"credential":              {},
		"credentials":             {},
		"deployment_credential":   {},
		"deploymentCredential":    {},
		"runtime":                 {},
		"runtime_provider":        {},
		"runtimeProvider":         {},
		"runtime_profile":         {},
		"runtimeProfile":          {},
		"profile":                 {},
		"runtime_grant":           {},
		"runtimeGrant":            {},
		"provider_binding":        {},
		"providerBinding":         {},
		"provider_binding_id":     {},
		"providerBindingID":       {},
		"agent_binding":           {},
		"agentBinding":            {},
		"agent_binding_id":        {},
		"agentBindingID":          {},
		"capability_servers":      {},
		"capabilityServers":       {},
		"capability_servers_json": {},
		"capabilityServersJSON":   {},
		"mcp_endpoint":            {},
		"mcpEndpoint":             {},
		"skills_endpoint":         {},
		"skillsEndpoint":          {},
		"model_gateway":           {},
		"modelGateway":            {},
		"model_gateway_id":        {},
		"modelGatewayID":          {},
		"models_endpoint":         {},
		"modelsEndpoint":          {},
		"selected_model":          {},
		"selectedModel":           {},
		"runtime_auth":            {},
		"runtimeAuth":             {},
		"runtime_policy":          {},
		"runtimePolicy":           {},
		"dependencyChecks":        {},
		"dependency_checks":       {},
		"connections":             {},
		"artifacts":               {},
		"conflictPolicy":          {},
		"requiredPermissions":     {},
		"tool_specs":              {},
		"toolSpecs":               {},
		"skill_specs":             {},
		"skillSpecs":              {},
		"custom_tools":            {},
		"customTools":             {},
		"apiVersion":              {},
		"kind":                    {},
	}
)

type Package struct {
	Manifest Manifest
	Files    map[string][]byte
	Digest   string
}

type Manifest struct {
	SchemaVersion   int             `yaml:"schemaVersion"`
	ID              string          `yaml:"id"`
	Version         string          `yaml:"version"`
	Name            string          `yaml:"name"`
	Description     *string         `yaml:"description,omitempty"`
	Model           string          `yaml:"model"`
	ModelConfigRef  string          `yaml:"model_config_ref,omitempty"`
	MinMOIVersion   string          `yaml:"minMOIVersion,omitempty"`
	Instruction     PathRef         `yaml:"instruction"`
	MOIResourceRefs MOIResourceRefs `yaml:"moiResourceRefs"`
}

type MOIResourceRefs struct {
	Tools           []NamedResourceRef  `yaml:"tools"`
	Skills          []NamedResourceRef  `yaml:"skills"`
	KnowledgeBases  []NamedResourceRef  `yaml:"knowledgeBases"`
	CatalogFiles    []CatalogFileRef    `yaml:"catalogFiles"`
	ChannelBindings []ChannelBindingRef `yaml:"channelBindings"`
}

type NamedResourceRef struct {
	Name        string `yaml:"name"`
	WorkspaceID string `yaml:"workspaceID"`
}

// CatalogFileRef identifies one file in a Catalog volume.
type CatalogFileRef struct {
	WorkspaceID string `yaml:"workspaceID" json:"workspace_id"`
	VolumeID    int64  `yaml:"volumeID" json:"volume_id"`
	FileID      string `yaml:"fileID" json:"file_id"`
}

// ChannelBindingRef binds either one selected tool or one selected Skill's
// Channel provider to one Channel instance in the target workspace.
// Credentials remain owned by the Channel catalog.
type ChannelBindingRef struct {
	ToolName         string `yaml:"toolName,omitempty" json:"tool_name,omitempty"`
	ToolWorkspaceID  string `yaml:"toolWorkspaceID,omitempty" json:"tool_workspace_id,omitempty"`
	SkillName        string `yaml:"skillName,omitempty" json:"skill_name,omitempty"`
	SkillWorkspaceID string `yaml:"skillWorkspaceID,omitempty" json:"skill_workspace_id,omitempty"`
	Provider         string `yaml:"provider,omitempty" json:"provider,omitempty"`
	WorkspaceID      string `yaml:"workspaceID" json:"workspace_id"`
	InstanceID       string `yaml:"instanceID" json:"instance_id"`
}

type PathRef struct {
	Path string `yaml:"path"`
}

type customToolEntryCandidate struct {
	Language  string
	EntryFile string
}

type CustomToolSpec struct {
	PackageDir  string
	PackageName string
	Language    string
	EntryFile   string
	Command     []string

	ArtifactPath   string
	ArtifactDigest string
	Platforms      []CustomToolArtifactPlatform
}

type CustomSkillSpec struct {
	PackageDir       string
	PackageName      string
	Name             string
	Description      string
	Category         string
	Tags             []string
	RoutingSummary   CustomSkillRoutingSummary
	Instruction      CustomSkillInstruction
	Requirements     CustomSkillRequirements
	ParametersSchema map[string]any
	OutputContract   map[string]any
}

type CustomSkillIdentity struct {
	PackageDir  string
	PackageName string
}

type CustomSkillRoutingSummary struct {
	Summary  string   `yaml:"summary"`
	Examples []string `yaml:"examples"`
}

type CustomSkillInstruction struct {
	Body            string         `yaml:"body"`
	VariablesSchema map[string]any `yaml:"variablesSchema"`
}

type CustomSkillRequirements struct {
	ToolRefs           []string                              `yaml:"toolRefs"`
	ToolResourceRefs   []CustomSkillToolResourceRef          `yaml:"toolResourceRefs"`
	KnowledgeBaseRoles []CustomSkillKnowledgeBaseRequirement `yaml:"knowledgeBaseRoles"`
	FileInputs         []CustomSkillFileInputRequirement     `yaml:"fileInputs"`
	ModelCapabilities  []string                              `yaml:"modelCapabilities"`
}

type CustomSkillToolResourceRef struct {
	ID          string `yaml:"id"`
	WorkspaceID string `yaml:"workspaceID"`
	Kind        string `yaml:"kind"`
}

type CustomSkillKnowledgeBaseRequirement struct {
	Role        string `yaml:"role"`
	Description string `yaml:"description"`
	Required    bool   `yaml:"required"`
}

type CustomSkillFileInputRequirement struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	MimeTypes   []string `yaml:"mimeTypes"`
	Required    bool     `yaml:"required"`
}

type customSkillManifest struct {
	Name             string                    `yaml:"name"`
	Description      string                    `yaml:"description"`
	Category         string                    `yaml:"category"`
	Tags             []string                  `yaml:"tags"`
	RoutingSummary   CustomSkillRoutingSummary `yaml:"routingSummary"`
	Instruction      CustomSkillInstruction    `yaml:"instruction"`
	Requirements     CustomSkillRequirements   `yaml:"requirements"`
	ParametersSchema map[string]any            `yaml:"parametersSchema"`
	OutputContract   map[string]any            `yaml:"outputContract"`
}

type CustomToolArtifactManifest struct {
	SchemaVersion  int                          `json:"schemaVersion"`
	ArtifactFormat string                       `json:"artifactFormat"`
	Command        []string                     `json:"command"`
	Platforms      []CustomToolArtifactPlatform `json:"platforms"`
	Language       string                       `json:"language"`
	EntryFile      string                       `json:"entryFile"`
	SourceDigest   string                       `json:"sourceDigest"`
	ArtifactDigest string                       `json:"artifactDigest"`
}

type CustomToolArtifactPlatform struct {
	OS      string   `json:"os"`
	Arch    string   `json:"arch"`
	Command []string `json:"command"`
}

func Parse(data []byte, runningMOIVersion string) (*Package, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("%w: package body is empty", ErrInvalidPackage)
	}
	if int64(len(data)) > MaxArchiveBytes {
		return nil, fmt.Errorf("%w: package archive exceeds %d bytes", ErrInvalidPackage, MaxArchiveBytes)
	}
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("%w: read zip: %v", ErrInvalidPackage, err)
	}
	if len(reader.File) == 0 {
		return nil, fmt.Errorf("%w: zip contains no files", ErrInvalidPackage)
	}
	if len(reader.File) > MaxEntries {
		return nil, fmt.Errorf("%w: package contains %d entries, max %d", ErrInvalidPackage, len(reader.File), MaxEntries)
	}
	if reader.File[0].Name != ManifestPath {
		return nil, fmt.Errorf("%w: agent.yaml must be the first entry", ErrInvalidPackage)
	}
	files := make(map[string][]byte, len(reader.File))
	entryNames := make([]string, 0, len(reader.File))
	var advertisedTotal uint64
	var actualTotal int64
	for _, file := range reader.File {
		if err := validatePath(file.Name); err != nil {
			return nil, err
		}
		if file.UncompressedSize64 > uint64(MaxEntryBytes) {
			return nil, fmt.Errorf("%w: package entry %s exceeds %d bytes", ErrInvalidPackage, file.Name, MaxEntryBytes)
		}
		if file.UncompressedSize64 > uint64(MaxUncompressedBytes) || advertisedTotal > uint64(MaxUncompressedBytes)-file.UncompressedSize64 {
			return nil, fmt.Errorf("%w: package uncompressed size exceeds %d bytes", ErrInvalidPackage, MaxUncompressedBytes)
		}
		advertisedTotal += file.UncompressedSize64
		if _, exists := files[file.Name]; exists {
			return nil, fmt.Errorf("%w: duplicate entry %s", ErrInvalidPackage, file.Name)
		}
		body, err := readZipFile(file)
		if err != nil {
			return nil, err
		}
		actualTotal += int64(len(body))
		if actualTotal > MaxUncompressedBytes {
			return nil, fmt.Errorf("%w: package uncompressed size exceeds %d bytes", ErrInvalidPackage, MaxUncompressedBytes)
		}
		files[file.Name] = body
		entryNames = append(entryNames, file.Name)
	}
	if err := validateEntryOrder(entryNames); err != nil {
		return nil, err
	}
	checksums, err := parseChecksums(files[ChecksumPath])
	if err != nil {
		return nil, err
	}
	if err := verifyChecksums(files, checksums); err != nil {
		return nil, err
	}
	var manifestNode yaml.Node
	if err := yaml.Unmarshal(files[ManifestPath], &manifestNode); err != nil {
		return nil, fmt.Errorf("%w: parse agent.yaml: %v", ErrInvalidPackage, err)
	}
	if err := rejectForbiddenManifestKeys(&manifestNode); err != nil {
		return nil, err
	}
	if err := rejectNullCatalogFiles(&manifestNode); err != nil {
		return nil, err
	}
	if err := rejectNullChannelBindings(&manifestNode); err != nil {
		return nil, err
	}
	manifest, err := decodeManifestStrict(files[ManifestPath])
	if err != nil {
		return nil, fmt.Errorf("%w: decode agent.yaml: %v", ErrInvalidPackage, err)
	}
	if err := validateManifest(manifest, runningMOIVersion, files); err != nil {
		return nil, err
	}
	digest := sha256.Sum256(data)
	return &Package{
		Manifest: manifest,
		Files:    cloneFiles(files),
		Digest:   "sha256:" + hex.EncodeToString(digest[:]),
	}, nil
}

func Build(files map[string][]byte, runningMOIVersion string) ([]byte, error) {
	if files == nil {
		return nil, fmt.Errorf("%w: package files are required", ErrInvalidPackage)
	}
	if err := validatePackageSizeLimits(files); err != nil {
		return nil, err
	}
	if _, ok := files[ManifestPath]; !ok {
		return nil, fmt.Errorf("%w: agent.yaml is required", ErrInvalidPackage)
	}
	if _, ok := files[ChecksumPath]; ok {
		return nil, fmt.Errorf("%w: checksums.txt is generated by package builder", ErrInvalidPackage)
	}
	normalized := cloneFiles(files)
	for name := range normalized {
		if err := validatePath(name); err != nil {
			return nil, err
		}
	}
	normalized[ChecksumPath] = buildChecksumFile(normalized)
	if err := validatePackageSizeLimits(normalized); err != nil {
		return nil, err
	}
	names := make([]string, 0, len(normalized))
	for name := range normalized {
		if name != ManifestPath {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	ordered := append([]string{ManifestPath}, names...)
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	for _, name := range ordered {
		header := &zip.FileHeader{Name: name, Method: zip.Store, Modified: zipTimestamp}
		entry, err := writer.CreateHeader(header)
		if err != nil {
			_ = writer.Close()
			return nil, fmt.Errorf("%w: create zip entry %s: %v", ErrInvalidPackage, name, err)
		}
		if _, err := entry.Write(normalized[name]); err != nil {
			_ = writer.Close()
			return nil, fmt.Errorf("%w: write zip entry %s: %v", ErrInvalidPackage, name, err)
		}
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("%w: close zip: %v", ErrInvalidPackage, err)
	}
	data := buf.Bytes()
	if _, err := Parse(data, runningMOIVersion); err != nil {
		return nil, err
	}
	return data, nil
}

func decodeManifestStrict(body []byte) (Manifest, error) {
	var manifest Manifest
	decoder := yaml.NewDecoder(bytes.NewReader(body))
	decoder.KnownFields(true)
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

func readZipFile(file *zip.File) ([]byte, error) {
	if file.FileInfo().IsDir() {
		return nil, fmt.Errorf("%w: directory entries are not allowed: %s", ErrInvalidPackage, file.Name)
	}
	reader, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("%w: open %s: %v", ErrInvalidPackage, file.Name, err)
	}
	defer reader.Close()
	body, err := io.ReadAll(io.LimitReader(reader, MaxEntryBytes+1))
	if err != nil {
		return nil, fmt.Errorf("%w: read %s: %v", ErrInvalidPackage, file.Name, err)
	}
	if int64(len(body)) > MaxEntryBytes {
		return nil, fmt.Errorf("%w: package entry %s exceeds %d bytes", ErrInvalidPackage, file.Name, MaxEntryBytes)
	}
	return body, nil
}

func validatePackageSizeLimits(files map[string][]byte) error {
	if len(files) > MaxEntries {
		return fmt.Errorf("%w: package contains %d entries, max %d", ErrInvalidPackage, len(files), MaxEntries)
	}
	var total int64
	for name, body := range files {
		if int64(len(body)) > MaxEntryBytes {
			return fmt.Errorf("%w: package entry %s exceeds %d bytes", ErrInvalidPackage, name, MaxEntryBytes)
		}
		total += int64(len(body))
		if total > MaxUncompressedBytes {
			return fmt.Errorf("%w: package uncompressed size exceeds %d bytes", ErrInvalidPackage, MaxUncompressedBytes)
		}
	}
	return nil
}

func validatePath(filePath string) error {
	if filePath == "" {
		return fmt.Errorf("%w: empty file path", ErrInvalidPackage)
	}
	if strings.Contains(filePath, "\\") {
		return fmt.Errorf("%w: file path must use forward slashes: %s", ErrInvalidPackage, filePath)
	}
	if strings.HasPrefix(filePath, "/") || path.IsAbs(filePath) {
		return fmt.Errorf("%w: file path must be relative: %s", ErrInvalidPackage, filePath)
	}
	clean := path.Clean(filePath)
	if clean != filePath || clean == "." || strings.HasPrefix(clean, "../") || clean == ".." || strings.Contains(clean, "/../") {
		return fmt.Errorf("%w: file path must not contain traversal or normalization segments: %s", ErrInvalidPackage, filePath)
	}
	return nil
}

func validateEntryOrder(names []string) error {
	if len(names) == 0 || names[0] != ManifestPath {
		return fmt.Errorf("%w: agent.yaml must be the first entry", ErrInvalidPackage)
	}
	rest := append([]string(nil), names[1:]...)
	sorted := append([]string(nil), rest...)
	sort.Strings(sorted)
	for i := range rest {
		if rest[i] != sorted[i] {
			return fmt.Errorf("%w: package entries after agent.yaml must be lexicographically sorted", ErrInvalidPackage)
		}
	}
	return nil
}

func parseChecksums(body []byte) (map[string]string, error) {
	if body == nil {
		return nil, fmt.Errorf("%w: checksums.txt is required", ErrInvalidPackage)
	}
	checksums := map[string]string{}
	lines := strings.Split(string(body), "\n")
	for idx, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) != 2 {
			return nil, fmt.Errorf("%w: invalid checksums.txt line %d", ErrInvalidPackage, idx+1)
		}
		sum, filePath := parts[0], parts[1]
		if len(sum) != sha256.Size*2 {
			return nil, fmt.Errorf("%w: invalid checksum length for %s", ErrInvalidPackage, filePath)
		}
		if _, err := hex.DecodeString(sum); err != nil {
			return nil, fmt.Errorf("%w: invalid checksum hex for %s", ErrInvalidPackage, filePath)
		}
		if err := validatePath(filePath); err != nil {
			return nil, err
		}
		if filePath == ChecksumPath {
			return nil, fmt.Errorf("%w: checksums.txt must not checksum itself", ErrInvalidPackage)
		}
		if _, exists := checksums[filePath]; exists {
			return nil, fmt.Errorf("%w: duplicate checksum entry %s", ErrInvalidPackage, filePath)
		}
		checksums[filePath] = sum
	}
	if len(checksums) == 0 {
		return nil, fmt.Errorf("%w: checksums.txt has no entries", ErrInvalidPackage)
	}
	return checksums, nil
}

func verifyChecksums(files map[string][]byte, checksums map[string]string) error {
	for name, body := range files {
		if name == ChecksumPath {
			continue
		}
		expected, ok := checksums[name]
		if !ok {
			return fmt.Errorf("%w: missing checksum for %s", ErrInvalidPackage, name)
		}
		actual := sha256.Sum256(body)
		if hex.EncodeToString(actual[:]) != expected {
			return fmt.Errorf("%w: checksum mismatch for %s", ErrInvalidPackage, name)
		}
	}
	for name := range checksums {
		if _, ok := files[name]; !ok {
			return fmt.Errorf("%w: checksum references missing file %s", ErrInvalidPackage, name)
		}
	}
	return nil
}

func buildChecksumFile(files map[string][]byte) []byte {
	names := make([]string, 0, len(files))
	for name := range files {
		if name != ChecksumPath {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	var buf strings.Builder
	for _, name := range names {
		sum := sha256.Sum256(files[name])
		buf.WriteString(hex.EncodeToString(sum[:]))
		buf.WriteString("  ")
		buf.WriteString(name)
		buf.WriteByte('\n')
	}
	return []byte(buf.String())
}

func validateManifest(manifest Manifest, runningMOIVersion string, files map[string][]byte) error {
	if manifest.SchemaVersion != SchemaVersion {
		return fmt.Errorf("%w: unsupported schemaVersion %d", ErrInvalidPackage, manifest.SchemaVersion)
	}
	if manifest.ID == "" {
		return fmt.Errorf("%w: id is required", ErrInvalidPackage)
	}
	if manifest.Version == "" {
		return fmt.Errorf("%w: version is required", ErrInvalidPackage)
	}
	if manifest.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidPackage)
	}
	if manifest.Model == "" {
		return fmt.Errorf("%w: model is required", ErrInvalidPackage)
	}
	if manifest.Description != nil && len(*manifest.Description) > MaxAgentDescriptionBytes {
		return fmt.Errorf("%w: description is too long", ErrInvalidPackage)
	}
	if err := validateModelConfigRef(manifest.ModelConfigRef); err != nil {
		return err
	}
	if manifest.Instruction.Path == "" {
		return fmt.Errorf("%w: instruction.path is required", ErrInvalidPackage)
	}
	if err := validateSemver(manifest.Version, "version"); err != nil {
		return err
	}
	if manifest.MinMOIVersion != "" {
		if err := validateSemver(manifest.MinMOIVersion, "minMOIVersion"); err != nil {
			return err
		}
		if runningMOIVersion == "" {
			return fmt.Errorf("%w: running MOI version is required when minMOIVersion is set", ErrInvalidPackage)
		}
		if err := validateSemver(runningMOIVersion, "running MOI version"); err != nil {
			return err
		}
		if compareSemver(runningMOIVersion, manifest.MinMOIVersion) < 0 {
			return fmt.Errorf("%w: minMOIVersion %s is greater than running MOI version %s", ErrInvalidPackage, manifest.MinMOIVersion, runningMOIVersion)
		}
	}
	for _, ref := range manifest.pathRefs() {
		if err := validatePath(ref.Path); err != nil {
			return err
		}
		if !strings.HasPrefix(ref.Path, "prompts/") {
			return fmt.Errorf("%w: instruction.path must be under prompts/: %s", ErrInvalidPackage, ref.Path)
		}
		if _, ok := files[ref.Path]; !ok {
			return fmt.Errorf("%w: manifest references missing file %s", ErrInvalidPackage, ref.Path)
		}
		if err := validateInstructionFile(ref.Path, files[ref.Path]); err != nil {
			return err
		}
	}
	if err := validatePackageFileSet(files); err != nil {
		return err
	}
	if err := validateNamedRefs("moiResourceRefs.tools", manifest.MOIResourceRefs.Tools, true); err != nil {
		return err
	}
	if err := validateNamedRefs("moiResourceRefs.skills", manifest.MOIResourceRefs.Skills, true); err != nil {
		return err
	}
	if err := validateNamedRefs("moiResourceRefs.knowledgeBases", manifest.MOIResourceRefs.KnowledgeBases, false); err != nil {
		return err
	}
	if err := ValidateCatalogFileRefs(manifest.MOIResourceRefs.CatalogFiles); err != nil {
		return err
	}
	if err := ValidateResourceChannelBindingRefs(manifest.MOIResourceRefs.Tools, manifest.MOIResourceRefs.Skills, manifest.MOIResourceRefs.ChannelBindings); err != nil {
		return err
	}
	if _, err := CustomToolSpecs(files); err != nil {
		return err
	}
	if _, err := CustomSkillSpecs(files); err != nil {
		return err
	}
	return nil
}

func validateInstructionFile(filePath string, body []byte) error {
	if len(body) == 0 {
		return fmt.Errorf("%w: instruction file %s is required", ErrInvalidPackage, filePath)
	}
	if !utf8.Valid(body) {
		return fmt.Errorf("%w: instruction file %s must be valid UTF-8", ErrInvalidPackage, filePath)
	}
	if bytes.IndexByte(body, 0) >= 0 {
		return fmt.Errorf("%w: instruction file %s must not contain NUL bytes", ErrInvalidPackage, filePath)
	}
	return nil
}

func validatePackageFileSet(files map[string][]byte) error {
	for name := range files {
		switch {
		case name == ManifestPath || name == ChecksumPath:
			continue
		case strings.HasPrefix(name, "prompts/"):
			continue
		case strings.HasPrefix(name, CustomToolRoot+"/"):
			if err := validateCustomRootFilePath(name, CustomToolRoot); err != nil {
				return err
			}
			continue
		case strings.HasPrefix(name, CustomSkillRoot+"/"):
			if err := validateCustomRootFilePath(name, CustomSkillRoot); err != nil {
				return err
			}
			continue
		default:
			return fmt.Errorf("%w: package file %s is not under allowed roots: prompts/, custom/tools/, custom/skills/", ErrInvalidPackage, name)
		}
	}
	return nil
}

func validateCustomRootFilePath(name, root string) error {
	rest := strings.TrimPrefix(name, root+"/")
	if slash := strings.IndexByte(rest, '/'); slash <= 0 {
		return fmt.Errorf("%w: package file %s must be under %s/<name>/", ErrInvalidPackage, name, root)
	}
	return nil
}

func CustomToolSpecs(files map[string][]byte) ([]CustomToolSpec, error) {
	dirs := CustomDirs(files, CustomToolRoot)
	specs := make([]CustomToolSpec, 0, len(dirs))
	for _, dir := range dirs {
		spec, err := InspectCustomTool(files, dir)
		if err != nil {
			return nil, err
		}
		specs = append(specs, spec)
	}
	return specs, nil
}

func CustomToolSourceSpecs(files map[string][]byte) ([]CustomToolSpec, error) {
	dirs := CustomDirs(files, CustomToolRoot)
	specs := make([]CustomToolSpec, 0, len(dirs))
	for _, dir := range dirs {
		spec, err := InspectCustomToolSource(files, dir)
		if err != nil {
			return nil, err
		}
		specs = append(specs, spec)
	}
	return specs, nil
}

func CustomSkillSpecs(files map[string][]byte) ([]CustomSkillSpec, error) {
	dirs := CustomDirs(files, CustomSkillRoot)
	specs := make([]CustomSkillSpec, 0, len(dirs))
	for _, dir := range dirs {
		spec, err := InspectCustomSkill(files, dir)
		if err != nil {
			return nil, err
		}
		specs = append(specs, spec)
	}
	return specs, nil
}

func CustomDirs(files map[string][]byte, prefix string) []string {
	seen := map[string]struct{}{}
	for filePath := range files {
		if !strings.HasPrefix(filePath, prefix+"/") {
			continue
		}
		rest := strings.TrimPrefix(filePath, prefix+"/")
		slash := strings.IndexByte(rest, '/')
		if slash <= 0 {
			continue
		}
		seen[prefix+"/"+rest[:slash]] = struct{}{}
	}
	if len(seen) == 0 {
		return nil
	}
	out := make([]string, 0, len(seen))
	for dir := range seen {
		out = append(out, dir)
	}
	sort.Strings(out)
	return out
}

func InspectCustomSkill(files map[string][]byte, dir string) (CustomSkillSpec, error) {
	identity, body, err := inspectCustomSkillIdentityBody(files, dir)
	if err != nil {
		return CustomSkillSpec{}, err
	}
	manifest, err := decodeCustomSkillManifestStrict(body)
	if err != nil {
		return CustomSkillSpec{}, fmt.Errorf("%w: decode custom skill %s %s: %v", ErrInvalidPackage, dir, CustomSkillManifestFile, err)
	}
	spec := CustomSkillSpec{
		PackageDir:       identity.PackageDir,
		PackageName:      identity.PackageName,
		Name:             manifest.Name,
		Description:      manifest.Description,
		Category:         manifest.Category,
		Tags:             append([]string(nil), manifest.Tags...),
		RoutingSummary:   manifest.RoutingSummary,
		Instruction:      manifest.Instruction,
		Requirements:     cloneCustomSkillRequirements(manifest.Requirements),
		ParametersSchema: cloneAnyMap(manifest.ParametersSchema),
		OutputContract:   cloneAnyMap(manifest.OutputContract),
	}
	if err := validateCustomSkillSpec(spec); err != nil {
		return CustomSkillSpec{}, err
	}
	return spec, nil
}

func InspectCustomSkillIdentity(files map[string][]byte, dir string) (CustomSkillIdentity, error) {
	identity, _, err := inspectCustomSkillIdentityBody(files, dir)
	return identity, err
}

func inspectCustomSkillIdentityBody(files map[string][]byte, dir string) (CustomSkillIdentity, []byte, error) {
	if !strings.HasPrefix(dir, CustomSkillRoot+"/") {
		return CustomSkillIdentity{}, nil, fmt.Errorf("%w: custom skill dir must be under %s: %s", ErrInvalidPackage, CustomSkillRoot, dir)
	}
	name := strings.TrimPrefix(dir, CustomSkillRoot+"/")
	if strings.Contains(name, "/") || name == "" {
		return CustomSkillIdentity{}, nil, fmt.Errorf("%w: custom skill dir must have exactly one name segment: %s", ErrInvalidPackage, dir)
	}
	if !customToolNameRE.MatchString(name) {
		return CustomSkillIdentity{}, nil, fmt.Errorf("%w: custom skill dir name must match %s: %s", ErrInvalidPackage, customToolNameRE.String(), name)
	}
	if len(name) > customSkillMaxDirLen {
		return CustomSkillIdentity{}, nil, fmt.Errorf("%w: custom skill dir name is too long: %s", ErrInvalidPackage, name)
	}
	manifestPath := CustomSkillManifestPath(dir)
	body, ok := files[manifestPath]
	if !ok {
		return CustomSkillIdentity{}, nil, fmt.Errorf("%w: custom skill %s must contain %s", ErrInvalidPackage, dir, CustomSkillManifestFile)
	}
	for filePath := range files {
		if strings.HasPrefix(filePath, dir+"/") && filePath != manifestPath {
			return CustomSkillIdentity{}, nil, fmt.Errorf("%w: custom skill %s must contain only %s, found %s", ErrInvalidPackage, dir, CustomSkillManifestFile, filePath)
		}
	}
	if !utf8.Valid(body) {
		return CustomSkillIdentity{}, nil, fmt.Errorf("%w: custom skill %s %s must be valid UTF-8", ErrInvalidPackage, dir, CustomSkillManifestFile)
	}
	return CustomSkillIdentity{PackageDir: dir, PackageName: name}, body, nil
}

func InspectCustomTool(files map[string][]byte, dir string) (CustomToolSpec, error) {
	spec, err := InspectCustomToolSource(files, dir)
	if err != nil {
		return CustomToolSpec{}, err
	}
	artifactPath := CustomToolArtifactPath(dir)
	manifestPath := CustomToolArtifactManifestPath(dir)
	artifact, ok := files[artifactPath]
	if !ok {
		return CustomToolSpec{}, fmt.Errorf("%w: custom tool %s is missing built artifact %s; run core-cli agent build before load", ErrInvalidPackage, dir, artifactPath)
	}
	manifestBody, ok := files[manifestPath]
	if !ok {
		return CustomToolSpec{}, fmt.Errorf("%w: custom tool %s is missing built artifact manifest %s; run core-cli agent build before load", ErrInvalidPackage, dir, manifestPath)
	}
	manifest, err := ParseCustomToolArtifactManifest(manifestBody)
	if err != nil {
		return CustomToolSpec{}, fmt.Errorf("%w: custom tool %s artifact manifest is invalid: %v", ErrInvalidPackage, dir, err)
	}
	if err := validateCustomToolArtifactManifest(spec, manifest, artifact); err != nil {
		return CustomToolSpec{}, err
	}
	spec.ArtifactPath = artifactPath
	spec.ArtifactDigest = manifest.ArtifactDigest
	spec.Command = append([]string(nil), manifest.Command...)
	spec.Platforms = cloneCustomToolArtifactPlatforms(manifest.Platforms)
	return spec, nil
}

func InspectCustomToolSource(files map[string][]byte, dir string) (CustomToolSpec, error) {
	if !strings.HasPrefix(dir, CustomToolRoot+"/") {
		return CustomToolSpec{}, fmt.Errorf("%w: custom tool dir must be under %s: %s", ErrInvalidPackage, CustomToolRoot, dir)
	}
	name := strings.TrimPrefix(dir, CustomToolRoot+"/")
	if strings.Contains(name, "/") || name == "" {
		return CustomToolSpec{}, fmt.Errorf("%w: custom tool dir must have exactly one name segment: %s", ErrInvalidPackage, dir)
	}
	if !customToolNameRE.MatchString(name) {
		return CustomToolSpec{}, fmt.Errorf("%w: custom tool dir name must match %s: %s", ErrInvalidPackage, customToolNameRE.String(), name)
	}
	if len(name) > customToolMaxNameLen {
		return CustomToolSpec{}, fmt.Errorf("%w: custom tool dir name is too long: %s", ErrInvalidPackage, name)
	}
	if _, reserved := files[dir+"/moi-agent-tool-wrapper"]; reserved {
		return CustomToolSpec{}, fmt.Errorf("%w: custom tool %s must not contain reserved file moi-agent-tool-wrapper", ErrInvalidPackage, dir)
	}
	if _, reserved := files[dir+"/"+CustomToolExecutableName]; reserved {
		return CustomToolSpec{}, fmt.Errorf("%w: custom tool %s must not contain reserved file %s", ErrInvalidPackage, dir, CustomToolExecutableName)
	}
	found := make([]customToolEntryCandidate, 0, 1)
	for _, candidate := range customToolEntryCandidates {
		if _, ok := files[dir+"/"+candidate.EntryFile]; ok {
			found = append(found, candidate)
		}
	}
	if len(found) != 1 {
		if len(found) == 0 {
			return CustomToolSpec{}, fmt.Errorf("%w: custom tool %s must contain exactly one supported entry file: %s", ErrInvalidPackage, dir, supportedCustomToolEntryFiles())
		}
		return CustomToolSpec{}, fmt.Errorf("%w: custom tool %s must contain exactly one supported entry file, found %s", ErrInvalidPackage, dir, foundCustomToolEntryFiles(found))
	}
	entry := found[0]
	return CustomToolSpec{
		PackageDir:  dir,
		PackageName: name,
		Language:    entry.Language,
		EntryFile:   entry.EntryFile,
		Command:     []string{CustomToolExecutableName},
	}, nil
}

func supportedCustomToolEntryFiles() string {
	names := make([]string, 0, len(customToolEntryCandidates))
	for _, candidate := range customToolEntryCandidates {
		names = append(names, candidate.EntryFile)
	}
	return strings.Join(names, ", ")
}

func foundCustomToolEntryFiles(found []customToolEntryCandidate) string {
	names := make([]string, 0, len(found))
	for _, candidate := range found {
		names = append(names, candidate.EntryFile)
	}
	return strings.Join(names, ", ")
}

func CustomToolArtifactPath(dir string) string {
	return dir + "/" + CustomToolArtifactFile
}

func CustomToolArtifactManifestPath(dir string) string {
	return dir + "/" + CustomToolArtifactManifestFile
}

func CustomSkillManifestPath(dir string) string {
	return dir + "/" + CustomSkillManifestFile
}

func ParseCustomToolArtifactManifest(body []byte) (CustomToolArtifactManifest, error) {
	var manifest CustomToolArtifactManifest
	if len(body) == 0 {
		return manifest, fmt.Errorf("artifact manifest is empty")
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		return manifest, err
	}
	if decoder.More() {
		return manifest, fmt.Errorf("artifact manifest must contain one JSON object")
	}
	return manifest, nil
}

func decodeCustomSkillManifestStrict(body []byte) (customSkillManifest, error) {
	var manifest customSkillManifest
	decoder := yaml.NewDecoder(bytes.NewReader(body))
	decoder.KnownFields(true)
	if err := decoder.Decode(&manifest); err != nil {
		return customSkillManifest{}, err
	}
	return manifest, nil
}

func validateCustomSkillSpec(spec CustomSkillSpec) error {
	if spec.Name == "" {
		return fmt.Errorf("%w: custom skill %s name is required", ErrInvalidPackage, spec.PackageDir)
	}
	if spec.Name != strings.TrimSpace(spec.Name) {
		return fmt.Errorf("%w: custom skill %s name must not contain leading or trailing whitespace", ErrInvalidPackage, spec.PackageDir)
	}
	if len(spec.Name) > 128 {
		return fmt.Errorf("%w: custom skill %s name is too long", ErrInvalidPackage, spec.PackageDir)
	}
	if spec.Description == "" {
		return fmt.Errorf("%w: custom skill %s description is required", ErrInvalidPackage, spec.PackageDir)
	}
	if spec.Description != strings.TrimSpace(spec.Description) {
		return fmt.Errorf("%w: custom skill %s description must not contain leading or trailing whitespace", ErrInvalidPackage, spec.PackageDir)
	}
	if len(spec.Description) > 4096 {
		return fmt.Errorf("%w: custom skill %s description is too long", ErrInvalidPackage, spec.PackageDir)
	}
	if spec.Category != strings.TrimSpace(spec.Category) {
		return fmt.Errorf("%w: custom skill %s category must not contain leading or trailing whitespace", ErrInvalidPackage, spec.PackageDir)
	}
	if spec.RoutingSummary.Summary != strings.TrimSpace(spec.RoutingSummary.Summary) {
		return fmt.Errorf("%w: custom skill %s routingSummary.summary must not contain leading or trailing whitespace", ErrInvalidPackage, spec.PackageDir)
	}
	if len(spec.RoutingSummary.Summary) > 4096 {
		return fmt.Errorf("%w: custom skill %s routingSummary.summary is too long", ErrInvalidPackage, spec.PackageDir)
	}
	if spec.Instruction.Body != strings.TrimSpace(spec.Instruction.Body) {
		return fmt.Errorf("%w: custom skill %s instruction.body must not contain leading or trailing whitespace", ErrInvalidPackage, spec.PackageDir)
	}
	if len(spec.Instruction.Body) > 64000 {
		return fmt.Errorf("%w: custom skill %s instruction.body is too long", ErrInvalidPackage, spec.PackageDir)
	}
	if spec.Instruction.Body == "" && spec.RoutingSummary.Summary == "" {
		return fmt.Errorf("%w: custom skill %s instruction.body or routingSummary.summary is required", ErrInvalidPackage, spec.PackageDir)
	}
	for index, tag := range spec.Tags {
		if tag == "" {
			return fmt.Errorf("%w: custom skill %s tags[%d] must not be empty", ErrInvalidPackage, spec.PackageDir, index)
		}
		if tag != strings.TrimSpace(tag) {
			return fmt.Errorf("%w: custom skill %s tags[%d] must not contain leading or trailing whitespace", ErrInvalidPackage, spec.PackageDir, index)
		}
	}
	for index, example := range spec.RoutingSummary.Examples {
		if example == "" {
			return fmt.Errorf("%w: custom skill %s routingSummary.examples[%d] must not be empty", ErrInvalidPackage, spec.PackageDir, index)
		}
		if example != strings.TrimSpace(example) {
			return fmt.Errorf("%w: custom skill %s routingSummary.examples[%d] must not contain leading or trailing whitespace", ErrInvalidPackage, spec.PackageDir, index)
		}
	}
	for index, capability := range spec.Requirements.ModelCapabilities {
		if capability == "" {
			return fmt.Errorf("%w: custom skill %s requirements.modelCapabilities[%d] must not be empty", ErrInvalidPackage, spec.PackageDir, index)
		}
		if capability != strings.TrimSpace(capability) {
			return fmt.Errorf("%w: custom skill %s requirements.modelCapabilities[%d] must not contain leading or trailing whitespace", ErrInvalidPackage, spec.PackageDir, index)
		}
	}
	for index, toolRef := range spec.Requirements.ToolRefs {
		if toolRef == "" {
			return fmt.Errorf("%w: custom skill %s requirements.toolRefs[%d] must not be empty", ErrInvalidPackage, spec.PackageDir, index)
		}
		if toolRef != strings.TrimSpace(toolRef) {
			return fmt.Errorf("%w: custom skill %s requirements.toolRefs[%d] must not contain leading or trailing whitespace", ErrInvalidPackage, spec.PackageDir, index)
		}
		if strings.Contains(toolRef, "/") || strings.Contains(toolRef, `\`) {
			return fmt.Errorf("%w: custom skill %s requirements.toolRefs[%d] must not contain path separators", ErrInvalidPackage, spec.PackageDir, index)
		}
	}
	for index, toolRef := range spec.Requirements.ToolResourceRefs {
		if toolRef.ID == "" {
			return fmt.Errorf("%w: custom skill %s requirements.toolResourceRefs[%d].id must not be empty", ErrInvalidPackage, spec.PackageDir, index)
		}
		if toolRef.ID != strings.TrimSpace(toolRef.ID) {
			return fmt.Errorf("%w: custom skill %s requirements.toolResourceRefs[%d].id must not contain leading or trailing whitespace", ErrInvalidPackage, spec.PackageDir, index)
		}
		if strings.Contains(toolRef.ID, "/") || strings.Contains(toolRef.ID, `\`) {
			return fmt.Errorf("%w: custom skill %s requirements.toolResourceRefs[%d].id must not contain path separators", ErrInvalidPackage, spec.PackageDir, index)
		}
		if toolRef.WorkspaceID != "" {
			if toolRef.WorkspaceID != strings.TrimSpace(toolRef.WorkspaceID) {
				return fmt.Errorf("%w: custom skill %s requirements.toolResourceRefs[%d].workspaceID must not contain leading or trailing whitespace", ErrInvalidPackage, spec.PackageDir, index)
			}
			if strings.Contains(toolRef.WorkspaceID, "/") || strings.Contains(toolRef.WorkspaceID, `\`) {
				return fmt.Errorf("%w: custom skill %s requirements.toolResourceRefs[%d].workspaceID must not contain path separators", ErrInvalidPackage, spec.PackageDir, index)
			}
			if toolRef.WorkspaceID != "system" {
				return fmt.Errorf("%w: custom skill %s requirements.toolResourceRefs[%d].workspaceID must be system when present", ErrInvalidPackage, spec.PackageDir, index)
			}
		}
		if toolRef.Kind != "" && toolRef.Kind != "tool" {
			return fmt.Errorf("%w: custom skill %s requirements.toolResourceRefs[%d].kind must be tool", ErrInvalidPackage, spec.PackageDir, index)
		}
	}
	for index, role := range spec.Requirements.KnowledgeBaseRoles {
		if role.Role == "" {
			return fmt.Errorf("%w: custom skill %s requirements.knowledgeBaseRoles[%d].role is required", ErrInvalidPackage, spec.PackageDir, index)
		}
		if role.Role != strings.TrimSpace(role.Role) {
			return fmt.Errorf("%w: custom skill %s requirements.knowledgeBaseRoles[%d].role must not contain leading or trailing whitespace", ErrInvalidPackage, spec.PackageDir, index)
		}
	}
	for index, input := range spec.Requirements.FileInputs {
		if input.Name == "" {
			return fmt.Errorf("%w: custom skill %s requirements.fileInputs[%d].name is required", ErrInvalidPackage, spec.PackageDir, index)
		}
		if input.Name != strings.TrimSpace(input.Name) {
			return fmt.Errorf("%w: custom skill %s requirements.fileInputs[%d].name must not contain leading or trailing whitespace", ErrInvalidPackage, spec.PackageDir, index)
		}
		for mimeIndex, mimeType := range input.MimeTypes {
			if mimeType == "" {
				return fmt.Errorf("%w: custom skill %s requirements.fileInputs[%d].mimeTypes[%d] must not be empty", ErrInvalidPackage, spec.PackageDir, index, mimeIndex)
			}
			if mimeType != strings.TrimSpace(mimeType) {
				return fmt.Errorf("%w: custom skill %s requirements.fileInputs[%d].mimeTypes[%d] must not contain leading or trailing whitespace", ErrInvalidPackage, spec.PackageDir, index, mimeIndex)
			}
		}
	}
	return nil
}

func validateCustomToolArtifactManifest(spec CustomToolSpec, manifest CustomToolArtifactManifest, artifact []byte) error {
	if manifest.SchemaVersion != 1 {
		return fmt.Errorf("%w: custom tool %s artifact manifest schemaVersion must be 1", ErrInvalidPackage, spec.PackageDir)
	}
	if manifest.ArtifactFormat != "tar" {
		return fmt.Errorf("%w: custom tool %s artifactFormat must be tar", ErrInvalidPackage, spec.PackageDir)
	}
	if len(manifest.Command) != 1 || manifest.Command[0] != CustomToolExecutableName {
		return fmt.Errorf("%w: custom tool %s artifact command must be [%q]", ErrInvalidPackage, spec.PackageDir, CustomToolExecutableName)
	}
	if manifest.Language != spec.Language {
		return fmt.Errorf("%w: custom tool %s artifact language %q does not match source language %q", ErrInvalidPackage, spec.PackageDir, manifest.Language, spec.Language)
	}
	if manifest.EntryFile != spec.EntryFile {
		return fmt.Errorf("%w: custom tool %s artifact entryFile %q does not match source entry %q", ErrInvalidPackage, spec.PackageDir, manifest.EntryFile, spec.EntryFile)
	}
	if len(manifest.Platforms) == 0 {
		return fmt.Errorf("%w: custom tool %s artifact manifest must list at least one platform", ErrInvalidPackage, spec.PackageDir)
	}
	seen := map[string]struct{}{}
	for index, platform := range manifest.Platforms {
		if platform.OS == "" || platform.Arch == "" {
			return fmt.Errorf("%w: custom tool %s artifact platforms[%d] must set os and arch", ErrInvalidPackage, spec.PackageDir, index)
		}
		key := platform.OS + "/" + platform.Arch
		if _, exists := seen[key]; exists {
			return fmt.Errorf("%w: custom tool %s artifact platform %s is duplicated", ErrInvalidPackage, spec.PackageDir, key)
		}
		seen[key] = struct{}{}
		if len(platform.Command) != 1 {
			return fmt.Errorf("%w: custom tool %s artifact platform %s command must contain exactly one executable", ErrInvalidPackage, spec.PackageDir, key)
		}
		if err := validateCustomToolArtifactCommandPath(platform.Command[0], key); err != nil {
			return fmt.Errorf("%w: custom tool %s artifact platform %s command is invalid: %v", ErrInvalidPackage, spec.PackageDir, key, err)
		}
	}
	digest := sha256.Sum256(artifact)
	actual := "sha256:" + hex.EncodeToString(digest[:])
	if manifest.ArtifactDigest != actual {
		return fmt.Errorf("%w: custom tool %s artifact digest mismatch: manifest=%s actual=%s", ErrInvalidPackage, spec.PackageDir, manifest.ArtifactDigest, actual)
	}
	if manifest.SourceDigest == "" {
		return fmt.Errorf("%w: custom tool %s artifact manifest sourceDigest is required", ErrInvalidPackage, spec.PackageDir)
	}
	return nil
}

func validateCustomToolArtifactCommandPath(command string, platform string) error {
	if command == "" {
		return fmt.Errorf("command is empty")
	}
	if strings.Contains(command, "\\") {
		return fmt.Errorf("command must use forward slashes")
	}
	if strings.HasPrefix(command, "/") || path.IsAbs(command) {
		return fmt.Errorf("command must be relative")
	}
	clean := path.Clean(command)
	if clean != command || clean == "." || clean == ".." || strings.HasPrefix(clean, "../") || strings.Contains(clean, "/../") {
		return fmt.Errorf("command must not contain traversal or normalization segments")
	}
	wantPrefix := "platforms/" + strings.ReplaceAll(platform, "/", "_") + "/"
	if !strings.HasPrefix(command, wantPrefix) {
		return fmt.Errorf("command must be under %s", wantPrefix)
	}
	return nil
}

func validateNamedRefs(field string, refs []NamedResourceRef, allowSystemWorkspaceID bool) error {
	seen := map[string]struct{}{}
	for index, ref := range refs {
		if ref.Name == "" {
			return fmt.Errorf("%w: %s[%d].name is required", ErrInvalidPackage, field, index)
		}
		if ref.Name != strings.TrimSpace(ref.Name) {
			return fmt.Errorf("%w: %s[%d].name must not contain leading or trailing whitespace", ErrInvalidPackage, field, index)
		}
		if strings.Contains(ref.Name, "/") || strings.Contains(ref.Name, "\\") {
			return fmt.Errorf("%w: %s[%d].name must not contain path separators", ErrInvalidPackage, field, index)
		}
		if ref.WorkspaceID != "" {
			if !allowSystemWorkspaceID {
				return fmt.Errorf("%w: %s[%d].workspaceID is not supported", ErrInvalidPackage, field, index)
			}
			if ref.WorkspaceID != strings.TrimSpace(ref.WorkspaceID) {
				return fmt.Errorf("%w: %s[%d].workspaceID must not contain leading or trailing whitespace", ErrInvalidPackage, field, index)
			}
			if strings.Contains(ref.WorkspaceID, "/") || strings.Contains(ref.WorkspaceID, "\\") {
				return fmt.Errorf("%w: %s[%d].workspaceID must not contain path separators", ErrInvalidPackage, field, index)
			}
			if ref.WorkspaceID != "system" {
				return fmt.Errorf("%w: %s[%d].workspaceID must be system when present", ErrInvalidPackage, field, index)
			}
		}
		key := ref.WorkspaceID + "\x00" + ref.Name
		if _, ok := seen[key]; ok {
			return fmt.Errorf("%w: duplicate %s name %q", ErrInvalidPackage, field, ref.Name)
		}
		seen[key] = struct{}{}
	}
	return nil
}

// ValidateCatalogFileRefs verifies canonical Catalog file identities without
// resolving their current Catalog existence or permissions.
func ValidateCatalogFileRefs(refs []CatalogFileRef) error {
	seen := make(map[CatalogFileRef]struct{}, len(refs))
	for index, ref := range refs {
		if ref.WorkspaceID == "" {
			return fmt.Errorf("%w: moiResourceRefs.catalogFiles[%d].workspaceID is required", ErrInvalidPackage, index)
		}
		if ref.WorkspaceID != strings.TrimSpace(ref.WorkspaceID) {
			return fmt.Errorf("%w: moiResourceRefs.catalogFiles[%d].workspaceID must not contain leading or trailing whitespace", ErrInvalidPackage, index)
		}
		if strings.Contains(ref.WorkspaceID, "/") || strings.Contains(ref.WorkspaceID, "\\") {
			return fmt.Errorf("%w: moiResourceRefs.catalogFiles[%d].workspaceID must not contain path separators", ErrInvalidPackage, index)
		}
		if ref.VolumeID <= 0 {
			return fmt.Errorf("%w: moiResourceRefs.catalogFiles[%d].volumeID must be positive", ErrInvalidPackage, index)
		}
		if ref.FileID == "" {
			return fmt.Errorf("%w: moiResourceRefs.catalogFiles[%d].fileID is required", ErrInvalidPackage, index)
		}
		if ref.FileID != strings.TrimSpace(ref.FileID) {
			return fmt.Errorf("%w: moiResourceRefs.catalogFiles[%d].fileID must not contain leading or trailing whitespace", ErrInvalidPackage, index)
		}
		if strings.ContainsAny(ref.FileID, `/\\`) {
			return fmt.Errorf("%w: moiResourceRefs.catalogFiles[%d].fileID must not contain path separators", ErrInvalidPackage, index)
		}
		if _, exists := seen[ref]; exists {
			return fmt.Errorf("%w: duplicate moiResourceRefs.catalogFiles reference", ErrInvalidPackage)
		}
		seen[ref] = struct{}{}
	}
	return nil
}

// ValidateChannelBindingRefs verifies explicit tool Channel bindings. Use
// ValidateResourceChannelBindingRefs when Skill provider bindings are allowed.
func ValidateChannelBindingRefs(tools []NamedResourceRef, refs []ChannelBindingRef) error {
	return ValidateResourceChannelBindingRefs(tools, nil, refs)
}

// ValidateResourceChannelBindingRefs verifies structural identity and resource
// ownership without resolving live Channel instances or Skill requirements.
func ValidateResourceChannelBindingRefs(tools, skills []NamedResourceRef, refs []ChannelBindingRef) error {
	selectedTools := make(map[string]struct{}, len(tools))
	for _, tool := range tools {
		selectedTools[namedResourceRefKey(tool.Name, tool.WorkspaceID)] = struct{}{}
	}
	selectedSkills := make(map[string]struct{}, len(skills))
	for _, skill := range skills {
		selectedSkills[namedResourceRefKey(skill.Name, skill.WorkspaceID)] = struct{}{}
	}
	seenBindings := make(map[string]struct{}, len(refs))
	for index, ref := range refs {
		field := fmt.Sprintf("moiResourceRefs.channelBindings[%d]", index)
		toolBinding := ref.ToolName != ""
		skillBinding := ref.SkillName != ""
		if toolBinding == skillBinding {
			return fmt.Errorf("%w: %s must bind exactly one tool or Skill provider", ErrInvalidPackage, field)
		}
		if err := validateOptionalChannelBindingIdentifier(field+".toolName", ref.ToolName); err != nil {
			return err
		}
		if err := validateOptionalChannelBindingIdentifier(field+".toolWorkspaceID", ref.ToolWorkspaceID); err != nil {
			return err
		}
		if err := validateOptionalChannelBindingIdentifier(field+".skillName", ref.SkillName); err != nil {
			return err
		}
		if err := validateOptionalChannelBindingIdentifier(field+".skillWorkspaceID", ref.SkillWorkspaceID); err != nil {
			return err
		}
		if err := validateOptionalChannelBindingIdentifier(field+".provider", ref.Provider); err != nil {
			return err
		}
		if err := validateChannelBindingIdentifier(field+".workspaceID", ref.WorkspaceID); err != nil {
			return err
		}
		var bindingKey string
		if toolBinding {
			if ref.SkillWorkspaceID != "" || ref.Provider != "" {
				return fmt.Errorf("%w: %s tool binding must not set Skill provider fields", ErrInvalidPackage, field)
			}
			toolKey := namedResourceRefKey(ref.ToolName, ref.ToolWorkspaceID)
			if _, ok := selectedTools[toolKey]; !ok {
				return fmt.Errorf("%w: %s references a tool not present in moiResourceRefs.tools", ErrInvalidPackage, field)
			}
			bindingKey = "tool\x00" + toolKey
		} else {
			if ref.ToolWorkspaceID != "" {
				return fmt.Errorf("%w: %s Skill provider binding must not set toolWorkspaceID", ErrInvalidPackage, field)
			}
			if ref.Provider == "" {
				return fmt.Errorf("%w: %s.provider is required for a Skill provider binding", ErrInvalidPackage, field)
			}
			skillKey := namedResourceRefKey(ref.SkillName, ref.SkillWorkspaceID)
			if _, ok := selectedSkills[skillKey]; !ok {
				return fmt.Errorf("%w: %s references a Skill not present in moiResourceRefs.skills", ErrInvalidPackage, field)
			}
			bindingKey = "skill\x00" + skillKey + "\x00" + ref.Provider
		}
		if _, exists := seenBindings[bindingKey]; exists {
			return fmt.Errorf("%w: duplicate moiResourceRefs.channelBindings resource reference", ErrInvalidPackage)
		}
		seenBindings[bindingKey] = struct{}{}
		if err := validateChannelBindingIdentifier(field+".instanceID", ref.InstanceID); err != nil {
			return err
		}
	}
	return nil
}

func validateOptionalChannelBindingIdentifier(field, value string) error {
	if value == "" {
		return nil
	}
	return validateChannelBindingIdentifier(field, value)
}

func namedResourceRefKey(name, workspaceID string) string {
	return workspaceID + "\x00" + name
}

func validateChannelBindingIdentifier(field, value string) error {
	if value == "" {
		return fmt.Errorf("%w: %s is required", ErrInvalidPackage, field)
	}
	if value != strings.TrimSpace(value) {
		return fmt.Errorf("%w: %s must not contain leading or trailing whitespace", ErrInvalidPackage, field)
	}
	if strings.ContainsAny(value, `/\\`) {
		return fmt.Errorf("%w: %s must not contain path separators", ErrInvalidPackage, field)
	}
	return nil
}

func validateModelConfigRef(ref string) error {
	if ref == "" {
		return nil
	}
	if ref != strings.TrimSpace(ref) {
		return fmt.Errorf("%w: model_config_ref must not contain leading or trailing whitespace", ErrInvalidPackage)
	}
	if strings.Contains(ref, "/") || strings.Contains(ref, "\\") {
		return fmt.Errorf("%w: model_config_ref must not contain path separators", ErrInvalidPackage)
	}
	return nil
}

func rejectForbiddenManifestKeys(node *yaml.Node) error {
	if node == nil {
		return nil
	}
	if node.Kind == yaml.MappingNode {
		for i := 0; i+1 < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			if keyNode.Kind == yaml.ScalarNode {
				if _, forbidden := forbiddenYAMLKeys[keyNode.Value]; forbidden {
					return fmt.Errorf("%w: agent.yaml must not contain deployment-specific field %q", ErrInvalidPackage, keyNode.Value)
				}
			}
			if err := rejectForbiddenManifestKeys(node.Content[i+1]); err != nil {
				return err
			}
		}
		return nil
	}
	for _, child := range node.Content {
		if err := rejectForbiddenManifestKeys(child); err != nil {
			return err
		}
	}
	return nil
}

func rejectNullCatalogFiles(node *yaml.Node) error {
	if node == nil {
		return nil
	}
	if node.Kind == yaml.MappingNode {
		for index := 0; index+1 < len(node.Content); index += 2 {
			key, value := node.Content[index], node.Content[index+1]
			if key.Value == "catalogFiles" && value.Tag == "!!null" {
				return fmt.Errorf("%w: moiResourceRefs.catalogFiles must not be null", ErrInvalidPackage)
			}
			if err := rejectNullCatalogFiles(value); err != nil {
				return err
			}
		}
		return nil
	}
	for _, child := range node.Content {
		if err := rejectNullCatalogFiles(child); err != nil {
			return err
		}
	}
	return nil
}

func rejectNullChannelBindings(node *yaml.Node) error {
	if node == nil {
		return nil
	}
	if node.Kind == yaml.MappingNode {
		for index := 0; index+1 < len(node.Content); index += 2 {
			key, value := node.Content[index], node.Content[index+1]
			if key.Value == "channelBindings" && value.Tag == "!!null" {
				return fmt.Errorf("%w: moiResourceRefs.channelBindings must not be null", ErrInvalidPackage)
			}
			if err := rejectNullChannelBindings(value); err != nil {
				return err
			}
		}
		return nil
	}
	for _, child := range node.Content {
		if err := rejectNullChannelBindings(child); err != nil {
			return err
		}
	}
	return nil
}

func (manifest Manifest) pathRefs() []PathRef {
	return []PathRef{manifest.Instruction}
}

func validateSemver(version string, field string) error {
	if !semverRE.MatchString(version) {
		return fmt.Errorf("%w: %s must be semantic version, got %q", ErrInvalidPackage, field, version)
	}
	return nil
}

func compareSemver(left string, right string) int {
	leftParts := semverCore(left)
	rightParts := semverCore(right)
	for i := 0; i < 3; i++ {
		if leftParts[i] < rightParts[i] {
			return -1
		}
		if leftParts[i] > rightParts[i] {
			return 1
		}
	}
	return 0
}

func semverCore(version string) [3]int {
	core := strings.SplitN(version, "-", 2)[0]
	parts := strings.Split(core, ".")
	var out [3]int
	for i := 0; i < len(parts) && i < 3; i++ {
		value := 0
		for _, ch := range parts[i] {
			value = value*10 + int(ch-'0')
		}
		out[i] = value
	}
	return out
}

func cloneFiles(files map[string][]byte) map[string][]byte {
	out := make(map[string][]byte, len(files))
	for name, body := range files {
		out[name] = append([]byte(nil), body...)
	}
	return out
}

func cloneAnyMap(input map[string]any) map[string]any {
	if input == nil {
		return nil
	}
	out := make(map[string]any, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func cloneCustomSkillRequirements(input CustomSkillRequirements) CustomSkillRequirements {
	return CustomSkillRequirements{
		ToolRefs:           append([]string(nil), input.ToolRefs...),
		ToolResourceRefs:   append([]CustomSkillToolResourceRef(nil), input.ToolResourceRefs...),
		KnowledgeBaseRoles: cloneCustomSkillKnowledgeBaseRequirements(input.KnowledgeBaseRoles),
		FileInputs:         cloneCustomSkillFileInputRequirements(input.FileInputs),
		ModelCapabilities:  append([]string(nil), input.ModelCapabilities...),
	}
}

func cloneCustomSkillKnowledgeBaseRequirements(input []CustomSkillKnowledgeBaseRequirement) []CustomSkillKnowledgeBaseRequirement {
	out := make([]CustomSkillKnowledgeBaseRequirement, 0, len(input))
	for _, item := range input {
		out = append(out, item)
	}
	return out
}

func cloneCustomSkillFileInputRequirements(input []CustomSkillFileInputRequirement) []CustomSkillFileInputRequirement {
	out := make([]CustomSkillFileInputRequirement, 0, len(input))
	for _, item := range input {
		out = append(out, CustomSkillFileInputRequirement{
			Name:        item.Name,
			Description: item.Description,
			MimeTypes:   append([]string(nil), item.MimeTypes...),
			Required:    item.Required,
		})
	}
	return out
}

func cloneCustomToolArtifactPlatforms(platforms []CustomToolArtifactPlatform) []CustomToolArtifactPlatform {
	out := make([]CustomToolArtifactPlatform, 0, len(platforms))
	for _, platform := range platforms {
		out = append(out, CustomToolArtifactPlatform{
			OS:      platform.OS,
			Arch:    platform.Arch,
			Command: append([]string(nil), platform.Command...),
		})
	}
	return out
}
