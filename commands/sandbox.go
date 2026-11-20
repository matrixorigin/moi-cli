package commands

import (
	"context"
	"flag"
	"fmt"
	"os"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
	"google.golang.org/protobuf/proto"
)

func ExecuteSandboxCommand(ctx *Context, args []string) error {
	if len(args) == 0 {
		ShowSandboxHelp(nil)
		return nil
	}
	switch args[0] {
	case "create", "update-network", "update-quota", "admin-update-network", "admin-update-quota":
		return runSandboxBody(ctx, args[0], args[1:])
	case "list":
		return runSandboxList(ctx, args[1:])
	case "get":
		return runSandboxGet(ctx, args[1:])
	case "delete":
		return runSandboxDelete(ctx, args[1:])
	case "get-quota":
		return runSandboxGetQuota(ctx, args[1:])
	case "terminal-token":
		return runSandboxTerminalToken(ctx, args[1:])
	case "connect-terminal":
		return runSandboxConnectTerminal(ctx, args[1:])
	case "create-ssh-session":
		return runSandboxCreateSSHSession(ctx, args[1:])
	case "admin-list":
		return runSandboxAdminList(ctx, args[1:])
	case "admin-get":
		return runSandboxAdminGet(ctx, args[1:])
	case "admin-delete":
		return runSandboxAdminDelete(ctx, args[1:])
	default:
		return fmt.Errorf("unknown sandbox subcommand: %s", args[0])
	}
}

func ShowSandboxHelp(_ []string) {
	fmt.Fprintln(os.Stderr, `Usage: moi-cli sandbox <subcommand> [options]
create, list, get, delete, update-network
get-quota, update-quota, terminal-token, connect-terminal, create-ssh-session
admin-list, admin-get, admin-delete, admin-update-network, admin-update-quota

Body commands accept --json / --json-file; inspect moi-cli schema sandbox COMMAND.
Admin commands use the same API key and the server's platform admission.
--workspace-id selects request context; admin-update-quota requires --target-workspace-id.
terminal-token and create-ssh-session return connection credentials; they do not open a terminal.
connect-terminal verifies the signed-token WebSocket handshake and then closes it.`)
}

func runSandboxBody(ctx *Context, operation string, args []string) error {
	fs := flag.NewFlagSet("sandbox "+operation, flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	inline := fs.String("json", "", "Typed request JSON")
	input := fs.String("json-file", "", "Typed request JSON file")
	var body proto.Message
	var service *sdk.SandboxService
	var invoke func() (any, error)
	switch operation {
	case "create":
		request := &sdk.SandboxCreateRequest{}
		body = request
		invoke = func() (any, error) { return service.Create(context.Background(), request) }
	case "update-network":
		id := fs.String("sandbox-id", "", "Sandbox ID")
		request := &sdk.SandboxNetworkRequest{}
		body = request
		invoke = func() (any, error) { return service.UpdateNetwork(context.Background(), *id, request) }
	case "admin-update-network":
		id := fs.String("sandbox-id", "", "Sandbox ID")
		request := &sdk.SandboxNetworkRequest{}
		body = request
		invoke = func() (any, error) { return service.AdminUpdateNetwork(context.Background(), *id, request) }
	case "update-quota":
		request := &sdk.SandboxQuotaRequest{}
		body = request
		invoke = func() (any, error) { return service.UpdateQuota(context.Background(), request) }
	case "admin-update-quota":
		targetWorkspaceID := fs.String("target-workspace-id", "", "Target workspace ID (distinct from request context)")
		request := &sdk.SandboxQuotaRequest{}
		body = request
		invoke = func() (any, error) {
			return service.AdminUpdateQuota(context.Background(), *targetWorkspaceID, request)
		}
	default:
		return fmt.Errorf("unknown sandbox body command: %s", operation)
	}
	return runWorkspaceRequest(ctx, fs, args, workspaceID, func(wid string) (any, error) {
		raw, err := loadJSONBytes(*inline, *input)
		if err != nil {
			return nil, err
		}
		if err := unmarshalProductJSON(raw, body); err != nil {
			return nil, fmt.Errorf("parse --json: %w", err)
		}
		service = ctx.Client.Sandboxes(wid)
		return invoke()
	})
}

func runSandboxList(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sandbox list", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	return runWorkspaceRequest(ctx, fs, args, workspaceID, func(wid string) (any, error) {
		service := ctx.Client.Sandboxes(wid)
		return service.List(context.Background())
	})
}

