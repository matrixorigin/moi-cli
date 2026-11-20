# workflow command reference

Generated from CLI source. Do not edit. Business requirements remain server-owned.
Use `moi-cli schema DOMAIN COMMAND` for typed JSON input and `--help` for defaults.

## workflow accept-candidate

- `--artifact-id`: server-owned workflow.candidate artifact ID
- `--idempotency-key`: caller-owned acceptance idempotency key
- `--proposal-id`: server-owned pending proposal ID
- `--task-id`: authenticated A2A task ID
- `--workspace-id`: Workspace ID (default: from config)

## workflow apps

- `--query-param`: Product query parameter name=value (repeatable; pagination/filter names follow the product API)
- `--workspace-id`: workspace-id

## workflow artifact-block-revisions

- `--artifact-id`: artifactID
- `--block-id`: blockID
- `--workspace-id`: Workspace ID (default: from config)

## workflow artifact-lineage-overview

- `--artifact-id`: artifactID
- `--workspace-id`: Workspace ID (default: from config)

## workflow artifact-node

- `--artifact-id`: artifactID
- `--node-id`: nodeID
- `--workspace-id`: Workspace ID (default: from config)

## workflow artifact-node-block-revisions

- `--artifact-id`: artifactID
- `--block-id`: blockID
- `--node-id`: nodeID
- `--workspace-id`: Workspace ID (default: from config)

## workflow artifact-node-output-blocks

- `--artifact-id`: artifactID
- `--node-id`: nodeID
- `--workspace-id`: Workspace ID (default: from config)

## workflow artifact-output-blocks

- `--artifact-id`: artifactID
- `--workspace-id`: Workspace ID (default: from config)

## workflow batch-retry-failed-files

- `--all-failed`: Retry every currently failed file in the source dispatch job
- `--execution-ids`: Comma-separated failed execution IDs
- `--request-id`: Client-generated idempotency key
- `--source-dispatch-job-id`: Source Volume dispatch job ID
- `--workflow-id`: Workflow ID
- `--workspace-id`: Workspace ID (default: from config)

## workflow cancel-run

- `--execution-id`: executionID
- `--reason`: Cancellation reason
- `--workflow-id`: workflowID
- `--workspace-id`: Workspace ID (default: from config)

## workflow cancel-run-rerun

- `--execution-id`: executionID
- `--reason`: Cancellation reason
- `--rerun-id`: rerunID
- `--workflow-id`: workflowID
- `--workspace-id`: Workspace ID (default: from config)

## workflow catalog-file-lineage-overview

- `--file-id`: fileID
- `--query-param`: Product query parameter name=value (repeatable; pagination/filter names follow the product API)
- `--workspace-id`: Workspace ID (default: from config)

## workflow create-artifact-block-revision

- `--artifact-id`: artifactID
- `--block-id`: blockID
- `--workspace-id`: Workspace ID (default: from config)

## workflow create-artifact-node-block-revision

- `--artifact-id`: artifactID
- `--block-id`: blockID
- `--node-id`: nodeID
- `--workspace-id`: Workspace ID (default: from config)

## workflow create-artifact-node-rerun

- `--artifact-id`: artifactID
- `--node-id`: nodeID
- `--workspace-id`: Workspace ID (default: from config)

## workflow create-custom-operator

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--workspace-id`: Workspace ID (default: from config)

## workflow create-mowl-workflow-template

- `--description`: Template description
- `--input-form`: JSON-encoded runtime fields string; empty clears the form
- `--is-builtin`: Builtin template flag; the server validates creation and type changes
- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--name`: name
- `--template-key`: Stable builtin template key (ignored for custom templates)
- `--workspace-id`: Workspace ID (default: from config)

## workflow create-run-node-rerun

