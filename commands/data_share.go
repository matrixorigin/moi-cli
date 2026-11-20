package commands

import (
	"context"
	"flag"
	"fmt"
	"os"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

// ExecuteDataShareCommand runs product SDK DataShareService methods via moi-cli.
func ExecuteDataShareCommand(ctx *Context, args []string) error {
	if len(args) == 0 {
		ShowDataShareHelp(nil)
		return nil
	}
	sub := args[0]
	switch sub {
	case "publishes":
		return runDataShareServicePublishes(ctx, args[1:])
	case "workspace-names-result":
		return runDataShareServiceWorkspaceNamesResult(ctx, args[1:])
	case "resolve-workspace-names":
		return runDataShareServiceResolveWorkspaceNames(ctx, args[1:])
	case "publish-summary":
		return runDataShareServicePublishSummary(ctx, args[1:])
	case "source-tables":
		return runDataShareServiceSourceTables(ctx, args[1:])
	case "check-publish-name":
		return runDataShareServiceCheckPublishName(ctx, args[1:])
	case "create-publish":
		return runDataShareServiceCreatePublish(ctx, args[1:])
	case "update-publish":
		return runDataShareServiceUpdatePublish(ctx, args[1:])
	case "delete-publish":
		return runDataShareServiceDeletePublish(ctx, args[1:])
	case "subscriptions":
		return runDataShareServiceSubscriptions(ctx, args[1:])
	case "subscribe":
		return runDataShareServiceSubscribe(ctx, args[1:])
	case "unsubscribe":
		return runDataShareServiceUnsubscribe(ctx, args[1:])
	default:
		return fmt.Errorf("unknown data-share subcommand: %s", sub)
	}
}

// ShowDataShareHelp prints help for data-share.
func ShowDataShareHelp(_ []string) {
	fmt.Fprintf(os.Stderr, `Usage: data-share <subcommand> [options]

Product SDK surface: DataShareService

Subcommands:
  publishes                        DataShareService.Publishes
  workspace-names-result           DataShareService.WorkspaceNamesResult
  resolve-workspace-names          DataShareService.ResolveWorkspaceNames
  publish-summary                  DataShareService.PublishSummary
  source-tables                    DataShareService.SourceTables
  check-publish-name               DataShareService.CheckPublishName
  create-publish                   DataShareService.CreatePublish
  update-publish                   DataShareService.UpdatePublish
  delete-publish                   DataShareService.DeletePublish
  subscriptions                    DataShareService.Subscriptions
  subscribe                        DataShareService.Subscribe
  unsubscribe                      DataShareService.Unsubscribe

Complex arguments accept --json / --json-file.
Stream methods accept --output <path>.

`)
}

func runDataShareServicePublishes(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("data-share publishes", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.String("keyword", "", "Publication name filter")
	fs.Int("page", 0, "Page number (omitted: server default)")
	fs.Int("page-size", 0, "Page size (omitted: server default)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli data-share publishes [options]\n\nProduct SDK: DataShareService.Publishes\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.DataShare(wid).Publishes(context.Background(), dataShareQueryOptions(fs)...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runDataShareServiceWorkspaceNamesResult(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("data-share workspace-names-result", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	ids := fs.String("ids", "", "ids (comma-separated)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli data-share workspace-names-result [options]\n\nProduct SDK: DataShareService.WorkspaceNamesResult\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.DataShare(wid).WorkspaceNamesResult(context.Background(), splitCSV(*ids))
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runDataShareServiceResolveWorkspaceNames(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("data-share resolve-workspace-names", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	ids := fs.String("ids", "", "ids (comma-separated)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli data-share resolve-workspace-names [options]\n\nProduct SDK: DataShareService.ResolveWorkspaceNames\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.DataShare(wid).ResolveWorkspaceNames(context.Background(), splitCSV(*ids))
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runDataShareServicePublishSummary(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("data-share publish-summary", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli data-share publish-summary [options]\n\nProduct SDK: DataShareService.PublishSummary\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.DataShare(wid).PublishSummary(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runDataShareServiceSourceTables(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("data-share source-tables", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	sourceDatabaseID := fs.String("source-database-id", "", "source database ID")
	fs.Int("page", 0, "Page number (omitted: server default)")
	fs.Int("page-size", 0, "Page size (omitted: server default)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli data-share source-tables [options]\n\nProduct SDK: DataShareService.SourceTables\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.DataShare(wid).SourceTables(context.Background(), *sourceDatabaseID, dataShareQueryOptions(fs)...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runDataShareServiceCheckPublishName(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("data-share check-publish-name", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.String("name", "", "Publication name to check")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli data-share check-publish-name [options]\n\nProduct SDK: DataShareService.CheckPublishName\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.DataShare(wid).CheckPublishName(context.Background(), dataShareQueryOptions(fs)...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runDataShareServiceCreatePublish(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("data-share create-publish", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	name := fs.String("name", "", "name")
	sourceDatabaseID := fs.String("source-database-id", "", "source database ID")
	tableScopeMode := fs.String("table-scope-mode", "selected", "table scope mode: all or selected")
	tableIDs := fs.String("table-ids", "", "MatrixOne table rel_ids (comma-separated; required for selected)")
	targetWorkspaceIDs := fs.String("target-workspace-ids", "", "target workspace IDs (comma-separated)")
	remark := fs.String("remark", "", "Publication remark (omitted: no remark; explicit empty sends an empty remark)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli data-share create-publish [options]\n\nProduct SDK: DataShareService.CreatePublish\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	tableScope := &sdk.DataShareScope{Mode: *tableScopeMode, ObjectIds: splitCSV(*tableIDs)}
	var options []sdk.DataSharePublishOption
	fs.Visit(func(option *flag.Flag) {
		if option.Name == "remark" {
			options = append(options, sdk.WithDataShareRemark(*remark))
		}
	})
	result, err := ctx.Client.DataShare(wid).CreatePublish(context.Background(), *name, *sourceDatabaseID, tableScope, splitCSV(*targetWorkspaceIDs), options...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runDataShareServiceUpdatePublish(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("data-share update-publish", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	publishID := fs.String("publish-id", "", "publishID")
	tableScopeMode := fs.String("table-scope-mode", "", "Replacement scope: all or selected (provide when changing tables)")
	tableIDs := fs.String("table-ids", "", "Replacement MatrixOne table rel_ids (comma-separated; use with selected scope)")
	targetWorkspaceIDs := fs.String("target-workspace-ids", "", "Replacement target workspace IDs (comma-separated)")
	remark := fs.String("remark", "", "Replacement remark (empty clears it)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli data-share update-publish [options]\n\nProduct SDK: DataShareService.UpdatePublish\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	var options []sdk.DataSharePublishUpdateOption
	scopeChanged := false
	fs.Visit(func(option *flag.Flag) {
		switch option.Name {
		case "table-scope-mode", "table-ids":
			scopeChanged = true
		case "target-workspace-ids":
			options = append(options, sdk.WithDataShareUpdateTargetWorkspaceIDs(splitCSV(*targetWorkspaceIDs)...))
		case "remark":
			options = append(options, sdk.WithDataShareUpdateRemark(*remark))
		}
	})
	if scopeChanged {
		options = append(options, sdk.WithDataShareUpdateTableScope(&sdk.DataShareScope{Mode: *tableScopeMode, ObjectIds: splitCSV(*tableIDs)}))
	}
	result, err := ctx.Client.DataShare(wid).UpdatePublish(context.Background(), *publishID, options...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runDataShareServiceDeletePublish(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("data-share delete-publish", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	publishID := fs.String("publish-id", "", "publishID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli data-share delete-publish [options]\n\nProduct SDK: DataShareService.DeletePublish\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if err := ctx.Client.DataShare(wid).DeletePublish(context.Background(), *publishID); err != nil {
		return err
	}
	fmt.Printf("data-share delete-publish: ok\n")
	return nil
}

func runDataShareServiceSubscriptions(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("data-share subscriptions", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.String("keyword", "", "Subscription name filter")
	fs.String("status", "", "Subscription status filter")
	fs.Int("page", 0, "Page number (omitted: server default)")
	fs.Int("page-size", 0, "Page size (omitted: server default)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli data-share subscriptions [options]\n\nProduct SDK: DataShareService.Subscriptions\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.DataShare(wid).Subscriptions(context.Background(), dataShareQueryOptions(fs)...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runDataShareServiceSubscribe(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("data-share subscribe", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	subscriptionID := fs.String("subscription-id", "", "subscriptionID")
	subName := fs.String("sub-name", "", "subName")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli data-share subscribe [options]\n\nProduct SDK: DataShareService.Subscribe\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.DataShare(wid).Subscribe(context.Background(), *subscriptionID, *subName)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runDataShareServiceUnsubscribe(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("data-share unsubscribe", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	subscriptionID := fs.String("subscription-id", "", "subscriptionID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli data-share unsubscribe [options]\n\nProduct SDK: DataShareService.Unsubscribe\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if err := ctx.Client.DataShare(wid).Unsubscribe(context.Background(), *subscriptionID); err != nil {
		return err
	}
	fmt.Printf("data-share unsubscribe: ok\n")
	return nil
}

// dataShareQueryOptions preserves omission so pagination defaults and validation
// remain with the product owner, including explicitly supplied zero values.
func dataShareQueryOptions(fs *flag.FlagSet) []sdk.QueryOption {
	var options []sdk.QueryOption
	fs.Visit(func(option *flag.Flag) {
		switch option.Name {
		case "page", "keyword", "status", "name":
			options = append(options, sdk.WithQueryParam(option.Name, option.Value.String()))
		case "page-size":
			options = append(options, sdk.WithQueryParam("page_size", option.Value.String()))
		}
	})
	return options
}
