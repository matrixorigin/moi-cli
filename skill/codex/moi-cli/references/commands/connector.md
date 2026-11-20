# connector command reference

Generated from CLI source. Do not edit. Business requirements remain server-owned.
Use `moi-cli schema DOMAIN COMMAND` for typed JSON input and `--help` for defaults.

## connector create

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
- `--workspace-id`: Workspace ID (default: from config)

## connector db-source-schema

- `--connector-id`: connectorID
- `--database`: database
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

- `--keyword`: Search task, connector and source path
- `--limit`: Page size
- `--offset`: Page offset
- `--target-category`: Target category: database, storage, model or other
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

- `--workspace-id`: Workspace ID (default: from config)

## connector get

- `--connector-id`: connectorID
- `--workspace-id`: Workspace ID (default: from config)

## connector get-task

- `--task-id`: taskID
- `--workspace-id`: Workspace ID (default: from config)

## connector list

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

- `--connector-id`: connectorID
- `--uri`: uri
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

- `--workspace-id`: Workspace ID (default: from config)

## connector task-runs

- `--task-id`: taskID
- `--workspace-id`: Workspace ID (default: from config)

## connector task-stats

- `--workspace-id`: Workspace ID (default: from config)

## connector tasks

- `--config-type`: Task configuration type: 1=file, 2=legacy database, 3=structured load (repeatable)
- `--page`: Page number
- `--page-size`: Tasks per page
- `--structured`: Filter structured tasks: true or false (omit for all)
- `--workspace-id`: Workspace ID (default: from config)

## connector update

- `--connector-id`: connectorID
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

- `--dedup`: Deduplication configuration JSON
- `--file-path`: Local file path (repeat only for connector uploads)
- `--file-types`: File types JSON
- `--max-upload-bytes`: Total byte limit for file-path input
- `--meta`: Connector upload metadata JSON
- `--path-regex`: Path selection regular expression
- `--table-config`: Table import configuration JSON
- `--unzip-keep-structure`: Preserve archive directory structure
- `--volume-id`: Target VolumeID
- `--workspace-id`: Workspace ID (default: from config)

## connector validate

- `--name`: name
- `--source-type`: sourceType
- `--usage-type`: usageType
- `--workspace-id`: Workspace ID (default: from config)
