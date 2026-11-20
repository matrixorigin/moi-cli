# sql command reference

Generated from CLI source. Do not edit. Business requirements remain server-owned.
Use `moi-cli schema DOMAIN COMMAND` for typed JSON input and `--help` for defaults.

## sql connect-query-web-socket

- `--workspace-id`: Workspace ID

## sql create-workbook

- `--name`: name
- `--workspace-id`: Workspace ID (default: from config)

## sql create-workbook-version

- `--sql-content`: sqlContent
- `--workbook-id`: workbookID
- `--workspace-id`: Workspace ID (default: from config)

## sql delete-workbook

- `--workbook-id`: workbookID
- `--workspace-id`: Workspace ID (default: from config)

## sql describe-query

- `--query-id`: queryID
- `--workspace-id`: Workspace ID (default: from config)

## sql download-query-result

- `--output`: Write CSV to file (default stdout)
- `--statement-id`: Statement ID
- `--workspace-id`: Workspace ID (default: from config)

## sql drop-meta-database

- `--db-name`: dbName
- `--workspace-id`: Workspace ID (default: from config)

## sql drop-meta-table

- `--db-name`: dbName
- `--table-name`: tableName
- `--workspace-id`: Workspace ID (default: from config)

## sql drop-meta-view

- `--db-name`: dbName
- `--table-name`: tableName
- `--workspace-id`: Workspace ID (default: from config)

## sql execute

- `--db`: Database name
- `--limit`: Limit
- `--query`: SQL query
- `--query-file`: SQL file
- `--workspace-id`: Workspace ID

## sql execute-query

- `--query`: query
- `--workspace-id`: Workspace ID (default: from config)

## sql kill-query

- `--query-id`: queryID
- `--workspace-id`: Workspace ID (default: from config)

## sql meta-databases

- `--workspace-id`: Workspace ID (default: from config)

## sql meta-db-tables

- `--db-name`: dbName
- `--workspace-id`: Workspace ID (default: from config)

## sql meta-db-tree

- `--workspace-id`: Workspace ID (default: from config)

## sql meta-table-column-stats

- `--db-name`: dbName
- `--table-name`: tableName
- `--workspace-id`: Workspace ID (default: from config)

## sql meta-table-columns

- `--db-name`: dbName
- `--table-name`: tableName
- `--workspace-id`: Workspace ID (default: from config)

## sql meta-table-ddl

- `--db-name`: dbName
- `--table-name`: tableName
- `--workspace-id`: Workspace ID (default: from config)

## sql meta-table-info

- `--db-name`: dbName
- `--table-name`: tableName
- `--workspace-id`: Workspace ID (default: from config)

## sql query-detail

- `--statement-id`: statementID
- `--workspace-id`: Workspace ID (default: from config)

## sql query-history

- `--workspace-id`: Workspace ID (default: from config)

## sql query-history-overview

- `--workspace-id`: Workspace ID (default: from config)

## sql query-profile

- `--statement-id`: statementID
- `--workspace-id`: Workspace ID (default: from config)

## sql query-result

- `--statement-id`: statementID
- `--workspace-id`: Workspace ID (default: from config)

## sql query-web-socket-url

- `--workspace-id`: Workspace ID (default: from config)

## sql run

- `--db`: Database name
- `--limit`: Limit
- `--query`: SQL query
- `--query-file`: SQL file
- `--workspace-id`: Workspace ID

## sql save-workbook-version

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--version-id`: versionID
- `--workbook-id`: workbookID
- `--workspace-id`: Workspace ID (default: from config)

## sql update-workbook

- `--name`: name
- `--workbook-id`: workbookID
- `--workspace-id`: Workspace ID (default: from config)

## sql update-workbook-version

- `--sql-content`: sqlContent
- `--version-id`: versionID
- `--workbook-id`: workbookID
- `--workspace-id`: Workspace ID (default: from config)

## sql workbook-detail

- `--version-id`: versionID
- `--workbook-id`: workbookID
- `--workspace-id`: Workspace ID (default: from config)

## sql workbook-versions

- `--workbook-id`: workbookID
- `--workspace-id`: Workspace ID (default: from config)

## sql workbooks

- `--workspace-id`: Workspace ID (default: from config)
