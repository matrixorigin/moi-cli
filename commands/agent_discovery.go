package commands

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func runAgentServiceList(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent list", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	catalog := fs.String("catalog", "", "Agent catalog: workspace or system")
	status := fs.String("status", "", "Agent status filter")
	query := fs.String("query", "", "Search agents")
	surface := fs.String("surface", "", "Agent list surface")
	includeDevelopment := fs.Bool("include-development", false, "Include development Agents")
	limit := fs.Int("limit", 0, "Page size")
	offset := fs.Int("offset", 0, "Page offset")
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, "Usage: moi-cli agent list [options]\n\nProduct SDK: AgentService.List\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	var options []sdk.QueryOption
	fs.Visit(func(current *flag.Flag) {
		switch current.Name {
		case "catalog":
			options = append(options, sdk.WithQueryParam("catalog", *catalog))
		case "status":
			options = append(options, sdk.WithQueryParam("status", *status))
		case "query":
			options = append(options, sdk.WithQueryParam("query", *query))
		case "surface":
			options = append(options, sdk.WithQueryParam("surface", *surface))
		case "include-development":
			options = append(options, sdk.WithQueryParam("include_development", fmt.Sprint(*includeDevelopment)))
		case "limit":
			options = append(options, sdk.WithQueryParam("limit", fmt.Sprint(*limit)))
		case "offset":
			options = append(options, sdk.WithQueryParam("offset", fmt.Sprint(*offset)))
		}
	})
	result, err := ctx.Client.Agents(wid).List(context.Background(), options...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceExecutorHealth(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent executor-health", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, "Usage: moi-cli agent executor-health [options]\n\nProduct SDK: AgentService.ExecutorHealth\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Agents(wid).ExecutorHealth(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServicePreviewQueryVisual(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent preview-query-visual", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	agentID := fs.String("agent-id", "", "Agent ID")
	fileID := fs.String("file-id", "", "Query visual file ID")
	output := fs.String("output", "", "Write image to file (default stdout)")
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, "Usage: moi-cli agent preview-query-visual [options]\n\nProduct SDK: AgentService.PreviewQueryVisual\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Agents(wid).PreviewQueryVisual(context.Background(), *agentID, *fileID)
	if err != nil {
		return err
	}
	defer result.Body.Close()
	return writeStream(result.Body, *output)
}
