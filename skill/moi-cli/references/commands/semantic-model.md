# semantic-model command reference

Generated from CLI source. Do not edit. Business requirements remain server-owned.
Use `moi-cli schema DOMAIN COMMAND` for typed JSON input and `--help` for defaults.

## semantic-model append-sources

- `--json`: JSON array of source objects; use --help or schema for source types and required fields
- `--json-file`: File containing the same source array as --json
- `--model-id`: modelID
- `--source-selections-file`: JSON file containing Catalog source selections
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model backfill-legacy-sources

- `--model-id`: modelID
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model check-source-existence

- `--file-ids`: fileIDs (comma-separated)
- `--model-id`: modelID
- `--table-ids`: tableIDs (comma-separated)
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model cite

- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--output`: Write the complete result to a new local file
- `--from`: Saved inventory, evidence result or previous source page (repeatable for citations)
- `--indices`: Evidence indices for citation or batch reads; citations may pair each --from with its own selection
- `--request-id`: Stable ID for replay of this exact request
- `--quote`: Exact inspected text identifying one canonical passage; pair each --from with one --quote

## semantic-model create

- `--description`: Knowledge base description
- `--image-index-enabled`: Enable image vector index
- `--name`: knowledge base name and immutable Catalog database name
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model create-empty

- `--description`: knowledge base description
- `--image-index-enabled`: Enable image vector index
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

- `--description`: Knowledge base description
- `--files-file`: JSON file containing semantic file configuration
- `--image-index-enabled`: Enable image vector index
- `--json`: JSON array of source objects; use --help or schema for source types and required fields
- `--json-file`: File containing the same source array as --json
- `--name`: knowledge base name and immutable Catalog database name
- `--source-selections-file`: JSON file containing Catalog source selections
- `--target-catalog-id`: Target Catalog ID (omit to let MOI choose)
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model create-with-spec

- `--description`: Knowledge base description
- `--files-file`: JSON file containing semantic file configuration
- `--json`: JSON creation spec with name and embedding/drawing controls; see moi-cli schema semantic-model create-with-spec
- `--json-file`: JSON file containing the creation spec
- `--tables-file`: JSON file containing semantic table definitions
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

## semantic-model describe-tables

- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--output`: Write the complete result to a new local file
- `--request-id`: Stable ID for replay of this exact request
- `--keywords`: Repeatable comma-separated schema keywords

## semantic-model export

- `--model-id`: modelID
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model find-files

- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--output`: Write the complete result to a new local file
- `--request-id`: Stable ID for replay of this exact request
- `--query`: File name or inventory query

## semantic-model get

- `--model-id`: modelID
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model get-drawing-provenance

- `--knowledge-base-id`: knowledge_base_id
- `--page-number`: optional page number; 0 returns every page
- `--source-file-id`: catalog source file id
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model get-source-document

- `--model-id`: modelID
- `--query-param`: Product query parameter name=value (repeatable; pagination/filter names follow the product API)
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

- `--query-param`: Product query parameter name=value (repeatable; pagination/filter names follow the product API)
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model list-entries

- `--model-id`: modelID
- `--query-param`: Product query parameter name=value (repeatable; pagination/filter names follow the product API)
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model list-passages

- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--output`: Write the complete result to a new local file
- `--from`: Saved inventory, evidence result or previous source page (repeatable for citations)
- `--index`: Zero-based evidence index for a read
- `--request-id`: Stable ID for replay of this exact request
- `--cursor`: Exact continuation cursor from this source's previous page
- `--page-size`: Enabled text passages per page (default 20, maximum 100)

## semantic-model list-referencing-agents

- `--cursor`: Cursor from the previous page
- `--limit`: Page size (omit to use the product default)
- `--model-id`: Knowledge base ID
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model list-source-jobs

- `--model-id`: modelID
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model list-sources

- `--include-display-status`: Include source display status
- `--model-id`: modelID
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model list-sources-page

- `--include-display-status`: Include source display status
- `--model-id`: modelID
- `--page`: page
- `--page-size`: pageSize
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model list-tags

