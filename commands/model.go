package commands

import (
	sdk "github.com/matrixorigin/moi-go-sdk/v5"

	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

// ExecuteModelCommand runs product SDK ModelService methods via moi-cli.
func ExecuteModelCommand(ctx *Context, args []string) error {
	if len(args) == 0 {
		ShowModelHelp(nil)
		return nil
	}
	sub := args[0]
	switch sub {
	case "set-workspace-llm-endpoint-status":
		return runModelServiceSetWorkspaceLLMEndpointStatus(ctx, args[1:])
	case "set-workspace-embedding-endpoint-status":
		return runModelServiceSetWorkspaceEmbeddingEndpointStatus(ctx, args[1:])
	case "create-workspace-embedding-endpoint":
		return runModelServiceCreateWorkspaceEmbeddingEndpoint(ctx, args[1:])
	case "update-workspace-embedding-router":
		return runModelServiceUpdateWorkspaceEmbeddingRouter(ctx, args[1:])
	case "update-workspace-embedding-backend":
		return runModelServiceUpdateWorkspaceEmbeddingBackend(ctx, args[1:])
	case "create-workspace-embedding-backend":
		return runModelServiceCreateWorkspaceEmbeddingBackend(ctx, args[1:])
	case "list-workspace-llm-endpoints":
		return runModelServiceListWorkspaceLLMEndpoints(ctx, args[1:])
	case "list-workspace-embedding-endpoints":
		return runModelServiceListWorkspaceEmbeddingEndpoints(ctx, args[1:])
	case "delete-workspace-embedding-backend":
		return runModelServiceDeleteWorkspaceEmbeddingBackend(ctx, args[1:])
	case "get-workspace-embedding-backend":
		return runModelServiceGetWorkspaceEmbeddingBackend(ctx, args[1:])
	case "list-workspace-llm-models":
		return runModelServiceListWorkspaceLLMModels(ctx, args[1:])
	case "workspace-embedding-router":
		return runModelServiceWorkspaceEmbeddingRouter(ctx, args[1:])
	case "list-workspace-embedding-backends":
		return runModelServiceListWorkspaceEmbeddingBackends(ctx, args[1:])
	case "workspace-chat-completion-stream":
		return runModelServiceWorkspaceChatCompletionStream(ctx, args[1:])
	case "workspace-chat-completion":
		return runModelServiceWorkspaceChatCompletion(ctx, args[1:])
	case "update-workspace-llm-router":
		return runModelServiceUpdateWorkspaceLLMRouter(ctx, args[1:])
	case "workspace-llm-router":
		return runModelServiceWorkspaceLLMRouter(ctx, args[1:])
	case "create-workspace-llm-endpoint":
		return runModelServiceCreateWorkspaceLLMEndpoint(ctx, args[1:])
	case "delete-workspace-llm-backend":
		return runModelServiceDeleteWorkspaceLLMBackend(ctx, args[1:])
	case "update-workspace-llm-backend":
		return runModelServiceUpdateWorkspaceLLMBackend(ctx, args[1:])
	case "get-workspace-llm-backend":
		return runModelServiceGetWorkspaceLLMBackend(ctx, args[1:])
	case "list-workspace-llm-backends":
		return runModelServiceListWorkspaceLLMBackends(ctx, args[1:])
	case "create-workspace-llm-backend":
		return runModelServiceCreateWorkspaceLLMBackend(ctx, args[1:])
	case "list":
		return runModelServiceList(ctx, args[1:])
	case "list-embedding-models":
		return runModelServiceListEmbeddingModels(ctx, args[1:])
	case "setup-llm-backend":
		return runModelServiceSetupLLMBackend(ctx, args[1:])
	case "probe-llm-models":
		return runModelServiceProbeLLMModels(ctx, args[1:])
	case "probe-embedding-models":
		return runModelServiceProbeEmbeddingModels(ctx, args[1:])
	default:
		return fmt.Errorf("unknown model subcommand: %s", sub)
	}
}

// ShowModelHelp prints help for model.
func ShowModelHelp(_ []string) {
	fmt.Fprintln(os.Stderr, "Additional commands: set-workspace-llm-endpoint-status, set-workspace-embedding-endpoint-status, create-workspace-embedding-endpoint, update-workspace-embedding-router, update-workspace-embedding-backend, create-workspace-embedding-backend, list-workspace-llm-endpoints, list-workspace-embedding-endpoints, delete-workspace-embedding-backend, get-workspace-embedding-backend, list-workspace-llm-models, workspace-embedding-router, list-workspace-embedding-backends")

	fmt.Fprintf(os.Stderr, `Usage: model <subcommand> [options]

Product SDK surface: ModelService

Subcommands:
  workspace-chat-completion-stream ModelService.WorkspaceChatCompletionStream
  workspace-chat-completion        ModelService.WorkspaceChatCompletion
  update-workspace-llm-router      ModelService.UpdateWorkspaceLLMRouter
  workspace-llm-router             ModelService.WorkspaceLLMRouter
  create-workspace-llm-endpoint    ModelService.CreateWorkspaceLLMEndpoint
  delete-workspace-llm-backend     ModelService.DeleteWorkspaceLLMBackend
  update-workspace-llm-backend     ModelService.UpdateWorkspaceLLMBackend
  get-workspace-llm-backend        ModelService.GetWorkspaceLLMBackend
  list-workspace-llm-backends      ModelService.ListWorkspaceLLMBackends
  create-workspace-llm-backend     ModelService.CreateWorkspaceLLMBackend
  list                             ModelService.List
  list-embedding-models            ModelService.ListEmbeddingModels
  setup-llm-backend                ModelService.SetupLLMBackend
  probe-llm-models                 ModelService.ProbeLLMModels
  probe-embedding-models           ModelService.ProbeEmbeddingModels

Complex arguments accept --json / --json-file.
Stream methods accept --output <path>.

`)
}

func runModelServiceList(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("model list", flag.ExitOnError)
	var queryParams queryParameters
	fs.Var(&queryParams, "query-param", "Product query parameter name=value (repeatable; pagination/filter names follow the product API)")
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli model list [options]\n\nProduct SDK: ModelService.List\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Models(wid).List(context.Background(), queryParams...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runModelServiceListEmbeddingModels(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("model list-embedding-models", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli model list-embedding-models [options]\n\nProduct SDK: ModelService.ListEmbeddingModels\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Models(wid).ListEmbeddingModels(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runModelServiceSetupLLMBackend(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("model setup-llm-backend", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	name := fs.String("name", "", "name")
	options := addProviderBackendCreateFlags(fs)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli model setup-llm-backend [options]\n\nProduct SDK: ModelService.SetupLLMBackend\n\nOptions:\n")
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
	result, err := ctx.Client.Models(wid).SetupLLMBackend(context.Background(), *name, options.address, splitCSV(options.models), values...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runModelServiceProbeLLMModels(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("model probe-llm-models", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	address := fs.String("address", "", "address")
	options := addProviderBackendProbeFlags(fs)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli model probe-llm-models [options]\n\nProduct SDK: ModelService.ProbeLLMModels\n\nOptions:\n")
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
	result, err := ctx.Client.Models(wid).ProbeLLMModels(context.Background(), *address, values...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runModelServiceProbeEmbeddingModels(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("model probe-embedding-models", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	address := fs.String("address", "", "address")
	options := addProviderBackendProbeFlags(fs)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli model probe-embedding-models [options]\n\nProduct SDK: ModelService.ProbeEmbeddingModels\n\nOptions:\n")
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
	result, err := ctx.Client.Models(wid).ProbeEmbeddingModels(context.Background(), *address, values...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}
func runModelServiceCreateWorkspaceLLMBackend(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("model create-workspace-llm-backend", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli model create-workspace-llm-backend [options]\n\nProduct SDK: ModelService.CreateWorkspaceLLMBackend\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	var request sdk.WorkspaceLLMBackendCreateRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Models(wid).CreateWorkspaceLLMBackend(context.Background(), &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runModelServiceListWorkspaceLLMBackends(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("model list-workspace-llm-backends", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli model list-workspace-llm-backends [options]\n\nProduct SDK: ModelService.ListWorkspaceLLMBackends\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Models(wid).ListWorkspaceLLMBackends(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runModelServiceGetWorkspaceLLMBackend(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("model get-workspace-llm-backend", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	backendID := fs.Int64("backend-id", 0, "backend-id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli model get-workspace-llm-backend [options]\n\nProduct SDK: ModelService.GetWorkspaceLLMBackend\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Models(wid).GetWorkspaceLLMBackend(context.Background(), *backendID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runModelServiceUpdateWorkspaceLLMBackend(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("model update-workspace-llm-backend", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	backendID := fs.Int64("backend-id", 0, "backend-id")
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli model update-workspace-llm-backend [options]\n\nProduct SDK: ModelService.UpdateWorkspaceLLMBackend\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	var request sdk.WorkspaceLLMBackendUpdateRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Models(wid).UpdateWorkspaceLLMBackend(context.Background(), *backendID, &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runModelServiceDeleteWorkspaceLLMBackend(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("model delete-workspace-llm-backend", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	backendID := fs.Int64("backend-id", 0, "backend-id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli model delete-workspace-llm-backend [options]\n\nProduct SDK: ModelService.DeleteWorkspaceLLMBackend\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	err = ctx.Client.Models(wid).DeleteWorkspaceLLMBackend(context.Background(), *backendID)
	if err != nil {
		return err
	}
	return nil
}

func runModelServiceCreateWorkspaceLLMEndpoint(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("model create-workspace-llm-endpoint", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	backendID := fs.Int64("backend-id", 0, "backend-id")
	address := fs.String("address", "", "address")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli model create-workspace-llm-endpoint [options]\n\nProduct SDK: ModelService.CreateWorkspaceLLMEndpoint\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Models(wid).CreateWorkspaceLLMEndpoint(context.Background(), *backendID, *address)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runModelServiceWorkspaceLLMRouter(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("model workspace-llm-router", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli model workspace-llm-router [options]\n\nProduct SDK: ModelService.WorkspaceLLMRouter\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Models(wid).WorkspaceLLMRouter(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runModelServiceUpdateWorkspaceLLMRouter(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("model update-workspace-llm-router", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli model update-workspace-llm-router [options]\n\nProduct SDK: ModelService.UpdateWorkspaceLLMRouter\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	var request sdk.WorkspaceLLMRouterUpdateRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Models(wid).UpdateWorkspaceLLMRouter(context.Background(), &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runModelServiceWorkspaceChatCompletion(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("model workspace-chat-completion", flag.ExitOnError)
	return runWorkspaceChatCompletion(ctx, args, fs, nil)
}

func runModelServiceWorkspaceChatCompletionStream(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("model workspace-chat-completion-stream", flag.ExitOnError)
	output := fs.String("output", "", "Output SSE file, or stdout")
	return runWorkspaceChatCompletion(ctx, args, fs, output)
}

func runWorkspaceChatCompletion(ctx *Context, args []string, fs *flag.FlagSet, output *string) error {
	method := "WorkspaceChatCompletion"
	if output != nil {
		method += "Stream"
	}
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	serviceSlot := fs.String("service-slot", "", "trusted service slot for billing scenario attribution")
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli %s [options]\n\nProduct SDK: ModelService.%s\n\nOptions:\n", fs.Name(), method)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	var request sdk.WorkspaceChatCompletionRequest
	if err := unmarshalProductJSON(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	var options []sdk.QueryOption
	if *serviceSlot != "" {
		options = append(options, sdk.WithQueryParam("service_slot", *serviceSlot))
	}
	if output != nil {
		result, err := ctx.Client.Models(wid).WorkspaceChatCompletionStream(context.Background(), &request, options...)
		if err != nil {
			return err
		}
		defer result.Body.Close()
		return writeStream(result.Body, *output)
	}
	result, err := ctx.Client.Models(wid).WorkspaceChatCompletion(context.Background(), &request, options...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runModelServiceListWorkspaceEmbeddingBackends(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("model list-workspace-embedding-backends", flag.ExitOnError)
	setSDKCommandUsage(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Models(wid).ListWorkspaceEmbeddingBackends(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runModelServiceWorkspaceEmbeddingRouter(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("model workspace-embedding-router", flag.ExitOnError)
	setSDKCommandUsage(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Models(wid).WorkspaceEmbeddingRouter(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runModelServiceListWorkspaceLLMModels(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("model list-workspace-llm-models", flag.ExitOnError)
	setSDKCommandUsage(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Models(wid).ListWorkspaceLLMModels(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runModelServiceGetWorkspaceEmbeddingBackend(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("model get-workspace-embedding-backend", flag.ExitOnError)
	setSDKCommandUsage(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	backendID := fs.Int64("backend-id", 0, "backend-id")
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Models(wid).GetWorkspaceEmbeddingBackend(context.Background(), *backendID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runModelServiceDeleteWorkspaceEmbeddingBackend(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("model delete-workspace-embedding-backend", flag.ExitOnError)
	setSDKCommandUsage(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	backendID := fs.Int64("backend-id", 0, "backend-id")
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	return ctx.Client.Models(wid).DeleteWorkspaceEmbeddingBackend(context.Background(), *backendID)
}

func runModelServiceListWorkspaceEmbeddingEndpoints(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("model list-workspace-embedding-endpoints", flag.ExitOnError)
	setSDKCommandUsage(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	backendID := fs.Int64("backend-id", 0, "backend-id")
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Models(wid).ListWorkspaceEmbeddingEndpoints(context.Background(), *backendID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runModelServiceListWorkspaceLLMEndpoints(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("model list-workspace-llm-endpoints", flag.ExitOnError)
	setSDKCommandUsage(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	backendID := fs.Int64("backend-id", 0, "backend-id")
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Models(wid).ListWorkspaceLLMEndpoints(context.Background(), *backendID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runModelServiceCreateWorkspaceEmbeddingBackend(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("model create-workspace-embedding-backend", flag.ExitOnError)
	setSDKCommandUsage(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	bodyJSON := fs.String("json", "", "Request JSON")
	bodyFile := fs.String("json-file", "", "Request JSON file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	var request sdk.WorkspaceEmbeddingBackendCreateRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Models(wid).CreateWorkspaceEmbeddingBackend(context.Background(), &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runModelServiceUpdateWorkspaceEmbeddingBackend(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("model update-workspace-embedding-backend", flag.ExitOnError)
	setSDKCommandUsage(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	backendID := fs.Int64("backend-id", 0, "backend-id")
	bodyJSON := fs.String("json", "", "Request JSON")
	bodyFile := fs.String("json-file", "", "Request JSON file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	var request sdk.WorkspaceEmbeddingBackendUpdateRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Models(wid).UpdateWorkspaceEmbeddingBackend(context.Background(), *backendID, &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runModelServiceUpdateWorkspaceEmbeddingRouter(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("model update-workspace-embedding-router", flag.ExitOnError)
	setSDKCommandUsage(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	bodyJSON := fs.String("json", "", "Request JSON")
	bodyFile := fs.String("json-file", "", "Request JSON file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	var request sdk.WorkspaceEmbeddingRouterUpdateRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Models(wid).UpdateWorkspaceEmbeddingRouter(context.Background(), &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runModelServiceCreateWorkspaceEmbeddingEndpoint(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("model create-workspace-embedding-endpoint", flag.ExitOnError)
	setSDKCommandUsage(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	backendID := fs.Int64("backend-id", 0, "backend-id")
	address := fs.String("address", "", "address")
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Models(wid).CreateWorkspaceEmbeddingEndpoint(context.Background(), *backendID, *address)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runModelServiceSetWorkspaceEmbeddingEndpointStatus(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("model set-workspace-embedding-endpoint-status", flag.ExitOnError)
	setSDKCommandUsage(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	backendID := fs.Int64("backend-id", 0, "backend-id")
	endpointID := fs.Int64("endpoint-id", 0, "endpoint-id")
	status := fs.Int("status", 0, "status")
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Models(wid).SetWorkspaceEmbeddingEndpointStatus(context.Background(), *backendID, *endpointID, int32(*status))
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runModelServiceSetWorkspaceLLMEndpointStatus(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("model set-workspace-llm-endpoint-status", flag.ExitOnError)
	setSDKCommandUsage(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	backendID := fs.Int64("backend-id", 0, "backend-id")
	endpointID := fs.Int64("endpoint-id", 0, "endpoint-id")
	status := fs.Int("status", 0, "status")
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	return ctx.Client.Models(wid).SetWorkspaceLLMEndpointStatus(context.Background(), *backendID, *endpointID, int32(*status))
}
