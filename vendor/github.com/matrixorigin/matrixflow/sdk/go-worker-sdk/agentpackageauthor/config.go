// Package agentpackageauthor converts structured Agent authoring input into
// Agent package source files.
package agentpackageauthor

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/matrixorigin/matrixflow/sdk/go-worker-sdk/agentpackage"
)

var ErrInvalidAuthoringInput = errors.New("invalid agent package authoring input")

const rolePromptPath = "prompts/role.md"

type CandidateConfig struct {
	AgentID            string                           `json:"agent_id"`
	Name               string                           `json:"name"`
	Description        string                           `json:"description"`
	ModelName          string                           `json:"model_name"`
	ModelConfigRef     string                           `json:"model_config_ref,omitempty"`
	ToolNames          []string                         `json:"tool_names"`
	SkillNames         []string                         `json:"skill_names"`
	KnowledgeBaseNames []string                         `json:"knowledge_base_names"`
	CatalogFiles       []agentpackage.CatalogFileRef    `json:"catalog_files"`
	ChannelBindings    []agentpackage.ChannelBindingRef `json:"channel_bindings"`
	AgentMD            string                           `json:"agent_md"`
	ChangeReason       string                           `json:"change_reason"`
}

type EditableConfig struct {
	AgentID            string                           `json:"agent_id"`
	Name               string                           `json:"name"`
	Description        string                           `json:"description"`
	ModelName          string                           `json:"model_name"`
	ModelConfigRef     string                           `json:"model_config_ref,omitempty"`
	ToolNames          []string                         `json:"tool_names"`
	SkillNames         []string                         `json:"skill_names"`
	KnowledgeBaseNames []string                         `json:"knowledge_base_names"`
	CatalogFiles       []agentpackage.CatalogFileRef    `json:"catalog_files"`
	ChannelBindings    []agentpackage.ChannelBindingRef `json:"channel_bindings"`
	AgentMD            string                           `json:"agent_md"`
}

type RenderInput struct {
	Config        CandidateConfig
	LoadVersion   string
	ToolRefs      []agentpackage.NamedResourceRef
	SkillRefs     []agentpackage.NamedResourceRef
	KnowledgeRefs []agentpackage.NamedResourceRef
}

type InitialSourceInput struct {
	AgentID     string
	Name        string
	Description *string
	Version     string
	Model       string
	Instruction string
}

func ParseCandidateConfigJSON(body []byte) (CandidateConfig, error) {
	required := []string{"agent_id", "name", "description", "model_name", "tool_names", "skill_names", "knowledge_base_names", "agent_md", "change_reason"}
	raw, err := parseObjectJSON(
		body,
		"candidate config",
		map[string]struct{}{
			"agent_id":             {},
			"name":                 {},
			"description":          {},
			"model_name":           {},
			"model_config_ref":     {},
			"tool_names":           {},
			"skill_names":          {},
			"knowledge_base_names": {},
			"catalog_files":        {},
			"channel_bindings":     {},
			"agent_md":             {},
			"change_reason":        {},
		},
		required,
	)
	if err != nil {
		return CandidateConfig{}, err
	}
	var cfg CandidateConfig
	if err := decodeRequired(raw, "agent_id", &cfg.AgentID); err != nil {
		return CandidateConfig{}, err
	}
	if err := decodeRequired(raw, "name", &cfg.Name); err != nil {
		return CandidateConfig{}, err
	}
	if err := decodeRequired(raw, "description", &cfg.Description); err != nil {
		return CandidateConfig{}, err
	}
	if err := decodeRequired(raw, "model_name", &cfg.ModelName); err != nil {
		return CandidateConfig{}, err
	}
	if raw["model_config_ref"] != nil {
		if err := decodeRequired(raw, "model_config_ref", &cfg.ModelConfigRef); err != nil {
			return CandidateConfig{}, err
		}
	}
	if err := decodeRequired(raw, "tool_names", &cfg.ToolNames); err != nil {
		return CandidateConfig{}, err
	}
	if err := decodeRequired(raw, "skill_names", &cfg.SkillNames); err != nil {
		return CandidateConfig{}, err
	}
	if err := decodeRequired(raw, "knowledge_base_names", &cfg.KnowledgeBaseNames); err != nil {
		return CandidateConfig{}, err
	}
	if _, exists := raw["catalog_files"]; exists {
		if err := decodeCatalogFiles(raw, "catalog_files", &cfg.CatalogFiles); err != nil {
			return CandidateConfig{}, err
		}
	} else {
		cfg.CatalogFiles = []agentpackage.CatalogFileRef{}
	}
	if err := validateCatalogFiles(cfg.CatalogFiles); err != nil {
		return CandidateConfig{}, err
	}
	if _, exists := raw["channel_bindings"]; exists {
		if err := decodeChannelBindings(raw, "channel_bindings", &cfg.ChannelBindings); err != nil {
			return CandidateConfig{}, err
		}
	} else {
		cfg.ChannelBindings = []agentpackage.ChannelBindingRef{}
	}
	if err := validateChannelBindings(cfg.ToolNames, cfg.SkillNames, cfg.ChannelBindings); err != nil {
		return CandidateConfig{}, err
	}
	if err := decodeRequired(raw, "agent_md", &cfg.AgentMD); err != nil {
		return CandidateConfig{}, err
	}
	if err := decodeRequired(raw, "change_reason", &cfg.ChangeReason); err != nil {
		return CandidateConfig{}, err
	}
	return cfg, nil
}

