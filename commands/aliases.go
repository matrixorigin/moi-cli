package commands

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func runWorkspaceEnsureAlias(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workspace ensure", flag.ExitOnError)
	name := fs.String("name", "", "Workspace name")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workspace ensure --name <name>\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *name == "" {
		return fmt.Errorf("--name is required")
	}
	ws, err := ctx.Client.EnsureWorkspace(context.Background(), *name)
	if err != nil {
		return err
	}
	out := map[string]any{"id": ws.ID(), "name": ws.Name()}
	return printJSONOrTable(ctx, out, func() error {
		fmt.Printf("Workspace:\n")
		printKV("id", ws.ID())
		printKV("name", ws.Name())
		return nil
	})
}

func runWorkspaceGetAlias(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workspace get", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workspace get --workspace-id <id>\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	ws, err := ctx.Client.Workspace(wid)
	if err != nil {
		return err
	}
	out := map[string]any{"id": ws.ID(), "name": ws.Name()}
	return printJSONOrTable(ctx, out, func() error {
		fmt.Printf("Workspace:\n")
		printKV("id", ws.ID())
		printKV("name", ws.Name())
		return nil
	})
}

func runWorkspaceGetDBConnectionAlias(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workspace get-db-connection", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workspace get-db-connection [--workspace-id <id>]\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	ws, err := ctx.Client.Workspace(wid)
	if err != nil {
		return err
	}
	conn, err := ws.DBConnection(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, conn, func() error { return printJSON(conn) })
}

func runCatalogGetAlias(ctx *Context, args []string) error {
	return runCatalogServiceInfo(ctx, args)
}

