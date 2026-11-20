# model command reference

Generated from CLI source. Do not edit. Business requirements remain server-owned.
Use `moi-cli schema DOMAIN COMMAND` for typed JSON input and `--help` for defaults.

## model create-workspace-embedding-backend

- `--json`: Request JSON
- `--json-file`: Request JSON file
- `--workspace-id`: Workspace ID

## model create-workspace-embedding-endpoint

- `--address`: address
- `--backend-id`: backend-id
- `--workspace-id`: Workspace ID

## model create-workspace-llm-backend

- `--json`: json
- `--json-file`: json-file
- `--workspace-id`: workspace-id

## model create-workspace-llm-endpoint

- `--address`: address
- `--backend-id`: backend-id
- `--workspace-id`: workspace-id

## model delete-workspace-embedding-backend

- `--backend-id`: backend-id
- `--workspace-id`: Workspace ID

## model delete-workspace-llm-backend

- `--backend-id`: backend-id
- `--workspace-id`: workspace-id

## model get-workspace-embedding-backend

- `--backend-id`: backend-id
- `--workspace-id`: Workspace ID

## model get-workspace-llm-backend

- `--backend-id`: backend-id
- `--workspace-id`: workspace-id

## model list

- `--workspace-id`: Workspace ID (default: from config)

## model list-embedding-models

- `--workspace-id`: Workspace ID (default: from config)

## model list-workspace-embedding-backends

- `--workspace-id`: Workspace ID

## model list-workspace-embedding-endpoints

- `--backend-id`: backend-id
- `--workspace-id`: Workspace ID

## model list-workspace-llm-backends

- `--workspace-id`: workspace-id

## model list-workspace-llm-endpoints

- `--backend-id`: backend-id
- `--workspace-id`: Workspace ID

## model list-workspace-llm-models

- `--workspace-id`: Workspace ID

## model probe-embedding-models

- `--address`: address
- `--workspace-id`: Workspace ID (default: from config)

## model probe-llm-models

- `--address`: address
- `--workspace-id`: Workspace ID (default: from config)

## model set-workspace-embedding-endpoint-status

- `--backend-id`: backend-id
- `--endpoint-id`: endpoint-id
- `--status`: status
- `--workspace-id`: Workspace ID

## model set-workspace-llm-endpoint-status

- `--backend-id`: backend-id
- `--endpoint-id`: endpoint-id
- `--status`: status
- `--workspace-id`: Workspace ID

## model setup-llm-backend

- `--address`: address
- `--models`: models (comma-separated)
- `--name`: name
- `--workspace-id`: Workspace ID (default: from config)

## model update-workspace-embedding-backend

- `--backend-id`: backend-id
- `--json`: Request JSON
- `--json-file`: Request JSON file
- `--workspace-id`: Workspace ID

## model update-workspace-embedding-router

- `--json`: Request JSON
- `--json-file`: Request JSON file
- `--workspace-id`: Workspace ID

## model update-workspace-llm-backend

- `--backend-id`: backend-id
- `--json`: json
- `--json-file`: json-file
- `--workspace-id`: workspace-id

## model update-workspace-llm-router

- `--json`: json
- `--json-file`: json-file
- `--workspace-id`: workspace-id

## model workspace-chat-completion

- `--json`: json
- `--json-file`: json-file
- `--service-slot`: trusted service slot for billing scenario attribution
- `--workspace-id`: workspace-id

## model workspace-chat-completion-stream

- `--json`: json
- `--json-file`: json-file
- `--output`: Output SSE file, or stdout
- `--service-slot`: trusted service slot for billing scenario attribution
- `--workspace-id`: workspace-id

## model workspace-embedding-router

- `--workspace-id`: Workspace ID

## model workspace-llm-router

- `--workspace-id`: workspace-id
