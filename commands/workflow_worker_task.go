package commands

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func runWorkflowServiceCreateWorkerTask(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow create-worker-task", flag.ExitOnError)
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
	if len(rawBody) == 0 {
		return sdk.ErrSpecRequired
	}
	var request sdk.WorkflowWorkerTaskCreateRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Workflows(wid).CreateWorkerTask(context.Background(), &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}
