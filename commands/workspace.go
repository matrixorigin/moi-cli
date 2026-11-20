package commands

import (
	"context"
	"flag"
	"fmt"
	"os"
)

// ExecuteWorkspaceCommand runs product SDK WorkspaceService methods via moi-cli.
func ExecuteWorkspaceCommand(ctx *Context, args []string) error {
	if len(args) == 0 {
		ShowWorkspaceHelp(nil)
		return nil
	}
	sub := args[0]
	switch sub {
	case "update":
		return runWorkspaceServiceUpdate(ctx, args[1:])
	case "start-creation":
		return runWorkspaceServiceStartCreation(ctx, args[1:])
	case "creations":
		return runWorkspaceServiceCreations(ctx, args[1:])
	case "ensure":
		return runWorkspaceEnsureAlias(ctx, args[1:])
	case "get":
		return runWorkspaceGetAlias(ctx, args[1:])
	case "get-db-connection":
		return runWorkspaceGetDBConnectionAlias(ctx, args[1:])
	case "list":
		return runWorkspaceServiceList(ctx, args[1:])
	case "create":
		return runWorkspaceServiceCreate(ctx, args[1:])
	case "claim":
		return runWorkspaceServiceClaim(ctx, args[1:])
	case "delete":
		return runWorkspaceServiceDelete(ctx, args[1:])
	default:
		return fmt.Errorf("unknown workspace subcommand: %s", sub)
	}
}

// ShowWorkspaceHelp prints help for workspace.
func ShowWorkspaceHelp(_ []string) {
	fmt.Fprintf(os.Stderr, `Usage: workspace <subcommand> [options]

Product SDK surface: WorkspaceService

Subcommands:
  update                           WorkspaceService.Update
  start-creation                   WorkspaceService.StartCreation
  creations                        WorkspaceService.Creations
  ensure                           (alias)
  get                              (alias)
  get-db-connection                (alias)
  list                             WorkspaceService.List
  create                           WorkspaceService.Create
  claim                            WorkspaceService.Claim
  delete                           WorkspaceService.Delete

Complex arguments accept --json / --json-file.
Stream methods accept --output <path>.

`)
}

func runWorkspaceServiceList(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workspace list", flag.ExitOnError)
	var queryParams queryParameters
	fs.Var(&queryParams, "query-param", "Product query parameter name=value (repeatable; pagination/filter names follow the product API)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workspace list [options]\n\nProduct SDK: WorkspaceService.List\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	result, err := ctx.Client.Workspaces().List(context.Background(), queryParams...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkspaceServiceCreate(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workspace create", flag.ExitOnError)
	name := fs.String("name", "", "name")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workspace create [options]\n\nProduct SDK: WorkspaceService.Create\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	result, err := ctx.Client.Workspaces().Create(context.Background(), *name)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkspaceServiceClaim(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workspace claim", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspaceID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workspace claim [options]\n\nProduct SDK: WorkspaceService.Claim\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	result, err := ctx.Client.Workspaces().Claim(context.Background(), *workspaceID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkspaceServiceDelete(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workspace delete", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspaceID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workspace delete [options]\n\nProduct SDK: WorkspaceService.Delete\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if err := ctx.Client.Workspaces().Delete(context.Background(), *workspaceID); err != nil {
		return err
	}
	fmt.Printf("workspace delete: ok\n")
	return nil
}
func runWorkspaceServiceCreations(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workspace creations", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workspace creations [options]\n\nProduct SDK: WorkspaceService.Creations\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	result, err := ctx.Client.Workspaces().Creations(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkspaceServiceStartCreation(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workspace start-creation", flag.ExitOnError)
	name := fs.String("name", "", "name")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workspace start-creation [options]\n\nProduct SDK: WorkspaceService.StartCreation\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	result, err := ctx.Client.Workspaces().StartCreation(context.Background(), *name)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkspaceServiceUpdate(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workspace update", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	name := fs.String("name", "", "name")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workspace update [options]\n\nProduct SDK: WorkspaceService.Update\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	return ctx.Client.Workspaces().Update(context.Background(), wid, *name)
}