func runSandboxGet(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sandbox get", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	id := fs.String("sandbox-id", "", "Sandbox ID")
	return runWorkspaceRequest(ctx, fs, args, workspaceID, func(wid string) (any, error) {
		service := ctx.Client.Sandboxes(wid)
		return service.Get(context.Background(), *id)
	})
}

func runSandboxDelete(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sandbox delete", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	id := fs.String("sandbox-id", "", "Sandbox ID")
	return runWorkspaceRequest(ctx, fs, args, workspaceID, func(wid string) (any, error) {
		service := ctx.Client.Sandboxes(wid)
		return nil, service.Delete(context.Background(), *id)
	})
}

func runSandboxGetQuota(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sandbox get-quota", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	var query queryParameters
	fs.Var(&query, "query-param", "Query name=value (repeatable); include_network_capabilities=true requests network capability metadata")
	return runWorkspaceRequest(ctx, fs, args, workspaceID, func(wid string) (any, error) {
		service := ctx.Client.Sandboxes(wid)
		return service.GetQuota(context.Background(), query...)
	})
}

func runSandboxTerminalToken(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sandbox terminal-token", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	id := fs.String("sandbox-id", "", "Sandbox ID")
	return runWorkspaceRequest(ctx, fs, args, workspaceID, func(wid string) (any, error) {
		service := ctx.Client.Sandboxes(wid)
		return service.TerminalToken(context.Background(), *id)
	})
}

func runSandboxConnectTerminal(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sandbox connect-terminal", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	id := fs.String("sandbox-id", "", "Sandbox ID")
	token := fs.String("token", "", "Short-lived token returned by sandbox terminal-token")
	cols := fs.Uint("cols", 80, "Initial terminal columns")
	rows := fs.Uint("rows", 24, "Initial terminal rows")
	return runWorkspaceRequest(ctx, fs, args, workspaceID, func(wid string) (any, error) {
		if uint64(*cols) > uint64(^uint32(0)) || uint64(*rows) > uint64(^uint32(0)) {
			return nil, fmt.Errorf("cols and rows must fit uint32")
		}
		terminal, err := ctx.Client.Sandboxes(wid).ConnectTerminal(context.Background(), *id, *token, uint32(*cols), uint32(*rows))
		if err != nil {
			return nil, err
		}
		if err := terminal.Close(); err != nil {
			return nil, err
		}
		return map[string]any{"connected": true}, nil
	})
}

func runSandboxCreateSSHSession(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sandbox create-ssh-session", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	id := fs.String("sandbox-id", "", "Sandbox ID")
	return runWorkspaceRequest(ctx, fs, args, workspaceID, func(wid string) (any, error) {
		service := ctx.Client.Sandboxes(wid)
		return service.CreateSSHSession(context.Background(), *id)
	})
}

func runSandboxAdminList(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sandbox admin-list", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	var query queryParameters
	fs.Var(&query, "query-param", "Query name=value (repeatable); workspace_id, limit, offset")
	return runWorkspaceRequest(ctx, fs, args, workspaceID, func(wid string) (any, error) {
		service := ctx.Client.Sandboxes(wid)
		return service.AdminList(context.Background(), query...)
	})
}

func runSandboxAdminGet(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sandbox admin-get", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	id := fs.String("sandbox-id", "", "Sandbox ID")
	return runWorkspaceRequest(ctx, fs, args, workspaceID, func(wid string) (any, error) {
		service := ctx.Client.Sandboxes(wid)
		return service.AdminGet(context.Background(), *id)
	})
}

func runSandboxAdminDelete(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sandbox admin-delete", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	id := fs.String("sandbox-id", "", "Sandbox ID")
	return runWorkspaceRequest(ctx, fs, args, workspaceID, func(wid string) (any, error) {
		service := ctx.Client.Sandboxes(wid)
		return nil, service.AdminDelete(context.Background(), *id)
	})
}
