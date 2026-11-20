# agent command reference

Generated from CLI source. Do not edit. Business requirements remain server-owned.
Use `moi-cli schema DOMAIN COMMAND` for typed JSON input and `--help` for defaults.

## agent a2a

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--no-wait`: Return a non-terminal message/send task admission without waiting for its terminal state
- `--workspace-id`: Workspace ID (default: from config)

## agent a2a-stream

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--output`: Write stream to file (default stdout)
- `--workspace-id`: Workspace ID (default: from config)

## agent add-team-member

- `--workspace-id`: Workspace ID (default: from config)
- `--team-id`: Team ID
- `--context-id`: Team Session context ID
- `--agent-workspace-id`: Member workspace ID
- `--agent-id`: Member Agent ID
- `--connection-id`: Ingress connection ID
- `--qr-session-token`: Existing WeCom QR session token (empty starts a session)
- `--expected-config-version`: Expected ingress configuration version
- `--target-kind`: Channel target kind: agent or team
- `--target-workspace-id`: Channel target workspace ID
- `--target-id`: Channel target ID
- `--expected-version`: Expected Team, membership or association version for this operation
- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--catalog`: Agent catalog filter
- `--query`: Agent catalog search query
- `--include-development`: Include development Agents
- `--limit`: Page size
- `--offset`: Page offset

## agent admin-feedback-ticket

- `--ticket-id`: Ticket ID
- `--workspace-id`: Workspace ID (default: from config)

## agent admin-feedback-tickets

- `--query-param`: Product query parameter name=value (repeatable; pagination/filter names follow the product API)
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

## agent artifact

- `--artifact-id`: 目标 ID
- `--version`: 成果版本
- `--workspace-id`: 工作区 ID，省略时读取配置

## agent artifacts

- `--query-param`: 筛选或分页参数 name=value，可重复
- `--workspace-id`: 工作区 ID，省略时读取配置

## agent backfill-artifacts

- `--json`: 请求 JSON
- `--json-file`: 请求 JSON 文件
- `--workspace-id`: 工作区 ID，省略时读取配置

## agent begin-plugin-oauth

- `--expected-version`: expected-version
- `--id`: id
- `--workspace-id`: Workspace ID (default: from config)

## agent bind-channel

- `--instance-id`: Channel instance ID returned by list-channel-instances (required)
- `--provider`: Provider of the existing channel instance (required)

## agent build

- `--output`: Output .moiagent file path
- `--overwrite`: Overwrite output file if it already exists
- `--running-moi-version`: Override MOI version for minMOIVersion validation
- `--source`: Package source directory

## agent build-feedback-github-issue-draft

- `--kind`: Issue kind: bug, feature, or auto
- `--ticket-id`: Ticket ID
- `--workspace-id`: Workspace ID (default: from config)

## agent builder context

- `--output-file`: Write trusted editing context to this JSON file

## agent builder propose

- `--json-file`: JSON object containing config or changes
- `--output-file`: Optionally save the complete server result to a JSON file

## agent builder workspace-mcp

- `--json-file`: Existing published capability preparation request JSON
- `--output-file`: Optionally save the complete server result to a JSON file

## agent cancel-task

- `--agent-code`: Agent code
- `--agent-id`: Agent ID
- `--request-id`: Request id
- `--task-id`: Task id
- `--workspace-id`: Workspace ID

## agent confirm-skill-candidate

- `--json`: JSON for the reviewed skill candidate
- `--json-file`: JSON file for the reviewed skill candidate
- `--task-id`: Runtime task ID
- `--workspace-id`: Workspace ID (default: from config)

## agent connect-browser

- `--workspace-id`: Workspace ID (default: from config)
- `--task-id`: Runtime task ID
- `--compression`: Negotiate browser stream compression

## agent connect-runtime-browser

- `--origin`: Allowed HTTP(S) origin for the WebSocket handshake
- `--task-id`: Agent runtime task ID
- `--token`: Short-lived token returned by agent runtime-browser-token
- `--workspace-id`: Workspace ID (default: from config)

## agent create-callback-rule

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--workspace-id`: Workspace ID (default: from config)

## agent create-channel-instance

- `--json`: Native channel JSON with config and secrets; prefer --json-file for credentials
- `--json-file`: Native channel JSON file; GitHub secrets.access_token accepts OAuth/PAT (no vault reference)
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

