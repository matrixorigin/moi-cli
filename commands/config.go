package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

// ExecuteConfigCommand runs the config command.
func ExecuteConfigCommand(_ *Context, args []string) error {
	if len(args) == 0 {
		ShowConfigHelp(nil)
		return nil
	}
	sub := args[0]
	switch sub {
	case "set-workspace":
		return runConfigSetWorkspace(args[1:])
	case "set-endpoint":
		return runConfigSetEndpoint(args[1:])
	case "set-api-key":
		return runConfigSetAPIKey(args[1:])
	case "set-role":
		return runConfigSetRole(args[1:])
	case "unset-role":
		return runConfigUnsetRole(args[1:])
	case "show":
		return runConfigShow(args[1:])
	default:
		return fmt.Errorf("unknown config subcommand: %s (use set-workspace, set-endpoint, set-api-key, set-role, unset-role, or show)", sub)
	}
}

func runConfigSetEndpoint(args []string) error {
	fs := flag.NewFlagSet("config set-endpoint", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: moi-cli config set-endpoint <url>

Set the product API endpoint and save to ~/.moi-cli.json. Subsequent
commands will use this value when -endpoint is not provided.

The endpoint must include the environment product base path, for example:
  http://127.0.0.1:8050/newmoi
  https://example.com/newmoi

Example:
  moi-cli config set-endpoint http://127.0.0.1:8050/newmoi
`)
	}
	_ = fs.Parse(args)
	if fs.NArg() == 0 {
		fs.Usage()
		return fmt.Errorf("endpoint URL is required")
	}
	endpoint := fs.Arg(0)
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	if cfg == nil {
		cfg = &Config{}
	}
	cfg.Endpoint = endpoint
	if err := SaveConfig(cfg); err != nil {
		return err
	}
	fmt.Printf("Endpoint set to: %s\n", endpoint)
	return nil
}

func runConfigSetAPIKey(args []string) error {
	fs := flag.NewFlagSet("config set-api-key", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: moi-cli config set-api-key <pat>

Set the human UC personal access token (PAT) and save to ~/.moi-cli.json.
Subsequent commands will use this value when -api-key is not provided.

Example:
  moi-cli config set-api-key moi_xxxxxxxx
`)
	}
	_ = fs.Parse(args)
	if fs.NArg() == 0 {
		fs.Usage()
		return fmt.Errorf("personal access token is required")
	}
	apiKey := fs.Arg(0)
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	if cfg == nil {
		cfg = &Config{}
	}
	cfg.APIKey = apiKey
	if err := SaveConfig(cfg); err != nil {
		return err
	}
	fmt.Printf("Personal access token set (saved to config).\n")
	return nil
}

