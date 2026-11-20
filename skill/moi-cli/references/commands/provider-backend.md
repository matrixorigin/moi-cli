# provider-backend command reference

Generated from CLI source. Do not edit. Business requirements remain server-owned.
Use `moi-cli schema DOMAIN COMMAND` for typed JSON input and `--help` for defaults.

## provider-backend create

- `--address`: Provider address
- `--api-key-file`: Provider API key file (read only when supplied)
- `--kind`: Backend kind: llm|embedding|parser
- `--models`: Provider models (comma-separated)
- `--name`: name
- `--reasoning-control-protocol`: Reasoning control protocol
- `--supported-mime-types`: Parser supported MIME types (comma-separated)
- `--timeout-seconds`: Provider timeout in seconds
- `--type`: Provider type
- `--workspace-id`: Workspace ID (default: from config)

## provider-backend create-endpoint

- `--address`: address
- `--backend-id`: backendID
- `--kind`: Backend kind: llm|embedding|parser
- `--workspace-id`: Workspace ID (default: from config)

## provider-backend delete

- `--id`: id
- `--kind`: Backend kind: llm|embedding|parser
- `--workspace-id`: Workspace ID (default: from config)

## provider-backend get

- `--id`: id
- `--kind`: Backend kind: llm|embedding|parser
- `--workspace-id`: Workspace ID (default: from config)

## provider-backend list

- `--kind`: Backend kind: llm|embedding|parser
- `--workspace-id`: Workspace ID (default: from config)

## provider-backend list-endpoints

- `--backend-id`: backendID
- `--kind`: Backend kind: llm|embedding|parser
- `--workspace-id`: Workspace ID (default: from config)

## provider-backend probe-backend-models

- `--api-key-file`: Provider API key file
- `--backend-id`: backendID
- `--kind`: Backend kind: llm|embedding|parser
- `--workspace-id`: Workspace ID (default: from config)

## provider-backend router

- `--kind`: Backend kind: llm|embedding|parser
- `--workspace-id`: Workspace ID (default: from config)

## provider-backend set-endpoint-status

- `--backend-id`: backendID
- `--endpoint-id`: endpointID
- `--kind`: Backend kind: llm|embedding|parser
- `--status`: status
- `--workspace-id`: Workspace ID (default: from config)

## provider-backend update

- `--api-key-file`: Provider API key file (empty file sends an explicit empty key)
- `--id`: id
- `--kind`: Backend kind: llm|embedding|parser
- `--models`: Provider models (comma-separated)
- `--name`: Updated backend name
- `--reasoning-control-protocol`: Reasoning control protocol
- `--supported-mime-types`: Supported MIME types (comma-separated)
- `--timeout-seconds`: Provider timeout in seconds
- `--workspace-id`: Workspace ID (default: from config)

## provider-backend update-router

- `--health-check-interval`: Router health check interval in seconds
- `--kind`: Backend kind: llm|embedding|parser
- `--max-retries`: Router maximum retries
- `--session-affinity`: Enable router session affinity
- `--strategy`: Router strategy
- `--workspace-id`: Workspace ID (default: from config)