## agent create-ingress-connection

- `--workspace-id`: Workspace ID (default: from config)
- `--team-id`: Team ID
- `--context-id`: Team Session context ID
- `--agent-workspace-id`: Member workspace ID
- `--agent-id`: Member Agent ID
- `--connection-id`: Ingress connection ID
- `--qr-session-token`: Existing WeCom QR session token (empty starts a session)
- `--expected-config-version`: Expected ingress configuration version
- `--target-kind`: Channel target kind: agent or team
- `--target-workspace-id`: Channel target workspace ID
- `--target-id`: Channel target ID
- `--expected-version`: Expected Team, membership or association version for this operation
- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--catalog`: Agent catalog filter
- `--query`: Agent catalog search query
- `--include-development`: Include development Agents
- `--limit`: Page size
- `--offset`: Page offset

## agent create-plugin-connection

- `--json`: Request JSON (secrets: prefer --json-file)
- `--json-file`: Request JSON file
- `--workspace-id`: Workspace ID (default: from config)

## agent create-team

- `--workspace-id`: Workspace ID (default: from config)
- `--team-id`: Team ID
- `--context-id`: Team Session context ID
- `--agent-workspace-id`: Member workspace ID
- `--agent-id`: Member Agent ID
- `--connection-id`: Ingress connection ID
- `--qr-session-token`: Existing WeCom QR session token (empty starts a session)
- `--expected-config-version`: Expected ingress configuration version
- `--target-kind`: Channel target kind: agent or team
- `--target-workspace-id`: Channel target workspace ID
- `--target-id`: Channel target ID
- `--expected-version`: Expected Team, membership or association version for this operation
- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--catalog`: Agent catalog filter
- `--query`: Agent catalog search query
- `--include-development`: Include development Agents
- `--limit`: Page size
- `--offset`: Page offset

## agent create-team-session

- `--workspace-id`: Workspace ID (default: from config)
- `--team-id`: Team ID
- `--context-id`: Team Session context ID
- `--agent-workspace-id`: Member workspace ID
- `--agent-id`: Member Agent ID
- `--connection-id`: Ingress connection ID
- `--qr-session-token`: Existing WeCom QR session token (empty starts a session)
- `--expected-config-version`: Expected ingress configuration version
- `--target-kind`: Channel target kind: agent or team
- `--target-workspace-id`: Channel target workspace ID
- `--target-id`: Channel target ID
- `--expected-version`: Expected Team, membership or association version for this operation
- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--catalog`: Agent catalog filter
- `--query`: Agent catalog search query
- `--include-development`: Include development Agents
- `--limit`: Page size
- `--offset`: Page offset

## agent delete-artifacts

- `--json`: 请求 JSON
- `--json-file`: 请求 JSON 文件
- `--workspace-id`: 工作区 ID，省略时读取配置

## agent delete-callback-rule

- `--rule-id`: ruleID
- `--workspace-id`: Workspace ID (default: from config)

## agent delete-ingress-connection

- `--workspace-id`: Workspace ID (default: from config)
- `--team-id`: Team ID
- `--context-id`: Team Session context ID
- `--agent-workspace-id`: Member workspace ID
- `--agent-id`: Member Agent ID
- `--connection-id`: Ingress connection ID
- `--qr-session-token`: Existing WeCom QR session token (empty starts a session)
- `--expected-config-version`: Expected ingress configuration version
- `--target-kind`: Channel target kind: agent or team
- `--target-workspace-id`: Channel target workspace ID
- `--target-id`: Channel target ID
- `--expected-version`: Expected Team, membership or association version for this operation
- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--catalog`: Agent catalog filter
- `--query`: Agent catalog search query
- `--include-development`: Include development Agents
- `--limit`: Page size
- `--offset`: Page offset

## agent delete-plugin-connection

- `--expected-version`: expected-version
- `--id`: id
- `--workspace-id`: Workspace ID (default: from config)

## agent delete-plugin-installation

- `--expected-version`: expected-version
- `--id`: id
- `--workspace-id`: Workspace ID (default: from config)

## agent delete-team