- `--config-override-file`: JSON file containing the rerun config override
- `--execution-id`: executionID
- `--expected-rerun-contract-hash`: Required rerun contract hash returned by get-run-node-rerun-plan
- `--mode`: mode
- `--node-key`: nodeKey
- `--reason`: Rerun reason
- `--source-node-execution-id`: sourceNodeExecutionID
- `--values-file`: JSON object containing rerun runtime values
- `--workflow-id`: workflowID
- `--workspace-id`: Workspace ID (default: from config)

## workflow create-target-artifact-node-revision-rerun

- `--artifact-id`: artifactID
- `--mode`: Revision rerun mode
- `--node-id`: nodeID
- `--selected-revisions-json-file`: JSON file containing selected revisions
- `--source-effective-set-version`: Source effective set version
- `--target-output-emission-id`: targetOutputEmissionID
- `--workspace-id`: Workspace ID (default: from config)

## workflow create-version

- `--json`: Request JSON
- `--json-file`: Request JSON file
- `--workflow-id`: workflow-id
- `--workspace-id`: Workspace ID

## workflow create-worker-task

- `--json`: Request JSON
- `--json-file`: Request JSON file
- `--workspace-id`: Workspace ID

## workflow create-workflow-template

- `--description`: Template description
- `--dsl-yaml`: dslYAML
- `--input-form`: JSON-encoded runtime fields string; empty clears the form
- `--is-builtin`: Builtin template flag; the server validates creation and type changes
- `--name`: name
- `--template-key`: Stable builtin template key (ignored for custom templates)
- `--workspace-id`: Workspace ID (default: from config)

## workflow custom-operators

- `--query-param`: Product query parameter name=value (repeatable; pagination/filter names follow the product API)
- `--workspace-id`: Workspace ID (default: from config)

## workflow delete

- `--workflow-id`: workflowID
- `--workspace-id`: Workspace ID (default: from config)

## workflow delete-custom-operator

- `--operator-id`: operatorID
- `--workspace-id`: Workspace ID (default: from config)

## workflow delete-run

- `--execution-id`: executionID
- `--workflow-id`: workflowID
- `--workspace-id`: Workspace ID (default: from config)

## workflow delete-version

- `--version-id`: version-id
- `--workflow-id`: workflow-id
- `--workspace-id`: Workspace ID

## workflow delete-workflow-template

- `--template-id`: templateID
- `--workspace-id`: Workspace ID (default: from config)

## workflow deploy

- `--cron-expression`: Cron expression; see workflow update --help for syntax, timezone and examples
- `--default-values-file`: Workflow default values JSON file
- `--description`: Workflow description
- `--dsl-yaml`: dslYAML
- `--dynamic-service-file`: Dynamic service configuration JSON file
- `--execution-mode`: Workflow execution mode: one_shot, cron, volume_trigger or dynamic_service
- `--name`: name
- `--runtime-fields-file`: Runtime parameter fields JSON file; see workflow deploy --help for an example
- `--volume-trigger-file`: Volume trigger configuration JSON file
- `--workflow-file`: Authored native MOWL JSON file (exclusive with --dsl-yaml)
- `--workflow-id`: Existing workflow definition ID to deploy a new version
- `--workspace-id`: Workspace ID (default: from config)

## workflow deploy-candidate

- `--acceptance-file`: path to complete JSON emitted by workflow accept-candidate
- `--compute-resource-id`: compute resource ID
- `--cron-expression`: Cron expression; see workflow update --help for syntax, timezone and examples
- `--default-values-file`: Workflow default values JSON file
- `--description`: workflow description
- `--dynamic-service-file`: Dynamic service configuration JSON file
- `--execution-mode`: execution mode
- `--name`: workflow name
- `--volume-trigger-file`: Volume trigger configuration JSON file
- `--workflow-id`: existing workflow ID
- `--workspace-id`: Workspace ID (default: from config)

