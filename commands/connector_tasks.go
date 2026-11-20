package commands

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func runConnectorServiceTasks(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("connector tasks", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	page := fs.Int("page", 0, "Page number (omit to use the MOI default)")
	pageSize := fs.Int("page-size", 0, "Tasks per page (omit to use the MOI default)")
	keyword := fs.String("keyword", "", "Search task name")
	orderBy := fs.String("order-by", "", "Sort field")
	isDesc := fs.Bool("is-desc", true, "Sort descending (omit to use the MOI default)")
	var statuses, connectorSources, loadIntervalTypes, sourceConnectorIDs []string
	fs.Func("status", "Task status (repeatable)", func(value string) error {
		statuses = append(statuses, value)
		return nil
	})
	fs.Func("connector-source", "Connector source type (repeatable)", func(value string) error {
		connectorSources = append(connectorSources, value)
		return nil
	})
	fs.Func("load-interval-type", "Load interval type (repeatable)", func(value string) error {
		loadIntervalTypes = append(loadIntervalTypes, value)
		return nil
	})
	fs.Func("source-connector-id", "Source connector ID (repeatable)", func(value string) error {
		sourceConnectorIDs = append(sourceConnectorIDs, value)
		return nil
	})
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli connector tasks [options]\n\nProduct SDK: ConnectorService.Tasks\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	var opts []sdk.QueryOption
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "page":
			opts = append(opts, sdk.WithPage(*page))
		case "page-size":
			opts = append(opts, sdk.WithPageSize(*pageSize))
		case "keyword":
			opts = append(opts, sdk.WithQueryParam("keyword", *keyword))
		case "order-by":
			opts = append(opts, sdk.WithQueryParam("order_by", *orderBy))
		case "is-desc":
			opts = append(opts, sdk.WithQueryParam("is_desc", strconv.FormatBool(*isDesc)))
		}
	})
	opts = append(opts,
		sdk.WithQueryValues("status", statuses...),
		sdk.WithQueryValues("connector_sources", connectorSources...),
		sdk.WithQueryValues("load_interval_types", loadIntervalTypes...),
		sdk.WithQueryValues("source_connector_ids", sourceConnectorIDs...),
	)
	result, err := ctx.Client.Connectors(wid).Tasks(context.Background(), opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}
