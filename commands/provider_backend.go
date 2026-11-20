package commands

import (
	"context"
	"flag"
	"fmt"
	"os"
)

// ExecuteProviderBackendCommand runs product SDK ProviderBackendService methods via moi-cli.
func ExecuteProviderBackendCommand(ctx *Context, args []string) error {
	if len(args) == 0 {
		ShowProviderBackendHelp(nil)
		return nil
	}
	sub := args[0]
	switch sub {
	case "list":
		return runProviderBackendServiceList(ctx, args[1:])
	case "create":
		return runProviderBackendServiceCreate(ctx, args[1:])
	case "get":
		return runProviderBackendServiceGet(ctx, args[1:])
	case "update":
		return runProviderBackendServiceUpdate(ctx, args[1:])
	case "delete":
		return runProviderBackendServiceDelete(ctx, args[1:])
	case "create-endpoint":
		return runProviderBackendServiceCreateEndpoint(ctx, args[1:])
	case "list-endpoints":
		return runProviderBackendServiceListEndpoints(ctx, args[1:])
	case "set-endpoint-status":
		return runProviderBackendServiceSetEndpointStatus(ctx, args[1:])
	case "router":
		return runProviderBackendServiceRouter(ctx, args[1:])
	case "update-router":
		return runProviderBackendServiceUpdateRouter(ctx, args[1:])
	case "probe-backend-models":
		return runProviderBackendServiceProbeBackendModels(ctx, args[1:])
	default:
		return fmt.Errorf("unknown provider-backend subcommand: %s", sub)
	}
}

// ShowProviderBackendHelp prints help for provider-backend.
func ShowProviderBackendHelp(_ []string) {
	fmt.Fprintf(os.Stderr, `Usage: provider-backend <subcommand> [options]

Product SDK surface: ProviderBackendService

Subcommands:
  list                             ProviderBackendService.List
  create                           ProviderBackendService.Create
  get                              ProviderBackendService.Get
  update                           ProviderBackendService.Update
  delete                           ProviderBackendService.Delete
  create-endpoint                  ProviderBackendService.CreateEndpoint
  list-endpoints                   ProviderBackendService.ListEndpoints
  set-endpoint-status              ProviderBackendService.SetEndpointStatus
  router                           ProviderBackendService.Router
  update-router                    ProviderBackendService.UpdateRouter
  probe-backend-models             ProviderBackendService.ProbeBackendModels

Complex arguments accept --json / --json-file.
Stream methods accept --output <path>.

`)
}

func runProviderBackendServiceList(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("provider-backend list", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	kind := fs.String("kind", "llm", "Backend kind: llm|embedding|parser")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli provider-backend list [options]\n\nProduct SDK: ProviderBackendService.List\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := providerBackendService(ctx, wid, *kind).List(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runProviderBackendServiceCreate(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("provider-backend create", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	kind := fs.String("kind", "llm", "Backend kind: llm|embedding|parser")
	name := fs.String("name", "", "name")
	options := addProviderBackendCreateFlags(fs)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli provider-backend create [options]\n\nProduct SDK: ProviderBackendService.Create\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	values, err := options.createOptions(fs)
	if err != nil {
		return err
	}
	result, err := providerBackendService(ctx, wid, *kind).Create(context.Background(), *name, values...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runProviderBackendServiceGet(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("provider-backend get", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	kind := fs.String("kind", "llm", "Backend kind: llm|embedding|parser")
	id := fs.String("id", "", "id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli provider-backend get [options]\n\nProduct SDK: ProviderBackendService.Get\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := providerBackendService(ctx, wid, *kind).Get(context.Background(), *id)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runProviderBackendServiceUpdate(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("provider-backend update", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	kind := fs.String("kind", "llm", "Backend kind: llm|embedding|parser")
	id := fs.String("id", "", "id")
	options := addProviderBackendUpdateFlags(fs)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli provider-backend update [options]\n\nProduct SDK: ProviderBackendService.Update\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	values, err := options.updateOptions(fs)
	if err != nil {
		return err
	}
	result, err := providerBackendService(ctx, wid, *kind).Update(context.Background(), *id, values...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runProviderBackendServiceDelete(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("provider-backend delete", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	kind := fs.String("kind", "llm", "Backend kind: llm|embedding|parser")
	id := fs.String("id", "", "id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli provider-backend delete [options]\n\nProduct SDK: ProviderBackendService.Delete\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := providerBackendService(ctx, wid, *kind).Delete(context.Background(), *id)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runProviderBackendServiceCreateEndpoint(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("provider-backend create-endpoint", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	kind := fs.String("kind", "llm", "Backend kind: llm|embedding|parser")
	backendID := fs.String("backend-id", "", "backendID")
	address := fs.String("address", "", "address")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli provider-backend create-endpoint [options]\n\nProduct SDK: ProviderBackendService.CreateEndpoint\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := providerBackendService(ctx, wid, *kind).CreateEndpoint(context.Background(), *backendID, *address)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runProviderBackendServiceListEndpoints(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("provider-backend list-endpoints", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	kind := fs.String("kind", "llm", "Backend kind: llm|embedding|parser")
	backendID := fs.String("backend-id", "", "backendID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli provider-backend list-endpoints [options]\n\nProduct SDK: ProviderBackendService.ListEndpoints\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := providerBackendService(ctx, wid, *kind).ListEndpoints(context.Background(), *backendID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runProviderBackendServiceSetEndpointStatus(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("provider-backend set-endpoint-status", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	kind := fs.String("kind", "llm", "Backend kind: llm|embedding|parser")
	backendID := fs.String("backend-id", "", "backendID")
	endpointID := fs.String("endpoint-id", "", "endpointID")
	status := fs.String("status", "", "status")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli provider-backend set-endpoint-status [options]\n\nProduct SDK: ProviderBackendService.SetEndpointStatus\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := providerBackendService(ctx, wid, *kind).SetEndpointStatus(context.Background(), *backendID, *endpointID, *status)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runProviderBackendServiceRouter(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("provider-backend router", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	kind := fs.String("kind", "llm", "Backend kind: llm|embedding|parser")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli provider-backend router [options]\n\nProduct SDK: ProviderBackendService.Router\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := providerBackendService(ctx, wid, *kind).Router(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runProviderBackendServiceUpdateRouter(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("provider-backend update-router", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	kind := fs.String("kind", "llm", "Backend kind: llm|embedding|parser")
	options := addProviderRouterFlags(fs)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli provider-backend update-router [options]\n\nProduct SDK: ProviderBackendService.UpdateRouter\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := providerBackendService(ctx, wid, *kind).UpdateRouter(context.Background(), options.options(fs)...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runProviderBackendServiceProbeBackendModels(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("provider-backend probe-backend-models", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	kind := fs.String("kind", "llm", "Backend kind: llm|embedding|parser")
	backendID := fs.String("backend-id", "", "backendID")
	options := addProviderBackendProbeFlags(fs)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli provider-backend probe-backend-models [options]\n\nProduct SDK: ProviderBackendService.ProbeBackendModels\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	values, err := options.createOptions(fs)
	if err != nil {
		return err
	}
	result, err := providerBackendService(ctx, wid, *kind).ProbeBackendModels(context.Background(), *backendID, values...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}
