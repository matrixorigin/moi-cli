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

- `--query-id`: query_id returned by sql execute-query
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
- `--limit`: Result limit
- `--offset`: Result offset
- `--query`: SQL query
- `--query-file`: SQL file
- `--sql-type`: SQL query type
- `--workspace-id`: Workspace ID

## sql execute-query

- `--db-name`: Database name
- `--limit`: Result limit
- `--offset`: Result offset
- `--query`: query
- `--sql-type`: SQL query type
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

- `--end`: End time (RFC3339)
- `--scope`: History category: self or workspace
- `--start`: Start time (RFC3339)
- `--statement-id`: statementID
- `--workspace-id`: Workspace ID (default: from config)

## sql query-history

- `--cu`: Compute usage filter (CU)
- `--databases`: Comma-separated database names
- `--end`: End time (RFC3339)
- `--limit`: Result limit
- `--min-duration`: Minimum query duration
- `--non-user`: Include non-user SQL
- `--offset`: Result offset
- `--order`: Order direction
- `--order-by`: Comma-separated order fields
- `--query-types`: Comma-separated query types
- `--scope`: History category: self or workspace
- `--selected-fields`: Comma-separated result field names
- `--session-id`: Session ID
- `--sql-source-types`: Comma-separated SQL source types
- `--start`: Start time (RFC3339)
- `--statement`: Statement search pattern
- `--statement-id`: Statement ID
- `--status`: Statement status
- `--transaction-id`: Transaction ID
- `--users`: Comma-separated user names
- `--workspace-id`: Workspace ID (default: from config)

## sql query-history-overview

- `--end`: End time (RFC3339)
- `--non-user`: Include non-user SQL
- `--scope`: History category: self or workspace
- `--sql-source-types`: Comma-separated SQL source types
- `--start`: Start time (RFC3339)
- `--workspace-id`: Workspace ID (default: from config)

## sql query-profile

- `--end`: End time (RFC3339)
- `--scope`: History category: self or workspace
- `--start`: Start time (RFC3339)
- `--statement-id`: statementID
- `--workspace-id`: Workspace ID (default: from config)

## sql query-result

- `--limit`: Result limit
- `--offset`: Result offset
- `--statement-id`: statementID
- `--workspace-id`: Workspace ID (default: from config)

## sql query-web-socket-url

- `--workspace-id`: Workspace ID (default: from config)

## sql run

- `--db`: Database name
- `--limit`: Result limit
- `--offset`: Result offset
- `--query`: SQL query
- `--query-file`: SQL file
- `--sql-type`: SQL query type
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

- `--page-size`: Page size
- `--page-token`: Opaque cursor from the previous page
- `--workbook-id`: workbookID
- `--workspace-id`: Workspace ID (default: from config)

## sql workbooks

- `--page-size`: Page size
- `--page-token`: Opaque cursor from the previous page
- `--search`: Workbook name search
- `--workspace-id`: Workspace ID (default: from config)