func ParseEditableConfigJSON(body []byte) (EditableConfig, error) {
	raw, err := parseObjectJSON(
		body,
		"editable agent config",
		map[string]struct{}{
			"agent_id":             {},
			"name":                 {},
			"description":          {},
			"model_name":           {},
			"model_config_ref":     {},
			"tool_names":           {},
			"skill_names":          {},
			"knowledge_base_names": {},
			"catalog_files":        {},
			"channel_bindings":     {},
			"agent_md":             {},
		},
		[]string{"agent_id", "name", "description", "model_name", "tool_names", "skill_names", "knowledge_base_names", "catalog_files", "channel_bindings", "agent_md"},
	)
	if err != nil {
		return EditableConfig{}, err
	}
	var cfg EditableConfig
	if err := decodeRequired(raw, "agent_id", &cfg.AgentID); err != nil {
		return EditableConfig{}, err
	}
	if err := decodeRequired(raw, "name", &cfg.Name); err != nil {
		return EditableConfig{}, err
	}
	if err := decodeRequired(raw, "description", &cfg.Description); err != nil {
		return EditableConfig{}, err
	}
	if err := decodeRequired(raw, "model_name", &cfg.ModelName); err != nil {
		return EditableConfig{}, err
	}
	if raw["model_config_ref"] != nil {
		if err := decodeRequired(raw, "model_config_ref", &cfg.ModelConfigRef); err != nil {
			return EditableConfig{}, err
		}
	}
	if err := decodeRequired(raw, "tool_names", &cfg.ToolNames); err != nil {
		return EditableConfig{}, err
	}
	if err := decodeRequired(raw, "skill_names", &cfg.SkillNames); err != nil {
		return EditableConfig{}, err
	}
	if err := decodeRequired(raw, "knowledge_base_names", &cfg.KnowledgeBaseNames); err != nil {
		return EditableConfig{}, err
	}
	if err := decodeCatalogFiles(raw, "catalog_files", &cfg.CatalogFiles); err != nil {
		return EditableConfig{}, err
	}
	if err := validateCatalogFiles(cfg.CatalogFiles); err != nil {
		return EditableConfig{}, err
	}
	if err := decodeChannelBindings(raw, "channel_bindings", &cfg.ChannelBindings); err != nil {
		return EditableConfig{}, err
	}
	if err := validateChannelBindings(cfg.ToolNames, cfg.SkillNames, cfg.ChannelBindings); err != nil {
		return EditableConfig{}, err
	}
	if err := decodeRequired(raw, "agent_md", &cfg.AgentMD); err != nil {
		return EditableConfig{}, err
	}
	return cfg, nil
}

func RenderCandidateSource(input RenderInput) (map[string][]byte, error) {
	if input.LoadVersion == "" {
		return nil, fmt.Errorf("%w: load_version is required", ErrInvalidAuthoringInput)
	}
	return renderSource(sourceInput{
		AgentID:           input.Config.AgentID,
		Name:              input.Config.Name,
		Description:       &input.Config.Description,
		Version:           input.LoadVersion,
		Model:             input.Config.ModelName,
		ModelConfigRef:    input.Config.ModelConfigRef,
		Instruction:       input.Config.AgentMD,
		ToolRefs:          renderResourceRefs(input.ToolRefs, input.Config.ToolNames),
		SkillRefs:         renderResourceRefs(input.SkillRefs, input.Config.SkillNames),
		KnowledgeBaseRefs: renderResourceRefs(input.KnowledgeRefs, input.Config.KnowledgeBaseNames),
		CatalogFiles:      input.Config.CatalogFiles,
		ChannelBindings:   input.Config.ChannelBindings,
	})
}

