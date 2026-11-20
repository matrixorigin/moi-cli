# connector command reference

Generated from CLI source. Do not edit. Business requirements remain server-owned.
Use `moi-cli schema DOMAIN COMMAND` for typed JSON input and `--help` for defaults.

## connector create

- `--config-file`: Connector config JSON object; use a private file or - for stdin to keep credentials out of process arguments
- `--name`: name
- `--source-type`: sourceType
- `--usage-type`: usageType
- `--workspace-id`: Workspace ID (default: from config)

## connector create-export-task

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--workspace-id`: Workspace ID (default: from config)

## connector create-task

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--workspace-id`: Workspace ID (default: from config)

## connector db-list-databases

- `--connector-id`: connectorID
- `--workspace-id`: Workspace ID (default: from config)

## connector db-list-schemas

- `--connector-id`: connectorID
- `--database`: database
- `--workspace-id`: Workspace ID (default: from config)

## connector db-list-tables

- `--connector-id`: connectorID
- `--database`: database
- `--schema`: Source database schema
- `--workspace-id`: Workspace ID (default: from config)

## connector db-source-schema

- `--connector-id`: connectorID
- `--database`: database
- `--schema`: Source database schema
- `--table`: table
- `--workspace-id`: Workspace ID (default: from config)

## connector delete

- `--connector-id`: connectorID
- `--workspace-id`: Workspace ID (default: from config)

## connector delete-export-task

- `--task-id`: taskID
- `--workspace-id`: Workspace ID (default: from config)

## connector delete-task

- `--task-id`: taskID
- `--workspace-id`: Workspace ID (default: from config)

## connector export-task-files

- `--limit`: Page size (omit to use the MOI default)
- `--offset`: Page offset
- `--order-by`: Sort field
- `--order-direction`: Sort direction: asc or desc
- `--statuses`: Comma-separated file status values
- `--task-id`: taskID
- `--workspace-id`: Workspace ID (default: from config)

## connector export-task-info

- `--task-id`: taskID
- `--workspace-id`: Workspace ID (default: from config)

## connector export-task-state

- `--task-id`: taskID
- `--workspace-id`: Workspace ID (default: from config)

## connector export-task-stats

- `--workspace-id`: Workspace ID (default: from config)

## connector export-tasks

- `--connector-name`: Filter by connector name
- `--creator`: Filter by creator label
- `--keyword`: Search task, connector and source path
- `--limit`: Page size (omit to use the MOI default)
- `--offset`: Page offset
- `--order-by`: Sort field
- `--order-direction`: Sort direction: asc or desc
- `--statuses`: Comma-separated task status values
- `--target-category`: Target category: database, storage, model or other
- `--task-id`: Filter by export task ID
- `--workspace-id`: Workspace ID (default: from config)

## connector file-delete

- `--conn-file-id`: connFileID
- `--workspace-id`: Workspace ID (default: from config)

## connector file-download

- `--conn-file-id`: connFileID
- `--workspace-id`: Workspace ID (default: from config)

## connector file-stream

- `--conn-file-id`: connFileID
- `--output`: Write stream to file (default stdout)
- `--workspace-id`: Workspace ID (default: from config)

## connector files

- `--query-param`: Product query parameter name=value (repeatable; pagination/filter names follow the product API)
- `--workspace-id`: Workspace ID (default: from config)

## connector get

- `--connector-id`: connectorID
- `--workspace-id`: Workspace ID (default: from config)

## connector get-task

- `--task-id`: taskID
- `--workspace-id`: Workspace ID (default: from config)

## connector list

- `--query-param`: Product query parameter name=value (repeatable; pagination/filter names follow the product API)
- `--workspace-id`: Workspace ID (default: from config)

## connector local-upload

- `--file-path`: Local file path (repeat only for connector uploads)
- `--max-upload-bytes`: Total byte limit for file-path input
- `--workspace-id`: Workspace ID (default: from config)

## connector mo-check-table-name

- `--connector-id`: connectorID
- `--database`: database
- `--table`: table
- `--workspace-id`: Workspace ID (default: from config)

## connector mo-column-mapping

