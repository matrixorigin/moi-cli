# Existing workflows

Use the workspace selected under SKILL.md. To find an existing workflow:

```text
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow list
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow get --workflow-id "$workflow_id"
```

Resolve `$workflow_id` from `workflows[].id` by the user's selection, not list
order. Inspect `workflow` in the get response for runnable state and configured
inputs. A name is not an execution ID. Remote jobs consume service references,
not laptop paths; use [files](files.md) when input upload is actually needed.

## Run a configured workflow

For an explicitly requested run of a ready, configured workflow:

```text
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow run --workflow-id "$workflow_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow get-run --workflow-id "$workflow_id" --execution-id "$execution_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" workflow get-run-result --workflow-id "$workflow_id" --execution-id "$execution_id"
```

For a confirmed single execution, use its real ID as `$execution_id`. For Volume inputs,
first follow the batch join described below. Get-run
returns `execution.status`; inspect it before fetching final results and poll
only within a finite wait budget. Get-run-result returns `result.status`,
`result.case_result`, `result.case_error` and `result.node_states`. Distinguish
failed/cancelled/still-running executions; inspect any expected output artifact
before saying it is available. Never substitute `moi_task_id` for `execution_id`.
Carry the selected workspace ID on all these calls, even when the run used an
explicit per-command override. Resume inspection using the returned execution
ID; a failed status read does not authorize another run. Download output only
when requested or needed for the requested analysis, not as a completion ritual.

The simple `run` command does not accept arbitrary JSON inputs. For a different
execution mode, use the matching [workflow reference](commands/workflow.md),
not guessed flags. Cancel, rerun, authoring, custom operators and API publishing
are separate tasks, not automatic steps in running an existing workflow.

When the Workflow consumes a Volume, follow the [cross-domain lineage recipe](recipes/catalog-workflow-knowledge-pipeline.md).
Production Volume runs can return the dispatch job ID inside `workflow_run.execution_id`,
even with `execution_mode=one_shot`. Call `workflow list-runs --workflow-id`, match
`executions[].dispatch_job_id` to that returned value, then use each matched record's
`execution_id` for `get-run*`. Check the exact input file set and version; a 404 on
the dispatch ID or an empty `list-file-executions` response does not prove processing failed.