## workflow deploy-mowl

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--name`: name
- `--workspace-id`: Workspace ID (default: from config)

## workflow deprecate-version

- `--version-id`: version-id
- `--workflow-id`: workflow-id
- `--workspace-id`: Workspace ID

## workflow disable-custom-operator

- `--operator-id`: operatorID
- `--workspace-id`: Workspace ID (default: from config)

## workflow disable-workitem-api-service

- `--node-id`: nodeID
- `--version`: version
- `--workspace-id`: Workspace ID (default: from config)

## workflow enable

- `--workflow-id`: workflowID
- `--workspace-id`: Workspace ID (default: from config)

## workflow enable-custom-operator

- `--operator-id`: operatorID
- `--workspace-id`: Workspace ID (default: from config)

## workflow enable-workitem-api-service

- `--node-id`: nodeID
- `--version`: version
- `--workspace-id`: Workspace ID (default: from config)

## workflow ensure-test-scratch

- `--json`: json
- `--json-file`: json-file
- `--workspace-id`: workspace-id

## workflow get

- `--workflow-id`: workflowID
- `--workspace-id`: Workspace ID (default: from config)

## workflow get-custom-operator

- `--operator-id`: operatorID
- `--workspace-id`: Workspace ID (default: from config)

## workflow get-custom-operator-code

- `--operator-id`: operatorID
- `--workspace-id`: Workspace ID (default: from config)

## workflow get-custom-operator-source

- `--node-id`: WorkItem node ID
- `--task-id`: Optional authoring task ID
- `--version`: WorkItem version
- `--workspace-id`: Workspace ID (default: from config)

## workflow get-dispatch-job

- `--dispatch-job-id`: dispatchJobID
- `--workflow-id`: workflowID
- `--workspace-id`: Workspace ID (default: from config)

## workflow get-run

- `--execution-id`: executionID
- `--workflow-id`: workflowID
- `--workspace-id`: Workspace ID (default: from config)

## workflow get-run-by-id

- `--execution-id`: executionID
- `--workspace-id`: Workspace ID (default: from config)

## workflow get-run-import-context

- `--execution-id`: executionID
- `--workflow-id`: workflowID
- `--workspace-id`: Workspace ID (default: from config)

## workflow get-run-import-context-by-id

- `--execution-id`: executionID
- `--workspace-id`: Workspace ID (default: from config)

## workflow get-run-node

- `--execution-id`: executionID
- `--node-key`: nodeKey
- `--workflow-id`: workflowID
- `--workspace-id`: Workspace ID (default: from config)

## workflow get-run-node-by-id

- `--execution-id`: executionID
- `--node-key`: nodeKey
- `--workspace-id`: Workspace ID (default: from config)

## workflow get-run-node-rerun-plan

- `--execution-id`: executionID
- `--node-key`: nodeKey
- `--query-param`: Product query parameter name=value (repeatable; pagination/filter names follow the product API)
- `--workflow-id`: workflowID
- `--workspace-id`: Workspace ID (default: from config)

## workflow get-run-rerun

- `--execution-id`: executionID
- `--rerun-id`: rerunID
- `--workflow-id`: workflowID
- `--workspace-id`: Workspace ID (default: from config)

## workflow get-run-result

- `--execution-id`: Execution ID returned by workflow run
- `--workflow-id`: Workflow application ID (use get-run-result-by-id when only the execution ID is known)
- `--workspace-id`: Workspace ID (default: from config)

## workflow get-run-result-by-id

- `--execution-id`: Execution ID returned by workflow run
- `--workspace-id`: Workspace ID (default: from config)

## workflow get-workflow-template

- `--template-id`: templateID
- `--workspace-id`: Workspace ID (default: from config)

## workflow get-workitem-api-service

- `--node-id`: nodeID
- `--version`: version
- `--workspace-id`: Workspace ID (default: from config)

## workflow history

- `--context-id`: context-id
- `--limit`: limit
- `--offset`: offset
- `--workspace-id`: workspace-id

## workflow invoke-dynamic-service

- `--file`: Explicit multipart upload file path
- `--file-field`: Multipart upload field
- `--json`: JSON object containing service_name and the business payload
- `--json-file`: JSON file containing service_name and the business payload
- `--workspace-id`: Workspace ID (default: from config)

## workflow invoke-workflow-dynamic-service

- `--file`: Explicit multipart upload file path
- `--file-field`: Multipart upload field
- `--json`: JSON object containing service_name and the business payload
- `--json-file`: JSON file containing service_name and the business payload
- `--workspace-id`: Workspace ID (default: from config)

## workflow invoke-workitem-api-service

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--node-id`: nodeID
- `--workspace-id`: Workspace ID (default: from config)

