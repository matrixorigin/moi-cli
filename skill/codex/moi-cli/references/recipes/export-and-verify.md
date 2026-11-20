# Recipe：下载数据并核对交付

先区分两种目标：本地下载已有表/查询/图表结果，或通过 Connector 写入外部目标。它们的任务和完成证据不同。

## 下载到本地

在任务目录执行；变量引用本次发现的真实对象，不覆盖用户原文件：

```sh
moi-cli -o -no-interactive -workspace-id "$workspace_id" catalog download-table --table-id "$table_id" --output "$table_csv"
moi-cli -o -no-interactive -workspace-id "$workspace_id" sql download-query-result --statement-id "$statement_id" --output "$query_csv"
moi-cli -o -no-interactive -workspace-id "$workspace_id" dashboard download-chart-result --dashboard-id "$dashboard_id" --chart-id "$chart_id" --output "$chart_csv"
```

查询先完成 describe/query-result，图表先完成 chart-result；下载参数不包含图表 execution ID，故并发刷新时还要核对下载范围和本次执行结果。文件存在或命令退出 0 不足以通过：用 CSV 解析器检查列名、行数、空值、关键记录和汇总，记录文件大小与 SHA256。可小规模逐行比对；不能把大文件全部放入模型上下文。各入口默认格式按实际 help 和文件内容确定，不凭扩展名推断。

Workflow 写入 Volume 的文件走 `catalog download-file --volume-id ID --file-id ID --output PATH`；Agent 发布的 workspace file 走 [智能体文件交付](agent-tools-and-deliver-files.md)，不能混用 ID。

普通 Catalog 文件先用 `catalog files --volume-id` 找到本次 file ID，再用同一 Volume 做
`file-info`、`preview-file` 和 `download-file`。源文件与下载文件至少核对字节数/SHA256；文本或 CSV
还要核对解析内容。`preview-file` 是预览，不保证返回完整文件，完整交付以实际下载产物为准。

2026-11-01 小表对照：table-info、table-data、preview-table、download-table 的3列6行一致，
CSV独立合计1500。只覆盖当前完整小表；默认预览窗口或 table-info 统计不保证任意大表已全部导出。

## 导出到外部系统

1. 固定源表、目标 Connector、目标对象、字段映射及追加/覆盖语义。目标凭证通过产品已有安全路径提供。
2. 读取 `schema connector create-export-task` 和对应源类型合同，验证请求文件后创建一次任务。
3. 保存 export task ID，用 `export-task-state`、`export-task-info`、`export-task-files` 查原任务。
4. 核验目标端本次数据范围及失败/跳过文件；重试前检查幂等与覆盖风险。

没有外部目标或合法连接时保持未验收；表 CSV 下载通过不证明 Connector 外部导出任务通过。

结构化文件的本地导出至少同时保留表/查询 ID、下载路径、字节数、SHA256、CSV 列名和解析记录数。
表行数由独立 SQL `COUNT(*)` 核对；NULL/空字符串由 typed JSON 或 SQL 布尔表达式核对，不能只靠
CSV 空字段。可复用[Catalog 文件与结构化表核验](verify-catalog-file-and-table.md)的证据清单。

## 展示窗口、完整CSV与历史结果

2026-11-02合成表1007行：table-data/preview各返回100行且total_rows=1007；
异步query-result返回1000行、total=1000、truncated=true，而表CSV及同一原statement CSV都完整1007行，
18091字节且字节一致。不要用预览或result.total冒充完整表行数；独立COUNT/业务主键/聚合与真实下载相互核对。
这只覆盖1007行小样本，当前query-result没有已确认分页输入，不能承诺任意大结果。

本次execute请求limit=2000仍是1000行结果窗口；另一请求limit=3加SQL LIMIT 9，结果及CSV实际9行。
不能把CLI limit当严格最多返回行数的保证，也不能推断所有入口均忽略它。需要严格业务范围时，
在获准SQL中显式限定并核对实际结果；没有完整合同就保留范围限制，不猜分页/续传参数。

CSV使用真正的解析器并保留换行（Python打开时`newline=''`），按业务主键逐字段核对，
不要按文本行数统计含换行字段的记录，也不假设无ORDER BY文件顺序。本次中文、emoji、逗号、引号、
真实LF/CRLF/TAB与尾空格均原样保留。普通CSV把NULL和空串都写为空字段，需要区分时同时检查typed JSON
的有效性字段；金额用Decimal或等价精确小数，不用浮点近似充当逐格一致。

区分“当前表下载”和“已完成查询的原statement下载”。源修改后本次当前表1006行/5165.11，
两条原statement的完整JSON/CSV不变；删除源表/库并回读404后，原查询结果仍可读且相同。
要当前数据须明确另起新查询；不能把历史CSV当当前表，也不能把删源称为删除了历史查询数据。
短时间历史保持不承诺无限保留期或权限撤回后仍可读取。
