package commands

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"unicode/utf8"

	sdk "github.com/matrixorigin/moi-go-sdk/v5"
)

// ExecuteWorkflowCommand runs product SDK WorkflowService methods via moi-cli.
func ExecuteWorkflowCommand(ctx *Context, args []string) error {
	if len(args) == 0 {
		ShowWorkflowHelp(nil)
		return nil
	}
	sub := args[0]
	switch sub {
	case "create-worker-task":
		return runWorkflowServiceCreateWorkerTask(ctx, args[1:])
	case "create-version":
		return runWorkflowServiceCreateVersion(ctx, args[1:])
	case "version-by-number":
		return runWorkflowServiceVersionByNumber(ctx, args[1:])
	case "delete-version":
		return runWorkflowServiceDeleteVersion(ctx, args[1:])
	case "deprecate-version":
		return runWorkflowServiceDeprecateVersion(ctx, args[1:])
	case "publish-version":
		return runWorkflowServicePublishVersion(ctx, args[1:])
	case "version":
		return runWorkflowServiceVersion(ctx, args[1:])
	case "latest-published-version":
		return runWorkflowServiceLatestPublishedVersion(ctx, args[1:])
	case "versions":
		return runWorkflowServiceVersions(ctx, args[1:])
	case "reset-test-scratch":
		return runWorkflowServiceResetTestScratch(ctx, args[1:])
	case "observe-test-scratch":
		return runWorkflowServiceObserveTestScratch(ctx, args[1:])
	case "ensure-test-scratch":
		return runWorkflowServiceEnsureTestScratch(ctx, args[1:])
	case "test-candidate":
		return runWorkflowServiceTestCandidate(ctx, args[1:])
	case "submit-candidate":
		return runWorkflowServiceSubmitCandidate(ctx, args[1:])
	case "apps":
		return runWorkflowServiceApps(ctx, args[1:])
	case "restore-history":
		return runWorkflowServiceRestoreHistory(ctx, args[1:])
	case "reject-candidate-proposal":
		return runWorkflowServiceRejectCandidateProposal(ctx, args[1:])
	case "history":
		return runWorkflowServiceHistory(ctx, args[1:])
	case "list":
		return runWorkflowServiceList(ctx, args[1:])
	case "deploy":
		return runWorkflowServiceDeploy(ctx, args[1:])
	case "deploy-mowl":
		return runWorkflowServiceDeployMOWL(ctx, args[1:])
	case "accept-candidate":
		return runWorkflowServiceAcceptCandidate(ctx, args[1:])
	case "deploy-candidate":
		return runWorkflowServiceDeployCandidate(ctx, args[1:])
	case "get":
		return runWorkflowServiceGet(ctx, args[1:])
	case "update":
		return runWorkflowServiceUpdate(ctx, args[1:])
	case "delete":
		return runWorkflowServiceDelete(ctx, args[1:])
	case "enable":
		return runWorkflowServiceEnable(ctx, args[1:])
	case "pause":
		return runWorkflowServicePause(ctx, args[1:])
	case "resume":
		return runWorkflowServiceResume(ctx, args[1:])
	case "run":
		return runWorkflowServiceRun(ctx, args[1:])
	case "list-runs":
		return runWorkflowServiceListRuns(ctx, args[1:])
	case "list-all-runs":
		return runWorkflowServiceListAllRuns(ctx, args[1:])
	case "get-dispatch-job":
		return runWorkflowServiceGetDispatchJob(ctx, args[1:])
	case "get-run":
		return runWorkflowServiceGetRun(ctx, args[1:])
	case "get-run-by-id":
		return runWorkflowServiceGetRunByID(ctx, args[1:])
	case "refresh-run":
		return runWorkflowServiceRefreshRun(ctx, args[1:])
	case "pause-run":
		return runWorkflowServicePauseRun(ctx, args[1:])
	case "resume-run":
		return runWorkflowServiceResumeRun(ctx, args[1:])
	case "get-run-result":
		return runWorkflowServiceGetRunResult(ctx, args[1:])
	case "get-run-result-by-id":
		return runWorkflowServiceGetRunResultByID(ctx, args[1:])
	case "get-run-import-context-by-id":
		return runWorkflowServiceGetRunImportContextByID(ctx, args[1:])
	case "list-run-import-files-by-id":
		return runWorkflowServiceListRunImportFilesByID(ctx, args[1:])
	case "get-run-node-by-id":
		return runWorkflowServiceGetRunNodeByID(ctx, args[1:])
	case "get-run-import-context":
		return runWorkflowServiceGetRunImportContext(ctx, args[1:])
	case "list-run-import-files":
		return runWorkflowServiceListRunImportFiles(ctx, args[1:])
	case "get-run-node":
		return runWorkflowServiceGetRunNode(ctx, args[1:])
	case "get-run-node-rerun-plan":
		return runWorkflowServiceGetRunNodeRerunPlan(ctx, args[1:])
	case "create-run-node-rerun":
		return runWorkflowServiceCreateRunNodeRerun(ctx, args[1:])
	case "list-run-rerun-branches":
		return runWorkflowServiceListRunRerunBranches(ctx, args[1:])
	case "get-run-rerun":
		return runWorkflowServiceGetRunRerun(ctx, args[1:])
	case "cancel-run-rerun":
		return runWorkflowServiceCancelRunRerun(ctx, args[1:])
	case "retry-run":
		return runWorkflowServiceRetryRun(ctx, args[1:])
	case "batch-retry-failed-files":
		return runWorkflowServiceBatchRetryFailedFiles(ctx, args[1:])
	case "cancel-run":
		return runWorkflowServiceCancelRun(ctx, args[1:])
	case "delete-run":
		return runWorkflowServiceDeleteRun(ctx, args[1:])
	case "list-file-executions":
		return runWorkflowServiceListFileExecutions(ctx, args[1:])
	case "workitem-catalog":
		return runWorkflowServiceWorkitemCatalog(ctx, args[1:])
	case "workitem-catalog-item":
		return runWorkflowServiceWorkitemCatalogItem(ctx, args[1:])
	case "get-workitem-api-service":
		return runWorkflowServiceGetWorkitemAPIService(ctx, args[1:])
	case "publish-workitem-api-service":
		return runWorkflowServicePublishWorkitemAPIService(ctx, args[1:])
	case "invoke-workitem-api-service":
		return runWorkflowServiceInvokeWorkitemAPIService(ctx, args[1:])
	case "invoke-dynamic-service":
		return runWorkflowServiceInvokeDynamicService(ctx, args[1:])
	case "invoke-workflow-dynamic-service":
		return runWorkflowServiceInvokeWorkflowDynamicService(ctx, args[1:])
	case "enable-workitem-api-service":
		return runWorkflowServiceEnableWorkitemAPIService(ctx, args[1:])
	case "disable-workitem-api-service":
		return runWorkflowServiceDisableWorkitemAPIService(ctx, args[1:])
	case "custom-operators":
		return runWorkflowServiceCustomOperators(ctx, args[1:])
	case "create-custom-operator":
		return runWorkflowServiceCreateCustomOperator(ctx, args[1:])
	case "get-custom-operator":
		return runWorkflowServiceGetCustomOperator(ctx, args[1:])
	case "get-custom-operator-code":
		return runWorkflowServiceGetCustomOperatorCode(ctx, args[1:])
	case "get-custom-operator-source":
		return runWorkflowServiceGetCustomOperatorSource(ctx, args[1:])
	case "test-custom-operator":
		return runWorkflowServiceTestCustomOperator(ctx, args[1:])
	case "update-custom-operator":
		return runWorkflowServiceUpdateCustomOperator(ctx, args[1:])
	case "enable-custom-operator":
		return runWorkflowServiceEnableCustomOperator(ctx, args[1:])
	case "disable-custom-operator":
		return runWorkflowServiceDisableCustomOperator(ctx, args[1:])
	case "delete-custom-operator":
		return runWorkflowServiceDeleteCustomOperator(ctx, args[1:])
	case "catalog-file-lineage-overview":
		return runWorkflowServiceCatalogFileLineageOverview(ctx, args[1:])
	case "artifact-lineage-overview":
		return runWorkflowServiceArtifactLineageOverview(ctx, args[1:])
	case "artifact-output-blocks":
		return runWorkflowServiceArtifactOutputBlocks(ctx, args[1:])
	case "artifact-block-revisions":
		return runWorkflowServiceArtifactBlockRevisions(ctx, args[1:])
	case "create-artifact-block-revision":
		return runWorkflowServiceCreateArtifactBlockRevision(ctx, args[1:])
	case "switch-artifact-effective-revisions":
		return runWorkflowServiceSwitchArtifactEffectiveRevisions(ctx, args[1:])
	case "artifact-node":
		return runWorkflowServiceArtifactNode(ctx, args[1:])
	case "artifact-node-output-blocks":
		return runWorkflowServiceArtifactNodeOutputBlocks(ctx, args[1:])
	case "artifact-node-block-revisions":
		return runWorkflowServiceArtifactNodeBlockRevisions(ctx, args[1:])
	case "create-artifact-node-block-revision":
		return runWorkflowServiceCreateArtifactNodeBlockRevision(ctx, args[1:])
	case "switch-artifact-node-effective-revisions":
		return runWorkflowServiceSwitchArtifactNodeEffectiveRevisions(ctx, args[1:])
	case "create-artifact-node-rerun":
		return runWorkflowServiceCreateArtifactNodeRerun(ctx, args[1:])
	case "create-target-artifact-node-revision-rerun":
		return runWorkflowServiceCreateTargetArtifactNodeRevisionRerun(ctx, args[1:])
	case "workflow-templates":
		return runWorkflowServiceWorkflowTemplates(ctx, args[1:])
	case "get-workflow-template":
		return runWorkflowServiceGetWorkflowTemplate(ctx, args[1:])
	case "create-workflow-template":
		return runWorkflowServiceCreateWorkflowTemplate(ctx, args[1:])
	case "create-mowl-workflow-template":
		return runWorkflowServiceCreateMOWLWorkflowTemplate(ctx, args[1:])
	case "update-workflow-template":
		return runWorkflowServiceUpdateWorkflowTemplate(ctx, args[1:])
	case "update-mowl-workflow-template":
		return runWorkflowServiceUpdateMOWLWorkflowTemplate(ctx, args[1:])
	case "delete-workflow-template":
		return runWorkflowServiceDeleteWorkflowTemplate(ctx, args[1:])
	default:
		return fmt.Errorf("unknown workflow subcommand: %s", sub)
	}
}

