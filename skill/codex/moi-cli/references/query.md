# Discover and query data

Use the workspace selected under SKILL.md. For a known table, inspect columns
unless already established in this conversation:

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" sql meta-table-columns --db-name "$database" --table-name "$table"
```

Read `columns[].name` and `columns[].data_type` to construct a bounded SELECT; quote
SQL identifiers separately from values. For unknown assets, use the relevant
[Catalog](commands/catalog.md) or [SQL metadata](commands/sql.md) command instead
of guessing identifiers or creating example tables.

For a row preview, include an explicit SQL `LIMIT` as well as the CLI `--limit`.
The result endpoint may report its own limit (observed: 1000 despite requested 1);
record requested and returned bounds separately. Verify actual row values and counts.
Creating synthetic fixtures is appropriate only for a user-authorized isolated test,
not as a substitute for discovering the user's existing data.

## Execute → describe → result

`$query` is the requested read-only SQL and `$database` is the selected database.
Run these steps separately, using each response to supply the next ID:

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" sql execute --db "$database" --query "$query" --limit 20
moi-cli -o -no-interactive -workspace-id "$workspace_id" sql describe-query --query-id "$query_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" sql query-result --statement-id "$statement_id"
```

1. Execute returns `query_id` (also exposed as `id`), **not rows**. Set
   `$query_id` from that response. Do not resubmit the query to poll it.
2. Describe returns `status`, `err_msg` and `statement_id`. While running, poll
   describe with a bounded wait. Continue after `status` is `success` and a
   nonempty `statement_id` is present; report other terminal outcomes or an
   exhausted wait budget. Set `$statement_id` from describe, never from `id`.
3. Result returns `columns`, `result` (rows) and `total`. Check its `status` and
   report the returned sample and limit, not an assumed complete dataset. Empty
   rows can be a valid result. A missing statement ID is not a reason to search
   query history, whose storage may lag behind this execution chain.

Stop after returning the requested rows/summary. Do not execute the same query
again to repair a result-fetch failure; retain both IDs and the known status.
Only poll states documented as pending, not permission/parameter failures.

For a selected SQL file, replace `--query` with `--query-file "$sql_file"`.
Do not query unrelated data or credential tables to diagnose access problems.