- `--query-param`: Product query parameter name=value (repeatable; pagination/filter names follow the product API)
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

## semantic-model query

- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--output`: Write the complete result to a new local file
- `--request-id`: Stable ID for replay of this exact request

## semantic-model read-document

- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--output`: Write the complete result to a new local file
- `--from`: Saved inventory, evidence result or previous source page (repeatable for citations)
- `--index`: Zero-based evidence index for a read
- `--request-id`: Stable ID for replay of this exact request

## semantic-model read-evidence

- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--output`: Write the complete result to a new local file
- `--from`: Saved inventory, evidence result or previous source page (repeatable for citations)
- `--indices`: Evidence indices for citation or batch reads; citations may pair each --from with its own selection
- `--index`: Zero-based evidence index for a read
- `--request-id`: Stable ID for replay of this exact request
- `--local-read`: Register a complete locally inspected passage; verify its digest against this task's evidence
- `--quote`: Exact inspected text identifying one canonical passage; pair each --from with one --quote

## semantic-model read-source

- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--output`: Write the complete result to a new local file
- `--from`: Saved inventory, evidence result or previous source page (repeatable for citations)
- `--index`: Zero-based evidence index for a read
- `--request-id`: Stable ID for replay of this exact request
- `--source-name`: Resolve exactly one selected source by filename query
- `--page-size`: Enabled text passages per page (default 100, maximum 100)

## semantic-model reconcile-source-jobs

- `--model-id`: modelID
- `--retry-source-row-id`: Retry this failed knowledge source row ID (omit to reconcile without retry)
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model reembed-segments

- `--json`: JSON for complex arguments
- `--json-file`: JSON file for complex arguments
- `--model-id`: modelID
- `--source-id`: sourceID
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model search

- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--output`: Write the complete result to a new local file
- `--from`: Saved inventory, evidence result or previous source page (repeatable for citations)
- `--index`: Zero-based evidence index for a read
- `--request-id`: Stable ID for replay of this exact request
- `--fulltext-only`: Use full-text index matching without embedding or vector candidates
- `--keywords`: Repeatable comma-separated retrieval keywords
- `--max-hits`: Candidate limit per query, retrieval route and index group (default 8); merged results and neighboring passages may exceed it

## semantic-model search-document

- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--output`: Write the complete result to a new local file
- `--from`: Saved inventory, evidence result or previous source page (repeatable for citations)
- `--index`: Zero-based evidence index for a read
- `--request-id`: Stable ID for replay of this exact request
- `--query`: Text to find within parsed Markdown

## semantic-model search-drawings

- `--json`: JSON DrawingSearchRequest (overrides individual flags when set)
- `--json-file`: JSON file for DrawingSearchRequest
- `--knowledge-base-id`: knowledge_base_id
- `--query`: text query
- `--query-image`: local query image path
- `--query-image-filename`: filename sent with the query image (default: local path base)
- `--top-k`: top_k
- `--workspace-id`: Workspace ID (default: from config)

## semantic-model search-images

- `--json`: Typed request JSON
- `--json-file`: Typed request JSON file
- `--output`: Write the complete result to a new local file
- `--request-id`: Stable ID for replay of this exact request
- `--query`: Description of visual evidence

## semantic-model update

- `--description`: Knowledge base description
- `--files-file`: JSON file containing semantic file configuration
- `--model-id`: modelID
- `--name`: existing immutable knowledge base and Catalog database name (rename unsupported)
- `--tables-file`: JSON file containing semantic table definitions
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

- `--clear-expires-at`: Clear expiry (exclusive with --expires-at)
- `--enabled`: Enable or disable the source; omission preserves its state
- `--expires-at`: Expiry as Unix seconds; omission preserves the current expiry
- `--force-enabled-after-expiry`: Allow or disallow use after expiry; omission preserves the current setting
- `--model-id`: modelID
- `--source-id`: sourceID
- `--tags`: JSON string array of replacement tags; [] clears tags
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