func CreateInitialSource(input InitialSourceInput) (map[string][]byte, error) {
	return renderSource(sourceInput{
		AgentID:         input.AgentID,
		Name:            input.Name,
		Description:     input.Description,
		Version:         input.Version,
		Model:           input.Model,
		Instruction:     input.Instruction,
		RequireNonEmpty: true,
	})
}

func ProjectEditableConfig(pkg *agentpackage.Package) (EditableConfig, error) {
	if pkg == nil {
		return EditableConfig{}, fmt.Errorf("%w: package is required", ErrInvalidAuthoringInput)
	}
	promptPath := pkg.Manifest.Instruction.Path
	if promptPath == "" {
		return EditableConfig{}, fmt.Errorf("%w: instruction path is required", ErrInvalidAuthoringInput)
	}
	prompt, exists := pkg.Files[promptPath]
	if !exists {
		return EditableConfig{}, fmt.Errorf("%w: instruction file is missing: %s", ErrInvalidAuthoringInput, promptPath)
	}
	if !utf8.Valid(prompt) {
		return EditableConfig{}, fmt.Errorf("%w: instruction file must be valid UTF-8: %s", ErrInvalidAuthoringInput, promptPath)
	}
	tools, err := projectNamedResourceRefs(pkg.Manifest.MOIResourceRefs.Tools, "tools")
	if err != nil {
		return EditableConfig{}, err
	}
	skills, err := projectNamedResourceRefs(pkg.Manifest.MOIResourceRefs.Skills, "skills")
	if err != nil {
		return EditableConfig{}, err
	}
	knowledgeBases, err := projectNamedResourceRefs(pkg.Manifest.MOIResourceRefs.KnowledgeBases, "knowledgeBases")
	if err != nil {
		return EditableConfig{}, err
	}
	catalogFiles := make([]agentpackage.CatalogFileRef, len(pkg.Manifest.MOIResourceRefs.CatalogFiles))
	copy(catalogFiles, pkg.Manifest.MOIResourceRefs.CatalogFiles)
	channelBindings := cloneChannelBindings(pkg.Manifest.MOIResourceRefs.ChannelBindings)
	return EditableConfig{
		AgentID:            pkg.Manifest.ID,
		Name:               pkg.Manifest.Name,
		Description:        stringValue(pkg.Manifest.Description),
		ModelName:          pkg.Manifest.Model,
		ModelConfigRef:     pkg.Manifest.ModelConfigRef,
		ToolNames:          tools,
		SkillNames:         skills,
		KnowledgeBaseNames: knowledgeBases,
		CatalogFiles:       catalogFiles,
		ChannelBindings:    channelBindings,
		AgentMD:            string(prompt),
	}, nil
}

func EditableSourceFiles(files map[string][]byte) (map[string][]byte, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("%w: package source files are required", ErrInvalidAuthoringInput)
	}
	out := make(map[string][]byte, len(files))
	for name, body := range files {
		if name == agentpackage.ChecksumPath || isGeneratedCustomToolArtifactFile(name) {
			continue
		}
		out[name] = append([]byte(nil), body...)
	}
	return out, nil
}

