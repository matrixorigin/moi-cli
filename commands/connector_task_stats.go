package commands

import (
	"context"
	"flag"
	"fmt"
	"os"
)

func runConnectorServiceTaskStats(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector task-stats", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector task-stats [options]\n\nProduct SDK: ConnectorService.TaskStats\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Connectors(wid).TaskStats(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}