// ShowWorkflowHelp prints help for workflow.
func ShowWorkflowHelp(_ []string) {
	fmt.Fprintln(os.Stderr, "Additional commands: create-worker-task, create-version, version-by-number, delete-version, deprecate-version, publish-version, version, latest-published-version, versions")

	fmt.Fprintf(os.Stderr, `Usage: workflow <subcommand> [options]

Product SDK surface: WorkflowService

Subcommands:
  create-worker-task               WorkflowService.CreateWorkerTask
  reset-test-scratch               WorkflowService.ResetTestScratch
  observe-test-scratch             WorkflowService.ObserveTestScratch
  ensure-test-scratch              WorkflowService.EnsureTestScratch
  test-candidate                   WorkflowService.TestCandidate
  submit-candidate                 WorkflowService.SubmitCandidate
  apps                             WorkflowService.Apps
  restore-history                  WorkflowService.RestoreHistory
  reject-candidate-proposal        WorkflowService.RejectCandidateProposal
  history                          WorkflowService.History
  list                             WorkflowService.List
  deploy                           WorkflowService.Deploy
  deploy-mowl                      WorkflowService.DeployMOWL
  accept-candidate                 WorkflowService.AcceptCandidate
  deploy-candidate                 WorkflowService.DeployCandidate
  get                              WorkflowService.Get
  update                           WorkflowService.Update
  delete                           WorkflowService.Delete
  enable                           WorkflowService.Enable
  pause                            WorkflowService.Pause
  resume                           WorkflowService.Resume
  run                              WorkflowService.Run
  list-runs                        WorkflowService.ListRuns
  list-all-runs                    WorkflowService.ListAllRuns
  get-dispatch-job                 WorkflowService.GetDispatchJob
  get-run                          WorkflowService.GetRun
  get-run-by-id                    WorkflowService.GetRunByID
  refresh-run                      WorkflowService.RefreshRun
  pause-run                        WorkflowService.PauseRun
  resume-run                       WorkflowService.ResumeRun
  get-run-result                   WorkflowService.GetRunResult
  get-run-result-by-id             WorkflowService.GetRunResultByID
  get-run-import-context-by-id     WorkflowService.GetRunImportContextByID
  list-run-import-files-by-id      WorkflowService.ListRunImportFilesByID
  get-run-node-by-id               WorkflowService.GetRunNodeByID
  get-run-import-context           WorkflowService.GetRunImportContext
  list-run-import-files            WorkflowService.ListRunImportFiles
  get-run-node                     WorkflowService.GetRunNode
  get-run-node-rerun-plan          WorkflowService.GetRunNodeRerunPlan
  create-run-node-rerun            WorkflowService.CreateRunNodeRerun
  list-run-rerun-branches          WorkflowService.ListRunRerunBranches
  get-run-rerun                    WorkflowService.GetRunRerun
  cancel-run-rerun                 WorkflowService.CancelRunRerun
  retry-run                        WorkflowService.RetryRun
  batch-retry-failed-files         WorkflowService.BatchRetryFailedFiles
  cancel-run                       WorkflowService.CancelRun
  delete-run                       WorkflowService.DeleteRun
  list-file-executions             WorkflowService.ListFileExecutions
  workitem-catalog                 WorkflowService.WorkitemCatalog
  workitem-catalog-item            WorkflowService.WorkitemCatalogItem
  get-workitem-api-service         WorkflowService.GetWorkitemAPIService
  publish-workitem-api-service     WorkflowService.PublishWorkitemAPIService
  invoke-workitem-api-service      WorkflowService.InvokeWorkitemAPIService
  invoke-dynamic-service           WorkflowService.InvokeDynamicService
  invoke-workflow-dynamic-service  WorkflowService.InvokeWorkflowDynamicService
  enable-workitem-api-service      WorkflowService.EnableWorkitemAPIService
  disable-workitem-api-service     WorkflowService.DisableWorkitemAPIService
  custom-operators                 WorkflowService.CustomOperators
  create-custom-operator           WorkflowService.CreateCustomOperator
  get-custom-operator              WorkflowService.GetCustomOperator
  get-custom-operator-code         WorkflowService.GetCustomOperatorCode
  get-custom-operator-source       WorkflowService.GetCustomOperatorSource
  test-custom-operator             WorkflowService.TestCustomOperator
  update-custom-operator           WorkflowService.UpdateCustomOperator
  enable-custom-operator           WorkflowService.EnableCustomOperator
  disable-custom-operator          WorkflowService.DisableCustomOperator
  delete-custom-operator           WorkflowService.DeleteCustomOperator
  catalog-file-lineage-overview    WorkflowService.CatalogFileLineageOverview
  artifact-lineage-overview        WorkflowService.ArtifactLineageOverview
  artifact-output-blocks           WorkflowService.ArtifactOutputBlocks
  artifact-block-revisions         WorkflowService.ArtifactBlockRevisions
  create-artifact-block-revision   WorkflowService.CreateArtifactBlockRevision
  switch-artifact-effective-revisions WorkflowService.SwitchArtifactEffectiveRevisions
  artifact-node                    WorkflowService.ArtifactNode
  artifact-node-output-blocks      WorkflowService.ArtifactNodeOutputBlocks
  artifact-node-block-revisions    WorkflowService.ArtifactNodeBlockRevisions
  create-artifact-node-block-revision WorkflowService.CreateArtifactNodeBlockRevision
  switch-artifact-node-effective-revisions WorkflowService.SwitchArtifactNodeEffectiveRevisions
  create-artifact-node-rerun       WorkflowService.CreateArtifactNodeRerun
  create-target-artifact-node-revision-rerun WorkflowService.CreateTargetArtifactNodeRevisionRerun
  workflow-templates               WorkflowService.WorkflowTemplates
  get-workflow-template            WorkflowService.GetWorkflowTemplate
  create-workflow-template         WorkflowService.CreateWorkflowTemplate
  create-mowl-workflow-template    WorkflowService.CreateMOWLWorkflowTemplate
  update-workflow-template         WorkflowService.UpdateWorkflowTemplate
  update-mowl-workflow-template    WorkflowService.UpdateMOWLWorkflowTemplate
  delete-workflow-template         WorkflowService.DeleteWorkflowTemplate

Complex arguments accept --json / --json-file.
Stream methods accept --output <path>.

`)
}

