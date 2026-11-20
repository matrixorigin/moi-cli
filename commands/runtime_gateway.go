package commands

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func ExecuteRuntimeGatewayCommand(ctx *Context, args []string) error {
	if len(args) == 0 {
		ShowRuntimeGatewayHelp(nil)
		return nil
	}
	sub := args[0]
	switch sub {
	case "invoke-model":
		return runRuntimeGatewayServiceInvokeModel(ctx, args[1:])

	case "apply-compute-binding-event":
		return runRuntimeGatewayServiceApplyComputeBindingEvent(ctx, args[1:])

	case "authorize-executor":
		return runRuntimeGatewayServiceAuthorizeExecutor(ctx, args[1:])

	case "resolve-model":
		return runRuntimeGatewayServiceResolveModel(ctx, args[1:])

	case "mcp":
		return runRuntimeGatewayServiceMCP(ctx, args[1:])

	case "renew-edge-token":
		return runRuntimeGatewayServiceRenewEdgeToken(ctx, args[1:])

	case "check-edge-token":
		return runRuntimeGatewayServiceCheckEdgeToken(ctx, args[1:])
	default:
		return fmt.Errorf("unknown runtime-gateway subcommand: %s", sub)
	}
}
func ShowRuntimeGatewayHelp(_ []string) {
	fmt.Fprintln(os.Stderr, "Usage: moi-cli runtime-gateway <subcommand> [options]\ninvoke-model\napply-compute-binding-event\nauthorize-executor\nresolve-model\nmcp\nrenew-edge-token\ncheck-edge-token")
}

func runRuntimeGatewayServiceCheckEdgeToken(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("runtime-gateway check-edge-token", flag.ExitOnError)
	setSDKCommandUsage(fs)
	token := fs.String("token", "", "token")
	grantFile := fs.String("runtime-token-file", "", "Runtime grant file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	client, err := runtimeGatewayCLIClient(ctx, *grantFile)
	if err != nil {
		return err
	}
	if client != ctx.Client {
		defer client.Close()
	}
	result, err := client.RuntimeGateway().CheckEdgeToken(context.Background(), *token)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runRuntimeGatewayServiceRenewEdgeToken(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("runtime-gateway renew-edge-token", flag.ExitOnError)
	setSDKCommandUsage(fs)
	token := fs.String("token", "", "token")
	grantFile := fs.String("runtime-token-file", "", "Runtime grant file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	client, err := runtimeGatewayCLIClient(ctx, *grantFile)
	if err != nil {
		return err
	}
	if client != ctx.Client {
		defer client.Close()
	}
	result, err := client.RuntimeGateway().RenewEdgeToken(context.Background(), *token)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runRuntimeGatewayServiceMCP(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("runtime-gateway mcp", flag.ExitOnError)
	setSDKCommandUsage(fs)
	bodyJSON := fs.String("json", "", "Request JSON")
	bodyFile := fs.String("json-file", "", "Request JSON file")
	grantFile := fs.String("runtime-token-file", "", "Runtime grant file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	var request sdk.RuntimeRPCRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	client, err := runtimeGatewayCLIClient(ctx, *grantFile)
	if err != nil {
		return err
	}
	if client != ctx.Client {
		defer client.Close()
	}
	result, err := client.RuntimeGateway().MCP(context.Background(), &request)
	if err != nil {
		return err
	}
	if result == nil {
		// A cancellation notification has an empty HTTP 202 acknowledgement,
		// not a JSON-RPC result to render.
		return nil
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runRuntimeGatewayServiceResolveModel(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("runtime-gateway resolve-model", flag.ExitOnError)
	setSDKCommandUsage(fs)
	bodyJSON := fs.String("json", "", "Request JSON")
	bodyFile := fs.String("json-file", "", "Request JSON file")
	grantFile := fs.String("runtime-token-file", "", "Runtime grant file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	var request sdk.RuntimeModelResolveRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	client, err := runtimeGatewayCLIClient(ctx, *grantFile)
	if err != nil {
		return err
	}
	if client != ctx.Client {
		defer client.Close()
	}
	result, err := client.RuntimeGateway().ResolveModel(context.Background(), &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runRuntimeGatewayServiceAuthorizeExecutor(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("runtime-gateway authorize-executor", flag.ExitOnError)
	setSDKCommandUsage(fs)
	bodyJSON := fs.String("json", "", "Request JSON")
	bodyFile := fs.String("json-file", "", "Request JSON file")
	grantFile := fs.String("runtime-token-file", "", "Runtime grant file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	var request sdk.RuntimeExecutorAuthorizationRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	client, err := runtimeGatewayCLIClient(ctx, *grantFile)
	if err != nil {
		return err
	}
	if client != ctx.Client {
		defer client.Close()
	}
	return client.RuntimeGateway().AuthorizeExecutor(context.Background(), &request)
}

func runRuntimeGatewayServiceApplyComputeBindingEvent(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("runtime-gateway apply-compute-binding-event", flag.ExitOnError)
	setSDKCommandUsage(fs)
	bodyJSON := fs.String("json", "", "Request JSON")
	bodyFile := fs.String("json-file", "", "Request JSON file")
	grantFile := fs.String("runtime-token-file", "", "Runtime grant file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	var request sdk.ComputeBindingEvent
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	client, err := runtimeGatewayCLIClient(ctx, *grantFile)
	if err != nil {
		return err
	}
	if client != ctx.Client {
		defer client.Close()
	}
	result, err := client.RuntimeGateway().ApplyComputeBindingEvent(context.Background(), &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runRuntimeGatewayServiceInvokeModel(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("runtime-gateway invoke-model", flag.ExitOnError)
	setSDKCommandUsage(fs)
	bodyJSON := fs.String("json", "", "Request JSON")
	bodyFile := fs.String("json-file", "", "Request JSON file")
	output := fs.String("output", "", "Output file, or stdout")
	grantFile := fs.String("runtime-token-file", "", "Runtime grant file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	var request sdk.RuntimeModelChatRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	client, err := runtimeGatewayCLIClient(ctx, *grantFile)
	if err != nil {
		return err
	}
	if client != ctx.Client {
		defer client.Close()
	}
	result, err := client.RuntimeGateway().InvokeModel(context.Background(), &request)
	if err != nil {
		return err
	}
	defer result.Body.Close()
	return writeStream(result.Body, *output)
}

func runtimeGatewayCLIClient(ctx *Context, grantFile string) (*sdk.Client, error) {
	if grantFile == "" {
		return ctx.Client, nil
	}
	raw, err := os.ReadFile(grantFile)
	if err != nil {
		return nil, err
	}
	return sdk.New(ctx.Endpoint, sdk.WithToken(strings.TrimSpace(string(raw))))
}
