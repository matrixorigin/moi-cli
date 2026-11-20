package commands

import (
	"context"
	"flag"
	"fmt"
	"os"
)

func runConnectorServiceExportTaskStats(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector export-task-stats", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector export-task-stats [options]\n\nProduct SDK: ConnectorService.ExportTaskStats\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Connectors(wid).ExportTaskStats(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}