- `--workspace-id`: Workspace ID (default: from config)
- `--team-id`: Team ID
- `--context-id`: Team Session context ID
- `--agent-workspace-id`: Member workspace ID
- `--agent-id`: Member Agent ID
- `--connection-id`: Ingress connection ID
- `--qr-session-token`: Existing WeCom QR session token (empty starts a session)
- `--expected-config-version`: Expected ingress configuration version
- `--target-kind`: Channel target kind: agent or team
- `--target-workspace-id`: Channel target workspace ID
- `--target-id`: Channel target ID
- `--expected-version`: Expected Team, membership or association version for this operation
- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--catalog`: Agent catalog filter
- `--query`: Agent catalog search query
- `--include-development`: Include development Agents
- `--limit`: Page size
- `--offset`: Page offset

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

## agent executor-health

- `--workspace-id`: Workspace ID (default: from config)

## agent export-agent-package

- `--agent-id`: agent-id
- `--output`: output
- `--version`: version
- `--workspace-id`: workspace-id

## agent generate-we-com-callback-secrets

- `--workspace-id`: Workspace ID (default: from config)
- `--team-id`: Team ID
- `--context-id`: Team Session context ID
- `--agent-workspace-id`: Member workspace ID
- `--agent-id`: Member Agent ID
- `--connection-id`: Ingress connection ID
- `--qr-session-token`: Existing WeCom QR session token (empty starts a session)
- `--expected-config-version`: Expected ingress configuration version
- `--target-kind`: Channel target kind: agent or team
- `--target-workspace-id`: Channel target workspace ID
- `--target-id`: Channel target ID
- `--expected-version`: Expected Team, membership or association version for this operation
- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--catalog`: Agent catalog filter
- `--query`: Agent catalog search query
- `--include-development`: Include development Agents
- `--limit`: Page size
- `--offset`: Page offset

## agent get-callback-rule

- `--rule-id`: ruleID
- `--workspace-id`: Workspace ID (default: from config)

## agent get-channel-instance

- `--instance-id`: instanceID
- `--provider`: provider
- `--workspace-id`: Workspace ID (default: from config)

## agent get-ingress-connection

- `--workspace-id`: Workspace ID (default: from config)
- `--team-id`: Team ID
- `--context-id`: Team Session context ID
- `--agent-workspace-id`: Member workspace ID
- `--agent-id`: Member Agent ID
- `--connection-id`: Ingress connection ID
- `--qr-session-token`: Existing WeCom QR session token (empty starts a session)
- `--expected-config-version`: Expected ingress configuration version
- `--target-kind`: Channel target kind: agent or team
- `--target-workspace-id`: Channel target workspace ID
- `--target-id`: Channel target ID
- `--expected-version`: Expected Team, membership or association version for this operation
- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--catalog`: Agent catalog filter
- `--query`: Agent catalog search query
- `--include-development`: Include development Agents
- `--limit`: Page size
- `--offset`: Page offset

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

## agent get-skill-candidate-confirmation

- `--call-id`: Skill Creator tool call ID
- `--source-digest`: Reviewed candidate source digest
- `--task-id`: Runtime task ID
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

## agent get-team

- `--workspace-id`: Workspace ID (default: from config)
- `--team-id`: Team ID
- `--context-id`: Team Session context ID
- `--agent-workspace-id`: Member workspace ID
- `--agent-id`: Member Agent ID
- `--connection-id`: Ingress connection ID
- `--qr-session-token`: Existing WeCom QR session token (empty starts a session)
- `--expected-config-version`: Expected ingress configuration version
- `--target-kind`: Channel target kind: agent or team
- `--target-workspace-id`: Channel target workspace ID
- `--target-id`: Channel target ID
- `--expected-version`: Expected Team, membership or association version for this operation
- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--catalog`: Agent catalog filter
- `--query`: Agent catalog search query
- `--include-development`: Include development Agents
- `--limit`: Page size
- `--offset`: Page offset

## agent import-skill-package

