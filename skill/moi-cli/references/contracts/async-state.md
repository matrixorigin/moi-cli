# 异步状态合同

异步任务分为四类事实：提交事实、运行事实、业务结果事实和产物事实。

- **提交事实**：命令接受并返回 task/query/execution/job ID。
- **运行事实**：详情接口返回当前状态、错误字段和可继续动作。
- **业务结果事实**：结果接口返回 rows、node states、source/job 状态、chart result 或目标资源状态。
- **产物事实**：按用户要求下载或检查文件、表、知识源或目标端对象。

轮询要求：

1. 使用返回的 ID，不重发原始写命令。
2. 只轮询文档声明的 pending 状态，并设置有限次数或时间预算。
3. 每次轮询记录状态变化和服务端错误；状态不变时不扩展结论。
4. 达到预算仍 pending 时报告 pending 和最后状态。
5. 失败、取消、拒绝或参数错误属于终态，先保存错误再判断是否有用户授权的恢复动作。

常见 ID 禁止混用：`query_id`、`statement_id`、`workflow_id`、`execution_id`、`task_id`、
`model_id`、`source_id`、`job_id` 和 `file_id` 各自只用于对应接口。

## 对象类型账本

跨 Catalog、Workflow、Knowledge 的任务必须在本地账本中为每个返回值记录“字段名、对象类型、
来源命令、workspace、上游 ID和可用查询命令”。推荐至少使用以下字段名：

| 字段 | 对象 | 可用于 | 不能替代 |
|---|---|---|---|
| `workflow_id` | Workflow/App | workflow 详情、运行和版本查询 | execution、batch task |
| `workflow_version_id` | 已发布版本 | 版本和执行绑定核对 | workflow_id、execution |
| 子记录的 `execution_id` | 单次/单文件 Workflow execution | get-run、get-run-result、import files | dispatch job、case、file |
| `dispatch_job_id`（run 可能错放在 `workflow_run.execution_id`） | Volume 批量分发任务 | list-runs 精确关联子记录、失败文件重试的 source-dispatch-job-id 输入 | 子 execution、case、file |
| `file_id` | Catalog 文件 | 下载、file-data-asset、lineage、Knowledge source 关联 | source_id、job_id |
| `model_id` / `source_id` / `source_job_id` | Knowledge 资源及准备任务 | source 详情、job、分段和引用核验 | file_id、workflow execution |

任何命令返回多个 `id` 时，先按字段名和后续可查询接口分类；如果服务只给无类型的 ID，保留
`unknown_id` 并停止依赖它的后续调用，不把 404 直接归因于服务故障。

Volume 输入必须结合配置与 `list-runs` 的 `dispatch_job_id` 判断返回 ID 的语义，见[Workflow](../domains/workflow.md#单次执行与-volume-批处理必须分开)。字段名、`one_shot` 和 available_actions 都不能单独证明对象类型。
