package commands

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
)

func runCatalogAllowedActions(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog allowed-actions", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	resourceType := fs.String("resource-type", "", "catalog, database, or volume")
	resourceIDs := fs.String("resource-ids", "", "Comma-separated resource IDs (1–100)")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: moi-cli catalog allowed-actions --resource-type <type> --resource-ids <ids> [options]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Catalog(wid).AllowedActions(context.Background(), *resourceType, strings.Split(*resourceIDs, ","))
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}
