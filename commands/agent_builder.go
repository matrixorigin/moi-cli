package commands

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func runAgentResourceMetadata(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent list-resource-meta", flag.ContinueOnError)
	setSDKCommandUsage(fs)
	output := fs.String("output-file", "", "Write all selectable resource metadata to this JSON file")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if *output == "" || fs.NArg() != 0 {
		return fmt.Errorf("usage: moi-cli agent list-resource-meta --output-file <resources.json>")
	}
	data, err := agentBuilderRequest(ctx, "resources", nil)
	if err != nil {
		return err
	}
	groups := []string{"skills", "tools", "knowledge_bases", "models"}
	counts := make(map[string]int, len(groups))
	var encoded bytes.Buffer
	encoded.WriteString("{\n  \"schema_version\": 1")
	for _, group := range groups {
		items, ok := data[group].([]any)
		if !ok {
			return &sdk.Error{Code: "CLI_RESOURCE_METADATA_INVALID", Message: "resource metadata response must contain skills, tools, knowledge_bases and models arrays"}
		}
		counts[group] = len(items)
		fmt.Fprintf(&encoded, ",\n  %q: [", group)
		for index, item := range items {
			if _, ok := item.(map[string]any); !ok {
				return &sdk.Error{Code: "CLI_RESOURCE_METADATA_INVALID", Message: "resource metadata must contain JSON objects"}
			}
			raw, err := json.Marshal(item)
			if err != nil {
				return err
			}
			if index > 0 {
				encoded.WriteByte(',')
			}
			encoded.WriteString("\n    ")
			encoded.Write(raw)
		}
		if len(items) > 0 {
			encoded.WriteString("\n  ")
		}
		encoded.WriteByte(']')
	}
	encoded.WriteString("\n}\n")
	target, err := filepath.Abs(*output)
	if err != nil {
		return err
	}
	if err := writeCommandResultFile(target, encoded.Bytes(), true); err != nil {
		return err
	}
	return printJSON(map[string]any{"complete": true, "output_file": target, "counts": counts})
}

func runAgentBuilder(ctx *Context, args []string) error {
	if len(args) == 0 || args[0] == "--help" || args[0] == "-h" {
		fmt.Fprintln(os.Stderr, "Usage: moi-cli agent builder <context|propose|workspace-mcp> [options]\n\ncontext saves trusted editing facts; propose submits a candidate; workspace-mcp prepares a selected published capability.")
		return nil
	}
	switch args[0] {
	case "context":
		return runAgentBuilderContext(ctx, args[1:])
	case "propose":
		return runAgentBuilderPropose(ctx, args[1:])
	case "workspace-mcp":
		return runAgentBuilderWorkspaceMCP(ctx, args[1:])
	default:
		return fmt.Errorf("unknown agent builder subcommand: %s", args[0])
	}
}

func runAgentBuilderContext(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent builder context", flag.ContinueOnError)
	setSDKCommandUsage(fs)
	output := fs.String("output-file", "", "Write trusted editing context to this JSON file")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if *output == "" || fs.NArg() != 0 {
		return fmt.Errorf("usage: moi-cli agent builder context --output-file <context.json>")
	}
	data, err := agentBuilderRequest(ctx, "context", nil)
	if err != nil {
		return err
	}
	raw, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	target, err := filepath.Abs(*output)
	if err != nil {
		return err
	}
	if err := writeCommandResultFile(target, append(raw, '\n'), true); err != nil {
		return err
	}
	return printJSON(map[string]any{"complete": true, "output_file": target, "operations": data["operations"]})
}

func runAgentBuilderPropose(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent builder propose", flag.ContinueOnError)
	setSDKCommandUsage(fs)
	input := fs.String("json-file", "", "JSON object containing config or changes")
	output := fs.String("output-file", "", "Optionally save the complete server result to a JSON file")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments")
	}
	return submitAgentBuilderFile(ctx, "propose", *input, *output)
}

func runAgentBuilderWorkspaceMCP(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent builder workspace-mcp", flag.ContinueOnError)
	setSDKCommandUsage(fs)
	input := fs.String("json-file", "", "Existing published capability preparation request JSON")
	output := fs.String("output-file", "", "Optionally save the complete server result to a JSON file")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments")
	}
	return submitAgentBuilderFile(ctx, "workspace_mcp", *input, *output)
}

func submitAgentBuilderFile(ctx *Context, operation, input, output string) error {
	if input == "" {
		return fmt.Errorf("--json-file is required")
	}
	rawBody, err := os.ReadFile(input)
	if err != nil {
		return err
	}
	if !json.Valid(rawBody) {
		return fmt.Errorf("--json-file must contain valid JSON")
	}
	// Keep the source JSON intact. The owner validates keys, duplicates and values.
	data, err := agentBuilderRequest(ctx, operation, json.RawMessage(rawBody))
	if err != nil {
		return err
	}
	if output != "" {
		raw, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return err
		}
		if err := writeCommandResultFile(output, append(raw, '\n'), true); err != nil {
			return err
		}
	}
	if operation == "workspace_mcp" {
		return printJSON(data)
	}
	// The card owns the full prompt and bindings. Keep terminal feedback compact.
	receipt := make(map[string]any)
	for _, key := range []string{"ok", "call_id", "task_id", "agent_id", "candidate_version", "load_version", "source_digest", "mode", "status", "message", "prompt_length", "diagnostics"} {
		if value, ok := data[key]; ok {
			receipt[key] = value
		}
	}
	if output != "" {
		receipt["output_file"] = output
	}
	return printJSON(receipt)
}

func agentBuilderRequest(ctx *Context, operation string, input any) (map[string]any, error) {
	workspace, err := resolveWorkspaceID(ctx, "")
	if err != nil {
		return nil, err
	}
	command := map[string]any{"operation": operation}
	if input != nil {
		command["input"] = input
	}
	result, err := ctx.Client.Agents(workspace).ResourcePost(context.Background(), "agent-builder", "/runtime", command)
	if err != nil {
		return nil, err
	}
	data, ok := result.AsMap()
	if !ok {
		return nil, &sdk.Error{Code: "CLI_AGENT_BUILDER_RESPONSE_INVALID", Message: "agent builder response must be a JSON object"}
	}
	switch operation {
	case "context":
		operations, ok := data["operations"].(map[string]any)
		if !ok || data["schema_version"] != json.Number("1") {
			return nil, invalidAgentBuilderResponse()
		}
		for _, action := range []string{"create", "revise_current"} {
			item, ok := operations[action].(map[string]any)
			if !ok {
				return nil, invalidAgentBuilderResponse()
			}
			if _, ok := item["allowed"].(bool); !ok {
				return nil, invalidAgentBuilderResponse()
			}
		}
	case "propose":
		if data["render_type"] != "agent_candidate" || data["status"] != "valid" {
			return nil, invalidAgentBuilderResponse()
		}
		for _, key := range []string{"agent_id", "candidate_version", "source_digest"} {
			if value, ok := data[key].(string); !ok || value == "" {
				return nil, invalidAgentBuilderResponse()
			}
		}
	}
	return data, nil
}

func invalidAgentBuilderResponse() error {
	return &sdk.Error{Code: "CLI_AGENT_BUILDER_RESPONSE_INVALID", Message: "agent builder response is missing required context or candidate fields"}
}
