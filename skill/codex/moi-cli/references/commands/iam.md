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

## iam create-role

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--workspace-id`: Workspace ID (default: from config)

## iam current-principal-capabilities

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

## iam object-permission-view

- `--resource-id`: resourceID
- `--resource-type`: resourceType
- `--workspace-id`: Workspace ID (default: from config)

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

## iam role-delete-impact

- `--expected-lifecycle-version`: expectedLifecycleVersion
- `--role-id`: roleID
- `--workspace-id`: Workspace ID (default: from config)

## iam role-inheritance

- `--role-id`: Canonical IAM role ID
- `--workspace-id`: Workspace ID (default: from config)

## iam role-inheritance-graph

- `--role-id`: Canonical IAM role ID
- `--workspace-id`: Workspace ID (default: from config)

## iam role-members

- `--role-id`: role-id
- `--workspace-id`: Workspace ID

## iam role-permission-view

- `--role-id`: roleID
- `--workspace-id`: Workspace ID (default: from config)

## iam roles

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
