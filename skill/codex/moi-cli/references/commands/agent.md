# agent command reference

Generated from CLI source. Do not edit. Business requirements remain server-owned.
Use `moi-cli schema DOMAIN COMMAND` for typed JSON input and `--help` for defaults.

## agent a2a

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--workspace-id`: Workspace ID (default: from config)

## agent a2a-stream

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--output`: Write stream to file (default stdout)
- `--workspace-id`: Workspace ID (default: from config)

## agent admin-feedback-ticket

- `--ticket-id`: Ticket ID
- `--workspace-id`: Workspace ID (default: from config)

## agent admin-feedback-tickets

- `--workspace-id`: Workspace ID

## agent agent-card

- `--agent-code`: Agent code
- `--agent-id`: Agent ID
- `--agent-workspace-id`: Workspace that owns the agent (query agent_workspace_id)
- `--workspace-id`: Workspace ID (default: from config)

## agent approve-feedback-reward

- `--json`: Approval update JSON
- `--json-file`: Approval update JSON file
- `--ticket-id`: Ticket ID
- `--workspace-id`: Workspace ID (default: from config)

## agent begin-plugin-oauth

- `--expected-version`: expected-version
- `--id`: id
- `--workspace-id`: Workspace ID (default: from config)

## agent build

- `--output`: Output .moiagent file path
- `--overwrite`: Overwrite output file if it already exists
- `--running-moi-version`: Override MOI version for minMOIVersion validation
- `--source`: Package source directory

## agent build-feedback-github-issue-draft

- `--kind`: Issue kind: bug, feature, or auto
- `--ticket-id`: Ticket ID
- `--workspace-id`: Workspace ID (default: from config)

## agent cancel-task

- `--agent-code`: Agent code
- `--agent-id`: Agent ID
- `--request-id`: Request id
- `--task-id`: Task id
- `--workspace-id`: Workspace ID

## agent create-callback-rule

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--workspace-id`: Workspace ID (default: from config)

## agent create-channel-instance

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--provider`: provider
- `--workspace-id`: Workspace ID (default: from config)

## agent create-feedback-github-issue

- `--body`: Issue body
- `--kind`: Issue kind: bug, feature, or auto
- `--retry`: Retry a failed issue creation
- `--ticket-id`: Ticket ID
- `--title`: Issue title
- `--workspace-id`: Workspace ID (default: from config)

## agent create-feedback-ticket

- `--json`: json
- `--json-file`: json-file
- `--workspace-id`: workspace-id

## agent create-plugin-connection

- `--json`: Request JSON (secrets: prefer --json-file)
- `--json-file`: Request JSON file
- `--workspace-id`: Workspace ID (default: from config)

## agent delete-callback-rule

- `--rule-id`: ruleID
- `--workspace-id`: Workspace ID (default: from config)

## agent delete-plugin-connection

- `--expected-version`: expected-version
- `--id`: id
- `--workspace-id`: Workspace ID (default: from config)

## agent delete-plugin-installation

- `--expected-version`: expected-version
- `--id`: id
- `--workspace-id`: Workspace ID (default: from config)

## agent describe-plugin-capability

- `--ref`: ref
- `--workspace-id`: Workspace ID (default: from config)

## agent disable-channel-instance

- `--instance-id`: instanceID
- `--provider`: provider
- `--workspace-id`: Workspace ID (default: from config)

## agent disconnect-plugin-connection

- `--expected-generation`: expected-generation
- `--id`: id
- `--workspace-id`: Workspace ID (default: from config)

## agent download-skill-files

- `--output-dir`: New directory for task-frozen files; must not exist
- `--skill-id`: Task-bound Skill ID

## agent executor-health

- `--workspace-id`: Workspace ID (default: from config)

## agent export-agent-package

- `--agent-id`: agent-id
- `--output`: output
- `--version`: version
- `--workspace-id`: workspace-id

## agent get-callback-rule

- `--rule-id`: ruleID
- `--workspace-id`: Workspace ID (default: from config)

## agent get-channel-instance

- `--instance-id`: instanceID
- `--provider`: provider
- `--workspace-id`: Workspace ID (default: from config)

## agent get-plugin-authorization

- `--id`: id
- `--workspace-id`: Workspace ID (default: from config)

## agent get-plugin-connection

- `--id`: id
- `--workspace-id`: Workspace ID (default: from config)

## agent get-plugin-installation

- `--id`: id
- `--workspace-id`: Workspace ID (default: from config)

## agent get-plugin-release

- `--namespace`: namespace query option
- `--release-id`: release-id
- `--workspace-id`: Workspace ID (default: from config)

## agent get-skill-file

- `--file-path`: filePath
- `--output`: Write stream to file (default stdout)
- `--skill-id`: skillID
- `--workspace-id`: Workspace ID (default: from config)

## agent get-task

- `--agent-code`: Agent code
- `--agent-id`: Agent ID
- `--request-id`: Request id
- `--task-id`: Task id
- `--workspace-id`: Workspace ID

## agent import-skill-package

- `--file-path`: Local file path (repeat only for connector uploads)
- `--filename`: filename
- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--max-upload-bytes`: Total byte limit for file-path input
- `--workspace-id`: Workspace ID (default: from config)

## agent init

- `--agent-id`: Agent ID
- `--description`: Agent display description
- `--instruction`: Initial role instruction
- `--model`: Default model name
- `--name`: Agent display name
- `--output`: Output source directory
- `--version`: Agent package semantic version

## agent inspect-skill-package

- `--file-path`: Local file path (repeat only for connector uploads)
- `--filename`: filename
- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--max-upload-bytes`: Total byte limit for file-path input
- `--workspace-id`: Workspace ID (default: from config)

