package commands

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

const redactedWorkspaceCurrentRole = "<redacted>"

// WorkspaceCurrentRoleIssue describes one invalid persisted workspace Role mapping.
type WorkspaceCurrentRoleIssue struct {
	WorkspaceID string
	RoleID      string
	Reason      string
}

// WorkspaceCurrentRoleConfigError prevents business requests from using a
// partially valid Role configuration and gives config commands enough detail to repair it.
type WorkspaceCurrentRoleConfigError struct {
	ConfigPath      string
	ContainerReason string
	Issues          []WorkspaceCurrentRoleIssue
}

func (e *WorkspaceCurrentRoleConfigError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "invalid current role configuration in %s:\n", e.ConfigPath)
	if e.ContainerReason != "" {
		fmt.Fprintf(&b, "  workspace_current_role_ids: %s\n", e.ContainerReason)
		b.WriteString("    fix: moi-cli config set-role <workspace-id> <valid-role-id> or moi-cli config unset-role <workspace-id>\n")
	}
	for _, issue := range e.Issues {
		fmt.Fprintf(&b, "  workspace %q, role %q: %s\n", issue.WorkspaceID, issue.RoleID, issue.Reason)
		fmt.Fprintf(&b, "    fix: moi-cli config set-role %q <valid-role-id> or moi-cli config unset-role %q\n", issue.WorkspaceID, issue.WorkspaceID)
	}
	b.WriteString("No request was sent.")
	return b.String()
}

// WorkspaceCurrentRoleIssues validates every persisted mapping in sorted workspace order.
func WorkspaceCurrentRoleIssues(cfg *Config) []WorkspaceCurrentRoleIssue {
	if cfg == nil || len(cfg.WorkspaceCurrentRoleIDs) == 0 {
		return nil
	}
	workspaceIDs := make([]string, 0, len(cfg.WorkspaceCurrentRoleIDs))
	for workspaceID := range cfg.WorkspaceCurrentRoleIDs {
		workspaceIDs = append(workspaceIDs, workspaceID)
	}
	sort.Strings(workspaceIDs)

	issues := make([]WorkspaceCurrentRoleIssue, 0)
	for _, workspaceID := range workspaceIDs {
		rawRoleID := cfg.WorkspaceCurrentRoleIDs[workspaceID]
		roleID, decodeErr := decodeWorkspaceCurrentRoleID(rawRoleID)
		displayRoleID := workspaceCurrentRoleDisplay(rawRoleID)
		if workspaceID == "" {
			issues = append(issues, WorkspaceCurrentRoleIssue{
				WorkspaceID: workspaceID,
				RoleID:      displayRoleID,
				Reason:      "workspace ID must not be empty",
			})
			continue
		}
		if decodeErr != nil {
			issues = append(issues, WorkspaceCurrentRoleIssue{
				WorkspaceID: workspaceID,
				RoleID:      displayRoleID,
				Reason:      "role ID must be a JSON string containing a canonical positive int64 decimal",
			})
			continue
		}
		if err := sdk.ValidateCurrentRoleID(roleID); err != nil {
			issues = append(issues, WorkspaceCurrentRoleIssue{
				WorkspaceID: workspaceID,
				RoleID:      displayRoleID,
				Reason:      currentRoleIssueReason(roleID),
			})
		}
	}
	return issues
}

// WorkspaceCurrentRoleSnapshot validates all persisted mappings and returns the
// typed values used to construct one immutable Product SDK client.
func WorkspaceCurrentRoleSnapshot(cfg *Config) (map[string]string, error) {
	containerReason := workspaceCurrentRoleContainerIssue(cfg)
	issues := WorkspaceCurrentRoleIssues(cfg)
	if containerReason != "" || len(issues) != 0 {
		path, err := GetConfigPath()
		if err != nil {
			return nil, err
		}
		return nil, &WorkspaceCurrentRoleConfigError{
			ConfigPath:      path,
			ContainerReason: containerReason,
			Issues:          issues,
		}
	}
	if cfg == nil || len(cfg.WorkspaceCurrentRoleIDs) == 0 {
		return nil, nil
	}

	roles := make(map[string]string, len(cfg.WorkspaceCurrentRoleIDs))
	for workspaceID, rawRoleID := range cfg.WorkspaceCurrentRoleIDs {
		roleID, err := decodeWorkspaceCurrentRoleID(rawRoleID)
		if err != nil {
			return nil, err
		}
		roles[workspaceID] = roleID
	}
	return roles, nil
}

func workspaceCurrentRoleContainerIssue(cfg *Config) string {
	if cfg != nil && cfg.hasInvalidWorkspaceCurrentRoleContainer() {
		return "must be a JSON object containing workspace-to-Role mappings"
	}
	return ""
}

func decodeWorkspaceCurrentRoleID(rawRoleID json.RawMessage) (string, error) {
	var roleID string
	if err := json.Unmarshal(rawRoleID, &roleID); err != nil {
		return "", err
	}
	return roleID, nil
}

func workspaceCurrentRoleDisplay(rawRoleID json.RawMessage) string {
	if roleID, err := decodeWorkspaceCurrentRoleID(rawRoleID); err == nil && sdk.ValidateCurrentRoleID(roleID) == nil {
		return roleID
	}
	return redactedWorkspaceCurrentRole
}

func currentRoleIssueReason(roleID string) string {
	if roleID == "" {
		return "role ID must be a canonical positive int64 decimal"
	}
	if strings.TrimSpace(roleID) != roleID {
		return "role ID must not contain surrounding whitespace"
	}
	if strings.HasPrefix(roleID, "+") || strings.HasPrefix(roleID, "-") {
		return "role ID must not include a sign"
	}
	if len(roleID) > 1 && roleID[0] == '0' {
		return "role ID must not have leading zeros"
	}
	for index := 0; index < len(roleID); index++ {
		if roleID[index] < '0' || roleID[index] > '9' {
			return "role ID must contain only decimal digits"
		}
	}
	if _, err := strconv.ParseInt(roleID, 10, 64); err != nil {
		return "role ID must fit in a positive int64"
	}
	return "role ID must be a canonical positive int64 decimal"
}
