# catalog command reference

Generated from CLI source. Do not edit. Business requirements remain server-owned.
Use `moi-cli schema DOMAIN COMMAND` for typed JSON input and `--help` for defaults.

## catalog allowed-actions

- `--resource-ids`: Comma-separated resource IDs (1–100)
- `--resource-type`: catalog, database, or volume
- `--workspace-id`: Workspace ID (default: from config)

## catalog batch-get-data-assets

- `--raw-file-ids`: rawFileIDs (comma-separated)
- `--workspace-id`: Workspace ID (default: from config)

## catalog batch-resolve-file-artifacts

- `--file-ids`: fileIDs (comma-separated)
- `--workspace-id`: Workspace ID (default: from config)

## catalog catalog-file-content

- `--capability`: capability
- `--file-id`: file-id
- `--output`: Output file, or stdout
- `--workspace-id`: Workspace ID

## catalog catalog-references

- `--id`: id
- `--workspace-id`: workspace-id

## catalog clean-folder

- `--folder-id`: folderID
- `--workspace-id`: Workspace ID (default: from config)

## catalog create

- `--name`: name
- `--workspace-id`: Workspace ID (default: from config)

## catalog create-database

- `--catalog-id`: catalogID
- `--name`: name
- `--workspace-id`: Workspace ID (default: from config)

## catalog create-file

- `--file-ids`: fileIDs (comma-separated)
- `--volume-id`: volumeID
- `--workspace-id`: Workspace ID (default: from config)

## catalog create-folder

- `--database-id`: databaseID
- `--description`: Folder description
- `--name`: name
- `--parent-volume-id`: parentVolumeID
- `--workspace-id`: Workspace ID (default: from config)

## catalog create-table

