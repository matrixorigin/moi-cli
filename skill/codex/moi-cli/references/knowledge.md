# Knowledge sources

## Create a knowledge base from a selected local document

Use the selected workspace, the user's `$file_path` and the requested `$name`.
Only follow this creation flow when the user asks for a new knowledge base:

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" semantic-model upload-local-file --file-path "$file_path"
moi-cli schema semantic-model create-with-sources
```

Upload returns `file_id` and may return `original_name`; if absent, retain the filename explicitly uploaded in this operation. Construct `$sources_file` as a JSON
array of CLI source inputs using `SourceType: "local_file"`, `FileID` from
`file_id` and `FileName` from the established upload filename. These are CLI input field names,
not HTTP body keys; use real values, not the local path as `FileID`.

```sh
moi-cli schema semantic-model create-with-sources --validate-file "$sources_file"
moi-cli -o -no-interactive -workspace-id "$workspace_id" semantic-model create-with-sources --name "$name" --json-file "$sources_file"
moi-cli -o -no-interactive -workspace-id "$workspace_id" semantic-model list-source-jobs --model-id "$model_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" semantic-model list-sources --model-id "$model_id"
```

Create returns `model.id`, `sources` and `jobs`. Use `model.id` as `$model_id`.
Job inspection returns `items[]` with `job_id`, `job_status`, `source_file_id`
and `error`; match jobs to this upload instead of treating unrelated jobs as
its outcome. Sources also come in `items[]`; match `source_file_id` to the
uploaded `file_id` and retain `row_id` as the source ID for later inspection.
Check parsing/indexing state: accepting a file is not search readiness.
Preserve the model, source and job IDs if preparation is pending or fails; do
not create another knowledge base to retry an inspection call. Report pending
preparation as pending rather than continuing without a time bound.

For an explicitly requested preparation, follow the controlled advancement and readiness checks in
[the preparation recipe](recipes/create-and-prepare-knowledge-base.md). For questions with sources,
continue with [answer and citation verification](recipes/query-knowledge-and-check-citations.md).

When a source is produced by a Workflow or comes from a Catalog Volume, use the
[cross-domain lineage recipe](recipes/catalog-workflow-knowledge-pipeline.md).
Match the same `file_id` across Catalog lineage, Knowledge source/job and the
formal answer `source_refs`; do not use a filename, an old execution or an empty
job list as a substitute.

For an existing knowledge base, resolve its ID first, use
`upload-model-local-file --model-id` and `append-sources` from the
[semantic-model reference](commands/semantic-model.md). Do not create a new
knowledge base as a substitute. Entry/segment changes and source deletion are
separate tasks, not automatic steps in importing a document.
