package commands

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func runAgentServiceAdminFeedbackTicket(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent admin-feedback-ticket", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	ticketID := fs.String("ticket-id", "", "Ticket ID")
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, "Usage: moi-cli agent admin-feedback-ticket [options]\n\nProduct SDK: AgentService.AdminFeedbackTicket\n\nOptions:\n")
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
	result, err := ctx.Client.Agents(wid).AdminFeedbackTicket(context.Background(), *ticketID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceBuildFeedbackGithubIssueDraft(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent build-feedback-github-issue-draft", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	ticketID := fs.String("ticket-id", "", "Ticket ID")
	kind := fs.String("kind", "", "Issue kind: bug, feature, or auto")
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, "Usage: moi-cli agent build-feedback-github-issue-draft [options]\n\nProduct SDK: AgentService.BuildFeedbackGithubIssueDraft\n\nOptions:\n")
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
	result, err := ctx.Client.Agents(wid).BuildFeedbackGithubIssueDraft(context.Background(), *ticketID, *kind)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceCreateFeedbackGithubIssue(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent create-feedback-github-issue", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	ticketID := fs.String("ticket-id", "", "Ticket ID")
	kind := fs.String("kind", "", "Issue kind: bug, feature, or auto")
	title := fs.String("title", "", "Issue title")
	body := fs.String("body", "", "Issue body")
	retry := fs.Bool("retry", false, "Retry a failed issue creation")
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, "Usage: moi-cli agent create-feedback-github-issue [options]\n\nProduct SDK: AgentService.CreateFeedbackGithubIssue\n\nOptions:\n")
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
	result, err := ctx.Client.Agents(wid).CreateFeedbackGithubIssue(context.Background(), *ticketID, *kind, *title, *body, *retry)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runAgentServiceApproveFeedbackReward(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("agent approve-feedback-reward", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	ticketID := fs.String("ticket-id", "", "Ticket ID")
	bodyJSON := fs.String("json", "", "Approval update JSON")
	bodyFile := fs.String("json-file", "", "Approval update JSON file")
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, "Usage: moi-cli agent approve-feedback-reward [options]\n\nProduct SDK: AgentService.ApproveFeedbackReward\n\nOptions:\n")
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
	rawBody, err := loadJSONBytes(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	var request sdk.FeedbackTicketAdminUpdateRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Agents(wid).ApproveFeedbackReward(context.Background(), *ticketID, &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}
