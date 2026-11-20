# Recipe：在本地筛选、排序和分页 CLI 结果

适用：CLI 已能取得完整或有界列表，用户需要在 Codex 中按名称、描述、类型等查找和组织结果。
优先使用命令正式提供的服务端筛选参数。只有列表已经取回、且用户接受客户端后处理时，
才使用本 Recipe；它不能证明服务端搜索、页面筛选或全量数据已经覆盖。

`scripts/json_query.py` 读取 CLI 原始 JSON，也读取 `record_cli.py` 或验收记录器保存的
`response.json`。它不会发起网络请求。输出固定标记 `mode=client_side_post_processing`、
`server_filter_claim=false` 和 `product_ui_claim=false`。

例如，从知识库列表按名称和描述模糊查找，再按名称排序，只输出必要字段：

```text
python3 scripts/json_query.py calls/kb-list/response.json \
  --items-path items --search-text kb046 \
  --search-field name --search-field description \
  --sort name --select id --select name --select source_counts --limit 20
```

精确筛选和本地分页：

```text
python3 scripts/json_query.py calls/model-list/response.json \
  --where model_type=chat --sort model --select model --select backend_name \
  --offset 0 --limit 20
```

规则：

1. 先记录源列表的真实数量和分页范围；源列表本身不完整时，本地筛选只能覆盖已取回的子集。
2. `--search-text` 是不区分大小写的子串匹配，`--where FIELD=VALUE` 是精确匹配；二者不要
   写成服务端全文检索。字段用点路径，例如 `files.image_index_status`。
3. 用 `--select` 只保留判断所需字段，避免把配置正文或敏感内容复制到报告。脚本会对常见
   凭证字段做脱敏，但不能识别任意业务秘密。
4. 输出中的 `source_count` 是源列表条数，`total` 是本地筛选后的数量，`returned` 是当前
   本地页条数。它们与服务端总数分开解释。
5. 对通用 `-query-param` 的探索必须检查结果数量和实际内容。未知参数可能被静默忽略；
   返回与未筛选完全相同不能证明筛选生效，更不能据此认定参数正确。