- `--category`: Imported Skill category
- `--description`: Imported Skill description
- `--file-path`: Local file path (repeat only for connector uploads)
- `--filename`: filename
- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--max-upload-bytes`: Total byte limit for file-path input
- `--name`: Imported Skill name
- `--tags`: Imported Skill tags as a JSON array or null
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

- `--catalog`: Agent catalog: workspace or system
- `--include-development`: Include development Agents
- `--limit`: Page size
- `--offset`: Page offset
- `--query`: Search agents
- `--status`: Agent status filter
- `--surface`: Agent list surface
- `--workspace-id`: Workspace ID (default: from config)

## agent list-agent-catalog-cards

- `--workspace-id`: Workspace ID (default: from config)
- `--team-id`: Team ID
- `--context-id`: Team Session context ID
- `--agent-workspace-id`: Member workspace ID
- `--agent-id`: Member Agent ID
- `--connection-id`: Ingress connection ID
- `--qr-session-token`: Existing WeCom QR session token (empty starts a session)
- `--expected-config-version`: Expected ingress configuration version
- `--target-kind`: Channel target kind: agent or team
- `--target-workspace-id`: Channel target workspace ID
- `--target-id`: Channel target ID
- `--expected-version`: Expected Team, membership or association version for this operation
- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--catalog`: Agent catalog filter
- `--query`: Agent catalog search query
- `--include-development`: Include development Agents
- `--limit`: Page size
- `--offset`: Page offset

## agent list-callback-rules

- `--query-param`: Product query parameter name=value (repeatable; pagination/filter names follow the product API)
- `--workspace-id`: Workspace ID (default: from config)

## agent list-channel-instances

- `--provider`: Channel provider ID (required; see agent list-connector-catalog)
- `--query-param`: Product query parameter name=value (repeatable; pagination/filter names follow the product API)
- `--workspace-id`: Workspace ID (default: from config)

## agent list-connector-catalog

- `--workspace-id`: Workspace ID (default: from config)

## agent list-ingress-capabilities

- `--workspace-id`: Workspace ID (default: from config)
- `--team-id`: Team ID
- `--context-id`: Team Session context ID
- `--agent-workspace-id`: Member workspace ID
- `--agent-id`: Member Agent ID
- `--connection-id`: Ingress connection ID
- `--qr-session-token`: Existing WeCom QR session token (empty starts a session)
- `--expected-config-version`: Expected ingress configuration version
- `--target-kind`: Channel target kind: agent or team
- `--target-workspace-id`: Channel target workspace ID
- `--target-id`: Channel target ID
- `--expected-version`: Expected Team, membership or association version for this operation
- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--catalog`: Agent catalog filter
- `--query`: Agent catalog search query
- `--include-development`: Include development Agents
- `--limit`: Page size
- `--offset`: Page offset

## agent list-ingress-connections

- `--workspace-id`: Workspace ID (default: from config)
- `--team-id`: Team ID
- `--context-id`: Team Session context ID
- `--agent-workspace-id`: Member workspace ID
- `--agent-id`: Member Agent ID
- `--connection-id`: Ingress connection ID
- `--qr-session-token`: Existing WeCom QR session token (empty starts a session)
- `--expected-config-version`: Expected ingress configuration version
- `--target-kind`: Channel target kind: agent or team
- `--target-workspace-id`: Channel target workspace ID
- `--target-id`: Channel target ID
- `--expected-version`: Expected Team, membership or association version for this operation
- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--catalog`: Agent catalog filter
- `--query`: Agent catalog search query
- `--include-development`: Include development Agents
- `--limit`: Page size
- `--offset`: Page offset

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

## agent list-resource-meta

- `--output-file`: Write all selectable resource metadata to this JSON file

## agent list-team-members

- `--workspace-id`: Workspace ID (default: from config)
- `--team-id`: Team ID
- `--context-id`: Team Session context ID
- `--agent-workspace-id`: Member workspace ID
- `--agent-id`: Member Agent ID
- `--connection-id`: Ingress connection ID
- `--qr-session-token`: Existing WeCom QR session token (empty starts a session)
- `--expected-config-version`: Expected ingress configuration version
- `--target-kind`: Channel target kind: agent or team
- `--target-workspace-id`: Channel target workspace ID
- `--target-id`: Channel target ID
- `--expected-version`: Expected Team, membership or association version for this operation
- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--catalog`: Agent catalog filter
- `--query`: Agent catalog search query
- `--include-development`: Include development Agents
- `--limit`: Page size
- `--offset`: Page offset

## agent list-team-messages

- `--workspace-id`: Workspace ID (default: from config)
- `--team-id`: Team ID
- `--context-id`: Team Session context ID
- `--agent-workspace-id`: Member workspace ID
- `--agent-id`: Member Agent ID
- `--connection-id`: Ingress connection ID
- `--qr-session-token`: Existing WeCom QR session token (empty starts a session)
- `--expected-config-version`: Expected ingress configuration version
- `--target-kind`: Channel target kind: agent or team
- `--target-workspace-id`: Channel target workspace ID
- `--target-id`: Channel target ID
- `--expected-version`: Expected Team, membership or association version for this operation
- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--catalog`: Agent catalog filter
- `--query`: Agent catalog search query
- `--include-development`: Include development Agents
- `--limit`: Page size
- `--offset`: Page offset