func runWorkflowServiceList(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow list", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	limit := fs.Int("limit", 0, "Maximum number of workflows to return")
	offset := fs.Int("offset", 0, "Number of workflows to skip")
	sourceType := fs.String("source-type", "", "Filter by workflow source type")
	status := fs.String("status", "", "Filter by workflow status")
	executionMode := fs.String("execution-mode", "", "Filter by execution mode")
	nameSearch := fs.String("name-search", "", "Filter by workflow name")
	workflowDefinitionID := fs.String("workflow-definition-id", "", "Filter by workflow definition ID")
	includeDynamicService := fs.Bool("include-dynamic-service", false, "Include dynamic service workflows")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow list [options]\n\nProduct SDK: WorkflowService.List\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	var opts []sdk.ListOption
	fs.Visit(func(value *flag.Flag) {
		switch value.Name {
		case "limit":
			opts = append(opts, sdk.WithLimit(*limit))
		case "offset":
			opts = append(opts, sdk.WithOffset(*offset))
		case "source-type":
			opts = append(opts, sdk.WithQueryParam("source_type", *sourceType))
		case "status":
			opts = append(opts, sdk.WithStatus(*status))
		case "execution-mode":
			opts = append(opts, sdk.WithQueryParam("execution_mode", *executionMode))
		case "name-search":
			opts = append(opts, sdk.WithQueryParam("name_search", *nameSearch))
		case "workflow-definition-id":
			opts = append(opts, sdk.WithQueryParam("workflow_definition_id", *workflowDefinitionID))
		case "include-dynamic-service":
			opts = append(opts, sdk.WithQueryParam("include_dynamic_service", strconv.FormatBool(*includeDynamicService)))
		}
	})
	result, err := ctx.Client.Workflows(wid).List(context.Background(), opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func printWorkflowScheduleGuidance(fs *flag.FlagSet) {
	fmt.Fprintln(fs.Output(), `
Cron schedules:
  Use --execution-mode cron together with --cron-expression.
  Five fields: minute hour day-of-month month day-of-week.
  Six fields: second minute hour day-of-month month day-of-week.
  Without a timezone prefix the server uses Asia/Shanghai. Use CRON_TZ=<IANA timezone> to set it explicitly.

  Daily at 07:00 Shanghai time, updating an existing saved workflow:
    moi-cli workflow update --workflow-id <id> --execution-mode cron --cron-expression 'CRON_TZ=Asia/Shanghai 0 7 * * *'
  Equivalent six-field expression: 'CRON_TZ=Asia/Shanghai 0 0 7 * * *'.
  Restore manual runs: moi-cli workflow update --workflow-id <id> --execution-mode one_shot.
  Confirm the saved execution_mode, cron_expression and trigger_summary with workflow get --workflow-id <id>.
  Updating schedule settings reuses the saved definition. Changing an accepted candidate's definition requires a new accepted candidate.`)
}

func printWorkflowRuntimeFieldsGuidance(fs *flag.FlagSet) {
	fmt.Fprintln(fs.Output(), `
Runtime configuration:
  --runtime-fields-file declares parameters; --default-values-file stores their values by field_id.
  Fields JSON example:
    {"fields":[{"field_id":"text","label":"Text","widget":"text","json_type":"string","bind_to":"vars.text"}]}
  Defaults JSON example: {"text":"saved value"}.
  bind_to uses input.<path> or vars.<path>. Configure these files directly; no UI form is required.
  Agent sessions can deploy authored MOWL with --workflow-file, run test inputs with workflow run --values-file,
  then inspect workflow get-run-result. The dedicated Workflow editor retains its candidate review.`)
}

func runWorkflowServiceDeploy(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow deploy", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	name := fs.String("name", "", "name")
	dslYAML := fs.String("dsl-yaml", "", "dslYAML")
	workflowFile := fs.String("workflow-file", "", "Authored native MOWL JSON file (exclusive with --dsl-yaml)")
	workflowID := fs.String("workflow-id", "", "Existing workflow definition ID to deploy a new version")
	description := fs.String("description", "", "Workflow description")
	executionMode := fs.String("execution-mode", "", "Workflow execution mode: one_shot, cron, volume_trigger or dynamic_service")
	cronExpression := fs.String("cron-expression", "", "Cron expression; see workflow update --help for syntax, timezone and examples")
	volumeTriggerFile := fs.String("volume-trigger-file", "", "Volume trigger configuration JSON file")
	dynamicServiceFile := fs.String("dynamic-service-file", "", "Dynamic service configuration JSON file")
	defaultValuesFile := fs.String("default-values-file", "", "Workflow default values JSON file")
	runtimeFieldsFile := fs.String("runtime-fields-file", "", "Runtime parameter fields JSON file; see workflow deploy --help for an example")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow deploy [options]\n\nProduct SDK: WorkflowService.Deploy\n\nOptions:\n")
		fs.PrintDefaults()
		printWorkflowScheduleGuidance(fs)
		printWorkflowRuntimeFieldsGuidance(fs)
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	var opts []sdk.WorkflowDeployOption
	if *runtimeFieldsFile != "" {
		fields, err := loadJSONFile[map[string]any](*runtimeFieldsFile)
		if err != nil {
			return err
		}
		opts = append(opts, sdk.WithWorkflowInputForm(fields))
	}
	if *workflowFile != "" {
		definition, err := loadJSONBytes("", *workflowFile)
		if err != nil {
			return err
		}
		opts = append(opts, sdk.WithWorkflowMOWL(definition))
	}
	if *executionMode != "" {
		opts = append(opts, sdk.WithWorkflowExecutionMode(*executionMode))
	}
	fs.Visit(func(value *flag.Flag) {
		switch value.Name {
		case "workflow-id":
			opts = append(opts, sdk.WithWorkflowID(*workflowID))
		case "description":
			opts = append(opts, sdk.WithWorkflowDescription(*description))
		case "cron-expression":
			opts = append(opts, sdk.WithWorkflowCronExpression(*cronExpression))
		}
	})
	configuration, err := workflowDeploymentFileOptions(*volumeTriggerFile, *dynamicServiceFile, *defaultValuesFile)
	if err != nil {
		return err
	}
	opts = append(opts, configuration...)
	result, err := ctx.Client.Workflows(wid).Deploy(context.Background(), *name, *dslYAML, opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func workflowDeploymentFileOptions(volumeTriggerFile, dynamicServiceFile, defaultValuesFile string) ([]sdk.WorkflowDeployOption, error) {
	var opts []sdk.WorkflowDeployOption
	if volumeTriggerFile != "" {
		trigger, err := loadJSONFile[sdk.WorkflowVolumeTriggerParam](volumeTriggerFile)
		if err != nil {
			return nil, err
		}
		opts = append(opts, sdk.WithWorkflowVolumeTrigger(trigger))
	}
	if dynamicServiceFile != "" {
		service, err := loadJSONFile[sdk.WorkflowDynamicServiceParam](dynamicServiceFile)
		if err != nil {
			return nil, err
		}
		opts = append(opts, sdk.WithWorkflowDynamicService(service))
	}
	if defaultValuesFile != "" {
		values, err := loadJSONFile[map[string]any](defaultValuesFile)
		if err != nil {
			return nil, err
		}
		opts = append(opts, sdk.WithWorkflowDefaultValues(values))
	}
	return opts, nil
}

func runWorkflowServiceGet(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow get", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	workflowID := fs.String("workflow-id", "", "workflowID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow get [options]\n\nProduct SDK: WorkflowService.Get\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).Get(context.Background(), *workflowID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceUpdate(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow update", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	workflowID := fs.String("workflow-id", "", "workflowID")
	name := fs.String("name", "", "Workflow name")
	description := fs.String("description", "", "Workflow description; pass an empty value to clear it")
	executionMode := fs.String("execution-mode", "", "Workflow execution mode: one_shot, cron, volume_trigger or dynamic_service")
	cronExpression := fs.String("cron-expression", "", "Cron expression; see workflow update --help for syntax, timezone and examples")
	dslYAML := fs.String("dsl-yaml", "", "Replacement DSL YAML; pass an empty value to clear it")
	volumeTriggerFile := fs.String("volume-trigger-file", "", "Volume trigger configuration JSON file")
	dynamicServiceFile := fs.String("dynamic-service-file", "", "Dynamic service configuration JSON file")
	defaultValuesFile := fs.String("default-values-file", "", "Workflow default values JSON file")
	runtimeFieldsFile := fs.String("runtime-fields-file", "", "Runtime parameter fields JSON file; see workflow deploy --help for an example")
	runtimeLayoutFile := fs.String("runtime-layout-file", "", "Runtime form layout JSON file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow update [options]\n\nProduct SDK: WorkflowService.Update\n\nOptions:\n")
		fs.PrintDefaults()
		printWorkflowScheduleGuidance(fs)
		printWorkflowRuntimeFieldsGuidance(fs)
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	var opts []sdk.WorkflowUpdateOption
	if *runtimeFieldsFile != "" {
		fields, err := loadJSONFile[map[string]any](*runtimeFieldsFile)
		if err != nil {
			return err
		}
		opts = append(opts, sdk.WithWorkflowUpdateInputForm(fields))
	}
	if *runtimeLayoutFile != "" {
		layout, err := loadJSONFile[json.RawMessage](*runtimeLayoutFile)
		if err != nil {
			return err
		}
		opts = append(opts, sdk.WithWorkflowUpdateFormLayout(layout))
	}
	fs.Visit(func(value *flag.Flag) {
		switch value.Name {
		case "name":
			opts = append(opts, sdk.WithWorkflowUpdateName(*name))
		case "description":
			opts = append(opts, sdk.WithWorkflowUpdateDescription(*description))
		case "execution-mode":
			opts = append(opts, sdk.WithWorkflowUpdateExecutionMode(*executionMode))
		case "cron-expression":
			opts = append(opts, sdk.WithWorkflowUpdateCronExpression(*cronExpression))
		case "dsl-yaml":
			opts = append(opts, sdk.WithWorkflowUpdateDSLYAML(*dslYAML))
		}
	})
	if *volumeTriggerFile != "" {
		trigger, err := loadJSONFile[sdk.WorkflowVolumeTriggerParam](*volumeTriggerFile)
		if err != nil {
			return err
		}
		opts = append(opts, sdk.WithWorkflowUpdateVolumeTrigger(trigger))
	}
	if *dynamicServiceFile != "" {
		service, err := loadJSONFile[sdk.WorkflowDynamicServiceParam](*dynamicServiceFile)
		if err != nil {
			return err
		}
		opts = append(opts, sdk.WithWorkflowUpdateDynamicService(service))
	}
	if *defaultValuesFile != "" {
		values, err := loadJSONFile[map[string]any](*defaultValuesFile)
		if err != nil {
			return err
		}
		opts = append(opts, sdk.WithWorkflowUpdateDefaultValues(values))
	}
	result, err := ctx.Client.Workflows(wid).Update(context.Background(), *workflowID, opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceDelete(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow delete", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	workflowID := fs.String("workflow-id", "", "workflowID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow delete [options]\n\nProduct SDK: WorkflowService.Delete\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).Delete(context.Background(), *workflowID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceEnable(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow enable", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	workflowID := fs.String("workflow-id", "", "workflowID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow enable [options]\n\nProduct SDK: WorkflowService.Enable\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).Enable(context.Background(), *workflowID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServicePause(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow pause", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	workflowID := fs.String("workflow-id", "", "workflowID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow pause [options]\n\nProduct SDK: WorkflowService.Pause\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).Pause(context.Background(), *workflowID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceResume(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow resume", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	workflowID := fs.String("workflow-id", "", "workflowID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow resume [options]\n\nProduct SDK: WorkflowService.Resume\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).Resume(context.Background(), *workflowID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

type workflowExecutionInputFlags struct {
	readValues           func() (map[string]any, error)
	replaceSavedDefaults *bool
	triggerNow           *bool
}

func addWorkflowExecutionInputFlags(fs *flag.FlagSet) workflowExecutionInputFlags {
	valuesFile := fs.String("values-file", "", "JSON object containing workflow runtime values")
	return workflowExecutionInputFlags{
		readValues: func() (map[string]any, error) {
			if *valuesFile == "" {
				return nil, nil
			}
			return loadJSONFile[map[string]any](*valuesFile)
		},
		replaceSavedDefaults: fs.Bool("replace-saved-defaults", false, "Replace saved workflow defaults with the supplied values"),
		triggerNow:           fs.Bool("trigger-now", false, "Request immediate workflow dispatch; may be set to false explicitly"),
	}
}

func (inputs workflowExecutionInputFlags) read(fs *flag.FlagSet) (values map[string]any, triggerNow *bool, err error) {
	values, err = inputs.readValues()
	if err != nil {
		return nil, nil, err
	}
	fs.Visit(func(value *flag.Flag) {
		if value.Name == "trigger-now" {
			triggerNow = inputs.triggerNow
		}
	})
	return values, triggerNow, nil
}

func runWorkflowServiceRun(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow run", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	workflowID := fs.String("workflow-id", "", "workflowID")
	inputs := addWorkflowExecutionInputFlags(fs)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow run [options]\n\nProduct SDK: WorkflowService.Run\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	var opts []sdk.WorkflowRunOption
	values, triggerNow, err := inputs.read(fs)
	if err != nil {
		return err
	}
	if values != nil {
		opts = append(opts, sdk.WithWorkflowRunValues(values))
	}
	if *inputs.replaceSavedDefaults {
		opts = append(opts, sdk.WithWorkflowRunReplaceSavedDefaults(true))
	}
	if triggerNow != nil {
		opts = append(opts, sdk.WithWorkflowRunTriggerNow(*triggerNow))
	}
	result, err := ctx.Client.Workflows(wid).Run(context.Background(), *workflowID, opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

type workflowRunListFlags struct {
	limit, offset                       *int
	status, executionMode, workflowName *string
}

func addWorkflowRunListFlags(fs *flag.FlagSet) workflowRunListFlags {
	return workflowRunListFlags{
		limit:         fs.Int("limit", 0, "Maximum number of executions to return"),
		offset:        fs.Int("offset", 0, "Number of executions to skip"),
		status:        fs.String("status", "", "Filter by execution status"),
		executionMode: fs.String("execution-mode", "", "Filter by execution mode"),
		workflowName:  fs.String("workflow-name", "", "Filter by workflow name"),
	}
}

func (flags workflowRunListFlags) options(fs *flag.FlagSet) []sdk.ListOption {
	var opts []sdk.ListOption
	fs.Visit(func(value *flag.Flag) {
		switch value.Name {
		case "limit":
			opts = append(opts, sdk.WithLimit(*flags.limit))
		case "offset":
			opts = append(opts, sdk.WithOffset(*flags.offset))
		case "status":
			opts = append(opts, sdk.WithStatus(*flags.status))
		case "execution-mode":
			opts = append(opts, sdk.WithQueryParam("execution_mode", *flags.executionMode))
		case "workflow-name":
			opts = append(opts, sdk.WithQueryParam("workflow_name", *flags.workflowName))
		}
	})
	return opts
}

func runWorkflowServiceListRuns(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow list-runs", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	workflowID := fs.String("workflow-id", "", "workflowID")
	listFlags := addWorkflowRunListFlags(fs)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow list-runs [options]\n\nProduct SDK: WorkflowService.ListRuns\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	opts := listFlags.options(fs)
	result, err := ctx.Client.Workflows(wid).ListRuns(context.Background(), *workflowID, opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceListAllRuns(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow list-all-runs", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	listFlags := addWorkflowRunListFlags(fs)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow list-all-runs [options]\n\nProduct SDK: WorkflowService.ListAllRuns\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	opts := listFlags.options(fs)
	result, err := ctx.Client.Workflows(wid).ListAllRuns(context.Background(), opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceGetDispatchJob(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow get-dispatch-job", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	workflowID := fs.String("workflow-id", "", "workflowID")
	dispatchJobID := fs.String("dispatch-job-id", "", "dispatchJobID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow get-dispatch-job [options]\n\nProduct SDK: WorkflowService.GetDispatchJob\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).GetDispatchJob(context.Background(), *workflowID, *dispatchJobID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceGetRun(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow get-run", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	workflowID := fs.String("workflow-id", "", "workflowID")
	executionID := fs.String("execution-id", "", "executionID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow get-run [options]\n\nProduct SDK: WorkflowService.GetRun\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).GetRun(context.Background(), *workflowID, *executionID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceGetRunByID(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow get-run-by-id", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	executionID := fs.String("execution-id", "", "executionID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow get-run-by-id [options]\n\nProduct SDK: WorkflowService.GetRunByID\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).GetRunByID(context.Background(), *executionID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceRefreshRun(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow refresh-run", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	workflowID := fs.String("workflow-id", "", "workflowID")
	executionID := fs.String("execution-id", "", "executionID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow refresh-run [options]\n\nProduct SDK: WorkflowService.RefreshRun\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).RefreshRun(context.Background(), *workflowID, *executionID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServicePauseRun(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow pause-run", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	workflowID := fs.String("workflow-id", "", "workflowID")
	executionID := fs.String("execution-id", "", "executionID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow pause-run [options]\n\nProduct SDK: WorkflowService.PauseRun\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).PauseRun(context.Background(), *workflowID, *executionID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceResumeRun(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow resume-run", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	workflowID := fs.String("workflow-id", "", "workflowID")
	executionID := fs.String("execution-id", "", "executionID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow resume-run [options]\n\nProduct SDK: WorkflowService.ResumeRun\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).ResumeRun(context.Background(), *workflowID, *executionID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceGetRunResult(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow get-run-result", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	workflowID := fs.String("workflow-id", "", "Workflow application ID (use get-run-result-by-id when only the execution ID is known)")
	executionID := fs.String("execution-id", "", "Execution ID returned by workflow run")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow get-run-result [options]\n\nProduct SDK: WorkflowService.GetRunResult\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).GetRunResult(context.Background(), *workflowID, *executionID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceGetRunResultByID(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow get-run-result-by-id", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	executionID := fs.String("execution-id", "", "Execution ID returned by workflow run")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow get-run-result-by-id [options]\n\nProduct SDK: WorkflowService.GetRunResultByID\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).GetRunResultByID(context.Background(), *executionID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceGetRunImportContextByID(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow get-run-import-context-by-id", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	executionID := fs.String("execution-id", "", "executionID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow get-run-import-context-by-id [options]\n\nProduct SDK: WorkflowService.GetRunImportContextByID\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).GetRunImportContextByID(context.Background(), *executionID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceListRunImportFilesByID(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow list-run-import-files-by-id", flag.ExitOnError)
	var queryParams queryParameters
	fs.Var(&queryParams, "query-param", "Product query parameter name=value (repeatable; pagination/filter names follow the product API)")
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	executionID := fs.String("execution-id", "", "executionID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow list-run-import-files-by-id [options]\n\nProduct SDK: WorkflowService.ListRunImportFilesByID\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).ListRunImportFilesByID(context.Background(), *executionID, queryParams...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceGetRunNodeByID(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow get-run-node-by-id", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	executionID := fs.String("execution-id", "", "executionID")
	nodeKey := fs.String("node-key", "", "nodeKey")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow get-run-node-by-id [options]\n\nProduct SDK: WorkflowService.GetRunNodeByID\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).GetRunNodeByID(context.Background(), *executionID, *nodeKey)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceGetRunImportContext(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow get-run-import-context", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	workflowID := fs.String("workflow-id", "", "workflowID")
	executionID := fs.String("execution-id", "", "executionID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow get-run-import-context [options]\n\nProduct SDK: WorkflowService.GetRunImportContext\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).GetRunImportContext(context.Background(), *workflowID, *executionID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceListRunImportFiles(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow list-run-import-files", flag.ExitOnError)
	var queryParams queryParameters
	fs.Var(&queryParams, "query-param", "Product query parameter name=value (repeatable; pagination/filter names follow the product API)")
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	workflowID := fs.String("workflow-id", "", "workflowID")
	executionID := fs.String("execution-id", "", "executionID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow list-run-import-files [options]\n\nProduct SDK: WorkflowService.ListRunImportFiles\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).ListRunImportFiles(context.Background(), *workflowID, *executionID, queryParams...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceGetRunNode(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow get-run-node", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	workflowID := fs.String("workflow-id", "", "workflowID")
	executionID := fs.String("execution-id", "", "executionID")
	nodeKey := fs.String("node-key", "", "nodeKey")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow get-run-node [options]\n\nProduct SDK: WorkflowService.GetRunNode\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).GetRunNode(context.Background(), *workflowID, *executionID, *nodeKey)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceGetRunNodeRerunPlan(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow get-run-node-rerun-plan", flag.ExitOnError)
	var queryParams queryParameters
	fs.Var(&queryParams, "query-param", "Product query parameter name=value (repeatable; pagination/filter names follow the product API)")
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	workflowID := fs.String("workflow-id", "", "workflowID")
	executionID := fs.String("execution-id", "", "executionID")
	nodeKey := fs.String("node-key", "", "nodeKey")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow get-run-node-rerun-plan [options]\n\nProduct SDK: WorkflowService.GetRunNodeRerunPlan\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).GetRunNodeRerunPlan(context.Background(), *workflowID, *executionID, *nodeKey, queryParams...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceCreateRunNodeRerun(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow create-run-node-rerun", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	workflowID := fs.String("workflow-id", "", "workflowID")
	executionID := fs.String("execution-id", "", "executionID")
	nodeKey := fs.String("node-key", "", "nodeKey")
	mode := fs.String("mode", "", "mode")
	sourceNodeExecutionID := fs.String("source-node-execution-id", "", "sourceNodeExecutionID")
	expectedRerunContractHash := fs.String("expected-rerun-contract-hash", "", "Required rerun contract hash returned by get-run-node-rerun-plan")
	valuesFile := fs.String("values-file", "", "JSON object containing rerun runtime values")
	configOverrideFile := fs.String("config-override-file", "", "JSON file containing the rerun config override")
	reason := fs.String("reason", "", "Rerun reason")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow create-run-node-rerun [options]\n\nProduct SDK: WorkflowService.CreateRunNodeRerun\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	opts := []sdk.WorkflowRerunOption{sdk.WithExpectedWorkflowRerunContractHash(*expectedRerunContractHash)}
	if *valuesFile != "" {
		values, err := loadJSONFile[map[string]any](*valuesFile)
		if err != nil {
			return err
		}
		opts = append(opts, sdk.WithWorkflowRerunValues(values))
	}
	if *configOverrideFile != "" {
		raw, err := loadJSONBytes("", *configOverrideFile)
		if err != nil {
			return err
		}
		opts = append(opts, sdk.WithWorkflowRerunConfigOverrideJSON(json.RawMessage(raw)))
	}
	fs.Visit(func(value *flag.Flag) {
		if value.Name == "reason" {
			opts = append(opts, sdk.WithWorkflowRerunReason(*reason))
		}
	})
	result, err := ctx.Client.Workflows(wid).CreateRunNodeRerun(context.Background(), *workflowID, *executionID, *nodeKey, *mode, *sourceNodeExecutionID, opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceListRunRerunBranches(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow list-run-rerun-branches", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	workflowID := fs.String("workflow-id", "", "workflowID")
	executionID := fs.String("execution-id", "", "executionID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow list-run-rerun-branches [options]\n\nProduct SDK: WorkflowService.ListRunRerunBranches\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).ListRunRerunBranches(context.Background(), *workflowID, *executionID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceGetRunRerun(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow get-run-rerun", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	workflowID := fs.String("workflow-id", "", "workflowID")
	executionID := fs.String("execution-id", "", "executionID")
	rerunID := fs.String("rerun-id", "", "rerunID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow get-run-rerun [options]\n\nProduct SDK: WorkflowService.GetRunRerun\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).GetRunRerun(context.Background(), *workflowID, *executionID, *rerunID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceCancelRunRerun(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow cancel-run-rerun", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	workflowID := fs.String("workflow-id", "", "workflowID")
	executionID := fs.String("execution-id", "", "executionID")
	rerunID := fs.String("rerun-id", "", "rerunID")
	reason := fs.String("reason", "", "Cancellation reason")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow cancel-run-rerun [options]\n\nProduct SDK: WorkflowService.CancelRunRerun\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	var opts []sdk.WorkflowCancelOption
	fs.Visit(func(value *flag.Flag) {
		if value.Name == "reason" {
			opts = append(opts, sdk.WithWorkflowCancelReason(*reason))
		}
	})
	result, err := ctx.Client.Workflows(wid).CancelRunRerun(context.Background(), *workflowID, *executionID, *rerunID, opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceRetryRun(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow retry-run", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	workflowID := fs.String("workflow-id", "", "workflowID")
	executionID := fs.String("execution-id", "", "executionID")
	inputs := addWorkflowExecutionInputFlags(fs)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow retry-run [options]\n\nProduct SDK: WorkflowService.RetryRun\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	values, triggerNow, err := inputs.read(fs)
	if err != nil {
		return err
	}
	var opts []sdk.WorkflowRerunOption
	if values != nil {
		opts = append(opts, sdk.WithWorkflowRerunValues(values))
	}
	if *inputs.replaceSavedDefaults {
		opts = append(opts, sdk.WithWorkflowRetryReplaceSavedDefaults(true))
	}
	if triggerNow != nil {
		opts = append(opts, sdk.WithWorkflowRerunTriggerNow(*triggerNow))
	}
	result, err := ctx.Client.Workflows(wid).RetryRun(context.Background(), *workflowID, *executionID, opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceBatchRetryFailedFiles(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow batch-retry-failed-files", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	workflowID := fs.String("workflow-id", "", "Workflow ID")
	sourceDispatchJobID := fs.String("source-dispatch-job-id", "", "Source Volume dispatch job ID")
	executionIDs := fs.String("execution-ids", "", "Comma-separated failed execution IDs")
	allFailed := fs.Bool("all-failed", false, "Retry every currently failed file in the source dispatch job")
	requestID := fs.String("request-id", "", "Client-generated idempotency key")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow batch-retry-failed-files [options]\n\nProduct SDK: WorkflowService.BatchRetryFailedFiles\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).BatchRetryFailedFiles(
		context.Background(),
		*workflowID,
		*sourceDispatchJobID,
		splitCSV(*executionIDs),
		*allFailed,
		*requestID,
	)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceCancelRun(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow cancel-run", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	workflowID := fs.String("workflow-id", "", "workflowID")
	executionID := fs.String("execution-id", "", "executionID")
	reason := fs.String("reason", "", "Cancellation reason")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow cancel-run [options]\n\nProduct SDK: WorkflowService.CancelRun\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	var opts []sdk.WorkflowCancelOption
	fs.Visit(func(value *flag.Flag) {
		if value.Name == "reason" {
			opts = append(opts, sdk.WithWorkflowCancelReason(*reason))
		}
	})
	result, err := ctx.Client.Workflows(wid).CancelRun(context.Background(), *workflowID, *executionID, opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceDeleteRun(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow delete-run", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	workflowID := fs.String("workflow-id", "", "workflowID")
	executionID := fs.String("execution-id", "", "executionID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow delete-run [options]\n\nProduct SDK: WorkflowService.DeleteRun\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).DeleteRun(context.Background(), *workflowID, *executionID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceListFileExecutions(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow list-file-executions", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fileID := fs.String("file-id", "", "Catalog file ID whose workflow executions should be returned")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow list-file-executions [options]\n\nProduct SDK: WorkflowService.ListFileExecutions\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	var opts []sdk.QueryOption
	if *fileID != "" {
		opts = append(opts, sdk.WithFileID(*fileID))
	}
	result, err := ctx.Client.Workflows(wid).ListFileExecutions(context.Background(), opts...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceWorkitemCatalog(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow workitem-catalog", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow workitem-catalog [options]\n\nProduct SDK: WorkflowService.WorkitemCatalog\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).WorkitemCatalog(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceWorkitemCatalogItem(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow workitem-catalog-item", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	nodeID := fs.String("node-id", "", "nodeID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow workitem-catalog-item [options]\n\nProduct SDK: WorkflowService.WorkitemCatalogItem\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).WorkitemCatalogItem(context.Background(), *nodeID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceGetWorkitemAPIService(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow get-workitem-api-service", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	nodeID := fs.String("node-id", "", "nodeID")
	version := fs.String("version", "", "version")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow get-workitem-api-service [options]\n\nProduct SDK: WorkflowService.GetWorkitemAPIService\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).GetWorkitemAPIService(context.Background(), *nodeID, *version)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServicePublishWorkitemAPIService(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow publish-workitem-api-service", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	nodeID := fs.String("node-id", "", "nodeID")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow publish-workitem-api-service [options]\n\nProduct SDK: WorkflowService.PublishWorkitemAPIService\n\nOptions:\n")
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
	var spec sdk.WorkitemAPIServicePublishSpec
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &spec); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.Workflows(wid).PublishWorkitemAPIService(context.Background(), *nodeID, spec)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

// workflowServiceInvokeGuidance documents the published invocation contract
// alongside the structural CLI schema. Business validation remains server-owned.
func workflowServiceInvokeGuidance(commandName string) map[string]any {
	if commandName == "workflow invoke-workitem-api-service" {
		return map[string]any{
			"required": []string{"service_name", "type", "version", "payload"},
			"field_contract": map[string]string{
				"service_name": "Use the published service.service_name.",
				"type":         "Use the literal operator for this legacy WorkItem API route.",
				"version":      "Use the published operator version.",
				"payload":      "Business input matching the published input_schema; do not wrap it in input.",
			},
			"example": map[string]any{
				"service_name": "<returned-service-name>",
				"type":         "operator",
				"version":      "<returned-version>",
				"payload":      map[string]any{"text": "hello"},
			},
			"preferred_invocation": "For newly published operator services, prefer workflow invoke-dynamic-service with the returned service.invoke.body_example; it does not require the legacy type field.",
		}
	}
	return map[string]any{
		"required": []string{"service_name", "payload"},
		"field_contract": map[string]string{
			"service_name":   "Use the published service.service_name.",
			"payload":        "Business input matching the published input_schema; do not wrap it in input or move its fields to the top level.",
			"version":        "Optional published version string; preserve it when the returned invocation contract includes it.",
			"version_number": "Optional numeric workflow version; preserve it when the returned invocation contract includes it.",
		},
		"example": map[string]any{
			"service_name": "<returned-service-name>",
			"payload":      map[string]any{"text": "hello"},
		},
		"publication_contract": "Copy the returned service.invoke.body_example and replace only its payload with representative business input. Do not add the legacy WorkItem API type field.",
	}
}

func runWorkflowServiceInvokeWorkitemAPIService(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow invoke-workitem-api-service", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	nodeID := fs.String("node-id", "", "nodeID")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow invoke-workitem-api-service [options]\n\nProduct SDK: WorkflowService.InvokeWorkitemAPIService\n\nOptions:\n")
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
	var spec sdk.WorkitemAPIServiceInvokeSpec
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &spec); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.Workflows(wid).InvokeWorkitemAPIService(context.Background(), *nodeID, spec)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceInvokeDynamicService(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow invoke-dynamic-service", flag.ExitOnError)
	return runWorkflowDynamicServiceCommand(ctx, args, fs, false)
}

func runWorkflowServiceInvokeWorkflowDynamicService(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow invoke-workflow-dynamic-service", flag.ExitOnError)
	return runWorkflowDynamicServiceCommand(ctx, args, fs, true)
}

func runWorkflowDynamicServiceCommand(ctx *Context, args []string, fs *flag.FlagSet, workflowRoute bool) error {
	commandName := fs.Name()
	sdkMethod := "WorkflowService.InvokeDynamicService"
	if workflowRoute {
		sdkMethod = "WorkflowService.InvokeWorkflowDynamicService"
	}
	filePath := fs.String("file", "", "Explicit multipart upload file path")
	fileField := fs.String("file-field", "file", "Multipart upload field")
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	bodyJSON := fs.String("json", "", "JSON object containing service_name and the business payload")
	bodyFile := fs.String("json-file", "", "JSON file containing service_name and the business payload")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli %s [options]\n\nProduct SDK: %s\n\nOptions:\n", commandName, sdkMethod)
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
	var spec sdk.DynamicServiceInvokeSpec
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &spec); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	if *filePath != "" {
		file, err := os.Open(*filePath)
		if err != nil {
			return err
		}
		defer file.Close()
		spec.Files = []sdk.MultipartFile{{FieldName: *fileField, Filename: filepath.Base(*filePath), Reader: file}}
	}
	var result *sdk.DynamicServiceInvokeResult
	if workflowRoute {
		result, err = ctx.Client.Workflows(wid).InvokeWorkflowDynamicService(context.Background(), spec)
	} else {
		result, err = ctx.Client.Workflows(wid).InvokeDynamicService(context.Background(), spec)
	}
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceEnableWorkitemAPIService(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow enable-workitem-api-service", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	nodeID := fs.String("node-id", "", "nodeID")
	version := fs.String("version", "", "version")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow enable-workitem-api-service [options]\n\nProduct SDK: WorkflowService.EnableWorkitemAPIService\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).EnableWorkitemAPIService(context.Background(), *nodeID, *version)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceDisableWorkitemAPIService(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow disable-workitem-api-service", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	nodeID := fs.String("node-id", "", "nodeID")
	version := fs.String("version", "", "version")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow disable-workitem-api-service [options]\n\nProduct SDK: WorkflowService.DisableWorkitemAPIService\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).DisableWorkitemAPIService(context.Background(), *nodeID, *version)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceCustomOperators(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow custom-operators", flag.ExitOnError)
	var queryParams queryParameters
	fs.Var(&queryParams, "query-param", "Product query parameter name=value (repeatable; pagination/filter names follow the product API)")
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow custom-operators [options]\n\nProduct SDK: WorkflowService.CustomOperators\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).CustomOperators(context.Background(), queryParams...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

type customOperatorCommandFlags struct {
	nodeID, name, description                        *string
	kind, language, handler, version, isolationLevel *string
	sourceFileID, baseNodeID, baseNodeVersion        *string
	codeFile, inputSchemaFile, outputSchemaFile      *string
	bindingConfigFile                                *string
	enabled                                          *bool
	catalogID, databaseID                            *int64
}

func bindCustomOperatorCommandFlags(fs *flag.FlagSet, create bool) customOperatorCommandFlags {
	flags := customOperatorCommandFlags{}
	if create {
		flags.nodeID = fs.String("node-id", "", "Operator WorkItem node ID")
	} else {
		flags.name = fs.String("name", "", "Updated operator name")
		flags.description = fs.String("description", "", "Updated operator description; an explicit empty value clears it")
		flags.inputSchemaFile = fs.String("input-schema-file", "", "JSON file containing the updated input schema")
		flags.outputSchemaFile = fs.String("output-schema-file", "", "JSON file containing the updated output schema")
	}
	flags.kind = fs.String("kind", "", "Operator kind")
	flags.language = fs.String("language", "", "Operator language")
	flags.handler = fs.String("handler", "", "Operator entry point")
	flags.version = fs.String("version", "", "Operator version")
	flags.isolationLevel = fs.String("isolation-level", "", "Operator isolation level")
	flags.enabled = fs.Bool("enabled", false, "Whether the operator is enabled")
	flags.codeFile = fs.String("code-file", "", "Read operator source from a local UTF-8 file")
	flags.sourceFileID = fs.String("source-file-id", "", "Operator source file ID")
	flags.baseNodeID = fs.String("base-node-id", "", "Base WorkItem node ID")
	flags.baseNodeVersion = fs.String("base-node-version", "", "Base WorkItem version")
	flags.bindingConfigFile = fs.String("binding-config-file", "", "JSON file containing the operator binding config")
	flags.catalogID = fs.Int64("catalog-id", 0, "Catalog ID")
	flags.databaseID = fs.Int64("database-id", 0, "Database ID")
	return flags
}

func loadCustomOperatorJSONFile(path string) (json.RawMessage, error) {
	raw, err := loadJSONBytes("", path)
	return json.RawMessage(raw), err
}

func runWorkflowServiceCreateCustomOperator(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow create-custom-operator", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	flags := bindCustomOperatorCommandFlags(fs, true)

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow create-custom-operator [options]\n\nProduct SDK: WorkflowService.CreateCustomOperator\n\nOptions:\n")
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
	var spec sdk.CustomOperatorSpec
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &spec); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	options := []sdk.CustomOperatorOption{}
	if *flags.codeFile != "" {
		code, err := readCustomOperatorSource(*flags.codeFile)
		if err != nil {
			return fmt.Errorf("read operator source: %w", err)
		}
		options = append(options, sdk.WithCustomOperatorCode(code))
	}
	bindingConfig, err := loadCustomOperatorJSONFile(*flags.bindingConfigFile)
	if err != nil {
		return err
	}
	fs.Visit(func(value *flag.Flag) {
		switch value.Name {
		case "node-id":
			options = append(options, sdk.WithCustomOperatorNodeID(*flags.nodeID))
		case "kind":
			options = append(options, sdk.WithCustomOperatorKind(*flags.kind))
		case "language":
			options = append(options, sdk.WithCustomOperatorLanguage(*flags.language))
		case "handler":
			options = append(options, sdk.WithCustomOperatorHandler(*flags.handler))
		case "version":
			options = append(options, sdk.WithCustomOperatorVersion(*flags.version))
		case "isolation-level":
			options = append(options, sdk.WithCustomOperatorIsolationLevel(*flags.isolationLevel))
		case "enabled":
			options = append(options, sdk.WithCustomOperatorEnabled(*flags.enabled))
		case "source-file-id":
			options = append(options, sdk.WithCustomOperatorSourceFileID(*flags.sourceFileID))
		case "base-node-id":
			options = append(options, sdk.WithCustomOperatorBaseNodeID(*flags.baseNodeID))
		case "base-node-version":
			options = append(options, sdk.WithCustomOperatorBaseNodeVersion(*flags.baseNodeVersion))
		case "binding-config-file":
			options = append(options, sdk.WithCustomOperatorBindingConfig(bindingConfig))
		case "catalog-id":
			options = append(options, sdk.WithCustomOperatorCatalogID(*flags.catalogID))
		case "database-id":
			options = append(options, sdk.WithCustomOperatorDatabaseID(*flags.databaseID))
		}
	})
	result, err := ctx.Client.Workflows(wid).CreateCustomOperator(context.Background(), spec, options...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceGetCustomOperator(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow get-custom-operator", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	operatorID := fs.String("operator-id", "", "operatorID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow get-custom-operator [options]\n\nProduct SDK: WorkflowService.GetCustomOperator\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).GetCustomOperator(context.Background(), *operatorID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceGetCustomOperatorCode(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow get-custom-operator-code", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	operatorID := fs.String("operator-id", "", "operatorID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow get-custom-operator-code [options]\n\nProduct SDK: WorkflowService.GetCustomOperatorCode\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).GetCustomOperatorCode(context.Background(), *operatorID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceGetCustomOperatorSource(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow get-custom-operator-source", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	nodeID := fs.String("node-id", "", "WorkItem node ID")
	version := fs.String("version", "", "WorkItem version")
	taskID := fs.String("task-id", "", "Optional authoring task ID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow get-custom-operator-source [options]\n\nProduct SDK: WorkflowService.GetCustomOperatorSource\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).GetCustomOperatorSource(context.Background(), *nodeID, *version, *taskID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceTestCustomOperator(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow test-custom-operator", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	operatorID := fs.String("operator-id", "", "operatorID")
	bodyJSON := fs.String("json", "", "JSON object containing the required input payload")
	bodyFile := fs.String("json-file", "", "JSON file containing the required input payload")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow test-custom-operator [options]\n\nProduct SDK: WorkflowService.TestCustomOperator\n\nOptions:\n")
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
	var input sdk.CustomOperatorTestInput
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &input); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	result, err := ctx.Client.Workflows(wid).TestCustomOperator(context.Background(), *operatorID, input)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceUpdateCustomOperator(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow update-custom-operator", flag.ContinueOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	operatorID := fs.String("operator-id", "", "operatorID")
	flags := bindCustomOperatorCommandFlags(fs, false)

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow update-custom-operator [options]\n\nProduct SDK: WorkflowService.UpdateCustomOperator\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	options := []sdk.CustomOperatorUpdateOption{}
	if *flags.codeFile != "" {
		code, err := readCustomOperatorSource(*flags.codeFile)
		if err != nil {
			return fmt.Errorf("read operator source: %w", err)
		}
		options = append(options, sdk.WithCustomOperatorUpdateCode(code))
	}
	inputSchema, err := loadCustomOperatorJSONFile(*flags.inputSchemaFile)
	if err != nil {
		return err
	}
	outputSchema, err := loadCustomOperatorJSONFile(*flags.outputSchemaFile)
	if err != nil {
		return err
	}
	bindingConfig, err := loadCustomOperatorJSONFile(*flags.bindingConfigFile)
	if err != nil {
		return err
	}
	fs.Visit(func(value *flag.Flag) {
		switch value.Name {
		case "kind":
			options = append(options, sdk.WithCustomOperatorUpdateKind(*flags.kind))
		case "language":
			options = append(options, sdk.WithCustomOperatorUpdateLanguage(*flags.language))
		case "handler":
			options = append(options, sdk.WithCustomOperatorUpdateHandler(*flags.handler))
		case "version":
			options = append(options, sdk.WithCustomOperatorUpdateVersion(*flags.version))
		case "isolation-level":
			options = append(options, sdk.WithCustomOperatorUpdateIsolationLevel(*flags.isolationLevel))
		case "input-schema-file":
			options = append(options, sdk.WithCustomOperatorUpdateInputSchema(inputSchema))
		case "output-schema-file":
			options = append(options, sdk.WithCustomOperatorUpdateOutputSchema(outputSchema))
		case "enabled":
			options = append(options, sdk.WithCustomOperatorUpdateEnabled(*flags.enabled))
		case "source-file-id":
			options = append(options, sdk.WithCustomOperatorUpdateSourceFileID(*flags.sourceFileID))
		case "base-node-id":
			options = append(options, sdk.WithCustomOperatorUpdateBaseNodeID(*flags.baseNodeID))
		case "base-node-version":
			options = append(options, sdk.WithCustomOperatorUpdateBaseNodeVersion(*flags.baseNodeVersion))
		case "binding-config-file":
			options = append(options, sdk.WithCustomOperatorUpdateBindingConfig(bindingConfig))
		case "catalog-id":
			options = append(options, sdk.WithCustomOperatorUpdateCatalogID(*flags.catalogID))
		case "database-id":
			options = append(options, sdk.WithCustomOperatorUpdateDatabaseID(*flags.databaseID))
		case "name":
			options = append(options, sdk.WithCustomOperatorUpdateName(*flags.name))
		case "description":
			options = append(options, sdk.WithCustomOperatorUpdateDescription(*flags.description))
		}
	})
	result, err := ctx.Client.Workflows(wid).UpdateCustomOperator(context.Background(), *operatorID, options...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceEnableCustomOperator(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow enable-custom-operator", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	operatorID := fs.String("operator-id", "", "operatorID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow enable-custom-operator [options]\n\nProduct SDK: WorkflowService.EnableCustomOperator\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).EnableCustomOperator(context.Background(), *operatorID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceDisableCustomOperator(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow disable-custom-operator", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	operatorID := fs.String("operator-id", "", "operatorID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow disable-custom-operator [options]\n\nProduct SDK: WorkflowService.DisableCustomOperator\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).DisableCustomOperator(context.Background(), *operatorID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceDeleteCustomOperator(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow delete-custom-operator", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	operatorID := fs.String("operator-id", "", "operatorID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow delete-custom-operator [options]\n\nProduct SDK: WorkflowService.DeleteCustomOperator\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).DeleteCustomOperator(context.Background(), *operatorID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceCatalogFileLineageOverview(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow catalog-file-lineage-overview", flag.ExitOnError)
	var queryParams queryParameters
	fs.Var(&queryParams, "query-param", "Product query parameter name=value (repeatable; pagination/filter names follow the product API)")
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fileID := fs.String("file-id", "", "fileID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow catalog-file-lineage-overview [options]\n\nProduct SDK: WorkflowService.CatalogFileLineageOverview\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).CatalogFileLineageOverview(context.Background(), *fileID, queryParams...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceArtifactLineageOverview(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow artifact-lineage-overview", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	artifactID := fs.String("artifact-id", "", "artifactID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow artifact-lineage-overview [options]\n\nProduct SDK: WorkflowService.ArtifactLineageOverview\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).ArtifactLineageOverview(context.Background(), *artifactID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceArtifactOutputBlocks(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow artifact-output-blocks", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	artifactID := fs.String("artifact-id", "", "artifactID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow artifact-output-blocks [options]\n\nProduct SDK: WorkflowService.ArtifactOutputBlocks\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).ArtifactOutputBlocks(context.Background(), *artifactID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceArtifactBlockRevisions(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow artifact-block-revisions", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	artifactID := fs.String("artifact-id", "", "artifactID")
	blockID := fs.String("block-id", "", "blockID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow artifact-block-revisions [options]\n\nProduct SDK: WorkflowService.ArtifactBlockRevisions\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).ArtifactBlockRevisions(context.Background(), *artifactID, *blockID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

type lineageRevisionCommandFlags struct {
	baseRevisionID, expectedBaseContentHash *string
	revisionContentRef, revisionContentHash *string
	revisionContentPayload, patchJSON       *string
	useOriginal, useCurrentEffective        *bool
	expectedEffectiveSetVersion             *int64
}

func addLineageRevisionCommandFlags(fs *flag.FlagSet) lineageRevisionCommandFlags {
	return lineageRevisionCommandFlags{
		baseRevisionID:              fs.String("base-revision-id", "", "Base revision ID"),
		useOriginal:                 fs.Bool("use-original", false, "Use original content as the base"),
		useCurrentEffective:         fs.Bool("use-current-effective", false, "Use the current effective revision as the base"),
		expectedEffectiveSetVersion: fs.Int64("expected-effective-set-version", 0, "Expected effective set version"),
		expectedBaseContentHash:     fs.String("expected-base-content-hash", "", "Expected base content hash"),
		revisionContentRef:          fs.String("revision-content-ref", "", "New revision content reference"),
		revisionContentHash:         fs.String("revision-content-hash", "", "New revision content hash"),
		revisionContentPayload:      fs.String("revision-content-payload", "", "New revision full content; an explicit empty value is preserved"),
		patchJSON:                   fs.String("patch-json", "", "Revision patch JSON string"),
	}
}

func (flags lineageRevisionCommandFlags) options(fs *flag.FlagSet) []sdk.LineageRevisionOption {
	var options []sdk.LineageRevisionOption
	fs.Visit(func(value *flag.Flag) {
		switch value.Name {
		case "base-revision-id":
			options = append(options, sdk.WithLineageBaseRevisionID(*flags.baseRevisionID))
		case "use-original":
			options = append(options, sdk.WithLineageUseOriginal(*flags.useOriginal))
		case "use-current-effective":
			options = append(options, sdk.WithLineageUseCurrentEffective(*flags.useCurrentEffective))
		case "expected-effective-set-version":
			options = append(options, sdk.WithLineageExpectedEffectiveSetVersion(*flags.expectedEffectiveSetVersion))
		case "expected-base-content-hash":
			options = append(options, sdk.WithLineageExpectedBaseContentHash(*flags.expectedBaseContentHash))
		case "revision-content-ref":
			options = append(options, sdk.WithLineageRevisionContentRef(*flags.revisionContentRef))
		case "revision-content-hash":
			options = append(options, sdk.WithLineageRevisionContentHash(*flags.revisionContentHash))
		case "revision-content-payload":
			options = append(options, sdk.WithLineageRevisionContentPayload(*flags.revisionContentPayload))
		case "patch-json":
			options = append(options, sdk.WithLineagePatchJSON(*flags.patchJSON))
		}
	})
	return options
}

func runWorkflowServiceCreateArtifactBlockRevision(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow create-artifact-block-revision", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	artifactID := fs.String("artifact-id", "", "artifactID")
	blockID := fs.String("block-id", "", "blockID")
	revisionFlags := addLineageRevisionCommandFlags(fs)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow create-artifact-block-revision [options]\n\nProduct SDK: WorkflowService.CreateArtifactBlockRevision\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).CreateArtifactBlockRevision(context.Background(), *artifactID, *blockID, revisionFlags.options(fs)...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceSwitchArtifactEffectiveRevisions(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow switch-artifact-effective-revisions", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	artifactID := fs.String("artifact-id", "", "artifactID")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	expectedVersion := fs.Int64("expected-effective-set-version", 0, "Expected effective set version")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow switch-artifact-effective-revisions [options]\n\nProduct SDK: WorkflowService.SwitchArtifactEffectiveRevisions\n\nOptions:\n")
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
	var switches []sdk.LineageRevisionSwitch
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &switches); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	var options []sdk.LineageSwitchOption
	fs.Visit(func(value *flag.Flag) {
		if value.Name == "expected-effective-set-version" {
			options = append(options, sdk.WithLineageSwitchExpectedEffectiveSetVersion(*expectedVersion))
		}
	})
	result, err := ctx.Client.Workflows(wid).SwitchArtifactEffectiveRevisions(context.Background(), *artifactID, switches, options...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceArtifactNode(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow artifact-node", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	artifactID := fs.String("artifact-id", "", "artifactID")
	nodeID := fs.String("node-id", "", "nodeID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow artifact-node [options]\n\nProduct SDK: WorkflowService.ArtifactNode\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).ArtifactNode(context.Background(), *artifactID, *nodeID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceArtifactNodeOutputBlocks(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow artifact-node-output-blocks", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	artifactID := fs.String("artifact-id", "", "artifactID")
	nodeID := fs.String("node-id", "", "nodeID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow artifact-node-output-blocks [options]\n\nProduct SDK: WorkflowService.ArtifactNodeOutputBlocks\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).ArtifactNodeOutputBlocks(context.Background(), *artifactID, *nodeID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceArtifactNodeBlockRevisions(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow artifact-node-block-revisions", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	artifactID := fs.String("artifact-id", "", "artifactID")
	nodeID := fs.String("node-id", "", "nodeID")
	blockID := fs.String("block-id", "", "blockID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow artifact-node-block-revisions [options]\n\nProduct SDK: WorkflowService.ArtifactNodeBlockRevisions\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).ArtifactNodeBlockRevisions(context.Background(), *artifactID, *nodeID, *blockID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceCreateArtifactNodeBlockRevision(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow create-artifact-node-block-revision", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	artifactID := fs.String("artifact-id", "", "artifactID")
	nodeID := fs.String("node-id", "", "nodeID")
	blockID := fs.String("block-id", "", "blockID")
	revisionFlags := addLineageRevisionCommandFlags(fs)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow create-artifact-node-block-revision [options]\n\nProduct SDK: WorkflowService.CreateArtifactNodeBlockRevision\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).CreateArtifactNodeBlockRevision(context.Background(), *artifactID, *nodeID, *blockID, revisionFlags.options(fs)...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceSwitchArtifactNodeEffectiveRevisions(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow switch-artifact-node-effective-revisions", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	artifactID := fs.String("artifact-id", "", "artifactID")
	nodeID := fs.String("node-id", "", "nodeID")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	expectedVersion := fs.Int64("expected-effective-set-version", 0, "Expected effective set version")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow switch-artifact-node-effective-revisions [options]\n\nProduct SDK: WorkflowService.SwitchArtifactNodeEffectiveRevisions\n\nOptions:\n")
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
	var switches []sdk.LineageRevisionSwitch
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &switches); err != nil {
			return fmt.Errorf("parse --json: %w", err)
		}
	}
	var options []sdk.LineageSwitchOption
	fs.Visit(func(value *flag.Flag) {
		if value.Name == "expected-effective-set-version" {
			options = append(options, sdk.WithLineageSwitchExpectedEffectiveSetVersion(*expectedVersion))
		}
	})
	result, err := ctx.Client.Workflows(wid).SwitchArtifactNodeEffectiveRevisions(context.Background(), *artifactID, *nodeID, switches, options...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

type lineageRerunCommandFlags struct {
	mode                                         *string
	configOverrideFile, inputPayloadOverrideFile *string
	varsPayloadOverrideFile                      *string
}

func addLineageRerunCommandFlags(fs *flag.FlagSet) lineageRerunCommandFlags {
	return lineageRerunCommandFlags{
		mode:                     fs.String("mode", "", "Lineage rerun mode"),
		configOverrideFile:       fs.String("config-override-file", "", "JSON file containing the config override"),
		inputPayloadOverrideFile: fs.String("input-payload-override-file", "", "JSON file containing the input payload override"),
		varsPayloadOverrideFile:  fs.String("vars-payload-override-file", "", "JSON file containing the vars payload override"),
	}
}

func (flags lineageRerunCommandFlags) options(fs *flag.FlagSet) ([]sdk.LineageRerunOption, error) {
	var options []sdk.LineageRerunOption
	fs.Visit(func(value *flag.Flag) {
		if value.Name == "mode" {
			options = append(options, sdk.WithLineageRerunMode(*flags.mode))
		}
	})
	for _, item := range []struct {
		path   string
		option func(json.RawMessage) sdk.LineageRerunOption
	}{
		{*flags.configOverrideFile, sdk.WithLineageRerunConfigOverride},
		{*flags.inputPayloadOverrideFile, sdk.WithLineageRerunInputPayloadOverride},
		{*flags.varsPayloadOverrideFile, sdk.WithLineageRerunVarsPayloadOverride},
	} {
		if item.path == "" {
			continue
		}
		raw, err := loadJSONBytes("", item.path)
		if err != nil {
			return nil, err
		}
		options = append(options, item.option(json.RawMessage(raw)))
	}
	return options, nil
}

func runWorkflowServiceCreateArtifactNodeRerun(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow create-artifact-node-rerun", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	artifactID := fs.String("artifact-id", "", "artifactID")
	nodeID := fs.String("node-id", "", "nodeID")
	rerunFlags := addLineageRerunCommandFlags(fs)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow create-artifact-node-rerun [options]\n\nProduct SDK: WorkflowService.CreateArtifactNodeRerun\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	options, err := rerunFlags.options(fs)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).CreateArtifactNodeRerun(context.Background(), *artifactID, *nodeID, options...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceCreateTargetArtifactNodeRevisionRerun(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow create-target-artifact-node-revision-rerun", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	artifactID := fs.String("artifact-id", "", "artifactID")
	nodeID := fs.String("node-id", "", "nodeID")
	targetOutputEmissionID := fs.String("target-output-emission-id", "", "targetOutputEmissionID")
	mode := fs.String("mode", "", "Revision rerun mode")
	sourceEffectiveSetVersion := fs.Int64("source-effective-set-version", 0, "Source effective set version")
	selectedRevisionsJSONFile := fs.String("selected-revisions-json-file", "", "JSON file containing selected revisions")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow create-target-artifact-node-revision-rerun [options]\n\nProduct SDK: WorkflowService.CreateTargetArtifactNodeRevisionRerun\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	var options []sdk.LineageRevisionRerunOption
	fs.Visit(func(value *flag.Flag) {
		switch value.Name {
		case "mode":
			options = append(options, sdk.WithLineageRevisionRerunMode(*mode))
		case "source-effective-set-version":
			options = append(options, sdk.WithLineageRevisionRerunSourceEffectiveSetVersion(*sourceEffectiveSetVersion))
		}
	})
	if *selectedRevisionsJSONFile != "" {
		raw, err := loadJSONBytes("", *selectedRevisionsJSONFile)
		if err != nil {
			return err
		}
		options = append(options, sdk.WithLineageSelectedRevisionsJSON(json.RawMessage(raw)))
	}
	result, err := ctx.Client.Workflows(wid).CreateTargetArtifactNodeRevisionRerun(context.Background(), *artifactID, *nodeID, *targetOutputEmissionID, options...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceWorkflowTemplates(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow workflow-templates", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow workflow-templates [options]\n\nProduct SDK: WorkflowService.WorkflowTemplates\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).WorkflowTemplates(context.Background())
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceGetWorkflowTemplate(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow get-workflow-template", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	templateID := fs.String("template-id", "", "templateID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow get-workflow-template [options]\n\nProduct SDK: WorkflowService.GetWorkflowTemplate\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).GetWorkflowTemplate(context.Background(), *templateID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceCreateWorkflowTemplate(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow create-workflow-template", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	addWorkflowTemplateFlags(fs)
	name := fs.String("name", "", "name")
	dslYAML := fs.String("dsl-yaml", "", "dslYAML")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow create-workflow-template [options]\n\nProduct SDK: WorkflowService.CreateWorkflowTemplate\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).CreateWorkflowTemplate(context.Background(), *name, *dslYAML, workflowTemplateCLIOptions(fs)...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceDeployMOWL(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow deploy-mowl", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	name := fs.String("name", "", "name")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow deploy-mowl [options]\n\nProduct SDK: WorkflowService.DeployMOWL\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	workflow, err := loadWorkflowMOWLGraph(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).DeployMOWL(context.Background(), *name, workflow)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceCreateMOWLWorkflowTemplate(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow create-mowl-workflow-template", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	addWorkflowTemplateFlags(fs)
	name := fs.String("name", "", "name")
	bodyJSON := fs.String("json", "", "JSON for complex arguments")
	bodyFile := fs.String("json-file", "", "JSON file for complex arguments")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow create-mowl-workflow-template [options]\n\nProduct SDK: WorkflowService.CreateMOWLWorkflowTemplate\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	workflow, err := loadWorkflowMOWLGraph(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).CreateMOWLWorkflowTemplate(context.Background(), *name, workflow, workflowTemplateCLIOptions(fs)...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceUpdateWorkflowTemplate(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow update-workflow-template", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	addWorkflowTemplateFlags(fs)
	templateID := fs.String("template-id", "", "templateID")
	name := fs.String("name", "", "name")
	dslYAML := fs.String("dsl-yaml", "", "dslYAML")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow update-workflow-template [options]\n\nProduct SDK: WorkflowService.UpdateWorkflowTemplate\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).UpdateWorkflowTemplate(context.Background(), *templateID, *name, *dslYAML, workflowTemplateCLIOptions(fs)...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceUpdateMOWLWorkflowTemplate(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow update-mowl-workflow-template", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	templateID := fs.String("template-id", "", "Template ID")
	name := fs.String("name", "", "Template name")
	bodyJSON := fs.String("json", "", "JSON MOWL publication graph")
	bodyFile := fs.String("json-file", "", "JSON file containing the MOWL graph")
	addWorkflowTemplateFlags(fs)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: moi-cli workflow update-mowl-workflow-template [options]\n\nProduct SDK: WorkflowService.UpdateMOWLWorkflowTemplate\n\nOptions:")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	workflow, err := loadWorkflowMOWLGraph(*bodyJSON, *bodyFile)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).UpdateMOWLWorkflowTemplate(context.Background(), *templateID, *name, workflow, workflowTemplateCLIOptions(fs)...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func addWorkflowTemplateFlags(fs *flag.FlagSet) {
	fs.String("template-key", "", "Stable builtin template key (ignored for custom templates)")
	fs.String("description", "", "Template description")
	fs.String("input-form", "", "JSON-encoded runtime fields string; empty clears the form")
	fs.Bool("is-builtin", false, "Builtin template flag; the server validates creation and type changes")
}

func workflowTemplateCLIOptions(fs *flag.FlagSet) []sdk.WorkflowTemplateOption {
	var opts []sdk.WorkflowTemplateOption
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "template-key":
			opts = append(opts, sdk.WithWorkflowTemplateKey(f.Value.String()))
		case "description":
			opts = append(opts, sdk.WithWorkflowTemplateDescription(f.Value.String()))
		case "input-form":
			opts = append(opts, sdk.WithWorkflowTemplateInputForm(f.Value.String()))
		case "is-builtin":
			opts = append(opts, sdk.WithWorkflowTemplateBuiltin(f.Value.(flag.Getter).Get().(bool)))
		}
	})
	return opts
}

func loadWorkflowMOWLGraph(bodyJSON, bodyFile string) (map[string]any, error) {
	rawBody, err := loadJSONBytes(bodyJSON, bodyFile)
	if err != nil {
		return nil, err
	}
	workflow := map[string]any{}
	if len(rawBody) > 0 {
		if err := json.Unmarshal(rawBody, &workflow); err != nil {
			return nil, fmt.Errorf("parse --json: %w", err)
		}
	}
	return workflow, nil
}

func runWorkflowServiceDeleteWorkflowTemplate(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow delete-workflow-template", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "Workspace ID (default: from config)")
	templateID := fs.String("template-id", "", "templateID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow delete-workflow-template [options]\n\nProduct SDK: WorkflowService.DeleteWorkflowTemplate\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).DeleteWorkflowTemplate(context.Background(), *templateID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceHistory(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow history", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	contextID := fs.String("context-id", "", "context-id")
	limit := fs.Int("limit", 20, "limit")
	offset := fs.Int("offset", 0, "offset")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow history [options]\n\nProduct SDK: WorkflowService.History\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).History(context.Background(), *contextID, *limit, *offset)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceRejectCandidateProposal(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow reject-candidate-proposal", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	proposalID := fs.String("proposal-id", "", "proposal-id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow reject-candidate-proposal [options]\n\nProduct SDK: WorkflowService.RejectCandidateProposal\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).RejectCandidateProposal(context.Background(), *proposalID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceRestoreHistory(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow restore-history", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	contextID := fs.String("context-id", "", "context-id")
	requestID := fs.String("request-id", "", "request-id")
	taskID := fs.String("task-id", "", "task-id")
	artifactID := fs.String("artifact-id", "", "artifact-id")
	parentTaskID := fs.String("parent-task-id", "", "parent history task ID")
	parentArtifactID := fs.String("parent-artifact-id", "", "parent history artifact ID")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow restore-history [options]\n\nProduct SDK: WorkflowService.RestoreHistory\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	request := sdk.WorkflowHistoryRestoreRequest{ContextID: *contextID, RequestID: *requestID, CandidateRef: sdk.WorkflowCandidateRef{TaskID: *taskID, ArtifactID: *artifactID}}
	if *parentTaskID != "" || *parentArtifactID != "" {
		request.ParentRef = &sdk.WorkflowCandidateRef{TaskID: *parentTaskID, ArtifactID: *parentArtifactID}
	}
	result, err := ctx.Client.Workflows(wid).RestoreHistory(context.Background(), request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceApps(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow apps", flag.ExitOnError)
	var queryParams queryParameters
	fs.Var(&queryParams, "query-param", "Product query parameter name=value (repeatable; pagination/filter names follow the product API)")
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow apps [options]\n\nProduct SDK: WorkflowService.Apps\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).Apps(context.Background(), queryParams...)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceSubmitCandidate(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow submit-candidate", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	projectDir := fs.String("project-dir", "", "Workflow authoring directory in this filesystem (requires a live runtime task)")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow submit-candidate [options]\n\nProduct SDK: WorkflowService.SubmitCandidate\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	rawBody, err := workflowCandidateBody(*bodyJSON, *bodyFile, *projectDir)
	if err != nil {
		return err
	}
	var request sdk.WorkflowCandidateSubmitRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Workflows(wid).SubmitCandidate(context.Background(), &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceTestCandidate(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow test-candidate", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	projectDir := fs.String("project-dir", "", "Workflow authoring directory in this filesystem (requires a live runtime task)")
	caseName := fs.String("case", "", "One JSON filename under project tests/cases (diagnostic only)")
	verificationOnly := fs.Bool("verification-only", false, "Re-run project verification against the preceding suite execution")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow test-candidate [options]\n\nProduct SDK: WorkflowService.TestCandidate\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *projectDir == "" && (*caseName != "" || *verificationOnly) {
		return fmt.Errorf("--case and --verification-only require --project-dir")
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	rawBody, err := workflowCandidateBody(*bodyJSON, *bodyFile, *projectDir)
	if err != nil {
		return err
	}
	var request sdk.WorkflowCandidateTestRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	if *projectDir != "" {
		request.Case, request.VerificationOnly = *caseName, *verificationOnly
	}
	result, err := ctx.Client.Workflows(wid).TestCandidate(context.Background(), &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

// Send authored files, never a sandbox path for the server to interpret.
// References and editable compiler/test output are not trusted evidence.
func workflowCandidateBody(bodyJSON, bodyFile, projectDir string) ([]byte, error) {
	if projectDir == "" {
		return loadJSONBytes(bodyJSON, bodyFile)
	}
	if bodyJSON != "" || bodyFile != "" {
		return nil, fmt.Errorf("--project-dir cannot be combined with --json or --json-file")
	}
	info, err := os.Lstat(projectDir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("--project-dir must be a directory")
	}
	files := make(map[string][]byte)
	for _, directory := range []string{"workflow", "operators", "tests"} {
		root := filepath.Join(projectDir, directory)
		if _, err := os.Lstat(root); err != nil {
			// Session checkpoints contain files; an unused optional directory
			// has no entry until the author creates its first operator or case.
			if os.IsNotExist(err) && directory != "workflow" {
				continue
			}
			return nil, err
		}
		err := filepath.WalkDir(root, func(name string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				return nil
			}
			if !entry.Type().IsRegular() {
				return fmt.Errorf("project contains a non-regular file: %s", name)
			}
			relative, err := filepath.Rel(projectDir, name)
			if err != nil {
				return err
			}
			content, err := os.ReadFile(name)
			if err == nil {
				files[filepath.ToSlash(relative)] = content
			}
			return err
		})
		if err != nil {
			return nil, err
		}
	}
	return json.Marshal(struct {
		Files map[string][]byte `json:"project_files"`
	}{files})
}

func runWorkflowServiceEnsureTestScratch(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow ensure-test-scratch", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow ensure-test-scratch [options]\n\nProduct SDK: WorkflowService.EnsureTestScratch\n\nOptions:\n")
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
	var request sdk.WorkflowScratchScope
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Workflows(wid).EnsureTestScratch(context.Background(), &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceObserveTestScratch(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow observe-test-scratch", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	bodyJSON := fs.String("json", "", "json")
	bodyFile := fs.String("json-file", "", "json-file")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow observe-test-scratch [options]\n\nProduct SDK: WorkflowService.ObserveTestScratch\n\nOptions:\n")
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
	var request sdk.WorkflowScratchScope
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Workflows(wid).ObserveTestScratch(context.Background(), &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceResetTestScratch(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow reset-test-scratch", flag.ExitOnError)
	workspaceID := fs.String("workspace-id", "", "workspace-id")
	taskID := fs.String("task-id", "", "task-id")
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: moi-cli workflow reset-test-scratch [options]\n\nProduct SDK: WorkflowService.ResetTestScratch\n\nOptions:\n")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).ResetTestScratch(context.Background(), *taskID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceVersions(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow versions", flag.ExitOnError)
	setSDKCommandUsage(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	workflowID := fs.String("workflow-id", "", "workflow-id")
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).Versions(context.Background(), *workflowID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceLatestPublishedVersion(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow latest-published-version", flag.ExitOnError)
	setSDKCommandUsage(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	workflowID := fs.String("workflow-id", "", "workflow-id")
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).LatestPublishedVersion(context.Background(), *workflowID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceVersion(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow version", flag.ExitOnError)
	setSDKCommandUsage(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	workflowID := fs.String("workflow-id", "", "workflow-id")
	versionID := fs.String("version-id", "", "version-id")
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).Version(context.Background(), *workflowID, *versionID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServicePublishVersion(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow publish-version", flag.ExitOnError)
	setSDKCommandUsage(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	workflowID := fs.String("workflow-id", "", "workflow-id")
	versionID := fs.String("version-id", "", "version-id")
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).PublishVersion(context.Background(), *workflowID, *versionID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceDeprecateVersion(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow deprecate-version", flag.ExitOnError)
	setSDKCommandUsage(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	workflowID := fs.String("workflow-id", "", "workflow-id")
	versionID := fs.String("version-id", "", "version-id")
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).DeprecateVersion(context.Background(), *workflowID, *versionID)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceDeleteVersion(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow delete-version", flag.ExitOnError)
	setSDKCommandUsage(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	workflowID := fs.String("workflow-id", "", "workflow-id")
	versionID := fs.String("version-id", "", "version-id")
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	return ctx.Client.Workflows(wid).DeleteVersion(context.Background(), *workflowID, *versionID)
}

func runWorkflowServiceVersionByNumber(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow version-by-number", flag.ExitOnError)
	setSDKCommandUsage(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	workflowID := fs.String("workflow-id", "", "workflow-id")
	number := fs.Int("number", 0, "number")
	if err := fs.Parse(args); err != nil {
		return err
	}
	wid, err := resolveWorkspaceID(ctx, *workspaceID)
	if err != nil {
		return err
	}
	result, err := ctx.Client.Workflows(wid).VersionByNumber(context.Background(), *workflowID, *number)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

func runWorkflowServiceCreateVersion(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("workflow create-version", flag.ExitOnError)
	setSDKCommandUsage(fs)
	workspaceID := fs.String("workspace-id", "", "Workspace ID")
	workflowID := fs.String("workflow-id", "", "workflow-id")
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
	var request sdk.WorkflowVersionCreateRequest
	if err := json.Unmarshal(rawBody, &request); err != nil {
		return fmt.Errorf("parse --json: %w", err)
	}
	result, err := ctx.Client.Workflows(wid).CreateVersion(context.Background(), *workflowID, &request)
	if err != nil {
		return err
	}
	return printJSONOrTable(ctx, result, func() error { return printJSON(result) })
}

// Source reading is shared by create and update; a failed local read never
// turns into a metadata-only mutation or silently changes source bytes.
func readCustomOperatorSource(name string) (string, error) {
	code, err := os.ReadFile(name)
	if err != nil {
		return "", err
	}
	if !utf8.Valid(code) {
		return "", fmt.Errorf("operator source must be UTF-8")
	}
	return string(code), nil
}
