# Recipe：Catalog 文件与结构化表核验

适用：在已选 workspace 的隔离数据库/Volume 中，核验本地文件上传挂载、目录筛选、预览下载、
CSV 入表和表/查询 CSV 下载。上传、挂载和导入是三个不同事实；任何一步状态未知都先查原对象，
不重发写命令。

## 1. 保存范围与真实 ID

记录 workspace、输入文件字节数/SHA256、隔离 database/Volume/folder 名称和预期行数。通过正式
tree/list/create 回执保存真实 `database_id`、`volume_id`、`folder_id`，不能按名称或数组顺序补 ID。
创建不是本 Recipe 的隐式授权；已有资源必须明确获准使用，新资源必须唯一命名且不碰业务数据。

## 2. 上传、挂载与目录筛选

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" catalog upload --file-path "$file_path"
moi-cli -o -no-interactive -workspace-id "$workspace_id" catalog create-file --volume-id "$volume_id" --file-ids "$file_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" catalog volume-info --volume-id "$volume_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" catalog files --volume-id "$volume_id" --parent-id "" --file-ids "$file_id"
```

`file_id` 来自上传回执；挂载回执和 Volume 数量仍不能替代列表中的真实条目。根目录使用空
`parent-id`；子文件夹也使用空或省略的 `parent-id`，但 `volume-id` 必须换成该 folder 自己的真实
Volume ID。2026-11-14 QA 中两种非空 parent 组合均返回 400，不把 folder 输出的 `parent_id` 直接
当成可用筛选输入。按名称、扩展名、file IDs、`file/folder` 类型筛选时保存全部条件和返回 ID；
翻页保持 `volume_id/parent-id/筛选/排序` 不变，不把第一页当完整清单。Volume 汇总可能递归统计
后代文件，当前目录列表则可能包含 folder 而不含后代文件，须按实际范围解释数量和字节数。

## 3. 详情、预览与下载互证

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" catalog file-info --file-id "$file_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" catalog preview-file --volume-id "$volume_id" --file-id "$file_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" catalog download-file --volume-id "$volume_id" --file-id "$file_id" --output "$downloaded_file"
```

核对返回的 Volume/placement 与本次目标一致。比较小文件内容与源文件；其它文件至少比较字节数和
SHA256。上传收据不证明可预览，预览不证明完整下载，下载存在也不证明入表或可检索。

## 4. CSV 入表与字段/NULL 核验

按[文件导入到表](import-file-to-table.md)选择一次 `connector upload` 流程，显式保存 table-config、
CSV 方言和 `csv.null_values`。保存 `task_id`、run、table ID 和 `latest_rows`，只轮询原任务；accepted、
任务完成和行数正确是不同证据。

对合成样例至少包含一个 NULL、一个空字符串、一个含逗号/引号字段和一个字段内换行。用正式 SQL
execute → describe → query-result 流程执行有界核验，示例表达式按实际表字段调整：

```sql
SELECT id, amount, note,
       note IS NULL AS note_is_null,
       note = '' AS note_is_empty,
       LENGTH(note) AS note_length
FROM isolated_table
ORDER BY id;

SELECT COUNT(*) AS row_count,
       SUM(CASE WHEN note IS NULL THEN 1 ELSE 0 END) AS null_count,
       SUM(CASE WHEN note = '' THEN 1 ELSE 0 END) AS empty_count,
       SUM(amount) AS amount_sum
FROM isolated_table;
```

保存 `query_id` 与 `statement_id`，核对字段元数据、每行有效性和精确小数合计。`note = ''` 不命中
NULL，不能用显示层的 `NULL` 文本、CSV 空单元格或字符串长度代替这两个判断。
2026-11-14 QA 的 MatrixOne 返回 `invalid argument aggregate function sum, bad value [BOOL]`，
因此不要直接 `SUM(note IS NULL)`；上面的 `CASE WHEN` 已实测返回 4 行、NULL=1、空串=1。

## 5. 表与查询结果下载

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" catalog download-table --table-id "$table_id" --output "$table_csv"
moi-cli -o -no-interactive -workspace-id "$workspace_id" sql download-query-result --statement-id "$statement_id" --output "$query_csv"
```

用 CSV 解析器按记录而不是物理行读取，核对列名、记录数、业务主键、字段内换行、字节数和 SHA256。
表下载与独立 COUNT、同一 statement 的结果/下载互证；预览窗口或 query-result 的 total/truncated 不
能冒充完整表行数。普通 CSV 中 NULL 与空串可能都为空字段，最终区分仍依赖上面的 typed JSON/SQL。

本 Recipe 的“导出”是本机表/查询 CSV 下载。外部 Connector 导出必须有明确目标、映射和写入授权，
并用原 export task 状态及目标端内容另行验收；本地 CSV 通过不能替代该结论。
