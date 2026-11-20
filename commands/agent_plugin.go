package commands

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func agentPluginQueryOptions(fs *flag.FlagSet, names ...string) []sdk.QueryOption {
	selected := make(map[string]bool, len(names))
	for _, name := range names {
		selected[name] = true
	}
	var options []sdk.QueryOption
	fs.Visit(func(current *flag.Flag) {
		if selected[current.Name] {
			options = append(options, sdk.WithQueryParam(current.Name, current.Value.String()))
		}
	})
	return options
}

func runAgentServiceSearchPlugins(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent search-plugins", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.String("q", "", "q query option")
	fs.String("availability", "", "availability query option")
	fs.String("cursor", "", "cursor query option")
	fs.String("limit", "", "limit query option")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent search-plugins [options]\n\nProduct SDK: AgentService.SearchPlugins\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	opts := agentPluginQueryOptions(fs, "q", "availability", "cursor", "limit")
	result, err := ctx.Client.Agents(wid).SearchPlugins(context.Background(), opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceDescribePluginCapability(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent describe-plugin-capability", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	ref := fs.String("ref", "", "ref")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent describe-plugin-capability [options]\n\nProduct SDK: AgentService.DescribePluginCapability\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Agents(wid).DescribePluginCapability(context.Background(), *ref)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceListPluginReleases(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent list-plugin-releases", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.String("namespace", "", "namespace query option")
	fs.String("q", "", "q query option")
	fs.String("limit", "", "limit query option")
	fs.String("offset", "", "offset query option")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent list-plugin-releases [options]\n\nProduct SDK: AgentService.ListPluginReleases\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	opts := agentPluginQueryOptions(fs, "namespace", "q", "limit", "offset")
	result, err := ctx.Client.Agents(wid).ListPluginReleases(context.Background(), opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceGetPluginRelease(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent get-plugin-release", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	releaseID := fs.String("release-id", "", "release-id")
	fs.String("namespace", "", "namespace query option")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent get-plugin-release [options]\n\nProduct SDK: AgentService.GetPluginRelease\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	opts := agentPluginQueryOptions(fs, "namespace")
	result, err := ctx.Client.Agents(wid).GetPluginRelease(context.Background(), *releaseID, opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceInstallPlugin(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent install-plugin", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	bodyJSON := fs.String("json", "", "Request JSON (secrets: prefer --json-file)")
	bodyFile := fs.String("json-file", "", "Request JSON file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent install-plugin [options]\n\nProduct SDK: AgentService.InstallPlugin\n\nOptions:\n")
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
	if len(rawBody) == 0 {
		return sdk.ErrSpecRequired
	}
	var req sdk.PluginInstallRequest
	if err := json.Unmarshal(rawBody, &req); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Agents(wid).InstallPlugin(context.Background(), &req)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceListPluginInstallations(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent list-plugin-installations", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.String("limit", "", "limit query option")
	fs.String("offset", "", "offset query option")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent list-plugin-installations [options]\n\nProduct SDK: AgentService.ListPluginInstallations\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	opts := agentPluginQueryOptions(fs, "limit", "offset")
	result, err := ctx.Client.Agents(wid).ListPluginInstallations(context.Background(), opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceGetPluginInstallation(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent get-plugin-installation", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	id := fs.String("id", "", "id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent get-plugin-installation [options]\n\nProduct SDK: AgentService.GetPluginInstallation\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Agents(wid).GetPluginInstallation(context.Background(), *id)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceUpdatePluginInstallation(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent update-plugin-installation", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	id := fs.String("id", "", "id")
	bodyJSON := fs.String("json", "", "Request JSON (secrets: prefer --json-file)")
	bodyFile := fs.String("json-file", "", "Request JSON file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent update-plugin-installation [options]\n\nProduct SDK: AgentService.UpdatePluginInstallation\n\nOptions:\n")
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
	if len(rawBody) == 0 {
		return sdk.ErrSpecRequired
	}
	var req sdk.PluginUpdateInstallationRequest
	if err := json.Unmarshal(rawBody, &req); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Agents(wid).UpdatePluginInstallation(context.Background(), *id, &req)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceDeletePluginInstallation(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent delete-plugin-installation", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	id := fs.String("id", "", "id")
	expectedVersion := fs.Int64("expected-version", 0, "expected-version")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent delete-plugin-installation [options]\n\nProduct SDK: AgentService.DeletePluginInstallation\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Agents(wid).DeletePluginInstallation(context.Background(), *id, *expectedVersion)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceCreatePluginConnection(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent create-plugin-connection", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	bodyJSON := fs.String("json", "", "Request JSON (secrets: prefer --json-file)")
	bodyFile := fs.String("json-file", "", "Request JSON file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent create-plugin-connection [options]\n\nProduct SDK: AgentService.CreatePluginConnection\n\nOptions:\n")
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
	if len(rawBody) == 0 {
		return sdk.ErrSpecRequired
	}
	var req sdk.PluginCreateConnectionRequest
	if err := json.Unmarshal(rawBody, &req); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Agents(wid).CreatePluginConnection(context.Background(), &req)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceListPluginConnections(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent list-plugin-connections", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.String("limit", "", "limit query option")
	fs.String("offset", "", "offset query option")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent list-plugin-connections [options]\n\nProduct SDK: AgentService.ListPluginConnections\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	opts := agentPluginQueryOptions(fs, "limit", "offset")
	result, err := ctx.Client.Agents(wid).ListPluginConnections(context.Background(), opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceGetPluginConnection(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent get-plugin-connection", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	id := fs.String("id", "", "id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent get-plugin-connection [options]\n\nProduct SDK: AgentService.GetPluginConnection\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Agents(wid).GetPluginConnection(context.Background(), *id)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceUpdatePluginConnection(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent update-plugin-connection", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	id := fs.String("id", "", "id")
	bodyJSON := fs.String("json", "", "Request JSON (secrets: prefer --json-file)")
	bodyFile := fs.String("json-file", "", "Request JSON file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent update-plugin-connection [options]\n\nProduct SDK: AgentService.UpdatePluginConnection\n\nOptions:\n")
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
	if len(rawBody) == 0 {
		return sdk.ErrSpecRequired
	}
	var req sdk.PluginUpdateConnectionRequest
	if err := json.Unmarshal(rawBody, &req); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Agents(wid).UpdatePluginConnection(context.Background(), *id, &req)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceDeletePluginConnection(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent delete-plugin-connection", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	id := fs.String("id", "", "id")
	expectedVersion := fs.Int("expected-version", 0, "expected-version")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent delete-plugin-connection [options]\n\nProduct SDK: AgentService.DeletePluginConnection\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Agents(wid).DeletePluginConnection(context.Background(), *id, *expectedVersion)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceGetPluginAuthorization(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent get-plugin-authorization", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	id := fs.String("id", "", "id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent get-plugin-authorization [options]\n\nProduct SDK: AgentService.GetPluginAuthorization\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Agents(wid).GetPluginAuthorization(context.Background(), *id)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceStorePluginSecret(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent store-plugin-secret", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	id := fs.String("id", "", "id")
	bodyJSON := fs.String("json", "", "Request JSON (secrets: prefer --json-file)")
	bodyFile := fs.String("json-file", "", "Request JSON file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent store-plugin-secret [options]\n\nProduct SDK: AgentService.StorePluginSecret\n\nOptions:\n")
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
	if len(rawBody) == 0 {
		return sdk.ErrSpecRequired
	}
	var req sdk.PluginStoreSecretRequest
	if err := json.Unmarshal(rawBody, &req); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Agents(wid).StorePluginSecret(context.Background(), *id, &req)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceBeginPluginOAuth(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent begin-plugin-oauth", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	id := fs.String("id", "", "id")
	expectedVersion := fs.Int("expected-version", 0, "expected-version")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent begin-plugin-oauth [options]\n\nProduct SDK: AgentService.BeginPluginOAuth\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Agents(wid).BeginPluginOAuth(context.Background(), *id, *expectedVersion)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServicePollPluginOAuth(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent poll-plugin-oauth", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	id := fs.String("id", "", "id")
	attemptID := fs.String("attempt-id", "", "attempt-id")
	expectedVersion := fs.Int("expected-version", 0, "expected-version")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent poll-plugin-oauth [options]\n\nProduct SDK: AgentService.PollPluginOAuth\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Agents(wid).PollPluginOAuth(context.Background(), *id, *attemptID, *expectedVersion)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceDisconnectPluginConnection(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent disconnect-plugin-connection", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	id := fs.String("id", "", "id")
	expectedGeneration := fs.Int64("expected-generation", 0, "expected-generation")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent disconnect-plugin-connection [options]\n\nProduct SDK: AgentService.DisconnectPluginConnection\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Agents(wid).DisconnectPluginConnection(context.Background(), *id, *expectedGeneration)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}
