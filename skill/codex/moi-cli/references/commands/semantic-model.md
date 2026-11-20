# semantic-model command reference

Generated from CLI source. Do not edit. Business requirements remain server-owned.
Use `moi-cli schema DOMAIN COMMAND` for typed JSON input and `--help` for defaults.

## semantic-model append-sources

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--model-id`: modelID
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model backfill-legacy-sources

- `--model-id`: modelID
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model check-source-existence

- `--file-i-ds`: fileIDs (comma-separated)
- `--model-id`: modelID
- `--table-i-ds`: tableIDs (comma-separated)
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model create

- `--name`: knowledge base name and immutable Catalog database name
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model create-empty

- `--description`: knowledge base description
- `--image-index-enabled`: enable image vector index (image_index_enabled=true)
- `--name`: knowledge base name and immutable Catalog database name
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model create-entry

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--model-id`: modelID
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model create-segment

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--model-id`: modelID
- `--source-id`: sourceID
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model create-with-sources

- `--image-index-enabled`: enable image vector index (image_index_enabled=true)
- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--name`: knowledge base name and immutable Catalog database name
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model delete

- `--model-id`: modelID
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model delete-entry

- `--entry-id`: entryID
- `--model-id`: modelID
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model delete-segment

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--model-id`: modelID
- `--segment-id`: segmentID
- `--source-id`: sourceID
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model delete-source

- `--model-id`: modelID
- `--source-id`: sourceID
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model export

- `--model-id`: modelID
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model get

- `--model-id`: modelID
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model get-source-document

- `--model-id`: modelID
- `--source-id`: sourceID
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model import

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--model-id`: modelID
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model import-model

- `--json`: json
- `--json-file`: json-file
- `--workspace-id`: workspace-id

## semantic-model import-models

- `--json`: json
- `--json-file`: json-file
- `--workspace-id`: workspace-id

## semantic-model list

- `--workspace-id`: Workspace ID (default: from config)

## semantic-model list-entries

- `--model-id`: modelID
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model list-referencing-agents

- `--cursor`: Cursor from the previous page
- `--limit`: Page size
- `--model-id`: Knowledge base ID
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model list-source-jobs

- `--model-id`: modelID
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model list-sources

- `--model-id`: modelID
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model list-sources-page

- `--model-id`: modelID
- `--page`: page
- `--page-size`: pageSize
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model list-tags

- `--workspace-id`: Workspace ID (default: from config)

## semantic-model preview-artifact

- `--file-id`: file-id
- `--model-id`: model-id
- `--output`: output
- `--workspace-id`: workspace-id

## semantic-model preview-model-source-selection-counts

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--model-id`: modelID
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model preview-source-file

- `--file-id`: file-id
- `--model-id`: model-id
- `--output`: output
- `--workspace-id`: workspace-id

## semantic-model preview-source-selection-counts

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model reconcile-source-jobs

- `--model-id`: modelID
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model reembed-segments

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--model-id`: modelID
- `--source-id`: sourceID
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model search-drawings

- `--json`: JSON DrawingSearchRequest (overrides individual flags when set)
- `--json-file`: JSON file for DrawingSearchRequest
- `--knowledge-base-id`: knowledge_base_id
- `--query`: text query
- `--query-image`: local query image path
- `--top-k`: top_k
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model update

- `--model-id`: modelID
- `--name`: existing immutable knowledge base and Catalog database name (rename unsupported)
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model update-entry

- `--entry-id`: entryID
- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--model-id`: modelID
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model update-segment

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--model-id`: modelID
- `--segment-id`: segmentID
- `--source-id`: sourceID
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model update-segment-enabled

- `--enabled`: enabled
- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--model-id`: modelID
- `--segment-id`: segmentID
- `--source-id`: sourceID
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model update-source-governance

- `--model-id`: modelID
- `--source-id`: sourceID
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model upload-local-file

- `--file-path`: Local file path (repeat only for connector uploads)
- `--filename`: filename
- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--max-upload-bytes`: Total byte limit for file-path input
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model upload-model-local-file

- `--file-path`: Local file path (repeat only for connector uploads)
- `--filename`: filename
- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--max-upload-bytes`: Total byte limit for file-path input
- `--model-id`: modelID
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model validate

- `--model-id`: modelID
- `--workspace-id`: Workspace ID (default: from config)
