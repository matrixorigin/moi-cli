package commands

import (
	"fmt"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

// Context holds the CLI execution context: product SDK client and resolved workspace ID.
type Context struct {
	Client            *sdk.Client
	WorkspaceID       string
	Verbose           bool
	OutputJSON        bool
	NoInteractive     bool
	Endpoint          string
	APIKey            string
	RunningMOIVersion string
	PreflightOnly     bool
}

// NewContext creates a context with the given client and workspace ID.
func NewContext(client *sdk.Client, workspaceID string, verbose, outputJSON, noInteractive bool) *Context {
	return &Context{
		Client:        client,
		WorkspaceID:   workspaceID,
		Verbose:       verbose,
		OutputJSON:    outputJSON,
		NoInteractive: noInteractive,
	}
}

// resolveWorkspaceID returns the workspace id from flag or config.
func resolveWorkspaceID(ctx *Context, fromFlag string) (string, error) {
	resolved := ""
	if fromFlag != "" {
		resolved = fromFlag
	} else if ctx != nil && ctx.WorkspaceID != "" {
		resolved = ctx.WorkspaceID
	} else {
		return "", fmt.Errorf("workspace-id is required (set with: moi-cli config set-workspace <id> or use --workspace-id)")
	}
	return resolved, nil
}
