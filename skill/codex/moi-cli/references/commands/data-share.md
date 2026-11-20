# data-share command reference

Generated from CLI source. Do not edit. Business requirements remain server-owned.
Use `moi-cli schema DOMAIN COMMAND` for typed JSON input and `--help` for defaults.

## data-share check-publish-name

- `--workspace-id`: Workspace ID (default: from config)

## data-share create-publish

- `--name`: name
- `--source-database-id`: source database ID
- `--table-ids`: MatrixOne table rel_ids (comma-separated; required for selected)
- `--table-scope-mode`: table scope mode: all or selected
- `--target-workspace-ids`: target workspace IDs (comma-separated)
- `--workspace-id`: Workspace ID (default: from config)

## data-share delete-publish

- `--publish-id`: publishID
- `--workspace-id`: Workspace ID (default: from config)

## data-share publish-summary

- `--workspace-id`: Workspace ID (default: from config)

## data-share publishes

- `--workspace-id`: Workspace ID (default: from config)

## data-share resolve-workspace-names

- `--ids`: ids (comma-separated)
- `--workspace-id`: Workspace ID (default: from config)

## data-share source-tables

- `--source-database-id`: source database ID
- `--workspace-id`: Workspace ID (default: from config)

## data-share subscribe

- `--sub-name`: subName
- `--subscription-id`: subscriptionID
- `--workspace-id`: Workspace ID (default: from config)

## data-share subscriptions

- `--workspace-id`: Workspace ID (default: from config)

## data-share unsubscribe

- `--subscription-id`: subscriptionID
- `--workspace-id`: Workspace ID (default: from config)

## data-share update-publish

- `--publish-id`: publishID
- `--workspace-id`: Workspace ID (default: from config)

## data-share workspace-names-result

- `--ids`: ids (comma-separated)
- `--workspace-id`: Workspace ID (default: from config)