func SourceDigest(files map[string][]byte) (string, error) {
	if len(files) == 0 {
		return "", fmt.Errorf("%w: package source files are required", ErrInvalidAuthoringInput)
	}
	names := make([]string, 0, len(files))
	for name := range files {
		if name == agentpackage.ChecksumPath {
			return "", fmt.Errorf("%w: source files must not include generated %s", ErrInvalidAuthoringInput, agentpackage.ChecksumPath)
		}
		names = append(names, name)
	}
	sort.Strings(names)
	h := sha256.New()
	for _, name := range names {
		sum := sha256.Sum256(files[name])
		h.Write([]byte(name))
		h.Write([]byte{0})
		h.Write([]byte(hex.EncodeToString(sum[:])))
		h.Write([]byte{0})
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}

func projectNamedResourceRefs(refs []agentpackage.NamedResourceRef, field string) ([]string, error) {
	if len(refs) == 0 {
		return []string{}, nil
	}
	out := make([]string, 0, len(refs))
	for _, ref := range refs {
		if ref.Name == "" {
			return nil, fmt.Errorf("%w: moiResourceRefs.%s contains empty name", ErrInvalidAuthoringInput, field)
		}
		out = append(out, ref.Name)
	}
	return out, nil
}

func isGeneratedCustomToolArtifactFile(name string) bool {
	return strings.HasPrefix(name, agentpackage.CustomToolRoot+"/") &&
		(strings.HasSuffix(name, "/"+agentpackage.CustomToolArtifactFile) ||
			strings.HasSuffix(name, "/"+agentpackage.CustomToolArtifactManifestFile))
}

type sourceInput struct {
	AgentID           string
	Name              string
	Description       *string
	Version           string
	Model             string
	ModelConfigRef    string
	Instruction       string
	ToolRefs          []agentpackage.NamedResourceRef
	SkillRefs         []agentpackage.NamedResourceRef
	KnowledgeBaseRefs []agentpackage.NamedResourceRef
	CatalogFiles      []agentpackage.CatalogFileRef
	ChannelBindings   []agentpackage.ChannelBindingRef
	RequireNonEmpty   bool
}

func renderSource(input sourceInput) (map[string][]byte, error) {
	if err := validateCatalogFiles(input.CatalogFiles); err != nil {
		return nil, err
	}
	if err := agentpackage.ValidateResourceChannelBindingRefs(input.ToolRefs, input.SkillRefs, input.ChannelBindings); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidAuthoringInput, err)
	}
	if input.RequireNonEmpty && input.AgentID == "" {
		return nil, fmt.Errorf("%w: agent_id is required", ErrInvalidAuthoringInput)
	}
	if input.RequireNonEmpty && input.Name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidAuthoringInput)
	}
	if input.RequireNonEmpty && input.Version == "" {
		return nil, fmt.Errorf("%w: version is required", ErrInvalidAuthoringInput)
	}
	if input.RequireNonEmpty && input.Model == "" {
		return nil, fmt.Errorf("%w: model_name is required", ErrInvalidAuthoringInput)
	}
	if input.RequireNonEmpty && input.Instruction == "" {
		return nil, fmt.Errorf("%w: agent_md is required", ErrInvalidAuthoringInput)
	}
	manifest := "schemaVersion: 1\n\n" +
		"id: " + yamlString(input.AgentID) + "\n" +
		"version: " + yamlString(input.Version) + "\n" +
		"name: " + yamlString(input.Name) + "\n" +
		renderOptionalDescription(input.Description) +
		"model: " + yamlString(input.Model) + "\n" +
		renderOptionalModelConfigRef(input.ModelConfigRef) + "\n" +
		"instruction:\n" +
		"  path: " + rolePromptPath + "\n\n" +
		"moiResourceRefs:\n" +
		renderNamedRefs("tools", input.ToolRefs) +
		renderNamedRefs("skills", input.SkillRefs) +
		renderNamedRefs("knowledgeBases", input.KnowledgeBaseRefs) +
		renderCatalogFiles(input.CatalogFiles) +
		renderChannelBindings(input.ChannelBindings)
	return map[string][]byte{
		agentpackage.ManifestPath: []byte(manifest),
		rolePromptPath:            []byte(input.Instruction),
	}, nil
}

func renderOptionalDescription(description *string) string {
	if description == nil {
		return ""
	}
	return "description: " + yamlString(*description) + "\n"
}

func renderOptionalModelConfigRef(ref string) string {
	if ref == "" {
		return ""
	}
	return "model_config_ref: " + yamlString(ref) + "\n"
}

func renderResourceRefs(refs []agentpackage.NamedResourceRef, names []string) []agentpackage.NamedResourceRef {
	if len(refs) != 0 || len(names) == 0 {
		return refs
	}
	out := make([]agentpackage.NamedResourceRef, 0, len(names))
	for _, name := range names {
		out = append(out, agentpackage.NamedResourceRef{Name: name})
	}
	return out
}

