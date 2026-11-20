package commands

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func runAgentServiceGetSkillCandidateConfirmation(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent get-skill-candidate-confirmation", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	taskID := fs.String("task-id", "", "Runtime task ID")
	callID := fs.String("call-id", "", "Skill Creator tool call ID")
	sourceDigest := fs.String("source-digest", "", "Reviewed candidate source digest")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent get-skill-candidate-confirmation --task-id <id> --call-id <id> --source-digest <digest> [options]\n\nProduct SDK: AgentService.GetSkillCandidateConfirmation\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	if *taskID == "" || *callID == "" || *sourceDigest == "" {
		return fmt.Errorf("--task-id, --call-id and --source-digest are required")
	}
	result, err := ctx.Client.Agents(wid).GetSkillCandidateConfirmation(context.Background(), *taskID, *callID, *sourceDigest)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceConfirmSkillCandidate(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent confirm-skill-candidate", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	taskID := fs.String("task-id", "", "Runtime task ID")
	bodyJSON := fs.String("json", "", "JSON for the reviewed skill candidate")
	bodyFile := fs.String("json-file", "", "JSON file for the reviewed skill candidate")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli agent confirm-skill-candidate --task-id <id> [options]\n\nProduct SDK: AgentService.ConfirmSkillCandidate\n\nOptions:\n")
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
	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	var request sdk.ConfirmSkillCandidateRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Agents(wid).ConfirmSkillCandidate(context.Background(), *taskID, &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}
