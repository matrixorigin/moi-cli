package commands

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

func runWorkflowServiceAcceptCandidate(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow accept-candidate", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	taskID := fs.String("task-id", "", "authenticated A2A task ID")
	artifactID := fs.String("artifact-id", "", "server-owned workflow.candidate artifact ID")
	proposalID := fs.String("proposal-id", "", "server-owned pending proposal ID")
	idempotencyKey := fs.String("idempotency-key", "", "caller-owned acceptance idempotency key")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow accept-candidate [options]\n\nProduct SDK: WorkflowService.AcceptCandidate\n\nOutputs the complete server acceptance JSON for explicit review and handoff.\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	accepted, err := ctx.Client.Workflows(wid).AcceptCandidate(
		context.Background(),
		sdk.WorkflowCandidateRef{TaskID: *taskID, ArtifactID: *artifactID},
		&sdk.WorkflowCandidateConfirmation{ProposalId: *proposalID, IdempotencyKey: *idempotencyKey},
	)
	if err != nil {
		return err
	}
	return printJSON(accepted.Acceptance())
}

func runWorkflowServiceDeployCandidate(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow deploy-candidate", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	name := fs.String("name", "", "workflow name")
	acceptanceFile := fs.String("acceptance-file", "", "path to complete JSON emitted by workflow accept-candidate")
	workflowID := fs.String("workflow-id", "", "existing workflow ID")
	description := fs.String("description", "", "workflow description")
	executionMode := fs.String("execution-mode", "", "execution mode")
	cronExpression := fs.String("cron-expression", "", "Cron expression; see workflow update --help for syntax, timezone and examples")
	computeResourceID := fs.String("compute-resource-id", "", "compute resource ID")
	volumeTriggerFile := fs.String("volume-trigger-file", "", "Volume trigger configuration JSON file")
	dynamicServiceFile := fs.String("dynamic-service-file", "", "Dynamic service configuration JSON file")
	defaultValuesFile := fs.String("default-values-file", "", "Workflow default values JSON file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow deploy-candidate [options]\n\nProduct SDK: WorkflowService.DeployCandidate\n\nRequires an explicit server acceptance result; this command never accepts a candidate.\n\nOptions:\n")
		fs.PrintDefaults()
		printWorkflowScheduleGuidance(fs)
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	accepted, err := acceptedWorkflowCandidateFromFile(*acceptanceFile)
	if err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	opts := make([]sdk.WorkflowDeployOption, 0, 5)
	fs.Visit(func(value *flag.Flag) {
		switch value.Name {
		case "workflow-id":
			opts = append(opts, sdk.WithWorkflowID(*workflowID))
		case "description":
			opts = append(opts, sdk.WithWorkflowDescription(*description))
		case "execution-mode":
			opts = append(opts, sdk.WithWorkflowExecutionMode(*executionMode))
		case "cron-expression":
			opts = append(opts, sdk.WithWorkflowCronExpression(*cronExpression))
		case "compute-resource-id":
			opts = append(opts, sdk.WithWorkflowComputeResourceID(*computeResourceID))
		}
	})
	configuration, err := workflowDeploymentFileOptions(*volumeTriggerFile, *dynamicServiceFile, *defaultValuesFile)
	if err != nil {
		return err
	}
	opts = append(opts, configuration...)
	result, err := ctx.Client.Workflows(wid).DeployCandidate(
		context.Background(),
		*name,
		accepted,
		opts...,
	)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func acceptedWorkflowCandidateFromFile(path string) (*sdk.AcceptedWorkflowCandidate, error) {
	if path == "" {
		return nil, fmt.Errorf("acceptance-file is required")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open workflow candidate acceptance file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var acceptance sdk.WorkflowCandidateAcceptance
	if err := decoder.Decode(&acceptance); err != nil {
		return nil, fmt.Errorf("decode workflow candidate acceptance file: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("decode workflow candidate acceptance file: multiple JSON values")
		}
		return nil, fmt.Errorf("decode workflow candidate acceptance file trailing data: %w", err)
	}
	accepted, err := sdk.AcceptedWorkflowCandidateFromAcceptance(&acceptance)
	if err != nil {
		return nil, fmt.Errorf("validate workflow candidate acceptance file: %w", err)
	}
	return accepted, nil
}