func renderNamedRefs(field string, refs []agentpackage.NamedResourceRef) string {
	if len(refs) == 0 {
		return "  " + field + ": []\n"
	}
	out := "  " + field + ":\n"
	for _, ref := range refs {
		out += "    - name: " + yamlString(ref.Name) + "\n"
		if ref.WorkspaceID != "" {
			out += "      workspaceID: " + yamlString(ref.WorkspaceID) + "\n"
		}
	}
	return out
}

func renderCatalogFiles(refs []agentpackage.CatalogFileRef) string {
	if len(refs) == 0 {
		return "  catalogFiles: []\n"
	}
	out := "  catalogFiles:\n"
	for _, ref := range refs {
		out += "    - workspaceID: " + yamlString(ref.WorkspaceID) + "\n"
		out += "      volumeID: " + fmt.Sprintf("%d", ref.VolumeID) + "\n"
		out += "      fileID: " + yamlString(ref.FileID) + "\n"
	}
	return out
}

func renderChannelBindings(refs []agentpackage.ChannelBindingRef) string {
	if len(refs) == 0 {
		return "  channelBindings: []\n"
	}
	out := "  channelBindings:\n"
	for _, ref := range refs {
		if ref.ToolName != "" {
			out += "    - toolName: " + yamlString(ref.ToolName) + "\n"
			if ref.ToolWorkspaceID != "" {
				out += "      toolWorkspaceID: " + yamlString(ref.ToolWorkspaceID) + "\n"
			}
		} else {
			out += "    - skillName: " + yamlString(ref.SkillName) + "\n"
			if ref.SkillWorkspaceID != "" {
				out += "      skillWorkspaceID: " + yamlString(ref.SkillWorkspaceID) + "\n"
			}
			out += "      provider: " + yamlString(ref.Provider) + "\n"
		}
		out += "      workspaceID: " + yamlString(ref.WorkspaceID) + "\n"
		out += "      instanceID: " + yamlString(ref.InstanceID) + "\n"
	}
	return out
}

func validateCatalogFiles(refs []agentpackage.CatalogFileRef) error {
	if err := agentpackage.ValidateCatalogFileRefs(refs); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidAuthoringInput, err)
	}
	return nil
}

func validateChannelBindings(toolNames, skillNames []string, refs []agentpackage.ChannelBindingRef) error {
	tools := make([]agentpackage.NamedResourceRef, 0, len(toolNames))
	for _, name := range toolNames {
		tools = append(tools, agentpackage.NamedResourceRef{Name: name})
	}
	for _, ref := range refs {
		if ref.ToolWorkspaceID == "" {
			continue
		}
		for index := range tools {
			if tools[index].Name == ref.ToolName {
				tools[index].WorkspaceID = ref.ToolWorkspaceID
			}
		}
	}
	skills := make([]agentpackage.NamedResourceRef, 0, len(skillNames))
	for _, name := range skillNames {
		skills = append(skills, agentpackage.NamedResourceRef{Name: name})
	}
	for _, ref := range refs {
		if ref.SkillWorkspaceID == "" {
			continue
		}
		for index := range skills {
			if skills[index].Name == ref.SkillName {
				skills[index].WorkspaceID = ref.SkillWorkspaceID
			}
		}
	}
	if err := agentpackage.ValidateResourceChannelBindingRefs(tools, skills, refs); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidAuthoringInput, err)
	}
	return nil
}

