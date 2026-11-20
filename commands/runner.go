package commands

import (
	"context"
	"flag"
	"fmt"
	"os"
)

func ExecuteRunnerCommand(ctx *Context, args []string) error {
	if len(args) == 0 {
		ShowRunnerHelp(nil)
		return nil
	}
	switch args[0] {
	case "create":
		return runRunnerCreate(ctx, args[1:])
	case "list":
		return runRunnerList(ctx, args[1:])
	case "get":
		return runRunnerGet(ctx, args[1:])
	case "refresh":
		return runRunnerRefresh(ctx, args[1:])
	case "delete":
		return runRunnerDelete(ctx, args[1:])
	default:
		return fmt.Errorf("unknown runner subcommand: %s", args[0])
	}
}

func ShowRunnerHelp(_ []string) {
	fmt.Fprintln(os.Stderr, `Usage: moi-cli runner <subcommand> [options]
create, list, get, refresh, delete

Create and refresh return a one-time connect_cmd. Get and list do not issue credentials.
Use --workspace-id for the workspace and --runner-id for an existing runner.`)
}

func runRunnerCreate(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("runner create", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	name := fs.String("name", "", "Runner name")
	return runWorkspaceRequest(ctx, fs, args, workspaceID, func(wid string) (any, error) {
		service := ctx.Client.Runners(wid)
		return service.Create(context.Background(), *name)
	})
}

func runRunnerList(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("runner list", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	return runWorkspaceRequest(ctx, fs, args, workspaceID, func(wid string) (any, error) {
		service := ctx.Client.Runners(wid)
		return service.List(context.Background())
	})
}

func runRunnerGet(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("runner get", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	id := fs.String("runner-id", "", "Runner ID")
	return runWorkspaceRequest(ctx, fs, args, workspaceID, func(wid string) (any, error) {
		service := ctx.Client.Runners(wid)
		return service.Get(context.Background(), *id)
	})
}

func runRunnerRefresh(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("runner refresh", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	id := fs.String("runner-id", "", "Runner ID")
	return runWorkspaceRequest(ctx, fs, args, workspaceID, func(wid string) (any, error) {
		service := ctx.Client.Runners(wid)
		return service.Refresh(context.Background(), *id)
	})
}

func runRunnerDelete(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("runner delete", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	id := fs.String("runner-id", "", "Runner ID")
	return runWorkspaceRequest(ctx, fs, args, workspaceID, func(wid string) (any, error) {
		service := ctx.Client.Runners(wid)
		return nil, service.Delete(context.Background(), *id)
	})
}