func runCatalogPrepareAlias(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("catalog prepare", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	name := fs.String("name", "", "Catalog name")
	description := fs.String("description", "", "Description")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli catalog prepare --name <name>\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if *name == "" {
		return fmt.Errorf("--name is required")
	}
	ws, err := ctx.Client.Workspace(wid)
	if err != nil {
		return err
	}
	var opts []sdk.CatalogInfoOption
	if *description != "" {
		opts = append(opts, sdk.WithCatalogDescription(*description))
	}
	handle, err := ws.PrepareCatalog(context.Background(), *name, opts...)
	if err != nil {
		return err
	}
	out := map[string]any{"id": handle.ID(), "name": handle.Name()}
	return printJSONOrTable(ctx, out, func() error {
		fmt.Printf("Catalog:\n")
		printKV("id", fmt.Sprintf("%d", handle.ID()))
		printKV("name", handle.Name())
		return nil
	})
}

func runSQLExecuteAlias(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sql execute", flag.ExitOnError)
	addSQLQueryFlags(fs)
	return runSQLStatement(ctx, fs, args, false)
}

func runSQLRunQuery(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("sql run", flag.ExitOnError)
	addSQLQueryFlags(fs)
	return runSQLStatement(ctx, fs, args, true)
}

func runSQLStatement(ctx *Context, fs *flag.FlagSet, args []string, returnResult bool) error {
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	query := fs.String("query", "", "SQL query")
	queryFile := fs.String("query-file", "", "SQL file")
	fs.String("db", "", "Database name")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli %s --query <sql>\n", fs.Name())
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	sqlText := *query
	if *queryFile != "" {
		data, err := os.ReadFile(*queryFile)
		if err != nil {
			return err
		}
		sqlText = string(data)
	}
	if strings.TrimSpace(sqlText) == "" {
		return fmt.Errorf("--query or --query-file is required")
	}
	opts := sqlQueryCLIOptions(fs, "db")
	if returnResult {
		result, err := ctx.Client.SQL(wid).RunQuery(context.Background(), sqlText, opts...)
		if err != nil {
			return err
		}
		if err := printJSONOrTable(ctx, result, func() error { return printJSON(result) }); err != nil {
			return err
		}
		if result.GetStatus() == "failed" || result.GetStatus() == "killed" {
			return fmt.Errorf("%s: %s", result.GetStatus(), result.GetErrMsg())
		}
		return nil
	}
	result, err := ctx.Client.SQL(wid).ExecuteQuery(context.Background(), sqlText, opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runSQLDescribeAlias(ctx *Context, args []string) error {
	return runSQLServiceDescribeQuery(ctx, args)
}

func runSQLKillAlias(ctx *Context, args []string) error {
	return runSQLServiceKillQuery(ctx, args)
}

func runSQLResultAlias(ctx *Context, args []string) error {
	return runSQLServiceQueryResult(ctx, args)
}

func runSQLHistoryAlias(ctx *Context, args []string) error {
	return runSQLServiceQueryHistory(ctx, args)
}

func runSQLMetaTablesAlias(ctx *Context, args []string) error {
	return runSQLServiceMetaDBTables(ctx, remapFlag(args, "db", "db-name"))
}

func runAgentSendAlias(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent send", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	agentCode := fs.String("agent-code", "", "Agent code")
	agentID := fs.String("agent-id", "", "Agent ID")
	text := fs.String("text", "", "User text")
	requestID := fs.String("request-id", "", "Request id")
	messageID := fs.String("message-id", "", "Message id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent send --text <text> (--agent-code|--agent-id)\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*text) == "" {
		return fmt.Errorf("--text is required")
	}
	ws, err := ctx.Client.Workspace(wid)
	if err != nil {
		return err
	}
	var agent *sdk.AgentHandle
	switch {
	case *agentCode != "":
		agent, err = ws.AgentByCode(*agentCode)
	case *agentID != "":
		agent, err = ws.AgentByID(*agentID)
	default:
		return fmt.Errorf("--agent-code or --agent-id is required")
	}
	if err != nil {
		return err
	}
	msg := sdk.AgentTextMessage{Text: *text, RequestID: *requestID, MessageID: *messageID}
	result, err := agent.SendText(context.Background(), msg)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentGetTaskAlias(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent get-task", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	agentCode := fs.String("agent-code", "", "Agent code")
	agentID := fs.String("agent-id", "", "Agent ID")
	taskID := fs.String("task-id", "", "Task id")
	requestID := fs.String("request-id", "", "Request id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent get-task --task-id <id> (--agent-code|--agent-id)\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if *taskID == "" {
		return fmt.Errorf("--task-id is required")
	}
	ws, err := ctx.Client.Workspace(wid)
	if err != nil {
		return err
	}
	var agent *sdk.AgentHandle
	switch {
	case *agentCode != "":
		agent, err = ws.AgentByCode(*agentCode)
	case *agentID != "":
		agent, err = ws.AgentByID(*agentID)
	default:
		return fmt.Errorf("--agent-code or --agent-id is required")
	}
	if err != nil {
		return err
	}
	req := sdk.AgentTaskRequest{TaskID: *taskID, RequestID: *requestID}
	result, err := agent.GetTask(context.Background(), req)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentCancelTaskAlias(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent cancel-task", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	agentCode := fs.String("agent-code", "", "Agent code")
	agentID := fs.String("agent-id", "", "Agent ID")
	taskID := fs.String("task-id", "", "Task id")
	requestID := fs.String("request-id", "", "Request id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent cancel-task --task-id <id> (--agent-code|--agent-id)\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if *taskID == "" {
		return fmt.Errorf("--task-id is required")
	}
	ws, err := ctx.Client.Workspace(wid)
	if err != nil {
		return err
	}
	var agent *sdk.AgentHandle
	switch {
	case *agentCode != "":
		agent, err = ws.AgentByCode(*agentCode)
	case *agentID != "":
		agent, err = ws.AgentByID(*agentID)
	default:
		return fmt.Errorf("--agent-code or --agent-id is required")
	}
	if err != nil {
		return err
	}
	req := sdk.AgentTaskRequest{TaskID: *taskID, RequestID: *requestID}
	result, err := agent.CancelTask(context.Background(), req)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentCardAlias(ctx *Context, args []string) error {
	return runAgentServiceAgentCard(ctx, args)
}

func remapFlag(args []string, from, to string) []string {
	out := make([]string, 0, len(args))
	prefix := "--" + from
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == prefix && i+1 < len(args) {
			out = append(out, "--"+to, args[i+1])
			i++
			continue
		}
		if strings.HasPrefix(a, prefix+"=") {
			out = append(out, "--"+to+"="+strings.TrimPrefix(a, prefix+"="))
			continue
		}
		out = append(out, a)
	}
	return out
}