## workflow latest-published-version

- `--workflow-id`: workflow-id
- `--workspace-id`: Workspace ID

## workflow list

- `--execution-mode`: Filter by execution mode
- `--include-dynamic-service`: Include dynamic service workflows
- `--limit`: Maximum number of workflows to return
- `--name-search`: Filter by workflow name
- `--offset`: Number of workflows to skip
- `--source-type`: Filter by workflow source type
- `--status`: Filter by workflow status
- `--workflow-definition-id`: Filter by workflow definition ID
- `--workspace-id`: Workspace ID (default: from config)

## workflow list-all-runs

- `--workspace-id`: Workspace ID (default: from config)

## workflow list-file-executions

- `--file-id`: Catalog file ID whose workflow executions should be returned
- `--workspace-id`: Workspace ID (default: from config)

## workflow list-run-import-files

- `--execution-id`: executionID
- `--query-param`: Product query parameter name=value (repeatable; pagination/filter names follow the product API)
- `--workflow-id`: workflowID
- `--workspace-id`: Workspace ID (default: from config)

## workflow list-run-import-files-by-id

- `--execution-id`: executionID
- `--query-param`: Product query parameter name=value (repeatable; pagination/filter names follow the product API)
- `--workspace-id`: Workspace ID (default: from config)

## workflow list-run-rerun-branches

- `--execution-id`: executionID
- `--workflow-id`: workflowID
- `--workspace-id`: Workspace ID (default: from config)

## workflow list-runs

- `--workflow-id`: workflowID
- `--workspace-id`: Workspace ID (default: from config)

## workflow observe-test-scratch

- `--json`: json
- `--json-file`: json-file
- `--workspace-id`: workspace-id

## workflow pause

- `--workflow-id`: workflowID
- `--workspace-id`: Workspace ID (default: from config)

## workflow pause-run

- `--execution-id`: executionID
- `--workflow-id`: workflowID
- `--workspace-id`: Workspace ID (default: from config)

## workflow publish-version

- `--version-id`: version-id
- `--workflow-id`: workflow-id
- `--workspace-id`: Workspace ID

## workflow publish-workitem-api-service

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--node-id`: nodeID
- `--workspace-id`: Workspace ID (default: from config)

## workflow refresh-run

- `--execution-id`: executionID
- `--workflow-id`: workflowID
- `--workspace-id`: Workspace ID (default: from config)

## workflow reject-candidate-proposal

- `--proposal-id`: proposal-id
- `--workspace-id`: workspace-id

## workflow reset-test-scratch

- `--task-id`: task-id
- `--workspace-id`: workspace-id

## workflow restore-history

- `--artifact-id`: artifact-id
- `--context-id`: context-id
- `--parent-artifact-id`: parent history artifact ID
- `--parent-task-id`: parent history task ID
- `--request-id`: request-id
- `--task-id`: task-id
- `--workspace-id`: workspace-id

## workflow resume

- `--workflow-id`: workflowID
- `--workspace-id`: Workspace ID (default: from config)

## workflow resume-run

- `--execution-id`: executionID
- `--workflow-id`: workflowID
- `--workspace-id`: Workspace ID (default: from config)

## workflow retry-run

- `--execution-id`: executionID
- `--replace-saved-defaults`: Replace saved workflow defaults with the supplied values
- `--trigger-now`: Request immediate workflow dispatch; may be set to false explicitly
- `--values-file`: JSON object containing workflow runtime values
- `--workflow-id`: workflowID
- `--workspace-id`: Workspace ID (default: from config)

## workflow run

- `--replace-saved-defaults`: Replace saved workflow defaults with the supplied values
- `--trigger-now`: Request immediate workflow dispatch; may be set to false explicitly
- `--values-file`: JSON object containing workflow runtime values
- `--workflow-id`: workflowID
- `--workspace-id`: Workspace ID (default: from config)

## workflow submit-candidate

- `--json`: json
- `--json-file`: json-file
- `--project-dir`: Workflow authoring directory in this filesystem (requires a live runtime task)
- `--workspace-id`: workspace-id

## workflow switch-artifact-effective-revisions

- `--artifact-id`: artifactID
- `--expected-effective-set-version`: Expected effective set version
- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--workspace-id`: Workspace ID (default: from config)

