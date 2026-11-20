# sandbox command reference

Generated from CLI source. Do not edit. Business requirements remain server-owned.
Use `moi-cli schema DOMAIN COMMAND` for typed JSON input and `--help` for defaults.

## sandbox admin-delete

- `--sandbox-id`: Sandbox ID
- `--workspace-id`: Workspace ID (default: from config)

## sandbox admin-get

- `--sandbox-id`: Sandbox ID
- `--workspace-id`: Workspace ID (default: from config)

## sandbox admin-list

- `--query-param`: Query name=value (repeatable); workspace_id, limit, offset
- `--workspace-id`: Workspace ID (default: from config)

## sandbox admin-update-network

- `--workspace-id`: Workspace ID (default: from config)
- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--sandbox-id`: Sandbox ID

## sandbox admin-update-quota

- `--workspace-id`: Workspace ID (default: from config)
- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--target-workspace-id`: Target workspace ID (distinct from request context)

## sandbox connect-terminal

- `--cols`: Initial terminal columns
- `--rows`: Initial terminal rows
- `--sandbox-id`: Sandbox ID
- `--token`: Short-lived token returned by sandbox terminal-token
- `--workspace-id`: Workspace ID (default: from config)

## sandbox create

- `--workspace-id`: Workspace ID (default: from config)
- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file

## sandbox create-ssh-session

- `--sandbox-id`: Sandbox ID
- `--workspace-id`: Workspace ID (default: from config)

## sandbox delete

- `--sandbox-id`: Sandbox ID
- `--workspace-id`: Workspace ID (default: from config)

## sandbox get

- `--sandbox-id`: Sandbox ID
- `--workspace-id`: Workspace ID (default: from config)

## sandbox get-quota

- `--query-param`: Query name=value (repeatable); include_network_capabilities=true requests network capability metadata
- `--workspace-id`: Workspace ID (default: from config)

## sandbox list

- `--workspace-id`: Workspace ID (default: from config)

## sandbox terminal-token

- `--sandbox-id`: Sandbox ID
- `--workspace-id`: Workspace ID (default: from config)

## sandbox update-network

- `--workspace-id`: Workspace ID (default: from config)
- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--sandbox-id`: Sandbox ID

## sandbox update-quota

- `--workspace-id`: Workspace ID (default: from config)
- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file

