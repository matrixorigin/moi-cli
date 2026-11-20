# provider-backend command reference

Generated from CLI source. Do not edit. Business requirements remain server-owned.
Use `moi-cli schema DOMAIN COMMAND` for typed JSON input and `--help` for defaults.

## provider-backend create

- `--kind`: Backend kind: llm|embedding|parser
- `--name`: name
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

- `--id`: id
- `--kind`: Backend kind: llm|embedding|parser
- `--workspace-id`: Workspace ID (default: from config)

## provider-backend update-router

- `--kind`: Backend kind: llm|embedding|parser
- `--workspace-id`: Workspace ID (default: from config)
