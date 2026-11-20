package commands

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"
)

func runAgentRuntimeBrowserToken(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent runtime-browser-token", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	taskID := fs.String("task-id", "", "Agent runtime task ID")
	return runWorkspaceRequest(ctx, fs, args, workspaceID, func(wid string) (any, error) {
		return ctx.Client.Agents(wid).RuntimeBrowserToken(context.Background(), *taskID)
	})
}

func runAgentConnectRuntimeBrowser(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent connect-runtime-browser", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	taskID := fs.String("task-id", "", "Agent runtime task ID")
	token := fs.String("token", "", "Short-lived token returned by agent runtime-browser-token")
	origin := fs.String("origin", "", "Allowed HTTP(S) origin for the WebSocket handshake")
	return runWorkspaceRequest(ctx, fs, args, workspaceID, func(wid string) (any, error) {
		browser, err := ctx.Client.Agents(wid).ConnectRuntimeBrowser(context.Background(), *taskID, *token, *origin)
		if err != nil {
			return nil, err
		}
		if err := browser.Close(); err != nil {
			return nil, err
		}
		// This command probes connectivity only; it is not an interactive
		// browser controller or evidence that the runtime task has completed.
		return map[string]any{"connected": true}, nil
	})
}

// Connect once and read the initial browser event, including the original image
// bytes (JSON base64) for a frame. Closing the viewer leaves the browser running.
func runAgentBrowser(ctx *Context, args []string, command string) error {
	fs := flag.NewFlagSet("agent "+command, flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	taskID := fs.String("task-id", "", "Runtime task ID")
	compression := fs.Bool("compression", true, "Negotiate browser stream compression")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent %s --task-id <id> [options]\n\nRead the initial browser event and image.\n\nOptions:\n", command)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	workspace, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	operation, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	connect := ctx.Client.Agents(workspace).ConnectBrowser
	switch command {
	case "connect-browser":
	case "watch-browser":
		connect = ctx.Client.Agents(workspace).WatchBrowser
	}
	browser, err := connect(operation, *taskID, *compression)
	if err != nil {
		return err
	}
	defer browser.Close()
	event, image, err := browser.Receive(operation)
	if err != nil {
		return err
	}
	return printJSON(map[string]any{"event": event, "image": image, "compression": browser.CompressionEnabled()})
}