- `--connector-id`: connectorID
- `--database`: database
- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--table`: table
- `--workspace-id`: Workspace ID (default: from config)

## connector mo-database-list

- `--connector-id`: connectorID
- `--workspace-id`: Workspace ID (default: from config)

## connector mo-database-table-list

- `--connector-id`: connectorID
- `--database`: database
- `--workspace-id`: Workspace ID (default: from config)

## connector pause-task

- `--task-id`: taskID
- `--workspace-id`: Workspace ID (default: from config)

## connector preview-file

- `--conn-file-id`: Temporary file ID returned by local-upload or preview-file
- `--connector-id`: Source connector ID
- `--csv-delimiter`: CSV quote character (empty disables quoting)
- `--csv-escape`: Use backslash escapes in CSV quoted fields
- `--csv-separator`: CSV field separator; provide all three csv-* options together
- `--file-type`: Product file type code; omit to use the stored file type
- `--has-header`: Use a row as column names
- `--header-row`: Column-name row (1-based)
- `--row-start`: First data row (1-based)
- `--sheet-name`: Spreadsheet sheet name
- `--uri`: Source file URI
- `--workspace-id`: Workspace ID (default: from config)

## connector preview-rest

- `--connector-id`: connectorID
- `--json`: REST preview JSON
- `--json-file`: REST preview JSON file
- `--workspace-id`: Workspace ID (default: from config)

## connector probe-rest

- `--connector-id`: connectorID
- `--workspace-id`: Workspace ID (default: from config)

## connector probe-rest-draft

- `--config-file`: REST API connector config JSON object; use a private file or - for stdin to keep credentials out of process arguments
- `--workspace-id`: Workspace ID (default: from config)

## connector resume-task

- `--task-id`: taskID
- `--workspace-id`: Workspace ID (default: from config)

## connector retry-export-task

- `--file-ids`: fileIDs to rerun (comma-separated; empty reruns the whole task)
- `--task-id`: taskID
- `--workspace-id`: Workspace ID (default: from config)

## connector retry-task-files

- `--ids`: ids of task files to retry (comma-separated; empty retries all failed files)
- `--task-id`: taskID
- `--workspace-id`: Workspace ID (default: from config)

## connector run-task-full-refresh

- `--task-id`: taskID
- `--workspace-id`: Workspace ID (default: from config)

## connector run-task-now

- `--task-id`: taskID
- `--workspace-id`: Workspace ID (default: from config)

## connector structured-backfill-start-resolve

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--workspace-id`: Workspace ID (default: from config)

## connector structured-source-capability

- `--connector-id`: connectorID
- `--database`: database
- `--schema`: schema
- `--source-type`: sourceType
- `--table`: table
- `--workspace-id`: Workspace ID (default: from config)

## connector structured-source-databases

- `--connector-id`: connectorID
- `--source-type`: sourceType
- `--workspace-id`: Workspace ID (default: from config)

## connector structured-source-objects

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--workspace-id`: Workspace ID (default: from config)

## connector structured-source-sample

- `--connector-id`: connectorID
- `--limit`: limit
- `--source-type`: sourceType
- `--topic-mode`: topicMode
- `--topic-pattern`: topicPattern
- `--topics-json`: JSON array of topics
- `--workspace-id`: Workspace ID (default: from config)

## connector structured-source-schema

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--workspace-id`: Workspace ID (default: from config)

## connector structured-source-schemas

- `--connector-id`: connectorID
- `--database`: database
- `--source-type`: sourceType
- `--workspace-id`: Workspace ID (default: from config)

## connector structured-sources

- `--workspace-id`: Workspace ID (default: from config)

## connector task-files

- `--query-param`: Product query parameter name=value (repeatable; pagination/filter names follow the product API)
- `--workspace-id`: Workspace ID (default: from config)

## connector task-runs

- `--task-id`: taskID
- `--workspace-id`: Workspace ID (default: from config)

## connector task-stats

- `--workspace-id`: Workspace ID (default: from config)

## connector tasks

- `--connector-source`: Connector source type (repeatable)
- `--is-desc`: Sort descending (omit to use the MOI default)
- `--keyword`: Search task name
- `--load-interval-type`: Load interval type (repeatable)
- `--order-by`: Sort field
- `--page`: Page number (omit to use the MOI default)
- `--page-size`: Tasks per page (omit to use the MOI default)
- `--source-connector-id`: Source connector ID (repeatable)
- `--status`: Task status (repeatable)
- `--workspace-id`: Workspace ID (default: from config)

## connector update

- `--config-file`: Replacement connector config JSON object; private file or - for stdin (unchanged when omitted)
- `--connector-id`: connectorID
- `--name`: New connector name (unchanged when omitted)
- `--source-type`: New source type (unchanged when omitted)
- `--usage-type`: New usage type (unchanged when omitted)
- `--workspace-id`: Workspace ID (default: from config)

## connector update-task

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--task-id`: taskID
- `--workspace-id`: Workspace ID (default: from config)

## connector update-task-from-spec

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--task-id`: taskID
- `--workspace-id`: Workspace ID (default: from config)

## connector upload

- `--dedup`: Deduplication configuration JSON (schema: moi-cli schema connector upload --argument dedup)
- `--file-path`: Local file path (repeat only for connector uploads)
- `--file-types`: File types JSON (schema: moi-cli schema connector upload --argument file-types)
- `--max-upload-bytes`: Total byte limit for file-path input
- `--meta`: Connector upload metadata JSON (schema: moi-cli schema connector upload --argument meta)
- `--name`: Import task name (omitted: server-generated name)
- `--path-regex`: Path selection regular expression
- `--table-config`: Table import configuration JSON (schema: moi-cli schema connector upload --argument table-config)
- `--unzip-keep-structure`: Preserve archive directory structure
- `--volume-id`: Target VolumeID
- `--workspace-id`: Workspace ID (default: from config)

## connector validate

- `--config-file`: Connector config JSON object; use a private file or - for stdin to keep credentials out of process arguments
- `--name`: name
- `--source-type`: sourceType
- `--usage-type`: usageType
- `--workspace-id`: Workspace ID (default: from config)

