package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// Config holds the product CLI configuration persisted to ~/.moi-cli.json.
type Config struct {
	Endpoint                string                     `json:"endpoint,omitempty"`
	APIKey                  string                     `json:"api_key,omitempty"`
	WorkspaceID             string                     `json:"workspace_id,omitempty"`
	WorkspaceCurrentRoleIDs map[string]json.RawMessage `json:"workspace_current_role_ids,omitempty"`

	invalidWorkspaceCurrentRoleIDs json.RawMessage
}

type configJSON struct {
	Endpoint                string          `json:"endpoint,omitempty"`
	APIKey                  string          `json:"api_key,omitempty"`
	WorkspaceID             string          `json:"workspace_id,omitempty"`
	WorkspaceCurrentRoleIDs json.RawMessage `json:"workspace_current_role_ids,omitempty"`
}

func (c *Config) UnmarshalJSON(data []byte) error {
	var wire configJSON
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}

	*c = Config{
		Endpoint:    wire.Endpoint,
		APIKey:      wire.APIKey,
		WorkspaceID: wire.WorkspaceID,
	}
	if len(wire.WorkspaceCurrentRoleIDs) == 0 {
		return nil
	}
	trimmed := bytes.TrimSpace(wire.WorkspaceCurrentRoleIDs)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		c.invalidWorkspaceCurrentRoleIDs = append(json.RawMessage(nil), wire.WorkspaceCurrentRoleIDs...)
		return nil
	}
	return json.Unmarshal(wire.WorkspaceCurrentRoleIDs, &c.WorkspaceCurrentRoleIDs)
}

func (c Config) MarshalJSON() ([]byte, error) {
	roleMappings := c.invalidWorkspaceCurrentRoleIDs
	if len(roleMappings) == 0 && len(c.WorkspaceCurrentRoleIDs) != 0 {
		encoded, err := json.Marshal(c.WorkspaceCurrentRoleIDs)
		if err != nil {
			return nil, err
		}
		roleMappings = encoded
	}
	return json.Marshal(configJSON{
		Endpoint:                c.Endpoint,
		APIKey:                  c.APIKey,
		WorkspaceID:             c.WorkspaceID,
		WorkspaceCurrentRoleIDs: roleMappings,
	})
}

func (c *Config) hasInvalidWorkspaceCurrentRoleContainer() bool {
	return c != nil && len(c.invalidWorkspaceCurrentRoleIDs) != 0
}

func (c *Config) clearInvalidWorkspaceCurrentRoleContainer() {
	if c != nil {
		c.invalidWorkspaceCurrentRoleIDs = nil
	}
}

// GetConfigPath selects an explicit installation config or the legacy default.
// An invalid explicit path must never fall back to another installation's PAT.
func GetConfigPath() (string, error) {
	if path, set := os.LookupEnv("MOI_CLI_CONFIG_FILE"); set {
		if !filepath.IsAbs(path) || filepath.Clean(path) != path || path == string(filepath.Separator) {
			return "", errors.New("CLI_CONFIG_PATH_INVALID")
		}
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".moi-cli.json"), nil
}

// LoadConfig reads the config from the default config file.
// Returns nil config and nil error if the file does not exist.
func LoadConfig() (*Config, error) {
	path, err := GetConfigPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// RuntimeConfig reads the connection injected into a conversation command.
// Saved interactive credentials and role selections do not participate.
func RuntimeConfig() (*Config, bool) {
	if _, present := os.LookupEnv("MOI_RUNTIME_AUTHORIZATION"); !present {
		return nil, false
	}
	return &Config{
		Endpoint: os.Getenv("MOI_ENDPOINT"), APIKey: os.Getenv("MOI_API_KEY"), WorkspaceID: os.Getenv("MOI_WORKSPACE_ID"),
	}, true
}

// SaveConfig writes the config to the default config file.
func SaveConfig(cfg *Config) error {
	if cfg == nil {
		return nil
	}
	path, err := GetConfigPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	// Never rewrite the existing inode: a reader that opened an older,
	// broader-mode file would keep access and observe the new credentials.
	// Write a private temporary file in the same directory and rename it
	// over the config path so the old inode never sees the new content.
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	committed := false
	defer func() {
		if !committed {
			_ = tmp.Close()
			_ = os.Remove(tmpPath)
		}
	}()
	if err := tmp.Chmod(0600); err != nil {
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}
	committed = true
	return nil
}