## agent list-team-sessions

- `--workspace-id`: Workspace ID (default: from config)
- `--team-id`: Team ID
- `--context-id`: Team Session context ID
- `--agent-workspace-id`: Member workspace ID
- `--agent-id`: Member Agent ID
- `--connection-id`: Ingress connection ID
- `--qr-session-token`: Existing WeCom QR session token (empty starts a session)
- `--expected-config-version`: Expected ingress configuration version
- `--target-kind`: Channel target kind: agent or team
- `--target-workspace-id`: Channel target workspace ID
- `--target-id`: Channel target ID
- `--expected-version`: Expected Team, membership or association version for this operation
- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--catalog`: Agent catalog filter
- `--query`: Agent catalog search query
- `--include-development`: Include development Agents
- `--limit`: Page size
- `--offset`: Page offset

## agent list-teams

- `--workspace-id`: Workspace ID (default: from config)
- `--team-id`: Team ID
- `--context-id`: Team Session context ID
- `--agent-workspace-id`: Member workspace ID
- `--agent-id`: Member Agent ID
- `--connection-id`: Ingress connection ID
- `--qr-session-token`: Existing WeCom QR session token (empty starts a session)
- `--expected-config-version`: Expected ingress configuration version
- `--target-kind`: Channel target kind: agent or team
- `--target-workspace-id`: Channel target workspace ID
- `--target-id`: Channel target ID
- `--expected-version`: Expected Team, membership or association version for this operation
- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--catalog`: Agent catalog filter
- `--query`: Agent catalog search query
- `--include-development`: Include development Agents
- `--limit`: Page size
- `--offset`: Page offset

## agent load-agent-package

- `--file-path`: Local file path (repeat only for connector uploads)
- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--max-upload-bytes`: Total byte limit for file-path input
- `--workspace-id`: Workspace ID (default: from config)

## agent my-feedback-tickets

- `--query-param`: Product query parameter name=value (repeatable; pagination/filter names follow the product API)
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

## agent post-channel-callback

- `--workspace-id`: Workspace ID (default: from config)
- `--provider`: Callback provider
- `--credential-ref`: Channel credential reference
- `--connection-id`: Ingress connection ID
- `--body-file`: Raw provider payload file; use - for stdin
- `--query-param`: Provider query parameter name=value (repeatable)

## agent post-ingress-callback

- `--workspace-id`: Workspace ID (default: from config)
- `--provider`: Callback provider
- `--credential-ref`: Channel credential reference
- `--connection-id`: Ingress connection ID
- `--body-file`: Raw provider payload file; use - for stdin
- `--query-param`: Provider query parameter name=value (repeatable)

## agent post-we-com-callback

- `--workspace-id`: Workspace ID (default: from config)
- `--provider`: Callback provider
- `--credential-ref`: Channel credential reference
- `--connection-id`: Ingress connection ID
- `--body-file`: Raw provider payload file; use - for stdin
- `--query-param`: Provider query parameter name=value (repeatable)

## agent preview-query-visual

- `--agent-id`: Agent ID
- `--file-id`: Query visual file ID
- `--output`: Write image to file (default stdout)
- `--workspace-id`: Workspace ID (default: from config)

## agent remove-team-member

- `--workspace-id`: Workspace ID (default: from config)
- `--team-id`: Team ID
- `--context-id`: Team Session context ID
- `--agent-workspace-id`: Member workspace ID
- `--agent-id`: Member Agent ID
- `--connection-id`: Ingress connection ID
- `--qr-session-token`: Existing WeCom QR session token (empty starts a session)
- `--expected-config-version`: Expected ingress configuration version
- `--target-kind`: Channel target kind: agent or team
- `--target-workspace-id`: Channel target workspace ID
- `--target-id`: Channel target ID
- `--expected-version`: Expected Team, membership or association version for this operation
- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--catalog`: Agent catalog filter
- `--query`: Agent catalog search query
- `--include-development`: Include development Agents
- `--limit`: Page size
- `--offset`: Page offset

