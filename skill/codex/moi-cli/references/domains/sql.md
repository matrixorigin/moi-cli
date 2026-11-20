# SQL 领域

## 元数据

未知表先获取列定义：

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" sql meta-table-columns --db-name "$database" --table-name "$table"
```

读取返回的列名和类型后再构造有界查询。不要查询无关凭证表来诊断权限。

## 执行闭环

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" sql execute --db "$database" --query "$query" --limit 20
moi-cli -o -no-interactive -workspace-id "$workspace_id" sql describe-query --query-id "$query_id"
moi-cli -o -no-interactive -workspace-id "$workspace_id" sql query-result --statement-id "$statement_id"
```

`execute` 返回 `query_id`，不是 rows；`describe-query` 返回 `statement_id`。只有 status 为
success 且 statement ID 非空时才调用 `query-result`。结果中的 `columns`、`result`、`total`
只代表本次 limit 和返回范围；空 rows 也可能是合法结果。

DDL/DML、删除/截断、工作簿保存和查询取消具有副作用，不能从简单 SELECT 的成功推断这些
能力已通过。查询历史只能用于查看历史事实，不能代替当前执行链的 statement ID。

## 工作簿与已知限制

工作簿保存和版本回读见[工作簿Recipe](../recipes/manage-sql-workbook.md)。
2026-10-31当前CLI的query-detail/query-profile只暴露statement-id，但服务要求同时提交非零start/end时间范围；
自己的真实statement也返回400。保留为CLI缺输入，不猜时间参数、不改用raw HTTP，也不误报登录失败。
本轮sql run执行SELECT SLEEP(0)时返回405（两种认证相同），但共享表的有界SELECT通过sql run成功。
因此不能判定sql run整体不可用；常规execute→describe→query-result在本轮工作簿/导入核验中已通过，按实际查询记录差异。

本轮空表有界SELECT的query-result返回500 `invalid input: empty query result`，因此错误响应不能当作零行证据；
需要确认内容时另做COUNT等有实际结果行的查询。本次COUNT独立证明零行，不推断所有空结果行为均相同。

后续补测已确认正式内联`sql run`能对合法零行SELECT返回success及列结构，result/total零值字段缺省；
原异步statement取结果和CSV下载仍500。任务开始可以按目标选择内联只读路径，详见
[合法零行查询的内联路径](../recipes/run-sql-and-fetch-result.md#合法零行查询的内联路径)。
已有失败不能静默重执行；内联生成新的query/statement，不是取回原查询历史快照，不能用于重跑副作用SQL。