- `--column-load-options`: JSON array of column load options
- `--comment`: Table comment
- `--data-file-url`: Source data file URL
- `--database-id`: databaseID
- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--name`: name
- `--workspace-id`: Workspace ID (default: from config)

## catalog create-volume

- `--database-id`: databaseID
- `--name`: name
- `--workspace-id`: Workspace ID (default: from config)

## catalog data-asset

- `--asset-id`: assetID
- `--workspace-id`: Workspace ID (default: from config)

## catalog database-children

- `--database-id`: databaseID
- `--include-stats`: Include child statistics (omit for server default)
- `--page-size`: Page size (omit for full list)
- `--page-token`: Continuation token returned by the previous page
- `--workspace-id`: Workspace ID (default: from config)

## catalog database-info

- `--database-id`: databaseID
- `--workspace-id`: Workspace ID (default: from config)

## catalog database-references

- `--id`: id
- `--workspace-id`: workspace-id

## catalog database-table-leaves

- `--database-id`: databaseID
- `--page-size`: Page size (omit to use the MOI default)
- `--page-token`: Cursor from the previous page
- `--search`: Search table names
- `--workspace-id`: Workspace ID (default: from config)

## catalog delete

- `--catalog-id`: catalogID
- `--workspace-id`: Workspace ID (default: from config)

## catalog delete-database

- `--database-id`: databaseID
- `--workspace-id`: Workspace ID (default: from config)

## catalog delete-file

- `--file-id`: fileID
- `--volume-id`: volumeID
- `--workspace-id`: Workspace ID (default: from config)

## catalog delete-file-ref

- `--file-ids`: fileIDs (comma-separated)
- `--volume-id`: volumeID
- `--workspace-id`: Workspace ID (default: from config)

## catalog delete-folder

- `--folder-id`: folderID
- `--workspace-id`: Workspace ID (default: from config)

## catalog delete-table

- `--table-id`: tableID
- `--workspace-id`: Workspace ID (default: from config)

## catalog delete-volume

- `--volume-id`: volumeID
- `--workspace-id`: Workspace ID (default: from config)

## catalog download-file

- `--as-zip`: Download the file or folder as a ZIP archive
- `--file-id`: fileID
- `--output`: Write stream to file (default stdout)
- `--volume-id`: volumeID
- `--workspace-id`: Workspace ID (default: from config)

## catalog download-table

- `--output`: Write stream to file (default stdout)
- `--table-id`: tableID
- `--workspace-id`: Workspace ID (default: from config)

## catalog file-artifact

- `--file-id`: fileID
- `--workspace-id`: Workspace ID (default: from config)

## catalog file-data-asset

- `--file-id`: fileID
- `--workspace-id`: Workspace ID (default: from config)

## catalog file-info

- `--file-id`: fileID
- `--workspace-id`: Workspace ID (default: from config)

## catalog files

- `--extensions`: File extensions (comma-separated)
- `--file-ids`: File IDs (comma-separated)
- `--file-type`: File entry type
- `--name`: File name substring
- `--order`: Sort direction
- `--order-by`: Sort field
- `--page`: Page number (omit for server default)
- `--page-size`: Page size (omit for server default)
- `--parent-id`: Parent folder ID (empty selects the Volume root)
- `--volume-id`: Volume ID to list
- `--workspace-id`: Workspace ID (default: from config)

## catalog folder-references

- `--id`: id
- `--workspace-id`: workspace-id

## catalog garbage-collection

- `--json`: json
- `--json-file`: json-file
- `--workspace-id`: workspace-id

## catalog health

- `--workspace-id`: Workspace ID (default: from config)

## catalog info

- `--catalog-id`: catalogID
- `--workspace-id`: Workspace ID (default: from config)

## catalog list

- `--include-stats`: Include child statistics (omit for server default)
- `--page-size`: Page size (omit for full list)
- `--page-token`: Continuation token returned by the previous page
- `--workspace-id`: Workspace ID (default: from config)

## catalog list-databases

- `--catalog-id`: catalogID
- `--include-stats`: Include child statistics (omit for server default)
- `--page-size`: Page size (omit for full list)
- `--page-token`: Continuation token returned by the previous page
- `--workspace-id`: Workspace ID (default: from config)

## catalog multi-table-info

- `--table-ids`: tableIDs (comma-separated)
- `--workspace-id`: Workspace ID (default: from config)

## catalog prepare

- `--description`: Description
- `--name`: Catalog name
- `--workspace-id`: Workspace ID

## catalog preview-file

- `--file-id`: fileID
- `--output`: Write stream to file (default stdout)
- `--volume-id`: volumeID
- `--workspace-id`: Workspace ID (default: from config)

## catalog preview-table

- `--table-id`: tableID
- `--workspace-id`: Workspace ID (default: from config)

## catalog preview-workspace-file

- `--file-id`: file-id
- `--output`: output
- `--workspace-id`: workspace-id

## catalog run-nl2sql

- `--database-id`: databaseID
- `--sql`: sql
- `--workspace-id`: Workspace ID (default: from config)

## catalog search-data-dashboard-tables

- `--page-size`: Page size (omit to use the MOI default)
- `--page-token`: Cursor from the previous page
- `--search`: Search tables across databases
- `--workspace-id`: Workspace ID (default: from config)

## catalog system-tree

- `--database-name`: Select a system database
- `--include-table-leaves`: Include table nodes
- `--page-size`: Page size (omit to use the MOI default)
- `--page-token`: Cursor from the previous page
- `--search`: Search system catalog resources
- `--table-name`: Resolve an exact system table name
- `--workspace-id`: Workspace ID (default: from config)

## catalog table-data

- `--table-id`: tableID
- `--workspace-id`: Workspace ID (default: from config)

## catalog table-exists

- `--database-id`: databaseID
- `--name`: name
- `--workspace-id`: Workspace ID (default: from config)

## catalog table-full-path

- `--table-id`: tableID
- `--workspace-id`: Workspace ID (default: from config)

## catalog table-info

- `--table-id`: tableID
- `--workspace-id`: Workspace ID (default: from config)

## catalog table-overview

- `--database-id`: databaseID
- `--workspace-id`: Workspace ID (default: from config)

## catalog table-references

- `--id`: id
- `--workspace-id`: workspace-id

## catalog tree

- `--include-table-leaves`: Include table nodes
- `--page-size`: Page size (omit to use the MOI default)
- `--page-token`: Cursor from the previous page
- `--resource-id`: Resolve a specific resource ID
- `--resource-types`: Comma-separated resource types
- `--search`: Search catalog resources
- `--workspace-id`: Workspace ID (default: from config)

## catalog trigger-file-workflow

- `--file-ids`: fileIDs (comma-separated)
- `--volume-id`: volumeID
- `--workspace-id`: Workspace ID (default: from config)

## catalog truncate-table

- `--table-id`: tableID
- `--workspace-id`: Workspace ID (default: from config)

## catalog update

- `--catalog-id`: catalogID
- `--name`: name
- `--workspace-id`: Workspace ID (default: from config)

## catalog update-database

- `--database-id`: databaseID
- `--name`: name
- `--workspace-id`: Workspace ID (default: from config)

## catalog update-database-description

- `--database-id`: databaseID
- `--workspace-id`: Workspace ID (default: from config)

## catalog update-folder

- `--description`: Folder description
- `--folder-id`: folderID
- `--name`: Folder name
- `--workspace-id`: Workspace ID (default: from config)

## catalog update-volume

- `--name`: name
- `--volume-id`: volumeID
- `--workspace-id`: Workspace ID (default: from config)

## catalog upload

- `--content-type`: File MIME type
- `--external-source-uri`: Trusted external source URI
- `--file-path`: Local file path (repeat only for connector uploads)
- `--filename`: filename
- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--max-upload-bytes`: Total byte limit for file-path input
- `--workspace-id`: Workspace ID (default: from config)

## catalog upload-bytes

- `--filename`: filename
- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--workspace-id`: Workspace ID (default: from config)

## catalog upload-file

- `--file-path`: filePath
- `--workspace-id`: Workspace ID (default: from config)

## catalog upload-private

- `--content-type`: File MIME type
- `--file-path`: Local file path (repeat only for connector uploads)
- `--filename`: filename
- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--max-upload-bytes`: Total byte limit for file-path input
- `--workspace-id`: Workspace ID (default: from config)

## catalog upload-private-bytes

- `--filename`: filename
- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--workspace-id`: Workspace ID (default: from config)

## catalog upload-private-file

- `--file-path`: filePath
- `--workspace-id`: Workspace ID (default: from config)

## catalog volume-full-path

- `--volume-id`: volumeID
- `--workspace-id`: Workspace ID (default: from config)

## catalog volume-info

- `--volume-id`: volumeID
- `--workspace-id`: Workspace ID (default: from config)

## catalog volume-references

- `--id`: id
- `--workspace-id`: workspace-id

## catalog workspace-database-tables

- `--database-id`: database-id
- `--workspace-id`: workspace-id

## catalog workspace-file-model-url

- `--file-id`: file-id
- `--workspace-id`: workspace-id

## catalog workspace-table-columns

- `--json`: json
- `--json-file`: json-file
- `--workspace-id`: workspace-id

## catalog workspace-table-overview

- `--workspace-id`: workspace-id