## workflow switch-artifact-node-effective-revisions

- `--artifact-id`: artifactID
- `--expected-effective-set-version`: Expected effective set version
- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--node-id`: nodeID
- `--workspace-id`: Workspace ID (default: from config)

## workflow test-candidate

- `--case`: One JSON filename under project tests/cases (diagnostic only)
- `--json`: json
- `--json-file`: json-file
- `--project-dir`: Workflow authoring directory in this filesystem (requires a live runtime task)
- `--verification-only`: Re-run project verification against the preceding suite execution
- `--workspace-id`: workspace-id

## workflow test-custom-operator

- `--json`: JSON object containing the required input payload
- `--json-file`: JSON file containing the required input payload
- `--operator-id`: operatorID
- `--workspace-id`: Workspace ID (default: from config)

## workflow update

- `--cron-expression`: Cron expression; see workflow update --help for syntax, timezone and examples
- `--default-values-file`: Workflow default values JSON file
- `--description`: Workflow description; pass an empty value to clear it
- `--dsl-yaml`: Replacement DSL YAML; pass an empty value to clear it
- `--dynamic-service-file`: Dynamic service configuration JSON file
- `--execution-mode`: Workflow execution mode: one_shot, cron, volume_trigger or dynamic_service
- `--name`: Workflow name
- `--runtime-fields-file`: Runtime parameter fields JSON file; see workflow deploy --help for an example
- `--runtime-layout-file`: Runtime form layout JSON file
- `--volume-trigger-file`: Volume trigger configuration JSON file
- `--workflow-id`: workflowID
- `--workspace-id`: Workspace ID (default: from config)

## workflow update-custom-operator

- `--operator-id`: operatorID
- `--workspace-id`: Workspace ID (default: from config)

## workflow update-mowl-workflow-template

- `--description`: Template description
- `--input-form`: JSON-encoded runtime fields string; empty clears the form
- `--is-builtin`: Builtin template flag; the server validates creation and type changes
- `--json`: JSON MOWL publication graph
- `--json-file`: JSON file containing the MOWL graph
- `--name`: Template name
- `--template-id`: Template ID
- `--template-key`: Stable builtin template key (ignored for custom templates)
- `--workspace-id`: Workspace ID (default: from config)

## workflow update-workflow-template

- `--description`: Template description
- `--dsl-yaml`: dslYAML
- `--input-form`: JSON-encoded runtime fields string; empty clears the form
- `--is-builtin`: Builtin template flag; the server validates creation and type changes
- `--name`: name
- `--template-id`: templateID
- `--template-key`: Stable builtin template key (ignored for custom templates)
- `--workspace-id`: Workspace ID (default: from config)

## workflow version

- `--version-id`: version-id
- `--workflow-id`: workflow-id
- `--workspace-id`: Workspace ID

## workflow version-by-number

- `--number`: number
- `--workflow-id`: workflow-id
- `--workspace-id`: Workspace ID

## workflow versions

- `--workflow-id`: workflow-id
- `--workspace-id`: Workspace ID

## workflow workflow-templates

- `--workspace-id`: Workspace ID (default: from config)

## workflow workitem-catalog

- `--workspace-id`: Workspace ID (default: from config)

## workflow workitem-catalog-item

- `--node-id`: nodeID
- `--workspace-id`: Workspace ID (default: from config)