func runConfigSetWorkspace(args []string) error {
	fs := flag.NewFlagSet("config set-workspace", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: moi-cli config set-workspace <workspace-id>

Set the current workspace and save to ~/.moi-cli.json. Subsequent commands
that need a workspace will use this value when --workspace-id is not provided.

Example:
  moi-cli config set-workspace my-workspace-id
`)
	}
	_ = fs.Parse(args)
	if fs.NArg() == 0 {
		fs.Usage()
		return fmt.Errorf("workspace-id is required")
	}
	workspaceID := fs.Arg(0)
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	if cfg == nil {
		cfg = &Config{}
	}
	cfg.WorkspaceID = workspaceID
	if err := SaveConfig(cfg); err != nil {
		return err
	}
	fmt.Printf("Workspace set to: %s\n", workspaceID)
	return nil
}

func runConfigSetRole(args []string) error {
	fs := flag.NewFlagSet("config set-role", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: moi-cli config set-role <workspace-id> <role-id>")
	}
	_ = fs.Parse(args)
	if fs.NArg() != 2 {
		fs.Usage()
		return fmt.Errorf("workspace-id and role-id are required")
	}
	workspaceID, roleID := fs.Arg(0), fs.Arg(1)
	if workspaceID == "" {
		return sdk.ErrWorkspaceRequired
	}
	if err := sdk.ValidateCurrentRoleID(roleID); err != nil {
		return err
	}
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	if cfg == nil {
		cfg = &Config{}
	}
	cfg.clearInvalidWorkspaceCurrentRoleContainer()
	if cfg.WorkspaceCurrentRoleIDs == nil {
		cfg.WorkspaceCurrentRoleIDs = make(map[string]json.RawMessage)
	}
	encodedRoleID, err := json.Marshal(roleID)
	if err != nil {
		return err
	}
	cfg.WorkspaceCurrentRoleIDs[workspaceID] = encodedRoleID
	if err := SaveConfig(cfg); err != nil {
		return err
	}
	fmt.Printf("Current role for workspace %s set to: %s\n", workspaceID, roleID)
	return nil
}

func runConfigUnsetRole(args []string) error {
	fs := flag.NewFlagSet("config unset-role", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: moi-cli config unset-role <workspace-id>")
	}
	_ = fs.Parse(args)
	if fs.NArg() != 1 {
		fs.Usage()
		return fmt.Errorf("workspace-id is required")
	}
	workspaceID := fs.Arg(0)
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}
	if cfg == nil || cfg.WorkspaceCurrentRoleIDs == nil {
		if cfg != nil && cfg.hasInvalidWorkspaceCurrentRoleContainer() {
			cfg.clearInvalidWorkspaceCurrentRoleContainer()
			if err := SaveConfig(cfg); err != nil {
				return err
			}
			fmt.Println("Malformed workspace current role mappings removed.")
			return nil
		}
		return fmt.Errorf("workspace %s has no configured current role", workspaceID)
	}
	if _, ok := cfg.WorkspaceCurrentRoleIDs[workspaceID]; !ok {
		return fmt.Errorf("workspace %s has no configured current role", workspaceID)
	}
	delete(cfg.WorkspaceCurrentRoleIDs, workspaceID)
	if err := SaveConfig(cfg); err != nil {
		return err
	}
	fmt.Printf("Current role for workspace %s removed.\n", workspaceID)
	return nil
}

func runConfigShow(args []string) error {
	fs := flag.NewFlagSet("config show", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, `Usage: moi-cli config show

Show the conversation's injected connection, or saved configuration outside a conversation.
`)
	}
	_ = fs.Parse(args)
	cfg, runtime := RuntimeConfig()
	if !runtime {
		var err error
		cfg, err = LoadConfig()
		if err != nil {
			return err
		}
	}
	if cfg == nil {
		fmt.Println("No config file (~/.moi-cli.json). Use config set-endpoint / set-api-key / set-workspace.")
		return nil
	}
	if runtime {
		fmt.Println("Runtime configuration (environment):")
	} else {
		path, _ := GetConfigPath()
		fmt.Printf("Config file: %s\n", path)
	}
	if cfg.Endpoint != "" {
		fmt.Printf("  endpoint:     %s\n", cfg.Endpoint)
	} else {
		fmt.Println("  endpoint:     (not set, use -endpoint or moi-cli config set-endpoint <url>)")
	}
	if cfg.APIKey != "" {
		fmt.Println("  api_key:      ********")
	} else {
		fmt.Println("  api_key:      (not set, use -api-key or moi-cli config set-api-key <key>)")
	}
	if cfg.WorkspaceID != "" {
		fmt.Printf("  workspace_id: %s\n", cfg.WorkspaceID)
	} else {
		fmt.Println("  workspace_id: (not set, use: moi-cli config set-workspace <id>)")
	}
	if reason := workspaceCurrentRoleContainerIssue(cfg); reason != "" {
		fmt.Printf("  workspace_current_role_ids: (invalid: %s)\n", reason)
		fmt.Println("    fix: moi-cli config set-role <workspace-id> <valid-role-id> or moi-cli config unset-role <workspace-id>")
	} else if len(cfg.WorkspaceCurrentRoleIDs) == 0 {
		fmt.Println("  workspace_current_role_ids: (none)")
	} else {
		fmt.Println("  workspace_current_role_ids:")
		issues := WorkspaceCurrentRoleIssues(cfg)
		issueReasons := make(map[string]string, len(issues))
		for _, issue := range issues {
			issueReasons[issue.WorkspaceID] = issue.Reason
		}
		workspaceIDs := make([]string, 0, len(cfg.WorkspaceCurrentRoleIDs))
		for workspaceID := range cfg.WorkspaceCurrentRoleIDs {
			workspaceIDs = append(workspaceIDs, workspaceID)
		}
		sort.Strings(workspaceIDs)
		for _, workspaceID := range workspaceIDs {
			roleID := workspaceCurrentRoleDisplay(cfg.WorkspaceCurrentRoleIDs[workspaceID])
			if reason, ok := issueReasons[workspaceID]; ok {
				fmt.Printf("    %s: %s [invalid: %s]\n", workspaceID, roleID, reason)
				continue
			}
			fmt.Printf("    %s: %s\n", workspaceID, roleID)
		}
	}
	return nil
}

// ShowConfigHelp prints help for the config command.
func ShowConfigHelp(_ []string) {
	fmt.Fprintf(os.Stderr, `Usage: config <subcommand> [options]

Manage moi-cli configuration (saved to ~/.moi-cli.json). Role selection is
workspace-scoped and persistent; there is no global one-shot Role override.

If config show marks a workspace Role mapping as invalid, use set-role to
replace it or unset-role to remove it before running business commands.

Subcommands:
  set-workspace <id>   Set current workspace; used when --workspace-id is omitted
  set-endpoint <url>   Set product API endpoint; used when -endpoint is omitted
  set-api-key <key>    Set human PAT; used when -api-key is omitted
  set-role <ws> <role> Set the current Role for one workspace
  unset-role <ws>      Remove the configured current Role for one workspace
  show                 Show current config and workspace Role mappings

Examples:
  moi-cli config set-endpoint http://127.0.0.1:8050/newmoi
  moi-cli config set-api-key moi_xxxxxxxx
  moi-cli config set-workspace my-workspace-id
  moi-cli config set-role my-workspace-id 12
  moi-cli config show

`)
}