## agent rename-team-session

- `--workspace-id`: Workspace ID (default: from config)
- `--team-id`: Team ID
- `--context-id`: Team Session context ID
- `--agent-workspace-id`: Member workspace ID
- `--agent-id`: Member Agent ID
- `--connection-id`: Ingress connection ID
- `--qr-session-token`: Existing WeCom QR session token (empty starts a session)
- `--expected-config-version`: Expected ingress configuration version
- `--target-kind`: Channel target kind: agent or team
- `--target-workspace-id`: Channel target workspace ID
- `--target-id`: Channel target ID
- `--expected-version`: Expected Team, membership or association version for this operation
- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--catalog`: Agent catalog filter
- `--query`: Agent catalog search query
- `--include-development`: Include development Agents
- `--limit`: Page size
- `--offset`: Page offset

## agent resolve-artifacts

- `--json`: 请求 JSON
- `--json-file`: 请求 JSON 文件
- `--workspace-id`: 工作区 ID，省略时读取配置

## agent resolve-we-com-qr-binding

- `--workspace-id`: Workspace ID (default: from config)
- `--team-id`: Team ID
- `--context-id`: Team Session context ID
- `--agent-workspace-id`: Member workspace ID
- `--agent-id`: Member Agent ID
- `--connection-id`: Ingress connection ID
- `--qr-session-token`: Existing WeCom QR session token (empty starts a session)
- `--expected-config-version`: Expected ingress configuration version
- `--target-kind`: Channel target kind: agent or team
- `--target-workspace-id`: Channel target workspace ID
- `--target-id`: Channel target ID
- `--expected-version`: Expected Team, membership or association version for this operation
- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--catalog`: Agent catalog filter
- `--query`: Agent catalog search query
- `--include-development`: Include development Agents
- `--limit`: Page size
- `--offset`: Page offset

## agent resource-any

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--method`: method
- `--query-param`: Product query parameter name=value (repeatable; pagination/filter names follow the product API)
- `--resource`: resource
- `--tail-path`: tailPath
- `--workspace-id`: Workspace ID (default: from config)

## agent resource-delete

- `--query-param`: Product query parameter name=value (repeatable; pagination/filter names follow the product API)
- `--resource`: resource
- `--tail-path`: tailPath
- `--workspace-id`: Workspace ID (default: from config)

## agent resource-get

- `--query-param`: Product query parameter name=value (repeatable; pagination/filter names follow the product API)
- `--resource`: resource
- `--tail-path`: tailPath
- `--workspace-id`: Workspace ID (default: from config)

## agent resource-patch

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--query-param`: Product query parameter name=value (repeatable; pagination/filter names follow the product API)
- `--resource`: resource
- `--tail-path`: tailPath
- `--workspace-id`: Workspace ID (default: from config)

## agent resource-post

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--query-param`: Product query parameter name=value (repeatable; pagination/filter names follow the product API)
- `--resource`: resource
- `--tail-path`: tailPath
- `--workspace-id`: Workspace ID (default: from config)

## agent resource-put

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--query-param`: Product query parameter name=value (repeatable; pagination/filter names follow the product API)
- `--resource`: resource
- `--tail-path`: tailPath
- `--workspace-id`: Workspace ID (default: from config)

## agent runtime-browser-token

- `--task-id`: Agent runtime task ID
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

## agent submit-team-message

