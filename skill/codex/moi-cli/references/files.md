# Local files and import/export

Use the selected workspace and a user-selected `$file_path`. Reuse an existing
successful upload receipt from this task. Choose one flow for the requested
outcome, not all of the following flows. If "upload to MOI" could mean either
staging bytes or placing a file in a visible platform folder, clarify the
destination before writing; do not silently choose table/knowledge import.

- catalog upload / upload-private upload to their respective Catalog endpoints.
  Private upload is explicit, not a fallback for a denied shared destination.
- connector local-upload stages files and returns service references.
- connector upload submits the existing import workflow. It can return accepted,
  partially accepted or rejected. Do not create a second task automatically.
- semantic-model uploads serve knowledge inputs; agent upload-workspace-file
  serves MOI agent files, not local chat attachments.

These reader upload commands accept --file-path. Connector accepts repeated paths.
Pass paths as quoted literal arguments; do not interpret shell metacharacters,
expand arbitrary globs, scan home directories or read secrets. New path input is
limited to 64 MiB total by default; --max-upload-bytes explicitly changes this
client memory guard, not server limits. No resumable/unlimited upload is promised.

## Import

For staging files for a Connector import:

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" connector local-upload --file-path "$file_path"
```

The response contains `conn_file_ids` and `file_ids`. Preserve each kind under
its original name; they are not local paths or interchangeable IDs. Resolve the
selected source connector and destination from actual resources. If creating a
typed import task, inspect its schema and construct `$request_file` with those
references and the user's format/table mapping choices:

```sh
moi-cli schema connector create-task
moi-cli schema connector create-task --validate-file "$request_file"
moi-cli -o -no-interactive -workspace-id "$workspace_id" connector create-task --json-file "$request_file"
moi-cli -o -no-interactive -workspace-id "$workspace_id" connector get-task --task-id "$task_id"
```

Create returns `task_id`; get-task returns `task`. Creation does not mean the
task ran. Use [Connector commands](commands/connector.md) to inspect task runs
and files, and start only when the selected operation requires it. After import,
check the destination table with a bounded [SQL query](query.md).

Direct `connector upload` is a different import flow: its `--volume-id`, `--meta`
and `--table-config` inputs depend on the selected destination and file format.
Inspect acceptance and rejected files; do not run both upload flows for one
request or substitute a URI for a file ID.

## Upload without importing (staging bytes)

For a Catalog upload only:

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" catalog upload --file-path "$file_path"
```

The product response returns `file_id`. Preserve it with the source path and
workspace; do not require `original_name` or `catalog_file`, which this endpoint
does not return. A successful response completes an explicitly requested
staging upload. Report that result and stop, without unrelated follow-up calls.

`catalog file-info` and `catalog file-data-asset` require the file to resolve to
a canonical Volume. They are not general verification endpoints for an
unattached upload. An `ErrIAMResourceNotFound` / missing Volume response does
not revoke the upload receipt or prove a broken environment. Do not repeat the
upload, switch identity or try other endpoints to bypass that boundary.

## Placement in a platform folder is a separate operation

The backend separately associates an uploaded file ID with a selected Volume;
uploading bytes alone does not perform that operation. Discover/confirm the
user's intended destination before promising that the file will appear there.

Attach uploaded files to the confirmed Volume with
`catalog create-file --volume-id <id> --file-ids <id1,id2>`. Pass the exact
file IDs from the upload receipt; the command fails with `ErrFileIDsRequired`
when none are given. `catalog files` does not expose the required Volume
filter, so do not invent `--filters` or switch to raw HTTP/Core APIs to list a
Volume. Do not create a different resource or upload another copy as a
workaround.

A file ID is not a SQL table, and a raw upload is not a searchable knowledge
source. Use the import or knowledge flow only when that is the requested goal.
For attaching files to an MOI agent, consult the [Agent reference](commands/agent.md).

## Export

Inspect `moi-cli schema connector create-export-task`, resolve authorized source
and format, then inspect export-task-state/info/files. Download only through a
supported path; never send credentials to arbitrary returned URLs. Ask before
overwriting a local file. A returned export ID is not a completed download.

Upload failure stops downstream processing; never replace missing data with an
empty file or claim that uploaded means imported/queryable.