func cloneChannelBindings(refs []agentpackage.ChannelBindingRef) []agentpackage.ChannelBindingRef {
	out := make([]agentpackage.ChannelBindingRef, len(refs))
	copy(out, refs)
	return out
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func yamlString(value string) string {
	data, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func parseRequiredObjectJSON(body []byte, subject string, allowed map[string]struct{}) (map[string]json.RawMessage, error) {
	required := make([]string, 0, len(allowed))
	for key := range allowed {
		required = append(required, key)
	}
	return parseObjectJSON(body, subject, allowed, required)
}

func parseObjectJSON(body []byte, subject string, allowed map[string]struct{}, required []string) (map[string]json.RawMessage, error) {
	var raw map[string]json.RawMessage
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&raw); err != nil {
		return nil, fmt.Errorf("%w: decode %s: %v", ErrInvalidAuthoringInput, subject, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return nil, fmt.Errorf("%w: %s must be one JSON object", ErrInvalidAuthoringInput, subject)
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return nil, fmt.Errorf("%w: unknown field %q", ErrInvalidAuthoringInput, key)
		}
	}
	for _, key := range required {
		if _, ok := raw[key]; !ok {
			return nil, fmt.Errorf("%w: missing required field %q", ErrInvalidAuthoringInput, key)
		}
	}
	return raw, nil
}

func decodeRequired(raw map[string]json.RawMessage, key string, out any) error {
	if bytes.Equal(raw[key], []byte("null")) {
		return fmt.Errorf("%w: field %q must not be null", ErrInvalidAuthoringInput, key)
	}
	if err := json.Unmarshal(raw[key], out); err != nil {
		return fmt.Errorf("%w: field %q has invalid JSON shape: %v", ErrInvalidAuthoringInput, key, err)
	}
	return nil
}

func decodeCatalogFiles(raw map[string]json.RawMessage, key string, out *[]agentpackage.CatalogFileRef) error {
	if bytes.Equal(raw[key], []byte("null")) {
		return fmt.Errorf("%w: field %q must not be null", ErrInvalidAuthoringInput, key)
	}
	var items []json.RawMessage
	if err := json.Unmarshal(raw[key], &items); err != nil {
		return fmt.Errorf("%w: field %q has invalid JSON shape: %v", ErrInvalidAuthoringInput, key, err)
	}
	refs := make([]agentpackage.CatalogFileRef, 0, len(items))
	for index, item := range items {
		entry, err := parseObjectJSON(item, fmt.Sprintf("%s[%d]", key, index), map[string]struct{}{
			"workspace_id": {},
			"volume_id":    {},
			"file_id":      {},
		}, []string{"workspace_id", "volume_id", "file_id"})
		if err != nil {
			return err
		}
		var ref agentpackage.CatalogFileRef
		if err := decodeRequired(entry, "workspace_id", &ref.WorkspaceID); err != nil {
			return err
		}
		if err := decodeRequired(entry, "volume_id", &ref.VolumeID); err != nil {
			return err
		}
		if err := decodeRequired(entry, "file_id", &ref.FileID); err != nil {
			return err
		}
		refs = append(refs, ref)
	}
	*out = refs
	return nil
}

func decodeChannelBindings(raw map[string]json.RawMessage, key string, out *[]agentpackage.ChannelBindingRef) error {
	if bytes.Equal(raw[key], []byte("null")) {
		return fmt.Errorf("%w: field %q must not be null", ErrInvalidAuthoringInput, key)
	}
	var items []json.RawMessage
	if err := json.Unmarshal(raw[key], &items); err != nil {
		return fmt.Errorf("%w: field %q has invalid JSON shape: %v", ErrInvalidAuthoringInput, key, err)
	}
	refs := make([]agentpackage.ChannelBindingRef, 0, len(items))
	for index, item := range items {
		entry, err := parseObjectJSON(item, fmt.Sprintf("%s[%d]", key, index), map[string]struct{}{
			"tool_name":          {},
			"tool_workspace_id":  {},
			"skill_name":         {},
			"skill_workspace_id": {},
			"provider":           {},
			"workspace_id":       {},
			"instance_id":        {},
		}, []string{"workspace_id", "instance_id"})
		if err != nil {
			return err
		}
		var ref agentpackage.ChannelBindingRef
		if entry["tool_name"] != nil {
			if err := decodeRequired(entry, "tool_name", &ref.ToolName); err != nil {
				return err
			}
		}
		if entry["tool_workspace_id"] != nil {
			if err := decodeRequired(entry, "tool_workspace_id", &ref.ToolWorkspaceID); err != nil {
				return err
			}
		}
		if entry["skill_name"] != nil {
			if err := decodeRequired(entry, "skill_name", &ref.SkillName); err != nil {
				return err
			}
		}
		if entry["skill_workspace_id"] != nil {
			if err := decodeRequired(entry, "skill_workspace_id", &ref.SkillWorkspaceID); err != nil {
				return err
			}
		}
		if entry["provider"] != nil {
			if err := decodeRequired(entry, "provider", &ref.Provider); err != nil {
				return err
			}
		}
		if err := decodeRequired(entry, "workspace_id", &ref.WorkspaceID); err != nil {
			return err
		}
		if err := decodeRequired(entry, "instance_id", &ref.InstanceID); err != nil {
			return err
		}
		refs = append(refs, ref)
	}
	*out = refs
	return nil
}
