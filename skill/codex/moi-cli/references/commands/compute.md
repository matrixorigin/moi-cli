# compute command reference

Generated from CLI source. Do not edit. Business requirements remain server-owned.
Use `moi-cli schema DOMAIN COMMAND` for typed JSON input and `--help` for defaults.

## compute admin-delete

- `--id`: id
- `--workspace-id`: Workspace ID (default: from config)

## compute admin-list

- `--workspace-id`: Workspace ID (default: from config)

## compute admin-preflight-delete

- `--id`: id
- `--workspace-id`: Workspace ID (default: from config)

## compute admin-resume

- `--id`: id
- `--workspace-id`: Workspace ID (default: from config)

## compute admin-retry

- `--id`: id
- `--workspace-id`: Workspace ID (default: from config)

## compute admin-runtime

- `--compute-resource-id`: Compute resource ID (required)
- `--workspace-id`: Workspace ID (default: from config)

## compute admin-suspend

- `--id`: id
- `--workspace-id`: Workspace ID (default: from config)

## compute admin-update

- `--cpu`: CPU cores
- `--description`: Compute resource description
- `--go-worker-image-id`: Go worker image ID
- `--gpu`: GPU count
- `--id`: Compute resource ID (required)
- `--max-replicas`: Maximum replicas
- `--memory-gib`: Memory in GiB
- `--min-replicas`: Minimum replicas
- `--name`: Compute resource name
- `--platform`: Worker platform
- `--python-worker-image-id`: Python worker image ID
- `--spec-id`: Compute resource spec ID
- `--workspace-id`: Workspace ID (default: from config)

## compute adopt-legacy-ownership

- `--id`: Compute resource ID (required)
- `--workspace-id`: Workspace ID (default: from config)

## compute create

- `--cpu`: cpu
- `--memory-gi-b`: memoryGiB
- `--name`: name
- `--spec-id`: specID
- `--workspace-id`: Workspace ID (default: from config)

## compute create-spec

- `--cpu-milli`: cpuMilli
- `--family`: family
- `--family-name`: familyName
- `--family-name-en`: familyNameEN
- `--kind`: kind
- `--memory-mi-b`: memoryMiB
- `--spec-id`: specID
- `--workspace-id`: Workspace ID (default: from config)

## compute delete

- `--id`: id
- `--workspace-id`: Workspace ID (default: from config)

## compute get

- `--id`: id
- `--workspace-id`: Workspace ID (default: from config)

## compute list

- `--workspace-id`: Workspace ID (default: from config)

## compute preflight-delete

- `--id`: id
- `--workspace-id`: Workspace ID (default: from config)

## compute resume

- `--id`: id
- `--workspace-id`: Workspace ID (default: from config)

## compute retry

- `--id`: id
- `--workspace-id`: Workspace ID (default: from config)

## compute runtime

- `--compute-resource-id`: Compute resource ID (required)
- `--workspace-id`: Workspace ID (default: from config)

## compute set-spec-enabled

- `--enabled`: enabled
- `--spec-id`: specID
- `--workspace-id`: Workspace ID (default: from config)

## compute spec-management-permission

- `--workspace-id`: Workspace ID (default: from config)

## compute specs

- `--workspace-id`: Workspace ID (default: from config)

## compute suspend

- `--id`: id
- `--workspace-id`: Workspace ID (default: from config)

## compute update

- `--id`: id
- `--workspace-id`: Workspace ID (default: from config)

## compute update-spec

- `--spec-id`: specID
- `--workspace-id`: Workspace ID (default: from config)

## compute worker-images

- `--workspace-id`: Workspace ID (default: from config)
