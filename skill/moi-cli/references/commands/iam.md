# iam command reference

Generated from CLI source. Do not edit. Business requirements remain server-owned.
Use `moi-cli schema DOMAIN COMMAND` for typed JSON input and `--help` for defaults.

## iam add-role-inheritance

- `--json`: JSON for IAM mutation
- `--json-file`: JSON file for IAM mutation
- `--role-id`: Canonical IAM role ID
- `--workspace-id`: Workspace ID (default: from config)

## iam apply-application-policy

- `--json`: JSON for IAM mutation
- `--json-file`: JSON file for IAM mutation
- `--role-id`: Canonical IAM role ID
- `--workspace-id`: Workspace ID (default: from config)

## iam apply-data-privileges

- `--json`: JSON for IAM data-privilege mutation
- `--json-file`: JSON file for IAM data-privilege mutation
- `--role-id`: Canonical IAM role ID
- `--workspace-id`: Workspace ID (default: from config)

## iam apply-object-grants

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--resource-id`: resourceID
- `--resource-type`: resourceType
- `--role-id`: roleID
- `--workspace-id`: Workspace ID (default: from config)

## iam batch-describe-permission-resources

- `--action-id`: IAM action ID
- `--purpose`: Resource picker purpose
- `--resource-id`: IAM resource ID (repeatable)
- `--resource-type`: IAM resource type
- `--workspace-id`: Workspace ID (default: from config)

## iam create-role

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--workspace-id`: Workspace ID (default: from config)

## iam current-principal-capabilities

- `--capability-id`: Capability ID to query (repeatable; omit for all; unsupported IDs are reported)
- `--workspace-id`: Workspace ID (default: from config)

## iam current-principal-roles

- `--workspace-id`: Workspace ID (default: from config)

## iam data-privilege-operation

- `--request-id`: Data-privilege operation request ID
- `--role-id`: Canonical IAM role ID
- `--workspace-id`: Workspace ID (default: from config)

## iam data-privileges

- `--role-id`: Canonical IAM role ID
- `--workspace-id`: Workspace ID (default: from config)

## iam delete-role

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--role-id`: roleID
- `--workspace-id`: Workspace ID (default: from config)

## iam list-permission-resources

- `--action-id`: IAM action ID
- `--cursor`: Pagination cursor
- `--limit`: Page size
- `--parent-id`: Parent resource ID
- `--purpose`: Resource picker purpose
- `--q`: Search text
- `--resource-type`: IAM resource type
- `--workspace-id`: Workspace ID (default: from config)

## iam object-permission-view

- `--resource-id`: resourceID
- `--resource-type`: resourceType
- `--workspace-id`: Workspace ID (default: from config)

## iam ownership-impact

- `--query-param`: 筛选或分页参数 name=value，可重复
- `--role-id`: 目标 ID
- `--workspace-id`: 工作区 ID，省略时读取配置

## iam permission-schema

- `--workspace-id`: Workspace ID (default: from config)

## iam put-user-role-bindings

- `--json`: JSON for IAM role-binding mutation
- `--json-file`: JSON file for IAM role-binding mutation
- `--use-current-versions`: Read current bindings and derive both expected CAS versions
- `--user-id`: Canonical IAM user ID
- `--workspace-id`: Workspace ID (default: from config)

## iam remove-role-inheritance

- `--json`: JSON for IAM mutation
- `--json-file`: JSON file for IAM mutation
- `--role-id`: Canonical IAM role ID
- `--workspace-id`: Workspace ID (default: from config)

## iam repair-data-privileges

- `--json`: JSON for IAM data-privilege repair
- `--json-file`: JSON file for IAM data-privilege repair
- `--repair-operation`: Data-privilege repair operation ID
- `--role-id`: Canonical IAM role ID
- `--workspace-id`: Workspace ID (default: from config)

## iam role-audit-events

- `--query-param`: 筛选或分页参数 name=value，可重复
- `--role-id`: 目标 ID
- `--workspace-id`: 工作区 ID，省略时读取配置

## iam role-delete-impact

- `--expected-lifecycle-version`: expectedLifecycleVersion
- `--include-ownership`: Include the live ownership page
- `--ownership-cursor`: Ownership pagination cursor
- `--ownership-limit`: Ownership page size
- `--role-id`: roleID
- `--transfer-requested`: Validate ownership transfer before deletion
- `--workspace-id`: Workspace ID (default: from config)

## iam role-delete-ownerships

- `--cursor`: Pagination cursor
- `--limit`: Page size
- `--role-id`: Canonical IAM role ID
- `--transfer-requested`: Read objects for ownership transfer
- `--workspace-id`: Workspace ID (default: from config)

## iam role-inheritance

- `--role-id`: Canonical IAM role ID
- `--workspace-id`: Workspace ID (default: from config)

## iam role-inheritance-graph

- `--role-id`: Canonical IAM role ID
- `--workspace-id`: Workspace ID (default: from config)

## iam role-members

- `--cursor`: Pagination cursor
- `--limit`: Page size
- `--role-id`: role-id
- `--workspace-id`: Workspace ID

## iam role-permission-view

- `--role-id`: roleID
- `--workspace-id`: Workspace ID (default: from config)

## iam roles

- `--workspace-id`: Workspace ID (default: from config)

## iam set-default-role

- `--json`: 请求 JSON
- `--json-file`: 请求 JSON 文件
- `--workspace-id`: 工作区 ID，省略时读取配置

## iam transfer-object-ownership

- `--json`: JSON for ownership transfer
- `--json-file`: JSON file for ownership transfer
- `--resource-id`: IAM resource ID
- `--resource-type`: IAM resource type
- `--workspace-id`: Workspace ID (default: from config)

## iam update-role

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--role-id`: roleID
- `--workspace-id`: Workspace ID (default: from config)

## iam update-role-lifecycle

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--role-id`: roleID
- `--workspace-id`: Workspace ID (default: from config)

## iam user-audit-events

- `--query-param`: 筛选或分页参数 name=value，可重复
- `--user-id`: 目标 ID
- `--workspace-id`: 工作区 ID，省略时读取配置

## iam user-role-bindings

- `--user-id`: Canonical IAM user ID
- `--workspace-id`: Workspace ID (default: from config)

## iam validate-application-policy

- `--json`: JSON for IAM mutation
- `--json-file`: JSON file for IAM mutation
- `--role-id`: Canonical IAM role ID
- `--workspace-id`: Workspace ID (default: from config)

## iam validate-data-privileges

- `--json`: JSON for IAM data-privilege mutation
- `--json-file`: JSON file for IAM data-privilege mutation
- `--role-id`: Canonical IAM role ID
- `--workspace-id`: Workspace ID (default: from config)

## iam validate-object-grants

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--resource-id`: resourceID
- `--resource-type`: resourceType
- `--role-id`: roleID
- `--workspace-id`: Workspace ID (default: from config)

