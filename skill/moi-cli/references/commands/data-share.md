# data-share command reference

Generated from CLI source. Do not edit. Business requirements remain server-owned.
Use `moi-cli schema DOMAIN COMMAND` for typed JSON input and `--help` for defaults.

## data-share check-publish-name

- `--name`: Publication name to check
- `--workspace-id`: Workspace ID (default: from config)

## data-share create-publish

- `--name`: name
- `--remark`: Publication remark (omitted: no remark; explicit empty sends an empty remark)
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

- `--keyword`: Publication name filter
- `--page`: Page number (omitted: server default)
- `--page-size`: Page size (omitted: server default)
- `--workspace-id`: Workspace ID (default: from config)

## data-share resolve-workspace-names

- `--ids`: ids (comma-separated)
- `--workspace-id`: Workspace ID (default: from config)

## data-share source-tables

- `--page`: Page number (omitted: server default)
- `--page-size`: Page size (omitted: server default)
- `--source-database-id`: source database ID
- `--workspace-id`: Workspace ID (default: from config)

## data-share subscribe

- `--sub-name`: subName
- `--subscription-id`: subscriptionID
- `--workspace-id`: Workspace ID (default: from config)

## data-share subscriptions

- `--keyword`: Subscription name filter
- `--page`: Page number (omitted: server default)
- `--page-size`: Page size (omitted: server default)
- `--status`: Subscription status filter
- `--workspace-id`: Workspace ID (default: from config)

## data-share unsubscribe

- `--subscription-id`: subscriptionID
- `--workspace-id`: Workspace ID (default: from config)

## data-share update-publish

- `--publish-id`: publishID
- `--remark`: Replacement remark (empty clears it)
- `--table-ids`: Replacement MatrixOne table rel_ids (comma-separated; use with selected scope)
- `--table-scope-mode`: Replacement scope: all or selected (provide when changing tables)
- `--target-workspace-ids`: Replacement target workspace IDs (comma-separated)
- `--workspace-id`: Workspace ID (default: from config)

## data-share workspace-names-result

- `--ids`: ids (comma-separated)
- `--workspace-id`: Workspace ID (default: from config)