- `--workspace-id`: Workspace ID (default: from config)
- `--team-id`: Team ID
- `--context-id`: Team Session context ID
- `--agent-workspace-id`: Member workspace ID
- `--agent-id`: Member Agent ID
- `--connection-id`: Ingress connection ID
- `--qr-session-token`: Existing WeCom QR session token (empty starts a session)
- `--expected-config-version`: Expected ingress configuration version
- `--target-kind`: Channel target kind: agent or team
- `--target-workspace-id`: Channel target workspace ID
- `--target-id`: Channel target ID
- `--expected-version`: Expected Team, membership or association version for this operation
- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--catalog`: Agent catalog filter
- `--query`: Agent catalog search query
- `--include-development`: Include development Agents
- `--limit`: Page size
- `--offset`: Page offset

## agent test-channel-configuration

- `--json`: Native channel JSON with config and secrets; prefer --json-file for credentials
- `--json-file`: Native channel JSON file; GitHub secrets.access_token accepts OAuth/PAT (no vault reference)
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
- `--json`: Native channel JSON with config and secrets; prefer --json-file for credentials
- `--json-file`: Native channel JSON file; GitHub secrets.access_token accepts OAuth/PAT (no vault reference)
- `--provider`: provider
- `--workspace-id`: Workspace ID (default: from config)

## agent update-ingress-connection

- `--workspace-id`: Workspace ID (default: from config)
- `--team-id`: Team ID
- `--context-id`: Team Session context ID
- `--agent-workspace-id`: Member workspace ID
- `--agent-id`: Member Agent ID
- `--connection-id`: Ingress connection ID
- `--qr-session-token`: Existing WeCom QR session token (empty starts a session)
- `--expected-config-version`: Expected ingress configuration version
- `--target-kind`: Channel target kind: agent or team
- `--target-workspace-id`: Channel target workspace ID
- `--target-id`: Channel target ID
- `--expected-version`: Expected Team, membership or association version for this operation
- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--catalog`: Agent catalog filter
- `--query`: Agent catalog search query
- `--include-development`: Include development Agents
- `--limit`: Page size
- `--offset`: Page offset

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

## agent update-team

- `--workspace-id`: Workspace ID (default: from config)
- `--team-id`: Team ID
- `--context-id`: Team Session context ID
- `--agent-workspace-id`: Member workspace ID
- `--agent-id`: Member Agent ID
- `--connection-id`: Ingress connection ID
- `--qr-session-token`: Existing WeCom QR session token (empty starts a session)
- `--expected-config-version`: Expected ingress configuration version
- `--target-kind`: Channel target kind: agent or team
- `--target-workspace-id`: Channel target workspace ID
- `--target-id`: Channel target ID
- `--expected-version`: Expected Team, membership or association version for this operation
- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--catalog`: Agent catalog filter
- `--query`: Agent catalog search query
- `--include-development`: Include development Agents
- `--limit`: Page size
- `--offset`: Page offset

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

## agent verify-channel-callback

- `--workspace-id`: Workspace ID (default: from config)
- `--provider`: Callback provider
- `--credential-ref`: Channel credential reference
- `--connection-id`: Ingress connection ID
- `--body-file`: Raw provider payload file; use - for stdin
- `--query-param`: Provider query parameter name=value (repeatable)

## agent verify-ingress-callback

- `--workspace-id`: Workspace ID (default: from config)
- `--provider`: Callback provider
- `--credential-ref`: Channel credential reference
- `--connection-id`: Ingress connection ID
- `--body-file`: Raw provider payload file; use - for stdin
- `--query-param`: Provider query parameter name=value (repeatable)

## agent verify-ingress-connection

- `--workspace-id`: Workspace ID (default: from config)
- `--team-id`: Team ID
- `--context-id`: Team Session context ID
- `--agent-workspace-id`: Member workspace ID
- `--agent-id`: Member Agent ID
- `--connection-id`: Ingress connection ID
- `--qr-session-token`: Existing WeCom QR session token (empty starts a session)
- `--expected-config-version`: Expected ingress configuration version
- `--target-kind`: Channel target kind: agent or team
- `--target-workspace-id`: Channel target workspace ID
- `--target-id`: Channel target ID
- `--expected-version`: Expected Team, membership or association version for this operation
- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--catalog`: Agent catalog filter
- `--query`: Agent catalog search query
- `--include-development`: Include development Agents
- `--limit`: Page size
- `--offset`: Page offset

## agent watch-browser

- `--workspace-id`: Workspace ID (default: from config)
- `--task-id`: Runtime task ID
- `--compression`: Negotiate browser stream compression

## agent workspace-mcp

- `--agent-id`: agent-id
- `--json`: Request JSON
- `--json-file`: Request JSON file
- `--runtime-token-file`: Runtime grant file
- `--workspace-id`: Workspace ID

