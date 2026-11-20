# compute command reference

Generated from CLI source. Do not edit. Business requirements remain server-owned.
Use `moi-cli schema DOMAIN COMMAND` for typed JSON input and `--help` for defaults.

## compute admin-delete

- `--id`: id
- `--query-param`: Product query parameter name=value (repeatable; pagination/filter names follow the product API)
- `--workspace-id`: Workspace ID (default: from config)

## compute admin-list

- `--query-param`: Product query parameter name=value (repeatable)
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

- `--auto-suspend-minutes`: Idle timeout in minutes (only submitted when specified)
- `--cpu`: CPU cores
- `--description`: Compute resource description
- `--go-worker-image-id`: Go worker image ID
- `--gpu`: GPU count
- `--id`: Compute resource ID (required)
- `--kind`: Compute resource kind
- `--max-replicas`: Maximum replicas
- `--memory-gib`: Memory in GiB
- `--min-replicas`: Minimum replicas
- `--name`: Compute resource name
- `--platform`: Worker platform
- `--py-requirements-txt`: Python worker requirements.txt contents; empty clears dependencies
- `--python-worker-image-id`: Python worker image ID
- `--spec-id`: Compute resource spec ID
- `--worker-images-file`: JSON file containing worker image selections; [] clears selection
- `--workspace-id`: Workspace ID (default: from config)

## compute adopt-legacy-ownership

- `--id`: Compute resource ID (required)
- `--workspace-id`: Workspace ID (default: from config)

## compute create

- `--auto-suspend-minutes`: Idle timeout in minutes (omit for the SDK default)
- `--cpu`: cpu
- `--description`: Compute resource description
- `--go-worker-image-id`: Go worker image ID
- `--gpu`: GPU count
- `--kind`: Compute resource kind
- `--max-replicas`: Maximum replicas
- `--memory-gi-b`: memoryGiB
- `--min-replicas`: Minimum replicas
- `--name`: name
- `--platform`: Worker platform
- `--py-requirements-txt`: Python worker requirements.txt contents
- `--python-worker-image-id`: Python worker image ID
- `--spec-id`: specID
- `--worker-images-file`: JSON file containing worker image selections
- `--workspace-id`: Workspace ID (default: from config)

## compute create-spec

- `--cpu-milli`: cpuMilli
- `--credit-per-hour`: Credits per hour
- `--description`: Spec description
- `--enabled`: Enabled value; use set-spec-enabled to disable an existing spec
- `--family`: family
- `--family-name`: familyName
- `--family-name-en`: familyNameEN
- `--gpu-cores`: GPU cores
- `--gpu-count`: GPU count
- `--gpu-memory-mib`: GPU memory in MiB
- `--kind`: kind
- `--memory-mi-b`: memoryMiB
- `--node-placement-file`: JSON file containing node placement; null clears placement
- `--spec-id`: specID
- `--workspace-id`: Workspace ID (default: from config)

## compute delete

- `--id`: id
- `--query-param`: Product query parameter name=value (repeatable; pagination/filter names follow the product API)
- `--workspace-id`: Workspace ID (default: from config)

## compute get

- `--id`: id
- `--workspace-id`: Workspace ID (default: from config)

## compute list

- `--query-param`: Product query parameter name=value (repeatable; pagination/filter names follow the product API)
- `--workspace-id`: Workspace ID (default: from config)

## compute preflight-delete

- `--id`: id
- `--workspace-id`: Workspace ID (default: from config)

## compute py-deps-limits

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

- `--auto-suspend-minutes`: Idle timeout in minutes (only submitted when specified)
- `--cpu`: CPU cores
- `--description`: Compute resource description
- `--go-worker-image-id`: Go worker image ID
- `--gpu`: GPU count
- `--id`: id
- `--kind`: Compute resource kind
- `--max-replicas`: Maximum replicas
- `--memory-gib`: Memory in GiB
- `--min-replicas`: Minimum replicas
- `--name`: Compute resource name
- `--platform`: Worker platform
- `--py-requirements-txt`: Python worker requirements.txt contents; empty clears dependencies
- `--python-worker-image-id`: Python worker image ID
- `--spec-id`: Compute resource spec ID
- `--worker-images-file`: JSON file containing worker image selections; [] clears selection
- `--workspace-id`: Workspace ID (default: from config)

## compute update-spec

- `--cpu-milli`: CPU in millicores
- `--credit-per-hour`: Credits per hour
- `--description`: Spec description
- `--enabled`: Enabled value; use set-spec-enabled to disable an existing spec
- `--family`: Spec family
- `--family-name`: Spec family display name
- `--family-name-en`: English spec family display name
- `--gpu-cores`: GPU cores
- `--gpu-count`: GPU count
- `--gpu-memory-mib`: GPU memory in MiB
- `--kind`: Spec kind
- `--memory-mib`: Memory in MiB
- `--node-placement-file`: JSON file containing node placement; null clears placement
- `--spec-id`: specID
- `--workspace-id`: Workspace ID (default: from config)

## compute worker-images

- `--query-param`: Product query parameter name=value (repeatable; pagination/filter names follow the product API)
- `--workspace-id`: Workspace ID (default: from config)

## compute workers

- `--compute-resource-id`: Compute resource ID (required)
- `--workspace-id`: Workspace ID (default: from config)