## agent install-plugin

- `--json`: Request JSON (secrets: prefer --json-file)
- `--json-file`: Request JSON file
- `--workspace-id`: Workspace ID (default: from config)

## agent list

- `--query`: Search agents
- `--workspace-id`: Workspace ID (default: from config)

## agent list-callback-rules

- `--workspace-id`: Workspace ID (default: from config)

## agent list-channel-instances

- `--provider`: provider
- `--workspace-id`: Workspace ID (default: from config)

## agent list-plugin-connections

- `--limit`: limit query option
- `--offset`: offset query option
- `--workspace-id`: Workspace ID (default: from config)

## agent list-plugin-installations

- `--limit`: limit query option
- `--offset`: offset query option
- `--workspace-id`: Workspace ID (default: from config)

## agent list-plugin-releases

- `--limit`: limit query option
- `--namespace`: namespace query option
- `--offset`: offset query option
- `--q`: q query option
- `--workspace-id`: Workspace ID (default: from config)

## agent load-agent-package

- `--file-path`: Local file path (repeat only for connector uploads)
- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--max-upload-bytes`: Total byte limit for file-path input
- `--workspace-id`: Workspace ID (default: from config)

## agent my-feedback-tickets

- `--workspace-id`: workspace-id

## agent polish-skill-stream

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--output`: Write stream to file (default stdout)
- `--workspace-id`: Workspace ID (default: from config)

## agent poll-plugin-oauth

- `--attempt-id`: attempt-id
- `--expected-version`: expected-version
- `--id`: id
- `--workspace-id`: Workspace ID (default: from config)

## agent preview-query-visual

- `--agent-id`: Agent ID
- `--file-id`: Query visual file ID
- `--output`: Write image to file (default stdout)
- `--workspace-id`: Workspace ID (default: from config)

## agent read-plugin-skill

- `--path`: path
- `--ref`: ref
- `--workspace-id`: Workspace ID (default: from config)

## agent resource-any

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--method`: method
- `--resource`: resource
- `--tail-path`: tailPath
- `--workspace-id`: Workspace ID (default: from config)

## agent resource-delete

- `--resource`: resource
- `--tail-path`: tailPath
- `--workspace-id`: Workspace ID (default: from config)

## agent resource-get

- `--resource`: resource
- `--tail-path`: tailPath
- `--workspace-id`: Workspace ID (default: from config)

## agent resource-patch

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--resource`: resource
- `--tail-path`: tailPath
- `--workspace-id`: Workspace ID (default: from config)

## agent resource-post

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--resource`: resource
- `--tail-path`: tailPath
- `--workspace-id`: Workspace ID (default: from config)

## agent resource-put

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--resource`: resource
- `--tail-path`: tailPath
- `--workspace-id`: Workspace ID (default: from config)

## agent search-plugins

- `--availability`: availability query option
- `--cursor`: cursor query option
- `--limit`: limit query option
- `--q`: q query option
- `--workspace-id`: Workspace ID (default: from config)

## agent send

- `--agent-code`: Agent code
- `--agent-id`: Agent ID
- `--message-id`: Message id
- `--request-id`: Request id
- `--text`: User text
- `--workspace-id`: Workspace ID

## agent store-plugin-secret

- `--id`: id
- `--json`: Request JSON (secrets: prefer --json-file)
- `--json-file`: Request JSON file
- `--workspace-id`: Workspace ID (default: from config)

## agent test-channel-configuration

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--provider`: provider
- `--workspace-id`: Workspace ID (default: from config)

## agent test-channel-instance

- `--instance-id`: instanceID
- `--provider`: provider
- `--workspace-id`: Workspace ID (default: from config)

## agent unpack

- `--file`: .moiagent package file
- `--output`: Output source directory
- `--running-moi-version`: Override MOI version for minMOIVersion validation

## agent update-admin-feedback-ticket

- `--json`: Request JSON
- `--json-file`: Request JSON file
- `--ticket-id`: ticket-id
- `--workspace-id`: Workspace ID

## agent update-callback-rule

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--rule-id`: ruleID
- `--workspace-id`: Workspace ID (default: from config)

## agent update-channel-instance

- `--instance-id`: instanceID
- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--provider`: provider
- `--workspace-id`: Workspace ID (default: from config)

## agent update-plugin-connection

- `--id`: id
- `--json`: Request JSON (secrets: prefer --json-file)
- `--json-file`: Request JSON file
- `--workspace-id`: Workspace ID (default: from config)

## agent update-plugin-installation

- `--id`: id
- `--json`: Request JSON (secrets: prefer --json-file)
- `--json-file`: Request JSON file
- `--workspace-id`: Workspace ID (default: from config)

## agent upload-workspace-file

- `--file-path`: Local file path (repeat only for connector uploads)
- `--filename`: filename
- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--max-upload-bytes`: Total byte limit for file-path input
- `--workspace-id`: Workspace ID (default: from config)

## agent verify

- `--file`: .moiagent package file
- `--running-moi-version`: Override MOI version for minMOIVersion validation
- `--source`: Package source directory containing agent.yaml

## agent workspace-mcp

- `--agent-id`: agent-id
- `--json`: Request JSON
- `--json-file`: Request JSON file
- `--runtime-token-file`: Runtime grant file
- `--workspace-id`: Workspace ID
