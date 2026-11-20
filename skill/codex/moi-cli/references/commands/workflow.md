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
- `--workflow-id`: workflowID
- `--workspace-id`: Workspace ID (default: from config)

## workflow cancel-run-rerun

- `--execution-id`: executionID
- `--rerun-id`: rerunID
- `--workflow-id`: workflowID
- `--workspace-id`: Workspace ID (default: from config)

## workflow catalog-file-lineage-overview

- `--file-id`: fileID
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

## workflow create-run-node-rerun

- `--execution-id`: executionID
- `--expected-rerun-contract-hash`: Required rerun contract hash returned by get-run-node-rerun-plan
- `--mode`: mode
- `--node-key`: nodeKey
- `--source-node-execution-id`: sourceNodeExecutionID
- `--workflow-id`: workflowID
- `--workspace-id`: Workspace ID (default: from config)

## workflow create-target-artifact-node-revision-rerun

- `--artifact-id`: artifactID
- `--node-id`: nodeID
- `--target-output-emission-id`: targetOutputEmissionID
- `--workspace-id`: Workspace ID (default: from config)

## workflow create-version

- `--json`: Request JSON
- `--json-file`: Request JSON file
- `--workflow-id`: workflow-id
- `--workspace-id`: Workspace ID

## workflow create-workflow-template

- `--dsl-yaml`: dslYAML
- `--name`: name
- `--workspace-id`: Workspace ID (default: from config)

## workflow custom-operators

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

- `--dsl-yaml`: dslYAML
- `--name`: name
- `--workspace-id`: Workspace ID (default: from config)

## workflow deploy-candidate

- `--acceptance-file`: path to complete JSON emitted by workflow accept-candidate
- `--compute-resource-id`: compute resource ID
- `--cron-expression`: cron expression
- `--description`: workflow description
- `--execution-mode`: execution mode
- `--name`: workflow name
- `--workflow-id`: existing workflow ID
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
- `--workflow-id`: workflowID
- `--workspace-id`: Workspace ID (default: from config)

## workflow get-run-rerun

- `--execution-id`: executionID
- `--rerun-id`: rerunID
- `--workflow-id`: workflowID
- `--workspace-id`: Workspace ID (default: from config)

## workflow get-run-result

- `--execution-id`: executionID
- `--workflow-id`: workflowID
- `--workspace-id`: Workspace ID (default: from config)

## workflow get-run-result-by-id

- `--execution-id`: executionID
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
- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--workspace-id`: Workspace ID (default: from config)

## workflow invoke-workflow-dynamic-service

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
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

- `--workspace-id`: Workspace ID (default: from config)

## workflow list-all-runs

- `--workspace-id`: Workspace ID (default: from config)

## workflow list-file-executions

- `--workspace-id`: Workspace ID (default: from config)

## workflow list-run-import-files

- `--execution-id`: executionID
- `--workflow-id`: workflowID
- `--workspace-id`: Workspace ID (default: from config)

## workflow list-run-import-files-by-id

- `--execution-id`: executionID
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
- `--workflow-id`: workflowID
- `--workspace-id`: Workspace ID (default: from config)

## workflow run

- `--workflow-id`: workflowID
- `--workspace-id`: Workspace ID (default: from config)

## workflow submit-candidate

- `--json`: json
- `--json-file`: json-file
- `--workspace-id`: workspace-id

## workflow switch-artifact-effective-revisions

- `--artifact-id`: artifactID
- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--workspace-id`: Workspace ID (default: from config)

## workflow switch-artifact-node-effective-revisions

- `--artifact-id`: artifactID
- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--node-id`: nodeID
- `--workspace-id`: Workspace ID (default: from config)

## workflow test-candidate

- `--json`: json
- `--json-file`: json-file
- `--workspace-id`: workspace-id

## workflow test-custom-operator

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--operator-id`: operatorID
- `--workspace-id`: Workspace ID (default: from config)

## workflow update

- `--workflow-id`: workflowID
- `--workspace-id`: Workspace ID (default: from config)

## workflow update-custom-operator

- `--operator-id`: operatorID
- `--workspace-id`: Workspace ID (default: from config)

## workflow update-workflow-template

- `--dsl-yaml`: dslYAML
- `--name`: name
- `--template-id`: templateID
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
